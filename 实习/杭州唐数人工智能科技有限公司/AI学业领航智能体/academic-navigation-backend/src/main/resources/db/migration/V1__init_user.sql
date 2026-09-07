-- V1: 用户主表初始化
-- 运行时机：服务首次启动时由 Flyway 自动执行

CREATE TABLE IF NOT EXISTS t_user (
    id          BIGINT          NOT NULL AUTO_INCREMENT COMMENT '主键',
    account     VARCHAR(30)     NOT NULL                COMMENT '学号/工号，唯一标识',
    name        VARCHAR(50)     NOT NULL                COMMENT '姓名',
    password    VARCHAR(100)    NOT NULL                COMMENT 'BCrypt 加密后的密码',
    role        VARCHAR(20)     NOT NULL                COMMENT '角色: student/teacher/department/dean',
    college     VARCHAR(100)    DEFAULT NULL            COMMENT '学院',
    major       VARCHAR(100)    DEFAULT NULL            COMMENT '专业',
    class_name  VARCHAR(50)     DEFAULT NULL            COMMENT '班级',
    grade       VARCHAR(10)     DEFAULT NULL            COMMENT '年级，如 2022',
    avatar_url  VARCHAR(255)    DEFAULT NULL            COMMENT '头像地址',
    is_active   TINYINT(1)      NOT NULL DEFAULT 1      COMMENT '是否启用: 1启用 0禁用',
    created_at  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_account (account),
    KEY idx_role (role)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户主表';

-- 密码均为 123456，BCrypt(strength=12) 加密
INSERT INTO t_user (account, name, password, role, college, major, class_name, grade) VALUES
('2022001',   '李明',   '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'student',    '数学与计算机科学学院', '计算机科学与技术', '计科2201', '2022'),
('T20180042', '张老师', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'teacher',    '数学与计算机科学学院', NULL,          NULL,       NULL),
('D20210001', '王主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'department', '数学与计算机科学学院', '计算机科学与技术', NULL,       NULL),
('L20210001', '陈院长', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'dean',       '数学与计算机科学学院', NULL,          NULL,       NULL);
