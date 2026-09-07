/* 监听元素高度：任何布局变化（窗口缩放、筛选区展开收起、侧边栏折叠）都实时反映 */
import { onBeforeUnmount, onMounted, ref, type Ref } from 'vue'

export function useAvailableHeight(target: Ref<HTMLElement | undefined>): Ref<number> {
  const height = ref(0)
  let observer: ResizeObserver | null = null

  onMounted(() => {
    if (!target.value) return
    height.value = target.value.clientHeight
    observer = new ResizeObserver((entries) => {
      height.value = entries[0].contentRect.height
    })
    observer.observe(target.value)
  })

  onBeforeUnmount(() => observer?.disconnect())
  return height
}
