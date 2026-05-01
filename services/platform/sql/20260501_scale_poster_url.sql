SET @has_poster_url := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'sys_scale'
    AND COLUMN_NAME = 'poster_url'
);
SET @ddl_poster_url := IF(
  @has_poster_url = 0,
  'ALTER TABLE sys_scale ADD COLUMN poster_url VARCHAR(500) DEFAULT NULL AFTER summary',
  'SELECT 1'
);
PREPARE stmt FROM @ddl_poster_url;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
