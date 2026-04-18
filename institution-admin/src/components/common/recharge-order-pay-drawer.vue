<script setup>
import {
  CloseOutlined,
  PlusOutlined,
} from '@ant-design/icons-vue'
import { computed, reactive, ref, watch } from 'vue'
import { message } from 'ant-design-vue'
import dayjs from 'dayjs'
import {
  getRechargeAccountByStudentApi,
  getRechargeAccountOrderDetailApi,
  getStudentDetailApi,
  payOrderBySchoolPalApi,
} from '@/api/finance-center/recharge-account'

const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
  saleOrderId: {
    type: [String, Number],
    default: '',
  },
})

const emit = defineEmits(['update:open', 'submitted'])

const drawerOpen = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

const loading = ref(false)
const submitting = ref(false)
const currentRechargeOrderDetail = ref(null)
const selectedStudent = ref(null)
const payType = ref('1')
const pay = ref(1)
const billRemarks = ref(undefined)
const fileList = ref([])
const payFormModel = reactive({
  account: 1,
  payDate: dayjs(),
})
const accountList = ref([{ value: 1, label: '默认账户' }])

const checkOptions = reactive([
  { id: 1, label: '微信', img: 'https://pcsys.admin.ybc365.com/e068b5e2-e27e-4228-8437-fae315326ced.png' },
  { id: 2, label: '支付宝', img: 'https://pcsys.admin.ybc365.com/3f396285-ac3a-43fe-94a8-adfef395d47d.png' },
  { id: 3, label: '银行转帐', img: 'https://pcsys.admin.ybc365.com/b05c2b92-6c09-46a0-8123-42446727d480.png' },
  { id: 4, label: 'POS机', img: 'https://pcsys.admin.ybc365.com/2a6137fb-3e35-470f-b368-2fe71e5439f1.png' },
  { id: 5, label: '现金', img: 'https://pcsys.admin.ybc365.com/bab59869-17ea-42c9-9354-d88a59ecf18a.png' },
  { id: 6, label: '其他', img: 'https://pcsys.admin.ybc365.com/6d6faf00-aaea-4f6c-b3ab-e581a9bcf7f6.png' },
])

const previewVisible = ref(false)
const previewImage = ref('')
const previewTitle = ref('')
const accountFormRefs = ref()

const drawerWidth = computed(() => (window.innerWidth <= 768 ? '100%' : '800px'))
const weekText = computed(() => {
  const value = dayjs(payFormModel.payDate)
  if (!value.isValid()) {
    return '-'
  }
  const names = ['周日', '周一', '周二', '周三', '周四', '周五', '周六']
  return names[value.day()]
})

function resetState() {
  currentRechargeOrderDetail.value = null
  selectedStudent.value = null
  payType.value = '1'
  pay.value = 1
  billRemarks.value = undefined
  fileList.value = []
  payFormModel.account = 1
  payFormModel.payDate = dayjs()
}

function handleClose() {
  resetState()
  drawerOpen.value = false
}

function disabledDate(current) {
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

async function handlePreview(file) {
  if (!file.url && !file.preview) {
    file.preview = await getBase64(file.originFileObj)
  }
  previewImage.value = file.url || file.preview
  previewVisible.value = true
  previewTitle.value = file.name || file.url?.substring(file.url.lastIndexOf('/') + 1) || ''
}

function handleCancelImg() {
  previewVisible.value = false
  previewTitle.value = ''
}

async function fetchRechargeOrderDetail() {
  const saleOrderId = String(props.saleOrderId || '').trim()
  if (!saleOrderId) {
    currentRechargeOrderDetail.value = null
    selectedStudent.value = null
    return
  }
  loading.value = true
  try {
    const { result: orderDetail } = await getRechargeAccountOrderDetailApi({ saleOrderId })
    currentRechargeOrderDetail.value = orderDetail || null
    const studentId = orderDetail?.studentId
    if (studentId) {
      const [{ result: studentDetail }, { result: rechargeAccount }] = await Promise.all([
        getStudentDetailApi({ studentId }),
        getRechargeAccountByStudentApi({ studentId }),
      ])
      selectedStudent.value = {
        ...studentDetail,
        rechargeAccountId: rechargeAccount?.id || '',
        rechargeAccountName: rechargeAccount?.accountName || rechargeAccount?.rechargeAccountName || '',
      }
    }
  }
  catch (error) {
    console.error('加载储值订单失败:', error)
    message.error(error?.message || '加载储值订单失败')
    handleClose()
  }
  finally {
    loading.value = false
  }
}

async function handleSubmit() {
  if (submitting.value) {
    return
  }
  if (!currentRechargeOrderDetail.value?.bill?.id) {
    message.error('未找到支付账单')
    return
  }
  submitting.value = true
  try {
    await accountFormRefs.value?.validate()
  }
  catch {
    submitting.value = false
    return
  }
  try {
    await payOrderBySchoolPalApi({
      billId: String(currentRechargeOrderDetail.value.bill.id),
      amount: Number(currentRechargeOrderDetail.value.amount || 0),
      remark: billRemarks.value || '',
      payMethod: pay.value,
      amountId: Number(payFormModel.account || 0),
      payTime: payFormModel.payDate ? dayjs(payFormModel.payDate).format('YYYY-MM-DD') : undefined,
    })
    message.success('储值订单付款成功')
    emit('submitted')
    handleClose()
  }
  catch (error) {
    console.error('储值订单付款失败:', error)
    message.error(error?.message || '储值订单付款失败')
  }
  finally {
    submitting.value = false
  }
}

watch(
  () => [props.open, props.saleOrderId],
  async ([open]) => {
    if (!open) {
      resetState()
      return
    }
    await fetchRechargeOrderDetail()
  },
  { immediate: true },
)
</script>

<template>
  <div>
    <a-drawer
      v-model:open="drawerOpen"
      :mask-closable="false"
      :keyboard="false"
      :body-style="{ padding: '0', background: '#fff' }"
      :closable="false"
      :width="drawerWidth"
      placement="right"
      @close="handleClose"
    >
      <template #title>
        <div class="custom-header flex justify-between h-4 flex-items-center">
          <div class="text-5">
            订单确认
          </div>
          <a-button type="text" class="close-btn" @click="handleClose">
            <template #icon>
              <CloseOutlined class="text-5 close-icon" />
            </template>
          </a-button>
        </div>
      </template>

      <div v-if="loading" class="py-20 flex justify-center">
        <a-spin size="large" tip="加载储值订单中..." />
      </div>

      <div v-else-if="currentRechargeOrderDetail">
        <span class="bg-#f6f7f8 flex h-2" />
        <div class="bg-white mt-2 justify-start p-8 pb2">
          <div class="Yprice mb1">
            <span>*</span>订单确认：
          </div>
          <div class="Yprice-num ml4">
            ¥ {{ Number(currentRechargeOrderDetail.amount || 0).toFixed(2) }}
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
              <div class="pay mt-3">
                <a-radio-group v-model:value="pay" class="custom-radio w-full">
                  <a-space :size="16" class="flex-wrap">
                    <label
                      v-for="item in checkOptions"
                      :key="item.id"
                      class="pay-box"
                      :class="{ active: pay === item.id }"
                    >
                      <span>
                        <img :src="item.img" alt="">
                        {{ item.label }}
                      </span>
                      <a-radio :value="item.id" />
                    </label>
                  </a-space>
                </a-radio-group>
              </div>
            </div>
          </div>
        </div>

        <div class="bg-white p-8 pt4 pb0 ">
          <a-form ref="accountFormRefs" class="flex flex-col" layout="vertical" :model="payFormModel">
            <div class="flex w-100%">
              <a-form-item label="收款账户" class="flex-1" name="account" :rules="[
                {
                  required: true,
                  message: '请选择收款账户',
                  trigger: 'change',
                },
              ]">
                <a-select
                  v-model:value="payFormModel.account"
                  :allow-clear="false"
                  placeholder="请选择收款账户"
                  :options="accountList"
                  style="flex:1"
                />
              </a-form-item>
              <a-form-item label="支付日期" class="flex-1 pl10" name="payDate" :rules="[
                {
                  required: true,
                  message: '请选择支付日期',
                  trigger: 'change',
                },
              ]">
                <div class="flex flex-items-center week-wrap">
                  <a-date-picker
                    v-model:value="payFormModel.payDate"
                    style="flex:1;"
                    :allow-clear="false"
                    :disabled-date="disabledDate"
                    :default-value="dayjs()"
                    format="YYYY-MM-DD"
                    class="w-35"
                    placeholder="请选择日期"
                  />
                  <div class="week">
                    {{ weekText }}
                  </div>
                </div>
              </a-form-item>
            </div>
            <a-form-item label="账单备注（选填）">
              <a-textarea
                v-model:value="billRemarks"
                placeholder="请输入内容，最多100字"
                :auto-size="{ minRows: 2, maxRows: 5 }"
              />
            </a-form-item>
          </a-form>
        </div>

        <div class="upload bg-white p-8 pt0 mt--4">
          <a-upload
            v-model:file-list="fileList"
            action="https://www.mocky.io/v2/5cc8019d300000980a055e76"
            list-type="picture-card"
            @preview="handlePreview"
          >
            <div v-if="fileList.length < 3">
              <PlusOutlined class="text-6" />
            </div>
          </a-upload>
          <span class="text-#888">最多上传 3 张图片，支持 BMP / JPG / JPEG / PNG，单张图片不超过 4 MB</span>
        </div>

      </div>

      <template #footer>
        <div v-if="currentRechargeOrderDetail && !loading" class="h-15 flex flex-center justify-end pr-8">
          <a-space :size="14">
            <a-button type="primary" class="h-12 w-35 text-5" :loading="submitting" :disabled="submitting" @click="handleSubmit">
              确定
            </a-button>
          </a-space>
        </div>
      </template>
    </a-drawer>

    <a-modal :open="previewVisible" :title="previewTitle" :footer="null" @cancel="handleCancelImg">
      <img alt="example" style="width: 100%" :src="previewImage">
    </a-modal>
  </div>
</template>

<style lang="less" scoped>
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

.Yprice {
  font-size: 14px;

  span {
    color: red;
  }
}

.Yprice-num {
  width: 560px;
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
  margin-bottom: 20px;

  .pay-box {
    border: 1px solid #eee;
    padding: 12px 16px;
    display: flex;
    align-items: center;
    border-radius: 6px;
    margin-right: 6px;
    user-select: none;
    cursor: pointer;
    margin-bottom: 16px;

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
</style>
