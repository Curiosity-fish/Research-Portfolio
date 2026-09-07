<script setup lang="ts">
/* 审计日志：admin 全量请求审计（AuditMiddleware 写入），支持操作者类型与动作精确过滤 */
import { reactive, ref } from 'vue'
import AdaptiveTable from '@/components/adaptive/AdaptiveTable.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import { listAuditLogs } from '@/api/admin/ops'
import type { AuditActorType, AuditLog, PageData } from '@/api/types'
import { formatTime } from '@/utils/format'

const ACTOR_META: Record<AuditActorType, { label: string; tag: 'primary' | 'success' | 'info' }> = {
  admin: { label: '管理员', tag: 'primary' },
  user: { label: '用户', tag: 'success' },
  system: { label: '系统', tag: 'info' },
}

interface TableExpose {
  reload: (p?: number) => void
}
const tableRef = ref<TableExpose>()

const filters = reactive<{ actor_type: AuditActorType | ''; action: string }>({ actor_type: '', action: '' })

async function fetchLogs(page: number, pageSize: number): Promise<PageData<AuditLog>> {
  return listAuditLogs({
    page,
    page_size: pageSize,
    ...(filters.actor_type ? { actor_type: filters.actor_type } : {}),
    ...(filters.action.trim() ? { action: filters.action.trim() } : {}),
  })
}

function applyFilters(): void {
  tableRef.value?.reload(1)
}

/** details 恒为对象，压成单行 JSON 供悬浮查看 */
function detailsText(log: AuditLog): string {
  return JSON.stringify(log.details)
}
</script>

<template>
  <div class="page-root">
    <PageHeader title="审计日志" description="管理员操作全量审计：动作形如「METHOD 路径」，支持精确匹配过滤" />

    <div class="audit__toolbar app-card">
      <el-select v-model="filters.actor_type" placeholder="操作者类型" clearable class="audit__filter" @change="applyFilters">
        <el-option v-for="(meta, val) in ACTOR_META" :key="val" :label="meta.label" :value="val" />
      </el-select>
      <el-input v-model="filters.action" placeholder="动作精确匹配，如 GET /api/v1/admin/users" clearable class="audit__action"
        @keyup.enter="applyFilters" @clear="applyFilters" />
      <el-button type="primary" @click="applyFilters">查询</el-button>
    </div>

    <AdaptiveTable ref="tableRef" :fetch="fetchLogs" class="audit__table">
      <el-table-column label="时间" min-width="150">
        <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
      </el-table-column>
      <el-table-column label="操作者类型" width="100">
        <template #default="{ row }">
          <el-tag :type="ACTOR_META[row.actor_type as AuditActorType]?.tag ?? 'info'" size="small" disable-transitions>
            {{ ACTOR_META[row.actor_type as AuditActorType]?.label ?? row.actor_type }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作者 ID" width="130">
        <template #default="{ row }"><span class="tnum">{{ row.actor_id.slice(0, 8) }}…</span></template>
      </el-table-column>
      <el-table-column prop="action" label="动作" min-width="240" show-overflow-tooltip />
      <el-table-column label="详情" min-width="180" show-overflow-tooltip>
        <template #default="{ row }">{{ detailsText(row) }}</template>
      </el-table-column>
      <el-table-column label="IP" width="130">
        <template #default="{ row }">{{ row.ip ?? '-' }}</template>
      </el-table-column>
    </AdaptiveTable>
  </div>
</template>

<style scoped>
.audit__toolbar {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
}
.audit__filter {
  width: 140px;
  flex-shrink: 0;
}
.audit__action {
  flex: 1;
  max-width: 480px;
}
.audit__table {
  min-height: 0;
}
</style>
