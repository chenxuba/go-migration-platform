<script setup>
// 引入icon
import { LeftOutlined, RightOutlined } from '@ant-design/icons-vue'
import { message } from 'ant-design-vue'
import { computed, ref, watch } from 'vue'
import dayjs from 'dayjs'

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
const successCount = ref(0)
let resetTimer = null
let messageInstance = null

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

// 将时间字符串转换为分钟数
function timeToMinutes(timeStr) {
  const [hours, minutes] = timeStr.split(':').map(Number)
  return hours * 60 + minutes
}

// 获取单元格的类名
// 添加控制显示的响应式变量
const showConflicts = ref(true)
const showScheduled = ref(true)

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

// 获取指定时段和组别的时间范围
function getTimeRangeForSlot(slot, group, timeIndex) {
  // 如果slot为null，从currentTimeSlots中获取对应时段的period
  if (!slot || !slot.period) {
    const timeSlots = currentTimeSlots.value
    // 使用timeIndex获取正确的时间段
    const currentSlot = timeSlots[timeIndex]
    if (currentSlot && currentSlot.period) {
      const [start, end] = currentSlot.period.split('-')
      return { start, end }
    }
  }
  else {
    const [start, end] = slot.period.split('-')
    return { start, end }
  }
  return { start: '00:00', end: '00:00' }
}

// 获取学生的所有课程信息（包括两个组）
function getStudentLessons(studentName, day) {
  const lessons = []
  const dayMap = {
    'monday': 0,
    'tuesday': 1,
    'wednesday': 2,
    'thursday': 3,
    'friday': 4,
    'saturday': 5,
    'sunday': 6,
  }
  const currentDayIndex = dayMap[day]
  const weekdays = ['mondayCourse', 'tuesdayCourse', 'wednesdayCourse', 'thursdayCourse', 'fridayCourse', 'saturdayCourse', 'sundayCourse']

  // 辅助函数：检查学生名字是否在课程中
  const isStudentInCourse = (course, student) => {
    if (!course || !course.studentName)
      return false

    // 分割学生名字字符串，比较是否包含目标学生
    const studentNames = course.studentName.split('、')
    return studentNames.includes(student)
  }

  // 检查感统组数据
  sensoryIntegrationData.value.forEach((teacher) => {
    teacher.soltTime.forEach((slot) => {
      const course = slot[weekdays[currentDayIndex]]
      // 使用辅助函数检查学生是否在课程中
      if (isStudentInCourse(course, studentName)) {
        lessons.push({
          id: course.studentId,
          group: '感统组',
          teacherId: teacher.teacherId,
          timeRange: getTimeRangeForSlot(slot, '感统组'),
        })
      }
    })
  })

  // 检查言语组数据
  speechGroupData.value.forEach((teacher) => {
    teacher.soltTime.forEach((slot) => {
      const course = slot[weekdays[currentDayIndex]]
      // 使用辅助函数检查学生是否在课程中
      if (isStudentInCourse(course, studentName)) {
        lessons.push({
          id: course.studentId,
          group: '言语组',
          teacherId: teacher.teacherId,
          timeRange: getTimeRangeForSlot(slot, '言语组'),
        })
      }
    })
  })

  return lessons
}

// 检查时间段是否可用（跨组冲突检测）
function isTimeSlotAvailable(slot, day, studentName, timeIndex) {
  // 获取当前选中学生在同一天的所有课程信息
  const studentLessons = getStudentLessons(studentName, day)

  // 获取当前时间段的实际时间范围
  const currentTimeRange = getTimeRangeForSlot(slot, currentGroup.value, timeIndex)

  // 检查是否与学生的其他课程时间冲突
  return !studentLessons.some(lesson =>
    isTimeOverlap(currentTimeRange, lesson.timeRange),
  )
}

// 修改获取单元格类名的方法
function getCellClass(record, day, timeIndex) {
  if (!record) {
    // 检查所有选中学生是否都可以在这个时间段排课
    const available = selectedStudents.value.length === 0
      || selectedStudents.value.every(student =>
        isTimeSlotAvailable(record, day, student, timeIndex),
      )

    return {
      'empty-lesson': true,
      'available-slot': available,
      'conflict-slot': !showConflicts.value ? false : !available && selectedStudents.value.length > 0,
    }
  }
  return {
    'scheduled-hidden': !showScheduled.value,
  }
}

// 时间维度选项
const timeOptions = [
  { key: 'day', label: '日' },
  { key: 'week', label: '周' },
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
// 定义时段数据
const sensoryTimeSlots = [
  { time: '第一节课', period: '08:00-08:45' },
  { time: '第二节课', period: '09:00-09:45' },
  { time: '第三节课', period: '10:00-10:45' },
]

const speechTimeSlots = [
  { time: '第一节课', period: '09:00-09:50' },
  { time: '第二节课', period: '10:00-10:50' },
  { time: '第三节课', period: '11:00-11:50' },
]

// 根据当前选中的组别获取对应的时段
const currentTimeSlots = computed(() => {
  return currentGroup.value === '感统组' ? sensoryTimeSlots : speechTimeSlots
})

// 定义课表的列
const columns = [
  {
    title: '教师',
    dataIndex: 'teacher',
    width: 100,
    fixed: 'left',
    align: 'center',
    customCell: (record, rowIndex) => {
      return {
        rowSpan: rowIndex % currentTimeSlots.value.length === 0 ? currentTimeSlots.value.length : 0,
      }
    },
  },
  {
    title: '时段',
    dataIndex: 'timeSlot',
    width: 120,
    fixed: 'left',
    align: 'center',
  },
  {
    title: '周一',
    dataIndex: 'monday',
    date: currentWeek.value.startOf('week').format('MM-DD'), // 添加日期字段
    width: 150,
    align: 'center',
  },
  {
    title: '周二',
    dataIndex: 'tuesday',
    date: currentWeek.value.startOf('week').add(1, 'day').format('MM-DD'), // 添加日期字段
    width: 150,
    align: 'center',
  },
  {
    title: '周三',
    dataIndex: 'wednesday',
    date: currentWeek.value.startOf('week').add(2, 'day').format('MM-DD'), // 添加日期字段
    width: 150,
    align: 'center',
  },
  {
    title: '周四',
    dataIndex: 'thursday',
    date: currentWeek.value.startOf('week').add(3, 'day').format('MM-DD'), // 添加日期字段
    width: 150,
    align: 'center',
  },
  {
    title: '周五',
    dataIndex: 'friday',
    date: currentWeek.value.startOf('week').add(4, 'day').format('MM-DD'), // 添加日期字段
    width: 150,
    align: 'center',
  },
  {
    title: '周六',
    dataIndex: 'saturday',
    date: currentWeek.value.startOf('week').add(5, 'day').format('MM-DD'), // 添加日期字段
    width: 150,
    align: 'center',
  },
  {
    title: '周日',
    dataIndex: 'sunday',
    date: currentWeek.value.startOf('week').add(6, 'day').format('MM-DD'), // 添加日期字段
    width: 150,
    align: 'center',
  },
]

// 模拟固定的课程数据 - 感统组
const sensoryIntegrationData = ref([
  {
    id: 1,
    teacherId: 1,
    teacherName: '王丽老师',
    soltTime: [
      {
        time: '第一节课',
        period: '08:00-08:45',
        mondayCourse: { studentName: '张三', studentId: '1', courseName: '感统训练' },
        tuesdayCourse: { studentName: '李四', studentId: '2', courseName: '语言训练' },
        wednesdayCourse: { studentName: '王五', studentId: '3', courseName: '认知训练' },
        thursdayCourse: null,
        fridayCourse: { studentName: '赵六', studentId: '4', courseName: '感统训练' },
        saturdayCourse: null,
        sundayCourse: null,
      },
      {
        time: '第二节课',
        period: '09:00-09:45',
        mondayCourse: { studentName: '李四', studentId: '2', courseName: '语言训练' },
        tuesdayCourse: null,
        wednesdayCourse: { studentName: '张三', studentId: '1', courseName: '感统训练' },
        thursdayCourse: { studentName: '王五', studentId: '3', courseName: '认知训练' },
        fridayCourse: null,
        saturdayCourse: { studentName: '赵六', studentId: '4', courseName: '感统训练' },
        sundayCourse: null,
      },
      {
        time: '第三节课',
        period: '10:00-10:45',
        mondayCourse: null,
        tuesdayCourse: { studentName: '王五', studentId: '3', courseName: '认知训练' },
        wednesdayCourse: null,
        thursdayCourse: { studentName: '张三', studentId: '1', courseName: '感统训练' },
        fridayCourse: { studentName: '李四', studentId: '2', courseName: '语言训练' },
        saturdayCourse: null,
        sundayCourse: null,
      },
    ],
  },
  {
    id: 2,
    teacherId: 2,
    teacherName: '李明老师',
    soltTime: [
      {
        time: '第一节课',
        period: '08:00-08:45',
        mondayCourse: { studentName: '赵六', studentId: '4', courseName: '数学' },
        tuesdayCourse: { studentName: '王五', studentId: '3', courseName: '英语' },
        wednesdayCourse: { studentName: '李四', studentId: '2', courseName: '物理' },
        thursdayCourse: null,
        fridayCourse: { studentName: '张三', studentId: '1', courseName: '化学' },
        saturdayCourse: null,
        sundayCourse: null,
      },
      {
        time: '第二节课',
        period: '09:00-09:45',
        mondayCourse: { studentName: '王五', studentId: '3', courseName: '英语' },
        tuesdayCourse: null,
        wednesdayCourse: { studentName: '赵六', studentId: '4', courseName: '数学' },
        thursdayCourse: { studentName: '李四', studentId: '2', courseName: '物理' },
        fridayCourse: null,
        saturdayCourse: null,
        sundayCourse: null,
      },
      {
        time: '第三节课',
        period: '10:00-10:45',
        mondayCourse: null,
        tuesdayCourse: { studentName: '张三', studentId: '1', courseName: '化学' },
        wednesdayCourse: null,
        thursdayCourse: { studentName: '赵六', studentId: '4', courseName: '数学' },
        fridayCourse: { studentName: '王五', studentId: '3', courseName: '英语' },
        saturdayCourse: null,
        sundayCourse: null,
      },
    ],
  },
])

// 模拟言语组的课程数据 - 使用不同的时段和不同的学生
const speechGroupData = ref([
  {
    id: 3,
    teacherId: 3,
    teacherName: '陈语老师',
    soltTime: [
      {
        time: '第一节课',
        period: '09:00-09:50',
        mondayCourse: { studentName: '刘一', studentId: '5', courseName: '语音训练' },
        tuesdayCourse: { studentName: '陈二', studentId: '6', courseName: '发音矫正' },
        wednesdayCourse: { studentName: '张三', studentId: '7', courseName: '语言理解' },
        thursdayCourse: null,
        fridayCourse: { studentName: '李四', studentId: '8', courseName: '口语表达' },
        saturdayCourse: { studentName: '王五', studentId: '9', courseName: '语言训练' },
        sundayCourse: null,
      },
      {
        time: '第二节课',
        period: '10:00-10:50',
        mondayCourse: { studentName: '陈二', studentId: '6', courseName: '发音矫正' },
        tuesdayCourse: null,
        wednesdayCourse: { studentName: '刘一', studentId: '5', courseName: '语音训练' },
        thursdayCourse: { studentName: '张三', studentId: '7', courseName: '语言理解' },
        fridayCourse: null,
        saturdayCourse: { studentName: '李四', studentId: '8', courseName: '口语表达' },
        sundayCourse: null,
      },
      {
        time: '第三节课',
        period: '11:00-11:50',
        mondayCourse: null,
        tuesdayCourse: { studentName: '李四', studentId: '8', courseName: '口语表达' },
        wednesdayCourse: null,
        thursdayCourse: { studentName: '王五', studentId: '9', courseName: '语言训练' },
        fridayCourse: { studentName: '刘一', studentId: '5', courseName: '语音训练' },
        saturdayCourse: null,
        sundayCourse: null,
      },
    ],
  },
  {
    id: 4,
    teacherId: 4,
    teacherName: '赵言老师',
    soltTime: [
      {
        time: '第一节课',
        period: '09:00-09:50',
        mondayCourse: { studentName: '赵六', studentId: '10', courseName: '语言发展' },
        tuesdayCourse: { studentName: '孙七', studentId: '11', courseName: '语言障碍矫正' },
        wednesdayCourse: { studentName: '周八', studentId: '12', courseName: '语言表达' },
        thursdayCourse: null,
        fridayCourse: { studentName: '吴九', studentId: '13', courseName: '语言理解' },
        saturdayCourse: null,
        sundayCourse: null,
      },
      {
        time: '第二节课',
        period: '10:00-10:50',
        mondayCourse: { studentName: '孙七', studentId: '11', courseName: '语言障碍矫正' },
        tuesdayCourse: null,
        wednesdayCourse: { studentName: '赵六', studentId: '10', courseName: '语言发展' },
        thursdayCourse: { studentName: '周八', studentId: '12', courseName: '语言表达' },
        fridayCourse: null,
        saturdayCourse: { studentName: '吴九', studentId: '13', courseName: '语言理解' },
        sundayCourse: null,
      },
      {
        time: '第三节课',
        period: '11:00-11:50',
        mondayCourse: null,
        tuesdayCourse: { studentName: '吴九', studentId: '13', courseName: '语言理解' },
        wednesdayCourse: null,
        thursdayCourse: { studentName: '赵六', studentId: '10', courseName: '语言发展' },
        fridayCourse: { studentName: '孙七', studentId: '11', courseName: '语言障碍矫正' },
        saturdayCourse: null,
        sundayCourse: null,
      },
    ],
  },
])

// 根据当前选中的组别获取对应的数据
const mockData = computed(() => {
  return currentGroup.value === '感统组' ? sensoryIntegrationData.value : speechGroupData.value
})

// 监听组别变化，重新生成表格数据
watch(currentGroup, () => {
  tableData.value = generateTableData()
})

// 生成表格数据
function generateTableData() {
  const data = []
  const timeSlots = currentTimeSlots.value

  mockData.value.forEach((teacher) => {
    teacher.soltTime.forEach((slot, timeIndex) => {
      data.push({
        key: `${teacher.teacherId}-${timeIndex}`,
        teacher: timeIndex === 0 ? teacher.teacherName : '', // 只在第一行显示教师名
        timeSlot: `${slot.time}\n${slot.period}`,
        monday: slot.mondayCourse,
        tuesday: slot.tuesdayCourse,
        wednesday: slot.wednesdayCourse,
        thursday: slot.thursdayCourse,
        friday: slot.fridayCourse,
        saturday: slot.saturdayCourse,
        sunday: slot.sundayCourse,
      })
    })
  })
  return data
}

const tableData = ref(generateTableData())

// 自定义单元格渲染
function customCell(record, day, timeIndex) {
  if (record) {
    // 如果是已排课程且设置为不显示，返回空的样式
    if (!showScheduled.value && !record.isNew) {
      return {
        content: '',
        style: {
          backgroundColor: '#fff',
          padding: '8px',
          height: '40px',
          lineHeight: '24px',
          cursor: 'default',
          pointerEvents: 'none',
          whiteSpace: 'nowrap',
          overflow: 'hidden',
          textOverflow: 'ellipsis',
        },
        html: '',
      }
    }

    // 检查是否是新添加的课程
    const isNew = record.isNew === true

    // 处理学生名字显示，如果包含顿号，说明是多人课程
    const displayName = record.studentName.includes('、')
      ? record.studentName // 如果已经包含顿号，直接使用
      : record.studentName // 否则使用单个名字

    return {
      content: `${displayName} - ${record.courseName}`,
      style: {
        backgroundColor: '#4e6dff',
        color: '#fff',
        borderRadius: '4px',
        padding: '8px',
        fontSize: '12px',
        fontWeight: 'bold',
        position: 'relative',
        height: '40px',
        lineHeight: '24px',
        whiteSpace: 'nowrap',
        overflow: 'hidden',
        textOverflow: 'ellipsis',
      },
      html: isNew
        ? `
        <div style="
          position: absolute;
          top: -5px;
          right: -5px;
          background-color: #ff4d4f;
          color: white;
          padding: 2px 6px;
          border-radius: 10px;
          font-size: 12px;
          transform: scale(0.8);
        ">新</div>
        <div style="white-space:nowrap;overflow:hidden;text-overflow:ellipsis;">${displayName} - ${record.courseName}</div>
      `
        : `<div style="white-space:nowrap;overflow:hidden;text-overflow:ellipsis;">${displayName} - ${record.courseName}</div>`,
    }
  }

  // 检查是否有学生被选中且该时段是否有冲突
  const hasConflict = selectedStudents.value.length > 0
    && !selectedStudents.value.every(student =>
      isTimeSlotAvailable(null, day, student, timeIndex),
    )

  // 如果设置为不显示冲突，则返回空的样式
  if (!showConflicts.value && hasConflict) {
    return {
      content: '',
      style: {
        backgroundColor: '#fff',
        padding: '8px',
        height: '40px',
        lineHeight: '24px',
        cursor: 'default', // 添加默认光标样式
        pointerEvents: 'none', // 禁用点击事件
      },
    }
  }

  return {
    content: hasConflict ? '点击查看冲突详情' : '',
    style: {
      backgroundColor: hasConflict ? '#ffe6e6' : '#fff',
      padding: '8px',
      color: '#ff4d4f',
      cursor: 'pointer', // 始终显示手型光标
      fontSize: '12px',
      textAlign: 'center',
      height: '40px', // 添加固定高度
      lineHeight: '24px', // 添加行高
    },
  }
}
const currentWeekday = ref('')

// 添加控制 Modal 的响应式变量
const showConflictModal = ref(false)
const conflictDetails = ref([])

// 修改获取冲突信息的方法
function getConflictInfo(studentName, slot, day) {
  const studentLessons = getStudentLessons(studentName, day)
  const currentTimeRange = getTimeRangeForSlot(slot, currentGroup.value)
  const weekdays = ['mondayCourse', 'tuesdayCourse', 'wednesdayCourse', 'thursdayCourse', 'fridayCourse', 'saturdayCourse', 'sundayCourse']
  const weekdayNames = ['一', '二', '三', '四', '五', '六', '日']
  const week = ['monday', 'tuesday', 'wednesday', 'thursday', 'friday', 'saturday', 'sunday']
  const dayIndex = week.indexOf(day)
  currentWeekday.value = weekdayNames[dayIndex]
  const dayMap = {
    'monday': 0,
    'tuesday': 1,
    'wednesday': 2,
    'thursday': 3,
    'friday': 4,
    'saturday': 5,
    'sunday': 6,
  }

  const conflictLessons = studentLessons.filter(lesson =>
    isTimeOverlap(currentTimeRange, lesson.timeRange),
  )

  if (conflictLessons.length > 0) {
    return conflictLessons.map((lesson) => {
      // 获取正确的数据源
      const dataSource = lesson.group === '感统组' ? sensoryIntegrationData.value : speechGroupData.value
      // 找到对应的老师
      const teacher = dataSource.find(t => t.teacherId === lesson.teacherId)
      // 找到对应的时间段
      const timeSlot = teacher?.soltTime.find(s =>
        s.period === `${lesson.timeRange.start}-${lesson.timeRange.end}`,
      )
      // 获取对应的课程信息
      const course = timeSlot?.[weekdays[dayMap[day]]]

      return {
        studentName,
        group: lesson.group,
        teacherName: teacher?.teacherName || '未知老师',
        timeRange: lesson.timeRange,
        courseName: course?.courseName || '未知课程',
        time: timeSlot?.time || '',
      }
    })
  }

  return []
}

function showSuccessMessage() {
  successCount.value++

  if (messageInstance) {
    messageInstance()
  }

  const content = successCount.value > 1
    ? `排课成功！+${successCount.value}`
    : '排课成功！'

  messageInstance = message.success({
    content,
    duration: 2,
    onClose: () => {
      if (resetTimer) {
        clearTimeout(resetTimer)
      }
      resetTimer = setTimeout(() => {
        successCount.value = 0
        messageInstance = null
      }, 3000)
    },
  })
}

// 修改处理单元格点击的方法
function handleCellClick(record, day, timeSlot, teacherId) {
  console.log(day)

  // 判断是否有选中学生
  if (!record && !selectedStudents.value.length) {
    message.warning('请先选择学生')
    return
  }

  // 如果是空单元格，检查是否可以排课
  if (!record) {
    // 从timeSlot中提取时间段索引
    const timeIndex = currentTimeSlots.value.findIndex(slot =>
      `${slot.time}\n${slot.period}` === timeSlot,
    )

    // 检查所有选中学生是否都可以在这个时间段排课
    const availableResults = selectedStudents.value.map((student) => {
      const available = isTimeSlotAvailable(null, day, student, timeIndex)
      const conflicts = available ? [] : getConflictInfo(student, currentTimeSlots.value[timeIndex], day)
      return {
        student,
        available,
        conflicts,
      }
    })

    const unavailableStudents = availableResults.filter(result => !result.available)

    if (unavailableStudents.length > 0) {
      // 显示所有冲突信息
      conflictDetails.value = unavailableStudents.flatMap(result => result.conflicts)
      showConflictModal.value = true
      return
    }

    // 所有学生都可以排课，执行排课逻辑
    try {
      // 获取当前老师的数据
      const currentTeacher = mockData.value.find(t => t.teacherId.toString() === teacherId)
      if (!currentTeacher) {
        message.error('未找到对应的教师信息')
        return
      }

      // 获取对应的时间段
      const slotObj = currentTeacher.soltTime[timeIndex]
      if (!slotObj) {
        message.error('未找到对应的时间段')
        return
      }

      // 获取对应的星期几的课程字段名
      const weekdayFields = {
        'monday': 'mondayCourse',
        'tuesday': 'tuesdayCourse',
        'wednesday': 'wednesdayCourse',
        'thursday': 'thursdayCourse',
        'friday': 'fridayCourse',
        'saturday': 'saturdayCourse',
        'sunday': 'sundayCourse',
      }

      // 拼接所有学生名字
      const studentNames = selectedStudents.value.join('、')

      // 新增课程对象
      const newCourse = {
        studentName: studentNames,
        courseName: '待定',
        isNew: true,
        // 可按需补充其他字段
      }

      // 将newCourse赋值到对应的表格数据
      slotObj[weekdayFields[day]] = newCourse

      // 重新生成表格数据
      tableData.value = generateTableData()

      showSuccessMessage()
      // 不再清空选中的学生
    }
    catch (error) {
      console.error('排课失败:', error)
      message.error('排课失败，请重试')
    }
  }
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
            v-model:value="selectedStudents" :max-tag-count="1" mode="multiple" style="width: 155px"
            placeholder="请选择学生" @change="handleStudentChange"
          >
            <a-select-option value="张三">
              张三
            </a-select-option>
            <a-select-option value="李四">
              李四
            </a-select-option>
            <a-select-option value="王五">
              王五
            </a-select-option>
            <a-select-option value="赵六">
              赵六
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
        <a-button type="primary">
          创建日程
        </a-button>
        <a-button>导出课表</a-button>
      </a-space>
    </div>
    <div class="center-content">
      <a-table
        size="small" :columns="columns" :data-source="tableData" :scroll="{ x: 1300 }"
        :sticky="{ offsetHeader: 100 }" :pagination="false" bordered
      >
        <template #headerCell="{ column }">
          <div v-if="column.dataIndex === 'teacher'">
            教师
          </div>
          <div v-else-if="column.dataIndex === 'timeSlot'">
            时段
          </div>
          <div v-else>
            {{ column.title }}
            <div style="line-height: 1;" class="text-3 text-#666 font-400">
              {{ column.date }}
            </div>
          </div>
        </template>
        <template #bodyCell="{ column, text, record, index }">
          <!-- 教师列合并处理 -->
          <template v-if="column.dataIndex === 'teacher'">
            <span v-if="text" class="teacher">{{ record.teacher }}</span>
          </template>
          <!-- 时段列处理 -->
          <template v-else-if="column.dataIndex === 'timeSlot'">
            <div class="time-slot-cell">
              <div class="time-name">
                {{ text.split('\n')[0] }}
              </div>
              <div class="time-period">
                {{ text.split('\n')[1] }}
              </div>
            </div>
          </template>
          <!-- 课程信息渲染 -->
          <template
            v-else-if="['monday', 'tuesday', 'wednesday', 'thursday', 'friday', 'saturday', 'sunday'].includes(column.dataIndex)"
          >
            <div
              v-if="text" :style="customCell(text).style"
              @click="handleCellClick(text, column.dataIndex, record.timeSlot, record.key.split('-')[0])"
              v-html="customCell(text).html"
            />
            <div
              v-else :style="customCell(null, column.dataIndex, index % currentTimeSlots.length).style"
              :class="getCellClass(text, column.dataIndex, index % currentTimeSlots.length)"
              @click="handleCellClick(text, column.dataIndex, record.timeSlot, record.key.split('-')[0])"
            >
&nbsp;
              {{ customCell(null, column.dataIndex, index % currentTimeSlots.length).content }}
            </div>
          </template>
        </template>
      </a-table>
    </div>
    <!-- 修改冲突详情 Modal -->
    <a-modal
      v-model:open="showConflictModal" title="日程冲突详情" :footer="null"
      width="700px" @ok="showConflictModal = false" @cancel="showConflictModal = false"
    >
      <div class="conflict-details">
        <a-alert v-for="(conflict, index) in conflictDetails" :key="index" show-icon type="error" class="mb-3">
          <template #message>
            <div class="conflict-message">
              <div class="font-bold mb-1">
                学生 {{ conflict.studentName }} 时段冲突
              </div>
              <div class="text-gray-600">
                已在{{ conflict.group }}{{ conflict.teacherName }} 周{{ currentWeekday }} 的
                {{ conflict.time }}（{{ conflict.timeRange.start }}-{{ conflict.timeRange.end }}）
                时段安排了【{{ conflict.courseName }}】课程
              </div>
            </div>
          </template>
        </a-alert>
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
}

// 调整行高和单元格高度
:deep(.ant-table-tbody .ant-table-cell) {
  line-height: 1.5;
  padding: 4px !important;
}

// 时段列特殊处理
.time-slot-cell {
  .time-name {
    line-height: 18px;
    font-size: 12px;
    font-weight: 500;
  }

  .time-period {
    line-height: 16px;
    font-size: 12px;
  }
}

// 保持表头高度不变
:deep(.ant-table-thead .ant-table-cell) {
  height: 40px;
  padding: 8px !important;
}

/* 添加冲突和可用状态的样式 */
.empty-lesson {
  cursor: pointer;
}

.available-slot {
  background-color: #e6ffe6 !important;
}

.conflict-slot {
  background-color: #ffe6e6 !important;
}

.scheduled-hidden {
  display: none;
}

.conflict-details {
  max-height: 400px;
  overflow-y: auto;
}

.conflict-message {
  line-height: 1.5;
}
</style>
