<script setup lang="ts">
/* 管理端骨架：侧边栏 + 内容区。
   根节点 100vh + overflow hidden，从结构上保证整页永不滚动；
   内容区高度固定，内部页面按「页头 + 自适应表格」配方消化高度。 */
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const collapsed = ref<boolean>(localStorage.getItem('admin_sidebar_collapsed') === '1')
const drawerOpen = ref(false)

function toggleCollapsed(): void {
  collapsed.value = !collapsed.value
  localStorage.setItem('admin_sidebar_collapsed', collapsed.value ? '1' : '0')
}

interface NavItem {
  to: string
  label: string
  icon: string
}

interface NavGroup {
  label: string
  items: NavItem[]
}

const navGroups = computed<NavGroup[]>(() => [
  {
    label: '数据',
    items: [{ to: '/admin', label: '概览', icon: 'Odometer' }],
  },
  {
    label: '管理',
    items: [
      { to: '/admin/users', label: '用户管理', icon: 'User' },
      { to: '/admin/relay', label: '转发资源', icon: 'Connection' },
    ],
  },
  {
    label: '运营',
    items: [
      { to: '/admin/finance', label: '资金与计费', icon: 'Wallet' },
      { to: '/admin/notifications', label: '通知', icon: 'Bell' },
      { to: '/admin/alerts', label: '告警', icon: 'Warning' },
      { to: '/admin/audit-logs', label: '审计日志', icon: 'Document' },
    ],
  },
])

function isActive(item: NavItem): boolean {
  if (item.to === '/admin') return route.path === '/admin'
  return route.path.startsWith(item.to)
}

function onNavigate(): void {
  drawerOpen.value = false
}

function logout(): void {
  auth.logout()
  router.push('/admin/login')
}

const roleLabel = computed(() =>
  auth.me?.role === 'super_admin' ? '超级管理员' : '管理员',
)

// 刷新页面后凭证仍在但 me 未加载：恢复管理员信息（token 失效时由拦截器跳登录）
onMounted(() => {
  void auth.ensureMe().catch(() => {})
})
</script>

<template>
  <div class="admin-layout">
    <!-- 移动端遮罩（<1024px，仅抽屉打开时渲染） -->
    <div v-if="drawerOpen" class="admin-layout__mask" @click="drawerOpen = false" />

    <!-- 侧边栏 -->
    <aside
      class="admin-layout__sidebar"
      :class="[
        collapsed ? 'admin-layout__sidebar--collapsed' : '',
        drawerOpen ? 'admin-layout__sidebar--drawer-open' : '',
      ]"
    >
      <div class="admin-layout__brand">
        <img src="/platform-logo.png" alt="校园Token管理平台" class="admin-layout__brand-logo" />
        <transition name="fade">
          <span v-if="!collapsed" class="admin-layout__brand-text">校园 Token 管理平台</span>
        </transition>
        <button
          class="admin-layout__collapse-btn"
          :class="{ 'admin-layout__collapse-btn--collapsed': collapsed }"
          :aria-label="collapsed ? '展开侧边栏' : '收起侧边栏'"
          @click="toggleCollapsed"
        >
          <el-icon :size="13"><Fold v-if="!collapsed" /><Expand v-else /></el-icon>
        </button>
      </div>

      <nav class="admin-layout__nav">
        <div v-for="group in navGroups" :key="group.label" class="admin-layout__group">
          <div v-if="!collapsed" class="admin-layout__group-label">{{ group.label }}</div>
          <router-link
            v-for="item in group.items"
            :key="item.to"
            :to="item.to"
            class="admin-layout__nav-item"
            :class="{ 'admin-layout__nav-item--active': isActive(item) }"
            :title="collapsed ? item.label : undefined"
            @click="onNavigate"
          >
            <el-icon :size="17"><component :is="item.icon" /></el-icon>
            <span v-if="!collapsed" class="admin-layout__nav-text">{{ item.label }}</span>
          </router-link>
        </div>
      </nav>

      <div class="admin-layout__sidebar-footer">
        <transition name="fade">
          <div v-if="!collapsed" class="admin-layout__me">
            <div class="admin-layout__me-avatar">{{ (auth.username || 'A')[0].toUpperCase() }}</div>
            <div class="admin-layout__me-meta">
              <div class="admin-layout__me-name">{{ auth.username || '-' }}</div>
              <div class="admin-layout__me-role">{{ roleLabel }}</div>
            </div>
          </div>
        </transition>
        <button class="admin-layout__logout" :title="collapsed ? '退出登录' : undefined" @click="logout">
          <el-icon :size="15"><SwitchButton /></el-icon>
          <span v-if="!collapsed">退出登录</span>
        </button>
      </div>
    </aside>

    <!-- 主内容区：高度固定，内部页面自管布局 -->
    <div class="admin-layout__main">
      <button
        class="admin-layout__menu-btn"
        aria-label="打开菜单"
        @click="drawerOpen = true"
      >
        <el-icon :size="18"><Expand /></el-icon>
      </button>
      <main class="admin-layout__content">
        <router-view v-slot="{ Component }">
          <component :is="Component" />
        </router-view>
      </main>
    </div>
  </div>
</template>

<style scoped>
.admin-layout {
  height: 100vh;
  display: flex;
  overflow: hidden;
  background: var(--bg-page);
}

/* 侧边栏 */
.admin-layout__sidebar {
  width: var(--sidebar-width);
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  background: var(--bg-sidebar);
  border-right: 1px solid var(--border-light);
  transition: width 0.2s ease;
  z-index: 50;
}
.admin-layout__sidebar--collapsed {
  width: var(--sidebar-collapsed-width);
}
.admin-layout__brand {
  height: 52px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 0 14px;
  border-bottom: 1px solid var(--border-light);
  overflow: hidden;
}
.admin-layout__brand-logo {
  width: 26px;
  height: 26px;
  flex-shrink: 0;
  object-fit: contain;
}
.admin-layout__brand-text {
  font-size: 13.5px;
  font-weight: 700;
  white-space: nowrap;
  color: var(--text-primary);
}
.admin-layout__nav {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  overflow-x: hidden;
  padding: 10px 8px;
}
.admin-layout__group {
  margin-bottom: 14px;
}
.admin-layout__group-label {
  padding: 0 10px 6px;
  font-size: 11px;
  font-weight: 600;
  color: var(--text-muted);
}
.admin-layout__nav-item {
  display: flex;
  align-items: center;
  gap: 10px;
  height: 38px;
  padding: 0 10px;
  margin-bottom: 2px;
  border-radius: var(--radius-sm);
  color: var(--text-secondary);
  text-decoration: none;
  white-space: nowrap;
  transition:
    background 0.15s,
    color 0.15s;
}
.admin-layout__nav-item:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}
.admin-layout__nav-item--active {
  background: var(--bg-active);
  color: var(--brand);
  font-weight: 600;
  box-shadow: inset 2px 0 0 var(--brand);
}
.admin-layout__sidebar-footer {
  flex-shrink: 0;
  padding: 10px;
  border-top: 1px solid var(--border-light);
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.admin-layout__me {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 6px 8px;
  border-radius: var(--radius-sm);
  background: var(--bg-hover);
}
.admin-layout__me-avatar {
  width: 30px;
  height: 30px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  background: linear-gradient(135deg, var(--brand), var(--brand-hover));
  color: #fff;
  font-size: 13px;
  font-weight: 700;
}
.admin-layout__me-name {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
  line-height: 18px;
}
.admin-layout__me-role {
  font-size: 11px;
  color: var(--text-muted);
  line-height: 15px;
}
.admin-layout__logout {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  height: 30px;
  border: 1px solid var(--border-light);
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--text-secondary);
  font-size: 12px;
  cursor: pointer;
  transition: all 0.15s;
}
.admin-layout__logout:hover {
  color: var(--danger);
  border-color: var(--danger);
  background: var(--danger-bg);
}

/* 主内容区 */
.admin-layout__main {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  position: relative;
}
.admin-layout__content {
  flex: 1;
  min-height: 0;
  padding: var(--page-padding);
}
.admin-layout__collapse-btn {
  margin-left: auto;
  flex-shrink: 0;
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px solid var(--border-default);
  border-radius: 50%;
  background: var(--bg-card);
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.15s;
}
.admin-layout__collapse-btn:hover {
  color: var(--brand);
  border-color: var(--brand);
}
.admin-layout__menu-btn {
  display: none;
  position: absolute;
  top: 12px;
  left: 12px;
  z-index: 30;
  width: 34px;
  height: 34px;
  align-items: center;
  justify-content: center;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  background: var(--bg-card);
  color: var(--text-secondary);
  cursor: pointer;
}
.admin-layout__mask {
  position: fixed;
  inset: 0;
  z-index: 40;
  background: var(--bg-mask);
}

/* 移动端（<1024px）：侧边栏变抽屉 */
@media (max-width: 1023px) {
  .admin-layout__sidebar {
    position: fixed;
    left: 0;
    top: 0;
    height: 100vh;
    width: var(--sidebar-width);
    transform: translateX(-100%);
    transition: transform 0.2s ease;
    box-shadow: var(--shadow-pop);
  }
  .admin-layout__sidebar--drawer-open {
    transform: translateX(0);
  }
  .admin-layout__menu-btn {
    display: flex;
  }
  .admin-layout__content {
    padding-top: calc(var(--page-padding) + 40px);
  }
}

/* 折叠态下隐藏文字 */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.15s;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
