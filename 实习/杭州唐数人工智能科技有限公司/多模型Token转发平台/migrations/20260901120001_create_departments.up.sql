-- Create the departments table with a self-referencing foreign key to support
-- a materialized-path tree structure. This is the first migration because
-- the users table references it.
CREATE TABLE IF NOT EXISTS departments (
    id UUID PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    name VARCHAR(100) NOT NULL,
    code VARCHAR(50) NOT NULL,
    level INTEGER NOT NULL DEFAULT 0,
    path VARCHAR(2000) NOT NULL DEFAULT '',
    sort_order INTEGER NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
    parent_id UUID,
    CONSTRAINT departments_parent_fk FOREIGN KEY (parent_id) REFERENCES departments(id) ON DELETE SET NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS department_code ON departments(code);
CREATE INDEX IF NOT EXISTS department_parent_id ON departments(parent_id);
CREATE INDEX IF NOT EXISTS department_path ON departments(path);
CREATE INDEX IF NOT EXISTS department_status ON departments(status);
