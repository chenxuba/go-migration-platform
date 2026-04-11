<script setup>
import { CloseOutlined, DownOutlined, ExclamationCircleFilled, ExclamationCircleOutlined } from '@ant-design/icons-vue'
import { Modal } from 'ant-design-vue'
import dayjs from 'dayjs'
import EditClassInfoModal from './edit-class-info-modal.vue'
import {
  batchEstimateRollCallSufficientTuitionAccountApi,
  checkRollCallTeachingRecordByTeacherAndTimeApi,
  confirmRollCallApi,
  getRollCallClassTimetableApi,
  getRollCallStudentLeaveCountApi,
  getRollCallStudentTuitionAccountsApi,
  getRollCallStudentTuitionExtraInfoApi,
  getRollCallTeachingRecordStudentListApi,
} from '@/api/edu-center/roll-call'
import { removeTeachingScheduleStudentCurrentApi } from '@/api/edu-center/teaching-schedule'
import { useStudentListRefresh } from '@/composables/useStudentListRefresh'
import { useStudentStore } from '@/stores/student'
import emitter, { EVENTS } from '@/utils/eventBus'
import messageService from '@/utils/messageService'

const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
  scheduleId: {
    type: String,
    default: '',
  },
  lessonDay: {
    type: String,
    default: '',
  },
})
const emit = defineEmits(['update:open', 'updated'])
const openDrawer = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})
const currentScheduleId = computed(() => String(props.scheduleId || '').trim())
const currentLessonDay = computed(() => {
  const text = String(props.lessonDay || '').trim()
  if (!text)
    return ''
  if (text.includes('T'))
    return text
  return `${text}T00:00:00`
})
const defaultStudentAvatar = 'https://cdn.schoolpal.cn/schoolpal/next-erp/avator_male.png'
const studentStore = useStudentStore()
const openStudentDrawer = ref(false)
const switchAccountModalOpen = ref(false)
const switchAccountLoading = ref(false)
const switchAccountOptions = ref([])
const switchAccountSelectedId = ref()
const switchAccountRecord = ref(null)
const switchedAccountOverrideMap = ref(new Map())
const submittingRollCall = ref(false)
const loading = ref(false)
const classTimetableDetail = ref(null)
const teachingRecordResult = ref(null)
const data = ref([])
const rollCallChanged = ref(false)
let loadSeq = 0
// 定义列
const columns = ref(
  [
    {
      title: '',
      dataIndex: 'index',
      fixed: 'left',
      width: 30,
      key: 'index',
    },
    {
      title: '学员/扣费课程账户',
      dataIndex: 'studentAccount',
      fixed: 'left',
      width: 220,
      key: 'studentAccount',
    },
    {
      title: '到课',
      dataIndex: 'attended',
      width: 95,
      key: 'attended',
    },
    {
      title: '旷课',
      dataIndex: 'absent',
      width: 95,
      key: 'absent',
    },
    {
      title: '请假',
      dataIndex: 'leave',
      width: 95,
      key: 'leave',
    },
    {
      title: '未记录',
      dataIndex: 'unrecorded',
      width: 130,
      key: 'unrecorded',
    },
    {
      title: '课消方式',
      dataIndex: 'consumptionMethod',
      width: 120,
      key: 'consumptionMethod',
    },
    {
      title: '上课点名数量',
      dataIndex: 'attendanceCount',
      width: 150,
      key: 'attendanceCount',
    },
    {
      title: '对内备注',
      dataIndex: 'internalNote',
      width: 150,
      key: 'internalNote',
    },
    {
      title: '对外备注',
      dataIndex: 'externalNote',
      width: 150,
      key: 'externalNote',
    },
    {
      title: '操作',
      dataIndex: 'action',
      fixed: 'right',
      width: 80,
      key: 'action',
    },
  ],
)
// 计算表格总宽度
const totalWidth = computed(() =>
  columns.value.reduce((acc, col) => acc + (col.width || 0), 0),
)
const userName = ref('')
// 修改为独立的状态控制
const headerStatus = ref('')
// 根据表头状态设置每个学生的状态
function setAllStudentStatus(status) {
  data.value.forEach((item) => {
    // 重置所有状态
    item.attended = false
    item.absent = false
    item.leave = false
    item.unrecorded = false
    // 设置选中的状态
    if (status) {
      item[status] = true
    }
  })
}
function syncHeaderStatus() {
  const statuses = ['attended', 'absent', 'leave', 'unrecorded']
  const matchedStatus = statuses.find(status => data.value.length > 0 && data.value.every(item => item[status]))
  headerStatus.value = matchedStatus || ''
}
// 计算属性监听表头状态变化
watch(() => headerStatus.value, (newStatus) => {
  // 当表头状态变化时，更新所有学生的状态
  if (newStatus === '') {
    return
  }
  setAllStudentStatus(newStatus)
})
// 处理表头批量操作
function handleHeaderStatusChange(status) {
  if (headerStatus.value === status) {
    // 如果点击当前选中的状态，则取消选择
    headerStatus.value = ''
    setAllStudentStatus('')
  }
  else {
    // 否则切换到新状态
    headerStatus.value = status
    setAllStudentStatus(status)
  }
}
// 处理单个学生状态变更
function handleStudentStatusChange(record, status) {
  // 判断是否是取消选择当前状态
  if (record[status]) {
    // 如果当前状态已经选中，则取消选择
    record[status] = false
  }
  else {
    // 重置该学生的所有状态
    record.attended = false
    record.absent = false
    record.leave = false
    record.unrecorded = false

    // 设置新状态
    record[status] = true
  }

  // 检查表头状态，只有全部学生选中同一状态时才改变表头状态
  const allSelected = data.value.every(item => item[status])
  const anySelected = data.value.some(item => item[status])

  if (allSelected) {
    headerStatus.value = status
  }
  else if (!anySelected && headerStatus.value === status) {
    // 如果没有学生选中此状态，且表头状态为此状态，则清除表头状态
    headerStatus.value = ''
  }
  else if (!anySelected) {
    headerStatus.value = ''
  }
  else {
    headerStatus.value = ''
  }
}
// 计算出席统计
const attendanceStats = computed(() => {
  return {
    attended: data.value.filter(student => student.attended).length,
    absent: data.value.filter(student => student.absent).length,
    leave: data.value.filter(student => student.leave).length,
    unrecorded: data.value.filter(student => student.unrecorded).length,
  }
})
const filteredData = computed(() => {
  const keyword = String(userName.value || '').trim()
  if (!keyword)
    return data.value
  return data.value.filter(item =>
    String(item.studentAccount || '').includes(keyword)
    || String(item.accountName || '').includes(keyword),
  )
})
const titleText = computed(() =>
  String(classTimetableDetail.value?.className || teachingRecordResult.value?.data?.sourceName || '上课点名').trim() || '上课点名',
)
const teacherBucket = computed(() => {
  const teachers = Array.isArray(classTimetableDetail.value?.teachers) ? classTimetableDetail.value.teachers : []
  const mainTeacher = teachers.find(item => Number(item.teacherDuty) === 1) || teachers[0]
  const assistantNames = teachers
    .filter(item => Number(item.teacherDuty) !== 1)
    .map(item => String(item.teacherName || '').trim())
    .filter(Boolean)
  return {
    teacherName: String(mainTeacher?.teacherName || '').trim() || '-',
    assistantText: assistantNames.length ? assistantNames.join('、') : '-',
  }
})
const classroomText = computed(() => String(classTimetableDetail.value?.addressName || '-').trim() || '-')
const showTip = computed(() => data.value.some(item => item.bindChildText === '未关注'))
const tipText = computed(() => '未关注家校平台的学员家长，无法接收课消提醒')
const timeText = computed(() => {
  const startTime = String(teachingRecordResult.value?.data?.startTime || '').trim()
  const endTime = String(teachingRecordResult.value?.data?.endTime || '').trim()
  if (!startTime || !endTime)
    return '-'
  const weekText = ['周日', '周一', '周二', '周三', '周四', '周五', '周六'][dayjs(startTime).day()] || '-'
  return `${dayjs(startTime).format('YYYY-MM-DD')}(${weekText}) ${dayjs(startTime).format('HH:mm')} ~ ${dayjs(endTime).format('HH:mm')}`
})
const durationText = computed(() => {
  const startTime = String(teachingRecordResult.value?.data?.startTime || '').trim()
  const endTime = String(teachingRecordResult.value?.data?.endTime || '').trim()
  if (!startTime || !endTime)
    return '-'
  const minutes = dayjs(endTime).diff(dayjs(startTime), 'minute')
  return minutes > 0 ? `${minutes}分钟` : '-'
})
const teacherClassTimeText = computed(() => {
  const value = Number(teachingRecordResult.value?.data?.teacherClassTime ?? classTimetableDetail.value?.defaultTeacherClassTime ?? 0)
  return `教师记录 ${Number.isInteger(value) ? value : value.toFixed(2).replace(/\.?0+$/, '')} 课时`
})
function getConsumptionMethodText(mode, studentType) {
  if (String(studentType || '') === '3')
    return '按课时'
  if (Number(mode) === 1)
    return '按课时'
  if (Number(mode) === 2)
    return '按时间'
  if (Number(mode) === 3)
    return '按金额'
  return '-'
}
function formatRemainingText(mode, quantity, paidRemaining) {
  const chargeMode = Number(mode || 0)
  if (chargeMode === 1)
    return `剩余课时：${Number(quantity || 0)}`
  if (chargeMode === 2)
    return `剩余天数：${Number(quantity || 0)}`
  if (chargeMode === 3)
    return `剩余金额：${Number(paidRemaining || 0)}`
  return ''
}
function normalizeAccountChargingMode(mode) {
  const parsed = Number(mode || 0)
  if (parsed === 4)
    return 3
  return parsed
}
function effectiveAccountChargingMode(acc) {
  const mode = normalizeAccountChargingMode(acc?.lessonChargingMode)
  if (mode > 0)
    return mode
  const totalQty = Number(acc?.totalQuantity || 0) + Number(acc?.totalFreeQuantity || 0)
  const remainQty = Number(acc?.quantity || 0) + Number(acc?.freeQuantity || 0)
  if ((totalQty > 0 || remainQty > 0) && acc?.enableExpireTime)
    return 2
  if (totalQty > 0 || remainQty > 0)
    return 1
  if (Number(acc?.totalTuition || 0) > 0 || Number(acc?.tuition || 0) > 0)
    return 3
  return 0
}
function isZeroDateTime(value) {
  const text = String(value || '').trim()
  return !text || text.startsWith('0001-01-01')
}
function getChargingModeText(mode) {
  if (Number(mode) === 1)
    return '课时'
  if (Number(mode) === 2)
    return '时段'
  if (Number(mode) === 3)
    return '金额'
  return '-'
}
function accountDeductTeachMethodText(lessonType) {
  const type = Number(lessonType || 0)
  if (type === 1)
    return '班级授课'
  if (type === 2)
    return '1对1授课'
  return '班级授课'
}
function switchAccountRemainText(acc) {
  const mode = effectiveAccountChargingMode(acc)
  const remainQuantity = Number(acc?.quantity || 0) + Number(acc?.freeQuantity || 0)
  if (mode === 1)
    return `剩余课时：${remainQuantity}`
  if (mode === 2)
    return `剩余天数：${remainQuantity}`
  if (mode === 3)
    return `剩余金额：${Number(acc?.tuition || 0)}`
  return '-'
}
function switchAccountExpireText(acc) {
  const startText = isZeroDateTime(acc?.startTime) ? '' : dayjs(acc.startTime).format('YYYY-MM-DD')
  const expireText = isZeroDateTime(acc?.expireTime) ? '' : dayjs(acc.expireTime).format('YYYY-MM-DD')
  if (startText && expireText)
    return `${startText} ~ ${expireText}`
  if (expireText)
    return expireText
  return '不限制'
}
function parseNumber(value) {
  const parsed = Number(value)
  return Number.isFinite(parsed) ? parsed : 0
}
function isRemainingInsufficient(record) {
  if (!record || record.type === '3' || record.unrecorded || !record.recordAttendance)
    return false
  if (String(record.consumptionMethod || '') !== '1')
    return false
  return parseNumber(record.attendanceCount) > parseNumber(record.remainingQuantity)
}
function teachingSourceTypeToTagType(sourceType) {
  if (Number(sourceType) === 4)
    return '3'
  if (Number(sourceType) === 2)
    return '2'
  if (Number(sourceType) === 3)
    return '4'
  return '1'
}
function defaultTeachingStatusToKey(value) {
  if (Number(value) === 2)
    return 'absent'
  if (Number(value) === 3)
    return 'leave'
  if (Number(value) === 4)
    return 'unrecorded'
  return 'attended'
}
function applyStudentState(item) {
  item.attended = false
  item.absent = false
  item.leave = false
  item.unrecorded = false
  const status = String(item.defaultAttendanceStatus || 'attended')
  if (status === 'absent')
    item.absent = true
  else if (status === 'leave')
    item.leave = true
  else if (status === 'unrecorded')
    item.unrecorded = true
  else
    item.attended = true
}
function mapStudentRow(item, leaveCountMap, tuitionExtraMap) {
  const extra = tuitionExtraMap.get(String(item.studentId || '')) || {}
  const leaveCount = Number(leaveCountMap.get(String(item.studentId || '')) || 0)
  const studentType = teachingSourceTypeToTagType(item.sourceType)
  const row = {
    id: String(item.studentId || ''),
    type: studentType,
    studentAccount: String(item.studentName || '-'),
    avatarUrl: String(item.avatar || defaultStudentAvatar),
    bindChildText: item.isBindChild ? '已关注' : '未关注',
    accountName: String(extra.bestMatchProductName || ''),
    remainingText: formatRemainingText(item.chargingMode, item.quantity, item.paidRemaining),
    remainingQuantity: Number(item.quantity || 0),
    paidRemaining: Number(item.paidRemaining || 0),
    leaveCountText: `已请假：${leaveCount}次`,
    consumptionMethod: String(item.chargingMode || ''),
    consumptionMethodText: getConsumptionMethodText(item.chargingMode, studentType),
    tuitionAccountId: String(item.tuitionAccountId || ''),
    sourceType: Number(item.sourceType || 0),
    rawChargingMode: Number(item.chargingMode || 0),
    recordAttendance: Number(item.chargingMode || 0) === 1,
    attendanceCount: Number(classTimetableDetail.value?.defaultStudentClassTime || 1),
    internalNote: '',
    externalNote: '',
    defaultAttendanceStatus: defaultTeachingStatusToKey(item.defaultStudentTeachingStatus),
    canSwitchAccount: Boolean(extra.mutilTuition),
    canRemove: true,
    attended: false,
    absent: false,
    leave: false,
    unrecorded: false,
  }
  applyStudentState(row)
  return row
}
function pickSwitchAccountSelectedId(list, record) {
  const currentId = String(record?.tuitionAccountId || '')
  const direct = list.find(acc => String(acc?.id || '') === currentId)
  if (direct?.id)
    return String(direct.id)
  return String(list[0]?.id || '')
}
function syncRecordTuitionAccount(record, acc) {
  const mode = effectiveAccountChargingMode(acc)
  const remainQuantity = Number(acc?.quantity || 0) + Number(acc?.freeQuantity || 0)
  record.tuitionAccountId = String(acc?.id || '')
  record.accountName = String(acc?.productName || acc?.lessonName || '')
  record.remainingQuantity = remainQuantity
  record.paidRemaining = Number(acc?.tuition || 0)
  record.remainingText = formatRemainingText(mode, remainQuantity, acc?.tuition)
  record.consumptionMethod = String(mode || '')
  record.consumptionMethodText = getConsumptionMethodText(mode, record.type)
  record.recordAttendance = mode === 1
}
function studentAccountOverrideKey(studentId) {
  return [
    String(currentScheduleId.value || '').trim(),
    String(currentLessonDay.value || '').trim(),
    String(studentId || '').trim(),
  ].join('#')
}
function saveStudentAccountOverride(studentId, acc) {
  const key = studentAccountOverrideKey(studentId)
  if (!key || !acc)
    return
  const nextMap = new Map(switchedAccountOverrideMap.value)
  nextMap.set(key, { ...acc })
  switchedAccountOverrideMap.value = nextMap
}
function getStudentAccountOverride(studentId) {
  const key = studentAccountOverrideKey(studentId)
  if (!key)
    return null
  return switchedAccountOverrideMap.value.get(key) || null
}
function applySwitchedAccountOverrides(rows) {
  rows.forEach((row) => {
    const override = getStudentAccountOverride(row?.id)
    if (override)
      syncRecordTuitionAccount(row, override)
  })
  return rows
}
function clearSwitchedAccountOverrides() {
  switchedAccountOverrideMap.value = new Map()
}
function shouldSkipManualErrorMessage(error) {
  return Number(error?.response?.status || 0) === 400
}
async function loadDetail() {
  if (!openDrawer.value || !currentScheduleId.value) {
    classTimetableDetail.value = null
    teachingRecordResult.value = null
    data.value = []
    headerStatus.value = ''
    closeSwitchAccountModal()
    clearSwitchedAccountOverrides()
    return
  }

  const seq = ++loadSeq
  loading.value = true
  try {
    const classTimetableRes = await getRollCallClassTimetableApi({
      id: currentScheduleId.value,
      lessonDay: currentLessonDay.value || undefined,
    })
    if (seq !== loadSeq)
      return
    if (classTimetableRes.code !== 200 || !classTimetableRes.result?.detail)
      throw new Error(classTimetableRes.message || '加载点名课表失败')

    classTimetableDetail.value = classTimetableRes.result.detail
    const lessonDay = String(classTimetableRes.result.detail.lessonDays?.[0]?.lessonDay || currentLessonDay.value || '')
    const classId = String(classTimetableRes.result.detail.classId || '')
    const lessonId = String(classTimetableRes.result.detail.lessonId || '')

    const recordRes = await getRollCallTeachingRecordStudentListApi({
      timetableSourceId: currentScheduleId.value,
      timetableSourceType: 1,
      classId,
      lessonId,
      one2OneId: '0',
      startDate: String(classTimetableRes.result.detail.startDate || '0001-01-01T00:00:00'),
      endDate: String(classTimetableRes.result.detail.endDate || '0001-01-01T00:00:00'),
      lessonDay,
    })
    if (seq !== loadSeq)
      return
    if (recordRes.code !== 200 || !recordRes.result)
      throw new Error(recordRes.message || '加载点名学员失败')

    teachingRecordResult.value = recordRes.result
    const studentIds = Array.isArray(recordRes.result.students) ? recordRes.result.students.map(item => String(item.studentId || '')).filter(Boolean) : []
    const [leaveCountRes, tuitionExtraRes] = await Promise.all([
      getRollCallStudentLeaveCountApi({
        studentIds,
        lessonId,
      }),
      getRollCallStudentTuitionExtraInfoApi({
        studentIds,
        lessonId,
      }),
    ])
    if (seq !== loadSeq)
      return
    if (leaveCountRes.code !== 200)
      throw new Error(leaveCountRes.message || '加载请假次数失败')
    if (tuitionExtraRes.code !== 200)
      throw new Error(tuitionExtraRes.message || '加载扣费补充信息失败')

    const leaveCountMap = new Map((Array.isArray(leaveCountRes.result) ? leaveCountRes.result : []).map(item => [String(item.studentId || ''), Number(item.leaveCount || 0)]))
    const tuitionExtraMap = new Map((Array.isArray(tuitionExtraRes.result) ? tuitionExtraRes.result : []).map(item => [String(item.studentId || ''), item]))
    data.value = applySwitchedAccountOverrides((Array.isArray(recordRes.result.students) ? recordRes.result.students : []).map(item =>
      mapStudentRow(item, leaveCountMap, tuitionExtraMap),
    ))
    syncHeaderStatus()
  }
  catch (error) {
    if (seq !== loadSeq)
      return
    classTimetableDetail.value = null
    teachingRecordResult.value = null
    data.value = []
    headerStatus.value = ''
    messageService.error(error?.response?.data?.message || error?.message || '加载点名详情失败')
  }
  finally {
    if (seq === loadSeq)
      loading.value = false
  }
}
// 编辑上课信息
const editClassInfoModal = ref(false)
function handleEditClassInfo() {
  editClassInfoModal.value = true
}

// 添加学员modal
const addStudentModal = ref(false)
const addStudentModalTitle = ref('')
const addStudentType = ref(4)
// 添加学员
function handleAddStudent({ key }) {
  if (key === '1') {
    messageService.info('补课学员功能暂未开发')
    return
  }
  else if (key === '2') {
    addStudentModalTitle.value = '添加临时学员'
    addStudentType.value = 2
  }
  else if (key === '3') {
    addStudentModalTitle.value = '添加试听学员'
    addStudentType.value = 3
  }
  addStudentModal.value = true
}
async function handleAddStudentSuccess() {
  rollCallChanged.value = true
  emit('updated')
  await loadDetail()
}
// 批量编辑modal
const batchEditModal = ref(false)
function handleBatchEdit() {
  batchEditModal.value = true
}
function isBatchEditableStudent(record) {
  return String(record?.type || '') !== '3' && String(record?.type || '') !== '4'
}
function matchesBatchEditRange(record, editRange) {
  if (!record || !isBatchEditableStudent(record))
    return false
  if (String(editRange) === '2')
    return Boolean(record.attended)
  if (String(editRange) === '3')
    return Boolean(record.leave)
  if (String(editRange) === '4')
    return Boolean(record.absent)
  return Boolean(record.attended || record.leave || record.absent)
}
function handleBatchEditSubmit(payload) {
  const classNumber = Number(payload?.classNumber || 0)
  const editRange = String(payload?.editRange || '1')
  const targetRows = data.value.filter(record => matchesBatchEditRange(record, editRange))
  if (targetRows.length === 0) {
    messageService.warning('当前没有符合条件的学员')
    return
  }
  targetRows.forEach((record) => {
    record.attendanceCount = classNumber
  })
  messageService.success(`已批量修改${targetRows.length}位学员的上课点名数量`)
}
function getRollCallRecordStatus(record) {
  if (record?.absent)
    return 2
  if (record?.leave)
    return 3
  if (record?.unrecorded)
    return 4
  return 1
}
function buildRollCallStudentQuantity(record) {
  if (!record || record.unrecorded)
    return 0
  if (!record.recordAttendance)
    return 0
  return Math.max(parseNumber(record.attendanceCount), 0)
}
function buildRollCallEstimatePayload() {
  return data.value
    .filter(record => Number(record?.consumptionMethod || 0) === 1 && buildRollCallStudentQuantity(record) > 0)
    .map(record => ({
      quantity: buildRollCallStudentQuantity(record),
      tuitionAccountId: String(record?.tuitionAccountId || ''),
      studentName: String(record?.studentAccount || ''),
    }))
}
function buildRollCallConfirmPayload() {
  const meta = teachingRecordResult.value?.data || {}
  const detail = classTimetableDetail.value || {}
  const teacherList = Array.isArray(teachingRecordResult.value?.teachers) ? teachingRecordResult.value.teachers : []
  return {
    sourceName: String(meta.sourceName || detail.className || ''),
    teachingContent: '',
    teachingContentImages: [],
    timetableSourceType: Number(meta.timetableSourceType || 0),
    timetableSourceId: String(meta.timetableSourceId || currentScheduleId.value || ''),
    sourceId: String(meta.sourceId || detail.classId || ''),
    sourceType: Number(meta.sourceType || 0),
    lessonId: String(meta.lessonId || detail.lessonId || ''),
    startTime: String(meta.startTime || ''),
    endTime: String(meta.endTime || ''),
    teacherClassTime: Number(meta.teacherClassTime || detail.defaultTeacherClassTime || 0),
    studentShouldDeduct: Number(detail.defaultStudentClassTime || 1),
    teacherList: teacherList.map(item => ({
      teacherId: String(item?.teacherId || ''),
      type: Number(item?.type || 0),
    })),
    studentList: data.value.map((record) => {
      const quantity = buildRollCallStudentQuantity(record)
      return {
        studentShouldDeduct: quantity,
        studentName: String(record?.studentAccount || ''),
        studentId: String(record?.id || ''),
        tuitionAccountId: String(record?.tuitionAccountId || '0'),
        absentTeachingRecordId: '0',
        status: getRollCallRecordStatus(record),
        sourceType: Number(record?.sourceType || 0),
        remark: String(record?.internalNote || ''),
        externalRemark: String(record?.externalNote || ''),
        skuMode: Number(record?.consumptionMethod || record?.rawChargingMode || 0),
        amount: 0,
        quantity,
      }
    }),
    subjectId: '0',
    classRoomId: String(meta.classroomId || detail.addressId || '0'),
  }
}
function showRollCallArrearWarning(names) {
  if (!Array.isArray(names) || names.length === 0)
    return
  Modal.warning({
    title: '超记提醒',
    content: `以下学员剩余课时不足，已按超记处理：${names.join('、')}`,
    okText: '我知道了',
  })
}
async function handleConfirmRollCall() {
  if (submittingRollCall.value)
    return
  if (!data.value.length) {
    messageService.warning('当前没有可提交的点名学员')
    return
  }
  const payload = buildRollCallConfirmPayload()
  if (!payload.startTime || !payload.endTime) {
    messageService.warning('缺少上课时间，请刷新后重试')
    return
  }
  const mainTeacher = payload.teacherList.find(item => Number(item?.type) === 1)
  const estimatePayload = buildRollCallEstimatePayload()
  submittingRollCall.value = true
  try {
    if (mainTeacher?.teacherId) {
      const checkRes = await checkRollCallTeachingRecordByTeacherAndTimeApi({
        startTime: payload.startTime,
        endTime: payload.endTime,
        teacherId: String(mainTeacher.teacherId || ''),
      })
      if (checkRes.code !== 200)
        throw new Error(checkRes.message || '上课教师时间冲突校验失败')
    }

    let insufficientNames = []
    if (estimatePayload.length > 0) {
      const estimateRes = await batchEstimateRollCallSufficientTuitionAccountApi({
        tuitionInfoList: estimatePayload,
      })
      if (estimateRes.code !== 200)
        throw new Error(estimateRes.message || '扣费账户剩余校验失败')
      const insufficientIdSet = new Set(
        (Array.isArray(estimateRes.result?.tuitionInfoList) ? estimateRes.result.tuitionInfoList : [])
          .filter(item => item?.isSufficient === false)
          .map(item => String(item?.tuitionAccountId || '')),
      )
      insufficientNames = estimatePayload
        .filter(item => insufficientIdSet.has(String(item?.tuitionAccountId || '')))
        .map(item => String(item?.studentName || ''))
        .filter(Boolean)
    }

    messageService.clear()
    const confirmRes = await confirmRollCallApi(payload)
    if (confirmRes.code !== 200)
      throw new Error(confirmRes.message || '点名提交失败')

    messageService.success('点名成功')
    rollCallChanged.value = true
    openDrawer.value = false
    showRollCallArrearWarning(Array.from(new Set(insufficientNames)))
  }
  catch (error) {
    if (!shouldSkipManualErrorMessage(error)) {
      messageService.error(error?.response?.data?.message || error?.message || '点名提交失败')
    }
  }
  finally {
    submittingRollCall.value = false
  }
}
async function handleSwitchAccount(record) {
  if (!record?.canSwitchAccount)
    return
  const lessonId = String(classTimetableDetail.value?.lessonId || teachingRecordResult.value?.data?.lessonId || '').trim()
  if (!lessonId) {
    messageService.error('缺少课程信息')
    return
  }
  switchAccountRecord.value = record
  switchAccountLoading.value = true
  try {
    const res = await getRollCallStudentTuitionAccountsApi({
      studentId: String(record.id || ''),
      lessonId,
    })
    if (res.code !== 200)
      throw new Error(res.message || '加载扣费课程账户失败')
    const list = Array.isArray(res.result?.list) ? res.result.list : []
    switchAccountOptions.value = list
    if (list.length === 0) {
      messageService.warning('暂无可切换的扣费课程账户')
      return
    }
    switchAccountSelectedId.value = pickSwitchAccountSelectedId(list, record)
    switchAccountModalOpen.value = true
  }
  catch (error) {
    console.error('load roll call tuition accounts failed', error)
    messageService.error(error?.response?.data?.message || error?.message || '加载扣费课程账户失败')
  }
  finally {
    switchAccountLoading.value = false
  }
}
function handleViewStudent(studentId) {
  const id = String(studentId || '').trim()
  if (!id)
    return
  studentStore.setStudentId(id)
  openStudentDrawer.value = true
}
function closeSwitchAccountModal() {
  switchAccountModalOpen.value = false
  switchAccountSelectedId.value = undefined
  switchAccountOptions.value = []
  switchAccountRecord.value = null
}
function submitSwitchAccount() {
  const currentId = String(switchAccountSelectedId.value || '').trim()
  if (!currentId) {
    messageService.warning('请选择扣费课程账户')
    return
  }
  const selectedAccount = switchAccountOptions.value.find(acc => String(acc?.id || '') === currentId)
  if (!selectedAccount || !switchAccountRecord.value) {
    messageService.warning('请选择扣费课程账户')
    return
  }
  syncRecordTuitionAccount(switchAccountRecord.value, selectedAccount)
  saveStudentAccountOverride(switchAccountRecord.value.id, selectedAccount)
  closeSwitchAccountModal()
}
function handleRemoveStudent(record) {
  const scheduleId = String(currentScheduleId.value || '').trim()
  const studentId = String(record?.id || '').trim()
  const name = String(record?.studentAccount || '').trim() || '当前学员'
  if (!scheduleId || !studentId) {
    messageService.warning('当前学员缺少移出标识，请刷新后重试')
    return
  }
  let removing = false
  Modal.confirm({
    title: '移出本节学员',
    content: `移出后仅影响本节课，不会影响班级成员和后续未开课。确认移出“${name}”吗？`,
    okText: '确认移出',
    cancelText: '取消',
    async onOk() {
      if (removing)
        return
      removing = true
      try {
        messageService.clear()
        const res = await removeTeachingScheduleStudentCurrentApi({
          scheduleId,
          studentId,
        })
        if (res.code !== 200)
          throw new Error(res.message || '移出本节失败')
        messageService.success(`已将${name}移出本节`)
        rollCallChanged.value = true
        emit('updated')
        await loadDetail()
      }
      catch (error) {
        if (!shouldSkipManualErrorMessage(error)) {
          messageService.error(error?.response?.data?.message || error?.message || '移出本节失败')
        }
        throw error
      }
      finally {
        removing = false
      }
    },
  })
}

useStudentListRefresh(() => {
  if (openDrawer.value && currentScheduleId.value) {
    loadDetail()
  }
})

watch(
  () => `${openDrawer.value}|${currentScheduleId.value}|${currentLessonDay.value}`,
  () => {
    loadDetail()
  },
  { immediate: true },
)

watch(
  () => openDrawer.value,
  (open, previous) => {
    if (!open && previous && rollCallChanged.value) {
      emit('updated')
      emitter.emit(EVENTS.REFRESH_DATA)
      rollCallChanged.value = false
    }
    if (open && !previous)
      rollCallChanged.value = false
  },
)
</script>

<template>
  <div>
    <a-drawer
      v-model:open="openDrawer" :push="{ distance: 80 }" :body-style="{ padding: '0', background: '#f7f7fd' }"
      :closable="false" width="1244px" placement="right"
    >
      <!-- 自定义头部 -->
      <template #title>
        <div class="custom-header flex justify-between h-4 flex-items-center">
          <div class="text-5">
            上课点名
          </div>
          <a-button type="text" class="close-btn" @click="openDrawer = false">
            <template #icon>
              <CloseOutlined class="text-5 close-icon" />
            </template>
          </a-button>
        </div>
      </template>
      <div v-if="showTip" class="tips bg-#e6f0ff py-12px px-16px text-#06f">
        <ExclamationCircleFilled /> {{ tipText }}
      </div>
      <div class="contenter flex flex-center bg-white px6 py3">
        <div class="avatarBox w-16 h-16 relative">
          <img
            width="64" height="64" src="https://pcsys.admin.ybc365.com/83b8fd68-2f9b-4a35-979f-1fd0ea349889.png"
            alt=""
          >
        </div>
        <div class="info flex flex-1 ml-4 flex-col">
          <div class="top flex justify-between flex-center flex-1">
            <a-space>
              <div class="name text-5 font-800">
                {{ titleText }}
              </div>
            </a-space>
          </div>
          <div class="bottom flex-1 flex flex-items-center mt-2">
            <div class="birthday flex-center">
              <span class="text-4 text-#222">{{ timeText }}</span>
              <span class="bg-#e6f0ff text-#06f text-3 px2 py1 rounded-10 ml2">{{ durationText }}</span>
            </div>
          </div>
        </div>
      </div>
      <div class="desc pt-4 bg-white px6 py3 pb0">
        <a-descriptions :column="3" size="small" :content-style="{ color: '#333' }">
          <a-descriptions-item label="上课教师">
            {{ teacherBucket.teacherName }}
          </a-descriptions-item>
          <a-descriptions-item label="上课助教">
            {{ teacherBucket.assistantText }}
          </a-descriptions-item>
          <a-descriptions-item label="上课教室">
            {{ classroomText }}
          </a-descriptions-item>
          <a-descriptions-item label="本次上课">
            {{ teacherClassTimeText }}
          </a-descriptions-item>
          <a-descriptions-item>
            <span class="text-#06f cursor-pointer" @click="handleEditClassInfo">编辑上课信息</span>
          </a-descriptions-item>
        </a-descriptions>
      </div>
      <div class="tables bg-#fff pt-16px px-24px pb-30px">
        <a-input v-model:value="userName" placeholder="搜索学员" class="h-48px rounded-12px">
          <template #prefix>
            <img
              src="https://prod-tbu-next-erp-cdn.schoolpal.cn/next-pc-static/static/12181/static/magnifying.2bcc08ab.svg"
              alt="" class="pr-6px mt--2px"
            >
          </template>
        </a-input>
        <!-- 用a-table 学员/扣费课程账户 到课 旷课 请假 未记录 课消方式 上课点名数量 对内备注 对外备注 操作 -->
        <!-- 带序号 -->
        <a-table
          :columns="columns" :data-source="filteredData" row-key="id" class="mt-12px" :pagination="false"
          :scroll="{ x: totalWidth }"
        >
          <template #headerCell="{ column }">
            <div v-if="column.dataIndex === 'studentAccount'">
              <div class="text-#333 font-800">
                {{ column.title }}
                <a-popover title="学员/扣费课程账户">
                  <template #content>
                    <div class="w-450px">
                      【扣费课程账户】当前点名所消耗的课程账户，报读相同课程且课消方式相同时，会合并计算为一个扣费课程账户，支持切换课程账户课消
                      <br>
                      【剩余数量】对应课程账户相关的剩余数量（课时/金额/天数）
                    </div>
                  </template>
                  <ExclamationCircleOutlined class="text-#06f cursor-pointer ml-4px" />
                </a-popover>
              </div>
            </div>
            <div v-if="column.dataIndex === 'attended'">
              <a-tooltip>
                <template #title>
                  {{ headerStatus === 'attended' ? '取消批量到课' : '批量到课' }}
                </template>
                <div class="text-#333 font-800">
                  <a-checkbox
                    :checked="headerStatus === 'attended'" class="status-checkbox attended-checkbox"
                    :class="{ 'active-checkbox': headerStatus === 'attended' }"
                    @click="() => handleHeaderStatusChange('attended')"
                  >
                    {{
                      column.title }}
                  </a-checkbox>
                </div>
              </a-tooltip>
            </div>
            <div v-if="column.dataIndex === 'absent'">
              <a-tooltip>
                <template #title>
                  {{ headerStatus === 'absent' ? '取消批量旷课' : '批量旷课' }}
                </template>
                <div class="text-#333 font-800">
                  <a-checkbox
                    :checked="headerStatus === 'absent'" class="status-checkbox absent-checkbox"
                    :class="{ 'active-checkbox': headerStatus === 'absent' }"
                    @click="() => handleHeaderStatusChange('absent')"
                  >
                    {{
                      column.title }}
                  </a-checkbox>
                </div>
              </a-tooltip>
            </div>
            <div v-if="column.dataIndex === 'leave'">
              <a-tooltip>
                <template #title>
                  {{ headerStatus === 'leave' ? '取消批量请假' : '批量请假' }}
                </template>
                <div class="text-#333 font-800">
                  <a-checkbox
                    :checked="headerStatus === 'leave'" class="status-checkbox leave-checkbox"
                    :class="{ 'active-checkbox': headerStatus === 'leave' }"
                    @click="() => handleHeaderStatusChange('leave')"
                  >
                    {{
                      column.title }}
                  </a-checkbox>
                </div>
              </a-tooltip>
            </div>
            <div v-if="column.dataIndex === 'unrecorded'">
              <a-popover title="字段说明">
                <template #content>
                  <div class="w-300px">
                    学员为"未记录"状态时，无法记录课时，也不会发送家长端消息提醒。
                  </div>
                </template>
                <div class="text-#333 font-800">
                  <a-checkbox
                    :checked="headerStatus === 'unrecorded'"
                    class="status-checkbox unrecorded-checkbox" :class="{ 'active-checkbox': headerStatus === 'unrecorded' }"
                    @click="() => handleHeaderStatusChange('unrecorded')"
                  >
                    {{
                      column.title }}
                    <ExclamationCircleOutlined class="cursor-pointer mr-4px" />
                  </a-checkbox>
                </div>
              </a-popover>
            </div>
            <div v-if="column.dataIndex === 'consumptionMethod'">
              <a-popover title="课消方式">
                <template #content>
                  <div class="w-300px">
                    【按时间】按天数计费 <br>
                    【按课时】按课时计费 <br>
                    【按金额】按金额计费 <br>
                    【提示】按时间、按金额计费，开启记录课时后仅作为「记录」，课时增减不产生学费变动。
                  </div>
                </template>
                <div class="text-#333 font-800">
                  {{ column.title }}
                  <ExclamationCircleOutlined class="cursor-pointer mr-4px" />
                </div>
              </a-popover>
            </div>
          </template>
          <template #bodyCell="{ column, record, index }">
            <div v-if="column.dataIndex === 'index'">
              {{ index + 1 }} {{ record.name }}
            </div>
            <div v-if="column.dataIndex === 'studentAccount'">
              <div class="flex flex-items-center text-3 cursor-pointer" @click="handleViewStudent(record.id)">
                <div class=" mr-4px">
                  <img
                    :src="record.avatarUrl || defaultStudentAvatar"
                    class="w-40px h-40px rounded-full mr-6px" alt=""
                  >
                  <span
                    v-if="record.type === '3'"
                    class=" flex bg-#fff5e6 text-#f90 w-120% justify-center ml--8px text-10px rounded-10"
                  >免费试听</span>
                  <span
                    v-if="record.type === '4'"
                    class=" flex bg-#734338 text-#fff w-120% justify-center ml--8px text-10px rounded-10"
                  >补课学员</span>
                  <span
                    v-if="record.type === '2'"
                    class=" flex bg-#888 text-#fff w-120% justify-center ml--8px text-10px rounded-10"
                  >临时学员</span>
                </div>
                <div class="text-#888">
                  <div class="text-14px text-#333 mb-2px">
                    {{ record.studentAccount }} <span
                      class="text-3 px2 py2px rounded-10 ml2px"
                      :class="record.bindChildText === '已关注' ? 'bg-#e6f0ff text-#06f' : 'bg-#f2f3f5 text-#8c8c8c'"
                    >{{ record.bindChildText }}</span>
                  </div>
                  <div v-if="record.accountName || record.canSwitchAccount">
                    {{ record.accountName }}
                    <a v-if="record.canSwitchAccount" class="text-#06f ml-4px" @click.stop="handleSwitchAccount(record)">切换</a>
                  </div>
                  <div v-if="record.remainingText">
                    {{ record.remainingText }}
                  </div>
                  <div class="text-#f90">
                    {{ record.leaveCountText }}
                  </div>
                </div>
              </div>
            </div>
            <div v-if="column.dataIndex === 'attended'">
              <div
                class="text-#333 attended-status"
                :class="{ 'active-status': record.attended, 'inactive-status': !record.attended }"
              >
                <a-checkbox
                  :checked="record.attended" class="status-checkbox attended-checkbox"
                  :class="{ 'active-checkbox': record.attended }" @click="() => handleStudentStatusChange(record, 'attended')"
                >
                  到课
                </a-checkbox>
              </div>
            </div>
            <div v-if="column.dataIndex === 'absent'">
              <div
                class="text-#333 cursor-pointer absent-status"
                :class="{ 'active-status': record.absent, 'inactive-status': !record.absent }"
              >
                <a-checkbox
                  :checked="record.absent" class="status-checkbox absent-checkbox"
                  :class="{ 'active-checkbox': record.absent }" @click="() => handleStudentStatusChange(record, 'absent')"
                >
                  旷课
                </a-checkbox>
              </div>
            </div>
            <div v-if="column.dataIndex === 'leave'">
              <div
                class="text-#333 cursor-pointer leave-status"
                :class="{ 'active-status': record.leave, 'inactive-status': !record.leave }"
              >
                <a-checkbox
                  :checked="record.leave" class="status-checkbox leave-checkbox"
                  :class="{ 'active-checkbox': record.leave }" @click="() => handleStudentStatusChange(record, 'leave')"
                >
                  请假
                </a-checkbox>
              </div>
            </div>
            <div v-if="column.dataIndex === 'unrecorded'">
              <div
                class="text-#333 cursor-pointer unrecorded-status"
                :class="{ 'active-status': record.unrecorded, 'inactive-status': !record.unrecorded }"
              >
                <a-checkbox
                  :checked="record.unrecorded" class="status-checkbox unrecorded-checkbox"
                  :class="{ 'active-checkbox': record.unrecorded }"
                  @click="() => handleStudentStatusChange(record, 'unrecorded')"
                >
                  未记录
                </a-checkbox>
              </div>
            </div>
            <div v-if="column.dataIndex === 'consumptionMethod'">
              <span>{{ record.consumptionMethodText || '-' }}</span>
              <!-- 分割线 -->
              <span class="flex w-47px">
                <a-divider
                  v-if="record.consumptionMethod !== '1' && !record.unrecorded && record.type !== '3'"
                  class="my-2px"
                />
              </span>
              <div
                v-if="record.consumptionMethod !== '1' && !record.unrecorded && record.type !== '3'"
                class="text-#888 text-3 flex flex-col"
              >
                <span>记录课时</span>
                <a-switch v-model:checked="record.recordAttendance" class="w-35px" />
              </div>
            </div>
            <div v-if="column.dataIndex === 'attendanceCount'">
              <div v-if="record.recordAttendance && !record.unrecorded" class="relative inline-flex flex-items-center">
                <span class="flex flex-items-center"><a-input-number
                  v-model:value="record.attendanceCount"
                  :min="0"
                  :precision="2"
                  :status="isRemainingInsufficient(record) ? 'error' : undefined"
                  class="w-80px mr-4px"
                />课时</span>
                <span
                  v-if="isRemainingInsufficient(record)"
                  class="absolute left-0 top-full text-#f33 text-12px leading-16px pt-2px whitespace-nowrap"
                >
                  剩余数量不足
                </span>
              </div>
              <!-- 当是未记录时，展示不计课时，不发送家长端消息提示 -->
              <span v-else-if="record.unrecorded && record.type !== '3'" class="flex flex-col">
                <span>不计课时</span>
                <span class="text-3 text-#999">不发送家长端 <br> 消息提示</span>
              </span>
              <span v-else-if="record.type === '3'">
                <div class="text-#888">免费试听学员</div>
                <span class="text-#999">不支持记课时</span>
              </span>
              <span v-else class="text-#888">不计课时</span>
            </div>
            <div v-if="column.dataIndex === 'internalNote'">
              <a-input v-model:value="record.internalNote" class="w-100px" placeholder="请输入" />
            </div>
            <div v-if="column.dataIndex === 'externalNote'">
              <a-input v-model:value="record.externalNote" class="w-100px" placeholder="请输入" />
            </div>
            <div v-else-if="column.dataIndex === 'action'">
              <a-space>
                <a v-if="record.canRemove" @click="handleRemoveStudent(record)">移出</a>
              </a-space>
            </div>
          </template>
        </a-table>
      </div>
      <!-- 自定义footer -->
      <template #footer>
        <div class="h-60px flex flex-items-center justify-between px-24px">
          <a-space :size="20">
            <a-dropdown>
              <template #overlay>
                <a-menu @click="handleAddStudent">
                  <a-menu-item key="1">
                    补课学员
                  </a-menu-item>
                  <a-menu-item key="2">
                    临时学员
                  </a-menu-item>
                  <a-menu-item key="3">
                    试听学员
                  </a-menu-item>
                </a-menu>
              </template>
              <a-button type="primary" ghost class="h-40px text-16px">
                添加学员
                <DownOutlined class="text-12px rotate-icon" />
              </a-button>
            </a-dropdown>
            <a-button type="primary" ghost class="h-40px text-16px" @click="handleBatchEdit">
              批量编辑点名数量
            </a-button>
          </a-space>
          <a-space :size="20">
            <div class="flex flex-col text-#222 text-16px font-500">
              <span class="mb-4px">共{{ data.length }}名学员</span>
              <span>到课{{ attendanceStats.attended }}人，请假{{ attendanceStats.leave }}人，旷课{{ attendanceStats.absent
              }}人，未记录{{
                attendanceStats.unrecorded }}人</span>
            </div>
            <a-button type="primary" class="h-48px text-18px w-140px font500" :loading="submittingRollCall" @click="handleConfirmRollCall">
              确认点名
            </a-button>
          </a-space>
        </div>
      </template>
    </a-drawer>
    <a-modal
      v-model:open="switchAccountModalOpen"
      class="roll-call-switch-account-modal"
      ok-text="确定"
      cancel-text="取消"
      :closable="false"
      width="720px"
      @ok="submitSwitchAccount"
      @cancel="closeSwitchAccountModal"
    >
      <template #title>
        <div class="roll-call-switch-account-modal__title">
          <span>选择扣费课程账户</span>
          <a-button type="text" class="close-btn" @click="closeSwitchAccountModal">
            <template #icon>
              <CloseOutlined class="text-5 close-icon" />
            </template>
          </a-button>
        </div>
      </template>

      <div class="roll-call-switch-account-modal__notice">
        <ExclamationCircleOutlined class="roll-call-switch-account-modal__notice-icon" />
        <span>以下为当前相关课程的扣费课程账户</span>
      </div>

      <div class="roll-call-switch-account-modal__table">
        <div class="roll-call-switch-account-modal__thead">
          <div class="roll-call-switch-account-modal__col roll-call-switch-account-modal__col--account">
            课程账户
          </div>
          <div class="roll-call-switch-account-modal__col roll-call-switch-account-modal__col--remain">
            剩余数量
          </div>
          <div class="roll-call-switch-account-modal__col roll-call-switch-account-modal__col--expire">
            有效日期/有效时段
          </div>
        </div>
        <a-spin :spinning="switchAccountLoading">
          <a-radio-group v-model:value="switchAccountSelectedId" class="roll-call-switch-account-modal__group custom-radio">
            <label
              v-for="acc in switchAccountOptions"
              :key="acc.id"
              class="roll-call-switch-account-modal__row"
              :class="{ 'is-active': String(switchAccountSelectedId) === String(acc.id) }"
            >
              <a-radio :value="acc.id" class="roll-call-switch-account-modal__radio" />
              <div class="roll-call-switch-account-modal__col roll-call-switch-account-modal__col--account">
                <div class="roll-call-switch-account-modal__account-name">
                  {{ acc.productName || acc.lessonName || '-' }}
                </div>
                <div class="roll-call-switch-account-modal__tags">
                  <a-tag color="#e9f2ff" :bordered="false">
                    {{ accountDeductTeachMethodText(acc.lessonType) }}
                  </a-tag>
                  <a-tag color="#eef3ff" :bordered="false">
                    {{ getChargingModeText(effectiveAccountChargingMode(acc)) }}
                  </a-tag>
                </div>
              </div>
              <div class="roll-call-switch-account-modal__col roll-call-switch-account-modal__col--remain">
                {{ switchAccountRemainText(acc) }}
              </div>
              <div class="roll-call-switch-account-modal__col roll-call-switch-account-modal__col--expire">
                {{ switchAccountExpireText(acc) }}
              </div>
            </label>
            <a-empty
              v-if="!switchAccountLoading && switchAccountOptions.length === 0"
              class="roll-call-switch-account-modal__empty"
              description="暂无可切换的扣费课程账户"
            />
          </a-radio-group>
        </a-spin>
      </div>
    </a-modal>
    <student-info-drawer v-model:open="openStudentDrawer" />
    <!-- 编辑上课信息 -->
    <EditClassInfoModal v-model:open="editClassInfoModal" />
    <!-- 添加学员modal -->
    <roll-call-add-student-modal
      v-model:open="addStudentModal"
      :title="addStudentModalTitle"
      :schedule-id="currentScheduleId"
      :student-type="addStudentType"
      @success="handleAddStudentSuccess"
    />
    <!-- 批量编辑 -->
    <roll-call-batch-edit-modal
      v-model:open="batchEditModal"
      :initial-class-number="Number(classTimetableDetail?.defaultStudentClassTime || 1)"
      @submit="handleBatchEditSubmit"
    />
  </div>
</template>

<style lang="less" scoped>
/* 添加旋转动画 */
@keyframes icon-rotate {
  from {
    transform: rotate(0deg);
  }

  to {
    transform: rotate(180deg);
  }
}

.close-btn {
  &:hover {
    background: transparent;

    .close-icon {
      animation: icon-rotate 0.3s linear;
    }
  }
}

/* 添加选中状态样式 */
.active-checkbox {
  &.attended-checkbox {
    :deep(.ant-checkbox-inner) {
      background-color: #1890ff !important;
      border-color: #1890ff !important;
    }
  }

  &.absent-checkbox {
    :deep(.ant-checkbox-inner) {
      background-color: #f33 !important;
      border-color: #f33 !important;
    }
  }

  &.leave-checkbox {
    :deep(.ant-checkbox-inner) {
      background-color: #f90 !important;
      border-color: #f90 !important;
    }
  }

  &.unrecorded-checkbox {
    :deep(.ant-checkbox-inner) {
      background-color: #888 !important;
      border-color: #888 !important;
    }
  }
}

.active-status {
  &.attended-status {
    color: #1890ff !important;
  }

  &.absent-status {
    color: #f33 !important;
  }

  &.leave-status {
    color: #f90 !important;
  }

  &.unrecorded-status {
    color: #888 !important;
  }

  opacity: 1 !important;
}

.inactive-status {
  opacity: 0.4;
  transition: opacity 0.3s, color 0.3s;

  &:hover {
    opacity: 1;
  }

  &.attended-status:hover {
    color: #1890ff !important;
  }

  &.absent-status:hover {
    color: #f33 !important;
  }

  &.leave-status:hover {
    color: #f90 !important;
  }

  &.unrecorded-status:hover {
    color: #888 !important;
  }

  /* 添加复选框hover效果 */
  &.attended-status:hover {
    :deep(.ant-checkbox-inner) {
      border-color: #1890ff !important;
    }
  }

  &.absent-status:hover {
    :deep(.ant-checkbox-inner) {
      border-color: #f33 !important;
    }
  }

  &.leave-status:hover {
    :deep(.ant-checkbox-inner) {
      border-color: #f90 !important;
    }
  }

  &.unrecorded-status:hover {
    :deep(.ant-checkbox-inner) {
      border-color: #888 !important;
    }
  }
}

/* 添加旋转过渡效果 */
.rotate-icon {
  display: inline-block;
  transition: transform 0.3s ease;
}

/* 当按钮悬停时旋转图标 */
.h-40px:hover .rotate-icon {
  transform: rotate(180deg);
}

.roll-call-switch-account-modal__title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  width: 100%;
  font-size: 18px;
  font-weight: 600;
  color: #1f1f1f;
}

.roll-call-switch-account-modal__notice {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 18px;
  padding: 11px 14px;
  border-radius: 10px;
  background: #edf4ff;
  color: #3f6fdc;
  font-size: 13px;
  line-height: 20px;
}

.roll-call-switch-account-modal__notice-icon {
  font-size: 14px;
}

.roll-call-switch-account-modal__table {
  overflow: hidden;
  border: 1px solid #edf0f5;
  border-radius: 14px;
  background: #fff;
}

.roll-call-switch-account-modal__thead {
  display: grid;
  grid-template-columns: minmax(0, 1.5fr) 132px 210px;
  align-items: center;
  padding: 14px 18px 14px 50px;
  background: #f7f9fc;
  border-bottom: 1px solid #edf0f5;
}

.roll-call-switch-account-modal__col {
  min-width: 0;
  font-size: 13px;
}

.roll-call-switch-account-modal__thead .roll-call-switch-account-modal__col {
  color: #667085;
  font-weight: 600;
}

.roll-call-switch-account-modal__group {
  display: block;
}

.roll-call-switch-account-modal__row {
  display: grid;
  grid-template-columns: 20px minmax(0, 1.5fr) 132px 210px;
  align-items: center;
  column-gap: 10px;
  padding: 16px 18px;
  border-bottom: 1px solid #f0f2f5;
  cursor: pointer;
  transition: background-color 0.2s ease, border-color 0.2s ease;
}

.roll-call-switch-account-modal__row:last-child {
  border-bottom: 0;
}

.roll-call-switch-account-modal__row:hover {
  background: #fafcff;
}

.roll-call-switch-account-modal__row.is-active {
  background: #eaf3ff;
}

.roll-call-switch-account-modal__radio {
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.roll-call-switch-account-modal__account-name {
  color: #1f1f1f;
  font-size: 15px;
  font-weight: 600;
  line-height: 22px;
}

.roll-call-switch-account-modal__tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 8px;
}

.roll-call-switch-account-modal__tags :deep(.ant-tag) {
  margin-inline-end: 0;
  border-radius: 999px;
  padding-inline: 10px;
  color: #4a67c7;
  font-size: 12px;
  line-height: 22px;
}

.roll-call-switch-account-modal__col--remain,
.roll-call-switch-account-modal__col--expire {
  color: #4b5565;
  line-height: 22px;
}

.roll-call-switch-account-modal__col--expire {
  white-space: nowrap;
}

.roll-call-switch-account-modal__empty {
  padding: 40px 0 36px;
}

.roll-call-switch-account-modal :deep(.ant-radio-wrapper) {
  margin-inline-end: 0;
}

.roll-call-switch-account-modal :deep(.ant-radio) {
  top: 0;
}

.roll-call-switch-account-modal :deep(.ant-radio-inner) {
  width: 18px;
  height: 18px;
}

.roll-call-switch-account-modal :deep(.ant-radio-checked .ant-radio-inner) {
  box-shadow: 0 0 0 4px rgba(76, 132, 255, 0.12);
}

.roll-call-switch-account-modal :deep(.ant-radio-input:focus + .ant-radio-inner) {
  box-shadow: 0 0 0 4px rgba(76, 132, 255, 0.12);
}

@media (max-width: 900px) {
  .roll-call-switch-account-modal__thead {
    grid-template-columns: minmax(0, 1.2fr) 112px 180px;
    padding-right: 14px;
    padding-left: 44px;
  }

  .roll-call-switch-account-modal__row {
    grid-template-columns: 18px minmax(0, 1.2fr) 112px 180px;
    padding-right: 14px;
    padding-left: 14px;
  }
}
</style>

<style>
.roll-call-switch-account-modal .ant-modal-header {
  padding: 18px 24px 14px !important;
  margin-bottom: 0;
  border-bottom: 0;
}

.roll-call-switch-account-modal .ant-modal-body {
  padding: 6px 24px 8px !important;
}

.roll-call-switch-account-modal .ant-modal-footer {
  padding: 16px 24px 20px !important;
  border-top-color: #f0f2f5;
}
</style>
