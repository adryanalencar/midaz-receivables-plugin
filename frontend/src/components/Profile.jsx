import PropTypes from 'prop-types';
// @mui
import Avatar from '@mui/material/Avatar';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';

// @types

// @icons
import { IconPhoto } from '@tabler/icons-react';

/***************************  PROFILE  ***************************/

export default function Profile({ avatar, title, caption, label, sx, titleProps, captionProps, placeholderIfEmpty }) {
  return (
    <Stack direction="row" sx={{ alignItems: 'center', justifyContent: 'space-between', gap: 1, width: 'fit-content', ...sx }}>
      {(avatar?.src || avatar?.children || placeholderIfEmpty) && (
        <Avatar
          {...avatar}
          alt="profile"
          sx={{
            bgcolor: '#F0FD63',
            color: '#1D1D1D',
            fontWeight: 700,
            fontSize: '12px',
            width: 32,
            height: 32,
            ...avatar?.sx,
            ...(placeholderIfEmpty && !avatar?.children && !avatar?.src && { fontSize: 20, '& svg': { width: 26, height: 26 } })
          }}
        >
          {avatar?.children || (!avatar?.src && placeholderIfEmpty && <IconPhoto stroke={1} />)}
        </Avatar>
      )}
      <Stack sx={{ gap: 0.25 }}>
        <Stack direction="row" sx={{ alignItems: 'center', gap: 0.5 }}>
          <Typography variant="subtitle2" {...titleProps} sx={{ color: 'text.primary', whiteSpace: 'nowrap', fontWeight: 600, ...titleProps?.sx }}>
            {title || (placeholderIfEmpty && 'N/A')}
          </Typography>
          {label}
        </Stack>
        <Typography variant="caption" {...captionProps} sx={{ color: '#919EAB', fontSize: '11px', ...captionProps?.sx }}>
          {caption || (placeholderIfEmpty && '---')}
        </Typography>
      </Stack>
    </Stack>
  );
}

Profile.propTypes = {
  avatar: PropTypes.any,
  title: PropTypes.any,
  caption: PropTypes.any,
  label: PropTypes.any,
  sx: PropTypes.any,
  titleProps: PropTypes.any,
  captionProps: PropTypes.any,
  placeholderIfEmpty: PropTypes.any
};
