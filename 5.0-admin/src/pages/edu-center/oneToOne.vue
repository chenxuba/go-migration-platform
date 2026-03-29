<script setup>
import { computed, createVNode, onMounted, reactive, ref } from 'vue'
import { debounce } from 'lodash-es'
import { CloseOutlined, DownOutlined, ExclamationCircleOutlined, QuestionCircleOutlined } from '@ant-design/icons-vue'
import { Modal } from 'ant-design-vue'
import dayjs from 'dayjs'
import { useRouter } from 'vue-router'
import StaffSelect from '@/components/common/staff-select.vue'
import FinishOneToOneCourseModal from '@/components/edu-center/one-to-one/finish-one-to-one-course-modal.vue'
import oneToOneDrawer from '@/components/edu-center/one-to-one/one-to-one-drawer.vue'
import { useTableColumns } from '@/composables/useTableColumns'
import { handleDateRangeParams } from '@/utils/dateRangeParams'
import messageService from '@/utils/messageService'
import {
  batchAssignOneToOneClassTeacherApi,
  batchUpdateOneToOneClassTimeApi,
  checkOneToOneNameApi,
  addCloseTuitionAccountOrderApi,
  closeOneToOneApi,
  reopenOneToOneApi,
  getOneToOneByIdApi,
  getOneToOneListApi,
  listTuitionAccountsByStudentAndLessonApi,
  updateOneToOneApi,
} from '@/api/edu-center/one-to-one'
import { Sex, SexLabel } from '@/enums'

const router = useRouter()
const allFilterRef = ref()
const loading = ref(false)
const dataSource = ref([])
const selectedRows = ref([])
const selectedRowKeys = ref([])
const actionRows = ref([])
const currentRecord = ref(null)
const drawerTuitionAccounts = ref([])
const drawerOpen = ref(false)
const totalStudentCount = ref(0)
const quickCounts = ref({
  unassignedTeacherCount: 0,
  unscheduledCount: 0,
})

const advisorModalOpen = ref(false)
const advisorModalTitle = ref('批量分配班主任')
const advisorSubmitting = ref(false)
const advisorForm = reactive({
  classTeacherIds: [],
})

const classTimeModalOpen = ref(false)
const classTimeSubmitting = ref(false)
const classTimeForm = reactive({
  classTimeRecordMode: 1,
  studentClassTime: 1,
  teacherClassTime: 0,
})

const classTimeBatchUnitLabel = computed(() =>
  Number(classTimeForm.classTimeRecordMode) === 2 ? '课时/小时' : '课时',
)

const classTimeBatchHint = computed(() =>
  Number(classTimeForm.classTimeRecordMode) === 2
    ? '每次点名，学员和上课教师记录的课时会根据日程时长自动计算课时（点名时支持调整）'
    : '每次点名，学员和上课教师记录的课时数默认为此数值（点名时支持调整）',
)

const classTimeBatchSelectionSummary = computed(() => {
  const rows = actionRows.value
  const n = rows.length
  const names = rows.map(r => r.name).filter(Boolean).join('，')
  return { n, names }
})

const finishCourseModalOpen = ref(false)
const finishCourseRecord = ref(null)

const editModalOpen = ref(false)
const editLoading = ref(false)
const editSubmitting = ref(false)
const currentEditRecord = ref(null)
const editForm = reactive({
  id: '',
  studentId: '',
  lessonId: '',
  studentName: '',
  lessonName: '',
  name: '',
  teacherIds: [],
  defaultTeacherId: undefined,
  classRoomId: undefined,
  classRoomName: '',
  defaultStudentClassTime: 1,
  defaultTeacherClassTime: 0,
  defaultClassTimeRecordMode: 1,
  remark: '',
  classProperties: [],
})

const editClassTimeUnitLabel = computed(() =>
  Number(editForm.defaultClassTimeRecordMode) === 2 ? '课时/小时' : '课时',
)

const editClassTimeHint = computed(() =>
  Number(editForm.defaultClassTimeRecordMode) === 2
    ? '每次点名，学员和上课教师记录的课时会根据日程时长自动计算课时（点名时支持调整）'
    : '每次点名，学员和上课教师记录的课时数默认为此数值（点名时支持调整）',
)

const displayArray = ref([
  'createTime',
  'enrolledCourse',
  'classTeacher',
  'createUser',
  'orNotFenClass',
  'doYouSchedule',
  'openClassStatus',
  'currentStatus',
])

const pagination = ref({
  current: 1,
  pageSize: 10,
  total: 0,
  showSizeChanger: true,
  pageSizeOptions: ['10', '20', '50', '100'],
  showQuickJumper: true,
  showTotal: total => `共 ${total} 条`,
})

const queryState = ref({
  studentId: undefined,
  lessonIds: undefined,
  classTeacherId: undefined,
  defaultTeacherId: undefined,
  hasClassTeacher: undefined,
  isScheduled: undefined,
  /** 默认仅看开班中（与 all-filter 开班状态 id=1 一致） */
  status: [1],
  classStudentStatus: undefined,
  startDate: undefined,
  endDate: undefined,
  createTime: undefined,
})

function resetQueryState() {
  Object.keys(queryState.value).forEach((key) => {
    queryState.value[key] = undefined
  })
}

const handleFilterUpdate = debounce((updates = {}, isClearAll = false, id, type) => {
  if (isClearAll) {
    resetQueryState()
  } else {
    Object.entries(updates).forEach(([key, value]) => {
      queryState.value[key] = value
    })
  }
  pagination.value.current = 1
  selectedRows.value = []
  selectedRowKeys.value = []
  getOneToOneList(queryState.value, id, type)
}, 200, { leading: true, trailing: false })

function handleQuickFilterChange(value, isClearAll, id, type) {
  if (isClearAll) {
    handleFilterUpdate({}, true, id, type)
    return
  }

  if (value === 1) {
    handleFilterUpdate({
      hasClassTeacher: false,
      isScheduled: undefined,
    }, false, id, type)
    return
  }
  if (value === 2) {
    handleFilterUpdate({
      hasClassTeacher: undefined,
      isScheduled: false,
    }, false, id, type)
    return
  }

  if (type === 'quickOneToOne' && id === 1) {
    handleFilterUpdate({ hasClassTeacher: undefined }, false, id, type)
    return
  }
  if (type === 'quickOneToOne' && id === 2) {
    handleFilterUpdate({ isScheduled: undefined }, false, id, type)
    return
  }

  handleFilterUpdate({
    hasClassTeacher: undefined,
    isScheduled: undefined,
  }, false, id, type)
}

const filterUpdateHandlers = computed(() => ({
  'update:stuPhoneSearchFilter': (val, isClearAll, id, type) => {
    handleFilterUpdate({ studentId: val || undefined }, isClearAll, id, type)
  },
  'update:enrolledCourseFilter': (val, isClearAll, id, type) => {
    handleFilterUpdate({ lessonIds: Array.isArray(val) && val.length ? val : undefined }, isClearAll, id, type)
  },
  'update:classTeacherFilter': (val, isClearAll, id, type) => {
    handleFilterUpdate({ classTeacherId: val || undefined }, isClearAll, id, type)
  },
  'update:createUserFilter': (val, isClearAll, id, type) => {
    handleFilterUpdate({ defaultTeacherId: val || undefined }, isClearAll, id, type)
  },
  'update:orNotFenClassFilter': (val, isClearAll, id, type) => {
    if (Array.isArray(val) && val.length === 1) {
      handleFilterUpdate({ hasClassTeacher: val[0] === 1 }, isClearAll, id, type)
      return
    }
    handleFilterUpdate({ hasClassTeacher: undefined }, isClearAll, id, type)
  },
  'update:doYouScheduleFilter': (val, isClearAll, id, type) => {
    if (val === 1) {
      handleFilterUpdate({ isScheduled: true }, isClearAll, id, type)
      return
    }
    if (val === 2) {
      handleFilterUpdate({ isScheduled: false }, isClearAll, id, type)
      return
    }
    handleFilterUpdate({ isScheduled: undefined }, isClearAll, id, type)
  },
  'update:openClassStatusFilter': (val, isClearAll, id, type) => {
    handleFilterUpdate({ status: val ? [val] : undefined }, isClearAll, id, type)
  },
  'update:currentStatusFilter': (val, isClearAll, id, type) => {
    handleFilterUpdate({ classStudentStatus: Array.isArray(val) && val.length ? val : undefined }, isClearAll, id, type)
  },
  'update:createTimeFilter': (val, isClearAll, id, type) => {
    handleFilterUpdate({ createTime: Array.isArray(val) && val.length ? val : undefined }, isClearAll, id, type)
  },
  'update:quickFilter': handleQuickFilterChange,
}))

const allColumns = ref([
  { title: '1对1', dataIndex: 'name', key: 'name', fixed: 'left', width: 180, required: true },
  { title: '学员/性别', dataIndex: 'student', key: 'student', width: 140, required: true },
  { title: '联系电话', dataIndex: 'phone', key: 'phone', width: 120 },
  { title: '上课课程', dataIndex: 'lesson', key: 'lesson', width: 180 },
  { title: '当前课程账户', dataIndex: 'account', key: 'account', width: 190 },
  { title: '报读数量', dataIndex: 'totalQuantity', key: 'totalQuantity', width: 150 },
  { title: '有效期至', dataIndex: 'expireTime', key: 'expireTime', width: 120 },
  { title: '停课日期', dataIndex: 'suspendedTime', key: 'suspendedTime', width: 120 },
  { title: '结课日期', dataIndex: 'classEndingTime', key: 'classEndingTime', width: 120 },
  { title: '已用数量', dataIndex: 'usedQuantity', key: 'usedQuantity', width: 120 },
  { title: '剩余数量', dataIndex: 'remainQuantity', key: 'remainQuantity', width: 120 },
  { title: '已用学费金额', dataIndex: 'usedTuition', key: 'usedTuition', width: 140 },
  { title: '剩余学费金额', dataIndex: 'remainTuition', key: 'remainTuition', width: 140 },
  { title: '总学费', dataIndex: 'totalTuition', key: 'totalTuition', width: 120 },
  { title: '班主任', dataIndex: 'classTeacher', key: 'classTeacher', width: 120 },
  { title: '默认上课教师', dataIndex: 'defaultTeacher', key: 'defaultTeacher', width: 140 },
  { title: '上课时间', dataIndex: 'classTime', key: 'classTime', width: 120 },
  { title: '最近上课时间', dataIndex: 'lastFinishedLessonDay', key: 'lastFinishedLessonDay', width: 150 },
  { title: '是否排课', dataIndex: 'isScheduled', key: 'isScheduled', width: 110 },
  { title: '已上/排课', dataIndex: 'lessonDayCount', key: 'lessonDayCount', width: 120 },
  { title: '创建时间', dataIndex: 'createdTime', key: 'createdTime', width: 160 },
  { title: '开班状态', dataIndex: 'status', key: 'status', width: 110 },
  { title: '开课状态', dataIndex: 'classStudentStatus', key: 'classStudentStatus', width: 110 },
  { title: '操作', dataIndex: 'action', key: 'action', fixed: 'right', width: 180 },
])

const { selectedValues, columnOptions, filteredColumns, totalWidth } = useTableColumns({
  storageKey: 'one-to-one-list',
  allColumns,
  excludeKeys: ['action'],
})

const rowSelection = {
  selectedRowKeys,
  onChange: (keys, rows) => {
    selectedRowKeys.value = keys
    selectedRows.value = rows
  },
}

function buildValidQueryParams(newQueryParams = {}) {
  const dateRangeMappings = {
    createTime: {
      begin: 'startDate',
      end: 'endDate',
    },
  }

  queryState.value.startDate = undefined
  queryState.value.endDate = undefined

  let merged = Object.assign({}, newQueryParams)
  if (Object.keys(merged).length > 0) {
    merged = handleDateRangeParams(merged, dateRangeMappings)
  }
  Object.assign(queryState.value, merged)

  const originalRangeFields = ['createTime']
  return Object.fromEntries(
    Object.entries(queryState.value).filter(([key, value]) => value !== undefined && !originalRangeFields.includes(key)),
  )
}

async function getOneToOneList(newQueryParams = {}, id, type) {
  loading.value = true
  try {
    const validQueryParams = buildValidQueryParams(newQueryParams)
    const res = await getOneToOneListApi({
      pageRequestModel: {
        needTotal: true,
        pageSize: pagination.value.pageSize,
        pageIndex: pagination.value.current,
        skipCount: 0,
      },
      queryModel: validQueryParams,
    })

    if (res.code === 200 && res.result) {
      dataSource.value = Array.isArray(res.result.list) ? res.result.list : []
      pagination.value.total = res.result.total || 0
      totalStudentCount.value = res.result.studentCount || 0
      quickCounts.value = {
        unassignedTeacherCount: dataSource.value.filter((item) => {
          const hasAdvisor = item.classTeacherId && item.classTeacherId !== '0'
          const hasClassTeachers = Array.isArray(item.teacherList) && item.teacherList.length > 0
          return !hasAdvisor && !hasClassTeachers
        }).length,
        unscheduledCount: dataSource.value.filter(item => !item.isScheduled).length,
      }
      allFilterRef.value?.clearQuickFilter(id, type)
      return
    }
    messageService.error(res.message || '获取1对1列表失败')
  } catch (error) {
    console.error('get one to one list failed', error)
    messageService.error('获取1对1列表失败')
  } finally {
    loading.value = false
  }
}

function handleTableChange(pageInfo) {
  pagination.value.current = pageInfo.current
  pagination.value.pageSize = pageInfo.pageSize
  getOneToOneList()
}

function handleEnroll() {
  router.push('/edu-center/registr-renewal')
}

function handleCreateOneToOne() {
  messageService.info('创建1对1功能暂未实现')
}

function handleReopenClass(record) {
  const id = record?.id
  if (!id) {
    messageService.error('缺少1对1班级ID')
    return
  }
  Modal.confirm({
    title: '恢复开班',
    centered: true,
    icon: createVNode(ExclamationCircleOutlined),
    content: '确定将该1对1恢复为开班中吗？',
    async onOk() {
      try {
        const res = await reopenOneToOneApi({ id: String(id) })
        if (res.code === 200) {
          messageService.success('已恢复开班')
          getOneToOneList()
          return
        }
        messageService.error(res.message || '恢复开班失败')
      }
      catch (err) {
        console.error(err)
        messageService.error('恢复开班失败')
      }
    },
  })
}

/** 二次确认后再弹出：结班并结课 / 仅结班 */
function openOneToOneCloseClassConfirm(record) {
  const id = record?.id
  if (!id) {
    messageService.error('缺少1对1班级ID')
    return
  }
  Modal.confirm({
    title: '1对1结班',
    centered: true,
    closable: false,
    maskClosable: false,
    keyboard: false,
    icon: createVNode(ExclamationCircleOutlined),
    content:
      '是否确认对1对1进行结班且结课，结班后会同步删除相关的日程，被删除的日程不可恢复，请谨慎操作',
    okText: '结班并结课',
    cancelText: '仅结班',
    async onOk() {
      try {
        const res = await closeOneToOneApi({ id: String(id) })
        if (res.code !== 200) {
          messageService.error(res.message || '结班失败')
          return Promise.reject(new Error(res.message || '结班失败'))
        }
        messageService.success('结班成功')
        await getOneToOneList()
        const updated = dataSource.value.find(r => String(r.id) === String(id)) || record
        finishCourseRecord.value = updated
        finishCourseModalOpen.value = true
      }
      catch (err) {
        console.error(err)
        messageService.error('结班失败')
        return Promise.reject(err)
      }
    },
    async onCancel() {
      try {
        const res = await closeOneToOneApi({ id: String(id) })
        if (res.code === 200) {
          messageService.success('结班成功')
          getOneToOneList()
          return
        }
        messageService.error(res.message || '结班失败')
      }
      catch (err) {
        console.error(err)
        messageService.error('结班失败')
      }
    },
  })
}

function handleFinishCourse(record) {
  openOneToOneCloseClassConfirm(record)
}

async function handleFinishCourseModalConfirm(payload) {
  const r = finishCourseRecord.value
  const ta = r?.tuitionAccount
  const tuitionAccountId = ta?.id || r?.tuitionAccountId
  if (!tuitionAccountId) {
    messageService.error('缺少学费账户ID')
    return
  }
  const quantity = Number(ta?.remainQuantity || 0)
  const freeQuantity = Number(ta?.remainFreeQuantity || 0)
  const tuition = Number(ta?.remainTuition ?? 0)
  if (quantity + freeQuantity <= 0 && tuition <= 0) {
    messageService.error('当前无可结课的剩余课时或学费')
    return
  }
  try {
    const res = await addCloseTuitionAccountOrderApi({
      tuitionAccountId: String(tuitionAccountId),
      quantity,
      freeQuantity,
      tuition,
      remark: payload?.remark || '',
    })
    if (res.code === 200) {
      messageService.success('结课成功')
      finishCourseModalOpen.value = false
      getOneToOneList()
      return
    }
    messageService.error(res.message || '结课失败')
  }
  catch (err) {
    console.error(err)
    messageService.error('结课失败')
  }
}

function closeEditModal() {
  editModalOpen.value = false
}

function resetEditForm() {
  editForm.id = ''
  editForm.studentId = ''
  editForm.lessonId = ''
  editForm.studentName = ''
  editForm.lessonName = ''
  editForm.name = ''
  editForm.teacherIds = []
  editForm.defaultTeacherId = undefined
  editForm.classRoomId = undefined
  editForm.classRoomName = ''
  editForm.defaultStudentClassTime = 1
  editForm.defaultTeacherClassTime = 0
  editForm.defaultClassTimeRecordMode = 1
  editForm.remark = ''
  editForm.classProperties = []
}

function getCurrentActionRows() {
  return actionRows.value.length > 0 ? actionRows.value : selectedRows.value
}

function getGenderText(sex) {
  return SexLabel[sex] || SexLabel[Sex.Unknown]
}

function isZeroDateValue(value) {
  if (!value)
    return true
  if (typeof value === 'string' && value.startsWith('0001-01-01'))
    return true
  const parsed = dayjs(value)
  return !parsed.isValid() || parsed.year() <= 1
}

function formatDateTime(value) {
  if (isZeroDateValue(value))
    return '-'
  return dayjs(value).format('YYYY-MM-DD HH:mm')
}

function formatDate(value) {
  if (isZeroDateValue(value))
    return '-'
  return dayjs(value).format('YYYY-MM-DD')
}

function formatMoney(value) {
  return `¥ ${Number(value || 0).toLocaleString('zh-CN', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  })}`
}

/** 列表行学费账户：lessonChargingMode 为 0 时按有效期+数量推断按时段（与后端聚合一致） */
function effectiveListLessonChargingMode(record) {
  const ta = record?.tuitionAccount
  const m = Number(ta?.lessonChargingMode)
  if (m > 0)
    return m
  if (ta?.enableExpireTime && Number(ta?.totalQuantity || 0) > 0)
    return 2
  return m || 0
}

function getChargingModeText(mode) {
  const modeMap = {
    1: '按课时',
    2: '按时段',
    3: '按金额',
  }
  return modeMap[mode] || '-'
}

function getQuantityUnit(mode) {
  if (mode === 1)
    return '课时'
  if (mode === 2)
    return '天'
  if (mode === 3)
    return '元'
  return ''
}

function calcUsedQuantity(record) {
  const total = Number(record.tuitionAccount?.totalQuantity || 0) + Number(record.tuitionAccount?.totalFreeQuantity || 0)
  const remain = Number(record.tuitionAccount?.remainQuantity || 0) + Number(record.tuitionAccount?.remainFreeQuantity || 0)
  return Math.max(total - remain, 0)
}

function calcRemainQuantity(record) {
  return Number(record.tuitionAccount?.remainQuantity || 0) + Number(record.tuitionAccount?.remainFreeQuantity || 0)
}

function calcUsedTuition(record) {
  return Math.max(Number(record.tuitionAccount?.totalTuition || 0) - Number(record.tuitionAccount?.remainTuition || 0), 0)
}

function shouldShowSchedulePlaceholder(record) {
  return !record.isScheduled || Number(record.one2OneLessonDayInfo?.lessonDayCount || 0) <= 0
}

function formatClassTime(record) {
  if (shouldShowSchedulePlaceholder(record))
    return '-'
  const ct = Number(record.classTime || 0)
  if (ct <= 0)
    return '-'
  const mode = effectiveListLessonChargingMode(record)
  const unit = mode === 2 ? '天' : '课时'
  return `${ct}${unit}/次`
}

function formatLessonDaySummary(record) {
  const total = Number(record.one2OneLessonDayInfo?.lessonDayCount || 0)
  const completed = Number(record.one2OneLessonDayInfo?.completeLessonDayCount || 0)
  if (total <= 0)
    return '-'
  return `${completed}/${total}节`
}

function getOpenClassStatus(status) {
  if (status === 2)
    return { text: '已结班', className: 'text-#888 bg-#f5f5f5' }
  return { text: '开班中', className: 'text-#06f bg-#e6f0ff' }
}

function isOneToOneClassClosed(record) {
  return Number(record?.status) === 2
}

function getClassStudentStatus(status) {
  if (status === 2)
    return { text: '已开课', className: 'text-#f90 bg-#fff5e6' }
  if (status === 3)
    return { text: '已结课', className: 'text-#888 bg-#f5f5f5' }
  return { text: '正常', className: 'text-#0c3 bg-#e6ffec' }
}

async function openDrawer(record) {
  currentRecord.value = record
  drawerTuitionAccounts.value = []
  drawerOpen.value = true
  try {
    const detailRes = await getOneToOneByIdApi(record?.id)
    if (detailRes.code === 200 && detailRes.result)
      currentRecord.value = { ...record, ...detailRes.result }

    const sid = currentRecord.value?.studentId
    const lid = currentRecord.value?.lessonId
    if (sid && lid) {
      const accRes = await listTuitionAccountsByStudentAndLessonApi({
        studentId: String(sid),
        lessonId: String(lid),
      })
      if (accRes.code === 200 && Array.isArray(accRes.result?.list))
        drawerTuitionAccounts.value = accRes.result.list
    }
  }
  catch (error) {
    console.error('open one-to-one drawer failed', error)
  }
}

function handleDrawerEdit(record) {
  drawerOpen.value = false
  openEditModal(record || currentRecord.value)
}

function ensureSelectedRows() {
  if (selectedRows.value.length > 0)
    return true
  messageService.warning('请先选择1对1记录')
  return false
}

function openBatchAction(action) {
  if (!ensureSelectedRows())
    return
  actionRows.value = [...selectedRows.value]

  if (action === 'assign') {
    advisorModalTitle.value = '批量分配班主任'
    advisorForm.classTeacherIds = []
    advisorModalOpen.value = true
    return
  }
  if (action === 'replace') {
    advisorModalTitle.value = '批量替换班主任'
    advisorForm.classTeacherIds = []
    advisorModalOpen.value = true
    return
  }
  if (action === 'classTime') {
    const current = actionRows.value[0]
    classTimeForm.classTimeRecordMode = Number(current?.defaultClassTimeRecordMode || 1)
    // 每次打开弹窗：学员默认 1、教师默认 0（不沿用列表当前行数值）
    classTimeForm.studentClassTime = 1
    classTimeForm.teacherClassTime = 0
    classTimeModalOpen.value = true
    return
  }
}

async function submitAdvisorBatch() {
  const teacherIds = Array.isArray(advisorForm.classTeacherIds)
    ? advisorForm.classTeacherIds.filter(id => id !== undefined && id !== null && `${id}` !== '')
    : []
  if (teacherIds.length === 0) {
    messageService.warning('请选择班主任')
    return
  }
  const rows = getCurrentActionRows()
  advisorSubmitting.value = true
  try {
    const res = await batchAssignOneToOneClassTeacherApi({
      ids: rows.map(item => item.id),
      classTeacherIds: teacherIds.map(id => String(id)),
    })
    if (res.code !== 200)
      throw new Error(res.message || '批量更新班主任失败')
    advisorModalOpen.value = false
    messageService.success(`${advisorModalTitle.value}成功`)
    resetSelection()
    await getOneToOneList()
  } catch (error) {
    console.error('batch assign advisor failed', error)
    messageService.error(error?.message || '批量更新班主任失败')
  } finally {
    advisorSubmitting.value = false
  }
}

async function submitClassTimeBatch() {
  const rows = getCurrentActionRows()
  classTimeSubmitting.value = true
  try {
    const studentCt = Number(classTimeForm.studentClassTime || 0)
    const res = await batchUpdateOneToOneClassTimeApi({
      ids: rows.map(item => item.id),
      classTime: studentCt,
      studentClassTime: studentCt,
      teacherClassTime: Number(classTimeForm.teacherClassTime || 0),
      classTimeRecordMode: Number(classTimeForm.classTimeRecordMode || 1),
    })
    if (res.code !== 200)
      throw new Error(res.message || '修改记录课时失败')
    classTimeModalOpen.value = false
    messageService.success('修改记录课时成功')
    resetSelection()
    await getOneToOneList()
  } catch (error) {
    console.error('batch update class time failed', error)
    messageService.error(error?.message || '修改记录课时失败')
  } finally {
    classTimeSubmitting.value = false
  }
}

function resetSelection() {
  selectedRows.value = []
  selectedRowKeys.value = []
  actionRows.value = []
}

async function openEditModal(record) {
  currentEditRecord.value = record
  resetEditForm()
  editModalOpen.value = true
  editLoading.value = true
  try {
    const res = await getOneToOneByIdApi(record?.id)
    if (res.code !== 200 || !res.result) {
      throw new Error(res.message || '获取1对1详情失败')
    }
    const detail = res.result
    editForm.id = detail.id || ''
    editForm.studentId = detail.studentId || ''
    editForm.lessonId = detail.lessonId || ''
    editForm.studentName = detail.studentName || ''
    editForm.lessonName = detail.lessonName || ''
    editForm.name = detail.name || ''
    editForm.teacherIds = Array.isArray(detail.teacherList)
      ? detail.teacherList.map(item => Number(item.teacherId)).filter(Boolean)
      : []
    editForm.defaultTeacherId = detail.defaultTeacherId && detail.defaultTeacherId !== '0'
      ? Number(detail.defaultTeacherId)
      : undefined
    editForm.classRoomId = detail.classroomId && detail.classroomId !== '0'
      ? Number(detail.classroomId)
      : undefined
    editForm.classRoomName = detail.classroomName || ''
    editForm.defaultStudentClassTime = Number(detail.defaultStudentClassTime || 1)
    editForm.defaultTeacherClassTime = Number(detail.defaultTeacherClassTime || 0)
    editForm.defaultClassTimeRecordMode = Number(detail.defaultClassTimeRecordMode || 1)
    editForm.remark = detail.remark || ''
    editForm.classProperties = Array.isArray(detail.classProperties) ? [...detail.classProperties] : []
  } catch (error) {
    console.error('get one to one detail failed', error)
    messageService.error(error?.message || '获取1对1详情失败')
    editModalOpen.value = false
  } finally {
    editLoading.value = false
  }
}

function buildEditTeacherIds() {
  const teacherIds = Array.isArray(editForm.teacherIds) ? [...editForm.teacherIds] : []
  if (editForm.defaultTeacherId && !teacherIds.includes(editForm.defaultTeacherId)) {
    teacherIds.push(editForm.defaultTeacherId)
  }
  return teacherIds
}

async function submitEditModal() {
  if (!editForm.name.trim()) {
    messageService.warning('请输入1对1名称')
    return
  }
  editSubmitting.value = true
  try {
    const checkRes = await checkOneToOneNameApi({
      name: editForm.name.trim(),
      exceptId: editForm.id,
      isOne2One: true,
    })
    if (checkRes.code !== 200) {
      throw new Error(checkRes.message || '校验1对1名称失败')
    }
    if (checkRes.result) {
      messageService.error('1对1名称已存在')
      return
    }

    const teacherIds = buildEditTeacherIds()
    const updateRes = await updateOneToOneApi({
      id: editForm.id,
      studentId: editForm.studentId,
      lessonId: editForm.lessonId,
      name: editForm.name.trim(),
      teacherId: teacherIds.map(id => String(id)),
      defaultTeacherId: editForm.defaultTeacherId ? String(editForm.defaultTeacherId) : '',
      defaultStudentClassTime: Number(editForm.defaultStudentClassTime || 0),
      defaultTeacherClassTime: Number(editForm.defaultTeacherClassTime || 0),
      defaultClassTimeRecordMode: Number(editForm.defaultClassTimeRecordMode || 1),
      remark: editForm.remark.trim(),
      classProperties: Array.isArray(editForm.classProperties) ? editForm.classProperties : [],
    })
    if (updateRes.code !== 200) {
      throw new Error(updateRes.message || '更新1对1失败')
    }

    messageService.success('编辑1对1成功')
    editModalOpen.value = false
    await getOneToOneList()
  } catch (error) {
    console.error('update one to one failed', error)
    messageService.error(error?.message || '编辑1对1失败')
  } finally {
    editSubmitting.value = false
  }
}

function handleSchedule(record) {
  router.push({
    path: '/edu-center/timetable',
    query: {
      mode: 'one-to-one',
      oneToOneId: record.id,
      studentId: record.studentId,
    },
  })
}

onMounted(() => {
  getOneToOneList()
})
</script>

<template>
  <div>
    <div class="filter-wrap bg-white pl-3 pr-3 rounded-4">
      <all-filter
        ref="allFilterRef"
        :display-array="displayArray"
        :is-quick-one-to-one-show="true"
        :is-show-search-stu-phone="true"
        :student-status="1"
        :default-open-class-status="1"
        :create-user-label="'默认上课教师'"
        :one-to-one-mode="true"
        :one-to-one-quick-counts="quickCounts"
        v-on="filterUpdateHandlers"
      />
    </div>

    <div class="student-list mt-2 pt-3 pb-3 pl-6 pr-6 bg-white rounded-4">
      <div class="tab-table">
        <div class="table-title flex justify-between">
          <div class="total">
            共{{ pagination.total || 0 }}条，关联学员 {{ totalStudentCount || 0 }} 人
            <span v-if="selectedRowKeys.length > 0" class="ml-2 text-blue-600">
              （已选 {{ selectedRowKeys.length }} 条）
              <a-button type="link" size="small" class="p-0 ml-1" @click="resetSelection">
                清空选择
              </a-button>
            </span>
          </div>
          <div class="edit flex">
            <a-dropdown class="mr-2">
              <template #overlay>
                <a-menu @click="({ key }) => openBatchAction(key)">
                  <a-menu-item key="assign">
                    批量分配班主任
                  </a-menu-item>
                  <a-menu-item key="replace">
                    批量替换班主任
                  </a-menu-item>
                  <a-menu-item key="classTime">
                    批量修改记录课时
                  </a-menu-item>
                </a-menu>
              </template>
              <a-button>
                批量操作
                <DownOutlined :style="{ fontSize: '10px' }" />
              </a-button>
            </a-dropdown>
            <a-button type="primary" class="mr-2 w-25 whitespace-nowrap" @click="handleEnroll">
              报名
            </a-button>
            <a-button type="primary" class="mr-2 w-30 whitespace-nowrap" @click="handleCreateOneToOne">
              创建1对1
            </a-button>
            <customize-code
              v-model:checked-values="selectedValues"
              :options="columnOptions"
              :total="allColumns.length - 1"
              :num="selectedValues.length - 1"
            />
          </div>
        </div>

        <div class="table-content mt-2">
          <a-table
            row-key="id"
            :data-source="dataSource"
            :loading="loading"
            :pagination="pagination"
            :columns="filteredColumns"
            :row-selection="rowSelection"
            :scroll="{ x: totalWidth }"
            size="small"
            @change="handleTableChange"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'name'">
                <a class="font500" @click="openDrawer(record)">{{ record.name || '-' }}</a>
              </template>
              <template v-if="column.key === 'student'">
                <student-avatar
                  :id="record.studentId"
                  :name="record.studentName || '-'"
                  :gender="getGenderText(record.sex)"
                  :avatar-url="record.avatar"
                  :show-age="false"
                  default-active-key="0"
                />
              </template>
              <template v-if="column.key === 'phone'">
                {{ record.phone || '-' }}
              </template>
              <template v-if="column.key === 'lesson'">
                <div>{{ record.lessonName || '-' }}</div>
              </template>
              <template v-if="column.key === 'account'">
                <div>{{ record.tuitionAccount?.productName || record.lessonName || '-' }}</div>
                <div class="text-3 text-#888">
                  1对1授课｜{{ getChargingModeText(effectiveListLessonChargingMode(record)) }}
                </div>
              </template>
              <template v-if="column.key === 'totalQuantity'">
                <div>
                  {{ Number(record.tuitionAccount?.totalQuantity || 0) + Number(record.tuitionAccount?.totalFreeQuantity || 0) }}{{ getQuantityUnit(effectiveListLessonChargingMode(record)) }}
                </div>
                <div class="text-3 text-#888">
                  购{{ record.tuitionAccount?.totalQuantity || 0 }}{{ getQuantityUnit(effectiveListLessonChargingMode(record)) }}
                  <span v-if="Number(record.tuitionAccount?.totalFreeQuantity || 0) > 0">
                    +赠{{ record.tuitionAccount?.totalFreeQuantity || 0 }}{{ getQuantityUnit(effectiveListLessonChargingMode(record)) }}
                  </span>
                </div>
              </template>
              <template v-if="column.key === 'expireTime'">
                {{ record.tuitionAccount?.enableExpireTime ? formatDate(record.tuitionAccount?.expireTime) : '-' }}
              </template>
              <template v-if="column.key === 'suspendedTime'">
                {{ formatDate(record.tuitionAccount?.suspendedTime) }}
              </template>
              <template v-if="column.key === 'classEndingTime'">
                {{ formatDate(record.tuitionAccount?.classEndingTime) }}
              </template>
              <template v-if="column.key === 'usedQuantity'">
                {{ calcUsedQuantity(record) }}{{ getQuantityUnit(effectiveListLessonChargingMode(record)) }}
              </template>
              <template v-if="column.key === 'remainQuantity'">
                {{ calcRemainQuantity(record) }}{{ getQuantityUnit(effectiveListLessonChargingMode(record)) }}
              </template>
              <template v-if="column.key === 'usedTuition'">
                {{ formatMoney(calcUsedTuition(record)) }}
              </template>
              <template v-if="column.key === 'remainTuition'">
                {{ formatMoney(record.tuitionAccount?.remainTuition) }}
              </template>
              <template v-if="column.key === 'totalTuition'">
                {{ formatMoney(record.tuitionAccount?.totalTuition) }}
              </template>
              <template v-if="column.key === 'classTeacher'">
                {{ record.classTeacherName || '-' }}
              </template>
              <template v-if="column.key === 'defaultTeacher'">
                {{ record.defaultTeacherName || '-' }}
              </template>
              <template v-if="column.key === 'classTime'">
                {{ formatClassTime(record) }}
              </template>
              <template v-if="column.key === 'lastFinishedLessonDay'">
                {{ shouldShowSchedulePlaceholder(record) ? '-' : formatDateTime(record.lastFinishedLessonDay) }}
              </template>
              <template v-if="column.key === 'isScheduled'">
                <span class="status-indicator" :class="record.isScheduled ? 'text-#0c3' : 'text-#666'">
                  <span class="status-dot" :class="record.isScheduled ? 'status-dot--success' : 'status-dot--warning'" />
                  {{ record.isScheduled ? '已排课' : '未排课' }}
                </span>
              </template>
              <template v-if="column.key === 'lessonDayCount'">
                {{ formatLessonDaySummary(record) }}
              </template>
              <template v-if="column.key === 'createdTime'">
                {{ formatDateTime(record.createdTime) }}
              </template>
              <template v-if="column.key === 'status'">
                <span :class="`${getOpenClassStatus(record.status).className} rounded-2.5 inline-block text-3 pt-0.5 pb-0.5 pl-2 pr-2`">
                  {{ getOpenClassStatus(record.status).text }}
                </span>
              </template>
              <template v-if="column.key === 'classStudentStatus'">
                <span :class="`${getClassStudentStatus(record.classStudentStatus).className} rounded-2.5 inline-block text-3 pt-0.5 pb-0.5 pl-2 pr-2`">
                  {{ getClassStudentStatus(record.classStudentStatus).text }}
                </span>
              </template>
              <template v-if="column.key === 'action'">
                <a-space :size="12">
                  <template v-if="isOneToOneClassClosed(record)">
                    <a @click="handleReopenClass(record)">恢复开班</a>
                    <a @click="openDrawer(record)">详情</a>
                  </template>
                  <template v-else>
                    <a @click="handleSchedule(record)">排课</a>
                    <a @click="handleFinishCourse(record)">结课</a>
                    <a @click="openEditModal(record)">编辑</a>
                    <a @click="openDrawer(record)">详情</a>
                  </template>
                </a-space>
              </template>
            </template>
          </a-table>
        </div>
      </div>
    </div>

    <one-to-one-drawer
      v-model:open="drawerOpen"
      :record="currentRecord"
      :tuition-accounts="drawerTuitionAccounts"
      @edit="handleDrawerEdit"
    />

    <FinishOneToOneCourseModal
      v-model:open="finishCourseModalOpen"
      title="结班并结课"
      :record="finishCourseRecord"
      @confirm="handleFinishCourseModalConfirm"
    />

    <a-modal v-model:open="advisorModalOpen" :title="advisorModalTitle" @ok="submitAdvisorBatch" :confirm-loading="advisorSubmitting">
      <a-form layout="vertical">
        <a-form-item label="班主任" required>
          <StaffSelect
            v-model="advisorForm.classTeacherIds"
            placeholder="请选择班主任（可多选）"
            width="100%"
            :status="0"
            :multiple="true"
          />
        </a-form-item>
      </a-form>
    </a-modal>

    <a-modal
      v-model:open="editModalOpen"
      centered
      class="createStu-modal-content-box"
      :keyboard="false"
      :closable="false"
      :mask-closable="false"
      :width="800"
      :confirm-loading="editSubmitting"
      ok-text="确定"
      cancel-text="取消"
      @ok="submitEditModal"
      @cancel="closeEditModal"
    >
      <template #title>
        <div class="text-5 flex justify-between flex-center">
          <span>编辑1对1</span>
          <a-button type="text" class="close-btn" @click="closeEditModal">
            <template #icon>
              <CloseOutlined class="text-5 close-icon" />
            </template>
          </a-button>
        </div>
      </template>
      <div class="stu-content scrollbar">
        <a-spin :spinning="editLoading">
          <a-form
            layout="horizontal"
            :label-col="{ span: 5 }"
            :wrapper-col="{ span: 16 }"
            label-align="right"
          >
          <a-form-item label="学员名称">
            <span>{{ editForm.studentName || '-' }}</span>
          </a-form-item>

          <a-form-item label="上课课程">
            <span>{{ editForm.lessonName || '-' }}</span>
          </a-form-item>

          <a-form-item label="1对1名称" required>
            <a-input v-model:value="editForm.name" :maxlength="100" placeholder="请输入1对1名称" style="width: 100%" />
          </a-form-item>

          <a-form-item label="班主任">
            <StaffSelect
              v-model="editForm.teacherIds"
              placeholder="请选择班主任"
              width="100%"
              :status="0"
              :multiple="true"
            />
          </a-form-item>

          <a-form-item>
            <template #label>
              <span>默认上课教师</span>
              <a-tooltip title="当课程未单独指定教师时，系统默认负责该班级日常教学的教师。">
                <QuestionCircleOutlined style="margin-left: 4px; color: #999" />
              </a-tooltip>
            </template>
            <StaffSelect
              v-model="editForm.defaultTeacherId"
              placeholder="请选择默认上课教师"
              width="100%"
              :status="0"
            />
          </a-form-item>

          <a-form-item label="上课教室">
            <a-select
              v-model:value="editForm.classRoomId"
              placeholder="请选择"
              style="width: 100%"
              :disabled="true"
              allow-clear
            >
              <a-select-option v-if="editForm.classRoomId" :value="editForm.classRoomId">
                {{ editForm.classRoomName }}
              </a-select-option>
            </a-select>
          </a-form-item>

          <a-form-item label="课时记录方式" required>
            <a-radio-group v-model:value="editForm.defaultClassTimeRecordMode" class="custom-radio">
              <a-radio :value="1">
                按固定课时记录
              </a-radio>
              <a-radio :value="2">
                按上课时长记录
              </a-radio>
            </a-radio-group>
          </a-form-item>

          <a-form-item label="默认记录课时" required>
            <div class="one-to-one-class-time-inputs">
              <span class="one-to-one-ct-group">
                <span>学员</span>
                <a-input-number v-model:value="editForm.defaultStudentClassTime" :min="0" :precision="2" style="width: 100px" />
                <span class="one-to-one-ct-unit">{{ editClassTimeUnitLabel }}</span>
              </span>
              <span class="one-to-one-ct-group">
                <span>上课教师课时</span>
                <a-input-number v-model:value="editForm.defaultTeacherClassTime" :min="0" :precision="2" style="width: 100px" />
                <span class="one-to-one-ct-unit">{{ editClassTimeUnitLabel }}</span>
              </span>
            </div>
            <div style="margin-top: 8px; color: #888; font-size: 13px;white-space: nowrap;">
              {{ editClassTimeHint }}
            </div>
          </a-form-item>

          <a-form-item label="备注">
            <a-textarea
              v-model:value="editForm.remark"
              :maxlength="150"
              :rows="1"
              placeholder="请输入"
              show-count
              style="width: 100%"
            />
          </a-form-item>
          </a-form>
        </a-spin>
      </div>
    </a-modal>

    <a-modal
      v-model:open="classTimeModalOpen"
      title="批量修改记录课时"
      width="640px"
      :confirm-loading="classTimeSubmitting"
      ok-text="确定"
      cancel-text="取消"
      @ok="submitClassTimeBatch"
    >
      <div v-if="classTimeBatchSelectionSummary.n" class="batch-class-time-summary">
        <div class="batch-class-time-summary-line">
          已选 <strong>{{ classTimeBatchSelectionSummary.n }}</strong> 个 <strong>1对1</strong> 记录课时
        </div>
        <div class="batch-class-time-summary-names">
          共 {{ classTimeBatchSelectionSummary.n }} 个，{{ classTimeBatchSelectionSummary.names || '—' }}
        </div>
      </div>
      <a-form layout="vertical" class="batch-class-time-form">
        <a-form-item label="课时记录方式" required>
          <a-radio-group v-model:value="classTimeForm.classTimeRecordMode" class="custom-radio">
            <a-radio :value="1">
              按固定课时记录
            </a-radio>
            <a-radio :value="2">
              按上课时长记录
            </a-radio>
          </a-radio-group>
        </a-form-item>
        <a-form-item>
          <template #label>
            <span><span class="batch-class-time-required">*</span> 默认记录学员</span>
          </template>
          <div class="one-to-one-class-time-inputs">
            <span class="one-to-one-ct-group">
              <a-input-number
                v-model:value="classTimeForm.studentClassTime"
                :min="0"
                :precision="2"
                style="width: 120px"
              />
              <span class="one-to-one-ct-unit">{{ classTimeBatchUnitLabel }}</span>
            </span>
            <span class="one-to-one-ct-group">
              <span class="one-to-one-ct-sep">，上课教师课时</span>
              <a-input-number
                v-model:value="classTimeForm.teacherClassTime"
                :min="0"
                :precision="2"
                style="width: 120px"
              />
              <span class="one-to-one-ct-unit">{{ classTimeBatchUnitLabel }}</span>
            </span>
          </div>
          <div class="batch-class-time-hint">
            {{ classTimeBatchHint }}
          </div>
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<style lang="less" scoped>
.total {
  position: relative;
  padding-left: 10px;
  color: #222;
  display: flex;
  align-items: center;

  &::before {
    display: inline-block;
    background: var(--pro-ant-color-primary);
    border-radius: 2px;
    content: "";
    height: 12px;
    left: 0;
    position: absolute;
    width: 4px;
  }
}

.status-indicator {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.status-dot {
  width: 6px;
  height: 6px;
  border-radius: 999px;
  display: inline-block;
}

.status-dot--warning {
  background: #fa8c16;
}

.status-dot--success {
  background: #52c41a;
}

@keyframes icon-rotate {
  from {
    transform: rotate(0deg);
  }

  to {
    transform: rotate(180deg);
  }
}

.createStu-modal-content-box {
  .stu-content {
    max-height: calc(100vh - 155px);
    padding: 24px 40px 0 !important;
    overflow: auto;
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

.batch-class-time-summary {
  margin-bottom: 16px;
  padding: 12px 14px;
  background: #e6f4ff;
  border-radius: 6px;
  font-size: 13px;
  line-height: 1.6;
  color: rgba(0, 0, 0, 0.85);
}

.batch-class-time-summary-line {
  margin-bottom: 4px;
}

.batch-class-time-summary-names {
  color: rgba(0, 0, 0, 0.65);
  word-break: break-all;
}

.batch-class-time-form {
  :deep(.ant-form-item) {
    margin-bottom: 16px;
  }
}

.batch-class-time-required {
  color: #ff4d4f;
  margin-right: 2px;
}

.batch-class-time-hint {
  margin-top: 8px;
  color: #888;
  font-size: 13px;
  line-height: 1.5;
}

/* 学员 + 教师两段同一行；组内输入与单位不换行。窄屏可横向滚动，避免第二段整体掉到下一行 */
.one-to-one-class-time-inputs {
  display: flex;
  flex-wrap: nowrap;
  align-items: center;
  column-gap: 12px;
  max-width: 100%;
  overflow-x: auto;
  padding-bottom: 2px;
}

.one-to-one-ct-group {
  display: inline-flex;
  align-items: center;
  flex-wrap: nowrap;
  flex-shrink: 0;
  gap: 8px;
}

.one-to-one-ct-unit,
.one-to-one-ct-sep {
  flex-shrink: 0;
  white-space: nowrap;
}
</style>

<style>
.createStu-modal-content-box .ant-modal-header {
  padding: 10px 16px !important;
  margin-bottom: 0;
}

.createStu-modal-content-box .ant-modal-body {
  padding: 0 !important;
}
</style>
