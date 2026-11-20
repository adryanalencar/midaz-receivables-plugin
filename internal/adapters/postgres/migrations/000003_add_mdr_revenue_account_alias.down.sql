-- 000003_add_mdr_revenue_account_alias.down.sql

DROP INDEX IF EXISTS idx_receivables_accounts_mdr_revenue;

ALTER TABLE receivables_accounts
DROP COLUMN IF EXISTS mdr_revenue_account_alias;
