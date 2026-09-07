-- V19: 补全大一至大四四届学生
-- 现有 320 名学生为 2021 届（班级名修正为 21XX），新增 2022/2023/2024 三届，
-- 每个专业每届 2 个班、每班 40 人，总计 1280 名学生。

-- 1) 将现有 2021 届班级名从 24XX 修正为 21XX（与届别一致）
UPDATE t_class SET name = '应统2101' WHERE id = 1;
UPDATE t_class SET name = '应统2102' WHERE id = 2;
UPDATE t_class SET name = '数科2101' WHERE id = 3;
UPDATE t_class SET name = '数科2102' WHERE id = 4;
UPDATE t_class SET name = '智能2101' WHERE id = 5;
UPDATE t_class SET name = '智能2102' WHERE id = 6;
UPDATE t_class SET name = '软工2101' WHERE id = 7;
UPDATE t_class SET name = '软工2102' WHERE id = 8;

UPDATE t_course_class SET class_name = '应统2101' WHERE class_name = '应统2401';
UPDATE t_course_class SET class_name = '应统2102' WHERE class_name = '应统2402';
UPDATE t_course_class SET class_name = '数科2101' WHERE class_name = '数科2401';
UPDATE t_course_class SET class_name = '数科2102' WHERE class_name = '数科2402';
UPDATE t_course_class SET class_name = '智能2101' WHERE class_name = '智能2401';
UPDATE t_course_class SET class_name = '智能2102' WHERE class_name = '智能2402';
UPDATE t_course_class SET class_name = '软工2101' WHERE class_name = '软工2401';
UPDATE t_course_class SET class_name = '软工2102' WHERE class_name = '软工2402';

UPDATE t_user SET class_name = '应统2101' WHERE class_name = '应统2401' AND role = 'student';
UPDATE t_user SET class_name = '应统2102' WHERE class_name = '应统2402' AND role = 'student';
UPDATE t_user SET class_name = '数科2101' WHERE class_name = '数科2401' AND role = 'student';
UPDATE t_user SET class_name = '数科2102' WHERE class_name = '数科2402' AND role = 'student';
UPDATE t_user SET class_name = '智能2101' WHERE class_name = '智能2401' AND role = 'student';
UPDATE t_user SET class_name = '智能2102' WHERE class_name = '智能2402' AND role = 'student';
UPDATE t_user SET class_name = '软工2101' WHERE class_name = '软工2401' AND role = 'student';
UPDATE t_user SET class_name = '软工2102' WHERE class_name = '软工2402' AND role = 'student';

-- 2) 补培养方案（各专业 2021-2024 届）
INSERT INTO t_major_plan (major_id, grade, dimensions, standard, actual, tips)
SELECT m.id, v.grade, v.dimensions, v.standard, v.actual, v.tips
FROM t_major m
CROSS JOIN (
    SELECT '2021' grade, '["公共基础","专业必修","专业选修","实践环节","通识教育"]' dimensions,
           '[48,62,20,18,12]' standard, '[46,58,24,16,12]' actual,
           '["专业必修完成度偏低","建议加强核心课程辅导"]' tips
    UNION ALL SELECT '2022', '["公共基础","专业必修","专业选修","实践环节","通识教育"]', '[48,62,20,18,12]', '[42,54,20,15,12]', '["大三阶段整体完成良好","建议关注考研与实习规划"]'
    UNION ALL SELECT '2023', '["公共基础","专业必修","专业选修","实践环节","通识教育"]', '[48,62,20,18,12]', '[30,38,14,10,12]', '["大二基础课程完成度正常","建议加强专业选修方向选择"]'
    UNION ALL SELECT '2024', '["公共基础","专业必修","专业选修","实践环节","通识教育"]', '[48,62,20,18,12]', '[20,28,10,8,10]', '["大一处于基础学习阶段","建议提前规划专业方向"]'
) v
WHERE NOT EXISTS (
    SELECT 1 FROM t_major_plan p WHERE p.major_id = m.id AND p.grade = v.grade
);

-- 3) 新增班主任与任课教师账号
INSERT INTO t_user (id, account, name, password, role, college, major, class_name, grade) VALUES
(1310, 'T2022A01', '应统2201班主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'teacher', '数学与计算机科学学院', '应用统计', NULL, NULL),
(1311, 'T2022A02', '应统2202班主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'teacher', '数学与计算机科学学院', '应用统计', NULL, NULL),
(1312, 'T2022B01', '数科2201班主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'teacher', '数学与计算机科学学院', '数据科学', NULL, NULL),
(1313, 'T2022B02', '数科2202班主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'teacher', '数学与计算机科学学院', '数据科学', NULL, NULL),
(1314, 'T2022C01', '智能2201班主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'teacher', '数学与计算机科学学院', '人工智能', NULL, NULL),
(1315, 'T2022C02', '智能2202班主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'teacher', '数学与计算机科学学院', '人工智能', NULL, NULL),
(1316, 'T2022D01', '软工2201班主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'teacher', '数学与计算机科学学院', '软件工程', NULL, NULL),
(1317, 'T2022D02', '软工2202班主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'teacher', '数学与计算机科学学院', '软件工程', NULL, NULL),
(1318, 'T2023A01', '应统2301班主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'teacher', '数学与计算机科学学院', '应用统计', NULL, NULL),
(1319, 'T2023A02', '应统2302班主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'teacher', '数学与计算机科学学院', '应用统计', NULL, NULL),
(1320, 'T2023B01', '数科2301班主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'teacher', '数学与计算机科学学院', '数据科学', NULL, NULL),
(1321, 'T2023B02', '数科2302班主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'teacher', '数学与计算机科学学院', '数据科学', NULL, NULL),
(1322, 'T2023C01', '智能2301班主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'teacher', '数学与计算机科学学院', '人工智能', NULL, NULL),
(1323, 'T2023C02', '智能2302班主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'teacher', '数学与计算机科学学院', '人工智能', NULL, NULL),
(1324, 'T2023D01', '软工2301班主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'teacher', '数学与计算机科学学院', '软件工程', NULL, NULL),
(1325, 'T2023D02', '软工2302班主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'teacher', '数学与计算机科学学院', '软件工程', NULL, NULL),
(1326, 'T2024A01', '应统2401班主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'teacher', '数学与计算机科学学院', '应用统计', NULL, NULL),
(1327, 'T2024A02', '应统2402班主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'teacher', '数学与计算机科学学院', '应用统计', NULL, NULL),
(1328, 'T2024B01', '数科2401班主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'teacher', '数学与计算机科学学院', '数据科学', NULL, NULL),
(1329, 'T2024B02', '数科2402班主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'teacher', '数学与计算机科学学院', '数据科学', NULL, NULL),
(1330, 'T2024C01', '智能2401班主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'teacher', '数学与计算机科学学院', '人工智能', NULL, NULL),
(1331, 'T2024C02', '智能2402班主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'teacher', '数学与计算机科学学院', '人工智能', NULL, NULL),
(1332, 'T2024D01', '软工2401班主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'teacher', '数学与计算机科学学院', '软件工程', NULL, NULL),
(1333, 'T2024D02', '软工2402班主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'teacher', '数学与计算机科学学院', '软件工程', NULL, NULL),
(1334, 'C2022A1', '统计22级任课一', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'course_teacher', '数学与计算机科学学院', '应用统计', NULL, NULL),
(1335, 'C2022A2', '统计22级任课二', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'course_teacher', '数学与计算机科学学院', '应用统计', NULL, NULL),
(1336, 'C2022B1', '数科22级任课一', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'course_teacher', '数学与计算机科学学院', '数据科学', NULL, NULL),
(1337, 'C2022B2', '数科22级任课二', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'course_teacher', '数学与计算机科学学院', '数据科学', NULL, NULL),
(1338, 'C2022C1', '智能22级任课一', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'course_teacher', '数学与计算机科学学院', '人工智能', NULL, NULL),
(1339, 'C2022C2', '智能22级任课二', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'course_teacher', '数学与计算机科学学院', '人工智能', NULL, NULL),
(1340, 'C2022D1', '软工22级任课一', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'course_teacher', '数学与计算机科学学院', '软件工程', NULL, NULL),
(1341, 'C2022D2', '软工22级任课二', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'course_teacher', '数学与计算机科学学院', '软件工程', NULL, NULL),
(1342, 'C2023A1', '统计23级任课一', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'course_teacher', '数学与计算机科学学院', '应用统计', NULL, NULL),
(1343, 'C2023A2', '统计23级任课二', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'course_teacher', '数学与计算机科学学院', '应用统计', NULL, NULL),
(1344, 'C2023B1', '数科23级任课一', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'course_teacher', '数学与计算机科学学院', '数据科学', NULL, NULL),
(1345, 'C2023B2', '数科23级任课二', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'course_teacher', '数学与计算机科学学院', '数据科学', NULL, NULL),
(1346, 'C2023C1', '智能23级任课一', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'course_teacher', '数学与计算机科学学院', '人工智能', NULL, NULL),
(1347, 'C2023C2', '智能23级任课二', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'course_teacher', '数学与计算机科学学院', '人工智能', NULL, NULL),
(1348, 'C2023D1', '软工23级任课一', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'course_teacher', '数学与计算机科学学院', '软件工程', NULL, NULL),
(1349, 'C2023D2', '软工23级任课二', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'course_teacher', '数学与计算机科学学院', '软件工程', NULL, NULL),
(1350, 'C2024A1', '统计24级任课一', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'course_teacher', '数学与计算机科学学院', '应用统计', NULL, NULL),
(1351, 'C2024A2', '统计24级任课二', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'course_teacher', '数学与计算机科学学院', '应用统计', NULL, NULL),
(1352, 'C2024B1', '数科24级任课一', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'course_teacher', '数学与计算机科学学院', '数据科学', NULL, NULL),
(1353, 'C2024B2', '数科24级任课二', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'course_teacher', '数学与计算机科学学院', '数据科学', NULL, NULL),
(1354, 'C2024C1', '智能24级任课一', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'course_teacher', '数学与计算机科学学院', '人工智能', NULL, NULL),
(1355, 'C2024C2', '智能24级任课二', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'course_teacher', '数学与计算机科学学院', '人工智能', NULL, NULL),
(1356, 'C2024D1', '软工24级任课一', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'course_teacher', '数学与计算机科学学院', '软件工程', NULL, NULL),
(1357, 'C2024D2', '软工24级任课二', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'course_teacher', '数学与计算机科学学院', '软件工程', NULL, NULL);

-- 4) 新增 2022/2023/2024 届班级
INSERT INTO t_class (id, name, grade, major_id, advisor_id, total_students) VALUES
(9,  '应统2201', '2022', 1, 25, 40),
(10, '应统2202', '2022', 1, 26, 40),
(11, '数科2201', '2022', 2, 27, 40),
(12, '数科2202', '2022', 2, 28, 40),
(13, '智能2201', '2022', 3, 29, 40),
(14, '智能2202', '2022', 3, 30, 40),
(15, '软工2201', '2022', 4, 31, 40),
(16, '软工2202', '2022', 4, 32, 40),
(17, '应统2301', '2023', 1, 33, 40),
(18, '应统2302', '2023', 1, 34, 40),
(19, '数科2301', '2023', 2, 35, 40),
(20, '数科2302', '2023', 2, 36, 40),
(21, '智能2301', '2023', 3, 37, 40),
(22, '智能2302', '2023', 3, 38, 40),
(23, '软工2301', '2023', 4, 39, 40),
(24, '软工2302', '2023', 4, 40, 40),
(25, '应统2401', '2024', 1, 41, 40),
(26, '应统2402', '2024', 1, 42, 40),
(27, '数科2401', '2024', 2, 43, 40),
(28, '数科2402', '2024', 2, 44, 40),
(29, '智能2401', '2024', 3, 45, 40),
(30, '智能2402', '2024', 3, 46, 40),
(31, '软工2401', '2024', 4, 47, 40),
(32, '软工2402', '2024', 4, 48, 40);

-- 5) 教师档案：24 名班主任 + 24 名任课教师
INSERT INTO t_teacher (id, teacher_id, user_id, department_id, title, is_class_advisor, class_id) VALUES
(25, 'T2022A01', 1310, 1, '讲师', 1, 9),
(26, 'T2022A02', 1311, 1, '讲师', 1, 10),
(27, 'T2022B01', 1312, 1, '讲师', 1, 11),
(28, 'T2022B02', 1313, 1, '讲师', 1, 12),
(29, 'T2022C01', 1314, 1, '讲师', 1, 13),
(30, 'T2022C02', 1315, 1, '讲师', 1, 14),
(31, 'T2022D01', 1316, 1, '讲师', 1, 15),
(32, 'T2022D02', 1317, 1, '讲师', 1, 16),
(33, 'T2023A01', 1318, 1, '讲师', 1, 17),
(34, 'T2023A02', 1319, 1, '讲师', 1, 18),
(35, 'T2023B01', 1320, 1, '讲师', 1, 19),
(36, 'T2023B02', 1321, 1, '讲师', 1, 20),
(37, 'T2023C01', 1322, 1, '讲师', 1, 21),
(38, 'T2023C02', 1323, 1, '讲师', 1, 22),
(39, 'T2023D01', 1324, 1, '讲师', 1, 23),
(40, 'T2023D02', 1325, 1, '讲师', 1, 24),
(41, 'T2024A01', 1326, 1, '讲师', 1, 25),
(42, 'T2024A02', 1327, 1, '讲师', 1, 26),
(43, 'T2024B01', 1328, 1, '讲师', 1, 27),
(44, 'T2024B02', 1329, 1, '讲师', 1, 28),
(45, 'T2024C01', 1330, 1, '讲师', 1, 29),
(46, 'T2024C02', 1331, 1, '讲师', 1, 30),
(47, 'T2024D01', 1332, 1, '讲师', 1, 31),
(48, 'T2024D02', 1333, 1, '讲师', 1, 32),
(49, 'C2022A1', 1334, 1, '讲师', 0, NULL),
(50, 'C2022A2', 1335, 1, '讲师', 0, NULL),
(51, 'C2022B1', 1336, 1, '讲师', 0, NULL),
(52, 'C2022B2', 1337, 1, '讲师', 0, NULL),
(53, 'C2022C1', 1338, 1, '讲师', 0, NULL),
(54, 'C2022C2', 1339, 1, '讲师', 0, NULL),
(55, 'C2022D1', 1340, 1, '讲师', 0, NULL),
(56, 'C2022D2', 1341, 1, '讲师', 0, NULL),
(57, 'C2023A1', 1342, 1, '讲师', 0, NULL),
(58, 'C2023A2', 1343, 1, '讲师', 0, NULL),
(59, 'C2023B1', 1344, 1, '讲师', 0, NULL),
(60, 'C2023B2', 1345, 1, '讲师', 0, NULL),
(61, 'C2023C1', 1346, 1, '讲师', 0, NULL),
(62, 'C2023C2', 1347, 1, '讲师', 0, NULL),
(63, 'C2023D1', 1348, 1, '讲师', 0, NULL),
(64, 'C2023D2', 1349, 1, '讲师', 0, NULL),
(65, 'C2024A1', 1350, 1, '讲师', 0, NULL),
(66, 'C2024A2', 1351, 1, '讲师', 0, NULL),
(67, 'C2024B1', 1352, 1, '讲师', 0, NULL),
(68, 'C2024B2', 1353, 1, '讲师', 0, NULL),
(69, 'C2024C1', 1354, 1, '讲师', 0, NULL),
(70, 'C2024C2', 1355, 1, '讲师', 0, NULL),
(71, 'C2024D1', 1356, 1, '讲师', 0, NULL),
(72, 'C2024D2', 1357, 1, '讲师', 0, NULL);

-- 6) 新增 960 名学生账号（2022/2023/2024 届）
INSERT INTO t_user (id, account, name, password, role, college, major, class_name, grade)
WITH RECURSIVE seq AS (SELECT 1 n UNION ALL SELECT n + 1 FROM seq WHERE n < 960),
cls AS (
    SELECT 9 id, '应统2201' name, 1 major_id, '2022' grade UNION ALL
    SELECT 10, '应统2202', 1, '2022' UNION ALL
    SELECT 11, '数科2201', 2, '2022' UNION ALL
    SELECT 12, '数科2202', 2, '2022' UNION ALL
    SELECT 13, '智能2201', 3, '2022' UNION ALL
    SELECT 14, '智能2202', 3, '2022' UNION ALL
    SELECT 15, '软工2201', 4, '2022' UNION ALL
    SELECT 16, '软工2202', 4, '2022' UNION ALL
    SELECT 17, '应统2301', 1, '2023' UNION ALL
    SELECT 18, '应统2302', 1, '2023' UNION ALL
    SELECT 19, '数科2301', 2, '2023' UNION ALL
    SELECT 20, '数科2302', 2, '2023' UNION ALL
    SELECT 21, '智能2301', 3, '2023' UNION ALL
    SELECT 22, '智能2302', 3, '2023' UNION ALL
    SELECT 23, '软工2301', 4, '2023' UNION ALL
    SELECT 24, '软工2302', 4, '2023' UNION ALL
    SELECT 25, '应统2401', 1, '2024' UNION ALL
    SELECT 26, '应统2402', 1, '2024' UNION ALL
    SELECT 27, '数科2401', 2, '2024' UNION ALL
    SELECT 28, '数科2402', 2, '2024' UNION ALL
    SELECT 29, '智能2401', 3, '2024' UNION ALL
    SELECT 30, '智能2402', 3, '2024' UNION ALL
    SELECT 31, '软工2401', 4, '2024' UNION ALL
    SELECT 32, '软工2402', 4, '2024'
)
SELECT 349 + s.n,
       CONCAT(cl.grade, cl.major_id, LPAD(MOD(s.n - 1, 80) + 1, 3, '0')),
       CONCAT('学生', LPAD(s.n, 3, '0')),
       '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO',
       'student', '数学与计算机科学学院',
       CASE cl.major_id WHEN 1 THEN '应用统计' WHEN 2 THEN '数据科学' WHEN 3 THEN '人工智能' ELSE '软件工程' END,
       cl.name, cl.grade
FROM seq s
JOIN cls cl ON cl.id = 9 + FLOOR((s.n - 1) / 40);

-- 7) 新增学生档案
INSERT INTO t_student (student_id, user_id, class_id, major_id, department_id, gpa, `rank`, total_students, alert_level, total_credits, required_credits)
SELECT u.account, u.id, c.id, c.major_id, 1,
       ROUND(2.0 + MOD(u.id * 7, 200) / 100, 2),
       MOD(u.id * 3, 40) + 1, 40, 'none',
       40 + MOD(u.id * 5, 60), 160.0
FROM t_user u
JOIN t_class c ON c.name = u.class_name
WHERE u.role = 'student' AND u.id BETWEEN 350 AND 1309;

-- 8) 新增开课记录（2022 届 5 学期、2023 届 3 学期、2024 届 1 学期）
INSERT INTO t_course_class (course_id, term, class_name, teacher_id, max_students, enrolled_students, status)
SELECT c.id, t.term, cl.name,
       CASE cl.grade
           WHEN '2022' THEN 49 + (cl.major_id - 1) * 2 + IF(MOD(c.id - 1, 6) IN (0, 1, 3), 0, 1)
           WHEN '2023' THEN 57 + (cl.major_id - 1) * 2 + IF(MOD(c.id - 1, 6) IN (0, 1, 3), 0, 1)
           ELSE 65 + (cl.major_id - 1) * 2 + IF(MOD(c.id - 1, 6) IN (0, 1, 3), 0, 1)
       END,
       40, 40,
       CASE WHEN t.term = '2024-2025-1' THEN 'active' ELSE 'finished' END
FROM t_course c
JOIN t_class cl ON cl.id BETWEEN 9 AND 32
JOIN (
    SELECT '2022' grade, '2022-2023-1' term UNION ALL
    SELECT '2022', '2022-2023-2' UNION ALL
    SELECT '2022', '2023-2024-1' UNION ALL
    SELECT '2022', '2023-2024-2' UNION ALL
    SELECT '2022', '2024-2025-1' UNION ALL
    SELECT '2023', '2023-2024-1' UNION ALL
    SELECT '2023', '2023-2024-2' UNION ALL
    SELECT '2023', '2024-2025-1' UNION ALL
    SELECT '2024', '2024-2025-1'
) t ON t.grade = cl.grade
WHERE c.id BETWEEN (cl.major_id - 1) * 6 + 1 AND cl.major_id * 6;

-- 9) 新增原始成绩（按各届已修学期）
INSERT INTO t_grade (student_id, course_class_id, score, status, term, exam_date)
SELECT s.id, cc.id,
       CASE WHEN MOD(s.id + cc.course_id, 11) = 0 THEN 45 + MOD(s.id + cc.id, 12)
            ELSE 70 + MOD(s.id * 13 + cc.id * 7, 26)
       END,
       'passed', cc.term, '2025-01-15'
FROM t_student s
JOIN t_class cl ON cl.id = s.class_id
JOIN t_course_class cc ON cc.class_name = cl.name
JOIN t_user u ON u.id = s.user_id
WHERE u.id BETWEEN 350 AND 1309;

-- 10) 当前学期保留部分挂科
UPDATE t_grade g
JOIN t_course_class cc ON cc.id = g.course_class_id
JOIN t_student s ON s.id = g.student_id
JOIN t_class cl ON cl.id = s.class_id
JOIN t_user u ON u.id = s.user_id
SET g.score = 42 + MOD(g.id, 15)
WHERE u.id BETWEEN 350 AND 1309
  AND cc.term = '2024-2025-1'
  AND MOD(s.id, 8) IN (0, 3)
  AND cc.course_id IN (
      (cl.major_id - 1) * 6 + 1 + MOD(s.id, 6),
      (cl.major_id - 1) * 6 + 1 + MOD(s.id + 3, 6)
  );

-- 11) 统一重算状态与绩点
UPDATE t_grade g
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
WHERE u.id BETWEEN 350 AND 1309;

-- 12) 体测、考勤、志愿、心理、竞赛（当前学期）
INSERT INTO t_health_report (student_id, term, total_score, items, report_url)
SELECT s.id, '2024-2025-1', 60 + MOD(s.id * 5, 36),
       JSON_ARRAY(
           JSON_OBJECT('name', '体重指数', 'score', 60 + MOD(s.id * 3, 36), 'level', '良好'),
           JSON_OBJECT('name', '肺活量', 'score', 60 + MOD(s.id * 5, 36), 'level', '良好'),
           JSON_OBJECT('name', '耐力跑', 'score', 60 + MOD(s.id * 7, 36), 'level', '良好')
       ), NULL
FROM t_student s
JOIN t_user u ON u.id = s.user_id
WHERE u.id BETWEEN 350 AND 1309;

INSERT INTO t_attendance (student_id, term, total_classes, absent_count, late_count, note)
SELECT s.id, '2024-2025-1', 60,
       CASE
           WHEN MOD(s.id, 6) = 0 THEN 6 + MOD(s.id, 4)
           WHEN MOD(s.id, 13) = 0 THEN 9 + MOD(s.id, 3)
           ELSE MOD(s.id, 5)
       END,
       MOD(s.id, 5), '模拟考勤数据'
FROM t_student s
JOIN t_user u ON u.id = s.user_id
WHERE u.id BETWEEN 350 AND 1309;

INSERT INTO t_volunteer (student_id, term, hours, description)
SELECT s.id, '2024-2025-1', 6 + MOD(s.id * 7, 60), '模拟志愿服务记录'
FROM t_student s
JOIN t_user u ON u.id = s.user_id
WHERE u.id BETWEEN 350 AND 1309;

INSERT INTO t_psychology (student_id, term, score, level, note)
SELECT s.id, '2024-2025-1',
       62 + MOD(s.id * 5, 36),
       CASE WHEN 62 + MOD(s.id * 5, 36) >= 85 THEN 'good'
            WHEN 62 + MOD(s.id * 5, 36) >= 70 THEN 'normal'
            ELSE 'attention' END,
       '模拟心理测评记录'
FROM t_student s
JOIN t_user u ON u.id = s.user_id
WHERE u.id BETWEEN 350 AND 1309;

INSERT INTO t_competition (student_id, term, competition_name, level, award, points)
SELECT s.id, '2024-2025-1',
       CASE cl.major_id WHEN 1 THEN '应用统计建模大赛'
                        WHEN 2 THEN '数据挖掘竞赛'
                        WHEN 3 THEN '人工智能创新赛'
                        ELSE '软件设计大赛' END,
       CASE WHEN MOD(s.id, 6) = 0 THEN 'national' WHEN MOD(s.id, 6) = 1 THEN 'provincial' ELSE 'school' END,
       CASE WHEN MOD(s.id, 4) = 0 THEN 'first' WHEN MOD(s.id, 4) = 1 THEN 'second' WHEN MOD(s.id, 4) = 2 THEN 'third' ELSE 'participation' END,
       CASE WHEN MOD(s.id, 6) = 0 AND MOD(s.id, 4) = 0 THEN 92
            WHEN MOD(s.id, 6) = 0 THEN 85
            WHEN MOD(s.id, 6) = 1 AND MOD(s.id, 4) = 0 THEN 82
            WHEN MOD(s.id, 6) = 1 THEN 76
            WHEN MOD(s.id, 4) = 0 THEN 74
            WHEN MOD(s.id, 4) = 1 THEN 70
            ELSE 64 END
FROM t_student s
JOIN t_class cl ON cl.id = s.class_id
JOIN t_user u ON u.id = s.user_id
WHERE u.id BETWEEN 350 AND 1309 AND MOD(s.id, 4) = 0;
