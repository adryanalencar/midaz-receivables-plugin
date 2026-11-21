import { useState, useEffect } from 'react';

// @mui
import Box from '@mui/material/Box';
import Grid from '@mui/material/Grid';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import Button from '@mui/material/Button';

// @project
import { BentoCard, KpiCard, StatusBadge } from '@/components/alignly';
import { receivablesApi } from '@/api/receivablesApi';

// @assets
import {
  IconArrowsShuffle,
  IconCheck,
  IconScale,
  IconBuildingBank,
  IconShieldCheck,
  IconArrowRight
} from '@tabler/icons-react';

/***************************  MATCHER 6 RECONCILIATION AUDIT VIEW  ***************************/

export default function ReconciliationView() {
  const [auditData, setAuditData] = useState(null);

  useEffect(() => {
    receivablesApi.getReconciliationAudit().then(setAuditData);
  }, []);

  const formatCurrency = (val) =>
    new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(val);

  return (
    <Box sx={{ pb: 6 }}>
      {/* Header */}
      <Stack
        direction={{ xs: 'column', sm: 'row' }}
        justifyContent="space-between"
        alignItems={{ xs: 'flex-start', sm: 'center' }}
        spacing={2}
        sx={{ mb: 4 }}
      >
        <Box>
          <Typography
            variant="h4"
            sx={{
              fontFamily: "'Plus Jakarta Sans', sans-serif",
              fontWeight: 700,
              color: '#1D1D1D',
              letterSpacing: '-0.03em',
              mb: 0.5
            }}
          >
            Matcher 6 Reconciliation
          </Typography>
          <Typography variant="body2" sx={{ color: '#919EAB' }}>
            Automated three-way bank settlement reconciliation &amp; balance sheet integrity
          </Typography>
        </Box>

        <StatusBadge variant="lime" icon={<IconShieldCheck size={14} />}>
          Matcher 6.0 Active
        </StatusBadge>
      </Stack>

      {/* KPI Bento Grid */}
      <Grid container spacing={3} sx={{ mb: 4 }}>
        <Grid item xs={12} md={4}>
          <KpiCard
            title="Reconciliation Match Rate"
            value="99.4%"
            subtitle="Auto-matched via Matcher 6"
            badgeText="Operational"
            badgeVariant="lime"
            variant="white"
            icon={<IconScale size={20} />}
          />
        </Grid>
        <Grid item xs={12} md={4}>
          <KpiCard
            title="Auto-Cleared Installments"
            value="1.420"
            subtitle="Extinguished into liquid cash"
            badgeText="Zero Duplication"
            badgeVariant="dark"
            variant="lime"
            icon={<IconCheck size={20} />}
          />
        </Grid>
        <Grid item xs={12} md={4}>
          <KpiCard
            title="Pending Source B Match"
            value="38"
            subtitle="Awaiting bank statement credit"
            badgeText="D-1 to D+1 Window"
            badgeVariant="pending"
            variant="white"
            icon={<IconArrowsShuffle size={20} />}
          />
        </Grid>
      </Grid>

      {/* Reconciliation Flow Architecture Showcase (Bento Card) */}
      <BentoCard variant="dark" sx={{ p: 4, mb: 4 }}>
        <Typography
          sx={{
            fontFamily: "'Plus Jakarta Sans', sans-serif",
            fontSize: '18px',
            fontWeight: 700,
            color: '#FFFFFF',
            letterSpacing: '-0.02em',
            mb: 1
          }}
        >
          Three-Way Double-Credit Prevention Architecture
        </Typography>
        <Typography sx={{ color: '#919EAB', fontSize: '13px', mb: 3 }}>
          How Receivables Plugin, Open Finance Plugin, and Matcher 6 prevent duplicating cash in Midaz
        </Typography>

        <Grid container spacing={2} alignItems="center">
          {/* Step 1: Source A */}
          <Grid item xs={12} md={3.6}>
            <Box sx={{ p: 2.5, borderRadius: '16px', bgcolor: '#262626', border: '1px solid rgba(255, 255, 255, 0.08)' }}>
              <StatusBadge variant="lime" sx={{ mb: 1.5 }}>
                Source A
              </StatusBadge>
              <Typography sx={{ color: '#FFFFFF', fontWeight: 600, fontSize: '14px', mb: 0.5 }}>
                Receivables Schedule
              </Typography>
              <Typography sx={{ color: '#919EAB', fontSize: '11px', lineHeight: 1.5 }}>
                Expected net settlement uploaded automatically at D+0 from card acquirer sales agenda.
              </Typography>
            </Box>
          </Grid>

          {/* Arrow */}
          <Grid item xs={12} md={0.6} sx={{ display: 'flex', justifyContent: 'center' }}>
            <IconArrowRight size={20} color="#F0FD63" />
          </Grid>

          {/* Step 2: Source B */}
          <Grid item xs={12} md={3.6}>
            <Box sx={{ p: 2.5, borderRadius: '16px', bgcolor: '#262626', border: '1px solid rgba(255, 255, 255, 0.08)' }}>
              <StatusBadge variant="dark" sx={{ mb: 1.5, bgcolor: '#1D1D1D', border: '1px solid #444' }}>
                Source B
              </StatusBadge>
              <Typography sx={{ color: '#FFFFFF', fontWeight: 600, fontSize: '14px', mb: 0.5 }}>
                Bank Statement Credit
              </Typography>
              <Typography sx={{ color: '#919EAB', fontSize: '11px', lineHeight: 1.5 }}>
                Intercepted by Plugin Open Finance via reconciliation.enabled flag without posting duplicate deposit.
              </Typography>
            </Box>
          </Grid>

          {/* Arrow */}
          <Grid item xs={12} md={0.6} sx={{ display: 'flex', justifyContent: 'center' }}>
            <IconArrowRight size={20} color="#F0FD63" />
          </Grid>

          {/* Step 3: Matcher + Midaz Extinguishment */}
          <Grid item xs={12} md={3.6}>
            <Box sx={{ p: 2.5, borderRadius: '16px', bgcolor: '#2C2C2C', border: '1px solid #F0FD63' }}>
              <StatusBadge variant="lime" sx={{ mb: 1.5 }}>
                Atomic Extinguishment
              </StatusBadge>
              <Typography sx={{ color: '#FFFFFF', fontWeight: 600, fontSize: '14px', mb: 0.5 }}>
                Midaz Ledger Posting
              </Typography>
              <Typography sx={{ color: '#F0FD63', fontSize: '11px', lineHeight: 1.5 }}>
                Receivable debited &amp; checking credited. Balance sheet remains perfectly balanced.
              </Typography>
            </Box>
          </Grid>
        </Grid>
      </BentoCard>

      {/* Reconciled Batches Table */}
      <BentoCard variant="white" sx={{ p: 3 }}>
        <Stack direction="row" justifyContent="space-between" alignItems="center" sx={{ mb: 3 }}>
          <Box>
            <Typography
              variant="h6"
              sx={{
                fontFamily: "'Plus Jakarta Sans', sans-serif",
                fontWeight: 700,
                color: '#1D1D1D',
                letterSpacing: '-0.02em'
              }}
            >
              Recent Reconciled Batches
            </Typography>
            <Typography variant="caption" sx={{ color: '#919EAB' }}>
              Settlements approved by Matcher 6 and committed to Midaz Ledger
            </Typography>
          </Box>
          <StatusBadge variant="lime">
            All Balanced
          </StatusBadge>
        </Stack>

        <Stack spacing={2}>
          {auditData?.recent_settlements?.map((batch) => (
            <Box
              key={batch.batch_id}
              sx={{
                p: 2.5,
                borderRadius: '16px',
                bgcolor: '#F3F3F3',
                border: '1px solid #E8E8E8',
                display: 'flex',
                flexDirection: { xs: 'column', md: 'row' },
                justifyContent: 'space-between',
                alignItems: { xs: 'flex-start', md: 'center' },
                gap: 2
              }}
            >
              <Stack direction="row" spacing={2} alignItems="center">
                <Box
                  sx={{
                    width: 44,
                    height: 44,
                    borderRadius: '12px',
                    bgcolor: '#1D1D1D',
                    color: '#F0FD63',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    fontWeight: 700,
                    fontSize: '12px'
                  }}
                >
                  <IconBuildingBank size={22} />
                </Box>
                <Box>
                  <Stack direction="row" spacing={1} alignItems="center">
                    <Typography sx={{ fontWeight: 700, fontSize: '14px', color: '#1D1D1D' }}>
                      {batch.batch_id}
                    </Typography>
                    <StatusBadge variant="lime">
                      {batch.acquirer}
                    </StatusBadge>
                  </Stack>
                  <Typography sx={{ fontSize: '12px', color: '#666666', mt: 0.25 }}>
                    Bank Extract: {batch.bank_credit_statement}
                  </Typography>
                  <Typography sx={{ fontSize: '11px', color: '#919EAB', mt: 0.25 }}>
                    Midaz TX: <code>{batch.midaz_tx_id}</code>
                  </Typography>
                </Box>
              </Stack>

              <Box sx={{ textAlign: { xs: 'left', md: 'right' } }}>
                <Typography sx={{ fontWeight: 700, fontSize: '15px', color: '#1D1D1D' }}>
                  {formatCurrency(batch.net_settlement)}
                </Typography>
                <Typography sx={{ fontSize: '11px', color: '#919EAB', mb: 1 }}>
                  Gross claim: {formatCurrency(batch.gross_amount)}
                </Typography>
                <StatusBadge variant="success">
                  Extinguished into Liquid Cash
                </StatusBadge>
              </Box>
            </Box>
          ))}
        </Stack>
      </BentoCard>
    </Box>
  );
}
