-- V20: 按届别修正已修学分，使大一至大四数据更贴近实际
UPDATE t_student s
JOIN t_class c ON c.id = s.class_id
JOIN t_user u ON u.id = s.user_id
SET s.total_credits = CASE c.grade
    WHEN '2021' THEN 118 + MOD(s.id, 18)
    WHEN '2022' THEN 78 + MOD(s.id, 10)
    WHEN '2023' THEN 46 + MOD(s.id, 8)
    WHEN '2024' THEN 16 + MOD(s.id, 4)
END
WHERE u.role = 'student';
