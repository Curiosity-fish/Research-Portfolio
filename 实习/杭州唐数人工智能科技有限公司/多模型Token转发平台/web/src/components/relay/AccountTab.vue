<script setup lang="ts">
/* 账号 tab：上游平台下的 API Key 账号池。列表不分页，按平台客户端过滤 */
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import AdaptiveTable from '@/components/adaptive/AdaptiveTable.vue'
import PlatformIcon from '@/components/common/PlatformIcon.vue'
import { createAccount, deleteAccount, listAccounts, listPlatforms, updateAccount } from '@/api/admin/relay'
import type { Account, AccountStatus, PageData, Platform } from '@/api/types'
import { formatTime } from '@/utils/format'

const STATUS_META: Record<AccountStatus, { label: string; tag: 'success' | 'info' | 'danger' }> = {
  active: { label: '正常', tag: 'success' },
  inactive: { label: '停用', tag: 'info' },
  error: { label: '异常', tag: 'danger' },
}

interface TableExpose {
  reload: (p?: number) => void
}
const tableRef = ref<TableExpose>()

const platforms = ref<Platform[]>([])
const platformMap = computed(() => new Map(platforms.value.map((p) => [p.id, p])))
const filterPlatform = ref('')

const accountsCache = ref<Account[]>([])

onMounted(async () => {
  platforms.value = await listPlatforms()
})

async function fetchAccounts(page: number, pageSize: number): Promise<PageData<Account>> {
  const all = await listAccounts()
  accountsCache.value = all
  const filtered = filterPlatform.value ? all.filter((a) => a.platform_id === filterPlatform.value) : all
  return {
    total: filtered.length,
    page,
    page_size: pageSize,
    list: filtered.slice((page - 1) * pageSize, page * pageSize),
  }
}

function applyFilter(): void {
  tableRef.value?.reload(1)
}

/* ── 新建 / 编辑 ── */
const dialogVisible = ref(false)
const dialogMode = ref<'create' | 'edit'>('create')
const editingId = ref('')
const saving = ref(false)
const formRef = ref<FormInstance>()
const form = reactive({
  platform_id: '',
  name: '',
  api_key: '',
  weight: 1,
  max_rpm: 0,
  status: 'active' as AccountStatus,
})

const rules: FormRules = {
  platform_id: [{ required: true, message: '请选择所属平台', trigger: 'change' }],
  name: [
    { required: true, message: '请输入账号名称', trigger: 'blur' },
    { max: 100, message: '不超过 100 个字符', trigger: 'blur' },
  ],
  api_key: [
    {
      validator: (_r, v: string, cb) => {
        if (dialogMode.value === 'create' && !v) cb(new Error('请输入 API Key'))
        else cb()
      },
      trigger: 'blur',
    },
  ],
}

function openCreate(): void {
  dialogMode.value = 'create'
  form.platform_id = filterPlatform.value
  form.name = ''
  form.api_key = ''
  form.weight = 1
  form.max_rpm = 0
  form.status = 'active'
  dialogVisible.value = true
}

function openEdit(row: Account): void {
  dialogMode.value = 'edit'
  editingId.value = row.id
  form.platform_id = row.platform_id
  form.name = row.name
  form.api_key = ''
  form.weight = row.weight
  form.max_rpm = row.max_rpm
  form.status = row.status
  dialogVisible.value = true
}

async function submit(): Promise<void> {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    saving.value = true
    try {
      if (dialogMode.value === 'create') {
        await createAccount({
          platform_id: form.platform_id,
          name: form.name.trim(),
          api_key: form.api_key.trim(),
          weight: form.weight,
          max_rpm: form.max_rpm,
          status: form.status,
        })
        ElMessage.success('账号已创建')
      } else {
        await updateAccount(editingId.value, {
          platform_id: form.platform_id,
          name: form.name.trim(),
          api_key: form.api_key.trim(),
          weight: form.weight,
          max_rpm: form.max_rpm,
          status: form.status,
        })
        ElMessage.success('已保存')
      }
      dialogVisible.value = false
      tableRef.value?.reload()
    } finally {
      saving.value = false
    }
  })
}

async function remove(row: Account): Promise<void> {
  try {
    await ElMessageBox.confirm(`确认删除账号「${row.name}」？`, '删除确认', {
      type: 'warning',
      confirmButtonText: '删除',
      cancelButtonText: '取消',
    })
  } catch {
    return
  }
  await deleteAccount(row.id)
  ElMessage.success('已删除')
  tableRef.value?.reload()
}
</script>

<template>
  <div class="relay-tab">
    <div class="relay-tab__bar">
      <el-select
        v-model="filterPlatform"
        placeholder="全部平台"
        clearable
        class="relay-tab__filter"
        @change="applyFilter"
      >
        <el-option v-for="p in platforms" :key="p.id" :label="p.name" :value="p.id" />
      </el-select>
      <span class="relay-tab__hint">API Key 由服务端加密存储，仅标记是否已设置；权重用于同平台多账号分流。</span>
      <el-button type="primary" :icon="'Plus'" @click="openCreate">新建账号</el-button>
    </div>

    <AdaptiveTable ref="tableRef" :fetch="fetchAccounts">
      <el-table-column prop="name" label="名称" min-width="140" show-overflow-tooltip />
      <el-table-column label="所属平台" min-width="150">
        <template #default="{ row }">
          <span v-if="platformMap.get(row.platform_id)" class="relay-tab__platform">
            <PlatformIcon :code="platformMap.get(row.platform_id)!.code" :size="16" />
            {{ platformMap.get(row.platform_id)!.name }}
          </span>
          <span v-else>-</span>
        </template>
      </el-table-column>
      <el-table-column label="API Key" width="90" align="center">
        <template #default="{ row }">
          <el-tag :type="row.api_key_encrypted ? 'success' : 'info'" size="small" disable-transitions>
            {{ row.api_key_encrypted ? '已设置' : '未设置' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="weight" label="权重" width="70" align="right" />
      <el-table-column label="限速 (RPM)" width="100" align="right">
        <template #default="{ row }">{{ row.max_rpm > 0 ? row.max_rpm : '不限' }}</template>
      </el-table-column>
      <el-table-column label="状态" width="90">
        <template #default="{ row }">
          <el-tag :type="STATUS_META[row.status as AccountStatus]?.tag ?? 'info'" size="small" disable-transitions>
            {{ STATUS_META[row.status as AccountStatus]?.label ?? row.status }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="error_count" label="错误次数" width="90" align="right" />
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
      :title="dialogMode === 'create' ? '新建账号' : '编辑账号'"
      width="520px"
      destroy-on-close
    >
      <el-form ref="formRef" :model="form" :rules="rules" label-width="92px">
        <el-form-item label="所属平台" prop="platform_id">
          <el-select v-model="form.platform_id" class="relay-tab__full" :disabled="dialogMode === 'edit'">
            <el-option v-for="p in platforms" :key="p.id" :label="p.name" :value="p.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="名称" prop="name">
          <el-input v-model="form.name" maxlength="100" />
        </el-form-item>
        <el-form-item label="API Key" prop="api_key">
          <el-input
            v-model="form.api_key"
            type="password"
            show-password
            :placeholder="dialogMode === 'create' ? '上游平台颁发的密钥' : '留空则不修改'"
          />
        </el-form-item>
        <el-form-item label="权重" prop="weight">
          <el-input-number v-model="form.weight" :min="1" :precision="0" controls-position="right" class="relay-tab__full" />
        </el-form-item>
        <el-form-item label="限速 RPM" prop="max_rpm">
          <el-input-number v-model="form.max_rpm" :min="0" :precision="0" controls-position="right" class="relay-tab__full" />
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
  gap: 12px;
}
.relay-tab__filter {
  width: 180px;
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
  gap: 6px;
}
.relay-tab__full {
  width: 100%;
}
</style>
