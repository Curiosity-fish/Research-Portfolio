<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import * as echarts from 'echarts'
import KpiCard from '@/components/common/KpiCard.vue'
import AiInsightPanel from '@/components/common/AiInsightPanel.vue'
import WarningBadge from '@/components/common/WarningBadge.vue'
import { getDepartmentOverview, getDepartmentTerms } from '@/api'
import { useAuthStore } from '@/stores/auth'
import type { DepartmentOverviewData } from '@/types'

const auth = useAuthStore()
const route = useRoute()
const router = useRouter()
const majorName = computed(() => auth.user?.major || '应用统计')

const emptyOverview: DepartmentOverviewData = {
  totalStudents: 0,
  avgGpa: 0,
  alertRate: 0,
  passRate: 0,
  gradeGpaTrend: [],
  alertDistribution: { yellow: 0, orange: 0, red: 0 },
  focusStudents: [],
}

const data = ref<DepartmentOverviewData>({ ...emptyOverview })
const loading = ref(true)
const aiLoading = ref(false)

const termOptions = ref<string[]>([])
const selectedTerm = ref('')

const gpaChartRef = ref<HTMLDivElement>()
const pieRef = ref<HTMLDivElement>()
let gpaChart: echarts.ECharts | null = null
let pieChart: echarts.ECharts | null = null

const seriesColors = ['#2A4D99', '#3db87e', '#7c3aed', '#f09a4e', '#0ea5e9', '#e04538']

const gradeLabels = computed(() => data.value.gradeGpaTrend.map(g => g.grade))
const gpaSeries = computed(() => [{
  name: selectedTerm.value,
  type: 'bar' as const,
  barWidth: '42%',
  barCategoryGap: '45%',
  data: data.value.gradeGpaTrend.map((g, i) => ({
    value: g.trend[0]?.gpa ?? 0,
    itemStyle: { color: seriesColors[i % seriesColors.length], borderRadius: [4, 4, 0, 0] },
  })),
}])

const alertData = computed(() => [
  { name: '黄色预警', value: data.value.alertDistribution.yellow, color: '#f5a623' },
  { name: '橙色预警', value: data.value.alertDistribution.orange, color: '#f09a4e' },
  { name: '红色预警', value: data.value.alertDistribution.red, color: '#e04538' },
])

const alertItems = computed(() => [
  { key: 'yellow', label: '黄色预警', count: data.value.alertDistribution.yellow, bg: '#fefce8', border: '#fde68a', text: '#854d0e', dot: '#eab308' },
  { key: 'orange', label: '橙色预警', count: data.value.alertDistribution.orange, bg: '#fff7ed', border: '#fed7aa', text: '#9a3412', dot: '#f09a4e' },
  { key: 'red', label: '红色预警', count: data.value.alertDistribution.red, bg: '#fef2f2', border: '#fecaca', text: '#991b1b', dot: '#e04538' },
])

const alertStudents = computed(() => data.value.focusStudents)

const aiSummary = computed(() =>
  `${majorName.value}专业当前学期整体${data.value.avgGpa >= 3.2 ? '平稳' : '需要关注'}，专业均GPA ${data.value.avgGpa.toFixed(2)}，预警率${data.value.alertRate}%，课程达标率${data.value.passRate}%。`
)
const aiPoints = computed(() => [
  `专业均GPA ${data.value.avgGpa.toFixed(2)}${data.value.avgGpa >= 3.2 ? '，处于良好水平' : '，低于预期基准，建议重点分析'}`,
  `预警学生 ${data.value.alertDistribution.yellow + data.value.alertDistribution.orange + data.value.alertDistribution.red} 人，其中红色 ${data.value.alertDistribution.red} 人需优先介入`,
  `课程达标率 ${data.value.passRate}%${data.value.passRate < 88 ? '，低于期望水平，需整体提升' : '，整体达标'}`,
  `专业总人数 ${data.value.totalStudents} 人，当前统计学期 ${selectedTerm.value}`,
])

async function load() {
  if (!selectedTerm.value) return
  loading.value = true
  try {
    data.value = await getDepartmentOverview(selectedTerm.value)
  } finally {
    loading.value = false
    renderCharts()
  }
}

async function initialize() {
  loading.value = true
  try {
    termOptions.value = await getDepartmentTerms()
    const queryTerm = typeof route.query.term === 'string' ? route.query.term : ''
    selectedTerm.value = termOptions.value.includes(queryTerm)
      ? queryTerm
      : (termOptions.value[0] ?? '')
    if (!selectedTerm.value) loading.value = false
  } catch {
    loading.value = false
  }
}

function renderCharts() {
  gpaChart?.dispose()
  pieChart?.dispose()
  if (!gpaChartRef.value || !pieRef.value) return

  gpaChart = echarts.init(gpaChartRef.value)
  gpaChart.setOption({
    tooltip: { trigger: 'axis', backgroundColor: '#fff', borderColor: '#edf0f6', borderWidth: 1, textStyle: { color: '#1a2233', fontSize: 12 } },
    legend: { show: false },
    grid: { top: 16, left: 40, right: 10, bottom: 40 },
    xAxis: {
      type: 'category',
      data: gradeLabels.value,
      axisLine: { show: false }, axisTick: { show: false },
      axisLabel: { color: '#9ca3af', fontSize: 11 },
    },
    yAxis: {
      type: 'value', min: 2.4, max: 4.0,
      axisLabel: { color: '#9ca3af', fontSize: 10 },
      splitLine: { lineStyle: { color: '#f2f4f9' } },
    },
    series: gpaSeries.value,
  })

  pieChart = echarts.init(pieRef.value)
  pieChart.setOption({
    tooltip: { trigger: 'item', formatter: '{b}: {c}人 ({d}%)' },
    legend: {
      orient: 'vertical', right: 4, top: 'middle',
      textStyle: { color: '#6b7280', fontSize: 11 }, itemWidth: 10, itemHeight: 10,
      formatter: (name: string) => {
        const item = alertData.value.find(d => d.name === name)
        return item ? `${name}  ${item.value}人` : name
      },
    },
    series: [{
      type: 'pie', radius: ['44%', '68%'], center: ['34%', '50%'],
      label: { show: false },
      data: alertData.value.map(d => ({ name: d.name, value: d.value, itemStyle: { color: d.color } })),
      emphasis: { itemStyle: { shadowBlur: 8, shadowColor: 'rgba(0,0,0,0.12)' } },
    }],
  })
}

watch(selectedTerm, async term => {
  if (!term) return
  await router.replace({ query: { ...route.query, term } })
  await load()
})
onMounted(initialize)
onBeforeUnmount(() => { gpaChart?.dispose(); pieChart?.dispose() })
</script>

<template>
  <div class="page" v-loading="loading">
    <div class="page-head">
      <div>
        <h2 class="page-title">专业学业态势</h2>
        <p class="page-sub">{{ majorName }}专业 · {{ selectedTerm }}</p>
      </div>
      <div class="term-selector">
        <span class="term-label">学期</span>
        <select v-model="selectedTerm" class="term-select">
          <option v-for="t in termOptions" :key="t" :value="t">{{ t }}</option>
        </select>
      </div>
    </div>

    <div class="kpi-grid">
      <KpiCard title="专业总人数" :value="data.totalStudents" unit="人" color="blue" icon="User" />
      <KpiCard title="专业均 GPA" :value="data.avgGpa.toFixed(2)" color="green" icon="TrendCharts" />
      <KpiCard title="预警比例" :value="`${data.alertRate}%`" color="orange" icon="Warning" />
      <KpiCard title="课程达标率" :value="`${data.passRate}%`" color="blue" icon="CircleCheck" />
    </div>

    <AiInsightPanel
      title="AI 专业学情诊断"
      :summary="aiSummary"
      :points="aiPoints"
      :actions="[{ label: '查看高风险学生', color: '#e04538' }, { label: '生成专业周报', color: '#2A4D99' }, { label: '启动课程评审', color: '#3db87e' }]"
      :loading="aiLoading"
      accent="#2A4D99"
    />

    <div class="mid-grid">
      <div class="card">
        <div class="card-title">各年级 GPA 对比（当前学期）</div>
        <div ref="gpaChartRef" style="height: 200px" />
      </div>
      <div class="card">
        <div class="card-title">预警分布</div>
        <div ref="pieRef" style="height: 200px" />
      </div>
      <div class="card">
        <div class="card-title">预警等级概览</div>
        <div class="alert-bars">
          <div v-for="item in alertItems" :key="item.key"
            class="alert-bar"
            :style="{ background: item.bg, border: `1px solid ${item.border}` }">
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
        <span class="card-title" style="margin:0">重点关注学生</span>
        <router-link :to="{ path: '/department/alerts', query: { term: selectedTerm } }" class="view-all">查看全部</router-link>
      </div>
      <el-table :data="alertStudents" stripe>
        <el-table-column label="姓名" prop="name" width="90" />
        <el-table-column label="学号" prop="studentId" width="120" />
        <el-table-column label="年级" prop="grade" width="90" />
        <el-table-column label="GPA" prop="gpa" width="80" />
        <el-table-column label="预警等级" width="130">
          <template #default="{ row }">
            <WarningBadge v-if="row.alertLevel !== 'none'" :level="row.alertLevel" />
          </template>
        </el-table-column>
        <el-table-column label="专业" prop="major" />
      </el-table>
    </div>
  </div>
</template>

<style scoped>
.page { padding: 28px 32px; display: flex; flex-direction: column; gap: 20px; }
.page-head { display: flex; align-items: flex-start; justify-content: space-between; }
.page-title { font-size: 20px; font-weight: 800; color: #101d3e; margin: 0; letter-spacing: -0.01em; }
.page-sub { font-size: 13px; color: #9facc5; margin: 4px 0 0; }
.term-label { font-size: 12.5px; color: #9facc5; font-weight: 600; }
.term-selector { display: flex; align-items: center; gap: 8px; }
.term-select {
  padding: 6px 28px 6px 12px; border: 1.5px solid #e5e9f0; border-radius: 9px;
  font-size: 13px; font-weight: 600; color: #374151;
  background: #fff url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 24 24'%3E%3Cpath fill='%239facc5' d='M7 10l5 5 5-5z'/%3E%3C/svg%3E") no-repeat right 8px center;
  appearance: none; cursor: pointer;
}
.term-select:focus { outline: none; border-color: #2A4D99; box-shadow: 0 0 0 3px rgba(42,77,153,0.1); }
.kpi-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 14px; }
.mid-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 14px; }
.card { background: #fff; border-radius: 16px; padding: 20px 22px; border: 1px solid rgba(10,19,41,0.07); box-shadow: 0 1px 2px rgba(10,19,41,0.03), 0 4px 16px rgba(10,19,41,0.05); }
.card-title { font-size: 14px; font-weight: 700; color: #111827; margin-bottom: 16px; }
.alert-bars { display: flex; flex-direction: column; gap: 10px; }
.alert-bar { display: flex; align-items: center; justify-content: space-between; padding: 12px 14px; border-radius: 10px; }
.alert-bar-left { display: flex; align-items: center; gap: 8px; }
.alert-dot { width: 8px; height: 8px; border-radius: 50%; flex-shrink: 0; }
.alert-label { font-size: 13px; font-weight: 600; }
.alert-count { font-size: 22px; font-weight: 800; font-variant-numeric: tabular-nums; }
.table-card { padding: 0; overflow: hidden; }
.table-head { display: flex; align-items: center; justify-content: space-between; padding: 16px 22px; border-bottom: 1px solid #f2f4f9; }
.view-all { font-size: 12.5px; color: #2A4D99; text-decoration: none; font-weight: 600; }
.view-all:hover { text-decoration: underline; }
</style>
