/** @type {import('tailwindcss').Config} */
export default {
  content: [
    './index.html',
    './src/**/*.{js,ts,jsx,tsx}',
  ],
  darkMode: ['class', '[data-color-scheme="dark"]'],
  theme: {
    extend: {
      colors: {
        brand: {
          lime: '#F0FD63',
          slate: '#919EAB',
        },
        surface: {
          dark: '#1D1D1D',
          'dark-elevated': '#262626',
          light: '#FFFFFF',
          subtle: '#F3F3F3',
        },
        border: {
          subtle: '#E8E8E8',
        },
      },
      fontFamily: {
        display: ['"Plus Jakarta Sans"', 'Gilroy', 'Poppins', 'sans-serif'],
        sans: ['Inter', 'system-ui', 'sans-serif'],
      },
      borderRadius: {
        container: '24px',
        bento: '20px',
        card: '16px',
        pill: '9999px',
      },
      letterSpacing: {
        tightest: '-0.03em',
        tighter: '-0.02em',
      },
    },
  },
  plugins: [],
  corePlugins: {
    preflight: false, // Keep MUI preflight intact while providing Tailwind utilities
  },
};
