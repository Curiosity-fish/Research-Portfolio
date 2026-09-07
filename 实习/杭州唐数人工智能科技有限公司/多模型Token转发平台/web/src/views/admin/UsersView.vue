<script setup lang="ts">
/* 用户管理页：筛选（关键字/角色/状态）+ 自适应表格 + 新建/编辑/状态管理 + Token 抽屉。
   契约已对照后端 handler/service 核验：列表分页，Token 列表不分页（见 UserTokenDrawer） */
import { reactive, ref } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import PageHeader from '@/components/common/PageHeader.vue'
import AdaptiveTable from '@/components/adaptive/AdaptiveTable.vue'
import UserTokenDrawer from '@/components/users/UserTokenDrawer.vue'
import {
  createUser,
  listUsers,
  updateUser,
  updateUserStatus,
  type UserListQuery,
} from '@/api/admin/user'
import type { User, UserRole, UserStatus } from '@/api/types'
import { formatTime } from '@/utils/format'

const ROLE_LABELS: Record<UserRole, string> = {
  student: '学生',
  teacher: '教师',
  staff: '教职工',
}

const STATUS_META: Record<UserStatus, { label: string; tag: 'success' | 'info' | 'danger'; done: string }> = {
  active: { label: '正常', tag: 'success', done: '已启用' },
  inactive: { label: '未激活', tag: 'info', done: '已设为未激活' },
  banned: { label: '已封禁', tag: 'danger', done: '已封禁' },
}

/* ── 筛选与表格 ── */
// AdaptiveTable 是泛型组件，InstanceType<typeof ...> 不适用；按其 defineExpose 结构取最小接口
interface TableExpose {
  reload: (p?: number) => void
}
const tableRef = ref<TableExpose>()
const filters = reactive({ keyword: '', role: '', status: '' })

async function fetchUsers(page: number, pageSize: number) {
  const query: UserListQuery = { page, page_size: pageSize }
  if (filters.keyword.trim()) query.keyword = filters.keyword.trim()
  if (filters.role) query.role = filters.role
  if (filters.status) query.status = filters.status
  return listUsers(query)
}

function applyFilters(): void {
  tableRef.value?.reload(1)
}

function resetFilters(): void {
  filters.keyword = ''
  filters.role = ''
  filters.status = ''
  applyFilters()
}

/* ── 新建 / 编辑 ── */
const dialogVisible = ref(false)
const dialogMode = ref<'create' | 'edit'>('create')
const editingId = ref('')
const saving = ref(false)
const formRef = ref<FormInstance>()
const form = reactive({
  username: '',
  password: '',
  name: '',
  role: 'student' as UserRole,
  email: '',
  phone: '',
})

const rules: FormRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 50, message: '长度 3-50 个字符', trigger: 'blur' },
  ],
  password: [
    {
      validator: (_r, v: string, cb) => {
        if (dialogMode.value === 'create' && !v) cb(new Error('请输入密码'))
        else if (v && v.length < 8) cb(new Error('至少 8 个字符'))
        else if (v && v.length > 72) cb(new Error('不超过 72 个字符'))
        else cb()
      },
      trigger: 'blur',
    },
  ],
  name: [
    { required: true, message: '请输入姓名', trigger: 'blur' },
    { max: 100, message: '不超过 100 个字符', trigger: 'blur' },
  ],
  role: [{ required: true, message: '请选择角色', trigger: 'change' }],
  email: [{ type: 'email', message: '邮箱格式不正确', trigger: 'blur' }],
  phone: [{ max: 20, message: '不超过 20 个字符', trigger: 'blur' }],
}

function openCreate(): void {
  dialogMode.value = 'create'
  form.username = ''
  form.password = ''
  form.name = ''
  form.role = 'student'
  form.email = ''
  form.phone = ''
  dialogVisible.value = true
}

function openEdit(row: User): void {
  dialogMode.value = 'edit'
  editingId.value = row.id
  form.username = row.username
  form.password = ''
  form.name = row.name
  form.role = row.role
  form.email = row.email ?? ''
  form.phone = row.phone ?? ''
  dialogVisible.value = true
}

async function submitUser(): Promise<void> {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    saving.value = true
    try {
      if (dialogMode.value === 'create') {
        await createUser({
          username: form.username.trim(),
          password: form.password,
          name: form.name.trim(),
          role: form.role,
          email: form.email.trim() || undefined,
          phone: form.phone.trim() || undefined,
        })
        ElMessage.success('用户已创建')
      } else {
        await updateUser(editingId.value, {
          name: form.name.trim(),
          role: form.role,
          email: form.email.trim(),
          phone: form.phone.trim(),
          ...(form.password ? { password: form.password } : {}),
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

/* ── 状态管理 ── */
async function changeStatus(row: User, status: UserStatus): Promise<void> {
  if (row.status === status) return
  if (status === 'banned') {
    try {
      await ElMessageBox.confirm(
        `确认封禁用户「${row.name}」？封禁后其 API Token 将无法调用。`,
        '封禁确认',
        { type: 'warning', confirmButtonText: '封禁', cancelButtonText: '取消' },
      )
    } catch {
      return
    }
  }
  await updateUserStatus(row.id, { status })
  ElMessage.success(STATUS_META[status].done)
  tableRef.value?.reload()
}

/* ── Token 抽屉 ── */
const tokenUser = ref<User | null>(null)

function openTokens(row: User): void {
  tokenUser.value = row
}
</script>

<template>
  <div class="page-root">
    <PageHeader title="用户管理" description="用户账号 CRUD、状态管理与 API Token 签发">
      <template #actions>
        <el-button type="primary" :icon="'Plus'" @click="openCreate">新建用户</el-button>
      </template>
    </PageHeader>

    <!-- 筛选栏：关键字匹配用户名/姓名（后端 ContainsFold），角色/状态精确过滤 -->
    <div class="users__toolbar app-card">
      <el-input
        v-model="filters.keyword"
        placeholder="用户名 / 姓名"
        clearable
        :prefix-icon="'Search'"
        class="users__keyword"
        @keyup.enter="applyFilters"
        @clear="applyFilters"
      />
      <el-select v-model="filters.role" placeholder="全部角色" clearable class="users__select" @change="applyFilters">
        <el-option v-for="(label, val) in ROLE_LABELS" :key="val" :label="label" :value="val" />
      </el-select>
      <el-select v-model="filters.status" placeholder="全部状态" clearable class="users__select" @change="applyFilters">
        <el-option v-for="(meta, val) in STATUS_META" :key="val" :label="meta.label" :value="val" />
      </el-select>
      <el-button type="primary" :icon="'Search'" @click="applyFilters">查询</el-button>
      <el-button :icon="'RefreshLeft'" @click="resetFilters">重置</el-button>
    </div>

    <AdaptiveTable ref="tableRef" :fetch="fetchUsers" class="users__table">
      <el-table-column prop="username" label="用户名" min-width="130" show-overflow-tooltip />
      <el-table-column prop="name" label="姓名" min-width="100" show-overflow-tooltip />
      <el-table-column label="角色" width="90">
        <template #default="{ row }">{{ ROLE_LABELS[row.role as UserRole] ?? row.role }}</template>
      </el-table-column>
      <el-table-column label="状态" width="90">
        <template #default="{ row }">
          <el-tag :type="STATUS_META[row.status as UserStatus]?.tag ?? 'info'" size="small" disable-transitions>
            {{ STATUS_META[row.status as UserStatus]?.label ?? row.status }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="邮箱" min-width="170" show-overflow-tooltip>
        <template #default="{ row }">{{ row.email || '-' }}</template>
      </el-table-column>
      <el-table-column label="手机" min-width="120">
        <template #default="{ row }">{{ row.phone || '-' }}</template>
      </el-table-column>
      <el-table-column label="创建时间" min-width="160">
        <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
      </el-table-column>
      <el-table-column label="操作" min-width="210">
        <template #default="{ row }">
          <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
          <el-button link type="primary" @click="openTokens(row)">Token 管理</el-button>
          <el-dropdown
            trigger="click"
            @command="(cmd: UserStatus) => changeStatus(row, cmd)"
          >
            <el-button link type="primary">
              状态<el-icon :size="12"><ArrowDown /></el-icon>
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item
                  v-for="(meta, val) in STATUS_META"
                  :key="val"
                  :command="val"
                  :disabled="row.status === val"
                >
                  {{ meta.label }}
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </template>
      </el-table-column>
    </AdaptiveTable>

    <!-- 新建 / 编辑用户 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogMode === 'create' ? '新建用户' : '编辑用户'"
      width="520px"
      destroy-on-close
    >
      <el-form ref="formRef" :model="form" :rules="rules" label-width="84px">
        <el-form-item v-if="dialogMode === 'create'" label="用户名" prop="username">
          <el-input v-model="form.username" placeholder="3-50 个字符，创建后不可修改" maxlength="50" />
        </el-form-item>
        <el-form-item :label="dialogMode === 'create' ? '密码' : '重置密码'" prop="password">
          <el-input
            v-model="form.password"
            type="password"
            show-password
            :placeholder="dialogMode === 'create' ? '至少 8 个字符' : '留空则不修改密码'"
            maxlength="72"
          />
        </el-form-item>
        <el-form-item label="姓名" prop="name">
          <el-input v-model="form.name" maxlength="100" />
        </el-form-item>
        <el-form-item label="角色" prop="role">
          <el-select v-model="form.role" class="users__full">
            <el-option v-for="(label, val) in ROLE_LABELS" :key="val" :label="label" :value="val" />
          </el-select>
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="form.email" placeholder="可选" maxlength="255" />
        </el-form-item>
        <el-form-item label="手机" prop="phone">
          <el-input v-model="form.phone" placeholder="可选" maxlength="20" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="submitUser">
          {{ dialogMode === 'create' ? '创建' : '保存' }}
        </el-button>
      </template>
    </el-dialog>

    <UserTokenDrawer :user="tokenUser" @closed="tokenUser = null" />
  </div>
</template>

<style scoped>
.users__toolbar {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
}
.users__keyword {
  width: 220px;
}
.users__select {
  width: 130px;
}
.users__full {
  width: 100%;
}
.users__table {
  min-height: 0;
}
</style>
