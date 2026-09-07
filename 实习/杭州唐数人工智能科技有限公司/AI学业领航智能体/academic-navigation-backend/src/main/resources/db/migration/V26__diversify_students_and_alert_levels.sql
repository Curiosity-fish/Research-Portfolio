-- V26: diversify the 2,560-student demo dataset and rebuild current alerts.
-- Per 40-student class: 28 normal, 8 yellow, 3 orange, 1 red.

DROP TEMPORARY TABLE IF EXISTS tmp_student_risk_profile;
CREATE TEMPORARY TABLE tmp_student_risk_profile AS
SELECT ranked.student_id,
       ranked.class_id,
       ranked.class_rank,
       CASE
           WHEN ranked.class_rank = 1 THEN 'red'
           WHEN ranked.class_rank <= 4 THEN 'orange'
           WHEN ranked.class_rank <= 12 THEN 'yellow'
           ELSE 'none'
       END AS risk_level,
       MOD(CRC32(CONCAT(ranked.student_no, '-risk-variant')), 4) AS risk_variant
FROM (
    SELECT s.id AS student_id,
           s.student_id AS student_no,
           s.class_id,
           ROW_NUMBER() OVER (
               PARTITION BY s.class_id
               ORDER BY CRC32(CONCAT(s.student_id, '-v26')), s.id
           ) AS class_rank
    FROM t_student s
) ranked;

ALTER TABLE tmp_student_risk_profile ADD PRIMARY KEY (student_id);

-- Give each current-term course a stable order so failure counts can be controlled.
DROP TEMPORARY TABLE IF EXISTS tmp_current_grade_profile;
CREATE TEMPORARY TABLE tmp_current_grade_profile AS
SELECT g.id AS grade_id,
       g.student_id,
       p.risk_level,
       p.risk_variant,
       ROW_NUMBER() OVER (
           PARTITION BY g.student_id
           ORDER BY CRC32(CONCAT(g.id, '-course-v26')), g.id
       ) AS course_rank
FROM t_grade g
JOIN tmp_student_risk_profile p ON p.student_id = g.student_id
WHERE g.term = '2024-2025-1';

ALTER TABLE tmp_current_grade_profile ADD PRIMARY KEY (grade_id);

UPDATE t_grade g
JOIN tmp_current_grade_profile p ON p.grade_id = g.id
SET g.score = CASE
    WHEN p.risk_level = 'none'
        THEN 80 + MOD(CRC32(CONCAT(g.id, '-normal')), 20)

    WHEN p.risk_level = 'yellow' AND p.risk_variant = 0
        THEN 74 + MOD(CRC32(CONCAT(g.id, '-yellow-gpa')), 6)
    WHEN p.risk_level = 'yellow' AND p.risk_variant = 1 AND p.course_rank = 1
        THEN 52 + MOD(CRC32(CONCAT(g.id, '-yellow-fail')), 7)
    WHEN p.risk_level = 'yellow' AND p.risk_variant = 1
        THEN 92 + MOD(CRC32(CONCAT(g.id, '-yellow-pass')), 8)
    WHEN p.risk_level = 'yellow'
        THEN 82 + MOD(CRC32(CONCAT(g.id, '-yellow')), 13)

    WHEN p.risk_level = 'orange' AND p.risk_variant = 0
        THEN 65 + MOD(CRC32(CONCAT(g.id, '-orange-gpa')), 8)
    WHEN p.risk_level = 'orange' AND p.risk_variant = 1 AND p.course_rank <= 2
        THEN 48 + MOD(CRC32(CONCAT(g.id, '-orange-fail')), 11)
    WHEN p.risk_level = 'orange' AND p.risk_variant = 1
        THEN 90 + MOD(CRC32(CONCAT(g.id, '-orange-pass')), 9)
    WHEN p.risk_level = 'orange'
        THEN 82 + MOD(CRC32(CONCAT(g.id, '-orange')), 13)

    WHEN p.risk_level = 'red' AND p.risk_variant = 0 AND p.course_rank <= 4
        THEN 42 + MOD(CRC32(CONCAT(g.id, '-red-fail')), 16)
    WHEN p.risk_level = 'red' AND p.risk_variant = 0
        THEN 72 + MOD(CRC32(CONCAT(g.id, '-red-pass')), 14)
    WHEN p.risk_level = 'red' AND p.risk_variant = 3 AND p.course_rank <= 2
        THEN 45 + MOD(CRC32(CONCAT(g.id, '-red-combined-fail')), 14)
    WHEN p.risk_level = 'red' AND p.risk_variant = 3
        THEN 78 + MOD(CRC32(CONCAT(g.id, '-red-combined-pass')), 10)
    ELSE 80 + MOD(CRC32(CONCAT(g.id, '-red-other')), 12)
END;

UPDATE t_grade g
JOIN tmp_current_grade_profile p ON p.grade_id = g.id
SET g.status = IF(g.score < 60, 'failed', 'passed'),
    g.grade_point = CASE
        WHEN g.score >= 95 THEN 4.0
        WHEN g.score >= 90 THEN 3.7
        WHEN g.score >= 85 THEN 3.3
        WHEN g.score >= 80 THEN 3.0
        WHEN g.score >= 75 THEN 2.7
        WHEN g.score >= 70 THEN 2.3
        WHEN g.score >= 65 THEN 2.0
        WHEN g.score >= 60 THEN 1.7
        ELSE 0.0
    END;

-- Diversify attendance. Only the attendance variant crosses attendance thresholds.
UPDATE t_attendance a
JOIN tmp_student_risk_profile p ON p.student_id = a.student_id
SET a.absent_count = CASE
        WHEN p.risk_level = 'none' THEN MOD(CRC32(CONCAT(a.student_id, '-absence')), 4)
        WHEN p.risk_level = 'yellow' AND p.risk_variant = 2
            THEN 5 + MOD(CRC32(CONCAT(a.student_id, '-yellow-absence')), 3)
        WHEN p.risk_level = 'yellow' THEN MOD(CRC32(CONCAT(a.student_id, '-yellow-low-absence')), 4)
        WHEN p.risk_level = 'orange' AND p.risk_variant = 2
            THEN 8 + MOD(CRC32(CONCAT(a.student_id, '-orange-absence')), 4)
        WHEN p.risk_level = 'orange' THEN MOD(CRC32(CONCAT(a.student_id, '-orange-low-absence')), 4)
        WHEN p.risk_level = 'red' AND p.risk_variant = 1
            THEN 15 + MOD(CRC32(CONCAT(a.student_id, '-red-absence')), 4)
        WHEN p.risk_level = 'red' AND p.risk_variant = 3
            THEN 8 + MOD(CRC32(CONCAT(a.student_id, '-red-combined-absence')), 4)
        ELSE 2 + MOD(CRC32(CONCAT(a.student_id, '-red-low-absence')), 3)
    END,
    a.late_count = CASE
        WHEN p.risk_level = 'none' THEN MOD(CRC32(CONCAT(a.student_id, '-late')), 3)
        WHEN p.risk_level = 'yellow' THEN 1 + MOD(CRC32(CONCAT(a.student_id, '-yellow-late')), 4)
        WHEN p.risk_level = 'orange' THEN 2 + MOD(CRC32(CONCAT(a.student_id, '-orange-late')), 5)
        ELSE 3 + MOD(CRC32(CONCAT(a.student_id, '-red-late')), 6)
    END,
    a.note = CONCAT('V26演示数据：',
        CASE p.risk_level WHEN 'red' THEN '高风险' WHEN 'orange' THEN '中风险'
             WHEN 'yellow' THEN '轻风险' ELSE '正常' END)
WHERE a.term = '2024-2025-1';

-- Keep psychology scores as independent profile data. They do not determine alert levels.
UPDATE t_psychology psy
JOIN tmp_student_risk_profile p ON p.student_id = psy.student_id
SET psy.score = 75 + MOD(CRC32(CONCAT(psy.student_id, '-psy-independent')), 21),
    psy.note = 'V26心理测评演示数据（不参与预警判级）'
WHERE psy.term = '2024-2025-1';

UPDATE t_psychology
SET level = CASE
        WHEN score >= 85 THEN 'good'
        WHEN score >= 72 THEN 'normal'
        ELSE 'attention'
    END
WHERE term = '2024-2025-1';

-- Recalculate current GPA from the diversified course results.
DROP TEMPORARY TABLE IF EXISTS tmp_current_gpa;
CREATE TEMPORARY TABLE tmp_current_gpa AS
SELECT g.student_id,
       ROUND(SUM(g.grade_point * c.credits) / NULLIF(SUM(c.credits), 0), 2) AS gpa
FROM t_grade g
JOIN t_course_class cc ON cc.id = g.course_class_id
JOIN t_course c ON c.id = cc.course_id
WHERE g.term = '2024-2025-1'
GROUP BY g.student_id;

ALTER TABLE tmp_current_gpa ADD PRIMARY KEY (student_id);

UPDATE t_student s
JOIN tmp_current_gpa cg ON cg.student_id = s.id
JOIN tmp_student_risk_profile p ON p.student_id = s.id
JOIN t_class cl ON cl.id = s.class_id
SET s.gpa = cg.gpa,
    s.alert_level = p.risk_level,
    s.total_credits = LEAST(s.required_credits,
        GREATEST(24,
            CASE cl.grade
                WHEN '2021' THEN 145
                WHEN '2022' THEN 110
                WHEN '2023' THEN 74
                ELSE 38
            END
            + MOD(CRC32(CONCAT(s.student_id, '-credits')), 7)
            - CASE p.risk_level WHEN 'red' THEN 14 WHEN 'orange' THEN 7
                  WHEN 'yellow' THEN 3 ELSE 0 END));

UPDATE t_gpa_history gh
JOIN tmp_current_gpa cg ON cg.student_id = gh.student_id
SET gh.gpa = cg.gpa
WHERE gh.term = '2024-2025-1';

-- Keep ranks and professional averages aligned with the new GPA values.
DROP TEMPORARY TABLE IF EXISTS tmp_current_rank;
CREATE TEMPORARY TABLE tmp_current_rank AS
SELECT s.id AS student_id,
       ROW_NUMBER() OVER (PARTITION BY s.major_id ORDER BY s.gpa DESC, s.student_id) AS major_rank,
       COUNT(*) OVER (PARTITION BY s.major_id) AS major_total,
       ROUND(AVG(s.gpa) OVER (PARTITION BY s.major_id), 2) AS major_avg_gpa
FROM t_student s;

ALTER TABLE tmp_current_rank ADD PRIMARY KEY (student_id);

UPDATE t_student s
JOIN tmp_current_rank r ON r.student_id = s.id
SET s.`rank` = r.major_rank,
    s.total_students = r.major_total;

UPDATE t_gpa_history gh
JOIN tmp_current_rank r ON r.student_id = gh.student_id
SET gh.`rank` = r.major_rank,
    gh.total_students = r.major_total,
    gh.avg_gpa = r.major_avg_gpa
WHERE gh.term = '2024-2025-1';

-- Refresh the academic and psychology-related profile dimensions.
UPDATE t_profile_score ps
JOIN t_student s ON s.id = ps.student_id
SET ps.score = GREATEST(30, LEAST(98, ROUND(25 + s.gpa / 4.0 * 70))),
    ps.description = '基于V26多样化成绩、GPA与排名计算'
WHERE ps.term = '2024-2025-1' AND ps.dimension_key = 'academic';

UPDATE t_profile_score ps
JOIN t_psychology psy ON psy.student_id = ps.student_id AND psy.term = ps.term
SET ps.score = psy.score,
    ps.description = '基于V26多样化心理测评结果计算'
WHERE ps.term = '2024-2025-1' AND ps.dimension_key = 'culture';

-- Replace only current, unresolved alerts. Historical resolved alerts remain for trends.
DELETE FROM t_alert WHERE status IN ('pending', 'processing');

DROP TEMPORARY TABLE IF EXISTS tmp_current_failures;
CREATE TEMPORARY TABLE tmp_current_failures AS
SELECT g.student_id,
       COUNT(*) AS failed_count,
       JSON_ARRAYAGG(c.name) AS failed_courses,
       MIN(c.name) AS first_course
FROM t_grade g
JOIN t_course_class cc ON cc.id = g.course_class_id
JOIN t_course c ON c.id = cc.course_id
WHERE g.term = '2024-2025-1' AND (g.status = 'failed' OR g.score < 60)
GROUP BY g.student_id;

ALTER TABLE tmp_current_failures ADD PRIMARY KEY (student_id);

INSERT INTO t_alert (
    student_id, level, type, title, description, course, failed_courses,
    trigger_date, status, suggestion, trigger_event, pushed_at
)
SELECT s.id,
       p.risk_level,
       CASE
           WHEN p.risk_variant = 3 THEN '复合风险预警'
           WHEN p.risk_variant = 0 THEN '学业预警'
           WHEN p.risk_variant = 1 THEN '课程预警'
           ELSE '出勤预警'
       END AS alert_type,
       CONCAT(
           CASE p.risk_level WHEN 'red' THEN '红色预警：' WHEN 'orange' THEN '橙色预警：'
                ELSE '黄色预警：' END,
           CASE
               WHEN p.risk_variant = 3 THEN '学业与出勤指标持续异常'
               WHEN p.risk_variant = 0 THEN '学业表现低于关注线'
               WHEN p.risk_variant = 1 THEN '课程不及格风险'
               WHEN p.risk_variant = 2 THEN '课堂出勤风险'
               ELSE '学业与出勤指标需要关注'
           END
       ),
       CONCAT('当前GPA ', FORMAT(s.gpa, 2),
              '，当学期不及格 ', COALESCE(f.failed_count, 0),
              ' 门，缺勤 ', a.absent_count, ' 次。'),
       f.first_course,
       COALESCE(f.failed_courses, JSON_ARRAY()),
       DATE '2025-01-15',
       CASE
           WHEN p.risk_level = 'red' THEN 'pending'
           WHEN p.risk_level = 'orange' AND MOD(s.id, 2) = 0 THEN 'processing'
           WHEN p.risk_level = 'yellow' AND MOD(s.id, 3) = 0 THEN 'processing'
           ELSE 'pending'
       END,
       CASE p.risk_level
           WHEN 'red' THEN '建议24小时内由班主任与辅导员联合核查，制定一人一策并持续跟踪'
           WHEN 'orange' THEN '建议一周内完成约谈，明确主要风险原因并安排针对性帮扶'
           ELSE '建议两周内进行提醒与复查，帮助学生调整学习或生活安排'
       END,
       CONCAT('GPA=', FORMAT(s.gpa, 2),
              '；挂科=', COALESCE(f.failed_count, 0),
              '；缺勤=', a.absent_count),
       '2025-01-15 14:00:00'
FROM t_student s
JOIN tmp_student_risk_profile p ON p.student_id = s.id
JOIN t_attendance a ON a.student_id = s.id AND a.term = '2024-2025-1'
LEFT JOIN tmp_current_failures f ON f.student_id = s.id
WHERE p.risk_level <> 'none';

DROP TEMPORARY TABLE IF EXISTS tmp_current_failures;
DROP TEMPORARY TABLE IF EXISTS tmp_current_rank;
DROP TEMPORARY TABLE IF EXISTS tmp_current_gpa;
DROP TEMPORARY TABLE IF EXISTS tmp_current_grade_profile;
DROP TEMPORARY TABLE IF EXISTS tmp_student_risk_profile;
