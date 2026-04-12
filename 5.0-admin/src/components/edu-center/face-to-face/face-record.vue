<script setup lang="ts">
import dayjs from 'dayjs'
import { pageFaceAttendanceRecordsApi, type FaceAttendanceRecordItem, type FaceAttendanceRelatedScheduleItem } from '@/api/edu-center/face'
import faceIcon from '@/assets/images/face.png'
import StudentAvatar from '@/components/common/StudentAvatar.vue'
import { useTableColumns } from '@/composables/useTableColumns'

const displayArray = ref(['createTime'])
const dataSource = ref<FaceAttendanceRecordItem[]>([])
const loading = ref(false)
const filterStudentId = ref<string | undefined>(undefined)
const filterCreateTime = ref<string[]>([])
const pagination = ref({
  current: 1,
  pageSize: 20,
  total: 0,
})
const selectedRowKeys = ref<string[]>([])
const scheduleModalOpen = ref(false)
const currentScheduleRecord = ref<FaceAttendanceRecordItem | null>(null)
const previewVisible = ref(false)
const previewImage = ref('')
const previewTitle = ref('')
const previewTimeText = ref('')

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
    title: '考勤时间',
    dataIndex: 'faceTime',
    width: 150,
    key: 'faceTime',
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
    width: 180,
  },
  {
    title: '相关日程',
    dataIndex: 'scheduleName',
    key: 'scheduleName',
    width: 220,
  },
  {
    title: '点名状态',
    dataIndex: 'rollCallStatus',
    key: 'rollCallStatus',
    width: 120,
  },
]

const relatedScheduleDataSource = computed<FaceAttendanceRelatedScheduleItem[]>(() => {
  const record = currentScheduleRecord.value
  if (!record)
    return []
  if (Array.isArray(record.relatedScheduleItems) && record.relatedScheduleItems.length) {
    return record.relatedScheduleItems
  }
  const classTimes = normalizeStringArray(record.classTimes)
  const relatedSchedules = normalizeStringArray(record.relatedSchedules)
  const length = Math.max(classTimes.length, relatedSchedules.length)
  return Array.from({ length }, (_, index) => ({
    classTime: classTimes[index] || '-',
    scheduleName: relatedSchedules[index] || '-',
    rollCallStatus: '-',
  }))
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

function getAttendancePhoto(record: Partial<FaceAttendanceRecordItem> | Record<string, any>) {
  if (record.action === 'sign_out')
    return String(record.signOutImage || '').trim()
  if (record.action === 'sign_in')
    return String(record.signInImage || '').trim()
  return ''
}

function getAttendancePhotoTitle(record: Partial<FaceAttendanceRecordItem> | Record<string, any>) {
  return record.action === 'sign_out' ? '签退照片' : '签到照片'
}

function openAttendancePhoto(record: Partial<FaceAttendanceRecordItem> | Record<string, any>) {
  const image = getAttendancePhoto(record)
  if (!image)
    return
  previewImage.value = image
  previewTitle.value = getAttendancePhotoTitle(record)
  previewTimeText.value = formatDateTime(String(record.actionTime || ''))
  previewVisible.value = true
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

function rollCallStatusClass(value?: string) {
  if (value === '已点名')
    return 'roll-call-status roll-call-status--signed'
  if (value === '未点名')
    return 'roll-call-status roll-call-status--unsigned'
  return 'roll-call-status'
}

function openRelatedScheduleModal(record: Record<string, any>) {
  currentScheduleRecord.value = record as FaceAttendanceRecordItem
  scheduleModalOpen.value = true
}

function handleCreateTimeFilter(value: unknown) {
  filterCreateTime.value = Array.isArray(value)
    ? value.map(item => String(item || '').trim()).filter(Boolean)
    : []
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
        beginAttendanceTime: filterCreateTime.value[0],
        endAttendanceTime: filterCreateTime.value[1],
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
        search-placeholder="搜索姓名/手机号"
        @update:createTimeFilter="handleCreateTimeFilter"
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
              <template v-else-if="column.key === 'faceTime'">
                {{ formatDateTime(record.attendanceTime) }}
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
                  <div class="attendance-action-cell__main attendance-action-cell__main--with-icon">
                    {{ record.actionLabel || '-' }}
                    <img
                      v-if="getAttendancePhoto(record)"
                      class="attendance-photo-trigger"
                      :src="faceIcon"
                      alt=""
                      @click="openAttendancePhoto(record)"
                    >
                  </div>
                  <div v-if="isPendingSignOut(record)" class="attendance-action-cell__sub">
                    待签退
                  </div>
                </div>
              </template>
              <template v-else-if="column.key === 'signInOutTime'">
                <div class="attendance-action-cell">
                  <div class="attendance-action-cell__main attendance-action-cell__time">
                    {{ formatDateTime(record.actionTime) }}
                  </div>
                  <div v-if="isPendingSignOut(record)" class="attendance-action-cell__sub">
                    -
                  </div>
                </div>
              </template>
              <template v-else-if="column.key === 'schedulePlan'">
                {{ record.hasSchedule ? '有' : '无' }}
              </template>
              <template v-else-if="column.key === 'linkSchedule'">
                <template v-if="Array.isArray(record.relatedScheduleItems) ? record.relatedScheduleItems.length : normalizeStringArray(record.relatedSchedules).length">
                  <a-button type="link" size="small" class="px-0" @click="openRelatedScheduleModal(record)">
                    查看
                  </a-button>
                </template>
                <template v-else>
                  -
                </template>
              </template>
              <template v-else-if="column.key === 'remarks'">
                {{ record.prompt || '-' }}
              </template>
            </template>
          </a-table>
        </div>
      </div>
    </div>
    <a-modal
      v-model:open="scheduleModalOpen"
      title="相关日程"
      width="680px"
      :footer="null"
      destroy-on-close
    >
      <a-table
        :row-key="record => `${record.classTime || ''}-${record.scheduleName || ''}-${record.rollCallStatus || ''}`"
        :columns="relatedScheduleColumns"
        :data-source="relatedScheduleDataSource"
        :pagination="false"
        size="small"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'classTime'">
            {{ formatSingleLineClassTime(record.classTime) }}
          </template>
          <template v-else-if="column.key === 'rollCallStatus'">
            <span :class="rollCallStatusClass(record.rollCallStatus)">
              {{ record.rollCallStatus || '-' }}
            </span>
          </template>
        </template>
      </a-table>
    </a-modal>
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

.attendance-action-cell {
  display: flex;
  flex-direction: column;
  gap: 4px;
  line-height: 20px;
}

.attendance-action-cell__main {
  color: #222;
  font-size: 14px;
  font-weight: 500;
}

.attendance-action-cell__main--with-icon {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.attendance-action-cell__time {
  font-variant-numeric: tabular-nums;
}

.attendance-action-cell__sub {
  color: #8c8c8c;
  font-size: 12px;
  line-height: 18px;
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
