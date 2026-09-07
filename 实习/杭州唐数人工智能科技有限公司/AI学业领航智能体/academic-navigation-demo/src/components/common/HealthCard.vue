<script setup lang="ts">
defineProps<{
  label: string
  score: number
  status: 'good' | 'normal' | 'warning' | 'danger'
  change?: number
  avgScore?: number
}>()

const cfg = {
  good:    { bar: '#35935a', tint: '#eef6f1', textColor: '#287350', label: '良好' },
  normal:  { bar: '#3a5fa0', tint: '#eef1f7', textColor: '#2f4f87', label: '正常' },
  warning: { bar: '#e89520', tint: '#fef6e8', textColor: '#cc7f18', label: '注意' },
  danger:  { bar: '#d94535', tint: '#fef0ee', textColor: '#bf3b2d', label: '预警' },
}
</script>

<template>
  <div class="health-card">
    <div class="hc-header">
      <span class="hc-label">{{ label }}</span>
      <span class="hc-status" :style="{ background: cfg[status].tint, color: cfg[status].textColor }">
        {{ cfg[status].label }}
      </span>
    </div>
    <div class="hc-score-row">
      <span class="hc-score">{{ score }}</span>
      <span class="hc-max">/ 100</span>
      <span v-if="change !== undefined" class="hc-change" :style="{ color: change >= 0 ? '#35935a' : '#d94535' }">
        {{ change >= 0 ? '+' : '' }}{{ change }}
      </span>
    </div>
    <div class="hc-bar-wrap">
      <div class="hc-bar" :style="{ width: score + '%', background: cfg[status].bar }" />
    </div>
    <div v-if="avgScore !== undefined" class="hc-avg">
      <span>均值 {{ avgScore }}</span>
    </div>
  </div>
</template>

<style scoped>
.health-card {
  background: #ffffff;
  border-radius: 12px;
  border: 1px solid #eceef2;
  box-shadow: 0 1px 3px rgba(17,31,56,0.05);
  padding: 16px;
  transition: box-shadow 0.2s ease, transform 0.2s ease;
}
.health-card:hover {
  box-shadow: 0 4px 12px rgba(17,31,56,0.09);
  transform: translateY(-1px);
}
.hc-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 10px;
}
.hc-label {
  font-size: 12.5px;
  font-weight: 600;
  color: #565d6d;
}
.hc-status {
  font-size: 11px;
  font-weight: 600;
  padding: 2px 8px;
  border-radius: 5px;
  letter-spacing: 0.02em;
}
.hc-score-row {
  display: flex;
  align-items: baseline;
  gap: 4px;
  margin-bottom: 8px;
}
.hc-score {
  font-size: 24px;
  font-weight: 700;
  color: #1a2845;
  letter-spacing: -0.02em;
}
.hc-max {
  font-size: 11.5px;
  color: #b3b8c4;
  font-weight: 500;
}
.hc-change {
  font-size: 11.5px;
  font-weight: 600;
  margin-left: auto;
}
.hc-bar-wrap {
  height: 4px;
  background: #f0f2f5;
  border-radius: 3px;
  overflow: hidden;
}
.hc-bar {
  height: 100%;
  border-radius: 3px;
  transition: width 0.7s cubic-bezier(0.4, 0, 0.2, 1);
}
.hc-avg {
  font-size: 11px;
  color: #b3b8c4;
  text-align: right;
  margin-top: 4px;
}
</style>
