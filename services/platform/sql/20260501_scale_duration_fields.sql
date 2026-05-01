SET @has_estimated_duration := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'sys_scale'
    AND COLUMN_NAME = 'estimated_duration'
);
SET @ddl_estimated_duration := IF(
  @has_estimated_duration = 0,
  'ALTER TABLE sys_scale ADD COLUMN estimated_duration VARCHAR(64) NOT NULL DEFAULT '''' AFTER age_range',
  'SELECT 1'
);
PREPARE stmt FROM @ddl_estimated_duration;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @has_duration_min_minutes := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'sys_scale'
    AND COLUMN_NAME = 'duration_min_minutes'
);
SET @ddl_duration_min_minutes := IF(
  @has_duration_min_minutes = 0,
  'ALTER TABLE sys_scale ADD COLUMN duration_min_minutes INT NOT NULL DEFAULT 0 AFTER estimated_duration',
  'SELECT 1'
);
PREPARE stmt FROM @ddl_duration_min_minutes;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @has_duration_max_minutes := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'sys_scale'
    AND COLUMN_NAME = 'duration_max_minutes'
);
SET @ddl_duration_max_minutes := IF(
  @has_duration_max_minutes = 0,
  'ALTER TABLE sys_scale ADD COLUMN duration_max_minutes INT NOT NULL DEFAULT 0 AFTER duration_min_minutes',
  'SELECT 1'
);
PREPARE stmt FROM @ddl_duration_max_minutes;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

UPDATE sys_scale
SET estimated_duration = '45-90分钟',
    duration_min_minutes = 45,
    duration_max_minutes = 90,
    update_time = NOW()
WHERE scale_code = 'PEP3'
  AND del_flag = 0;
