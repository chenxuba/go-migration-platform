<script setup>
import {
  CloseOutlined,
  DownOutlined,
  ExclamationCircleOutlined,
  PlusOutlined,
} from '@ant-design/icons-vue'
import { reactive, ref, computed, watch, onMounted, onUnmounted, nextTick } from 'vue'
import { message } from 'ant-design-vue'
import dayjs from 'dayjs'
import { addIntendedStudentApi } from '@/api/enroll-center/intention-student'
import { getOrderTagListPagedApi } from '@/api/finance-center/order-tag'
import {
  createRechargeAccountOrderApi,
  getRechargeAccountByStudentApi,
  getRechargeAccountOrderDetailApi,
  getStudentDetailApi,
  payOrderBySchoolPalApi,
} from '@/api/finance-center/recharge-account'
import { openOrderReceiptPage } from '@/utils/order-receipt'
import StaffSelect from './staff-select.vue'
import StudentSelect from './student-select.vue'
import CreateStudent from './create-student.vue'
import OrderDetailDrawer from './order-detail-drawer.vue'
import { payMethodOptionsWithIcons } from './pay-method-options-data'

const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
  stuId: {
    type: [Number, String],
    default: undefined,
  },
})
const emit = defineEmits(['update:open', 'submitted'])
const studentSelectRef = ref()
const payType = ref('1')
const pay = ref(1)
const checkOptions = reactive([...payMethodOptionsWithIcons])
const fileList = ref([])
const accountList = ref([{ value: 1, label: '默认账户' }])
const confirmFormState = reactive({
  account: 1,
  payDate: undefined,
  billRemarks: undefined,
})
// 禁止选择今天之后的日期
function disabledDate(current) {
  // 当天结束时间（23:59:59）之后的时间不可选
  return current > dayjs().endOf('day')
}
function getBase64(file) {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.readAsDataURL(file)
    reader.onload = () => resolve(reader.result)
    reader.onerror = error => reject(error)
  })
}
function handleCancelImg() {
  previewVisible.value = false
  previewTitle.value = ''
}
const previewVisible = ref(false)
const previewImage = ref('')
const previewTitle = ref('')
const accountFormRefs = ref()
async function handlePreview(file) {
  if (!file.url && !file.preview) {
    file.preview = await getBase64(file.originFileObj)
  }
  previewImage.value = file.url || file.preview
  previewVisible.value = true
  previewTitle.value = file.name || file.url.substring(file.url.lastIndexOf('/') + 1)
}
// 处理双向绑定
const openDrawer = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})
function closeFun() {
  // 重置表单状态
  Object.assign(formState, {
    ...initialFormState,
    createTime: dayjs().format('YYYY-MM-DD'), // 确保时间动态更新
  })
  checkPayPrice.value = false
  checkPayPriceLine.value = false
  formState.studentId = undefined
  selectedStudent.value = null
  currentRechargeOrderDetail.value = null
  // 重置学员选择组件，确保重新打开抽屉时重新请求接口
  if (studentSelectRef.value) {
    studentSelectRef.value.reset()
  }
  // 关闭创建学员modal
  openCreateStudent.value = false
  Object.assign(confirmFormState, {
    account: 1,
    payDate: undefined,
    billRemarks: undefined,
  })
  openDrawer.value = false
}

// 处理销售员选择变化
function handleSalespersonChange(value, staffInfo) {
  console.log('选择的销售员:', value, staffInfo)
}

// 监听抽屉打开状态
watch(() => props.open, (newVal) => {
  if (newVal) {
    // 抽屉打开时重置学员选择组件，确保获取最新数据
    if (studentSelectRef.value) {
      studentSelectRef.value.reset()
    }
    // 重置创建学员modal状态
    openCreateStudent.value = false
    if (props.stuId) {
      nextTick(() => {
        handleStudentSelect({ id: props.stuId })
      })
    }
  }
})

// 选中的学员信息
const selectedStudent = ref(null)

// 处理学员选择
async function handleStudentSelect(student) {
  console.log('选择的学员:', student)
  const studentId = student?.id || formState.studentId
  if (!studentId) {
    selectedStudent.value = null
    return
  }
  try {
    const [{ result: studentDetail }, { result: rechargeAccount }] = await Promise.all([
      getStudentDetailApi({ studentId }),
      getRechargeAccountByStudentApi({ studentId }),
    ])
    selectedStudent.value = {
      ...student,
      ...studentDetail,
      stuName: studentDetail?.name || student?.stuName || student?.name,
      mobile: studentDetail?.phone || student?.mobile || student?.phone,
      avatarUrl: studentDetail?.avatar || student?.avatarUrl || student?.avatar,
      studentStatus: studentDetail?.status ?? student?.studentStatus,
      salesPersonId: studentDetail?.salespersonId || '0',
      salesPersonName: studentDetail?.salespersonName || '',
      collectorStaffId: studentDetail?.collectorStaffId || '0',
      phoneSellStaffId: studentDetail?.phoneSellStaffId || '0',
      foregroundStaffId: studentDetail?.foregroundStaffId || '0',
      viceSellStaffStaffId: studentDetail?.viceSellStaffStaffId || '0',
      rechargeAccountId: rechargeAccount?.id || '',
      rechargeAccountName: rechargeAccount?.accountName || rechargeAccount?.rechargeAccountName || '',
      rechargeAccountPhone: rechargeAccount?.phone || '',
      mainStudentId: rechargeAccount?.mainStudentId || '',
      accountBalance: rechargeAccount?.balance || 0,
      giftBalance: rechargeAccount?.givingBalance || 0,
      residualBalance: rechargeAccount?.residualBalance || 0,
      rechargeAccountStudents: rechargeAccount?.students || [],
    }
    formState.salesperson = selectedStudent.value.salesPersonId || undefined
  }
  catch (error) {
    console.error('加载学员储值信息失败:', error)
    message.error('加载学员信息失败')
  }
}

// 格式化手机号码
function formatMobile(mobile) {
  if (!mobile) return ''
  if (mobile.length === 11) {
    return mobile.replace(/(\d{3})\d{4}(\d{4})/, '$1****$2')
  }
  return mobile
}

function formatMoney2(value) {
  return Number(value || 0).toFixed(2)
}
const checkPayPrice = ref(false)
const checkPayPriceLine = ref(false)
const openConfirmDrawer = ref(false)
const openModal = ref(false)
const completionSnapshot = ref({
  payDate: '',
  billRemarks: '',
  payMethodLabel: '',
  accountLabel: '',
})

const formRef = ref()
const rules = {
  studentId: [
    {
      required: true,
      message: '请选择学员',
      trigger: 'change',
    },
  ],
}
const initialFormState = { // 可抽离初始状态（可选）
  studentId: undefined,
  createTime: dayjs().format('YYYY-MM-DD'),
  givePayPrice: undefined,
  cldPayPrice: undefined,
  payPrice: undefined,
   salesperson: undefined,
   orderTag: [],
  inRemarks: undefined,
  outRemarks: undefined,
}

const formState = reactive({ ...initialFormState })
watch(
  () => props.open,
  (newOpen) => {
    if (newOpen) {
      // 重置所有字段，并设置最新的 stuId
      Object.assign(formState, {
        ...initialFormState,
        studentId: props.stuId, // 关键步骤
        createTime: dayjs().format('YYYY-MM-DD'), // 更新时间（如需）
      })
      if (props.stuId) {
        handleStudentSelect({ id: props.stuId })
      }
    }
  },
  { immediate: true }, // 初始化时立即执行一次
)
function changeStudent() {
  formState.studentId = undefined
  selectedStudent.value = null
  // 重置创建学员modal状态
  openCreateStudent.value = false
}
function handleChange(value) {
  console.log(`selected ${value}`)
}
// 提交时校验
async function handleSubmit() {
  formRef.value.validate().then(async () => {
    if (!formState.payPrice && !formState.givePayPrice && !formState.cldPayPrice) {
      checkPayPrice.value = true
      setTimeout(() => {
        checkPayPriceLine.value = true
      }, 400)
    }
    else {
      checkPayPrice.value = false
      checkPayPriceLine.value = false
      if (!selectedStudent.value?.rechargeAccountId) {
        message.error('未找到关联储值账户')
        return
      }
      try {
        const { result: orderResult } = await createRechargeAccountOrderApi({
          rechargeAccountId: String(selectedStudent.value.rechargeAccountId),
          amount: Number(formState.payPrice || 0),
          givingAmount: Number(formState.givePayPrice || 0),
          residualAmount: Number(formState.cldPayPrice || 0),
          dealDate: formState.createTime,
          salePersonId: String(formState.salesperson || selectedStudent.value.salesPersonId || '0'),
          collectorStaffId: String(selectedStudent.value.collectorStaffId || '0'),
          phoneSellStaffId: String(selectedStudent.value.phoneSellStaffId || '0'),
          foregroundStaffId: String(selectedStudent.value.foregroundStaffId || '0'),
          viceSellStaffStaffId: String(selectedStudent.value.viceSellStaffStaffId || '0'),
          remark: formState.inRemarks || '',
          orderTagIds: Array.isArray(formState.orderTag) ? formState.orderTag.map(String) : [],
          externalRemark: formState.outRemarks || '',
          studentId: String(formState.studentId),
        })
        const rechargeAccountOrderId = orderResult?.id
        if (!rechargeAccountOrderId) {
          throw new Error('创建充值订单失败')
        }
        const { result: orderDetail } = await getRechargeAccountOrderDetailApi({
          rechargeAccountOrderId: String(rechargeAccountOrderId),
        })
        currentRechargeOrderDetail.value = orderDetail
        const billId = Number(orderDetail?.bill?.id || 0)
        const requiresBill = Number(orderDetail?.amount ?? formState.payPrice ?? 0) > 0
        if (!requiresBill || billId <= 0) {
          pay.value = 1
          fileList.value = []
          Object.assign(confirmFormState, {
            account: 1,
            payDate: undefined,
            billRemarks: undefined,
          })
          completionSnapshot.value = {
            payDate: '-',
            billRemarks: '实收金额为0，系统已直接完成订单，未生成账单',
            payMethodLabel: '无需收款',
            accountLabel: '-',
          }
          message.success('账户充值成功')
          openOverDrawer.value = true
          openDrawer.value = false
          emit('submitted')
          return
        }
      }
      catch (error) {
        console.error('创建充值订单失败:', error)
        message.error(error?.message || '创建充值订单失败')
        return
      }
      openConfirmDrawer.value = true
      openDrawer.value = false
    }
  }).catch((err) => {
    console.log('验证失败', err)
  })
}

// 监听金额字段变化，实时移除验证提示
watch( 
  [() => formState.payPrice, () => formState.givePayPrice, () => formState.cldPayPrice],
  ([payPrice, givePayPrice, cldPayPrice]) => {
    // 如果任何一个字段有值，就移除验证提示
    if (payPrice || givePayPrice || cldPayPrice) {
      checkPayPrice.value = false
      checkPayPriceLine.value = false
    }
  }
)
function handleCancel() {
  openModal.value = false
  openConfirmDrawer.value = false
  openDrawer.value = false
  closeFun()
  fileList.value = []
  pay.value = 1
  // 关闭创建学员modal
  openCreateStudent.value = false
}
function handleOk() {
  openModal.value = false
}

async function handleOver() {
  if (!currentRechargeOrderDetail.value?.bill?.id) {
    message.error('未找到支付账单')
    return
  }
  try {
    await accountFormRefs.value?.validate()
  }
  catch {
    return
  }
  try {
    const payDateTimeStr = confirmFormState.payDate
      ? dayjs(confirmFormState.payDate)
          .hour(dayjs().hour())
          .minute(dayjs().minute())
          .second(dayjs().second())
          .format('YYYY-MM-DD HH:mm:ss')
      : ''
    await payOrderBySchoolPalApi({
      billId: String(currentRechargeOrderDetail.value.bill.id),
      amount: Number(formState.payPrice || 0),
      remark: confirmFormState.billRemarks || '',
      payMethod: pay.value,
      amountId: Number(confirmFormState.account || 0),
      payTime: payDateTimeStr || undefined,
    })
    completionSnapshot.value = {
      payDate: payDateTimeStr,
      billRemarks: confirmFormState.billRemarks || '',
      payMethodLabel: checkOptions.find(item => item.id === pay.value)?.label || '-',
      accountLabel: accountList.value.find(item => item.value === confirmFormState.account)?.label || '-',
    }
    message.success('账户充值成功')
    openOverDrawer.value = true
    openConfirmDrawer.value = false
    emit('submitted')
  }
  catch (error) {
    console.error('账户充值失败:', error)
    message.error(error?.response?.data?.message || error?.message || '账户充值失败')
    return
  }
  pay.value = 1
  Object.assign(confirmFormState, {
    account: 1,
    payDate: undefined,
    billRemarks: undefined,
  })
  fileList.value = []
}
const openOverDrawer = ref(false)
function closeOverDrawer() {
  openOverDrawer.value = false
}

const openOrderDetailDrawer = ref(false)
const orderDetailDrawerOrderId = ref('')
function openRechargeSaleOrderDetailDrawer() {
  const id = String(currentRechargeOrderDetail.value?.saleOrderId ?? '').trim()
  if (!id) {
    message.warning('暂无关联系统订单，无法查看详情')
    return
  }
  orderDetailDrawerOrderId.value = id
  openOrderDetailDrawer.value = true
}

function handlePrintReceipt() {
  const id = String(currentRechargeOrderDetail.value?.saleOrderId ?? '').trim()
  if (!id) {
    message.warning('暂无关联系统订单，无法打印收据')
    return
  }
  openOrderReceiptPage(id, { template: 'a4' })
}

function handleDownloadReceipt() {
  const id = String(currentRechargeOrderDetail.value?.saleOrderId ?? '').trim()
  if (!id) {
    message.warning('暂无关联系统订单，无法下载收据')
    return
  }
  openOrderReceiptPage(id, { template: 'a4', autoPrint: true })
}

// 创建学员相关
const openCreateStudent = ref(false)
const createStudentRef = ref()
const orderTagOptions = ref([])
const currentRechargeOrderDetail = ref(null)

// 处理创建新学员按钮点击
function handleCreateNewStudent() {
  openCreateStudent.value = true
}

// 处理创建学员取消
function handleCreateStudentCancel() {
  openCreateStudent.value = false
  if (createStudentRef.value) {
    createStudentRef.value.closeSpinning()
  }
}

// 处理创建学员成功
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

      // 关闭创建学员modal
      openCreateStudent.value = false
      if (createStudentRef.value) {
        createStudentRef.value.resetForm()
        createStudentRef.value.closeSpinning() // 关闭loading状态
      }

      // 重置学员选择组件，确保重新请求接口获取最新数据
      if (studentSelectRef.value) {
        studentSelectRef.value.reset()
      }
    }
    else {
      message.error(`创建学员失败：${res.message || '未知错误'}`)
      // 关闭loading状态但不关闭modal
      if (createStudentRef.value) {
        createStudentRef.value.closeSpinning()
      }
    }
  }
  catch (error) {
    console.error('创建学员失败:', error)
    message.error('创建学员失败，请重试')
    // 关闭loading状态但不关闭modal
    if (createStudentRef.value) {
      createStudentRef.value.closeSpinning()
    }
  }
}
function closePage() {
  openOverDrawer.value = false
  closeFun()
}
function closeConfirmDrawer() {
  openModal.value = true
}

// 响应式宽度
const windowWidth = ref(window.innerWidth)

// 监听窗口尺寸变化
function handleResize() {
  windowWidth.value = window.innerWidth
}

onMounted(() => {
  window.addEventListener('resize', handleResize)
  fetchOrderTagOptions()
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
})

async function fetchOrderTagOptions() {
  try {
    const { result } = await getOrderTagListPagedApi({
      queryModel: { enable: true },
      sortModel: {},
      pageRequestModel: {
        needTotal: true,
        pageSize: 50,
        pageIndex: 1,
        skipCount: 0,
      },
    })
    orderTagOptions.value = Array.isArray(result?.list)
      ? result.list.map(item => ({ value: String(item.id), label: item.name }))
      : []
  }
  catch (error) {
    console.error('加载订单标签失败:', error)
  }
}

// 计算抽屉宽度
const drawerWidth = computed(() => {
  const width = windowWidth.value
  if (width <= 768) {
    return '100%'
  } else if (width <= 1024) {
    return '90%'
  } else if (width <= 1440) {
    return '800px'
  } else {
    return '800px'
  }
})

watch(openConfirmDrawer, (opened) => {
  if (opened && !confirmFormState.payDate) {
    confirmFormState.payDate = dayjs()
    accountFormRefs.value?.clearValidate?.(['account', 'payDate'])
  }
})

watch(() => confirmFormState.account, () => {
  accountFormRefs.value?.clearValidate?.('account')
})

watch(() => confirmFormState.payDate, () => {
  accountFormRefs.value?.clearValidate?.('payDate')
})
</script>

<template>
  <div>
    <a-drawer v-model:open="openDrawer" :body-style="{ padding: '0', background: '#f6f7f8' }" :push="false"
      :closable="false" :width="drawerWidth" placement="right" @close="closeFun">
      <!-- 自定义头部 -->
      <template #title>
        <div class="custom-header flex justify-between h-4 flex-items-center">
          <div class="text-5">
            账户充值
          </div>
          <a-button type="text" class="close-btn" @click="closeFun">
            <template #icon>
              <CloseOutlined class="text-5 close-icon" />
            </template>
          </a-button>
        </div>
      </template>
      <a-form ref="formRef" layout="vertical" :model="formState" :rules="rules">
        <div class=" py4 bg-white">
          <div class="stuSelectBox px8 border border-b-#eee border-solid border-x-none border-t-none"
            :class="formState.studentId ? 'pb4' : ''">
            <a-form-item v-if="!formState.studentId" class="selectStu" name="studentId">
              <div class="flex flex-items-center">
                <span class="flex flex-items-center whitespace-nowrap"> <span class="text-#f03 mr1"
                    style="font-family: SimSun, sans-serif;">*</span> 搜索/选择学员：</span>
                <div class="selectBox flex flex-items-center">
                  <div>
                    <StudentSelect ref="studentSelectRef" v-model="formState.studentId" placeholder="搜索姓名/手机号"
                      width="300px" allow-clear @select="handleStudentSelect" @change="handleChange" />
                  </div>
                  <a-button type="primary" class="ml3" ghost @click="handleCreateNewStudent">
                    创建新学员
                  </a-button>
                </div>
              </div>
            </a-form-item>
            <div v-if="formState.studentId && selectedStudent" class="flex  justify-between">
              <div class="flex bg-#fafafa pr4 rounded-10 flex-items-center">
                <img width="40" class="rounded-10" height="40"
                  :src="selectedStudent.avatarUrl || 'https://cdn.schoolpal.cn/schoolpal/next-erp/avator_female.png?x-oss-process=image/resize,w_120'"
                  alt="">
                <span class="ml2 mr4 text-5 font-500">{{ selectedStudent.stuName || '未知' }}</span>
                <span class="text-4 mr4">{{ formatMobile(selectedStudent.mobile) || '未知' }}</span>
                <span 
                  class="text-3 px2 py1 rounded-10"
                  :class="{
                    'bg-#e6f0ff text-#06f': selectedStudent.studentStatus === 1,
                    'bg-#fff7e6 text-#fa8c16': selectedStudent.studentStatus === 0,
                    'bg-#f5f5f5 text-#999': selectedStudent.studentStatus === 2
                  }"
                >
                  {{ selectedStudent.studentStatus === 1 ? '在读学员' : selectedStudent.studentStatus === 2 ? '历史学员' : '意向学员' }}
                </span>
              </div>
              <a-button type="primary" class="ml3" ghost @click="changeStudent">
                更换学员
              </a-button>
            </div>
          </div>
          <div v-if="formState.studentId && selectedStudent"
            class="linkAccount px8 pt4 flex justify-between flex-center">
            <div class="text-3.5">
              <span class="text-#888">关联储值账户：<span class="text-#222 text-4 font-500">{{ selectedStudent.rechargeAccountName || selectedStudent.rechargeAccountId || '-' }}</span> </span>
            </div>
            <div class="text-3 flex-col flex flex-items-end">
              <div class="text-#222">
                账户余额：<span>¥</span> <span class="text-6"
                  style="font-family: 'DIN Alternate', DINAlternate, sans-serif;">{{
                    formatMoney2(selectedStudent.accountBalance) }}</span>
              </div>
              <div class="text-#888 line-height-2 flex">
               <div class="mr-12px">
                含赠送余额：<span>¥</span> <span class="text-3.5"
                  style="font-family: 'DIN Alternate', DINAlternate, sans-serif;">{{
                    formatMoney2(selectedStudent.giftBalance) }}</span>
               </div>
               <div>
                含残联余额：<span>¥</span> <span class="text-3.5"
                  style="font-family: 'DIN Alternate', DINAlternate, sans-serif;">{{
                    formatMoney2(selectedStudent.residualBalance) }}</span>
               </div>
              </div>
            </div>
          </div>
        </div>
        <div class="middleBox mt2 py6 pb2 px8 bg-white">
          <div>
            <div class="lebal text-#666">
              充值金额：
            </div>
            <div class="payPrice h-20 border border-b-#eee border-solid border-x-none border-t-none"
              :class="{ 'animate-border': checkPayPrice }">
              <a-input-number v-model:value="formState.payPrice" :bordered="false" :controls="false"
                class="h-100% w-100% text-12" :min="0" :max="100000" :precision="2" placeholder="输入充值金额">
                <template #addonBefore>
                  <span class="text-12">¥</span>
                </template>
              </a-input-number>
              <span v-if="checkPayPriceLine" class="text-4 text-#f33 line-height-6">请输入充值金额、残联金额或赠送金额</span>
            </div>
          </div>
          <div class="mt7">
            <div class="lebal text-#666">
              残联金额：
            </div>
            <div class="payPrice h-20 border border-b-#eee border-solid border-x-none border-t-none"
              :class="{ 'animate-border': checkPayPrice }">
              <a-input-number v-model:value="formState.cldPayPrice" :bordered="false" :controls="false"
                class="h-100% w-100% text-12" :min="0" :max="100000" :precision="2" placeholder="输入残联金额">
                <template #addonBefore>
                  <span class="text-12">¥</span>
                </template>
              </a-input-number>
              <span v-if="checkPayPriceLine" class="text-4 text-#f33 line-height-6 ">请输入充值金额、残联金额或赠送金额</span>
            </div>
          </div>
          <div class="mt7">
            <div class="lebal text-#666">
              赠送金额：
            </div>
            <div class="payPrice h-20 border border-b-#eee border-solid border-x-none border-t-none"
              :class="{ 'animate-border': checkPayPrice }">
              <a-input-number v-model:value="formState.givePayPrice" :bordered="false" :controls="false"
                class="h-100% w-100% text-12" :min="0" :max="100000" :precision="2" placeholder="输入赠送金额">
                <template #addonBefore>
                  <span class="text-12">¥</span>
                </template>
              </a-input-number>
              <span v-if="checkPayPriceLine" class="text-4 text-#f33 line-height-6 ">请输入充值金额、残联金额或赠送金额</span>
            </div>
          </div>
          <div class="mt6">
            <div class="flex justify-between w-70%">
              <a-form-item label="经办日期：">
                <a-date-picker v-model:value="formState.createTime" class="w-60" :allow-clear="false"
                  :disabled-date="disabledDate" format="YYYY-MM-DD" value-format="YYYY-MM-DD" />
              </a-form-item>
              <a-form-item label="订单销售员：">
                <StaffSelect v-model="formState.salesperson" placeholder="请选择销售员" width="240px"
                  @change="handleSalespersonChange" />
              </a-form-item>
            </div>
          </div>
        </div>
        <div class="middleBox mt2 py6 px8 bg-white">
          <div class="w-100%">
            <a-form-item label="订单标签：" class="w-100%">
              <a-select v-model:value="formState.orderTag" mode="multiple" style="width:100%"
                placeholder="请选择订单标签 (最多可选5个)" :options="orderTagOptions" />
            </a-form-item>
          </div>
          <div class="w-100%">
            <a-form-item label="对内备注：" class="w-100%">
              <a-textarea v-model:value="formState.inRemarks" placeholder="此备注仅内部员工可见"
                :auto-size="{ minRows: 2, maxRows: 5 }" />
            </a-form-item>
          </div>
          <div class="w-100%">
            <a-form-item label="对外备注：" class="w-100%">
              <a-textarea v-model:value="formState.outRemarks" placeholder="此备注打印时将显示"
                :auto-size="{ minRows: 2, maxRows: 5 }" />
            </a-form-item>
          </div>
        </div>
      </a-form>
      <template #footer>
        <div class="h-15 flex flex-center justify-end pr-8">
          <a-space :size="14">
            <a-button type="primary" class="h-12 w-35 text-5" @click="handleSubmit">
              确定提交
            </a-button>
          </a-space>
        </div>
      </template>
      <!-- 订单确认 -->
      <a-drawer v-model:open="openConfirmDrawer" :mask-closable="false" :keyboard="false"
        :body-style="{ padding: '0', background: '#fff' }" :closable="false" :width="drawerWidth" placement="right">
        <!-- 自定义头部 -->
        <template #title>
          <div class="custom-header flex justify-between h-4 flex-items-center">
            <div class="text-5">
              订单确认
            </div>
            <a-button type="text" class="close-btn" @click="closeConfirmDrawer">
              <template #icon>
                <CloseOutlined class="text-5 close-icon" />
              </template>
            </a-button>
          </div>
        </template>
        <span class="bg-#f6f7f8 flex h-2" />
        <div class="bg-white  mt-2 justify-start p-8 pb2">
          <div class="Yprice mb1">
            <span>*</span>订单确认：
          </div>
          <div class="Yprice-num ml4">
            ¥ {{ Number(currentRechargeOrderDetail?.amount ?? formState.payPrice ?? 0).toFixed(2) }}
          </div>
          <div class="mt6">
            <a-radio-group v-model:value="payType" button-style="solid">
              <a-radio-button value="1" class="px8">
                已收款只记账
              </a-radio-button>
              <a-radio-button value="2" class="px10">
                面对面收款
              </a-radio-button>
            </a-radio-group>
            <div class="payList mt-4">
              <div class="payList-title">
                <span>*</span>收款方式
                <span class="payList-tip ml-2">请选择</span>
              </div>
              <div class="pay">
                <a-radio-group v-model:value="pay" class="custom-radio">
                  <a-space :size="16" class="flex-wrap">
                    <label v-for="(item, index) in checkOptions" :key="index" class="pay-box"
                      :class="{ active: pay === item.id }">
                      <span> <img :src="item.img" alt=""> {{ item.label }}</span>
                      <a-radio :value="item.id" />
                    </label>
                  </a-space>
                </a-radio-group>
              </div>
            </div>
          </div>
        </div>
        <div class="bg-white p-8 pt4 pb0 ">
          <a-form ref="accountFormRefs" class="flex flex-col" layout="vertical" :model="confirmFormState">
            <div class="flex w-100%">
              <!-- 收款账户必选校验 -->
              <a-form-item label="收款账户" class="flex-1" name="account" :rules="[
                {
                  required: true,
                  message: '请选择收款账户',
                  trigger: 'change',
                },
              ]">
                <a-select v-model:value="confirmFormState.account" :allow-clear="false" placeholder="请选择收款账户" :options="accountList"
                  style="flex:1" />
              </a-form-item>
              <!-- 支付日期 -->
              <a-form-item label="支付日期" class="flex-1 pl10" name="payDate" :rules="[
                {
                  required: true,
                  message: '请选择支付日期',
                  trigger: 'change',
                },
              ]">
                <div class="flex flex-items-center week-wrap">
                  <a-date-picker v-model:value="confirmFormState.payDate" style="flex:1;" :allow-clear="false"
                    :disabled-date="disabledDate" format="YYYY-MM-DD" class="w-35"
                    placeholder="请选择日期" />
                  <div class="week">
                    周一
                  </div>
                </div>
              </a-form-item>
            </div>
            <a-form-item label="账单备注（选填）">
              <a-textarea v-model:value="confirmFormState.billRemarks" placeholder="请输入内容，最多100字"
                :auto-size="{ minRows: 2, maxRows: 5 }" />
            </a-form-item>
          </a-form>
        </div>
        <div class="upload bg-white p-8 pt0 mt--4">
          <a-upload v-model:file-list="fileList" action="https://www.mocky.io/v2/5cc8019d300000980a055e76"
            list-type="picture-card" @preview="handlePreview">
            <div v-if="fileList.length < 3">
              <PlusOutlined class="text-6" />
            </div>
          </a-upload>
          <span class="text-#888">最多上传 3 张图片，支持 BMP / JPG / JPEG / PNG，单张图片不超过 4 MB</span>
        </div>
        <template #footer>
          <div class="h-15 flex flex-center justify-end pr-8">
            <a-space :size="14">
              <a-button type="primary" class="h-12 w-35 text-5" @click="handleOver">
                确定
              </a-button>
            </a-space>
          </div>
        </template>
      </a-drawer>
    </a-drawer>
    <a-modal v-model:open="openModal" centered :keyboard="false" :closable="false" :mask-closable="false" :width="380">
      <template #title>
        <div class="text-4">
          <ExclamationCircleOutlined class="mr2 text-#f90" /> 退出本次收款
        </div>
      </template>
      如您选择发送二维码让家长付款，请停留在此页面直到家长完成付款
      <template #footer>
        <a-button key="back" danger ghost @click="handleCancel">
          确定离开
        </a-button>
        <a-button key="submit" type="primary" ghost @click="handleOk">
          继续收款
        </a-button>
      </template>
    </a-modal>
    <a-modal :open="previewVisible" :title="previewTitle" :footer="null" @cancel="handleCancelImg">
      <img alt="example" style="width: 100%" :src="previewImage">
    </a-modal>
    <!-- 订单完成 -->
    <a-drawer v-model:open="openOverDrawer" :mask-closable="false" :keyboard="false"
      :body-style="{ padding: '0', background: '#fff' }" :closable="false" :width="drawerWidth" placement="right">
      <!-- 自定义头部 -->
      <template #title>
        <div class="custom-header flex justify-between h-4 flex-items-center">
          <div class="text-5">
            订单完成
          </div>
          <a-button type="text" class="close-btn" @click="closeOverDrawer">
            <template #icon>
              <CloseOutlined class="text-5 close-icon" />
            </template>
          </a-button>
        </div>
      </template>
      <span class="bg-#f6f7f8 flex h-2 " />
      <div class="flex justify-center border border-b-#eee border-solid border-x-none border-t-none">
        <a-result status="success">
          <template #title>
            <span class="text-8 text-#222" style="font-family: 'DIN alternate', sans-serif;">¥{{ Number(currentRechargeOrderDetail?.amount ?? formState.payPrice ?? 0).toFixed(2) }}</span>
          </template>
          <template #subTitle>
            <span class="text-4 text-#666">- 订单完成 -</span>
          </template>
          <template #extra>
            <a-button type="primary" @click="openRechargeSaleOrderDetailDrawer">
              查看订单详情
            </a-button>
            <a-dropdown>
              <template #overlay>
                <a-menu>
                  <a-menu-item key="0" @click="handlePrintReceipt">
                    打印收据
                  </a-menu-item>
                  <a-menu-item key="1" @click="handleDownloadReceipt">
                    下载收据
                  </a-menu-item>
                  <a-menu-item key="3">
                    发送短信
                  </a-menu-item>
                </a-menu>
              </template>
              <a-button>
                查看收据
                <DownOutlined :style="{ fontSize: '10px' }" />
              </a-button>
            </a-dropdown>
            <a-button @click="closePage">
              关闭页面
            </a-button>
          </template>
        </a-result>
      </div>
      <div class="px8 py6">
        <a-descriptions :column="3" :content-style="{ color: '#888' }">
          <a-descriptions-item label="储值账户：">
            {{ selectedStudent?.rechargeAccountName || selectedStudent?.rechargeAccountId || '-' }}
          </a-descriptions-item>
          <a-descriptions-item label="关联学员：">
            {{ currentRechargeOrderDetail?.studentName || selectedStudent?.stuName || '-' }}
          </a-descriptions-item>
          <a-descriptions-item label="收款方式：">
            {{ completionSnapshot.payMethodLabel || '-' }}
          </a-descriptions-item>
          <a-descriptions-item label="收款账户：">
            {{ completionSnapshot.accountLabel || '-' }}
          </a-descriptions-item>
          <a-descriptions-item label="支付单号：">
            -
          </a-descriptions-item>
          <a-descriptions-item label="对方账户：">
            -
          </a-descriptions-item>
          <a-descriptions-item label="支付时间：">
            {{ completionSnapshot.payDate || '-' }}
          </a-descriptions-item>
          <a-descriptions-item label="账单备注：">
            {{ completionSnapshot.billRemarks || '-' }}
          </a-descriptions-item>
        </a-descriptions>
      </div>
    </a-drawer>

    <!-- 创建学员Modal -->
    <CreateStudent
      ref="createStudentRef"
      v-model:open="openCreateStudent"
      :type="1"
      @submit="handleCreateStudentSuccess"
    />

    <order-detail-drawer v-model:open="openOrderDetailDrawer" :order-id="orderDetailDrawerOrderId" />
  </div>
</template>

<style lang="less" scoped>
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

.payPrice {
  :deep(.ant-input-number-input) {
    height: 80px;
    line-height: 80px;
    font-family: "DIN alternate", sans-serif;
  }

  :deep(.ant-input-number-group-addon) {
    background: transparent;
    border: none;
  }
}

.selectStu {
  :deep(.ant-form-item-explain-error) {
    padding-left: 115px;
  }
}

/* 动画关键帧 */
@keyframes borderExpand {
  0% {
    transform: scaleX(0);
    opacity: 0;
  }

  100% {
    transform: scaleX(1);
    opacity: 1;
  }
}

/* 动画容器 */
.animate-border {
  position: relative;
}

/* 动画线条 */
.animate-border::after {

  content: '';
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  height: 1px;
  background: #ff3333;
  transform-origin: center;
  /* 缩放中心点 */
  animation: borderExpand 0.4s cubic-bezier(0.68, -0.55, 0.27, 1.55) forwards;
}

.Yprice {
  font-size: 14px;

  span {
    color: red;
  }
}

.Yprice-num {
  padding-top: 12px;
  font-weight: 700;
  font-size: 48px;
  color: #000;
  line-height: 56px;
  font-family: "DIN alternate", sans-serif;
}

.payList {
  span {
    color: red;
  }

  span.payList-tip {
    color: var(--pro-ant-color-primary);
  }
}

.pay {
  margin-top: 10px;

  .pay-box {
    border: 1px solid #eee;
    padding: 12px 16px;
    display: flex;
    align-items: center;
    border-radius: 6px;
    user-select: none;
    cursor: pointer;

    &:hover {
      border-color: var(--pro-ant-color-primary);
    }

    span {
      color: #000;
      margin-right: 20px;
      display: flex;
      align-items: center;

      img {
        width: 20px;
        height: 20px;
        margin-right: 6px;
      }
    }
  }

  .active {
    border-color: var(--pro-ant-color-primary);
  }
}

.week-wrap {
  display: flex;

  .week {
    background: #e6f0ff;
    border-radius: 14px;
    color: var(--pro-ant-color-primary);
    font-size: 14px;
    font-weight: 400;
    line-height: 20px;
    margin-left: 8px;
    padding: 2px 14px;
  }
}

:deep(.ant-upload) {
  width: 78px !important;
  height: 78px !important;
}

:deep(.ant-upload-list-item-container) {
  width: 78px !important;
  height: 78px !important;
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

:deep(.ant-result-icon>.anticon) {
  color: #06f;
}
</style>
