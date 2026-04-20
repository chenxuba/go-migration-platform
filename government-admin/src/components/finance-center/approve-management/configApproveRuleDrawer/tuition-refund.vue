<script setup>
import { ref, watch } from 'vue'

const props = defineProps({
  checked: {
    type: Boolean,
    default: false,
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

const giveClassTime = ref(0) // 赠送课时
const givePrice = ref(0) // 赠送金额
const giveDays = ref(0) // 赠送天数

const discountRateChecked = ref(false) // 整单优惠折扣开关
const discountAmountChecked = ref(false) // 整单优惠金额开关

const discountRate = ref(0) // 整单优惠折扣
const discountAmount = ref(0) // 整单优惠金额

const list = ref([]) // 审批流程

watch(() => props.templateData, (val) => {
  list.value = Array.isArray(val?.flowModels)
    ? val.flowModels.map(item => ({
        name: Array.isArray(item.staffNames) && item.staffNames.length > 0
          ? item.staffNames.join('、')
          : (item.staffIds || []).join('、'),
      }))
    : []
}, { immediate: true, deep: true })

const approvalProcessOpen = ref(false) // 审批流程弹窗

function handleConfigApprovalProcess() {
  approvalProcessOpen.value = true
}

function handleProcessSave(flowModels) {
  emit('update:templateData', {
    ...props.templateData,
    enable: checkedValue.value,
    flowModels,
  })
}
</script>

<template>
  <div class="mt-12px">
    <div class="flex flex-items-center">
      <div class="text-#222 font500 text-20px mr-14px w-126px text-right">
        退学杂教材费
      </div>
      <a-switch v-model:checked="checkedValue" />
    </div>
    <!-- 审批条件 -->
    <div v-if="checkedValue" class="mt-20px w-700px ">
      <!-- 审批条件 单选 -->
      <div class="flex flex-items-center">
        <div class="text-#222  text-14px mr-0px w-100px text-right">
          审批条件：
        </div>
        <a-radio-group v-model:value="approvalCriteria" class="custom-radio">
          <a-radio value="1" class="text-#666">
            不限制，订单提交/支付后即生成审批
          </a-radio>
          <!-- <a-radio value="2" class="text-#666">限制条件 (可多选，满足任一条件即生成审批)</a-radio> -->
        </a-radio-group>
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
      title="退学杂教材费审批流程"
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
</style>
