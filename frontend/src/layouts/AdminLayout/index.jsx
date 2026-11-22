import { useEffect } from 'react';
import { Outlet } from 'react-router-dom';

import useMediaQuery from '@mui/material/useMediaQuery';
import Container from '@mui/material/Container';
import Toolbar from '@mui/material/Toolbar';
import Stack from '@mui/material/Stack';
import Box from '@mui/material/Box';

// @project
import Drawer from './Drawer';
import Header from './Header';
import { handlerDrawerOpen, useGetMenuMaster } from '@/states/menu';
import Breadcrumbs from '@/components/Breadcrumbs';
import Loader from '@/components/Loader';
import { DRAWER_WIDTH } from '@/config';
import useConfig from '@/hooks/useConfig';

/***************************  ADMIN LAYOUT  ***************************/

export default function DashboardLayout() {
  const { menuMaster, menuMasterLoading } = useGetMenuMaster();
  const drawerOpen = menuMaster?.isDashboardDrawerOpened ?? true;
  const downMD = useMediaQuery((theme) => theme.breakpoints.down('md'));

  useEffect(() => {
    handlerDrawerOpen(!downMD);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [downMD]);

  if (menuMasterLoading) return <Loader />;

  return (
    <Stack direction="row" sx={{ width: 1, minHeight: '100vh', bgcolor: '#F3F3F3' }}>
      <Header />
      <Drawer />
      <Box
        component="main"
        sx={{
          width: {
            xs: '100%',
            md: drawerOpen ? `calc(100% - ${DRAWER_WIDTH}px)` : '100%'
          },
          flexGrow: 1,
          p: { xs: 2, sm: 3 },
          bgcolor: '#F3F3F3',
          minHeight: '100vh'
        }}
      >
        <Toolbar sx={{ minHeight: { xs: 54, sm: 46, md: 76 } }} />
        <Box
          sx={{
            py: 0.4,
            px: 1.5,
            mx: { xs: -2, sm: -3 },
            display: { xs: 'block', md: 'none' },
            borderBottom: '1px solid #E8E8E8',
            mb: 2
          }}
        >
          <Breadcrumbs />
        </Box>
        <Container maxWidth="xl" sx={{ px: { xs: 0, sm: 2 } }}>
          <Outlet />
        </Container>
      </Box>
    </Stack>
  );
}
