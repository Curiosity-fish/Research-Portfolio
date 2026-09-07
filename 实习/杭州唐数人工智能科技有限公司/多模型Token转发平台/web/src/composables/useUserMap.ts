/* 用户 ID → 用户名映射：订单/流水只携带 user_id，展示时解析为用户名。
   一次性拉取前 100 条（mock 数据量足够；超出时回退展示短 ID） */
import { onMounted, ref } from 'vue'
import { listUsers } from '@/api/admin/user'

export function shortId(id: string | null | undefined): string {
  if (!id) return '-'
  return `${id.slice(0, 8)}…`
}

export function useUserMap() {
  const userMap = ref<Map<string, string>>(new Map())

  onMounted(async () => {
    const data = await listUsers({ page: 1, page_size: 100 })
    userMap.value = new Map(data.list.map((u) => [u.id, u.username]))
  })

  function userName(id: string): string {
    return userMap.value.get(id) ?? shortId(id)
  }

  return { userMap, userName }
}
