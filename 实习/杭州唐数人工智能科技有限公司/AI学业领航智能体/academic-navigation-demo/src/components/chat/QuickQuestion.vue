<script setup lang="ts">
defineProps<{
  question: string
}>()
const emit = defineEmits<{ click: [q: string] }>()
</script>

<template>
  <button
    class="px-3 py-1.5 rounded-full text-sm border border-zjnu-blue-200 text-zjnu-blue-600 bg-zjnu-blue-50 hover:bg-zjnu-blue-100 transition-colors whitespace-nowrap"
    @click="emit('click', question)"
  >
    {{ question }}
  </button>
</template>
