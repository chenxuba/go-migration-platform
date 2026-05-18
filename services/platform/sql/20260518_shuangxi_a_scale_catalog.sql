INSERT INTO sys_scale (
  scale_name, scale_code, category, scenario, age_range, current_version,
  item_count, domain_count, institution_count, month_usage, data_status, summary,
  execution_entry, api_package, sort, age_min_months, age_max_months,
  estimated_duration, duration_min_minutes, duration_max_minutes,
  create_time, update_time, del_flag
)
SELECT
  '双溪课程评量表A', 'SHUANGXI_A', '课程评量', '课程评估', '2岁-16岁', 'A-2012-doc',
  209, 7, 0, 0, '题库已入库；Pad静态测评工作台已接入；评分保存和报告待接入',
  '面向心智障碍儿童个别化教育课程的0-3级课程评量表，覆盖感官知觉、粗大动作、精细动作、生活自理、沟通、认知和社会技能。',
  'Pad /shuangxi-a-assessment', '', 4, 24, 192,
  '60-90分钟', 60, 90,
  NOW(), NOW(), 0
WHERE NOT EXISTS (
  SELECT 1 FROM sys_scale WHERE scale_code = 'SHUANGXI_A' AND del_flag = 0
);

UPDATE sys_scale
SET scale_name = '双溪课程评量表A',
    category = '课程评量',
    scenario = '课程评估',
    age_range = '2岁-16岁',
    age_min_months = 24,
    age_max_months = 192,
    estimated_duration = '60-90分钟',
    duration_min_minutes = 60,
    duration_max_minutes = 90,
    current_version = 'A-2012-doc',
    item_count = 209,
    domain_count = 7,
    data_status = '题库已入库；Pad静态测评工作台已接入；评分保存和报告待接入',
    summary = '面向心智障碍儿童个别化教育课程的0-3级课程评量表，覆盖感官知觉、粗大动作、精细动作、生活自理、沟通、认知和社会技能。',
    execution_entry = 'Pad /shuangxi-a-assessment',
    api_package = '',
    sort = 4,
    update_time = NOW()
WHERE scale_code = 'SHUANGXI_A' AND del_flag = 0;

INSERT INTO sys_scale_reference (scale_id, content, sort, create_time, update_time, del_flag, version)
SELECT s.id, '台北市双溪启智文教基金会《双溪心智障碍儿童个别化教育课程》。', 1, NOW(), NOW(), 0, 1
FROM sys_scale s
WHERE s.scale_code = 'SHUANGXI_A' AND s.del_flag = 0
  AND NOT EXISTS (
    SELECT 1 FROM sys_scale_reference r
    WHERE r.scale_id = s.id
      AND r.content = '台北市双溪启智文教基金会《双溪心智障碍儿童个别化教育课程》。'
      AND r.del_flag = 0
  );

INSERT INTO sys_scale_reference (scale_id, content, sort, create_time, update_time, del_flag, version)
SELECT s.id, '《双溪课程评量表.doc》结构化题库，0-3级评量标准，共209题。', 2, NOW(), NOW(), 0, 1
FROM sys_scale s
WHERE s.scale_code = 'SHUANGXI_A' AND s.del_flag = 0
  AND NOT EXISTS (
    SELECT 1 FROM sys_scale_reference r
    WHERE r.scale_id = s.id
      AND r.content = '《双溪课程评量表.doc》结构化题库，0-3级评量标准，共209题。'
      AND r.del_flag = 0
  );
