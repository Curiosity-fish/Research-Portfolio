-- V24: 修复近四学期预警趋势
-- 1) 现有未闭环预警归入最近学期 2024-2025-1
UPDATE t_alert
SET trigger_date = '2025-01-15', pushed_at = '2025-01-15 14:00'
WHERE status IN ('pending', 'processing') AND trigger_date = '2026-08-13';

-- 2) 依据历史学期挂科生成已闭环预警，供近四学期趋势使用
INSERT INTO t_alert (student_id, level, type, title, description, course, failed_courses,
                     trigger_date, status, suggestion, trigger_event, pushed_at)
SELECT s.id,
       CASE WHEN f.failed_cnt >= 4 THEN 'red'
            WHEN f.failed_cnt >= 2 THEN 'orange'
            ELSE 'yellow' END,
       '挂科预警',
       CASE WHEN f.failed_cnt >= 4 THEN '多门课程不及格'
            WHEN f.failed_cnt >= 2 THEN '累计挂科2门及以上'
            ELSE '单门课程不及格' END,
       CONCAT('该学期累计不及格 ', f.failed_cnt, ' 门，已完成闭环处理'),
       JSON_UNQUOTE(JSON_EXTRACT(f.failed_names, '$[0]')),
       f.failed_names,
       f.term_date,
       'resolved',
       '已完成约谈并制定学业恢复计划',
       CASE WHEN f.failed_cnt >= 4 THEN '累计不及格超4门'
            WHEN f.failed_cnt >= 2 THEN '累计挂科2门及以上'
            ELSE '课程成绩低于60分' END,
       CONCAT(DATE_FORMAT(f.term_date, '%Y-%m-%d'), ' 10:00')
FROM (
    SELECT g.student_id, g.term,
           COUNT(*) AS failed_cnt,
           CONCAT('["', GROUP_CONCAT(DISTINCT c.name SEPARATOR '","'), '"]') AS failed_names,
           CASE g.term
               WHEN '2022-2023-2' THEN DATE '2023-07-10'
               WHEN '2023-2024-1' THEN DATE '2024-01-15'
               WHEN '2023-2024-2' THEN DATE '2024-07-10'
               ELSE DATE '2025-01-15'
           END AS term_date
    FROM t_grade g
    JOIN t_course_class cc ON cc.id = g.course_class_id
    JOIN t_course c ON c.id = cc.course_id
    WHERE g.status = 'failed'
      AND g.term IN ('2022-2023-2', '2023-2024-1', '2023-2024-2')
    GROUP BY g.student_id, g.term, term_date
) f
JOIN t_student s ON s.id = f.student_id
WHERE NOT EXISTS (
    SELECT 1 FROM t_alert a
    WHERE a.student_id = s.id AND a.status = 'resolved' AND a.trigger_date = f.term_date
);
