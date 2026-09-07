<script setup lang="ts">
defineProps<{
  title?: string
  description?: string
  icon?: string
}>()
</script>

<template>
  <div class="flex flex-col items-center justify-center py-16 text-gray-400">
    <el-icon class="text-5xl mb-4 text-gray-200"><component :is="icon || 'DocumentRemove'" /></el-icon>
    <p class="text-base font-medium text-gray-400">{{ title || '暂无数据' }}</p>
    <p v-if="description" class="text-sm mt-1">{{ description }}</p>
  </div>
</template>
