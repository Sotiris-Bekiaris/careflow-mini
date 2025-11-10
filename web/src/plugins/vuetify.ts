import 'vuetify/styles'
import { createVuetify } from 'vuetify'
import * as components from 'vuetify/components'
import * as directives from 'vuetify/directives'
import { mdi } from 'vuetify/iconsets/mdi'

export default createVuetify({
  components,
  directives,
  iconsets: {
    mdi,
  },
  theme: {
    themes: {
      light: {
        colors: {
          primary: '#1976D2',
          secondary: '#424242',
          success: '#4CAF50',
          warning: '#FB8C00',
          error: '#D32F2F',
          info: '#2196F3',
          background: '#FFFFFF',
          surface: '#FAFAFA',
        },
      },
      dark: {
        colors: {
          primary: '#1976D2',
          secondary: '#424242',
          success: '#4CAF50',
          warning: '#FB8C00',
          error: '#D32F2F',
          info: '#2196F3',
          background: '#121212',
          surface: '#1E1E1E',
        },
      },
    },
  },
})
