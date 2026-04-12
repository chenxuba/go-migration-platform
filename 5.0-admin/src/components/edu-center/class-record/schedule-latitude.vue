<script setup lang="ts">
import { DownOutlined } from '@ant-design/icons-vue'
import dayjs, { type Dayjs } from 'dayjs'
import { computed, onMounted, ref, watch } from 'vue'
import { getScheduleTeachingRecordPagedListApi, type ScheduleTeachingRecordItem } from '@/api/edu-center/class-record'

const monthStart = dayjs().startOf('month')
const today = dayjs()
const defaultScheduleDateVals = [monthStart.format('YYYY-MM-DD'), today.format('YYYY-MM-DD')]
const displayArray = ref(['scheduleDate', 'scheduleType', 'isArrears'])
const scheduleTypeOptions = [
  { id: '1', value: '班级日程' },
  { id: '2', value: '1对1日程' },
  { id: '3', value: '试听日程' },
]

const dataSource = ref<ScheduleTeachingRecordItem[]>([])
const loading = ref(false)
const openClassRecordDrawer = ref(false)
const currentTeachingRecordId = ref('')
const filterDateRange = ref<[Dayjs, Dayjs]>([monthStart, today])
const filterScheduleTypes = ref<string[]>([])
const filterIsArrear = ref<boolean | null>(null)
const summary = ref({
  total: 0,
  totalClassTimes: 0,
  totalTeacherTimes: 0,
  totalTuition: 0,
})
const pagination = ref({
  current: 1,
  pageSize: 50,
  total: 0,
})

function handleSeeClassRecord(record?: Partial<ScheduleTeachingRecordItem>) {
  currentTeachingRecordId.value = String(record?.teachingRecordId || '').trim()
  openClassRecordDrawer.value = true
}

const allColumns = ref<any[]>([
  {
    title: '上课日期/时段',
    dataIndex: 'classDateTime',
    key: 'classDateTime',
    fixed: 'left',
    width: 160,
    required: true,
  },
  {
    title: '日程类型',
    dataIndex: 'scheduleType',
    key: 'scheduleType',
    width: 140,
  },
  {
    title: '所属班级/1v1',
    key: 'linkClass1v1',
    dataIndex: 'cloud',
    width: 180,
  },
  {
    title: '所属课程',
    key: 'course',
    dataIndex: 'course',
    width: 160,
  },
  {
    title: '科目',
    dataIndex: 'subject',
    key: 'subject',
    width: 110,
  },
  {
    title: '点名状态',
    dataIndex: 'callStatus',
    key: 'callStatus',
    width: 130,
  },
  {
    title: '出勤率',
    dataIndex: 'attendanceRate',
    key: 'attendanceRate',
    width: 150,
  },
  {
    title: '消耗数量',
    dataIndex: 'useNum',
    key: 'useNum',
    width: 140,
  },
  {
    title: '消耗学费',
    dataIndex: 'usePrice',
    key: 'usePrice',
    width: 140,
  },
  {
    title: '上课老师',
    dataIndex: 'mainTeacher',
    key: 'mainTeacher',
    width: 140,
  },
  {
    title: '上课助教',
    dataIndex: 'subTeacher',
    key: 'subTeacher',
    width: 140,
  },
  {
    title: '教师记录课时',
    dataIndex: 'teacherOweNum',
    key: 'teacherOweNum',
    width: 140,
  },
  {
    title: '创建时间',
    key: 'createTime',
    dataIndex: 'createTime',
    width: 170,
  },
  {
    title: '操作',
    dataIndex: 'action',
    key: 'action',
    fixed: 'right',
    width: 140,
    required: true,
  },
])

const savedSelected = localStorage.getItem('schedule-latitude')
const keysArray = allColumns.value
  .map(column => column?.key)
  .filter(key => typeof key !== 'undefined')
const initialSelectedValues = savedSelected ? JSON.parse(savedSelected) : keysArray
const selectedValues = ref(initialSelectedValues)
const columnOptions = computed(() =>
  allColumns.value
    .filter(col => col.key !== 'action')
    .map(col => ({
      id: col.key,
      value: col.title,
      disabled: col.required,
    })),
)
const filteredColumns = computed(() => {
  const requiredColumns = allColumns.value.filter(col => col.required)
  const optionalColumns = allColumns.value.filter(col =>
    selectedValues.value.includes(col.key) && !col.required,
  )
  return [
    ...requiredColumns.filter(col => col.fixed === 'left'),
    ...optionalColumns,
    ...requiredColumns.filter(col => col.fixed === 'right'),
  ]
})

watch(selectedValues, (newVal) => {
  const requiredKeys = allColumns.value.filter(col => col.required).map(col => col.key)
  if (!requiredKeys.every(key => newVal.includes(key))) {
    selectedValues.value = Array.from(new Set([
      ...newVal.filter(value => !requiredKeys.includes(value)),
      ...requiredKeys,
    ]))
  }
}, { deep: true })

watch(selectedValues, (newVal) => {
  localStorage.setItem('schedule-latitude', JSON.stringify(newVal))
}, { deep: true })

const totalWidth = computed(() =>
  filteredColumns.value.reduce((acc, column) => acc + (column.width || 0), 0),
)

const rowSelection = {
  onChange: (_selectedRowKeys: (string | number)[], _selectedRows: ScheduleTeachingRecordItem[]) => {},
}

const tablePagination = computed(() => ({
  current: pagination.value.current,
  pageSize: pagination.value.pageSize,
  total: pagination.value.total,
  showSizeChanger: true,
  showQuickJumper: true,
  showTotal: (total: number) => `共 ${total} 条`,
}))

function formatNumber(value?: number, suffix = '') {
  const num = Number(value || 0)
  if (!Number.isFinite(num))
    return suffix ? `0${suffix}` : '0'
  const text = Number.isInteger(num) ? String(num) : num.toFixed(2).replace(/\.?0+$/, '')
  return suffix ? `${text}${suffix}` : text
}

function formatCurrency(value?: number) {
  return `¥${Number(value || 0).toFixed(2)}`
}

function formatDateTime(record: Partial<ScheduleTeachingRecordItem> | Record<string, any>) {
  const start = dayjs(record.startTime)
  const end = dayjs(record.endTime)
  if (!start.isValid() || !end.isValid()) {
    return {
      dateText: '-',
      timeText: '--:-- ~ --:--',
    }
  }
  const weekday = ['周日', '周一', '周二', '周三', '周四', '周五', '周六'][start.day()] || ''
  return {
    dateText: `${start.format('YYYY-MM-DD')} (${weekday})`,
    timeText: `${start.format('HH:mm')} ~ ${end.format('HH:mm')}`,
  }
}

function scheduleTypeText(value?: number) {
  const type = Number(value || 0)
  if (type === 2)
    return '1对1日程'
  if (type === 3)
    return '试听日程'
  return '班级日程'
}

function rollCallStatusText(value?: number) {
  return Number(value || 0) === 1 ? '部分点名' : '全部点名'
}

function rollCallStatusTagClass(value?: number) {
  return Number(value || 0) === 1
    ? 'roll-call-status-tag roll-call-status-tag--partial'
    : 'roll-call-status-tag roll-call-status-tag--full'
}

function classDisplay(record: Partial<ScheduleTeachingRecordItem> | Record<string, any>) {
  return record.className || record.one2OneName || '-'
}

function attendanceRateText(record: Partial<ScheduleTeachingRecordItem> | Record<string, any>) {
  const shouldAttendCount = Number(record.shouldAttendCount || 0)
  if (shouldAttendCount <= 0)
    return '--'
  return `${Math.round(Number(record.attendanceRate || 0) * 100)}%`
}

function displayConsumedQuantity(record: Partial<ScheduleTeachingRecordItem> | Record<string, any>) {
  const actualTuition = Number(record.actualTuition || 0)
  const actualQuantity = Number(record.actualQuantity || 0)
  if (actualTuition <= 0 && actualQuantity > 0)
    return 0
  return actualQuantity
}

function buildQueryModel() {
  return {
    beginStartTime: filterDateRange.value[0]?.format('YYYY-MM-DD'),
    endStartTime: filterDateRange.value[1]?.format('YYYY-MM-DD'),
    timetableSourceTypes: filterScheduleTypes.value.map(item => Number(item)).filter(Boolean),
    isArrear: filterIsArrear.value,
  }
}

async function loadList() {
  loading.value = true
  try {
    const res = await getScheduleTeachingRecordPagedListApi({
      queryModel: buildQueryModel(),
      pageRequestModel: {
        needTotal: true,
        pageIndex: pagination.value.current,
        pageSize: pagination.value.pageSize,
        skipCount: (pagination.value.current - 1) * pagination.value.pageSize,
      },
      sortModel: {
        startTime: 2,
        updatedTime: 0,
      },
    })
    if (res.code === 200) {
      dataSource.value = Array.isArray(res.result?.list) ? res.result.list : []
      summary.value = {
        total: Number(res.result?.total || 0),
        totalClassTimes: Number(res.result?.totalClassTimes || 0),
        totalTeacherTimes: Number(res.result?.totalTeacherTimes || 0),
        totalTuition: Number(res.result?.totalTuition || 0),
      }
      pagination.value.total = Number(res.result?.total || 0)
      return
    }
    dataSource.value = []
    summary.value = { total: 0, totalClassTimes: 0, totalTeacherTimes: 0, totalTuition: 0 }
    pagination.value.total = 0
  }
  catch (error) {
    console.error('load schedule teaching records failed', error)
    dataSource.value = []
    summary.value = { total: 0, totalClassTimes: 0, totalTeacherTimes: 0, totalTuition: 0 }
    pagination.value.total = 0
  }
  finally {
    loading.value = false
  }
}

function handleScheduleDateFilter(value: unknown) {
  if (!Array.isArray(value) || value.length < 2) {
    filterDateRange.value = [monthStart, today]
    return
  }
  const start = dayjs(String(value[0] || ''))
  const end = dayjs(String(value[1] || ''))
  filterDateRange.value = [
    start.isValid() ? start : monthStart,
    end.isValid() ? end : today,
  ]
}

function handleScheduleTypeFilter(value: unknown) {
  filterScheduleTypes.value = Array.isArray(value) ? value.map(item => String(item || '')).filter(Boolean) : []
}

function handleIsArrearsFilter(value: unknown) {
  if (value === null || value === undefined || value === '')
    filterIsArrear.value = null
  else
    filterIsArrear.value = Number(value) === 1
}

function handleTableChange(page: { current?: number, pageSize?: number }) {
  pagination.value.current = Number(page?.current || 1)
  pagination.value.pageSize = Number(page?.pageSize || pagination.value.pageSize)
  loadList()
}

watch(
  [filterDateRange, filterScheduleTypes, filterIsArrear],
  () => {
    pagination.value.current = 1
    loadList()
  },
  { deep: true },
)

onMounted(() => {
  loadList()
})
</script>

<template>
  <div>
    <div class="filter-wrap bg-white  pl-3 pr-3 rounded-4">
      <all-filter
        :display-array="displayArray"
        :default-schedule-date-vals="defaultScheduleDateVals"
        :schedule-date-disable-future="true"
        :schedule-type-options="scheduleTypeOptions"
        :is-quick-show="false"
        @update:schedule-date-filter="handleScheduleDateFilter"
        @update:schedule-type-filter="handleScheduleTypeFilter"
        @update:is-arrears-filter="handleIsArrearsFilter"
      />
    </div>
    <div class="student-list mt-3 pt-3 pb-3 pl-6 pr-6 bg-white rounded-4">
      <div class="tab-table">
        <div class="table-title flex justify-between">
          <div class="total">
            共 {{ summary.total }} 条记录 ，学员总计 {{ formatNumber(summary.totalClassTimes, '课时') }}，上课教师总计 {{ formatNumber(summary.totalTeacherTimes, '课时') }} ，共消耗学费 {{ formatCurrency(summary.totalTuition) }}
          </div>
          <div class="edit flex">
            <a-button class="mr-2">
              变更日志
            </a-button>
            <a-dropdown class="mr-2">
              <template #overlay>
                <a-menu>
                  <a-menu-item key="1">
                    批量导出
                  </a-menu-item>
                  <a-menu-item key="3">
                    导出记录
                  </a-menu-item>
                </a-menu>
              </template>
              <a-button>
                导出数据
                <DownOutlined :style="{ fontSize: '10px' }" />
              </a-button>
            </a-dropdown>
            <customize-code
              v-model:checked-values="selectedValues" :options="columnOptions" :total="allColumns.length - 1"
              :num="selectedValues.length - 1"
            />
          </div>
        </div>
        <div class="table-content mt-2">
          <a-table
            row-key="teachingRecordId"
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
              <template v-if="column.key === 'classDateTime'">
                <div class="name">
                  <div class="text-#000">
                    {{ formatDateTime(record).dateText }}
                  </div>
                  <div class="text-3 text-#888 flex flex-items-center">
                    {{ formatDateTime(record).timeText }}
                  </div>
                </div>
              </template>
              <template v-if="column.key === 'linkClass1v1'">
                {{ classDisplay(record) }}
              </template>
              <template v-if="column.key === 'course'">
                {{ record.lessonName || '-' }}
              </template>
              <template v-if="column.key === 'subject'">
                {{ record.subjectName || '-' }}
              </template>
              <template v-if="column.key === 'scheduleType'">
                {{ scheduleTypeText(record.timetableSourceType) }}
              </template>
              <template v-if="column.key === 'useNum'">
                {{ formatNumber(displayConsumedQuantity(record), '课时') }}
              </template>
              <template v-if="column.key === 'callStatus'">
                <span :class="rollCallStatusTagClass(record.rollCallStatus)">
                  {{ rollCallStatusText(record.rollCallStatus) }}
                </span>
              </template>
              <template v-if="column.key === 'attendanceRate'">
                <div class="name">
                  <div class="text-#000">
                    {{ attendanceRateText(record) }}
                  </div>
                  <div class="text-3 text-#888 flex flex-items-center">
                    实到{{ record.attendCount || 0 }}人 / 应到{{ record.shouldAttendCount || 0 }}人
                  </div>
                </div>
              </template>
              <template v-if="column.key === 'usePrice'">
                {{ formatCurrency(record.actualTuition) }}
              </template>
              <template v-if="column.key === 'mainTeacher'">
                {{ record.teacherName || '-' }}
              </template>
              <template v-if="column.key === 'subTeacher'">
                {{ record.assistants || '-' }}
              </template>
              <template v-if="column.key === 'teacherOweNum'">
                {{ formatNumber(record.teacherClassTime, '课时') }}
              </template>
              <template v-if="column.key === 'createTime'">
                {{ record.createdTime || '-' }}
              </template>
              <template v-if="column.key === 'action'">
                <a class="font500" @click="handleSeeClassRecord(record)">上课记录详情</a>
              </template>
            </template>
          </a-table>
        </div>
      </div>
    </div>
    <class-record-details v-model:open="openClassRecordDrawer" :teaching-record-id="currentTeachingRecordId" />
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

.roll-call-status-tag {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 72px;
  padding: 2px 10px;
  border-radius: 999px;
  font-size: 12px;
  line-height: 20px;
  white-space: nowrap;
  border: 1px solid transparent;
}

.roll-call-status-tag--full {
  color: #166534;
  background: #f0fdf4;
  border-color: #bbf7d0;
}

.roll-call-status-tag--partial {
  color: #b45309;
  background: #fff7ed;
  border-color: #fed7aa;
}
</style>
