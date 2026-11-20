# Product Requirements Document (PRD): Midaz Receivables Plugin

## 1. Problem Statement & Context
Fintechs, marketplaces, sub-acquirers, and merchant aggregators utilizing **Midaz** as their immutable core ledger frequently operate across multiple legal entities, merchant tenants, and corporate tax IDs (CNPJs). 

When managing credit card receivables (**Agenda de Recebíveis**) and bank account settlements across hundreds or thousands of distinct merchants, two architectural imperatives arise:
1. **Multi-Tenant Acquirer Credential Isolation:** Each merchant or business unit possesses distinct credentials with payment acquirers (e.g., SumUp, PagBank). A merchant may hold accounts across multiple acquirers simultaneously. The platform must provision a designated **Receivables Account** in Midaz for each merchant and dynamically associate multiple isolated **Acquirer Connections** to that specific account.
2. **Elimination of Balance Duplication:** Card sales are captured on D+0, establishing an accrued credit right. When the acquirer executes the bank payout on D+30, standard bank statement integrations (such as Open Finance) mirror the deposit as new cash, causing a duplicate balance unless coordinated through automated reconciliation with **Matcher 6**.

---

## 2. Product Objectives
The **Midaz Receivables Plugin (`midaz-plugin-receivables`)** provides an autonomous, multi-tenant settlement and receivables orchestration service:
1. **Autonomous Receivables Account Provisioning:** Automatically provisions dedicated merchant receivables accounts in Midaz (`@<alias>/receivables`) and associates arbitrary acquirer connections to them.
2. **Multi-Credential Acquirer Aggregation:** Allows a single Receivables Account to bind multiple acquirer connections (e.g., SumUp Connection with Merchant A's credentials, PagBank Connection with Merchant A's credentials) without coupling the core domain to tax IDs.
3. **Double-Entry Receivables Ingestion:** Ingests sales and future payout schedules into Midaz using balanced double-entry transactions counterbalanced by universal clearing accounts (`@external/acquirer_<name>`).
4. **Three-Way Settlement Reconciliation with Matcher 6 & Open Finance:** Interoperates with `midaz-plugin-openfinance` and the local **Matcher 6** engine (`http://localhost:4018`) to ensure that bank credits extinguish accrued receivables rights into the merchant's liquid checking account (`@<alias>/checking`), guaranteeing zero balance inflation.
5. **Strict English Standardization:** All domain models, APIs, database schemas, code, logs, and technical specifications adhere strictly to English nomenclature.

---

## 3. Scope Definition

### 3.1. In-Scope
- **Account & Connection Governance:**
  - Automated provisioning of Midaz receivables accounts (`type: asset`).
  - Secure storage and association of acquirer credentials to receivables accounts (1 Account : $N$ Connections).
- **Multi-Acquirer Connectors:**
  - **SumUp:** Ingestion of sales, installments, payout calendars, and fee breakdowns via REST API and Webhooks.
  - **PagBank (PagSeguro):** Ingestion of orders, charges, and settlement schedules via REST API and Webhooks.
- **Double-Entry Booking (Midaz v3.4.8+):**
  - **Phase 1 (Sale Ingestion D+0):** Debit `@external/acquirer_<name>` (Gross), Credit `@<alias>/receivables` (Net), Credit `@platform/mdr_fee` (Spread).
  - **Phase 2 (Settlement D+X):** Transfer from `@<alias>/receivables` to `@<alias>/checking` upon Matcher 6 approval.
  - **Phase 3 (Disputes / Chargebacks):** Reversal or debiting of receivables balance.
- **Matcher 6 Integration (v6.0.0-beta.1):**
  - Push Source A (Expected Receivables Schedule) to Matcher 6.
  - Process Matcher 6 webhook callbacks and execute atomic settlement transfers.
- **Inter-Plugin Protocol with Open Finance:**
  - Interception of bank deposits flagged with `reconciliation.enabled == true`.

### 3.2. Out-of-Scope (V1)
- **Direct Card Pinpad / POS Hardware Communication:** Strictly ingests finalized acquirer sales and payout schedules.
- **Bacen CERC/CIP Trade Repository Registration:** Scheduled for V2.

---

## 4. User Personas

| Persona | Role | Primary Goal |
| :--- | :--- | :--- |
| **Merchant Administrator** | Manages merchant onboarding and financial operations | Needs to easily connect multiple acquirers (SumUp, PagBank) to a single receivables account and monitor expected cash inflows. |
| **Financial Controller / Auditor** | Oversees balance sheet accuracy and regulatory compliance | Demands mathematical proof that every bank deposit matches individual card installments minus contracted MDR fees. |
| **Platform Engineer** | Operates Midaz microservices infrastructure | Requires a stateless, containerized Go service configurable via environment variables with zero downtime migrations. |

---

## 5. Functional Requirements (FR)

### FR-01: Receivables Account Provisioning in Midaz
- The plugin exposes `POST /v1/accounts/receivables` to create an internal receivables wallet.
- Calls Midaz Onboarding API (`POST /v1/organizations/{org_id}/ledgers/{ledger_id}/accounts`) with `type: "asset"` and alias `@<alias>/receivables`.
- Stores the mapping locally with Midaz `account_id`, `ledger_id`, and `organization_id`.

### FR-02: Multi-Acquirer Connection Linking (1 Account : N Connections)
- The plugin exposes `POST /v1/accounts/receivables/{id}/connections` to link credentials for a given acquirer:
  - `acquirer_type`: `SUMUP`, `PAGBANK`.
  - `credentials`: Acquirer-specific API keys, merchant codes, or OAuth tokens.
  - `auto_sync_enabled`: Boolean flag enabling background polling.
  - `sync_interval_minutes`: Interval for background synchronization.
- A single Receivables Account may have both a SumUp connection and a PagBank connection active simultaneously.

### FR-03: Universal Acquirer Clearing Account Bootstrapping
- Upon startup, the plugin verifies that universal clearing accounts exist in Midaz:
  - `@external/acquirer_sumup`
  - `@external/acquirer_pagbank`
- If missing, it provisions them via Midaz Onboarding API with `type: "external"`.

### FR-04: Normalized Sales & Installment Ingestion
- Ingests card sales from active connections, calculating MDR and net values per installment using arbitrary-precision decimal arithmetic.
- Normalized fields:
  - `connection_id`, `receivables_account_id`.
  - `acquirer`, `nsu`, `tid`, `authorization_code`.
  - `gross_amount`, `mdr_fee`, `net_amount`.
  - `installment_number`, `total_installments`.
  - `expected_settlement_date`.

### FR-05: Double-Entry Posting to Midaz (D+0)
- For every confirmed sale installment, post a balanced transaction to Midaz:
  - **Debit:** `@external/acquirer_<name>` (Gross amount).
  - **Credit:** `@<alias>/receivables` (Net amount).
  - **Credit (Optional):** `@platform/mdr_fee` (Platform spread).
  - **Deterministic Code:** `REC-<ACQUIRER>-<NSU>-<INSTALLMENT>`.

### FR-06: Matcher 6 Reconciliation Bridge
- Pushes expected installments to Matcher 6 (`POST /v1/imports/contexts/{context_id}/sources/{source_agenda_id}/upload`).
- Listens on `/v1/webhooks/matcher` for approved match events.
- Upon receiving a match, triggers atomic settlement in Midaz:
  - **Debit:** `@<alias>/receivables`.
  - **Credit:** `@<alias>/checking`.
  - **Code:** `SETTLE-<ACQUIRER>-<NSU>-<INSTALLMENT>`.

---

## 6. Non-Functional Requirements (NFR)
- **NFR-01 (Strict English Standard):** All code, schemas, documentation, and error codes are in English.
- **NFR-02 (Mathematical Precision):** All monetary amounts use fixed-point arithmetic (`shopspring/decimal`).
- **NFR-03 (Idempotency):** Re-delivered webhooks or sync retries produce zero duplicate transactions in Midaz.
- **NFR-04 (Traceability):** Every ledger transaction carries metadata referencing Acquirer NSU, Matcher Run ID, and Connection ID.
- **NFR-05 (Cryptographic Security at Rest):** All merchant acquirer credentials (API keys, merchant codes, OAuth tokens, secrets) must be encrypted at rest using AES-256-GCM authenticated encryption with unique per-record nonces. Plaintext credentials must never be written to persistent storage or logs, and must be decrypted strictly in memory at runtime during API calls.
