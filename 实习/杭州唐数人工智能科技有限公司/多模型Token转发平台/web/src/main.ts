/* 应用入口：全局样式（顺序即优先级）→ Element Plus → 图标全局注册 → 路由与状态 */
import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import 'element-plus/dist/index.css'

import '@fontsource/noto-sans-sc/400.css'
import '@fontsource/noto-sans-sc/500.css'
import '@fontsource/noto-sans-sc/700.css'
import '@fontsource/noto-serif-sc/700.css'
import '@fontsource/noto-serif-sc/900.css'

import './styles/tokens.css'
import './styles/element-theme.css'
import './styles/base.css'

import App from './App.vue'
import router from './router'

const app = createApp(App)

// EP 图标全局注册：模板里直接写组件名（如 <el-icon><User /></el-icon>）
for (const [name, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(name, component)
}

app.use(createPinia())
app.use(router)
app.use(ElementPlus, { locale: zhCn })
app.mount('#app')
