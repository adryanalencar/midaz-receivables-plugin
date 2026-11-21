// @project
import { AuthType } from '@/enum';

/***************************  THEME CONSTANT  ***************************/

export const APP_DEFAULT_PATH = '/receivables';
export const APP_SUPPORT_PATH = 'https://phoenixcoded.authordesk.app/';

export const DRAWER_WIDTH = 240;
export const MINI_DRAWER_WIDTH = 76 + 1; // 1px - for right-side border

export const CSS_VAR_PREFIX = '';

/***************************  AUTH CONSTANT  ***************************/

export const AUTH_USER_KEY = 'auth-user';
export const AUTH_PROVIDER = AuthType.MOCK;

/***************************  THEME ENUM  ***************************/

export let Themes;

(function (Themes) {
  Themes['THEME_ALIGNLY'] = 'alignly';
  Themes['THEME_CRM'] = 'crm';
  Themes['THEME_AI'] = 'ai';
  Themes['THEME_HOSTING'] = 'hosting';
})(Themes || (Themes = {}));

export let ThemeMode;

(function (ThemeMode) {
  ThemeMode['LIGHT'] = 'light';
  ThemeMode['DARK'] = 'dark';
  ThemeMode['SYSTEM'] = 'system';
})(ThemeMode || (ThemeMode = {}));

export let ThemeDirection;

(function (ThemeDirection) {
  ThemeDirection['LTR'] = 'ltr';
  ThemeDirection['RTL'] = 'rtl';
})(ThemeDirection || (ThemeDirection = {}));

export let ThemeI18n;

(function (ThemeI18n) {
  ThemeI18n['EN'] = 'en';
  ThemeI18n['FR'] = 'fr';
  ThemeI18n['RO'] = 'ro';
  ThemeI18n['ZH'] = 'zh';
})(ThemeI18n || (ThemeI18n = {}));

export let ThemeFonts;

(function (ThemeFonts) {
  ThemeFonts['FONT_ALIGNLY'] = "'Inter', system-ui, -apple-system, sans-serif";
  ThemeFonts['FONT_DISPLAY'] = "'Plus Jakarta Sans', 'Gilroy', 'Poppins', sans-serif";
  ThemeFonts['FONT_ROBOTO'] = "'Roboto', sans-serif";
  ThemeFonts['FONT_ARCHIVO'] = "'Archivo', sans-serif";
  ThemeFonts['FONT_FIGTREE'] = "'Figtree', sans-serif";
})(ThemeFonts || (ThemeFonts = {}));

export const DEFAULT_THEME_MODE = ThemeMode.LIGHT;

/***************************  CONFIG  ***************************/

const config = {
  currentTheme: Themes.THEME_ALIGNLY,
  themeDirection: ThemeDirection.LTR,
  miniDrawer: false,
  i18n: ThemeI18n.EN
};

export default config;
