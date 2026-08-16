-- Older Mihomo/import flows stored standby state in schedulable=false.
-- Restore active Grok OAuth accounts unless the new explicit administrator
-- marker says that scheduling was intentionally disabled. Runtime quota and
-- error state live in their dedicated columns/status and remain untouched.
UPDATE accounts
SET schedulable = TRUE,
    updated_at = NOW()
WHERE deleted_at IS NULL
  AND platform = 'grok'
  AND type = 'oauth'
  AND status = 'active'
  AND schedulable = FALSE
  AND COALESCE(extra->>'manual_schedulable_disabled', 'false') <> 'true';
