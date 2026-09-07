<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import PathCard from '@/components/common/PathCard.vue'
import MilestoneTimeline from '@/components/common/MilestoneTimeline.vue'
import JobCard from '@/components/common/JobCard.vue'
import { getDevelopmentJobs, getDevelopmentPaths, getDevelopmentSchools } from '@/api'
import type { DevelopmentPath, JobItem, SchoolItem } from '@/types'

const router = useRouter()

// 全部五条路径
const paths = ref<DevelopmentPath[]>([])
const loading = ref(true)

const activeKey = ref<'graduate' | 'overseas' | 'employment' | 'civil' | 'institution'>('graduate')
const activePath = computed(() => paths.value.find(p => p.key === activeKey.value))

// 路径对应的智能体 id
const pathAgentMap: Record<string, string> = {
  graduate:    'grad-school-advisor',
  overseas:    'study-abroad-advisor',
  employment:  'job-advisor',
  civil:       'civil-service-advisor',
  institution: 'institution-exam-advisor',
}

function selectPath(key: 'graduate' | 'overseas' | 'employment' | 'civil' | 'institution') {
  activeKey.value = key
}

function goToAgent() {
  if (activePath.value) {
    router.push(`/student/agents/${pathAgentMap[activePath.value.key]}`)
  }
}

const urgencyConfig = {
  high:   { label: '紧迫', color: '#e04538', bg: '#fef2f2' },
  medium: { label: '中等', color: '#f09a4e', bg: '#fff7ed' },
  low:    { label: '较低', color: '#3db87e', bg: '#f0fdf4' },
}

const expandedAction = ref<number | null>(null)
function toggleAction(i: number) {
  expandedAction.value = expandedAction.value === i ? null : i
}

const activeJobCategory = ref<'backend' | 'ai' | 'pm'>('backend')
const jobsByCategory = ref<Record<string, JobItem[]>>({ backend: [], ai: [], pm: [] })
const filteredJobs = computed(() => jobsByCategory.value[activeJobCategory.value] ?? [])

type GradTierKey = 'top' | 'good' | 'match'
const activeGradTier = ref<GradTierKey>('good')
const gradSchools = ref<SchoolItem[]>([])
const filteredGradSchools = computed(() => gradSchools.value.filter(s => s.tier === activeGradTier.value))

type OverseasTierKey = 'top' | 'good' | 'match'
const activeOverseasTier = ref<OverseasTierKey>('good')
const overseasSchools = ref<SchoolItem[]>([])
const filteredOverseasSchools = computed(() => overseasSchools.value.filter(s => s.tier === activeOverseasTier.value))

const jobCategories = [
  { key: 'backend', label: '后端开发', desc: 'Java / Go / Python', color: '#2A4D99', bg: 'rgba(42,77,153,0.08)', icon: 'Cpu' },
  { key: 'ai', label: 'AI 算法', desc: '机器学习 / 大模型', color: '#7c3aed', bg: 'rgba(124,58,237,0.08)', icon: 'MagicStick' },
  { key: 'pm', label: '产品经理', desc: '需求 / 项目管理', color: '#f09a4e', bg: 'rgba(240,154,78,0.08)', icon: 'Briefcase' },
] as const

const gradSchoolTiers = [
  { key: 'top', label: '冲刺', desc: '高难度目标', color: '#e04538', bg: 'rgba(224,69,56,0.08)' },
  { key: 'good', label: '稳妥', desc: '匹配度高', color: '#2A4D99', bg: 'rgba(42,77,153,0.08)' },
  { key: 'match', label: '保底', desc: '录取概率高', color: '#3db87e', bg: 'rgba(61,184,126,0.08)' },
] as const

const overseasSchoolTiers = gradSchoolTiers

const window = {
  open: (url: string, target: string, features: string) => {
    globalThis.open(url, target, features)
  },
}

async function loadAll() {
  loading.value = true
  try {
    paths.value = await getDevelopmentPaths()
    const [backend, ai, pm] = await Promise.all([
      getDevelopmentJobs('backend', 1, 100),
      getDevelopmentJobs('ai', 1, 100),
      getDevelopmentJobs('pm', 1, 100),
    ])
    jobsByCategory.value = { backend: backend.list, ai: ai.list, pm: pm.list }
    const [grad, overseas] = await Promise.all([
      getDevelopmentSchools('grad', undefined, 1, 100),
      getDevelopmentSchools('overseas', undefined, 1, 100),
    ])
    gradSchools.value = grad.list
    overseasSchools.value = overseas.list
  } finally {
    loading.value = false
  }
}

onMounted(loadAll)
</script>

<template>
  <div class="page" v-loading="loading">
    <div>
      <h2 class="page-title">发展引导</h2>
      <p class="page-sub">选择你的发展方向，点击路径卡片进入专属 AI 智能体</p>
    </div>

    <!-- Path selector：三条路径 + 跳转按钮 -->
    <div class="path-grid">
      <PathCard
        v-for="p in paths" :key="p.key"
        :path-key="p.key" :label="p.label"
        :active="activeKey === p.key"
        @select="selectPath(p.key as 'graduate' | 'overseas' | 'employment' | 'civil' | 'institution')"
      />
    </div>

    <!-- 跳转到 AI 智能体的入口 -->
    <div class="agent-entry-bar">
      <div class="agent-entry-left">
        <span class="agent-entry-icon">{{
          activeKey === 'graduate' ? '🎓'
          : activeKey === 'overseas' ? '✈️'
          : activeKey === 'civil' ? '🏛️'
          : activeKey === 'institution' ? '📌'
          : '💼'
        }}</span>
        <div>
          <div class="agent-entry-name">{{
            activeKey === 'graduate' ? 'AI 考研助手'
            : activeKey === 'overseas' ? 'AI 留学助手'
            : activeKey === 'civil' ? 'AI 考公助手'
            : activeKey === 'institution' ? 'AI 考编助手'
            : 'AI 就业助手'
          }}</div>
          <div class="agent-entry-tip">点击进入专属 AI 智能体，获取个性化规划建议</div>
        </div>
      </div>
      <button class="agent-entry-btn" @click="goToAgent">立即咨询 →</button>
    </div>

    <!-- Gap analysis（原 two-col 只保留差距分析） -->
    <div v-if="activePath" class="card">
        <div class="card-title">差距分析</div>
        <div class="gap-list">
          <div v-for="gap in activePath.gapItems" :key="gap.dimension" class="gap-row">
            <div class="gap-dim">{{ gap.dimension }}</div>
            <div class="gap-values">
              <span class="gap-current">当前：{{ gap.current }}</span>
              <span class="gap-arrow">→</span>
              <span class="gap-target">目标：{{ gap.target }}</span>
            </div>
            <span
              class="gap-urgency"
              :style="{ color: urgencyConfig[gap.urgent].color, background: urgencyConfig[gap.urgent].bg }"
            >{{ urgencyConfig[gap.urgent].label }}</span>
          </div>
        </div>
    </div>

    <!-- Action checklist -->
    <div v-if="activePath" class="card">
      <div class="card-title">行动清单</div>
      <div class="action-list">
        <div
          v-for="(action, i) in activePath.actionItems" :key="i"
          class="action-item"
          @click="toggleAction(i)"
        >
          <div class="action-header">
            <span
              class="action-urgency"
              :style="{ color: urgencyConfig[action.urgency].color, background: urgencyConfig[action.urgency].bg }"
            >{{ urgencyConfig[action.urgency].label }}</span>
            <span class="action-title">{{ action.title }}</span>
            <div class="action-right">
              <span v-if="action.deadline" class="action-deadline">{{ action.deadline }}</span>
              <el-icon class="action-chevron" :class="{ open: expandedAction === i }"><ArrowRight /></el-icon>
            </div>
          </div>
          <div v-if="expandedAction === i" class="action-detail">{{ action.description }}</div>
        </div>
      </div>
    </div>

    <!-- Milestone timeline -->
    <div v-if="activePath" class="card">
      <div class="card-title">关键里程碑</div>
      <MilestoneTimeline :milestones="activePath.milestones" />
    </div>

    <!-- Job recommendation — employment path only -->
    <template v-if="activeKey === 'employment'">
      <div class="card jobs-section">
        <div class="jobs-header">
          <div class="jobs-header-left">
            <div class="jobs-header-icon">
              <el-icon><Briefcase /></el-icon>
            </div>
            <div>
              <div class="card-title" style="margin-bottom: 2px">就业岗位推荐</div>
              <div class="jobs-header-sub">基于你的学业画像，智能匹配浙江省内在招岗位 · 数据来源：BOSS直聘 / 前程无忧 / 智联招聘</div>
            </div>
          </div>
          <div class="jobs-freshness">
            <span class="freshness-dot" />
            <span>数据更新于 2026-06-10</span>
          </div>
        </div>

        <!-- 职位方向选择 -->
        <div class="job-cat-grid">
          <button
            v-for="cat in jobCategories" :key="cat.key"
            class="job-cat-btn"
            :class="{ active: activeJobCategory === cat.key }"
            :style="activeJobCategory === cat.key ? { borderColor: cat.color, background: cat.bg } : {}"
            @click="activeJobCategory = cat.key"
          >
            <div class="job-cat-icon" :style="{ background: cat.bg, color: cat.color }">
              <el-icon><component :is="cat.icon" /></el-icon>
            </div>
            <div class="job-cat-body">
              <div class="job-cat-label" :style="activeJobCategory === cat.key ? { color: cat.color } : {}">{{ cat.label }}</div>
              <div class="job-cat-desc">{{ cat.desc }}</div>
            </div>
            <div class="job-cat-count" :style="{ color: cat.color }">
              {{ (jobsByCategory[cat.key] || []).length }} 个岗位
            </div>
          </button>
        </div>

        <!-- 岗位卡片列表 -->
        <div class="jobs-grid">
          <JobCard v-for="job in filteredJobs" :key="job.id" :job="job" />
        </div>

        <div class="jobs-footer">
          <el-icon><InfoFilled /></el-icon>
          <span>以上岗位来自后端职位缓存，由实时爬虫服务定期更新。点击「查看详情」将跳转至原招聘平台页面。</span>
        </div>
      </div>
    </template>

    <!-- Grad school recommendation — graduate path only -->
    <template v-if="activeKey === 'graduate'">
      <div class="card school-section">
        <div class="school-header">
          <div class="school-header-left">
            <div class="school-header-icon" style="background: rgba(42,77,153,0.1); color: #2A4D99">
              <el-icon><School /></el-icon>
            </div>
            <div>
              <div class="card-title" style="margin-bottom: 2px">考研院校推荐</div>
              <div class="school-header-sub">基于你的 GPA 3.62 · 专业背景 · 学科排名，推荐最匹配的国内研究生院校</div>
            </div>
          </div>
          <div class="school-freshness">
            <span class="freshness-dot-blue" />
            <span>数据参考2025年录取情况</span>
          </div>
        </div>

        <div class="school-tier-btns">
          <button
            v-for="tier in gradSchoolTiers" :key="tier.key"
            class="tier-btn"
            :class="{ active: activeGradTier === tier.key }"
            :style="activeGradTier === tier.key ? { borderColor: tier.color, background: tier.bg, color: tier.color } : {}"
            @click="activeGradTier = tier.key"
          >
            <span class="tier-dot" :style="{ background: tier.color }" />
            <div>
              <div class="tier-btn-label">{{ tier.label }}</div>
              <div class="tier-btn-desc">{{ tier.desc }}</div>
            </div>
            <span class="tier-count" :style="{ color: tier.color }">
              {{ gradSchools.filter(s => s.tier === tier.key).length }} 所
            </span>
          </button>
        </div>

        <div class="school-grid">
          <div
            v-for="school in filteredGradSchools" :key="school.id"
            class="school-card"
          >
            <div class="school-card-top">
              <div class="school-avatar">{{ school.name[0] }}</div>
              <div class="school-info">
                <div class="school-name">{{ school.name }}</div>
                <div class="school-meta">{{ school.location }} · {{ school.rank }}</div>
              </div>
              <div
                class="school-match-badge"
                :style="{
                  color: school.matchScore >= 85 ? '#3db87e' : school.matchScore >= 70 ? '#2A4D99' : '#f09a4e',
                  background: school.matchScore >= 85 ? 'rgba(61,184,126,0.08)' : school.matchScore >= 70 ? 'rgba(42,77,153,0.08)' : 'rgba(240,154,78,0.08)',
                }"
              >{{ school.matchScore }}% 匹配</div>
            </div>

            <div class="school-programs">
              <span v-for="p in school.programs.slice(0, 2)" :key="p" class="program-chip">{{ p }}</span>
            </div>

            <div class="school-req-row">
              <div class="req-item">
                <span class="req-label">录取GPA</span>
                <span class="req-value">{{ school.admissionGpa }}</span>
              </div>
              <div class="req-item">
                <span class="req-label">考试科目</span>
                <span class="req-value">{{ school.examRequirements.join(' / ') }}</span>
              </div>
            </div>

            <div class="school-highlights">
              <span v-for="h in school.highlights.slice(0, 3)" :key="h" class="highlight-chip">{{ h }}</span>
            </div>

            <div class="school-match-bar-row">
              <span class="match-bar-label">画像匹配</span>
              <div class="match-bar-track">
                <div
                  class="match-bar-fill"
                  :style="{
                    width: school.matchScore + '%',
                    background: school.matchScore >= 85 ? '#3db87e' : school.matchScore >= 70 ? '#2A4D99' : '#f09a4e'
                  }"
                />
              </div>
            </div>

            <ul class="school-reasons">
              <li v-for="r in school.matchReasons" :key="r">{{ r }}</li>
            </ul>

            <div class="school-card-footer">
              <button class="btn-school" @click="() => window.open(school.officialUrl, '_blank', 'noopener,noreferrer')">
                访问官网
                <el-icon><Link /></el-icon>
              </button>
            </div>
          </div>
        </div>

        <div class="school-disclaimer">
          <el-icon><InfoFilled /></el-icon>
          <span>以上推荐基于历史录取数据与学业画像匹配，仅供参考，实际录取情况以各院校官方公告为准。</span>
        </div>
      </div>
    </template>

    <!-- Overseas school recommendation — overseas path only -->
    <template v-if="activeKey === 'overseas'">
      <div class="card school-section">
        <div class="school-header">
          <div class="school-header-left">
            <div class="school-header-icon" style="background: rgba(139,92,246,0.1); color: #8b5cf6">
              <el-icon><Promotion /></el-icon>
            </div>
            <div>
              <div class="card-title" style="margin-bottom: 2px">留学院校推荐</div>
              <div class="school-header-sub">基于你的 GPA 3.62 · 专业背景 · 留学意向，推荐最匹配的海外研究生院校</div>
            </div>
          </div>
          <div class="school-freshness">
            <span class="freshness-dot-purple" />
            <span>数据参考2025/26申请季</span>
          </div>
        </div>

        <div class="school-tier-btns">
          <button
            v-for="tier in overseasSchoolTiers" :key="tier.key"
            class="tier-btn"
            :class="{ active: activeOverseasTier === tier.key }"
            :style="activeOverseasTier === tier.key ? { borderColor: tier.color, background: tier.bg, color: tier.color } : {}"
            @click="activeOverseasTier = tier.key"
          >
            <span class="tier-dot" :style="{ background: tier.color }" />
            <div>
              <div class="tier-btn-label">{{ tier.label }}</div>
              <div class="tier-btn-desc">{{ tier.desc }}</div>
            </div>
            <span class="tier-count" :style="{ color: tier.color }">
              {{ overseasSchools.filter(s => s.tier === tier.key).length }} 所
            </span>
          </button>
        </div>

        <div class="school-grid">
          <div
            v-for="school in filteredOverseasSchools" :key="school.id"
            class="school-card overseas-card"
          >
            <div class="school-card-top">
              <div class="school-avatar overseas-avatar">{{ school.countryEmoji }}</div>
              <div class="school-info">
                <div class="school-name">{{ school.nameZh }}</div>
                <div class="school-meta">{{ school.name }}</div>
                <div class="school-country">{{ school.country }} · {{ school.rank }}</div>
              </div>
              <div
                class="school-match-badge"
                :style="{
                  color: school.matchScore >= 85 ? '#3db87e' : school.matchScore >= 70 ? '#2A4D99' : '#f09a4e',
                  background: school.matchScore >= 85 ? 'rgba(61,184,126,0.08)' : school.matchScore >= 70 ? 'rgba(42,77,153,0.08)' : 'rgba(240,154,78,0.08)',
                }"
              >{{ school.matchScore }}% 匹配</div>
            </div>

            <div class="school-programs">
              <span v-for="p in school.programs.slice(0, 1)" :key="p" class="program-chip purple">{{ p }}</span>
            </div>

            <div class="school-req-row">
              <div class="req-item">
                <span class="req-label">录取GPA</span>
                <span class="req-value">{{ school.admissionGpa }}</span>
              </div>
              <div class="req-item">
                <span class="req-label">语言要求</span>
                <span class="req-value">{{ school.languageRequirement }}</span>
              </div>
            </div>

            <div class="school-highlights">
              <span v-for="h in school.highlights.slice(0, 3)" :key="h" class="highlight-chip purple">{{ h }}</span>
            </div>

            <div class="school-match-bar-row">
              <span class="match-bar-label">画像匹配</span>
              <div class="match-bar-track">
                <div
                  class="match-bar-fill"
                  :style="{
                    width: school.matchScore + '%',
                    background: school.matchScore >= 85 ? '#3db87e' : school.matchScore >= 70 ? '#2A4D99' : '#8b5cf6'
                  }"
                />
              </div>
            </div>

            <ul class="school-reasons">
              <li v-for="r in school.matchReasons" :key="r">{{ r }}</li>
            </ul>

            <div class="school-card-footer">
              <span class="deadline-badge">
                <el-icon><Calendar /></el-icon>
                截止 {{ school.deadline }}
              </span>
              <button class="btn-school purple" @click="() => window.open(school.officialUrl, '_blank', 'noopener,noreferrer')">
                访问官网
                <el-icon><Link /></el-icon>
              </button>
            </div>
          </div>
        </div>

        <div class="school-disclaimer">
          <el-icon><InfoFilled /></el-icon>
          <span>以上推荐基于历史录取数据与学业画像匹配，仅供参考。申请截止日期及录取要求请以各院校官网公告为准。</span>
        </div>
      </div>
    </template>

  </div>
</template>

<style scoped>
.page { padding: 28px 32px; display: flex; flex-direction: column; gap: 20px; }
.page-title { font-size: 20px; font-weight: 800; color: #101d3e; margin: 0; letter-spacing: -0.01em; }
.page-sub { font-size: 13px; color: #9facc5; margin: 4px 0 0; }

.path-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 14px; }

/* ── AI 智能体入口条 ── */
.agent-entry-bar {
  display: flex; align-items: center; justify-content: space-between;
  background: linear-gradient(135deg, #0c1a38 0%, #1a2d5a 60%, #2A4D99 100%);
  border-radius: 14px; padding: 16px 20px;
}
.agent-entry-left { display: flex; align-items: center; gap: 14px; }
.agent-entry-icon { font-size: 28px; line-height: 1; }
.agent-entry-name { font-size: 14px; font-weight: 700; color: #fff; margin-bottom: 3px; }
.agent-entry-tip { font-size: 12px; color: rgba(255,255,255,0.6); }
.agent-entry-btn {
  padding: 9px 20px; background: #fff; color: #2A4D99;
  border: none; border-radius: 20px; font-size: 13px; font-weight: 700;
  cursor: pointer; white-space: nowrap; transition: opacity 0.15s;
  flex-shrink: 0;
}
.agent-entry-btn:hover { opacity: 0.9; }

.card {
  background: #fff; border-radius: 14px; border: 1px solid #edf0f6;
  padding: 20px; box-shadow: 0 1px 4px rgba(0,0,0,0.04);
}
.card-title { font-size: 14px; font-weight: 700; color: #101d3e; margin-bottom: 16px; }

.match-score-wrap { display: flex; align-items: center; gap: 20px; }
.match-ring { flex-shrink: 0; }
.match-desc { flex: 1; }
.match-path-label { font-size: 16px; font-weight: 700; color: #2A4D99; margin-bottom: 8px; }
.match-tip { font-size: 13px; color: #6b7280; line-height: 1.6; margin: 0; }

.gap-list { display: flex; flex-direction: column; gap: 10px; }
.gap-row { display: flex; align-items: center; gap: 10px; padding: 10px 12px; background: #f9fafb; border-radius: 8px; }
.gap-dim { font-size: 13px; font-weight: 600; color: #374151; min-width: 80px; }
.gap-values { display: flex; align-items: center; gap: 6px; flex: 1; font-size: 12px; color: #6b7280; }
.gap-current { color: #6b7280; }
.gap-arrow { color: #9ca3af; }
.gap-target { color: #2A4D99; font-weight: 600; }
.gap-urgency { font-size: 11px; font-weight: 600; padding: 2px 7px; border-radius: 5px; white-space: nowrap; }

.action-list { display: flex; flex-direction: column; gap: 2px; }
.action-item {
  border-radius: 10px; overflow: hidden; cursor: pointer;
  border: 1px solid #edf0f6; transition: border-color 0.15s;
}
.action-item:hover { border-color: #c8d4e8; }
.action-header {
  display: flex; align-items: center; gap: 10px;
  padding: 12px 14px;
}
.action-urgency { font-size: 11px; font-weight: 600; padding: 2px 7px; border-radius: 5px; white-space: nowrap; flex-shrink: 0; }
.action-title { flex: 1; font-size: 13.5px; font-weight: 600; color: #1f2937; }
.action-right { display: flex; align-items: center; gap: 8px; flex-shrink: 0; }
.action-deadline { font-size: 12px; color: #9ca3af; }
.action-chevron { font-size: 12px; color: #9ca3af; transition: transform 0.2s; }
.action-chevron.open { transform: rotate(90deg); }
.action-detail {
  padding: 0 14px 12px 14px; font-size: 13px; color: #6b7280;
  line-height: 1.6; border-top: 1px solid #f0f0f0; padding-top: 10px;
}

.placeholder-card {
  display: flex; align-items: center; gap: 14px;
  padding: 18px 20px; border-radius: 14px;
  border: 1px dashed #c8d4e8; background: #fafbfd;
}
.placeholder-icon { font-size: 24px; flex-shrink: 0; }
.placeholder-title { font-size: 14px; font-weight: 600; color: #374151; }
.placeholder-desc { font-size: 12.5px; color: #9ca3af; margin-top: 2px; }
.placeholder-badge {
  margin-left: auto; flex-shrink: 0;
  font-size: 11px; font-weight: 600; color: #7c3aed;
  background: #f5f3ff; padding: 3px 10px; border-radius: 6px;
}

/* ── Job Recommendation ── */
.jobs-section { padding: 20px 22px; }

.jobs-header {
  display: flex; align-items: flex-start;
  justify-content: space-between; gap: 16px;
  margin-bottom: 20px; flex-wrap: wrap;
}
.jobs-header-left { display: flex; align-items: center; gap: 12px; }
.jobs-header-icon {
  width: 36px; height: 36px; border-radius: 10px; flex-shrink: 0;
  background: rgba(61,184,126,0.1); color: #3db87e;
  display: flex; align-items: center; justify-content: center;
  font-size: 17px;
}
.jobs-header-sub { font-size: 12px; color: #9facc5; margin-top: 2px; }

.jobs-freshness {
  display: flex; align-items: center; gap: 5px;
  font-size: 11.5px; color: #9facc5; flex-shrink: 0;
}
.freshness-dot {
  width: 6px; height: 6px; border-radius: 50%;
  background: #3db87e;
  box-shadow: 0 0 0 3px rgba(61,184,126,0.2);
}

.job-cat-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 10px; margin-bottom: 20px; }

.job-cat-btn {
  display: flex; align-items: center; gap: 12px;
  padding: 14px 16px; border-radius: 12px;
  border: 1.5px solid #e5e9f0; background: #fff;
  cursor: pointer; transition: all 0.15s ease; text-align: left;
}
.job-cat-btn:hover { background: #f8f9fc; border-color: #d1d8e8; }
.job-cat-btn.active { box-shadow: 0 2px 10px rgba(0,0,0,0.07); }

.job-cat-icon {
  width: 36px; height: 36px; border-radius: 9px; flex-shrink: 0;
  display: flex; align-items: center; justify-content: center;
  font-size: 17px;
}
.job-cat-body { flex: 1; min-width: 0; }
.job-cat-label { font-size: 13.5px; font-weight: 700; color: #374151; margin-bottom: 3px; }
.job-cat-desc { font-size: 11.5px; color: #9facc5; line-height: 1.4; }
.job-cat-count { font-size: 12px; font-weight: 800; flex-shrink: 0; }

.jobs-grid { display: grid; grid-template-columns: repeat(2, 1fr); gap: 14px; margin-bottom: 16px; }

.jobs-footer {
  display: flex; align-items: flex-start; gap: 6px;
  font-size: 11.5px; color: #b4bed2; line-height: 1.5;
  padding: 10px 14px; background: #f8f9fc; border-radius: 9px;
}
.jobs-footer .el-icon { font-size: 13px; flex-shrink: 0; margin-top: 1px; }

/* ── School Recommendation ── */
.school-section { padding: 20px 22px; }

.school-header {
  display: flex; align-items: flex-start;
  justify-content: space-between; gap: 16px;
  margin-bottom: 20px; flex-wrap: wrap;
}
.school-header-left { display: flex; align-items: center; gap: 12px; }
.school-header-icon {
  width: 36px; height: 36px; border-radius: 10px; flex-shrink: 0;
  display: flex; align-items: center; justify-content: center;
  font-size: 17px;
}
.school-header-sub { font-size: 12px; color: #9facc5; margin-top: 2px; }

.school-freshness {
  display: flex; align-items: center; gap: 5px;
  font-size: 11.5px; color: #9facc5; flex-shrink: 0;
}
.freshness-dot-blue {
  width: 6px; height: 6px; border-radius: 50%;
  background: #2A4D99; box-shadow: 0 0 0 3px rgba(42,77,153,0.15);
}
.freshness-dot-purple {
  width: 6px; height: 6px; border-radius: 50%;
  background: #8b5cf6; box-shadow: 0 0 0 3px rgba(139,92,246,0.15);
}

.school-tier-btns {
  display: grid; grid-template-columns: repeat(3, 1fr);
  gap: 10px; margin-bottom: 20px;
}
.tier-btn {
  display: flex; align-items: center; gap: 10px;
  padding: 12px 14px; border-radius: 12px;
  border: 1.5px solid #e5e9f0; background: #fff;
  cursor: pointer; transition: all 0.15s ease; text-align: left;
}
.tier-btn:hover { background: #f8f9fc; border-color: #d1d8e8; }
.tier-btn.active { box-shadow: 0 2px 10px rgba(0,0,0,0.07); }
.tier-dot { width: 8px; height: 8px; border-radius: 50%; flex-shrink: 0; }
.tier-btn > div { flex: 1; min-width: 0; }
.tier-btn-label { font-size: 13px; font-weight: 700; color: #374151; margin-bottom: 2px; }
.tier-btn-desc { font-size: 11px; color: #9facc5; }
.tier-btn.active .tier-btn-label { color: inherit; }
.tier-count { font-size: 12px; font-weight: 800; flex-shrink: 0; }

.school-grid {
  display: grid; grid-template-columns: repeat(3, 1fr);
  gap: 14px; margin-bottom: 16px;
}

.school-card {
  background: #fafbfd; border-radius: 14px;
  border: 1px solid #edf0f6; padding: 16px 18px;
  display: flex; flex-direction: column; gap: 11px;
  transition: box-shadow 0.18s ease, transform 0.18s ease;
}
.school-card:hover {
  box-shadow: 0 4px 14px rgba(10,19,41,0.08);
  transform: translateY(-1px);
  background: #fff;
}

.school-card-top { display: flex; align-items: flex-start; gap: 10px; }
.school-avatar {
  width: 38px; height: 38px; border-radius: 10px; flex-shrink: 0;
  background: linear-gradient(135deg, #e8edf8, #d0d9f0);
  display: flex; align-items: center; justify-content: center;
  font-size: 14px; font-weight: 800; color: #2A4D99;
}
.overseas-avatar {
  background: #f4f4f4; font-size: 20px; color: unset;
}
.school-info { flex: 1; min-width: 0; }
.school-name { font-size: 14px; font-weight: 700; color: #111827; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.school-meta { font-size: 11px; color: #9facc5; margin-top: 1px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.school-country { font-size: 11px; color: #6b7280; margin-top: 1px; }
.school-match-badge {
  font-size: 11px; font-weight: 700;
  padding: 3px 8px; border-radius: 6px; flex-shrink: 0; white-space: nowrap;
}

.school-programs { display: flex; gap: 6px; flex-wrap: wrap; }
.program-chip {
  font-size: 11px; color: #2A4D99;
  background: rgba(42,77,153,0.07);
  padding: 2px 8px; border-radius: 5px;
}
.program-chip.purple { color: #7c3aed; background: rgba(124,58,237,0.07); }

.school-req-row { display: flex; flex-direction: column; gap: 5px; }
.req-item { display: flex; align-items: flex-start; gap: 8px; }
.req-label { font-size: 11.5px; font-weight: 600; color: #9facc5; white-space: nowrap; min-width: 48px; flex-shrink: 0; }
.req-value { font-size: 11.5px; color: #4b5563; line-height: 1.5; }

.school-highlights { display: flex; gap: 5px; flex-wrap: wrap; }
.highlight-chip {
  font-size: 11px; color: #374151;
  background: #f0f2f7; padding: 2px 7px; border-radius: 5px;
}
.highlight-chip.purple { color: #7c3aed; background: rgba(124,58,237,0.07); }

.school-match-bar-row { display: flex; align-items: center; gap: 8px; }
.match-bar-label { font-size: 11px; color: #9facc5; font-weight: 600; white-space: nowrap; }
.match-bar-track { flex: 1; height: 4px; background: #edf0f6; border-radius: 99px; overflow: hidden; }
.match-bar-fill { height: 100%; border-radius: 99px; transition: width 0.5s cubic-bezier(0.34,1.56,0.64,1); }

.school-reasons {
  margin: 0; padding: 0; list-style: none;
  display: flex; flex-direction: column; gap: 4px;
}
.school-reasons li {
  font-size: 11.5px; color: #6b7280; padding-left: 10px; position: relative; line-height: 1.5;
}
.school-reasons li::before {
  content: ''; position: absolute; left: 0; top: 7px;
  width: 3px; height: 3px; border-radius: 50%; background: #3db87e;
}

.school-card-footer {
  display: flex; align-items: center; justify-content: space-between;
  margin-top: auto; padding-top: 4px;
}
.deadline-badge {
  display: flex; align-items: center; gap: 4px;
  font-size: 11px; color: #9facc5;
}
.deadline-badge .el-icon { font-size: 11px; }

.btn-school {
  display: flex; align-items: center; gap: 4px;
  padding: 6px 12px; border-radius: 8px;
  background: #2A4D99; color: #fff;
  font-size: 12px; font-weight: 700;
  border: none; cursor: pointer; flex-shrink: 0;
  transition: background 0.15s ease, transform 0.15s ease;
}
.btn-school:hover { background: #3a5fba; transform: translateY(-1px); }
.btn-school:active { transform: translateY(0); }
.btn-school.purple { background: #7c3aed; }
.btn-school.purple:hover { background: #6d28d9; }
.btn-school .el-icon { font-size: 12px; }

.school-disclaimer {
  display: flex; align-items: flex-start; gap: 6px;
  font-size: 11.5px; color: #b4bed2; line-height: 1.5;
  padding: 10px 14px; background: #f8f9fc; border-radius: 9px;
}
.school-disclaimer .el-icon { font-size: 13px; flex-shrink: 0; margin-top: 1px; }

</style>
