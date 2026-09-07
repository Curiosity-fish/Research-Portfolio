<script setup lang="ts">
import type { ChatMessage } from '@/types'

defineProps<{ message: ChatMessage }>()

function renderContent(text: string) {
  return text
    .replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
    .replace(/\n/g, '<br/>')
}
</script>

<template>
  <div class="flex gap-3" :class="message.role === 'user' ? 'flex-row-reverse' : 'flex-row'">
    <!-- Avatar -->
    <div
      class="w-9 h-9 rounded-full flex-shrink-0 flex items-center justify-center text-white text-sm font-bold"
      :class="message.role === 'ai' ? 'bg-zjnu-blue-500' : 'bg-gray-300'"
    >
      <span v-if="message.role === 'ai'">AI</span>
      <el-icon v-else><UserFilled /></el-icon>
    </div>

    <!-- Bubble -->
    <div class="max-w-[75%] flex flex-col gap-2">
      <div
        class="px-4 py-3 rounded-2xl text-sm leading-relaxed"
        :class="message.role === 'user'
          ? 'bg-zjnu-blue-500 text-white rounded-tr-sm'
          : 'bg-white text-gray-700 rounded-tl-sm shadow-sm border border-gray-100'"
        v-html="renderContent(message.content)"
      />

      <!-- Score card -->
      <div v-if="message.cardData?.type === 'score'" class="bg-white rounded-xl p-4 shadow-sm border border-gray-100">
        <div class="text-xs text-gray-400 mb-2">学业数据卡片</div>
        <div class="flex gap-6">
          <div class="text-center">
            <div class="text-2xl font-bold text-zjnu-blue-500">3.62</div>
            <div class="text-xs text-gray-400 mt-1">累计GPA</div>
          </div>
          <div class="text-center">
            <div class="text-2xl font-bold text-growth-green-500">15</div>
            <div class="text-xs text-gray-400 mt-1">专业排名</div>
          </div>
          <div class="text-center">
            <div class="text-2xl font-bold text-energy-orange-500">12.5%</div>
            <div class="text-xs text-gray-400 mt-1">专业百分位</div>
          </div>
        </div>
      </div>

      <!-- Alert card -->
      <div v-if="message.cardData?.type === 'alert'" class="bg-green-50 rounded-xl p-4 border border-green-100">
        <div class="flex items-center gap-2 text-growth-green-600">
          <el-icon><CircleCheckFilled /></el-icon>
          <span class="text-sm font-medium">学业状态良好，无预警</span>
        </div>
      </div>

      <div class="text-xs text-gray-300 mt-1" :class="message.role === 'user' ? 'text-right' : 'text-left'">
        {{ new Date(message.timestamp).toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' }) }}
      </div>
    </div>
  </div>
</template>
