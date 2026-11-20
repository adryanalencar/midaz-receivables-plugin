-- 000003_add_mdr_revenue_account_alias.up.sql

ALTER TABLE receivables_accounts
ADD COLUMN IF NOT EXISTS mdr_revenue_account_alias VARCHAR(255);

CREATE INDEX IF NOT EXISTS idx_receivables_accounts_mdr_revenue ON receivables_accounts(mdr_revenue_account_alias);
