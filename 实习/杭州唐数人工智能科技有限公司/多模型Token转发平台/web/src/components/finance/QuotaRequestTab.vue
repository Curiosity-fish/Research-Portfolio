<script setup lang="ts">
/* 配额审批 tab：pending 申请可一键通过/拒绝；status 过滤由服务端承担 */
import { ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import AdaptiveTable from '@/components/adaptive/AdaptiveTable.vue'
import { approveQuotaRequest, listQuotaRequests, rejectQuotaRequest } from '@/api/admin/finance'
import type { PageData, QuotaRequest, QuotaRequestStatus } from '@/api/types'
import { shortId, useUserMap } from '@/composables/useUserMap'
import { formatCount, formatTime } from '@/utils/format'

const { userName } = useUserMap()

const STATUS_META: Record<QuotaRequestStatus, { label: string; tag: 'warning' | 'success' | 'danger' }> = {
  pending: { label: '待审批', tag: 'warning' },
  approved: { label: '已通过', tag: 'success' },
  rejected: { label: '已拒绝', tag: 'danger' },
}

interface TableExpose {
  reload: (p?: number) => void
}
const tableRef = ref<TableExpose>()

const filterStatus = ref<QuotaRequestStatus | ''>('')

async function fetchRequests(page: number, pageSize: number): Promise<PageData<QuotaRequest>> {
  return listQuotaRequests({
    page,
    page_size: pageSize,
    ...(filterStatus.value ? { status: filterStatus.value } : {}),
  })
}

function applyFilter(): void {
  tableRef.value?.reload(1)
}

async function approve(row: QuotaRequest): Promise<void> {
  try {
    await ElMessageBox.confirm(
      `确认通过用户「${userName(row.user_id)}」的配额申请（${formatCount(row.requested_amount)} Token）？`,
      '通过确认',
      { type: 'warning', confirmButtonText: '通过', cancelButtonText: '取消' },
    )
  } catch {
    return
  }
  await approveQuotaRequest(row.id)
  ElMessage.success('已通过，配额已发放')
  tableRef.value?.reload()
}

async function reject(row: QuotaRequest): Promise<void> {
  try {
    await ElMessageBox.confirm(
      `确认拒绝用户「${userName(row.user_id)}」的配额申请？`,
      '拒绝确认',
      { type: 'warning', confirmButtonText: '拒绝', cancelButtonText: '取消' },
    )
  } catch {
    return
  }
  await rejectQuotaRequest(row.id)
  ElMessage.success('已拒绝')
  tableRef.value?.reload()
}
</script>

<template>
  <div class="fin-tab">
    <div class="fin-tab__bar">
      <el-select v-model="filterStatus" placeholder="全部状态" clearable class="fin-tab__filter" @change="applyFilter">
        <el-option v-for="(meta, val) in STATUS_META" :key="val" :label="meta.label" :value="val" />
      </el-select>
      <span class="fin-tab__hint">配额申请由终端用户对其 Token 发起；通过后额度直接发放到该 Token。</span>
    </div>

    <AdaptiveTable ref="tableRef" :fetch="fetchRequests">
      <el-table-column label="用户" min-width="100" show-overflow-tooltip>
        <template #default="{ row }">{{ userName(row.user_id) }}</template>
      </el-table-column>
      <el-table-column label="Token" width="120">
        <template #default="{ row }"><span class="tnum">{{ shortId(row.token_id) }}</span></template>
      </el-table-column>
      <el-table-column label="申请额度" min-width="100" align="right">
        <template #default="{ row }"><span class="tnum">{{ formatCount(row.requested_amount) }}</span></template>
      </el-table-column>
      <el-table-column label="申请原因" min-width="130" show-overflow-tooltip>
        <template #default="{ row }">{{ row.reason ?? '-' }}</template>
      </el-table-column>
      <el-table-column label="状态" width="90">
        <template #default="{ row }">
          <el-tag :type="STATUS_META[row.status as QuotaRequestStatus]?.tag ?? 'info'" size="small" disable-transitions>
            {{ STATUS_META[row.status as QuotaRequestStatus]?.label ?? row.status }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="申请时间" min-width="150">
        <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
      </el-table-column>
      <el-table-column label="审核时间" min-width="150">
        <template #default="{ row }">{{ row.reviewed_at ? formatTime(row.reviewed_at) : '-' }}</template>
      </el-table-column>
      <el-table-column label="操作" width="120">
        <template #default="{ row }">
          <template v-if="row.status === 'pending'">
            <el-button link type="primary" @click="approve(row)">通过</el-button>
            <el-button link type="danger" @click="reject(row)">拒绝</el-button>
          </template>
          <span v-else class="fin-tab__done">已处理</span>
        </template>
      </el-table-column>
    </AdaptiveTable>
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
  gap: 12px;
}
.fin-tab__filter {
  width: 140px;
  flex-shrink: 0;
}
.fin-tab__hint {
  flex: 1;
  font-size: 12px;
  color: var(--text-muted);
}
.fin-tab__done {
  font-size: 12px;
  color: var(--text-muted);
}
</style>
