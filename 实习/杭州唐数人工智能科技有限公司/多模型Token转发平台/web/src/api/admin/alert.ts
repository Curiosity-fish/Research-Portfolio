/* 告警 API：规则 CRUD + 手动评估 + 记录查询/解决 */
import { del, get, patch, post, put } from '../http'
import type { AlertRecord, AlertRule, AlertRuleCreateReq, AlertRuleUpdateReq, PageData } from '../types'

export interface AlertRuleListQuery {
  enabled?: boolean
  page?: number
  page_size?: number
}

export function listAlertRules(query: AlertRuleListQuery): Promise<PageData<AlertRule>> {
  return get<PageData<AlertRule>>('/admin/alert-rules', { params: query })
}

export function createAlertRule(req: AlertRuleCreateReq): Promise<AlertRule> {
  return post<AlertRule>('/admin/alert-rules', req)
}

export function updateAlertRule(id: string, req: AlertRuleUpdateReq): Promise<AlertRule> {
  return put<AlertRule>(`/admin/alert-rules/${id}`, req)
}

export function deleteAlertRule(id: string): Promise<void> {
  return del<void>(`/admin/alert-rules/${id}`)
}

/** 手动评估全部启用规则，返回本次新产生（去重后）的告警数 */
export function evaluateAlerts(): Promise<{ created: number }> {
  return post<{ created: number }>('/admin/alerts/evaluate')
}

export interface AlertRecordListQuery {
  rule_id?: string
  is_resolved?: boolean
  page?: number
  page_size?: number
}

export function listAlertRecords(query: AlertRecordListQuery): Promise<PageData<AlertRecord>> {
  return get<PageData<AlertRecord>>('/admin/alerts', { params: query })
}

export function resolveAlertRecord(id: string): Promise<AlertRecord> {
  return patch<AlertRecord>(`/admin/alerts/${id}/resolve`)
}
