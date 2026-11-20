# Accounting & Double-Entry Specification: Receivables Lifecycle

## 1. Accounting Foundations in Midaz

Midaz enforces the double-entry bookkeeping identity formulated by Luca Pacioli:
$$\sum \text{Debits} - \sum \text{Credits} = 0$$

In financial systems, **Credit Card Receivables** are accrued financial assets representing future claims against a card acquirer (e.g., SumUp, PagBank). They cannot be credited directly into a merchant's liquid bank account upon capture without inflating the money supply and distorting ledger invariants.

Furthermore, future receivables are inherently **unsettled commitments**:
- On sale capture ($D+0$), the claim is recorded in Midaz with **`pending: true`**, placing funds on hold in the acquirer's internal liquidity account without releasing spendable liquid balance.
- On settlement ($D+X$), once bank statement credits are conciliated via **Matcher 6**, two actions occur atomically:
  1. The pending transactions in the acquirer account are committed (`POST .../transactions/{id}/commit`), actualizing the credit in `@<merchant>/receivables`.
  2. An **atomic $N:M$ multi-leg transfer** moves the funds from `@<merchant>/receivables` into the merchant's mirrored bank account (`@<merchant>/bank_<name>`), routing any fee deductions (e.g., machine rental) to `@<merchant>/receivables/fees`.
  3. An internal acknowledgment notification is dispatched to `plugin-openfinance` to mark the bank transaction as processed.

---

## 2. Account Hierarchy & Subaccount Topology

To guarantee strict tenant isolation, multi-CNPJ segregation, and avoid Midaz balance restrictions (`ErrOnHoldExternalAccount` - Code 0098), the architecture structures accounts as follows:

```
[Universal Fiat Universe]
           │
           ▼ (Step 1: Immediate Influx Funding)
[@external/BRL:universal] (Universal External Fiat Account)
           │
           ▼
[@acquirer/sumup] (Internal Liquidity Account: type = deposit)
           │
           ▼ (Step 2: Pending Allocation: pending = true)
           │── (Net Claim: Pending Credit) ───► [@merchant_alpha/receivables] (Subaccount: Future Claims)
           └── (Platform MDR Fee Credit) ─────► [@platform/mdr_revenue] (Platform Operational Revenue)
                                                                │
                                                                ▼ (D+X Matcher 6 Approval + Multi-Leg Transfer)
                                                    [@merchant_alpha/bank_sicredi] (Mirrored Bank Subaccount)
                                                                │
                                                                ├── (Optional Deductions: Aluguel Maquininha)
                                                                ▼
                                                    [@merchant_alpha/receivables/fees] (Tariff Subaccount)
```

### Account Definitions:
1. **Universal External Fiat Account (`@external/BRL:universal`)**:
   - `type: "external"`
   - Permitted by Midaz to hold negative balances.
   - Injects gross liquidity into internal acquirer accounts upon sale ingestion.
2. **Internal Acquirer Liquidity Accounts (`@acquirer/<name>`)**:
   - `type: "deposit"` (Internal account, NOT external).
   - Allows Midaz to safely place funds `on_hold` when `pending: true` without triggering Error 0098.
   - **Operational Metric:** The `on_hold` balance of `@acquirer/sumup` represents the exact total gross volume of future receivables currently floating across all clients.
3. **Mother Account (Master CNPJ Profile: `@<merchant>`)**:
   - Master account in Midaz (e.g. `@merchant_alpha`).
   - Mapped 1:1 to a dedicated **Tenant / Context in Matcher 6**.
4. **Merchant Receivables Subaccount (`@<merchant>/receivables`)**:
   - `parentAccountId = merchant_alpha.id`
   - Holds the merchant's unsettled, accrued credit rights across all bound acquirer connections.
5. **Merchant Tariff & Deductions Subaccount (`@<merchant>/receivables/fees`)**:
   - `parentAccountId = merchant_alpha.id`
   - Absorbs settlement deductions (e.g., POS terminal rentals, chargebacks, anticipation fees).
6. **Mirrored Bank Subaccount (`@<merchant>/bank_<name>`)**:
   - `parentAccountId = merchant_alpha.id`
   - Mirrors bank statements via Open Finance (e.g., Sicredi, Itaú).

---

## 3. The 3-Phase Accounting Lifecycle

### Phase 1: Sale Ingestion & Agenda Booking ($D+0$ Two-Step Sequence)
* **Real-world event:** A customer purchases R$ 300.00 in 3 installments at Merchant Alpha via SumUp.
* **Acquirer MDR fee:** 3.00% (R$ 9.00 total, R$ 3.00 per installment).
* **Net amount to merchant:** R$ 291.00 (R$ 97.00 per installment).

#### Step 1: Influx Funding (Immediate)
```json
{
  "chartOfAccountsGroupName": "ACQUIRER_FUNDING",
  "description": "Funding SumUp Acquirer Liquidity for Sale NSU 987654321 (Installment 1/3)",
  "code": "FUND-SUMUP-987654321-1",
  "send": {
    "asset": "BRL",
    "value": "100.00",
    "source": {
      "from": [{ "accountAlias": "@external/BRL:universal", "amount": { "asset": "BRL", "value": "100.00" } }]
    },
    "distribute": {
      "to": [{ "accountAlias": "@acquirer/sumup", "amount": { "asset": "BRL", "value": "100.00" } }]
    }
  }
}
```

#### Step 2: Agenda Booking (Pending: true)
```json
{
  "chartOfAccountsGroupName": "ACQUIRER_RECEIVABLES",
  "description": "SumUp Card Receivable - Installment 1/3 (NSU: 987654321)",
  "code": "REC-SUMUP-987654321-1",
  "pending": true,
  "metadata": {
    "acquirer": "SUMUP",
    "nsu": "987654321",
    "installment_number": 1,
    "gross_value": "100.00",
    "mdr_fee": "3.00",
    "net_value": "97.00",
    "expected_settlement_date": "2026-12-20"
  },
  "send": {
    "asset": "BRL",
    "value": "100.00",
    "source": {
      "from": [{ "accountAlias": "@acquirer/sumup", "amount": { "asset": "BRL", "value": "100.00" } }]
    },
    "distribute": {
      "to": [
        { "accountAlias": "@merchant_alpha/receivables", "amount": { "asset": "BRL", "value": "97.00" } },
        { "accountAlias": "@platform/mdr_revenue", "amount": { "asset": "BRL", "value": "3.00" } }
      ]
    }
  }
}
```
* **Ledger state:** `@acquirer/sumup` has Available: 0, OnHold: R$ 100.00. The Midaz transaction ID is saved in `receivables_installments.midaz_receivable_tx_id`.

---

### Phase 2: Reconciliation & Multi-Leg Settlement ($D+X$)

1. **Bank Statement Sync:** SumUp deposits R$ 250.10 in Sicredi (R$ 300.00 in sales minus R$ 49.90 POS terminal rental).
2. **Matcher 6 Reconciliation:** Matcher 6 (running in the dedicated Context for `@merchant_alpha`) reconciles the bank credit against the 3 receivables.
3. **Execution Sequence:**
   - **Step A:** `POST /transactions/{midaz_receivable_tx_id}/commit` for each of the 3 receivables.
   - **Step B:** Atomic $N:M$ settlement transfer:

```json
{
  "chartOfAccountsGroupName": "OPEN_FINANCE_SYNC",
  "description": "Open Finance Deposit OFP-SICREDI-998877 (Conciliated with 3 Receivables)",
  "code": "OFP-SICREDI-998877",
  "send": {
    "asset": "BRL",
    "value": "300.00",
    "source": {
      "from": [
        { "accountAlias": "@merchant_alpha/receivables", "amount": { "asset": "BRL", "value": "200.00" } },
        { "accountAlias": "@merchant_alpha/receivables", "amount": { "asset": "BRL", "value": "50.00" } },
        { "accountAlias": "@merchant_alpha/receivables", "amount": { "asset": "BRL", "value": "50.00" } }
      ]
    },
    "distribute": {
      "to": [
        {
          "accountAlias": "@merchant_alpha/bank_sicredi",
          "amount": { "asset": "BRL", "value": "250.10" },
          "description": "Liquid Bank Deposit"
        },
        {
          "accountAlias": "@merchant_alpha/receivables/fees",
          "amount": { "asset": "BRL", "value": "49.90" },
          "description": "SumUp POS Terminal Rental",
          "metadata": {
            "charged_by": "SUMUP",
            "fee_type": "MACHINE_RENTAL"
          }
        }
      ]
    }
  }
}
```
   - **Step C:** Inter-plugin ACK:
     `POST /v1/internal/transactions/mark-processed` sent to `plugin-openfinance`.
