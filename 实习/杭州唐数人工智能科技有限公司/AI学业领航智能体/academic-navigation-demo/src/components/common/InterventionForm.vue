<script setup lang="ts">
import { reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { interveneAlert } from '@/api'

const props = defineProps<{
  alertId?: string
}>()

const emit = defineEmits<{
  saved: []
}>()
const submitting = ref(false)

const form = reactive({
  interventionDate: '',
  methods: [] as string[],
  content: '',
  studentResponse: '',
  followUpPlan: '',
})

const methodOptions = [
  { label: '电话沟通', value: 'phone' },
  { label: '当面谈话', value: 'meeting' },
  { label: '发送消息', value: 'message' },
  { label: '其他', value: 'other' },
]

async function handleSave() {
  if (!form.interventionDate) {
    ElMessage.warning('请选择干预时间')
    return
  }
  if (!form.methods.length) {
    ElMessage.warning('请选择至少一种干预方式')
    return
  }
  if (!form.content.trim()) {
    ElMessage.warning('请填写干预内容')
    return
  }
  if (!form.studentResponse) {
    ElMessage.warning('请选择学生响应情况')
    return
  }
  if (!props.alertId) {
    ElMessage.error('缺少预警记录 ID')
    return
  }
  submitting.value = true
  try {
    await interveneAlert(props.alertId, {
      interventionDate: form.interventionDate,
      methods: form.methods,
      content: form.content,
      studentResponse: form.studentResponse,
      followUpPlan: form.followUpPlan,
    })
    ElMessage.success('干预记录已保存')
    resetForm()
    emit('saved')
  } catch (err: any) {
    ElMessage.error(err?.message || err?.response?.data?.message || '干预记录保存失败')
  } finally {
    submitting.value = false
  }
}

function resetForm() {
  form.interventionDate = ''
  form.methods = []
  form.content = ''
  form.studentResponse = ''
  form.followUpPlan = ''
}

defineExpose({ resetForm })
</script>

<template>
  <div class="intervention-form">
    <div class="form-section-title">干预回填</div>

    <div class="form-item">
      <label class="form-label required">干预时间</label>
      <el-date-picker
        v-model="form.interventionDate"
        type="date" placeholder="选择干预日期"
        format="YYYY-MM-DD" value-format="YYYY-MM-DD"
        style="width: 100%"
      />
    </div>

    <div class="form-item">
      <label class="form-label required">干预方式</label>
      <el-checkbox-group v-model="form.methods" class="methods-group">
        <el-checkbox v-for="opt in methodOptions" :key="opt.value" :value="opt.value">
          {{ opt.label }}
        </el-checkbox>
      </el-checkbox-group>
    </div>

    <div class="form-item">
      <label class="form-label required">干预内容</label>
      <el-input
        v-model="form.content"
        type="textarea" :rows="3"
        placeholder="描述本次干预的具体内容和情况..."
        maxlength="200" show-word-limit
      />
    </div>

    <div class="form-item">
      <label class="form-label required">学生响应情况</label>
      <el-radio-group v-model="form.studentResponse" class="response-group">
        <el-radio value="positive">积极配合</el-radio>
        <el-radio value="neutral">一般</el-radio>
        <el-radio value="resistant">抵触</el-radio>
      </el-radio-group>
    </div>

    <div class="form-item">
      <label class="form-label">后续跟进计划 <span class="optional">（可选）</span></label>
      <el-input
        v-model="form.followUpPlan"
        type="textarea" :rows="2"
        placeholder="后续跟进安排..."
        maxlength="200" show-word-limit
      />
    </div>

    <div class="form-actions">
      <el-button type="primary" :loading="submitting" @click="handleSave">保存干预记录</el-button>
    </div>
  </div>
</template>

<style scoped>
.intervention-form {
  padding: 4px 0;
}
.form-section-title {
  font-size: 14px;
  font-weight: 700;
  color: #1f2937;
  padding-bottom: 14px;
  border-bottom: 1px solid #f0f0f0;
  margin-bottom: 16px;
}
.form-item {
  margin-bottom: 18px;
}
.form-label {
  display: block;
  font-size: 12.5px;
  font-weight: 600;
  color: #374151;
  margin-bottom: 7px;
  letter-spacing: 0.02em;
}
.form-label.required::before {
  content: '*';
  color: #ef4444;
  margin-right: 3px;
}
.optional {
  font-weight: 400;
  color: #9ca3af;
  font-size: 11.5px;
}
.methods-group {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}
.response-group {
  display: flex;
  gap: 20px;
}
.form-actions {
  padding-top: 8px;
  border-top: 1px solid #f0f0f0;
  display: flex;
  justify-content: flex-end;
}
</style>
