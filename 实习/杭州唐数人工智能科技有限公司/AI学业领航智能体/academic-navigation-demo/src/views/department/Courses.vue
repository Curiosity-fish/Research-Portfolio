<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import * as echarts from 'echarts'
import AiInsightPanel from '@/components/common/AiInsightPanel.vue'
import { getDepartmentCourseHeatmap, getDepartmentCourses, getDepartmentTerms } from '@/api'
import type { CourseHeatmapData, DepartmentCourseItem } from '@/types'

const route = useRoute()
const router = useRouter()
const termOptions = ref<string[]>([])
const selectedTerm = ref('')
const loading = ref(true)
const aiCoursesLoading = ref(false)

const heatmapRef = ref<HTMLDivElement>()
let heatChart: echarts.ECharts | null = null

const heatmap = ref<CourseHeatmapData>({ courses: [], grades: [], heatData: [] })
const courseList = ref<DepartmentCourseItem[]>([])

const courses = computed(() => heatmap.value.courses)
const grades = computed(() => heatmap.value.grades)
const heatData = computed<[number, number, number][]>(() => heatmap.value.heatData)

const courseHealthItems = computed(() => courseList.value.map(c => ({
  course: c.name,
  teacher: '',
  risk: courseRisk(c),
  grades: c.gradeStats.map(g => ({
    grade: g.grade,
    avgScore: g.avgScore,
    passRate: g.passRate,
    highRate: g.highRate,
    lowRate: g.lowRate,
  })),
})))

const warnTips = computed(() => {
  const tips: string[] = []
  for (const c of courseList.value) {
    for (const g of c.gradeStats) {
      if (g.passRate < 75) tips.push(`${c.name}（${g.grade}）- 通过率 ${g.passRate}%，建议增加辅导课时`)
    }
  }
  return tips.slice(0, 3)
})

const goodTips = computed(() => {
  const tips: string[] = []
  for (const c of courseList.value) {
    for (const g of c.gradeStats) {
      if (g.passRate >= 92) tips.push(`${c.name}（${g.grade}）- 通过率 ${g.passRate}%`)
    }
  }
  return tips.slice(0, 3)
})

const lowestStat = computed(() => {
  let result = { course: '', grade: '', passRate: 101 }
  for (const c of courseList.value) {
    for (const g of c.gradeStats) {
      if (g.passRate < result.passRate) {
        result = { course: c.name, grade: g.grade, passRate: g.passRate }
      }
    }
  }
  return result
})

const aiCoursesSummary = computed(() =>
  lowestStat.value.course
    ? `基于当前学期课程通过率与成绩分布分析，${lowestStat.value.course}（${lowestStat.value.grade}）通过率仅 ${lowestStat.value.passRate}%，为专业内最低，建议优先核查教学方案与补考安排。`
    : '当前学期暂无课程成绩数据，无法生成健康度诊断。'
)

const aiCoursesPoints = computed(() => [
  lowestStat.value.course ? `${lowestStat.value.course}（${lowestStat.value.grade}）通过率 ${lowestStat.value.passRate}%，建议安排专项辅导` : '暂无低通过率课程',
  courseList.value.some(c => c.gradeStats.some(g => g.highRate > 50 || g.avgScore >= 90))
    ? '部分课程高分率偏高，建议复核评分标准'
    : '各课程成绩分布整体正常',
  `共统计 ${courseList.value.length} 门课程，覆盖 ${grades.value.length} 个年级`,
  `当前统计学期 ${selectedTerm.value}`,
])

const detailVisible = ref(false)
const detailCourse = ref('')
const detailGrade = ref('')
const detailPassRate = ref(0)

function courseRisk(c: DepartmentCourseItem) {
  if (c.gradeStats.some(g => g.highRate > 50 || g.avgScore >= 90)) return 'high'
  if (c.gradeStats.some(g => g.passRate < 75)) return 'low'
  return 'none'
}

function healthScoreColor(score: number) {
  if (score >= 85) return '#3db87e'
  if (score >= 70) return '#f09a4e'
  return '#e04538'
}

function rateColor(r: number) {
  return r >= 88 ? '#3db87e' : r >= 75 ? '#f09a4e' : '#e04538'
}

async function load() {
  if (!selectedTerm.value) return
  loading.value = true
  try {
    const [h, list] = await Promise.all([
      getDepartmentCourseHeatmap(selectedTerm.value),
      getDepartmentCourses(selectedTerm.value),
    ])
    heatmap.value = h
    courseList.value = list
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
  heatChart?.dispose()
  if (!heatmapRef.value) return
  heatChart = echarts.init(heatmapRef.value)
  heatChart.setOption({
    tooltip: {
      trigger: 'item',
      backgroundColor: '#fff', borderColor: '#edf0f6', borderWidth: 1,
      textStyle: { color: '#1a2233', fontSize: 12 },
      extraCssText: 'box-shadow:0 8px 24px rgba(10,19,41,0.1);border-radius:10px',
      formatter: (p: any) =>
        `${courses.value[p.value[1]]} · ${grades.value[p.value[0]]}<br/>通过率：<b>${p.value[2]}%</b>`,
    },
    grid: { top: 20, left: 90, right: 20, bottom: 60 },
    xAxis: {
      type: 'category', data: grades.value,
      axisLabel: { color: '#6b7280', fontSize: 12 },
      axisLine: { show: false }, axisTick: { show: false },
    },
    yAxis: {
      type: 'category', data: courses.value,
      axisLabel: { color: '#374151', fontSize: 12, width: 80, overflow: 'truncate' },
      axisLine: { show: false }, axisTick: { show: false },
    },
    visualMap: {
      min: 60, max: 100, orient: 'horizontal', left: 'center', bottom: 4,
      text: ['100%', '60%'], textStyle: { color: '#6b7280', fontSize: 11 },
      inRange: { color: ['#e04538', '#f5a623', '#5ed29c'] },
    },
    series: [{
      type: 'heatmap', data: heatData.value,
      label: { show: true, formatter: (p: any) => `${p.value[2]}%`, fontSize: 12, fontWeight: 600, color: '#1a2233' },
      itemStyle: { borderRadius: 8, borderWidth: 4, borderColor: '#f6f8fb' },
      emphasis: { itemStyle: { shadowBlur: 8, shadowColor: 'rgba(0,0,0,0.12)' } },
    }],
  })

  heatChart.on('click', (p: any) => {
    detailCourse.value = courses.value[p.value[1]]
    detailGrade.value = grades.value[p.value[0]]
    detailPassRate.value = p.value[2]
    detailVisible.value = true
  })
}

watch(selectedTerm, async term => {
  if (!term) return
  await router.replace({ query: { ...route.query, term } })
  await load()
})
onMounted(initialize)
onBeforeUnmount(() => { heatChart?.dispose() })
</script>

<template>
  <div class="page" v-loading="loading">
    <div class="page-head">
      <div>
        <h2 class="page-title">课程达成度分析</h2>
        <p class="page-sub">各课程在不同年级学生中的通过率分布，点击色块查看详情</p>
      </div>
      <div class="term-selector">
        <span class="term-label">学期</span>
        <select v-model="selectedTerm" class="term-select">
          <option v-for="t in termOptions" :key="t" :value="t">{{ t }}</option>
        </select>
      </div>
    </div>

    <div class="card">
      <div class="card-head">
        <span class="card-title">课程通过率热力图</span>
        <div class="legend-inline">
          <span class="legend-dot" style="background:#e04538" />
          <span class="legend-text">较低 (&lt;75%)</span>
          <span class="legend-dot" style="background:#f5a623" />
          <span class="legend-text">中等 (75-88%)</span>
          <span class="legend-dot" style="background:#5ed29c" />
          <span class="legend-text">达标 (&gt;88%)</span>
        </div>
      </div>
      <div ref="heatmapRef" style="height: 360px" />
    </div>

    <div class="tips-grid">
      <div class="tip-card warn">
        <div class="tip-head">
          <span class="tip-icon">!</span>
          <span class="tip-title">需重点关注</span>
        </div>
        <ul class="tip-list">
          <li v-for="tip in warnTips" :key="tip">{{ tip }}</li>
        </ul>
      </div>
      <div class="tip-card good">
        <div class="tip-head">
          <span class="tip-icon good-icon">✓</span>
          <span class="tip-title">表现优秀</span>
        </div>
        <ul class="tip-list">
          <li v-for="tip in goodTips" :key="tip">{{ tip }}</li>
        </ul>
      </div>
    </div>

    <!-- AI 课程健康度 -->
    <AiInsightPanel
      title="AI 课程健康度诊断"
      :summary="aiCoursesSummary"
      :points="aiCoursesPoints"
      :actions="[{ label: '生成课程健康报告', color: '#2A4D99' }, { label: '预警课程详情', color: '#e04538' }]"
      :loading="aiCoursesLoading"
      accent="#f09a4e"
    />

    <!-- 各课程成绩明细 -->
    <div class="card">
      <div class="card-head">
        <span class="card-title">各课程成绩明细（按年级）</span>
        <div class="legend-inline">
          <span class="risk-tag risk-high">评分偏松</span>
          <span class="risk-tag risk-low">通过率偏低</span>
        </div>
      </div>
      <div class="score-table">
        <div class="score-thead">
          <span>课程 / 年级</span>
          <span>均分</span>
          <span>通过率</span>
          <span>高分率(≥90)</span>
          <span>不及格率</span>
          <span>AI 风险标注</span>
        </div>
        <template v-for="item in courseHealthItems" :key="item.course">
          <!-- 课程主行：课程名 + 授课老师 -->
          <div class="score-course-row">
            <span class="score-course-name">
              {{ item.course }}
              <span class="score-teacher-inline">{{ item.teacher }}</span>
            </span>
            <span></span><span></span><span></span><span></span><span></span>
          </div>
          <!-- 每个年级子行 -->
          <div v-for="g in item.grades" :key="g.grade" class="score-trow">
            <span class="score-grade-cell">{{ g.grade }}</span>
            <span class="score-avg" :style="{ color: g.avgScore >= 90 ? '#f09a4e' : g.avgScore < 75 ? '#e04538' : '#374151' }">
              {{ g.avgScore }}
            </span>
            <span class="score-pass" :style="{ color: healthScoreColor(g.passRate) }">{{ g.passRate }}%</span>
            <span :style="{ color: g.highRate > 50 ? '#f09a4e' : '#374151', fontWeight: g.highRate > 50 ? 700 : 400 }">
              {{ g.highRate }}%
            </span>
            <span :style="{ color: g.lowRate > 25 ? '#e04538' : '#374151', fontWeight: g.lowRate > 25 ? 700 : 400 }">
              {{ g.lowRate }}%
            </span>
            <span>
              <span v-if="g.highRate > 50 || g.avgScore >= 90" class="risk-tag risk-high">评分偏松</span>
              <span v-else-if="g.passRate < 75" class="risk-tag risk-low">通过率偏低</span>
              <span v-else class="risk-tag risk-ok">正常</span>
            </span>
          </div>
        </template>
      </div>
      <div class="score-note">
        <span class="note-icon">ℹ</span>
        高分率超过50%或均分超过90分的课程，建议核查评分标准是否过于宽松；不及格率超过25%的课程，建议重点复查教学质量。
      </div>
    </div>

    <el-dialog v-model="detailVisible" :title="`${detailCourse} · ${detailGrade}`" width="400px">
      <div class="detail-body">
        <div class="detail-rate" :style="{ color: rateColor(detailPassRate) }">{{ detailPassRate }}%</div>
        <div class="detail-label">课程通过率</div>
        <el-progress
          :percentage="detailPassRate" :stroke-width="10"
          :color="rateColor(detailPassRate)"
          style="margin-top: 18px"
        />
        <div class="detail-tip" :style="{ borderLeft: `3px solid ${rateColor(detailPassRate)}` }">
          <template v-if="detailPassRate < 75">通过率偏低，建议安排补充课时或提前预警高风险学生</template>
          <template v-else-if="detailPassRate < 88">表现中等，关注末位学生，建议提前介入</template>
          <template v-else>课程达成情况良好，保持当前教学质量</template>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<style scoped>
.page { padding: 28px 32px; display: flex; flex-direction: column; gap: 20px; }
.page-head { display: flex; align-items: flex-start; justify-content: space-between; }
.term-selector { display: flex; align-items: center; gap: 8px; flex-shrink: 0; }
.term-label { font-size: 12.5px; color: #9facc5; font-weight: 600; }
.term-select {
  padding: 6px 28px 6px 12px; border: 1.5px solid #e5e9f0; border-radius: 9px;
  font-size: 13px; font-weight: 600; color: #374151;
  background: #fff url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 24 24'%3E%3Cpath fill='%239facc5' d='M7 10l5 5 5-5z'/%3E%3C/svg%3E") no-repeat right 8px center;
  appearance: none; cursor: pointer;
}
.term-select:focus { outline: none; border-color: #2A4D99; box-shadow: 0 0 0 3px rgba(42,77,153,0.1); }
.page-title { font-size: 20px; font-weight: 800; color: #101d3e; margin: 0; letter-spacing: -0.01em; }
.page-sub { font-size: 13px; color: #9facc5; margin: 4px 0 0; }

.card {
  background: #fff; border-radius: 16px; padding: 20px 22px;
  border: 1px solid rgba(10,19,41,0.07);
  box-shadow: 0 1px 2px rgba(10,19,41,0.03), 0 4px 16px rgba(10,19,41,0.05);
}
.card-head { display: flex; align-items: center; justify-content: space-between; margin-bottom: 12px; }
.card-title { font-size: 14px; font-weight: 700; color: #111827; }

.legend-inline { display: flex; align-items: center; gap: 8px; }
.legend-dot { width: 8px; height: 8px; border-radius: 50%; flex-shrink: 0; }
.legend-text { font-size: 11.5px; color: #6b7280; margin-right: 4px; }

.tips-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 14px; }
.tip-card {
  border-radius: 14px; padding: 16px 18px;
  border: 1px solid;
}
.tip-card.warn { background: #fef9f9; border-color: #fecaca; }
.tip-card.good { background: #f6fdf9; border-color: #bbf7d0; }

.tip-head { display: flex; align-items: center; gap: 8px; margin-bottom: 10px; }
.tip-icon {
  width: 20px; height: 20px; border-radius: 50%;
  display: flex; align-items: center; justify-content: center;
  background: #e04538; color: #fff; font-size: 11px; font-weight: 800; flex-shrink: 0;
}
.good-icon { background: #3db87e; }
.tip-title { font-size: 13.5px; font-weight: 700; color: #111827; }

.tip-list { margin: 0; padding-left: 0; list-style: none; display: flex; flex-direction: column; gap: 6px; }
.tip-list li { font-size: 12.5px; color: #6b7280; line-height: 1.5; padding-left: 12px; position: relative; }
.tip-list li::before { content: ''; position: absolute; left: 0; top: 7px; width: 4px; height: 4px; border-radius: 50%; background: #d1d5db; }

.detail-body { padding: 8px 0 16px; text-align: center; }
.detail-rate { font-size: 52px; font-weight: 800; line-height: 1; letter-spacing: -0.02em; }
.detail-label { font-size: 13px; color: #9ca3af; margin-top: 4px; }
.detail-tip {
  margin-top: 16px; padding: 10px 14px;
  background: #f8f9fc; border-radius: 8px;
  font-size: 13px; color: #4b5563; line-height: 1.6;
  text-align: left;
}

.card-badge-ai {
  font-size: 11px; font-weight: 600; color: #2A4D99;
  background: rgba(42,77,153,0.08); padding: 3px 9px; border-radius: 10px;
}

.health-list { display: flex; flex-direction: column; gap: 12px; }
.health-item {
  display: grid; grid-template-columns: 200px 1fr 1fr;
  align-items: center; gap: 16px;
  padding: 12px 14px; border-radius: 12px;
  border: 1px solid #f0f2f7; transition: background 0.15s;
}
.health-item:hover { background: #fafbfd; }

.health-item-left { display: flex; align-items: center; gap: 8px; }
.health-course { font-size: 13.5px; font-weight: 700; color: #1f2937; }
.health-status-tag {
  font-size: 11px; font-weight: 700;
  padding: 2px 8px; border-radius: 20px;
}

.health-bar-wrap { display: flex; align-items: center; gap: 10px; }
.health-bar-track {
  flex: 1; height: 7px; background: #f0f2f7;
  border-radius: 99px; overflow: hidden;
}
.health-bar-fill {
  height: 100%; border-radius: 99px;
  transition: width 0.5s cubic-bezier(0.34,1.56,0.64,1);
}
.health-score { font-size: 15px; font-weight: 800; min-width: 28px; text-align: right; }

.health-suggestion { font-size: 12px; color: #6b7280; line-height: 1.5; }
.pass-rate-list { display: flex; flex-direction: column; gap: 14px; }
.pass-rate-row { display: flex; align-items: center; gap: 16px; }
.pr-left { width: 220px; flex-shrink: 0; }
.pr-course { font-size: 13.5px; font-weight: 700; color: #111827; display: block; }
.pr-suggestion { font-size: 12px; color: #9facc5; margin-top: 2px; display: block; }
.pr-bar-wrap { flex: 1; display: flex; align-items: center; gap: 10px; }
.pr-track { flex: 1; height: 10px; background: #f0f2f7; border-radius: 99px; overflow: hidden; }
.pr-fill { height: 100%; border-radius: 99px; transition: width 0.5s ease; }
.pr-val { font-size: 14px; font-weight: 800; width: 42px; text-align: right; font-variant-numeric: tabular-nums; }

.score-table { display: flex; flex-direction: column; }
.score-thead {
  display: grid; grid-template-columns: 1.2fr 1fr 0.8fr 0.8fr 1fr 0.8fr 1.4fr;
  padding: 8px 10px; background: #f8f9fc; border-radius: 8px;
  font-size: 12px; font-weight: 700; color: #9facc5; margin-bottom: 4px;
}
.score-trow {
  display: grid; grid-template-columns: 1.2fr 1fr 0.8fr 0.8fr 1fr 0.8fr 1.4fr;
  padding: 13px 10px; border-bottom: 1px solid #f0f2f7;
  font-size: 13px; align-items: center;
  transition: background 0.12s;
}
.score-trow:hover { background: #f8f9fc; }
.score-trow:last-child { border-bottom: none; }
.score-thead {
  display: grid; grid-template-columns: 1.4fr 0.8fr 0.8fr 1fr 0.8fr 1.2fr;
  padding: 8px 10px; background: #f8f9fc; border-radius: 8px;
  font-size: 12px; font-weight: 700; color: #9facc5; margin-bottom: 4px;
}
.score-trow {
  display: grid; grid-template-columns: 1.4fr 0.8fr 0.8fr 1fr 0.8fr 1.2fr;
  padding: 10px 10px; border-bottom: 1px solid #f0f2f7;
  font-size: 13px; align-items: center;
}
.score-trow:last-child { border-bottom: none; }
.score-trow:hover { background: #f8f9fc; }
.score-course-row {
  display: grid; grid-template-columns: 1.4fr 0.8fr 0.8fr 1fr 0.8fr 1.2fr;
  padding: 10px 10px 6px; background: #f4f6fa;
  border-top: 2px solid #edf0f6; align-items: center;
}
.score-course-name { font-size: 14px; font-weight: 800; color: #111827; display: flex; align-items: center; gap: 8px; }
.score-teacher-inline { font-size: 12px; font-weight: 400; color: #9facc5; }
.score-grade-cell { font-size: 12.5px; color: #374151; font-weight: 600; padding-left: 8px; }

.risk-tag {
  display: inline-block; font-size: 11.5px; font-weight: 600;
  padding: 3px 8px; border-radius: 6px;
}
.risk-high { color: #f09a4e; background: rgba(240,154,78,0.1); }
.risk-low  { color: #e04538; background: rgba(224,69,56,0.08); }
.risk-ok   { color: #3db87e; background: rgba(61,184,126,0.08); }

.score-note {
  margin-top: 14px; padding: 10px 14px;
  background: #f8f9fc; border-radius: 8px;
  font-size: 12.5px; color: #6b7280; line-height: 1.6;
  display: flex; gap: 8px;
}
.note-icon { color: #9facc5; flex-shrink: 0; font-style: normal; }
</style>
