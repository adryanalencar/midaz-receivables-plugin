-- 000001_init_schema.up.sql

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'connection_status') THEN
        CREATE TYPE connection_status AS ENUM (
            'ACTIVE',
            'INACTIVE',
            'EXPIRED',
            'ERROR'
        );
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'installment_status') THEN
        CREATE TYPE installment_status AS ENUM (
            'PENDING_SETTLEMENT',
            'RECONCILED',
            'CANCELLED',
            'CHARGEBACK',
            'OVERDUE'
        );
    END IF;
END $$;

-- Receivables Account (1 per merchant/tenant)
CREATE TABLE IF NOT EXISTS receivables_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL,
    ledger_id UUID NOT NULL,
    midaz_account_id UUID NOT NULL,
    account_alias VARCHAR(255) NOT NULL UNIQUE,
    checking_account_alias VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Acquirer Connections (N per Receivables Account)
CREATE TABLE IF NOT EXISTS acquirer_connections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    receivables_account_id UUID NOT NULL REFERENCES receivables_accounts(id) ON DELETE CASCADE,
    acquirer_type VARCHAR(50) NOT NULL,       -- 'SUMUP', 'PAGBANK'
    display_name VARCHAR(100) NOT NULL,
    credentials_encrypted TEXT NOT NULL,      -- AES-256-GCM encrypted base64 ciphertext
    status connection_status NOT NULL DEFAULT 'ACTIVE',
    auto_sync_enabled BOOLEAN NOT NULL DEFAULT true,
    sync_interval_minutes INT NOT NULL DEFAULT 60,
    last_sync_at TIMESTAMP WITH TIME ZONE,
    last_sync_status VARCHAR(50),
    last_sync_error TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_account_acquirer UNIQUE (receivables_account_id, acquirer_type, display_name)
);

-- Ingested Acquirer Sales
CREATE TABLE IF NOT EXISTS acquirer_sales (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    connection_id UUID NOT NULL REFERENCES acquirer_connections(id) ON DELETE CASCADE,
    receivables_account_id UUID NOT NULL REFERENCES receivables_accounts(id) ON DELETE CASCADE,
    acquirer VARCHAR(50) NOT NULL,           -- 'SUMUP', 'PAGBANK'
    nsu VARCHAR(100) NOT NULL,
    tid VARCHAR(100),
    authorization_code VARCHAR(100),
    sale_date TIMESTAMP WITH TIME ZONE NOT NULL,
    gross_amount NUMERIC(18, 4) NOT NULL,
    total_mdr_fee NUMERIC(18, 4) NOT NULL,
    total_net_amount NUMERIC(18, 4) NOT NULL,
    installments_count INT NOT NULL DEFAULT 1,
    payment_method VARCHAR(50) NOT NULL,    -- 'CREDIT_CARD', 'DEBIT_CARD', 'PIX'
    card_brand VARCHAR(50),                 -- 'VISA', 'MASTERCARD', 'ELO'
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_connection_sale UNIQUE (connection_id, nsu)
);

-- Installments / Receivables Schedule
CREATE TABLE IF NOT EXISTS receivables_installments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sale_id UUID NOT NULL REFERENCES acquirer_sales(id) ON DELETE CASCADE,
    connection_id UUID NOT NULL REFERENCES acquirer_connections(id) ON DELETE CASCADE,
    receivables_account_id UUID NOT NULL REFERENCES receivables_accounts(id) ON DELETE CASCADE,
    acquirer VARCHAR(50) NOT NULL,
    nsu VARCHAR(100) NOT NULL,
    installment_number INT NOT NULL,
    gross_amount NUMERIC(18, 4) NOT NULL,
    mdr_fee NUMERIC(18, 4) NOT NULL,
    net_amount NUMERIC(18, 4) NOT NULL,
    expected_settlement_date DATE NOT NULL,
    actual_settlement_date DATE,
    status installment_status NOT NULL DEFAULT 'PENDING_SETTLEMENT',
    midaz_receivable_tx_id UUID,            -- Transaction ID of the D+0 agenda booking
    midaz_settlement_tx_id UUID,            -- Transaction ID of the D+X settlement transfer
    matcher_record_id VARCHAR(255),         -- Ingestion ID pushed to Matcher 6
    matcher_run_id UUID,                    -- ID of Matcher 6 execution confirming match
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_connection_installment UNIQUE (connection_id, nsu, installment_number)
);

CREATE INDEX IF NOT EXISTS idx_installments_expected_date ON receivables_installments(expected_settlement_date, status);
CREATE INDEX IF NOT EXISTS idx_installments_account ON receivables_installments(receivables_account_id, status);
CREATE INDEX IF NOT EXISTS idx_connections_sync ON acquirer_connections(auto_sync_enabled, status);
