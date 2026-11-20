# Accounting & Double-Entry Specification: Receivables Lifecycle

## 1. Accounting Foundations in Midaz

Midaz enforces the double-entry bookkeeping identity formulated by Luca Pacioli:
$$\sum \text{Debits} - \sum \text{Credits} = 0$$

In financial systems, **Credit Card Receivables** are accrued financial assets representing future claims against a card acquirer (e.g., SumUp, PagBank). They cannot be credited directly into a merchant's liquid bank account upon capture without inflating the money supply and distorting ledger invariants.

Furthermore, future receivables are inherently **unsettled commitments**:
- On sale capture ($D+0$), the claim is recorded in Midaz with **`pending: true`**, preserving full auditability of the future schedule without releasing available spendable balance.
- On settlement ($D+X$), once bank statement credits are conciliated via **Matcher 6**, two actions occur atomically:
  1. The pending transactions in the receivables account are committed (`POST .../transactions/{id}/commit`), actualizing the credit in `@<merchant>/receivables`.
  2. An **atomic $N:1$ multi-source transfer** moves the funds from `@<merchant>/receivables` into the merchant's mirrored bank account (`@<merchant>/openfinance`), with each source leg reflecting its respective installment claim.

---

## 2. Account Topology

The plugin designs a multi-tiered account hierarchy within the target Midaz Ledger:

```
[External Acquirer Universe]
            │
            ▼ (D+0 Pending Sale Ingestion: pending: true)
[@external/acquirer_sumup] (Universal External Clearing Account)
            │
            ├── (Net Claim: Pending Credit) ───────► [@merchant_alpha/receivables] (Asset: Future Claim on Acquirer)
            └── (Platform Fee Credit) ─────────────► [@platform/mdr_revenue] (Platform Operational Revenue)
                                                                │
                                                                ▼ (D+X Matcher 6 Approval + Multi-Source Transfer)
                                                    [@merchant_alpha/openfinance] (Mirrored Bank Account)
                                                                ▲
                                                                │ (Non-card generic bank deposits)
                                                    [@external/BRL:openfinance]
```

### Account Hierarchy Definitions:
1. **Universal Acquirer Clearing Account (`@external/acquirer_<name>`)**:
   - `type: "external"`
   - Permitted by Midaz to hold negative balances.
   - Represents the gross aggregated claims owed by the card acquirer to our platform.
2. **Merchant Receivables Account (`@<alias>/receivables`)**:
   - `type: "asset"`
   - Holds the merchant's unsettled, accrued credit rights across all bound acquirer connections.
   - Installments are initially posted as `pending: true` and committed upon settlement confirmation.
3. **Mirrored Bank Account (`@<alias>/openfinance`)**:
   - `type: "liability"` (from the ledger's perspective: mirrored checking balance from Open Finance).
   - Receives funds either from `@<alias>/receivables` (when reconciled against acquirer receivables) or from `@external/BRL:openfinance` (for generic external deposits).
4. **Platform Fee Account (`@platform/mdr_revenue`)**:
   - `type: "revenue"`
   - Records the platform's spread/commission on card transactions.

---

## 3. The 3-Phase Accounting Lifecycle

### Phase 1: Sale Ingestion & Agenda Booking ($D+0$ Pending Transactions)
* **Real-world event:** A customer purchases R$ 300.00 in 3 installments at Merchant Alpha via SumUp.
* **Acquirer MDR fee:** 3.00% (R$ 9.00 total, R$ 3.00 per installment).
* **Net amount to merchant:** R$ 291.00 (R$ 97.00 per installment).
* **Midaz Ledger Entry (Per Installment, with `"pending": true`):**
  - **Debit:** `@external/acquirer_sumup` $\rightarrow$ R$ 100.00 (Gross claim against SumUp).
  - **Credit (Pending):** `@merchant_alpha/receivables` $\rightarrow$ R$ 97.00 (Net merchant receivable, on-hold).
  - **Credit:** `@platform/mdr_revenue` $\rightarrow$ R$ 3.00 (Fee retained).
* **Captured Identifier:** The plugin stores the resulting `midaz_receivable_tx_id` in `receivables_installments` for later commitment.

#### Midaz Transaction Payload ($D+0$, Pending):
```json
{
  "chartOfAccountsGroupName": "ACQUIRER_RECEIVABLES",
  "description": "SumUp Card Receivable - Installment 1/3 (NSU: 987654321)",
  "code": "REC-SUMUP-987654321-1",
  "pending": true,
  "metadata": {
    "acquirer": "SUMUP",
    "connection_id": "01948df2-0000-7000-8000-000000000001",
    "nsu": "987654321",
    "tid": "TID-554433",
    "installment_number": 1,
    "total_installments": 3,
    "gross_value": "100.00",
    "mdr_fee": "3.00",
    "net_value": "97.00",
    "expected_settlement_date": "2026-12-20"
  },
  "send": {
    "asset": "BRL",
    "value": "100.00",
    "source": {
      "from": [
        {
          "accountAlias": "@external/acquirer_sumup",
          "amount": { "asset": "BRL", "value": "100.00" },
          "description": "Debit SumUp Clearing Account (Gross Claim)"
        }
      ]
    },
    "distribute": {
      "to": [
        {
          "accountAlias": "@merchant_alpha/receivables",
          "amount": { "asset": "BRL", "value": "97.00" },
          "description": "Credit Merchant Receivables Account (Net Value, Pending)"
        },
        {
          "accountAlias": "@platform/mdr_revenue",
          "amount": { "asset": "BRL", "value": "3.00" },
          "description": "Credit Platform MDR Fee Revenue"
        }
      ]
    }
  }
}
```

---

### Phase 2: Reconciliation & Two-Phase Settlement ($D+X$)

1. **Bank Statement Sync:** On 2026-12-20, an acquirer deposit of R$ 300.00 arrives in the merchant's Sicredi bank account.
2. **Pre-Posting Matcher Gate:** Because `reconciliation_enabled: true` on `@merchant_alpha/openfinance`, the bank transaction is NOT immediately posted from `@external/BRL:openfinance`. Instead, it is sent to **Matcher 6** as Source B.
3. **Matcher 6 Resolution:** Matcher 6 finds 3 scheduled receivables (e.g., R$ 200.00 from SumUp, R$ 50.00 from PagBank, R$ 50.00 from PagBank) totaling R$ 300.00.
4. **Reconciliation Execution:**
   - **Step A (Commit Pending Operations):** For each of the 3 receivables, the plugin calls Midaz:
     ```http
     POST /v1/organizations/{org_id}/ledgers/{ledger_id}/transactions/{midaz_receivable_tx_id}/commit
     ```
     This moves each receivable out of `PENDING` status, actualizing the credit in `@merchant_alpha/receivables`.
   - **Step B (Multi-Source $N:1$ Transfer):** The plugin executes an atomic transaction with 3 source legs originating from `@merchant_alpha/receivables` and 1 destination leg to `@merchant_alpha/openfinance`.

#### Midaz Multi-Source Settlement Transaction Payload ($N:1$):
```json
{
  "chartOfAccountsGroupName": "OPEN_FINANCE_SYNC",
  "description": "Open Finance Deposit OFP-SICREDI-998877 (Conciliated with 3 Receivables)",
  "code": "OFP-SICREDI-998877",
  "metadata": {
    "matcher_run_id": "a9b8c7d6-e5f4-3210-9876-543210abcdef",
    "match_id": "match_01948df2_998877",
    "bank_tx_id": "OFP-SICREDI-998877",
    "receivables_count": 3,
    "settlement_date": "2026-12-20",
    "reconciliation_strategy": "MULTI_SOURCE_SETTLEMENT"
  },
  "send": {
    "asset": "BRL",
    "value": "300.00",
    "source": {
      "from": [
        {
          "accountAlias": "@merchant_alpha/receivables",
          "amount": { "asset": "BRL", "value": "200.00" },
          "description": "Receivable REC-SUMUP-1001-1 (NSU: 1001)"
        },
        {
          "accountAlias": "@merchant_alpha/receivables",
          "amount": { "asset": "BRL", "value": "50.00" },
          "description": "Receivable REC-PAGBANK-2001-1 (NSU: 2001)"
        },
        {
          "accountAlias": "@merchant_alpha/receivables",
          "amount": { "asset": "BRL", "value": "50.00" },
          "description": "Receivable REC-PAGBANK-2002-1 (NSU: 2002)"
        }
      ]
    },
    "distribute": {
      "to": [
        {
          "accountAlias": "@merchant_alpha/openfinance",
          "amount": { "asset": "BRL", "value": "300.00" },
          "description": "Credit Mirrored Open Finance Bank Account"
        }
      ]
    }
  }
}
```

---

### Phase 3: Non-Reconciled Bank Transactions

Any transaction entering the mirrored bank account that is **not** matched to card receivables (e.g., standard Pix incoming, payroll, vendor wire) follows the standard Open Finance direct posting:

* **Debit:** `@external/BRL:openfinance` $\rightarrow$ Total amount.
* **Credit:** `@merchant_alpha/openfinance` $\rightarrow$ Total amount.

This guarantees that:
1. Receivables balances in Midaz are never double-counted.
2. The bank statement mirror (`@merchant_alpha/openfinance`) precisely reflects bank reality.
3. Every card settlement in the bank statement is directly traceable to the specific receivables that funded it.
