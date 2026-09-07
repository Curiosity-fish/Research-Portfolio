<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getTeacherGpaProgress, getTeacherTerms } from '@/api'
import type { GpaProgressData } from '@/types'

const data = ref<GpaProgressData>({ terms: [], students: [] })
const loading = ref(false)
const route = useRoute()
const router = useRouter()
const className = ref('')
const availableTerms = ref<string[]>([])
const allTerms = computed(() => data.value.terms)
const students = computed(() => data.value.students)

async function loadProgress() {
  loading.value = true
  try {
    data.value = await getTeacherGpaProgress(selectedTerm.value)
  } finally {
    loading.value = false
  }
}

// 进退步计算（首学期 → 末学期）
const ranked = computed(() => {
  return students.value.map(s => {
    const first = s.gpaByTerm[allTerms.value[0]] ?? 0
    const last  = s.gpaByTerm[allTerms.value[allTerms.value.length - 1]] ?? 0
    const delta = +((last - first).toFixed(2))
    return { ...s, first, last, delta }
  }).sort((a, b) => b.delta - a.delta)
})

const topProgress  = computed(() => ranked.value.slice(0, 5))
const topRegress   = computed(() => [...ranked.value].reverse().slice(0, 5))

function deltaColor(d: number) {
  if (d > 0.15)  return '#3db87e'
  if (d < -0.15) return '#e04538'
  return '#9facc5'
}
function deltaBg(d: number) {
  if (d > 0.15)  return 'rgba(61,184,126,0.08)'
  if (d < -0.15) return 'rgba(224,69,56,0.08)'
  return 'rgba(159,172,197,0.08)'
}
function deltaSign(d: number) { return d > 0 ? '+' : '' }
function termX(index: number) {
  return terms.value.length <= 1 ? 282 : 44 + (index / (terms.value.length - 1)) * 476
}

// 当前选中学生（查看趋势用）
const selectedId = ref<string | null>(null)
const selectedStudent = computed(() => students.value.find(s => s.id === selectedId.value) ?? null)

const terms = allTerms

// 学期选择器
const termOptions = computed(() => availableTerms.value)
const selectedTerm = ref('')

function formatTerm(term: string) {
  const parts = term.split('-')
  if (parts.length !== 3) return term
  return `${parts[0]}-${parts[1]} 学年${parts[2] === '2' ? '第二' : '第一'}学期`
}

async function initialize() {
  loading.value = true
  try {
    const context = await getTeacherTerms()
    availableTerms.value = context.terms
    className.value = context.className
    const queryTerm = typeof route.query.term === 'string' ? route.query.term : ''
    selectedTerm.value = context.terms.includes(queryTerm) ? queryTerm : context.latestTerm
  } finally {
    loading.value = false
  }
}

onMounted(initialize)
watch(selectedTerm, async term => {
  if (!term) return
  selectedId.value = null
  await router.replace({ query: { ...route.query, term } })
  loadProgress()
})
</script>

<template>
  <div v-loading="loading" class="page">
    <div class="page-head">
      <div>
        <h2 class="page-title">GPA 进退情况</h2>
        <p class="page-sub">截至 {{ formatTerm(selectedTerm) }} · {{ className }} · 共 {{ students.length }} 名学生</p>
      </div>
      <div class="term-selector">
        <span class="term-selector-label">学期</span>
        <select v-if="termOptions.length > 1" v-model="selectedTerm" class="term-select">
          <option v-for="t in termOptions" :key="t" :value="t">{{ t }}</option>
        </select>
        <span v-else class="term-static">{{ selectedTerm }}</span>
      </div>
    </div>

    <!-- 进退步榜 -->
    <div class="rank-grid">
      <!-- 进步榜 -->
      <div class="card">
        <div class="card-title-row">
          <span class="card-title">进步最大 Top 5</span>
          <span class="badge-green">首学期至本学期</span>
        </div>
        <div class="rank-list">
          <div
            v-for="(s, i) in topProgress" :key="s.id"
            class="rank-item"
            :class="{ selected: selectedId === s.id }"
            @click="selectedId = selectedId === s.id ? null : s.id"
          >
            <span class="rank-no" :class="i < 3 ? 'top3' : ''">{{ i + 1 }}</span>
            <div class="rank-info">
              <span class="rank-name">{{ s.name }}</span>
              <span class="rank-id">{{ s.studentId }}</span>
            </div>
            <div class="rank-gpa-row">
              <span class="gpa-val">{{ s.first }}</span>
              <span class="gpa-arrow">→</span>
              <span class="gpa-val bold">{{ s.last }}</span>
            </div>
            <span
              class="delta-badge"
              :style="{ color: deltaColor(s.delta), background: deltaBg(s.delta) }"
            >{{ deltaSign(s.delta) }}{{ s.delta }}</span>
          </div>
        </div>
      </div>

      <!-- 退步榜 -->
      <div class="card">
        <div class="card-title-row">
          <span class="card-title">退步最大 Top 5</span>
          <span class="badge-red">需重点关注</span>
        </div>
        <div class="rank-list">
          <div
            v-for="(s, i) in topRegress" :key="s.id"
            class="rank-item"
            :class="{ selected: selectedId === s.id }"
            @click="selectedId = selectedId === s.id ? null : s.id"
          >
            <span class="rank-no warn">{{ i + 1 }}</span>
            <div class="rank-info">
              <span class="rank-name">{{ s.name }}</span>
              <span class="rank-id">{{ s.studentId }}</span>
            </div>
            <div class="rank-gpa-row">
              <span class="gpa-val">{{ s.first }}</span>
              <span class="gpa-arrow">→</span>
              <span class="gpa-val bold">{{ s.last }}</span>
            </div>
            <span
              class="delta-badge"
              :style="{ color: deltaColor(s.delta), background: deltaBg(s.delta) }"
            >{{ deltaSign(s.delta) }}{{ s.delta }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 全班 GPA 趋势表 -->
    <div class="card table-card">
      <div class="card-title-row" style="margin-bottom: 16px">
        <span class="card-title">全班学期 GPA 明细</span>
        <span class="tip-text">点击学生行可查看趋势</span>
      </div>
      <div class="table-wrap">
        <table class="gpa-table">
          <thead>
            <tr>
              <th class="col-name">姓名</th>
              <th class="col-id">学号</th>
              <th v-for="t in terms" :key="t">{{ t }}</th>
              <th class="col-delta">变化幅度</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="s in ranked" :key="s.id"
              :class="{ 'row-selected': selectedId === s.id }"
              @click="selectedId = selectedId === s.id ? null : s.id"
            >
              <td class="col-name bold-name">{{ s.name }}</td>
              <td class="col-id gray">{{ s.studentId }}</td>
              <td
                v-for="t in terms" :key="t"
                class="gpa-cell"
                :class="{
                  'cell-high': (s.gpaByTerm[t] ?? 0) >= 3.5,
                  'cell-low':  (s.gpaByTerm[t] ?? 0) > 0 && (s.gpaByTerm[t] ?? 0) < 2.5,
                }"
              >
                {{ s.gpaByTerm[t] ?? '-' }}
              </td>
              <td>
                <span
                  class="delta-badge"
                  :style="{ color: deltaColor(s.delta), background: deltaBg(s.delta) }"
                >{{ deltaSign(s.delta) }}{{ s.delta }}</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <div class="legend-row">
        <span class="legend-item"><span class="legend-dot green" />GPA ≥ 3.5</span>
        <span class="legend-item"><span class="legend-dot red" />GPA &lt; 2.5</span>
      </div>
    </div>

    <!-- 学生趋势折线（选中后展示） -->
    <div v-if="selectedStudent" class="card trend-card">
      <!-- 卡片标题行 -->
      <div class="card-title-row">
        <div class="trend-header-left">
          <div class="trend-avatar">{{ selectedStudent.name[0] }}</div>
          <div>
            <span class="card-title">{{ selectedStudent.name }} · GPA 走势</span>
            <div class="trend-sub">{{ selectedStudent.studentId }} · 点击其他学生可切换查看</div>
          </div>
        </div>
        <button class="close-btn" @click="selectedId = null">✕ 关闭</button>
      </div>

      <!-- 左右布局 -->
      <div class="trend-body-row">

        <!-- 左：折线图 -->
        <div class="trend-chart-col">
          <svg class="line-chart" viewBox="0 0 540 190" preserveAspectRatio="xMidYMid meet">
            <defs>
              <linearGradient id="areaGrad" x1="0" y1="0" x2="0" y2="1">
                <stop offset="0%" stop-color="#2A4D99" stop-opacity="0.14" />
                <stop offset="100%" stop-color="#2A4D99" stop-opacity="0.01" />
              </linearGradient>
            </defs>
            <!-- 网格线 + Y轴刻度 -->
            <g v-for="(v, i) in [4.0, 3.5, 3.0, 2.5, 2.0, 1.5]" :key="v">
              <line :x1="44" :y1="16 + i * 26" :x2="528" :y2="16 + i * 26"
                stroke="#f0f2f7" stroke-width="1" />
              <text :x="38" :y="20 + i * 26" text-anchor="end" font-size="10" fill="#c8d4e8">{{ v.toFixed(1) }}</text>
            </g>
            <!-- 2.5 预警线 -->
            <line x1="44" y1="90" x2="520" y2="90" stroke="#fde68a" stroke-width="1.5" stroke-dasharray="5,4" />
            <text x="523" y="94" font-size="9" fill="#f09a4e" font-weight="600">预警</text>

            <!-- 面积填充 -->
            <path
              :d="(() => {
                const pts = terms.map((t, i) => {
                  const gpa = selectedStudent!.gpaByTerm[t] ?? 2.0
                  const x = termX(i)
                  const y = 16 + (4.0 - gpa) / 2.5 * 130
                  return [x, y]
                })
                const line = pts.map(([x,y], i) => `${i===0?'M':'L'}${x},${y}`).join(' ')
                return line + ` L${pts[pts.length-1][0]},146 L${pts[0][0]},146 Z`
              })()"
              fill="url(#areaGrad)"
            />
            <!-- 折线 -->
            <polyline
              :points="terms.map((t, i) => {
                const gpa = selectedStudent!.gpaByTerm[t] ?? 2.0
                const x = termX(i)
                const y = 16 + (4.0 - gpa) / 2.5 * 130
                return `${x},${y}`
              }).join(' ')"
              fill="none" stroke="#2A4D99" stroke-width="2.5"
              stroke-linejoin="round" stroke-linecap="round"
            />
            <!-- 数据点 -->
            <g v-for="(t, i) in terms" :key="t">
              <circle
                :cx="termX(i)"
                :cy="16 + (4.0 - (selectedStudent!.gpaByTerm[t] ?? 2.0)) / 2.5 * 130"
                r="3" fill="#fff" stroke="#2A4D99" stroke-width="2"
              />
            </g>
            <!-- X轴标签 -->
            <text v-for="(t, i) in terms" :key="`lbl-${t}`"
              :x="termX(i)"
              y="178" text-anchor="middle" font-size="10" fill="#b4bed2"
            >{{ t }}</text>
          </svg>

          <!-- 总变化量 -->
          <div class="trend-summary">
            <span class="ts-label">首学期</span>
            <span class="ts-val">{{ selectedStudent.gpaByTerm[terms[0]] }}</span>
            <span class="ts-arrow">→</span>
            <span class="ts-label">最新</span>
            <span class="ts-val">{{ selectedStudent.gpaByTerm[terms[terms.length - 1]] }}</span>
            <span
              class="ts-delta"
              :style="{
                color: deltaColor(+(selectedStudent.gpaByTerm[terms[terms.length-1]] - selectedStudent.gpaByTerm[terms[0]]).toFixed(2)),
                background: deltaBg(+(selectedStudent.gpaByTerm[terms[terms.length-1]] - selectedStudent.gpaByTerm[terms[0]]).toFixed(2))
              }"
            >{{ deltaSign(selectedStudent.gpaByTerm[terms[terms.length-1]] - selectedStudent.gpaByTerm[terms[0]]) }}{{
              +((selectedStudent.gpaByTerm[terms[terms.length-1]] - selectedStudent.gpaByTerm[terms[0]]).toFixed(2))
            }}</span>
          </div>
        </div>

        <!-- 右：学期数据列 -->
        <div class="trend-data-col">
          <div
            v-for="t in terms" :key="t"
            class="term-data-item"
            :class="{ 'is-warn': (selectedStudent.gpaByTerm[t] ?? 0) > 0 && (selectedStudent.gpaByTerm[t] ?? 0) < 2.5 }"
          >
            <span class="tdi-term">{{ t }}</span>
            <span class="tdi-gpa">{{ selectedStudent.gpaByTerm[t] ?? '-' }}</span>
            <span v-if="(selectedStudent.gpaByTerm[t] ?? 0) > 0 && (selectedStudent.gpaByTerm[t] ?? 0) < 2.5" class="tdi-warn-tag">预警</span>
          </div>
        </div>

      </div>
    </div>
  </div>
</template>

<style scoped>
.page { padding: 28px 32px; display: flex; flex-direction: column; gap: 20px; }
.page-head { display: flex; align-items: flex-start; justify-content: space-between; }
.page-title { font-size: 20px; font-weight: 800; color: #101d3e; margin: 0; letter-spacing: -0.01em; }
.page-sub { font-size: 13px; color: #9facc5; margin: 4px 0 0; }
.term-static { font-size: 13px; font-weight: 600; color: #374151; }

.view-all { font-size: 12px; color: #2A4D99; text-decoration: none; }

.view-toggle { display: none; }
.toggle-btn { display: none; }

.card {
  background: #fff; border-radius: 16px; padding: 20px 22px;
  border: 1px solid rgba(10,19,41,0.07);
  box-shadow: 0 1px 2px rgba(10,19,41,0.03), 0 4px 16px rgba(10,19,41,0.05);
}
.card-title { font-size: 14px; font-weight: 700; color: #111827; }
.card-title-row { display: flex; align-items: center; justify-content: space-between; }

.badge-green { font-size: 11px; font-weight: 600; color: #3db87e; background: rgba(61,184,126,0.1); padding: 2px 8px; border-radius: 6px; }
.badge-red   { font-size: 11px; font-weight: 600; color: #e04538; background: rgba(224,69,56,0.08); padding: 2px 8px; border-radius: 6px; }
.tip-text    { font-size: 12px; color: #9facc5; }

.rank-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; }

.rank-list { display: flex; flex-direction: column; gap: 6px; margin-top: 14px; }
.rank-item {
  display: flex; align-items: center; gap: 12px;
  padding: 10px 12px; border-radius: 10px; cursor: pointer;
  border: 1.5px solid transparent;
  transition: background 0.15s, border-color 0.15s;
  background: #f9fafb;
}
.rank-item:hover { background: #f0f4fb; }
.rank-item.selected { background: rgba(42,77,153,0.06); border-color: rgba(42,77,153,0.2); }

.rank-no {
  width: 22px; height: 22px; border-radius: 6px; flex-shrink: 0;
  display: flex; align-items: center; justify-content: center;
  font-size: 11px; font-weight: 800;
  background: #edf0f6; color: #6b7280;
}
.rank-no.top3 { background: rgba(42,77,153,0.12); color: #2A4D99; }
.rank-no.warn { background: rgba(224,69,56,0.1); color: #e04538; }

.rank-info { flex: 1; display: flex; flex-direction: column; gap: 1px; }
.rank-name { font-size: 13.5px; font-weight: 600; color: #111827; }
.rank-id   { font-size: 11px; color: #9facc5; }

.rank-gpa-row { display: flex; align-items: center; gap: 5px; font-size: 12.5px; color: #6b7280; }
.gpa-val { font-variant-numeric: tabular-nums; }
.gpa-val.bold { font-weight: 700; color: #111827; }
.gpa-arrow { color: #c8d4e8; }

.delta-badge {
  font-size: 12px; font-weight: 700;
  padding: 3px 8px; border-radius: 6px;
  white-space: nowrap; flex-shrink: 0;
}

.table-card { padding: 20px 22px; overflow-x: auto; }
.table-wrap { overflow-x: auto; }
.gpa-table {
  width: 100%; border-collapse: collapse; font-size: 13px;
  min-width: 700px;
}
.gpa-table thead th {
  padding: 8px 10px; text-align: center;
  font-size: 12px; font-weight: 700; color: #9facc5;
  border-bottom: 1.5px solid #f0f2f7;
  white-space: nowrap;
}
.gpa-table thead th.col-name { text-align: left; }
.gpa-table tbody tr {
  border-bottom: 1px solid #f8f9fb;
  cursor: pointer; transition: background 0.12s;
}
.gpa-table tbody tr:hover { background: #f6f8fc; }
.gpa-table tbody tr.row-selected { background: rgba(42,77,153,0.06); }
.gpa-table td { padding: 9px 10px; text-align: center; }
.col-name { text-align: left !important; min-width: 60px; }
.col-id   { min-width: 90px; }
.col-delta { min-width: 80px; }
.bold-name { font-weight: 600; color: #111827; }
.gray { color: #9facc5; font-size: 12px; }
.gpa-cell { font-variant-numeric: tabular-nums; font-weight: 500; }
.cell-high { color: #3db87e; font-weight: 700; }
.cell-low  { color: #e04538; font-weight: 700; }

.legend-row { display: flex; gap: 16px; margin-top: 14px; }
.legend-item { display: flex; align-items: center; gap: 5px; font-size: 12px; color: #6b7280; }
.legend-dot { width: 8px; height: 8px; border-radius: 50%; flex-shrink: 0; }
.legend-dot.green { background: #3db87e; }
.legend-dot.red   { background: #e04538; }

.trend-card {}

.trend-header-left { display: flex; align-items: center; gap: 12px; }
.trend-avatar {
  width: 36px; height: 36px; border-radius: 50%; flex-shrink: 0;
  background: linear-gradient(135deg, #2A4D99, #5580d4);
  display: flex; align-items: center; justify-content: center;
  color: #fff; font-size: 14px; font-weight: 700;
}
.trend-sub { font-size: 11.5px; color: #9facc5; margin-top: 2px; }

.close-btn {
  padding: 5px 14px; border: 1px solid #e5e9f0; border-radius: 8px;
  font-size: 12px; color: #6b7280; background: #fff; cursor: pointer;
  flex-shrink: 0; transition: background 0.15s;
}
.close-btn:hover { background: #f4f6fa; }

/* 左右布局 */
.trend-body-row {
  display: flex; gap: 24px; margin-top: 18px; align-items: flex-start;
}
.trend-chart-col {
  flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 12px;
}
.line-chart { width: 100%; height: auto; display: block; }

.trend-summary {
  display: flex; align-items: center; gap: 10px;
  padding: 10px 14px; background: #f8f9fc; border-radius: 10px;
  font-size: 13px;
}
.ts-label { color: #9facc5; font-size: 12px; }
.ts-val { font-size: 16px; font-weight: 800; color: #111827; font-variant-numeric: tabular-nums; }
.ts-arrow { color: #c8d4e8; font-size: 14px; }
.ts-delta {
  font-size: 13px; font-weight: 700; padding: 3px 10px;
  border-radius: 7px; margin-left: 4px;
}

/* 右侧学期数据列 */
.trend-data-col {
  width: 148px; flex-shrink: 0;
  display: flex; flex-direction: column; gap: 6px;
  padding-top: 4px;
}
.term-data-item {
  display: flex; align-items: center; justify-content: space-between;
  padding: 9px 12px; border-radius: 9px;
  background: #f8f9fc; border: 1.5px solid transparent;
  transition: border-color 0.15s;
}
.term-data-item.is-warn {
  background: rgba(224,69,56,0.04); border-color: rgba(224,69,56,0.18);
}
.tdi-term {
  font-size: 12px; font-weight: 600; color: #9facc5; flex-shrink: 0;
}
.tdi-gpa {
  font-size: 16px; font-weight: 800; color: #2A4D99;
  font-variant-numeric: tabular-nums;
}
.term-data-item.is-warn .tdi-gpa { color: #e04538; }
.tdi-warn-tag {
  font-size: 10px; font-weight: 700; color: #e04538;
  background: rgba(224,69,56,0.1); padding: 1px 5px; border-radius: 4px;
  flex-shrink: 0;
}

/* 学期选择器 */
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
