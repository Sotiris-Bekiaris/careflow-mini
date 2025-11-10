import 'vuetify/styles'
import { createVuetify, type ThemeDefinition } from 'vuetify'
import * as components from 'vuetify/components'
import * as directives from 'vuetify/directives'
import { mdi } from 'vuetify/iconsets/mdi'

const appleLight: ThemeDefinition = {
  dark: false,
  colors: {
    primary: '#0a84ff',
    secondary: '#1c1c1e',
    success: '#34c759',
    warning: '#ff9f0a',
    error: '#ff453a',
    info: '#64d2ff',
    background: '#f5f5f7',
    surface: '#ffffff',
    'surface-variant': '#f9fafb',
    outline: '#e5e7eb',
  },
}

const appleDark: ThemeDefinition = {
  dark: true,
  colors: {
    primary: '#0a84ff',
    secondary: '#f2f2f7',
    success: '#30d158',
    warning: '#ffd60a',
    error: '#ff453a',
    info: '#64d2ff',
    background: '#000000',
    surface: '#1c1c1e',
    'surface-variant': '#2c2c2e',
    outline: '#3a3a3c',
  },
}

export default createVuetify({
  components,
  directives,
  icons: {
    defaultSet: 'mdi',
    sets: {
      mdi,
    },
  },
  theme: {
    defaultTheme: 'appleLight',
    themes: {
      appleLight,
      appleDark,
    },
    variations: {
      colors: ['primary', 'secondary', 'surface'],
      lighten: 2,
      darken: 2,
    },
  },
})
