-- 000004_add_external_account_alias.up.sql

ALTER TABLE receivables_accounts
ADD COLUMN IF NOT EXISTS external_account_alias VARCHAR(255);
