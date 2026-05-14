SET @has_assessment_scale_dataset_metadata := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'assessment_scale_dataset'
    AND COLUMN_NAME = 'metadata_json'
);
SET @ddl_assessment_scale_dataset_metadata := IF(
  @has_assessment_scale_dataset_metadata = 0,
  'ALTER TABLE assessment_scale_dataset ADD COLUMN metadata_json LONGTEXT NULL AFTER sources_json',
  'SELECT 1'
);
PREPARE stmt FROM @ddl_assessment_scale_dataset_metadata;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
