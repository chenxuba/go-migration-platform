-- PEP-3 儿童表现记录选项绑定值迁移为数字序号。
-- 本迁移按产品要求清空旧 PEP-3 测评记录，不做旧 value 兼容映射。
-- 执行后重启 education-service；启动时会重新同步 PEP-3 题库记录字段。

DELETE FROM assessment_caregiver_invite
WHERE draft_id IN (
  SELECT id FROM assessment_draft WHERE assessment_code = 'PEP3'
)
OR record_id IN (
  SELECT id FROM assessment_record WHERE assessment_code = 'PEP3'
);

DELETE FROM assessment_draft_item_record_value
WHERE draft_id IN (
  SELECT id FROM assessment_draft WHERE assessment_code = 'PEP3'
);

DELETE FROM assessment_draft_item_score
WHERE draft_id IN (
  SELECT id FROM assessment_draft WHERE assessment_code = 'PEP3'
);

DELETE FROM assessment_draft_raw_score
WHERE draft_id IN (
  SELECT id FROM assessment_draft WHERE assessment_code = 'PEP3'
);

DELETE FROM assessment_draft
WHERE assessment_code = 'PEP3';

DELETE FROM assessment_record
WHERE assessment_code = 'PEP3';

DELETE FROM assessment_scale_item_record_field
WHERE scale_code = 'PEP3';
