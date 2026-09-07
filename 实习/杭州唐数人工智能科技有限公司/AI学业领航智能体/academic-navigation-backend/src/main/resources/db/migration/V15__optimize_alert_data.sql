-- V15: 优化成绩与考勤分布，避免 80 名学生全员预警
-- 大部分学生成绩调整到通过区间，仅部分学生保留挂科；考勤拉开高低分布

-- 1) 全部成绩先调整为“通过为主”的区间（60-95）
UPDATE t_grade g
JOIN t_course_class cc ON cc.id = g.course_class_id
SET g.score = 70 + MOD(g.score, 26);

-- 2) 仅部分学生保留挂科（每名选定学生挂 1-2 门课）
UPDATE t_grade g
JOIN t_student s ON s.id = g.student_id
JOIN t_course_class cc ON cc.id = g.course_class_id
SET g.score = 42 + MOD(g.id, 15)
WHERE MOD(s.id, 8) IN (0, 3)
  AND cc.course_id IN (1 + MOD(s.id, 6), 1 + MOD(s.id + 3, 6))
  AND cc.term = '2024-2025-1';

-- 3) 按分数统一重算状态与绩点
UPDATE t_grade
SET status = IF(score >= 60, 'passed', 'failed'),
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

-- 4) 考勤分布优化：多数学生缺勤少，少数学生缺勤多
UPDATE t_attendance
SET absent_count = CASE
        WHEN MOD(student_id, 6) = 0 THEN 6 + MOD(student_id, 4)
        WHEN MOD(student_id, 13) = 0 THEN 9 + MOD(student_id, 3)
        ELSE MOD(student_id, 5)
    END,
    late_count = MOD(student_id, 5),
    note = '模拟考勤数据（已优化分布）';
