<script setup lang="ts">
/* 页头：标题 + 右侧操作区，所有列表页统一高度，保证高度计算有确定值 */
defineProps<{
  title: string
  description?: string
}>()
</script>

<template>
  <header class="page-header">
    <div class="page-header__text">
      <h1 class="page-header__title">{{ title }}</h1>
      <p v-if="description" class="page-header__desc">{{ description }}</p>
    </div>
    <div class="page-header__actions">
      <slot name="actions" />
    </div>
  </header>
</template>

<style scoped>
.page-header {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  min-height: 40px;
}
.page-header__title {
  margin: 0;
  font-size: 17px;
  font-weight: 700;
  line-height: 24px;
  color: var(--text-primary);
}
.page-header__desc {
  margin: 2px 0 0;
  font-size: 12px;
  color: var(--text-muted);
}
.page-header__actions {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}
</style>
