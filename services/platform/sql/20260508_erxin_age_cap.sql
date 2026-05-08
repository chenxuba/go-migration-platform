UPDATE sys_scale
SET age_range = '0岁-6岁',
    age_min_months = 0,
    age_max_months = 72,
    update_time = NOW()
WHERE scale_code = 'ERXIN2'
  AND del_flag = 0;
