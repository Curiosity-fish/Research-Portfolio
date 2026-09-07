<script setup lang="ts">
/* 密度填充统计卡栅格：auto-fill minmax 让卡片列数随内容区宽度自动增减，
   1366 下一行约 5-6 张，1920 下 7-8 张，大屏不再出现两侧大片空白 */
defineProps<{
  minCardWidth?: number
}>()
</script>

<template>
  <div class="stat-grid" :style="{ '--stat-min-width': (minCardWidth ?? 200) + 'px' }">
    <slot />
  </div>
</template>

<style scoped>
/* auto-fit 而非 auto-fill：空轨道坍缩并均分给现有卡片，
   固定数量的统计卡在任何宽度下都铺满整行，不产生右侧大片空白 */
.stat-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(min(100%, var(--stat-min-width)), 1fr));
  gap: 12px;
}
</style>
