<script setup>
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { debounce } from 'lodash-es'
import { DownOutlined, ExclamationCircleOutlined, InfoCircleOutlined } from '@ant-design/icons-vue'
import dayjs from 'dayjs'
import { getEnrolledStudentListApi } from '~@/api/edu-center/enrolled-student'
import { useStudentFields } from '@/composables/useStudentFields'
import { useTableColumns } from '@/composables/useTableColumns'
import { useRouter } from 'vue-router'
import { calculateAge } from '@/utils/date'
import { ParentRelationshipLabel } from '@/enums'
import messageService from '~@/utils/messageService'
import emitter, { EVENTS } from '~@/utils/eventBus'
import { useStudentListRefresh } from '@/composables/useStudentListRefresh'
import { handleDateRangeParams } from '~@/utils/dateRangeParams'

const props = defineProps({
  currentType: {
    type: Number,
    default: 1, // 1: 在读学员, 2: 历史学员
  },
})

const router = useRouter()
const openDrawer = ref(false)
const activeKey = ref('1')
// 筛选条件显示配置 - 排除意向课程、跟进状态、下次跟进、未跟进天数、是否被推荐、分配时间、体验课购买状态
const displayArray = ref(['customSearch', 'sex', 'createUser', 'createTime', 'age', 'channelCategory','birthday',  'salesPerson',   'wxChat', 'grade', 'school', 'address', 'hobbies', 'studentStatus'])
const dataSource = ref([])
const loading = ref(false)
const selectedRows = ref([])
const selectedRowKeys = ref([])
const exportModalVisible = ref(false)
const exportRecordModalVisible = ref(false)
const exportMode = ref('all')
const exportReportType = ref('student')
const exportFileType = ref('excel')
const exportConditionItems = ref([])

const exportPreviewColumns = [
  { title: '学员姓名', dataIndex: 'stuName', key: 'stuName' },
  { title: '学员年龄', dataIndex: 'age', key: 'age' },
  { title: '学员生日', dataIndex: 'birthDay', key: 'birthDay' },
  { title: '学员性别', dataIndex: 'sex', key: 'sex' },
  { title: '学员电话', dataIndex: 'mobile', key: 'mobile' },
  { title: '电话关系', dataIndex: 'relation', key: 'relation' },
  { title: '备用手机号', dataIndex: 'backupPhone', key: 'backupPhone' },
  { title: '微信', dataIndex: 'wechat', key: 'wechat' },
  { title: '学员备注', dataIndex: 'remark', key: 'remark' },
  { title: '家校通关注状态', dataIndex: 'bindStatus', key: 'bindStatus' },
  { title: '人脸采集状态', dataIndex: 'faceStatus', key: 'faceStatus' },
  { title: '学员状态', dataIndex: 'studentStatusLabel', key: 'studentStatusLabel' },
  { title: '创建人', dataIndex: 'creator', key: 'creator' },
  { title: '创建日期', dataIndex: 'createdAt', key: 'createdAt' },
  { title: '首次报读时间', dataIndex: 'firstEnrollAt', key: 'firstEnrollAt' },
  { title: '渠道', dataIndex: 'channel', key: 'channel' },
  { title: '转介绍推荐人', dataIndex: 'recommender', key: 'recommender' },
  { title: '销售员', dataIndex: 'salesperson', key: 'salesperson' },
  { title: '最新跟进时间', dataIndex: 'latestFollowAt', key: 'latestFollowAt' },
  { title: '关联储值账户余额', dataIndex: 'balance', key: 'balance' },
  { title: '关联储值账户赠送余额', dataIndex: 'giftBalance', key: 'giftBalance' },
  { title: '订单欠费金额', dataIndex: 'arrearAmount', key: 'arrearAmount' },
  { title: '剩余积分数量', dataIndex: 'points', key: 'points' },
  { title: '家庭住宅', dataIndex: 'residence', key: 'residence' },
]
const exportPreviewRows = [
  {
    stuName: '王小明',
    age: '18周岁',
    birthDay: '2010-01-01',
    sex: '男',
    mobile: '18818888888',
    relation: '母亲',
    backupPhone: '13800138000',
    wechat: '18818888888',
    remark: '测试学员',
    bindStatus: '已采集',
    faceStatus: '已关注',
    studentStatusLabel: '在读',
    creator: '李晨',
    createdAt: '2022-10-18 14:46:50',
    firstEnrollAt: '2024-11-21 12:00:00',
    channel: '转介绍',
    recommender: '李晨',
    salesperson: '孙勇',
    latestFollowAt: '2023-02-14 10:36:35',
    balance: '10000',
    giftBalance: '1000',
    arrearAmount: '0',
    points: '980',
    residence: 'xxx',
  },
]
const exportFieldCount = computed(() => exportPreviewColumns.length)
const exportQuerySummary = computed(() => {
  if (exportConditionItems.value.length === 0) {
    return ['全部导出']
  }
  return exportConditionItems.value.map(item => `${item.label}：${item.value}`)
})
const exportRecordConditions = computed(() => {
  if (exportConditionItems.value.length === 0) {
    return [{ label: '查询条件', value: '全部' }]
  }
  return exportConditionItems.value
})

// 计算年级选项数据
const gradeOptionsData = computed(() => {
  const gradeField = systemDefaultIsDisplayList.value.find(item => item.fieldKey === '年级')
  if (gradeField && gradeField.optionsJson) {
    return gradeField.optionsJson.split(',').filter(option => option.trim())
  }
  return []
})

const allColumns = ref([
  {
    title: '学员/性别/年龄',
    dataIndex: 'name',
    key: 'name',
    fixed: 'left',
    width: 160,
    required: true,
  },
  {
    title: '联系电话',
    dataIndex: 'mobile',
    width: 120,
    key: 'mobile',
  },
  {
    title: '家校云',
    key: 'isBindChild',
    dataIndex: 'isBindChild',
    width: 100,
  },
  {
    title: '人脸采集',
    key: 'isCollect',
    dataIndex: 'isCollect',
    width: 100,
  },
  {
    title: '学员状态',
    dataIndex: 'studentStatus',
    key: 'studentStatus',
    width: 110,
  },
  {
    title: '创建人',
    dataIndex: 'createName',
    key: 'createName',
    width: 100,
  },
  {
    title: '创建时间',
    dataIndex: 'createTime',
    key: 'createTime',
    width: 170,
  },
  {
    title: '首次报读时间',
    dataIndex: 'firstEnrolledTime',
    key: 'firstEnrolledTime',
    width: 170,
  },
  {
    title: '渠道',
    dataIndex: 'channelName',
    key: 'channelName',
    width: 100,
  },
  {
    title: '生日',
    key: 'birthDay',
    dataIndex: 'birthDay',
    width: 120,
  },
  {
    title: '销售',
    key: 'salePersonName',
    dataIndex: 'salePersonName',
    width: 100,
  },
  {
    title: '最新跟进',
    dataIndex: 'followUpTime',
    key: 'followUpTime',
    fixed: 'right',
    width: 170,
    required: true,
  },
])

// 系统默认可显示字段列表
const systemDefaultIsDisplayCodeList = ref([
  {
    title: '微信号',
    dataIndex: 'weChatNumber',
    key: 'weChatNumber',
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
    dataIndex: 'studySchool',
    key: 'studySchool',
    width: 140,
    show: false,
    isDynamic: true,
  },
  {
    title: '家庭地址',
    key: 'address',
    dataIndex: 'address',
    width: 140,
    show: false,
    isDynamic: true,
  },
])

const customIsDisplayCodeList = ref([])
const callCustomIsDisplayList = ref([])

const { systemDefaultIsDisplayList, customIsDisplaySearchList, getAllStuFields, getCustomField } = useStudentFields()

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
  // 获取所有非动态的基础列
  const baseColumns = allColumns.value.filter(col => !col.isDynamic)

  // 合并所有需要显示的动态列
  const visibleSystemColumns = systemDefaultIsDisplayCodeList.value.filter(item => item.show)
  const allDynamicColumns = [...visibleSystemColumns, ...customIsDisplayCodeList.value]

  // 重新组装列顺序：基础列 -> 动态列
  allColumns.value = [
    ...baseColumns,
    ...allDynamicColumns,
  ]
}

const rowSelection = {
  selectedRowKeys: selectedRowKeys,
  onChange: (keys, rows) => {
    selectedRowKeys.value = keys
    selectedRows.value = rows
  },
}

// 定义字段映射关系
const fieldMappings = {
  age: ['age', 'ageMin', 'ageMax'],
  createTime: ['createTime', 'createTimeBegin', 'createTimeEnd'],
  birthday: ['birthday', 'birthDayBegin', 'birthDayEnd'],
  lastFollowTime: ['lastFollowTime', 'followUpTimeBegin', 'followUpTimeEnd'],
}

// 存储所有查询条件
const queryState = ref({
  studentId: undefined, // 学员ID
  mobile: undefined, // 手机号
  stuName: undefined,
  sexes: undefined,
  customFieldSearchList: undefined,
  studentStatuses: [props.currentType], // 根据currentType设置默认值
  createTimeBegin: undefined,
  createTimeEnd: undefined,
  birthDayBegin: undefined,
  birthDayEnd: undefined,
  followUpTimeBegin: undefined,
  followUpTimeEnd: undefined,
  ageMin: undefined,
  ageMax: undefined,
  channelIds: undefined,
  wechatNumber: undefined,
  schoolSearchKey: undefined,
  addressSearchKey: undefined,
  interestSearchKey: undefined,
  grades: undefined,
  createId: undefined,
  salespersonId: undefined,
  isHasSalePerson: undefined,
  // 原始字段
  age: undefined,
  createTime: undefined,
  birthday: undefined,
  lastFollowTime: undefined,
})

// 监听 currentType 变化，更新 studentStatuses
watch(() => props.currentType, (newType) => {
  // 切换学员类型时，重置所有筛选条件，但保留当前类型的学员状态
  resetQueryState([newType])
  // 重置分页
  pagination.value.current = 1
  // 清空选中的行
  selectedRows.value = []
  selectedRowKeys.value = []
  // 清空筛选器的显示状态（如果需要）
  if (allFilterRef.value?.clearQuickFilter) {
    allFilterRef.value.clearQuickFilter()
  }
  // 重新获取数据
  getEnrolledStudentList()
}, { immediate: false })

// 重置所有查询条件
function resetQueryState(studentStatuses = [1, 2]) {
  Object.keys(queryState.value).forEach((key) => {
    if (key === 'studentStatuses') {
      queryState.value[key] = studentStatuses // 使用传入的学员状态，默认为所有学员（在读+历史）
    } else {
      queryState.value[key] = undefined
    }
  })
}

// 使用防抖处理筛选条件更新
const handleFilterUpdate = debounce((updates, isClearAll = false, id, type) => {
  if (isClearAll) {
    resetQueryState()
  }
  else {
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
  selectedRows.value = []
  selectedRowKeys.value = []
  getEnrolledStudentList(queryState.value, id, type)
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
  hideOnSinglePage: false,
  showQuickJumper: true,
})

// 获取在读学员列表
async function getEnrolledStudentList(newQueryParams = {}, id, type) {
  // 定义时间范围字段映射
  const dateRangeMappings = {
    createTime: {
      begin: 'createTimeBegin',
      end: 'createTimeEnd',
    },
    birthday: {
      begin: 'birthDayBegin',
      end: 'birthDayEnd',
    },
    lastFollowTime: {
      begin: 'followUpTimeBegin',
      end: 'followUpTimeEnd',
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
    const originalFields = ['age', 'createTime', 'birthday', 'lastFollowTime']
    const validQueryParams = Object.fromEntries(
      Object.entries(queryState.value)
        .filter(([key, value]) => value !== undefined && !originalFields.includes(key)),
    )

    const res = await getEnrolledStudentListApi({
      pageRequestModel: {
        pageSize: pagination.value.pageSize,
        pageIndex: pagination.value.current,
      },
      queryModel: {
        ...validQueryParams,
      },
    })
    
    if (res.code === 200 && res.result) {
      dataSource.value = res.result || []
      pagination.value.total = res.total || 0
    }
    else {
      messageService.error(res.message || '获取数据失败')
    }
    
    // 清空快捷筛选
    allFilterRef.value?.clearQuickFilter(id, type)
  }
  catch (error) {
    console.error('获取在读学员列表失败:', error)
    messageService.error('获取数据失败')
  }
  finally {
    loading.value = false
  }
}

// 处理表格变化
function handleTableChange(paginationInfo) {
  pagination.value.current = paginationInfo.current
  pagination.value.pageSize = paginationInfo.pageSize
  getEnrolledStudentList()
}
function handleSeeStuData() {
  openDrawer.value = true
}

const defaultStudentStatus = computed(() => props.currentType)
const allFilterRef = ref(null)

// 过滤器字段映射
const filterFieldMapping = {
  stuNameFilter: 'stuName',
  sexFilter: 'sexes',
  customSearchInputFilter: 'customFieldSearchList',
  stuStatusFilter: 'studentStatuses',
  stuPhoneSearchFilter: 'studentId', // 学员/电话搜索映射到studentId
  createTimeFilter: 'createTime',
  birthdayFilter: 'birthday',
  lastFollowTimeFilter: 'lastFollowTime',
  ageFilter: 'age',
  channelFilter: 'channelIds',
  wxChatFilter: 'wechatNumber',
  schoolFilter: 'schoolSearchKey',
  addressFilter: 'addressSearchKey',
  hobbiesFilter: 'interestSearchKey',
  gradeFilter: 'grades',
  createUserFilter: 'createId',
  salesPersonFilter: 'salespersonId',
  hasSalesPersonFilter: 'isHasSalePerson',
}

// 生成所有过滤器的更新处理器
const filterUpdateHandlers = computed(() => {
  const handlers = {}
  Object.entries(filterFieldMapping).forEach(([eventKey, fieldName]) => {
    handlers[`update:${eventKey}`] = (val, isClearAll, id, type) => {
      // 特殊处理 studentStatuses
      if (fieldName === 'studentStatuses') {
        // 如果清除学员状态筛选，查询所有学员（在读+历史）
        if (!val || val === undefined) {
          handleFilterUpdate({ [fieldName]: [1, 2] }, isClearAll, id, type)
        } else {
          handleFilterUpdate({ [fieldName]: [val] }, isClearAll, id, type)
        }
      } else {
        handleFilterUpdate({ [fieldName]: val }, isClearAll, id, type)
      }
    }
  })
  return handlers
})

const { selectedValues, columnOptions, filteredColumns, totalWidth }
  = useTableColumns({
    storageKey: 'studyingOrHistoryColumns',
    allColumns,
    excludeKeys: [],
    defaultSelectedKeys: ['mobile', 'isBindChild', 'isCollect', 'studentStatus', 'createName', 'createTime', 'firstEnrolledTime', 'channelName', 'birthDay', 'salePersonName', 'followUpTime'],
  })

function handleImportExportAction({ key }) {
  syncExportConditions()
  switch (key) {
    case '1':
      router.push('/import-center/starter/order')
      break
    case '2':
      router.push('/import-center/starter/intentionStudent')
      break
    case '3':
      exportModalVisible.value = true
      break
    case '4':
      exportRecordModalVisible.value = true
      break
  }
}

function syncExportConditions() {
  const conditions = allFilterRef.value?.getOrderedConditions?.() || []
  exportConditionItems.value = conditions.map((item) => {
    const values = Array.isArray(item.values) ? item.values : []
    const displayValue = values.length > 0
      ? values.map(valueItem => `${valueItem?.value || ''}`.replace(' 至 ', ' ~ ')).filter(Boolean).join('、')
      : '全部'
    return {
      label: item.label,
      value: displayValue || '全部',
    }
  })
}

function handleViewExportRecord() {
  exportModalVisible.value = false
  exportRecordModalVisible.value = true
}

function handleSubmitExport() {
  exportModalVisible.value = false
  messageService.success('批量导出静态页面已展示')
}

function handleDownloadExportRecord() {
  messageService.info('导出记录下载功能待接入')
}

function formatTime(time) {
  return time ? dayjs(time).format('YYYY-MM-DD HH:mm') : '-'
}

onMounted(() => {
  getEnrolledStudentList()
  getAllStuFields({ filter: 3 })
  getCustomField().then((res) => {
    callCustomIsDisplayList.value = res
    customIsDisplayCodeList.value = res.map(item => ({
      title: item.fieldKey,
      dataIndex: item.fieldKey + item.id,
      key: item.fieldKey + item.id,
      width: 120,
      show: false,
      isDynamic: true,
    }))
  })
})

// 统一的学员列表刷新事件监听
useStudentListRefresh(getEnrolledStudentList)

// 暴露方法给父组件调用
defineExpose({
  getEnrolledStudentList
})
</script>

<template>
  <div>
    <!-- 学员筛选条件 -->
    <div class="filter-wrap mt-2 bg-white pl-3 pr-3 rounded-4">
      <all-filter 
        ref="allFilterRef"
        :default-student-status="defaultStudentStatus" 
        :display-array="displayArray" 
        :is-quick-show="false"
        :is-show-search-stu-phonefilter="true"
        :custom-is-display-list="customIsDisplaySearchList"
        :student-status="currentType"
        :grade-options-data="gradeOptionsData"
        v-on="filterUpdateHandlers"
      />
    </div>
    <div class="student-list mt-2 pt-3 pb-3 pl-6 pr-6 bg-white rounded-4">
      <div class="tab-table">
        <div class="table-title flex justify-between">
          <div class="total">
            当前共{{ pagination.total || 0 }}名学员
            <span v-if="selectedRowKeys.length > 0" class="ml-2 text-blue-600">
              （已选中{{ selectedRowKeys.length }}名学员）
              <a-button type="link" size="small" class="p-0 ml-1" @click="selectedRowKeys = []; selectedRows = []">
                清空选择
              </a-button>
            </span>
          </div>
          <div class="edit flex">
            <div class="upNew">
              <a-button class="mr-2">
                群发短信
              </a-button>
            </div>
            <a-dropdown class="mr-2">
              <template #overlay>
                <a-menu>
                  <a-menu-item key="1">
                    批量分配销售
                  </a-menu-item>
                  <a-menu-item key="3">
                    批量删除学员
                  </a-menu-item>
                </a-menu>
              </template>
              <a-button>
                批量操作
                <DownOutlined :style="{ fontSize: '10px' }" />
              </a-button>
            </a-dropdown>
            <a-dropdown class="mr-2" overlay-class-name="student-import-export-dropdown">
              <template #overlay>
                <a-menu @click="handleImportExportAction">
                  <a-menu-item key="1">
                    导入学员订单
                  </a-menu-item>
                  <a-menu-item key="2">
                    导入意向学员
                  </a-menu-item>
                  <a-menu-item key="3">
                    批量导出
                  </a-menu-item>
                  <a-menu-item key="4">
                    导出记录
                  </a-menu-item>
                </a-menu>
              </template>
              <a-button>
                导入/导出学员
                <DownOutlined :style="{ fontSize: '10px' }" />
              </a-button>
            </a-dropdown>
            <!-- 自定义字段 -->
            <customize-code 
              v-model:checked-values="selectedValues" 
              :options="columnOptions" 
              :total="allColumns.length"
              :num="selectedValues.length" 
            />
          </div>
        </div>
        <div class="table-content mt-2">
          <div class="tip">
            家校微信关注数为 {{ dataSource.filter(item => item.isBindChild).length }}，关注率 {{ dataSource.length > 0 ? ((dataSource.filter(item => item.isBindChild).length / dataSource.length) * 100).toFixed(2) : 0 }}%。引导家长关注家校平台，发送学员成果，提升续费率！ <a class="font500">点击下载家校物料（易拉宝、台卡等）</a>
          </div>
          <a-table 
            :data-source="dataSource" 
            row-key="id"
            :loading="loading"
            :pagination="pagination" 
            :columns="filteredColumns"
            :row-selection="rowSelection" 
            :scroll="{ x: totalWidth }" 
            :sticky="{ offsetHeader: 100 }"
            size="small"
            @change="handleTableChange"
          >
            <template #headerCell="{ column }">
              <!-- 动态自定义字段表头 -->
              <template v-if="customIsDisplayCodeList.some(item => item.key === column.key)">
                <clamped-text :text="column.title" :lines="1" />
              </template>
              <template v-if="column.key === 'studentStatus'">
                <span class="mr-1">{{ column.title }}</span>
                <a-tooltip color="#666">
                  <template #title>
                    在读学员：当前报读课程有一门或多门课程有剩余课时/天数/金额的学员。
                    历史学员：报读课程中全部课程都已结课的学员。
                  </template>
                  <ExclamationCircleOutlined />
                </a-tooltip>
              </template>
            </template>
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'name'">
                <student-avatar 
                  :id="record.id"
                  :name="record.stuName" 
                  :gender="record.stuSex === 0 ? '女' : record.stuSex === 1 ? '男' : '未知'" 
                  :age="calculateAge(record.birthDay)" 
                  :avatar-url="record.avatarUrl"
                  default-active-key="0" 
                />
              </template>
              <template v-if="column.key === 'mobile'">
                <div class="name">
                  <div class="text-#222" v-if="record.phoneRelationship">
                    {{ ParentRelationshipLabel[record.phoneRelationship] }}
                  </div>
                  <div class="text-3 text-#666">
                    {{ record.mobile || '-' }}
                  </div>
                </div>
              </template>
              <template v-if="column.key === 'isBindChild'">
                <a-tooltip placement="right">
                  <template #title>
                    <span>{{ record.isBindChild ? '已关注' : '点击邀请关注' }}</span>
                  </template>
                  <div class="flex flex-items-center cursor-pointer">
                    <span class="whitespace-nowrap" :class="record.isBindChild ? 'text-green-600' : 'text-#ccc'">
                      {{ record.isBindChild ? '已关注' : '未关注' }}
                    </span>
                    <svg v-if="!record.isBindChild" width="16px" height="16px" class="ml-2" viewBox="0 0 16 16">
                      <g id="\u9875\u9762-2" stroke="none" stroke-width="1" fill="none" fill-rule="evenodd">
                        <g id="\u753B\u677F\u5907\u4EFD-21" transform="translate(-474.000000, -608.000000)"
                          fill="#CCCCCC">
                          <g id="Rectangle-2\u5907\u4EFD-89" transform="translate(398.000000, 580.000000)">
                            <g id="\u7F16\u7EC4" transform="translate(76.000000, 21.600000)">
                              <g id="\u7F16\u7EC4" transform="translate(0.000000, 6.400000)">
                                <path id="\u8DEF\u5F84"
                                  d="M12.5488957,14.2844713 L11.5010486,14.2844713 C11.1341596,14.280754 10.8398076,13.9883197 10.843536,13.6312425 C10.8398076,13.2741654 11.1341596,12.9817311 11.5010486,12.9780138 L12.5488957,12.9780138 C13.1929132,12.9707828 13.7094253,12.457622 13.7035882,11.8308133 L13.7035882,5.51659915 C13.7049584,5.07149643 13.4426881,4.66546656 13.0299588,4.47372986 L8.49973266,2.41098807 C8.19497588,2.2717625 7.84236314,2.2717625 7.53760636,2.41098807 L3.00725203,4.47372986 C2.59455747,4.66549483 2.33231941,5.07151473 2.33368895,5.51659915 L2.33368895,8.11331051 C2.33741739,8.47038769 2.04306536,8.76282195 1.67617635,8.76653928 C1.30928733,8.76282195 1.0149353,8.47038769 1.01862871,8.11331051 L1.01862871,5.51659915 C1.01573797,4.56462047 1.57664141,3.69619985 2.4593492,3.28605311 L6.98970297,1.22331132 C7.64157303,0.925562892 8.39577156,0.925562892 9.04764162,1.22331132 L13.5778672,3.28605311 C14.460609,3.69617215 15.0215439,4.56460257 15.0186287,5.51659915 L15.0186287,11.8308133 C15.0186287,13.1837748 13.9107309,14.2844713 12.5488957,14.2844713 Z" />
                                <path id="\u8DEF\u5F84"
                                  d="M1.56733162,10.2194036 C1.40127346,10.2195233 1.23544109,10.2313173 1.07112909,10.2546935 C1.02383678,10.4730961 1,10.6956916 1,10.9188916 C1,11.7700045 1.34739282,12.5862583 1.96575882,13.1880863 C2.58412481,13.7899143 3.42280952,14.1280178 4.29731162,14.1280178 C4.51607755,14.1280178 4.73430678,14.1069519 4.94880175,14.0650857 C4.97598308,13.8947261 4.98963678,13.7225811 4.98963678,13.5501797 C4.98963678,12.666803 4.6290758,11.8196066 3.98726883,11.1949647 C3.34546185,10.5703228 2.4749842,10.2194036 1.56733162,10.2194036 Z" />
                                <path id="\u8DEF\u5F84"
                                  d="M4.04965057,14.1112242 C4.36580361,14.1624804 4.68574686,14.1883875 5.00625618,14.1886844 C6.55014367,14.1886844 8.03079835,13.5917848 9.12249298,12.529291 C10.2141876,11.4667972 10.8274979,10.0257453 10.8274979,8.5231503 C10.8271446,8.25723104 10.8076999,7.99166078 10.7693057,7.72837983 C10.4531526,7.67712408 10.1332094,7.65121719 9.81270009,7.65092023 C8.26881275,7.65092023 6.78815822,8.24781981 5.69646369,9.3103135 C4.60476916,10.3728072 3.99145897,11.813859 3.99145897,13.3164538 C3.99181199,13.582373 4.01125654,13.8479433 4.04965057,14.1112242 Z" />
                              </g>
                            </g>
                          </g>
                        </g>
                      </g>
                    </svg>
                  </div>
                </a-tooltip>
              </template>
              <template v-if="column.key === 'isCollect'">
                <a-tooltip placement="right">
                  <template #title>
                    <span>{{ record.isCollect ? '已采集' : '点击采集人脸' }}</span>
                  </template>
                  <div class="flex flex-items-center cursor-pointer">
                    <span class="whitespace-nowrap" :class="record.isCollect ? 'text-green-600' : 'text-#ccc'">
                      {{ record.isCollect ? '已采集' : '未采集' }}
                    </span>
                    <svg v-if="!record.isCollect" width="16px" height="16px" viewBox="0 0 16 16" class="ml-2">
                      <g id="\u9875\u9762-2" stroke="none" stroke-width="1" fill="none" fill-rule="evenodd">
                        <g id="\u753B\u677F\u5907\u4EFD-21" transform="translate(-594.000000, -608.000000)">
                          <g id="\u7F16\u7EC4-11" transform="translate(518.000000, 310.000000)">
                            <g id="Rectangle-2\u5907\u4EFD-88" transform="translate(0.000000, 270.000000)">
                              <g id="\u7F16\u7EC4" transform="translate(76.000000, 21.600000)">
                                <g id="\u7F16\u7EC4" transform="translate(0.000000, 6.400000)">
                                  <polygon id="\u77E9\u5F62" fill="#000000" fill-rule="nonzero" opacity="0"
                                    points="0 0 16 0 16 16 8 16 0 16" />
                                  <path id="\u5F62\u72B6"
                                    d="M1.49983336,11 C1.74529324,10.9999182 1.94950067,11.1767253 1.99191437,11.4099604 L2,11.4998334 L2,14 L4.5,14 C4.74545992,14 4.9496084,14.1768752 4.99194436,14.4101244 L5,14.5 C5,14.7454599 4.82312487,14.9496084 4.58987566,14.9919444 L4.5,15 L1.50100003,15 C1.25559799,15 1.05147725,14.8232051 1.00908211,14.5900195 L1.00100006,14.5001667 L1,11.5001667 C0.999908009,11.2240243 1.223691,11.0000921 1.49983336,11 Z M14.4988336,11 C14.7442935,10.9999183 14.9485009,11.1767254 14.9909146,11.4099605 L14.9990002,11.4998334 L15,14.4998334 C15.0000818,14.7453511 14.8231944,14.9495863 14.5898958,14.9919408 L14.5,15 L11.5,15 C11.2238576,15 11,14.7761424 11,14.5 C11,14.2545401 11.1768752,14.0503917 11.4101244,14.0080557 L14.4988336,11 Z M4.5,9 L11.5,9 L11.4931641,9.38828125 L11.4931641,9.38828125 L11.4769287,9.60498047 L11.4769287,9.60498047 L11.4453125,9.83125 C11.28125,10.75 10.625,11.8 8,11.8 C5.484375,11.8 4.77685547,10.8356771 4.5778656,9.94669189 L4.53663635,9.71717529 C4.53140259,9.67943522 4.5269165,9.64200846 4.52307129,9.60498047 L4.50683594,9.38828125 L4.50683594,9.38828125 L4.5,9 Z M11,5.5 C11.5522847,5.5 12,5.94771525 12,6.5 C12,7.05228475 11.5522847,7.5 11,7.5 C10.4477153,7.5 10,7.05228475 10,6.5 C10,5.94771525 10.4477153,5.5 11,5.5 Z M5,5.5 C5.55228475,5.5 6,5.94771525 6,6.5 C6,7.05228475 5.55228475,7.5 5,7.5 C4.44771525,7.5 4,7.05228475 4,6.5 C4,5.94771525 4.44771525,5.5 5,5.5 Z M14.5,1 C14.7455177,1 14.9496939,1.17695541 14.9919707,1.41026814 L15,1.50016663 L14.9990002,4.50016663 C14.9989082,4.77630898 14.774976,5.000092 14.4988336,5 C14.2533737,4.99991817 14.0492842,4.82297499 14.007026,4.58971169 L13.9990003,4.49983337 L14,2 L11.5,2 C11.2545401,2 11.0503916,1.82312484 11.0080557,1.58987563 L11,1.5 C11,1.25454011 11.1768752,1.05039163 11.4101244,1.00805567 L11.5,1 L14.5,1 Z M4.5,1 C4.77614235,1 5,1.22385763 5,1.5 C5,1.74545989 4.82312481,1.94960837 4.5898756,1.99194433 L4.5,2 L2,2 L2,4.50016667 C1.99991812,4.74562654 1.82297492,4.94971605 1.58971162,4.99197426 L1.49983331,5 C1.25437343,4.99991815 1.05028392,4.82297495 1.00802571,4.58971165 L1,4.49983333 L1.001,1.49983333 C1.0010818,1.25443131 1.17794474,1.05036951 1.41114451,1.0080521 L1.50099997,1 L4.5,1 Z"
                                    fill="#CCCCCC" />
                                </g>
                              </g>
                            </g>
                          </g>
                        </g>
                      </g>
                    </svg>
                  </div>
                </a-tooltip>
              </template>
              <template v-if="column.key === 'studentStatus'">
                <div class="flex flex-items-center studentStatus">
                  <span :class="record.studentStatus === 1?'dot':'lishidot'" />
                  <span>{{ record.studentStatus === 1 ? '在读学员' : '历史学员' }}</span>
                </div>
              </template>
              <template v-if="column.key === 'createName'">
                <clamped-text :text="record.createName || '-'" />
              </template>
              <template v-if="column.key === 'createTime'">
                <clamped-text :text="formatTime(record.createTime)" />
              </template>
              <template v-if="column.key === 'firstEnrolledTime'">
                <clamped-text :text="formatTime(record.firstEnrolledTime)" />
              </template>
              <template v-if="column.key === 'channelName'">
                <clamped-text :text="record.channelName || '-'" />
              </template>
              <template v-if="column.key === 'birthDay'">
                {{ record.birthDay || '-' }}
              </template>
              <template v-if="column.key === 'weChatNumber'">
                <clamped-text :lines="1" :text="record.weChatNumber || '-'" />
              </template>
              <template v-if="column.key === 'grade'">
                <clamped-text :lines="1" :text="record.grade || '-'" />
              </template>
              <template v-if="column.key === 'studySchool'">
                <clamped-text :lines="2" :text="record.studySchool || '-'" />
              </template>
              <template v-if="column.key === 'address'">
                <clamped-text :lines="2" :text="record.address || '-'" />
              </template>
              <template v-if="column.key === 'salePersonName'">
                <clamped-text :text="record.salePersonName || '-'" />
              </template>
              <template v-if="column.key === 'followUpTime'">
                <clamped-text :text="formatTime(record.followUpTime)" />
              </template>
              <!-- 动态自定义字段内容 customInfo 要处理customInfo为null的情况 -->
              <template v-if="customIsDisplayCodeList.some(item => item.key === column.key)">
                <clamped-text
                  :text="record.customInfo && record.customInfo.find(item => item.fieldName + item.fieldId === column.key)?.value || '-'" 
                />
              </template>
            </template>
          </a-table>
        </div>
      </div>
    </div>
    <student-info-drawer v-model:open="openDrawer" />

    <a-modal
      v-model:open="exportModalVisible"
      title="批量导出"
      :footer="null"
      :width="820"
      class="student-export-modal"
      destroy-on-close
    >
      <div class="export-tip-bar">
        <InfoCircleOutlined class="export-tip-icon" />
        <span>当前列表最多支持导出 10000 条数据。若超出，请前往【数据中心-报表管理-明细表】导出</span>
      </div>

      <div class="export-modal-content">
        <div class="export-row">
          <div class="export-label">
            查询条件：
          </div>
          <div class="export-query-box">
            <div v-for="item in exportQuerySummary" :key="item" class="export-query-line">
              {{ item }}
            </div>
          </div>
        </div>

        <div class="export-row export-row--compact">
          <div class="export-label">
            导出方式：
          </div>
          <a-radio-group v-model:value="exportMode" class="custom-radio export-radio-group">
            <a-radio value="all">
              全部导出
            </a-radio>
          </a-radio-group>
        </div>

        <div class="export-row export-row--compact">
          <div class="export-label">
            报表类型：
          </div>
          <a-radio-group v-model:value="exportReportType" class="custom-radio export-radio-group">
            <a-radio value="student">
              学员维度
            </a-radio>
          </a-radio-group>
        </div>

        <div class="export-row export-row--stacked">
          <div class="export-label">
            导出范例：
          </div>
          <div class="export-preview-title">
            共{{ exportFieldCount }}个字段
          </div>
          <div class="export-preview-card">
            <div class="export-preview-scroll">
              <a-table
                :data-source="exportPreviewRows"
                :columns="exportPreviewColumns"
                :pagination="false"
                size="small"
                :scroll="{ x: 3200 }"
                row-key="stuName"
              />
            </div>
          </div>
        </div>

        <div class="export-row export-row--compact">
          <div class="export-label">
            生成类型：
          </div>
          <a-radio-group v-model:value="exportFileType" class="custom-radio export-radio-group">
            <a-radio value="excel">
              EXCEL格式文件
            </a-radio>
          </a-radio-group>
        </div>
      </div>

      <div class="export-modal-footer">
        <a-button @click="handleViewExportRecord">
          查看导出记录
        </a-button>
        <a-button type="primary" class="ml-3" @click="handleSubmitExport">
          导出
        </a-button>
      </div>
    </a-modal>

    <a-modal
      v-model:open="exportRecordModalVisible"
      title="导出记录"
      :footer="null"
      :width="800"
      class="student-export-record-modal"
      destroy-on-close
    >
      <div class="export-record-card">
        <div class="export-record-header">
          <div class="export-record-meta">
            <span>报表生成时间：2026-03-27 11:34:06</span>
            <span class="ml-6">导出人：陈瑞</span>
          </div>
          <a-button @click="handleDownloadExportRecord">
            下载
          </a-button>
        </div>

        <div class="export-record-body">
          <div class="export-record-top">
            <div class="export-record-title">
              查询条件
            </div>
            <div class="export-record-expire">
              请在一周内下载，过期将失效
            </div>
          </div>
          <div class="export-record-grid">
            <div v-for="item in exportRecordConditions" :key="`${item.label}-${item.value}`" class="export-record-item">
              <span class="export-record-item-label">{{ item.label }}：</span>
              <span>{{ item.value }}</span>
            </div>
          </div>
        </div>
      </div>
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
  span.lishidot{
    border-radius: 50%;
    display: inline-block;
    height: 6px;
    position: relative;
    vertical-align: middle;
    width: 6px;
    margin-right: 4px;
    background: #888;
  }
}

.tip {
  padding: 10px 24px 10px 14px;
  background: #e6f0ff;
  color: #333;

  a {
    color: var(--pro-ant-color-primary);
  }
}

:deep(.student-export-modal .ant-modal-body),
:deep(.student-export-record-modal .ant-modal-body) {
  padding-top: 0;
}

.export-tip-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 0 -24px;
  padding: 12px 20px;
  background: #eaf3ff;
  color: #1668dc;
  font-size: 15px;
  line-height: 22px;
}

.export-tip-icon {
  flex-shrink: 0;
  font-size: 16px;
}

.export-modal-content {
  padding-top: 22px;
}

.export-row {
  display: flex;
  align-items: center;
  margin-bottom: 18px;
}

.export-row--compact {
  align-items: center;
  margin-bottom: 16px;
}

.export-row--block {
  display: flex;
  align-items: center;
}

.export-row--stacked {
  display: grid;
  grid-template-columns: 88px minmax(0, 1fr);
  row-gap: 12px;
  align-items: start;
}

.export-label {
  flex-shrink: 0;
  width: 88px;
  color: #595959;
  font-size: 15px;
  line-height: 22px;
}

.export-query-box {
  flex: 1;
  min-height: 56px;
  padding: 16px 18px;
  border-radius: 12px;
  background: #f5f7fb;
  color: #262626;
  font-size: 15px;
  line-height: 24px;
}

.export-query-line + .export-query-line {
  margin-top: 6px;
}

.export-radio-group {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
}

.export-radio-group :deep(.ant-radio-wrapper) {
  margin-right: 24px;
  color: #262626;
  font-size: 15px;
  line-height: 22px;
}

.export-preview-title {
  flex: 1;
  color: #262626;
  font-size: 15px;
  line-height: 22px;
}

.export-preview-card {
  flex: 1;
  overflow: hidden;
  border: 1px solid #edf0f5;
  border-radius: 12px;
  margin-top: 0;
}

.export-row--stacked .export-preview-card {
  grid-column: 2;
}

.export-preview-scroll {
  overflow-x: auto;
}

.export-modal-footer {
  display: flex;
  justify-content: flex-end;
  padding-top: 16px;
  border-top: 1px solid #f0f0f0;
}

.export-record-card {
  border: 1px solid #edf0f5;
  border-radius: 12px;
  overflow: hidden;
  background: #fff;
}

.export-record-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 24px;
  border-bottom: 1px solid #edf0f5;
}

.export-record-meta {
  color: #262626;
  font-size: 15px;
  font-weight: 600;
  line-height: 24px;
}

.export-record-body {
  padding: 18px 24px 20px;
}

.export-record-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 10px;
}

.export-record-title {
  color: #262626;
  font-size: 15px;
  font-weight: 600;
}

.export-record-expire {
  color: #1668dc;
  font-size: 14px;
  font-weight: 500;
}

.export-record-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px 40px;
}

.export-record-item {
  color: #262626;
  font-size: 15px;
  line-height: 24px;
}

.export-record-item-label {
  color: #595959;
}

.custom-radio ::v-deep(.ant-radio-wrapper:hover .ant-radio),
.custom-radio ::v-deep(.ant-radio:hover .ant-radio-inner),
.custom-radio ::v-deep(.ant-radio-input:focus + .ant-radio-inner) {
  border-color: var(--pro-ant-color-primary);
}

.custom-radio ::v-deep(.ant-radio-inner) {
  background-color: transparent;
  border-color: #d9d9d9;
}

.custom-radio ::v-deep(.ant-radio-checked .ant-radio-inner) {
  background-color: transparent;
  border-color: var(--pro-ant-color-primary);
}

.custom-radio ::v-deep(.ant-radio-inner::after) {
  background-color: var(--pro-ant-color-primary);
  transform: scale(0.5);
}

.upNew {
  position: relative;

  &::before {
    position: absolute;
    top: -12px;
    left: -22px;
    z-index: 1;
    width: 39px;
    height: 22px;
    background: url(data:image/png;base64,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);
    background-size: contain;
    content: "";
  }
}

:deep(.student-import-export-dropdown .ant-dropdown-menu) {
  min-width: 156px;
  padding: 8px 0;
  border-radius: 14px;
  box-shadow: 0 10px 28px rgba(15, 23, 42, 0.12);
}

:deep(.student-import-export-dropdown .ant-dropdown-menu-item) {
  height: 34px;
  padding: 0 16px;
  color: #262626;
  font-size: 14px;
  line-height: 34px;
}

:deep(.student-import-export-dropdown .ant-dropdown-menu-item:hover) {
  background: #f5f8ff;
}

/* 添加旋转动画 */
@keyframes icon-rotate {
  from {
    transform: rotate(0deg);
  }

  to {
    transform: rotate(180deg);
  }
}

.close-btn {
  &:hover {
    background: transparent;

    .close-icon {
      animation: icon-rotate 0.3s linear;
    }
  }
}

.tabs {
  width: 100%;
  border-radius: 10px;

  :deep(.ant-tabs-nav) {
    background: #fff;
    margin: 0;
  }

  :deep(.ant-tabs-ink-bar) {
    text-align: center;
    height: 12px !important;
    background: transparent;
    bottom: 0px !important;

    &::after {
      position: absolute;
      top: 0;
      left: calc(50% - 12px);
      width: 24px !important;
      height: 4px !important;
      border-radius: 2px;
      background-color: var(--pro-ant-color-primary);
      content: "";
    }
  }
}
</style>
