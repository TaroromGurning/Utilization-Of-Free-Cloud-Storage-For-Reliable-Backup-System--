import { createApp } from 'vue'
import { createPinia } from 'pinia' // Import Pinia

import App from './App.vue'
import router from './router' // Import Router
import './assets/main.css'

const app = createApp(App)

app.use(createPinia()) // Aktifkan Pinia
app.use(router)      // Aktifkan Router

app.mount('#app')