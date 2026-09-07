-- V30: complete historical attendance snapshots and rebuild the three-level
-- warning trend with the current failed-course and absence rules.

DROP TEMPORARY TABLE IF EXISTS tmp_historical_population_v30;
DROP TEMPORARY TABLE IF EXISTS tmp_historical_failures_v30;
DROP TEMPORARY TABLE IF EXISTS tmp_historical_profile_v30;

CREATE TEMPORARY TABLE tmp_historical_population_v30 AS
SELECT DISTINCT g.student_id, g.term
FROM t_grade g
WHERE g.term IN ('2022-2023-2', '2023-2024-1', '2023-2024-2');

ALTER TABLE tmp_historical_population_v30 ADD PRIMARY KEY (student_id, term);

-- Historical attendance was previously absent. Seed stable demo snapshots:
-- about 4% severe absence, 8% three absences, 48% one or two absences.
INSERT INTO t_attendance (
    student_id, term, total_classes, absent_count, late_count, note
)
SELECT p.student_id,
       p.term,
       60,
       CASE
           WHEN MOD(CRC32(CONCAT(s.student_id, '-', p.term, '-history-alert')), 100) < 4
               THEN 4 + MOD(CRC32(CONCAT(s.student_id, '-', p.term, '-red-absence')), 3)
           WHEN MOD(CRC32(CONCAT(s.student_id, '-', p.term, '-history-alert')), 100) < 12
               THEN 3
           WHEN MOD(CRC32(CONCAT(s.student_id, '-', p.term, '-history-alert')), 100) < 60
               THEN 1 + MOD(CRC32(CONCAT(s.student_id, '-', p.term, '-yellow-absence')), 2)
           ELSE 0
       END,
       MOD(CRC32(CONCAT(s.student_id, '-', p.term, '-history-late')), 4),
       '历史趋势演示考勤数据（按新预警标准生成）'
FROM tmp_historical_population_v30 p
JOIN t_student s ON s.id = p.student_id
ON DUPLICATE KEY UPDATE
    total_classes = VALUES(total_classes),
    absent_count = VALUES(absent_count),
    late_count = VALUES(late_count),
    note = VALUES(note);

CREATE TEMPORARY TABLE tmp_historical_failures_v30 AS
SELECT g.student_id,
       g.term,
       COUNT(*) AS failed_count,
       JSON_ARRAYAGG(c.name) AS failed_courses,
       MIN(c.name) AS first_course
FROM t_grade g
JOIN t_course_class cc ON cc.id = g.course_class_id
JOIN t_course c ON c.id = cc.course_id
WHERE g.term IN ('2022-2023-2', '2023-2024-1', '2023-2024-2')
  AND (g.status = 'failed' OR g.score < 60)
GROUP BY g.student_id, g.term;

ALTER TABLE tmp_historical_failures_v30 ADD PRIMARY KEY (student_id, term);

CREATE TEMPORARY TABLE tmp_historical_profile_v30 AS
SELECT p.student_id,
       p.term,
       COALESCE(f.failed_count, 0) AS failed_count,
       COALESCE(f.failed_courses, JSON_ARRAY()) AS failed_courses,
       f.first_course,
       a.absent_count,
       CASE
           WHEN COALESCE(f.failed_count, 0) >= 2 OR a.absent_count >= 4 THEN 'red'
           WHEN COALESCE(f.failed_count, 0) = 1 OR a.absent_count = 3 THEN 'orange'
           WHEN COALESCE(f.failed_count, 0) = 0 AND a.absent_count BETWEEN 1 AND 2 THEN 'yellow'
           ELSE 'none'
       END AS level
FROM tmp_historical_population_v30 p
JOIN t_attendance a ON a.student_id = p.student_id AND a.term = p.term
LEFT JOIN tmp_historical_failures_v30 f
       ON f.student_id = p.student_id AND f.term = p.term;

ALTER TABLE tmp_historical_profile_v30 ADD PRIMARY KEY (student_id, term);

DELETE FROM t_alert
WHERE status = 'resolved'
  AND trigger_date IN (DATE '2023-07-10', DATE '2024-01-15', DATE '2024-07-10');

INSERT INTO t_alert (
    student_id, level, type, title, description, course, failed_courses,
    trigger_date, status, suggestion, trigger_event, pushed_at
)
SELECT p.student_id,
       p.level,
       CASE
           WHEN p.failed_count > 0 AND p.absent_count > 0 THEN '复合风险预警'
           WHEN p.failed_count > 0 THEN '课程预警'
           ELSE '出勤预警'
       END,
       CONCAT(
           CASE p.level WHEN 'red' THEN '红色预警：'
                        WHEN 'orange' THEN '橙色预警：'
                        ELSE '黄色预警：' END,
           CASE
               WHEN p.failed_count >= 2 THEN '挂科两门及以上'
               WHEN p.absent_count >= 4 THEN '缺勤四次及以上'
               WHEN p.failed_count = 1 THEN '挂科一门'
               WHEN p.absent_count = 3 THEN '缺勤三次'
               ELSE '缺勤一到两次且无挂科'
           END
       ),
       CONCAT('该学期挂科 ', p.failed_count, ' 门，缺勤 ', p.absent_count, ' 次，已完成闭环处理。'),
       p.first_course,
       p.failed_courses,
       CASE p.term
           WHEN '2022-2023-2' THEN DATE '2023-07-10'
           WHEN '2023-2024-1' THEN DATE '2024-01-15'
           ELSE DATE '2024-07-10'
       END,
       'resolved',
       '历史预警已完成约谈、帮扶和闭环复查',
       CONCAT('挂科=', p.failed_count, '；缺勤=', p.absent_count),
       CASE p.term
           WHEN '2022-2023-2' THEN TIMESTAMP '2023-07-10 10:00:00'
           WHEN '2023-2024-1' THEN TIMESTAMP '2024-01-15 10:00:00'
           ELSE TIMESTAMP '2024-07-10 10:00:00'
       END
FROM tmp_historical_profile_v30 p
WHERE p.level <> 'none';

DROP TEMPORARY TABLE IF EXISTS tmp_historical_profile_v30;
DROP TEMPORARY TABLE IF EXISTS tmp_historical_failures_v30;
DROP TEMPORARY TABLE IF EXISTS tmp_historical_population_v30;
