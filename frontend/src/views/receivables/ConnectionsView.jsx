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
import MenuItem from '@mui/material/MenuItem';
import CircularProgress from '@mui/material/CircularProgress';

// @project
import { BentoCard, StatusBadge } from '@/components/alignly';
import { receivablesApi } from '@/api/receivablesApi';

// @assets
import {
  IconPlus,
  IconRefresh,
  IconPlugConnected,
  IconWebhook,
  IconClock,
  IconKey,
  IconCheck
} from '@tabler/icons-react';

/***************************  ACQUIRER CONNECTIONS VIEW  ***************************/

export default function ConnectionsView() {
  const [connections, setConnections] = useState([]);
  const [loading, setLoading] = useState(false);
  const [syncingId, setSyncingId] = useState(null);
  const [openModal, setOpenModal] = useState(false);
  const [syncedFeedback, setSyncedFeedback] = useState(null);

  // Form state
  const [acquirerType, setAcquirerType] = useState('SUMUP');
  const [displayName, setDisplayName] = useState('');
  const [merchantCode, setMerchantCode] = useState('');
  const [apiKey, setApiKey] = useState('');

  const loadConnections = async () => {
    setLoading(true);
    try {
      const data = await receivablesApi.getConnections();
      setConnections(data || []);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadConnections();
  }, []);

  const handleSync = async (connId) => {
    setSyncingId(connId);
    try {
      await receivablesApi.triggerSync(connId);
      setSyncedFeedback(connId);
      setTimeout(() => setSyncedFeedback(null), 3000);
      loadConnections();
    } finally {
      setSyncingId(null);
    }
  };

  const handleCreate = async () => {
    if (!displayName || !merchantCode) return;
    setLoading(true);
    try {
      await receivablesApi.createConnection('01948df2-aaaa-7000-8000-000000000001', {
        acquirer_type: acquirerType,
        display_name: displayName,
        credentials: {
          merchant_code: merchantCode,
          api_key: apiKey
        },
        auto_sync_enabled: true,
        sync_interval_minutes: 60
      });
      setOpenModal(false);
      setDisplayName('');
      setMerchantCode('');
      setApiKey('');
      loadConnections();
    } finally {
      setLoading(false);
    }
  };

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
            Acquirer Connections
          </Typography>
          <Typography variant="body2" sx={{ color: '#919EAB' }}>
            Multi-tenant credential isolation (1 Receivables Account : N Acquirer Connections)
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
          Connect Acquirer
        </Button>
      </Stack>

      {/* Connection Cards Grid */}
      <Grid container spacing={3}>
        {connections.map((conn) => {
          const isSyncing = syncingId === conn.id;
          const isRecentlySynced = syncedFeedback === conn.id;

          return (
            <Grid item xs={12} md={6} lg={4} key={conn.id}>
              <BentoCard
                variant="white"
                sx={{
                  height: '100%',
                  display: 'flex',
                  flexDirection: 'column',
                  justifyContent: 'space-between',
                  p: 3
                }}
              >
                <Box>
                  <Stack direction="row" justifyContent="space-between" alignItems="flex-start" sx={{ mb: 2 }}>
                    <Box
                      sx={{
                        width: 44,
                        height: 44,
                        borderRadius: '12px',
                        bgcolor: conn.acquirer_type === 'SUMUP' ? '#F0FD63' : conn.acquirer_type === 'ADITUM' ? '#1D1D1D' : '#919EAB',
                        color: conn.acquirer_type === 'SUMUP' ? '#1D1D1D' : '#FFFFFF',
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        fontWeight: 700,
                        fontSize: '13px'
                      }}
                    >
                      {conn.acquirer_type.slice(0, 3)}
                    </Box>
                    <StatusBadge variant={conn.status === 'ACTIVE' ? 'lime' : 'neutral'}>
                      {conn.status}
                    </StatusBadge>
                  </Stack>

                  <Typography
                    sx={{
                      fontFamily: "'Plus Jakarta Sans', sans-serif",
                      fontSize: '16px',
                      fontWeight: 700,
                      color: '#1D1D1D',
                      letterSpacing: '-0.02em',
                      mb: 0.5
                    }}
                  >
                    {conn.display_name}
                  </Typography>
                  <Typography sx={{ color: '#919EAB', fontSize: '12px', mb: 2 }}>
                    {conn.acquirer_type} Gateway Connection
                  </Typography>

                  <Stack spacing={1.5} sx={{ p: 2, borderRadius: '14px', bgcolor: '#F3F3F3', mb: 2.5 }}>
                    <Stack direction="row" spacing={1} alignItems="center">
                      <IconKey size={15} color="#919EAB" />
                      <Typography sx={{ fontSize: '11px', color: '#666666' }}>
                        Merchant ID: <strong>{conn.merchant_code}</strong>
                      </Typography>
                    </Stack>
                    <Stack direction="row" spacing={1} alignItems="center">
                      <IconClock size={15} color="#919EAB" />
                      <Typography sx={{ fontSize: '11px', color: '#666666' }}>
                        Sync Every: <strong>{conn.sync_interval_minutes} mins</strong>
                      </Typography>
                    </Stack>
                    <Stack direction="row" spacing={1} alignItems="center">
                      <IconWebhook size={15} color="#919EAB" />
                      <Typography
                        sx={{
                          fontSize: '11px',
                          color: '#666666',
                          whiteSpace: 'nowrap',
                          overflow: 'hidden',
                          textOverflow: 'ellipsis'
                        }}
                      >
                        Webhook: {conn.webhook_url}
                      </Typography>
                    </Stack>
                  </Stack>
                </Box>

                <Stack direction="row" spacing={1} justifyContent="space-between" alignItems="center" sx={{ pt: 1, borderTop: '1px solid #E8E8E8' }}>
                  <Typography sx={{ fontSize: '11px', color: '#919EAB' }}>
                    Last sync: {new Date(conn.last_sync_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                  </Typography>

                  <Button
                    size="small"
                    variant="outlined"
                    onClick={() => handleSync(conn.id)}
                    disabled={isSyncing}
                    startIcon={
                      isSyncing ? (
                        <CircularProgress size={14} color="inherit" />
                      ) : isRecentlySynced ? (
                        <IconCheck size={14} color="#15803d" />
                      ) : (
                        <IconRefresh size={14} />
                      )
                    }
                    sx={{
                      borderRadius: '9999px',
                      color: isRecentlySynced ? '#15803d' : '#1D1D1D',
                      borderColor: isRecentlySynced ? 'rgba(34, 197, 94, 0.4)' : '#E8E8E8',
                      bgcolor: isRecentlySynced ? 'rgba(34, 197, 94, 0.08)' : '#FFFFFF',
                      fontSize: '11px',
                      py: 0.5,
                      px: 2,
                      '&:hover': { bgcolor: '#F3F3F3' }
                    }}
                  >
                    {isSyncing ? 'Syncing...' : isRecentlySynced ? 'Synced!' : 'Sync Now'}
                  </Button>
                </Stack>
              </BentoCard>
            </Grid>
          );
        })}
      </Grid>

      {/* Connect Acquirer Modal */}
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
          Connect New Acquirer
        </DialogTitle>
        <DialogContent>
          <Typography variant="body2" sx={{ color: '#919EAB', mb: 3 }}>
            Provision isolated credentials for a payment acquirer bound to this Receivables Account.
          </Typography>

          <Stack spacing={2.5}>
            <TextField
              select
              label="Acquirer Provider"
              value={acquirerType}
              onChange={(e) => setAcquirerType(e.target.value)}
              fullWidth
              size="small"
              sx={{ '& .MuiOutlinedInput-root': { borderRadius: '12px' } }}
            >
              <MenuItem value="SUMUP">SumUp (REST APIs &amp; Webhooks)</MenuItem>
              <MenuItem value="ADITUM">Aditum (Multi-Acquirer Gateway &amp; EDI 3.2)</MenuItem>
              <MenuItem value="PAGBANK">PagBank (REST APIs &amp; Webhooks)</MenuItem>
            </TextField>

            <TextField
              label="Display Name / Terminal Tag"
              placeholder="e.g. Aditum Gateway SP Terminal"
              value={displayName}
              onChange={(e) => setDisplayName(e.target.value)}
              fullWidth
              size="small"
              sx={{ '& .MuiOutlinedInput-root': { borderRadius: '12px' } }}
            />

            <TextField
              label="Merchant Code / ID"
              placeholder="e.g. ADT-98214 or SU-BRA-99120"
              value={merchantCode}
              onChange={(e) => setMerchantCode(e.target.value)}
              fullWidth
              size="small"
              sx={{ '& .MuiOutlinedInput-root': { borderRadius: '12px' } }}
            />

            <TextField
              label="API Key / Secret Token"
              type="password"
              placeholder="Enter secure API secret"
              value={apiKey}
              onChange={(e) => setApiKey(e.target.value)}
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
            disabled={!displayName || !merchantCode || loading}
            sx={{
              borderRadius: '9999px',
              bgcolor: '#1D1D1D',
              color: '#FFFFFF',
              px: 3,
              '&:hover': { bgcolor: '#000000' }
            }}
          >
            Save Connection
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}
