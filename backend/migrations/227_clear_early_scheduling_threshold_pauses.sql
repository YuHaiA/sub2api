-- Percentage-based scheduling thresholds are no longer allowed to pause an
-- account before its quota reaches 100%. Clear only historical threshold
-- pauses that were recorded below exhaustion; real quota and error cooldowns
-- use different state/reasons and remain untouched.
UPDATE accounts
SET temp_unschedulable_until = NULL,
    temp_unschedulable_reason = NULL,
    updated_at = NOW()
WHERE temp_unschedulable_until IS NOT NULL
  AND temp_unschedulable_reason LIKE '%"source":"account_scheduling_threshold"%'
  AND COALESCE(
        NULLIF(
          substring(temp_unschedulable_reason FROM '"used_percent"[[:space:]]*:[[:space:]]*([0-9]+([.][0-9]+)?)'),
          ''
        )::numeric,
        0
      ) < 100;
