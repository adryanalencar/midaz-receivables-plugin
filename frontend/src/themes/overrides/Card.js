/***************************  ALIGNLY OVERRIDES - CARD (BENTO 20PX FLAT)  ***************************/

export default function Card(theme) {
  return {
    MuiCard: {
      defaultProps: {
        elevation: 0
      },
      styleOverrides: {
        root: {
          borderRadius: 20,
          border: `1px solid ${theme.vars.palette.divider}`,
          boxShadow: 'none',
          backgroundImage: 'none',
          backgroundColor: theme.vars.palette.background.paper,
          transition: 'all 0.2s ease',
          overflow: 'hidden'
        }
      }
    },
    MuiPaper: {
      defaultProps: {
        elevation: 0
      },
      styleOverrides: {
        root: {
          backgroundImage: 'none'
        },
        rounded: {
          borderRadius: 20
        }
      }
    }
  };
}
