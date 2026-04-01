<script setup lang="ts">
import {
  BookOutlined,
  CalendarOutlined,
  ClockCircleOutlined,
  CloseOutlined,
  EnvironmentOutlined,
  InfoCircleOutlined,
  ThunderboltOutlined,
  UserOutlined,
} from '@ant-design/icons-vue'
import type { OneToOneItem } from '../../../api/edu-center/one-to-one'
import dayjs, { type Dayjs } from 'dayjs'
import { computed, ref, watch } from 'vue'

type RecommendationKey = 'stable' | 'balanced' | 'efficient'
type PreviewTone = 'ready' | 'warning' | 'blocked' | 'deferred'

interface PreviewPlan {
  date: string
  week: string
  time: string
  teacher: string
  classroom: string
  status: string
  tone: PreviewTone
  reasonTags: string[]
  riskNote: string
  resolutionNote: string
  actions: string[]
}

interface InsightItem {
  label: string
  value: string
  desc: string
}

interface RecommendationScenario {
  estimatedCount: number
  cycleWeeks: number
  previewPlans: PreviewPlan[]
  suggestions: string[]
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

const durationOptions = ['45分钟 / 1课时', '60分钟 / 1课时', '90分钟 / 2课时']
const weekDayOptions = ['周一', '周二', '周三', '周四', '周五', '周六', '周日']
const recommendationPlans: Array<{
  key: RecommendationKey
  badge: string
  title: string
  desc: string
  metrics: string[]
}> = [
  {
    key: 'stable',
    badge: '推荐',
    title: '方案 A · 稳定优先',
    desc: '优先固定周一 / 周三 / 周五，尽量保持同老师与同教室。',
    metrics: ['老师稳定 92%', '教室稳定 86%'],
  },
  {
    key: 'balanced',
    badge: '均衡',
    title: '方案 B · 均衡优先',
    desc: '兼顾老师负载和固定时段，尽量减少需人工确认的场景。',
    metrics: ['冲突风险更低', '备选老师 2 位'],
  },
  {
    key: 'efficient',
    badge: '高效率',
    title: '方案 C · 排满优先',
    desc: '优先最早可用资源，更快排满本周期课次，适合补排。',
    metrics: ['最早 2 周排满', '老师切换更多'],
  },
]

const teacherModeCards = [
  {
    key: 'fixed',
    title: '固定老师优先',
    desc: '同一周期尽量由同一位老师连续授课',
  },
  {
    key: 'smart',
    title: '智能分配老师',
    desc: '优先空闲老师，同时保留时段稳定性',
  },
]

const roomModeCards = [
  {
    key: 'smart',
    title: '智能教室分配',
    desc: '优先同楼层空闲教室，自动预留备选',
  },
  {
    key: 'fixed',
    title: '固定教室优先',
    desc: '尽量固定在同一教室，减少学员切换',
  },
]

const smartRuleOptions = [
  { key: 'avoidConflict', label: '自动避让冲突' },
  { key: 'sameTeacher', label: '优先同老师' },
  { key: 'sameRoom', label: '优先同教室' },
  { key: 'continuous', label: '优先连续排课' },
]
const holidayRuleOptions = [
  { key: 'skipHoliday', label: '节假日自动跳过' },
  { key: 'deferHoliday', label: '自动顺延到下个可排日' },
  { key: 'keepLessonCount', label: '保持总课次数不变' },
]
const deferRuleOptions = [
  { key: 'teacherLeave', label: '老师请假自动换备选' },
  { key: 'roomBusy', label: '教室占用自动换房' },
  { key: 'studentLeave', label: '学员请假顺延补排' },
]

const recommendationScenarios: Record<RecommendationKey, RecommendationScenario> = {
  stable: {
    estimatedCount: 12,
    cycleWeeks: 4,
    previewPlans: [
      {
        date: '2026-04-06',
        week: '周一',
        time: '16:00 - 16:45',
        teacher: '李老师',
        classroom: '个训室 03',
        status: '待创建',
        tone: 'ready',
        reasonTags: ['同老师', '同教室', '无冲突'],
        riskNote: '',
        resolutionNote: '',
        actions: ['直接创建'],
      },
      {
        date: '2026-04-08',
        week: '周三',
        time: '16:00 - 16:45',
        teacher: '李老师',
        classroom: '个训室 03',
        status: '待创建',
        tone: 'ready',
        reasonTags: ['同老师', '连续时段'],
        riskNote: '',
        resolutionNote: '',
        actions: ['直接创建'],
      },
      {
        date: '2026-04-10',
        week: '周五',
        time: '16:00 - 16:45',
        teacher: '李老师',
        classroom: '个训室 03',
        status: '已顺延',
        tone: 'deferred',
        reasonTags: ['节假日顺延', '保持总课次'],
        riskNote: '原日期命中节假日，不建议创建当日课程。',
        resolutionNote: '系统已按顺延规则调整到 2026-04-11 16:00-16:45，老师与教室均已校验可用。',
        actions: ['查看顺延', '恢复原日期'],
      },
      {
        date: '2026-04-13',
        week: '周一',
        time: '16:00 - 16:45',
        teacher: '周老师',
        classroom: '言语教室 02',
        status: '需确认',
        tone: 'warning',
        reasonTags: ['老师请假备选', '原教室占用'],
        riskNote: '原方案冲突：李老师该时段请假，原教室同期已有课程占用。',
        resolutionNote: '备选方案已二次校验通过：周老师 16:00-16:45 空闲，言语教室 02 未占用；确认后将按备选方案创建，不会强行占用冲突资源。',
        actions: ['接受备选', '改老师'],
      },
      {
        date: '2026-04-15',
        week: '周三',
        time: '16:00 - 16:45',
        teacher: '李老师',
        classroom: '个训室 03',
        status: '待创建',
        tone: 'ready',
        reasonTags: ['同老师', '无冲突'],
        riskNote: '',
        resolutionNote: '',
        actions: ['直接创建'],
      },
      {
        date: '2026-04-17',
        week: '周五',
        time: '16:00 - 16:45',
        teacher: '李老师',
        classroom: '个训室 03',
        status: '待创建',
        tone: 'ready',
        reasonTags: ['同教室', '连续排课'],
        riskNote: '',
        resolutionNote: '',
        actions: ['直接创建'],
      },
    ],
    suggestions: [
      '建议按固定周一 / 周三 / 周五生成，方便家长形成稳定到课习惯。',
      '第 4 次课保留一个备选老师，可降低请假导致的人工调整成本。',
      '若后续接入逻辑，可在这里叠加余额不足、老师忙闲和教室占用明细。',
    ],
  },
  balanced: {
    estimatedCount: 12,
    cycleWeeks: 4,
    previewPlans: [
      {
        date: '2026-04-06',
        week: '周一',
        time: '16:00 - 16:45',
        teacher: '李老师',
        classroom: '个训室 03',
        status: '待创建',
        tone: 'ready',
        reasonTags: ['老师负载均衡', '无冲突'],
        riskNote: '',
        resolutionNote: '',
        actions: ['直接创建'],
      },
      {
        date: '2026-04-08',
        week: '周三',
        time: '16:00 - 16:45',
        teacher: '周老师',
        classroom: '言语教室 02',
        status: '待创建',
        tone: 'ready',
        reasonTags: ['老师负载均衡', '空闲教室'],
        riskNote: '',
        resolutionNote: '',
        actions: ['直接创建'],
      },
      {
        date: '2026-04-10',
        week: '周五',
        time: '16:00 - 16:45',
        teacher: '李老师',
        classroom: '个训室 03',
        status: '已顺延',
        tone: 'deferred',
        reasonTags: ['节假日顺延', '自动补齐'],
        riskNote: '原日期命中节假日，不建议创建当日课程。',
        resolutionNote: '系统已顺延到最近可排日，且保持总课次不变。',
        actions: ['查看顺延'],
      },
      {
        date: '2026-04-13',
        week: '周一',
        time: '16:00 - 16:45',
        teacher: '周老师',
        classroom: '言语教室 02',
        status: '待创建',
        tone: 'ready',
        reasonTags: ['备选已固化', '无冲突'],
        riskNote: '',
        resolutionNote: '',
        actions: ['直接创建'],
      },
      {
        date: '2026-04-15',
        week: '周三',
        time: '16:00 - 16:45',
        teacher: '黄老师',
        classroom: 'OT 教室 01',
        status: '需确认',
        tone: 'warning',
        reasonTags: ['老师负载切换', '教室更近'],
        riskNote: '原固定老师该时段连续授课过密，继续插入会导致负载过高。',
        resolutionNote: '系统推荐黄老师与 OT 教室 01，二次校验通过；确认后可降低后续连续课冲突风险。',
        actions: ['接受备选', '改教室'],
      },
      {
        date: '2026-04-17',
        week: '周五',
        time: '16:00 - 16:45',
        teacher: '李老师',
        classroom: '个训室 03',
        status: '待创建',
        tone: 'ready',
        reasonTags: ['无冲突', '周期闭环'],
        riskNote: '',
        resolutionNote: '',
        actions: ['直接创建'],
      },
    ],
    suggestions: [
      '如果更重视老师稳定性，可以切回方案 A。',
      '当前方案更适合老师资源紧张、但又不希望出现不可创建的场景。',
      '后续可把老师负载阈值做成可调参数，让均衡策略更贴近运营偏好。',
    ],
  },
  efficient: {
    estimatedCount: 14,
    cycleWeeks: 4,
    previewPlans: [
      {
        date: '2026-04-05',
        week: '周日',
        time: '09:00 - 09:45',
        teacher: '黄老师',
        classroom: 'OT 教室 01',
        status: '待创建',
        tone: 'ready',
        reasonTags: ['最早可排', '补排优先'],
        riskNote: '',
        resolutionNote: '',
        actions: ['直接创建'],
      },
      {
        date: '2026-04-07',
        week: '周二',
        time: '09:00 - 09:45',
        teacher: '李老师',
        classroom: '个训室 03',
        status: '待创建',
        tone: 'ready',
        reasonTags: ['最早空档', '自动补齐'],
        riskNote: '',
        resolutionNote: '',
        actions: ['直接创建'],
      },
      {
        date: '2026-04-09',
        week: '周四',
        time: '09:00 - 09:45',
        teacher: '周老师',
        classroom: '言语教室 02',
        status: '需确认',
        tone: 'warning',
        reasonTags: ['老师切换', '教室切换'],
        riskNote: '原偏好老师与教室在该时段均无可用资源。',
        resolutionNote: '备选方案已校验通过：切换周老师与言语教室 02 后可正常创建，但稳定性会下降。',
        actions: ['接受备选', '改时段'],
      },
      {
        date: '2026-04-12',
        week: '周日',
        time: '09:00 - 09:45',
        teacher: '-',
        classroom: '-',
        status: '不可创建',
        tone: 'blocked',
        reasonTags: ['学员冲突', '无空闲教室'],
        riskNote: '该时段学员已有其他课程，且同楼层教室全部占满。',
        resolutionNote: '当前没有可用备选方案，建议改时段或关闭“最早排满”策略后重新生成。',
        actions: ['改时段', '跳过本次'],
      },
      {
        date: '2026-04-14',
        week: '周二',
        time: '09:00 - 09:45',
        teacher: '黄老师',
        classroom: 'OT 教室 01',
        status: '待创建',
        tone: 'ready',
        reasonTags: ['最早空档', '补排优先'],
        riskNote: '',
        resolutionNote: '',
        actions: ['直接创建'],
      },
      {
        date: '2026-04-16',
        week: '周四',
        time: '09:00 - 09:45',
        teacher: '李老师',
        classroom: '个训室 03',
        status: '已顺延',
        tone: 'deferred',
        reasonTags: ['老师请假顺延', '自动补齐'],
        riskNote: '原计划老师临时请假，系统未找到同日可用备选。',
        resolutionNote: '已顺延到最近可排空档，并保持总课次数不变。',
        actions: ['查看顺延'],
      },
    ],
    suggestions: [
      '这套方案更适合补排和短期消课，不适合作为长期固定模板。',
      '出现 `不可创建` 时，优先改时段比强行换老师更稳妥。',
      '如果你希望减少人工确认与失败项，建议切回方案 B。',
    ],
  },
}

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
    defaultStudentClassTime: 1,
    teacherClassTime: 0,
    defaultTeacherClassTime: 0,
    defaultClassTimeRecordMode: 1,
    classTime: 45,
    createdTime: '2026-03-20 16:20:00',
    tuitionAccountCount: 1,
    remark: '家长希望固定周一 / 周三 / 周五放学后安排课程。',
    teacherList: [
      { teacherId: 'teacher-001', name: '李老师', status: 1, classId: '10001', isDefault: true },
      { teacherId: 'teacher-002', name: '周老师', status: 1, classId: '10001' },
    ],
  },
  {
    id: '10002',
    name: '李欣然-认知理解1对1',
    studentName: '李欣然',
    studentId: 'stu-002',
    lessonId: 'lesson-002',
    lessonName: '认知理解训练',
    classRoomId: 'room-03',
    classRoomName: '个训室 03',
    defaultTeacherId: 'teacher-001',
    defaultTeacherName: '李老师',
    classTeacherId: 'advisor-002',
    classTeacherName: '黄老师',
    status: 1,
    classStudentStatus: 1,
    isScheduled: true,
    defaultStudentClassTime: 1,
    teacherClassTime: 0,
    defaultTeacherClassTime: 0,
    defaultClassTimeRecordMode: 1,
    classTime: 45,
    createdTime: '2026-03-18 11:05:00',
    tuitionAccountCount: 2,
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
    classRoomId: 'room-03',
    classRoomName: '个训室 03',
    defaultTeacherId: 'teacher-001',
    defaultTeacherName: '李老师',
    classTeacherId: '',
    classTeacherName: '',
    status: 1,
    classStudentStatus: 2,
    isScheduled: false,
    defaultStudentClassTime: 1,
    teacherClassTime: 0,
    defaultTeacherClassTime: 0,
    defaultClassTimeRecordMode: 2,
    classTime: 45,
    createdTime: '2026-03-12 09:40:00',
    tuitionAccountCount: 1,
    remark: '当前处于停课中，恢复后再批量排课。',
    teacherList: [
      { teacherId: 'teacher-001', name: '李老师', status: 1, classId: '10003', isDefault: true },
      { teacherId: 'teacher-002', name: '周老师', status: 1, classId: '10003' },
    ],
  },
]

const formState = ref({
  duration: durationOptions[0],
  teacher: oneToOneRecords[0]?.defaultTeacherName || '',
  classroom: oneToOneRecords[0]?.classRoomName || '',
  teacherMode: teacherModeCards[0].key,
  roomMode: roomModeCards[0].key,
})

const aiPromptText = ref('')
const aiGenerating = ref(false)
const aiLoadingStep = ref(0)
const aiInsightText = ref('')
const aiSteps = [
  '正在分析学员历史排课偏好...',
  '正在调用约束求解器计算时间资源...',
  '正在匹配最优教师与教室...',
  '正在生成人类可读的排课报告...',
]

const scheduleRange = ref<[Dayjs, Dayjs]>([
  dayjs().add(5, 'day').startOf('day'),
  dayjs().add(32, 'day').startOf('day'),
])
const startTime = ref(dayjs().hour(16).minute(0).second(0))
const endTime = ref(dayjs().hour(16).minute(45).second(0))
const selectedOneToOneId = ref(oneToOneRecords[0]?.id || '')
const selectedRecommendation = ref<RecommendationKey>(recommendationPlans[0].key)
const selectedWeekdays = ref(['周一', '周三', '周五'])
const selectedSmartRules = ref(['avoidConflict', 'sameTeacher', 'sameRoom'])
const selectedHolidayRules = ref(['skipHoliday', 'deferHoliday'])
const selectedDeferRules = ref(['teacherLeave', 'roomBusy'])

const selectedOneToOne = computed(() =>
  oneToOneRecords.find(item => item.id === selectedOneToOneId.value) || oneToOneRecords[0],
)

const oneToOneSelectOptions = computed(() =>
  oneToOneRecords.map(item => ({
    label: `${item.name} · ${item.studentName || '-'} · ${item.lessonName || '-'}`,
    value: item.id,
  })),
)

const oneToOneRecordModeText = computed(() =>
  Number(selectedOneToOne.value?.defaultClassTimeRecordMode) === 2 ? '按上课时长记录' : '按固定课时记录',
)

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

const selectedOneToOneSummary = computed(() => [
  { label: '1对1名称', value: selectedOneToOne.value?.name || '-' },
  { label: '学员', value: selectedOneToOne.value?.studentName || '-' },
  { label: '关联课程', value: selectedOneToOne.value?.lessonName || '-' },
  { label: '默认上课教师', value: selectedOneToOne.value?.defaultTeacherName || '-' },
  { label: '班主任', value: selectedOneToOne.value?.classTeacherName || '-' },
  { label: '上课教室', value: selectedOneToOne.value?.classRoomName || '-' },
  { label: '课时记录方式', value: oneToOneRecordModeText.value },
  {
    label: '默认记录课时',
    value: `学员 ${selectedOneToOne.value?.defaultStudentClassTime ?? 0} / 教师 ${selectedOneToOne.value?.defaultTeacherClassTime ?? 0}`,
  },
])

const selectedOneToOneTeacherOptions = computed(() => {
  const options = new Set<string>()
  selectedOneToOne.value?.teacherList?.forEach((item) => {
    if (item?.name)
      options.add(item.name)
  })
  if (selectedOneToOne.value?.defaultTeacherName)
    options.add(selectedOneToOne.value.defaultTeacherName)
  return [...options]
})

const selectedOneToOneClassroomOptions = computed(() =>
  selectedOneToOne.value?.classRoomName ? [selectedOneToOne.value.classRoomName] : [],
)

watch(
  selectedOneToOne,
  (value) => {
    if (!value)
      return
    if (value.defaultTeacherName)
      formState.value.teacher = value.defaultTeacherName
    if (value.classRoomName)
      formState.value.classroom = value.classRoomName
    formState.value.duration = Number(value.defaultClassTimeRecordMode) === 2 ? '按上课时长记录' : '45分钟 / 1课时'
  },
  { immediate: true },
)

const activeRecommendationScenario = computed(() => recommendationScenarios[selectedRecommendation.value])
const previewPlans = computed(() => activeRecommendationScenario.value.previewPlans)
const estimatedCount = computed(() => activeRecommendationScenario.value.estimatedCount)
const cycleWeeks = computed(() => activeRecommendationScenario.value.cycleWeeks)
const readyCount = computed(() => previewPlans.value.filter(item => item.tone === 'ready').length)
const deferredCount = computed(() => previewPlans.value.filter(item => item.tone === 'deferred').length)
const pendingCount = computed(() => previewPlans.value.filter(item => item.tone === 'warning').length)
const blockedCount = computed(() => previewPlans.value.filter(item => item.tone === 'blocked').length)
const rangeText = computed(() => `${scheduleRange.value[0].format('MM/DD')} - ${scheduleRange.value[1].format('MM/DD')}`)
const selectedWeekdaysText = computed(() => selectedWeekdays.value.join(' / '))
const timeRangeText = computed(() => `${startTime.value.format('HH:mm')} - ${endTime.value.format('HH:mm')}`)
const selectedRecommendationLabel = computed(() =>
  recommendationPlans.find(item => item.key === selectedRecommendation.value)?.title || '',
)
const schedulablePlans = computed(() =>
  previewPlans.value.filter(item => item.teacher !== '-' && item.classroom !== '-'),
)

function buildDistribution(values: string[]) {
  const counter = new Map<string, number>()
  values.filter(Boolean).forEach((value) => {
    counter.set(value, (counter.get(value) || 0) + 1)
  })
  return [...counter.entries()].sort((a, b) => b[1] - a[1])
}

const teacherDistribution = computed(() => buildDistribution(schedulablePlans.value.map(item => item.teacher)))
const classroomDistribution = computed(() => buildDistribution(schedulablePlans.value.map(item => item.classroom)))

const smartInsightList = computed<InsightItem[]>(() => {
  const totalSchedulable = Math.max(schedulablePlans.value.length, 1)
  const [primaryTeacher = ['未分配', 0], backupTeacher = ['无', 0]] = teacherDistribution.value
  const [primaryRoom = ['未分配', 0], backupRoom = ['无', 0]] = classroomDistribution.value
  const manualCount = pendingCount.value + blockedCount.value
  const manualDesc = [
    pendingCount.value > 0 ? `${pendingCount.value} 节需确认` : '',
    blockedCount.value > 0 ? `${blockedCount.value} 节不可创建` : '',
    deferredCount.value > 0 ? `${deferredCount.value} 节已顺延` : '',
  ].filter(Boolean).join('，') || '当前无需人工介入'

  return [
    {
      label: '主排老师',
      value: `${primaryTeacher[0]} · ${primaryTeacher[1]}节`,
      desc: backupTeacher[1] > 0
        ? `备选老师 ${backupTeacher[0]} · ${backupTeacher[1]}节，主老师命中 ${Math.round((Number(primaryTeacher[1]) / totalSchedulable) * 100)}%`
        : '全周期命中同一位老师',
    },
    {
      label: '主排教室',
      value: `${primaryRoom[0]} · ${primaryRoom[1]}节`,
      desc: backupRoom[1] > 0
        ? `备选教室 ${backupRoom[0]} · ${backupRoom[1]}节，主教室命中 ${Math.round((Number(primaryRoom[1]) / totalSchedulable) * 100)}%`
        : '全周期命中同一间教室',
    },
    {
      label: '人工干预',
      value: `${manualCount}节`,
      desc: manualDesc,
    },
  ]
})

const smartSuggestionList = computed(() => {
  const notes = [
    `当前按 ${selectedWeekdaysText.value} ${timeRangeText.value} 生成，预计覆盖 ${cycleWeeks.value} 周，共 ${estimatedCount.value} 节。`,
  ]
  const firstDeferred = previewPlans.value.find(item => item.tone === 'deferred')
  if (firstDeferred) {
    notes.push(`${firstDeferred.date} 已按规则顺延处理，系统保留总课次数不变并重新校验老师、教室可用性。`)
  }
  const firstWarning = previewPlans.value.find(item => item.tone === 'warning')
  if (firstWarning) {
    notes.push(`${firstWarning.date} 存在备选方案，当前等待确认是否接受 ${firstWarning.teacher} · ${firstWarning.classroom} 的替换结果。`)
  }
  const firstBlocked = previewPlans.value.find(item => item.tone === 'blocked')
  if (firstBlocked) {
    notes.push(`${firstBlocked.date} 当前无可用备选资源，建议优先改时段或切换到更稳妥的推荐方案。`)
  }
  else {
    notes.push('当前预览中没有发现“无备选仍强排”的风险项，系统不会直接创建冲突课程。')
  }
  return notes
})

function closeModal() {
  modalOpen.value = false
}

function toggleWeekday(day: string) {
  const active = selectedWeekdays.value.includes(day)
  if (active && selectedWeekdays.value.length === 1)
    return

  selectedWeekdays.value = active
    ? selectedWeekdays.value.filter(item => item !== day)
    : [...selectedWeekdays.value, day]
}

function toggleSmartRule(ruleKey: string) {
  const active = selectedSmartRules.value.includes(ruleKey)
  selectedSmartRules.value = active
    ? selectedSmartRules.value.filter(item => item !== ruleKey)
    : [...selectedSmartRules.value, ruleKey]
}

function toggleHolidayRule(key: string) {
  const active = selectedHolidayRules.value.includes(key)
  selectedHolidayRules.value = active
    ? selectedHolidayRules.value.filter(item => item !== key)
    : [...selectedHolidayRules.value, key]
}

function toggleDeferRule(key: string) {
  const active = selectedDeferRules.value.includes(key)
  selectedDeferRules.value = active
    ? selectedDeferRules.value.filter(item => item !== key)
    : [...selectedDeferRules.value, key]
}

function handleAIParseRemark() {
  if (!selectedOneToOne.value?.remark) return

  aiGenerating.value = true
  aiLoadingStep.value = 0

  const stepInterval = setInterval(() => {
    aiLoadingStep.value = (aiLoadingStep.value + 1) % aiSteps.length
  }, 600)

  setTimeout(() => {
    clearInterval(stepInterval)
    aiGenerating.value = false

    // Simulate AI parsing result
    selectedWeekdays.value = ['周一', '周三', '周五']
    selectedRecommendation.value = 'stable'
    formState.value.teacherMode = 'fixed'

    aiInsightText.value = '已根据备注为您自动勾选“周一、周三、周五”，并启用了“老师稳定优先”策略。节假日将按顺延处理，确保家长体验最优。'
  }, 2500)
}

function handleAIGenerate() {
  if (!aiPromptText.value.trim()) return

  aiGenerating.value = true
  aiLoadingStep.value = 0

  const stepInterval = setInterval(() => {
    aiLoadingStep.value = (aiLoadingStep.value + 1) % aiSteps.length
  }, 800)

  setTimeout(() => {
    clearInterval(stepInterval)
    aiGenerating.value = false

    // Simulate AI generation result based on prompt
    selectedRecommendation.value = 'efficient'
    formState.value.teacherMode = 'smart'

    aiInsightText.value = `已为您应用自然语言排课：“${aiPromptText.value}”。系统切换为【方案C：排满优先】，优先填补早盘空档，预计比原计划提早一周消课完毕。`
    aiPromptText.value = ''
  }, 3200)
}
</script>

<template>
  <a-modal
    v-model:open="modalOpen"
    centered
    class="one-to-one-schedule-modal"
    :footer="null"
    :width="1080"
    :keyboard="false"
    :closable="false"
    :mask-closable="true"
    @cancel="closeModal"
  >
    <template #title>
      <div class="batch-modal-header">
        <div class="batch-modal-header__main">
          <span class="batch-modal-header__badge">智能批量</span>
          <div>
            <div class="batch-modal-header__title">
              1 对 1 批量排课
            </div>
            <div class="batch-modal-header__subtitle">
              先设定周期、时段和策略，再批量生成整段日程。这一版先实现静态样式，不接排课逻辑。
            </div>
          </div>
        </div>

        <a-button type="text" class="batch-modal-header__close" @click="closeModal">
          <template #icon>
            <CloseOutlined />
          </template>
        </a-button>
      </div>
    </template>

    <div class="batch-modal-body">
      <!-- AI Loading Overlay -->
      <div v-if="aiGenerating" class="ai-loading-overlay">
        <div class="ai-loading-content">
          <div class="ai-loading-spinner"></div>
          <div class="ai-loading-text">
            <span class="ai-icon-sparkle">✨</span>
            {{ aiSteps[aiLoadingStep] }}
          </div>
        </div>
      </div>

      <section class="batch-overview">
        <div class="batch-overview__top">
          <div class="batch-overview__content">
            <span class="batch-overview__eyebrow">批量排课工作台</span>
            <h3>一次配置批量规则，直接预览未来 4 周的排课结果</h3>
            <p>把“对象选择、周期模板、智能策略、生成预览”放在同一个弹窗里，先看清批量结果，再决定是否创建。</p>
          </div>

          <div class="batch-overview__hero-card">
            <span class="batch-overview__hero-label">预计生成</span>
            <div class="batch-overview__hero-headline">
              <strong>{{ estimatedCount }}</strong>
              <span>个候选日程</span>
            </div>
            <div class="batch-overview__hero-meta">
              {{ cycleWeeks }} 周周期 · {{ rangeText }}
            </div>
          </div>
        </div>

        <div class="batch-overview__metrics">
          <div class="batch-overview-metric">
            <span class="batch-overview-metric__label">固定上课日</span>
            <div class="batch-overview-metric__chips">
              <span v-for="item in selectedWeekdays" :key="item" class="batch-overview-chip">
                {{ item }}
              </span>
            </div>
          </div>

          <div class="batch-overview-metric">
            <span class="batch-overview-metric__label">固定时段</span>
            <div class="batch-overview-metric__value">
              {{ timeRangeText }}
            </div>
            <div class="batch-overview-metric__subvalue">
              {{ formState.duration }}
            </div>
          </div>

          <div class="batch-overview-metric">
            <span class="batch-overview-metric__label">预览结果</span>
            <div class="batch-overview-metric__result">
              <span class="batch-overview-result batch-overview-result--ready">
                {{ readyCount }} 个可直接创建
              </span>
              <span class="batch-overview-result batch-overview-result--info">
                {{ deferredCount }} 个已自动顺延
              </span>
              <span class="batch-overview-result batch-overview-result--warning">
                {{ pendingCount }} 个需确认
              </span>
              <span class="batch-overview-result batch-overview-result--danger">
                {{ blockedCount }} 个不可创建
              </span>
            </div>
          </div>
        </div>
      </section>

      <div class="batch-layout">
        <section class="batch-main-panel">
          <section class="batch-section">
            <div class="batch-section__header">
              <div class="batch-section__title">
                推荐方案
              </div>
              <div class="batch-section__desc">
                系统根据固定时段、老师稳定性和冲突规避倾向，给出三种默认排课方案
              </div>
            </div>

            <div class="batch-recommend-grid">
              <button
                v-for="item in recommendationPlans"
                :key="item.key"
                type="button"
                class="batch-recommend-card"
                :class="{ 'batch-recommend-card--active': selectedRecommendation === item.key }"
                @click="selectedRecommendation = item.key"
              >
                <div class="batch-recommend-card__top">
                  <span class="batch-recommend-card__badge">{{ item.badge }}</span>
                  <span class="batch-recommend-card__title">{{ item.title }}</span>
                </div>
                <div class="batch-recommend-card__desc">
                  {{ item.desc }}
                </div>
                <div class="batch-recommend-card__metrics">
                  <span v-for="metric in item.metrics" :key="metric" class="batch-recommend-card__metric">
                    {{ metric }}
                  </span>
                </div>
              </button>
            </div>
          </section>

          <section class="batch-section">
            <div class="batch-section__header">
              <div class="batch-section__title">
                排课对象
              </div>
              <div class="batch-section__desc">
                选择已有的 1 对 1 记录，系统自动带出当前班级的学员、课程、教师与记录课时信息
              </div>
            </div>

            <div class="batch-field-grid">
              <label class="batch-field batch-field--full">
                <span class="batch-field__label">
                  <BookOutlined />
                  1对1记录
                </span>
                <a-select v-model:value="selectedOneToOneId" size="large" show-search option-filter-prop="label">
                  <a-select-option
                    v-for="item in oneToOneSelectOptions"
                    :key="item.value"
                    :value="item.value"
                    :label="item.label"
                  >
                    {{ item.label }}
                  </a-select-option>
                </a-select>
              </label>

              <div class="batch-one-to-one-card batch-field--full">
                <div class="batch-one-to-one-card__top">
                  <div>
                    <div class="batch-one-to-one-card__title">
                      {{ selectedOneToOne?.name || '-' }}
                    </div>
                    <div class="batch-one-to-one-card__meta">
                      {{ selectedOneToOne?.studentName || '-' }} · {{ selectedOneToOne?.lessonName || '-' }}
                    </div>
                  </div>

                  <div class="batch-one-to-one-card__chips">
                    <span class="batch-one-to-one-card__chip">
                      {{ oneToOneStatusText }}
                    </span>
                    <span class="batch-one-to-one-card__chip batch-one-to-one-card__chip--sub">
                      {{ oneToOneStudentStatusText }}
                    </span>
                    <span class="batch-one-to-one-card__chip batch-one-to-one-card__chip--sub">
                      {{ selectedOneToOne?.isScheduled ? '已排课' : '未排课' }}
                    </span>
                  </div>
                </div>

                <div class="batch-one-to-one-card__grid">
                  <div v-for="item in selectedOneToOneSummary" :key="item.label" class="batch-one-to-one-card__item">
                    <div class="batch-one-to-one-card__item-label">
                      {{ item.label }}
                    </div>
                    <div class="batch-one-to-one-card__item-value">
                      {{ item.value }}
                    </div>
                  </div>
                </div>

                <div class="batch-one-to-one-card__remark">
                  <span class="batch-one-to-one-card__remark-label">备注</span>
                  <span>{{ selectedOneToOne?.remark || '-' }}</span>
                </div>
              </div>
            </div>
          </section>

          <section class="batch-section">
            <div class="batch-section__header">
              <div class="batch-section__title">
                周期模板
              </div>
              <div class="batch-section__desc">
                定义排课区间、重复周次和固定时段
              </div>
            </div>

            <div class="batch-field-grid">
              <label class="batch-field batch-field--full">
                <span class="batch-field__label">
                  <CalendarOutlined />
                  排课周期
                </span>
                <a-range-picker v-model:value="scheduleRange" size="large" class="batch-field__control" />
              </label>

              <div class="batch-field batch-field--full">
                <span class="batch-field__label">
                  <CalendarOutlined />
                  每周重复
                </span>
                <div class="batch-weekday-group">
                  <button
                    v-for="item in weekDayOptions"
                    :key="item"
                    type="button"
                    class="batch-weekday-chip"
                    :class="{ 'batch-weekday-chip--active': selectedWeekdays.includes(item) }"
                    @click="toggleWeekday(item)"
                  >
                    {{ item }}
                  </button>
                </div>
              </div>

              <div class="batch-field">
                <span class="batch-field__label">
                  <ClockCircleOutlined />
                  固定时段
                </span>
                <div class="batch-time-range">
                  <a-time-picker v-model:value="startTime" size="large" format="HH:mm" class="batch-field__control" />
                  <span class="batch-time-range__split">至</span>
                  <a-time-picker v-model:value="endTime" size="large" format="HH:mm" class="batch-field__control" />
                </div>
              </div>

              <label class="batch-field">
                <span class="batch-field__label">
                  <ClockCircleOutlined />
                  单次课时
                </span>
                <a-select v-model:value="formState.duration" size="large">
                  <a-select-option v-for="item in durationOptions" :key="item" :value="item">
                    {{ item }}
                  </a-select-option>
                </a-select>
              </label>

              <div class="batch-field batch-field--full">
                <span class="batch-field__label">
                  <CalendarOutlined />
                  节假日处理
                </span>
                <div class="batch-rule-group">
                  <button
                    v-for="item in holidayRuleOptions"
                    :key="item.key"
                    type="button"
                    class="batch-rule-chip"
                    :class="{ 'batch-rule-chip--active': selectedHolidayRules.includes(item.key) }"
                    @click="toggleHolidayRule(item.key)"
                  >
                    {{ item.label }}
                  </button>
                </div>
              </div>

              <div class="batch-field batch-field--full">
                <span class="batch-field__label">
                  <ThunderboltOutlined />
                  异常顺延
                </span>
                <div class="batch-rule-group">
                  <button
                    v-for="item in deferRuleOptions"
                    :key="item.key"
                    type="button"
                    class="batch-rule-chip"
                    :class="{ 'batch-rule-chip--active': selectedDeferRules.includes(item.key) }"
                    @click="toggleDeferRule(item.key)"
                  >
                    {{ item.label }}
                  </button>
                </div>
              </div>
            </div>
          </section>

          <section class="batch-section">
            <div class="batch-section__header">
              <div class="batch-section__title">
                智能策略
              </div>
              <div class="batch-section__desc">
                决定老师、教室与冲突处理的默认偏好
              </div>
            </div>

            <div class="batch-strategy-grid">
              <button
                v-for="item in teacherModeCards"
                :key="item.key"
                type="button"
                class="batch-strategy-card"
                :class="{ 'batch-strategy-card--active': formState.teacherMode === item.key }"
                @click="formState.teacherMode = item.key"
              >
                <span class="batch-strategy-card__title">{{ item.title }}</span>
                <span class="batch-strategy-card__desc">{{ item.desc }}</span>
              </button>
            </div>

            <div class="batch-strategy-grid batch-strategy-grid--secondary">
              <button
                v-for="item in roomModeCards"
                :key="item.key"
                type="button"
                class="batch-strategy-card"
                :class="{ 'batch-strategy-card--active': formState.roomMode === item.key }"
                @click="formState.roomMode = item.key"
              >
                <span class="batch-strategy-card__title">{{ item.title }}</span>
                <span class="batch-strategy-card__desc">{{ item.desc }}</span>
              </button>
            </div>

            <div class="batch-field-grid batch-field-grid--compact">
              <label class="batch-field">
                <span class="batch-field__label">
                  <UserOutlined />
                  优先老师
                </span>
                <a-select v-model:value="formState.teacher" size="large">
                  <a-select-option v-for="item in selectedOneToOneTeacherOptions" :key="item" :value="item">
                    {{ item }}
                  </a-select-option>
                </a-select>
              </label>

              <label class="batch-field">
                <span class="batch-field__label">
                  <EnvironmentOutlined />
                  优先教室
                </span>
                <a-select v-model:value="formState.classroom" size="large">
                  <a-select-option v-for="item in selectedOneToOneClassroomOptions" :key="item" :value="item">
                    {{ item }}
                  </a-select-option>
                </a-select>
              </label>

              <div class="batch-field batch-field--full">
                <span class="batch-field__label">
                  <ThunderboltOutlined />
                  智能规则
                </span>
                <div class="batch-rule-group">
                  <button
                    v-for="item in smartRuleOptions"
                    :key="item.key"
                    type="button"
                    class="batch-rule-chip"
                    :class="{ 'batch-rule-chip--active': selectedSmartRules.includes(item.key) }"
                    @click="toggleSmartRule(item.key)"
                  >
                    {{ item.label }}
                  </button>
                </div>
              </div>

              <label class="batch-field batch-field--full">
                <span class="batch-field__label">
                  <InfoCircleOutlined />
                  1对1备注
                </span>
                <div class="batch-existing-remark" :class="{ 'batch-existing-remark--has-content': selectedOneToOne?.remark }">
                  <div class="batch-existing-remark__content">
                    {{ selectedOneToOne?.remark || '暂无备注' }}
                  </div>
                  <a-button
                    v-if="selectedOneToOne?.remark"
                    type="primary"
                    size="small"
                    class="ai-parse-btn"
                    :loading="aiGenerating"
                    @click="handleAIParseRemark"
                  >
                    <template #icon>
                      <span class="ai-icon-sparkle">✨</span>
                    </template>
                    AI 智能提取偏好
                  </a-button>
                </div>
              </label>
            </div>
          </section>
        </section>

        <aside class="batch-side-panel">
          <section class="batch-side-card batch-side-card--summary">
            <div class="batch-side-card__header">
              <div class="batch-side-card__title">
                生成摘要
              </div>
            </div>

            <div class="batch-summary-board">
              <div class="batch-summary-board__headline">
                未来 {{ cycleWeeks }} 周 · {{ rangeText }}
              </div>
              <div class="batch-summary-board__count">
                {{ estimatedCount }} 个候选日程
              </div>
              <div class="batch-summary-board__meta">
                {{ selectedOneToOne?.name || '-' }} · {{ selectedOneToOne?.lessonName || '-' }} · {{ selectedRecommendationLabel }}
              </div>
            </div>

            <div class="batch-summary-tags">
              <span class="batch-summary-tag">
                {{ selectedWeekdaysText }}
              </span>
              <span class="batch-summary-tag">
                {{ timeRangeText }}
              </span>
              <span class="batch-summary-tag">
                {{ formState.duration }}
              </span>
              <span class="batch-summary-tag">
                {{ selectedRecommendationLabel }}
              </span>
            </div>
          </section>

          <section class="batch-side-card batch-side-card--preview-list">
            <div class="batch-side-card__header">
              <div>
                <div class="batch-side-card__title">
                  批量预览
                </div>
                <div class="batch-side-card__helper">
                  待创建=已校验；需确认=备选已校验；不可创建=需人工处理
                </div>
              </div>
              <div class="batch-side-card__extra">
                共 {{ previewPlans.length }} 条
              </div>
            </div>

            <div class="batch-preview-list">
              <div
                v-for="item in previewPlans"
                :key="`${selectedRecommendation}-${item.date}-${item.week}`"
                class="batch-preview-row"
                :class="`batch-preview-row--${item.tone}`"
              >
                <div class="batch-preview-row__top">
                  <div class="batch-preview-row__main">
                    <span class="batch-preview-row__date">{{ item.date }}</span>
                    <a-tooltip :title="`${item.week} · ${item.time} · ${item.teacher} · ${item.classroom}`">
                      <div class="batch-preview-row__meta">
                        {{ item.week }} · {{ item.time }} · {{ item.teacher }} · {{ item.classroom }}
                      </div>
                    </a-tooltip>
                  </div>

                  <div class="batch-preview-row__side">
                    <span
                      class="batch-preview-row__status"
                      :class="{
                        'batch-preview-row__status--ready': item.tone === 'ready',
                        'batch-preview-row__status--warning': item.tone === 'warning',
                        'batch-preview-row__status--blocked': item.tone === 'blocked',
                        'batch-preview-row__status--deferred': item.tone === 'deferred',
                      }"
                    >
                      {{ item.status }}
                    </span>

                    <div class="batch-preview-row__actions">
                      <a-button
                        v-for="action in item.actions"
                        :key="action"
                        type="link"
                        size="small"
                        class="batch-preview-row__action"
                      >
                        {{ action }}
                      </a-button>
                    </div>
                  </div>
                </div>

                <div class="batch-preview-row__bottom">
                  <div class="batch-preview-row__reasons">
                    <span v-for="reason in item.reasonTags" :key="reason" class="batch-preview-row__reason-chip">
                      {{ reason }}
                    </span>
                  </div>

                  <div v-if="item.riskNote || item.resolutionNote" class="batch-preview-row__notes">
                    <div v-if="item.riskNote" class="batch-preview-row__note batch-preview-row__note--risk">
                      {{ item.riskNote }}
                    </div>
                    <div v-if="item.resolutionNote" class="batch-preview-row__note batch-preview-row__note--resolved">
                      {{ item.resolutionNote }}
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </section>

          <section class="batch-side-card">
            <div class="batch-side-card__header">
              <div class="batch-side-card__title">
                智能判断
              </div>
            </div>

            <transition name="ai-fade">
              <div v-if="aiInsightText" class="ai-insight-box">
                <span class="ai-insight-box__icon">✨</span>
                <div class="ai-insight-box__content">
                  <div class="ai-insight-box__title">AI 排课洞察</div>
                  <div class="ai-insight-box__text">{{ aiInsightText }}</div>
                </div>
                <a-button type="text" size="small" class="ai-insight-box__close" @click="aiInsightText = ''">
                  <CloseOutlined />
                </a-button>
              </div>
            </transition>

            <div class="batch-insight-list">
              <div v-for="item in smartInsightList" :key="item.label" class="batch-insight-item">
                <div>
                  <div class="batch-insight-item__label">
                    {{ item.label }}
                  </div>
                  <div class="batch-insight-item__desc">
                    {{ item.desc }}
                  </div>
                </div>
                <div class="batch-insight-item__value">
                  {{ item.value }}
                </div>
              </div>
            </div>

            <div class="batch-suggestion-list">
              <div v-for="item in smartSuggestionList" :key="item" class="batch-suggestion-item">
                {{ item }}
              </div>
            </div>
          </section>
        </aside>
      </div>

      <div class="batch-modal-footer">
        <div class="batch-modal-footer__ai-prompt">
          <a-input-group compact class="ai-prompt-group">
            <a-input
              v-model:value="aiPromptText"
              placeholder="✨ 例如：将下周二的课都尽量挪到上午..."
              :disabled="aiGenerating"
              @pressEnter="handleAIGenerate"
            >
              <template #prefix>
                <span class="ai-icon-sparkle">✨</span>
              </template>
            </a-input>
            <a-button type="primary" :loading="aiGenerating" @click="handleAIGenerate">
              AI 调整
            </a-button>
          </a-input-group>
        </div>

        <div class="batch-modal-footer__actions">
          <a-button size="large" @click="closeModal">
            取消
          </a-button>
          <a-button size="large">
            预览排课
          </a-button>
          <a-button size="large" type="primary">
            批量创建 {{ estimatedCount }} 个日程
          </a-button>
        </div>
      </div>
    </div>
  </a-modal>
</template>

<style scoped lang="less">
.batch-modal-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}

.batch-modal-header__main {
  display: flex;
  align-items: flex-start;
  gap: 14px;
}

.batch-modal-header__badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 70px;
  height: 30px;
  padding: 0 12px;
  border-radius: 999px;
  background: linear-gradient(135deg, #1f6dff 0%, #3a87ff 100%);
  color: #fff;
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.04em;
}

.batch-modal-header__title {
  color: #1f2329;
  font-size: 22px;
  font-weight: 700;
  line-height: 1.25;
}

.batch-modal-header__subtitle {
  margin-top: 4px;
  color: #7c8698;
  font-size: 13px;
  line-height: 1.6;
}

.batch-modal-header__close {
  color: #9aa4b2;
}

.batch-modal-body {
  display: flex;
  flex-direction: column;
  gap: 18px;
  position: relative;
}

/* AI Elements Styles */
.ai-icon-sparkle {
  color: #1f6dff;
  font-size: 14px;
  animation: pulse 2s infinite;
}

@keyframes pulse {
  0% { opacity: 0.6; transform: scale(0.9); }
  50% { opacity: 1; transform: scale(1.1); }
  100% { opacity: 0.6; transform: scale(0.9); }
}

.ai-loading-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(255, 255, 255, 0.85);
  backdrop-filter: blur(4px);
  z-index: 100;
  border-radius: 22px;
}

.ai-loading-content {
  position: sticky;
  top: 50%;
  transform: translateY(-50%);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}

.ai-loading-spinner {
  width: 40px;
  height: 40px;
  border: 3px solid #eef5ff;
  border-top-color: #1f6dff;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin-bottom: 16px;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.ai-loading-text {
  font-size: 15px;
  font-weight: 600;
  color: #1f2329;
  display: flex;
  align-items: center;
  gap: 8px;
  background: linear-gradient(90deg, #1f6dff, #8b5cf6, #1f6dff);
  background-size: 200% auto;
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  animation: shimmer 3s linear infinite;
}

@keyframes shimmer {
  to { background-position: 200% center; }
}

.batch-existing-remark {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 12px 14px;
  border-radius: 12px;
  background: #f8fafc;
  color: #556070;
  font-size: 13px;
  line-height: 1.5;
  transition: all 0.3s ease;
}

.batch-existing-remark--has-content {
  background: #f4f7fb;
  border: 1px dashed #dbe3ee;
}

.batch-existing-remark__content {
  flex: 1;
}

.ai-parse-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  background: linear-gradient(135deg, #1f6dff 0%, #0056f7 100%);
  border: none;
  font-weight: 600;
  border-radius: 8px;
}

.ai-parse-btn:hover {
  background: linear-gradient(135deg, #3a87ff 0%, #1f6dff 100%);
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(31, 109, 255, 0.2);
}

.ai-insight-box {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  margin-bottom: 16px;
  padding: 14px 16px;
  background: linear-gradient(135deg, #f0f7ff 0%, #eef3fc 100%);
  border: 1px solid #d3e2ff;
  border-radius: 14px;
  box-shadow: 0 4px 16px rgba(31, 109, 255, 0.06);
  position: relative;
}

.ai-insight-box__icon {
  font-size: 18px;
  margin-top: 2px;
}

.ai-insight-box__content {
  flex: 1;
}

.ai-insight-box__title {
  font-size: 13px;
  font-weight: 700;
  color: #1f6dff;
  margin-bottom: 4px;
}

.ai-insight-box__text {
  font-size: 12px;
  color: #3b4b68;
  line-height: 1.6;
}

.ai-insight-box__close {
  position: absolute;
  top: 8px;
  right: 8px;
  color: #8b95a7;
}

.ai-fade-enter-active,
.ai-fade-leave-active {
  transition: all 0.4s ease;
}

.ai-fade-enter-from,
.ai-fade-leave-to {
  opacity: 0;
  transform: translateY(-10px);
}

.batch-modal-footer__ai-prompt {
  flex: 1;
  max-width: 480px;
}

.ai-prompt-group {
  display: flex !important;
}

:deep(.ai-prompt-group .ant-input) {
  border-top-right-radius: 0 !important;
  border-bottom-right-radius: 0 !important;
  border-color: #dbe3ee;
}

:deep(.ai-prompt-group .ant-input:focus) {
  border-color: #1f6dff;
  box-shadow: 0 0 0 2px rgba(31, 109, 255, 0.1);
}

:deep(.ai-prompt-group .ant-btn) {
  border-top-left-radius: 0 !important;
  border-bottom-left-radius: 0 !important;
  height: 44px;
  font-weight: 600;
  background: linear-gradient(135deg, #1f6dff 0%, #3a87ff 100%);
  border: none;
}

.batch-overview {
  display: flex;
  flex-direction: column;
  gap: 14px;
  padding: 18px 20px;
  border: 1px solid #e7edf5;
  border-radius: 22px;
  background:
    radial-gradient(circle at left top, rgb(31 109 255 / 10%) 0, rgb(31 109 255 / 0%) 42%),
    linear-gradient(135deg, #fbfdff 0%, #f8fbff 55%, #fff9f9 100%);
}

.batch-overview__top {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 360px;
  gap: 16px;
  align-items: center;
}

.batch-overview__eyebrow {
  display: inline-block;
  margin-bottom: 8px;
  color: #1f6dff;
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.08em;
}

.batch-overview__content h3 {
  margin: 0;
  color: #1f2329;
  font-size: 20px;
  line-height: 1.35;
}

.batch-overview__content p {
  margin: 8px 0 0;
  color: #667085;
  font-size: 13px;
  line-height: 1.65;
}

.batch-overview__hero-card {
  padding: 16px 18px;
  border: 1px solid rgb(232 238 247 / 96%);
  border-radius: 20px;
  background:
    radial-gradient(circle at top left, rgb(31 109 255 / 12%) 0, rgb(31 109 255 / 0%) 55%),
    linear-gradient(135deg, #f7fbff 0%, #eef5ff 100%);
  box-shadow:
    inset 0 1px 0 rgb(255 255 255 / 78%),
    0 12px 28px rgb(31 51 84 / 6%);
}

.batch-overview__hero-label {
  display: inline-block;
  color: #7f8ba0;
  font-size: 12px;
  font-weight: 600;
}

.batch-overview__hero-headline {
  display: flex;
  align-items: baseline;
  gap: 8px;
  margin-top: 8px;
  color: #1f2329;
}

.batch-overview__hero-headline strong {
  font-size: 30px;
  line-height: 1;
  font-weight: 700;
}

.batch-overview__hero-headline span {
  font-size: 15px;
  font-weight: 600;
}

.batch-overview__hero-meta {
  margin-top: 8px;
  color: #7f8ba0;
  font-size: 12px;
  line-height: 1.5;
}

.batch-overview__metrics {
  display: grid;
  grid-template-columns: 1.2fr 0.9fr 1fr;
  gap: 12px;
}

.batch-overview-metric {
  min-height: 108px;
  padding: 14px 16px;
  border: 1px solid #ebf0f6;
  border-radius: 16px;
  background: rgb(255 255 255 / 88%);
}

.batch-overview-metric__label {
  display: block;
  margin-bottom: 10px;
  color: #8b95a7;
  font-size: 12px;
  font-weight: 600;
}

.batch-overview-metric__chips,
.batch-overview-metric__result {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.batch-overview-chip {
  display: inline-flex;
  align-items: center;
  height: 30px;
  padding: 0 12px;
  border-radius: 999px;
  background: #eef5ff;
  color: #1f6dff;
  font-size: 12px;
  font-weight: 700;
}

.batch-overview-metric__value {
  color: #1f2329;
  font-size: 20px;
  font-weight: 700;
  line-height: 1.2;
}

.batch-overview-metric__subvalue {
  margin-top: 6px;
  color: #8b95a7;
  font-size: 12px;
  line-height: 1.5;
}

.batch-overview-result {
  display: inline-flex;
  align-items: center;
  min-height: 30px;
  padding: 0 12px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 700;
}

.batch-overview-result--ready {
  background: #edf9f1;
  color: #16a34a;
}

.batch-overview-result--warning {
  background: #fff5e8;
  color: #d97706;
}

.batch-overview-result--info {
  background: #eef5ff;
  color: #1f6dff;
}

.batch-overview-result--danger {
  background: #fff1f2;
  color: #dc2626;
}

.batch-layout {
  display: grid;
  grid-template-columns: minmax(0, 1.7fr) minmax(320px, 0.98fr);
  gap: 16px;
}

.batch-main-panel,
.batch-side-card {
  border: 1px solid #e8eef6;
  border-radius: 20px;
  background: #fff;
  box-shadow: 0 10px 30px rgb(15 23 42 / 4%);
}

.batch-main-panel {
  padding: 18px;
}

.batch-section + .batch-section {
  margin-top: 20px;
  padding-top: 20px;
  border-top: 1px dashed #ecf1f6;
}

.batch-section__header {
  margin-bottom: 14px;
}

.batch-section__title,
.batch-side-card__title {
  color: #1f2329;
  font-size: 16px;
  font-weight: 700;
  line-height: 1.3;
}

.batch-section__desc {
  margin-top: 4px;
  color: #8b95a7;
  font-size: 12px;
  line-height: 1.6;
}

.batch-field-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px;
}

.batch-recommend-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
}

.batch-recommend-card {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 10px;
  padding: 12px;
  border: 1px solid #dbe3ee;
  border-radius: 18px;
  background: linear-gradient(180deg, #fff 0%, #fbfcff 100%);
  text-align: left;
  cursor: pointer;
  transition:
    border-color 0.18s ease,
    box-shadow 0.18s ease,
    transform 0.18s ease;
}

.batch-recommend-card:hover {
  border-color: #a9c6ff;
  transform: translateY(-1px);
}

.batch-recommend-card--active {
  border-color: #1f6dff;
  background: linear-gradient(180deg, #f8fbff 0%, #eef5ff 100%);
  box-shadow: 0 10px 24px rgb(31 109 255 / 8%);
}

.batch-recommend-card__top {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  flex-wrap: nowrap;
}

.batch-recommend-card__badge {
  display: inline-flex;
  align-items: center;
  height: 24px;
  padding: 0 8px;
  border-radius: 999px;
  background: #eef5ff;
  color: #1f6dff;
  font-size: 11px;
  font-weight: 700;
  flex-shrink: 0;
}

.batch-recommend-card__title {
  min-width: 0;
  flex: 1;
  color: #1f2329;
  font-size: 14px;
  font-weight: 700;
  line-height: 1.35;
}

.batch-recommend-card__desc {
  color: #667085;
  font-size: 12px;
  line-height: 1.6;
}

.batch-recommend-card__metrics {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.batch-recommend-card__metric {
  display: inline-flex;
  align-items: center;
  height: 26px;
  padding: 0 10px;
  border-radius: 999px;
  background: #f4f7fb;
  color: #5f6b7c;
  font-size: 11px;
  font-weight: 600;
}

.batch-field-grid--compact {
  margin-top: 16px;
}

.batch-one-to-one-card {
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding: 16px 20px;
  border: 1px solid #dbe3ee;
  border-radius: 16px;
  background: #fcfdfe;
}

.batch-one-to-one-card__top {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
}

.batch-one-to-one-card__title {
  color: #1f2329;
  font-size: 16px;
  font-weight: 700;
  line-height: 1.35;
}

.batch-one-to-one-card__meta {
  margin-top: 6px;
  color: #7c8698;
  font-size: 13px;
  line-height: 1.5;
}

.batch-one-to-one-card__chips {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  justify-content: flex-end;
}

.batch-one-to-one-card__chip {
  display: inline-flex;
  align-items: center;
  height: 24px;
  padding: 0 10px;
  border-radius: 999px;
  background: #edf9f1;
  color: #16a34a;
  font-size: 11px;
  font-weight: 700;
}

.batch-one-to-one-card__chip--sub {
  background: #f4f7fb;
  color: #586174;
}

.batch-one-to-one-card__grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
  padding: 16px 0;
  border-top: 1px dashed #edf2f7;
  border-bottom: 1px dashed #edf2f7;
}

.batch-one-to-one-card__item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.batch-one-to-one-card__item-label {
  color: #8b95a7;
  font-size: 12px;
  line-height: 1.5;
}

.batch-one-to-one-card__item-value {
  color: #1f2329;
  font-size: 13px;
  font-weight: 600;
  line-height: 1.4;
}

.batch-one-to-one-card__remark {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  color: #617086;
  font-size: 12px;
  line-height: 1.6;
}

.batch-one-to-one-card__remark-label {
  display: inline-flex;
  align-items: center;
  height: 20px;
  padding: 0 6px;
  border-radius: 4px;
  background: #f0f4fa;
  color: #5f6b7c;
  font-weight: 600;
}

.batch-field {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.batch-field--full {
  grid-column: 1 / -1;
}

.batch-field__label {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: #586174;
  font-size: 13px;
  font-weight: 600;
}

.batch-field__control {
  width: 100%;
}

.batch-weekday-group,
.batch-rule-group {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.batch-weekday-chip,
.batch-rule-chip,
.batch-strategy-card {
  border: 1px solid #dbe3ee;
  background: #fff;
  cursor: pointer;
  transition:
    border-color 0.18s ease,
    box-shadow 0.18s ease,
    background-color 0.18s ease,
    transform 0.18s ease;
}

.batch-weekday-chip,
.batch-rule-chip {
  height: 36px;
  padding: 0 14px;
  border-radius: 12px;
  color: #556070;
  font-size: 13px;
  font-weight: 600;
}

.batch-weekday-chip:hover,
.batch-rule-chip:hover {
  border-color: #9ec0ff;
  color: #1f6dff;
}

.batch-weekday-chip--active,
.batch-rule-chip--active {
  border-color: #1f6dff;
  background: #eef5ff;
  color: #1f6dff;
  box-shadow: inset 0 0 0 1px rgb(31 109 255 / 12%);
}

.batch-time-range {
  display: grid;
  grid-template-columns: 1fr 24px 1fr;
  gap: 8px;
  align-items: center;
}

.batch-time-range__split {
  color: #98a2b3;
  font-size: 12px;
  text-align: center;
}

.batch-strategy-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.batch-strategy-grid--secondary {
  margin-top: 12px;
}

.batch-strategy-card {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 8px;
  padding: 14px 16px;
  border-radius: 16px;
  text-align: left;
}

.batch-strategy-card:hover {
  border-color: #aac7ff;
  transform: translateY(-1px);
}

.batch-strategy-card--active {
  border-color: #1f6dff;
  background: linear-gradient(180deg, #f8fbff 0%, #eef5ff 100%);
  box-shadow: 0 8px 18px rgb(31 109 255 / 8%);
}

.batch-strategy-card__title {
  color: #1f2329;
  font-size: 14px;
  font-weight: 700;
  line-height: 1.3;
}

.batch-strategy-card__desc {
  color: #7c8698;
  font-size: 12px;
  line-height: 1.6;
}

.batch-side-panel {
  display: flex;
  flex-direction: column;
  gap: 16px;
  height: 100%;
  min-height: 0;
}

.batch-side-card {
  padding: 16px;
}

.batch-side-card__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 12px;
}

.batch-side-card--summary {
  background: linear-gradient(180deg, #1f6dff 0%, #1558db 100%);
  color: #fff;
}

.batch-side-card--summary .batch-side-card__title {
  color: rgb(255 255 255 / 92%);
}

.batch-summary-board {
  padding: 16px;
  border-radius: 18px;
  background: rgb(255 255 255 / 12%);
  backdrop-filter: blur(10px);
}

.batch-summary-board__headline {
  color: rgb(255 255 255 / 74%);
  font-size: 12px;
}

.batch-summary-board__count {
  margin-top: 8px;
  font-size: 24px;
  font-weight: 700;
  line-height: 1.2;
}

.batch-summary-board__meta {
  margin-top: 8px;
  color: rgb(255 255 255 / 82%);
  font-size: 12px;
  line-height: 1.6;
}

.batch-summary-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 14px;
}

.batch-summary-tag {
  display: inline-flex;
  align-items: center;
  padding: 8px 10px;
  border-radius: 999px;
  background: rgb(255 255 255 / 12%);
  color: rgb(255 255 255 / 86%);
  font-size: 12px;
  line-height: 1;
}

.batch-preview-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  padding-right: 4px;
}

.batch-side-card--preview-list {
  display: flex;
  flex-direction: column;
  flex: 1;
  min-height: 320px;
}

.batch-side-card__extra {
  color: #98a2b3;
  font-size: 12px;
  line-height: 1;
  white-space: nowrap;
}

.batch-side-card__helper {
  margin-top: 4px;
  color: #8b95a7;
  font-size: 10px;
  line-height: 1.5;
}

.batch-preview-row {
  padding: 12px 14px;
  border: 1px solid #edf2f7;
  border-radius: 16px;
  background: linear-gradient(180deg, #fff 0%, #fbfcff 100%);
}

.batch-preview-row__top {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
}

.batch-preview-row__main {
  min-width: 0;
  flex: 1;
}

.batch-preview-row__date {
  color: #1f2329;
  font-size: 13px;
  font-weight: 700;
  line-height: 1.2;
}

.batch-preview-row__meta {
  margin-top: 4px;
  overflow: hidden;
  color: #7c8698;
  font-size: 12px;
  line-height: 1.45;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.batch-preview-row__side {
  display: flex;
  align-items: flex-end;
  flex-direction: column;
  gap: 4px;
  flex-shrink: 0;
}

.batch-preview-row__bottom {
  margin-top: 8px;
}

.batch-preview-row__reasons {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.batch-preview-row__reason-chip {
  display: inline-flex;
  align-items: center;
  height: 22px;
  padding: 0 8px;
  border-radius: 999px;
  background: #f4f7fb;
  color: #617086;
  font-size: 10px;
  font-weight: 600;
}

.batch-preview-row__note {
  margin-top: 6px;
  font-size: 10px;
  line-height: 1.55;
}

.batch-preview-row__note--risk {
  color: #d97706;
}

.batch-preview-row__note--resolved {
  color: #15803d;
}

.batch-preview-row__status {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 64px;
  height: 26px;
  padding: 0 8px;
  border-radius: 999px;
  font-size: 11px;
  font-weight: 700;
  white-space: nowrap;
}

.batch-preview-row__status--ready {
  background: #edf9f1;
  color: #16a34a;
}

.batch-preview-row__status--warning {
  background: #fff5e8;
  color: #d97706;
}

.batch-preview-row__status--blocked {
  background: #fff1f2;
  color: #dc2626;
}

.batch-preview-row__status--deferred {
  background: #eef5ff;
  color: #1f6dff;
}

.batch-preview-row__actions {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 2px;
}

.batch-preview-row__action {
  height: auto;
  display: inline-flex;
  padding: 0 !important;
  font-size: 11px;
  line-height: 1.5;
  white-space: nowrap;
}

:deep(.batch-preview-row__action.ant-btn-link) {
  color: #1f6dff;
}

.batch-preview-row__notes {
  display: flex;
  flex-direction: column;
  gap: 4px;
  margin-top: 8px;
}

.batch-preview-list::-webkit-scrollbar {
  width: 6px;
}

.batch-preview-list::-webkit-scrollbar-thumb {
  border-radius: 999px;
  background: #d7deea;
}

.batch-preview-list::-webkit-scrollbar-track {
  background: transparent;
}

.batch-suggestion-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin-top: 12px;
}

.batch-insight-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.batch-insight-item {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  padding: 12px 14px;
  border-radius: 14px;
  background: #f8fafc;
}

.batch-insight-item__label {
  color: #1f2329;
  font-size: 13px;
  font-weight: 700;
  line-height: 1.4;
}

.batch-insight-item__desc {
  margin-top: 4px;
  color: #7c8698;
  font-size: 12px;
  line-height: 1.6;
}

.batch-insight-item__value {
  color: #1f6dff;
  font-size: 18px;
  font-weight: 700;
  line-height: 1.2;
  white-space: nowrap;
}

.batch-suggestion-item {
  padding: 12px 14px;
  border-radius: 14px;
  background: #f8fafc;
  color: #667085;
  font-size: 12px;
  line-height: 1.7;
}

.batch-modal-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.batch-modal-footer__note {
  color: #8b95a7;
  font-size: 12px;
  line-height: 1.6;
}

.batch-modal-footer__actions {
  display: flex;
  gap: 12px;
}

:deep(.one-to-one-schedule-modal .ant-modal-content) {
  border-radius: 24px;
  overflow: hidden;
}

:deep(.one-to-one-schedule-modal .ant-modal-header) {
  padding: 20px 24px 12px;
  margin-bottom: 0;
  border-bottom: 0;
}

:deep(.one-to-one-schedule-modal .ant-modal-body) {
  padding: 0 24px 24px;
  max-height: calc(100vh - 140px);
  overflow-y: auto;
}

:deep(.one-to-one-schedule-modal .ant-select-selector),
:deep(.one-to-one-schedule-modal .ant-picker),
:deep(.one-to-one-schedule-modal .ant-input),
:deep(.one-to-one-schedule-modal .ant-input-affix-wrapper),
:deep(.one-to-one-schedule-modal .ant-input-textarea textarea) {
  border-radius: 12px;
}

:deep(.one-to-one-schedule-modal .ant-select-selector),
:deep(.one-to-one-schedule-modal .ant-picker),
:deep(.one-to-one-schedule-modal .ant-input),
:deep(.one-to-one-schedule-modal .ant-input-affix-wrapper) {
  min-height: 44px;
  border-color: #dbe3ee;
  box-shadow: none !important;
}

:deep(.one-to-one-schedule-modal .ant-input-textarea textarea) {
  border-color: #dbe3ee;
}

@media (max-width: 1100px) {
  .batch-overview__top,
  .batch-overview__metrics,
  .batch-layout {
    grid-template-columns: 1fr;
  }

  .batch-field-grid,
  .batch-strategy-grid,
  .batch-recommend-grid {
    grid-template-columns: 1fr;
  }

  .batch-preview-row__top {
    flex-direction: column;
  }

  .batch-preview-row__side {
    align-items: flex-start;
  }

  .batch-preview-row__actions {
    align-items: flex-start;
  }

  .batch-modal-footer {
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>
