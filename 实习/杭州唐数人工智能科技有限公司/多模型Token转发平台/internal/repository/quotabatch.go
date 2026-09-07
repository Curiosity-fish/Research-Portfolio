package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/ent/predicate"
	"github.com/school-api/school-api-v1/ent/quotarecord"
	"github.com/school-api/school-api-v1/ent/user"
	"github.com/school-api/school-api-v1/ent/usertoken"
)

// QuotaBatchMode selects how QuotaBatchInput.Value is applied to quota_limit.
type QuotaBatchMode string

const (
	// QuotaBatchModeSet overwrites quota_limit with Value.
	QuotaBatchModeSet QuotaBatchMode = "set"
	// QuotaBatchModeAdd adds the (possibly negative) Value to quota_limit.
	QuotaBatchModeAdd QuotaBatchMode = "add"
)

// QuotaBatchInput describes one batch quota operation. Exactly one of UserIDs
// and DepartmentID selects the target users; MaxQuota comes from the settings
// quota range and bounds add-mode results (the set-mode value is range-checked
// by the service before the repository is called).
type QuotaBatchInput struct {
	UserIDs      []uuid.UUID
	DepartmentID *uuid.UUID
	Mode         QuotaBatchMode
	Value        int64
	MaxQuota     int64
}

// QuotaBatchResult reports the per-target outcome of a batch operation. Users
// without an enabled token are skipped, as are tokens whose add-mode result
// would leave the allowed range.
type QuotaBatchResult struct {
	MatchedUsers  int `json:"matched_users"`
	UpdatedTokens int `json:"updated_tokens"`
	SkippedUsers  int `json:"skipped_users"`
	SkippedTokens int `json:"skipped_tokens"`
}

// QuotaBatchRepository applies batch quota changes with an audit trail.
type QuotaBatchRepository interface {
	Apply(ctx context.Context, input QuotaBatchInput) (*QuotaBatchResult, error)
}

// EntQuotaBatchRepository implements QuotaBatchRepository using Ent.
type EntQuotaBatchRepository struct {
	client *ent.Client
}

// NewEntQuotaBatchRepository creates a new Ent-backed quota batch repository.
func NewEntQuotaBatchRepository(client *ent.Client) *EntQuotaBatchRepository {
	return &EntQuotaBatchRepository{client: client}
}

// Apply updates quota_limit on every enabled token of the target users in a
// single transaction and appends an admin_adjust quota record per token.
//
// Semantics:
//   - only enabled tokens are touched; disabled tokens and users without any
//     enabled token are counted as skipped (user status itself is ignored, the
//     token-level switch is the meaningful gate)
//   - set mode assumes Value was already range-validated by the service and
//     overwrites quota_limit unconditionally, including on unlimited tokens
//   - add mode skips unlimited tokens (nil quota_limit) and tokens whose
//     resulting limit would be negative or above MaxQuota
//   - quota record amount stores the magnitude of the change (the column has
//     a CHECK amount > 0 constraint); quota_after follows the existing
//     convention of recording quota_used after the change
func (r *EntQuotaBatchRepository) Apply(ctx context.Context, input QuotaBatchInput) (*QuotaBatchResult, error) {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("start quota batch transaction: %w", err)
	}

	rollback := func(err error) (*QuotaBatchResult, error) {
		if rerr := tx.Rollback(); rerr != nil {
			err = fmt.Errorf("%v (rollback failed: %v)", err, rerr)
		}
		return nil, err
	}

	var userPred predicate.User
	if input.DepartmentID != nil {
		userPred = user.DepartmentID(*input.DepartmentID)
	} else {
		userPred = user.IDIn(input.UserIDs...)
	}
	// Classes are leaf departments whose users attach directly, so an exact
	// department match (no subtree expansion) is the intended targeting.
	userIDs, err := tx.User.Query().Where(userPred).IDs(ctx)
	if err != nil {
		return rollback(fmt.Errorf("resolve batch target users: %w", err))
	}

	result := &QuotaBatchResult{MatchedUsers: len(userIDs)}
	if len(userIDs) == 0 {
		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("commit empty quota batch: %w", err)
		}
		return result, nil
	}

	tokens, err := tx.UserToken.Query().
		Where(
			usertoken.HasUserWith(user.IDIn(userIDs...)),
			usertoken.IsEnabledEQ(true),
		).
		All(ctx)
	if err != nil {
		return rollback(fmt.Errorf("load batch target tokens: %w", err))
	}

	tokensByUser := make(map[uuid.UUID][]*ent.UserToken, len(userIDs))
	for _, tok := range tokens {
		tokensByUser[tok.UserID] = append(tokensByUser[tok.UserID], tok)
	}

	for _, uid := range userIDs {
		userTokens := tokensByUser[uid]
		if len(userTokens) == 0 {
			result.SkippedUsers++
			continue
		}
		for _, tok := range userTokens {
			newLimit, ok := batchNewLimit(tok, input)
			if !ok {
				result.SkippedTokens++
				continue
			}
			amount, changed := batchAdjustAmount(tok, newLimit)
			// A set-mode no-op (token already at the target limit) needs no
			// audit row: the CHECK amount > 0 constraint forbids recording a
			// zero change, and inventing a magnitude would falsify the trail.
			if changed {
				if _, err := tx.QuotaRecord.Create().
					SetUserID(uid).
					SetTokenID(tok.ID).
					SetType(quotarecord.TypeAdminAdjust).
					SetAmount(amount).
					SetQuotaAfter(tok.QuotaUsed).
					Save(ctx); err != nil {
					return rollback(fmt.Errorf("create batch quota record: %w", err))
				}
			}
			if _, err := tx.UserToken.UpdateOneID(tok.ID).
				SetQuotaLimit(newLimit).
				Save(ctx); err != nil {
				return rollback(fmt.Errorf("update token quota: %w", err))
			}
			result.UpdatedTokens++
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit quota batch: %w", err)
	}
	return result, nil
}

// batchNewLimit computes the new quota_limit for one token, reporting whether
// the token is adjustable at all.
func batchNewLimit(tok *ent.UserToken, input QuotaBatchInput) (int64, bool) {
	switch input.Mode {
	case QuotaBatchModeSet:
		return input.Value, true
	case QuotaBatchModeAdd:
		if tok.QuotaLimit == nil {
			// Unlimited tokens cannot absorb a delta.
			return 0, false
		}
		newLimit := *tok.QuotaLimit + input.Value
		if newLimit < 0 || newLimit > input.MaxQuota {
			return 0, false
		}
		return newLimit, true
	}
	return 0, false
}

// batchAdjustAmount returns the positive magnitude written to the quota
// record: the size of the change for finite tokens, or the new limit itself
// when a previously unlimited token is given one. changed is false for a
// no-op (unchanged limit), in which case no record row may be written.
func batchAdjustAmount(tok *ent.UserToken, newLimit int64) (int64, bool) {
	if tok.QuotaLimit == nil {
		return newLimit, newLimit > 0
	}
	delta := newLimit - *tok.QuotaLimit
	if delta == 0 {
		return 0, false
	}
	if delta < 0 {
		delta = -delta
	}
	return delta, true
}

var _ QuotaBatchRepository = (*EntQuotaBatchRepository)(nil)
