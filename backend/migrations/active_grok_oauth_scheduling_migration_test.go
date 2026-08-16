package migrations

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRestoreActiveGrokOAuthSchedulingMigrationPreservesExplicitManualPause(t *testing.T) {
	raw, err := FS.ReadFile("228_restore_active_grok_oauth_scheduling.sql")
	require.NoError(t, err)

	sql := strings.ToLower(string(raw))
	require.Contains(t, sql, "platform = 'grok'")
	require.Contains(t, sql, "type = 'oauth'")
	require.Contains(t, sql, "status = 'active'")
	require.Contains(t, sql, "manual_schedulable_disabled")
	require.NotContains(t, sql, "rate_limit_reset_at = null")
	require.NotContains(t, sql, "temp_unschedulable_until = null")
}
