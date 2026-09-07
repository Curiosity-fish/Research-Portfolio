-- V17: 课程从 6 门扩充到 24 门（4 个专业各 6 门）
-- 并为 6 个新班级班主任补充对应 C 工号任课教师账号

-- 1) 新增 18 门课程
INSERT INTO t_course (id, code, name, credits, type, department_id) VALUES
(7,  'DS201', '数据挖掘导论', 3.0, 'required', 1),
(8,  'DS202', '机器学习基础', 3.5, 'required', 1),
(9,  'DS203', '数据库原理', 3.5, 'required', 1),
(10, 'DS204', '数值分析', 3.0, 'required', 1),
(11, 'DS205', '数据可视化', 2.5, 'elective', 1),
(12, 'DS206', '大数据技术', 3.0, 'elective', 1),
(13, 'AI201', '人工智能导论', 3.0, 'required', 1),
(14, 'AI202', '深度学习基础', 3.5, 'required', 1),
(15, 'AI203', '自然语言处理', 3.0, 'required', 1),
(16, 'AI204', '计算机视觉', 3.0, 'required', 1),
(17, 'AI205', '知识图谱', 2.5, 'elective', 1),
(18, 'AI206', '强化学习导论', 3.0, 'elective', 1),
(19, 'SE201', '软件工程导论', 3.0, 'required', 1),
(20, 'SE202', '面向对象程序设计', 3.5, 'required', 1),
(21, 'SE203', '数据库系统', 3.5, 'required', 1),
(22, 'SE204', '软件测试', 2.5, 'required', 1),
(23, 'SE205', '项目管理', 2.5, 'elective', 1),
(24, 'SE206', 'Web开发技术', 3.0, 'elective', 1);

-- 2) 新增 6 位 C 工号任课教师（对应 6 个新班级班主任）
INSERT INTO t_user (id, account, name, password, role, college, major, class_name, grade) VALUES
(337, 'C2026001', '数科一班任课', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'course_teacher', '数学与计算机科学学院', '数据科学', NULL, NULL),
(338, 'C2026002', '数科二班任课', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'course_teacher', '数学与计算机科学学院', '数据科学', NULL, NULL),
(339, 'C2026003', '智能一班任课', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'course_teacher', '数学与计算机科学学院', '人工智能', NULL, NULL),
(340, 'C2026004', '智能二班任课', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'course_teacher', '数学与计算机科学学院', '人工智能', NULL, NULL),
(341, 'C2026005', '软件一班任课', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'course_teacher', '数学与计算机科学学院', '软件工程', NULL, NULL),
(342, 'C2026006', '软件二班任课', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'course_teacher', '数学与计算机科学学院', '软件工程', NULL, NULL);

INSERT INTO t_teacher (id, teacher_id, user_id, department_id, title, is_class_advisor, class_id) VALUES
(12, 'C2026001', 337, 1, '讲师', 0, NULL),
(13, 'C2026002', 338, 1, '讲师', 0, NULL),
(14, 'C2026003', 339, 1, '讲师', 0, NULL),
(15, 'C2026004', 340, 1, '讲师', 0, NULL),
(16, 'C2026005', 341, 1, '讲师', 0, NULL),
(17, 'C2026006', 342, 1, '讲师', 0, NULL);

-- 3) 替换新班级的旧课程与成绩（原 6 门通用课改为专业课程）
DELETE FROM t_grade WHERE student_id IN (
    SELECT id FROM t_student WHERE class_id IN (3, 4, 5, 6, 7, 8)
);

DELETE FROM t_course_class WHERE class_name IN (
    SELECT name FROM t_class WHERE id IN (3, 4, 5, 6, 7, 8)
) AND course_id <= 6;

-- 4) 新班级按专业插入课程开课（7 学期）
INSERT INTO t_course_class (course_id, term, class_name, teacher_id, max_students, enrolled_students, status)
SELECT c.id, t.term, cl.name,
       CASE
           WHEN c.id IN (7, 8, 9) THEN 12
           WHEN c.id IN (10, 11, 12) THEN 13
           WHEN c.id IN (13, 14, 15) THEN 14
           WHEN c.id IN (16, 17, 18) THEN 15
           WHEN c.id IN (19, 20, 21) THEN 16
           ELSE 17
       END,
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
WHERE cl.id IN (3, 4, 5, 6, 7, 8)
  AND c.id BETWEEN 7 AND 24
  AND (
      (cl.id IN (3, 4) AND c.id BETWEEN 7 AND 12)
      OR (cl.id IN (5, 6) AND c.id BETWEEN 13 AND 18)
      OR (cl.id IN (7, 8) AND c.id BETWEEN 19 AND 24)
  );

-- 5) 新学生按专业课程生成 7 学期成绩
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
WHERE (u.account LIKE '20263%' OR u.account LIKE '20264%'
   OR u.account LIKE '20265%' OR u.account LIKE '20266%'
   OR u.account LIKE '20267%' OR u.account LIKE '20268%')
  AND cc.course_id BETWEEN 7 AND 24;

-- 6) 当前学期部分学生保留挂科
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
  AND cc.course_id IN (7 + MOD(s.id, 6), 7 + MOD(s.id + 3, 6));

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
