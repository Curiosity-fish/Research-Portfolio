-- V10: 重置为“24级应用统计”最小数据集
-- 1 个专业、2 个班级、2 名班主任、80 名学生及原始成绩/体测/考勤/志愿/心理/竞赛数据
-- 清空此前演示与多样化数据；职位/院校缓存属于发展引导参考数据，保留不动

DELETE FROM t_grade;
DELETE FROM t_gpa_history;
DELETE FROM t_profile_score;
DELETE FROM t_alert;
DELETE FROM t_intervention_record;
DELETE FROM t_health_report;
DELETE FROM t_competition;
DELETE FROM t_volunteer;
DELETE FROM t_attendance;
DELETE FROM t_psychology;
DELETE FROM t_major_plan;
DELETE FROM t_course_class;
DELETE FROM t_course;
DELETE FROM t_student;
DELETE FROM t_class;
DELETE FROM t_teacher;
DELETE FROM t_user;
DELETE FROM t_major;
DELETE FROM t_department;

-- ============ 账号（80 名学生 + 2 班主任 + 1 系主任 + 1 院长，密码均为 123456） ============

INSERT INTO t_user (id, account, name, password, role, college, major, class_name, grade)
WITH RECURSIVE seq AS (SELECT 1 n UNION ALL SELECT n + 1 FROM seq WHERE n < 80)
SELECT n,
       CONCAT(CASE WHEN n <= 40 THEN '202610' ELSE '202620' END,
              LPAD(CASE WHEN n <= 40 THEN n ELSE n - 40 END, 2, '0')),
       CONCAT('学生', LPAD(n, 2, '0')),
       '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO',
       'student', '数学与计算机科学学院', '应用统计',
       CASE WHEN n <= 40 THEN '应统2401' ELSE '应统2402' END,
       '2024'
FROM seq;

INSERT INTO t_user (id, account, name, password, role, college, major, class_name, grade) VALUES
(81, 'T20180042', '王老师', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'teacher', '数学与计算机科学学院', NULL, NULL, NULL),
(82, 'T20180043', '李老师', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'teacher', '数学与计算机科学学院', NULL, NULL, NULL),
(83, 'D20210001', '陈主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'department', '数学与计算机科学学院', '应用统计', NULL, NULL),
(84, 'L20210001', '张院长', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'dean', '数学与计算机科学学院', NULL, NULL, NULL);

-- ============ 组织架构 ============

INSERT INTO t_department (id, name, code, dean_id) VALUES
(1, '数学与计算机科学学院', 'MATH_CS', 84);

INSERT INTO t_major (id, name, code, department_id, degree_type) VALUES
(1, '应用统计', 'APP_STAT', 1, '理学学士');

INSERT INTO t_teacher (id, teacher_id, user_id, department_id, title, is_class_advisor, class_id) VALUES
(1, 'T20180042', 81, 1, '副教授', 1, 1),
(2, 'T20180043', 82, 1, '讲师', 1, 2);

INSERT INTO t_class (id, name, grade, major_id, advisor_id, total_students) VALUES
(1, '应统2401', '2024', 1, 1, 40),
(2, '应统2402', '2024', 1, 2, 40);

INSERT INTO t_student (student_id, user_id, class_id, major_id, department_id, gpa, `rank`, total_students, alert_level, total_credits, required_credits)
SELECT u.account, u.id, cl.id, 1, 1,
       ROUND(2.0 + MOD(u.id * 7, 200) / 100, 2),
       MOD(u.id * 3, 40) + 1, 40, 'none',
       40 + MOD(u.id * 5, 60), 160.0
FROM t_user u
JOIN t_class cl ON cl.id = CASE WHEN u.id <= 40 THEN 1 ELSE 2 END
WHERE u.role = 'student';

-- ============ 课程与开课 ============

INSERT INTO t_course (id, code, name, credits, type, department_id) VALUES
(1, 'MA201', '数学分析', 4.0, 'required', 1),
(2, 'ST201', '概率论与数理统计', 3.5, 'required', 1),
(3, 'ST202', '统计学导论', 3.0, 'required', 1),
(4, 'ST203', '应用回归分析', 3.0, 'elective', 1),
(5, 'EN104', '大学英语（四）', 2.0, 'public', 1),
(6, 'CS102', '计算机基础', 3.0, 'public', 1);

INSERT INTO t_course_class (course_id, term, class_name, teacher_id, max_students, enrolled_students, status)
SELECT c.id, '2024-2025-1', cl.name, CASE WHEN cl.id = 1 THEN 1 ELSE 2 END, 40, 40, 'active'
FROM t_course c
CROSS JOIN t_class cl;

-- ============ 原始成绩（模拟校方数据） ============

INSERT INTO t_grade (student_id, course_class_id, score, status, term, exam_date)
SELECT s.id, cc.id,
       CASE WHEN MOD(s.id + cc.course_id, 7) = 0 THEN 45 + MOD(s.id + cc.id, 15)
            ELSE 58 + MOD(s.id * 7 + cc.course_id * 11, 42)
       END,
       'passed', '2024-2025-1', '2025-01-15'
FROM t_student s
JOIN t_class cl ON cl.id = s.class_id
JOIN t_course_class cc ON cc.class_name = cl.name AND cc.term = '2024-2025-1';

UPDATE t_grade g
JOIN t_course_class cc ON cc.id = g.course_class_id
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
WHERE g.term = '2024-2025-1';

-- ============ 体测与校方补充数据 ============

INSERT INTO t_health_report (student_id, term, total_score, items, report_url)
SELECT s.id, '2024-2025-1', 60 + MOD(s.id * 5, 36),
       JSON_ARRAY(
           JSON_OBJECT('name', '体重指数', 'score', 60 + MOD(s.id * 3, 36), 'level', '良好'),
           JSON_OBJECT('name', '肺活量', 'score', 60 + MOD(s.id * 5, 36), 'level', '良好'),
           JSON_OBJECT('name', '耐力跑', 'score', 60 + MOD(s.id * 7, 36), 'level', '良好')
       ),
       NULL
FROM t_student s;

INSERT INTO t_attendance (student_id, term, total_classes, absent_count, late_count, note)
SELECT s.id, '2024-2025-1', 60, MOD(s.id * 3, 12), MOD(s.id, 5), '模拟考勤数据'
FROM t_student s;

INSERT INTO t_volunteer (student_id, term, hours, description)
SELECT s.id, '2024-2025-1', 6 + MOD(s.id * 7, 60), '模拟志愿服务记录'
FROM t_student s;

INSERT INTO t_psychology (student_id, term, score, level, note)
SELECT s.id, '2024-2025-1',
       62 + MOD(s.id * 5, 36),
       CASE WHEN 62 + MOD(s.id * 5, 36) >= 85 THEN 'good'
            WHEN 62 + MOD(s.id * 5, 36) >= 70 THEN 'normal'
            ELSE 'attention' END,
       '模拟心理测评记录'
FROM t_student s;

INSERT INTO t_competition (student_id, term, competition_name, level, award, points)
SELECT s.id, '2024-2025-1',
       '应用统计建模大赛',
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
WHERE MOD(s.id, 4) = 0;

-- ============ 培养方案（仅应用统计） ============

INSERT INTO t_major_plan (major_id, grade, dimensions, standard, actual, tips) VALUES
(1, '2024', '["公共基础","专业必修","专业选修","实践环节","通识教育"]', '[48,62,20,18,12]', '[20,30,10,6,8]', '["大一基础阶段完成度符合预期","建议关注数学与统计基础"]');
