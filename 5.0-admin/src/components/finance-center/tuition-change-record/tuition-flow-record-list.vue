<script setup>
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { CloseOutlined, DeleteOutlined, DownOutlined } from '@ant-design/icons-vue'
import { debounce } from 'lodash-es'
import { useTableColumns } from '@/composables/useTableColumns'
import { getCourseIdAndNameApi } from '@/api/edu-center/registr-renewal'
import { getTuitionAccountFlowRecordListApi } from '@/api/finance-center/tuition-account-flow'
import { getRecommenderPageApi } from '@/api/enroll-center/intention-student'
import messageService from '@/utils/messageService'

const FLOW_TYPE_GROUPS = [
  {
    id: 'registration',
    label: '报名学费',
    children: [
      { id: 1, label: '报名', direction: 'in' },
    ],
  },
  {
    id: 'transferIn',
    label: '转入学费',
    children: [
      { id: 2, label: '转入', direction: 'in' },
      { id: 3, label: '跨校转入', direction: 'in' },
      { id: 4, label: '跨校上课转入', direction: 'in' },
    ],
  },
  {
    id: 'return',
    label: '返还学费',
    children: [
      { id: 5, label: '课消退还', direction: 'in' },
      { id: 6, label: '撤销结课', direction: 'in' },
      { id: 7, label: '过期撤回返还', direction: 'in' },
      { id: 8, label: '撤回退课订单', direction: 'in' },
      { id: 9, label: '撤销转出', direction: 'in' },
      { id: 10, label: '撤回导入课消', direction: 'in' },
      { id: 11, label: '撤回每日自动课消', direction: 'in' },
    ],
  },
  {
    id: 'consume',
    label: '课消学费',
    children: [
      { id: 12, label: '课消', direction: 'out' },
      { id: 13, label: '导入课消', direction: 'out' },
      { id: 14, label: '课消补扣', direction: 'out' },
      { id: 15, label: '每日自动课消', direction: 'out' },
      { id: 16, label: '课消欠费清算', direction: 'out' },
    ],
  },
  {
    id: 'transferOut',
    label: '转出学费',
    children: [
      { id: 17, label: '转出', direction: 'out' },
      { id: 18, label: '跨校转出', direction: 'out' },
      { id: 19, label: '跨校上课转出', direction: 'out' },
    ],
  },
  {
    id: 'graduate',
    label: '结课学费',
    children: [
      { id: 20, label: '结课', direction: 'out' },
      { id: 21, label: '到期结算', direction: 'out' },
      { id: 25, label: '手动结课', direction: 'out' },
    ],
  },
  {
    id: 'refund',
    label: '退费学费',
    children: [
      { id: 22, label: '退费', direction: 'out' },
    ],
  },
  {
    id: 'void',
    label: '作废学费',
    children: [
      { id: 23, label: '订单作废', direction: 'out' },
      { id: 24, label: '作废跨校转入', direction: 'out' },
    ],
  },
]

const LESSON_TYPE_MAP = {
  1: '班级授课',
  2: '1v1授课',
}

const LESSON_CHARGING_MODE_MAP = {
  1: '按课时',
  2: '按时段',
  3: '按金额',
}

const LESSON_CHARGING_UNIT_MAP = {
  1: '课时',
  2: '天',
  3: '元',
}

const sourceTypeLabelMap = FLOW_TYPE_GROUPS.reduce((acc, group) => {
  group.children.forEach((child) => {
    acc[child.id] = child.label
  })
  return acc
}, {})

const sourceTypeDirectionMap = FLOW_TYPE_GROUPS.reduce((acc, group) => {
  group.children.forEach((child) => {
    acc[child.id] = child.direction
  })
  return acc
}, {})

const sourceTypeCascaderOptions = FLOW_TYPE_GROUPS.map(group => ({
  id: group.id,
  name: group.label,
  channelList: group.children.map(child => ({
    id: child.id,
    name: child.label,
  })),
}))

const filterState = ref({
  productId: '',
  studentId: '',
  sourceTypePaths: [],
  changeTime: [],
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
  { title: '上课课程', dataIndex: 'teachingCourseName', key: 'teachingCourseName', width: 140 },
  { title: '扣费课程账户', dataIndex: 'productName', key: 'productName', width: 160 },
  { title: '授课方式', dataIndex: 'lessonType', key: 'lessonType', width: 120 },
  { title: '收费方式', dataIndex: 'lessonChargingMode', key: 'lessonChargingMode', width: 120 },
  { title: '变动类型', dataIndex: 'sourceType', key: 'sourceType', width: 140 },
  { title: '变动数量', dataIndex: 'quantity', key: 'quantity', width: 150 },
  { title: '变动数量对应学费（元）', dataIndex: 'tuition', key: 'tuition', width: 170 },
  { title: '操作', dataIndex: 'action', key: 'action', fixed: 'right', width: 140 },
])

const { selectedValues, columnOptions, filteredColumns, totalWidth } = useTableColumns({
  storageKey: 'tuition-change-record-list',
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
    list.push({
      key: 'productId',
      label: '扣费课程账户',
      value: courseOptions.value.find(item => String(item.id) === String(filterState.value.productId))?.value || filterState.value.productId,
    })
  }
  if (filterState.value.studentId) {
    list.push({
      key: 'studentId',
      label: '学员/电话',
      value: selectedStudentOption.value?.stuName || filterState.value.studentId,
    })
  }
  if (selectedSourceTypes.value.length) {
    list.push({
      key: 'sourceTypes',
      label: '变动类型',
      value: selectedSourceTypes.value.map(id => sourceTypeLabelMap[id]).filter(Boolean).join('、'),
    })
  }
  if (Array.isArray(filterState.value.changeTime) && filterState.value.changeTime.filter(Boolean).length === 2) {
    list.push({
      key: 'changeTime',
      label: '变动时间',
      value: `${filterState.value.changeTime[0]} 至 ${filterState.value.changeTime[1]}`,
    })
  }
  return list
})

const hasSelectedFilters = computed(() => selectedFilterList.value.length > 0)

watch(filterState, () => {
  pagination.value.current = 1
  fetchFlowRecordList()
}, { deep: true })

onUnmounted(() => {
  debouncedSearchStuPhone.cancel()
})

function formatDate(dateStr) {
  if (!dateStr)
    return '-'
  return String(dateStr).replace('T', ' ').slice(0, 16)
}

function formatChangeValue(value, sourceType, mode) {
  const direction = sourceTypeDirectionMap[sourceType]
  const prefix = direction === 'out' ? '-' : '+'
  const num = Math.abs(Number(value || 0))
  const unit = LESSON_CHARGING_UNIT_MAP[mode] || ''
  return `${prefix}${num}${unit}`
}

function formatChangeTuition(value, sourceType) {
  const direction = sourceTypeDirectionMap[sourceType]
  const prefix = direction === 'out' ? '-' : '+'
  return `${prefix}${Math.abs(Number(value || 0)).toFixed(2)}`
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
    if (stuPhoneSearchFinished.value && stuPhoneSearchPagination.value.current !== 1)
      return
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
      if (stuPhoneSearchPagination.value.current === 1)
        stuPhoneSearchOptions.value = resultData
      else
        stuPhoneSearchOptions.value = [...stuPhoneSearchOptions.value, ...resultData]
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
  if (!event)
    return
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
  if (id) {
    selectedStudentOption.value = stuPhoneSearchOptions.value.find(user => String(user.id) === String(id)) || null
  }
  else {
    selectedStudentOption.value = null
  }
  nextTick(() => {
    filterState.value.studentId = id || ''
  })
}

async function fetchFlowRecordList() {
  try {
    loading.value = true
    const { result } = await getTuitionAccountFlowRecordListApi({
      queryModel: {
        productId: filterState.value.productId || undefined,
        studentId: filterState.value.studentId || undefined,
        sourceTypes: selectedSourceTypes.value.length ? selectedSourceTypes.value : undefined,
        startTime: filterState.value.changeTime?.[0] || undefined,
        endTime: filterState.value.changeTime?.[1] || undefined,
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
    console.error('获取学费变动记录失败:', error)
    messageService.error('获取学费变动记录失败')
  }
  finally {
    loading.value = false
  }
}

function handleTableChange(pag) {
  pagination.value.current = pag.current
  pagination.value.pageSize = pag.pageSize
  fetchFlowRecordList()
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
  fetchFlowRecordList()
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
              :options="sourceTypeCascaderOptions"
              label="变动类型"
              type="cascader"
              placeholder="搜索变动类型"
            />
            <checkbox-filter
              v-model:checked-values="filterState.changeTime"
              label="变动时间"
              type="dateTime"
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
              清空所有
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
            row-key="tutionAccountFlowId"
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
              <template v-if="column.key === 'teachingCourseName'">
                {{ record.teachingCourseName || '-' }}
              </template>
              <template v-if="column.key === 'lessonType'">
                {{ LESSON_TYPE_MAP[record.lessonType] || '-' }}
              </template>
              <template v-if="column.key === 'lessonChargingMode'">
                {{ LESSON_CHARGING_MODE_MAP[record.lessonChargingMode] || '-' }}
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
                  {{ formatChangeTuition(record.tuition, record.sourceType) }}
                </span>
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

.selectBox {
  justify-content: flex-end;
  align-items: center;
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
