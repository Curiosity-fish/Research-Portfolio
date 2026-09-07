-- V27: remove psychology scores from alert grading and rebuild current alerts.
-- Psychology records remain available for authorized human follow-up, but the
-- warning level is determined by GPA, failed courses, and absences only.

-- Decouple the demo psychology values from the old risk variants as well.
UPDATE t_psychology
SET score = 75 + MOD(CRC32(CONCAT(student_id, '-psy-independent')), 21),
    note = 'V27心理测评演示数据（不参与预警判级）'
WHERE term = '2024-2025-1';

UPDATE t_psychology
SET level = CASE
        WHEN score >= 85 THEN 'good'
        WHEN score >= 72 THEN 'normal'
        ELSE 'attention'
    END
WHERE term = '2024-2025-1';

DROP TEMPORARY TABLE IF EXISTS tmp_alert_level_v27;
DROP TEMPORARY TABLE IF EXISTS tmp_failed_courses_v27;

CREATE TEMPORARY TABLE tmp_failed_courses_v27 AS
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

ALTER TABLE tmp_failed_courses_v27 ADD PRIMARY KEY (student_id);

CREATE TEMPORARY TABLE tmp_alert_level_v27 AS
SELECT s.id AS student_id,
       s.gpa,
       COALESCE(f.failed_count, 0) AS failed_count,
       COALESCE(f.failed_courses, JSON_ARRAY()) AS failed_courses,
       f.first_course,
       COALESCE(a.absent_count, 0) AS absences,
       (CASE WHEN s.gpa < 2.40 THEN 1 ELSE 0 END
        + CASE WHEN COALESCE(f.failed_count, 0) >= 2 THEN 1 ELSE 0 END
        + CASE WHEN COALESCE(a.absent_count, 0) >= 8 THEN 1 ELSE 0 END) AS moderate_signals,
       CASE
           WHEN s.gpa < 1.80
                OR COALESCE(f.failed_count, 0) >= 4
                OR COALESCE(a.absent_count, 0) >= 15
                OR (CASE WHEN s.gpa < 2.40 THEN 1 ELSE 0 END
                    + CASE WHEN COALESCE(f.failed_count, 0) >= 2 THEN 1 ELSE 0 END
                    + CASE WHEN COALESCE(a.absent_count, 0) >= 8 THEN 1 ELSE 0 END) >= 3
               THEN 'red'
           WHEN (s.gpa < 2.40
                 OR COALESCE(f.failed_count, 0) >= 2
                 OR COALESCE(a.absent_count, 0) >= 8)
                OR ((CASE WHEN s.gpa < 2.80 THEN 1 ELSE 0 END
                     + CASE WHEN COALESCE(f.failed_count, 0) >= 1 THEN 1 ELSE 0 END
                     + CASE WHEN COALESCE(a.absent_count, 0) >= 5 THEN 1 ELSE 0 END) >= 2)
               THEN 'orange'
           WHEN (CASE WHEN s.gpa < 2.80 THEN 1 ELSE 0 END
                 + CASE WHEN COALESCE(f.failed_count, 0) >= 1 THEN 1 ELSE 0 END
                 + CASE WHEN COALESCE(a.absent_count, 0) >= 5 THEN 1 ELSE 0 END) = 1
               THEN 'yellow'
           ELSE 'none'
       END AS level
FROM t_student s
LEFT JOIN tmp_failed_courses_v27 f ON f.student_id = s.id
LEFT JOIN t_attendance a ON a.student_id = s.id AND a.term = '2024-2025-1';

ALTER TABLE tmp_alert_level_v27 ADD PRIMARY KEY (student_id);

UPDATE t_student s
JOIN tmp_alert_level_v27 p ON p.student_id = s.id
SET s.alert_level = p.level;

-- Rebuild only unresolved records. Historical resolved alerts remain available
-- for trend reporting.
DELETE FROM t_alert WHERE status IN ('pending', 'processing');

INSERT INTO t_alert (
    student_id, level, type, title, description, course, failed_courses,
    trigger_date, status, suggestion, trigger_event, pushed_at
)
SELECT p.student_id,
       p.level,
       CASE
           WHEN p.moderate_signals >= 2 THEN '复合风险预警'
           WHEN p.gpa < 2.80 THEN '学业预警'
           WHEN p.failed_count >= 1 THEN '课程预警'
           ELSE '出勤预警'
       END AS alert_type,
       CONCAT(
           CASE p.level WHEN 'red' THEN '红色预警：'
                       WHEN 'orange' THEN '橙色预警：'
                       ELSE '黄色预警：' END,
           CASE
               WHEN p.moderate_signals >= 2 THEN '学业与出勤指标持续异常'
               WHEN p.gpa < 2.80 THEN '学业表现低于关注线'
               WHEN p.failed_count >= 1 THEN '课程不及格风险'
               ELSE '课堂出勤风险'
           END
       ),
       CONCAT('当前GPA ', FORMAT(p.gpa, 2),
              '，当学期不及格 ', p.failed_count,
              ' 门，缺勤 ', p.absences, ' 次。'),
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
       CONCAT('GPA=', FORMAT(p.gpa, 2),
              '；挂科=', p.failed_count,
              '；缺勤=', p.absences),
       '2025-01-15 14:00:00'
FROM tmp_alert_level_v27 p
WHERE p.level <> 'none';

DROP TEMPORARY TABLE IF EXISTS tmp_alert_level_v27;
DROP TEMPORARY TABLE IF EXISTS tmp_failed_courses_v27;
