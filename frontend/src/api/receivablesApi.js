import axios from 'axios';

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8082';

const api = axios.create({
  baseURL: API_BASE_URL,
  timeout: 5000,
  headers: {
    'Content-Type': 'application/json'
  }
});

// Mock fallback data for standalone frontend development
const MOCK_ACCOUNTS = [
  {
    id: '01948df2-aaaa-7000-8000-000000000001',
    organization_id: '01948df2-0000-7000-8000-000000000001',
    ledger_id: '01948df2-0000-7000-8000-000000000002',
    account_alias: '@merchant_alpha/receivables',
    checking_account_alias: '@merchant_alpha/checking',
    name: 'Alpha Supermarket - Card Receivables',
    clearing_account_alias: '@external/acquirer_sumup',
    status: 'ACTIVE',
    created_at: '2026-11-20T14:50:00Z',
    metrics: {
      total_pending_gross: 1849200.0,
      total_pending_net: 1782900.0,
      total_settled: 8450000.0,
      pending_installments_count: 142
    }
  },
  {
    id: '01948df2-bbbb-7000-8000-000000000002',
    organization_id: '01948df2-0000-7000-8000-000000000001',
    ledger_id: '01948df2-0000-7000-8000-000000000002',
    account_alias: '@bravo_retail/receivables',
    checking_account_alias: '@bravo_retail/checking',
    name: 'Bravo Electronics - Multi-Acquirer Agenda',
    clearing_account_alias: '@external/acquirer_pagbank',
    status: 'ACTIVE',
    created_at: '2026-11-21T09:15:00Z',
    metrics: {
      total_pending_gross: 940300.0,
      total_pending_net: 908150.0,
      total_settled: 3200000.0,
      pending_installments_count: 86
    }
  }
];

const MOCK_CONNECTIONS = [
  {
    id: 'conn_sumup_01',
    receivables_account_id: '01948df2-aaaa-7000-8000-000000000001',
    acquirer_type: 'SUMUP',
    display_name: 'SumUp POS Terminal Fleet',
    merchant_code: 'SU-BRA-99120',
    status: 'ACTIVE',
    auto_sync_enabled: true,
    sync_interval_minutes: 60,
    last_sync_at: '2026-11-21T19:30:00Z',
    webhook_url: 'http://localhost:8082/v1/webhooks/sumup/conn_sumup_01'
  },
  {
    id: 'conn_aditum_01',
    receivables_account_id: '01948df2-aaaa-7000-8000-000000000001',
    acquirer_type: 'ADITUM',
    display_name: 'Aditum Multi-Acquirer Gateway',
    merchant_code: 'ADT-98214',
    status: 'ACTIVE',
    auto_sync_enabled: true,
    sync_interval_minutes: 30,
    last_sync_at: '2026-11-21T20:10:00Z',
    webhook_url: 'http://localhost:8082/v1/webhooks/aditum/conn_aditum_01'
  },
  {
    id: 'conn_pagbank_01',
    receivables_account_id: '01948df2-bbbb-7000-8000-000000000002',
    acquirer_type: 'PAGBANK',
    display_name: 'PagBank Online Checkout',
    merchant_code: 'PB-MERCH-4491',
    status: 'ACTIVE',
    auto_sync_enabled: true,
    sync_interval_minutes: 60,
    last_sync_at: '2026-11-21T18:00:00Z',
    webhook_url: 'http://localhost:8082/v1/webhooks/pagbank/conn_pagbank_01'
  }
];

const MOCK_INSTALLMENTS = [
  {
    id: 'inst_01',
    acquirer: 'SUMUP',
    nsu: '987654321',
    account_alias: '@merchant_alpha/receivables',
    installment_number: 1,
    total_installments: 3,
    card_brand: 'Mastercard Black',
    gross_amount: 1250.0,
    mdr_fee: 43.75,
    net_amount: 1206.25,
    expected_settlement_date: '2026-11-22',
    status: 'PENDING_SETTLEMENT',
    midaz_receivable_tx_id: 'tx_midaz_01948df2-0001',
    matcher_status: 'SOURCE_A_UPLOADED'
  },
  {
    id: 'inst_02',
    acquirer: 'ADITUM',
    nsu: '876543210',
    account_alias: '@merchant_alpha/receivables',
    installment_number: 2,
    total_installments: 6,
    card_brand: 'Visa Infinite',
    gross_amount: 3400.0,
    mdr_fee: 108.8,
    net_amount: 3291.2,
    expected_settlement_date: '2026-11-23',
    status: 'PENDING_SETTLEMENT',
    midaz_receivable_tx_id: 'tx_midaz_01948df2-0002',
    matcher_status: 'SOURCE_A_UPLOADED'
  },
  {
    id: 'inst_03',
    acquirer: 'SUMUP',
    nsu: '765432109',
    account_alias: '@merchant_alpha/receivables',
    installment_number: 1,
    total_installments: 1,
    card_brand: 'Elo Grafite',
    gross_amount: 45200.0,
    mdr_fee: 1356.0,
    net_amount: 43844.0,
    expected_settlement_date: '2026-11-20',
    status: 'RECONCILED',
    midaz_receivable_tx_id: 'tx_midaz_01948df2-0003',
    midaz_settlement_tx_id: 'tx_midaz_settle_01948df2-0003',
    matcher_status: 'MATCH_APPROVED'
  },
  {
    id: 'inst_04',
    acquirer: 'PAGBANK',
    nsu: '654321098',
    account_alias: '@bravo_retail/receivables',
    installment_number: 1,
    total_installments: 2,
    card_brand: 'Visa Platinum',
    gross_amount: 890.0,
    mdr_fee: 31.15,
    net_amount: 858.85,
    expected_settlement_date: '2026-11-24',
    status: 'PENDING_SETTLEMENT',
    midaz_receivable_tx_id: 'tx_midaz_01948df2-0004',
    matcher_status: 'SOURCE_A_UPLOADED'
  },
  {
    id: 'inst_05',
    acquirer: 'ADITUM',
    nsu: '543210987',
    account_alias: '@merchant_alpha/receivables',
    installment_number: 3,
    total_installments: 10,
    card_brand: 'Mastercard Platinum',
    gross_amount: 5200.0,
    mdr_fee: 171.6,
    net_amount: 5028.4,
    expected_settlement_date: '2026-11-25',
    status: 'PENDING_SETTLEMENT',
    midaz_receivable_tx_id: 'tx_midaz_01948df2-0005',
    matcher_status: 'SOURCE_A_UPLOADED'
  }
];

export const receivablesApi = {
  // Receivables Accounts
  async getAccounts() {
    try {
      const response = await api.get('/v1/accounts/receivables');
      return response.data;
    } catch {
      return MOCK_ACCOUNTS;
    }
  },

  async createAccount(data) {
    try {
      const response = await api.post('/v1/accounts/receivables', data);
      return response.data;
    } catch {
      const newAcc = {
        id: `01948df2-gen-${Date.now()}`,
        ...data,
        status: 'ACTIVE',
        created_at: new Date().toISOString(),
        metrics: {
          total_pending_gross: 0,
          total_pending_net: 0,
          total_settled: 0,
          pending_installments_count: 0
        }
      };
      MOCK_ACCOUNTS.push(newAcc);
      return newAcc;
    }
  },

  // Connections
  async getConnections(accountId) {
    try {
      const url = accountId
        ? `/v1/accounts/receivables/${accountId}/connections`
        : '/v1/connections';
      const response = await api.get(url);
      return response.data;
    } catch {
      if (accountId) {
        return MOCK_CONNECTIONS.filter((c) => c.receivables_account_id === accountId);
      }
      return MOCK_CONNECTIONS;
    }
  },

  async createConnection(accountId, data) {
    try {
      const response = await api.post(`/v1/accounts/receivables/${accountId}/connections`, data);
      return response.data;
    } catch {
      const newConn = {
        id: `conn_${data.acquirer_type.toLowerCase()}_${Date.now()}`,
        receivables_account_id: accountId,
        acquirer_type: data.acquirer_type,
        display_name: data.display_name,
        merchant_code: data.credentials?.merchant_code || 'CODE-1234',
        status: 'ACTIVE',
        auto_sync_enabled: data.auto_sync_enabled ?? true,
        sync_interval_minutes: data.sync_interval_minutes || 60,
        last_sync_at: new Date().toISOString(),
        webhook_url: `http://localhost:8082/v1/webhooks/${data.acquirer_type.toLowerCase()}/new`
      };
      MOCK_CONNECTIONS.push(newConn);
      return newConn;
    }
  },

  async triggerSync(connectionId) {
    try {
      const response = await api.post(`/v1/connections/${connectionId}/sync`);
      return response.data;
    } catch {
      return {
        status: 'SYNC_COMPLETED',
        connection_id: connectionId,
        synced_installments: 14,
        synced_at: new Date().toISOString()
      };
    }
  },

  // Calendar & Installments
  async getCalendar(accountId, params = {}) {
    try {
      const response = await api.get(`/v1/accounts/receivables/${accountId}/calendar`, { params });
      return response.data;
    } catch {
      return {
        receivables_account_id: accountId,
        total_pending_gross: 1849200.0,
        total_pending_net: 1782900.0,
        total_reconciled: 8450000.0,
        installments: MOCK_INSTALLMENTS
      };
    }
  },

  // Reconciliation summary with Matcher 6
  async getReconciliationAudit() {
    return {
      matcher_engine: 'Matcher 6.0.0-beta.1 (Port 4018)',
      context_slug: 'acquirers-settlement',
      health: 'HEALTHY',
      stats: {
        total_matched_pct: '99.4%',
        auto_cleared_count: 1420,
        pending_source_b: 38,
        discrepancies_flagged: 0
      },
      recent_settlements: [
        {
          batch_id: 'MATCH-BATCH-8901',
          acquirer: 'SUMUP',
          gross_amount: 45200.0,
          net_settlement: 43844.0,
          bank_credit_statement: 'CRÉDITO CARTÕES SUMUP REF: 20261120',
          reconciled_at: '2026-11-20T18:15:32Z',
          midaz_tx_id: 'tx_midaz_settle_01948df2-0003',
          status: 'SETTLED_EXTINGUISHED'
        },
        {
          batch_id: 'MATCH-BATCH-8902',
          acquirer: 'ADITUM',
          gross_amount: 12890.5,
          net_settlement: 12516.68,
          bank_credit_statement: 'ADITUM MEIOS DE PAGAMENTO D+1',
          reconciled_at: '2026-11-19T17:42:10Z',
          midaz_tx_id: 'tx_midaz_settle_01948df2-0008',
          status: 'SETTLED_EXTINGUISHED'
        }
      ]
    };
  }
};
