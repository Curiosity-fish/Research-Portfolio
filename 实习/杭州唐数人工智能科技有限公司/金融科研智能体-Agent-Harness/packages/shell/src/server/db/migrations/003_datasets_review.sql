-- 003_datasets_review.sql — 数据集与技能对齐:公开资源先上线后审核(A7)
ALTER TABLE datasets ADD COLUMN review_state TEXT NOT NULL DEFAULT 'none' CHECK (review_state IN ('none','pending','approved','rejected'));
