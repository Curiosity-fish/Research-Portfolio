-- V18: 班主任与任课教师彻底分离
-- 移除班主任对应的 C 工号，新增 7 位独立普通任课教师，24 门课全部由独立任课教师承担

-- 1) 新增 7 位普通任课教师
INSERT INTO t_user (id, account, name, password, role, college, major, class_name, grade) VALUES
(343, 'C20180050', '周老师', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'course_teacher', '数学与计算机科学学院', NULL, NULL, NULL),
(344, 'C20180051', '吴老师', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'course_teacher', '数学与计算机科学学院', NULL, NULL, NULL),
(345, 'C20180052', '郑老师', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'course_teacher', '数学与计算机科学学院', NULL, NULL, NULL),
(346, 'C20180053', '王老师', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'course_teacher', '数学与计算机科学学院', NULL, NULL, NULL),
(347, 'C20180054', '赵老师', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'course_teacher', '数学与计算机科学学院', NULL, NULL, NULL),
(348, 'C20180055', '孙老师', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'course_teacher', '数学与计算机科学学院', NULL, NULL, NULL),
(349, 'C20180056', '钱老师', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'course_teacher', '数学与计算机科学学院', NULL, NULL, NULL);

INSERT INTO t_teacher (id, teacher_id, user_id, department_id, title, is_class_advisor, class_id) VALUES
(18, 'C20180050', 343, 1, '讲师', 0, NULL),
(19, 'C20180051', 344, 1, '讲师', 0, NULL),
(20, 'C20180052', 345, 1, '讲师', 0, NULL),
(21, 'C20180053', 346, 1, '讲师', 0, NULL),
(22, 'C20180054', 347, 1, '讲师', 0, NULL),
(23, 'C20180055', 348, 1, '讲师', 0, NULL),
(24, 'C20180056', 349, 1, '讲师', 0, NULL);

-- 2) 课程班级全部改挂到独立任课教师
UPDATE t_course_class SET teacher_id = 18 WHERE teacher_id IN (4, 5);
UPDATE t_course_class SET teacher_id = 19 WHERE teacher_id = 12;
UPDATE t_course_class SET teacher_id = 20 WHERE teacher_id = 13;
UPDATE t_course_class SET teacher_id = 21 WHERE teacher_id = 14;
UPDATE t_course_class SET teacher_id = 22 WHERE teacher_id = 15;
UPDATE t_course_class SET teacher_id = 23 WHERE teacher_id = 16;
UPDATE t_course_class SET teacher_id = 24 WHERE teacher_id = 17;

-- 3) 删除班主任对应的 C 工号（班主任不再授课）
DELETE FROM t_teacher WHERE id IN (4, 5, 12, 13, 14, 15, 16, 17);
DELETE FROM t_user WHERE account IN ('C20180042', 'C20180043', 'C2026001', 'C2026002', 'C2026003', 'C2026004', 'C2026005', 'C2026006');
