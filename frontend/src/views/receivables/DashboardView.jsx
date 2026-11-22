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
  IconBuildingBank,
  IconShieldCheck,
  IconDeviceDesktop
} from '@tabler/icons-react';

/***************************  RECEIVABLES DASHBOARD VIEW  ***************************/

export default function DashboardView() {
  const [activeTab, setActiveTab] = useState('all');
  const [loading, setLoading] = useState(false);
  const [installments, setInstallments] = useState([]);
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedAcquirer, setSelectedAcquirer] = useState('ALL');

  const tabs = [
    { label: 'Todas as Operações', value: 'all' },
    { label: 'Aguardando Depósito', value: 'pending' },
    { label: 'Conciliadas em Conta', value: 'reconciled' }
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
            Agenda de Recebíveis de Cartão
          </Typography>
          <Typography variant="body2" sx={{ color: '#919EAB' }}>
            Previsão consolidada de depósitos, taxas das credenciadoras e rastreabilidade por terminal (TID)
          </Typography>
        </Box>

        <Stack direction="row" spacing={1.5} alignItems="center" flexWrap="wrap" gap={1}>
          <Button
            variant="outlined"
            onClick={loadData}
            disabled={loading}
            startIcon={loading ? <CircularProgress size={16} /> : <IconRefresh size={16} />}
            sx={{
              borderRadius: '9999px',
              color: '#1D1D1D',
              borderColor: '#E8E8E8',
              bgcolor: '#FFFFFF',
              '&:hover': { bgcolor: '#F3F3F3', borderColor: '#D0D0D0' }
            }}
          >
            Atualizar Agenda
          </Button>
          <PillTabs tabs={tabs} activeTab={activeTab} onChange={setActiveTab} />
        </Stack>
      </Stack>

      {/* KPI Bento Grid */}
      <Grid container spacing={3} sx={{ mb: 3 }}>
        <Grid size={{ xs: 12, md: 4 }}>
          <KpiCard
            title="Total Bruto a Receber"
            value="R$ 1.849.200,00"
            subtitle="Vendas capturadas em todas as credenciadoras"
            badgeText="+18.4% este mês"
            badgeVariant="lime"
            variant="white"
            icon={<IconTrendingUp size={20} />}
          />
        </Grid>
        <Grid size={{ xs: 12, md: 4 }}>
          <KpiCard
            title="Líquido Previsto a Depositar"
            value="R$ 1.782.900,00"
            subtitle="Descontada taxa MDR média de 3.58%"
            badgeText="Janela D+30"
            badgeVariant="success"
            variant="white"
            icon={<IconCreditCard size={20} />}
          />
        </Grid>
        <Grid size={{ xs: 12, md: 4 }}>
          <KpiCard
            title="Vendas já Conciliadas em Caixa"
            value="R$ 8.450.000,00"
            subtitle="Depósitos confirmados via Open Finance"
            badgeText="Sem Duplicação"
            badgeVariant="dark"
            variant="lime"
            icon={<IconBuildingBank size={20} />}
          />
        </Grid>
      </Grid>

      {/* Featured Bento Highlight (Clearing & Double Entry Status) */}
      <Grid container spacing={3} sx={{ mb: 4 }}>
        <Grid size={{ xs: 12, lg: 7 }}>
          <BentoCard variant="dark" sx={{ height: '100%', p: 3.5 }}>
            <Stack direction="row" justifyContent="space-between" alignItems="flex-start" sx={{ mb: 2.5 }}>
              <Box>
                <StatusBadge variant="lime" sx={{ mb: 1.5 }}>
                  Conciliação Financeira Ativa
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
                  Liquidação Automática e Rastreabilidade Total
                </Typography>
                <Typography sx={{ color: '#919EAB', fontSize: '13px', lineHeight: 1.6, maxWidth: 540 }}>
                  Cada transação capturada por suas credenciadoras (SumUp, Aditum, PagBank) identifica individualmente o número do comprovante (NSU) e o terminal (TID). Ao cruzar com os depósitos bancários, o valor é liquidado no seu saldo disponível de forma transparente.
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
                <IconShieldCheck size={28} />
              </Box>
            </Stack>

            <Stack direction="row" spacing={3} sx={{ pt: 2, borderTop: '1px solid rgba(255, 255, 255, 0.1)' }}>
              <Box>
                <Typography sx={{ color: '#919EAB', fontSize: '11px', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
                  Credenciadoras
                </Typography>
                <Typography sx={{ color: '#FFFFFF', fontWeight: 600, fontSize: '14px', mt: 0.5 }}>
                  SumUp, Aditum e PagBank
                </Typography>
              </Box>
              <Box>
                <Typography sx={{ color: '#919EAB', fontSize: '11px', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
                  Identificadores
                </Typography>
                <Typography sx={{ color: '#FFFFFF', fontWeight: 600, fontSize: '14px', mt: 0.5 }}>
                  NSU + Terminal (TID)
                </Typography>
              </Box>
              <Box>
                <Typography sx={{ color: '#919EAB', fontSize: '11px', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
                  Conferência de Saldo
                </Typography>
                <Typography sx={{ color: '#F0FD63', fontWeight: 600, fontSize: '14px', mt: 0.5 }}>
                  100% Conciliado
                </Typography>
              </Box>
            </Stack>
          </BentoCard>
        </Grid>

        <Grid size={{ xs: 12, lg: 5 }}>
          <BentoCard variant="white" sx={{ height: '100%', p: 3.5, display: 'flex', flexDirection: 'column', justifyContent: 'space-between' }}>
            <Box>
              <Stack direction="row" justifyContent="space-between" alignItems="center" sx={{ mb: 2 }}>
                <Typography sx={{ fontFamily: "'Plus Jakarta Sans', sans-serif", fontWeight: 700, fontSize: '16px' }}>
                  Credenciadoras Ativas
                </Typography>
                <StatusBadge variant="lime">
                  3 Contratos
                </StatusBadge>
              </Stack>

              <Stack spacing={2}>
                {[
                  { name: 'SumUp Pagamentos', type: 'SUMUP', status: 'Sincronizado', rate: 'Taxa MDR 3.50%', terminals: '12 Terminais POS' },
                  { name: 'Aditum Gateway / TEF', type: 'ADITUM', status: 'Sincronizado', rate: 'Taxa MDR 3.20%', terminals: '8 Caixas TEF' },
                  { name: 'PagBank Checkout', type: 'PAGBANK', status: 'Programado', rate: 'Taxa MDR 3.99%', terminals: 'Online & POS' }
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
                          {acq.terminals} • {acq.rate}
                        </Typography>
                      </Box>
                    </Stack>
                    <StatusBadge variant={acq.status === 'Sincronizado' ? 'lime' : 'neutral'}>
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
              Gerenciar Contratos de Credenciadoras
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
              Previsão de Depósitos e Rastreio de Transações
            </Typography>
            <Typography variant="caption" sx={{ color: '#919EAB' }}>
              Identificação por NSU, Terminal de Captura (TID) e cronograma diário de liquidação
            </Typography>
          </Box>

          <Stack direction="row" spacing={1.5} alignItems="center" flexWrap="wrap" gap={1}>
            {/* Acquirer Filter Pills */}
            {[
              { id: 'ALL', label: 'Todas' },
              { id: 'SUMUP', label: 'SumUp' },
              { id: 'ADITUM', label: 'Aditum' },
              { id: 'PAGBANK', label: 'PagBank' }
            ].map((acq) => (
              <Button
                key={acq.id}
                onClick={() => setSelectedAcquirer(acq.id)}
                sx={{
                  px: 2,
                  py: 0.5,
                  borderRadius: '9999px',
                  fontSize: '11px',
                  fontWeight: selectedAcquirer === acq.id ? 600 : 500,
                  bgcolor: selectedAcquirer === acq.id ? '#1D1D1D' : '#F3F3F3',
                  color: selectedAcquirer === acq.id ? '#FFFFFF' : '#1D1D1D',
                  '&:hover': {
                    bgcolor: selectedAcquirer === acq.id ? '#1D1D1D' : '#E8E8E8'
                  }
                }}
              >
                {acq.label}
              </Button>
            ))}

            <TextField
              size="small"
              placeholder="Buscar NSU, terminal ou bandeira..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              InputProps={{
                startAdornment: (
                  <InputAdornment position="start">
                    <IconSearch size={16} color="#919EAB" />
                  </InputAdornment>
                ),
                sx: { borderRadius: '9999px', bgcolor: '#F3F3F3', fontSize: '12px', width: { xs: 200, sm: 260 } }
              }}
            />
          </Stack>
        </Stack>

        {/* Operations List */}
        <Stack spacing={1.5}>
          {filteredInstallments.map((inst, idx) => {
            const terminalMock = inst.acquirer === 'SUMUP' ? `POS-0${(idx % 4) + 1}` : inst.acquirer === 'ADITUM' ? `TEF-CX0${(idx % 3) + 1}` : `ONLINE-CHK`;

            return (
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
                {/* Col 1: Acquirer + NSU + Terminal */}
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
                        NSU #{inst.nsu}
                      </Typography>
                      <StatusBadge variant={inst.acquirer === 'SUMUP' ? 'lime' : 'neutral'}>
                        {inst.acquirer}
                      </StatusBadge>
                    </Stack>
                    <Stack direction="row" spacing={1} alignItems="center" sx={{ mt: 0.25 }}>
                      <Stack direction="row" spacing={0.5} alignItems="center">
                        <IconDeviceDesktop size={13} color="#919EAB" />
                        <Typography sx={{ fontSize: '11px', color: '#666666', fontWeight: 600 }}>
                          Terminal: {terminalMock}
                        </Typography>
                      </Stack>
                      <Typography sx={{ fontSize: '11px', color: '#919EAB' }}>
                        • {inst.card_brand} • Parcela {inst.installment_number}/{inst.total_installments}
                      </Typography>
                    </Stack>
                  </Box>
                </Stack>

                {/* Col 2: Date + Schedule */}
                <Box>
                  <Stack direction="row" spacing={1} alignItems="center">
                    <IconCalendarEvent size={15} color="#919EAB" />
                    <Typography sx={{ fontSize: '12px', color: '#1D1D1D', fontWeight: 600 }}>
                      Previsão: {inst.expected_settlement_date}
                    </Typography>
                  </Stack>
                  <Typography sx={{ fontSize: '11px', color: '#919EAB', mt: 0.25 }}>
                    Conta Corrente Principal
                  </Typography>
                </Box>

                {/* Col 3: Gross / Fee / Net */}
                <Box sx={{ textAlign: { xs: 'left', md: 'right' } }}>
                  <Typography sx={{ fontWeight: 700, fontSize: '14px', color: '#1D1D1D' }}>
                    {formatCurrency(inst.net_amount)} (Líquido)
                  </Typography>
                  <Typography sx={{ fontSize: '11px', color: '#919EAB' }}>
                    Bruto: {formatCurrency(inst.gross_amount)} | Taxa MDR: {formatCurrency(inst.mdr_fee)}
                  </Typography>
                </Box>

                {/* Col 4: Status Badge */}
                <Box sx={{ alignSelf: { xs: 'flex-start', md: 'center' } }}>
                  <StatusBadge variant={inst.status === 'RECONCILED' ? 'lime' : 'pending'}>
                    {inst.status === 'RECONCILED' ? 'Conciliado em Conta' : 'Aguardando Depósito'}
                  </StatusBadge>
                </Box>
              </Box>
            );
          })}
        </Stack>
      </BentoCard>
    </Box>
  );
}
