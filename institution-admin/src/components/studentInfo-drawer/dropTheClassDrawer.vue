<script setup>
import { CloseOutlined, PlusOutlined, QuestionCircleFilled } from '@ant-design/icons-vue'
import { h } from 'vue'
import dayjs from 'dayjs'

const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },

})
const emit = defineEmits(['update:open'])
// 处理双向绑定
const openDrawer = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})
// 获取当前日期（格式：YYYY-MM-DD）
const getCurrentDate = () => dayjs().format('YYYY-MM-DD')
// 确定退课提示
const openModal = ref(false)
// defineEmits(['update:open']);
const current = ref(0)
const formRef = ref(null)
const items = computed(() => [
  {
    title: '填写退课单',
    icon: h('img', {
      src: 'https://prod-tbu-next-erp-cdn.schoolpal.cn/next-pc-static/static/12181/static/transfer-pre.b79033da.svg',
      style: { width: '32px', height: '32px', marginTop: '-3px', marginRight: '-8px' },
    }),
  },
  {
    title: '确认实退金额',
    icon: h('img', {
      src: current.value === 1
        ? 'https://prod-tbu-next-erp-cdn.schoolpal.cn/next-pc-static/static/12181/static/transfer-next.4b9c0674.svg'
        : 'https://prod-tbu-next-erp-cdn.schoolpal.cn/next-pc-static/static/12181/static/transfer-next-normal.11eb34d7.svg',
      style: { width: '32px', height: '32px', marginTop: '-3px', marginRight: '-6px' },
    }),
  },
])
// 初始化表单状态的函数
function initFormState() {
  return {
  // 剩余课时
    surplusClass: 5,
    // 有效期至
    effectiveDate: '2025-06-01',
    // 退课数量
    dropTheClassNumber: '',
    // 应退金额
    price: '200.99',
    // 退款方式
    payType: 1,
    // 实退金额
    dropPayPrice: '200.99',
    // 退款账户
    dropPayAccount: '1',
    // 经办日期
    date: getCurrentDate(), // 使用dayjs更简洁,
    // 订单标签
    orderLabel: [],
    // 订单销售员
    salesperson: undefined,
    // 对内备注
    remarks1: '',
    // 对外备注
    remarks2: '',
    // 账单备注
    billRemarks: '',
    fileList: [
      {
        uid: '-1',
        name: 'image.png',
        status: 'done',
        url: 'https://zos.alipayobjects.com/rmsportal/jkjgkEfvpUPVyRjUImniVslZfWPnJuuZ.png',
      },
      {
        uid: '-2',
        name: 'image.png',
        status: 'done',
        url: 'https://zos.alipayobjects.com/rmsportal/jkjgkEfvpUPVyRjUImniVslZfWPnJuuZ.png',
      },
      {
        uid: '-3',
        name: 'image.png',
        status: 'done',
        url: 'https://zos.alipayobjects.com/rmsportal/jkjgkEfvpUPVyRjUImniVslZfWPnJuuZ.png',
      },
    ],
  }
}

const formState = reactive(initFormState())
// 订单标签
const orderLabelOptions = ref([
  { id: '1', name: '内荐' },
  { id: '2', name: '转介绍' },
  { id: '3', name: '双11订单' },
  { id: '4', name: '会员日订单' },
])
// 订单销售员
const salespersonOptions = ref([
  { id: '1', name: '陈瑞生', phone: '17601241636' },
  { id: '2', name: '刘明', phone: '18876552232' },
  { id: '3', name: '张望名', phone: '17601241636' },
  { id: '4', name: '李元芳', phone: '17601241636' },
])
function orderLabelFilterOption(input, option) {
  const keyword = input.toLowerCase()
  const nameMatch = option.name.toLowerCase().includes(keyword)
  return nameMatch
}
// 禁止选择今天之后的日期
function disabledDate(current) {
  // 当天结束时间（23:59:59）之后的时间不可选
  return current > dayjs().endOf('day')
}
// 处理日期更新
function handleDateChange(dateObj) {
  if (!dateObj) {
    // 仅当值被清空时恢复当前日期
    formState.date = getCurrentDate()
  }
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
async function handlePreview(file) {
  if (!file.url && !file.preview) {
    file.preview = await getBase64(file.originFileObj)
  }
  previewImage.value = file.url || file.preview
  previewVisible.value = true
  previewTitle.value = file.name || file.url.substring(file.url.lastIndexOf('/') + 1)
}
function handleCancelPreview() {
  previewVisible.value = false
  previewTitle.value = ''
}
// 全部退还
function handleAllReturn() {
  formState.dropTheClassNumber = formState.surplusClass
}
function handleNext() {
  formRef.value.validate().then(() => {
    current.value++
  })
}
function handleConfirm() {
  openModal.value = true
}
function handleBack() {
  current.value--
}
function handleCancel() {
  current.value = 0
  openDrawer.value = false
  // 重置整个formState为初始状态
  Object.assign(formState, initFormState())
  formRef.value.resetFields()
}
const checkOptions = reactive([
  {
    id: 1,
    label: '微信',
    img: 'https://pcsys.admin.ybc365.com/e068b5e2-e27e-4228-8437-fae315326ced.png',
  },
  {
    id: 2,
    label: '支付宝',
    img: 'https://pcsys.admin.ybc365.com/3f396285-ac3a-43fe-94a8-adfef395d47d.png',
  },
  {
    id: 3,
    label: '银行转帐',
    img: 'https://pcsys.admin.ybc365.com/b05c2b92-6c09-46a0-8123-42446727d480.png',
  },
  {
    id: 4,
    label: 'POS机',
    img: 'https://pcsys.admin.ybc365.com/2a6137fb-3e35-470f-b368-2fe71e5439f1.png',
  },
  {
    id: 5,
    label: '现金',
    img: 'https://pcsys.admin.ybc365.com/bab59869-17ea-42c9-9354-d88a59ecf18a.png',
  },
  {
    id: 6,
    label: '其他',
    img: 'https://pcsys.admin.ybc365.com/6d6faf00-aaea-4f6c-b3ab-e581a9bcf7f6.png',
  },
  {
    id: 7,
    label: '储值账户',
    img: 'https://pcsys.admin.ybc365.com/bca6bad7-3180-4b01-b766-4b05013347f7.png',
  },

])
const readonly = ref(true)
function handleModify() {
  readonly.value = false
}
</script>

<template>
  <div>
    <a-drawer
      v-model:open="openDrawer" :push="{ distance: 80 }" :body-style="{ padding: '0', background: '#f7f7fd' }"
      :closable="false" width="1165px" placement="right" :mask-closable="false" :keyboard="false"
    >
      <!-- 自定义头部 -->
      <template #title>
        <div class="custom-header flex justify-between h-4 flex-items-center">
          <div class="text-5">
            退课
          </div>
          <a-button type="text" class="close-btn" @click="handleCancel">
            <template #icon>
              <CloseOutlined class="text-5 close-icon" />
            </template>
          </a-button>
        </div>
      </template>
      <a-alert
        message="退课后，将会扣除退课所填写的课时并生成退费订单，请谨慎操作！" show-icon type="warning"
        class="text-#f90 border-none bg-#fff5e6"
      />
      <div class="contenter">
        <div class="steps mt-24px mx-24px bg-#fff rounded-16px flex flex-center px-28% py-24px">
          <a-steps :current="current" :items="items" />
        </div>
        <a-form ref="formRef" :model="formState">
          <div v-if="current === 0" class="mt-24px mx-24px p-24px bg-#fff rounded-16px flex flex-col   py-24px">
            <h1 class="text-20px">
              听觉训练课
            </h1>
            <a-space class="flex-1 flex flex-items-center">
              <span class="bg-#e6f0ff text-#06f text-3 px3 py2px rounded-10 ">1v1授课</span>
              <span class="bg-#e6f0ff text-#06f text-3 px3 py2px rounded-10">课时</span>
            </a-space>
            <div class="flex flex-items-center my-24px">
              <span>剩余课时：</span>
              <span class="text-#888">5课时</span>
            </div>
            <div class="flex flex-items-center">
              <span>有效期至：</span>
              <span class="text-#888">2025-06-01</span>
            </div>
            <!-- 分割线 -->
            <a-divider />
            <!-- 退课数量 -->
            <a-form-item label="退课数量" name="dropTheClassNumber" :rules="[{ required: true, message: '请输入退课数量' }]">
              <div class="flex flex-items-center">
                <a-input-number
                  v-model:value="formState.dropTheClassNumber" placeholder="请输入退课数量" :min="0" :max="10000"
                  :precision="2" class="w-160px"
                />
                <a-button type="link" class="text-#06f" @click="handleAllReturn">
                  全部退还
                </a-button>
              </div>
            </a-form-item>
          </div>
          <div
            v-if="current === 1"
            class="mt-24px mx-24px p-24px pb-0 mb-24px bg-#fff rounded-16px flex flex-col   py-24px"
          >
            <!-- 应退金额 -->
            <div class="flex flex-col">
              <span class="text-#666 mb-2px">应退金额：</span>
              <span class="text-#000 text-48px custom-num-font-family">¥ {{ formState.price }} <span
                class="text-#888 text-14px "
              >退还
                5.00
                课时，退赠
                0.00 课时</span> </span>
            </div>
            <!-- 退款方式 -->
            <a-form-item name="payType">
              <div class="payList mt-4 text-#666">
                <div>
                  <span class=" mr-2px">*</span>退款方式
                  <span class="payList-tip ml-2">请选择</span>
                </div>
                <div class="pay">
                  <a-radio-group v-model:value="formState.payType" class="custom-radio">
                    <a-space :size="16" class="flex-wrap">
                      <label
                        v-for="(item, index) in checkOptions" :key="index"
                        class="pay-box" :class="{ active: formState.payType === item.id }"
                      >
                        <span> <img :src="item.img" alt=""> {{ item.label }}</span>
                        <a-radio :value="item.id" />
                      </label>
                    </a-space>
                  </a-radio-group>
                </div>
              </div>
            </a-form-item>
            <a-form-item name="dropPayPrice">
              <!-- 实退金额 -->
              <div class="flex flex-col mt-20px mb-40px">
                <span class="text-#666 mb-2px"><span class="text-red mr-2px">*</span>实退金额：</span>
                <div class="text-#000 text-48px custom-num-font-family">
                  <div
                    class="payPrice h-20 border border-b-#eee border-solid border-x-none border-t-none w-100%"
                    :class="{ 'animate-border': !formState.dropPayPrice }"
                  >
                    <a-input-number
                      v-model:value="formState.dropPayPrice" :precision="2"
                      :bordered="false" :controls="false" :readonly="readonly" class="h-100% w-100% text-12"
                      :min="1" :max="100000" placeholder="输入实退金额" @blur="formState.dropPayPrice ? readonly = true : readonly = false"
                    >
                      <template #addonBefore>
                        <span class="text-12">¥</span>
                      </template>
                      <template #addonAfter>
                        <a-button v-if="readonly" @click="handleModify">
                          修改
                        </a-button>
                      </template>
                    </a-input-number>
                    <span v-if="!formState.dropPayPrice" class="text-3.5 text-#f33 relative top--27px">请输入实退金额</span>
                    <span
                      v-if="formState.dropPayPrice && formState.dropPayPrice < formState.price"
                      class="text-3.5 text-#888 relative top--27px"
                    >应退金额：¥ {{ formState.dropPayPrice }}，手续费 ¥
                      {{ (formState.price
                        - formState.dropPayPrice).toFixed(2) }}</span>
                    <span
                      v-if="formState.dropPayPrice && formState.dropPayPrice == formState.price"
                      class="text-3.5 text-#888 relative top--27px"
                    >应退金额：¥ {{ formState.dropPayPrice }}</span>
                    <span
                      v-if="formState.dropPayPrice && formState.dropPayPrice > formState.price"
                      class="text-3.5 text-#888 relative top--27px"
                    >应退金额：¥ {{ formState.dropPayPrice }}，亏损费 ¥
                      {{ (formState.dropPayPrice - formState.price).toFixed(2) }}</span>
                  </div>
                </div>
              </div>
            </a-form-item>
            <!-- 退款账户 -->
            <a-form-item name="dropPayAccount" :rules="[{ required: true, message: '请选择退款账户' }]">
              <div class="text-#666 flex flex-items-center mb-6px">
                <span class="text-red mr-2px">*</span>退款账户：
              </div>
              <a-select v-model:value="formState.dropPayAccount" placeholder="请选择退款账户" style="width: 300px;">
                <a-select-option value="1">
                  默认账户
                </a-select-option>
              </a-select>
            </a-form-item>
            <!-- 经办日期 -->
            <a-form-item class="custom-label mt--2">
              <div class="text-#666 flex flex-items-center mb-6px">
                经办日期：
              </div>
              <a-date-picker
                v-model:value="formState.date" class="w-300px" :disabled-date="disabledDate"
                value-format="YYYY-MM-DD" format="YYYY-MM-DD" @change="handleDateChange"
              />
            </a-form-item>
            <!-- 订单标签 -->
            <a-form-item>
              <div class="text-#666 flex flex-items-center mb-6px">
                订单标签：
              </div>
              <a-select
                v-model:value="formState.orderLabel" mode="multiple" placeholder="请选择订单标签" show-search
                class="multiple-select" style="width: 100%" :options="orderLabelOptions"
                :filter-option="orderLabelFilterOption" :field-names="{ label: 'name', value: 'id' }"
              />
            </a-form-item>
            <!-- 订单销售员 -->
            <a-form-item>
              <div class="text-#666 flex flex-items-center mb-6px">
                订单销售员：
              </div>
              <a-select
                v-model:value="formState.salesperson" placeholder="请选择销售员" show-search style="width: 320px"
                :options="salespersonOptions" :field-names="{ label: 'name', value: 'id' }"
              >
                <template #option="{ name, phone }">
                  <div class="flex justify-between flex-items-center">
                    <span>{{ name }}</span>
                    <span class="text-#999 text-3">{{ phone }}</span>
                  </div>
                </template>
              </a-select>
            </a-form-item>
            <a-form-item>
              <div class="text-#666 flex flex-items-center mb-6px">
                对内备注：
              </div>
              <a-input
                v-model:value="formState.remarks1" placeholder="请输入对内备注，此备注仅内部员工可见"
                style="width: 100%"
              />
            </a-form-item>
            <a-form-item>
              <div class="text-#666 flex flex-items-center mb-6px">
                对外备注：
              </div>
              <a-input v-model:value="formState.remarks2" placeholder="请输入对内备注，此备注打印时将显示" style="width: 100%" />
            </a-form-item>
            <a-form-item>
              <div class="text-#666 flex flex-items-center mb-6px">
                账单备注：
              </div>
              <a-textarea
                v-model:value="formState.billRemarks" placeholder="请输入内容，最多100字"
                :auto-size="{ minRows: 2, maxRows: 5 }"
              />
            </a-form-item>
            <a-form-item class="w-80%">
              <a-form-item-rest>
                <div class="mt--10px">
                  <a-upload
                    v-model:file-list="formState.fileList"
                    action="https://www.mocky.io/v2/5cc8019d300000980a055e76" list-type="picture-card"
                    accept=".jpg,.jpeg,.png" @preview="handlePreview"
                  >
                    <div v-if="formState.fileList.length < 3">
                      <PlusOutlined class="text-20px" />
                    </div>
                  </a-upload>
                  <span class="text-#888 text-12px">最多上传3张，支持JPG、JPEG、PNG，单张图片不超过 4 MB</span>
                  <a-modal :open="previewVisible" :title="previewTitle" :footer="null" @cancel="handleCancelPreview">
                    <img alt="example" style="width: 100%" :src="previewImage">
                  </a-modal>
                </div>
              </a-form-item-rest>
            </a-form-item>
          </div>
        </a-form>
      </div>
      <template #footer>
        <div v-if="current === 0" class="px-24px py-16px flex  justify-end flex-items-center">
          <div class="l flex flex-col items-end">
            <span class="text-16px text-#333 font-bold mb-8px">总计学费：¥ {{ formState.price }}</span>
            <span class="text-14px text-#888">退还 5.00 课时，退赠 0 课时（赠送不计入总计）</span>
          </div>
          <div class="flex flex-items-center ml-24px">
            <a-button type="primary" class="text-20px h-48px w-140px" @click="handleNext">
              下一步
            </a-button>
          </div>
        </div>
        <div v-if="current === 1" class="px-24px py-16px flex  justify-between flex-items-center">
          <div>
            <a-button type="primary" class="text-20px h-48px w-140px" ghost @click="handleBack">
              返回上一步
            </a-button>
          </div>
          <div class="flex flex-items-center">
            <div class="l flex flex-col items-end">
              <span class="text-16px text-#333 font-bold mb-8px">实退金额：¥ {{ formState.price }}</span>
              <span class="text-14px text-#888">应退金额：¥ {{ formState.price }}</span>
            </div>
            <div class="flex flex-items-center ml-24px">
              <a-button type="primary" class="text-20px h-48px w-140px" @click="handleConfirm">
                确定
              </a-button>
            </div>
          </div>
        </div>
      </template>
    </a-drawer>
    <!-- 确定退课提示 -->
    <a-modal
      v-model:open="openModal" centered :mask-closable="false" :keyboard="false" :footer="false" :width="416"
      :closable="false"
    >
      <div class="text-16px font-bold flex flex-items-center">
        <QuestionCircleFilled class="text-#f33 text-22px mr-12px" />确定退课？
      </div>
      <div class="pl-35px text-#666 mt-12px">
        退课后，将会扣除退课所填写的课时并生成退费订单，请谨慎操作！
      </div>
      <div class="flex flex-items-center justify-end mt-24px">
        <a-button @click="openModal = false">
          再想想
        </a-button>
        <a-button danger ghost class="ml-12px" @click="handleConfirm">
          确定退课
        </a-button>
      </div>
    </a-modal>
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
    margin-right: 6px;
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
  animation: borderExpand 0.4s cubic-bezier(0.68, -0.55, 0.27, 1.2) forwards;
}

.multiple-select {
  :deep(.ant-select-selection-item) {
    background-color: #e6f0ff;
    border: 1px solid #99c2ff;
  }
}
</style>
