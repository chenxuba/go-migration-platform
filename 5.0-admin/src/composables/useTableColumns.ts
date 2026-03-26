// composables/useTableColumns.js
import { computed, ref, watch } from 'vue'

export function useTableColumns(options) {
  const { storageKey, allColumns, excludeKeys = [], defaultSelectedKeys = [] } = options

  // 计算所有有效的列键（包括动态列）
  const keysArray = computed(() => {
    const staticKeys = allColumns.value
      .map(column => column?.key)
      .filter(key => typeof key !== 'undefined')

    // 获取动态列的键
    const dynamicKeys = allColumns.value
      .filter(col => col.isDynamic) // 假设我们用 isDynamic 标记动态列
      .map(col => col.key)
    return [...staticKeys, ...dynamicKeys]
  })

  // 从 localStorage 获取保存的选择状态
  const getSavedSelection = () => {
    try {
      const savedSelected = localStorage.getItem(storageKey)
      return savedSelected ? JSON.parse(savedSelected) : defaultSelectedKeys
    }
    catch (e) {
      console.error('Error parsing saved column selection:', e)
      return []
    }
  }

  // 初始化选中值（包括有效的存储值和必选列）
  const savedSelectedArray = getSavedSelection()
  const validSavedSelected = savedSelectedArray.filter(key =>
    keysArray.value.includes(key),
  )

  // 获取所有必选列的键
  const requiredKeys = computed(() =>
    allColumns.value
      .filter(col => col.required)
      .map(col => col.key),
  )

  // 初始选中值：包括保存的有效选择和必选列
  const initialSelectedValues = validSavedSelected.length > 0
    ? Array.from(new Set([...validSavedSelected, ...requiredKeys.value]))
    : keysArray.value

  // 响应式选中值
  const selectedValues = ref(initialSelectedValues)

  // 生成可选项（排除指定列）
  const columnOptions = computed(() =>
    allColumns.value
      .filter(col => !excludeKeys.includes(col.key))
      .map(col => ({
        id: col.key,
        value: col.title,
        disabled: col.required,
      })),
  )

  // 处理后的列配置
  const filteredColumns = computed(() => {
    const requiredColumns = allColumns.value.filter(col => col.required)
    const excludedColumns = allColumns.value.filter(col =>
      excludeKeys.includes(col.key),
    )

    const optionalColumns = allColumns.value.filter(
      col =>
        selectedValues.value.includes(col.key)
        && !col.required
        && !excludeKeys.includes(col.key),
    )

    // 合并逻辑（保持顺序）
    return [
      ...requiredColumns.filter(col => col.fixed === 'left'),
      ...optionalColumns,
      ...requiredColumns.filter(col => col.fixed === 'right'),
      ...excludedColumns.filter(col => col.fixed === 'right'),
    ]
  })

  // 监听选中值变化，确保必选字段始终被选中
  watch(selectedValues, (newVal) => {
    const allRequired = requiredKeys.value

    if (!allRequired.every(k => newVal.includes(k))) {
      selectedValues.value = Array.from(new Set([...newVal, ...allRequired]))
    }

    // 保存到 localStorage
    try {
      localStorage.setItem(storageKey, JSON.stringify(newVal))
    }
    catch (e) {
      console.error('Error saving column selection:', e)
    }
  }, { deep: true })

  // 监听列配置变化，更新选中状态
  watch(() => allColumns.value, (newColumns) => {
    const savedSelection = getSavedSelection()
    const newKeys = newColumns.map(col => col.key)

    // 合并现有选择和新的必选列
    const newRequired = newColumns
      .filter(col => col.required)
      .map(col => col.key)

    // 更新选中值，保留已保存的选择并添加新的必选列
    selectedValues.value = Array.from(new Set([
      ...savedSelection.filter(key => newKeys.includes(key)),
      ...newRequired,
    ]))
  }, { deep: true })

  // 计算表格总宽度
  const totalWidth = computed(() =>
    filteredColumns.value.reduce((acc, col) => acc + (col.width || 0), 0),
  )

  return {
    selectedValues,
    columnOptions,
    filteredColumns,
    totalWidth,
  }
}
