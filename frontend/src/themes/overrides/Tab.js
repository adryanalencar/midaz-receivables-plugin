// @project
import { withAlpha } from '@/utils/colorUtils';

/***************************  ALIGNLY OVERRIDES - TAB (PILL FORMAT)  ***************************/

export default function Tab(theme) {
  return {
    MuiTab: {
      defaultProps: {
        disableFocusRipple: true
      },
      styleOverrides: {
        root: {
          ...theme.typography.button,
          fontWeight: 500,
          fontSize: '12px',
          textTransform: 'none',
          minWidth: 'auto',
          minHeight: 36,
          padding: '8px 18px',
          borderRadius: 9999,
          color: 'rgba(29, 29, 29, 0.7)',
          transition: 'all 0.2s ease',
          margin: '0 4px',
          '&:hover': {
            color: '#1D1D1D',
            backgroundColor: 'rgba(29, 29, 29, 0.04)'
          },
          '&.Mui-selected': {
            backgroundColor: '#1D1D1D',
            color: '#FFFFFF !important',
            fontWeight: 500,
            '&:hover': {
              backgroundColor: '#333333'
            }
          },
          ...theme.applyStyles('dark', {
            color: 'rgba(255, 255, 255, 0.7)',
            '&:hover': {
              color: '#FFFFFF',
              backgroundColor: 'rgba(255, 255, 255, 0.08)'
            },
            '&.Mui-selected': {
              backgroundColor: '#F0FD63',
              color: '#1D1D1D !important',
              '&:hover': {
                backgroundColor: '#D4E33E'
              }
            }
          })
        }
      }
    }
  };
}
