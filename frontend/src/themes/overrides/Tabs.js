/***************************  ALIGNLY OVERRIDES - TABS (PILL NAVIGATION)  ***************************/

export default function Tabs(theme) {
  return {
    MuiTabs: {
      defaultProps: {
        indicatorColor: 'primary',
        textColor: 'inherit'
      },
      styleOverrides: {
        root: {
          minHeight: 40,
          '& .MuiTabs-indicator': {
            display: 'none' // Alignly uses individual Pill background on active tab
          },
          '& .MuiTabs-flexContainer': {
            gap: 4,
            alignItems: 'center'
          }
        }
      }
    }
  };
}
