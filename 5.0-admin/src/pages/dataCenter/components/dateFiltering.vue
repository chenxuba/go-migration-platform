<script setup>
import dayjs from 'dayjs'

const emit = defineEmits(['change'])
const dateType = ref('week')
const date = ref([])
const isSelecting = ref(false)
const currentTime = ref(dayjs().format('YYYY-MM-DD HH:mm:ss'))

// 更新当前时间
function updateCurrentTime() {
  currentTime.value = dayjs().format('HH:mm:ss')
}

// 组件挂载时更新当前时间
onMounted(() => {
  updateCurrentTime()
})

// 根据日期类型和选择的日期，转换为年月日格式的日期范围
function convertToDateRange(type, selectedDate) {
  if (!selectedDate || selectedDate.length !== 2 || !selectedDate[0] || !selectedDate[1]) {
    return []
  }

  const [start, end] = selectedDate
  let startDate, endDate

  switch (type) {
    case 'day':
      // 日期类型为天时，不需要特殊处理
      return selectedDate
    case 'week':
      // 周的第一天和最后一天
      startDate = dayjs(start).startOf('week')
      endDate = dayjs(end).endOf('week')
      break
    case 'month':
      // 月的第一天和最后一天
      startDate = dayjs(start).startOf('month')
      endDate = dayjs(end).endOf('month')
      break
    case 'quarter':
      // 季度的第一天和最后一天
      startDate = dayjs(start).startOf('quarter')
      endDate = dayjs(end).endOf('quarter')
      break
    default:
      return selectedDate
  }

  return [startDate.format('YYYY-MM-DD'), endDate.format('YYYY-MM-DD')]
}

// 切换日期类型时，重置日期
function handleChange(value) {
  date.value = []
}

// 监听日期变化
watch(date, (newValue) => {
  if (newValue && newValue.length === 2 && !isSelecting.value) {
    // 设置选择中状态，防止循环更新
    isSelecting.value = true

    // 转换日期范围
    const dateRange = convertToDateRange(dateType.value, newValue)

    // 更新UI显示的日期范围
    if (dateRange.length === 2) {
      date.value = dateRange
    }

    // 触发change事件
    emit('change', {
      dateType: dateType.value,
      date: dateRange,
    })

    // 重置选择状态
    setTimeout(() => {
      isSelecting.value = false
    }, 0)
  }
})

// 监听日期类型变化
watch(dateType, () => {
  // 如果有选中的日期，则更新日期范围
  if (date.value && date.value.length === 2) {
    const dateRange = convertToDateRange(dateType.value, date.value)
    date.value = dateRange
  }
})
</script>

<template>
  <div class="date-filtering">
    <div class="ml-2">
      <a-select v-model:value="dateType" class="w-90px" @change="handleChange">
        <a-select-option value="day">
          按天
        </a-select-option>
        <a-select-option value="week">
          按周
        </a-select-option>
        <a-select-option value="month">
          按月
        </a-select-option>
        <a-select-option value="quarter">
          按季
        </a-select-option>
      </a-select>
      <a-range-picker v-model:value="date" value-format="YYYY-MM-DD" format="YYYY-MM-DD" :picker="dateType" />
      <span class="text-gray-400 font-size-12px ml-4">数据更新于：今天 {{ currentTime }}</span>
    </div>
  </div>
</template>

<style lang="less" scoped>
.date-filtering {
  background: #fff;
  margin-top: 8px;
  padding: 12px;
  border-radius: 12px;

  :deep(.ant-select-selector) {
    border-top-right-radius: 0;
    border-bottom-right-radius: 0;
    border-right: none;
  }

  :deep(.ant-picker) {
    border-top-left-radius: 0;
    border-bottom-left-radius: 0;
  }
}
</style>
