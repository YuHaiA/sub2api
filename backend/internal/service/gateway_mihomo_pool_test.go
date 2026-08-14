//go:build unit

package service

import (
	"context"
	"testing"

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

func TestExpandMihomoPoolProxyExclusionsExcludesSiblingEgressAccount(t *testing.T) {
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
