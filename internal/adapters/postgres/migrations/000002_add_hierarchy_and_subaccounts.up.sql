-- 000002_add_hierarchy_and_subaccounts.up.sql

ALTER TABLE receivables_accounts
ADD COLUMN IF NOT EXISTS parent_account_id UUID,
ADD COLUMN IF NOT EXISTS mother_account_alias VARCHAR(255),
ADD COLUMN IF NOT EXISTS fee_account_alias VARCHAR(255),
ADD COLUMN IF NOT EXISTS matcher_context_id VARCHAR(255);

CREATE INDEX IF NOT EXISTS idx_receivables_accounts_parent ON receivables_accounts(parent_account_id);
CREATE INDEX IF NOT EXISTS idx_receivables_accounts_context ON receivables_accounts(matcher_context_id);
