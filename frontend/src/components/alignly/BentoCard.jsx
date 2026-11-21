import PropTypes from 'prop-types';
import Box from '@mui/material/Box';

/**
 * Alignly Bento Card (rounded-[20px], no heavy shadows)
 * @param {'white' | 'lime' | 'dark' | 'slate' | 'subtle'} variant
 */
export default function BentoCard({ children, variant = 'white', sx = {}, ...props }) {
  const styles = {
    white: {
      bgcolor: '#FFFFFF',
      color: '#1D1D1D',
      border: '1px solid #E8E8E8',
      boxShadow: 'none'
    },
    lime: {
      bgcolor: '#F0FD63',
      color: '#1D1D1D',
      border: 'none',
      boxShadow: '0 0 24px rgba(240, 253, 99, 0.15)'
    },
    dark: {
      bgcolor: '#1D1D1D',
      color: '#FFFFFF',
      border: 'none',
      boxShadow: 'none'
    },
    slate: {
      bgcolor: '#919EAB',
      color: '#FFFFFF',
      border: 'none',
      boxShadow: 'none'
    },
    subtle: {
      bgcolor: '#F3F3F3',
      color: '#1D1D1D',
      border: '1px solid #E8E8E8',
      boxShadow: 'none'
    }
  };

  const currentStyle = styles[variant] || styles.white;

  return (
    <Box
      sx={{
        borderRadius: '20px',
        p: { xs: 2.5, md: 3 },
        position: 'relative',
        overflow: 'hidden',
        transition: 'transform 0.15s ease, box-shadow 0.15s ease',
        ...currentStyle,
        ...sx
      }}
      {...props}
    >
      {children}
    </Box>
  );
}

BentoCard.propTypes = {
  children: PropTypes.node,
  variant: PropTypes.oneOf(['white', 'lime', 'dark', 'slate', 'subtle']),
  sx: PropTypes.object
};
