# Midaz Plugin: Receivables & Settlement Reconciliation

An enterprise-grade plugin for the **[Midaz Ledger](https://github.com/LerianStudio/midaz)** and **[Matcher 6](http://localhost:4018)** that manages card acquirers' sales agendas (**Agenda de Recebíveis**) and choreographs three-way bank settlement reconciliation with the **Midaz Open Finance Plugin**.

---

## 🌟 Executive Summary & Problem Solved

Fintechs, marketplaces, and merchant aggregators operating across multiple business entities, merchant tenants, and corporate tax IDs (CNPJs) face two critical challenges:

1. **Multi-Tenant Acquirer Credential Isolation:** Each merchant possesses separate credentials with payment acquirers (e.g., SumUp, PagBank, Aditum). A single merchant may operate across multiple acquirers simultaneously. The platform must provision a designated **Receivables Account** in Midaz for each merchant and dynamically associate multiple isolated **Acquirer Connections** to that specific account.
2. **The Double-Credit Dilemma:** When a customer buys via credit card (D+0), the merchant earns an accrued future credit right. If this credit right is booked in the ledger, and 30 days later when the acquirer deposits money into the merchant's bank account an Open Finance sync credits the account again, the merchant's balance is **duplicated**.

**The Solution:**
This plugin acts as the central bridge between **Card Acquirers**, **Midaz Ledger**, **Plugin Open Finance**, and **Matcher 6**:
* **Receivables Account Provisioning:** Provisions merchant receivables accounts (`@<alias>/receivables`) in Midaz and allows binding multiple acquirer credentials to them (1 Account : $N$ Connections).
* **D+0 (Sale Ingestion):** Ingests card sales from SumUp/PagBank/Aditum and books them into `@<alias>/receivables` counterbalanced by universal clearing accounts (`@external/acquirer_<name>`).
* **D+X (Settlement Day):** When the bank deposit arrives, `plugin-openfinance` intercepts the credit and routes it to **Matcher 6** as Source B instead of posting a direct deposit.
* **Matcher 6 (Reconciliation):** Matches the expected receivable (Source A) with the real bank statement credit (Source B).
* **Midaz Settlement (Double-Entry Extinguishment):** Upon Matcher 6 approval, this plugin executes an atomic transfer in Midaz:
  $$\texttt{@<alias>/receivables} \longrightarrow \texttt{@<alias>/checking}$$
  The receivable is cleared, liquid cash is made available, and balance sheet integrity is preserved with **zero duplication**.

---

## 🗺️ System Architecture

```mermaid
flowchart TD
    subgraph Acquirers["1. Payment Acquirers (Multi-Tenant)"]
        SU["SumUp REST APIs & Webhooks<br/>(Merchant A Credentials)"]
        PB["PagBank REST APIs & Webhooks<br/>(Merchant A Credentials)"]
        AD["Aditum REST APIs, EDI 3.2 & Webhooks<br/>(Merchant A Credentials)"]
    end

    subgraph ReceivablesPlugin["2. Midaz Plugin: Receivables (Port 8082)"]
        direction TB
        subgraph AccountModel["Multi-Connection Model"]
            RecAcc["Receivables Account<br/>@merchant_alpha/receivables"]
            Conn1["Connection 1: SumUp"]
            Conn2["Connection 2: PagBank"]
            RecAcc --- Conn1
            RecAcc --- Conn2
        end
        IngestEngine["Ingestion Engine<br/>• MDR Calculation & Splitting"]
        SettleWorker["Settlement Worker<br/>• Matcher 6 Webhook Listener"]
    end

    subgraph MidazPlatform["3. Midaz Core Ledger (Port 3000 / 3001)"]
        Clearance["@external/acquirer_sumup<br/>(Universal External Clearing)"]
        AgendaAcc["@merchant_alpha/receivables<br/>(Asset: Accrued Claims D+30)"]
        LiquidAcc["@merchant_alpha/checking<br/>(Liability: Available Liquid Funds)"]
    end

    subgraph OFFlow["4. Plugin Open Finance (Port 8080)"]
        BankExtract["Real Bank Statement Credit<br/>(Itaú, Bradesco, etc.)"]
        ReconcileFlag{"Account flagged with<br/>reconciliation.enabled == true?"}
    end

    subgraph MatcherPlatform["5. Matcher 6 Engine (Port 4018)"]
        SourceA["Source A: Receivables Schedule<br/>(Uploaded by Receivables Plugin)"]
        SourceB["Source B: Bank Statement Credits<br/>(Uploaded by Open Finance Plugin)"]
        MatchRules["Matching Rules:<br/>• Net Amount == Bank Credit<br/>• Date tolerance D-1 to D+1<br/>• Merchant & Acquirer ID"]
        SourceA --> MatchRules
        SourceB --> MatchRules
    end

    SU --> Conn1
    PB --> Conn2
    Conn1 --> IngestEngine
    Conn2 --> IngestEngine

    %% D+0 Flow
    IngestEngine -->|1. Post D+0: Gross Claim| Clearance
    IngestEngine -->|1. Post D+0: Net Receivable| AgendaAcc
    IngestEngine -->|2. Upload Expectation| SourceA

    %% Bank Flow
    BankExtract --> ReconcileFlag
    ReconcileFlag -->|Yes: Intercept & Push to Matcher| SourceB

    %% Matcher Execution
    MatchRules -->|3. Match Approved!| SettleWorker

    %% Settlement Flow
    SettleWorker -->|4. Post D+X: Extinguish Claim| AgendaAcc
    SettleWorker -->|4. Post D+X: Credit Liquid Cash| LiquidAcc
```

---

## ⚡ Core Capabilities

- **Multi-Tenant Acquirer Connections:** Binds multiple acquirer credentials to a single Receivables Account (1 Account : $N$ Connections) without coupling core logic to tax IDs.
- **Double-Entry Invariants:** Strictly adheres to Midaz double-entry rules. Clean mathematical boundary between *Receivables Claims* (unsettled asset) and *Liquid Cash* (settled liability).
- **Matcher 6 Native Integration:** Direct integration with Matcher 6 (`:4018`) via Huma v2 OpenAPI 3.1 endpoints and webhook callbacks.
- **Inter-Plugin Protocol:** Coordinates with `plugin-openfinance` using the `reconciliation.enabled` account flag to prevent double-credit balance inflation.
- **Multi-Item & Batch Settlement ($N:1$ and $N:N$):** Consolidates multiple card installments into a single lump-sum bank wire or splits multi-merchant batches.
- **Strict English Standard:** All code, domain models, database schemas, APIs, error messages, and tests are in English.
- **Midaz Console Dynamic Discovery:** Serves `/api/manifest` allowing the plugin to mount dynamically inside the Midaz Console.
- **Embedded Database Governance:** Runs automated PostgreSQL migrations on boot using `golang-migrate` with distributed advisory locks.

---

## 🛠️ Quickstart (Local Development)

### 1. Prerequisites
- Docker & Docker Compose
- Midaz Ledger running locally (v3.4.8+)
- Matcher 6 running locally (`http://localhost:4018`)
- PostgreSQL 16 (for plugin state)

### 2. Environment Configuration (`.env`)
```env
# Server
PORT=8082
ENVIRONMENT_NAME=local
LOG_LEVEL=info

# PostgreSQL
DATABASE_URL=postgres://midaz:midaz123@localhost:5432/midaz_receivables?sslmode=disable

# Midaz Ledger Core
MIDAZ_ONBOARDING_URL=http://localhost:3000
MIDAZ_TRANSACTION_URL=http://localhost:3001
MIDAZ_ORGANIZATION_ID=01948df2-0000-7000-8000-000000000001
MIDAZ_LEDGER_ID=01948df2-0000-7000-8000-000000000002

# Matcher 6 Reconciliation
MATCHER_URL=http://localhost:4018
MATCHER_CONTEXT_SLUG=acquirers-settlement
MATCHER_WEBHOOK_SECRET=secret_matcher_local_123
```

---

## 📚 Detailed Documentation

All business logic, accounting rules, and technical designs are documented in depth inside the `docs/` directory:

| Document | Description |
| :--- | :--- |
| **[01 - Product Requirements (PRD)](file:///Users/adryanalencar/Documents/projetos%202/plugin-receivables/docs/01-PRODUCT-REQUIREMENTS.md)** | Multi-tenant problem statement, personas, functional requirements, and edge cases. |
| **[02 - Technical Architecture](file:///Users/adryanalencar/Documents/projetos%202/plugin-receivables/docs/02-TECHNICAL-ARCHITECTURE.md)** | Hexagonal architecture, CQRS, directory structure, PostgreSQL schemas, and 1:N connection model. |
| **[03 - Accounting & Double-Entry](file:///Users/adryanalencar/Documents/projetos%202/plugin-receivables/docs/03-ACCOUNTING-DOUBLE-ENTRY.md)** | Midaz double-entry rules, `@external` clearing accounts, D+0 and D+X payloads, and $N:1$ batch settlements. |
| **[04 - Acquirers Specification](file:///Users/adryanalencar/Documents/projetos%202/plugin-receivables/docs/04-ACQUIRER-PROVIDERS-SPEC.md)** | SumUp and PagBank integration specs with connection-scoped credentials and webhooks. |
| **[05 - Matcher 6 Reconciliation](file:///Users/adryanalencar/Documents/projetos%202/plugin-receivables/docs/05-MATCHER-RECONCILIATION-SPEC.md)** | Matcher 6 Context/Sources setup, matching rules, ingestion APIs, and webhook processing. |
| **[06 - Inter-Plugin Protocol](file:///Users/adryanalencar/Documents/projetos%202/plugin-receivables/docs/06-INTER-PLUGIN-PROTOCOL-OPENFINANCE.md)** | Choreography between `plugin-openfinance` and `plugin-receivables` via the `reconciliation.enabled` flag. |
| **[07 - API & Console Manifest](file:///Users/adryanalencar/Documents/projetos%202/plugin-receivables/docs/07-API-AND-MANIFEST-SPEC.md)** | Complete REST endpoints for accounts, connections, sync, webhooks, and Midaz Console manifest. |
| **[08 - Implementation Plan](file:///Users/adryanalencar/Documents/projetos%202/plugin-receivables/docs/08-IMPLEMENTATION-PLAN.md)** | 6-phase disciplined execution roadmap from domain scaffolding to container packaging. |
| **[09 - Testing Strategy](file:///Users/adryanalencar/Documents/projetos%202/plugin-receivables/docs/09-TESTING-STRATEGY.md)** | Test pyramid, unit tests, PostgreSQL integration, and edge-case simulation suites. |
