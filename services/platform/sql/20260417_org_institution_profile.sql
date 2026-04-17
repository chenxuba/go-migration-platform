CREATE TABLE IF NOT EXISTS org_institution_profile (
  id BIGINT NOT NULL AUTO_INCREMENT,
  institution_id BIGINT NOT NULL,
  organ_label VARCHAR(127) DEFAULT NULL,
  description TEXT DEFAULT NULL,
  business_time VARCHAR(255) DEFAULT NULL,
  video VARCHAR(2000) DEFAULT NULL,
  gallery_images JSON DEFAULT NULL,
  create_id BIGINT DEFAULT NULL,
  create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
  update_id BIGINT DEFAULT NULL,
  update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  del_flag TINYINT(1) DEFAULT 0,
  PRIMARY KEY (id),
  UNIQUE KEY uk_org_institution_profile_inst (institution_id),
  KEY idx_org_institution_profile_inst_del (institution_id, del_flag)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO org_institution_profile (
  institution_id,
  organ_label,
  description,
  business_time,
  video,
  gallery_images,
  create_id,
  create_time,
  update_id,
  update_time,
  del_flag
)
SELECT oi.id,
       NULLIF(TRIM(IFNULL(oi.organ_label, '')), ''),
       NULLIF(TRIM(IFNULL(oi.description, '')), ''),
       NULLIF(TRIM(IFNULL(oi.business_time, '')), ''),
       NULLIF(TRIM(IFNULL(oi.video, '')), ''),
       oi.inst_images,
       oi.create_id,
       COALESCE(oi.create_time, NOW()),
       oi.update_id,
       COALESCE(oi.update_time, NOW()),
       0
FROM org_institution oi
LEFT JOIN org_institution_profile oip ON oip.institution_id = oi.id AND oip.del_flag = 0
WHERE oi.del_flag = 0
  AND oip.id IS NULL
  AND (
    NULLIF(TRIM(IFNULL(oi.organ_label, '')), '') IS NOT NULL
    OR NULLIF(TRIM(IFNULL(oi.description, '')), '') IS NOT NULL
    OR NULLIF(TRIM(IFNULL(oi.business_time, '')), '') IS NOT NULL
    OR NULLIF(TRIM(IFNULL(oi.video, '')), '') IS NOT NULL
    OR oi.inst_images IS NOT NULL
  );
