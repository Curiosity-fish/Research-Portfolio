-- V13: 恢复任课教师账号并分配课程

INSERT IGNORE INTO t_user (id, account, name, password, role, college, major, class_name, grade) VALUES
(85, 'C20180099', '林老师', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'course_teacher', '数学与计算机科学学院', NULL, NULL, NULL);

INSERT IGNORE INTO t_teacher (id, teacher_id, user_id, department_id, title, is_class_advisor, class_id) VALUES
(3, 'C20180099', 85, 1, '讲师', 0, NULL);

-- 将部分课程班级划给任课教师（概率论、统计学导论、计算机基础）
UPDATE t_course_class SET teacher_id = 3 WHERE course_id IN (2, 3, 6) AND teacher_id IN (1, 2);
