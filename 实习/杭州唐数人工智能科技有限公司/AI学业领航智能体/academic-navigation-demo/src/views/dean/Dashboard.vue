<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import * as echarts from 'echarts'
import KpiCard from '@/components/common/KpiCard.vue'
import { getDeanDashboard, getDeanTerms } from '@/api'
import type { DeanData } from '@/types'

const emptyDean: DeanData = {
  totalStudents: 0,
  avgGpa: 0,
  alertRate: 0,
  coursePassRate: 0,
  interventionResponseRate: 0,
  departmentRanking: [],
  alertHeatmap: [],
}

const data = ref<DeanData>({ ...emptyDean })
const route = useRoute()
const router = useRouter()
const loading = ref(true)
const termOptions = ref<string[]>([])
const selectedTerm = ref('')

const rankChartRef = ref<HTMLDivElement>()
const gpaBarRef = ref<HTMLDivElement>()
let rankChart: echarts.ECharts | null = null
let gpaBarChart: echarts.ECharts | null = null

async function load() {
  if (!selectedTerm.value) return
  loading.value = true
  try {
    data.value = await getDeanDashboard(selectedTerm.value)
  } finally {
    loading.value = false
    renderCharts()
  }
}

async function initialize() {
  loading.value = true
  try {
    termOptions.value = await getDeanTerms()
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
  rankChart?.dispose()
  gpaBarChart?.dispose()
  if (!rankChartRef.value || !gpaBarRef.value) return

  const sorted = [...data.value.departmentRanking].sort((a, b) => a.passRate - b.passRate)
  rankChart = echarts.init(rankChartRef.value)
  rankChart.setOption({
    tooltip: {
      trigger: 'axis', axisPointer: { type: 'shadow' },
      backgroundColor: '#fff', borderColor: '#edf0f6', borderWidth: 1,
      textStyle: { color: '#1a2233', fontSize: 12 },
      formatter: (p: { name: string; value: number }[]) => `${p[0].name}：${p[0].value}%`,
    },
    grid: { top: 8, left: 88, right: 56, bottom: 8 },
    xAxis: {
      type: 'value', min: 0, max: 100,
      axisLabel: { color: '#9ca3af', fontSize: 11, formatter: '{value}%' },
      splitLine: { lineStyle: { color: '#f2f4f9' } },
    },
    yAxis: {
      type: 'category', data: sorted.map(d => d.dept),
      axisLabel: { color: '#374151', fontSize: 12 },
      axisLine: { show: false }, axisTick: { show: false },
    },
    series: [{
      type: 'bar', barMaxWidth: 18,
      itemStyle: {
        color: (p: { dataIndex: number }) => {
          const r = sorted[p.dataIndex].passRate
          return r >= 88 ? '#3db87e' : r >= 78 ? '#f09a4e' : '#e04538'
        },
        borderRadius: [0, 6, 6, 0],
      },
      label: { show: true, position: 'right', formatter: (p: { value: number }) => `${p.value}%`, fontSize: 12, color: '#6b7280', fontWeight: 600 },
      data: sorted.map(d => d.passRate),
    }],
  })

  const depts = data.value.departmentRanking.map(d => d.dept)
  const gpas = data.value.departmentRanking.map(d => d.avgGpa)
  gpaBarChart = echarts.init(gpaBarRef.value)
  gpaBarChart.setOption({
    tooltip: {
      trigger: 'axis', axisPointer: { type: 'shadow' },
      backgroundColor: '#fff', borderColor: '#edf0f6', borderWidth: 1,
      textStyle: { color: '#1a2233', fontSize: 12 },
      formatter: (p: { name: string; value: number }[]) => `${p[0].name}：GPA ${p[0].value}`,
    },
    grid: { top: 16, left: 36, right: 16, bottom: 36 },
    xAxis: {
      type: 'category', data: depts,
      axisLine: { show: false }, axisTick: { show: false },
      axisLabel: { color: '#9ca3af', fontSize: 11, interval: 0, rotate: 0 },
    },
    yAxis: {
      type: 'value', min: 0, max: 4,
      axisLabel: { color: '#9ca3af', fontSize: 10 },
      splitLine: { lineStyle: { color: '#f2f4f9' } },
    },
    series: [{
      type: 'bar', barMaxWidth: 32,
      itemStyle: {
        color: (p: { dataIndex: number }) => {
          const colors = ['#2A4D99', '#3db87e', '#7c3aed', '#f09a4e', '#0ea5e9', '#e04538']
          return colors[p.dataIndex % colors.length]
        },
        borderRadius: [4, 4, 0, 0],
      },
      label: { show: true, position: 'top', formatter: (p: { value: number }) => `${p.value}`, fontSize: 11, color: '#6b7280', fontWeight: 700 },
      data: gpas,
    }],
  })

  gpaBarChart.setOption({
    series: [{ markLine: {
      silent: true,
      lineStyle: { color: '#e04538', type: 'dashed', width: 1.5 },
      label: { formatter: `院均 ${data.value.avgGpa}`, position: 'end', color: '#e04538', fontSize: 11 },
      data: [{ yAxis: data.value.avgGpa }],
    }}],
  }, false)
}

watch(selectedTerm, async term => {
  if (!term) return
  await router.replace({ query: { ...route.query, term } })
  await load()
})
onMounted(initialize)
onBeforeUnmount(() => { rankChart?.dispose(); gpaBarChart?.dispose() })
</script>

<template>
  <div class="page" v-loading="loading">
    <div class="page-head">
      <div>
        <h2 class="page-title">全院学业质量仪表盘</h2>
        <p class="page-sub">数学与计算机科学学院 · {{ selectedTerm }}</p>
      </div>
      <div class="term-selector">
        <span class="term-label">学期</span>
        <select v-model="selectedTerm" class="term-select">
          <option v-for="t in termOptions" :key="t" :value="t">{{ t }}</option>
        </select>
      </div>
    </div>

    <div class="kpi-grid">
      <KpiCard title="全院在校生" :value="data.totalStudents.toLocaleString()" unit="人" color="blue" icon="User" />
      <KpiCard title="全院平均 GPA" :value="data.avgGpa.toFixed(2)" color="green" icon="TrendCharts" trend="up" />
      <KpiCard title="综合预警率" :value="`${data.alertRate}%`" color="orange" icon="Warning" />
      <KpiCard title="课程达标率" :value="`${data.coursePassRate}%`" color="blue" icon="CircleCheck" trend="up" />
      <KpiCard title="干预响应率" :value="`${data.interventionResponseRate}%`" color="green" icon="Finished" />
    </div>

    <div class="charts-grid">
      <div class="card">
        <div class="card-title">各专业课程通过率排名</div>
        <div ref="rankChartRef" style="height: 240px" />
      </div>
      <div class="card">
        <div class="card-title">各专业 GPA 对比</div>
        <div ref="gpaBarRef" style="height: 240px" />
      </div>
    </div>

    <div class="card table-card">
      <div class="table-head">
        <span class="card-title" style="margin: 0">各专业预警分布</span>
      </div>
      <el-table :data="data.alertHeatmap" stripe>
        <el-table-column label="专业" prop="dept" width="160" />
        <el-table-column label="黄色预警" prop="yellow" width="120">
          <template #default="{ row }">
            <span class="badge yellow">{{ row.yellow }} 人</span>
          </template>
        </el-table-column>
        <el-table-column label="橙色预警" prop="orange" width="120">
          <template #default="{ row }">
            <span class="badge orange">{{ row.orange }} 人</span>
          </template>
        </el-table-column>
        <el-table-column label="红色预警" prop="red" width="120">
          <template #default="{ row }">
            <span class="badge red">{{ row.red }} 人</span>
          </template>
        </el-table-column>
        <el-table-column label="合计">
          <template #default="{ row }">
            <span class="total">{{ row.yellow + row.orange + row.red }} 人</span>
          </template>
        </el-table-column>
      </el-table>
    </div>
  </div>
</template>

<style scoped>
.page { padding: 28px 32px; display: flex; flex-direction: column; gap: 20px; }
.page-head { display: flex; align-items: flex-start; }
.term-selector { display: flex; align-items: center; gap: 8px; margin-left: auto; }
.term-label { font-size: 12.5px; color: #9facc5; font-weight: 600; }
.term-select {
  padding: 6px 28px 6px 12px; border: 1.5px solid #e5e9f0; border-radius: 9px;
  font-size: 13px; font-weight: 600; color: #374151;
  background: #fff url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 24 24'%3E%3Cpath fill='%239facc5' d='M7 10l5 5 5-5z'/%3E%3C/svg%3E") no-repeat right 8px center;
  appearance: none; cursor: pointer;
}
.term-select:focus { outline: none; border-color: #2A4D99; }
.page-title { font-size: 20px; font-weight: 800; color: #101d3e; margin: 0; letter-spacing: -0.01em; }
.page-sub { font-size: 13px; color: #9facc5; margin: 4px 0 0; }

.kpi-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(180px, 1fr)); gap: 14px; }
.charts-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 14px; }

.card {
  background: #fff; border-radius: 16px; padding: 20px 22px;
  border: 1px solid rgba(10,19,41,0.07);
  box-shadow: 0 1px 2px rgba(10,19,41,0.03), 0 4px 16px rgba(10,19,41,0.05);
}
.card-title { font-size: 14px; font-weight: 700; color: #111827; margin-bottom: 16px; }

.table-card { padding: 0; overflow: hidden; }
.table-head { padding: 16px 22px; border-bottom: 1px solid #f2f4f9; }

.badge { font-size: 12.5px; font-weight: 600; padding: 3px 10px; border-radius: 6px; }
.badge.yellow { color: #92650a; background: #fef3c7; }
.badge.orange { color: #9a3412; background: #fff7ed; }
.badge.red    { color: #991b1b; background: #fef2f2; }
.total { font-size: 13px; font-weight: 700; color: #374151; }
</style>
