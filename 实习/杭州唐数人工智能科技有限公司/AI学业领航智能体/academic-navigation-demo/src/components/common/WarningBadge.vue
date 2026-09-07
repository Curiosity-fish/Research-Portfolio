<script setup lang="ts">
defineProps<{
  level: 'yellow' | 'orange' | 'red'
  label?: string
}>()

const config = {
  yellow: {
    bg: '#fffbeb', text: '#92650a', border: '#fde68a',
    dot: '#f59e0b', label: '黄色预警',
  },
  orange: {
    bg: '#fff7ed', text: '#9a3f12', border: '#fed7aa',
    dot: '#f97316', label: '橙色预警',
  },
  red: {
    bg: '#fef2f2', text: '#991b1b', border: '#fecaca',
    dot: '#ef4444', label: '红色预警',
  },
}
</script>

<template>
  <span
    class="warning-badge"
    :style="{
      background: config[level].bg,
      color: config[level].text,
      borderColor: config[level].border,
    }"
  >
    <span class="badge-dot" :style="{ background: config[level].dot }" />
    {{ label || config[level].label }}
  </span>
</template>

<style scoped>
.warning-badge {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 3px 10px;
  border-radius: 6px;
  border: 1px solid;
  font-size: 12px;
  font-weight: 500;
  letter-spacing: 0.01em;
  white-space: nowrap;
}
.badge-dot {
  width: 5px;
  height: 5px;
  border-radius: 50%;
  flex-shrink: 0;
}
</style>
