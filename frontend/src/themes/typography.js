// @project
import { ThemeFonts } from '@/config';

/***************************  ALIGNLY - TYPOGRAPHY  ***************************/

export default function typography() {
  const displayFont = ThemeFonts.FONT_DISPLAY || "'Plus Jakarta Sans', 'Gilroy', sans-serif";
  const bodyFont = ThemeFonts.FONT_ALIGNLY || "'Inter', system-ui, sans-serif";

  return {
    fontFamily: bodyFont,
    letterSpacing: '-0.01em',

    // Heading - Display / Hero
    h1: {
      fontFamily: displayFont,
      fontWeight: 600,
      fontSize: 40,
      lineHeight: '48px',
      letterSpacing: '-0.03em'
    },

    // Heading - Section / Page Title
    h2: {
      fontFamily: displayFont,
      fontWeight: 600,
      fontSize: 32,
      lineHeight: '40px',
      letterSpacing: '-0.02em'
    },

    // Heading - Card / Sub-section
    h3: {
      fontFamily: displayFont,
      fontWeight: 600,
      fontSize: 24,
      lineHeight: '32px',
      letterSpacing: '-0.02em'
    },

    // Heading - Medium Card Title
    h4: {
      fontFamily: displayFont,
      fontWeight: 600,
      fontSize: 20,
      lineHeight: '28px',
      letterSpacing: '-0.01em'
    },

    // Heading - Small Card Title
    h5: {
      fontFamily: displayFont,
      fontWeight: 500,
      fontSize: 16,
      lineHeight: '24px',
      letterSpacing: '-0.01em'
    },

    // Heading - Micro Card Title
    h6: {
      fontFamily: displayFont,
      fontWeight: 500,
      fontSize: 14,
      lineHeight: '20px'
    },

    // Subtitle - 1
    subtitle1: {
      fontFamily: bodyFont,
      fontWeight: 500,
      fontSize: 14,
      lineHeight: '20px'
    },

    // Subtitle - 2
    subtitle2: {
      fontFamily: bodyFont,
      fontWeight: 500,
      fontSize: 12,
      lineHeight: '18px'
    },

    // Body - 1
    body1: {
      fontFamily: bodyFont,
      fontWeight: 400,
      fontSize: 14,
      lineHeight: '22px'
    },

    // Body - 2
    body2: {
      fontFamily: bodyFont,
      fontWeight: 400,
      fontSize: 12,
      lineHeight: '18px'
    },

    // Caption
    caption: {
      fontFamily: bodyFont,
      fontWeight: 400,
      fontSize: 11,
      lineHeight: '16px'
    },

    // Overline
    overline: {
      fontFamily: bodyFont,
      fontWeight: 600,
      fontSize: 10,
      lineHeight: '14px',
      textTransform: 'uppercase',
      letterSpacing: '0.05em'
    },

    // Button
    button: {
      fontFamily: bodyFont,
      fontWeight: 500,
      fontSize: 12,
      lineHeight: '18px',
      textTransform: 'none'
    }
  };
}
