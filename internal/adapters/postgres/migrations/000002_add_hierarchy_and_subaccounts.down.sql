-- 000002_add_hierarchy_and_subaccounts.down.sql

DROP INDEX IF EXISTS idx_receivables_accounts_context;
DROP INDEX IF EXISTS idx_receivables_accounts_parent;

ALTER TABLE receivables_accounts
DROP COLUMN IF EXISTS matcher_context_id,
DROP COLUMN IF EXISTS fee_account_alias,
DROP COLUMN IF EXISTS mother_account_alias,
DROP COLUMN IF EXISTS parent_account_id;
