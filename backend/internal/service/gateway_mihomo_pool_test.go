//go:build unit

package service

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/stretchr/testify/require"
)

type mihomoPoolAccountRepo struct {
	AccountRepository
	accounts []Account
}

func (r *mihomoPoolAccountRepo) GetByID(_ context.Context, id int64) (*Account, error) {
	for i := range r.accounts {
		if r.accounts[i].ID == id {
			account := r.accounts[i]
			return &account, nil
		}
	}
	return nil, ErrAccountNotFound
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
		if account.Platform == platform && account.Schedulable && account.IsSchedulable() {
			accounts = append(accounts, account)
		}
	}
	return accounts, nil
}

func init() {
	mihomoPoolAccountCacheDisabled = true
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
		{ID: 2, Platform: PlatformGrok, Type: AccountTypeOAuth, Status: StatusActive, Schedulable: false, ProxyID: &proxyOne, Extra: map[string]any{"mihomo_pool_managed": true}},
		{ID: 3, Platform: PlatformGrok, Type: AccountTypeOAuth, Status: StatusActive, Schedulable: false, ProxyID: &proxyTwo, Extra: map[string]any{"mihomo_pool_managed": true}},
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

func TestMihomoPoolManagedLegacySchedulableFalseHonorsCooldown(t *testing.T) {
	account := &Account{
		ID:          10,
		Platform:    PlatformGrok,
		Status:      StatusActive,
		Schedulable: false,
		Extra:       map[string]any{"mihomo_pool_managed": true},
	}
	require.True(t, account.IsSchedulable())

	resetAt := time.Now().Add(time.Hour)
	account.RateLimitResetAt = &resetAt
	require.False(t, account.IsSchedulable())

	resetAt = time.Now().Add(-time.Second)
	account.RateLimitResetAt = &resetAt
	require.True(t, account.IsSchedulable())
}

func TestMihomoPoolStandbyWaitsForEgressBinding(t *testing.T) {
	account := &Account{
		ID:          13,
		Platform:    PlatformGrok,
		Type:        AccountTypeOAuth,
		Status:      StatusActive,
		Schedulable: true,
		Extra: map[string]any{
			"mihomo_pool_managed": true,
			"mihomo_pool_standby": true,
		},
	}

	require.False(t, account.IsSchedulable())

	account.Extra["mihomo_pool_standby"] = false
	require.True(t, account.IsSchedulable())
}

func TestGrokOAuthAccountIgnoresLegacyStandbyFlagButHonorsManualPause(t *testing.T) {
	account := &Account{
		ID:          11,
		Platform:    PlatformGrok,
		Type:        AccountTypeOAuth,
		Status:      StatusActive,
		Schedulable: false,
	}
	require.True(t, account.IsSchedulable())

	account.Extra = map[string]any{ManualSchedulingDisabledExtraKey: true}
	require.False(t, account.IsSchedulable())
}

func TestGrokOAuthAccountWithoutPoolMarkerReturnsToCandidatePool(t *testing.T) {
	account := Account{
		ID:          12,
		Platform:    PlatformGrok,
		Type:        AccountTypeOAuth,
		Status:      StatusActive,
		Schedulable: false,
	}
	svc := &GatewayService{
		accountRepo: &mihomoPoolAccountRepo{accounts: []Account{account}},
		cfg:         &config.Config{RunMode: config.RunModeSimple},
	}

	selected, err := svc.SelectAccountForModelWithExclusions(
		context.WithValue(context.Background(), ctxkey.ForcePlatform, PlatformGrok),
		nil,
		"",
		"",
		nil,
	)
	require.NoError(t, err)
	require.Equal(t, account.ID, selected.ID)
}

func TestSelectAccountForModelWithExclusionsSkipsExhaustedMihomoAccountAndRestoresStandby(t *testing.T) {
	proxyOne := int64(101)
	proxyTwo := int64(102)
	resetAt := time.Now().Add(time.Hour)
	accounts := []Account{
		{ID: 1, Platform: PlatformGrok, Type: AccountTypeOAuth, Status: StatusActive, Schedulable: true, ProxyID: &proxyOne, RateLimitResetAt: &resetAt, Extra: map[string]any{"mihomo_pool_managed": true}},
		{ID: 2, Platform: PlatformGrok, Type: AccountTypeOAuth, Status: StatusActive, Schedulable: false, ProxyID: &proxyTwo, Extra: map[string]any{"mihomo_pool_managed": true}},
	}
	svc := &GatewayService{
		accountRepo: &mihomoPoolAccountRepo{accounts: accounts},
		cfg:         &config.Config{RunMode: config.RunModeSimple},
	}

	selected, err := svc.SelectAccountForModelWithExclusions(
		context.WithValue(context.Background(), ctxkey.ForcePlatform, PlatformGrok),
		nil,
		"",
		"",
		map[int64]struct{}{1: {}},
	)
	require.NoError(t, err)
	require.Equal(t, int64(2), selected.ID)
}

func TestOpenAICompatibleGrokSelectionRestoresMihomoStandbyAccount(t *testing.T) {
	proxyOne := int64(101)
	proxyTwo := int64(102)
	resetAt := time.Now().Add(time.Hour)
	accounts := []Account{
		{ID: 1, Platform: PlatformGrok, Type: AccountTypeOAuth, Status: StatusActive, Schedulable: true, ProxyID: &proxyOne, RateLimitResetAt: &resetAt, Extra: map[string]any{"mihomo_pool_managed": true}},
		{ID: 2, Platform: PlatformGrok, Type: AccountTypeOAuth, Status: StatusActive, Schedulable: false, ProxyID: &proxyTwo, Extra: map[string]any{"mihomo_pool_managed": true}},
	}
	svc := &OpenAIGatewayService{
		accountRepo: &mihomoPoolAccountRepo{accounts: accounts},
		cfg:         &config.Config{RunMode: config.RunModeSimple},
	}

	selected, err := svc.selectAccountForModelWithExclusions(
		context.Background(),
		nil,
		PlatformGrok,
		"",
		"",
		map[int64]struct{}{1: {}},
		false,
		0,
		"",
		false,
	)
	require.NoError(t, err)
	require.Equal(t, int64(2), selected.ID)
}

func TestOpenAIAdvancedSchedulerGrokPrefersAnotherMihomoEgress(t *testing.T) {
	resetOpenAIAdvancedSchedulerSettingCacheForTest()
	defer resetOpenAIAdvancedSchedulerSettingCacheForTest()

	proxyOne := int64(101)
	proxyTwo := int64(102)
	accounts := []Account{
		{ID: 1, Platform: PlatformGrok, Type: AccountTypeOAuth, Status: StatusActive, Schedulable: true, Concurrency: 1, Priority: 0, ProxyID: &proxyOne, Extra: map[string]any{"mihomo_pool_managed": true}},
		{ID: 2, Platform: PlatformGrok, Type: AccountTypeOAuth, Status: StatusActive, Schedulable: true, Concurrency: 1, Priority: 0, ProxyID: &proxyOne, Extra: map[string]any{"mihomo_pool_managed": true}},
		{ID: 3, Platform: PlatformGrok, Type: AccountTypeOAuth, Status: StatusActive, Schedulable: true, Concurrency: 1, Priority: 10, ProxyID: &proxyTwo, Extra: map[string]any{"mihomo_pool_managed": true}},
	}
	cfg := &config.Config{RunMode: config.RunModeSimple}
	cfg.Gateway.OpenAIWS.LBTopK = 1
	svc := &OpenAIGatewayService{
		accountRepo:        &mihomoPoolAccountRepo{accounts: accounts},
		cfg:                cfg,
		concurrencyService: NewConcurrencyService(schedulerTestConcurrencyCache{}),
		rateLimitService:   newOpenAIAdvancedSchedulerRateLimitService("true"),
	}

	selection, _, err := svc.SelectAccountWithSchedulerForCapability(
		context.Background(), nil, "", "", "grok-4.5",
		map[int64]struct{}{1: {}}, OpenAIUpstreamTransportHTTPSSE, "",
		false, false, false, PlatformGrok,
	)

	require.NoError(t, err)
	require.NotNil(t, selection)
	require.NotNil(t, selection.Account)
	require.Equal(t, int64(3), selection.Account.ID, "another egress should win before the higher-priority sibling")
}

func TestOpenAIAdvancedSchedulerDoesNotExpandMihomoExclusionsForOpenAI(t *testing.T) {
	proxyID := int64(101)
	svc := &OpenAIGatewayService{accountRepo: &mihomoPoolAccountRepo{accounts: []Account{
		{ID: 1, Platform: PlatformGrok, ProxyID: &proxyID, Extra: map[string]any{"mihomo_pool_managed": true}},
		{ID: 2, Platform: PlatformGrok, ProxyID: &proxyID, Extra: map[string]any{"mihomo_pool_managed": true}},
	}}}
	original := map[int64]struct{}{1: {}}

	preferred := svc.expandGrokPreferredExclusions(context.Background(), PlatformOpenAI, original)

	require.Equal(t, original, preferred)
	require.NotContains(t, preferred, int64(2))
}

func TestOpenAIAdvancedSchedulerGrokFallsBackToSameMihomoEgress(t *testing.T) {
	resetOpenAIAdvancedSchedulerSettingCacheForTest()
	defer resetOpenAIAdvancedSchedulerSettingCacheForTest()

	proxyID := int64(101)
	accounts := []Account{
		{ID: 1, Platform: PlatformGrok, Type: AccountTypeOAuth, Status: StatusActive, Schedulable: true, Concurrency: 1, ProxyID: &proxyID, Extra: map[string]any{"mihomo_pool_managed": true}},
		{ID: 2, Platform: PlatformGrok, Type: AccountTypeOAuth, Status: StatusActive, Schedulable: true, Concurrency: 1, ProxyID: &proxyID, Extra: map[string]any{"mihomo_pool_managed": true}},
	}
	cfg := &config.Config{RunMode: config.RunModeSimple}
	cfg.Gateway.OpenAIWS.LBTopK = 1
	svc := &OpenAIGatewayService{
		accountRepo:        &mihomoPoolAccountRepo{accounts: accounts},
		cfg:                cfg,
		concurrencyService: NewConcurrencyService(schedulerTestConcurrencyCache{}),
		rateLimitService:   newOpenAIAdvancedSchedulerRateLimitService("true"),
	}

	selection, _, err := svc.SelectAccountWithSchedulerForCapability(
		context.Background(), nil, "", "", "grok-4.5",
		map[int64]struct{}{1: {}}, OpenAIUpstreamTransportHTTPSSE, "",
		false, false, false, PlatformGrok,
	)

	require.NoError(t, err)
	require.NotNil(t, selection)
	require.NotNil(t, selection.Account)
	require.Equal(t, int64(2), selection.Account.ID, "same-egress sibling must remain available as the fail-open pass")
}

func TestOpenAISSECommentLine(t *testing.T) {
	require.True(t, openAISSECommentLine(": ping"))
	require.True(t, openAISSECommentLine(":"))
	require.True(t, openAISSECommentLine(" : keepalive"))
	require.False(t, openAISSECommentLine(""))
	require.False(t, openAISSECommentLine(`data: {"type":"response.created"}`))
	require.False(t, openAISSECommentLine("event: response.created"))
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

type countingMihomoPoolRepo struct {
	mihomoPoolAccountRepo
	listCalls int
}

func (r *countingMihomoPoolRepo) ListByPlatform(ctx context.Context, platform string) ([]Account, error) {
	r.listCalls++
	return r.mihomoPoolAccountRepo.ListByPlatform(ctx, platform)
}

func TestMihomoPoolAccountListCacheHitsWithinTTL(t *testing.T) {
	repo := &countingMihomoPoolRepo{mihomoPoolAccountRepo: mihomoPoolAccountRepo{accounts: []Account{
		{ID: 1, Platform: PlatformGrok, Extra: map[string]any{"mihomo_pool_managed": true}},
	}}}
	cache := &mihomoPoolAccountListCache{}

	first, err := cache.listWithTTL(context.Background(), repo, time.Minute)
	require.NoError(t, err)
	require.Len(t, first, 1)
	require.Equal(t, 1, repo.listCalls)

	second, err := cache.listWithTTL(context.Background(), repo, time.Minute)
	require.NoError(t, err)
	require.Equal(t, first[0].ID, second[0].ID)
	require.Equal(t, 1, repo.listCalls)

	third, err := cache.listWithTTL(context.Background(), repo, 0)
	require.NoError(t, err)
	require.Equal(t, first[0].ID, third[0].ID)
	require.Equal(t, 2, repo.listCalls)
}
