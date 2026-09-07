<script setup lang="ts" generic="T">
/**
 * 自适应表格：把「可用高度 → 每页行数 → 分页」整条链内置。
 * 表格区不设 height/max-height，el-table 不产生内部滚动条；
 * 行数由容器高度计算，装不下的数据交给分页器，保证 B 档滚动策略。
 * 列定义通过默认 slot 透传给 el-table，页面写法与原生 EP 一致。
 */
import { onBeforeUnmount, onMounted, ref, type Ref } from 'vue'
import type { PageData } from '@/api/types'

const props = withDefaults(
  defineProps<{
    fetch: (page: number, pageSize: number) => Promise<PageData<T>>
    rowHeight?: number
    minRows?: number
  }>(),
  { rowHeight: 44, minRows: 5 },
)

// 与 styles/element-theme.css 中的 --table-header-height 保持一致
const HEADER_HEIGHT = 40
// 表格与分页器之间的间距，预留避免取整误差导致溢出
const BOTTOM_GAP = 4

const rows = ref<T[]>([]) as Ref<T[]>
const total = ref(0)
const loading = ref(false)
const page = ref(1)
const pageSize = ref(10)

const tableWrapRef = ref<HTMLElement>()
let observer: ResizeObserver | null = null

function computeRows(height: number): number {
  const available = height - HEADER_HEIGHT - BOTTOM_GAP
  return Math.max(props.minRows, Math.floor(available / props.rowHeight))
}

async function load(p: number): Promise<void> {
  loading.value = true
  try {
    const data = await props.fetch(p, pageSize.value)
    // 部分接口（如 stats/usage）不分页、无 total 字段，按 0 兜底防止分页器失效
    const resolvedTotal = data.total ?? 0
    const maxPage = Math.max(1, Math.ceil(resolvedTotal / pageSize.value))
    if (p > maxPage && resolvedTotal > 0) {
      // 当前页在新条件下越界（筛选或行数变化后 total 缩小），回退到最后一页
      page.value = maxPage
      const again = await props.fetch(maxPage, pageSize.value)
      rows.value = again.list
      total.value = again.total ?? 0
    } else {
      page.value = p
      rows.value = data.list
      total.value = resolvedTotal
    }
  } finally {
    loading.value = false
  }
}

/** 页面操作（编辑/删除/筛选变更）后刷新；传 1 表示回到首页 */
function reload(p?: number): void {
  void load(p ?? page.value)
}

onMounted(() => {
  const el = tableWrapRef.value
  if (!el) return
  // 首屏同步读一次，避免默认行数闪变；后续交给 ResizeObserver
  pageSize.value = computeRows(el.clientHeight)
  observer = new ResizeObserver((entries) => {
    const next = computeRows(entries[0].contentRect.height)
    if (next !== pageSize.value) {
      pageSize.value = next
      void load(1)
    }
  })
  observer.observe(el)
  void load(1)
})

onBeforeUnmount(() => observer?.disconnect())

defineExpose({ reload })
</script>

<template>
  <div class="adaptive-table">
    <div ref="tableWrapRef" class="adaptive-table__wrap">
      <el-table v-loading="loading" :data="rows" class="adaptive-table__grid" stripe>
        <slot />
        <template #empty>
          <el-empty description="暂无数据" :image-size="72" />
        </template>
      </el-table>
    </div>
    <div class="adaptive-table__pager">
      <el-pagination
        :total="total"
        layout="total, prev, pager, next"
        :page-size="pageSize"
        @current-change="reload"
      />
    </div>
  </div>
</template>

<style scoped>
.adaptive-table {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  gap: 8px;
}
/* 表格区吃掉除分页器外的全部高度；overflow 兜底取整误差，保证永不出现内部滚动条 */
.adaptive-table__wrap {
  flex: 1;
  min-height: 0;
  overflow: hidden;
}
.adaptive-table__grid {
  width: 100%;
  /* 高度撑满 wrapper：行数由容器高度计算并留有余量，内容不会超高，
     EP 只在内容超出时才出现 body 滚动条，因此这里仍是零滚动条，
     同时让空态占位垂直居中、表头固定 */
  height: 100%;
}
.adaptive-table__pager {
  flex-shrink: 0;
  display: flex;
  justify-content: flex-end;
}
</style>
