<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'

const props = withDefaults(defineProps<{
  title?: string
  summary?: string
  points?: string[]
  actions?: { label: string; color?: string }[]
  accent?: string
  loading?: boolean
  compact?: boolean
}>(), {
  title: 'AI 分析',
  accent: '#2A4D99',
  loading: false,
  compact: false,
})

const visibleCount = ref(0)
const appeared = ref(false)

function animate() {
  appeared.value = false
  visibleCount.value = 0
  setTimeout(() => {
    appeared.value = true
    if (props.points?.length) {
      let i = 0
      const t = setInterval(() => {
        visibleCount.value = ++i
        if (i >= props.points!.length) clearInterval(t)
      }, 160)
    }
  }, 200)
}

onMounted(animate)
watch(() => props.points, animate)
</script>

<template>
  <div class="ai-panel" :class="{ compact }">
    <div class="ai-hd" :style="{ borderLeftColor: accent }">
      <div class="ai-icon-wrap" :style="{ background: accent + '18', color: accent }">
        <svg width="13" height="13" viewBox="0 0 24 24" fill="currentColor">
          <path d="M12 2l2.4 7.4H22l-6.2 4.5 2.4 7.4L12 17l-6.2 4.3 2.4-7.4L2 9.4h7.6z"/>
        </svg>
      </div>
      <span class="ai-title-text">{{ title }}</span>
      <span class="ai-badge" :style="{ background: accent }">AI</span>
    </div>

    <div v-if="loading || !appeared" class="ai-skeleton">
      <div class="sk-line" style="width: 85%" />
      <div class="sk-line" style="width: 65%" />
      <div class="sk-line" style="width: 75%" />
    </div>

    <Transition name="ai-fade">
      <div v-if="appeared && !loading" class="ai-body">
        <p v-if="summary" class="ai-summary">{{ summary }}</p>
        <div v-if="points?.length" class="ai-points">
          <div
            v-for="(pt, i) in points" :key="i"
            class="ai-point"
            :class="{ shown: i < visibleCount }"
          >
            <span class="pt-dot" :style="{ background: accent }" />
            <span class="pt-text">{{ pt }}</span>
          </div>
        </div>
        <div v-if="actions?.length" class="ai-actions">
          <span
            v-for="a in actions" :key="a.label"
            class="ai-chip"
            :style="{ background: (a.color || accent) + '12', color: a.color || accent, borderColor: (a.color || accent) + '30' }"
          >{{ a.label }}</span>
        </div>
      </div>
    </Transition>
  </div>
</template>

<style scoped>
.ai-panel {
  background: #fff;
  border-radius: 14px;
  border: 1px solid rgba(10,19,41,0.07);
  box-shadow: 0 1px 2px rgba(10,19,41,0.03), 0 4px 16px rgba(10,19,41,0.05);
  overflow: hidden;
}

.ai-hd {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 14px 16px 12px;
  border-bottom: 1px solid #f2f4f9;
  border-left: 3px solid var(--accent, #2A4D99);
  background: linear-gradient(135deg, #fafbfd 0%, #fff 100%);
}
.ai-icon-wrap {
  width: 24px; height: 24px;
  border-radius: 7px;
  display: flex; align-items: center; justify-content: center;
  flex-shrink: 0;
}
.ai-title-text {
  font-size: 13px; font-weight: 700; color: #1f2937; flex: 1;
}
.ai-badge {
  font-size: 10px; font-weight: 800; color: #fff;
  padding: 2px 7px; border-radius: 8px; letter-spacing: 0.05em;
}

.ai-skeleton {
  padding: 14px 16px;
  display: flex; flex-direction: column; gap: 8px;
}
.sk-line {
  height: 10px; border-radius: 6px;
  background: linear-gradient(90deg, #f0f2f7 25%, #e8eaef 50%, #f0f2f7 75%);
  background-size: 200% 100%;
  animation: shimmer 1.4s infinite;
}
@keyframes shimmer { 0% { background-position: -200% 0 } 100% { background-position: 200% 0 } }

.ai-body { padding: 14px 16px; display: flex; flex-direction: column; gap: 10px; }

.ai-summary {
  font-size: 13px; color: #374151; line-height: 1.65;
  margin: 0; font-weight: 450;
}

.ai-points { display: flex; flex-direction: column; gap: 7px; }
.ai-point {
  display: flex; align-items: flex-start; gap: 8px;
  opacity: 0; transform: translateY(6px);
  transition: opacity 0.3s ease, transform 0.3s ease;
}
.ai-point.shown { opacity: 1; transform: none; }

.pt-dot {
  width: 6px; height: 6px; border-radius: 50%;
  flex-shrink: 0; margin-top: 6px;
}
.pt-text { font-size: 12.5px; color: #4b5563; line-height: 1.6; }

.ai-actions { display: flex; flex-wrap: wrap; gap: 6px; padding-top: 4px; }
.ai-chip {
  font-size: 11.5px; font-weight: 600;
  padding: 4px 10px; border-radius: 20px;
  border: 1px solid; cursor: default;
  transition: opacity 0.15s;
}

.ai-fade-enter-active { transition: opacity 0.35s ease; }
.ai-fade-enter-from { opacity: 0; }

.compact .ai-hd { padding: 10px 14px 9px; }
.compact .ai-body { padding: 10px 14px; gap: 8px; }
.compact .ai-summary { font-size: 12.5px; }
.compact .pt-text { font-size: 12px; }
</style>
