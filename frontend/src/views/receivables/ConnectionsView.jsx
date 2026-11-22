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
  IconClock,
  IconKey,
  IconCheck,
  IconBuildingStore
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
            Operadoras e Credenciadoras Conectadas
          </Typography>
          <Typography variant="body2" sx={{ color: '#919EAB' }}>
            Contratos de adquirência e credenciamentos integrados para captura automática de vendas
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
          Conectar Nova Credenciadora
        </Button>
      </Stack>

      {/* Connection Cards Grid */}
      <Grid container spacing={3}>
        {connections.map((conn) => {
          const isSyncing = syncingId === conn.id;
          const isRecentlySynced = syncedFeedback === conn.id;

          return (
            <Grid size={{ xs: 12, sm: 6, lg: 4 }} key={conn.id}>
              <BentoCard
                variant="white"
                sx={{
                  height: '100%',
                  display: 'flex',
                  flexDirection: 'column',
                  justifyContent: 'space-between',
                  p: 3.5
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
                      {conn.status === 'ACTIVE' ? 'Contrato Ativo' : 'Pausado'}
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
                    Credenciadora: {conn.acquirer_type}
                  </Typography>

                  <Stack spacing={1.5} sx={{ p: 2, borderRadius: '14px', bgcolor: '#F3F3F3', mb: 2.5 }}>
                    <Stack direction="row" spacing={1} alignItems="center">
                      <IconKey size={15} color="#919EAB" />
                      <Typography sx={{ fontSize: '12px', color: '#666666' }}>
                        Merchant ID (Afiliação): <strong>{conn.merchant_code || '12345678000195'}</strong>
                      </Typography>
                    </Stack>
                    <Stack direction="row" spacing={1} alignItems="center">
                      <IconBuildingStore size={15} color="#919EAB" />
                      <Typography sx={{ fontSize: '12px', color: '#666666' }}>
                        Terminais e POS vinculados: <strong>Múltiplos</strong>
                      </Typography>
                    </Stack>
                    <Stack direction="row" spacing={1} alignItems="center">
                      <IconClock size={15} color="#919EAB" />
                      <Typography sx={{ fontSize: '12px', color: '#666666' }}>
                        Sincronização: a cada 60 min (Automática)
                      </Typography>
                    </Stack>
                  </Stack>
                </Box>

                <Button
                  variant="outlined"
                  onClick={() => handleSync(conn.id)}
                  disabled={isSyncing}
                  fullWidth
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
                    borderColor: isRecentlySynced ? 'rgba(34, 197, 94, 0.4)' : '#E8E8E8',
                    color: isRecentlySynced ? '#15803d' : '#1D1D1D',
                    bgcolor: isRecentlySynced ? 'rgba(34, 197, 94, 0.08)' : '#FFFFFF',
                    fontSize: '12px',
                    py: 0.75,
                    '&:hover': { bgcolor: '#F3F3F3' }
                  }}
                >
                  {isSyncing ? 'Sincronizando transações...' : isRecentlySynced ? 'Vendas Atualizadas!' : 'Sincronizar Vendas'}
                </Button>
              </BentoCard>
            </Grid>
          );
        })}
      </Grid>

      {/* Modal - Conectar Credenciadora */}
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
          Conectar Credenciadora / Adquirente
        </DialogTitle>
        <DialogContent>
          <Stack spacing={2.5} sx={{ mt: 1 }}>
            <TextField
              select
              label="Operadora / Adquirente"
              value={acquirerType}
              onChange={(e) => setAcquirerType(e.target.value)}
              fullWidth
              size="small"
            >
              <MenuItem value="SUMUP">SumUp Pagamentos</MenuItem>
              <MenuItem value="ADITUM">Aditum Gateway / TEF</MenuItem>
              <MenuItem value="PAGBANK">PagBank (PagSeguro)</MenuItem>
              <MenuItem value="STONE">Stone Pagamentos</MenuItem>
            </TextField>

            <TextField
              label="Identificação do Credenciamento"
              placeholder="Ex: Afiliação Matriz Centro"
              value={displayName}
              onChange={(e) => setDisplayName(e.target.value)}
              fullWidth
              size="small"
            />

            <TextField
              label="Código do Estabelecimento (Merchant ID / StoneCode)"
              placeholder="Ex: 01948df2-aditum-merchant-uuid"
              value={merchantCode}
              onChange={(e) => setMerchantCode(e.target.value)}
              fullWidth
              size="small"
            />

            <TextField
              label="Token de Acesso / API Key da Credenciadora"
              placeholder="Chave secreta ou credencial de API"
              type="password"
              value={apiKey}
              onChange={(e) => setApiKey(e.target.value)}
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
            disabled={loading || !displayName || !merchantCode}
            sx={{
              borderRadius: '9999px',
              bgcolor: '#1D1D1D',
              color: '#FFFFFF',
              px: 3,
              '&:hover': { bgcolor: '#000000' }
            }}
          >
            Salvar Credenciamento
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}
