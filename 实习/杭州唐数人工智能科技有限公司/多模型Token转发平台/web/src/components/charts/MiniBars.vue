<script setup lang="ts">
/* 轻量横向条形图（纯 SVG，零依赖）：Top-N 排行场景（模型/用户用量分布） */
import { computed } from 'vue'

const props = defineProps<{
  items: { label: string; value: number }[]
  color?: string
}>()

const color = computed(() => props.color ?? 'var(--brand)')
const max = computed(() => Math.max(1, ...props.items.map((i) => i.value)))

function barWidth(v: number): number {
  return Math.max(2, Math.round((v / max.value) * 100))
}

function fmt(v: number): string {
  return v >= 10000 ? `${(v / 10000).toFixed(1)}w` : v.toLocaleString('zh-CN')
}
</script>

<template>
  <div class="mini-bars">
    <div v-if="!items.length" class="mini-bars__empty">暂无数据</div>
    <div v-for="item in items" :key="item.label" class="mini-bars__row">
      <span class="mini-bars__label" :title="item.label">{{ item.label }}</span>
      <div class="mini-bars__track">
        <div class="mini-bars__bar" :style="{ width: barWidth(item.value) + '%', background: color }" />
      </div>
      <span class="mini-bars__value tnum">{{ fmt(item.value) }}</span>
    </div>
  </div>
</template>

<style scoped>
.mini-bars {
  display: flex;
  flex-direction: column;
  gap: 8px;
  overflow: hidden auto;
}
.mini-bars__empty {
  color: var(--text-muted);
  font-size: 12px;
  text-align: center;
  padding: 24px 0;
}
.mini-bars__row {
  display: flex;
  align-items: center;
  gap: 8px;
}
.mini-bars__label {
  width: 88px;
  flex-shrink: 0;
  font-size: 12px;
  color: var(--text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.mini-bars__track {
  flex: 1;
  height: 12px;
  background: var(--bg-page);
  border-radius: 6px;
  overflow: hidden;
}
.mini-bars__bar {
  height: 100%;
  border-radius: 6px;
  opacity: 0.85;
}
.mini-bars__value {
  width: 56px;
  flex-shrink: 0;
  text-align: right;
  font-size: 12px;
  color: var(--text-muted);
}
</style>
