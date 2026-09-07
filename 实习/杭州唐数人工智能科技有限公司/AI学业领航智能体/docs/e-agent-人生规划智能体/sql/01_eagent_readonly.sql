-- E-Agent 人生规划智能体：只读数据库账号与视图
-- 执行前请替换强密码，并确认当前 MySQL 用户有 CREATE USER / CREATE VIEW 权限。

-- ============ 1. 只读账号 ============

CREATE USER IF NOT EXISTS 'eagent_readonly'@'%'
  IDENTIFIED BY '请替换为高强度密码';

CREATE USER IF NOT EXISTS 'eagent_readonly'@'localhost'
  IDENTIFIED BY '请替换为高强度密码';

-- 清理旧权限，仅保留视图只读权限。
REVOKE ALL PRIVILEGES, GRANT OPTION FROM 'eagent_readonly'@'%';
REVOKE ALL PRIVILEGES, GRANT OPTION FROM 'eagent_readonly'@'localhost';

-- ============ 2. 当前学生档案 ============

CREATE OR REPLACE VIEW v_agent_student_profile AS
SELECT
    s.id AS student_pk,
    s.student_id AS student_no,
    u.name AS student_name,
    d.name AS department_name,
    m.name AS major_name,
    m.degree_type,
    c.name AS class_name,
    c.grade,
    s.gpa,
    s.`rank`,
    s.total_students,
    s.total_credits,
    s.required_credits,
    s.alert_level
FROM t_student s
JOIN t_user u ON u.id = s.user_id
LEFT JOIN t_class c ON c.id = s.class_id
LEFT JOIN t_major m ON m.id = s.major_id
LEFT JOIN t_department d ON d.id = s.department_id;

-- ============ 3. GPA 与排名趋势 ============

CREATE OR REPLACE VIEW v_agent_gpa_trend AS
SELECT
    h.student_id,
    h.term,
    h.gpa,
    h.avg_gpa,
    h.`rank`,
    h.total_students
FROM t_gpa_history h
ORDER BY h.student_id, h.term;

-- ============ 4. 未闭环预警 ============

CREATE OR REPLACE VIEW v_agent_recent_alerts AS
SELECT
    a.id AS alert_id,
    a.student_id,
    a.level,
    a.type,
    a.title,
    a.description,
    a.course,
    a.failed_courses,
    a.trigger_date,
    a.status,
    a.suggestion
FROM t_alert a
WHERE a.status IN ('pending', 'processing')
ORDER BY a.trigger_date DESC, a.id DESC;

-- ============ 5. 最近课程成绩 ============

CREATE OR REPLACE VIEW v_agent_recent_grades AS
SELECT
    g.id AS grade_id,
    g.student_id,
    g.term,
    c.code AS course_code,
    c.name AS course_name,
    c.credits,
    g.score,
    g.grade_point,
    g.status
FROM t_grade g
JOIN t_course_class cc ON cc.id = g.course_class_id
JOIN t_course c ON c.id = cc.course_id;

-- ============ 6. 人生规划综合快照 ============

CREATE OR REPLACE VIEW v_agent_development_snapshot AS
SELECT
    s.id AS student_pk,
    s.student_id AS student_no,
    s.gpa,
    s.`rank`,
    s.total_students,
    p.dimension_key,
    p.label AS dimension_label,
    p.score,
    p.avg_score,
    p.description,
    a.absent_count,
    a.late_count,
    psy.score AS psychology_score,
    psy.level AS psychology_level,
    v.hours AS volunteer_hours,
    COUNT(DISTINCT comp.id) AS competition_count
FROM t_student s
LEFT JOIN t_profile_score p
  ON p.student_id = s.id
 AND p.term = (
      SELECT h.term
      FROM t_gpa_history h
      WHERE h.student_id = s.id
      ORDER BY h.term DESC
      LIMIT 1
 )
LEFT JOIN t_attendance a
  ON a.student_id = s.id
 AND a.term = (
      SELECT h.term
      FROM t_gpa_history h
      WHERE h.student_id = s.id
      ORDER BY h.term DESC
      LIMIT 1
 )
LEFT JOIN t_psychology psy
  ON psy.student_id = s.id
 AND psy.term = (
      SELECT h.term
      FROM t_gpa_history h
      WHERE h.student_id = s.id
      ORDER BY h.term DESC
      LIMIT 1
 )
LEFT JOIN t_volunteer v
  ON v.student_id = s.id
 AND v.term = (
      SELECT h.term
      FROM t_gpa_history h
      WHERE h.student_id = s.id
      ORDER BY h.term DESC
      LIMIT 1
 )
LEFT JOIN t_competition comp
  ON comp.student_id = s.id
 AND comp.term = (
      SELECT h.term
      FROM t_gpa_history h
      WHERE h.student_id = s.id
      ORDER BY h.term DESC
      LIMIT 1
 )
GROUP BY
    s.id,
    s.student_id,
    s.gpa,
    s.`rank`,
    s.total_students,
    p.dimension_key,
    p.label,
    p.score,
    p.avg_score,
    p.description,
    a.absent_count,
    a.late_count,
    psy.score,
    psy.level,
    v.hours;

-- ============ 7. 专业培养方案 ============

CREATE OR REPLACE VIEW v_agent_major_plan AS
SELECT
    mp.id AS plan_id,
    m.id AS major_id,
    m.name AS major_name,
    mp.grade,
    mp.dimensions,
    mp.standard,
    mp.actual,
    mp.tips
FROM t_major_plan mp
JOIN t_major m ON m.id = mp.major_id;

-- ============ 8. 就业岗位库 ============

CREATE OR REPLACE VIEW v_agent_jobs AS
SELECT
    j.id,
    j.company_name,
    j.company_size,
    j.company_stage,
    j.job_title,
    j.salary_range,
    j.city,
    j.education,
    j.experience,
    j.tags,
    j.highlights,
    j.match_score,
    j.match_reasons,
    j.category,
    j.publish_date,
    j.source_url
FROM t_job_cache j;

-- ============ 9. 考研 / 留学院校库 ============

CREATE OR REPLACE VIEW v_agent_schools AS
SELECT
    sc.id,
    sc.type,
    sc.name,
    sc.name_zh,
    sc.country,
    sc.location,
    sc.`rank`,
    sc.tier,
    sc.programs,
    sc.admission_gpa,
    sc.exam_requirements,
    sc.language_requirement,
    sc.highlights,
    sc.match_score,
    sc.match_reasons,
    sc.official_url,
    sc.deadline
FROM t_school_cache sc;

-- ============ 10. 授予视图只读权限 ============
-- 生产环境应将 '%' 替换为 E-Agent 后端服务器的固定 IP。

GRANT SELECT ON academic_nav.v_agent_student_profile TO 'eagent_readonly'@'%';
GRANT SELECT ON academic_nav.v_agent_gpa_trend TO 'eagent_readonly'@'%';
GRANT SELECT ON academic_nav.v_agent_recent_alerts TO 'eagent_readonly'@'%';
GRANT SELECT ON academic_nav.v_agent_recent_grades TO 'eagent_readonly'@'%';
GRANT SELECT ON academic_nav.v_agent_development_snapshot TO 'eagent_readonly'@'%';
GRANT SELECT ON academic_nav.v_agent_major_plan TO 'eagent_readonly'@'%';
GRANT SELECT ON academic_nav.v_agent_jobs TO 'eagent_readonly'@'%';
GRANT SELECT ON academic_nav.v_agent_schools TO 'eagent_readonly'@'%';

GRANT SELECT ON academic_nav.v_agent_student_profile TO 'eagent_readonly'@'localhost';
GRANT SELECT ON academic_nav.v_agent_gpa_trend TO 'eagent_readonly'@'localhost';
GRANT SELECT ON academic_nav.v_agent_recent_alerts TO 'eagent_readonly'@'localhost';
GRANT SELECT ON academic_nav.v_agent_recent_grades TO 'eagent_readonly'@'localhost';
GRANT SELECT ON academic_nav.v_agent_development_snapshot TO 'eagent_readonly'@'localhost';
GRANT SELECT ON academic_nav.v_agent_major_plan TO 'eagent_readonly'@'localhost';
GRANT SELECT ON academic_nav.v_agent_jobs TO 'eagent_readonly'@'localhost';
GRANT SELECT ON academic_nav.v_agent_schools TO 'eagent_readonly'@'localhost';

FLUSH PRIVILEGES;
