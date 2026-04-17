<script setup lang="ts">
import type { TableColumnsType } from 'ant-design-vue'
import type { FormInstance, Rule } from 'ant-design-vue/es/form'
import type {
  InstitutionDetail,
  InstitutionRenewalMutationPayload,
  InstitutionRenewalRecord,
} from '@/api/platform/institutions'
import { CloseOutlined } from '@ant-design/icons-vue'
import {
  getInstitutionDetailApi,
  getInstitutionRenewalRecordsApi,
  renewInstitutionApi,
} from '@/api/platform/institutions'
import messageService from '@/utils/messageService'

const props = defineProps<{
  open: boolean
  institutionId?: number | null
}>()

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'renewed'): void
}>()

interface RenewalFormState {
  openType?: number
  openDuration?: string
}

const openModal = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

const formRef = ref<FormInstance>()
const loading = ref(false)
const submitting = ref(false)
const detail = ref<InstitutionDetail>()
const records = ref<InstitutionRenewalRecord[]>([])
const formState = reactive<RenewalFormState>({
  openType: undefined,
  openDuration: undefined,
})

const openTypeOptions = [
  { value: 1, label: '体验版' },
  { value: 2, label: '正式版' },
]
const openDurationOptionMap: Record<number, { value: string, label: string }[]> = {
  1: [
    { value: '3d', label: '3天' },
    { value: '5d', label: '5天' },
    { value: '7d', label: '7天' },
  ],
  2: [
    { value: '1y', label: '1年' },
    { value: '2y', label: '2年' },
    { value: '3y', label: '3年' },
    { value: '5y', label: '5年' },
    { value: '99y', label: '99年' },
  ],
}

const availableOpenTypeOptions = computed(() => {
  if (Number(detail.value?.openType || 2) === 2)
    return openTypeOptions.filter(item => item.value === 2)
  return openTypeOptions
})
const openDurationOptions = computed(() => openDurationOptionMap[Number(formState.openType) || 2] || openDurationOptionMap[2])
const previewExpireEndTime = computed(() => {
  const duration = String(formState.openDuration || '').trim()
  if (!duration)
    return '--'

  const currentExpireEnd = parseDateTime(detail.value?.expireEndTime)
  const now = new Date()
  const effectiveEnd = currentExpireEnd && currentExpireEnd.getTime() > now.getTime() ? currentExpireEnd : now

  return formatDateMinute(formatDateTime(addDuration(effectiveEnd, duration)))
})

const columns: TableColumnsType<InstitutionRenewalRecord> = [
  {
    title: '续期时间',
    dataIndex: 'createTime',
    key: 'createTime',
    width: 160,
  },
  {
    title: '续前类型',
    dataIndex: 'beforeOpenType',
    key: 'beforeOpenType',
    width: 100,
    align: 'center' as const,
  },
  {
    title: '续期类型',
    dataIndex: 'afterOpenType',
    key: 'afterOpenType',
    width: 100,
    align: 'center' as const,
  },
  {
    title: '续期时长',
    dataIndex: 'renewDuration',
    key: 'renewDuration',
    width: 100,
    align: 'center' as const,
  },
  {
    title: '续前到期时间',
    dataIndex: 'beforeExpireEndTime',
    key: 'beforeExpireEndTime',
    width: 160,
  },
  {
    title: '续后到期时间',
    dataIndex: 'afterExpireEndTime',
    key: 'afterExpireEndTime',
    width: 160,
  },
]

const rules: Record<string, Rule[]> = {
  openType: [{ required: true, message: '请选择开通类型', trigger: 'change' }],
  openDuration: [{ required: true, message: '请选择续期时长', trigger: 'change' }],
}

function closeModal() {
  emit('update:open', false)
}

function formatDateMinute(value?: string) {
  const raw = String(value || '').trim()
  if (!raw)
    return '--'

  return raw.length >= 16 ? raw.slice(0, 16) : raw
}

function parseDateTime(value?: string) {
  const raw = String(value || '').trim()
  if (!raw)
    return null

  const normalized = raw.replace(/-/g, '/')
  const parsed = new Date(normalized)
  if (Number.isNaN(parsed.getTime()))
    return null

  return parsed
}

function formatDateTime(value: Date) {
  const year = value.getFullYear()
  const month = String(value.getMonth() + 1).padStart(2, '0')
  const date = String(value.getDate()).padStart(2, '0')
  const hours = String(value.getHours()).padStart(2, '0')
  const minutes = String(value.getMinutes()).padStart(2, '0')
  const seconds = String(value.getSeconds()).padStart(2, '0')
  return `${year}-${month}-${date} ${hours}:${minutes}:${seconds}`
}

function addDuration(source: Date, duration: string) {
  const result = new Date(source.getTime())

  switch (duration) {
    case '3d':
      result.setDate(result.getDate() + 3)
      break
    case '5d':
      result.setDate(result.getDate() + 5)
      break
    case '7d':
      result.setDate(result.getDate() + 7)
      break
    case '2y':
      result.setFullYear(result.getFullYear() + 2)
      break
    case '3y':
      result.setFullYear(result.getFullYear() + 3)
      break
    case '5y':
      result.setFullYear(result.getFullYear() + 5)
      break
    case '99y':
      result.setFullYear(result.getFullYear() + 99)
      break
    default:
      result.setFullYear(result.getFullYear() + 1)
      break
  }

  return result
}

function getOpenTypeLabel(value?: number) {
  return Number(value) === 1 ? '体验版' : '正式版'
}

function getOpenDurationLabel(openType: number, value?: string) {
  return openDurationOptionMap[openType]?.find(item => item.value === String(value || '').trim())?.label || '--'
}

function resolveDurationValue(openType: number, preferred?: string) {
  const options = openDurationOptionMap[openType] || []
  const matched = options.find(item => item.value === String(preferred || '').trim())
  return matched?.value || options[0]?.value
}

function resetState() {
  detail.value = undefined
  records.value = []
  formState.openType = undefined
  formState.openDuration = undefined
  nextTick(() => {
    formRef.value?.clearValidate?.()
  })
}

function applyDefaultForm(detailData: InstitutionDetail) {
  const currentType = Number(detailData.openType || 2) || 2
  formState.openType = currentType
  formState.openDuration = resolveDurationValue(currentType, detailData.openDuration)
  nextTick(() => {
    formRef.value?.clearValidate?.()
  })
}

function handleOpenTypeChange(value?: number | string) {
  formState.openType = value ? Number(value) : undefined
  formState.openDuration = resolveDurationValue(Number(formState.openType) || 2)
}

async function loadData(id: number) {
  loading.value = true
  try {
    const [detailRes, recordRes] = await Promise.all([
      getInstitutionDetailApi({ id }),
      getInstitutionRenewalRecordsApi({ institutionId: id }),
    ])

    if (detailRes.code !== 200 || !detailRes.result) {
      messageService.error(detailRes.message || '获取机构信息失败')
      return
    }
    if (recordRes.code !== 200) {
      messageService.error(recordRes.message || '获取续期记录失败')
      return
    }

    detail.value = detailRes.result
    records.value = Array.isArray(recordRes.result) ? recordRes.result : []
    applyDefaultForm(detailRes.result)
  }
  catch (error: any) {
    console.error('load institution renewal data failed', error)
    messageService.error(error?.message || '获取续期信息失败')
  }
  finally {
    loading.value = false
  }
}

function buildPayload(): InstitutionRenewalMutationPayload | null {
  const institutionId = Number(props.institutionId || 0)
  const openType = Number(formState.openType || 0)
  const openDuration = String(formState.openDuration || '').trim()
  if (!institutionId || !openType || !openDuration)
    return null

  return {
    institutionId,
    openType,
    openDuration,
  }
}

async function submitRenewal() {
  try {
    await formRef.value?.validate()
  }
  catch {
    return
  }

  const payload = buildPayload()
  if (!payload)
    return

  submitting.value = true
  try {
    const res = await renewInstitutionApi(payload)
    if (res.code !== 200) {
      messageService.error(res.message || '机构续期失败')
      return
    }

    messageService.success('机构续期成功')
    emit('renewed')
    await loadData(payload.institutionId)
  }
  catch (error: any) {
    console.error('renew institution failed', error)
    messageService.error(error?.message || '机构续期失败')
  }
  finally {
    submitting.value = false
  }
}

watch(
  () => [props.open, props.institutionId] as const,
  ([open, institutionId]) => {
    if (!open) {
      resetState()
      return
    }

    if (institutionId)
      void loadData(Number(institutionId))
  },
  { immediate: true },
)
</script>

<template>
  <a-modal
    v-model:open="openModal"
    centered
    destroy-on-close
    :keyboard="false"
    :closable="false"
    :mask-closable="false"
    :width="1080"
    class="createStu-modal-content-box institution-renewal-modal"
  >
    <template #title>
      <div class="institution-renewal__titlebar">
        <span>机构续期</span>
        <a-button type="text" class="close-btn" @click="closeModal">
          <template #icon>
            <CloseOutlined class="close-icon" />
          </template>
        </a-button>
      </div>
    </template>

    <a-spin :spinning="loading">
      <div class="institution-renewal__content scrollbar">
        <div class="renewal-toolbar">
          <div class="renewal-toolbar__row">
            <div class="renewal-toolbar__caption">
              当前信息
            </div>
            <div class="renewal-facts">
              <div class="renewal-fact renewal-fact--wide">
                <span class="renewal-fact__label">机构名称</span>
                <span class="renewal-fact__value renewal-fact__value--ellipsis" :title="detail?.organName || '--'">
                  {{ detail?.organName || '--' }}
                </span>
              </div>

              <div class="renewal-fact">
                <span class="renewal-fact__label">当前开通类型</span>
                <span class="renewal-fact__value">{{ getOpenTypeLabel(detail?.openType) }}</span>
              </div>

              <div class="renewal-fact">
                <span class="renewal-fact__label">当前到期时间</span>
                <span class="renewal-fact__value">{{ formatDateMinute(detail?.expireEndTime) }}</span>
              </div>
            </div>
          </div>

          <div class="renewal-toolbar__row renewal-toolbar__row--form">
            <div class="renewal-toolbar__caption">
              续期设置
            </div>
            <a-form ref="formRef" layout="inline" :model="formState" :rules="rules" class="renewal-inline-form">
              <a-form-item label="开通类型" name="openType">
                <a-select
                  v-model:value="formState.openType"
                  :options="availableOpenTypeOptions"
                  class="renewal-inline-form__select"
                  placeholder="请选择开通类型"
                  @change="handleOpenTypeChange"
                />
              </a-form-item>

              <a-form-item label="续期时长" name="openDuration">
                <a-select
                  v-model:value="formState.openDuration"
                  :options="openDurationOptions"
                  :disabled="!formState.openType"
                  class="renewal-inline-form__select"
                  placeholder="请选择续期时长"
                />
              </a-form-item>

              <a-form-item label="续后到期时间" class="renewal-inline-form__preview-item">
                <div class="renewal-inline-preview">
                  {{ previewExpireEndTime }}
                </div>
              </a-form-item>
            </a-form>
          </div>
        </div>

        <div class="renewal-section renewal-section--records">
          <div class="renewal-section__title">
            续期记录
          </div>
          <a-table
            :columns="columns"
            :data-source="records"
            :pagination="false"
            :scroll="{ x: 880, y: 320 }"
            row-key="id"
            size="small"
            class="renewal-table"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'createTime'">
                {{ formatDateMinute(record.createTime) }}
              </template>

              <template v-else-if="column.key === 'beforeOpenType'">
                {{ getOpenTypeLabel(record.beforeOpenType) }}
              </template>

              <template v-else-if="column.key === 'afterOpenType'">
                {{ getOpenTypeLabel(record.afterOpenType) }}
              </template>

              <template v-else-if="column.key === 'renewDuration'">
                {{ getOpenDurationLabel(record.afterOpenType, record.renewDuration) }}
              </template>

              <template v-else-if="column.key === 'beforeExpireEndTime'">
                {{ formatDateMinute(record.beforeExpireEndTime) }}
              </template>

              <template v-else-if="column.key === 'afterExpireEndTime'">
                {{ formatDateMinute(record.afterExpireEndTime) }}
              </template>
            </template>
          </a-table>
        </div>
      </div>
    </a-spin>

    <template #footer>
      <a-button danger ghost @click="closeModal">
        关闭
      </a-button>
      <a-button type="primary" ghost :loading="submitting" @click="submitRenewal">
        确定续期
      </a-button>
    </template>
  </a-modal>
</template>

<style scoped>
.institution-renewal__titlebar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  color: #1f2329;
  font-size: 22px;
  font-weight: 700;
  line-height: 32px;
}

.institution-renewal__content {
  max-height: calc(100vh - 155px);
  padding: 20px 32px 0 !important;
  overflow: auto;
}

.renewal-toolbar {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-bottom: 14px;
  padding: 14px 18px;
  background: #fff;
  border: 1px solid #edf1f7;
  border-radius: 14px;
}

.renewal-toolbar__row {
  display: grid;
  grid-template-columns: 72px minmax(0, 1fr);
  align-items: start;
  gap: 12px;
  min-width: 0;
}

.renewal-toolbar__caption {
  color: #262626;
  font-size: 13px;
  font-weight: 600;
  line-height: 32px;
}

.renewal-facts {
  display: flex;
  flex-wrap: wrap;
  gap: 8px 10px;
}

.renewal-fact {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
  padding: 8px 12px;
  background: #fafbfc;
  border: 1px solid #f0f2f5;
  border-radius: 10px;
}

.renewal-fact--wide {
  flex: 1 1 320px;
}

.renewal-fact__label {
  flex-shrink: 0;
  color: #8c8c8c;
  font-size: 12px;
  line-height: 20px;
}

.renewal-fact__value {
  color: #262626;
  font-size: 13px;
  font-weight: 600;
  line-height: 20px;
}

.renewal-fact__value--ellipsis {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.renewal-inline-form {
  display: flex;
  flex-wrap: wrap;
  gap: 0 12px;
  width: 100%;
}

.renewal-inline-form :deep(.ant-form-item) {
  margin-right: 0;
  margin-bottom: 0;
}

.renewal-inline-form__select {
  width: 152px;
}

.renewal-inline-form__preview-item {
  flex: 1 1 0;
  min-width: 260px;
}

.renewal-inline-form__preview-item :deep(.ant-form-item-control) {
  flex: 1 1 auto;
  min-width: 0;
}

.renewal-inline-form__preview-item :deep(.ant-form-item-control-input) {
  min-height: 32px;
  width: 100%;
}

.renewal-inline-form__preview-item :deep(.ant-form-item-control-input-content) {
  width: 100%;
}

.renewal-inline-preview {
  display: flex;
  align-items: center;
  width: 100%;
  min-width: 0;
  height: 32px;
  padding: 0 11px;
  color: #262626;
  font-size: 13px;
  font-weight: 600;
  background: #fafbfc;
  border: 1px solid #d9d9d9;
  border-radius: 6px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.renewal-section {
  margin-bottom: 14px;
  padding: 16px 18px 12px;
  background: #fff;
  border: 1px solid #edf1f7;
  border-radius: 14px;
}

.renewal-section__title {
  margin-bottom: 16px;
  color: #262626;
  font-size: 15px;
  font-weight: 600;
  line-height: 22px;
}

.renewal-section--records {
  padding-bottom: 8px;
}

:deep(.renewal-table .ant-table) {
  background: transparent;
}

:deep(.institution-renewal-modal .ant-form-item) {
  margin-bottom: 8px;
}

:deep(.institution-renewal-modal .ant-modal-footer) {
  padding: 18px 32px 24px;
}

:deep(.createStu-modal-content-box.institution-renewal-modal .ant-modal-header) {
  padding: 26px 32px 18px;
}

:deep(.createStu-modal-content-box.institution-renewal-modal .ant-modal-body) {
  padding: 0;
}

@media (max-width: 768px) {
  .institution-renewal__content {
    padding: 16px 16px 0 !important;
  }

  .renewal-toolbar {
    padding: 14px;
  }

  .renewal-toolbar__row {
    grid-template-columns: 1fr;
    gap: 8px;
  }

  .renewal-inline-form {
    width: 100%;
  }

  .renewal-inline-form__select {
    width: 100%;
  }

  .renewal-inline-preview {
    min-width: 0;
    width: 100%;
  }
}
</style>
