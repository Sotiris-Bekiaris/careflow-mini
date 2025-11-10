/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{vue,js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        primary: '#1976D2',
        secondary: '#424242',
        success: '#4CAF50',
        warning: '#FB8C00',
        error: '#D32F2F',
        info: '#2196F3',
      },
      fontFamily: {
        sans: ['Roboto', 'system-ui', 'sans-serif'],
      },
    },
  },
  plugins: [],
}
