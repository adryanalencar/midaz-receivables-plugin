import { useState, useEffect } from 'react';

// @mui
import Box from '@mui/material/Box';
import Grid from '@mui/material/Grid';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import Button from '@mui/material/Button';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import TextField from '@mui/material/TextField';

// @project
import { BentoCard, StatusBadge } from '@/components/alignly';
import { receivablesApi } from '@/api/receivablesApi';

// @assets
import {
  IconPlus,
  IconBuildingBank,
  IconWallet,
  IconArrowsExchange
} from '@tabler/icons-react';

/***************************  RECEIVABLES ACCOUNTS & CLEARING VIEW  ***************************/

export default function AccountsView() {
  const [accounts, setAccounts] = useState([]);
  const [openModal, setOpenModal] = useState(false);

  // Form state
  const [accountAlias, setAccountAlias] = useState('');
  const [checkingAlias, setCheckingAlias] = useState('');
  const [accountName, setAccountName] = useState('');
  const [clearingAlias, setClearingAlias] = useState('SumUp Pagamentos');

  const loadAccounts = async () => {
    const data = await receivablesApi.getAccounts();
    setAccounts(data || []);
  };

  useEffect(() => {
    loadAccounts();
  }, []);

  const handleCreate = async () => {
    if (!accountAlias || !checkingAlias) return;
    await receivablesApi.createAccount({
      account_alias: accountAlias,
      checking_account_alias: checkingAlias,
      name: accountName || accountAlias,
      clearing_account_alias: clearingAlias
    });
    setOpenModal(false);
    setAccountAlias('');
    setCheckingAlias('');
    setAccountName('');
    loadAccounts();
  };

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
            Carteiras de Recebimento de Vendas
          </Typography>
          <Typography variant="body2" sx={{ color: '#919EAB' }}>
            Controle de saldo a receber de cartões e contas correntes bancárias associadas
          </Typography>
        </Box>

        <Button
          variant="contained"
          onClick={() => setOpenModal(true)}
          startIcon={<IconPlus size={16} />}
          sx={{
            borderRadius: '9999px',
            bgcolor: '#1D1D1D',
            color: '#FFFFFF',
            px: 3,
            '&:hover': { bgcolor: '#000000' }
          }}
        >
          Cadastrar Nova Carteira
        </Button>
      </Stack>

      {/* Grid of Accounts (3 cards per row on desktop/laptop) */}
      <Grid container spacing={3}>
        {accounts.map((acc) => (
          <Grid size={{ xs: 12, sm: 6, lg: 4 }} key={acc.id}>
            <BentoCard
              variant="white"
              sx={{
                p: 3.5,
                height: '100%',
                display: 'flex',
                flexDirection: 'column',
                justifyContent: 'space-between'
              }}
            >
              <Box>
                <Stack direction="row" justifyContent="space-between" alignItems="flex-start" sx={{ mb: 2 }}>
                  <Stack direction="row" spacing={1.5} alignItems="center">
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
                        fontSize: '13px'
                      }}
                    >
                      <IconWallet size={22} />
                    </Box>
                    <Box>
                      <Typography
                        sx={{
                          fontFamily: "'Plus Jakarta Sans', sans-serif",
                          fontWeight: 700,
                          fontSize: '16px',
                          color: '#1D1D1D'
                        }}
                      >
                        {acc.name}
                      </Typography>
                      <Typography sx={{ fontSize: '11px', color: '#919EAB' }}>
                        Carteira Ativa de Maquininhas
                      </Typography>
                    </Box>
                  </Stack>
                  <StatusBadge variant="lime">
                    Ativa
                  </StatusBadge>
                </Stack>

                {/* Account Details Box */}
                <Box sx={{ p: 2, borderRadius: '14px', bgcolor: '#F3F3F3', mb: 2.5 }}>
                  <Stack spacing={1.2}>
                    <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                      <Stack direction="row" spacing={1} alignItems="center">
                        <IconBuildingBank size={15} color="#666666" />
                        <Typography sx={{ fontSize: '12px', color: '#666666' }}>Conta de Depósito:</Typography>
                      </Stack>
                      <Typography sx={{ fontSize: '12px', fontWeight: 600, color: '#1D1D1D' }}>
                        Banco Itaú PJ
                      </Typography>
                    </Box>

                    <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                      <Stack direction="row" spacing={1} alignItems="center">
                        <IconArrowsExchange size={15} color="#666666" />
                        <Typography sx={{ fontSize: '12px', color: '#666666' }}>Operadora Principal:</Typography>
                      </Stack>
                      <Typography sx={{ fontSize: '12px', fontWeight: 600, color: '#15803d' }}>
                        SumUp / Aditum
                      </Typography>
                    </Box>
                  </Stack>
                </Box>
              </Box>

              <Stack
                direction="row"
                justifyContent="space-between"
                alignItems="center"
                sx={{ pt: 2, borderTop: '1px solid #E8E8E8' }}
              >
                <Box>
                  <Typography sx={{ fontSize: '11px', color: '#919EAB' }}>Saldo Bruto a Receber</Typography>
                  <Typography sx={{ fontSize: '18px', fontWeight: 700, color: '#1D1D1D' }}>
                    {formatCurrency(acc.balance)}
                  </Typography>
                </Box>

                <StatusBadge variant="lime">
                  100% Conciliado
                </StatusBadge>
              </Stack>
            </BentoCard>
          </Grid>
        ))}
      </Grid>

      {/* Create Account Modal */}
      <Dialog
        open={openModal}
        onClose={() => setOpenModal(false)}
        maxWidth="xs"
        fullWidth
        slotProps={{
          paper: {
            sx: {
              borderRadius: '24px',
              p: 1,
              bgcolor: '#FFFFFF',
              boxShadow: '0 20px 40px rgba(0,0,0,0.1)'
            }
          }
        }}
      >
        <DialogTitle sx={{ fontFamily: "'Plus Jakarta Sans', sans-serif", fontWeight: 700, fontSize: '18px' }}>
          Cadastrar Nova Carteira de Recebíveis
        </DialogTitle>
        <DialogContent>
          <Stack spacing={2.5} sx={{ mt: 1 }}>
            <TextField
              label="Nome Amigável da Carteira"
              placeholder="Ex: Vendas Filial Centro"
              value={accountName}
              onChange={(e) => setAccountName(e.target.value)}
              fullWidth
              size="small"
            />

            <TextField
              label="Código Interno da Carteira"
              placeholder="Ex: filial_centro_recebiveis"
              value={accountAlias}
              onChange={(e) => setAccountAlias(e.target.value)}
              fullWidth
              size="small"
            />

            <TextField
              label="Conta Corrente de Depósito"
              placeholder="Ex: Banco Itaú Ag 0812 CC 18920-4"
              value={checkingAlias}
              onChange={(e) => setCheckingAlias(e.target.value)}
              fullWidth
              size="small"
            />
          </Stack>
        </DialogContent>
        <DialogActions sx={{ px: 3, pb: 3 }}>
          <Button
            onClick={() => setOpenModal(false)}
            sx={{ borderRadius: '9999px', color: '#666666' }}
          >
            Cancelar
          </Button>
          <Button
            variant="contained"
            onClick={handleCreate}
            disabled={!accountAlias || !checkingAlias}
            sx={{
              borderRadius: '9999px',
              bgcolor: '#1D1D1D',
              color: '#FFFFFF',
              px: 3,
              '&:hover': { bgcolor: '#000000' }
            }}
          >
            Criar Carteira
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}
