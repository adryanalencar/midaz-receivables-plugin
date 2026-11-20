# Implementation Plan: Midaz Receivables Plugin

## 1. Overview & Phased Roadmap

This roadmap breaks down the development of `midaz-plugin-receivables` into 6 disciplined, test-driven phases. Every line of code, entity, database table, API route, and test is authored strictly in **English**.

---

## 2. Phase Breakdown

### Phase 1: Core Domain, Entity Modeling, Encryption & PostgreSQL Migrations
- **Deliverables:**
  - `go.mod` module initialization (`github.com/LerianStudio/midaz-plugin-receivables`).
  - Core domain models in `internal/domain/`:
    - `ReceivablesAccount`: Midaz account wallet representation (`@<alias>/receivables`).
    - `AcquirerConnection`: Credentials and synchronization state for an acquirer.
    - `Sale` and `Installment`: Normalized card transactions, fees, and settlement dates.
    - `SettlementMatch`: Matcher 6 matched pair representation.
  - Domain Ports: `AcquirerProvider`, `MidazLedgerClient`, `MatcherClient`, `ReceivablesRepository`.
  - Cryptographic engine in `internal/crypto/encryptor.go`:
    - AES-256-GCM encryption and decryption with 12-byte random nonces and 16-byte authentication tags.
    - Base64 packing/unpacking and fail-fast master key verification.
  - PostgreSQL schema migration files: `000001_init_schema.up.sql` and `000001_init_schema.down.sql`.
  - Database migrator and repository implementation (`internal/adapters/postgres/repository.go`).
- **Validation:** Cryptographic unit tests (tamper resistance, non-repeatable nonces, round-trip encrypt/decrypt) and repository unit tests using `go-sqlmock` validating account and connection CRUD, upserts, and idempotency.

### Phase 2: Acquirer Provider Adapters (SumUp & PagBank)
- **Deliverables:**
  - `internal/adapters/providers/factory.go`: Provider registry resolving acquirer adapters by type.
  - `internal/adapters/providers/sumup/`:
    - REST client pulling sales history and payout schedules with connection credentials.
    - Inbound webhook parser with HMAC signature verification.
  - `internal/adapters/providers/pagbank/`:
    - REST client pulling orders, charges, and settlement statements.
    - Inbound webhook parser with token verification.
- **Validation:** Unit tests for parsers with real anonymized JSON fixtures from SumUp and PagBank.

### Phase 3: Midaz Double-Entry Ledger Client & Account Provisioning
- **Deliverables:**
  - `internal/adapters/http/out/midaz_client.go`: High-performance HTTP client for Midaz Onboarding & Transaction APIs.
  - `internal/services/command/ensure_acquirer_accounts.go`: Idempotent provisioning of universal `@external/acquirer_*` accounts.
  - `internal/services/command/create_receivables_account.go`: Provisions the merchant's receivables wallet in Midaz.
  - `internal/services/command/ingest_acquirer_sale.go`: Double-entry booking of D+0 sales:
    $$\texttt{@external/acquirer\_<name>} \longrightarrow \texttt{@<alias>/receivables}$$
- **Validation:** Integration test asserting balanced debits and credits on Midaz transaction generation.

### Phase 4: Matcher 6 Reconciliation Bridge
- **Deliverables:**
  - `internal/adapters/http/out/matcher_client.go`: Client for Matcher 6 (`:4018`):
    - Context & Sources provisioning (`acquirers-settlement`).
    - Source A batch upload (`POST /v1/imports/contexts/{ctxId}/sources/{srcId}/upload`).
    - Triggering reconciliation runs (`POST /v1/contexts/{ctxId}/reconciliations/runs`).
  - `internal/workers/matcher_upload_worker.go`: Background worker periodically uploading pending installments.
- **Validation:** Live API integration tests against local Matcher 6 running on `http://localhost:4018`.

### Phase 5: Settlement Webhooks & Inter-Plugin Coordination
- **Deliverables:**
  - `internal/adapters/http/in/matcher_webhook_handler.go`: Webhook receiver for Matcher 6 match approvals.
  - `internal/services/command/settle_matched_receivable.go`: Atomic execution of settlement transfer in Midaz:
    $$\texttt{@<alias>/receivables} \longrightarrow \texttt{@<alias>/checking}$$
  - Marking installment status as `RECONCILED` with `matcher_run_id` and `midaz_settlement_tx_id`.
- **Validation:** End-to-end integration test: Account created -> Connection linked -> Sale ingested -> Schedule uploaded -> Matcher webhook received -> Settlement posted in Midaz.

### Phase 6: Midaz Console Manifest & Production Packaging
- **Deliverables:**
  - `internal/adapters/http/in/manifest_handler.go`: Dynamic `/api/manifest` microfrontend descriptor.
  - Health & readiness probes (`/health`, `/ready`).
  - Multi-stage production `Dockerfile` (distroless / static binary).
  - Complete `docker-compose.yml` linking Plugin Receivables, PostgreSQL, Midaz, and Matcher 6.
- **Validation:** Full stack startup via `docker compose up -d` with automated health checks passing.
