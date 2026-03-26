<script setup>
import {
  CloseOutlined,
  ExclamationCircleOutlined,
} from '@ant-design/icons-vue'
import { computed, reactive, ref, watch } from 'vue'
import { message } from 'ant-design-vue'
import dayjs from 'dayjs'
import { getOrderTagListPagedApi } from '@/api/finance-center/order-tag'
import {
  createRechargeAccountRefundOrderApi,
  getRechargeAccountApi,
  getRechargeAccountOrderDetailApi,
  getStudentDetailApi,
  payOrderBySchoolPalApi,
} from '@/api/finance-center/recharge-account'
import StaffSelect from './staff-select.vue'
import { payMethodOptionsWithIcons } from './pay-method-options-data'

function payMethodImgById(id) {
  return payMethodOptionsWithIcons.find(o => o.id === id)?.img || ''
}

const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
  rechargeAccountId: {
    type: [String, Number],
    default: undefined,
  },
})
const emit = defineEmits(['update:open', 'submitted'])

const payType = ref('1')
const pay = ref(1)
/** 与充值抽屉一致：微信、支付宝、银行转账、POS机、现金、其他 */
const checkOptions = reactive([...payMethodOptionsWithIcons])

const openDrawer = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

const loadingAccount = ref(false)
const selectedStudent = ref(null)
const currentOrderDetail = ref(null)

const formRef = ref()
const initialFormState = {
  createTime: dayjs().format('YYYY-MM-DD'),
  refundAmount: undefined,
  cldRefundAmount: undefined,
  giveDeduct: undefined,
  salesperson: undefined,
  orderTag: [],
  inRemarks: undefined,
  outRemarks: undefined,
}
const formState = reactive({ ...initialFormState })

const checkRefundAmount = ref(false)
const checkRefundAmountLine = ref(false)

const openConfirmDrawer = ref(false)
const openModal = ref(false)

function disabledDate(current) {
  return current > dayjs().endOf('day')
}

function formatMobile(mobile) {
  if (!mobile)
    return ''
  if (mobile.length === 11)
    return mobile.replace(/(\d{3})\d{4}(\d{4})/, '$1****$2')
  return mobile
}

function formatMoney2(value) {
  return Number(value || 0).toFixed(2)
}

/** 可用总余额 = 接口 balance（充值+残联）+ 赠送 */
const totalUsableBalance = computed(() => {
  const s = selectedStudent.value
  if (!s)
    return 0
  return Number(s.accountBalance || 0) + Number(s.giftBalance || 0)
})

const rechargeBalanceCap = computed(() => {
  const s = selectedStudent.value
  if (!s)
    return 0
  return Math.min(100000, Math.max(0, Number(s.rechargeBalancePure ?? 0)))
})
const residualBalanceCap = computed(() => {
  const s = selectedStudent.value
  if (!s)
    return 0
  return Math.min(100000, Math.max(0, Number(s.residualBalance ?? 0)))
})
const giftBalanceCap = computed(() => {
  const s = selectedStudent.value
  if (!s)
    return 0
  return Math.min(100000, Math.max(0, Number(s.giftBalance ?? 0)))
})

/** 退费确认抽屉：实退/残联/赠送及完成后约剩余 */
const refundConfirmBreakdown = computed(() => {
  const s = selectedStudent.value
  const refundRecharge = Number(currentOrderDetail.value?.amount ?? formState.refundAmount ?? 0)
  const refundCld = Number(currentOrderDetail.value?.residualAmount ?? formState.cldRefundAmount ?? 0)
  const refundGift = Number(currentOrderDetail.value?.givingAmount ?? formState.giveDeduct ?? 0)
  const rechargePure = Number(s?.rechargeBalancePure ?? 0)
  const residual = Number(s?.residualBalance ?? 0)
  const gift = Number(s?.giftBalance ?? 0)
  return {
    refundRecharge,
    refundCld,
    refundGift,
    remainRecharge: s ? Math.max(0, rechargePure - refundRecharge) : null,
    remainCld: s ? Math.max(0, residual - refundCld) : null,
    remainGift: s ? Math.max(0, gift - refundGift) : null,
    hasStudent: !!s,
  }
})
const refundRequiresBill = computed(() => Number(currentOrderDetail.value?.amount ?? formState.refundAmount ?? 0) > 0)

function fillRefundAmountFull() {
  if (!selectedStudent.value)
    return
  formState.refundAmount = rechargeBalanceCap.value
}
function fillCldRefundFull() {
  if (!selectedStudent.value)
    return
  formState.cldRefundAmount = residualBalanceCap.value
}
function fillGiveDeductFull() {
  if (!selectedStudent.value)
    return
  formState.giveDeduct = giftBalanceCap.value
}

async function loadAccountAndStudent() {
  const id = props.rechargeAccountId
  if (!id) {
    selectedStudent.value = null
    return
  }
  loadingAccount.value = true
  try {
    const { result: acc } = await getRechargeAccountApi({ rechargeAccountId: String(id) })
    if (!acc?.id) {
      message.error('未找到储值账户')
      selectedStudent.value = null
      return
    }
    const mainId = acc.mainStudentId
    let detail = {}
    if (mainId) {
      try {
        const { result: stu } = await getStudentDetailApi({ studentId: mainId })
        detail = stu || {}
      }
      catch {
        detail = {}
      }
    }
    const mainStu = (acc.students || []).find(s => s.isMainStudent) || acc.students?.[0]
    const bal = Number(acc.balance || 0)
    const res = Number(acc.residualBalance || 0)
    const giv = Number(acc.givingBalance || 0)
    const rechargePure = Math.max(0, bal - res)
    selectedStudent.value = {
      stuName: mainStu?.name || detail?.name || '未知',
      mobile: mainStu?.phone || detail?.phone || '',
      avatarUrl: mainStu?.avatar || detail?.avatar,
      studentStatus: detail?.status ?? 0,
      salesPersonId: detail?.salespersonId || '0',
      rechargeAccountId: acc.id,
      rechargeAccountName: acc.accountName || acc.phone || acc.id,
      rechargeAccountPhone: acc.phone || '',
      mainStudentId: acc.mainStudentId || '',
      accountBalance: bal,
      rechargeBalancePure: rechargePure,
      giftBalance: giv,
      residualBalance: res,
      collectorStaffId: detail?.collectorStaffId || '0',
      phoneSellStaffId: detail?.phoneSellStaffId || '0',
      foregroundStaffId: detail?.foregroundStaffId || '0',
      viceSellStaffStaffId: detail?.viceSellStaffStaffId || '0',
    }
    formState.salesperson = selectedStudent.value.salesPersonId && selectedStudent.value.salesPersonId !== '0'
      ? selectedStudent.value.salesPersonId
      : undefined
  }
  catch (e) {
    console.error(e)
    message.error('加载储值账户失败')
    selectedStudent.value = null
  }
  finally {
    loadingAccount.value = false
  }
}

watch(
  () => [props.open, props.rechargeAccountId],
  ([isOpen]) => {
    if (isOpen) {
      openConfirmDrawer.value = false
      openModal.value = false
      Object.assign(formState, {
        ...initialFormState,
        createTime: dayjs().format('YYYY-MM-DD'),
      })
      checkRefundAmount.value = false
      checkRefundAmountLine.value = false
      currentOrderDetail.value = null
      loadAccountAndStudent()
    }
  },
)

watch(
  [() => formState.refundAmount, () => formState.cldRefundAmount, () => formState.giveDeduct],
  ([a, b, c]) => {
    if (a || b || c) {
      checkRefundAmount.value = false
      checkRefundAmountLine.value = false
    }
  },
)

function closeFun() {
  Object.assign(formState, { ...initialFormState, createTime: dayjs().format('YYYY-MM-DD') })
  selectedStudent.value = null
  currentOrderDetail.value = null
  openDrawer.value = false
}

/** 仅「确定退费」时调用，避免提前建单产生待付款记录 */
async function createRefundOrderAndLoadDetail() {
  const tagIds = Array.isArray(formState.orderTag) ? formState.orderTag.map(String) : []
  const { result: orderResult } = await createRechargeAccountRefundOrderApi({
    rechargeAccountId: String(selectedStudent.value.rechargeAccountId),
    amount: Number(formState.refundAmount || 0),
    givingAmount: Number(formState.giveDeduct || 0),
    residualAmount: Number(formState.cldRefundAmount || 0),
    dealDate: formState.createTime,
    salePersonId: String(formState.salesperson || selectedStudent.value.salesPersonId || '0'),
    collectorStaffId: String(selectedStudent.value.collectorStaffId || '0'),
    phoneSellStaffId: String(selectedStudent.value.phoneSellStaffId || '0'),
    foregroundStaffId: String(selectedStudent.value.foregroundStaffId || '0'),
    viceSellStaffStaffId: String(selectedStudent.value.viceSellStaffStaffId || '0'),
    remark: formState.inRemarks || '',
    orderTagIds: tagIds,
    externalRemark: formState.outRemarks || '',
    studentId: String(selectedStudent.value.mainStudentId || ''),
  })
  const orderId = orderResult?.id
  if (!orderId)
    throw new Error('创建退费订单失败')
  const { result: orderDetail } = await getRechargeAccountOrderDetailApi({
    rechargeAccountOrderId: String(orderId),
  })
  return orderDetail
}

async function handleSubmit() {
  if (!selectedStudent.value?.rechargeAccountId) {
    message.error('储值账户未就绪')
    return
  }
  if (!formState.refundAmount && !formState.cldRefundAmount && !formState.giveDeduct) {
    checkRefundAmount.value = true
    setTimeout(() => {
      checkRefundAmountLine.value = true
    }, 400)
    return
  }
  checkRefundAmount.value = false
  checkRefundAmountLine.value = false

  const tagIds = Array.isArray(formState.orderTag) ? formState.orderTag.map(String) : []
  if (tagIds.length > 5) {
    message.error('订单标签最多可选5个')
    return
  }

  currentOrderDetail.value = null
  openConfirmDrawer.value = true
  openDrawer.value = false
}

const accountFormRefs = ref()
const confirmFormState = reactive({
  account: 1,
  payDate: undefined,
  billRemarks: undefined,
})
const accountList = ref([{ value: 1, label: '默认账户' }])

const submittingRefund = ref(false)

async function handleOver() {
  if (!selectedStudent.value?.rechargeAccountId) {
    message.error('储值账户未就绪')
    return
  }
  if (refundRequiresBill.value) {
    try {
      await accountFormRefs.value?.validate()
    }
    catch {
      return
    }
  }
  const tagIds = Array.isArray(formState.orderTag) ? formState.orderTag.map(String) : []
  if (tagIds.length > 5) {
    message.error('订单标签最多可选5个')
    return
  }

  submittingRefund.value = true
  try {
    let orderDetail = currentOrderDetail.value
    if (!orderDetail?.id) {
      orderDetail = await createRefundOrderAndLoadDetail()
      currentOrderDetail.value = orderDetail
    }
    const billId = Number(orderDetail?.bill?.id || 0)
    if (!refundRequiresBill.value || billId <= 0) {
      message.success('账户退费成功')
      openConfirmDrawer.value = false
      Object.assign(confirmFormState, {
        account: 1,
        payDate: undefined,
        billRemarks: undefined,
      })
      emit('submitted')
      closeFun()
      return
    }
    await payOrderBySchoolPalApi({
      billId: String(orderDetail.bill.id),
      amount: Number(formState.refundAmount || 0),
      remark: confirmFormState.billRemarks || '',
      payMethod: pay.value,
      amountId: Number(confirmFormState.account || 0),
      payTime: confirmFormState.payDate
        ? dayjs(confirmFormState.payDate)
            .hour(dayjs().hour())
            .minute(dayjs().minute())
            .second(dayjs().second())
            .format('YYYY-MM-DD HH:mm:ss')
        : undefined,
    })
    message.success('账户退费成功')
    openConfirmDrawer.value = false
    Object.assign(confirmFormState, {
      account: 1,
      payDate: undefined,
      billRemarks: undefined,
    })
    emit('submitted')
    closeFun()
  }
  catch (e) {
    console.error(e)
    message.error(e?.response?.data?.message || e?.message || '退费失败')
  }
  finally {
    submittingRefund.value = false
  }
}

function handleCancel() {
  openModal.value = false
  openConfirmDrawer.value = false
  closeFun()
}

function handleOk() {
  openModal.value = false
}

function closeConfirmDrawer() {
  openModal.value = true
}

const orderTagOptions = ref([])
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
  catch (e) {
    console.error(e)
  }
}

fetchOrderTagOptions()

const windowWidth = ref(typeof window !== 'undefined' ? window.innerWidth : 1200)
const drawerWidth = computed(() => {
  const w = windowWidth.value
  if (w <= 768)
    return '100%'
  if (w <= 1024)
    return '90%'
  return '800px'
})

watch(openConfirmDrawer, (opened) => {
  if (opened && !confirmFormState.payDate) {
    confirmFormState.payDate = dayjs()
    accountFormRefs.value?.clearValidate?.(['account', 'payDate'])
  }
})

function onOrderTagChange(val) {
  if (Array.isArray(val) && val.length > 5)
    formState.orderTag = val.slice(0, 5)
}
</script>

<template>
  <div>
    <a-drawer
      v-model:open="openDrawer"
      :body-style="{ padding: '0', background: '#f6f7f8' }"
      :push="false"
      :closable="false"
      :width="drawerWidth"
      placement="right"
      @close="closeFun"
    >
      <template #title>
        <div class="custom-header flex justify-between h-4 flex-items-center">
          <div class="text-5">
            账户退费
          </div>
          <a-button type="text" class="close-btn" @click="closeFun">
            <template #icon>
              <CloseOutlined class="text-5 close-icon" />
            </template>
          </a-button>
        </div>
      </template>

      <a-spin :spinning="loadingAccount">
        <a-form ref="formRef" layout="vertical" :model="formState">
          <div class="py4 bg-white">
            <div
              class="stuSelectBox px8 border border-b-#eee border-solid border-x-none border-t-none"
              :class="selectedStudent ? 'pb4' : ''"
            >
              <div v-if="selectedStudent" class="flex justify-between">
                <div class="flex bg-#fafafa pr4 rounded-10 flex-items-center">
                  <img
                    width="40"
                    class="rounded-10"
                    height="40"
                    :src="selectedStudent.avatarUrl || 'https://cdn.schoolpal.cn/schoolpal/next-erp/avator_female.png?x-oss-process=image/resize,w_120'"
                    alt=""
                  >
                  <span class="ml2 mr4 text-5 font-500">{{ selectedStudent.stuName || '未知' }}</span>
                  <span class="text-4 mr4">{{ formatMobile(selectedStudent.mobile) || '未知' }}</span>
                  <span
                    class="text-3 px2 py1 rounded-10"
                    :class="{
                      'bg-#e6f0ff text-#06f': selectedStudent.studentStatus === 1,
                      'bg-#fff7e6 text-#fa8c16': selectedStudent.studentStatus === 0,
                      'bg-#f5f5f5 text-#999': selectedStudent.studentStatus === 2,
                    }"
                  >
                    {{ selectedStudent.studentStatus === 1 ? '在读学员' : selectedStudent.studentStatus === 2 ? '历史学员' : '意向学员' }}
                  </span>
                </div>
              </div>
            </div>
            <div
              v-if="selectedStudent"
              class="linkAccount px8 pt4 flex justify-between flex-center"
            >
              <div class="text-3.5">
                <span class="text-#888">关联储值账户：<span class="text-#222 text-4 font-500">{{ selectedStudent.rechargeAccountName || selectedStudent.rechargeAccountPhone || selectedStudent.rechargeAccountId || '-' }}</span></span>
              </div>
              <div class="text-3 flex-col flex flex-items-end">
                <div class="text-#222">
                  可用总余额：<span>¥</span> <span class="text-6" style="font-family: 'DIN Alternate', DINAlternate, sans-serif;">{{
                    formatMoney2(totalUsableBalance) }}</span>
                </div>
                <div class="text-#888 line-height-2 flex flex-wrap justify-end gap-x-12px">
                  <div>
                    含残联余额：<span>¥</span> <span class="text-3.5" style="font-family: 'DIN Alternate', DINAlternate, sans-serif;">{{
                      formatMoney2(selectedStudent.residualBalance) }}</span>
                  </div>
                  <div>
                    含赠送余额：<span>¥</span> <span class="text-3.5" style="font-family: 'DIN Alternate', DINAlternate, sans-serif;">{{
                      formatMoney2(selectedStudent.giftBalance) }}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
          <div class="middleBox mt2 py6 pb2 px8 bg-white">
            <div class="refund-amount-item w-full">
              <div class="lebal text-#666">
                充值退款：
              </div>
              <div class="payPrice refund-amount-input w-full">
                <div
                  class="refund-input-surface h-20 w-full"
                  :class="{ 'animate-border': checkRefundAmount }"
                >
                  <a-input-number
                    v-model:value="formState.refundAmount"
                    :bordered="false"
                    :controls="false"
                    class="h-100% w-100% text-12"
                    :min="0"
                    :max="rechargeBalanceCap"
                    :precision="2"
                    placeholder="输入退费金额"
                  >
                    <template #addonBefore>
                      <span class="text-12">¥</span>
                    </template>
                  </a-input-number>
                </div>
                <div
                  v-if="selectedStudent"
                  class="refund-amount-footer mt-2 mb-2 flex flex-items-center justify-between text-3.5 px1 pb-1"
                >
                  <span class="text-#888">充值余额 <span class="text-#222 font-500">¥ {{ formatMoney2(selectedStudent.rechargeBalancePure) }}</span></span>
                  <a
                    class="text-#06f cursor-pointer select-none"
                    :class="{ 'pointer-events-none text-#ccc': rechargeBalanceCap <= 0 }"
                    @click="fillRefundAmountFull"
                  >全部</a>
                </div>
              </div>
            </div>
            <span class="refund-item-sep block h-2 bg-#f6f7f8" aria-hidden="true" />
            <div class="refund-amount-item w-full">
              <div class="lebal text-#666 mt-2">
                残联金额：
              </div>
              <div class="payPrice refund-amount-input w-full">
                <div
                  class="refund-input-surface h-20 w-full"
                  :class="{ 'animate-border': checkRefundAmount }"
                >
                  <a-input-number
                    v-model:value="formState.cldRefundAmount"
                    :bordered="false"
                    :controls="false"
                    class="h-100% w-100% text-12"
                    :min="0"
                    :max="residualBalanceCap"
                    :precision="2"
                    placeholder="输入残联退费金额"
                  >
                    <template #addonBefore>
                      <span class="text-12">¥</span>
                    </template>
                  </a-input-number>
                </div>
                <div
                  v-if="selectedStudent"
                  class="refund-amount-footer mt-2 mb-2 flex flex-items-center justify-between text-3.5 px1 pb-1"
                >
                  <span class="text-#888">残联余额 <span class="text-#222 font-500">¥ {{ formatMoney2(selectedStudent.residualBalance) }}</span></span>
                  <a
                    class="text-#06f cursor-pointer select-none"
                    :class="{ 'pointer-events-none text-#ccc': residualBalanceCap <= 0 }"
                    @click="fillCldRefundFull"
                  >全部</a>
                </div>
              </div>
            </div>
            <span class="refund-item-sep block h-2 bg-#f6f7f8" aria-hidden="true" />
            <div class="refund-amount-item w-full">
              <div class="lebal text-#666 mt-2">
                赠送扣减：
              </div>
              <div class="payPrice refund-amount-input w-full">
                <div
                  class="refund-input-surface h-20 w-full"
                  :class="{ 'animate-border': checkRefundAmount }"
                >
                  <a-input-number
                    v-model:value="formState.giveDeduct"
                    :bordered="false"
                    :controls="false"
                    class="h-100% w-100% text-12"
                    :min="0"
                    :max="giftBalanceCap"
                    :precision="2"
                    placeholder="请输入赠送扣减金额"
                  >
                    <template #addonBefore>
                      <span class="text-12">¥</span>
                    </template>
                  </a-input-number>
                </div>
                <div
                  v-if="selectedStudent"
                  class="refund-amount-footer mt-2 mb-2 flex flex-items-center justify-between text-3.5 px1 pb-1"
                >
                  <span class="text-#888">赠送余额 <span class="text-#222 font-500">¥ {{ formatMoney2(selectedStudent.giftBalance) }}</span></span>
                  <a
                    class="text-#06f cursor-pointer select-none"
                    :class="{ 'pointer-events-none text-#ccc': giftBalanceCap <= 0 }"
                    @click="fillGiveDeductFull"
                  >全部</a>
                </div>
              </div>
            </div>
            <span v-if="checkRefundAmountLine" class="mb-4 block text-4 text-#f33 line-height-6">请输入退费金额、残联金额或赠送扣减金额</span>
            <span class="refund-item-sep block h-2 bg-#f6f7f8" aria-hidden="true" />
            <div class="w-full pt-8">
              <div class="flex w-full flex-wrap gap-x-8 gap-y-4">
                <a-form-item class="min-w-0 flex-1" label="经办日期：">
                  <a-date-picker
                    v-model:value="formState.createTime"
                    class="w-full"
                    :allow-clear="false"
                    :disabled-date="disabledDate"
                    format="YYYY-MM-DD"
                    value-format="YYYY-MM-DD"
                  />
                </a-form-item>
                <a-form-item class="min-w-0 flex-1" label="订单销售员：">
                  <StaffSelect
                    v-model="formState.salesperson"
                    placeholder="请选择"
                    width="100%"
                  />
                </a-form-item>
              </div>
            </div>
          </div>
          <div class="middleBox mt2 py6 px8 bg-white">
            <div class="w-100%">
              <a-form-item label="订单标签：" class="w-100%">
                <a-select
                  v-model:value="formState.orderTag"
                  mode="multiple"
                  style="width:100%"
                  placeholder="请选择订单标签（最多可选5个）"
                  :options="orderTagOptions"
                  @change="onOrderTagChange"
                />
              </a-form-item>
            </div>
            <div class="w-100%">
              <a-form-item label="对内备注：" class="w-100%">
                <a-textarea
                  v-model:value="formState.inRemarks"
                  placeholder="此备注仅内部员工可见"
                  :auto-size="{ minRows: 2, maxRows: 5 }"
                />
              </a-form-item>
            </div>
            <div class="w-100%">
              <a-form-item label="对外备注：" class="w-100%">
                <a-textarea
                  v-model:value="formState.outRemarks"
                  placeholder="此备注打印时将显示"
                  :auto-size="{ minRows: 2, maxRows: 5 }"
                />
              </a-form-item>
            </div>
          </div>
        </a-form>
      </a-spin>
      <template #footer>
        <div class="h-15 flex flex-center justify-end pr-8">
          <a-button type="primary" class="h-12 w-35 text-5" :disabled="loadingAccount" @click="handleSubmit">
            确定提交
          </a-button>
        </div>
      </template>
    </a-drawer>

    <a-drawer
      v-model:open="openConfirmDrawer"
      :mask-closable="false"
      :keyboard="false"
      :body-style="{ padding: '0', background: '#fff' }"
      :closable="false"
      :width="drawerWidth"
      placement="right"
    >
      <template #title>
        <div class="custom-header flex justify-between h-4 flex-items-center">
          <div class="text-5">
            退费确认
          </div>
          <a-button type="text" class="close-btn" @click="closeConfirmDrawer">
            <template #icon>
              <CloseOutlined class="text-5 close-icon" />
            </template>
          </a-button>
        </div>
      </template>
      <span class="bg-#f6f7f8 flex h-2" />
      <div class="bg-white mt-2 justify-start p-8 pb2">
        <div class="Yprice mb1">
          <span>*</span>实退金额（充值部分）
        </div>
        <div class="Yprice-num refund-hero-amount ml4">
          ¥ {{ formatMoney2(refundConfirmBreakdown.refundRecharge) }}
        </div>

        <div class="refund-breakdown-grid">
          <div class="refund-bd-card">
            <img class="refund-bd-icon" :src="payMethodImgById(3)" alt="">
            <div class="refund-bd-body">
              <div class="refund-bd-label">
                退充值金额
              </div>
              <div class="refund-bd-note">
                从账户充值余额扣减
              </div>
              <div class="refund-bd-val minus tabular-nums">
                ¥ {{ formatMoney2(refundConfirmBreakdown.refundRecharge) }}
              </div>
            </div>
          </div>
          <div class="refund-bd-card">
            <img class="refund-bd-icon" :src="payMethodImgById(5)" alt="">
            <div class="refund-bd-body">
              <div class="refund-bd-label">
                清除残联金额
              </div>
              <div class="refund-bd-note">
                从残联余额扣减
              </div>
              <div class="refund-bd-val minus tabular-nums">
                ¥ {{ formatMoney2(refundConfirmBreakdown.refundCld) }}
              </div>
            </div>
          </div>
          <div class="refund-bd-card">
            <img class="refund-bd-icon" :src="payMethodImgById(6)" alt="">
            <div class="refund-bd-body">
              <div class="refund-bd-label">
                清除赠送金额
              </div>
              <div class="refund-bd-note">
                从赠送余额扣减
              </div>
              <div class="refund-bd-val minus tabular-nums">
                ¥ {{ formatMoney2(refundConfirmBreakdown.refundGift) }}
              </div>
            </div>
          </div>
        </div>

        <div v-if="refundConfirmBreakdown.hasStudent" class="refund-remain-panel">
          <div class="refund-remain-title">
            完成后账户剩余
          </div>
          <div class="refund-remain-cols">
            <div class="refund-remain-cell">
              <span class="refund-remain-k">充值</span>
              <span class="refund-remain-v tabular-nums">¥ {{ formatMoney2(refundConfirmBreakdown.remainRecharge) }}</span>
            </div>
            <div class="refund-remain-cell">
              <span class="refund-remain-k">残联</span>
              <span class="refund-remain-v tabular-nums">¥ {{ formatMoney2(refundConfirmBreakdown.remainCld) }}</span>
            </div>
            <div class="refund-remain-cell">
              <span class="refund-remain-k">赠送</span>
              <span class="refund-remain-v tabular-nums">¥ {{ formatMoney2(refundConfirmBreakdown.remainGift) }}</span>
            </div>
          </div>
        </div>
        <div v-else class="refund-remain-panel refund-remain-panel--muted">
          <div class="refund-remain-title refund-remain-title--muted">
            加载账户信息后显示完成后约剩余
          </div>
        </div>

        <div v-if="refundRequiresBill" class="mt6">
          <a-radio-group v-model:value="payType" button-style="solid">
            <a-radio-button value="1" class="px8">
              已退款只记账
            </a-radio-button>
            <a-radio-button value="2" class="px10">
              面对面退款
            </a-radio-button>
          </a-radio-group>
          <div class="payList mt-4">
            <div class="payList-title">
              <span>*</span>退款方式
              <span class="payList-tip ml-2">请选择</span>
            </div>
            <div class="pay">
              <a-radio-group v-model:value="pay" class="custom-radio">
                <a-space :size="16" class="flex-wrap">
                  <label
                    v-for="(item, index) in checkOptions"
                    :key="index"
                    class="pay-box"
                    :class="{ active: pay === item.id }"
                  >
                    <span><img v-if="item.img" :src="item.img" alt="">{{ item.label }}</span>
                    <a-radio :value="item.id" />
                  </label>
                </a-space>
              </a-radio-group>
            </div>
          </div>
        </div>
        <div v-else class="refund-no-ledger-tip mt6">
          实退金额为 0，本次仅扣减账户余额，不生成账单记录，也无需填写退款方式、退款账户和退款日期。
        </div>
      </div>
      <div v-if="refundRequiresBill" class="bg-white p-8 pt4 pb0">
        <a-form ref="accountFormRefs" class="flex flex-col" layout="vertical" :model="confirmFormState">
          <div class="flex w-100%">
            <a-form-item
              label="退款账户"
              class="flex-1"
              name="account"
              :rules="[{ required: true, message: '请选择退款账户', trigger: 'change' }]"
            >
              <a-select
                v-model:value="confirmFormState.account"
                :allow-clear="false"
                placeholder="请选择"
                :options="accountList"
                style="flex:1"
              />
            </a-form-item>
            <a-form-item
              label="退款日期"
              class="flex-1 pl10"
              name="payDate"
              :rules="[{ required: true, message: '请选择日期', trigger: 'change' }]"
            >
              <a-date-picker
                v-model:value="confirmFormState.payDate"
                style="flex:1;"
                :allow-clear="false"
                :disabled-date="disabledDate"
                format="YYYY-MM-DD"
                class="w-35"
                placeholder="请选择日期"
              />
            </a-form-item>
          </div>
          <a-form-item label="账单备注（选填）">
            <a-textarea
              v-model:value="confirmFormState.billRemarks"
              placeholder="请输入内容，最多100字"
              :auto-size="{ minRows: 2, maxRows: 5 }"
            />
          </a-form-item>
        </a-form>
      </div>
      <template #footer>
        <div class="h-15 flex flex-center justify-end pr-8">
          <a-button
            type="primary"
            class="h-12 w-35 text-5"
            :loading="submittingRefund"
            @click="handleOver"
          >
            {{ refundRequiresBill ? '确定退费' : '确认扣减' }}
          </a-button>
        </div>
      </template>
    </a-drawer>

    <a-modal v-model:open="openModal" centered :keyboard="false" :closable="false" :mask-closable="false" :width="380">
      <template #title>
        <div class="text-4">
          <ExclamationCircleOutlined class="mr2 text-#f90" /> 退出本次退费
        </div>
      </template>
      确定要离开退费确认吗？
      <template #footer>
        <a-button key="back" danger ghost @click="handleCancel">
          确定离开
        </a-button>
        <a-button key="submit" type="primary" ghost @click="handleOk">
          继续
        </a-button>
      </template>
    </a-modal>
  </div>
</template>

<style lang="less" scoped>
.close-btn {
  &:hover {
    background: transparent;
  }
}
.refund-amount-item {
  width: 100%;
  max-width: 100%;
  box-sizing: border-box;
}

/* 与确认抽屉 `bg-#f6f7f8 h-2` 一致，横向铺满 white 区块（抵消 middleBox 的 px8） */
.refund-item-sep {
  width: calc(100% + 4rem);
  max-width: none;
  margin-left: -2rem;
  margin-right: -2rem;
  flex-shrink: 0;
}

.refund-no-ledger-tip {
  padding: 12px 16px;
  border-radius: 12px;
  background: #f6f8fc;
  color: #5c6470;
  line-height: 1.7;
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

.refund-amount-input {
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  :deep(.ant-input-number-group) {
    width: 100%;
    display: table;
  }
  :deep(.ant-input-number) {
    width: 100%;
  }
}

.refund-input-surface {
  flex-shrink: 0;
  /* 无 bordered 时补回输入框下划线（与 Ant Design 默认分割线一致） */
  border-bottom: 1px solid #d9d9d9;
  transition: border-color 0.2s ease;

  &:focus-within:not(.animate-border) {
    border-bottom-color: var(--pro-ant-color-success, #52c41a);
  }

  /* 校验红线由 ::after 动画展示，底边暂隐避免双线 */
  &.animate-border {
    border-bottom-color: transparent;
  }
}

.refund-amount-footer {
  flex-shrink: 0;
}
@keyframes borderExpand {
  0% { transform: scaleX(0); opacity: 0; }
  100% { transform: scaleX(1); opacity: 1; }
}
.animate-border {
  position: relative;
}
.animate-border::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  height: 1px;
  background: #ff3333;
  transform-origin: center;
  animation: borderExpand 0.4s cubic-bezier(0.68, -0.55, 0.27, 1.55) forwards;
}
.Yprice {
  font-size: 14px;
  span { color: red; }
}
.Yprice-num {
  padding-top: 12px;
  font-weight: 700;
  font-size: 48px;
  color: #000;
  line-height: 56px;
  font-family: "DIN alternate", sans-serif;
}
.refund-hero-amount {
  color: #f5222d;
}
.refund-breakdown-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 12px;
  margin-top: 24px;
}
.refund-bd-card {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 14px 16px;
  background: #fafafa;
  border: 1px solid #f0f0f0;
  border-radius: 8px;
}
.refund-bd-icon {
  width: 28px;
  height: 28px;
  flex-shrink: 0;
  margin-top: 2px;
}
.refund-bd-body {
  min-width: 0;
  flex: 1;
}
.refund-bd-label {
  font-size: 14px;
  font-weight: 600;
  color: #222;
  line-height: 20px;
}
.refund-bd-note {
  font-size: 12px;
  color: #888;
  margin-top: 4px;
  line-height: 18px;
}
.refund-bd-val {
  margin-top: 8px;
  font-size: 20px;
  font-weight: 700;
  font-family: "DIN alternate", sans-serif;
  line-height: 26px;
  &.minus {
    color: #f5222d;
  }
}
.refund-remain-panel {
  margin-top: 12px;
  padding: 8px 10px 10px;
  background: linear-gradient(135deg, #f6f9ff 0%, #fafbff 100%);
  border: 1px solid #e6f0ff;
  border-radius: 6px;
}
.refund-remain-panel--muted {
  background: #fafafa;
  border-color: #eee;
  padding: 8px 10px;
}
.refund-remain-title {
  font-size: 12px;
  font-weight: 600;
  color: #555;
  margin-bottom: 6px;
  line-height: 18px;
}
.refund-remain-title--muted {
  font-weight: 400;
  color: #999;
  margin-bottom: 0;
}
.refund-remain-cols {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  align-items: start;
  gap: 0 4px;
}
.refund-remain-cell {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
  min-width: 0;
  text-align: center;
  padding: 0 4px;
}
.refund-remain-cell:not(:first-child) {
  border-left: 1px solid rgba(24, 88, 255, 0.12);
}
.refund-remain-k {
  font-size: 11px;
  color: #888;
  line-height: 16px;
  white-space: nowrap;
}
.refund-remain-v {
  font-size: 13px;
  font-weight: 600;
  font-family: "DIN alternate", sans-serif;
  color: #222;
  line-height: 18px;
}
.payList {
  span { color: red; }
  span.payList-tip { color: var(--pro-ant-color-primary); }
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
    &:hover { border-color: var(--pro-ant-color-primary); }
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
  .active { border-color: var(--pro-ant-color-primary); }
}
</style>
