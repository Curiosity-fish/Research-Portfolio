package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/school-api/school-api-v1/internal/domain"
	"github.com/school-api/school-api-v1/internal/testutil"
)

func TestEntBillingRepository_PreDeductSuccess(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	defer client.Close()

	u, _, tok, _ := testutil.CreateTestAPIKey(t, ctx, client)
	limit := int64(1000)
	_, err := client.UserToken.UpdateOneID(tok.ID).
		SetQuotaLimit(limit).
		Save(ctx)
	if err != nil {
		t.Fatalf("set quota limit: %v", err)
	}
	_, err = client.User.UpdateOneID(u.ID).SetBalance(5000).Save(ctx)
	if err != nil {
		t.Fatalf("set balance: %v", err)
	}

	repo := NewEntBillingRepository(client)
	hold, err := repo.PreDeduct(ctx, PreDeductInput{
		UserID:          u.ID,
		TokenID:         tok.ID,
		EstimatedTokens: 100,
		EstimatedCost:   50,
	})
	if err != nil {
		t.Fatalf("pre-deduct: %v", err)
	}
	if hold.EstimatedTokens != 100 || hold.EstimatedCost != 50 {
		t.Fatalf("unexpected hold: %+v", hold)
	}

	freshUser, err := client.User.Get(ctx, u.ID)
	if err != nil {
		t.Fatalf("get user: %v", err)
	}
	if freshUser.Balance != 4950 {
		t.Fatalf("expected balance 4950, got %d", freshUser.Balance)
	}

	freshTok, err := client.UserToken.Get(ctx, tok.ID)
	if err != nil {
		t.Fatalf("get token: %v", err)
	}
	if freshTok.QuotaUsed != 100 {
		t.Fatalf("expected quota_used 100, got %d", freshTok.QuotaUsed)
	}

	records, err := client.QuotaRecord.Query().All(ctx)
	_ = records
	if err != nil {
		t.Fatalf("list quota records: %v", err)
	}
	cnt, err := client.QuotaRecord.Query().Count(ctx)
	if err != nil {
		t.Fatalf("count quota records: %v", err)
	}
	if cnt != 1 {
		t.Fatalf("expected 1 quota record, got %d", cnt)
	}
	bcnt, err := client.BalanceRecord.Query().Count(ctx)
	if err != nil {
		t.Fatalf("count balance records: %v", err)
	}
	if bcnt != 1 {
		t.Fatalf("expected 1 balance record, got %d", bcnt)
	}
}

func TestEntBillingRepository_PreDeductInsufficientQuota(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	defer client.Close()

	u, _, tok, _ := testutil.CreateTestAPIKey(t, ctx, client)
	limit := int64(50)
	_, err := client.UserToken.UpdateOneID(tok.ID).SetQuotaLimit(limit).Save(ctx)
	if err != nil {
		t.Fatalf("set quota limit: %v", err)
	}

	repo := NewEntBillingRepository(client)
	_, err = repo.PreDeduct(ctx, PreDeductInput{
		UserID:          u.ID,
		TokenID:         tok.ID,
		EstimatedTokens: 100,
		EstimatedCost:   0,
	})
	if !errors.Is(err, domain.ErrInsufficientQuota) {
		t.Fatalf("expected ErrInsufficientQuota, got %v", err)
	}

	freshTok, err := client.UserToken.Get(ctx, tok.ID)
	if err != nil {
		t.Fatalf("get token: %v", err)
	}
	if freshTok.QuotaUsed != 0 {
		t.Fatalf("expected quota_used unchanged, got %d", freshTok.QuotaUsed)
	}
}

func TestEntBillingRepository_PreDeductInsufficientBalance(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	defer client.Close()

	u, _, tok, _ := testutil.CreateTestAPIKey(t, ctx, client)
	_, err := client.UserToken.UpdateOneID(tok.ID).SetQuotaLimit(1000).Save(ctx)
	if err != nil {
		t.Fatalf("set quota limit: %v", err)
	}

	repo := NewEntBillingRepository(client)
	_, err = repo.PreDeduct(ctx, PreDeductInput{
		UserID:          u.ID,
		TokenID:         tok.ID,
		EstimatedTokens: 10,
		EstimatedCost:   100,
	})
	if !errors.Is(err, domain.ErrInsufficientBalance) {
		t.Fatalf("expected ErrInsufficientBalance, got %v", err)
	}
}

func TestEntBillingRepository_SettleRefund(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	defer client.Close()

	u, _, tok, _ := testutil.CreateTestAPIKey(t, ctx, client)
	_, err := client.UserToken.UpdateOneID(tok.ID).SetQuotaLimit(1000).Save(ctx)
	if err != nil {
		t.Fatalf("set quota limit: %v", err)
	}
	_, err = client.User.UpdateOneID(u.ID).SetBalance(5000).Save(ctx)
	if err != nil {
		t.Fatalf("set balance: %v", err)
	}

	repo := NewEntBillingRepository(client)
	hold, err := repo.PreDeduct(ctx, PreDeductInput{
		UserID:          u.ID,
		TokenID:         tok.ID,
		EstimatedTokens: 100,
		EstimatedCost:   50,
	})
	if err != nil {
		t.Fatalf("pre-deduct: %v", err)
	}

	if err := repo.Settle(ctx, hold, 40, 20, nil); err != nil {
		t.Fatalf("settle: %v", err)
	}

	freshUser, _ := client.User.Get(ctx, u.ID)
	if freshUser.Balance != 4980 {
		t.Fatalf("expected balance 4980, got %d", freshUser.Balance)
	}
	freshTok, _ := client.UserToken.Get(ctx, tok.ID)
	if freshTok.QuotaUsed != 40 {
		t.Fatalf("expected quota_used 40, got %d", freshTok.QuotaUsed)
	}

	qcnt, _ := client.QuotaRecord.Query().Count(ctx)
	bcnt, _ := client.BalanceRecord.Query().Count(ctx)
	if qcnt != 2 {
		t.Fatalf("expected 2 quota records, got %d", qcnt)
	}
	if bcnt != 2 {
		t.Fatalf("expected 2 balance records, got %d", bcnt)
	}
}

func TestEntBillingRepository_Refund(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	defer client.Close()

	u, _, tok, _ := testutil.CreateTestAPIKey(t, ctx, client)
	_, err := client.UserToken.UpdateOneID(tok.ID).SetQuotaLimit(1000).Save(ctx)
	if err != nil {
		t.Fatalf("set quota limit: %v", err)
	}
	_, err = client.User.UpdateOneID(u.ID).SetBalance(5000).Save(ctx)
	if err != nil {
		t.Fatalf("set balance: %v", err)
	}

	repo := NewEntBillingRepository(client)
	hold, err := repo.PreDeduct(ctx, PreDeductInput{
		UserID:          u.ID,
		TokenID:         tok.ID,
		EstimatedTokens: 100,
		EstimatedCost:   50,
	})
	if err != nil {
		t.Fatalf("pre-deduct: %v", err)
	}

	if err := repo.Refund(ctx, hold, nil); err != nil {
		t.Fatalf("refund: %v", err)
	}

	freshUser, _ := client.User.Get(ctx, u.ID)
	if freshUser.Balance != 5000 {
		t.Fatalf("expected balance 5000, got %d", freshUser.Balance)
	}
	freshTok, _ := client.UserToken.Get(ctx, tok.ID)
	if freshTok.QuotaUsed != 0 {
		t.Fatalf("expected quota_used 0, got %d", freshTok.QuotaUsed)
	}
}

func TestEntBillingRepository_UnlimitedQuota(t *testing.T) {
	ctx := context.Background()
	client := openTestClient(t)
	defer client.Close()

	u, _, tok, _ := testutil.CreateTestAPIKey(t, ctx, client)

	repo := NewEntBillingRepository(client)
	_, err := repo.PreDeduct(ctx, PreDeductInput{
		UserID:          u.ID,
		TokenID:         tok.ID,
		EstimatedTokens: 100000,
		EstimatedCost:   0,
	})
	if err != nil {
		t.Fatalf("pre-deduct with unlimited quota: %v", err)
	}
}
