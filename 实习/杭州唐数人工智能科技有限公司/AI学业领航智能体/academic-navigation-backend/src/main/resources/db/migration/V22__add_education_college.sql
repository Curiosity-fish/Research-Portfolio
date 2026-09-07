-- V22: 新增教育学院，仿照数学与计算机科学学院结构
-- 4 个专业 × 4 届 × 2 个班 × 40 人，共 1280 名学生

-- 1) 学院
INSERT INTO t_department (id, name, code, dean_id) VALUES
(2, '教育学院', 'EDU', 1358);

-- 2) 专业
INSERT INTO t_major (id, name, code, department_id, degree_type) VALUES
(5, '教育学', 'EDU', 2, '教育学学士'),
(6, '学前教育', 'PRE', 2, '教育学学士'),
(7, '小学教育', 'PRI', 2, '教育学学士'),
(8, '教育技术学', 'EDT', 2, '理学学士');

-- 3) 课程
INSERT INTO t_course (id, code, name, credits, type, department_id) VALUES
(25, 'EDU201', '教育学原理', 3.5, 'required', 2),
(26, 'EDU202', '中外教育史', 3.0, 'required', 2),
(27, 'EDU203', '课程与教学论', 3.5, 'required', 2),
(28, 'EDU204', '教育心理学', 3.0, 'required', 2),
(29, 'EDU205', '教育研究方法', 2.5, 'elective', 2),
(30, 'EDU206', '教育政策与法规', 2.0, 'elective', 2),
(31, 'PRE201', '学前教育学', 3.5, 'required', 2),
(32, 'PRE202', '儿童发展心理学', 3.0, 'required', 2),
(33, 'PRE203', '学前课程论', 3.0, 'required', 2),
(34, 'PRE204', '幼儿园活动设计', 3.0, 'required', 2),
(35, 'PRE205', '学前教育评价', 2.5, 'elective', 2),
(36, 'PRE206', '儿童卫生与保健', 2.5, 'elective', 2),
(37, 'PRI201', '小学教育学', 3.5, 'required', 2),
(38, 'PRI202', '小学语文教学法', 3.0, 'required', 2),
(39, 'PRI203', '小学数学教学法', 3.0, 'required', 2),
(40, 'PRI204', '小学英语教学法', 3.0, 'required', 2),
(41, 'PRI205', '小学班级管理', 2.5, 'elective', 2),
(42, 'PRI206', '小学教育科研方法', 2.5, 'elective', 2),
(43, 'EDT201', '教育技术学导论', 3.5, 'required', 2),
(44, 'EDT202', '教学系统设计', 3.0, 'required', 2),
(45, 'EDT203', '多媒体课件制作', 3.0, 'required', 2),
(46, 'EDT204', '远程教育原理', 3.0, 'required', 2),
(47, 'EDT205', '教育数据挖掘', 2.5, 'elective', 2),
(48, 'EDT206', '智能学习环境', 2.5, 'elective', 2);

-- 4) 院长、系主任账号
INSERT INTO t_user (id, account, name, password, role, college, major, class_name, grade) VALUES
(1358, 'L20210002', '李院长', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'dean', '教育学院', NULL, NULL, NULL),
(1359, 'D20210005', '教育学系主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'department', '教育学院', '教育学', NULL, NULL),
(1360, 'D20210006', '学前教育系主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'department', '教育学院', '学前教育', NULL, NULL),
(1361, 'D20210007', '小学教育系主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'department', '教育学院', '小学教育', NULL, NULL),
(1362, 'D20210008', '教育技术系主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'department', '教育学院', '教育技术学', NULL, NULL);

-- 5) 班主任账号（32 人）
INSERT INTO t_user (id, account, name, password, role, college, major, class_name, grade) VALUES
(1363, 'T2021E01', '教育学2101班主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'teacher', '教育学院', '教育学', NULL, NULL),
(1364, 'T2021E02', '教育学2102班主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'teacher', '教育学院', '教育学', NULL, NULL),
(1365, 'T2021P01', '学前2101班主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'teacher', '教育学院', '学前教育', NULL, NULL),
(1366, 'T2021P02', '学前2102班主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'teacher', '教育学院', '学前教育', NULL, NULL),
(1367, 'T2021X01', '小教2101班主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'teacher', '教育学院', '小学教育', NULL, NULL),
(1368, 'T2021X02', '小教2102班主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'teacher', '教育学院', '小学教育', NULL, NULL),
(1369, 'T2021T01', '教技2101班主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'teacher', '教育学院', '教育技术学', NULL, NULL),
(1370, 'T2021T02', '教技2102班主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'teacher', '教育学院', '教育技术学', NULL, NULL),
(1371, 'T2022E01', '教育学2201班主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'teacher', '教育学院', '教育学', NULL, NULL),
(1372, 'T2022E02', '教育学2202班主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'teacher', '教育学院', '教育学', NULL, NULL),
(1373, 'T2022P01', '学前2201班主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'teacher', '教育学院', '学前教育', NULL, NULL),
(1374, 'T2022P02', '学前2202班主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'teacher', '教育学院', '学前教育', NULL, NULL),
(1375, 'T2022X01', '小教2201班主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'teacher', '教育学院', '小学教育', NULL, NULL),
(1376, 'T2022X02', '小教2202班主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'teacher', '教育学院', '小学教育', NULL, NULL),
(1377, 'T2022T01', '教技2201班主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'teacher', '教育学院', '教育技术学', NULL, NULL),
(1378, 'T2022T02', '教技2202班主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'teacher', '教育学院', '教育技术学', NULL, NULL),
(1379, 'T2023E01', '教育学2301班主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'teacher', '教育学院', '教育学', NULL, NULL),
(1380, 'T2023E02', '教育学2302班主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'teacher', '教育学院', '教育学', NULL, NULL),
(1381, 'T2023P01', '学前2301班主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'teacher', '教育学院', '学前教育', NULL, NULL),
(1382, 'T2023P02', '学前2302班主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'teacher', '教育学院', '学前教育', NULL, NULL),
(1383, 'T2023X01', '小教2301班主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'teacher', '教育学院', '小学教育', NULL, NULL),
(1384, 'T2023X02', '小教2302班主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'teacher', '教育学院', '小学教育', NULL, NULL),
(1385, 'T2023T01', '教技2301班主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'teacher', '教育学院', '教育技术学', NULL, NULL),
(1386, 'T2023T02', '教技2302班主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'teacher', '教育学院', '教育技术学', NULL, NULL),
(1387, 'T2024E01', '教育学2401班主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'teacher', '教育学院', '教育学', NULL, NULL),
(1388, 'T2024E02', '教育学2402班主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'teacher', '教育学院', '教育学', NULL, NULL),
(1389, 'T2024P01', '学前2401班主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'teacher', '教育学院', '学前教育', NULL, NULL),
(1390, 'T2024P02', '学前2402班主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'teacher', '教育学院', '学前教育', NULL, NULL),
(1391, 'T2024X01', '小教2401班主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'teacher', '教育学院', '小学教育', NULL, NULL),
(1392, 'T2024X02', '小教2402班主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'teacher', '教育学院', '小学教育', NULL, NULL),
(1393, 'T2024T01', '教技2401班主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'teacher', '教育学院', '教育技术学', NULL, NULL),
(1394, 'T2024T02', '教技2402班主任', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'teacher', '教育学院', '教育技术学', NULL, NULL);

-- 6) 任课教师账号（32 人）
INSERT INTO t_user (id, account, name, password, role, college, major, class_name, grade) VALUES
(1395, 'C2021E1', '教育学21级任课一', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'course_teacher', '教育学院', '教育学', NULL, NULL),
(1396, 'C2021E2', '教育学21级任课二', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'course_teacher', '教育学院', '教育学', NULL, NULL),
(1397, 'C2021P1', '学前21级任课一', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'course_teacher', '教育学院', '学前教育', NULL, NULL),
(1398, 'C2021P2', '学前21级任课二', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'course_teacher', '教育学院', '学前教育', NULL, NULL),
(1399, 'C2021X1', '小教21级任课一', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'course_teacher', '教育学院', '小学教育', NULL, NULL),
(1400, 'C2021X2', '小教21级任课二', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'course_teacher', '教育学院', '小学教育', NULL, NULL),
(1401, 'C2021T1', '教技21级任课一', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'course_teacher', '教育学院', '教育技术学', NULL, NULL),
(1402, 'C2021T2', '教技21级任课二', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'course_teacher', '教育学院', '教育技术学', NULL, NULL),
(1403, 'C2022E1', '教育学22级任课一', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'course_teacher', '教育学院', '教育学', NULL, NULL),
(1404, 'C2022E2', '教育学22级任课二', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'course_teacher', '教育学院', '教育学', NULL, NULL),
(1405, 'C2022P1', '学前22级任课一', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'course_teacher', '教育学院', '学前教育', NULL, NULL),
(1406, 'C2022P2', '学前22级任课二', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'course_teacher', '教育学院', '学前教育', NULL, NULL),
(1407, 'C2022X1', '小教22级任课一', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'course_teacher', '教育学院', '小学教育', NULL, NULL),
(1408, 'C2022X2', '小教22级任课二', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'course_teacher', '教育学院', '小学教育', NULL, NULL),
(1409, 'C2022T1', '教技22级任课一', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'course_teacher', '教育学院', '教育技术学', NULL, NULL),
(1410, 'C2022T2', '教技22级任课二', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'course_teacher', '教育学院', '教育技术学', NULL, NULL),
(1411, 'C2023E1', '教育学23级任课一', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'course_teacher', '教育学院', '教育学', NULL, NULL),
(1412, 'C2023E2', '教育学23级任课二', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'course_teacher', '教育学院', '教育学', NULL, NULL),
(1413, 'C2023P1', '学前23级任课一', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'course_teacher', '教育学院', '学前教育', NULL, NULL),
(1414, 'C2023P2', '学前23级任课二', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'course_teacher', '教育学院', '学前教育', NULL, NULL),
(1415, 'C2023X1', '小教23级任课一', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'course_teacher', '教育学院', '小学教育', NULL, NULL),
(1416, 'C2023X2', '小教23级任课二', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'course_teacher', '教育学院', '小学教育', NULL, NULL),
(1417, 'C2023T1', '教技23级任课一', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'course_teacher', '教育学院', '教育技术学', NULL, NULL),
(1418, 'C2023T2', '教技23级任课二', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'course_teacher', '教育学院', '教育技术学', NULL, NULL),
(1419, 'C2024E1', '教育学24级任课一', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'course_teacher', '教育学院', '教育学', NULL, NULL),
(1420, 'C2024E2', '教育学24级任课二', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'course_teacher', '教育学院', '教育学', NULL, NULL),
(1421, 'C2024P1', '学前24级任课一', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'course_teacher', '教育学院', '学前教育', NULL, NULL),
(1422, 'C2024P2', '学前24级任课二', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'course_teacher', '教育学院', '学前教育', NULL, NULL),
(1423, 'C2024X1', '小教24级任课一', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'course_teacher', '教育学院', '小学教育', NULL, NULL),
(1424, 'C2024X2', '小教24级任课二', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'course_teacher', '教育学院', '小学教育', NULL, NULL),
(1425, 'C2024T1', '教技24级任课一', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'course_teacher', '教育学院', '教育技术学', NULL, NULL),
(1426, 'C2024T2', '教技24级任课二', '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO', 'course_teacher', '教育学院', '教育技术学', NULL, NULL);

-- 7) 班级（32 个）
INSERT INTO t_class (id, name, grade, major_id, advisor_id, total_students) VALUES
(33, '教育学2101', '2021', 5, 73, 40),
(34, '教育学2102', '2021', 5, 74, 40),
(35, '学前2101', '2021', 6, 75, 40),
(36, '学前2102', '2021', 6, 76, 40),
(37, '小教2101', '2021', 7, 77, 40),
(38, '小教2102', '2021', 7, 78, 40),
(39, '教技2101', '2021', 8, 79, 40),
(40, '教技2102', '2021', 8, 80, 40),
(41, '教育学2201', '2022', 5, 81, 40),
(42, '教育学2202', '2022', 5, 82, 40),
(43, '学前2201', '2022', 6, 83, 40),
(44, '学前2202', '2022', 6, 84, 40),
(45, '小教2201', '2022', 7, 85, 40),
(46, '小教2202', '2022', 7, 86, 40),
(47, '教技2201', '2022', 8, 87, 40),
(48, '教技2202', '2022', 8, 88, 40),
(49, '教育学2301', '2023', 5, 89, 40),
(50, '教育学2302', '2023', 5, 90, 40),
(51, '学前2301', '2023', 6, 91, 40),
(52, '学前2302', '2023', 6, 92, 40),
(53, '小教2301', '2023', 7, 93, 40),
(54, '小教2302', '2023', 7, 94, 40),
(55, '教技2301', '2023', 8, 95, 40),
(56, '教技2302', '2023', 8, 96, 40),
(57, '教育学2401', '2024', 5, 97, 40),
(58, '教育学2402', '2024', 5, 98, 40),
(59, '学前2401', '2024', 6, 99, 40),
(60, '学前2402', '2024', 6, 100, 40),
(61, '小教2401', '2024', 7, 101, 40),
(62, '小教2402', '2024', 7, 102, 40),
(63, '教技2401', '2024', 8, 103, 40),
(64, '教技2402', '2024', 8, 104, 40);

-- 8) 教师档案：32 名班主任 + 32 名任课教师
INSERT INTO t_teacher (id, teacher_id, user_id, department_id, title, is_class_advisor, class_id) VALUES
(73, 'T2021E01', 1363, 2, '讲师', 1, 33),
(74, 'T2021E02', 1364, 2, '讲师', 1, 34),
(75, 'T2021P01', 1365, 2, '讲师', 1, 35),
(76, 'T2021P02', 1366, 2, '讲师', 1, 36),
(77, 'T2021X01', 1367, 2, '讲师', 1, 37),
(78, 'T2021X02', 1368, 2, '讲师', 1, 38),
(79, 'T2021T01', 1369, 2, '讲师', 1, 39),
(80, 'T2021T02', 1370, 2, '讲师', 1, 40),
(81, 'T2022E01', 1371, 2, '讲师', 1, 41),
(82, 'T2022E02', 1372, 2, '讲师', 1, 42),
(83, 'T2022P01', 1373, 2, '讲师', 1, 43),
(84, 'T2022P02', 1374, 2, '讲师', 1, 44),
(85, 'T2022X01', 1375, 2, '讲师', 1, 45),
(86, 'T2022X02', 1376, 2, '讲师', 1, 46),
(87, 'T2022T01', 1377, 2, '讲师', 1, 47),
(88, 'T2022T02', 1378, 2, '讲师', 1, 48),
(89, 'T2023E01', 1379, 2, '讲师', 1, 49),
(90, 'T2023E02', 1380, 2, '讲师', 1, 50),
(91, 'T2023P01', 1381, 2, '讲师', 1, 51),
(92, 'T2023P02', 1382, 2, '讲师', 1, 52),
(93, 'T2023X01', 1383, 2, '讲师', 1, 53),
(94, 'T2023X02', 1384, 2, '讲师', 1, 54),
(95, 'T2023T01', 1385, 2, '讲师', 1, 55),
(96, 'T2023T02', 1386, 2, '讲师', 1, 56),
(97, 'T2024E01', 1387, 2, '讲师', 1, 57),
(98, 'T2024E02', 1388, 2, '讲师', 1, 58),
(99, 'T2024P01', 1389, 2, '讲师', 1, 59),
(100, 'T2024P02', 1390, 2, '讲师', 1, 60),
(101, 'T2024X01', 1391, 2, '讲师', 1, 61),
(102, 'T2024X02', 1392, 2, '讲师', 1, 62),
(103, 'T2024T01', 1393, 2, '讲师', 1, 63),
(104, 'T2024T02', 1394, 2, '讲师', 1, 64),
(105, 'C2021E1', 1395, 2, '讲师', 0, NULL),
(106, 'C2021E2', 1396, 2, '讲师', 0, NULL),
(107, 'C2021P1', 1397, 2, '讲师', 0, NULL),
(108, 'C2021P2', 1398, 2, '讲师', 0, NULL),
(109, 'C2021X1', 1399, 2, '讲师', 0, NULL),
(110, 'C2021X2', 1400, 2, '讲师', 0, NULL),
(111, 'C2021T1', 1401, 2, '讲师', 0, NULL),
(112, 'C2021T2', 1402, 2, '讲师', 0, NULL),
(113, 'C2022E1', 1403, 2, '讲师', 0, NULL),
(114, 'C2022E2', 1404, 2, '讲师', 0, NULL),
(115, 'C2022P1', 1405, 2, '讲师', 0, NULL),
(116, 'C2022P2', 1406, 2, '讲师', 0, NULL),
(117, 'C2022X1', 1407, 2, '讲师', 0, NULL),
(118, 'C2022X2', 1408, 2, '讲师', 0, NULL),
(119, 'C2022T1', 1409, 2, '讲师', 0, NULL),
(120, 'C2022T2', 1410, 2, '讲师', 0, NULL),
(121, 'C2023E1', 1411, 2, '讲师', 0, NULL),
(122, 'C2023E2', 1412, 2, '讲师', 0, NULL),
(123, 'C2023P1', 1413, 2, '讲师', 0, NULL),
(124, 'C2023P2', 1414, 2, '讲师', 0, NULL),
(125, 'C2023X1', 1415, 2, '讲师', 0, NULL),
(126, 'C2023X2', 1416, 2, '讲师', 0, NULL),
(127, 'C2023T1', 1417, 2, '讲师', 0, NULL),
(128, 'C2023T2', 1418, 2, '讲师', 0, NULL),
(129, 'C2024E1', 1419, 2, '讲师', 0, NULL),
(130, 'C2024E2', 1420, 2, '讲师', 0, NULL),
(131, 'C2024P1', 1421, 2, '讲师', 0, NULL),
(132, 'C2024P2', 1422, 2, '讲师', 0, NULL),
(133, 'C2024X1', 1423, 2, '讲师', 0, NULL),
(134, 'C2024X2', 1424, 2, '讲师', 0, NULL),
(135, 'C2024T1', 1425, 2, '讲师', 0, NULL),
(136, 'C2024T2', 1426, 2, '讲师', 0, NULL);

-- 9) 960 名学生账号
INSERT INTO t_user (id, account, name, password, role, college, major, class_name, grade)
WITH RECURSIVE seq AS (SELECT 1 n UNION ALL SELECT n + 1 FROM seq WHERE n < 960),
cls AS (
    SELECT 33 id, '教育学2101' name, 5 major_id, '2021' grade UNION ALL
    SELECT 34, '教育学2102', 5, '2021' UNION ALL
    SELECT 35, '学前2101', 6, '2021' UNION ALL
    SELECT 36, '学前2102', 6, '2021' UNION ALL
    SELECT 37, '小教2101', 7, '2021' UNION ALL
    SELECT 38, '小教2102', 7, '2021' UNION ALL
    SELECT 39, '教技2101', 8, '2021' UNION ALL
    SELECT 40, '教技2102', 8, '2021' UNION ALL
    SELECT 41, '教育学2201', 5, '2022' UNION ALL
    SELECT 42, '教育学2202', 5, '2022' UNION ALL
    SELECT 43, '学前2201', 6, '2022' UNION ALL
    SELECT 44, '学前2202', 6, '2022' UNION ALL
    SELECT 45, '小教2201', 7, '2022' UNION ALL
    SELECT 46, '小教2202', 7, '2022' UNION ALL
    SELECT 47, '教技2201', 8, '2022' UNION ALL
    SELECT 48, '教技2202', 8, '2022' UNION ALL
    SELECT 49, '教育学2301', 5, '2023' UNION ALL
    SELECT 50, '教育学2302', 5, '2023' UNION ALL
    SELECT 51, '学前2301', 6, '2023' UNION ALL
    SELECT 52, '学前2302', 6, '2023' UNION ALL
    SELECT 53, '小教2301', 7, '2023' UNION ALL
    SELECT 54, '小教2302', 7, '2023' UNION ALL
    SELECT 55, '教技2301', 8, '2023' UNION ALL
    SELECT 56, '教技2302', 8, '2023' UNION ALL
    SELECT 57, '教育学2401', 5, '2024' UNION ALL
    SELECT 58, '教育学2402', 5, '2024' UNION ALL
    SELECT 59, '学前2401', 6, '2024' UNION ALL
    SELECT 60, '学前2402', 6, '2024' UNION ALL
    SELECT 61, '小教2401', 7, '2024' UNION ALL
    SELECT 62, '小教2402', 7, '2024' UNION ALL
    SELECT 63, '教技2401', 8, '2024' UNION ALL
    SELECT 64, '教技2402', 8, '2024'
)
SELECT 1426 + s.n,
       CONCAT(cl.grade, cl.major_id, LPAD(((cl.id - 33) % 2) * 40 + MOD(s.n - 1, 40) + 1, 3, '0')),
       CONCAT('教院学生', LPAD(s.n, 3, '0')),
       '$2a$12$maQjoyay/Wbtatsza7qnIe0SzLSUiip3PzmAtEz1F415jS.w6z3HO',
       'student', '教育学院',
       CASE cl.major_id WHEN 5 THEN '教育学' WHEN 6 THEN '学前教育' WHEN 7 THEN '小学教育' ELSE '教育技术学' END,
       cl.name, cl.grade
FROM seq s
JOIN cls cl ON cl.id = 33 + FLOOR((s.n - 1) / 40);

-- 10) 学生档案
INSERT INTO t_student (student_id, user_id, class_id, major_id, department_id, gpa, `rank`, total_students, alert_level, total_credits, required_credits)
SELECT u.account, u.id, c.id, c.major_id, 2,
       ROUND(2.0 + MOD(u.id * 7, 200) / 100, 2),
       MOD(u.id * 3, 40) + 1, 40, 'none',
       CASE c.grade
           WHEN '2021' THEN 118 + MOD(u.id, 18)
           WHEN '2022' THEN 78 + MOD(u.id, 10)
           WHEN '2023' THEN 46 + MOD(u.id, 8)
           ELSE 16 + MOD(u.id, 4)
       END, 160.0
FROM t_user u
JOIN t_class c ON c.name = u.class_name
WHERE u.role = 'student' AND u.id BETWEEN 1427 AND 2386;

-- 11) 开课记录
INSERT INTO t_course_class (course_id, term, class_name, teacher_id, max_students, enrolled_students, status)
SELECT c.id, t.term, cl.name,
       CASE cl.grade
           WHEN '2021' THEN 105 + (cl.major_id - 5) * 2 + IF(MOD(c.id - 1, 6) IN (0, 1, 3), 0, 1)
           WHEN '2022' THEN 113 + (cl.major_id - 5) * 2 + IF(MOD(c.id - 1, 6) IN (0, 1, 3), 0, 1)
           WHEN '2023' THEN 121 + (cl.major_id - 5) * 2 + IF(MOD(c.id - 1, 6) IN (0, 1, 3), 0, 1)
           ELSE 129 + (cl.major_id - 5) * 2 + IF(MOD(c.id - 1, 6) IN (0, 1, 3), 0, 1)
       END,
       40, 40,
       CASE WHEN t.term = '2024-2025-1' THEN 'active' ELSE 'finished' END
FROM t_course c
JOIN t_class cl ON cl.id BETWEEN 33 AND 64
JOIN (
    SELECT '2021' grade, '2021-2022-1' term UNION ALL
    SELECT '2021', '2021-2022-2' UNION ALL
    SELECT '2021', '2022-2023-1' UNION ALL
    SELECT '2021', '2022-2023-2' UNION ALL
    SELECT '2021', '2023-2024-1' UNION ALL
    SELECT '2021', '2023-2024-2' UNION ALL
    SELECT '2021', '2024-2025-1' UNION ALL
    SELECT '2022', '2022-2023-1' UNION ALL
    SELECT '2022', '2022-2023-2' UNION ALL
    SELECT '2022', '2023-2024-1' UNION ALL
    SELECT '2022', '2023-2024-2' UNION ALL
    SELECT '2022', '2024-2025-1' UNION ALL
    SELECT '2023', '2023-2024-1' UNION ALL
    SELECT '2023', '2023-2024-2' UNION ALL
    SELECT '2023', '2024-2025-1' UNION ALL
    SELECT '2024', '2024-2025-1'
) t ON t.grade = cl.grade
WHERE c.id BETWEEN 25 + (cl.major_id - 5) * 6 AND 30 + (cl.major_id - 5) * 6;

-- 12) 原始成绩
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
WHERE u.id BETWEEN 1427 AND 2386;

-- 13) 当前学期保留部分挂科
UPDATE t_grade g
JOIN t_course_class cc ON cc.id = g.course_class_id
JOIN t_student s ON s.id = g.student_id
JOIN t_class cl ON cl.id = s.class_id
JOIN t_user u ON u.id = s.user_id
SET g.score = 42 + MOD(g.id, 15)
WHERE u.id BETWEEN 1427 AND 2386
  AND cc.term = '2024-2025-1'
  AND MOD(s.id, 8) IN (0, 3)
  AND cc.course_id IN (
      25 + (cl.major_id - 5) * 6 + MOD(s.id, 6),
      25 + (cl.major_id - 5) * 6 + MOD(s.id + 3, 6)
  );

-- 14) 状态与绩点
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
WHERE u.id BETWEEN 1427 AND 2386;

-- 15) 体测、考勤、志愿、心理、竞赛
INSERT INTO t_health_report (student_id, term, total_score, items, report_url)
SELECT s.id, '2024-2025-1', 60 + MOD(s.id * 5, 36),
       JSON_ARRAY(
           JSON_OBJECT('name', '体重指数', 'score', 60 + MOD(s.id * 3, 36), 'level', '良好'),
           JSON_OBJECT('name', '肺活量', 'score', 60 + MOD(s.id * 5, 36), 'level', '良好'),
           JSON_OBJECT('name', '耐力跑', 'score', 60 + MOD(s.id * 7, 36), 'level', '良好')
       ), NULL
FROM t_student s
JOIN t_user u ON u.id = s.user_id
WHERE u.id BETWEEN 1427 AND 2386;

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
WHERE u.id BETWEEN 1427 AND 2386;

INSERT INTO t_volunteer (student_id, term, hours, description)
SELECT s.id, '2024-2025-1', 6 + MOD(s.id * 7, 60), '模拟志愿服务记录'
FROM t_student s
JOIN t_user u ON u.id = s.user_id
WHERE u.id BETWEEN 1427 AND 2386;

INSERT INTO t_psychology (student_id, term, score, level, note)
SELECT s.id, '2024-2025-1',
       62 + MOD(s.id * 5, 36),
       CASE WHEN 62 + MOD(s.id * 5, 36) >= 85 THEN 'good'
            WHEN 62 + MOD(s.id * 5, 36) >= 70 THEN 'normal'
            ELSE 'attention' END,
       '模拟心理测评记录'
FROM t_student s
JOIN t_user u ON u.id = s.user_id
WHERE u.id BETWEEN 1427 AND 2386;

INSERT INTO t_competition (student_id, term, competition_name, level, award, points)
SELECT s.id, '2024-2025-1',
       CASE cl.major_id WHEN 5 THEN '师范生教学技能大赛'
                        WHEN 6 THEN '学前教育创新大赛'
                        WHEN 7 THEN '小学教育案例大赛'
                        ELSE '教育技术应用竞赛' END,
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
WHERE u.id BETWEEN 1427 AND 2386 AND MOD(s.id, 4) = 0;

-- 16) 培养方案
INSERT INTO t_major_plan (major_id, grade, dimensions, standard, actual, tips)
SELECT m.id, v.grade, v.dimensions, v.standard, v.actual, v.tips
FROM t_major m
CROSS JOIN (
    SELECT '2021' grade, '["公共基础","专业必修","专业选修","实践环节","通识教育"]' dimensions,
           '[48,62,20,18,12]' standard, '[46,58,24,16,12]' actual,
           '["专业必修完成度偏低","建议加强师范技能训练"]' tips
    UNION ALL SELECT '2022', '["公共基础","专业必修","专业选修","实践环节","通识教育"]', '[48,62,20,18,12]', '[42,54,20,15,12]', '["大三阶段整体完成良好","建议关注实习与教师资格准备"]'
    UNION ALL SELECT '2023', '["公共基础","专业必修","专业选修","实践环节","通识教育"]', '[48,62,20,18,12]', '[30,38,14,10,12]', '["大二基础课程完成度正常","建议加强教学实践"]'
    UNION ALL SELECT '2024', '["公共基础","专业必修","专业选修","实践环节","通识教育"]', '[48,62,20,18,12]', '[20,28,10,8,10]', '["大一处于基础学习阶段","建议提前规划师范方向"]'
) v
WHERE m.id BETWEEN 5 AND 8
  AND NOT EXISTS (
      SELECT 1 FROM t_major_plan p WHERE p.major_id = m.id AND p.grade = v.grade
  );
