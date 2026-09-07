package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/ent/account"
	"github.com/school-api/school-api-v1/ent/calllog"
	"github.com/school-api/school-api-v1/ent/platform"
	"github.com/school-api/school-api-v1/internal/testutil"
)

func TestEntCallLogRepository_Create(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	repo := NewEntCallLogRepository(client)

	u, _, _, _ := testutil.CreateTestAPIKey(t, ctx, client)

	p, err := client.Platform.Create().
		SetName("OpenAI").
		SetCode("openai").
		SetType(platform.TypeOpenai).
		SetBaseURL("https://api.openai.com/v1").
		SetStatus(platform.StatusActive).
		Save(ctx)
	if err != nil {
		t.Fatalf("create platform: %v", err)
	}

	a, err := client.Account.Create().
		SetPlatform(p).
		SetName("Primary").
		SetAPIKeyEncrypted("encrypted-key").
		SetWeight(1).
		SetMaxRpm(0).
		SetStatus(account.StatusActive).
		Save(ctx)
	if err != nil {
		t.Fatalf("create account: %v", err)
	}

	errMsg := "upstream timeout"
	log, err := repo.Create(ctx, CreateCallLogInput{
		UserID:           u.ID,
		TokenID:          uuid.New(),
		PlatformID:       p.ID,
		AccountID:        a.ID,
		Model:            "gpt-4o",
		PromptTokens:     10,
		CompletionTokens: 20,
		TotalTokens:      30,
		LatencyMs:        123,
		StatusCode:       200,
		ErrorMsg:         &errMsg,
	})
	if err != nil {
		t.Fatalf("create call log: %v", err)
	}

	if log.UserID != u.ID {
		t.Errorf("expected user_id %s, got %s", u.ID, log.UserID)
	}
	if log.TokenID == uuid.Nil {
		t.Error("expected non-nil token_id")
	}
	if log.PlatformID != p.ID {
		t.Errorf("expected platform_id %s, got %s", p.ID, log.PlatformID)
	}
	if log.AccountID != a.ID {
		t.Errorf("expected account_id %s, got %s", a.ID, log.AccountID)
	}
	if log.Model != "gpt-4o" {
		t.Errorf("expected model gpt-4o, got %s", log.Model)
	}
	if log.PromptTokens != 10 || log.CompletionTokens != 20 || log.TotalTokens != 30 {
		t.Errorf("unexpected token counts: %v", log)
	}
	if log.LatencyMs != 123 {
		t.Errorf("expected latency 123, got %d", log.LatencyMs)
	}
	if log.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", log.StatusCode)
	}
	if log.ErrorMsg == nil || *log.ErrorMsg != errMsg {
		t.Errorf("expected error_msg %q, got %v", errMsg, log.ErrorMsg)
	}

	// Verify immutability: created_at should be set.
	if log.CreatedAt.IsZero() {
		t.Error("expected created_at to be set")
	}
}

func TestEntCallLogRepository_CreateNoErrorMsg(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	repo := NewEntCallLogRepository(client)

	u, _, _, _ := testutil.CreateTestAPIKey(t, ctx, client)
	p := createTestPlatform(t, ctx, client)
	a := createTestAccount(t, ctx, client, p)

	log, err := repo.Create(ctx, CreateCallLogInput{
		UserID:     u.ID,
		TokenID:    uuid.New(),
		PlatformID: p.ID,
		AccountID:  a.ID,
		Model:      "gpt-4o",
		StatusCode: 200,
	})
	if err != nil {
		t.Fatalf("create call log: %v", err)
	}

	if log.ErrorMsg != nil {
		t.Errorf("expected nil error_msg, got %v", *log.ErrorMsg)
	}
}

func createTestPlatform(t *testing.T, ctx context.Context, client *ent.Client) *ent.Platform {
	t.Helper()
	p, err := client.Platform.Create().
		SetName("OpenAI").
		SetCode("openai").
		SetType(platform.TypeOpenai).
		SetBaseURL("https://api.openai.com/v1").
		SetStatus(platform.StatusActive).
		Save(ctx)
	if err != nil {
		t.Fatalf("create platform: %v", err)
	}
	return p
}

func createTestAccount(t *testing.T, ctx context.Context, client *ent.Client, p *ent.Platform) *ent.Account {
	t.Helper()
	a, err := client.Account.Create().
		SetPlatform(p).
		SetName("Primary").
		SetAPIKeyEncrypted("encrypted-key").
		SetWeight(1).
		SetMaxRpm(0).
		SetStatus(account.StatusActive).
		Save(ctx)
	if err != nil {
		t.Fatalf("create account: %v", err)
	}
	return a
}

func TestEntCallLogRepository_DeleteOlderThan(t *testing.T) {
	ctx := context.Background()
	client, db := openTestClientDB(t)
	repo := NewEntCallLogRepository(client)

	u, _, _, _ := testutil.CreateTestAPIKey(t, ctx, client)
	p := createTestPlatform(t, ctx, client)
	a := createTestAccount(t, ctx, client, p)

	mkLog := func(model string) *ent.CallLog {
		t.Helper()
		log, err := repo.Create(ctx, CreateCallLogInput{
			UserID:     u.ID,
			TokenID:    uuid.New(),
			PlatformID: p.ID,
			AccountID:  a.ID,
			Model:      model,
			StatusCode: 200,
		})
		if err != nil {
			t.Fatalf("create call log %s: %v", model, err)
		}
		return log
	}

	// created_at is immutable through Ent, so backdate rows with raw SQL. The
	// stored format is RFC3339 (T + Z), and the predicate compares as strings,
	// so write the same format to keep ordering correct.
	backdate := func(id uuid.UUID, at time.Time) {
		t.Helper()
		if _, err := db.ExecContext(ctx,
			"UPDATE call_logs SET created_at = ? WHERE id = ?",
			at.UTC().Format(time.RFC3339Nano), id.String(),
		); err != nil {
			t.Fatalf("backdate log: %v", err)
		}
	}

	old := mkLog("old-model")    // will be aged to 40 days ago
	recent := mkLog("new-model") // stays at now

	backdate(old.ID, time.Now().AddDate(0, 0, -40))

	cutoff := time.Now().AddDate(0, 0, -30)
	deleted, err := repo.DeleteOlderThan(ctx, cutoff)
	if err != nil {
		t.Fatalf("delete older than: %v", err)
	}
	if deleted != 1 {
		t.Fatalf("expected 1 row deleted, got %d", deleted)
	}

	remaining, err := client.CallLog.Query().All(ctx)
	if err != nil {
		t.Fatalf("query remaining: %v", err)
	}
	if len(remaining) != 1 || remaining[0].ID != recent.ID {
		t.Fatalf("expected only recent log to survive, got %d rows", len(remaining))
	}

	// Boundary: a log 1 second newer than the cutoff must survive the sweep.
	// Exact equality is fragile across timestamp precision, so use +1s.
	atCutoff := mkLog("at-cutoff")
	backdate(atCutoff.ID, cutoff.Add(time.Second))
	deleted, err = repo.DeleteOlderThan(ctx, cutoff)
	if err != nil {
		t.Fatalf("second delete: %v", err)
	}
	if deleted != 0 {
		t.Fatalf("expected 0 deleted for a log inside the window, got %d", deleted)
	}
	n, err := client.CallLog.Query().Where(calllog.ModelEQ("at-cutoff")).Count(ctx)
	if err != nil {
		t.Fatalf("count at-cutoff: %v", err)
	}
	if n != 1 {
		t.Fatalf("expected in-window log to survive, got count %d", n)
	}
}
