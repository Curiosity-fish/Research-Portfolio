<script setup lang="ts">
defineProps<{
  title?: string
  label?: string
  value: string | number
  unit?: string
  sub?: string
  trend?: 'up' | 'down' | 'flat'
  trendValue?: string
  color?: 'blue' | 'green' | 'orange' | 'red'
  icon?: string
}>()
</script>

<template>
  <div class="kpi" :class="`kpi-${color || 'blue'}`">
    <div class="kpi-top">
      <span class="kpi-title">{{ title || label }}</span>
      <div v-if="icon" class="kpi-icon-wrap">
        <el-icon><component :is="icon" /></el-icon>
      </div>
    </div>
    <div class="kpi-value-row">
      <span class="kpi-value">{{ value }}</span>
      <span v-if="unit" class="kpi-unit">{{ unit }}</span>
    </div>
    <div v-if="trendValue || sub" class="kpi-footer">
      <el-icon v-if="trend === 'up'" class="trend-up"><ArrowUp /></el-icon>
      <el-icon v-else-if="trend === 'down'" class="trend-down"><ArrowDown /></el-icon>
      <span class="kpi-sub">{{ trendValue || sub }}</span>
    </div>
  </div>
</template>

<style scoped>
.kpi {
  --accent: #2A4D99;
  --accent-bg: rgba(42, 77, 153, 0.07);
  position: relative;
  background: #ffffff;
  border-radius: 16px;
  padding: 20px 22px 18px;
  border: 1px solid rgba(10, 19, 41, 0.07);
  box-shadow: 0 1px 2px rgba(10, 19, 41, 0.03), 0 4px 16px rgba(10, 19, 41, 0.05);
  display: flex;
  flex-direction: column;
  gap: 8px;
  overflow: hidden;
  transition: box-shadow 0.18s ease, transform 0.18s ease;
}

.kpi::before {
  content: '';
  position: absolute;
  left: 0; top: 0; bottom: 0;
  width: 3px;
  background: var(--accent);
  border-radius: 16px 0 0 16px;
}


.kpi:hover {
  box-shadow: 0 4px 8px rgba(10, 19, 41, 0.06), 0 16px 40px rgba(10, 19, 41, 0.08);
  transform: translateY(-2px);
}

.kpi-blue   { --accent: #2A4D99; --accent-bg: rgba(42,77,153,0.08); }
.kpi-green  { --accent: #3db87e; --accent-bg: rgba(61,184,126,0.08); }
.kpi-orange { --accent: #f09a4e; --accent-bg: rgba(240,154,78,0.08); }
.kpi-red    { --accent: #e04538; --accent-bg: rgba(224,69,56,0.08); }

.kpi-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.kpi-title {
  font-size: 12px;
  font-weight: 600;
  color: #8a99b4;
  letter-spacing: 0.04em;
  white-space: nowrap;
}

.kpi-icon-wrap {
  width: 34px; height: 34px;
  border-radius: 9px;
  display: flex; align-items: center; justify-content: center;
  flex-shrink: 0;
  font-size: 18px;
  background: var(--accent-bg);
  color: var(--accent);
}

.kpi-value-row {
  display: flex;
  align-items: baseline;
  gap: 4px;
}

.kpi-value {
  font-size: 30px;
  font-weight: 800;
  line-height: 1;
  color: #111827;
  font-variant-numeric: tabular-nums;
  letter-spacing: -0.02em;
}

.kpi-unit {
  font-size: 13px;
  color: #b4bed2;
  font-weight: 500;
  margin-bottom: 2px;
}

.kpi-footer {
  display: flex;
  align-items: center;
  gap: 4px;
  margin-top: 2px;
}

.kpi-sub { font-size: 11.5px; color: #9facc5; }
.trend-up { color: #3db87e; font-size: 11px; }
.trend-down { color: #e04538; font-size: 11px; }
</style>
