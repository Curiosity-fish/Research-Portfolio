-- V6: 多样化数据集
-- 目标：学生总数 1000（现有 143 + 新增 857），覆盖 2021-2024 四个年级，
-- 三个学院，每个学院 4-5 个专业；同步补齐新专业对应的系主任账号。
SET NAMES utf8mb4;

-- ============ 确保三学院结构与样本专业存在（干净库也可直接执行） ============
INSERT INTO t_department (id, name, code, dean_id)
SELECT 100100, '数据科学与工程测试学院', 'TEST_01', NULL
WHERE NOT EXISTS (SELECT 1 FROM t_department WHERE name = '数据科学与工程测试学院');
INSERT INTO t_department (id, name, code, dean_id)
SELECT 100101, '智能计算测试学院', 'TEST_02', NULL
WHERE NOT EXISTS (SELECT 1 FROM t_department WHERE name = '智能计算测试学院');
INSERT INTO t_major (id, name, code, department_id, degree_type)
SELECT 100200, '数据科学与大数据技术', 'MAJOR_01', 100100, '工学学士'
WHERE NOT EXISTS (SELECT 1 FROM t_major WHERE name = '数据科学与大数据技术' AND department_id = 100100);
INSERT INTO t_major (id, name, code, department_id, degree_type)
SELECT 100201, '软件工程', 'MAJOR_02', 100101, '工学学士'
WHERE NOT EXISTS (SELECT 1 FROM t_major WHERE name = '软件工程' AND department_id = 100101);
INSERT INTO t_major (id, name, code, department_id, degree_type)
SELECT 100202, '计算机科学与技术', 'MAJOR_03', 100100, '工学学士'
WHERE NOT EXISTS (SELECT 1 FROM t_major WHERE name = '计算机科学与技术' AND department_id = 100100);
INSERT INTO t_major (id, name, code, department_id, degree_type)
SELECT 100203, '人工智能', 'MAJOR_04', 100101, '工学学士'
WHERE NOT EXISTS (SELECT 1 FROM t_major WHERE name = '人工智能' AND department_id = 100101);
UPDATE t_department SET dean_id = (SELECT id FROM (SELECT id FROM t_user WHERE account = 'L2024001') t) WHERE id = 100100 AND dean_id IS NULL;
UPDATE t_department SET dean_id = (SELECT id FROM (SELECT id FROM t_user WHERE account = 'L2024002') t) WHERE id = 100101 AND dean_id IS NULL;

-- ============ 新增专业（每学院 4-5 个） ============
INSERT INTO t_major (id, name, code, department_id, degree_type)
SELECT 5, '人工智能', 'AI1', 1, '工学学士'
WHERE NOT EXISTS (SELECT 1 FROM t_major WHERE name = '人工智能' AND department_id = 1);
INSERT INTO t_major (id, name, code, department_id, degree_type)
SELECT 6, '网络空间安全', 'NSS', 1, '工学学士'
WHERE NOT EXISTS (SELECT 1 FROM t_major WHERE name = '网络空间安全' AND department_id = 1);
INSERT INTO t_major (id, name, code, department_id, degree_type)
SELECT 100204, '人工智能', 'AI2', 100100, '工学学士'
WHERE NOT EXISTS (SELECT 1 FROM t_major WHERE name = '人工智能' AND department_id = 100100);
INSERT INTO t_major (id, name, code, department_id, degree_type)
SELECT 100205, '统计学', 'STAT', 100100, '工学学士'
WHERE NOT EXISTS (SELECT 1 FROM t_major WHERE name = '统计学' AND department_id = 100100);
INSERT INTO t_major (id, name, code, department_id, degree_type)
SELECT 100206, '信息管理与信息系统', 'IMIS', 100100, '工学学士'
WHERE NOT EXISTS (SELECT 1 FROM t_major WHERE name = '信息管理与信息系统' AND department_id = 100100);
INSERT INTO t_major (id, name, code, department_id, degree_type)
SELECT 100207, '物联网工程', 'IOT', 100101, '工学学士'
WHERE NOT EXISTS (SELECT 1 FROM t_major WHERE name = '物联网工程' AND department_id = 100101);
INSERT INTO t_major (id, name, code, department_id, degree_type)
SELECT 100208, '信息安全', 'SEC', 100101, '工学学士'
WHERE NOT EXISTS (SELECT 1 FROM t_major WHERE name = '信息安全' AND department_id = 100101);
INSERT INTO t_major (id, name, code, department_id, degree_type)
SELECT 100209, '数字媒体技术', 'DMT', 100101, '工学学士'
WHERE NOT EXISTS (SELECT 1 FROM t_major WHERE name = '数字媒体技术' AND department_id = 100101);

-- ============ 新增专业对应的系主任账号（密码 123456） ============
INSERT IGNORE INTO t_user (account, name, password, role, college, major, class_name, grade) VALUES
    ('D20210004', '人工智能系主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'department', '数学与计算机科学学院', '人工智能', NULL, NULL),
    ('D20210005', '网络空间安全系主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'department', '数学与计算机科学学院', '网络空间安全', NULL, NULL),
    ('D2024005', '人工智能系主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'department', '数据科学与工程测试学院', '人工智能', NULL, NULL),
    ('D2024006', '统计学系主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'department', '数据科学与工程测试学院', '统计学', NULL, NULL),
    ('D2024007', '信息管理系主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'department', '数据科学与工程测试学院', '信息管理与信息系统', NULL, NULL),
    ('D2024008', '物联网系主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'department', '智能计算测试学院', '物联网工程', NULL, NULL),
    ('D2024009', '信息安全系主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'department', '智能计算测试学院', '信息安全', NULL, NULL),
    ('D2024010', '数字媒体系主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'department', '智能计算测试学院', '数字媒体技术', NULL, NULL)
;

-- ============ 为 15 个专业 x 4 个年级创建班级 ============
INSERT INTO t_class (name, grade, major_id, advisor_id, total_students)
SELECT CONCAT(ma.short, g.grade, '01'), g.grade, ma.major_id, 1, 0
FROM (SELECT 2021 grade UNION ALL SELECT 2022 UNION ALL SELECT 2023 UNION ALL SELECT 2024) g
CROSS JOIN (
    SELECT 0 idx, 1 major_id, 'CS' short
UNION ALL
    SELECT 1 idx, 2 major_id, 'SE' short
UNION ALL
    SELECT 2 idx, 3 major_id, 'DS' short
UNION ALL
    SELECT 3 idx, 5 major_id, 'AI1' short
UNION ALL
    SELECT 4 idx, 6 major_id, 'NSS' short
UNION ALL
    SELECT 5 idx, 100200 major_id, 'BDA' short
UNION ALL
    SELECT 6 idx, 100202 major_id, 'CS2' short
UNION ALL
    SELECT 7 idx, 100204 major_id, 'AI2' short
UNION ALL
    SELECT 8 idx, 100205 major_id, 'STAT' short
UNION ALL
    SELECT 9 idx, 100206 major_id, 'IMIS' short
UNION ALL
    SELECT 10 idx, 100201 major_id, 'SE2' short
UNION ALL
    SELECT 11 idx, 100203 major_id, 'AI3' short
UNION ALL
    SELECT 12 idx, 100207 major_id, 'IOT' short
UNION ALL
    SELECT 13 idx, 100208 major_id, 'SEC' short
UNION ALL
    SELECT 14 idx, 100209 major_id, 'DMT' short
) ma
WHERE NOT EXISTS (SELECT 1 FROM t_class c WHERE c.name = CONCAT(ma.short, g.grade, '01'));

-- ============ 新增 857 名学生账号（20990001-20990857） ============
INSERT IGNORE INTO t_user (account, name, password, role, college, major, class_name, grade)
SELECT CONCAT('2099', LPAD(nums.seq, 4, '0')),
       CONCAT(ELT(1 + MOD(nums.seq - 1, 20), '张',
'李',
'王',
'刘',
'陈',
'杨',
'赵',
'黄',
'周',
'吴',
'徐',
'孙',
'胡',
'朱',
'高',
'林',
'何',
'郭',
'马',
'罗'),
              ELT(1 + MOD(nums.seq * 7, 20), '伟',
'芳',
'娜',
'敏',
'静',
'磊',
'军',
'洋',
'勇',
'艳',
'杰',
'娟',
'涛',
'明',
'超',
'霞',
'平',
'刚',
'桂英',
'志强')),
       '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'student', dep.name, m.name,
       CONCAT(ma.short, 2021 + MOD(FLOOR((nums.seq - 1) / 15), 4), '01'),
       2021 + MOD(FLOOR((nums.seq - 1) / 15), 4)
FROM (
  SELECT d0.n + d1.n * 10 + d2.n * 100 + 1 AS seq
  FROM (SELECT 0 n UNION ALL SELECT 1 UNION ALL SELECT 2 UNION ALL SELECT 3 UNION ALL SELECT 4
        UNION ALL SELECT 5 UNION ALL SELECT 6 UNION ALL SELECT 7 UNION ALL SELECT 8 UNION ALL SELECT 9) d0
  CROSS JOIN (SELECT 0 n UNION ALL SELECT 1 UNION ALL SELECT 2 UNION ALL SELECT 3 UNION ALL SELECT 4
              UNION ALL SELECT 5 UNION ALL SELECT 6 UNION ALL SELECT 7 UNION ALL SELECT 8 UNION ALL SELECT 9) d1
  CROSS JOIN (SELECT 0 n UNION ALL SELECT 1 UNION ALL SELECT 2 UNION ALL SELECT 3 UNION ALL SELECT 4
              UNION ALL SELECT 5 UNION ALL SELECT 6 UNION ALL SELECT 7 UNION ALL SELECT 8 UNION ALL SELECT 9) d2
) nums
JOIN (
    SELECT 0 idx, 1 major_id, 'CS' short
UNION ALL
    SELECT 1 idx, 2 major_id, 'SE' short
UNION ALL
    SELECT 2 idx, 3 major_id, 'DS' short
UNION ALL
    SELECT 3 idx, 5 major_id, 'AI1' short
UNION ALL
    SELECT 4 idx, 6 major_id, 'NSS' short
UNION ALL
    SELECT 5 idx, 100200 major_id, 'BDA' short
UNION ALL
    SELECT 6 idx, 100202 major_id, 'CS2' short
UNION ALL
    SELECT 7 idx, 100204 major_id, 'AI2' short
UNION ALL
    SELECT 8 idx, 100205 major_id, 'STAT' short
UNION ALL
    SELECT 9 idx, 100206 major_id, 'IMIS' short
UNION ALL
    SELECT 10 idx, 100201 major_id, 'SE2' short
UNION ALL
    SELECT 11 idx, 100203 major_id, 'AI3' short
UNION ALL
    SELECT 12 idx, 100207 major_id, 'IOT' short
UNION ALL
    SELECT 13 idx, 100208 major_id, 'SEC' short
UNION ALL
    SELECT 14 idx, 100209 major_id, 'DMT' short
) ma ON ma.idx = MOD(nums.seq - 1, 15)
JOIN t_major m ON m.id = ma.major_id
JOIN t_department dep ON dep.id = m.department_id
JOIN (SELECT 1000 - (SELECT COUNT(*) FROM t_student) AS target) tg ON 1 = 1
WHERE nums.seq <= tg.target;

-- ============ 新增学生档案（GPA/排名/预警等级按序号确定性生成） ============
INSERT INTO t_student (student_id, user_id, class_id, major_id, department_id, gpa, `rank`, total_students, alert_level, total_credits, required_credits)
SELECT u.account, u.id, cl.id, m.id, m.department_id,
       ROUND(2.10 + MOD(seq * 7, 190) / 100.0, 2),
       1 + MOD(seq * 13, 69),
       70,
       CASE WHEN MOD(seq, 17) = 0 THEN 'red' WHEN MOD(seq, 11) = 0 THEN 'orange' WHEN MOD(seq, 7) = 0 THEN 'yellow' ELSE 'none' END,
       (CAST(u.grade AS UNSIGNED) - 2021) * 40 + 20,
       176.0
FROM t_user u
JOIN (SELECT account, CAST(SUBSTRING(account, 5) AS UNSIGNED) AS seq FROM t_user WHERE account LIKE '2099%') ns ON ns.account = u.account
JOIN t_department dep ON dep.name = u.college
JOIN t_major m ON m.name = u.major AND m.department_id = dep.id
JOIN (
    SELECT 0 idx, 1 major_id, 'CS' short
UNION ALL
    SELECT 1 idx, 2 major_id, 'SE' short
UNION ALL
    SELECT 2 idx, 3 major_id, 'DS' short
UNION ALL
    SELECT 3 idx, 5 major_id, 'AI1' short
UNION ALL
    SELECT 4 idx, 6 major_id, 'NSS' short
UNION ALL
    SELECT 5 idx, 100200 major_id, 'BDA' short
UNION ALL
    SELECT 6 idx, 100202 major_id, 'CS2' short
UNION ALL
    SELECT 7 idx, 100204 major_id, 'AI2' short
UNION ALL
    SELECT 8 idx, 100205 major_id, 'STAT' short
UNION ALL
    SELECT 9 idx, 100206 major_id, 'IMIS' short
UNION ALL
    SELECT 10 idx, 100201 major_id, 'SE2' short
UNION ALL
    SELECT 11 idx, 100203 major_id, 'AI3' short
UNION ALL
    SELECT 12 idx, 100207 major_id, 'IOT' short
UNION ALL
    SELECT 13 idx, 100208 major_id, 'SEC' short
UNION ALL
    SELECT 14 idx, 100209 major_id, 'DMT' short
) ma ON ma.major_id = m.id
JOIN t_class cl ON cl.name = CONCAT(ma.short, u.grade, '01')
WHERE u.account LIKE '2099%'
  AND NOT EXISTS (SELECT 1 FROM t_student s WHERE s.student_id = u.account);

-- ============ 回填班级人数 ============
UPDATE t_class c
SET c.total_students = (SELECT COUNT(*) FROM t_student s JOIN t_user u ON u.id = s.user_id WHERE u.class_name = c.name);

-- ============ 为新班级创建当前学期课程班级 ============
INSERT INTO t_course_class (course_id, term, class_name, teacher_id, max_students, enrolled_students, status)
SELECT c.id, '2024-2025-1', cl.name, 2, cl.total_students, cl.total_students, 'active'
FROM t_course c
JOIN t_class cl ON cl.name REGEXP '^(CS|SE|DS|AI1|NSS|BDA|CS2|AI2|STAT|IMIS|SE2|AI3|IOT|SEC|DMT)[0-9]{4}01$'
WHERE c.id IN (1,2,3,4,5,6)
  AND NOT EXISTS (SELECT 1 FROM t_course_class cc WHERE cc.course_id = c.id AND cc.class_name = cl.name AND cc.term = '2024-2025-1');

-- ============ 生成当前学期成绩 ============
INSERT IGNORE INTO t_grade (student_id, course_class_id, score, status, term, exam_date)
SELECT s.id, cc.id,
       45 + MOD(seq * 7 + cc.course_id * 11, 56),
       'passed', '2024-2025-1', '2025-01-15'
FROM t_student s
JOIN t_user u ON u.id = s.user_id
JOIN (SELECT account, CAST(SUBSTRING(account, 5) AS UNSIGNED) AS seq FROM t_user WHERE account LIKE '2099%') ns ON ns.account = u.account
JOIN t_class cl ON cl.id = s.class_id
JOIN t_course_class cc ON cc.class_name = cl.name AND cc.term = '2024-2025-1' AND cc.course_id IN (1,2,3,4,5,6)
WHERE u.account LIKE '2099%';

UPDATE t_grade g
JOIN t_student s ON s.id = g.student_id
JOIN t_user u ON u.id = s.user_id
SET g.status = IF(g.score >= 60, 'passed', 'failed'),
    g.grade_point = CASE
        WHEN g.score >= 95 THEN 4.0 WHEN g.score >= 90 THEN 3.7 WHEN g.score >= 85 THEN 3.3
        WHEN g.score >= 80 THEN 3.0 WHEN g.score >= 75 THEN 2.7 WHEN g.score >= 70 THEN 2.3
        WHEN g.score >= 65 THEN 2.0 WHEN g.score >= 60 THEN 1.7 ELSE 0.0 END
WHERE u.account LIKE '2099%' AND g.term = '2024-2025-1';

-- ============ 生成各年级 GPA 历史 ============
INSERT IGNORE INTO t_gpa_history (student_id, term, gpa, avg_gpa, `rank`, total_students)
SELECT s.id, t.term,
       ROUND(LEAST(4.0, GREATEST(0.8, s.gpa + (t.idx - 7) * 0.04 + MOD(seq * 3, 5) * 0.01)), 2),
       ROUND(3.0 + (t.idx - 7) * 0.03, 2),
       GREATEST(1, s.`rank` + (7 - t.idx) * 2),
       s.total_students
FROM t_student s
JOIN t_user u ON u.id = s.user_id
JOIN (SELECT account, CAST(SUBSTRING(account, 5) AS UNSIGNED) AS seq FROM t_user WHERE account LIKE '2099%') ns ON ns.account = u.account
JOIN (
    SELECT '2021-2022-1' term, 1 idx UNION ALL SELECT '2021-2022-2', 2
    UNION ALL SELECT '2022-2023-1', 3 UNION ALL SELECT '2022-2023-2', 4
    UNION ALL SELECT '2023-2024-1', 5 UNION ALL SELECT '2023-2024-2', 6
    UNION ALL SELECT '2024-2025-1', 7
) t
WHERE u.account LIKE '2099%'
  AND CAST(LEFT(t.term, 4) AS SIGNED) - CAST(u.grade AS SIGNED) BETWEEN 0 AND 3;

-- ============ 生成当前学期画像分数 ============
INSERT IGNORE INTO t_profile_score (student_id, term, dimension_key, label, score, avg_score, max_score, description, details)
SELECT s.id, '2024-2025-1', 'academic', '学业成绩',
       CASE WHEN s.gpa >= 3.5 THEN 88 WHEN s.gpa >= 3.0 THEN 76 WHEN s.gpa >= 2.5 THEN 64 WHEN s.gpa >= 2.0 THEN 54 ELSE 42 END,
       72, 100,
       CONCAT('GPA ', FORMAT(s.gpa, 2), '，专业排名 ', s.`rank`),
       JSON_ARRAY(CONCAT('GPA: ', FORMAT(s.gpa, 2), ' / 4.0'))
FROM t_student s
JOIN t_user u ON u.id = s.user_id
WHERE u.account LIKE '2099%';

INSERT IGNORE INTO t_profile_score (student_id, term, dimension_key, label, score, avg_score, max_score, description, details)
SELECT s.id, '2024-2025-1', 'ability', '实践能力',
       CASE WHEN s.alert_level = 'none' THEN 72 WHEN s.alert_level = 'yellow' THEN 60 ELSE 50 END,
       68, 100,
       '课程项目与竞赛参与情况正常',
       JSON_ARRAY('参与课程项目 1 次', '建议参与学科竞赛')
FROM t_student s
JOIN t_user u ON u.id = s.user_id
WHERE u.account LIKE '2099%';

-- ============ 为预警学生生成预警记录 ============
INSERT INTO t_alert (student_id, level, type, title, description, course, failed_courses, trigger_date, status, suggestion, trigger_event, pushed_at)
SELECT s.id, s.alert_level,
       CASE s.alert_level WHEN 'red' THEN '学业危机' WHEN 'orange' THEN '挂科预警' ELSE '成绩预警' END,
       CASE s.alert_level WHEN 'red' THEN '多门课程不及格' WHEN 'orange' THEN 'GPA 低于 2.5' ELSE '成绩处于预警线' END,
       CONCAT(CASE s.alert_level WHEN 'yellow' THEN '当前成绩' ELSE 'GPA ' END, FORMAT(s.gpa, 2), '，低于专业平均水平，需要重点关注'),
       ELT(1 + MOD(seq, 6), '高等数学（下）', '线性代数', '数据结构', '操作系统', '大学英语（四）', '计算机网络'),
       CASE WHEN s.alert_level = 'yellow' THEN JSON_ARRAY() ELSE JSON_ARRAY(ELT(1 + MOD(seq, 6), '高等数学（下）', '线性代数', '数据结构', '操作系统', '大学英语（四）', '计算机网络')) END,
       DATE_ADD('2024-09-01', INTERVAL MOD(seq, 90) DAY),
       CASE s.alert_level WHEN 'red' THEN 'pending' WHEN 'orange' THEN 'processing' ELSE 'resolved' END,
       CASE s.alert_level WHEN 'red' THEN '建议立即约谈并制定学业恢复计划' WHEN 'orange' THEN '建议参加辅导并定期复盘' ELSE '建议持续跟踪成绩变化' END,
       CASE s.alert_level WHEN 'red' THEN '累计不及格达3门' WHEN 'orange' THEN 'GPA低于2.5' ELSE '成绩接近预警线' END,
       DATE_FORMAT(DATE_ADD('2024-09-01', INTERVAL MOD(seq, 90) DAY), '%Y-%m-%d %H:%i')
FROM t_student s
JOIN t_user u ON u.id = s.user_id
JOIN (SELECT account, CAST(SUBSTRING(account, 5) AS UNSIGNED) AS seq FROM t_user WHERE account LIKE '2099%') ns ON ns.account = u.account
WHERE u.account LIKE '2099%' AND s.alert_level <> 'none'
  AND NOT EXISTS (
      SELECT 1 FROM t_alert a WHERE a.student_id = s.id
      AND a.title = CASE s.alert_level WHEN 'red' THEN '多门课程不及格' WHEN 'orange' THEN 'GPA 低于 2.5' ELSE '成绩处于预警线' END
      AND a.trigger_date = DATE_ADD('2024-09-01', INTERVAL MOD(seq, 90) DAY)
  );

