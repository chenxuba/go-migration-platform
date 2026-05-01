-- PEP-3 题库儿童表现记录选项绑定值迁移为数字序号。
-- 平台题库静态记录字段直接清空，重启 platform-service 后会按代码模板重新同步。

DELETE FROM assessment_scale_item_record_field
WHERE scale_code = 'PEP3';
