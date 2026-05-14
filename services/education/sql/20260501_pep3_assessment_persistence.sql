-- PEP-3断点续测与静态量表数据表。
-- 静态题库/分测验域/常模内容由 education-service 启动时从已整理 JSON 种子同步。

CREATE TABLE IF NOT EXISTS assessment_draft_item_score (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  inst_id BIGINT NOT NULL DEFAULT 0,
  draft_id BIGINT NOT NULL DEFAULT 0,
  item_no INT NOT NULL DEFAULT 0,
  score INT NOT NULL DEFAULT 0,
  create_id BIGINT NOT NULL DEFAULT 0,
  update_id BIGINT NOT NULL DEFAULT 0,
  create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  del_flag TINYINT(1) NOT NULL DEFAULT 0,
  UNIQUE KEY uk_assessment_draft_item_score (draft_id, item_no),
  KEY idx_assessment_draft_item_score_inst (inst_id, draft_id, item_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS assessment_draft_raw_score (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  inst_id BIGINT NOT NULL DEFAULT 0,
  draft_id BIGINT NOT NULL DEFAULT 0,
  scale_code VARCHAR(32) NOT NULL DEFAULT '',
  raw_score INT NOT NULL DEFAULT 0,
  create_id BIGINT NOT NULL DEFAULT 0,
  update_id BIGINT NOT NULL DEFAULT 0,
  create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  del_flag TINYINT(1) NOT NULL DEFAULT 0,
  UNIQUE KEY uk_assessment_draft_raw_score (draft_id, scale_code),
  KEY idx_assessment_draft_raw_score_inst (inst_id, draft_id, scale_code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS assessment_draft_item_record_value (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  inst_id BIGINT NOT NULL DEFAULT 0,
  draft_id BIGINT NOT NULL DEFAULT 0,
  item_no INT NOT NULL DEFAULT 0,
  field_key VARCHAR(100) NOT NULL DEFAULT '',
  value_json LONGTEXT NOT NULL,
  create_id BIGINT NOT NULL DEFAULT 0,
  update_id BIGINT NOT NULL DEFAULT 0,
  create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  del_flag TINYINT(1) NOT NULL DEFAULT 0,
  UNIQUE KEY uk_assessment_draft_item_record_value (draft_id, item_no, field_key),
  KEY idx_assessment_draft_item_record_inst (inst_id, draft_id, item_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS assessment_scale_dataset (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  scale_code VARCHAR(64) NOT NULL DEFAULT '',
  scale_version VARCHAR(64) NOT NULL DEFAULT '',
  data_status VARCHAR(1000) NOT NULL DEFAULT '',
  sources_json LONGTEXT NOT NULL,
  metadata_json LONGTEXT NULL,
  create_id BIGINT NOT NULL DEFAULT 0,
  update_id BIGINT NOT NULL DEFAULT 0,
  create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  del_flag TINYINT(1) NOT NULL DEFAULT 0,
  UNIQUE KEY uk_assessment_scale_dataset (scale_code, scale_version)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS assessment_scale_item (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  scale_code VARCHAR(64) NOT NULL DEFAULT '',
  scale_version VARCHAR(64) NOT NULL DEFAULT '',
  item_no INT NOT NULL DEFAULT 0,
  item_json LONGTEXT NOT NULL,
  create_id BIGINT NOT NULL DEFAULT 0,
  update_id BIGINT NOT NULL DEFAULT 0,
  create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  del_flag TINYINT(1) NOT NULL DEFAULT 0,
  UNIQUE KEY uk_assessment_scale_item (scale_code, scale_version, item_no),
  KEY idx_assessment_scale_item_version (scale_code, scale_version, item_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS assessment_scale_domain (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  scale_code VARCHAR(64) NOT NULL DEFAULT '',
  scale_version VARCHAR(64) NOT NULL DEFAULT '',
  domain_code VARCHAR(64) NOT NULL DEFAULT '',
  sort_no INT NOT NULL DEFAULT 0,
  domain_json LONGTEXT NOT NULL,
  create_id BIGINT NOT NULL DEFAULT 0,
  update_id BIGINT NOT NULL DEFAULT 0,
  create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  del_flag TINYINT(1) NOT NULL DEFAULT 0,
  UNIQUE KEY uk_assessment_scale_domain (scale_code, scale_version, domain_code),
  KEY idx_assessment_scale_domain_version (scale_code, scale_version, sort_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS assessment_scale_norm_record (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  scale_code VARCHAR(64) NOT NULL DEFAULT '',
  scale_version VARCHAR(64) NOT NULL DEFAULT '',
  record_key VARCHAR(128) NOT NULL DEFAULT '',
  sort_no INT NOT NULL DEFAULT 0,
  norm_json LONGTEXT NOT NULL,
  create_id BIGINT NOT NULL DEFAULT 0,
  update_id BIGINT NOT NULL DEFAULT 0,
  create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  del_flag TINYINT(1) NOT NULL DEFAULT 0,
  UNIQUE KEY uk_assessment_scale_norm_record (scale_code, scale_version, record_key),
  KEY idx_assessment_scale_norm_version (scale_code, scale_version, sort_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
