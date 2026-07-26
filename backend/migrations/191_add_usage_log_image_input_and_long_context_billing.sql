-- Persist image-input billing details and whether the long-context pricing
-- branch was applied to a usage record. Defaults preserve existing rows and
-- keep the migration safe for rolling deployments.
ALTER TABLE usage_logs
    ADD COLUMN IF NOT EXISTS image_input_tokens INTEGER NOT NULL DEFAULT 0;

ALTER TABLE usage_logs
    ADD COLUMN IF NOT EXISTS image_input_cost DECIMAL(20, 10) NOT NULL DEFAULT 0;

ALTER TABLE usage_logs
    ADD COLUMN IF NOT EXISTS long_context_billing_applied BOOLEAN NOT NULL DEFAULT FALSE;
