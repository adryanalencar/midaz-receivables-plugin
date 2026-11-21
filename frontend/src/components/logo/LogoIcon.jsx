// @mui
import Box from '@mui/material/Box';

/***************************  LOGO - ICON  ***************************/

export default function LogoIcon() {
  return (
    <Box
      sx={{
        width: 32,
        height: 32,
        borderRadius: '9px',
        bgcolor: '#F0FD63',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        cursor: 'pointer'
      }}
    >
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none">
        <rect x="3" y="3" width="7" height="7" rx="2" fill="#1D1D1D" />
        <rect x="14" y="3" width="7" height="7" rx="2" fill="#1D1D1D" fillOpacity="0.4" />
        <rect x="14" y="14" width="7" height="7" rx="2" fill="#1D1D1D" />
        <rect x="3" y="14" width="7" height="7" rx="2" fill="#1D1D1D" fillOpacity="0.4" />
      </svg>
    </Box>
  );
}
