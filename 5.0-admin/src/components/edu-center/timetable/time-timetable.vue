<script setup>
// 引入icon
import { DownOutlined, LeftOutlined, RightOutlined } from '@ant-design/icons-vue'
import dayjs from 'dayjs'

import { watch } from 'vue'

const displayArray = ref([
  'intentionCourse', // 意向课程
  'reference', // 推荐人
  'department', // 所属部门（仅在 type='dpt' 时显示）
  'channelCategory', // 渠道
  'channelStatus', // 渠道状态
  'channelType', // 渠道类型
  'subject', // 科目
])

// 模式选项
const modeOptions = {
  1: '标准模式',
  2: '纵向平铺',
  3: '横向平铺',
}
// 当前选中的模式
const currentMode = ref('1')

// 处理模式切换
function handleMenuClick({ key }) {
  currentMode.value = key
}

// 时间维度选项
const timeOptions = [
  { key: 'day', label: '日' },
  { key: 'week', label: '周' },
  { key: 'month', label: '月' },
]
// 当前选中的时间维度
const currentTime = ref('week')

// 当前的日期区间 - 默认设置为本周
const currentWeek = ref(dayjs())

// 监听时间维度变化
watch(currentTime, () => {
  // 切换时始终使用当前时间
  currentWeek.value = dayjs()
})

// 格式化日期显示
function formatDateRange(value) {
  if (!value)
    return ''

  switch (currentTime.value) {
    case 'day':
      return value.format('YYYY年MM月DD日')
    case 'week':
      const start = value.startOf('week')
      const end = value.endOf('week')

      if (start.year() === end.year() && start.month() === end.month()) {
        return `${start.format('YYYY年MM月DD日')} ~ ${end.format('DD日')}`
      }
      else if (start.year() === end.year()) {
        return `${start.format('YYYY年MM月DD日')} ~ ${end.format('MM月DD日')}`
      }
      else {
        return `${start.format('YYYY年MM月DD日')} ~ ${end.format('YYYY年MM月DD日')}`
      }
    case 'month':
      return value.format('YYYY年MM月')
    default:
      return ''
  }
}

// 处理前一个时间段
function handlePrev() {
  switch (currentTime.value) {
    case 'day':
      currentWeek.value = currentWeek.value.subtract(1, 'day')
      break
    case 'week':
      currentWeek.value = currentWeek.value.subtract(1, 'week')
      break
    case 'month':
      currentWeek.value = currentWeek.value.subtract(1, 'month')
      break
  }
}

// 处理后一个时间段
function handleNext() {
  switch (currentTime.value) {
    case 'day':
      currentWeek.value = currentWeek.value.add(1, 'day')
      break
    case 'week':
      currentWeek.value = currentWeek.value.add(1, 'week')
      break
    case 'month':
      currentWeek.value = currentWeek.value.add(1, 'month')
      break
  }
}
</script>

<template>
  <div class="filter-wrap  bg-white  pl-3 pr-3 rounded-4 rounded-lt-0 rounded-rt-0">
    <all-filter :display-array="displayArray" :is-show-search-stu-phonefilter="true" />
  </div>
  <div class="time-template mt2 bg-white  py3 px5 rounded-4">
    <div class="top-filter flex justify-between flex-items-center">
      <div>
        <a-dropdown>
          <template #overlay>
            <a-menu @click="handleMenuClick">
              <a-menu-item key="1">
                标准模式
              </a-menu-item>
              <a-menu-item key="2">
                纵向平铺
              </a-menu-item>
              <a-menu-item key="3">
                横向平铺
              </a-menu-item>
            </a-menu>
          </template>
          <a-button type="primary" class="font-800">
            {{ modeOptions[currentMode] }}
            <DownOutlined class="text-3" />
          </a-button>
        </a-dropdown>
      </div>
      <div class="time-selector flex-center">
        <a-radio-group v-model:value="currentTime" button-style="solid" size="small">
          <a-radio-button v-for="opt in timeOptions" :key="opt.key" :value="opt.key">
            {{ opt.label }}
          </a-radio-button>
        </a-radio-group>
        <div class="ml3 text-#0061ff font-800 text-5 flex-center">
          <a-popover trigger="hover">
            <template #content>
              {{ currentTime === 'day' ? '前一天' : currentTime === 'week' ? '上一周' : '上个月' }}
            </template>
            <span
              class="cursor-pointer text-3 text-#888 flex w6 h6 bg-#eee rounded-10 flex-center font-500 hover-text-#06f hover-bg-#e6f0ff"
              @click="handlePrev"
            >
              <LeftOutlined />
            </span>
          </a-popover>
          <span class="mx-2">
            <div class="relative cursor-pointer">{{ formatDateRange(currentWeek) }}
              <a-date-picker
                v-if="currentTime === 'day'"
                v-model:value="currentWeek"
                class="absolute top-0 left-0 right-0 bottom-0 z-10 opacity-0"
                :allow-clear="false"
                :bordered="false"
                :format="formatDateRange"
                style="cursor:pointer;"
              />
              <a-date-picker
                v-else-if="currentTime === 'week'"
                v-model:value="currentWeek"
                class="absolute top-0 left-0 right-0 bottom-0 z-10 opacity-0"
                picker="week"
                :allow-clear="false"
                :bordered="false"
                :format="formatDateRange"
                style="cursor:pointer;"
              />
              <a-date-picker
                v-else
                v-model:value="currentWeek"
                class="absolute top-0 left-0 right-0 bottom-0 z-10 opacity-0"
                picker="month"
                :allow-clear="false"
                :bordered="false"
                :format="formatDateRange"
                style="cursor:pointer;"
              />
            </div>
          </span>
          <a-popover trigger="hover">
            <template #content>
              {{ currentTime === 'day' ? '后一天' : currentTime === 'week' ? '下一周' : '下个月' }}
            </template>
            <span
              class="cursor-pointer text-3 text-#888 flex w6 h6 bg-#eee rounded-10 flex-center font-500 hover-text-#06f hover-bg-#e6f0ff"
              @click="handleNext"
            >
              <RightOutlined />
            </span>
          </a-popover>
        </div>
      </div>
      <a-space>
        <a-button type="primary">
          创建日程
        </a-button>
        <a-button>导出课表</a-button>
      </a-space>
    </div>
  </div>
</template>

<style scoped lang="less">
.time-selector {
  font-family: DINAlternate-Bold, DINAlternate;

  .ant-radio-button-wrapper {
    padding: 0 16px;
  }
}
</style>
