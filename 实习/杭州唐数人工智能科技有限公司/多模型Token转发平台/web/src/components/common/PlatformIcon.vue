<script setup lang="ts">
/* 平台品牌图标：simple-icons 按需具名引入（包内大量品牌图标因合规被移除，
   仅映射 v16 中确认存在的 slug；未命中映射的 code 一律降级为文字徽标） */
import { computed } from 'vue'
import {
  siAlibabacloud,
  siAnthropic,
  siBaidu,
  siBytedance,
  siDeepseek,
  siGooglegemini,
  siMeta,
  siMistralai,
  siMoonshotai,
  siOllama,
  siOpenrouter,
  siPerplexity,
  siVllm,
  siX,
  type SimpleIcon,
} from 'simple-icons'

const props = defineProps<{
  /** 平台 code（如 openai / anthropic / deepseek），大小写不敏感 */
  code: string
  size?: number
}>()

// 平台 code → 已确认的图标；openai/azure/zhipu 等品牌图标在 simple-icons v16 已移除，刻意不映射
const CODE_TO_ICON: Record<string, SimpleIcon> = {
  anthropic: siAnthropic,
  deepseek: siDeepseek,
  google: siGooglegemini,
  gemini: siGooglegemini,
  mistral: siMistralai,
  meta: siMeta,
  llama: siMeta,
  moonshot: siMoonshotai,
  kimi: siMoonshotai,
  qwen: siAlibabacloud,
  alibaba: siAlibabacloud,
  baidu: siBaidu,
  ernie: siBaidu,
  doubao: siBytedance,
  bytedance: siBytedance,
  volcengine: siBytedance,
  openrouter: siOpenrouter,
  xai: siX,
  perplexity: siPerplexity,
  ollama: siOllama,
  vllm: siVllm,
}

const icon = computed<SimpleIcon | null>(() => CODE_TO_ICON[props.code?.toLowerCase()] ?? null)
const brandColor = computed(() => (icon.value ? `#${icon.value.hex}` : 'var(--brand)'))
const sizePx = computed(() => `${props.size ?? 16}px`)
</script>

<template>
  <svg
    v-if="icon"
    role="img"
    :aria-label="icon.title"
    viewBox="0 0 24 24"
    :width="sizePx"
    :height="sizePx"
    :fill="brandColor"
  >
    <path :d="icon.path" />
  </svg>
  <span v-else class="platform-icon__fallback" :style="{ width: sizePx, height: sizePx }">
    {{ code?.slice(0, 1).toUpperCase() ?? '?' }}
  </span>
</template>

<style scoped>
.platform-icon__fallback {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 6px;
  background: var(--brand-bg);
  color: var(--brand);
  font-size: 10px;
  font-weight: 700;
}
</style>
