<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import type { AlertItem } from '@/types'
import WarningBadge from '@/components/common/WarningBadge.vue'
import { getCourseTeacherAlerts, getCourseTeacherCourses, getCourseTeacherTerms } from '@/api'

const alerts = ref<AlertItem[]>([])
const route = useRoute()
const router = useRouter()
const loading = ref(false)
const levelFilter = ref<'all' | 'red' | 'orange' | 'yellow'>('all')
const courseFilter = ref('全部')
const generalAlertLabel = '综合预警'
const taughtCourseNames = ref(new Set<string>())
const termOptions = ref<string[]>([])
const selectedTerm = ref('')

const courses = computed(() => {
  const options = ['全部', ...taughtCourseNames.value]
  if (alerts.value.some(alert => !alert.course)) options.push(generalAlertLabel)
  return options
})

function courseCount(course: string) {
  if (course === '全部') return alerts.value.length
  return alerts.value.filter(alert => (alert.course || generalAlertLabel) === course).length
}

const filtered = computed(() => {
  let list = alerts.value
  if (levelFilter.value !== 'all') list = list.filter(a => a.level === levelFilter.value)
  if (courseFilter.value !== '全部') {
    list = list.filter(a => (a.course || generalAlertLabel) === courseFilter.value)
  }
  return list
})

const counts = computed(() => ({
  red: alerts.value.filter(a => a.level === 'red').length,
  orange: alerts.value.filter(a => a.level === 'orange').length,
  yellow: alerts.value.filter(a => a.level === 'yellow').length,
}))

const levelConfig = {
  red:    { label: '高风险', color: '#e04538', bg: '#fef2f2', border: '#fecaca' },
  orange: { label: '中风险', color: '#f09a4e', bg: '#fff7ed', border: '#fed7aa' },
  yellow: { label: '关注',   color: '#ca8a04', bg: '#fefce8', border: '#fde68a' },
}

async function loadAlerts() {
  if (!selectedTerm.value) return
  loading.value = true
  try {
    const [res, taughtCourses] = await Promise.all([
      getCourseTeacherAlerts(selectedTerm.value, 1, 10000),
      getCourseTeacherCourses(selectedTerm.value).catch(() => []),
    ])
    taughtCourseNames.value = new Set(taughtCourses.map(course => course.name))
    alerts.value = res.list.map(alert => ({ ...alert, failedCourses: alert.failedCourses ?? [] }))
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

onMounted(initialize)
watch(selectedTerm, async term => {
  if (!term) return
  courseFilter.value = '全部'
  levelFilter.value = 'all'
  await router.replace({ query: { ...route.query, term } })
  loadAlerts()
})
</script>

<template>
  <div v-loading="loading" class="page">
    <div class="page-head">
      <div>
        <h2 class="page-title">学情预警</h2>
        <p class="page-sub">{{ selectedTerm }} · 按真实选课关系归属，共 {{ alerts.length }} 条课程预警</p>
      </div>
      <el-select v-if="termOptions.length > 1" v-model="selectedTerm" placeholder="选择学期" style="width: 180px">
        <el-option v-for="term in termOptions" :key="term" :label="term" :value="term" />
      </el-select>
      <span v-else class="term-static">{{ selectedTerm }}</span>
    </div>

    <div class="summary-row">
      <div
        v-for="(cfg, key) in levelConfig" :key="key"
        class="summary-card"
        :class="{ active: levelFilter === key }"
        :style="levelFilter === key ? { borderColor: cfg.color, background: cfg.bg } : {}"
        @click="levelFilter = levelFilter === key ? 'all' : key"
      >
        <div class="summary-num" :style="{ color: cfg.color }">{{ counts[key as keyof typeof counts] }}</div>
        <div class="summary-label">{{ cfg.label }}</div>
      </div>
      <div
        class="summary-card"
        :class="{ active: levelFilter === 'all' }"
        @click="levelFilter = 'all'"
      >
        <div class="summary-num" style="color:#374151">{{ alerts.length }}</div>
        <div class="summary-label">全部预警</div>
      </div>
    </div>

    <div class="filter-bar">
      <div class="filter-tabs">
        <button
          v-for="c in courses" :key="c"
          class="filter-tab" :class="{ active: courseFilter === c }"
          @click="courseFilter = c"
        >{{ c }} <span class="filter-count">{{ courseCount(c) }}</span></button>
      </div>
      <span class="result-count">共 {{ filtered.length }} 条</span>
    </div>

    <div class="alert-list">
      <div v-for="a in filtered" :key="a.id" class="alert-row">
        <div class="alert-level-bar" :style="{ background: levelConfig[a.level].color }" />
        <div class="alert-main">
          <div class="alert-header">
            <span class="alert-name">{{ a.studentName }}</span>
            <span class="alert-id">{{ a.studentId }}</span>
            <span class="alert-course-tag">{{ a.course || generalAlertLabel }}</span>
            <WarningBadge :level="a.level" />
          </div>
          <div class="alert-title">{{ a.title }}</div>
          <div class="alert-desc">{{ a.description }}</div>
          <div v-if="a.failedCourses?.length" class="failed-tags">
            <span v-for="c in a.failedCourses" :key="c" class="failed-tag">{{ c }}</span>
          </div>
          <div class="alert-meta">
            <span v-if="a.triggerEvent">{{ a.triggerEvent }}</span>
            <span>{{ a.pushedAt || a.date }}</span>
            <span class="status-text">{{ a.status }}</span>
          </div>
        </div>
      </div>

      <div v-if="filtered.length === 0" class="empty">
        <el-icon style="font-size:32px;color:#d1d5db"><CircleCheck /></el-icon>
        <div>当前筛选条件下无预警学生</div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.page { padding: 28px 32px; display: flex; flex-direction: column; gap: 20px; }
.page-head { display: flex; align-items: flex-start; justify-content: space-between; gap: 16px; }
.page-title { font-size: 20px; font-weight: 800; color: #101d3e; margin: 0; letter-spacing: -0.01em; }
.page-sub { font-size: 13px; color: #9facc5; margin: 4px 0 0; }
.term-static { font-size: 13px; color: #374151; font-weight: 600; }

.summary-row { display: flex; gap: 12px; }
.summary-card {
  flex: 1; background: #fff; border-radius: 12px; padding: 16px 20px;
  border: 2px solid #edf0f6; cursor: pointer; text-align: center;
  transition: all 0.15s;
}
.summary-card:hover:not(.active) { border-color: #c8d4e8; }
.summary-num { font-size: 28px; font-weight: 800; line-height: 1; }
.summary-label { font-size: 12px; color: #9facc5; margin-top: 6px; }

.filter-bar {
  display: flex; align-items: center; justify-content: space-between; gap: 12px;
  background: #fff; border: 1px solid #edf0f6; border-radius: 12px; padding: 10px 14px;
}
.filter-tabs { display: flex; gap: 6px; flex-wrap: wrap; }
.filter-tab {
  padding: 5px 12px; border: 1px solid #e5e9f0; border-radius: 8px;
  background: #fff; color: #6b7280; font-size: 12.5px; font-weight: 600;
  cursor: pointer; transition: all 0.15s;
}
.filter-tab.active { background: rgba(42,77,153,0.08); border-color: rgba(42,77,153,0.3); color: #2A4D99; }
.filter-count { margin-left: 4px; font-size: 11px; font-weight: 800; color: inherit; }
.result-count { font-size: 12.5px; color: #9facc5; white-space: nowrap; }

.alert-list { display: flex; flex-direction: column; gap: 10px; }
.alert-row {
  display: flex; background: #fff; border: 1px solid #edf0f6; border-radius: 12px;
  overflow: hidden; box-shadow: 0 1px 2px rgba(10,19,41,0.03);
}
.alert-level-bar { width: 4px; flex-shrink: 0; }
.alert-main { flex: 1; min-width: 0; padding: 16px 18px; display: flex; flex-direction: column; gap: 6px; }
.alert-header { display: flex; align-items: center; gap: 10px; flex-wrap: wrap; }
.alert-name { font-size: 15px; font-weight: 700; color: #111827; }
.alert-id { font-size: 12px; color: #9facc5; }
.alert-course-tag {
  font-size: 11.5px; font-weight: 600; color: #2A4D99;
  background: rgba(42,77,153,0.08); padding: 2px 8px; border-radius: 6px;
}
.alert-title { font-size: 13.5px; font-weight: 600; color: #1f2937; }
.alert-desc { font-size: 12.5px; color: #6b7280; line-height: 1.6; }
.failed-tags { display: flex; gap: 5px; flex-wrap: wrap; }
.failed-tag {
  font-size: 11px; color: #b45309; background: #fef3c7;
  border: 1px solid #fde68a; padding: 1px 7px; border-radius: 5px;
}
.alert-meta {
  display: flex; gap: 14px; margin-top: 4px; font-size: 12px; color: #9ca3af; flex-wrap: wrap;
}
.status-text { text-transform: capitalize; }
.empty {
  display: flex; flex-direction: column; align-items: center; gap: 10px;
  padding: 50px 0; color: #9facc5; font-size: 13px; background: #fff;
  border: 1px solid #edf0f6; border-radius: 12px;
}
</style>
