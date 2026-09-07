package repository

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent/quotarecord"
	"github.com/school-api/school-api-v1/internal/auth"
	"github.com/school-api/school-api-v1/internal/testutil"
)

func TestEntQuotaBatchRepository_SetByUserIDs(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	repo := NewEntQuotaBatchRepository(client)

	// Target user: one enabled token with a limit, one disabled token.
	user1, _, enabledTok, _ := testutil.CreateTestAPIKey(t, ctx, client)
	limit := int64(1_000_000)
	if _, err := client.UserToken.UpdateOneID(enabledTok.ID).SetQuotaLimit(limit).Save(ctx); err != nil {
		t.Fatalf("seed token limit: %v", err)
	}
	plaintext, hash, err := auth.GenerateAPIKey()
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	if _, err := client.UserToken.Create().
		SetUserID(user1.ID).
		SetName("disabled").
		SetTokenHash(hash).
		SetTokenLast4(plaintext[len(plaintext)-4:]).
		SetIsEnabled(false).
		Save(ctx); err != nil {
		t.Fatalf("seed disabled token: %v", err)
	}

	// User whose only token is deleted, plus a user ID that matches nobody.
	user2, _, user2Tok, _ := testutil.CreateTestAPIKey(t, ctx, client)
	if err := client.UserToken.DeleteOneID(user2Tok.ID).Exec(ctx); err != nil {
		t.Fatalf("delete user2 token: %v", err)
	}

	result, err := repo.Apply(ctx, QuotaBatchInput{
		UserIDs:  []uuid.UUID{user1.ID, user2.ID, uuid.New()},
		Mode:     QuotaBatchModeSet,
		Value:    2_000_000,
		MaxQuota: 3_000_000,
	})
	if err != nil {
		t.Fatalf("apply batch: %v", err)
	}
	// Only existing users count as matched; the unknown ID matches nobody.
	if result.MatchedUsers != 2 || result.UpdatedTokens != 1 ||
		result.SkippedUsers != 1 || result.SkippedTokens != 0 {
		t.Fatalf("unexpected result: %+v", result)
	}

	tok, err := client.UserToken.Get(ctx, enabledTok.ID)
	if err != nil {
		t.Fatalf("reload token: %v", err)
	}
	if tok.QuotaLimit == nil || *tok.QuotaLimit != 2_000_000 {
		t.Fatalf("expected limit 2000000, got %v", tok.QuotaLimit)
	}

	recs, err := client.QuotaRecord.Query().
		Where(quotarecord.TokenIDEQ(enabledTok.ID)).
		All(ctx)
	if err != nil {
		t.Fatalf("load records: %v", err)
	}
	if len(recs) != 1 {
		t.Fatalf("expected 1 quota record, got %d", len(recs))
	}
	if recs[0].Type != quotarecord.TypeAdminAdjust || recs[0].Amount != 1_000_000 {
		t.Fatalf("unexpected record: type=%s amount=%d", recs[0].Type, recs[0].Amount)
	}
}

func TestEntQuotaBatchRepository_SetNoOpWritesNoRecord(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	repo := NewEntQuotaBatchRepository(client)

	user, _, tok, _ := testutil.CreateTestAPIKey(t, ctx, client)
	limit := int64(1_500_000)
	if _, err := client.UserToken.UpdateOneID(tok.ID).SetQuotaLimit(limit).Save(ctx); err != nil {
		t.Fatalf("seed token limit: %v", err)
	}

	result, err := repo.Apply(ctx, QuotaBatchInput{
		UserIDs:  []uuid.UUID{user.ID},
		Mode:     QuotaBatchModeSet,
		Value:    limit,
		MaxQuota: 3_000_000,
	})
	if err != nil {
		t.Fatalf("apply batch: %v", err)
	}
	if result.UpdatedTokens != 1 || result.SkippedTokens != 0 {
		t.Fatalf("expected no-op counted as updated: %+v", result)
	}

	count, err := client.QuotaRecord.Query().Where(quotarecord.TokenIDEQ(tok.ID)).Count(ctx)
	if err != nil {
		t.Fatalf("count records: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected no record for no-op, got %d", count)
	}
}

func TestEntQuotaBatchRepository_AddModeBounds(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	repo := NewEntQuotaBatchRepository(client)

	// Finite-limit token that stays in range, and an unlimited token.
	user, _, tok, _ := testutil.CreateTestAPIKey(t, ctx, client)
	limit := int64(1_000_000)
	if _, err := client.UserToken.UpdateOneID(tok.ID).SetQuotaLimit(limit).Save(ctx); err != nil {
		t.Fatalf("seed token limit: %v", err)
	}
	plaintext, hash, err := auth.GenerateAPIKey()
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	if _, err := client.UserToken.Create().
		SetUserID(user.ID).
		SetName("unlimited").
		SetTokenHash(hash).
		SetTokenLast4(plaintext[len(plaintext)-4:]).
		Save(ctx); err != nil {
		t.Fatalf("seed unlimited token: %v", err)
	}

	// +2M lands the finite token exactly on MaxQuota (allowed); the
	// unlimited token is skipped.
	result, err := repo.Apply(ctx, QuotaBatchInput{
		UserIDs:  []uuid.UUID{user.ID},
		Mode:     QuotaBatchModeAdd,
		Value:    2_000_000,
		MaxQuota: 3_000_000,
	})
	if err != nil {
		t.Fatalf("apply batch: %v", err)
	}
	if result.UpdatedTokens != 1 || result.SkippedTokens != 1 {
		t.Fatalf("unexpected result: %+v", result)
	}

	tok, err = client.UserToken.Get(ctx, tok.ID)
	if err != nil {
		t.Fatalf("reload token: %v", err)
	}
	if tok.QuotaLimit == nil || *tok.QuotaLimit != 3_000_000 {
		t.Fatalf("expected limit 3000000, got %v", tok.QuotaLimit)
	}

	// A further +1 would exceed MaxQuota, so the token is now skipped.
	result, err = repo.Apply(ctx, QuotaBatchInput{
		UserIDs:  []uuid.UUID{user.ID},
		Mode:     QuotaBatchModeAdd,
		Value:    1,
		MaxQuota: 3_000_000,
	})
	if err != nil {
		t.Fatalf("apply overflow batch: %v", err)
	}
	if result.UpdatedTokens != 0 || result.SkippedTokens != 2 {
		t.Fatalf("expected overflow skipped: %+v", result)
	}

	// Negative delta below zero is also skipped.
	result, err = repo.Apply(ctx, QuotaBatchInput{
		UserIDs:  []uuid.UUID{user.ID},
		Mode:     QuotaBatchModeAdd,
		Value:    -4_000_000,
		MaxQuota: 3_000_000,
	})
	if err != nil {
		t.Fatalf("apply negative batch: %v", err)
	}
	if result.UpdatedTokens != 0 || result.SkippedTokens != 2 {
		t.Fatalf("expected negative overflow skipped: %+v", result)
	}

	// A valid reduction updates the limit with a positive record amount.
	result, err = repo.Apply(ctx, QuotaBatchInput{
		UserIDs:  []uuid.UUID{user.ID},
		Mode:     QuotaBatchModeAdd,
		Value:    -500_000,
		MaxQuota: 3_000_000,
	})
	if err != nil {
		t.Fatalf("apply reduction batch: %v", err)
	}
	if result.UpdatedTokens != 1 {
		t.Fatalf("expected reduction applied: %+v", result)
	}
	tok, err = client.UserToken.Get(ctx, tok.ID)
	if err != nil {
		t.Fatalf("reload token after reduction: %v", err)
	}
	if tok.QuotaLimit == nil || *tok.QuotaLimit != 2_500_000 {
		t.Fatalf("expected limit 2500000 after reduction, got %v", tok.QuotaLimit)
	}
	recs, err := client.QuotaRecord.Query().Where(quotarecord.TokenIDEQ(tok.ID)).All(ctx)
	if err != nil {
		t.Fatalf("load records: %v", err)
	}
	if len(recs) != 2 {
		t.Fatalf("expected 2 records (increase + reduction), got %d", len(recs))
	}
	for _, rec := range recs {
		if rec.Amount <= 0 {
			t.Fatalf("record amount must be positive, got %d", rec.Amount)
		}
	}
}

func TestEntQuotaBatchRepository_ByDepartment(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	repo := NewEntQuotaBatchRepository(client)

	dep, err := client.Department.Create().
		SetName("计算机1班").
		SetCode("cls-" + uuid.NewString()[:8]).
		Save(ctx)
	if err != nil {
		t.Fatalf("create department: %v", err)
	}
	other, err := client.Department.Create().
		SetName("计算机2班").
		SetCode("cls-" + uuid.NewString()[:8]).
		Save(ctx)
	if err != nil {
		t.Fatalf("create other department: %v", err)
	}

	inClass, _, inTok, _ := testutil.CreateTestAPIKey(t, ctx, client)
	if _, err := client.User.UpdateOneID(inClass.ID).SetDepartmentID(dep.ID).Save(ctx); err != nil {
		t.Fatalf("assign department: %v", err)
	}
	outClass, _, outTok, _ := testutil.CreateTestAPIKey(t, ctx, client)
	if _, err := client.User.UpdateOneID(outClass.ID).SetDepartmentID(other.ID).Save(ctx); err != nil {
		t.Fatalf("assign other department: %v", err)
	}

	result, err := repo.Apply(ctx, QuotaBatchInput{
		DepartmentID: &dep.ID,
		Mode:         QuotaBatchModeSet,
		Value:        1_000_000,
		MaxQuota:     3_000_000,
	})
	if err != nil {
		t.Fatalf("apply batch by department: %v", err)
	}
	if result.MatchedUsers != 1 || result.UpdatedTokens != 1 {
		t.Fatalf("unexpected result: %+v", result)
	}

	got, err := client.UserToken.Get(ctx, inTok.ID)
	if err != nil {
		t.Fatalf("reload in-class token: %v", err)
	}
	if got.QuotaLimit == nil || *got.QuotaLimit != 1_000_000 {
		t.Fatalf("expected in-class limit 1000000, got %v", got.QuotaLimit)
	}
	untouched, err := client.UserToken.Get(ctx, outTok.ID)
	if err != nil {
		t.Fatalf("reload out-class token: %v", err)
	}
	if untouched.QuotaLimit != nil {
		t.Fatalf("out-class token must stay unlimited, got %v", *untouched.QuotaLimit)
	}
}
