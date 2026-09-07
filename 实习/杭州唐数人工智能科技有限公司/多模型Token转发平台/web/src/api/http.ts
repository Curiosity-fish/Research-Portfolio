/* HTTP 基础设施：axios 实例、JWT 注入、统一错误处理、401 跳登录。
   拦截器只透传成功响应与处理错误，信封解包在各请求方法内完成，类型链路最清晰。 */

import axios, { AxiosError, type AxiosRequestConfig } from 'axios'
import { ElMessage } from 'element-plus'
import type { ApiErrorBody, Envelope } from './types'

const TOKEN_KEY = 'admin_access_token'

export function getToken(): string {
  return localStorage.getItem(TOKEN_KEY) ?? ''
}

export function setToken(token: string): void {
  localStorage.setItem(TOKEN_KEY, token)
}

export function clearToken(): void {
  localStorage.removeItem(TOKEN_KEY)
}

/** 401 时清理凭证并回登录页；整页跳转而非 router push，避免拦截器与路由循环依赖 */
function handleUnauthorized(): void {
  clearToken()
  if (!location.pathname.startsWith('/admin/login')) {
    location.assign('/admin/login')
  }
}

export const http = axios.create({
  baseURL: '/api/v1',
  timeout: 30000,
})

http.interceptors.request.use((config) => {
  const token = getToken()
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

http.interceptors.response.use(
  (response) => response,
  (error: AxiosError<ApiErrorBody>) => {
    const status = error.response?.status
    const errBody = error.response?.data?.error

    if (status === 401) {
      handleUnauthorized()
      return Promise.reject(error)
    }

    // 限流单独提示，其余按后端 message 展示
    const message =
      status === 429
        ? '请求过于频繁，请稍后再试'
        : (errBody?.message ?? error.message ?? '网络异常，请稍后重试')
    ElMessage.error(message)
    return Promise.reject(error)
  },
)

/** 解包响应信封：后端成功响应统一为 { data: ... }（无 code/message 包裹），直接取 data 本体 */
function unwrap<T>(body: unknown): T {
  if (body === null || body === undefined || body === '') return null as T
  if (typeof body === 'object' && !Array.isArray(body) && 'data' in body) {
    return (body as { data: T }).data
  }
  return body as T
}

export async function get<T>(url: string, config?: AxiosRequestConfig): Promise<T> {
  const resp = await http.get<Envelope<T>>(url, config)
  return unwrap(resp.data)
}

export async function post<T>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<T> {
  const resp = await http.post<Envelope<T>>(url, data, config)
  return unwrap(resp.data)
}

export async function put<T>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<T> {
  const resp = await http.put<Envelope<T>>(url, data, config)
  return unwrap(resp.data)
}

export async function patch<T>(
  url: string,
  data?: unknown,
  config?: AxiosRequestConfig,
): Promise<T> {
  const resp = await http.patch<Envelope<T>>(url, data, config)
  return unwrap(resp.data)
}

export async function del<T>(url: string, config?: AxiosRequestConfig): Promise<T> {
  const resp = await http.delete<Envelope<T>>(url, config)
  return unwrap(resp.data)
}
