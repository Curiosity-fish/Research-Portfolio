-- V4: M3 系主任/院长/发展引导模块
-- 新增职位缓存、院校缓存、培养方案表，并补充多年级/多专业演示数据

CREATE TABLE IF NOT EXISTS t_job_cache (
    id             BIGINT       NOT NULL AUTO_INCREMENT COMMENT '主键',
    platform       VARCHAR(30)  DEFAULT NULL            COMMENT '平台 key: boss/51job/zhilian/liepin',
    platform_label VARCHAR(50)  DEFAULT NULL            COMMENT '平台名称',
    company_name   VARCHAR(100) NOT NULL                COMMENT '公司名称',
    company_size   VARCHAR(50)  DEFAULT NULL            COMMENT '公司规模',
    company_stage  VARCHAR(50)  DEFAULT NULL            COMMENT '公司阶段',
    job_title      VARCHAR(100) NOT NULL                COMMENT '职位名称',
    salary_range   VARCHAR(50)  DEFAULT NULL            COMMENT '薪资范围',
    city           VARCHAR(50)  DEFAULT NULL            COMMENT '城市',
    district       VARCHAR(100) DEFAULT NULL            COMMENT '区/园区',
    education      VARCHAR(50)  DEFAULT NULL            COMMENT '学历要求',
    experience     VARCHAR(50)  DEFAULT NULL            COMMENT '经验要求',
    tags           JSON         DEFAULT NULL            COMMENT '技能标签数组',
    highlights     JSON         DEFAULT NULL            COMMENT '亮点数组',
    match_score    INT          NOT NULL DEFAULT 0      COMMENT '画像匹配度',
    match_reasons  JSON         DEFAULT NULL            COMMENT '匹配原因数组',
    publish_date   VARCHAR(20)  DEFAULT NULL            COMMENT '发布日期',
    source_url     VARCHAR(255) DEFAULT NULL            COMMENT '原始链接',
    category       VARCHAR(30)  DEFAULT NULL            COMMENT '岗位分类: backend/ai/pm/test/data/security',
    published_at   DATETIME     DEFAULT NULL            COMMENT '平台发布时间',
    crawled_at     DATETIME     DEFAULT NULL            COMMENT '抓取时间',
    created_at     DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_job_category (category),
    KEY idx_job_match (match_score)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='招聘职位缓存';

CREATE TABLE IF NOT EXISTS t_school_cache (
    id                   BIGINT       NOT NULL AUTO_INCREMENT COMMENT '主键',
    type                 VARCHAR(20)  NOT NULL                COMMENT 'grad/overseas',
    name                 VARCHAR(150) NOT NULL                COMMENT '院校英文名/中文名',
    name_zh              VARCHAR(150) DEFAULT NULL            COMMENT '院校中文名',
    country              VARCHAR(50)  DEFAULT NULL            COMMENT '国家',
    country_emoji        VARCHAR(20)  DEFAULT NULL            COMMENT '国家 emoji',
    location             VARCHAR(100) DEFAULT NULL            COMMENT '所在地',
    `rank`               VARCHAR(100) DEFAULT NULL            COMMENT '排名',
    tier                 VARCHAR(20)  DEFAULT NULL            COMMENT 'top/good/match',
    programs             JSON         DEFAULT NULL            COMMENT '招生专业数组',
    admission_gpa        VARCHAR(50)  DEFAULT NULL            COMMENT '录取 GPA 要求',
    exam_requirements    JSON         DEFAULT NULL            COMMENT '考试科目数组（考研）',
    language_requirement VARCHAR(100) DEFAULT NULL            COMMENT '语言要求（留学）',
    highlights           JSON         DEFAULT NULL            COMMENT '亮点数组',
    match_score          INT          NOT NULL DEFAULT 0      COMMENT '画像匹配度',
    match_reasons        JSON         DEFAULT NULL            COMMENT '匹配原因数组',
    official_url         VARCHAR(255) DEFAULT NULL            COMMENT '官网地址',
    deadline             VARCHAR(50)  DEFAULT NULL            COMMENT '申请截止时间',
    created_at           DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at           DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_school_type_tier (type, tier),
    KEY idx_school_match (match_score)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='院校库缓存';

CREATE TABLE IF NOT EXISTS t_major_plan (
    id         BIGINT      NOT NULL AUTO_INCREMENT COMMENT '主键',
    major_id   BIGINT      NOT NULL                COMMENT '专业ID',
    grade      VARCHAR(10) NOT NULL                COMMENT '年级，如 2022',
    dimensions JSON        NOT NULL                COMMENT '培养维度数组',
    standard   JSON        NOT NULL                COMMENT '标准学分数组',
    actual     JSON        NOT NULL                COMMENT '实际学分数组',
    tips       JSON        DEFAULT NULL            COMMENT '建议数组',
    created_at DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_major_plan (major_id, grade)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='专业培养方案';

-- ============ 补充专业与班级 ============

INSERT INTO t_major (id, name, code, department_id, degree_type)
SELECT 3, '数据科学', 'DS', 1, '理学学士'
WHERE NOT EXISTS (SELECT 1 FROM t_major WHERE name = '数据科学');

INSERT INTO t_class (id, name, grade, major_id, advisor_id, total_students)
SELECT 3, '计科2101', '2021', 1, 1, 40
WHERE NOT EXISTS (SELECT 1 FROM t_class WHERE name = '计科2101');

INSERT INTO t_class (id, name, grade, major_id, advisor_id, total_students)
SELECT 4, '计科2301', '2023', 1, 1, 50
WHERE NOT EXISTS (SELECT 1 FROM t_class WHERE name = '计科2301');

INSERT INTO t_class (id, name, grade, major_id, advisor_id, total_students)
SELECT 5, '计科2401', '2024', 1, 1, 50
WHERE NOT EXISTS (SELECT 1 FROM t_class WHERE name = '计科2401');

INSERT INTO t_class (id, name, grade, major_id, advisor_id, total_students)
SELECT 6, '软工2301', '2023', 2, 1, 45
WHERE NOT EXISTS (SELECT 1 FROM t_class WHERE name = '软工2301');

INSERT INTO t_class (id, name, grade, major_id, advisor_id, total_students)
SELECT 7, '数科2401', '2024', 3, 1, 40
WHERE NOT EXISTS (SELECT 1 FROM t_class WHERE name = '数科2401');

-- ============ 补充学生账号（密码均 123456） ============

INSERT IGNORE INTO t_user (account, name, password, role, college, major, class_name, grade) VALUES
('2021001', '赵敏', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'student', '数学与计算机科学学院', '计算机科学与技术', '计科2101', '2021'),
('2021002', '钱程', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'student', '数学与计算机科学学院', '计算机科学与技术', '计科2101', '2021'),
('2023001', '孙婷', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'student', '数学与计算机科学学院', '计算机科学与技术', '计科2301', '2023'),
('2023002', '李浩', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'student', '数学与计算机科学学院', '计算机科学与技术', '计科2301', '2023'),
('2024001', '周晓', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'student', '数学与计算机科学学院', '计算机科学与技术', '计科2401', '2024'),
('2024002', '吴凡', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'student', '数学与计算机科学学院', '计算机科学与技术', '计科2401', '2024'),
('2023003', '郑楠', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'student', '数学与计算机科学学院', '软件工程', '软工2301', '2023'),
('2023004', '王磊', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'student', '数学与计算机科学学院', '软件工程', '软工2301', '2023'),
('2023005', '冯雪', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'student', '数学与计算机科学学院', '软件工程', '软工2301', '2023'),
('2024003', '何宇', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'student', '数学与计算机科学学院', '数据科学', '数科2401', '2024'),
('2024004', '高远', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'student', '数学与计算机科学学院', '数据科学', '数科2401', '2024');

INSERT INTO t_student (student_id, user_id, class_id, major_id, department_id, gpa, `rank`, total_students, alert_level, total_credits, required_credits)
SELECT v.student_id, u.id, cl.id, m.id, 1, v.gpa, v.rank, v.total, v.alert, v.credits, 176.0
FROM (
    SELECT '2021001' student_id, '计科2101' class_name, '计算机科学与技术' major_name, 3.42 gpa, 18 `rank`, 40 total, 'none' alert, 118.0 credits
    UNION ALL SELECT '2021002', '计科2101', '计算机科学与技术', 2.05, 75, 40, 'red', 88.0
    UNION ALL SELECT '2023001', '计科2301', '计算机科学与技术', 3.55, 9, 50, 'none', 72.0
    UNION ALL SELECT '2023002', '计科2301', '计算机科学与技术', 2.22, 62, 50, 'orange', 58.0
    UNION ALL SELECT '2024001', '计科2401', '计算机科学与技术', 3.28, 22, 50, 'none', 36.0
    UNION ALL SELECT '2024002', '计科2401', '计算机科学与技术', 2.48, 48, 50, 'yellow', 30.0
    UNION ALL SELECT '2023003', '软工2301', '软件工程', 3.30, 16, 45, 'none', 70.0
    UNION ALL SELECT '2023004', '软工2301', '软件工程', 1.95, 42, 45, 'red', 52.0
    UNION ALL SELECT '2023005', '软工2301', '软件工程', 2.62, 28, 45, 'yellow', 66.0
    UNION ALL SELECT '2024003', '数科2401', '数据科学', 3.66, 6, 40, 'none', 38.0
    UNION ALL SELECT '2024004', '数科2401', '数据科学', 2.15, 34, 40, 'orange', 28.0
) v
JOIN t_user u ON u.account = v.student_id
JOIN t_class cl ON cl.name = v.class_name
JOIN t_major m ON m.name = v.major_name AND m.department_id = 1
WHERE NOT EXISTS (SELECT 1 FROM t_student s WHERE s.student_id = v.student_id);

-- ============ 补充课程班级与成绩 ============

INSERT INTO t_course_class (course_id, term, class_name, teacher_id, max_students, enrolled_students, status)
SELECT c.id, '2024-2025-1', cl.name, 2, cl.total_students, cl.total_students, 'active'
FROM t_course c
CROSS JOIN t_class cl
WHERE cl.id IN (3, 4, 5, 6, 7)
  AND NOT EXISTS (
      SELECT 1 FROM t_course_class cc
      WHERE cc.course_id = c.id AND cc.class_name = cl.name AND cc.term = '2024-2025-1'
  );

INSERT IGNORE INTO t_grade (student_id, course_class_id, score, status, term, exam_date)
SELECT s.id, cc.id,
       CASE
           WHEN MOD(s.id + cc.course_id, 5) = 0 THEN 45 + MOD(s.id + cc.id, 16)
           WHEN MOD(s.id + cc.course_id, 7) = 0 THEN 55 + MOD(s.id * 3 + cc.course_id * 5, 12)
           ELSE 62 + MOD(s.id * 7 + cc.course_id * 11, 36)
       END,
       'passed',
       '2024-2025-1',
       '2025-01-15'
FROM t_student s
JOIN t_class cl ON cl.id = s.class_id
JOIN t_course_class cc ON cc.class_name = cl.name AND cc.term = '2024-2025-1'
JOIN t_user u ON u.id = s.user_id
WHERE u.account IN ('2021001','2021002','2023001','2023002','2024001','2024002','2023003','2023004','2023005','2024003','2024004');

UPDATE t_grade g
JOIN t_course_class cc ON cc.id = g.course_class_id
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
WHERE u.account IN ('2021001','2021002','2023001','2023002','2024001','2024002','2023003','2023004','2023005','2024003','2024004');

-- ============ 补充 GPA 历史（按年级过滤学期） ============

INSERT IGNORE INTO t_gpa_history (student_id, term, gpa, avg_gpa, `rank`, total_students)
SELECT s.id, t.term,
       ROUND(LEAST(4.0, GREATEST(0.8, s.gpa + (t.idx - 7) * 0.05 + MOD(s.id, 7) * 0.01)), 2),
       ROUND(3.05 + (t.idx - 7) * 0.04, 2),
       GREATEST(1, s.`rank` + (7 - t.idx) * 3),
       s.total_students
FROM t_student s
JOIN t_user u ON u.id = s.user_id
JOIN (
    SELECT '2021-2022-1' term, 1 idx UNION ALL SELECT '2021-2022-2', 2
    UNION ALL SELECT '2022-2023-1', 3 UNION ALL SELECT '2022-2023-2', 4
    UNION ALL SELECT '2023-2024-1', 5 UNION ALL SELECT '2023-2024-2', 6
    UNION ALL SELECT '2024-2025-1', 7
) t
WHERE u.account IN ('2021001','2021002','2023001','2023002','2024001','2024002','2023003','2023004','2023005','2024003','2024004')
  AND CAST(LEFT(t.term, 4) AS SIGNED) - CAST(u.grade AS SIGNED) BETWEEN 0 AND 3;

-- ============ 补充画像分数（当前学期） ============

INSERT IGNORE INTO t_profile_score (student_id, term, dimension_key, label, score, avg_score, max_score, description, details)
SELECT s.id, '2024-2025-1', 'academic', '学业成绩',
       CASE
           WHEN s.gpa >= 3.5 THEN 88
           WHEN s.gpa >= 3.0 THEN 76
           WHEN s.gpa >= 2.5 THEN 64
           WHEN s.gpa >= 2.0 THEN 54
           ELSE 42
       END,
       72, 100,
       CONCAT('GPA ', FORMAT(s.gpa, 2), '，位于专业对应梯队'),
       JSON_ARRAY(CONCAT('GPA: ', FORMAT(s.gpa, 2), ' / 4.0'))
FROM t_student s
JOIN t_user u ON u.id = s.user_id
WHERE u.account IN ('2021001','2021002','2023001','2023002','2024001','2024002','2023003','2023004','2023005','2024003','2024004');

INSERT IGNORE INTO t_profile_score (student_id, term, dimension_key, label, score, avg_score, max_score, description, details)
SELECT s.id, '2024-2025-1', 'ability', '实践能力',
       CASE
           WHEN s.alert_level = 'none' THEN 74
           WHEN s.alert_level = 'yellow' THEN 62
           ELSE 52
       END,
       68, 100,
       '课程项目与竞赛参与情况正常',
       JSON_ARRAY('参与课程项目 1 次', '建议参与学科竞赛')
FROM t_student s
JOIN t_user u ON u.id = s.user_id
WHERE u.account IN ('2021001','2021002','2023001','2023002','2024001','2024002','2023003','2023004','2023005','2024003','2024004');

-- ============ 补充预警记录 ============

INSERT INTO t_alert (student_id, level, type, title, description, course, failed_courses, trigger_date, status, suggestion, trigger_event, pushed_at)
SELECT s.id, v.level, v.type, v.title, v.description, v.course, v.failed, v.trigger_date, v.status, v.suggestion, v.trigger_event, v.pushed_at
FROM (
    SELECT '2021002' account, 'red' level, '挂科预警' type, '累计挂科超过3门' title, '多门核心课程不及格，GPA 2.05，面临学业危机，需院级介入' description, '数据结构' course, '["高等数学（上）","高等数学（下）","线性代数","数据结构"]' failed, '2025-04-06' trigger_date, 'pending' status, '建议立即约谈，制定学业恢复计划' suggestion, '累计不及格超3门' trigger_event, '2025-04-06 15:20' pushed_at
    UNION ALL SELECT '2023002', 'orange', '成绩预警', 'GPA 低于 2.5', '当前 GPA 2.22，低于专业平均水平，核心课程存在挂科风险', '操作系统', '["操作系统"]', '2025-04-07', 'processing', '建议参加课后辅导并定期复盘', 'GPA低于2.5', '2025-04-07 10:30'
    UNION ALL SELECT '2024002', 'yellow', '出勤预警', '课堂参与度不足', '新生适应期内课堂出勤率低于 80%，需持续关注', NULL, '[]', '2025-04-08', 'pending', '建议班主任谈心并关注学习习惯', '出勤率低于80%', '2025-04-08 09:10'
    UNION ALL SELECT '2023004', 'red', '挂科预警', '多门课程不及格', '高等数学、数据结构等课程成绩低于 60 分，学业风险较高', '数据结构', '["高等数学（下）","数据结构"]', '2025-04-09', 'pending', '建议安排一对一辅导并联系家长', '累计不及格达3门', '2025-04-09 14:00'
    UNION ALL SELECT '2023005', 'yellow', '成绩预警', '成绩处于边缘', 'GPA 2.62 接近预警线，个别课程成绩偏低', '计算机网络', '[]', '2025-04-10', 'resolved', '已完成谈话并约定每周复盘', 'GPA接近预警线', '2025-04-10 11:00'
    UNION ALL SELECT '2024004', 'orange', '成绩预警', 'GPA 低于 2.5', '当前 GPA 2.15，数据类课程成绩偏低，需制定提升计划', NULL, '["高等数学（下）"]', '2025-04-11', 'pending', '建议参加数学辅导并调整学习计划', 'GPA低于2.5', '2025-04-11 16:40'
) v
JOIN t_user u ON u.account = v.account
JOIN t_student s ON s.user_id = u.id
WHERE NOT EXISTS (
    SELECT 1 FROM t_alert a WHERE a.student_id = s.id AND a.title = v.title AND a.trigger_date = v.trigger_date
);

-- 历史预警（近四学期趋势展示用，均已闭环）
INSERT INTO t_alert (student_id, level, type, title, description, course, failed_courses, trigger_date, status, suggestion, trigger_event, pushed_at)
SELECT s.id, v.level, v.type, v.title, v.description, v.course, v.failed, v.trigger_date, v.status, v.suggestion, v.trigger_event, v.pushed_at
FROM (
    SELECT '2022003' account, 'orange' level, '成绩预警' type, 'GPA 低于 2.5' title, '上学期 GPA 2.40，低于专业平均水平' description, NULL course, '[]' failed, '2023-11-15' trigger_date, 'resolved' status, '已完成约谈并制定提升计划' suggestion, 'GPA低于2.5' trigger_event, '2023-11-15 10:00' pushed_at
    UNION ALL SELECT '2022005', 'red', '挂科预警', '累计挂科超过3门', '多门课程不及格，GPA 1.60', '高等数学（下）', '["高等数学（上）","高等数学（下）","线性代数"]', '2024-01-20', 'resolved', '已安排专项帮扶并跟踪复查', '累计不及格超3门', '2024-01-20 14:30'
    UNION ALL SELECT '2022007', 'yellow', '出勤预警', '缺勤次数较多', '上学期缺勤累计 7 次，存在学业下滑风险', NULL, '[]', '2024-05-10', 'resolved', '已谈话并约定出勤目标', '学期缺勤达7次', '2024-05-10 09:20'
    UNION ALL SELECT '2022010', 'orange', '成绩预警', 'GPA 低于 2.5', '上学期 GPA 1.85，成绩持续偏低', NULL, '["线性代数"]', '2024-11-18', 'resolved', '已安排一对一学业咨询', 'GPA低于2.5', '2024-11-18 16:00'
    UNION ALL SELECT '2022002', 'yellow', '成绩预警', '成绩出现下滑', '上学期 GPA 由 3.10 降至 2.95，成绩呈下滑趋势', NULL, '[]', '2025-01-15', 'resolved', '已完成谈心谈话并约定每周复盘', 'GPA连续下滑', '2025-01-15 11:00'
    UNION ALL SELECT '2022012', 'orange', '挂科预警', '多门课程成绩偏低', '上学期操作系统成绩低于 60 分，多门课程需补考', '操作系统', '["操作系统"]', '2025-03-20', 'resolved', '已参加补考辅导小组', '课程成绩低于60分', '2025-03-20 15:10'
) v
JOIN t_user u ON u.account = v.account
JOIN t_student s ON s.user_id = u.id
WHERE NOT EXISTS (
    SELECT 1 FROM t_alert a WHERE a.student_id = s.id AND a.title = v.title AND a.trigger_date = v.trigger_date
);

-- ============ 培养方案种子数据 ============

INSERT IGNORE INTO t_major_plan (major_id, grade, dimensions, standard, actual, tips)
SELECT m.id, v.grade, v.dimensions, v.standard, v.actual, v.tips
FROM (
    SELECT '计算机科学与技术' major_name, '2021' grade, '["公共基础","专业必修","专业选修","实践环节","通识教育"]' dimensions, '[48,62,20,18,12]' standard, '[48,62,22,18,12]' actual, '["专业选修超出计划，建议聚焦核心方向","实践环节完成度良好"]' tips
    UNION ALL SELECT '计算机科学与技术', '2022', '["公共基础","专业必修","专业选修","实践环节","通识教育"]', '[48,62,20,18,12]', '[46,58,24,16,12]', '["公共基础学分完成度略低","专业选修超出计划，建议聚焦核心方向"]'
    UNION ALL SELECT '计算机科学与技术', '2023', '["公共基础","专业必修","专业选修","实践环节","通识教育"]', '[48,62,20,18,12]', '[44,55,26,14,12]', '["公共基础与专业必修完成度偏低","建议加强实践环节学分积累"]'
    UNION ALL SELECT '计算机科学与技术', '2024', '["公共基础","专业必修","专业选修","实践环节","通识教育"]', '[48,62,20,18,12]', '[24,32,10,8,12]', '["大一处于基础学习阶段，学分完成符合预期","建议提前规划专业选修方向"]'
    UNION ALL SELECT '软件工程', '2023', '["公共基础","专业必修","专业选修","实践环节","通识教育"]', '[48,60,22,20,12]', '[45,56,24,17,12]', '["专业必修完成度略低","实践环节需加强项目训练"]'
    UNION ALL SELECT '数据科学', '2024', '["公共基础","专业必修","专业选修","实践环节","通识教育"]', '[48,62,20,18,12]', '[25,33,11,8,12]', '["大一基础阶段完成度符合预期","建议关注数学与编程基础"]'
) v
JOIN t_major m ON m.name = v.major_name AND m.department_id = 1;

-- ============ 职位缓存种子数据 ============

INSERT INTO t_job_cache (platform, platform_label, company_name, company_size, company_stage, job_title, salary_range, city, district, education, experience, tags, highlights, match_score, match_reasons, publish_date, source_url, category, published_at, crawled_at) VALUES
('boss', 'BOSS直聘', '阿里云计算有限公司', '10000人以上', '上市企业', '后端开发工程师（Java）', '20-35K·14薪', '杭州', '余杭区·阿里巴巴西溪园区', '本科', '应届/1年以内', '["Java","Spring Boot","MySQL","分布式"]', '["顶级平台背书","应届可投","校招直通"]', 91, '["Java基础课成绩优秀","应届生专项岗位","GPA 3.62 符合校招要求"]', '2026-07-02', 'https://www.zhipin.com/job_detail/', 'backend', '2026-07-02 09:00:00', '2026-07-02 10:00:00'),
('boss', 'BOSS直聘', '网易杭州研究院', '10000人以上', '上市企业', 'Java后端开发（校招）', '18-28K·13薪', '杭州', '滨江区·网易大厦', '本科', '应届', '["Java","Netty","Redis","Kafka"]', '["网易游戏业务线","校招专项","导师制"]', 87, '["计算机科学专业对口","校招岗位，应届可投","网络编程技术与课程知识匹配"]', '2026-07-01', 'https://campus.163.com/', 'backend', '2026-07-01 09:00:00', '2026-07-01 10:30:00'),
('51job', '前程无忧', '海康威视数字技术股份有限公司', '10000人以上', '上市企业', '软件研发工程师（后端）', '15-22K·14薪', '杭州', '滨江区·海康威视园区', '本科', '应届/1年以内', '["Java","Go","Linux","C++"]', '["物联网龙头企业","技术积累强","福利完善"]', 85, '["操作系统课程背景契合嵌入式方向","本地头部企业","应届生岗位"]', '2026-07-01', 'https://www.51job.com/', 'backend', '2026-07-01 11:00:00', '2026-07-01 12:00:00'),
('zhilian', '智联招聘', '浙江大华技术股份有限公司', '10000人以上', '上市企业', '后端研发工程师（Go/Java）', '14-20K·13薪', '杭州', '滨江区', '本科', '应届', '["Go","Java","gRPC","Docker"]', '["全球视频监控龙头","应届专项通道","股票激励"]', 82, '["计算机网络课程成绩良好契合网络编程岗","计算机科学专业对口","杭州本地"]', '2026-06-30', 'https://www.zhaopin.com/', 'backend', '2026-06-30 09:00:00', '2026-06-30 10:00:00'),
('boss', 'BOSS直聘', '新华三技术有限公司（H3C）', '10000人以上', '上市企业', '云平台后端开发工程师', '15-24K·13薪', '杭州', '滨江区', '本科', '应届/1年', '["Java","Kubernetes","OpenStack","MySQL"]', '["云计算国产化先驱","技术氛围好","晋升体系完善"]', 80, '["数据库原理课程成绩优异","云平台方向契合专业知识体系","杭州本地大厂"]', '2026-06-29', 'https://www.zhipin.com/job_detail/', 'backend', '2026-06-29 09:00:00', '2026-06-29 10:00:00'),
('liepin', '猎聘', '浙江蘑菇街有限公司', '2000-9999人', '上市企业', 'Java后端开发（电商）', '13-20K·14薪', '杭州', '西湖区', '本科', '应届/1年以内', '["Java","Spring Cloud","ElasticSearch","RabbitMQ"]', '["电商微服务架构","技术栈新","弹性工作"]', 76, '["Spring微服务架构与课程项目经验匹配","应届友好","杭州互联网环境"]', '2026-06-28', 'https://www.liepin.com/', 'backend', '2026-06-28 09:00:00', '2026-06-28 10:00:00'),
('boss', 'BOSS直聘', '阿里巴巴达摩院', '10000人以上', '上市企业', 'AI算法工程师（校招）', '25-45K·16薪', '杭州', '余杭区', '本科/硕士', '应届', '["PyTorch","LLM","Python","深度学习"]', '["世界顶级AI研究院","顶级薪资","论文产出机会"]', 78, '["数学建模竞赛背景加分","GPA 3.62符合校招门槛","计算机专业背景匹配"]', '2026-07-02', 'https://damo.alibaba.com/', 'ai', '2026-07-02 09:00:00', '2026-07-02 10:00:00'),
('boss', 'BOSS直聘', '网易灵犀互娱', '5000-9999人', '上市企业', 'AI应用开发工程师', '18-30K·13薪', '杭州', '滨江区', '本科', '应届/1年', '["Python","LangChain","RAG","Prompt Engineering"]', '["游戏AI方向","应用落地项目","技术成长快"]', 75, '["Python基础与AI应用开发契合","计算机专业背景","应届生通道"]', '2026-07-01', 'https://www.zhipin.com/job_detail/', 'ai', '2026-07-01 09:00:00', '2026-07-01 10:00:00'),
('51job', '前程无忧', '浙江工业大学科研院（校企合作）', '500-999人', '事业单位', '机器学习算法研究员（应届）', '12-18K', '杭州', '拱墅区', '本科', '应届', '["Python","sklearn","TensorFlow","数据分析"]', '["高校科研环境","稳定福利","可申请科研补贴"]', 72, '["GPA 3.62符合高校科研门槛","有助于积累科研经验","计算机+数学背景契合"]', '2026-06-30', 'https://www.51job.com/', 'ai', '2026-06-30 09:00:00', '2026-06-30 10:00:00'),
('zhilian', '智联招聘', '杭州深睿博联科技有限公司', '500-999人', 'B轮', '计算机视觉工程师（应届）', '15-22K·13薪', '杭州', '西湖区', '本科', '应届', '["OpenCV","YOLO","Python","ONNX"]', '["医疗AI赛道","成长型公司","股权激励"]', 70, '["数据库+算法基础契合CV工程岗","成长型公司发展空间大","杭州本地AI明星企业"]', '2026-06-29', 'https://www.zhaopin.com/', 'ai', '2026-06-29 09:00:00', '2026-06-29 10:00:00'),
('boss', 'BOSS直聘', '华数传媒网络有限公司', '2000-4999人', '上市企业', 'AIGC内容工程师', '12-18K·13薪', '杭州', '上城区', '本科', '应届/1年', '["Python","Stable Diffusion","ChatGPT API","内容生成"]', '["传媒+AI结合","应届友好","杭州国企背景"]', 65, '["AIGC方向入门门槛相对较低","计算机专业背景满足要求","本地国企稳定"]', '2026-06-27', 'https://www.zhipin.com/job_detail/', 'ai', '2026-06-27 09:00:00', '2026-06-27 10:00:00'),
('boss', 'BOSS直聘', '浙江省数字经济研究院', '100-499人', '事业单位', '数字化项目助理（应届生培养计划）', '9-14K', '杭州', '滨江区', '本科', '应届', '["项目管理","Axure","Visio","需求分析"]', '["政府背景稳定","数字化转型核心岗","晋升空间清晰"]', 74, '["GPA 3.62符合事业单位门槛","计算机专业为项目管理加分","发展意向维度得分高"]', '2026-07-01', 'https://www.zhipin.com/job_detail/', 'pm', '2026-07-01 09:00:00', '2026-07-01 10:00:00'),
('51job', '前程无忧', '中国联合网络通信有限公司浙江分公司', '10000人以上', '央企', 'IT项目管理培训生', '8-12K', '杭州', '西湖区·联通大厦', '本科', '应届', '["PMP","项目管理","IT运维","需求调研"]', '["央企背景","五险一金+公积金","带薪培训6个月"]', 71, '["计算机专业背景是IT项管的核心优势","央企稳定性高","GPA符合培训生要求"]', '2026-06-30', 'https://www.51job.com/', 'pm', '2026-06-30 09:00:00', '2026-06-30 10:00:00'),
('boss', 'BOSS直聘', '浙商银行股份有限公司', '10000人以上', '上市企业', '科技项目经理（校招）', '10-16K', '杭州', '上城区', '本科', '应届', '["金融科技","项目管理","敏捷开发","Scrum"]', '["金融科技方向","银行福利待遇好","技术与管理双通道"]', 68, '["计算机+管理背景契合金融科技","银行业技术岗起点稳","GPA 3.62满足校招门槛"]', '2026-06-29', 'https://www.zhipin.com/job_detail/', 'pm', '2026-06-29 09:00:00', '2026-06-29 10:00:00'),
('liepin', '猎聘', '用友网络科技股份有限公司（浙江区）', '10000人以上', '上市企业', '实施项目管理工程师', '10-15K·13薪', '杭州', '拱墅区', '本科', '应届/1年', '["ERP实施","项目管理","SQL","客户沟通"]', '["ERP行业龙头","全国出差机会","快速积累项目经验"]', 64, '["数据库原理课程成绩优异契合ERP实施","计算机背景满足技术需求","积累项目管理经验的好起点"]', '2026-06-27', 'https://www.liepin.com/', 'pm', '2026-06-27 09:00:00', '2026-06-27 10:00:00');

-- ============ 院校缓存种子数据 ============

INSERT INTO t_school_cache (type, name, name_zh, country, country_emoji, location, `rank`, tier, programs, admission_gpa, exam_requirements, language_requirement, highlights, match_score, match_reasons, official_url, deadline) VALUES
('grad', '浙江大学', NULL, NULL, NULL, '浙江 · 杭州', 'QS 44 / 软科 3', 'top', '["计算机科学与技术","软件工程","网络空间安全"]', '3.7+ 推荐', '["数学一","英语一","408计算机学科专业基础"]', NULL, '["C9联盟","双一流A+","本省顶尖","导师资源丰富"]', 72, '["本省院校，地域优势明显","GPA 3.62 接近往年录取线下限","计算机专业背景完全对口"]', 'https://www.zju.edu.cn', NULL),
('grad', '南京大学', NULL, NULL, NULL, '江苏 · 南京', 'QS 133 / 软科 6', 'top', '["计算机科学与技术","软件工程","人工智能"]', '3.6+ 推荐', '["数学一","英语一","408计算机学科专业基础"]', NULL, '["C9联盟","双一流A","学术氛围浓厚","长三角核心"]', 75, '["GPA 3.62 达到往年录取中位线","华东五校之一，就业认可度高","长三角地区，地理位置优"]', 'https://www.nju.edu.cn', NULL),
('grad', '东南大学', NULL, NULL, NULL, '江苏 · 南京', 'QS 262 / 软科 18', 'good', '["计算机科学与技术","软件工程","网络空间安全"]', '3.5+', '["数学一","英语一","408计算机学科专业基础"]', NULL, '["985","双一流A-","工科强校","就业口碑好"]', 82, '["GPA 3.62 超过往年录取平均线","985院校，性价比高","计算机学科评估A-，实力强劲"]', 'https://www.seu.edu.cn', NULL),
('grad', '华东师范大学', NULL, NULL, NULL, '上海', 'QS 291 / 软科 28', 'good', '["计算机科学与技术","软件工程","数据科学"]', '3.4+', '["数学一","英语一","408计算机学科专业基础"]', NULL, '["985","双一流B+","上海户口优势","数据科学特色"]', 85, '["师范类985，竞争相对缓和","GPA 3.62 优势明显","上海地理位置，就业机会多"]', 'https://www.ecnu.edu.cn', NULL),
('grad', '杭州电子科技大学', NULL, NULL, NULL, '浙江 · 杭州', '软科 91', 'match', '["计算机科学与技术","软件工程","网络空间安全"]', '3.0+', '["数学一","英语一","408计算机学科专业基础"]', NULL, '["电子信息特色","杭州互联网近水楼台","就业率高","本地优势"]', 92, '["本省院校，地域熟悉","GPA 3.62 优势极大，几乎稳进","杭州互联网企业认可度高"]', 'https://www.hdu.edu.cn', NULL),
('grad', '宁波大学', NULL, NULL, NULL, '浙江 · 宁波', '软科 109', 'match', '["计算机科学与技术","软件工程"]', '3.0+', '["数学一","英语一","408计算机学科专业基础"]', NULL, '["双一流","浙江省重点","宁波地理优势","学费适中"]', 91, '["本省双一流，保底稳妥","GPA 3.62 远超录取线","宁波经济发达，就业前景好"]', 'https://www.nbu.edu.cn', NULL),
('overseas', 'Carnegie Mellon University', '卡耐基梅隆大学', '美国', '🇺🇸', NULL, 'QS CS 1 / USNews CS 1', 'top', '["Master of Science in Computer Science","Master of Software Engineering"]', '3.7+ (4.0 scale)', NULL, '托福 100+ / 雅思 7.0+', '["全球CS第一","SV直通车","AI/ML全球顶尖","就业薪资极高"]', 62, '["计算机专业背景完全对口","GPA 3.62 低于往年录取中位线","需托福/GRE高分+强科研背景补强"]', 'https://www.cmu.edu', '2026-12-15'),
('overseas', 'ETH Zurich', '苏黎世联邦理工学院', '瑞士', '🇨🇭', NULL, 'QS 7 / QS CS 7', 'top', '["Master in Computer Science","Master in Data Science"]', '3.7+ (4.0 scale)', NULL, '托福 100+ / 雅思 7.0+', '["欧陆第一","学费低","AI研究前沿","毕业留欧机会"]', 68, '["欧洲顶尖CS项目，性价比高","GPA 3.62 接近录取下限","学费仅1500瑞郎/年"]', 'https://ethz.ch', '2026-12-15'),
('overseas', 'University of Southern California', '南加州大学', '美国', '🇺🇸', NULL, 'QS 116 / USNews CS 20', 'good', '["Master of Science in Computer Science","Master of Science in Data Science"]', '3.3+', NULL, '托福 90+ / 雅思 6.5+', '["洛杉矶地理优势","中国学生友好","CS就业强","STEM OPT 3年"]', 82, '["GPA 3.62 高于录取平均线","CS项目规模大，录取相对友好","洛杉矶科技公司多，实习机会丰富"]', 'https://www.usc.edu', '2027-01-15'),
('overseas', 'The Chinese University of Hong Kong', '香港中文大学', '中国香港', '🇭🇰', NULL, 'QS 47 / QS CS 30', 'good', '["MSc in Computer Science","MSc in Information Engineering"]', '3.3+', NULL, '托福 79+ / 雅思 6.5+', '["港三甲","一年制","性价比高","离家近"]', 85, '["GPA 3.62 优势明显","一年制硕士，时间成本低","香港就业+回内地认可度双优"]', 'https://www.cuhk.edu.hk', '2027-02-28'),
('overseas', 'The University of Hong Kong', '香港大学', '中国香港', '🇭🇰', NULL, 'QS 26 / QS CS 31', 'match', '["MSc in Computer Science","MSc in Data Science"]', '3.0+', NULL, '托福 80+ / 雅思 6.0+ (小分5.5)', '["港排名第一","一年制","rolling录取","就业认可度高"]', 88, '["GPA 3.62 远超录取线","rolling录取，早申优势大","一年制+离家近，性价比极高"]', 'https://www.hku.hk', '2027-04-30 (rolling)'),
('overseas', 'University of Melbourne', '墨尔本大学', '澳大利亚', '🇦🇺', NULL, 'QS 13 / QS CS 42', 'match', '["Master of Information Technology","Master of Data Science"]', '3.0+', NULL, '托福 79+ / 雅思 6.5+', '["澳洲第一","移民友好","2年工签","气候宜居"]', 90, '["GPA 3.62 优势极大","澳洲移民政策友好，有PR机会","CS就业市场稳定"]', 'https://www.unimelb.edu.au', '2027-03-31');
