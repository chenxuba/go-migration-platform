<script setup>
import { LeftOutlined, RightOutlined } from '@ant-design/icons-vue'
import { Modal } from 'ant-design-vue'
import dayjs from 'dayjs'
import { h } from 'vue'
import CreateSchedulePopover from './create-schedule-popover.vue'
import { getOneToOneListApi } from '@/api/edu-center/one-to-one'
import { checkOneToOneScheduleAvailabilityApi, createOneToOneSchedulesApi, listTeachingSchedulesByTeacherMatrixApi } from '@/api/edu-center/teaching-schedule'
import { useUserStore } from '@/stores/user'
import messageService from '@/utils/messageService'
import {
  DEFAULT_UNIFIED_TIME_PERIOD_CONFIG,
  buildQuickHourlySlots,
  configGroupsSorted,
  parseUnifiedTimePeriodConfig,
} from '@/utils/unified-time-period'
import emitter, { EVENTS } from '@/utils/eventBus'

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
/** 课表时间视图：下拉与日期导航联动 */
const timeViewOptions = [
  { value: 'day', label: '日视图' },
  { value: 'week', label: '周视图' },
]
/** 1=1v1，2=班课 */
const currentModel = ref('1')
const currentGroup = ref('A')
/** 与 matrixDays、表头节次列对齐；切换 A/B 时在新数据返回前不改，避免清空矩阵导致整页高度塌缩抖动 */
const displayedGroupKey = ref('A')
const timetableRootRef = ref(null)
/** 当前选中的 1 对 1 记录 id（非学员 id，避免同一学员多门课冲突） */
const oneToOneRecordId = ref(undefined)
const studentIds = ref([])
const courseId = ref(null)
const courseName = ref(null)
const classId = ref(null)
const className = ref(null)
const teacherId = ref(null)

function getWeekStart(value = dayjs()) {
  const d = dayjs(value)
  const diff = d.day() === 0 ? -6 : 1 - d.day()
  return d.add(diff, 'day').startOf('day')
}

// 监听时间维度变化
watch(currentTime, () => {
  currentWeek.value = dayjs()
})

// 格式化日期显示
function formatDateRange(value) {
  if (!value)
    return ''

  switch (currentTime.value) {
    case 'day':
      return value.format('YYYY年MM月DD日')
    case 'week': {
      const start = getWeekStart(value)
      const end = start.add(6, 'day')
      if (start.year() === end.year() && start.month() === end.month()) {
        return `${start.format('YYYY年MM月DD日')} ~ ${end.format('DD日')}`
      }
      else if (start.year() === end.year()) {
        return `${start.format('YYYY年MM月DD日')} ~ ${end.format('MM月DD日')}`
      }
      else {
        return `${start.format('YYYY年MM月DD日')} ~ ${end.format('YYYY年MM月DD日')}`
      }
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

function handleThisWeek() {
  currentWeek.value = dayjs()
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
const userStore = useUserStore()
const matrixDays = ref([])
const timetableLoading = ref(false)
const oneToOneAvailabilityLoading = ref(false)
const creatingOneToOneSchedule = ref(false)
const conflictDetailModalOpen = ref(false)
const conflictDetailState = ref({
  summary: '',
  attempted: null,
  items: [],
})
/** 防止快速切换周次/组别时旧请求晚到覆盖新矩阵 */
let matrixLoadSeq = 0
let oneToOneAvailabilitySeq = 0
let pendingConflictJump = null
let focusedScheduleCellTimer = null
const focusedScheduleCellKey = ref('')

const displayDates = computed(() => {
  if (currentTime.value === 'day')
    return [dayjs(currentWeek.value).startOf('day')]
  const start = getWeekStart(currentWeek.value)
  return Array.from({ length: 7 }, (_, i) => start.add(i, 'day'))
})

const queryDateRange = computed(() => {
  const dates = displayDates.value
  if (!dates.length) {
    const d = dayjs().format('YYYY-MM-DD')
    return { startDate: d, endDate: d }
  }
  return {
    startDate: dates[0].format('YYYY-MM-DD'),
    endDate: dates[dates.length - 1].format('YYYY-MM-DD'),
  }
})

const periodConfig = computed(() => {
  const parsed = parseUnifiedTimePeriodConfig(userStore.instConfig?.unifiedTimePeriodJson)
  return parsed ?? DEFAULT_UNIFIED_TIME_PERIOD_CONFIG
})

const sortedPeriodGroups = computed(() => configGroupsSorted(periodConfig.value))

const groupOptions = computed(() => {
  const g = sortedPeriodGroups.value
  if (!g.length)
    return [{ key: 'A', label: '默认时段' }]
  if (g.length === 1)
    return [{ key: 'A', label: g[0].name || '时段' }]
  return [
    { key: 'A', label: g[0].name || 'A时段' },
    { key: 'B', label: g[1].name || 'B时段' },
  ]
})

function slotsForGroupKey(key) {
  const groups = sortedPeriodGroups.value
  const fallback = buildQuickHourlySlots().filter(s => s.enabled !== false)
  if (!groups.length)
    return [...fallback].sort((a, b) => a.index - b.index)
  const idx = key === 'B' ? 1 : 0
  const g = groups[idx] || groups[0]
  return [...g.slots].filter(s => s.enabled !== false).sort((a, b) => a.index - b.index)
}

const activePeriodSlots = computed(() => slotsForGroupKey(displayedGroupKey.value))

function periodGroupForKey(key) {
  const groups = sortedPeriodGroups.value
  if (!groups.length)
    return null
  const idx = key === 'B' ? 1 : 0
  return groups[idx] || groups[0] || null
}

/** 矩阵接口：时段组 UUID + 回退 teacherIds（按请求时的组别快照，避免加载途中切换导致参数错位） */
function teacherMatrixGroupParamsForKey(key) {
  const g = periodGroupForKey(key)
  if (!g)
    return {}
  const periodGroupUuid = String(g.id || '').trim()
  const bound = g.boundTeachers
  const ids = Array.isArray(bound)
    ? bound.map(t => String(t.id ?? '').trim()).filter(Boolean)
    : []
  return {
    ...(periodGroupUuid ? { periodGroupUuid } : {}),
    ...(ids.length ? { matrixTeacherIds: ids.join(',') } : {}),
  }
}

function normalizeHHMM(t) {
  const s = String(t || '').trim()
  if (!s)
    return ''
  return s.length >= 5 ? s.slice(0, 5) : s
}

function minutesFromHHMM(t) {
  const n = normalizeHHMM(t)
  const m = /^(\d{1,2}):(\d{2})$/.exec(n)
  if (!m)
    return null
  const h = Number(m[1])
  const mi = Number(m[2])
  if (!Number.isFinite(h) || !Number.isFinite(mi))
    return null
  return h * 60 + mi
}

function emptyLessonCell(slot) {
  return {
    startTime: slot.start,
    endTime: slot.end,
    courseName: null,
    studentId: null,
    classId: null,
    className: null,
    studentNames: null,
    courseType: null,
    conflict: false,
    conflictReason: null,
    serverConflict: false,
    serverConflictReason: null,
    isMain: undefined,
  }
}

function buildLessonsForRow(slots, legacyList) {
  const lessons = slots.map(s => emptyLessonCell(s))
  const list = Array.isArray(legacyList) ? legacyList : []
  for (const leg of list) {
    const st = normalizeHHMM(leg.scheduleStartTime)
    const et = normalizeHHMM(leg.scheduleEndTime)
    const sm = minutesFromHHMM(st)
    const em = minutesFromHHMM(et)
    let idx = lessons.findIndex(l => l.startTime === st && l.endTime === et)
    if (idx < 0 && sm != null && em != null) {
      let best = -1
      let bestOverlap = 0
      lessons.forEach((l, i) => {
        const ls = minutesFromHHMM(l.startTime)
        const le = minutesFromHHMM(l.endTime)
        if (ls == null || le == null)
          return
        const ov = Math.min(le, em) - Math.max(ls, sm)
        if (ov > bestOverlap) {
          bestOverlap = ov
          best = i
        }
      })
      if (best >= 0 && bestOverlap > 0)
        idx = best
    }
    if (idx < 0)
      continue

    const displayCourseType = leg.courseType === 2 ? 1 : 2
    const studentIds = (leg.studentList || []).map(s => String(s.id)).filter(Boolean)
    const names = (leg.studentList || []).map(s => ({ id: String(s.id), name: s.name }))

    lessons[idx] = {
      ...lessons[idx],
      courseName: leg.courseName || null,
      studentId: studentIds.length ? studentIds : null,
      classId: leg.classId != null ? String(leg.classId) : null,
      className: leg.className || null,
      studentNames: names.length ? names : null,
      courseType: displayCourseType,
      isMain: true,
    }
  }
  return lessons
}

const gridRows = computed(() => {
  const slots = activePeriodSlots.value
  const dates = displayDates.value.map(d => d.format('YYYY-MM-DD'))
  if (!slots.length || !dates.length)
    return []

  let teacherCols = []
  for (const d of dates) {
    const day = matrixDays.value.find(x => x.scheduleDate === d)
    if (day?.scheduleListVoList?.length) {
      teacherCols = day.scheduleListVoList
      break
    }
  }
  if (!teacherCols.length && matrixDays.value[0]?.scheduleListVoList)
    teacherCols = matrixDays.value[0].scheduleListVoList || []

  if (!teacherCols.length)
    return []

  const dayMap = new Map()
  for (const day of matrixDays.value) {
    const m = new Map()
    for (const col of day.scheduleListVoList || []) {
      m.set(String(col.teacherId), col.scheduleInfoVoList || [])
    }
    dayMap.set(day.scheduleDate, m)
  }

  const rows = []
  for (const col of teacherCols) {
    const tid = String(col.teacherId)
    const tname = col.teacherName
    for (const dateStr of dates) {
      const schedules = dayMap.get(dateStr)?.get(tid) || []
      rows.push({
        key: `${tid}-${dateStr}`,
        name: tname,
        teacherId: tid,
        date: dateStr,
        lessons: buildLessonsForRow(slots, schedules),
      })
    }
  }
  return rows
})

const dataSource = computed(() => {
  return [...gridRows.value].sort((a, b) => {
    if (a.teacherId !== b.teacherId)
      return a.teacherId.localeCompare(b.teacherId)
    return a.date.localeCompare(b.date)
  })
})

function syncLessonConflictState(lesson) {
  const serverReason = lesson.serverConflict ? lesson.serverConflictReason : null
  lesson.conflict = Boolean(serverReason)
  lesson.conflictReason = serverReason || null
}

function clearLessonConflictState(lesson, scope = 'all') {
  if (scope === 'all' || scope === 'server') {
    lesson.serverConflict = false
    lesson.serverConflictReason = null
  }
  if (scope === 'all') {
    lesson.conflict = false
    lesson.conflictReason = null
  }
  else {
    syncLessonConflictState(lesson)
  }
}

function resetEmptyLessonConflicts(scope = 'all') {
  dataSource.value.forEach((teacher) => {
    teacher.lessons.forEach((lesson) => {
      if (!lesson.studentId)
        clearLessonConflictState(lesson, scope)
    })
  })
}

/** 当前展示范围内每位老师已占用的节次数（与格子里蓝色已排课一致：有 studentId 即计入） */
const scheduledLessonCountByTeacher = computed(() => {
  const map = new Map()
  for (const row of dataSource.value) {
    const tid = String(row.teacherId)
    let n = 0
    for (const lesson of row.lessons || []) {
      if (lesson.studentId)
        n++
    }
    map.set(tid, (map.get(tid) || 0) + n)
  }
  return map
})

function teacherScheduledLessonCount(teacherId) {
  if (teacherId == null)
    return 0
  return scheduledLessonCountByTeacher.value.get(String(teacherId)) ?? 0
}

function teacherLessonCountLabel(teacherId) {
  const n = teacherScheduledLessonCount(teacherId)
  const scope = currentTime.value === 'day' ? '当日' : '本周'
  return `${scope}共${n}节课`
}

const activeGroupLabel = computed(() => {
  return groupOptions.value.find(o => o.key === displayedGroupKey.value)?.label || ''
})

watch(
  groupOptions,
  (opts) => {
    if (!opts.some(o => o.key === currentGroup.value))
      currentGroup.value = 'A'
  },
  { immediate: true },
)

async function loadTimetableMatrix() {
  const seq = ++matrixLoadSeq
  const requestedGroup = currentGroup.value
  timetableLoading.value = true
  try {
    await userStore.getInstConfig()
    if (seq !== matrixLoadSeq)
      return
    const { startDate, endDate } = queryDateRange.value
    const classType = currentModel.value === '1' ? 2 : 1
    const res = await listTeachingSchedulesByTeacherMatrixApi({
      startDate,
      endDate,
      classType,
      ...teacherMatrixGroupParamsForKey(requestedGroup),
    })
    if (seq !== matrixLoadSeq)
      return
    if (res.code === 200 && Array.isArray(res.result))
      matrixDays.value = res.result
    else
      matrixDays.value = []
    displayedGroupKey.value = requestedGroup
    await nextTick()
    if (seq !== matrixLoadSeq)
      return
    if (currentModel.value === '1' && oneToOneRecordId.value)
      await detectOneToOneAvailability(oneToOneRecordId.value)
    else if (currentModel.value === '2' && classId.value)
      handleClass(classId.value)
    else
      resetEmptyLessonConflicts()
    await flushPendingConflictJump()
  }
  catch (e) {
    console.error('loadTimetableMatrix failed', e)
  }
  finally {
    if (seq === matrixLoadSeq)
      timetableLoading.value = false
  }
}

watch(
  [currentWeek, currentTime, currentModel, currentGroup],
  () => {
    void loadTimetableMatrix()
  },
)

function refreshTimetableRelatedData() {
  void loadTimetableMatrix()
  void fetchOneToOneOptionsForTimetable()
}

onMounted(() => {
  void loadTimetableMatrix()
  void fetchOneToOneOptionsForTimetable()
  emitter.on(EVENTS.REFRESH_DATA, refreshTimetableRelatedData)
})

onUnmounted(() => {
  emitter.off(EVENTS.REFRESH_DATA, refreshTimetableRelatedData)
  if (focusedScheduleCellTimer)
    clearTimeout(focusedScheduleCellTimer)
})
const columns = computed(() => {
  const slots = activePeriodSlots.value
  const baseColumns = [
    {
      title: '教师',
      dataIndex: 'name',
      key: 'name',
      width: 120,
      align: 'center',
      fixed: 'left',
      customCell: (_, index) => {
        if (!dataSource.value.length)
          return {}
        const currentTeacherId = dataSource.value[index].teacherId
        if (index === 0 || dataSource.value[index - 1].teacherId !== currentTeacherId) {
          let count = 1
          for (let i = index + 1; i < dataSource.value.length; i++) {
            if (dataSource.value[i].teacherId === currentTeacherId)
              count++
            else
              break
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

  const lessonColumns = slots.map((slot, index) => ({
    title: `第${slot.index}节课`,
    startTime: slot.start,
    endTime: slot.end,
    dataIndex: ['lessons', index],
    key: `lesson-${slot.index}-${index}`,
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
/** 开班中的 1 对 1 列表（来自 getOneToOneListApi，下拉 value 为 record.id） */
const oneToOneData = ref([])
const oneToOneListLoading = ref(false)

function mapRowToOneToOneOption(row) {
  const id = String(row.id || '').trim()
  const studentId = String(row.studentId || '').trim()
  const studentName = String(row.studentName || '').trim()
  const lessonName = String(row.lessonName || '').trim()
  const name = String(row.name || '').trim()
    || (studentName && lessonName ? `${studentName}-${lessonName}` : studentName || lessonName || id)
  return {
    id,
    studentId,
    studentName,
    courseId: row.lessonId != null ? String(row.lessonId) : '',
    courseName: lessonName,
    name,
  }
}

async function fetchOneToOneOptionsForTimetable() {
  oneToOneListLoading.value = true
  try {
    const res = await getOneToOneListApi({
      pageRequestModel: {
        needTotal: false,
        pageSize: 500,
        pageIndex: 1,
        skipCount: 0,
      },
      queryModel: {
        /** 与一对一列表页默认一致：仅开班中 */
        status: [1],
      },
    })
    if (res.code === 200 && res.result) {
      const list = Array.isArray(res.result.list) ? res.result.list : []
      oneToOneData.value = list.map(mapRowToOneToOneOption).filter(item => item.id)
    }
    else {
      oneToOneData.value = []
      messageService.error(res.message || '获取1对1列表失败')
    }
  }
  catch (e) {
    console.error('fetchOneToOneOptionsForTimetable', e)
    oneToOneData.value = []
    messageService.error('获取1对1列表失败')
  }
  finally {
    oneToOneListLoading.value = false
  }
}

function filterOneToOneOption(input, option) {
  const q = (input || '').trim().toLowerCase()
  if (!q)
    return true
  const id = option?.value != null ? String(option.value) : ''
  const item = oneToOneData.value.find(r => r.id === id)
  if (!item)
    return true
  const blob = `${item.name} ${item.studentName} ${item.courseName} ${item.studentId}`.toLowerCase()
  return blob.includes(q)
}

// 当前视图下的全部行（时段 A/B 切换后数据源已重建；跨组检测以当前页为准）
const allDataSource = computed(() => dataSource.value)

function buildAvailabilitySlotKey(teacherId, lessonDate, startTime, endTime) {
  return `${String(teacherId)}|${lessonDate}|${startTime}|${endTime}`
}

function parseConflictTimeRange(timeText) {
  const m = String(timeText || '').trim().match(/(\d{2}:\d{2})\s*[~-]\s*(\d{2}:\d{2})/)
  if (!m)
    return null
  return { startTime: m[1], endTime: m[2] }
}

function resolveConflictScheduleGroupInfo(item) {
  const teacherId = String(item?.teacherId || '').trim()
  const timeRange = parseConflictTimeRange(item?.timeText)
  const matches = groupOptions.value
    .map((opt) => {
      const group = periodGroupForKey(opt.key)
      const teacherMatched = !teacherId || !(group?.boundTeachers?.length)
        || group.boundTeachers.some(t => String(t.id) === teacherId)
      const timeMatched = !timeRange || slotsForGroupKey(opt.key).some(slot =>
        slot.start === timeRange.startTime && slot.end === timeRange.endTime,
      )
      return teacherMatched && timeMatched
        ? { key: opt.key, label: opt.label }
        : null
    })
    .filter(Boolean)

  const unique = matches.filter((item, index, arr) =>
    arr.findIndex(x => x.key === item.key) === index,
  )
  return {
    keys: unique.map(item => item.key),
    labels: unique.map(item => item.label),
  }
}

function resolveConflictScheduleGroupLabel(item) {
  return resolveConflictScheduleGroupInfo(item).labels.join('/')
}

function setFocusedScheduleCell(key) {
  focusedScheduleCellKey.value = key || ''
  if (focusedScheduleCellTimer)
    clearTimeout(focusedScheduleCellTimer)
  if (key) {
    focusedScheduleCellTimer = window.setTimeout(() => {
      if (focusedScheduleCellKey.value === key)
        focusedScheduleCellKey.value = ''
    }, 3000)
  }
}

async function focusScheduleCell(key) {
  await nextTick()
  const root = timetableRootRef.value
  if (!root || !key)
    return false
  const cell = Array.from(root.querySelectorAll('[data-schedule-cell-key]')).find(
    el => el.getAttribute('data-schedule-cell-key') === key,
  )
  if (!cell)
    return false
  cell.scrollIntoView({
    behavior: 'smooth',
    block: 'center',
    inline: 'center',
  })
  setFocusedScheduleCell(key)
  return true
}

async function flushPendingConflictJump() {
  if (!pendingConflictJump?.cellKey)
    return
  const pending = pendingConflictJump
  const found = await focusScheduleCell(pending.cellKey)
  if (found) {
    pendingConflictJump = null
    messageService.success('已定位到冲突课程')
  }
}

function openApiConflictModal(reason, column, record) {
  const existingSchedules = Array.isArray(reason?.existingSchedules) ? reason.existingSchedules : []
  const items = existingSchedules.map((item, index) => {
    const groupInfo = resolveConflictScheduleGroupInfo(item)
    const timeRange = parseConflictTimeRange(item.timeText)
    const jumpGroupKey = groupInfo.keys.includes(currentGroup.value)
      ? currentGroup.value
      : groupInfo.keys[0] || ''
    return {
      key: `${item.teacherId || item.teacherName || 'teacher'}-${item.date}-${item.timeText}-${index}`,
      name: item.name || '-',
      classTypeText: item.classTypeText || '日程',
      date: item.date,
      week: item.week || '',
      timeText: item.timeText,
      teacherId: item.teacherId || '',
      teacherName: item.teacherName || '-',
      groupLabel: groupInfo.labels.join('/') || '未知组别',
      classroomName: item.classroomName || '-',
      studentText: (item.studentNames || []).join('、') || '-',
      conflictTypes: item.conflictTypes || [],
      jumpCellKey: timeRange && item.teacherId
        ? buildAvailabilitySlotKey(item.teacherId, item.date, timeRange.startTime, timeRange.endTime)
        : '',
      jumpGroupKey,
    }
  })

  conflictDetailState.value = {
    summary: `${reason.message || '当前空位存在时间冲突'}，共发现 ${items.length} 条冲突日程。`,
    attempted: {
      date: record.date,
      week: formatWeek(record.date),
      timeText: `${column.startTime}-${column.endTime}`,
      teacherName: record.name,
      lessonIndex: getLessonIndex(column.startTime),
      groupLabel: activeGroupLabel.value || '当前组',
    },
    items,
  }
  conflictDetailModalOpen.value = true
}

async function jumpToConflictSchedule(item) {
  if (!item?.jumpCellKey) {
    messageService.warning('当前冲突课程暂不支持定位')
    return
  }

  conflictDetailModalOpen.value = false
  pendingConflictJump = {
    cellKey: item.jumpCellKey,
  }

  let needReload = false
  if (item.jumpGroupKey && item.jumpGroupKey !== currentGroup.value) {
    currentGroup.value = item.jumpGroupKey
    needReload = true
  }

  const targetDate = String(item.date || '').trim()
  if (targetDate) {
    if (currentTime.value === 'day') {
      if (dayjs(currentWeek.value).format('YYYY-MM-DD') !== targetDate) {
        currentWeek.value = dayjs(targetDate)
        needReload = true
      }
    }
    else {
      const { startDate, endDate } = queryDateRange.value
      if (targetDate < startDate || targetDate > endDate) {
        currentWeek.value = dayjs(targetDate)
        needReload = true
      }
    }
  }

  if (!needReload) {
    const found = await focusScheduleCell(item.jumpCellKey)
    if (found) {
      pendingConflictJump = null
      messageService.success('已定位到冲突课程')
    }
    else {
      await loadTimetableMatrix()
    }
  }
}

function buildCurrentOneToOneAvailabilityPayload(oneToOneId) {
  const schedules = []
  dataSource.value.forEach((teacher) => {
    teacher.lessons.forEach((lesson) => {
      if (!lesson.studentId) {
        schedules.push({
          teacherId: String(teacher.teacherId),
          lessonDate: teacher.date,
          startTime: lesson.startTime,
          endTime: lesson.endTime,
        })
      }
    })
  })
  return {
    oneToOneId: String(oneToOneId || ''),
    schedules,
  }
}

function buildAvailabilityConflictDetail(reason) {
  const schedules = Array.isArray(reason?.existingSchedules) ? reason.existingSchedules : []
  if (!schedules.length)
    return reason?.message || '该时间段不可排课'
  const detail = schedules.map((item) => {
    const groupText = resolveConflictScheduleGroupLabel(item)
    const studentText = (item.studentNames || []).length ? `，学员：${item.studentNames.join('、')}` : ''
    const groupSuffix = groupText ? `(${groupText})` : ''
    return `${item.date} ${item.timeText} ${item.teacherName || '-'}${groupSuffix} ${item.name}${studentText}`
  }).join('；')
  return `${reason.message || '该时间段不可排课'}。已有日程：${detail}`
}

function applyServerAvailabilityResult(result) {
  resetEmptyLessonConflicts('server')
  const invalidMap = new Map()
  const items = Array.isArray(result?.items) ? result.items : []
  items.forEach((item) => {
    if (item?.valid === false) {
      invalidMap.set(
        buildAvailabilitySlotKey(item.teacherId, item.lessonDate, item.startTime, item.endTime),
        item,
      )
    }
  })

  dataSource.value.forEach((teacher) => {
    teacher.lessons.forEach((lesson) => {
      if (lesson.studentId)
        return
      const matched = invalidMap.get(
        buildAvailabilitySlotKey(teacher.teacherId, teacher.date, lesson.startTime, lesson.endTime),
      )
      lesson.serverConflict = Boolean(matched)
      lesson.serverConflictReason = matched
        ? {
            type: '1v1-api',
            message: matched.message || '该时间段不可排课',
            conflictTypes: matched.conflictTypes || [],
            existingSchedules: matched.existingSchedules || [],
          }
        : null
      syncLessonConflictState(lesson)
    })
  })
}

async function detectOneToOneAvailability(value) {
  const seq = ++oneToOneAvailabilitySeq
  const oneToOneId = String(value || '').trim()
  if (!oneToOneId || currentModel.value !== '1') {
    oneToOneAvailabilityLoading.value = false
    resetEmptyLessonConflicts()
    return
  }

  if (!oneToOneData.value.some(item => item.id === oneToOneId)) {
    oneToOneAvailabilityLoading.value = false
    resetEmptyLessonConflicts()
    return
  }

  const payload = buildCurrentOneToOneAvailabilityPayload(oneToOneId)
  if (!payload.schedules.length) {
    oneToOneAvailabilityLoading.value = false
    resetEmptyLessonConflicts()
    return
  }

  oneToOneAvailabilityLoading.value = true
  try {
    const res = await checkOneToOneScheduleAvailabilityApi(payload)
    if (seq !== oneToOneAvailabilitySeq)
      return
    if (res.code !== 200 || !res.result)
      throw new Error(res.message || '检测课表空位失败')
    applyServerAvailabilityResult(res.result)
  }
  catch (error) {
    if (seq !== oneToOneAvailabilitySeq)
      return
    console.error('detectOneToOneAvailability failed', error)
    resetEmptyLessonConflicts('server')
    messageService.error(error?.response?.data?.message || error?.message || '检测课表空位失败')
  }
  finally {
    if (seq === oneToOneAvailabilitySeq)
      oneToOneAvailabilityLoading.value = false
  }
}

function handle1v1(value) {
  void detectOneToOneAvailability(value)
}

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
    resetEmptyLessonConflicts()
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

// 检查班课交叉时间冲突
function checkClassCrossTimeConflicts(classInfo) {
  console.log('运行班课冲突检测', classInfo)

  resetEmptyLessonConflicts()

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
          const conflictGroup = activeGroupLabel.value

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
          for (const t of allDataSource.value) {
            // 只检查同一天的课程
            if (t.date !== currentTime.date)
              continue

            const sameTimeLessons = t.lessons.filter(l =>
              l.studentId
              && isTimeOverlap(
                { start: l.startTime, end: l.endTime },
                { start: currentTime.startTime, end: currentTime.endTime },
              ),
            )

            let matchedStudentConflict = false
            for (const sameTimeLesson of sameTimeLessons) {
              if (sameTimeLesson.classId === classInfo.id)
                continue

              for (const sid of classInfo.studentIds) {
                if (sameTimeLesson.studentId?.includes(sid)) {
                  console.log('学生时间冲突', currentTime.date, currentTime.startTime, sameTimeLesson.startTime)
                  hasConflict = true

                  const studentIndex = classInfo.studentIds.indexOf(sid)
                  const studentName = studentIndex >= 0 ? classInfo.studentNames[studentIndex] : '未知学生'
                  const month = dayjs(t.date).format('M')
                  const day = dayjs(t.date).format('D')
                  const conflictGroup = activeGroupLabel.value

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

                  matchedStudentConflict = true
                  break
                }
              }
              if (matchedStudentConflict)
                break
            }

            if (matchedStudentConflict)
              break
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
function handleConflictClick(timeSlot, column, record) {
  let content = '该时间段已有课程安排，无法排课'

  // 根据冲突原因提供更详细的信息
  if (timeSlot.conflictReason) {
    const reason = timeSlot.conflictReason
    const groupInfo = reason.group ? `(${reason.group})` : ''
    const timeInfo = reason.time ? `[${reason.time}]` : ''

    if (reason.type === '1v1-api') {
      openApiConflictModal(reason, column, record)
      return
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

function buildConfirmField(label, value, valueColor = '#1f2329') {
  return h('div', {
    style: {
      display: 'grid',
      gridTemplateColumns: '76px 1fr',
      gap: '10px',
      alignItems: 'start',
      fontSize: '14px',
      lineHeight: '22px',
    },
  }, [
    h('div', {
      style: {
        color: '#8c8c8c',
      },
    }, label),
    h('div', {
      style: {
        color: valueColor,
        fontWeight: 600,
        wordBreak: 'break-word',
      },
    }, value || '-'),
  ])
}

const SMART_TIMETABLE_SKIP_CONFIRM_KEY = 'smart-timetable-skip-confirm-date'

function todayScheduleConfirmKey() {
  return dayjs().format('YYYY-MM-DD')
}

function shouldSkipScheduleConfirmToday() {
  if (typeof window === 'undefined')
    return false
  try {
    return window.localStorage.getItem(SMART_TIMETABLE_SKIP_CONFIRM_KEY) === todayScheduleConfirmKey()
  }
  catch {
    return false
  }
}

function setSkipScheduleConfirmToday(enabled) {
  if (typeof window === 'undefined')
    return
  try {
    if (enabled)
      window.localStorage.setItem(SMART_TIMETABLE_SKIP_CONFIRM_KEY, todayScheduleConfirmKey())
    else
      window.localStorage.removeItem(SMART_TIMETABLE_SKIP_CONFIRM_KEY)
  }
  catch {
  }
}

function buildScheduleConfirmContent({
  modeLabel,
  modeColor,
  targetLabel,
  targetValue,
  courseName,
  dateLabel,
  timeLabel,
  teacherName,
  groupLabel,
  onSkipTodayChange,
}) {
  return h('div', {
    style: {
      display: 'flex',
      flexDirection: 'column',
      gap: '14px',
      marginTop: '4px',
    },
  }, [
    h('div', {
      style: {
        display: 'flex',
        alignItems: 'center',
        gap: '10px',
        flexWrap: 'wrap',
      },
    }, [
      h('span', {
        style: {
          display: 'inline-flex',
          alignItems: 'center',
          justifyContent: 'center',
          minWidth: '52px',
          height: '28px',
          padding: '0 10px',
          borderRadius: '999px',
          background: modeColor,
          color: '#fff',
          fontSize: '13px',
          fontWeight: 700,
        },
      }, modeLabel),
      h('span', {
        style: {
          color: '#262626',
          fontSize: '16px',
          fontWeight: 700,
        },
      }, `${targetValue}${courseName ? ` · ${courseName}` : ''}`),
    ]),
    h('div', {
      style: {
        padding: '14px 16px',
        borderRadius: '14px',
        background: '#f8fafc',
        border: '1px solid #edf2f7',
        display: 'flex',
        flexDirection: 'column',
        gap: '10px',
      },
    }, [
      buildConfirmField(targetLabel, targetValue),
      buildConfirmField('上课时间', `${dateLabel} · ${timeLabel}`, '#1677ff'),
      buildConfirmField('上课老师', teacherName),
      buildConfirmField('所在组别', groupLabel || '当前组'),
    ]),
    h('div', {
      style: {
        padding: '12px 14px',
        borderRadius: '12px',
        background: '#fff7e6',
        color: '#ad6800',
        fontSize: '13px',
        lineHeight: '22px',
      },
    }, '确认后将立即创建日程并占用该时段；如果此时课表已被别人占用，系统会再次拦截。'),
    h('label', {
      style: {
        display: 'inline-flex',
        alignItems: 'center',
        gap: '8px',
        cursor: 'pointer',
        color: '#595959',
        fontSize: '13px',
        userSelect: 'none',
      },
    }, [
      h('input', {
        type: 'checkbox',
        onChange: event => onSkipTodayChange?.(Boolean(event?.target?.checked)),
        style: {
          width: '14px',
          height: '14px',
          cursor: 'pointer',
        },
      }),
      h('span', null, '今日不再提示'),
    ]),
  ])
}

function confirmScheduleWithOptionalSkip({
  modeLabel,
  modeColor,
  targetLabel,
  targetValue,
  courseName,
  dateLabel,
  timeLabel,
  teacherName,
  groupLabel,
  onConfirm,
}) {
  if (shouldSkipScheduleConfirmToday()) {
    return Promise.resolve(onConfirm())
  }

  let skipToday = false
  return new Promise((resolve, reject) => {
    Modal.confirm({
      title: '确认排课',
      width: 620,
      okText: '确认排课',
      cancelText: '再想想',
      content: buildScheduleConfirmContent({
        modeLabel,
        modeColor,
        targetLabel,
        targetValue,
        courseName,
        dateLabel,
        timeLabel,
        teacherName,
        groupLabel,
        onSkipTodayChange: (checked) => { skipToday = checked },
      }),
      async onOk() {
        if (skipToday)
          setSkipScheduleConfirmToday(true)
        try {
          const result = await onConfirm()
          resolve(result)
          return result
        }
        catch (error) {
          reject(error)
          throw error
        }
      },
      onCancel() {
        resolve(false)
      },
    })
  })
}

// 排课
function handleScheduleClick(timeSlot, column, record) {
  if (currentModel.value === '1') {
    if (!oneToOneRecordId.value) {
      Modal.warning({
        title: '请先选择一对一',
        content: '请先在上方选择要排课的 1 对 1 记录',
      })
      return
    }

    const studentInfo = oneToOneData.value.find(
      item => item.id === String(oneToOneRecordId.value),
    )

    if (!studentInfo) {
      Modal.warning({
        title: '记录不存在',
        content: '所选 1 对 1 已不在列表中，请重新选择或刷新页面',
      })
      return
    }

    // 获取月份和日期信息
    const dateObj = dayjs(record.date)
    const month = dateObj.format('M')
    const day = dateObj.format('D')
    const lessonIndex = getLessonIndex(column.startTime)

    void confirmScheduleWithOptionalSkip({
      modeLabel: '1v1',
      modeColor: '#1677ff',
      targetLabel: '排课对象',
      targetValue: studentInfo.studentName,
      courseName: studentInfo.courseName,
      dateLabel: `${month}月${day}日 ${formatWeek(record.date)} 第${lessonIndex}节`,
      timeLabel: `${column.startTime}-${column.endTime}`,
      teacherName: record.name,
      groupLabel: activeGroupLabel.value || '当前组',
      async onConfirm() {
        creatingOneToOneSchedule.value = true
        try {
          const res = await createOneToOneSchedulesApi({
            oneToOneId: String(oneToOneRecordId.value),
            teacherId: String(record.teacherId),
            schedules: [{
              lessonDate: record.date,
              startTime: column.startTime,
              endTime: column.endTime,
            }],
          })
          if (res.code !== 200)
            throw new Error(res.message || '创建1对1日程失败')

          messageService.success(`已为 ${studentInfo.studentName} 创建 ${month}月${day}日 第${lessonIndex}节课`)
          emitter.emit(EVENTS.REFRESH_DATA)
        }
        catch (error) {
          console.error('create one-to-one schedule failed', error)
          messageService.error(error?.response?.data?.message || error?.message || '创建1对1日程失败')
          await loadTimetableMatrix()
        }
        finally {
          creatingOneToOneSchedule.value = false
        }
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

    void confirmScheduleWithOptionalSkip({
      modeLabel: '班课',
      modeColor: '#13c2c2',
      targetLabel: '排课班级',
      targetValue: classInfo.name,
      courseName: classInfo.courseName,
      dateLabel: `${month}月${day}日 ${formatWeek(record.date)} 第${lessonIndex}节`,
      timeLabel: `${column.startTime}-${column.endTime}`,
      teacherName: record.name,
      groupLabel: activeGroupLabel.value || '当前组',
      onConfirm() {
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
          conflictReason: null,
          serverConflict: false,
          serverConflictReason: null,
        })

        console.log('更新课程信息完成', targetLesson)

        // 重新检查班课交叉时间冲突
        checkClassCrossTimeConflicts(classInfo)
      },
    })
  }
}

function getLessonIndex(startTime) {
  const slots = activePeriodSlots.value
  const i = slots.findIndex(s => s.start === startTime)
  return i >= 0 ? i + 1 : ''
}

// 添加监听，当模式切换时清空之前的选择
watch(currentModel, (newValue) => {
  console.log('切换模式', newValue)

  oneToOneAvailabilitySeq += 1
  oneToOneAvailabilityLoading.value = false
  resetEmptyLessonConflicts()

  if (newValue === '1') {
    // 切换到1v1模式，清空班级选择
    classId.value = null
    className.value = null
  }
  else {
    oneToOneRecordId.value = undefined
    courseId.value = null
    courseName.value = null
  }
})
</script>

<template>
  <div ref="timetableRootRef">
    <div class="filter-wrap bg-white pl-3 pr-3 rounded-4 rounded-lt-0 rounded-rt-0">
      <all-filter :display-array="displayArray" :is-show-search-stu-phonefilter="true" />
    </div>
    <div class="time-template mt2 bg-white py3 px5 rounded-4 rounded-lb-0 rounded-rb-0">
      <div class="top-filter st-top-filter-bar flex flex-nowrap items-center gap-1 overflow-x-auto">
        <div class="shrink-0">
          <a-radio-group v-model:value="currentModel" button-style="solid">
            <a-radio-button value="1">
              1v1
            </a-radio-button>
            <a-radio-button value="2">
              班课
            </a-radio-button>
          </a-radio-group>
        </div>
        <div class="shrink-0">
          <div v-if="currentModel === '1'" class="flex items-center shrink-0 gap-1">
            <span class="whitespace-nowrap w-71px text-right">选择1v1：</span>
            <a-select
              v-model:value="oneToOneRecordId"
              allow-clear
              show-search
              :loading="oneToOneListLoading"
              :filter-option="filterOneToOneOption"
              placeholder="搜索/选择"
              class="st-top-1v1-select"
              option-label-prop="label"
              @change="handle1v1"
            >
              <a-select-option
                v-for="item in oneToOneData"
                :key="item.id"
                :value="item.id"
                :label="item.name"
              >
                <div>{{ item.name }}</div>
              </a-select-option>
            </a-select>
          </div>
          <div v-if="currentModel === '2'" class="flex items-center">
            <!-- 写一个 select下拉选择框，使用 班级数据  -->
            <span class="w-75px">选择班级：</span>
            <a-select
              v-model:value="classId"
              allow-clear
              placeholder="请搜索/选择班级"
              class="st-top-class-select"
              option-label-prop="label"
              @change="handleClass"
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
        <div class="time-selector flex items-center shrink-0 st-time-selector--after-filters">
          <a-select
            v-model:value="currentTime"
            :options="timeViewOptions"
            class="st-time-view-select"
          />
          <div
            class="text-#0061ff font-800 text-5 flex items-center shrink-0 st-date-nav"
            :class="
              currentTime === 'day'
                ? 'st-date-nav--day'
                : currentTime === 'week'
                  ? 'st-date-nav--week'
                  : 'st-date-nav--month'
            "
          >
            <a-popover trigger="hover">
              <template #content>
                {{ currentTime === 'day' ? '前一天' : currentTime === 'week' ? '上一周' : '上个月' }}
              </template>
              <span
                class="cursor-pointer text-3 text-#888 flex w6 h6 bg-#eee rounded-10 flex-center font-500 shrink-0 hover-text-#06f hover-bg-#e6f0ff"
                @click="handlePrev"
              >
                <LeftOutlined />
              </span>
            </a-popover>
            <span class="mx-1 min-w-0 flex-1 st-date-nav__mid">
              <div class="relative cursor-pointer whitespace-nowrap text-center st-date-nav__text">
                {{ formatDateRange(currentWeek) }}
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
                class="cursor-pointer text-3 text-#888 flex w6 h6 bg-#eee rounded-10 flex-center font-500 shrink-0 hover-text-#06f hover-bg-#e6f0ff"
                @click="handleNext"
              >
                <RightOutlined />
              </span>
            </a-popover>
          </div>
          <a-button size="small" class="shrink-0 st-this-week-btn" @click="handleThisWeek">
            本周
          </a-button>
        </div>
        <div class="ml-auto flex shrink-0 items-center gap-2">
          <!-- 添加组别选择 -->
          <a-radio-group v-model:value="currentGroup" button-style="solid">
            <a-radio-button v-for="opt in groupOptions" :key="opt.key" :value="opt.key">
              {{ opt.label }}
            </a-radio-button>
          </a-radio-group>
          <a-space>
            <CreateSchedulePopover />
            <a-button>导出课表</a-button>
          </a-space>
        </div>
      </div>
    </div>
    <a-spin :spinning="timetableLoading || oneToOneAvailabilityLoading || creatingOneToOneSchedule">
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
            <div
              v-if="text.studentId"
              :data-schedule-cell-key="buildAvailabilitySlotKey(record.teacherId, record.date, column.startTime, column.endTime)"
              class="st-schedule-cell flex flex-col bg-#4e6dff1f h-11 rounded-1 text-3 text-#fff"
              :class="{ 'st-schedule-cell--focused': focusedScheduleCellKey === buildAvailabilitySlotKey(record.teacherId, record.date, column.startTime, column.endTime) }"
            >
              <!-- 方格头部时间 -->
              <!-- 班课 -->
              <div class="pl1 bg-#06f rounded-1 rounded-lb-0 rounded-rb-0 flex relative h-5">
                {{ column.startTime }}-{{ column.endTime }}
                <!-- 标记 -->
                <span
                  class="absolute right-0 pl-2 pr-1  h-4 bg-#00000080 text-#fff text-2.5 font-500 rounded-rt-1 rounded-lb-2"
                >
                  <span v-if="text.courseType === 1">1v1</span>
                  <span v-if="text.courseType === 2">班课(<span>{{ text.isMain ? '主教' : '辅教' }}</span>)</span>
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
              ]" @click="text.conflict ? handleConflictClick(text, column, record) : handleScheduleClick(text, column, record)"
            >
              {{ text.conflict ? '时间冲突(不可排)' : '空闲时段(可排)' }}
            </div>
          </template>
          <template v-if="column.key === 'date'">
            <div class="text-3.5 ">
              {{ formatWeek(text) }}
            </div>
            <div class="text-3 font-500 line-height-3 text-#666">
              {{ formatDate(text) }}
            </div>
          </template>
          <template v-if="column.key === 'name'">
            <div>{{ text }}</div>
            <div class="text-3 text-#666 leading-snug">
              {{ teacherLessonCountLabel(record.teacherId) }}
            </div>
          </template>
        </template>
      </a-table>
    </a-spin>

    <a-modal
      v-model:open="conflictDetailModalOpen"
      title="冲突详情"
      :footer="null"
      width="760px"
      centered
    >
      <div class="st-conflict-modal">
        <div class="st-conflict-summary">
          {{ conflictDetailState.summary }}
        </div>

        <div v-if="conflictDetailState.attempted" class="st-conflict-attempt">
          <div class="st-conflict-section-title">
            你正在选择的空位
          </div>
          <div class="st-conflict-attempt__card">
            <div class="st-conflict-attempt__headline">
              {{ conflictDetailState.attempted.date }} {{ conflictDetailState.attempted.week }}
              第{{ conflictDetailState.attempted.lessonIndex }}节
            </div>
            <div class="st-conflict-attempt__meta">
              {{ conflictDetailState.attempted.timeText }} · {{ conflictDetailState.attempted.teacherName }} · {{ conflictDetailState.attempted.groupLabel }}
            </div>
          </div>
        </div>

        <div class="st-conflict-section-title">
          冲突课程
        </div>
        <div class="st-conflict-list">
          <div v-for="item in conflictDetailState.items" :key="item.key" class="st-conflict-item">
            <div class="st-conflict-item__main">
              <div class="st-conflict-item__headline">
                <span>{{ item.name }}</span>
                <a-tag color="blue" :bordered="false">
                  {{ item.classTypeText }}
                </a-tag>
                <a-tag color="orange" :bordered="false">
                  {{ item.groupLabel }}
                </a-tag>
              </div>
              <div class="st-conflict-item__meta">
                {{ item.date }} {{ item.week }} · {{ item.timeText }}
              </div>
              <div class="st-conflict-item__meta">
                教师：{{ item.teacherName }}｜学员：{{ item.studentText }}
              </div>
              <div class="st-conflict-item__meta">
                冲突原因：{{ (item.conflictTypes || []).map(type => `${type}冲突`).join('、') || '时间冲突' }}
              </div>
            </div>
            <div class="st-conflict-item__side">
              <a-button type="primary" ghost :disabled="!item.jumpCellKey" @click="jumpToConflictSchedule(item)">
                定位到课程
              </a-button>
            </div>
          </div>
        </div>
      </div>
    </a-modal>
  </div>
</template>

<style lang="less" scoped>
/* 与班课下拉同量级宽度，避免顶栏把日期区挤换行 */
.st-top-1v1-select {
  width: 180px;
  max-width: 180px;
}

.st-top-class-select {
  width: 180px;
  max-width: 180px;
}

.st-time-view-select {
  width: 112px;
  min-width: 112px;
  flex-shrink: 0;
}

/* 左箭头 + 日期 + 右箭头 整体固定宽度，避免仅中间字数变化时整块左右滑 */
.st-date-nav {
  box-sizing: border-box;
}
.st-date-nav--day {
  width: 300px;
  min-width: 300px;
  max-width: 300px;
}
.st-date-nav--week {
  width: 300px;
  min-width: 300px;
  max-width: 300px;
}
.st-date-nav--month {
  width: 180px;
  min-width: 180px;
  max-width: 180px;
}
.st-date-nav__mid {
  overflow: hidden;
}
.st-date-nav__text {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* 单行不换行；极窄时出现横向滚动条，避免「创建日程」掉到第二行 */
.st-top-filter-bar {
  scrollbar-width: thin;
  -webkit-overflow-scrolling: touch;
}

.time-selector {
  font-family: DINAlternate-Bold, DINAlternate;
  gap: 6px;

  .ant-radio-button-wrapper {
    padding: 0 16px;
  }
}

/* 一对一/班课筛选与「周视图」之间略增间距 */
.st-time-selector--after-filters {
  margin-left: 8px;
}

.st-this-week-btn {
  padding: 0 10px;
  height: 28px;
  line-height: 26px;
  border-radius: 8px;
}

:deep(td.ant-table-cell.ant-table-cell-row-hover) {
  background-color: rgb(231, 236, 255) !important;
}

:deep(td.ant-table-cell) {
  padding: 4px !important;
}

.st-schedule-cell {
  position: relative;
  transition: box-shadow 0.25s ease, transform 0.25s ease;
}

.st-schedule-cell--focused {
  animation: st-schedule-cell-flash 0.5s ease-in-out 6;
  box-shadow:
    0 0 0 3px rgba(255, 77, 79, 0.98),
    0 0 0 8px rgba(255, 77, 79, 0.26),
    0 0 20px rgba(255, 77, 79, 0.5);
  transform: scale(1.015);
  z-index: 2;
}

@keyframes st-schedule-cell-flash {
  0%, 100% {
    box-shadow:
      0 0 0 2px rgba(255, 77, 79, 0.9),
      0 0 0 5px rgba(255, 77, 79, 0.14),
      0 0 10px rgba(255, 77, 79, 0.22);
    transform: scale(1.005);
  }

  50% {
    box-shadow:
      0 0 0 4px rgba(255, 77, 79, 1),
      0 0 0 10px rgba(255, 77, 79, 0.34),
      0 0 26px rgba(255, 77, 79, 0.62);
    transform: scale(1.022);
  }
}

.st-conflict-modal {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.st-conflict-summary {
  padding: 14px 16px;
  border-radius: 12px;
  background: #fff7e6;
  color: #ad6800;
  font-size: 14px;
  font-weight: 600;
  line-height: 1.7;
}

.st-conflict-section-title {
  color: #1f2329;
  font-size: 15px;
  font-weight: 700;
}

.st-conflict-attempt__card,
.st-conflict-item {
  border: 1px solid #edf2f7;
  border-radius: 14px;
  background: #fff;
}

.st-conflict-attempt__card {
  padding: 14px 16px;
}

.st-conflict-attempt__headline,
.st-conflict-item__headline {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  color: #1f2329;
  font-size: 15px;
  font-weight: 700;
}

.st-conflict-attempt__meta,
.st-conflict-item__meta {
  margin-top: 6px;
  color: #4b5563;
  font-size: 13px;
  line-height: 1.7;
}

.st-conflict-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.st-conflict-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 16px;
}

.st-conflict-item__main {
  min-width: 0;
  flex: 1;
}

.st-conflict-item__side {
  flex-shrink: 0;
}
</style>
