-- Build a realistic, deterministic demo distribution across cohorts, classes,
-- courses and terms. Warning levels still follow AcademicAlertPolicy exactly:
-- yellow = 1-2 absences and no failure; orange = 1 failure or 3 absences;
-- red = 2+ failures or 4+ absences.

DROP TEMPORARY TABLE IF EXISTS tmp_v33_student_term_base;
DROP TEMPORARY TABLE IF EXISTS tmp_v33_student_term_targets;
DROP TEMPORARY TABLE IF EXISTS tmp_v33_student_term_risk;
DROP TEMPORARY TABLE IF EXISTS tmp_v33_grade_profile;
DROP TEMPORARY TABLE IF EXISTS tmp_v33_gpa;
DROP TEMPORARY TABLE IF EXISTS tmp_v33_gpa_rank;
DROP TEMPORARY TABLE IF EXISTS tmp_v33_latest_gpa;
DROP TEMPORARY TABLE IF EXISTS tmp_v33_credits;
DROP TEMPORARY TABLE IF EXISTS tmp_v33_failures;
DROP TEMPORARY TABLE IF EXISTS tmp_v33_profile_avg;

-- Each class-term receives a slightly different warning mix. Across the school,
-- roughly 88% remain normal, yellow is the largest warning group and red is rare.
CREATE TEMPORARY TABLE tmp_v33_student_term_base AS
SELECT source.*,
       ROW_NUMBER() OVER (
           PARTITION BY source.class_id, source.term
           ORDER BY CRC32(CONCAT('risk-order-', source.student_no, '-', source.term)), source.student_id
       ) AS risk_rank
FROM (
    SELECT DISTINCT s.id AS student_id,
           s.student_id AS student_no,
           s.class_id,
           s.major_id,
           cl.grade AS cohort,
           g.term,
           MOD(CRC32(CONCAT('class-risk-', s.class_id, '-', g.term)), 4) AS class_profile,
           MOD(CRC32(CONCAT('risk-variant-', s.student_id, '-', g.term)), 4) AS risk_variant
    FROM t_student s
    JOIN t_class cl ON cl.id = s.class_id
    JOIN t_grade g ON g.student_id = s.id
) source;

CREATE TEMPORARY TABLE tmp_v33_student_term_targets AS
SELECT b.*,
       CASE b.cohort
           WHEN '2021' THEN 1 + MOD(b.class_profile, 3)
           WHEN '2022' THEN 2 + MOD(b.class_profile, 3)
           WHEN '2023' THEN 2 + MOD(b.class_profile, 3)
           ELSE 2 + MOD(b.class_profile, 3)
       END AS yellow_count,
       CASE b.cohort
           WHEN '2021' THEN 1 + MOD(b.class_profile, 2)
           WHEN '2022' THEN 1 + MOD(b.class_profile + 1, 2)
           WHEN '2023' THEN 1 + MOD(b.class_profile, 2)
           ELSE 1 + MOD(b.class_profile, 2)
       END AS orange_count,
       CASE b.cohort
           WHEN '2021' THEN IF(b.class_profile = 3, 1, 0)
           WHEN '2022' THEN IF(b.class_profile >= 2, 1, 0)
           WHEN '2023' THEN IF(b.class_profile <> 0, 1, 0)
           ELSE IF(b.class_profile = 3, 1, 0)
       END AS red_count
FROM tmp_v33_student_term_base b;

CREATE TEMPORARY TABLE tmp_v33_student_term_risk AS
SELECT t.*,
       CASE
           WHEN t.risk_rank <= t.red_count THEN 'red'
           WHEN t.risk_rank <= t.red_count + t.orange_count THEN 'orange'
           WHEN t.risk_rank <= t.red_count + t.orange_count + t.yellow_count THEN 'yellow'
           ELSE 'none'
       END AS level
FROM tmp_v33_student_term_targets t;

ALTER TABLE tmp_v33_student_term_risk ADD PRIMARY KEY (student_id, term);

-- Course results combine four stable signals: student ability, class baseline,
-- cohort stage and course difficulty. Failures are concentrated in harder
-- courses, while normal and yellow-warning students always pass.
CREATE TEMPORARY TABLE tmp_v33_grade_profile AS
SELECT g.id AS grade_id,
       g.student_id,
       g.term,
       r.class_id,
       r.cohort,
       r.level,
       r.risk_variant,
       c.id AS course_id,
       c.code AS course_code,
       MOD(CRC32(CONCAT('course-difficulty-', c.code)), 5) AS difficulty_tier,
       ROW_NUMBER() OVER (
           PARTITION BY g.student_id, g.term
           ORDER BY MOD(CRC32(CONCAT('course-difficulty-', c.code)), 5),
                    CRC32(CONCAT('failure-course-', g.student_id, '-', c.code, '-', g.term)),
                    g.id
       ) AS course_rank
FROM t_grade g
JOIN tmp_v33_student_term_risk r
  ON r.student_id = g.student_id AND r.term = g.term
JOIN t_course_class cc ON cc.id = g.course_class_id
JOIN t_course c ON c.id = cc.course_id;

ALTER TABLE tmp_v33_grade_profile ADD PRIMARY KEY (grade_id);

UPDATE t_grade g
JOIN tmp_v33_grade_profile p ON p.grade_id = g.id
JOIN t_student s ON s.id = g.student_id
SET g.score = CASE
        WHEN p.level = 'orange' AND p.risk_variant <> 3 AND p.course_rank = 1
            THEN 50 + MOD(CRC32(CONCAT('orange-fail-', g.id)), 9)
        WHEN p.level = 'red' AND p.risk_variant <> 3 AND p.course_rank <= 2
            THEN 47 + MOD(CRC32(CONCAT('red-fail-', g.id)), 12)
        ELSE LEAST(98, GREATEST(60,
            74
            + (CAST(MOD(CRC32(CONCAT('class-quality-', p.class_id)), 9) AS SIGNED) - 4)
            + p.difficulty_tier * 2
            + (CAST(MOD(CRC32(CONCAT('student-ability-', s.student_id)), 13) AS SIGNED) - 6)
            + (CAST(MOD(CRC32(CONCAT('term-form-', s.student_id, '-', p.course_code, '-', p.term)), 7) AS SIGNED) - 3)
            + CASE p.cohort WHEN '2021' THEN 2 WHEN '2022' THEN 0
                            WHEN '2023' THEN -1 ELSE 1 END
            + CASE p.level WHEN 'yellow' THEN -1 WHEN 'orange' THEN -2
                           WHEN 'red' THEN -3 ELSE 0 END
        ))
    END;

UPDATE t_grade
SET status = IF(score < 60, 'failed', 'passed'),
    grade_point = CASE
        WHEN score >= 95 THEN 4.0
        WHEN score >= 90 THEN 3.7
        WHEN score >= 85 THEN 3.3
        WHEN score >= 80 THEN 3.0
        WHEN score >= 75 THEN 2.7
        WHEN score >= 70 THEN 2.3
        WHEN score >= 65 THEN 2.0
        WHEN score >= 60 THEN 1.7
        ELSE 0.0
    END;

-- Attendance is sparse by design. Under the agreed policy, any 1-2 absences
-- already produce yellow, so normal students must have zero absences.
UPDATE t_attendance a
JOIN tmp_v33_student_term_risk r
  ON r.student_id = a.student_id AND r.term = a.term
SET a.total_classes = 56 + MOD(CRC32(CONCAT('class-hours-', r.class_id, '-', r.term)), 13),
    a.absent_count = CASE
        WHEN r.level = 'yellow' THEN 1 + MOD(CRC32(CONCAT('yellow-absence-', a.student_id, '-', a.term)), 2)
        WHEN r.level = 'orange' AND r.risk_variant = 3 THEN 3
        WHEN r.level = 'orange' THEN MOD(CRC32(CONCAT('orange-absence-', a.student_id, '-', a.term)), 2)
        WHEN r.level = 'red' AND r.risk_variant = 3
            THEN 4 + MOD(CRC32(CONCAT('red-absence-', a.student_id, '-', a.term)), 2)
        WHEN r.level = 'red' THEN MOD(CRC32(CONCAT('red-low-absence-', a.student_id, '-', a.term)), 2)
        ELSE 0
    END,
    a.late_count = CASE
        WHEN r.level = 'none' THEN IF(MOD(CRC32(CONCAT('normal-late-', a.student_id, '-', a.term)), 5) = 0, 1, 0)
        WHEN r.level = 'yellow' THEN MOD(CRC32(CONCAT('yellow-late-', a.student_id, '-', a.term)), 3)
        WHEN r.level = 'orange' THEN 1 + MOD(CRC32(CONCAT('orange-late-', a.student_id, '-', a.term)), 2)
        ELSE 1 + MOD(CRC32(CONCAT('red-late-', a.student_id, '-', a.term)), 3)
    END,
    a.note = 'V33院校演示数据：按班级、年级与学期差异化生成';

-- Recompute term GPA and major ranking from the new course results.
CREATE TEMPORARY TABLE tmp_v33_gpa AS
SELECT g.student_id,
       g.term,
       ROUND(SUM(g.grade_point * c.credits) / NULLIF(SUM(c.credits), 0), 2) AS gpa
FROM t_grade g
JOIN t_course_class cc ON cc.id = g.course_class_id
JOIN t_course c ON c.id = cc.course_id
GROUP BY g.student_id, g.term;

ALTER TABLE tmp_v33_gpa ADD PRIMARY KEY (student_id, term);

CREATE TEMPORARY TABLE tmp_v33_gpa_rank AS
SELECT cg.student_id,
       cg.term,
       cg.gpa,
       ROW_NUMBER() OVER (PARTITION BY s.major_id, cg.term ORDER BY cg.gpa DESC, s.student_id) AS major_rank,
       COUNT(*) OVER (PARTITION BY s.major_id, cg.term) AS major_total,
       ROUND(AVG(cg.gpa) OVER (PARTITION BY s.major_id, cg.term), 2) AS major_avg_gpa
FROM tmp_v33_gpa cg
JOIN t_student s ON s.id = cg.student_id;

ALTER TABLE tmp_v33_gpa_rank ADD PRIMARY KEY (student_id, term);

UPDATE t_gpa_history gh
JOIN tmp_v33_gpa_rank r ON r.student_id = gh.student_id AND r.term = gh.term
SET gh.gpa = r.gpa,
    gh.avg_gpa = r.major_avg_gpa,
    gh.`rank` = r.major_rank,
    gh.total_students = r.major_total;

CREATE TEMPORARY TABLE tmp_v33_latest_gpa AS
SELECT ranked.*
FROM (
    SELECT r.*,
           ROW_NUMBER() OVER (PARTITION BY r.student_id ORDER BY r.term DESC) AS term_rank
    FROM tmp_v33_gpa_rank r
) ranked
WHERE ranked.term_rank = 1;

ALTER TABLE tmp_v33_latest_gpa ADD PRIMARY KEY (student_id);

UPDATE t_student s
JOIN tmp_v33_latest_gpa latest ON latest.student_id = s.id
JOIN tmp_v33_student_term_risk risk
  ON risk.student_id = s.id AND risk.term = latest.term
SET s.gpa = latest.gpa,
    s.`rank` = latest.major_rank,
    s.total_students = latest.major_total,
    s.alert_level = risk.level;

CREATE TEMPORARY TABLE tmp_v33_credits AS
SELECT g.student_id,
       ROUND(SUM(CASE WHEN g.status = 'passed' THEN c.credits ELSE 0 END), 1) AS earned_credits
FROM t_grade g
JOIN t_course_class cc ON cc.id = g.course_class_id
JOIN t_course c ON c.id = cc.course_id
GROUP BY g.student_id;

ALTER TABLE tmp_v33_credits ADD PRIMARY KEY (student_id);

UPDATE t_student s
JOIN tmp_v33_credits cr ON cr.student_id = s.id
SET s.total_credits = LEAST(s.required_credits, cr.earned_credits);

-- Keep the non-academic dimensions varied by student and term as well.
UPDATE t_health_report h
JOIN t_student s ON s.id = h.student_id
JOIN t_class cl ON cl.id = s.class_id
SET h.total_score = LEAST(98, GREATEST(60,
        76
        + (CAST(MOD(CRC32(CONCAT('fitness-', s.student_id)), 19) AS SIGNED) - 8)
        + (CAST(MOD(CRC32(CONCAT('fitness-term-', s.student_id, '-', h.term)), 9) AS SIGNED) - 4)
        + (CAST(MOD(CRC32(CONCAT('fitness-class-', cl.id)), 5) AS SIGNED) - 2)
    )),
    h.items = JSON_ARRAY(
        JSON_OBJECT('name', '体重指数', 'score', LEAST(100, h.total_score + MOD(CRC32(CONCAT('bmi-', h.student_id, '-', h.term)), 6)), 'level', '正常'),
        JSON_OBJECT('name', '肺活量', 'score', GREATEST(55, h.total_score - MOD(CRC32(CONCAT('lung-', h.student_id, '-', h.term)), 7)), 'level', '有效'),
        JSON_OBJECT('name', '耐力跑', 'score', GREATEST(55, h.total_score - MOD(CRC32(CONCAT('run-', h.student_id, '-', h.term)), 9)), 'level', '有效')
    );

UPDATE t_volunteer v
JOIN t_student s ON s.id = v.student_id
SET v.hours = 4
        + MOD(CRC32(CONCAT('volunteer-class-', s.class_id)), 13)
        + MOD(CRC32(CONCAT('volunteer-student-', s.student_id)), 29)
        + MOD(CRC32(CONCAT('volunteer-term-', v.term)), 9),
    v.description = '学期志愿服务汇总（含班级活动与个人服务）';

UPDATE t_psychology p
JOIN t_student s ON s.id = p.student_id
SET p.score = LEAST(96, GREATEST(62,
        69 + MOD(CRC32(CONCAT('assessment-', s.student_id)), 23)
        + (CAST(MOD(CRC32(CONCAT('assessment-term-', s.student_id, '-', p.term)), 9) AS SIGNED) - 4)
    )),
    p.note = '学期综合测评记录（不参与预警判级）';

UPDATE t_psychology
SET level = CASE WHEN score >= 85 THEN 'good'
                 WHEN score >= 70 THEN 'normal' ELSE 'attention' END;

UPDATE t_profile_score ps
JOIN tmp_v33_gpa_rank gh ON gh.student_id = ps.student_id AND gh.term = ps.term
JOIN (
    SELECT student_id, term, ROUND(AVG(score), 1) AS avg_score,
           SUM(score < 60 OR status = 'failed') AS failed_count
    FROM t_grade
    GROUP BY student_id, term
) gs ON gs.student_id = ps.student_id AND gs.term = ps.term
SET ps.score = LEAST(98, GREATEST(45, ROUND(
        28 + gh.gpa / 4 * 54
        + (1 - gh.major_rank / NULLIF(gh.major_total, 0)) * 12
        + gs.avg_score / 100 * 6
        - gs.failed_count * 4
    ))),
    ps.description = '基于当学期GPA、专业排名、课程均分与挂科数计算',
    ps.details = JSON_ARRAY(
        CONCAT('GPA: ', FORMAT(gh.gpa, 2)),
        CONCAT('专业排名: ', gh.major_rank, ' / ', gh.major_total),
        CONCAT('课程均分: ', ROUND(gs.avg_score)),
        CONCAT('挂科门数: ', gs.failed_count)
    )
WHERE ps.dimension_key = 'academic';

UPDATE t_profile_score ps
LEFT JOIN (
    SELECT student_id, term, MAX(points) AS max_points, COUNT(*) AS item_count
    FROM t_competition
    GROUP BY student_id, term
) cs ON cs.student_id = ps.student_id AND cs.term = ps.term
SET ps.score = LEAST(95, 52
        + MOD(CRC32(CONCAT('practice-base-', ps.student_id, '-', ps.term)), 16)
        + LEAST(22, COALESCE(cs.item_count, 0) * 6 + GREATEST(0, COALESCE(cs.max_points, 0) - 70) / 4)),
    ps.description = '基于当学期竞赛、项目参与和稳定实践表现计算',
    ps.details = JSON_ARRAY(
        CONCAT('竞赛记录: ', COALESCE(cs.item_count, 0), ' 项'),
        CONCAT('最高实践积分: ', COALESCE(cs.max_points, 0))
    )
WHERE ps.dimension_key = 'practice';

UPDATE t_profile_score ps
JOIN t_volunteer v ON v.student_id = ps.student_id AND v.term = ps.term
JOIN t_attendance a ON a.student_id = ps.student_id AND a.term = ps.term
SET ps.score = LEAST(96, GREATEST(55, ROUND(
        59 + LEAST(v.hours, 45) * 0.68 - a.absent_count * 3
        + MOD(CRC32(CONCAT('quality-', ps.student_id, '-', ps.term)), 7)
    ))),
    ps.description = '基于当学期志愿服务、课堂出勤和综合表现计算',
    ps.details = JSON_ARRAY(
        CONCAT('志愿时长: ', v.hours, ' 小时'),
        CONCAT('缺勤次数: ', a.absent_count, ' 次')
    )
WHERE ps.dimension_key = 'quality';

UPDATE t_profile_score ps
JOIN t_psychology p ON p.student_id = ps.student_id AND p.term = ps.term
SET ps.score = LEAST(95, GREATEST(60, ROUND(
        59 + p.score * 0.28
        + MOD(CRC32(CONCAT('culture-', ps.student_id, '-', ps.term)), 9)
    ))),
    ps.description = '基于当学期综合测评与人文活动表现计算',
    ps.details = JSON_ARRAY(CONCAT('综合测评: ', p.score, ' 分'))
WHERE ps.dimension_key = 'culture';

UPDATE t_profile_score ps
JOIN t_health_report h ON h.student_id = ps.student_id AND h.term = ps.term
SET ps.score = h.total_score,
    ps.description = '基于同一学期体测总分计算',
    ps.details = JSON_ARRAY(CONCAT('体测总分: ', h.total_score, ' 分'))
WHERE ps.dimension_key = 'health';

-- Store the real class-term average for every radar dimension.
CREATE TEMPORARY TABLE tmp_v33_profile_avg AS
SELECT s.class_id,
       ps.term,
       ps.dimension_key,
       ROUND(AVG(ps.score)) AS class_avg
FROM t_profile_score ps
JOIN t_student s ON s.id = ps.student_id
GROUP BY s.class_id, ps.term, ps.dimension_key;

ALTER TABLE tmp_v33_profile_avg ADD PRIMARY KEY (class_id, term, dimension_key);

UPDATE t_profile_score ps
JOIN t_student s ON s.id = ps.student_id
JOIN tmp_v33_profile_avg avg_score
  ON avg_score.class_id = s.class_id
 AND avg_score.term = ps.term
 AND avg_score.dimension_key = ps.dimension_key
SET ps.avg_score = avg_score.class_avg,
    ps.max_score = 100;

-- Rebuild alert records from the same grade and attendance facts used by all
-- dashboards. There are no intervention records in the demo dataset.
CREATE TEMPORARY TABLE tmp_v33_failures AS
SELECT g.student_id,
       g.term,
       COUNT(*) AS failed_count,
       JSON_ARRAYAGG(c.name) AS failed_courses,
       MIN(c.name) AS first_course
FROM t_grade g
JOIN t_course_class cc ON cc.id = g.course_class_id
JOIN t_course c ON c.id = cc.course_id
WHERE g.status = 'failed' OR g.score < 60
GROUP BY g.student_id, g.term;

ALTER TABLE tmp_v33_failures ADD PRIMARY KEY (student_id, term);

DELETE FROM t_alert;

INSERT INTO t_alert (
    student_id, level, type, title, description, course, failed_courses,
    trigger_date, status, suggestion, trigger_event, pushed_at
)
SELECT risk.student_id,
       risk.level,
       CASE
           WHEN COALESCE(f.failed_count, 0) > 0 AND a.absent_count > 0 THEN '复合风险预警'
           WHEN COALESCE(f.failed_count, 0) > 0 THEN '课程预警'
           ELSE '出勤预警'
       END,
       CONCAT(
           CASE risk.level WHEN 'red' THEN '红色预警：'
                           WHEN 'orange' THEN '橙色预警：'
                           ELSE '黄色预警：' END,
           CASE
               WHEN COALESCE(f.failed_count, 0) > 0 AND a.absent_count > 0 THEN '挂科与缺勤风险'
               WHEN COALESCE(f.failed_count, 0) > 0 THEN '课程不及格风险'
               ELSE '课堂出勤风险'
           END
       ),
       CONCAT('当学期不及格 ', COALESCE(f.failed_count, 0), ' 门，缺勤 ', a.absent_count, ' 次。'),
       f.first_course,
       COALESCE(f.failed_courses, JSON_ARRAY()),
       STR_TO_DATE(CONCAT(SUBSTRING(risk.term, 6, 4),
            IF(RIGHT(risk.term, 1) = '1', '-01-15', '-07-10')), '%Y-%m-%d'),
       CASE
           WHEN risk.term <> latest.term THEN 'resolved'
           WHEN risk.level = 'orange' AND MOD(risk.student_id, 2) = 0 THEN 'processing'
           WHEN risk.level = 'yellow' AND MOD(risk.student_id, 3) = 0 THEN 'processing'
           ELSE 'pending'
       END,
       CASE risk.level
           WHEN 'red' THEN '建议24小时内联合核查，制定一人一策并持续跟踪'
           WHEN 'orange' THEN '建议一周内完成约谈，明确风险原因并安排针对性帮扶'
           ELSE '建议两周内进行提醒与复查，帮助学生及时调整'
       END,
       CONCAT('挂科=', COALESCE(f.failed_count, 0), '；缺勤=', a.absent_count),
       TIMESTAMP(
           STR_TO_DATE(CONCAT(SUBSTRING(risk.term, 6, 4),
               IF(RIGHT(risk.term, 1) = '1', '-01-15', '-07-10')), '%Y-%m-%d'),
           '14:00:00'
       )
FROM tmp_v33_student_term_risk risk
JOIN tmp_v33_latest_gpa latest ON latest.student_id = risk.student_id
JOIN t_attendance a ON a.student_id = risk.student_id AND a.term = risk.term
LEFT JOIN tmp_v33_failures f ON f.student_id = risk.student_id AND f.term = risk.term
WHERE risk.level <> 'none';

DROP TEMPORARY TABLE IF EXISTS tmp_v33_profile_avg;
DROP TEMPORARY TABLE IF EXISTS tmp_v33_failures;
DROP TEMPORARY TABLE IF EXISTS tmp_v33_credits;
DROP TEMPORARY TABLE IF EXISTS tmp_v33_latest_gpa;
DROP TEMPORARY TABLE IF EXISTS tmp_v33_gpa_rank;
DROP TEMPORARY TABLE IF EXISTS tmp_v33_gpa;
DROP TEMPORARY TABLE IF EXISTS tmp_v33_grade_profile;
DROP TEMPORARY TABLE IF EXISTS tmp_v33_student_term_risk;
DROP TEMPORARY TABLE IF EXISTS tmp_v33_student_term_targets;
DROP TEMPORARY TABLE IF EXISTS tmp_v33_student_term_base;
