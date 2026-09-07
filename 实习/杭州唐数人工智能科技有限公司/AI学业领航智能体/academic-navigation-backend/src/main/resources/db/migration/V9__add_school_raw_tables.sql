-- V6: 校方原始补充数据表（竞赛/志愿时长/考勤/心理测评）
-- 这些数据模拟“校方提供”，用于计算五维画像中的实践能力、思想品德、心理素质等派生指标

CREATE TABLE IF NOT EXISTS t_competition (
    id               BIGINT       NOT NULL AUTO_INCREMENT COMMENT '主键',
    student_id       BIGINT       NOT NULL                COMMENT '学生档案ID',
    term             VARCHAR(20)  DEFAULT NULL            COMMENT '学期',
    competition_name VARCHAR(200) DEFAULT NULL            COMMENT '竞赛名称',
    level            VARCHAR(20)  DEFAULT NULL            COMMENT '级别: national/provincial/school',
    award            VARCHAR(50)  DEFAULT NULL            COMMENT '奖项: first/second/third/participation',
    points           INT          NOT NULL DEFAULT 0      COMMENT '实践积分',
    created_at       DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at       DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_competition_student (student_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='学生竞赛记录（校方原始数据）';

CREATE TABLE IF NOT EXISTS t_volunteer (
    id          BIGINT       NOT NULL AUTO_INCREMENT COMMENT '主键',
    student_id  BIGINT       NOT NULL                COMMENT '学生档案ID',
    term        VARCHAR(20)  DEFAULT NULL            COMMENT '学期',
    hours       DECIMAL(5,1) NOT NULL DEFAULT 0      COMMENT '志愿时长（小时）',
    description VARCHAR(200) DEFAULT NULL            COMMENT '服务内容',
    created_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_volunteer (student_id, term),
    KEY idx_volunteer_student (student_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='志愿时长（校方原始数据）';

CREATE TABLE IF NOT EXISTS t_attendance (
    id            BIGINT       NOT NULL AUTO_INCREMENT COMMENT '主键',
    student_id    BIGINT       NOT NULL                COMMENT '学生档案ID',
    term          VARCHAR(20)  DEFAULT NULL            COMMENT '学期',
    total_classes INT          NOT NULL DEFAULT 0      COMMENT '应到总课时',
    absent_count  INT          NOT NULL DEFAULT 0      COMMENT '缺勤次数',
    late_count    INT          NOT NULL DEFAULT 0      COMMENT '迟到次数',
    note          VARCHAR(200) DEFAULT NULL            COMMENT '备注',
    created_at    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_attendance (student_id, term),
    KEY idx_attendance_student (student_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='课堂考勤（校方原始数据）';

CREATE TABLE IF NOT EXISTS t_psychology (
    id          BIGINT       NOT NULL AUTO_INCREMENT COMMENT '主键',
    student_id  BIGINT       NOT NULL                COMMENT '学生档案ID',
    term        VARCHAR(20)  DEFAULT NULL            COMMENT '学期',
    score       INT          NOT NULL DEFAULT 0      COMMENT '心理测评分数',
    level       VARCHAR(20)  DEFAULT NULL            COMMENT '等级: good/normal/attention',
    note        VARCHAR(200) DEFAULT NULL            COMMENT '备注',
    created_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_psychology (student_id, term),
    KEY idx_psychology_student (student_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='心理测评（校方原始数据）';

-- ============ 模拟校方提供的原始数据 ============

INSERT IGNORE INTO t_attendance (student_id, term, total_classes, absent_count, late_count, note)
SELECT s.id, '2024-2025-1', 60, MOD(s.id * 3, 12), MOD(s.id, 5), '模拟考勤数据'
FROM t_student s;

INSERT IGNORE INTO t_volunteer (student_id, term, hours, description)
SELECT s.id, '2024-2025-1', 6 + MOD(s.id * 7, 60), '模拟志愿服务记录'
FROM t_student s;

INSERT IGNORE INTO t_psychology (student_id, term, score, level, note)
SELECT s.id, '2024-2025-1',
       62 + MOD(s.id * 5, 36),
       CASE WHEN 62 + MOD(s.id * 5, 36) >= 85 THEN 'good'
            WHEN 62 + MOD(s.id * 5, 36) >= 70 THEN 'normal'
            ELSE 'attention' END,
       '模拟心理测评记录'
FROM t_student s;

INSERT INTO t_competition (student_id, term, competition_name, level, award, points)
SELECT s.id, '2024-2025-1',
       CASE WHEN MOD(s.id, 6) = 0 THEN '全国大学生数学建模竞赛'
            WHEN MOD(s.id, 6) = 1 THEN '省级程序设计大赛'
            WHEN MOD(s.id, 6) = 2 THEN '校级创新创业大赛'
            ELSE '大学生课外学术竞赛' END,
       CASE WHEN MOD(s.id, 6) = 0 THEN 'national'
            WHEN MOD(s.id, 6) = 1 THEN 'provincial'
            ELSE 'school' END,
       CASE WHEN MOD(s.id, 4) = 0 THEN 'first'
            WHEN MOD(s.id, 4) = 1 THEN 'second'
            WHEN MOD(s.id, 4) = 2 THEN 'third'
            ELSE 'participation' END,
       CASE WHEN MOD(s.id, 6) = 0 AND MOD(s.id, 4) = 0 THEN 95
            WHEN MOD(s.id, 6) = 0 THEN 88
            WHEN MOD(s.id, 6) = 1 AND MOD(s.id, 4) = 0 THEN 85
            WHEN MOD(s.id, 6) = 1 THEN 80
            WHEN MOD(s.id, 4) = 0 THEN 78
            WHEN MOD(s.id, 4) = 1 THEN 72
            ELSE 65 END
FROM t_student s
WHERE MOD(s.id, 4) = 0
  AND NOT EXISTS (
      SELECT 1 FROM t_competition c
      WHERE c.student_id = s.id AND c.term = '2024-2025-1' AND c.competition_name = (
          CASE WHEN MOD(s.id, 6) = 0 THEN '全国大学生数学建模竞赛'
               WHEN MOD(s.id, 6) = 1 THEN '省级程序设计大赛'
               WHEN MOD(s.id, 6) = 2 THEN '校级创新创业大赛'
               ELSE '大学生课外学术竞赛' END
      )
  );
