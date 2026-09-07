<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getDepartmentAlerts, getDepartmentTerms } from '@/api'
import { useAuthStore } from '@/stores/auth'
import type { DepartmentAlertItem } from '@/types'

const auth = useAuthStore()
const route = useRoute()
const router = useRouter()
const majorName = computed(() => auth.user?.major || '应用统计')

const students = ref<DepartmentAlertItem[]>([])
const total = ref(0)
const loading = ref(true)

type GroupMode = 'grade' | 'level'
const groupMode = ref<GroupMode>('level')
const levelFilter = ref<'all' | 'red' | 'orange' | 'yellow'>('all')
const gradeFilter = ref('全部年级')
const currentPage = ref(1)
const pageSize = 10
const termOptions = ref<string[]>([])
const selectedTerm = ref('')

const grades = ref<string[]>(['全部年级'])

const levelConfig = {
  red:    { label: '红色预警', color: '#e04538', bg: '#fef2f2', border: '#fecaca', desc: '挂科两门及以上或缺勤四次及以上' },
  orange: { label: '橙色预警', color: '#f09a4e', bg: '#fff7ed', border: '#fed7aa', desc: '挂科一门或缺勤三次' },
  yellow: { label: '黄色预警', color: '#ca8a04', bg: '#fefce8', border: '#fde68a', desc: '缺勤一到两次且无挂科' },
}

const trendIcon = { up: '↑', down: '↓', flat: '→' }
const trendColor = { up: '#3db87e', down: '#e04538', flat: '#9facc5' }

const counts = computed(() => ({
  red:    students.value.filter(s => s.level === 'red').length,
  orange: students.value.filter(s => s.level === 'orange').length,
  yellow: students.value.filter(s => s.level === 'yellow').length,
}))

const filtered = computed(() => {
  let list = students.value
  if (levelFilter.value !== 'all') list = list.filter(s => s.level === levelFilter.value)
  if (gradeFilter.value !== '全部年级') list = list.filter(s => s.grade === gradeFilter.value)
  return list
})

const pagedFiltered = computed(() => {
  const start = (currentPage.value - 1) * pageSize
  return filtered.value.slice(start, start + pageSize)
})

const byGrade = computed(() => {
  const map: Record<string, DepartmentAlertItem[]> = {}
  for (const g of grades.value.filter(g => g !== '全部年级')) {
    const items = pagedFiltered.value.filter(s => s.grade === g)
    if (items.length) map[g] = items
  }
  return map
})

const byLevel = computed(() => {
  const map: Record<string, DepartmentAlertItem[]> = {}
  for (const lv of ['red', 'orange', 'yellow'] as const) {
    const items = pagedFiltered.value.filter(s => s.level === lv)
    if (items.length) map[lv] = items
  }
  return map
})

async function load() {
  if (!selectedTerm.value) return
  loading.value = true
  try {
    // This page is a full professional overview, so load the complete active
    // result set before calculating level and grade counts.
    const page = await getDepartmentAlerts({ term: selectedTerm.value, page: 1, size: 1000 })
    students.value = page.list
    total.value = page.total
    const gradeSet = new Set(page.list.map(s => s.grade).filter(Boolean))
    grades.value = ['全部年级', ...Array.from(gradeSet).sort()]
  } finally {
    loading.value = false
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

watch([levelFilter, gradeFilter], () => {
  currentPage.value = 1
})

watch(selectedTerm, async term => {
  if (!term) return
  currentPage.value = 1
  await router.replace({ query: { ...route.query, term } })
  await load()
})

onMounted(initialize)
</script>

<template>
  <div class="page" v-loading="loading">
    <div class="page-head">
      <div>
        <h2 class="page-title">预警管理</h2>
        <p class="page-sub">{{ majorName }}专业 · {{ selectedTerm }} · 共 {{ total }} 名预警学生</p>
      </div>
      <div class="head-actions">
        <select v-model="selectedTerm" class="filter-select">
          <option v-for="t in termOptions" :key="t" :value="t">{{ t }}</option>
        </select>
        <select v-model="gradeFilter" class="filter-select">
          <option v-for="g in grades" :key="g">{{ g }}</option>
        </select>
        <div class="group-toggle">
          <button :class="['toggle-btn', { active: groupMode === 'level' }]" @click="groupMode = 'level'">按预警等级</button>
          <button :class="['toggle-btn', { active: groupMode === 'grade' }]" @click="groupMode = 'grade'">按年级</button>
        </div>
      </div>
    </div>

    <div class="summary-row">
      <div v-for="(cfg, key) in levelConfig" :key="key"
        class="summary-card"
        :class="{ active: levelFilter === key }"
        :style="levelFilter === key ? { borderColor: cfg.color, background: cfg.bg } : {}"
        @click="levelFilter = levelFilter === key ? 'all' : key">
        <div class="sc-top">
          <span class="sc-dot" :style="{ background: cfg.color }" />
          <span class="sc-label" :style="levelFilter === key ? { color: cfg.color } : {}">{{ cfg.label }}</span>
        </div>
        <div class="sc-num" :style="{ color: cfg.color }">{{ counts[key as keyof typeof counts] }}</div>
        <div class="sc-desc">{{ cfg.desc }}</div>
      </div>
      <div class="summary-card total" :class="{ active: levelFilter === 'all' }" @click="levelFilter = 'all'">
        <div class="sc-top"><span class="sc-dot" style="background:#374151" /><span class="sc-label">全部预警</span></div>
        <div class="sc-num" style="color:#374151">{{ total }}</div>
        <div class="sc-desc">点击查看所有预警学生</div>
      </div>
    </div>

    <template v-if="groupMode === 'level'">
      <div v-for="(list, lv) in byLevel" :key="lv" class="group-section">
        <div class="group-header" :style="{ borderLeftColor: levelConfig[lv as keyof typeof levelConfig].color, background: levelConfig[lv as keyof typeof levelConfig].bg }">
          <span class="group-title" :style="{ color: levelConfig[lv as keyof typeof levelConfig].color }">{{ levelConfig[lv as keyof typeof levelConfig].label }}</span>
          <span class="group-count">{{ list.length }} 人</span>
        </div>
        <div class="alert-list">
          <div v-for="s in list" :key="s.id" class="alert-row">
            <div class="alert-level-bar" :style="{ background: levelConfig[s.level].color }" />
            <div class="alert-avatar" :style="{ background: levelConfig[s.level].color + '22', color: levelConfig[s.level].color }">{{ s.studentName[0] }}</div>
            <div class="alert-main">
              <div class="alert-top-row">
                <span class="alert-name">{{ s.studentName }}</span>
                <span class="alert-id">{{ s.studentId }}</span>
                <span class="alert-grade-tag">{{ s.grade }}</span>
                <span class="alert-class">{{ s.className }}</span>
              </div>
              <div class="alert-stats">
                <span class="stat-item">GPA <b :style="{ color: s.gpa < 2.0 ? '#e04538' : s.gpa < 2.3 ? '#f09a4e' : '#374151' }">{{ s.gpa }}</b></span>
                <span class="stat-sep" />
                <span class="stat-item">趋势 <b :style="{ color: trendColor[s.trend] }">{{ trendIcon[s.trend] }}</b></span>
                <span class="stat-sep" />
                <span class="stat-item">缺勤 <b :style="{ color: s.absences > 4 ? '#e04538' : '#374151' }">{{ s.absences }}次</b></span>
              </div>
              <div class="alert-reason">{{ s.reason }}</div>
            </div>
          </div>
        </div>
      </div>
    </template>

    <template v-else>
      <div v-for="(list, grade) in byGrade" :key="grade" class="group-section">
        <div class="group-header grade-header">
          <span class="group-title grade-title">{{ grade }}</span>
          <span class="group-count">{{ list.length }} 人</span>
          <div class="grade-level-pills">
            <span v-if="list.filter(s=>s.level==='red').length" class="pill red">红 {{ list.filter(s=>s.level==='red').length }}</span>
            <span v-if="list.filter(s=>s.level==='orange').length" class="pill orange">橙 {{ list.filter(s=>s.level==='orange').length }}</span>
            <span v-if="list.filter(s=>s.level==='yellow').length" class="pill yellow">黄 {{ list.filter(s=>s.level==='yellow').length }}</span>
          </div>
        </div>
        <div class="alert-list">
          <div v-for="s in list" :key="s.id" class="alert-row">
            <div class="alert-level-bar" :style="{ background: levelConfig[s.level].color }" />
            <div class="alert-avatar" :style="{ background: levelConfig[s.level].color + '22', color: levelConfig[s.level].color }">{{ s.studentName[0] }}</div>
            <div class="alert-main">
              <div class="alert-top-row">
                <span class="alert-name">{{ s.studentName }}</span>
                <span class="alert-id">{{ s.studentId }}</span>
                <span class="level-badge" :style="{ color: levelConfig[s.level].color, background: levelConfig[s.level].bg, borderColor: levelConfig[s.level].border }">{{ levelConfig[s.level].label }}</span>
                <span class="alert-class">{{ s.className }}</span>
              </div>
              <div class="alert-stats">
                <span class="stat-item">GPA <b :style="{ color: s.gpa < 2.0 ? '#e04538' : s.gpa < 2.3 ? '#f09a4e' : '#374151' }">{{ s.gpa }}</b></span>
                <span class="stat-sep" />
                <span class="stat-item">趋势 <b :style="{ color: trendColor[s.trend] }">{{ trendIcon[s.trend] }}</b></span>
                <span class="stat-sep" />
                <span class="stat-item">缺勤 <b :style="{ color: s.absences > 4 ? '#e04538' : '#374151' }">{{ s.absences }}次</b></span>
              </div>
              <div class="alert-reason">{{ s.reason }}</div>
            </div>
          </div>
        </div>
      </div>
    </template>

    <div v-if="filtered.length === 0" class="empty">
      <div>当前筛选条件下无预警学生</div>
    </div>

    <div v-if="filtered.length > pageSize" class="pagination-row">
      <el-pagination
        v-model:current-page="currentPage"
        :total="filtered.length"
        :page-size="pageSize"
        :pager-count="5"
        layout="prev, pager, next"
        prev-text="上一页"
        next-text="下一页"
      />
    </div>
  </div>
</template>

<style scoped>
.page { padding: 28px 32px; display: flex; flex-direction: column; gap: 20px; }
.page-head { display: flex; align-items: flex-start; justify-content: space-between; flex-wrap: wrap; gap: 12px; }
.page-title { font-size: 20px; font-weight: 800; color: #101d3e; margin: 0; letter-spacing: -0.01em; }
.page-sub { font-size: 13px; color: #9facc5; margin: 4px 0 0; }
.head-actions { display: flex; align-items: center; gap: 10px; }
.filter-select { padding: 6px 28px 6px 12px; border: 1.5px solid #e5e9f0; border-radius: 9px; font-size: 13px; font-weight: 600; color: #374151; background: #fff url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 24 24'%3E%3Cpath fill='%239facc5' d='M7 10l5 5 5-5z'/%3E%3C/svg%3E") no-repeat right 8px center; appearance: none; cursor: pointer; }
.filter-select:focus { outline: none; border-color: #2A4D99; }
.group-toggle { display: flex; gap: 3px; background: #f0f2f7; border-radius: 10px; padding: 3px; }
.toggle-btn { padding: 6px 14px; border: none; border-radius: 8px; font-size: 12.5px; font-weight: 600; color: #6b7280; background: transparent; cursor: pointer; transition: all 0.15s; }
.toggle-btn.active { background: #fff; color: #2A4D99; box-shadow: 0 1px 4px rgba(0,0,0,0.1); }
.summary-row { display: grid; grid-template-columns: repeat(4, 1fr); gap: 12px; }
.summary-card { background: #fff; border-radius: 14px; padding: 16px 18px; border: 2px solid #edf0f6; cursor: pointer; transition: all 0.15s; display: flex; flex-direction: column; gap: 4px; }
.summary-card:hover { border-color: #c8d4e8; }
.summary-card.active { box-shadow: 0 4px 16px rgba(0,0,0,0.08); }
.summary-card.total.active { border-color: #2A4D99; background: rgba(42,77,153,0.04); }
.sc-top { display: flex; align-items: center; gap: 7px; }
.sc-dot { width: 8px; height: 8px; border-radius: 50%; flex-shrink: 0; }
.sc-label { font-size: 13px; font-weight: 700; color: #374151; }
.sc-num { font-size: 32px; font-weight: 800; line-height: 1.1; font-variant-numeric: tabular-nums; }
.sc-desc { font-size: 11.5px; color: #9facc5; line-height: 1.4; }
.group-section { display: flex; flex-direction: column; border-radius: 16px; overflow: hidden; border: 1px solid rgba(10,19,41,0.07); box-shadow: 0 1px 2px rgba(10,19,41,0.03), 0 4px 16px rgba(10,19,41,0.04); }
.group-header { padding: 12px 20px; display: flex; align-items: center; gap: 10px; border-left: 4px solid transparent; }
.group-title { font-size: 14px; font-weight: 800; }
.group-count { font-size: 12px; font-weight: 600; padding: 2px 8px; border-radius: 20px; background: rgba(0,0,0,0.06); color: #374151; }
.grade-header { background: #f8f9fc; border-left-color: #2A4D99; }
.grade-title { color: #111827; }
.grade-level-pills { display: flex; gap: 5px; margin-left: 4px; }
.pill { font-size: 11px; font-weight: 700; padding: 2px 7px; border-radius: 99px; }
.pill.red    { background: rgba(224,69,56,0.1); color: #e04538; }
.pill.orange { background: rgba(240,154,78,0.1); color: #f09a4e; }
.pill.yellow { background: rgba(202,138,4,0.1);  color: #ca8a04; }
.alert-list { background: #fff; display: flex; flex-direction: column; }
.alert-row { display: flex; align-items: flex-start; padding: 14px 20px; border-top: 1px solid #f0f2f7; transition: background 0.12s; }
.alert-row:hover { background: #f8f9fc; }
.alert-level-bar { width: 3px; min-height: 48px; border-radius: 2px; flex-shrink: 0; margin-right: 12px; align-self: stretch; }
.alert-avatar { width: 36px; height: 36px; border-radius: 50%; flex-shrink: 0; display: flex; align-items: center; justify-content: center; font-size: 14px; font-weight: 700; margin-right: 14px; }
.alert-main { flex: 1; min-width: 0; }
.alert-top-row { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; margin-bottom: 5px; }
.alert-name { font-size: 14px; font-weight: 700; color: #111827; }
.alert-id { font-size: 12px; color: #9facc5; }
.alert-grade-tag { font-size: 12px; color: #6b7280; background: #f0f2f7; padding: 2px 6px; border-radius: 5px; }
.alert-class { font-size: 12px; color: #6b7280; background: #f0f2f7; padding: 2px 6px; border-radius: 5px; }
.level-badge { font-size: 11.5px; font-weight: 700; padding: 2px 8px; border-radius: 6px; border: 1px solid; }
.alert-stats { display: flex; align-items: center; gap: 12px; margin: 5px 0; font-size: 12.5px; }
.stat-item { color: #6b7280; }
.stat-item b { font-weight: 700; }
.stat-sep { width: 1px; height: 12px; background: #e5e9f0; display: inline-block; }
.alert-reason { font-size: 12.5px; color: #6b7280; line-height: 1.5; }
.empty { text-align: center; padding: 40px; color: #9facc5; font-size: 14px; background: #fff; border-radius: 16px; border: 1px solid rgba(10,19,41,0.07); }
.pagination-row { display: flex; justify-content: flex-end; align-items: center; min-height: 40px; padding: 2px 0; }
.pagination-row :deep(.el-pagination) { min-height: 32px; }
.pagination-row :deep(.btn-prev),
.pagination-row :deep(.btn-next),
.pagination-row :deep(.el-pager li) { min-width: 32px; height: 32px; line-height: 32px; border-radius: 6px; }
.pagination-row :deep(.btn-prev), .pagination-row :deep(.btn-next) { padding: 0 8px; }
</style>
