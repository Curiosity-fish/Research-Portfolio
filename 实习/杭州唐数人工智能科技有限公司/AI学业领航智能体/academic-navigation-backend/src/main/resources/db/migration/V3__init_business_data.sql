-- V3: M1/M2 核心业务表与演示种子数据
-- AI 智能体、职位/院校缓存等表推迟到 M3/M4 建表，避免 M1 范围膨胀

CREATE TABLE IF NOT EXISTS t_department (
    id          BIGINT       NOT NULL AUTO_INCREMENT COMMENT '主键',
    name        VARCHAR(100) NOT NULL                COMMENT '学院名称',
    code        VARCHAR(50)  DEFAULT NULL            COMMENT '学院编码',
    dean_id     BIGINT       DEFAULT NULL            COMMENT '院长用户ID',
    created_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_department_name (name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='学院';

CREATE TABLE IF NOT EXISTS t_major (
    id             BIGINT       NOT NULL AUTO_INCREMENT COMMENT '主键',
    name           VARCHAR(100) NOT NULL                COMMENT '专业名称',
    code           VARCHAR(50)  DEFAULT NULL            COMMENT '专业编码',
    department_id  BIGINT       NOT NULL                COMMENT '所属学院ID',
    degree_type    VARCHAR(50)  DEFAULT NULL            COMMENT '学位类型',
    created_at     DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_major_department (department_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='专业';

CREATE TABLE IF NOT EXISTS t_teacher (
    id               BIGINT      NOT NULL AUTO_INCREMENT COMMENT '主键',
    teacher_id       VARCHAR(30) NOT NULL                COMMENT '教师工号',
    user_id          BIGINT      NOT NULL                COMMENT '用户ID',
    department_id    BIGINT      DEFAULT NULL            COMMENT '所属学院ID',
    title            VARCHAR(50) DEFAULT NULL            COMMENT '职称',
    is_class_advisor TINYINT(1)  NOT NULL DEFAULT 0      COMMENT '是否班主任',
    class_id         BIGINT      DEFAULT NULL            COMMENT '负责班级ID',
    created_at       DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at       DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_teacher_id (teacher_id),
    KEY idx_teacher_user (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='教师档案';

CREATE TABLE IF NOT EXISTS t_class (
    id             BIGINT      NOT NULL AUTO_INCREMENT COMMENT '主键',
    name           VARCHAR(50) NOT NULL                COMMENT '班级名称',
    grade          VARCHAR(10) DEFAULT NULL            COMMENT '年级，如 2022',
    major_id       BIGINT      NOT NULL                COMMENT '专业ID',
    advisor_id     BIGINT      DEFAULT NULL            COMMENT '班主任教师ID',
    total_students INT         NOT NULL DEFAULT 0      COMMENT '班级总人数',
    created_at     DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_class_name (name),
    KEY idx_class_major (major_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='班级';

CREATE TABLE IF NOT EXISTS t_student (
    id               BIGINT       NOT NULL AUTO_INCREMENT COMMENT '主键',
    student_id       VARCHAR(30)  NOT NULL                COMMENT '学号',
    user_id          BIGINT       NOT NULL                COMMENT '用户ID',
    class_id         BIGINT       DEFAULT NULL            COMMENT '班级ID',
    major_id         BIGINT       DEFAULT NULL            COMMENT '专业ID',
    department_id    BIGINT       DEFAULT NULL            COMMENT '学院ID',
    gpa              DECIMAL(4,2) DEFAULT NULL            COMMENT '当前GPA',
    `rank`           INT          DEFAULT NULL            COMMENT '专业排名',
    total_students   INT          DEFAULT NULL            COMMENT '专业总人数',
    alert_level      VARCHAR(10)  NOT NULL DEFAULT 'none' COMMENT '预警等级: none/yellow/orange/red',
    total_credits    DECIMAL(5,1) DEFAULT NULL            COMMENT '已修学分',
    required_credits DECIMAL(5,1) DEFAULT NULL            COMMENT '要求总学分',
    created_at       DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at       DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_student_no (student_id),
    KEY idx_student_user (user_id),
    KEY idx_student_class (class_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='学生档案';

CREATE TABLE IF NOT EXISTS t_course (
    id             BIGINT       NOT NULL AUTO_INCREMENT COMMENT '主键',
    code           VARCHAR(30)  NOT NULL                COMMENT '课程编码',
    name           VARCHAR(100) NOT NULL                COMMENT '课程名称',
    credits        DECIMAL(4,1) NOT NULL DEFAULT 0      COMMENT '学分',
    type           VARCHAR(20)  NOT NULL DEFAULT 'required' COMMENT 'required/elective/public',
    department_id  BIGINT       DEFAULT NULL            COMMENT '开课学院ID',
    created_at     DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_course_code (code),
    KEY idx_course_department (department_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='课程';

CREATE TABLE IF NOT EXISTS t_course_class (
    id                 BIGINT      NOT NULL AUTO_INCREMENT COMMENT '主键',
    course_id          BIGINT      NOT NULL                COMMENT '课程ID',
    term               VARCHAR(20) NOT NULL                COMMENT '学期，如 2024-2025-1',
    class_name         VARCHAR(50) DEFAULT NULL            COMMENT '上课班级',
    teacher_id         BIGINT      DEFAULT NULL            COMMENT '任课教师ID',
    max_students       INT         NOT NULL DEFAULT 0      COMMENT '容量',
    enrolled_students  INT         NOT NULL DEFAULT 0      COMMENT '已选人数',
    status             VARCHAR(20) NOT NULL DEFAULT 'active' COMMENT 'active/finished',
    created_at         DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at         DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_course_class_course (course_id),
    KEY idx_course_class_term (term)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='课程班级（开课）';

CREATE TABLE IF NOT EXISTS t_grade (
    id               BIGINT       NOT NULL AUTO_INCREMENT COMMENT '主键',
    student_id       BIGINT       NOT NULL                COMMENT '学生档案ID',
    course_class_id  BIGINT       NOT NULL                COMMENT '课程班级ID',
    score            DECIMAL(5,1) DEFAULT NULL            COMMENT '成绩',
    grade_point      DECIMAL(3,1) DEFAULT NULL            COMMENT '绩点',
    status           VARCHAR(20)  NOT NULL DEFAULT 'pending' COMMENT 'pending/passed/failed/retake',
    term             VARCHAR(20)  NOT NULL                COMMENT '学期',
    exam_date        DATE         DEFAULT NULL            COMMENT '考试日期',
    created_at       DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at       DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_grade (student_id, course_class_id),
    KEY idx_grade_student (student_id),
    KEY idx_grade_course (course_class_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='成绩';

CREATE TABLE IF NOT EXISTS t_alert (
    id                   BIGINT       NOT NULL AUTO_INCREMENT COMMENT '主键',
    student_id           BIGINT       NOT NULL                COMMENT '学生档案ID',
    level                VARCHAR(10)  NOT NULL                COMMENT 'yellow/orange/red',
    type                 VARCHAR(50)  NOT NULL                COMMENT '预警类型',
    title                VARCHAR(200) NOT NULL                COMMENT '预警标题',
    description          VARCHAR(500) DEFAULT NULL            COMMENT '预警描述',
    course               VARCHAR(100) DEFAULT NULL            COMMENT '关联课程',
    failed_courses       JSON         DEFAULT NULL            COMMENT '挂科课程数组',
    trigger_date         DATE         DEFAULT NULL            COMMENT '触发日期',
    status               VARCHAR(20)  NOT NULL DEFAULT 'pending' COMMENT 'pending/processing/resolved',
    suggestion           VARCHAR(500) DEFAULT NULL            COMMENT '建议',
    trigger_event        VARCHAR(200) DEFAULT NULL            COMMENT '触发事件',
    pushed_at            DATETIME     DEFAULT NULL            COMMENT '推送时间',
    handler_id           BIGINT       DEFAULT NULL            COMMENT '处理人用户ID',
    handle_note          VARCHAR(500) DEFAULT NULL            COMMENT '处理备注',
    student_confirmed_at DATETIME     DEFAULT NULL            COMMENT '学生确认时间',
    created_at           DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at           DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_alert_student (student_id),
    KEY idx_alert_status (status),
    KEY idx_alert_level (level)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='预警记录';

CREATE TABLE IF NOT EXISTS t_intervention_record (
    id                  BIGINT       NOT NULL AUTO_INCREMENT COMMENT '主键',
    alert_id            BIGINT       NOT NULL                COMMENT '预警ID',
    teacher_id          BIGINT       NOT NULL                COMMENT '教师档案ID',
    intervention_date   DATE         DEFAULT NULL            COMMENT '干预日期',
    methods             JSON         DEFAULT NULL            COMMENT '干预方式数组',
    content             VARCHAR(1000) DEFAULT NULL           COMMENT '干预内容',
    student_response    VARCHAR(50)  DEFAULT NULL            COMMENT '学生反馈',
    follow_up_plan      VARCHAR(500) DEFAULT NULL            COMMENT '后续计划',
    created_at          DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at          DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_intervention_alert (alert_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='教师干预回填';

CREATE TABLE IF NOT EXISTS t_gpa_history (
    id             BIGINT       NOT NULL AUTO_INCREMENT COMMENT '主键',
    student_id     BIGINT       NOT NULL                COMMENT '学生档案ID',
    term           VARCHAR(20)  NOT NULL                COMMENT '学期',
    gpa            DECIMAL(4,2) NOT NULL                COMMENT '学期GPA',
    avg_gpa        DECIMAL(4,2) DEFAULT NULL            COMMENT '专业平均GPA',
    `rank`         INT          DEFAULT NULL            COMMENT '专业排名',
    total_students INT          DEFAULT NULL            COMMENT '专业总人数',
    created_at     DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_gpa_history (student_id, term),
    KEY idx_gpa_term (term)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='学期GPA历史';

CREATE TABLE IF NOT EXISTS t_profile_score (
    id             BIGINT       NOT NULL AUTO_INCREMENT COMMENT '主键',
    student_id     BIGINT       NOT NULL                COMMENT '学生档案ID',
    term           VARCHAR(20)  NOT NULL                COMMENT '学期',
    dimension_key  VARCHAR(50)  NOT NULL                COMMENT '维度 key',
    label          VARCHAR(50)  NOT NULL                COMMENT '维度名称',
    score          INT          NOT NULL DEFAULT 0      COMMENT '分数',
    avg_score      INT          DEFAULT NULL            COMMENT '专业平均分',
    max_score      INT          DEFAULT NULL            COMMENT '最高分',
    description    VARCHAR(200) DEFAULT NULL            COMMENT '描述',
    details        JSON         DEFAULT NULL            COMMENT '明细数组',
    created_at     DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_profile_score (student_id, term, dimension_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='五维画像分数';

CREATE TABLE IF NOT EXISTS t_health_report (
    id          BIGINT       NOT NULL AUTO_INCREMENT COMMENT '主键',
    student_id  BIGINT       NOT NULL                COMMENT '学生档案ID',
    term        VARCHAR(20)  NOT NULL                COMMENT '学期',
    total_score INT          DEFAULT NULL            COMMENT '总分',
    items       JSON         DEFAULT NULL            COMMENT '项目明细数组',
    report_url  VARCHAR(255) DEFAULT NULL            COMMENT '报告地址',
    created_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_health_report (student_id, term)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='体测报告';

-- ============ 种子数据 ============

INSERT INTO t_department (id, name, code, dean_id) VALUES
(1, '数学与计算机科学学院', 'MATH_CS', (SELECT id FROM t_user WHERE account = 'L20210001'));

INSERT INTO t_major (id, name, code, department_id, degree_type) VALUES
(1, '计算机科学与技术', 'CS', 1, '工学学士'),
(2, '软件工程', 'SE', 1, '工学学士');

INSERT INTO t_teacher (id, teacher_id, user_id, department_id, title, is_class_advisor, class_id) VALUES
(1, 'T20180042', (SELECT id FROM t_user WHERE account = 'T20180042'), 1, '副教授', 1, NULL),
(2, 'C20180099', (SELECT id FROM t_user WHERE account = 'C20180099'), 1, '讲师', 0, NULL);

INSERT INTO t_class (id, name, grade, major_id, advisor_id, total_students) VALUES
(1, '计科2201', '2022', 1, 1, 60),
(2, '计科2202', '2022', 2, 1, 60);

UPDATE t_teacher SET class_id = 1 WHERE id = 1;

-- 密码均为 123456
INSERT INTO t_user (account, name, password, role, college, major, class_name, grade) VALUES
('2022002', '王芳', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'student', '数学与计算机科学学院', '计算机科学与技术', '计科2201', '2022'),
('2022003', '赵磊', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'student', '数学与计算机科学学院', '计算机科学与技术', '计科2201', '2022'),
('2022004', '陈晨', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'student', '数学与计算机科学学院', '计算机科学与技术', '计科2201', '2022'),
('2022005', '刘洋', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'student', '数学与计算机科学学院', '计算机科学与技术', '计科2201', '2022'),
('2022006', '孙悦', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'student', '数学与计算机科学学院', '计算机科学与技术', '计科2201', '2022'),
('2022007', '周杰', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'student', '数学与计算机科学学院', '计算机科学与技术', '计科2201', '2022'),
('2022008', '吴桐', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'student', '数学与计算机科学学院', '计算机科学与技术', '计科2201', '2022'),
('2022009', '郑爽', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'student', '数学与计算机科学学院', '计算机科学与技术', '计科2201', '2022'),
('2022010', '钱进', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'student', '数学与计算机科学学院', '计算机科学与技术', '计科2201', '2022'),
('2022011', '何静', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'student', '数学与计算机科学学院', '计算机科学与技术', '计科2201', '2022'),
('2022012', '罗浩', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'student', '数学与计算机科学学院', '计算机科学与技术', '计科2201', '2022');

INSERT INTO t_student (id, student_id, user_id, class_id, major_id, department_id, gpa, `rank`, total_students, alert_level, total_credits, required_credits)
SELECT 1, '2022001', id, 1, 1, 1, 3.62, 15, 120, 'none', 89.0, 176.0 FROM t_user WHERE account = '2022001'
UNION ALL SELECT 2, '2022002', id, 1, 1, 1, 2.80, 62, 120, 'orange', 82.0, 176.0 FROM t_user WHERE account = '2022002'
UNION ALL SELECT 3, '2022003', id, 1, 1, 1, 1.80, 108, 120, 'red', 68.0, 176.0 FROM t_user WHERE account = '2022003'
UNION ALL SELECT 4, '2022004', id, 1, 1, 1, 3.78, 10, 120, 'none', 95.0, 176.0 FROM t_user WHERE account = '2022004'
UNION ALL SELECT 5, '2022005', id, 1, 1, 1, 1.50, 118, 120, 'red', 62.0, 176.0 FROM t_user WHERE account = '2022005'
UNION ALL SELECT 6, '2022006', id, 1, 1, 1, 3.50, 22, 120, 'none', 93.0, 176.0 FROM t_user WHERE account = '2022006'
UNION ALL SELECT 7, '2022007', id, 1, 1, 1, 2.42, 72, 120, 'yellow', 78.0, 176.0 FROM t_user WHERE account = '2022007'
UNION ALL SELECT 8, '2022008', id, 1, 1, 1, 3.80, 7, 120, 'none', 98.0, 176.0 FROM t_user WHERE account = '2022008'
UNION ALL SELECT 9, '2022009', id, 1, 1, 1, 2.75, 60, 120, 'yellow', 84.0, 176.0 FROM t_user WHERE account = '2022009'
UNION ALL SELECT 10, '2022010', id, 1, 1, 1, 1.70, 110, 120, 'orange', 65.0, 176.0 FROM t_user WHERE account = '2022010'
UNION ALL SELECT 11, '2022011', id, 1, 1, 1, 3.92, 3, 120, 'none', 102.0, 176.0 FROM t_user WHERE account = '2022011'
UNION ALL SELECT 12, '2022012', id, 1, 1, 1, 2.40, 78, 120, 'orange', 76.0, 176.0 FROM t_user WHERE account = '2022012';

INSERT INTO t_course (id, code, name, credits, type, department_id) VALUES
(1, 'MA102',  '高等数学（下）', 4.0, 'required', 1),
(2, 'MA103',  '线性代数',      3.0, 'required', 1),
(3, 'CS2021', '数据结构',      4.0, 'required', 1),
(4, 'CS2022', '操作系统',      3.5, 'required', 1),
(5, 'EN104',  '大学英语（四）', 2.0, 'public',   1),
(6, 'CS2023', '计算机网络',    3.0, 'elective', 1);

INSERT INTO t_course_class (id, course_id, term, class_name, teacher_id, max_students, enrolled_students, status) VALUES
(1, 1, '2024-2025-1', '计科2201', 1, 60, 60, 'active'),
(2, 2, '2024-2025-1', '计科2201', 1, 60, 60, 'active'),
(3, 3, '2024-2025-1', '计科2201', 2, 60, 60, 'active'),
(4, 4, '2024-2025-1', '计科2201', 2, 60, 60, 'active'),
(5, 5, '2024-2025-1', '计科2201', 1, 60, 60, 'active'),
(6, 6, '2024-2025-1', '计科2201', 2, 60, 60, 'active');

-- 高等数学（下）
INSERT INTO t_grade (student_id, course_class_id, score, grade_point, status, term, exam_date) VALUES
(1, 1, 92, 4.0, 'passed', '2024-2025-1', '2025-01-15'),
(2, 1, 76, 3.0, 'passed', '2024-2025-1', '2025-01-15'),
(3, 1, 55, 0.0, 'failed', '2024-2025-1', '2025-01-15'),
(4, 1, 88, 3.7, 'passed', '2024-2025-1', '2025-01-15'),
(5, 1, 45, 0.0, 'failed', '2024-2025-1', '2025-01-15'),
(6, 1, 82, 3.3, 'passed', '2024-2025-1', '2025-01-15'),
(7, 1, 61, 2.0, 'passed', '2024-2025-1', '2025-01-15'),
(8, 1, 90, 4.0, 'passed', '2024-2025-1', '2025-01-15'),
(9, 1, 72, 2.7, 'passed', '2024-2025-1', '2025-01-15'),
(10, 1, 50, 0.0, 'failed', '2024-2025-1', '2025-01-15'),
(11, 1, 95, 4.0, 'passed', '2024-2025-1', '2025-01-15'),
(12, 1, 68, 2.3, 'passed', '2024-2025-1', '2025-01-15');

-- 线性代数
INSERT INTO t_grade (student_id, course_class_id, score, grade_point, status, term, exam_date) VALUES
(1, 2, 88, 3.7, 'passed', '2024-2025-1', '2025-01-16'),
(2, 2, 70, 2.7, 'passed', '2024-2025-1', '2025-01-16'),
(3, 2, 60, 2.0, 'passed', '2024-2025-1', '2025-01-16'),
(4, 2, 90, 4.0, 'passed', '2024-2025-1', '2025-01-16'),
(5, 2, 50, 0.0, 'failed', '2024-2025-1', '2025-01-16'),
(6, 2, 78, 3.0, 'passed', '2024-2025-1', '2025-01-16'),
(7, 2, 55, 0.0, 'failed', '2024-2025-1', '2025-01-16'),
(8, 2, 85, 3.3, 'passed', '2024-2025-1', '2025-01-16'),
(9, 2, 65, 2.3, 'passed', '2024-2025-1', '2025-01-16'),
(10, 2, 42, 0.0, 'failed', '2024-2025-1', '2025-01-16'),
(11, 2, 92, 4.0, 'passed', '2024-2025-1', '2025-01-16'),
(12, 2, 62, 2.0, 'passed', '2024-2025-1', '2025-01-16');

-- 数据结构
INSERT INTO t_grade (student_id, course_class_id, score, grade_point, status, term, exam_date) VALUES
(1, 3, 90, 4.0, 'passed', '2024-2025-1', '2025-01-17'),
(2, 3, 58, 0.0, 'failed', '2024-2025-1', '2025-01-17'),
(3, 3, 49, 0.0, 'failed', '2024-2025-1', '2025-01-17'),
(4, 3, 92, 4.0, 'passed', '2024-2025-1', '2025-01-17'),
(5, 3, 60, 2.0, 'passed', '2024-2025-1', '2025-01-17'),
(6, 3, 85, 3.3, 'passed', '2024-2025-1', '2025-01-17'),
(7, 3, 70, 2.7, 'passed', '2024-2025-1', '2025-01-17'),
(8, 3, 88, 3.7, 'passed', '2024-2025-1', '2025-01-17'),
(9, 3, 58, 0.0, 'failed', '2024-2025-1', '2025-01-17'),
(10, 3, 66, 2.3, 'passed', '2024-2025-1', '2025-01-17'),
(11, 3, 96, 4.0, 'passed', '2024-2025-1', '2025-01-17'),
(12, 3, 55, 0.0, 'failed', '2024-2025-1', '2025-01-17');

-- 操作系统
INSERT INTO t_grade (student_id, course_class_id, score, grade_point, status, term, exam_date) VALUES
(1, 4, 85, 3.3, 'passed', '2024-2025-1', '2025-01-18'),
(2, 4, 62, 2.0, 'passed', '2024-2025-1', '2025-01-18'),
(3, 4, 52, 0.0, 'failed', '2024-2025-1', '2025-01-18'),
(4, 4, 86, 3.3, 'passed', '2024-2025-1', '2025-01-18'),
(5, 4, 55, 0.0, 'failed', '2024-2025-1', '2025-01-18'),
(6, 4, 80, 3.3, 'passed', '2024-2025-1', '2025-01-18'),
(7, 4, 58, 0.0, 'failed', '2024-2025-1', '2025-01-18'),
(8, 4, 92, 4.0, 'passed', '2024-2025-1', '2025-01-18'),
(9, 4, 60, 2.0, 'passed', '2024-2025-1', '2025-01-18'),
(10, 4, 55, 0.0, 'failed', '2024-2025-1', '2025-01-18'),
(11, 4, 90, 4.0, 'passed', '2024-2025-1', '2025-01-18'),
(12, 4, 50, 0.0, 'failed', '2024-2025-1', '2025-01-18');

-- 大学英语（四）
INSERT INTO t_grade (student_id, course_class_id, score, grade_point, status, term, exam_date) VALUES
(1, 5, 95, 4.0, 'passed', '2024-2025-1', '2025-01-19'),
(2, 5, 80, 3.3, 'passed', '2024-2025-1', '2025-01-19'),
(3, 5, 65, 2.3, 'passed', '2024-2025-1', '2025-01-19'),
(4, 5, 90, 4.0, 'passed', '2024-2025-1', '2025-01-19'),
(5, 5, 70, 2.7, 'passed', '2024-2025-1', '2025-01-19'),
(6, 5, 88, 3.7, 'passed', '2024-2025-1', '2025-01-19'),
(7, 5, 75, 3.0, 'passed', '2024-2025-1', '2025-01-19'),
(8, 5, 94, 4.0, 'passed', '2024-2025-1', '2025-01-19'),
(9, 5, 75, 3.0, 'passed', '2024-2025-1', '2025-01-19'),
(10, 5, 72, 2.7, 'passed', '2024-2025-1', '2025-01-19'),
(11, 5, 98, 4.0, 'passed', '2024-2025-1', '2025-01-19'),
(12, 5, 78, 3.0, 'passed', '2024-2025-1', '2025-01-19');

-- 计算机网络
INSERT INTO t_grade (student_id, course_class_id, score, grade_point, status, term, exam_date) VALUES
(1, 6, 87, 3.7, 'passed', '2024-2025-1', '2025-01-20'),
(2, 6, 68, 2.3, 'passed', '2024-2025-1', '2025-01-20'),
(3, 6, 58, 0.0, 'failed', '2024-2025-1', '2025-01-20'),
(4, 6, 91, 4.0, 'passed', '2024-2025-1', '2025-01-20'),
(5, 6, 48, 0.0, 'failed', '2024-2025-1', '2025-01-20'),
(6, 6, 84, 3.3, 'passed', '2024-2025-1', '2025-01-20'),
(7, 6, 66, 2.3, 'passed', '2024-2025-1', '2025-01-20'),
(8, 6, 89, 3.7, 'passed', '2024-2025-1', '2025-01-20'),
(9, 6, 63, 2.0, 'passed', '2024-2025-1', '2025-01-20'),
(10, 6, 60, 2.0, 'passed', '2024-2025-1', '2025-01-20'),
(11, 6, 93, 4.0, 'passed', '2024-2025-1', '2025-01-20'),
(12, 6, 58, 0.0, 'failed', '2024-2025-1', '2025-01-20');

INSERT INTO t_alert (id, student_id, level, type, title, description, course, failed_courses, trigger_date, status, suggestion, trigger_event, pushed_at) VALUES
(1, 3,  'red',    '挂科预警', '累计挂科3门', '累计不及格课程达到3门，触发红色预警，班主任已收到通知', '数据结构', '["高等数学（下）","线性代数","数据结构"]', '2025-04-08', 'processing', '建议制定补考计划，安排一对一辅导', '累计不及格达3门', '2025-04-08 14:30'),
(2, 2,  'orange', '挂科预警', '数据结构不及格', '本学期数据结构成绩低于60分，存在挂科风险', '数据结构', '["数据结构"]', '2025-04-08', 'processing', '建议参加课程辅导并加强练习', '课程成绩低于60分', '2025-04-08 15:02'),
(3, 5,  'red',    '挂科预警', '累计挂科3门', '高等数学、线性代数、计算机网络均不及格', '计算机网络', '["高等数学（下）","线性代数","计算机网络"]', '2025-04-09', 'pending',   '建议家长会面并制定专项帮扶方案', '累计不及格达3门', '2025-04-09 09:10'),
(4, 10, 'orange', '成绩预警', 'GPA 低于 2.5', '当前学期GPA为1.70，明显低于专业平均水平', NULL, '[]', '2025-04-09', 'pending', '建议一对一学业咨询并调整学习计划', 'GPA低于2.5', '2025-04-09 10:00'),
(5, 7,  'yellow', '出勤预警', '缺勤次数较多', '本学期缺勤累计8次，存在学业下滑风险', '操作系统', '[]', '2025-04-10', 'pending', '建议与学生谈话并关注课堂出勤', '学期缺勤达8次', '2025-04-10 11:20'),
(6, 12, 'orange', '挂科预警', '操作系统不及格', '操作系统成绩低于60分，且多门课程成绩偏低', '操作系统', '["操作系统"]', '2025-04-10', 'pending', '建议参加补考辅导小组', '课程成绩低于60分', '2025-04-10 14:00'),
(7, 9,  'yellow', '成绩预警', '本学期成绩下滑', 'GPA由上学期3.05降至2.75，成绩呈下滑趋势', NULL, '[]', '2025-04-11', 'resolved', '已完成谈心谈话并约定每周复盘', 'GPA连续下滑', '2025-04-11 16:30'),
(8, 3,  'orange', '考勤预警', '课堂参与度低', '数据结构课堂出勤率低于70%，影响学习效果', '数据结构', '[]', '2025-04-12', 'resolved', '已安排学习伙伴帮扶', '出勤率低于70%', '2025-04-12 09:00');

INSERT INTO t_intervention_record (id, alert_id, teacher_id, intervention_date, methods, content, student_response, follow_up_plan) VALUES
(1, 1, 1, '2025-04-12', '["phone","meeting"]', '与家长沟通，并约定每周两次辅导', 'positive', '两周后复查作业完成情况'),
(2, 2, 2, '2025-04-13', '["meeting"]', '课后一对一讲解数据结构重点章节', 'neutral', '下次测验后评估进步情况');

-- 五个学期 GPA 历史
INSERT INTO t_gpa_history (student_id, term, gpa, avg_gpa, `rank`, total_students) VALUES
(1, '2022-2023-1', 3.45, 3.10, 25, 120),
(2, '2022-2023-1', 3.10, 3.05, 35, 120),
(3, '2022-2023-1', 2.60, 3.00, 60, 120),
(4, '2022-2023-1', 3.55, 3.12, 20, 120),
(5, '2022-2023-1', 2.20, 2.95, 85, 120),
(6, '2022-2023-1', 3.30, 3.08, 30, 120),
(7, '2022-2023-1', 2.80, 3.00, 50, 120),
(8, '2022-2023-1', 3.60, 3.15, 15, 120),
(9, '2022-2023-1', 3.05, 3.05, 38, 120),
(10, '2022-2023-1', 2.40, 2.95, 72, 120),
(11, '2022-2023-1', 3.75, 3.20, 8, 120),
(12, '2022-2023-1', 2.95, 3.05, 45, 120);

INSERT INTO t_gpa_history (student_id, term, gpa, avg_gpa, `rank`, total_students) VALUES
(1, '2022-2023-2', 3.52, 3.12, 22, 120),
(2, '2022-2023-2', 3.02, 3.02, 40, 120),
(3, '2022-2023-2', 2.40, 2.95, 70, 120),
(4, '2022-2023-2', 3.60, 3.14, 18, 120),
(5, '2022-2023-2', 2.00, 2.90, 95, 120),
(6, '2022-2023-2', 3.35, 3.10, 28, 120),
(7, '2022-2023-2', 2.70, 2.98, 55, 120),
(8, '2022-2023-2', 3.65, 3.17, 13, 120),
(9, '2022-2023-2', 2.98, 3.02, 42, 120),
(10, '2022-2023-2', 2.20, 2.90, 82, 120),
(11, '2022-2023-2', 3.80, 3.22, 6, 120),
(12, '2022-2023-2', 2.85, 3.00, 52, 120);

INSERT INTO t_gpa_history (student_id, term, gpa, avg_gpa, `rank`, total_students) VALUES
(1, '2023-2024-1', 3.62, 3.15, 18, 120),
(2, '2023-2024-1', 2.95, 3.00, 48, 120),
(3, '2023-2024-1', 2.20, 2.90, 80, 120),
(4, '2023-2024-1', 3.68, 3.16, 15, 120),
(5, '2023-2024-1', 1.85, 2.85, 102, 120),
(6, '2023-2024-1', 3.40, 3.12, 26, 120),
(7, '2023-2024-1', 2.60, 2.95, 60, 120),
(8, '2023-2024-1', 3.70, 3.19, 11, 120),
(9, '2023-2024-1', 2.90, 2.98, 48, 120),
(10, '2023-2024-1', 2.00, 2.85, 92, 120),
(11, '2023-2024-1', 3.85, 3.24, 5, 120),
(12, '2023-2024-1', 2.70, 2.95, 60, 120);

INSERT INTO t_gpa_history (student_id, term, gpa, avg_gpa, `rank`, total_students) VALUES
(1, '2023-2024-2', 3.70, 3.18, 15, 120),
(2, '2023-2024-2', 2.88, 2.98, 55, 120),
(3, '2023-2024-2', 1.95, 2.85, 95, 120),
(4, '2023-2024-2', 3.72, 3.18, 12, 120),
(5, '2023-2024-2', 1.60, 2.80, 110, 120),
(6, '2023-2024-2', 3.45, 3.15, 24, 120),
(7, '2023-2024-2', 2.50, 2.92, 66, 120),
(8, '2023-2024-2', 3.75, 3.21, 9, 120),
(9, '2023-2024-2', 2.82, 2.95, 54, 120),
(10, '2023-2024-2', 1.85, 2.80, 100, 120),
(11, '2023-2024-2', 3.90, 3.26, 4, 120),
(12, '2023-2024-2', 2.55, 2.90, 68, 120);

INSERT INTO t_gpa_history (student_id, term, gpa, avg_gpa, `rank`, total_students) VALUES
(1, '2024-2025-1', 3.75, 3.20, 12, 120),
(2, '2024-2025-1', 2.80, 2.95, 62, 120),
(3, '2024-2025-1', 1.80, 2.80, 108, 120),
(4, '2024-2025-1', 3.78, 3.20, 10, 120),
(5, '2024-2025-1', 1.50, 2.75, 118, 120),
(6, '2024-2025-1', 3.50, 3.17, 22, 120),
(7, '2024-2025-1', 2.42, 2.90, 72, 120),
(8, '2024-2025-1', 3.80, 3.23, 7, 120),
(9, '2024-2025-1', 2.75, 2.92, 60, 120),
(10, '2024-2025-1', 1.70, 2.76, 110, 120),
(11, '2024-2025-1', 3.92, 3.28, 3, 120),
(12, '2024-2025-1', 2.40, 2.85, 78, 120);

-- 五维画像（当前学期）
INSERT INTO t_profile_score (student_id, term, dimension_key, label, score, avg_score, max_score, description, details) VALUES
(1, '2024-2025-1', 'academic', '学业成绩', 88, 75, 100, 'GPA 位于专业前 13%', '["GPA: 3.75 / 4.0","核心课均分 86.5 分"]'),
(2, '2024-2025-1', 'academic', '学业成绩', 70, 75, 100, '成绩处于中游，个别课程偏弱', '["GPA: 2.80 / 4.0","数据结构需要加强"]'),
(3, '2024-2025-1', 'academic', '学业成绩', 45, 75, 100, '挂科课程较多，需重点帮扶', '["GPA: 1.80 / 4.0","累计挂科 3 门"]'),
(4, '2024-2025-1', 'academic', '学业成绩', 90, 75, 100, '成绩优秀，保持稳定', '["GPA: 3.78 / 4.0","核心课均分 89 分"]'),
(5, '2024-2025-1', 'academic', '学业成绩', 40, 75, 100, '成绩垫底，存在退学风险', '["GPA: 1.50 / 4.0","累计挂科 3 门"]'),
(6, '2024-2025-1', 'academic', '学业成绩', 82, 75, 100, '成绩良好，具备提升空间', '["GPA: 3.50 / 4.0","专业课表现稳定"]'),
(7, '2024-2025-1', 'academic', '学业成绩', 60, 75, 100, '成绩偏低，出勤问题突出', '["GPA: 2.42 / 4.0","缺勤次数较多"]'),
(8, '2024-2025-1', 'academic', '学业成绩', 92, 75, 100, '成绩优秀，专业排名靠前', '["GPA: 3.80 / 4.0","核心课均分 91 分"]'),
(9, '2024-2025-1', 'academic', '学业成绩', 66, 75, 100, '成绩下滑，需要关注', '["GPA: 2.75 / 4.0","近两学期持续下降"]'),
(10, '2024-2025-1', 'academic', '学业成绩', 50, 75, 100, '成绩偏低，需制定提升计划', '["GPA: 1.70 / 4.0","多门课程低于 60 分"]'),
(11, '2024-2025-1', 'academic', '学业成绩', 95, 75, 100, '成绩顶尖，具备保研潜力', '["GPA: 3.92 / 4.0","专业排名第 3"]'),
(12, '2024-2025-1', 'academic', '学业成绩', 58, 75, 100, '成绩偏低，专业课偏弱', '["GPA: 2.40 / 4.0","操作系统不及格"]');

INSERT INTO t_profile_score (student_id, term, dimension_key, label, score, avg_score, max_score, description, details) VALUES
(1, '2024-2025-1', 'ability', '实践能力', 78, 70, 100, '竞赛与项目经历较丰富', '["参与省级竞赛 2 次","课程项目完成度 90%"]'),
(2, '2024-2025-1', 'ability', '实践能力', 68, 70, 100, '实践能力中等', '["参与课程项目 1 次"]'),
(3, '2024-2025-1', 'ability', '实践能力', 55, 70, 100, '实践参与较少', '["建议参与实验室开放项目"]'),
(4, '2024-2025-1', 'ability', '实践能力', 80, 70, 100, '实践能力良好', '["主持课程项目 1 次"]'),
(5, '2024-2025-1', 'ability', '实践能力', 50, 70, 100, '实践参与不足', '["建议从基础训练开始"]'),
(6, '2024-2025-1', 'ability', '实践能力', 75, 70, 100, '实践能力良好', '["参与校创项目 1 次"]'),
(7, '2024-2025-1', 'ability', '实践能力', 62, 70, 100, '实践能力一般', '["建议加强编程练习"]'),
(8, '2024-2025-1', 'ability', '实践能力', 82, 70, 100, '实践能力优秀', '["参与国家级竞赛 1 次"]'),
(9, '2024-2025-1', 'ability', '实践能力', 64, 70, 100, '实践能力一般', '["建议参与学科竞赛"]'),
(10, '2024-2025-1', 'ability', '实践能力', 52, 70, 100, '实践参与较少', '["建议参加实训项目"]'),
(11, '2024-2025-1', 'ability', '实践能力', 88, 70, 100, '实践能力突出', '["论文/专利 1 项","竞赛获奖 2 次"]'),
(12, '2024-2025-1', 'ability', '实践能力', 60, 70, 100, '实践能力一般', '["建议加强专业实训"]');

INSERT INTO t_profile_score (student_id, term, dimension_key, label, score, avg_score, max_score, description, details) VALUES
(1, '2024-2025-1', 'psychology', '心理素质', 82, 75, 100, '心理状态稳定', '["测评结果良好"]'),
(2, '2024-2025-1', 'psychology', '心理素质', 74, 75, 100, '心理状态良好', '["测评结果正常"]'),
(3, '2024-2025-1', 'psychology', '心理素质', 60, 75, 100, '学业压力较大', '["建议预约心理辅导"]'),
(4, '2024-2025-1', 'psychology', '心理素质', 85, 75, 100, '心理状态优秀', '["测评结果优秀"]'),
(5, '2024-2025-1', 'psychology', '心理素质', 55, 75, 100, '情绪波动明显', '["建议重点关注"]'),
(6, '2024-2025-1', 'psychology', '心理素质', 80, 75, 100, '心理状态良好', '["测评结果正常"]'),
(7, '2024-2025-1', 'psychology', '心理素质', 70, 75, 100, '心理状态一般', '["建议增加沟通交流"]'),
(8, '2024-2025-1', 'psychology', '心理素质', 86, 75, 100, '心理状态优秀', '["测评结果优秀"]'),
(9, '2024-2025-1', 'psychology', '心理素质', 72, 75, 100, '心理状态一般', '["测评结果正常"]'),
(10, '2024-2025-1', 'psychology', '心理素质', 60, 75, 100, '学业焦虑偏高', '["建议预约心理辅导"]'),
(11, '2024-2025-1', 'psychology', '心理素质', 90, 75, 100, '心理状态优秀', '["测评结果优秀"]'),
(12, '2024-2025-1', 'psychology', '心理素质', 68, 75, 100, '心理状态一般', '["建议加强关注"]');

INSERT INTO t_profile_score (student_id, term, dimension_key, label, score, avg_score, max_score, description, details) VALUES
(1, '2024-2025-1', 'health', '体质健康', 85, 80, 100, '体测成绩良好', '["体测总分 85 分"]'),
(2, '2024-2025-1', 'health', '体质健康', 80, 80, 100, '体测成绩良好', '["体测总分 80 分"]'),
(3, '2024-2025-1', 'health', '体质健康', 70, 80, 100, '体测成绩一般', '["建议加强锻炼"]'),
(4, '2024-2025-1', 'health', '体质健康', 88, 80, 100, '体测成绩优秀', '["体测总分 88 分"]'),
(5, '2024-2025-1', 'health', '体质健康', 68, 80, 100, '体测成绩一般', '["建议规律运动"]'),
(6, '2024-2025-1', 'health', '体质健康', 84, 80, 100, '体测成绩良好', '["体测总分 84 分"]'),
(7, '2024-2025-1', 'health', '体质健康', 78, 80, 100, '体测成绩良好', '["体测总分 78 分"]'),
(8, '2024-2025-1', 'health', '体质健康', 90, 80, 100, '体测成绩优秀', '["体测总分 90 分"]'),
(9, '2024-2025-1', 'health', '体质健康', 80, 80, 100, '体测成绩良好', '["体测总分 80 分"]'),
(10, '2024-2025-1', 'health', '体质健康', 72, 80, 100, '体测成绩一般', '["建议增加运动量"]'),
(11, '2024-2025-1', 'health', '体质健康', 93, 80, 100, '体测成绩优秀', '["体测总分 93 分"]'),
(12, '2024-2025-1', 'health', '体质健康', 76, 80, 100, '体测成绩一般', '["建议规律锻炼"]');

INSERT INTO t_profile_score (student_id, term, dimension_key, label, score, avg_score, max_score, description, details) VALUES
(1, '2024-2025-1', 'thought', '思想品德', 90, 85, 100, '思想表现优秀', '["志愿服务 20 小时"]'),
(2, '2024-2025-1', 'thought', '思想品德', 85, 85, 100, '思想表现良好', '["班级活动参与积极"]'),
(3, '2024-2025-1', 'thought', '思想品德', 80, 85, 100, '思想表现良好', '["日常表现正常"]'),
(4, '2024-2025-1', 'thought', '思想品德', 92, 85, 100, '思想表现优秀', '["担任班干部"]'),
(5, '2024-2025-1', 'thought', '思想品德', 78, 85, 100, '思想表现一般', '["建议参与集体活动"]'),
(6, '2024-2025-1', 'thought', '思想品德', 88, 85, 100, '思想表现良好', '["志愿服务 12 小时"]'),
(7, '2024-2025-1', 'thought', '思想品德', 84, 85, 100, '思想表现良好', '["日常表现正常"]'),
(8, '2024-2025-1', 'thought', '思想品德', 94, 85, 100, '思想表现优秀', '["志愿服务 30 小时"]'),
(9, '2024-2025-1', 'thought', '思想品德', 86, 85, 100, '思想表现良好', '["班级活动参与积极"]'),
(10, '2024-2025-1', 'thought', '思想品德', 80, 85, 100, '思想表现良好', '["日常表现正常"]'),
(11, '2024-2025-1', 'thought', '思想品德', 96, 85, 100, '思想表现优秀', '["校级表彰 1 次"]'),
(12, '2024-2025-1', 'thought', '思想品德', 82, 85, 100, '思想表现一般', '["建议参与班级活动"]');

INSERT INTO t_health_report (student_id, term, total_score, items, report_url) VALUES
(1,  '2024-2025-1', 85, '[{"name":"体重指数","score":82,"level":"良好"},{"name":"肺活量","score":88,"level":"优秀"},{"name":"耐力跑","score":85,"level":"良好"}]', NULL),
(2,  '2024-2025-1', 80, '[{"name":"体重指数","score":78,"level":"良好"},{"name":"肺活量","score":82,"level":"良好"},{"name":"耐力跑","score":80,"level":"良好"}]', NULL),
(3,  '2024-2025-1', 70, '[{"name":"体重指数","score":68,"level":"一般"},{"name":"肺活量","score":72,"level":"良好"},{"name":"耐力跑","score":70,"level":"一般"}]', NULL),
(4,  '2024-2025-1', 88, '[{"name":"体重指数","score":86,"level":"优秀"},{"name":"肺活量","score":90,"level":"优秀"},{"name":"耐力跑","score":88,"level":"优秀"}]', NULL),
(5,  '2024-2025-1', 68, '[{"name":"体重指数","score":65,"level":"一般"},{"name":"肺活量","score":70,"level":"一般"},{"name":"耐力跑","score":68,"level":"一般"}]', NULL),
(6,  '2024-2025-1', 84, '[{"name":"体重指数","score":82,"level":"良好"},{"name":"肺活量","score":85,"level":"良好"},{"name":"耐力跑","score":84,"level":"良好"}]', NULL),
(7,  '2024-2025-1', 78, '[{"name":"体重指数","score":75,"level":"良好"},{"name":"肺活量","score":80,"level":"良好"},{"name":"耐力跑","score":78,"level":"良好"}]', NULL),
(8,  '2024-2025-1', 90, '[{"name":"体重指数","score":88,"level":"优秀"},{"name":"肺活量","score":92,"level":"优秀"},{"name":"耐力跑","score":90,"level":"优秀"}]', NULL),
(9,  '2024-2025-1', 80, '[{"name":"体重指数","score":78,"level":"良好"},{"name":"肺活量","score":82,"level":"良好"},{"name":"耐力跑","score":80,"level":"良好"}]', NULL),
(10, '2024-2025-1', 72, '[{"name":"体重指数","score":70,"level":"一般"},{"name":"肺活量","score":74,"level":"良好"},{"name":"耐力跑","score":72,"level":"一般"}]', NULL),
(11, '2024-2025-1', 93, '[{"name":"体重指数","score":92,"level":"优秀"},{"name":"肺活量","score":95,"level":"优秀"},{"name":"耐力跑","score":93,"level":"优秀"}]', NULL),
(12, '2024-2025-1', 76, '[{"name":"体重指数","score":74,"level":"良好"},{"name":"肺活量","score":78,"level":"良好"},{"name":"耐力跑","score":76,"level":"良好"}]', NULL);
