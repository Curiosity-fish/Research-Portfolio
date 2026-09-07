package service

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/ent/account"
	"github.com/school-api/school-api-v1/ent/platform"
	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/relay"
	"github.com/school-api/school-api-v1/internal/repository"
)

// randInt is a testable wrapper around crypto/rand.Int.
var randInt = func(max *big.Int) (*big.Int, error) {
	return rand.Int(rand.Reader, max)
}

// RelayInput carries the authenticated caller identity and raw request body.
type RelayInput struct {
	UserID      uuid.UUID
	TokenID     uuid.UUID
	GroupID     uuid.UUID
	RequestBody []byte
}

// RelayUsage holds token counts parsed from an upstream response.
type RelayUsage struct {
	PromptTokens     int64
	CompletionTokens int64
	TotalTokens      int64
}

// RelayResult is the outcome of a non-streaming relay.
type RelayResult struct {
	StatusCode int
	Header     http.Header
	Body       []byte
	Usage      RelayUsage
}

// ModelInfo represents an entry in the /v1/models list.
type ModelInfo struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}

// RelayService defines AI relay business logic.
type RelayService interface {
	ChatCompletions(ctx context.Context, input RelayInput) (*RelayResult, error)
	ChatCompletionsStream(ctx context.Context, input RelayInput, w http.ResponseWriter) error
	Messages(ctx context.Context, input RelayInput) (*RelayResult, error)
	MessagesStream(ctx context.Context, input RelayInput, w http.ResponseWriter) error
	ListModels(ctx context.Context, groupID uuid.UUID) ([]ModelInfo, error)
}

type relayService struct {
	modelRepo         repository.AIModelRepository
	groupPlatformRepo repository.GroupPlatformRepository
	platformRepo      repository.PlatformRepository
	accountRepo       repository.AccountRepository
	callLogRepo       repository.CallLogRepository
	accountService    AccountService
	billingService    BillingService
	forwarder         relay.Forwarder
	encryptKey        []byte
	maxBodyBytes      int64
	maxRetries        int
}

// NewRelayService creates a new RelayService.
func NewRelayService(
	modelRepo repository.AIModelRepository,
	groupPlatformRepo repository.GroupPlatformRepository,
	platformRepo repository.PlatformRepository,
	accountRepo repository.AccountRepository,
	callLogRepo repository.CallLogRepository,
	accountService AccountService,
	billingService BillingService,
	forwarder relay.Forwarder,
	encryptKey []byte,
	maxBodyBytes int64,
	maxRetries int,
) RelayService {
	return &relayService{
		modelRepo:         modelRepo,
		groupPlatformRepo: groupPlatformRepo,
		platformRepo:      platformRepo,
		accountRepo:       accountRepo,
		callLogRepo:       callLogRepo,
		accountService:    accountService,
		billingService:    billingService,
		forwarder:         forwarder,
		encryptKey:        encryptKey,
		maxBodyBytes:      maxBodyBytes,
		maxRetries:        maxRetries,
	}
}

// ChatCompletions routes a non-streaming OpenAI chat completion request upstream.
func (s *relayService) ChatCompletions(ctx context.Context, input RelayInput) (*RelayResult, error) {
	return s.relayNonStream(ctx, input, platform.TypeOpenai, "/chat/completions", parseOpenAIUsage)
}

// ChatCompletionsStream forwards a streaming OpenAI chat completion request upstream.
func (s *relayService) ChatCompletionsStream(ctx context.Context, input RelayInput, w http.ResponseWriter) error {
	return s.relayStream(ctx, input, w, platform.TypeOpenai, "/chat/completions", relay.ParseUsage)
}

// Messages routes a non-streaming Anthropic Messages request upstream.
func (s *relayService) Messages(ctx context.Context, input RelayInput) (*RelayResult, error) {
	return s.relayNonStream(ctx, input, platform.TypeAnthropic, "/v1/messages", parseAnthropicUsage)
}

// MessagesStream forwards a streaming Anthropic Messages request upstream.
func (s *relayService) MessagesStream(ctx context.Context, input RelayInput, w http.ResponseWriter) error {
	return s.relayStream(ctx, input, w, platform.TypeAnthropic, "/v1/messages", relay.ParseAnthropicUsage)
}

// ListModels returns the models available to the given group.
func (s *relayService) ListModels(ctx context.Context, groupID uuid.UUID) ([]ModelInfo, error) {
	platformIDs, err := s.groupPlatformRepo.ListPlatformsByGroupID(ctx, groupID)
	if err != nil {
		return nil, domain.WrapInternal(err)
	}

	hasActive := false
	for _, pid := range platformIDs {
		p, err := s.platformRepo.GetByID(ctx, pid)
		if err != nil {
			continue
		}
		if p.Status == platform.StatusActive {
			hasActive = true
			break
		}
	}

	if !hasActive {
		return []ModelInfo{}, nil
	}

	models, err := s.modelRepo.List(ctx)
	if err != nil {
		return nil, domain.WrapInternal(err)
	}

	now := time.Now().Unix()
	result := make([]ModelInfo, 0, len(models))
	for _, m := range models {
		if !m.IsEnabled {
			continue
		}
		result = append(result, ModelInfo{
			ID:      m.Name,
			Object:  "model",
			Created: now,
			OwnedBy: "school-api",
		})
	}

	return result, nil
}

// relayNonStream handles the common non-streaming relay path for OpenAI and Anthropic.
func (s *relayService) relayNonStream(
	ctx context.Context,
	input RelayInput,
	platformType platform.Type,
	upstreamPath string,
	usageParser func([]byte) RelayUsage,
) (*RelayResult, error) {
	start := time.Now()

	modelName, model, rewrittenBody, err := s.prepareChatRequest(ctx, input)
	if err != nil {
		return nil, err
	}

	candidates, err := s.listCandidatesByType(ctx, input.GroupID, platformType)
	if err != nil {
		return nil, err
	}
	if len(candidates) == 0 {
		return nil, domain.NewAppError(503, "service_unavailable", "No available account")
	}

	hold, err := s.billingService.PreDeduct(ctx, input.UserID, input.TokenID, model, input.RequestBody)
	if err != nil {
		return nil, err
	}

	attempts := len(candidates)
	if s.maxRetries+1 < attempts {
		attempts = s.maxRetries + 1
	}

	var successResult *RelayResult
	var successCallLogID uuid.UUID
	var finalResult *RelayResult
	var finalCallLogID uuid.UUID
	var lastErr error

	for i := 0; i < attempts && len(candidates) > 0; i++ {
		c, remaining, pickErr := s.pickWeightedCandidate(candidates)
		if pickErr != nil {
			_ = s.billingService.Refund(ctx, hold, nil)
			return nil, pickErr
		}
		candidates = remaining

		result, callLogID, retryable, attemptErr := s.attemptNonStream(ctx, c, input, modelName, rewrittenBody, upstreamPath, usageParser, start)
		if attemptErr != nil {
			if !retryable {
				_ = s.billingService.Refund(ctx, hold, logIDPtr(callLogID))
				return nil, attemptErr
			}
			s.markAccountError(ctx, c.account.ID)
			if result != nil {
				finalResult = result
				finalCallLogID = callLogID
			} else {
				lastErr = attemptErr
			}
			continue
		}

		if result.StatusCode >= 200 && result.StatusCode < 300 {
			successResult = result
			successCallLogID = callLogID
			break
		}
		if !retryable {
			_ = s.billingService.Refund(ctx, hold, logIDPtr(callLogID))
			return result, nil
		}

		s.markAccountError(ctx, c.account.ID)
		finalResult = result
		finalCallLogID = callLogID
	}

	if successResult != nil {
		if settleErr := s.billingService.Settle(ctx, hold, model, successResult.Usage, successCallLogID); settleErr != nil {
			// Settlement failures must not break the caller's response.
		}
		return successResult, nil
	}

	if finalResult != nil {
		_ = s.billingService.Refund(ctx, hold, logIDPtr(finalCallLogID))
		return finalResult, nil
	}

	_ = s.billingService.Refund(ctx, hold, nil)
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, domain.NewAppError(503, "service_unavailable", "No available account")
}

// attemptNonStream performs one upstream attempt and reports whether the failure is retryable.
func (s *relayService) attemptNonStream(
	ctx context.Context,
	c *candidate,
	input RelayInput,
	modelName string,
	rewrittenBody []byte,
	upstreamPath string,
	usageParser func([]byte) RelayUsage,
	start time.Time,
) (*RelayResult, uuid.UUID, bool, error) {
	apiKey, err := s.accountService.DecryptAPIKey(ctx, c.account.ID)
	if err != nil {
		return nil, uuid.UUID{}, true, domain.NewAppError(503, "service_unavailable", "Failed to prepare upstream credentials")
	}

	req, err := s.buildUpstreamRequest(ctx, c.platform, rewrittenBody, apiKey, "application/json", upstreamPath)
	if err != nil {
		return nil, uuid.UUID{}, false, err
	}

	resp, err := s.forwarder.Forward(ctx, req)
	if err != nil {
		return nil, uuid.UUID{}, true, domain.NewAppError(504, "gateway_timeout", "Upstream request failed")
	}

	limitedReader := io.LimitReader(resp.Body, s.maxBodyBytes+1)
	respBody, err := io.ReadAll(limitedReader)
	resp.Body.Close()
	if err != nil {
		return nil, uuid.UUID{}, true, domain.WrapInternal(err)
	}
	if int64(len(respBody)) > s.maxBodyBytes {
		return nil, uuid.UUID{}, false, domain.NewAppError(502, "bad_gateway", "Upstream response body too large")
	}

	usage := usageParser(respBody)

	var errMsg *string
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg := truncateString(string(respBody), 500)
		errMsg = &msg
	}

	callLogID := uuid.UUID{}
	if log, logErr := s.callLogRepo.Create(ctx, repository.CreateCallLogInput{
		UserID:           input.UserID,
		TokenID:          input.TokenID,
		PlatformID:       c.platform.ID,
		AccountID:        c.account.ID,
		Model:            modelName,
		PromptTokens:     usage.PromptTokens,
		CompletionTokens: usage.CompletionTokens,
		TotalTokens:      usage.TotalTokens,
		LatencyMs:        time.Since(start).Milliseconds(),
		StatusCode:       resp.StatusCode,
		ErrorMsg:         errMsg,
	}); logErr == nil {
		callLogID = log.ID
	}

	result := &RelayResult{
		StatusCode: resp.StatusCode,
		Header:     resp.Header.Clone(),
		Body:       respBody,
		Usage:      usage,
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return result, callLogID, false, nil
	}

	return result, callLogID, isRetryableStatus(resp.StatusCode), nil
}

// relayStream handles the common streaming relay path for OpenAI and Anthropic.
func (s *relayService) relayStream(
	ctx context.Context,
	input RelayInput,
	w http.ResponseWriter,
	platformType platform.Type,
	upstreamPath string,
	usageParser func([]string) relay.Usage,
) error {
	start := time.Now()

	modelName, model, rewrittenBody, err := s.prepareChatRequest(ctx, input)
	if err != nil {
		return err
	}

	candidates, err := s.listCandidatesByType(ctx, input.GroupID, platformType)
	if err != nil {
		return err
	}
	if len(candidates) == 0 {
		return domain.NewAppError(503, "service_unavailable", "No available account")
	}

	hold, err := s.billingService.PreDeduct(ctx, input.UserID, input.TokenID, model, input.RequestBody)
	if err != nil {
		return err
	}

	attempts := len(candidates)
	if s.maxRetries+1 < attempts {
		attempts = s.maxRetries + 1
	}

	var lastRespBody []byte
	var lastRespHeader http.Header
	var lastRespStatus int
	var lastErr error

	for i := 0; i < attempts && len(candidates) > 0; i++ {
		c, remaining, pickErr := s.pickWeightedCandidate(candidates)
		if pickErr != nil {
			_ = s.billingService.Refund(ctx, hold, nil)
			return pickErr
		}
		candidates = remaining

		apiKey, err := s.accountService.DecryptAPIKey(ctx, c.account.ID)
		if err != nil {
			s.markAccountError(ctx, c.account.ID)
			lastErr = domain.NewAppError(503, "service_unavailable", "Failed to prepare upstream credentials")
			continue
		}

		req, err := s.buildUpstreamRequest(ctx, c.platform, rewrittenBody, apiKey, "text/event-stream", upstreamPath)
		if err != nil {
			_ = s.billingService.Refund(ctx, hold, nil)
			return err
		}

		resp, err := s.forwarder.Forward(ctx, req)
		if err != nil {
			s.markAccountError(ctx, c.account.ID)
			lastErr = domain.NewAppError(504, "gateway_timeout", "Upstream request failed")
			continue
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			if isRetryableStatus(resp.StatusCode) {
				body, _ := io.ReadAll(io.LimitReader(resp.Body, s.maxBodyBytes+1))
				resp.Body.Close()
				s.markAccountError(ctx, c.account.ID)
				lastRespBody = body
				lastRespHeader = resp.Header.Clone()
				lastRespStatus = resp.StatusCode
				lastErr = nil
				continue
			}
			copyHeaders(w.Header(), resp.Header)
			w.WriteHeader(resp.StatusCode)
			_, _ = io.Copy(w, resp.Body)
			resp.Body.Close()
			_ = s.billingService.Refund(ctx, hold, nil)
			return nil
		}

		copyHeaders(w.Header(), resp.Header)
		w.WriteHeader(resp.StatusCode)

		var flusher http.Flusher
		if f, ok := w.(http.Flusher); ok {
			flusher = f
		}

		var captured []string
		onData := func(data string) {
			captured = append(captured, data)
		}

		streamErr := relay.StreamReader(w, resp.Body, flusher, onData)
		resp.Body.Close()
		if streamErr != nil {
			_ = s.billingService.Refund(ctx, hold, nil)
			return domain.WrapInternal(streamErr)
		}

		usage := usageParser(captured)

		callLogID := uuid.UUID{}
		if log, logErr := s.callLogRepo.Create(ctx, repository.CreateCallLogInput{
			UserID:           input.UserID,
			TokenID:          input.TokenID,
			PlatformID:       c.platform.ID,
			AccountID:        c.account.ID,
			Model:            modelName,
			PromptTokens:     usage.PromptTokens,
			CompletionTokens: usage.OutputTokens,
			TotalTokens:      usage.TotalTokens(),
			LatencyMs:        time.Since(start).Milliseconds(),
			StatusCode:       resp.StatusCode,
			ErrorMsg:         nil,
		}); logErr == nil {
			callLogID = log.ID
		}

		if settleErr := s.billingService.Settle(ctx, hold, model, RelayUsage{
			PromptTokens:     usage.PromptTokens,
			CompletionTokens: usage.OutputTokens,
			TotalTokens:      usage.TotalTokens(),
		}, callLogID); settleErr != nil {
			// Settlement failures must not break the caller's response.
		}

		return nil
	}

	if lastErr != nil {
		_ = s.billingService.Refund(ctx, hold, nil)
		return lastErr
	}
	if len(lastRespBody) > 0 || lastRespStatus != 0 {
		copyHeaders(w.Header(), lastRespHeader)
		w.WriteHeader(lastRespStatus)
		_, _ = w.Write(lastRespBody)
		_ = s.billingService.Refund(ctx, hold, nil)
		return nil
	}
	_ = s.billingService.Refund(ctx, hold, nil)
	return domain.NewAppError(503, "service_unavailable", "No available account")
}

// prepareChatRequest validates and rewrites the incoming request body.
// It returns the public model name, the model entity, the rewritten JSON body, and any error.
func (s *relayService) prepareChatRequest(ctx context.Context, input RelayInput) (string, *ent.AIModel, []byte, error) {
	var bodyMap map[string]json.RawMessage
	if err := json.Unmarshal(input.RequestBody, &bodyMap); err != nil {
		return "", nil, nil, domain.NewAppError(400, "invalid_request_error", "Invalid JSON body")
	}

	modelRaw, ok := bodyMap["model"]
	if !ok {
		return "", nil, nil, domain.NewAppError(400, "invalid_request_error", "Missing 'model' field")
	}

	var modelName string
	if err := json.Unmarshal(modelRaw, &modelName); err != nil {
		return "", nil, nil, domain.NewAppError(400, "invalid_request_error", "Invalid 'model' field")
	}

	model, err := s.modelRepo.GetByName(ctx, modelName)
	if err != nil {
		if ent.IsNotFound(err) {
			return "", nil, nil, domain.NewAppError(400, "model_not_found", "The model '"+modelName+"' does not exist")
		}
		return "", nil, nil, domain.WrapInternal(err)
	}
	if !model.IsEnabled {
		return "", nil, nil, domain.NewAppError(400, "model_not_found", "The model '"+modelName+"' is not available")
	}

	bodyMap["model"] = json.RawMessage(fmt.Sprintf("%q", model.UpstreamName))
	rewrittenBody, err := json.Marshal(bodyMap)
	if err != nil {
		return "", nil, nil, domain.WrapInternal(err)
	}

	return modelName, model, rewrittenBody, nil
}

// candidate pairs an account with the platform it belongs to.
type candidate struct {
	account  *ent.Account
	platform *ent.Platform
}

// listCandidatesByType returns all active accounts of the requested platform type accessible by the group.
func (s *relayService) listCandidatesByType(ctx context.Context, groupID uuid.UUID, platformType platform.Type) ([]*candidate, error) {
	platformIDs, err := s.groupPlatformRepo.ListPlatformsByGroupID(ctx, groupID)
	if err != nil {
		return nil, domain.WrapInternal(err)
	}
	if len(platformIDs) == 0 {
		return nil, domain.NewAppError(503, "service_unavailable", "No platform configured for this group")
	}

	var candidates []*candidate

	for _, pid := range platformIDs {
		p, err := s.platformRepo.GetByID(ctx, pid)
		if err != nil {
			continue
		}
		if p.Status != platform.StatusActive || p.Type != platformType {
			continue
		}

		accounts, err := s.accountRepo.ListByPlatformID(ctx, p.ID)
		if err != nil {
			continue
		}
		for _, a := range accounts {
			if a.Status == account.StatusActive {
				candidates = append(candidates, &candidate{account: a, platform: p})
			}
		}
	}

	return candidates, nil
}

// pickWeightedCandidate selects one account by weighted random sampling and returns the remaining candidates.
func (s *relayService) pickWeightedCandidate(candidates []*candidate) (*candidate, []*candidate, error) {
	totalWeight := 0
	for _, c := range candidates {
		w := c.account.Weight
		if w < 1 {
			w = 1
		}
		totalWeight += w
	}

	r, err := randInt(big.NewInt(int64(totalWeight)))
	if err != nil {
		return nil, nil, domain.WrapInternal(err)
	}

	cursor := int64(0)
	pick := r.Int64()
	for i, c := range candidates {
		w := c.account.Weight
		if w < 1 {
			w = 1
		}
		cursor += int64(w)
		if pick < cursor {
			remaining := append([]*candidate{}, candidates[:i]...)
			remaining = append(remaining, candidates[i+1:]...)
			return c, remaining, nil
		}
	}

	// Fallback for any rounding issues.
	lastIdx := len(candidates) - 1
	return candidates[lastIdx], candidates[:lastIdx], nil
}

// markAccountError increments the account's consecutive failure counter.
// Failures here must not fail the caller's request.
func (s *relayService) markAccountError(ctx context.Context, accountID uuid.UUID) {
	_ = s.accountRepo.IncrementErrorCount(ctx, accountID)
}

// isRetryableStatus reports whether an upstream HTTP status warrants switching to another account.
func isRetryableStatus(code int) bool {
	return code >= 500 || code == http.StatusTooManyRequests
}

// logIDPtr returns a pointer to the given UUID when it is non-zero, otherwise nil.
func logIDPtr(id uuid.UUID) *uuid.UUID {
	if id == (uuid.UUID{}) {
		return nil
	}
	return &id
}

// buildUpstreamRequest constructs the HTTP request sent to the provider.
func (s *relayService) buildUpstreamRequest(ctx context.Context, p *ent.Platform, body []byte, apiKey, accept, path string) (*http.Request, error) {
	url := strings.TrimSuffix(p.BaseURL, "/") + path
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, domain.WrapInternal(err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", accept)
	// Anthropic requires an x-api-key header in addition to Authorization.
	if strings.Contains(p.BaseURL, "anthropic") || path == "/v1/messages" {
		req.Header.Set("x-api-key", apiKey)
	}
	return req, nil
}

// parseOpenAIUsage extracts token counts from an OpenAI-style usage object.
func parseOpenAIUsage(body []byte) RelayUsage {
	var response struct {
		Usage struct {
			PromptTokens     int64 `json:"prompt_tokens"`
			CompletionTokens int64 `json:"completion_tokens"`
			TotalTokens      int64 `json:"total_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return RelayUsage{}
	}
	return RelayUsage{
		PromptTokens:     response.Usage.PromptTokens,
		CompletionTokens: response.Usage.CompletionTokens,
		TotalTokens:      response.Usage.TotalTokens,
	}
}

// parseAnthropicUsage extracts token counts from an Anthropic-style usage object.
func parseAnthropicUsage(body []byte) RelayUsage {
	var response struct {
		Usage struct {
			InputTokens  int64 `json:"input_tokens"`
			OutputTokens int64 `json:"output_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return RelayUsage{}
	}
	return RelayUsage{
		PromptTokens:     response.Usage.InputTokens,
		CompletionTokens: response.Usage.OutputTokens,
		TotalTokens:      response.Usage.InputTokens + response.Usage.OutputTokens,
	}
}

func truncateString(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max]
}

// copyHeaders copies selected upstream headers to the downstream response.
func copyHeaders(dst, src http.Header) {
	for _, key := range []string{"Content-Type", "Cache-Control", "X-Request-ID"} {
		if v := src.Get(key); v != "" {
			dst.Set(key, v)
		}
	}
}
