import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import './composables/useTheme'
import './styles/tokens.css'
import './styles/base.css'

const app = createApp(App)

app.use(router)

app.mount('#app')
