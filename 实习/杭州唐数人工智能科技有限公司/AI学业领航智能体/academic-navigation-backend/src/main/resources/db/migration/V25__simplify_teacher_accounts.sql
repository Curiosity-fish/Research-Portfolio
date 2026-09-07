-- V25: 班主任工号统一为 T001-T064，任课教师工号统一为 C001-C064

UPDATE t_user u
JOIN (
    SELECT t.user_id, ROW_NUMBER() OVER (ORDER BY t.id) AS rn
    FROM t_teacher t
    WHERE t.is_class_advisor = 1
) m ON m.user_id = u.id
SET u.account = CONCAT('T', LPAD(m.rn, 3, '0'))
WHERE u.role = 'teacher';

UPDATE t_teacher t
JOIN t_user u ON u.id = t.user_id
SET t.teacher_id = u.account
WHERE t.is_class_advisor = 1;

UPDATE t_user u
JOIN (
    SELECT t.user_id, ROW_NUMBER() OVER (ORDER BY t.id) AS rn
    FROM t_teacher t
    WHERE t.is_class_advisor = 0
) m ON m.user_id = u.id
SET u.account = CONCAT('C', LPAD(m.rn, 3, '0'))
WHERE u.role = 'course_teacher';

UPDATE t_teacher t
JOIN t_user u ON u.id = t.user_id
SET t.teacher_id = u.account
WHERE t.is_class_advisor = 0;
