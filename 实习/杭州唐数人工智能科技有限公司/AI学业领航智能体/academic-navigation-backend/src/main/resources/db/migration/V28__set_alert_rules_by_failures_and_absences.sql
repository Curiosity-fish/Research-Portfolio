-- V28: use the new warning standard.
-- Yellow: 1-2 absences and no failed course.
-- Orange: 1 failed course or 3 absences.
-- Red: 2+ failed courses or 4+ absences.
-- GPA and psychology scores are not used for alert grading.

DROP TEMPORARY TABLE IF EXISTS tmp_failed_courses_v28;
DROP TEMPORARY TABLE IF EXISTS tmp_absences_v28;
DROP TEMPORARY TABLE IF EXISTS tmp_alert_level_v28;

CREATE TEMPORARY TABLE tmp_failed_courses_v28 AS
SELECT g.student_id,
       COUNT(*) AS failed_count,
       JSON_ARRAYAGG(c.name) AS failed_courses,
       MIN(c.name) AS first_course
FROM t_grade g
JOIN t_course_class cc ON cc.id = g.course_class_id
JOIN t_course c ON c.id = cc.course_id
WHERE g.term = '2024-2025-1'
  AND (g.status = 'failed' OR g.score < 60)
GROUP BY g.student_id;

ALTER TABLE tmp_failed_courses_v28 ADD PRIMARY KEY (student_id);

CREATE TEMPORARY TABLE tmp_absences_v28 AS
SELECT student_id, SUM(COALESCE(absent_count, 0)) AS absences
FROM t_attendance
WHERE term = '2024-2025-1'
GROUP BY student_id;

ALTER TABLE tmp_absences_v28 ADD PRIMARY KEY (student_id);

CREATE TEMPORARY TABLE tmp_alert_level_v28 AS
SELECT s.id AS student_id,
       COALESCE(f.failed_count, 0) AS failed_count,
       COALESCE(f.failed_courses, JSON_ARRAY()) AS failed_courses,
       f.first_course,
       COALESCE(a.absences, 0) AS absences,
       CASE
           WHEN COALESCE(f.failed_count, 0) >= 2 OR COALESCE(a.absences, 0) >= 4 THEN 'red'
           WHEN COALESCE(f.failed_count, 0) >= 1 OR COALESCE(a.absences, 0) = 3 THEN 'orange'
           WHEN COALESCE(f.failed_count, 0) = 0
                AND COALESCE(a.absences, 0) BETWEEN 1 AND 2 THEN 'yellow'
           ELSE 'none'
       END AS level
FROM t_student s
LEFT JOIN tmp_failed_courses_v28 f ON f.student_id = s.id
LEFT JOIN tmp_absences_v28 a ON a.student_id = s.id;

ALTER TABLE tmp_alert_level_v28 ADD PRIMARY KEY (student_id);

UPDATE t_student s
JOIN tmp_alert_level_v28 p ON p.student_id = s.id
SET s.alert_level = p.level;

DELETE FROM t_alert WHERE status IN ('pending', 'processing');

INSERT INTO t_alert (
    student_id, level, type, title, description, course, failed_courses,
    trigger_date, status, suggestion, trigger_event, pushed_at
)
SELECT p.student_id,
       p.level,
       CASE
           WHEN p.failed_count > 0 AND p.absences > 0 THEN '复合风险预警'
           WHEN p.failed_count > 0 THEN '课程预警'
           ELSE '出勤预警'
       END,
       CONCAT(
           CASE p.level WHEN 'red' THEN '红色预警：'
                       WHEN 'orange' THEN '橙色预警：'
                       ELSE '黄色预警：' END,
           CASE
               WHEN p.failed_count > 0 AND p.absences > 0 THEN '挂科与缺勤风险'
               WHEN p.failed_count > 0 THEN '课程不及格风险'
               ELSE '课堂出勤风险'
           END
       ),
       CONCAT('当学期不及格 ', p.failed_count, ' 门，缺勤 ', p.absences, ' 次。'),
       p.first_course,
       p.failed_courses,
       DATE '2025-01-15',
       CASE
           WHEN p.level = 'red' THEN 'pending'
           WHEN p.level = 'orange' AND MOD(p.student_id, 2) = 0 THEN 'processing'
           WHEN p.level = 'yellow' AND MOD(p.student_id, 3) = 0 THEN 'processing'
           ELSE 'pending'
       END,
       CASE p.level
           WHEN 'red' THEN '建议24小时内由班主任与辅导员联合核查，制定一人一策并持续跟踪'
           WHEN 'orange' THEN '建议一周内完成约谈，明确主要风险原因并安排针对性帮扶'
           ELSE '建议两周内进行提醒与复查，帮助学生调整学习或生活安排'
       END,
       CONCAT('挂科=', p.failed_count, '；缺勤=', p.absences),
       '2025-01-15 14:00:00'
FROM tmp_alert_level_v28 p
WHERE p.level <> 'none';

DROP TEMPORARY TABLE IF EXISTS tmp_alert_level_v28;
DROP TEMPORARY TABLE IF EXISTS tmp_absences_v28;
DROP TEMPORARY TABLE IF EXISTS tmp_failed_courses_v28;
