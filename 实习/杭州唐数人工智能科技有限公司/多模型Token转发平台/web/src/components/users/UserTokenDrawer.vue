<script setup lang="ts">
/* 用户 Token 抽屉：列表（后端不分页、一次返回全部）、新建、启停、删除。
   明文 Token 仅在创建响应中出现一次，用独立对话框展示并提供复制。 */
import { reactive, ref, watch } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { createToken, deleteToken, listTokens, updateTokenStatus } from '@/api/admin/user'
import type { TokenCreateReq, TokenCreateResp, User, UserToken } from '@/api/types'
import { formatCount, formatTime } from '@/utils/format'

const props = defineProps<{ user: User | null }>()
const emit = defineEmits<{ (e: 'closed'): void }>()

const visible = ref(false)
const loading = ref(false)
const tokens = ref<UserToken[]>([])

watch(
  () => props.user,
  (u) => {
    if (u) {
      visible.value = true
      void loadTokens()
    }
  },
)

async function loadTokens(): Promise<void> {
  if (!props.user) return
  loading.value = true
  try {
    const data = await listTokens(props.user.id)
    tokens.value = data.list
  } finally {
    loading.value = false
  }
}

/* ── 新建 Token ── */
const createVisible = ref(false)
const saving = ref(false)
const formRef = ref<FormInstance>()
const form = reactive({ name: '', quota_limit: undefined as number | undefined, expires_at: '' })

const rules: FormRules = {
  name: [
    { required: true, message: '请输入 Token 名称', trigger: 'blur' },
    { max: 100, message: '不超过 100 个字符', trigger: 'blur' },
  ],
}

function openCreate(): void {
  form.name = ''
  form.quota_limit = undefined
  form.expires_at = ''
  createVisible.value = true
}

async function submitCreate(): Promise<void> {
  if (!formRef.value || !props.user) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    const req: TokenCreateReq = { name: form.name.trim() }
    if (form.quota_limit !== undefined) req.quota_limit = form.quota_limit
    if (form.expires_at) req.expires_at = form.expires_at
    saving.value = true
    try {
      const resp = await createToken(props.user!.id, req)
      createVisible.value = false
      plainResp.value = resp
      plainVisible.value = true
      void loadTokens()
    } finally {
      saving.value = false
    }
  })
}

/* ── 明文一次性展示 ── */
const plainVisible = ref(false)
const plainResp = ref<TokenCreateResp | null>(null)

async function copyPlaintext(): Promise<void> {
  if (!plainResp.value) return
  try {
    await navigator.clipboard.writeText(plainResp.value.plaintext_token)
    ElMessage.success('已复制到剪贴板')
  } catch {
    ElMessage.warning('复制失败，请手动选择复制')
  }
}

/* ── 启停 / 删除 ── */
async function toggleEnabled(t: UserToken, val: boolean): Promise<void> {
  if (!props.user) return
  try {
    await updateTokenStatus(props.user.id, t.id, val)
    t.is_enabled = val
    ElMessage.success(val ? '已启用' : '已禁用')
  } catch {
    // model-value 绑定，不改 t.is_enabled 即自动回退，拦截器已提示原因
  }
}

async function removeToken(t: UserToken): Promise<void> {
  if (!props.user) return
  try {
    await ElMessageBox.confirm(`确认删除 Token「${t.name}」？删除后立即失效且不可恢复。`, '删除确认', {
      type: 'warning',
      confirmButtonText: '删除',
      cancelButtonText: '取消',
    })
  } catch {
    return
  }
  await deleteToken(props.user.id, t.id)
  ElMessage.success('已删除')
  void loadTokens()
}

function onDrawerClose(): void {
  visible.value = false
  plainVisible.value = false
  plainResp.value = null
  emit('closed')
}
</script>

<template>
  <el-drawer
    :model-value="visible"
    :title="`Token 管理 · ${user?.name ?? ''}`"
    size="660px"
    @close="onDrawerClose"
  >
    <div class="token-drawer">
      <div class="token-drawer__bar">
        <span class="token-drawer__hint">API Token 用于调用转发接口，明文仅在创建时展示一次。</span>
        <el-button type="primary" :icon="'Plus'" @click="openCreate">新建 Token</el-button>
      </div>

      <el-table v-loading="loading" :data="tokens" class="token-drawer__table">
        <el-table-column prop="name" label="名称" min-width="110" show-overflow-tooltip />
        <el-table-column label="Token" min-width="110">
          <template #default="{ row }">
            <code class="token-drawer__mask">sk-****{{ row.token_last4 }}</code>
          </template>
        </el-table-column>
        <el-table-column label="额度（已用/上限）" min-width="130" align="right">
          <template #default="{ row }">
            <span class="tnum">{{ formatCount(row.quota_used) }}</span>
            <span class="token-drawer__sep">/</span>
            <span class="tnum">{{ row.quota_limit !== undefined ? formatCount(row.quota_limit) : '不限' }}</span>
          </template>
        </el-table-column>
        <el-table-column label="过期时间" min-width="130">
          <template #default="{ row }">{{ row.expires_at ? formatTime(row.expires_at) : '永不过期' }}</template>
        </el-table-column>
        <el-table-column label="启用" width="56" align="center">
          <template #default="{ row }">
            <el-switch :model-value="row.is_enabled" @change="(v: boolean) => toggleEnabled(row, v)" />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="60" align="center">
          <template #default="{ row }">
            <el-button link type="danger" @click="removeToken(row)">删除</el-button>
          </template>
        </el-table-column>
        <template #empty>
          <el-empty description="该用户暂无 Token" :image-size="72" />
        </template>
      </el-table>
    </div>

    <!-- 新建 Token -->
    <el-dialog
      v-model="createVisible"
      title="新建 Token"
      width="480px"
      destroy-on-close
      append-to-body
    >
      <el-form ref="formRef" :model="form" :rules="rules" label-width="92px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="form.name" placeholder="如：毕业设计调用" maxlength="100" />
        </el-form-item>
        <el-form-item label="额度上限" prop="quota_limit">
          <el-input-number
            v-model="form.quota_limit"
            :min="0"
            :precision="0"
            :step="100000"
            controls-position="right"
            placeholder="不填则不限"
            class="token-drawer__quota"
          />
        </el-form-item>
        <el-form-item label="过期时间" prop="expires_at">
          <el-date-picker
            v-model="form.expires_at"
            type="datetime"
            value-format="YYYY-MM-DDTHH:mm:ssZ"
            placeholder="不填则永不过期"
            class="token-drawer__quota"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="submitCreate">创建</el-button>
      </template>
    </el-dialog>

    <!-- 明文一次性展示 -->
    <el-dialog v-model="plainVisible" title="Token 创建成功" width="580px" :close-on-click-modal="false">
      <el-alert
        type="warning"
        :closable="false"
        title="明文仅此一次展示，关闭后无法再次查看，请立即复制保存。"
        class="token-drawer__alert"
      />
      <div class="token-drawer__plaintext">
        <el-input :model-value="plainResp?.plaintext_token" type="textarea" :rows="3" readonly resize="none" />
        <el-button type="primary" :icon="'CopyDocument'" @click="copyPlaintext">复制</el-button>
      </div>
      <div class="token-drawer__meta">
        名称：{{ plainResp?.name }} · 创建于 {{ formatTime(plainResp?.created_at) }}
      </div>
      <template #footer>
        <el-button type="primary" @click="plainVisible = false">我已保存</el-button>
      </template>
    </el-dialog>
  </el-drawer>
</template>

<style scoped>
.token-drawer {
  display: flex;
  flex-direction: column;
  gap: 12px;
  height: 100%;
}
.token-drawer__bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  flex-shrink: 0;
}
.token-drawer__hint {
  font-size: 12px;
  color: var(--text-muted);
}
.token-drawer__table {
  flex: 1;
}
.token-drawer__mask {
  font-family: monospace;
  font-size: 12px;
  color: var(--text-secondary);
}
.token-drawer__sep {
  margin: 0 2px;
  color: var(--text-muted);
}
.token-drawer__quota {
  width: 100%;
}
.token-drawer__alert {
  margin-bottom: 12px;
}
.token-drawer__plaintext {
  display: flex;
  gap: 8px;
  align-items: flex-start;
}
.token-drawer__plaintext :deep(.el-input__inner),
.token-drawer__plaintext :deep(textarea) {
  font-family: monospace;
  font-size: 12px;
}
.token-drawer__meta {
  margin-top: 10px;
  font-size: 12px;
  color: var(--text-muted);
}
</style>
