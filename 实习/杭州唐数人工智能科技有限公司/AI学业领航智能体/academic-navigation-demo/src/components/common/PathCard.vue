<script setup lang="ts">
defineProps<{
  pathKey: 'graduate' | 'overseas' | 'employment' | 'civil' | 'institution'
  label: string
  active: boolean
}>()

const emit = defineEmits<{ select: [] }>()

const iconMap = {
  graduate:    '🎓',
  overseas:    '✈️',
  employment:  '💼',
  civil:       '🏛️',
  institution: '📌',
}

const colorMap = {
  graduate:    { accent: '#2A4D99', tint: '#eef3fb' },
  overseas:    { accent: '#0ea5e9', tint: '#f0f9ff' },
  employment:  { accent: '#059669', tint: '#ecfdf5' },
  civil:       { accent: '#b45309', tint: '#fef3e2' },
  institution: { accent: '#0f766e', tint: '#f0fdfa' },
}
</script>

<template>
  <div
    class="path-card"
    :class="{ active }"
    :style="active ? { borderColor: colorMap[pathKey].accent, background: colorMap[pathKey].tint } : {}"
    @click="emit('select')"
  >
    <div class="path-icon">{{ iconMap[pathKey] }}</div>
    <div class="path-label">{{ label }}</div>
    <div v-if="active" class="path-active-tag" :style="{ background: colorMap[pathKey].accent }">当前选择</div>
  </div>
</template>

<style scoped>
.path-card {
  position: relative;
  display: flex; flex-direction: column; align-items: center;
  padding: 24px 16px 20px;
  border-radius: 14px;
  border: 2px solid #edf0f6;
  background: #fff;
  cursor: pointer;
  transition: all 0.2s ease;
  text-align: center;
  user-select: none;
}
.path-card:hover:not(.active) {
  border-color: #c8d4e8;
  transform: translateY(-2px);
  box-shadow: 0 4px 16px rgba(0,0,0,0.07);
}
.path-card.active {
  border-width: 2px;
  box-shadow: 0 4px 20px rgba(42,77,153,0.12);
}
.path-icon { font-size: 32px; margin-bottom: 10px; line-height: 1; }
.path-label { font-size: 15px; font-weight: 700; color: #1f2937; }
.path-active-tag {
  position: absolute; top: -1px; right: 12px;
  font-size: 11px; font-weight: 600; color: #fff;
  padding: 2px 8px; border-radius: 0 0 8px 8px;
}
</style>
