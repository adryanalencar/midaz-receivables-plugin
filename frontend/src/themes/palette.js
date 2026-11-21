// @project
import { extendPaletteWithChannels, withAlpha } from '@/utils/colorUtils';

/***************************  ALIGNLY - PALETTE  ***************************/

export function buildPalette() {
  // Alignly Brand Tokens
  const surfaceDark = '#1D1D1D';
  const surfaceDarkElevated = '#262626';
  const brandLime = '#F0FD63';
  const brandSlate = '#919EAB';
  const surfaceLight = '#FFFFFF';
  const surfaceSubtle = '#F3F3F3';
  const borderSubtle = '#E8E8E8';

  const textPrimary = surfaceDark;
  const textSecondary = brandSlate;

  const lightPalette = {
    primary: {
      lighter: surfaceSubtle,
      light: '#333333',
      main: surfaceDark,
      dark: '#141414',
      darker: '#000000',
      contrastText: '#FFFFFF'
    },
    secondary: {
      lighter: surfaceSubtle,
      light: '#B0B8C1',
      main: brandSlate,
      dark: '#637381',
      darker: '#454F5B',
      contrastText: '#FFFFFF'
    },
    lime: {
      lighter: withAlpha(brandLime, 0.25),
      light: '#F7FE9F',
      main: brandLime,
      dark: '#D4E33E',
      darker: '#869600',
      contrastText: surfaceDark
    },
    error: {
      lighter: '#FEE2E2',
      light: '#FCA5A5',
      main: '#991B1B',
      dark: '#7F1D1D',
      darker: '#450A0A',
      contrastText: '#FFFFFF'
    },
    warning: {
      lighter: '#FEF3C7',
      light: '#FCD34D',
      main: '#D97706',
      dark: '#B45309',
      darker: '#78350F',
      contrastText: surfaceDark
    },
    success: {
      lighter: withAlpha(brandLime, 0.35),
      light: brandLime,
      main: '#22892F',
      dark: '#14631D',
      darker: '#0A380F',
      contrastText: surfaceDark
    },
    info: {
      lighter: surfaceSubtle,
      light: brandSlate,
      main: surfaceDark,
      dark: '#141414',
      darker: '#000000',
      contrastText: '#FFFFFF'
    },
    grey: {
      50: '#FAFAFA',
      100: surfaceSubtle, // #F3F3F3 - surface container / subtle
      200: borderSubtle, // #E8E8E8 - border
      300: '#DCDCDC',
      400: '#B0B8C1',
      500: brandSlate, // #919EAB - slate
      600: '#637381',
      700: '#454F5B',
      800: surfaceDarkElevated, // #262626
      900: surfaceDark // #1D1D1D
    },
    text: {
      primary: textPrimary,
      secondary: textSecondary,
      disabled: '#B0B8C1'
    },
    divider: borderSubtle,
    background: {
      default: surfaceSubtle, // Canvas background
      paper: surfaceLight // Card surface
    },
    action: {
      hover: withAlpha(surfaceDark, 0.05),
      selected: withAlpha(brandLime, 0.25),
      disabled: withAlpha(surfaceDark, 0.26),
      disabledBackground: surfaceSubtle
    }
  };

  // @dark mode (Alignly high-contrast dark surfaces)
  const textPrimaryDark = '#FFFFFF';
  const textSecondaryDark = brandSlate;
  const dividerDark = '#2A2A2A';
  const backgroundDark = '#141414';
  const paperDark = surfaceDark;

  const darkPalette = {
    primary: {
      lighter: '#333333',
      light: '#F7FE9F',
      main: brandLime, // Lime accent in dark mode
      dark: '#D4E33E',
      darker: '#869600',
      contrastText: surfaceDark
    },
    secondary: {
      lighter: '#2A2A2A',
      light: '#B0B8C1',
      main: brandSlate,
      dark: '#E0E0FF',
      darker: '#F1EFFF',
      contrastText: '#FFFFFF'
    },
    lime: {
      lighter: withAlpha(brandLime, 0.25),
      light: '#F7FE9F',
      main: brandLime,
      dark: '#D4E33E',
      darker: '#869600',
      contrastText: surfaceDark
    },
    error: {
      lighter: '#450A0A',
      light: '#F87171',
      main: '#EF4444',
      dark: '#B91C1C',
      darker: '#7F1D1D',
      contrastText: '#FFFFFF'
    },
    warning: {
      lighter: '#451A03',
      light: '#FBBF24',
      main: '#F59E0B',
      dark: '#D97706',
      darker: '#92400E',
      contrastText: surfaceDark
    },
    success: {
      lighter: withAlpha(brandLime, 0.25),
      light: brandLime,
      main: brandLime,
      dark: '#14631D',
      darker: '#0A380F',
      contrastText: surfaceDark
    },
    info: {
      lighter: '#262626',
      light: brandSlate,
      main: '#FFFFFF',
      dark: '#DCDCDC',
      darker: '#B0B8C1',
      contrastText: surfaceDark
    },
    grey: {
      50: '#141414',
      100: '#1A1A1A',
      200: dividerDark,
      300: '#333333',
      400: '#4D4D4D',
      500: brandSlate,
      600: '#808080',
      700: '#999999',
      800: surfaceDarkElevated,
      900: '#FAFAFA'
    },
    text: {
      primary: textPrimaryDark,
      secondary: textSecondaryDark,
      disabled: '#666666'
    },
    divider: dividerDark,
    background: {
      default: backgroundDark,
      paper: paperDark
    },
    action: {
      hover: withAlpha('#FFFFFF', 0.08),
      selected: withAlpha(brandLime, 0.2),
      disabled: withAlpha('#FFFFFF', 0.3),
      disabledBackground: withAlpha('#FFFFFF', 0.08)
    }
  };

  const commonColor = { common: { black: '#000', white: '#fff' } };

  const extendedLight = extendPaletteWithChannels(lightPalette);
  const extendedDark = extendPaletteWithChannels(darkPalette);
  const extendedCommon = extendPaletteWithChannels(commonColor);

  return {
    light: {
      mode: 'light',
      ...extendedCommon,
      ...extendedLight
    },
    dark: {
      mode: 'dark',
      ...extendedCommon,
      ...extendedDark
    }
  };
}
