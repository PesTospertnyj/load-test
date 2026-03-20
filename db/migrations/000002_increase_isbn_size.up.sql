-- Increase ISBN column size to accommodate longer ISBN values
ALTER TABLE books ALTER COLUMN isbn TYPE VARCHAR(100);
