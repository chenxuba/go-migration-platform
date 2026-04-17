<script setup>
// 引入icon
import { LeftOutlined, RightOutlined } from '@ant-design/icons-vue'
import { message } from 'ant-design-vue'
import dayjs from 'dayjs'

import { watch } from 'vue'
import CreateSchedulePopover from './create-schedule-popover.vue'

const displayArray = ref([
  'intentionCourse', // 意向课程
  'reference', // 推荐人
  'department', // 所属部门（仅在 type='dpt' 时显示）
  'channelCategory', // 渠道
  'channelStatus', // 渠道状态
  'channelType', // 渠道类型
  'subject', // 科目
])
// 当前选中的学生
const selectedStudent = ref('张小明')

// 处理学生选择变化
function handleStudentChange(value) {
  selectedStudent.value = value
}

// 添加处理单元格点击的方法
function handleCellClick(text, column, teacher) {
  // 如果没有选择学生，提示先选择学生
  if (!selectedStudent.value) {
    message.warning('请先选择学生')
    return
  }

  // 如果是空单元格，检查是否可以排课
  if (!text) {
    const available = isTimeSlotAvailable(text, column)
    if (available) {
      // 生成新的课程ID
      const newId = generateNewId()

      // 可以排课
      dataSource.value = dataSource.value.map((item) => {
        if (item.teacher === teacher) {
          return {
            ...item,
            [column.dataIndex]: {
              id: newId,
              studentName: selectedStudent.value,
              courseName: '待定', // 这里可以添加课程选择功能
              isNewScheduled: true, // 添加标记，表示这是新排的课
            },
          }
        }
        return item
      })
      message.success('排课成功！')
    }
    else {
      // 显示冲突警告
      message.warning(`${selectedStudent.value}在该时间段已有其他课程安排`)
    }
  }
}

// 生成新的课程ID
function generateNewId() {
  // 获取当前最大ID
  let maxId = 1000
  dataSource.value.forEach((teacher) => {
    for (let i = 1; i <= 12; i++) {
      const lesson = teacher[`lesson${i}`]
      if (lesson && lesson.id) {
        const id = Number.parseInt(lesson.id)
        if (id > maxId) {
          maxId = id
        }
      }
    }
  })
  // 返回新ID
  return (maxId + 1).toString()
}

// 检查时间段是否可用
function isTimeSlotAvailable(text, column) {
  // 如果格子为空，说明这个时间段是空闲的
  if (!text) {
    // 获取当前选中学生的ID
    const studentIds = getStudentIds(selectedStudent.value)

    // 检查这个学生是否已经在其他老师的这个时间段有课
    return !dataSource.value.some((teacher) => {
      const lesson = teacher[column.dataIndex]
      return lesson && (
        lesson.studentName === selectedStudent.value
        || (studentIds.length > 0 && studentIds.includes(lesson.id))
      )
    })
  }
  return false
}

// 获取学生的所有课程ID
function getStudentIds(studentName) {
  const ids = []
  dataSource.value.forEach((teacher) => {
    for (let i = 1; i <= 12; i++) {
      const lesson = teacher[`lesson${i}`]
      if (lesson && lesson.studentName === studentName && lesson.id) {
        ids.push(lesson.id)
      }
    }
  })
  return ids
}

// 获取单元格的类名
// 添加控制显示的响应式变量
const showConflicts = ref(true)
const showScheduled = ref(true)

// 修改获取单元格类名的方法
function getCellClass(text, column) {
  if (!text) {
    const available = isTimeSlotAvailable(text, column)
    return {
      'empty-lesson': true,
      'available-slot': available,
      'conflict-slot': !showConflicts.value ? false : !available,
    }
  }
  return {
    'scheduled-hidden': !showScheduled.value && !text.isNewScheduled, // 修改这里，新排的课不会被隐藏
  }
}
// 时间维度选项
const timeOptions = [
  { key: 'day', label: '日' },
  { key: 'week', label: '周' },
]
// 当前选中的时间维度
const currentTime = ref('day')

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

// 表格列定义
const columns = [
  {
    title: '教师',
    dataIndex: 'teacher',
    fixed: 'left',
    width: 135,
  },
  {
    title: '第一节课',
    time: '08:00-08:50',
    width: 135,
    dataIndex: 'lesson1',
  },
  {
    title: '第二节课',
    time: '09:00-09:50',
    width: 135,
    dataIndex: 'lesson2',
  },
  {
    title: '第三节课',
    time: '10:00-10:50',
    width: 135,
    dataIndex: 'lesson3',
  },
  {
    title: '第四节课',
    time: '11:00-11:50',
    width: 135,
    dataIndex: 'lesson4',
  },
  {
    title: '第五节课',
    time: '13:00-13:50',
    width: 135,
    dataIndex: 'lesson5',
  },
  {
    title: '第六节课',
    time: '14:00-14:50',
    dataIndex: 'lesson6',
    width: 135,
  },
  {
    title: '第七节课',
    time: '15:00-15:50',
    width: 135,
    dataIndex: 'lesson7',
  },
  {
    title: '第八节课',
    time: '16:00-16:50',
    width: 135,
    dataIndex: 'lesson8',
  },
  {
    title: '第九节课',
    time: '17:00-17:50',
    width: 135,
    dataIndex: 'lesson9',
  },
  {
    title: '第十节课',
    time: '18:00-18:50',
    width: 135,
    dataIndex: 'lesson10',
  },
  {
    title: '第十一节课',
    time: '19:00-19:50',
    width: 135,
    dataIndex: 'lesson11',
  },
  {
    title: '第十二节课',
    time: '20:00-20:50',
    width: 135,
    dataIndex: 'lesson12',
  },
]

// 表格数据
const dataSource = ref([
  {
    key: '1',
    teacher: '张老师',
    lesson1: { id: '1001', studentName: '李明', courseName: '语文' },
    lesson2: '',
    lesson3: { id: '1002', studentName: '王芳', courseName: '数学' },
    lesson4: { id: '1003', studentName: '赵强', courseName: '英语' },
    lesson5: { id: '1004', studentName: '刘洋', courseName: '物理' },
    lesson6: '',
    lesson7: { id: '1005', studentName: '张小明', courseName: '化学' },
    lesson8: { id: '1006', studentName: '陈静', courseName: '生物' },
    lesson9: '',
    lesson10: { id: '1007', studentName: '杨帆', courseName: '历史' },
    lesson11: '',
    lesson12: { id: '1008', studentName: '周婷', courseName: '地理' },
  },
  {
    key: '2',
    teacher: '李老师',
    lesson1: { id: '1009', studentName: '林涛', courseName: '语文' },
    lesson2: { id: '1010', studentName: '张小明', courseName: '语文' },
    lesson3: { id: '1011', studentName: '黄磊', courseName: '数学' },
    lesson4: { id: '1012', studentName: '张小明', courseName: '英语' },
    lesson5: { id: '1013', studentName: '孙宇', courseName: '物理' },
    lesson6: { id: '1014', studentName: '郑华', courseName: '化学' },
    lesson7: '',
    lesson8: { id: '1015', studentName: '吴凡', courseName: '生物' },
    lesson9: { id: '1016', studentName: '徐佳', courseName: '历史' },
    lesson10: '',
    lesson11: { id: '1017', studentName: '马超', courseName: '地理' },
    lesson12: '',
  },
  {
    key: '3',
    teacher: '王老师',
    lesson1: { id: '1018', studentName: '钱伟', courseName: '语文' },
    lesson2: { id: '1019', studentName: '孙艺', courseName: '数学' },
    lesson3: { id: '1020', studentName: '周杰', courseName: '英语' },
    lesson4: '',
    lesson5: { id: '1021', studentName: '吴琳', courseName: '物理' },
    lesson6: { id: '1022', studentName: '郑小红', courseName: '化学' },
    lesson7: { id: '1023', studentName: '王浩', courseName: '生物' },
    lesson8: '',
    lesson9: { id: '1024', studentName: '张小明', courseName: '历史' },
    lesson10: { id: '1025', studentName: '李娜', courseName: '地理' },
    lesson11: '',
    lesson12: { id: '1026', studentName: '刘志', courseName: '语文' },
  },
  {
    key: '4',
    teacher: '刘老师',
    lesson1: { id: '1027', studentName: '陈明', courseName: '语文' },
    lesson2: { id: '1028', studentName: '张伟', courseName: '数学' },
    lesson3: '',
    lesson4: { id: '1029', studentName: '王丽', courseName: '英语' },
    lesson5: { id: '1030', studentName: '李强', courseName: '物理' },
    lesson6: { id: '1031', studentName: '赵芳', courseName: '化学' },
    lesson7: '',
    lesson8: { id: '1032', studentName: '刘洋', courseName: '生物' },
    lesson9: { id: '1033', studentName: '陈刚', courseName: '历史' },
    lesson10: '',
    lesson11: { id: '1034', studentName: '杨丽', courseName: '地理' },
    lesson12: { id: '1035', studentName: '周杰', courseName: '语文' },
  },
  {
    key: '5',
    teacher: '陈老师',
    lesson1: '',
    lesson2: { id: '1036', studentName: '林小红', courseName: '数学' },
    lesson3: { id: '1037', studentName: '张强', courseName: '英语' },
    lesson4: { id: '1038', studentName: '王磊', courseName: '物理' },
    lesson5: '',
    lesson6: { id: '1039', studentName: '李明', courseName: '化学' },
    lesson7: { id: '1040', studentName: '赵丽', courseName: '生物' },
    lesson8: { id: '1041', studentName: '刘芳', courseName: '历史' },
    lesson9: '',
    lesson10: { id: '1042', studentName: '徐佳', courseName: '地理' },
    lesson11: { id: '1043', studentName: '陈浩', courseName: '语文' },
    lesson12: '',
  },
  {
    key: '6',
    teacher: '赵老师',
    lesson1: { id: '1044', studentName: '杨帆', courseName: '数学' },
    lesson2: '',
    lesson3: { id: '1045', studentName: '周婷', courseName: '英语' },
    lesson4: { id: '1046', studentName: '吴涛', courseName: '物理' },
    lesson5: { id: '1047', studentName: '郑明', courseName: '化学' },
    lesson6: '',
    lesson7: { id: '1048', studentName: '马丽', courseName: '生物' },
    lesson8: { id: '1049', studentName: '钱芳', courseName: '历史' },
    lesson9: { id: '1050', studentName: '孙强', courseName: '地理' },
    lesson10: '',
    lesson11: { id: '1051', studentName: '周杰', courseName: '语文' },
    lesson12: { id: '1052', studentName: '王芳', courseName: '数学' },
  },
  {
    key: '7',
    teacher: '孙老师',
    lesson1: { id: '1053', studentName: '李华', courseName: '英语' },
    lesson2: { id: '1054', studentName: '张伟', courseName: '物理' },
    lesson3: '',
    lesson4: { id: '1055', studentName: '王丽', courseName: '化学' },
    lesson5: { id: '1056', studentName: '刘强', courseName: '生物' },
    lesson6: { id: '1057', studentName: '陈芳', courseName: '历史' },
    lesson7: '',
    lesson8: { id: '1058', studentName: '赵明', courseName: '地理' },
    lesson9: { id: '1059', studentName: '杨洋', courseName: '语文' },
    lesson10: { id: '1060', studentName: '周刚', courseName: '数学' },
    lesson11: '',
    lesson12: { id: '1061', studentName: '吴丽', courseName: '英语' },
  },
  {
    key: '8',
    teacher: '周老师',
    lesson1: '',
    lesson2: { id: '1062', studentName: '郑伟', courseName: '物理' },
    lesson3: { id: '1063', studentName: '马芳', courseName: '化学' },
    lesson4: { id: '1064', studentName: '钱明', courseName: '生物' },
    lesson5: '',
    lesson6: { id: '1065', studentName: '孙丽', courseName: '历史' },
    lesson7: { id: '1066', studentName: '周强', courseName: '地理' },
    lesson8: { id: '1067', studentName: '吴芳', courseName: '语文' },
    lesson9: '',
    lesson10: { id: '1068', studentName: '郑洋', courseName: '数学' },
    lesson11: { id: '1069', studentName: '马明', courseName: '英语' },
    lesson12: '',
  },
  {
    key: '9',
    teacher: '吴老师',
    lesson1: { id: '1070', studentName: '钱丽', courseName: '化学' },
    lesson2: '',
    lesson3: { id: '1071', studentName: '孙芳', courseName: '生物' },
    lesson4: { id: '1072', studentName: '周明', courseName: '历史' },
    lesson5: { id: '1073', studentName: '吴强', courseName: '地理' },
    lesson6: '',
    lesson7: { id: '1074', studentName: '郑丽', courseName: '语文' },
    lesson8: { id: '1075', studentName: '马芳', courseName: '数学' },
    lesson9: { id: '1076', studentName: '钱明', courseName: '英语' },
    lesson10: '',
    lesson11: { id: '1077', studentName: '孙洋', courseName: '物理' },
    lesson12: { id: '1078', studentName: '周刚', courseName: '化学' },
  },
  {
    key: '10',
    teacher: '郑老师',
    lesson1: { id: '1079', studentName: '吴丽', courseName: '生物' },
    lesson2: { id: '1080', studentName: '郑明', courseName: '历史' },
    lesson3: '',
    lesson4: { id: '1081', studentName: '马芳', courseName: '地理' },
    lesson5: { id: '1082', studentName: '钱强', courseName: '语文' },
    lesson6: { id: '1083', studentName: '孙丽', courseName: '数学' },
    lesson7: '',
    lesson8: { id: '1084', studentName: '周芳', courseName: '英语' },
    lesson9: { id: '1085', studentName: '吴明', courseName: '物理' },
    lesson10: { id: '1086', studentName: '郑强', courseName: '化学' },
    lesson11: '',
    lesson12: { id: '1087', studentName: '马丽', courseName: '生物' },
  },
  {
    key: '11',
    teacher: '黄老师',
    lesson1: { id: '1088', studentName: '钱芳', courseName: '历史' },
    lesson2: { id: '1089', studentName: '孙明', courseName: '地理' },
    lesson3: { id: '1090', studentName: '周丽', courseName: '语文' },
    lesson4: '',
    lesson5: { id: '1091', studentName: '吴芳', courseName: '数学' },
    lesson6: { id: '1092', studentName: '郑明', courseName: '英语' },
    lesson7: { id: '1093', studentName: '马强', courseName: '物理' },
    lesson8: '',
    lesson9: { id: '1094', studentName: '钱丽', courseName: '化学' },
    lesson10: { id: '1095', studentName: '孙芳', courseName: '生物' },
    lesson11: '',
    lesson12: { id: '1096', studentName: '周明', courseName: '历史' },
  },
  {
    key: '12',
    teacher: '徐老师',
    lesson1: '',
    lesson2: { id: '1097', studentName: '吴强', courseName: '地理' },
    lesson3: { id: '1098', studentName: '郑丽', courseName: '语文' },
    lesson4: { id: '1099', studentName: '马芳', courseName: '数学' },
    lesson5: { id: '1100', studentName: '钱明', courseName: '英语' },
    lesson6: '',
    lesson7: { id: '1101', studentName: '孙洋', courseName: '物理' },
    lesson8: { id: '1102', studentName: '周刚', courseName: '化学' },
    lesson9: { id: '1103', studentName: '吴丽', courseName: '生物' },
    lesson10: { id: '1104', studentName: '郑明', courseName: '历史' },
    lesson11: '',
    lesson12: { id: '1105', studentName: '马芳', courseName: '地理' },
  },
])
// 添加计算属性计算表格高度
const tableHeight = computed(() => {
  // 获取视窗高度
  const windowHeight = window.innerHeight
  // 减去其他元素的高度（头部过滤器、顶部操作栏等）
  // 这里的数值需要根据实际情况调整
  const otherHeight = 390 // 预估其他元素总高度
  return windowHeight - otherHeight
})
</script>

<template>
  <div class="filter-wrap  bg-white  pl-3 pr-3 rounded-4 rounded-lt-0 rounded-rt-0">
    <all-filter :display-array="displayArray" :is-show-search-stu-phonefilter="true" />
  </div>
  <div class="time-template mt2 bg-white  py3 px5 rounded-4">
    <div class="top-filter flex justify-between flex-items-center">
      <div>
        <div class="flex items-center">
          <span class="mr-2">选择学生：</span>
          <a-select
            v-model:value="selectedStudent" style="width: 200px" placeholder="请选择学生"
            @change="handleStudentChange"
          >
            <a-select-option value="李明">
              李明
            </a-select-option>
            <a-select-option value="张小明">
              张小明
            </a-select-option>
            <a-select-option value="孙强">
              孙强
            </a-select-option>
          </a-select>
          <div class="ml-4 flex items-center">
            <a-checkbox v-model:checked="showConflicts" class="mr-4 flex items-center">
              <div class="flex items-center">
                <div class="w-4 h-4 bg-#ffe6e6 mr-1" />
                <span>显示冲突时间</span>
              </div>
            </a-checkbox>
            <a-checkbox v-model:checked="showScheduled" class="mr-4 items-center">
              <div class="flex items-center">
                <div class="w-4 h-4 bg-#06f border border-solid border-#06f mr-1" />
                <span>显示已排课程</span>
              </div>
            </a-checkbox>
            <div class="w-4 h-4 bg-#e6ffe6 mr-1" />
            <span>可排课时间</span>
          </div>
        </div>
      </div>
      <div class="time-selector flex-center flex-1">
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
                v-if="currentTime === 'day'" v-model:value="currentWeek"
                class="absolute top-0 left-0 right-0 bottom-0 z-10 opacity-0" :allow-clear="false" :bordered="false" :format="formatDateRange"
                style="cursor:pointer;"
              />
              <a-date-picker
                v-else-if="currentTime === 'week'"
                v-model:value="currentWeek" class="absolute top-0 left-0 right-0 bottom-0 z-10 opacity-0" picker="week"
                :allow-clear="false" :bordered="false" :format="formatDateRange" style="cursor:pointer;"
              />
              <a-date-picker
                v-else v-model:value="currentWeek"
                class="absolute top-0 left-0 right-0 bottom-0 z-10 opacity-0" picker="month" :allow-clear="false" :bordered="false"
                :format="formatDateRange" style="cursor:pointer;"
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
        <create-schedule-popover />
        <a-button>导出课表</a-button>
      </a-space>
    </div>
    <div class="center-content">
      <a-table
        :columns="columns" :data-source="dataSource" :pagination="false" :scroll="{ x: 1560 }"
        :sticky="{ offsetHeader: 100 }" bordered
      >
        <template #headerCell="{ column }">
          <template v-if="column.time">
            <div class="py1 whitespace-nowrap">
              <div>{{ column.title }}</div>
              <div class="text-12px text-#666">
                {{ column.time }}
              </div>
            </div>
          </template>
          <template v-else>
            {{ column.title }}
          </template>
        </template>
        <template #bodyCell="{ column, text, record }">
          <template v-if="column.dataIndex === 'teacher'">
            <div>{{ text }}</div>
          </template>
          <template v-else>
            <div
              :class="getCellClass(text, column)" class=" cursor-pointer relative"
              @click="handleCellClick(text, column, record.teacher)"
            >
              <template v-if="text">
                <div class="con">
                  <div class="t">
                    {{ text.studentName }}
                    <div v-if="text.isNewScheduled" class="scheduled-badge">
                      新
                    </div>
                  </div>
                  <div class="flex flex-col  flex-items-start pl2 pt1">
                    <div class="text-12px text-#666 courseName">
                      {{ text.courseName }}
                    </div>
                    <div class="name">
                      1v1
                    </div>
                  </div>
                </div>
              </template>
            </div>
          </template>
        </template>
      </a-table>
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

.center-content {
  margin-top: 16px;

  :deep(.ant-table-sticky-scroll-bar) {
    display: none;
  }

  :deep(.ant-table-cell) {
    padding: 2px;
    height: 40px;
    min-height: 40px;
  }

  :deep(.ant-table-cell) {
    text-align: center;
  }

  // 修改教师列的背景色样式
  :deep(.ant-table-cell.ant-table-cell-fix-left) {
    background-color: #fafafa !important;
  }

  // 添加行悬停效果
  :deep(.ant-table-tbody > tr:hover > td) {
    background-color: #dcecff !important; // 使用浅蓝色作为悬停背景色
    transition: background-color 0.3s; // 添加过渡效果
    cursor: pointer;
  }

  .scheduled-hidden {
    opacity: 0;
    transition: opacity 0.3s;
  }

  .empty-lesson {
    min-height: 40px;
    width: 100%;
    padding: 8px 0;
    height: 70px;
  }

  .available-slot {
    background-color: #e6ffe6; // 绿色背景表示可用
  }

  .conflict-slot {
    background-color: #ffe6e6; // 红色背景表示冲突
    transition: background-color 0.3s;
  }

  .scheduled-badge {
    position: absolute;
    top: 1px;
    right: 1px;
    bottom: 1px;
    background-color: #00000080;
    color: #fff;
    font-size: 12px;
    width: 20px;
    height: 20px;
    border-radius: 4px;
    display: flex;
    align-items: center;
    justify-content: center;
  }
  .con{
    background: rgba(78, 109, 255, 0.12);
    border-radius: 4px;
    height: 70px;
    font-weight: 400;
    position: relative;
    .t{
      display: flex;
    justify-content: start;
    padding-left: 8px;
      background: rgb(78, 109, 255);
      color: #fff;
      font-weight: 500;
      font-size: 14px;
      border-radius: 4px 4px 0 0;
      position: relative;
    }
    .name{
      font-size: 12px;
      color: #002cfd;
      font-weight: 500;

    }
    .courseName{
      color: rgb(0, 44, 253);
      font-size: 12px;
     }
  }
}
</style>
