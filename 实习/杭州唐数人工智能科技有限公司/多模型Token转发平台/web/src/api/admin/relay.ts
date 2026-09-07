/* 转发资源 API：平台 / 账号 / 模型 / 分组绑定。
   三个实体列表后端均不分页（一次性返回数组），分页由各 tab 客户端切片完成 */
import { del, get, post, put } from '../http'
import type {
  Account,
  AccountUpsertReq,
  AIModel,
  Group,
  GroupPlatformBindReq,
  ModelUpsertReq,
  Platform,
  PlatformUpsertReq,
} from '../types'

const P = '/admin/platforms'
const A = '/admin/accounts'
const M = '/admin/models'
const G = '/admin/groups'

export function listPlatforms(): Promise<Platform[]> {
  return get<Platform[]>(P)
}

export function createPlatform(req: PlatformUpsertReq): Promise<Platform> {
  return post<Platform>(P, req)
}

export function updatePlatform(id: string, req: PlatformUpsertReq): Promise<Platform> {
  return put<Platform>(`${P}/${id}`, req)
}

export function deletePlatform(id: string): Promise<null> {
  return del<null>(`${P}/${id}`)
}

export function listAccounts(): Promise<Account[]> {
  return get<Account[]>(A)
}

export function createAccount(req: AccountUpsertReq): Promise<Account> {
  return post<Account>(A, req)
}

export function updateAccount(id: string, req: AccountUpsertReq): Promise<Account> {
  return put<Account>(`${A}/${id}`, req)
}

export function deleteAccount(id: string): Promise<null> {
  return del<null>(`${A}/${id}`)
}

export function listModels(): Promise<AIModel[]> {
  return get<AIModel[]>(M)
}

export function createModel(req: ModelUpsertReq): Promise<AIModel> {
  return post<AIModel>(M, req)
}

export function updateModel(id: string, req: ModelUpsertReq): Promise<AIModel> {
  return put<AIModel>(`${M}/${id}`, req)
}

export function deleteModel(id: string): Promise<null> {
  return del<null>(`${M}/${id}`)
}

export function listGroups(): Promise<Group[]> {
  return get<Group[]>(G)
}

/** 绑定关系列表：后端只返回平台 ID 数组 */
export function listGroupPlatforms(groupId: string): Promise<string[]> {
  return get<string[]>(`${G}/${groupId}/platforms`)
}

export function bindPlatform(groupId: string, req: GroupPlatformBindReq): Promise<null> {
  return post<null>(`${G}/${groupId}/platforms`, req)
}

export function unbindPlatform(groupId: string, platformId: string): Promise<null> {
  return del<null>(`${G}/${groupId}/platforms/${platformId}`)
}
