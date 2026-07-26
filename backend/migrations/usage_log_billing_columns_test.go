package migrations

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUsageLogImageInputAndLongContextBillingMigration(t *testing.T) {
	content, err := FS.ReadFile("191_add_usage_log_image_input_and_long_context_billing.sql")
	require.NoError(t, err)

	sql := strings.ToLower(string(content))
	require.Contains(t, sql, "add column if not exists image_input_tokens")
	require.Contains(t, sql, "add column if not exists image_input_cost")
	require.Contains(t, sql, "add column if not exists long_context_billing_applied")
	require.Contains(t, sql, "integer not null default 0")
	require.Contains(t, sql, "decimal(20, 10) not null default 0")
	require.Contains(t, sql, "boolean not null default false")
}
