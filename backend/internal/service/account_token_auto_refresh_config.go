package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"golang.org/x/sync/singleflight"
)

const (
	sfKeyAccountTokenAutoRefreshConfig = "account_token_auto_refresh_config"
	accountTokenAutoConfigCacheTTL     = 60 * time.Second
	accountTokenAutoConfigErrorTTL     = 5 * time.Second
	accountTokenAutoConfigDBTimeout    = 5 * time.Second
	DefaultTokenRefreshIntervalValue   = 1
	DefaultTokenRefreshIntervalUnit    = "day"
	DefaultTokenRefreshBatchSize       = 10
	maxTokenRefreshIntervalValue       = 365
	maxTokenRefreshBatchSize           = 50
)

type AccountTokenAutoRefreshConfig struct {
	Enabled        bool   `json:"enabled"`
	IntervalValue  int    `json:"interval_value"`
	IntervalUnit   string `json:"interval_unit"`
	BatchSize      int    `json:"batch_size"`
	LastRunAt      *int64 `json:"last_run_at,omitempty"`
	LastRunTotal   int    `json:"last_run_total,omitempty"`
	LastRunSuccess int    `json:"last_run_success,omitempty"`
	LastRunFailed  int    `json:"last_run_failed,omitempty"`
}

type cachedAccountTokenAutoRefreshConfig struct {
	config    *AccountTokenAutoRefreshConfig
	expiresAt int64
}

var accountTokenAutoRefreshCache atomic.Value
var accountTokenAutoRefreshSF singleflight.Group

func defaultAccountTokenAutoRefreshConfig() *AccountTokenAutoRefreshConfig {
	return &AccountTokenAutoRefreshConfig{
		Enabled:       false,
		IntervalValue: DefaultTokenRefreshIntervalValue,
		IntervalUnit:  DefaultTokenRefreshIntervalUnit,
		BatchSize:     DefaultTokenRefreshBatchSize,
	}
}

func normalizeAccountTokenAutoRefreshConfig(cfg *AccountTokenAutoRefreshConfig) *AccountTokenAutoRefreshConfig {
	if cfg == nil {
		return defaultAccountTokenAutoRefreshConfig()
	}
	out := *cfg
	out.IntervalUnit = strings.ToLower(strings.TrimSpace(out.IntervalUnit))
	if out.IntervalUnit != "hour" && out.IntervalUnit != "day" {
		out.IntervalUnit = DefaultTokenRefreshIntervalUnit
	}
	if out.IntervalValue <= 0 {
		out.IntervalValue = DefaultTokenRefreshIntervalValue
	}
	if out.BatchSize <= 0 {
		out.BatchSize = DefaultTokenRefreshBatchSize
	}
	return &out
}

func validateAccountTokenAutoRefreshConfig(cfg *AccountTokenAutoRefreshConfig) error {
	if cfg == nil {
		return nil
	}
	if cfg.IntervalValue < 1 || cfg.IntervalValue > maxTokenRefreshIntervalValue {
		return fmt.Errorf("interval_value must be between 1 and %d", maxTokenRefreshIntervalValue)
	}
	if cfg.IntervalUnit != "hour" && cfg.IntervalUnit != "day" {
		return errors.New("interval_unit must be hour or day")
	}
	if cfg.BatchSize < 1 || cfg.BatchSize > maxTokenRefreshBatchSize {
		return fmt.Errorf("batch_size must be between 1 and %d", maxTokenRefreshBatchSize)
	}
	return nil
}

func parseAccountTokenAutoRefreshConfigJSON(raw string) *AccountTokenAutoRefreshConfig {
	cfg := defaultAccountTokenAutoRefreshConfig()
	if strings.TrimSpace(raw) == "" {
		return cfg
	}
	if err := json.Unmarshal([]byte(raw), cfg); err != nil {
		return defaultAccountTokenAutoRefreshConfig()
	}
	return normalizeAccountTokenAutoRefreshConfig(cfg)
}

func accountTokenAutoRefreshDuration(cfg *AccountTokenAutoRefreshConfig) time.Duration {
	normalized := normalizeAccountTokenAutoRefreshConfig(cfg)
	if normalized.IntervalUnit == "hour" {
		return time.Duration(normalized.IntervalValue) * time.Hour
	}
	return time.Duration(normalized.IntervalValue) * 24 * time.Hour
}

func (s *SettingService) GetAccountTokenAutoRefreshConfig(ctx context.Context) (*AccountTokenAutoRefreshConfig, error) {
	if cached := accountTokenAutoRefreshCache.Load(); cached != nil {
		if c, ok := cached.(*cachedAccountTokenAutoRefreshConfig); ok && time.Now().UnixNano() < c.expiresAt {
			return c.config, nil
		}
	}
	result, err, _ := accountTokenAutoRefreshSF.Do(sfKeyAccountTokenAutoRefreshConfig, func() (any, error) {
		return s.loadAccountTokenAutoRefreshConfigFromDB()
	})
	if err != nil {
		return defaultAccountTokenAutoRefreshConfig(), err
	}
	if cfg, ok := result.(*AccountTokenAutoRefreshConfig); ok {
		return cfg, nil
	}
	return defaultAccountTokenAutoRefreshConfig(), nil
}

func (s *SettingService) loadAccountTokenAutoRefreshConfigFromDB() (*AccountTokenAutoRefreshConfig, error) {
	dbCtx, cancel := context.WithTimeout(context.Background(), accountTokenAutoConfigDBTimeout)
	defer cancel()

	raw, err := s.settingRepo.GetValue(dbCtx, SettingKeyAccountTokenAutoRefreshConfig)
	if err != nil {
		cfg := defaultAccountTokenAutoRefreshConfig()
		if errors.Is(err, ErrSettingNotFound) {
			accountTokenAutoRefreshCache.Store(&cachedAccountTokenAutoRefreshConfig{
				config:    cfg,
				expiresAt: time.Now().Add(accountTokenAutoConfigCacheTTL).UnixNano(),
			})
			return cfg, nil
		}
		accountTokenAutoRefreshCache.Store(&cachedAccountTokenAutoRefreshConfig{
			config:    cfg,
			expiresAt: time.Now().Add(accountTokenAutoConfigErrorTTL).UnixNano(),
		})
		return cfg, err
	}
	cfg := parseAccountTokenAutoRefreshConfigJSON(raw)
	accountTokenAutoRefreshCache.Store(&cachedAccountTokenAutoRefreshConfig{
		config:    cfg,
		expiresAt: time.Now().Add(accountTokenAutoConfigCacheTTL).UnixNano(),
	})
	return cfg, nil
}

func (s *SettingService) SaveAccountTokenAutoRefreshConfig(ctx context.Context, cfg *AccountTokenAutoRefreshConfig) error {
	cfg = normalizeAccountTokenAutoRefreshConfig(cfg)
	if err := validateAccountTokenAutoRefreshConfig(cfg); err != nil {
		return infraerrors.BadRequest("INVALID_ACCOUNT_TOKEN_AUTO_REFRESH_CONFIG", err.Error())
	}
	existing, _ := s.GetAccountTokenAutoRefreshConfig(ctx)
	if existing != nil {
		cfg.LastRunAt = existing.LastRunAt
		cfg.LastRunTotal = existing.LastRunTotal
		cfg.LastRunSuccess = existing.LastRunSuccess
		cfg.LastRunFailed = existing.LastRunFailed
	}
	return s.storeAccountTokenAutoRefreshConfig(ctx, cfg)
}

func (s *SettingService) MarkAccountTokenAutoRefreshRun(
	ctx context.Context,
	runAt time.Time,
	total int,
	success int,
	failed int,
) error {
	cfg, err := s.GetAccountTokenAutoRefreshConfig(ctx)
	if err != nil {
		cfg = defaultAccountTokenAutoRefreshConfig()
	}
	cfg = normalizeAccountTokenAutoRefreshConfig(cfg)
	ts := runAt.Unix()
	cfg.LastRunAt = &ts
	cfg.LastRunTotal = total
	cfg.LastRunSuccess = success
	cfg.LastRunFailed = failed
	return s.storeAccountTokenAutoRefreshConfig(ctx, cfg)
}

func (s *SettingService) storeAccountTokenAutoRefreshConfig(ctx context.Context, cfg *AccountTokenAutoRefreshConfig) error {
	data, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("account token auto refresh: marshal config: %w", err)
	}
	if err := s.settingRepo.Set(ctx, SettingKeyAccountTokenAutoRefreshConfig, string(data)); err != nil {
		return fmt.Errorf("account token auto refresh: save config: %w", err)
	}
	accountTokenAutoRefreshSF.Forget(sfKeyAccountTokenAutoRefreshConfig)
	accountTokenAutoRefreshCache.Store(&cachedAccountTokenAutoRefreshConfig{
		config:    cfg,
		expiresAt: time.Now().Add(accountTokenAutoConfigCacheTTL).UnixNano(),
	})
	return nil
}
