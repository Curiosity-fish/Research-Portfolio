/* 路由表与守卫：/admin 需 JWT 凭证，401 由 http 拦截器统一处理 */
import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const routes: RouteRecordRaw[] = [
  {
    path: '/admin/login',
    name: 'admin-login',
    component: () => import('@/views/admin/LoginView.vue'),
    meta: { public: true, title: '登录' },
  },
  {
    path: '/admin',
    component: () => import('@/components/layout/AdminLayout.vue'),
    children: [
      {
        path: '',
        name: 'admin-dashboard',
        component: () => import('@/views/admin/DashboardView.vue'),
        meta: { title: '概览' },
      },
      {
        path: 'users',
        name: 'admin-users',
        component: () => import('@/views/admin/UsersView.vue'),
        meta: { title: '用户管理' },
      },
      // 以下为后续步骤实现的页面，先注册路由保证导航可用
      {
        path: 'relay',
        name: 'admin-relay',
        component: () => import('@/views/admin/RelayView.vue'),
        meta: { title: '转发资源' },
      },
      {
        path: 'finance',
        name: 'admin-finance',
        component: () => import('@/views/admin/FinanceView.vue'),
        meta: { title: '资金与计费' },
      },
      {
        path: 'notifications',
        name: 'admin-notifications',
        component: () => import('@/views/admin/NotificationsView.vue'),
        meta: { title: '通知' },
      },
      {
        path: 'alerts',
        name: 'admin-alerts',
        component: () => import('@/views/admin/AlertsView.vue'),
        meta: { title: '告警' },
      },
      {
        path: 'audit-logs',
        name: 'admin-audit-logs',
        component: () => import('@/views/admin/AuditLogsView.vue'),
        meta: { title: '审计日志' },
      },
    ],
  },
  { path: '/', redirect: '/admin' },
  { path: '/:pathMatch(.*)*', redirect: '/admin' },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach((to) => {
  const auth = useAuthStore()
  if (to.meta.public) {
    // 已登录用户访问登录页时回到首页
    if (auth.hasCredential && to.name === 'admin-login') return { name: 'admin-dashboard' }
    return true
  }
  if (!auth.hasCredential) {
    return { name: 'admin-login', query: { redirect: to.fullPath } }
  }
  return true
})

router.afterEach((to) => {
  document.title = to.meta.title ? `${to.meta.title} · 校园 Token 管理平台` : '校园 Token 管理平台'
})

export default router
