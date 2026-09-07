import http from 'node:http'
import fs from 'node:fs/promises'
import path from 'node:path'

const args = process.argv.slice(2)
function readArg(name, fallback) {
  const idx = args.indexOf(`--${name}`)
  return idx >= 0 && args[idx + 1] !== undefined ? args[idx + 1] : fallback
}

const port = Number(readArg('port', process.env.PORT || '9090'))
const token = process.env.SCHOOL_API_TOKEN || ''
const dataPath = path.resolve(readArg('data', process.env.SAMPLE_DATA_JSON || 'sample_data.json'))

const raw = await fs.readFile(dataPath, 'utf8')
const dataset = JSON.parse(raw)
const tables = dataset.tables

const studentsByNo = new Map()
for (const row of tables.t_student) studentsByNo.set(String(row.student_id), row)

const usersByAccount = new Map()
for (const row of tables.t_user) usersByAccount.set(String(row.account), row)

function ok(data) {
  return JSON.stringify({ code: 0, message: 'success', data })
}

function fail(code, message) {
  return JSON.stringify({ code, message, data: null })
}

function respond(res, status, body) {
  res.writeHead(status, { 'Content-Type': 'application/json; charset=utf-8' })
  res.end(body)
}

function pageList(list, page, size) {
  const p = Math.max(1, Number(page) || 1)
  const s = Math.max(1, Number(size) || 1000)
  const start = (p - 1) * s
  return { list: list.slice(start, start + s), total: list.length, page: p, size: s }
}

function authorize(req) {
  if (!token) return null
  const header = req.headers.authorization || ''
  return header === `Bearer ${token}` ? null : 'invalid token'
}

const server = http.createServer((req, res) => {
  res.setHeader('Content-Type', 'application/json; charset=utf-8')
  res.setHeader('Access-Control-Allow-Origin', '*')
  res.setHeader('Access-Control-Allow-Headers', 'Authorization, Content-Type')
  res.setHeader('Access-Control-Allow-Methods', 'GET, POST, OPTIONS')
  if (req.method === 'OPTIONS') {
    res.writeHead(204)
    res.end()
    return
  }

  const authError = authorize(req)
  if (authError) {
    respond(res, 401, fail(401, authError))
    return
  }

  const url = new URL(req.url, `http://${req.headers.host || 'localhost'}`)
  const segments = url.pathname.split('/').filter(Boolean)

  if (url.pathname === '/health') {
    const records = Object.fromEntries(Object.entries(tables).map(([k, v]) => [k, v.length]))
    respond(res, 200, ok({ status: 'up', records }))
    return
  }

  if (segments[0] === 'api' && segments[1] === 'v1') {
    const [entity, id, child] = segments.slice(2)
    if (entity === 'students' && id) {
      const student = studentsByNo.get(String(id))
      if (!student) {
        respond(res, 404, fail(404, 'student not found'))
        return
      }
      if (!child) {
        const user = usersByAccount.get(String(student.student_id))
        respond(res, 200, ok({ ...student, user }))
        return
      }
      let data
      if (child === 'grades') {
        const term = url.searchParams.get('term')
        data = tables.t_grade.filter((g) => String(g.student_id) === String(student.id) && (!term || g.term === term))
      } else if (child === 'gpa-history') {
        data = tables.t_gpa_history.filter((g) => String(g.student_id) === String(student.id))
      } else if (child === 'profile-scores') {
        const term = url.searchParams.get('term')
        data = tables.t_profile_score.filter((p) => String(p.student_id) === String(student.id) && (!term || p.term === term))
      } else if (child === 'health-reports') {
        data = tables.t_health_report.filter((h) => String(h.student_id) === String(student.id))
      } else if (child === 'alerts') {
        data = tables.t_alert.filter((a) => String(a.student_id) === String(student.id))
      } else {
        respond(res, 404, fail(404, `unknown student child: ${child}`))
        return
      }
      respond(res, 200, ok(data))
      return
    }

    switch (entity) {
      case 'departments':
        respond(res, 200, ok(tables.t_department))
        return
      case 'majors':
        respond(res, 200, ok(tables.t_major))
        return
      case 'classes':
        respond(res, 200, ok(tables.t_class))
        return
      case 'teachers':
        respond(res, 200, ok(tables.t_teacher))
        return
      case 'users':
        respond(res, 200, ok(tables.t_user))
        return
      case 'students': {
        const page = Number(url.searchParams.get('page') || '1')
        const size = Number(url.searchParams.get('size') || '1000')
        respond(res, 200, ok(pageList(tables.t_student, page, size)))
        return
      }
      case 'courses':
        respond(res, 200, ok(tables.t_course))
        return
      case 'course-classes': {
        const term = url.searchParams.get('term')
        respond(res, 200, ok(term ? tables.t_course_class.filter((c) => c.term === term) : tables.t_course_class))
        return
      }
      case 'alerts':
        respond(res, 200, ok(tables.t_alert))
        return
      case 'interventions':
        respond(res, 200, ok(tables.t_intervention_record))
        return
      default:
        break
    }
  }

  respond(res, 404, fail(404, 'not found'))
})

server.listen(port, () => {
  console.log(`[school-api-server] listening on http://localhost:${port}, data=${dataPath}`)
})
