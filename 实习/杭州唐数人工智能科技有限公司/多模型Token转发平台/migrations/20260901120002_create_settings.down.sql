-- Roll back the settings table and its indexes.
DROP INDEX IF EXISTS setting_key;
DROP TABLE IF EXISTS settings;
