-- V2: 新增任课老师角色用户
-- 密码均为 123456，BCrypt(strength=12) 加密

INSERT INTO t_user (account, name, password, role, college, major, class_name, grade)
VALUES ('C20180099', '林老师', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'course_teacher', '数学与计算机科学学院', NULL, NULL, NULL);
