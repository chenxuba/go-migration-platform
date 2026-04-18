<script setup>
import { ExclamationCircleFilled } from '@ant-design/icons-vue'
import { computed, reactive, ref, watch } from 'vue'
import messageService from '~@/utils/messageService'

const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
  employeeNames: {
    type: String,
    required: true,
    default: '',
  },
  employeeCount: {
    type: Number,
    default: 1,
  },
  loading: {
    type: Boolean,
    default: false,
  },
})

const emit = defineEmits(['update:open', 'cancel', 'confirm'])

const formRef = ref(null)
const formState = reactive({
  confirmInput: '',
  riskAcknowledged: false,
})

const openModal = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

const isBatch = computed(() => props.employeeCount > 1)

const title = computed(() => isBatch.value ? `确定批量离职 ${props.employeeCount} 位员工？` : `确定要将"${props.employeeNames}"设置为离职吗？`)

const expectedInput = computed(() => {
  return isBatch.value ? `确认删除${props.employeeCount}人` : props.employeeNames
})

function validateInput(rule, value) {
  if (!value) {
    return Promise.reject('请输入确认信息')
  }
  return Promise.resolve()
}

function handleCancel() {
  openModal.value = false
  // 立即清空表单数据
  resetForm()
  // 延迟清空验证状态，确保在 modal 动画完成后执行
  setTimeout(() => {
    if (formRef.value) {
      formRef.value.clearValidate()
    }
  }, 100)
  emit('cancel')
}

function resetForm() {
  formState.confirmInput = ''
  formState.riskAcknowledged = false
  // 清空表单验证状态
  if (formRef.value) {
    formRef.value.clearValidate()
  }
}

async function handleSubmit() {
  try {
    await formRef.value.validate()

    // 单人：检测输入的值是否与员工姓名一致
    if (props.employeeCount === 1) {
      if (formState.confirmInput.trim() !== props.employeeNames) {
        messageService.error('输入员工姓名与所选离职员工不一致')
        return
      }
    }

    // 批量：检测输入的值是否与确认指令一致
    if (props.employeeCount > 1) {
      if (formState.confirmInput.trim() !== `确认离职${props.employeeCount}人`) {
        messageService.error('确认指令输入错误')
        return
      }
    }

    if (!formState.riskAcknowledged) {
      messageService.warning('请勾选"我已阅读并知晓以上风险"')
      return
    }

    emit('confirm')
  }
  catch (error) {
    console.error('Validation failed:', error)
  }
}

// 监听openModal的变化，当关闭时清空表单
watch(() => openModal.value, (newVal) => {
  if (!newVal) {
    resetForm()
    // 确保验证状态也被清空
    setTimeout(() => {
      if (formRef.value) {
        formRef.value.clearValidate()
      }
    }, 100)
  }
})
</script>

<template>
  <a-modal
    v-model:open="openModal" centered :footer="false" :closable="false" :mask-closable="false"
    :keyboard="false" width="580px"
  >
    <div class="text-16px mb-16px font-400 flex items-start gap-12px">
      <ExclamationCircleFilled class="text-#f90 text-24px" />
      <div class="flex-1">
        <div class="text-#222 font-400 mb-8px">
          {{ title }}
        </div>
        <div v-if="isBatch" class="text-#666 text-14px">
          <span class="text-#222 font500">
            员工姓名：
          </span>{{ employeeNames }}
        </div>
      </div>
    </div>

    <div class="ml-30px text-#666 text-14px leading-20px">
      <div class="mb-8px">
        1、离职后，该账户将无法登录系统
      </div>
      <div class="mb-8px">
        2、该员工的离职，不会影响历史已生成的相关工作记录，也不会影响未来已安排的工作，如需更改负责老师，请去相关日程进行手动更改。
      </div>
      <div class="mb-16px">
        3、员工离职后，意向学员所填写销售为此员工，会清空意向学员所属销售清空。
      </div>
      <div class="mb-16px">
        4、操作员工离职会自动清除流程审批中的审批流，如果审批流里只有该员工，请先更改其他员工后再操作离职。
      </div>

      <template v-if="isBatch">
        <div class="mb-8px">
          如确认离职，请输入以下确认指令：<span class="text-#f00">确认离职{{ employeeCount }}人</span>
        </div>
      </template>
      <template v-else>
        <div class="mb-8px">
          如确认离职，请输入员工姓名：<span class="text-#f00">{{ employeeNames }}</span>
        </div>
      </template>

      <a-form ref="formRef" :model="formState" class="mb-16px">
        <a-form-item name="confirmInput" :rules="[{ required: true, message: '请输入确认信息', validator: validateInput }]">
          <a-input
            v-model:value="formState.confirmInput"
            :placeholder="isBatch ? '请输入离职指令' : '请输入要离职的员工姓名'"
          />
        </a-form-item>
      </a-form>

      <div class="mb-24px flex items-center">
        <a-checkbox v-model:checked="formState.riskAcknowledged" class="text-#666">
          <span class="text-#222 text-15px font500">
            我已阅读并知晓以上风险
          </span>
        </a-checkbox>
      </div>
    </div>

    <a-space class="flex justify-end">
      <a-button @click="handleCancel">
        取消
      </a-button>
      <a-button type="primary" :loading="loading" :disabled="!formState.riskAcknowledged" @click="handleSubmit">
        确认离职
      </a-button>
    </a-space>
  </a-modal>
</template>

<style scoped>
:deep(.ant-form-item) {
  margin-bottom: 0;
}

:deep(.ant-form-item-explain-error) {
  margin-top: 4px;
}
</style>
