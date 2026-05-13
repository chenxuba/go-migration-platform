INSERT INTO sys_scale (
  scale_name, scale_code, category, scenario, age_range, current_version,
  item_count, domain_count, institution_count, month_usage, data_status, summary,
  execution_entry, api_package, sort, age_min_months, age_max_months,
  estimated_duration, duration_min_minutes, duration_max_minutes,
  create_time, update_time, del_flag
)
SELECT
  '孤独症儿童发展评估表（试行）', 'AUTISMDEV', '标准化测评', '现场测评', '0岁-6岁', '2010-revised-trainer',
  493, 8, 0, 0, '题库、评分规则和教育端接口已串联；前端工作台待接入',
  '面向0岁至6岁孤独症儿童的发展评估表，覆盖感知觉、动作、语言沟通、认知、社会交往、生活自理及情绪行为。',
  '机构端 /teacherCenter/scale-library', '/api/v1/assessments/autismdev/*', 3, 0, 72,
  '60-120分钟', 60, 120,
  NOW(), NOW(), 0
WHERE NOT EXISTS (
  SELECT 1 FROM sys_scale WHERE scale_code = 'AUTISMDEV' AND del_flag = 0
);

UPDATE sys_scale
SET scale_name = '孤独症儿童发展评估表（试行）',
    category = '标准化测评',
    scenario = '现场测评',
    age_range = '0岁-6岁',
    age_min_months = 0,
    age_max_months = 72,
    estimated_duration = '60-120分钟',
    duration_min_minutes = 60,
    duration_max_minutes = 120,
    current_version = '2010-revised-trainer',
    item_count = 493,
    domain_count = 8,
    data_status = '题库、评分规则和教育端接口已串联；前端工作台待接入',
    summary = '面向0岁至6岁孤独症儿童的发展评估表，覆盖感知觉、动作、语言沟通、认知、社会交往、生活自理及情绪行为。',
    execution_entry = '机构端 /teacherCenter/scale-library',
    api_package = '/api/v1/assessments/autismdev/*',
    sort = 3,
    update_time = NOW()
WHERE scale_code = 'AUTISMDEV' AND del_flag = 0;

INSERT INTO sys_scale_reference (scale_id, content, sort, create_time, update_time, del_flag, version)
SELECT s.id, '中国残疾人联合会康复部《孤独症儿童发展评估表（试行）》使用手册及评估量表。', 1, NOW(), NOW(), 0, 1
FROM sys_scale s
WHERE s.scale_code = 'AUTISMDEV' AND s.del_flag = 0
  AND NOT EXISTS (
    SELECT 1 FROM sys_scale_reference r
    WHERE r.scale_id = s.id
      AND r.content = '中国残疾人联合会康复部《孤独症儿童发展评估表（试行）》使用手册及评估量表。'
      AND r.del_flag = 0
  );

INSERT INTO sys_scale_reference (scale_id, content, sort, create_time, update_time, del_flag, version)
SELECT s.id, '孤独症儿童发展评估表结构化题库、领域划分和 P/E/F/X、A/M/S 评分规则。', 2, NOW(), NOW(), 0, 1
FROM sys_scale s
WHERE s.scale_code = 'AUTISMDEV' AND s.del_flag = 0
  AND NOT EXISTS (
    SELECT 1 FROM sys_scale_reference r
    WHERE r.scale_id = s.id
      AND r.content = '孤独症儿童发展评估表结构化题库、领域划分和 P/E/F/X、A/M/S 评分规则。'
      AND r.del_flag = 0
  );
