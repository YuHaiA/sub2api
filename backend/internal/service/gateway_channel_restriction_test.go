//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	channelRestrictionTestModel      = "gpt-5.5"
	channelRestrictionAlternateModel = "gpt-5.4-mini"
)

// --- billingModelForRestriction ---

func TestBillingModelForRestriction_Requested(t *testing.T) {
	t.Parallel()
	got := billingModelForRestriction(BillingModelSourceRequested, channelRestrictionTestModel, channelRestrictionAlternateModel)
	require.Equal(t, channelRestrictionTestModel, got)
}

func TestBillingModelForRestriction_ChannelMapped(t *testing.T) {
	t.Parallel()
	got := billingModelForRestriction(BillingModelSourceChannelMapped, channelRestrictionAlternateModel, channelRestrictionTestModel)
	require.Equal(t, channelRestrictionTestModel, got)
}

func TestBillingModelForRestriction_Upstream(t *testing.T) {
	t.Parallel()
	got := billingModelForRestriction(BillingModelSourceUpstream, channelRestrictionTestModel, channelRestrictionAlternateModel)
	require.Equal(t, "", got, "upstream should return empty (per-account check needed)")
}

func TestBillingModelForRestriction_Empty(t *testing.T) {
	t.Parallel()
	got := billingModelForRestriction("", channelRestrictionAlternateModel, channelRestrictionTestModel)
	require.Equal(t, channelRestrictionTestModel, got, "empty source defaults to channel_mapped")
}

// --- resolveAccountUpstreamModel ---

func TestResolveAccountUpstreamModel_OpenAIPassthrough(t *testing.T) {
	t.Parallel()
	account := &Account{
		Platform: PlatformOpenAI,
	}
	got := resolveAccountUpstreamModel(account, channelRestrictionTestModel)
	require.Equal(t, channelRestrictionTestModel, got)
}

func TestResolveAccountUpstreamModel_NonOpenAIPassthrough(t *testing.T) {
	t.Parallel()
	account := &Account{
		Platform: PlatformAnthropic,
	}
	got := resolveAccountUpstreamModel(account, channelRestrictionTestModel)
	require.Equal(t, channelRestrictionTestModel, got, "no mapping = passthrough")
}

// --- checkChannelPricingRestriction ---

func TestCheckChannelPricingRestriction_NilGroupID(t *testing.T) {
	t.Parallel()
	svc := &GatewayService{channelService: &ChannelService{}}
	require.False(t, svc.checkChannelPricingRestriction(context.Background(), nil, channelRestrictionTestModel))
}

func TestCheckChannelPricingRestriction_NilChannelService(t *testing.T) {
	t.Parallel()
	svc := &GatewayService{}
	gid := int64(10)
	require.False(t, svc.checkChannelPricingRestriction(context.Background(), &gid, channelRestrictionTestModel))
}

func TestCheckChannelPricingRestriction_EmptyModel(t *testing.T) {
	t.Parallel()
	svc := &GatewayService{channelService: &ChannelService{}}
	gid := int64(10)
	require.False(t, svc.checkChannelPricingRestriction(context.Background(), &gid, ""))
}

func TestCheckChannelPricingRestriction_ChannelMapped_Restricted(t *testing.T) {
	t.Parallel()
	// 渠道映射到 gpt-5.5，但定价列表只有其他模型。
	ch := Channel{
		ID:                 1,
		Status:             StatusActive,
		GroupIDs:           []int64{10},
		RestrictModels:     true,
		BillingModelSource: BillingModelSourceChannelMapped,
		ModelPricing: []ChannelModelPricing{
			{Platform: PlatformOpenAI, Models: []string{channelRestrictionAlternateModel}},
		},
		ModelMapping: map[string]map[string]string{
			PlatformOpenAI: {channelRestrictionTestModel: channelRestrictionTestModel},
		},
	}
	channelSvc := newTestChannelService(makeStandardRepo(ch, map[int64]string{10: PlatformOpenAI}))
	svc := &GatewayService{channelService: channelSvc}

	gid := int64(10)
	require.True(t, svc.checkChannelPricingRestriction(context.Background(), &gid, channelRestrictionTestModel),
		"mapped model gpt-5.5 is NOT in pricing -> restricted")
}

func TestCheckChannelPricingRestriction_ChannelMapped_Allowed(t *testing.T) {
	t.Parallel()
	// 渠道映射到 gpt-5.5，且定价列表包含 gpt-5.5。
	ch := Channel{
		ID:                 1,
		Status:             StatusActive,
		GroupIDs:           []int64{10},
		RestrictModels:     true,
		BillingModelSource: BillingModelSourceChannelMapped,
		ModelPricing: []ChannelModelPricing{
			{Platform: PlatformOpenAI, Models: []string{channelRestrictionTestModel}},
		},
		ModelMapping: map[string]map[string]string{
			PlatformOpenAI: {channelRestrictionTestModel: channelRestrictionTestModel},
		},
	}
	channelSvc := newTestChannelService(makeStandardRepo(ch, map[int64]string{10: PlatformOpenAI}))
	svc := &GatewayService{channelService: channelSvc}

	gid := int64(10)
	require.False(t, svc.checkChannelPricingRestriction(context.Background(), &gid, channelRestrictionTestModel),
		"mapped model gpt-5.5 IS in pricing -> allowed")
}

func TestCheckChannelPricingRestriction_Requested_Restricted(t *testing.T) {
	t.Parallel()
	// billing_model_source=requested，定价列表有其他模型但请求的是 gpt-5.5。
	ch := Channel{
		ID:                 1,
		Status:             StatusActive,
		GroupIDs:           []int64{10},
		RestrictModels:     true,
		BillingModelSource: BillingModelSourceRequested,
		ModelPricing: []ChannelModelPricing{
			{Platform: PlatformOpenAI, Models: []string{channelRestrictionAlternateModel}},
		},
	}
	channelSvc := newTestChannelService(makeStandardRepo(ch, map[int64]string{10: PlatformOpenAI}))
	svc := &GatewayService{channelService: channelSvc}

	gid := int64(10)
	require.True(t, svc.checkChannelPricingRestriction(context.Background(), &gid, channelRestrictionTestModel),
		"requested model gpt-5.5 is NOT in pricing -> restricted")
}

func TestCheckChannelPricingRestriction_Requested_Allowed(t *testing.T) {
	t.Parallel()
	ch := Channel{
		ID:                 1,
		Status:             StatusActive,
		GroupIDs:           []int64{10},
		RestrictModels:     true,
		BillingModelSource: BillingModelSourceRequested,
		ModelPricing: []ChannelModelPricing{
			{Platform: PlatformOpenAI, Models: []string{channelRestrictionTestModel}},
		},
	}
	channelSvc := newTestChannelService(makeStandardRepo(ch, map[int64]string{10: PlatformOpenAI}))
	svc := &GatewayService{channelService: channelSvc}

	gid := int64(10)
	require.False(t, svc.checkChannelPricingRestriction(context.Background(), &gid, channelRestrictionTestModel),
		"requested model IS in pricing → allowed")
}

func TestCheckChannelPricingRestriction_EmptyPricingAllowlistDoesNotDenyAll(t *testing.T) {
	t.Parallel()
	ch := Channel{
		ID:                 1,
		Status:             StatusActive,
		GroupIDs:           []int64{10},
		RestrictModels:     true,
		BillingModelSource: BillingModelSourceChannelMapped,
		ModelMapping: map[string]map[string]string{
			PlatformOpenAI: {channelRestrictionTestModel: channelRestrictionTestModel},
		},
	}
	channelSvc := newTestChannelService(makeStandardRepo(ch, map[int64]string{10: PlatformOpenAI}))
	svc := &GatewayService{channelService: channelSvc}

	gid := int64(10)
	require.False(t, svc.checkChannelPricingRestriction(context.Background(), &gid, channelRestrictionTestModel),
		"restrict_models without any pricing rows must not turn into deny-all")
}

func TestCheckChannelPricingRestriction_Upstream_SkipsPreCheck(t *testing.T) {
	t.Parallel()
	// upstream 模式：预检查始终跳过（返回 false），需逐账号检查
	ch := Channel{
		ID:                 1,
		Status:             StatusActive,
		GroupIDs:           []int64{10},
		RestrictModels:     true,
		BillingModelSource: BillingModelSourceUpstream,
		ModelPricing: []ChannelModelPricing{
			{Platform: PlatformOpenAI, Models: []string{channelRestrictionAlternateModel}},
		},
	}
	channelSvc := newTestChannelService(makeStandardRepo(ch, map[int64]string{10: PlatformOpenAI}))
	svc := &GatewayService{channelService: channelSvc}

	gid := int64(10)
	require.False(t, svc.checkChannelPricingRestriction(context.Background(), &gid, channelRestrictionTestModel),
		"upstream mode should skip pre-check (per-account check needed)")
}

func TestCheckChannelPricingRestriction_RestrictModelsDisabled(t *testing.T) {
	t.Parallel()
	ch := Channel{
		ID:             1,
		Status:         StatusActive,
		GroupIDs:       []int64{10},
		RestrictModels: false, // 未开启模型限制
		ModelPricing: []ChannelModelPricing{
			{Platform: PlatformOpenAI, Models: []string{channelRestrictionAlternateModel}},
		},
	}
	channelSvc := newTestChannelService(makeStandardRepo(ch, map[int64]string{10: PlatformOpenAI}))
	svc := &GatewayService{channelService: channelSvc}

	gid := int64(10)
	require.False(t, svc.checkChannelPricingRestriction(context.Background(), &gid, channelRestrictionTestModel),
		"RestrictModels=false → always allowed")
}

func TestCheckChannelPricingRestriction_NoChannel(t *testing.T) {
	t.Parallel()
	// 分组没有关联渠道
	repo := &mockChannelRepository{
		listAllFn: func(_ context.Context) ([]Channel, error) { return nil, nil },
	}
	channelSvc := newTestChannelService(repo)
	svc := &GatewayService{channelService: channelSvc}

	gid := int64(999)
	require.False(t, svc.checkChannelPricingRestriction(context.Background(), &gid, channelRestrictionTestModel),
		"no channel for group → allowed")
}

// --- isUpstreamModelRestrictedByChannel ---

func TestIsUpstreamModelRestrictedByChannel_Restricted(t *testing.T) {
	t.Parallel()
	ch := Channel{
		ID:             1,
		Status:         StatusActive,
		GroupIDs:       []int64{10},
		RestrictModels: true,
		ModelPricing: []ChannelModelPricing{
			{Platform: PlatformOpenAI, Models: []string{channelRestrictionAlternateModel}},
		},
	}
	channelSvc := newTestChannelService(makeStandardRepo(ch, map[int64]string{10: PlatformOpenAI}))
	svc := &GatewayService{channelService: channelSvc}

	account := &Account{Platform: PlatformOpenAI}
	require.True(t, svc.isUpstreamModelRestrictedByChannel(context.Background(), 10, account, channelRestrictionTestModel),
		"upstream model gpt-5.5 NOT in pricing -> restricted")
}

func TestIsUpstreamModelRestrictedByChannel_Allowed(t *testing.T) {
	t.Parallel()
	ch := Channel{
		ID:             1,
		Status:         StatusActive,
		GroupIDs:       []int64{10},
		RestrictModels: true,
		ModelPricing: []ChannelModelPricing{
			{Platform: PlatformOpenAI, Models: []string{channelRestrictionTestModel}},
		},
	}
	channelSvc := newTestChannelService(makeStandardRepo(ch, map[int64]string{10: PlatformOpenAI}))
	svc := &GatewayService{channelService: channelSvc}

	account := &Account{Platform: PlatformOpenAI}
	require.False(t, svc.isUpstreamModelRestrictedByChannel(context.Background(), 10, account, channelRestrictionTestModel),
		"upstream model gpt-5.5 IS in pricing -> allowed")
}

func TestIsUpstreamModelRestrictedByChannel_EmptyRequestedModel(t *testing.T) {
	t.Parallel()
	ch := Channel{
		ID:             1,
		Status:         StatusActive,
		GroupIDs:       []int64{10},
		RestrictModels: true,
		ModelPricing: []ChannelModelPricing{
			{Platform: PlatformOpenAI, Models: []string{channelRestrictionTestModel}},
		},
	}
	channelSvc := newTestChannelService(makeStandardRepo(ch, map[int64]string{10: PlatformOpenAI}))
	svc := &GatewayService{channelService: channelSvc}

	account := &Account{Platform: PlatformOpenAI}
	require.False(t, svc.isUpstreamModelRestrictedByChannel(context.Background(), 10, account, ""),
		"empty upstream model should not be restricted")
}

func TestIsUpstreamModelRestrictedByChannel_EmptyPricingAllowlistDoesNotDenyAll(t *testing.T) {
	t.Parallel()
	ch := Channel{
		ID:             1,
		Status:         StatusActive,
		GroupIDs:       []int64{10},
		RestrictModels: true,
	}
	channelSvc := newTestChannelService(makeStandardRepo(ch, map[int64]string{10: PlatformOpenAI}))
	svc := &GatewayService{channelService: channelSvc}

	account := &Account{Platform: PlatformOpenAI}
	require.False(t, svc.isUpstreamModelRestrictedByChannel(context.Background(), 10, account, channelRestrictionTestModel),
		"empty pricing allowlist must not block every upstream model")
}
