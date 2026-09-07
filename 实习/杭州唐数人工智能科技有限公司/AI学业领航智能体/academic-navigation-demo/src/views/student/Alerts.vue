<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import type { AlertItem } from '@/types'
import WarningBadge from '@/components/common/WarningBadge.vue'
import { getStudentAlerts, getStudentTerms, getAlertDetail, confirmAlert } from '@/api'
import { useAlertStore } from '@/stores/alerts'

const alertStore = useAlertStore()
const route = useRoute()
const router = useRouter()
const alerts = ref<AlertItem[]>([])
const drawerVisible = ref(false)
const selectedAlert = ref<AlertItem | null>(null)
const confirmedIds = ref<Set<string>>(new Set())
const loading = ref(false)
const errorMessage = ref('')
const detailError = ref('')
const detailLoading = ref(false)
const total = ref(0)
const page = ref(1)
const size = 20
const termOptions = ref<string[]>([])
const selectedTerm = ref('')

async function loadAlerts() {
  loading.value = true
  errorMessage.value = ''
  try {
    const res = await getStudentAlerts(selectedTerm.value || undefined, page.value, size)
    alerts.value = res.list.map(a => ({ ...a, failedCourses: a.failedCourses ?? [] }))
    total.value = res.total
    alertStore.refreshUnread()
  } catch {
    errorMessage.value = '预警通知暂时无法加载，请检查网络后重试'
  } finally {
    loading.value = false
  }
}

async function openDrawer(alert: AlertItem) {
  selectedAlert.value = alert
  drawerVisible.value = true
  detailLoading.value = true
  detailError.value = ''
  try {
    const detail = await getAlertDetail(alert.id)
    selectedAlert.value = { ...detail, failedCourses: detail.failedCourses ?? [] }
  } catch {
    detailError.value = '详情加载失败，请关闭后重试'
  } finally {
    detailLoading.value = false
  }
}

async function confirmViewed(id: string) {
  await confirmAlert(id)
  confirmedIds.value.add(id)
  const target = alerts.value.find(a => a.id === id)
  if (target) target.status = 'processing'
  ElMessage.success('已确认查看')
  alertStore.refreshUnread()
}

function handlePageChange(next: number) {
  page.value = next
  loadAlerts()
}

async function initialize() {
  loading.value = true
  try {
    termOptions.value = await getStudentTerms()
    const queryTerm = typeof route.query.term === 'string' ? route.query.term : ''
    selectedTerm.value = termOptions.value.includes(queryTerm) ? queryTerm : (termOptions.value[0] ?? '')
    if (!selectedTerm.value) await loadAlerts()
  } finally {
    loading.value = false
  }
}

const yellowAlerts = computed(() => alerts.value.filter(a => a.level === 'yellow'))
const orangeAlerts = computed(() => alerts.value.filter(a => a.level === 'orange'))
const redAlerts = computed(() => alerts.value.filter(a => a.level === 'red'))

onMounted(initialize)

watch(selectedTerm, async term => {
  if (!term) return
  page.value = 1
  await router.replace({ query: { ...route.query, term } })
  loadAlerts()
})
</script>

<template>
  <div v-loading="loading" class="page">
    <div v-if="errorMessage" class="load-error" role="alert">
      <el-icon><WarningFilled /></el-icon>
      <span>{{ errorMessage }}</span>
      <el-button text type="primary" size="small" @click="loadAlerts">重新加载</el-button>
    </div>
    <div class="page-head">
      <div>
        <h2 class="page-title">预警通知</h2>
        <p class="page-sub">{{ selectedTerm }} · 系统自动检测并推送的学业预警信息</p>
      </div>
      <el-select v-if="termOptions.length > 1" v-model="selectedTerm" placeholder="选择学期" style="width: 180px">
        <el-option v-for="term in termOptions" :key="term" :label="term" :value="term" />
      </el-select>
      <span v-else-if="selectedTerm" class="term-static">{{ selectedTerm }}</span>
    </div>

    <div v-if="alerts.length === 0" class="status-ok">
      <div class="status-ok-icon">
        <el-icon class="text-white text-xl"><CircleCheckFilled /></el-icon>
      </div>
      <div>
        <div class="status-ok-title">学业状态正常</div>
        <div class="status-ok-sub">本学期暂无预警记录，继续保持良好学习状态</div>
      </div>
    </div>

    <!-- Red alerts -->
    <template v-if="redAlerts.length">
      <div class="section-header">
        <WarningBadge level="red" />
        <span class="section-count">{{ redAlerts.length }} 条严重预警</span>
      </div>
      <div v-for="alert in redAlerts" :key="alert.id" class="alert-card red">
        <div class="alert-banner red-banner">
          <el-icon><Warning /></el-icon>
          严重学业危机 — 需配合院级干预
        </div>
        <div class="alert-card-body">
          <div class="alert-title-row">
            <span class="alert-title">{{ alert.title }}</span>
            <WarningBadge :level="alert.level" />
          </div>
          <p class="alert-desc">{{ alert.description }}</p>
          <div v-if="alert.failedCourses?.length" class="failed-courses-block">
            <span class="failed-courses-label">挂科课程</span>
            <div class="failed-courses-tags">
              <span v-for="c in alert.failedCourses" :key="c" class="failed-tag red-tag">{{ c }}</span>
            </div>
          </div>
          <div class="alert-meta-row">
            <span v-if="alert.triggerEvent" class="alert-trigger">
              <el-icon><Bell /></el-icon>{{ alert.triggerEvent }}
            </span>
            <span class="alert-date">{{ alert.pushedAt || alert.date }}</span>
          </div>
          <div class="alert-contact">
            <el-icon><Phone /></el-icon>
            院系联系人：李老师（辅导员）— 15888001234
          </div>
        </div>
        <div class="alert-card-footer">
          <el-button size="small" @click="openDrawer(alert)">查看详情</el-button>
        </div>
      </div>
    </template>

    <!-- Orange alerts -->
    <template v-if="orangeAlerts.length">
      <div class="section-header">
        <WarningBadge level="orange" />
        <span class="section-count">{{ orangeAlerts.length }} 条警示预警</span>
      </div>
      <div v-for="alert in orangeAlerts" :key="alert.id" class="alert-card orange">
        <div class="alert-card-body">
          <div class="alert-title-row">
            <span class="alert-title">{{ alert.title }}</span>
            <WarningBadge :level="alert.level" />
          </div>
          <p class="alert-desc">{{ alert.description }}</p>
          <div v-if="alert.failedCourses?.length" class="failed-courses-block">
            <span class="failed-courses-label">挂科课程</span>
            <div class="failed-courses-tags">
              <span v-for="c in alert.failedCourses" :key="c" class="failed-tag orange-tag">{{ c }}</span>
            </div>
          </div>
          <div class="alert-meta-row">
            <span v-if="alert.triggerEvent" class="orange-trigger">
              <el-icon><Bell /></el-icon>{{ alert.triggerEvent }}
            </span>
            <span class="alert-date">{{ alert.pushedAt || alert.date }}</span>
          </div>
          <div class="alert-teacher-notice">
            <el-icon><UserFilled /></el-icon>
            班主任王教授已收到通知，将于近期联系你
          </div>
        </div>
        <div class="alert-card-footer">
          <el-button size="small" @click="openDrawer(alert)">查看详情</el-button>
          <el-button
            size="small" type="primary"
            :disabled="confirmedIds.has(alert.id)"
            @click="confirmViewed(alert.id)"
          >
            {{ confirmedIds.has(alert.id) ? '已确认' : '确认已查看' }}
          </el-button>
        </div>
      </div>
    </template>

    <!-- Yellow alerts table -->
    <template v-if="yellowAlerts.length">
      <div class="section-header">
        <WarningBadge level="yellow" />
        <span class="section-count">{{ yellowAlerts.length }} 条关注预警</span>
      </div>
      <div class="alert-table-wrap">
        <el-table :data="yellowAlerts" stripe>
          <el-table-column label="预警类型" prop="type" width="120" />
          <el-table-column label="预警内容" prop="title" />
          <el-table-column label="触发事件" prop="triggerEvent" width="180" />
          <el-table-column label="推送时间" width="160">
            <template #default="{ row }">{{ row.pushedAt || row.date }}</template>
          </el-table-column>
          <el-table-column label="操作" width="90">
            <template #default="{ row }">
              <el-button text type="primary" size="small" @click="openDrawer(row)">详情</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </template>

    <div v-if="total > size" class="pagination-row">
      <el-pagination
        v-model:current-page="page"
        :total="total"
        :page-size="size"
        layout="prev, pager, next"
        @current-change="handlePageChange"
      />
    </div>

    <!-- Detail drawer -->
    <el-drawer v-model="drawerVisible" title="预警详情" size="400px">
      <div v-loading="detailLoading" v-if="selectedAlert" class="flex flex-col gap-4">
        <div v-if="detailError" class="detail-error" role="alert">
          <el-icon><WarningFilled /></el-icon>{{ detailError }}
        </div>
        <div class="drawer-header">
          <span class="text-base font-semibold" style="color: #101d3e">{{ selectedAlert.title }}</span>
          <WarningBadge :level="selectedAlert.level" />
        </div>
        <el-descriptions :column="1" border>
          <el-descriptions-item label="预警类型">{{ selectedAlert.type }}</el-descriptions-item>
          <el-descriptions-item label="触发事件">{{ selectedAlert.triggerEvent || '—' }}</el-descriptions-item>
          <el-descriptions-item label="推送时间">{{ selectedAlert.pushedAt || selectedAlert.date }}</el-descriptions-item>
          <el-descriptions-item label="详细描述">{{ selectedAlert.description }}</el-descriptions-item>
          <el-descriptions-item v-if="selectedAlert.failedCourses?.length" label="挂科课程">
            <div class="failed-courses-tags" style="margin-top:2px">
              <span v-for="c in selectedAlert.failedCourses" :key="c" class="failed-tag orange-tag">{{ c }}</span>
            </div>
          </el-descriptions-item>
        </el-descriptions>
        <div v-if="selectedAlert.suggestion" class="suggestion-box">
          <div class="suggestion-title">改进建议</div>
          <p class="suggestion-body">{{ selectedAlert.suggestion }}</p>
        </div>
      </div>
    </el-drawer>
  </div>
</template>

<style scoped>
.page { padding: 28px 32px; display: flex; flex-direction: column; gap: 20px; }
.load-error, .detail-error {
  display: flex; align-items: center; gap: 10px; padding: 11px 14px;
  border: 1px solid #f6c8c3; border-radius: 10px; background: #fff7f6;
  color: #a63026; font-size: 13px;
}
.load-error .el-button { margin-left: auto; }
.detail-error { margin-bottom: 2px; }
.page-head { display: flex; align-items: flex-start; justify-content: space-between; gap: 16px; }
.page-title { font-size: 20px; font-weight: 800; color: #101d3e; margin: 0; letter-spacing: -0.01em; }
.page-sub { font-size: 13px; color: #9facc5; margin: 4px 0 0; }
.term-static { font-size: 13px; color: #374151; font-weight: 600; }

.status-ok {
  display: flex; align-items: center; gap: 16px;
  background: #f0fdf4; border: 1px solid #bbf7d0;
  border-radius: 14px; padding: 18px 20px;
}
.status-ok-icon {
  width: 44px; height: 44px; flex-shrink: 0;
  border-radius: 12px; background: #3db87e;
  display: flex; align-items: center; justify-content: center;
}
.status-ok-title { font-weight: 600; color: #166534; font-size: 14px; }
.status-ok-sub { font-size: 13px; color: #16a34a; margin-top: 2px; }

.section-header { display: flex; align-items: center; gap: 10px; }
.section-count { font-size: 13px; color: #6b7280; }

.alert-card { border-radius: 14px; overflow: hidden; border: 1px solid; }
.alert-card.red { border-color: #fecaca; background: #fff5f5; }
.alert-card.orange { border-color: #fed7aa; background: #fffbf5; }

.alert-banner {
  display: flex; align-items: center; gap: 8px;
  padding: 10px 16px; font-size: 13px; font-weight: 600;
}
.red-banner { background: #fee2e2; color: #991b1b; }

.alert-card-body { padding: 16px; }
.alert-title-row { display: flex; align-items: center; justify-content: space-between; margin-bottom: 8px; }
.alert-title { font-size: 15px; font-weight: 700; color: #1f2937; }
.alert-desc { font-size: 13px; color: #6b7280; margin: 0 0 10px; line-height: 1.6; }

.alert-meta-row {
  display: flex; align-items: center; justify-content: space-between; margin-bottom: 10px;
}
.alert-trigger, .orange-trigger {
  display: flex; align-items: center; gap: 4px;
  font-size: 12px; padding: 2px 8px; border-radius: 6px;
}
.alert-trigger { color: #991b1b; background: #fee2e2; }
.orange-trigger { color: #92400e; background: #fef3c7; }
.alert-date { font-size: 12px; color: #9ca3af; }

.alert-contact, .alert-teacher-notice {
  display: flex; align-items: center; gap: 6px;
  font-size: 12.5px; padding: 8px 12px; border-radius: 8px;
}
.alert-contact { background: #fef2f2; color: #b91c1c; }
.alert-teacher-notice { background: #fff7ed; color: #c2410c; }

.alert-card-footer {
  padding: 10px 16px; border-top: 1px solid rgba(0,0,0,0.06);
  display: flex; gap: 8px; justify-content: flex-end;
}
.alert-table-wrap { background: #fff; border-radius: 14px; border: 1px solid #edf0f6; overflow: hidden; }

.drawer-header {
  display: flex; align-items: center; justify-content: space-between;
  padding-bottom: 12px; border-bottom: 1px solid #f0f0f0; margin-bottom: 4px;
}
.suggestion-box { background: #eff6ff; border-radius: 10px; padding: 14px; }
.suggestion-title { font-size: 13px; font-weight: 600; color: #1e40af; margin-bottom: 6px; }
.suggestion-body { font-size: 13px; color: #2563eb; line-height: 1.6; margin: 0; }

.failed-courses-block {
  display: flex; align-items: flex-start; gap: 8px;
  margin-bottom: 10px;
}
.failed-courses-label {
  font-size: 11.5px; font-weight: 600; color: #6b7280;
  white-space: nowrap; padding-top: 3px;
}
.failed-courses-tags { display: flex; flex-wrap: wrap; gap: 5px; }
.failed-tag {
  font-size: 11.5px; padding: 2px 8px; border-radius: 5px;
  border: 1px solid; white-space: nowrap;
}
.red-tag { color: #991b1b; background: #fee2e2; border-color: #fca5a5; }
.orange-tag { color: #92400e; background: #fef3c7; border-color: #fde68a; }

.pagination-row {
  display: flex;
  justify-content: center;
  padding-top: 4px;
}
</style>
