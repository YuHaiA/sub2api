-- Mihomo pool members used to encode standby state with schedulable=false,
-- which left only the account bound to the active egress eligible for routing.
-- Pool members are now independently schedulable; quota/rate-limit/temporary
-- cooldown fields continue to remove only the affected account until reset.
UPDATE accounts
SET schedulable = TRUE,
    updated_at = NOW()
WHERE deleted_at IS NULL
  AND platform = 'grok'
  AND status = 'active'
  AND schedulable = FALSE
  AND COALESCE(extra->>'mihomo_pool_managed', 'false') = 'true';
