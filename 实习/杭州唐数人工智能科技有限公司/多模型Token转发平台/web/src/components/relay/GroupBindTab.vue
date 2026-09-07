<script setup lang="ts">
/* 分组绑定 tab：用户分组 → 可用平台的路由白名单。后端绑定关系仅返回平台 ID 数组 */
import { computed, onMounted, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import AdaptiveTable from '@/components/adaptive/AdaptiveTable.vue'
import PlatformIcon from '@/components/common/PlatformIcon.vue'
import { bindPlatform, listGroupPlatforms, listGroups, listPlatforms, unbindPlatform } from '@/api/admin/relay'
import type { Group, PageData, Platform } from '@/api/types'

interface BoundRow {
  platform_id: string
  platform: Platform
}

const groups = ref<Group[]>([])
const platforms = ref<Platform[]>([])
const platformMap = computed(() => new Map(platforms.value.map((p) => [p.id, p])))
const activeGroup = ref('')
const boundIds = ref<string[]>([])
const binding = ref(false)

const bindDialogVisible = ref(false)
const bindPlatformId = ref('')

/** 未绑定平台（绑定对话框候选） */
const unboundPlatforms = computed(() =>
  platforms.value.filter((p) => p.status === 'active' && !boundIds.value.includes(p.id)),
)

interface TableExpose {
  reload: (p?: number) => void
}
const tableRef = ref<TableExpose>()

onMounted(async () => {
  const [g, p] = await Promise.all([listGroups(), listPlatforms()])
  groups.value = g
  platforms.value = p
  if (g.length) activeGroup.value = g[0].id
})

async function fetchBound(page: number, pageSize: number): Promise<PageData<BoundRow>> {
  if (!activeGroup.value) return { total: 0, page, page_size: pageSize, list: [] }
  const ids = await listGroupPlatforms(activeGroup.value)
  boundIds.value = ids
  const rows = ids
    .map((id) => platformMap.value.get(id))
    .filter((p): p is Platform => Boolean(p))
    .map((platform) => ({ platform_id: platform.id, platform }))
  return {
    total: rows.length,
    page,
    page_size: pageSize,
    list: rows.slice((page - 1) * pageSize, page * pageSize),
  }
}

/* 分组可能由 onMounted 异步赋值（非用户交互），统一由 watch 触发刷新 */
watch(activeGroup, () => {
  tableRef.value?.reload(1)
})

function openBind(): void {
  bindPlatformId.value = ''
  bindDialogVisible.value = true
}

async function submitBind(): Promise<void> {
  if (!activeGroup.value || !bindPlatformId.value) return
  binding.value = true
  try {
    await bindPlatform(activeGroup.value, { platform_id: bindPlatformId.value })
    ElMessage.success('已绑定')
    bindDialogVisible.value = false
    tableRef.value?.reload()
  } finally {
    binding.value = false
  }
}

async function removeBind(row: BoundRow): Promise<void> {
  const groupName = groups.value.find((g) => g.id === activeGroup.value)?.name ?? ''
  try {
    await ElMessageBox.confirm(
      `确认将平台「${row.platform.name}」从分组「${groupName}」解绑？`,
      '解绑确认',
      { type: 'warning', confirmButtonText: '解绑', cancelButtonText: '取消' },
    )
  } catch {
    return
  }
  await unbindPlatform(activeGroup.value, row.platform_id)
  ElMessage.success('已解绑')
  tableRef.value?.reload()
}
</script>

<template>
  <div class="relay-tab">
    <div class="relay-tab__bar">
      <el-select v-model="activeGroup" class="relay-tab__group">
        <el-option v-for="g in groups" :key="g.id" :label="g.name" :value="g.id" />
      </el-select>
      <span class="relay-tab__hint">终端用户按所属分组路由到已绑定的平台；未绑定任何平台时分组不可用。</span>
      <el-button type="primary" :icon="'Plus'" :disabled="!activeGroup" @click="openBind">绑定平台</el-button>
    </div>

    <AdaptiveTable ref="tableRef" :fetch="fetchBound" :min-rows="5">
      <el-table-column label="平台" min-width="170">
        <template #default="{ row }">
          <span class="relay-tab__platform">
            <PlatformIcon :code="row.platform.code" :size="18" />
            <span class="relay-tab__name">{{ row.platform.name }}</span>
          </span>
        </template>
      </el-table-column>
      <el-table-column prop="platform.code" label="标识码" min-width="110" show-overflow-tooltip />
      <el-table-column prop="platform.base_url" label="Base URL" min-width="220" show-overflow-tooltip />
      <el-table-column label="状态" width="90">
        <template #default="{ row }">
          <el-tag :type="row.platform.status === 'active' ? 'success' : 'info'" size="small" disable-transitions>
            {{ row.platform.status === 'active' ? '正常' : '停用' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="90">
        <template #default="{ row }">
          <el-button link type="danger" @click="removeBind(row)">解绑</el-button>
        </template>
      </el-table-column>
    </AdaptiveTable>

    <el-dialog v-model="bindDialogVisible" title="绑定平台" width="440px" destroy-on-close>
      <el-select v-model="bindPlatformId" placeholder="选择要绑定的平台" class="relay-tab__full">
        <el-option v-for="p in unboundPlatforms" :key="p.id" :label="p.name" :value="p.id" />
      </el-select>
      <template #footer>
        <el-button @click="bindDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="binding" :disabled="!bindPlatformId" @click="submitBind">
          绑定
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.relay-tab {
  height: 100%;
  min-height: 0;
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.relay-tab__bar {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  gap: 12px;
}
.relay-tab__group {
  width: 220px;
  flex-shrink: 0;
}
.relay-tab__hint {
  flex: 1;
  font-size: 12px;
  color: var(--text-muted);
}
.relay-tab__platform {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}
.relay-tab__name {
  font-weight: 600;
}
.relay-tab__full {
  width: 100%;
}
</style>
