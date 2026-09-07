import { ref } from 'vue'
import { defineStore } from 'pinia'
import { getUnreadCount } from '@/api'

export const useAlertStore = defineStore('alerts', () => {
  const unreadCount = ref(0)

  async function refreshUnread() {
    try {
      unreadCount.value = await getUnreadCount()
    } catch {
      unreadCount.value = 0
    }
  }

  return { unreadCount, refreshUnread }
})
