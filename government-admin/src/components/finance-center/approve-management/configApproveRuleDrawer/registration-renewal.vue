<script setup>
import { ref, watch } from 'vue'
import { QuestionCircleOutlined } from '@ant-design/icons-vue'

const props = defineProps({
  checked: {
    type: Boolean,
    default: false,
  },
  submitAttempted: {
    type: Number,
    default: 0,
  },
  templateData: {
    type: Object,
    default: () => ({}),
  },
})

const emit = defineEmits(['update:checked', 'update:templateData'])

const checkedValue = ref(props.checked)

watch(() => props.checked, (val) => {
  checkedValue.value = val
}, { immediate: true })

watch(checkedValue, (val) => {
  emit('update:checked', val)
})

const approvalCriteria = ref('1') // 审批条件

const giveClassTimeChecked = ref(false) // 赠送课时开关
const givePriceChecked = ref(false) // 赠送金额开关
const giveDaysChecked = ref(false) // 赠送天数开关

const giveClassTime = ref(null) // 赠送课时
const givePrice = ref(null) // 赠送金额
const giveDays = ref(null) // 赠送天数

const discountRateChecked = ref(false) // 整单优惠折扣开关
const discountAmountChecked = ref(false) // 整单优惠金额开关

const discountRate = ref(null) // 整单优惠折扣
const discountAmount = ref(null) // 整单优惠金额

const list = ref([]) // 审批流程
const hydrating = ref(false)
const showValidation = ref(false)

function isPositiveValue(value) {
  return value !== null && value !== undefined && value !== '' && Number(value) > 0
}

function fieldInvalid(checked, value) {
  return showValidation.value && approvalCriteria.value === '2' && checked && !isPositiveValue(value)
}

function resetRuleState() {
  approvalCriteria.value = '1'
  giveClassTimeChecked.value = false
  givePriceChecked.value = false
  giveDaysChecked.value = false
  giveClassTime.value = null
  givePrice.value = null
  giveDays.value = null
  discountRateChecked.value = false
  discountAmountChecked.value = false
  discountRate.value = null
  discountAmount.value = null
}

function applyRuleDraft(draft) {
  if (!draft || typeof draft !== 'object') {
    return false
  }
  approvalCriteria.value = draft.approvalCriteria || '1'
  giveClassTimeChecked.value = !!draft.giveClassTimeChecked
  givePriceChecked.value = !!draft.givePriceChecked
  giveDaysChecked.value = !!draft.giveDaysChecked
  giveClassTime.value = draft.giveClassTime ?? null
  givePrice.value = draft.givePrice ?? null
  giveDays.value = draft.giveDays ?? null
  discountRateChecked.value = !!draft.discountRateChecked
  discountAmountChecked.value = !!draft.discountAmountChecked
  discountRate.value = draft.discountRate ?? null
  discountAmount.value = draft.discountAmount ?? null
  return true
}

function buildRuleDraft() {
  return {
    approvalCriteria: approvalCriteria.value,
    giveClassTimeChecked: giveClassTimeChecked.value,
    givePriceChecked: givePriceChecked.value,
    giveDaysChecked: giveDaysChecked.value,
    giveClassTime: giveClassTime.value,
    givePrice: givePrice.value,
    giveDays: giveDays.value,
    discountRateChecked: discountRateChecked.value,
    discountAmountChecked: discountAmountChecked.value,
    discountRate: discountRate.value,
    discountAmount: discountAmount.value,
  }
}

function parseRuleJson(ruleJson) {
  if (applyRuleDraft(props.templateData?.__ruleState)) {
    return
  }
  const savedCriteria = props.templateData?.__approvalCriteria
  resetRuleState()
  if (savedCriteria === '2') {
    approvalCriteria.value = '2'
  }
  const raw = String(ruleJson || '').trim()
  if (!raw) {
    return
  }
  try {
    const parsed = JSON.parse(raw)
    const hasAnyRule = Object.keys(parsed || {}).length > 0
    if (!hasAnyRule) {
      return
    }
    approvalCriteria.value = '2'
    if (parsed.classTimeFreeQuantity > 0) {
      giveClassTimeChecked.value = true
      giveClassTime.value = Number(parsed.classTimeFreeQuantity)
    }
    if (parsed.priceFreeQuantity > 0) {
      givePriceChecked.value = true
      givePrice.value = Number(parsed.priceFreeQuantity)
    }
    if (parsed.dateFreeQuantity > 0) {
      giveDaysChecked.value = true
      giveDays.value = Number(parsed.dateFreeQuantity)
    }
    if (parsed.discount > 0) {
      discountRateChecked.value = true
      discountRate.value = Number(parsed.discount)
    }
    if (parsed.discountPrice > 0) {
      discountAmountChecked.value = true
      discountAmount.value = Number(parsed.discountPrice)
    }
  }
  catch (error) {
    console.error('解析报名续费审批规则失败:', error)
  }
}

function buildRulePayload() {
  if (approvalCriteria.value !== '2') {
    return ''
  }
  const payload = {}
  if (giveClassTimeChecked.value && Number(giveClassTime.value) > 0) {
    payload.classTimeFreeQuantity = Number(giveClassTime.value)
  }
  if (givePriceChecked.value && Number(givePrice.value) > 0) {
    payload.priceFreeQuantity = Number(givePrice.value)
  }
  if (giveDaysChecked.value && Number(giveDays.value) > 0) {
    payload.dateFreeQuantity = Number(giveDays.value)
  }
  if (discountRateChecked.value && Number(discountRate.value) > 0) {
    payload.discount = Number(discountRate.value)
  }
  if (discountAmountChecked.value && Number(discountAmount.value) > 0) {
    payload.discountPrice = Number(discountAmount.value)
  }
  return Object.keys(payload).length > 0 ? JSON.stringify(payload) : ''
}

function emitTemplateRuleUpdate() {
  if (hydrating.value) {
    return
  }
  emit('update:templateData', {
    ...props.templateData,
    __approvalCriteria: approvalCriteria.value,
    __ruleState: buildRuleDraft(),
    ruleJson: buildRulePayload(),
  })
}

watch(() => props.templateData, (val) => {
  hydrating.value = true
  list.value = Array.isArray(val?.flowModels)
    ? val.flowModels.map(item => ({
        name: Array.isArray(item.staffNames) && item.staffNames.length > 0
          ? item.staffNames.join('、')
          : (item.staffIds || []).join('、'),
      }))
    : []
  parseRuleJson(val?.ruleJson)
  hydrating.value = false
}, { immediate: true, deep: true })

watch([
  approvalCriteria,
  giveClassTimeChecked,
  givePriceChecked,
  giveDaysChecked,
  discountRateChecked,
  discountAmountChecked,
  giveClassTime,
  givePrice,
  giveDays,
  discountRate,
  discountAmount,
], () => {
  emitTemplateRuleUpdate()
})

watch(() => props.submitAttempted, (value) => {
  if (value > 0) {
    showValidation.value = true
  }
})

const approvalProcessOpen = ref(false) // 审批流程弹窗

function handleConfigApprovalProcess() {
  approvalProcessOpen.value = true
}

function handleProcessSave(flowModels) {
  emit('update:templateData', {
    ...props.templateData,
    enable: checkedValue.value,
    __approvalCriteria: approvalCriteria.value,
    __ruleState: buildRuleDraft(),
    ruleJson: buildRulePayload(),
    flowModels,
  })
}
</script>

<template>
  <div class="mt-12px">
    <div class="flex flex-items-center">
      <div class="text-#222 font500 text-20px mr-14px w-86px text-right">
        报名续费
      </div>
      <a-switch v-model:checked="checkedValue" />
    </div>
    <!-- 审批条件 -->
    <div v-if="checkedValue" class="mt-20px w-700px">
      <!-- 审批条件 单选 -->
      <div class="flex flex-items-center">
        <div class="text-#222  text-14px mr-0px w-100px text-right">
          审批条件：
        </div>
        <a-radio-group v-model:value="approvalCriteria" class="custom-radio">
          <a-radio value="1" class="text-#666">
            不限制，订单提交/支付后即生成审批
          </a-radio>
          <a-radio value="2" class="text-#666">
            限制条件 (可多选，满足任一条件即生成审批)
          </a-radio>
        </a-radio-group>
      </div>
      <!-- 按赠送金额 -->
      <div v-if="approvalCriteria === '2'" class="mt-20px flex flex-items-start">
        <div class="text-#222  text-14px mt-8px w-100px text-right">
          按赠送金额：
        </div>
        <!-- 多选 -->
        <div class="flex  flex-col">
          <div class="flex flex-items-center mb-12px">
            <a-checkbox v-model:checked="giveClassTimeChecked" class="text-#666 ">
              <div class="flex flex-items-center" @click.stop>
                赠送课时 ＞
              </div>
            </a-checkbox>
            <div class="flex flex-col">
              <a-input-number
                v-model:value="giveClassTime" :precision="2" placeholder="请输入" :min="0.01"
                class="w-100px mr-6px input-stop-propagation"
                :status="fieldInvalid(giveClassTimeChecked, giveClassTime) ? 'error' : ''"
              />
              <span v-if="fieldInvalid(giveClassTimeChecked, giveClassTime)" class="error-text">请输入</span>
            </div>
            <span class="ml-6px">课时</span>
          </div>
          <div class="flex flex-items-center mb-12px">
            <a-checkbox v-model:checked="givePriceChecked" class="text-#666 ">
              <div class="flex flex-items-center" @click.stop>
                赠送金额 ＞
              </div>
            </a-checkbox>
            <div class="flex flex-col">
              <a-input-number
                v-model:value="givePrice" :precision="2" placeholder="请输入" :min="0.01"
                class="w-100px mr-6px input-stop-propagation"
                :status="fieldInvalid(givePriceChecked, givePrice) ? 'error' : ''"
              />
              <span v-if="fieldInvalid(givePriceChecked, givePrice)" class="error-text">请输入</span>
            </div>
            <span class="ml-6px">元</span>
          </div>
          <div class="flex flex-items-center mb-12px">
            <a-checkbox v-model:checked="giveDaysChecked" class="text-#666 ">
              <div class="flex flex-items-center" @click.stop>
                赠送天数 ＞
              </div>
            </a-checkbox>
            <div class="flex flex-col">
              <a-input-number
                v-model:value="giveDays" :precision="0" placeholder="请输入" :min="1"
                class="w-100px mr-6px input-stop-propagation"
                :status="fieldInvalid(giveDaysChecked, giveDays) ? 'error' : ''"
              />
              <span v-if="fieldInvalid(giveDaysChecked, giveDays)" class="error-text">请输入</span>
            </div>
            <span class="ml-6px">天</span>
          </div>
        </div>
      </div>
      <!-- 按优惠金额 -->
      <div v-if="approvalCriteria === '2'" class="mt-18px flex flex-items-start">
        <div class="text-#222  text-14px mt-8px w-100px text-right">
          按优惠金额：
        </div>
        <!-- 多选 -->
        <div class="flex  flex-col">
          <div class="flex flex-items-center mb-12px">
            <a-checkbox v-model:checked="discountRateChecked" class="text-#666 ">
              <div class="flex flex-items-center" @click.stop>
                整单优惠折扣低于
              </div>
            </a-checkbox>
            <div class="flex flex-col">
              <a-input-number
                v-model:value="discountRate" :precision="1" placeholder="请输入" :min="0.1" :max="9.9"
                class="w-100px mr-6px input-stop-propagation"
                :status="fieldInvalid(discountRateChecked, discountRate) ? 'error' : ''"
              />
              <span v-if="fieldInvalid(discountRateChecked, discountRate)" class="error-text">请输入</span>
            </div>
            <span class="ml-6px">折</span>

            <a-popover title="说明">
              <template #content>
                <div class="w-300px">
                  如：整单优惠折扣低于 9 折，即生成审批。一笔应付 1000 元的报名订单，优惠金额大于 100 元时，需要审批。
                </div>
              </template>
              <QuestionCircleOutlined class="cursor-pointer ml-6px" />
            </a-popover>
          </div>
          <div class="flex flex-items-center mb-12px">
            <a-checkbox v-model:checked="discountAmountChecked" class="text-#666 ">
              <div class="flex flex-items-center" @click.stop>
                整单优惠金额超过
              </div>
            </a-checkbox>
            <div class="flex flex-col">
              <a-input-number
                v-model:value="discountAmount" :precision="2" placeholder="请输入" :min="0.01"
                class="w-100px mr-6px input-stop-propagation"
                :status="fieldInvalid(discountAmountChecked, discountAmount) ? 'error' : ''"
              />
              <span v-if="fieldInvalid(discountAmountChecked, discountAmount)" class="error-text">请输入</span>
            </div>
            <span class="ml-6px">元</span>
          </div>
        </div>
      </div>
      <!-- 审批流程 -->
      <div class="mt-20px flex flex-items-start">
        <div class="text-#222  text-14px mt-7px w-100px text-right">
          审批流程：
        </div>
        <div class="flex-1">
          <a-button type="primary" ghost @click="handleConfigApprovalProcess">
            配置审批流程
          </a-button>
          <a-alert v-if="list.length > 0" class="bg-#fafafa border-#ddd mt-8px px-16px py-4px">
            <template #message>
              <a-timeline class="mt-10px">
                <a-timeline-item v-for="(item, index) in list" :key="index">
                  <span class="text-#888 relative top-2px"> {{ item.name }}</span>
                </a-timeline-item>
              </a-timeline>
            </template>
          </a-alert>
        </div>
      </div>
    </div>
    <ConfigApprovalProcess
      v-model:open="approvalProcessOpen"
      title="报名续费审批流程"
      :flow-models="props.templateData?.flowModels || []"
      @save="handleProcessSave"
    />
  </div>
</template>

<style lang="less" scoped>
:deep(.ant-timeline .ant-timeline-item-content){
  min-height: 20px;
  margin-left: 15px !important;
}
:deep(.ant-timeline .ant-timeline-item){
  padding-bottom: 6px;
}

:deep(.ant-timeline .ant-timeline-item-head){
  background-color: #06f;
  border:none;
  margin-top: 2px;
  width: 7px;
  height: 7px;
  margin-left: 1px;
}

:deep(.ant-timeline .ant-timeline-item:last-child){
  padding-bottom: 0px;
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

.input-stop-propagation {

  ::v-deep(.ant-input-number-handler-wrap),
  ::v-deep(.ant-input-number-input-wrap) {
    pointer-events: auto;
  }

  ::v-deep(.ant-input-number-handler-up),
  ::v-deep(.ant-input-number-handler-down) {
    pointer-events: auto;
  }
}

.flex-items-center {
  position: relative;
}

.error-text {
  margin-top: 6px;
  font-size: 12px;
  line-height: 1;
  color: #ff4d4f;
}
</style>
