import { useState } from 'react';

// @mui
import Box from '@mui/material/Box';
import Grid from '@mui/material/Grid';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import Button from '@mui/material/Button';

// @project
import { BentoCard, KpiCard, StatusBadge, PillTabs } from '@/components/alignly';

// @assets
import { IconArrowUpRight, IconPlus, IconTrendingUp, IconWallet, IconUsers, IconCreditCard } from '@tabler/icons-react';

/***************************  SAMPLE / DASHBOARD SHOWCASE  ***************************/

export default function SamplePage() {
  const [activeTab, setActiveTab] = useState('overview');

  const tabs = [
    { label: 'Overview', value: 'overview' },
    { label: 'Transactions', value: 'transactions' },
    { label: 'Settlements', value: 'settlements' },
    { label: 'Audit Log', value: 'audit' }
  ];

  return (
    <Box sx={{ pb: 6 }}>
      {/* Header Bar with Pill Navigation */}
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
            Dashboard
          </Typography>
          <Typography variant="body2" sx={{ color: '#919EAB' }}>
            Midaz Plugin Core Overview &amp; Real-time Operations
          </Typography>
        </Box>

        <Stack direction="row" spacing={1.5} alignItems="center">
          <PillTabs tabs={tabs} activeTab={activeTab} onChange={setActiveTab} />
          <Button
            variant="contained"
            color="primary"
            startIcon={<IconPlus size={16} />}
            sx={{
              bgcolor: '#1D1D1D',
              color: '#FFFFFF',
              px: 3,
              '&:hover': { bgcolor: '#000000' }
            }}
          >
            New Operation
          </Button>
        </Stack>
      </Stack>

      {/* Row 1: KPI Bento Grid (3 Columns) */}
      <Grid container spacing={3} sx={{ mb: 3 }}>
        <Grid size={{ xs: 12, md: 4 }}>
          <KpiCard
            title="Total Receivables"
            value="R$ 1.849.200,00"
            subtitle="vs last month"
            badgeText="+18.4%"
            badgeVariant="lime"
            variant="white"
            icon={<IconTrendingUp size={20} />}
          />
        </Grid>
        <Grid size={{ xs: 12, md: 4 }}>
          <KpiCard
            title="Active Contracts"
            value="342"
            subtitle="12 awaiting approval"
            badgeText="98% healthy"
            badgeVariant="success"
            variant="white"
            icon={<IconCreditCard size={20} />}
          />
        </Grid>
        <Grid size={{ xs: 12, md: 4 }}>
          <KpiCard
            title="Connected Accounts"
            value="18"
            subtitle="Pluggy & Aditum"
            badgeText="Operational"
            badgeVariant="lime"
            variant="white"
            icon={<IconWallet size={20} />}
          />
        </Grid>
      </Grid>

      {/* Row 2: Featured Bento Cards (2 Columns - Lime & Slate/Dark Highlights) */}
      <Grid container spacing={3} sx={{ mb: 3 }}>
        <Grid size={{ xs: 12, md: 6 }}>
          <KpiCard
            title="Settlements Pending Clearing"
            value="R$ 482.150,00"
            subtitle="D+1 Clearing Window scheduled for 18:00"
            badgeText="Scheduled"
            badgeVariant="dark"
            variant="lime"
            icon={<IconWallet size={20} />}
            action={
              <Box
                sx={{
                  width: 32,
                  height: 32,
                  borderRadius: '9999px',
                  bgcolor: '#1D1D1D',
                  color: '#FFFFFF',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  cursor: 'pointer'
                }}
              >
                <IconArrowUpRight size={18} />
              </Box>
            }
          />
        </Grid>

        <Grid size={{ xs: 12, md: 6 }}>
          <KpiCard
            title="Open Finance Consent Pipeline"
            value="94.2% Complete"
            subtitle="28 active sync jobs running across institutions"
            badgeText="Live Sync"
            badgeVariant="lime"
            variant="dark"
            icon={<IconUsers size={20} />}
            action={
              <Box
                sx={{
                  width: 32,
                  height: 32,
                  borderRadius: '9999px',
                  bgcolor: '#262626',
                  color: '#F0FD63',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  cursor: 'pointer'
                }}
              >
                <IconArrowUpRight size={18} />
              </Box>
            }
          />
        </Grid>
      </Grid>

      {/* Row 3: Standard Bento Container with Table / Activity List */}
      <BentoCard variant="white">
        <Stack direction="row" justifyContent="space-between" alignItems="center" sx={{ mb: 3 }}>
          <Box>
            <Typography
              variant="h6"
              sx={{
                fontFamily: "'Plus Jakarta Sans', sans-serif",
                fontWeight: 600,
                color: '#1D1D1D',
                letterSpacing: '-0.02em'
              }}
            >
              Recent Operations
            </Typography>
            <Typography variant="caption" sx={{ color: '#919EAB' }}>
              Real-time audit trail of plugin transactions &amp; contracts
            </Typography>
          </Box>
          <StatusBadge variant="lime">
            Live Stream
          </StatusBadge>
        </Stack>

        <Stack spacing={2}>
          {[
            { id: 'TX-901', org: 'SumUp Receivables Acquirer', type: 'Settlement', value: 'R$ 45.200,00', status: 'Concluído', variant: 'lime' },
            { id: 'TX-902', org: 'Banco do Brasil (OpenFinance)', type: 'Account Sync', value: '1.240 entries', status: 'Concluído', variant: 'lime' },
            { id: 'TX-903', org: 'Aditum Multi-Acquirer Gateway', type: 'Credit Card Liquidation', value: 'R$ 12.890,50', status: 'Pendente', variant: 'pending' },
            { id: 'TX-904', org: 'Pluggy Banking Connector', type: 'Consent Renewal', value: 'Active', status: 'Ativo', variant: 'success' }
          ].map((op) => (
            <Box
              key={op.id}
              sx={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
                p: 2,
                borderRadius: '16px',
                bgcolor: '#F3F3F3',
                transition: 'all 0.15s ease',
                '&:hover': { bgcolor: '#EAEAEA' }
              }}
            >
              <Stack direction="row" spacing={2} alignItems="center">
                <Box
                  sx={{
                    width: 40,
                    height: 40,
                    borderRadius: '12px',
                    bgcolor: '#FFFFFF',
                    border: '1px solid #E8E8E8',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    fontWeight: 700,
                    fontSize: '12px',
                    color: '#1D1D1D'
                  }}
                >
                  {op.id.split('-')[1]}
                </Box>
                <Box>
                  <Typography sx={{ fontWeight: 600, fontSize: '13px', color: '#1D1D1D' }}>
                    {op.org}
                  </Typography>
                  <Typography sx={{ fontSize: '11px', color: '#919EAB' }}>
                    {op.type}
                  </Typography>
                </Box>
              </Stack>

              <Stack direction="row" spacing={3} alignItems="center">
                <Typography sx={{ fontWeight: 600, fontSize: '13px', color: '#1D1D1D' }}>
                  {op.value}
                </Typography>
                <StatusBadge variant={op.variant}>
                  {op.status}
                </StatusBadge>
              </Stack>
            </Box>
          ))}
        </Stack>
      </BentoCard>
    </Box>
  );
}
