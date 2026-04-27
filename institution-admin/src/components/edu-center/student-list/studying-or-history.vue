<script setup>
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { debounce } from 'lodash-es'
import { DownOutlined, ExclamationCircleOutlined, InfoCircleOutlined } from '@ant-design/icons-vue'
import { Empty } from 'ant-design-vue'
import dayjs from 'dayjs'
import {
  downloadEnrolledStudentExportRecordApi,
  exportEnrolledStudentsApi,
  getEnrolledStudentExportRecordsApi,
  getEnrolledStudentListApi,
} from '~@/api/edu-center/enrolled-student'
import { getInstConfigModuleApi } from '@/api/common/config'
import { useStudentFields } from '@/composables/useStudentFields'
import { useTableColumns } from '@/composables/useTableColumns'
import { useRouter } from 'vue-router'
import { calculateAge } from '@/utils/date'
import { ParentRelationshipLabel, StudentStatus, StudentStatusLabel } from '@/enums'
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

const simpleImage = Empty.PRESENTED_IMAGE_SIMPLE

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
const exportModalConditionItems = ref([])
const exportConditionItems = ref([])
const exportSubmitting = ref(false)
const exportRecordsLoading = ref(false)
const exportRecords = ref([])
const supervisorEnabled = ref(false)

const exportPreviewColumns = [
  { title: '学员姓名', dataIndex: 'stuName', key: 'stuName' },
  { title: '学员年龄', dataIndex: 'age', key: 'age' },
  { title: '学员生日', dataIndex: 'birthDay', key: 'birthDay' },
  { title: '学员性别', dataIndex: 'sex', key: 'sex' },
  { title: '学员电话', dataIndex: 'mobile', key: 'mobile' },
  { title: '电话关系', dataIndex: 'relation', key: 'relation' },
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
    wechat: '18818888888',
    remark: '测试学员',
    bindStatus: '已关注',
    faceStatus: '已采集',
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
  if (exportModalConditionItems.value.length === 0) {
    return ['全部导出']
  }
  return exportModalConditionItems.value.map(item => `${item.label}：${item.value}`)
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
    title: '销售员',
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

const supervisorColumn = {
  title: '督导',
  key: 'supervisorName',
  dataIndex: 'supervisorName',
  width: 100,
  isDynamic: true,
  columnLabel: '督导',
}

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
  const salePersonColumnIndex = baseColumns.findIndex(col => col.key === 'salePersonName')

  // 合并所有需要显示的动态列
  const visibleSystemColumns = systemDefaultIsDisplayCodeList.value.filter(item => item.show)
  const trailingDynamicColumns = [...visibleSystemColumns, ...customIsDisplayCodeList.value]

  const columnsWithSupervisor = (() => {
    if (!supervisorEnabled.value || salePersonColumnIndex === -1)
      return baseColumns

    return [
      ...baseColumns.slice(0, salePersonColumnIndex),
      supervisorColumn,
      ...baseColumns.slice(salePersonColumnIndex),
    ]
  })()

  // 重新组装列顺序：基础列 -> 动态列
  allColumns.value = [
    ...columnsWithSupervisor,
    ...trailingDynamicColumns,
  ]

  void nextTick(() => {
    if (!supervisorEnabled.value) {
      if (selectedValues.value.includes(supervisorColumn.key)) {
        selectedValues.value = selectedValues.value.filter(key => key !== supervisorColumn.key)
      }
      return
    }

    if (!selectedValues.value.includes(supervisorColumn.key)) {
      selectedValues.value = [...selectedValues.value, supervisorColumn.key]
    }
  })
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

function isConfigEnabled(value) {
  if (typeof value === 'boolean')
    return value
  if (typeof value === 'number')
    return value !== 0
  if (typeof value === 'string')
    return value === '1' || value.toLowerCase() === 'true'
  return false
}

async function loadSupervisorConfig() {
  try {
    const res = await getInstConfigModuleApi('enrollment')
    supervisorEnabled.value = isConfigEnabled(res.result?.enableSupervisor)
    updateDynamicColumns()
  }
  catch (error) {
    console.error('获取督导配置失败:', error)
  }
}

function handleImportExportAction({ key }) {
  switch (key) {
    case '1':
      router.push('/import-center/starter/order')
      break
    case '2':
      router.push('/import-center/starter/intentionStudent')
      break
    case '3':
      openExportModal()
      break
    case '4':
      openExportRecordModal()
      break
  }
}

function getDefaultStudentStatusLabel() {
  return props.currentType === 1 ? '在读学员' : props.currentType === 2 ? '历史学员' : '全部'
}

function buildFixedExportConditions(conditionMap = new Map()) {
  return [
    { label: '学员状态', value: conditionMap.get('学员状态') || getDefaultStudentStatusLabel() },
    { label: '家校通', value: conditionMap.get('家校通') || '全部' },
    { label: '人脸采集', value: conditionMap.get('人脸采集') || '全部' },
    { label: '性别', value: conditionMap.get('性别') || '全部' },
    { label: '创建时间', value: conditionMap.get('创建时间') || '-' },
  ]
}

function syncExportConditions() {
  const conditions = allFilterRef.value?.getOrderedConditions?.() || []
  const mappedConditions = conditions.map((item) => {
    const values = Array.isArray(item.values) ? item.values : []
    const displayValue = values.length > 0
      ? values.map(valueItem => `${valueItem?.value || ''}`.replace(' 至 ', ' ~ ')).filter(Boolean).join('、')
      : '全部'
    return {
      label: item.label,
      value: displayValue || '全部',
    }
  })
  const conditionMap = new Map(mappedConditions.map(item => [item.label, item.value]))
  const fixedConditions = buildFixedExportConditions(conditionMap)
  const extraConditions = mappedConditions.filter(item => !fixedConditions.some(fixed => fixed.label === item.label))
  exportModalConditionItems.value = [
    { label: '学员状态', value: conditionMap.get('学员状态') || getDefaultStudentStatusLabel() },
    ...mappedConditions.filter(item => item.label !== '学员状态'),
  ]
  exportConditionItems.value = [...fixedConditions, ...extraConditions]
}

function getExportRecordDisplayConditions(record) {
  const recordConditions = Array.isArray(record?.queryConditions) ? record.queryConditions : []
  const conditionMap = new Map(recordConditions.map(item => [item.label, item.value]))
  return buildFixedExportConditions(conditionMap)
}

function buildExportQueryModel() {
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
  const normalizedQuery = handleDateRangeParams({ ...queryState.value }, dateRangeMappings)
  return Object.fromEntries(
    Object.entries(normalizedQuery)
      .filter(([, value]) => value !== undefined && value !== null && value !== '' && (!Array.isArray(value) || value.length > 0)),
  )
}

async function openExportModal() {
  syncExportConditions()
  exportModalVisible.value = true
}

async function openExportRecordModal() {
  syncExportConditions()
  exportRecordModalVisible.value = true
  await loadExportRecords()
}

async function loadExportRecords() {
  exportRecordsLoading.value = true
  try {
    const res = await getEnrolledStudentExportRecordsApi()
    exportRecords.value = res.result || res.data || []
  }
  catch (error) {
    console.error('load export records failed', error)
    messageService.error('获取导出记录失败')
  }
  finally {
    exportRecordsLoading.value = false
  }
}

async function downloadExportRecord(record) {
  try {
    const response = await downloadEnrolledStudentExportRecordApi(record.id)
    triggerBlobDownload(response)
  }
  catch (error) {
    console.error('download export record failed', error)
    messageService.error('下载失败，请稍后重试')
  }
}

function triggerBlobDownload(response) {
  const blob = new Blob([response.data], { type: response.headers['content-type'] || 'application/octet-stream' })
  const disposition = response.headers['content-disposition'] || ''
  const matched = disposition.match(/filename\*=UTF-8''([^;]+)/i)
  const fileName = matched ? decodeURIComponent(matched[1]) : `学员批量导出-${dayjs().format('YYYYMMDDHHmmss')}.xlsx`
  const url = window.URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = fileName
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  window.URL.revokeObjectURL(url)
}

async function handleViewExportRecord() {
  exportModalVisible.value = false
  await openExportRecordModal()
}

async function handleSubmitExport() {
  if (pagination.value.total === 0 || dataSource.value.length === 0) {
    messageService.error('没有符合条件的报读列表可以导出')
    return
  }
  exportSubmitting.value = true
  try {
    syncExportConditions()
    const res = await exportEnrolledStudentsApi({
      queryModel: buildExportQueryModel(),
      queryConditions: exportConditionItems.value,
    })
    const record = res.result || res.data
    if (!record?.id) {
      throw new Error(res.message || '导出失败')
    }
    const response = await downloadEnrolledStudentExportRecordApi(record.id)
    triggerBlobDownload(response)
    exportModalVisible.value = false
    await openExportRecordModal()
  }
  catch (error) {
    console.error('export enrolled students failed', error)
    messageService.error(error?.message || '导出失败，请稍后重试')
  }
  finally {
    exportSubmitting.value = false
  }
}

function handleDownloadExportRecord() {
  messageService.info('请选择具体导出记录下载')
}

function formatTime(time) {
  return time ? dayjs(time).format('YYYY-MM-DD HH:mm') : '-'
}

function formatBirthday(value) {
  if (!value)
    return '-'
  const date = dayjs(value)
  return date.isValid() ? date.format('YYYY-MM-DD') : value
}

function getStudentStatusClass(status) {
  if (status === StudentStatus.Intention)
    return 'is-intention'
  if (status === StudentStatus.History)
    return 'is-history'
  return 'is-reading'
}

onMounted(() => {
  getEnrolledStudentList()
  getAllStuFields({ filter: 3 })
  loadSupervisorConfig()
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
                    <img
                      class="ml-2 follow-bind-icon"
                      :class="{ 'follow-bind-icon--inactive': !record.isBindChild }"
                      src="~@/assets/images/follow.svg"
                      alt=""
                    >
                  </div>
                </a-tooltip>
              </template>
              <template v-if="column.key === 'isCollect'">
                <a-tooltip placement="right">
                  <template #title>
                    <span>{{ record.isCollect ? '已采集' : '点击采集人脸' }}</span>
                  </template>
                  <div class="flex flex-items-center cursor-pointer">
                    <template v-if="record.isCollect">
                      <span class="whitespace-nowrap text-#222">
                        已采集
                      </span>
                      <img class="ml-2 face-collect-icon face-collect-icon--collected" src="~@/assets/images/face.svg" alt="">
                    </template>
                    <template v-else>
                      <span class="whitespace-nowrap text-#ccc">
                        未采集
                      </span>
                      <img class="ml-2 face-collect-icon" src="~@/assets/images/face.svg" alt="">
                    </template>
                  </div>
                </a-tooltip>
              </template>
              <template v-if="column.key === 'studentStatus'">
                <div class="status-cell" :class="getStudentStatusClass(record.studentStatus)">
                  <span class="dot" />
                  <span>{{ StudentStatusLabel[record.studentStatus] || '-' }}</span>
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
                {{ formatBirthday(record.birthDay) }}
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
              <template v-if="column.key === 'supervisorName'">
                <clamped-text :text="record.supervisorName || '-'" />
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
        <a-button type="primary" class="ml-3" :loading="exportSubmitting" @click="handleSubmitExport">
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
      <a-spin :spinning="exportRecordsLoading">
        <div class="export-record-list" v-if="exportRecords.length > 0">
          <div v-for="record in exportRecords" :key="record.id" class="export-record-card">
            <div class="export-record-header">
              <div class="export-record-meta">
                <span>报表生成时间：{{ record.createdTime ? dayjs(record.createdTime).format('YYYY-MM-DD HH:mm:ss') : '-' }}</span>
                <span class="ml-6">导出人：{{ record.exporterName || '-' }}</span>
              </div>
              <a-button @click="downloadExportRecord(record)">
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
                <div v-for="item in getExportRecordDisplayConditions(record)" :key="`${record.id}-${item.label}-${item.value}`" class="export-record-item">
                  <span class="export-record-item-label">{{ item.label }}：</span>
                  <span>{{ item.value }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
        <div v-else-if="!exportRecordsLoading" class="export-record-empty">
          <a-empty :image="simpleImage" description="暂无数据" />
        </div>
      </a-spin>
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

.status-cell {
  display: inline-flex;
  align-items: center;
  color: #222;

  .dot {
    width: 6px;
    height: 6px;
    margin-right: 6px;
    border-radius: 999px;
    background: #52c41a;
  }

  &.is-intention .dot {
    background: #faad14;
  }

  &.is-history .dot {
    background: #d9d9d9;
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

.export-record-list {
  max-height: 520px;
  overflow-y: auto;
  padding-right: 4px;
}

.export-record-empty {
  min-height: 220px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.export-record-card {
  border: 1px solid #edf0f5;
  border-radius: 12px;
  overflow: hidden;
  background: #fff;
  margin-bottom: 16px;
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
    background: url(data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAE4AAAAsCAYAAADLlo5MAAAAAXNSR0IArs4c6QAABjtJREFUaEPtm3lo1EcUxz+zRrwtgmiNf4hBvEFkd0m8Fa1XbdGWBlERFVsFj1ovPEGsfxk86omK4IEiFg/EQkHFekATknjfSETQKKKoVfFKdsrbybq7yR6//e3+4prkwWJI3nsz8913z6hIgrTWipycbHy+b/H5slAqE8hEa/m3aRKqUyeq1CvgEVCK1qW4XCW4XH+Rn1+glNJ2F1J2BLXXOwStfwK+R+uv7ej47DJKPQaOodSfqrDwZKL7SQg4nZ2dQ1nZaqBfogulOf85MjIWqoKCfKv7tASc9nqz0DoPrX+wqviL5FPqMEotUIWFJfH2Hxc4v1v6fAeBFvGU1ZC/P8flyo3nvjGB0273LJRah9b1aggo1o6hVDla/6aKizdGE4gKnHa71wO/WlupxnL9oYqL50Q6XUTg/JYGG2osHIkdbHYky6sCXEWp8Xetc8+oPqnKUWp45ZgXBpw/e/p8RbUoEVi1PUkYntBsGw6cx3OoxpccVqGqzKfUYVVU9GPg15+Aqyhu/7Wrt1bIZWT0ChTJQeDc7nNA35QC0KULTJliVC5dCh8+2FffsiUsXgxZWbBsGVy/bl2XywXdukH9+nDhgnW5qpznVXGxv2vyA1dR5J5IRmNE2X79YN068yf5+e3b5JbYvBmys+H4cVixoqqujAwQgAOfVq2gZ08j07w5PH8Oo0fDmzf29+FyfSOJwgDndm8HfravLYpkssBNngwDBgSVt2gBbdvCx49w+3b4otu2QY8eMHVq5M1obWTWrIGLF+0fVantqqhomvKPhrxeGbmkfsqRLHDikmIhVmj5cmjXzgAnFnXzJpSWms+9e1BUBC9fWtEUm0emKoWFmcrRpJAscJ07Q2YmNG1qYtuVK8FDNWgAbjcUFEB5Ody4YUAW4M6ehblzkwcpmgZJEtrr/R2fb5kjqyQLnGyqQwfYtQvevYPhw6GszGxVXFjc7u5dGDvW/G769OoBzuVapbTbvQ8Yl7bAycYOHjQWN2cOnD9vtirJYdQoA+qmTdULHOxX2uM5jdYDHQduy5bY5YiUKgJQKPXqBU2aQP/+MHIk5OfD0aOGQ8qbZs1gwwYTx0pKYOhQY3Hi0lu3Rj/SpUsmwdglpf4R4G6jdUe7OmLKhbpqvAUkcA8eHM516JAJ+FZoxw5QKnpWDdUhX8KTJ1a0RuZR6o64qlxmOHOxEgqcfMsSxKORZMLKAX3lSmjdOijRuDFIUS1UWZ/UdlKqiMWJNQVqNUkijRqZtV/JUTEx8elT+8DBa7G4/9C6WTJaosqmIjmEKu/UCfZJSAYGDoTXr8OXjpQccnNh4UK4dQsmTEjZMavPVe10Dg0bGmsJkGTYQOwaMyYcuBcvYNq0qlnVQeCqJznYAW7iRJg925qVDBsG48eDyJw8CYsWGTnHgEvnckRca8aMIHAS/KUfFZJ6TtqoAElpsmABDBkCu3fDxorrAseAS/cCOF6Mk+D//r3h2rMHunaFVauCZYtjwJlLZmfmcKlIDu3bw9q1JoseOBBMDpIIpD+9fz/ozqdOwVdfmQ5CelNHXTWdm3w5+KRJMHOmKX7F/QJZVWqxI0egXj0YMcIU12fOGLDEbR/LCwcHY5zo1h7PNrT+xVoUToArFRYnLVX37rB6NVy+HF6OSNslZUlengFKelcBsE+fYPxzylX9wJnb+vQbZEqxu3dv0IrEDUPruL59TTy7ds0MATweY3Xz5gW/XSeB84Pndp9N+DGNVODSfEejNm1A+k2hY8eCk41YRvvwocmKQuvXg4Ajjb00+JULYMmqs2bBnTuwZImRkc5B4mGAHAfOTpKQqUROTgK+a4FVGnS5p5Bpr4AtBbCAIe4qHyk3JIsOGhQcGsyfb9qoq1dBpsah5DRwFbEusevBceNiW5wFnKqwPHhgRkVCYrHSIchkZf9+6FgxizhxwlzcBEj62Z07TYw7ffozAJfOF9IyxJSJsCQIybCVL35kUvzoUXhRLBBKXde7Nzx7ZrJwiqjuCYRNIOse3aQSOH+8q3vmFRPSuoeFqba4gL5a+JTVEpRx3wD73ba2PJ62BJlhsgTcJ+szRXJeyh/nJLDhdGFNCLhK7puLUt858nQiXdCJsQ9bwH0C8Ev4L0kOfQn/A6jssToWH7guAAAAAElFTkSuQmCC);
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

.face-collect-icon {
  width: 16px;
  height: 16px;
}

.follow-bind-icon {
  width: 16px;
  height: 16px;
}

.follow-bind-icon--inactive {
  filter: grayscale(1) opacity(0.45);
}

.face-collect-icon--collected {
  filter: brightness(0) saturate(100%) invert(49%) sepia(88%) saturate(3657%) hue-rotate(208deg) brightness(101%) contrast(101%);
}
</style>
