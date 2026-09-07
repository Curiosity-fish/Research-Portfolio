<script setup lang="ts">
import type { MilestoneItem } from '@/types'

defineProps<{ milestones: MilestoneItem[] }>()
</script>

<template>
  <div class="timeline">
    <div v-for="(m, i) in milestones" :key="i" class="timeline-item" :class="m.status">
      <div class="timeline-line-wrap">
        <div class="timeline-dot" :class="m.status" />
        <div v-if="i < milestones.length - 1" class="timeline-connector" :class="m.status === 'done' ? 'done' : 'pending'" />
      </div>
      <div class="timeline-content">
        <div class="timeline-header">
          <span class="timeline-title">{{ m.title }}</span>
          <span class="timeline-date">{{ m.date }}</span>
        </div>
        <p v-if="m.description" class="timeline-desc">{{ m.description }}</p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.timeline { display: flex; flex-direction: column; gap: 0; }

.timeline-item {
  display: flex;
  gap: 14px;
  min-height: 52px;
}

.timeline-line-wrap {
  display: flex; flex-direction: column; align-items: center;
  flex-shrink: 0; width: 20px;
}

.timeline-dot {
  width: 14px; height: 14px; border-radius: 50%; flex-shrink: 0;
  margin-top: 3px;
}
.timeline-dot.done { background: #3db87e; }
.timeline-dot.current {
  background: #2A4D99;
  box-shadow: 0 0 0 4px rgba(42,77,153,0.18);
  animation: pulse-glow 2s infinite;
}
.timeline-dot.pending { background: #e5e7eb; border: 2px solid #d1d5db; }

.timeline-connector {
  flex: 1; width: 2px; margin: 4px 0;
}
.timeline-connector.done { background: #3db87e; }
.timeline-connector.pending { background: #e5e7eb; }

.timeline-content { padding-bottom: 16px; flex: 1; }

.timeline-header {
  display: flex; align-items: baseline; justify-content: space-between; gap: 8px;
  margin-bottom: 4px;
}

.timeline-title {
  font-size: 13.5px; font-weight: 600; color: #1f2937;
}
.timeline-item.pending .timeline-title { color: #9ca3af; }

.timeline-date {
  font-size: 11.5px; color: #9ca3af; flex-shrink: 0;
}

.timeline-desc {
  font-size: 12.5px; color: #6b7280; margin: 0; line-height: 1.5;
}
</style>
