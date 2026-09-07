<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getCourseTeacherCourses, getCourseTeacherTerms, interpretMetrics } from '@/api'
import type { CourseTeacherCourse } from '@/types'

interface DecoratedCourse extends CourseTeacherCourse {
  semester: string
  aiInsight: string
  aiPoints: string[]
  color: string
  bg: string
  classNames: string
  classStats: CourseTeacherCourse['classes']
}

const courses = ref<DecoratedCourse[]>([])
const route = useRoute()
const router = useRouter()
const loading = ref(false)
const selectedCourse = ref<DecoratedCourse | null>(null)
const aiLoading = ref(false)
const aiSummary = ref('正在基于课程数据生成 AI 学情诊断...')
const aiPoints = ref<string[]>([])
const aiCache = new Map<string, { summary: string; points: string[] }>()
const filterMode = ref<'all' | 'excellent' | 'attention' | 'highRisk'>('all')
const keyword = ref('')
const scoreLabels = ['0-59', '60-69', '70-79', '80-89', '90-94', '95-100']

const termOptions = ref<string[]>([])
const selectedTerm = ref('')

function decorate(course: CourseTeacherCourse): DecoratedCourse {
  const color = course.avgScore >= 80 ? '#3db87e' : course.avgScore >= 70 ? '#2A4D99' : '#e04538'
  const shown = course.classes.slice(0, 3).map(s => s.className)
  const classNames = course.classes.length > 3
    ? `${shown.join(' / ')} 等 ${course.classes.length} 个班级`
    : shown.join(' / ')
  return {
    ...course,
    semester: course.term,
    classStats: course.classes,
    classNames,
    aiInsight: `${course.name}本学期均分 ${course.avgScore} 分，通过率 ${course.passRate}%，预警学生 ${course.highRiskCount} 人。`,
    aiPoints: course.classes
      .map(cls => `${cls.className}均分 ${cls.avg}，通过率 ${cls.pass}%，挂科 ${cls.failStudents.length} 人`)
      .slice(0, 4),
    color,
    bg: 'rgba(42,77,153,0.08)',
  }
}

async function loadAiCourse(course: DecoratedCourse) {
  if (!course) return
  const cacheKey = `${course.id}|${course.term}`
  const cached = aiCache.get(cacheKey)
  if (cached) {
    aiSummary.value = cached.summary
    aiPoints.value = cached.points
    return
  }
  aiLoading.value = true
  try {
    const res = await interpretMetrics(
      '请基于以下课程成绩与综合预警数据生成简明的AI学情诊断：highRiskCount和班级risk表示该课程选课学生中的预警人数，包含黄色、橙色、红色预警，并非仅指挂科人数。第一行输出一句总体评价，后续每行用“- ”开头输出分析或建议，最多5条；不要出现#、**、反引号，不要编造数据。',
      {
        name: course.name,
        term: course.term,
        students: course.students,
        avgScore: course.avgScore,
        passRate: course.passRate,
        highRiskCount: course.highRiskCount,
        scoreDistribution: course.scoreDistribution,
        classes: course.classes.map(c => ({
          className: c.className,
          avg: c.avg,
          pass: c.pass,
          high: c.high,
          low: c.low,
          count: c.count,
          risk: c.risk,
          failStudents: c.failStudents.length,
        })),
      },
    )
    const clean = (s: string) => s.replace(/[#*`]/g, '').trim()
    const lines = res.content.split('\n').map(clean).filter(Boolean)
    const bullets = lines.filter(l => /^[-•]/.test(l)).map(l => l.replace(/^[-•]\s*/, ''))
    if (bullets.length > 0) {
      aiSummary.value = lines.find(l => !/^[-•]/.test(l)) || 'AI 学情诊断'
      aiPoints.value = bullets.slice(0, 5)
    } else {
      aiSummary.value = res.content.replace(/[#*`]/g, '').trim()
      aiPoints.value = []
    }
    aiCache.set(cacheKey, { summary: aiSummary.value, points: aiPoints.value })
  } catch {
    aiSummary.value = 'AI 学情诊断生成失败，请稍后重试'
    aiPoints.value = []
  } finally {
    aiLoading.value = false
  }
}

async function loadCourses() {
  if (!selectedTerm.value) return
  loading.value = true
  try {
    const list = await getCourseTeacherCourses(selectedTerm.value)
    courses.value = list.map(decorate)
    selectedCourse.value = filteredCourses.value[0] ?? null
  } finally {
    loading.value = false
  }
}

async function initialize() {
  loading.value = true
  try {
    termOptions.value = await getCourseTeacherTerms()
    const queryTerm = typeof route.query.term === 'string' ? route.query.term : ''
    selectedTerm.value = termOptions.value.includes(queryTerm) ? queryTerm : (termOptions.value[0] ?? '')
  } finally {
    loading.value = false
  }
}

const categoryCounts = computed(() => ({
  all: courses.value.length,
  excellent: courses.value.filter(c => c.avgScore >= 80).length,
  attention: courses.value.filter(c => c.avgScore >= 70 && c.avgScore < 80).length,
  highRisk: courses.value.filter(c => c.highRiskCount >= 20 || c.passRate < 75).length,
}))

const filteredCourses = computed(() => {
  let list = courses.value
  if (filterMode.value === 'excellent') list = list.filter(c => c.avgScore >= 80)
  if (filterMode.value === 'attention') list = list.filter(c => c.avgScore >= 70 && c.avgScore < 80)
  if (filterMode.value === 'highRisk') list = list.filter(c => c.highRiskCount >= 20 || c.passRate < 75)
  const kw = keyword.value.trim()
  if (kw) list = list.filter(c => c.name.includes(kw))
  return list
})

const filters = [
  { key: 'all', label: '全部课程', color: '#2A4D99' },
  { key: 'excellent', label: '表现优秀', color: '#3db87e' },
  { key: 'attention', label: '需要关注', color: '#f09a4e' },
  { key: 'highRisk', label: '高风险', color: '#e04538' },
] as const

function statusLabel(c: DecoratedCourse) {
  if (c.avgScore >= 80) return { label: '优秀', color: '#3db87e', bg: 'rgba(61,184,126,0.1)' }
  if (c.passRate < 75 || c.highRiskCount >= 20) return { label: '高风险', color: '#e04538', bg: 'rgba(224,69,56,0.1)' }
  if (c.avgScore >= 70) return { label: '关注', color: '#f09a4e', bg: 'rgba(240,154,78,0.1)' }
  return { label: '预警', color: '#ca8a04', bg: 'rgba(202,138,4,0.1)' }
}

function barHeight(val: number) {
  if (!selectedCourse.value || selectedCourse.value.scoreDistribution.length === 0) return 0
  const max = Math.max(...selectedCourse.value.scoreDistribution)
  return max === 0 ? 0 : Math.round((val / max) * 72)
}

onMounted(initialize)
watch(selectedTerm, async term => {
  if (!term) return
  await router.replace({ query: { ...route.query, term } })
  loadCourses()
})
watch(selectedCourse, c => {
  if (c) loadAiCourse(c)
})
watch([filterMode, keyword], () => {
  if (!selectedCourse.value || !filteredCourses.value.some(c => c.id === selectedCourse.value?.id)) {
    selectedCourse.value = filteredCourses.value[0] ?? null
  }
})
</script>

<template>
  <div v-loading="loading" class="page">
    <div class="page-head">
      <div>
        <h2 class="page-title">我的课程</h2>
        <p class="page-sub">本学期共承担 {{ courses.length }} 门课程，合计 {{ courses.reduce((s,c) => s+c.students, 0) }} 选课人次</p>
      </div>
      <el-select v-model="selectedTerm" placeholder="选择学期" style="width: 180px">
        <el-option v-for="t in termOptions" :key="t" :label="t" :value="t" />
      </el-select>
    </div>

    <div class="filter-bar">
      <div class="filter-tabs">
        <button
          v-for="f in filters" :key="f.key"
          class="filter-tab"
          :class="{ active: filterMode === f.key }"
          :style="filterMode === f.key ? { color: f.color, borderColor: f.color, background: `${f.color}12` } : {}"
          @click="filterMode = f.key"
        >
          {{ f.label }}
          <span class="filter-count" :style="filterMode === f.key ? { color: f.color } : {}">{{ categoryCounts[f.key] }}</span>
        </button>
      </div>
      <el-input v-model="keyword" placeholder="搜索课程名称" clearable style="width: 220px" />
    </div>

    <div v-if="!loading && courses.length === 0" class="empty-tip">当前学期暂无课程数据</div>
    <div v-else-if="!loading && filteredCourses.length === 0" class="empty-tip">当前筛选条件下暂无课程</div>

    <template v-if="selectedCourse">
    <!-- 课程卡片 -->
    <div class="course-list">
      <div
        v-for="c in filteredCourses" :key="c.id"
        class="course-card"
        :class="{ active: selectedCourse.id === c.id }"
        :style="selectedCourse.id === c.id ? { borderColor: c.color, background: c.bg } : {}"
        @click="selectedCourse = c"
      >
        <div class="course-card-top">
          <div>
            <div class="course-name">{{ c.name }}</div>
            <div class="course-code">{{ c.code }} · {{ c.classNames }}</div>
          </div>
          <div class="course-students">
            <span class="course-status" :style="{ color: statusLabel(c).color, background: statusLabel(c).bg }">{{ statusLabel(c).label }}</span>
            <div class="course-count">
              <span class="cs-num">{{ c.students }}</span>
              <span class="cs-label">人</span>
            </div>
          </div>
        </div>
        <div class="course-stats">
          <div class="cs-stat">
            <div class="cs-stat-label">均分</div>
            <div class="cs-stat-val" :style="{ color: c.color }">{{ c.avgScore }}</div>
          </div>
          <div class="cs-stat-divider" />
          <div class="cs-stat">
            <div class="cs-stat-label">通过率</div>
            <div class="cs-stat-val" :style="{ color: c.passRate >= 90 ? '#3db87e' : c.passRate >= 80 ? '#f09a4e' : '#e04538' }">{{ c.passRate }}%</div>
          </div>
          <div class="cs-stat-divider" />
          <div class="cs-stat">
            <div class="cs-stat-label">预警学生</div>
            <div class="cs-stat-val" :style="{ color: c.highRiskCount >= 10 ? '#e04538' : '#f09a4e' }">{{ c.highRiskCount }}人</div>
          </div>
        </div>
      </div>
    </div>

    <!-- AI 诊断 + 成绩分布 并排 -->
    <div class="ai-dist-row">
      <div class="ai-panel">
        <div class="ai-panel-header">
          <div class="ai-badge">AI</div>
          <div class="ai-header-text">
            <div class="ai-panel-title">{{ selectedCourse.name }} · AI 学情诊断</div>
            <div class="ai-panel-sub">基于期末/期中成绩综合分析</div>
          </div>
        </div>
        <p class="ai-summary" v-if="aiLoading">正在生成 AI 学情诊断...</p>
        <p v-else class="ai-summary">{{ aiSummary }}</p>
        <ul class="ai-points">
          <li v-for="(pt, i) in aiPoints" :key="i">{{ pt }}</li>
        </ul>
      </div>

      <div class="card dist-card">
        <div class="card-title">{{ selectedCourse.name }} · 成绩分布</div>
        <div class="bar-chart">
          <div v-for="(val, i) in selectedCourse.scoreDistribution" :key="i" class="bar-col">
            <div class="bar-val">{{ val }}</div>
            <div class="bar-wrap">
              <div class="bar-fill" :style="{ height: barHeight(val) + 'px', background: selectedCourse.color }" />
            </div>
            <div class="bar-label">{{ scoreLabels[i] }}</div>
          </div>
        </div>
      </div>
    </div>

    <!-- 班级对比 -->
    <div class="card">
      <div class="cls-table-head-row">
        <div class="card-title" style="margin-bottom:0">{{ selectedCourse.name }} · 班级成绩对比</div>
        <div class="cls-meta">{{ selectedCourse.semester }} · 共 {{ selectedCourse.students }} 名学生</div>
      </div>

      <!-- 表头 -->
      <div class="cls-table">
        <div class="cls-thead">
          <span>班级</span>
          <span>均分</span>
          <span>通过率</span>
          <span>最高分</span>
          <span>最低分</span>
          <span>人数</span>
          <span>预警学生</span>
          <span>均分可视化</span>
        </div>
        <div v-for="cls in selectedCourse.classStats" :key="cls.className" class="cls-trow">
          <span class="cls-name-cell">{{ cls.className }}</span>
          <span class="cls-avg-cell" :style="{ color: cls.avg >= selectedCourse.avgScore ? '#3db87e' : '#f09a4e' }">{{ cls.avg }}</span>
          <span class="cls-pass-cell" :style="{ color: cls.pass >= 90 ? '#3db87e' : cls.pass >= 80 ? '#2A4D99' : '#f09a4e' }">{{ cls.pass }}%</span>
          <span class="cls-hi-cell" style="color:#3db87e">{{ cls.high }}</span>
          <span class="cls-lo-cell" :style="{ color: cls.low < 60 ? '#e04538' : '#9facc5' }">{{ cls.low }}</span>
          <span class="cls-cnt-cell">{{ cls.count }}人</span>
          <span class="cls-risk-cell" :style="{ color: cls.risk >= 5 ? '#e04538' : '#f09a4e' }">{{ cls.risk }}人</span>
          <div class="cls-bar-cell">
            <div class="cls-bar-track">
              <div class="cls-bar-fill"
                :style="{ width: ((cls.avg - 50) / 50 * 100) + '%', background: selectedCourse.color }" />
            </div>
            <span class="cls-bar-num">{{ cls.avg }}</span>
          </div>
        </div>
      </div>

      <!-- 各班本学期挂科学生 -->
      <div class="fail-section">
        <div class="fail-title">各班本学期挂科学生名单</div>
        <div class="fail-grid">
          <div v-for="cls in selectedCourse.classStats" :key="cls.className" class="fail-cls-col">
            <div class="fail-cls-header">
              <span class="fail-cls-name">{{ cls.className }}</span>
              <span class="fail-cls-count" :style="{ color: cls.termFailStudents.length >= 5 ? '#e04538' : '#f09a4e' }">
                {{ cls.termFailStudents.length }}人本学期有挂科
              </span>
            </div>
            <div v-if="cls.termFailStudents.length === 0" class="fail-empty">本学期暂无挂科</div>
            <div v-else class="fail-list">
              <div v-for="s in cls.termFailStudents" :key="s.studentId" class="fail-item">
                <div class="fail-avatar">{{ s.name[0] }}</div>
                <div class="fail-info">
                  <span class="fail-name">{{ s.name }}</span>
                  <span class="fail-id">{{ s.studentId }}</span>
                  <span class="fail-courses">{{ s.failedCourses.join('、') }}</span>
                </div>
                <span v-if="s.lowestScore !== null" class="fail-score" :style="{ color: s.lowestScore < 50 ? '#e04538' : '#f09a4e' }">
                  最低{{ s.lowestScore }}分
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
    </template>
  </div>
</template>

<style scoped>
.page { padding: 28px 32px; display: flex; flex-direction: column; gap: 20px; }
.page-head { display: flex; align-items: flex-start; justify-content: space-between; gap: 12px; }
.page-title { font-size: 20px; font-weight: 800; color: #101d3e; margin: 0; letter-spacing: -0.01em; }
.page-sub { font-size: 13px; color: #9facc5; margin: 4px 0 0; }
.empty-tip { font-size: 13px; color: #9facc5; padding: 20px 0; text-align: center; }
.filter-bar { display: flex; align-items: center; justify-content: space-between; gap: 12px; flex-wrap: wrap; }
.filter-tabs { display: flex; gap: 8px; flex-wrap: wrap; }
.filter-tab {
  display: flex; align-items: center; gap: 6px;
  padding: 7px 14px; border-radius: 9px;
  border: 1.5px solid #e5e9f0; background: #fff;
  font-size: 12.5px; font-weight: 600; color: #6b7280;
  cursor: pointer; transition: all 0.15s ease;
}
.filter-tab:hover { border-color: #c8d4e8; }
.filter-tab.active { box-shadow: 0 2px 8px rgba(10,19,41,0.06); }
.filter-count { font-size: 11.5px; font-weight: 800; color: #9facc5; }
.course-list { display: grid; grid-template-columns: repeat(3, 1fr); gap: 14px; }
.course-card { background: #fff; border-radius: 14px; padding: 18px 20px; border: 2px solid #edf0f6; cursor: pointer; transition: all 0.18s ease; box-shadow: 0 1px 4px rgba(0,0,0,0.04); }
.course-card:hover { border-color: #c8d4e8; }
.course-card.active { box-shadow: 0 4px 16px rgba(0,0,0,0.08); }
.course-card-top { display: flex; align-items: flex-start; justify-content: space-between; margin-bottom: 14px; }
.course-name { font-size: 15px; font-weight: 700; color: #111827; }
.course-code { font-size: 11.5px; color: #9facc5; margin-top: 3px; }
.course-students { display: flex; flex-direction: column; align-items: flex-end; gap: 6px; flex-shrink: 0; }
.course-status { font-size: 11px; font-weight: 700; padding: 3px 9px; border-radius: 6px; white-space: nowrap; }
.course-count { display: flex; align-items: baseline; gap: 2px; }
.cs-num { font-size: 26px; font-weight: 800; color: #111827; }
.cs-label { font-size: 12px; color: #9facc5; }
.course-stats { display: flex; align-items: center; gap: 0; margin-top: 12px; padding-top: 12px; border-top: 1px solid rgba(255,255,255,0.08); }
.cs-stat { flex: 1; text-align: center; }
.cs-stat-label { font-size: 11px; color: #9facc5; margin-bottom: 3px; }
.cs-stat-val { font-size: 18px; font-weight: 800; line-height: 1; }
.cs-stat-divider { width: 1px; height: 28px; background: rgba(0,0,0,0.08); flex-shrink: 0; }
.ai-dist-row { display: grid; grid-template-columns: 1fr 300px; gap: 16px; align-items: stretch; }
.ai-panel { background: linear-gradient(135deg, #0c1a38 0%, #1a2d5a 60%, #2A4D99 100%); border-radius: 16px; padding: 20px 24px; display: flex; flex-direction: column; justify-content: center; }
.ai-panel-header { display: flex; align-items: center; gap: 12px; margin-bottom: 12px; flex-wrap: wrap; }
.ai-badge { width: 30px; height: 30px; border-radius: 8px; flex-shrink: 0; background: rgba(255,255,255,0.15); color: #fff; display: flex; align-items: center; justify-content: center; font-size: 12px; font-weight: 800; }
.ai-header-text { flex: 1; }
.ai-panel-title { font-size: 14px; font-weight: 700; color: #fff; }
.ai-panel-sub { font-size: 12px; color: rgba(255,255,255,0.45); margin-top: 2px; }
.health-pill { font-size: 12px; font-weight: 700; padding: 4px 12px; border-radius: 20px; flex-shrink: 0; }
.ai-summary { font-size: 13px; color: rgba(255,255,255,0.82); line-height: 1.7; margin: 0 0 12px; }
.ai-points { margin: 0; padding: 0; list-style: none; display: flex; flex-direction: column; gap: 5px; }
.ai-points li { font-size: 12.5px; color: rgba(255,255,255,0.65); padding-left: 14px; position: relative; line-height: 1.5; }
.ai-points li::before { content: ''; position: absolute; left: 0; top: 7px; width: 4px; height: 4px; border-radius: 50%; background: #5580d4; }
.card { background: #fff; border-radius: 14px; padding: 20px 22px; border: 1px solid rgba(10,19,41,0.07); box-shadow: 0 1px 2px rgba(10,19,41,0.03), 0 4px 16px rgba(10,19,41,0.05); }
.card-title { font-size: 14px; font-weight: 700; color: #111827; margin-bottom: 16px; }
.dist-card { display: flex; flex-direction: column; }
.bar-chart { display: flex; align-items: flex-end; gap: 6px; padding: 4px 0; flex: 1; }
.bar-col { display: flex; flex-direction: column; align-items: center; gap: 4px; flex: 1; }
.bar-val { font-size: 11px; font-weight: 600; color: #374151; }
.bar-wrap { width: 100%; display: flex; align-items: flex-end; height: 72px; }
.bar-fill { width: 100%; border-radius: 4px 4px 0 0; transition: height 0.4s ease; min-height: 3px; }
.bar-label { font-size: 9.5px; color: #9facc5; white-space: nowrap; }
.cls-table-head-row { display: flex; align-items: center; justify-content: space-between; margin-bottom: 16px; }
.cls-meta { font-size: 12px; color: #9facc5; }
.cls-table { display: flex; flex-direction: column; gap: 0; margin-bottom: 20px; }
.cls-thead { display: grid; grid-template-columns: 1.4fr 0.8fr 0.9fr 0.7fr 0.7fr 0.7fr 0.7fr 2fr; padding: 8px 10px; background: #f8f9fc; border-radius: 8px; font-size: 12px; font-weight: 700; color: #9facc5; margin-bottom: 4px; }
.cls-trow { display: grid; grid-template-columns: 1.4fr 0.8fr 0.9fr 0.7fr 0.7fr 0.7fr 0.7fr 2fr; padding: 12px 10px; border-bottom: 1px solid #f0f2f7; font-size: 13px; align-items: center; transition: background 0.12s; }
.cls-trow:hover { background: #f8f9fc; }
.cls-trow:last-child { border-bottom: none; }
.cls-name-cell { font-weight: 700; color: #111827; }
.cls-avg-cell { font-weight: 800; font-size: 14px; }
.cls-pass-cell, .cls-hi-cell, .cls-lo-cell { font-weight: 600; }
.cls-cnt-cell { color: #6b7280; }
.cls-risk-cell { font-weight: 700; }
.cls-bar-cell { display: flex; align-items: center; gap: 8px; }
.cls-bar-track { flex: 1; height: 6px; background: #f0f2f7; border-radius: 99px; overflow: hidden; }
.cls-bar-fill { height: 100%; border-radius: 99px; transition: width 0.5s ease; }
.cls-bar-num { font-size: 12px; font-weight: 700; color: #374151; width: 32px; text-align: right; }
.cls-chart-row { display: none; }
.cls-chart-col { display: none; }
.cls-chart-bar-wrap { display: none; }
.cls-chart-bar { display: none; }
.cls-chart-label { display: none; }
.cls-chart-val { display: none; }
.cls-chart-avg-line { display: none; }
.cls-chart-avg-tag { display: none; }

/* 挂科学生 */
.fail-section { border-top: 1px solid #f0f2f7; padding-top: 18px; }
.fail-title { font-size: 13.5px; font-weight: 700; color: #374151; margin-bottom: 14px; }
.fail-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 14px; }
.fail-cls-col { background: #f9fafb; border-radius: 10px; padding: 14px 16px; }
.fail-cls-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 10px; }
.fail-cls-name { font-size: 13px; font-weight: 700; color: #111827; }
.fail-cls-count { font-size: 12px; font-weight: 700; }
.fail-empty { font-size: 12.5px; color: #3db87e; font-weight: 600; }
.fail-list { display: flex; flex-direction: column; gap: 7px; }
.fail-item { display: flex; align-items: center; gap: 8px; background: #fff; border-radius: 8px; padding: 7px 10px; border: 1px solid #f0f2f7; }
.fail-avatar { width: 26px; height: 26px; border-radius: 50%; background: linear-gradient(135deg, #f09a4e, #f5c518); display: flex; align-items: center; justify-content: center; color: #fff; font-size: 11px; font-weight: 700; flex-shrink: 0; }
.fail-info { flex: 1; display: flex; flex-direction: column; gap: 1px; }
.fail-name { font-size: 13px; font-weight: 600; color: #111827; }
.fail-id { font-size: 10.5px; color: #9facc5; }
.fail-courses { font-size: 11px; color: #b45309; margin-top: 2px; }
.fail-score { font-size: 13px; font-weight: 800; flex-shrink: 0; }
</style>
