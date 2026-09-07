-- 补齐每个已有成绩学期的学生端基础数据与五维画像。
-- 学生-学期范围始终取自真实成绩表，避免为尚未入学的年级生成历史学期。

DROP TEMPORARY TABLE IF EXISTS tmp_student_terms;
CREATE TEMPORARY TABLE tmp_student_terms AS
SELECT DISTINCT student_id, term
FROM t_grade;
ALTER TABLE tmp_student_terms ADD PRIMARY KEY (student_id, term);

INSERT IGNORE INTO t_health_report (student_id, term, total_score, items, report_url)
SELECT st.student_id,
       st.term,
       60 + MOD(CRC32(CONCAT(st.student_id, '-', st.term)), 36),
       JSON_ARRAY(
           JSON_OBJECT('name', '体重指数', 'score', 60 + MOD(CRC32(CONCAT('bmi-', st.student_id, '-', st.term)), 36), 'level', '良好'),
           JSON_OBJECT('name', '肺活量', 'score', 60 + MOD(CRC32(CONCAT('lung-', st.student_id, '-', st.term)), 36), 'level', '良好'),
           JSON_OBJECT('name', '耐力跑', 'score', 60 + MOD(CRC32(CONCAT('run-', st.student_id, '-', st.term)), 36), 'level', '良好')
       ),
       NULL
FROM tmp_student_terms st;

INSERT IGNORE INTO t_attendance (student_id, term, total_classes, absent_count, late_count, note)
SELECT st.student_id,
       st.term,
       60,
       CASE
           WHEN MOD(CRC32(CONCAT('absence-', st.student_id, '-', st.term)), 17) = 0
               THEN 4 + MOD(CRC32(CONCAT('absence-high-', st.student_id, '-', st.term)), 4)
           ELSE MOD(CRC32(CONCAT('absence-low-', st.student_id, '-', st.term)), 4)
       END,
       MOD(CRC32(CONCAT('late-', st.student_id, '-', st.term)), 4),
       '学期考勤汇总'
FROM tmp_student_terms st;

INSERT IGNORE INTO t_volunteer (student_id, term, hours, description)
SELECT st.student_id,
       st.term,
       6 + MOD(CRC32(CONCAT('volunteer-', st.student_id, '-', st.term)), 55),
       '学期志愿服务汇总'
FROM tmp_student_terms st;

INSERT IGNORE INTO t_psychology (student_id, term, score, level, note)
SELECT st.student_id,
       st.term,
       62 + MOD(CRC32(CONCAT('psychology-', st.student_id, '-', st.term)), 36),
       CASE
           WHEN 62 + MOD(CRC32(CONCAT('psychology-', st.student_id, '-', st.term)), 36) >= 85 THEN 'good'
           WHEN 62 + MOD(CRC32(CONCAT('psychology-', st.student_id, '-', st.term)), 36) >= 70 THEN 'normal'
           ELSE 'attention'
       END,
       '学期测评记录'
FROM tmp_student_terms st;

INSERT INTO t_competition (student_id, term, competition_name, level, award, points)
SELECT st.student_id,
       st.term,
       '学科专业能力竞赛',
       CASE MOD(CRC32(CONCAT('competition-level-', st.student_id, '-', st.term)), 3)
           WHEN 0 THEN 'school'
           WHEN 1 THEN 'provincial'
           ELSE 'national'
       END,
       CASE MOD(CRC32(CONCAT('competition-award-', st.student_id, '-', st.term)), 4)
           WHEN 0 THEN 'first'
           WHEN 1 THEN 'second'
           WHEN 2 THEN 'third'
           ELSE 'participation'
       END,
       64 + MOD(CRC32(CONCAT('competition-points-', st.student_id, '-', st.term)), 30)
FROM tmp_student_terms st
WHERE MOD(CRC32(CONCAT('competition-', st.student_id, '-', st.term)), 4) = 0
  AND NOT EXISTS (
      SELECT 1 FROM t_competition c
      WHERE c.student_id = st.student_id AND c.term = st.term
  );

-- 学业成绩：按当学期 GPA、专业排名、课程均分和挂科数计算。
INSERT IGNORE INTO t_profile_score
    (student_id, term, dimension_key, label, score, avg_score, max_score, description, details)
SELECT gh.student_id,
       gh.term,
       'academic',
       '学业成绩',
       LEAST(100, GREATEST(0, ROUND(
           35 + gh.gpa / 4 * 40
           + (1 - gh.`rank` / NULLIF(gh.total_students, 0)) * 12
           + gs.avg_score / 100 * 8
           - gs.failed_count * 5
       ))),
       72,
       100,
       '基于当学期GPA、排名、挂科数与课程均分计算',
       JSON_ARRAY(
           CONCAT('GPA: ', FORMAT(gh.gpa, 2)),
           CONCAT('专业排名: ', gh.`rank`, ' / ', gh.total_students),
           CONCAT('挂科门数: ', gs.failed_count),
           CONCAT('课程均分: ', ROUND(gs.avg_score))
       )
FROM t_gpa_history gh
JOIN (
    SELECT student_id,
           term,
           AVG(score) avg_score,
           SUM(score < 60 OR status = 'failed') failed_count
    FROM t_grade
    GROUP BY student_id, term
) gs ON gs.student_id = gh.student_id AND gs.term = gh.term;

-- 实践能力：按当学期竞赛记录计算，无竞赛时保留基础分。
INSERT IGNORE INTO t_profile_score
    (student_id, term, dimension_key, label, score, avg_score, max_score, description, details)
SELECT st.student_id,
       st.term,
       'practice',
       '实践能力',
       CASE
           WHEN COALESCE(cs.max_points, 0) >= 90 THEN 90
           WHEN COALESCE(cs.max_points, 0) >= 80 THEN 82
           WHEN COALESCE(cs.max_points, 0) >= 70 THEN 74
           WHEN COALESCE(cs.item_count, 0) > 0 THEN 66
           ELSE 55
       END,
       68,
       100,
       '基于当学期竞赛经历与实践积分计算',
       JSON_ARRAY(
           CONCAT('竞赛积分: ', COALESCE(cs.max_points, 0)),
           CONCAT('竞赛记录: ', COALESCE(cs.item_count, 0), ' 项')
       )
FROM tmp_student_terms st
LEFT JOIN (
    SELECT student_id, term, MAX(points) max_points, COUNT(*) item_count
    FROM t_competition
    GROUP BY student_id, term
) cs ON cs.student_id = st.student_id AND cs.term = st.term;

-- 综合素质：按当学期志愿服务和出勤计算。
INSERT IGNORE INTO t_profile_score
    (student_id, term, dimension_key, label, score, avg_score, max_score, description, details)
SELECT st.student_id,
       st.term,
       'quality',
       '综合素质',
       CASE
           WHEN v.hours >= 40 THEN 92
           WHEN v.hours >= 20 THEN 85
           WHEN v.hours >= 8 THEN 78
           WHEN v.hours > 0 THEN 70
           ELSE 62
       END,
       85,
       100,
       '基于当学期志愿服务时长与出勤情况计算',
       JSON_ARRAY(
           CONCAT('志愿时长: ', v.hours, ' 小时'),
           CONCAT('出勤率: ', GREATEST(0, 100 - ROUND(a.absent_count * 100 / NULLIF(a.total_classes, 0))), '%')
       )
FROM tmp_student_terms st
JOIN t_volunteer v ON v.student_id = st.student_id AND v.term = st.term
JOIN t_attendance a ON a.student_id = st.student_id AND a.term = st.term;

-- 人文素养：沿用现有数据模型中的学期综合测评结果。
INSERT IGNORE INTO t_profile_score
    (student_id, term, dimension_key, label, score, avg_score, max_score, description, details)
SELECT st.student_id,
       st.term,
       'culture',
       '人文素养',
       CASE WHEN p.score >= 85 THEN 90 WHEN p.score >= 70 THEN 78 ELSE 68 END,
       75,
       100,
       '基于当学期综合测评结果计算',
       JSON_ARRAY(CONCAT('综合测评: ', p.score, ' 分'))
FROM tmp_student_terms st
JOIN t_psychology p ON p.student_id = st.student_id AND p.term = st.term;

-- 身心健康：严格使用同一学期体测报告，不沿用最新学期数据。
INSERT IGNORE INTO t_profile_score
    (student_id, term, dimension_key, label, score, avg_score, max_score, description, details)
SELECT st.student_id,
       st.term,
       'health',
       '身心健康',
       h.total_score,
       80,
       100,
       '基于当学期体测总分计算',
       JSON_ARRAY(CONCAT('体测总分: ', h.total_score, ' 分'))
FROM tmp_student_terms st
JOIN t_health_report h ON h.student_id = st.student_id AND h.term = st.term;

DROP TEMPORARY TABLE IF EXISTS tmp_student_terms;
