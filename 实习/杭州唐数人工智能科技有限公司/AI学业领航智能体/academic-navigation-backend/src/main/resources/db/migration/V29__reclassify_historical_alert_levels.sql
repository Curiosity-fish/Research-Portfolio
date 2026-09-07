-- V29: reclassify historical warning levels using the current standard.
-- Historical attendance snapshots are not available, so historical records
-- are classified from the actual failed-course count for that term:
-- 1 failed course -> orange; 2+ failed courses -> red.

DROP TEMPORARY TABLE IF EXISTS tmp_historical_failed_v29;

CREATE TEMPORARY TABLE tmp_historical_failed_v29 AS
SELECT g.student_id,
       g.term,
       COUNT(*) AS failed_count
FROM t_grade g
WHERE g.status = 'failed'
  AND g.term IN ('2022-2023-2', '2023-2024-1', '2023-2024-2')
GROUP BY g.student_id, g.term;

ALTER TABLE tmp_historical_failed_v29 ADD PRIMARY KEY (student_id, term);

UPDATE t_alert a
JOIN tmp_historical_failed_v29 f
  ON f.student_id = a.student_id
 AND f.term = CASE a.trigger_date
        WHEN DATE '2023-07-10' THEN '2022-2023-2'
        WHEN DATE '2024-01-15' THEN '2023-2024-1'
        WHEN DATE '2024-07-10' THEN '2023-2024-2'
    END
SET a.level = CASE
        WHEN f.failed_count >= 2 THEN 'red'
        WHEN f.failed_count = 1 THEN 'orange'
        ELSE 'yellow'
    END,
    a.title = CASE
        WHEN f.failed_count >= 2 THEN '红色预警：挂科两门及以上'
        WHEN f.failed_count = 1 THEN '橙色预警：挂科一门'
        ELSE '黄色预警：学业表现需要关注'
    END,
    a.description = CONCAT('该学期挂科 ', f.failed_count, ' 门，已完成闭环处理。'),
    a.trigger_event = CASE
        WHEN f.failed_count >= 2 THEN '挂科两门及以上'
        WHEN f.failed_count = 1 THEN '挂科一门'
        ELSE '学业表现需要关注'
    END
WHERE a.status = 'resolved'
  AND a.trigger_date IN (DATE '2023-07-10', DATE '2024-01-15', DATE '2024-07-10');

DROP TEMPORARY TABLE IF EXISTS tmp_historical_failed_v29;
