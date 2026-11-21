// @mui
import { styled } from '@mui/material/styles';
import Drawer from '@mui/material/Drawer';

// @project
import { DRAWER_WIDTH, MINI_DRAWER_WIDTH } from '@/config';

// Mixin for common (open/closed) drawer state
const commonDrawerStyles = () => ({
  backgroundColor: '#1D1D1D',
  color: '#FFFFFF',
  borderRight: 'none',
  boxShadow: 'none',
  overflowX: 'hidden'
});

// Mixin for opened drawer state
const openedMixin = (theme) => ({
  ...commonDrawerStyles(theme),
  width: DRAWER_WIDTH,
  backgroundColor: '#1D1D1D',
  color: '#FFFFFF',
  transition: theme.transitions.create('width', {
    easing: theme.transitions.easing.sharp,
    duration: theme.transitions.duration.enteringScreen
  })
});

// Mixin for closed drawer state
const closedMixin = (theme) => ({
  ...commonDrawerStyles(theme),
  width: MINI_DRAWER_WIDTH,
  backgroundColor: '#1D1D1D',
  color: '#FFFFFF',
  transition: theme.transitions.create('width', {
    easing: theme.transitions.easing.sharp,
    duration: theme.transitions.duration.leavingScreen
  })
});

/***************************  DRAWER - MINI STYLED  ***************************/

const MiniDrawerStyled = styled(Drawer, { shouldForwardProp: (prop) => prop !== 'open' })(({ theme, open }) => ({
  width: DRAWER_WIDTH,
  flexShrink: 0,
  whiteSpace: 'nowrap',
  boxSizing: 'border-box',
  '& .MuiDrawer-paper': {
    backgroundColor: '#1D1D1D',
    color: '#FFFFFF',
    borderRight: 'none',
    boxShadow: 'none',
    ...(open ? openedMixin(theme) : closedMixin(theme))
  },
  ...(open && {
    ...openedMixin(theme)
  }),
  ...(!open && {
    ...closedMixin(theme)
  })
}));

export default MiniDrawerStyled;
