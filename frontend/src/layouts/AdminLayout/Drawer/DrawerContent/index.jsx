import useMediaQuery from '@mui/material/useMediaQuery';
import Stack from '@mui/material/Stack';

// @project
import MiniDrawer from './MiniDrawer';
import ResponsiveDrawer from './ResponsiveDrawer';

import { useGetMenuMaster } from '@/states/menu';
import SimpleBar from '@/components/third-party/SimpleBar';

/***************************  DRAWER - CONTENT (NO USER FOOTER)  ***************************/

export default function DrawerContent() {
  const upLG = useMediaQuery((theme) => theme.breakpoints.up('lg'));
  const { menuMaster } = useGetMenuMaster();
  const drawerOpen = menuMaster.isDashboardDrawerOpened;

  return (
    <SimpleBar sx={{ height: 'calc(100vh - 70px)' }}>
      <Stack sx={{ px: !drawerOpen && upLG ? 0 : 2, pt: 1 }}>
        {!drawerOpen && upLG ? <MiniDrawer /> : <ResponsiveDrawer />}
      </Stack>
    </SimpleBar>
  );
}
