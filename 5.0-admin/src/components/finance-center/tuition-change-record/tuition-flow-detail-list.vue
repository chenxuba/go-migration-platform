<script setup>
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { CloseOutlined, DeleteOutlined } from '@ant-design/icons-vue'
import { debounce } from 'lodash-es'
import { useTableColumns } from '@/composables/useTableColumns'
import { getCourseIdAndNameApi } from '@/api/edu-center/registr-renewal'
import { getRecommenderPageApi } from '@/api/enroll-center/intention-student'
import { getSubTuitionAccountFlowRecordListApi } from '@/api/finance-center/tuition-account-flow'
import messageService from '@/utils/messageService'

const FLOW_TYPE_GROUPS = [
  { id: 'registration', name: '报名学费', label: '报名学费', channelList: [{ id: 1, name: '报名' }] },
  { id: 'transferIn', name: '转入学费', label: '转入学费', channelList: [{ id: 2, name: '转入' }, { id: 3, name: '跨校转入' }, { id: 4, name: '跨校上课转入' }] },
  { id: 'return', name: '返还学费', label: '返还学费', channelList: [{ id: 5, name: '课消退还' }, { id: 6, name: '撤销结课' }, { id: 7, name: '过期撤回返还' }, { id: 8, name: '撤回退课订单' }, { id: 9, name: '撤销转出' }, { id: 10, name: '撤回导入课消' }, { id: 11, name: '撤回每日自动课消' }] },
  { id: 'consume', name: '课消学费', label: '课消学费', channelList: [{ id: 12, name: '课消' }, { id: 13, name: '导入课消' }, { id: 14, name: '课消补扣' }, { id: 15, name: '每日自动课消' }, { id: 16, name: '课消欠费清算' }] },
  { id: 'transferOut', name: '转出学费', label: '转出学费', channelList: [{ id: 17, name: '转出' }, { id: 18, name: '跨校转出' }, { id: 19, name: '跨校上课转出' }] },
  { id: 'graduate', name: '结课学费', label: '结课学费', channelList: [{ id: 20, name: '结课' }, { id: 21, name: '到期结算' }, { id: 25, name: '手动结课' }] },
  { id: 'refund', name: '退费学费', label: '退费学费', channelList: [{ id: 22, name: '退费' }] },
  { id: 'void', name: '作废学费', label: '作废学费', channelList: [{ id: 23, name: '订单作废' }, { id: 24, name: '作废跨校转入' }] },
]

const LESSON_CHARGING_UNIT_MAP = {
  1: '课时',
  2: '天',
  3: '元',
}

const sourceTypeLabelMap = FLOW_TYPE_GROUPS.reduce((acc, group) => {
  group.channelList.forEach((item) => {
    acc[item.id] = item.name
  })
  return acc
}, {})

const sourceTypeDirectionMap = {
  1: 'in',
  2: 'in',
  3: 'in',
  4: 'in',
  5: 'in',
  6: 'in',
  7: 'in',
  8: 'in',
  9: 'in',
  10: 'in',
  11: 'in',
  12: 'out',
  13: 'out',
  14: 'out',
  15: 'out',
  16: 'out',
  17: 'out',
  18: 'out',
  19: 'out',
  20: 'out',
  21: 'out',
  22: 'out',
  23: 'out',
  24: 'out',
  25: 'out',
}

const filterState = ref({
  productId: '',
  studentId: '',
  sourceTypePaths: [],
  changeTime: [],
  orderNumber: '',
})

const searchKeyStuPhoneModel = ref(undefined)
const selectedStudentOption = ref(null)
const stuPhoneSearchOptions = ref([])
const stuPhoneSearchFinished = ref(false)
const stuPhoneSearchLoading = ref(false)
const stuPhoneSearchPagination = ref({
  current: 1,
  pageSize: 10,
  total: 0,
})
const courseOptions = ref([])
const courseFinished = ref(false)
const dataSource = ref([])
const loading = ref(false)
const pagination = ref({
  current: 1,
  pageSize: 20,
  total: 0,
  showSizeChanger: true,
  showQuickJumper: true,
  showTotal: total => `共 ${total} 条`,
})

const allColumns = ref([
  { title: '变动时间', dataIndex: 'createdTime', key: 'createdTime', fixed: 'left', width: 160, required: true },
  { title: '学员/电话', dataIndex: 'studentName', key: 'studentName', fixed: 'left', width: 180, required: true },
  { title: '扣费课程账户', dataIndex: 'productName', key: 'productName', width: 160 },
  { title: '变动类型', dataIndex: 'sourceType', key: 'sourceType', width: 140 },
  { title: '变动数量', dataIndex: 'quantity', key: 'quantity', width: 150 },
  { title: '变动数量对应学费（元）', dataIndex: 'tuition', key: 'tuition', width: 180 },
  { title: '变动后剩余数量', dataIndex: 'balanceQuantity', key: 'balanceQuantity', width: 160 },
  { title: '变动后剩余学费（元）', dataIndex: 'balanceTuition', key: 'balanceTuition', width: 190 },
  { title: '订单编号', dataIndex: 'orderNumber', key: 'orderNumber', width: 220 },
  { title: '操作', dataIndex: 'action', key: 'action', fixed: 'right', width: 140 },
])

const { selectedValues, columnOptions, filteredColumns, totalWidth } = useTableColumns({
  storageKey: 'tuition-change-detail-list',
  allColumns,
  excludeKeys: ['action'],
})

const selectedSourceTypes = computed(() => {
  return (Array.isArray(filterState.value.sourceTypePaths) ? filterState.value.sourceTypePaths : [])
    .map(path => Array.isArray(path) ? path[path.length - 1] : undefined)
    .filter(Boolean)
})

const selectedFilterList = computed(() => {
  const list = []
  if (filterState.value.productId) {
    list.push({ key: 'productId', label: '扣费课程账户', value: courseOptions.value.find(item => String(item.id) === String(filterState.value.productId))?.value || filterState.value.productId })
  }
  if (selectedSourceTypes.value.length) {
    list.push({ key: 'sourceTypes', label: '变动类型', value: selectedSourceTypes.value.map(id => sourceTypeLabelMap[id]).filter(Boolean).join('、') })
  }
  if (Array.isArray(filterState.value.changeTime) && filterState.value.changeTime.filter(Boolean).length === 2) {
    list.push({ key: 'changeTime', label: '变动时间', value: `${filterState.value.changeTime[0]} 至 ${filterState.value.changeTime[1]}` })
  }
  if (filterState.value.orderNumber) {
    list.push({ key: 'orderNumber', label: '订单编号', value: filterState.value.orderNumber })
  }
  if (filterState.value.studentId) {
    list.push({ key: 'studentId', label: '学员/电话', value: selectedStudentOption.value?.stuName || filterState.value.studentId })
  }
  return list
})

const hasSelectedFilters = computed(() => selectedFilterList.value.length > 0)

watch(filterState, () => {
  pagination.value.current = 1
  fetchFlowDetailList()
}, { deep: true })

onUnmounted(() => {
  debouncedSearchStuPhone.cancel()
})

function formatDate(dateStr) {
  if (!dateStr) return '-'
  return String(dateStr).replace('T', ' ').slice(0, 16)
}

function formatChangeValue(value, sourceType, mode) {
  const direction = sourceTypeDirectionMap[sourceType]
  const prefix = direction === 'out' ? '-' : '+'
  // 接口可能已存负数（扣减），避免与 prefix 叠加成「--29」
  const num = Math.abs(Number(value || 0))
  const unit = LESSON_CHARGING_UNIT_MAP[mode] || ''
  return `${prefix}${num}${unit}`
}

function formatChangeMoney(value, sourceType) {
  const direction = sourceTypeDirectionMap[sourceType]
  const prefix = direction === 'out' ? '-' : '+'
  const num = Math.abs(Number(value || 0))
  return `${prefix}¥${num.toFixed(2)}`
}

function formatBalanceValue(value, mode) {
  const num = Number(value || 0)
  const unit = LESSON_CHARGING_UNIT_MAP[mode] || ''
  return `${num}${unit}`
}

function formatBalanceMoney(value) {
  const num = Number(value || 0)
  return `¥${num.toFixed(2)}`
}

function getChangeValueClass(sourceType) {
  return sourceTypeDirectionMap[sourceType] === 'out' ? 'change-out' : 'change-in'
}

function getSourceTypeText(sourceType) {
  return sourceTypeLabelMap[sourceType] || '-'
}

function clearFilter(key) {
  if (key === 'changeTime') {
    filterState.value.changeTime = []
    return
  }
  if (key === 'sourceTypes') {
    filterState.value.sourceTypePaths = []
    return
  }
  if (key === 'studentId') {
    filterState.value.studentId = ''
    searchKeyStuPhoneModel.value = undefined
    selectedStudentOption.value = null
    return
  }
  filterState.value[key] = ''
}

function clearAllFilters() {
  filterState.value = {
    productId: '',
    studentId: '',
    sourceTypePaths: [],
    changeTime: [],
    orderNumber: '',
  }
  searchKeyStuPhoneModel.value = undefined
  selectedStudentOption.value = null
}

async function loadCourseOptions(searchKey = '') {
  try {
    const res = await getCourseIdAndNameApi({
      pageRequestModel: {
        needTotal: true,
        pageSize: 50,
        pageIndex: 1,
        skipCount: 0,
      },
      queryModel: {
        delFlag: false,
        productType: 1,
        searchKey,
        saleStatus: true,
      },
      sortModel: {},
    })
    if (res.code === 200) {
      const resultData = Array.isArray(res.result) ? res.result : []
      courseOptions.value = resultData.map(item => ({
        id: item.id,
        value: item.name || '未命名课程',
        ...item,
      }))
      courseFinished.value = true
    }
  }
  catch (error) {
    console.error('加载课程列表失败:', error)
  }
}

async function getStuPhoneSearchPage(params = { key: undefined, studentStatus: 1 }) {
  try {
    if (stuPhoneSearchFinished.value && stuPhoneSearchPagination.value.current !== 1) return
    stuPhoneSearchLoading.value = true
    const res = await getRecommenderPageApi({
      pageRequestModel: {
        needTotal: true,
        pageSize: stuPhoneSearchPagination.value.pageSize,
        pageIndex: stuPhoneSearchPagination.value.current,
        skipCount: 0,
      },
      queryModel: {
        searchKey: params.key,
        studentStatus: params.studentStatus,
      },
      sortModel: {},
    })
    if (res.code === 200) {
      const resultData = Array.isArray(res.result) ? res.result : []
      if (stuPhoneSearchPagination.value.current === 1) stuPhoneSearchOptions.value = resultData
      else stuPhoneSearchOptions.value = [...stuPhoneSearchOptions.value, ...resultData]
      stuPhoneSearchPagination.value.total = res.total || 0
      stuPhoneSearchFinished.value = stuPhoneSearchOptions.value.length >= stuPhoneSearchPagination.value.total
    }
  }
  catch (error) {
    console.error('加载学员/电话搜索数据失败:', error)
  }
  finally {
    stuPhoneSearchLoading.value = false
  }
}

function dropdownVisibleChangeFun(event) {
  if (!event) return
  stuPhoneSearchPagination.value.current = 1
  stuPhoneSearchOptions.value = []
  stuPhoneSearchFinished.value = false
  getStuPhoneSearchPage({ studentStatus: 1 })
}

const debouncedSearchStuPhone = debounce((value) => {
  stuPhoneSearchPagination.value.current = 1
  stuPhoneSearchFinished.value = false
  stuPhoneSearchOptions.value = []
  getStuPhoneSearchPage({ key: value, studentStatus: 1 })
}, 300)

function handleSearchStuPhone(value) {
  debouncedSearchStuPhone(value)
}

function handleStuPhonePopupScroll(event) {
  const { target } = event
  const { scrollTop, scrollHeight, clientHeight } = target
  if (scrollHeight - scrollTop - clientHeight < 1) {
    if (!stuPhoneSearchLoading.value && stuPhoneSearchPagination.value.current * stuPhoneSearchPagination.value.pageSize < stuPhoneSearchPagination.value.total) {
      stuPhoneSearchPagination.value.current += 1
      getStuPhoneSearchPage({ studentStatus: 1 })
    }
  }
}

function handleStuPhoneSearchChange(value) {
  const id = typeof value === 'object' && value !== null ? value.value : value
  if (id) selectedStudentOption.value = stuPhoneSearchOptions.value.find(user => String(user.id) === String(id)) || null
  else selectedStudentOption.value = null
  nextTick(() => {
    filterState.value.studentId = id || ''
  })
}

async function fetchFlowDetailList() {
  try {
    loading.value = true
    const { result } = await getSubTuitionAccountFlowRecordListApi({
      queryModel: {
        productId: filterState.value.productId || undefined,
        studentId: filterState.value.studentId || undefined,
        sourceTypes: selectedSourceTypes.value.length ? selectedSourceTypes.value : undefined,
        startTime: filterState.value.changeTime?.[0] || undefined,
        endTime: filterState.value.changeTime?.[1] || undefined,
        orderNumber: filterState.value.orderNumber || undefined,
      },
      pageRequestModel: {
        needTotal: true,
        pageSize: pagination.value.pageSize,
        pageIndex: pagination.value.current,
        skipCount: (pagination.value.current - 1) * pagination.value.pageSize,
      },
      sortModel: {
        orderByCreatedTime: 0,
      },
    })
    dataSource.value = Array.isArray(result?.list) ? result.list : []
    pagination.value.total = result?.total || 0
  }
  catch (error) {
    console.error('获取学费变动明细失败:', error)
    messageService.error('获取学费变动明细失败')
  }
  finally {
    loading.value = false
  }
}

function handleTableChange(pag) {
  pagination.value.current = pag.current
  pagination.value.pageSize = pag.pageSize
  fetchFlowDetailList()
}

function handleOpenClassRecord(record) {
  if (!record?.teachingRecordId) {
    messageService.info('暂无上课记录详情')
    return
  }
  messageService.info('上课记录详情待实现')
}

onMounted(async () => {
  await loadCourseOptions()
  fetchFlowDetailList()
})
</script>

<template>
  <div>
    <div class="filter-wrap bg-white pl-3 pr-3 pb-3 rounded-lb-4 rounded-rb-4">
      <div class="filter-section pt-3">
        <span class="section-title mt-0.5 text-#222">筛选条件：</span>
        <div class="filter-toolbar">
          <div class="standard-filters">
            <checkbox-filter
              v-model:checked-values="filterState.productId"
              :options="courseOptions"
              label="扣费课程账户"
              type="radio"
              category="course"
              placeholder="请输入课程..."
              :finished="courseFinished"
              @on-dropdown-visible-change="loadCourseOptions"
              @on-search="loadCourseOptions"
            />
            <checkbox-filter
              v-model:checked-values="filterState.sourceTypePaths"
              :options="FLOW_TYPE_GROUPS"
              label="变动类型"
              type="cascader"
              placeholder="搜索变动类型"
            />
            <checkbox-filter
              v-model:checked-values="filterState.changeTime"
              label="变动时间"
              type="dateTime"
            />
            <checkbox-filter
              v-model:checked-values="filterState.orderNumber"
              label="订单编号"
              type="inputType"
              placeholder="请输入订单编号"
            />
          </div>
          <div class="student-filter-wrap">
            <div class="label">
              学员/电话
            </div>
            <a-select
              v-model:value="searchKeyStuPhoneModel"
              class="searchKeyStuPhone"
              allow-clear
              show-search
              placeholder="搜索姓名/手机号"
              :filter-option="false"
              style="width: 240px"
              option-label-prop="label"
              :label-in-value="true"
              @change="handleStuPhoneSearchChange"
              @dropdown-visible-change="dropdownVisibleChangeFun"
              @search="handleSearchStuPhone"
              @popup-scroll="handleStuPhonePopupScroll"
            >
              <a-select-option
                v-for="item in stuPhoneSearchOptions"
                :key="item.id"
                :value="item.id"
                :data="item"
                :label="item.stuName"
              >
                <div class="flex flex-center mb-1">
                  <div>
                    <img class="w-10 rounded-10" :src="item.avatarUrl" alt="">
                  </div>
                  <div class="ml-2 mr-3">
                    <div class="text-sm text-#666 leading-7">
                      {{ item.stuName }}
                    </div>
                    <div class="text-xs text-#888">
                      {{ item.mobile }}
                    </div>
                  </div>
                  <div>
                    <a-tag v-if="item.studentStatus == 1" :bordered="false" color="processing">
                      在读学员
                    </a-tag>
                    <a-tag v-if="item.studentStatus == 0" :bordered="false" color="orange">
                      意向学员
                    </a-tag>
                  </div>
                </div>
              </a-select-option>
            </a-select>
          </div>
        </div>
      </div>
      <div v-if="hasSelectedFilters" class="selected-conditions">
        <span class="section-title text-#222">已选条件：</span>
        <div class="condition-tags">
          <a-popconfirm title="确定要清空所有条件吗？" @confirm="clearAllFilters">
            <a-tag color="red" class="clear-all mb-2">
              清空已选
              <DeleteOutlined class="text-3 ml-4px mt-0.6px" />
            </a-tag>
          </a-popconfirm>
          <a-tag v-for="item in selectedFilterList" :key="item.key" color="blue" class="condition-tag mb-2">
            <div class="tag-content">
              <span class="condition-label">{{ item.label }}：</span>
              <div class="condition-values">
                <span class="value-item">
                  {{ item.value }}
                  <CloseOutlined class="close-icon" @click.stop="clearFilter(item.key)" />
                </span>
              </div>
            </div>
          </a-tag>
        </div>
      </div>
    </div>
    <div class="student-list mt-3 pt-3 pb-3 pl-6 pr-6 bg-white rounded-4">
      <div class="tab-table">
        <div class="table-title flex justify-between">
          <div class="total">
            当前共计 {{ pagination.total }} 条记录
          </div>
          <div class="edit flex">
            <a-button class="mr-2">
              导出数据
            </a-button>
            <customize-code
              v-model:checked-values="selectedValues"
              :options="columnOptions"
              :total="allColumns.length - 1"
              :num="selectedValues.length"
            />
          </div>
        </div>
        <div class="table-content mt-2">
          <a-table
            :data-source="dataSource"
            :pagination="pagination"
            :columns="filteredColumns"
            :scroll="{ x: totalWidth }"
            :loading="loading"
            row-key="id"
            size="small"
            @change="handleTableChange"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'createdTime'">
                {{ formatDate(record.createdTime) }}
              </template>
              <template v-if="column.key === 'studentName'">
                <student-avatar
                  :id="record.studentId"
                  :name="record.studentName || '-'"
                  :phone="record.studentPhone || ''"
                  :avatar-url="record.studentAvatar"
                  :show-gender="false"
                  :show-age="false"
                  default-active-key="0"
                />
              </template>
              <template v-if="column.key === 'productName'">
                {{ record.productName || '-' }}
              </template>
              <template v-if="column.key === 'sourceType'">
                {{ getSourceTypeText(record.sourceType) }}
              </template>
              <template v-if="column.key === 'quantity'">
                <span :class="getChangeValueClass(record.sourceType)">
                  {{ formatChangeValue(record.quantity, record.sourceType, record.lessonChargingMode) }}
                </span>
              </template>
              <template v-if="column.key === 'tuition'">
                <span :class="getChangeValueClass(record.sourceType)">
                  {{ formatChangeMoney(record.tuition, record.sourceType) }}
                </span>
              </template>
              <template v-if="column.key === 'balanceQuantity'">
                {{ formatBalanceValue(record.balanceQuantity, record.lessonChargingMode) }}
              </template>
              <template v-if="column.key === 'balanceTuition'">
                {{ formatBalanceMoney(record.balanceTuition) }}
              </template>
              <template v-if="column.key === 'orderNumber'">
                {{ record.orderNumber || '-' }}
              </template>
              <template v-if="column.key === 'action'">
                <a v-if="record.teachingRecordId" @click="handleOpenClassRecord(record)">上课记录详情</a>
                <span v-else class="text-#999">-</span>
              </template>
            </template>
          </a-table>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped lang="less">
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

.filter-section {
  display: flex;
  align-items: flex-start;
  }

.section-title {
  white-space: nowrap;
}

.standard-filters {
  display: flex;
  flex-wrap: wrap;
}

.filter-toolbar {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  width: 100%;
}

.student-filter-wrap {
  align-items: center;
  display: flex;
  margin-left: auto;
  padding-right: 8px;
}

.student-filter-wrap .label {
  border: 1px solid #f0f0f0;
  border-radius: 8px 0 0 8px !important;
  color: #222;
  font-size: 14px;
  height: 32px;
  line-height: 32px;
  min-width: 104px;
  padding: 0 16px 0 8px;
  text-align: center;
  white-space: nowrap;
  border-right: 0;
}

:deep(.student-filter-wrap .ant-select-selector) {
  border-radius: 0 6px 6px 0 !important;
}

.selected-conditions {
  display: flex;
  align-items: flex-start;
}

.condition-tags {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
}

.condition-tag {
  display: flex;
  align-items: center;
  border-radius: 4px;
}

.tag-content {
  display: flex;
  align-items: center;
}

.condition-values {
  display: flex;
  align-items: center;
}

.value-item {
  display: inline-flex;
  align-items: center;
}

.close-icon {
  margin-left: 6px;
  font-size: 12px;
  cursor: pointer;
  color: rgba(92, 92, 92, 0.45);
  transition: color 0.3s;
}

.close-icon:hover {
  color: rgba(0, 0, 0, 0.75);
}

.clear-all {
  cursor: pointer;
  display: flex;
  align-items: center;
}

.change-in {
  color: inherit;
  font-weight: inherit;
}

.change-out {
  color: #ff4d4f;
  font-weight: 500;
}
</style>
