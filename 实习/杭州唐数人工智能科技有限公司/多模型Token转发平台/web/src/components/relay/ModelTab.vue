<script setup lang="ts">
/* 模型 tab：可转发的模型目录与计价。列表不分页；价格单位 元/1K Token（提交时转 micro） */
import { reactive, ref } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import AdaptiveTable from '@/components/adaptive/AdaptiveTable.vue'
import { createModel, deleteModel, listModels, updateModel } from '@/api/admin/relay'
import type { AIModel, ModelType, PageData } from '@/api/types'
import { formatMicroTrim, formatTime, toMicro } from '@/utils/format'

const TYPE_LABELS: Record<ModelType, string> = {
  chat: '对话',
  embedding: '嵌入',
  image: '图像',
}

const TYPE_TAGS: Record<ModelType, 'primary' | 'success' | 'warning'> = {
  chat: 'primary',
  embedding: 'success',
  image: 'warning',
}

interface TableExpose {
  reload: (p?: number) => void
}
const tableRef = ref<TableExpose>()

async function fetchModels(page: number, pageSize: number): Promise<PageData<AIModel>> {
  const all = await listModels()
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
  upstream_name: '',
  type: 'chat' as ModelType,
  input_price_yuan: 0,
  output_price_yuan: 0,
  is_enabled: true,
})

const rules: FormRules = {
  name: [
    { required: true, message: '请输入模型名称', trigger: 'blur' },
    { max: 100, message: '不超过 100 个字符', trigger: 'blur' },
  ],
  upstream_name: [
    { required: true, message: '请输入上游模型名', trigger: 'blur' },
    { max: 100, message: '不超过 100 个字符', trigger: 'blur' },
  ],
  type: [{ required: true, message: '请选择模型类型', trigger: 'change' }],
}

function openCreate(): void {
  dialogMode.value = 'create'
  form.name = ''
  form.upstream_name = ''
  form.type = 'chat'
  form.input_price_yuan = 0
  form.output_price_yuan = 0
  form.is_enabled = true
  dialogVisible.value = true
}

function openEdit(row: AIModel): void {
  dialogMode.value = 'edit'
  editingId.value = row.id
  form.name = row.name
  form.upstream_name = row.upstream_name
  form.type = row.type
  form.input_price_yuan = Number(formatMicroTrim(row.input_price))
  form.output_price_yuan = Number(formatMicroTrim(row.output_price))
  form.is_enabled = row.is_enabled
  dialogVisible.value = true
}

function buildReq() {
  return {
    name: form.name.trim(),
    upstream_name: form.upstream_name.trim(),
    type: form.type,
    input_price: toMicro(form.input_price_yuan) ?? 0,
    output_price: toMicro(form.output_price_yuan) ?? 0,
    is_enabled: form.is_enabled,
  }
}

async function submit(): Promise<void> {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    saving.value = true
    try {
      if (dialogMode.value === 'create') {
        await createModel(buildReq())
        ElMessage.success('模型已创建')
      } else {
        await updateModel(editingId.value, buildReq())
        ElMessage.success('已保存')
      }
      dialogVisible.value = false
      tableRef.value?.reload()
    } finally {
      saving.value = false
    }
  })
}

async function toggleEnabled(row: AIModel, val: boolean): Promise<void> {
  try {
    await updateModel(row.id, {
      name: row.name,
      upstream_name: row.upstream_name,
      type: row.type,
      input_price: row.input_price,
      output_price: row.output_price,
      is_enabled: val,
    })
    row.is_enabled = val
    ElMessage.success(val ? '已启用' : '已停用')
  } catch {
    // model-value 绑定，失败自动回退，拦截器已提示
  }
}

async function remove(row: AIModel): Promise<void> {
  try {
    await ElMessageBox.confirm(`确认删除模型「${row.name}」？删除后不再可转发。`, '删除确认', {
      type: 'warning',
      confirmButtonText: '删除',
      cancelButtonText: '取消',
    })
  } catch {
    return
  }
  await deleteModel(row.id)
  ElMessage.success('已删除')
  tableRef.value?.reload()
}
</script>

<template>
  <div class="relay-tab">
    <div class="relay-tab__bar">
      <span class="relay-tab__hint">名称对终端用户暴露，上游名称是转发给平台的 model 字段；计价为 元/1K Token。</span>
      <el-button type="primary" :icon="'Plus'" @click="openCreate">新建模型</el-button>
    </div>

    <AdaptiveTable ref="tableRef" :fetch="fetchModels">
      <el-table-column prop="name" label="名称" min-width="140" show-overflow-tooltip />
      <el-table-column prop="upstream_name" label="上游名称" min-width="150" show-overflow-tooltip />
      <el-table-column label="类型" width="90">
        <template #default="{ row }">
          <el-tag :type="TYPE_TAGS[row.type as ModelType] ?? 'info'" size="small" disable-transitions>
            {{ TYPE_LABELS[row.type as ModelType] ?? row.type }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="输入价（元/1K）" min-width="120" align="right">
        <template #default="{ row }"><span class="tnum">{{ formatMicroTrim(row.input_price) }}</span></template>
      </el-table-column>
      <el-table-column label="输出价（元/1K）" min-width="120" align="right">
        <template #default="{ row }"><span class="tnum">{{ formatMicroTrim(row.output_price) }}</span></template>
      </el-table-column>
      <el-table-column label="启用" width="70" align="center">
        <template #default="{ row }">
          <el-switch :model-value="row.is_enabled" @change="(v: boolean) => toggleEnabled(row, v)" />
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
      :title="dialogMode === 'create' ? '新建模型' : '编辑模型'"
      width="520px"
      destroy-on-close
    >
      <el-form ref="formRef" :model="form" :rules="rules" label-width="110px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="form.name" placeholder="终端用户可见，如 GPT-4o" maxlength="100" />
        </el-form-item>
        <el-form-item label="上游名称" prop="upstream_name">
          <el-input v-model="form.upstream_name" placeholder="转发给上游的 model 值，如 gpt-4o-2024-08-06" maxlength="100" />
        </el-form-item>
        <el-form-item label="类型" prop="type">
          <el-select v-model="form.type" class="relay-tab__full">
            <el-option v-for="(label, val) in TYPE_LABELS" :key="val" :label="label" :value="val" />
          </el-select>
        </el-form-item>
        <el-form-item label="输入价" prop="input_price_yuan">
          <el-input-number v-model="form.input_price_yuan" :min="0" :precision="4" :step="0.001" controls-position="right" class="relay-tab__full" />
        </el-form-item>
        <el-form-item label="输出价" prop="output_price_yuan">
          <el-input-number v-model="form.output_price_yuan" :min="0" :precision="4" :step="0.001" controls-position="right" class="relay-tab__full" />
        </el-form-item>
        <el-form-item label="启用" prop="is_enabled">
          <el-switch v-model="form.is_enabled" />
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
.relay-tab__full {
  width: 100%;
}
</style>
