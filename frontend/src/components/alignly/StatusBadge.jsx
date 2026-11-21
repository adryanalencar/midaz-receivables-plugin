import PropTypes from 'prop-types';
import Box from '@mui/material/Box';

/**
 * Alignly Status Badge (Always Pill format `rounded-full`)
 * @param {'success' | 'lime' | 'warning' | 'pending' | 'danger' | 'neutral' | 'dark'} variant
 */
export default function StatusBadge({ children, variant = 'neutral', icon, sx = {} }) {
  const styles = {
    lime: {
      bgcolor: '#F0FD63',
      color: '#1D1D1D',
      border: 'none'
    },
    success: {
      bgcolor: 'rgba(34, 197, 94, 0.12)',
      color: '#15803d',
      border: '1px solid rgba(34, 197, 94, 0.2)'
    },
    pending: {
      bgcolor: '#F3F3F3',
      color: '#1D1D1D',
      border: '1px solid #E8E8E8'
    },
    warning: {
      bgcolor: 'rgba(234, 179, 8, 0.12)',
      color: '#a16207',
      border: '1px solid rgba(234, 179, 8, 0.25)'
    },
    danger: {
      bgcolor: 'rgba(239, 68, 68, 0.1)',
      color: '#b91c1c',
      border: '1px solid rgba(239, 68, 68, 0.2)'
    },
    dark: {
      bgcolor: '#1D1D1D',
      color: '#FFFFFF',
      border: 'none'
    },
    neutral: {
      bgcolor: '#F3F3F3',
      color: '#666666',
      border: '1px solid #E8E8E8'
    }
  };

  const currentStyle = styles[variant] || styles.neutral;

  return (
    <Box
      component="span"
      sx={{
        display: 'inline-flex',
        alignItems: 'center',
        gap: 0.75,
        px: 1.5,
        py: 0.5,
        borderRadius: '9999px',
        fontSize: '11px',
        fontWeight: 600,
        lineHeight: 1,
        letterSpacing: '0.01em',
        ...currentStyle,
        ...sx
      }}
    >
      {icon && <Box component="span" sx={{ display: 'inline-flex', alignItems: 'center' }}>{icon}</Box>}
      {children}
    </Box>
  );
}

StatusBadge.propTypes = {
  children: PropTypes.node.isRequired,
  variant: PropTypes.oneOf(['success', 'lime', 'warning', 'pending', 'danger', 'neutral', 'dark']),
  icon: PropTypes.node,
  sx: PropTypes.object
};
