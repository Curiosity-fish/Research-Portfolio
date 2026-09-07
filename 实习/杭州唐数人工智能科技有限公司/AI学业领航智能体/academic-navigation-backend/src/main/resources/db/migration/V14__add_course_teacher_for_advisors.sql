-- V14: 两位班主任同时以任课教师身份授课，补充对应 course_teacher 工号

INSERT IGNORE INTO t_user (id, account, name, password, role, college, major, class_name, grade) VALUES
(86, 'C20180042', '王老师', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'course_teacher', '数学与计算机科学学院', NULL, NULL, NULL),
(87, 'C20180043', '李老师', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'course_teacher', '数学与计算机科学学院', NULL, NULL, NULL);

INSERT IGNORE INTO t_teacher (id, teacher_id, user_id, department_id, title, is_class_advisor, class_id) VALUES
(4, 'C20180042', 86, 1, '副教授', 0, NULL),
(5, 'C20180043', 87, 1, '讲师', 0, NULL);

-- 班主任原授课课程改挂到对应任课教师账号（数学分析、应用回归分析、大学英语四）
UPDATE t_course_class SET teacher_id = 4 WHERE teacher_id = 1 AND course_id IN (1, 4, 5);
UPDATE t_course_class SET teacher_id = 5 WHERE teacher_id = 2 AND course_id IN (1, 4, 5);
