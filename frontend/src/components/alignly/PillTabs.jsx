import PropTypes from 'prop-types';
import Box from '@mui/material/Box';
import Stack from '@mui/material/Stack';
import ButtonBase from '@mui/material/ButtonBase';

/**
 * Alignly Pill Tabs Navigation
 * (Active tab: bg-[#1D1D1D] text-white rounded-full px-5 py-2 text-xs font-medium)
 */
export default function PillTabs({ tabs, activeTab, onChange, sx = {} }) {
  return (
    <Box
      sx={{
        display: 'inline-flex',
        p: 0.5,
        borderRadius: '9999px',
        bgcolor: '#FFFFFF',
        border: '1px solid #E8E8E8',
        ...sx
      }}
    >
      <Stack direction="row" spacing={0.5}>
        {tabs.map((tab) => {
          const isActive = tab.value === activeTab;
          return (
            <ButtonBase
              key={tab.value}
              onClick={() => onChange(tab.value)}
              sx={{
                px: 2.5,
                py: 1,
                borderRadius: '9999px',
                fontSize: '12px',
                fontWeight: isActive ? 600 : 500,
                color: isActive ? '#FFFFFF' : 'rgba(29, 29, 29, 0.7)',
                bgcolor: isActive ? '#1D1D1D' : 'transparent',
                transition: 'all 0.15s ease',
                '&:hover': {
                  color: isActive ? '#FFFFFF' : '#1D1D1D',
                  bgcolor: isActive ? '#1D1D1D' : '#F3F3F3'
                }
              }}
            >
              {tab.label}
            </ButtonBase>
          );
        })}
      </Stack>
    </Box>
  );
}

PillTabs.propTypes = {
  tabs: PropTypes.arrayOf(
    PropTypes.shape({
      label: PropTypes.string.isRequired,
      value: PropTypes.oneOfType([PropTypes.string, PropTypes.number]).isRequired
    })
  ).isRequired,
  activeTab: PropTypes.oneOfType([PropTypes.string, PropTypes.number]).isRequired,
  onChange: PropTypes.func.isRequired,
  sx: PropTypes.object
};
