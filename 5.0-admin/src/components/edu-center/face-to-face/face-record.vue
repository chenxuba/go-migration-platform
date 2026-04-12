<script setup lang="ts">
import dayjs from 'dayjs'
import { pageFaceAttendanceRecordsApi, type FaceAttendanceRecordItem, type FaceAttendanceRelatedScheduleItem } from '@/api/edu-center/face'
import faceIcon from '@/assets/images/face.png'
import StudentAvatar from '@/components/common/StudentAvatar.vue'
import RollCallDrawer from '@/components/common/roll-call-drawer.vue'
import { useTableColumns } from '@/composables/useTableColumns'

const displayArray = ref(['scheduleType', 'createTime', 'lastEditedTime', 'scheduleCallStatus'])
const dataSource = ref<FaceAttendanceRecordItem[]>([])
const loading = ref(false)
const filterStudentId = ref<string | undefined>(undefined)
const filterSignInTime = ref<string[]>([])
const filterSignOutTime = ref<string[]>([])
const filterActionTypes = ref<string[]>([])
const filterPendingSignOut = ref<string | undefined>(undefined)
const pagination = ref({
  current: 1,
  pageSize: 20,
  total: 0,
})
const selectedRowKeys = ref<string[]>([])
const scheduleModalOpen = ref(false)
const currentScheduleRecord = ref<FaceAttendanceRecordItem | null>(null)
const rollCallDrawerOpen = ref(false)
const currentRollCallScheduleId = ref('')
const currentRollCallLessonDay = ref('')
const previewVisible = ref(false)
const previewImage = ref('')
const previewTitle = ref('')
const previewTimeText = ref('')
const attendanceActionOptions = [
  { id: 'sign_in', value: '自动签到' },
  { id: 'sign_out', value: '自动签退' },
]
const pendingSignOutOptions = [
  { id: '1', value: '含待签退的数据' },
  { id: '0', value: '不含待签退的数据' },
]

const allColumns = ref([
  {
    title: '学员姓名',
    dataIndex: 'name',
    key: 'name',
    fixed: 'left',
    width: 130,
    required: true,
  },
  {
    title: '人脸采集',
    dataIndex: 'face',
    key: 'face',
    width: 120,
  },
  {
    title: '考勤类型',
    dataIndex: 'faceType',
    key: 'faceType',
    width: 120,
  },
  {
    title: '签到/签退类型',
    dataIndex: 'signInOutType',
    key: 'signInOutType',
    width: 160,
  },
  {
    title: '签到/签退时间',
    dataIndex: 'signInOutTime',
    key: 'signInOutTime',
    width: 160,
  },
  {
    title: '排课计划',
    dataIndex: 'schedulePlan',
    key: 'schedulePlan',
    width: 120,
  },
  {
    title: '相关日程',
    dataIndex: 'linkSchedule',
    key: 'linkSchedule',
    width: 160,
  },
  {
    title: '提示',
    dataIndex: 'remarks',
    key: 'remarks',
    width: 160,
  },
])

const { selectedValues, columnOptions, filteredColumns, totalWidth }
  = useTableColumns({
    storageKey: 'face-record',
    allColumns,
    excludeKeys: ['action'],
  })

const tablePagination = computed(() => ({
  current: pagination.value.current,
  pageSize: pagination.value.pageSize,
  total: pagination.value.total,
  showSizeChanger: true,
  showQuickJumper: true,
  showTotal: (total: number) => `共 ${total} 条`,
}))

const relatedScheduleColumns = [
  {
    title: '上课时间',
    dataIndex: 'classTime',
    key: 'classTime',
    width: 240,
  },
  {
    title: '相关日程',
    dataIndex: 'scheduleName',
    key: 'scheduleName',
    width: 320,
  },
  {
    title: '点名状态',
    dataIndex: 'rollCallStatus',
    key: 'rollCallStatus',
    width: 160,
  },
  {
    title: '操作',
    dataIndex: 'action',
    key: 'action',
    width: 110,
  },
]

const relatedScheduleDataSource = computed<FaceAttendanceRelatedScheduleItem[]>(() => {
  const record = currentScheduleRecord.value
  if (!record)
    return []
  return getRelatedScheduleItems(record)
})

const rowSelection = computed(() => ({
  selectedRowKeys: selectedRowKeys.value,
  onChange: (keys: (string | number)[]) => {
    selectedRowKeys.value = keys.map(key => String(key))
  },
}))

function formatDateTime(value?: string) {
  if (!value)
    return '-'
  const parsed = dayjs(value)
  return parsed.isValid() ? parsed.format('YYYY-MM-DD HH:mm') : '-'
}

function isPendingSignOut(record: Partial<FaceAttendanceRecordItem> | Record<string, any>) {
  return record.action === 'sign_in' && Number(record.sessionStatus) === 1 && !record.signOutTime
}

function getAttendancePhoto(record: Partial<FaceAttendanceRecordItem> | Record<string, any>, action = String(record.action || '')) {
  if (action === 'sign_out')
    return String(record.signOutImage || '').trim()
  if (action === 'sign_in')
    return String(record.signInImage || '').trim()
  return ''
}

function getAttendancePhotoTitle(action = '') {
  return action === 'sign_out' ? '签退照片' : '签到照片'
}

function openAttendancePhoto(record: Partial<FaceAttendanceRecordItem> | Record<string, any>, action = String(record.action || '')) {
  const image = getAttendancePhoto(record, action)
  if (!image)
    return
  previewImage.value = image
  previewTitle.value = getAttendancePhotoTitle(action)
  previewTimeText.value = formatDateTime(String(action === 'sign_out' ? record.signOutTime : record.attendanceTime || ''))
  previewVisible.value = true
}

function getAttendancePrimaryActionLabel(record: Partial<FaceAttendanceRecordItem> | Record<string, any>) {
  if (record.attendanceTime)
    return '自动签到'
  return record.actionLabel || '-'
}

function getAttendanceSecondaryActionLabel(record: Partial<FaceAttendanceRecordItem> | Record<string, any>) {
  if (record.signOutTime)
    return '自动签退'
  if (isPendingSignOut(record))
    return '待签退'
  return ''
}

function sexText(value?: number) {
  if (Number(value) === 1)
    return '男'
  if (Number(value) === 2)
    return '女'
  return '-'
}

function normalizeStringArray(value?: string[]) {
  if (!Array.isArray(value))
    return []
  return value.map(item => String(item || '').trim()).filter(Boolean)
}

function splitBlockLines(value: string) {
  return String(value || '').split('\n').map(item => item.trim()).filter(Boolean)
}

function formatSingleLineClassTime(value: string) {
  return splitBlockLines(value).join(' ')
}

function getRelatedScheduleItems(record?: Partial<FaceAttendanceRecordItem> | Record<string, any>) {
  if (!record)
    return []
  if (Array.isArray(record.relatedScheduleItems) && record.relatedScheduleItems.length)
    return record.relatedScheduleItems
  const classTimes = normalizeStringArray(record.classTimes)
  const relatedSchedules = normalizeStringArray(record.relatedSchedules)
  const length = Math.max(classTimes.length, relatedSchedules.length)
  return Array.from({ length }, (_, index) => ({
    scheduleId: '',
    classTime: classTimes[index] || '-',
    scheduleName: relatedSchedules[index] || '-',
    rollCallStatus: '-',
  }))
}

function parseRelatedScheduleEndTime(value?: string) {
  const text = formatSingleLineClassTime(String(value || ''))
  const matched = text.match(/(\d{4}-\d{2}-\d{2})\s+(\d{2}:\d{2})\s*~\s*(\d{2}:\d{2})/)
  if (!matched)
    return null
  const [, dateText, _startText, endText] = matched
  const parsed = dayjs(`${dateText} ${endText}`)
  return parsed.isValid() ? parsed : null
}

function isOverdueUnsignedSchedule(record?: FaceAttendanceRelatedScheduleItem) {
  if (!record || String(record.rollCallStatus || '').trim() !== '未点名')
    return false
  const endTime = parseRelatedScheduleEndTime(record.classTime)
  return !!endTime && endTime.isBefore(dayjs())
}

function hasOverdueUnsignedSchedules(record?: Partial<FaceAttendanceRecordItem> | Record<string, any>) {
  return getRelatedScheduleItems(record).some(item => isOverdueUnsignedSchedule(item))
}

function faceAttendancePromptText(record?: Partial<FaceAttendanceRecordItem> | Record<string, any>) {
  if (hasOverdueUnsignedSchedules(record))
    return '存在已过时未点名日程，请手动点名'
  return String(record?.prompt || '-').trim() || '-'
}

function relatedScheduleStatusText(record?: FaceAttendanceRelatedScheduleItem) {
  if (isOverdueUnsignedSchedule(record))
    return '待手动点名'
  return String(record?.rollCallStatus || '-').trim() || '-'
}

function rollCallStatusClass(value?: string) {
  if (value === '已点名')
    return 'roll-call-status roll-call-status--signed'
  if (value === '未点名')
    return 'roll-call-status roll-call-status--unsigned'
  return 'roll-call-status'
}

function relatedScheduleStatusClass(record?: FaceAttendanceRelatedScheduleItem) {
  if (isOverdueUnsignedSchedule(record))
    return 'roll-call-status roll-call-status--manual'
  return rollCallStatusClass(record?.rollCallStatus)
}

function relatedScheduleRowClassName(record: FaceAttendanceRelatedScheduleItem) {
  return isOverdueUnsignedSchedule(record) ? 'related-schedule-row related-schedule-row--manual' : 'related-schedule-row'
}

function getRelatedScheduleLessonDay(record?: FaceAttendanceRelatedScheduleItem) {
  const text = formatSingleLineClassTime(String(record?.classTime || ''))
  const matched = text.match(/(\d{4}-\d{2}-\d{2})/)
  return matched?.[1] || ''
}

function openRelatedScheduleModal(record: Record<string, any>) {
  currentScheduleRecord.value = record as FaceAttendanceRecordItem
  scheduleModalOpen.value = true
}

function handleGoRollCall(record?: FaceAttendanceRelatedScheduleItem) {
  const scheduleId = String(record?.scheduleId || '').trim()
  if (!scheduleId)
    return
  currentRollCallScheduleId.value = scheduleId
  currentRollCallLessonDay.value = getRelatedScheduleLessonDay(record)
  rollCallDrawerOpen.value = true
}

async function handleRollCallUpdated() {
  const currentRecordId = String(currentScheduleRecord.value?.id || '').trim()
  await loadList()
  if (currentRecordId) {
    const nextRecord = dataSource.value.find(item => String(item.id || '').trim() === currentRecordId)
    currentScheduleRecord.value = nextRecord || null
  }
}

function handleCreateTimeFilter(value: unknown) {
  filterSignInTime.value = Array.isArray(value)
    ? value.map(item => String(item || '').trim()).filter(Boolean)
    : []
  pagination.value.current = 1
  loadList()
}

function handleSignOutTimeFilter(value: unknown) {
  filterSignOutTime.value = Array.isArray(value)
    ? value.map(item => String(item || '').trim()).filter(Boolean)
    : []
  pagination.value.current = 1
  loadList()
}

function handleActionTypeFilter(value: unknown) {
  filterActionTypes.value = Array.isArray(value)
    ? value.map(item => String(item || '').trim()).filter(Boolean)
    : []
  pagination.value.current = 1
  loadList()
}

function handlePendingSignOutFilter(value: unknown) {
  const normalized = Array.isArray(value) ? value[0] : value
  const text = String(normalized || '').trim()
  filterPendingSignOut.value = text || undefined
  pagination.value.current = 1
  loadList()
}

function handleStudentFilter(value: unknown) {
  const normalized = Array.isArray(value) ? value[0] : value
  const text = String(normalized || '').trim()
  filterStudentId.value = text || undefined
  pagination.value.current = 1
  loadList()
}

async function loadList() {
  loading.value = true
  try {
    const res = await pageFaceAttendanceRecordsApi({
      pageRequestModel: {
        pageSize: pagination.value.pageSize,
        pageIndex: pagination.value.current,
      },
      queryModel: {
        studentId: filterStudentId.value,
        actionTypes: filterActionTypes.value,
        beginSignInTime: filterSignInTime.value[0],
        endSignInTime: filterSignInTime.value[1],
        beginSignOutTime: filterSignOutTime.value[0],
        endSignOutTime: filterSignOutTime.value[1],
        pendingSignOut: filterPendingSignOut.value === '1' ? true : (filterPendingSignOut.value === '0' ? false : undefined),
      },
    })
    if (res.code === 200) {
      dataSource.value = Array.isArray(res.result) ? res.result : []
      pagination.value.total = Number(res.total || 0)
      return
    }
    dataSource.value = []
    pagination.value.total = 0
  }
  catch (error) {
    console.error('load face attendance records failed', error)
    dataSource.value = []
    pagination.value.total = 0
  }
  finally {
    loading.value = false
  }
}

function handleTableChange(page: { current?: number, pageSize?: number }) {
  const nextCurrent = Number(page.current || 1)
  const nextSize = Number(page.pageSize || pagination.value.pageSize)
  const sizeChanged = nextSize !== pagination.value.pageSize
  pagination.value.pageSize = nextSize
  pagination.value.current = sizeChanged ? 1 : nextCurrent
  loadList()
}

onMounted(() => {
  loadList()
})
</script>

<template>
  <div>
    <div class="filter-wrap  bg-white  pl-3 pr-3 rounded-4">
      <all-filter
        :display-array="displayArray"
        :is-quick-show="false"
        :is-show-search-stu-phonefilter="true"
        :whole-condition-clear-types="['scheduleType']"
        create-time-label="签到时间"
        last-edited-time-label="签退时间"
        schedule-type-label="签到/签退类型"
        :schedule-type-options="attendanceActionOptions"
        schedule-call-status-label="待签退"
        :schedule-call-status-options="pendingSignOutOptions"
        search-placeholder="搜索姓名/手机号"
        @update:createTimeFilter="handleCreateTimeFilter"
        @update:lastEditedTimeFilter="handleSignOutTimeFilter"
        @update:scheduleTypeFilter="handleActionTypeFilter"
        @update:scheduleCallStatusFilter="handlePendingSignOutFilter"
        @update:stuPhoneSearchFilter="handleStudentFilter"
      />
    </div>
    <div class="student-list mt-2 pt-3 pb-3 pl-6 pr-6 bg-white rounded-4">
      <div class="tab-table">
        <div class="table-title flex justify-between">
          <div class="total">
            共 {{ pagination.total }} 条
          </div>
          <div class="edit flex">
            <a-button class="mr-2">
              导出数据
            </a-button>
            <customize-code
              v-model:checked-values="selectedValues"
              :options="columnOptions"
              :total="allColumns.length"
              :num="selectedValues.length"
            />
          </div>
        </div>
        <div class="table-content mt-2">
          <a-table
            row-key="id"
            :loading="loading"
            :data-source="dataSource"
            :pagination="tablePagination"
            :columns="filteredColumns"
            :row-selection="rowSelection"
            :scroll="{ x: totalWidth }"
            size="small"
            @change="handleTableChange"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'name'">
                <StudentAvatar
                  :id="record.studentId"
                  :name="record.studentName || '-'"
                  :avatar-url="record.avatarUrl"
                  :phone="record.studentMobile || ''"
                  :show-gender="false"
                  :show-age="false"
                  :auto-width="false"
                />
              </template>
              <template v-else-if="column.key === 'face'">
                <div class="flex flex-items-center">
                  <span class="whitespace-nowrap" :class="record.isCollect ? 'text-#333' : 'text-#999'">
                    {{ record.isCollect ? '已采集' : '未采集' }}
                  </span>
                  <svg
                    v-if="record.isCollect"
                    width="16px"
                    height="16px"
                    viewBox="0 0 16 16"
                    style="margin-left: 12px;"
                  >
                    <g stroke="none" stroke-width="1" fill="none" fill-rule="evenodd">
                      <g transform="translate(-594.000000, -464.000000)">
                        <g transform="translate(518.000000, 310.000000)">
                          <g transform="translate(0.000000, 126.000000)">
                            <g transform="translate(76.000000, 21.600000)">
                              <g transform="translate(0.000000, 6.400000)">
                                <polygon fill="#000000" fill-rule="nonzero" opacity="0" points="0 0 16 0 16 16 8 16 0 16" />
                                <path
                                  d="M1.49983336,11 C1.74529324,10.9999182 1.94950067,11.1767253 1.99191437,11.4099604 L2,11.4998334 L2,14 L4.5,14 C4.74545992,14 4.9496084,14.1768752 4.99194436,14.4101244 L5,14.5 C5,14.7454599 4.82312487,14.9496084 4.58987566,14.9919444 L4.5,15 L1.50100003,15 C1.25559799,15 1.05147725,14.8232051 1.00908211,14.5900195 L1.00100006,14.5001667 L1,11.5001667 C0.999908009,11.2240243 1.223691,11.0000921 1.49983336,11 Z M14.4988336,11 C14.7442935,10.9999183 14.9485009,11.1767254 14.9909146,11.4099605 L14.9990002,11.4998334 L15,14.4998334 C15.0000818,14.7453511 14.8231944,14.9495863 14.5898958,14.9919408 L14.5,15 L11.5,15 C11.2238576,15 11,14.7761424 11,14.5 C11,14.2545401 11.1768752,14.0503917 11.4101244,14.0080557 L11.5,14 L14,14 L13.9990003,11.5001667 C13.9989185,11.2547068 14.1757256,11.0504994 14.4089607,11.0080857 L14.4988336,11 Z M4.5,9 L11.5,9 L11.4931641,9.38828125 L11.4931641,9.38828125 L11.4769287,9.60498047 L11.4769287,9.60498047 L11.4453125,9.83125 C11.28125,10.75 10.625,11.8 8,11.8 C5.484375,11.8 4.77685547,10.8356771 4.5778656,9.94669189 L4.53663635,9.71717529 C4.53140259,9.67943522 4.5269165,9.64200846 4.52307129,9.60498047 L4.50683594,9.38828125 L4.50683594,9.38828125 L4.5,9 Z M11,5.5 C11.5522847,5.5 12,5.94771525 12,6.5 C12,7.05228475 11.5522847,7.5 11,7.5 C10.4477153,7.5 10,7.05228475 10,6.5 C10,5.94771525 10.4477153,5.5 11,5.5 Z M5,5.5 C5.55228475,5.5 6,5.94771525 6,6.5 C6,7.05228475 5.55228475,7.5 5,7.5 C4.44771525,7.5 4,7.05228475 4,6.5 C4,5.94771525 4.44771525,5.5 5,5.5 Z M14.5,1 C14.7455177,1 14.9496939,1.17695541 14.9919707,1.41026814 L15,1.50016663 L14.9990002,4.50016663 C14.9989082,4.77630898 14.774976,5.000092 14.4988336,5 C14.2533737,4.99991817 14.0492842,4.82297499 14.007026,4.58971169 L13.9990003,4.49983337 L14,2 L11.5,2 C11.2545401,2 11.0503916,1.82312484 11.0080557,1.58987563 L11,1.5 C11,1.25454011 11.1768752,1.05039163 11.4101244,1.00805567 L11.5,1 L14.5,1 Z M4.5,1 C4.77614235,1 5,1.22385763 5,1.5 C5,1.74545989 4.82312481,1.94960837 4.5898756,1.99194433 L4.5,2 L2,2 L2,4.50016667 C1.99991812,4.74562654 1.82297492,4.94971605 1.58971162,4.99197426 L1.49983331,5 C1.25437343,4.99991815 1.05028392,4.82297495 1.00802571,4.58971165 L1,4.49983333 L1.001,1.49983333 C1.0010818,1.25443131 1.17794474,1.05036951 1.41114451,1.0080521 L1.50099997,1 L4.5,1 Z"
                                  fill="#0066FF"
                                />
                              </g>
                            </g>
                          </g>
                        </g>
                      </g>
                    </g>
                  </svg>
                </div>
              </template>
              <template v-else-if="column.key === 'faceType'">
                {{ record.attendanceType || '人脸考勤' }}
              </template>
              <template v-else-if="column.key === 'signInOutType'">
                <div class="attendance-action-cell">
                  <div class="attendance-entry attendance-entry--sign-in">
                    <span class="attendance-entry__tag">签到</span>
                    <span class="attendance-entry__text">{{ getAttendancePrimaryActionLabel(record) }}</span>
                    <img
                      v-if="getAttendancePhoto(record, 'sign_in')"
                      class="attendance-photo-trigger"
                      :src="faceIcon"
                      alt=""
                      @click="openAttendancePhoto(record, 'sign_in')"
                    >
                  </div>
                  <div
                    v-if="getAttendanceSecondaryActionLabel(record)"
                    class="attendance-entry"
                    :class="record.signOutTime ? 'attendance-entry--sign-out' : 'attendance-entry--pending'"
                  >
                    <span class="attendance-entry__tag">签退</span>
                    <span class="attendance-entry__text">{{ getAttendanceSecondaryActionLabel(record) }}</span>
                    <img
                      v-if="record.signOutTime && getAttendancePhoto(record, 'sign_out')"
                      class="attendance-photo-trigger"
                      :src="faceIcon"
                      alt=""
                      @click="openAttendancePhoto(record, 'sign_out')"
                    >
                  </div>
                </div>
              </template>
              <template v-else-if="column.key === 'signInOutTime'">
                <div class="attendance-action-cell">
                  <div class="attendance-entry attendance-entry--sign-in">
                    <span class="attendance-entry__tag">签到</span>
                    <span class="attendance-entry__text attendance-entry__text--time">{{ formatDateTime(record.attendanceTime) }}</span>
                  </div>
                  <div
                    v-if="record.signOutTime || isPendingSignOut(record)"
                    class="attendance-entry"
                    :class="record.signOutTime ? 'attendance-entry--sign-out' : 'attendance-entry--pending'"
                  >
                    <span class="attendance-entry__tag">签退</span>
                    <span class="attendance-entry__text attendance-entry__text--time">
                      {{ record.signOutTime ? formatDateTime(record.signOutTime) : '待签退' }}
                    </span>
                  </div>
                </div>
              </template>
              <template v-else-if="column.key === 'schedulePlan'">
                {{ record.hasSchedule ? '有' : '无' }}
              </template>
              <template v-else-if="column.key === 'linkSchedule'">
                <template v-if="Array.isArray(record.relatedScheduleItems) ? record.relatedScheduleItems.length : normalizeStringArray(record.relatedSchedules).length">
                  <div class="face-record-link-schedule">
                    <a-button type="link" size="small" class="px-0" @click="openRelatedScheduleModal(record)">
                      查看
                    </a-button>
                    <span v-if="hasOverdueUnsignedSchedules(record)" class="face-record-alert-badge">
                      待手动点名
                    </span>
                  </div>
                </template>
                <template v-else>
                  -
                </template>
              </template>
              <template v-else-if="column.key === 'remarks'">
                <span :class="{ 'face-record-alert-text': hasOverdueUnsignedSchedules(record) }">
                  {{ faceAttendancePromptText(record) }}
                </span>
              </template>
            </template>
          </a-table>
        </div>
      </div>
    </div>
    <a-modal
      v-model:open="scheduleModalOpen"
      title="相关日程"
      width="920px"
      :footer="null"
      destroy-on-close
    >
      <a-table
        :row-key="record => `${record.classTime || ''}-${record.scheduleName || ''}-${record.rollCallStatus || ''}`"
        :columns="relatedScheduleColumns"
        :data-source="relatedScheduleDataSource"
        :pagination="false"
        :row-class-name="relatedScheduleRowClassName"
        :scroll="{ x: 860 }"
        size="small"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'classTime'">
            <span class="related-schedule-time">
              {{ formatSingleLineClassTime(record.classTime) }}
            </span>
          </template>
          <template v-else-if="column.key === 'scheduleName'">
            <div class="related-schedule-name-cell">
              <span class="related-schedule-name-text" :title="record.scheduleName || '-'">{{ record.scheduleName || '-' }}</span>
              <span v-if="isOverdueUnsignedSchedule(record)" class="related-schedule-hint">
                已过上课时间
              </span>
            </div>
          </template>
          <template v-else-if="column.key === 'rollCallStatus'">
            <div class="related-schedule-status-cell">
              <span :class="relatedScheduleStatusClass(record)">
                {{ relatedScheduleStatusText(record) }}
              </span>
            </div>
          </template>
          <template v-else-if="column.key === 'action'">
            <div class="related-schedule-action-cell">
              <a-button
                v-if="isOverdueUnsignedSchedule(record) && record.scheduleId"
                type="link"
                size="small"
                class="px-0"
                @click="handleGoRollCall(record)"
              >
                去点名
              </a-button>
              <template v-else>
                -
              </template>
            </div>
          </template>
        </template>
      </a-table>
    </a-modal>
    <RollCallDrawer
      v-model:open="rollCallDrawerOpen"
      :schedule-id="currentRollCallScheduleId"
      :lesson-day="currentRollCallLessonDay"
      @updated="handleRollCallUpdated"
      @confirmed="handleRollCallUpdated"
    />
    <a-modal v-model:open="previewVisible" :title="previewTitle" :footer="null" @cancel="previewVisible = false">
      <div class="attendance-preview">
        <img alt="attendance" class="attendance-preview__image" :src="previewImage">
        <div v-if="previewTimeText" class="attendance-preview__stamp">
          {{ previewTimeText }}
        </div>
      </div>
    </a-modal>
  </div>
</template>

<style lang="less" scoped>
.actionbtn {
  color: var(--pro-ant-color-primary);
  cursor: pointer;
}

.filter-wrap {
  border-top-left-radius: 0 !important;
  border-top-right-radius: 0 !important;
}

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

.roll-call-status {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 60px;
  height: 24px;
  padding: 0 10px;
  border-radius: 12px;
  font-size: 12px;
  line-height: 24px;
  color: #666;
  background: #f5f5f5;
}

.roll-call-status--signed {
  color: #1f8f55;
  background: #edf9f2;
}

.roll-call-status--unsigned {
  color: #8a5a00;
  background: #fff6e8;
}

.roll-call-status--manual {
  color: #cf1322;
  background: #fff1f0;
  box-shadow: inset 0 0 0 1px #ffccc7;
  font-weight: 600;
}

.related-schedule-name-cell {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
  min-width: 0;
}

.related-schedule-time,
.related-schedule-status-cell,
.related-schedule-action-cell {
  white-space: nowrap;
}

.related-schedule-name-text {
  flex: 1 1 auto;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: #1f2937;
}

.related-schedule-status-cell,
.related-schedule-action-cell {
  display: flex;
  align-items: center;
}

.related-schedule-hint {
  display: inline-flex;
  align-items: center;
  height: 20px;
  padding: 0 8px;
  border-radius: 999px;
  background: #fff7e6;
  color: #d46b08;
  font-size: 12px;
  line-height: 20px;
  white-space: nowrap;
  flex-shrink: 0;
}

:deep(.related-schedule-row--manual td) {
  background: #fffafa !important;
}

.face-record-link-schedule {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.face-record-alert-badge {
  display: inline-flex;
  align-items: center;
  height: 22px;
  padding: 0 8px;
  border-radius: 999px;
  background: #fff1f0;
  color: #cf1322;
  box-shadow: inset 0 0 0 1px #ffccc7;
  font-size: 12px;
  line-height: 22px;
  font-weight: 600;
  white-space: nowrap;
}

.face-record-alert-text {
  color: #cf1322;
  font-weight: 500;
  font-size: 12px;
  line-height: 18px;
}

.attendance-action-cell {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.attendance-entry {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  min-height: 22px;
}

.attendance-entry__tag {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 34px;
  height: 20px;
  padding: 0 8px;
  border-radius: 999px;
  font-size: 12px;
  line-height: 20px;
  font-weight: 600;
  white-space: nowrap;
}

.attendance-entry__text {
  color: #334155;
  font-size: 13px;
  line-height: 20px;
}

.attendance-entry__text--time {
  font-variant-numeric: tabular-nums;
}

.attendance-entry--sign-in .attendance-entry__tag {
  color: #1668dc;
  background: #e8f3ff;
}

.attendance-entry--sign-out .attendance-entry__tag {
  color: #1f8f55;
  background: #edf9f2;
}

.attendance-entry--pending .attendance-entry__tag {
  color: #ad6800;
  background: #fff7e6;
}

.attendance-entry--pending .attendance-entry__text {
  color: #935f00;
}

.attendance-photo-trigger {
  width: 16px;
  height: 16px;
  cursor: pointer;
  flex: 0 0 auto;
}

.attendance-preview {
  position: relative;
}

.attendance-preview__image {
  width: 100%;
  display: block;
}

.attendance-preview__stamp {
  position: absolute;
  right: 12px;
  bottom: 12px;
  padding: 4px 10px;
  border-radius: 4px;
  background: rgba(0, 0, 0, 0.58);
  color: #fff;
  font-size: 12px;
  line-height: 18px;
  font-variant-numeric: tabular-nums;
}
</style>
