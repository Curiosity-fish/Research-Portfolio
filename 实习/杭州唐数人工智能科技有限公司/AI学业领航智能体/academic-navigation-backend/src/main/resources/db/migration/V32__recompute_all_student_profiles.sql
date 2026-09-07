-- 统一重算新旧学期画像，避免旧脚本生成的详情与当前成绩/GPA口径不一致。

UPDATE t_profile_score ps
JOIN t_gpa_history gh ON gh.student_id = ps.student_id AND gh.term = ps.term
JOIN (
    SELECT student_id,
           term,
           AVG(score) avg_score,
           SUM(score < 60 OR status = 'failed') failed_count
    FROM t_grade
    GROUP BY student_id, term
) gs ON gs.student_id = ps.student_id AND gs.term = ps.term
SET ps.score = LEAST(100, GREATEST(0, ROUND(
        35 + gh.gpa / 4 * 40
        + (1 - gh.`rank` / NULLIF(gh.total_students, 0)) * 12
        + gs.avg_score / 100 * 8
        - gs.failed_count * 5
    ))),
    ps.avg_score = 72,
    ps.max_score = 100,
    ps.description = '基于当学期GPA、排名、挂科数与课程均分计算',
    ps.details = JSON_ARRAY(
        CONCAT('GPA: ', FORMAT(gh.gpa, 2)),
        CONCAT('专业排名: ', gh.`rank`, ' / ', gh.total_students),
        CONCAT('挂科门数: ', gs.failed_count),
        CONCAT('课程均分: ', ROUND(gs.avg_score))
    )
WHERE ps.dimension_key = 'academic';

UPDATE t_profile_score ps
LEFT JOIN (
    SELECT student_id, term, MAX(points) max_points, COUNT(*) item_count
    FROM t_competition
    GROUP BY student_id, term
) cs ON cs.student_id = ps.student_id AND cs.term = ps.term
SET ps.score = CASE
        WHEN COALESCE(cs.max_points, 0) >= 90 THEN 90
        WHEN COALESCE(cs.max_points, 0) >= 80 THEN 82
        WHEN COALESCE(cs.max_points, 0) >= 70 THEN 74
        WHEN COALESCE(cs.item_count, 0) > 0 THEN 66
        ELSE 55
    END,
    ps.avg_score = 68,
    ps.max_score = 100,
    ps.description = '基于当学期竞赛经历与实践积分计算',
    ps.details = JSON_ARRAY(
        CONCAT('竞赛积分: ', COALESCE(cs.max_points, 0)),
        CONCAT('竞赛记录: ', COALESCE(cs.item_count, 0), ' 项')
    )
WHERE ps.dimension_key = 'practice';

UPDATE t_profile_score ps
JOIN t_volunteer v ON v.student_id = ps.student_id AND v.term = ps.term
JOIN t_attendance a ON a.student_id = ps.student_id AND a.term = ps.term
SET ps.score = CASE
        WHEN v.hours >= 40 THEN 92
        WHEN v.hours >= 20 THEN 85
        WHEN v.hours >= 8 THEN 78
        WHEN v.hours > 0 THEN 70
        ELSE 62
    END,
    ps.avg_score = 85,
    ps.max_score = 100,
    ps.description = '基于当学期志愿服务时长与出勤情况计算',
    ps.details = JSON_ARRAY(
        CONCAT('志愿时长: ', v.hours, ' 小时'),
        CONCAT('出勤率: ', GREATEST(0, 100 - ROUND(a.absent_count * 100 / NULLIF(a.total_classes, 0))), '%')
    )
WHERE ps.dimension_key = 'quality';

UPDATE t_profile_score ps
JOIN t_psychology p ON p.student_id = ps.student_id AND p.term = ps.term
SET ps.score = CASE WHEN p.score >= 85 THEN 90 WHEN p.score >= 70 THEN 78 ELSE 68 END,
    ps.avg_score = 75,
    ps.max_score = 100,
    ps.description = '基于当学期综合测评结果计算',
    ps.details = JSON_ARRAY(CONCAT('综合测评: ', p.score, ' 分'))
WHERE ps.dimension_key = 'culture';

UPDATE t_profile_score ps
JOIN t_health_report h ON h.student_id = ps.student_id AND h.term = ps.term
SET ps.score = h.total_score,
    ps.avg_score = 80,
    ps.max_score = 100,
    ps.description = '基于当学期体测总分计算',
    ps.details = JSON_ARRAY(CONCAT('体测总分: ', h.total_score, ' 分'))
WHERE ps.dimension_key = 'health';
