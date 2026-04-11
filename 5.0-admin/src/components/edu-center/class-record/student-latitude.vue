<script setup lang="ts">
import { DownOutlined } from '@ant-design/icons-vue'
import dayjs, { type Dayjs } from 'dayjs'
import { computed, onMounted, ref, watch } from 'vue'
import { getStudentTeachingRecordPagedListApi, type StudentTeachingRecordItem } from '@/api/edu-center/class-record'
import { useTableColumns } from '@/composables/useTableColumns'

const monthStart = dayjs().startOf('month')
const today = dayjs()
const defaultScheduleDateVals = [monthStart.format('YYYY-MM-DD'), today.format('YYYY-MM-DD')]
const displayArray = ref(['scheduleDate', 'scheduleType', 'isArrears'])
const scheduleTypeOptions = [
  { id: '1', value: '班级日程' },
  { id: '2', value: '1对1日程' },
  { id: '3', value: '试听日程' },
]

const loading = ref(false)
const openClassRecordDrawer = ref(false)
const dataSource = ref<StudentTeachingRecordItem[]>([])
const filterDateRange = ref<[Dayjs, Dayjs]>([monthStart, today])
const filterScheduleTypes = ref<string[]>([])
const filterIsArrear = ref<boolean | null>(null)
const summary = ref({
  total: 0,
  totalClassTimes: 0,
  totalTuition: 0,
})
const pagination = ref({
  current: 1,
  pageSize: 50,
  total: 0,
})

function handleSeeClassRecord() {
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
    title: '学员/电话',
    dataIndex: 'name',
    key: 'name',
    fixed: 'left',
    width: 160,
    required: true,
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
    title: '日程类型',
    dataIndex: 'scheduleType',
    key: 'scheduleType',
    width: 140,
  },
  {
    title: '学员身份',
    dataIndex: 'studentIdentity',
    key: 'studentIdentity',
    width: 140,
  },
  {
    title: '上课状态',
    dataIndex: 'classStatus',
    key: 'classStatus',
    width: 120,
  },
  {
    title: '扣费课程账户',
    dataIndex: 'deductionAccount',
    key: 'deductionAccount',
    width: 160,
  },
  {
    title: '课消方式',
    key: 'courseNotMethod',
    dataIndex: 'courseNotMethod',
    width: 110,
  },
  {
    title: '上课点名数量',
    dataIndex: 'classCallNum',
    key: 'classCallNum',
    width: 160,
  },
  {
    title: '消耗数量',
    dataIndex: 'useNum',
    key: 'useNum',
    width: 140,
  },
  {
    title: '拖欠数量',
    dataIndex: 'oweNum',
    key: 'oweNum',
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
    title: '点名更新时间',
    key: 'callupdateTime',
    dataIndex: 'callupdateTime',
    width: 200,
  },
  {
    title: '对内备注',
    dataIndex: 'externalRemarks',
    key: 'externalRemarks',
    width: 140,
  },
  {
    title: '对外备注',
    dataIndex: 'remarks',
    key: 'remarks',
    width: 140,
  },
  {
    title: '操作',
    dataIndex: 'action',
    key: 'action',
    fixed: 'right',
    width: 140,
  },
])

const { selectedValues, columnOptions, filteredColumns, totalWidth } = useTableColumns({
  storageKey: 'student-latitude',
  allColumns,
  excludeKeys: ['action'],
})

const rowSelection = {
  onChange: (_selectedRowKeys: (string | number)[], _selectedRows: StudentTeachingRecordItem[]) => {},
}

const tablePagination = computed(() => ({
  current: pagination.value.current,
  pageSize: pagination.value.pageSize,
  total: pagination.value.total,
  showSizeChanger: true,
  showQuickJumper: true,
  showTotal: (total: number) => `共 ${total} 条`,
}))

function formatDateTimeRange(record: Partial<StudentTeachingRecordItem> | Record<string, any>) {
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

function sourceTypeText(value?: number) {
  const type = Number(value || 0)
  if (type === 2)
    return '临时学员'
  if (type === 3 || type === 7)
    return '补课学员'
  if (type === 4)
    return '试听学员'
  if (type === 6)
    return '1对1学员'
  return '班课学员'
}

function statusText(value?: number) {
  const status = Number(value || 0)
  if (status === 2)
    return '旷课'
  if (status === 3)
    return '请假'
  if (status === 4)
    return '未记录'
  return '到课'
}

function scheduleTypeText(value?: number) {
  const type = Number(value || 0)
  if (type === 2)
    return '1对1日程'
  if (type === 3)
    return '试听日程'
  return '班级日程'
}

function chargingModeText(value?: number) {
  const mode = Number(value || 0)
  if (mode === 2)
    return '按时间'
  if (mode === 3)
    return '按金额'
  return '按课时'
}

function classDisplay(record: Partial<StudentTeachingRecordItem> | Record<string, any>) {
  return record.className || record.one2OneName || '-'
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
    const res = await getStudentTeachingRecordPagedListApi({
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
        totalTuition: Number(res.result?.totalTuition || 0),
      }
      pagination.value.total = Number(res.result?.total || 0)
      return
    }
    dataSource.value = []
    summary.value = { total: 0, totalClassTimes: 0, totalTuition: 0 }
    pagination.value.total = 0
  }
  catch (error) {
    console.error('load student teaching records failed', error)
    dataSource.value = []
    summary.value = { total: 0, totalClassTimes: 0, totalTuition: 0 }
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
    <div class="filter-wrap bg-white  pl-3 pr-3 rounded-lb-4 rounded-rb-4">
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
            共 {{ summary.total }} 条记录 ，共记录 {{ formatNumber(summary.totalClassTimes, '课时') }}，共消耗学费 {{ formatCurrency(summary.totalTuition) }}
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
            row-key="studentTeachingRecordId"
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
                    {{ formatDateTimeRange(record).dateText }}
                  </div>
                  <div class="text-3 text-#888 flex flex-items-center">
                    {{ formatDateTimeRange(record).timeText }}
                  </div>
                </div>
              </template>
              <template v-if="column.key === 'name'">
                <div class="flex">
                  <img
                    width="40" height="40" class="mr-2" style="border-radius: 100%;"
                    :src="record.avatar || 'https://cdn.schoolpal.cn/schoolpal/next-erp/avator_male.png?x-oss-process=image/resize,w_120'"
                    alt=""
                  >
                  <div class="name mt-1">
                    <div class="text-#222">
                      {{ record.studentName || '-' }}
                    </div>
                    <div class="text-3 text-#888 flex flex-items-center">
                      {{ record.studentPhone || '-' }}
                    </div>
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
              <template v-if="column.key === 'studentIdentity'">
                {{ sourceTypeText(record.sourceType) }}
              </template>
              <template v-if="column.key === 'classStatus'">
                {{ statusText(record.status) }}
              </template>
              <template v-if="column.key === 'deductionAccount'">
                {{ record.tuitionAccountName || '-' }}
              </template>
              <template v-if="column.key === 'courseNotMethod'">
                {{ chargingModeText(record.skuMode) }}
              </template>
              <template v-if="column.key === 'classCallNum'">
                {{ formatNumber(record.quantity, '课时') }}
              </template>
              <template v-if="column.key === 'useNum'">
                {{ formatNumber(record.actualQuantity, '课时') }}
              </template>
              <template v-if="column.key === 'oweNum'">
                {{ formatNumber(record.arrearQuantity, '课时') }}
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
              <template v-if="column.key === 'callupdateTime'">
                {{ record.updatedTime ? dayjs(record.updatedTime).format('YYYY-MM-DD HH:mm') : '-' }}
              </template>
              <template v-if="column.key === 'externalRemarks'">
                {{ record.remark || '-' }}
              </template>
              <template v-if="column.key === 'remarks'">
                {{ record.externalRemark || '-' }}
              </template>
              <template v-if="column.key === 'action'">
                <a class="font500" @click="handleSeeClassRecord()">上课记录详情</a>
              </template>
            </template>
          </a-table>
        </div>
      </div>
    </div>
    <class-record-details v-model:open="openClassRecordDrawer" />
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

.studentStatus {
  span.dot {
    border-radius: 50%;
    display: inline-block;
    height: 6px;
    margin-right: 6px;
    width: 6px;
  }
}
</style>
