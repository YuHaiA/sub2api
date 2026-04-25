package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAutoHealthStatusFromResult_UsesCurrentBuckets(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		result   *ScheduledTestResult
		expected string
	}{
		{
			name:     "nil result stays unchecked",
			result:   nil,
			expected: "unchecked",
		},
		{
			name: "success stays healthy",
			result: &ScheduledTestResult{
				Status: "success",
			},
			expected: "healthy",
		},
		{
			name: "429 becomes constrained",
			result: &ScheduledTestResult{
				Status:       "failed",
				ErrorMessage: "API returned 429: too many requests",
			},
			expected: "constrained",
		},
		{
			name: "402 becomes constrained",
			result: &ScheduledTestResult{
				Status:       "failed",
				ErrorMessage: "payment required (402)",
			},
			expected: "constrained",
		},
		{
			name: "banned becomes unavailable",
			result: &ScheduledTestResult{
				Status:       "failed",
				ErrorMessage: "account banned due to violation",
			},
			expected: "unavailable",
		},
		{
			name: "generic failure becomes unavailable",
			result: &ScheduledTestResult{
				Status:       "failed",
				ErrorMessage: "dial tcp timeout",
			},
			expected: "unavailable",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.expected, autoHealthStatusFromResult(tc.result))
		})
	}
}
