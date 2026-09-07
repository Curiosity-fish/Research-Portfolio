-- V23: 补齐教育学院 2024 届 320 名学生

-- 1) 学生账号
INSERT INTO t_user (id, account, name, password, role, college, major, class_name, grade)
WITH RECURSIVE seq AS (SELECT 1 n UNION ALL SELECT n + 1 FROM seq WHERE n < 320),
cls AS (
    SELECT 57 id, '教育学2401' name, 5 major_id, '2024' grade UNION ALL
    SELECT 58, '教育学2402', 5, '2024' UNION ALL
    SELECT 59, '学前2401', 6, '2024' UNION ALL
    SELECT 60, '学前2402', 6, '2024' UNION ALL
    SELECT 61, '小教2401', 7, '2024' UNION ALL
    SELECT 62, '小教2402', 7, '2024' UNION ALL
    SELECT 63, '教技2401', 8, '2024' UNION ALL
    SELECT 64, '教技2402', 8, '2024'
)
SELECT 2386 + s.n,
       CONCAT(cl.grade, cl.major_id, LPAD(((cl.id - 33) % 2) * 40 + MOD(s.n - 1, 40) + 1, 3, '0')),
       CONCAT('教院学生', LPAD(960 + s.n, 3, '0')),
       '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO',
       'student', '教育学院',
       CASE cl.major_id WHEN 5 THEN '教育学' WHEN 6 THEN '学前教育' WHEN 7 THEN '小学教育' ELSE '教育技术学' END,
       cl.name, cl.grade
FROM seq s
JOIN cls cl ON cl.id = 57 + FLOOR((s.n - 1) / 40);

-- 2) 学生档案
INSERT INTO t_student (student_id, user_id, class_id, major_id, department_id, gpa, `rank`, total_students, alert_level, total_credits, required_credits)
SELECT u.account, u.id, c.id, c.major_id, 2,
       ROUND(2.0 + MOD(u.id * 7, 200) / 100, 2),
       MOD(u.id * 3, 40) + 1, 40, 'none',
       16 + MOD(u.id, 4), 160.0
FROM t_user u
JOIN t_class c ON c.name = u.class_name
WHERE u.role = 'student' AND u.id BETWEEN 2387 AND 2706;

-- 3) 成绩
INSERT INTO t_grade (student_id, course_class_id, score, status, term, exam_date)
SELECT s.id, cc.id,
       CASE WHEN MOD(s.id + cc.course_id, 11) = 0 THEN 45 + MOD(s.id + cc.id, 12)
            ELSE 70 + MOD(s.id * 13 + cc.id * 7, 26)
       END,
       'passed', cc.term, '2025-01-15'
FROM t_student s
JOIN t_class cl ON cl.id = s.class_id
JOIN t_course_class cc ON cc.class_name = cl.name
JOIN t_user u ON u.id = s.user_id
WHERE u.id BETWEEN 2387 AND 2706;

-- 4) 当前学期保留部分挂科
UPDATE t_grade g
JOIN t_course_class cc ON cc.id = g.course_class_id
JOIN t_student s ON s.id = g.student_id
JOIN t_class cl ON cl.id = s.class_id
JOIN t_user u ON u.id = s.user_id
SET g.score = 42 + MOD(g.id, 15)
WHERE u.id BETWEEN 2387 AND 2706
  AND cc.term = '2024-2025-1'
  AND MOD(s.id, 8) IN (0, 3)
  AND cc.course_id IN (
      25 + (cl.major_id - 5) * 6 + MOD(s.id, 6),
      25 + (cl.major_id - 5) * 6 + MOD(s.id + 3, 6)
  );

-- 5) 状态与绩点
UPDATE t_grade g
JOIN t_student s ON s.id = g.student_id
JOIN t_user u ON u.id = s.user_id
SET g.status = IF(g.score >= 60, 'passed', 'failed'),
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
    END
WHERE u.id BETWEEN 2387 AND 2706;

-- 6) 体测、考勤、志愿、心理、竞赛
INSERT INTO t_health_report (student_id, term, total_score, items, report_url)
SELECT s.id, '2024-2025-1', 60 + MOD(s.id * 5, 36),
       JSON_ARRAY(
           JSON_OBJECT('name', '体重指数', 'score', 60 + MOD(s.id * 3, 36), 'level', '良好'),
           JSON_OBJECT('name', '肺活量', 'score', 60 + MOD(s.id * 5, 36), 'level', '良好'),
           JSON_OBJECT('name', '耐力跑', 'score', 60 + MOD(s.id * 7, 36), 'level', '良好')
       ), NULL
FROM t_student s
JOIN t_user u ON u.id = s.user_id
WHERE u.id BETWEEN 2387 AND 2706;

INSERT INTO t_attendance (student_id, term, total_classes, absent_count, late_count, note)
SELECT s.id, '2024-2025-1', 60,
       CASE
           WHEN MOD(s.id, 6) = 0 THEN 6 + MOD(s.id, 4)
           WHEN MOD(s.id, 13) = 0 THEN 9 + MOD(s.id, 3)
           ELSE MOD(s.id, 5)
       END,
       MOD(s.id, 5), '模拟考勤数据'
FROM t_student s
JOIN t_user u ON u.id = s.user_id
WHERE u.id BETWEEN 2387 AND 2706;

INSERT INTO t_volunteer (student_id, term, hours, description)
SELECT s.id, '2024-2025-1', 6 + MOD(s.id * 7, 60), '模拟志愿服务记录'
FROM t_student s
JOIN t_user u ON u.id = s.user_id
WHERE u.id BETWEEN 2387 AND 2706;

INSERT INTO t_psychology (student_id, term, score, level, note)
SELECT s.id, '2024-2025-1',
       62 + MOD(s.id * 5, 36),
       CASE WHEN 62 + MOD(s.id * 5, 36) >= 85 THEN 'good'
            WHEN 62 + MOD(s.id * 5, 36) >= 70 THEN 'normal'
            ELSE 'attention' END,
       '模拟心理测评记录'
FROM t_student s
JOIN t_user u ON u.id = s.user_id
WHERE u.id BETWEEN 2387 AND 2706;

INSERT INTO t_competition (student_id, term, competition_name, level, award, points)
SELECT s.id, '2024-2025-1',
       CASE cl.major_id WHEN 5 THEN '师范生教学技能大赛'
                        WHEN 6 THEN '学前教育创新大赛'
                        WHEN 7 THEN '小学教育案例大赛'
                        ELSE '教育技术应用竞赛' END,
       CASE WHEN MOD(s.id, 6) = 0 THEN 'national' WHEN MOD(s.id, 6) = 1 THEN 'provincial' ELSE 'school' END,
       CASE WHEN MOD(s.id, 4) = 0 THEN 'first' WHEN MOD(s.id, 4) = 1 THEN 'second' WHEN MOD(s.id, 4) = 2 THEN 'third' ELSE 'participation' END,
       CASE WHEN MOD(s.id, 6) = 0 AND MOD(s.id, 4) = 0 THEN 92
            WHEN MOD(s.id, 6) = 0 THEN 85
            WHEN MOD(s.id, 6) = 1 AND MOD(s.id, 4) = 0 THEN 82
            WHEN MOD(s.id, 6) = 1 THEN 76
            WHEN MOD(s.id, 4) = 0 THEN 74
            WHEN MOD(s.id, 4) = 1 THEN 70
            ELSE 64 END
FROM t_student s
JOIN t_class cl ON cl.id = s.class_id
JOIN t_user u ON u.id = s.user_id
WHERE u.id BETWEEN 2387 AND 2706 AND MOD(s.id, 4) = 0;
