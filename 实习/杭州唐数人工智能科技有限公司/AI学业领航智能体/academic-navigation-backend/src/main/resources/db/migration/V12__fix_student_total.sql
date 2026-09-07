-- V12: 修正专业总人数（应用统计共 2 个班，80 人）
UPDATE t_student s
SET total_students = (
    SELECT cnt FROM (
        SELECT major_id, COUNT(*) cnt FROM t_student GROUP BY major_id
    ) m WHERE m.major_id = s.major_id
);
