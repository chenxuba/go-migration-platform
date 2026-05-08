SET @has_assessment_draft_reuse_lookup_index := (
  SELECT COUNT(*)
  FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'assessment_draft'
    AND INDEX_NAME = 'idx_assessment_draft_reuse_lookup'
);
SET @ddl_assessment_draft_reuse_lookup_index := IF(
  @has_assessment_draft_reuse_lookup_index = 0,
  'ALTER TABLE assessment_draft ADD INDEX idx_assessment_draft_reuse_lookup (inst_id, student_id, assessment_code, assessment_date, submitted_record_id, del_flag, update_time, id)',
  'SELECT 1'
);
PREPARE stmt FROM @ddl_assessment_draft_reuse_lookup_index;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
