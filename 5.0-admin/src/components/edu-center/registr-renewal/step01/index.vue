<script setup>
import { computed, nextTick, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import { message, Modal } from 'ant-design-vue'
import { InfoCircleOutlined, FormOutlined } from '@ant-design/icons-vue'
import dayjs from 'dayjs'
import { debounce } from 'lodash-es'
import { useRoute } from 'vue-router'
import ActiveCourseModal from './active-course-modal.vue'
import UpdateDateModal from './update-date-modal.vue'
import CouponInputModal from './coupon-input-modal.vue'
import CreateStudent from '@/components/common/create-student.vue'
import StaffSelect from '@/components/common/staff-select.vue'
import StudentSelect from '@/components/common/student-select.vue'
import { getRechargeAccountByStudentApi } from '@/api/finance-center/recharge-account'
import { useUserStore } from '@/stores/user'
import { addIntendedStudentApi, getRecommenderPageApi } from '~@/api/enroll-center/intention-student'
import { getOrderTagListPagedApi } from '@/api/finance-center/order-tag'
import { getCalcCourseEnrollTypeApi, getCheckQuoteInfoApi, postCreateOrderApi, postPayOrderApi } from '~@/api/edu-center/registr-renewal'
import messageService from '~@/utils/messageService'

// 获取路由信息
const route = useRoute()
const userStore = useUserStore()
const currentInstUserId = computed(() => userStore.instUserId)

// Props
const props = defineProps({
  formState: {
    type: Object,
    required: true,
  },
  handleOver: {
    type: Boolean,
    default: false,
  },
})

// Emits
const emit = defineEmits([
  'update:formState',
  'update:handleOver',
  'submit-order',
])

const openSelect = ref(false)
const activeCourseOver = ref(false)
const activeCourseModalRef = ref(null)

// 学生选择组件引用
const studentSelectRef = ref(null)

// 获取当前日期（格式：YYYY-MM-DD）
const getCurrentDate = () => dayjs().format('YYYY-MM-DD')
const settingForm = reactive({
  orderDiscountType: '0',
  zdiscountNumber: undefined,
  zdiscountRate: undefined,
  dealDate: getCurrentDate(),
  salePerson: currentInstUserId.value || undefined,
  orderLabel: undefined,
  internalRemark: '',
  externalRemark: '',
  useBalance: 0, // 使用充值余额
  useResidualBalance: 0, // 使用残联余额
  useGiftBalance: 0, // 使用赠送余额
})

// 初始化表单数据
const formData = reactive([])





const orderLabelOptions = ref([])

async function fetchOrderTagOptions() {
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
      orderLabelOptions.value = Array.isArray(res.result?.list) ? res.result.list : []
      return
    }
    messageService.error(res.message || '获取订单标签失败')
  }
  catch (error) {
    console.error('获取订单标签失败:', error)
    messageService.error('获取订单标签失败')
  }
}

function orderLabelFilterOption(input, option) {
  const keyword = input.toLowerCase()
  const nameMatch = option.name.toLowerCase().includes(keyword)
  return nameMatch
}

function handleCourseModalConfirm(selectedCourses) {
  openSelect.value = false
  activeCourseOver.value = true
  // console.log('Selected courses:', selectedCourses)
  // 处理 selectedCourses 中的 productSku ,格式如下
  //     {
  //       label: '单次报价单',
  //       options: [
  //         {
  //           value: '1',
  //           label: '1课时｜100元',
  //         },
  //       ],
  // }

  // 获取新选择的课程ID集合
  const selectedCourseIds = new Set(selectedCourses.map(course => course.id))

  // 保留仍然被选中的已有课程（保持其配置）
  const keepCourses = formData.filter(course => selectedCourseIds.has(course.id))

  // 找出新增的课程
  const existingCourseIds = new Set(formData.map(course => course.id))
  const newCourses = selectedCourses.filter(course => !existingCourseIds.has(course.id))

  // 处理新增课程的默认配置
  newCourses.forEach((item) => {
    // 添加表单默认值
    item.handleType = 0 // 默认为"无"
    item.productSkuId = undefined // 默认未选择报价单
    item.skuCount = 1 // 默认购买1份
    item.freeQuantity = 0 // 默认赠送0课时
    item.discountType = '0' // 默认无优惠
    item.discountNumber = undefined // 优惠金额
    item.discountRate = undefined // 优惠折扣 优惠折扣*10
    item.validDate = '' // 有效期
    item.endDate = '' // 结束时间（时段模式使用）
    item.endDateWeek = '' // 结束时间对应的周几
    item.totalDays = 0 // 总天数（含赠）
    item.error = false // 报价单错误状态

    item.priceList = []
    item.productSku.forEach((sku) => {
      // 根据 unit 生成 label
      const generateLabel = (sku) => {
        const { unit, quantity, price } = sku
        switch (unit) {
          case 1:
            return `${quantity}课时｜${price}元`
          case 2:
            return `${quantity}天｜${price}元`
          case 3:
            return `${quantity}月｜${price}元`
          case 4:
            return `${quantity}年｜${price}元`
          case 5:
            return `${price}元`
          default:
            return `${price}元`
        }
      }

      item.priceList.push({
        label: sku.name,
        options: [
          {
            value: sku.id,
            label: generateLabel(sku),
            lessonAudition: sku.lessonAudition,
            lessonModel: sku.lessonModel,
            unit: sku.unit,
            quantity: sku.quantity,
            price: sku.price
          },
        ],
      })
    })

    // 对 priceList 进行排序：按课时-时段-金额类型排序
    item.priceList.sort((a, b) => {
      const aOption = a.options[0]
      const bOption = b.options[0]

      // 定义unit的优先级：课时(1) > 天(2) > 月(3) > 年(4) > 其他(5)
      const unitPriority = { 1: 1, 2: 2, 3: 3, 4: 4, 5: 5 }
      const aUnitPriority = unitPriority[aOption.unit] || 6
      const bUnitPriority = unitPriority[bOption.unit] || 6

      // 首先按unit类型排序
      if (aUnitPriority !== bUnitPriority) {
        return aUnitPriority - bUnitPriority
      }

      // 同一unit类型内，按quantity排序
      if (aOption.quantity !== bOption.quantity) {
        return (aOption.quantity || 0) - (bOption.quantity || 0)
      }

      // 相同quantity，按价格排序
      return (bOption.price || 0) - (aOption.price || 0)
    })
  })

  // 重新构建formData：保留已配置的课程 + 新增课程
  formData.splice(0, formData.length, ...keepCourses, ...newCourses)
}

// 当前选中的学员信息
const currentSelectedStudent = ref(null)

function handleStudentChange(value) {
  // 更新formState中的学员ID
  if (!value) {
    currentSelectedStudent.value = null
    settingForm.useBalance = 0
    settingForm.useResidualBalance = 0
    settingForm.useGiftBalance = 0
  }
  emit('update:formState', { ...props.formState, studentId: value, selectedStudentInfo: value ? props.formState.selectedStudentInfo : undefined })
}

async function enrichStudentRechargeInfo(studentInfo) {
  if (!studentInfo?.id) {
    return studentInfo
  }
  try {
    const { result: rechargeAccount } = await getRechargeAccountByStudentApi({ studentId: studentInfo.id })
    const residualBalance = Number(rechargeAccount?.residualBalance || 0)
    const accountBalance = Number(rechargeAccount?.balance || 0)
    return {
      ...studentInfo,
      rechargeAccountId: rechargeAccount?.id || '',
      rechargeAccountName: rechargeAccount?.accountName || rechargeAccount?.rechargeAccountName || '',
      rechargeBalance: Math.max(0, accountBalance - residualBalance),
      residualBalance,
      giftBalance: Number(rechargeAccount?.givingBalance || 0),
    }
  }
  catch (error) {
    console.error('加载学员储值余额失败:', error)
    return {
      ...studentInfo,
      rechargeAccountId: '',
      rechargeAccountName: '',
      rechargeBalance: 0,
      residualBalance: 0,
      giftBalance: 0,
    }
  }
}

async function handleStudentSelect(studentInfo) {
  // 保存选中的学员信息
  const enrichedStudent = await enrichStudentRechargeInfo(studentInfo)
  currentSelectedStudent.value = enrichedStudent
  settingForm.useBalance = 0
  settingForm.useResidualBalance = 0
  settingForm.useGiftBalance = 0
  emit('update:formState', { ...props.formState, studentId: enrichedStudent?.id, selectedStudentInfo: enrichedStudent })
  console.log('选中的学员信息:', enrichedStudent)
}



// 检查路由参数并加载学生信息
async function checkRouteAndLoadStudent() {
  const studentId = route.query.id || route.params.id
  if (studentId && studentSelectRef.value) {
    try {
      // 使用StudentSelect组件的方法加载学生信息
      const student = await studentSelectRef.value.loadStudentById(studentId)
      if (student) {
        const enrichedStudent = await enrichStudentRechargeInfo(student)
        // 保存学员信息
        currentSelectedStudent.value = enrichedStudent
        // 更新formState中的studentId
        emit('update:formState', { ...props.formState, studentId: enrichedStudent.id, selectedStudentInfo: enrichedStudent })
        console.log('通过路由参数加载的学员信息:', enrichedStudent)
      } else {
        console.warn('未找到指定ID的学员信息:', studentId)
      }
    } catch (error) {
      console.error('加载路由指定学员失败:', error)
    }
  }
}

// 组件挂载时检查路由参数
onMounted(() => {
  // 延迟执行以确保 studentSelectRef 已经准备好
  nextTick(() => {
    checkRouteAndLoadStudent()
  })
  fetchOrderTagOptions()
})

// 优惠券弹窗状态
const openCouponModal = ref(false)

// 打开优惠券弹窗
function handleOpenCouponModal() {
  openCouponModal.value = true
}

// 处理优惠券确认事件
function handleCouponConfirm(couponCode) {
  // 这里可以添加验证逻辑
  console.log('输入的优惠券码:', couponCode)
  // 关闭弹窗在子组件中处理
}

// 关闭优惠券弹窗
function handleCloseCouponModal() {
  openCouponModal.value = false
}

// 提取验证错误信息的通用函数
function extractValidationErrors(error) {
  let errorMessages = []

  if (error && typeof error === 'object') {
    // 处理 Ant Design Vue 表单验证错误
    if (error.values && error.errorFields) {
      error.errorFields.forEach(field => {
        if (field.errors && field.errors.length > 0) {
          field.errors.forEach(err => {
            errorMessages.push(err)
          })
        }
      })
    }
    // 处理单个表单验证错误（没有 values 属性的情况）
    else if (error.errorFields && Array.isArray(error.errorFields)) {
      error.errorFields.forEach(field => {
        if (field.errors && field.errors.length > 0) {
          field.errors.forEach(err => {
            errorMessages.push(err)
          })
        }
      })
    }
    // 处理普通错误对象
    else if (error.message) {
      errorMessages.push(error.message)
    }
    // 处理错误数组
    else if (Array.isArray(error)) {
      error.forEach(err => {
        if (err && typeof err === 'object') {
          if (err.message) {
            errorMessages.push(err.message)
          } else if (err.errors && Array.isArray(err.errors)) {
            err.errors.forEach(e => errorMessages.push(e))
          }
        } else if (typeof err === 'string') {
          errorMessages.push(err)
        }
      })
    }
    // 处理嵌套的错误对象
    else if (error.error && error.error.message) {
      errorMessages.push(error.error.message)
    }
  }

  // 如果错误是字符串类型
  if (typeof error === 'string') {
    errorMessages.push(error)
  }

  return errorMessages
}

// 显示验证错误信息
function showValidationErrors(errorMessages, defaultMessage = '验证失败，请检查填写内容') {
  // 如果没有提取到具体错误信息，显示默认信息
  if (errorMessages.length === 0) {
    errorMessages.push(defaultMessage)
  }

  // 显示错误信息（最多显示前3条）
  const displayMessages = errorMessages.slice(0, 3)
  displayMessages.forEach((msg, index) => {
    setTimeout(() => {
      messageService.error(msg)
    }, index * 200) // 错开显示时间，避免重叠
  })

  if (errorMessages.length > 3) {
    setTimeout(() => {
      messageService.error(`还有 ${errorMessages.length - 3} 个验证错误...`)
    }, 600)
  }
}

// 组件卸载时取消防抖
onUnmounted(() => {
  debouncedValidateAll.cancel()
  debouncedSubmitOrder.cancel()
})

// 获取当前选中的学员信息
const selectedStudent = computed(() => {
  if (!props.formState.studentId) return null

  // 优先使用保存的学员信息
  if (currentSelectedStudent.value && currentSelectedStudent.value.id === props.formState.studentId) {
    return currentSelectedStudent.value
  }

  if (props.formState.selectedStudentInfo && props.formState.selectedStudentInfo.id === props.formState.studentId) {
    return props.formState.selectedStudentInfo
  }

  // 如果没有保存的信息，尝试从组件获取
  if (studentSelectRef.value) {
    return studentSelectRef.value.selectedStudent
  }

  return null
})

// 计算单个课程的合计金额
function calculateCourseTotal(courseItem) {
  if (!courseItem.productSkuId || !courseItem.skuCount)
    return 0

  // 找到选中的报价单价格
  let selectedPrice = 0

  // 直接从 productSku 中获取实际价格
  const sku = courseItem.productSku?.find(sku => sku.id === courseItem.productSkuId)
  selectedPrice = sku ? sku.price : 0

  if (selectedPrice === 0)
    return 0

  // 计算小计：价格 * 购买份数
  let subtotal = selectedPrice * (courseItem.skuCount || 0)

  // 应用优惠
  if (courseItem.discountType === '1' && courseItem.discountNumber) {
    // 金额优惠
    subtotal -= courseItem.discountNumber
  }
  else if (courseItem.discountType === '2' && courseItem.discountRate) {
    // 折扣优惠 (discountRate是0.1-9.9，需要除以10得到实际折扣)
    subtotal = subtotal * (courseItem.discountRate / 10)
  }

  return Math.max(0, subtotal) // 确保不为负数
}

// 计算总合计金额
const discountedTotalAmount = computed(() => {
  // 先计算所有课程的小计总和
  const subtotal = formData.reduce((total, item) => {
    return total + calculateCourseTotal(item)
  }, 0)

  // 应用整单优惠
  let finalAmount = subtotal
  if (settingForm.orderDiscountType === '1' && settingForm.zdiscountNumber) {
    // 整单金额优惠
    finalAmount -= settingForm.zdiscountNumber
  }
  else if (settingForm.orderDiscountType === '2' && settingForm.zdiscountRate) {
    // 整单折扣优惠
    finalAmount = finalAmount * (settingForm.zdiscountRate / 10)
  }

  return Math.max(0, finalAmount) // 确保不为负数
})

const totalStoragePayment = computed(() => {
  return Number(settingForm.useBalance || 0) + Number(settingForm.useResidualBalance || 0) + Number(settingForm.useGiftBalance || 0)
})

const totalAmount = computed(() => {
  return Math.max(0, discountedTotalAmount.value - totalStoragePayment.value)
})

const availableRechargeBalance = computed(() => Number(selectedStudent.value?.rechargeBalance || 0))
const availableResidualBalance = computed(() => Number(selectedStudent.value?.residualBalance || 0))
const availableGiftBalance = computed(() => Number(selectedStudent.value?.giftBalance || 0))

function getUsageMax(currentKey) {
  const baseAmount = discountedTotalAmount.value
  const otherUsage = [
    currentKey === 'useBalance' ? 0 : Number(settingForm.useBalance || 0),
    currentKey === 'useResidualBalance' ? 0 : Number(settingForm.useResidualBalance || 0),
    currentKey === 'useGiftBalance' ? 0 : Number(settingForm.useGiftBalance || 0),
  ].reduce((sum, value) => sum + value, 0)

  const remaining = Math.max(0, baseAmount - otherUsage)
  const availableMap = {
    useBalance: availableRechargeBalance.value,
    useResidualBalance: availableResidualBalance.value,
    useGiftBalance: availableGiftBalance.value,
  }
  return Math.max(0, Math.min(availableMap[currentKey] || 0, remaining))
}

function normalizeUsageValue(currentKey) {
  const max = getUsageMax(currentKey)
  const currentValue = Number(settingForm[currentKey] || 0)
  if (currentValue > max) {
    settingForm[currentKey] = Number(max.toFixed(2))
  } else if (currentValue < 0) {
    settingForm[currentKey] = 0
  }
}

function applyFullUsage(currentKey) {
  settingForm[currentKey] = Number(getUsageMax(currentKey).toFixed(2))
}

watch(
  () => [
    discountedTotalAmount.value,
    availableRechargeBalance.value,
    availableResidualBalance.value,
    availableGiftBalance.value,
    settingForm.useBalance,
    settingForm.useResidualBalance,
    settingForm.useGiftBalance,
  ],
  () => {
    normalizeUsageValue('useBalance')
    normalizeUsageValue('useResidualBalance')
    normalizeUsageValue('useGiftBalance')
  },
  { deep: true },
)

// 创建响应式的课程合计计算
const courseTotals = computed(() => {
  return formData.map(item => calculateCourseTotal(item))
})

// 获取指定索引的课程合计
function getCourseTotal(index) {
  return courseTotals.value[index] || 0
}

// 计算包含整单优惠分摊的课程最终合计金额
const courseFinalTotals = computed(() => {
  return formData.map((item, index) => {
    const courseTotal = calculateCourseTotal(item)
    const distributedDiscount = getCourseDistributedDiscount(index)
    return Math.max(0, courseTotal - distributedDiscount)
  })
})

// 获取指定索引的课程最终合计（包含整单优惠分摊）
function getCourseFinalTotal(index) {
  return courseFinalTotals.value[index] || 0
}

// 计算整单优惠金额
const totalDiscountAmount = computed(() => {
  const subtotal = formData.reduce((total, item) => {
    return total + calculateCourseTotal(item)
  }, 0)

  let discountAmount = 0
  if (settingForm.orderDiscountType === '1' && settingForm.zdiscountNumber) {
    // 整单金额优惠
    discountAmount = settingForm.zdiscountNumber
  }
  else if (settingForm.orderDiscountType === '2' && settingForm.zdiscountRate) {
    // 整单折扣优惠
    discountAmount = subtotal * (1 - settingForm.zdiscountRate / 10)
  }

  return Math.max(0, discountAmount)
})

// 计算当前设置的折扣减免金额（实时显示）
const currentDiscountAmount = computed(() => {
  if (settingForm.orderDiscountType === '2' && settingForm.zdiscountRate) {
    const subtotal = formData.reduce((total, item) => {
      return total + calculateCourseTotal(item)
    }, 0)
    return subtotal * (1 - settingForm.zdiscountRate / 10)
  }
  return 0
})

// 计算当前设置的金额优惠（实时显示）
const currentAmountDiscount = computed(() => {
  if (settingForm.orderDiscountType === '1' && settingForm.zdiscountNumber) {
    return settingForm.zdiscountNumber
  }
  return 0
})

// 计算单个课程的分摊优惠金额
function getCourseDistributedDiscount(index) {
  if (formData.length === 0 || totalDiscountAmount.value === 0)
    return 0

  const courseTotal = getCourseTotal(index)
  const totalBeforeDiscount = formData.reduce((total, item) => {
    return total + calculateCourseTotal(item)
  }, 0)

  if (totalBeforeDiscount === 0)
    return 0

  // 按比例分摊整单优惠
  return (courseTotal / totalBeforeDiscount) * totalDiscountAmount.value
}

// 取消选择课程
function cancelSelectCourse(index) {
  // 获取要删除的课程信息，用于在modal中取消选中
  const courseToRemove = formData[index]

  // 从表单数据中移除
  formData.splice(index, 1)

  // 如果没有课程了，重置状态
  if (formData.length === 0) {
    activeCourseOver.value = false
  }

  // 通知ActiveCourseModal组件取消选中该课程
  if (activeCourseModalRef.value) {
    activeCourseModalRef.value.cancelCourseSelection(courseToRemove)
  }
}

// 处理日期更新
function handleDateChange(dateObj) {
  if (!dateObj) {
    settingForm.dealDate = getCurrentDate()
  }
}

// 禁止选择今天之后的日期
function disabledDate(current) {
  return current > dayjs().endOf('day')
}

// 禁止选择今天之前的日期（用于有效期）
function disabledValidDate(current) {
  return current < dayjs().startOf('day')
}

// 开始时间不禁用任何日期（允许选择之前的日期）
function disabledStartDate(current) {
  return false
}

// 表单引用
const formRefs = ref([])
const settingOrderForm = ref(null)

// 校验规则
const validateMessages = {
  required: '${label}不能为空',
  types: {
    number: '请输入有效的数字',
  },
}

const rules = {
  productSkuId: [{ required: true, message: '请选择报价单' }],
}

// 动态校验规则
function numRules(index) {
  return [
    {
      validator: async () => {
        const total = formData[index].skuCount + formData[index].freeQuantity
        if (total <= 0) {
          return Promise.reject(new Error('购买数+赠送数的总和不可为0'))
        }
        return Promise.resolve()
      },
    },
  ]
}

function giftNumRules(index) {
  return [
    {
      validator: async () => {
        const total = formData[index].skuCount + formData[index].freeQuantity
        if (total <= 0) {
          return Promise.reject(new Error('购买数+赠送数的总和不可为0'))
        }
        return Promise.resolve()
      },
    },
  ]
}

function discountNumberRules(index) {
  return [
    {
      required: formData[index].discountType === '1',
      message: '请输入优惠金额',
      type: 'number',
      min: 0.01,
    },
  ]
}

function discountRateRules(index) {
  return [
    {
      required: formData[index].discountType === '2',
      message: '请输入折扣（0.1-9.9）',
      type: 'number',
      min: 0.1,
      max: 10,
    },
  ]
}

function zdiscountNumberRules() {
  return [
    {
      required: settingForm.orderDiscountType === '1',
      message: '请输入优惠金额',
    },
    {
      type: 'number',
      min: 0.01,
      message: '优惠金额必须大于0.01',
      trigger: 'blur',
    },
  ]
}

function zdiscountRateRules() {
  return [
    {
      required: settingForm.orderDiscountType === '2',
      message: '请输入折扣（0.1-9.9）',
    },
    {
      type: 'number',
      min: 0.1,
      max: 9.9,
      message: '折扣必须在0.1-9.9之间',
      trigger: 'blur',
    },
  ]
}

// 保存按钮loading状态
const saveLoading = ref(false)

// 提交订单loading状态
const submitOrderLoading = ref(false)

// 提交时统一校验
async function validateAll() {
  if (saveLoading.value) return // 防止重复点击

  saveLoading.value = true
  // 清除之前的错误状态
  clearAllErrors()

  try {
    await Promise.all(
      formRefs.value.map((formRef) => {
        return formRef.validate()
      }),
    )
    // console.log('所有表单验证通过')

    // 如果未选择开始时间  先弹窗提示
    const hasTimePeriodCourse = formData.some(course => shouldShowTimePeriod(course))
    const hasUnsetStartTime = formData.some(course =>
      shouldShowTimePeriod(course) && (!course.validDate || course.validDate === '')
    )

    if (hasTimePeriodCourse && hasUnsetStartTime) {
      return new Promise((resolve) => {
        Modal.confirm({
          title: '暂未设置"开始时间"?',
          content: '若想下次设置"开始时间"，可至"订单模块—待处理"中设置',
          okText: '暂不设置',
          centered: true,
          cancelText: '返回编辑',
          onOk() {
            // 暂不设置，继续执行
            emit('update:handleOver', true, formData)
            resolve()
          },
          onCancel() {
            // 返回编辑，隐藏弹窗，终止逻辑
            resolve()
          }
        })
      })
    }
    const quoteDetailList = formData.map(course => {
      // 从productSku中找到选中的报价单
      const selectedSku = course.productSku?.find(sku => sku.id === course.productSkuId)
      // console.log(selectedSku);

      return {
        courseId: course.id,
        quoteId: course.productSkuId,
        price: selectedSku?.price || 0,
        quantity: selectedSku?.quantity,
        lessonModel: selectedSku?.lessonModel || 1,
      }
    })

    // 检测报价单信息
    const result = await getCheckQuoteInfo({ quoteDetailList })
    if (!result.error) {
      console.log(formData);

      emit('update:handleOver', true, formData)
    } else {
      // console.log(result.result)

      // 标记出错的课程
      if (result.result && Array.isArray(result.result)) {
        result.result.forEach(errorItem => {
          if (errorItem.error === 1) {
            // 找到对应的课程并标记错误
            const courseIndex = formData.findIndex(course => course.id === errorItem.courseId)
            if (courseIndex !== -1) {
              formData[courseIndex].error = true
            }
          }
        })
      }

      messageService.error('报价单失效，请检查报价单信息')
    }

  }
  catch (error) {
    emit('update:handleOver', false)
    console.log('验证失败:', error)

    // 提取并显示具体的验证错误信息
    const errorMessages = extractValidationErrors(error)
    showValidationErrors(errorMessages, '表单验证失败，请检查填写内容')
  }
  finally {
    saveLoading.value = false
  }
}

// 创建防抖的保存函数
const debouncedValidateAll = debounce(validateAll, 300, {
  leading: true,
  trailing: true,
})

// 创建防抖的提交订单函数
const debouncedSubmitOrder = debounce(submitOrder, 300, {
  leading: true,
  trailing: false,
})

// 处理取消选择课程（用于错误状态下的重新选择）
function handleCancel() {
  // 找出所有有错误的课程
  const errorCourses = formData.filter(course => course.error === true)

  // 通知ActiveCourseModal组件取消这些课程的选中状态
  if (activeCourseModalRef.value && errorCourses.length > 0) {
    errorCourses.forEach(errorCourse => {
      activeCourseModalRef.value.cancelCourseSelection(errorCourse)
    })
  }

  // 只删除有错误的课程数据
  const validCourses = formData.filter(course => course.error !== true)

  // 清空原数组并重新填充有效课程
  formData.splice(0, formData.length, ...validCourses)

  // 如果没有课程了，重置状态
  if (formData.length === 0) {
    activeCourseOver.value = false
  }

  emit('update:handleOver', false)

  // 打开选择课程modal弹窗
  openSelect.value = true
}

// 清除所有课程的错误状态
function clearAllErrors() {
  formData.forEach(course => {
    course.error = false
  })
}
async function getCheckQuoteInfo(quoteDetailList) {
  try {
    const res = await getCheckQuoteInfoApi(quoteDetailList)
    if (res.code === 200) {
      if (res.result.length == 0) {
        return {
          error: false
        }
      } else {
        return {
          error: true,
          result: res.result,
        }
      }
    } else {
      messageService.error(res.message)
    }
  } catch (error) {
    // console.log('验证失败:', error)
  }
}

// 重置订单设置表单
function resetSettingForm() {
  settingForm.orderDiscountType = '0'
  settingForm.zdiscountNumber = undefined
  settingForm.zdiscountRate = undefined
  settingForm.dealDate = getCurrentDate()
  settingForm.salePerson = currentInstUserId.value || undefined
  settingForm.orderLabel = undefined
  settingForm.internalRemark = ''
  settingForm.externalRemark = ''
  settingForm.useBalance = 0
  settingForm.useResidualBalance = 0
  settingForm.useGiftBalance = 0
}

watch(currentInstUserId, (value) => {
  if (value && !settingForm.salePerson) {
    settingForm.salePerson = value
  }
}, { immediate: true })

function handleEdit() {
  resetSettingForm() // 重置订单设置表单
  clearAllErrors() // 清除所有错误状态
  emit('update:handleOver', false)
}

function confirm(e) {
  const newFormState = { ...props.formState, studentId: undefined }
  newFormState.selectedStudentInfo = undefined
  emit('update:formState', newFormState)

  // 清空当前选中的学员信息
  currentSelectedStudent.value = null
  settingForm.useBalance = 0
  settingForm.useResidualBalance = 0
  settingForm.useGiftBalance = 0

  // 重置学生选择组件
  if (studentSelectRef.value) {
    studentSelectRef.value.reset()
  }
}

function handleSelect(e) {
  if (e === 1) {
    // 检查是否已选择学员
    if (!props.formState.studentId) {
      messageService.error("请先选择办理学员")
      return
    }
    openSelect.value = true
  }
}

// 按班级选择的处理函数
function handleClassSelect() {
  // 检查是否已选择学员
  if (!props.formState.studentId) {
    message.error('请先选择办理学员')
    return
  }
  // TODO: 这里添加按班级选择的逻辑
  message.info('按班级选择功能暂未实现')
}

async function submitOrder() {
  if (submitOrderLoading.value) return // 防止重复点击

  if (!props.formState.studentId) {
    message.error('请选择学员')
    return
  }

  submitOrderLoading.value = true

  try {
    await settingOrderForm.value.validate()

    // 构造订单数据
    const orderData = {
      // 学员信息
      studentInfo: {
        id: props.formState.studentId,
        ...selectedStudent.value,
      },
      // 课程数据
      courseData: formData.map(course => ({
        ...course,
        courseTotal: calculateCourseTotal(course),
        distributedDiscount: getCourseDistributedDiscount(formData.indexOf(course)),
      })),
      // 订单设置
      orderSettings: {
        ...settingForm,
      },
      // 金额信息
      amountInfo: {
        totalAmount: (totalAmount.value).toFixed(2),
        subtotal: formData.reduce((total, item) => total + calculateCourseTotal(item), 0),
        totalDiscountAmount: totalDiscountAmount.value,
        currentDiscountAmount: currentDiscountAmount.value,
        currentAmountDiscount: currentAmountDiscount.value,
      },
      orderId: undefined,
    }
    // console.log(orderData)

    // 按照API要求的结构组装数据
    const apiOrderData = {
      studentId: props.formState.studentId,
      orderDetail: {
        quoteDetailList: formData.map(course => {
          // 获取选中的sku信息
          const selectedSku = course.productSku?.find(sku => sku.id === course.productSkuId)
          const courseTotal = calculateCourseTotal(course)
          const distributedDiscount = getCourseDistributedDiscount(formData.indexOf(course))
          const realQuantity = calculateCourseRealQuantity(course, selectedSku)

          return {
            handleType: course.handleType || 0,
            courseId: course.id,
            quoteId: course.productSkuId,
            lessonMode: selectedSku?.lessonModel || 1,
            classId: course.classId || null, // 如果没有班级信息可以设为null
            count: course.skuCount || 1,
            unit: selectedSku?.unit || 1,//是否需要赋默认值
            freeQuantity: course.freeQuantity || 0,
            discountType: parseInt(course.discountType) || 0,
            discountNumber: parseInt(course.discountType) === 1 ? (course.discountNumber || 0) : parseInt(course.discountType) === 2 ? (course.discountRate || 0) : 0,
            hasValidDate: !!course.validDate,
            validDate: course.validDate || null,
            endDate: course.endDate || null,
            shareDiscount: distributedDiscount.toFixed(2),
            amount: courseTotal.toFixed(2),
            quantity: selectedSku?.quantity || 1,
            realQuantity,
            realAmount: (courseTotal - distributedDiscount).toFixed(2)
          }
        }),
        orderDiscountType: parseInt(settingForm.orderDiscountType) || 0,
        orderDiscountNumber: settingForm.orderDiscountType === '1' ? settingForm.zdiscountNumber : settingForm.zdiscountRate,
        orderDiscountAmount: (totalDiscountAmount.value).toFixed(2),
        orderRealQuantity: formData.reduce((total, course) => {
          const selectedSku = course.productSku?.find(sku => sku.id === course.productSkuId)
          return total + calculateCourseRealQuantity(course, selectedSku)
        }, 0),
        orderRealAmount: (totalAmount.value).toFixed(2),
        rechargeAccountId: selectedStudent.value?.rechargeAccountId || '',
        useBalance: Number(settingForm.useBalance || 0),
        useResidualBalance: Number(settingForm.useResidualBalance || 0),
        useGiftBalance: Number(settingForm.useGiftBalance || 0),
        internalRemark: settingForm.internalRemark || '',
        externalRemark: settingForm.externalRemark || '',
        dealDate: settingForm.dealDate,
        salePerson: settingForm.salePerson || null,
        orderTagIds: settingForm.orderLabel || []
      }
    }

    console.log('API订单数据:', apiOrderData)

    // 在提交订单前再次检测报价单信息
    const quoteDetailList = formData.map(course => {
      // 从productSku中找到选中的报价单
      const selectedSku = course.productSku?.find(sku => sku.id === course.productSkuId)

      return {
        courseId: course.id,
        quoteId: course.productSkuId,
        price: selectedSku?.price || 0,
        quantity: selectedSku?.quantity,
        lessonModel: selectedSku?.lessonModel || 1,
      }
    })

    // 检测报价单信息
    const checkResult = await getCheckQuoteInfo({ quoteDetailList })
    if (checkResult.error) {
      // 标记出错的课程
      if (checkResult.result && Array.isArray(checkResult.result)) {
        checkResult.result.forEach(errorItem => {
          if (errorItem.error === 1) {
            // 找到对应的课程并标记错误
            const courseIndex = formData.findIndex(course => course.id === errorItem.courseId)
            if (courseIndex !== -1) {
              formData[courseIndex].error = true
            }
          }
        })
      }

      messageService.error('报价单失效，请检查报价单信息')
      submitOrderLoading.value = false
      return
    }

    const res = await postCreateOrderApi(apiOrderData)
    if (res.code === 200) {
      messageService.success('创建订单成功')
      orderData.orderId = res.result
      if (Number(totalAmount.value || 0) <= 0) {
        const payRes = await postPayOrderApi({
          orderId: res.result,
          payAmount: 0,
          payAccounts: [],
        })
        if (payRes.code !== 200) {
          messageService.error(payRes.message || '订单自动完成失败')
          return
        }
        emit('submit-order', {
          ...orderData,
          autoComplete: true,
          paymentData: {
            orderId: res.result,
            payAmount: 0,
            paymentMethods: [],
          },
        })
        return
      }
      emit('submit-order', orderData)
    } else {
      messageService.error(res.message)
    }

  } catch (error) {
    console.log('提交订单失败:', error)

    // 提取并显示具体的验证错误信息
    const errorMessages = extractValidationErrors(error)
    showValidationErrors(errorMessages, '提交订单失败，请检查订单信息')
  } finally {
    submitOrderLoading.value = false
  }
}

// 处理报价单选择变化
async function handlePriceListChange(value, courseIndex) {

  if (!value) {
    return
  }

  try {
    // 获取选中的课程信息
    const courseItem = formData[courseIndex]
    if (!courseItem) {
      return
    }
    // console.log(courseItem);
    courseItem.skuCount = 1
    // 从priceList中找到选中的报价单选项，获取lessonAudition和lessonModel
    let selectedOption = null
    for (const priceGroup of courseItem.priceList) {
      const option = priceGroup.options.find(opt => opt.value === value)
      if (option) {
        selectedOption = option
        break
      }
    }

    if (!selectedOption) {
      // console.error('未找到选中的报价单选项')
      return
    }

    // console.log('选中的报价单信息:', selectedOption)
    // console.log('lessonAudition:', selectedOption.lessonAudition)
    // console.log('lessonModel:', selectedOption.lessonModel)
    // console.log('unit:', selectedOption.unit)

    // 根据lessonModel设置日期
    if (selectedOption.lessonModel === 2) {
      // 时段模式：设置开始时间为当天（使用dayjs对象）
      courseItem.validDate = dayjs().format('YYYY-MM-DD')
      // 计算结束时间和总天数
      calculateEndDate(courseItem)
    } else if (selectedOption.lessonModel === 1 || selectedOption.lessonModel === 3) {
      // 课时模式或金额模式：清空有效期
      courseItem.validDate = ''
      courseItem.endDate = ''
      courseItem.endDateWeek = ''
      courseItem.totalDays = 0
    }

    if (props.formState.studentId) {
      // 构造API参数
      const params = {
        studentId: props.formState.studentId,
        courses: [{
          courseId: courseItem.id, // 使用课程ID，不是productSkuId
          isAudition: selectedOption.lessonAudition || false // 使用实际的lessonAudition值
        }]
      }

      // 调用API查询报名类型
      const res = await getCalcCourseEnrollTypeApi(params)

      if (res.code === 200 && res.result) {
        // 根据API返回结果设置报名类型
        // 假设API返回的结果中包含enrollType字段：1-新报，2-续费
        // 实际字段名可能需要根据API文档调整
        const enrollType = res.result[0].enrollType

        formData[courseIndex].handleType = enrollType

        // console.log('查询报名类型成功:', res.result)
      }
    }
  } catch (error) {
    // console.error('查询报名类型失败:', error)
  }
}

// 根据lessonModel获取赠送标签文字
function getGiftLabel(courseItem) {
  if (!courseItem.productSkuId) {
    return '赠送课时' // 默认值
  }

  // 从priceList中找到选中的报价单选项，获取lessonModel
  let selectedOption = null
  for (const priceGroup of courseItem.priceList) {
    const option = priceGroup.options.find(opt => opt.value === courseItem.productSkuId)
    if (option) {
      selectedOption = option
      break
    }
  }

  if (!selectedOption) {
    return '赠送课时' // 默认值
  }

  // 根据lessonModel返回对应的标签文字
  switch (selectedOption.lessonModel) {
    case 1:
      return '赠送课时'
    case 2:
      return '赠送天数'
    case 3:
      return '赠送金额'
    default:
      return '赠送课时'
  }
}

// 根据lessonModel获取placeholder文字
function getGiftPlaceholder(courseItem) {
  if (!courseItem.productSkuId) {
    return '请输入课时' // 默认值
  }

  // 从priceList中找到选中的报价单选项，获取lessonModel
  let selectedOption = null
  for (const priceGroup of courseItem.priceList) {
    const option = priceGroup.options.find(opt => opt.value === courseItem.productSkuId)
    if (option) {
      selectedOption = option
      break
    }
  }

  if (!selectedOption) {
    return '请输入课时' // 默认值
  }

  // 根据lessonModel返回对应的placeholder文字
  switch (selectedOption.lessonModel) {
    case 1:
      return '请输入课时'
    case 2:
      return '请输入天数'
    case 3:
      return '请输入金额'
    default:
      return '请输入课时'
  }
}

// 判断是否显示有效期（课时模式和金额模式）
function shouldShowValidDate(courseItem) {
  if (!courseItem.productSkuId) {
    return true // 默认显示有效期
  }

  // 从priceList中找到选中的报价单选项，获取lessonModel
  let selectedOption = null
  for (const priceGroup of courseItem.priceList) {
    const option = priceGroup.options.find(opt => opt.value === courseItem.productSkuId)
    if (option) {
      selectedOption = option
      break
    }
  }

  if (!selectedOption) {
    return true // 默认显示有效期
  }

  // lessonModel === 1(课时) 或 === 3(金额) 时显示有效期
  return selectedOption.lessonModel === 1 || selectedOption.lessonModel === 3
}

// 判断是否显示开始时间和结束时间（时段模式）
function shouldShowTimePeriod(courseItem) {
  if (!courseItem.productSkuId) {
    return false // 默认不显示时段
  }

  // 从priceList中找到选中的报价单选项，获取lessonModel
  let selectedOption = null
  for (const priceGroup of courseItem.priceList) {
    const option = priceGroup.options.find(opt => opt.value === courseItem.productSkuId)
    if (option) {
      selectedOption = option
      break
    }
  }

  if (!selectedOption) {
    return false // 默认不显示时段
  }

  // lessonModel === 2(时段) 时显示开始和结束时间
  return selectedOption.lessonModel === 2
}

// 计算总天数（含赠）
function calculateTotalDays(courseItem) {
  if (!courseItem.productSkuId || !courseItem.skuCount) {
    courseItem.totalDays = 0
    return
  }

  // 从priceList中找到选中的报价单选项
  let selectedOption = null
  for (const priceGroup of courseItem.priceList) {
    const option = priceGroup.options.find(opt => opt.value === courseItem.productSkuId)
    if (option) {
      selectedOption = option
      break
    }
  }

  if (!selectedOption) {
    courseItem.totalDays = 0
    return
  }

  // 只有时段模式才需要计算总天数
  if (selectedOption.lessonModel !== 2) {
    courseItem.totalDays = 0
    return
  }

  // 如果有开始时间和结束时间，使用实际日期差计算总天数（包含开始和结束日期）
  if (courseItem.validDate && courseItem.endDate) {
    const startDate = dayjs(courseItem.validDate)
    const endDate = dayjs(courseItem.endDate)
    if (startDate.isValid() && endDate.isValid()) {
      // 计算日期差，+1 表示包含开始和结束日期
      courseItem.totalDays = endDate.diff(startDate, 'day') + 1
      return
    }
  }

  // 如果没有结束时间，使用原来的计算方式作为后备
  const { unit, quantity } = selectedOption
  // 购买数量
  const purchaseQuantity = quantity * courseItem.skuCount
  const freeQuantity = courseItem.freeQuantity || 0

  let totalDays = 0
  switch (unit) {
    case 1: // 课时 - 假设每课时1.5小时，按天计算（可根据实际业务调整）
      // 假设每天可安排2-3节课时，这里按平均每天2课时计算
      const totalCourseHours = purchaseQuantity + freeQuantity
      totalDays = Math.ceil(totalCourseHours / 2)
      break
    case 2: // 天
      totalDays = purchaseQuantity + freeQuantity
      break
    case 3: // 月
      // 购买数量按月计算（1个月=30天），赠送数量按天计算
      totalDays = purchaseQuantity * 30 + freeQuantity
      break
    case 4: // 年
      // 购买数量按年计算（1年=365天），赠送数量按天计算
      totalDays = purchaseQuantity * 365 + freeQuantity
      break
    case 5: // 其他 - 按天计算
      totalDays = purchaseQuantity + freeQuantity
      break
    default:
      totalDays = purchaseQuantity + freeQuantity
  }

  courseItem.totalDays = totalDays
}

function calculateCourseRealQuantity(courseItem, selectedSku) {
  if (!selectedSku) {
    return 0
  }

  const purchaseQuantity = (selectedSku.quantity || 1) * (courseItem.skuCount || 1)
  const freeQuantity = courseItem.freeQuantity || 0

  if (selectedSku.lessonModel === 2) {
    const totalDays = Number(courseItem.totalDays || 0)
    if (totalDays > 0) {
      return totalDays
    }

    if (courseItem.validDate && courseItem.endDate) {
      const startDate = dayjs(courseItem.validDate)
      const endDate = dayjs(courseItem.endDate)
      if (startDate.isValid() && endDate.isValid()) {
        return Math.max(0, endDate.diff(startDate, 'day') + 1)
      }
    }
  }

  return purchaseQuantity + freeQuantity
}

// 计算结束时间
function calculateEndDate(courseItem) {
  if (!courseItem.productSkuId || !courseItem.validDate || !courseItem.skuCount) {
    courseItem.endDate = ''
    courseItem.endDateWeek = ''
    return
  }

  // 从priceList中找到选中的报价单选项
  let selectedOption = null
  for (const priceGroup of courseItem.priceList) {
    const option = priceGroup.options.find(opt => opt.value === courseItem.productSkuId)
    if (option) {
      selectedOption = option
      break
    }
  }

  if (!selectedOption) {
    courseItem.endDate = ''
    courseItem.endDateWeek = ''
    return
  }

  // 只有时段模式才需要计算结束时间
  if (selectedOption.lessonModel !== 2) {
    courseItem.endDate = ''
    courseItem.endDateWeek = ''
    return
  }

  const startDate = dayjs(courseItem.validDate)
  if (!startDate.isValid()) {
    courseItem.endDate = ''
    courseItem.endDateWeek = ''
    return
  }

  const { unit, quantity } = selectedOption
  // 购买数量
  const purchaseQuantity = quantity * courseItem.skuCount
  const freeQuantity = courseItem.freeQuantity || 0

  let endDate
  switch (unit) {
    case 1: // 课时 - 假设每课时1.5小时，按天计算（可根据实际业务调整）
      // 假设每天可安排2-3节课时，这里按平均每天2课时计算
      const totalCourseHours = purchaseQuantity + freeQuantity
      const totalDays = Math.ceil(totalCourseHours / 2)
      endDate = startDate.add(totalDays, 'day')
      break
    case 2: // 天
      // 总天数 = 购买天数 + 赠送天数
      // 结束日期 = 开始日期 + (总天数 - 1) 天
      // 例如：1天是19-19，2天是19-20，3天是19-21
      const totalDaysUnit2 = purchaseQuantity + freeQuantity
      endDate = startDate.add(totalDaysUnit2 - 1, 'day')
      break
    case 3: // 月
      // 购买数量按月计算，赠送数量按天计算
      endDate = startDate.add(purchaseQuantity, 'month').subtract(1, 'day').add(freeQuantity, 'day')
      break
    case 4: // 年
      // 购买数量按年计算，赠送数量按天计算
      endDate = startDate.add(purchaseQuantity, 'year').subtract(1, 'day').add(freeQuantity, 'day')
      break
    case 5: // 其他 - 按天计算
      const totalDaysUnit5 = purchaseQuantity + freeQuantity
      endDate = startDate.add(totalDaysUnit5, 'day')
      break
    default:
      const totalDaysDefault = purchaseQuantity + freeQuantity
      endDate = startDate.add(totalDaysDefault, 'day')
  }

  // 格式化结束时间为 YYYY-MM-DD 格式，周几单独存储
  courseItem.endDate = endDate.format('YYYY-MM-DD')

  // 计算周几并单独存储
  const weekDays = ['日', '一', '二', '三', '四', '五', '六']
  const weekDay = weekDays[endDate.day()]
  courseItem.endDateWeek = `周${weekDay}`

  // 同时计算总天数
  calculateTotalDays(courseItem)
}

// 根据lessonModel获取precision值，并自动调整freeQuantity的精度
function getGiftPrecision(courseItem) {
  if (!courseItem.productSkuId) {
    return 2 // 默认值
  }

  // 从priceList中找到选中的报价单选项，获取lessonModel
  let selectedOption = null
  for (const priceGroup of courseItem.priceList) {
    const option = priceGroup.options.find(opt => opt.value === courseItem.productSkuId)
    if (option) {
      selectedOption = option
      break
    }
  }

  if (!selectedOption) {
    return 2 // 默认值
  }

  // 根据lessonModel返回对应的precision值
  let newPrecision
  switch (selectedOption.lessonModel) {
    case 1:
      newPrecision = 2 // 课时：2位小数
      break
    case 2:
      newPrecision = 0 // 天数：0位小数
      break
    case 3:
      newPrecision = 2 // 金额：2位小数
      break
    default:
      newPrecision = 2
  }

  // 自动调整freeQuantity的精度以匹配新的precision
  if (courseItem.freeQuantity !== undefined && courseItem.freeQuantity !== null) {
    const currentValue = courseItem.freeQuantity
    const adjustedValue = Number(currentValue.toFixed(newPrecision))

    // 只有当值确实发生变化时才更新，避免无限循环
    if (currentValue !== adjustedValue) {
      // 使用nextTick确保在下一个渲染周期更新
      nextTick(() => {
        courseItem.freeQuantity = adjustedValue
      })
    }
  }

  return newPrecision
}

// 获取课程标签列表
function getCourseTagList(item) {
  const tags = []

  if (item.teachMethod === 1) {
    tags.push({
      text: '班级授课',
      color: '#e6f0ff',
      textColor: '#0066ff',
      type: 'normal',
    })
  }

  if (item.teachMethod === 2) {
    tags.push({
      text: '1v1授课',
      color: '#e6f0ff',
      textColor: '#0066ff',
      type: 'normal',
    })
  }

  tags.push({
    text: item.chargeMethods,
    color: '#e6f0ff',
    textColor: '#0066ff',
    type: 'normal',
  })

  // 注意：排除课程分类标签和课程属性标签

  return tags
}

// 通用标签样式
function getTagStyle(type = 'normal') {
  const baseStyle = {
    borderRadius: '20px',
    marginRight: '0',
    height: '20px',
  }

  if (type === 'primary') {
    return {
      ...baseStyle,
      color: '#fff',
    }
  }

  return {
    ...baseStyle,
    color: '#0066ff',
  }
}
const openCreateStudent = ref(false)
const createStudentRef = ref(null)
function handleCreateNewStudent() {
  openCreateStudent.value = true
}

// 创建学员成功的回调
async function handleCreateStudentSuccess(studentData) {
  try {
    // 创建一个新对象用于API调用，避免修改原始数据
    const apiData = { ...studentData }

    // 处理渠道ID
    if (apiData.channelId && apiData.channelId.length === 1) {
      apiData.channelId = apiData.channelId[0]
    }
    else if (apiData.channelId && apiData.channelId.length > 1) {
      apiData.channelId = apiData.channelId[1]
    }
    // 调用创建学员API
    const res = await addIntendedStudentApi(apiData)

    if (res.code === 200) {
      message.success('创建学员成功')

      // 关闭modal
      openCreateStudent.value = false
      createStudentRef.value.resetForm()
      // 重置学生选择组件，让其重新查询
      if (studentSelectRef.value) {
        studentSelectRef.value.reset()
      }

      // 将新创建的学员添加到学员列表顶部
      // if (res.result && res.result.id) {
      //   const newStudent = {
      //     id: res.result.id,
      //     stuName: studentData.stuName,
      //     mobile: studentData.mobile,
      //     avatarUrl: studentData.avatar,
      //     studentStatus: 0, // 新创建的学员默认为意向学员
      //     ...res.result
      //   }

      //   // 将新学员添加到列表顶部
      //   stuListOptions.value.unshift(newStudent)

      //   // 自动选中新创建的学员
      //   const newStudentId = res.result.id
      //   emit('update:formState', { ...props.formState, id: newStudentId })

      //   console.log('新创建的学员已选中:', newStudent)
      // }
    }
    else {
      message.error(`创建学员失败：${res.message || '未知错误'}`)
    }
  }
  catch (error) {
    // console.error('创建学员失败:', error)
    message.error('创建学员失败')
  }
}
const openUpdateDate = ref(false) // 修改结束时间modal
const config = ref({})
function handleUpdateDate(validDate, endDate, courseIndex, freeQuantity) {
  openUpdateDate.value = true
  config.value = {
    validDate,
    endDate,
    courseIndex, // 保存课程索引
    freeQuantity, // 赠送天数
  }
}
function handleUpdateDateSubmit(formState) {
  console.log('formState', formState)

  // 获取要更新的课程索引
  const courseIndex = config.value.courseIndex
  if (courseIndex !== undefined && formData[courseIndex]) {
    const courseItem = formData[courseIndex]

    // 更新结束时间为用户选择的新时间
    if (formState.beforeTotalDays) {
      courseItem.endDate = formState.beforeTotalDays

      // 重新计算并设置周几
      const endDate = dayjs(formState.beforeTotalDays)
      const weekDays = ['日', '一', '二', '三', '四', '五', '六']
      const weekDay = weekDays[endDate.day()]
      courseItem.endDateWeek = `周${weekDay}`

      // 重新计算总天数
      if (courseItem.validDate) {
        const startDate = dayjs(courseItem.validDate)
        const totalDays = endDate.diff(startDate, 'day') + 1
        courseItem.totalDays = Math.max(0, totalDays)
      }

      console.log('已更新课程结束时间:', {
        courseIndex,
        newEndDate: courseItem.endDate,
        newEndDateWeek: courseItem.endDateWeek,
        newTotalDays: courseItem.totalDays
      })
    }
  }

  // 关闭modal并清空配置
  openUpdateDate.value = false
  config.value = {}
}
</script>

<template>
  <div class="current0 contenter">
    <div class="center scrollbar rounded-4">
      <div class="step student h-34 bg-white rounded-4  justify-start p-6">
        <custom-title class="mb-4" title="学员" font-size="20px" font-weight="500" />
        <a-form v-if="formState.studentId === undefined" :model="formState" name="basic">
          <a-form-item label="搜索/选择学员：" name="id" :rules="[{ required: true, message: '请选择学员' }]">
            <div class="selectBox flex">
              <div class="flex">
                <StudentSelect ref="studentSelectRef" v-model="formState.studentId" placeholder="搜索姓名/手机号" width="360px"
                  @change="handleStudentChange" @select="handleStudentSelect" />
                <a-button class="ml-6" type="primary" @click="handleCreateNewStudent">
                  创建新学员
                </a-button>
              </div>
            </div>
          </a-form-item>
          <a-form-item :wrapper-col="{ offset: 8, span: 16 }" />
        </a-form>
        <div v-if="formState.studentId" class="stuInfo mt-2 flex">
          <div class="flex">
            <div class="avatar mr-4">
              <img :src="selectedStudent?.avatarUrl || 'https://cdn.schoolpal.cn/schoolpal/next-erp/avator_female.png'"
                alt="">
            </div>
            <div class="name text-5 mr-4">
              {{ selectedStudent?.stuName || '学员姓名' }}
            </div>
            <div class="phone text-4 mr-5">
              {{ selectedStudent?.mobile || '手机号码' }}
            </div>
            <span v-if="selectedStudent?.studentStatus === 1"
              class="text-#108ee9 bg-#e6f7ff rounded-4 px-10px h-20px text-3 font-500">
              在读学员
            </span>
            <span v-else-if="selectedStudent?.studentStatus === 0"
              class="text-#f90 bg-#fff5e6 rounded-4 px-10px h-20px text-3 font-500">
              意向学员
            </span>
            <span v-else-if="selectedStudent?.studentStatus === 2"
              class="text-#999 bg-#f5f5f5 rounded-4 px-10px h-20px text-3 font-500">
              历史学员
            </span>
          </div>
          <div class="btn">
            <a-popconfirm placement="topRight" title="提示：更换学员后会重置订单设置的内容。" ok-text="我知道了" cancel-text="再想想"
              @confirm="confirm">
              <a-button type="primary" ghost>
                更换学员
              </a-button>
            </a-popconfirm>
          </div>
        </div>
      </div>
      <div class="step bg-white rounded-4 mt-3 justify-start p-6">
        <custom-title class="mb-2" title="办理内容" font-size="20px" font-weight="500" />
        <div v-if="!activeCourseOver || formData.length === 0" class="select-box flex flex-center">
          <div class="select flex flex-col flex-center">
            <img src="@/assets/images/book.svg" alt="book">
            <a-button class="mt-4" type="primary" @click="handleSelect(1)">
              按课程/学杂费/教材商品选择
            </a-button>
          </div>
          <div class="select ml-4 flex flex-col flex-center">
            <img src="@/assets/images/classes.svg" alt="classes">
            <a-button class="mt-4 w-34 classes" type="primary" @click="handleClassSelect">
              按班级选择
            </a-button>
          </div>
        </div>
        <div v-if="activeCourseOver && formData.length > 0" class="conductContent mt-5">
          <div v-for="(_, index) in formData" :key="index" class="container-box mb-4">
            <div class="container-box-top">
              <div class="flex flex-items-center">
                <span class="font-600 text-#222 text-4">{{
                  _.name
                  }}</span>
                <span class="ml-4">
                  <a-space :size="5" class="flex flex-wrap">
                    <template v-for="tag in getCourseTagList(_)" :key="tag.key || tag.text">
                      <a-tag v-if="tag.type === 'tooltip'" class="font500" :style="getTagStyle(tag.type)"
                        :color="tag.color">
                        {{ tag.text }}
                        <a-tooltip>
                          <template #title>
                            {{ tag.tooltipTitle }}
                          </template>
                          <InfoCircleOutlined class="ml-1" />
                        </a-tooltip>
                      </a-tag>
                      <a-tag v-else class="font500" :style="getTagStyle(tag.type)" :color="tag.color">
                        {{ tag.text }}
                      </a-tag>
                    </template>
                  </a-space>
                </span>
                <span class="ml-4">
                  <a-select v-model:value="_.handleType" placeholder="请选择" style="width: 100px">
                    <a-select-option :value="0">无</a-select-option>
                    <a-select-option :value="1">新报</a-select-option>
                    <a-select-option :value="2">续费</a-select-option>
                    <a-select-option :value="3">扩科</a-select-option>
                  </a-select>
                </span>
              </div>
              <div class="right">
                <span class="price">合计：¥ {{ getCourseFinalTotal(index).toFixed(2) }}</span>
                <span v-if="!handleOver" class="line" />
                <a-popconfirm v-if="!handleOver" placement="topRight" title="确定要取消选择这门课程吗？" ok-text="确定"
                  cancel-text="取消" @confirm="cancelSelectCourse(index)">
                  <span class="cancel">取消选择</span>
                </a-popconfirm>
              </div>
            </div>
            <div class="container-box-bottom scrollbar relative">
              <a-form ref="formRefs" layout="vertical" :model="_" :rules="rules" :validate-messages="validateMessages">
                <a-space :size="24" class="pr-30px ">
                  <!-- 报价单必选校验 -->
                  <a-form-item name="productSkuId" label="报价单">
                    <a-select v-model:value="_.productSkuId" :disabled="handleOver" placeholder="请选择报价单"
                      style="width: 170px" popup-class-name="auto-width-dropdown"
                      @change="(value) => handlePriceListChange(value, index)">
                      <a-select-opt-group v-for="group in _.priceList" :key="group.label" :label="group.label">
                        <a-select-option v-for="option in group.options" :key="option.value" :value="option.value">
                          <div class="flex flex-items-center ">
                            <span class="tagCustom mr-6px" v-if="option.lessonAudition">体验价</span>
                            <span class=" text-ellipsis overflow-hidden ">{{ option.label }}</span>
                          </div>
                        </a-select-option>
                      </a-select-opt-group>
                    </a-select>
                  </a-form-item>

                  <!-- 购买份数 -->
                  <a-form-item name="num" label="购买份数" :rules="numRules(index)">
                    <a-input-number v-model:value="_.skuCount" placeholder="请输入份数" :disabled="handleOver"
                      style="width: 120px" :precision="0" :min="1" @change="() => {
                        if (shouldShowTimePeriod(_)) {
                          calculateEndDate(_)
                        } else {
                          calculateTotalDays(_)
                        }
                      }" />
                  </a-form-item>

                  <!-- 赠送课时 -->
                  <a-form-item name="giftNum" :label="getGiftLabel(_)" :rules="giftNumRules(index)">
                    <a-input-number v-model:value="_.freeQuantity" :placeholder="getGiftPlaceholder(_)"
                      :disabled="handleOver" style="width: 120px" :precision="getGiftPrecision(_)" :min="0" @change="() => {
                        if (shouldShowTimePeriod(_)) {
                          calculateEndDate(_)
                        } else {
                          calculateTotalDays(_)
                        }
                      }" />
                  </a-form-item>
                  <!-- 有效期 (课时模式和金额模式) -->
                  <a-form-item v-if="shouldShowValidDate(_)" label="有效期至">
                    <a-date-picker v-if="!handleOver" v-model:value="_.validDate" format="YYYY-MM-DD"
                      value-format="YYYY-MM-DD" class="w-35" :disabled="handleOver"
                      :disabled-date="disabledValidDate" />
                    <a-date-picker v-if="handleOver && _.validDate" format="YYYY-MM-DD" value-format="YYYY-MM-DD"
                      v-model:value="_.validDate" class="w-35" :disabled="handleOver" />
                    <div v-if="handleOver && _.validDate === ''" class="w-35">
                      不限制
                    </div>
                  </a-form-item>
                  <!-- 开始时间 (时段模式) -->
                  <a-form-item v-if="shouldShowTimePeriod(_)" label="开始时间">
                    <a-date-picker v-if="!handleOver" v-model:value="_.validDate" format="YYYY-MM-DD"
                      value-format="YYYY-MM-DD" class="w-35" :disabled="handleOver" :disabled-date="disabledStartDate"
                      @change="() => calculateEndDate(_)" />
                    <a-date-picker v-if="handleOver && _.validDate" format="YYYY-MM-DD" value-format="YYYY-MM-DD"
                      v-model:value="_.validDate" class="w-35" :disabled="handleOver" />
                    <div v-if="handleOver && !_.validDate" class="w-35">
                      未设置
                    </div>
                  </a-form-item>
                  <!-- 结束时间 (时段模式) -->
                  <a-form-item v-if="shouldShowTimePeriod(_)" label="结束时间">
                    <div v-if="!_.endDate" class="w-35 text-#ccc">
                      系统会自动计算
                    </div>
                    <div v-if="_.endDate" class="w-35 flex">
                      {{ _.endDate }}{{ _.endDateWeek ? `(${_.endDateWeek})` : '' }}
                      <!-- 编辑icon -->
                      <FormOutlined v-if="!handleOver" class="text-#06f cursor-pointer ml1"
                        @click="handleUpdateDate(_.validDate, _.endDate, index, _.freeQuantity)" />
                    </div>
                  </a-form-item>
                  <!-- 总天数（含赠） -->
                  <a-form-item v-if="shouldShowTimePeriod(_) && handleOver" label="总天数（含赠）">
                    <div class="w-35 text-#333">
                      {{ _.totalDays }}
                    </div>
                  </a-form-item>
                  <!-- 单课优惠 -->
                  <a-form-item name="radio" label="单课优惠">
                    <div v-if="!handleOver" class="flex flex-items-center ">
                      <a-radio-group v-model:value="_.discountType" class="custom-radio  whitespace-nowrap">
                        <a-radio value="0">
                          无
                        </a-radio>
                        <a-radio value="1">
                          金额
                        </a-radio>
                        <a-radio value="2">
                          折扣
                        </a-radio>
                      </a-radio-group>

                      <template v-if="_.discountType === '1'">
                        <a-form-item name="discountNumber" :rules="discountNumberRules(index)" class="ml-2 mgnone">
                          <div class="flex flex-center">
                            <a-input-number v-model:value="_.discountNumber" :disabled="handleOver" style="width: 100px"
                              :precision="2" :min="0.01" placeholder="金额" />
                            <span class="ml-1">元</span>
                          </div>
                        </a-form-item>
                      </template>

                      <template v-if="_.discountType === '2'">
                        <a-form-item name="discountRate" :rules="discountRateRules(index)" class="ml-2 mgnone">
                          <div class="flex flex-center styleCss relative">
                            <a-input-number v-model:value="_.discountRate" :disabled="handleOver" style="width: 100px"
                              :precision="1" :min="0.1" :max="9.9" placeholder="折扣" />
                            <span class="ml-1">折</span>
                            <span v-if="_.discountRate && _.productSkuId && _.skuCount"
                              class="ml-5 text-12px text-#f03 absolute left--18px bottom--18px">
                              - ¥{{((_.productSku?.find(sku => sku.id === _.productSkuId)?.price || 0)
                                * _.skuCount * (1 - _.discountRate / 10)).toFixed(2)}}
                            </span>
                          </div>
                        </a-form-item>
                      </template>
                    </div>
                    <div v-else class="text-#f00 whitespace-nowrap">
                      <span v-if="_.discountType === '1' && _.discountNumber">
                        -¥ {{ _.discountNumber.toFixed(2) }}
                      </span>
                      <span v-else-if="_.discountType === '2' && _.discountRate">
                        - ¥{{((_.productSku?.find(sku => sku.id === _.productSkuId)?.price || 0)
                          * _.skuCount * (1
                            - _.discountRate / 10)).toFixed(2)}}
                      </span>
                      <span v-else>
                        无优惠
                      </span>
                    </div>
                  </a-form-item>
                  <!-- 整单分摊优惠 -->
                  <a-form-item v-if="handleOver" class="ml-10" label="整单分摊优惠">
                    <div class="text-#f00 whitespace-nowrap">
                      -¥ <span>{{ getCourseDistributedDiscount(index).toFixed(2) }}</span>
                    </div>
                  </a-form-item>
                  <a-form-item v-if="handleOver" class="ml-6" label="分摊优惠券优惠">
                    <div>无</div>
                  </a-form-item>
                  <a-form-item v-if="handleOver" class="ml-6" label="分摊赠送金额">
                    <div>无</div>
                  </a-form-item>
                </a-space>
                <div v-if="_.error"
                  class="mask absolute left-0 top-0 w-full h-full bg-#fff opacity-90 z-10 flex flex-center flex-col">
                  <div class="text-#333 mb-8px">出错了，报价单规格发生变化</div>
                  <a-button danger ghost @click="handleCancel">重新选择课程/教材商品</a-button>
                </div>
              </a-form>
            </div>
          </div>
          <a-space v-if="!handleOver" :size="14" class="flex justify-end">
            <a-button type="primary" ghost @click="handleClassSelect">
              按班级选择
            </a-button>
            <a-button type="primary" ghost @click="handleSelect(1)">
              选择课程/学杂费/教材商品
            </a-button>
            <a-button type="primary" style="width: 128px" :loading="saveLoading" @click="debouncedValidateAll">
              保存
            </a-button>
          </a-space>
          <a-space v-if="handleOver" :size="14" class="flex justify-end">
            <a-button type="primary" ghost @click="handleEdit">
              编辑内容
            </a-button>
          </a-space>
        </div>
      </div>
      <div class="step bg-white rounded-4 mt-3 justify-start p-6">
        <custom-title class="mb-5" title="订单设置" font-size="20px" font-weight="500" />
        <a-form v-if="handleOver && formState.studentId && activeCourseOver" ref="settingOrderForm" :model="settingForm"
          label-align="left">
          <div class="order-settings-layout">
            <div class="order-settings-left">
              <a-form-item class="custom-label" label="整单优惠设置：">
                <div class="flex flex-items-center">
                  <a-radio-group v-model:value="settingForm.orderDiscountType" class="custom-radio whitespace-nowrap">
                    <a-radio value="0">
                      无
                    </a-radio>
                    <a-radio value="1">
                      金额
                    </a-radio>
                    <a-radio value="2">
                      折扣
                    </a-radio>
                  </a-radio-group>

                  <template v-if="settingForm.orderDiscountType === '1'">
                    <a-form-item name="zdiscountNumber" :rules="zdiscountNumberRules()" class="ml-2 mgnone">
                      <div class="flex flex-center ">
                        <a-input-number v-model:value="settingForm.zdiscountNumber" style="width: 100px" :precision="2"
                          :min="0.01" placeholder="输入金额" />
                        <span class="ml-1">元</span>
                        <span class="ml-5 text-12px text-#f03 " />
                      </div>
                    </a-form-item>
                  </template>

                  <template v-if="settingForm.orderDiscountType === '2'">
                    <a-form-item name="zdiscountRate" :rules="zdiscountRateRules()" class="ml-2 mgnone">
                      <div class="flex flex-center styleCss relative">
                        <a-input-number v-model:value="settingForm.zdiscountRate" style="width: 100px" :precision="1"
                          :min="0.1" :max="9.9" placeholder="输入折扣" />
                        <span class="ml-1">折</span>
                        <span class="ml-5 text-12px text-#f03 absolute left--18px bottom--18px">
                          <template v-if="currentDiscountAmount > 0">
                            减 ¥{{ currentDiscountAmount.toFixed(2) }}
                          </template>
                        </span>
                      </div>
                    </a-form-item>
                  </template>
                </div>
              </a-form-item>
              <a-form-item class="custom-label mt--2" label="经办日期：">
                <a-date-picker v-model:value="settingForm.dealDate" class="w-80" :disabled-date="disabledDate"
                  value-format="YYYY-MM-DD" format="YYYY-MM-DD" @change="handleDateChange" />
              </a-form-item>
              <div class="order-settings-right-spacer" />
            </div>

            <div class="order-settings-right">
              <div class="storage-account-panel">
                <div class="storage-account-panel__label">储值账户抵扣</div>
                <div class="storage-account-panel__content">
                  <div class="storage-account-inline">
                    <span class="storage-account-inline__title">充值余额</span>
                    <a-input-number v-model:value="settingForm.useBalance" :precision="2" :min="0"
                      :max="getUsageMax('useBalance')" class="storage-account-input" placeholder="请输入"
                      @change="normalizeUsageValue('useBalance')">
                      <template #addonAfter>
                        元
                      </template>
                    </a-input-number>
                  </div>
                  <div class="storage-account-inline">
                    <span class="storage-account-inline__title">残联余额</span>
                    <a-input-number v-model:value="settingForm.useResidualBalance" :precision="2" :min="0"
                      :max="getUsageMax('useResidualBalance')" class="storage-account-input" placeholder="请输入"
                      @change="normalizeUsageValue('useResidualBalance')">
                      <template #addonAfter>
                        元
                      </template>
                    </a-input-number>
                  </div>
                  <div class="storage-account-inline">
                    <span class="storage-account-inline__title">赠送余额</span>
                    <a-input-number v-model:value="settingForm.useGiftBalance" :precision="2" :min="0"
                      :max="getUsageMax('useGiftBalance')" class="storage-account-input" placeholder="请输入"
                      @change="normalizeUsageValue('useGiftBalance')">
                      <template #addonAfter>
                        元
                      </template>
                    </a-input-number>
                  </div>
                </div>
                <div class="storage-account-panel__hint-row">
                  <div class="storage-account-hint-item">
                    <span>可用 ¥ {{ availableRechargeBalance.toFixed(2) }}</span>
                    <span v-if="getUsageMax('useBalance') > 0" class="storage-account-link" @click.stop="applyFullUsage('useBalance')">全部使用</span>
                  </div>
                  <div class="storage-account-hint-item">
                    <span>可用 ¥ {{ availableResidualBalance.toFixed(2) }}</span>
                    <span v-if="getUsageMax('useResidualBalance') > 0" class="storage-account-link" @click.stop="applyFullUsage('useResidualBalance')">全部使用</span>
                  </div>
                  <div class="storage-account-hint-item">
                    <span>可用 ¥ {{ availableGiftBalance.toFixed(2) }}</span>
                    <span v-if="getUsageMax('useGiftBalance') > 0" class="storage-account-link" @click.stop="applyFullUsage('useGiftBalance')">全部使用</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
          <a-form-item class="custom-label" label="优惠券：">
            <a-button class="w-35" ghost type="primary" @click="handleOpenCouponModal">
              输入优惠券码
            </a-button>
          </a-form-item>
          <a-form-item class="custom-label" label="订单销售员：">
            <StaffSelect v-model="settingForm.salePerson" placeholder="请选择销售员" width="320px" :status="0" />
          </a-form-item>
          <a-form-item class="custom-label" label="订单标签：">
            <a-select v-model:value="settingForm.orderLabel" mode="multiple" placeholder="请选择订单标签" show-search
              style="width: 100%" :options="orderLabelOptions" :filter-option="orderLabelFilterOption"
              :field-names="{ label: 'name', value: 'id' }" />
          </a-form-item>
          <a-form-item class="custom-label" label="对内备注：">
            <a-input v-model:value="settingForm.internalRemark" placeholder="请输入对内备注，此备注仅内部员工可见" style="width: 100%" />
          </a-form-item>
          <a-form-item class="custom-label" label="对外备注：">
            <a-input v-model:value="settingForm.externalRemark" placeholder="请输入对外备注，此备注打印时将显示" style="width: 100%" />
          </a-form-item>
        </a-form>
      </div>
    </div>
    <a-affix :offset-bottom="0">
      <div class="step h-20 flex flex-center justify-end pr-10 footer bg-white rounded-4 mt-3">
        <div v-if="handleOver" class="text-5.6 totalPrice setting mr-12">
          应收金额：<span>¥ {{ totalAmount.toFixed(2) }}</span>
        </div>
        <a-button class="h-11 w-35" :disabled="!handleOver" :loading="submitOrderLoading" type="primary"
          @click="debouncedSubmitOrder">
          提交订单
        </a-button>
      </div>
    </a-affix>
    <!-- 选择课程/学杂费/教材商品 -->
    <ActiveCourseModal ref="activeCourseModalRef" v-model:open="openSelect" :selected-courses="formData"
      @confirm="handleCourseModalConfirm" />

    <!-- 创建新学员 -->
    <CreateStudent ref="createStudentRef" v-model:open="openCreateStudent" :type="1"
      @submit="handleCreateStudentSuccess" />
  </div>
  <UpdateDateModal v-model:open="openUpdateDate" :config="config" @submit="handleUpdateDateSubmit" />
  
  <!-- 优惠券输入弹窗 -->
  <CouponInputModal 
    v-model:open="openCouponModal" 
    @confirm="handleCouponConfirm"
  />
</template>

<style lang="less" scoped>
.step {
  box-shadow: 0 0 8px 0 rgba(94, 188, 255, 0.08);
}

.tagCustom {
  display: inline-block;
  padding: 2px 5px;
  background: linear-gradient(212deg, #fad961, #f76b1c);
  border-radius: 8px;
  font-size: 10px;
  font-weight: 500;
  color: #fff;
  line-height: 12px;
}

/* 自适应下拉框宽度 */
:global(.auto-width-dropdown) {
  width: auto !important;
  min-width: 170px !important;
  max-width: 400px !important;
}

:global(.auto-width-dropdown .ant-select-item-option-content) {
  white-space: nowrap !important;
}

:global(.auto-width-dropdown .ant-select-item) {
  white-space: nowrap !important;
  padding: 8px 12px !important;
}

:global(.auto-width-dropdown .ant-select-item-group) {
  padding: 8px 12px 4px !important;
  font-weight: 600 !important;
  color: #666 !important;
}

:deep(.custom-label .ant-form-item-label) {
  min-width: 104px;

  label {
    color: #333 !important;
    font-weight: 500 !important;
  }
}

.order-settings-layout {
  position: relative;
}

.order-settings-left {
  min-width: 0;
  padding-right: 736px;
}

.order-settings-right-spacer {
  height: 4px;
}

.order-settings-right {
  position: absolute;
  top: 0;
  right: 0;
  width: 716px;
  z-index: 3;
}

.storage-account-panel {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  width: 100%;
  padding: 12px 14px 10px;
  border: 1px solid #edf2ff;
  border-radius: 14px;
  background: linear-gradient(180deg, #fcfdff 0%, #f7f9ff 100%);
  box-sizing: border-box;
  position: relative;
  z-index: 3;
}

.storage-account-panel__label {
  color: #333;
  font-size: 14px;
  font-weight: 600;
  line-height: 20px;
  margin-bottom: 10px;
}

.storage-account-panel__content {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 10px;
  width: 100%;
}

.storage-account-inline {
  min-width: 0;
}

.storage-account-inline__title {
  color: #666;
  display: block;
  font-size: 12px;
  margin-bottom: 4px;
  white-space: nowrap;
}

.storage-account-input {
  width: 100%;
}

:deep(.storage-account-input .ant-input-number-group-addon) {
  min-width: 34px;
  padding: 0 8px;
}

.storage-account-panel__hint-row {
  color: #666;
  display: grid;
  font-size: 11px;
  grid-template-columns: repeat(3, 1fr);
  gap: 10px;
  line-height: 16px;
  margin-top: 6px;
  width: 100%;
}

.storage-account-hint-item {
  display: flex;
  align-items: center;
  gap: 6px;
  min-height: 24px;
}

.storage-account-link {
  display: inline-flex;
  align-items: center;
  min-height: 24px;
  padding: 2px 6px;
  border-radius: 6px;
  color: #1677ff;
  cursor: pointer;
  font-size: 11px;
  line-height: 1;
  user-select: none;
  flex-shrink: 0;
}

.storage-account-link:hover {
  background: rgba(22, 119, 255, 0.08);
  color: #0958d9;
}

@media (max-width: 1480px) {
  .order-settings-left {
    padding-right: 676px;
  }
  .order-settings-right {
    width: 656px;
  }
}

@media (max-width: 1280px) {
  .order-settings-layout {
    min-height: unset;
  }

  .order-settings-left {
    padding-right: 0;
  }

  .order-settings-right {
    position: static;
    width: 640px;
    margin-left: auto;
    margin-top: 8px;
  }

  .storage-account-panel__content,
  .storage-account-panel__hint-row {
    grid-template-columns: 1fr;
  }
}

.totalPrice {
  font-weight: bold;
  display: flex;
  align-items: flex-end;
  flex-direction: column;
  line-height: 1.2;
  font-family: "DIN alternate", sans-serif;

  span {
    color: var(--pro-ant-color-primary);
    font-family: "DIN alternate", sans-serif;
    font-size: 38px;
  }
}

.totalPrice.setting {
  flex-direction: row;
  align-items: center;
}

.contenter {
  display: flex;
  flex-direction: column;
  height: calc(100vh - 195px);
  overflow: hidden;

  .center {
    flex: 1;
    overflow: auto;
    margin-top: 12px;
  }
}

.current0 {
  .m-r-b-r a {
    color: var(--pro-ant-color-primary);
  }

  .classes {
    background: var(--pro-ant-color-success);

    &:hover {
      background: var(--pro-ant-color-success-hover);
    }
  }

  .stuInfo {
    display: flex;
    align-items: center;
    justify-content: space-between;

    .flex {
      align-items: center;
      background: #fafafa;
      border-radius: 100px;
      padding-right: 16px;

      .avatar {
        width: 40px;
        height: 40px;

        img {
          width: 100%;
          height: 100%;
          border-radius: 100%;
        }
      }

      .name {
        font-weight: 500;
      }
    }
  }

  .conductContent {
    .container-box {
      background: #fafafa;

      .container-box-top {
        height: 44px;
        background: #f0f5fe;
        padding: 10px 24px;
        display: flex;
        justify-content: space-between;
        align-items: center;
        border-top-left-radius: 16px;
        border-top-right-radius: 16px;

        .right {
          display: flex;
          align-items: center;

          span.price {
            font-weight: 500;
            color: #222;
            white-space: nowrap;
            font-size: 14px;
          }

          .line {
            height: 12px;
            width: 1px;
            background: #ccc;
            display: inline-block;
            margin: 0 18px;
          }

          .cancel {
            color: var(--pro-ant-color-primary);
            font-weight: bold;
            cursor: pointer;
          }
        }
      }

      .container-box-bottom {
        padding: 10px 8px 0 24px;
        overflow-x: auto;

        :deep(.ant-form-item-label label) {
          white-space: nowrap;
        }
      }
    }
  }

  .mgnone {
    margin: 0 !important;

    :deep(.ant-form-item-explain-error) {
      position: absolute;
    }
  }

  :deep(.ant-form-item-explain-error) {
    font-size: 12px !important;
  }

  :deep(.ant-select-disabled .ant-select-selector) {
    border-color: transparent !important;
    background: transparent !important;
    color: #333 !important;
    cursor: default !important;
  }

  :deep(.ant-select-disabled .ant-select-arrow) {
    display: none !important;
  }

  :deep(.ant-input-number-disabled) {
    border-color: transparent !important;
    background: transparent !important;
    color: #333 !important;
    cursor: default !important;
  }

  :deep(.ant-input-number-disabled .ant-input-number-input) {
    padding-left: 0;
  }

  :deep(.ant-picker-disabled) {
    border-color: transparent !important;
    background: transparent !important;
    padding-left: 0;
  }

  :deep(.ant-picker-disabled input) {
    color: #333 !important;
  }

  :deep(.ant-picker-disabled .ant-picker-suffix) {
    display: none !important;
  }

  /* 自定义镂空样式 */
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

}

:deep(.ant-select-selection-item) {
  color: #333;
}
</style>

<style lang="less">
.modal {
  .ant-modal-header {
    margin-bottom: 0;
  }

  .ant-modal-footer {
    margin-top: 0;
  }
}
</style>
