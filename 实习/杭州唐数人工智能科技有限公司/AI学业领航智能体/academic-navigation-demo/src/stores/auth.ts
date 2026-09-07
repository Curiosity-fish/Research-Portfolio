import { ref } from 'vue'
import { defineStore } from 'pinia'
import type { User } from '@/types'
import request from '@/utils/request'

type Role = 'student' | 'teacher' | 'course_teacher' | 'department' | 'dean'

const routeMap: Record<Role, string> = {
  student:       '/student/dashboard',
  teacher:       '/teacher/dashboard',
  course_teacher:'/course-teacher/dashboard',
  department:    '/department/overview',
  dean:          '/dean/dashboard',
}

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  const isLoggedIn = ref(false)

  function init() {
    const saved = localStorage.getItem('user')
    const token = localStorage.getItem('token')
    if (saved && token && saved !== 'undefined' && saved !== 'null') {
      try {
        user.value = JSON.parse(saved)
        isLoggedIn.value = true
      } catch {
        localStorage.removeItem('user')
        localStorage.removeItem('token')
      }
    }
  }

  async function restore(): Promise<boolean> {
    const token = localStorage.getItem('token')
    if (!token) return false
    try {
      const res = await request.get('/auth/me') as any
      const userInfo = res.data ?? res
      localStorage.setItem('user', JSON.stringify(userInfo))
      user.value = userInfo
      isLoggedIn.value = true
      return true
    } catch {
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      user.value = null
      isLoggedIn.value = false
      return false
    }
  }

  async function login(account: string, password: string): Promise<{ ok: boolean; path?: string; error?: string }> {
    try {
      const res = await request.post('/auth/login', { account, password }) as any
      const { token, user: userInfo } = res.data ?? res

      // 持久化到 localStorage
      localStorage.setItem('token', token)
      localStorage.setItem('user', JSON.stringify(userInfo))

      user.value = userInfo
      isLoggedIn.value = true

      const path = routeMap[userInfo.role as Role] ?? '/student/dashboard'
      return { ok: true, path }
    } catch (err: any) {
      const message = err?.message || err.response?.data?.message || '登录失败'
      return { ok: false, error: message }
    }
  }

  async function logout() {
    try {
      await request.post('/auth/logout')
    } catch {
      // 即使请求失败，本地也要清除
    } finally {
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      user.value = null
      isLoggedIn.value = false
    }
  }

  return { user, isLoggedIn, init, restore, login, logout }
})
