-- Roll back the departments table and its indexes.
DROP INDEX IF EXISTS department_status;
DROP INDEX IF EXISTS department_path;
DROP INDEX IF EXISTS department_parent_id;
DROP INDEX IF EXISTS department_code;
DROP TABLE IF EXISTS departments;
