package admin

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestClassifyAccountHealthStatus_MapsConstrainedAndUnavailable(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		result   *service.ScheduledTestResult
		expected string
	}{
		{
			name: "success stays healthy",
			result: &service.ScheduledTestResult{
				Status: "success",
			},
			expected: accountHealthStatusHealthy,
		},
		{
			name: "429 becomes constrained",
			result: &service.ScheduledTestResult{
				Status:       "failed",
				ErrorMessage: "API returned 429: too many requests",
			},
			expected: accountHealthStatusConstrained,
		},
		{
			name: "402 becomes constrained",
			result: &service.ScheduledTestResult{
				Status:       "failed",
				ErrorMessage: "payment required (402)",
			},
			expected: accountHealthStatusConstrained,
		},
		{
			name: "banned becomes unavailable",
			result: &service.ScheduledTestResult{
				Status:       "failed",
				ErrorMessage: "account banned due to violation",
			},
			expected: accountHealthStatusUnavailable,
		},
		{
			name: "generic failure becomes unavailable",
			result: &service.ScheduledTestResult{
				Status:       "failed",
				ErrorMessage: "dial tcp timeout",
			},
			expected: accountHealthStatusUnavailable,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.expected, classifyAccountHealthStatus(tc.result))
		})
	}
}

func TestParseStoredAccountHealth_NormalizesLegacyStatuses(t *testing.T) {
	t.Parallel()

	quotaAccount := &service.Account{
		Extra: map[string]any{
			accountHealthCheckExtraKey: map[string]any{
				"status":  "banned_or_exhausted",
				"message": "insufficient quota (402)",
			},
		},
	}
	bannedAccount := &service.Account{
		Extra: map[string]any{
			accountHealthCheckExtraKey: map[string]any{
				"status":  "banned_or_exhausted",
				"message": "account banned",
			},
		},
	}
	rateLimitedAccount := &service.Account{
		Extra: map[string]any{
			accountHealthCheckExtraKey: map[string]any{
				"status":  "rate_limited",
				"message": "API returned 429",
			},
		},
	}

	require.Equal(t, accountHealthStatusConstrained, parseStoredAccountHealth(quotaAccount).Status)
	require.Equal(t, accountHealthStatusUnavailable, parseStoredAccountHealth(bannedAccount).Status)
	require.Equal(t, accountHealthStatusConstrained, parseStoredAccountHealth(rateLimitedAccount).Status)
}

func TestSummarizeAccountHealthSnapshots_UsesNewBuckets(t *testing.T) {
	t.Parallel()

	summary := summarizeAccountHealthSnapshots([]accountHealthSnapshot{
		{Status: accountHealthStatusHealthy},
		{Status: accountHealthStatusConstrained},
		{Status: accountHealthStatusUnavailable},
		{Status: accountHealthStatusUnchecked},
	})

	require.Equal(t, 4, summary.TotalAccounts)
	require.Equal(t, 1, summary.HealthyAccounts)
	require.Equal(t, 1, summary.ConstrainedAccounts)
	require.Equal(t, 1, summary.UnavailableAccounts)
	require.Equal(t, 1, summary.UncheckedAccounts)
}
