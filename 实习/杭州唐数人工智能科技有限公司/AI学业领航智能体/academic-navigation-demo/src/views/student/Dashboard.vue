<script setup lang="ts">
import { reactive, ref, onMounted, onBeforeUnmount, watch, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import * as echarts from 'echarts'
import KpiCard from '@/components/common/KpiCard.vue'
import WarningBadge from '@/components/common/WarningBadge.vue'
import AiInsightPanel from '@/components/common/AiInsightPanel.vue'
import { getStudentDashboard, getStudentGrades, getStudentTerms, interpretMetrics } from '@/api'
import type { StudentDashboardData, StudentTermGradesData } from '@/types'

const data = reactive<StudentDashboardData>({
  gpa: 0,
  rank: 0,
  totalStudents: 0,
  classRank: 0,
  classTotalStudents: 0,
  credits: 0,
  totalCredits: 0,
  courseCount: 0,
  alertCount: 0,
  healthScore: 0,
  gpaTrend: [],
  recentAlerts: [],
})
const route = useRoute()
const router = useRouter()
const termGrades = reactive<StudentTermGradesData>({
  term: '',
  gpa: 0,
  avgGpa: null,
  rank: null,
  totalStudents: null,
  earnedCredits: 0,
  courseCount: 0,
  passedCount: 0,
  failedCount: 0,
  courses: [],
})
const loading = ref(false)
const errorMessage = ref('')
const chartRef = ref<HTMLDivElement>()
let chart: echarts.ECharts | null = null

// 学期选择
const selectedSemester = ref('')
const semesters = ref<string[]>([])

function formatSemester(term: string) {
  const parts = term.split('-')
  const semester = parts[2] === '2' ? '第二学期' : '第一学期'
  return parts.length === 3 ? `${parts[0]}-${parts[1]}学年 ${semester}` : term
}

async function loadTerms() {
  semesters.value = await getStudentTerms()
  const queryTerm = typeof route.query.term === 'string' ? route.query.term : ''
  selectedSemester.value = semesters.value.includes(queryTerm) ? queryTerm : (semesters.value[0] ?? '')
  if (!selectedSemester.value) await loadDashboard()
}

// AI 学业报告（由本地模型基于学生真实数据生成）
const aiLoading = ref(false)
const aiSummary = ref('正在基于学业数据生成 AI 报告...')
const aiPoints = ref<string[]>([])
const aiReportCache = new Map<string, { summary: string; points: string[] }>()

// AI 行动建议
const aiActions = [
  { label: '算法复习计划', color: '#2A4D99' },
  { label: '查看推荐资源', color: '#3db87e' },
  { label: '预约辅导老师', color: '#f09a4e' },
]

async function loadDashboard() {
  loading.value = true
  errorMessage.value = ''
  try {
    const [dashboard, grades] = await Promise.all([
      getStudentDashboard(selectedSemester.value || undefined),
      getStudentGrades(selectedSemester.value || undefined),
    ])
    Object.assign(data, dashboard)
    Object.assign(termGrades, grades)
    await nextTick()
    renderChart()
  } catch {
    errorMessage.value = '学业总览暂时无法加载，请检查网络后重试'
  } finally {
    loading.value = false
  }
  if (!errorMessage.value) loadAiReport()
}

async function loadAiReport() {
  const cacheKey = selectedSemester.value
  const cached = aiReportCache.get(cacheKey)
  if (cached) {
    aiSummary.value = cached.summary
    aiPoints.value = cached.points
    return
  }
  aiLoading.value = true
  try {
    const res = await interpretMetrics(
      '请基于以下学生学业总览数据生成一份简明的AI学业报告：先给出总体评价，再列出亮点、风险与行动建议。不要编造数据。',
      {
        gpa: data.gpa,
        rank: data.rank,
        totalStudents: data.totalStudents,
        credits: data.credits,
        totalCredits: data.totalCredits,
        alertCount: data.alertCount,
        healthScore: data.healthScore,
        gpaTrend: data.gpaTrend,
        recentAlerts: data.recentAlerts.map(a => ({
          level: a.level,
          title: a.title,
          description: a.description,
          status: a.status,
        })),
      },
    )
    const clean = (s: string) => s.replace(/[#*`]/g, '').trim()
    const lines = res.content.split('\n').map(clean).filter(Boolean)
    const bullets = lines.filter(l => /^[-*•]/.test(l)).map(l => l.replace(/^[-*•]\s*/, ''))
    if (bullets.length > 0) {
      aiSummary.value = lines.find(l => !/^[-*•]/.test(l)) || 'AI 学业报告'
      aiPoints.value = bullets.slice(0, 5)
    } else {
      aiSummary.value = res.content.replace(/[#*`]/g, '').trim()
      aiPoints.value = []
    }
    aiReportCache.set(cacheKey, { summary: aiSummary.value, points: aiPoints.value })
  } catch {
    aiSummary.value = 'AI 学业报告生成失败，请稍后重试'
    aiPoints.value = []
  } finally {
    aiLoading.value = false
  }
}

function renderChart() {
  if (!chartRef.value) return
  if (!chart) chart = echarts.init(chartRef.value)
  const values = [
    ...data.gpaTrend.map(d => d.gpa),
    ...data.gpaTrend.map(d => d.avg),
  ]
  const rawMin = values.length ? Math.min(...values) : 3
  const rawMax = values.length ? Math.max(...values) : 4
  chart.setOption({
    tooltip: {
      trigger: 'axis',
      backgroundColor: '#fff',
      borderColor: '#edf0f6',
      borderWidth: 1,
      textStyle: { color: '#1a2233', fontSize: 12 },
      extraCssText: 'box-shadow:0 8px 24px rgba(10,19,41,0.1);border-radius:10px;padding:10px 14px',
    },
    legend: {
      data: ['我的GPA', '年级均值'],
      bottom: 0,
      textStyle: { fontSize: 11.5, color: '#8a99b4' },
      itemWidth: 16, itemHeight: 2, icon: 'roundRect', itemGap: 20,
    },
    grid: { left: 44, right: 20, top: 16, bottom: 40 },
    xAxis: {
      type: 'category',
      data: data.gpaTrend.map(d => d.semester),
      axisLine: { lineStyle: { color: '#edf0f6' } },
      axisTick: { show: false },
      axisLabel: { color: '#9facc5', fontSize: 11.5 },
    },
    yAxis: {
      type: 'value',
      min: Math.max(0, Math.floor((rawMin - 0.2) * 2) / 2),
      max: Math.min(4.2, Math.ceil((rawMax + 0.2) * 2) / 2),
      splitLine: { lineStyle: { color: '#f2f4f9' } },
      axisLabel: { color: '#9facc5', fontSize: 11 },
      axisLine: { show: false },
    },
    series: [
      {
        name: '我的GPA', type: 'line', smooth: 0.5,
        data: data.gpaTrend.map(d => d.gpa),
        symbol: 'circle', symbolSize: 7,
        lineStyle: { color: '#2A4D99', width: 2.5 },
        itemStyle: { color: '#2A4D99', borderWidth: 2, borderColor: '#fff' },
        areaStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: 'rgba(42,77,153,0.14)' },
            { offset: 1, color: 'rgba(42,77,153,0)' },
          ]),
        },
      },
      {
        name: '年级均值', type: 'line', smooth: 0.5,
        data: data.gpaTrend.map(d => d.avg),
        symbol: 'circle', symbolSize: 4,
        lineStyle: { color: '#c8d0df', width: 1.5, type: 'dashed' },
        itemStyle: { color: '#c8d0df' },
      },
      {
        name: '我的GPA', type: 'bar', barWidth: 16,
        itemStyle: { color: 'rgba(42,77,153,0.14)', borderRadius: [4, 4, 0, 0] },
        data: data.gpaTrend.map(d => d.gpa),
        emphasis: { itemStyle: { color: 'rgba(42,77,153,0.24)' } },
      },
    ],
  })
}

onMounted(loadTerms)

watch(selectedSemester, async semester => {
  if (!semester) return
  await router.replace({ query: { ...route.query, term: semester } })
  loadDashboard()
})

onBeforeUnmount(() => { chart?.dispose() })
</script>

<template>
  <div v-loading="loading" class="dash">
    <div v-if="errorMessage" class="load-error" role="alert">
      <el-icon><WarningFilled /></el-icon>
      <span>{{ errorMessage }}</span>
      <el-button text type="primary" size="small" @click="loadDashboard">重新加载</el-button>
    </div>
    <div class="dash-header">
      <div>
        <h2 class="dash-title">学业总览</h2>
        <p class="dash-sub">实时掌握本学期 GPA、排名、学分与预警状态</p>
      </div>
      <el-select v-if="semesters.length > 1" v-model="selectedSemester" placeholder="选择学期" style="width: 230px">
        <el-option
          v-for="item in semesters"
          :key="item"
          :label="formatSemester(item)"
          :value="item"
        />
      </el-select>
      <span v-else-if="selectedSemester" class="single-term">{{ formatSemester(selectedSemester) }}</span>
    </div>

    <div class="kpi-grid">
      <KpiCard title="当前 GPA" :value="data.gpa" unit="/4.0" color="blue" icon="TrendCharts" />
      <KpiCard title="班级排名" :value="data.classRank" :unit="`/ ${data.classTotalStudents}`" color="green" icon="User" />
      <KpiCard title="专业排名" :value="data.rank" :unit="`/ ${data.totalStudents}`" color="orange" icon="Trophy" />
      <KpiCard title="体测得分" :value="data.healthScore" unit="分" color="blue" icon="Football" />
    </div>

    <!-- AI 学业报告 + GPA 趋势 双栏 -->
    <div class="ai-trend-grid">
      <AiInsightPanel
        title="AI 学业报告"
        :summary="aiSummary"
        :points="aiPoints"
        :actions="aiActions"
        :loading="aiLoading"
        accent="#2A4D99"
      />

      <div class="card">
        <div class="card-head">
          <span class="card-title">GPA 趋势</span>
          <span class="card-badge">截至所选学期</span>
        </div>
        <div ref="chartRef" class="trend-chart" />
      </div>
    </div>

    <div class="middle-grid">
      <div class="side-col" style="flex-direction: row; gap: 14px;">
        <div class="card side-card">
          <div class="side-label">本学期课程</div>
          <div class="side-num">{{ data.courseCount }}</div>
          <div class="side-desc">门课程有成绩记录</div>
        </div>
        <div class="card side-card">
          <div class="side-label">预警状态</div>
          <div v-if="data.alertCount === 0" class="status-ok">
            <span class="status-dot" />
            <span>学业正常</span>
          </div>
          <WarningBadge v-else level="yellow" />
        </div>
        <div class="card side-card">
          <div class="side-label">学分进度</div>
          <div class="side-num">{{ data.credits }}</div>
          <div class="side-desc">本学期已获 · 培养方案共 {{ data.totalCredits }} 学分</div>
        </div>
      </div>
    </div>

    <div class="card grade-card">
      <div class="grade-head">
        <div>
          <div class="card-title">本学期成绩</div>
          <div class="grade-sub">{{ formatSemester(termGrades.term) }}</div>
        </div>
        <div class="grade-summary">
          <span>已获学分 <strong>{{ termGrades.earnedCredits }}</strong></span>
          <span>通过 <strong class="passed-text">{{ termGrades.passedCount }}</strong></span>
          <span>未通过 <strong :class="{ 'failed-text': termGrades.failedCount > 0 }">{{ termGrades.failedCount }}</strong></span>
        </div>
      </div>
      <el-table v-if="termGrades.courses.length" :data="termGrades.courses" class="grade-table">
        <el-table-column prop="courseCode" label="课程代码" width="130" />
        <el-table-column prop="courseName" label="课程名称" min-width="190" />
        <el-table-column prop="credits" label="学分" width="90" align="center" />
        <el-table-column label="成绩" width="100" align="center">
          <template #default="scope">
            <strong :class="scope.row.score !== null && scope.row.score < 60 ? 'failed-text' : 'score-text'">
              {{ scope.row.score ?? '--' }}
            </strong>
          </template>
        </el-table-column>
        <el-table-column label="绩点" width="90" align="center">
          <template #default="scope">{{ scope.row.gradePoint ?? '--' }}</template>
        </el-table-column>
        <el-table-column label="状态" width="100" align="center">
          <template #default="scope">
            <span v-if="scope.row.status === 'failed' || (scope.row.score !== null && scope.row.score < 60)" class="grade-status failed">未通过</span>
            <span v-else-if="scope.row.status === 'pending'" class="grade-status pending">待录入</span>
            <span v-else class="grade-status passed">已通过</span>
          </template>
        </el-table-column>
      </el-table>
      <div v-else class="empty-grade">该学期暂无课程成绩</div>
    </div>

    <div class="card">
      <div class="card-head">
        <span class="card-title">近期通知</span>
      </div>
      <div v-if="data.recentAlerts.length === 0" class="empty">
        <div class="empty-icon">
          <el-icon><CircleCheckFilled /></el-icon>
        </div>
        <div class="empty-title">暂无预警通知</div>
        <div class="empty-desc">保持当前良好的学习状态，继续加油</div>
      </div>
      <div v-else class="recent-alert-list">
        <div v-for="alert in data.recentAlerts" :key="alert.id" class="recent-alert-item">
          <WarningBadge :level="alert.level" />
          <div class="recent-alert-main">
            <span class="recent-alert-title">{{ alert.title }}</span>
            <span class="recent-alert-desc">{{ alert.description }}</span>
          </div>
          <span class="recent-alert-date">{{ alert.pushedAt || alert.date }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.dash { padding: 28px 32px; display: flex; flex-direction: column; gap: 20px; }
.load-error {
  display: flex; align-items: center; gap: 10px; padding: 11px 14px;
  border: 1px solid #f6c8c3; border-radius: 10px; background: #fff7f6;
  color: #a63026; font-size: 13px;
}
.load-error .el-button { margin-left: auto; }

.dash-header { display: flex; align-items: center; justify-content: space-between; }
.dash-title { font-size: 20px; font-weight: 800; color: #101d3e; margin: 0; letter-spacing: -0.01em; }
.dash-sub { font-size: 13px; color: #9facc5; margin: 4px 0 0; }
.single-term {
  min-width: 210px; padding: 8px 12px; text-align: center;
  border: 1px solid #dcdfe6; border-radius: 4px;
  color: #606266; background: #fff; font-size: 14px;
}

.kpi-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 14px; }

.ai-trend-grid { display: grid; grid-template-columns: 340px 1fr; gap: 14px; }

.middle-grid { display: grid; grid-template-columns: 1fr; gap: 14px; }

.side-col { display: flex; flex-direction: column; gap: 14px; }
.side-card { flex: 1; }

.card {
  background: #fff;
  border-radius: 16px;
  padding: 20px 22px;
  border: 1px solid rgba(10,19,41,0.07);
  box-shadow: 0 1px 2px rgba(10,19,41,0.03), 0 4px 16px rgba(10,19,41,0.05);
}

.card-head { display: flex; align-items: center; gap: 8px; margin-bottom: 16px; }
.card-title { font-size: 14px; font-weight: 700; color: #111827; }
.trend-chart { height: 200px; }
.card-badge {
  font-size: 11px; font-weight: 600; color: #9facc5;
  background: #f2f4f9; padding: 2px 8px; border-radius: 10px;
}

.grade-card { padding-bottom: 12px; }
.grade-head { display: flex; align-items: center; justify-content: space-between; gap: 20px; margin-bottom: 14px; }
.grade-sub { margin-top: 5px; color: #9facc5; font-size: 12px; }
.grade-summary { display: flex; align-items: center; gap: 22px; color: #7b879d; font-size: 12.5px; }
.grade-summary strong { margin-left: 4px; color: #25324a; font-size: 14px; }
.passed-text { color: #2d9c68 !important; }
.failed-text { color: #d95858 !important; }
.score-text { color: #25324a; }
.grade-status { display: inline-flex; min-width: 54px; justify-content: center; padding: 3px 8px; border-radius: 4px; font-size: 12px; font-weight: 600; }
.grade-status.passed { color: #27845a; background: #eaf8f1; }
.grade-status.failed { color: #c74848; background: #fff0f0; }
.grade-status.pending { color: #9a6a22; background: #fff7e7; }
.empty-grade { padding: 34px 0; text-align: center; color: #9facc5; font-size: 13px; }
:deep(.grade-table .el-table__header th) { background: #f8f9fc; color: #7b879d; font-weight: 600; }
:deep(.grade-table .el-table__cell) { padding: 10px 0; }

.side-label { font-size: 12px; font-weight: 600; color: #9facc5; letter-spacing: 0.02em; margin-bottom: 10px; }
.side-num { font-size: 34px; font-weight: 800; color: #2A4D99; line-height: 1; margin-bottom: 4px; letter-spacing: -0.02em; }
.side-desc { font-size: 12px; color: #b4bed2; }

.status-ok {
  display: flex; align-items: center; gap: 8px;
  font-size: 13.5px; font-weight: 600; color: #3db87e; margin-top: 4px;
}
.status-dot {
  width: 8px; height: 8px; border-radius: 50%; background: #3db87e;
  box-shadow: 0 0 0 3px rgba(61,184,126,0.15);
  animation: pulse 2s infinite;
}
@keyframes pulse {
  0%, 100% { box-shadow: 0 0 0 3px rgba(61,184,126,0.15); }
  50% { box-shadow: 0 0 0 6px rgba(61,184,126,0); }
}

.empty { display: flex; flex-direction: column; align-items: center; padding: 28px 0; gap: 8px; }
.empty-icon {
  width: 44px; height: 44px; border-radius: 50%;
  background: #ecfbf3; display: flex; align-items: center;
  justify-content: center; color: #3db87e; font-size: 22px;
}
.empty-title { font-size: 14px; font-weight: 700; color: #3e4759; }
.empty-desc { font-size: 12.5px; color: #9facc5; }

.recent-alert-list { display: flex; flex-direction: column; gap: 8px; }
.recent-alert-item {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 12px 14px;
  border: 1px solid #edf0f6;
  border-radius: 10px;
  background: #fafbfd;
}
.recent-alert-main { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 3px; }
.recent-alert-title { font-size: 13.5px; font-weight: 600; color: #1f2937; }
.recent-alert-desc {
  font-size: 12.5px;
  color: #6b7280;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.recent-alert-date { font-size: 12px; color: #9ca3af; white-space: nowrap; }

@media (max-width: 1080px) {
  .dash { padding: 24px; }
  .kpi-grid { grid-template-columns: repeat(2, 1fr); }
  .ai-trend-grid { grid-template-columns: 1fr; }
  .ai-trend-grid :deep(.ai-insight-panel) { min-height: auto; }
}

@media (max-width: 760px) {
  .dash { padding: 20px 16px 28px; gap: 16px; }
  .dash-header { align-items: flex-start; flex-direction: column; gap: 12px; }
  .dash-title { font-size: 19px; }
  .dash-sub { line-height: 1.5; }
  .dash-header .el-select, .single-term { width: 100% !important; min-width: 0; }
  .kpi-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 10px; }
  .side-col { flex-direction: column !important; gap: 10px !important; }
  .card { border-radius: 12px; padding: 16px; }
  .trend-chart { height: 220px; }
  .grade-head { align-items: flex-start; flex-direction: column; gap: 12px; }
  .grade-summary { width: 100%; justify-content: space-between; gap: 8px; }
  .grade-card { overflow: hidden; }
  .grade-table { min-width: 600px; }
  .grade-card :deep(.el-table) { overflow-x: auto; }
  .recent-alert-item { padding: 11px 12px; gap: 9px; }
  .recent-alert-desc { white-space: normal; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; }
  .recent-alert-date { font-size: 11px; }
}

@media (max-width: 420px) {
  .kpi-grid { grid-template-columns: 1fr; }
  .grade-summary { align-items: flex-start; flex-direction: column; }
  .trend-chart { height: 200px; }
}
</style>
