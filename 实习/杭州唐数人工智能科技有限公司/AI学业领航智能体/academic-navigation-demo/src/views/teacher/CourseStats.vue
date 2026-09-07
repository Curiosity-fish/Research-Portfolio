<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getTeacherCourseStats, getTeacherTerms } from '@/api'
import type { CourseStatsItem } from '@/types'

const courses = ref<CourseStatsItem[]>([])
const loading = ref(false)
const route = useRoute()
const router = useRouter()

async function loadStats() {
  loading.value = true
  try {
    courses.value = await getTeacherCourseStats(selectedTerm.value)
  } finally {
    loading.value = false
  }
}

const expandedId = ref<string | null>(null)

function toggleExpand(id: string) {
  expandedId.value = expandedId.value === id ? null : id
}

const termOptions = ref<string[]>([])
const selectedTerm = ref('')
const className = ref('')

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
  } finally {
    loading.value = false
  }
}

function failColor(rate: number) {
  if (rate >= 30) return '#e04538'
  if (rate >= 15) return '#f09a4e'
  return '#3db87e'
}
function failBg(rate: number) {
  if (rate >= 30) return 'rgba(224,69,56,0.08)'
  if (rate >= 15) return 'rgba(240,154,78,0.08)'
  return 'rgba(61,184,126,0.08)'
}
function scoreColor(s: number) {
  if (s >= 85) return '#3db87e'
  if (s < 70)  return '#e04538'
  return '#2A4D99'
}

// 汇总：挂科最多的学生
const failCountMap = computed(() => {
  const m: Record<string, { name: string; count: number; courses: string[] }> = {}
  for (const c of courses.value) {
    for (const s of c.failStudents) {
      if (!m[s.studentId]) m[s.studentId] = { name: s.name, count: 0, courses: [] }
      m[s.studentId].count++
      m[s.studentId].courses.push(c.name)
    }
  }
  return Object.entries(m)
    .map(([id, v]) => ({ studentId: id, ...v }))
    .sort((a, b) => b.count - a.count)
    .slice(0, 5)
})

onMounted(initialize)
watch(selectedTerm, async term => {
  if (!term) return
  await router.replace({ query: { ...route.query, term } })
  loadStats()
})
</script>

<template>
  <div v-loading="loading" class="page">
    <div class="page-head">
      <div>
        <h2 class="page-title">课程成绩统计</h2>
        <p class="page-sub">{{ formatTerm(selectedTerm) }} · {{ className }} · {{ courses.length }} 门课程</p>
      </div>
      <div class="term-selector">
        <span class="term-selector-label">学期</span>
        <select v-if="termOptions.length > 1" v-model="selectedTerm" class="term-select">
          <option v-for="t in termOptions" :key="t" :value="t">{{ t }}</option>
        </select>
        <span v-else class="term-static">{{ selectedTerm }}</span>
      </div>
    </div>

    <!-- 重点关注汇总 -->
    <div class="card summary-card">
      <div class="card-title-row">
        <span class="card-title">重点关注学生（挂科 ≥ 2 门）</span>
        <span class="badge-red">{{ failCountMap.filter(s => s.count >= 2).length }} 人</span>
      </div>
      <div v-if="failCountMap.filter(s => s.count >= 2).length === 0" class="empty-tip">暂无学生挂科 2 门及以上</div>
      <div v-else class="fail-student-list">
        <div v-for="s in failCountMap.filter(s => s.count >= 2)" :key="s.studentId" class="fail-student-row">
          <div class="fs-avatar">{{ s.name[0] }}</div>
          <div class="fs-info">
            <span class="fs-name">{{ s.name }}</span>
            <span class="fs-id">{{ s.studentId }}</span>
          </div>
          <div class="fs-courses">
            <span v-for="c in s.courses" :key="c" class="course-chip">{{ c }}</span>
          </div>
          <span class="fail-count-badge">挂科 {{ s.count }} 门</span>
        </div>
      </div>
    </div>

    <!-- 课程列表 -->
    <div class="card table-card">
      <div class="card-title-row" style="margin-bottom: 16px">
        <span class="card-title">各课程成绩概览</span>
        <span class="tip-text">点击课程行展开挂科学生名单</span>
      </div>

      <div class="course-list">
        <div
          v-for="c in courses" :key="c.id"
          class="course-block"
        >
          <!-- 主行 -->
          <div
            class="course-row"
            :class="{ expanded: expandedId === c.id, 'has-fail': c.failStudents.length > 0 }"
            @click="c.failStudents.length > 0 ? toggleExpand(c.id) : null"
          >
            <div class="course-name-col">
              <span class="course-name">{{ c.name }}</span>
              <span class="course-credits">{{ c.credits }} 学分</span>
            </div>

            <div class="course-metric">
              <span class="metric-label">平均分</span>
              <span class="metric-val" :style="{ color: scoreColor(c.avgScore) }">{{ c.avgScore }}</span>
            </div>

            <div class="course-metric">
              <span class="metric-label">通过率</span>
              <div class="bar-wrap">
                <div class="bar-track">
                  <div
                    class="bar-fill"
                    :style="{ width: c.passRate + '%', background: c.passRate >= 90 ? '#3db87e' : c.passRate >= 75 ? '#2A4D99' : '#f09a4e' }"
                  />
                </div>
                <span class="bar-pct" :style="{ color: c.passRate >= 90 ? '#3db87e' : c.passRate >= 75 ? '#2A4D99' : '#f09a4e' }">{{ c.passRate }}%</span>
              </div>
            </div>

            <div class="course-metric">
              <span class="metric-label">挂科率</span>
              <span
                class="fail-badge"
                :style="{ color: failColor(c.failRate), background: failBg(c.failRate) }"
              >{{ c.failRate }}%</span>
            </div>

            <div class="course-tail">
              <span v-if="c.failStudents.length > 0" class="fail-count">{{ c.failStudents.length }} 人挂科</span>
              <span v-else class="no-fail">无挂科</span>
              <el-icon v-if="c.failStudents.length > 0" class="expand-icon" :class="{ open: expandedId === c.id }">
                <ArrowRight />
              </el-icon>
            </div>
          </div>

          <!-- 展开：挂科学生列表 -->
          <div v-if="expandedId === c.id && c.failStudents.length > 0" class="fail-list">
            <div class="fail-list-head">
              <span>挂科学生名单</span>
              <span class="fail-list-sub">共 {{ c.failStudents.length }} 人</span>
            </div>
            <div class="fail-items">
              <div v-for="s in c.failStudents" :key="s.studentId" class="fail-item">
                <div class="fi-avatar">{{ s.name[0] }}</div>
                <span class="fi-name">{{ s.name }}</span>
                <span class="fi-id">{{ s.studentId }}</span>
                <span class="fi-score" :style="{ color: s.score < 50 ? '#e04538' : '#f09a4e' }">{{ s.score }} 分</span>
              </div>
            </div>
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
.term-static { font-size: 13px; font-weight: 600; color: #374151; }

.card {
  background: #fff; border-radius: 16px; padding: 20px 22px;
  border: 1px solid rgba(10,19,41,0.07);
  box-shadow: 0 1px 2px rgba(10,19,41,0.03), 0 4px 16px rgba(10,19,41,0.05);
}
.card-title { font-size: 14px; font-weight: 700; color: #111827; }
.card-title-row { display: flex; align-items: center; justify-content: space-between; }
.tip-text { font-size: 12px; color: #9facc5; }
.badge-red { font-size: 11px; font-weight: 600; color: #e04538; background: rgba(224,69,56,0.08); padding: 2px 8px; border-radius: 6px; }

/* Summary */
.empty-tip { font-size: 13px; color: #9facc5; margin-top: 12px; }
.fail-student-list { display: flex; flex-direction: column; gap: 8px; margin-top: 14px; }
.fail-student-row {
  display: flex; align-items: center; gap: 12px;
  padding: 10px 14px; border-radius: 10px;
  background: rgba(224,69,56,0.04); border: 1px solid rgba(224,69,56,0.12);
}
.fs-avatar {
  width: 32px; height: 32px; border-radius: 50%; flex-shrink: 0;
  background: linear-gradient(135deg, #e04538, #f87171);
  display: flex; align-items: center; justify-content: center;
  color: #fff; font-size: 13px; font-weight: 700;
}
.fs-info { display: flex; flex-direction: column; gap: 1px; min-width: 80px; }
.fs-name { font-size: 13.5px; font-weight: 600; color: #111827; }
.fs-id   { font-size: 11px; color: #9facc5; }
.fs-courses { flex: 1; display: flex; gap: 5px; flex-wrap: wrap; }
.course-chip {
  font-size: 11px; color: #374151;
  background: #f0f2f7; padding: 2px 7px; border-radius: 5px;
}
.fail-count-badge {
  font-size: 12px; font-weight: 700; color: #e04538;
  background: rgba(224,69,56,0.1); padding: 3px 9px; border-radius: 6px;
  white-space: nowrap; flex-shrink: 0;
}

/* Course list */
.table-card { padding: 20px 22px; }
.course-list { display: flex; flex-direction: column; gap: 6px; }
.course-block { border-radius: 10px; overflow: hidden; border: 1.5px solid #edf0f6; }

.course-row {
  display: flex; align-items: center; gap: 16px;
  padding: 13px 16px; background: #fff;
  transition: background 0.15s;
}
.course-row.has-fail { cursor: pointer; }
.course-row.has-fail:hover { background: #f8f9fc; }
.course-row.expanded { background: #f4f7fd; border-bottom: 1px solid #e8edf6; }

.course-name-col { flex: 1.5; min-width: 140px; }
.course-name   { font-size: 13.5px; font-weight: 600; color: #111827; display: block; }
.course-credits { font-size: 11px; color: #9facc5; margin-top: 2px; display: block; }

.course-metric { display: flex; flex-direction: column; gap: 4px; min-width: 90px; flex: 1; }
.metric-label { font-size: 11px; color: #9facc5; font-weight: 600; }
.metric-val { font-size: 15px; font-weight: 700; font-variant-numeric: tabular-nums; }

.bar-wrap { display: flex; align-items: center; gap: 8px; }
.bar-track { flex: 1; height: 5px; background: #f0f2f7; border-radius: 99px; overflow: hidden; min-width: 60px; }
.bar-fill  { height: 100%; border-radius: 99px; transition: width 0.4s ease; }
.bar-pct   { font-size: 13px; font-weight: 700; white-space: nowrap; }

.fail-badge {
  font-size: 13px; font-weight: 700;
  padding: 3px 8px; border-radius: 6px;
  display: inline-block; width: fit-content;
}

.course-tail { display: flex; align-items: center; gap: 8px; min-width: 90px; justify-content: flex-end; }
.fail-count { font-size: 12.5px; font-weight: 600; color: #e04538; }
.no-fail    { font-size: 12.5px; color: #3db87e; font-weight: 600; }
.expand-icon { font-size: 13px; color: #9facc5; transition: transform 0.2s; }
.expand-icon.open { transform: rotate(90deg); }

/* Fail student expansion */
.fail-list { padding: 14px 16px; background: #fafbfd; }
.fail-list-head {
  display: flex; align-items: center; gap: 8px;
  font-size: 12.5px; font-weight: 700; color: #374151;
  margin-bottom: 10px;
}
.fail-list-sub { font-size: 11.5px; color: #9facc5; font-weight: 400; }
.fail-items { display: flex; flex-wrap: wrap; gap: 8px; }
.fail-item {
  display: flex; align-items: center; gap: 8px;
  padding: 7px 12px; border-radius: 8px;
  background: #fff; border: 1px solid #fecaca;
}
.fi-avatar {
  width: 26px; height: 26px; border-radius: 50%; flex-shrink: 0;
  background: rgba(224,69,56,0.1);
  display: flex; align-items: center; justify-content: center;
  color: #e04538; font-size: 11px; font-weight: 700;
}
.fi-name  { font-size: 13px; font-weight: 600; color: #111827; }
.fi-id    { font-size: 11px; color: #9facc5; }
.fi-score { font-size: 13px; font-weight: 700; margin-left: 4px; }

.term-selector { display: flex; align-items: center; gap: 8px; }
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
