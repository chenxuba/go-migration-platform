<script setup>
import { computed, nextTick, onBeforeUnmount, onMounted, onUnmounted, ref, toRaw, watch, watchEffect } from 'vue'
import { debounce } from 'lodash-es'
import {
  CloseOutlined,
  DeleteOutlined,
  SearchOutlined,
} from '@ant-design/icons-vue'
import {
  FollowMethod,
  FollowMethodLabel,
  FollowUpStatus,
  FollowUpStatusLabel,
  IntentionLevel,
  IntentionLevelLabel,
  IsCommonCourse,
  IsCommonCourseLabel,
  IsCommonYesNo,
  IsCommonYesNoLabel,
  SellStatus,
  SellStatusLabel,
  Sex,
  SexLabel,
  StudentStatus,
  StudentStatusLabel,
  TeachingMethod,
  TeachingMethodLabel,
  VisitStatus,
  VisitStatusLabel,
} from '@/enums'
import { getChannelTreeApi, getRecommenderPageApi } from '~@/api/enroll-center/intention-student'
import { getCourseCategoryPageApi, getCoursePropertyOptionsApi } from '~@/api/edu-center/course-list'
import { getCourseIdAndNameApi } from '~@/api/edu-center/registr-renewal'
import { getOrderTagListPagedApi } from '@/api/finance-center/order-tag'
import { getUserListApi, getListTreeDepartApi } from '~@/api/internal-manage/staff-manage'
import { useUserStore } from '~@/stores/user'

const props = defineProps({
  createTimeLabel: {
    type: String,
    default: '创建时间',
  },
  studentStatus: {
    type: Number,
    default: null,
  },
  defaultStudentStatus: {
    type: Number,
    default: null,
  },
  defaultAccountStatus: {
    type: String,
    default: null,
  },
  defaultCreateTimeVals: {
    type: Array,
    default: () => {
      return []
    },
  },
  defaultOrderStatusVals: {
    type: Array,
    default: () => {
      return []
    },
  },
  defaultOrderTagIds: {
    type: Array,
    default: () => {
      return []
    },
  },
  defaultOrderArrearStatusVals: {
    type: Array,
    default: () => {
      return []
    },
  },
  defaultEnableStatus: {
    type: [Boolean, Number],
    default: null,
  },
  defaultOrNotFenClass: {
    type: Array,
    default: () => {
      return []
    },
  },
  defaultCurrentStatus: {
    type: Array,
    default: () => {
      return []
    },
  },
  defaultOpenClassStatus: {
    type: Number,
    default: null,
  },
  type: {
    type: String,
    default: 'all',
  },
  isShowChangTeacherSearch: {
    type: Boolean,
    default: false,
  },
  isQuickShow: {
    type: Boolean,
    default: false,
  },
  isApproveQuickShow: {
    type: Boolean,
    default: false,
  },
  isQuickOneToOneShow: {
    type: Boolean,
    default: false,
  },
  displayArray: {
    type: Array,
    default: () => {
      return []
    },
  },
  // 新增：是否需要获取部门数据
  needDepartmentData: {
    type: Boolean,
    default: false,
  },
  isShowSearchStuPhone: {
    type: Boolean,
    default: false,
  },
  isShowSearchStuPhonefilter: {
    type: Boolean,
    default: false,
  },
  isShowOneToOne: {
    type: Boolean,
    default: false,
  },
  isShowSearchInput: {
    type: Boolean,
    default: false,
  },
  isShowClsssOrCourseSearch: {
    type: Boolean,
    default: false,
  },
  renderClassListOptions: {
    type: Boolean,
    default: true,
  },
  searchPlaceholder: {
    type: String,
    default: '请输入关键字',
  },
  searchLabel: {
    type: String,
    default: '搜索纬度',
  },
  createUserLabel: {
    type: String,
    default: '创建人',
  },
  salesPersonLabel: {
    type: String,
    default: '销售员',
  },
  createUserPlaceholder: {
    type: String,
    default: '请输入创建人',
  },
  salesPersonPlaceholder: {
    type: String,
    default: '请输入销售员',
  },
  selectDptVals: {
    type: [String, Number],
    default: null,
  },
  channelCategories: {
    type: Array,
    default: () => [],
  },
  channelPositionRole: {
    type: Array,
    default: () => [],
  },
  customIsDisplayList: {
    type: Array,
    default: () => [],
  },
  // 课程属性开启列表
  courseAttributeList: {
    type: Array,
    default: () => [],
  },
  followUpCount: {
    type: Object,
    default: () => ({}),
  },
  approveQuickCounts: {
    type: Object,
    default: () => ({}),
  },
  // 隐藏指定的快捷筛选项（传入id数组）
  hideQuickFilters: {
    type: Array,
    default: () => [],
  },
  // 自定义快捷筛选的传值（key为筛选项id，value为要传递的值）
  customQuickFilterValues: {
    type: Object,
    default: () => ({}),
  },
  searchFilterList: {
    type: Array,
    default: () => [],
  },
  // 年级选项数据
  gradeOptionsData: {
    type: Array,
    default: () => [],
  },
  // // 课程类别
  // courseCategory: {
  //   type: Array,
  //   default: () => [],
  // },
})
const emit = defineEmits(['update:channelTypeFilter', 'update:channelStatusFilter', 'update:channelCategoryFilter',
  'update:quickFilter', 'update:intentionLevelFilter', 'update:createTimeFilter', 'update:birthdayFilter',
  'update:lastFollowTimeFilter', 'update:nextFollowTimeFilter', 'update:assignTimeFilter', 'update:followStatusFilter',
  'update:sexFilter', 'update:ageFilter', 'update:channelFilter',
  'update:notFollowDaysFilter', 'update:wxChatFilter', 'update:schoolFilter', 'update:addressFilter',
  'update:hobbiesFilter', 'update:orderNumberFilter', 'update:gradeFilter', 'update:stuPhoneSearchFilter', 'update:tuiJianUserFilter',
  'update:recommendedFilter', 'update:followMethodFilter', 'update:visitStatusFilter', 'update:payrollStatusFilter',
  'update:stuStatusFilter', 'update:followTimeFilter', 'update:courseCategoryFilter', 'update:isCommonCourseFilter',
  'update:teachingMethodFilter', 'update:sellStatusFilter', 'update:chargingMethodFilter', 'update:hasTrialPriceFilter',
  'update:trialPurchaseStatusFilter',
  'update:isMicroSchoolSaleFilter', 'update:isMicroSchoolDisplayFilter',
  'update:lastEditedTimeFilter', 'update:channelPositionRoleFilter', 'update:channelUserType',
  'update:channelAccountStatus', 'update:performanceAllocationStatusFilter',
  'update:enableStatusFilter',
  'update:orderTypeFilter', 'update:orderTagFilter', 'update:enrollTypeFilter', 'update:productTypeFilter', 'update:approveNumberFilter', 'update:approvalStatusFilter', 'update:finishTimeFilter', 'update:departmentFilter', 'staff-search', 'update:createUserFilter',
  'update:salesPersonFilter', 'update:hasSalesPersonFilter', 'searchInputFun', 'update:customSearchInputFilter',
  'update:courseAttributeFilter',
  'update:orderStatusFilter', 'update:orderSourceFilter', 'update:dealDateFilter', 'update:latestPaidTimeFilter', 'update:handleContentFilter', 'update:orderArrearStatusFilter',
  'update:intentionCourseFilter', 'update:currentStatusFilter', 'update:orNotFenClassFilter',
  'update:isSetExpirationDateFilter', 'update:classEndingTimeFilter', 'update:classStopTimeFilter',
  'update:expiryDateFilter', 'update:classNameFilter', 'update:enrolledCourseFilter', 
  'update:validityPeriodFilter', 'update:classTeacherFilter', 'update:isArrearsFilter', 
  'update:lastClassTimeFilter', 'update:remainingFilter', 'update:billingModeFilter',])
const spinning = ref(false)

// 用户store
const userStore = useUserStore()
const userDeptIds = computed(() => userStore.deptIds || [])
// 配置参数
const DEBOUNCE_TIME = 300 // 防抖时间（毫秒）
// 实时响应的中间值（用于v-model绑定）
const inputValue = ref('')
// 创建防抖函数（确保单例）
const triggerDebounce = debounce((value) => {
  searchInputKey.value = value
  console.log('执行防抖后的操作:', value) // 替换为实际业务逻辑
  emit('searchInputFun', value)
}, DEBOUNCE_TIME)

// 是否跳过下次watch触发（用于避免重复调用）
let skipNextWatch = false

// 监听输入框变化
watch(inputValue, (newVal) => {
  if (skipNextWatch) {
    skipNextWatch = false
    return
  }
  triggerDebounce(newVal)
})

// 组件卸载时清理
onBeforeUnmount(() => {
  triggerDebounce.cancel() // 清除未执行的防抖任务
  debouncedEmit.cancel()
})
// 创建通用的防抖emit函数
const debouncedEmit = debounce((eventName, value) => {
  emit(`update:${eventName}`, value)
}, 500)

function getNameById(id) {
  if (id === undefined || id === null || id === '')
    return null
  const search = (nodes) => {
    for (const node of nodes) {
      const nid = node.id
      if (nid === id || String(nid) === String(id) || Number(nid) === Number(id))
        return node.name || node.departName || null
      if (node.children) {
        const result = search(node.children)
        if (result)
          return result
      }
    }
    return null
  }
  return search(dptListOptions.value)
}
// 添加一个递归查找函数，用于在嵌套结构中查找节点
function findNodeById(nodes, id) {
  for (const node of nodes) {
    if (node.id === id) {
      return node
    }
    if (node.channelList && node.channelList.length > 0) {
      const found = findNodeById(node.channelList, id)
      if (found)
        return found
    }
  }
  return null
}

const dptName = ref('')
const searchKeyOneToOne = ref(undefined) // 一对一搜索框的值
const searchKeyStuPhone = ref(undefined) // 学员/电话搜索的值（真正用于请求 & 传参的 studentId）
// 学员/电话下拉框的 v-model，使用 labelInValue 结构，保证有名字可用于回显
const searchKeyStuPhoneModel = ref()
const searchInputKey = ref(undefined)
const selectInputKey = ref(undefined)
const teacherType = ref(1)
const selectTeacher = ref(undefined)
const childRefs = ref({}) // 存储子组件实例（按 category 分类）
// 动态收集/清理子组件实例
function handleRef(el, category) {
  if (el) {
    childRefs.value[category] = el // 组件挂载时存储实例
  }
  else {
    delete childRefs.value[category] // 组件卸载时删除实例
  }
}
const lastUpdated = reactive({})
const conditionOrder = ref([]) // 存储条件类型的添加顺序
// 未跟进天数筛选相关
const notFollowDaysInput = ref(null) // 输入框
const notFollowDaysQuick = ref(null) // 快捷选择
const notFollowDaysQuickOptions = [7, 15, 30, 60]
const notFollowDaysSelected = ref(null) // 已选的未跟进天数

const minAge = ref(null)
const maxAge = ref(null)
const selectedAgeRange = ref(null)



// 员工搜索结果 - 新增用于存储搜索结果
const searchResultOptions = ref([])
// 员工搜索相关状态
const searchFilterPagination = ref({
  current: 1,
  pageSize: 10,
  total: 0,
})
const searchFilterFinished = ref(false)
const searchFilterLoading = ref(false)
const searchFilterSearchKey = ref('')

// 意向度选项
const customOptions = ref([
  { id: IntentionLevel.High, value: IntentionLevelLabel[IntentionLevel.High] },
  { id: IntentionLevel.Medium, value: IntentionLevelLabel[IntentionLevel.Medium] },
  { id: IntentionLevel.Low, value: IntentionLevelLabel[IntentionLevel.Low] },
  { id: IntentionLevel.Unknown, value: IntentionLevelLabel[IntentionLevel.Unknown] },
])
const selectedValues = ref([])

// 年级选项 - 使用传入的数据或默认值
const gradeOptions = computed(() => {
  if (props.gradeOptionsData && props.gradeOptionsData.length > 0) {
    return props.gradeOptionsData.map(item => ({
      id: item,
      value: item
    }))
  }
  // 默认年级选项
  return [
    { id: "一年级", value: '一年级' },
    { id: "二年级", value: '二年级' },
    { id: "三年级", value: '三年级' },
    { id: "四年级", value: '四年级' },
    { id: "五年级", value: '五年级' },
    { id: "六年级", value: '六年级' },
    { id: "七年级", value: '七年级' },
    { id: "八年级", value: '八年级' },
    { id: "九年级", value: '九年级' },
    { id: "高一", value: '高一' },
  ]
})
const gradeVals = ref([])

// 是否有销售员选项
const hasSalesPersonOptions = ref([
  { id: 1, value: '有销售员' },
  { id: 0, value: '无销售员' },
])
const hasSalesPersonVals = ref([])
// 跟进状态选项
const followStatusOptions = ref([
  // 使用枚举
  { id: FollowUpStatus.Pending, value: FollowUpStatusLabel[FollowUpStatus.Pending] },
  { id: FollowUpStatus.InProgress, value: FollowUpStatusLabel[FollowUpStatus.InProgress] },
  { id: FollowUpStatus.NoAnswer, value: FollowUpStatusLabel[FollowUpStatus.NoAnswer] },
  { id: FollowUpStatus.Invited, value: FollowUpStatusLabel[FollowUpStatus.Invited] },
  { id: FollowUpStatus.Audited, value: FollowUpStatusLabel[FollowUpStatus.Audited] },
  { id: FollowUpStatus.Visited, value: FollowUpStatusLabel[FollowUpStatus.Visited] },
  { id: FollowUpStatus.Invalid, value: FollowUpStatusLabel[FollowUpStatus.Invalid] },
])
const followStatusVals = ref([])

// 性别选项
const sexOptions = ref([
  { id: Sex.Male, value: SexLabel[Sex.Male] },
  { id: Sex.Female, value: SexLabel[Sex.Female] },
  { id: Sex.Unknown, value: SexLabel[Sex.Unknown] },
])
const sexVals = ref([])

// 是否是通用课选项
const isCommonCourseOptions = ref([
  { id: IsCommonCourse.NotCommon, value: IsCommonCourseLabel[IsCommonCourse.NotCommon] },
  { id: IsCommonCourse.AllCommon, value: IsCommonCourseLabel[IsCommonCourse.AllCommon] },
  { id: IsCommonCourse.PartCommon, value: IsCommonCourseLabel[IsCommonCourse.PartCommon] },
])
const isCommonCourseVals = ref([])

// 跟进方式选项
const followMethodOptions = ref([
  { id: FollowMethod.None, value: FollowMethodLabel[FollowMethod.None] },
  { id: FollowMethod.Phone, value: FollowMethodLabel[FollowMethod.Phone] },
  { id: FollowMethod.WeChat, value: FollowMethodLabel[FollowMethod.WeChat] },
  { id: FollowMethod.Interview, value: FollowMethodLabel[FollowMethod.Interview] },
  { id: FollowMethod.Other, value: FollowMethodLabel[FollowMethod.Other] },
])
const followMethodVals = ref([])

// 回访状态选项
const visitStatusOptions = ref([
  { id: VisitStatus.Visited, value: VisitStatusLabel[VisitStatus.Visited] },
  { id: VisitStatus.NotVisited, value: VisitStatusLabel[VisitStatus.NotVisited] },
])
const visitStatusVals = ref([])

// 工资条状态选项
const payrollStatusOptions = ref([
  { id: 1, value: '待确认' },
  { id: 2, value: '已确认' },
  { id: 3, value: '已作废' },
])
const payrollStatusVals = ref([])

// 学员状态选项
const stuStatusOptions = ref([
  { id: StudentStatus.Intention, value: StudentStatusLabel[StudentStatus.Intention] },
  { id: StudentStatus.Reading, value: StudentStatusLabel[StudentStatus.Reading] },
  { id: StudentStatus.History, value: StudentStatusLabel[StudentStatus.History] },
])
const stuStatusVals = ref([])

// 创建人选项
const createUserOptions = ref([])
const createUserVals = ref(null)
const selectedCreateUserInfo = ref(null) // 缓存已选创建人的详细信息

// 销售员选项 - 独立的数据源和状态
const salesPersonOptions = ref([])
const selectSalesPersonVals = ref(null)
const selectedSalesPersonInfo = ref(null) // 缓存已选销售员的详细信息

// 授课方式选项
const teachingMethodOptions = ref([
  { id: TeachingMethod.Class, value: TeachingMethodLabel[TeachingMethod.Class] },
  { id: TeachingMethod.OneToOne, value: TeachingMethodLabel[TeachingMethod.OneToOne] },
])
const teachingMethodVals = ref(null)

// 售卖状态选项
const sellStatusOptions = ref([
  { id: SellStatus.OnSale, value: SellStatusLabel[SellStatus.OnSale] },
  { id: SellStatus.StopSale, value: SellStatusLabel[SellStatus.StopSale] },
])
const sellStatusVals = ref(null)

// 是否有体验价选项
const hasTrialPriceOptions = ref([
  { id: IsCommonYesNo.Yes, value: IsCommonYesNoLabel[IsCommonYesNo.Yes] },
  { id: IsCommonYesNo.No, value: IsCommonYesNoLabel[IsCommonYesNo.No] },
])
const hasTrialPriceVals = ref(null)

// 是否开启微校售卖选项
const isMicroSchoolSaleOptions = ref([
  { id: IsCommonYesNo.Yes, value: IsCommonYesNoLabel[IsCommonYesNo.Yes] },
  { id: IsCommonYesNo.No, value: IsCommonYesNoLabel[IsCommonYesNo.No] },
])
const isMicroSchoolSaleVals = ref(null)

// 是否开启微校展示选项
const isMicroSchoolDisplayOptions = ref([
  { id: IsCommonYesNo.Yes, value: IsCommonYesNoLabel[IsCommonYesNo.Yes] },
  { id: IsCommonYesNo.No, value: IsCommonYesNoLabel[IsCommonYesNo.No] },
])
const isMicroSchoolDisplayVals = ref(null)

// 班级名称选项（多选）
const classNameOptions = ref([])
const classNameVals = ref([])
const classNamePagination = ref({
  current: 1,
  pageSize: 10,
  total: 0,
})
const classNameFinished = ref(false)

// 报读课程选项（多选）- 复用意向课程的数据源
const enrolledCourseVals = ref([])

// 有效期至（复用下次跟进时间组件）
const validityPeriodVals = ref([])

// 班主任（复用销售员组件）
const classTeacherOptions = ref([])
const classTeacherVals = ref(null)
const selectedClassTeacherInfo = ref(null)
const classTeacherPagination = ref({
  current: 1,
  pageSize: 100,
  total: 0,
})
const classTeacherFinished = ref(false)

// 是否欠费（复用是否有体验价组件）
const isArrearsOptions = ref([
  { id: IsCommonYesNo.Yes, value: '是' },
  { id: IsCommonYesNo.No, value: '否' },
])
const isArrearsVals = ref(null)
const orderArrearStatusOptions = ref([
  { id: 2, value: '有欠费未补费' },
  { id: 3, value: '有欠费已补费' },
  { id: 1, value: '无欠费' },
  { id: 4, value: '已坏账' },
  { id: 5, value: '补费完成' },
])
const orderArrearStatusVals = ref([])

// 最近上课时间（复用下次跟进时间组件）
const lastClassTimeVals = ref([])

// 创建时间选项
const createTimeVals = ref([])
const finishTimeVals = ref([])
const birthdayVals = ref([]) // 新增生日值
const approveNumberVals = ref('')
const orderNumberVals = ref('')
const wxChatVals = ref('') // 新增微信号值
const schoolVals = ref('') // 新增学校值
const addressVals = ref('') // 新增地址值
const hobbiesVals = ref('') // 新增爱好值
const lastEditedTimeVals = ref([])
// 分配时间选项
const assignTimeVals = ref([])
// 跟进时间选项
const followTimeVals = ref([])
// 最近跟进时间选项
const lastFollowTimeVals = ref([])
// 下次跟进时间选项
const nextFollowTimeVals = ref([])
// 申请时间选项
const applyTimeVals = ref([])
// 支付时间选项
const payTimeVals = ref([])
// 结课日期选项
const classEndingTimeVals = ref([])
// 停课日期选项
const classStopTimeVals = ref([])

// 课程列表选项
const courseListOptions = ref([])
const selectCourseValues = ref(null)
const selectedCourseInfo = ref(null) // 缓存已选课程的详细信息

// 一对一选项
const oneToOneOptions = ref([
  { id: 1, name: '张学良', course: '作业OT训练' },
  { id: 2, name: '高保庆', course: '交互训练' },
])
// 分页参数 - 原有的用于创建人
const pagination = ref({
  current: 1,
  pageSize: 100,
  total: 0,
  showTotal: total => `共 ${total} 条`,
})
const finished = ref(false)

// 推荐人独立的分页参数
const recommendPagination = ref({
  current: 1,
  pageSize: 100,
  total: 0,
  showTotal: total => `共 ${total} 条`,
})
const recommendFinished = ref(false)

// 课程列表独立的分页参数
const courseListPagination = ref({
  current: 1,
  pageSize: 10,
  total: 0,
  showTotal: total => `共 ${total} 条`,
})
const courseListFinished = ref(false)

// 学员/电话搜索独立的分页参数
const stuPhoneSearchPagination = ref({
  current: 1,
  pageSize: 100,
  total: 0,
  showTotal: total => `共 ${total} 条`,
})
const stuPhoneSearchFinished = ref(false)

// 销售员独立的分页参数
const salesPersonPagination = ref({
  current: 1,
  pageSize: 100,
  total: 0,
  showTotal: total => `共 ${total} 条`,
})
const salesPersonFinished = ref(false)

// 获取推荐人数据 - 独立的函数
async function getRecommendPage(params = { key: undefined, studentStatus: undefined }) {
  try {
    if (recommendFinished.value) return

    // 开启loading
    childRefs.value.recommend?.openSpinning()
    isLoading.value = true

    const res = await getRecommenderPageApi({
      'pageRequestModel': {
        'needTotal': true,
        'pageSize': recommendPagination.value.pageSize,
        'pageIndex': recommendPagination.value.current,
        'skipCount': 0,
      },
      'queryModel': {
        'searchKey': params.key,
        'studentStatus': params.studentStatus,
      },
      'sortModel': {},
    })

    if (res.code === 200) {
      let resultData = res.result || []

      // 保留首次加载的清空逻辑
      if (recommendPagination.value.current === 1) {
        recommendOptions.value = resultData
      }
      else {
        recommendOptions.value = [...recommendOptions.value, ...resultData]
      }
      recommendPagination.value.total = res.total || 0

      if (recommendOptions.value.length >= recommendPagination.value.total) {
        recommendFinished.value = true
      }
    }
  }
  catch (error) {
    console.error('加载推荐人数据失败:', error)
    // 发生错误时回退页码
    if (recommendPagination.value.current > 1) {
      recommendPagination.value.current -= 1
    }
  }
  finally {
    childRefs.value.recommend?.resetSpinning()
    isLoading.value = false
  }
}

// 获取课程列表数据 - 独立的函数
async function getCourseListPage(params = { key: undefined }) {
  try {
    if (courseListFinished.value) return

    // 开启loading - 同时支持意向课程和报读课程
    childRefs.value.yiXiangcourse?.openSpinning()
    childRefs.value.enrolledCourse?.openSpinning()
    isLoading.value = true

    const res = await getCourseIdAndNameApi({
      'pageRequestModel': {
        'needTotal': true,
        'pageSize': courseListPagination.value.pageSize,
        'pageIndex': courseListPagination.value.current,
        'skipCount': 0,
      },
      'queryModel': {
        'delFlag': false,
        'productType': 1,
        'searchKey': params.key,
        'saleStatus': true,
      },
      'sortModel': {},
    })

    if (res.code === 200) {
      let resultData = res.result || []

      // 转换数据格式，适配组件需要的格式
      const formattedData = resultData.map(item => ({
        id: item.id,
        value: item.name || '未命名课程',
        ...item // 保留原始数据，以备后用
      }))

      // 保留首次加载的清空逻辑
      if (courseListPagination.value.current === 1) {
        courseListOptions.value = formattedData
      }
      else {
        courseListOptions.value = [...courseListOptions.value, ...formattedData]
      }
      courseListPagination.value.total = res.total || 0

      if (courseListOptions.value.length >= courseListPagination.value.total) {
        courseListFinished.value = true
      }
    }
  }
  catch (error) {
    console.error('加载课程列表数据失败:', error)
    // 发生错误时回退页码
    if (courseListPagination.value.current > 1) {
      courseListPagination.value.current -= 1
    }
  }
  finally {
    childRefs.value.yiXiangcourse?.resetSpinning()
    childRefs.value.enrolledCourse?.resetSpinning()
    isLoading.value = false
  }
}

// 获取学员/电话搜索数据 - 独立的函数
async function getStuPhoneSearchPage(params = { key: undefined, studentStatus: undefined }) {
  try {
    if (stuPhoneSearchFinished.value) return

    // 开启loading
    childRefs.value.stuPhoneSearch?.openSpinning()
    spinning.value = true
    isLoading.value = true

    const res = await getRecommenderPageApi({
      'pageRequestModel': {
        'needTotal': true,
        'pageSize': stuPhoneSearchPagination.value.pageSize,
        'pageIndex': stuPhoneSearchPagination.value.current,
        'skipCount': 0,
      },
      'queryModel': {
        'searchKey': params.key,
        'studentStatus': params.studentStatus,
      },
      'sortModel': {},
    })

    if (res.code === 200) {
      let resultData = res.result || []

      // 保留首次加载的清空逻辑
      if (stuPhoneSearchPagination.value.current === 1) {
        stuPhoneSearchOptions.value = resultData
      }
      else {
        stuPhoneSearchOptions.value = [...stuPhoneSearchOptions.value, ...resultData]
      }
      stuPhoneSearchPagination.value.total = res.total || 0

      if (stuPhoneSearchOptions.value.length >= stuPhoneSearchPagination.value.total) {
        stuPhoneSearchFinished.value = true
      }
    }
  }
  catch (error) {
    console.error('加载学员/电话搜索数据失败:', error)
    // 发生错误时回退页码
    if (stuPhoneSearchPagination.value.current > 1) {
      stuPhoneSearchPagination.value.current -= 1
    }
  }
  finally {
    childRefs.value.stuPhoneSearch?.resetSpinning()
    spinning.value = false
    isLoading.value = false
  }
}

// 保留原有的getRecommenderPage函数用于向后兼容
async function getRecommenderPage(params = { key: undefined, studentStatus: undefined }) {
  // 这个函数现在主要用于向后兼容，实际应该使用具体的函数
  return await getStuPhoneSearchPage(params)
}

// 递归处理部门权限，如果用户有父级部门权限，则下级部门不禁用
function processDepartmentPermissions(departments, userDeptIds, rootDepartments = departments) {
  if (!Array.isArray(departments)) return []

  return departments.map(dept => {
    const processedDept = { ...dept }

    // 检查当前部门是否在用户的权限列表中
    const hasPermission = userDeptIds.includes(dept.id)

    // 如果当前部门有权限，则保持启用状态，不需要设置disabled
    if (hasPermission) {
      processedDept.disabled = false
      // console.log(`部门 "${dept.name}" (ID: ${dept.id}) - 直接有权限，启用`)
    } else {
      // 如果当前部门没有权限，需要检查父级是否有权限
      const hasParentAuth = hasParentPermission(dept, rootDepartments, userDeptIds)
      processedDept.disabled = !hasParentAuth
      // console.log(`部门 "${dept.name}" (ID: ${dept.id}) - 无直接权限，父级权限: ${hasParentAuth ? '有' : '无'}，${hasParentAuth ? '启用' : '禁用'}`)
    }

    // 递归处理子部门，传递原始的根部门数据
    if (dept.children && dept.children.length > 0) {
      processedDept.children = processDepartmentPermissions(dept.children, userDeptIds, rootDepartments)
    }

    return processedDept
  })
}

// 检查是否有父级权限
function hasParentPermission(currentDept, allDepts, userDeptIds) {
  // 递归查找当前部门的所有父级部门ID
  function findParentIds(deptId, departments, currentPath = []) {
    for (const dept of departments) {
      const newPath = [...currentPath, dept.id]

      if (dept.id === deptId) {
        // 找到目标部门，返回路径中除了自己以外的所有父级ID
        return currentPath
      }

      if (dept.children && dept.children.length > 0) {
        const result = findParentIds(deptId, dept.children, newPath)
        if (result !== null) {
          return result
        }
      }
    }
    return null
  }

  const parentIds = findParentIds(currentDept.id, allDepts)

  // 如果找到了父级部门，检查用户是否有任何父级部门的权限
  if (parentIds && parentIds.length > 0) {
    return parentIds.some(parentId => userDeptIds.includes(parentId))
  }

  return false
}

// 获取部门数据
async function getDepartmentList() {
  try {
    const res = await getListTreeDepartApi()
    if (res.code === 200 && res.result) {
      // 转换数据格式
      const transformedData = transformDepartmentData(res.result)
      originalDptListOptions.value = transformedData

      // console.log('用户部门权限IDs:', userDeptIds.value)
      // console.log('原始部门数据:', transformedData)

      // 根据用户权限处理部门数据
      dptListOptions.value = processDepartmentPermissions(transformedData, userDeptIds.value)

      // 部门树异步返回后再解析名称（避免已选条件里「所属部门」为空）
      nextTick(() => {
        dptName.value = getNameById(selectDptVals.value) ?? ''
      })
    }
  } catch (error) {
    console.error('获取部门数据失败:', error)
  }
}

// 转换部门数据格式
function transformDepartmentData(data) {
  if (!Array.isArray(data)) return []

  return data.map(item => {
    const transformed = {
      id: item.id || item.departId,
      name: item.departName || item.name,
      ...item
    }

    // 递归处理子部门
    if (item.children && Array.isArray(item.children) && item.children.length > 0) {
      transformed.children = transformDepartmentData(item.children)
    }

    return transformed
  })
}

// 获取全部创建人/员工
async function getCreateUserPage(params = { key: undefined }) {
  try {
    if (finished.value) return

    // 开启loading
    childRefs.value.createUser?.openSpinning()
    isLoading.value = true

    const res = await getUserListApi({
      'pageRequestModel': {
        'needTotal': true,
        'pageSize': pagination.value.pageSize,
        'pageIndex': pagination.value.current,
        'skipCount': 0,
      },
      'queryModel': {
        'searchKey': params.key,
      },
      'sortModel': {},
    })

    if (res.code === 200) {
      let resultData = res.result || []

      // 保留首次加载的清空逻辑
      if (pagination.value.current === 1) {
        createUserOptions.value = resultData
      }
      else {
        createUserOptions.value = [...createUserOptions.value, ...resultData]
      }
      pagination.value.total = res.total || 0

      if (createUserOptions.value.length >= pagination.value.total) {
        finished.value = true
      }
    }
  }
  catch (error) {
    console.error('加载创建人数据失败:', error)
    // 发生错误时回退页码
    if (pagination.value.current > 1) {
      pagination.value.current -= 1
    }
  }
  finally {
    childRefs.value.createUser?.resetSpinning()
    isLoading.value = false
  }
}

// 获取销售员数据 - 独立的函数
async function getSalesPersonPage(params = { key: undefined }) {
  try {
    if (salesPersonFinished.value) return

    // 开启loading
    childRefs.value.salesPerson?.openSpinning()
    isLoading.value = true

    const res = await getUserListApi({
      'pageRequestModel': {
        'needTotal': true,
        'pageSize': salesPersonPagination.value.pageSize,
        'pageIndex': salesPersonPagination.value.current,
        'skipCount': 0,
      },
      'queryModel': {
        'searchKey': params.key,
      },
      'sortModel': {},
    })

    if (res.code === 200) {
      let resultData = res.result || []

      // 保留首次加载的清空逻辑
      if (salesPersonPagination.value.current === 1) {
        salesPersonOptions.value = resultData
      }
      else {
        salesPersonOptions.value = [...salesPersonOptions.value, ...resultData]
      }
      salesPersonPagination.value.total = res.total || 0

      if (salesPersonOptions.value.length >= salesPersonPagination.value.total) {
        salesPersonFinished.value = true
      }
    }
  }
  catch (error) {
    console.error('加载销售员数据失败:', error)
    // 发生错误时回退页码
    if (salesPersonPagination.value.current > 1) {
      salesPersonPagination.value.current -= 1
    }
  }
  finally {
    childRefs.value.salesPerson?.resetSpinning()
    isLoading.value = false
  }
}

// 推荐人选项 - 独立的数据源和状态
const recommendOptions = ref([])
const recommendVals = ref(null)
const selectedRecommendInfo = ref(null) // 缓存已选推荐人的详细信息

// 学员/电话搜索选项 - 独立的数据源和状态  
const stuPhoneSearchOptions = ref([])
const stuPhoneSearchVals = ref(null)
const selectedStuPhoneSearchInfo = ref(null) // 缓存已选学员/电话的详细信息

// 保留原始的stuListOptions用于向后兼容(如果有其他地方使用)
const stuListOptions = ref([])

// 部门列表选项
const dptListOptions = ref([])
const originalDptListOptions = ref([]) // 保存原始部门数据
const selectDptVals = ref(null)

// 渠道状态列表选项
const channelListOptions = ref([
  { id: false, value: '启用中' },
  { id: true, value: '已停用' },
])
const selectChannelVals = ref([])

// 渠道类型列表选项
const channelTypeOptions = ref([
  { id: false, value: '自定义' },
  { id: true, value: '系统默认' },
])
const selectChannelTypeVals = ref([])

// 渠道树列表选项
const channelOptions = ref([])
const selectChannelTreeVals = ref([])
const selectChannelValsFilter = ref([])

// 渠道分类列表选项
const channelCategoryOptions = computed(() => {
  const defaultOption = { id: 0, value: '无分类' }

  if (props.channelCategories && props.channelCategories.length > 0) {
    return [
      defaultOption,
      ...props.channelCategories.map(item => ({
        id: item.id,
        value: item.categoryName,
      })),
    ]
  }

  return [defaultOption]
})

// 任职角色列表选项
const positionRoleOptions = computed(() => {
  return props.channelPositionRole.map((item) => {
    return {
      id: item.id,
      value: item.roleName,
    }
  })
})
const selectChannelCategoryVals = ref([])
const positionRoleVals = ref([])

// 科目列表选项
const subjectOptions = ref([
  { id: 1, value: '残联' },
  { id: 2, value: '自费' },
  { id: 3, value: '民生' },
])
const selectSubjectVals = ref(null)

// 课程类别列表选项
const courseCategoryOptions = ref([])
const selectCourseCategoryVals = ref(null)

// 学员状态选项
const studentStatusOptions = ref([
  { id: 1, value: '在读学员' },
  { id: 2, value: '历史学员' },
])
const selectStudentStatusVals = ref(1) // 默认选中在读学员
// 账号状态
const accountStatusOptions = ref([
  { id: '0', value: '在职中' },
  { id: '1', value: '已离职' },
])
const selectAccountStatusVals = ref(null)
// 员工类型
const userTypeOptions = ref([
  { id: 1, value: '正式员工' },
  { id: 2, value: '兼职员工' },
])
const selectUserTypeVals = ref(null)

// 业绩分配状态选项
const performanceAllocationStatusOptions = ref([
  { id: 1, value: '待分配' },
  { id: 2, value: '已分配' },
  { id: 3, value: '无需分配' },
])
const performanceAllocationStatusVals = ref(null)

// 订单类型选项
const orderTypeOptions = ref([
  { id: 1, value: '报名续费' },
  { id: 2, value: '储值账户充值' },
  { id: 3, value: '退课' },
  { id: 4, value: '储值账户退费' },
  { id: 5, value: '转课' },
  { id: 6, value: '退教材费' },
  { id: 7, value: '退学杂费' },
])
const orderTypeVals = ref([])
const orderTagOptions = ref([])
const orderTagVals = ref([])
const enrollTypeOptions = ref([
  { id: 1, value: '新报' },
  { id: 3, value: '扩科' },
  { id: 2, value: '续费' },
  { id: 0, value: '无' },
])
const enrollTypeVals = ref([])
const productTypeOptions = ref([
  { id: 1, value: '课程' },
  { id: 2, value: '教学用品' },
  { id: 3, value: '约课付费' },
  { id: 4, value: '储值账户' },
  { id: 5, value: '场地预约' },
  { id: 6, value: '学杂费' },
])
const productTypeVals = ref([])
const orderStatusOptions = ref([
  { id: 1, value: '待付款' },
  { id: 3, value: '已完成' },
  { id: 7, value: '退费中' },
  { id: 8, value: '已退费' },
  { id: 4, value: '已关闭' },
  { id: 5, value: '已作废' },
  { id: 6, value: '待处理' },
  { id: 2, value: '审批中' },
])
const orderStatusVals = ref([])
const approvalStatusOptions = ref([
  { id: 1, value: '审批中' },
  { id: 2, value: '审批通过' },
  { id: 3, value: '审批拒绝' },
  { id: 4, value: '已作废' },
])
const approvalStatusVals = ref([])
const enableStatusOptions = ref([
  { id: 1, value: '启用' },
  { id: 0, value: '停用' },
])
const enableStatusVals = ref(null)
const orderSourceOptions = ref([
  { id: 1, value: '线下办理' },
  { id: 2, value: '微校报名' },
  { id: 3, value: '线下导入' },
  { id: 4, value: '续费订单' },
])
const orderSourceVals = ref([])
const dealDateVals = ref([])
const latestPaidTimeVals = ref([])
// 当前状态选项
const currentStatusOptions = ref([
  { id: 3, value: '已结课' },
  { id: 2, value: '已停课' },
  { id: 1, value: '正常' },
])
const selectCurrentStatusVals = ref([])

// 是否分班
const orNotFenClassOptions = ref([
  { id: 0, value: '未分班' },
  { id: 1, value: '已分班' },
])
const selectOrNotFenClassVals = ref([])

// 收费方式选项
const billingModeOptions = ref([
  { id: 1, value: '按课时' },
  { id: 2, value: '按时段' },
  { id: 3, value: '按金额' },
])
const selectBillingModeVals = ref([])

// 是否设置有效期
const isSetExpirationDateOptions = ref([
  { id: 1, value: '已设置有效期' },
  { id: 2, value: '未设置有效期' },
])
const selectIsSetExpirationDateVals = ref(null)

// 开班状态选项
const openClassStatusOptions = ref([
  { id: 1, value: '开班中' },
  { id: 2, value: '未开班' },
])
const selectOpenClassStatusVals = ref(null)

// 是否排课选项
const doYouScheduleOptions = ref([
  { id: 1, value: '已排课' },
  { id: 2, value: '未排课' },
])
const selectDoYouScheduleVals = ref(null)

// 是否被推荐
const recommendedOptions = ref([
  { id: IsCommonYesNo.Yes, value: IsCommonYesNoLabel[IsCommonYesNo.Yes] },
  { id: IsCommonYesNo.No, value: IsCommonYesNoLabel[IsCommonYesNo.No] },
])
const recommendedVals = ref([])

// 自定义的文本输入框值 - 修改为对象，以存储不同字段的值
const searchInputVals = ref({})

// 剩余数量相关变量
const remainingModeOptions = [
  { key: 'lesson', label: '按课时' },
  { key: 'period', label: '按时段' },
  { key: 'amount', label: '按金额' },
]
const remainingMode = ref('lesson') // 默认按课时
const minRemaining = ref(null)
const maxRemaining = ref(null)
const selectedRemainingRange = ref(null)
// 快捷筛选选项（单选）
const quickFilters = ref([
  { id: 1, name: '今日待跟进', count: 1, selected: false },
  { id: 2, name: '本周新增', count: 0, selected: false },
  { id: 3, name: '逾期未回访', count: 0, selected: false },
])

// 过滤要显示的快捷筛选项
const visibleQuickFilters = computed(() => {
  return quickFilters.value.filter(filter => !props.hideQuickFilters.includes(filter.id))
})

// 监听跟进数量
watch(() => props.followUpCount, (newVal) => {
  quickFilters.value[0].count = newVal.toBeFollowedUpTodayCount || 0
  quickFilters.value[1].count = newVal.newInquiriesAddedWeekCount || 0
  quickFilters.value[2].count = newVal.overdueForFollowUpInterviewCount || 0
}, { immediate: true })
const quickOneToOneFilters = ref([
  { id: 1, name: '未分配班主任学员', count: 1, selected: false },
  { id: 2, name: '待排课学员', count: 0, selected: false },
])
const quickApproveFilters = ref([
  { id: 1, name: '待我审批', count: 1, selected: false },
  { id: 2, name: '我已审批', count: 10, selected: false },
])
watch(() => props.approveQuickCounts, (newVal) => {
  quickApproveFilters.value[0].count = newVal.truntoMyApproveCount || 0
  quickApproveFilters.value[1].count = newVal.myHaveApprovedCount || 0
}, { immediate: true, deep: true })
// 体验课购买状态选项
const trialPurchaseStatusOptions = ref([
  { id: 1, value: '已购买' },
  { id: 2, value: '未购买' },
])
const trialPurchaseStatusVals = ref([])

// 处理快捷筛选单选
function selectQuickFilter(selectedFilter, type) {
  if (type == 1) {
    quickFilters.value.forEach((filter) => {
      filter.selected
        = filter.id === selectedFilter.id ? !filter.selected : false
    })
    const selectedQuickFilter = quickFilters.value.find(q => q.selected)
    let quickFilter = selectedQuickFilter?.id

    // 如果有自定义传值，使用自定义值
    if (selectedQuickFilter && props.customQuickFilterValues[selectedQuickFilter.id] !== undefined) {
      quickFilter = props.customQuickFilterValues[selectedQuickFilter.id]
    }

    emit('update:quickFilter', quickFilter)
  }
  else if (type == 2) {
    quickOneToOneFilters.value.forEach((filter) => {
      filter.selected
        = filter.id === selectedFilter.id ? !filter.selected : false
    })
    console.log(
      '当前快捷筛选:',
      quickOneToOneFilters.value.find(q => q.selected)?.name,
    )
  }
  else if (type == 3) {
    quickApproveFilters.value.forEach((filter) => {
      filter.selected
        = filter.id === selectedFilter.id ? !filter.selected : false
    })
    const selectedQuickFilter = quickApproveFilters.value.find(q => q.selected)
    emit('update:quickFilter', selectedQuickFilter?.id)
  }
}
function handleNotFollowDaysInput(val) {
  if (val) {
    notFollowDaysQuick.value = null
  }
}
function selectNotFollowDaysQuick(d) {
  notFollowDaysQuick.value = d
  notFollowDaysInput.value = d // 让输入框也同步显示
}

function resetNotFollowDays() {
  notFollowDaysInput.value = null
  // 保留快捷选择
  notFollowDaysQuick.value = null
  // notFollowDaysSelected.value = null; // 如果需要保留已选条件，也不要清空
}
function handleNotFollowDaysConfirm() {
  notFollowDaysSelected.value = notFollowDaysInput.value || notFollowDaysQuick.value
  // 关闭下拉
  if (childRefs.value.notFollowDays && childRefs.value.notFollowDays.closeDropdown) {
    childRefs.value.notFollowDays.closeDropdown()
  }
  // 这里可以触发实际的筛选逻辑
  console.log(notFollowDaysSelected.value)
  emit('update:notFollowDaysFilter', notFollowDaysSelected.value)
}

function resetAge() {
  minAge.value = null
  maxAge.value = null
  // selectedAgeRange.value = null;
}

function handleAgeConfirm() {
  if (minAge.value == null && maxAge.value == null) {
    selectedAgeRange.value = null // 清空已选条件
    // 关闭下拉（如果有ref，类似notFollowDays的处理）
    if (childRefs.value.age && childRefs.value.age.closeDropdown) {
      childRefs.value.age.closeDropdown()
    }
    emit('update:ageFilter', [])
    return
  }
  let min = minAge.value
  let max = maxAge.value
  if (min != null && max != null && min > max) {
    // 交换
    [min, max] = [max, min]
    minAge.value = min
    maxAge.value = max
  }
  selectedAgeRange.value = { min, max }
  // 关闭下拉（如果有ref，类似notFollowDays的处理）
  if (childRefs.value.age && childRefs.value.age.closeDropdown) {
    childRefs.value.age.closeDropdown()
  }
  // 这里可以触发实际的筛选逻辑
  console.log(minAge.value, maxAge.value)
  emit('update:ageFilter', [min, max])
}

function handleGradeChange() {
  nextTick(() => {
    debouncedEmit('gradeFilter', gradeVals.value)
  })
}

function handleIntentionChange() {
  nextTick(() => {
    debouncedEmit('intentionLevelFilter', selectedValues.value)
  })
}

function handleChannelTypeChange() {
  nextTick(() => {
    let isDefault
    if (selectChannelTypeVals.value.length === 1) {
      isDefault = selectChannelTypeVals.value[0]
    }
    else if (selectChannelTypeVals.value.length === 2) {
      isDefault = undefined
    }
    else {
      isDefault = undefined
    }
    debouncedEmit('channelTypeFilter', isDefault)
  })
}

function handleChannelChange() {
  nextTick(() => {
    let isDisabled
    if (selectChannelVals.value.length === 1) {
      isDisabled = selectChannelVals.value[0]
    }
    else if (selectChannelVals.value.length === 2) {
      isDisabled = undefined
    }
    else {
      isDisabled = undefined
    }
    debouncedEmit('channelStatusFilter', isDisabled)
  })
}

function handleFollowChange() {
  nextTick(() => {
    // console.log("跟进状态:", toRaw(followStatusVals.value));
    debouncedEmit('followStatusFilter', followStatusVals.value)
  })
}

function handleSexChange() {
  nextTick(() => {
    debouncedEmit('sexFilter', sexVals.value)
  })
}

function handleIsCommonCourseChange() {
  nextTick(() => {
    console.log(`是否是通用课：${isCommonCourseVals.value}`)
    debouncedEmit('isCommonCourseFilter', isCommonCourseVals.value)
  })
}

function handleFollowMethodChange() {
  nextTick(() => {
    debouncedEmit('followMethodFilter', followMethodVals.value)
  })
}
function handleVisitStatusChange() {
  nextTick(() => {
    debouncedEmit('visitStatusFilter', visitStatusVals.value)
  })
}
function handlePayrollStatusChange() {
  nextTick(() => {
    console.log('工资条状态:', payrollStatusVals.value)
    debouncedEmit('payrollStatusFilter', payrollStatusVals.value)
  })
}
function handleStuStatusChange() {
  nextTick(() => {
    debouncedEmit('stuStatusFilter', stuStatusVals.value)
  })
}
function handleCreateUserChange(e) {
  // 缓存已选创建人的详细信息
  if (e) {
    const selectedUser = createUserOptions.value.find(user => user.id === e)
    selectedCreateUserInfo.value = selectedUser || null
  } else {
    selectedCreateUserInfo.value = null
  }

  nextTick(() => {
    console.log('创建人:', e)
    emit('update:createUserFilter', e)
  })
}
function handleTeachingMethodChange(e) {
  nextTick(() => {
    console.log('授课方式:', e)
    emit('update:teachingMethodFilter', e)
  })
}

function handleSellStatusChange(e) {
  nextTick(() => {
    console.log('售卖状态:', e)
    emit('update:sellStatusFilter', e)
  })
}
function handleSalesPersonChange(e) {
  // 缓存已选销售员的详细信息
  if (e) {
    const selectedUser = salesPersonOptions.value.find(user => user.id === e)
    selectedSalesPersonInfo.value = selectedUser || null
  } else {
    selectedSalesPersonInfo.value = null
  }

  nextTick(() => {
    console.log('销售员:', e)
    emit('update:salesPersonFilter', e)
  })
}
function handleHasSalesPersonChange() {
  nextTick(() => {
    console.log('是否有销售员:', toRaw(hasSalesPersonVals.value))
    // 处理下 如果全部选择 则传undefined 如果选择否 则传false 如果选择是 则传true
    let isDisabled
    if (hasSalesPersonVals.value.length === 1) {
      isDisabled = hasSalesPersonVals.value[0] !== 0
    }
    else if (hasSalesPersonVals.value.length === 2) {
      isDisabled = undefined
    }
    else {
      isDisabled = undefined
    }
    debouncedEmit('hasSalesPersonFilter', isDisabled)
  })
}
function handleCreateTimeChange(e) {
  nextTick(() => {
    // console.log("创建时间2:", e);
    emit('update:createTimeFilter', e)
  })
}

function handlelastEditedTimeChange(e) {
  nextTick(() => {
    emit('update:lastEditedTimeFilter', e)
  })
}

function handleFinishTimeChange(e) {
  nextTick(() => {
    emit('update:finishTimeFilter', e)
  })
}

function handleBirthdayChange(e) {
  nextTick(() => {
    // console.log("生日:", e);
    emit('update:birthdayFilter', e)
  })
}

function handleFollowTimeChange(e) {
  nextTick(() => {
    console.log('跟进时间:', e)
    emit('update:followTimeFilter', e)
  })
}

function handleTrialPurchaseStatusChange() {
  nextTick(() => {
    emit('update:trialPurchaseStatusFilter', toRaw(trialPurchaseStatusVals.value))
  })
}
function handleAssignTimeChange() {
  nextTick(() => {
    // console.log("分配时间:", assignTimeVals.value);
    emit('update:assignTimeFilter', assignTimeVals.value)
  })
}
function handleWxChatChange() {
  nextTick(() => {
    console.log('微信号:', wxChatVals.value)
    if (childRefs.value.wxChat && childRefs.value.wxChat.closeDropdown) {
      childRefs.value.wxChat.closeDropdown()
    }
    emit('update:wxChatFilter', wxChatVals.value)
  })
}
function handleSchoolChange() {
  nextTick(() => {
    console.log('就读学校:', schoolVals.value)
    if (childRefs.value.school && childRefs.value.school.closeDropdown) {
      childRefs.value.school.closeDropdown()
    }
    emit('update:schoolFilter', schoolVals.value)
  })
}

function handleOrderNumberChange() {
  nextTick(() => {
    if (childRefs.value.orderNumber && childRefs.value.orderNumber.closeDropdown) {
      childRefs.value.orderNumber.closeDropdown()
    }
    emit('update:orderNumberFilter', orderNumberVals.value)
  })
}

function handleApproveNumberChange() {
  nextTick(() => {
    if (childRefs.value.approveNumber && childRefs.value.approveNumber.closeDropdown) {
      childRefs.value.approveNumber.closeDropdown()
    }
    emit('update:approveNumberFilter', approveNumberVals.value)
  })
}

function handleAddressChange() {
  nextTick(() => {
    console.log('家庭住址:', addressVals.value)
    if (childRefs.value.address && childRefs.value.address.closeDropdown) {
      childRefs.value.address.closeDropdown()
    }
    emit('update:addressFilter', addressVals.value)
  })
}

function handleHobbiesChange() {
  nextTick(() => {
    console.log('兴趣爱好:', hobbiesVals.value)
    if (childRefs.value.hobbies && childRefs.value.hobbies.closeDropdown) {
      childRefs.value.hobbies.closeDropdown()
    }
    emit('update:hobbiesFilter', hobbiesVals.value)
  })
}
function handleLastFollowTimeChange() {
  nextTick(() => {
    // console.log("最近跟进时间:", lastFollowTimeVals.value);
    emit('update:lastFollowTimeFilter', lastFollowTimeVals.value)
  })
}
function handleNextFollowTimeChange() {
  nextTick(() => {
    // console.log("下次跟进时间:", nextFollowTimeVals.value);
    emit('update:nextFollowTimeFilter', nextFollowTimeVals.value)
  })
}
function handleCourseChange(e) {
  // 缓存已选课程的详细信息
  if (e) {
    const selectedCourse = courseListOptions.value.find(course => course.id === e)
    selectedCourseInfo.value = selectedCourse || null
  } else {
    selectedCourseInfo.value = null
  }

  nextTick(() => {
    console.log('意向课程:', e)
    emit('update:intentionCourseFilter', e)
  })
}
function handleTuiJianUserChange() {
  // 缓存已选推荐人的详细信息
  if (recommendVals.value) {
    const selectedUser = recommendOptions.value.find(user => user.id === recommendVals.value)
    selectedRecommendInfo.value = selectedUser || null
  } else {
    selectedRecommendInfo.value = null
  }

  nextTick(() => {
    console.log('推荐人:', recommendVals.value)
    emit('update:tuiJianUserFilter', recommendVals.value)
  })
}

// 推荐人下拉框可见状态变化 - 独立的函数
async function onRecommendDropdownVisibleChange(value) {
  if (value)
    return
  recommendOptions.value = []
  recommendPagination.value.current = 1
  recommendFinished.value = false
  await getRecommendPage()
}

// 推荐人加载更多 - 独立的函数
async function loadMoreRecommend() {
  // 检查是否正在加载且还有更多数据
  if (!isLoading.value && recommendPagination.value.current * recommendPagination.value.pageSize < recommendPagination.value.total) {
    isLoading.value = true
    recommendPagination.value.current += 1
    await getRecommendPage()
  }
}

// 课程列表下拉框可见状态变化 - 独立的函数
async function onCourseListDropdownVisibleChange(value) {
  if (value) {
    // 打开下拉框时，如果数据为空则加载
    if (courseListOptions.value.length === 0) {
      await getCourseListPage()
    }
    return
  }
  // 关闭下拉框时重置数据
  courseListOptions.value = []
  courseListPagination.value.current = 1
  courseListFinished.value = false
  await getCourseListPage()
}

// 课程列表加载更多 - 独立的函数
async function loadMoreCourseList() {
  // 检查是否正在加载且还有更多数据
  if (!isLoading.value && courseListPagination.value.current * courseListPagination.value.pageSize < courseListPagination.value.total) {
    isLoading.value = true
    courseListPagination.value.current += 1
    await getCourseListPage()
  }
}

// 学员/电话搜索下拉框可见状态变化 - 独立的函数
async function onStuPhoneSearchDropdownVisibleChange(value) {
  if (value)
    return
  stuPhoneSearchOptions.value = []
  stuPhoneSearchPagination.value.current = 1
  stuPhoneSearchFinished.value = false
  await getStuPhoneSearchPage()
}

// 学员/电话搜索加载更多 - 独立的函数
async function loadMoreStuPhoneSearch() {
  // 检查是否正在加载且还有更多数据
  if (!isLoading.value && stuPhoneSearchPagination.value.current * stuPhoneSearchPagination.value.pageSize < stuPhoneSearchPagination.value.total) {
    isLoading.value = true
    stuPhoneSearchPagination.value.current += 1
    await getStuPhoneSearchPage()
  }
}

// 保留原有的函数用于向后兼容
async function onDropdownVisibleChange(value) {
  return await onStuPhoneSearchDropdownVisibleChange(value)
}
async function loadMore() {
  return await loadMoreStuPhoneSearch()
}
// 获取全部创建人/员工
async function onDropdownVisibleChangeTeacher(value) {
  createUserOptions.value = []
  pagination.value.current = 1
  finished.value = false
  await getCreateUserPage({ key: value })
}

// 销售员下拉框可见状态变化 - 独立的函数
async function onDropdownVisibleChangeSalesPerson(value) {
  salesPersonOptions.value = []
  salesPersonPagination.value.current = 1
  salesPersonFinished.value = false
  await getSalesPersonPage({ key: value })
}

// 创建人搜索
async function onSearchCreateUserFun(value) {
  pagination.value.current = 1
  finished.value = false
  await getCreateUserPage({ key: value })
}

// 销售员搜索 - 独立的函数
async function onSearchSalesPersonFun(value) {
  salesPersonPagination.value.current = 1
  salesPersonFinished.value = false
  await getSalesPersonPage({ key: value })
}

// 创建人加载更多
async function loadMoreCreateUser() {
  // 检查是否正在加载且还有更多数据
  if (!isLoading.value && pagination.value.current * pagination.value.pageSize < pagination.value.total) {
    isLoading.value = true
    pagination.value.current += 1
    await getCreateUserPage()
  }
}

// 销售员加载更多 - 独立的函数
async function loadMoreSalesPerson() {
  // 检查是否正在加载且还有更多数据
  if (!isLoading.value && salesPersonPagination.value.current * salesPersonPagination.value.pageSize < salesPersonPagination.value.total) {
    isLoading.value = true
    salesPersonPagination.value.current += 1
    await getSalesPersonPage()
  }
}

// 推荐人搜索 - 独立的函数
async function onSearchRecommendFun(value) {
  recommendPagination.value.current = 1
  recommendFinished.value = false
  await getRecommendPage({ key: value, studentStatus: 0 })
}

// 课程列表搜索 - 独立的函数
async function onSearchCourseListFun(value) {
  courseListPagination.value.current = 1
  courseListFinished.value = false
  await getCourseListPage({ key: value })
}

// 学员/电话搜索搜索 - 独立的函数
async function onSearchStuPhoneSearchFun(value) {
  stuPhoneSearchPagination.value.current = 1
  stuPhoneSearchFinished.value = false
  await getStuPhoneSearchPage({ key: value, studentStatus: 0 })
}

// 保留原有的搜索函数用于向后兼容
async function onSearchFun(value) {
  return await onSearchStuPhoneSearchFun(value)
}
async function onSearchCourseCategoryFun(value) {
  courseCategoryPagination.value.current = 1
  courseCategoryFinished.value = false
  await queryCourseCategory(value)
}
// 分页参数
const courseCategoryPagination = ref({
  current: 1,
  pageSize: 5,
  total: 0,
  showTotal: total => `共 ${total} 条`,
})
const courseCategoryFinished = ref(false)
// 查询课程类别
async function queryCourseCategory(value) {
  // console.log("查询课程类别:", value);
  if (courseCategoryFinished.value)
    return
  try {
    const res = await getCourseCategoryPageApi({
      'pageRequestModel': {
        'needTotal': true,
        'pageSize': courseCategoryPagination.value.pageSize, // 获取所有数据
        'pageIndex': courseCategoryPagination.value.current,
        'skipCount': 1,
      },
      'queryModel': {
        'searchKey': value, // 使用单独的搜索关键字
      },
      'sortModel': {
        'orderBySortNumber': 2,
      },
    })
    if (res.code === 200) {
      courseCategoryOptions.value = res.result
      courseCategoryPagination.value.total = res.total
      childRefs.value.courseType?.resetSpinning()
    }
  }
  catch (error) {
    console.log(error)
  }
}
const courseAttributeOptions = ref([])
const courseAttributeId = ref(null)
const courseAttributeSearchValue = ref(null)
// 查询课程属性
async function handleCourseAttributeDropdownVisibleChange(val, id) {
  // console.log('课程属性下拉框展开:', id);
  // 下拉框展开时 的 id
  courseAttributeId.value = id
  courseAttributeSearchValue.value = val
  courseAttributeOptions.value = []
  childRefs.value[`courseAttribute_${id}`]?.openSpinning()
  const res = await getCoursePropertyOptionsApi({
    'propertyId': id,
  })
  if (res.code === 200) {
    // console.log(res.result);
    courseAttributeOptions.value = res.result
    childRefs.value[`courseAttribute_${id}`]?.resetSpinning()
  }
}
// 搜索课程属性
async function filterCourseAttributeFun(val) {
  // 如果没有搜索值，显示所有选项
  if (!val || val.trim() === '') {
    // 重新获取完整的课程属性选项
    const res = await getCoursePropertyOptionsApi({
      'propertyId': courseAttributeId.value,
    })
    if (res.code === 200) {
      courseAttributeOptions.value = res.result
    }
  }
  else {
    // 使用filter筛选匹配的选项
    const filteredOptions = courseAttributeOptions.value.filter(option =>
      option.name && option.name.toLowerCase().includes(val.toLowerCase()),
    )

    // 更新筛选后的选项
    courseAttributeOptions.value = filteredOptions

    // 如果筛选结果为空，可以考虑重新请求接口
    if (filteredOptions.length === 0) {
      console.log('没有匹配的选项，可能需要从服务器搜索')
      // 可以在这里调用接口进行服务器端搜索
      courseAttributeOptions.value = []
    }
  }

  childRefs.value[`courseAttribute_${courseAttributeId.value}`]?.resetSpinning()
}
// 自定义搜索课程属性
function handleCourseAttributeChange(e) {
  nextTick(() => {
    const item = props.courseAttributeList?.find(attr => attr.id === courseAttributeId.value || attr.id === e)
    const optionId = typeof e === 'object' && e !== null ? e.id : e
    emit('update:courseAttributeFilter', {
      itemId: item?.id || courseAttributeId.value,
      itemName: item?.name,
      value: optionId,
    }, false, item?.id || courseAttributeId.value, `courseAttribute_${item?.id || courseAttributeId.value}`)
  })
}
// 渠道树下拉框可见状态 请求数据
async function getChannelTree() {
  try {
    const res = await getChannelTreeApi()
    // 过滤 channelList长度为0的数据
    channelOptions.value = res.result.filter(item => item.channelList.length > 0)
  }
  catch (error) {
    console.log(error)
  }
}
function handleDropdownVisibleChange(visible) {
  getChannelTree()
}
function handleChannelCategoryChange() {
  nextTick(() => {
    // 获取原始数据
    const originalData = toRaw(selectChannelTreeVals.value)
    // console.log("渠道原始数据:", originalData);

    // 提取叶子节点值
    const leafValues = originalData.map(subArr => subArr[subArr.length - 1])

    // console.log("渠道处理后:", leafValues);

    // 更新过滤后的值
    selectChannelValsFilter.value = leafValues

    debouncedEmit('channelFilter', selectChannelValsFilter.value)
  })
}

function handleSubjectChange(e) {
  nextTick(() => {
    console.log('科目:', e)
  })
}
function handleCourseCategoryChange(e) {
  nextTick(() => {
    console.log('课程类别:', e)
    emit('update:courseCategoryFilter', e)
  })
}
function handleStudentStatusChange(e) {
  nextTick(() => {
    console.log('学员状态:', e)
    // 发出学员状态筛选事件，传递选中的值
    emit('update:stuStatusFilter', e)
  })
}
function handleCurrentStatusChange(e) {
  nextTick(() => {
    console.log('当前状态:', e)
    debouncedEmit('currentStatusFilter', selectCurrentStatusVals.value)
  })
}

function handleAccountStatusChange(e) {
  nextTick(() => {
    emit('update:channelAccountStatus', e)
    console.log('账号状态:', e)
  })
}

function handleUserTypeChange(e) {
  nextTick(() => {
    emit('update:channelUserType', e)
    console.log('用户类型:', e)
  })
}

function handleOrNotFenClassChange(e) {
  nextTick(() => {
    console.log('是否分班:', e)
    debouncedEmit('orNotFenClassFilter', selectOrNotFenClassVals.value)
  })
}

function handleBillingModeChange(e) {
  nextTick(() => {
    console.log('收费方式:', selectBillingModeVals.value)
    debouncedEmit('chargingMethodFilter', selectBillingModeVals.value)
  })
}
function handleIsSetExpirationDateChange(e) {
  nextTick(() => {
    console.log('有效期状态:', e)
    debouncedEmit('isSetExpirationDateFilter', selectIsSetExpirationDateVals.value)
  })
}
function handleOpenClassStatusChange(e) {
  nextTick(() => {
    console.log('开班状态:', e)
  })
}
function handleDoYouScheduleChange(e) {
  nextTick(() => {
    console.log('是否排课:', e)
  })
}
function handleRecommendedChange() {
  nextTick(() => {
    console.log('是否被推荐:', toRaw(recommendedVals.value))
    // 处理下 如果全部选择 则传undefined 如果选择否 则传false 如果选择是 则传true
    let isDisabled
    if (recommendedVals.value.length === 1) {
      isDisabled = recommendedVals.value[0] !== 0
    }
    else if (recommendedVals.value.length === 2) {
      isDisabled = undefined
    }
    else {
      isDisabled = undefined
    }
    debouncedEmit('recommendedFilter', isDisabled)
  })
}
// 渠道分类处理函数
function handleChannelCategoryValueChange() {
  nextTick(() => {
    console.log('渠道分类:', toRaw(selectChannelCategoryVals.value))

    // 根据选择确定categoryId参数
    let categoryId
    if (selectChannelCategoryVals.value.length === 1) {
      // 只选了一个选项时，直接使用该选项的id值
      categoryId = selectChannelCategoryVals.value
    }
    else if (selectChannelCategoryVals.value.length > 1) {
      // 选了多个选项时，传数组
      categoryId = selectChannelCategoryVals.value
    }
    else {
      // 没有选择任何选项
      categoryId = undefined
    }

    // 发出事件通知父组件
    // emit('update:channelCategoryFilter', categoryId);
    debouncedEmit('channelCategoryFilter', categoryId)
  })
}
// 任职角色处理函数
function handleRoleValueChange() {
  nextTick(() => {
    // 根据选择确定Id参数
    console.log('任职角色:', positionRoleVals.value)
    let id
    if (positionRoleVals.value.length === 1) {
      // 只选了一个选项时，直接使用该选项的id值
      id = positionRoleVals.value
    }
    else if (positionRoleVals.value.length > 1) {
      // 选了多个选项时，传数组
      id = positionRoleVals.value
    }
    else {
      // 没有选择任何选项
      id = undefined
    }

    // 发出事件通知父组件
    // emit('update:channelCategoryFilter', id);
    debouncedEmit('channelPositionRoleFilter', id)
  })
}

function resetRemaining() {
  minRemaining.value = null
  maxRemaining.value = null
}
function handleRemainingConfirm() {
  if (minRemaining.value == null && maxRemaining.value == null) {
    selectedRemainingRange.value = null // 清空已选条件
    // 关闭下拉
    if (childRefs.value.remaining && childRefs.value.remaining.closeDropdown) {
      childRefs.value.remaining.closeDropdown()
    }
    // emit清空事件
    debouncedEmit('remainingFilter', null)
    return
  }
  let min = minRemaining.value
  let max = maxRemaining.value
  if (min != null && max != null && min > max) {
    // 交换
    [min, max] = [max, min]
    minRemaining.value = min
    maxRemaining.value = max
  }
  
  // 将mode转换为对应的数字：lesson=1, period=2, amount=3
  const modeMap = { lesson: 1, period: 2, amount: 3 }
  const modeValue = modeMap[remainingMode.value] || 1
  
  selectedRemainingRange.value = { min, max, mode: remainingMode.value }

  // emit事件，传递给父组件
  debouncedEmit('remainingFilter', { mode: modeValue, min, max })

  // 清空输入框但保留已选条件
  minRemaining.value = null
  maxRemaining.value = null

  // 关闭下拉
  if (childRefs.value.remaining && childRefs.value.remaining.closeDropdown) {
    childRefs.value.remaining.closeDropdown()
  }
}
// 已选条件计算
const selectedConditions = computed(() => {
  const conditions = [
    {
      type: 'quick',
      label: '快捷筛选',
      show: true,
      values: quickFilters.value
        .filter(q => q.selected)
        .map(q => ({ id: q.id, value: q.name })),
    },
    {
      type: 'quickOneToOne',
      label: '快捷筛选',
      show: true,
      values: quickOneToOneFilters.value
        .filter(q => q.selected)
        .map(q => ({ id: q.id, value: q.name })),
    },
    {
      type: 'quickApprove',
      label: '快捷筛选',
      show: true,
      values: quickApproveFilters.value
        .filter(q => q.selected)
        .map(q => ({ id: q.id, value: q.name })),
    },
    {
      type: 'teacherSelect',
      label: teacherType.value == 1 ? '上课老师' : '上课助教',
      show: true,
      values: (() => {
        if (!selectTeacher.value)
          return []
        const item = createUserOptions.value.find(
          opt => opt.id === selectTeacher.value,
        )
        return item ? [{ id: item.id, value: `${item.value}` }] : []
      })(),
    },
    {
      type: 'oneToOneSearch',
      label: '一对一',
      show: true,
      values: (() => {
        if (!searchKeyOneToOne.value)
          return []
        const item = oneToOneOptions.value.find(
          opt => opt.id === searchKeyOneToOne.value,
        )
        return item
          ? [{ id: item.id, value: `${item.name}～${item.course}` }]
          : []
      })(),
    },
    {
      type: 'stuPhoneSearchNew',
      label: '学员/电话',
      show: true,
      values: (() => {
        if (!searchKeyStuPhone.value)
          return []

        // 优先使用缓存的已选学员/电话详细信息
        if (selectedStuPhoneSearchInfo.value && selectedStuPhoneSearchInfo.value.id === searchKeyStuPhone.value) {
          return [{ id: selectedStuPhoneSearchInfo.value.id, value: `${selectedStuPhoneSearchInfo.value.stuName}` }]
        }

        // 其次从学员/电话选项中查找
        const item = stuPhoneSearchOptions.value.find(
          opt => opt.id === searchKeyStuPhone.value,
        )
        if (item) {
          return [{ id: item.id, value: `${item.stuName}` }]
        }

        // 最后显示ID作为后备方案
        return [{ id: searchKeyStuPhone.value, value: searchKeyStuPhone.value }]
      })(),
    },
    {
      type: 'selectInputKeySearch',
      label: props.searchLabel,
      show: true,
      values: (() => {
        if (!selectInputKey.value)
          return []
        if (props.renderClassListOptions) {
          const item = searchResultOptions.value.find(
            opt => opt.id === selectInputKey.value,
          )
          return item ? [{ id: item.id, value: `${item.value}` }] : []
        }
        else {
          const item = courseListOptions.value.find(
            opt => opt.id === selectInputKey.value,
          )
          return item ? [{ id: item.id, value: `${item.value}` }] : []
        }
      })(),
    },
    {
      type: 'searchInputKeySearch',
      label: props.searchLabel,
      show: true,
      values: (() => {
        if (!searchInputKey.value)
          return []
        return [{ id: 0, value: searchInputKey.value }]
      })(),
    },
    // searchInputKey
    {
      type: 'notFollowDays',
      label: '未跟进天数',
      show: props.displayArray.includes('notFollowDays'),
      values: notFollowDaysSelected.value
        ? [{ id: 'notFollowDays', value: `> ${notFollowDaysSelected.value}天` }]
        : [],
    },
    {
      type: 'intention',
      label: '意向度',
      show: props.displayArray.includes('intention'),
      values: customOptions.value.filter(opt =>
        selectedValues.value.includes(opt.id),
      ),
    },
    {
      type: 'grade',
      label: '年级',
      show: props.displayArray.includes('grade'),
      values: gradeOptions.value.filter(opt => gradeVals.value.includes(opt.id)),
    },

    {
      type: 'followStatus',
      label: '跟进状态',
      show: props.displayArray.includes('followStatus'),
      values: followStatusOptions.value.filter(opt =>
        followStatusVals.value.includes(opt.id),
      ),
    },

    {
      type: 'sex',
      label: '性别',
      show: props.displayArray.includes('sex'),
      values: sexOptions.value.filter(opt => sexVals.value.includes(opt.id)),
    },
    {
      type: 'isCommonCourse',
      label: '是否是通用课',
      show: props.displayArray.includes('isCommonCourse'),
      values: isCommonCourseOptions.value.filter(opt => isCommonCourseVals.value.includes(opt.id)),
    },
    {
      type: 'followMethod',
      label: '跟进方式',
      show: props.displayArray.includes('followMethod'),
      values: followMethodOptions.value.filter(opt => followMethodVals.value.includes(opt.id)),
    },
    {
      type: 'visitStatus',
      label: '回访状态',
      show: props.displayArray.includes('visitStatus'),
      values: visitStatusOptions.value.filter(opt => visitStatusVals.value.includes(opt.id)),
    },
    {
      type: 'payrollStatus',
      label: '工资条状态',
      show: props.displayArray.includes('payrollStatus'),
      values: payrollStatusOptions.value.filter(opt => payrollStatusVals.value.includes(opt.id)),
    },
    {
      type: 'stuStatus',
      label: '学员状态',
      show: props.displayArray.includes('stuStatus'),
      values: stuStatusOptions.value.filter(opt => stuStatusVals.value.includes(opt.id)),
    },
    {
      type: 'createUser',
      label: props.createUserLabel,
      show: props.displayArray.includes('createUser'),
      values: (() => {
        if (!createUserVals.value)
          return []

        // 优先使用缓存的详细信息
        if (selectedCreateUserInfo.value && selectedCreateUserInfo.value.id === createUserVals.value) {
          return [{ id: selectedCreateUserInfo.value.id, value: `${selectedCreateUserInfo.value.nickName}` }]
        }

        // 其次从当前选项中查找
        const item = createUserOptions.value.find(
          opt => opt.id === createUserVals.value,
        )
        if (item) {
          return [{ id: item.id, value: `${item.nickName}` }]
        }

        // 最后显示ID作为后备方案
        return [{ id: createUserVals.value, value: createUserVals.value }]
      })(),
    },
    {
      type: 'teachingMethod',
      label: '授课方式',
      show: props.displayArray.includes('teachingMethod'),
      values: teachingMethodOptions.value.filter(
        opt => opt.id === teachingMethodVals.value,
      ),
    },
    {
      type: 'sellStatus',
      label: '售卖状态',
      show: props.displayArray.includes('sellStatus'),
      values: sellStatusOptions.value.filter(
        opt => opt.id === sellStatusVals.value,
      ),
    },
    {
      type: 'hasTrialPrice',
      label: '是否有体验价',
      show: props.displayArray.includes('hasTrialPrice'),
      values: hasTrialPriceOptions.value.filter(
        opt => opt.id === hasTrialPriceVals.value,
      ),
    },
    {
      type: 'isMicroSchoolSale',
      label: '是否开启微校售卖',
      show: props.displayArray.includes('isMicroSchoolSale'),
      values: isMicroSchoolSaleOptions.value.filter(
        opt => opt.id === isMicroSchoolSaleVals.value,
      ),
    },
    {
      type: 'isMicroSchoolDisplay',
      label: '是否开启微校展示',
      show: props.displayArray.includes('isMicroSchoolDisplay'),
      values: isMicroSchoolDisplayOptions.value.filter(
        opt => opt.id === isMicroSchoolDisplayVals.value,
      ),
    },
    {
      type: 'salesPerson',
      label: props.salesPersonLabel,
      show: props.displayArray.includes('salesPerson'),
      values: (() => {
        if (!selectSalesPersonVals.value)
          return []

        // 优先使用缓存的详细信息
        if (selectedSalesPersonInfo.value && selectedSalesPersonInfo.value.id === selectSalesPersonVals.value) {
          return [{ id: selectedSalesPersonInfo.value.id, value: `${selectedSalesPersonInfo.value.nickName}` }]
        }

        // 其次从销售员选项中查找（不再从创建人选项中查找）
        const item = salesPersonOptions.value.find(
          opt => opt.id === selectSalesPersonVals.value,
        )
        if (item) {
          return [{ id: item.id, value: `${item.nickName}` }]
        }

        // 最后显示ID作为后备方案
        return [{ id: selectSalesPersonVals.value, value: selectSalesPersonVals.value }]
      })(),
    },
    {
      type: 'className',
      label: '班级名称',
      show: props.displayArray.includes('className'),
      values: classNameOptions.value.filter(opt => classNameVals.value.includes(opt.id)),
    },
    {
      type: 'enrolledCourse',
      label: '报读课程',
      show: props.displayArray.includes('enrolledCourse'),
      values: courseListOptions.value.filter(opt => enrolledCourseVals.value.includes(opt.id)),
    },
    {
      type: 'validityPeriod',
      label: '有效期至',
      show: props.displayArray.includes('validityPeriod'),
      values:
        validityPeriodVals.value.length === 2
          ? [
            {
              id: 'dateRange',
              value: `${validityPeriodVals.value[0]} 至 ${validityPeriodVals.value[1]}`,
            },
          ]
          : [],
    },
    {
      type: 'classTeacher',
      label: '班主任',
      show: props.displayArray.includes('classTeacher'),
      values: (() => {
        if (!classTeacherVals.value)
          return []

        // 优先使用缓存的详细信息
        if (selectedClassTeacherInfo.value && selectedClassTeacherInfo.value.id === classTeacherVals.value) {
          return [{ id: selectedClassTeacherInfo.value.id, value: `${selectedClassTeacherInfo.value.nickName}` }]
        }

        // 其次从班主任选项中查找
        const item = classTeacherOptions.value.find(
          opt => opt.id === classTeacherVals.value,
        )
        if (item) {
          return [{ id: item.id, value: `${item.nickName}` }]
        }

        // 最后显示ID作为后备方案
        return [{ id: classTeacherVals.value, value: classTeacherVals.value }]
      })(),
    },
    {
      type: 'isArrears',
      label: '是否欠费',
      show: props.displayArray.includes('isArrears'),
      values: isArrearsOptions.value.filter(
        opt => opt.id === isArrearsVals.value,
      ),
    },
    {
      type: 'enableStatus',
      label: '启用状态',
      show: props.displayArray.includes('enableStatus'),
      values: enableStatusOptions.value.filter(
        opt => opt.id === enableStatusVals.value,
      ),
    },
    {
      type: 'orderArrearStatus',
      label: '欠费状态',
      show: props.displayArray.includes('orderArrearStatus'),
      values: orderArrearStatusOptions.value.filter(opt => orderArrearStatusVals.value.includes(opt.id)),
    },
    {
      type: 'lastClassTime',
      label: '最近上课时间',
      show: props.displayArray.includes('lastClassTime'),
      values:
        lastClassTimeVals.value.length === 2
          ? [
            {
              id: 'dateRange',
              value: `${lastClassTimeVals.value[0]} 至 ${lastClassTimeVals.value[1]}`,
            },
          ]
          : [],
    },
    {
      type: 'lastEditedTime',
      label: '最近编辑时间',
      show: props.displayArray.includes('lastEditedTime'),
      values:
        lastEditedTimeVals.value.length === 2
          ? [
            {
              id: 'dateRange',
              value: `${lastEditedTimeVals.value[0]} 至 ${lastEditedTimeVals.value[1]}`,
            },
          ]
          : [],
    },
    {
      type: 'createTime',
      label: props.createTimeLabel,
      show: props.displayArray.includes('createTime'),
      values:
        createTimeVals.value.length === 2
          ? [
            {
              id: 'dateRange',
              value: `${createTimeVals.value[0]} 至 ${createTimeVals.value[1]}`,
            },
          ]
          : [],
    },
    {
      type: 'followTime',
      label: '跟进时间',
      show: props.displayArray.includes('followTime'),
      values:
        followTimeVals.value.length === 2
          ? [
            {
              id: 'dateRange',
              value: `${followTimeVals.value[0]} 至 ${followTimeVals.value[1]}`,
            },
          ]
          : [],
    },
    {
      type: 'birthday',
      label: '生日',
      show: props.displayArray.includes('birthday'),
      values:
        birthdayVals.value.length === 2
          ? [
            {
              id: 'dateRange',
              value: `${birthdayVals.value[0]} 至 ${birthdayVals.value[1]}`,
            },
          ]
          : [],
    },
    {
      type: 'approveNumber',
      label: '审批编号',
      show: props.displayArray.includes('approveNumber'),
      values: (() => {
        if (!approveNumberVals.value)
          return []
        return [{ id: 0, value: approveNumberVals.value }]
      })(),
    },
    {
      type: 'orderNumber',
      label: '订单编号',
      show: props.displayArray.includes('orderNumber'),
      values: (() => {
        if (!orderNumberVals.value)
          return []
        return [{ id: 0, value: orderNumberVals.value }]
      })(),
    },
    {
      type: 'wxChat',
      label: '微信号',
      show: props.displayArray.includes('wxChat'),
      values: (() => {
        if (!wxChatVals.value)
          return []
        return [{ id: 0, value: wxChatVals.value }]
      })(),
    },
    {
      type: 'school',
      label: '就读学校',
      show: props.displayArray.includes('school'),
      values: (() => {
        if (!schoolVals.value)
          return []
        return [{ id: 0, value: schoolVals.value }]
      })(),
    },
    {
      type: 'address',
      label: '家庭住址',
      show: props.displayArray.includes('address'),
      values: (() => {
        if (!addressVals.value)
          return []
        return [{ id: 0, value: addressVals.value }]
      })(),
    },
    {
      type: 'hobbies',
      label: '兴趣爱好',
      show: props.displayArray.includes('hobbies'),
      values: (() => {
        if (!hobbiesVals.value)
          return []
        return [{ id: 0, value: hobbiesVals.value }]
      })(),
    },
    {
      type: 'assignTime',
      label: '分配时间',
      show: props.displayArray.includes('assignTime'),
      values:
        assignTimeVals.value.length === 2
          ? [
            {
              id: 'dateRange',
              value: `${assignTimeVals.value[0]} 至 ${assignTimeVals.value[1]}`,
            },
          ]
          : [],
    },
    {
      type: 'lastFollowTime',
      label: '最近跟进',
      show: props.displayArray.includes('lastFollowTime'),
      values:
        lastFollowTimeVals.value.length === 2
          ? [
            {
              id: 'dateRange',
              value: `${lastFollowTimeVals.value[0]} 至 ${lastFollowTimeVals.value[1]}`,
            },
          ]
          : [],
    },
    {
      type: 'nextFollowTime',
      label: '下次跟进',
      show: props.displayArray.includes('nextFollowTime'),
      values:
        nextFollowTimeVals.value.length === 2
          ? [
            {
              id: 'dateRange',
              value: `${nextFollowTimeVals.value[0]} 至 ${nextFollowTimeVals.value[1]}`,
            },
          ]
          : [],
    },
    {
      type: 'applyTime',
      label: '申请时间',
      show: props.displayArray.includes('applyTime'),
      values:
        applyTimeVals.value.length === 2
          ? [
            {
              id: 'dateRange',
              value: `${applyTimeVals.value[0]} 至 ${applyTimeVals.value[1]}`,
            },
          ]
          : [],
    },
    {
      type: 'finishTime',
      label: '审批完成时间',
      show: props.displayArray.includes('finishTime'),
      values:
        finishTimeVals.value.length === 2
          ? [
            {
              id: 'dateRange',
              value: `${finishTimeVals.value[0]} 至 ${finishTimeVals.value[1]}`,
            },
          ]
          : [],
    },
    {
      type: 'payTime',
      label: '支付日期',
      show: props.displayArray.includes('payTime'),
      values:
        payTimeVals.value.length === 2
          ? [
            {
              id: 'dateRange',
              value: `${payTimeVals.value[0]} 至 ${payTimeVals.value[1]}`,
            },
          ]
          : [],
    },
    {
      type: 'classEndingTime',
      label: '结课时间',
      show: props.displayArray.includes('classEndingTime'),
      values:
        classEndingTimeVals.value.length === 2
          ? [
            {
              id: 'dateRange',
              value: `${classEndingTimeVals.value[0]} 至 ${classEndingTimeVals.value[1]}`,
            },
          ]
          : [],
    },
    {
      type: 'classStopTime',
      label: '停课时间',
      show: props.displayArray.includes('classStopTime'),
      values:
        classStopTimeVals.value.length === 2
          ? [
            {
              id: 'dateRange',
              value: `${classStopTimeVals.value[0]} 至 ${classStopTimeVals.value[1]}`,
            },
          ]
          : [],
    },
    {
      type: 'intentionCourse',
      label: '意向课程',
      show: props.displayArray.includes('intentionCourse'),
      values: (() => {
        if (!selectCourseValues.value)
          return []

        // 优先使用缓存的详细信息
        if (selectedCourseInfo.value && selectedCourseInfo.value.id === selectCourseValues.value) {
          return [{ id: selectedCourseInfo.value.id, value: selectedCourseInfo.value.value }]
        }

        // 其次从课程列表选项中查找
        const item = courseListOptions.value.find(
          opt => opt.id === selectCourseValues.value,
        )
        if (item) {
          return [{ id: item.id, value: item.value }]
        }

        // 最后显示ID作为后备方案
        return [{ id: selectCourseValues.value, value: selectCourseValues.value }]
      })(),
    },
    {
      type: 'recommend',
      label: '推荐人',
      show: props.displayArray.includes('recommend'),
      values: (() => {
        if (!recommendVals.value)
          return []

        // 优先使用缓存的详细信息
        if (selectedRecommendInfo.value && selectedRecommendInfo.value.id === recommendVals.value) {
          return [{ id: selectedRecommendInfo.value.id, value: `${selectedRecommendInfo.value.stuName}` }]
        }

        // 其次从推荐人选项中查找
        const item = recommendOptions.value.find(
          opt => opt.id === recommendVals.value,
        )
        if (item) {
          return [{ id: item.id, value: `${item.stuName}` }]
        }

        // 最后显示ID作为后备方案
        return [{ id: recommendVals.value, value: recommendVals.value }]
      })(),
    },
    {
      type: 'stuPhoneSearch',
      label: '学员/电话',
      show: props.displayArray.includes('stuPhoneSearch'),
      values: (() => {
        if (!stuPhoneSearchVals.value)
          return []
        const item = stuPhoneSearchOptions.value.find(
          opt => opt.id === stuPhoneSearchVals.value,
        )
        return item ? [{ id: item.id, value: `${item.stuName}` }] : []
      })(),
    },
    {
      type: 'channelCategory',
      label: '渠道',
      show: props.displayArray.includes('channelCategory'),
      values: (() => {
        // 确保有选中值
        if (!selectChannelValsFilter.value || selectChannelValsFilter.value.length === 0) {
          return []
        }

        // 查找匹配的选项（包括子节点）
        return selectChannelValsFilter.value.map((id) => {
          // 递归查找节点
          const node = findNodeById(channelOptions.value, id)
          return node ? { id: node.id, value: node.name } : null
        }).filter(Boolean) // 过滤掉未找到的节点
      })(),
    },
    {
      type: 'channelStatus',
      label: '渠道状态',
      show: props.displayArray.includes('channelStatus'),
      values: channelListOptions.value.filter(opt =>
        selectChannelVals.value.includes(opt.id),
      ),
    },
    {
      type: 'trialPurchaseStatus',
      label: '体验课购买状态',
      show: props.displayArray.includes('trialPurchaseStatus'),
      values: trialPurchaseStatusOptions.value.filter(opt =>
        trialPurchaseStatusVals.value.includes(opt.id),
      ),
    },
    {
      type: 'channelType',
      label: '渠道类型',
      show: props.displayArray.includes('channelType'),
      values: channelTypeOptions.value.filter(opt =>
        selectChannelTypeVals.value.includes(opt.id),
      ),
    },
    {
      type: 'channelCategoryType',
      label: '渠道分类',
      show: props.displayArray.includes('channelCategoryType'),
      values: channelCategoryOptions.value.filter(opt =>
        selectChannelCategoryVals.value.includes(opt.id),
      ),
    },
    {
      type: 'positionRole',
      label: '任职角色',
      show: props.displayArray.includes('positionRole'),
      values: positionRoleOptions.value.filter(opt =>
        positionRoleVals.value.includes(opt.id),
      ),
    },
    {
      type: 'subject',
      label: '科目',
      show: props.displayArray.includes('subject'),
      values: subjectOptions.value.filter(
        opt => opt.id === selectSubjectVals.value,
      ),
    },
    {
      type: 'courseCategory',
      label: '课程类别',
      show: props.displayArray.includes('courseCategory'),
      values: courseCategoryOptions.value.filter(
        opt => opt.id === selectCourseCategoryVals.value,
      ),
    },
    {
      type: 'studentStatus',
      label: '学员状态',
      show: props.displayArray.includes('studentStatus'),
      values: studentStatusOptions.value.filter(
        opt => opt.id === selectStudentStatusVals.value,
      ),
    },
    {
      type: 'accountStatus',
      label: '账号状态',
      show: props.displayArray.includes('accountStatus'),
      values: accountStatusOptions.value.filter(
        opt => opt.id === selectAccountStatusVals.value,
      ),
    },
    {
      type: 'userType',
      label: '员工类型',
      show: props.displayArray.includes('userType'),
      values: userTypeOptions.value.filter(
        opt => opt.id === selectUserTypeVals.value,
      ),
    },
    {
      type: 'currentStatus',
      label: '当前状态',
      show: props.displayArray.includes('currentStatus'),
      values: currentStatusOptions.value.filter(opt =>
        selectCurrentStatusVals.value.includes(opt.id),
      ),
    },
    {
      type: 'orNotFenClass',
      label: '是否分班',
      show: props.displayArray.includes('orNotFenClass'),
      values: orNotFenClassOptions.value.filter(opt =>
        selectOrNotFenClassVals.value.includes(opt.id),
      ),
    },
    {
      type: 'billingMode',
      label: '收费方式',
      show: props.displayArray.includes('billingMode'),
      values: billingModeOptions.value.filter(opt =>
        selectBillingModeVals.value.includes(opt.id),
      ),
    },
    {
      type: 'isSetExpirationDate',
      label: '是否设置有效期',
      show: props.displayArray.includes('isSetExpirationDate'),
      values: isSetExpirationDateOptions.value.filter(
        opt => opt.id === selectIsSetExpirationDateVals.value,
      ),
    },
    {
      type: 'openClassStatus',
      label: '开班状态',
      show: props.displayArray.includes('openClassStatus'),
      values: openClassStatusOptions.value.filter(
        opt => opt.id === selectOpenClassStatusVals.value,
      ),
    },
    {
      type: 'doYouSchedule',
      label: '是否排课',
      show: props.displayArray.includes('doYouSchedule'),
      values: doYouScheduleOptions.value.filter(
        opt => opt.id === selectDoYouScheduleVals.value,
      ),
    },
    {
      type: 'age',
      label: '年龄',
      show: props.displayArray.includes('age'),
      values: selectedAgeRange.value && (selectedAgeRange.value.min != null || selectedAgeRange.value.max != null)
        ? [{
          id: 'age',
          value: selectedAgeRange.value.min != null && selectedAgeRange.value.max != null
            ? `${selectedAgeRange.value.min} ~ ${selectedAgeRange.value.max} 岁`
            : selectedAgeRange.value.min != null
              ? `≥ ${selectedAgeRange.value.min} 岁`
              : `≤ ${selectedAgeRange.value.max} 岁`,
        }]
        : [],
    },
    {
      type: 'recommended',
      label: '是否被推荐',
      show: props.displayArray.includes('recommended'),
      values: recommendedOptions.value.filter(opt =>
        recommendedVals.value.includes(opt.id),
      ),
    },
    {
      type: 'hasSalesPerson',
      label: '是否有销售员',
      show: props.displayArray.includes('hasSalesPerson'),
      values: hasSalesPersonOptions.value.filter(opt =>
        hasSalesPersonVals.value.includes(opt.id),
      ),
    },
    {
      type: 'remaining',
      label: '剩余数量',
      show: props.displayArray.includes('remaining'),
      values: selectedRemainingRange.value && (selectedRemainingRange.value.min != null || selectedRemainingRange.value.max != null)
        ? [{
          id: 'remaining',
          value: (() => {
            const { min, max, mode } = selectedRemainingRange.value
            const modeText = mode === 'lesson' ? '按课时' : mode === 'period' ? '按时段' : '按金额'
            const unit = mode === 'lesson' ? '课时' : mode === 'period' ? '天' : '元'

            if (min != null && max != null) {
              return `${modeText} ${min} ~ ${max} ${unit}`
            }
            else if (min != null) {
              return `${modeText} ≥ ${min} ${unit}`
            }
            else {
              return `${modeText} ≤ ${max} ${unit}`
            }
          })(),
        }]
        : [],
    },

    // 为自定义搜索字段添加条件
    ...props.customIsDisplayList.filter(item => searchInputVals.value[item.id]).map(item => ({
      type: `customSearch_${item.id}`,
      label: item.fieldKey,
      show: props.displayArray.includes('customSearch'),
      values: (() => {
        const value = searchInputVals.value[item.id]
        if (item.fieldType === 3 && Array.isArray(value) && value.length === 2) {
          // 日期类型，格式化为范围显示
          return [{ id: item.id, value: `${value[0]} 至 ${value[1]}` }]
        }
        else if (value) {
          // 其他类型（文本、数字）
          return [{ id: item.id, value }]
        }
        return []
      })(),
    })).filter(item => item.values.length > 0),

    // 为课程属性列表添加条件
    ...props.courseAttributeList.filter(item => searchInputVals.value[item.id]).map(item => ({
      type: `courseAttribute_${item.id}`,
      label: item.name,
      show: props.displayArray.includes('courseAttribute'),
      values: (() => {
        const value = searchInputVals.value[item.id]
        if (value) {
          // 从课程属性选项中找到对应的显示文本
          const option = courseAttributeOptions.value.find(opt => opt.id === value.id)
          // 优先使用找到的选项名称，其次使用value.name，最后使用value本身
          const displayValue = option ? option.name : (value.name || value)
          return [{ id: item.id, value: displayValue }]
        }
        return []
      })(),
    })).filter(item => item.values.length > 0),

    {
      type: 'performanceAllocationStatus',
      label: '业绩分配状态',
      show: props.displayArray.includes('performanceAllocationStatus'),
      values: performanceAllocationStatusOptions.value.filter(
        opt => opt.id === performanceAllocationStatusVals.value,
      ),
    },
    {
      type: 'orderType',
      label: '订单类型',
      show: props.displayArray.includes('orderType'),
      values: orderTypeOptions.value.filter(opt => orderTypeVals.value.includes(opt.id)),
    },
    {
      type: 'orderTag',
      label: '订单标签',
      show: props.displayArray.includes('orderTag'),
      values: orderTagOptions.value.filter(opt => orderTagVals.value.includes(opt.id)),
    },
    {
      type: 'enrollType',
      label: '报读类型',
      show: props.displayArray.includes('enrollType'),
      values: enrollTypeOptions.value.filter(opt => enrollTypeVals.value.includes(opt.id)),
    },
    {
      type: 'productType',
      label: '商品类型',
      show: props.displayArray.includes('productType'),
      values: productTypeOptions.value.filter(opt => productTypeVals.value.includes(opt.id)),
    },
    {
      type: 'orderStatus',
      label: '订单状态',
      show: props.displayArray.includes('orderStatus'),
      values: orderStatusOptions.value.filter(opt => orderStatusVals.value.includes(opt.id)),
    },
    {
      type: 'approvalStatus',
      label: '审批状态',
      show: props.displayArray.includes('approvalStatus'),
      values: approvalStatusOptions.value.filter(opt => approvalStatusVals.value.includes(opt.id)),
    },
    {
      type: 'orderSource',
      label: '订单来源',
      show: props.displayArray.includes('orderSource'),
      values: orderSourceOptions.value.filter(opt => orderSourceVals.value.includes(opt.id)),
    },
    {
      type: 'dealDate',
      label: '经办日期',
      show: props.displayArray.includes('dealDate'),
      values: dealDateVals.value.length === 2
        ? [{ id: 'dealDate', value: `${dealDateVals.value[0]} 至 ${dealDateVals.value[1]}` }]
        : [],
    },
    {
      type: 'latestPaidTime',
      label: '最近支付时间',
      show: props.displayArray.includes('latestPaidTime'),
      values: latestPaidTimeVals.value.length === 2
        ? [{ id: 'latestPaidTime', value: `${latestPaidTimeVals.value[0]} 至 ${latestPaidTimeVals.value[1]}` }]
        : [],
    },
    {
      type: 'handleContent',
      label: '办理内容',
      show: props.displayArray.includes('handleContent'),
      values: courseListOptions.value.filter(opt => enrolledCourseVals.value.includes(opt.id)),
    },
  ]
  return conditions
    .filter(item => item.values.length > 0 && item.show)
    .sort((a, b) => (lastUpdated[a.type] || 0) - (lastUpdated[b.type] || 0))
})
// watch(selectDptVals, () => {
//   dptName.value = getNameById(selectDptVals.value)
// })
// 监听各条件变化，更新最后操作时间
watch(selectedValues, () => (lastUpdated.intention = Date.now()))
watch(hasSalesPersonVals, () => (lastUpdated.hasSalesPerson = Date.now()))
watch(followStatusVals, () => (lastUpdated.followStatus = Date.now()))
watch(sexVals, () => (lastUpdated.sex = Date.now()))
watch(followMethodVals, () => (lastUpdated.followMethod = Date.now()))
watch(visitStatusVals, () => (lastUpdated.visitStatus = Date.now()))
watch(payrollStatusVals, () => (lastUpdated.payrollStatus = Date.now()))
watch(stuStatusVals, () => (lastUpdated.stuStatus = Date.now()))
watch(selectChannelVals, () => (lastUpdated.channelStatus = Date.now()))
watch(selectChannelTypeVals, () => {
  lastUpdated.channelType = Date.now()
  // 移除这一行，避免重复调用
  // handleChannelTypeChange();
})
watch(
  selectChannelValsFilter,
  () => (lastUpdated.channelCategory = Date.now()),
)
watch(createUserVals, () => {
  lastUpdated.createUser = Date.now()
  // 自动更新创建人缓存信息
  if (createUserVals.value && createUserOptions.value.length > 0) {
    const selectedUser = createUserOptions.value.find(user => user.id === createUserVals.value)
    if (selectedUser) {
      selectedCreateUserInfo.value = selectedUser
    }
  }
})
watch(teachingMethodVals, () => (lastUpdated.teachingMethod = Date.now()))
watch(selectSalesPersonVals, () => {
  lastUpdated.salesPerson = Date.now()
  // 自动更新销售员缓存信息
  if (selectSalesPersonVals.value && salesPersonOptions.value.length > 0) {
    const selectedUser = salesPersonOptions.value.find(user => user.id === selectSalesPersonVals.value)
    if (selectedUser) {
      selectedSalesPersonInfo.value = selectedUser
    }
  }
})
watch(isArrearsVals, () => (lastUpdated.isArrears = Date.now()))
watch(classTeacherVals, () => (lastUpdated.classTeacher = Date.now()))
watch(lastClassTimeVals, () => (lastUpdated.lastClassTime = Date.now()))
watch(createTimeVals, () => (lastUpdated.createTime = Date.now()))
watch(birthdayVals, () => (lastUpdated.birthday = Date.now())) // 添加生日的watch
watch(assignTimeVals, () => (lastUpdated.assignTime = Date.now()))
watch(followTimeVals, () => (lastUpdated.followTime = Date.now()))
watch(lastFollowTimeVals, () => (lastUpdated.lastFollowTime = Date.now()))
watch(nextFollowTimeVals, () => (lastUpdated.nextFollowTime = Date.now()))
watch(applyTimeVals, () => (lastUpdated.applyTime = Date.now()))
watch(payTimeVals, () => (lastUpdated.payTime = Date.now()))
watch(classEndingTimeVals, () => {
  lastUpdated.classEndingTime = Date.now()
  debouncedEmit('classEndingTimeFilter', classEndingTimeVals.value)
})
watch(classStopTimeVals, () => {
  lastUpdated.classStopTime = Date.now()
  debouncedEmit('classStopTimeFilter', classStopTimeVals.value)
})
watch(selectCourseValues, () => {
  lastUpdated.intentionCourse = Date.now()
  // 自动更新课程缓存信息
  if (selectCourseValues.value && courseListOptions.value.length > 0) {
    const selectedCourse = courseListOptions.value.find(course => course.id === selectCourseValues.value)
    if (selectedCourse) {
      selectedCourseInfo.value = selectedCourse
    }
  }
})
watch(recommendedVals, () => (lastUpdated.recommended = Date.now()))
watch(trialPurchaseStatusVals, () => (lastUpdated.trialPurchaseStatus = Date.now()))
watch(lastEditedTimeVals, () => (lastUpdated.lastEditedTime = Date.now()))
watch(finishTimeVals, () => (lastUpdated.finishTime = Date.now()))

watch(
  () => quickFilters.value.map(q => q.selected),
  () => (lastUpdated.quick = Date.now()),
  { deep: true },
)
watch(
  () => quickOneToOneFilters.value.map(q => q.selected),
  () => (lastUpdated.quickOneToOne = Date.now()),
  { deep: true },
)
watch(
  () => quickApproveFilters.value.map(q => q.selected),
  () => (lastUpdated.quickApprove = Date.now()),
  { deep: true },
)
watch(recommendVals, () => {
  lastUpdated.recommend = Date.now()
  // 自动更新推荐人缓存信息
  if (recommendVals.value && recommendOptions.value.length > 0) {
    const selectedUser = recommendOptions.value.find(user => user.id === recommendVals.value)
    if (selectedUser) {
      selectedRecommendInfo.value = selectedUser
    }
  }
})
watch(searchKeyStuPhone, () => {
  lastUpdated.stuPhoneSearchNew = Date.now()
  // 自动更新学员/电话搜索缓存信息
  if (searchKeyStuPhone.value && stuPhoneSearchOptions.value.length > 0) {
    const selectedUser = stuPhoneSearchOptions.value.find(user => user.id === searchKeyStuPhone.value)
    if (selectedUser) {
      selectedStuPhoneSearchInfo.value = selectedUser
    }
  }
})

// 同步「真正用于请求的 id」和下拉框回显对象，保证始终能显示学员姓名
watch(
  [searchKeyStuPhone, selectedStuPhoneSearchInfo, stuPhoneSearchOptions],
  () => {
    if (!searchKeyStuPhone.value) {
      searchKeyStuPhoneModel.value = undefined
      return
    }

    // 优先使用缓存的详细信息
    let user = null
    if (selectedStuPhoneSearchInfo.value && selectedStuPhoneSearchInfo.value.id === searchKeyStuPhone.value) {
      user = selectedStuPhoneSearchInfo.value
    }
    else if (stuPhoneSearchOptions.value && stuPhoneSearchOptions.value.length > 0) {
      user = stuPhoneSearchOptions.value.find(u => u.id === searchKeyStuPhone.value) || null
    }

    if (user) {
      searchKeyStuPhoneModel.value = { value: user.id, label: user.stuName }
    }
    else {
      // 兜底：至少让回显不是纯 id，可以按需要再优化
      searchKeyStuPhoneModel.value = { value: searchKeyStuPhone.value, label: String(searchKeyStuPhone.value) }
    }
  },
  { immediate: true },
)
watch(selectSubjectVals, () => (lastUpdated.subject = Date.now()))
watch(
  selectCourseCategoryVals,
  () => (lastUpdated.courseCategory = Date.now()),
)
watch(selectStudentStatusVals, () => (lastUpdated.studentStatus = Date.now()))
watch(selectAccountStatusVals, () => (lastUpdated.accountStatus = Date.now()))
watch(selectUserTypeVals, () => (lastUpdated.userType = Date.now()))
watch(selectCurrentStatusVals, () => (lastUpdated.currentStatus = Date.now()))
watch(selectOrNotFenClassVals, () => (lastUpdated.orNotFenClass = Date.now()))
watch(selectBillingModeVals, () => (lastUpdated.billingMode = Date.now()))
watch(orderTagVals, () => (lastUpdated.orderTag = Date.now()))
watch(enrollTypeVals, () => (lastUpdated.enrollType = Date.now()))
watch(productTypeVals, () => (lastUpdated.productType = Date.now()))
watch(orderStatusVals, () => (lastUpdated.orderStatus = Date.now()))
watch(orderArrearStatusVals, () => (lastUpdated.orderArrearStatus = Date.now()))
watch(orderSourceVals, () => (lastUpdated.orderSource = Date.now()))
watch(dealDateVals, () => (lastUpdated.dealDate = Date.now()))
watch(latestPaidTimeVals, () => (lastUpdated.latestPaidTime = Date.now()))
watch(selectedRemainingRange, () => (lastUpdated.remaining = Date.now()))
watch(wxChatVals, () => (lastUpdated.wxChat = Date.now()))
watch(schoolVals, () => (lastUpdated.school = Date.now()))
watch(addressVals, () => (lastUpdated.address = Date.now()))
watch(hobbiesVals, () => (lastUpdated.hobbies = Date.now()))
watch(selectChannelCategoryVals, () => (lastUpdated.channelCategoryType = Date.now()))
watch(positionRoleVals, () => (lastUpdated.positionRole = Date.now()))
watch(
  selectIsSetExpirationDateVals,
  () => (lastUpdated.isSetExpirationDate = Date.now()),
)
watch(
  selectOpenClassStatusVals,
  () => (lastUpdated.openClassStatus = Date.now()),
)
watch(selectDoYouScheduleVals, () => (lastUpdated.doYouSchedule = Date.now()))
watch(isCommonCourseVals, () => (lastUpdated.isCommonCourse = Date.now()))
watch(teachingMethodVals, () => (lastUpdated.teachingMethod = Date.now()))
watch(sellStatusVals, () => (lastUpdated.sellStatus = Date.now()))
watch(hasTrialPriceVals, () => (lastUpdated.hasTrialPrice = Date.now()))
watch(isMicroSchoolSaleVals, () => (lastUpdated.isMicroSchoolSale = Date.now()))
watch(isMicroSchoolDisplayVals, () => (lastUpdated.isMicroSchoolDisplay = Date.now()))
watch(performanceAllocationStatusVals, () => (lastUpdated.performanceAllocationStatus = Date.now()))
watch(orderTypeVals, () => (lastUpdated.orderType = Date.now()))
watch(approvalStatusVals, () => (lastUpdated.approvalStatus = Date.now()))
// 观察筛选条件变化，维护顺序队列
watch(
  selectedConditions,
  (newConditions) => {
    const newTypes = newConditions.map(c => c.type)

    // 保留仍然存在的类型
    conditionOrder.value = conditionOrder.value.filter(t =>
      newTypes.includes(t),
    )

    // 添加新增的类型到队列末尾
    newTypes.forEach((t) => {
      if (!conditionOrder.value.includes(t)) {
        conditionOrder.value.push(t)
      }
    })
  },
  { deep: true },
)
// 最终使用的排序条件
const orderedConditions = computed(() => {
  return [...selectedConditions.value].sort(
    (a, b) =>
      conditionOrder.value.indexOf(a.type)
      - conditionOrder.value.indexOf(b.type),
  )
})
// 清空所有筛选
const clearAll = debounce(() => {
  // 重置多选类
  [
    selectedValues,
    gradeVals,
    hasSalesPersonVals,
    followStatusVals,
    sexVals,
    isCommonCourseVals,
    followMethodVals,
    visitStatusVals,
    payrollStatusVals,
    stuStatusVals,
    ...(props.type !== 'noDelCreateTime' ? [createTimeVals] : []), // 当type为noDelCreateTime时，不清除创建时间
    lastEditedTimeVals,
    birthdayVals,
    assignTimeVals,
    followTimeVals,
    lastFollowTimeVals,
    nextFollowTimeVals,
    applyTimeVals,
    payTimeVals,
    classEndingTimeVals,
    classStopTimeVals,
    // 报读列表新增：有效期至、最近上课时间
    validityPeriodVals,
    lastClassTimeVals,
    selectChannelVals,
    selectChannelTreeVals,
    selectChannelValsFilter,
    selectChannelTypeVals,
    selectCurrentStatusVals,
    selectOrNotFenClassVals,
    selectBillingModeVals,
    selectAccountStatusVals,
    selectUserTypeVals,
    recommendedVals,
    selectChannelCategoryVals,
    positionRoleVals,
    trialPurchaseStatusVals, // 添加新的状态到清空列表
    orderTypeVals, // 添加订单类型到清空列表
    orderTagVals,
    enrollTypeVals,
    productTypeVals,
    orderStatusVals,
    orderArrearStatusVals,
    orderSourceVals,
  ].forEach(ref => (ref.value = []))

  // 使用nextTick确保所有状态更新完成后再一次性发出所有更新事件
  nextTick(() => {
    // 一次性发出所有更新事件，并标记为清空所有
    emit('update:channelTypeFilter', undefined, true)
    emit('update:channelStatusFilter', undefined, true)
    emit('update:channelCategoryFilter', undefined, true)
    emit('update:quickFilter', undefined, true)
    emit('update:intentionLevelFilter', [], true)
    emit('update:gradeFilter', [], true)
    emit('update:followStatusFilter', [], true)
    // 当type为noDelCreateTime时，不清除创建时间
    if (props.type !== 'noDelCreateTime') {
      emit('update:createTimeFilter', undefined, true)
    }
    emit('update:birthdayFilter', undefined, true)
    emit('update:lastFollowTimeFilter', undefined, true)
    emit('update:followTimeFilter', undefined, true)
    emit('update:sexFilter', [], true)
    emit('update:isCommonCourseFilter', [], true)
    emit('update:ageFilter', undefined, true)
    emit('update:channelFilter', undefined, true)
    emit('update:notFollowDaysFilter', undefined, true)
    emit('update:approveNumberFilter', undefined, true)
    emit('update:orderNumberFilter', undefined, true)
    emit('update:wxChatFilter', undefined, true)
    emit('update:schoolFilter', undefined, true)
    emit('update:addressFilter', undefined, true)
    emit('update:hobbiesFilter', undefined, true)
    emit('update:recommendedFilter', undefined, true)
    emit('update:hasSalesPersonFilter', undefined, true)
    emit('update:createUserFilter', undefined, true)
    emit('update:salesPersonFilter', undefined, true)
    emit('update:followMethodFilter', [], true)
    emit('update:visitStatusFilter', [], true)
    emit('update:payrollStatusFilter', [], true)
    emit('update:stuStatusFilter', [], true)
    emit('update:stuPhoneSearchFilter', [], true)
    emit('update:courseCategoryFilter', undefined, true)
    emit('update:teachingMethodFilter', undefined, true)
    emit('update:sellStatusFilter', undefined, true)
    emit('update:chargingMethodFilter', [], true)
    emit('update:hasTrialPriceFilter', undefined, true)
    emit('update:isMicroSchoolSaleFilter', undefined, true)
    emit('update:isMicroSchoolDisplayFilter', undefined, true)
    emit('update:lastEditedTimeFilter', [], true)
    emit('update:finishTimeFilter', [], true)
    emit('update:birthdayFilter', undefined, true)
    emit('update:lastFollowTimeFilter', undefined, true)
    emit('update:channelAccountStatus', undefined, true)
    emit('update:channelUserType', undefined, true)
    emit('update:performanceAllocationStatusFilter', undefined, true)
    emit('update:enableStatusFilter', undefined, true)
    emit('update:orderTypeFilter', [], true)
    emit('update:orderTagFilter', [], true)
    emit('update:enrollTypeFilter', [], true)
    emit('update:productTypeFilter', [], true)
    emit('update:approvalStatusFilter', [], true)
    emit('update:orderStatusFilter', [], true)
    emit('update:orderArrearStatusFilter', [], true)
    emit('update:orderSourceFilter', [], true)
    emit('update:dealDateFilter', undefined, true)
    emit('update:latestPaidTimeFilter', undefined, true)
    emit('update:handleContentFilter', undefined, true)
    emit('update:intentionCourseFilter', undefined, true)
  })

  resetNotFollowDays()
  notFollowDaysSelected.value = null
  resetAge()
  selectedAgeRange.value = null
  resetRemaining()
  selectedRemainingRange.value = null
  // 重置单选类
  quickFilters.value.forEach(q => (q.selected = false))
  quickOneToOneFilters.value.forEach(q => (q.selected = false))
  quickApproveFilters.value.forEach(q => (q.selected = false))

  // 重置创建人相关状态
  createUserVals.value = null
  selectedCreateUserInfo.value = null
  createUserOptions.value = []
  pagination.value.current = 1
  finished.value = false
  if (childRefs.value.createUser?.resetSearch) {
    childRefs.value.createUser.resetSearch()
  }

  // 重置销售员相关状态 - 独立处理
  selectSalesPersonVals.value = null
  selectedSalesPersonInfo.value = null
  salesPersonOptions.value = []
  salesPersonPagination.value.current = 1
  salesPersonFinished.value = false
  if (childRefs.value.salesPerson?.resetSearch) {
    childRefs.value.salesPerson.resetSearch()
  }

  // 继续重置其他状态
  teachingMethodVals.value = null
  dealDateVals.value = []
  latestPaidTimeVals.value = []
  sellStatusVals.value = null
  hasTrialPriceVals.value = null
  isMicroSchoolSaleVals.value = null
  isMicroSchoolDisplayVals.value = null
  enableStatusVals.value = null
  selectCourseValues.value = null
  selectedCourseInfo.value = null
  courseListOptions.value = []
  courseListPagination.value.current = 1
  courseListFinished.value = false
  if (childRefs.value.yiXiangcourse?.resetSearch) {
    childRefs.value.yiXiangcourse.resetSearch()
  }
  selectAccountStatusVals.value = null
  selectUserTypeVals.value = null
  // 报读列表新增：是否欠费、班主任等状态也一并重置
  isArrearsVals.value = null
  classTeacherVals.value = null
  selectedClassTeacherInfo.value = null
  // 重置推荐人相关状态 - 独立处理
  recommendVals.value = null
  selectedRecommendInfo.value = null
  recommendOptions.value = []
  recommendPagination.value.current = 1
  recommendFinished.value = false
  if (childRefs.value.recommend?.resetSearch) {
    childRefs.value.recommend.resetSearch()
  }

  // 重置学员/电话搜索相关状态 - 独立处理
  stuPhoneSearchVals.value = null
  selectedStuPhoneSearchInfo.value = null
  stuPhoneSearchOptions.value = []
  stuPhoneSearchPagination.value.current = 1
  stuPhoneSearchFinished.value = false
  if (childRefs.value.stuPhoneSearch?.resetSearch) {
    childRefs.value.stuPhoneSearch.resetSearch()
  }
  selectSubjectVals.value = null
  selectCourseCategoryVals.value = null
  selectStudentStatusVals.value = null
  performanceAllocationStatusVals.value = null
  selectIsSetExpirationDateVals.value = null
  selectOpenClassStatusVals.value = null
  selectDoYouScheduleVals.value = null
  wxChatVals.value = null
  orderNumberVals.value = ''
  schoolVals.value = null
  addressVals.value = null
  hobbiesVals.value = null
  teachingMethodVals.value = null

  // 向后兼容：清理原始的stuListOptions
  stuListOptions.value = []

  Object.values(childRefs.value).forEach((child) => {
    if (child?.resetSearch) {
      child.resetSearch()
    }
  })

  searchInputKey.value = undefined
  searchKeyOneToOne.value = undefined
  searchKeyStuPhone.value = undefined
  selectInputKey.value = undefined
  skipNextWatch = true // 跳过下次watch触发
  inputValue.value = undefined
  selectTeacher.value = undefined

  // 清空所有自定义搜索字段
  searchInputVals.value = {}

  // 单独处理自定义搜索字段的子组件重置
  if (props.customIsDisplayList && props.customIsDisplayList.length > 0) {
    props.customIsDisplayList.forEach((item) => {
      if ((item.fieldType == 1 || item.fieldType == 2 || item.fieldType == 3)) {
        const childRef = childRefs.value[`customSearchInput_${item.id}`]
        if (childRef && childRef.resetSearch) {
          childRef.resetSearch()
        }
      }
    })
  }

  // 单独处理课程属性字段的子组件重置
  if (props.courseAttributeList && props.courseAttributeList.length > 0) {
    props.courseAttributeList.forEach((item) => {
      const childRef = childRefs.value[`courseAttribute_${item.id}`]
      if (childRef && childRef.resetSearch) {
        childRef.resetSearch()
      }
    })
  }

  // 重置员工搜索状态
  searchResultOptions.value = []
  searchFilterPagination.value.current = 1
  searchFilterFinished.value = false
  searchFilterLoading.value = false
  searchFilterSearchKey.value = ''
}, 300)
// 移除单个条件
function removeCondition(type, id) {
  console.log(type, id)

  // 处理自定义搜索字段
  if (type.startsWith('customSearch_')) {
    const itemId = type.split('_')[1]
    // searchInputVals.value[itemId] = ''
    // 移除子组件input框内的值
    const childRef = childRefs.value[`customSearchInput_${itemId}`]
    if (childRef && childRef.resetSearch) {
      childRef.resetSearch()
    }
    emit('update:customSearchInputFilter', null, false, itemId, 'clear')
    return
  }

  // 处理课程属性字段
  if (type.startsWith('courseAttribute_')) {
    const itemId = type.split('_')[1]
    searchInputVals.value[itemId] = ''
    // 移除子组件选择框内的值
    const childRef = childRefs.value[`courseAttribute_${itemId}`]
    if (childRef && childRef.resetSearch) {
      childRef.resetSearch()
    }
    emit('update:courseAttributeFilter', {
      itemId: Number(itemId),
      value: undefined,
    }, false, itemId, type)
    return
  }

  switch (type) {
    case 'recommended':
      // recommendedVals.value = [];
      emit('update:recommendedFilter', undefined, false, id, type)
      break
    case 'age':
      // resetAge();
      // selectedAgeRange.value = null;
      emit('update:ageFilter', [], false, id, type)
      break
    case 'teacherSelect':
      selectTeacher.value = undefined
      break
    case 'oneToOneSearch':
      searchKeyOneToOne.value = undefined
      break
    case 'stuPhoneSearch':
      // searchKeyStuPhone.value = undefined;
      emit('update:stuPhoneSearchFilter', undefined, false, id, type)
      break
    case 'stuPhoneSearchNew':
      searchKeyStuPhone.value = undefined
      emit('update:stuPhoneSearchFilter', undefined, false, id, type)
      break
    case 'selectInputKeySearch':
      // selectInputKey.value = undefined
      emit('update:stuPhoneSearchFilter', undefined, false, id, type)
      break
    case 'searchInputKeySearch':
      // 直接处理，避免通过watch重复触发
      // searchInputKey.value = undefined
      skipNextWatch = true // 跳过下次watch触发
      // inputValue.value = undefined
      emit('searchInputFun', undefined, id, type)
      break

    case 'intention':
      // selectedValues.value = [];
      emit('update:intentionLevelFilter', [], false, id = null, type)
      break
    case 'grade':
      // gradeVals.value = [];
      emit('update:gradeFilter', [], false, id, type)
      break
    case 'hasSalesPerson':
      // hasSalesPersonVals.value = []
      emit('update:hasSalesPersonFilter', undefined, false, id, type)
      break
    case 'followStatus':
      // followStatusVals.value = [];
      emit('update:followStatusFilter', [], false, id, type)
      break
    case 'sex':
      // sexVals.value = [];
      emit('update:sexFilter', [], false, id, type)
      break
    case 'isCommonCourse':
      // isCommonCourseVals.value = [];
      emit('update:isCommonCourseFilter', [], false, id, type)
      break
    case 'followMethod':
      // followMethodVals.value = [];
      emit('update:followMethodFilter', [], false, id, type)
      break
    case 'visitStatus':
      // visitStatusVals.value = [];
      emit('update:visitStatusFilter', [], false, id, type)
      break
    case 'payrollStatus':
      // payrollStatusVals.value = [];
      emit('update:payrollStatusFilter', [], false, id, type)
      break
    case 'stuStatus':
      // stuStatusVals.value = [];
      emit('update:stuStatusFilter', [], false, id, type)
      break
    case 'channelStatus':
      // selectChannelVals.value = [];
      emit('update:channelStatusFilter', undefined, false, id, type)
      break
    case 'channelCategory':
      // selectChannelTreeVals.value = [];
      // selectChannelValsFilter.value = [];
      emit('update:channelFilter', undefined, false, id, type)
      break
    case 'channelType':
      // selectChannelTypeVals.value = [];
      emit('update:channelTypeFilter', undefined, false, id, type)
      break
    case 'channelCategoryType':
      // selectChannelCategoryVals.value = [];
      emit('update:channelCategoryFilter', undefined, false, id, type)
      break
    case 'positionRole':
      // positionRoleVals.value = []
      emit('update:channelPositionRoleFilter', undefined, false, id, type)
      break
    case 'quick':
      // const filter = quickFilters.value.find((q) => q.id === id);
      // if (filter) filter.selected = false;
      // 与父组件通讯，告诉父组件id，父组件根据id清除对应的筛选
      emit('update:quickFilter', undefined, false, id, type)
      break
    case 'notFollowDays':
      // resetNotFollowDays();
      // notFollowDaysSelected.value = null;
      emit('update:notFollowDaysFilter', undefined, false, id, type)
      break
    case 'approveNumber':
      emit('update:approveNumberFilter', undefined, false, id, type)
      break
    case 'orderNumber':
      emit('update:orderNumberFilter', undefined, false, id, type)
      break
    case 'validityPeriod':
      // 有效期至：只通知父组件清除筛选条件，本地选中值在 clearQuickFilter 中、接口成功后再清
      emit('update:validityPeriodFilter', [], false, id, type)
      break
    case 'classTeacher':
      // 班主任：只通知父组件清除筛选条件，本地选中值在 clearQuickFilter 中、接口成功后再清
      emit('update:classTeacherFilter', undefined, false, id, type)
      break
    case 'wxChat':
      // wxChatVals.value = null;
      // if (childRefs.value['wxChat'] && childRefs.value['wxChat'].closeDropdown) {
      //   childRefs.value['wxChat'].resetSearch();
      // }
      emit('update:wxChatFilter', undefined, false, id, type)
      break
    case 'school':
      // schoolVals.value = null;
      // if (childRefs.value['school'] && childRefs.value['school'].closeDropdown) {
      //   childRefs.value['school'].resetSearch();
      // }
      emit('update:schoolFilter', undefined, false, id, type)
      break
    case 'address':
      // addressVals.value = null;
      // if (childRefs.value['address'] && childRefs.value['address'].closeDropdown) {
      //   childRefs.value['address'].resetSearch();
      // }
      emit('update:addressFilter', undefined, false, id, type)
      break
    case 'hobbies':
      // hobbiesVals.value = null;
      // if (childRefs.value['hobbies'] && childRefs.value['hobbies'].closeDropdown) {
      //   childRefs.value['hobbies'].resetSearch();
      // }
      emit('update:hobbiesFilter', undefined, false, id, type)
      break
    case 'channelStatus':
      // 先记录以前的值
      const prevChannelStatusLength = selectChannelVals.value.length

      // 清空值
      selectChannelVals.value = []

      // 只有当之前确实有值的时候才触发事件
      if (prevChannelStatusLength > 0) {
        emit('update:channelStatusFilter', undefined)
      }
      break
    case 'channelType':
      // 先记录以前的值
      const prevLength = selectChannelTypeVals.value.length

      // 清空值
      selectChannelTypeVals.value = []

      // 只有当之前确实有值的时候才触发事件，避免不必要的请求
      if (prevLength > 0) {
        emit('update:channelTypeFilter', undefined)
      }
      break
    case 'channelCategoryType':
      // 先记录以前的值
      const prevCategoriesLength = selectChannelCategoryVals.value.length

      // 清空值
      selectChannelCategoryVals.value = []

      // 只有当之前确实有值的时候才触发事件
      if (prevCategoriesLength > 0) {
        emit('update:channelCategoryFilter', undefined)
      }
      break
    case 'quickOneToOne':
      const filterOneToOne = quickOneToOneFilters.value.find(
        q => q.id === id,
      )
      if (filterOneToOne)
        filterOneToOne.selected = false
      break
    case 'quickApprove':
      const filterApprove = quickApproveFilters.value.find(q => q.id === id)
      if (filterApprove)
        filterApprove.selected = false
      break
    case 'createUser': // 新增创建人移除逻辑
      // createUserVals.value = null
      emit('update:createUserFilter', undefined, false, id, type)
      break
    case 'teachingMethod': // 新增授课方式移除逻辑
      // teachingMethodVals.value = null;
      emit('update:teachingMethodFilter', undefined, false, id, type)
      break
    case 'salesPerson': // 销售员移除逻辑
      // selectSalesPersonVals.value = null
      emit('update:salesPersonFilter', undefined, false, id, type)
      break
    case 'createTime': // 新增创建时间移除逻辑
      // 当type为noDelCreateTime时，不清除创建时间
      if (props.type !== 'noDelCreateTime') {
        // createTimeVals.value = []
        emit('update:createTimeFilter', [], false, id, type)
      }
      break
    case 'lastEditedTime': // 最近修改时间移除逻辑
      // lastEditedTimeVals.value = []
      // console.log('单个清楚')
      emit('update:lastEditedTimeFilter', [], false, id, type)
      break
    case 'assignTime': // 分配时间移除逻辑
      // assignTimeVals.value = [];
      emit('update:assignTimeFilter', [], false, id, type)
      break
    case 'lastFollowTime': // 最近跟进时间移除逻辑
      // lastFollowTimeVals.value = [];
      emit('update:lastFollowTimeFilter', [], false, id, type)
      break
    case 'nextFollowTime': // 下次跟进时间移除逻辑
      // nextFollowTimeVals.value = [];
      emit('update:nextFollowTimeFilter', [], false, id, type)
      break
    case 'applyTime': // 新增申请时间移除逻辑
      applyTimeVals.value = []
      break
    case 'finishTime':
      emit('update:finishTimeFilter', [], false, id, type)
      break
    case 'payTime': // 新增支付时间移除逻辑
      payTimeVals.value = []
      break
    case 'classEndingTime': // 新增创建时间移除逻辑
      classEndingTimeVals.value = []
      break
    case 'classStopTime': // 新增创建时间移除逻辑
      classStopTimeVals.value = []
      break
    case 'intentionCourse': // 新增意向课程移除逻辑
      // selectCourseValues.value = null
      emit('update:intentionCourseFilter', undefined, false, id, type)
      break
    case 'recommend': // 新增推荐人移除逻辑
      // recommendedVals.value = [];
      emit('update:tuiJianUserFilter', undefined, false, id, type)
      break
    case 'subject': // 新增科目移除逻辑
      selectSubjectVals.value = null
      break
    case 'courseCategory': // 课程类别移除逻辑
      // selectCourseCategoryVals.value = null;
      emit('update:courseCategoryFilter', undefined, false, id, type)
      break
    case 'studentStatus':
      // 重置选中值
      selectStudentStatusVals.value = null
      // 重置子组件
      const studentStatusRef = childRefs.value['studentStatus']
      if (studentStatusRef && studentStatusRef.resetSearch) {
        studentStatusRef.resetSearch()
      }
      // 发出事件通知父组件
      emit('update:stuStatusFilter', undefined, false, id, type)
      break
    case 'accountStatus': // 账号状态移除逻辑
      // selectAccountStatusVals.value = null
      emit('update:channelAccountStatus', undefined, false, id, type)
      break
    case 'userType':
      // selectUserTypeVals.value = null
      emit('update:channelUserType', undefined, false, id, type)
      break
    case 'currentStatus':
      // 清空当前状态多选
      selectCurrentStatusVals.value = []
      emit('update:currentStatusFilter', [], false, id, type)
      break
    case 'orNotFenClass':
      // 清空分班状态多选
      selectOrNotFenClassVals.value = []
      emit('update:orNotFenClassFilter', [], false, id, type)
      break
    case 'billingMode':
      // 清空收费方式多选
      selectBillingModeVals.value = []
      emit('update:billingModeFilter', [], false, id, type)
      break
    case 'isSetExpirationDate':
      // 清空是否设置有效期
      selectIsSetExpirationDateVals.value = null
      emit('update:isSetExpirationDateFilter', undefined, false, id, type)
      break
    case 'openClassStatus':
      selectOpenClassStatusVals.value = null
      break
    case 'doYouSchedule':
      selectDoYouScheduleVals.value = null
      break
    case 'remaining':
      // 清空剩余数量区间
      resetRemaining()
      selectedRemainingRange.value = null
      emit('update:remainingFilter', null, false, id, type)
      break
    case 'birthday': // 添加生日移除逻辑
      // birthdayVals.value = [];
      emit('update:birthdayFilter', [], false, id, type)
      break
    case 'trialPurchaseStatus':
      trialPurchaseStatusVals.value = []
      break
    case 'followTime': // 跟进时间移除逻辑
      // followTimeVals.value = [];
      emit('update:followTimeFilter', [], false, id, type)
      break
    case 'teachingMethod': // 新增授课方式移除逻辑
      teachingMethodVals.value = null
      emit('update:teachingMethodFilter', undefined, false, id, type)
      break
    case 'isArrears':
      // 是否欠费：只通知父组件清除筛选条件，本地选中值在 clearQuickFilter 中、接口成功后再清
      emit('update:isArrearsFilter', undefined, false, id, type)
      break
    case 'lastClassTime':
      // 最近上课时间：只通知父组件清除筛选条件，本地选中值在 clearQuickFilter 中、接口成功后再清
      emit('update:lastClassTimeFilter', [], false, id, type)
      break
    case 'sellStatus': // 新增售卖状态移除逻辑
      sellStatusVals.value = null
      emit('update:sellStatusFilter', undefined, false, id, type)
      break
    case 'hasTrialPrice': // 新增是否有体验价移除逻辑
      // hasTrialPriceVals.value = null;
      emit('update:hasTrialPriceFilter', undefined, false, id, type)
      break
    case 'isMicroSchoolSale': // 新增是否开启微校售卖移除逻辑
      // isMicroSchoolSaleVals.value = null;
      emit('update:isMicroSchoolSaleFilter', undefined, false, id, type)
      break
    case 'isMicroSchoolDisplay': // 新增是否开启微校展示移除逻辑
      // isMicroSchoolDisplayVals.value = null;
      emit('update:isMicroSchoolDisplayFilter', undefined, false, id, type)
      break
    case 'performanceAllocationStatus':
      performanceAllocationStatusVals.value = null
      emit('update:performanceAllocationStatusFilter', undefined, false, id, type)
      break
    case 'orderType':
      orderTypeVals.value = []
      emit('update:orderTypeFilter', undefined, false, id, type)
      break
    case 'orderTag':
      orderTagVals.value = []
      emit('update:orderTagFilter', undefined, false, id, type)
      break
    case 'enrollType':
      enrollTypeVals.value = []
      emit('update:enrollTypeFilter', undefined, false, id, type)
      break
    case 'productType':
      productTypeVals.value = []
      emit('update:productTypeFilter', undefined, false, id, type)
      break
    case 'approvalStatus':
      approvalStatusVals.value = []
      emit('update:approvalStatusFilter', undefined, false, id, type)
      break
    case 'orderStatus':
      orderStatusVals.value = []
      emit('update:orderStatusFilter', undefined, false, id, type)
      break
    case 'orderArrearStatus':
      orderArrearStatusVals.value = []
      emit('update:orderArrearStatusFilter', undefined, false, id, type)
      break
    case 'enableStatus':
      enableStatusVals.value = null
      emit('update:enableStatusFilter', undefined, false, id, type)
      break
    case 'orderSource':
      orderSourceVals.value = []
      emit('update:orderSourceFilter', undefined, false, id, type)
      break
    case 'dealDate':
      emit('update:dealDateFilter', [], false, id, type)
      break
    case 'latestPaidTime':
      emit('update:latestPaidTimeFilter', [], false, id, type)
      break
    case 'handleContent':
      emit('update:handleContentFilter', [], false, id, type)
      break
    case 'enrolledCourse':
      // 只通知父组件清除报读课程筛选，真正清空本地选中值放在 clearQuickFilter 里，等接口成功后再处理
      emit('update:enrolledCourseFilter', [], false, id, type)
      break
    default:

      break
  }
}
// 封装清除快捷筛选
function clearQuickFilter(id, type) {
  console.log(id, type);
  switch (type) {
    case 'clear':
      // 清除已选条件
      searchInputVals.value[id] = ''
      break
    case 'quick':
      const filter = quickFilters.value.find(q => q.id === id)
      if (filter)
        filter.selected = false
      break
    case 'intention':
      selectedValues.value = []
      break
    case 'grade':
      gradeVals.value = []
      break
    case 'createTime':
      // 当type为noDelCreateTime时，不清除创建时间
      if (props.type !== 'noDelCreateTime') {
        createTimeVals.value = []
      }
      break
    case 'createUser':
      createUserVals.value = null
      selectedCreateUserInfo.value = null
      createUserOptions.value = []
      pagination.value.current = 1
      finished.value = false
      if (childRefs.value.createUser?.resetSearch) {
        childRefs.value.createUser.resetSearch()
      }
      break
    case 'followTime':
      followTimeVals.value = []
      break
    case 'birthday':
      birthdayVals.value = []
      break
    case 'lastFollowTime':
      lastFollowTimeVals.value = []
      break
    case 'nextFollowTime':
      nextFollowTimeVals.value = []
      break
    case 'assignTime':
      assignTimeVals.value = []
      break
    case 'followStatus':
      followStatusVals.value = []
      break
    case 'age':
      resetAge()
      selectedAgeRange.value = null
      break
    case 'channelCategoryType':
      selectChannelCategoryVals.value = []
      break
    case 'positionRole':
      positionRoleVals.value = []
      break
    case 'followMethod':
      followMethodVals.value = []
      break
    case 'visitStatus':
      visitStatusVals.value = []
      break
    case 'payrollStatus':
      payrollStatusVals.value = []
      break
    case 'stuStatus':
      stuStatusVals.value = []
      break
    case 'age':
      resetAge()
      selectedAgeRange.value = null
      break
    case 'channelCategoryType':
      selectChannelCategoryVals.value = []
      break
    case 'channelType':
      selectChannelTypeVals.value = []
      break
    case 'channelStatus':
      selectChannelVals.value = []
      break
    case 'channelCategory':
      selectChannelTreeVals.value = []
      selectChannelValsFilter.value = []
      break
    case 'notFollowDays':
      resetNotFollowDays()
      notFollowDaysSelected.value = null
      break
    case 'validityPeriod':
      // 清空“有效期至”时间范围的本地选中值（等待父组件接口成功后调用）
      validityPeriodVals.value = []
      break
    case 'isArrears':
      // 清空“是否欠费”的本地选中值（等待父组件接口成功后调用）
      isArrearsVals.value = null
      break
    case 'orderTag':
      orderTagVals.value = []
      break
    case 'enrollType':
      enrollTypeVals.value = []
      break
    case 'productType':
      productTypeVals.value = []
      break
    case 'orderArrearStatus':
      orderArrearStatusVals.value = []
      break
    case 'lastClassTime':
      // 清空“最近上课时间”本地选中值（等待父组件接口成功后调用）
      lastClassTimeVals.value = []
      break
    case 'classTeacher':
      // 清空班主任本地选中值（等待父组件接口成功后调用）
      classTeacherVals.value = null
      selectedClassTeacherInfo.value = null
      classTeacherOptions.value = []
      classTeacherPagination.value.current = 1
      classTeacherFinished.value = false
      if (childRefs.value.classTeacher?.resetSearch) {
        childRefs.value.classTeacher.resetSearch()
      }
      break
    case 'stuPhoneSearch':
      stuPhoneSearchVals.value = null
      selectedStuPhoneSearchInfo.value = null
      stuPhoneSearchOptions.value = []
      stuPhoneSearchPagination.value.current = 1
      stuPhoneSearchFinished.value = false
      searchKeyStuPhone.value = undefined
      if (childRefs.value.stuPhoneSearch && childRefs.value.stuPhoneSearch.closeDropdown) {
        childRefs.value.stuPhoneSearch.resetSearch()
      }
      break
    case 'stuPhoneSearchNew':
      searchKeyStuPhone.value = undefined
      stuPhoneSearchOptions.value = []
      stuPhoneSearchPagination.value.current = 1
      stuPhoneSearchFinished.value = false
      if (childRefs.value.stuPhoneSearch && childRefs.value.stuPhoneSearch.closeDropdown) {
        childRefs.value.stuPhoneSearch.resetSearch()
      }
      break
    case 'wxChat':
      wxChatVals.value = null
      if (childRefs.value.wxChat && childRefs.value.wxChat.closeDropdown) {
        childRefs.value.wxChat.resetSearch()
      }
      break
    case 'school':
      schoolVals.value = null
      if (childRefs.value.school && childRefs.value.school.closeDropdown) {
        childRefs.value.school.resetSearch()
      }
      break
    case 'orderNumber':
      orderNumberVals.value = ''
      if (childRefs.value.orderNumber && childRefs.value.orderNumber.closeDropdown) {
        childRefs.value.orderNumber.resetSearch()
      }
      break
    case 'approveNumber':
      approveNumberVals.value = ''
      if (childRefs.value.approveNumber && childRefs.value.approveNumber.closeDropdown) {
        childRefs.value.approveNumber.resetSearch()
      }
      break
    case 'address':
      addressVals.value = null
      if (childRefs.value.address && childRefs.value.address.closeDropdown) {
        childRefs.value.address.resetSearch()
      }
      break
    case 'recommend':
      recommendVals.value = null
      selectedRecommendInfo.value = null
      recommendOptions.value = []
      recommendPagination.value.current = 1
      recommendFinished.value = false
      if (childRefs.value.recommend && childRefs.value.recommend.closeDropdown) {
        childRefs.value.recommend.resetSearch()
      }
      break
    case 'createUser':
      createUserVals.value = null
      selectedCreateUserInfo.value = null
      if (childRefs.value.createUser && childRefs.value.createUser.closeDropdown) {
        childRefs.value.createUser.resetSearch()
      }
      break
    case 'salesPerson':
      selectSalesPersonVals.value = null
      selectedSalesPersonInfo.value = null
      salesPersonOptions.value = []
      salesPersonPagination.value.current = 1
      salesPersonFinished.value = false
      if (childRefs.value.salesPerson && childRefs.value.salesPerson.closeDropdown) {
        childRefs.value.salesPerson.resetSearch()
      }
      break
    case 'hobbies':
      hobbiesVals.value = null
      if (childRefs.value.hobbies && childRefs.value.hobbies.closeDropdown) {
        childRefs.value.hobbies.resetSearch()
      }
      break
    case 'recommended':
      recommendedVals.value = []
      break
    case 'hasSalesPerson':
      hasSalesPersonVals.value = []
      break
    case 'courseCategory': // 课程类别移除逻辑
      selectCourseCategoryVals.value = null
      break
    case 'intentionCourse': // 意向课程移除逻辑
      selectCourseValues.value = null
      selectedCourseInfo.value = null
      courseListOptions.value = []
      courseListPagination.value.current = 1
      courseListFinished.value = false
      if (childRefs.value.yiXiangcourse && childRefs.value.yiXiangcourse.closeDropdown) {
        childRefs.value.yiXiangcourse.resetSearch()
      }
      break
    case 'isCommonCourse':
      isCommonCourseVals.value = []
      break
    case 'sex':
      sexVals.value = []
      break
    case 'followMethod':
      followMethodVals.value = []
      break
    case 'teachingMethod': // 新增授课方式移除逻辑
      teachingMethodVals.value = null
      break
    case 'billingMode':
      selectBillingModeVals.value = []
      break
    case 'hasTrialPrice': // 新增是否有体验价移除逻辑
      hasTrialPriceVals.value = null
      break
    case 'isMicroSchoolSale': // 新增是否开启微校售卖移除逻辑
      isMicroSchoolSaleVals.value = null
      break
    case 'isMicroSchoolDisplay': // 新增是否开启微校展示移除逻辑
      isMicroSchoolDisplayVals.value = null
      break
    case 'performanceAllocationStatus':
      performanceAllocationStatusVals.value = null
      break
    case 'orderType':
      orderTypeVals.value = []
      break
    case 'approvalStatus':
      approvalStatusVals.value = []
      break
    case 'orderStatus':
      orderStatusVals.value = []
      break
    case 'orderSource':
      orderSourceVals.value = []
      break
    case 'dealDate':
      dealDateVals.value = []
      break
    case 'finishTime':
      finishTimeVals.value = []
      break
    case 'latestPaidTime':
      latestPaidTimeVals.value = []
      break
    case 'handleContent':
      enrolledCourseVals.value = []
      break
    case 'lastEditedTime': // 最近修改时间移除逻辑
      lastEditedTimeVals.value = []
      break
    case 'accountStatus': // 账号状态移除逻辑
      selectAccountStatusVals.value = null
      break
    case 'userType':
      selectUserTypeVals.value = null
      break
    case 'positionRole':
      positionRoleVals.value = []
      break
    case 'selectInputKeySearch':
      selectInputKey.value = undefined
      // 重置所有搜索相关状态后重新查询第一页
      searchResultOptions.value = []
      searchFilterPagination.value.current = 1
      searchFilterPagination.value.total = 0
      searchFilterFinished.value = false
      searchFilterLoading.value = false
      searchFilterSearchKey.value = ''
      // getStaffSearchData({ searchKey: '' })
      break
    case 'searchInputKeySearch':
      searchInputKey.value = undefined
      skipNextWatch = true // 跳过下次watch触发
      inputValue.value = undefined
      break
    case 'enrolledCourse':
      enrolledCourseVals.value = []
      break
    default:
      break
  }
}

// 展开下拉菜单的回调函数
function dropdownVisibleChangeFun(type, event) {
  if (!event || searchKeyStuPhone.value || stuPhoneSearchFinished.value)
    return
  stuPhoneSearchPagination.value.current = 1
  stuPhoneSearchOptions.value = []
  if (type == 'isShowSearchStuPhone') {
    currentSearchStudentStatus.value = 0
    getStuPhoneSearchPage({ studentStatus: 0 })
  }
  else if (type == 'isShowSearchStuPhonefilter') {
    currentSearchStudentStatus.value = 1
    getStuPhoneSearchPage({ studentStatus: 1 })
  }
}
// 创建防抖的搜索函数
const debouncedSearch = debounce((value, studentStatus = 0) => {
  console.log(`搜索文本框的值 ${value}`)
  stuPhoneSearchPagination.value.current = 1
  stuPhoneSearchFinished.value = false
  spinning.value = true
  stuPhoneSearchOptions.value = []
  getStuPhoneSearchPage({ key: value, studentStatus })
}, 300) // 300ms 的防抖延迟

// 当前搜索的学员状态
const currentSearchStudentStatus = ref(0)

// 文本框变化触发
function handleSearchStuPhone(value) {
  debouncedSearch(value, currentSearchStudentStatus.value)
}
// 添加加载状态标记
const isLoading = ref(false)
function handlePopupScroll(event) {
  const { target } = event
  const { scrollTop, scrollHeight, clientHeight } = target
  // 判断是否滚动到底部
  if (scrollHeight - scrollTop - clientHeight < 1) {
    console.log('滚动到底部了')
    // 检查是否正在加载且还有更多数据
    if (!isLoading.value && stuPhoneSearchPagination.value.current * stuPhoneSearchPagination.value.pageSize < stuPhoneSearchPagination.value.total) {
      isLoading.value = true
      stuPhoneSearchPagination.value.current += 1
      getStuPhoneSearchPage({ studentStatus: currentSearchStudentStatus.value })
    }
  }
}

// 组件卸载时取消防抖
onUnmounted(() => {
  debouncedSearch.cancel()
  debouncedSearchStaff.cancel()
})

// 员工搜索防抖函数
const debouncedSearchStaff = debounce((value) => {
  searchFilterSearchKey.value = value
  searchFilterPagination.value.current = 1
  searchResultOptions.value = []
  searchFilterFinished.value = false
  getStaffSearchData({ searchKey: value })
}, 300)

// 获取员工搜索数据的方法 - 需要父组件提供
async function getStaffSearchData(params = { searchKey: '' }) {
  if (searchFilterFinished.value || searchFilterLoading.value) return

  searchFilterLoading.value = true
  try {
    // 通过emit触发父组件提供搜索数据
    emit('staff-search', {
      ...params,
      pageRequestModel: {
        needTotal: true,
        pageSize: searchFilterPagination.value.pageSize,
        pageIndex: searchFilterPagination.value.current,
        skipCount: 1,
      }
    })
  } catch (error) {
    console.error('获取员工搜索数据失败:', error)
    // 只有在出错时才重置 loading 状态
    searchFilterLoading.value = false
  }
  // 不在 finally 中重置 loading，让 updateStaffSearchData 来控制
}

// 更新员工搜索数据的方法 - 供父组件调用
function updateStaffSearchData(data) {
  // 处理搜索结果数据格式，保持与计算属性一致
  const formattedResults = (data.result || []).map((item) => {
    return {
      id: item.id,
      value: item.nickName || item.roleName,
    }
  })

  if (searchFilterPagination.value.current === 1) {
    searchResultOptions.value = formattedResults
  } else {
    searchResultOptions.value = [...searchResultOptions.value, ...formattedResults]
  }

  searchFilterPagination.value.total = data.total || 0

  if (searchResultOptions.value.length >= searchFilterPagination.value.total) {
    searchFilterFinished.value = true
  }

  searchFilterLoading.value = false
}

// 员工搜索输入处理
function handleStaffSearch(value) {
  debouncedSearchStaff(value)
}

// 员工搜索滚动处理
function handleStaffPopupScroll(event) {
  const { target } = event
  const { scrollTop, scrollHeight, clientHeight } = target

  if (scrollHeight - scrollTop - clientHeight < 1) {
    if (!searchFilterLoading.value && searchFilterPagination.value.current * searchFilterPagination.value.pageSize < searchFilterPagination.value.total) {
      searchFilterPagination.value.current += 1
      getStaffSearchData({ searchKey: searchFilterSearchKey.value })
    }
  }
}

// 员工选择变化处理
function handleStaffChange(value) {
  emit('update:stuPhoneSearchFilter', value)

  if (value) {
    // 有选中值时，重置搜索状态，但保留已选中的值
    searchFilterFinished.value = false
  } else {
    // 清除选择时，重置所有搜索相关状态
    searchResultOptions.value = []
    searchFilterPagination.value.current = 1
    searchFilterPagination.value.total = 0
    searchFilterFinished.value = false
    searchFilterLoading.value = false
    searchFilterSearchKey.value = ''
  }
}

// 员工下拉可见状态变化处理
function handleStaffDropdownVisible(visible) {
  if (visible && searchResultOptions.value.length === 0 && !searchFilterLoading.value) {
    // 下拉展开时如果没有数据且不在加载中，则开始加载
    searchFilterPagination.value.current = 1
    searchFilterFinished.value = false
    getStaffSearchData({ searchKey: '' })
  }
}





function handleChange(value) {
  // console.log(`selected 学生id ${value}`);
  finished.value = false
  emit('update:stuPhoneSearchFilter', value)
}

// 学员/电话搜索选择变化处理函数 - 新增缓存机制
function handleStuPhoneSearchChange(value) {
  // 当 a-select 使用 labelInValue 时，value 形如 { value: studentId, label: stuName }
  const id = typeof value === 'object' && value !== null ? value.value : value

  // 缓存已选学员/电话的详细信息
  if (id) {
    const selectedUser = stuPhoneSearchOptions.value.find(user => user.id === id)
    selectedStuPhoneSearchInfo.value = selectedUser || null
  }
  else {
    selectedStuPhoneSearchInfo.value = null
  }

  searchKeyStuPhone.value = id
  nextTick(() => {
    console.log('学员/电话:', id)
    emit('update:stuPhoneSearchFilter', id)
  })
}

async function fetchOrderTagOptions() {
  if (!props.displayArray.includes('orderTag')) {
    return
  }
  try {
    const res = await getOrderTagListPagedApi({
      queryModel: { enable: true },
      sortModel: {},
      pageRequestModel: {
        pageSize: 50,
        pageIndex: 1,
        needTotal: true,
        skipCount: 0,
      },
    })
    if (res.code === 200) {
      orderTagOptions.value = Array.isArray(res.result?.list)
        ? res.result.list.map(item => ({
            id: item.id,
            value: item.name,
          }))
        : []
      return
    }
    console.error('获取订单标签失败:', res.message)
  }
  catch (error) {
    console.error('获取订单标签失败:', error)
  }
}

// 组件挂载时初始化数据
onMounted(() => {
  // 不再在这里直接调用 getDepartmentList()
  // 改为通过 watch 监听来按需调用
  fetchOrderTagOptions()
})

// 监听用户权限变化，重新处理部门数据
watch(userDeptIds, (newDeptIds) => {
  if (newDeptIds && originalDptListOptions.value.length > 0) {
    dptListOptions.value = processDepartmentPermissions(originalDptListOptions.value, newDeptIds)
    nextTick(() => {
      dptName.value = getNameById(selectDptVals.value) ?? ''
    })
  }
}, { deep: true })

// 使用 watchEffect 监听 props 变化并自动更新相应的值
watchEffect(() => {
  // 先同步父组件传入的部门 id（不能用 if(props.selectDptVals)，0 也是合法 id）
  if (props.selectDptVals !== undefined && props.selectDptVals !== null) {
    selectDptVals.value = props.selectDptVals
  }
  dptName.value = getNameById(selectDptVals.value) ?? ''

  if (props.defaultStudentStatus) {
    selectStudentStatusVals.value = props.defaultStudentStatus
  }
  if (props.defaultAccountStatus) {
    // console.log('默认账号状态', props.defaultAccountStatus)
    selectAccountStatusVals.value = props.defaultAccountStatus
    // emit('update:channelAccountStatus', props.defaultAccountStatus)
  }

  if (props.defaultCurrentStatus) {
    selectCurrentStatusVals.value = props.defaultCurrentStatus
  }

  if (props.defaultOrNotFenClass) {
    selectOrNotFenClassVals.value = props.defaultOrNotFenClass
  }

  if (props.defaultCreateTimeVals) {
    createTimeVals.value = props.defaultCreateTimeVals
  }

  if (props.defaultOrderStatusVals?.length) {
    orderStatusVals.value = props.defaultOrderStatusVals
  }

  if (props.defaultOrderTagIds?.length) {
    orderTagVals.value = props.defaultOrderTagIds
  }

  if (props.defaultOrderArrearStatusVals?.length) {
    orderArrearStatusVals.value = props.defaultOrderArrearStatusVals
  }

  if (props.defaultEnableStatus !== null && props.defaultEnableStatus !== undefined) {
    enableStatusVals.value = props.defaultEnableStatus === true ? 1 : props.defaultEnableStatus === false ? 0 : props.defaultEnableStatus
  }

  if (props.defaultOpenClassStatus) {
    selectOpenClassStatusVals.value = props.defaultOpenClassStatus
  }
})

function filterClassOrCourseOption(input, option) {
  return option.value.toLowerCase().includes(input.toLowerCase())
}
// 改变老师类型选择
function changeTeacherType() {
  selectTeacher.value = undefined
}
function handleCustomSingleSearchInputChange(e, itemId) {
  nextTick(() => {
    console.log(e)
    console.log(itemId)

    console.log('文本输入框的值', searchInputVals.value[itemId])
    emit('update:customSearchInputFilter', {
      item: e,
      value: searchInputVals.value[itemId]
    })
    // 关闭下拉
    if (childRefs.value[`customSearchInput_${itemId}`] && childRefs.value[`customSearchInput_${itemId}`].closeDropdown) {
      childRefs.value[`customSearchInput_${itemId}`].closeDropdown()
    }
  })
}

function handleHasTrialPriceChange(e) {
  nextTick(() => {
    console.log('是否有体验价:', e)
    emit('update:hasTrialPriceFilter', e)
  })
}

function handleIsMicroSchoolSaleChange(e) {
  nextTick(() => {
    console.log('是否开启微校售卖:', e)
    emit('update:isMicroSchoolSaleFilter', e)
  })
}

function handleIsMicroSchoolDisplayChange(e) {
  nextTick(() => {
    console.log('是否开启微校展示:', e)
    emit('update:isMicroSchoolDisplayFilter', e)
  })
}

// 班级名称处理函数
function handleClassNameChange() {
  nextTick(() => {
    console.log('班级名称:', classNameVals.value)
    debouncedEmit('classNameFilter', classNameVals.value)
  })
}

// 报读课程处理函数
function handleEnrolledCourseChange() {
  nextTick(() => {
    console.log('报读课程:', enrolledCourseVals.value)
    debouncedEmit('enrolledCourseFilter', enrolledCourseVals.value)
    debouncedEmit('handleContentFilter', enrolledCourseVals.value)
  })
}

// 有效期至处理函数
function handleValidityPeriodChange() {
  nextTick(() => {
    console.log('有效期至:', validityPeriodVals.value)
    emit('update:validityPeriodFilter', validityPeriodVals.value)
  })
}

// 班主任处理函数
function handleClassTeacherChange(e) {
  if (e) {
    const selectedUser = classTeacherOptions.value.find(user => user.id === e)
    selectedClassTeacherInfo.value = selectedUser || null
  } else {
    selectedClassTeacherInfo.value = null
  }

  nextTick(() => {
    console.log('班主任:', e)
    emit('update:classTeacherFilter', e)
  })
}

// 是否欠费处理函数
function handleIsArrearsChange(e) {
  nextTick(() => {
    console.log('是否欠费:', e)
    emit('update:isArrearsFilter', e)
  })
}

function handleEnableStatusChange(e) {
  nextTick(() => {
    emit('update:enableStatusFilter', e)
  })
}

function handleOrderArrearStatusChange() {
  nextTick(() => {
    emit('update:orderArrearStatusFilter', orderArrearStatusVals.value)
  })
}

// 最近上课时间处理函数
function handleLastClassTimeChange() {
  nextTick(() => {
    console.log('最近上课时间:', lastClassTimeVals.value)
    emit('update:lastClassTimeFilter', lastClassTimeVals.value)
  })
}

// 班主任下拉框可见状态变化
async function onDropdownVisibleChangeClassTeacher(value) {
  if (value)
    return
  classTeacherPagination.value.current = 1
  classTeacherFinished.value = false
  await getClassTeacherPage()
}

// 班主任搜索
async function onSearchClassTeacherFun(value) {
  classTeacherPagination.value.current = 1
  classTeacherFinished.value = false
  await getClassTeacherPage({ key: value })
}

// 班主任加载更多
async function loadMoreClassTeacher() {
  if (!isLoading.value && classTeacherPagination.value.current * classTeacherPagination.value.pageSize < classTeacherPagination.value.total) {
    isLoading.value = true
    classTeacherPagination.value.current += 1
    await getClassTeacherPage()
  }
}

// 获取班主任数据
async function getClassTeacherPage(params = { key: undefined }) {
  try {
    if (classTeacherFinished.value) return

    childRefs.value.classTeacher?.openSpinning()
    isLoading.value = true

    const res = await getUserListApi({
      'pageRequestModel': {
        'needTotal': true,
        'pageSize': classTeacherPagination.value.pageSize,
        'pageIndex': classTeacherPagination.value.current,
        'skipCount': 0,
      },
      'queryModel': {
        'searchKey': params.key,
      },
      'sortModel': {},
    })

    if (res.code === 200) {
      let resultData = res.result || []

      if (classTeacherPagination.value.current === 1) {
        classTeacherOptions.value = resultData
      }
      else {
        classTeacherOptions.value = [...classTeacherOptions.value, ...resultData]
      }
      classTeacherPagination.value.total = res.total || 0

      if (classTeacherOptions.value.length >= classTeacherPagination.value.total) {
        classTeacherFinished.value = true
      }
    }
  }
  catch (error) {
    console.error('加载班主任数据失败:', error)
    if (classTeacherPagination.value.current > 1) {
      classTeacherPagination.value.current -= 1
    }
  }
  finally {
    childRefs.value.classTeacher?.resetSpinning()
    isLoading.value = false
  }
}

function handlePerformanceAllocationStatusChange(e) {
  nextTick(() => {
    console.log('业绩分配状态:', e)
    emit('update:performanceAllocationStatusFilter', e)
  })
}

function handleOrderTypeChange() {
  nextTick(() => {
    console.log('订单类型:', orderTypeVals.value)
    debouncedEmit('orderTypeFilter', orderTypeVals.value)
  })
}

function handleOrderTagChange() {
  nextTick(() => {
    debouncedEmit('orderTagFilter', orderTagVals.value)
  })
}

function handleEnrollTypeChange() {
  nextTick(() => {
    debouncedEmit('enrollTypeFilter', enrollTypeVals.value)
  })
}

function handleProductTypeChange() {
  nextTick(() => {
    debouncedEmit('productTypeFilter', productTypeVals.value)
  })
}

function handleOrderStatusChange() {
  nextTick(() => {
    debouncedEmit('orderStatusFilter', orderStatusVals.value)
  })
}

function handleApprovalStatusChange() {
  nextTick(() => {
    debouncedEmit('approvalStatusFilter', approvalStatusVals.value)
  })
}

function handleOrderSourceChange() {
  nextTick(() => {
    debouncedEmit('orderSourceFilter', orderSourceVals.value)
  })
}

function handleDealDateChange() {
  nextTick(() => {
    emit('update:dealDateFilter', dealDateVals.value)
  })
}

function handleLatestPaidTimeChange() {
  nextTick(() => {
    emit('update:latestPaidTimeFilter', latestPaidTimeVals.value)
  })
}

// 处理部门数据更新（v-model）
function handleDepartmentUpdate(value) {
  console.log('handleDepartmentUpdate called with:', value)
  selectDptVals.value = value
  dptName.value = getNameById(value)
}

// 处理部门切换
function handleDepartmentChange(value) {
  nextTick(() => {
    // console.log('所属部门切换:', value)
    // console.log('设置前的 selectDptVals:', selectDptVals.value)
    // 更新部门名称显示
    dptName.value = getNameById(value)
    selectDptVals.value = value
    // console.log('设置后的 selectDptVals:', selectDptVals.value);

    // 通知父组件部门已切换
    emit('update:departmentFilter', value)
  })
}

// 2. 修改部门数据获取逻辑 - 监听 displayArray 和 needDepartmentData 变化
const shouldLoadDepartmentData = computed(() => {
  return props.needDepartmentData || (props.displayArray.includes('department') && props.type === 'dpt')
})

// 监听是否需要获取部门数据
watch(shouldLoadDepartmentData, (needLoad) => {
  if (needLoad && dptListOptions.value.length === 0) {
    getDepartmentList()
  }
}, { immediate: true })

// 暴露方法供父组件调用
defineExpose({
  clearQuickFilter,
  updateStaffSearchData,
  getOrderedConditions: () => orderedConditions.value,
  // 新增：暴露部门数据获取方法
  getDepartmentList,
  // 新增：重置部门数据方法
  resetDepartmentData: () => {
    dptListOptions.value = []
    originalDptListOptions.value = []
  }
})
</script>

<template>
  <div class="home flex">
    <div class="flex-1 mr-0">
      <!-- 快捷筛选区域 -->
      <div class="flex flex-wrap">
        <div v-if="isQuickShow" class="filter-section mb-2 flex-1">
          <span class="section-title mt-0.5">快捷筛选：</span>
          <div class="quick-filters">
            <a-button v-for="filter in visibleQuickFilters" :key="filter.id"
              :type="filter.selected ? 'primary' : 'default'" class="filter-btn" @click="selectQuickFilter(filter, 1)">
              {{ filter.name }}（{{ filter.count }}）
            </a-button>
          </div>
        </div>
        <div v-if="isQuickOneToOneShow" class="filter-section mb-2 flex-1">
          <span class="section-title mt-0.5">快捷筛选：</span>
          <div class="quick-filters">
            <a-button v-for="filter in quickOneToOneFilters" :key="filter.id"
              :type="filter.selected ? 'primary' : 'default'" class="filter-btn" @click="selectQuickFilter(filter, 2)">
              {{ filter.name }}（{{ filter.count }}）
            </a-button>
          </div>
        </div>
        <div v-if="isApproveQuickShow" class="filter-section mb-2 flex-1">
          <span class="section-title mt-0.5">快捷筛选：</span>
          <div class="quick-filters">
            <a-button v-for="filter in quickApproveFilters" :key="filter.id"
              :type="filter.selected ? 'primary' : 'default'" class="filter-btn" @click="selectQuickFilter(filter, 3)">
              {{ filter.name }}（{{ filter.count }}）
            </a-button>
          </div>
        </div>
        <div v-if="isShowSearchStuPhone" class="w-100 mt--1 ml-14px mb-4px hidden-below-1100">
          <div class="selectBox flex">
            <div class="label">
              学员/电话
            </div>
            <div>
              <a-select v-model:value="searchKeyStuPhoneModel" class="searchKeyStuPhone" allow-clear show-search
                placeholder="搜索姓名/手机号" :filter-option="false" style="width: 240px" option-label-prop="label"
                :label-in-value="true"
                @change="handleStuPhoneSearchChange"
                @dropdown-visible-change="dropdownVisibleChangeFun('isShowSearchStuPhone', $event)"
                @search="handleSearchStuPhone" @popup-scroll="handlePopupScroll">
                <a-select-option v-for="(item) in stuPhoneSearchOptions" :key="item.id" :value="item.id" :data="item"
                  :label="item.stuName">
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
                <a-select-option v-if="spinning" :key="1" :value="12">
                  <a-spin class="flex justify-center" />
                </a-select-option>
                <a-select-option v-if="stuPhoneSearchOptions.length > 0 && stuPhoneSearchFinished" disabled
                  class="no-more-data-option">
                  <div class="no-more-data">
                    没有更多了
                  </div>
                </a-select-option>
              </a-select>
            </div>
          </div>
        </div>
        <div v-if="isShowOneToOne" class="w-100 mt--1 hidden-below-1100">
          <div class="selectBox flex">
            <div class="label">
              一对一
            </div>
            <div>
              <a-select v-model:value="searchKeyOneToOne" allow-clear show-search placeholder="请输入关键字"
                style="width: 240px" option-label-prop="label" @change="handleChange">
                <a-select-option v-for="item in oneToOneOptions" :key="item.id" :value="item.id" :data="item"
                  :label="`${item.name}～${item.course}`">
                  <div class="flex flex-items-center mb-1">
                    <div class="ml-2">
                      <div class="text-sm text-#666 leading-7">
                        {{ item.name }}
                      </div>
                    </div>
                    <span>~</span>
                    <div>
                      <div class="text-sm text-#666 leading-7">
                        {{ item.course }}
                      </div>
                    </div>
                  </div>
                </a-select-option>
              </a-select>
            </div>
          </div>
        </div>
      </div>
      <!-- 常规筛选条件 -->
      <div class="filter-section flex justify-between">
        <div class="flex">
          <span class="section-title mt-0.5 text-#222">筛选条件：</span>
          <div class="standard-filters">
            <!-- 将固定顺序的筛选条件替换为动态顺序 -->
            <template v-for="(filterType, index) in displayArray" :key="filterType + index">
              <!-- 意向度 -->
              <checkbox-filter v-if="filterType === 'intention'" v-model:checked-values="selectedValues"
                :options="customOptions" label="意向度" type="checkbox" @change="handleIntentionChange" />
              <!-- 年级 -->
              <checkbox-filter v-if="filterType === 'grade'" v-model:checked-values="gradeVals" :options="gradeOptions"
                label="年级" type="checkbox" @change="handleGradeChange" />
              <!-- 是否有销售员 -->
              <checkbox-filter v-if="filterType === 'hasSalesPerson'" v-model:checked-values="hasSalesPersonVals"
                :options="hasSalesPersonOptions" label="是否有销售员" type="checkbox" @change="handleHasSalesPersonChange" />

              <!-- 跟进状态 -->
              <checkbox-filter v-if="filterType === 'followStatus'" v-model:checked-values="followStatusVals"
                :options="followStatusOptions" label="跟进状态" type="checkbox" @change="handleFollowChange" />

              <!-- 性别 -->
              <checkbox-filter v-if="filterType === 'sex'" v-model:checked-values="sexVals" :options="sexOptions"
                label="性别" type="checkbox" @change="handleSexChange" />

              <!-- 是否是通用课 -->
              <checkbox-filter v-if="filterType === 'isCommonCourse'" v-model:checked-values="isCommonCourseVals"
                :options="isCommonCourseOptions" label="是否是通用课" type="checkbox" @change="handleIsCommonCourseChange" />

              <!-- 跟进方式 -->
              <checkbox-filter v-if="filterType === 'followMethod'" v-model:checked-values="followMethodVals"
                :options="followMethodOptions" label="跟进方式" type="checkbox" @change="handleFollowMethodChange" />

              <!-- 回访状态 -->
              <checkbox-filter v-if="filterType === 'visitStatus'" v-model:checked-values="visitStatusVals"
                :options="visitStatusOptions" label="回访状态" type="checkbox" @change="handleVisitStatusChange" />

              <!-- 工资条状态 -->
              <checkbox-filter v-if="filterType === 'payrollStatus'" v-model:checked-values="payrollStatusVals"
                :options="payrollStatusOptions" label="工资条状态" type="checkbox" @change="handlePayrollStatusChange" />

              <!-- 学员状态 -->
              <checkbox-filter v-if="filterType === 'stuStatus'" v-model:checked-values="stuStatusVals"
                :options="stuStatusOptions" label="学员状态" type="checkbox" @change="handleStuStatusChange" />

              <!-- 创建人 -->
              <checkbox-filter v-if="filterType === 'createUser'" :ref="(el) => handleRef(el, 'createUser')"
                v-model:checked-values="createUserVals" category="teacher" :placeholder="createUserPlaceholder"
                :options="createUserOptions" :label="createUserLabel" type="radio" :finished="finished"
                @radio-change="handleCreateUserChange" @on-dropdown-visible-change="onDropdownVisibleChangeTeacher"
                @on-search="onSearchCreateUserFun" @load-more="loadMoreCreateUser" />

              <!-- 授课方式 -->
              <checkbox-filter v-if="filterType === 'teachingMethod'" :ref="(el) => handleRef(el, 'teachingMethod')"
                v-model:checked-values="teachingMethodVals" category="noSearchRadio" placeholder="请输入授课方式"
                :options="teachingMethodOptions" label="授课方式" type="radio" @radio-change="handleTeachingMethodChange" />

              <!-- 售卖状态 -->
              <checkbox-filter v-if="filterType === 'sellStatus'" :ref="(el) => handleRef(el, 'sellStatus')"
                v-model:checked-values="sellStatusVals" category="noSearchRadio" placeholder="请选择售卖状态"
                :options="sellStatusOptions" label="售卖状态" type="radio" @radio-change="handleSellStatusChange" />

              <!-- 是否有体验价 -->
              <checkbox-filter v-if="filterType === 'hasTrialPrice'" :ref="(el) => handleRef(el, 'hasTrialPrice')"
                v-model:checked-values="hasTrialPriceVals" category="noSearchRadio" placeholder="请选择是否有体验价"
                :options="hasTrialPriceOptions" label="是否有体验价" type="radio" @radio-change="handleHasTrialPriceChange" />

              <!-- 是否开启微校售卖 -->
              <checkbox-filter v-if="filterType === 'isMicroSchoolSale'"
                :ref="(el) => handleRef(el, 'isMicroSchoolSale')" v-model:checked-values="isMicroSchoolSaleVals"
                category="noSearchRadio" placeholder="请选择是否开启微校售卖" :options="isMicroSchoolSaleOptions" label="是否开启微校售卖"
                type="radio" @radio-change="handleIsMicroSchoolSaleChange" />

              <!-- 是否开启微校展示 -->
              <checkbox-filter v-if="filterType === 'isMicroSchoolDisplay'"
                :ref="(el) => handleRef(el, 'isMicroSchoolDisplay')" v-model:checked-values="isMicroSchoolDisplayVals"
                category="noSearchRadio" placeholder="请选择是否开启微校展示" :options="isMicroSchoolDisplayOptions"
                label="是否开启微校展示" type="radio" @radio-change="handleIsMicroSchoolDisplayChange" />

              <!-- 销售员 -->
              <checkbox-filter v-if="filterType === 'salesPerson'" :ref="(el) => handleRef(el, 'salesPerson')"
                v-model:checked-values="selectSalesPersonVals" category="teacher" :placeholder="salesPersonPlaceholder"
                :options="salesPersonOptions" :label="salesPersonLabel" type="radio" :finished="salesPersonFinished"
                @radio-change="handleSalesPersonChange" @on-dropdown-visible-change="onDropdownVisibleChangeSalesPerson"
                @on-search="onSearchSalesPersonFun" @load-more="loadMoreSalesPerson" />

              <!-- 班级名称（多选） -->
              <checkbox-filter v-if="filterType === 'className'" :ref="(el) => handleRef(el, 'className')"
                v-model:checked-values="classNameVals" :options="classNameOptions" label="班级名称" 
                type="checkbox" @change="handleClassNameChange" />

              <!-- 报读课程（多选）- 复用意向课程的数据源 -->
              <checkbox-filter v-if="filterType === 'enrolledCourse'" :ref="(el) => handleRef(el, 'enrolledCourse')"
                v-model:checked-values="enrolledCourseVals" category="course" placeholder="请输入报读课程"
                :options="courseListOptions" label="报读课程" show-search type="checkbox" :finished="courseListFinished"
                @change="handleEnrolledCourseChange" @on-dropdown-visible-change="onCourseListDropdownVisibleChange"
                @on-search="onSearchCourseListFun" @load-more="loadMoreCourseList" />

              <!-- 有效期至（复用下次跟进时间组件） -->
              <checkbox-filter v-if="filterType === 'validityPeriod'" v-model:checked-values="validityPeriodVals"
                label="有效期至" type="dateTimeQuick" @date-picker-change="handleValidityPeriodChange" />

              <!-- 班主任（复用销售员组件） -->
              <checkbox-filter v-if="filterType === 'classTeacher'" :ref="(el) => handleRef(el, 'classTeacher')"
                v-model:checked-values="classTeacherVals" category="teacher" placeholder="请输入班主任"
                :options="classTeacherOptions" label="班主任" type="radio" :finished="classTeacherFinished"
                @radio-change="handleClassTeacherChange" @on-dropdown-visible-change="onDropdownVisibleChangeClassTeacher"
                @on-search="onSearchClassTeacherFun" @load-more="loadMoreClassTeacher" />

              <!-- 是否欠费（复用是否有体验价组件） -->
              <checkbox-filter v-if="filterType === 'isArrears'" :ref="(el) => handleRef(el, 'isArrears')"
                v-model:checked-values="isArrearsVals" category="noSearchRadio" placeholder="请选择是否欠费"
                :options="isArrearsOptions" label="是否欠费" type="radio" @radio-change="handleIsArrearsChange" />

              <checkbox-filter v-if="filterType === 'enableStatus'" :ref="(el) => handleRef(el, 'enableStatus')"
                v-model:checked-values="enableStatusVals" category="noSearchRadio" placeholder="请选择启用状态"
                :options="enableStatusOptions" label="启用状态" type="radio" @radio-change="handleEnableStatusChange" />

              <!-- 最近上课时间（复用下次跟进时间组件） -->
              <checkbox-filter v-if="filterType === 'lastClassTime'" v-model:checked-values="lastClassTimeVals"
                label="最近上课时间" type="dateTimeQuick" @date-picker-change="handleLastClassTimeChange" />

              <!-- 创建时间 -->
              <checkbox-filter v-if="filterType === 'createTime'" v-model:checked-values="createTimeVals"
                :label="createTimeLabel" type="dateTime" @date-picker-change="handleCreateTimeChange" />

              <!-- 跟进时间 -->
              <checkbox-filter v-if="filterType === 'followTime'" v-model:checked-values="followTimeVals" label="跟进时间"
                type="dateTimeQuick" @date-picker-change="handleFollowTimeChange" />

              <!-- 最近编辑时间 -->
              <checkbox-filter v-if="filterType === 'lastEditedTime'" v-model:checked-values="lastEditedTimeVals"
                label="最近编辑时间" type="dateTime" @date-picker-change="handlelastEditedTimeChange" />

              <!-- 生日 -->
              <checkbox-filter v-if="filterType === 'birthday'" v-model:checked-values="birthdayVals" label="生日"
                type="dateTime" @date-picker-change="handleBirthdayChange" />

              <checkbox-filter v-if="filterType === 'approveNumber'" :ref="(el) => handleRef(el, 'approveNumber')"
                v-model:checked-values="approveNumberVals" placeholder="请输入审批编号" label="审批编号" type="inputType"
                @change="handleApproveNumberChange" />

              <checkbox-filter v-if="filterType === 'orderNumber'" :ref="(el) => handleRef(el, 'orderNumber')"
                v-model:checked-values="orderNumberVals" placeholder="请输入订单编号" label="订单编号" type="inputType"
                @change="handleOrderNumberChange" />

              <!-- 微信号 -->
              <checkbox-filter v-if="filterType === 'wxChat'" :ref="(el) => handleRef(el, 'wxChat')"
                v-model:checked-values="wxChatVals" placeholder="请输入微信号" label="微信号" type="inputType"
                @change="handleWxChatChange" />

              <!-- 分配时间 -->
              <checkbox-filter v-if="filterType === 'assignTime'" v-model:checked-values="assignTimeVals" label="分配时间"
                type="dateTime" @date-picker-change="handleAssignTimeChange" />

              <!-- 最近跟进 -->
              <checkbox-filter v-if="filterType === 'lastFollowTime'" v-model:checked-values="lastFollowTimeVals"
                label="最近跟进" type="dateTime" @date-picker-change="handleLastFollowTimeChange" />

              <!-- 下次跟进 -->
              <checkbox-filter v-if="filterType === 'nextFollowTime'" v-model:checked-values="nextFollowTimeVals"
                label="下次跟进" type="dateTimeQuick" @date-picker-change="handleNextFollowTimeChange" />

              <!-- 申请时间 -->
              <checkbox-filter v-if="filterType === 'applyTime'" v-model:checked-values="applyTimeVals" label="申请时间"
                type="dateTime" @date-picker-change="handleCreateTimeChange" />

              <checkbox-filter v-if="filterType === 'finishTime'" v-model:checked-values="finishTimeVals" label="审批完成时间"
                type="dateTime" @date-picker-change="handleFinishTimeChange" />

              <!-- 支付日期 -->
              <checkbox-filter v-if="filterType === 'payTime'" v-model:checked-values="payTimeVals" label="支付日期"
                type="dateTime" @date-picker-change="handleCreateTimeChange" />

              <!-- 结课时间 -->
              <checkbox-filter v-if="filterType === 'classEndingTime'" v-model:checked-values="classEndingTimeVals"
                label="结课时间" type="dateTimeQuick" @date-picker-change="handleCreateTimeChange" />

              <!-- 停课时间 -->
              <checkbox-filter v-if="filterType === 'classStopTime'" v-model:checked-values="classStopTimeVals"
                label="停课时间" type="dateTimeQuick" @date-picker-change="handleCreateTimeChange" />

              <!-- 意向课程 -->
              <checkbox-filter v-if="filterType === 'intentionCourse'" :ref="(el) => handleRef(el, 'yiXiangcourse')"
                v-model:checked-values="selectCourseValues" category="course" placeholder="请输入意向课程"
                :options="courseListOptions" label="意向课程" type="radio" :finished="courseListFinished"
                @radio-change="handleCourseChange" @on-dropdown-visible-change="onCourseListDropdownVisibleChange"
                @on-search="onSearchCourseListFun" @load-more="loadMoreCourseList" />

              <!-- 推荐人 -->
              <checkbox-filter v-if="filterType === 'recommend'" :ref="(el) => handleRef(el, 'recommend')"
                v-model:checked-values="recommendVals" placeholder="请输入推荐人" label="推荐人" category="stu"
                :options="recommendOptions" type="radio" :finished="recommendFinished"
                @radio-change="handleTuiJianUserChange" @on-dropdown-visible-change="onRecommendDropdownVisibleChange"
                @on-search="onSearchRecommendFun" @load-more="loadMoreRecommend" />

              <!-- 学员/电话 -->
              <checkbox-filter v-if="filterType === 'stuPhoneSearch'" :ref="(el) => handleRef(el, 'stuPhoneSearch')"
                v-model:checked-values="stuPhoneSearchVals" placeholder="请输入学员/电话" label="学员/电话" category="stu"
                :options="stuPhoneSearchOptions" type="radio" :finished="stuPhoneSearchFinished"
                @radio-change="handleChange" @on-dropdown-visible-change="onStuPhoneSearchDropdownVisibleChange"
                @on-search="onSearchStuPhoneSearchFun" @load-more="loadMoreStuPhoneSearch" />
              <!-- 所属部门 -->
              <checkbox-filter v-if="filterType === 'department' && type == 'dpt'"
                v-model:checked-values="selectDptVals" :options="dptListOptions" label="所属部门" type="tree"
                @change="handleDepartmentChange" @update:checked-values="handleDepartmentUpdate" />

              <!-- 渠道树 -->
              <checkbox-filter v-if="filterType === 'channelCategory'" v-model:checked-values="selectChannelTreeVals"
                show-search :options="channelOptions" label="渠道" type="cascader" @change="handleChannelCategoryChange"
                @dropdown-visible-change="handleDropdownVisibleChange" />

              <!-- 渠道状态 -->
              <checkbox-filter v-if="filterType === 'channelStatus'" v-model:checked-values="selectChannelVals"
                :options="channelListOptions" label="渠道状态" type="checkbox" @change="handleChannelChange" />

              <!-- 渠道类型 -->
              <checkbox-filter v-if="filterType === 'channelType'" v-model:checked-values="selectChannelTypeVals"
                :options="channelTypeOptions" label="渠道类型" type="checkbox" @change="handleChannelTypeChange" />

              <!-- 科目 -->
              <checkbox-filter v-if="filterType === 'subject'" :ref="(el) => handleRef(el, 'kemu')"
                v-model:checked-values="selectSubjectVals" category="course" placeholder="请输入科目"
                :options="subjectOptions" label="科目" type="radio" @radio-change="handleSubjectChange" />

              <!-- 课程类别 -->
              <checkbox-filter v-if="filterType === 'courseCategory'" :ref="(el) => handleRef(el, 'courseType')"
                v-model:checked-values="selectCourseCategoryVals" category="course" placeholder="请搜索"
                :options="courseCategoryOptions" label="课程类别" type="radio" @radio-change="handleCourseCategoryChange"
                @on-search="onSearchCourseCategoryFun" @on-dropdown-visible-change="queryCourseCategory" />

              <!-- 学员状态 -->
              <checkbox-filter v-if="filterType === 'studentStatus'" :ref="(el) => handleRef(el, 'studentStatus')"
                v-model:checked-values="selectStudentStatusVals" category="noSearchRadio" placeholder="选择学员状态"
                :options="studentStatusOptions" label="学员状态" type="radio" @radio-change="handleStudentStatusChange" />
              <!-- 账号状态 -->
              <checkbox-filter v-if="filterType === 'accountStatus'" :ref="(el) => handleRef(el, 'accountStatus')"
                v-model:checked-values="selectAccountStatusVals" category="noSearchRadio" placeholder="选择账号状态"
                :options="accountStatusOptions" label="账号状态" type="radio" @radio-change="handleAccountStatusChange" />
              <!-- 员工类型 -->
              <checkbox-filter v-if="filterType === 'userType'" :ref="(el) => handleRef(el, 'userType')"
                v-model:checked-values="selectUserTypeVals" category="noSearchRadio" placeholder="选择员工类型"
                :options="userTypeOptions" label="员工类型" type="radio" @radio-change="handleUserTypeChange" />

              <!-- 业绩分配状态 -->
              <checkbox-filter v-if="filterType === 'performanceAllocationStatus'"
                :ref="(el) => handleRef(el, 'performanceAllocationStatus')"
                v-model:checked-values="performanceAllocationStatusVals" category="noSearchRadio" placeholder="选择业绩分配状态"
                :options="performanceAllocationStatusOptions" label="业绩分配状态" type="radio"
                @radio-change="handlePerformanceAllocationStatusChange" />

              <!-- 订单类型 -->
              <checkbox-filter v-if="filterType === 'orderType'" v-model:checked-values="orderTypeVals"
                :options="orderTypeOptions" label="订单类型" type="checkbox" @change="handleOrderTypeChange" />

              <checkbox-filter v-if="filterType === 'orderTag'" v-model:checked-values="orderTagVals"
                :options="orderTagOptions" label="订单标签" type="checkbox" @change="handleOrderTagChange" />

              <checkbox-filter v-if="filterType === 'enrollType'" v-model:checked-values="enrollTypeVals"
                :options="enrollTypeOptions" label="报读类型" type="checkbox" @change="handleEnrollTypeChange" />

              <checkbox-filter v-if="filterType === 'productType'" v-model:checked-values="productTypeVals"
                :options="productTypeOptions" label="商品类型" type="checkbox" @change="handleProductTypeChange" />

              <checkbox-filter v-if="filterType === 'approvalStatus'" v-model:checked-values="approvalStatusVals"
                :options="approvalStatusOptions" label="审批状态" type="checkbox" @change="handleApprovalStatusChange" />

              <checkbox-filter v-if="filterType === 'orderStatus'" v-model:checked-values="orderStatusVals"
                :options="orderStatusOptions" label="订单状态" type="checkbox" @change="handleOrderStatusChange" />

              <checkbox-filter v-if="filterType === 'orderArrearStatus'" v-model:checked-values="orderArrearStatusVals"
                :options="orderArrearStatusOptions" label="欠费状态" type="checkbox" @change="handleOrderArrearStatusChange" />

              <checkbox-filter v-if="filterType === 'orderSource'" v-model:checked-values="orderSourceVals"
                :options="orderSourceOptions" label="订单来源" type="checkbox" @change="handleOrderSourceChange" />

              <checkbox-filter v-if="filterType === 'dealDate'" v-model:checked-values="dealDateVals"
                label="经办日期" type="dateTime" @date-picker-change="handleDealDateChange" />

              <checkbox-filter v-if="filterType === 'latestPaidTime'" v-model:checked-values="latestPaidTimeVals"
                label="最近支付时间" type="dateTime" @date-picker-change="handleLatestPaidTimeChange" />

              <checkbox-filter v-if="filterType === 'handleContent'" :ref="(el) => handleRef(el, 'enrolledCourse')"
                v-model:checked-values="enrolledCourseVals" category="course" placeholder="请输入办理内容"
                :options="courseListOptions" label="办理内容" show-search type="checkbox" :finished="courseListFinished"
                @change="handleEnrolledCourseChange" @on-dropdown-visible-change="onCourseListDropdownVisibleChange"
                @on-search="onSearchCourseListFun" @load-more="loadMoreCourseList" />

              <!-- 当前状态 -->
              <checkbox-filter v-if="filterType === 'currentStatus'" :ref="(el) => handleRef(el, 'currentStatus')"
                v-model:checked-values="selectCurrentStatusVals" category="course" placeholder="选择状态"
                :options="currentStatusOptions" label="当前状态" type="checkbox" @change="handleCurrentStatusChange" />

              <!-- 是否分班 -->
              <checkbox-filter v-if="filterType === 'orNotFenClass'" :ref="(el) => handleRef(el, 'orNotFenClass')"
                v-model:checked-values="selectOrNotFenClassVals" :options="orNotFenClassOptions" label="是否分班"
                type="checkbox" @change="handleOrNotFenClassChange" />

              <!-- 渠道分类 -->
              <checkbox-filter v-if="filterType === 'channelCategoryType'"
                :ref="(el) => handleRef(el, 'channelCategoryType')" v-model:checked-values="selectChannelCategoryVals"
                :options="channelCategoryOptions" label="渠道分类" show-search type="checkbox"
                @change="handleChannelCategoryValueChange" />
              <!-- 任职角色 -->
              <checkbox-filter v-if="filterType === 'positionRole'" :ref="(el) => handleRef(el, 'positionRole')"
                v-model:checked-values="positionRoleVals" :options="positionRoleOptions" label="任职角色" show-search
                type="checkbox" @change="handleRoleValueChange" />

              <!-- 收费方式 -->
              <checkbox-filter v-if="filterType === 'billingMode'" :ref="(el) => handleRef(el, 'billingMode')"
                v-model:checked-values="selectBillingModeVals" :options="billingModeOptions" label="收费方式"
                type="checkbox" @change="handleBillingModeChange" />

              <!-- 是否设置有效期 -->
              <checkbox-filter v-if="filterType === 'isSetExpirationDate'"
                :ref="(el) => handleRef(el, 'isSetExpirationDate')"
                v-model:checked-values="selectIsSetExpirationDateVals" category="noSearchRadio" placeholder="请选择有效期状态"
                :options="isSetExpirationDateOptions" label="是否设置有效期" type="radio"
                @radio-change="handleIsSetExpirationDateChange" />

              <!-- 开班状态 -->
              <checkbox-filter v-if="filterType === 'openClassStatus'" :ref="(el) => handleRef(el, 'openClassStatus')"
                v-model:checked-values="selectOpenClassStatusVals" category="noSearchRadio" placeholder="请选择开班状态"
                :options="openClassStatusOptions" label="开班状态" type="radio"
                @radio-change="handleOpenClassStatusChange" />

              <!-- 是否排课 -->
              <checkbox-filter v-if="filterType === 'doYouSchedule'" :ref="(el) => handleRef(el, 'doYouSchedule')"
                v-model:checked-values="selectDoYouScheduleVals" category="noSearchRadio" placeholder="请选择排课状态"
                :options="doYouScheduleOptions" label="是否排课" type="radio" @radio-change="handleDoYouScheduleChange" />

              <!-- 未跟进天数 -->
              <checkbox-filter v-if="filterType === 'notFollowDays'" :ref="(el) => handleRef(el, 'notFollowDays')"
                type="custom" label="未跟进天数" :checked-values="notFollowDaysSelected ? [notFollowDaysSelected] : []"
                @change="handleNotFollowDaysConfirm">
                <template #custom>
                  <div class="not-follow-days-filter">
                    <div class="flex items-center mb-2">
                      <span style="white-space:nowrap;">大于</span>
                      <a-input-number v-model:value="notFollowDaysInput" :controls="false" class="w-40 mx2" :min="1"
                        :precision="0" placeholder="天数" @change="handleNotFollowDaysInput" />
                      <span>天</span>
                    </div>
                    <div class="quick-btns mb-2">
                      <a-button v-for="d in notFollowDaysQuickOptions" :key="d"
                        :type="notFollowDaysQuick === d ? 'primary' : 'default'" size="small" class="mr-2"
                        style="border-radius: 6px;" @click="selectNotFollowDaysQuick(d)">
                        {{ d }}天
                      </a-button>
                    </div>
                    <div class="flex justify-between px4 b-btn">
                      <a-button size="small" type="text" class="mr-2 w-18 text-#06f reset-btn"
                        :disabled="!notFollowDaysInput && !notFollowDaysQuick" @click="resetNotFollowDays">
                        重置
                      </a-button>
                      <a-button size="small" class="w-18" type="primary" @click="handleNotFollowDaysConfirm">
                        确定
                      </a-button>
                    </div>
                  </div>
                </template>
              </checkbox-filter>

              <!-- 年龄 -->
              <checkbox-filter v-if="filterType === 'age'" :ref="(el) => handleRef(el, 'age')" type="custom" label="年龄"
                :checked-values="selectedAgeRange ? [selectedAgeRange] : []" @change="handleAgeConfirm">
                <template #custom>
                  <div class="age-filter">
                    <div class="flex justify-between mb-2">
                      <span>最小年龄（岁）</span>
                      <span>最大年龄（岁）</span>
                    </div>
                    <div class="flex items-center mb-2">
                      <a-input-number v-model:value="minAge" :controls="false" class="w-25 mr-2" :min="0" :max="100"
                        :precision="0" placeholder="最小" />
                      <span>-</span>
                      <a-input-number v-model:value="maxAge" :controls="false" class="w-25 ml-2" :min="0" :max="100"
                        :precision="0" placeholder="最大" />
                    </div>
                    <div class="flex justify-between px4 b-btn">
                      <a-button size="small" type="text" class="mr-2 w-18 text-#06f reset-btn"
                        :disabled="minAge == null && maxAge == null" @click="resetAge">
                        重置
                      </a-button>
                      <a-button size="small" class="w-18" type="primary" @click="handleAgeConfirm">
                        确定
                      </a-button>
                    </div>
                  </div>
                </template>
              </checkbox-filter>
              <!-- 是否被推荐 -->
              <checkbox-filter v-if="filterType === 'recommended'" v-model:checked-values="recommendedVals"
                :options="recommendedOptions" label="是否被推荐" type="checkbox" @change="handleRecommendedChange" />

              <!-- 剩余数量 -->
              <checkbox-filter v-if="filterType === 'remaining'" :ref="(el) => handleRef(el, 'remaining')" type="custom"
                label="剩余数量" :checked-values="selectedRemainingRange ? [selectedRemainingRange] : []"
                @change="handleRemainingConfirm">
                <template #custom>
                  <div class="remaining-filter">
                    <a-tabs v-model:active-key="remainingMode" centered size="small">
                      <a-tab-pane key="lesson" tab="按课时">
                        <div class="px-3">
                          <div class="flex justify-between mb-1  text-3">
                            <span>最小课时数（课时）</span>
                            <span>最大课时数（课时）</span>
                          </div>
                          <div class="flex items-center mb-4">
                            <a-input-number v-model:value="minRemaining" :controls="false" class="w-28 mr-2" :min="0"
                              :precision="0" placeholder="请输入" />
                            <span>~</span>
                            <a-input-number v-model:value="maxRemaining" :controls="false" class="w-28 ml-2" :min="0"
                              :precision="0" placeholder="请输入" />
                          </div>
                        </div>
                      </a-tab-pane>
                      <a-tab-pane key="period" tab="按时段">
                        <div class="px-3">
                          <div class="flex justify-between mb-1 text-3">
                            <span>最小天数（天）</span>
                            <span>最大天数（天）</span>
                          </div>
                          <div class="flex items-center mb-4">
                            <a-input-number v-model:value="minRemaining" :controls="false" class="w-28 mr-2" :min="0"
                              :precision="0" placeholder="请输入" />
                            <span>~</span>
                            <a-input-number v-model:value="maxRemaining" :controls="false" class="w-28 ml-2" :min="0"
                              :precision="0" placeholder="请输入" />
                          </div>
                        </div>
                      </a-tab-pane>
                      <a-tab-pane key="amount" tab="按金额">
                        <div class="px-3">
                          <div class="flex justify-between mb-1 text-3">
                            <span>最小金额（元）</span>
                            <span>最大金额（元）</span>
                          </div>
                          <div class="flex items-center mb-4">
                            <a-input-number v-model:value="minRemaining" :controls="false" class="w-28 mr-2" :min="0"
                              :precision="2" placeholder="请输入" />
                            <span>~</span>
                            <a-input-number v-model:value="maxRemaining" :controls="false" class="w-28 ml-2" :min="0"
                              :precision="2" placeholder="请输入" />
                          </div>
                        </div>
                      </a-tab-pane>
                    </a-tabs>
                    <div class="flex justify-between px4 mb-1 b-btn">
                      <a-button size="small" type="text" class="mr-2 w-18 text-#06f reset-btn"
                        :disabled="minRemaining == null && maxRemaining == null" @click="resetRemaining">
                        重置
                      </a-button>
                      <a-button size="small" class="w-18" type="primary" @click="handleRemainingConfirm">
                        确定
                      </a-button>
                    </div>
                  </div>
                </template>
              </checkbox-filter>
              <!-- 自定义的单选select选项 -->
              <div v-if="filterType === 'customSearch' && customIsDisplayList.length > 0" class="flex">
                <span v-for="item in customIsDisplayList" :key="item.id">
                  <!-- 文本 数字 日期 选项 -->
                  <div v-if="item.fieldType == 1">
                    <checkbox-filter :ref="(el) => handleRef(el, `customSearchInput_${item.id}`)"
                      v-model:checked-values="searchInputVals[item.id]" :placeholder="`请输入${item.fieldKey}`"
                      :label="item.fieldKey" type="inputType"
                      @change="(e) => handleCustomSingleSearchInputChange(item, item.id)" />
                  </div>
                  <div v-if="item.fieldType == 2">
                    <checkbox-filter :ref="(el) => handleRef(el, `customSearchInput_${item.id}`)"
                      v-model:checked-values="searchInputVals[item.id]" :placeholder="`请输入${item.fieldKey}`"
                      :label="item.fieldKey" type="numberInputType"
                      @change="(e) => handleCustomSingleSearchInputChange(item, item.id)" />
                  </div>
                  <div v-if="item.fieldType == 3">
                    <checkbox-filter :ref="(el) => handleRef(el, `customSearchInput_${item.id}`)"
                      v-model:checked-values="searchInputVals[item.id]" :placeholder="`请输入${item.fieldKey}`"
                      :label="item.fieldKey" type="dateSelectType"
                      @date-picker-change="(e) => handleCustomSingleSearchInputChange(item, item.id)" />
                  </div>
                  <div v-if="item.fieldType == 4">
                    <checkbox-filter :ref="(el) => handleRef(el, `customSearchInput_${item.id}`)"
                      v-model:checked-values="searchInputVals[item.id]" :placeholder="`请输入${item.fieldKey}`"
                      :options="item.optionsList" :label="item.fieldKey" type="radioType"
                      @radio-change="handleCustomSingleSearchInputChange(item, item.id)" />
                  </div>
                </span>
              </div>

              <!-- 体验课购买状态 -->
              <checkbox-filter v-if="filterType === 'trialPurchaseStatus'"
                v-model:checked-values="trialPurchaseStatusVals" :options="trialPurchaseStatusOptions" label="体验课购买状态"
                type="checkbox" @change="handleTrialPurchaseStatusChange" />

              <!-- 就读学校 -->
              <checkbox-filter v-if="filterType === 'school'" :ref="(el) => handleRef(el, 'school')"
                v-model:checked-values="schoolVals" placeholder="请输入(支持模糊搜索)" label="就读学校" type="inputType"
                @change="handleSchoolChange" />

              <!-- 家庭住址 -->
              <checkbox-filter v-if="filterType === 'address'" :ref="(el) => handleRef(el, 'address')"
                v-model:checked-values="addressVals" placeholder="请输入(支持模糊搜索)" label="家庭住址" type="inputType"
                @change="handleAddressChange" />

              <!-- 兴趣爱好 -->
              <checkbox-filter v-if="filterType === 'hobbies'" :ref="(el) => handleRef(el, 'hobbies')"
                v-model:checked-values="hobbiesVals" placeholder="请输入(支持模糊搜索)" label="兴趣爱好" type="inputType"
                @change="handleHobbiesChange" />

              <!-- 开启的课程属性列表，下拉options 要下拉的时候触发接口查询 -->
              <div v-if="filterType === 'courseAttribute' && courseAttributeList.length > 0" class="flex">
                <span v-for="item in courseAttributeList" :key="item.id">
                  <div>
                    <checkbox-filter :id="item.id" :ref="(el) => handleRef(el, `courseAttribute_${item.id}`)"
                      v-model:checked-values="searchInputVals[item.id]" :placeholder="`请输入${item.name}`"
                      :options="courseAttributeOptions" :label="item.name" category="courseAttribute" type="radioType"
                      @on-search="filterCourseAttributeFun" @radio-change="handleCourseAttributeChange"
                      @on-dropdown-visible-change="handleCourseAttributeDropdownVisibleChange" />
                  </div>
                </span>
              </div>
            </template>
          </div>
        </div>
        <div v-if="isShowSearchInput" class="w-70 mt--1 hidden-below-1100">
          <div class="selectBox flex">
            <div class="label searchLabel">
              {{ searchLabel }}
            </div>
            <div>
              <a-input v-model:value="inputValue" class="searchInput w-165px" allow-clear
                :placeholder="searchPlaceholder">
                <template #suffix>
                  <SearchOutlined style="color: #bbb" />
                </template>
              </a-input>
            </div>
          </div>
        </div>
        <div v-if="isShowClsssOrCourseSearch" class="w-100 mt--0.5 hidden-below-1100">
          <div class="selectBox flex">
            <div class="label searchLabel">
              {{ searchLabel }}
            </div>
            <div v-if="renderClassListOptions">
              <a-select v-model:value="selectInputKey" allow-clear show-search :placeholder="searchPlaceholder"
                style="width: 200px" :filter-option="false" :loading="searchFilterLoading" @change="handleStaffChange"
                @search="handleStaffSearch" @popup-scroll="handleStaffPopupScroll"
                @dropdown-visible-change="handleStaffDropdownVisible">
                <a-select-option v-for="item in searchResultOptions" :key="item.id" :value="item.id">
                  {{ item.value }}
                </a-select-option>
                <a-select-option v-if="searchFilterLoading" key="loading" disabled>
                  <a-spin class="flex justify-center" />
                </a-select-option>
                <a-select-option v-if="searchResultOptions.length > 0 && searchFilterFinished" disabled
                  class="no-more-data-option">
                  <div class="no-more-data">
                    没有更多了
                  </div>
                </a-select-option>
              </a-select>
            </div>
            <div v-if="!renderClassListOptions">
              <a-select v-model:value="selectInputKey" allow-clear show-search :placeholder="searchPlaceholder"
                style="width: 200px" :options="courseListOptions" :field-names="{ label: 'value', value: 'id' }"
                :filter-option="filterClassOrCourseOption" @change="handleChange" />
            </div>
          </div>
        </div>
        <div v-if="isShowSearchStuPhonefilter" class="w-100 mt--1 hidden-below-1100">
          <div class="selectBox flex">
            <div class="label">
              学员/电话
            </div>
            <div>
              <a-select v-model:value="searchKeyStuPhoneModel" class="searchKeyStuPhone" allow-clear show-search
                placeholder="搜索姓名/手机号" :filter-option="false" style="width: 240px" option-label-prop="label"
                :label-in-value="true"
                @change="handleStuPhoneSearchChange"
                @dropdown-visible-change="dropdownVisibleChangeFun('isShowSearchStuPhonefilter', $event)"
                @search="handleSearchStuPhone" @popup-scroll="handlePopupScroll">
                <a-select-option v-for="(item) in stuPhoneSearchOptions" :key="item.id" :value="item.id" :data="item"
                  :label="item.stuName">
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
                <a-select-option v-if="spinning" :key="1" :value="12">
                  <a-spin class="flex justify-center" />
                </a-select-option>
                <a-select-option v-if="stuPhoneSearchOptions.length > 0 && stuPhoneSearchFinished" disabled
                  class="no-more-data-option">
                  <div class="no-more-data">
                    没有更多了
                  </div>
                </a-select-option>
              </a-select>
            </div>
          </div>
        </div>
        <div v-if="isShowChangTeacherSearch" class="w-100 mt--0.5 hidden-below-1100">
          <div>
            <a-input-group compact class="compactInput">
              <a-select v-model:value="teacherType" class="w-105px" @change="changeTeacherType">
                <a-select-option :value="1">
                  上课老师
                </a-select-option>
                <a-select-option :value="2">
                  上课助教
                </a-select-option>
              </a-select>
              <a-select v-model:value="selectTeacher" show-search class="w-50" placeholder="请输入上课老师">
                <a-select-option v-for="(item, index) in createUserOptions" :key="index" :value="item.id" :data="item">
                  {{
                    item.value }}
                </a-select-option>
              </a-select>
            </a-input-group>
          </div>
        </div>
      </div>
      <!-- 已选条件展示 -->
      <div v-if="orderedConditions.length > 0 || type == 'dpt' || type == 'noDelCreateTime'"
        class="selected-conditions">
        <span class="section-title text-#222">已选条件：</span>
        <div class="condition-tags">
          <a-popconfirm title="确定要清空所有条件吗？" @confirm="clearAll">
            <a-tag v-if="(type != 'dpt' || type != 'noDelCreateTime') || orderedConditions.length > 0" color="red"
              class="clear-all mb-2">
              清空所有
              <DeleteOutlined class="text-3 ml-4px mt-0.6px" />
            </a-tag>
          </a-popconfirm>

          <a-tag v-if="type == 'dpt'" color="blue" class="condition-tag mb-2">
            <div class="tag-content">
              <span class="condition-label">所属部门：</span>
              <div class="condition-values">
                <span class="value-item">
                  {{ dptName }}
                </span>
              </div>
            </div>
          </a-tag>
          <a-tag v-for="condition in orderedConditions" :key="condition.type" color="blue" class="condition-tag mb-2">
            <div class="tag-content">
              <span class="condition-label">{{ condition.label }}：</span>
              <div class="condition-values">
                <span v-for="(value, index) in condition.values" :key="value.id" class="value-item ">
                  {{ value.value ?? value.name }}
                  <CloseOutlined v-if="index === condition.values.length - 1" class="close-icon"
                    @click.stop="removeCondition(condition.type, value.id)" />
                  <span v-else class="separator">、</span>
                </span>
              </div>
            </div>
          </a-tag>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.remaining-filter :deep(.ant-tabs-nav) {
  margin-bottom: 12px;
}

:deep(.ant-tabs-ink-bar) {
  text-align: center;
  height: 9px !important;
  background: transparent;
  bottom: 1px !important;
}

:deep(.ant-tabs-ink-bar::after) {
  position: absolute;
  top: 3px;
  left: calc(50% - 12px);
  width: 24px !important;
  height: 4px !important;
  border-radius: 2px;
  background-color: var(--pro-ant-color-primary);
  content: "";
}

.age-filter {
  display: inline-block;
  padding: 8px;
  border-radius: 6px;
  min-width: 100px;
  color: #666;
}

.not-follow-days-filter {
  display: inline-block;
  padding: 8px;
  border-radius: 6px;
  min-width: 210px;
  color: #666;
}

.quick-btns .ant-btn {
  min-width: 48px;
}

.b-btn {
  border-top: 1px solid #e9e9e9;
  padding-top: 10px;
}

.b-btn .ant-btn-text:not(:disabled):hover {
  color: #06f;
  background: transparent;
}

.selectBox {
  justify-content: flex-end;
  align-items: center;
}

.searchInput {
  border-top-left-radius: 0;
  border-bottom-left-radius: 0;
}

.selectBox .label {
  border: 1px solid #f0f0f0;
  height: 32px;
  padding: 0 10px;
  line-height: 32px;
  text-align: left;
  width: 100px;
  border-radius: 8px 0 0 8px !important;
  color: #222;
  font-size: 14px;
  min-width: 104px;
  padding-left: 8px;
  padding-right: 16px;
  border-right: 0;
  text-align: center;
}

.selectBox .searchLabel {
  border-right: 0 !important;
}

:deep(.selectBox .ant-select-selector) {
  border-radius: 0 6px 6px 0 !important;
}

.home {
  padding: 12px 12px 6px 12px;
  background: #ffffff;
  border-radius: 8px;
  align-items: flex-start;
  width: 100%;
}

.debug-panel {
  padding: 16px;
  margin-bottom: 24px;
  background: #f8f9fa;
  border-radius: 6px;
}

.filter-section {
  display: flex;
  align-items: flex-start;
}

.section-title {
  white-space: nowrap;
}

.quick-filters {
  display: flex;
  gap: 8px;
}

.standard-filters {
  display: flex;
  flex-wrap: wrap;
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

.filter-btn {
  height: 28px;
  padding: 0 12px;
}

.no-more-data {
  text-align: center;
  color: #999;
  font-size: 12px;
  padding: 8px 0;
}

.no-more-data-option {
  cursor: default !important;
}

:deep(.compactInput.ant-input-group.ant-input-group-compact) {
  display: flex !important;
  justify-content: flex-end !important;
}

/* 添加响应式隐藏样式 */
@media screen and (max-width: 1110px) {
  .hidden-below-1100 {
    display: none !important;
  }
}
</style>
