<script setup lang="ts">
/* 通知管理：服务端分页；支持类型过滤与新建（不选用户即全体广播） */
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import AdaptiveTable from '@/components/adaptive/AdaptiveTable.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import { createNotification, listNotifications } from '@/api/admin/ops'
import { listUsers } from '@/api/admin/user'
import type { Notification, NotificationType, PageData, User } from '@/api/types'
import { useUserMap } from '@/composables/useUserMap'
import { formatTime } from '@/utils/format'

const { userName } = useUserMap()

const TYPE_META: Record<NotificationType, { label: string; tag: 'primary' | 'warning' }> = {
  announcement: { label: '公告', tag: 'primary' },
  system: { label: '系统', tag: 'warning' },
}

interface TableExpose {
  reload: (p?: number) => void
}
const tableRef = ref<TableExpose>()

const filterType = ref<NotificationType | ''>('')

async function fetchNotifications(page: number, pageSize: number): Promise<PageData<Notification>> {
  return listNotifications({
    page,
    page_size: pageSize,
    ...(filterType.value ? { type: filterType.value } : {}),
  })
}

function applyFilter(): void {
  tableRef.value?.reload(1)
}

/* ── 新建通知 ── */
const users = ref<User[]>([])
onMounted(async () => {
  const data = await listUsers({ page: 1, page_size: 100 })
  users.value = data.list
})

const dialogVisible = ref(false)
const saving = ref(false)
const formRef = ref<FormInstance>()
const form = reactive<{ type: NotificationType; title: string; content: string; user_id: string }>({
  type: 'announcement',
  title: '',
  content: '',
  user_id: '',
})

const rules: FormRules = {
  type: [{ required: true, message: '请选择类型', trigger: 'change' }],
  title: [
    { required: true, message: '请输入标题', trigger: 'blur' },
    { max: 200, message: '不超过 200 个字符', trigger: 'blur' },
  ],
  content: [
    { required: true, message: '请输入内容', trigger: 'blur' },
    { max: 2000, message: '不超过 2000 个字符', trigger: 'blur' },
  ],
}

function openCreate(): void {
  form.type = 'announcement'
  form.title = ''
  form.content = ''
  form.user_id = ''
  dialogVisible.value = true
}

async function submit(): Promise<void> {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    saving.value = true
    try {
      await createNotification({
        type: form.type,
        title: form.title.trim(),
        content: form.content.trim(),
        ...(form.user_id ? { user_id: form.user_id } : {}),
      })
      ElMessage.success('通知已发布')
      dialogVisible.value = false
      tableRef.value?.reload()
    } finally {
      saving.value = false
    }
  })
}
</script>

<template>
  <div class="page-root">
    <PageHeader title="通知" description="站内通知发布与查询；不指定用户时为全体广播" />

    <div class="notify__toolbar app-card">
      <el-select v-model="filterType" placeholder="全部类型" clearable class="notify__filter" @change="applyFilter">
        <el-option v-for="(meta, val) in TYPE_META" :key="val" :label="meta.label" :value="val" />
      </el-select>
      <span class="notify__hint">广播通知面向全部终端用户；定向通知仅指定用户可见。</span>
      <el-button type="primary" :icon="'Plus'" @click="openCreate">新建通知</el-button>
    </div>

    <AdaptiveTable ref="tableRef" :fetch="fetchNotifications" class="notify__table">
      <el-table-column prop="title" label="标题" min-width="150" show-overflow-tooltip />
      <el-table-column label="类型" width="90">
        <template #default="{ row }">
          <el-tag :type="TYPE_META[row.type as NotificationType]?.tag ?? 'info'" size="small" disable-transitions>
            {{ TYPE_META[row.type as NotificationType]?.label ?? row.type }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="接收范围" min-width="130" show-overflow-tooltip>
        <template #default="{ row }">
          {{ row.user_id ? userName(row.user_id) : '全体用户' }}
        </template>
      </el-table-column>
      <el-table-column prop="content" label="内容" min-width="220" show-overflow-tooltip />
      <el-table-column label="创建时间" min-width="150">
        <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
      </el-table-column>
    </AdaptiveTable>

    <el-dialog v-model="dialogVisible" title="新建通知" width="520px" destroy-on-close>
      <el-form ref="formRef" :model="form" :rules="rules" label-width="92px">
        <el-form-item label="类型" prop="type">
          <el-radio-group v-model="form.type">
            <el-radio-button value="announcement">公告</el-radio-button>
            <el-radio-button value="system">系统</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="标题" prop="title">
          <el-input v-model="form.title" maxlength="200" placeholder="通知标题" />
        </el-form-item>
        <el-form-item label="内容" prop="content">
          <el-input v-model="form.content" type="textarea" :rows="4" maxlength="2000" placeholder="通知正文" />
        </el-form-item>
        <el-form-item label="接收用户" prop="user_id">
          <el-select v-model="form.user_id" filterable clearable placeholder="不选择即全体广播" class="notify__full">
            <el-option v-for="u in users" :key="u.id" :label="u.username" :value="u.id" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="submit">发布</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.notify__toolbar {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
}
.notify__filter {
  width: 140px;
  flex-shrink: 0;
}
.notify__hint {
  flex: 1;
  font-size: 12px;
  color: var(--text-muted);
}
.notify__table {
  min-height: 0;
}
.notify__full {
  width: 100%;
}
</style>
