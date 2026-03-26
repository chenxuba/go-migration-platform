<script setup>
import { CloseCircleOutlined } from '@ant-design/icons-vue'
import { computed, reactive, ref, watch } from 'vue'
import messageService from '~@/utils/messageService'

const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
  studentNames: {
    type: String,
    required: true,
    default: '',
  },
  studentCount: {
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
  deleteConfirm: '',
  checked: false,
})

const openModal = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

const isBatch = computed(() => props.studentCount > 1)

const title = computed(() => isBatch.value ? `确认批量删除 ${props.studentCount} 位学员` : '确认删除学员？')

const placeholder = computed(() => {
  return isBatch.value ? '请输入删除指令' : '请输入要删除的学员姓名'
})

function validateInput(rule, value) {
  if (!value) {
    return Promise.reject('请输入确认信息')
  }
  return Promise.resolve()
}

function handleCancel() {
  openModal.value = false
  formState.deleteConfirm = ''
  emit('cancel')
}

async function handleSubmit() {
  try {
    await formRef.value.validate()
    // 单人 检测输入的值是否与学员姓名一致
    if (props.studentCount === 1) {
      if (formState.deleteConfirm.trim() !== props.studentNames) {
        messageService.error('输入学员姓名与所选删除学员不一致')
        return
      }
    }
    // 批量 检测输入的值是否与删除指令一致
    if (props.studentCount > 1) {
      if (formState.deleteConfirm.trim() !== `确认删除${props.studentCount}人`) {
        messageService.error('删除指令输入错误')
        return
      }
      if (!formState.checked) {
        messageService.warning('请勾选“我已核对批量删除学员姓名”')
        return
      }
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
    formState.deleteConfirm = ''
    formState.checked = false
  }
})
</script>

<template>
  <a-modal
    v-model:open="openModal" centered :footer="false" :closable="false" :mask-closable="false"
    :keyboard="false" width="480px"
  >
    <div class="text-18px mb-12px font500 flex items-center">
      <CloseCircleOutlined class="text-#f00 mr2 text-5" /> {{ title }}
    </div>
    <div class="pl-30px text-#666">
      <template v-if="isBatch">
        <div>已选择删除学员：<span class="text-#f00">{{ studentNames }}</span></div>
        <div>删除后，此学员的基本信息，包括相关订单、报读课程、上课记录等相关功能的 <span class="text-#f00">数据将会被永久清除，此操作不可撤销</span></div>
        <div>备注：在报表功能中将会自动生成冲正记录进行统计，保证历史数据的准确性</div>
        <div class="mt-12px mb-6px">
          如确认删除，请输入以下删除指令：<span class="text-#f00">确认删除{{ studentCount }}人</span>
        </div>
      </template>
      <template v-else>
        <div>已选择删除学员：<span class="text-#f00">{{ studentNames }}</span></div>
        <div>删除后，此学员的基本信息，包括相关订单、报读课程、上课记录等相关功能的 <span class="text-#f00">数据将会被永久清除，此操作不可撤销</span></div>
        <div class="mt-12px mb-6px">
          如确认删除，请输入学员姓名：<span class="text-#f00">{{ studentNames }}</span>
        </div>
      </template>
      <a-form ref="formRef" :model="formState">
        <a-form-item name="deleteConfirm" :rules="[{ required: true, message: '请输入确认信息', validator: validateInput }]">
          <a-input v-model:value="formState.deleteConfirm" :placeholder="placeholder" @copy.prevent @paste.prevent @cut.prevent @contextmenu.prevent />
        </a-form-item>
        <a-form-item v-if="isBatch">
          <a-checkbox v-model:checked="formState.checked" class="text-#666">
            我已核对批量删除学员姓名
          </a-checkbox>
        </a-form-item>
      </a-form>
    </div>
    <a-space class="mt-24px flex justify-end">
      <a-button @click="handleCancel">
        取消
      </a-button>
      <a-button type="primary" :loading="loading" @click="handleSubmit">
        确认
      </a-button>
    </a-space>
  </a-modal>
</template>
