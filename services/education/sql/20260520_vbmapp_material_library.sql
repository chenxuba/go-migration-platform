-- VB-MAPP素材库。平台默认素材由 docs/vbmapp/response-material-profiles.json 初始化；
-- 后续机构素材和学生偏好可在 library_scope / inst_id 维度扩展。

CREATE TABLE IF NOT EXISTS vbmapp_material_profile (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  scale_version VARCHAR(64) NOT NULL DEFAULT '',
  library_scope VARCHAR(16) NOT NULL DEFAULT 'platform',
  inst_id BIGINT NOT NULL DEFAULT 0,
  profile_id VARCHAR(100) NOT NULL DEFAULT '',
  label VARCHAR(255) NOT NULL DEFAULT '',
  source_logic TEXT NOT NULL,
  suggested_types_json LONGTEXT NOT NULL,
  preparation_checks_json LONGTEXT NOT NULL,
  sort_no INT NOT NULL DEFAULT 0,
  status VARCHAR(16) NOT NULL DEFAULT 'active',
  create_id BIGINT NOT NULL DEFAULT 0,
  update_id BIGINT NOT NULL DEFAULT 0,
  create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  del_flag TINYINT(1) NOT NULL DEFAULT 0,
  UNIQUE KEY uk_vbmapp_material_profile (scale_version, library_scope, inst_id, profile_id),
  KEY idx_vbmapp_material_profile_scope (scale_version, library_scope, inst_id, status, del_flag)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS vbmapp_material_item (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  scale_version VARCHAR(64) NOT NULL DEFAULT '',
  library_scope VARCHAR(16) NOT NULL DEFAULT 'platform',
  inst_id BIGINT NOT NULL DEFAULT 0,
  profile_id VARCHAR(100) NOT NULL DEFAULT '',
  material_code VARCHAR(100) NOT NULL DEFAULT '',
  material_name VARCHAR(255) NOT NULL DEFAULT '',
  material_type VARCHAR(100) NOT NULL DEFAULT '',
  sort_no INT NOT NULL DEFAULT 0,
  status VARCHAR(16) NOT NULL DEFAULT 'active',
  create_id BIGINT NOT NULL DEFAULT 0,
  update_id BIGINT NOT NULL DEFAULT 0,
  create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  del_flag TINYINT(1) NOT NULL DEFAULT 0,
  UNIQUE KEY uk_vbmapp_material_item (scale_version, library_scope, inst_id, profile_id, material_code),
  KEY idx_vbmapp_material_item_profile (scale_version, library_scope, inst_id, profile_id, status, del_flag),
  KEY idx_vbmapp_material_item_name (scale_version, material_name, status, del_flag)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
