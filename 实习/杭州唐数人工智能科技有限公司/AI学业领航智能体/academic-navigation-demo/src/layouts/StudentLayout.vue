<script setup lang="ts">
import { computed, onMounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useAlertStore } from '@/stores/alerts'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()
const alertStore = useAlertStore()

const navItems = [
  { name: 'AI 领航助手', icon: 'ChatDotRound', path: '/student/agents', highlight: true },
  { name: '学业总览', icon: 'DataBoard', path: '/student/dashboard' },
  { name: '学业画像', icon: 'Odometer', path: '/student/profile' },
  { name: '预警通知', icon: 'Bell', path: '/student/alerts' },
  { name: '发展引导', icon: 'Compass', path: '/student/development-path' },
  { name: '快捷访问', icon: 'Link', path: '/student/quick-access' },
]

const activeIndex = computed(() => route.path)

onMounted(() => {
  alertStore.refreshUnread()
})

watch(() => route.path, () => {
  alertStore.refreshUnread()
})

function logout() {
  auth.logout()
  router.push('/login')
}
</script>

<template>
  <div class="layout-root">
    <aside class="sidebar">
      <div class="sidebar-brand">
        <img src="/zjnu-logo.png" alt="浙江师范大学校徽" class="sidebar-logo-img" />
        <div>
          <div class="sidebar-logo-title">学业领航</div>
          <div class="sidebar-logo-sub">学生端</div>
        </div>
      </div>

      <div class="sidebar-user">
        <div class="sidebar-avatar">{{ auth.user?.name?.[0] || '学' }}</div>
        <div>
          <div class="sidebar-user-name">{{ auth.user?.name }}</div>
          <div class="sidebar-user-meta">{{ auth.user?.className }}</div>
        </div>
      </div>

      <nav class="sidebar-nav">
        <router-link
          v-for="item in navItems" :key="item.path" :to="item.path"
          class="nav-item" :class="{ active: activeIndex.startsWith(item.path), highlight: item.highlight }"
          :aria-current="activeIndex.startsWith(item.path) ? 'page' : undefined"
        >
          <span class="nav-accent" />
          <el-icon class="nav-icon"><component :is="item.icon" /></el-icon>
          <span>{{ item.name }}</span>
          <span
            v-if="item.path === '/student/alerts' && alertStore.unreadCount > 0"
            class="nav-badge"
          >{{ alertStore.unreadCount }}</span>
        </router-link>
      </nav>

      <div class="sidebar-bottom">
        <button class="sidebar-logout" @click="logout">
          <el-icon><SwitchButton /></el-icon>
          <span>退出登录</span>
        </button>
        <div class="sidebar-brand-footer">
          <img src="/zjnu-logo.png" alt="浙江师范大学校徽" class="sidebar-footer-logo" />
          <span>浙江师范大学</span>
        </div>
      </div>
    </aside>

    <main class="layout-main">
      <router-view />
    </main>
  </div>
</template>

<style scoped>
.layout-root { display: flex; height: 100vh; background: #f6f8fb; }

.sidebar {
  width: 220px;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  background: linear-gradient(180deg, #0e1a32 0%, #101d3e 100%);
  border-right: 1px solid rgba(255,255,255,0.04);
}

.sidebar-brand {
  padding: 20px 18px;
  display: flex;
  align-items: center;
  gap: 12px;
  border-bottom: 1px solid rgba(255,255,255,0.06);
}

.sidebar-logo-img {
  width: 34px;
  height: 34px;
  object-fit: contain;
  flex-shrink: 0;
  max-width: 100%;
  border-radius: 50%;
  background: rgba(255,255,255,0.06);
  padding: 2px;
}

.sidebar-logo-title { color: rgba(255,255,255,0.92); font-size: 14px; font-weight: 700; letter-spacing: 0.04em; }
.sidebar-logo-sub { color: rgba(255,255,255,0.3); font-size: 11px; margin-top: 1px; }

.sidebar-user {
  padding: 16px 18px;
  display: flex; align-items: center; gap: 10px;
  border-bottom: 1px solid rgba(255,255,255,0.06);
}

.sidebar-avatar {
  width: 32px; height: 32px; border-radius: 50%;
  background: linear-gradient(135deg, #2A4D99, #5580d4);
  display: flex; align-items: center; justify-content: center;
  color: white; font-size: 13px; font-weight: 700; flex-shrink: 0;
  box-shadow: 0 0 0 2px rgba(85,128,212,0.3);
}

.sidebar-user-name { color: rgba(255,255,255,0.88); font-size: 13px; font-weight: 600; }
.sidebar-user-meta { color: rgba(255,255,255,0.3); font-size: 11px; margin-top: 1px; }

.sidebar-nav { flex: 1; padding: 12px 10px; display: flex; flex-direction: column; gap: 2px; }

.nav-item {
  position: relative;
  display: flex; align-items: center; gap: 10px;
  padding: 10px 14px; border-radius: 10px;
  font-size: 13.5px; font-weight: 500;
  color: rgba(165,185,220,0.65);
  text-decoration: none;
  transition: background 0.15s ease, color 0.15s ease;
}
.nav-item:hover { background: rgba(255,255,255,0.06); color: rgba(210,225,245,0.9); }
.nav-item.active { background: rgba(42,77,153,0.25); color: #ffffff; }

.nav-accent {
  position: absolute; left: 0; top: 50%; transform: translateY(-50%);
  width: 3px; height: 0;
  border-radius: 0 3px 3px 0; background: #5580d4;
  transition: height 0.2s cubic-bezier(0.34,1.56,0.64,1);
}
.nav-item.active .nav-accent { height: 20px; }

.nav-icon { font-size: 15px; flex-shrink: 0; }
.nav-item.active .nav-icon { color: #7f9fdf; }

.nav-badge {
  margin-left: auto;
  min-width: 18px;
  height: 18px;
  padding: 0 5px;
  border-radius: 9px;
  background: #ef4444;
  color: #fff;
  font-size: 11px;
  font-weight: 700;
  line-height: 18px;
  text-align: center;
}

.nav-item.highlight {
  background: rgba(34,197,94,0.12);
  color: #4ade80;
  border: 1px solid rgba(34,197,94,0.25);
}
.nav-item.highlight .nav-icon { color: #4ade80; }
.nav-item.highlight:hover { background: rgba(34,197,94,0.2); }
.nav-item.highlight.active { background: rgba(34,197,94,0.22); border-color: rgba(34,197,94,0.45); color: #86efac; }

.sidebar-bottom { padding: 12px 10px; border-top: 1px solid rgba(255,255,255,0.06); }

.sidebar-brand-footer {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 14px 2px;
  font-size: 11px;
  color: rgba(255,255,255,0.2);
  font-weight: 500;
  letter-spacing: 0.03em;
}
.sidebar-footer-logo {
  width: 14px;
  height: 14px;
  object-fit: contain;
  opacity: 0.35;
  max-width: 100%;
}

.sidebar-logout {
  width: 100%;
  display: flex; align-items: center; gap: 10px;
  padding: 10px 14px; border: none; border-radius: 10px;
  background: transparent;
  color: rgba(165,185,220,0.45);
  font-size: 13px; font-weight: 500;
  cursor: pointer; transition: all 0.15s ease; font-family: inherit;
}
.sidebar-logout:hover { background: rgba(255,255,255,0.06); color: rgba(210,225,245,0.7); }

.layout-main { flex: 1; overflow-y: scroll; }

@media (max-width: 760px) {
  .layout-root { height: auto; min-height: 100vh; display: block; }
  .sidebar { width: 100%; min-height: auto; }
  .sidebar-brand { padding: 14px 18px; }
  .sidebar-user { display: none; }
  .sidebar-nav {
    flex-direction: row;
    overflow-x: auto;
    padding: 8px 12px 10px;
    gap: 6px;
    scrollbar-width: none;
  }
  .sidebar-nav::-webkit-scrollbar { display: none; }
  .nav-item { flex: 0 0 auto; padding: 9px 12px; font-size: 12px; white-space: nowrap; }
  .nav-item .nav-accent { display: none; }
  .nav-item.highlight { order: -1; }
  .sidebar-bottom { display: none; }
  .layout-main { overflow: visible; }
}
</style>
