package migrations

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnableMihomoPoolAccountsMigrationPreservesRuntimeCooldowns(t *testing.T) {
	raw, err := FS.ReadFile("224_enable_mihomo_pool_accounts.sql")
	require.NoError(t, err)

	sql := strings.ToLower(string(raw))
	require.Contains(t, sql, "set schedulable = true")
	require.Contains(t, sql, "extra->>'mihomo_pool_managed'")
	require.Contains(t, sql, "status = 'active'")
	require.NotContains(t, sql, "rate_limit_reset_at = null")
	require.NotContains(t, sql, "temp_unschedulable_until = null")
	require.NotContains(t, sql, "overload_until = null")
}
