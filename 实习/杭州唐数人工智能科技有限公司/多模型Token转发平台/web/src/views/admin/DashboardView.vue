<script setup lang="ts">
/* 概览页：核心指标统计卡 + 用量图表（纯 SVG，零依赖）+ 近 7 日用量自适应表格。
   大屏增强：≥2560px 视口或手动开启展示模式，统计卡字号放大到远观可读档位 */
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import PageHeader from '@/components/common/PageHeader.vue'
import StatCardGrid from '@/components/adaptive/StatCardGrid.vue'
import AdaptiveTable from '@/components/adaptive/AdaptiveTable.vue'
import MiniLine from '@/components/charts/MiniLine.vue'
import MiniBars from '@/components/charts/MiniBars.vue'
import { getDashboardStats, getUsageStats } from '@/api/admin/ops'
import type { DashboardStats, PageData, UsageStat } from '@/api/types'
import { useUserMap } from '@/composables/useUserMap'
import { formatCount, formatMicro } from '@/utils/format'

const stats = ref<DashboardStats | null>(null)
const loading = ref(false)

// 大屏展示模式：自动（≥2560px）或手动开关
const viewportWidth = ref(window.innerWidth)
const manualDisplay = ref(false)
const displayMode = computed(() => manualDisplay.value || viewportWidth.value >= 2560)

let ro: ResizeObserver | null = null
const rootRef = ref<HTMLElement>()

onMounted(() => {
  void loadStats()
  void loadCharts()
  if (rootRef.value) {
    ro = new ResizeObserver(() => {
      viewportWidth.value = window.innerWidth
    })
    // 监听根元素即等于监听内容区宽度（侧边栏变化也覆盖）
    ro.observe(rootRef.value)
  }
})

onBeforeUnmount(() => ro?.disconnect())

async function loadStats(): Promise<void> {
  loading.value = true
  try {
    stats.value = await getDashboardStats()
  } finally {
    loading.value = false
  }
}

const cards = computed(() => [
  { label: '调用次数', value: formatCount(stats.value?.calls), unit: '次' },
  { label: '输入 Token', value: formatCount(stats.value?.prompt_tokens), unit: '' },
  { label: '输出 Token', value: formatCount(stats.value?.completion_tokens), unit: '' },
  { label: '总 Token', value: formatCount(stats.value?.total_tokens), unit: '' },
  { label: '累计充值', value: formatMicro(stats.value?.recharge_amount), unit: '元' },
  { label: '累计消费', value: formatMicro(stats.value?.consume_amount), unit: '元' },
])

/* ── 图表数据：usage 接口不分页，一次拉全量后前端聚合 ── */
const { userName } = useUserMap()
const dayList = ref<UsageStat[]>([])
const modelList = ref<UsageStat[]>([])
const userList = ref<UsageStat[]>([])
const chartLoading = ref(false)

async function loadCharts(): Promise<void> {
  chartLoading.value = true
  try {
    const [day, model, user] = await Promise.all([
      getUsageStats({ group_by: 'day' }),
      getUsageStats({ group_by: 'model' }),
      getUsageStats({ group_by: 'user' }),
    ])
    dayList.value = day.list
    modelList.value = model.list
    userList.value = user.list
  } finally {
    chartLoading.value = false
  }
}

/** 近 7 日调用趋势，按日期升序，标签取 MM-DD */
const trendPoints = computed(() => {
  const days = [...dayList.value].sort((a, b) => (a.day ?? '').localeCompare(b.day ?? '')).slice(-7)
  return days.map((d) => ({
    label: (d.day ?? '').slice(5),
    value: d.calls,
  }))
})

const modelTop = computed(() =>
  [...modelList.value]
    .sort((a, b) => b.total_tokens - a.total_tokens)
    .slice(0, 8)
    .map((m) => ({ label: m.model ?? '-', value: m.total_tokens })),
)

const userTop = computed(() =>
  [...userList.value]
    .sort((a, b) => b.total_tokens - a.total_tokens)
    .slice(0, 8)
    .map((u) => ({ label: userName(u.user_id ?? ''), value: u.total_tokens })),
)

/* ── 近 7 日用量表：UsageByDay 契约仅含 calls/total_tokens ── */
interface DayRow {
  day: string
  calls: number
  total_tokens: number
}

let usageCache: DayRow[] | null = null

async function fetchUsage(page: number, pageSize: number): Promise<PageData<DayRow>> {
  if (!usageCache) {
    const data = await getUsageStats({ group_by: 'day' })
    usageCache = data.list.map((u: UsageStat) => ({
      day: u.day ?? '-',
      calls: u.calls,
      total_tokens: u.total_tokens,
    }))
  }
  return {
    total: usageCache.length,
    page,
    page_size: pageSize,
    list: usageCache.slice((page - 1) * pageSize, page * pageSize),
  }
}
</script>

<template>
  <div ref="rootRef" class="page-root" :class="{ 'app-display-mode': displayMode }">
    <PageHeader title="概览">
      <template #actions>
        <el-switch
          v-model="manualDisplay"
          inline-prompt
          active-text="展示"
          inactive-text="展示"
          title="大屏展示模式"
          style="--el-switch-on-color: var(--brand)"
        />
        <el-button :icon="'Refresh'" :loading="loading" @click="loadStats">刷新</el-button>
      </template>
    </PageHeader>

    <StatCardGrid :min-card-width="displayMode ? 260 : 170">
      <div v-for="card in cards" :key="card.label" class="stat-card app-card" v-loading="loading">
        <div class="stat-card__label">{{ card.label }}</div>
        <div class="stat-card__value tnum">
          {{ card.value }}<span v-if="card.unit" class="stat-card__unit">{{ card.unit }}</span>
        </div>
      </div>
    </StatCardGrid>

    <!-- 图表区：纯 SVG 渲染，密度随宽度自适应换行 -->
    <div class="dashboard__charts" v-loading="chartLoading">
      <div class="app-card dashboard__chart">
        <div class="dashboard__chart-title">近 7 日调用趋势</div>
        <div class="dashboard__chart-body">
          <MiniLine :points="trendPoints" />
        </div>
      </div>
      <div class="app-card dashboard__chart">
        <div class="dashboard__chart-title">模型用量 Top 8（Token）</div>
        <div class="dashboard__chart-body">
          <MiniBars :items="modelTop" />
        </div>
      </div>
      <div class="app-card dashboard__chart">
        <div class="dashboard__chart-title">用户用量 Top 8（Token）</div>
        <div class="dashboard__chart-body">
          <MiniBars :items="userTop" color="var(--brand-hover)" />
        </div>
      </div>
    </div>

    <AdaptiveTable :fetch="fetchUsage" class="dashboard__table">
      <el-table-column prop="day" label="日期" min-width="120" />
      <el-table-column prop="calls" label="调用次数" min-width="110" align="right">
        <template #default="{ row }">{{ formatCount(row.calls) }}</template>
      </el-table-column>
      <el-table-column prop="total_tokens" label="总 Token" min-width="130" align="right">
        <template #default="{ row }">{{ formatCount(row.total_tokens) }}</template>
      </el-table-column>
    </AdaptiveTable>
  </div>
</template>

<style scoped>
.stat-card {
  padding: 14px 16px;
}
.stat-card__label {
  font-size: 12px;
  color: var(--text-muted);
  margin-bottom: 6px;
}
.stat-card__value {
  font-size: 22px;
  font-weight: 700;
  color: var(--text-primary);
  line-height: 28px;
}
.stat-card__unit {
  font-size: 12px;
  font-weight: 400;
  color: var(--text-muted);
  margin-left: 4px;
}

.dashboard__charts {
  flex-shrink: 0;
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 12px;
}
.dashboard__chart {
  height: 200px;
  padding: 12px 16px;
  display: flex;
  flex-direction: column;
}
.dashboard__chart-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 8px;
  flex-shrink: 0;
}
.dashboard__chart-body {
  flex: 1;
  min-height: 0;
}

/* 大屏展示模式：远观可读档位（配合 tokens.css 的 --display-scale 预留） */
.app-display-mode .stat-card {
  padding: 22px 24px;
}
.app-display-mode .stat-card__label {
  font-size: 15px;
  margin-bottom: 10px;
}
.app-display-mode .stat-card__value {
  font-size: 34px;
  line-height: 42px;
}
.app-display-mode .stat-card__unit {
  font-size: 16px;
}
.app-display-mode .dashboard__chart {
  height: 260px;
}
.dashboard__table {
  min-height: 0;
}
</style>
