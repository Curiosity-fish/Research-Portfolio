<script setup lang="ts">
import { reactive, ref, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import RadarChart from '@/components/common/RadarChart.vue'
import AiInsightPanel from '@/components/common/AiInsightPanel.vue'
import { getStudentProfile, getStudentTerms, interpretMetrics } from '@/api'
import type { AcademicProfile, ProfileDimension } from '@/types'

const profile = reactive<AcademicProfile>({
  dimensions: [],
  gpaTrend: [],
  rankTrend: [],
})
const route = useRoute()
const router = useRouter()
const loading = ref(false)
const activeDimension = ref<ProfileDimension | null>(null)

// 学期选择
const selectedSemester = ref('')
const semesters = ref<string[]>([])

function formatSemester(term: string) {
  const parts = term.split('-')
  const semester = parts[2] === '2' ? '第二学期' : '第一学期'
  return parts.length === 3 ? `${parts[0]}-${parts[1]}学年 ${semester}` : term
}

async function loadTerms() {
  semesters.value = await getStudentTerms()
  const queryTerm = typeof route.query.term === 'string' ? route.query.term : ''
  selectedSemester.value = semesters.value.includes(queryTerm) ? queryTerm : (semesters.value[0] ?? '')
  if (!selectedSemester.value) await loadProfile()
}

async function loadProfile() {
  loading.value = true
  try {
    const res = await getStudentProfile(selectedSemester.value)
    Object.assign(profile, res)
    activeDimension.value = profile.dimensions[0] ?? null
  } finally {
    loading.value = false
  }
  loadAiProfile()
}

// AI 画像解读（本地模型基于真实画像数据生成）
const aiProfileLoading = ref(false)
const aiProfileSummary = ref('正在基于五维画像数据生成 AI 解读...')
const aiProfilePoints = ref<string[]>([])
const aiProfileCache = new Map<string, { summary: string; points: string[] }>()

async function loadAiProfile() {
  const cacheKey = selectedSemester.value
  const cached = aiProfileCache.get(cacheKey)
  if (cached) {
    aiProfileSummary.value = cached.summary
    aiProfilePoints.value = cached.points
    return
  }
  aiProfileLoading.value = true
  try {
    const res = await interpretMetrics(
      '请基于以下五维画像数据生成简明的AI画像解读：第一行输出一句总体评价，后续每行用“- ”开头输出分析或建议，最多5条；不要出现#、**、反引号，不要编造数据。',
      {
        dimensions: profile.dimensions.map(d => ({
          key: d.key,
          label: d.label,
          score: d.score,
          avgScore: d.avgScore,
          maxScore: d.maxScore,
          description: d.description,
          details: d.details,
        })),
        gpaTrend: profile.gpaTrend,
        rankTrend: profile.rankTrend,
      },
    )
    const clean = (s: string) => s.replace(/[#*`]/g, '').trim()
    const lines = res.content.split('\n').map(clean).filter(Boolean)
    const bullets = lines.filter(l => /^[-•]/.test(l)).map(l => l.replace(/^[-•]\s*/, ''))
    if (bullets.length > 0) {
      aiProfileSummary.value = lines.find(l => !/^[-•]/.test(l)) || 'AI 画像解读'
      aiProfilePoints.value = bullets.slice(0, 5)
    } else {
      aiProfileSummary.value = res.content.replace(/[#*`]/g, '').trim()
      aiProfilePoints.value = []
    }
    aiProfileCache.set(cacheKey, { summary: aiProfileSummary.value, points: aiProfilePoints.value })
  } catch {
    aiProfileSummary.value = 'AI 画像解读生成失败，请稍后重试'
    aiProfilePoints.value = []
  } finally {
    aiProfileLoading.value = false
  }
}

function scoreStatus(score: number) {
  if (score >= 80) return { label: '良好', color: '#3db87e', bg: 'rgba(61,184,126,0.1)' }
  if (score >= 65) return { label: '正常', color: '#2A4D99', bg: 'rgba(42,77,153,0.08)' }
  return { label: '注意', color: '#f09a4e', bg: 'rgba(240,154,78,0.1)' }
}

onMounted(loadTerms)

watch(selectedSemester, async semester => {
  if (!semester) return
  await router.replace({ query: { ...route.query, term: semester } })
  loadProfile()
})
</script>

<template>
  <div v-loading="loading" class="page">
    <div class="page-head">
      <div>
        <h2 class="page-title">学业画像</h2>
        <p class="page-sub">学业成绩 · 实践能力 · 综合素质 · 人文素养 · 身心健康</p>
      </div>
      <el-select v-if="semesters.length > 1" v-model="selectedSemester" placeholder="选择学期" style="width: 230px">
        <el-option
          v-for="item in semesters"
          :key="item"
          :label="formatSemester(item)"
          :value="item"
        />
      </el-select>
      <span v-else-if="selectedSemester" class="single-term">{{ formatSemester(selectedSemester) }}</span>
    </div>

    <div class="top-grid">
      <div class="card">
        <div class="card-title">五维雷达图</div>
        <RadarChart :dimensions="profile.dimensions.map(d => ({ label: d.label, score: d.score, avgScore: d.avgScore }))" />
      </div>

      <div class="card dim-card">
        <div class="card-title">维度详情</div>
        <div class="dim-list">
          <div
            v-for="dim in profile.dimensions" :key="dim.key"
            class="dim-row"
            :class="{ active: activeDimension?.key === dim.key }"
            @click="activeDimension = dim"
          >
            <span class="dim-name">{{ dim.label }}</span>
            <div class="dim-bar-wrap">
              <div class="dim-bar-track">
                <div class="dim-bar-fill" :style="{ width: dim.score + '%', background: scoreStatus(dim.score).color }" />
              </div>
            </div>
            <span class="dim-score" :style="{ color: scoreStatus(dim.score).color }">{{ dim.score }}</span>
            <span class="dim-badge" :style="{ color: scoreStatus(dim.score).color, background: scoreStatus(dim.score).bg }">
              {{ scoreStatus(dim.score).label }}
            </span>
          </div>
        </div>
        <div v-if="activeDimension" class="dim-detail">
          <div class="detail-desc">{{ activeDimension.description }}</div>
          <div class="detail-tags">
            <span v-for="d in activeDimension.details" :key="d" class="detail-tag">{{ d }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- AI 画像解读 -->
    <AiInsightPanel
      title="AI 画像解读"
      :summary="aiProfileSummary"
      :points="aiProfilePoints"
      :actions="[{ label: '查看提升建议', color: '#2A4D99' }, { label: '对比发展路径', color: '#3db87e' }]"
      :loading="aiProfileLoading"
      accent="#2A4D99"
    />
  </div>
</template>

<style scoped>
.page { padding: 28px 32px; display: flex; flex-direction: column; gap: 20px; }
.page-head { display: flex; align-items: center; justify-content: space-between; }
.page-title { font-size: 20px; font-weight: 800; color: #101d3e; margin: 0; letter-spacing: -0.01em; }
.page-sub { font-size: 13px; color: #9facc5; margin: 4px 0 0; }
.single-term {
  min-width: 210px; padding: 8px 12px; text-align: center;
  border: 1px solid #dcdfe6; border-radius: 4px;
  color: #606266; background: #fff; font-size: 14px;
}

.top-grid { display: grid; grid-template-columns: 300px 1fr; gap: 14px; }

.card {
  background: #fff; border-radius: 16px; padding: 20px 22px;
  border: 1px solid rgba(10,19,41,0.07);
  box-shadow: 0 1px 2px rgba(10,19,41,0.03), 0 4px 16px rgba(10,19,41,0.05);
}
.card-title { font-size: 14px; font-weight: 700; color: #111827; margin-bottom: 16px; }

.dim-list { display: flex; flex-direction: column; gap: 6px; }

.dim-row {
  display: grid;
  grid-template-columns: 56px 1fr 36px 44px;
  align-items: center;
  gap: 12px;
  padding: 10px 12px;
  border-radius: 10px;
  cursor: pointer;
  transition: background 0.15s ease;
}
.dim-row:hover { background: #f8f9fc; }
.dim-row.active { background: rgba(42,77,153,0.06); }

.dim-name {
  font-size: 13px; font-weight: 600; color: #374151;
  white-space: nowrap;
}
.dim-bar-wrap { display: flex; align-items: center; }
.dim-bar-track {
  width: 100%; height: 6px;
  background: #f0f2f7; border-radius: 99px; overflow: hidden;
}
.dim-bar-fill {
  height: 100%; border-radius: 99px;
  transition: width 0.5s cubic-bezier(0.34, 1.56, 0.64, 1);
}
.dim-score {
  font-size: 14px; font-weight: 800;
  text-align: right; font-variant-numeric: tabular-nums;
}
.dim-badge {
  font-size: 11px; font-weight: 700;
  padding: 2px 7px; border-radius: 20px;
  white-space: nowrap; text-align: center;
}

.dim-detail {
  margin-top: 14px; padding: 14px 16px;
  background: #f8f9fc; border-radius: 12px;
}
.detail-desc { font-size: 13px; color: #4b5563; font-weight: 500; margin-bottom: 10px; }
.detail-tags { display: flex; flex-wrap: wrap; gap: 6px; }
.detail-tag {
  font-size: 12px; color: #6b7280;
  background: #fff; border: 1px solid #e5e7eb;
  padding: 4px 10px; border-radius: 8px;
}
</style>
