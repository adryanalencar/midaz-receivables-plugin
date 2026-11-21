import PropTypes from 'prop-types';
import Box from '@mui/material/Box';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import BentoCard from './BentoCard';
import StatusBadge from './StatusBadge';

/**
 * Alignly KPI Card
 * High contrast Bento card showing key metrics with badge and optional icon/action
 */
export default function KpiCard({
  title,
  value,
  subtitle,
  badgeText,
  badgeVariant = 'lime',
  variant = 'white',
  icon,
  action,
  sx = {}
}) {
  const isLime = variant === 'lime';
  const isDark = variant === 'dark';
  const isSlate = variant === 'slate';

  const textColor = isLime || variant === 'white' || variant === 'subtle' ? '#1D1D1D' : '#FFFFFF';
  const secondaryColor = isLime
    ? 'rgba(29, 29, 29, 0.7)'
    : isDark
      ? '#919EAB'
      : isSlate
        ? 'rgba(255, 255, 255, 0.85)'
        : '#919EAB';

  return (
    <BentoCard variant={variant} sx={{ height: '100%', display: 'flex', flexDirection: 'column', justifyContent: 'space-between', ...sx }}>
      <Stack direction="row" justifyContent="space-between" alignItems="flex-start" sx={{ mb: 2 }}>
        <Box>
          <Typography
            sx={{
              fontSize: '13px',
              fontWeight: 500,
              color: secondaryColor,
              letterSpacing: '-0.01em',
              mb: 0.5
            }}
          >
            {title}
          </Typography>
          <Typography
            sx={{
              fontFamily: "'Plus Jakarta Sans', sans-serif",
              fontSize: { xs: '24px', sm: '28px', md: '32px' },
              fontWeight: 700,
              color: textColor,
              letterSpacing: '-0.03em',
              lineHeight: 1.15
            }}
          >
            {value}
          </Typography>
        </Box>
        {icon && (
          <Box
            sx={{
              p: 1,
              borderRadius: '9999px',
              bgcolor: isLime ? 'rgba(29, 29, 29, 0.08)' : isDark ? '#262626' : '#F3F3F3',
              color: textColor,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center'
            }}
          >
            {icon}
          </Box>
        )}
      </Stack>

      <Stack direction="row" justifyContent="space-between" alignItems="center" sx={{ mt: 'auto', pt: 1 }}>
        <Stack direction="row" spacing={1} alignItems="center">
          {badgeText && (
            <StatusBadge variant={badgeVariant}>
              {badgeText}
            </StatusBadge>
          )}
          {subtitle && (
            <Typography
              sx={{
                fontSize: '12px',
                color: secondaryColor,
                fontWeight: 400
              }}
            >
              {subtitle}
            </Typography>
          )}
        </Stack>
        {action}
      </Stack>
    </BentoCard>
  );
}

KpiCard.propTypes = {
  title: PropTypes.string.isRequired,
  value: PropTypes.oneOfType([PropTypes.string, PropTypes.number]).isRequired,
  subtitle: PropTypes.string,
  badgeText: PropTypes.string,
  badgeVariant: PropTypes.string,
  variant: PropTypes.oneOf(['white', 'lime', 'dark', 'slate', 'subtle']),
  icon: PropTypes.node,
  action: PropTypes.node,
  sx: PropTypes.object
};
