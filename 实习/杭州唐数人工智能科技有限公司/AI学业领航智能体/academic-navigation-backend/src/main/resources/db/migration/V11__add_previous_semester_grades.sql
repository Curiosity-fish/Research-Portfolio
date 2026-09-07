-- V11: 假设 80 名应用统计学生为 2021 级（现为大四），
-- 补充大一至大三 6 个学期的课程成绩，使 GPA 趋势完整（大一上 ~ 大四上 共 7 个学期）

UPDATE t_user SET grade = '2021' WHERE role = 'student' AND grade = '2024';
UPDATE t_class SET grade = '2021' WHERE grade = '2024';

-- 前 6 个学期开课记录
INSERT INTO t_course_class (course_id, term, class_name, teacher_id, max_students, enrolled_students, status)
SELECT c.id, t.term, cl.name, CASE WHEN cl.id = 1 THEN 1 ELSE 2 END, 40, 40, 'finished'
FROM t_course c
CROSS JOIN t_class cl
CROSS JOIN (
    SELECT '2021-2022-1' term UNION ALL SELECT '2021-2022-2'
    UNION ALL SELECT '2022-2023-1' UNION ALL SELECT '2022-2023-2'
    UNION ALL SELECT '2023-2024-1' UNION ALL SELECT '2023-2024-2'
) t
WHERE NOT EXISTS (
    SELECT 1 FROM t_course_class cc
    WHERE cc.course_id = c.id AND cc.class_name = cl.name AND cc.term = t.term
);

-- 前 6 个学期成绩（模拟校方原始成绩）
INSERT INTO t_grade (student_id, course_class_id, score, status, term, exam_date)
SELECT s.id, cc.id,
       CASE WHEN MOD(s.id + cc.id, 11) = 0 THEN 45 + MOD(s.id + cc.id, 12)
            ELSE 60 + MOD(s.id * 13 + cc.id * 7, 40)
       END,
       'passed', cc.term, '2025-01-15'
FROM t_student s
JOIN t_class cl ON cl.id = s.class_id
JOIN t_course_class cc ON cc.class_name = cl.name
WHERE cc.term IN ('2021-2022-1','2021-2022-2','2022-2023-1','2022-2023-2','2023-2024-1','2023-2024-2');

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
WHERE cc.term IN ('2021-2022-1','2021-2022-2','2022-2023-1','2022-2023-2','2023-2024-1','2023-2024-2');
