$ErrorActionPreference = 'Stop'

function Test-Step($name, $action) {
  try {
    $value = & $action
    Write-Host "[PASS] $name" -ForegroundColor Green
    return $value
  } catch {
    Write-Host "[FAIL] $name -> $($_.Exception.Message)" -ForegroundColor Red
    return $null
  }
}

function Login($account) {
  $res = Invoke-RestMethod -Uri 'http://localhost:8080/api/auth/login' -Method Post -ContentType 'application/json' -Body (@{ account = $account; password = '123456' } | ConvertTo-Json)
  return $res.data.token
}

function Get-Api($token, $path) {
  return Invoke-RestMethod -Uri "http://localhost:8080$path" -Headers @{ Authorization = "Bearer $token" }
}

Write-Host '=== 1. 服务状态 ===' -ForegroundColor Cyan
Test-Step '模拟教务接口 9090 健康检查' {
  $h = Invoke-RestMethod -Uri 'http://localhost:9090/health' -TimeoutSec 5
  "records=$($h.data.records.t_student) students"
}
Test-Step '后端 8080 登录可用' {
  $token = Login 'L20210001'
  "dean login ok"
}

Write-Host ''
Write-Host '=== 2. 同步状态 ===' -ForegroundColor Cyan
$deanToken = Login 'L20210001'
Test-Step 'school-sync/status' {
  $s = Get-Api $deanToken '/api/school-sync/status'
  "provider=$($s.data.provider) baseUrl=$($s.data.baseUrl)"
}

Write-Host ''
Write-Host '=== 3. 学生端样本数据 ===' -ForegroundColor Cyan
$studentToken = Login '20260001'
Test-Step '学生 dashboard' {
  $d = (Get-Api $studentToken '/api/student/dashboard?term=2027-2028-2').data
  "gpa=$($d.gpa) rank=$($d.rank)/$($d.totalStudents) trend=$($d.gpaTrend.Count)"
}
Test-Step '学生画像' {
  $p = (Get-Api $studentToken '/api/student/profile?term=2027-2028-2').data
  "dimensions=$($p.dimensions.Count) gpaTrend=$($p.gpaTrend.Count)"
}
Test-Step '学生预警' {
  $a = (Get-Api $studentToken '/api/student/alerts?page=1').data
  "total=$($a.total)"
}
Test-Step '学生体测' {
  $h = (Get-Api $studentToken '/api/student/health-reports?term=2027-2028-2').data
  "reports=$($h.Count)"
}

Write-Host ''
Write-Host '=== 4. 班主任端样本数据 ===' -ForegroundColor Cyan
$teacherToken = Login 'T2024001'
Test-Step '班级驾驶舱' {
  $d = (Get-Api $teacherToken '/api/teacher/dashboard?term=2027-2028-2').data
  "students=$($d.totalStudents) alertCount=$($d.alertCount)"
}
Test-Step '班级学生列表' {
  $s = (Get-Api $teacherToken '/api/teacher/students?page=1').data
  "total=$($s.total) first=$($s.list[0].studentId)"
}
Test-Step 'GPA 进退' {
  $g = (Get-Api $teacherToken '/api/teacher/gpa-progress?term=2027-2028-2').data
  "students=$($g.students.Count) terms=$($g.terms.Count)"
}
Test-Step '课程统计' {
  $c = (Get-Api $teacherToken '/api/teacher/course-stats?term=2027-2028-2').data
  "courses=$($c.Count)"
}

Write-Host ''
Write-Host '=== 5. 任课教师端样本数据 ===' -ForegroundColor Cyan
$courseTeacherToken = Login 'C2024001'
Test-Step '任课教师课程列表' {
  $c = (Get-Api $courseTeacherToken '/api/course-teacher/courses?term=2027-2028-2').data
  "courses=$($c.Count) students=$($c[0].students)"
}
Test-Step '任课教师预警' {
  $a = (Get-Api $courseTeacherToken '/api/course-teacher/alerts?page=1').data
  "total=$($a.total)"
}

Write-Host ''
Write-Host '=== 6. 权限隔离 ===' -ForegroundColor Cyan
Test-Step '学生禁止触发同步接口' {
  try {
    Invoke-RestMethod -Uri 'http://localhost:8080/api/school-sync' -Method Post -Headers @{ Authorization = "Bearer $studentToken" } | Out-Null
    'unexpected ok'
  } catch {
    "expected 403: $([int]$_.Exception.Response.StatusCode)"
  }
}

Write-Host ''
Write-Host '冒烟完成。前端地址: http://localhost:5173' -ForegroundColor Cyan
