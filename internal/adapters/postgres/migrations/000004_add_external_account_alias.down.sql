-- 000004_add_external_account_alias.down.sql

ALTER TABLE receivables_accounts
DROP COLUMN IF EXISTS external_account_alias;
