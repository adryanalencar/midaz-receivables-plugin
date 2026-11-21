import PropTypes from 'prop-types';
import { Activity } from 'react';

// @mui
import IconButton from '@mui/material/IconButton';
import Stack from '@mui/material/Stack';
import Box from '@mui/material/Box';

// @project
import { handlerDrawerOpen, useGetMenuMaster } from '@/states/menu';
import Logo from '@/components/logo';

// @assets
import { IconLayoutSidebarLeftCollapse, IconLayoutSidebarRightCollapse } from '@tabler/icons-react';

/***************************  DRAWER HEADER  ***************************/

export default function DrawerHeader({ open }) {
  const { menuMaster } = useGetMenuMaster();
  const drawerOpen = menuMaster.isDashboardDrawerOpened;

  return (
    <Box sx={{ width: 1, px: 2, py: { xs: 2, md: 2.5 } }}>
      <Stack direction="row" sx={{ alignItems: 'center', justifyContent: open ? 'space-between' : 'center', height: 36 }}>
        <Activity mode={open ? 'visible' : 'hidden'}>
          <Logo />
        </Activity>
        <IconButton
          aria-label="open drawer"
          onClick={() => handlerDrawerOpen(!drawerOpen)}
          size="small"
          sx={{
            color: '#919EAB',
            bgcolor: 'rgba(255, 255, 255, 0.05)',
            border: '1px solid rgba(255, 255, 255, 0.1)',
            borderRadius: '9999px',
            '&:hover': {
              color: '#FFFFFF',
              bgcolor: 'rgba(255, 255, 255, 0.1)'
            }
          }}
        >
          {!drawerOpen ? <IconLayoutSidebarRightCollapse size={18} /> : <IconLayoutSidebarLeftCollapse size={18} />}
        </IconButton>
      </Stack>
    </Box>
  );
}

DrawerHeader.propTypes = { open: PropTypes.bool };
