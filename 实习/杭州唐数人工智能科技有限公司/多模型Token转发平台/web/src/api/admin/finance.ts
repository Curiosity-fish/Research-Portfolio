/* 计费与资金 API：订单 / 流水 / 配额审批 */
import { get, patch, post } from '../http'
import type {
  BalanceRecord,
  OrderReconciliation,
  PageData,
  PageQuery,
  QuotaRecord,
  QuotaRequest,
  RechargeOrder,
  RefundOrderResp,
  RefundReq,
} from '../types'

/** admin 订单列表仅支持 user_id 过滤（后端无 status 参数） */
export interface OrderListQuery extends PageQuery {
  user_id?: string
}

export function listOrders(query: OrderListQuery): Promise<PageData<RechargeOrder>> {
  return get<PageData<RechargeOrder>>('/admin/recharge-orders', { params: query })
}

export function getOrderReconciliation(id: string): Promise<OrderReconciliation> {
  return get<OrderReconciliation>(`/admin/recharge-orders/${id}/reconciliation`)
}

export function createRefund(orderId: string, req: RefundReq): Promise<RefundOrderResp> {
  return post<RefundOrderResp>(`/admin/recharge-orders/${orderId}/refunds`, req)
}

export interface QuotaRequestListQuery extends PageQuery {
  status?: string
}

export function listQuotaRequests(query: QuotaRequestListQuery): Promise<PageData<QuotaRequest>> {
  return get<PageData<QuotaRequest>>('/admin/quota-requests', { params: query })
}

export function approveQuotaRequest(id: string): Promise<QuotaRequest> {
  return patch<QuotaRequest>(`/admin/quota-requests/${id}/approve`)
}

export function rejectQuotaRequest(id: string): Promise<QuotaRequest> {
  return patch<QuotaRequest>(`/admin/quota-requests/${id}/reject`)
}

/** admin 余额流水仅支持 user_id 过滤（后端无 type 参数） */
export interface BalanceRecordQuery extends PageQuery {
  user_id?: string
}

export function listBalanceRecords(query: BalanceRecordQuery): Promise<PageData<BalanceRecord>> {
  return get<PageData<BalanceRecord>>('/admin/balance-records', { params: query })
}

export function listQuotaRecords(query: PageQuery): Promise<PageData<QuotaRecord>> {
  return get<PageData<QuotaRecord>>('/admin/quota-records', { params: query })
}
