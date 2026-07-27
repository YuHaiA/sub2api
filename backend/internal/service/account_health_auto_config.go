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
	accountHealthAutoConfigRunningTTL   = 2 * time.Second
	accountHealthAutoConfigErrorTTL     = 5 * time.Second
	accountHealthAutoConfigDBTimeout    = 5 * time.Second
	accountHealthAutoConfigStaleAfter   = 20 * time.Minute
	DefaultAccountHealthIntervalMinutes = 60
	maxAccountHealthIntervalMinutes     = 24 * 60
)

type AccountHealthAutoCheckConfig struct {
	Enabled           bool   `json:"enabled"`
	IntervalMinutes   int    `json:"interval_minutes"`
	ModelID           string `json:"model_id"`
	Running           bool   `json:"running,omitempty"`
	CurrentTotal      int    `json:"current_total,omitempty"`
	CurrentSuccess    int    `json:"current_success,omitempty"`
	CurrentFailed     int    `json:"current_failed,omitempty"`
	QueueRunning      string `json:"queue_running,omitempty"`
	QueuePending      string `json:"queue_pending,omitempty"`
	LastRunAt         *int64 `json:"last_run_at,omitempty"`
	ProgressUpdatedAt *int64 `json:"progress_updated_at,omitempty"`
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
			cfg := s.withAccountHealthQueueSnapshot(c.config)
			return s.clearStaleAccountHealthAutoCheckRunning(cfg), nil
		}
	}
	result, err, _ := accountHealthAutoCheckSF.Do(sfKeyAccountHealthAutoCheckConfig, func() (any, error) {
		return s.loadAccountHealthAutoCheckConfigFromDB()
	})
	if err != nil {
		return defaultAccountHealthAutoCheckConfig(), err
	}
	if cfg, ok := result.(*AccountHealthAutoCheckConfig); ok {
		cfg = s.withAccountHealthQueueSnapshot(cfg)
		return s.clearStaleAccountHealthAutoCheckRunning(cfg), nil
	}
	return defaultAccountHealthAutoCheckConfig(), nil
}

func (s *SettingService) withAccountHealthQueueSnapshot(cfg *AccountHealthAutoCheckConfig) *AccountHealthAutoCheckConfig {
	if cfg == nil {
		return defaultAccountHealthAutoCheckConfig()
	}
	out := *cfg
	queue := GetBackgroundMaintenanceSnapshot()
	out.QueueRunning = queue.Running
	out.QueuePending = queue.Pending
	return &out
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
	cfg = s.clearStaleAccountHealthAutoCheckRunning(cfg)
	ttl := accountHealthAutoConfigCacheTTL
	if cfg != nil && cfg.Running {
		ttl = accountHealthAutoConfigRunningTTL
	}
	accountHealthAutoCheckCache.Store(&cachedAccountHealthAutoCheckConfig{
		config:    cfg,
		expiresAt: time.Now().Add(ttl).UnixNano(),
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
		cfg.Running = existing.Running
		cfg.CurrentTotal = existing.CurrentTotal
		cfg.CurrentSuccess = existing.CurrentSuccess
		cfg.CurrentFailed = existing.CurrentFailed
		cfg.QueueRunning = existing.QueueRunning
		cfg.QueuePending = existing.QueuePending
		cfg.LastRunAt = existing.LastRunAt
		cfg.ProgressUpdatedAt = existing.ProgressUpdatedAt
	}
	return s.storeAccountHealthAutoCheckConfig(ctx, cfg)
}

func (s *SettingService) MarkAccountHealthAutoCheckRun(ctx context.Context, runAt time.Time) error {
	cfg, err := s.GetAccountHealthAutoCheckConfig(ctx)
	if err != nil {
		cfg = defaultAccountHealthAutoCheckConfig()
	}
	cfg = normalizeAccountHealthAutoCheckConfig(cfg)
	queue := GetBackgroundMaintenanceSnapshot()
	cfg.QueueRunning = queue.Running
	cfg.QueuePending = queue.Pending
	ts := runAt.Unix()
	cfg.Running = false
	cfg.CurrentTotal = 0
	cfg.CurrentSuccess = 0
	cfg.CurrentFailed = 0
	cfg.ProgressUpdatedAt = nil
	cfg.LastRunAt = &ts
	return s.storeAccountHealthAutoCheckConfig(ctx, cfg)
}

func (s *SettingService) MarkAccountHealthAutoCheckProgress(
	ctx context.Context,
	total int,
	success int,
	failed int,
) error {
	cfg, err := s.GetAccountHealthAutoCheckConfig(ctx)
	if err != nil {
		cfg = defaultAccountHealthAutoCheckConfig()
	}
	cfg = normalizeAccountHealthAutoCheckConfig(cfg)
	queue := GetBackgroundMaintenanceSnapshot()
	cfg.QueueRunning = queue.Running
	cfg.QueuePending = queue.Pending
	cfg.Running = true
	cfg.CurrentTotal = total
	cfg.CurrentSuccess = success
	cfg.CurrentFailed = failed
	ts := time.Now().Unix()
	cfg.ProgressUpdatedAt = &ts
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
	ttl := accountHealthAutoConfigCacheTTL
	if cfg != nil && cfg.Running {
		ttl = accountHealthAutoConfigRunningTTL
	}
	accountHealthAutoCheckCache.Store(&cachedAccountHealthAutoCheckConfig{
		config:    cfg,
		expiresAt: time.Now().Add(ttl).UnixNano(),
	})
	return nil
}

// clearStaleAccountHealthAutoCheckRunning drops a stuck "running" flag when progress
// has not advanced for a long time (process crash, hung upstream, deploy restart).
func (s *SettingService) clearStaleAccountHealthAutoCheckRunning(cfg *AccountHealthAutoCheckConfig) *AccountHealthAutoCheckConfig {
	if cfg == nil || !cfg.Running {
		return cfg
	}

	queue := GetBackgroundMaintenanceSnapshot()
	healthQueueActive := queue.Running == "account_health_manual" ||
		queue.Running == "account_health_auto" ||
		queue.Pending == "account_health_manual" ||
		queue.Pending == "account_health_auto"

	stale := false
	switch progressUpdatedAt := cfg.ProgressUpdatedAt; {
	case progressUpdatedAt == nil:
		// Legacy stuck "running" without progress heartbeat: only clear when this
		// process is not currently executing/queueing a health task.
		stale = !healthQueueActive
	default:
		updatedAt := time.Unix(*progressUpdatedAt, 0)
		stale = time.Since(updatedAt) >= accountHealthAutoConfigStaleAfter && !healthQueueActive
	}
	if !stale {
		return cfg
	}

	cleared := *cfg
	cleared.Running = false
	cleared.CurrentTotal = 0
	cleared.CurrentSuccess = 0
	cleared.CurrentFailed = 0
	cleared.ProgressUpdatedAt = nil
	cleared.QueueRunning = ""
	cleared.QueuePending = ""
	go func() {
		_ = s.storeAccountHealthAutoCheckConfig(context.Background(), &cleared)
	}()
	return &cleared
}
