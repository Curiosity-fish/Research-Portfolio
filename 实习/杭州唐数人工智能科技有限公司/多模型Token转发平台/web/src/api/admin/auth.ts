/* 认证 API */
import { get, post } from '../http'
import type { AdminMe, LoginReq, LoginResp } from '../types'

const BASE = '/admin'

export function login(req: LoginReq): Promise<LoginResp> {
  return post<LoginResp>(`${BASE}/login`, req)
}

export function fetchMe(): Promise<AdminMe> {
  return get<AdminMe>(`${BASE}/me`)
}
