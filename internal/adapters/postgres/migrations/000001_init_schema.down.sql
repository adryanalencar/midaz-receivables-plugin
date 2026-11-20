-- 000001_init_schema.down.sql

DROP TABLE IF EXISTS receivables_installments CASCADE;
DROP TABLE IF EXISTS acquirer_sales CASCADE;
DROP TABLE IF EXISTS acquirer_connections CASCADE;
DROP TABLE IF EXISTS receivables_accounts CASCADE;

DROP TYPE IF EXISTS installment_status;
DROP TYPE IF EXISTS connection_status;
