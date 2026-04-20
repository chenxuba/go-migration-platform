/**
 * 处理时间范围查询参数
 * @param {object} params - 原始查询参数对象
 * @param {object} fieldMappings - 字段映射配置对象
 * @returns {object} - 处理后的查询参数对象
 *
 * @example
 * const params = {
 *   createTime: ['2024-01-01', '2024-01-31'],
 *   birthday: ['2000-01-01', '2000-12-31']
 * };
 *
 * const mappings = {
 *   createTime: {
 *     begin: 'createTimeBegin',
 *     end: 'createTimeEnd'
 *   },
 *   birthday: {
 *     begin: 'birthdayBegin',
 *     end: 'birthdayEnd'
 *   }
 * };
 *
 * const result = handleDateRangeParams(params, mappings);
 */
export function handleDateRangeParams(params: Record<string, any>, fieldMappings: Record<string, { begin: string, end: string }>) {
  const newParams = { ...params }

  Object.entries(fieldMappings).forEach(([originalField, { begin, end }]) => {
    const rangeValue = newParams[originalField]

    if (rangeValue && Array.isArray(rangeValue) && rangeValue.length > 0) {
      // 设置开始和结束时间
      newParams[begin] = rangeValue[0]
      newParams[end] = rangeValue[1]
    }
    delete newParams[originalField]
    if (!rangeValue || !rangeValue.length) {
      delete newParams[begin]
      delete newParams[end]
    }
  })

  return newParams
}
