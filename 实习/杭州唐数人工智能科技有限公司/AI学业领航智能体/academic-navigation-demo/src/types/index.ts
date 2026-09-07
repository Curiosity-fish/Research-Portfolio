export interface User {
  id: string | number
  account?: string
  name: string
  role: 'student' | 'teacher' | 'course_teacher' | 'department' | 'dean'
  avatar?: string
  studentId?: string
  teacherId?: string
  college?: string
  major?: string
  className?: string
  grade?: string
}

export interface Student {
  id: string
  name: string
  studentId: string
  college: string
  major: string
  className: string
  grade: string
  gpa: number
  rank: number
  totalStudents: number
  alertLevel: 'none' | 'yellow' | 'orange' | 'red'
  avatar?: string
}

export interface ProfileDimension {
  key: string
  label: string
  score: number
  avgScore: number
  maxScore: number
  description: string
  details: string[]
}

export interface AcademicProfile {
  dimensions: ProfileDimension[]
  gpaTrend: GpaTrendItem[]
  rankTrend: RankTrendItem[]
}

export interface GpaTrendItem {
  semester: string
  gpa: number
  avg: number
}

export interface RankTrendItem {
  semester: string
  rank: number
  total: number
}

export interface AlertItem {
  id: string
  studentId: string
  studentName: string
  type: string
  level: 'yellow' | 'orange' | 'red'
  title: string
  description: string
  course?: string
  failedCourses?: string[]
  date: string
  status: 'pending' | 'processing' | 'resolved'
  suggestion?: string
  triggerEvent?: string
  pushedAt?: string
}

export interface ChatMessage {
  id: string
  role: 'user' | 'ai'
  content: string
  timestamp: number
  cardData?: ChatCardData
}

export interface ChatCardData {
  type: 'score' | 'radar' | 'alert' | 'trend'
  title: string
  data: Record<string, unknown>
}

export interface StudentDashboardData {
  gpa: number
  rank: number
  totalStudents: number
  classRank: number
  classTotalStudents: number
  credits: number
  totalCredits: number
  courseCount: number
  alertCount: number
  healthScore: number
  gpaTrend: GpaTrendItem[]
  recentAlerts: AlertItem[]
}

export interface StudentCourseGrade {
  courseCode: string
  courseName: string
  credits: number
  score: number | null
  gradePoint: number | null
  status: 'pending' | 'passed' | 'failed' | 'retake'
}

export interface StudentTermGradesData {
  term: string
  gpa: number
  avgGpa: number | null
  rank: number | null
  totalStudents: number | null
  earnedCredits: number
  courseCount: number
  passedCount: number
  failedCount: number
  courses: StudentCourseGrade[]
}

export interface PageResult<T> {
  list: T[]
  total: number
  page: number
  size: number
}

export interface HealthReportItem {
  name: string
  score: number
  level?: string
}

export interface HealthReport {
  id: string
  studentId: string
  term: string
  totalScore: number | null
  items: HealthReportItem[]
  reportUrl?: string | null
}

export interface ResourceItem {
  id: string
  type: 'video' | 'doc' | 'exercise' | 'book'
  subject: string
  title: string
  source: string
  duration?: string
  matchScore: number
  matchReason: string
  tags: string[]
  difficulty: 'easy' | 'medium' | 'hard'
}

export interface PeerStudent {
  id: string
  studentId: string
  name: string
  gpa: number
  rank: number
  similarity: number
  highlights: string[]
}

export interface GpaProgressStudent {
  id: string
  name: string
  studentId: string
  gpaByTerm: Record<string, number>
}

export interface GpaProgressData {
  terms: string[]
  students: GpaProgressStudent[]
}

export interface TeacherTermContext {
  className: string
  latestTerm: string
  terms: string[]
}

export interface CourseFailStudent {
  name: string
  studentId: string
  score: number
}

export interface TermFailStudent {
  name: string
  studentId: string
  failedCourses: string[]
  lowestScore: number | null
}

export interface CourseStatsItem {
  id: string
  name: string
  credits: number
  avgScore: number
  passRate: number
  failRate: number
  failStudents: CourseFailStudent[]
}

export interface CourseClassStat {
  className: string
  avg: number
  pass: number
  high: number
  low: number
  count: number
  risk: number
  failStudents: CourseFailStudent[]
  termFailStudents: TermFailStudent[]
}

export interface CourseTeacherCourse {
  id: string
  name: string
  code: string
  term: string
  students: number
  avgScore: number
  passRate: number
  highRiskCount: number
  scoreDistribution: number[]
  classes: CourseClassStat[]
}

export interface StudentDetailData {
  student: Student
  profile: AcademicProfile
  alerts: AlertItem[]
  gpaTrend: GpaTrendItem[]
}

export interface InterventionPayload {
  interventionDate: string
  methods: string[]
  content: string
  studentResponse: string
  followUpPlan?: string
}

export interface AlertStats {
  total: number
  yellow: number
  orange: number
  red: number
}

export interface TeacherDashboardData {
  totalStudents: number
  avgGpa: number
  alertCount: number
  passRate: number
  alertDistribution: { yellow: number; orange: number; red: number }
  dimensionAvg: { label: string; value: number }[]
  recentAlerts: AlertItem[]
  gradeDistribution: { name: string; value: number }[]
}

// ── Phase 2 Types ──────────────────────────────────────────────

export interface InterventionRecord {
  id: string
  alertId: string
  teacherId: string
  interventionDate: string
  methods: ('phone' | 'meeting' | 'message' | 'other')[]
  content: string
  studentResponse: 'positive' | 'neutral' | 'resistant'
  followUpPlan?: string
  createdAt: number
}

export interface GapItem {
  dimension: string
  current: number | string
  target: number | string
  urgent: 'high' | 'medium' | 'low'
}

export interface ActionItem {
  title: string
  description: string
  urgency: 'high' | 'medium' | 'low'
  deadline?: string
}

export interface MilestoneItem {
  title: string
  date: string
  status: 'done' | 'current' | 'pending'
  description?: string
}

export interface DevelopmentPath {
  key: 'graduate' | 'overseas' | 'civil' | 'employment' | 'institution'
  label: string
  matchScore: number
  gapItems: GapItem[]
  actionItems: ActionItem[]
  milestones: MilestoneItem[]
}

export interface DepartmentData {
  totalStudents: number
  avgGpa: number
  alertRate: number
  passRate: number
  gradeGpaTrend: { grade: string; trend: GpaTrendItem[] }[]
  courseHeatmap: { course: string; periods: { period: string; passRate: number }[] }[]
}

export interface DeanData {
  totalStudents: number
  avgGpa: number
  alertRate: number
  coursePassRate: number
  interventionResponseRate: number
  departmentRanking: { dept: string; healthScore: number; passRate: number; avgGpa: number }[]
  alertHeatmap: { dept: string; yellow: number; orange: number; red: number }[]
}

export interface SandboxParams {
  studyHoursPerWeek: number
  attendanceRate: number
  assignmentCompletionRate: number
  libraryVisitsPerMonth: number
  targetSemester: '下学期' | '下下学期' | string
}

export interface SandboxPrediction {
  predictedGpa: number
  predictedRank: number
  gpaDelta: number
  rankDelta: number
  riskLevel: 'high' | 'medium' | 'low'
  insights: string[]
  suggestions: string[]
}

// ── Phase 3 Types ──────────────────────────────────────────────

export interface DepartmentGradeTrend {
  grade: string
  trend: GpaTrendItem[]
}

export interface DepartmentFocusStudent {
  id: string
  name: string
  studentId: string
  grade: string
  major: string
  className?: string
  gpa: number
  alertLevel: 'none' | 'yellow' | 'orange' | 'red'
}

export interface DepartmentOverviewData extends Omit<DepartmentData, 'courseHeatmap'> {
  alertDistribution: { yellow: number; orange: number; red: number }
  focusStudents: DepartmentFocusStudent[]
}

export interface DepartmentCourseGradeStat {
  grade: string
  avgScore: number
  passRate: number
  highRate: number
  lowRate: number
}

export interface DepartmentCourseItem {
  id: string
  name: string
  code: string
  credits: number
  avgScore: number
  passRate: number
  highRate: number
  lowRate: number
  gradeStats: DepartmentCourseGradeStat[]
}

export interface CourseHeatmapData {
  courses: string[]
  grades: string[]
  heatData: [number, number, number][]
}

export interface PlanAnalysisData {
  grade: string
  dimensions: string[]
  standard: number[]
  actual: number[]
  tips: string[]
}

export interface DepartmentAlertItem extends AlertItem {
  grade: string
  className: string
  major: string
  gpa: number
  trend: 'up' | 'down' | 'flat'
  absences: number
  reason: string
}

export interface DeanAlertRow {
  dept: string
  totalStudents: number
  red: number
  orange: number
  yellow: number
  rate: number
}

export interface DeanGradeAlertRow {
  grade: string
  red: number
  orange: number
  yellow: number
  total: number
  note: string
}

export interface DeanAlertStatsData {
  term: string
  total: { red: number; orange: number; yellow: number; total: number }
  byDept: DeanAlertRow[]
  byGrade: DeanGradeAlertRow[]
}

export interface AlertTrendData {
  terms: string[]
  red: number[]
  orange: number[]
  yellow: number[]
}

export type SchoolItem = GradSchool & OverseasSchool

export interface DevelopmentAnalyzePayload {
  targetKey: string
  targetLabel?: string
  targetScore?: number
}

export interface DevelopmentAnalyzeData {
  targetKey: string
  targetLabel: string
  matchScore: number
  gaps: GapItem[]
  actions: ActionItem[]
  milestones: MilestoneItem[]
  suggestions: string[]
}

export interface JobItem {
  id: string
  platform: 'boss' | '51job' | 'zhilian' | 'liepin'
  platformLabel: string
  companyName: string
  companySize: string
  companyStage: string
  jobTitle: string
  salaryRange: string
  city: string
  district: string
  education: string
  experience: string
  tags: string[]
  highlights: string[]
  matchScore: number
  matchReasons: string[]
  publishDate: string
  sourceUrl: string
  category: 'backend' | 'frontend' | 'ai' | 'test' | 'pm' | 'presales' | 'data' | 'security'
}

export interface GradSchool {
  id: string
  name: string
  location: string
  rank: string
  tier: 'top' | 'good' | 'match'
  programs: string[]
  admissionGpa: string
  examRequirements: string[]
  highlights: string[]
  matchScore: number
  matchReasons: string[]
  officialUrl: string
}

export interface OverseasSchool {
  id: string
  name: string
  nameZh: string
  country: string
  countryEmoji: string
  rank: string
  tier: 'top' | 'good' | 'match'
  programs: string[]
  admissionGpa: string
  languageRequirement: string
  highlights: string[]
  matchScore: number
  matchReasons: string[]
  officialUrl: string
  deadline: string
}

// ── Phase 3 Types ──────────────────────────────────────────────
