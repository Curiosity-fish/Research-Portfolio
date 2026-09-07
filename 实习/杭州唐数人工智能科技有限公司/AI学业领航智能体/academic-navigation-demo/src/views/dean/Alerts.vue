<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import * as echarts from 'echarts'
import { getDeanAlertTrend, getDeanAlerts, getDeanTerms } from '@/api'
import type { AlertTrendData, DeanAlertStatsData } from '@/types'

const emptyStats: DeanAlertStatsData = {
  term: '',
  total: { red: 0, orange: 0, yellow: 0, total: 0 },
  byDept: [],
  byGrade: [],
}

const stats = ref<DeanAlertStatsData>({ ...emptyStats })
const route = useRoute()
const router = useRouter()
const trendData = ref<AlertTrendData>({ terms: [], red: [], orange: [], yellow: [] })
const loading = ref(true)

const barRef = ref<HTMLDivElement>()
const trendRef = ref<HTMLDivElement>()
let barChart: echarts.ECharts | null = null
let trendChart: echarts.ECharts | null = null

const viewMode = ref<'dept' | 'grade'>('dept')
const termOptions = ref<string[]>([])
const selectedTerm = ref('')

const deptAlerts = computed(() => stats.value.byDept)
const gradeAlerts = computed(() => stats.value.byGrade)

const totalRed    = computed(() => stats.value.total.red)
const totalOrange = computed(() => stats.value.total.orange)
const totalYellow = computed(() => stats.value.total.yellow)
const totalAll    = computed(() => stats.value.total.total)

const sortedDepts = computed(() =>
  [...deptAlerts.value].sort((a, b) => b.rate - a.rate)
)

function riskColor(rate: number) {
  if (rate >= 9) return '#e04538'
  if (rate >= 7) return '#f09a4e'
  return '#3db87e'
}

function riskLabel(rate: number) {
  if (rate >= 9) return '高风险'
  if (rate >= 7) return '需关注'
  return '正常'
}

function riskBg(rate: number) {
  if (rate >= 9) return 'rgba(224,69,56,0.08)'
  if (rate >= 7) return 'rgba(240,154,78,0.08)'
  return 'rgba(61,184,126,0.08)'
}

async function load() {
  if (!selectedTerm.value) return
  loading.value = true
  try {
    const [s, t] = await Promise.all([
      getDeanAlerts(selectedTerm.value),
      getDeanAlertTrend(),
    ])
    stats.value = s
    trendData.value = t
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
  barChart?.dispose()
  trendChart?.dispose()
  if (!barRef.value || !trendRef.value) return

  const sorted = [...deptAlerts.value].sort((a, b) => a.rate - b.rate)
  barChart = echarts.init(barRef.value)
  barChart.setOption({
    tooltip: {
      trigger: 'axis', axisPointer: { type: 'shadow' },
      backgroundColor: '#fff', borderColor: '#edf0f6', borderWidth: 1,
      textStyle: { color: '#1a2233', fontSize: 12 },
      formatter: (p: { name: string; value: number }[]) => `${p[0].name}：预警率 ${p[0].value}%`,
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
        color: (p: { dataIndex: number }) => riskColor(sorted[p.dataIndex].rate),
        borderRadius: [0, 6, 6, 0],
      },
      label: { show: true, position: 'right', formatter: (p: { value: number }) => `${p.value}%`, fontSize: 12, color: '#6b7280', fontWeight: 600 },
      data: sorted.map(d => d.rate),
    }],
  })

  trendChart = echarts.init(trendRef.value)
  trendChart.setOption({
    tooltip: {
      trigger: 'axis',
      backgroundColor: '#fff', borderColor: '#edf0f6', borderWidth: 1,
      textStyle: { color: '#1a2233', fontSize: 12 },
    },
    legend: { data: ['红色预警', '橙色预警', '黄色预警'], bottom: 0, textStyle: { color: '#6b7280', fontSize: 11 }, itemWidth: 10, itemHeight: 10 },
    grid: { top: 16, left: 36, right: 16, bottom: 40 },
    xAxis: {
      type: 'category', data: trendData.value.terms,
      axisLine: { show: false }, axisTick: { show: false },
      axisLabel: { color: '#9ca3af', fontSize: 11 },
    },
    yAxis: {
      type: 'value',
      axisLabel: { color: '#9ca3af', fontSize: 10 },
      splitLine: { lineStyle: { color: '#f2f4f9' } },
    },
    series: [
      { name: '红色预警', type: 'line', smooth: true, symbol: 'circle', symbolSize: 5,
        lineStyle: { color: '#e04538', width: 2 }, itemStyle: { color: '#e04538', borderWidth: 2, borderColor: '#fff' },
        data: trendData.value.red },
      { name: '橙色预警', type: 'line', smooth: true, symbol: 'circle', symbolSize: 5,
        lineStyle: { color: '#f09a4e', width: 2 }, itemStyle: { color: '#f09a4e', borderWidth: 2, borderColor: '#fff' },
        data: trendData.value.orange },
      { name: '黄色预警', type: 'line', smooth: true, symbol: 'circle', symbolSize: 5,
        lineStyle: { color: '#ca8a04', width: 2 }, itemStyle: { color: '#ca8a04', borderWidth: 2, borderColor: '#fff' },
        data: trendData.value.yellow },
    ],
  })
}

watch(selectedTerm, async term => {
  if (!term) return
  await router.replace({ query: { ...route.query, term } })
  await load()
})
onMounted(initialize)
onBeforeUnmount(() => { barChart?.dispose(); trendChart?.dispose() })
</script>

<template>
  <div class="page" v-loading="loading">
    <div class="page-head">
      <div>
        <h2 class="page-title">全院预警管理</h2>
        <p class="page-sub">各专业学业预警分布与趋势 · {{ selectedTerm }}</p>
      </div>
      <div class="head-right">
        <select v-model="selectedTerm" class="filter-select">
          <option v-for="t in termOptions" :key="t">{{ t }}</option>
        </select>
        <div class="view-toggle">
          <button :class="['toggle-btn', { active: viewMode === 'dept' }]" @click="viewMode = 'dept'">按专业</button>
          <button :class="['toggle-btn', { active: viewMode === 'grade' }]" @click="viewMode = 'grade'">按年级</button>
        </div>
      </div>
    </div>

    <!-- KPI 卡片 -->
    <div class="kpi-row">
      <div class="kpi-card red">
        <div class="kc-num">{{ totalRed }}</div>
        <div class="kc-label">红色预警</div>
        <div class="kc-desc">挂科两门及以上或缺勤四次及以上</div>
      </div>
      <div class="kpi-card orange">
        <div class="kc-num">{{ totalOrange }}</div>
        <div class="kc-label">橙色预警</div>
        <div class="kc-desc">挂科一门或缺勤三次</div>
      </div>
      <div class="kpi-card yellow">
        <div class="kc-num">{{ totalYellow }}</div>
        <div class="kc-label">黄色预警</div>
        <div class="kc-desc">缺勤一到两次且无挂科</div>
      </div>
      <div class="kpi-card total">
        <div class="kc-num">{{ totalAll }}</div>
        <div class="kc-label">预警合计</div>
        <div class="kc-desc">全院预警总人数</div>
      </div>
    </div>

    <!-- 图表区 -->
    <div class="charts-row">
      <div class="card chart-card">
        <div class="card-title">各专业预警率对比</div>
        <div ref="barRef" style="height: 240px" />
      </div>
      <div class="card chart-card">
        <div class="card-title">近四学期预警趋势</div>
        <div ref="trendRef" style="height: 240px" />
      </div>
    </div>

    <!-- 按专业详情 -->
    <div v-if="viewMode === 'dept'" class="card">
      <div class="card-title">各专业预警明细</div>
      <div class="detail-table">
        <div class="dt-head">
          <span>专业</span>
          <span>在校生</span>
          <span>红色</span>
          <span>橙色</span>
          <span>黄色</span>
          <span>预警率</span>
          <span>风险等级</span>
        </div>
        <div v-for="d in sortedDepts" :key="d.dept" class="dt-row">
          <span class="dt-dept">{{ d.dept }}</span>
          <span class="dt-total-s">{{ d.totalStudents }}人</span>
          <span class="dt-red">{{ d.red }}</span>
          <span class="dt-orange">{{ d.orange }}</span>
          <span class="dt-yellow">{{ d.yellow }}</span>
          <span class="dt-rate" :style="{ color: riskColor(d.rate), fontWeight: 700 }">{{ d.rate }}%</span>
          <span>
            <span class="risk-tag" :style="{ color: riskColor(d.rate), background: riskBg(d.rate) }">
              {{ riskLabel(d.rate) }}
            </span>
          </span>
        </div>
      </div>
    </div>

    <!-- 按年级详情 -->
    <div v-else class="card">
      <div class="card-title">各年级预警明细（全院）</div>
      <div class="grade-list">
        <div v-for="g in gradeAlerts" :key="g.grade" class="grade-row">
          <div class="gr-left">
            <span class="gr-grade">{{ g.grade }}</span>
            <span class="gr-note">{{ g.note }}</span>
          </div>
          <div class="gr-badges">
            <span class="gb red">红 {{ g.red }}</span>
            <span class="gb orange">橙 {{ g.orange }}</span>
            <span class="gb yellow">黄 {{ g.yellow }}</span>
          </div>
          <div class="gr-bar-wrap">
            <div class="gr-bar-track">
              <div class="gr-seg red-seg"  :style="{ flex: g.red }" />
              <div class="gr-seg org-seg"  :style="{ flex: g.orange }" />
              <div class="gr-seg yel-seg"  :style="{ flex: g.yellow }" />
            </div>
            <span class="gr-total">{{ g.total }}人</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.page { padding: 28px 32px; display: flex; flex-direction: column; gap: 20px; }
.page-head { display: flex; align-items: flex-start; justify-content: space-between; flex-wrap: wrap; gap: 12px; }
.page-title { font-size: 20px; font-weight: 800; color: #101d3e; margin: 0; letter-spacing: -0.01em; }
.page-sub { font-size: 13px; color: #9facc5; margin: 4px 0 0; }
.head-right { display: flex; align-items: center; gap: 10px; }
.filter-select {
  padding: 6px 28px 6px 12px; border: 1.5px solid #e5e9f0; border-radius: 9px;
  font-size: 13px; font-weight: 600; color: #374151;
  background: #fff url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 24 24'%3E%3Cpath fill='%239facc5' d='M7 10l5 5 5-5z'/%3E%3C/svg%3E") no-repeat right 8px center;
  appearance: none; cursor: pointer;
}
.filter-select:focus { outline: none; border-color: #2A4D99; }
.view-toggle { display: flex; gap: 3px; background: #f0f2f7; border-radius: 10px; padding: 3px; }
.toggle-btn { padding: 6px 14px; border: none; border-radius: 8px; font-size: 12.5px; font-weight: 600; color: #6b7280; background: transparent; cursor: pointer; transition: all 0.15s; }
.toggle-btn.active { background: #fff; color: #2A4D99; box-shadow: 0 1px 4px rgba(0,0,0,0.1); }

.kpi-row { display: grid; grid-template-columns: repeat(4, 1fr); gap: 14px; }
.kpi-card { background: #fff; border-radius: 14px; padding: 18px 20px; border: 2px solid #edf0f6; box-shadow: 0 1px 4px rgba(0,0,0,0.04); }
.kpi-card.red    { border-color: #fecaca; background: #fef2f2; }
.kpi-card.orange { border-color: #fed7aa; background: #fff7ed; }
.kpi-card.yellow { border-color: #fde68a; background: #fefce8; }
.kpi-card.total  { border-color: #e5e9f0; }
.kc-num { font-size: 36px; font-weight: 800; color: #111827; line-height: 1.1; font-variant-numeric: tabular-nums; }
.kpi-card.red .kc-num    { color: #e04538; }
.kpi-card.orange .kc-num { color: #f09a4e; }
.kpi-card.yellow .kc-num { color: #ca8a04; }
.kc-label { font-size: 13px; font-weight: 700; color: #374151; margin: 4px 0 2px; }
.kc-desc { font-size: 11.5px; color: #9facc5; }

.charts-row { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; }
.chart-card {}
.card { background: #fff; border-radius: 16px; padding: 20px 22px; border: 1px solid rgba(10,19,41,0.07); box-shadow: 0 1px 2px rgba(10,19,41,0.03), 0 4px 16px rgba(10,19,41,0.05); }
.card-title { font-size: 14px; font-weight: 700; color: #111827; margin-bottom: 16px; }

.detail-table { display: flex; flex-direction: column; }
.dt-head { display: grid; grid-template-columns: 1.4fr 0.8fr 0.6fr 0.6fr 0.6fr 0.8fr 1fr; padding: 8px 10px; background: #f8f9fc; border-radius: 8px; font-size: 12px; font-weight: 700; color: #9facc5; margin-bottom: 4px; }
.dt-row { display: grid; grid-template-columns: 1.4fr 0.8fr 0.6fr 0.6fr 0.6fr 0.8fr 1fr; padding: 12px 10px; border-bottom: 1px solid #f0f2f7; font-size: 13px; align-items: center; transition: background 0.12s; }
.dt-row:hover { background: #f8f9fc; }
.dt-row:last-child { border-bottom: none; }
.dt-dept { font-weight: 700; color: #111827; }
.dt-total-s { color: #6b7280; }
.dt-red    { color: #e04538; font-weight: 700; }
.dt-orange { color: #f09a4e; font-weight: 700; }
.dt-yellow { color: #ca8a04; font-weight: 700; }
.dt-rate { font-variant-numeric: tabular-nums; }
.risk-tag { font-size: 11.5px; font-weight: 600; padding: 3px 8px; border-radius: 6px; }

.grade-list { display: flex; flex-direction: column; gap: 14px; }
.grade-row { display: flex; align-items: center; gap: 16px; padding: 12px 14px; background: #f8f9fc; border-radius: 12px; }
.gr-left { width: 180px; flex-shrink: 0; }
.gr-grade { font-size: 13.5px; font-weight: 700; color: #111827; display: block; }
.gr-note { font-size: 12px; color: #9facc5; margin-top: 2px; display: block; }
.gr-badges { display: flex; gap: 6px; flex-shrink: 0; }
.gb { font-size: 12px; font-weight: 700; padding: 3px 8px; border-radius: 6px; }
.gb.red    { background: rgba(224,69,56,0.1); color: #e04538; }
.gb.orange { background: rgba(240,154,78,0.1); color: #f09a4e; }
.gb.yellow { background: rgba(202,138,4,0.1);  color: #ca8a04; }
.gr-bar-wrap { flex: 1; display: flex; align-items: center; gap: 10px; }
.gr-bar-track { flex: 1; height: 10px; border-radius: 99px; overflow: hidden; display: flex; }
.gr-seg { height: 100%; }
.red-seg  { background: #e04538; }
.org-seg  { background: #f09a4e; }
.yel-seg  { background: #fde68a; }
.gr-total { font-size: 13px; font-weight: 700; color: #374151; width: 36px; text-align: right; white-space: nowrap; }
</style>
