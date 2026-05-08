CREATE TABLE IF NOT EXISTS assessment_report_interpretation (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  inst_id BIGINT NOT NULL DEFAULT 0,
  record_id BIGINT NOT NULL DEFAULT 0,
  assessment_code VARCHAR(64) NOT NULL DEFAULT '',
  source_hash VARCHAR(64) NOT NULL DEFAULT '',
  content_json LONGTEXT NOT NULL,
  model VARCHAR(100) NOT NULL DEFAULT '',
  generated_by VARCHAR(32) NOT NULL DEFAULT '',
  create_id BIGINT NOT NULL DEFAULT 0,
  update_id BIGINT NOT NULL DEFAULT 0,
  create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  del_flag TINYINT(1) NOT NULL DEFAULT 0,
  UNIQUE KEY uk_assessment_report_interpretation_record (inst_id, record_id, assessment_code),
  KEY idx_assessment_report_interpretation_update (inst_id, assessment_code, update_time)
);
