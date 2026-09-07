package repository

import (
	"context"
	"fmt"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"

	"github.com/school-api/school-api-v1/ent"
	"github.com/school-api/school-api-v1/ent/balancerecord"
	"github.com/school-api/school-api-v1/ent/predicate"
	"github.com/school-api/school-api-v1/ent/quotarecord"
	"github.com/school-api/school-api-v1/ent/user"
	"github.com/school-api/school-api-v1/ent/usertoken"
	"github.com/school-api/school-api-v1/internal/domain"
)

// BillingHold captures the state of a pre-deduct so that a later settle/refund
// can restore the correct balance/quota.
type BillingHold struct {
	UserID          uuid.UUID
	TokenID         uuid.UUID
	EstimatedTokens int64
	EstimatedCost   int64
	QuotaRecordID   uuid.UUID
	BalanceRecordID *uuid.UUID // nil when estimated cost is zero
}

// PreDeductInput is the data required to reserve quota and balance before a relay.
type PreDeductInput struct {
	UserID          uuid.UUID
	TokenID         uuid.UUID
	EstimatedTokens int64
	EstimatedCost   int64
}

// BillingRepository handles cross-table transactions for quota and balance.
type BillingRepository interface {
	PreDeduct(ctx context.Context, input PreDeductInput) (*BillingHold, error)
	Settle(ctx context.Context, hold *BillingHold, actualTokens, actualCost int64, callLogID *uuid.UUID) error
	Refund(ctx context.Context, hold *BillingHold, callLogID *uuid.UUID) error
	AdminAdjustBalance(ctx context.Context, userID, adminID uuid.UUID, amount int64, remark *string) (*ent.User, error)
}

// EntBillingRepository implements BillingRepository using Ent transactions.
type EntBillingRepository struct {
	client *ent.Client
}

// NewEntBillingRepository creates a new Ent-backed billing repository.
func NewEntBillingRepository(client *ent.Client) *EntBillingRepository {
	return &EntBillingRepository{client: client}
}

// PreDeduct reserves estimated quota and balance in a single transaction.
// It returns a hold that must be passed to Settle or Refund after the relay.
func (r *EntBillingRepository) PreDeduct(ctx context.Context, input PreDeductInput) (*BillingHold, error) {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("start billing transaction: %w", err)
	}

	rollback := func(err error) (*BillingHold, error) {
		if rerr := tx.Rollback(); rerr != nil {
			err = fmt.Errorf("%v (rollback failed: %v)", err, rerr)
		}
		return nil, err
	}

	// Verify the user exists and has enough balance.
	u, err := tx.User.Get(ctx, input.UserID)
	if err != nil {
		if ent.IsNotFound(err) {
			return rollback(fmt.Errorf("billing user not found"))
		}
		return rollback(fmt.Errorf("get billing user: %w", err))
	}
	if input.EstimatedCost > 0 {
		if u.Balance < input.EstimatedCost {
			return rollback(domain.ErrInsufficientBalance)
		}
		balanceAffected, err := tx.User.Update().
			Where(user.IDEQ(input.UserID), user.BalanceGTE(input.EstimatedCost)).
			AddBalance(-input.EstimatedCost).
			Save(ctx)
		if err != nil {
			return rollback(fmt.Errorf("deduct balance: %w", err))
		}
		if balanceAffected == 0 {
			return rollback(domain.ErrInsufficientBalance)
		}
	}

	// Verify the token exists and has enough quota.
	tok, err := tx.UserToken.Get(ctx, input.TokenID)
	if err != nil {
		if ent.IsNotFound(err) {
			return rollback(domain.ErrInsufficientQuota)
		}
		return rollback(fmt.Errorf("get billing token: %w", err))
	}
	if tok.QuotaLimit != nil && tok.QuotaUsed+input.EstimatedTokens > *tok.QuotaLimit {
		return rollback(domain.ErrInsufficientQuota)
	}
	quotaAffected, err := tx.UserToken.Update().
		Where(usertoken.IDEQ(input.TokenID), quotaAvailable(input.EstimatedTokens)).
		AddQuotaUsed(input.EstimatedTokens).
		Save(ctx)
	if err != nil {
		return rollback(fmt.Errorf("deduct quota: %w", err))
	}
	if quotaAffected == 0 {
		return rollback(domain.ErrInsufficientQuota)
	}

	// Re-read the mutated rows to capture authoritative after-values.
	u, err = tx.User.Get(ctx, input.UserID)
	if err != nil {
		return rollback(fmt.Errorf("get updated user: %w", err))
	}
	tok, err = tx.UserToken.Get(ctx, input.TokenID)
	if err != nil {
		return rollback(fmt.Errorf("get updated token: %w", err))
	}

	quotaRec, err := tx.QuotaRecord.Create().
		SetUserID(input.UserID).
		SetTokenID(input.TokenID).
		SetType(quotarecord.TypePreDeduct).
		SetAmount(input.EstimatedTokens).
		SetQuotaAfter(tok.QuotaUsed).
		Save(ctx)
	if err != nil {
		return rollback(fmt.Errorf("create pre-deduct quota record: %w", err))
	}

	hold := &BillingHold{
		UserID:          input.UserID,
		TokenID:         input.TokenID,
		EstimatedTokens: input.EstimatedTokens,
		EstimatedCost:   input.EstimatedCost,
		QuotaRecordID:   quotaRec.ID,
	}

	if input.EstimatedCost > 0 {
		balanceRec, err := tx.BalanceRecord.Create().
			SetUserID(input.UserID).
			SetType(balancerecord.TypeConsume).
			SetAmount(input.EstimatedCost).
			SetBalanceAfter(u.Balance).
			Save(ctx)
		if err != nil {
			return rollback(fmt.Errorf("create pre-deduct balance record: %w", err))
		}
		hold.BalanceRecordID = &balanceRec.ID
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit billing pre-deduct: %w", err)
	}
	return hold, nil
}

// Settle adjusts the pre-deducted reservation to match actual usage.
func (r *EntBillingRepository) Settle(ctx context.Context, hold *BillingHold, actualTokens, actualCost int64, callLogID *uuid.UUID) error {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return fmt.Errorf("start settle transaction: %w", err)
	}

	rollback := func(err error) error {
		if rerr := tx.Rollback(); rerr != nil {
			err = fmt.Errorf("%v (rollback failed: %v)", err, rerr)
		}
		return err
	}

	tokenDelta := actualTokens - hold.EstimatedTokens
	costDelta := actualCost - hold.EstimatedCost

	if tokenDelta != 0 {
		if err := r.adjustQuotaInTx(ctx, tx, hold.UserID, hold.TokenID, tokenDelta, callLogID); err != nil {
			return rollback(err)
		}
	}
	if costDelta != 0 {
		if err := r.adjustBalanceInTx(ctx, tx, hold.UserID, -costDelta, callLogID, nil); err != nil {
			return rollback(err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit settle: %w", err)
	}
	return nil
}

// Refund releases a pre-deducted reservation when the relay did not succeed.
func (r *EntBillingRepository) Refund(ctx context.Context, hold *BillingHold, callLogID *uuid.UUID) error {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return fmt.Errorf("start refund transaction: %w", err)
	}

	rollback := func(err error) error {
		if rerr := tx.Rollback(); rerr != nil {
			err = fmt.Errorf("%v (rollback failed: %v)", err, rerr)
		}
		return err
	}

	if err := r.adjustQuotaInTx(ctx, tx, hold.UserID, hold.TokenID, -hold.EstimatedTokens, callLogID); err != nil {
		return rollback(err)
	}
	if hold.EstimatedCost > 0 {
		if err := r.adjustBalanceInTx(ctx, tx, hold.UserID, hold.EstimatedCost, callLogID, nil); err != nil {
			return rollback(err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit refund: %w", err)
	}
	return nil
}

// adjustQuotaInTx applies a signed delta to a token's quota_used and writes a record.
// A negative delta creates a refund record; a positive delta creates a consume record.
func (r *EntBillingRepository) adjustQuotaInTx(ctx context.Context, tx *ent.Tx, userID, tokenID uuid.UUID, delta int64, callLogID *uuid.UUID) error {
	if delta == 0 {
		return nil
	}

	if _, err := tx.UserToken.UpdateOneID(tokenID).AddQuotaUsed(delta).Save(ctx); err != nil {
		return fmt.Errorf("adjust quota_used: %w", err)
	}

	tok, err := tx.UserToken.Get(ctx, tokenID)
	if err != nil {
		return fmt.Errorf("get token after quota adjustment: %w", err)
	}

	recType := quotarecord.TypeConsume
	amount := delta
	if delta < 0 {
		recType = quotarecord.TypeRefund
		amount = -delta
	}

	b := tx.QuotaRecord.Create().
		SetUserID(userID).
		SetTokenID(tokenID).
		SetType(recType).
		SetAmount(amount).
		SetQuotaAfter(tok.QuotaUsed)
	if callLogID != nil {
		b.SetCallLogID(*callLogID)
	}
	if _, err := b.Save(ctx); err != nil {
		return fmt.Errorf("create quota adjustment record: %w", err)
	}
	return nil
}

// adjustBalanceInTx applies a signed delta to a user's balance and writes a record.
// A negative delta creates a consume record; a positive delta creates a refund record.
func (r *EntBillingRepository) adjustBalanceInTx(ctx context.Context, tx *ent.Tx, userID uuid.UUID, delta int64, callLogID, orderID *uuid.UUID) error {
	if delta == 0 {
		return nil
	}

	if _, err := tx.User.UpdateOneID(userID).AddBalance(delta).Save(ctx); err != nil {
		return fmt.Errorf("adjust balance: %w", err)
	}

	u, err := tx.User.Get(ctx, userID)
	if err != nil {
		return fmt.Errorf("get user after balance adjustment: %w", err)
	}

	recType := balancerecord.TypeRefund
	amount := delta
	if delta < 0 {
		recType = balancerecord.TypeConsume
		amount = -delta
	}

	b := tx.BalanceRecord.Create().
		SetUserID(userID).
		SetType(recType).
		SetAmount(amount).
		SetBalanceAfter(u.Balance)
	if callLogID != nil {
		b.SetCallLogID(*callLogID)
	}
	if orderID != nil {
		b.SetRelatedOrderID(*orderID)
	}
	if _, err := b.Save(ctx); err != nil {
		return fmt.Errorf("create balance adjustment record: %w", err)
	}
	return nil
}

// AdminAdjustBalance applies a signed balance adjustment by an admin and writes a balance record.
func (r *EntBillingRepository) AdminAdjustBalance(ctx context.Context, userID, adminID uuid.UUID, amount int64, remark *string) (*ent.User, error) {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("start admin adjust transaction: %w", err)
	}

	rollback := func(err error) (*ent.User, error) {
		if rerr := tx.Rollback(); rerr != nil {
			err = fmt.Errorf("%v (rollback failed: %v)", err, rerr)
		}
		return nil, err
	}

	if _, err := tx.User.UpdateOneID(userID).AddBalance(amount).Save(ctx); err != nil {
		return rollback(fmt.Errorf("adjust user balance: %w", err))
	}

	u, err := tx.User.Get(ctx, userID)
	if err != nil {
		return rollback(fmt.Errorf("get updated user: %w", err))
	}

	b := tx.BalanceRecord.Create().
		SetUserID(userID).
		SetType(balancerecord.TypeAdminAdjust).
		SetAmount(amount).
		SetBalanceAfter(u.Balance)
	if remark != nil {
		b.SetRemark(*remark)
	}
	if _, err := b.Save(ctx); err != nil {
		return rollback(fmt.Errorf("create admin adjust balance record: %w", err))
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit admin adjust: %w", err)
	}
	return u, nil
}

// quotaAvailable returns a predicate that is true when the token has unlimited
// quota or enough remaining quota for the estimated amount.
func quotaAvailable(estimated int64) predicate.UserToken {
	return predicate.UserToken(func(s *sql.Selector) {
		s.Where(sql.ExprP("(quota_limit IS NULL OR quota_limit - quota_used >= ?)", estimated))
	})
}

var _ BillingRepository = (*EntBillingRepository)(nil)
