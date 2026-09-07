<script setup lang="ts">
/* 告警规则：列表 + 新建/编辑/删除 + 行内启停（复用更新接口，只传 enabled） */
import { reactive, ref } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import AdaptiveTable from '@/components/adaptive/AdaptiveTable.vue'
import { createAlertRule, deleteAlertRule, listAlertRules, updateAlertRule } from '@/api/admin/alert'
import type { AlertMetric, AlertRule, AlertRuleCreateReq, PageData } from '@/api/types'
import { formatTime } from '@/utils/format'

const METRIC_META: Record<AlertMetric, { label: string; tag: 'warning' | 'danger' }> = {
  balance_low: { label: '余额过低', tag: 'warning' },
  quota_low: { label: '配额过低', tag: 'warning' },
  error_rate: { label: '错误次数', tag: 'danger' },
  cost_spike: { label: '消费突增', tag: 'danger' },
}

/* 阈值单位说明：与 service/alert.go 评估逻辑逐项对应 */
const THRESHOLD_HINT =
  '阈值单位：余额过低 / 消费突增为 micro 元（1 元 = 1,000,000）；配额过低为 token 数；错误次数为账号连续错误计数。'

interface TableExpose {
  reload: (p?: number) => void
}
const tableRef = ref<TableExpose>()

async function fetchRules(page: number, pageSize: number): Promise<PageData<AlertRule>> {
  return listAlertRules({ page, page_size: pageSize })
}

async function toggleEnabled(row: AlertRule, value: boolean): Promise<void> {
  try {
    await updateAlertRule(row.id, { enabled: value })
    ElMessage.success(value ? '规则已启用' : '规则已停用')
  } finally {
    // 无论成败都重载，保证开关与后端状态一致
    tableRef.value?.reload()
  }
}

async function removeRule(row: AlertRule): Promise<void> {
  await ElMessageBox.confirm(`确认删除规则「${row.name}」？删除后不再评估该规则。`, '删除确认', {
    type: 'warning',
    confirmButtonText: '删除',
    cancelButtonText: '取消',
  })
  await deleteAlertRule(row.id)
  ElMessage.success('规则已删除')
  tableRef.value?.reload()
}

/* ── 新建/编辑对话框 ── */
const dialogVisible = ref(false)
const editingId = ref<string | null>(null)
const saving = ref(false)
const formRef = ref<FormInstance>()
const form = reactive<{
  name: string
  metric: AlertMetric
  threshold: number | null
  enabled: boolean
  description: string
}>({ name: '', metric: 'balance_low', threshold: null, enabled: true, description: '' })

const rules: FormRules = {
  name: [
    { required: true, message: '请输入规则名称', trigger: 'blur' },
    { max: 200, message: '不超过 200 个字符', trigger: 'blur' },
  ],
  metric: [{ required: true, message: '请选择监控指标', trigger: 'change' }],
  threshold: [{ required: true, message: '请输入阈值', trigger: 'blur' }],
}

function openCreate(): void {
  editingId.value = null
  form.name = ''
  form.metric = 'balance_low'
  form.threshold = null
  form.enabled = true
  form.description = ''
  dialogVisible.value = true
}

function openEdit(row: AlertRule): void {
  editingId.value = row.id
  form.name = row.name
  form.metric = row.metric
  form.threshold = row.threshold
  form.enabled = row.enabled
  form.description = row.description ?? ''
  dialogVisible.value = true
}

async function submit(): Promise<void> {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid || form.threshold === null) return
    saving.value = true
    const payload: AlertRuleCreateReq = {
      name: form.name.trim(),
      metric: form.metric,
      threshold: form.threshold,
      enabled: form.enabled,
      ...(form.description.trim() ? { description: form.description.trim() } : {}),
    }
    try {
      if (editingId.value) {
        await updateAlertRule(editingId.value, payload)
        ElMessage.success('规则已更新')
      } else {
        await createAlertRule(payload)
        ElMessage.success('规则已创建')
      }
      dialogVisible.value = false
      tableRef.value?.reload()
    } finally {
      saving.value = false
    }
  })
}
</script>

<template>
  <div class="rule-tab">
    <div class="rule-tab__toolbar app-card">
      <span class="rule-tab__hint">{{ THRESHOLD_HINT }}</span>
      <el-button type="primary" :icon="'Plus'" @click="openCreate">新建规则</el-button>
    </div>

    <AdaptiveTable ref="tableRef" :fetch="fetchRules" class="rule-tab__table">
      <el-table-column prop="name" label="名称" min-width="150" show-overflow-tooltip />
      <el-table-column label="指标" width="110">
        <template #default="{ row }">
          <el-tag :type="METRIC_META[row.metric as AlertMetric]?.tag ?? 'info'" size="small" disable-transitions>
            {{ METRIC_META[row.metric as AlertMetric]?.label ?? row.metric }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="threshold" label="阈值" width="120" align="right" class-name="tnum" />
      <el-table-column label="启用" width="90">
        <template #default="{ row }">
          <el-switch :model-value="(row as AlertRule).enabled" @change="(v: boolean) => toggleEnabled(row as AlertRule, v)" />
        </template>
      </el-table-column>
      <el-table-column label="描述" min-width="170" show-overflow-tooltip>
        <template #default="{ row }">{{ (row as AlertRule).description || '—' }}</template>
      </el-table-column>
      <el-table-column label="更新时间" min-width="150">
        <template #default="{ row }">{{ formatTime((row as AlertRule).updated_at) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="130" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="openEdit(row as AlertRule)">编辑</el-button>
          <el-button link type="danger" @click="removeRule(row as AlertRule)">删除</el-button>
        </template>
      </el-table-column>
    </AdaptiveTable>

    <el-dialog v-model="dialogVisible" :title="editingId ? '编辑规则' : '新建规则'" width="520px" destroy-on-close>
      <el-form ref="formRef" :model="form" :rules="rules" label-width="92px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="form.name" maxlength="200" placeholder="规则名称" />
        </el-form-item>
        <el-form-item label="指标" prop="metric">
          <el-select v-model="form.metric" class="rule-tab__full">
            <el-option v-for="(meta, val) in METRIC_META" :key="val" :label="meta.label" :value="val" />
          </el-select>
        </el-form-item>
        <el-form-item label="阈值" prop="threshold">
          <el-input-number v-model="form.threshold" :min="1" :precision="0" class="rule-tab__full" controls-position="right" />
        </el-form-item>
        <el-form-item label="启用" prop="enabled">
          <el-switch v-model="form.enabled" />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input v-model="form.description" type="textarea" :rows="3" maxlength="500" placeholder="可选，说明规则用途" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="submit">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.rule-tab {
  height: 100%;
  min-height: 0;
  display: flex;
  flex-direction: column;
}
.rule-tab__toolbar {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
}
.rule-tab__hint {
  flex: 1;
  font-size: 12px;
  color: var(--text-muted);
}
.rule-tab__table {
  min-height: 0;
}
.rule-tab__full {
  width: 100%;
}
</style>
