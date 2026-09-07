<script setup lang="ts">
import { reactive, computed, ref, onMounted, onBeforeUnmount, watch, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import * as echarts from 'echarts'
import KpiCard from '@/components/common/KpiCard.vue'
import WarningBadge from '@/components/common/WarningBadge.vue'
import RadarChart from '@/components/common/RadarChart.vue'
import AiInsightPanel from '@/components/common/AiInsightPanel.vue'
import { getTeacherDashboard, getTeacherTerms, interpretMetrics } from '@/api'
import type { TeacherDashboardData } from '@/types'

const data = reactive<TeacherDashboardData>({
  totalStudents: 0,
  avgGpa: 0,
  alertCount: 0,
  passRate: 0,
  alertDistribution: { yellow: 0, orange: 0, red: 0 },
  dimensionAvg: [],
  recentAlerts: [],
  gradeDistribution: [],
})
const route = useRoute()
const router = useRouter()
const loading = ref(false)
const pieRef = ref<HTMLDivElement>()
let pieChart: echarts.ECharts | null = null

const termOptions = ref<string[]>([])
const selectedTerm = ref('')
const className = ref('')

function formatTerm(term: string) {
  const parts = term.split('-')
  if (parts.length !== 3) return term
  return `${parts[0]}-${parts[1]} 学年${parts[2] === '2' ? '第二' : '第一'}学期`
}

const gradeColors = ['#2A4D99', '#3db87e', '#f09a4e', '#e04538']

// AI 班级诊断（本地模型基于真实班级数据生成）
const aiClassLoading = ref(false)
const aiClassSummary = ref('正在基于班级数据生成 AI 诊断...')
const aiClassPoints = ref<string[]>([])
const aiClassCache = new Map<string, { summary: string; points: string[] }>()

async function loadAiClass() {
  const cacheKey = selectedTerm.value
  const cached = aiClassCache.get(cacheKey)
  if (cached) {
    aiClassSummary.value = cached.summary
    aiClassPoints.value = cached.points
    return
  }
  aiClassLoading.value = true
  try {
    const res = await interpretMetrics(
      '请基于以下班级学情数据生成简明的AI班级诊断：第一行输出一句总体评价，后续每行用“- ”开头输出分析或建议，最多5条；不要出现#、**、反引号，不要编造数据。',
      {
        totalStudents: data.totalStudents,
        avgGpa: data.avgGpa,
        alertCount: data.alertCount,
        passRate: data.passRate,
        alertDistribution: data.alertDistribution,
        dimensionAvg: data.dimensionAvg,
        gradeDistribution: data.gradeDistribution,
        recentAlerts: data.recentAlerts.map(a => ({ id: a.id, level: a.level, title: a.title, status: a.status })),
      },
    )
    const clean = (s: string) => s.replace(/[#*`]/g, '').trim()
    const lines = res.content.split('\n').map(clean).filter(Boolean)
    const bullets = lines.filter(l => /^[-•]/.test(l)).map(l => l.replace(/^[-•]\s*/, ''))
    if (bullets.length > 0) {
      aiClassSummary.value = lines.find(l => !/^[-•]/.test(l)) || 'AI 班级诊断'
      aiClassPoints.value = bullets.slice(0, 5)
    } else {
      aiClassSummary.value = res.content.replace(/[#*`]/g, '').trim()
      aiClassPoints.value = []
    }
    aiClassCache.set(cacheKey, { summary: aiClassSummary.value, points: aiClassPoints.value })
  } catch {
    aiClassSummary.value = 'AI 班级诊断生成失败，请稍后重试'
    aiClassPoints.value = []
  } finally {
    aiClassLoading.value = false
  }
}

async function loadDashboard() {
  if (!selectedTerm.value) return
  loading.value = true
  try {
    Object.assign(data, await getTeacherDashboard(selectedTerm.value))
    await nextTick()
    renderPie()
  } finally {
    loading.value = false
  }
  loadAiClass()
}

async function initialize() {
  loading.value = true
  try {
    const context = await getTeacherTerms()
    termOptions.value = context.terms
    className.value = context.className
    const queryTerm = typeof route.query.term === 'string' ? route.query.term : ''
    selectedTerm.value = context.terms.includes(queryTerm) ? queryTerm : context.latestTerm
  } finally {
    loading.value = false
  }
}

function renderPie() {
  if (!pieRef.value) return
  if (!pieChart) pieChart = echarts.init(pieRef.value)
  pieChart.setOption({
    tooltip: { trigger: 'item', formatter: '{b}: {c}人 ({d}%)' },
    legend: {
      orient: 'vertical', right: '4%', top: 'middle',
      textStyle: { fontSize: 12, color: '#6b7280' },
      itemWidth: 10, itemHeight: 10,
      itemGap: 10,
      formatter: (name: string) => {
        const item = data.gradeDistribution.find(d => d.name === name)
        return item ? `${name}  ${item.value}人` : name
      },
    },
    color: gradeColors,
    series: [{
      type: 'pie',
      radius: ['44%', '68%'],
      center: ['28%', '50%'],
      data: data.gradeDistribution,
      label: { show: false },
      emphasis: { itemStyle: { shadowBlur: 8, shadowColor: 'rgba(0,0,0,0.12)' } },
    }],
  })
}

onMounted(initialize)

watch(selectedTerm, async term => {
  if (!term) return
  await router.replace({ query: { ...route.query, term } })
  loadDashboard()
})

onBeforeUnmount(() => { pieChart?.dispose() })

const radarDims = computed(() => data.dimensionAvg.map(d => ({ label: d.label, score: d.value, avgScore: d.value - 5 })))

const alertItems = computed(() => [
  { key: 'yellow', label: '黄色预警', count: data.alertDistribution.yellow, bg: '#fefce8', border: '#fde68a', text: '#854d0e', dot: '#eab308' },
  { key: 'orange', label: '橙色预警', count: data.alertDistribution.orange, bg: '#fff7ed', border: '#fed7aa', text: '#9a3412', dot: '#f09a4e' },
  { key: 'red',    label: '红色预警', count: data.alertDistribution.red,    bg: '#fef2f2', border: '#fecaca', text: '#991b1b', dot: '#e04538' },
])
</script>

<template>
  <div v-loading="loading" class="page">
    <div class="page-head">
      <div>
        <h2 class="page-title">班级驾驶舱</h2>
        <p class="page-sub">{{ formatTerm(selectedTerm) }} · {{ className }}</p>
      </div>
      <div class="term-selector">
        <span class="term-selector-label">学期</span>
        <select v-if="termOptions.length > 1" v-model="selectedTerm" class="term-select">
          <option v-for="t in termOptions" :key="t" :value="t">{{ t }}</option>
        </select>
        <span v-else class="term-static">{{ selectedTerm }}</span>
      </div>
    </div>

    <div class="kpi-grid">
      <KpiCard title="班级人数" :value="data.totalStudents" unit="人" color="blue" icon="User" />
      <router-link :to="{ path: '/teacher/gpa-progress', query: { term: selectedTerm } }" class="kpi-link">
        <KpiCard title="班级均 GPA" :value="data.avgGpa" unit="/4.0" color="green" icon="TrendCharts" />
      </router-link>
      <router-link :to="{ path: '/teacher/alerts', query: { term: selectedTerm } }" class="kpi-link">
        <KpiCard title="当前预警" :value="data.alertCount" unit="人" color="orange" icon="Warning" />
      </router-link>
      <router-link :to="{ path: '/teacher/course-stats', query: { term: selectedTerm } }" class="kpi-link">
        <KpiCard title="课程通过率" :value="data.passRate" unit="%" color="blue" icon="CircleCheckFilled" />
      </router-link>
    </div>

    <!-- AI 班级诊断 -->
    <AiInsightPanel
      title="AI 班级诊断报告"
      :summary="aiClassSummary"
      :points="aiClassPoints"
      :actions="[{ label: '查看高风险学生', color: '#e04538' }, { label: '生成班级周报', color: '#2A4D99' }, { label: '一键批量通知', color: '#3db87e' }]"
      :loading="aiClassLoading"
      accent="#e04538"
    />

    <div class="mid-grid">
      <div class="card">
        <div class="card-title">班级五维画像</div>
        <RadarChart :dimensions="radarDims" />
      </div>
      <div class="card">
        <div class="card-title">成绩分布</div>
        <div ref="pieRef" style="height: 220px" />
      </div>
      <div class="card">
        <div class="card-title">预警分布</div>
        <div class="alert-bars">
          <div
            v-for="item in alertItems" :key="item.key"
            class="alert-bar"
            :style="{ background: item.bg, border: `1px solid ${item.border}` }"
          >
            <div class="alert-bar-left">
              <span class="alert-dot" :style="{ background: item.dot }" />
              <span class="alert-label" :style="{ color: item.text }">{{ item.label }}</span>
            </div>
            <span class="alert-count" :style="{ color: item.text }">{{ item.count }}</span>
          </div>
        </div>
      </div>
    </div>

    <div class="card table-card">
      <div class="table-head">
        <span class="card-title">本学期预警列表</span>
        <router-link :to="{ path: '/teacher/alerts', query: { term: selectedTerm } }" class="view-all">查看全部</router-link>
      </div>
      <el-table :data="data.recentAlerts" stripe>
        <el-table-column label="学生" prop="studentName" width="90" />
        <el-table-column label="学号" prop="studentId" width="120" />
        <el-table-column label="预警等级" width="130">
          <template #default="{ row }"><WarningBadge :level="row.level" /></template>
        </el-table-column>
        <el-table-column label="预警内容" prop="title" />
        <el-table-column label="日期" prop="date" width="120" />
        <el-table-column label="操作" width="100">
          <template #default="{ row }">
            <router-link :to="{ path: `/teacher/student/${row.studentId}`, query: { term: selectedTerm } }">
              <el-button text type="primary" size="small">查看</el-button>
            </router-link>
          </template>
        </el-table-column>
      </el-table>
    </div>
  </div>
</template>

<style scoped>
.page { padding: 28px 32px; display: flex; flex-direction: column; gap: 20px; }
.page-head { display: flex; align-items: flex-start; justify-content: space-between; }
.page-title { font-size: 20px; font-weight: 800; color: #101d3e; margin: 0; letter-spacing: -0.01em; }
.page-sub { font-size: 13px; color: #9facc5; margin: 4px 0 0; }
.term-static { font-size: 13px; font-weight: 600; color: #374151; }

.kpi-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 14px; }

.mid-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 14px; }

.card {
  background: #fff; border-radius: 16px; padding: 20px 22px;
  border: 1px solid rgba(10,19,41,0.07);
  box-shadow: 0 1px 2px rgba(10,19,41,0.03), 0 4px 16px rgba(10,19,41,0.05);
}

.card-title { font-size: 14px; font-weight: 700; color: #111827; margin-bottom: 16px; }

.alert-bars { display: flex; flex-direction: column; gap: 10px; }
.alert-bar {
  display: flex; align-items: center; justify-content: space-between;
  padding: 12px 14px; border-radius: 10px;
}
.alert-bar-left { display: flex; align-items: center; gap: 8px; }
.alert-dot { width: 8px; height: 8px; border-radius: 50%; flex-shrink: 0; }
.alert-label { font-size: 13px; font-weight: 600; }
.alert-count { font-size: 22px; font-weight: 800; font-variant-numeric: tabular-nums; }

.table-card { padding: 0; overflow: hidden; }
.table-head {
  display: flex; align-items: center; justify-content: space-between;
  padding: 16px 22px; border-bottom: 1px solid #f2f4f9;
}
.view-all { font-size: 12.5px; color: #2A4D99; text-decoration: none; font-weight: 600; }
.view-all:hover { text-decoration: underline; }

.kpi-link { text-decoration: none; display: block; }

.term-selector { display: flex; align-items: center; gap: 8px; flex-shrink: 0; }
.term-selector-label { font-size: 12.5px; color: #9facc5; font-weight: 600; }
.term-select {
  padding: 6px 28px 6px 12px;
  border: 1.5px solid #e5e9f0; border-radius: 9px;
  font-size: 13px; font-weight: 600; color: #374151;
  background: #fff url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 24 24'%3E%3Cpath fill='%239facc5' d='M7 10l5 5 5-5z'/%3E%3C/svg%3E") no-repeat right 8px center;
  appearance: none; cursor: pointer;
  transition: border-color 0.15s, box-shadow 0.15s;
}
.term-select:focus { outline: none; border-color: #2A4D99; box-shadow: 0 0 0 3px rgba(42,77,153,0.1); }
.term-select:hover { border-color: #c8d4e8; }
</style>
