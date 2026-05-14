UPDATE sys_scale
SET scale_name = '孤独症儿童发展评估表',
    update_time = NOW()
WHERE scale_code = 'AUTISMDEV'
  AND del_flag = 0
  AND scale_name = '孤独症儿童发展评估表（试行）';

UPDATE sys_scale_reference r
JOIN sys_scale s ON s.id = r.scale_id
SET r.content = REPLACE(r.content, '孤独症儿童发展评估表（试行）', '孤独症儿童发展评估表'),
    r.update_time = NOW()
WHERE s.scale_code = 'AUTISMDEV'
  AND s.del_flag = 0
  AND r.del_flag = 0
  AND r.content LIKE '%孤独症儿童发展评估表（试行）%';

UPDATE assessment_draft
SET assessment_name = '孤独症儿童发展评估表',
    update_time = NOW()
WHERE assessment_code = 'AUTISMDEV'
  AND del_flag = 0
  AND assessment_name = '孤独症儿童发展评估表（试行）';

UPDATE assessment_record
SET assessment_name = '孤独症儿童发展评估表',
    update_time = NOW()
WHERE assessment_code = 'AUTISMDEV'
  AND del_flag = 0
  AND assessment_name = '孤独症儿童发展评估表（试行）';
