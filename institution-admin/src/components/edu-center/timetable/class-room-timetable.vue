<script setup>
// 引入icon
import { LeftOutlined, RightOutlined } from '@ant-design/icons-vue'
import { message } from 'ant-design-vue'
import { computed, ref, watch } from 'vue'
import dayjs from 'dayjs'
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

// 当前选中的学生（改为数组）
const selectedStudents = ref([])

// 处理学生选择变化
function handleStudentChange(value) {
  selectedStudents.value = value
}

// 当前选中的组别
const currentGroup = ref('感统组')
const groupOptions = [
  { key: '感统组', label: '感统组' },
  { key: '言语组', label: '言语组' },
]

// 根据当前选中的组别获取课表数据
const currentDataSource = computed(() => {
  return currentGroup.value === '感统组' ? dataSourceSensory.value : dataSourceSpeech.value
})

// 根据当前选中的组别获取课表列定义
const currentColumns = computed(() => {
  return currentGroup.value === '感统组' ? columnsSensory : columnsSpeech
})

// 添加处理单元格点击的方法
function handleCellClick(text, column, teacher) {
  // 如果没有选择学生，提示先选择学生
  if (!text && !selectedStudents.value.length) {
    message.warning('请先选择学生')
    return
  }

  // 如果是空单元格，检查是否可以排课
  if (!text) {
    // 检查所有选中学生是否都可以在这个时间段排课
    const availableResults = selectedStudents.value.map((student) => {
      return {
        student,
        available: isTimeSlotAvailable(text, column, student),
        conflictInfo: getConflictInfo(student, column),
      }
    })

    const unavailableStudents = availableResults.filter(result => !result.available)

    if (unavailableStudents.length > 0) {
      // 显示所有冲突学生的信息
      const conflictMessages = unavailableStudents.map((result) => {
        return result.conflictInfo
      })
      // message.warning(conflictMessages.join('\n'))
      return
    }

    // 所有学生都可以排课，生成新的课程ID
    const newId = generateNewId()

    // 可以排课
    if (currentGroup.value === '感统组') {
      dataSourceSensory.value = dataSourceSensory.value.map((item) => {
        if (item.teacherId === teacher.teacherId) {
          return {
            ...item,
            [column.dataIndex]: {
              id: newId,
              studentName: selectedStudents.value.join('、'), // 使用顿号连接多个学生名字
              courseName: '待定',
              isNewScheduled: true,
            },
          }
        }
        return item
      })
    }
    else {
      dataSourceSpeech.value = dataSourceSpeech.value.map((item) => {
        if (item.teacherId === teacher.teacherId) {
          return {
            ...item,
            [column.dataIndex]: {
              id: newId,
              studentName: selectedStudents.value.join('、'), // 使用顿号连接多个学生名字
              courseName: '待定',
              isNewScheduled: true,
            },
          }
        }
        return item
      })
    }
    message.success('排课成功！')
  }
}

// 生成新的课程ID
function generateNewId() {
  // 获取当前最大ID
  let maxId = 1000

  // 检查感统组数据
  dataSourceSensory.value.forEach((teacher) => {
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

  // 检查言语组数据
  dataSourceSpeech.value.forEach((teacher) => {
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

// 检查时间段是否可用（跨组冲突检测）
function isTimeSlotAvailable(text, column, studentName) {
  // 如果格子为空，说明这个时间段是空闲的
  if (!text) {
    // 获取当前选中学生的所有课程信息
    const studentLessons = getStudentLessons(studentName)

    // 获取当前时间段的实际时间范围
    const currentTimeRange = getTimeRangeForColumn(column, currentGroup.value)

    // 检查是否与学生的其他课程时间冲突
    return !studentLessons.some((lesson) => {
      // 检查时间是否重叠
      return isTimeOverlap(currentTimeRange, lesson.timeRange)
    })
  }
  return false
}

// 获取学生的所有课程信息（包括两个组）
function getStudentLessons(studentName) {
  const lessons = []

  // 检查感统组数据
  dataSourceSensory.value.forEach((teacher) => {
    for (let i = 1; i <= 12; i++) {
      const lesson = teacher[`lesson${i}`]
      // 修改判断逻辑，支持多人课程
      if (lesson && lesson.studentName) {
        const students = lesson.studentName.split('、')
        if (students.includes(studentName)) {
          lessons.push({
            id: lesson.id,
            group: '感统组',
            teacherId: teacher.teacherId,
            lessonIndex: i,
            timeRange: getTimeRangeForColumn({ dataIndex: `lesson${i}` }, '感统组'),
          })
        }
      }
    }
  })

  // 检查言语组数据
  dataSourceSpeech.value.forEach((teacher) => {
    for (let i = 1; i <= 12; i++) {
      const lesson = teacher[`lesson${i}`]
      // 修改判断逻辑，支持多人课程
      if (lesson && lesson.studentName) {
        const students = lesson.studentName.split('、')
        if (students.includes(studentName)) {
          lessons.push({
            id: lesson.id,
            group: '言语组',
            teacherId: teacher.teacherId,
            lessonIndex: i,
            timeRange: getTimeRangeForColumn({ dataIndex: `lesson${i}` }, '言语组'),
          })
        }
      }
    }
  })

  return lessons
}

// 获取指定列和组别的时间范围
function getTimeRangeForColumn(column, group) {
  const columns = group === '感统组' ? columnsSensory : columnsSpeech
  const targetColumn = columns.find(col => col.dataIndex === column.dataIndex)

  if (targetColumn && targetColumn.time) {
    const [start, end] = targetColumn.time.split('-')
    return { start, end }
  }

  return { start: '00:00', end: '00:00' }
}

// 检查两个时间范围是否重叠
function isTimeOverlap(range1, range2) {
  // 将时间转换为分钟数进行比较
  const start1 = timeToMinutes(range1.start)
  const end1 = timeToMinutes(range1.end)
  const start2 = timeToMinutes(range2.start)
  const end2 = timeToMinutes(range2.end)

  // 检查是否重叠
  return (start1 < end2 && start2 < end1)
}

// 将时间字符串转换为分钟数
function timeToMinutes(timeStr) {
  const [hours, minutes] = timeStr.split(':').map(Number)
  return hours * 60 + minutes
}

// 获取单元格的类名
// 添加控制显示的响应式变量
const showConflicts = ref(true)
const showScheduled = ref(true)

// 修改获取单元格类名的方法
function getCellClass(text, column) {
  if (!text) {
    // 检查所有选中学生是否都可以在这个时间段排课
    const available = selectedStudents.value.every(student =>
      isTimeSlotAvailable(text, column, student),
    )
    return {
      'empty-lesson': true,
      'available-slot': true, // 空闲时间格子始终显示绿色背景
      'conflict-slot': !available && showConflicts.value && selectedStudents.value.length > 0,
      'hidden-conflict': !available && !showConflicts.value && selectedStudents.value.length > 0,
    }
  }
  return {
    'scheduled-hidden': !showScheduled.value && !text.isNewScheduled,
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

// 感统组表格列定义
const columnsSensory = [
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

// 言语组表格列定义（时间不同）
const columnsSpeech = [
  {
    title: '教师',
    dataIndex: 'teacher',
    fixed: 'left',
    width: 135,
  },
  {
    title: '第一节课',
    time: '08:30-09:20',
    width: 135,
    dataIndex: 'lesson1',
  },
  {
    title: '第二节课',
    time: '09:30-10:20',
    width: 135,
    dataIndex: 'lesson2',
  },
  {
    title: '第三节课',
    time: '10:30-11:20',
    width: 135,
    dataIndex: 'lesson3',
  },
  {
    title: '第四节课',
    time: '11:30-12:20',
    width: 135,
    dataIndex: 'lesson4',
  },
  {
    title: '第五节课',
    time: '13:30-14:20',
    width: 135,
    dataIndex: 'lesson5',
  },
  {
    title: '第六节课',
    time: '14:30-15:20',
    dataIndex: 'lesson6',
    width: 135,
  },
  {
    title: '第七节课',
    time: '15:30-16:20',
    width: 135,
    dataIndex: 'lesson7',
  },
  {
    title: '第八节课',
    time: '16:30-17:20',
    width: 135,
    dataIndex: 'lesson8',
  },
  {
    title: '第九节课',
    time: '17:30-18:20',
    width: 135,
    dataIndex: 'lesson9',
  },
  {
    title: '第十节课',
    time: '18:30-19:20',
    width: 135,
    dataIndex: 'lesson10',
  },
  {
    title: '第十一节课',
    time: '19:30-20:20',
    width: 135,
    dataIndex: 'lesson11',
  },
  {
    title: '第十二节课',
    time: '20:30-21:20',
    width: 135,
    dataIndex: 'lesson12',
  },
]

// 感统组表格数据
const dataSourceSensory = ref([
  {
    key: '1',
    teacher: '张老师',
    teacherId: 'T001',
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
    teacherId: 'T002',
    lesson1: { id: '1009', studentName: '林涛', courseName: '语文' },
    lesson2: '', // 移除这节课，因为与言语组的课程时间冲突
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
])

// 言语组表格数据
const dataSourceSpeech = ref([
  {
    key: '3',
    teacher: '王老师',
    teacherId: 'T003',
    lesson1: { id: '2001', studentName: '张小明', courseName: '语言训练' },
    lesson2: '',
    lesson3: { id: '2002', studentName: '李华', courseName: '发音练习' },
    lesson4: { id: '2003', studentName: '赵敏', courseName: '口语表达' },
    lesson5: { id: '2004', studentName: '刘芳', courseName: '语言理解' },
    lesson6: '',
    lesson7: { id: '2005', studentName: '陈明', courseName: '阅读训练' },
    lesson8: { id: '2006', studentName: '杨丽', courseName: '语言游戏' },
    lesson9: '',
    lesson10: { id: '2007', studentName: '周强', courseName: '语言康复' },
    lesson11: '',
    lesson12: { id: '2008', studentName: '吴芳', courseName: '语言评估' },
  },
  {
    key: '4',
    teacher: '赵老师',
    teacherId: 'T004',
    lesson1: { id: '2009', studentName: '林小', courseName: '语言训练' },
    lesson2: { id: '2010', studentName: '黄强', courseName: '发音练习' },
    lesson3: { id: '2011', studentName: '张小明', courseName: '口语表达' },
    lesson4: { id: '2012', studentName: '孙明', courseName: '语言理解' },
    lesson5: { id: '2013', studentName: '郑丽', courseName: '阅读训练' },
    lesson6: { id: '2014', studentName: '吴强', courseName: '语言游戏' },
    lesson7: '',
    lesson8: { id: '2015', studentName: '徐芳', courseName: '语言康复' },
    lesson9: { id: '2016', studentName: '马丽', courseName: '语言评估' },
    lesson10: '',
    lesson11: { id: '2017', studentName: '刘强', courseName: '语言训练' },
    lesson12: '',
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
// 获取冲突信息
function getConflictInfo(studentName, column) {
  const studentLessons = getStudentLessons(studentName)
  const currentTimeRange = getTimeRangeForColumn(column, currentGroup.value)

  // 找到所有冲突的课程
  const conflictLessons = studentLessons.filter((lesson) => {
    return isTimeOverlap(currentTimeRange, lesson.timeRange)
  })

  if (conflictLessons.length === 0) {
    return ''
  }

  // 获取并格式化所有冲突信息
  return conflictLessons.map((conflictLesson) => {
    try {
      // 获取冲突课程的具体信息
      const conflictTeacher = conflictLesson.group === '感统组'
        ? dataSourceSensory.value.find(t => t.teacherId === conflictLesson.teacherId)
        : dataSourceSpeech.value.find(t => t.teacherId === conflictLesson.teacherId)

      if (!conflictTeacher) {
        return `未找到教师信息`
      }

      const columns = conflictLesson.group === '感统组' ? columnsSensory : columnsSpeech
      const timeSlot = columns.find(col => col.dataIndex === `lesson${conflictLesson.lessonIndex}`)

      // 获取课程名称
      const lessonData = conflictTeacher[`lesson${conflictLesson.lessonIndex}`]
      const courseName = lessonData && lessonData.courseName ? lessonData.courseName : '未知课程'

      if (!timeSlot) {
        return `未找到时间段信息`
      }

      return `${studentName}已在${conflictLesson.group}${conflictTeacher.teacher}的第${conflictLesson.lessonIndex}节课(${timeSlot.time})有【${courseName}】课程安排`
    }
    catch (error) {
      console.error('获取冲突信息时出错:', error)
      // 返回更具体的错误信息
      return `获取${studentName}的课程冲突信息时出错：${error.message || '未知错误'}`
    }
  }).join('\n')
}
const conflictModalVisible = ref(false)
const currentConflicts = ref([])

// 显示冲突详情弹窗
// 修改 showConflictModal 方法
function showConflictModal(text, column) {
  // 获取所有冲突信息,并将每条信息单独存储
  currentConflicts.value = selectedStudents.value
    .filter(student => !isTimeSlotAvailable(text, column, student))
    .flatMap((student) => {
      const conflicts = getConflictInfo(student, column).split('\n')
      return conflicts.filter(conflict => conflict.trim() !== '')
    })

  conflictModalVisible.value = true
}
</script>

<template>
  <div class="filter-wrap bg-white pl-3 pr-3 rounded-4 rounded-lt-0 rounded-rt-0">
    <all-filter :display-array="displayArray" :is-show-search-stu-phonefilter="true" />
  </div>
  <div class="time-template mt2 bg-white py3 px5 rounded-4">
    <div class="top-filter flex justify-between flex-items-center">
      <div>
        <div class="flex items-center">
          <span class="mr-2">学生:</span>
          <a-select
            v-model:value="selectedStudents" :max-tag-count="1" mode="multiple" style="width: 155px" placeholder="请选择学生"
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
            <a-select-option value="王芳">
              王芳
            </a-select-option>
            <a-select-option value="赵强">
              赵强
            </a-select-option>
            <a-select-option value="刘洋">
              刘洋
            </a-select-option>
            <a-select-option value="陈静">
              陈静
            </a-select-option>
            <a-select-option value="杨帆">
              杨帆
            </a-select-option>
            <a-select-option value="周婷">
              周婷
            </a-select-option>
            <a-select-option value="林涛">
              林涛
            </a-select-option>
          </a-select>
          <div class="ml-4 flex items-center">
            <a-checkbox v-model:checked="showConflicts" class="mr-1 flex items-center">
              <div class="flex items-center">
                <div class="w-4 h-4 bg-#ffe6e6 mr-1" />
                <span>冲突</span>
              </div>
            </a-checkbox>
            <a-checkbox v-model:checked="showScheduled" class="mr-1 items-center">
              <div class="flex items-center">
                <div class="w-4 h-4 bg-#06f border border-solid border-#06f mr-1" />
                <span>已排</span>
              </div>
            </a-checkbox>
            <div class="w-4 h-4 bg-#e6ffe6 mr-1" />
            <span>可排</span>
          </div>
        </div>
      </div>
      <div class="time-selector flex-center flex-1">
        <!-- 添加组别选择 -->
        <a-radio-group v-model:value="currentGroup" button-style="solid" size="small" class="mr-4">
          <a-radio-button v-for="opt in groupOptions" :key="opt.key" :value="opt.key">
            {{ opt.label }}
          </a-radio-button>
        </a-radio-group>

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
        :columns="currentColumns" :data-source="currentDataSource" :pagination="false" :scroll="{ x: 1560 }"
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
              :class="getCellClass(text, column)" class="lesson-cell py1  cursor-pointer"
              @click="handleCellClick(text, column, record)"
            >
              <template v-if="text">
                <div class="con">
                  <div class="t">
                    <clamped-text :lines="1" :text="text.studentName" />
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
              <template v-else>
                <div v-if="!text" :class="getCellClass(text, column)" class="empty-cell-hint">
                  <template v-if="selectedStudents.every(student => isTimeSlotAvailable(text, column, student))">
                    <span class="text-green-600">空闲时间(可排)</span>
                  </template>
                  <template v-else>
                    <div class="text-red-600">
                      <a class="text-red hover-text-red" @click="showConflictModal(text, column)">
                        查看冲突详情
                      </a>
                    </div>
                  </template>
                </div>
              </template>
            </div>
          </template>
        </template>
      </a-table>
    </div>
    <!-- 在组件末尾添加弹窗组件 -->
    <a-modal
      v-model:open="conflictModalVisible"
      title="日程冲突详情"
      :footer="null"
      width="600px"
    >
      <div class="conflict-details">
        <a-alert
          v-for="(conflict, index) in currentConflicts"
          :key="index"
          :message="conflict"
          type="error"
          class="mb-2"
          show-icon
        />
      </div>
    </a-modal>
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
    border-radius: 4px;
    overflow: hidden;
  }

  .conflict-slot {
    background-color: #ffe6e6; // 红色背景表示冲突
    transition: background-color 0.3s;
    overflow: hidden;
    border-radius: 4px;
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

  .con {
    background: rgba(78, 109, 255, 0.12);
    border-radius: 4px;
    height: 70px;
    font-weight: 400;
    position: relative;

    .t {
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

    .name {
      font-size: 12px;
      color: #002cfd;
      font-weight: 500;

    }

    .courseName {
      color: rgb(0, 44, 253);
      font-size: 12px;
    }
  }

  .empty-cell-hint {
    padding: 0px 4px;
    font-size: 12px;
    text-align: center;
  }
  .hidden-conflict {
  visibility: hidden;
}
}
</style>
