from pathlib import Path

HASH = '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO'
OUT = Path(r'E:\Academic Navigation-test\Academic Navigation - 副本\academic-navigation-backend\src\main\resources\db\migration\V6__diversify_dataset.sql')

MAJORS = [
    (0, 1, 'CS', 1), (1, 2, 'SE', 1), (2, 3, 'DS', 1), (3, 5, 'AI1', 1), (4, 6, 'NSS', 1),
    (5, 100200, 'BDA', 100100), (6, 100202, 'CS2', 100100), (7, 100204, 'AI2', 100100),
    (8, 100205, 'STAT', 100100), (9, 100206, 'IMIS', 100100),
    (10, 100201, 'SE2', 100101), (11, 100203, 'AI3', 100101), (12, 100207, 'IOT', 100101),
    (13, 100208, 'SEC', 100101), (14, 100209, 'DMT', 100101),
]

NEW_MAJORS = [
    (5, '人工智能', 'AI1', 1),
    (6, '网络空间安全', 'NSS', 1),
    (100204, '人工智能', 'AI2', 100100),
    (100205, '统计学', 'STAT', 100100),
    (100206, '信息管理与信息系统', 'IMIS', 100100),
    (100207, '物联网工程', 'IOT', 100101),
    (100208, '信息安全', 'SEC', 100101),
    (100209, '数字媒体技术', 'DMT', 100101),
]

NEW_DEPT_ACCOUNTS = [
    ('D20210004', '人工智能系主任', '数学与计算机科学学院', '人工智能'),
    ('D20210005', '网络空间安全系主任', '数学与计算机科学学院', '网络空间安全'),
    ('D2024005', '人工智能系主任', '数据科学与工程测试学院', '人工智能'),
    ('D2024006', '统计学系主任', '数据科学与工程测试学院', '统计学'),
    ('D2024007', '信息管理系主任', '数据科学与工程测试学院', '信息管理与信息系统'),
    ('D2024008', '物联网系主任', '智能计算测试学院', '物联网工程'),
    ('D2024009', '信息安全系主任', '智能计算测试学院', '信息安全'),
    ('D2024010', '数字媒体系主任', '智能计算测试学院', '数字媒体技术'),
]

SURNAMES = ['张', '李', '王', '刘', '陈', '杨', '赵', '黄', '周', '吴', '徐', '孙', '胡', '朱', '高', '林', '何', '郭', '马', '罗']
GIVEN = ['伟', '芳', '娜', '敏', '静', '磊', '军', '洋', '勇', '艳', '杰', '娟', '涛', '明', '超', '霞', '平', '刚', '桂英', '志强']

COURSES = [1, 2, 3, 4, 5, 6]
COURSE_NAMES = ['高等数学（下）', '线性代数', '数据结构', '操作系统', '大学英语（四）', '计算机网络']

lines = []
w = lines.append

w('-- V6: 多样化数据集')
w('-- 目标：学生总数 1000（现有 143 + 新增 857），覆盖 2021-2024 四个年级，')
w('-- 三个学院，每个学院 4-5 个专业；同步补齐新专业对应的系主任账号。')
w('SET NAMES utf8mb4;')
w('')

# ---------- 自包含学院与样本专业 ----------
w('-- ============ 确保三学院结构与样本专业存在（干净库也可直接执行） ============')
w("INSERT INTO t_department (id, name, code, dean_id)")
w("SELECT 100100, '数据科学与工程测试学院', 'TEST_01', NULL")
w("WHERE NOT EXISTS (SELECT 1 FROM t_department WHERE name = '数据科学与工程测试学院');")
w("INSERT INTO t_department (id, name, code, dean_id)")
w("SELECT 100101, '智能计算测试学院', 'TEST_02', NULL")
w("WHERE NOT EXISTS (SELECT 1 FROM t_department WHERE name = '智能计算测试学院');")
w("INSERT INTO t_major (id, name, code, department_id, degree_type)")
w("SELECT 100200, '数据科学与大数据技术', 'MAJOR_01', 100100, '工学学士'")
w("WHERE NOT EXISTS (SELECT 1 FROM t_major WHERE name = '数据科学与大数据技术' AND department_id = 100100);")
w("INSERT INTO t_major (id, name, code, department_id, degree_type)")
w("SELECT 100201, '软件工程', 'MAJOR_02', 100101, '工学学士'")
w("WHERE NOT EXISTS (SELECT 1 FROM t_major WHERE name = '软件工程' AND department_id = 100101);")
w("INSERT INTO t_major (id, name, code, department_id, degree_type)")
w("SELECT 100202, '计算机科学与技术', 'MAJOR_03', 100100, '工学学士'")
w("WHERE NOT EXISTS (SELECT 1 FROM t_major WHERE name = '计算机科学与技术' AND department_id = 100100);")
w("INSERT INTO t_major (id, name, code, department_id, degree_type)")
w("SELECT 100203, '人工智能', 'MAJOR_04', 100101, '工学学士'")
w("WHERE NOT EXISTS (SELECT 1 FROM t_major WHERE name = '人工智能' AND department_id = 100101);")
w("UPDATE t_department SET dean_id = (SELECT id FROM (SELECT id FROM t_user WHERE account = 'L2024001') t) WHERE id = 100100 AND dean_id IS NULL;")
w("UPDATE t_department SET dean_id = (SELECT id FROM (SELECT id FROM t_user WHERE account = 'L2024002') t) WHERE id = 100101 AND dean_id IS NULL;")
w('')

digits = ',\n      '.join(f'SELECT {i} n' for i in range(10))
nums = (
    '(\n'
    '  SELECT d0.n + d1.n * 10 + d2.n * 100 + 1 AS seq\n'
    '  FROM (SELECT 0 n UNION ALL SELECT 1 UNION ALL SELECT 2 UNION ALL SELECT 3 UNION ALL SELECT 4\n'
    '        UNION ALL SELECT 5 UNION ALL SELECT 6 UNION ALL SELECT 7 UNION ALL SELECT 8 UNION ALL SELECT 9) d0\n'
    '  CROSS JOIN (SELECT 0 n UNION ALL SELECT 1 UNION ALL SELECT 2 UNION ALL SELECT 3 UNION ALL SELECT 4\n'
    '              UNION ALL SELECT 5 UNION ALL SELECT 6 UNION ALL SELECT 7 UNION ALL SELECT 8 UNION ALL SELECT 9) d1\n'
    '  CROSS JOIN (SELECT 0 n UNION ALL SELECT 1 UNION ALL SELECT 2 UNION ALL SELECT 3 UNION ALL SELECT 4\n'
    '              UNION ALL SELECT 5 UNION ALL SELECT 6 UNION ALL SELECT 7 UNION ALL SELECT 8 UNION ALL SELECT 9) d2\n'
    ') nums'
)

mapping_rows = '\nUNION ALL\n'.join(
    f"    SELECT {idx} idx, {mid} major_id, '{short}' short" for idx, mid, short, _ in MAJORS
)
mapping = f'(\n{mapping_rows}\n) ma'

# ---------- 新专业 ----------
w('-- ============ 新增专业（每学院 4-5 个） ============')
for mid, name, code, dept in NEW_MAJORS:
    w(f"INSERT INTO t_major (id, name, code, department_id, degree_type)")
    w(f"SELECT {mid}, '{name}', '{code}', {dept}, '工学学士'")
    w(f"WHERE NOT EXISTS (SELECT 1 FROM t_major WHERE name = '{name}' AND department_id = {dept});")
w('')

# ---------- 新系主任账号 ----------
w('-- ============ 新增专业对应的系主任账号（密码 123456） ============')
rows = ',\n'.join(
    f"    ('{acc}', '{name}', '{HASH}', 'department', '{college}', '{major}', NULL, NULL)"
    for acc, name, college, major in NEW_DEPT_ACCOUNTS
)
w('INSERT IGNORE INTO t_user (account, name, password, role, college, major, class_name, grade) VALUES')
w(rows)
w(';')
w('')

# ---------- 班级 ----------
w('-- ============ 为 15 个专业 x 4 个年级创建班级 ============')
w('INSERT INTO t_class (name, grade, major_id, advisor_id, total_students)')
w(f"SELECT CONCAT(ma.short, g.grade, '01'), g.grade, ma.major_id, 1, 0")
w('FROM (SELECT 2021 grade UNION ALL SELECT 2022 UNION ALL SELECT 2023 UNION ALL SELECT 2024) g')
w(f'CROSS JOIN {mapping}')
w("WHERE NOT EXISTS (SELECT 1 FROM t_class c WHERE c.name = CONCAT(ma.short, g.grade, '01'));")
w('')

# ---------- 学生用户 ----------
w('-- ============ 新增 857 名学生账号（20990001-20990857） ============')
sur = ',\n'.join(f"'{s}'" for s in SURNAMES)
giv = ',\n'.join(f"'{g}'" for g in GIVEN)
w('INSERT IGNORE INTO t_user (account, name, password, role, college, major, class_name, grade)')
w('SELECT CONCAT(\'2099\', LPAD(nums.seq, 4, \'0\')),')
w(f"       CONCAT(ELT(1 + MOD(nums.seq - 1, 20), {sur}),")
w(f"              ELT(1 + MOD(nums.seq * 7, 20), {giv})),")
w(f"       '{HASH}', 'student', dep.name, m.name,")
w("       CONCAT(ma.short, 2021 + MOD(FLOOR((nums.seq - 1) / 15), 4), '01'),")
w('       2021 + MOD(FLOOR((nums.seq - 1) / 15), 4)')
w(f'FROM {nums}')
w(f'JOIN {mapping} ON ma.idx = MOD(nums.seq - 1, 15)')
w('JOIN t_major m ON m.id = ma.major_id')
w('JOIN t_department dep ON dep.id = m.department_id')
w('JOIN (SELECT 1000 - (SELECT COUNT(*) FROM t_student) AS target) tg ON 1 = 1')
w('WHERE nums.seq <= tg.target;')
w('')

# ---------- 学生档案 ----------
w('-- ============ 新增学生档案（GPA/排名/预警等级按序号确定性生成） ============')
w('INSERT INTO t_student (student_id, user_id, class_id, major_id, department_id, gpa, `rank`, total_students, alert_level, total_credits, required_credits)')
w('SELECT u.account, u.id, cl.id, m.id, m.department_id,')
w('       ROUND(2.10 + MOD(seq * 7, 190) / 100.0, 2),')
w('       1 + MOD(seq * 13, 69),')
w('       70,')
w("       CASE WHEN MOD(seq, 17) = 0 THEN 'red' WHEN MOD(seq, 11) = 0 THEN 'orange' WHEN MOD(seq, 7) = 0 THEN 'yellow' ELSE 'none' END,")
w('       (CAST(u.grade AS UNSIGNED) - 2021) * 40 + 20,')
w('       176.0')
w("FROM t_user u")
w("JOIN (SELECT account, CAST(SUBSTRING(account, 5) AS UNSIGNED) AS seq FROM t_user WHERE account LIKE '2099%') ns ON ns.account = u.account")
w("JOIN t_department dep ON dep.name = u.college")
w("JOIN t_major m ON m.name = u.major AND m.department_id = dep.id")
w(f"JOIN {mapping} ON ma.major_id = m.id")
w("JOIN t_class cl ON cl.name = CONCAT(ma.short, u.grade, '01')")
w("WHERE u.account LIKE '2099%'")
w('  AND NOT EXISTS (SELECT 1 FROM t_student s WHERE s.student_id = u.account);')
w('')

# ---------- 班级人数回填 ----------
w('-- ============ 回填班级人数 ============')
w('UPDATE t_class c')
w('SET c.total_students = (SELECT COUNT(*) FROM t_student s JOIN t_user u ON u.id = s.user_id WHERE u.class_name = c.name);')
w('')

# ---------- 课程班级 ----------
w('-- ============ 为新班级创建当前学期课程班级 ============')
w("INSERT INTO t_course_class (course_id, term, class_name, teacher_id, max_students, enrolled_students, status)")
w("SELECT c.id, '2024-2025-1', cl.name, 2, cl.total_students, cl.total_students, 'active'")
w('FROM t_course c')
w('JOIN t_class cl ON cl.name REGEXP \'^(CS|SE|DS|AI1|NSS|BDA|CS2|AI2|STAT|IMIS|SE2|AI3|IOT|SEC|DMT)[0-9]{4}01$\'')
w('WHERE c.id IN (1,2,3,4,5,6)')
w("  AND NOT EXISTS (SELECT 1 FROM t_course_class cc WHERE cc.course_id = c.id AND cc.class_name = cl.name AND cc.term = '2024-2025-1');")
w('')

# ---------- 成绩 ----------
w('-- ============ 生成当前学期成绩 ============')
w('INSERT IGNORE INTO t_grade (student_id, course_class_id, score, status, term, exam_date)')
w('SELECT s.id, cc.id,')
w('       45 + MOD(seq * 7 + cc.course_id * 11, 56),')
w("       'passed', '2024-2025-1', '2025-01-15'")
w('FROM t_student s')
w('JOIN t_user u ON u.id = s.user_id')
w("JOIN (SELECT account, CAST(SUBSTRING(account, 5) AS UNSIGNED) AS seq FROM t_user WHERE account LIKE '2099%') ns ON ns.account = u.account")
w('JOIN t_class cl ON cl.id = s.class_id')
w("JOIN t_course_class cc ON cc.class_name = cl.name AND cc.term = '2024-2025-1' AND cc.course_id IN (1,2,3,4,5,6)")
w("WHERE u.account LIKE '2099%';")
w('')
w('UPDATE t_grade g')
w('JOIN t_student s ON s.id = g.student_id')
w('JOIN t_user u ON u.id = s.user_id')
w("SET g.status = IF(g.score >= 60, 'passed', 'failed'),")
w('    g.grade_point = CASE')
w('        WHEN g.score >= 95 THEN 4.0 WHEN g.score >= 90 THEN 3.7 WHEN g.score >= 85 THEN 3.3')
w('        WHEN g.score >= 80 THEN 3.0 WHEN g.score >= 75 THEN 2.7 WHEN g.score >= 70 THEN 2.3')
w('        WHEN g.score >= 65 THEN 2.0 WHEN g.score >= 60 THEN 1.7 ELSE 0.0 END')
w("WHERE u.account LIKE '2099%' AND g.term = '2024-2025-1';")
w('')

# ---------- GPA 历史 ----------
w('-- ============ 生成各年级 GPA 历史 ============')
w('INSERT IGNORE INTO t_gpa_history (student_id, term, gpa, avg_gpa, `rank`, total_students)')
w('SELECT s.id, t.term,')
w('       ROUND(LEAST(4.0, GREATEST(0.8, s.gpa + (t.idx - 7) * 0.04 + MOD(seq * 3, 5) * 0.01)), 2),')
w('       ROUND(3.0 + (t.idx - 7) * 0.03, 2),')
w('       GREATEST(1, s.`rank` + (7 - t.idx) * 2),')
w('       s.total_students')
w('FROM t_student s')
w('JOIN t_user u ON u.id = s.user_id')
w("JOIN (SELECT account, CAST(SUBSTRING(account, 5) AS UNSIGNED) AS seq FROM t_user WHERE account LIKE '2099%') ns ON ns.account = u.account")
w('JOIN (')
w("    SELECT '2021-2022-1' term, 1 idx UNION ALL SELECT '2021-2022-2', 2")
w("    UNION ALL SELECT '2022-2023-1', 3 UNION ALL SELECT '2022-2023-2', 4")
w("    UNION ALL SELECT '2023-2024-1', 5 UNION ALL SELECT '2023-2024-2', 6")
w("    UNION ALL SELECT '2024-2025-1', 7")
w(') t')
w("WHERE u.account LIKE '2099%'")
w('  AND CAST(LEFT(t.term, 4) AS SIGNED) - CAST(u.grade AS SIGNED) BETWEEN 0 AND 3;')
w('')

# ---------- 画像分数 ----------
w('-- ============ 生成当前学期画像分数 ============')
w('INSERT IGNORE INTO t_profile_score (student_id, term, dimension_key, label, score, avg_score, max_score, description, details)')
w("SELECT s.id, '2024-2025-1', 'academic', '学业成绩',")
w('       CASE WHEN s.gpa >= 3.5 THEN 88 WHEN s.gpa >= 3.0 THEN 76 WHEN s.gpa >= 2.5 THEN 64 WHEN s.gpa >= 2.0 THEN 54 ELSE 42 END,')
w('       72, 100,')
w("       CONCAT('GPA ', FORMAT(s.gpa, 2), '，专业排名 ', s.`rank`),")
w("       JSON_ARRAY(CONCAT('GPA: ', FORMAT(s.gpa, 2), ' / 4.0'))")
w('FROM t_student s')
w('JOIN t_user u ON u.id = s.user_id')
w("WHERE u.account LIKE '2099%';")
w('')
w('INSERT IGNORE INTO t_profile_score (student_id, term, dimension_key, label, score, avg_score, max_score, description, details)')
w("SELECT s.id, '2024-2025-1', 'ability', '实践能力',")
w("       CASE WHEN s.alert_level = 'none' THEN 72 WHEN s.alert_level = 'yellow' THEN 60 ELSE 50 END,")
w('       68, 100,')
w("       '课程项目与竞赛参与情况正常',")
w("       JSON_ARRAY('参与课程项目 1 次', '建议参与学科竞赛')")
w('FROM t_student s')
w('JOIN t_user u ON u.id = s.user_id')
w("WHERE u.account LIKE '2099%';")
w('')

# ---------- 预警 ----------
course_elt = 'ELT(1 + MOD(seq, 6), ' + ', '.join(f"'{c}'" for c in COURSE_NAMES) + ')'
w('-- ============ 为预警学生生成预警记录 ============')
w('INSERT INTO t_alert (student_id, level, type, title, description, course, failed_courses, trigger_date, status, suggestion, trigger_event, pushed_at)')
w('SELECT s.id, s.alert_level,')
w("       CASE s.alert_level WHEN 'red' THEN '学业危机' WHEN 'orange' THEN '挂科预警' ELSE '成绩预警' END,")
w("       CASE s.alert_level WHEN 'red' THEN '多门课程不及格' WHEN 'orange' THEN 'GPA 低于 2.5' ELSE '成绩处于预警线' END,")
w("       CONCAT(CASE s.alert_level WHEN 'yellow' THEN '当前成绩' ELSE 'GPA ' END, FORMAT(s.gpa, 2), '，低于专业平均水平，需要重点关注'),")
w(f'       {course_elt},')
w(f"       CASE WHEN s.alert_level = 'yellow' THEN JSON_ARRAY() ELSE JSON_ARRAY({course_elt}) END,")
w("       DATE_ADD('2024-09-01', INTERVAL MOD(seq, 90) DAY),")
w("       CASE s.alert_level WHEN 'red' THEN 'pending' WHEN 'orange' THEN 'processing' ELSE 'resolved' END,")
w("       CASE s.alert_level WHEN 'red' THEN '建议立即约谈并制定学业恢复计划' WHEN 'orange' THEN '建议参加辅导并定期复盘' ELSE '建议持续跟踪成绩变化' END,")
w("       CASE s.alert_level WHEN 'red' THEN '累计不及格达3门' WHEN 'orange' THEN 'GPA低于2.5' ELSE '成绩接近预警线' END,")
w("       DATE_FORMAT(DATE_ADD('2024-09-01', INTERVAL MOD(seq, 90) DAY), '%Y-%m-%d %H:%i')")
w('FROM t_student s')
w('JOIN t_user u ON u.id = s.user_id')
w("JOIN (SELECT account, CAST(SUBSTRING(account, 5) AS UNSIGNED) AS seq FROM t_user WHERE account LIKE '2099%') ns ON ns.account = u.account")
w("WHERE u.account LIKE '2099%' AND s.alert_level <> 'none'")
w('  AND NOT EXISTS (')
w('      SELECT 1 FROM t_alert a WHERE a.student_id = s.id')
w("      AND a.title = CASE s.alert_level WHEN 'red' THEN '多门课程不及格' WHEN 'orange' THEN 'GPA 低于 2.5' ELSE '成绩处于预警线' END")
w("      AND a.trigger_date = DATE_ADD('2024-09-01', INTERVAL MOD(seq, 90) DAY)")
w('  );')
w('')

OUT.write_text('\n'.join(lines) + '\n', encoding='utf-8')
print(f'generated {OUT}')
print(f'lines={len(lines)}')
