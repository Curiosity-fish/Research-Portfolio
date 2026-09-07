<script setup lang="ts">
/* 告警记录：列表（规则/状态过滤）+ 手动评估 + 标记解决。
   后端评估自动按规则去重，同一问题未解决前重复评估不会产生新记录 */
import { onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import AdaptiveTable from '@/components/adaptive/AdaptiveTable.vue'
import { evaluateAlerts, listAlertRecords, listAlertRules, resolveAlertRecord } from '@/api/admin/alert'
import type { AlertMetric, AlertRecord, AlertRule, PageData } from '@/api/types'
import { shortId, useUserMap } from '@/composables/useUserMap'
import { formatCount, formatTime } from '@/utils/format'

const METRIC_META: Record<AlertMetric, { label: string; tag: 'warning' | 'danger' }> = {
  balance_low: { label: '余额过低', tag: 'warning' },
  quota_low: { label: '配额过低', tag: 'warning' },
  error_rate: { label: '错误次数', tag: 'danger' },
  cost_spike: { label: '消费突增', tag: 'danger' },
}

const { userName } = useUserMap()

interface TableExpose {
  reload: (p?: number) => void
}
const tableRef = ref<TableExpose>()

/* ── 过滤条件 ── */
const filterRule = ref('')
const filterResolved = ref<'' | 'unresolved' | 'resolved'>('')

/* 规则下拉与 rule_id → 名称映射共用一份数据 */
const rules = ref<AlertRule[]>([])
onMounted(async () => {
  const data = await listAlertRules({ page: 1, page_size: 100 })
  rules.value = data.list
})

function ruleName(id?: string): string {
  if (!id) return '—'
  return rules.value.find((r) => r.id === id)?.name ?? shortId(id)
}

/* 关联对象：优先用户，其次上游账号，再次 token */
function targetText(row: AlertRecord): string {
  if (row.user_id) return userName(row.user_id)
  if (row.account_id) return `账号 ${shortId(row.account_id)}`
  if (row.token_id) return `Token ${shortId(row.token_id)}`
  return '—'
}

async function fetchRecords(page: number, pageSize: number): Promise<PageData<AlertRecord>> {
  return listAlertRecords({
    page,
    page_size: pageSize,
    ...(filterRule.value ? { rule_id: filterRule.value } : {}),
    ...(filterResolved.value ? { is_resolved: filterResolved.value === 'resolved' } : {}),
  })
}

function applyFilter(): void {
  tableRef.value?.reload(1)
}

const evaluating = ref(false)
async function evaluate(): Promise<void> {
  evaluating.value = true
  try {
    const { created } = await evaluateAlerts()
    ElMessage.success(`评估完成，本次新产生 ${created} 条告警`)
    tableRef.value?.reload(1)
  } finally {
    evaluating.value = false
  }
}

async function resolve(row: AlertRecord): Promise<void> {
  await ElMessageBox.confirm('确认将该告警标记为已解决？', '标记解决', {
    type: 'info',
    confirmButtonText: '标记解决',
    cancelButtonText: '取消',
  })
  await resolveAlertRecord(row.id)
  ElMessage.success('已标记解决')
  tableRef.value?.reload()
}
</script>

<template>
  <div class="record-tab">
    <div class="record-tab__toolbar app-card">
      <el-select v-model="filterRule" placeholder="全部规则" clearable class="record-tab__filter" @change="applyFilter">
        <el-option v-for="r in rules" :key="r.id" :label="r.name" :value="r.id" />
      </el-select>
      <el-select v-model="filterResolved" placeholder="全部状态" clearable class="record-tab__filter" @change="applyFilter">
        <el-option label="未处理" value="unresolved" />
        <el-option label="已解决" value="resolved" />
      </el-select>
      <span class="record-tab__hint">告警由规则评估产生；同一问题未解决前重复评估会自动去重。</span>
      <el-button type="primary" plain :icon="'Refresh'" :loading="evaluating" @click="evaluate">评估一次</el-button>
    </div>

    <AdaptiveTable ref="tableRef" :fetch="fetchRecords" class="record-tab__table">
      <el-table-column label="规则" min-width="130" show-overflow-tooltip>
        <template #default="{ row }">{{ ruleName((row as AlertRecord).rule_id) }}</template>
      </el-table-column>
      <el-table-column label="指标" width="100">
        <template #default="{ row }">
          <el-tag :type="METRIC_META[row.metric as AlertMetric]?.tag ?? 'info'" size="small" disable-transitions>
            {{ METRIC_META[row.metric as AlertMetric]?.label ?? row.metric }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="message" label="内容" min-width="220" show-overflow-tooltip />
      <el-table-column label="关联对象" min-width="150" show-overflow-tooltip>
        <template #default="{ row }">{{ targetText(row as AlertRecord) }}</template>
      </el-table-column>
      <el-table-column label="触发值" width="100" align="right">
        <template #default="{ row }">{{ formatCount((row as AlertRecord).triggered_value) }}</template>
      </el-table-column>
      <el-table-column label="状态" width="90">
        <template #default="{ row }">
          <el-tag :type="(row as AlertRecord).is_resolved ? 'info' : 'danger'" size="small" disable-transitions>
            {{ (row as AlertRecord).is_resolved ? '已解决' : '未处理' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="触发时间" min-width="150">
        <template #default="{ row }">{{ formatTime((row as AlertRecord).created_at) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="110" fixed="right">
        <template #default="{ row }">
          <el-button
            v-if="!(row as AlertRecord).is_resolved"
            link
            type="primary"
            @click="resolve(row as AlertRecord)"
          >
            标记解决
          </el-button>
          <span v-else class="record-tab__done">—</span>
        </template>
      </el-table-column>
    </AdaptiveTable>
  </div>
</template>

<style scoped>
.record-tab {
  height: 100%;
  min-height: 0;
  display: flex;
  flex-direction: column;
}
.record-tab__toolbar {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
}
.record-tab__filter {
  width: 150px;
  flex-shrink: 0;
}
.record-tab__hint {
  flex: 1;
  font-size: 12px;
  color: var(--text-muted);
}
.record-tab__table {
  min-height: 0;
}
.record-tab__done {
  color: var(--text-muted);
}
</style>
