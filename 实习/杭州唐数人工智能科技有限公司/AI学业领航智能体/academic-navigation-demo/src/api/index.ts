import request from '@/utils/request'
import type {
  AcademicProfile,
  AlertTrendData,
  AlertItem,
  AlertStats,
  CourseHeatmapData,
  CourseTeacherCourse,
  CourseStatsItem,
  DeanAlertStatsData,
  DeanData,
  DepartmentAlertItem,
  DepartmentCourseItem,
  DepartmentOverviewData,
  DevelopmentAnalyzeData,
  DevelopmentAnalyzePayload,
  DevelopmentPath,
  GpaProgressData,
  GpaTrendItem,
  InterventionPayload,
  JobItem,
  PageResult,
  PlanAnalysisData,
  SchoolItem,
  Student,
  StudentDashboardData,
  StudentTermGradesData,
  StudentDetailData,
  TeacherDashboardData,
  TeacherTermContext,
} from '@/types'

export interface ApiResult<T> {
  code: number
  message: string
  data: T
  timestamp: number
}

export interface AgentMemoryItem {
  id: string
  memory: string
  score: number | null
  createdAt: string | null
}

export interface LifePlanningChatPayload {
  message: string
  remember: boolean
  history: Array<{ role: 'user' | 'assistant'; content: string }>
}

export interface LifePlanningChatResult {
  answer: string
  memoryAvailable: boolean
  remembered: boolean
  memoryNotice: string | null
}

async function get<T>(url: string, params?: Record<string, unknown>): Promise<T> {
  const res = await request.get(url, { params })
  return (res as unknown as ApiResult<T>).data
}

async function post<T>(url: string, body?: unknown): Promise<T> {
  const res = await request.post(url, body)
  return (res as unknown as ApiResult<T>).data
}

// 学生端
export function getStudentDashboard(term?: string) {
  return get<StudentDashboardData>('/student/dashboard', { term })
}

export function getStudentTerms() {
  return get<string[]>('/student/terms')
}

export function getStudentGrades(term?: string) {
  return get<StudentTermGradesData>('/student/grades', { term })
}

export function getStudentProfile(term?: string) {
  return get<AcademicProfile>('/student/profile', { term })
}

export function getStudentGpaTrend(term?: string) {
  return get<GpaTrendItem[]>('/student/gpa-trend', { term })
}

export function getStudentAlerts(term?: string, page = 1, size = 20) {
  return get<PageResult<AlertItem>>('/student/alerts', { term, page, size })
}

export function chatWithLifePlanningAgent(payload: LifePlanningChatPayload) {
  return post<LifePlanningChatResult>('/agents/life-planning/chat', payload)
}

export function getLifePlanningMemories() {
  return get<AgentMemoryItem[]>('/agents/life-planning/memories')
}

export function deleteLifePlanningMemory(id: string) {
  return request.delete(`/agents/life-planning/memories/${encodeURIComponent(id)}`)
}

// 班主任端
export function getTeacherDashboard(term?: string) {
  return get<TeacherDashboardData>('/teacher/dashboard', { term })
}

export function getTeacherTerms() {
  return get<TeacherTermContext>('/teacher/terms')
}

export function getTeacherGpaProgress(term?: string) {
  return get<GpaProgressData>('/teacher/gpa-progress', { term })
}

export function getTeacherCourseStats(term?: string) {
  return get<CourseStatsItem[]>('/teacher/course-stats', { term })
}

export function getTeacherStudents(keyword?: string, page = 1, size = 20, term?: string) {
  return get<PageResult<Student>>('/teacher/students', { keyword, page, size, term })
}

export function getTeacherStudentDetail(studentId: string, term?: string) {
  return get<StudentDetailData>(`/teacher/students/${studentId}`, { term })
}

// 任课教师端
export function getCourseTeacherCourses(term?: string) {
  return get<CourseTeacherCourse[]>('/course-teacher/courses', { term })
}

export function getCourseTeacherTerms() {
  return get<string[]>('/course-teacher/terms')
}

export function getCourseTeacherAlerts(term?: string, page = 1, size = 20) {
  return get<PageResult<AlertItem>>('/course-teacher/alerts', { term, page, size })
}

// 预警模块
export interface AlertQuery {
  term?: string
  level?: string
  status?: string
  keyword?: string
  page?: number
  size?: number
}

export function getAlerts(query: AlertQuery = {}) {
  const params: Record<string, unknown> = {}
  if (query.term) params.term = query.term
  if (query.level && query.level !== 'all') params.level = query.level
  if (query.status && query.status !== 'all') params.status = query.status
  if (query.keyword) params.keyword = query.keyword
  params.page = query.page ?? 1
  params.size = query.size ?? 20
  return get<PageResult<AlertItem>>('/alerts', params)
}

export function getAlertDetail(id: string) {
  return get<AlertItem>(`/alerts/${id}`)
}

export function getUnreadCount() {
  return get<number>('/alerts/unread-count')
}

export function confirmAlert(id: string) {
  return post<void>(`/alerts/${id}/confirm`)
}

export function interveneAlert(id: string, payload: InterventionPayload) {
  return post<void>(`/alerts/${id}/intervention`, payload)
}

export function getAlertStats(term?: string) {
  return get<AlertStats>('/alerts/stats', { term })
}

// 系主任端
export function getDepartmentTerms() {
  return get<string[]>('/department/terms')
}

export function getDepartmentOverview(term?: string) {
  return get<DepartmentOverviewData>('/department/overview', { term })
}

export function getDepartmentCourseHeatmap(term?: string) {
  return get<CourseHeatmapData>('/department/courses/heatmap', { term })
}

export function getDepartmentCourses(term?: string) {
  return get<DepartmentCourseItem[]>('/department/courses', { term })
}

export function getDepartmentPlanAnalysis(grade?: string) {
  return get<PlanAnalysisData>('/department/plan-analysis', { grade })
}

export interface DepartmentAlertQuery {
  term?: string
  grade?: string
  level?: string
  status?: string
  page?: number
  size?: number
}

export function getDepartmentAlerts(query: DepartmentAlertQuery = {}) {
  const params: Record<string, unknown> = {}
  if (query.term) params.term = query.term
  if (query.grade && query.grade !== '全部年级') params.grade = query.grade
  if (query.level && query.level !== 'all') params.level = query.level
  if (query.status && query.status !== 'all') params.status = query.status
  params.page = query.page ?? 1
  params.size = query.size ?? 20
  return get<PageResult<DepartmentAlertItem>>('/department/alerts', params)
}

// 院长端
export function getDeanTerms() {
  return get<string[]>('/dean/terms')
}

export function getDeanDashboard(term?: string) {
  return get<DeanData>('/dean/dashboard', { term })
}

export function getDeanAlerts(term?: string, groupBy?: string) {
  const params: Record<string, unknown> = {}
  if (term) params.term = term
  if (groupBy) params.groupBy = groupBy
  return get<DeanAlertStatsData>('/dean/alerts', params)
}

export function getDeanAlertTrend(terms?: string) {
  return get<AlertTrendData>('/dean/alerts/trend', { terms })
}

// 发展引导
export function getDevelopmentPaths() {
  return get<DevelopmentPath[]>('/development/paths')
}

export function getDevelopmentJobs(category?: string, page = 1, size = 20) {
  return get<PageResult<JobItem>>('/development/jobs', { category, page, size })
}

export function getDevelopmentSchools(type: 'grad' | 'overseas', tier?: string, page = 1, size = 20) {
  const params: Record<string, unknown> = { type, page, size }
  if (tier) params.tier = tier
  return get<PageResult<SchoolItem>>('/development/schools', params)
}

export function analyzeDevelopmentTarget(payload: DevelopmentAnalyzePayload) {
  return post<DevelopmentAnalyzeData>('/development/analyze', payload)
}

// AI 指标解读
export interface AiInterpretResult {
  content: string
  source: 'local' | 'fallback'
}

export function interpretMetrics(instruction: string, data: Record<string, unknown>) {
  return post<AiInterpretResult>('/ai/interpret', { instruction, data })
}
