# Accounting & Double-Entry Specification: Receivables Lifecycle

## 1. Accounting Foundations in Midaz

Midaz enforces the double-entry bookkeeping identity formulated by Luca Pacioli:
$$\sum \text{Debits} - \sum \text{Credits} = 0$$

In financial systems, **Credit Card Receivables** are accrued financial assets representing future claims against a card acquirer. They cannot be credited directly into a merchant's liquid checking account upon capture without inflating the money supply and distorting ledger invariants.

---

## 2. Account Topology

The plugin designs a multi-tiered account hierarchy within the target Midaz Ledger:

```
[External Acquirer Universe]
            │
            ▼ (Initial Gross Claim Debit)
[@external/acquirer_sumup] (Universal External Clearing Account)
            │
            ├── (Net Claim Credit) ──────────────► [@merchant_alpha/receivables] (Asset: Accrued Future Claims)
            └── (Platform Fee Credit) ───────────► [@platform/mdr_revenue] (Platform Operational Revenue)
                                                                │
                                                                ▼ (Settlement via Matcher 6 Approval)
                                                   [@merchant_alpha/checking] (Liability: Available Liquid Funds)
```

### Account Hierarchy Definitions:
1. **Universal Acquirer Clearing Account (`@external/acquirer_<name>`)**:
   - `type: "external"`
   - Permitted by Midaz to hold negative balances.
   - Represents the gross aggregated claims owed by the card acquirer to our platform.
2. **Merchant Receivables Account (`@<alias>/receivables`)**:
   - `type: "asset"`
   - Holds the merchant's unsettled, accrued credit rights across all bound acquirer connections.
3. **Merchant Checking Account (`@<alias>/checking`)**:
   - `type: "liability"` (from the ledger's perspective: funds available for merchant withdrawal or transfers).
4. **Platform Fee Account (`@platform/mdr_revenue`)**:
   - `type: "revenue"`
   - Records the platform's spread/commission on card transactions.

---

## 3. The 3-Phase Accounting Lifecycle

### Phase 1: Sale Ingestion & Agenda Booking (D+0)
* **Real-world event:** A customer purchases R$ 300.00 in 3 installments at Merchant Alpha via SumUp.
* **Acquirer MDR fee:** 3.00% (R$ 9.00 total, R$ 3.00 per installment).
* **Net amount to merchant:** R$ 291.00 (R$ 97.00 per installment).
* **Midaz Ledger Entry (Per Installment):**
  - **Debit:** `@external/acquirer_sumup` $\rightarrow$ R$ 100.00 (Gross claim against SumUp).
  - **Credit:** `@merchant_alpha/receivables` $\rightarrow$ R$ 97.00 (Net merchant receivable).
  - **Credit:** `@platform/mdr_revenue` $\rightarrow$ R$ 3.00 (Fee retained).

#### Midaz Transaction Payload (D+0):
```json
{
  "chartOfAccountsGroupName": "ACQUIRER_RECEIVABLES",
  "description": "SumUp Card Receivable - Installment 1/3 (NSU: 987654321)",
  "code": "REC-SUMUP-987654321-1",
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
          "description": "Credit Merchant Receivables Account (Net Value)"
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

### Phase 2: Reconciliation & Settlement Extinguishment (D+X)
* **Real-world event:** On 2026-12-20, SumUp deposits R$ 97.00 into the merchant's bank account.
* **The Open Finance Plugin:** Reads the bank credit line and pushes it to **Matcher 6** as Source B.
* **The Receivables Plugin:** Pushed expected installment R$ 97.00 to **Matcher 6** as Source A.
* **Matcher 6:** Matches the records and dispatches an approved match event.
* **Midaz Ledger Entry (Settlement Transfer):**
  - **Debit:** `@merchant_alpha/receivables` $\rightarrow$ R$ 97.00 (Extinguishes the accrued receivable).
  - **Credit:** `@merchant_alpha/checking` $\rightarrow$ R$ 97.00 (Avails liquid spendable funds).

#### Midaz Settlement Transaction Payload:
```json
{
  "chartOfAccountsGroupName": "RECEIVABLES_SETTLEMENT",
  "description": "SumUp Settlement Extinguishment - Matcher 6 Confirmed",
  "code": "SETTLE-SUMUP-987654321-1",
  "metadata": {
    "matcher_run_id": "a9b8c7d6-e5f4-3210-9876-543210abcdef",
    "acquirer": "SUMUP",
    "nsu": "987654321",
    "installment_number": 1,
    "settlement_date": "2026-12-20",
    "bank_tx_id": "OFP-ITAU-998877"
  },
  "send": {
    "asset": "BRL",
    "value": "97.00",
    "source": {
      "from": [
        {
          "accountAlias": "@merchant_alpha/receivables",
          "amount": { "asset": "BRL", "value": "97.00" },
          "description": "Debit Receivables Account (Extinguish Claim)"
        }
      ]
    },
    "distribute": {
      "to": [
        {
          "accountAlias": "@merchant_alpha/checking",
          "amount": { "asset": "BRL", "value": "97.00" },
          "description": "Credit Checking Account (Available Cash)"
        }
      ]
    }
  }
}
```

---

### Phase 3: Batch Settlements ($N:1$ Matching)

When an acquirer combines multiple sales into a single wire transfer (e.g., 3 sales of R$ 50, R$ 70, R$ 80 settled in a single bank credit of R$ 200.00):

Midaz's native **$N:1$ transaction support** executes this atomically:

```json
{
  "chartOfAccountsGroupName": "BATCH_RECEIVABLES_SETTLEMENT",
  "description": "Batch Settlement SumUp - 3 Receivables to 1 Bank Deposit",
  "code": "SETTLE-BATCH-SUMUP-20261220-001",
  "send": {
    "asset": "BRL",
    "value": "200.00",
    "source": {
      "from": [
        {
          "accountAlias": "@merchant_alpha/receivables",
          "amount": { "asset": "BRL", "value": "50.00" },
          "description": "Settle Receivable NSU 101"
        },
        {
          "accountAlias": "@merchant_alpha/receivables",
          "amount": { "asset": "BRL", "value": "70.00" },
          "description": "Settle Receivable NSU 102"
        },
        {
          "accountAlias": "@merchant_alpha/receivables",
          "amount": { "asset": "BRL", "value": "80.00" },
          "description": "Settle Receivable NSU 103"
        }
      ]
    },
    "distribute": {
      "to": [
        {
          "accountAlias": "@merchant_alpha/checking",
          "amount": { "asset": "BRL", "value": "200.00" },
          "description": "Credit Unified Checking Account"
        }
      ]
    }
  }
}
```
