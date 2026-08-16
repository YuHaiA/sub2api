package migrations

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClearEarlySchedulingThresholdPausesMigrationPreservesRealCooldowns(t *testing.T) {
	raw, err := FS.ReadFile("227_clear_early_scheduling_threshold_pauses.sql")
	require.NoError(t, err)

	sql := strings.ToLower(string(raw))
	require.Contains(t, sql, "account_scheduling_threshold")
	require.Contains(t, sql, "used_percent")
	require.Contains(t, sql, ") < 100")
	require.Contains(t, sql, "temp_unschedulable_until = null")
	require.Contains(t, sql, "temp_unschedulable_reason = null")
	require.NotContains(t, sql, "rate_limit_reset_at = null")
	require.NotContains(t, sql, "overload_until = null")
}
