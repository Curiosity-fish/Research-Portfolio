/* API 契约类型：字段与 learning/frontend-handoff.md 逐项对应 */

/** 成功响应信封 */
export interface Envelope<T> {
  code: number
  message: string
  data: T
}

/** 失败响应体 */
export interface ApiErrorBody {
  error: {
    code: string
    message: string
    request_id: string
    details?: Record<string, string>
  }
}

/** 统一分页结构 */
export interface PageData<T> {
  total: number
  page: number
  page_size: number
  list: T[]
}

/** 统一分页查询参数 */
export interface PageQuery {
  page?: number
  page_size?: number
}

// ---------- 认证 ----------

export interface LoginReq {
  username: string
  password: string
}

export interface LoginResp {
  access_token: string
  token_type: string
  expires_in: number
}

export interface AdminMe {
  admin_id: string
  role: string
  status: string
}

// ---------- 用户与 Token ----------

export type UserRole = 'student' | 'teacher' | 'staff'
export type UserStatus = 'active' | 'inactive' | 'banned'

export interface User {
  id: string
  username: string
  name: string
  role: UserRole
  status: UserStatus
  email?: string
  phone?: string
  gender?: string
  department_id?: string
  group_id?: string
  created_at: string
  updated_at: string
}

export interface UserCreateReq {
  username: string
  password: string
  name: string
  role?: UserRole
  status?: UserStatus
  email?: string
  phone?: string
  gender?: string
  department_id?: string
  group_id?: string
}

export interface UserUpdateReq {
  name?: string
  password?: string
  role?: UserRole
  email?: string
  phone?: string
  gender?: string
  department_id?: string
  group_id?: string
}

export interface UserStatusReq {
  status: UserStatus
}

export interface UserToken {
  id: string
  user_id: string
  name: string
  token_last4: string
  is_enabled: boolean
  /** 后端 omitempty：未设置时字段缺失，表示不限额度 */
  quota_limit?: number
  quota_used: number
  expires_at?: string
  created_at: string
  updated_at: string
}

export interface TokenCreateReq {
  name: string
  /** 可选，≥0，缺省不限额度 */
  quota_limit?: number
  /** 可选，RFC3339，缺省永不过期 */
  expires_at?: string
}

/** 创建 Token 的响应：plaintext_token 仅返回一次 */
export interface TokenCreateResp extends UserToken {
  plaintext_token: string
}

// ---------- 转发资源 ----------

export type PlatformType = 'openai' | 'anthropic'
export type PlatformStatus = 'active' | 'inactive'

/** 平台列表不分页：后端一次性返回全部 */
export interface Platform {
  id: string
  name: string
  code: string
  type: PlatformType
  base_url: string
  status: PlatformStatus
  created_at: string
  updated_at: string
}

export interface PlatformUpsertReq {
  name: string
  code: string
  type?: PlatformType
  base_url: string
  status?: PlatformStatus
}

export type AccountStatus = 'active' | 'inactive' | 'error'

/** 账号列表不分页；api_key 永不回传，仅有是否已加密的标记 */
export interface Account {
  id: string
  platform_id: string
  name: string
  api_key_encrypted: boolean
  weight: number
  max_rpm: number
  status: AccountStatus
  error_count: number
  created_at: string
  updated_at: string
}

export interface AccountUpsertReq {
  platform_id: string
  name: string
  /** 编辑时留空表示不修改 */
  api_key: string
  weight?: number
  max_rpm?: number
  status?: AccountStatus
}

export type ModelType = 'chat' | 'embedding' | 'image'

/** 模型列表不分页；价格为 micro-元/1K Token（后端结算公式 tokens×price/1000） */
export interface AIModel {
  id: string
  name: string
  upstream_name: string
  type: ModelType
  input_price: number
  output_price: number
  is_enabled: boolean
  created_at: string
  updated_at: string
}

export interface ModelUpsertReq {
  name: string
  upstream_name: string
  type?: ModelType
  input_price?: number
  output_price?: number
  is_enabled?: boolean
}

/** 分组（relay 路由层实体），GET /admin/groups 只读列表 */
export interface Group {
  id: string
  name: string
  code: string
  description?: string
  sort_order: number
  status: PlatformStatus
  created_at: string
  updated_at: string
}

/** 分组-平台绑定请求（绑定本身无权重等附加字段） */
export interface GroupPlatformBindReq {
  platform_id: string
}

// ---------- 计费与资金 ----------

export type OrderStatus = 'pending' | 'paid' | 'failed' | 'cancelled'

/** 充值订单。金额 micro（1 元 = 1e6）；admin 列表仅支持 user_id 过滤 */
export interface RechargeOrder {
  id: string
  user_id: string
  amount: number
  status: OrderStatus
  provider: string
  provider_order_id?: string
  paid_at?: string
  created_at: string
  updated_at: string
}

/** 订单对账单（含已退/可退金额与关联余额流水） */
export interface OrderReconciliation {
  id: string
  user_id: string
  amount: number
  refunded_amount: number
  refundable_amount: number
  status: string
  provider: string
  created_at: string
  updated_at: string
  balance_records: BalanceRecord[]
}

export interface RefundReq {
  amount: number
  reason?: string
}

/** 退款结果：最新对账单 + 本次退款产生的余额流水 */
export interface RefundOrderResp {
  order: OrderReconciliation
  balance_record: BalanceRecord
}

export type BalanceRecordType = 'recharge' | 'consume' | 'refund' | 'admin_adjust'

/** 余额流水。金额 micro；balance_after 为变动后余额 */
export interface BalanceRecord {
  id: string
  user_id: string
  type: BalanceRecordType
  amount: number
  balance_after: number
  call_log_id?: string
  related_order_id?: string
  remark?: string
  created_at: string
  updated_at: string
}

export type QuotaRequestStatus = 'pending' | 'approved' | 'rejected'

/** 配额申请。requested_amount 单位为 Token 数；审批通过后产生 request_approved 配额流水 */
export interface QuotaRequest {
  id: string
  user_id: string
  token_id: string
  requested_amount: number
  reason?: string
  status: QuotaRequestStatus
  reviewed_by?: string
  reviewed_at?: string
  created_at: string
  updated_at: string
}

export type QuotaRecordType = 'pre_deduct' | 'consume' | 'refund' | 'admin_adjust' | 'request_approved'

/** 配额流水。amount 单位为 Token 数；quota_after 为变动后配额 */
export interface QuotaRecord {
  id: string
  user_id: string
  token_id: string
  call_log_id?: string
  type: QuotaRecordType
  amount: number
  quota_after: number
  created_at: string
  updated_at: string
}

/** 人工调账。金额 micro，正数加余额、负数扣减 */
export interface BalanceAdjustReq {
  amount: number
  remark?: string
}

// ---------- 运营 ----------

export type NotificationType = 'announcement' | 'system'

/** 通知。user_id 为空表示广播；is_read 为当前查看者维度，admin 列表恒 false */
export interface Notification {
  id: string
  type: NotificationType
  title: string
  content: string
  user_id?: string
  is_read: boolean
  created_at: string
  updated_at: string
}

export interface NotificationCreateReq {
  type: NotificationType
  title: string
  content: string
  /** 缺省为广播 */
  user_id?: string
}

export type AuditActorType = 'admin' | 'user' | 'system'

/** 审计日志。action 为「METHOD 路径」；details 恒为对象（path/query/method/status） */
export interface AuditLog {
  id: string
  actor_type: AuditActorType
  actor_id: string
  action: string
  target_type?: string
  target_id?: string
  details: Record<string, unknown>
  ip?: string
  user_agent?: string
  created_at: string
  updated_at: string
}

// ---------- 统计 ----------

export interface DashboardStats {
  calls: number
  prompt_tokens: number
  completion_tokens: number
  total_tokens: number
  recharge_amount: number
  consume_amount: number
}

/** stats/usage 查询：后端不分页（忽略 page/page_size），仅支持日期过滤 */
export interface UsageQuery {
  group_by: 'model' | 'user' | 'day'
  start_date?: string
  end_date?: string
}

/** stats/usage 响应：固定为全量 list，无 total/page/page_size */
export interface UsageStatsResponse {
  list: UsageStat[]
}

/* ── 告警 ── */

export type AlertMetric = 'balance_low' | 'quota_low' | 'error_rate' | 'cost_spike'

export interface AlertRule {
  id: string
  name: string
  metric: AlertMetric
  /** 阈值单位随指标而定：余额/消费为 micro 元，配额为 token 数，错误次数为计数 */
  threshold: number
  enabled: boolean
  description?: string
  created_at: string
  updated_at: string
}

export interface AlertRuleCreateReq {
  name: string
  metric: AlertMetric
  threshold: number
  enabled: boolean
  description?: string
}

/** 更新为指针语义：不传表示不修改 */
export interface AlertRuleUpdateReq {
  name?: string
  metric?: AlertMetric
  threshold?: number
  enabled?: boolean
  description?: string
}

export interface AlertRecord {
  id: string
  rule_id?: string
  metric: AlertMetric
  user_id?: string
  token_id?: string
  account_id?: string
  triggered_value: number
  message: string
  is_resolved: boolean
  created_at: string
  updated_at: string
}

export interface UsageStat {
  model?: string
  user_id?: string
  day?: string
  calls: number
  prompt_tokens?: number
  completion_tokens?: number
  total_tokens: number
}
