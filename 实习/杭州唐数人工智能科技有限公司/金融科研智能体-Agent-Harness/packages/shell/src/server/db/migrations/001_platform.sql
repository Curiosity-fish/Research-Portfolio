-- 001_platform.sql — 平台数据模型（spec 第 3 节）
-- 时间戳一律为毫秒 INTEGER；id 一律 TEXT(uuid)。

CREATE TABLE users (
	id TEXT PRIMARY KEY,
	username TEXT NOT NULL UNIQUE,
	password_hash TEXT NOT NULL,
	role TEXT NOT NULL CHECK (role IN ('admin','teacher','student')),
	status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active','disabled')),
	created_at INTEGER NOT NULL
);

CREATE TABLE course_groups (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	type TEXT NOT NULL CHECK (type IN ('course','research')),
	owner_id TEXT NOT NULL REFERENCES users(id),
	status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active','archived')),
	created_at INTEGER NOT NULL
);

CREATE TABLE group_members (
	group_id TEXT NOT NULL REFERENCES course_groups(id) ON DELETE CASCADE,
	user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	role TEXT NOT NULL CHECK (role IN ('owner','member')),
	PRIMARY KEY (group_id, user_id)
);

-- 项目：scope_type = private | group；挂组时 group_id 指向课程组（一个组多个项目）
CREATE TABLE projects (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	folder TEXT,
	owner_id TEXT REFERENCES users(id),
	scope_type TEXT NOT NULL DEFAULT 'private' CHECK (scope_type IN ('private','group')),
	group_id TEXT REFERENCES course_groups(id),
	created_at INTEGER NOT NULL,
	updated_at INTEGER NOT NULL
);
CREATE INDEX idx_projects_owner ON projects(owner_id);
CREATE INDEX idx_projects_group ON projects(group_id);

-- 项目与 agent 会话的归属关系（session_id 为 agent 会话 id，唯一）
CREATE TABLE project_session_links (
	project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
	session_id TEXT PRIMARY KEY
);
CREATE INDEX idx_psl_project ON project_session_links(project_id);

-- 技能库
CREATE TABLE skills (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	repo_url TEXT,
	version TEXT,
	visibility TEXT NOT NULL CHECK (visibility IN ('private','group','public')),
	group_id TEXT REFERENCES course_groups(id),
	owner_id TEXT REFERENCES users(id),
	path TEXT NOT NULL,
	status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active','removed')),
	review_state TEXT NOT NULL DEFAULT 'none' CHECK (review_state IN ('none','pending','approved','rejected')),
	created_at INTEGER NOT NULL
);
CREATE INDEX idx_skills_owner ON skills(owner_id);
CREATE INDEX idx_skills_group ON skills(group_id);

-- 数据集库
CREATE TABLE datasets (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	size_bytes INTEGER NOT NULL DEFAULT 0,
	visibility TEXT NOT NULL CHECK (visibility IN ('private','group','public')),
	group_id TEXT REFERENCES course_groups(id),
	owner_id TEXT REFERENCES users(id),
	path TEXT NOT NULL,
	status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active','removed')),
	created_at INTEGER NOT NULL
);
CREATE INDEX idx_datasets_owner ON datasets(owner_id);
CREATE INDEX idx_datasets_group ON datasets(group_id);

-- 平台会话（执行器跟踪；区别于 agent 会话文件）
CREATE TABLE sessions (
	id TEXT PRIMARY KEY,
	project_id TEXT REFERENCES projects(id),
	user_id TEXT REFERENCES users(id),
	container_id TEXT,
	status TEXT NOT NULL DEFAULT 'created' CHECK (status IN ('created','running','ended','error')),
	cwd TEXT,
	started_at INTEGER,
	ended_at INTEGER
);
CREATE INDEX idx_sessions_project ON sessions(project_id);
