/***************************  OVERRIDES - MENU  ***************************/

export default function ListItemButton() {
  return {
    MuiListItemButton: {
      styleOverrides: {
        root: {
          padding: '10px 12px',
          borderRadius: 12,
          transition: 'all 0.15s ease'
        }
      }
    }
  };
}
