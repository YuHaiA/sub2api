package service

import (
	"testing"
	"time"
)

func TestNormalizeAccountTokenAutoRefreshConfig(t *testing.T) {
	cfg := normalizeAccountTokenAutoRefreshConfig(&AccountTokenAutoRefreshConfig{
		IntervalValue: 0,
		IntervalUnit:  "WEEK",
		BatchSize:     0,
		Scope:         "bad",
		GroupID:       99,
		HealthStatus:  "weird",
	})

	if cfg.IntervalValue != DefaultTokenRefreshIntervalValue {
		t.Fatalf("expected default interval value, got %d", cfg.IntervalValue)
	}
	if cfg.IntervalUnit != DefaultTokenRefreshIntervalUnit {
		t.Fatalf("expected default interval unit, got %s", cfg.IntervalUnit)
	}
	if cfg.BatchSize != DefaultTokenRefreshBatchSize {
		t.Fatalf("expected default batch size, got %d", cfg.BatchSize)
	}
	if cfg.Scope != "all" {
		t.Fatalf("expected default scope all, got %s", cfg.Scope)
	}
	if cfg.GroupID != 0 {
		t.Fatalf("expected group id reset to 0, got %d", cfg.GroupID)
	}
	if cfg.HealthStatus != "" {
		t.Fatalf("expected health status reset to empty, got %q", cfg.HealthStatus)
	}
}

func TestValidateAccountTokenAutoRefreshConfig(t *testing.T) {
	if err := validateAccountTokenAutoRefreshConfig(&AccountTokenAutoRefreshConfig{
		IntervalValue: 1,
		IntervalUnit:  "hour",
		BatchSize:     10,
		Scope:         "all",
	}); err != nil {
		t.Fatalf("expected valid config, got %v", err)
	}

	if err := validateAccountTokenAutoRefreshConfig(&AccountTokenAutoRefreshConfig{
		IntervalValue: 1,
		IntervalUnit:  "day",
		BatchSize:     10,
		Scope:         "group",
		GroupID:       8,
		HealthStatus:  "constrained",
	}); err != nil {
		t.Fatalf("expected valid group config, got %v", err)
	}

	if err := validateAccountTokenAutoRefreshConfig(&AccountTokenAutoRefreshConfig{
		IntervalValue: 0,
		IntervalUnit:  "hour",
		BatchSize:     10,
	}); err == nil {
		t.Fatal("expected interval validation error")
	}

	if err := validateAccountTokenAutoRefreshConfig(&AccountTokenAutoRefreshConfig{
		IntervalValue: 1,
		IntervalUnit:  "hour",
		BatchSize:     10,
		Scope:         "group",
		GroupID:       0,
	}); err == nil {
		t.Fatal("expected group validation error")
	}

	if err := validateAccountTokenAutoRefreshConfig(&AccountTokenAutoRefreshConfig{
		IntervalValue: 1,
		IntervalUnit:  "hour",
		BatchSize:     10,
		Scope:         "all",
		HealthStatus:  "bad",
	}); err == nil {
		t.Fatal("expected health status validation error")
	}
}

func TestAccountTokenAutoRefreshDuration(t *testing.T) {
	hourCfg := &AccountTokenAutoRefreshConfig{IntervalValue: 6, IntervalUnit: "hour", BatchSize: 10}
	dayCfg := &AccountTokenAutoRefreshConfig{IntervalValue: 2, IntervalUnit: "day", BatchSize: 10}

	if got := accountTokenAutoRefreshDuration(hourCfg); got != 6*time.Hour {
		t.Fatalf("expected 6h, got %s", got)
	}
	if got := accountTokenAutoRefreshDuration(dayCfg); got != 48*time.Hour {
		t.Fatalf("expected 48h, got %s", got)
	}
}
