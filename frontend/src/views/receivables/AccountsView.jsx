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
  const [clearingAlias, setClearingAlias] = useState('@external/acquirer_sumup');

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
            Receivables Accounts &amp; Clearing
          </Typography>
          <Typography variant="body2" sx={{ color: '#919EAB' }}>
            Managed Midaz Ledger receivables wallets and external clearing counterbalances
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
          Provision Account
        </Button>
      </Stack>

      {/* Account Cards */}
      <Grid container spacing={3}>
        {accounts.map((acc) => (
          <Grid item xs={12} md={6} key={acc.id}>
            <BentoCard variant="white" sx={{ p: 3.5, height: '100%', display: 'flex', flexDirection: 'column', justifyContent: 'space-between' }}>
              <Box>
                <Stack direction="row" justifyContent="space-between" alignItems="flex-start" sx={{ mb: 2.5 }}>
                  <Box>
                    <Typography
                      sx={{
                        fontFamily: "'Plus Jakarta Sans', sans-serif",
                        fontSize: '18px',
                        fontWeight: 700,
                        color: '#1D1D1D',
                        letterSpacing: '-0.02em'
                      }}
                    >
                      {acc.name}
                    </Typography>
                    <Typography sx={{ color: '#919EAB', fontSize: '12px', mt: 0.25 }}>
                      ID: <code>{acc.id}</code>
                    </Typography>
                  </Box>
                  <StatusBadge variant="lime">
                    {acc.status}
                  </StatusBadge>
                </Stack>

                {/* Double Entry Visualizer within card */}
                <Box sx={{ p: 2, borderRadius: '16px', bgcolor: '#F3F3F3', mb: 3 }}>
                  <Typography sx={{ fontSize: '11px', color: '#919EAB', textTransform: 'uppercase', letterSpacing: '0.05em', mb: 1.5, fontWeight: 600 }}>
                    Double-Entry Routing Pair
                  </Typography>

                  <Stack spacing={1.5}>
                    <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                      <Stack direction="row" spacing={1} alignItems="center">
                        <IconWallet size={16} color="#1D1D1D" />
                        <Typography sx={{ fontSize: '13px', fontWeight: 600, color: '#1D1D1D' }}>
                          Receivables Asset:
                        </Typography>
                      </Stack>
                      <Typography sx={{ fontSize: '13px', fontWeight: 600, color: '#1D1D1D', bgcolor: '#FFFFFF', px: 1.5, py: 0.25, borderRadius: '6px', border: '1px solid #E8E8E8' }}>
                        {acc.account_alias}
                      </Typography>
                    </Box>

                    <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                      <Stack direction="row" spacing={1} alignItems="center">
                        <IconBuildingBank size={16} color="#1D1D1D" />
                        <Typography sx={{ fontSize: '13px', fontWeight: 600, color: '#1D1D1D' }}>
                          Liquid Checking:
                        </Typography>
                      </Stack>
                      <Typography sx={{ fontSize: '13px', fontWeight: 600, color: '#1D1D1D', bgcolor: '#FFFFFF', px: 1.5, py: 0.25, borderRadius: '6px', border: '1px solid #E8E8E8' }}>
                        {acc.checking_account_alias}
                      </Typography>
                    </Box>

                    <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                      <Stack direction="row" spacing={1} alignItems="center">
                        <IconArrowsExchange size={16} color="#1D1D1D" />
                        <Typography sx={{ fontSize: '13px', fontWeight: 600, color: '#1D1D1D' }}>
                          Clearing Account:
                        </Typography>
                      </Stack>
                      <Typography sx={{ fontSize: '13px', fontWeight: 600, color: '#15803d', bgcolor: 'rgba(34, 197, 94, 0.08)', px: 1.5, py: 0.25, borderRadius: '6px', border: '1px solid rgba(34, 197, 94, 0.2)' }}>
                        {acc.clearing_account_alias || '@external/acquirer_sumup'}
                      </Typography>
                    </Box>
                  </Stack>
                </Box>
              </Box>

              <Stack direction="row" justifyContent="space-between" alignItems="center" sx={{ pt: 2, borderTop: '1px solid #E8E8E8' }}>
                <Box>
                  <Typography sx={{ fontSize: '11px', color: '#919EAB' }}>Pending Liquidity</Typography>
                  <Typography sx={{ fontSize: '16px', fontWeight: 700, color: '#1D1D1D' }}>
                    {formatCurrency(acc.metrics?.total_pending_net || 0)}
                  </Typography>
                </Box>
                <Box sx={{ textAlign: 'right' }}>
                  <Typography sx={{ fontSize: '11px', color: '#919EAB' }}>Cleared to Checking</Typography>
                  <Typography sx={{ fontSize: '16px', fontWeight: 700, color: '#15803d' }}>
                    {formatCurrency(acc.metrics?.total_settled || 0)}
                  </Typography>
                </Box>
              </Stack>
            </BentoCard>
          </Grid>
        ))}
      </Grid>

      {/* Provision Account Modal */}
      <Dialog
        open={openModal}
        onClose={() => setOpenModal(false)}
        maxWidth="sm"
        fullWidth
        PaperProps={{
          sx: {
            borderRadius: '24px',
            p: 1.5,
            bgcolor: '#FFFFFF',
            border: '1px solid #E8E8E8',
            boxShadow: '0 24px 48px rgba(0, 0, 0, 0.08)'
          }
        }}
      >
        <DialogTitle sx={{ fontFamily: "'Plus Jakarta Sans', sans-serif", fontWeight: 700, fontSize: '18px' }}>
          Provision Receivables Account
        </DialogTitle>
        <DialogContent>
          <Typography variant="body2" sx={{ color: '#919EAB', mb: 3 }}>
            Registers a receivables wallet in Midaz Ledger linked to the merchant's checking balance and clearing account.
          </Typography>

          <Stack spacing={2.5}>
            <TextField
              label="Account Name"
              placeholder="e.g. Gamma Electronics Receivables"
              value={accountName}
              onChange={(e) => setAccountName(e.target.value)}
              fullWidth
              size="small"
              sx={{ '& .MuiOutlinedInput-root': { borderRadius: '12px' } }}
            />

            <TextField
              label="Receivables Account Alias"
              placeholder="@merchant_alias/receivables"
              value={accountAlias}
              onChange={(e) => setAccountAlias(e.target.value)}
              fullWidth
              size="small"
              sx={{ '& .MuiOutlinedInput-root': { borderRadius: '12px' } }}
            />

            <TextField
              label="Liquid Checking Account Alias"
              placeholder="@merchant_alias/checking"
              value={checkingAlias}
              onChange={(e) => setCheckingAlias(e.target.value)}
              fullWidth
              size="small"
              sx={{ '& .MuiOutlinedInput-root': { borderRadius: '12px' } }}
            />

            <TextField
              label="External Clearing Account"
              placeholder="@external/acquirer_sumup"
              value={clearingAlias}
              onChange={(e) => setClearingAlias(e.target.value)}
              helperText="Universal or custom clearing account in Midaz (counterbalance)"
              fullWidth
              size="small"
              sx={{ '& .MuiOutlinedInput-root': { borderRadius: '12px' } }}
            />
          </Stack>
        </DialogContent>

        <DialogActions sx={{ px: 3, pb: 2.5 }}>
          <Button
            onClick={() => setOpenModal(false)}
            sx={{ borderRadius: '9999px', color: '#666666' }}
          >
            Cancel
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
            Provision in Ledger
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}
