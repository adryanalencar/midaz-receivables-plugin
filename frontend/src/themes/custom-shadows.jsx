// @project
import { withAlpha } from '@/utils/colorUtils';

/***************************  ALIGNLY THEME - SHADOWS  ***************************/

export default function CustomShadows(palette) {
  const isDark = palette.mode === 'dark';
  const shadowBase = isDark ? '#000000' : '#1D1D1D';

  return {
    button: 'none',
    section: 'none',
    card: 'none',
    tooltip: `0px 4px 12px ${withAlpha(shadowBase, isDark ? 0.4 : 0.08)}`,
    popover: `0px 8px 30px ${withAlpha(shadowBase, isDark ? 0.5 : 0.04)}`,
    dialog: `0px 16px 40px ${withAlpha(shadowBase, isDark ? 0.6 : 0.06)}`,
    focus: `0px 0px 0px 2px #F0FD63`
  };
}
