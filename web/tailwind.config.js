/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{vue,js,ts,jsx,tsx}'],
  theme: {
    extend: {
      colors: {
        primary: '#0a84ff',
        secondary: '#1c1c1e',
        success: '#34c759',
        warning: '#ff9f0a',
        error: '#ff453a',
        info: '#64d2ff',
        slate: {
          50: '#f5f5f7',
          100: '#e9ebf2',
          300: '#c7c9d3',
          500: '#7a7c86',
          700: '#3a3a43',
          900: '#111114',
        },
      },
      fontFamily: {
        sans: ['SF Pro Display', 'SF Pro Text', 'Inter', 'Helvetica Neue', 'sans-serif'],
      },
      borderRadius: {
        xl: '1.75rem',
      },
    },
  },
  plugins: [],
}
