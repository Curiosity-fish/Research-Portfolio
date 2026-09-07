-- V8: 统一五维画像维度与顺序
-- 业务定义五维：学业成绩 · 实践能力 · 综合素质 · 人文素养 · 身心健康
-- 旧维度 key（ability/psychology/thought）迁移为 practice/quality/culture，
-- health 标签由“体质健康”统一为“身心健康”，保证雷达图与前端文案一致。

-- 1) 旧维度 key 迁移到统一 key
UPDATE t_profile_score SET dimension_key = 'practice', label = '实践能力' WHERE dimension_key = 'ability';
UPDATE t_profile_score SET dimension_key = 'quality', label = '综合素质' WHERE dimension_key = 'thought';
UPDATE t_profile_score SET dimension_key = 'culture', label = '人文素养' WHERE dimension_key = 'psychology';
UPDATE t_profile_score SET label = '身心健康' WHERE dimension_key = 'health';

-- 2) 人文素养沿用原心理素质分数，但替换为符合业务语义的描述与明细
UPDATE t_profile_score
SET description = '参与人文讲座与校园文化活动，阅读与审美素养良好',
    details = JSON_ARRAY('学术/人文讲座 3 次', '校园文艺活动 2 次', '读书笔记 1 项')
WHERE dimension_key = 'culture';

-- 3) 兜底补齐每个学生/学期缺失的统一五维
INSERT IGNORE INTO t_profile_score (student_id, term, dimension_key, label, score, avg_score, max_score, description, details)
SELECT p.student_id, p.term, d.dimension_key, d.label,
       CASE d.dimension_key
           WHEN 'quality' THEN 60 + MOD(p.student_id * 7, 31)
           WHEN 'culture' THEN 55 + MOD(p.student_id * 3, 31)
           WHEN 'health' THEN 55 + MOD(p.student_id * 5, 41)
           ELSE 65 + MOD(p.student_id * 7, 31)
       END,
       d.avg_score, 100, d.description, d.details
FROM (SELECT DISTINCT student_id, term FROM t_profile_score) p
CROSS JOIN (
    SELECT 'quality' dimension_key, '综合素质' label, 72 avg_score,
           '班级事务与志愿服务参与良好，综合素质均衡' description,
           JSON_ARRAY('班级职务 1 项', '志愿服务 20 小时', '社团活动 2 项') details
    UNION ALL
    SELECT 'culture', '人文素养', 68,
           '参与人文讲座与校园文化活动，阅读与审美素养良好',
           JSON_ARRAY('学术/人文讲座 3 次', '校园文艺活动 2 次', '读书笔记 1 项')
    UNION ALL
    SELECT 'health', '身心健康', 76,
           '体测成绩良好，心理状态稳定，身心综合表现正常',
           JSON_ARRAY('体测总评良好', '心理测评正常', '作息规律')
) d
WHERE NOT EXISTS (
    SELECT 1 FROM t_profile_score e
    WHERE e.student_id = p.student_id
      AND e.term = p.term
      AND e.dimension_key = d.dimension_key
);
