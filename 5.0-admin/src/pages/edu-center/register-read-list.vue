<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { debounce } from 'lodash-es'
import { DownOutlined } from '@ant-design/icons-vue'
import dayjs from 'dayjs'
import { getRegisterReadListApi } from '~@/api/edu-center/register-read-list'
import emitter, { EVENTS } from '~@/utils/eventBus'
import { useStudentListRefresh } from '@/composables/useStudentListRefresh'
import { useTableColumns } from '@/composables/useTableColumns'
import messageService from '~@/utils/messageService'
import { handleDateRangeParams } from '~@/utils/dateRangeParams'
import { Sex, SexLabel } from '@/enums'

// 报读列表页面筛选条件（仅静态展示，后续根据需求再处理参数传递）
const displayArray = ref([
  // 报读课程
  'enrolledCourse',
  // 班级名称
  'className',
  // 有效期至 & 是否设置有效期
  'validityPeriod',
  'isSetExpirationDate',
  // 停课时间 & 结课时间
  'classStopTime',
  'classEndingTime',
  // 当前状态 & 分班状态
  'currentStatus',
  'orNotFenClass',
  // 授课方式 & 收费方式
  'teachingMethod',
  'billingMode',
  // 剩余数量
  'remaining',
  // 班主任 & 销售员
  'classTeacher',
  'salesPerson',
  // 是否欠费
  'isArrears',
  // 最近上课时间
  'lastClassTime',
])
const dataSource = ref([])
const loading = ref(false)
const selectedRows = ref([])
const selectedRowKeys = ref([])
const totalStats = ref({
  totalRemainedTuition: 0,
  totalConfirmedTuition: 0,
  totalPaidRemainedTuition: 0,
  total: 0,
  studentCount: 0,
})

const allColumns = ref([
  {
    title: '学员/性别',
    dataIndex: 'name',
    key: 'name',
    fixed: 'left',
    width: 160,
    required: true,
  },
  {
    title: '联系电话',
    dataIndex: 'phone',
    width: 120,
    fixed: 'left',
    key: 'phone',
  },
  {
    title: '报读课程',
    key: 'registerReadCourse',
    dataIndex: 'registerReadCourse',
    width: 160,
  },
  {
    title: '报读数量',
    key: 'registerReadNum',
    dataIndex: 'registerReadNum',
    width: 160,
  },
  {
    title: '当前状态',
    dataIndex: 'currentStatus',
    key: 'currentStatus',
    width: 120,
  },
  {
    title: '分班状态',
    dataIndex: 'orNotFenClass',
    key: 'orNotFenClass',
    width: 120,
  },
  {
    title: '班主任',
    dataIndex: 'headTeacher',
    key: 'headTeacher',
    width: 120,
  },
  {
    title: '有效期至',
    dataIndex: 'expiryDate',
    key: 'expiryDate',
    width: 150,
  },
  {
    title: '已用数量',
    dataIndex: 'usedNumber',
    key: 'usedNumber',
    width: 120,
  },
  {
    title: '剩余数量',
    dataIndex: 'remainNumber',
    key: 'remainNumber',
    width: 140,
  },
  {
    title: '已用学费金额',
    key: 'usedtuitionfeeamount',
    dataIndex: 'usedtuitionfeeamount',
    width: 140,
  },
  {
    title: '剩余学费金额',
    dataIndex: 'remaintuitionfeeamount',
    key: 'remaintuitionfeeamount',
    width: 140,
  },
  {
    title: '总学费',
    dataIndex: 'totaltuitionfee',
    key: 'totaltuitionfee',
    width: 140,
  },
])

const rowSelection = {
  selectedRowKeys: selectedRowKeys,
  onChange: (keys, rows) => {
    selectedRowKeys.value = keys
    selectedRows.value = rows
  },
}

const defaultOrNotFenClass = ref([0])
const defaultCurrentStatus = ref([1,2])

// 存储所有查询条件
const queryState = ref({
  studentId: undefined,
  fromExpireTime: undefined,
  toExpireTime: undefined,
  fromSuspendedTime: undefined,
  toSuspendedTime: undefined,
  fromClosedTime: undefined,
  toClosedTime: undefined,
  isSetExpireTime: undefined,
  assignedClass: undefined,
  lessonType: undefined,
  remainLessonChargingMode: undefined,
  fromRemainQuantity: undefined,
  toRemainQuantity: undefined,
  lessonChargingList: undefined,
  statusList: [1, 2], // 默认状态：正常 + 停课
  classTeacherId: undefined,
  salespersonId: undefined,
  classIds: undefined,
  productIds: undefined,
  isArrears: undefined,
  lastestTeachingRecordStartTime: undefined,
  lastestTeachingRecordEndTime: undefined,
  // 时间范围原始字段（仅用于中间转换，不直接传给接口）
  classEndingTime: undefined,
  classStopTime: undefined,
  expiryDate: undefined,
  lastClassTime: undefined,
})

// 重置所有查询条件（保留默认状态）
function resetQueryState() {
  Object.keys(queryState.value).forEach((key) => {
    if (key === 'statusList') {
      queryState.value.statusList = [1, 2]
    } else {
      queryState.value[key] = undefined
    }
  })
}

// 分页参数
const pagination = ref({
  current: 1,
  pageSize: 10,
  total: 0,
  showSizeChanger: true,
  showTotal: total => `共 ${total} 条`,
  pageSizeOptions: ['10', '20', '50', '100'],
  hideOnSinglePage: false,
  showQuickJumper: true,
})

// 使用防抖处理筛选条件更新
const handleFilterUpdate = debounce((updates, isClearAll = false, id, type) => {
  if (isClearAll) {
    // 清除所有条件，恢复默认状态
    resetQueryState()
  } else {
    Object.entries(updates).forEach(([key, value]) => {
      queryState.value[key] = value
    })
  }

  pagination.value.current = 1
  selectedRows.value = []
  selectedRowKeys.value = []
  // 将最新的查询条件传入，方便在接口里做时间/区间字段转换
  getRegisterReadList(queryState.value, id, type)
}, 300, { leading: true, trailing: false })

// 过滤器字段映射
const filterFieldMapping = {
  stuPhoneSearchFilter: 'studentId',
  currentStatusFilter: 'statusList',
  orNotFenClassFilter: 'assignedClass',
  billingModeFilter: 'lessonChargingList',
  chargingMethodFilter: 'lessonChargingList', // 收费方式（all-filter组件实际emit的事件名）
  isSetExpirationDateFilter: 'isSetExpireTime',
  classEndingTimeFilter: 'classEndingTime',
  classStopTimeFilter: 'classStopTime',
  expiryDateFilter: 'expiryDate',
  // 有效期至（新组件，对应同一组过期时间字段）
  validityPeriodFilter: 'expiryDate',
  createUserFilter: 'createId',
  salesPersonFilter: 'salespersonId',
  intentionCourseFilter: 'productIds',
  teachingMethodFilter: 'lessonType',
  classTeacherFilter: 'classTeacherId',
  classNameFilter: 'classIds',
  enrolledCourseFilter: 'productIds',
  isArrearsFilter: 'isArrears',
  lastClassTimeFilter: 'lastClassTime',
  remainingFilter: 'remaining',
}

// 生成所有过滤器的更新处理器（对标学员管理页）
const filterUpdateHandlers = computed(() => {
  const handlers = {}
  Object.entries(filterFieldMapping).forEach(([eventKey, fieldName]) => {
    handlers[`update:${eventKey}`] = (val, isClearAll, id, type) => {
      // 分班状态：0-未分班，1-已分班
      if (fieldName === 'assignedClass') {
        if (Array.isArray(val) && val.length > 0) {
          if (val.includes(0) && val.includes(1)) {
            handleFilterUpdate({ assignedClass: undefined }, isClearAll, id, type)
          } else if (val.includes(0)) {
            handleFilterUpdate({ assignedClass: false }, isClearAll, id, type)
          } else if (val.includes(1)) {
            handleFilterUpdate({ assignedClass: true }, isClearAll, id, type)
          }
        } else {
          handleFilterUpdate({ assignedClass: undefined }, isClearAll, id, type)
        }
        return
      }

      // 是否设置有效期：1-已设置，2-未设置
      if (fieldName === 'isSetExpireTime') {
        if (val === 2) {
          handleFilterUpdate({ isSetExpireTime: false }, isClearAll, id, type)
        } else if (val === 1) {
          handleFilterUpdate({ isSetExpireTime: true }, isClearAll, id, type)
        } else {
          handleFilterUpdate({ isSetExpireTime: undefined }, isClearAll, id, type)
        }
        return
      }

      // 是否欠费：0-不欠费，1-欠费
      if (fieldName === 'isArrears') {
        if (val === 0) {
          handleFilterUpdate({ isArrears: false }, isClearAll, id, type)
        } else if (val === 1) {
          handleFilterUpdate({ isArrears: true }, isClearAll, id, type)
        } else {
          handleFilterUpdate({ isArrears: undefined }, isClearAll, id, type)
        }
        return
      }

      // 剩余数量：{ mode, min, max } -> remainLessonChargingMode / fromRemainQuantity / toRemainQuantity
      if (fieldName === 'remaining') {
        if (val && typeof val === 'object') {
          const { mode, min, max } = val
          handleFilterUpdate({
            remainLessonChargingMode: mode,
            fromRemainQuantity: min,
            toRemainQuantity: max,
          }, isClearAll, id, type)
        } else {
          handleFilterUpdate({
            remainLessonChargingMode: undefined,
            fromRemainQuantity: undefined,
            toRemainQuantity: undefined,
          }, isClearAll, id, type)
        }
        return
      }

      // 其它普通字段：空数组视为清空筛选
      let finalVal = val
      if (Array.isArray(val) && val.length === 0) {
        finalVal = undefined
      }
      handleFilterUpdate({ [fieldName]: finalVal }, isClearAll, id, type)
    }
  })
  return handlers
})

// 获取报读列表（对标学员管理页调用方式）
async function getRegisterReadList(newQueryParams = {}, id, type) {
  loading.value = true
  try {
    // 定义时间范围字段映射
    const dateRangeMappings = {
      classEndingTime: {
        begin: 'fromClosedTime',
        end: 'toClosedTime',
      },
      classStopTime: {
        begin: 'fromSuspendedTime',
        end: 'toSuspendedTime',
      },
      expiryDate: {
        begin: 'fromExpireTime',
        end: 'toExpireTime',
      },
      lastClassTime: {
        begin: 'lastestTeachingRecordStartTime',
        end: 'lastestTeachingRecordEndTime',
      },
    }

    // 先清除 queryState 中的所有时间范围字段
    Object.values(dateRangeMappings).forEach(({ begin, end }) => {
      queryState.value[begin] = undefined
      queryState.value[end] = undefined
    })

    // 如果有新的查询参数，则处理时间范围
    if (Object.keys(newQueryParams).length > 0) {
      newQueryParams = handleDateRangeParams(newQueryParams, dateRangeMappings)
    }

    // 合并新的查询参数到 queryState
    Object.assign(queryState.value, newQueryParams)

    // 不直接传递原始时间范围字段，只传 begin/end 等实际查询字段
    const originalRangeFields = ['classEndingTime', 'classStopTime', 'expiryDate', 'lastClassTime']

    // 过滤掉 undefined 的值和原始时间范围字段
    const validQueryParams = Object.fromEntries(
      Object.entries(queryState.value)
        .filter(([key, value]) => value !== undefined && !originalRangeFields.includes(key)),
    )

    const res = await getRegisterReadListApi({
      pageRequestModel: {
        needTotal: true,
        pageSize: pagination.value.pageSize,
        pageIndex: pagination.value.current,
        skipCount: 0,
      },
      queryModel: validQueryParams,
    })

    if (res.code === 200 && res.result) {
      dataSource.value = res.result.studentTutionAccounts || []
      pagination.value.total = res.result.total || 0
      totalStats.value = {
        totalRemainedTuition: res.result.totalRemainedTuition || 0,
        totalConfirmedTuition: res.result.totalConfirmedTuition || 0,
        totalPaidRemainedTuition: res.result.totalPaidRemainedTuition || 0,
        total: res.result.total || 0,
        studentCount: res.result.studentCount || 0,
      }
      // 清空快捷筛选
      allFilterRef.value?.clearQuickFilter(id, type)
    } else {
      messageService.error(res.message || '获取数据失败')
    }
  } catch (error) {
    console.error('获取报读列表失败:', error)
    messageService.error('获取数据失败')
  } finally {
    loading.value = false
  }
}

// 处理表格变化
function handleTableChange(paginationInfo) {
  pagination.value.current = paginationInfo.current
  pagination.value.pageSize = paginationInfo.pageSize
  getRegisterReadList()
}

// 格式化金额
function formatMoney(amount) {
  if (amount === null || amount === undefined) return '¥ 0'
  return `¥ ${Number(amount).toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`
}

// 格式化日期
function formatDate(dateStr) {
  if (!dateStr || dateStr === '0001-01-01T00:00:00' || dateStr === '') return '-'
  return dayjs(dateStr).format('YYYY-MM-DD')
}

// 获取性别文本
function getSexText(sex) {
  // 统一使用全局枚举：0=女，1=男，2=未知
  return SexLabel[sex] || SexLabel[Sex.Unknown]
}

// 获取状态文本和样式
function getStatusInfo(status) {
  const statusMap = {
    1: { text: '正常', class: 'text-#0c3 bg-#e6ffec' },
    2: { text: '停课', class: 'text-#f90 bg-#fff5e6' },
    3: { text: '结课', class: 'text-#888 bg-#f5f5f5' },
  }
  return statusMap[status] || { text: '未知', class: 'text-#888 bg-#f5f5f5' }
}

// 获取计费模式文本
function getChargingModeText(mode) {
  const modeMap = {
    1: '按课时',
    2: '按时段',
    3: '按金额',
  }
  return modeMap[mode] || '未知'
}

// 获取授课方式文本
function getLessonTypeText(type) {
  const typeMap = {
    1: '班级授课',
    2: '1v1授课',
  }
  return typeMap[type] || '未知'
}

// 获取办理类型文本
function getHandleTypeText(type) {
  const typeMap = {
    0: '试听',
    1: '报读',
    2: '续费',
    3: '转课',
  }
  return typeMap[type] || '未知'
}

// 获取数量文本（根据计费模式）
function getQuantityText(quantity, chargingMode) {
  if (chargingMode === 1) {
    return `${quantity}课时`
  } else if (chargingMode === 2) {
    return `${quantity}天`
  } else if (chargingMode === 3) {
    return `${quantity}元`
  } else {
    return `${quantity}`
  }
}

const allFilterRef = ref(null)

const { selectedValues, columnOptions, filteredColumns, totalWidth }
  = useTableColumns({
    storageKey: 'register-read-list',
    allColumns,
    excludeKeys: ['action'],
  })

onMounted(() => {
  getRegisterReadList()
})

// 统一的学员列表刷新事件监听
useStudentListRefresh(getRegisterReadList)
</script>

<template>
  <div>
    <!-- 学员筛选条件 -->
    <div class="filter-wrap bg-white pl-3 pr-3 rounded-4">
      <all-filter
        ref="allFilterRef"
        :default-or-not-fen-class="defaultOrNotFenClass"
        :default-current-status="defaultCurrentStatus"
        :display-array="displayArray"
        :is-quick-show="false"
        :is-show-search-stu-phonefilter="true"
        v-on="filterUpdateHandlers"
      />
    </div>
    <div class="student-list mt-2 pt-3 pb-3 pl-6 pr-6 bg-white rounded-4">
      <div class="tab-table">
        <div class="table-title flex justify-between">
          <div class="total">
            总计（{{ pagination.total }}条）：已课消金额：{{ formatMoney(totalStats.totalConfirmedTuition) }}，剩余学费金额：{{ formatMoney(totalStats.totalRemainedTuition) }}，学员人数：{{ totalStats.studentCount }}人
          </div>
          <div class="edit flex">
            <a-dropdown class="mr-2">
              <template #overlay>
                <a-menu>
                  <a-menu-item key="1">
                    批量停课
                  </a-menu-item>
                  <a-menu-item key="2">
                    批量复课
                  </a-menu-item>
                  <a-menu-item key="3">
                    批量结课
                  </a-menu-item>
                  <a-menu-item key="4">
                    批量转课
                  </a-menu-item>
                  <a-menu-item key="5">
                    批量修改有效期
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
            <!-- 自定义字段 -->
            <customize-code
              :checked-values="selectedValues"
              :options="columnOptions"
              :total="allColumns.length"
              :num="selectedValues.length"
              @update:checked-values="(val) => (selectedValues = val)"
            />
          </div>
        </div>
        <div class="table-content mt-2">
          <a-table
            :data-source="dataSource"
            :loading="loading"
            :pagination="pagination"
            :columns="filteredColumns"
            :row-selection="rowSelection"
            :scroll="{ x: totalWidth }"
            row-key="tuitionAccountId"
            size="small"
            @change="handleTableChange"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'name'">
                <student-avatar
                  :id="record.studentId"
                  :name="record.studentName || '-'"
                  :gender="getSexText(record.sex)"
                  :avatar-url="record.avatar || 'https://pcsys.admin.ybc365.com/c04d0ea2-a8b0-4001-b19b-946a980cb726.png'"
                  :show-age="false"
                  default-active-key="0"
                />
              </template>
              <template v-if="column.key === 'phone'">
                <div class="name">
                  <div class="text-3.5 text-#333">
                    {{ record.phone || '-' }}
                  </div>
                </div>
              </template>
              <template v-if="column.key === 'registerReadCourse'">
                <div class="text-#222">
                  {{ record.lessonName || '-' }}
                </div>
                <div class="text-3 text-#888 flex flex-items-center">
                  {{ getLessonTypeText(record.lessonType) }} | {{ getChargingModeText(record.lessonChargingMode) }}
                </div>
              </template>
              <template v-if="column.key === 'registerReadNum'">
                <div class="text-#222">
                  {{ getQuantityText((record.totalQuantity || 0) + (record.totalFreeQuantity || 0), record.lessonChargingMode) }}
                </div>
                <div class="text-3 text-#888 flex flex-items-center">
                  购{{ record.totalQuantity || 0 }}{{ record.lessonChargingMode === 1 ? '课时' : record.lessonChargingMode === 2 ? '天' : record.lessonChargingMode === 3 ? '元' : '' }}<span v-if="(record.totalFreeQuantity || 0) > 0">+赠{{ record.totalFreeQuantity || 0 }}{{ record.lessonChargingMode === 1 ? '课时' : record.lessonChargingMode === 2 ? '天' : record.lessonChargingMode === 3 ? '元' : '' }}</span>
                </div>
              </template>
              <template v-if="column.key === 'currentStatus'">
                <div :class="`${getStatusInfo(record.tuitionAccountStatus).class} rounded-2.5 inline-block text-3 pt-0.5 pb-0.5 pl-2 pr-2`">
                  {{ getStatusInfo(record.tuitionAccountStatus).text }}
                </div>
              </template>
              <template v-if="column.key === 'orNotFenClass'">
                <div :class="`${record.assignedClass ? 'text-#0c3 bg-#e6ffec' : 'text-#f90 bg-#fff5e6'} rounded-2.5 inline-block text-3 pt-0.5 pb-0.5 pl-2 pr-2`">
                  {{ record.assignedClass ? '已分班' : '未分班' }}
                </div>
              </template>
              <template v-if="column.key === 'headTeacher'">
                {{ record.classTeacherList && record.classTeacherList.length > 0 ? record.classTeacherList.map(t => t.name).join('、') : '-' }}
              </template>
              <template v-if="column.key === 'expiryDate'">
                {{ formatDate(record.expireTime) }}
              </template>
              <template v-if="column.key === 'usedNumber'">
                {{ getQuantityText(((record.totalQuantity || 0) + (record.totalFreeQuantity || 0)) - ((record.quantity || 0) + (record.freeQuantity || 0)), record.lessonChargingMode) }}
              </template>
              <template v-if="column.key === 'remainNumber'">
                {{ getQuantityText((record.quantity || 0) + (record.freeQuantity || 0), record.lessonChargingMode) }}
              </template>
              <template v-if="column.key === 'usedtuitionfeeamount'">
                {{ formatMoney(record.confirmedTuition) }}
              </template>
              <template v-if="column.key === 'remaintuitionfeeamount'">
                {{ formatMoney(record.paidRemaining) }}
              </template>
              <template v-if="column.key === 'totaltuitionfee'">
                {{ formatMoney(record.totalTuition) }}
              </template>
            </template>
          </a-table>
        </div>
      </div>
    </div>
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
    position: relative;
    vertical-align: middle;
    width: 6px;
    margin-right: 4px;
    background: var(--pro-ant-color-primary);
  }
}
</style>
