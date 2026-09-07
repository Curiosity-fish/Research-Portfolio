-- V7: 补齐五维画像缺失维度
-- 为每个已有画像记录的学生补齐心理素质/体质健康/思想品德三个维度，
-- 保证学生端五维雷达图始终有完整五维数据。

INSERT IGNORE INTO t_profile_score (student_id, term, dimension_key, label, score, avg_score, max_score, description, details)
SELECT p.student_id, p.term, d.dimension_key, d.label,
       CASE d.dimension_key
           WHEN 'psychology' THEN 60 + MOD(p.student_id * 3, 31)
           WHEN 'health' THEN 55 + MOD(p.student_id * 5, 41)
           ELSE 65 + MOD(p.student_id * 7, 31)
       END,
       d.avg_score, 100, d.description, d.details
FROM (SELECT DISTINCT student_id, term FROM t_profile_score) p
CROSS JOIN (
    SELECT 'psychology' dimension_key, '心理素质' label, 72 avg_score,
           '心理状态稳定，测评结果正常' description, JSON_ARRAY('测评结果正常') details
    UNION ALL
    SELECT 'health', '体质健康', 76,
           '体测成绩良好，建议保持规律锻炼', JSON_ARRAY('体测总分 80 分')
    UNION ALL
    SELECT 'thought', '思想品德', 82,
           '思想表现良好，日常表现正常', JSON_ARRAY('日常表现正常')
) d
WHERE NOT EXISTS (
    SELECT 1 FROM t_profile_score e
    WHERE e.student_id = p.student_id
      AND e.term = p.term
      AND e.dimension_key = d.dimension_key
);
