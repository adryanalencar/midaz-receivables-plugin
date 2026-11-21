# Roadmap: Advanced Frontend Flows for Midaz Receivables Plugin

This document outlines the advanced operational flows and edge cases scheduled for future implementation in `frontend/`.

---

## 1. Sale Simulation & Inbound Webhook Emulator
- **Objective:** Enable developers and operational teams to test D+0 sale ingestion in local/staging environments without relying on real terminal hardware.
- **Components to Add:**
  - `SaleSimulatorModal`: Dialog to create synthetic sales specifying:
    - Acquirer (`SUMUP`, `ADITUM`, `PAGBANK`)
    - NSU, TID, Authorization Code
    - Card Brand (Visa, Mastercard, Elo)
    - Gross Amount & Installments Count ($1$ to $12$)
  - Immediate trigger of `POST /v1/webhooks/:acquirer/:connection_id` with signature payload.
  - Live refresh of the Receivables Agenda table and Midaz double-entry transaction link.

---

## 2. Receivables Anticipation Flow (Antecipação de Recebíveis)
- **Objective:** Support merchant liquidity acceleration (anticipating D+30 installments to D+0).
- **Components to Add:**
  - `AnticipateInstallmentModal`: Action button in the agenda table row.
  - Calculates anticipation discount rate ($N\%$ monthly pro-rata).
  - Posts anticipation transaction in Midaz:
    - Debit: `@<merchant>/checking` (Net Anticipated)
    - Debit: `@<merchant>/receivables/fees` (Anticipation Spread)
    - Credit: `@<merchant>/receivables` (Original Accrued Claim)
  - Updates installment status to `ANTICIPATED_EXTINGUISHED`.

---

## 3. Disputes, Reversals & Chargebacks Handling
- **Objective:** Manage post-sale cancellations or consumer chargebacks occurring prior to or after settlement.
- **Components to Add:**
  - `DisputeResolutionDrawer`:
    - Flags installment as `DISPUTED` or `CHARGEBACK_DEBITED`.
    - Double-entry adjustment: Debit `@<merchant>/receivables` or `@<merchant>/checking`, Credit `@external/acquirer_*`.

---

## 4. Batch Settlement Visualizer ($N:1$ and $N:N$)
- **Objective:** Give financial controllers full clarity on how multiple individual card installments correspond to a single consolidated bank payout.
- **Components to Add:**
  - `BatchConsolidationModal`:
    - Shows batch breakdown: e.g. 84 installments totaling R$ 45.200,00 matching 1 bank deposit credit.
    - Drill-down list of NSUs included in the settlement bundle.
