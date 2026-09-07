<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import type { AlertItem } from '@/types'
import WarningBadge from '@/components/common/WarningBadge.vue'
import InterventionForm from '@/components/common/InterventionForm.vue'
import AiInsightPanel from '@/components/common/AiInsightPanel.vue'
import { getAlerts, getAlertDetail, getTeacherTerms, interpretMetrics } from '@/api'
import { useAlertStore } from '@/stores/alerts'

const alertStore = useAlertStore()
const route = useRoute()
const router = useRouter()
const alertList = ref<AlertItem[]>([])
const loading = ref(false)
const errorMessage = ref('')
const filterLevel = ref<string>('all')
const filterStatus = ref<string>('all')
const keyword = ref('')
const page = ref(1)
const size = 10
const total = ref(0)
const termOptions = ref<string[]>([])
const selectedTerm = ref('')
const className = ref('')
const drawerVisible = ref(false)
const selectedAlert = ref<AlertItem | null>(null)
const interventionFormRef = ref<InstanceType<typeof InterventionForm> | null>(null)
const aiAnalysisLoading = ref(false)
let latestRequestId = 0

async function loadAlerts(requestedPage = page.value) {
  const requestId = ++latestRequestId
  loading.value = true
  errorMessage.value = ''
  try {
    const res = await getAlerts({
      term: selectedTerm.value,
      level: filterLevel.value,
      status: filterStatus.value,
      keyword: keyword.value,
      page: requestedPage,
      size,
    })

    // Ignore responses from an older filter/page request.
    if (requestId !== latestRequestId) return

    const lastPage = Math.max(1, Math.ceil(res.total / size))
    if (requestedPage > lastPage) {
      page.value = lastPage
      return loadAlerts(lastPage)
    }

    page.value = requestedPage
    alertList.value = res.list.map(a => ({ ...a, failedCourses: a.failedCourses ?? [] }))
    total.value = res.total
  } catch {
    if (requestId !== latestRequestId) return
    alertList.value = []
    total.value = 0
    errorMessage.value = '预警数据加载失败，请稍后重试'
  } finally {
    if (requestId === latestRequestId) loading.value = false
  }
}

function formatTerm(term: string) {
  const parts = term.split('-')
  if (parts.length !== 3) return term
  return `${parts[0]}-${parts[1]} 学年${parts[2] === '2' ? '第二' : '第一'}学期`
}

async function initialize() {
  loading.value = true
  try {
    const context = await getTeacherTerms()
    termOptions.value = context.terms
    className.value = context.className
    const queryTerm = typeof route.query.term === 'string' ? route.query.term : ''
    selectedTerm.value = context.terms.includes(queryTerm) ? queryTerm : context.latestTerm
  } catch {
    errorMessage.value = '学期数据加载失败，请稍后重试'
    loading.value = false
  }
}

function search() {
  page.value = 1
  void loadAlerts(1)
}

function handlePageChange(next: number) {
  page.value = next
  void loadAlerts(next)
}

// AI 根因分析（本地模型基于预警真实数据生成）
const aiAnalysisSummary = ref('')
const aiAnalysisPoints = ref<string[]>([])
const aiAnalysisCache = new Map<string, { summary: string; points: string[] }>()

async function loadAiAnalysis(alert: AlertItem) {
  if (!alert) return
  const cacheKey = alert.id
  const cached = aiAnalysisCache.get(cacheKey)
  if (cached) {
    aiAnalysisSummary.value = cached.summary
    aiAnalysisPoints.value = cached.points
    return
  }
  aiAnalysisLoading.value = true
  try {
    const res = await interpretMetrics(
      '请基于以下预警记录生成简明的AI根因分析与面谈建议：第一行输出一句总体判断，后续每行用“- ”开头输出分析或建议，最多5条；不要出现#、**、反引号，不要编造数据。',
      {
        level: alert.level,
        type: alert.type,
        title: alert.title,
        description: alert.description,
        course: alert.course,
        failedCourses: alert.failedCourses ?? [],
        status: alert.status,
        suggestion: alert.suggestion,
      },
    )
    const clean = (s: string) => s.replace(/[#*`]/g, '').trim()
    const lines = res.content.split('\n').map(clean).filter(Boolean)
    const bullets = lines.filter(l => /^[-•]/.test(l)).map(l => l.replace(/^[-•]\s*/, ''))
    if (bullets.length > 0) {
      aiAnalysisSummary.value = lines.find(l => !/^[-•]/.test(l)) || 'AI 根因分析'
      aiAnalysisPoints.value = bullets.slice(0, 5)
    } else {
      aiAnalysisSummary.value = res.content.replace(/[#*`]/g, '').trim()
      aiAnalysisPoints.value = []
    }
    aiAnalysisCache.set(cacheKey, { summary: aiAnalysisSummary.value, points: aiAnalysisPoints.value })
  } catch {
    aiAnalysisSummary.value = 'AI 根因分析生成失败，请稍后重试'
    aiAnalysisPoints.value = []
  } finally {
    aiAnalysisLoading.value = false
  }
}

async function openIntervention(alert: AlertItem) {
  selectedAlert.value = alert
  drawerVisible.value = true
  try {
    const detail = await getAlertDetail(alert.id)
    selectedAlert.value = { ...detail, failedCourses: detail.failedCourses ?? [] }
  } finally {
    loadAiAnalysis(selectedAlert.value)
  }
}

function onInterventionSaved() {
  drawerVisible.value = false
  void loadAlerts(page.value)
  alertStore.refreshUnread()
}

function onDrawerClose() {
  interventionFormRef.value?.resetForm()
}

const statusConfig = {
  pending: { label: '待处理', type: 'warning' as const },
  processing: { label: '处理中', type: 'primary' as const },
  resolved: { label: '已解决', type: 'success' as const },
}

function getStatusConfig(status: string) {
  return statusConfig[status as keyof typeof statusConfig] || {
    label: '未知状态',
    type: 'info' as const,
  }
}

const accentMap: Record<string, string> = {
  red: '#e04538', orange: '#f09a4e', yellow: '#eab308',
}

onMounted(initialize)

watch(selectedTerm, async term => {
  if (!term) return
  page.value = 1
  await router.replace({ query: { ...route.query, term } })
  void loadAlerts(1)
})
</script>

<template>
  <div v-loading="loading" class="alert-page p-6 flex flex-col gap-6">
    <div class="page-heading">
      <div>
        <h2 class="text-xl font-bold" style="color: #101d3e">预警管理</h2>
        <p class="text-sm mt-1" style="color: #9facc5">{{ formatTerm(selectedTerm) }} · {{ className }} · 共 {{ total }} 条预警</p>
      </div>
      <div class="term-selector">
        <span class="term-selector-label">学期</span>
        <select v-if="termOptions.length > 1" v-model="selectedTerm" class="term-select">
          <option v-for="term in termOptions" :key="term" :value="term">{{ term }}</option>
        </select>
        <span v-else class="term-static">{{ selectedTerm }}</span>
      </div>
    </div>

    <!-- Filters -->
    <div class="filter-bar">
      <div class="filter-control">
        <span class="filter-label">预警等级</span>
        <el-select v-model="filterLevel" aria-label="预警等级" style="width: 150px" @change="search">
          <el-option label="全部等级" value="all" />
          <el-option label="黄色预警" value="yellow" />
          <el-option label="橙色预警" value="orange" />
          <el-option label="红色预警" value="red" />
        </el-select>
      </div>
      <div class="filter-control">
        <span class="filter-label">处理状态</span>
        <el-select v-model="filterStatus" aria-label="处理状态" style="width: 150px" @change="search">
          <el-option label="全部状态" value="all" />
          <el-option label="待处理" value="pending" />
          <el-option label="处理中" value="processing" />
          <el-option label="已解决" value="resolved" />
        </el-select>
      </div>
      <div class="filter-control filter-keyword">
        <span class="filter-label">关键词</span>
        <el-input
          v-model="keyword"
          placeholder="学生姓名 / 学号"
          clearable
          @keyup.enter="search"
          @clear="search"
        />
      </div>
      <el-button class="filter-submit" type="primary" @click="search">查询</el-button>
    </div>

    <!-- Table -->
    <div class="table-wrap">
      <el-table :data="alertList" height="100%" stripe>
        <template #empty>
          <div class="table-empty">
            <span v-if="errorMessage">{{ errorMessage }}</span>
            <span v-else>暂无符合条件的预警记录</span>
            <el-button v-if="errorMessage" text type="primary" size="small" @click="loadAlerts(page)">重试</el-button>
          </div>
        </template>
        <el-table-column label="学生姓名" prop="studentName" width="100" />
        <el-table-column label="学号" prop="studentId" width="110" />
        <el-table-column label="预警类型" prop="type" width="110" />
        <el-table-column label="预警等级" width="130">
          <template #default="{ row }">
            <WarningBadge :level="row.level" />
          </template>
        </el-table-column>
        <el-table-column label="预警内容" min-width="220">
          <template #default="{ row }">
            <div class="alert-content-cell">
              <span class="alert-title-text">{{ row.title }}</span>
              <div v-if="row.failedCourses?.length" class="failed-course-tags">
                <span
                  v-for="c in row.failedCourses"
                  :key="c"
                  class="course-tag"
                >{{ c }}</span>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="触发事件" prop="triggerEvent" width="160" />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusConfig(row.status).type" size="small">
              {{ getStatusConfig(row.status).label }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="140">
          <template #default="{ row }">
            <el-button text type="primary" size="small" @click="openIntervention(row)">干预</el-button>
            <router-link :to="{ path: `/teacher/student/${row.studentId}`, query: { term: selectedTerm } }">
              <el-button text size="small">画像</el-button>
            </router-link>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <div v-if="total > size" class="pagination-row">
      <el-pagination
        v-model:current-page="page"
        :total="total"
        :page-size="size"
        :pager-count="5"
        layout="prev, pager, next"
        prev-text="上一页"
        next-text="下一页"
        @current-change="handlePageChange"
      />
    </div>

    <!-- Intervention drawer -->
    <el-drawer
      v-model="drawerVisible"
      title="干预处理"
      size="500px"
      @close="onDrawerClose"
    >
      <div v-if="selectedAlert" class="drawer-content">
        <!-- Student + alert info -->
        <div class="alert-summary" :class="selectedAlert.level">
          <div class="alert-summary-left">
            <div class="student-name">{{ selectedAlert.studentName }}</div>
            <div class="student-meta">{{ selectedAlert.studentId }} · {{ selectedAlert.type }}</div>
          </div>
          <WarningBadge :level="selectedAlert.level" />
        </div>

        <el-descriptions :column="1" border size="small" class="mb-4">
          <el-descriptions-item label="预警内容">{{ selectedAlert.title }}</el-descriptions-item>
          <el-descriptions-item label="详细描述">{{ selectedAlert.description }}</el-descriptions-item>
          <el-descriptions-item label="触发事件">{{ selectedAlert.triggerEvent || '—' }}</el-descriptions-item>
          <el-descriptions-item label="推送时间">{{ selectedAlert.pushedAt || selectedAlert.date }}</el-descriptions-item>
          <el-descriptions-item v-if="selectedAlert.suggestion" label="处置建议">{{ selectedAlert.suggestion }}</el-descriptions-item>
        </el-descriptions>

        <!-- AI 根因分析 -->
        <AiInsightPanel
          title="AI 根因分析 & 面谈建议"
          :summary="aiAnalysisSummary"
          :points="aiAnalysisPoints"
          :loading="aiAnalysisLoading"
          :accent="accentMap[selectedAlert.level] || '#2A4D99'"
          compact
        />

        <el-divider />

        <InterventionForm
          ref="interventionFormRef"
          :alert-id="selectedAlert.id"
          @saved="onInterventionSaved"
        />
      </div>
    </el-drawer>
  </div>
</template>

<style scoped>
.alert-page { min-height: calc(100vh - 48px); box-sizing: border-box; }
.page-heading { display: flex; align-items: flex-start; justify-content: space-between; }
.term-selector { display: flex; align-items: center; gap: 8px; }
.term-selector-label { font-size: 12px; color: #94a3b8; font-weight: 600; }
.term-select {
  height: 34px; padding: 0 30px 0 11px; border: 1px solid #dbe3ee; border-radius: 8px;
  background: #fff; color: #334155; font-size: 13px; font-weight: 600; cursor: pointer;
}
.term-select:focus { outline: none; border-color: #2a4d99; box-shadow: 0 0 0 3px rgba(42,77,153,0.1); }
.term-static { font-size: 13px; color: #334155; font-weight: 600; }
.filter-bar {
  display: flex; align-items: flex-end; gap: 14px; flex-wrap: wrap;
  padding: 16px 18px; background: #fff; border: 1px solid #e6ebf3;
  border-radius: 12px; box-shadow: 0 4px 14px rgba(31, 55, 96, 0.04);
}
.filter-control { display: flex; flex-direction: column; gap: 6px; }
.filter-label { color: #64748b; font-size: 12px; line-height: 1; font-weight: 600; }
.filter-keyword { width: 220px; }
.filter-submit { min-width: 76px; height: 32px; font-weight: 600; }

.table-wrap {
  flex: 1; min-height: 520px; height: clamp(520px, calc(100vh - 330px), 1080px);
  background: #fff; border: 1px solid #e6ebf3; border-radius: 12px; overflow: hidden;
  box-shadow: 0 4px 16px rgba(31, 55, 96, 0.04);
}
.table-empty { display: flex; align-items: center; justify-content: center; gap: 12px; min-height: 120px; color: #9facc5; font-size: 13px; }
.pagination-row {
  display: flex; justify-content: flex-end; align-items: center;
  min-height: 40px; padding: 2px 2px 0;
}
.pagination-row :deep(.el-pagination) {
  width: 360px; min-height: 32px; justify-content: flex-end; white-space: nowrap;
}
.pagination-row :deep(.btn-prev),
.pagination-row :deep(.btn-next),
.pagination-row :deep(.el-pager li) {
  min-width: 32px; height: 32px; line-height: 32px; border-radius: 6px;
}
.pagination-row :deep(.btn-prev), .pagination-row :deep(.btn-next) { padding: 0 8px; }

.table-wrap :deep(.el-table) { --el-table-header-bg-color: #f8fafc; --el-table-border-color: #edf1f6; color: #334155; }
.table-wrap :deep(.el-table th.el-table__cell) { height: 52px; color: #64748b; font-size: 12px; font-weight: 700; }
.table-wrap :deep(.el-table td.el-table__cell) { padding: 14px 0; }
.table-wrap :deep(.el-table .cell) { line-height: 1.5; }
.table-wrap :deep(.el-table__body-wrapper) { scrollbar-color: #cbd5e1 transparent; }
.table-wrap :deep(.el-table__body tr:hover > td.el-table__cell) { background: #f8fbff; }
.table-wrap :deep(.el-button.is-text) { font-weight: 600; }

.alert-content-cell { display: flex; flex-direction: column; gap: 6px; padding: 4px 0; }
.alert-title-text { font-size: 13px; color: #1f2937; font-weight: 500; }
.failed-course-tags { display: flex; flex-wrap: wrap; gap: 4px; }
.course-tag {
  display: inline-block;
  font-size: 11px;
  color: #b45309;
  background: #fef3c7;
  border: 1px solid #fde68a;
  padding: 1px 7px;
  border-radius: 4px;
  white-space: nowrap;
}

.drawer-content { display: flex; flex-direction: column; gap: 16px; }

.alert-summary {
  display: flex; align-items: center; justify-content: space-between;
  padding: 14px 16px; border-radius: 12px; border: 1px solid;
}
.alert-summary.yellow { background: #fffbeb; border-color: #fde68a; }
.alert-summary.orange { background: #fff7ed; border-color: #fed7aa; }
.alert-summary.red { background: #fef2f2; border-color: #fecaca; }

.student-name { font-size: 15px; font-weight: 700; color: #1f2937; }
.student-meta { font-size: 12px; color: #9ca3af; margin-top: 2px; }

.mb-4 { margin-bottom: 0; }

@media (max-width: 900px) {
  .alert-page { min-height: calc(100vh - 32px); }
  .filter-bar { align-items: stretch; }
  .filter-control, .filter-keyword { width: 100%; }
  .filter-control :deep(.el-select), .filter-keyword :deep(.el-input) { width: 100% !important; }
  .filter-submit { width: 100%; }
  .table-wrap { height: 560px; }
  .pagination-row :deep(.el-pagination) { width: 100%; }
}
</style>
