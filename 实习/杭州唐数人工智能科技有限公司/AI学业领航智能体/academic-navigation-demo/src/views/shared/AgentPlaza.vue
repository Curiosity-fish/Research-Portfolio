<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { agents, type AgentConfig } from '@/mock/agents'

const router = useRouter()
const auth = useAuthStore()
const role = computed(() => auth.user?.role || 'student')

const searchQuery = ref('')
const activeCategory = ref('全部')

const allVisible = computed(() => agents.filter(a => a.roles.includes(role.value)))

const roleCategories = computed(() => {
  const seen = new Set<string>()
  const cats: string[] = []
  for (const a of allVisible.value) {
    if (!seen.has(a.category)) { seen.add(a.category); cats.push(a.category) }
  }
  return ['全部', ...cats]
})

const filtered = computed(() => {
  let list = allVisible.value
  if (searchQuery.value) {
    const q = searchQuery.value.toLowerCase()
    return list.filter(a => a.name.toLowerCase().includes(q) || a.description.toLowerCase().includes(q))
  }
  if (activeCategory.value !== '全部') {
    return list.filter(a => a.category === activeCategory.value)
  }
  return list
})

const sections = computed(() => {
  if (searchQuery.value || activeCategory.value !== '全部') {
    return [{ label: '', items: filtered.value }]
  }
  // 全部 tab：不分类，直接平铺
  return [{ label: '', items: allVisible.value }]
})

const heroMap: Record<string, { title: string; sub: string }> = {
  student:        { title: 'AI 领航助手', sub: '你的专属智能助手，7×24 在线，帮你解决学业与发展的每一个问题' },
  teacher:        { title: 'AI 领航助手', sub: '覆盖班级管理全场景，让每一项日常工作都有智能支撑' },
  course_teacher: { title: 'AI 领航助手', sub: '从备课到批改，全流程 AI 辅助，让教学更高效、更有深度' },
  department:     { title: 'AI 领航助手', sub: '系部教学质量管理的智能决策中枢，数据驱动每一个管理判断' },
  dean:           { title: 'AI 领航助手', sub: '院级视角的全局数据洞察，辅助战略决策与资源配置' },
}

function formatCount(n: number) {
  if (n >= 10000) return (n / 10000).toFixed(1) + 'w'
  if (n >= 1000) return (n / 1000).toFixed(1) + 'k'
  return n.toString()
}

const routePrefix: Record<string, string> = {
  student:        'student',
  teacher:        'teacher',
  course_teacher: 'course-teacher',
  department:     'department',
  dean:           'dean',
}

function openAgent(agent: AgentConfig) {
  if (agent.status === 'coming_soon') return
  const prefix = routePrefix[role.value] ?? role.value
  router.push(`/${prefix}/agents/${agent.id}`)
}
</script>

<template>
  <div class="plaza">

    <!-- ══ 粘性头部：Hero + Tabs ══ -->
    <div class="sticky-header">

    <!-- ══ Hero Banner ══ -->
    <div class="hero">
      <div class="hero-bg" />
      <div class="hero-content">
        <div class="hero-left">
          <div class="hero-badge">
            <svg viewBox="0 0 24 24" fill="currentColor" class="hero-badge-icon">
              <path d="M12 2l2.4 7.4H22l-6.2 4.5 2.4 7.4L12 17l-6.2 4.3 2.4-7.4L2 9.4h7.6z"/>
            </svg>
            智能助手平台
          </div>
          <h1 class="hero-title">{{ heroMap[role]?.title ?? 'AI 领航助手' }}</h1>
          <p class="hero-sub">{{ heroMap[role]?.sub }}</p>
          <div class="hero-stats">
            <div class="hero-stat">
              <span class="hero-stat-num">{{ allVisible.filter(a => a.status === 'active').length }}</span>
              <span class="hero-stat-label">个助手上线</span>
            </div>
            <div class="hero-divider" />
            <div class="hero-stat">
              <span class="hero-stat-num">{{ roleCategories.length - 1 }}</span>
              <span class="hero-stat-label">个功能分类</span>
            </div>
            <div class="hero-divider" />
            <div class="hero-stat">
              <span class="hero-stat-num">7×24</span>
              <span class="hero-stat-label">随时可用</span>
            </div>
          </div>
        </div>
        <div class="hero-right">
          <div class="search-wrap">
            <svg class="search-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="11" cy="11" r="8"/><path d="m21 21-4.35-4.35"/>
            </svg>
            <input v-model="searchQuery" class="search-input" placeholder="搜索助手名称或功能..." />
          </div>
        </div>
      </div>
    </div>

    <!-- ══ 分类 Tabs ══ -->
    <div class="tabs-row">
      <div class="tabs">
        <button
          v-for="cat in roleCategories" :key="cat"
          class="tab" :class="{ active: activeCategory === cat }"
          @click="activeCategory = cat; searchQuery = ''"
        >{{ cat }}</button>
      </div>
      <span v-if="searchQuery" class="search-result-tip">
        找到 {{ filtered.length }} 个相关助手
      </span>
    </div>

    </div><!-- /sticky-header -->

    <!-- ══ 助手卡片 ══ -->
    <div class="body">
      <template v-if="sections.length > 0">
        <div v-for="section in sections" :key="section.label" class="section">
          <div v-if="section.label" class="section-header">
            <span class="section-name">{{ section.label }}</span>
            <span class="section-rule" />
          </div>
          <div class="card-grid">
            <div
              v-for="agent in section.items" :key="agent.id"
              class="agent-card"
              :class="{ 'is-soon': agent.status === 'coming_soon' }"
              :style="{ '--accent': agent.accentColor }"
              @click="openAgent(agent)"
            >
              <!-- 封面图 -->
              <div class="card-img-wrap">
                <img :src="agent.imgUrl" :alt="agent.name" class="card-img" loading="lazy" />
                <span v-if="agent.status === 'coming_soon'" class="badge-soon">即将上线</span>
              </div>

              <!-- 内容区 -->
              <div class="card-content">
                <div class="card-top">
                  <span class="cat-pill">{{ agent.category }}</span>
                  <span class="usage-count">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/>
                      <path d="M23 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/>
                    </svg>
                    {{ formatCount(agent.usageCount) }}
                  </span>
                </div>
                <h3 class="card-name">{{ agent.name }}</h3>
                <p class="card-desc">{{ agent.description }}</p>
                <div class="card-footer">
                  <div class="data-dots">
                    <span v-for="d in (agent.dataUsed || []).slice(0, 2)" :key="d" class="data-dot">{{ d }}</span>
                  </div>
                  <span class="enter-btn" v-if="agent.status === 'active'">
                    开始使用
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M5 12h14M12 5l7 7-7 7"/></svg>
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </template>
      <div v-else class="empty">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
          <circle cx="11" cy="11" r="8"/><path d="m21 21-4.35-4.35"/>
        </svg>
        <p>未找到相关助手</p>
      </div>
    </div>

  </div>
</template>

<style scoped>
/* ══ 页面底色 ══ */
.plaza {
  display: flex;
  flex-direction: column;
  min-height: 100%;
  background: #f0f2f8;
}

/* ══ 粘性头部 ══ */
.sticky-header {
  position: sticky;
  top: 0;
  z-index: 10;
  background: #f0f2f8;
}

/* ══ Hero ══ */
.hero {
  position: relative;
  overflow: hidden;
  padding: 36px 44px 32px;
  background: linear-gradient(135deg, #0b1830 0%, #162544 45%, #1e3a7a 100%);
}
.hero-bg {
  position: absolute; inset: 0; pointer-events: none;
  background:
    radial-gradient(ellipse 55% 100% at 90% 50%, rgba(85,128,212,0.22) 0%, transparent 65%),
    radial-gradient(ellipse 35% 70% at 5% 90%, rgba(124,58,237,0.14) 0%, transparent 60%),
    radial-gradient(circle 200px at 50% -20%, rgba(255,255,255,0.04) 0%, transparent 70%);
}
/* 装饰圆环 */
.hero-bg::before {
  content: '';
  position: absolute;
  right: -60px; top: -60px;
  width: 320px; height: 320px;
  border-radius: 50%;
  border: 1px solid rgba(255,255,255,0.06);
}
.hero-bg::after {
  content: '';
  position: absolute;
  right: 20px; top: -20px;
  width: 200px; height: 200px;
  border-radius: 50%;
  border: 1px solid rgba(255,255,255,0.04);
}
.hero-content {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 40px;
}
.hero-left { flex: 1; min-width: 0; }
.hero-badge {
  display: inline-flex; align-items: center; gap: 6px;
  font-size: 11px; font-weight: 600; letter-spacing: 0.08em;
  color: rgba(255,255,255,0.55);
  background: rgba(255,255,255,0.08);
  border: 1px solid rgba(255,255,255,0.1);
  padding: 5px 12px; border-radius: 20px;
  margin-bottom: 14px;
  backdrop-filter: blur(4px);
}
.hero-badge-icon { width: 10px; height: 10px; color: #fbbf24; }
.hero-title {
  font-size: 28px; font-weight: 800; color: #fff;
  margin: 0 0 10px; letter-spacing: -0.025em; line-height: 1.15;
}
.hero-sub {
  font-size: 13.5px; color: rgba(255,255,255,0.5);
  margin: 0 0 24px; line-height: 1.65; max-width: 500px;
}
.hero-stats { display: flex; align-items: stretch; gap: 0; }
.hero-stat {
  display: flex; flex-direction: column; align-items: center;
  padding: 10px 20px;
  background: rgba(255,255,255,0.06);
  border: 1px solid rgba(255,255,255,0.08);
  backdrop-filter: blur(6px);
}
.hero-stat:first-child { border-radius: 10px 0 0 10px; }
.hero-stat:last-child { border-radius: 0 10px 10px 0; }
.hero-stat + .hero-stat { border-left: none; }
.hero-stat-num { font-size: 20px; font-weight: 800; color: #fff; line-height: 1; }
.hero-stat-label { font-size: 11px; color: rgba(255,255,255,0.4); margin-top: 3px; }

.hero-right { flex-shrink: 0; }
.search-wrap { position: relative; }
.search-icon {
  position: absolute; left: 14px; top: 50%; transform: translateY(-50%);
  width: 16px; height: 16px; color: rgba(255,255,255,0.38); pointer-events: none;
}
.search-input {
  width: 300px; height: 44px; padding: 0 18px 0 42px;
  border-radius: 12px;
  border: 1px solid rgba(255,255,255,0.14);
  background: rgba(255,255,255,0.09);
  color: #fff; font-size: 13.5px; outline: none; font-family: inherit;
  backdrop-filter: blur(8px);
  transition: background 0.15s, border-color 0.15s, box-shadow 0.15s;
}
.search-input::placeholder { color: rgba(255,255,255,0.28); }
.search-input:focus {
  background: rgba(255,255,255,0.14);
  border-color: rgba(255,255,255,0.28);
  box-shadow: 0 0 0 3px rgba(255,255,255,0.06);
}

/* ══ Tabs ══ */
.tabs-row {
  display: flex; align-items: center; justify-content: space-between;
  padding: 22px 44px 0; gap: 16px;
}
.tabs { display: flex; gap: 6px; flex-wrap: wrap; }
.tab {
  padding: 7px 18px; border-radius: 20px;
  border: 1.5px solid #e0e4f0; background: #fff;
  font-size: 13px; font-weight: 500; color: #6b7280;
  cursor: pointer; transition: all 0.15s; font-family: inherit;
  box-shadow: 0 1px 3px rgba(0,0,0,0.04);
}
.tab:hover { border-color: #2A4D99; color: #2A4D99; background: #f0f4ff; }
.tab.active {
  background: #2A4D99; border-color: #2A4D99; color: #fff;
  box-shadow: 0 3px 10px rgba(42,77,153,0.28);
}
.search-result-tip { font-size: 12.5px; color: #9facc5; flex-shrink: 0; }

/* ══ Body ══ */
.body { padding: 20px 44px 52px; display: flex; flex-direction: column; gap: 28px; }

.section { display: flex; flex-direction: column; gap: 14px; }
.section-header { display: flex; align-items: center; gap: 12px; }
.section-name { font-size: 14px; font-weight: 700; color: #1f2937; letter-spacing: -0.01em; flex-shrink: 0; }
.section-rule { flex: 1; height: 1px; background: linear-gradient(90deg, #d1d5e8 0%, transparent 100%); }

/* ══ Card Grid ══ */
.card-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 20px;
}

/* ══ Agent Card ══ */
.agent-card {
  background: #fff;
  border-radius: 16px;
  border: 1px solid #e8ecf3;
  overflow: hidden;
  display: flex; flex-direction: column;
  cursor: pointer;
  box-shadow: 0 1px 3px rgba(0,0,0,0.06), 0 4px 12px rgba(0,0,0,0.04);
  transition: box-shadow 0.2s ease, transform 0.2s ease;
}
.agent-card:not(.is-soon):hover {
  box-shadow: 0 8px 28px rgba(0,0,0,0.12), 0 2px 6px rgba(0,0,0,0.06);
  transform: translateY(-4px);
}
.agent-card.is-soon { cursor: default; opacity: 0.55; }

/* ── 封面图 ── */
.card-img-wrap {
  position: relative;
  height: 160px;
  overflow: hidden;
  flex-shrink: 0;
  /* inset shadow 绘制在容器自身，不受子元素 scale 影响，无黑线 */
  box-shadow: inset 0 -60px 40px -10px #fff;
}
.card-img {
  width: 100%; height: 100%;
  object-fit: cover; display: block;
  transition: transform 0.4s ease;
}
.agent-card:not(.is-soon):hover .card-img { transform: scale(1.05); }

/* 底部渐变过渡 —— 已改用 inset box-shadow，此规则保留为空以防残留引用 */

/* 即将上线 badge */
.badge-soon {
  position: absolute; top: 10px; right: 10px;
  font-size: 10.5px; font-weight: 600; color: rgba(255,255,255,0.85);
  background: rgba(0,0,0,0.38); backdrop-filter: blur(4px);
  padding: 3px 9px; border-radius: 20px;
}

/* ── 内容区 ── */
.card-content {
  padding: 14px 18px 16px;
  display: flex; flex-direction: column; gap: 8px;
  background: #fff;
}
.card-top {
  display: flex; align-items: center; justify-content: space-between; gap: 8px;
}
.cat-pill {
  font-size: 11px; font-weight: 600;
  color: var(--accent, #2A4D99);
  background: color-mix(in srgb, var(--accent, #2A4D99) 10%, white);
  padding: 3px 10px; border-radius: 20px;
  border: 1px solid color-mix(in srgb, var(--accent, #2A4D99) 22%, white);
  flex-shrink: 0;
}
.usage-count {
  display: flex; align-items: center; gap: 4px;
  font-size: 11.5px; color: #94a3b8; flex-shrink: 0;
}
.usage-count svg { width: 13px; height: 13px; }
.card-name {
  font-size: 15px; font-weight: 700; color: #0f172a;
  line-height: 1.3; margin: 0;
}
.card-desc {
  font-size: 12.5px; color: #64748b; line-height: 1.6; margin: 0;
}



/* 底部 */
.card-footer {
  display: flex; align-items: center; justify-content: space-between;
  border-top: 1px solid #f1f5fb; padding-top: 10px; margin-top: auto; gap: 8px;
}
.data-dots { display: flex; gap: 6px; flex-wrap: wrap; flex: 1; min-width: 0; }
.data-dot {
  font-size: 10.5px; color: #94a3b8;
  background: #f8fafc; border: 1px solid #e8ecf3;
  padding: 2px 8px; border-radius: 20px; white-space: nowrap;
}
.enter-btn {
  display: flex; align-items: center; gap: 4px;
  font-size: 12px; font-weight: 600;
  color: var(--accent, #2A4D99); flex-shrink: 0; white-space: nowrap;
  opacity: 0; transition: opacity 0.18s;
}
.enter-btn svg { width: 12px; height: 12px; }
.agent-card:not(.is-soon):hover .enter-btn { opacity: 1; }

/* ══ Empty ══ */
.empty {
  display: flex; flex-direction: column; align-items: center;
  padding: 80px 0; gap: 10px; color: #9facc5; font-size: 14px;
}
.empty svg { width: 36px; height: 36px; color: #d1d5db; }
.empty p { margin: 0; }
</style>
