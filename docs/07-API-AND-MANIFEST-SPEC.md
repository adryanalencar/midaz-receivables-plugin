# API Specification & Midaz Console Manifest

## 1. REST Endpoints Summary

| Method | Path | Description | Auth |
| :--- | :--- | :--- | :--- |
| `GET` | `/health` | Service liveness probe | None |
| `GET` | `/ready` | Service readiness probe (DB + Midaz ping) | None |
| `GET` | `/api/manifest` | Midaz Console microfrontend plugin descriptor | None |
| `POST` | `/v1/accounts/receivables` | Provisions a new receivables account in Midaz | Bearer / API Key |
| `GET` | `/v1/accounts/receivables` | Lists all managed receivables accounts | Bearer / API Key |
| `GET` | `/v1/accounts/receivables/:id` | Retrieves receivables account details | Bearer / API Key |
| `POST` | `/v1/accounts/receivables/:id/connections` | Links an acquirer connection (credentials) to account | Bearer / API Key |
| `GET` | `/v1/accounts/receivables/:id/connections` | Lists all acquirer connections bound to account | Bearer / API Key |
| `POST` | `/v1/connections/:id/sync` | Manually triggers sync for a specific connection | Bearer / API Key |
| `GET` | `/v1/accounts/receivables/:id/calendar` | Retrieves future receivables calendar and settlements | Bearer / API Key |
| `POST` | `/v1/webhooks/sumup/:connection_id` | Inbound webhook for SumUp events (HMAC signed) | Webhook Signature |
| `POST` | `/v1/webhooks/pagbank/:connection_id` | Inbound webhook for PagBank events | Webhook Token |
| `POST` | `/v1/webhooks/matcher` | Receives approved settlement matches from Matcher 6 | Matcher Secret / Token |

---

## 2. Receivables Account & Connection APIs

### 2.1. Create Receivables Account (`POST /v1/accounts/receivables`)
Provisions an account in Midaz with `type: "asset"` and stores the local mapping.

**Request:**
```json
{
  "organization_id": "01948df2-0000-7000-8000-000000000001",
  "ledger_id": "01948df2-0000-7000-8000-000000000002",
  "account_alias": "@merchant_alpha/receivables",
  "checking_account_alias": "@merchant_alpha/checking",
  "name": "Receivables Wallet - Merchant Alpha"
}
```

**Response (HTTP 201 Created):**
```json
{
  "id": "01948df2-aaaa-7000-8000-000000000001",
  "organization_id": "01948df2-0000-7000-8000-000000000001",
  "ledger_id": "01948df2-0000-7000-8000-000000000002",
  "midaz_account_id": "01948df2-bbbb-7000-8000-000000000002",
  "account_alias": "@merchant_alpha/receivables",
  "checking_account_alias": "@merchant_alpha/checking",
  "name": "Receivables Wallet - Merchant Alpha",
  "status": "ACTIVE",
  "created_at": "2026-11-20T14:50:00Z"
}
```

---

### 2.2. Link Acquirer Connection (`POST /v1/accounts/receivables/:id/connections`)
Associates specific acquirer credentials to the target Receivables Account.

**Request:**
```json
{
  "acquirer_type": "SUMUP",
  "display_name": "SumUp Main POS Terminal",
  "credentials": {
    "api_key": "sup_live_abcd1234efgh5678",
    "merchant_code": "MC123456"
  },
  "auto_sync_enabled": true,
  "sync_interval_minutes": 60
}
```

**Response (HTTP 201 Created):**
```json
{
  "id": "01948df2-cccc-7000-8000-000000000003",
  "receivables_account_id": "01948df2-aaaa-7000-8000-000000000001",
  "acquirer_type": "SUMUP",
  "display_name": "SumUp Main POS Terminal",
  "status": "ACTIVE",
  "auto_sync_enabled": true,
  "sync_interval_minutes": 60,
  "webhook_url": "https://api.platform.com/v1/webhooks/sumup/01948df2-cccc-7000-8000-000000000003",
  "created_at": "2026-11-20T14:52:00Z"
}
```

---

### 2.3. Query Receivables Calendar (`GET /v1/accounts/receivables/:id/calendar`)

**Query Parameters:**
* `status`: `PENDING_SETTLEMENT`, `RECONCILED`, `OVERDUE`
* `start_date`: `2026-12-01`
* `end_date`: `2026-12-31`
* `acquirer`: `SUMUP`

**Response (HTTP 200 OK):**
```json
{
  "receivables_account_id": "01948df2-aaaa-7000-8000-000000000001",
  "account_alias": "@merchant_alpha/receivables",
  "total_pending_gross": "15000.00",
  "total_pending_net": "14550.00",
  "total_reconciled": "8500.00",
  "installments": [
    {
      "id": "01948df2-dddd-7000-8000-000000000004",
      "connection_id": "01948df2-cccc-7000-8000-000000000003",
      "acquirer": "SUMUP",
      "nsu": "987654321",
      "installment_number": 1,
      "total_installments": 3,
      "gross_amount": "100.00",
      "mdr_fee": "3.00",
      "net_amount": "97.00",
      "expected_settlement_date": "2026-12-20",
      "actual_settlement_date": null,
      "status": "PENDING_SETTLEMENT",
      "midaz_receivable_tx_id": "01948df2-eeee-7000-8000-000000000005",
      "midaz_settlement_tx_id": null
    }
  ]
}
```

---

## 3. Midaz Console Manifest (`GET /api/manifest`)

```json
{
  "name": "midaz-plugin-receivables",
  "version": "1.0.0",
  "displayName": "Acquirer Receivables & Settlement",
  "description": "Multi-tenant card acquirers agenda orchestration and automated three-way bank settlement reconciliation with Matcher 6",
  "category": "reconciliation",
  "icon": "CreditCard",
  "routes": [
    {
      "path": "/receivables",
      "label": "Receivables Agenda",
      "componentUrl": "/remoteEntry.js",
      "scope": "midazReceivables",
      "module": "./ReceivablesDashboard"
    },
    {
      "path": "/receivables/connections",
      "label": "Acquirer Connections",
      "componentUrl": "/remoteEntry.js",
      "scope": "midazReceivables",
      "module": "./ConnectionsManager"
    },
    {
      "path": "/receivables/reconciliation",
      "label": "Matcher Reconciliation",
      "componentUrl": "/remoteEntry.js",
      "scope": "midazReceivables",
      "module": "./ReconciliationView"
    }
  ],
  "configuration": {
    "supportedAcquirers": ["SUMUP", "PAGBANK"],
    "matcherEngineVersion": "6.0.0-beta.1"
  }
}
```
