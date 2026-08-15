//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/stretchr/testify/require"
)

type mihomoPoolAccountRepo struct {
	AccountRepository
	accounts []Account
}

func (r *mihomoPoolAccountRepo) ListByPlatform(_ context.Context, platform string) ([]Account, error) {
	accounts := make([]Account, 0, len(r.accounts))
	for _, account := range r.accounts {
		if account.Platform == platform {
			accounts = append(accounts, account)
		}
	}
	return accounts, nil
}

func (r *mihomoPoolAccountRepo) ListSchedulableByPlatform(_ context.Context, platform string) ([]Account, error) {
	accounts := make([]Account, 0, len(r.accounts))
	for _, account := range r.accounts {
		if account.Platform == platform && account.IsSchedulable() {
			accounts = append(accounts, account)
		}
	}
	return accounts, nil
}

func TestExpandMihomoPoolProxyExclusionsDefersSiblingEgressAccount(t *testing.T) {
	proxyOne := int64(101)
	proxyTwo := int64(102)
	repo := &mihomoPoolAccountRepo{accounts: []Account{
		{ID: 1, Platform: PlatformGrok, ProxyID: &proxyOne, Extra: map[string]any{"mihomo_pool_managed": true}},
		{ID: 2, Platform: PlatformGrok, ProxyID: &proxyOne, Extra: map[string]any{"mihomo_pool_managed": true}},
		{ID: 3, Platform: PlatformGrok, ProxyID: &proxyTwo, Extra: map[string]any{"mihomo_pool_managed": true}},
		{ID: 4, Platform: PlatformGrok, ProxyID: &proxyOne},
	}}
	svc := &GatewayService{accountRepo: repo}

	excluded := svc.expandMihomoPoolProxyExclusions(context.Background(), map[int64]struct{}{1: {}})

	require.Contains(t, excluded, int64(1))
	require.Contains(t, excluded, int64(2))
	require.NotContains(t, excluded, int64(3))
	require.NotContains(t, excluded, int64(4))
}

func TestHasAdditionalAccountExclusionsAllowsSiblingFallback(t *testing.T) {
	original := map[int64]struct{}{1: {}}
	preferred := map[int64]struct{}{1: {}, 2: {}}

	require.True(t, hasAdditionalAccountExclusions(original, preferred))
	require.False(t, hasAdditionalAccountExclusions(original, original))
}

func TestSelectAccountForModelWithExclusionsPrefersOtherEgressAndFallsBackToSibling(t *testing.T) {
	proxyOne := int64(101)
	proxyTwo := int64(102)
	accounts := []Account{
		{ID: 1, Platform: PlatformGrok, Type: AccountTypeOAuth, Status: StatusActive, Schedulable: true, ProxyID: &proxyOne, Extra: map[string]any{"mihomo_pool_managed": true}},
		{ID: 2, Platform: PlatformGrok, Type: AccountTypeOAuth, Status: StatusActive, Schedulable: true, ProxyID: &proxyOne, Extra: map[string]any{"mihomo_pool_managed": true}},
		{ID: 3, Platform: PlatformGrok, Type: AccountTypeOAuth, Status: StatusActive, Schedulable: true, ProxyID: &proxyTwo, Extra: map[string]any{"mihomo_pool_managed": true}},
	}
	ctx := context.WithValue(context.Background(), ctxkey.ForcePlatform, PlatformGrok)

	svc := &GatewayService{
		accountRepo: &mihomoPoolAccountRepo{accounts: accounts},
		cfg:         &config.Config{RunMode: config.RunModeSimple},
	}
	selected, err := svc.SelectAccountForModelWithExclusions(ctx, nil, "", "", map[int64]struct{}{1: {}})
	require.NoError(t, err)
	require.Equal(t, int64(3), selected.ID, "a healthy account on another egress should be preferred")

	svc.accountRepo = &mihomoPoolAccountRepo{accounts: accounts[:2]}
	selected, err = svc.SelectAccountForModelWithExclusions(ctx, nil, "", "", map[int64]struct{}{1: {}})
	require.NoError(t, err)
	require.Equal(t, int64(2), selected.ID, "same-egress sibling remains eligible as fallback")
}

func TestExpandMihomoPoolProxyExclusionsLeavesNonPoolAccountAlone(t *testing.T) {
	proxyID := int64(101)
	repo := &mihomoPoolAccountRepo{accounts: []Account{
		{ID: 1, Platform: PlatformGrok, ProxyID: &proxyID},
		{ID: 2, Platform: PlatformGrok, ProxyID: &proxyID, Extra: map[string]any{"mihomo_pool_managed": true}},
	}}
	svc := &GatewayService{accountRepo: repo}

	excluded := svc.expandMihomoPoolProxyExclusions(context.Background(), map[int64]struct{}{1: {}})

	require.Equal(t, map[int64]struct{}{1: {}}, excluded)
}
