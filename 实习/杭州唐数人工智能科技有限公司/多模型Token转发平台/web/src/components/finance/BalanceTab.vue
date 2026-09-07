<script setup lang="ts">
/* 余额流水 tab：服务端分页；支持人工调账（正数加余额、负数扣减） */
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import AdaptiveTable from '@/components/adaptive/AdaptiveTable.vue'
import { listBalanceRecords } from '@/api/admin/finance'
import { adjustBalance, listUsers } from '@/api/admin/user'
import type { BalanceRecord, PageData, User } from '@/api/types'
import { useUserMap } from '@/composables/useUserMap'
import { formatMicro, formatTime } from '@/utils/format'

const { userName } = useUserMap()

const TYPE_META: Record<BalanceRecord['type'], { label: string; tag: 'success' | 'warning' | 'danger' | 'primary' }> = {
  recharge: { label: '充值', tag: 'success' },
  consume: { label: '消费', tag: 'warning' },
  refund: { label: '退款', tag: 'primary' },
  admin_adjust: { label: '人工调账', tag: 'danger' },
}

interface TableExpose {
  reload: (p?: number) => void
}
const tableRef = ref<TableExpose>()

async function fetchRecords(page: number, pageSize: number): Promise<PageData<BalanceRecord>> {
  return listBalanceRecords({ page, page_size: pageSize })
}

/* ── 人工调账 ── */
const users = ref<User[]>([])
onMounted(async () => {
  const data = await listUsers({ page: 1, page_size: 100 })
  users.value = data.list
})

const adjustVisible = ref(false)
const adjustSaving = ref(false)
const adjustFormRef = ref<FormInstance>()
const adjustForm = reactive({ user_id: '', amount_yuan: 1, direction: 'add' as 'add' | 'deduct', remark: '' })

const adjustRules: FormRules = {
  user_id: [{ required: true, message: '请选择用户', trigger: 'change' }],
  amount_yuan: [
    { required: true, message: '请输入金额', trigger: 'blur' },
    { validator: (_r, v: number, cb) => (v > 0 ? cb() : cb(new Error('金额需大于 0'))), trigger: 'blur' },
  ],
}

function openAdjust(): void {
  adjustForm.user_id = ''
  adjustForm.amount_yuan = 1
  adjustForm.direction = 'add'
  adjustForm.remark = ''
  adjustVisible.value = true
}

async function submitAdjust(): Promise<void> {
  if (!adjustFormRef.value) return
  await adjustFormRef.value.validate(async (valid) => {
    if (!valid) return
    adjustSaving.value = true
    try {
      const micro = Math.round(adjustForm.amount_yuan * 1e6)
      await adjustBalance(adjustForm.user_id, {
        amount: adjustForm.direction === 'add' ? micro : -micro,
        remark: adjustForm.remark.trim() || undefined,
      })
      ElMessage.success('调账成功')
      adjustVisible.value = false
      tableRef.value?.reload()
    } finally {
      adjustSaving.value = false
    }
  })
}
</script>

<template>
  <div class="fin-tab">
    <div class="fin-tab__bar">
      <span class="fin-tab__hint">余额变动全量流水：充值/消费/退款/人工调账；金额为 micro 存储，展示按元。</span>
      <el-button type="primary" :icon="'Plus'" @click="openAdjust">人工调账</el-button>
    </div>

    <AdaptiveTable ref="tableRef" :fetch="fetchRecords">
      <el-table-column label="用户" min-width="110" show-overflow-tooltip>
        <template #default="{ row }">{{ userName(row.user_id) }}</template>
      </el-table-column>
      <el-table-column label="类型" width="100">
        <template #default="{ row }">
          <el-tag :type="TYPE_META[row.type as BalanceRecord['type']]?.tag ?? 'info'" size="small" disable-transitions>
            {{ TYPE_META[row.type as BalanceRecord['type']]?.label ?? row.type }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="变动金额（元）" min-width="110" align="right">
        <template #default="{ row }">
          <span class="tnum" :class="row.amount >= 0 ? 'fin-tab__plus' : 'fin-tab__minus'">
            {{ row.amount >= 0 ? '+' : '' }}{{ formatMicro(row.amount) }}
          </span>
        </template>
      </el-table-column>
      <el-table-column label="变动后余额（元）" min-width="120" align="right">
        <template #default="{ row }"><span class="tnum">{{ formatMicro(row.balance_after) }}</span></template>
      </el-table-column>
      <el-table-column label="备注" min-width="140" show-overflow-tooltip>
        <template #default="{ row }">{{ row.remark ?? '-' }}</template>
      </el-table-column>
      <el-table-column label="时间" min-width="150">
        <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
      </el-table-column>
    </AdaptiveTable>

    <el-dialog v-model="adjustVisible" title="人工调账" width="460px" destroy-on-close>
      <el-form ref="adjustFormRef" :model="adjustForm" :rules="adjustRules" label-width="92px">
        <el-form-item label="用户" prop="user_id">
          <el-select v-model="adjustForm.user_id" filterable placeholder="选择用户" class="fin-tab__full">
            <el-option v-for="u in users" :key="u.id" :label="u.username" :value="u.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="方向" prop="direction">
          <el-radio-group v-model="adjustForm.direction">
            <el-radio-button value="add">增加余额</el-radio-button>
            <el-radio-button value="deduct">扣减余额</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="金额（元）" prop="amount_yuan">
          <el-input-number v-model="adjustForm.amount_yuan" :min="0.01" :precision="2" :step="1"
            controls-position="right" class="fin-tab__full" />
        </el-form-item>
        <el-form-item label="备注" prop="remark">
          <el-input v-model="adjustForm.remark" type="textarea" :rows="2" maxlength="500" placeholder="选填" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="adjustVisible = false">取消</el-button>
        <el-button type="primary" :loading="adjustSaving" @click="submitAdjust">确认调账</el-button>
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
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}
.fin-tab__hint {
  font-size: 12px;
  color: var(--text-muted);
}
.fin-tab__plus {
  color: var(--success, #529b2e);
}
.fin-tab__minus {
  color: var(--danger, #c45656);
}
.fin-tab__full {
  width: 100%;
}
</style>
