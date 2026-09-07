<script setup lang="ts">
import { computed, onMounted, onBeforeUnmount, ref, watch, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import * as echarts from 'echarts'
import RadarChart from '@/components/common/RadarChart.vue'
import WarningBadge from '@/components/common/WarningBadge.vue'
import AiInsightPanel from '@/components/common/AiInsightPanel.vue'
import { getTeacherStudentDetail, getTeacherTerms, interpretMetrics } from '@/api'
import type { AcademicProfile, StudentDetailData } from '@/types'

const route = useRoute()
const router = useRouter()

const detail = ref<StudentDetailData | null>(null)
const loading = ref(false)
const availableTerms = ref<string[]>([])
const selectedTerm = ref('')
const emptyProfile: AcademicProfile = { dimensions: [], gpaTrend: [], rankTrend: [] }
const student = computed(() => detail.value?.student ?? null)
const studentAlerts = computed(() => detail.value?.alerts ?? [])
const profile = computed(() => detail.value?.profile ?? emptyProfile)

// AI 学生画像叙述（本地模型基于真实学生数据生成）
const aiPortraitLoading = ref(false)
const aiPortraitSummary = ref('正在基于学生画像数据生成 AI 叙述...')
const aiPortraitPoints = ref<string[]>([])
const aiPortraitCache = new Map<string, { summary: string; points: string[] }>()

async function loadAiPortrait() {
  if (!detail.value) return
  const cacheKey = `${route.params.id}-${selectedTerm.value}`
  const cached = aiPortraitCache.get(cacheKey)
  if (cached) {
    aiPortraitSummary.value = cached.summary
    aiPortraitPoints.value = cached.points
    return
  }
  aiPortraitLoading.value = true
  try {
    const res = await interpretMetrics(
      '请基于以下学生画像数据生成简明的AI画像叙述：第一行输出一句总体评价，后续每行用“- ”开头输出分析或建议，最多5条；不要出现#、**、反引号，不要编造数据。',
      {
        student: student.value ? {
          name: student.value.name,
          grade: student.value.grade,
          major: student.value.major,
          gpa: student.value.gpa,
          rank: student.value.rank,
          totalStudents: student.value.totalStudents,
          alertLevel: student.value.alertLevel,
        } : null,
        term: selectedTerm.value,
        dimensions: profile.value.dimensions.map(d => ({ label: d.label, score: d.score, description: d.description })),
        alerts: studentAlerts.value.map(a => ({ level: a.level, title: a.title, status: a.status })),
      },
    )
    const clean = (s: string) => s.replace(/[#*`]/g, '').trim()
    const lines = res.content.split('\n').map(clean).filter(Boolean)
    const bullets = lines.filter(l => /^[-•]/.test(l)).map(l => l.replace(/^[-•]\s*/, ''))
    if (bullets.length > 0) {
      aiPortraitSummary.value = lines.find(l => !/^[-•]/.test(l)) || 'AI 学生画像叙述'
      aiPortraitPoints.value = bullets.slice(0, 5)
    } else {
      aiPortraitSummary.value = res.content.replace(/[#*`]/g, '').trim()
      aiPortraitPoints.value = []
    }
    aiPortraitCache.set(cacheKey, { summary: aiPortraitSummary.value, points: aiPortraitPoints.value })
  } catch {
    aiPortraitSummary.value = 'AI 学生画像叙述生成失败，请稍后重试'
    aiPortraitPoints.value = []
  } finally {
    aiPortraitLoading.value = false
  }
}

const trendRef = ref<HTMLDivElement>()
let trendChart: echarts.ECharts | null = null

async function loadDetail() {
  if (!selectedTerm.value) return
  loading.value = true
  try {
    detail.value = await getTeacherStudentDetail(String(route.params.id), selectedTerm.value)
    await nextTick()
    renderTrend()
  } finally {
    loading.value = false
  }
  loadAiPortrait()
}

function renderTrend() {
  if (!trendRef.value) return
  if (!trendChart) trendChart = echarts.init(trendRef.value)
  trendChart.setOption({
    tooltip: { trigger: 'axis' },
    grid: { left: 40, right: 20, top: 20, bottom: 36 },
    xAxis: { type: 'category', data: profile.value.gpaTrend.map(d => d.semester), axisTick: { show: false }, axisLine: { lineStyle: { color: '#eee' } } },
    yAxis: { type: 'value', min: 0, max: 4, splitLine: { lineStyle: { color: '#f5f5f5' } } },
    series: [{ name: 'GPA', type: 'bar', data: profile.value.gpaTrend.map(d => d.gpa), barWidth: '40%', itemStyle: { color: '#1E5BB5', borderRadius: [4, 4, 0, 0] } }],
  })
}

async function initialize() {
  loading.value = true
  try {
    const context = await getTeacherTerms()
    availableTerms.value = context.terms
    const queryTerm = typeof route.query.term === 'string' ? route.query.term : ''
    selectedTerm.value = context.terms.includes(queryTerm) ? queryTerm : context.latestTerm
  } finally {
    loading.value = false
  }
}

onMounted(initialize)

watch(() => route.params.id, loadDetail)
watch(selectedTerm, async term => {
  if (!term) return
  await router.replace({ query: { ...route.query, term } })
  loadDetail()
})

onBeforeUnmount(() => { trendChart?.dispose() })

const radarDims = computed(() => profile.value.dimensions.map(d => ({ label: d.label, score: d.score, avgScore: d.avgScore })))
</script>

<template>
  <div v-loading="loading" v-if="student" class="p-6 flex flex-col gap-6">
    <!-- Back + header -->
    <div class="flex items-center justify-between gap-3">
      <div class="flex items-center gap-3">
        <el-button text @click="router.back()"><el-icon><ArrowLeft /></el-icon> 返回</el-button>
        <h2 class="text-xl font-bold text-gray-800">学生画像详情</h2>
      </div>
      <div class="flex items-center gap-2">
        <span class="text-xs font-semibold text-gray-400">学期</span>
        <select
          v-if="availableTerms.length > 1"
          v-model="selectedTerm"
          class="rounded-lg border border-gray-200 bg-white px-3 py-2 text-sm font-semibold text-gray-700 outline-none focus:border-zjnu-blue-500"
        >
          <option v-for="term in availableTerms" :key="term" :value="term">{{ term }}</option>
        </select>
        <span v-else class="text-sm font-semibold text-gray-600">{{ selectedTerm }}</span>
      </div>
    </div>

    <!-- Student info card -->
    <div class="bg-white rounded-2xl p-5 shadow-sm border border-gray-100">
      <div class="flex items-center gap-5">
        <div class="w-16 h-16 rounded-2xl bg-zjnu-blue-500 flex items-center justify-center text-white text-2xl font-bold">
          {{ student.name[0] }}
        </div>
        <div class="flex-1">
          <div class="flex items-center gap-3">
            <span class="text-xl font-bold text-gray-800">{{ student.name }}</span>
            <WarningBadge v-if="student.alertLevel !== 'none'" :level="student.alertLevel as 'yellow' | 'orange' | 'red'" />
            <span v-else class="text-xs text-growth-green-600 bg-growth-green-50 px-2 py-0.5 rounded-full">正常</span>
          </div>
          <div class="flex gap-4 mt-2 text-sm text-gray-500">
            <span>{{ student.studentId }}</span>
            <span>{{ student.major }}</span>
            <span>{{ student.className }}</span>
          </div>
        </div>
        <div class="flex gap-6 text-center">
          <div>
            <div class="text-2xl font-bold text-zjnu-blue-500">{{ student.gpa }}</div>
            <div class="text-xs text-gray-400 mt-1">GPA</div>
          </div>
          <div>
            <div class="text-2xl font-bold text-growth-green-500">{{ student.rank }}</div>
            <div class="text-xs text-gray-400 mt-1">专业排名</div>
          </div>
        </div>
      </div>
    </div>

    <!-- AI 学生画像叙述 -->
    <AiInsightPanel
      title="AI 学生画像叙述"
      :summary="aiPortraitSummary"
      :points="aiPortraitPoints"
      :actions="[{ label: '生成面谈提纲', color: '#2A4D99' }, { label: '查看预警历史', color: '#e04538' }]"
      :loading="aiPortraitLoading"
      accent="#2A4D99"
    />

    <!-- Radar + GPA trend -->
    <div class="grid grid-cols-2 gap-5">
      <div class="bg-white rounded-2xl p-5 shadow-sm border border-gray-100">
        <div class="text-base font-semibold text-gray-700 mb-2">五维学业画像</div>
        <RadarChart :dimensions="radarDims" />
      </div>
      <div class="bg-white rounded-2xl p-5 shadow-sm border border-gray-100">
        <div class="text-base font-semibold text-gray-700 mb-4">GPA 趋势</div>
        <div ref="trendRef" class="h-56" />
      </div>
    </div>

    <!-- Alerts -->
    <div class="bg-white rounded-2xl p-5 shadow-sm border border-gray-100">
      <div class="text-base font-semibold text-gray-700 mb-4">本学期预警记录</div>
      <div v-if="studentAlerts.length === 0" class="text-sm text-gray-400 py-4 text-center">该学生暂无预警记录</div>
      <div v-else class="flex flex-col gap-3">
        <div v-for="alert in studentAlerts" :key="alert.id" class="flex items-start gap-3 p-3 rounded-xl bg-bg-gray-50">
          <WarningBadge :level="alert.level" />
          <div class="flex-1">
            <div class="text-sm font-medium text-gray-700">{{ alert.title }}</div>
            <div class="text-xs text-gray-400 mt-0.5">{{ alert.description }}</div>
          </div>
          <span class="text-xs text-gray-400">{{ alert.date }}</span>
        </div>
      </div>
    </div>
  </div>
  <div v-else-if="!loading" class="p-6 text-center text-gray-400">学生不存在或无权查看</div>
</template>
