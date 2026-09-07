-- V5: 补齐按专业/学院口径查看数据所需的系主任与院长账号
-- 密码均为 123456（与 V1-V4 种子一致）

INSERT IGNORE INTO t_user (account, name, password, role, college, major, class_name, grade) VALUES
('D20210002', '软件系主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'department', '数学与计算机科学学院', '软件工程', NULL, NULL),
('D20210003', '数据系主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'department', '数学与计算机科学学院', '数据科学', NULL, NULL),
('D2024003', '计算系主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'department', '数据科学与工程测试学院', '计算机科学与技术', NULL, NULL),
('D2024004', '智能系主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'department', '智能计算测试学院', '人工智能', NULL, NULL),
('L2024002', '智能计算院长', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'dean', '智能计算测试学院', NULL, NULL, NULL);

-- 现有测试学院系主任账号补齐专业归属
UPDATE t_user SET major = '数据科学与大数据技术' WHERE account = 'D2024001' AND major IS NULL;
UPDATE t_user SET major = '软件工程' WHERE account = 'D2024002' AND major IS NULL;

-- 智能计算测试学院挂接独立院长账号
UPDATE t_department
SET dean_id = (SELECT id FROM (SELECT id FROM t_user WHERE account = 'L2024002') t)
WHERE id = 100101 AND dean_id IS NOT NULL;
