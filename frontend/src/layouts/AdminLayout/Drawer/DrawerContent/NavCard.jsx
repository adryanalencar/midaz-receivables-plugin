import PropTypes from 'prop-types';

// @mui
import Avatar from '@mui/material/Avatar';
import Box from '@mui/material/Box';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';

// @project
import useCurrentUser from '@/hooks/useCurrentUser';

// @assets
import { IconChevronDown } from '@tabler/icons-react';

/***************************  DRAWER CONTENT - USER PROFILE FOOTER  ***************************/

export default function NavCard({ isMiniDrawer }) {
  const { userData } = useCurrentUser();

  const name = userData?.name || 'Midaz User';
  const email = userData?.email || 'user@midaz.io';
  const initials = name
    .split(' ')
    .map((n) => n[0])
    .slice(0, 2)
    .join('')
    .toUpperCase();

  if (isMiniDrawer) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', mb: 2.5, width: 1 }}>
        <Avatar
          sx={{
            width: 36,
            height: 36,
            bgcolor: '#F0FD63',
            color: '#1D1D1D',
            fontSize: '12px',
            fontWeight: 700,
            cursor: 'pointer',
            transition: 'transform 0.15s ease',
            '&:hover': { transform: 'scale(1.05)' }
          }}
        >
          {initials}
        </Avatar>
      </Box>
    );
  }

  return (
    <Box
      sx={{
        mx: 0.5,
        mb: 2.5,
        p: 1.25,
        borderRadius: '16px',
        bgcolor: 'rgba(255, 255, 255, 0.05)',
        border: '1px solid rgba(255, 255, 255, 0.08)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        cursor: 'pointer',
        transition: 'all 0.15s ease',
        '&:hover': {
          bgcolor: 'rgba(255, 255, 255, 0.08)',
          borderColor: 'rgba(255, 255, 255, 0.12)'
        }
      }}
    >
      <Stack direction="row" spacing={1.5} alignItems="center" sx={{ overflow: 'hidden' }}>
        <Avatar
          sx={{
            width: 36,
            height: 36,
            bgcolor: '#F0FD63',
            color: '#1D1D1D',
            fontSize: '12px',
            fontWeight: 700,
            flexShrink: 0
          }}
        >
          {initials}
        </Avatar>
        <Box sx={{ overflow: 'hidden', minWidth: 0 }}>
          <Typography
            sx={{
              color: '#FFFFFF',
              fontSize: '12px',
              fontWeight: 600,
              lineHeight: 1.2,
              whiteSpace: 'nowrap',
              overflow: 'hidden',
              textOverflow: 'ellipsis'
            }}
          >
            {name}
          </Typography>
          <Typography
            sx={{
              color: '#919EAB',
              fontSize: '11px',
              lineHeight: 1.2,
              mt: 0.25,
              whiteSpace: 'nowrap',
              overflow: 'hidden',
              textOverflow: 'ellipsis'
            }}
          >
            {email}
          </Typography>
        </Box>
      </Stack>
      <IconChevronDown size={16} color="#919EAB" style={{ flexShrink: 0, marginLeft: 8 }} />
    </Box>
  );
}

NavCard.propTypes = { isMiniDrawer: PropTypes.bool };
