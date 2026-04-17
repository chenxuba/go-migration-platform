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
        <div class="renewal-section">
          <div class="renewal-section__title">
            当前信息
          </div>
          <div class="renewal-summary">
            <div class="renewal-summary__item">
              <div class="renewal-summary__label">
                机构名称
              </div>
              <div class="renewal-summary__value renewal-summary__value--wide" :title="detail?.organName || '--'">
                {{ detail?.organName || '--' }}
              </div>
            </div>

            <div class="renewal-summary__item">
              <div class="renewal-summary__label">
                当前开通类型
              </div>
              <div class="renewal-summary__value">
                {{ getOpenTypeLabel(detail?.openType) }}
              </div>
            </div>

            <div class="renewal-summary__item">
              <div class="renewal-summary__label">
                当前开通时长
              </div>
              <div class="renewal-summary__value">
                {{ getOpenDurationLabel(Number(detail?.openType || 2), detail?.openDuration) }}
              </div>
            </div>

            <div class="renewal-summary__item">
              <div class="renewal-summary__label">
                当前到期时间
              </div>
              <div class="renewal-summary__value">
                {{ formatDateMinute(detail?.expireEndTime) }}
              </div>
            </div>
          </div>
        </div>

        <div class="renewal-section">
          <div class="renewal-section__title">
            续期设置
          </div>
          <a-form ref="formRef" layout="vertical" :model="formState" :rules="rules">
            <a-row :gutter="24">
              <a-col :xs="24" :md="12">
                <a-form-item label="开通类型：" name="openType">
                  <a-select
                    v-model:value="formState.openType"
                    :options="availableOpenTypeOptions"
                    placeholder="请选择开通类型"
                    @change="handleOpenTypeChange"
                  />
                </a-form-item>
              </a-col>

              <a-col :xs="24" :md="12">
                <a-form-item label="续期时长：" name="openDuration">
                  <a-select
                    v-model:value="formState.openDuration"
                    :options="openDurationOptions"
                    :disabled="!formState.openType"
                    placeholder="请选择续期时长"
                  />
                </a-form-item>
              </a-col>
            </a-row>
          </a-form>
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
  padding: 24px 40px 0 !important;
  overflow: auto;
}

.renewal-section {
  margin-bottom: 16px;
  padding: 20px 22px 18px;
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

.renewal-summary {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14px 24px;
}

.renewal-summary__item {
  min-width: 0;
}

.renewal-summary__label {
  margin-bottom: 6px;
  color: #8c8c8c;
  font-size: 12px;
  line-height: 20px;
}

.renewal-summary__value {
  color: #262626;
  font-size: 14px;
  font-weight: 600;
  line-height: 22px;
}

.renewal-summary__value--wide {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
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
    padding: 20px 20px 0 !important;
  }

  .renewal-summary {
    grid-template-columns: 1fr;
    gap: 12px;
  }
}
</style>
