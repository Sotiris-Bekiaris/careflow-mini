import { createApp } from 'vue'
import './style.css'
import 'vuetify/styles'

import App from './App.vue'
import router from './router'
import pinia from './stores'
import vuetify from './plugins/vuetify'

const app = createApp(App)

app.use(pinia)
app.use(router)
app.use(vuetify)

app.mount('#app')
