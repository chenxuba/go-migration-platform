<script setup>
import { CaretDownOutlined, DownOutlined, ExclamationCircleOutlined } from '@ant-design/icons-vue'
import { Modal } from 'ant-design-vue'
import dayjs from 'dayjs'
import { debounce } from 'lodash-es'
import { createVNode, onMounted } from 'vue'
import { listClassroomsApi } from '@/api/business-settings/classroom'
import {
  batchAssignGroupClassTeacherApi,
  batchCloseGroupClassesApi,
  batchReplaceGroupClassTeacherApi,
  batchUpdateGroupClassClassTimeApi,
  batchUpdateGroupClassMaxCountApi,
  closeGroupClassApi,
  groupClassStatisticsApi,
  pageGroupClassesApi,
  reopenGroupClassApi,
} from '@/api/edu-center/group-class'
import CreateClassModal from '@/components/common/create-class-modal.vue'
import StaffSelect from '@/components/common/staff-select.vue'
import ClassAddStudentModal from '@/components/edu-center/class-list/class-add-student-modal.vue'
import ClassListDrawer from '@/components/edu-center/class-list/class-list-drawer.vue'
import GroupClassUnscheduledRollCallModal from '@/components/edu-center/class-list/group-class-unscheduled-roll-call-modal.vue'
import GroupClassFinishCourseModal from '@/components/edu-center/class-list/group-class-finish-course-modal.vue'
import GroupClassScheduleModal from '@/components/edu-center/timetable/group-class-schedule-modal.vue'
import { useTableColumns } from '@/composables/useTableColumns'
import { openCloseClassConfirm } from '@/utils/closeClassConfirm'
import messageService from '@/utils/messageService'

const defaultOpenClassStatus = 1

const allFilterRef = ref()
const createClassModal = ref(false)
const editClassRecord = ref(null)
const classListDrawerFlag = ref(false)
const currentClassRecord = ref(null)
const classListDrawerInitialTab = ref('0')
const addStudentModalOpen = ref(false)
const addStudentModalTitle = ref('')
const addStudentModalLessonName = ref('')
const addStudentModalClassId = ref('')
const addStudentModalLessonId = ref('')
const finishCourseModalOpen = ref(false)
const finishCourseRecord = ref(null)
const scheduleModalOpen = ref(false)
const scheduleModalClassId = ref('')
const unscheduledRollCallModalOpen = ref(false)
const unscheduledRollCallClassId = ref('')
const listLoading = ref(false)
const dataSource = ref([])
const selectedRowKeys = ref([])
const selectedRows = ref([])
const batchActionRows = ref([])

const batchTeacherModalOpen = ref(false)
const batchTeacherModalTitle = ref('批量分配班主任')
const batchTeacherAction = ref('assign')
const batchTeacherSubmitting = ref(false)
const batchTeacherForm = reactive({
  teacherIds: [],
})

const batchClassTimeModalOpen = ref(false)
const batchClassTimeSubmitting = ref(false)
const batchClassTimeForm = reactive({
  defaultClassTimeRecordMode: 1,
  defaultStudentClassTime: 1,
  defaultTeacherClassTime: 0,
})

const batchMaxCountModalOpen = ref(false)
const batchMaxCountSubmitting = ref(false)
const batchMaxCountForm = reactive({
  maxCount: undefined,
})

function syncBatchMaxCountInput(value) {
  if (value == null || value === '' || Number(value) <= 0)
    batchMaxCountForm.maxCount = undefined
  else
    batchMaxCountForm.maxCount = Number(value)
}

const displayArray = ref([
  'customSearch',
  'classTeacher',
  'createUser',
  'doYouSchedule',
  'openClassStatus',
  'classEndingTime',
  'createTime',
  'salesPerson',
])

const customSearchFilters = ref([
  {
    id: 'lessonKey',
    fieldKey: '关联课程',
    fieldType: 4,
    optionsList: [],
  },
  {
    id: 'classRoomName',
    fieldKey: '上课教室',
    fieldType: 4,
    optionsList: [],
  },
  {
    id: 'courseType',
    fieldKey: '关联课程类型',
    fieldType: 4,
    optionsList: [
      { id: 'single', value: '课程' },
      { id: 'compose', value: '组合课' },
    ],
  },
])

const stats = ref({
  classCount: 0,
  openClassCount: 0,
  studentCount: 0,
  studentPersonTime: 0,
})

const pagination = reactive({
  current: 1,
  pageSize: 20,
  total: 0,
  showSizeChanger: true,
  showTotal: total => `共 ${total} 条`,
})

const batchClassTimeUnitLabel = computed(() =>
  Number(batchClassTimeForm.defaultClassTimeRecordMode) === 2 ? '课时/小时' : '课时',
)

const batchClassTimeHint = computed(() =>
  Number(batchClassTimeForm.defaultClassTimeRecordMode) === 2
    ? '每次点名，学员和上课教师记录的课时会根据日程时长自动计算课时（点名时支持调整）'
    : '每次点名，学员和上课教师记录的课时数默认为此数值（点名时支持调整）',
)

const batchSelectionSummary = computed(() => {
  const rows = batchActionRows.value
  const count = rows.length
  const names = rows.map(item => item?.name).filter(Boolean).join('，')
  return { count, names }
})

const queryState = ref({
  classIds: undefined,
  lessonKey: undefined,
  teacherId: undefined,
  defaultTeacherId: undefined,
  classRoomName: undefined,
  courseType: undefined,
  isScheduled: undefined,
  statues: [defaultOpenClassStatus],
  className: undefined,
  createdStaffIds: undefined,
  createdTime: undefined,
  closedTime: undefined,
})

function resetQueryState() {
  Object.keys(queryState.value).forEach((key) => {
    queryState.value[key] = undefined
  })
}

function updateCustomSearchOptions(id, optionsList) {
  customSearchFilters.value = customSearchFilters.value.map(item => item.id === id
    ? { ...item, optionsList }
    : item)
}

async function loadClassroomFilterOptions() {
  try {
    const res = await listClassroomsApi()
    if (res.code !== 200) {
      messageService.error(res.message || '获取教室列表失败')
      return
    }
    const roomMap = new Map()
    for (const item of Array.isArray(res.result) ? res.result : []) {
      const roomName = String(item?.name || '').trim()
      if (!roomName || roomMap.has(roomName))
        continue
      roomMap.set(roomName, {
        id: roomName,
        value: roomName,
      })
    }
    if (roomMap.size > 0)
      updateCustomSearchOptions('classRoomName', Array.from(roomMap.values()))
  }
  catch (error) {
    console.error('load classroom filter options failed', error)
    messageService.error(error?.message || '获取教室列表失败')
  }
}

/**
 * 课程筛选项仍从当前列表结果合并；教室筛选项以机构教室列表为主，
 * 当前列表仅补充班级默认教室，避免页面内筛选项丢失。
 */
function mergeCustomFilterOptionsFromClassList(list) {
  if (!Array.isArray(list) || list.length === 0)
    return

  const lessonItem = customSearchFilters.value.find(f => f.id === 'lessonKey')
  const roomItem = customSearchFilters.value.find(f => f.id === 'classRoomName')
  const lessonMap = new Map((lessonItem?.optionsList || []).map(o => [o.id, o]))
  const roomMap = new Map((roomItem?.optionsList || []).map(o => [o.id, o]))

  for (const item of list) {
    const roomName = String(item.classRoomName || '').trim()
    if (roomName && !roomMap.has(roomName)) {
      roomMap.set(roomName, {
        id: roomName,
        value: roomName,
      })
    }
    const lid = String(item.lessonId ?? '').trim()
    if (lid) {
      const optId = item.isMultiProduct ? `compose:${lid}` : `single:${lid}`
      if (!lessonMap.has(optId)) {
        lessonMap.set(optId, {
          id: optId,
          value: item.lessonName || lid,
        })
      }
    }
  }

  updateCustomSearchOptions('lessonKey', Array.from(lessonMap.values()))
  updateCustomSearchOptions('classRoomName', Array.from(roomMap.values()))
}

const handleFilterUpdate = debounce((updates = {}, isClearAll = false, id, type) => {
  if (isClearAll) {
    resetQueryState()
  }
  else {
    Object.entries(updates).forEach(([key, value]) => {
      queryState.value[key] = value
    })
  }

  pagination.current = 1
  selectedRows.value = []
  selectedRowKeys.value = []
  batchActionRows.value = []
  getClassList(queryState.value, id, type)
}, 200, { leading: true, trailing: false })

const filterUpdateHandlers = computed(() => ({
  'update:customSearchInputFilter': (payload, isClearAll, id, type) => {
    if (isClearAll) {
      handleFilterUpdate({}, true, id, type)
      return
    }

    const fieldId = id || payload?.item?.id
    const value = payload?.value
    if (fieldId === 'lessonKey') {
      handleFilterUpdate({ lessonKey: value || undefined }, false, id, type)
      return
    }
    if (fieldId === 'classRoomName') {
      handleFilterUpdate({ classRoomName: value || undefined }, false, id, type)
      return
    }
    if (fieldId === 'courseType') {
      handleFilterUpdate({ courseType: value || undefined }, false, id, type)
    }
  },
  'update:classTeacherFilter': (val, isClearAll, id, type) => {
    handleFilterUpdate({ teacherId: val || undefined }, isClearAll, id, type)
  },
  'update:stuPhoneSearchFilter': (val, isClearAll, id, type) => {
    handleFilterUpdate({ classIds: val ? [String(val)] : undefined }, isClearAll, id, type)
  },
  'update:createUserFilter': (val, isClearAll, id, type) => {
    handleFilterUpdate({ defaultTeacherId: val || undefined }, isClearAll, id, type)
  },
  'update:salesPersonFilter': (val, isClearAll, id, type) => {
    handleFilterUpdate({ createdStaffIds: val ? [String(val)] : undefined }, isClearAll, id, type)
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
    handleFilterUpdate({ statues: val ? [val] : undefined }, isClearAll, id, type)
  },
  'update:createTimeFilter': (val, isClearAll, id, type) => {
    handleFilterUpdate({ createdTime: Array.isArray(val) && val.length ? val : undefined }, isClearAll, id, type)
  },
  'update:classEndingTimeFilter': (val, isClearAll, id, type) => {
    handleFilterUpdate({ closedTime: Array.isArray(val) && val.length ? val : undefined }, isClearAll, id, type)
  },
}))

async function handleClassSearch(searchParams) {
  try {
    const res = await pageGroupClassesApi({
      pageRequestModel: searchParams.pageRequestModel || {
        needTotal: true,
        pageSize: 10,
        pageIndex: 1,
        skipCount: 0,
      },
      queryModel: {
        className: searchParams.searchKey || undefined,
      },
    })
    if (res.code !== 200) {
      messageService.error(res.message || '搜索班级失败')
      return
    }
    const list = Array.isArray(res.result?.list) ? res.result.list : []
    allFilterRef.value?.updateStaffSearchData?.({
      result: list.map(item => ({
        id: item.id,
        name: item.name,
      })),
      total: Number(res.result?.total || 0),
    })
  }
  catch (error) {
    console.error('search class failed', error)
    messageService.error('搜索班级失败')
  }
}

function buildQueryModel(source = {}) {
  const queryModel = {}

  if (source.className) {
    queryModel.className = String(source.className).trim()
  }
  if (Array.isArray(source.classIds) && source.classIds.length > 0) {
    queryModel.classIds = source.classIds
  }
  if (source.teacherId) {
    queryModel.teacherId = String(source.teacherId)
  }
  if (source.defaultTeacherId) {
    queryModel.defaultTeacherId = String(source.defaultTeacherId)
  }
  if (source.classRoomName) {
    queryModel.classRoomName = String(source.classRoomName).trim()
  }
  if (Array.isArray(source.statues) && source.statues.length > 0) {
    queryModel.statues = source.statues
  }
  if (Array.isArray(source.createdStaffIds) && source.createdStaffIds.length > 0) {
    queryModel.createdStaffIds = source.createdStaffIds
  }
  if (typeof source.isScheduled === 'boolean') {
    queryModel.isScheduled = source.isScheduled
  }
  if (source.courseType === 'single') {
    queryModel.isMultiProduct = false
  }
  else if (source.courseType === 'compose') {
    queryModel.isMultiProduct = true
  }
  if (source.lessonKey) {
    const raw = String(source.lessonKey)
    const split = raw.split(':')
    queryModel.lessonIds = [split.length > 1 ? split[1] : raw]
  }
  if (Array.isArray(source.createdTime) && source.createdTime.length === 2) {
    queryModel.createdStartTime = source.createdTime[0]
    queryModel.createdEndTime = source.createdTime[1]
  }
  if (Array.isArray(source.closedTime) && source.closedTime.length === 2) {
    queryModel.closedStartDate = source.closedTime[0]
    queryModel.closedEndDate = source.closedTime[1]
  }

  return queryModel
}

async function getClassList(newQueryParams = {}, id, type) {
  listLoading.value = true
  try {
    const queryModel = buildQueryModel(newQueryParams)
    const [listRes, statsRes] = await Promise.all([
      pageGroupClassesApi({
        queryModel,
        pageRequestModel: {
          needTotal: true,
          pageSize: pagination.pageSize,
          pageIndex: pagination.current,
          skipCount: 0,
        },
      }),
      groupClassStatisticsApi(queryModel),
    ])

    if (listRes.code === 200 && listRes.result) {
      dataSource.value = Array.isArray(listRes.result.list) ? listRes.result.list : []
      pagination.total = Number(listRes.result.total || 0)
      mergeCustomFilterOptionsFromClassList(dataSource.value)
      allFilterRef.value?.clearQuickFilter?.(id, type)
    }
    else {
      dataSource.value = []
      pagination.total = 0
      messageService.error(listRes.message || '获取班级列表失败')
    }

    if (statsRes.code === 200 && statsRes.result) {
      stats.value = {
        classCount: Number(statsRes.result.classCount || 0),
        openClassCount: Number(statsRes.result.openClassCount || 0),
        studentCount: Number(statsRes.result.studentCount || 0),
        studentPersonTime: Number(statsRes.result.studentPersonTime || 0),
      }
    }
    else {
      stats.value = {
        classCount: 0,
        openClassCount: 0,
        studentCount: 0,
        studentPersonTime: 0,
      }
    }
  }
  catch (error) {
    console.error('get class list failed', error)
    dataSource.value = []
    pagination.total = 0
    stats.value = {
      classCount: 0,
      openClassCount: 0,
      studentCount: 0,
      studentPersonTime: 0,
    }
    messageService.error('获取班级列表失败')
  }
  finally {
    listLoading.value = false
  }
}

function onTableChange(pageInfo) {
  pagination.current = pageInfo.current
  pagination.pageSize = pageInfo.pageSize
  batchActionRows.value = []
  getClassList(queryState.value)
}

function formatDt(value) {
  if (value == null || value === '')
    return '-'
  const date = dayjs(value)
  return date.isValid() ? date.format('YYYY-MM-DD HH:mm') : '-'
}

function formatClosed(value) {
  if (value == null || value === '')
    return '-'
  const date = dayjs(value)
  if (!date.isValid() || date.year() < 1900)
    return '-'
  return date.format('YYYY-MM-DD')
}

function formatClassTime(times) {
  if (!Array.isArray(times) || times.length === 0)
    return '-'

  return times.map((item) => {
    const startTime = item?.startTime ? dayjs(item.startTime).format('MM-DD HH:mm') : ''
    const endTime = item?.endTime ? dayjs(item.endTime).format('HH:mm') : ''
    if (startTime && endTime)
      return `${startTime}-${endTime}`
    return item?.name || '日程'
  }).join('；')
}

function statusLabel(status) {
  if (status === 1)
    return '开班中'
  if (status === 2)
    return '已结班'
  return `状态${status}`
}

function teacherNames(teachers) {
  if (!Array.isArray(teachers) || teachers.length === 0)
    return '-'
  return teachers.map(item => item.name).filter(Boolean).join('、')
}

/** 学员数：有最大学员数时展示 N/maxCount，否则仅 N */
function formatStudentCountDisplay(record) {
  const n = Number(record?.studentCount)
  const max = Number(record?.maxCount)
  const safeN = Number.isFinite(n) ? n : 0
  if (Number.isFinite(max) && max > 0)
    return `${safeN}/${max}`
  return String(safeN)
}

function createClass() {
  editClassRecord.value = null
  createClassModal.value = true
}

function openClassListDrawer(record, initialActiveKey = '0') {
  currentClassRecord.value = record || null
  classListDrawerInitialTab.value = String(initialActiveKey || '0')
  classListDrawerFlag.value = true
}

function openAddStudentModal(record) {
  addStudentModalTitle.value = String(record?.name || '').trim() || '班级'
  addStudentModalLessonName.value = String(record?.lessonName || '').trim()
  addStudentModalClassId.value = String(record?.id ?? '').trim()
  addStudentModalLessonId.value = String(record?.lessonId ?? '').trim()
  addStudentModalOpen.value = true
}

function openScheduleModal(record) {
  const classId = String(record?.id || '').trim()
  if (!classId) {
    messageService.warning('当前班级信息不完整，暂不可排课')
    return
  }
  scheduleModalClassId.value = classId
  scheduleModalOpen.value = true
}

function getUnscheduledRollCallDisabledReason(record) {
  const classId = String(record?.id || '').trim()
  if (!classId)
    return '当前班级信息不完整，暂不可创建未排课点名'
  if (Number(record?.studentCount || 0) <= 0)
    return `${String(record?.name || '当前班级').trim() || '当前班级'}暂无在班学员，不能创建未排课点名`
  return ''
}

function openUnscheduledRollCallModal(record) {
  const disabledReason = getUnscheduledRollCallDisabledReason(record)
  if (disabledReason) {
    messageService.warning(disabledReason)
    return
  }
  const classId = String(record?.id || '').trim()
  unscheduledRollCallClassId.value = classId
  unscheduledRollCallModalOpen.value = true
}

function isUnscheduledRollCallDisabled(record) {
  return Boolean(getUnscheduledRollCallDisabledReason(record))
}

async function closeGroupClass(record) {
  const id = String(record?.id || '').trim()
  if (!id) {
    messageService.error('缺少班级ID')
    return Promise.reject(new Error('缺少班级ID'))
  }
  try {
    const res = await closeGroupClassApi({ id })
    if (res.code === 200) {
      messageService.success('结班成功')
      await getClassList(queryState.value)
      syncCurrentClassRecordFromList()
      return
    }
    messageService.error(res.message || '结班失败')
    return Promise.reject(new Error(res.message || '结班失败'))
  }
  catch (error) {
    console.error('close group class failed', error)
    messageService.error(error?.response?.data?.message || error?.message || '结班失败')
    return Promise.reject(error)
  }
}

function openGroupClassFinishCourseModal(record) {
  finishCourseRecord.value = record || null
  finishCourseModalOpen.value = true
}

async function reopenGroupClass(record) {
  const id = String(record?.id || '').trim()
  if (!id) {
    messageService.error('缺少班级ID')
    return
  }
  try {
    const res = await reopenGroupClassApi({ id })
    if (res.code === 200) {
      messageService.success('已恢复开班')
      await getClassList(queryState.value)
      syncCurrentClassRecordFromList()
      return
    }
    messageService.error(res.message || '恢复开班失败')
  }
  catch (error) {
    console.error('reopen group class failed', error)
    if (error?.response)
      return
    messageService.error('恢复开班失败')
  }
}

function openGroupClassCloseConfirm(record) {
  openCloseClassConfirm({
    async onOk() {
      await closeGroupClass(record)
      const latest = dataSource.value.find(item => String(item?.id || '').trim() === String(record?.id || '').trim()) || record
      openGroupClassFinishCourseModal(latest)
    },
    onCancel() {
      return closeGroupClass(record)
    },
  })
}

function openGroupClassReopenConfirm(record) {
  const id = String(record?.id || '').trim()
  if (!id) {
    messageService.error('缺少班级ID')
    return
  }
  Modal.confirm({
    title: '恢复开班',
    centered: true,
    icon: createVNode(ExclamationCircleOutlined),
    content: '确定将该班课恢复为开班中吗？',
    async onOk() {
      await reopenGroupClass(record)
    },
  })
}

function resetSelection() {
  selectedRows.value = []
  selectedRowKeys.value = []
  batchActionRows.value = []
}

function getCurrentBatchRows() {
  return batchActionRows.value.length > 0 ? batchActionRows.value : selectedRows.value
}

function collectBatchRows(actionLabel, { activeOnly = true } = {}) {
  if (selectedRows.value.length <= 0) {
    messageService.warning('请先选择班级')
    return []
  }

  const rows = activeOnly
    ? selectedRows.value.filter(item => Number(item?.status) === defaultOpenClassStatus)
    : [...selectedRows.value]

  if (rows.length <= 0) {
    messageService.warning(`请选择开班中的班级后再${actionLabel}`)
    return []
  }

  const ignoredCount = selectedRows.value.length - rows.length
  if (ignoredCount > 0 && activeOnly)
    messageService.warning(`已自动忽略 ${ignoredCount} 个已结班班级`)

  return rows
}

function openBatchCloseConfirm() {
  const rows = collectBatchRows('批量结班')
  if (rows.length <= 0)
    return
  batchActionRows.value = rows

  Modal.confirm({
    title: '批量结班',
    centered: true,
    icon: createVNode(ExclamationCircleOutlined),
    content: `已选 ${rows.length} 个班级。确认后将同步删除这些班级未点名的后续日程，且不可恢复，请谨慎操作。`,
    okText: '确定',
    cancelText: '取消',
    async onOk() {
      const currentRows = getCurrentBatchRows()
      try {
        const res = await batchCloseGroupClassesApi({
          ids: currentRows.map(item => String(item.id)),
        })
        if (res.code !== 200)
          throw new Error(res.message || '批量结班失败')
        messageService.success('批量结班成功')
        resetSelection()
        await getClassList(queryState.value)
        syncCurrentClassRecordFromList()
      }
      catch (error) {
        console.error('batch close group classes failed', error)
        messageService.error(error?.message || '批量结班失败')
      }
    },
  })
}

function openBatchAction(action) {
  if (action === 'close') {
    openBatchCloseConfirm()
    return
  }

  const actionLabelMap = {
    assign: '批量分配班主任',
    replace: '批量替换班主任',
    classTime: '批量修改记录课时',
    maxCount: '批量修改满班人数',
  }
  const rows = collectBatchRows(actionLabelMap[action] || '批量操作')
  if (rows.length <= 0)
    return
  batchActionRows.value = rows

  if (action === 'assign' || action === 'replace') {
    batchTeacherAction.value = action
    batchTeacherModalTitle.value = action === 'assign' ? '批量分配班主任' : '批量替换班主任'
    batchTeacherForm.teacherIds = []
    batchTeacherModalOpen.value = true
    return
  }

  if (action === 'classTime') {
    const current = rows[0]
    batchClassTimeForm.defaultClassTimeRecordMode = Number(current?.defaultClassTimeRecordMode || 1)
    batchClassTimeForm.defaultStudentClassTime = Number(current?.defaultStudentClassTime ?? 1) || 1
    batchClassTimeForm.defaultTeacherClassTime = Number(current?.defaultTeacherClassTime ?? 0) || 0
    batchClassTimeModalOpen.value = true
    return
  }

  if (action === 'maxCount') {
    const current = rows[0]
    syncBatchMaxCountInput(current?.maxCount)
    batchMaxCountModalOpen.value = true
  }
}

async function submitBatchTeacher() {
  const teacherIds = Array.isArray(batchTeacherForm.teacherIds)
    ? batchTeacherForm.teacherIds.filter(id => id !== undefined && id !== null && `${id}` !== '')
    : []
  if (teacherIds.length === 0) {
    messageService.warning('请选择班主任')
    return
  }

  const currentRows = getCurrentBatchRows()
  batchTeacherSubmitting.value = true
  try {
    const api = batchTeacherAction.value === 'replace'
      ? batchReplaceGroupClassTeacherApi
      : batchAssignGroupClassTeacherApi
    const res = await api({
      ids: currentRows.map(item => String(item.id)),
      teacherIds: teacherIds.map(id => String(id)),
    })
    if (res.code !== 200)
      throw new Error(res.message || '批量更新班主任失败')
    batchTeacherModalOpen.value = false
    messageService.success(`${batchTeacherModalTitle.value}成功`)
    resetSelection()
    await getClassList(queryState.value)
    syncCurrentClassRecordFromList()
  }
  catch (error) {
    console.error('batch update group class teachers failed', error)
    messageService.error(error?.message || '批量更新班主任失败')
  }
  finally {
    batchTeacherSubmitting.value = false
  }
}

async function submitBatchClassTime() {
  const currentRows = getCurrentBatchRows()
  batchClassTimeSubmitting.value = true
  try {
    const res = await batchUpdateGroupClassClassTimeApi({
      ids: currentRows.map(item => String(item.id)),
      defaultStudentClassTime: Number(batchClassTimeForm.defaultStudentClassTime || 0),
      defaultTeacherClassTime: Number(batchClassTimeForm.defaultTeacherClassTime || 0),
      defaultClassTimeRecordMode: Number(batchClassTimeForm.defaultClassTimeRecordMode || 1),
    })
    if (res.code !== 200)
      throw new Error(res.message || '批量修改记录课时失败')
    batchClassTimeModalOpen.value = false
    messageService.success('批量修改记录课时成功')
    resetSelection()
    await getClassList(queryState.value)
    syncCurrentClassRecordFromList()
  }
  catch (error) {
    console.error('batch update group class class time failed', error)
    messageService.error(error?.message || '批量修改记录课时失败')
  }
  finally {
    batchClassTimeSubmitting.value = false
  }
}

async function submitBatchMaxCount() {
  const currentRows = getCurrentBatchRows()
  batchMaxCountSubmitting.value = true
  try {
    const res = await batchUpdateGroupClassMaxCountApi({
      ids: currentRows.map(item => String(item.id)),
      maxCount: Number(batchMaxCountForm.maxCount || 0),
    })
    if (res.code !== 200)
      throw new Error(res.message || '批量修改满班人数失败')
    batchMaxCountModalOpen.value = false
    messageService.success('批量修改满班人数成功')
    resetSelection()
    await getClassList(queryState.value)
    syncCurrentClassRecordFromList()
  }
  catch (error) {
    console.error('batch update group class max count failed', error)
    messageService.error(error?.message || '批量修改满班人数失败')
  }
  finally {
    batchMaxCountSubmitting.value = false
  }
}

function onClassRowMenuClick({ key }, record) {
  if (key === '1') {
    openClassListDrawer(record, '2')
    return
  }
  if (key === '2') {
    openUnscheduledRollCallModal(record)
    return
  }
  if (key === '3') {
    editClassRecord.value = record
    createClassModal.value = true
    return
  }
  if (key === '4') {
    openGroupClassCloseConfirm(record)
    return
  }
  console.log(key, record)
}

function syncCurrentClassRecordFromList() {
  const currentId = String(currentClassRecord.value?.id || '').trim()
  if (!currentId)
    return
  const latest = dataSource.value.find(item => String(item?.id || '').trim() === currentId)
  if (latest)
    currentClassRecord.value = latest
}

async function afterClassModalSave() {
  await getClassList(queryState.value)
  syncCurrentClassRecordFromList()
}

function handleDrawerEdit(record) {
  editClassRecord.value = record
  createClassModal.value = true
}

async function handleDrawerRefresh() {
  await getClassList(queryState.value)
  syncCurrentClassRecordFromList()
}

function handleDrawerFinishCourse(record) {
  openGroupClassFinishCourseModal(record)
}

async function handleFinishCourseSuccess() {
  finishCourseModalOpen.value = false
  await getClassList(queryState.value)
  syncCurrentClassRecordFromList()
}

watch(createClassModal, (open) => {
  if (!open)
    editClassRecord.value = null
})

watch(scheduleModalOpen, (open) => {
  if (!open)
    scheduleModalClassId.value = ''
})

const allColumns = ref([
  {
    title: '班级名称',
    dataIndex: 'name',
    key: 'name',
    fixed: 'left',
    width: 200,
    required: true,
  },
  {
    title: '关联课程',
    key: 'linkCourse',
    dataIndex: 'linkCourse',
    width: 160,
  },
  {
    title: '学员数',
    key: 'studentNum',
    dataIndex: 'studentNum',
    width: 110,
  },
  {
    title: '班主任',
    key: 'headTeacher',
    dataIndex: 'headTeacher',
    width: 140,
  },
  {
    title: '默认上课教师',
    key: 'defaultTeacher',
    dataIndex: 'defaultTeacher',
    width: 140,
  },
  {
    title: '上课教室',
    dataIndex: 'classRoom',
    key: 'classRoom',
    width: 140,
  },
  {
    title: '上课时间',
    dataIndex: 'classTime',
    key: 'classTime',
    width: 220,
  },
  {
    title: '是否排课',
    dataIndex: 'doYouSchedule',
    key: 'doYouSchedule',
    width: 120,
  },
  {
    title: '已上/日程总数',
    dataIndex: 'alreadyOnOrtotal',
    key: 'alreadyOnOrtotal',
    width: 150,
  },
  {
    title: '状态',
    dataIndex: 'openClassStatus',
    key: 'openClassStatus',
    width: 120,
  },
  {
    title: '创建时间',
    dataIndex: 'createTime',
    key: 'createTime',
    width: 180,
  },
  {
    title: '创建人',
    key: 'createUser',
    dataIndex: 'createUser',
    width: 120,
  },
  {
    title: '备注',
    key: 'remark',
    dataIndex: 'remark',
    width: 160,
  },
  {
    title: '结班日期',
    key: 'classEndingTime',
    dataIndex: 'classEndingTime',
    width: 120,
  },
  {
    title: '操作',
    dataIndex: 'action',
    key: 'action',
    fixed: 'right',
    width: 220,
  },
])

const { selectedValues, columnOptions, filteredColumns, totalWidth }
  = useTableColumns({
    storageKey: 'class-list',
    allColumns,
    excludeKeys: ['action'],
  })

const rowSelection = computed(() => ({
  selectedRowKeys: selectedRowKeys.value,
  onChange: (keys, rows) => {
    selectedRowKeys.value = keys
    selectedRows.value = rows
    batchActionRows.value = []
  },
}))

onMounted(async () => {
  await Promise.all([
    loadClassroomFilterOptions(),
    getClassList(queryState.value),
  ])
})
</script>

<template>
  <div>
    <div class="filter-wrap bg-white pl-3 pr-3 rounded-4">
      <all-filter
        ref="allFilterRef"
        :display-array="displayArray"
        :is-quick-show="false"
        :is-show-clsss-or-course-search="true"
        :custom-is-display-list="customSearchFilters"
        search-label="班级名称"
        search-placeholder="请选择班级名称"
        create-user-label="默认上课教师"
        create-user-placeholder="请输入默认上课教师"
        sales-person-label="创建人"
        sales-person-placeholder="请输入创建人"
        :default-open-class-status="defaultOpenClassStatus"
        @staff-search="handleClassSearch"
        v-on="filterUpdateHandlers"
      />
    </div>

    <div class="student-list mt-2 pt-3 pb-3 pl-6 pr-6 bg-white rounded-4">
      <div class="tab-table">
        <div class="table-title flex justify-between">
          <div class="total">
            总计 {{ stats.classCount }} 个班级，{{ stats.openClassCount }} 个开班中，在读学员 {{ stats.studentCount }} 人，在读人次 {{ stats.studentPersonTime }} 人
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
                  <a-menu-item key="close">
                    批量结班
                  </a-menu-item>
                  <a-menu-item key="assign">
                    批量分配班主任
                  </a-menu-item>
                  <a-menu-item key="replace">
                    批量替换班主任
                  </a-menu-item>
                  <a-menu-item key="classTime">
                    批量修改记录课时
                  </a-menu-item>
                  <a-menu-item key="maxCount">
                    批量修改满班人数
                  </a-menu-item>
                </a-menu>
              </template>
              <a-button>
                批量操作
                <DownOutlined :style="{ fontSize: '10px' }" />
              </a-button>
            </a-dropdown>
            <a-dropdown class="mr-2">
              <template #overlay>
                <a-menu>
                  <a-menu-item key="0">
                    导入班级
                  </a-menu-item>
                  <a-menu-item key="1">
                    批量导出
                  </a-menu-item>
                  <a-menu-item key="2">
                    导出记录
                  </a-menu-item>
                </a-menu>
              </template>
              <a-button>
                导出数据
                <DownOutlined :style="{ fontSize: '10px' }" />
              </a-button>
            </a-dropdown>
            <a-button type="primary" class="mr-2" @click="createClass">
              创建班级
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
            :data-source="dataSource"
            :loading="listLoading"
            :pagination="pagination"
            :columns="filteredColumns"
            :row-selection="rowSelection"
            :scroll="{ x: totalWidth }"
            row-key="id"
            size="small"
            @change="onTableChange"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'name'">
                <a-button type="link" @click="openClassListDrawer(record)">
                  {{ record.name || '-' }}
                </a-button>
              </template>
              <template v-else-if="column.key === 'linkCourse'">
                <div class="text-#222">
                  {{ record.lessonName || '-' }}
                </div>
                <div class="text-3 text-#888 flex flex-items-center">
                  {{ record.isMultiProduct ? '组合课程' : '课程' }}
                </div>
              </template>
              <template v-else-if="column.key === 'studentNum'">
                {{ formatStudentCountDisplay(record) }}
              </template>
              <template v-else-if="column.key === 'headTeacher'">
                {{ teacherNames(record.teachers) }}
              </template>
              <template v-else-if="column.key === 'defaultTeacher'">
                {{ record.defaultTeacherName || '-' }}
              </template>
              <template v-else-if="column.key === 'classRoom'">
                {{ record.classRoomName || '-' }}
              </template>
              <template v-else-if="column.key === 'classTime'">
                {{ formatClassTime(record.classLessonTimes) }}
              </template>
              <template v-else-if="column.key === 'doYouSchedule'">
                <div class="studentStatus">
                  <span class="dot" />
                  <span>{{ record.isScheduled ? '已排课' : '未排课' }}</span>
                </div>
              </template>
              <template v-else-if="column.key === 'alreadyOnOrtotal'">
                {{ record.classLessonDayInfos?.completeLessonDayCount ?? 0 }}/{{ record.classLessonDayInfos?.lessonDayCount ?? 0 }}
              </template>
              <template v-else-if="column.key === 'openClassStatus'">
                <div
                  class="rounded-2.5 inline-block text-3 pt-0.5 pb-0.5 pl-2 pr-2"
                  :class="record.status === 1 ? 'text-#06f bg-#e6f0ff' : 'text-#666 bg-#f5f5f5'"
                >
                  {{ statusLabel(record.status) }}
                </div>
              </template>
              <template v-else-if="column.key === 'createTime'">
                {{ formatDt(record.createdTime) }}
              </template>
              <template v-else-if="column.key === 'createUser'">
                {{ record.createdStaffName || '-' }}
              </template>
              <template v-else-if="column.key === 'remark'">
                {{ record.remark || '-' }}
              </template>
              <template v-else-if="column.key === 'classEndingTime'">
                {{ formatClosed(record.closedTime) }}
              </template>
              <template v-else-if="column.key === 'action'">
                <span class="flex action">
                  <template v-if="record.status === 2">
                    <a @click.prevent="openGroupClassReopenConfirm(record)">恢复开班</a>
                  </template>
                  <template v-else>
                    <a class="mr-3" @click.prevent="openScheduleModal(record)">排课</a>
                    <a class="mr-3" @click.prevent="openAddStudentModal(record)">添加学员</a>
                    <div style="cursor: pointer;">
                      <a-dropdown :trigger="['hover']" placement="bottom">
                        <a @click.prevent>
                          <div class="intention">
                            更多
                            <CaretDownOutlined
                              class="text-#1677ff"
                              :style="{ fontSize: '12px' }"
                            />
                          </div>
                        </a>
                        <template #overlay>
                          <a-menu style="text-align: center;width: 120px;" @click="(e) => onClassRowMenuClick(e, record)">
                            <a-menu-item key="1">
                              上课点名
                            </a-menu-item>
                            <a-menu-item key="2" :class="{ 'menu-item-disabled': isUnscheduledRollCallDisabled(record) }">
                              <a-tooltip :title="getUnscheduledRollCallDisabledReason(record) || undefined" placement="left">
                                <span class="menu-item-label">未排课点名</span>
                              </a-tooltip>
                            </a-menu-item>
                            <a-menu-item key="3">
                              编辑班级
                            </a-menu-item>
                            <a-menu-item key="4" danger>
                              结班
                            </a-menu-item>
                          </a-menu>
                        </template>
                      </a-dropdown>
                    </div>
                  </template>
                </span>
              </template>
            </template>
          </a-table>
        </div>
      </div>
    </div>
    <CreateClassModal
      v-model:open="createClassModal"
      :edit-record="editClassRecord"
      @created="afterClassModalSave"
      @updated="afterClassModalSave"
    />
    <ClassListDrawer
      v-model:open="classListDrawerFlag"
      :initial-active-key="classListDrawerInitialTab"
      :record="currentClassRecord"
      @edit="handleDrawerEdit"
      @finish-course="handleDrawerFinishCourse"
      @refresh="handleDrawerRefresh"
    />
    <GroupClassFinishCourseModal
      v-model:open="finishCourseModalOpen"
      :record="finishCourseRecord"
      @success="handleFinishCourseSuccess"
    />
    <ClassAddStudentModal
      v-model:open="addStudentModalOpen"
      :title="addStudentModalTitle"
      :lesson-name="addStudentModalLessonName"
      :class-id="addStudentModalClassId"
      :lesson-id="addStudentModalLessonId"
      @success="afterClassModalSave"
    />
    <GroupClassScheduleModal
      v-model:open="scheduleModalOpen"
      :initial-group-class-id="scheduleModalClassId"
      @updated="afterClassModalSave"
    />
    <GroupClassUnscheduledRollCallModal
      v-model:open="unscheduledRollCallModalOpen"
      :class-id="unscheduledRollCallClassId"
      @updated="afterClassModalSave"
      @confirmed="afterClassModalSave"
    />
    <a-modal
      v-model:open="batchTeacherModalOpen"
      :title="batchTeacherModalTitle"
      :confirm-loading="batchTeacherSubmitting"
      ok-text="确定"
      cancel-text="取消"
      @ok="submitBatchTeacher"
    >
      <div v-if="batchSelectionSummary.count" class="batch-class-summary">
        <div class="batch-class-summary-line">
          已选 <strong>{{ batchSelectionSummary.count }}</strong> 个班级
        </div>
        <div class="batch-class-summary-names">
          {{ batchSelectionSummary.names || '—' }}
        </div>
      </div>
      <a-form layout="vertical">
        <a-form-item label="班主任" required>
          <StaffSelect
            v-model="batchTeacherForm.teacherIds"
            placeholder="请选择班主任（可多选）"
            width="100%"
            :status="0"
            :multiple="true"
          />
        </a-form-item>
      </a-form>
    </a-modal>
    <a-modal
      v-model:open="batchClassTimeModalOpen"
      title="批量修改记录课时"
      width="640px"
      :confirm-loading="batchClassTimeSubmitting"
      ok-text="确定"
      cancel-text="取消"
      @ok="submitBatchClassTime"
    >
      <div v-if="batchSelectionSummary.count" class="batch-class-summary">
        <div class="batch-class-summary-line">
          已选 <strong>{{ batchSelectionSummary.count }}</strong> 个班级
        </div>
        <div class="batch-class-summary-names">
          {{ batchSelectionSummary.names || '—' }}
        </div>
      </div>
      <a-form layout="vertical" class="batch-class-time-form">
        <a-form-item label="课时记录方式" required>
          <a-radio-group v-model:value="batchClassTimeForm.defaultClassTimeRecordMode" class="custom-radio">
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
            <span><span class="batch-class-required">*</span> 默认记录学员</span>
          </template>
          <div class="one-to-one-class-time-inputs">
            <span class="one-to-one-ct-group">
              <a-input-number
                v-model:value="batchClassTimeForm.defaultStudentClassTime"
                :min="0"
                :precision="2"
                style="width: 120px"
              />
              <span class="one-to-one-ct-unit">{{ batchClassTimeUnitLabel }}</span>
            </span>
            <span class="one-to-one-ct-group">
              <span class="one-to-one-ct-sep">，上课教师课时</span>
              <a-input-number
                v-model:value="batchClassTimeForm.defaultTeacherClassTime"
                :min="0"
                :precision="2"
                style="width: 120px"
              />
              <span class="one-to-one-ct-unit">{{ batchClassTimeUnitLabel }}</span>
            </span>
          </div>
          <div class="batch-class-time-hint">
            {{ batchClassTimeHint }}
          </div>
        </a-form-item>
      </a-form>
    </a-modal>
    <a-modal
      v-model:open="batchMaxCountModalOpen"
      title="批量修改满班人数"
      :confirm-loading="batchMaxCountSubmitting"
      ok-text="确定"
      cancel-text="取消"
      @ok="submitBatchMaxCount"
    >
      <div v-if="batchSelectionSummary.count" class="batch-class-summary">
        <div class="batch-class-summary-line">
          已选 <strong>{{ batchSelectionSummary.count }}</strong> 个班级
        </div>
        <div class="batch-class-summary-names">
          {{ batchSelectionSummary.names || '—' }}
        </div>
      </div>
      <a-form layout="vertical">
        <a-form-item label="满班人数">
          <a-input-number
            :value="batchMaxCountForm.maxCount"
            :min="0"
            :precision="0"
            placeholder="不限"
            style="width: 160px"
            @update:value="syncBatchMaxCountInput"
          />
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
    content: '';
    height: 12px;
    left: 0;
    position: absolute;
    width: 4px;
  }
}

.studentStatus {
  display: flex;
  align-items: center;

  span.dot {
    border-radius: 50%;
    display: inline-block;
    height: 6px;
    position: relative;
    vertical-align: middle;
    width: 6px;
    margin-right: 4px;
    background: var(--pro-ant-color-primary);
  }
}

.menu-item-label {
  display: inline-block;
  width: 100%;
}

.menu-item-disabled {
  :deep(.ant-dropdown-menu-title-content) {
    color: rgba(0, 0, 0, 0.25);
    cursor: not-allowed;
  }
}

.batch-class-summary {
  margin-bottom: 16px;
  padding: 12px 14px;
  border-radius: 8px;
  background: #f8fbff;
  border: 1px solid #e6f0ff;
}

.batch-class-summary-line {
  color: #1f1f1f;
  font-weight: 500;
}

.batch-class-summary-names {
  margin-top: 6px;
  color: #666;
  font-size: 13px;
  line-height: 1.6;
  word-break: break-all;
}

.batch-class-required {
  color: #ff4d4f;
  margin-right: 2px;
}

.one-to-one-class-time-inputs {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
}

.one-to-one-ct-group {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  color: #262626;
}

.one-to-one-ct-unit {
  color: #8c8c8c;
}

.one-to-one-ct-sep {
  color: #595959;
}

.batch-class-time-hint {
  margin-top: 8px;
  color: #8c8c8c;
  font-size: 13px;
}
</style>
