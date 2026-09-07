import fs from 'node:fs/promises'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { SpreadsheetFile, Workbook } from '@oai/artifact-tool'

const scriptDir = path.dirname(fileURLToPath(import.meta.url))

function readArg(name, fallback) {
  const flag = `--${name}`
  const idx = process.argv.indexOf(flag)
  if (idx >= 0 && process.argv[idx + 1] !== undefined) return process.argv[idx + 1]
  return fallback
}

const studentCount = Math.max(1, Number(readArg('students', '120')))
const classCount = Math.max(1, Number(readArg('classes', '4')))
const termCount = Math.max(1, Number(readArg('terms', '8')))
const majorCount = Math.max(1, Number(readArg('majors', '4')))
const courseCount = Math.max(1, Number(readArg('courses', '12')))
const departmentCount = Math.max(1, Number(readArg('departments', '2')))
const firstYear = Number(readArg('first-year', '2024'))
const idOffset = Number(readArg('id-offset', '100000'))
const seed = Number(readArg('seed', '20260812'))
const outDir = path.resolve(readArg('out', path.join(scriptDir, '../../outputs/sample-data')))

function mulberry32(a) {
  return function () {
    a |= 0
    a = (a + 0x6D2B79F5) | 0
    let t = Math.imul(a ^ (a >>> 15), 1 | a)
    t = (t + Math.imul(t ^ (t >>> 7), 61 | t)) ^ t
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296
  }
}

const rand = mulberry32(seed)
const randInt = (min, max) => Math.floor(rand() * (max - min + 1)) + min
const randNum = (min, max, digits = 1) => Number((min + rand() * (max - min)).toFixed(digits))
const pick = (arr) => arr[Math.floor(rand() * arr.length)]

const surnames = ['李', '王', '张', '刘', '陈', '杨', '赵', '黄', '周', '吴', '徐', '孙', '马', '朱', '胡', '郭', '何', '高', '林', '罗']
const givenNames = ['明', '芳', '磊', '洋', '静', '晨', '欣', '杰', '婷', '浩', '雪', '超', '宇', '悦', '鹏', '敏', '航', '思', '强', '怡']

const passwordHash = '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO'

const terms = []
for (let t = 0; t < termCount; t++) {
  const startYear = firstYear + Math.floor(t / 2)
  const termNo = (t % 2) + 1
  terms.push(`${startYear}-${startYear + 1}-${termNo}`)
}

const latestTerm = terms[terms.length - 1]

// ---- 基础档案 ----
const departmentNames = [
  '数据科学与工程测试学院',
  '智能计算测试学院',
  '网络与信息安全测试学院',
  '数字媒体测试学院',
]
const departments = []
for (let d = 0; d < departmentCount; d++) {
  departments.push({
    id: idOffset + 100 + d,
    name: departmentNames[d] ?? `测试学院${d + 1}`,
    code: `TEST_${String(d + 1).padStart(2, '0')}`,
    dean_id: null,
  })
}

const majorNames = [
  '数据科学与大数据技术',
  '软件工程',
  '计算机科学与技术',
  '人工智能',
  '信息安全',
  '物联网工程',
  '数字媒体技术',
  '智能科学与技术',
]
const majors = []
for (let m = 0; m < majorCount; m++) {
  const name = majorNames[m] ?? `测试专业${m + 1}`
  const department = departments[m % departmentCount]
  majors.push({
    id: idOffset + 200 + m,
    name,
    code: `MAJOR_${String(m + 1).padStart(2, '0')}`,
    department_id: department.id,
    degree_type: '工学学士',
  })
}

const classes = []
for (let c = 0; c < classCount; c++) {
  const major = majors[c % majors.length]
  const department = departments.find((d) => d.id === major.department_id)
  const shortName = major.name.replace(/[（(].*[）)]/, '').slice(0, 2)
  classes.push({
    id: idOffset + 300 + c,
    name: `${shortName}${firstYear}${Math.floor(c / majors.length) + 1}`,
    grade: String(firstYear),
    major_id: major.id,
    department_id: department.id,
    advisor_id: null,
    total_students: 0,
  })
}

const teacherDefs = []
for (let c = 0; c < classCount; c++) {
  const department = departments.find((d) => d.id === classes[c].department_id)
  teacherDefs.push({
    teacherId: `T${firstYear}${String(c + 1).padStart(3, '0')}`,
    name: `班主任${c + 1}`,
    role: 'teacher',
    isClassAdvisor: 1,
    title: '副教授',
    departmentId: department.id,
  })
}
for (let d = 0; d < departments.length; d++) {
  teacherDefs.push({
    teacherId: `C${firstYear}${String(d + 1).padStart(3, '0')}`,
    name: `任课教师${d + 1}`,
    role: 'course_teacher',
    isClassAdvisor: 0,
    title: '讲师',
    departmentId: departments[d].id,
  })
}
for (let d = 0; d < departments.length; d++) {
  teacherDefs.push({
    teacherId: `D${firstYear}${String(d + 1).padStart(3, '0')}`,
    name: `系主任${d + 1}`,
    role: 'department',
    isClassAdvisor: 0,
    title: '系主任',
    departmentId: departments[d].id,
  })
}
teacherDefs.push({
  teacherId: `L${firstYear}001`,
  name: '赵院长',
  role: 'dean',
  isClassAdvisor: 0,
  title: '院长',
  departmentId: departments[0].id,
})

const users = []
const teachers = []
const studentNoBase = (firstYear + 2) * 10000

for (let i = 0; i < teacherDefs.length; i++) {
  const def = teacherDefs[i]
  const userId = idOffset + 500 + i
  const deptName = departments.find((d) => d.id === def.departmentId)?.name ?? departments[0].name
  def.userId = userId
  users.push({
    id: userId,
    account: def.teacherId,
    name: def.name,
    password: passwordHash,
    role: def.role,
    college: deptName,
    major: null,
    class_name: null,
    grade: null,
    avatar_url: null,
    is_active: 1,
  })
  teachers.push({
    id: idOffset + 400 + i,
    teacher_id: def.teacherId,
    user_id: userId,
    department_id: def.departmentId,
    title: def.title,
    is_class_advisor: def.isClassAdvisor,
    class_id: i < classCount ? classes[i].id : null,
  })
}

const deanUser = users.find((u) => u.account.startsWith('L'))
for (const d of departments) {
  d.dean_id = deanUser?.id ?? null
}

const students = []
const perClass = Math.floor(studentCount / classCount)
const remainder = studentCount % classCount

for (let c = 0; c < classCount; c++) {
  const count = perClass + (c < remainder ? 1 : 0)
  const cls = classes[c]
  const dept = departments.find((d) => d.id === cls.department_id)
  cls.total_students = count
  classes[c].advisor_id = teachers.find((t) => t.class_id === cls.id)?.id ?? null
  for (let i = 0; i < count; i++) {
    const idx = students.length
    const userId = idOffset + 500 + teacherDefs.length + idx
    const account = String(studentNoBase + idx + 1)
    users.push({
      id: userId,
      account,
      name: pick(surnames) + pick(givenNames),
      password: passwordHash,
      role: 'student',
      college: dept.name,
      major: majors.find((m) => m.id === cls.major_id).name,
      class_name: cls.name,
      grade: cls.grade,
      avatar_url: null,
      is_active: 1,
    })
    students.push({
      id: idOffset + 600 + idx,
      student_id: account,
      user_id: userId,
      class_id: cls.id,
      major_id: cls.major_id,
      department_id: cls.department_id,
      gpa: 0,
      rank: 0,
      total_students: 0,
      alert_level: 'none',
      total_credits: 0,
      required_credits: 0,
      name: users[users.length - 1].name,
    })
  }
}

// ---- 课程与开课 ----
const courseNames = [
  '数据结构',
  '高等数学',
  '线性代数',
  '操作系统',
  '大学英语',
  '计算机网络',
  '数据库原理',
  '软件工程',
  '编译原理',
  '计算机组成原理',
  '离散数学',
  '概率论与数理统计',
]
const courseCredits = [4, 4, 3, 3.5, 2, 3, 3, 3, 3, 3.5, 3, 3]
const courses = []
for (let ci = 0; ci < courseCount; ci++) {
  const department = departments[ci % departmentCount]
  courses.push({
    id: idOffset + 700 + ci,
    code: `TEST${String(101 + ci).padStart(3, '0')}`,
    name: courseNames[ci] ?? `测试课程${ci + 1}`,
    credits: courseCredits[ci] ?? 3,
    type: ci % 6 === 4 ? 'public' : ci % 6 === 5 ? 'elective' : 'required',
    department_id: department.id,
  })
}

const courseClasses = []
let courseClassId = idOffset + 800
for (const cls of classes) {
  for (const term of terms) {
    for (let ci = 0; ci < courses.length; ci++) {
      const course = courses[ci]
      const classAdvisor = teachers.find((t) => t.class_id === cls.id)
      const courseTeacher = teachers.find((t) => t.teacher_id.startsWith('C') && t.department_id === cls.department_id)
      courseClasses.push({
        id: courseClassId++,
        course_id: course.id,
        term,
        class_name: cls.name,
        teacher_id: ci % 3 === 0 ? classAdvisor.id : courseTeacher.id,
        max_students: 60,
        enrolled_students: cls.total_students,
        status: term === latestTerm ? 'active' : 'finished',
      })
    }
  }
}

function gradePoint(score) {
  if (score >= 90) return 4.0
  if (score >= 85) return 3.7
  if (score >= 80) return 3.3
  if (score >= 75) return 3.0
  if (score >= 70) return 2.7
  if (score >= 65) return 2.3
  if (score >= 60) return 2.0
  return 0.0
}

// ---- 成绩、GPA 历史、画像、体测 ----
const grades = []
const gpaHistory = []
const profileScores = []
const healthReports = []
const studentTermGpa = new Map()
let gradeId = idOffset + 900
let gpaId = idOffset + 1000
let profileId = idOffset + 2000
let healthId = idOffset + 3000

const dimensionMeta = [
  { key: 'academic', label: '学业成绩' },
  { key: 'ability', label: '实践能力' },
  { key: 'psychology', label: '心理素质' },
  { key: 'health', label: '体质健康' },
  { key: 'thought', label: '思想品德' },
]

for (const student of students) {
  const termGpaMap = {}
  for (const term of terms) {
    const cls = classes.find((c) => c.id === student.class_id)
    const termClasses = courseClasses.filter((cc) => cc.term === term && cc.class_name === cls.name)
    let weightedPoint = 0
    let totalCredit = 0
    const failedNames = []
    for (const cc of termClasses) {
      const course = courses.find((c) => c.id === cc.course_id)
      const ability = 55 + rand() * 35
      const trendBoost = Math.min(10, Math.max(-8, (students.indexOf(student) % 5) - 2 + (terms.indexOf(term) - termCount / 2) * 1.2))
      const score = Math.max(0, Math.min(100, Math.round(ability + trendBoost + (rand() * 12 - 6))))
      const point = gradePoint(score)
      const status = score >= 60 ? 'passed' : rand() < 0.18 ? 'retake' : 'failed'
      grades.push({
        id: gradeId++,
        student_id: student.id,
        course_class_id: cc.id,
        score,
        grade_point: point,
        status,
        term,
        exam_date: new Date(`${term.slice(0, 4)}-01-${String(randInt(5, 20)).padStart(2, '0')}`),
      })
      weightedPoint += point * course.credits
      totalCredit += course.credits
      if (score < 60) failedNames.push(course.name)
    }
    const gpa = totalCredit > 0 ? Number((weightedPoint / totalCredit).toFixed(2)) : 0
    termGpaMap[term] = gpa
    studentTermGpa.set(`${student.id}:${term}`, { gpa, failedNames })
    student.total_credits += totalCredit
    student.required_credits += totalCredit

    const dimensionScores = {
      academic: Math.max(35, Math.min(98, Math.round(35 + gpa * 16))),
      ability: randInt(45, 90),
      psychology: randInt(50, 95),
      health: randInt(55, 95),
      thought: randInt(60, 98),
    }
    const detailMap = {
      academic: [
        `GPA: ${gpa.toFixed(2)} / 4.0`,
        failedNames.length ? `挂科课程: ${failedNames.join('、')}` : '无挂科记录',
      ],
      ability: [`竞赛/项目经历 ${randInt(0, 6)} 项`, `实践测评 ${dimensionScores.ability} 分`],
      psychology: ['心理测评结果正常', `学业压力指数 ${randInt(20, 60)}`],
      health: [`体测总分 ${randInt(60, 95)} 分`, '建议每周保持 3 次以上运动'],
      thought: [`志愿服务累计 ${randInt(4, 60)} 小时`, '德育测评结果良好'],
    }
    for (const dim of dimensionMeta) {
      const score = dimensionScores[dim.key]
      profileScores.push({
        id: profileId++,
        student_id: student.id,
        term,
        dimension_key: dim.key,
        label: dim.label,
        score,
        avg_score: randInt(55, 80),
        max_score: 100,
        description: `${dim.label}测评结果（测试数据）`,
        details: JSON.stringify(detailMap[dim.key]),
      })
    }

    const healthItems = [
      { name: '体重指数', score: randInt(65, 95), level: pick(['良好', '优秀', '一般']) },
      { name: '肺活量', score: randInt(65, 95), level: pick(['良好', '优秀', '一般']) },
      { name: '耐力跑', score: randInt(60, 95), level: pick(['良好', '优秀', '一般']) },
      { name: '坐位体前屈', score: randInt(60, 95), level: pick(['良好', '优秀', '一般']) },
      { name: '立定跳远', score: randInt(55, 95), level: pick(['良好', '优秀', '一般']) },
    ]
    healthReports.push({
      id: healthId++,
      student_id: student.id,
      term,
      total_score: Math.round(healthItems.reduce((sum, item) => sum + item.score, 0) / healthItems.length),
      items: JSON.stringify(healthItems),
      report_url: null,
    })
  }

}

// 全部学生 GPA 生成后，再按专业统一计算平均 GPA 与排名
for (const student of students) {
  const majorStudents = students.filter((s) => s.major_id === student.major_id)
  for (const term of terms) {
    const gpas = majorStudents.map((s) => studentTermGpa.get(`${s.id}:${term}`)?.gpa ?? 0)
    const avgGpa = Number((gpas.reduce((a, b) => a + b, 0) / gpas.length).toFixed(2))
    const sorted = [...majorStudents].sort(
      (a, b) => (studentTermGpa.get(`${b.id}:${term}`)?.gpa ?? 0) - (studentTermGpa.get(`${a.id}:${term}`)?.gpa ?? 0),
    )
    const rank = sorted.findIndex((s) => s.id === student.id) + 1
    gpaHistory.push({
      id: gpaId++,
      student_id: student.id,
      term,
      gpa: studentTermGpa.get(`${student.id}:${term}`).gpa,
      avg_gpa: avgGpa,
      rank,
      total_students: majorStudents.length,
    })
  }

  const lastGpa = studentTermGpa.get(`${student.id}:${latestTerm}`).gpa
  const lastRank = gpaHistory.filter((g) => g.student_id === student.id && g.term === latestTerm)[0].rank
  student.gpa = lastGpa
  student.rank = lastRank
  student.total_students = majorStudents.length
  student.alert_level = lastGpa >= 3.5 ? 'none' : lastGpa >= 2.8 ? 'yellow' : lastGpa >= 2.0 ? 'orange' : 'red'
}

// ---- 预警与干预 ----
const alerts = []
const interventions = []
let alertId = idOffset + 4000
let interventionId = idOffset + 5000

const alertStatuses = ['pending', 'processing', 'resolved']

for (let i = 0; i < students.length; i++) {
  const student = students[i]
  const advisor = teachers.find((t) => t.class_id === student.class_id)
  const handlerUserId = advisor?.userId ?? teacherDefs[0].userId
  const termData = studentTermGpa.get(`${student.id}:${latestTerm}`)
  const failed = termData?.failedNames ?? []

  let alertCount = 0
  if (student.alert_level !== 'none' || failed.length) alertCount += 1
  else if (rand() < 0.55) alertCount += 1
  if (student.alert_level !== 'none' && rand() < 0.55) alertCount += 1
  if (rand() < 0.2) alertCount += 1

  for (let n = 0; n < alertCount; n++) {
    const day = 10 + n * 3 + randInt(0, 2)
    const status = alertStatuses[(i + n) % alertStatuses.length]
    let type, level, title, description, triggerEvent, suggestion, course, failedCourses
    if (n === 0 && failed.length) {
      type = failed.length >= 3 ? '学业危机' : '挂科预警'
      level = failed.length >= 3 ? 'red' : failed.length >= 2 ? 'orange' : 'yellow'
      title = `累计挂科 ${failed.length} 门`
      description = `最近学期挂科课程：${failed.join('、')}`
      triggerEvent = `累计不及格课程达 ${failed.length} 门`
      suggestion = '建议安排一对一学业辅导并制定补考计划'
      course = failed[0]
      failedCourses = failed
    } else if (n === 0) {
      type = '成绩预警'
      level = student.alert_level !== 'none' ? student.alert_level : pick(['yellow', 'yellow', 'orange'])
      const threshold = level === 'red' ? '2.0' : level === 'orange' ? '2.3' : '2.8'
      title = `GPA 低于 ${threshold}`
      description = `最近学期 GPA ${student.gpa.toFixed(2)}，存在学业下滑风险`
      triggerEvent = `GPA 低于 ${threshold} 预警线`
      suggestion = '建议预约学业咨询并调整学习计划'
      course = null
      failedCourses = []
    } else if (n === 1) {
      type = '出勤预警'
      level = 'yellow'
      title = `学期缺勤累计 ${randInt(5, 15)} 次`
      description = '近期课堂出勤率下降，存在学习状态下滑风险'
      triggerEvent = '月出勤率低于 85%'
      suggestion = '建议与辅导员沟通并关注课堂出勤'
      course = null
      failedCourses = []
    } else {
      type = '课堂参与预警'
      level = pick(['yellow', 'orange'])
      title = '课堂互动与作业提交率偏低'
      description = '近一个月作业按时提交率低于 70%，课堂参与度一般'
      triggerEvent = '作业提交率低于 70%'
      suggestion = '建议加入学习小组并设置作业提醒'
      course = null
      failedCourses = []
    }

    const alert = {
      id: alertId++,
      student_id: student.id,
      level,
      type,
      title,
      description,
      course,
      failed_courses: JSON.stringify(failedCourses),
      trigger_date: new Date(`${latestTerm.slice(0, 4)}-01-${String(day).padStart(2, '0')}`),
      status,
      suggestion,
      trigger_event: triggerEvent,
      pushed_at: `${latestTerm.slice(0, 4)}-01-${String(day).padStart(2, '0')} 09:${String(randInt(0, 59)).padStart(2, '0')}`,
      handler_id: status === 'pending' ? null : handlerUserId,
      handle_note: status === 'pending' ? null : '已完成初步沟通（测试数据）',
      student_confirmed_at: status === 'pending'
        ? null
        : `${latestTerm.slice(0, 4)}-01-${String(day + 1).padStart(2, '0')} 10:00`,
    }
    alerts.push(alert)

    if (status !== 'pending') {
      interventions.push({
        id: interventionId++,
        alert_id: alert.id,
        teacher_id: advisor?.id ?? teachers[0].id,
        intervention_date: new Date(`${latestTerm.slice(0, 4)}-01-${String(day + 2).padStart(2, '0')}`),
        methods: JSON.stringify(pick([['meeting'], ['phone', 'message'], ['meeting', 'message'], ['message']])),
        content: `已围绕“${title}”与学生面谈并制定跟进方案（测试数据）`,
        student_response: pick(['positive', 'neutral', 'resistant']),
        follow_up_plan: pick(['两周后复查作业完成情况', '每周复盘一次学习计划', '下次考试后评估改进情况']),
      })
    }
  }
}

// ---- SQL 导出 ----
function sqlEscape(value) {
  if (value === null || value === undefined) return 'NULL'
  if (value instanceof Date) {
    const d = value.toISOString().slice(0, 10)
    return `'${d}'`
  }
  if (typeof value === 'number') return String(value)
  return `'${String(value).replace(/'/g, "''")}'`
}

function insertSql(table, columns, rows, extra = '') {
  if (rows.length === 0) return `-- ${table}: 无数据\n`
  const colList = columns.map((c) => (c === 'rank' ? '`rank`' : c)).join(', ')
  const values = rows
    .map((row) => `  (${columns.map((c) => sqlEscape(row[c])).join(', ')})`)
    .join(',\n')
  return `INSERT INTO ${table} (${colList}) VALUES\n${values}${extra};\n`
}

const sqlParts = []
sqlParts.push(`-- 样本数据生成脚本输出`)
sqlParts.push(`-- 参数: departments=${departmentCount} students=${studentCount} classes=${classCount} majors=${majorCount} courses=${courseCount} terms=${termCount} firstYear=${firstYear} idOffset=${idOffset} seed=${seed}`)
sqlParts.push(`-- 注意: 使用显式 ID，请导入独立测试库或先清空对应表。密码统一为 123456 的 BCrypt 哈希。`)
sqlParts.push('SET NAMES utf8mb4;')
sqlParts.push('')
sqlParts.push(insertSql('t_department', ['id', 'name', 'code', 'dean_id'], departments))
sqlParts.push(insertSql('t_major', ['id', 'name', 'code', 'department_id', 'degree_type'], majors))
sqlParts.push(insertSql('t_teacher', ['id', 'teacher_id', 'user_id', 'department_id', 'title', 'is_class_advisor', 'class_id'], teachers))
sqlParts.push(insertSql('t_class', ['id', 'name', 'grade', 'major_id', 'advisor_id', 'total_students'], classes))
sqlParts.push(insertSql('t_user', ['id', 'account', 'name', 'password', 'role', 'college', 'major', 'class_name', 'grade', 'avatar_url', 'is_active'], users))
sqlParts.push(insertSql('t_student', ['id', 'student_id', 'user_id', 'class_id', 'major_id', 'department_id', 'gpa', 'rank', 'total_students', 'alert_level', 'total_credits', 'required_credits'], students))
sqlParts.push(insertSql('t_course', ['id', 'code', 'name', 'credits', 'type', 'department_id'], courses))
sqlParts.push(insertSql('t_course_class', ['id', 'course_id', 'term', 'class_name', 'teacher_id', 'max_students', 'enrolled_students', 'status'], courseClasses))
sqlParts.push(insertSql('t_grade', ['id', 'student_id', 'course_class_id', 'score', 'grade_point', 'status', 'term', 'exam_date'], grades))
sqlParts.push(insertSql('t_gpa_history', ['id', 'student_id', 'term', 'gpa', 'avg_gpa', 'rank', 'total_students'], gpaHistory))
sqlParts.push(insertSql('t_profile_score', ['id', 'student_id', 'term', 'dimension_key', 'label', 'score', 'avg_score', 'max_score', 'description', 'details'], profileScores))
sqlParts.push(insertSql('t_health_report', ['id', 'student_id', 'term', 'total_score', 'items', 'report_url'], healthReports))
sqlParts.push(insertSql('t_alert', ['id', 'student_id', 'level', 'type', 'title', 'description', 'course', 'failed_courses', 'trigger_date', 'status', 'suggestion', 'trigger_event', 'pushed_at', 'handler_id', 'handle_note', 'student_confirmed_at'], alerts))
sqlParts.push(insertSql('t_intervention_record', ['id', 'alert_id', 'teacher_id', 'intervention_date', 'methods', 'content', 'student_response', 'follow_up_plan'], interventions))

const sqlText = sqlParts.join('\n')
await fs.mkdir(outDir, { recursive: true })
await fs.writeFile(path.join(outDir, 'sample_data.sql'), sqlText, 'utf8')
console.log(`SQL written: ${path.join(outDir, 'sample_data.sql')} (${sqlText.length} bytes)`)

const parseJsonField = (row, field) => ({
  ...row,
  [field]: typeof row[field] === 'string' ? JSON.parse(row[field]) : row[field],
})
const toDateString = (value) => (value instanceof Date ? value.toISOString().slice(0, 10) : value)
const toIsoDateTime = (value) => (typeof value === 'string' && value.includes(' ') ? value.replace(' ', 'T') : value)
const mapDateFields = (row, fields) => {
  const result = { ...row }
  for (const field of fields) result[field] = toDateString(result[field])
  return result
}

const jsonTables = {
  t_department: departments,
  t_major: majors,
  t_class: classes,
  t_teacher: teachers,
  t_user: users,
  t_student: students,
  t_course: courses,
  t_course_class: courseClasses,
  t_grade: grades.map((r) => mapDateFields(r, ['exam_date'])),
  t_gpa_history: gpaHistory,
  t_profile_score: profileScores.map((r) => parseJsonField(r, 'details')),
  t_health_report: healthReports.map((r) => parseJsonField(r, 'items')),
  t_alert: alerts.map((r) => {
    const row = mapDateFields(parseJsonField(r, 'failed_courses'), ['trigger_date'])
    row.pushed_at = toIsoDateTime(row.pushed_at)
    row.student_confirmed_at = toIsoDateTime(row.student_confirmed_at)
    return row
  }),
  t_intervention_record: interventions.map((r) => mapDateFields(parseJsonField(r, 'methods'), ['intervention_date'])),
}

const jsonData = {
  generatedAt: new Date().toISOString(),
  params: {
    departments: departmentCount,
    students: studentCount,
    classes: classCount,
    majors: majorCount,
    courses: courseCount,
    terms: termCount,
    firstYear,
    idOffset,
    seed,
  },
  tables: jsonTables,
}
const jsonPath = path.join(outDir, 'sample_data.json')
await fs.writeFile(jsonPath, JSON.stringify(jsonData), 'utf8')
console.log(`JSON written: ${jsonPath}`)

// ---- Excel 导出 ----
const workbook = Workbook.create()
const readmeSheet = workbook.worksheets.add('说明')
const dictSheet = workbook.worksheets.add('字段字典')

const sheetDefs = [
  { name: 't_department', headers: ['id', 'name', 'code', 'dean_id'] },
  { name: 't_major', headers: ['id', 'name', 'code', 'department_id', 'degree_type'] },
  { name: 't_teacher', headers: ['id', 'teacher_id', 'user_id', 'department_id', 'title', 'is_class_advisor', 'class_id'] },
  { name: 't_class', headers: ['id', 'name', 'grade', 'major_id', 'advisor_id', 'total_students'] },
  { name: 't_user', headers: ['id', 'account', 'name', 'password', 'role', 'college', 'major', 'class_name', 'grade', 'avatar_url', 'is_active'] },
  { name: 't_student', headers: ['id', 'student_id', 'user_id', 'class_id', 'major_id', 'department_id', 'gpa', 'rank', 'total_students', 'alert_level', 'total_credits', 'required_credits'] },
  { name: 't_course', headers: ['id', 'code', 'name', 'credits', 'type', 'department_id'] },
  { name: 't_course_class', headers: ['id', 'course_id', 'term', 'class_name', 'teacher_id', 'max_students', 'enrolled_students', 'status'] },
  { name: 't_grade', headers: ['id', 'student_id', 'course_class_id', 'score', 'grade_point', 'status', 'term', 'exam_date'] },
  { name: 't_gpa_history', headers: ['id', 'student_id', 'term', 'gpa', 'avg_gpa', 'rank', 'total_students'] },
  { name: 't_profile_score', headers: ['id', 'student_id', 'term', 'dimension_key', 'label', 'score', 'avg_score', 'max_score', 'description', 'details'] },
  { name: 't_health_report', headers: ['id', 'student_id', 'term', 'total_score', 'items', 'report_url'] },
  { name: 't_alert', headers: ['id', 'student_id', 'level', 'type', 'title', 'description', 'course', 'failed_courses', 'trigger_date', 'status', 'suggestion', 'trigger_event', 'pushed_at', 'handler_id', 'handle_note', 'student_confirmed_at'] },
  { name: 't_intervention_record', headers: ['id', 'alert_id', 'teacher_id', 'intervention_date', 'methods', 'content', 'student_response', 'follow_up_plan'] },
]

function colName(index) {
  let n = index
  let s = ''
  while (n > 0) {
    const rem = (n - 1) % 26
    s = String.fromCharCode(65 + rem) + s
    n = Math.floor((n - 1) / 26)
  }
  return s
}

function addDataSheet(sheetName, headers, rows, options = {}) {
  const sheet = workbook.worksheets.add(sheetName)
  const lastRow = rows.length + 1
  const lastCol = headers.length
  const fullRange = `A1:${colName(lastCol)}${lastRow}`
  sheet.getRange(fullRange).values = [headers, ...rows.map((r) => headers.map((h) => r[h] ?? null))]
  sheet.showGridLines = false

  const headerRange = sheet.getRange(`A1:${colName(lastCol)}1`)
  headerRange.format = {
    fill: '#1F3864',
    font: { bold: true, color: '#FFFFFF' },
    horizontalAlignment: 'center',
  }
  sheet.getRange(`A2:${colName(lastCol)}${lastRow}`).format = {
    verticalAlignment: 'middle',
  }
  sheet.freezePanes.freezeRows(1)

  for (const [key, fmt] of Object.entries(options.numberFormats ?? {})) {
    const colIndex = headers.indexOf(key)
    if (colIndex >= 0) {
      sheet.getRange(`${colName(colIndex + 1)}2:${colName(colIndex + 1)}${lastRow}`).setNumberFormat(fmt)
    }
  }

  for (const [key, values] of Object.entries(options.validations ?? {})) {
    const colIndex = headers.indexOf(key)
    if (colIndex >= 0 && lastRow > 1) {
      sheet.getRange(`${colName(colIndex + 1)}2:${colName(colIndex + 1)}${lastRow}`).dataValidation = {
        rule: { type: 'list', values },
      }
    }
  }

  const tableName = `Table_${sheetName.replace(/[^a-zA-Z0-9_]/g, '_')}`
  sheet.tables.add(`A1:${colName(lastCol)}${lastRow}`, true, tableName)
  sheet.getRange(`A1:${colName(lastCol)}${Math.min(lastRow, 200)}`).format.autofitColumns()
  return sheet
}

const sheetNameToRows = {
  t_department: departments,
  t_major: majors,
  t_teacher: teachers,
  t_class: classes,
  t_user: users,
  t_student: students,
  t_course: courses,
  t_course_class: courseClasses,
  t_grade: grades,
  t_gpa_history: gpaHistory,
  t_profile_score: profileScores,
  t_health_report: healthReports,
  t_alert: alerts,
  t_intervention_record: interventions,
}

for (const def of sheetDefs) {
  addDataSheet(def.name, def.headers, sheetNameToRows[def.name], {
    numberFormats: {
      gpa: '0.00',
      avg_gpa: '0.00',
      score: '0.0',
      avg_score: '0',
      total_score: '0',
      credits: '0.0',
      total_credits: '0.0',
      required_credits: '0.0',
      grade_point: '0.0',
      exam_date: 'yyyy-mm-dd',
      intervention_date: 'yyyy-mm-dd',
      trigger_date: 'yyyy-mm-dd',
    },
    validations: {
      role: ['student', 'teacher', 'course_teacher', 'department', 'dean'],
      alert_level: ['none', 'yellow', 'orange', 'red'],
      level: ['yellow', 'orange', 'red'],
      status: ['pending', 'processing', 'resolved', 'active', 'finished', 'passed', 'failed', 'retake'],
      dimension_key: ['academic', 'ability', 'psychology', 'health', 'thought'],
      type: ['required', 'elective', 'public'],
    },
  })
}

// 字段字典
const dictRows = sheetDefs.flatMap((def) =>
  def.headers.map((h) => ({
    table: def.name,
    field: h,
    required: '视业务需要',
    type: h.includes('date') ? 'DATE' : h.includes('at') ? 'DATETIME/TEXT' : h.includes('score') || h.includes('gpa') || h.includes('credit') ? 'DECIMAL' : h === 'id' || h.endsWith('_id') ? 'BIGINT' : h.includes('json') || h.includes('methods') || h.includes('details') || h.includes('items') ? 'JSON/TEXT' : 'VARCHAR',
    note: '',
  })),
)
dictSheet.showGridLines = false
dictSheet.getRange(`A1:E${dictRows.length + 1}`).values = [
  ['表名', '字段', '必填', '类型', '说明'],
  ...dictRows.map((r) => [r.table, r.field, r.required, r.type, r.note]),
]
dictSheet.getRange('A1:E1').format = { fill: '#1F3864', font: { bold: true, color: '#FFFFFF' }, horizontalAlignment: 'center' }
dictSheet.freezePanes.freezeRows(1)
dictSheet.getRange('A1:E1').format.autofitColumns()

// 说明页
readmeSheet.showGridLines = false
readmeSheet.getRange('A1:B8').values = [
  ['Academic Navigation 样本数据包', ''],
  ['生成参数', `students=${studentCount}, classes=${classCount}, terms=${termCount}, seed=${seed}`],
  ['ID 偏移', String(idOffset)],
  ['首个学期', terms[0]],
  ['最近学期', latestTerm],
  ['密码', '统一 123456（BCrypt 哈希已写入 t_user）'],
  ['导入建议', '导入独立测试库或先清空目标表；SQL 文件 sample_data.sql 与 Excel 数据一致'],
  ['', ''],
]
readmeSheet.getRange('A1:B1').format = { fill: '#1F3864', font: { bold: true, color: '#FFFFFF' }, fontSize: 14 }
readmeSheet.getRange('A1:B1').merge()
readmeSheet.getRange('A2:B7').format = { verticalAlignment: 'middle' }
readmeSheet.getRange('A2:A7').format = { font: { bold: true } }

const summaryHeaders = ['数据表', '行数']
const summaryRows = sheetDefs.map((def, i) => [
  def.name,
  `=COUNTA('${def.name}'!A2:A${sheetNameToRows[def.name].length + 2})`,
])
readmeSheet.getRange(`A9:B${9 + summaryRows.length}`).values = [summaryHeaders, ...summaryRows]
readmeSheet.getRange('A9:B9').format = { fill: '#D9E2F3', font: { bold: true } }
readmeSheet.getRange('A1:B16').format.autofitColumns()

const xlsxPath = path.join(outDir, 'sample_data.xlsx')
const output = await SpreadsheetFile.exportXlsx(workbook)
await output.save(xlsxPath)
console.log(`Excel written: ${xlsxPath}`)

// 渲染预览（供本地检查）
const previewDir = path.join(outDir, 'preview')
await fs.mkdir(previewDir, { recursive: true })
for (const sheetName of ['说明', 't_student', 't_grade', 't_gpa_history', 't_profile_score', 't_alert']) {
  try {
    const blob = await workbook.render({ sheetName, range: 'A1:H16', autoCrop: 'all', scale: 1, format: 'png' })
    await fs.writeFile(path.join(previewDir, `${sheetName}.png`), new Uint8Array(await blob.arrayBuffer()))
  } catch (err) {
    console.warn(`preview skip ${sheetName}: ${err.message}`)
  }
}
console.log(`Preview written: ${previewDir}`)
