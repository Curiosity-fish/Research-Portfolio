param(
    [string]$MysqlHost = "127.0.0.1",
    [string]$MysqlUser = "root",
    [string]$MysqlDatabase = "academic_nav",
    [string]$OutDir = "outputs\eagent-formal-data"
)

$ErrorActionPreference = "Stop"
$mysql = "C:\Program Files\MySQL\MySQL Server 8.0\bin\mysql.exe"

if (-not (Test-Path $mysql)) {
    throw "MySQL client not found at $mysql"
}

if (-not $env:MYSQL_PWD) {
    $secure = Read-Host "Enter MySQL password" -AsSecureString
    $env:MYSQL_PWD = [System.Runtime.InteropServices.Marshal]::PtrToStringUni(
        [System.Runtime.InteropServices.Marshal]::SecureStringToBSTR($secure)
    )
}

$outDir = Join-Path -Path (Resolve-Path ".").Path -ChildPath $OutDir
$tablesDir = Join-Path $outDir "tables"
$knowledgeDir = Join-Path $outDir "knowledge_package"
New-Item -ItemType Directory -Force -Path $tablesDir, $knowledgeDir | Out-Null

$tables = @(
    "t_department",
    "t_major",
    "t_class",
    "t_teacher",
    "t_student",
    "t_course",
    "t_course_class",
    "t_grade",
    "t_gpa_history",
    "t_profile_score",
    "t_health_report",
    "t_alert",
    "t_attendance",
    "t_psychology",
    "t_volunteer",
    "t_competition",
    "t_intervention_record",
    "t_job_cache",
    "t_school_cache",
    "t_major_plan"
)

$counts = @{}

foreach ($table in $tables) {
    if ($table -eq "t_psychology") {
        $sql = "SELECT id, student_id, term, score, level FROM t_psychology"
    } elseif ($table -eq "t_health_report") {
        $sql = "SELECT id, student_id, term, total_score, items FROM t_health_report"
    } else {
        $sql = "SELECT * FROM $table"
    }
    $outputFile = Join-Path $tablesDir "$table.tsv"
    & $mysql --host=$MysqlHost --user=$MysqlUser --database=$MysqlDatabase --default-character-set=utf8mb4 --batch --raw --execute=$sql |
        Out-File -FilePath $outputFile -Encoding utf8

    $countSql = "SELECT COUNT(*) FROM $table"
    $countText = & $mysql --host=$MysqlHost --user=$MysqlUser --database=$MysqlDatabase --batch --skip-column-names --execute=$countSql
    $counts[$table] = [int](($countText | Select-Object -First 1) -replace '\D', '')
}

$userPublicSql = "SELECT id, account, name, role, college, major, class_name, grade, is_active FROM t_user ORDER BY id"
& $mysql --host=$MysqlHost --user=$MysqlUser --database=$MysqlDatabase --default-character-set=utf8mb4 --batch --raw --execute=$userPublicSql |
    Out-File -FilePath (Join-Path $tablesDir "t_user_public.tsv") -Encoding utf8

$userCountSql = "SELECT COUNT(*) FROM t_user"
$userCountText = & $mysql --host=$MysqlHost --user=$MysqlUser --database=$MysqlDatabase --batch --skip-column-names --execute=$userCountSql
$counts["t_user_public"] = [int](($userCountText | Select-Object -First 1) -replace '\D', '')

$knowledgeSource = Join-Path -Path (Resolve-Path ".").Path -ChildPath "docs\e-agent-人生规划智能体\knowledge"
if (Test-Path $knowledgeSource) {
    Get-ChildItem -Path $knowledgeSource -Filter "*.md" |
        Copy-Item -Destination $knowledgeDir -Force
}

$manifest = [ordered]@{
    generated_at = (Get-Date).ToString("yyyy-MM-dd HH:mm:ss")
    source = [ordered]@{
        type = "mysql"
        host = $MysqlHost
        database = $MysqlDatabase
    }
    tables = $counts
    sensitive_fields_excluded = @(
        "t_user.password",
        "t_user.avatar_url",
        "t_psychology.note",
        "t_health_report.report_url"
    )
    usage = @(
        "tables/ for offline validation and formal handoff",
        "knowledge_package/ for E-Agent document and QA knowledge base",
        "Do not upload live student tables to E-Agent; use the backend API tools instead."
    )
}

$manifest | ConvertTo-Json -Depth 5 | Out-File -FilePath (Join-Path $outDir "manifest.json") -Encoding utf8
Write-Host "Formal data package written to $outDir"
