package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExpandMihomoPoolProxyExclusionsExcludesSiblingEgressAccount(t *testing.T) {
	proxyOne := int64(101)
	proxyTwo := int64(102)
	repo := &mockAccountRepoForPlatform{accounts: []Account{
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
	repo := &mockAccountRepoForPlatform{accounts: []Account{
		{ID: 1, Platform: PlatformGrok, ProxyID: &proxyID},
		{ID: 2, Platform: PlatformGrok, ProxyID: &proxyID, Extra: map[string]any{"mihomo_pool_managed": true}},
	}}
	svc := &GatewayService{accountRepo: repo}

	excluded := svc.expandMihomoPoolProxyExclusions(context.Background(), map[int64]struct{}{1: {}})

	require.Equal(t, map[int64]struct{}{1: {}}, excluded)
}
