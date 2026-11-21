import { useState, useEffect } from 'react';

// @mui
import Box from '@mui/material/Box';
import Grid from '@mui/material/Grid';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import Button from '@mui/material/Button';
import TextField from '@mui/material/TextField';
import InputAdornment from '@mui/material/InputAdornment';
import CircularProgress from '@mui/material/CircularProgress';

// @project
import { BentoCard, KpiCard, StatusBadge, PillTabs } from '@/components/alignly';
import { receivablesApi } from '@/api/receivablesApi';

// @assets
import {
  IconCalendarEvent,
  IconCreditCard,
  IconTrendingUp,
  IconArrowUpRight,
  IconSearch,
  IconRefresh,
  IconBuildingBank
} from '@tabler/icons-react';

/***************************  RECEIVABLES DASHBOARD VIEW  ***************************/

export default function DashboardView() {
  const [activeTab, setActiveTab] = useState('all');
  const [loading, setLoading] = useState(false);
  const [installments, setInstallments] = useState([]);
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedAcquirer, setSelectedAcquirer] = useState('ALL');

  const tabs = [
    { label: 'All Operations', value: 'all' },
    { label: 'Pending Settlement (D+X)', value: 'pending' },
    { label: 'Reconciled', value: 'reconciled' }
  ];

  const loadData = async () => {
    setLoading(true);
    try {
      const data = await receivablesApi.getCalendar('01948df2-aaaa-7000-8000-000000000001');
      setInstallments(data.installments || []);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, []);

  const filteredInstallments = installments.filter((item) => {
    const matchesTab =
      activeTab === 'all' ||
      (activeTab === 'pending' && item.status === 'PENDING_SETTLEMENT') ||
      (activeTab === 'reconciled' && item.status === 'RECONCILED');

    const matchesAcquirer = selectedAcquirer === 'ALL' || item.acquirer === selectedAcquirer;

    const matchesSearch =
      item.nsu.toLowerCase().includes(searchQuery.toLowerCase()) ||
      item.card_brand.toLowerCase().includes(searchQuery.toLowerCase()) ||
      item.acquirer.toLowerCase().includes(searchQuery.toLowerCase());

    return matchesTab && matchesAcquirer && matchesSearch;
  });

  const formatCurrency = (val) =>
    new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(val);

  return (
    <Box sx={{ pb: 6 }}>
      {/* Header Bar */}
      <Stack
        direction={{ xs: 'column', md: 'row' }}
        justifyContent="space-between"
        alignItems={{ xs: 'flex-start', md: 'center' }}
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
            Receivables Agenda
          </Typography>
          <Typography variant="body2" sx={{ color: '#919EAB' }}>
            Multi-tenant card acquirers agenda orchestration &amp; D+X settlement calendar
          </Typography>
        </Box>

        <Stack direction="row" spacing={1.5} alignItems="center">
          <Button
            variant="outlined"
            onClick={loadData}
            startIcon={loading ? <CircularProgress size={16} /> : <IconRefresh size={16} />}
            sx={{
              borderRadius: '9999px',
              color: '#1D1D1D',
              borderColor: '#E8E8E8',
              bgcolor: '#FFFFFF',
              '&:hover': { bgcolor: '#F3F3F3', borderColor: '#D0D0D0' }
            }}
          >
            Refresh
          </Button>
          <PillTabs tabs={tabs} activeTab={activeTab} onChange={setActiveTab} />
        </Stack>
      </Stack>

      {/* KPI Bento Grid */}
      <Grid container spacing={3} sx={{ mb: 3 }}>
        <Grid item xs={12} md={4}>
          <KpiCard
            title="Total Gross Receivables (D+0)"
            value="R$ 1.849.200,00"
            subtitle="Accrued claims booked in Ledger"
            badgeText="+18.4% MoM"
            badgeVariant="lime"
            variant="white"
            icon={<IconTrendingUp size={20} />}
          />
        </Grid>
        <Grid item xs={12} md={4}>
          <KpiCard
            title="Net Settlement Liquidity"
            value="R$ 1.782.900,00"
            subtitle="After MDR deduction (avg 3.58%)"
            badgeText="D+30 Window"
            badgeVariant="success"
            variant="white"
            icon={<IconCreditCard size={20} />}
          />
        </Grid>
        <Grid item xs={12} md={4}>
          <KpiCard
            title="Extinguished into Liquid Cash"
            value="R$ 8.450.000,00"
            subtitle="Reconciled via Matcher 6"
            badgeText="Zero Duplication"
            badgeVariant="dark"
            variant="lime"
            icon={<IconBuildingBank size={20} />}
          />
        </Grid>
      </Grid>

      {/* Featured Bento Highlight (Clearing & Double Entry Status) */}
      <Grid container spacing={3} sx={{ mb: 4 }}>
        <Grid item xs={12} md={7}>
          <BentoCard variant="dark" sx={{ height: '100%', p: 3.5 }}>
            <Stack direction="row" justifyContent="space-between" alignItems="flex-start" sx={{ mb: 2.5 }}>
              <Box>
                <StatusBadge variant="lime" sx={{ mb: 1.5 }}>
                  Double-Entry Invariant Guarded
                </StatusBadge>
                <Typography
                  sx={{
                    fontFamily: "'Plus Jakarta Sans', sans-serif",
                    fontSize: '20px',
                    fontWeight: 700,
                    color: '#FFFFFF',
                    letterSpacing: '-0.02em',
                    mb: 1
                  }}
                >
                  Universal Clearing Counterbalance
                </Typography>
                <Typography sx={{ color: '#919EAB', fontSize: '13px', lineHeight: 1.6, maxWidth: 540 }}>
                  Sales ingested at D+0 are debited into{' '}
                  <Box component="span" sx={{ color: '#F0FD63', fontWeight: 600 }}>
                    @merchant/receivables
                  </Box>{' '}
                  and credited against{' '}
                  <Box component="span" sx={{ color: '#FFFFFF', fontWeight: 600 }}>
                    @external/acquirer_sumup
                  </Box>
                  . Upon Matcher 6 approval, claims are extinguished into liquid checking cash without double-credit inflation.
                </Typography>
              </Box>
              <Box
                sx={{
                  p: 1.5,
                  borderRadius: '16px',
                  bgcolor: '#262626',
                  color: '#F0FD63',
                  display: { xs: 'none', sm: 'flex' }
                }}
              >
                <IconBuildingBank size={28} />
              </Box>
            </Stack>

            <Stack direction="row" spacing={3} sx={{ pt: 2, borderTop: '1px solid rgba(255, 255, 255, 0.1)' }}>
              <Box>
                <Typography sx={{ color: '#919EAB', fontSize: '11px', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
                  Universal Clearance
                </Typography>
                <Typography sx={{ color: '#FFFFFF', fontWeight: 600, fontSize: '14px', mt: 0.5 }}>
                  @external/acquirer_sumup
                </Typography>
              </Box>
              <Box>
                <Typography sx={{ color: '#919EAB', fontSize: '11px', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
                  Liquid Checking
                </Typography>
                <Typography sx={{ color: '#FFFFFF', fontWeight: 600, fontSize: '14px', mt: 0.5 }}>
                  @merchant_alpha/checking
                </Typography>
              </Box>
              <Box>
                <Typography sx={{ color: '#919EAB', fontSize: '11px', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
                  Status
                </Typography>
                <Typography sx={{ color: '#F0FD63', fontWeight: 600, fontSize: '14px', mt: 0.5 }}>
                  Balanced 100%
                </Typography>
              </Box>
            </Stack>
          </BentoCard>
        </Grid>

        <Grid item xs={12} md={5}>
          <BentoCard variant="white" sx={{ height: '100%', p: 3.5, display: 'flex', flexDirection: 'column', justifyContent: 'space-between' }}>
            <Box>
              <Stack direction="row" justifyContent="space-between" alignItems="center" sx={{ mb: 2 }}>
                <Typography sx={{ fontFamily: "'Plus Jakarta Sans', sans-serif", fontWeight: 700, fontSize: '16px' }}>
                  Connected Acquirers
                </Typography>
                <StatusBadge variant="lime">
                  3 Active
                </StatusBadge>
              </Stack>

              <Stack spacing={2}>
                {[
                  { name: 'SumUp POS Fleet', type: 'SUMUP', status: 'Live Sync', rate: '3.50% MDR' },
                  { name: 'Aditum Gateway', type: 'ADITUM', status: 'Live Sync', rate: '3.20% MDR' },
                  { name: 'PagBank Online', type: 'PAGBANK', status: 'Scheduled', rate: '3.99% MDR' }
                ].map((acq) => (
                  <Box
                    key={acq.name}
                    sx={{
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'space-between',
                      p: 1.5,
                      borderRadius: '14px',
                      bgcolor: '#F3F3F3'
                    }}
                  >
                    <Stack direction="row" spacing={1.5} alignItems="center">
                      <Box
                        sx={{
                          width: 32,
                          height: 32,
                          borderRadius: '10px',
                          bgcolor: '#1D1D1D',
                          color: '#FFFFFF',
                          display: 'flex',
                          alignItems: 'center',
                          justifyContent: 'center',
                          fontWeight: 700,
                          fontSize: '11px'
                        }}
                      >
                        {acq.type[0]}
                      </Box>
                      <Box>
                        <Typography sx={{ fontSize: '13px', fontWeight: 600, color: '#1D1D1D' }}>
                          {acq.name}
                        </Typography>
                        <Typography sx={{ fontSize: '11px', color: '#919EAB' }}>
                          {acq.rate}
                        </Typography>
                      </Box>
                    </Stack>
                    <StatusBadge variant={acq.status === 'Live Sync' ? 'lime' : 'neutral'}>
                      {acq.status}
                    </StatusBadge>
                  </Box>
                ))}
              </Stack>
            </Box>

            <Button
              variant="text"
              href="/receivables/connections"
              endIcon={<IconArrowUpRight size={16} />}
              sx={{
                alignSelf: 'flex-start',
                mt: 2,
                p: 0,
                color: '#1D1D1D',
                fontWeight: 600,
                fontSize: '12px',
                '&:hover': { bgcolor: 'transparent', textDecoration: 'underline' }
              }}
            >
              Manage Acquirer Connections
            </Button>
          </BentoCard>
        </Grid>
      </Grid>

      {/* Main Operations & Settlements Table Bento Card */}
      <BentoCard variant="white" sx={{ p: 3 }}>
        <Stack
          direction={{ xs: 'column', sm: 'row' }}
          justifyContent="space-between"
          alignItems={{ xs: 'flex-start', sm: 'center' }}
          spacing={2}
          sx={{ mb: 3 }}
        >
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
              Settlement Operations &amp; Installments
            </Typography>
            <Typography variant="caption" sx={{ color: '#919EAB' }}>
              Real-time audit trail of D+X clearing schedule
            </Typography>
          </Box>

          <Stack direction="row" spacing={1.5} alignItems="center">
            {/* Acquirer Filter Pills */}
            {['ALL', 'SUMUP', 'ADITUM', 'PAGBANK'].map((acq) => (
              <Button
                key={acq}
                onClick={() => setSelectedAcquirer(acq)}
                sx={{
                  px: 2,
                  py: 0.5,
                  borderRadius: '9999px',
                  fontSize: '11px',
                  fontWeight: selectedAcquirer === acq ? 600 : 500,
                  bgcolor: selectedAcquirer === acq ? '#1D1D1D' : '#F3F3F3',
                  color: selectedAcquirer === acq ? '#FFFFFF' : '#1D1D1D',
                  '&:hover': {
                    bgcolor: selectedAcquirer === acq ? '#1D1D1D' : '#E8E8E8'
                  }
                }}
              >
                {acq}
              </Button>
            ))}

            <TextField
              size="small"
              placeholder="Search NSU or brand..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              InputProps={{
                startAdornment: (
                  <InputAdornment position="start">
                    <IconSearch size={16} color="#919EAB" />
                  </InputAdornment>
                ),
                sx: { borderRadius: '9999px', bgcolor: '#F3F3F3', fontSize: '12px', width: { xs: 160, sm: 200 } }
              }}
            />
          </Stack>
        </Stack>

        {/* Operations List */}
        <Stack spacing={1.5}>
          {filteredInstallments.map((inst) => (
            <Box
              key={inst.id}
              sx={{
                display: 'flex',
                flexDirection: { xs: 'column', md: 'row' },
                alignItems: { xs: 'flex-start', md: 'center' },
                justifyContent: 'space-between',
                p: 2,
                borderRadius: '16px',
                bgcolor: '#FFFFFF',
                border: '1px solid #E8E8E8',
                gap: 2,
                transition: 'all 0.15s ease',
                '&:hover': {
                  borderColor: '#D0D0D0',
                  boxShadow: '0 4px 16px rgba(0, 0, 0, 0.04)'
                }
              }}
            >
              {/* Col 1: Acquirer + NSU */}
              <Stack direction="row" spacing={2} alignItems="center">
                <Box
                  sx={{
                    width: 44,
                    height: 44,
                    borderRadius: '12px',
                    bgcolor: inst.acquirer === 'SUMUP' ? '#F0FD63' : inst.acquirer === 'ADITUM' ? '#1D1D1D' : '#919EAB',
                    color: inst.acquirer === 'SUMUP' ? '#1D1D1D' : '#FFFFFF',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    fontWeight: 700,
                    fontSize: '12px',
                    flexShrink: 0
                  }}
                >
                  {inst.acquirer.slice(0, 3)}
                </Box>
                <Box>
                  <Stack direction="row" spacing={1} alignItems="center">
                    <Typography sx={{ fontWeight: 700, fontSize: '13px', color: '#1D1D1D' }}>
                      NSU {inst.nsu}
                    </Typography>
                    <StatusBadge variant={inst.acquirer === 'SUMUP' ? 'lime' : 'neutral'}>
                      {inst.acquirer}
                    </StatusBadge>
                  </Stack>
                  <Typography sx={{ fontSize: '11px', color: '#919EAB', mt: 0.25 }}>
                    {inst.card_brand} • Installment {inst.installment_number}/{inst.total_installments}
                  </Typography>
                </Box>
              </Stack>

              {/* Col 2: Date + Schedule */}
              <Box>
                <Stack direction="row" spacing={1} alignItems="center">
                  <IconCalendarEvent size={15} color="#919EAB" />
                  <Typography sx={{ fontSize: '12px', color: '#1D1D1D', fontWeight: 600 }}>
                    Expected: {inst.expected_settlement_date}
                  </Typography>
                </Stack>
                <Typography sx={{ fontSize: '11px', color: '#919EAB', mt: 0.25 }}>
                  {inst.account_alias}
                </Typography>
              </Box>

              {/* Col 3: Gross / Fee / Net */}
              <Box sx={{ textAlign: { xs: 'left', md: 'right' } }}>
                <Typography sx={{ fontWeight: 700, fontSize: '14px', color: '#1D1D1D' }}>
                  {formatCurrency(inst.net_amount)} (Net)
                </Typography>
                <Typography sx={{ fontSize: '11px', color: '#919EAB' }}>
                  Gross: {formatCurrency(inst.gross_amount)} | Fee: {formatCurrency(inst.mdr_fee)}
                </Typography>
              </Box>

              {/* Col 4: Status Badge */}
              <Box sx={{ alignSelf: { xs: 'flex-start', md: 'center' } }}>
                <StatusBadge variant={inst.status === 'RECONCILED' ? 'lime' : 'pending'}>
                  {inst.status === 'RECONCILED' ? 'Reconciled & Cleared' : 'Pending D+X'}
                </StatusBadge>
              </Box>
            </Box>
          ))}
        </Stack>
      </BentoCard>
    </Box>
  );
}
