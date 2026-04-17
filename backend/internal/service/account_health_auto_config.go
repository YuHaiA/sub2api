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
	sfKeyAccountHealthAutoCheckConfig   = "account_health_auto_check_config"
	accountHealthAutoConfigCacheTTL     = 60 * time.Second
	accountHealthAutoConfigErrorTTL     = 5 * time.Second
	accountHealthAutoConfigDBTimeout    = 5 * time.Second
	DefaultAccountHealthIntervalMinutes = 60
	maxAccountHealthIntervalMinutes     = 24 * 60
)

type AccountHealthAutoCheckConfig struct {
	Enabled         bool   `json:"enabled"`
	IntervalMinutes int    `json:"interval_minutes"`
	ModelID         string `json:"model_id"`
	LastRunAt       *int64 `json:"last_run_at,omitempty"`
}

type cachedAccountHealthAutoCheckConfig struct {
	config    *AccountHealthAutoCheckConfig
	expiresAt int64
}

var accountHealthAutoCheckCache atomic.Value // *cachedAccountHealthAutoCheckConfig
var accountHealthAutoCheckSF singleflight.Group

func defaultAccountHealthAutoCheckConfig() *AccountHealthAutoCheckConfig {
	return &AccountHealthAutoCheckConfig{
		Enabled:         false,
		IntervalMinutes: DefaultAccountHealthIntervalMinutes,
		ModelID:         "",
	}
}

func normalizeAccountHealthAutoCheckConfig(cfg *AccountHealthAutoCheckConfig) *AccountHealthAutoCheckConfig {
	if cfg == nil {
		return defaultAccountHealthAutoCheckConfig()
	}
	out := *cfg
	out.ModelID = strings.TrimSpace(out.ModelID)
	if out.IntervalMinutes <= 0 {
		out.IntervalMinutes = DefaultAccountHealthIntervalMinutes
	}
	return &out
}

func validateAccountHealthAutoCheckConfig(cfg *AccountHealthAutoCheckConfig) error {
	if cfg == nil {
		return nil
	}
	if cfg.IntervalMinutes < 1 || cfg.IntervalMinutes > maxAccountHealthIntervalMinutes {
		return fmt.Errorf("interval_minutes must be between 1 and %d", maxAccountHealthIntervalMinutes)
	}
	return nil
}

func parseAccountHealthAutoCheckConfigJSON(raw string) *AccountHealthAutoCheckConfig {
	cfg := defaultAccountHealthAutoCheckConfig()
	if strings.TrimSpace(raw) == "" {
		return cfg
	}
	if err := json.Unmarshal([]byte(raw), cfg); err != nil {
		return defaultAccountHealthAutoCheckConfig()
	}
	return normalizeAccountHealthAutoCheckConfig(cfg)
}

func (s *SettingService) GetAccountHealthAutoCheckConfig(ctx context.Context) (*AccountHealthAutoCheckConfig, error) {
	if cached := accountHealthAutoCheckCache.Load(); cached != nil {
		if c, ok := cached.(*cachedAccountHealthAutoCheckConfig); ok && time.Now().UnixNano() < c.expiresAt {
			return c.config, nil
		}
	}
	result, err, _ := accountHealthAutoCheckSF.Do(sfKeyAccountHealthAutoCheckConfig, func() (any, error) {
		return s.loadAccountHealthAutoCheckConfigFromDB()
	})
	if err != nil {
		return defaultAccountHealthAutoCheckConfig(), err
	}
	if cfg, ok := result.(*AccountHealthAutoCheckConfig); ok {
		return cfg, nil
	}
	return defaultAccountHealthAutoCheckConfig(), nil
}

func (s *SettingService) loadAccountHealthAutoCheckConfigFromDB() (*AccountHealthAutoCheckConfig, error) {
	dbCtx, cancel := context.WithTimeout(context.Background(), accountHealthAutoConfigDBTimeout)
	defer cancel()

	raw, err := s.settingRepo.GetValue(dbCtx, SettingKeyAccountHealthAutoCheckConfig)
	if err != nil {
		cfg := defaultAccountHealthAutoCheckConfig()
		if errors.Is(err, ErrSettingNotFound) {
			accountHealthAutoCheckCache.Store(&cachedAccountHealthAutoCheckConfig{
				config:    cfg,
				expiresAt: time.Now().Add(accountHealthAutoConfigCacheTTL).UnixNano(),
			})
			return cfg, nil
		}
		accountHealthAutoCheckCache.Store(&cachedAccountHealthAutoCheckConfig{
			config:    cfg,
			expiresAt: time.Now().Add(accountHealthAutoConfigErrorTTL).UnixNano(),
		})
		return cfg, err
	}
	cfg := parseAccountHealthAutoCheckConfigJSON(raw)
	accountHealthAutoCheckCache.Store(&cachedAccountHealthAutoCheckConfig{
		config:    cfg,
		expiresAt: time.Now().Add(accountHealthAutoConfigCacheTTL).UnixNano(),
	})
	return cfg, nil
}

func (s *SettingService) SaveAccountHealthAutoCheckConfig(ctx context.Context, cfg *AccountHealthAutoCheckConfig) error {
	cfg = normalizeAccountHealthAutoCheckConfig(cfg)
	if err := validateAccountHealthAutoCheckConfig(cfg); err != nil {
		return infraerrors.BadRequest("INVALID_ACCOUNT_HEALTH_AUTO_CHECK_CONFIG", err.Error())
	}
	existing, _ := s.GetAccountHealthAutoCheckConfig(ctx)
	if existing != nil {
		cfg.LastRunAt = existing.LastRunAt
	}
	return s.storeAccountHealthAutoCheckConfig(ctx, cfg)
}

func (s *SettingService) MarkAccountHealthAutoCheckRun(ctx context.Context, runAt time.Time) error {
	cfg, err := s.GetAccountHealthAutoCheckConfig(ctx)
	if err != nil {
		cfg = defaultAccountHealthAutoCheckConfig()
	}
	cfg = normalizeAccountHealthAutoCheckConfig(cfg)
	ts := runAt.Unix()
	cfg.LastRunAt = &ts
	return s.storeAccountHealthAutoCheckConfig(ctx, cfg)
}

func (s *SettingService) storeAccountHealthAutoCheckConfig(ctx context.Context, cfg *AccountHealthAutoCheckConfig) error {
	data, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("account health auto check: marshal config: %w", err)
	}
	if err := s.settingRepo.Set(ctx, SettingKeyAccountHealthAutoCheckConfig, string(data)); err != nil {
		return fmt.Errorf("account health auto check: save config: %w", err)
	}
	accountHealthAutoCheckSF.Forget(sfKeyAccountHealthAutoCheckConfig)
	accountHealthAutoCheckCache.Store(&cachedAccountHealthAutoCheckConfig{
		config:    cfg,
		expiresAt: time.Now().Add(accountHealthAutoConfigCacheTTL).UnixNano(),
	})
	return nil
}
