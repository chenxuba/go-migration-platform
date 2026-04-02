<script setup lang="ts">
import {
  BookOutlined,
  CalendarOutlined,
  ClockCircleOutlined,
  CloseOutlined,
  EnvironmentOutlined,
  PlusOutlined,
  QuestionCircleOutlined,
  TeamOutlined,
  UserOutlined,
} from '@ant-design/icons-vue'
import dayjs, { type Dayjs } from 'dayjs'
import { computed, ref, watch } from 'vue'
import type { OneToOneItem } from '../../../api/edu-center/one-to-one'

type PreviewTone = 'pending' | 'blocked'
type ScheduleType = 'oneToOne' | 'studentLesson'
type SchedulingMode = 'repeat' | 'free'
type RepeatRule = 'none' | 'weekly' | 'biweekly' | 'daily' | 'alternateDay'
type HolidayPolicy = 'include' | 'filter'
type TimeMode = 'school' | 'custom'

interface PreviewItem {
  date: string
  week: string
  rule: string
  time: string
  teacher: string
  classroom: string
  tone: PreviewTone
}

interface SummaryItem {
  label: string
  value: string
}

interface QuickTag {
  text: string
  tone: 'primary' | 'warning' | 'neutral'
}

interface OptionItem<T = string> {
  value: T
  label: string
  desc?: string
}

interface SchoolTimeSlot {
  value: string
  label: string
  desc: string
  start: string
  end: string
}

interface TimeBlock {
  key: string
  rangeText: string
  minutes: number
}

interface CustomTimeRangeRow {
  start: Dayjs | null
  end: Dayjs | null
}

const props = defineProps<{
  open: boolean
}>()

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
}>()

const modalOpen = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

const weekDayOptions = ['周一', '周二', '周三', '周四', '周五', '周六', '周日']
const weekdayToNumber: Record<string, number> = {
  周日: 0,
  周一: 1,
  周二: 2,
  周三: 3,
  周四: 4,
  周五: 5,
  周六: 6,
}
const weekDisplayMap: Record<number, string> = {
  0: '周日',
  1: '周一',
  2: '周二',
  3: '周三',
  4: '周四',
  5: '周五',
  6: '周六',
}

const scheduleTypeOptions: OptionItem<ScheduleType>[] = [
  { value: 'oneToOne', label: '按1对1', desc: '从现有1对1档案直接带出排课信息' },
  { value: 'studentLesson', label: '按学员和课程', desc: '按学员与课程组合查看待创建对象' },
]

const schedulingModeOptions: OptionItem<SchedulingMode>[] = [
  { value: 'repeat', label: '重复排课', desc: '批量生成固定周期日程' },
  { value: 'free', label: '自由排课', desc: '先创建单次日程，后续再补排' },
]

const repeatRuleOptions: OptionItem<RepeatRule>[] = [
  { value: 'none', label: '不重复' },
  { value: 'weekly', label: '每周重复' },
  { value: 'biweekly', label: '隔周重复' },
  { value: 'daily', label: '每天重复' },
  { value: 'alternateDay', label: '隔天重复' },
]

const holidayPolicyOptions: OptionItem<HolidayPolicy>[] = [
  { value: 'include', label: '不过滤节假日' },
  { value: 'filter', label: '过滤节假日' },
]

const timeModeOptions: OptionItem<TimeMode>[] = [
  { value: 'school', label: '课表节次', desc: '按学校课表「第N节课」时间段排课' },
  { value: 'custom', label: '自定义时段', desc: '自定义开始时间和结束时间' },
]

const repeatRuleLabelMap: Record<RepeatRule, string> = {
  none: '不重复',
  weekly: '每周重复',
  biweekly: '隔周重复',
  daily: '每天重复',
  alternateDay: '隔天重复',
}

const schoolTimeSlotOptions: SchoolTimeSlot[] = [
  {
    value: 'slot-a',
    label: '第1节课',
    desc: '第1节课',
    start: '08:00',
    end: '08:45',
  },
  {
    value: 'slot-b',
    label: '第2节课',
    desc: '第2节课',
    start: '08:55',
    end: '09:40',
  },
  {
    value: 'slot-c',
    label: '第3节课',
    desc: '第3节课',
    start: '09:50',
    end: '10:35',
  },
  {
    value: 'slot-d',
    label: '第4节课',
    desc: '第4节课',
    start: '10:45',
    end: '11:30',
  },
  {
    value: 'slot-e',
    label: '第5节课',
    desc: '第5节课',
    start: '14:00',
    end: '14:45',
  },
  {
    value: 'slot-f',
    label: '第6节课',
    desc: '第6节课',
    start: '14:55',
    end: '15:40',
  },
]

const schoolHolidaySet = new Set(['2026-05-01', '2026-05-02', '2026-05-03'])
const assistantPool = ['王助教', '刘助教', '陈助教']
const sharedClassrooms = ['个训室 03', '个训室 05', '言语室 02', '感统室 01']

const oneToOneRecords: OneToOneItem[] = [
  {
    id: '10001',
    name: '王小明-语言表达1对1',
    studentName: '王小明',
    studentId: 'stu-001',
    lessonId: 'lesson-001',
    lessonName: '语言表达提升',
    classRoomId: 'room-03',
    classRoomName: '个训室 03',
    defaultTeacherId: 'teacher-001',
    defaultTeacherName: '李老师',
    classTeacherId: 'advisor-001',
    classTeacherName: '陈老师',
    status: 1,
    classStudentStatus: 1,
    isScheduled: true,
    classTime: 45,
    studentClassTime: 1,
    teacherClassTime: 0,
    defaultClassTimeRecordMode: 1,
    createdTime: '2026-03-20 16:20:00',
    tuitionAccountCount: 1,
    tuitionAccount: {
      id: 'ta-10001',
      totalQuantity: 12,
      totalFreeQuantity: 0,
      remainQuantity: 8,
      remainFreeQuantity: 0,
      lessonChargingMode: 1,
    },
    remark: '',
    teacherList: [
      { teacherId: 'teacher-001', name: '李老师', status: 1, classId: '10001', isDefault: true },
      { teacherId: 'teacher-002', name: '周老师', status: 1, classId: '10001' },
      { teacherId: 'teacher-003', name: '孙老师', status: 1, classId: '10001' },
    ],
  },
  {
    id: '10002',
    name: '李欣然-认知理解1对1',
    studentName: '李欣然',
    studentId: 'stu-002',
    lessonId: 'lesson-002',
    lessonName: '认知理解训练',
    classRoomId: 'room-05',
    classRoomName: '个训室 05',
    defaultTeacherId: 'teacher-001',
    defaultTeacherName: '李老师',
    classTeacherId: 'advisor-002',
    classTeacherName: '黄老师',
    status: 1,
    classStudentStatus: 1,
    isScheduled: true,
    classTime: 45,
    studentClassTime: 1,
    teacherClassTime: 0,
    defaultClassTimeRecordMode: 1,
    createdTime: '2026-03-18 11:05:00',
    tuitionAccountCount: 2,
    tuitionAccount: {
      id: 'ta-10002',
      totalQuantity: 10,
      totalFreeQuantity: 0,
      remainQuantity: 6,
      remainFreeQuantity: 0,
      lessonChargingMode: 2,
    },
    remark: '优先稳定老师，家长可接受节假日顺延。',
    teacherList: [
      { teacherId: 'teacher-001', name: '李老师', status: 1, classId: '10002', isDefault: true },
      { teacherId: 'teacher-003', name: '黄老师', status: 1, classId: '10002' },
    ],
  },
  {
    id: '10003',
    name: '周可心-精细动作1对1',
    studentName: '周可心',
    studentId: 'stu-003',
    lessonId: 'lesson-003',
    lessonName: '精细动作干预',
    classRoomId: 'room-07',
    classRoomName: '感统室 01',
    defaultTeacherId: 'teacher-002',
    defaultTeacherName: '周老师',
    classTeacherId: '',
    classTeacherName: '',
    status: 1,
    classStudentStatus: 2,
    isScheduled: false,
    classTime: 45,
    studentClassTime: 1,
    teacherClassTime: 0,
    defaultClassTimeRecordMode: 2,
    createdTime: '2026-03-12 09:40:00',
    tuitionAccountCount: 1,
    tuitionAccount: {
      id: 'ta-10003',
      totalQuantity: 20,
      totalFreeQuantity: 2,
      remainQuantity: 11,
      remainFreeQuantity: 1,
      lessonChargingMode: 1,
      status: 2,
    },
    remark: '当前处于停课中，恢复后再排课。',
    teacherList: [
      { teacherId: 'teacher-001', name: '李老师', status: 1, classId: '10003' },
      { teacherId: 'teacher-002', name: '周老师', status: 1, classId: '10003', isDefault: true },
    ],
  },
]

const selectedOneToOneId = ref(oneToOneRecords[0]?.id || '')
const scheduleType = ref<ScheduleType>('oneToOne')
const schedulingMode = ref<SchedulingMode>('repeat')
const repeatRule = ref<RepeatRule>('weekly')
const holidayPolicy = ref<HolidayPolicy>('filter')
const timeMode = ref<TimeMode>('school')
const selectedWeekdays = ref(['周一', '周三', '周五'])
/** 同一选课可多次勾选（例如上午一节 + 下午一节），每个上课日按所选时段各生成一节 */
const selectedSchoolTimeSlots = ref<string[]>(['slot-d', 'slot-e'])
const selectedTeacher = ref(oneToOneRecords[0]?.defaultTeacherName || '')
const selectedAssistant = ref<string[] | undefined>(undefined)
const selectedClassroom = ref(oneToOneRecords[0]?.classRoomName || '')
const selectedStudentLessonPath = ref<string[] | undefined>(undefined)
const previewModalOpen = ref(false)
const scheduleStartDate = ref(dayjs().add(5, 'day').startOf('day'))
const freeSelectedDates = ref<Dayjs[]>([dayjs().add(5, 'day').startOf('day')])
const freeCalendarPanelDate = ref(dayjs().add(5, 'day').startOf('month'))
const freeCalendarOpen = ref(false)
const plannedClassCount = ref(1)
const customTimeRanges = ref<CustomTimeRangeRow[]>([
  {
    start: dayjs().hour(10).minute(30).second(0),
    end: dayjs().hour(11).minute(15).second(0),
  },
  {
    start: dayjs().hour(14).minute(0).second(0),
    end: dayjs().hour(14).minute(45).second(0),
  },
])

const selectedOneToOne = computed(() =>
  oneToOneRecords.find(item => item.id === selectedOneToOneId.value) || oneToOneRecords[0],
)

const oneToOneSelectOptions = computed(() =>
  oneToOneRecords.map(item => ({
    value: item.id || '',
    label: `${item.name || '-'} · 课程：${item.lessonName || '-'}`,
    desc: `${item.defaultTeacherName || '-'} · ${item.classRoomName || '-'} · ${item.tuitionAccountCount ?? 0} 个账户`,
  })),
)

const studentLessonCascaderOptions = computed(() => {
  const grouped = new Map<string, { value: string, label: string, children: { value: string, label: string }[] }>()

  oneToOneRecords.forEach((item) => {
    const studentValue = item.studentId || `student-${item.id || item.name || Math.random()}`
    if (!grouped.has(studentValue)) {
      grouped.set(studentValue, {
        value: studentValue,
        label: item.studentName || '-',
        children: [],
      })
    }

    grouped.get(studentValue)?.children.push({
      value: item.id || '',
      label: `课程：${item.lessonName || '-'}`,
    })
  })

  return [...grouped.values()]
})

const selectedTeacherOptions = computed(() => {
  const teacherSet = new Set<string>()
  selectedOneToOne.value?.teacherList?.forEach((item) => {
    if (item?.name)
      teacherSet.add(item.name)
  })
  if (selectedOneToOne.value?.defaultTeacherName)
    teacherSet.add(selectedOneToOne.value.defaultTeacherName)
  return [...teacherSet]
})

const assistantOptions = computed(() =>
  assistantPool.map(item => ({ value: item, label: item })),
)

const classroomOptions = computed(() => {
  const classroomSet = new Set<string>()
  if (selectedOneToOne.value?.classRoomName)
    classroomSet.add(selectedOneToOne.value.classRoomName)
  sharedClassrooms.forEach(item => classroomSet.add(item))
  return [...classroomSet].map(item => ({ value: item, label: item }))
})

const recordSessionMinutes = computed(() => Math.max(Number(selectedOneToOne.value?.classTime || 45), 1))

watch(
  selectedOneToOne,
  (value) => {
    const teacherNames = selectedTeacherOptions.value
    selectedTeacher.value = value?.defaultTeacherName && teacherNames.includes(value.defaultTeacherName)
      ? value.defaultTeacherName
      : (teacherNames[0] || '')
    selectedClassroom.value = value?.classRoomName || classroomOptions.value[0]?.value || ''
    selectedAssistant.value = undefined
    const s = dayjs().hour(10).minute(30).second(0)
    customTimeRanges.value = [
      { start: s, end: s.add(recordSessionMinutes.value, 'minute') },
      {
        start: dayjs().hour(14).minute(0).second(0),
        end: dayjs().hour(14).minute(45).second(0),
      },
    ]
    selectedSchoolTimeSlots.value = ['slot-d', 'slot-e']
    freeSelectedDates.value = [dayjs().add(5, 'day').startOf('day')]
    freeCalendarPanelDate.value = freeSelectedDates.value[0].startOf('month')
  },
  { immediate: true },
)

watch(selectedOneToOneId, (value) => {
  const current = oneToOneRecords.find(item => item.id === value)
  if (!current?.studentId || !current?.id) {
    selectedStudentLessonPath.value = undefined
    return
  }
  selectedStudentLessonPath.value = [current.studentId, current.id]
  const ta = current?.tuitionAccount
  const remain = Number(ta?.remainQuantity || 0) + Number(ta?.remainFreeQuantity || 0)
  plannedClassCount.value = remain > 0 ? Math.max(1, Math.ceil(remain)) : 1
}, { immediate: true })

watch(modalOpen, (value) => {
  if (!value)
    previewModalOpen.value = false
})

const oneToOneStatusText = computed(() =>
  Number(selectedOneToOne.value?.status) === 2 ? '已结班' : '开班中',
)

const oneToOneStudentStatusText = computed(() => {
  const status = Number(selectedOneToOne.value?.classStudentStatus)
  if (status === 2)
    return '停课中'
  if (status === 3)
    return '已结课'
  return '开课中'
})

const recordModeText = computed(() =>
  Number(selectedOneToOne.value?.defaultClassTimeRecordMode) === 2 ? '按上课时长记录' : '按固定课时记录',
)

const effectiveLessonChargingMode = computed(() => {
  const ta = selectedOneToOne.value?.tuitionAccount
  const mode = Number(ta?.lessonChargingMode || 0)
  if (mode > 0)
    return mode
  if (ta?.enableExpireTime && Number(ta?.totalQuantity || 0) > 0)
    return 2
  return mode || 0
})

const quantityUnit = computed(() => {
  if (effectiveLessonChargingMode.value === 1)
    return '课时'
  if (effectiveLessonChargingMode.value === 2)
    return '天'
  if (effectiveLessonChargingMode.value === 3)
    return '元'
  return ''
})

function formatBalanceValue(value: number) {
  if (!Number.isFinite(value))
    return '--'
  return `${value.toFixed(2)}${quantityUnit.value ? ` ${quantityUnit.value}` : ''}`
}

const recordClassroomText = computed(() => selectedOneToOne.value?.classRoomName || '-')
const scheduledClassroomText = computed(() => selectedClassroom.value || '-')
const selectedAssistantText = computed(() =>
  selectedAssistant.value?.length ? selectedAssistant.value.join('、') : '未安排',
)

const usedQuantityValue = computed(() => {
  const ta = selectedOneToOne.value?.tuitionAccount
  const total = Number(ta?.totalQuantity || 0) + Number(ta?.totalFreeQuantity || 0)
  const remain = Number(ta?.remainQuantity || 0) + Number(ta?.remainFreeQuantity || 0)
  return Math.max(total - remain, 0)
})

const remainQuantityValue = computed(() => {
  const ta = selectedOneToOne.value?.tuitionAccount
  return Number(ta?.remainQuantity || 0) + Number(ta?.remainFreeQuantity || 0)
})

/** 剩余可排课节数上限（与学费账户粒度一致，向上取整） */
const remainSessionCap = computed(() => {
  const r = remainQuantityValue.value
  return r > 0 ? Math.ceil(r) : 0
})

function slotDurationMinutes(start: string, end: string) {
  const [sh, sm] = start.split(':').map(Number)
  const [eh, em] = end.split(':').map(Number)
  return Math.max(0, eh * 60 + em - (sh * 60 + sm))
}

const activeTimeBlocks = computed<TimeBlock[]>(() => {
  if (timeMode.value === 'school') {
    const rows: (TimeBlock & { sortKey: string })[] = []
    for (const id of selectedSchoolTimeSlots.value) {
      const slot = schoolTimeSlotOptions.find(s => s.value === id)
      if (!slot)
        continue
      rows.push({
        key: slot.value,
        rangeText: `${slot.desc} · ${slot.start}-${slot.end}`,
        minutes: slotDurationMinutes(slot.start, slot.end),
        sortKey: slot.start,
      })
    }
    return rows
      .sort((a, b) => a.sortKey.localeCompare(b.sortKey))
      .map(({ sortKey: _s, ...r }) => r)
  }
  const rows: (TimeBlock & { sortKey: string })[] = []
  customTimeRanges.value.forEach((row, index) => {
    if (!row.start || !row.end || !row.end.isAfter(row.start))
      return
    rows.push({
      key: `custom-${index}-${row.start.valueOf()}`,
      rangeText: `${row.start.format('HH:mm')} - ${row.end.format('HH:mm')}`,
      minutes: row.end.diff(row.start, 'minute'),
      sortKey: row.start.format('HH:mm'),
    })
  })
  return rows
    .sort((a, b) => a.sortKey.localeCompare(b.sortKey))
    .map(({ sortKey: _s, ...r }) => r)
})

const scheduleSessionMinutesText = computed(() => {
  const blocks = activeTimeBlocks.value
  if (!blocks.length)
    return '--'
  const uniq = [...new Set(blocks.map(b => `${b.minutes} 分钟`))]
  return uniq.join('、')
})

const selectedWeekdaysText = computed(() => selectedWeekdays.value.join(' / '))
const freeSelectedDatesSorted = computed(() =>
  [...freeSelectedDates.value].sort((a, b) => a.valueOf() - b.valueOf()),
)

const freeSelectedDateKeys = computed(() =>
  new Set(freeSelectedDatesSorted.value.map(item => item.format('YYYY-MM-DD'))),
)

const freeSelectedDatesText = computed(() => {
  const dates = freeSelectedDatesSorted.value.map(item => item.format('YYYY-MM-DD'))
  if (dates.length === 0)
    return '请选择上课日期'
  return dates.join('、')
})

const scheduleTargetCount = computed(() => {
  const slots = Math.max(activeTimeBlocks.value.length, 1)
  if (schedulingMode.value === 'free')
    return freeSelectedDatesSorted.value.length * slots
  const planned = Math.max(0, Math.floor(Number(plannedClassCount.value) || 0))
  if (planned < 1)
    return 0
  if (repeatRule.value === 'none')
    return Math.min(planned, slots)
  if (schedulingMode.value === 'repeat')
    return planned
  return 0
})

const repeatRuleText = computed(() => {
  if (schedulingMode.value === 'free')
    return '自由排课 · 单次日程'
  const base = repeatRuleLabelMap[repeatRule.value]
  if (repeatRule.value === 'weekly' || repeatRule.value === 'biweekly')
    return `${base} · ${selectedWeekdaysText.value || '-'}`
  return base
})

const dateSettingText = computed(() => {
  if (schedulingMode.value === 'free')
    return freeSelectedDatesText.value
  return rangeText.value
})

const timeModeText = computed(() => {
  const blocks = activeTimeBlocks.value
  if (!blocks.length)
    return timeMode.value === 'school' ? '请选择课表节次' : '请配置自定义时段'
  const blocksDesc = blocks.map(b => b.rangeText).join('；')
  if (timeMode.value === 'school') {
    const n = blocks.length
    const head = n > 1 ? `课表节次（共 ${n} 节）` : '课表节次'
    return `${head} · ${blocksDesc}`
  }
  const n = blocks.length
  const head = n > 1 ? `自定义时段（共 ${n} 节）` : '自定义时段'
  return `${head} · ${blocksDesc}`
})

const selectedOneToOneSummary = computed<SummaryItem[]>(() => [
  { label: '学员', value: selectedOneToOne.value?.studentName || '-' },
  { label: '课程', value: selectedOneToOne.value?.lessonName || '-' },
  { label: '默认老师', value: selectedOneToOne.value?.defaultTeacherName || '-' },
  { label: '班主任', value: selectedOneToOne.value?.classTeacherName || '-' },
  { label: '默认教室', value: recordClassroomText.value },
  { label: '档案节课时长', value: `${recordSessionMinutes.value} 分钟` },
  { label: '本次时段', value: scheduleSessionMinutesText.value },
  { label: '记录方式', value: recordModeText.value },
  { label: '学费账户', value: `${selectedOneToOne.value?.tuitionAccountCount ?? 0} 个` },
])

const overviewItems = computed<SummaryItem[]>(() => [
  { label: '排课类型', value: scheduleType.value === 'oneToOne' ? '按1对1' : '按学员和课程' },
  { label: '排课方式', value: schedulingMode.value === 'repeat' ? '重复排课' : '自由排课' },
  { label: '日期设置', value: dateSettingText.value },
  {
    label: schedulingMode.value === 'free' ? '已选日期' : '计划上课次数',
    value: schedulingMode.value === 'free'
      ? `${freeSelectedDatesSorted.value.length} 天，共 ${scheduleTargetCount.value} 节`
      : `${Math.floor(Number(plannedClassCount.value) || 0)} 节（手填；剩余可排约 ${remainSessionCap.value || 0} 节）`,
  },
  { label: '重复规则', value: repeatRuleText.value },
  { label: '上课时间', value: timeModeText.value },
  { label: '上课老师', value: selectedTeacher.value || '-' },
  { label: '上课助教', value: selectedAssistantText.value },
  { label: '上课教室', value: scheduledClassroomText.value },
])

/** 侧边「创建摘要」里 value 会被单行省略，这些行悬停展示全文 */
const overviewTooltipLabels = new Set(['重复规则', '计划上课次数', '已选日期', '上课时间', '上课教室'])

const planExceedsRemainHint = computed(() => {
  if (schedulingMode.value === 'free') {
    const cap = remainSessionCap.value
    const planned = scheduleTargetCount.value
    if (planned < 1)
      return ''
    if (cap <= 0)
      return '当前剩余可排为 0，已选日期仅用于预估；正式创建前请处理账户。'
    if (planned > cap)
      return `当前已选 ${planned} 节，多于剩余可排 ${cap} 节；正式创建时请核对账户或减少日期。`
    return ''
  }
  const cap = remainSessionCap.value
  const planned = Math.floor(Number(plannedClassCount.value) || 0)
  if (planned < 1)
    return ''
  if (cap <= 0)
    return '当前剩余可排为 0，仍可按手填节数推算结束日期；正式创建前请处理账户。'
  if (planned > cap)
    return `当前手填 ${planned} 节，多于剩余可排 ${cap} 节；结束日期仍按手填次数推算，正式创建时请核对账户/是否截断。`
  return ''
})

const recordQuickTags = computed<QuickTag[]>(() => [
  { text: oneToOneStatusText.value, tone: 'primary' },
  {
    text: oneToOneStudentStatusText.value,
    tone: oneToOneStudentStatusText.value === '开课中' ? 'neutral' : 'warning',
  },
])

const blockedReason = computed(() => {
  if (Number(selectedOneToOne.value?.status) === 2)
    return '当前 1 对 1 已结班，暂不可创建新日程。'
  if (Number(selectedOneToOne.value?.classStudentStatus) === 2)
    return '当前学员处于停课中，恢复后再创建日程。'
  if (Number(selectedOneToOne.value?.classStudentStatus) === 3)
    return '当前学员已结课，暂不可创建新日程。'
  if (!selectedTeacher.value)
    return '请先选择上课教师。'
  if (schedulingMode.value === 'free' && freeSelectedDatesSorted.value.length === 0)
    return '请至少选择一个上课日期。'
  const planned = Math.floor(Number(plannedClassCount.value) || 0)
  if (schedulingMode.value !== 'free' && planned < 1)
    return '请填写计划上课次数（至少为 1）。'
  if (timeMode.value === 'school' && selectedSchoolTimeSlots.value.length === 0)
    return '请至少选择一节课表节次。'
  if (timeMode.value === 'custom') {
    if (customTimeRanges.value.length === 0)
      return '请至少添加一行自定义时段。'
    for (const row of customTimeRanges.value) {
      if (!row.start || !row.end)
        return '请补全每一行自定义时段的开始与结束时间。'
      if (!row.end.isAfter(row.start))
        return '每一行自定义时段的结束时间需晚于开始时间。'
    }
  }
  if (!activeTimeBlocks.value.length)
    return '当前没有可用的上课时段，请检查时间配置。'
  if (schedulingMode.value === 'repeat' && (repeatRule.value === 'weekly' || repeatRule.value === 'biweekly') && selectedWeekdays.value.length === 0)
    return '请至少选择一个每周上课日。'
  return ''
})

const isSchedulable = computed(() => !blockedReason.value)

function isHoliday(date: Dayjs) {
  return schoolHolidaySet.has(date.format('YYYY-MM-DD'))
}

const rawPlannedDates = computed(() => {
  if (schedulingMode.value === 'free')
    return freeSelectedDatesSorted.value.map(item => item.startOf('day'))

  const start = scheduleStartDate.value
  if (!start)
    return []

  const rangeStart = start.startOf('day')
  const sessionCap = scheduleTargetCount.value
  if (sessionCap <= 0)
    return []
  if (repeatRule.value === 'none')
    return [rangeStart]

  const slotsPerDay = Math.max(activeTimeBlocks.value.length, 1)
  const dateTarget = Math.max(1, Math.ceil(sessionCap / slotsPerDay))

  const result: Dayjs[] = []
  let cursor = rangeStart
  let guard = 0

  while (result.length < dateTarget && guard < 5000) {
    if (repeatRule.value === 'daily') {
      result.push(cursor)
    }
    else if (repeatRule.value === 'alternateDay') {
      const diff = cursor.diff(rangeStart, 'day')
      if (diff % 2 === 0)
        result.push(cursor)
    }
    else {
      const selectedNumbers = new Set(selectedWeekdays.value.map(day => weekdayToNumber[day]))
      const matchedWeekday = selectedNumbers.has(cursor.day())
      const weekDiff = Math.floor(cursor.startOf('day').diff(rangeStart, 'day') / 7)
      if (matchedWeekday && (repeatRule.value === 'weekly' || weekDiff % 2 === 0))
        result.push(cursor)
    }
    cursor = cursor.add(1, 'day')
    guard += 1
  }

  return result
})

const excludedHolidayCount = computed(() => {
  if (holidayPolicy.value !== 'filter')
    return 0
  return rawPlannedDates.value.filter(date => isHoliday(date)).length
})

const plannedDates = computed(() => {
  if (holidayPolicy.value !== 'filter')
    return rawPlannedDates.value
  return rawPlannedDates.value.filter(date => !isHoliday(date))
})

const autoScheduleEndDate = computed(() => {
  if (schedulingMode.value === 'free')
    return freeSelectedDatesSorted.value[freeSelectedDatesSorted.value.length - 1] || scheduleStartDate.value
  return plannedDates.value[plannedDates.value.length - 1] || scheduleStartDate.value
})

const rangeText = computed(() =>
  `${scheduleStartDate.value.format('YYYY-MM-DD')} 至 ${autoScheduleEndDate.value.format('YYYY-MM-DD')}`,
)

const previewRuleText = computed(() => {
  if (schedulingMode.value === 'free')
    return '单次日程'
  if (repeatRule.value === 'weekly' || repeatRule.value === 'biweekly')
    return `${repeatRuleLabelMap[repeatRule.value]} · ${selectedWeekdaysText.value || '-'}`
  return repeatRuleLabelMap[repeatRule.value]
})

const previewPlans = computed<PreviewItem[]>(() => {
  const blocks = activeTimeBlocks.value
  if (!blocks.length)
    return []
  const tone: PreviewTone = isSchedulable.value ? 'pending' : 'blocked'
  const unclipped = plannedDates.value.flatMap(date =>
    blocks.map(block => ({
      date: date.format('YYYY-MM-DD'),
      week: weekDisplayMap[date.day()] || '-',
      rule: previewRuleText.value,
      time: block.rangeText,
      teacher: selectedTeacher.value || selectedOneToOne.value?.defaultTeacherName || '-',
      classroom: scheduledClassroomText.value,
      tone,
    })),
  )
  const cap = scheduleTargetCount.value
  if (cap <= 0)
    return []
  return unclipped.slice(0, cap)
})

const estimatedCount = computed(() => previewPlans.value.length)

const previewHelperText = computed(() => {
  if (blockedReason.value)
    return blockedReason.value
  if (!estimatedCount.value && excludedHolidayCount.value > 0)
    return '当前日期都命中节假日且已被过滤，请调整日期或关闭节假日过滤。'
  if (!estimatedCount.value)
    return '请先选择有效的排课日期。'
  if (excludedHolidayCount.value > 0)
    return `已根据节假日规则过滤 ${excludedHolidayCount.value} 节，剩余 ${estimatedCount.value} 节待创建。`
  return '正式创建前会再做一次资源校验，确认老师、教室和时间可用。'
})

const footerTipText = computed(() => {
  if (selectedOneToOne.value?.remark)
    return `档案备注：${selectedOneToOne.value.remark}`
  return '创建后仍可在日程列表中继续调整老师、教室和具体时间。'
})

const actionButtonText = computed(() => {
  if (schedulingMode.value === 'free')
    return estimatedCount.value > 0 ? '创建单次日程' : '创建日程'
  return estimatedCount.value > 0 ? `批量创建 ${estimatedCount.value} 节` : '批量创建日程'
})

/** 多选时段超出可视区域时展示 +N（配合 maxTagCount=responsive） */
function schoolSlotMaxTagPlaceholder(omittedValues: { label?: unknown, value?: unknown }[]) {
  const n = omittedValues?.length ?? 0
  return n > 0 ? `+${n}` : ''
}

function closeModal() {
  previewModalOpen.value = false
  modalOpen.value = false
}

function openPreviewModal() {
  if (!isSchedulable.value || estimatedCount.value === 0)
    return
  previewModalOpen.value = true
}

function closePreviewModal() {
  previewModalOpen.value = false
}

function confirmBatchCreate() {
  previewModalOpen.value = false
  modalOpen.value = false
}

function handleStudentLessonChange(value: string[]) {
  selectedStudentLessonPath.value = value?.length ? value : undefined
  if (value?.length >= 2)
    selectedOneToOneId.value = value[1]
}

function toggleFreeScheduleDate(date: Dayjs) {
  const key = date.startOf('day').format('YYYY-MM-DD')
  const exists = freeSelectedDates.value.some(item => item.startOf('day').format('YYYY-MM-DD') === key)
  freeSelectedDates.value = exists
    ? freeSelectedDates.value.filter(item => item.startOf('day').format('YYYY-MM-DD') !== key)
    : [...freeSelectedDates.value, date.startOf('day')]
}

function handleFreeCalendarPanelChange(date: Dayjs) {
  freeCalendarPanelDate.value = date.startOf('month')
}

function clearFreeSelectedDates() {
  freeSelectedDates.value = []
}

function toggleWeekday(day: string) {
  const active = selectedWeekdays.value.includes(day)
  if (active && selectedWeekdays.value.length === 1)
    return

  selectedWeekdays.value = active
    ? selectedWeekdays.value.filter(item => item !== day)
    : [...selectedWeekdays.value, day]
}

const isAllWeekdaysSelected = computed(() => selectedWeekdays.value.length === weekDayOptions.length)

function toggleAllWeekdays() {
  selectedWeekdays.value = isAllWeekdaysSelected.value ? [] : [...weekDayOptions]
}

function invertWeekdays() {
  const inverted = weekDayOptions.filter(item => !selectedWeekdays.value.includes(item))
  selectedWeekdays.value = inverted
}

watch(timeMode, (value) => {
  if (value === 'custom' && customTimeRanges.value.length === 0) {
    const s = dayjs().hour(10).minute(30).second(0)
    customTimeRanges.value = [
      { start: s, end: s.add(recordSessionMinutes.value, 'minute') },
    ]
  }
})

function disabledCustomEndTime(start: Dayjs | null) {
  if (!start) {
    return {
      disabledHours: () => [],
      disabledMinutes: () => [],
      disabledSeconds: () => [],
    }
  }

  const startHour = start.hour()
  const startMinute = start.minute()

  return {
    disabledHours: () => Array.from({ length: startHour }, (_, index) => index),
    disabledMinutes: (selectedHour: number) => {
      if (selectedHour === startHour)
        return Array.from({ length: startMinute + 1 }, (_, index) => index)
      return []
    },
    disabledSeconds: () => [],
  }
}

function addCustomTimeRow() {
  const last = customTimeRanges.value[customTimeRanges.value.length - 1]
  const base = last?.end?.isValid()
    ? last.end!
    : dayjs().hour(14).minute(0).second(0)
  customTimeRanges.value = [
    ...customTimeRanges.value,
    {
      start: base,
      end: base.add(recordSessionMinutes.value, 'minute'),
    },
  ]
}

function removeCustomTimeRow(index: number) {
  if (customTimeRanges.value.length <= 1)
    return
  customTimeRanges.value = customTimeRanges.value.filter((_, i) => i !== index)
}
</script>

<template>
  <div class="one-to-one-schedule-modal-root">
    <a-modal
      v-model:open="modalOpen"
      centered
      class="one-to-one-schedule-modal"
      :width="1140"
      :body-style="{ padding: '0' }"
      :keyboard="false"
      :closable="false"
      :mask-closable="true"
      @cancel="closeModal"
    >
      <template #title>
        <div class="planner-head">
          <div class="planner-head__main">
            <div class="planner-head__title">
              创建1对1日程
            </div>
            <div class="planner-head__subtitle">
              参考当前1对1档案快速生成日程，先把规则配置清楚，再确认创建。
            </div>
          </div>

          <div class="planner-head__stats">
            <div class="planner-stat">
              <strong>{{ estimatedCount }}</strong>
              <span>节预计生成</span>
            </div>
            <div v-if="excludedHolidayCount" class="planner-stat planner-stat--subtle">
              <strong>{{ excludedHolidayCount }}</strong>
              <span>节已过滤</span>
            </div>
          </div>

          <a-button type="text" class="planner-head__close" @click="closeModal">
            <template #icon>
              <CloseOutlined />
            </template>
          </a-button>
        </div>
      </template>

      <template #footer>
        <div class="planner-footer">
          <div class="planner-footer__tip">
            {{ footerTipText }}
          </div>

          <div class="planner-footer__actions">
            <a-button @click="closeModal">
              取消
            </a-button>
            <a-button type="primary" :disabled="!isSchedulable || estimatedCount === 0" @click="openPreviewModal">
              {{ actionButtonText }}
            </a-button>
          </div>
        </div>
      </template>

      <div class="planner-shell">
        <section class="planner-strip planner-strip--top">
          <div class="planner-inline planner-inline--top">
            <div class="planner-inline__group planner-inline__group--type">
              <span class="planner-label planner-label--inline">排课类型</span>
              <div class="planner-choice-row planner-choice-row--type-inline">
                <button
                  v-for="item in scheduleTypeOptions"
                  :key="item.value"
                  type="button"
                  class="planner-choice planner-choice--type"
                  :class="{ 'planner-choice--active': scheduleType === item.value }"
                  @click="scheduleType = item.value"
                >
                  <span class="planner-choice__title">{{ item.label }}</span>
                </button>
              </div>
            </div>

            <div class="planner-inline__group planner-inline__group--record">
              <span class="planner-label planner-label--inline">
                <BookOutlined />
                {{ scheduleType === 'oneToOne' ? '选择1对1' : '选择学员和课程' }}
              </span>
              <a-select
                v-if="scheduleType === 'oneToOne'"
                v-model:value="selectedOneToOneId"
                size="large"
                show-search
                option-filter-prop="label"
                option-label-prop="label"
                popup-class-name="planner-record-select-dropdown"
                :allow-clear="false"
                class="planner-control planner-control--record"
              >
                <a-select-option
                  v-for="item in oneToOneSelectOptions"
                  :key="item.value"
                  :value="item.value"
                  :label="item.label"
                >
                  <div class="planner-option">
                    <div class="planner-option__title">
                      {{ item.label }}
                    </div>
                    <div class="planner-option__desc">
                      {{ item.desc }}
                    </div>
                  </div>
                </a-select-option>
              </a-select>
              <a-cascader
                v-else
                v-model:value="selectedStudentLessonPath"
                :options="studentLessonCascaderOptions"
                placeholder="请选择学员和课程"
                class="planner-control planner-control--record"
                @change="handleStudentLessonChange"
              />
            </div>

            <div class="planner-inline__group planner-inline__group--status">
              <span class="planner-label planner-label--inline">账户余额</span>
              <div class="planner-balance">
                <div class="planner-balance__item">
                  <span>已用数量</span>
                  <strong>{{ formatBalanceValue(usedQuantityValue) }}</strong>
                </div>
                <div class="planner-balance__item">
                  <span>剩余数量</span>
                  <strong>{{ formatBalanceValue(remainQuantityValue) }}</strong>
                </div>
              </div>
            </div>
          </div>
        </section>

        <div class="planner-layout">
          <aside class="planner-aside">
            <section class="planner-card planner-card--summary">
              <div class="planner-card__head">
                <div class="planner-card__title">
                  当前档案
                </div>
                <div class="planner-card__desc">
                  来自当前1对1记录的基础信息与创建摘要。
                </div>
              </div>

              <div class="planner-profile">
                <div class="planner-profile__name">
                  {{ selectedOneToOne?.name || '-' }}
                </div>
                <div class="planner-profile__meta">
                  {{ selectedOneToOne?.studentName || '-' }} · {{ selectedOneToOne?.lessonName || '-' }}
                </div>
              </div>

              <div class="planner-summary-list">
                <div
                  v-for="item in selectedOneToOneSummary"
                  :key="item.label"
                  class="planner-summary-list__row"
                >
                  <span>{{ item.label }}</span>
                  <strong class="planner-summary-list__value">{{ item.value }}</strong>
                </div>
              </div>

              <div class="planner-card__subhead">
                创建摘要
              </div>

              <div class="planner-summary-list planner-summary-list--secondary">
                <div
                  v-for="item in overviewItems"
                  :key="item.label"
                  class="planner-summary-list__row"
                >
                  <span>{{ item.label }}</span>
                  <a-tooltip v-if="overviewTooltipLabels.has(item.label)" :title="item.value">
                    <span class="planner-summary-list__value-wrap">
                      <strong class="planner-summary-list__value">{{ item.value }}</strong>
                    </span>
                  </a-tooltip>
                  <strong v-else class="planner-summary-list__value">{{ item.value }}</strong>
                </div>
              </div>

              <div v-if="selectedOneToOne?.remark" class="planner-note">
                {{ selectedOneToOne.remark }}
              </div>

              <div v-if="planExceedsRemainHint" class="planner-alert planner-alert--soft">
                {{ planExceedsRemainHint }}
              </div>

              <div v-if="blockedReason" class="planner-alert">
                {{ blockedReason }}
              </div>
            </section>
          </aside>

          <main class="planner-main">
            <section class="planner-card planner-card--form">
              <div class="planner-card__head">
                <div class="planner-card__title">
                  排课设置
                </div>
                <div class="planner-card__desc">
                  按顺序完成排课方式、日期规则和时间资源。
                </div>
              </div>

              <div class="planner-section">
                <div class="planner-section__title">
                  排课方式
                </div>

                <div class="planner-choice-row">
                  <button
                    v-for="item in schedulingModeOptions"
                    :key="item.value"
                    type="button"
                    class="planner-choice"
                    :class="{ 'planner-choice--active': schedulingMode === item.value }"
                    @click="schedulingMode = item.value"
                  >
                    <span class="planner-choice__title">{{ item.label }}</span>
                    <span class="planner-choice__desc">{{ item.desc }}</span>
                  </button>
                </div>
              </div>

              <div class="planner-section">
                <div class="planner-section__title">
                  日期规则
                </div>

                <div class="planner-form-grid">
                  <div
                    v-if="schedulingMode === 'repeat'"
                    class="planner-date-plan-row"
                  >
                    <label class="planner-field planner-date-plan-row__cell planner-date-plan-row__cell--start">
                      <span class="planner-label planner-label--required">
                        <CalendarOutlined />
                        开始日期
                      </span>
                      <a-date-picker
                        v-model:value="scheduleStartDate"
                        size="large"
                        class="planner-control"
                        :allow-clear="false"
                      />
                    </label>

                    <label class="planner-field planner-date-plan-row__cell planner-date-plan-row__cell--count">
                      <span class="planner-label planner-label--required">
                        计划上课次数
                      </span>
                      <a-input-number
                        v-model:value="plannedClassCount"
                        :min="1"
                        :precision="0"
                        size="large"
                        class="planner-control"
                      />
                    </label>

                    <div class="planner-field planner-date-plan-row__cell">
                      <span class="planner-label">
                        <CalendarOutlined />
                        结束日期
                      </span>
                      <div class="planner-static-field planner-static-field--compact planner-static-field--inline-hint">
                        <strong>{{ autoScheduleEndDate.format('YYYY-MM-DD') }}</strong>
                        <span>根据计划上课次数与重复规则自动推算</span>
                      </div>
                    </div>

                    <span class="planner-field__hint planner-date-plan-row__hint">可自由填写节数；结束日期由开始日期、重复规则与本次数推算。参考剩余可排 {{ remainSessionCap }} 节。</span>
                  </div>

                  <div
                    v-else
                    class="planner-date-plan-row planner-date-plan-row--free"
                  >
                    <label class="planner-field planner-date-plan-row__cell planner-date-plan-row__cell--start">
                      <div class="planner-field__label-row">
                        <span class="planner-label planner-label--required">
                          <CalendarOutlined />
                          上课日期
                        </span>
                        <a-button
                          type="link"
                          size="small"
                          :disabled="freeSelectedDatesSorted.length === 0"
                          @click.stop.prevent="clearFreeSelectedDates"
                        >
                          一键清空
                        </a-button>
                      </div>
                      <a-popover
                        v-model:open="freeCalendarOpen"
                        trigger="click"
                        placement="bottomLeft"
                        overlay-class-name="planner-free-calendar-popover"
                      >
                        <template #content>
                          <div class="planner-free-calendar">
                            <a-calendar
                              :fullscreen="false"
                              :value="freeCalendarPanelDate"
                              @panelChange="handleFreeCalendarPanelChange"
                              @select="toggleFreeScheduleDate"
                            >
                              <template #dateFullCellRender="{ current }">
                                <div
                                  class="planner-free-calendar__cell"
                                  :class="{ 'planner-free-calendar__cell--selected': freeSelectedDateKeys.has(current.format('YYYY-MM-DD')) }"
                                >
                                  {{ current.date() }}
                                </div>
                              </template>
                            </a-calendar>
                          </div>
                        </template>

                        <div class="planner-static-field planner-static-field--compact planner-static-field--free-trigger">
                          <strong :title="freeSelectedDatesText">{{ freeSelectedDatesText }}</strong>
                        </div>
                      </a-popover>
                    </label>
                  </div>

                  <div v-if="schedulingMode === 'repeat'" class="planner-field planner-field--full">
                    <span class="planner-label planner-label--required">
                      重复规则
                    </span>
                    <div class="planner-chip-row">
                      <button
                        v-for="item in repeatRuleOptions"
                        :key="item.value"
                        type="button"
                        class="planner-chip"
                        :class="{ 'planner-chip--active': repeatRule === item.value }"
                        @click="repeatRule = item.value"
                      >
                        {{ item.label }}
                      </button>
                    </div>
                  </div>

                <div
                  v-if="schedulingMode === 'repeat' && (repeatRule === 'weekly' || repeatRule === 'biweekly')"
                  class="planner-field planner-field--full"
                >
                  <span class="planner-label planner-label--required">
                    每周上课日
                  </span>
                  <div class="planner-chip-row planner-chip-row--weekday">
                    <button
                      v-for="item in weekDayOptions"
                      :key="item"
                      type="button"
                      class="planner-chip"
                        :class="{ 'planner-chip--active': selectedWeekdays.includes(item) }"
                        @click="toggleWeekday(item)"
                    >
                      {{ item }}
                    </button>
                    <div class="planner-chip-actions">
                      <a-button type="link" size="small" @click="toggleAllWeekdays">
                        {{ isAllWeekdaysSelected ? '取消全选' : '全选' }}
                      </a-button>
                      <a-button type="link" size="small" @click="invertWeekdays">
                        反选
                      </a-button>
                    </div>
                  </div>
                </div>

                  <div class="planner-field planner-field--full">
                    <span class="planner-label planner-label--required">
                      节假日
                      <a-tooltip title="当前示例会按学校节假日配置过滤 2026-05-01 至 2026-05-03。">
                        <QuestionCircleOutlined class="planner-label__tip" />
                      </a-tooltip>
                    </span>
                    <div class="planner-chip-row">
                      <button
                        v-for="item in holidayPolicyOptions"
                        :key="item.value"
                        type="button"
                        class="planner-chip"
                        :class="{ 'planner-chip--active': holidayPolicy === item.value }"
                        @click="holidayPolicy = item.value"
                      >
                        {{ item.label }}
                      </button>
                    </div>
                  </div>
                </div>
              </div>

              <div class="planner-section">
                <div class="planner-section__title">
                  时间与资源
                </div>

                <div class="planner-form-grid">
                  <div class="planner-field planner-field--full">
                    <span class="planner-label planner-label--required">
                      上课时间
                    </span>
                    <div class="planner-choice-row planner-choice-row--time-mode">
                      <button
                        v-for="item in timeModeOptions"
                        :key="item.value"
                        type="button"
                        class="planner-choice"
                        :class="{ 'planner-choice--active': timeMode === item.value }"
                        @click="timeMode = item.value"
                      >
                        <span class="planner-choice__title">{{ item.label }}</span>
                        <span class="planner-choice__desc">{{ item.desc }}</span>
                      </button>
                    </div>
                  </div>

                  <label v-if="timeMode === 'school'" class="planner-field planner-field--major planner-field--full">
                    <span class="planner-label planner-label--required">
                      <ClockCircleOutlined />
                      课表节次（可多选）
                    </span>
                    <a-tooltip title="同一上课日内按所选「第几节课」各生成一节日程；重复排课时每个上课日都会生成这些节，总节数仍受剩余课时/额度约束。">
                      <span class="planner-control-tooltip-wrap">
                        <a-select
                          v-model:value="selectedSchoolTimeSlots"
                          mode="multiple"
                          size="large"
                          allow-clear
                          placeholder="同一天内可勾选多节，例如上午 + 下午"
                          max-tag-count="responsive"
                          :max-tag-text-length="14"
                          :max-tag-placeholder="schoolSlotMaxTagPlaceholder"
                          class="planner-control planner-control--major planner-multi-slot-select"
                        >
                          <a-select-option
                            v-for="item in schoolTimeSlotOptions"
                            :key="item.value"
                            :value="item.value"
                            :label="`${item.desc} · ${item.start} - ${item.end}`"
                          >
                            <div class="planner-option planner-option--inline">
                              <div class="planner-option__title">{{ item.desc }} · {{ item.start }} - {{ item.end }}</div>
                            </div>
                          </a-select-option>
                        </a-select>
                      </span>
                    </a-tooltip>
                  </label>

                  <div v-else class="planner-field planner-field--major planner-field--full">
                    <span class="planner-label planner-label--required">
                      <ClockCircleOutlined />
                      自定义时段（可多行）
                    </span>
                    <div class="planner-custom-time-list">
                      <div
                        v-for="(row, rowIndex) in customTimeRanges"
                        :key="rowIndex"
                        class="planner-time-range planner-time-range--stacked"
                      >
                        <a-time-picker
                          v-model:value="row.start"
                          size="large"
                          format="HH:mm"
                          class="planner-control"
                          allow-clear
                          placeholder="开始"
                          :minute-step="5"
                        />
                        <span class="planner-time-range__sep">至</span>
                        <a-time-picker
                          v-model:value="row.end"
                          size="large"
                          format="HH:mm"
                          class="planner-control"
                          allow-clear
                          placeholder="结束"
                          :disabled="!row.start"
                          :disabled-time="() => disabledCustomEndTime(row.start)"
                          :minute-step="5"
                        />
                        <a-button
                          v-if="customTimeRanges.length > 1"
                          type="link"
                          danger
                          size="small"
                          class="planner-time-range__remove"
                          @click="removeCustomTimeRow(rowIndex)"
                        >
                          删除
                        </a-button>
                      </div>
                      <a-button type="dashed" block class="planner-custom-time-add" @click="addCustomTimeRow">
                        <template #icon>
                          <PlusOutlined />
                        </template>
                        添加时段
                      </a-button>
                    </div>
                  </div>

                  <div class="planner-field planner-field--major">
                    <span class="planner-label">
                      <span class="planner-label__lead-spacer" aria-hidden="true" />
                      各时段课长
                    </span>
                    <div class="planner-static-field planner-static-field--major planner-static-field--inline">
                      <strong>{{ scheduleSessionMinutesText }}</strong>
                      <span>按每行开始/结束时间分别计算</span>
                    </div>
                  </div>

                  <label class="planner-field">
                    <span class="planner-label planner-label--required">
                      <UserOutlined />
                      上课教师
                    </span>
                    <a-select
                      v-model:value="selectedTeacher"
                      size="large"
                      :allow-clear="false"
                      class="planner-control"
                    >
                      <a-select-option v-for="item in selectedTeacherOptions" :key="item" :value="item">
                        {{ item }}
                      </a-select-option>
                    </a-select>
                  </label>

                  <label class="planner-field">
                    <span class="planner-label">
                      <TeamOutlined />
                      上课助教
                    </span>
                    <a-select
                      v-model:value="selectedAssistant"
                      mode="multiple"
                      size="large"
                      allow-clear
                      placeholder="可不选"
                      max-tag-count="responsive"
                      :options="assistantOptions"
                      class="planner-control"
                    />
                  </label>

                  <label class="planner-field">
                    <span class="planner-label">
                      <EnvironmentOutlined />
                      上课教室
                    </span>
                    <a-select
                      v-model:value="selectedClassroom"
                      size="large"
                      allow-clear
                      placeholder="可不选"
                      :options="classroomOptions"
                      class="planner-control"
                    />
                  </label>
                </div>
              </div>
            </section>
          </main>
        </div>

      </div>
    </a-modal>

    <a-modal
      v-model:open="previewModalOpen"
      centered
      class="planner-review-modal"
      :footer="null"
      :width="980"
      :body-style="{ padding: '0 24px 16px' }"
      :keyboard="false"
      :mask-closable="false"
      @cancel="closePreviewModal"
    >
      <template #title>
        <div class="planner-review__head">
          <div>
            <div class="planner-review__title">
              预计排课清单
            </div>
            <div class="planner-review__subtitle">
              先确认本次将创建的日程，再执行批量创建。
            </div>
          </div>
          <div class="planner-review__count">
            共 {{ estimatedCount }} 节
          </div>
        </div>
      </template>

      <div class="planner-review">
        <div
          class="planner-review__tip"
          :class="{ 'planner-review__tip--warning': !isSchedulable }"
        >
          {{ previewHelperText }}
        </div>

        <div class="planner-table-wrap planner-table-wrap--modal">
          <table class="planner-table">
            <thead>
              <tr>
                <th>日期</th>
                <th>星期</th>
                <th>规则</th>
                <th>时间</th>
                <th>老师</th>
                <th>教室</th>
                <th>状态</th>
              </tr>
            </thead>
            <tbody>
              <tr v-if="!previewPlans.length">
                <td colspan="7" class="planner-table__empty">
                  暂无预计日程
                </td>
              </tr>
              <template v-else>
                <tr
                  v-for="(item, planIdx) in previewPlans"
                  :key="`${item.date}-${item.time}-${planIdx}`"
                  :class="{ 'planner-table__row--blocked': item.tone === 'blocked' }"
                >
                  <td>{{ item.date }}</td>
                  <td>{{ item.week }}</td>
                  <td>{{ item.rule }}</td>
                  <td>{{ item.time }}</td>
                  <td>{{ item.teacher }}</td>
                  <td>{{ item.classroom }}</td>
                  <td>
                    <span
                      class="planner-tag planner-tag--table"
                      :class="{ 'planner-tag--warning': item.tone === 'blocked' }"
                    >
                      {{ item.tone === 'blocked' ? '不可创建' : '待校验' }}
                    </span>
                  </td>
                </tr>
              </template>
            </tbody>
          </table>
        </div>

        <div class="planner-review__footer">
          <div class="planner-footer__tip">
            {{ footerTipText }}
          </div>

          <div class="planner-footer__actions">
            <a-button @click="closePreviewModal">
              返回修改
            </a-button>
            <a-button type="primary" :disabled="!isSchedulable || estimatedCount === 0" @click="confirmBatchCreate">
              {{ actionButtonText }}
            </a-button>
          </div>
        </div>
      </div>
    </a-modal>
  </div>
</template>

<style scoped lang="less">
.one-to-one-schedule-modal-root {
  display: contents;
}

.planner-head {
  display: flex;
  align-items: flex-start;
  gap: 12px;
}

.planner-head__main {
  flex: 1;
  min-width: 0;
}

.planner-head__title {
  color: #1f2329;
  font-size: 20px;
  font-weight: 600;
  line-height: 1.25;
}

.planner-head__subtitle {
  margin-top: 2px;
  color: #86909c;
  font-size: 12px;
  line-height: 1.5;
}

.planner-head__stats {
  display: flex;
  align-items: center;
  gap: 8px;
}

.planner-head__close {
  color: #94a3b8;
}

.planner-stat {
  min-width: 108px;
  padding: 8px 10px;
  border-radius: 12px;
  background: linear-gradient(180deg, #f6fbff 0%, #edf6ff 100%);
  text-align: right;
}

.planner-stat strong {
  display: block;
  color: #1677ff;
  font-size: 20px;
  font-weight: 700;
  line-height: 1;
}

.planner-stat span {
  display: block;
  margin-top: 4px;
  color: #6b7785;
  font-size: 11px;
  line-height: 1.4;
}

.planner-stat--subtle {
  background: #f7f8fa;
}

.planner-stat--subtle strong {
  color: #1f2329;
}

.planner-shell {
  --planner-major-height: 42px;
  display: flex;
  max-height: calc(100vh - 170px);
  flex-direction: column;
  gap: 10px;
  overflow: auto;
  padding: 10px 14px 10px;
  background: #f6f8fb;
}

.planner-strip,
.planner-card {
  border: 1px solid #e8edf3;
  border-radius: 16px;
  background: #fff;
  box-shadow: 0 8px 24px rgb(15 23 42 / 4%);
}

.planner-strip {
  padding: 14px 18px;
}

.planner-inline {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.planner-inline--top {
  align-items: center;
  gap: 18px;
}

.planner-inline__group {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: 12px;
}

.planner-inline__group--record {
  flex: 1;
}

.planner-inline__group--status {
  flex-shrink: 0;
  justify-content: flex-end;
}

.planner-inline__group--type {
  flex: 0 0 auto;
}

.planner-control--record {
  flex: 1;
  min-width: 0;
}

.planner-label {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: #4e5969;
  font-size: 13px;
  font-weight: 600;
  white-space: nowrap;
}

.planner-label--inline {
  flex-shrink: 0;
  margin-right: 2px;
}

.planner-label--required::before {
  content: '*';
  color: #ff4d4f;
}

.planner-label__tip {
  color: #9aa4b2;
  font-size: 14px;
}

.planner-tag-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  justify-content: flex-start;
}

.planner-balance {
  display: flex;
  gap: 6px;
}

.planner-balance__item {
  min-width: 100px;
  padding: 6px 10px;
  border: 1px solid #f90;
  border-radius: 10px;
  background: #fff5e6;
}

.planner-balance__item span {
  display: block;
  color: #86909c;
  font-size: 11px;
  line-height: 1.4;
}

.planner-balance__item strong {
  display: block;
  margin-top: 2px;
  color: #1f2329;
  font-size: 14px;
  font-weight: 600;
  line-height: 1.45;
}

.planner-tag {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 30px;
  padding: 0 12px;
  border-radius: 999px;
  background: #f2f4f7;
  color: #4e5969;
  font-size: 12px;
  font-weight: 500;
}

.planner-tag--primary {
  background: #e6f4ff;
  color: #1677ff;
}

.planner-tag--warning {
  background: #fff7e6;
  color: #d46b08;
}

.planner-tag--table {
  min-height: 26px;
  padding: 0 10px;
}

.planner-choice-row {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.planner-choice-row--type-inline {
  display: flex;
  flex-wrap: nowrap;
  gap: 8px;
}

/* 上课时间：双卡片略压低高度，与其它大块表单项区分 */
.planner-choice-row--time-mode button.planner-choice {
  min-height: 54px;
  padding: 8px 12px;
  gap: 4px;
}

.planner-choice-row--time-mode .planner-choice__title {
  font-size: 14px;
  line-height: 1.35;
}

.planner-choice-row--time-mode .planner-choice__desc {
  line-height: 1.45;
}

button.planner-choice {
  display: flex;
  width: 100%;
  min-height: 58px;
  flex-direction: column;
  justify-content: center;
  gap: 4px;
  padding: 10px 12px;
  border: 1px solid #d8e1eb;
  border-radius: 14px;
  background: #fff;
  cursor: pointer;
  appearance: none;
  font: inherit;
  outline: none;
  text-align: left;
  transition:
    border-color 0.2s ease,
    background-color 0.2s ease,
    box-shadow 0.2s ease;
}

button.planner-choice:hover {
  border-color: #91caff;
}

button.planner-choice.planner-choice--active {
  border-color: #1677ff;
  background: #f7fbff;
  box-shadow: 0 0 0 3px rgb(22 119 255 / 8%);
}

.planner-choice__title {
  color: #1f2329;
  font-size: 14px;
  font-weight: 600;
  line-height: 1.5;
}

.planner-choice__desc {
  color: #7d8898;
  font-size: 11px;
  line-height: 1.45;
}

button.planner-choice.planner-choice--active .planner-choice__title {
  color: #1677ff;
}

button.planner-choice.planner-choice--type {
  width: auto;
  min-height: 40px;
  flex-direction: row;
  align-items: center;
  gap: 0;
  padding: 0 14px;
  border-radius: 999px;
}

button.planner-choice.planner-choice--type .planner-choice__title {
  font-size: 14px;
}

.planner-option__title {
  color: #1f2329;
  font-size: 13px;
  line-height: 1.5;
}

.planner-option__desc {
  color: #86909c;
  font-size: 12px;
  line-height: 1.5;
}

.planner-option--inline {
  display: flex;
  align-items: center;
}

.planner-layout {
  display: grid;
  grid-template-columns: 300px minmax(0, 1fr);
  gap: 14px;
  align-items: stretch;
}

.planner-aside,
.planner-main {
  display: flex;
  min-width: 0;
}

.planner-card {
  overflow: hidden;
}

.planner-card--summary,
.planner-card--form {
  display: flex;
  flex: 1;
  flex-direction: column;
}

.planner-card__head {
  padding: 16px 18px 0;
}

.planner-card__title {
  color: #1f2329;
  font-size: 15px;
  font-weight: 600;
  line-height: 1.5;
}

.planner-card__desc {
  margin-top: 4px;
  color: #86909c;
  font-size: 12px;
  line-height: 1.6;
}

.planner-card__subhead {
  margin: 0 18px 10px;
  padding-top: 14px;
  border-top: 1px solid #f1f4f8;
  color: #1f2329;
  font-size: 13px;
  font-weight: 600;
  line-height: 1.5;
}

.planner-profile {
  padding: 14px 18px 0;
}

.planner-profile__name {
  color: #1f2329;
  font-size: 18px;
  font-weight: 600;
  line-height: 1.4;
}

.planner-profile__meta {
  margin-top: 4px;
  color: #86909c;
  font-size: 13px;
  line-height: 1.6;
}

.planner-summary-list {
  display: flex;
  flex-direction: column;
  gap: 0;
  padding: 14px 18px 16px;
}

.planner-summary-list--secondary {
  padding-top: 0;
}

.planner-summary-list__row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 10px 0;
  border-bottom: 1px dashed #edf0f3;
}

.planner-summary-list__row:last-child {
  border-bottom: none;
}

.planner-summary-list__row span {
  flex-shrink: 0;
  width: 72px;
  color: #86909c;
  font-size: 13px;
}

.planner-summary-list__value-wrap {
  display: block;
  min-width: 0;
  flex: 1;
  text-align: right;
  cursor: default;
}

.planner-summary-list__value {
  display: block;
  min-width: 0;
  flex: 1;
  overflow: hidden;
  color: #1f2329;
  font-size: 12px;
  font-weight: 600;
  line-height: 1.6;
  text-align: right;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.planner-summary-list__value-wrap .planner-summary-list__value {
  flex: none;
  width: 100%;
}

.planner-note,
.planner-alert {
  margin: 0 18px 18px;
  padding: 12px 14px;
  border-radius: 12px;
  font-size: 13px;
  line-height: 1.7;
}

.planner-note {
  background: #fafbfc;
  color: #4e5969;
}

.planner-alert {
  background: #fff7e6;
  color: #d46b08;
}

.planner-alert--soft {
  background: #e8f4ff;
  color: #1664c0;
}

.planner-field__hint {
  display: block;
  margin-top: 6px;
  color: #86909c;
  font-size: 12px;
  line-height: 1.55;
}

.planner-section {
  padding: 16px 18px 18px;
}

.planner-section + .planner-section {
  border-top: 1px solid #f1f4f8;
}

.planner-section__title {
  margin-bottom: 10px;
  color: #1f2329;
  font-size: 15px;
  font-weight: 600;
  line-height: 1.5;
}

.planner-form-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14px 16px;
}

.planner-form-grid > .planner-field .planner-control {
  width: 100%;
}

:deep(.planner-form-grid > .planner-field .ant-select) {
  width: 100% !important;
}

.planner-date-plan-row {
  grid-column: 1 / -1;
  display: grid;
  /* 开始日期固定舒适宽度，结束日期列仍用 1fr 吃剩余空间 */
  grid-template-columns: minmax(148px, 200px) minmax(100px, 132px) minmax(0, 1fr);
  column-gap: 14px;
  row-gap: 10px;
  align-items: start;
}

.planner-date-plan-row__cell {
  min-width: 0;
}

.planner-date-plan-row__cell--start {
  width: 100%;
  max-width: 200px;
}

.planner-date-plan-row__cell--start .planner-control {
  width: 100%;
  max-width: 200px;
}

:deep(.planner-date-plan-row__cell--start .ant-picker) {
  width: 100% !important;
  max-width: 200px !important;
}

/* 计划次数输入撑满中间列，避免单元格内右侧留白导致与结束日期间距「看起来」不对称 */
.planner-date-plan-row__cell--count {
  width: 100%;
  min-width: 0;
}

.planner-date-plan-row__cell--count .planner-control {
  width: 100%;
}

:deep(.planner-date-plan-row__cell--count .ant-input-number) {
  width: 100% !important;
}

.planner-date-plan-row__hint {
  grid-column: 1 / -1;
  margin-top: 0;
}

.planner-date-plan-row--free {
  grid-template-columns: minmax(0, 1fr);
  column-gap: 14px;
  row-gap: 10px;
}

.planner-date-plan-row--free .planner-date-plan-row__cell--start {
  max-width: none;
}

.planner-date-plan-row--free .planner-date-plan-row__cell--start .planner-control {
  max-width: none;
}

:deep(.planner-date-plan-row--free .planner-date-plan-row__cell--start .ant-picker) {
  max-width: none !important;
}

.planner-static-field--free-trigger {
  cursor: pointer;
}

.planner-static-field--free-trigger strong {
  display: block;
  font-size: 13px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.planner-static-field--free-trigger span {
  font-size: 11px;
}

.planner-free-calendar {
  width: 300px;
}

.planner-free-calendar__cell {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 34px;
  border-radius: 10px;
  color: #1f2329;
  transition: all 0.18s ease;
}

.planner-free-calendar__cell--selected {
  background: #1677ff;
  color: #fff;
  font-weight: 600;
}

/* 与左侧列「图标 + gap」同宽，使「各时段课长」与「上课助教」文案纵列对齐 */
.planner-label__lead-spacer {
  display: inline-block;
  width: calc(14px + 6px);
  flex-shrink: 0;
}

.planner-date-plan-row__cell .planner-static-field--inline-hint span {
  white-space: normal;
}

.planner-date-plan-row__cell .planner-static-field--inline-hint strong {
  white-space: nowrap;
  flex-shrink: 0;
}

.planner-field {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: 8px;
}

.planner-field--full {
  grid-column: 1 / -1;
}

.planner-field--major {
  align-self: start;
}

.planner-field__label-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.planner-time-range {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto minmax(0, 1fr);
  gap: 10px;
  align-items: center;
}

.planner-time-range__sep {
  color: #98a2b3;
  font-size: 12px;
  white-space: nowrap;
}

.planner-control-tooltip-wrap {
  display: block;
  width: 100%;
  /* 避免 grid/flex 子项 min-width:0 把多选压成一条竖线 */
  min-width: 320px;
}

.planner-custom-time-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
  width: 100%;
}

.planner-time-range--stacked {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 10px;
}

.planner-time-range--stacked .planner-control {
  flex: 1 1 120px;
  min-width: 100px;
}

.planner-time-range__remove {
  flex-shrink: 0;
}

.planner-custom-time-add {
  margin-top: 2px;
}

.planner-chip-row {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.planner-chip-row--weekday {
  align-items: center;
  flex-wrap: nowrap;
  overflow-x: auto;
}

.planner-chip-actions {
  display: inline-flex;
  align-items: center;
  gap: 2px;
  margin-left: 6px;
  flex-shrink: 0;
  white-space: nowrap;
}

button.planner-chip {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 68px;
  height: 34px;
  padding: 0 14px;
  border: 1px solid #d7dfe9;
  border-radius: 999px;
  background: #fff;
  color: #4e5969;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  appearance: none;
  font: inherit;
  outline: none;
  transition:
    color 0.2s ease,
    border-color 0.2s ease,
    background-color 0.2s ease,
    box-shadow 0.2s ease;
}

button.planner-chip:hover {
  border-color: #91caff;
  color: #1677ff;
}

button.planner-chip.planner-chip--active {
  border-color: #1677ff;
  background: #e6f4ff;
  color: #1677ff;
  box-shadow: 0 0 0 3px rgb(22 119 255 / 8%);
}

.planner-static-field {
  display: flex;
  min-height: 42px;
  flex-direction: column;
  justify-content: center;
  gap: 4px;
  padding: 10px 12px;
  border: 1px solid #e5ebf2;
  border-radius: 12px;
  background: #fafbfd;
}

.planner-static-field strong {
  color: #1f2329;
  font-size: 14px;
  font-weight: 600;
  line-height: 1.5;
}

.planner-static-field span {
  color: #86909c;
  font-size: 12px;
  line-height: 1.5;
}

.planner-static-field--compact {
  min-height: 42px;
  justify-content: center;
  padding: 9px 12px;
}

.planner-static-field--compact strong {
  font-size: 14px;
  line-height: 1.35;
}

.planner-static-field--compact span {
  margin-top: 2px;
  font-size: 11px;
  line-height: 1.35;
}

.planner-static-field--inline-hint {
  flex-direction: row;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.planner-static-field--inline-hint span {
  margin-top: 0;
  text-align: right;
  white-space: nowrap;
}

.planner-static-field--major {
  height: var(--planner-major-height);
  padding: 10px 12px;
  border-radius: 12px;
}

.planner-static-field--inline {
  flex-direction: row;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.planner-static-field--inline strong {
  font-size: 14px;
  flex-shrink: 0;
}

.planner-static-field--inline span {
  font-size: 12px;
  text-align: right;
}

.planner-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 14px 6px 0;
}

.planner-footer__tip {
  flex: 1;
  min-width: 0;
  color: #86909c;
  font-size: 12px;
  line-height: 1.6;
}

.planner-footer__actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.planner-review {
  display: flex;
  max-height: calc(100vh - 220px);
  flex-direction: column;
}

.planner-review__head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}

.planner-review__title {
  color: #1f2329;
  font-size: 20px;
  font-weight: 600;
  line-height: 1.4;
}

.planner-review__subtitle {
  margin-top: 4px;
  color: #86909c;
  font-size: 13px;
  line-height: 1.6;
}

.planner-review__count {
  flex-shrink: 0;
  padding-right: 36px;
  color: #1677ff;
  font-size: 16px;
  font-weight: 600;
  line-height: 1.6;
}

.planner-review__tip {
  padding: 12px 14px;
  border-radius: 12px;
  background: #f6f8fa;
  color: #4e5969;
  font-size: 13px;
  line-height: 1.6;
}

.planner-review__tip--warning {
  background: #fff7e6;
  color: #d46b08;
}

.planner-review__footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding-top: 16px;
  border-top: 1px solid #f0f0f0;
}

.planner-table-wrap {
  overflow: auto;
  padding-top: 14px;
}

.planner-table-wrap--modal {
  max-height: min(56vh, 540px);
}

.planner-table {
  width: 100%;
  min-width: 860px;
  border-collapse: separate;
  border-spacing: 0;
}

.planner-table thead th {
  padding: 12px 14px;
  border-bottom: 1px solid #f0f0f0;
  background: #fafafa;
  color: #4e5969;
  font-size: 13px;
  font-weight: 600;
  text-align: left;
  white-space: nowrap;
}

.planner-table tbody td {
  padding: 14px;
  border-bottom: 1px solid #f5f5f5;
  color: #1f2329;
  font-size: 13px;
  line-height: 1.6;
  vertical-align: middle;
}

.planner-table tbody tr:last-child td {
  border-bottom: none;
}

.planner-table__row--blocked td {
  background: #fffaf0;
}

.planner-table__empty {
  padding: 36px 14px !important;
  color: #86909c !important;
  text-align: center;
}

:deep(.one-to-one-schedule-modal .ant-modal-content),
:deep(.planner-review-modal .ant-modal-content) {
  overflow: hidden;
  padding: 0;
  border-radius: 20px;
}

:deep(.one-to-one-schedule-modal .ant-modal-header) {
  margin-bottom: 0;
  padding: 14px 20px 10px;
  border-bottom: none;
}

:deep(.one-to-one-schedule-modal .ant-modal-body) {
  padding: 0;
}

:deep(.one-to-one-schedule-modal .ant-modal-footer) {
  padding: 14px 20px 16px;
  border-top: 1px solid #eef2f6;
  background: #fff;
}




:deep(.planner-review-modal .ant-modal-header) {
  margin-bottom: 0;
  padding: 18px 24px 14px;
  border-bottom: none;
}

:deep(.planner-review-modal .ant-modal-body) {
  padding: 0 24px 16px;
}

:deep(.planner-control.ant-picker),
:deep(.planner-control .ant-select-selector) {
  border-color: #d9e1ea !important;
  border-radius: 12px !important;
  box-shadow: none !important;
  min-height: 42px;
}

:deep(.planner-multi-slot-select.ant-select) {
  width: 100% !important;
  min-width: 320px !important;
}

:deep(.planner-multi-slot-select .ant-select-selector) {
  width: 100% !important;
  min-width: 0;
  flex-wrap: nowrap !important;
  align-items: center !important;
}

:deep(.planner-multi-slot-select .ant-select-selection-overflow) {
  flex-wrap: nowrap !important;
}

:deep(.planner-control--major.ant-picker),
:deep(.planner-control--major .ant-select-selector) {
  height: var(--planner-major-height) !important;
  min-height: var(--planner-major-height) !important;
  padding-inline: 12px !important;
  border-radius: 12px !important;
}

:deep(.planner-multi-slot-select.ant-select-multiple .ant-select-selector) {
  height: auto !important;
  min-height: var(--planner-major-height) !important;
}

:deep(.planner-control--major.ant-picker .ant-picker-input) {
  height: 100%;
}

:deep(.planner-control--major.ant-picker .ant-picker-input > input) {
  height: 100%;
  font-size: 14px;
  font-weight: 400;
}

:deep(.planner-control--major .ant-select-selection-item) {
  display: flex;
  align-items: center;
  font-size: 14px;
  font-weight: 400;
  line-height: 1.45;
  white-space: normal;
}

:deep(.planner-multi-slot-select.ant-select-multiple .ant-select-selection-item) {
  flex-shrink: 0;
  white-space: nowrap !important;
}

:deep(.planner-control--major .ant-select-arrow),
:deep(.planner-control--major.ant-picker .ant-picker-suffix) {
  color: #b0b8c5;
  font-size: 16px;
}

:deep(.planner-control--record .ant-select-selector),
:deep(.planner-control--record .ant-select-selection-item),
:deep(.planner-control--record .ant-select-selection-search),
:deep(.planner-control--record .ant-select-selection-search-input),
:deep(.planner-control--record .ant-select-selection-placeholder) {
  font-size: 14px !important;
  font-weight: 400 !important;
}

:deep(.planner-control--record .ant-select-selector),
:deep(.planner-control--record .ant-cascader-picker),
:deep(.planner-control--record .ant-cascader-picker-label),
:deep(.planner-control--record .ant-cascader-input) {
  display: flex;
  align-items: center;
  min-height: 42px;
}

:deep(.planner-control--record .ant-cascader-picker-label) {
  inset: 0 40px 0 12px;
  line-height: 42px;
}

:deep(.ant-select-selection-item) {
  font-size: 14px !important;
}

:deep(.ant-select-selection-search-input),
:deep(.ant-select-selection-placeholder) {
  font-size: 14px !important;
}

:deep(.planner-control .ant-select-selection-item),
:deep(.planner-control.ant-picker .ant-picker-input > input) {
  font-size: 14px !important;
}

:deep(.planner-record-select-dropdown) {
  padding: 8px;
}

:deep(.planner-record-select-dropdown .ant-select-item) {
  padding: 0;
}

:deep(.planner-record-select-dropdown .ant-select-item-option) {
  min-height: auto;
  margin-bottom: 6px;
  padding: 0;
  border: 1px solid transparent;
  border-radius: 12px;
  background: #fff;
  transition:
    border-color 0.18s ease,
    background-color 0.18s ease,
    box-shadow 0.18s ease;
}

:deep(.planner-record-select-dropdown .ant-select-item-option:last-child) {
  margin-bottom: 0;
}

:deep(.planner-record-select-dropdown .ant-select-item-option-content) {
  padding: 10px 12px;
}

:deep(.planner-record-select-dropdown .ant-select-item-option-active:not(.ant-select-item-option-disabled)) {
  border-color: #dbeafe;
  background: #f8fbff;
}

:deep(.planner-record-select-dropdown .ant-select-item-option-selected:not(.ant-select-item-option-disabled)) {
  border-color: #bfd8ff;
  background: linear-gradient(180deg, #f7fbff 0%, #eef6ff 100%);
  box-shadow: inset 0 0 0 1px rgb(22 119 255 / 10%);
}

:deep(.planner-record-select-dropdown .planner-option) {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

:deep(.planner-record-select-dropdown .planner-option__title) {
  overflow: hidden;
  color: #1f2329;
  font-size: 14px;
  font-weight: 600;
  line-height: 1.45;
  text-overflow: ellipsis;
  white-space: nowrap;
}

:deep(.planner-record-select-dropdown .planner-option__desc) {
  overflow: hidden;
  color: #7d8898;
  font-size: 12px;
  line-height: 1.5;
  text-overflow: ellipsis;
  white-space: nowrap;
}

:deep(.one-to-one-schedule-modal .ant-select-focused .ant-select-selector),
:deep(.one-to-one-schedule-modal .ant-picker-focused) {
  border-color: #1677ff !important;
}

@media (max-width: 1200px) {
  .planner-layout {
    grid-template-columns: minmax(0, 1fr);
  }

  .planner-inline,
  .planner-inline__group {
    align-items: stretch;
  }

  .planner-inline {
    flex-direction: column;
  }

  .planner-inline__group--record,
  .planner-inline__group--status,
  .planner-inline__group--type {
    width: 100%;
  }

  .planner-inline__group--status {
    justify-content: flex-start;
  }

  .planner-balance {
    flex-wrap: wrap;
  }

  .planner-choice-row--type-inline {
    flex-wrap: wrap;
  }

  .planner-control--record {
    width: 100%;
  }
}

@media (max-width: 768px) {
  .planner-head,
  .planner-footer,
  .planner-review__head,
  .planner-review__footer {
    flex-direction: column;
    align-items: stretch;
  }

  .planner-head__stats {
    justify-content: space-between;
  }

  .planner-choice-row,
  .planner-form-grid {
    grid-template-columns: minmax(0, 1fr);
  }

  .planner-date-plan-row {
    grid-template-columns: minmax(0, 1fr);
  }

  .planner-date-plan-row__cell--start,
  .planner-date-plan-row__cell--start .planner-control {
    max-width: none;
  }

  :deep(.planner-date-plan-row__cell--start .ant-picker) {
    max-width: none !important;
  }

  .planner-time-range {
    grid-template-columns: minmax(0, 1fr);
  }

  .planner-time-range__sep {
    display: none;
  }

  .planner-inline__group {
    flex-direction: column;
    align-items: flex-start;
  }

  .planner-choice-row--type-inline {
    width: 100%;
    flex-direction: column;
  }
}
</style>

<style lang="less">
.planner-record-select-dropdown {
  padding: 8px;
}


.planner-record-select-dropdown .ant-select-item {
  padding: 0;
}

.planner-record-select-dropdown .ant-select-item-option {
  min-height: auto;
  margin-bottom: 6px;
  padding: 0;
  border: 1px solid transparent;
  border-radius: 12px;
  background: #fff;
  transition:
    border-color 0.18s ease,
    background-color 0.18s ease,
    box-shadow 0.18s ease;
}

.planner-record-select-dropdown .ant-select-item-option:last-child {
  margin-bottom: 0;
}

.planner-record-select-dropdown .ant-select-item-option-content {
  padding: 10px 12px;
}

.planner-record-select-dropdown .ant-select-item-option-active:not(.ant-select-item-option-disabled) {
  border-color: #dbeafe;
  background: #f8fbff;
}

.planner-record-select-dropdown .ant-select-item-option-selected:not(.ant-select-item-option-disabled) {
  border-color: #bfd8ff;
  background: linear-gradient(180deg, #f7fbff 0%, #eef6ff 100%);
  box-shadow: inset 0 0 0 1px rgb(22 119 255 / 10%);
}

.planner-record-select-dropdown .planner-option {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.planner-record-select-dropdown .planner-option__title {
  overflow: hidden;
  color: #1f2329;
  font-size: 14px;
  font-weight: 600;
  line-height: 1.45;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.planner-record-select-dropdown .planner-option__desc {
  overflow: hidden;
  color: #7d8898;
  font-size: 12px;
  line-height: 1.5;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
