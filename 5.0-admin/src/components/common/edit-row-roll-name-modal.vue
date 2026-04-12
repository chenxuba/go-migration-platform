<script setup lang="ts">
import { CloseOutlined } from '@ant-design/icons-vue'
import type { TeachingRecordDetailStudent } from '@/api/edu-center/class-record'
import messageService from '@/utils/messageService'

const props = withDefaults(defineProps<{
  open: boolean
  student?: TeachingRecordDetailStudent | null
}>(), {
  student: null,
})

const emit = defineEmits(['update:open'])
const formRef = ref()

const openModal = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

const formState = reactive({
  status: 0,
  editRecord: undefined as number | undefined,
  internalNote: '',
  externalNote: '',
})

const avatarUrl = computed(() => String(props.student?.avatar || '').trim() || 'https://cdn.schoolpal.cn/schoolpal/next-erp/avator_male.png')
const studentName = computed(() => String(props.student?.studentName || '').trim() || '-')
const chargingModeText = computed(() => {
  const mode = Number(props.student?.skuMode || 0)
  if (mode === 2)
    return '按时间'
  if (mode === 3)
    return '按金额'
  if (mode === 1)
    return '按课时'
  return '-'
})
const sourceTypeText = computed(() => {
  const type = Number(props.student?.sourceType || 0)
  if (type === 2)
    return '临时学员'
  if (type === 3 || type === 7)
    return '补课学员'
  if (type === 4)
    return '试听学员'
  if (type === 6)
    return '1对1学员'
  return '班级学员'
})
const tuitionAccountText = computed(() => String(props.student?.tuitionAccountName || '').trim() || '-')
const leftQuantityText = computed(() => {
  const quantity = Number(props.student?.leftQuantity || 0)
  if (!Number.isFinite(quantity))
    return '-'
  const text = Number.isInteger(quantity) ? String(quantity) : quantity.toFixed(2).replace(/\.?0+$/, '')
  return `${text}课时`
})
const quantityDisabled = computed(() => Number(props.student?.sourceType || 0) === 4 || Number(props.student?.skuMode || 0) === 2)

function syncFormState() {
  const student = props.student
  formState.status = Number(student?.status ?? 0)
  formState.editRecord = student && String(student.studentTeachingRecordId || '').trim()
    ? Number(student.quantity ?? 0)
    : undefined
  formState.internalNote = String(student?.remark || '')
  formState.externalNote = String(student?.externalRemark || '')
}

watch(
  () => [openModal.value, props.student] as const,
  ([open]) => {
    if (!open)
      return
    syncFormState()
  },
  { immediate: true },
)

async function handleSubmit() {
  try {
    await formRef.value?.validate()
    messageService.info('编辑点名保存接口暂未接入，当前先支持真实数据回显')
    openModal.value = false
  }
  catch {
  }
}

function closeFun() {
  formRef.value?.resetFields()
  openModal.value = false
}
</script>

<template>
  <a-modal
    v-model:open="openModal"
    centered
    class="modal-content-box"
    :keyboard="false"
    :closable="false"
    :mask-closable="false"
    :width="800"
  >
    <template #title>
      <div class="text-5 flex justify-between flex-center">
        <span>编辑点名</span>
        <a-button type="text" class="close-btn" @click="closeFun">
          <template #icon>
            <CloseOutlined class="text-5 close-icon" />
          </template>
        </a-button>
      </div>
    </template>
    <div class="rounded-12px bg-#fafafa mx-24px">
      <div class="contenter px6 py3" style="margin-bottom: 0;">
        <div class="avatar flex flex-items-center">
          <img :src="avatarUrl" class="w-40px h-40px rounded-full mr-8px" alt="">
          <span class="text-5 text-#222 font-800">{{ studentName }}</span>
        </div>
        <div class="text-14px text-#222 flex mt-12px flex-wrap">
          <div class="mr-60px mb-10px flex flex-items-center">
            <span class="text-#888">学员身份：</span>
            <span>{{ sourceTypeText }}</span>
          </div>
          <div class="mr-60px mb-10px">
            <span class="text-#888">课消方式：</span>
            <span>{{ chargingModeText }}</span>
          </div>
          <div class="mr-60px mb-10px">
            <span class="text-#888">扣费课程账户：</span>
            <span>{{ tuitionAccountText }}</span>
          </div>
          <div class="mr-60px mb-10px">
            <span class="text-#888">剩余课时：</span>
            <span>{{ leftQuantityText }}</span>
          </div>
        </div>
      </div>
      <div class="w-752px">
        <a-divider class="my-0" />
      </div>
      <div class="contenter scrollbar" style="margin: 0 24px 24px 24px;">
        <a-form ref="formRef" :model="formState" :label-col="{ span: 4 }" :wrapper-col="{ span: 18 }">
          <a-form-item label="编辑状态" name="status" :rules="[{ required: true, message: '请选择编辑状态' }]" class="mb-40px">
            <a-radio-group v-model:value="formState.status" class="custom-radio">
              <a-radio :value="0">
                未点名
              </a-radio>
              <a-radio :value="1">
                到课
              </a-radio>
              <a-radio :value="3">
                请假
              </a-radio>
              <a-radio :value="2">
                旷课
              </a-radio>
              <a-radio :value="4">
                未记录
              </a-radio>
            </a-radio-group>
          </a-form-item>

          <a-form-item label="编辑记录" name="editRecord">
            <div class="flex flex-items-center">
              <a-input-number v-model:value="formState.editRecord" :precision="2" :min="0" :max="100" :disabled="quantityDisabled" />
              <span class="ml-4px">课时</span>
            </div>
            <div v-if="quantityDisabled" class="text-12px text-#888 mt-6px">
              当前学员课消方式不记录课时
            </div>
          </a-form-item>

          <a-form-item label="编辑对内备注" name="internalNote">
            <a-input v-model:value="formState.internalNote" placeholder="选填（200字以内）" :maxlength="200" />
          </a-form-item>

          <a-form-item label="编辑对外备注" name="externalNote">
            <a-input v-model:value="formState.externalNote" placeholder="选填（200字以内）" :maxlength="200" />
          </a-form-item>
        </a-form>
      </div>
    </div>
    <template #footer>
      <a-button danger ghost @click="closeFun">
        取消
      </a-button>
      <a-button type="primary" ghost @click="handleSubmit">
        确定
      </a-button>
    </template>
  </a-modal>
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

.contenter {
  padding: 24px;
  background: #fafafa;
  margin: 24px;
}
</style>

<style>
.modal-content-box .ant-modal-header {
  padding: 10px 16px !important;
  margin-bottom: 0;
}

.modal-content-box .ant-modal-body {
  padding: 0 !important;
}
</style>
