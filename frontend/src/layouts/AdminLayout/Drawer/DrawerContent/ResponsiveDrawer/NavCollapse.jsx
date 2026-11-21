import PropTypes from 'prop-types';
import { Activity, useState } from 'react';

// @mui
import Collapse from '@mui/material/Collapse';
import List from '@mui/material/List';
import ListItemButton from '@mui/material/ListItemButton';
import ListItemIcon from '@mui/material/ListItemIcon';
import ListItemText from '@mui/material/ListItemText';
import Typography from '@mui/material/Typography';

// @project
import NavItem from './NavItem';
import DynamicIcon from '@/components/DynamicIcon';
import useCurrentUser from '@/hooks/useCurrentUser';
import useMenuCollapse from '@/hooks/useMenuCollapse';
import { usePathname } from '@/utils/navigation';

// @third-party
import { FormattedMessage } from 'react-intl';

// @assets
import { IconChevronDown, IconChevronUp } from '@tabler/icons-react';

// @style
const verticalDivider = {
  '&:after': {
    content: "''",
    position: 'absolute',
    left: 16,
    top: -2,
    height: `calc(100% + 2px)`,
    width: '1px',
    opacity: 1,
    bgcolor: 'rgba(255, 255, 255, 0.1)'
  }
};

/***************************  COLLAPSE - LOOP  ***************************/

function NavCollapseLoop({ item, userRole }) {
  return item.children?.map((item) => {
    if (item.roles?.length && userRole && !item.roles.includes(userRole)) {
      return null;
    }

    switch (item.type) {
      case 'collapse':
        return <NavCollapse key={item.id} item={item} level={1} />;
      case 'item':
        return <NavItem key={item.id} item={item} level={1} />;
      default:
        return (
          <Typography key={item.id} variant="h6" color="error" align="center">
            Fix - Collapse or Item
          </Typography>
        );
    }
  });
}

/***************************  RESPONSIVE DRAWER - COLLAPSE  ***************************/

export default function NavCollapse({ item, level = 0 }) {
  const { userData } = useCurrentUser();

  const [open, setOpen] = useState(false);
  const [selected, setSelected] = useState(null);

  const userRole = userData?.role;
  const pathname = usePathname();

  useMenuCollapse(item, pathname, false, setSelected, setOpen);

  const handleClick = () => {
    setOpen(!open);
  };

  const isCurrent = open || selected === item.id;

  return (
    <>
      <ListItemButton
        id={`${item.id}-btn`}
        selected={isCurrent}
        sx={{
          my: 0.35,
          px: 1.5,
          py: 0.85,
          borderRadius: '12px',
          color: isCurrent ? '#FFFFFF' : '#919EAB',
          bgcolor: isCurrent ? '#2C2C2C' : 'transparent',
          transition: 'all 0.15s ease',
          '&:hover': {
            bgcolor: isCurrent ? '#2C2C2C' : 'rgba(255, 255, 255, 0.06)',
            color: '#FFFFFF',
            '& .MuiListItemIcon-root': {
              color: isCurrent ? '#F0FD63' : '#FFFFFF'
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
        onClick={handleClick}
      >
        <Activity mode={level === 0 ? 'visible' : 'hidden'}>
          <ListItemIcon sx={{ minWidth: 28, color: isCurrent ? '#F0FD63' : '#919EAB', transition: 'color 0.15s ease' }}>
            <DynamicIcon name={item.icon} color="currentColor" size={18} stroke={1.5} />
          </ListItemIcon>
        </Activity>
        <ListItemText
          primary={<FormattedMessage id={item.title} />}
          slotProps={{
            primary: {
              sx: {
                fontSize: '13px',
                fontWeight: isCurrent ? 600 : 400,
                letterSpacing: '-0.01em'
              }
            }
          }}
          sx={{ mb: '-1px', my: 0 }}
        />
        {open ? <IconChevronUp size={16} stroke={1.5} /> : <IconChevronDown size={16} stroke={1.5} />}
      </ListItemButton>
      <Collapse in={open} timeout="auto" unmountOnExit>
        <List component="div" sx={{ p: 0, pl: 2.5, position: 'relative', ...verticalDivider }}>
          <NavCollapseLoop item={item} userRole={userRole} />
        </List>
      </Collapse>
    </>
  );
}

NavCollapseLoop.propTypes = { item: PropTypes.any, userRole: PropTypes.any };
NavCollapse.propTypes = { item: PropTypes.any, level: PropTypes.number };
