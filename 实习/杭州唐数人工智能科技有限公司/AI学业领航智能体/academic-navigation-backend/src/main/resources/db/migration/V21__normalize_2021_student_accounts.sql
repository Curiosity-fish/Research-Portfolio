-- V21: 将 2021 届学生学号统一为 2021xxxx（与班级届别一致）
UPDATE t_user u
JOIN (
    SELECT s.user_id, c.major_id, c.id AS class_id,
           ROW_NUMBER() OVER (PARTITION BY s.class_id ORDER BY s.user_id) AS rn
    FROM t_student s
    JOIN t_class c ON c.id = s.class_id
    WHERE c.grade = '2021'
) m ON m.user_id = u.id
SET u.account = CONCAT('2021', m.major_id, LPAD(((m.class_id - 1) % 2) * 40 + m.rn, 3, '0'))
WHERE u.role = 'student';

UPDATE t_student s
JOIN t_user u ON u.id = s.user_id
SET s.student_id = u.account
WHERE u.role = 'student' AND u.account LIKE '2021%';
