<script setup>
import { LeftOutlined, RightOutlined } from '@ant-design/icons-vue'
import { Modal } from 'ant-design-vue'
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
// 当前选中的时间维度
const currentTime = ref('week')
// 当前的日期区间 - 默认设置为本周
const currentWeek = ref(dayjs())
// 时间维度选项
const timeOptions = [
  { key: 'day', label: '日' },
  { key: 'week', label: '周' },
]
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
// 创建一个方法 用于格式化时间xx月-xx日
function formatDate(date) {
  return dayjs(date).format('MM-DD')
}
// 创建一个方法 用于格式化时间为周x，非星期x
function formatWeek(date) {
  const day = dayjs(date).day()
  const weekMap = {
    0: '日',
    1: '一',
    2: '二',
    3: '三',
    4: '四',
    5: '五',
    6: '六',
  }
  return `周${weekMap[day]}`
}
const dataSource = computed(() => {
  // 根据当前选中的组别选择数据源
  const sourceData = currentGroup.value === 'A' ? rawDataSourceA.value : rawDataSourceB.value

  // 对选中的数据源进行排序
  return [...sourceData].sort((a, b) => {
    // 首先按照teacherId排序
    if (a.teacherId !== b.teacherId) {
      return a.teacherId.localeCompare(b.teacherId)
    }
    // teacherId相同时，按照日期排序
    return a.date.localeCompare(b.date)
  })
})

const rawDataSourceA = ref([
  {
    key: '1',
    name: '张老师',
    teacherId: 't001',
    date: '2025-04-28',
    lessons: [
      {
        startTime: '08:00',
        endTime: '08:40',
        courseName: '口肌训练课',
        studentId: ['10001'],
        classId: null,
        className: null,
        studentNames: [{ id: '10001', name: '张三' }],
        courseType: 1,
      },
      {
        startTime: '08:50',
        endTime: '09:30',
        courseName: '初级感统课',
        studentId: ['10002'],
        classId: null,
        className: null,
        studentNames: [{ id: '10002', name: '李四' }],
        courseType: 1,
      },
      {
        startTime: '09:40',
        endTime: '10:20',
        courseName: 'OT精细课',
        studentId: ['10003'],
        classId: null,
        className: null,
        studentNames: [{ id: '10003', name: '王五' }],
        courseType: 1,
      },
      {
        startTime: '10:30',
        endTime: '11:10',
        courseName: '初级认知课',
        studentId: ['10004', '10009'],
        classId: 'C-01',
        className: '苹果基础班',
        studentNames: [{ id: '10004', name: '赵六' }, { id: '10009', name: '孙八' }],
        courseType: 2,
        isMain: true,
      },
      {
        startTime: '11:20',
        endTime: '12:00',
        courseName: 'PT治疗课',
        studentId: ['10005'],
        classId: null,
        className: null,
        studentNames: [{ id: '10005', name: '钱七' }],
        courseType: 1,
      },
      {
        startTime: '12:10',
        endTime: '12:50',
        courseName: null,
        studentId: null,
        classId: null,
        className: null,
        studentNames: null,
        courseType: null,
      },
      // 新增的6节课
      {
        startTime: '13:00',
        endTime: '13:40',
        courseName: null,
        studentId: null,
        classId: null,
        className: null,
        studentNames: null,
        courseType: null,
      },
      {
        startTime: '13:50',
        endTime: '14:30',
        courseName: null,
        studentId: null,
        classId: null,
        className: null,
        studentNames: null,
        courseType: null,
      },
      {
        startTime: '14:40',
        endTime: '15:20',
        courseName: null,
        studentId: null,
        classId: null,
        className: null,
        studentNames: null,
        courseType: null,
      },
      {
        startTime: '15:30',
        endTime: '16:10',
        courseName: null,
        studentId: null,
        classId: null,
        className: null,
        studentNames: null,
        courseType: null,
      },
      {
        startTime: '16:20',
        endTime: '17:00',
        courseName: null,
        studentId: null,
        classId: null,
        className: null,
        studentNames: null,
        courseType: null,
      },
      {
        startTime: '17:10',
        endTime: '17:50',
        courseName: null,
        studentId: null,
        classId: null,
        className: null,
        studentNames: null,
        courseType: null,
      },
    ],
  },
  {
    key: '1',
    name: '张老师',
    teacherId: 't001',
    date: '2025-04-29',
    lessons: [
      {
        startTime: '08:00',
        endTime: '08:40',
        courseName: null,
        studentId: null,
        classId: null,
        className: null,
        studentNames: null,
      },
      {
        startTime: '08:50',
        endTime: '09:30',
        courseName: '初级感统课',
        studentId: ['10002'],
        classId: null,
        className: null,
        studentNames: [{ id: '10002', name: '李四' }],
        courseType: 1,
      },
      {
        startTime: '09:40',
        endTime: '10:20',
        courseName: null,
        studentId: null,
        classId: null,
        className: null,
        studentNames: null,
        courseType: null,
      },
      {
        startTime: '10:30',
        endTime: '11:10',
        courseName: '初级认知课',
        studentId: ['10004'],
        classId: null,
        className: null,
        studentNames: [{ id: '10004', name: '赵六' }],
        courseType: 1,
      },
      {
        startTime: '11:20',
        endTime: '12:00',
        courseName: null,
        studentId: null,
        classId: null,
        className: null,
        studentNames: null,
        courseType: null,
      },
      {
        startTime: '12:10',
        endTime: '12:50',
        courseName: null,
        studentId: null,
        classId: null,
        className: null,
        studentNames: null,
        courseType: null,
      },
      // 新增的6节课
      {
        startTime: '13:00',
        endTime: '13:40',
        courseName: null,
        studentId: null,
        classId: null,
        className: null,
        studentNames: null,
        courseType: null,
      },
      {
        startTime: '13:50',
        endTime: '14:30',
        courseName: null,
        studentId: null,
        classId: null,
        className: null,
        studentNames: null,
        courseType: null,
      },
      {
        startTime: '14:40',
        endTime: '15:20',
        courseName: null,
        studentId: null,
        classId: null,
        className: null,
        studentNames: null,
        courseType: null,
      },
      {
        startTime: '15:30',
        endTime: '16:10',
        courseName: null,
        studentId: null,
        classId: null,
        className: null,
        studentNames: null,
        courseType: null,
      },
      {
        startTime: '16:20',
        endTime: '17:00',
        courseName: null,
        studentId: null,
        classId: null,
        className: null,
        studentNames: null,
        courseType: null,
      },
      {
        startTime: '17:10',
        endTime: '17:50',
        courseName: null,
        studentId: null,
        classId: null,
        className: null,
        studentNames: null,
        courseType: null,
      },
    ],
  },
  {
    key: '1',
    name: '王老师',
    teacherId: 't002',
    date: '2025-04-28',
    lessons: [
      {
        startTime: '08:00',
        endTime: '08:40',
        courseName: null,
        studentId: null,
        classId: null,
        className: null,
        studentNames: null,
        courseType: null,
      },
      {
        startTime: '08:50',
        endTime: '09:30',
        courseName: null,
        studentId: null,
        classId: null,
        className: null,
        studentNames: null,
        courseType: null,
      },
      {
        startTime: '09:40',
        endTime: '10:20',
        courseName: null,
        studentId: null,
        classId: null,
        className: null,
        studentNames: null,
        courseType: null,
      },
      {
        startTime: '10:30',
        endTime: '11:10',
        courseName: '初级认知课',
        studentId: ['10004', '10009'],
        classId: 'C-01',
        className: '苹果基础班',
        studentNames: [{ id: '10004', name: '赵六' }, { id: '10009', name: '孙八' }],
        courseType: 2,
        isMain: false,
      },
      {
        startTime: '11:20',
        endTime: '12:00',
        courseName: 'PT治疗课',
        studentId: ['10005'],
        classId: null,
        className: null,
        studentNames: [{ id: '10005', name: '钱七' }],
        courseType: 1,
      },
      {
        startTime: '12:10',
        endTime: '12:50',
        courseName: null,
        studentId: null,
        classId: null,
        className: null,
        studentNames: null,
        courseType: null,
      },
      // 新增的6节课
      {
        startTime: '13:00',
        endTime: '13:40',
        courseName: null,
        studentId: null,
        classId: null,
        className: null,
        studentNames: null,
        courseType: null,
      },
      {
        startTime: '13:50',
        endTime: '14:30',
        courseName: null,
        studentId: null,
        classId: null,
        className: null,
        studentNames: null,
        courseType: null,
      },
      {
        startTime: '14:40',
        endTime: '15:20',
        courseName: null,
        studentId: null,
        classId: null,
        className: null,
        studentNames: null,
        courseType: null,
      },
      {
        startTime: '15:30',
        endTime: '16:10',
        courseName: null,
        studentId: null,
        classId: null,
        className: null,
        studentNames: null,
        courseType: null,
      },
      {
        startTime: '16:20',
        endTime: '17:00',
        courseName: null,
        studentId: null,
        classId: null,
        className: null,
        studentNames: null,
        courseType: null,
      },
      {
        startTime: '17:10',
        endTime: '17:50',
        courseName: null,
        studentId: null,
        classId: null,
        className: null,
        studentNames: null,
        courseType: null,
      },
    ],
  },
  {
    key: '1',
    name: '王老师',
    teacherId: 't002',
    date: '2025-04-29',
    lessons: [
      {
        startTime: '08:00',
        endTime: '08:40',
        courseName: '口肌训练课',
        studentId: ['10001'],
        classId: null,
        className: null,
        studentNames: [{ id: '10001', name: '张三' }],
        courseType: 1,
      },
      {
        startTime: '08:50',
        endTime: '09:30',
        courseName: null,
        studentId: null,
        classId: null,
        className: null,
        studentNames: null,
        courseType: null,
      },
      {
        startTime: '09:40',
        endTime: '10:20',
        courseName: null,
        studentId: null,
        classId: null,
        className: null,
        studentNames: null,
        courseType: null,
      },
      {
        startTime: '10:30',
        endTime: '11:10',
        courseName: null,
        studentId: null,
        classId: null,
        className: null,
        studentNames: null,
        courseType: null,
      },
      {
        startTime: '11:20',
        endTime: '12:00',
        courseName: null,
        studentId: null,
        classId: null,
        className: null,
        studentNames: null,
        courseType: null,
      },
      {
        startTime: '12:10',
        endTime: '12:50',
        courseName: null,
        studentId: null,
        classId: null,
        className: null,
        studentNames: null,
        courseType: null,
      },
      // 新增的6节课
      {
        startTime: '13:00',
        endTime: '13:40',
        courseName: null,
        studentId: null,
        classId: null,
        className: null,
        studentNames: null,
        courseType: null,
      },
      {
        startTime: '13:50',
        endTime: '14:30',
        courseName: null,
        studentId: null,
        classId: null,
        className: null,
        studentNames: null,
        courseType: null,
      },
      {
        startTime: '14:40',
        endTime: '15:20',
        courseName: null,
        studentId: null,
        classId: null,
        className: null,
        studentNames: null,
        courseType: null,
      },
      {
        startTime: '15:30',
        endTime: '16:10',
        courseName: null,
        studentId: null,
        classId: null,
        className: null,
        studentNames: null,
        courseType: null,
      },
      {
        startTime: '16:20',
        endTime: '17:00',
        courseName: null,
        studentId: null,
        classId: null,
        className: null,
        studentNames: null,
        courseType: null,
      },
      {
        startTime: '17:10',
        endTime: '17:50',
        courseName: null,
        studentId: null,
        classId: null,
        className: null,
        studentNames: null,
        courseType: null,
      },
    ],
  },
])
const rawDataSourceB = ref([
  {
    key: '1',
    name: '李老师',
    teacherId: 't003',
    date: '2025-04-28',
    lessons: [
      {
        startTime: '08:30',
        endTime: '09:10',
        courseName: '口肌训练课',
        studentId: ['20001'],
        classId: null,
        className: null,
        studentNames: [{ id: '20001', name: '刘一' }],
        courseType: 1,
      },
      {
        startTime: '09:20',
        endTime: '10:00',
        courseName: '初级感统课',
        studentId: ['20002'],
        classId: null,
        className: null,
        studentNames: [{ id: '20002', name: '陈二' }],
        courseType: 1,
      },
      {
        startTime: '10:10',
        endTime: '10:50',
        courseName: null,
        studentId: null,
        classId: null,
        className: null,
        studentNames: null,
      },
      {
        startTime: '11:00',
        endTime: '11:40',
        courseName: '初级认知课',
        studentId: ['20004', '20009'],
        classId: 'C-02',
        className: '橙子基础班',
        studentNames: [{ id: '20004', name: '张四' }, { id: '20009', name: '王九' }],
        courseType: 2,
        isMain: true,
      },
      {
        startTime: '11:50',
        endTime: '12:30',
        courseName: 'PT治疗课',
        studentId: ['20005'],
        classId: null,
        className: null,
        studentNames: [{ id: '20005', name: '李五' }],
        courseType: 1,
      },
      {
        startTime: '12:40',
        endTime: '13:20',
        courseName: 'PT治疗课',
        studentId: ['20006'],
        classId: null,
        className: null,
        studentNames: [{ id: '20006', name: '赵六' }],
        courseType: 1,
      },
      // 新增的6节课
      {
        startTime: '13:30',
        endTime: '14:10',
        courseName: null,
        studentId: null,
        classId: null,
        className: null,
        studentNames: null,
        courseType: null,
      },
      {
        startTime: '14:20',
        endTime: '15:00',
        courseName: null,
        studentId: null,
        classId: null,
        className: null,
        studentNames: null,
        courseType: null,
      },
      {
        startTime: '15:10',
        endTime: '15:50',
        courseName: null,
        studentId: null,
        classId: null,
        className: null,
        studentNames: null,
        courseType: null,
      },
      {
        startTime: '16:00',
        endTime: '16:40',
        courseName: null,
        studentId: null,
        classId: null,
        className: null,
        studentNames: null,
        courseType: null,
      },
      {
        startTime: '16:50',
        endTime: '17:30',
        courseName: null,
        studentId: null,
        classId: null,
        className: null,
        studentNames: null,
        courseType: null,
      },
      {
        startTime: '17:40',
        endTime: '18:20',
        courseName: null,
        studentId: null,
        classId: null,
        className: null,
        studentNames: null,
        courseType: null,
      },
    ],
  },
  {
    key: '1',
    name: '李老师',
    teacherId: 't003',
    date: '2025-04-29',
    lessons: [
      {
        startTime: '08:30',
        endTime: '09:10',
        courseName: null,
        studentId: null,
        classId: null,
        className: null,
        studentNames: null,
      },
      {
        startTime: '09:20',
        endTime: '10:00',
        courseName: '初级感统课',
        studentId: ['20007'],
        classId: null,
        className: null,
        studentNames: [{ id: '20007', name: '钱七' }],
        courseType: 1,
      },
      {
        startTime: '10:10',
        endTime: '10:50',
        courseName: 'OT精细课',
        studentId: ['10009'],
        classId: null,
        className: null,
        studentNames: [{ id: '10009', name: '孙八' }],
        courseType: 1,
      },
      {
        startTime: '11:00',
        endTime: '11:40',
        courseName: '初级认知课',
        studentId: ['20010'],
        classId: null,
        className: null,
        studentNames: [{ id: '20010', name: '周十' }],
        courseType: 1,
      },
      {
        startTime: '11:50',
        endTime: '12:30',
        courseName: null,
        studentId: null,
        classId: null,
        className: null,
        studentNames: null,
      },
      {
        startTime: '12:40',
        endTime: '13:20',
        courseName: null,
        studentId: null,
        classId: null,
        className: null,
        studentNames: null,
        courseType: null,
      },
      // 新增的6节课
      {
        startTime: '13:30',
        endTime: '14:10',
        courseName: null,
        studentId: null,
        classId: null,
        className: null,
        studentNames: null,
        courseType: null,
      },
      {
        startTime: '14:20',
        endTime: '15:00',
        courseName: null,
        studentId: null,
        classId: null,
        className: null,
        studentNames: null,
        courseType: null,
      },
      {
        startTime: '15:10',
        endTime: '15:50',
        courseName: null,
        studentId: null,
        classId: null,
        className: null,
        studentNames: null,
        courseType: null,
      },
      {
        startTime: '16:00',
        endTime: '16:40',
        courseName: null,
        studentId: null,
        classId: null,
        className: null,
        studentNames: null,
        courseType: null,
      },
      {
        startTime: '16:50',
        endTime: '17:30',
        courseName: null,
        studentId: null,
        classId: null,
        className: null,
        studentNames: null,
        courseType: null,
      },
      {
        startTime: '17:40',
        endTime: '18:20',
        courseName: null,
        studentId: null,
        classId: null,
        className: null,
        studentNames: null,
        courseType: null,
      },
    ],
  },
  {
    key: '1',
    name: '陈老师',
    teacherId: 't004',
    date: '2025-04-28',
    lessons: [
      {
        startTime: '08:30',
        endTime: '09:10',
        courseName: '口肌训练课',
        studentId: ['20011'],
        classId: null,
        className: null,
        studentNames: [{ id: '20011', name: '吴十一' }],
        courseType: 1,
      },
      {
        startTime: '09:20',
        endTime: '10:00',
        courseName: '初级感统课',
        studentId: ['20012'],
        classId: null,
        className: null,
        studentNames: [{ id: '20012', name: '郑十二' }],
        courseType: 1,
      },
      {
        startTime: '10:10',
        endTime: '10:50',
        courseName: null,
        studentId: null,
        classId: null,
        className: null,
        studentNames: null,
        courseType: null,
      },
      {
        startTime: '11:00',
        endTime: '11:40',
        courseName: '初级认知课',
        studentId: ['20004', '20009'],
        classId: 'C-02',
        className: '橙子基础班',
        studentNames: [{ id: '20004', name: '张四' }, { id: '20009', name: '王九' }],
        courseType: 2,
        isMain: false,
      },
      {
        startTime: '11:50',
        endTime: '12:30',
        courseName: 'PT治疗课',
        studentId: ['20013'],
        classId: null,
        className: null,
        studentNames: [{ id: '20013', name: '马十三' }],
        courseType: 1,
      },
      {
        startTime: '12:40',
        endTime: '13:20',
        courseName: null,
        studentId: null,
        classId: null,
        className: null,
        studentNames: null,
        courseType: null,
      },
      // 新增的6节课
      {
        startTime: '13:30',
        endTime: '14:10',
        courseName: null,
        studentId: null,
        classId: null,
        className: null,
        studentNames: null,
        courseType: null,
      },
      {
        startTime: '14:20',
        endTime: '15:00',
        courseName: null,
        studentId: null,
        classId: null,
        className: null,
        studentNames: null,
        courseType: null,
      },
      {
        startTime: '15:10',
        endTime: '15:50',
        courseName: null,
        studentId: null,
        classId: null,
        className: null,
        studentNames: null,
        courseType: null,
      },
      {
        startTime: '16:00',
        endTime: '16:40',
        courseName: null,
        studentId: null,
        classId: null,
        className: null,
        studentNames: null,
        courseType: null,
      },
      {
        startTime: '16:50',
        endTime: '17:30',
        courseName: null,
        studentId: null,
        classId: null,
        className: null,
        studentNames: null,
        courseType: null,
      },
      {
        startTime: '17:40',
        endTime: '18:20',
        courseName: null,
        studentId: null,
        classId: null,
        className: null,
        studentNames: null,
        courseType: null,
      },
    ],
  },
  {
    key: '1',
    name: '陈老师',
    teacherId: 't004',
    date: '2025-04-29',
    lessons: [
      {
        startTime: '08:30',
        endTime: '09:10',
        courseName: '口肌训练课',
        studentId: ['20014'],
        classId: null,
        className: null,
        studentNames: [{ id: '20014', name: '林十四' }],
        courseType: 1,
      },
      {
        startTime: '09:20',
        endTime: '10:00',
        courseName: null,
        studentId: null,
        classId: null,
        className: null,
        studentNames: null,
        courseType: null,
      },
      {
        startTime: '10:10',
        endTime: '10:50',
        courseName: 'OT精细课',
        studentId: ['20015'],
        classId: null,
        className: null,
        studentNames: [{ id: '20015', name: '王十五' }],
        courseType: 1,
      },
      {
        startTime: '11:00',
        endTime: '11:40',
        courseName: null,
        studentId: null,
        classId: null,
        className: null,
        studentNames: null,
        courseType: null,
      },
      {
        startTime: '11:50',
        endTime: '12:30',
        courseName: null,
        studentId: null,
        className: null,
        studentNames: null,
        courseType: null,
      },
      {
        startTime: '12:40',
        endTime: '13:20',
        courseName: null,
        studentId: null,
        classId: null,
        className: null,
        studentNames: null,
        courseType: null,
      },
      // 新增的6节课
      {
        startTime: '13:30',
        endTime: '14:10',
        courseName: null,
        studentId: null,
        classId: null,
        className: null,
        studentNames: null,
        courseType: null,
      },
      {
        startTime: '14:20',
        endTime: '15:00',
        courseName: null,
        studentId: null,
        classId: null,
        className: null,
        studentNames: null,
        courseType: null,
      },
      {
        startTime: '15:10',
        endTime: '15:50',
        courseName: null,
        studentId: null,
        classId: null,
        className: null,
        studentNames: null,
        courseType: null,
      },
      {
        startTime: '16:00',
        endTime: '16:40',
        courseName: null,
        studentId: null,
        classId: null,
        className: null,
        studentNames: null,
        courseType: null,
      },
      {
        startTime: '16:50',
        endTime: '17:30',
        courseName: null,
        studentId: null,
        classId: null,
        className: null,
        studentNames: null,
        courseType: null,
      },
      {
        startTime: '17:40',
        endTime: '18:20',
        courseName: null,
        studentId: null,
        classId: null,
        className: null,
        studentNames: null,
        courseType: null,
      },
    ],
  },
])
// columns的定义也需要相应修改
const columns = computed(() => {
  // 获取最大课程数
  const maxLessons = Math.max(...dataSource.value.map(item => item.lessons.length))

  // 基础列定义
  const baseColumns = [
    {
      title: '教师',
      dataIndex: 'name',
      key: 'name',
      width: 120,
      align: 'center',
      fixed: 'left',
      customCell: (_, index) => {
        const currentTeacherId = dataSource.value[index].teacherId
        if (index === 0 || dataSource.value[index - 1].teacherId !== currentTeacherId) {
          let count = 1
          for (let i = index + 1; i < dataSource.value.length; i++) {
            if (dataSource.value[i].teacherId === currentTeacherId) {
              count++
            }
            else {
              break
            }
          }
          return { rowSpan: count }
        }
        return { rowSpan: 0 }
      },
    },
    {
      title: '日期',
      dataIndex: 'date',
      key: 'date',
      width: 80,
      fixed: 'left',
      align: 'center',
    },
  ]

  // 动态生成课程列
  const lessonColumns = Array.from({ length: maxLessons }, (_, index) => ({
    title: `第${index + 1}节课`,
    startTime: dataSource.value[0].lessons[index]?.startTime || '',
    endTime: dataSource.value[0].lessons[index]?.endTime || '',
    dataIndex: ['lessons', index],
    key: `lesson${index}`,
    width: 160,
    align: 'center',
  }))

  return [...baseColumns, ...lessonColumns]
})
// 课程列表数据
const courseList = ref([
  {
    id: '589251114063479808',
    name: '初级认知课',
    courseType: 1,
  },
  {
    id: '58925112157479108',
    name: '初级感统课',
    courseType: 1,
  },
  {
    id: '589251121574791081',
    name: 'PT治疗课',
    courseType: 1,
  },
  {
    id: '589251121574791082',
    name: 'OT精细课',
    courseType: 1,
  },
  {
    id: '589251121574791083',
    name: '口肌训练课',
    courseType: 1,
  },
  {
    id: '589251121574791084',
    name: '初级认知课',
    courseType: 2,
  },
])
// 一对一课程数据
const oneToOneData = ref([
  {
    id: '589251755896808448',
    courseId: '589251114063479808',
    courseName: '初级认知课',
    studentId: '10001',
    studentName: '张三',
    name: '张三-初级认知课',
    remainQuantity: 9,
  },
  {
    id: '5892517551234546775',
    courseId: '58925112157479108',
    courseName: '初级感统课',
    studentId: '10002',
    studentName: '李四',
    name: '李四-初级感统课',
    remainQuantity: 6,
  },
  {
    id: '5892517551234546776',
    courseId: '589251121574791081',
    courseName: 'PT治疗课',
    studentId: '10003',
    studentName: '王五',
    name: '王五-PT治疗课',
    remainQuantity: 10,
  },
  {
    id: '5892517551234546777',
    courseId: '589251121574791082',
    courseName: 'OT精细课',
    studentId: '20014',
    studentName: '林十四',
    name: '林十四-OT精细课',
    remainQuantity: 10,
  },
  {
  // 孙八
    id: '5892517551234546778',
    courseId: '589251121574791083',
    courseName: '口肌训练课',
    studentId: '10009',
    studentName: '孙八',
    name: '孙八-口肌训练课',
    remainQuantity: 10,
  },
])
const studentId = ref(null)
const studentIds = ref([])
const courseId = ref(null)
const courseName = ref(null)
const classId = ref(null)
const className = ref(null)
const teacherId = ref(null)
// 选择1v1触发
function timeToMinutes(timeStr) {
  const [hours, minutes] = timeStr.split(':').map(Number)
  return hours * 60 + minutes
}

function handle1v1(value) {
  if (!value) {
    // 清除所有冲突标记
    dataSource.value.forEach((teacher) => {
      teacher.lessons.forEach((lesson) => {
        if (!lesson.studentId) {
          lesson.conflict = false
        }
      })
    })
    return
  }

  // 获取学生信息
  const studentInfo = oneToOneData.value.find(item => item.studentId === value)
  if (!studentInfo)
    return

  // 检查冲突
  checkConflicts(studentInfo)
}

// 获取所有组合并的数据源(用于冲突检测)
const allDataSource = computed(() => {
  return [...rawDataSourceA.value, ...rawDataSourceB.value]
})

// 检查两个时间段是否有交叉
function isTimeOverlap(time1, time2) {
  // 将时间转换为分钟数进行比较
  const timeToMinutes = (timeStr) => {
    const [hours, minutes] = timeStr.split(':').map(Number)
    return hours * 60 + minutes
  }

  const start1 = timeToMinutes(time1.start)
  const end1 = timeToMinutes(time1.end)
  const start2 = timeToMinutes(time2.start)
  const end2 = timeToMinutes(time2.end)

  // 检查时间是否交叉
  return (start1 < end2 && start2 < end1)
}

// 检查1v1冲突
function checkConflicts(studentInfo) {
  console.log('运行1v1冲突检测', studentInfo)

  // 先清除所有冲突标记
  dataSource.value.forEach((teacher) => {
    teacher.lessons.forEach((lesson) => {
      if (!lesson.studentId) {
        lesson.conflict = false
        lesson.conflictReason = null
      }
    })
  })

  // 遍历所有老师的课表查找学生已有的课程时间段 (包括A组和B组)
  const studentLessonTimes = []

  allDataSource.value.forEach((teacher) => {
    teacher.lessons.forEach((lesson) => {
      // 如果该时间段已经有这个学生的课，记录下来
      if (lesson.studentId && lesson.studentId.includes(studentInfo.studentId)) {
        studentLessonTimes.push({
          date: teacher.date,
          startTime: lesson.startTime,
          endTime: lesson.endTime,
          teacherName: teacher.name,
          courseName: lesson.courseName,
          lessonIndex: getLessonIndex(lesson.startTime),
          studentName: studentInfo.studentName,
          group: teacher.teacherId.startsWith('t00') ? 'A组' : 'B组', // 根据老师ID判断所属组别
        })
      }
    })
  })

  // 标记冲突时间段
  if (studentLessonTimes.length > 0) {
    dataSource.value.forEach((teacher) => {
      teacher.lessons.forEach((lesson) => {
        // 只检查空闲时段
        if (!lesson.studentId) {
          // 检查是否与学生已有课程在同一天且时间重叠
          const conflictLesson = studentLessonTimes.find(existingLesson =>
            existingLesson.date === teacher.date
            && isTimeOverlap(
              { start: existingLesson.startTime, end: existingLesson.endTime },
              { start: lesson.startTime, end: lesson.endTime },
            ),
          )

          if (conflictLesson) {
            lesson.conflict = true
            // 记录冲突原因
            const month = dayjs(conflictLesson.date).format('M')
            const day = dayjs(conflictLesson.date).format('D')
            lesson.conflictReason = {
              type: '1v1',
              studentName: conflictLesson.studentName,
              date: `${month}月${day}日`,
              lessonIndex: conflictLesson.lessonIndex,
              teacherName: conflictLesson.teacherName,
              courseName: conflictLesson.courseName,
              group: conflictLesson.group,
              time: `${conflictLesson.startTime}-${conflictLesson.endTime}`,
            }
          }
        }
      })
    })
  }
}
const currentModel = ref('1')
// 班级数据
const classData = ref([
  {
    id: 'C-01',
    name: '苹果基础班',
    studentIds: ['589250903194799104', '5892509031876223323', '10001'],
    studentNames: ['陈陈', '晨晨', '张三'],
    courseId: '589251114063479808',
    courseName: '初级认知课',
    mainTeacherId: 't001',
    mainTeacherName: '张老师',
  },
  {
    id: 'C-02',
    name: '橙子基础班',
    studentIds: ['20004', '20009', '5892509031876223323'],
    studentNames: ['张四', '王九', '晨晨'],
    courseId: '589251121574791084',
    courseName: '初级认知课',
    mainTeacherId: 't003',
    mainTeacherName: '李老师',
  },
])
// 选择班级触发
function handleClass(value) {
  if (!value) {
    // 清除所有冲突标记
    dataSource.value.forEach((teacher) => {
      teacher.lessons.forEach((lesson) => {
        if (!lesson.studentId) {
          lesson.conflict = false
        }
      })
    })
    return
  }

  // 获取班级信息
  const classInfo = classData.value.find(item => item.id === value)
  if (!classInfo)
    return

  console.log('选择班级', classInfo.name)

  // 检查班课冲突
  checkClassCrossTimeConflicts(classInfo)
}
// 当前选中的组别
const currentGroup = ref('A')
const groupOptions = [
  { key: 'A', label: 'A组' },
  { key: 'B', label: 'B组' },
]

// 监听组别变化
watch(currentGroup, () => {
  // 如果当前有选中的学生，重新进行冲突检测
  if (studentId.value) {
    handle1v1(studentId.value)
  }
  // 如果当前有选中的班级，重新进行冲突检测
  if (classId.value) {
    handleClass(classId.value)
  }
})

// 检查班课交叉时间冲突
function checkClassCrossTimeConflicts(classInfo) {
  console.log('运行班课冲突检测', classInfo)

  // 先清除所有冲突标记
  dataSource.value.forEach((teacher) => {
    teacher.lessons.forEach((lesson) => {
      if (!lesson.studentId) {
        lesson.conflict = false
        lesson.conflictReason = null
      }
    })
  })

  // 首先收集这个班级在所有组已排课的时间段（跨组检测）
  const classExistingLessons = []

  allDataSource.value.forEach((teacher) => {
    teacher.lessons.forEach((lesson) => {
      // 如果这个时间段已经排了当前班级的课
      if (lesson.classId === classInfo.id) {
        classExistingLessons.push({
          date: teacher.date,
          startTime: lesson.startTime,
          endTime: lesson.endTime,
          teacherName: teacher.name,
          teacherId: teacher.teacherId,
          lessonIndex: getLessonIndex(lesson.startTime),
        })
      }
    })
  })

  console.log('班级已排课时间段', classExistingLessons)

  // 遍历所有老师的课表
  dataSource.value.forEach((teacher) => {
    // 检查每个时间段
    teacher.lessons.forEach((lesson, lessonIndex) => {
      if (!lesson.studentId) {
        // 获取当前时间段信息
        const currentTime = {
          date: teacher.date,
          startTime: lesson.startTime,
          endTime: lesson.endTime,
        }

        let hasConflict = false
        let conflictReason = null

        // 1. 检查班级跨组交叉时段冲突 - 只检查时间段不同但有重叠的情况
        // 注意：同一班级在完全相同的时间段不算冲突（允许安排主教+辅教）
        const classTimeConflict = classExistingLessons.find(existingLesson =>
          existingLesson.date === currentTime.date
          // 关键逻辑：只有当时间段不完全相同但有重叠时才算冲突
          && (existingLesson.startTime !== currentTime.startTime
            || existingLesson.endTime !== currentTime.endTime)
          && isTimeOverlap(
            { start: existingLesson.startTime, end: existingLesson.endTime },
            { start: currentTime.startTime, end: currentTime.endTime },
          ),
        )

        if (classTimeConflict) {
          console.log('班级跨组交叉时段冲突', classInfo.name, currentTime.date, currentTime.startTime)
          hasConflict = true

          // 记录冲突原因
          const month = dayjs(classTimeConflict.date).format('M')
          const day = dayjs(classTimeConflict.date).format('D')

          // 获取冲突课程所在组别
          const conflictGroup = classTimeConflict.teacherId.startsWith('t00') ? 'A组' : 'B组'

          conflictReason = {
            type: '班级时间段交叉冲突',
            className: classInfo.name,
            date: `${month}月${day}日`,
            lessonIndex: classTimeConflict.lessonIndex,
            teacherName: classTimeConflict.teacherName,
            group: conflictGroup,
            time: `${classTimeConflict.startTime}-${classTimeConflict.endTime}`,
          }
        }

        // 2. 检查教师冲突 - 同一教师在同一时间是否有其他班级的课
        if (!hasConflict) {
          const teacherOtherLesson = teacher.lessons.find((l, idx) =>
            idx !== lessonIndex
            && l.courseType === 2
            && l.classId !== classInfo.id
            && isTimeOverlap(
              { start: l.startTime, end: l.endTime },
              { start: currentTime.startTime, end: currentTime.endTime },
            ),
          )

          if (teacherOtherLesson) {
            console.log('教师已有其他班级课程', teacher.name, currentTime.startTime)
            hasConflict = true

            // 记录冲突原因
            const month = dayjs(teacher.date).format('M')
            const day = dayjs(teacher.date).format('D')
            conflictReason = {
              type: '教师班课冲突',
              teacherName: teacher.name,
              date: `${month}月${day}日`,
              lessonIndex: getLessonIndex(currentTime.startTime),
              className: teacherOtherLesson.className,
              courseName: teacherOtherLesson.courseName,
              time: `${teacherOtherLesson.startTime}-${teacherOtherLesson.endTime}`,
            }
          }
        }

        // 3. 检查学生冲突 - 班级学生是否在同一时间有其他课程 (跨组检测)
        if (!hasConflict && classInfo.studentIds?.length > 0) {
          // 遍历所有组的老师课表，查找同一时间的课程
          teacherLoop: for (const t of allDataSource.value) {
            // 只检查同一天的课程
            if (t.date === currentTime.date) {
              const sameTimeLessons = t.lessons.filter(l =>
                l.studentId
                && isTimeOverlap(
                  { start: l.startTime, end: l.endTime },
                  { start: currentTime.startTime, end: currentTime.endTime },
                ),
              )

              // 检查每个同时间的课程
              for (const sameTimeLesson of sameTimeLessons) {
                // 如果是当前选中的班级课程，不算冲突
                if (sameTimeLesson.classId === classInfo.id)
                  continue

                // 检查学生是否有交集
                for (const sid of classInfo.studentIds) {
                  if (sameTimeLesson.studentId?.includes(sid)) {
                    console.log('学生时间冲突', currentTime.date, currentTime.startTime, sameTimeLesson.startTime)
                    hasConflict = true

                    // 查找冲突学生姓名
                    const studentIndex = classInfo.studentIds.indexOf(sid)
                    const studentName = studentIndex >= 0 ? classInfo.studentNames[studentIndex] : '未知学生'

                    // 记录冲突原因
                    const month = dayjs(t.date).format('M')
                    const day = dayjs(t.date).format('D')

                    // 获取冲突课程所在组别
                    const conflictGroup = t.teacherId.startsWith('t00') ? 'A组' : 'B组'

                    conflictReason = {
                      type: '学生课程冲突',
                      studentName,
                      date: `${month}月${day}日`,
                      lessonIndex: getLessonIndex(sameTimeLesson.startTime),
                      teacherName: t.name,
                      courseName: sameTimeLesson.courseName,
                      className: sameTimeLesson.className,
                      group: conflictGroup,
                      time: `${sameTimeLesson.startTime}-${sameTimeLesson.endTime}`,
                    }

                    break teacherLoop // 找到一个冲突就跳出循环
                  }
                }
              }
            }
          }
        }

        // 设置冲突标记和原因
        lesson.conflict = hasConflict
        lesson.conflictReason = conflictReason
      }
    })
  })
}

// 检查班级在某个时间段是否已经有课程安排及主教设置
function checkClassExistingTeacherRole(classId, teacherId, startTime, endTime) {
  console.log('检查班级主教/辅教角色', classId, teacherId, startTime)

  // 获取班级信息
  const classInfo = classData.value.find(item => item.id === classId)
  if (!classInfo) {
    console.log('未找到班级信息，默认设置为主教')
    return { isMainTeacher: true, hasExistingArrangement: false }
  }

  // 统一仅使用mainTeacherId判断
  // 如果老师ID等于班级配置的主教ID，则为主教；否则为辅教
  const isMainTeacher = classInfo.mainTeacherId === teacherId

  console.log('根据班级配置判断角色:', isMainTeacher ? '主教' : '辅教')
  console.log('班级配置的主教ID:', classInfo.mainTeacherId, '当前老师ID:', teacherId)

  // 检查是否已存在该班级课程安排
  let hasExistingArrangement = false

  // 遍历所有老师的所有日期，检查是否已有该班级同时段的课程
  allDataSource.value.forEach((teacher) => {
    teacher.lessons.forEach((lesson) => {
      // 只检查与当前时间段相同的时间段
      if (lesson.startTime === startTime && lesson.endTime === endTime) {
        // 检查是否是同一个班级的课程
        if (lesson.classId === classId) {
          hasExistingArrangement = true
        }
      }
    })
  })

  console.log('是否已有该班级课程安排:', hasExistingArrangement)
  console.log('最终角色设置:', isMainTeacher ? '主教' : '辅教')
  return { isMainTeacher, hasExistingArrangement }
}

// 处理冲突点击
function handleConflictClick(timeSlot, column) {
  let content = '该时间段已有课程安排，无法排课'

  // 根据冲突原因提供更详细的信息
  if (timeSlot.conflictReason) {
    const reason = timeSlot.conflictReason
    const groupInfo = reason.group ? `(${reason.group})` : ''
    const timeInfo = reason.time ? `[${reason.time}]` : ''

    if (reason.type === '1v1') {
      content = `该时间段${reason.studentName}在${reason.date}第${reason.lessonIndex}节课${timeInfo}已有${reason.teacherName}${groupInfo}的${reason.courseName}课程安排，无法排课`
    }
    else if (reason.type === '教师班课冲突') {
      content = `该时间段${reason.teacherName}在${reason.date}第${reason.lessonIndex}节课${timeInfo}已有${reason.className}的${reason.courseName}班课安排，无法排课`
    }
    else if (reason.type === '学生课程冲突') {
      content = `该时间段${reason.studentName}在${reason.date}第${reason.lessonIndex}节课${timeInfo}已有${reason.teacherName}${groupInfo}的${reason.courseName || (`${reason.className}班课`)}课程安排，无法排课`
    }
    else if (reason.type === '班级时间段交叉冲突') {
      content = `该时间段${reason.className}在${reason.date}第${reason.lessonIndex}节课${timeInfo}已有${reason.teacherName}${groupInfo}的课程安排，不支持交叉时间段排课`
    }
  }

  Modal.info({
    title: '时间冲突',
    content,
  })
}

// 排课
function handleScheduleClick(timeSlot, column, record) {
  if (currentModel.value === '1') {
    // 1v1排课逻辑
    if (!studentId.value) {
      Modal.warning({
        title: '请先选择学生',
        content: '请先在上方选择要排课的学生',
      })
      return
    }

    // 获取学生信息
    const studentInfo = oneToOneData.value.find(
      item => item.studentId === studentId.value,
    )

    if (!studentInfo) {
      Modal.warning({
        title: '学生信息不存在',
        content: '请选择有效的学生',
      })
      return
    }

    // 获取月份和日期信息
    const dateObj = dayjs(record.date)
    const month = dateObj.format('M')
    const day = dateObj.format('D')
    const lessonIndex = getLessonIndex(column.startTime)

    Modal.confirm({
      title: '确认排课',
      content: `确定要为 ${studentInfo.studentName} 安排 ${month}月${day}日 ${record.name} 第${lessonIndex}节课 ${column.startTime}-${column.endTime} 的课程吗？`,
      onOk() {
        console.log('确认排课', studentInfo.studentName, column.startTime, column.endTime)

        // 更新数据源
        const targetTeacher = dataSource.value.find(
          t => t.teacherId === record.teacherId && t.date === record.date,
        )

        if (!targetTeacher)
          return

        // 获取列索引
        const columnIndex = column.dataIndex[1]

        // 使用列索引直接获取正确的时间槽
        const targetLesson = targetTeacher.lessons[columnIndex]

        if (!targetLesson)
          return

        // 更新课程信息
        Object.assign(targetLesson, {
          studentId: [studentInfo.studentId],
          courseName: studentInfo.courseName,
          courseType: 1,
          studentNames: [{
            id: studentInfo.studentId,
            name: studentInfo.studentName,
          }],
          classId: null,
          className: null,
          conflict: false,
        })

        // 重新检查冲突
        handle1v1(studentId.value)
      },
    })
  }
  else {
    // 班课排课逻辑
    if (!classId.value) {
      Modal.warning({
        title: '请先选择班级',
        content: '请先在上方选择要排课的班级',
      })
      return
    }

    const classInfo = classData.value.find(
      item => item.id === classId.value,
    )

    if (!classInfo) {
      Modal.warning({
        title: '班级信息不存在',
        content: '请选择有效的班级',
      })
      return
    }

    // 检查时间冲突
    if (timeSlot.conflict) {
      Modal.warning({
        title: '时间冲突',
        content: '该时间段已有冲突，不可排课',
      })
      return
    }

    // 获取月份和日期信息
    const dateObj = dayjs(record.date)
    const month = dateObj.format('M')
    const day = dateObj.format('D')
    const lessonIndex = getLessonIndex(column.startTime)

    Modal.confirm({
      title: '确认排课',
      content: `确定要为 ${classInfo.name} 安排 ${month}月${day}日 ${record.name} 第${lessonIndex}节课 ${column.startTime}-${column.endTime} 的课程吗？`,
      onOk() {
        console.log('确认排课', classInfo.name, column.startTime, column.endTime)

        // 更新数据源
        const targetTeacher = dataSource.value.find(
          t => t.teacherId === record.teacherId && t.date === record.date,
        )

        if (!targetTeacher)
          return

        // 获取列索引
        const columnIndex = column.dataIndex[1]

        // 使用列索引直接获取正确的时间槽
        const targetLesson = targetTeacher.lessons[columnIndex]

        if (!targetLesson)
          return

        // 检查主教/辅教角色
        const { isMainTeacher } = checkClassExistingTeacherRole(
          classInfo.id,
          record.teacherId,
          targetLesson.startTime,
          targetLesson.endTime,
        )

        // 更新课程信息
        Object.assign(targetLesson, {
          classId: classInfo.id,
          className: classInfo.name,
          courseName: classInfo.courseName,
          courseType: 2,
          isMain: isMainTeacher, // 根据检查结果设置是否为主教
          studentNames: classInfo.studentNames.map(name => ({ name })),
          studentId: classInfo.studentIds,
          conflict: false,
        })

        console.log('更新课程信息完成', targetLesson)

        // 重新检查班课交叉时间冲突
        checkClassCrossTimeConflicts(classInfo)
      },
    })
  }
}

// 获取课程节数
function getLessonIndex(startTime) {
  const timeMapA = {
    '08:00': 1,
    '08:50': 2,
    '09:40': 3,
    '10:30': 4,
    '11:20': 5,
    '12:10': 6,
    '13:00': 7,
    '13:50': 8,
    '14:40': 9,
    '15:30': 10,
    '16:20': 11,
    '17:10': 12,
  }

  const timeMapB = {
    '08:30': 1,
    '09:20': 2,
    '10:10': 3,
    '11:00': 4,
    '11:50': 5,
    '12:40': 6,
    '13:30': 7,
    '14:20': 8,
    '15:10': 9,
    '16:00': 10,
    '16:50': 11,
    '17:40': 12,
  }

  return timeMapA[startTime] || timeMapB[startTime] || ''
}

// 添加监听，当模式切换时清空之前的选择
watch(currentModel, (newValue) => {
  console.log('切换模式', newValue)

  // 清除所有冲突标记
  dataSource.value.forEach((teacher) => {
    teacher.lessons.forEach((lesson) => {
      if (!lesson.studentId) {
        lesson.conflict = false
        lesson.conflictReason = null
      }
    })
  })

  if (newValue === '1') {
    // 切换到1v1模式，清空班级选择
    classId.value = null
    className.value = null
  }
  else {
    // 切换到班课模式，清空学生选择
    studentId.value = null
    courseId.value = null
    courseName.value = null
  }
})

// 生成空闲课程数据
function createEmptyLessonsA() {
  return [
    {
      startTime: '08:00',
      endTime: '08:40',
      courseName: null,
      studentId: null,
      classId: null,
      className: null,
      studentNames: null,
      courseType: null,
    },
    {
      startTime: '08:50',
      endTime: '09:30',
      courseName: null,
      studentId: null,
      classId: null,
      className: null,
      studentNames: null,
      courseType: null,
    },
    {
      startTime: '09:40',
      endTime: '10:20',
      courseName: null,
      studentId: null,
      classId: null,
      className: null,
      studentNames: null,
      courseType: null,
    },
    {
      startTime: '10:30',
      endTime: '11:10',
      courseName: null,
      studentId: null,
      classId: null,
      className: null,
      studentNames: null,
      courseType: null,
    },
    {
      startTime: '11:20',
      endTime: '12:00',
      courseName: null,
      studentId: null,
      classId: null,
      className: null,
      studentNames: null,
      courseType: null,
    },
    {
      startTime: '12:10',
      endTime: '12:50',
      courseName: null,
      studentId: null,
      classId: null,
      className: null,
      studentNames: null,
      courseType: null,
    },
    {
      startTime: '13:00',
      endTime: '13:40',
      courseName: null,
      studentId: null,
      classId: null,
      className: null,
      studentNames: null,
      courseType: null,
    },
    {
      startTime: '13:50',
      endTime: '14:30',
      courseName: null,
      studentId: null,
      classId: null,
      className: null,
      studentNames: null,
      courseType: null,
    },
    {
      startTime: '14:40',
      endTime: '15:20',
      courseName: null,
      studentId: null,
      classId: null,
      className: null,
      studentNames: null,
      courseType: null,
    },
    {
      startTime: '15:30',
      endTime: '16:10',
      courseName: null,
      studentId: null,
      classId: null,
      className: null,
      studentNames: null,
      courseType: null,
    },
    {
      startTime: '16:20',
      endTime: '17:00',
      courseName: null,
      studentId: null,
      classId: null,
      className: null,
      studentNames: null,
      courseType: null,
    },
    {
      startTime: '17:10',
      endTime: '17:50',
      courseName: null,
      studentId: null,
      classId: null,
      className: null,
      studentNames: null,
      courseType: null,
    },
  ]
}

function createEmptyLessonsB() {
  return [
    {
      startTime: '08:30',
      endTime: '09:10',
      courseName: null,
      studentId: null,
      classId: null,
      className: null,
      studentNames: null,
      courseType: null,
    },
    {
      startTime: '09:20',
      endTime: '10:00',
      courseName: null,
      studentId: null,
      classId: null,
      className: null,
      studentNames: null,
      courseType: null,
    },
    {
      startTime: '10:10',
      endTime: '10:50',
      courseName: null,
      studentId: null,
      classId: null,
      className: null,
      studentNames: null,
      courseType: null,
    },
    {
      startTime: '11:00',
      endTime: '11:40',
      courseName: null,
      studentId: null,
      classId: null,
      className: null,
      studentNames: null,
      courseType: null,
    },
    {
      startTime: '11:50',
      endTime: '12:30',
      courseName: null,
      studentId: null,
      classId: null,
      className: null,
      studentNames: null,
      courseType: null,
    },
    {
      startTime: '12:40',
      endTime: '13:20',
      courseName: null,
      studentId: null,
      classId: null,
      className: null,
      studentNames: null,
      courseType: null,
    },
    {
      startTime: '13:30',
      endTime: '14:10',
      courseName: null,
      studentId: null,
      classId: null,
      className: null,
      studentNames: null,
      courseType: null,
    },
    {
      startTime: '14:20',
      endTime: '15:00',
      courseName: null,
      studentId: null,
      classId: null,
      className: null,
      studentNames: null,
      courseType: null,
    },
    {
      startTime: '15:10',
      endTime: '15:50',
      courseName: null,
      studentId: null,
      classId: null,
      className: null,
      studentNames: null,
      courseType: null,
    },
    {
      startTime: '16:00',
      endTime: '16:40',
      courseName: null,
      studentId: null,
      classId: null,
      className: null,
      studentNames: null,
      courseType: null,
    },
    {
      startTime: '16:50',
      endTime: '17:30',
      courseName: null,
      studentId: null,
      classId: null,
      className: null,
      studentNames: null,
      courseType: null,
    },
    {
      startTime: '17:40',
      endTime: '18:20',
      courseName: null,
      studentId: null,
      classId: null,
      className: null,
      studentNames: null,
      courseType: null,
    },
  ]
}

// 添加A组老师额外的5天数据
const extraDaysA = [
  '2025-04-30',
  '2025-05-01',
  '2025-05-02',
  '2025-05-03',
  '2025-05-04',
]

// 为张老师添加额外的5天数据
extraDaysA.forEach((date) => {
  rawDataSourceA.value.push({
    key: '1',
    name: '张老师',
    teacherId: 't001',
    date,
    lessons: createEmptyLessonsA(),
  })
})

// 为王老师添加额外的5天数据
extraDaysA.forEach((date) => {
  rawDataSourceA.value.push({
    key: '1',
    name: '王老师',
    teacherId: 't002',
    date,
    lessons: createEmptyLessonsA(),
  })
})

// 添加B组老师额外的5天数据
const extraDaysB = [
  '2025-04-30',
  '2025-05-01',
  '2025-05-02',
  '2025-05-03',
  '2025-05-04',
]

// 为李老师添加额外的5天数据
extraDaysB.forEach((date) => {
  rawDataSourceB.value.push({
    key: '1',
    name: '李老师',
    teacherId: 't003',
    date,
    lessons: createEmptyLessonsB(),
  })
})

// the same for 陈老师
extraDaysB.forEach((date) => {
  rawDataSourceB.value.push({
    key: '1',
    name: '陈老师',
    teacherId: 't004',
    date,
    lessons: createEmptyLessonsB(),
  })
})
</script>

<template>
  <div>
    <div class="filter-wrap bg-white pl-3 pr-3 rounded-4 rounded-lt-0 rounded-rt-0">
      <all-filter :display-array="displayArray" :is-show-search-stu-phonefilter="true" />
    </div>
    <div class="time-template mt2 bg-white py3 px5 rounded-4 rounded-lb-0 rounded-rb-0">
      <div class="top-filter flex justify-between flex-items-center">
        <div class="mr2">
          <a-radio-group v-model:value="currentModel" button-style="solid">
            <a-radio-button value="1">
              1v1
            </a-radio-button>
            <a-radio-button value="2">
              班课
            </a-radio-button>
          </a-radio-group>
        </div>
        <div>
          <div v-if="currentModel == 1" class="flex items-center">
            <!-- 写一个 select下拉选择框，使用 一对一课程数据  -->
            <span>选择一对一：</span>
            <a-select
              v-model:value="studentId" allow-clear placeholder="请搜索/选择一对一" style="width: 160px"
              option-label-prop="label" @change="handle1v1"
            >
              <!-- 原有选项内容保持不变 -->
              <a-select-option
                v-for="item in oneToOneData" :key="item.id" :value="item.studentId" :data="item"
                :label="item.name"
              >
                <div>{{ item.name }}</div>
              </a-select-option>
            </a-select>
          </div>
          <div v-if="currentModel == 2" class="flex items-center">
            <!-- 写一个 select下拉选择框，使用 班级数据  -->
            <span>选择班级：</span>
            <a-select
              v-model:value="classId" allow-clear placeholder="请搜索/选择班级" style="width: 160px"
              option-label-prop="label" @change="handleClass"
            >
              <!-- 原有选项内容保持不变 -->
              <a-select-option
                v-for="item in classData" :key="item.id" :value="item.id" :data="item"
                :label="item.name"
              >
                <div>{{ item.name }}</div>
                <div class="text-3 text-#666">
                  主教：{{ item.mainTeacherName }}
                </div>
              </a-select-option>
            </a-select>
          </div>
        </div>
        <div class="time-selector flex-center flex-1">
          <a-radio-group v-model:value="currentTime" button-style="solid">
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
                  v-model:value="currentWeek" class="absolute top-0 left-0 right-0 bottom-0 z-10 opacity-0"
                  :allow-clear="false" :bordered="false" :format="formatDateRange" style="cursor:pointer;"
                />
                <a-date-picker
                  v-else-if="currentTime === 'week'"
                  v-model:value="currentWeek" class="absolute top-0 left-0 right-0 bottom-0 z-10 opacity-0"
                  picker="week" :allow-clear="false" :bordered="false" :format="formatDateRange"
                  style="cursor:pointer;"
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
        <div>
          <!-- 添加组别选择 -->
          <a-radio-group v-model:value="currentGroup" button-style="solid" class="mr-4">
            <a-radio-button v-for="opt in groupOptions" :key="opt.key" :value="opt.key">
              {{ opt.label }}
            </a-radio-button>
          </a-radio-group>
        </div>
        <a-space>
          <create-schedule-popover />
          <a-button>导出课表</a-button>
        </a-space>
      </div>
    </div>
    <a-table
      :scroll="{ x: 1300 }" :sticky="{ offsetHeader: 100 }" size="small" :pagination="false" bordered
      :data-source="dataSource" :columns="columns"
    >
      <template #headerCell="{ column }">
        <template v-if="column.startTime && column.endTime">
          <div>{{ column.title }}</div>
          <div class="text-12px text-#666 line-height-2">
            {{ column.startTime }}-{{ column.endTime }}
          </div>
        </template>
        <template v-else>
          {{ column.title }}
        </template>
      </template>
      <template #bodyCell="{ column, record, text }">
        <template v-if="column.dataIndex?.[0] === 'lessons'">
          <div v-if="text.studentId" class=" flex  flex-col bg-#4e6dff1f h-11 rounded-1 text-3 text-#fff">
            <!-- 方格头部时间 -->
            <!-- 班课 -->
            <div class="pl1 bg-#06f rounded-1 rounded-lb-0 rounded-rb-0 flex relative h-5">
              {{ column.startTime }}-{{ column.endTime }}
              <!-- 标记 -->
              <span
                class="absolute right-0 pl-2 pr-1  h-4 bg-#00000080 text-#fff text-2.5 font-500 rounded-rt-1 rounded-lb-2"
              >
                <span v-if="text.courseType == 1">1v1</span>
                <span v-if="text.courseType == 2">班课(<span>{{ text.isMain ? '主教' : '辅教' }}</span>)</span>
              </span>
            </div>
            <!-- 1v1 -->
            <div v-if="!text.classId" class="flex pl-1 flex-1 text-#002cfd cursor-pointer flex-items-center">
              <span v-for="(item, index) in text.studentNames" :key="index">
                <div class="flex">{{ item.name }}{{ index !== text.studentNames.length - 1 ? '、' : '' }}-{{ text.courseName }}</div>
              </span>
            </div>
            <!-- 班课 -->
            <div v-else class="flex  pl-1 flex-1 text-#002cfd cursor-pointer line-height-4 flex-items-center">
              <div class="flex">
                {{ text.className }}-{{ text.courseName }}
              </div>
            </div>
          </div>
          <!-- 空闲时段 -->
          <div
            v-else class="h-11 rounded-1 text-3 flex-center cursor-pointer" :class="[
              text.conflict ? 'bg-#ffe6e6 text-#a31616' : 'bg-#e6ffe6 text-#16a34a',
            ]" @click="text.conflict ? handleConflictClick(text, column) : handleScheduleClick(text, column, record)"
          >
            {{ text.conflict ? '时间冲突(不可排)' : '空闲时段(可排)' }}
          </div>
        </template>
        <template v-if="column.key == 'date'">
          <div class="text-3.5 ">
            {{ formatWeek(text) }}
          </div>
          <div class="text-3 font-500 line-height-3 text-#666">
            {{ formatDate(text) }}
          </div>
        </template>
        <template v-if="column.key == 'name'">
          <div>{{ text }}</div>
          <div class="text-3 cursor-pointer text-#06f">
            查看空闲时间
          </div>
        </template>
      </template>
    </a-table>
  </div>
</template>

<style lang="less" scoped>
.time-selector {
  font-family: DINAlternate-Bold, DINAlternate;

  .ant-radio-button-wrapper {
    padding: 0 16px;
  }
}

:deep(td.ant-table-cell.ant-table-cell-row-hover) {
  background-color: rgb(231, 236, 255) !important;
}

:deep(td.ant-table-cell) {
  padding: 4px !important;
}
</style>
