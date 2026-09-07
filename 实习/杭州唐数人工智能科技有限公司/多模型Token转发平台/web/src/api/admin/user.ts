/* 用户与 Token 管理 API */
import { del, get, patch, post, put } from '../http'
import type {
  BalanceAdjustReq,
  BalanceRecord,
  PageData,
  PageQuery,
  TokenCreateReq,
  TokenCreateResp,
  User,
  UserCreateReq,
  UserStatusReq,
  UserToken,
  UserUpdateReq,
} from '../types'

const BASE = '/admin/users'

export interface UserListQuery extends PageQuery {
  keyword?: string
  status?: string
  role?: string
}

export function listUsers(query: UserListQuery): Promise<PageData<User>> {
  return get<PageData<User>>(BASE, { params: query })
}

export function getUser(id: string): Promise<User> {
  return get<User>(`${BASE}/${id}`)
}

export function createUser(req: UserCreateReq): Promise<User> {
  return post<User>(BASE, req)
}

export function updateUser(id: string, req: UserUpdateReq): Promise<User> {
  return put<User>(`${BASE}/${id}`, req)
}

export function updateUserStatus(id: string, req: UserStatusReq): Promise<User> {
  return patch<User>(`${BASE}/${id}/status`, req)
}

/** Token 列表不分页：后端一次性返回该用户全部 Token */
export function listTokens(userId: string): Promise<{ list: UserToken[] }> {
  return get<{ list: UserToken[] }>(`${BASE}/${userId}/tokens`)
}

export function createToken(userId: string, req: TokenCreateReq): Promise<TokenCreateResp> {
  return post<TokenCreateResp>(`${BASE}/${userId}/tokens`, req)
}

export function deleteToken(userId: string, tokenId: string): Promise<null> {
  return del<null>(`${BASE}/${userId}/tokens/${tokenId}`)
}

export function updateTokenStatus(
  userId: string,
  tokenId: string,
  isEnabled: boolean,
): Promise<UserToken> {
  return patch<UserToken>(`${BASE}/${userId}/tokens/${tokenId}/status`, { is_enabled: isEnabled })
}

export function adjustBalance(userId: string, req: BalanceAdjustReq): Promise<BalanceRecord> {
  return post<BalanceRecord>(`${BASE}/${userId}/balance-adjust`, req)
}
