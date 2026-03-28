<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { debounce } from 'lodash-es'
import { DownOutlined } from '@ant-design/icons-vue'
import dayjs from 'dayjs'
import { useRouter } from 'vue-router'
import StaffSelect from '@/components/common/staff-select.vue'
import { useTableColumns } from '@/composables/useTableColumns'
import { handleDateRangeParams } from '@/utils/dateRangeParams'
import messageService from '@/utils/messageService'
import {
  batchAssignOneToOneClassTeacherApi,
  batchUpdateOneToOneAttributesApi,
  batchUpdateOneToOneClassTimeApi,
  getOneToOneListApi,
} from '@/api/edu-center/one-to-one'
import { Sex, SexLabel } from '@/enums'

const router = useRouter()
const allFilterRef = ref()
const loading = ref(false)
const dataSource = ref([])
const selectedRows = ref([])
const selectedRowKeys = ref([])
const currentRecord = ref(null)
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
  classTeacherId: undefined,
})

const classTimeModalOpen = ref(false)
const classTimeSubmitting = ref(false)
const classTimeForm = reactive({
  classTime: 1,
  studentClassTime: 1,
  teacherClassTime: 0,
})

const attributeModalOpen = ref(false)
const attributeSubmitting = ref(false)
const attributeForm = reactive({
  defaultTeacherId: undefined,
  status: undefined,
  classStudentStatus: undefined,
})

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
  status: undefined,
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
  { title: '操作', dataIndex: 'action', key: 'action', fixed: 'right', width: 90 },
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
        unassignedTeacherCount: dataSource.value.filter(item => !item.classTeacherId || item.classTeacherId === '0').length,
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

function getGenderText(sex) {
  return SexLabel[sex] || SexLabel[Sex.Unknown]
}

function formatDateTime(value) {
  if (!value || value === '0001-01-01T00:00:00')
    return '-'
  return dayjs(value).format('YYYY-MM-DD HH:mm')
}

function formatDate(value) {
  if (!value || value === '0001-01-01T00:00:00')
    return '-'
  return dayjs(value).format('YYYY-MM-DD')
}

function formatMoney(value) {
  return `¥ ${Number(value || 0).toLocaleString('zh-CN', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  })}`
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

function getOpenClassStatus(status) {
  if (status === 2)
    return { text: '已结班', className: 'text-#888 bg-#f5f5f5' }
  return { text: '开班中', className: 'text-#06f bg-#e6f0ff' }
}

function getClassStudentStatus(status) {
  if (status === 2)
    return { text: '已停课', className: 'text-#f90 bg-#fff5e6' }
  if (status === 3)
    return { text: '已结课', className: 'text-#888 bg-#f5f5f5' }
  return { text: '开课中', className: 'text-#0c3 bg-#e6ffec' }
}

function openDrawer(record) {
  currentRecord.value = record
  drawerOpen.value = true
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

  if (action === 'assign') {
    advisorModalTitle.value = '批量分配班主任'
    advisorForm.classTeacherId = undefined
    advisorModalOpen.value = true
    return
  }
  if (action === 'replace') {
    advisorModalTitle.value = '批量替换班主任'
    advisorForm.classTeacherId = undefined
    advisorModalOpen.value = true
    return
  }
  if (action === 'classTime') {
    const current = selectedRows.value[0]
    classTimeForm.classTime = Number(current?.classTime || 1)
    classTimeForm.studentClassTime = Number(current?.studentClassTime || 1)
    classTimeForm.teacherClassTime = Number(current?.teacherClassTime || 0)
    classTimeModalOpen.value = true
    return
  }
  if (action === 'attribute') {
    const current = selectedRows.value[0]
    attributeForm.defaultTeacherId = current?.defaultTeacherId && current.defaultTeacherId !== '0'
      ? Number(current.defaultTeacherId)
      : undefined
    attributeForm.status = current?.status
    attributeForm.classStudentStatus = current?.classStudentStatus
    attributeModalOpen.value = true
  }
}

async function submitAdvisorBatch() {
  if (!advisorForm.classTeacherId) {
    messageService.warning('请选择班主任')
    return
  }
  advisorSubmitting.value = true
  try {
    const res = await batchAssignOneToOneClassTeacherApi({
      ids: selectedRows.value.map(item => item.id),
      classTeacherId: String(advisorForm.classTeacherId),
    })
    if (res.code !== 200)
      throw new Error(res.message || '批量更新班主任失败')
    advisorModalOpen.value = false
    messageService.success(`${advisorModalTitle.value}成功`)
    await getOneToOneList()
  } catch (error) {
    console.error('batch assign advisor failed', error)
    messageService.error(error?.message || '批量更新班主任失败')
  } finally {
    advisorSubmitting.value = false
  }
}

async function submitClassTimeBatch() {
  classTimeSubmitting.value = true
  try {
    const res = await batchUpdateOneToOneClassTimeApi({
      ids: selectedRows.value.map(item => item.id),
      classTime: Number(classTimeForm.classTime || 0),
      studentClassTime: Number(classTimeForm.studentClassTime || 0),
      teacherClassTime: Number(classTimeForm.teacherClassTime || 0),
    })
    if (res.code !== 200)
      throw new Error(res.message || '修改记录课时失败')
    classTimeModalOpen.value = false
    messageService.success('修改记录课时成功')
    await getOneToOneList()
  } catch (error) {
    console.error('batch update class time failed', error)
    messageService.error(error?.message || '修改记录课时失败')
  } finally {
    classTimeSubmitting.value = false
  }
}

async function submitAttributeBatch() {
  if (!attributeForm.defaultTeacherId && !attributeForm.status && !attributeForm.classStudentStatus) {
    messageService.warning('请至少修改一项属性')
    return
  }
  attributeSubmitting.value = true
  try {
    const payload = {
      ids: selectedRows.value.map(item => item.id),
      status: attributeForm.status,
      classStudentStatus: attributeForm.classStudentStatus,
    }
    if (attributeForm.defaultTeacherId) {
      payload.defaultTeacherId = String(attributeForm.defaultTeacherId)
    }
    const res = await batchUpdateOneToOneAttributesApi(payload)
    if (res.code !== 200)
      throw new Error(res.message || '修改1对1属性失败')
    attributeModalOpen.value = false
    messageService.success('修改1对1属性成功')
    await getOneToOneList()
  } catch (error) {
    console.error('batch update attributes failed', error)
    messageService.error(error?.message || '修改1对1属性失败')
  } finally {
    attributeSubmitting.value = false
  }
}

function resetSelection() {
  selectedRows.value = []
  selectedRowKeys.value = []
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
                    修改记录课时
                  </a-menu-item>
                  <a-menu-item key="attribute">
                    修改1对1属性
                  </a-menu-item>
                </a-menu>
              </template>
              <a-button>
                批量操作
                <DownOutlined :style="{ fontSize: '10px' }" />
              </a-button>
            </a-dropdown>
            <a-button type="primary" class="mr-2 w-25" @click="handleEnroll">
              报名
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
                <div class="text-3 text-#888">
                  1对1授课
                </div>
              </template>
              <template v-if="column.key === 'account'">
                <div>{{ record.tuitionAccount?.productName || record.lessonName || '-' }}</div>
                <div class="text-3 text-#888">
                  {{ getChargingModeText(record.tuitionAccount?.lessonChargingMode) }}
                </div>
              </template>
              <template v-if="column.key === 'totalQuantity'">
                <div>
                  {{ Number(record.tuitionAccount?.totalQuantity || 0) + Number(record.tuitionAccount?.totalFreeQuantity || 0) }}{{ getQuantityUnit(record.tuitionAccount?.lessonChargingMode) }}
                </div>
                <div class="text-3 text-#888">
                  购{{ record.tuitionAccount?.totalQuantity || 0 }}{{ getQuantityUnit(record.tuitionAccount?.lessonChargingMode) }}
                  <span v-if="Number(record.tuitionAccount?.totalFreeQuantity || 0) > 0">
                    +赠{{ record.tuitionAccount?.totalFreeQuantity || 0 }}{{ getQuantityUnit(record.tuitionAccount?.lessonChargingMode) }}
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
                {{ calcUsedQuantity(record) }}{{ getQuantityUnit(record.tuitionAccount?.lessonChargingMode) }}
              </template>
              <template v-if="column.key === 'remainQuantity'">
                {{ calcRemainQuantity(record) }}{{ getQuantityUnit(record.tuitionAccount?.lessonChargingMode) }}
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
                {{ Number(record.classTime || 0) > 0 ? `${record.classTime}课时/次` : '-' }}
              </template>
              <template v-if="column.key === 'lastFinishedLessonDay'">
                {{ formatDateTime(record.lastFinishedLessonDay) }}
              </template>
              <template v-if="column.key === 'isScheduled'">
                <span :class="record.isScheduled ? 'text-#0c3' : 'text-#999'">
                  {{ record.isScheduled ? '已排课' : '未排课' }}
                </span>
              </template>
              <template v-if="column.key === 'lessonDayCount'">
                {{ record.one2OneLessonDayInfo?.completeLessonDayCount || 0 }}/{{ record.one2OneLessonDayInfo?.lessonDayCount || 0 }}节
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
                <a @click="openDrawer(record)">查看</a>
              </template>
            </template>
          </a-table>
        </div>
      </div>
    </div>

    <a-drawer
      v-model:open="drawerOpen"
      :closable="false"
      width="780px"
      placement="right"
      :body-style="{ background: '#f7f7fd' }"
    >
      <template #title>
        <div class="flex justify-between items-center">
          <span>1对1详情</span>
          <a-button type="text" @click="drawerOpen = false">
            关闭
          </a-button>
        </div>
      </template>

      <div v-if="currentRecord" class="bg-white rounded-4 p-6">
        <div class="flex items-center">
          <student-avatar
            :id="currentRecord.studentId"
            :name="currentRecord.studentName || '-'"
            :gender="getGenderText(currentRecord.sex)"
            :avatar-url="currentRecord.avatar"
            :show-age="false"
            default-active-key="0"
          />
          <div class="ml-4">
            <div class="text-5 font-600">
              {{ currentRecord.name || '-' }}
            </div>
            <div class="text-#888 mt-1">
              创建于 {{ formatDateTime(currentRecord.createdTime) }}
            </div>
          </div>
        </div>

        <a-descriptions class="mt-6" :column="2" size="small">
          <a-descriptions-item label="上课课程">
            {{ currentRecord.lessonName || '-' }}
          </a-descriptions-item>
          <a-descriptions-item label="当前课程账户">
            {{ currentRecord.tuitionAccount?.productName || '-' }}
          </a-descriptions-item>
          <a-descriptions-item label="班主任">
            {{ currentRecord.classTeacherName || '-' }}
          </a-descriptions-item>
          <a-descriptions-item label="默认上课教师">
            {{ currentRecord.defaultTeacherName || '-' }}
          </a-descriptions-item>
          <a-descriptions-item label="是否排课">
            {{ currentRecord.isScheduled ? '已排课' : '未排课' }}
          </a-descriptions-item>
          <a-descriptions-item label="已上/排课">
            {{ currentRecord.one2OneLessonDayInfo?.completeLessonDayCount || 0 }}/{{ currentRecord.one2OneLessonDayInfo?.lessonDayCount || 0 }}节
          </a-descriptions-item>
          <a-descriptions-item label="总学费">
            {{ formatMoney(currentRecord.tuitionAccount?.totalTuition) }}
          </a-descriptions-item>
          <a-descriptions-item label="剩余学费金额">
            {{ formatMoney(currentRecord.tuitionAccount?.remainTuition) }}
          </a-descriptions-item>
        </a-descriptions>
      </div>
    </a-drawer>

    <a-modal v-model:open="advisorModalOpen" :title="advisorModalTitle" @ok="submitAdvisorBatch" :confirm-loading="advisorSubmitting">
      <a-form layout="vertical">
        <a-form-item label="班主任" required>
          <StaffSelect v-model="advisorForm.classTeacherId" placeholder="请选择班主任" width="100%" :status="0" />
        </a-form-item>
      </a-form>
    </a-modal>

    <a-modal v-model:open="classTimeModalOpen" title="修改记录课时" @ok="submitClassTimeBatch" :confirm-loading="classTimeSubmitting">
      <a-form layout="vertical">
        <a-form-item label="上课时间">
          <a-input-number v-model:value="classTimeForm.classTime" :min="0" :precision="2" style="width: 100%" />
        </a-form-item>
        <a-form-item label="学员记录课时">
          <a-input-number v-model:value="classTimeForm.studentClassTime" :min="0" :precision="2" style="width: 100%" />
        </a-form-item>
        <a-form-item label="老师授课课时">
          <a-input-number v-model:value="classTimeForm.teacherClassTime" :min="0" :precision="2" style="width: 100%" />
        </a-form-item>
      </a-form>
    </a-modal>

    <a-modal v-model:open="attributeModalOpen" title="修改1对1属性" @ok="submitAttributeBatch" :confirm-loading="attributeSubmitting">
      <a-form layout="vertical">
        <a-form-item label="默认上课教师">
          <StaffSelect v-model="attributeForm.defaultTeacherId" placeholder="请选择默认上课教师" width="100%" :status="0" />
        </a-form-item>
        <a-form-item label="开班状态">
          <a-select v-model:value="attributeForm.status" allow-clear placeholder="请选择开班状态">
            <a-select-option :value="1">
              开班中
            </a-select-option>
            <a-select-option :value="2">
              已结班
            </a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item label="开课状态">
          <a-select v-model:value="attributeForm.classStudentStatus" allow-clear placeholder="请选择开课状态">
            <a-select-option :value="1">
              开课中
            </a-select-option>
            <a-select-option :value="2">
              已停课
            </a-select-option>
            <a-select-option :value="3">
              已结课
            </a-select-option>
          </a-select>
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
</style>
