<script setup>
import { computed, h, ref, watch } from 'vue'
import {
  ExclamationCircleFilled,
  LeftOutlined,
  RightOutlined,
  DownOutlined,
} from '@ant-design/icons-vue'
import { getApprovalTemplatesApi } from '@/api/finance-center/approval-manage'
import { getOrderDetailApi } from '@/api/finance-center/order-manage'
import messageService from '@/utils/messageService'
import { openOrderReceiptPage } from '@/utils/order-receipt'

// Props
const props = defineProps({
  paymentData: {
    type: Object,
    default: () => ({}),
  },
})
const orderInfo = ref({})
const approvalEnabled = ref(false)
const loading = ref(false)
const openOrderDetailDrawer = ref(false)
// Emits
const emit = defineEmits([
  'go-back',
  'create-new-order',
  'view-order-detail',
])

const orderId = computed(() => String(props.paymentData?.orderId || ''))
const isApprovalPending = computed(() => approvalEnabled.value && orderInfo.value?.orderStatus === 2)
/** 支付成功（订单完成展示）时按钮区顶间距为 0，审批中/加载中等为 34px */
const payButtonMarginTop = computed(() =>
  !loading.value && !isApprovalPending.value ? '0' : '34px',
)
const amountText = computed(() => Number(props.paymentData?.payAmount || 0).toFixed(2))
const payMethodMap = {
  1: '微信',
  2: '支付宝',
  3: '银行转账',
  4: 'POS机',
  5: '现金',
  6: '其他',
}
const paymentMethodList = computed(() => {
  if (Array.isArray(props.paymentData?.paymentMethods) && props.paymentData.paymentMethods.length) {
    return props.paymentData.paymentMethods
  }
  return Array.isArray(orderInfo.value?.paymentRecords)
    ? orderInfo.value.paymentRecords.map(item => ({
        payTitle: payMethodMap[item.payMethod] || '-',
        payAmount: Number(item.payAmount || 0),
        accountName: item.accountName,
        paymentNo: item.paymentId,
        payTime: item.payTime,
      }))
    : []
})

function goBack() {
  emit('go-back')
}

function createNewOrder() {
  emit('create-new-order')
}

function viewOrderDetail() {
  if (!orderId.value) {
    return
  }
  openOrderDetailDrawer.value = true
}

function handlePrintReceipt() {
  if (!orderId.value) {
    messageService.warning('订单不存在')
    return
  }
  openOrderReceiptPage(orderId.value, { template: 'a4' })
}

function handleDownloadReceipt() {
  if (!orderId.value) {
    messageService.warning('订单不存在')
    return
  }
  openOrderReceiptPage(orderId.value, { template: 'a4', autoDownload: true })
}

function formatDateTime(value) {
  if (!value)
    return '-'
  return String(value).replace('T', ' ').slice(0, 16)
}

function formatDateOnly(value) {
  if (!value)
    return '-'
  return String(value).split('T')[0]
}

async function fetchApprovalEnabled() {
  try {
    const res = await getApprovalTemplatesApi()
    if (res.code === 200) {
      const list = Array.isArray(res.result) ? res.result : []
      approvalEnabled.value = !!list.find(item => item.type === 1 && item.enable)
      return
    }
  }
  catch (error) {
    console.error('获取审批模板失败:', error)
  }
  approvalEnabled.value = false
}

async function fetchOrderInfo() {
  if (!orderId.value) {
    orderInfo.value = {}
    return
  }
  try {
    loading.value = true
    const res = await getOrderDetailApi({ orderId: orderId.value })
    if (res.code === 200 && res.result) {
      orderInfo.value = res.result
      return
    }
    orderInfo.value = {}
  }
  catch (error) {
    console.error('获取订单详情失败:', error)
    messageService.error('获取订单详情失败')
    orderInfo.value = {}
  }
  finally {
    loading.value = false
  }
}

async function loadPageData() {
  await Promise.all([
    fetchApprovalEnabled(),
    fetchOrderInfo(),
  ])
}

watch(() => props.paymentData?.orderId, () => {
  loadPageData()
}, { immediate: true })
</script>

<template>
  <div class="current2">
    <div class="step bg-white rounded-4 mt-3 justify-start p-6">
      <div class="unfollowContainer flex flex-items-center">
        <ExclamationCircleFilled class="ExclamationCircleFilled" />
        <div class="unfollowText">
          此学员家长尚未关注家校平台
        </div>
        <div class="unfollowButtonContainer">
          点击邀请关注，推送家校通知
          <RightOutlined />
        </div>
      </div>
      <div v-if="loading" class="py-10 flex justify-center">
        <a-spin />
      </div>
      <div v-else class="patStatus">
        <template v-if="isApprovalPending">
          <div class="icon">
            <svg width="72px" height="72px" viewBox="0 0 72 72">
              <title>形状结合</title>
              <g id="\u9875\u9762-1" stroke="none" stroke-width="1" fill="none" fill-rule="evenodd">
                <g id="PC\u62A5\u540D\u7EED\u8D39" transform="translate(-694.000000, -288.000000)" fill="#FF9900">
                  <path
                    id="\u5F62\u72B6\u7ED3\u5408"
                    d="M730,288 C749.882251,288 766,304.117749 766,324 C766,343.882251 749.882251,360 730,360 C710.117749,360 694,343.882251 694,324 C694,304.117749 710.117749,288 730,288 Z M725.982064,302.866071 C724.206907,302.866071 722.767857,304.305132 722.767857,306.080302 L722.767857,306.080302 L722.767857,326.698561 L722.768382,326.772524 C722.803014,329.20975 724.546618,331.292687 726.947622,331.750368 L726.947622,331.750368 L742.897637,334.790775 L742.94993,334.800305 C744.674923,335.100229 746.327776,333.961588 746.656847,332.235253 L746.656847,332.235253 L746.666377,332.18296 C746.966299,330.457955 745.827666,328.805089 744.101344,328.476016 L744.101344,328.476016 L729.196271,325.634757 L729.196271,306.080302 L729.19584,306.027149 C729.167456,304.276494 727.73947,302.866071 725.982064,302.866071 Z"
                  />
                </g>
              </g>
            </svg>
          </div>
          <div class="payPrice">
            ¥{{ amountText }}
          </div>
          <div class="payTip">
            - 订单审批中（机构已开启订单审批）-
          </div>
        </template>
        <a-result v-else status="success">
          <template #title>
            <span class="payPrice">¥{{ amountText }}</span>
          </template>
          <template #subTitle>
            <span class="payment-summary">- 订单完成 -</span>
          </template>
        </a-result>
      </div>
      <div class="payButton" :style="{ marginTop: payButtonMarginTop }">
        <a-button :icon="h(LeftOutlined)" @click="goBack">
          返回
        </a-button>
        <a-button type="primary" class="ml4 mr4" @click="createNewOrder">
          再报一笔
        </a-button>
        <a-button @click="viewOrderDetail">
          查看订单详情
        </a-button>
        <a-dropdown v-if="orderId">
          <template #overlay>
            <a-menu>
              <a-menu-item key="print" @click="handlePrintReceipt">
                打印收据
              </a-menu-item>
              <a-menu-item key="download" @click="handleDownloadReceipt">
                下载收据
              </a-menu-item>
            </a-menu>
          </template>
          <a-button class="ml4">
            查看收据
            <DownOutlined :style="{ fontSize: '10px' }" />
          </a-button>
        </a-dropdown>
      </div>
      <div class="detailInfo">
        <a-form>
          <a-row :gutter="[16, 8]">
            <a-col :xs="24" :sm="12" :md="12" :lg="6" :xl="6" :xxl="6">
              <a-form-item label="学员姓名">
                <span>{{ orderInfo?.studentName || '-' }}</span>
              </a-form-item>
            </a-col>
            <a-col :xs="24" :sm="12" :md="12" :lg="6" :xl="6" :xxl="6">
              <a-form-item label="手机号">
                <span>{{ orderInfo?.studentPhone || '-' }}</span>
              </a-form-item>
            </a-col>
            <a-col :xs="24" :sm="12" :md="12" :lg="6" :xl="6" :xxl="6">
              <a-form-item label="订单编号">
                <span class="text-ellipsis" :title="orderInfo?.orderNumber || '-'">{{ orderInfo?.orderNumber || '-' }}</span>
              </a-form-item>
            </a-col>
            <a-col :xs="24" :sm="12" :md="12" :lg="6" :xl="6" :xxl="6">
              <a-form-item label="对内备注">
                <span>{{ orderInfo?.remark || '无' }}</span>
              </a-form-item>
            </a-col>
          </a-row>
          <a-row :gutter="[16, 8]">
            <a-col :xs="24" :sm="12" :md="12" :lg="6" :xl="6" :xxl="6">
              <a-form-item label="对外备注">
                <span>{{ orderInfo?.externalRemark || '无' }}</span>
              </a-form-item>
            </a-col>
            <a-col :xs="24" :sm="12" :md="12" :lg="18" :xl="18" :xxl="18">
              <a-form-item label="办理时间">
                <span>{{ formatDateOnly(orderInfo?.dealDate) }}</span>
              </a-form-item>
            </a-col>
          </a-row>
          <a-row
            v-for="(item, index) in paymentMethodList"
            :key="index"
            :gutter="[16, 8]"
            class="bg-#f6f7f8 pt-10px"
          >
            <a-col :xs="24" :sm="12" :md="12" :lg="6" :xl="6" :xxl="6">
              <a-form-item :label="`收款方式${index + 1}`">
                <span class="text-ellipsis" :title="`${item.payTitle}(¥${item.payAmount.toFixed(2)})`">{{ item.payTitle }}(¥{{ item.payAmount.toFixed(2) }})</span>
              </a-form-item>
            </a-col>
            <a-col :xs="24" :sm="12" :md="12" :lg="6" :xl="6" :xxl="6">
              <a-form-item label="收款账户">
                <span>{{ item.accountName || '默认账户' }}</span>
              </a-form-item>
            </a-col>
            <a-col :xs="24" :sm="12" :md="12" :lg="6" :xl="6" :xxl="6">
              <a-form-item label="支付单号">
                <span>{{ item.paymentNo || '-' }}</span>
              </a-form-item>
            </a-col>
            <a-col :xs="24" :sm="12" :md="12" :lg="6" :xl="6" :xxl="6">
              <a-form-item label="支付日期">
                <span>{{ formatDateOnly(item.payTime) }}</span>
              </a-form-item>
            </a-col>
          </a-row>
        </a-form>
      </div>
      <order-detail-drawer v-model:open="openOrderDetailDrawer" :order-id="orderId" />
    </div>
  </div>
</template>

<style lang="less" scoped>
.step {
  box-shadow: 0 0 8px 0 rgba(94, 188, 255, 0.08);
}

.current2 {
  .unfollowContainer {
    background: var(--pro-ant-color-primary-bg-hover);
    padding: 15px;
    margin: 0 0 20px 0;
    border-radius: 4px;

    .unfollowButtonContainer {
      margin-left: 10px;
      color: var(--pro-ant-color-primary);
      cursor: pointer;
    }

    .ExclamationCircleFilled {
      color: var(--pro-ant-color-primary);
      margin-right: 4px;
      font-size: 16px;
    }
  }

  .patStatus {
    text-align: center;

    .payPrice {
      margin-top: 12px;
      margin-bottom: 8px;
      line-height: 38px;
      font-family: "DIN alternate", sans-serif;
      font-size: 32px;
      font-weight: 700;
      color: #222;
    }

    .payTip {
      font-family: PingFangSC, PingFang SC;
      font-weight: 500;
      font-size: 16px;
      color: #f90;
      line-height: 22px;
      font-style: normal;
    }

    .payment-summary {
      font-size: 14px;
      color: #666;
      line-height: 20px;
    }
  }

  .payButton {
    text-align: center;
    border-bottom: 1px solid #eee;
    padding-bottom: 40px;
  }

  .detailInfo {
    margin-top: 24px;

    span {
      color: #888;
    }

    .text-ellipsis {
      display: inline-block;
      max-width: 100%;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
      vertical-align: top;
    }

    :deep(.ant-form-item) {
      margin-bottom: 12px !important;
    }

    .rounded-2 {
      :deep(.ant-form-item) {
        margin-bottom: 2px !important;
      }
    }
  }
}
</style>
