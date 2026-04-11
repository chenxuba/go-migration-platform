<script setup lang="ts">
import { ExclamationCircleOutlined } from '@ant-design/icons-vue'
import { computed, ref, watch } from 'vue'
import type { TeachingRecordDetailResult, TeachingRecordDetailStudent } from '@/api/edu-center/class-record'
import StudentAvatar from '@/components/common/StudentAvatar.vue'
import { useStudentStore } from '@/stores/student'

const props = withDefaults(defineProps<{
  detail?: TeachingRecordDetailResult | null
  loading?: boolean
}>(), {
  detail: null,
  loading: false,
})

const displayArray = ref(['studentIdentity', 'classStatus', 'billingMode'])
const searchKeyword = ref('')
const filterStudentIdentityValues = ref<string[]>([])
const filterClassStatusValues = ref<string[]>([])
const filterBillingModeValues = ref<number[]>([])
const openDrawer = ref(false)
const studentStore = useStudentStore()

const allColumns = ref<any[]>([
  {
    title: '学员/电话',
    dataIndex: 'name',
    key: 'name',
    fixed: 'left',
    width: 180,
    required: true,
  },
  {
    title: '详情',
    key: 'detail',
    dataIndex: 'detail',
    fixed: 'left',
    width: 80,
  },
  {
    title: '学员身份',
    dataIndex: 'studentIdentity',
    key: 'studentIdentity',
    width: 120,
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
    width: 120,
  },
  {
    title: '上课点名数量',
    dataIndex: 'classCallNum',
    key: 'classCallNum',
    width: 120,
  },
  {
    title: '消耗数量',
    dataIndex: 'useNum',
    key: 'useNum',
    width: 120,
  },
  {
    title: '拖欠数量',
    dataIndex: 'oweNum',
    key: 'oweNum',
    width: 120,
  },
  {
    title: '消耗学费',
    dataIndex: 'usePrice',
    key: 'usePrice',
    width: 120,
  },
  {
    title: '对内备注',
    dataIndex: 'externalRemarks',
    key: 'externalRemarks',
    width: 160,
  },
  {
    title: '对外备注',
    dataIndex: 'remarks',
    key: 'remarks',
    width: 160,
  },
  {
    title: '点名更新时间',
    key: 'callupdateTime',
    dataIndex: 'callupdateTime',
    width: 170,
    fixed: 'right',
  },
])

const savedSelected = localStorage.getItem('call-name-details')
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
  const optionalColumns = allColumns.value
    .filter(col => selectedValues.value.includes(col.key) && !col.required)
  return [
    ...requiredColumns.filter(col => col.fixed === 'left'),
    ...optionalColumns,
    ...requiredColumns.filter(col => col.fixed === 'right'),
  ]
})

watch(selectedValues, (newVal) => {
  const requiredKeys = allColumns.value
    .filter(col => col.required)
    .map(col => col.key)
  if (!requiredKeys.every(k => newVal.includes(k))) {
    selectedValues.value = Array.from(new Set([
      ...newVal.filter(v => !requiredKeys.includes(v)),
      ...requiredKeys,
    ]))
  }
}, { deep: true })

watch(selectedValues, (newVal) => {
  localStorage.setItem('call-name-details', JSON.stringify(newVal))
}, { deep: true })

const totalWidth = computed(() =>
  filteredColumns.value.reduce((acc, column) => acc + (column.width || 0), 0),
)
const tablePagination = computed(() => (filteredData.value.length > 10 ? { hideOnSinglePage: true } : false))

const studentIdentityOptions = [
  { id: 'class', value: '班级学员' },
  { id: 'one_to_one', value: '1对1学员' },
  { id: 'trial', value: '试听学员' },
  { id: 'temporary', value: '临时学员' },
  { id: 'makeup', value: '补课学员' },
]

const classStatusOptions = [
  { id: '1', value: '到课' },
  { id: '3', value: '请假' },
  { id: '2', value: '旷课' },
  { id: '0', value: '未记录' },
]

const billingModeOptions = [
  { id: 1, value: '按课时' },
  { id: 2, value: '按时间' },
  { id: 3, value: '按金额' },
  { id: 4, value: '不记课时' },
]

function formatNumber(value?: number, suffix = '') {
  const num = Number(value || 0)
  if (!Number.isFinite(num))
    return suffix ? `0${suffix}` : '0'
  const text = Number.isInteger(num) ? String(num) : num.toFixed(2).replace(/\.?0+$/, '')
  return suffix ? `${text}${suffix}` : text
}

function formatCurrency(value?: number) {
  return `¥ ${Number(value || 0).toFixed(0)}`
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
  return '班级学员'
}

function studentIdentityTagClass(value?: number) {
  const type = Number(value || 0)
  if (type === 2)
    return 'record-meta-tag record-meta-tag--temporary'
  if (type === 3 || type === 7)
    return 'record-meta-tag record-meta-tag--make-up'
  if (type === 4)
    return 'record-meta-tag record-meta-tag--trial-student'
  if (type === 6)
    return 'record-meta-tag record-meta-tag--one-to-one-student'
  return 'record-meta-tag record-meta-tag--class-student'
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

function statusTagClass(value?: number) {
  const status = Number(value || 0)
  if (status === 2)
    return 'record-status-tag record-status-tag--absent'
  if (status === 3)
    return 'record-status-tag record-status-tag--leave'
  if (status === 4)
    return 'record-status-tag record-status-tag--pending'
  return 'record-status-tag record-status-tag--arrived'
}

function chargingModeText(value?: number) {
  const mode = Number(value || 0)
  if (mode === 2)
    return '按时间'
  if (mode === 3)
    return '按金额'
  return '按课时'
}

function isTrialStudent(record: Partial<TeachingRecordDetailStudent>) {
  return Number(record.sourceType || 0) === 4
}

function isTimeChargingMode(record: Partial<TeachingRecordDetailStudent>) {
  return Number(record.skuMode || 0) === 2
}

function hasArrearQuantity(record: Partial<TeachingRecordDetailStudent>) {
  return Number(record.arrearQuantity || 0) > 0
}

function buildStudentSourceTypes(values: string[]) {
  const result = new Set<number>()
  values.forEach((item) => {
    if (item === 'class')
      result.add(5)
    else if (item === 'one_to_one')
      result.add(6)
    else if (item === 'trial')
      result.add(4)
    else if (item === 'temporary')
      result.add(2)
    else if (item === 'makeup') {
      result.add(3)
      result.add(7)
    }
  })
  return result
}

function buildClassStatusValues(values: string[]) {
  return new Set(values.map((item) => {
    if (item === '0')
      return 4
    return Number(item)
  }).filter(item => Number.isFinite(item)))
}

const rawStudentList = computed(() => Array.isArray(props.detail?.studentList) ? props.detail?.studentList || [] : [])

const filteredData = computed(() => {
  const search = searchKeyword.value.trim().toLowerCase()
  const sourceTypeSet = buildStudentSourceTypes(filterStudentIdentityValues.value)
  const statusSet = buildClassStatusValues(filterClassStatusValues.value)

  return rawStudentList.value.filter((item) => {
    if (search) {
      const name = String(item.studentName || '').toLowerCase()
      if (!name.includes(search))
        return false
    }
    if (sourceTypeSet.size > 0 && !sourceTypeSet.has(Number(item.sourceType || 0)))
      return false
    if (statusSet.size > 0 && !statusSet.has(Number(item.status || 0)))
      return false
    if (filterBillingModeValues.value.length > 0) {
      const currentMode = Number(item.skuMode || 0)
      const matches = filterBillingModeValues.value.some((mode) => {
        if (mode === 4)
          return isTrialStudent(item)
        return currentMode === mode
      })
      if (!matches)
        return false
    }
    return true
  })
})

const summary = computed(() => {
  return {
    total: Number(props.detail?.shouldAttendanceCount || 0),
    arrived: Number(props.detail?.actualAttendanceCount || 0),
    leave: Number(props.detail?.leaveCount || 0),
    absent: Number(props.detail?.truancyCount || 0),
    quantity: Number(props.detail?.studentTotalClassTime || 0),
    tuition: Number(props.detail?.studentActualTuition || 0),
  }
})

function openStudentDetail(studentId?: string) {
  const id = String(studentId || '').trim()
  if (!id)
    return
  studentStore.setStudentId(id)
  openDrawer.value = true
}

function handleSearchInput(value?: string) {
  searchKeyword.value = String(value || '').trim()
}

function handleStudentIdentityFilter(value: unknown) {
  filterStudentIdentityValues.value = Array.isArray(value)
    ? value.map(item => String(item || '').trim()).filter(Boolean)
    : []
}

function handleClassStatusFilter(value: unknown) {
  filterClassStatusValues.value = Array.isArray(value)
    ? value.map(item => String(item || '').trim()).filter(Boolean)
    : []
}

function handleBillingModeFilter(value: unknown) {
  filterBillingModeValues.value = Array.isArray(value)
    ? value.map(item => Number(item)).filter(item => Number.isFinite(item))
    : []
}
</script>

<template>
  <div>
    <div class="filter-wrap bg-white pl-3 pr-3 rounded-4">
      <all-filter
        :display-array="displayArray"
        :is-quick-show="false"
        :is-show-search-input="true"
        :student-identity-options="studentIdentityOptions"
        :class-status-options="classStatusOptions"
        :billing-mode-label="'课消方式'"
        :billing-mode-options-data="billingModeOptions"
        :whole-condition-clear-types="['studentIdentity', 'classStatus', 'billingMode']"
        search-label="学员姓名"
        @searchInputFun="handleSearchInput"
        @update:student-identity-filter="handleStudentIdentityFilter"
        @update:class-status-filter="handleClassStatusFilter"
        @update:billing-mode-filter="handleBillingModeFilter"
        @update:charging-method-filter="handleBillingModeFilter"
      />
    </div>

    <div class="student-list mt-3 pt-3 pb-3 pl-6 pr-6 bg-white rounded-4">
      <div class="tab-table">
        <div class="table-title flex justify-between">
          <div class="total">
            共 {{ summary.total }} 人（到课 {{ summary.arrived }}人，请假 {{ summary.leave }}人，旷课{{ summary.absent }}人）；共记 {{ formatNumber(summary.quantity, '课时') }}，共消耗学费 {{ formatCurrency(summary.tuition) }}
          </div>
        </div>
        <div class="table-content mt-2">
          <a-table
            :loading="loading"
            :data-source="filteredData"
            row-key="studentTeachingRecordId"
            :pagination="tablePagination"
            :columns="filteredColumns"
            :scroll="{ x: totalWidth }"
            size="small"
          >
            <template #headerCell="{ column }">
              <template v-if="column.key === 'courseNotMethod'">
                <span class="mr-1">{{ column.title }}</span>
                <a-popover placement="top">
                  <template #title>
                    <div style="border-bottom:1px solid #eee;padding-bottom: 6px;">
                      课消方式
                    </div>
                  </template>
                  <template #content>
                    <div>【课消方式】课消方式决定了点名时的记录内容。</div>
                    <div>“按课时”：可以记录课时。</div>
                    <div>“按金额”：可以记录课时和金额。</div>
                    <div>“按时间”：可以记录课时。</div>
                  </template>
                  <ExclamationCircleOutlined />
                </a-popover>
              </template>
              <template v-if="column.key === 'useNum'">
                <span class="mr-1">{{ column.title }}</span>
                <a-popover placement="top">
                  <template #title>
                    <div style="border-bottom:1px solid #eee;padding-bottom: 6px;">
                      消耗数量
                    </div>
                  </template>
                  <template #content>
                    <div>【消耗数量】当次课程真实消耗了多少课时/金额。</div>
                  </template>
                  <ExclamationCircleOutlined />
                </a-popover>
              </template>
              <template v-if="column.key === 'oweNum'">
                <span class="mr-1">{{ column.title }}</span>
                <a-popover placement="top">
                  <template #title>
                    <div style="border-bottom:1px solid #eee;padding-bottom: 6px;">
                      拖欠数量
                    </div>
                  </template>
                  <template #content>
                    <div>【拖欠数量】该学员“剩余数量 &lt; 点名数量时”，会产生“拖欠数量”。</div>
                  </template>
                  <ExclamationCircleOutlined />
                </a-popover>
              </template>
              <template v-if="column.key === 'usePrice'">
                <span class="mr-1">{{ column.title }}</span>
                <a-popover placement="top">
                  <template #title>
                    <div style="border-bottom:1px solid #eee;padding-bottom: 6px;">
                      消耗学费
                    </div>
                  </template>
                  <template #content>
                    <div>【消耗学费】本次点名数量对应的学费（钱），即机构实际确认收入。</div>
                  </template>
                  <ExclamationCircleOutlined />
                </a-popover>
              </template>
            </template>

            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'name'">
                <StudentAvatar
                  :id="record.studentId"
                  :name="record.studentName || '-'"
                  :avatar-url="record.avatar"
                  :phone="record.studentPhone"
                  :show-gender="false"
                  :show-age="false"
                  default-active-key="0"
                />
              </template>
              <template v-if="column.key === 'detail'">
                <span class="detail-link" @click="openStudentDetail(record.studentId)">详情</span>
              </template>
              <template v-if="column.key === 'studentIdentity'">
                <span :class="studentIdentityTagClass(record.sourceType)">
                  {{ sourceTypeText(record.sourceType) }}
                </span>
              </template>
              <template v-if="column.key === 'classStatus'">
                <span :class="statusTagClass(record.status)">
                  {{ statusText(record.status) }}
                </span>
              </template>
              <template v-if="column.key === 'deductionAccount'">
                {{ isTrialStudent(record) ? '-' : (record.tuitionAccountName || '-') }}
              </template>
              <template v-if="column.key === 'courseNotMethod'">
                {{ isTrialStudent(record) ? '-' : chargingModeText(record.skuMode) }}
              </template>
              <template v-if="column.key === 'classCallNum'">
                {{ isTrialStudent(record) || isTimeChargingMode(record) ? '不记课时' : formatNumber(record.quantity, '课时') }}
              </template>
              <template v-if="column.key === 'useNum'">
                {{ isTrialStudent(record) || isTimeChargingMode(record) ? '-' : formatNumber(record.actualQuantity, '课时') }}
              </template>
              <template v-if="column.key === 'oweNum'">
                <span :class="{ 'owe-num-text': hasArrearQuantity(record) }">
                  {{ isTrialStudent(record) || isTimeChargingMode(record) ? '-' : formatNumber(record.arrearQuantity, '课时') }}
                </span>
              </template>
              <template v-if="column.key === 'usePrice'">
                <span class="use-price-text">
                  {{ isTrialStudent(record) || isTimeChargingMode(record) ? '-' : `¥${Number(record.actualTuition || 0).toFixed(2)}` }}
                </span>
              </template>
              <template v-if="column.key === 'externalRemarks'">
                {{ record.remark || '-' }}
              </template>
              <template v-if="column.key === 'remarks'">
                {{ record.externalRemark || '-' }}
              </template>
              <template v-if="column.key === 'callupdateTime'">
                {{ record.updatedTime || '-' }}
              </template>
            </template>
          </a-table>
        </div>
      </div>
    </div>

    <student-info-drawer v-model:open="openDrawer" />
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

.detail-link {
  color: #1677ff;
  cursor: pointer;
}

.record-status-tag {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 48px;
  padding: 2px 10px;
  border-radius: 999px;
  font-size: 12px;
  line-height: 20px;
  white-space: nowrap;
}

.record-meta-tag {
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

.record-meta-tag--class-student {
  color: #166534;
  background: #f0fdf4;
  border-color: #bbf7d0;
}

.record-meta-tag--one-to-one-student {
  color: #1d4ed8;
  background: #eff6ff;
  border-color: #bfdbfe;
}

.record-meta-tag--trial-student {
  color: #c2410c;
  background: #fff7ed;
  border-color: #fed7aa;
}

.record-meta-tag--temporary {
  color: #7c3aed;
  background: #f5f3ff;
  border-color: #ddd6fe;
}

.record-meta-tag--make-up {
  color: #0f766e;
  background: #f0fdfa;
  border-color: #99f6e4;
}

.record-status-tag--arrived {
  color: #2f6bff;
  background: #eef4ff;
}

.record-status-tag--leave {
  color: #fa8c16;
  background: #fff4e8;
}

.record-status-tag--absent {
  color: #f5222d;
  background: #fff1f0;
}

.record-status-tag--pending {
  color: #8c8c8c;
  background: #f5f5f5;
}

.use-price-text {
  font-weight: 600;
}

.owe-num-text {
  color: #f5222d;
}
</style>
