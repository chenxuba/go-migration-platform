<script setup>
import { h, reactive, ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import {
  CreditCardOutlined,
  FileDoneOutlined,
  SolutionOutlined,
} from '@ant-design/icons-vue'
import { getOrderDetailApi } from '@/api/finance-center/order-manage'
import messageService from '@/utils/messageService'

// 引入步骤组件
import Step01 from '@/components/edu-center/registr-renewal/step01/index.vue'
import Step02 from '@/components/edu-center/registr-renewal/step02/index.vue'
import Step03 from '@/components/edu-center/registr-renewal/step03/index.vue'

const route = useRoute()
const current = ref(0)
const bottom = ref(0)
const handleOver = ref(false)
const orderData = ref({}) // 存储从step01传递过来的订单数据
const paymentData = ref({}) // 存储从step02传递过来的支付数据
const loading = ref(false)

const items = [
  {
    title: '生成订单',
    icon: h(SolutionOutlined),
  },
  {
    title: '订单支付',
    icon: h(CreditCardOutlined),
  },
  {
    title: '订单完成',
    icon: h(FileDoneOutlined),
  },
]

const formState = reactive({
  studentId: undefined,
  selectedStudentInfo: undefined,
  orderDetail:{
    quoteDetailList:[]
  }
})

// 处理步骤1的事件
function handleFormStateUpdate(newFormState) {
  Object.assign(formState, newFormState)
}

function handleHandleOverUpdate(value, formData) {
  // console.log('handleHandleOverUpdate', formData)
  handleOver.value = value
}

function handleSubmitOrder(data) {
  orderData.value = data // 存储订单数据
  current.value++
}

// 处理步骤2的事件
function handleSubmitPayment(data) {
  paymentData.value = data // 存储支付数据
  current.value++
}

// 处理步骤3的事件
function handleGoBack() {
  current.value = 0
  handleOver.value = false
  // 重置状态
  formState.studentId = undefined
  formState.selectedStudentInfo = undefined
  orderData.value = {} // 清空订单数据
  paymentData.value = {} // 清空支付数据
}

function handleCreateNewOrder() {
  const selectedStudentInfo = orderData.value?.studentInfo || undefined
  // 重置状态，回到第一步
  current.value = 0
  handleOver.value = false
  formState.studentId = selectedStudentInfo?.id || selectedStudentInfo?.studentId || undefined
  formState.selectedStudentInfo = selectedStudentInfo
  orderData.value = {} // 清空订单数据
  paymentData.value = {} // 清空支付数据
}

function handleViewOrderDetail() {
  // 查看订单详情逻辑
  console.log('查看订单详情')
}

// 从路由参数加载订单数据
async function loadOrderFromRoute() {
  const orderId = route.params.id
  const step = route.query.step
  
  if (orderId) {
    try {
      loading.value = true
      const { result } = await getOrderDetailApi({ orderId })
      
      if (result) {
        const isRepayment = step === '1' && result.isAmountOwed
        const receivableAmount = isRepayment
          ? (result.arrearAmount ?? 0)
          : (result.totalAmount ?? result.amount ?? 0)

        // 将订单详情数据转换为orderData格式
        orderData.value = {
          orderId: result.orderId,
          orderNumber: result.orderNumber,
          amountInfo: {
            totalAmount: receivableAmount,
            paidAmount: result.paidAmount || 0,
            arrearAmount: result.arrearAmount || 0,
          },
          studentInfo: {
            studentId: result.studentId,
            studentName: result.studentName,
          },
        }
        
        // 如果指定了步骤，直接跳转到对应步骤
        if (step === '1') {
          current.value = 1 // 跳转到付款步骤
        }
      }
    }
    catch (error) {
      console.error('加载订单失败:', error)
      messageService.error('加载订单失败')
    }
    finally {
      loading.value = false
    }
  }
}

// 页面加载时检查路由参数
onMounted(() => {
  loadOrderFromRoute()
})
</script>

<template>
  <div v-if="loading" class="main flex items-center justify-center" style="min-height: 400px;">
    <a-spin size="large" tip="加载订单数据中..." />
  </div>
  <div v-else class="main">
    <div class="step header h-20 bg-white rounded-4 flex flex-center">
      <a-steps style="padding: 0 20%;" :current="current" :items="items" />
    </div>
    <!-- 步骤1：生成订单 -->
    <Step01
      v-if="current === 0" :form-state="formState" :handle-over="handleOver"
      @update:form-state="handleFormStateUpdate" @update:handle-over="handleHandleOverUpdate"
      @submit-order="handleSubmitOrder"
    />

    <!-- 步骤2：订单支付 -->
    <Step02 v-if="current === 1" :order-data="orderData" @submit-payment="handleSubmitPayment" />

    <!-- 步骤3：订单完成 -->
    <Step03
      v-if="current === 2" :payment-data="paymentData" @go-back="handleGoBack"
      @create-new-order="handleCreateNewOrder" @view-order-detail="handleViewOrderDetail"
    />
  </div>
</template>

<style lang="less" scoped>
.step {
  box-shadow: 0 0 8px 0 rgba(94, 188, 255, 0.08);
}
</style>
