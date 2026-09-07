-- Create the users table for the relay service. It references departments
-- so that users can be associated with the school organization tree.
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    username VARCHAR(50) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(255),
    phone VARCHAR(20),
    role VARCHAR(20) NOT NULL DEFAULT 'student' CHECK (role IN ('student', 'teacher', 'staff')),
    gender VARCHAR(10),
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'banned')),
    department_id UUID,
    CONSTRAINT users_department_fk FOREIGN KEY (department_id) REFERENCES departments(id) ON DELETE SET NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS user_username ON users(username);
CREATE UNIQUE INDEX IF NOT EXISTS user_email ON users(email);
CREATE UNIQUE INDEX IF NOT EXISTS user_phone ON users(phone);
CREATE INDEX IF NOT EXISTS user_department_id ON users(department_id);
CREATE INDEX IF NOT EXISTS user_status ON users(status);
