<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, watch, nextTick } from 'vue'
import * as echarts from 'echarts'

const props = defineProps<{
  dimensions: { label: string; score: number; avgScore: number }[]
}>()

const chartRef = ref<HTMLDivElement>()
let chart: echarts.ECharts | null = null

function buildOption() {
  const labels = props.dimensions.map(d => d.label)
  const myScores = props.dimensions.map(d => d.score)
  const avgScores = props.dimensions.map(d => d.avgScore)

  return {
    tooltip: {
      trigger: 'item',
      backgroundColor: '#ffffff',
      borderColor: '#eceef2',
      borderWidth: 1,
      textStyle: { color: '#1f2937', fontSize: 12 },
      extraCssText: 'border-radius:8px;box-shadow:0 4px 14px rgba(17,31,56,0.1);padding:8px 12px',
    },
    legend: {
      bottom: 0,
      data: ['我的得分', '年级平均'],
      textStyle: { color: '#717888', fontSize: 11.5 },
      icon: 'circle',
      itemWidth: 7, itemHeight: 7,
      itemGap: 20,
    },
    radar: {
      indicator: labels.map(l => ({ name: l, max: 100 })),
      shape: 'polygon',
      splitNumber: 4,
      radius: '62%',
      center: ['50%', '48%'],
      axisName: {
        color: '#717888',
        fontSize: 11.5,
        fontWeight: '500',
      },
      splitLine: { lineStyle: { color: '#eceef2', width: 1 } },
      splitArea: {
        areaStyle: {
          color: ['rgba(58,95,160,0.02)', 'rgba(58,95,160,0.04)'],
        },
      },
      axisLine: { lineStyle: { color: '#eceef2', width: 1 } },
    },
    series: [{
      type: 'radar',
      data: [
        {
          value: myScores,
          name: '我的得分',
          symbol: 'circle',
          symbolSize: 5,
          lineStyle: { color: '#3a5fa0', width: 2 },
          areaStyle: { color: 'rgba(58,95,160,0.12)' },
          itemStyle: { color: '#ffffff', borderColor: '#3a5fa0', borderWidth: 2 },
        },
        {
          value: avgScores,
          name: '年级平均',
          symbol: 'none',
          lineStyle: { color: '#d9dce4', width: 1.5, type: [5, 4] },
          areaStyle: { color: 'rgba(217,220,228,0.08)' },
          itemStyle: { color: '#d9dce4' },
        },
      ],
    }],
  }
}

function renderChart() {
  if (!chartRef.value || props.dimensions.length === 0) return
  if (!chart) chart = echarts.init(chartRef.value, undefined, { renderer: 'canvas' })
  chart.setOption(buildOption(), true)
}

onMounted(() => {
  renderChart()
})
watch(() => props.dimensions, async () => {
  await nextTick()
  renderChart()
  chart?.resize()
}, { deep: true, immediate: true })
onBeforeUnmount(() => { chart?.dispose() })
</script>

<template>
  <div ref="chartRef" class="w-full h-72" />
</template>
