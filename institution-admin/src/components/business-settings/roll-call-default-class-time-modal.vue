<script setup lang="ts">
import { CloseOutlined, InfoCircleFilled } from '@ant-design/icons-vue'
import { computed, reactive, ref, watch } from 'vue'
import { type InstConfig, setInstConfigApi } from '~@/api/common/config'
import { useUserStore } from '~@/stores/user'
import messageService from '~@/utils/messageService'

const props = defineProps<{
  open: boolean
  instConfig?: Partial<InstConfig>
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
  saved: []
}>()

const userStore = useUserStore()
const submitting = ref(false)

const openModal = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

const formState = reactive({
  defaultClassTimeRecordMode: 1,
  defaultStudentClassTime: 1,
  defaultTeacherClassTime: 0,
})

const classTimeUnitLabel = computed(() =>
  Number(formState.defaultClassTimeRecordMode) === 2 ? '课时/小时' : '课时',
)

function syncFormState() {
  formState.defaultClassTimeRecordMode = Number(props.instConfig?.defaultClassTimeRecordMode || 1)
  formState.defaultStudentClassTime = Number(props.instConfig?.defaultStudentClassTime ?? 1)
  formState.defaultTeacherClassTime = Number(props.instConfig?.defaultTeacherClassTime ?? 0)
}

watch(() => props.open, (value) => {
  if (value)
    syncFormState()
}, { immediate: true })

function closeFun() {
  if (submitting.value)
    return
  openModal.value = false
}

async function handleSubmit() {
  if (submitting.value)
    return
  if (!Number.isFinite(Number(formState.defaultStudentClassTime)) || Number(formState.defaultStudentClassTime) <= 0) {
    messageService.error('学员记录课时需大于 0')
    return
  }
  if (!Number.isFinite(Number(formState.defaultTeacherClassTime)) || Number(formState.defaultTeacherClassTime) < 0) {
    messageService.error('教师授课课时不能小于 0')
    return
  }

  submitting.value = true
  try {
    await setInstConfigApi({
      ...(props.instConfig as InstConfig),
      defaultClassTimeRecordMode: Number(formState.defaultClassTimeRecordMode || 1),
      defaultStudentClassTime: String(Number(formState.defaultStudentClassTime || 0)),
      defaultTeacherClassTime: String(Number(formState.defaultTeacherClassTime || 0)),
    })
    await userStore.getInstConfig()
    messageService.success('默认记录课时已更新')
    emit('saved')
    openModal.value = false
  }
  catch (error) {
    console.error('save default class time failed', error)
    messageService.error('保存失败，请稍后重试')
  }
  finally {
    submitting.value = false
  }
}
</script>

<template>
  <a-modal
    v-model:open="openModal"
    class="modal-content-box"
    wrap-class-name="roll-call-default-class-time-modal-wrap"
    centered
    :keyboard="false"
    :closable="false"
    :mask-closable="false"
    :width="560"
    destroy-on-close
    :confirm-loading="submitting"
    :body-style="{ maxHeight: '68vh', overflowY: 'auto', padding: '0' }"
  >
    <template #title>
      <div class="modal-title-bar">
        <span>默认记录课时</span>
        <a-button type="text" class="close-btn" @click="closeFun">
          <template #icon>
            <CloseOutlined class="close-icon" />
          </template>
        </a-button>
      </div>
    </template>

    <div class="default-class-time-modal__body">
      <div class="default-class-time-modal__notice">
        <InfoCircleFilled class="default-class-time-modal__notice-icon" />
        <span>编辑后，不会影响已创建的班级和1对1</span>
      </div>

      <div class="default-class-time-modal__form">
        <div class="default-class-time-modal__row">
          <div class="default-class-time-modal__label">
            <span class="default-class-time-modal__required">*</span>
            <span>课时记录方式:</span>
          </div>
          <div class="default-class-time-modal__content default-class-time-modal__content--radio">
            <a-radio-group v-model:value="formState.defaultClassTimeRecordMode" class="custom-radio">
              <a-radio :value="1">
                按固定课时记录
              </a-radio>
              <a-radio :value="2">
                按上课时长记录
              </a-radio>
            </a-radio-group>
          </div>
        </div>

        <div class="default-class-time-modal__row">
          <div class="default-class-time-modal__label">
            <span class="default-class-time-modal__required">*</span>
            <span>学员记录课时:</span>
          </div>
          <div class="default-class-time-modal__content">
            <a-input-number
              v-model:value="formState.defaultStudentClassTime"
              :min="0"
              :precision="2"
              :controls="false"
              class="default-class-time-modal__input"
            />
            <span class="default-class-time-modal__unit">{{ classTimeUnitLabel }}</span>
          </div>
        </div>

        <div class="default-class-time-modal__row">
          <div class="default-class-time-modal__label">
            <span class="default-class-time-modal__required">*</span>
            <span>教师授课课时:</span>
          </div>
          <div class="default-class-time-modal__content">
            <a-input-number
              v-model:value="formState.defaultTeacherClassTime"
              :min="0"
              :precision="2"
              :controls="false"
              class="default-class-time-modal__input"
            />
            <span class="default-class-time-modal__unit">{{ classTimeUnitLabel }}</span>
          </div>
        </div>
      </div>
    </div>

    <template #footer>
      <a-button danger ghost @click="closeFun">
        取消
      </a-button>
      <a-button type="primary" ghost :loading="submitting" @click="handleSubmit">
        确定
      </a-button>
    </template>
  </a-modal>
</template>

<style lang="less" scoped>
.modal-title-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 20px;
  font-weight: 600;
  color: #1f2329;
}

.close-btn {
  &:hover {
    background: transparent;
  }
}

.close-icon {
  font-size: 18px;
  color: #909399;
}

.default-class-time-modal__body {
  padding: 18px 20px 8px;
  background: #fff;
}

.default-class-time-modal__notice {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  margin-bottom: 18px;
  color: #1d4ed8;
  background: #edf4ff;
  border-radius: 8px;
  line-height: 22px;
}

.default-class-time-modal__notice-icon {
  font-size: 16px;
  flex-shrink: 0;
}

.default-class-time-modal__form {
  display: flex;
  flex-direction: column;
  gap: 18px;
  padding-left: 10px;
}

.default-class-time-modal__row {
  display: flex;
  align-items: center;
  min-height: 32px;
  column-gap: 4px;
}

.default-class-time-modal__label {
  width: 136px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: flex-start;
  color: #1f2329;
  line-height: 22px;
}

.default-class-time-modal__required {
  margin-right: 4px;
  color: #f53f3f;
}

.default-class-time-modal__content {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}

.default-class-time-modal__content--radio {
  :deep(.ant-radio-group) {
    display: flex;
    align-items: center;
    gap: 18px;
    flex-wrap: nowrap;
  }

  :deep(.ant-radio-wrapper) {
    margin-inline-start: 0;
    white-space: nowrap;
  }
}

.default-class-time-modal__input {
  width: 240px;
  flex: 0 0 240px;
}

.default-class-time-modal__unit {
  color: #4e5969;
  white-space: nowrap;
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

<style>
.roll-call-default-class-time-modal-wrap .ant-modal-content {
  overflow: hidden;
  border-radius: 16px;
}

.roll-call-default-class-time-modal-wrap .ant-modal-header {
  padding: 14px 20px !important;
  margin-bottom: 0;
  border-bottom: 1px solid #f0f0f0;
}

.roll-call-default-class-time-modal-wrap .ant-modal-body {
  padding: 0 !important;
}

.roll-call-default-class-time-modal-wrap .ant-modal-footer {
  padding: 12px 20px 16px;
  border-top: 1px solid #f0f0f0;
}
</style>
