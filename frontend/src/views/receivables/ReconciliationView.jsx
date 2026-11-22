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
  IconCheck,
  IconScale,
  IconBuildingBank,
  IconShieldCheck,
  IconArrowRight,
  IconCreditCard,
  IconCash
} from '@tabler/icons-react';

/***************************  RECONCILIATION AUDIT VIEW  ***************************/

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
            Conciliação de Vendas e Depósitos Bancários
          </Typography>
          <Typography variant="body2" sx={{ color: '#919EAB' }}>
            Auditoria automática entre as vendas nas maquininhas e os créditos na sua conta corrente
          </Typography>
        </Box>

        <StatusBadge variant="lime" icon={<IconShieldCheck size={14} />}>
          Motor de Conciliação Ativo
        </StatusBadge>
      </Stack>

      {/* KPI Bento Grid */}
      <Grid container spacing={3} sx={{ mb: 4 }}>
        <Grid size={{ xs: 12, md: 4 }}>
          <KpiCard
            title="Taxa de Precisão na Conciliação"
            value="99.4%"
            subtitle="Conferência automática sem divergências"
            badgeText="Operação Normal"
            badgeVariant="lime"
            variant="white"
            icon={<IconScale size={20} />}
          />
        </Grid>
        <Grid size={{ xs: 12, md: 4 }}>
          <KpiCard
            title="Vendas Conciliadas no Mês"
            value="1.420"
            subtitle="Transferidas diretamente para seu caixa"
            badgeText="Sem Duplicações"
            badgeVariant="dark"
            variant="lime"
            icon={<IconCheck size={20} />}
          />
        </Grid>
        <Grid size={{ xs: 12, md: 4 }}>
          <KpiCard
            title="Aguardando Depósito Bancário"
            value="38"
            subtitle="Dentro do prazo normal das operadoras"
            badgeText="Janela D+X"
            badgeVariant="pending"
            variant="white"
            icon={<IconBuildingBank size={20} />}
          />
        </Grid>
      </Grid>

      {/* Reconciliation Flow Visual Card */}
      <BentoCard variant="white" sx={{ p: 4, mb: 4 }}>
        <Typography
          variant="h6"
          sx={{
            fontFamily: "'Plus Jakarta Sans', sans-serif",
            fontWeight: 700,
            color: '#1D1D1D',
            letterSpacing: '-0.02em',
            mb: 1
          }}
        >
          Como Funciona a Conciliação Automática
        </Typography>
        <Typography variant="body2" sx={{ color: '#919EAB', mb: 3.5 }}>
          Fluxo ponta a ponta que protege o saldo da sua empresa contra duplicidades
        </Typography>

        <Grid container spacing={2} alignItems="center">
          {/* Step 1 */}
          <Grid size={{ xs: 12, md: 3.6 }}>
            <Box sx={{ p: 2.5, borderRadius: '16px', bgcolor: '#F3F3F3', border: '1px solid #E8E8E8' }}>
              <Stack direction="row" spacing={1.5} alignItems="center" sx={{ mb: 1 }}>
                <Box sx={{ p: 1, borderRadius: '10px', bgcolor: '#1D1D1D', color: '#F0FD63', display: 'flex' }}>
                  <IconCreditCard size={18} />
                </Box>
                <Typography sx={{ fontWeight: 700, fontSize: '14px', color: '#1D1D1D' }}>
                  1. Venda na Maquininha
                </Typography>
              </Stack>
              <Typography sx={{ fontSize: '12px', color: '#666666', lineHeight: 1.5 }}>
                A venda é registrada com bandeira, parcelas e taxa MDR. O valor entra na sua agenda como a receber.
              </Typography>
            </Box>
          </Grid>

          <Grid size={{ xs: 12, md: 0.6 }} sx={{ display: { xs: 'none', md: 'flex' }, justifyContent: 'center' }}>
            <IconArrowRight size={20} color="#919EAB" />
          </Grid>

          {/* Step 2 */}
          <Grid size={{ xs: 12, md: 3.6 }}>
            <Box sx={{ p: 2.5, borderRadius: '16px', bgcolor: '#F3F3F3', border: '1px solid #E8E8E8' }}>
              <Stack direction="row" spacing={1.5} alignItems="center" sx={{ mb: 1 }}>
                <Box sx={{ p: 1, borderRadius: '10px', bgcolor: '#1D1D1D', color: '#FFFFFF', display: 'flex' }}>
                  <IconBuildingBank size={18} />
                </Box>
                <Typography sx={{ fontWeight: 700, fontSize: '14px', color: '#1D1D1D' }}>
                  2. Depósito Bancário
                </Typography>
              </Stack>
              <Typography sx={{ fontSize: '12px', color: '#666666', lineHeight: 1.5 }}>
                Na data prevista, a adquirente deposita o valor líquido no banco. O Midaz identifica o crédito pelo Open Finance.
              </Typography>
            </Box>
          </Grid>

          <Grid size={{ xs: 12, md: 0.6 }} sx={{ display: { xs: 'none', md: 'flex' }, justifyContent: 'center' }}>
            <IconArrowRight size={20} color="#919EAB" />
          </Grid>

          {/* Step 3 */}
          <Grid size={{ xs: 12, md: 3.6 }}>
            <Box sx={{ p: 2.5, borderRadius: '16px', bgcolor: '#F0FD63', border: '1px solid #D8E645' }}>
              <Stack direction="row" spacing={1.5} alignItems="center" sx={{ mb: 1 }}>
                <Box sx={{ p: 1, borderRadius: '10px', bgcolor: '#1D1D1D', color: '#F0FD63', display: 'flex' }}>
                  <IconCash size={18} />
                </Box>
                <Typography sx={{ fontWeight: 700, fontSize: '14px', color: '#1D1D1D' }}>
                  3. Saldo Liberado
                </Typography>
              </Stack>
              <Typography sx={{ fontSize: '12px', color: '#1D1D1D', lineHeight: 1.5 }}>
                O valor a receber é baixado com sucesso e o saldo disponível é atualizado com segurança total.
              </Typography>
            </Box>
          </Grid>
        </Grid>
      </BentoCard>

      {/* Detailed Audit Table */}
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
              Auditoria de Conciliações Recentes
            </Typography>
            <Typography variant="caption" sx={{ color: '#919EAB' }}>
              Conferência de transações liquidadas entre maquininhas e extratos
            </Typography>
          </Box>
          <StatusBadge variant="lime">
            Auditoria Ativa
          </StatusBadge>
        </Stack>

        <Stack spacing={1.5}>
          {[
            {
              id: '1',
              desc: 'Venda Débito Visa • NSU 998231',
              acquirer: 'SumUp POS',
              expected: 150.0,
              deposited: 145.5,
              fee: 4.5,
              date: '21/11/2026',
              status: 'CONCILIADO'
            },
            {
              id: '2',
              desc: 'Venda Crédito 2x Mastercard • NSU 551920',
              acquirer: 'Aditum Gateway',
              expected: 480.0,
              deposited: 462.24,
              fee: 17.76,
              date: '21/11/2026',
              status: 'CONCILIADO'
            },
            {
              id: '3',
              desc: 'Venda Débito Elo • NSU 331892',
              acquirer: 'PagBank Checkout',
              expected: 220.0,
              deposited: 213.4,
              fee: 6.6,
              date: '20/11/2026',
              status: 'CONCILIADO'
            }
          ].map((item) => (
            <Box
              key={item.id}
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
              <Stack direction="row" spacing={2} alignItems="center">
                <Box
                  sx={{
                    width: 44,
                    height: 44,
                    borderRadius: '12px',
                    bgcolor: '#F0FD63',
                    color: '#1D1D1D',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    fontWeight: 700,
                    fontSize: '12px',
                    flexShrink: 0
                  }}
                >
                  <IconCheck size={20} />
                </Box>
                <Box>
                  <Typography sx={{ fontWeight: 700, fontSize: '13px', color: '#1D1D1D' }}>
                    {item.desc}
                  </Typography>
                  <Typography sx={{ fontSize: '11px', color: '#919EAB', mt: 0.25 }}>
                    Operadora: {item.acquirer} • Data: {item.date}
                  </Typography>
                </Box>
              </Stack>

              <Box sx={{ textAlign: { xs: 'left', md: 'right' } }}>
                <Typography sx={{ fontWeight: 700, fontSize: '14px', color: '#15803d' }}>
                  {formatCurrency(item.deposited)} (Creditado)
                </Typography>
                <Typography sx={{ fontSize: '11px', color: '#919EAB' }}>
                  Bruto: {formatCurrency(item.expected)} | Taxa: {formatCurrency(item.fee)}
                </Typography>
              </Box>

              <Box sx={{ alignSelf: { xs: 'flex-start', md: 'center' } }}>
                <StatusBadge variant="lime">
                  Conciliado com Sucesso
                </StatusBadge>
              </Box>
            </Box>
          ))}
        </Stack>
      </BentoCard>
    </Box>
  );
}
