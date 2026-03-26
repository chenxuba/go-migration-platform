<script setup>
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { debounce } from 'lodash-es'
import { CaretDownOutlined, CaretUpOutlined, DownOutlined, ExclamationCircleOutlined } from '@ant-design/icons-vue'
import { Modal } from 'ant-design-vue'
import dayjs from 'dayjs'
import { addIntendedStudentApi, batchDeleteIntendedStudentApi, batchAssignSalespersonApi, batchTransferToPublicPoolApi, createStudentFollowUpApi, getIntentStudentListApi, updateIntendedStudentApi, updateStatusApi } from '~@/api/enroll-center/intention-student'
import { useStudentFields } from '~@/composables/useStudentFields'
import messageService from '~@/utils/messageService'
import { useTableColumns } from '@/composables/useTableColumns'
import { calculateAge } from '@/utils/date'
import emitter, { EVENTS } from '~@/utils/eventBus'
import { FollowUpStatus, FollowUpStatusLabel, FollowUpStatusStyle, IntentionLevel, IntentionLevelLabel, IntentionLevelStyle, ParentRelationshipLabel } from '@/enums'
import DeleteConfirmModal from '@/components/common/DeleteConfirmModal.vue'
import { handleDateRangeParams } from '~@/utils/dateRangeParams'
import { useUserStore } from '~@/stores/user'

const userStore = useUserStore()

const displayArray = ref(['customSearch', 'stuPhoneSearch', 'intention', 'intentionCourse',
  'lastFollowTime', 'nextFollowTime', 'createTime', 'age'])

const assignSalesVisible = ref(false)
const modalType = ref(1)
const modalTitle = ref('')
const selectedRows = ref([])
const selectedRowKeys = ref([])
// 跨页选择：存储所有选中的数据
const allSelectedRows = ref(new Map()) // 使用Map存储，key为id，value为完整的行数据
const allSelectedRowKeys = ref(new Set()) // 使用Set存储所有选中的key
const dataSource = ref([])
const allColumns = ref([
  {
    title: '学员/性别/年龄',
    dataIndex: 'name',
    key: 'name',
    fixed: 'left',
    width: 160,
    required: true, // 新增必选标识
  },
  {
    title: '联系电话',
    dataIndex: 'mobile',
    width: 120,
    key: 'mobile',
  },
  {
    title: '意向度',
    key: 'intentionLevel',
    dataIndex: 'addresss',
    width: 100,
  },
  {
    title: '意向课程',
    key: 'intentionCourse',
    dataIndex: 'addresss',
    width: 120,

  },
  {
    title: '渠道',
    dataIndex: 'channelName',
    key: 'channelName',
    width: 100,
  },
  
  {
    title: '最近跟进',
    dataIndex: 'followed',
    key: 'followed',
    width: 170,
    sorter: true,
  },
  {
    title: '下次跟进',
    dataIndex: 'nextTime',
    key: 'nextTime',
    width: 170,
    sorter: true,
  },
  {
    title: '创建时间',
    dataIndex: 'createTime',
    key: 'createTime',
    width: 170,
    sorter: true,
  },
 
  {
    title: '生日',
    dataIndex: 'birthday',
    key: 'birthday',
    width: 120,
  },



])

const systemDefaultIsDisplayCodeList = ref([
  {
    title: '微信号',
    dataIndex: 'wxChat',
    key: 'wxChat',
    width: 120,
    show: false,
    isDynamic: true,
  },
  {
    title: '年级',
    dataIndex: 'grade',
    key: 'grade',
    width: 100,
    show: false,
    isDynamic: true,
  },
  {
    title: '就读学校',
    dataIndex: 'school',
    key: 'school',
    width: 120,
    show: false,
    isDynamic: true,
  },
  {
    title: '家庭住址',
    key: 'address',
    dataIndex: 'address',
    width: 150,
    show: false,
    isDynamic: true,
  },
  {
    title: '兴趣爱好',
    key: 'hobbies',
    dataIndex: 'hobbies',
    width: 150,
    show: false,
    isDynamic: true,
  },
])
const customIsDisplayCodeList = ref([])
const callCustomIsDisplayList = ref([])
const { systemDefaultIsDisplayList, customIsDisplaySearchList, getAllStuFields, getCustomField } = useStudentFields()

// 计算年级选项数据
const gradeOptionsData = computed(() => {
  const gradeField = systemDefaultIsDisplayList.value.find(item => item.fieldKey === '年级')
  if (gradeField && gradeField.optionsJson) {
    return gradeField.optionsJson.split(',').filter(option => option.trim())
  }
  return []
})

// Add watch effect for systemDefaultIsDisplayList
// 控制显示自定义字段和列的逻辑
watch(systemDefaultIsDisplayList, (newList) => {
  // Update show field based on systemDefaultIsDisplayList
  systemDefaultIsDisplayCodeList.value.forEach((item) => {
    const matchingField = newList.find(field => field.fieldKey === item.title)

    if (matchingField) {
      item.show = matchingField.isDisplay
    }
  })

  const fieldsToCheck = [
    { key: 'channelCategory', displayKey: '渠道' },
    { key: 'sex', displayKey: '性别' },
    { key: 'birthday', displayKey: '生日' },
    { key: 'wxChat', displayKey: '微信号' },
    { key: 'grade', displayKey: '年级' },
    { key: 'school', displayKey: '就读学校' },
    { key: 'hobbies', displayKey: '兴趣爱好' },
    { key: 'address', displayKey: '家庭住址' },
  ]

  fieldsToCheck.forEach((field) => {
    const foundField = newList.find(item => item.fieldKey === field.displayKey)

    if (foundField && foundField.isDisplay && foundField.searched) {
      if (!displayArray.value.includes(field.key)) {
        displayArray.value.push(field.key)
      }
    }
  })

  // updateDynamicColumns();
}, { deep: true })

watch(callCustomIsDisplayList, (newList) => {
  customIsDisplayCodeList.value = newList.map(item => ({
    title: item.fieldKey,
    dataIndex: item.fieldKey + item.id,
    key: item.fieldKey + item.id,
    width: 120,
    show: false,
    isDynamic: true,
  }))

  updateDynamicColumns()
}, { deep: true })

// 新增一个函数来统一处理动态列的更新
function updateDynamicColumns() {

  // 获取所有非动态、非操作、非倒计时的基础列
  const baseColumns = allColumns.value.filter(col =>
    !col.isDynamic
    && col.key !== 'action'
  )

  // 合并所有需要显示的动态列
  const visibleSystemColumns = systemDefaultIsDisplayCodeList.value.filter(item => item.show)
  const allDynamicColumns = [...visibleSystemColumns, ...customIsDisplayCodeList.value]

  // 重新组装列顺序：基础列 -> 动态列 -> 操作列
  allColumns.value = [
    ...baseColumns,
    ...allDynamicColumns,
  ]
}

const { selectedValues, columnOptions, filteredColumns, totalWidth }
  = useTableColumns({
    storageKey: 'public-pool-student', // 本地存储键名
    allColumns, // 原始列配置
    excludeKeys: ['action'], // 需要排除的列键
    defaultSelectedKeys: ['mobile', 'intentionLevel', 'intentionCourse','channelName', 'teacher', 'status', 'followed', 'nextTime', 'createTime',  'birthday'],
  })

// 批量分配销售
function handleBatchAssign() {
  const allSelectedRowsArray = Array.from(allSelectedRows.value.values())
  
  if (allSelectedRowsArray.length === 0) {
    messageService.error('请选择要分配销售的学员')
    return
  }
  
  // 使用全局选中数据
  selectedRows.value = allSelectedRowsArray
  selectedRowKeys.value = Array.from(allSelectedRowKeys.value)
  assignSalesVisible.value = true
  modalType.value = 1
  modalTitle.value = '批量分配销售'
}

// 批量认领
function handleBatchClaim() {
  const allSelectedRowsArray = Array.from(allSelectedRows.value.values())
  
  if (allSelectedRowsArray.length === 0) {
    messageService.error('请选择要认领的学员')
    return
  }
  
  // 显示确认弹窗
  Modal.confirm({
    title: `确认批量认领${allSelectedRowsArray.length}名学员？`,
    content: '认领后，你将成为这些学员的所属销售员',
    okText: '确认',
    cancelText: '我再想想',
    centered: true,
    async onOk() {
      try {
        // 认领逻辑：将选中的学员分配给当前用户
        const claimData = {
          studentIds: allSelectedRowsArray.map(row => row.id),
          salespersonId: userStore.userInfo.instUserId
        }
        
        const res = await batchAssignSalespersonApi(claimData)
        if (res.code === 200) {
          messageService.success(`成功认领${allSelectedRowsArray.length}名学员`)
          // 刷新列表
          getIntentStudentList()
          // 清空选择（包括跨页选择）
          clearAllSelection()
        }
      } catch (error) {
        messageService.error('认领失败')
      }
    }
  })
}

const rowSelection = {
  selectedRowKeys: selectedRowKeys,
  onChange: (keys, rows) => {
    // 更新当前页的选中状态
    selectedRowKeys.value = keys
    selectedRows.value = rows

    // 跨页选择逻辑：更新全局选中状态
    // 1. 先从全局状态中移除当前页面的所有数据
    dataSource.value.forEach(item => {
      allSelectedRows.value.delete(item.id)
      allSelectedRowKeys.value.delete(item.id)
    })

    // 2. 将当前页面选中的数据添加到全局状态
    rows.forEach(row => {
      allSelectedRows.value.set(row.id, row)
      allSelectedRowKeys.value.add(row.id)
    })
  },
  onSelect: (record, selected, selectedRows, nativeEvent) => {
    // 单行选择时的处理
    if (selected) {
      allSelectedRows.value.set(record.id, record)
      allSelectedRowKeys.value.add(record.id)
    } else {
      allSelectedRows.value.delete(record.id)
      allSelectedRowKeys.value.delete(record.id)
    }
  },
  onSelectAll: (selected, selectedRows, changeRows) => {
    // 全选/取消全选时的处理
    if (selected) {
      // 全选当前页
      changeRows.forEach(row => {
        allSelectedRows.value.set(row.id, row)
        allSelectedRowKeys.value.add(row.id)
      })
    } else {
      // 取消全选当前页
      changeRows.forEach(row => {
        allSelectedRows.value.delete(row.id)
        allSelectedRowKeys.value.delete(row.id)
      })
    }
  }
}
const loading = ref(false)
const assignSalesRef = ref(null)

// 定义字段映射关系
const fieldMappings = {
  age: ['age', 'ageMin', 'ageMax'],
  birthday: ['birthday', 'birthDayBegin', 'birthDayEnd'],
  lastFollowTime: ['lastFollowTime', 'followUpTimeBegin', 'followUpTimeEnd'],
  nextFollowTime: ['nextFollowTime', 'nextFollowUpTimeBegin', 'nextFollowUpTimeEnd'],
}

// 存储所有查询条件的响应式对象
const queryState = ref({
  queryAllOrDepartment: 1, // 全部（含所有部门）
  isHasSalePerson: false,
  intentionLevels: undefined,
  followUpStatuses: undefined,
  birthDayBegin: undefined,
  birthDayEnd: undefined,
  followUpTimeBegin: undefined,
  followUpTimeEnd: undefined,
  nextFollowUpTimeBegin: undefined,
  nextFollowUpTimeEnd: undefined,
  sexes: undefined,
  ageMin: undefined,
  ageMax: undefined,
  channelIds: undefined,
  // 原始字段
  age: undefined,
  birthday: undefined,
  lastFollowTime: undefined,
  nextFollowTime: undefined,
  // 可以添加更多查询条件...
})

// 重置所有查询条件
function resetQueryState() {
  Object.keys(queryState.value).forEach((key) => {
    queryState.value[key] = undefined
  })
}

// 使用防抖处理所有筛选条件的更新
const handleFilterUpdate = debounce((updates, isClearAll = false, id, type) => {
  if (isClearAll) {
    // 如果是清空所有，重置所有查询条件
    resetQueryState()
  }
  else {
    // 处理更新
    Object.entries(updates).forEach(([key, value]) => {
      if (Array.isArray(value) && value.length === 0 && fieldMappings[key]) {
        // 如果是空数组且在映射中存在，清除所有相关字段
        fieldMappings[key].forEach((field) => {
          queryState.value[field] = undefined
        })
      }
      else if (key === 'customFieldSearchList') {
        // 特殊处理自定义字段搜索列表
        handleCustomFieldSearchUpdate(value, id, type)
      }
      else {
        queryState.value[key] = value
      }
    })
  }

  pagination.value.current = 1
  // 筛选条件改变时清空选择状态
  clearAllSelection()
  getIntentStudentList(queryState.value, id, type)
}, 300, { leading: true, trailing: false })

// 处理自定义字段搜索更新
function handleCustomFieldSearchUpdate(data, id, type) {
  if (!data || (!data.item && !data.value)) {
    // 如果是清空操作，根据id和type来决定清空方式
    if (id && type === 'clear') {
      // 清空特定字段
      const currentList = queryState.value.customFieldSearchList || []
      queryState.value.customFieldSearchList = currentList.filter(item =>
        item.studentCustomFieldId !== id.toString()
      )
      if (queryState.value.customFieldSearchList.length === 0) {
        queryState.value.customFieldSearchList = undefined
      }
    } else {
      // 清空所有自定义字段搜索
      queryState.value.customFieldSearchList = undefined
    }
    return
  }

  const { item, value } = data

  if (!item || !value) {
    return
  }

  // 确保 customFieldSearchList 是数组
  if (!queryState.value.customFieldSearchList) {
    queryState.value.customFieldSearchList = []
  }

  // 查找是否已存在该字段的搜索条件
  const existingIndex = queryState.value.customFieldSearchList.findIndex(
    searchItem => searchItem.studentCustomFieldId === item.id.toString()
  )

  // 构造搜索对象
  const searchObject = {
    studentCustomFieldId: item.id.toString(),
    type: item.fieldType,
    searchOptions: item.fieldType === 4 ? [value] : null, // 选择类型使用searchOptions
    searchKey: item.fieldType === 1 || item.fieldType === 2 ? value : null, // 文本/数字类型使用searchKey
    searchTimeBegin: null,
    searchTimeEnd: null
  }

  if (existingIndex > -1) {
    // 更新已存在的搜索条件
    queryState.value.customFieldSearchList[existingIndex] = searchObject
  } else {
    // 添加新的搜索条件
    queryState.value.customFieldSearchList.push(searchObject)
  }
}

// 分页参数
const pagination = ref({
  current: 1,
  pageSize: 10,
  total: 0,
  showSizeChanger: true,
  showTotal: total => `共 ${total} 条`,
  pageSizeOptions: ['5', '10', '20', '50'],
  hideOnSinglePage: true,
  showQuickJumper: true,
})

// 排序状态管理
const sortModel = ref({
  byFollowUpTime: 0,
  byNextFlowTime: 0,
})
// 清空所有选择状态
function clearAllSelection() {
  selectedRows.value = []
  selectedRowKeys.value = []
  allSelectedRows.value.clear()
  allSelectedRowKeys.value.clear()
}

// 根据全局选中状态更新当前页的选中状态
function updateCurrentPageSelection() {
  const currentPageKeys = []
  const currentPageRows = []

  dataSource.value.forEach(item => {
    if (allSelectedRowKeys.value.has(item.id)) {
      currentPageKeys.push(item.id)
      currentPageRows.push(item)
    }
  })

  selectedRowKeys.value = currentPageKeys
  selectedRows.value = currentPageRows
}

// 获取意向学员列表
async function getIntentStudentList(newQueryParams = {}, id, type) {
  // 定义时间范围字段映射
  const dateRangeMappings = {
    birthday: {
      begin: 'birthDayBegin',
      end: 'birthDayEnd',
    },
    lastFollowTime: {
      begin: 'followUpTimeBegin',
      end: 'followUpTimeEnd',
    },
    nextFollowTime: {
      begin: 'nextFollowUpTimeBegin',
      end: 'nextFollowUpTimeEnd',
    },
    age: {
      begin: 'ageMin',
      end: 'ageMax',
    },
  }

  loading.value = true
  try {
    // 先清除 queryState 中的所有时间范围字段
    Object.values(dateRangeMappings).forEach(({ begin, end }) => {
      queryState.value[begin] = undefined
      queryState.value[end] = undefined
    })

    // 如果有新的查询参数，则处理时间范围
    if (Object.keys(newQueryParams).length > 0) {
      newQueryParams = handleDateRangeParams(newQueryParams, dateRangeMappings)
    }

    // 合并新的查询参数到queryState
    Object.assign(queryState.value, newQueryParams)

    // 过滤掉undefined的值和原始字段，只传递有效的查询条件
    const originalFields = ['age', 'createTime', 'salesAssignedTime', 'birthday', 'lastFollowTime', 'nextFollowTime']
    const validQueryParams = Object.fromEntries(
      Object.entries(queryState.value)
        .filter(([key, value]) => value !== undefined && !originalFields.includes(key)),
    )

    const res = await getIntentStudentListApi({
      'pageRequestModel': {
        'pageSize': pagination.value.pageSize,
        'pageIndex': pagination.value.current,
      },
      'sortModel': sortModel.value,
      'queryModel': {
        ...validQueryParams, // 展开所有有效的查询条件
      },
    })
    dataSource.value = res.result || []
    pagination.value.total = res.total
    allFilterRef.value.clearQuickFilter(id, type)

    // 更新当前页的选中状态（跨页选择）
    updateCurrentPageSelection()

  }
  catch (error) {
    // 获取意向学员列表失败时的错误处理
  }
  finally {
    loading.value = false
  }
}

// 处理表格排序变化
function handleTableChange(paginationInfo, filters, sorter) {
  // 重置所有排序字段为0
  Object.keys(sortModel.value).forEach((key) => {
    sortModel.value[key] = 0
  })

  // 处理排序逻辑（支持单列排序和多列排序）
  const sortFieldMap = {
    'followed': 'byFollowUpTime',
    'nextTime': 'byNextFlowTime',
  }

  // 如果是数组，说明是多列排序，我们只取第一个（因为后端只支持单列排序）
  const currentSorter = Array.isArray(sorter) ? sorter[0] : sorter

  if (currentSorter && currentSorter.order) {
    const sortField = sortFieldMap[currentSorter.field]
    if (sortField) {
      // ascend: 升序(1), descend: 降序(2)
      sortModel.value[sortField] = currentSorter.order === 'ascend' ? 1 : 2
    }
  }

  // 更新分页信息
  pagination.value.current = paginationInfo.current
  pagination.value.pageSize = paginationInfo.pageSize

  // 重新获取数据
  getIntentStudentList()
}

onUnmounted(() => {
  // 组件卸载时移除事件监听
  emitter.off(EVENTS.REFRESH_STUDENT_LIST, getIntentStudentList)
})


onMounted(() => {
  getIntentStudentList()
  getAllStuFields({ filter: 3 })
  getCustomField().then((res) => {
    callCustomIsDisplayList.value = res
  })
  // 监听刷新列表事件
  emitter.on(EVENTS.REFRESH_STUDENT_LIST, getIntentStudentList)
})
const allFilterRef = ref(null)

// 过滤器字段映射
const filterFieldMapping = {
  intentionLevelFilter: 'intentionLevels',
  intentionCourseFilter: 'courseId',
  birthdayFilter: 'birthday',
  lastFollowTimeFilter: 'lastFollowTime',
  nextFollowTimeFilter: 'nextFollowTime',
  followStatusFilter: 'followUpStatuses',
  sexFilter: 'sexes',
  ageFilter: 'age',
  channelFilter: 'channelIds',
  wxChatFilter: 'wechatNumber',
  schoolFilter: 'schoolSearchKey',
  addressFilter: 'addressSearchKey',
  hobbiesFilter: 'interestSearchKey',
  gradeFilter: 'grades',
  stuPhoneSearchFilter: 'studentId',
  customSearchInputFilter: 'customFieldSearchList'
}

// 生成所有过滤器的更新处理器
const filterUpdateHandlers = computed(() => {
  const handlers = {}
  Object.entries(filterFieldMapping).forEach(([eventKey, fieldName]) => {
    handlers[`update:${eventKey}`] = (val, isClearAll, id, type) =>
      handleFilterUpdate({ [fieldName]: val }, isClearAll, id, type)
  })
  return handlers
})
function formatTime(time) {
  return time ? dayjs(time).format('YYYY-MM-DD HH:mm') : '-'
}
function isOverdue(nextFollowUpTime) {
  if (!nextFollowUpTime)
    return false

  const targetTime = dayjs(nextFollowUpTime)
  const now = dayjs()

  return targetTime.isValid() && targetTime.isBefore(now)
}
// 处理分配销售和批量转入公有池
async function handleAssignSales(data) {
  try {
    let res
    let successMessage = ''
    let errorMessage = ''
    
    // 分配销售
    res = await batchAssignSalespersonApi(data)
      successMessage = '分配销售成功'
      errorMessage = '分配销售失败'
    
    if (res.code === 200) {
      messageService.success(successMessage)
      // 刷新列表
      getIntentStudentList()
      // 清空选择（包括跨页选择）
      clearAllSelection()
      // 关闭弹窗
      assignSalesVisible.value = false
    }
  }
  catch (error) {
    console.log(error)
    messageService.error(modalType.value === 3 ? '转入公有池失败' : '分配销售失败')
  }
  finally {
    // 关闭modal中的loading
    assignSalesRef.value?.closeLoading()
  }
}
// 暴露方法给父组件调用
defineExpose({
  getIntentStudentList,
  clearAllSelection
})
</script>

<template>
  <div class="tab-content">
    <all-filter ref="allFilterRef" :display-array="displayArray" :custom-is-display-list="customIsDisplaySearchList"
      :student-status="0" :grade-options-data="gradeOptionsData" v-on="filterUpdateHandlers" />
    <div class="tab-table">
      <div class="table-title flex justify-between flex-items-center">
        <div class="total whitespace-nowrap mr-12px">
          当前共{{ pagination.total || 0 }}名学员
          <span v-if="allSelectedRowKeys.size > 0" class="ml-2 text-blue-600">
            （已选中{{ allSelectedRowKeys.size }}名学员）
            <a-button type="link" size="small" class="p-0 ml-1" @click="clearAllSelection">
              清空选择
            </a-button>
          </span>
        </div>
        <div class="edit flex overflow-x-auto">
          <a-button type="primary" class="mr-2" @click="handleBatchAssign">
            批量分配
          </a-button>
          <a-button class="mr-2" @click="handleBatchClaim">
            批量认领
          </a-button>
          <!-- 自定义字段 -->
          <customize-code v-model:checked-values="selectedValues" :options="columnOptions"
            :total="allColumns.length - 1" :num="selectedValues.length" />
        </div>
      </div>
      <div class="table-content mt-2">
        <a-table :data-source="dataSource" row-key="id" :loading="loading" :pagination="pagination"
          :columns="filteredColumns" :row-selection="rowSelection" :scroll="{ x: totalWidth }"
          :sticky="{ offsetHeader: 100 }" size="small" @change="handleTableChange">
          <template #headerCell="{ column }">
            <!-- 动态自定义字段表头 -->
            <template v-if="customIsDisplayCodeList.some(item => item.key === column.key)">
              <clamped-text :text="column.title" :lines="1" />
            </template>
            <template v-if="column.key === 'status'">
              <span class="mr-1">{{ column.title }}</span>
              <a-popover color="#fff" title="跟进状态">
                <template #content>
                  跟进状态为手动标记，仅为区分跟进状态，与学员实际系统状态无关
                </template>
                <ExclamationCircleOutlined />
              </a-popover>
            </template>
            <template v-if="column.key === 'recommendStudentName'">
              <span class="mr-1">{{ column.title }}</span>
              <a-popover color="#fff" title="推荐人">
                <template #content>
                  通过转介绍推荐的此意向学员
                </template>
                <ExclamationCircleOutlined />
              </a-popover>
            </template>
          </template>
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'name'">
              <student-avatar :id="record.id" :name="record.stuName"
                :gender="record.stuSex == 1 ? '男' : record.stuSex == 0 ? '女' : '未知'"
                :age="calculateAge(record.birthDay)" :avatar-url="record.avatarUrl" default-active-key="4" />
            </template>
            <!-- 联系电话 -->
            <template v-if="column.key === 'mobile'">
              <div class="flex flex-col gap-1">
                <span v-if="record.phoneRelationship" class="text-14px  leading-none">
                  {{ ParentRelationshipLabel[record.phoneRelationship] }}
                </span>
                <span class="text-14px  leading-none">{{ record.mobile || '-' }}</span>
              </div>
            </template>
            <!-- 意向度 -->
            <template v-if="column.key === 'intentionLevel'">
                <div >
                    <div class="intention">
                      <span class="intentionTag"
                        :style="{ background: IntentionLevelStyle[record.intentLevel]?.color }" />
                      {{ IntentionLevelLabel[record.intentLevel] || IntentionLevelLabel[IntentionLevel.Unknown] }}
                    </div>
                  </div>
            </template>
            <!-- 意向课程 -->
            <template v-if="column.key === 'intentionCourse'">
              <clamped-text :text="record.lessons && Array.isArray(record.lessons) && record.lessons.length > 0
                ? record.lessons.map(course => course.name).join(', ')
                : '-'" />
            </template>
            <!-- 渠道 -->
            <template v-if="column.key === 'channelName'">
              <clamped-text :text="record.channelName || '-'" />
            </template>
          
            <!-- 最近跟进 -->
            <template v-if="column.key == 'followed'">
              <clamped-text :text="formatTime(record.followUpTime) || '-'" />
            </template>
            <!-- 下次跟进 -->
            <template v-if="column.key == 'nextTime'">
              <clamped-text
                :class="{ 'text-#f90': isOverdue(record.nextFollowUpTime) && record.followUpInterviewStatus === 1 }"
                :text="formatTime(record.nextFollowUpTime) || '-'" />
            </template>
            <!-- 创建时间 -->
            <template v-if="column.key == 'createTime'">
              <clamped-text :text="formatTime(record.createTime) || '-'" />
            </template>
         
            <!-- 生日 -->
            <template v-if="column.key === 'birthday'">
              {{ record.birthDay || '-' }}
            </template>
          
            <!-- 微信号 -->
            <template v-if="column.key === 'wxChat'">
              <clamped-text :text="record.weChatNumber || '-'" />
            </template>
            <!-- 年级 -->
            <template v-if="column.key === 'grade'">
              <clamped-text :text="record.grade || '-'" />
            </template>
            <!-- 就读学校 -->
            <template v-if="column.key === 'school'">
              <clamped-text :text="record.studySchool || '-'" />
            </template>
            <!-- 家庭住址 -->
            <template v-if="column.key === 'address'">
              <clamped-text :text="record.address || '-'" />
            </template>
            <!-- 兴趣爱好 -->
            <template v-if="column.key === 'hobbies'">
              <clamped-text :text="record.interest || '-'" />
            </template>
            <!-- 动态自定义字段内容 customInfo 要处理customInfo为null的情况 -->
            <template v-if="customIsDisplayCodeList.some(item => item.key === column.key)">
              <clamped-text
                :text="record.customInfo && record.customInfo.find(item => item.fieldName + item.fieldId === column.key)?.value || '-'" />
            </template>
          
          </template>
        </a-table>
      </div>
    </div>
  </div>
  <!-- 分配销售 -->
  <assign-sales-modal ref="assignSalesRef" v-model:open="assignSalesVisible" :type="modalType" :title="modalTitle"
    :selected-students="selectedRows"  @submit="handleAssignSales"/>
 
</template>

<style lang="less" scoped>
.tab-content {
  margin: 8px 0;

  .tab-table {
    background: #fff;
    margin-top: 8px;
    padding: 12px;
    border-radius: 12px;

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

    .intention {
      display: flex;
      align-items: center;

      .statusTag {
        width: 54px;
        height: 22px;
        border-radius: 100px;
        display: flex;
        align-items: center;
        justify-content: center;
        font-size: 12px;
        font-family: PingFangSC-Regular, PingFang SC, sans-serif;
        font-weight: 600;
      }
    }

    .intentionTag {
      display: inline-block;
      width: 6px;
      height: 6px;
      border-radius: 100px;
      margin-right: 3px;
    }

    .action {
      a {
        color: var(--pro-ant-color-primary);
        ;
      }
    }
  }

}

.hover {
  &:hover {
    .name {
      color: #06f;
    }
  }
}
</style>
