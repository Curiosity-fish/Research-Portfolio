-- V16: 从 1 个专业扩展为同学院下 4 个专业，每个专业 2 个班
-- 应用统计 / 数据科学 / 人工智能 / 软件工程，共 8 个班、320 名学生

-- 1) 新增专业
INSERT INTO t_major (id, name, code, department_id, degree_type) VALUES
(2, '数据科学', 'DS', 1, '理学学士'),
(3, '人工智能', 'AI', 1, '工学学士'),
(4, '软件工程', 'SE', 1, '工学学士');

-- 2) 新增班主任账号（6 位，对应 6 个新班级）
INSERT INTO t_user (id, account, name, password, role, college, major, class_name, grade) VALUES
(328, 'T2026001', '数据一班班主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'teacher', '数学与计算机科学学院', NULL, NULL, NULL),
(329, 'T2026002', '数据二班班主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'teacher', '数学与计算机科学学院', NULL, NULL, NULL),
(330, 'T2026003', '智能一班班主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'teacher', '数学与计算机科学学院', NULL, NULL, NULL),
(331, 'T2026004', '智能二班班主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'teacher', '数学与计算机科学学院', NULL, NULL, NULL),
(332, 'T2026005', '软件一班班主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'teacher', '数学与计算机科学学院', NULL, NULL, NULL),
(333, 'T2026006', '软件二班班主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'teacher', '数学与计算机科学学院', NULL, NULL, NULL);

INSERT INTO t_teacher (id, teacher_id, user_id, department_id, title, is_class_advisor, class_id) VALUES
(6, 'T2026001', 328, 1, '讲师', 1, 3),
(7, 'T2026002', 329, 1, '讲师', 1, 4),
(8, 'T2026003', 330, 1, '讲师', 1, 5),
(9, 'T2026004', 331, 1, '讲师', 1, 6),
(10, 'T2026005', 332, 1, '讲师', 1, 7),
(11, 'T2026006', 333, 1, '讲师', 1, 8);

-- 3) 新增班级
INSERT INTO t_class (id, name, grade, major_id, advisor_id, total_students) VALUES
(3, '数科2401', '2021', 2, 6, 40),
(4, '数科2402', '2021', 2, 7, 40),
(5, '智能2401', '2021', 3, 8, 40),
(6, '智能2402', '2021', 3, 9, 40),
(7, '软工2401', '2021', 4, 10, 40),
(8, '软工2402', '2021', 4, 11, 40);

-- 4) 新增 240 名学生（3 个专业 × 2 个班 × 40 人）
INSERT INTO t_user (account, name, password, role, college, major, class_name, grade)
WITH RECURSIVE seq AS (SELECT 1 n UNION ALL SELECT n + 1 FROM seq WHERE n < 240)
SELECT CONCAT(
           CASE
               WHEN n <= 40 THEN '202630'
               WHEN n <= 80 THEN '202640'
               WHEN n <= 120 THEN '202650'
               WHEN n <= 160 THEN '202660'
               WHEN n <= 200 THEN '202670'
               ELSE '202680'
           END,
           LPAD(CASE
               WHEN n <= 40 THEN n
               WHEN n <= 80 THEN n - 40
               WHEN n <= 120 THEN n - 80
               WHEN n <= 160 THEN n - 120
               WHEN n <= 200 THEN n - 160
               ELSE n - 200
           END, 2, '0')
       ),
       CONCAT('学生', LPAD(n, 3, '0')),
       '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO',
       'student', '数学与计算机科学学院',
       CASE WHEN n <= 80 THEN '数据科学'
            WHEN n <= 160 THEN '人工智能'
            ELSE '软件工程' END,
       CASE WHEN n <= 40 THEN '数科2401'
            WHEN n <= 80 THEN '数科2402'
            WHEN n <= 120 THEN '智能2401'
            WHEN n <= 160 THEN '智能2402'
            WHEN n <= 200 THEN '软工2401'
            ELSE '软工2402' END,
       '2021'
FROM seq;

INSERT INTO t_student (student_id, user_id, class_id, major_id, department_id, gpa, `rank`, total_students, alert_level, total_credits, required_credits)
SELECT u.account, u.id, c.id, c.major_id, 1,
       ROUND(2.0 + MOD(u.id * 7, 200) / 100, 2),
       MOD(u.id * 3, 40) + 1, 80, 'none',
       40 + MOD(u.id * 5, 60), 160.0
FROM t_user u
JOIN t_class c ON c.name = u.class_name
WHERE u.role = 'student' AND u.id >= 88;

-- 5) 新班级的开课记录（7 个学期 × 6 门课；课程按原任课教师分配）
INSERT INTO t_course_class (course_id, term, class_name, teacher_id, max_students, enrolled_students, status)
SELECT c.id, t.term, cl.name,
       CASE WHEN c.id IN (2, 3, 6) THEN 3
            WHEN c.id IN (1, 4) THEN 4
            ELSE 5 END,
       40, 40,
       CASE WHEN t.term = '2024-2025-1' THEN 'active' ELSE 'finished' END
FROM t_course c
CROSS JOIN t_class cl
CROSS JOIN (
    SELECT '2021-2022-1' term UNION ALL SELECT '2021-2022-2'
    UNION ALL SELECT '2022-2023-1' UNION ALL SELECT '2022-2023-2'
    UNION ALL SELECT '2023-2024-1' UNION ALL SELECT '2023-2024-2'
    UNION ALL SELECT '2024-2025-1'
) t
WHERE cl.id IN (3, 4, 5, 6, 7, 8);

-- 6) 新学生成绩（7 个学期）
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
WHERE u.account LIKE '20263%' OR u.account LIKE '20264%'
   OR u.account LIKE '20265%' OR u.account LIKE '20266%'
   OR u.account LIKE '20267%' OR u.account LIKE '20268%';

UPDATE t_grade g
JOIN t_course_class cc ON cc.id = g.course_class_id
JOIN t_student s ON s.id = g.student_id
JOIN t_user u ON u.id = s.user_id
SET g.score = 42 + MOD(g.id, 15)
WHERE (u.account LIKE '20263%' OR u.account LIKE '20264%'
   OR u.account LIKE '20265%' OR u.account LIKE '20266%'
   OR u.account LIKE '20267%' OR u.account LIKE '20268%')
  AND cc.term = '2024-2025-1'
  AND MOD(s.id, 8) IN (0, 3)
  AND cc.course_id IN (1 + MOD(s.id, 6), 1 + MOD(s.id + 3, 6));

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
WHERE u.account LIKE '20263%' OR u.account LIKE '20264%'
   OR u.account LIKE '20265%' OR u.account LIKE '20266%'
   OR u.account LIKE '20267%' OR u.account LIKE '20268%';

-- 7) 新学生体测/考勤/志愿/心理
INSERT INTO t_health_report (student_id, term, total_score, items, report_url)
SELECT s.id, '2024-2025-1', 60 + MOD(s.id * 5, 36),
       JSON_ARRAY(
           JSON_OBJECT('name', '体重指数', 'score', 60 + MOD(s.id * 3, 36), 'level', '良好'),
           JSON_OBJECT('name', '肺活量', 'score', 60 + MOD(s.id * 5, 36), 'level', '良好'),
           JSON_OBJECT('name', '耐力跑', 'score', 60 + MOD(s.id * 7, 36), 'level', '良好')
       ), NULL
FROM t_student s
JOIN t_user u ON u.id = s.user_id
WHERE u.account LIKE '20263%' OR u.account LIKE '20264%'
   OR u.account LIKE '20265%' OR u.account LIKE '20266%'
   OR u.account LIKE '20267%' OR u.account LIKE '20268%';

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
WHERE u.account LIKE '20263%' OR u.account LIKE '20264%'
   OR u.account LIKE '20265%' OR u.account LIKE '20266%'
   OR u.account LIKE '20267%' OR u.account LIKE '20268%';

INSERT INTO t_volunteer (student_id, term, hours, description)
SELECT s.id, '2024-2025-1', 6 + MOD(s.id * 7, 60), '模拟志愿服务记录'
FROM t_student s
JOIN t_user u ON u.id = s.user_id
WHERE u.account LIKE '20263%' OR u.account LIKE '20264%'
   OR u.account LIKE '20265%' OR u.account LIKE '20266%'
   OR u.account LIKE '20267%' OR u.account LIKE '20268%';

INSERT INTO t_psychology (student_id, term, score, level, note)
SELECT s.id, '2024-2025-1',
       62 + MOD(s.id * 5, 36),
       CASE WHEN 62 + MOD(s.id * 5, 36) >= 85 THEN 'good'
            WHEN 62 + MOD(s.id * 5, 36) >= 70 THEN 'normal'
            ELSE 'attention' END,
       '模拟心理测评记录'
FROM t_student s
JOIN t_user u ON u.id = s.user_id
WHERE u.account LIKE '20263%' OR u.account LIKE '20264%'
   OR u.account LIKE '20265%' OR u.account LIKE '20266%'
   OR u.account LIKE '20267%' OR u.account LIKE '20268%';

-- 8) 新增 3 位系主任（每个专业一位）
INSERT INTO t_user (id, account, name, password, role, college, major, class_name, grade) VALUES
(334, 'D20210002', '数据系主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'department', '数学与计算机科学学院', '数据科学', NULL, NULL),
(335, 'D20210003', '智能系主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'department', '数学与计算机科学学院', '人工智能', NULL, NULL),
(336, 'D20210004', '软件系主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'department', '数学与计算机科学学院', '软件工程', NULL, NULL);

-- 9) 新增专业培养方案
INSERT INTO t_major_plan (major_id, grade, dimensions, standard, actual, tips) VALUES
(2, '2021', '["公共基础","专业必修","专业选修","实践环节","通识教育"]', '[48,62,20,18,12]', '[46,58,24,16,12]', '["公共基础学分完成度略低","建议加强实践环节"]'),
(3, '2021', '["公共基础","专业必修","专业选修","实践环节","通识教育"]', '[48,62,20,18,12]', '[45,56,26,15,12]', '["专业必修完成度偏低","建议加强算法与数学基础"]'),
(4, '2021', '["公共基础","专业必修","专业选修","实践环节","通识教育"]', '[48,62,20,18,12]', '[47,60,22,17,12]', '["培养方案整体完成良好","建议增加工程实践项目"]');
