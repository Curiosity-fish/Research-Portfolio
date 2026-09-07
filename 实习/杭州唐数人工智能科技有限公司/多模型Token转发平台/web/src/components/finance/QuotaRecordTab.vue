<script setup lang="ts">
/* 配额流水 tab：只读，服务端分页 */
import AdaptiveTable from '@/components/adaptive/AdaptiveTable.vue'
import { listQuotaRecords } from '@/api/admin/finance'
import type { PageData, QuotaRecord, QuotaRecordType } from '@/api/types'
import { shortId, useUserMap } from '@/composables/useUserMap'
import { formatCount, formatTime } from '@/utils/format'

const { userName } = useUserMap()

const TYPE_META: Record<QuotaRecordType, { label: string; tag: 'warning' | 'success' | 'danger' | 'primary' | 'info' }> = {
  pre_deduct: { label: '预扣', tag: 'warning' },
  consume: { label: '消费', tag: 'success' },
  refund: { label: '退回', tag: 'primary' },
  admin_adjust: { label: '人工调整', tag: 'danger' },
  request_approved: { label: '审批发放', tag: 'info' },
}

async function fetchRecords(page: number, pageSize: number): Promise<PageData<QuotaRecord>> {
  return listQuotaRecords({ page, page_size: pageSize })
}
</script>

<template>
  <div class="fin-tab">
    <div class="fin-tab__bar">
      <span class="fin-tab__hint">Token 配额变动全量流水；调用先预扣、完成后按实际消费结算。</span>
    </div>

    <AdaptiveTable :fetch="fetchRecords">
      <el-table-column label="用户" min-width="110" show-overflow-tooltip>
        <template #default="{ row }">{{ userName(row.user_id) }}</template>
      </el-table-column>
      <el-table-column label="Token" width="120">
        <template #default="{ row }"><span class="tnum">{{ shortId(row.token_id) }}</span></template>
      </el-table-column>
      <el-table-column label="类型" width="100">
        <template #default="{ row }">
          <el-tag :type="TYPE_META[row.type as QuotaRecordType]?.tag ?? 'info'" size="small" disable-transitions>
            {{ TYPE_META[row.type as QuotaRecordType]?.label ?? row.type }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="变动额度" min-width="100" align="right">
        <template #default="{ row }">
          <span class="tnum" :class="row.amount >= 0 ? 'fin-tab__plus' : 'fin-tab__minus'">
            {{ row.amount >= 0 ? '+' : '' }}{{ formatCount(row.amount) }}
          </span>
        </template>
      </el-table-column>
      <el-table-column label="变动后配额" min-width="110" align="right">
        <template #default="{ row }"><span class="tnum">{{ formatCount(row.quota_after) }}</span></template>
      </el-table-column>
      <el-table-column label="时间" min-width="150">
        <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
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
</style>
