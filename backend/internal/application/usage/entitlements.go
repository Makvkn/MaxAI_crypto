package usage

import (
	"context"

	"github.com/google/uuid"

	"github.com/maxaicrypto/backend/internal/domain/subscription"
	"github.com/maxaicrypto/backend/internal/domain/wallet"
)

// EntitlementApp resolves plan limits for a user.
type EntitlementApp struct {
	subscriptions subscription.Repository
	wallets       wallet.Repository
	aiDailyLimit  int
}

// NewEntitlementApp wires the entitlement service.
func NewEntitlementApp(
	subscriptions subscription.Repository,
	wallets wallet.Repository,
	aiDailyLimit int,
) *EntitlementApp {
	return &EntitlementApp{
		subscriptions: subscriptions,
		wallets:       wallets,
		aiDailyLimit:  aiDailyLimit,
	}
}

// Entitlements implements EntitlementService.
func (a *EntitlementApp) Entitlements(ctx context.Context, userID uuid.UUID) (subscription.Entitlements, error) {
	sub, ok, err := a.subscriptions.GetByUser(ctx, userID)
	if err != nil {
		return subscription.Entitlements{}, err
	}
	if !ok || sub.Plan == subscription.PlanFree {
		ent := subscription.FreeEntitlements
		ent.AIOperationsPerDay = a.aiDailyLimit
		return ent, nil
	}
	// PRO limits are not part of the MVP; keep the abstraction ready.
	ent := subscription.FreeEntitlements
	ent.MaxWallets = 10
	ent.AIOperationsPerDay = a.aiDailyLimit
	return ent, nil
}

// Can implements EntitlementService.
func (a *EntitlementApp) Can(ctx context.Context, userID uuid.UUID, feature subscription.Feature) (bool, error) {
	ent, err := a.Entitlements(ctx, userID)
	if err != nil {
		return false, err
	}
	return ent.Allows(feature), nil
}

// CanCreateWallet implements EntitlementService.
func (a *EntitlementApp) CanCreateWallet(ctx context.Context, userID uuid.UUID) (bool, error) {
	ent, err := a.Entitlements(ctx, userID)
	if err != nil {
		return false, err
	}
	count, err := a.wallets.CountByUser(ctx, userID)
	if err != nil {
		return false, err
	}
	return count < ent.MaxWallets, nil
}

var _ EntitlementService = (*EntitlementApp)(nil)
