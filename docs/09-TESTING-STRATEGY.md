# Testing Strategy: Midaz Receivables Plugin

## 1. Testing Philosophy & Standards

In mission-critical financial ledger systems, **testing is not an afterthought—it is a core invariant**. The `midaz-plugin-receivables` maintains a strict testing pyramid:

```
               ▲
              / \
             /E2E\     End-to-End Simulation (Matcher + Midaz)
            /-----\
           / Integ \   PostgreSQL, HTTP Clients & Webhooks
          /---------\
         /   Unit    \ Pure Domain Logic, Accounting Invariants, Parsers
        /-------------\
```

---

## 2. Test Suites & Categorization

### 2.1. Unit Testing (Isolated Domain & Parsers)
* **Framework:** `testing` + `github.com/stretchr/testify`.
* **Zero External Dependencies:** No network, no database.
* **Scope:**
  - **MDR & Net Value Calculations:** Validates that `Gross - MDR == Net` down to the cent using `shopspring/decimal`.
  - **Installment Splitting:** Asserts that odd amounts (e.g. R$ 100 in 3 installments) account for remainder cents ($33.34 + 33.33 + 33.33 = 100.00$) without losing or creating currency.
  - **Acquirer Parser Tests:** Ingestion tests parsing anonymized JSON payloads from SumUp and PagBank.

### 2.2. Adapter & Database Integration Testing
* **Database Mocking:** `github.com/DATA-DOG/go-sqlmock` for repository query unit tests.
* **Real Database Integration:** Runs against Dockerized PostgreSQL in CI/CD.
* **Scope:**
  - Idempotent upsert of sales and installments on duplicate NSU.
  - Status transition transitions: `PENDING_SETTLEMENT` $\rightarrow$ `RECONCILED`.

### 2.3. Midaz & Matcher 6 Contract Testing
* **HTTP Mocking:** In-memory HTTP mock servers simulating Midaz and Matcher 6 endpoints.
* **Payload Verification:** Asserts that every transaction sent to Midaz strictly adheres to Midaz DSL:
  - Balanced `send.source.from` and `send.distribute.to`.
  - Correct `@external` clearing account formatting.
  - Proper metadata tags.

---

## 3. High-Priority Edge Cases

| Test Case | Scenario | Expected Outcome |
| :--- | :--- | :--- |
| **TC-01: Odd Cent Installment Split** | Sale of R$ 100.00 in 3 installments. | First installment = R$ 33.34, second = R$ 33.33, third = R$ 33.33. Sum strictly equals R$ 100.00. |
| **TC-02: Duplicate Sale Ingestion** | Same webhook payload delivered twice by SumUp. | Returns HTTP 200 without creating duplicate database rows or duplicate Midaz transactions. |
| **TC-03: Pre-Settlement Cancellation** | Customer cancels sale before D+30 payout. | Midaz reversal transaction executed; installment marked `CANCELLED`. |
| **TC-04: Matcher 6 Re-Delivery** | Matcher webhook delivered multiple times for same match. | First call settles receivable in Midaz; subsequent calls acknowledge idempotently with HTTP 200. |
| **TC-05: Out-of-Window Settlement** | Acquirer pays on D+35 instead of D+30. | Matcher 6 tolerance rule (or manual queue) matches the delayed credit and settles the specific installment. |
