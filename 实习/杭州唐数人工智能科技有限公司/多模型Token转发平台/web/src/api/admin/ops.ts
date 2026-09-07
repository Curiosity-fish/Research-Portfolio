/* 运营 API：通知 / 审计日志 / 统计 */
import { get, post } from '../http'
import type {
  AuditLog,
  DashboardStats,
  Notification,
  NotificationCreateReq,
  PageData,
  PageQuery,
  UsageQuery,
  UsageStatsResponse,
} from '../types'

export function createNotification(req: NotificationCreateReq): Promise<Notification> {
  return post<Notification>('/admin/notifications', req)
}

export interface NotificationListQuery extends PageQuery {
  user_id?: string
  type?: string
}

export function listNotifications(query?: NotificationListQuery): Promise<PageData<Notification>> {
  return get<PageData<Notification>>('/admin/notifications', { params: query })
}

export interface AuditLogQuery extends PageQuery {
  actor_type?: string
  actor_id?: string
  /** 精确匹配，形如「GET /api/v1/admin/users」 */
  action?: string
}

export function listAuditLogs(query: AuditLogQuery): Promise<PageData<AuditLog>> {
  return get<PageData<AuditLog>>('/admin/audit-logs', { params: query })
}

export interface DashboardStatsQuery {
  start_date?: string
  end_date?: string
}

export function getDashboardStats(query?: DashboardStatsQuery): Promise<DashboardStats> {
  return get<DashboardStats>('/admin/dashboard/stats', { params: query })
}

export function getUsageStats(query: UsageQuery): Promise<UsageStatsResponse> {
  return get<UsageStatsResponse>('/admin/stats/usage', { params: query })
}
