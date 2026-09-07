-- 004_session_permission.sql — 会话权限模式(request_approval / full_access)
-- 默认 full_access;项目会话由服务层在创建时置为 request_approval
ALTER TABLE sessions ADD COLUMN permission_mode TEXT NOT NULL DEFAULT 'full_access' CHECK (permission_mode IN ('request_approval','full_access'));
