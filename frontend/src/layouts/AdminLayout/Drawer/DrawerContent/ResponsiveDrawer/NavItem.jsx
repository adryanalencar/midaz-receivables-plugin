import PropTypes from 'prop-types';
import { Activity, useEffect } from 'react';

// @mui
import { useTheme } from '@mui/material/styles';
import useMediaQuery from '@mui/material/useMediaQuery';
import ListItemButton from '@mui/material/ListItemButton';
import ListItemIcon from '@mui/material/ListItemIcon';
import ListItemText from '@mui/material/ListItemText';

// @project
import { handlerActiveItem, handlerDrawerOpen, useGetMenuMaster } from '@/states/menu';
import DynamicIcon from '@/components/DynamicIcon';
import RouterLink from '@/components/Link';
import { usePathname } from '@/utils/navigation';

// @third-party
import { FormattedMessage } from 'react-intl';

/***************************  RESPONSIVE DRAWER - ITEM  ***************************/

export default function NavItem({ item, level = 0 }) {
  const theme = useTheme();
  const { menuMaster } = useGetMenuMaster();
  const openItem = menuMaster.openedItem;

  const downMD = useMediaQuery(theme.breakpoints.down('md'));

  // Active menu item on page load
  const pathname = usePathname();

  useEffect(() => {
    if (pathname === item.url) handlerActiveItem(item.id);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [pathname]);

  const isSelected = openItem === item.id;

  const itemHandler = () => {
    if (downMD) handlerDrawerOpen(false);
  };

  return (
    <ListItemButton
      id={`${item.id}-btn`}
      component={RouterLink}
      to={item.url}
      {...(item?.target && { target: '_blank' })}
      selected={isSelected}
      disabled={item.disabled}
      onClick={itemHandler}
      sx={{
        my: 0.35,
        px: 1.5,
        py: 0.85,
        borderRadius: '12px',
        color: isSelected ? '#FFFFFF' : '#919EAB',
        bgcolor: isSelected ? '#2C2C2C' : 'transparent',
        transition: 'all 0.15s ease',
        '&:hover': {
          bgcolor: isSelected ? '#2C2C2C' : 'rgba(255, 255, 255, 0.06)',
          color: '#FFFFFF',
          '& .MuiListItemIcon-root': {
            color: isSelected ? '#F0FD63' : '#FFFFFF'
          }
        },
        '&.Mui-selected': {
          bgcolor: '#2C2C2C',
          color: '#FFFFFF',
          '&:hover': {
            bgcolor: '#333333'
          },
          '& .MuiListItemIcon-root': {
            color: '#F0FD63'
          }
        }
      }}
    >
      <Activity mode={level === 0 ? 'visible' : 'hidden'}>
        <ListItemIcon sx={{ minWidth: 28, color: isSelected ? '#F0FD63' : '#919EAB', transition: 'color 0.15s ease' }}>
          <DynamicIcon name={item.icon} color="currentColor" size={18} stroke={1.5} />
        </ListItemIcon>
      </Activity>
      <ListItemText
        primary={<FormattedMessage id={item.title} defaultMessage={item.title} />}
        slotProps={{
          primary: {
            sx: {
              fontSize: '13px',
              fontWeight: isSelected ? 600 : 400,
              letterSpacing: '-0.01em'
            }
          }
        }}
        sx={{ mb: '-1px', my: 0 }}
      />
    </ListItemButton>
  );
}

NavItem.propTypes = { item: PropTypes.any, level: PropTypes.number };
