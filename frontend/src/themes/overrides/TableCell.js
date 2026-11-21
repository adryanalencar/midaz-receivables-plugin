/***************************  ALIGNLY OVERRIDES - TABLE CELL  ***************************/

export default function TableCell(theme) {
  return {
    MuiTableCell: {
      styleOverrides: {
        root: ({ ownerState }) => {
          const baseStyle = {
            whiteSpace: 'nowrap',
            borderBottom: `1px solid ${theme.vars.palette.grey[100]}`,
            borderLeft: 'none',
            borderRight: 'none',
            borderTop: 'none'
          };

          const align = ownerState.align;

          if (align === 'right') {
            return {
              ...baseStyle,
              justifyContent: 'flex-end',
              textAlign: 'right',
              '& > *': {
                justifyContent: 'flex-end',
                margin: '0 0 0 auto'
              }
            };
          }

          if (align === 'center') {
            return {
              ...baseStyle,
              justifyContent: 'center',
              textAlign: 'center',
              '& > *': {
                justifyContent: 'center',
                margin: '0 auto'
              }
            };
          }

          return baseStyle;
        },
        head: {
          ...theme.typography.caption,
          color: theme.vars.palette.text.secondary,
          fontWeight: 500,
          fontSize: '11px',
          textTransform: 'uppercase',
          letterSpacing: '0.04em',
          padding: '12px 20px',
          backgroundColor: theme.vars.palette.background.default,
          borderBottom: `1px solid ${theme.vars.palette.divider}`
        },
        body: {
          ...theme.typography.body2,
          color: theme.vars.palette.text.primary,
          padding: '14px 20px'
        }
      }
    }
  };
}
