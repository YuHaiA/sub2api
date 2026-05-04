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
}

func TestValidateAccountTokenAutoRefreshConfig(t *testing.T) {
	if err := validateAccountTokenAutoRefreshConfig(&AccountTokenAutoRefreshConfig{
		IntervalValue: 1,
		IntervalUnit:  "hour",
		BatchSize:     10,
	}); err != nil {
		t.Fatalf("expected valid config, got %v", err)
	}

	if err := validateAccountTokenAutoRefreshConfig(&AccountTokenAutoRefreshConfig{
		IntervalValue: 0,
		IntervalUnit:  "hour",
		BatchSize:     10,
	}); err == nil {
		t.Fatal("expected interval validation error")
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
