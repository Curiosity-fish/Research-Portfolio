/* 管理员认证状态：登录凭证 + 当前用户 */
import { defineStore } from 'pinia'
import { ref } from 'vue'
import { fetchMe, login } from '@/api/admin/auth'
import { clearToken, setToken } from '@/api/http'
import type { AdminMe, LoginReq } from '@/api/types'

export const useAuthStore = defineStore('auth', () => {
  const me = ref<AdminMe | null>(null)
  /** 登录用户名：/me 响应不含 username，登录时落盘供布局展示 */
  const username = ref<string>(localStorage.getItem('admin_username') ?? '')
  /** 是否有本地凭证（未校验有效性，路由守卫据此决定是否放行） */
  const hasCredential = ref<boolean>(!!localStorage.getItem('admin_access_token'))

  async function loginAndFetch(req: LoginReq): Promise<void> {
    const resp = await login(req)
    setToken(resp.access_token)
    localStorage.setItem('admin_username', req.username)
    username.value = req.username
    hasCredential.value = true
    me.value = await fetchMe()
  }

  async function ensureMe(): Promise<void> {
    if (!me.value) {
      me.value = await fetchMe()
    }
  }

  function logout(): void {
    clearToken()
    localStorage.removeItem('admin_username')
    hasCredential.value = false
    me.value = null
    username.value = ''
  }

  return { me, username, hasCredential, loginAndFetch, ensureMe, logout }
})
