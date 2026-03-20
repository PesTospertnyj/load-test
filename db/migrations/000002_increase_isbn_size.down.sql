-- Revert ISBN column size back to original VARCHAR(20)
ALTER TABLE books ALTER COLUMN isbn TYPE VARCHAR(20);
