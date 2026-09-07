<script setup lang="ts">
/* 充值订单 tab：服务端分页；支持对账（查看可退金额与关联流水）与退款 */
import { reactive, ref } from 'vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import AdaptiveTable from '@/components/adaptive/AdaptiveTable.vue'
import { createRefund, getOrderReconciliation, listOrders } from '@/api/admin/finance'
import type { OrderReconciliation, PageData, RechargeOrder } from '@/api/types'
import { useUserMap } from '@/composables/useUserMap'
import { formatMicro, formatTime } from '@/utils/format'

const { userName } = useUserMap()

const STATUS_META: Record<RechargeOrder['status'], { label: string; tag: 'warning' | 'success' | 'danger' | 'info' }> = {
  pending: { label: '待支付', tag: 'warning' },
  paid: { label: '已支付', tag: 'success' },
  failed: { label: '失败', tag: 'danger' },
  cancelled: { label: '已取消', tag: 'info' },
}

interface TableExpose {
  reload: (p?: number) => void
}
const tableRef = ref<TableExpose>()

async function fetchOrders(page: number, pageSize: number): Promise<PageData<RechargeOrder>> {
  return listOrders({ page, page_size: pageSize })
}

/* ── 对账 ── */
const reconVisible = ref(false)
const reconLoading = ref(false)
const recon = ref<OrderReconciliation | null>(null)

async function openRecon(row: RechargeOrder): Promise<void> {
  reconVisible.value = true
  reconLoading.value = true
  try {
    recon.value = await getOrderReconciliation(row.id)
  } finally {
    reconLoading.value = false
  }
}

/* ── 退款 ── */
const refundVisible = ref(false)
const refundSaving = ref(false)
const refundOrderId = ref('')
const refundMaxYuan = ref(0)
const refundFormRef = ref<FormInstance>()
const refundForm = reactive({ amount_yuan: 0, reason: '' })

const refundRules: FormRules = {
  amount_yuan: [
    { required: true, message: '请输入退款金额', trigger: 'blur' },
    {
      validator: (_r, v: number, cb) => {
        if (v <= 0) cb(new Error('金额需大于 0'))
        else if (v > refundMaxYuan.value) cb(new Error(`超出可退金额 ${refundMaxYuan.value.toFixed(2)} 元`))
        else cb()
      },
      trigger: 'blur',
    },
  ],
}

async function openRefund(row: RechargeOrder): Promise<void> {
  try {
    const r = await getOrderReconciliation(row.id)
    if (r.refundable_amount <= 0) {
      ElMessage.warning('该订单无可退金额')
      return
    }
    refundOrderId.value = row.id
    refundMaxYuan.value = r.refundable_amount / 1e6
    refundForm.amount_yuan = refundMaxYuan.value
    refundForm.reason = ''
    refundVisible.value = true
  } catch {
    // 拦截器已提示
  }
}

async function submitRefund(): Promise<void> {
  if (!refundFormRef.value) return
  await refundFormRef.value.validate(async (valid) => {
    if (!valid) return
    refundSaving.value = true
    try {
      await createRefund(refundOrderId.value, {
        amount: Math.round(refundForm.amount_yuan * 1e6),
        reason: refundForm.reason.trim() || undefined,
      })
      ElMessage.success('退款成功')
      refundVisible.value = false
      tableRef.value?.reload()
    } finally {
      refundSaving.value = false
    }
  })
}
</script>

<template>
  <div class="fin-tab">
    <div class="fin-tab__bar">
      <span class="fin-tab__hint">充值订单由终端用户发起；支付走 mock 渠道，退款按可退金额部分或全额退回。</span>
    </div>

    <AdaptiveTable ref="tableRef" :fetch="fetchOrders">
      <el-table-column label="订单号" width="120">
        <template #default="{ row }"><span class="tnum">{{ row.id.slice(0, 8) }}…</span></template>
      </el-table-column>
      <el-table-column label="用户" min-width="110" show-overflow-tooltip>
        <template #default="{ row }">{{ userName(row.user_id) }}</template>
      </el-table-column>
      <el-table-column label="金额（元）" min-width="100" align="right">
        <template #default="{ row }"><span class="tnum">{{ formatMicro(row.amount) }}</span></template>
      </el-table-column>
      <el-table-column label="状态" width="90">
        <template #default="{ row }">
          <el-tag :type="STATUS_META[row.status as RechargeOrder['status']]?.tag ?? 'info'" size="small" disable-transitions>
            {{ STATUS_META[row.status as RechargeOrder['status']]?.label ?? row.status }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="provider" label="渠道" width="80" />
      <el-table-column label="支付时间" min-width="150">
        <template #default="{ row }">{{ row.paid_at ? formatTime(row.paid_at) : '-' }}</template>
      </el-table-column>
      <el-table-column label="创建时间" min-width="150">
        <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="120">
        <template #default="{ row }">
          <el-button link type="primary" @click="openRecon(row)">对账</el-button>
          <el-button v-if="row.status === 'paid'" link type="danger" @click="openRefund(row)">退款</el-button>
        </template>
      </el-table-column>
    </AdaptiveTable>

    <el-dialog v-model="reconVisible" title="订单对账" width="640px">
      <div v-loading="reconLoading">
        <template v-if="recon">
          <div class="fin-recon__summary">
            <div class="fin-recon__item">
              <span class="fin-recon__label">订单金额</span>
              <span class="tnum">{{ formatMicro(recon.amount) }} 元</span>
            </div>
            <div class="fin-recon__item">
              <span class="fin-recon__label">已退金额</span>
              <span class="tnum">{{ formatMicro(recon.refunded_amount) }} 元</span>
            </div>
            <div class="fin-recon__item">
              <span class="fin-recon__label">可退金额</span>
              <span class="tnum fin-recon__strong">{{ formatMicro(recon.refundable_amount) }} 元</span>
            </div>
          </div>
          <el-table :data="recon.balance_records" size="small" class="fin-recon__table">
            <el-table-column label="类型" width="90">
              <template #default="{ row }">
                <el-tag size="small" disable-transitions :type="row.type === 'refund' ? 'warning' : 'success'">
                  {{ row.type === 'refund' ? '退款' : '充值' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="金额（元）" min-width="100" align="right">
              <template #default="{ row }"><span class="tnum">{{ formatMicro(row.amount) }}</span></template>
            </el-table-column>
            <el-table-column label="变动后余额（元）" min-width="120" align="right">
              <template #default="{ row }"><span class="tnum">{{ formatMicro(row.balance_after) }}</span></template>
            </el-table-column>
            <el-table-column label="备注" min-width="120" show-overflow-tooltip>
              <template #default="{ row }">{{ row.remark ?? '-' }}</template>
            </el-table-column>
            <el-table-column label="时间" min-width="150">
              <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
            </el-table-column>
          </el-table>
        </template>
      </div>
    </el-dialog>

    <el-dialog v-model="refundVisible" title="订单退款" width="460px" destroy-on-close>
      <el-alert type="warning" :closable="false" class="fin-tab__alert"
        :title="`最多可退 ${refundMaxYuan.toFixed(2)} 元，退款将实时冲减用户余额。`" />
      <el-form ref="refundFormRef" :model="refundForm" :rules="refundRules" label-width="92px" class="fin-tab__form">
        <el-form-item label="退款金额" prop="amount_yuan">
          <el-input-number v-model="refundForm.amount_yuan" :min="0.01" :max="refundMaxYuan" :precision="2" :step="1"
            controls-position="right" class="fin-tab__full" />
        </el-form-item>
        <el-form-item label="退款原因" prop="reason">
          <el-input v-model="refundForm.reason" type="textarea" :rows="2" maxlength="500" placeholder="选填" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="refundVisible = false">取消</el-button>
        <el-button type="danger" :loading="refundSaving" @click="submitRefund">确认退款</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.fin-tab {
  height: 100%;
  min-height: 0;
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.fin-tab__bar {
  flex-shrink: 0;
}
.fin-tab__hint {
  font-size: 12px;
  color: var(--text-muted);
}
.fin-tab__alert {
  margin-bottom: 16px;
}
.fin-tab__form {
  margin-top: 4px;
}
.fin-tab__full {
  width: 100%;
}
.fin-recon__summary {
  display: flex;
  gap: 24px;
  margin-bottom: 16px;
  padding: 12px 16px;
  background: var(--bg-page);
  border-radius: var(--radius-md);
}
.fin-recon__item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.fin-recon__label {
  font-size: 12px;
  color: var(--text-muted);
}
.fin-recon__strong {
  font-weight: 700;
  color: var(--brand);
}
.fin-recon__table {
  width: 100%;
}
</style>
