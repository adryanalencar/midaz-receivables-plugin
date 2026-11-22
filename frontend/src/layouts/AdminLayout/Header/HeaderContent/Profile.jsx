import { useState } from 'react';

// @mui
import { useTheme } from '@mui/material/styles';
import Avatar from '@mui/material/Avatar';
import Button from '@mui/material/Button';
import ClickAwayListener from '@mui/material/ClickAwayListener';
import Divider from '@mui/material/Divider';
import Fade from '@mui/material/Fade';
import List from '@mui/material/List';
import ListItem from '@mui/material/ListItem';
import ListItemButton from '@mui/material/ListItemButton';
import ListItemIcon from '@mui/material/ListItemIcon';
import ListItemText from '@mui/material/ListItemText';
import Popper from '@mui/material/Popper';
import Stack from '@mui/material/Stack';
import Box from '@mui/material/Box';

// @project
import { ThemeDirection } from '@/config';
import MainCard from '@/components/MainCard';
import Profile from '@/components/Profile';
import { AuthRole, AvatarSize } from '@/enum';
import useConfig from '@/hooks/useConfig';
import useCurrentUser from '@/hooks/useCurrentUser';
import { logout } from '@/utils/api/auth';

// @assets
import { IconLogout, IconSettings, IconUser } from '@tabler/icons-react';

/***************************  HEADER - PROFILE  ***************************/

const RoleTitles = {
  [AuthRole.SUPER_ADMIN]: 'Administrador Financeiro',
  [AuthRole.ADMIN]: 'Operador Financeiro',
  [AuthRole.USER]: 'Auditor Contábil'
};

export default function ProfileSection() {
  const theme = useTheme();
  const { userData } = useCurrentUser();

  const [anchorEl, setAnchorEl] = useState(null);

  const open = Boolean(anchorEl);
  const id = open ? 'profile-action-popper' : undefined;
  const buttonStyle = { borderRadius: 2, p: 1 };

  const name = userData && Object.keys(userData).length > 0 && userData.firstname
    ? `${userData.firstname} ${userData.lastname}`.trim()
    : 'Adryan Alencar';
  const roleTitle = userData?.role ? RoleTitles[userData.role] || 'Administrador Financeiro' : 'Administrador Financeiro';
  const initials = name
    .split(' ')
    .map((n) => n[0])
    .slice(0, 2)
    .join('')
    .toUpperCase() || 'AA';

  const currentProfileData = {
    avatar: {
      children: initials,
      size: AvatarSize.XS,
      sx: { width: 34, height: 34, bgcolor: '#F0FD63', color: '#1D1D1D', fontWeight: 700 }
    },
    title: name,
    caption: roleTitle
  };

  const handleActionClick = (event) => {
    setAnchorEl(anchorEl ? null : event.currentTarget);
  };

  const logoutAccount = () => {
    setAnchorEl(null);
    logout();
  };

  return (
    <>
      <Box onClick={handleActionClick} sx={{ cursor: 'pointer' }}>
        <Box sx={{ display: { xs: 'none', sm: 'flex' } }}>
          <Profile {...currentProfileData} />
        </Box>
        <Box sx={{ display: { xs: 'block', sm: 'none' } }}>
          <Avatar
            sx={{
              width: 32,
              height: 32,
              bgcolor: '#F0FD63',
              color: '#1D1D1D',
              fontSize: '12px',
              fontWeight: 700
            }}
          >
            {initials}
          </Avatar>
        </Box>
      </Box>
      <Popper
        placement="bottom-end"
        id={id}
        open={open}
        anchorEl={anchorEl}
        transition
        popperOptions={{ modifiers: [{ name: 'offset', options: { offset: [theme.direction === ThemeDirection.RTL ? -8 : 8, 8] } }] }}
      >
        {({ TransitionProps }) => (
          <Fade in={open} {...TransitionProps}>
            <MainCard sx={{ borderRadius: 2, boxShadow: theme.vars.customShadows.tooltip, minWidth: 220, p: 0.5 }}>
              <ClickAwayListener onClickAway={() => setAnchorEl(null)}>
                <Stack sx={{ px: 0.5, py: 0.75 }}>
                  <Profile
                    {...currentProfileData}
                    sx={{
                      flexDirection: 'column',
                      justifyContent: 'center',
                      textAlign: 'center',
                      width: 1,
                      '& .MuiAvatar-root': { width: 44, height: 44, fontSize: '15px', mx: 'auto', mb: 0.5 }
                    }}
                  />
                  <Divider sx={{ my: 1 }} />
                  <List disablePadding>
                    <ListItemButton href="#" sx={buttonStyle}>
                      <ListItemIcon>
                        <IconUser size={16} />
                      </ListItemIcon>
                      <ListItemText primary="Minha Conta" />
                    </ListItemButton>
                    <ListItemButton href="#" sx={{ ...buttonStyle, my: 0.5 }}>
                      <ListItemIcon>
                        <IconSettings size={16} />
                      </ListItemIcon>
                      <ListItemText primary="Configurações" />
                    </ListItemButton>
                    <ListItem disablePadding sx={{ mt: 1 }}>
                      <Button
                        fullWidth
                        variant="outlined"
                        color="secondary"
                        size="small"
                        endIcon={<IconLogout size={16} />}
                        onClick={logoutAccount}
                        sx={{ borderRadius: '8px' }}
                      >
                        Sair
                      </Button>
                    </ListItem>
                  </List>
                </Stack>
              </ClickAwayListener>
            </MainCard>
          </Fade>
        )}
      </Popper>
    </>
  );
}
