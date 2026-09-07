import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', redirect: '/login' },
    {
      path: '/login',
      name: 'Login',
      component: () => import('@/views/login/Index.vue'),
    },

    // 学生端
    {
      path: '/student',
      component: () => import('@/layouts/StudentLayout.vue'),
      children: [
        { path: '', redirect: '/student/dashboard' },
        { path: 'dashboard', name: 'StudentDashboard', component: () => import('@/views/student/Dashboard.vue') },
        { path: 'profile', name: 'AcademicProfile', component: () => import('@/views/student/Profile.vue') },
        { path: 'alerts', name: 'StudentAlerts', component: () => import('@/views/student/Alerts.vue') },
        { path: 'development-path', name: 'DevelopmentPath', component: () => import('@/views/student/DevelopmentPath.vue') },
        { path: 'agents', name: 'StudentAgentPlaza', component: () => import('@/views/shared/AgentPlaza.vue') },
        { path: 'agents/:id', name: 'StudentAgentChat', component: () => import('@/views/shared/AgentChat.vue') },
        { path: 'quick-access', name: 'StudentQuickAccess', component: () => import('@/views/shared/QuickAccess.vue') },
      ],
    },

    // 教师端
    {
      path: '/teacher',
      component: () => import('@/layouts/TeacherLayout.vue'),
      children: [
        { path: '', redirect: '/teacher/dashboard' },
        { path: 'dashboard', name: 'TeacherDashboard', component: () => import('@/views/teacher/Dashboard.vue') },
        { path: 'alerts', name: 'AlertManagement', component: () => import('@/views/teacher/AlertManagement.vue') },
        { path: 'gpa-progress', name: 'GpaProgress', component: () => import('@/views/teacher/GpaProgress.vue') },
        { path: 'course-stats', name: 'CourseStats', component: () => import('@/views/teacher/CourseStats.vue') },
        { path: 'student/:id', name: 'StudentDetail', component: () => import('@/views/teacher/StudentDetail.vue') },
        { path: 'agents', name: 'TeacherAgentPlaza', component: () => import('@/views/shared/AgentPlaza.vue') },
        { path: 'agents/:id', name: 'TeacherAgentChat', component: () => import('@/views/shared/AgentChat.vue') },
        { path: 'quick-access', name: 'TeacherQuickAccess', component: () => import('@/views/shared/QuickAccess.vue') },
      ],
    },

    // 专业负责人端
    {
      path: '/department',
      component: () => import('@/layouts/DepartmentLayout.vue'),
      children: [
        { path: '', redirect: '/department/overview' },
        { path: 'overview', name: 'DepartmentOverview', component: () => import('@/views/department/Overview.vue') },
        { path: 'courses', name: 'CourseHeatmap', component: () => import('@/views/department/Courses.vue') },
        { path: 'alerts', name: 'DepartmentAlerts', component: () => import('@/views/department/Alerts.vue') },
        { path: 'agents', name: 'DepartmentAgentPlaza', component: () => import('@/views/shared/AgentPlaza.vue') },
        { path: 'agents/:id', name: 'DepartmentAgentChat', component: () => import('@/views/shared/AgentChat.vue') },
        { path: 'quick-access', name: 'DepartmentQuickAccess', component: () => import('@/views/shared/QuickAccess.vue') },
      ],
    },

    // 院领导端
    {
      path: '/dean',
      component: () => import('@/layouts/DeanLayout.vue'),
      children: [
        { path: '', redirect: '/dean/dashboard' },
        { path: 'dashboard', name: 'DeanDashboard', component: () => import('@/views/dean/Dashboard.vue') },
        { path: 'alerts', name: 'DeanAlerts', component: () => import('@/views/dean/Alerts.vue') },
        { path: 'agents', name: 'DeanAgentPlaza', component: () => import('@/views/shared/AgentPlaza.vue') },
        { path: 'agents/:id', name: 'DeanAgentChat', component: () => import('@/views/shared/AgentChat.vue') },
        { path: 'quick-access', name: 'DeanQuickAccess', component: () => import('@/views/shared/QuickAccess.vue') },
      ],
    },

    // 任课老师端
    {
      path: '/course-teacher',
      component: () => import('@/layouts/CourseTeacherLayout.vue'),
      children: [
        { path: '', redirect: '/course-teacher/dashboard' },
        { path: 'dashboard', name: 'CourseTeacherDashboard', component: () => import('@/views/course-teacher/Dashboard.vue') },
        { path: 'alerts', name: 'CourseTeacherAlerts', component: () => import('@/views/course-teacher/Alerts.vue') },
        { path: 'agents', name: 'CourseTeacherAgentPlaza', component: () => import('@/views/shared/AgentPlaza.vue') },
        { path: 'agents/:id', name: 'CourseTeacherAgentChat', component: () => import('@/views/shared/AgentChat.vue') },
        { path: 'quick-access', name: 'CourseTeacherQuickAccess', component: () => import('@/views/shared/QuickAccess.vue') },
      ],
    },
  ],
})

const publicPaths = ['/login']

router.beforeEach(async (to) => {
  if (publicPaths.includes(to.path)) return true

  const token = localStorage.getItem('token')
  if (!token) return { path: '/login' }

  const auth = useAuthStore()
  if (!auth.isLoggedIn) {
    const ok = await auth.restore()
    if (!ok) return { path: '/login' }
  }
  return true
})

export default router
