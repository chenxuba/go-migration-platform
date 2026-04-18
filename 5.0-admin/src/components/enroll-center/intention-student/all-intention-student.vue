<script setup>
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { debounce } from 'lodash-es'
import { CaretDownOutlined, CaretUpOutlined, DownOutlined, ExclamationCircleOutlined, InfoCircleOutlined } from '@ant-design/icons-vue'
import { Empty, Modal } from 'ant-design-vue'
import dayjs from 'dayjs'
import { addIntendedStudentApi, batchDeleteIntendedStudentApi, batchAssignSalespersonApi, batchTransferToPublicPoolApi, createStudentFollowUpApi, downloadIntentionStudentExportRecordApi, exportIntentionStudentsApi, getFollowUpCountApi, getIntentStudentListApi, getIntentionStudentExportRecordsApi, updateIntendedStudentApi, updateStatusApi } from '~@/api/enroll-center/intention-student'
import { useStudentFields } from '~@/composables/useStudentFields'
import messageService from '~@/utils/messageService'
import { useTableColumns } from '@/composables/useTableColumns'
import { calculateAge } from '@/utils/date'
import emitter, { EVENTS } from '~@/utils/eventBus'
import { FollowUpStatus, FollowUpStatusLabel, FollowUpStatusStyle, IntentionLevel, IntentionLevelLabel, IntentionLevelStyle, ParentRelationshipLabel } from '@/enums'
import DeleteConfirmModal from '@/components/common/DeleteConfirmModal.vue'
import { handleDateRangeParams } from '~@/utils/dateRangeParams'
import { useUserStore } from '~@/stores/user'
import { AccessEnum } from '@/constants/access'
import { useAccess } from '@/composables/access'

const props = defineProps({
  publicDataIsShow: {
    type: Boolean,
    default: false,
  },
})

const simpleImage = Empty.PRESENTED_IMAGE_SIMPLE

const userStore = useUserStore()
const router = useRouter()
const { hasAccess } = useAccess()

const canManageIntentionStudent = computed(() => hasAccess(AccessEnum.enroll_intention_manage))
const canEditIntentionFollowStatus = computed(() => hasAccess(AccessEnum.enroll_intention_follow_status))
const canEditFollowRecord = computed(() => hasAccess(AccessEnum.enroll_follow_edit))
const canImportIntentionStudent = computed(() => hasAccess(AccessEnum.enroll_intention_import))
const canExportIntentionStudent = computed(() => hasAccess(AccessEnum.enroll_intention_export))
const canViewIntentionExportRecords = computed(() => canExportIntentionStudent.value)
const canAssignIntentionSales = computed(() => hasAccess(AccessEnum.enroll_intention_assign_sales))
const canTransferIntentionPublicPool = computed(() =>
  props.publicDataIsShow && hasAccess(AccessEnum.enroll_intention_transfer_public_pool),
)
const visibleImportExportActions = computed(() => {
  const items = []
  if (canImportIntentionStudent.value) {
    items.push({
      key: '1',
      label: '导入意向学员',
    })
  }
  if (canExportIntentionStudent.value) {
    items.push({
      key: '2',
      label: '批量导出',
    })
  }
  if (canViewIntentionExportRecords.value) {
    items.push({
      key: '3',
      label: '导出记录',
    })
  }
  return items
})
const showImportExportDropdown = computed(() => visibleImportExportActions.value.length > 0)
const importExportButtonText = computed(() => {
  if (visibleImportExportActions.value.length === 1)
    return visibleImportExportActions.value[0].label
  return '导出/导入学员'
})
const visibleBatchActions = computed(() => {
  const items = []
  if (canAssignIntentionSales.value) {
    items.push({
      key: '1',
      label: '批量分配销售',
    })
  }
  if (canManageIntentionStudent.value) {
    items.push({
      key: '3',
      label: '批量删除学员',
    })
  }
  if (canTransferIntentionPublicPool.value) {
    items.push({
      key: '4',
      label: '批量转入公有池',
      class: 'whitespace-nowrap',
    })
  }
  return items
})
const showBatchDropdown = computed(() => visibleBatchActions.value.length > 0)

const displayArray = ref(['customSearch', 'intention', 'intentionCourse', 'salesPerson', 'hasSalesPerson', 'followStatus', 'lastFollowTime', 'nextFollowTime', 'notFollowDays', 'createTime', 'age', 'recommended', 'recommend', 'assignTime', 'trialPurchaseStatus', 'createUser'])

const assignSalesVisible = ref(false)
const modalType = ref(1)
const modalTitle = ref('')
const visible = ref(false)
const openDropdowns = ref(new Set())
const openStatusDropdowns = ref(new Set())
const openDeleteModal = ref(false)
const selectedRows = ref([])
const selectedRowKeys = ref([])
// 跨页选择：存储所有选中的数据
const allSelectedRows = ref(new Map()) // 使用Map存储，key为id，value为完整的行数据
const allSelectedRowKeys = ref(new Set()) // 使用Set存储所有选中的key
const btnLoading = ref(false)
const selectStuInfo = ref({})
const dataSource = ref([])
const exportModalVisible = ref(false)
const exportRecordModalVisible = ref(false)
const exportMode = ref('all')
const exportReportType = ref('student')
const exportFileType = ref('excel')
const exportSubmitting = ref(false)
const exportRecordsLoading = ref(false)
const exportRecords = ref([])
const exportModalConditionItems = ref([])
const exportConditionItems = ref([])
const exportPreviewColumns = [
  { title: '学员姓名', dataIndex: 'stuName', key: 'stuName', width: 140 },
  { title: '学员年龄', dataIndex: 'age', key: 'age', width: 120 },
  { title: '学员生日', dataIndex: 'birthDay', key: 'birthDay', width: 140 },
  { title: '学员性别', dataIndex: 'sex', key: 'sex', width: 120 },
  { title: '学员电话', dataIndex: 'mobile', key: 'mobile', width: 150 },
  { title: '电话关系', dataIndex: 'relation', key: 'relation', width: 120 },
  { title: '微信', dataIndex: 'wechat', key: 'wechat', width: 150 },
  { title: '转介绍推荐人', dataIndex: 'referralRecommender', key: 'referralRecommender', width: 150 },
  { title: '学员备注', dataIndex: 'remark', key: 'remark', width: 140 },
  { title: '意向度', dataIndex: 'intentionLevel', key: 'intentionLevel', width: 120 },
  { title: '意向课程', dataIndex: 'lessons', key: 'lessons', width: 160 },
  { title: '渠道分类', dataIndex: 'channelCategory', key: 'channelCategory', width: 120 },
  { title: '渠道', dataIndex: 'channel', key: 'channel', width: 120 },
  { title: '销售员', dataIndex: 'salesperson', key: 'salesperson', width: 120 },
  { title: '跟进状态', dataIndex: 'followUpStatus', key: 'followUpStatus', width: 120 },
  { title: '最近跟进', dataIndex: 'latestFollowTime', key: 'latestFollowTime', width: 180 },
  { title: '下次跟进时间', dataIndex: 'nextFollowTime', key: 'nextFollowTime', width: 180 },
  { title: '创建时间', dataIndex: 'createTime', key: 'createTime', width: 180 },
  { title: '创建人', dataIndex: 'creator', key: 'creator', width: 120 },
  { title: '是否被推荐', dataIndex: 'isRecommended', key: 'isRecommended', width: 120 },
  { title: '推荐人', dataIndex: 'recommender', key: 'recommender', width: 120 },
  { title: '最近试听课程', dataIndex: 'latestTrialCourse', key: 'latestTrialCourse', width: 160 },
  { title: '最近试听时间', dataIndex: 'latestTrialTime', key: 'latestTrialTime', width: 180 },
  { title: '分配销售时间', dataIndex: 'salesAssignedTime', key: 'salesAssignedTime', width: 180 },
  { title: '体验课购买状态', dataIndex: 'trialPurchaseStatus', key: 'trialPurchaseStatus', width: 160 },
  { title: '公有池倒计时', dataIndex: 'publicPoolCountdown', key: 'publicPoolCountdown', width: 140 },
  { title: '最近沟通内容', dataIndex: 'latestCommunication', key: 'latestCommunication', width: 160 },
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
    referralRecommender: '李晨',
    remark: '测试学员',
    intentionLevel: '高',
    lessons: '钢琴课',
    channelCategory: '推广',
    channel: '转介绍',
    salesperson: '孙勇',
    followUpStatus: '',
    latestFollowTime: '2023-02-14 10:36:35',
    nextFollowTime: '2023-02-18',
    createTime: '2022-10-18 14:46:50',
    creator: '王老师',
    isRecommended: '是',
    recommender: '王同学',
    latestTrialCourse: '钢琴课',
    latestTrialTime: '2023-02-14 09:36:35',
    salesAssignedTime: '2023-02-14 09:36:35',
    trialPurchaseStatus: '未购买',
    publicPoolCountdown: '3天',
    latestCommunication: '推广',
  },
]
const exportFieldCount = computed(() => exportPreviewColumns.length)
const exportQuerySummary = computed(() => {
  if (exportModalConditionItems.value.length === 0)
    return ['当前列表全部数据']
  return exportModalConditionItems.value.map(item => `${item.label}：${item.value}`)
})
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
    title: '渠道分类',
    dataIndex: 'channelCategoryName',
    key: 'channelCategoryName',
    width: 100,
  },
  {
    title: '渠道',
    dataIndex: 'channelName',
    key: 'channelName',
    width: 100,

  },
  {
    title: '销售员',
    dataIndex: 'teacher',
    key: 'teacher',
    width: 100,

  },
  {
    title: '跟进状态',
    key: 'status',
    dataIndex: 'status',
    width: 120,
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
    title: '创建人',
    key: 'createUser',
    width: 100,
  },
  {
    title: '是否被推荐',
    dataIndex: 'isRecommend',
    key: 'isRecommend',
    width: 110,
  },
  {
    title: '推荐人',
    dataIndex: 'recommendStudentName',
    key: 'recommendStudentName',
    width: 100,
  },
  {
    title: '生日',
    dataIndex: 'birthday',
    key: 'birthday',
    width: 120,
  },

  {
    title: '分配销售时间',
    key: 'salesAssignedTime',
    dataIndex: 'salesAssignedTime',
    width: 180,
    sorter: true,
  },
  {
    title: '体验课购买状态',
    key: 'experienceClassPurchaseStatus',
    dataIndex: 'experienceClassPurchaseStatus',
    width: 150,
  },
  {
    title: '操作',
    key: 'action',
    fixed: 'right',
    width: 180,
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
  // 保存操作列和倒计时列
  const actionColumn = allColumns.value.find(item => item.key === 'action')
  const countdownColumn = allColumns.value.find(item => item.key === 'countdown')

  // 获取所有非动态、非操作、非倒计时的基础列
  const baseColumns = allColumns.value.filter(col =>
    !col.isDynamic
    && col.key !== 'action'
    && col.key !== 'countdown',
  )

  // 合并所有需要显示的动态列
  const visibleSystemColumns = systemDefaultIsDisplayCodeList.value.filter(item => item.show)
  const allDynamicColumns = [...visibleSystemColumns, ...customIsDisplayCodeList.value]

  // 重新组装列顺序：基础列 -> 动态列 -> 倒计时列 -> 操作列
  allColumns.value = [
    ...baseColumns,
    ...allDynamicColumns,
  ]

  // 如果存在倒计时列，添加到动态列后面
  if (countdownColumn) {
    allColumns.value.push(countdownColumn)
  }

  // 添加操作列到最后
  if (actionColumn) {
    allColumns.value.push(actionColumn)
  }

  // 在所有列处理完成后执行倒计时列逻辑
  handleCountdownColumn()
}
// 处理公有池倒计时列的显示逻辑
function handleCountdownColumn() {
  const countdownColumns = {
    title: '公有池倒计时',
    key: 'countdown',
    dataIndex: 'countdown',
    width: 140,
    fixed: 'right',
    sorter: true,
  }

  watch(() => props.publicDataIsShow, (newVal) => {
    const currentColumns = allColumns.value
    const actionIndex = currentColumns.findIndex(col => col.key === 'action')
    const countdownIndex = currentColumns.findIndex(col => col.key === 'countdown')

    if (newVal) {
      if (countdownIndex === -1 && actionIndex !== -1) {
        currentColumns.splice(actionIndex, 0, countdownColumns)
        // 处理本地存储
        const intentionStudent = JSON.parse(localStorage.getItem('intention-student'))
        if (intentionStudent && !intentionStudent.includes('countdown')) {
          intentionStudent.push('countdown')
          localStorage.setItem('intention-student', JSON.stringify(intentionStudent))
        }
      }
    }
    else {
      if (countdownIndex !== -1) {
        currentColumns.splice(countdownIndex, 1)
      }
    }
  }, { immediate: true })
}
const { selectedValues, columnOptions, filteredColumns, totalWidth }
  = useTableColumns({
    storageKey: 'intention-student', // 本地存储键名
    allColumns, // 原始列配置
    excludeKeys: ['action'], // 需要排除的列键
    defaultSelectedKeys: ['mobile', 'intentionLevel', 'intentionCourse', 'channelCategoryName', 'channelName', 'teacher', 'status', 'followed', 'nextTime', 'createTime', 'createUser', 'isRecommend', 'recommendStudentName', 'birthday', 'salesAssignedTime', 'experienceClassPurchaseStatus', 'countdown'],
  })
// 创建学员
const openCreateStu = ref(false)
function handleAddStu() {
  openCreateStu.value = true
  modalType.value = 1
}
const openAddFollowUpModal = ref(false)
function handleAddFollowUp(record) {
  if (!canEditFollowRecord.value)
    return

  selectStuInfo.value = record
  openAddFollowUpModal.value = true
}
// 提交跟进记录
async function handleFollowUpSubmit(data) {
  try {
    const res = await createStudentFollowUpApi(data)
    if (res.code === 200) {
      messageService.success('添加跟进记录成功')
      openAddFollowUpModal.value = false
      getIntentStudentList()
    }
    emitter.emit(EVENTS.CLOSE_LOADING_EVENT)
  }
  catch (error) {
    console.log(error)
    // 关闭按钮loading
    emitter.emit(EVENTS.CLOSE_LOADING_EVENT)
  }
}
const rowSelection = {
  selectedRowKeys: selectedRowKeys,
  onChange: (keys, rows) => {
    // console.log(`selectedRowKeys: ${keys}`, 'selectedRows: ', rows);
    
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
    
    // console.log('全局选中数据:', Array.from(allSelectedRows.value.values()));
    // console.log('全局选中Keys:', Array.from(allSelectedRowKeys.value));
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
function onClickMenu(record, { key }) {
  // console.log(`Click on item ${key}`);
  // console.log(record);
  if (key === '1' || key === 1) {
    const sid = record?.id
    if (sid == null || sid === '') {
      messageService.error('学员信息不完整，无法报名')
      return
    }
    router.push({
      path: '/edu-center/registr-renewal',
      query: { id: String(sid) },
    })
  }
  else if (key === '2') {
    // 分配销售
    selectedRows.value = [record]
    selectedRowKeys.value = [record.id]
    assignSalesVisible.value = true
    modalType.value = 2
    modalTitle.value = '分配销售'
  }
  else if (key === '3') {
    // 编辑
    selectedRows.value = [record]
    selectedRowKeys.value = [record.id]
    modalType.value = 2
    openCreateStu.value = true
  }
  else if (key === '4') {
    // 删除
    selectedRows.value = [record]
    selectedRowKeys.value = [record.id]
    openDeleteModal.value = true
  }
  else if (key === '5') {
    // 放弃 - 使用Modal.confirm确认
    Modal.confirm({
      title: '确定放弃该学员？',
      content: '放弃后，您就不再是该学员的销售员，此学员将进入公有池',
      okText: '放弃',
      okType: 'danger',
      cancelText: '取消',
      centered: true,
      async onOk() {
        try {
          const res = await batchTransferToPublicPoolApi({ studentIds: [record.id] })
          if (res.code === 200) {
            messageService.success('放弃成功，学员已转入公有池')
            // 刷新列表
            getIntentStudentList()
          }
        } catch (error) {
          console.log(error)
          messageService.error('放弃失败')
        }
      }
    })
  }
}
function handleBatchOperation({ key }) {
  // console.log('批量操作')
  // 获取跨页选中的所有数据
  const allSelectedRowsArray = Array.from(allSelectedRows.value.values())
  
  if (key === '1') {
    // 批量分配销售
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
  else if (key === '2') {
    // 批量编辑学员
  }
  else if (key === '3') {
    // 批量删除学员
    if (allSelectedRowsArray.length === 0) {
      messageService.error('请选择要删除的学员')
      return
    }
    // 使用全局选中数据
    selectedRows.value = allSelectedRowsArray
    selectedRowKeys.value = Array.from(allSelectedRowKeys.value)
    openDeleteModal.value = true
  }
  else if (key === '4') {
    // 批量转入公有池
    if (allSelectedRowsArray.length === 0) {
      messageService.error('请选择要转入公有池的学员')
      return
    }
    // 使用全局选中数据
    selectedRows.value = allSelectedRowsArray
    selectedRowKeys.value = Array.from(allSelectedRowKeys.value)
    assignSalesVisible.value = true
    modalType.value = 3
    modalTitle.value = '批量转入公有池'
  }
}

function handleImportExportAction({ key }) {
  if (key === '1' && canImportIntentionStudent.value) {
    router.push('/import-center/starter/intentionStudent')
  }
  else if (key === '2' && canExportIntentionStudent.value) {
    openExportModal()
  }
  else if (key === '3' && canViewIntentionExportRecords.value) {
    openExportRecordModal()
  }
}

function buildFixedExportConditions(conditionMap = new Map()) {
  return [
    { label: '意向度', value: conditionMap.get('意向度') || '全部' },
    { label: '性别', value: conditionMap.get('性别') || '全部' },
    { label: '最近跟进', value: conditionMap.get('最近跟进') || '-' },
    { label: '下次跟进', value: conditionMap.get('下次跟进') || '-' },
  ]
}

function syncExportConditions() {
  const conditions = allFilterRef.value?.getOrderedConditions?.() || []
  const mappedConditions = conditions
    .map((item) => {
      const values = Array.isArray(item.values) ? item.values : []
      const displayValue = values.length > 0
        ? values.map(valueItem => `${valueItem?.value || ''}`.replace(' 至 ', ' ~ ')).filter(Boolean).join('、')
        : '全部'
      return {
        label: item.label,
        value: displayValue || '全部',
      }
    })
    .filter(item => item.label && item.value)

  const conditionMap = new Map(mappedConditions.map(item => [item.label, item.value]))
  const fixedConditions = buildFixedExportConditions(conditionMap)
  const extraConditions = mappedConditions.filter(item => !fixedConditions.some(fixed => fixed.label === item.label))

  if (mappedConditions.length === 0) {
    exportModalConditionItems.value = [{
      label: '导出范围',
      value: '当前列表全部数据',
    }]
    exportConditionItems.value = fixedConditions
    return
  }

  exportModalConditionItems.value = mappedConditions
  exportConditionItems.value = [...fixedConditions, ...extraConditions]
}

function buildExportQueryModel() {
  const dateRangeMappings = {
    createTime: {
      begin: 'createTimeBegin',
      end: 'createTimeEnd',
    },
    salesAssignedTime: {
      begin: 'salesAssignedTimeBegin',
      end: 'salesAssignedTimeEnd',
    },
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
  const normalizedQuery = handleDateRangeParams({ ...queryState.value }, dateRangeMappings)
  const originalFields = ['age', 'createTime', 'salesAssignedTime', 'birthday', 'lastFollowTime', 'nextFollowTime']
  return Object.fromEntries(
    Object.entries(normalizedQuery)
      .filter(([key, value]) => value !== undefined && value !== null && value !== '' && !originalFields.includes(key) && (!Array.isArray(value) || value.length > 0)),
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
    const res = await getIntentionStudentExportRecordsApi()
    exportRecords.value = res.result || res.data || []
  }
  catch (error) {
    console.error('load intention export records failed', error)
    messageService.error('获取导出记录失败')
  }
  finally {
    exportRecordsLoading.value = false
  }
}

async function downloadExportRecord(record) {
  try {
    const response = await downloadIntentionStudentExportRecordApi(record.id)
    triggerBlobDownload(response)
  }
  catch (error) {
    console.error('download intention export record failed', error)
    messageService.error('下载失败，请稍后重试')
  }
}

function triggerBlobDownload(response) {
  const blob = new Blob([response.data], { type: response.headers['content-type'] || 'application/octet-stream' })
  const disposition = response.headers['content-disposition'] || ''
  const matched = disposition.match(/filename\*=UTF-8''([^;]+)/i)
  const fileName = matched ? decodeURIComponent(matched[1]) : `意向学员批量导出-${dayjs().format('YYYYMMDDHHmmss')}.xlsx`
  const url = window.URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = fileName
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  window.URL.revokeObjectURL(url)
}

function getExportRecordDisplayConditions(record) {
  const recordConditions = Array.isArray(record?.queryConditions) ? record.queryConditions : []
  const conditionMap = new Map(recordConditions.map(item => [item.label, item.value]))
  const fixedConditions = buildFixedExportConditions(conditionMap)
  const hiddenLabels = new Set(['导出范围'])
  const extraConditions = recordConditions.filter(item =>
    item?.label
    && item?.value
    && !hiddenLabels.has(item.label)
    && !fixedConditions.some(fixed => fixed.label === item.label),
  )
  return [...fixedConditions, ...extraConditions]
}

async function handleViewExportRecord() {
  exportModalVisible.value = false
  await openExportRecordModal()
}

async function handleSubmitExport() {
  if (pagination.value.total === 0 || dataSource.value.length === 0) {
    messageService.error('没有符合条件的意向学员可以导出')
    return
  }
  exportSubmitting.value = true
  try {
    syncExportConditions()
    const res = await exportIntentionStudentsApi({
      queryModel: buildExportQueryModel(),
      sortModel: sortModel.value,
      queryConditions: exportConditionItems.value,
    })
    const record = res.result || res.data
    if (!record?.id)
      throw new Error(res.message || '导出失败')
    const response = await downloadIntentionStudentExportRecordApi(record.id)
    triggerBlobDownload(response)
    exportModalVisible.value = false
    if (canViewIntentionExportRecords.value)
      await openExportRecordModal()
  }
  catch (error) {
    console.error('export intention students failed', error)
    messageService.error(error?.message || '导出失败，请稍后重试')
  }
  finally {
    exportSubmitting.value = false
  }
}

// 处理批量删除
async function handleBatchDelete() {
  btnLoading.value = true
  try {
    const studentIds = selectedRows.value.map(row => row.id)
    const res = await batchDeleteIntendedStudentApi({ studentIds })
    if (res.code === 200) {
      messageService.success('删除成功')
      // 刷新列表
      getIntentStudentList()
      // 清空选择（包括跨页选择）
      clearAllSelection()
      // 关闭弹窗
      openDeleteModal.value = false
    }
  }
  catch (error) {
    console.log(error)
    messageService.error('删除失败')
  }
  finally {
    btnLoading.value = false
  }
}

// 处理分配销售和批量转入公有池
async function handleAssignSales(data) {
  try {
    let res
    let successMessage = ''
    let errorMessage = ''
    
    if (modalType.value === 3) {
      // 批量转入公有池
      const studentIds = selectedRows.value.map(row => row.id)
      res = await batchTransferToPublicPoolApi({ studentIds })
      successMessage = '转入公有池成功'
      errorMessage = '转入公有池失败'
    } else {
      // 分配销售
      res = await batchAssignSalespersonApi(data)
      successMessage = '分配销售成功'
      errorMessage = '分配销售失败'
    }
    
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

const loading = ref(false)
const createStudentRef = ref(null)
const assignSalesRef = ref(null)

// 定义字段映射关系
const fieldMappings = {
  age: ['age', 'ageMin', 'ageMax'],
  createTime: ['createTime', 'createTimeBegin', 'createTimeEnd'],
  birthday: ['birthday', 'birthDayBegin', 'birthDayEnd'],
  lastFollowTime: ['lastFollowTime', 'followUpTimeBegin', 'followUpTimeEnd'],
  nextFollowTime: ['nextFollowTime', 'nextFollowUpTimeBegin', 'nextFollowUpTimeEnd'],
  salesAssignedTime: ['salesAssignedTime', 'salesAssignedTimeBegin', 'salesAssignedTimeEnd'],
}

// 存储所有查询条件的响应式对象
const queryState = ref({
  queryAllOrDepartment: 1, // 全部（含所有部门）
  quickFilter: undefined,
  intentionLevels: undefined,
  followUpStatuses: undefined,
  createTimeBegin: undefined,
  createTimeEnd: undefined,
  birthDayBegin: undefined,
  birthDayEnd: undefined,
  followUpTimeBegin: undefined,
  followUpTimeEnd: undefined,
  nextFollowUpTimeBegin: undefined,
  nextFollowUpTimeEnd: undefined,
  salesAssignedTimeBegin: undefined,
  salesAssignedTimeEnd: undefined,
  sexes: undefined,
  ageMin: undefined,
  ageMax: undefined,
  channelIds: undefined,
  notFollowUpDay: undefined,
  // 原始字段
  age: undefined,
  createTime: undefined,
  birthday: undefined,
  lastFollowTime: undefined,
  nextFollowTime: undefined,
  salesAssignedTime: undefined,
  // 可以添加更多查询条件...
})

// 重置所有查询条件
function resetQueryState() {
  Object.keys(queryState.value).forEach((key) => {
    if (key !== 'queryAllOrDepartment') {
      queryState.value[key] = undefined
    }
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
  byCreatedTime: 0,
  byFollowUpTime: 0,
  byNextFlowTime: 0,
  byDaysUntilReturn: 0,
  bySalesAssignedTime: 0,
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
    createTime: {
      begin: 'createTimeBegin',
      end: 'createTimeEnd',
    },
    salesAssignedTime: {
      begin: 'salesAssignedTimeBegin',
      end: 'salesAssignedTimeEnd',
    },
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

    getFollowUpCount()
  }
  catch (error) {
    console.log('error: ', error)
  }
  finally {
    loading.value = false
  }
}

// 快捷筛选
function handleQuickFilter(quickFilter) {
  handleFilterUpdate({ quickFilter })
}

// 意向度筛选
function handleIntentionLevelFilter(intentionLevelFilter) {
  handleFilterUpdate({ intentionLevels: intentionLevelFilter })
}

// 处理表格排序变化
function handleTableChange(paginationInfo, filters, sorter) {
  // console.log('排序变化:', sorter);

  // 重置所有排序字段为0
  Object.keys(sortModel.value).forEach((key) => {
    sortModel.value[key] = 0
  })

  // 处理排序逻辑（支持单列排序和多列排序）
  const sortFieldMap = {
    'createTime': 'byCreatedTime',
    'followed': 'byFollowUpTime',
    'nextTime': 'byNextFlowTime',
    'countdown': 'byDaysUntilReturn',
    'salesAssignedTime': 'bySalesAssignedTime',
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

async function handleCreateStu(data) {
  try {
    // 创建一个新对象用于API调用，避免修改原始数据
    const apiData = { ...data }

    // 处理渠道ID
    if (apiData.channelId && apiData.channelId.length === 1) {
      apiData.channelId = apiData.channelId[0]
    }
    else if (apiData.channelId && apiData.channelId.length > 1) {
      apiData.channelId = apiData.channelId[1]
    }

    const res = modalType.value === 1 ? await addIntendedStudentApi(apiData) : await updateIntendedStudentApi(apiData)

    if (res.code === 200) {
      messageService.success(modalType.value === 1 ? '创建成功' : '编辑成功')
      getIntentStudentList()
      // 如果是编辑操作，清空选择状态（包括跨页选择）
      if (modalType.value === 2) {
        clearAllSelection()
      }
      openCreateStu.value = false
      createStudentRef.value.resetForm()
    }
  }
  catch (error) {
    console.log('error: ', error)
  }
  finally {
    createStudentRef.value.closeSpinning()
  }
}

function toggleDropdown(recordId) {
  // Clear all dropdowns first
  openDropdowns.value.clear()
  openStatusDropdowns.value.clear()

  // Then add the current one if it wasn't previously open
  if (!openDropdowns.value.has(recordId)) {
    openDropdowns.value.add(recordId)
  }
}

function toggleStatusDropdown(recordId) {
  // Clear all dropdowns first
  openDropdowns.value.clear()
  openStatusDropdowns.value.clear()

  // Then add the current one if it wasn't previously open
  if (!openStatusDropdowns.value.has(recordId)) {
    openStatusDropdowns.value.add(recordId)
  }
}

// Add a method to handle dropdown close from outside
function handleDropdownVisibleChange(visible, recordId, type) {
  if (!visible) {
    if (type === 'intention') {
      openDropdowns.value.delete(recordId)
    }
    else {
      openStatusDropdowns.value.delete(recordId)
    }
  }
}

// 获取状态配置的计算属性
const getStatusConfig = computed(() => (status) => {
  return {
    text: FollowUpStatusLabel[status] || '-',
    className: FollowUpStatusStyle[status]?.className || '',
  }
})

// 使用防抖包装状态变更处理函数
const debouncedStatusChange = debounce(async (record, status) => {
  loading.value = true
  try {
    const res = await updateStatusApi({
      id: record.id,
      uuid: record.uuid,
      version: record.version,
      followUpStatus: status,
    })
    if (res.code === 200) {
      messageService.success('更新状态成功')
      // 关闭下拉框
      openStatusDropdowns.value.delete(record.id)
      // 刷新列表
      getIntentStudentList()
    }
  }
  catch (error) {
    messageService.error('更新状态失败')
    console.error('更新状态失败:', error)
    loading.value = false
  }
}, 300, { leading: true, trailing: false })

async function handleStatusChange(record, status) {
  if (record.followUpStatus === status) {
    // 可选：提示用户状态未变更
    // 关闭下拉框
    openStatusDropdowns.value.delete(record.id)
    return // 直接返回，不执行后续操作
  }
  debouncedStatusChange(record, status)
}

// 使用防抖包装意向度变更处理函数
const debouncedIntentionChange = debounce(async (record, intentLevel) => {
  loading.value = true
  try {
    const res = await updateStatusApi({
      id: record.id,
      uuid: record.uuid,
      version: record.version,
      intentLevel,
    })
    if (res.code === 200) {
      messageService.success('更新意向度成功')
      // 关闭下拉框
      openDropdowns.value.delete(record.id)
      // 刷新列表
      getIntentStudentList()
    }
  }
  catch (error) {
    messageService.error('更新意向度失败')
    console.error('更新意向度失败:', error)
    loading.value = false
  }
}, 300, { leading: true, trailing: false })

async function handleIntentionChange(record, intentLevel) {
  if (record.intentLevel === intentLevel) {
    // 如果选择的意向度与当前相同，直接关闭下拉框
    openDropdowns.value.delete(record.id)
    return
  }
  debouncedIntentionChange(record, intentLevel)
}
const followUpCount = ref({})
// 获取跟进数量 getFollowUpCountApi
async function getFollowUpCount() {
  try {
    const res = await getFollowUpCountApi({ queryAllOrDepartment: 1 })
    followUpCount.value = res.result
  }
  catch (error) {
    console.log('error: ', error)
  }
}

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
  quickFilter: 'quickFilter',
  intentionLevelFilter: 'intentionLevels',
  intentionCourseFilter: 'courseId',
  createTimeFilter: 'createTime',
  birthdayFilter: 'birthday',
  lastFollowTimeFilter: 'lastFollowTime',
  nextFollowTimeFilter: 'nextFollowTime',
  assignTimeFilter: 'salesAssignedTime',
  followStatusFilter: 'followUpStatuses',
  sexFilter: 'sexes',
  ageFilter: 'age',
  channelFilter: 'channelIds',
  notFollowDaysFilter: 'notFollowUpDay',
  wxChatFilter: 'wechatNumber',
  schoolFilter: 'schoolSearchKey',
  addressFilter: 'addressSearchKey',
  hobbiesFilter: 'interestSearchKey',
  gradeFilter: 'grades',
  stuPhoneSearchFilter: 'studentId',
  tuiJianUserFilter: 'recommendStudentId',
  recommendedFilter: 'isRecommend',
  createUserFilter: 'createId',
  salesPersonFilter: 'salespersonId',
  hasSalesPersonFilter: 'isHasSalePerson',
  trialPurchaseStatusFilter: 'purchasedAuditionProduct',
  customSearchInputFilter: 'customFieldSearchList'
}

// 生成所有过滤器的更新处理器
const filterUpdateHandlers = computed(() => {
  const handlers = {}
  Object.entries(filterFieldMapping).forEach(([eventKey, fieldName]) => {
    handlers[`update:${eventKey}`] = (val, isClearAll, id, type) => {
      if (fieldName === 'purchasedAuditionProduct') {
        if (Array.isArray(val) && val.length > 0) {
          if (val.includes(1) && !val.includes(2)) {
            handleFilterUpdate({ purchasedAuditionProduct: true }, isClearAll, id, type)
          } else if (val.includes(2) && !val.includes(1)) {
            handleFilterUpdate({ purchasedAuditionProduct: false }, isClearAll, id, type)
          } else {
            handleFilterUpdate({ purchasedAuditionProduct: undefined }, isClearAll, id, type)
          }
        } else {
          handleFilterUpdate({ purchasedAuditionProduct: undefined }, isClearAll, id, type)
        }
        return
      }
      handleFilterUpdate({ [fieldName]: val }, isClearAll, id, type)
    }
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

// 暴露方法给父组件调用
defineExpose({
  getIntentStudentList
})
</script>

<template>
  <div class="tab-content">
    <all-filter ref="allFilterRef" :display-array="displayArray" :is-quick-show="true" :is-show-search-stu-phone="true"
      :custom-is-display-list="customIsDisplaySearchList" :student-status="0" :follow-up-count="followUpCount"
      :grade-options-data="gradeOptionsData" v-on="filterUpdateHandlers" />
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
          <a-button v-if="canManageIntentionStudent" type="primary" class="mr-2" @click="handleAddStu">
            创建学员
          </a-button>
          <a-button type="primary" class="mr-2">
            安排试听
          </a-button>
          <a-dropdown v-if="showImportExportDropdown" class="mr-2">
            <template #overlay>
              <a-menu @click="handleImportExportAction">
                <a-menu-item v-for="item in visibleImportExportActions" :key="item.key">
                  {{ item.label }}
                </a-menu-item>
              </a-menu>
            </template>
            <a-button>
              {{ importExportButtonText }}
              <DownOutlined :style="{ fontSize: '10px' }" />
            </a-button>
          </a-dropdown>
          <a-tooltip placement="topLeft">
            <template #title>
              可选择多种营销短信模板
            </template>
            <a-button class="mr-2">
              群发短信
            </a-button>
          </a-tooltip>
          <a-dropdown v-if="showBatchDropdown" class="mr-2">
            <template #overlay>
              <a-menu @click="handleBatchOperation">
                <a-menu-item v-for="item in visibleBatchActions" :key="item.key" :class="item.class">
                  {{ item.label }}
                </a-menu-item>
              </a-menu>
            </template>
            <a-button>
              批量操作
              <DownOutlined :style="{ fontSize: '10px' }" />
            </a-button>
          </a-dropdown>
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
            <template v-if="column.key === 'countdown'">
              <div class="flex">
                <clamped-text :text="column.title" :lines="1" />
                <a-popover color="#fff" title="放回到公有池倒计时剩余天数">
                  <template #content>
                    1. 未跟进剩余天数倒计时结束后此线索将放回至公有池 <br>
                    2. 添加跟进后倒计时将重新计算<br>
                    3. 当前线索分配至另一位销售后倒计时将重新计算<br>
                    4. 设置剩余天数：请至移动端"我的-业务设置-"
                  </template>
                  <ExclamationCircleOutlined class="ml-4px" />
                </a-popover>
              </div>
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
              <div :style="{ cursor: canManageIntentionStudent ? 'pointer' : 'default' }">
                <template v-if="!canManageIntentionStudent">
                  <div class="intention">
                    <span class="intentionTag"
                      :style="{ background: IntentionLevelStyle[record.intentLevel]?.color }" />
                    {{ IntentionLevelLabel[record.intentLevel] || IntentionLevelLabel[IntentionLevel.Unknown] }}
                  </div>
                </template>
                <a-dropdown v-else :trigger="['click']" :open="openDropdowns.has(record.id)"
                  @update:open="(val) => handleDropdownVisibleChange(val, record.id, 'intention')">
                  <div @click.prevent="toggleDropdown(record.id)">
                    <div class="intention">
                      <span class="intentionTag"
                        :style="{ background: IntentionLevelStyle[record.intentLevel]?.color }" />
                      {{ IntentionLevelLabel[record.intentLevel] || IntentionLevelLabel[IntentionLevel.Unknown] }}
                      <CaretDownOutlined v-if="!openDropdowns.has(record.id)" class="ml-1 text-#ccc"
                        :style="{ 'font-size': '10px' }" />
                      <CaretUpOutlined v-if="openDropdowns.has(record.id)" class="ml-1 text-#1677ff"
                        :style="{ 'font-size': '10px' }" />
                    </div>
                  </div>
                  <template #overlay>
                    <a-menu style="text-align: center;width: 70px;"
                      @click="({ key }) => handleIntentionChange(record, Number(key))">
                      <a-menu-item v-for="level in Object.values(IntentionLevel).filter(v => !isNaN(Number(v)))"
                        :key="level">
                        {{ IntentionLevelLabel[level] }}
                      </a-menu-item>
                    </a-menu>
                  </template>
                </a-dropdown>
              </div>
            </template>
            <!-- 意向课程 -->
            <template v-if="column.key === 'intentionCourse'">
              <clamped-text :text="record.lessons && Array.isArray(record.lessons) && record.lessons.length > 0
                ? record.lessons.map(course => course.name).join(', ')
                : '-'" />
            </template>
            <!-- 渠道分类 -->
            <template v-if="column.key === 'channelCategoryName'">
              <clamped-text :lines="1" :text="record.channelCategoryName || '-'" />
            </template>
            <!-- 渠道 -->
            <template v-if="column.key === 'channelName'">
              <clamped-text :text="record.channelName || '-'" />
            </template>
            <!-- 销售员 -->
            <template v-if="column.key === 'teacher'">
              <clamped-text :text="record.salePersonName || '-'" />
            </template>
            <!-- 跟进状态 -->
            <template v-if="column.key === 'status'">
              <div :style="{ cursor: canEditIntentionFollowStatus ? 'pointer' : 'default' }">
                <template v-if="!canEditIntentionFollowStatus">
                  <div class="intention">
                    <span class="statusTag" :class="getStatusConfig(record.followUpStatus).className">
                      {{ getStatusConfig(record.followUpStatus).text }}
                    </span>
                  </div>
                </template>
                <a-dropdown v-else :trigger="['click']" :open="openStatusDropdowns.has(record.id)"
                  @update:open="(val) => handleDropdownVisibleChange(val, record.id, 'status')">
                  <div @click.prevent="toggleStatusDropdown(record.id)">
                    <div class="intention">
                      <span class="statusTag" :class="getStatusConfig(record.followUpStatus).className">
                        {{ getStatusConfig(record.followUpStatus).text }}
                      </span>
                      <component :is="openStatusDropdowns.has(record.id) ? CaretUpOutlined : CaretDownOutlined"
                        class="ml-1px" :class="openStatusDropdowns.has(record.id) ? 'text-#1677ff' : 'text-#ccc'"
                        :style="{ fontSize: '10px' }" />
                    </div>
                  </div>
                  <template #overlay>
                    <a-menu style="text-align: center;width: 100px;"
                      @click="({ key }) => handleStatusChange(record, Number(key))">
                      <a-menu-item v-for="status in Object.values(FollowUpStatus).filter(v => !isNaN(Number(v)))"
                        :key="status">
                        {{ FollowUpStatusLabel[status] }}
                      </a-menu-item>
                    </a-menu>
                  </template>
                </a-dropdown>
              </div>
            </template>
            <!-- 最近跟进 -->
            <template v-if="column.key == 'followed'">
              <clamped-text :text="formatTime(record.followUpTime) || '-'" />
            </template>
            <!-- 下次跟进 -->
            <template v-if="column.key == 'nextTime'">
              <clamped-text :class="{ 'text-#f90': isOverdue(record.nextFollowUpTime) && record.followUpInterviewStatus === 1 }"
                :text="formatTime(record.nextFollowUpTime) || '-'" />
            </template>
            <!-- 创建时间 -->
            <template v-if="column.key == 'createTime'">
              <clamped-text :text="formatTime(record.createTime) || '-'" />
            </template>
            <!-- 创建人 -->
            <template v-if="column.key === 'createUser'">
              <clamped-text :text="record.createName || '-'" />
            </template>
            <!-- 生日 -->
            <template v-if="column.key === 'birthday'">
              {{ record.birthDay ? dayjs(record.birthDay).format('YYYY-MM-DD') : '-' }}
            </template>
            <!-- 是否被推荐 -->
            <template v-if="column.key === 'isRecommend'">
              <div class="flex items-center">
                <a-badge :status="record.isRecommend == 1 ? 'success' : 'default'" class="mr--4px" />{{
                  record.isRecommend == 1 ? '是' : '否' }}
              </div>
            </template>
            <!-- 推荐人 -->
            <template v-if="column.key === 'recommendStudentName'">
              <clamped-text :text="record.recommendStudentName || '-'" />
            </template>
            <!-- 分配销售时间 -->
            <template v-if="column.key === 'salesAssignedTime'">
              <clamped-text :text="formatTime(record.salesAssignedTime) || '-'" />
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
            <!-- 自定义字段展示 -->
            <!-- <template v-for="item in record.customInfo" :key="item.fieldId">
              <template v-if="column.key == item.fieldName + item.fieldId">
                <clamped-text :lines="2" :text="item.value || '-'"></clamped-text>
              </template>
            </template> -->
            <!-- 动态自定义字段内容 customInfo 要处理customInfo为null的情况 -->
            <template v-if="customIsDisplayCodeList.some(item => item.key === column.key)">
              <clamped-text
                :text="record.customInfo && record.customInfo.find(item => item.fieldName + item.fieldId === column.key)?.value || '-'" />
            </template>
            <!-- 体验课购买状态 -->
            <template v-if="column.key === 'experienceClassPurchaseStatus'">
              <clamped-text :text="record.experienceClassPurchaseStatus || '-'" />
            </template>
            <template v-if="column.key === 'countdown'">
              <span :class="{ 'text-red-500': record.daysUntilReturn !== null && record.daysUntilReturn !== undefined && record.daysUntilReturn <= 3 }">
                {{ record.daysUntilReturn !== null && record.daysUntilReturn !== undefined ? record.daysUntilReturn + '天' : '-' }}
              </span>
            </template>
            <template v-else-if="column.key === 'action'">
              <a-space>
                <a v-if="canEditFollowRecord" @click="handleAddFollowUp(record)">添加跟进</a>
                <a>试听</a>
                <div style="cursor: pointer;">
                  <a-dropdown placement="bottomRight" :arrow="true">
                    <a @click.prevent>
                      <div class="intention">更多
                        <CaretDownOutlined class="ml-1 text-#1677ff" :style="{ 'font-size': '12px' }" />
                      </div>
                    </a>
                    <template #overlay>
                      <a-menu style="text-align: center;width: 100px;" @click="onClickMenu(record, $event)">
                        <a-menu-item key="1">
                          报名
                        </a-menu-item>
                        <a-menu-item v-if="canAssignIntentionSales" key="2">
                          分配销售
                        </a-menu-item>
                        <a-menu-item key="5" v-if="canTransferIntentionPublicPool && record.salePerson == userStore.userInfo.instUserId">
                          放弃
                        </a-menu-item>
                        <a-menu-item v-if="canManageIntentionStudent" key="3">
                          编辑
                        </a-menu-item>
                        <a-menu-item v-if="canManageIntentionStudent" key="4">
                          删除
                        </a-menu-item>
                      </a-menu>
                    </template>
                  </a-dropdown>
                </div>
              </a-space>
            </template>
          </template>
        </a-table>
      </div>
    </div>
  </div>
  <create-student ref="createStudentRef" v-model:open="openCreateStu" :record="selectedRows[0]" :type="modalType"
    @submit="handleCreateStu" />
  <!-- 跟进记录 -->
  <add-follow-up-modal v-model:open="openAddFollowUpModal" :record="selectStuInfo"
    @handle-follow-up-submit="handleFollowUpSubmit" />
  <!-- 分配销售 -->
  <assign-sales-modal ref="assignSalesRef" v-model:open="assignSalesVisible" :type="modalType" :title="modalTitle" 
    :selected-students="selectedRows" @submit="handleAssignSales" />
  <!-- 批量删除确认弹窗 -->
  <DeleteConfirmModal v-model:open="openDeleteModal"
    :student-names="(selectedRows || []).map(row => row.stuName).join('、')" :student-count="selectedRows?.length || 0"
    :loading="btnLoading" @confirm="handleBatchDelete" @cancel="openDeleteModal = false" />
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
              :scroll="{ x: 4200 }"
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
      <a-button v-if="canViewIntentionExportRecords" @click="handleViewExportRecord">
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
      <div v-if="exportRecords.length > 0" class="export-record-list">
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
  font-size: 14px;
  line-height: 22px;
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
  font-size: 14px;
  font-weight: 600;
}

.export-record-expire {
  color: #1668dc;
  font-size: 13px;
}

.export-record-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px 40px;
}

.export-record-item {
  color: #262626;
  font-size: 14px;
  line-height: 22px;
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

.hover {
  &:hover {
    .name {
      color: #06f;
    }
  }
}
</style>
