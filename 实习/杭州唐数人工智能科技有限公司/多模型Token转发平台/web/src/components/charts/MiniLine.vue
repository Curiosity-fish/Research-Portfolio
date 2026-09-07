<script setup lang="ts">
/* 轻量折线图（纯 SVG，零依赖）：按 viewBox 等比缩放，标签稀疏渲染避免拥挤。
   用于 Dashboard 近 7 日趋势等小数据量场景，替代引入 echarts 这类重型依赖 */
import { computed } from 'vue'

const props = defineProps<{
  points: { label: string; value: number }[]
  /** 折线/面积主题色，默认品牌金 */
  color?: string
}>()

const W = 560
const H = 150
const PAD_L = 36
const PAD_R = 8
const PAD_T = 10
const PAD_B = 20

const color = computed(() => props.color ?? 'var(--brand)')

const max = computed(() => Math.max(1, ...props.points.map((p) => p.value)))

function x(i: number): number {
  const n = Math.max(1, props.points.length - 1)
  return PAD_L + ((W - PAD_L - PAD_R) * i) / n
}

function y(v: number): number {
  return PAD_T + (H - PAD_T - PAD_B) * (1 - v / max.value)
}

const linePath = computed(() =>
  props.points.map((p, i) => `${i === 0 ? 'M' : 'L'}${x(i).toFixed(1)},${y(p.value).toFixed(1)}`).join(' '),
)

const areaPath = computed(() => {
  if (!props.points.length) return ''
  const base = H - PAD_B
  return `${linePath.value} L${x(props.points.length - 1).toFixed(1)},${base} L${x(0).toFixed(1)},${base} Z`
})

/** 稀疏刻度：最多 5 个 x 标签、3 条 y 网格线 */
const xTicks = computed(() => {
  const n = props.points.length
  if (!n) return []
  const step = Math.max(1, Math.ceil((n - 1) / 4))
  const ticks = []
  for (let i = 0; i < n; i += step) ticks.push({ i, label: props.points[i].label })
  if ((n - 1) % step !== 0) ticks.push({ i: n - 1, label: props.points[n - 1].label })
  return ticks
})

const yTicks = computed(() => [0, 0.5, 1].map((r) => ({ ratio: r, value: Math.round(max.value * r) })))
</script>

<template>
  <svg class="mini-line" :viewBox="`0 0 ${W} ${H}`" preserveAspectRatio="xMidYMid meet" role="img">
    <line v-for="t in yTicks" :key="t.ratio" :x1="PAD_L" :x2="W - PAD_R"
      :y1="y(max * t.ratio)" :y2="y(max * t.ratio)" class="mini-line__grid" />
    <path v-if="areaPath" :d="areaPath" class="mini-line__area" :style="{ fill: color }" />
    <path v-if="linePath" :d="linePath" class="mini-line__line" :style="{ stroke: color }" />
    <circle v-for="(p, i) in points" :key="i" :cx="x(i)" :cy="y(p.value)" r="2.4" class="mini-line__dot"
      :style="{ fill: color }" />
    <text v-for="t in yTicks" :key="'y' + t.ratio" :x="PAD_L - 6" :y="y(max * t.ratio) + 3" class="mini-line__tick" text-anchor="end">
      {{ t.value >= 10000 ? (t.value / 10000).toFixed(0) + 'w' : t.value }}
    </text>
    <text v-for="t in xTicks" :key="'x' + t.i" :x="x(t.i)" :y="H - 4" class="mini-line__tick" text-anchor="middle">
      {{ t.label }}
    </text>
  </svg>
</template>

<style scoped>
.mini-line {
  width: 100%;
  height: 100%;
  display: block;
}
.mini-line__grid {
  stroke: var(--border-light);
  stroke-width: 1;
}
.mini-line__area {
  opacity: 0.12;
}
.mini-line__line {
  fill: none;
  stroke-width: 2;
  stroke-linejoin: round;
  stroke-linecap: round;
}
.mini-line__dot {
  stroke: var(--bg-card);
  stroke-width: 1;
}
.mini-line__tick {
  font-size: 10px;
  fill: var(--text-muted);
}
</style>
