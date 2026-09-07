<script setup lang="ts">
/* 平台 tab：列表 + 新建/编辑/删除。列表不分页（后端一次返回），客户端切片 */
import { reactive, ref } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import AdaptiveTable from '@/components/adaptive/AdaptiveTable.vue'
import PlatformIcon from '@/components/common/PlatformIcon.vue'
import { createPlatform, deletePlatform, listPlatforms, updatePlatform } from '@/api/admin/relay'
import type { PageData, Platform, PlatformStatus, PlatformType } from '@/api/types'
import { formatTime } from '@/utils/format'

const TYPE_LABELS: Record<PlatformType, string> = {
  openai: 'OpenAI 兼容',
  anthropic: 'Anthropic 兼容',
}

const STATUS_META: Record<string, { label: string; tag: 'success' | 'info' }> = {
  active: { label: '正常', tag: 'success' },
  inactive: { label: '停用', tag: 'info' },
}

interface TableExpose {
  reload: (p?: number) => void
}
const tableRef = ref<TableExpose>()

async function fetchPlatforms(page: number, pageSize: number): Promise<PageData<Platform>> {
  const all = await listPlatforms()
  return {
    total: all.length,
    page,
    page_size: pageSize,
    list: all.slice((page - 1) * pageSize, page * pageSize),
  }
}

/* ── 新建 / 编辑 ── */
const dialogVisible = ref(false)
const dialogMode = ref<'create' | 'edit'>('create')
const editingId = ref('')
const saving = ref(false)
const formRef = ref<FormInstance>()
const form = reactive({
  name: '',
  code: '',
  type: 'openai' as PlatformType,
  base_url: '',
  status: 'active' as PlatformStatus,
})

const rules: FormRules = {
  name: [
    { required: true, message: '请输入平台名称', trigger: 'blur' },
    { max: 100, message: '不超过 100 个字符', trigger: 'blur' },
  ],
  code: [
    { required: true, message: '请输入标识码', trigger: 'blur' },
    { max: 50, message: '不超过 50 个字符', trigger: 'blur' },
  ],
  type: [{ required: true, message: '请选择协议类型', trigger: 'change' }],
  base_url: [
    { required: true, message: '请输入 Base URL', trigger: 'blur' },
    { type: 'url', message: 'URL 格式不正确', trigger: 'blur' },
    { max: 500, message: '不超过 500 个字符', trigger: 'blur' },
  ],
}

function openCreate(): void {
  dialogMode.value = 'create'
  form.name = ''
  form.code = ''
  form.type = 'openai'
  form.base_url = ''
  form.status = 'active'
  dialogVisible.value = true
}

function openEdit(row: Platform): void {
  dialogMode.value = 'edit'
  editingId.value = row.id
  form.name = row.name
  form.code = row.code
  form.type = row.type
  form.base_url = row.base_url
  form.status = row.status
  dialogVisible.value = true
}

async function submit(): Promise<void> {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    saving.value = true
    try {
      const req = {
        name: form.name.trim(),
        code: form.code.trim(),
        type: form.type,
        base_url: form.base_url.trim(),
        status: form.status,
      }
      if (dialogMode.value === 'create') {
        await createPlatform(req)
        ElMessage.success('平台已创建')
      } else {
        await updatePlatform(editingId.value, req)
        ElMessage.success('已保存')
      }
      dialogVisible.value = false
      tableRef.value?.reload()
    } finally {
      saving.value = false
    }
  })
}

async function remove(row: Platform): Promise<void> {
  try {
    await ElMessageBox.confirm(
      `确认删除平台「${row.name}」？其下账号与模型将不可用。`,
      '删除确认',
      { type: 'warning', confirmButtonText: '删除', cancelButtonText: '取消' },
    )
  } catch {
    return
  }
  await deletePlatform(row.id)
  ElMessage.success('已删除')
  tableRef.value?.reload()
}
</script>

<template>
  <div class="relay-tab">
    <div class="relay-tab__bar">
      <span class="relay-tab__hint">上游平台定义：协议类型决定转发适配器，标识码用于图标匹配。</span>
      <el-button type="primary" :icon="'Plus'" @click="openCreate">新建平台</el-button>
    </div>

    <AdaptiveTable ref="tableRef" :fetch="fetchPlatforms">
      <el-table-column label="平台" min-width="170">
        <template #default="{ row }">
          <span class="relay-tab__platform">
            <PlatformIcon :code="row.code" :size="18" />
            <span class="relay-tab__name">{{ row.name }}</span>
          </span>
        </template>
      </el-table-column>
      <el-table-column prop="code" label="标识码" min-width="110" show-overflow-tooltip />
      <el-table-column label="协议类型" width="130">
        <template #default="{ row }">{{ TYPE_LABELS[row.type as PlatformType] ?? row.type }}</template>
      </el-table-column>
      <el-table-column prop="base_url" label="Base URL" min-width="200" show-overflow-tooltip />
      <el-table-column label="状态" width="90">
        <template #default="{ row }">
          <el-tag :type="STATUS_META[row.status]?.tag ?? 'info'" size="small" disable-transitions>
            {{ STATUS_META[row.status]?.label ?? row.status }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="创建时间" min-width="160">
        <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="120">
        <template #default="{ row }">
          <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
          <el-button link type="danger" @click="remove(row)">删除</el-button>
        </template>
      </el-table-column>
    </AdaptiveTable>

    <el-dialog
      v-model="dialogVisible"
      :title="dialogMode === 'create' ? '新建平台' : '编辑平台'"
      width="520px"
      destroy-on-close
    >
      <el-form ref="formRef" :model="form" :rules="rules" label-width="92px">
        <el-form-item label="平台名称" prop="name">
          <el-input v-model="form.name" maxlength="100" />
        </el-form-item>
        <el-form-item label="标识码" prop="code">
          <el-input v-model="form.code" placeholder="如 deepseek / openai，用于品牌图标匹配" maxlength="50" />
        </el-form-item>
        <el-form-item label="协议类型" prop="type">
          <el-select v-model="form.type" class="relay-tab__full">
            <el-option v-for="(label, val) in TYPE_LABELS" :key="val" :label="label" :value="val" />
          </el-select>
        </el-form-item>
        <el-form-item label="Base URL" prop="base_url">
          <el-input v-model="form.base_url" placeholder="https://api.example.com/v1" maxlength="500" />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-select v-model="form.status" class="relay-tab__full">
            <el-option v-for="(meta, val) in STATUS_META" :key="val" :label="meta.label" :value="val" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="submit">
          {{ dialogMode === 'create' ? '创建' : '保存' }}
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
  justify-content: space-between;
  gap: 12px;
}
.relay-tab__hint {
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
