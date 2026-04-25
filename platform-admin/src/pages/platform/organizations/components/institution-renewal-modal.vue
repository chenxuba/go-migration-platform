<script setup lang="ts">
import type { TableColumnsType } from 'ant-design-vue'
import type { Dayjs } from 'dayjs'
import type { FormInstance, Rule } from 'ant-design-vue/es/form'
import dayjs from 'dayjs'
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
import { pageVersionsApi, type VersionItem } from '@/api/platform/versions'
import { sortVersionsByDisplayOrder } from '../../shared/version-order'
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
  moduleId?: number
  renewMode: 'duration' | 'adjust'
  openDuration?: string
  customExpireEndTime?: Dayjs
}

const openModal = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

const formRef = ref<FormInstance>()
const loading = ref(false)
const submitting = ref(false)
const renewConfirmOpen = ref(false)
const detail = ref<InstitutionDetail>()
const records = ref<InstitutionRenewalRecord[]>([])
const versionLoading = ref(false)
const tenantVersions = ref<VersionItem[]>([])
const formState = reactive<RenewalFormState>({
  moduleId: undefined,
  renewMode: 'duration',
  openDuration: undefined,
  customExpireEndTime: undefined,
})
const renewActionLocks = reactive({
  openConfirm: false,
  submitRenewal: false,
})
const renewActionLockTimers: Partial<Record<'openConfirm' | 'submitRenewal', ReturnType<typeof setTimeout>>> = {}

const openTypeOptions = [
  { value: 1, label: '体验版' },
  { value: 2, label: '基础版' },
  { value: 3, label: '高级版' },
  { value: 4, label: '旗舰版' },
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
  3: [
    { value: '1y', label: '1年' },
    { value: '2y', label: '2年' },
    { value: '3y', label: '3年' },
    { value: '5y', label: '5年' },
    { value: '99y', label: '99年' },
  ],
  4: [
    { value: '1y', label: '1年' },
    { value: '2y', label: '2年' },
    { value: '3y', label: '3年' },
    { value: '5y', label: '5年' },
    { value: '99y', label: '99年' },
  ],
}

const availableOpenTypeOptions = computed(() => tenantVersions.value.map(item => ({ value: Number(item.id), label: item.name })))
const openDurationOptions = computed(() => openDurationOptionMap[2])
const confirmOpenTypeLabel = computed(() => getVersionName(formState.moduleId))
const confirmOpenDurationLabel = computed(() => formState.renewMode === 'adjust' ? getAdjustedDurationLabel(detail.value?.expireEndTime, previewExpireEndTime.value) : getOpenDurationLabel(2, formState.openDuration))
const previewExpireEndTime = computed(() => {
  if (formState.renewMode === 'adjust')
    return formState.customExpireEndTime ? formState.customExpireEndTime.format('YYYY-MM-DD HH:mm') : '--'

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
  {
    title: '操作人',
    dataIndex: 'operatorName',
    key: 'operatorName',
    width: 140,
    fixed: 'right' as const,
  },
]

const rules: Record<string, Rule[]> = {
  moduleId: [{ required: true, message: '请选择开通版本', trigger: 'change' }],
  openDuration: [
    {
      validator: async () => {
        if (formState.renewMode === 'duration' && !formState.openDuration)
          return Promise.reject(new Error('请选择续期时长'))
        return Promise.resolve()
      },
      trigger: 'change',
    },
  ],
  customExpireEndTime: [
    {
      validator: async () => {
        if (formState.renewMode === 'adjust' && !formState.customExpireEndTime)
          return Promise.reject(new Error('请选择自定义到期时间'))
        return Promise.resolve()
      },
      trigger: 'change',
    },
  ],
}

function closeModal() {
  emit('update:open', false)
}

function acquireRenewActionLock(key: 'openConfirm' | 'submitRenewal', delay = 600) {
  if (renewActionLocks[key])
    return false

  renewActionLocks[key] = true
  if (renewActionLockTimers[key])
    clearTimeout(renewActionLockTimers[key])

  renewActionLockTimers[key] = setTimeout(() => {
    renewActionLocks[key] = false
    renewActionLockTimers[key] = undefined
  }, delay)

  return true
}

function closeRenewConfirm() {
  renewConfirmOpen.value = false
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
  const currentValue = Number(value || 0)
  return openTypeOptions.find(item => item.value === currentValue)?.label || '基础版'
}

function getOpenDurationLabel(openType: number, value?: string) {
  const duration = String(value || '').trim()
  return openDurationOptionMap[openType]?.find(item => item.value === duration)?.label || '--'
}

function getAdjustedDurationLabel(beforeTime?: string, afterTime?: string) {
  const before = parseDateTime(beforeTime)
  const after = parseDateTime(afterTime)
  if (!before || !after)
    return '--'

  const diff = after.getTime() - before.getTime()
  if (diff === 0)
    return '0天'

  const dayMs = 24 * 60 * 60 * 1000
  const days = Math.ceil(Math.abs(diff) / dayMs)
  return `${diff > 0 ? '' : '-'}${days}天`
}

function getRenewDurationLabel(record: Partial<InstitutionRenewalRecord>) {
  const duration = String(record.renewDuration || '').trim()
  if (duration === 'adjust' || duration === '调整')
    return getAdjustedDurationLabel(record.beforeExpireEndTime, record.afterExpireEndTime)
  return getOpenDurationLabel(record.afterOpenType, record.renewDuration)
}

function getVersionName(moduleId?: number) {
  const id = Number(moduleId || 0)
  return tenantVersions.value.find(item => Number(item.id) === id)?.name || detail.value?.currentModuleName || '--'
}

function resolveDurationValue(openType: number, preferred?: string) {
  const options = openDurationOptionMap[openType] || []
  const matched = options.find(item => item.value === String(preferred || '').trim())
  return matched?.value || options[0]?.value
}

function resetState() {
  renewConfirmOpen.value = false
  detail.value = undefined
  records.value = []
  formState.moduleId = undefined
  formState.renewMode = 'duration'
  formState.openDuration = undefined
  formState.customExpireEndTime = undefined
  tenantVersions.value = []
  nextTick(() => {
    formRef.value?.clearValidate?.()
  })
}

function applyDefaultForm(detailData: InstitutionDetail) {
  const currentModuleId = Number(detailData.currentModuleId || 0)
  formState.moduleId = currentModuleId || tenantVersions.value[0]?.id
  formState.renewMode = 'duration'
  formState.openDuration = resolveDurationValue(2, detailData.openDuration)
  formState.customExpireEndTime = undefined
  nextTick(() => {
    formRef.value?.clearValidate?.()
  })
}

function handleOpenTypeChange(value?: number | string) {
  formState.moduleId = value ? Number(value) : undefined
  formState.openDuration = resolveDurationValue(2, formState.openDuration)
}

function handleRenewModeChange() {
  if (formState.renewMode === 'duration') {
    formState.openDuration = resolveDurationValue(2, formState.openDuration)
    formState.customExpireEndTime = undefined
  }
  else {
    formState.openDuration = undefined
    const currentExpireEnd = detail.value?.expireEndTime ? dayjs(detail.value.expireEndTime) : undefined
    formState.customExpireEndTime = currentExpireEnd?.isValid() ? currentExpireEnd : dayjs().add(1, 'day')
  }
  nextTick(() => formRef.value?.clearValidate?.())
}

async function loadData(id: number) {
  loading.value = true
  try {
    versionLoading.value = true
    const [detailRes, recordRes, versionRes] = await Promise.all([
      getInstitutionDetailApi({ id }),
      getInstitutionRenewalRecordsApi({ institutionId: id }),
      pageVersionsApi({ current: 1, size: 200, type: 1, institutionId: id }),
    ])

    if (detailRes.code !== 200 || !detailRes.result) {
      messageService.error(detailRes.message || '获取机构信息失败')
      return
    }
    if (recordRes.code !== 200) {
      messageService.error(recordRes.message || '获取续期记录失败')
      return
    }
    if (versionRes.code !== 200) {
      messageService.error(versionRes.message || '获取版本列表失败')
      return
    }

    detail.value = detailRes.result
    tenantVersions.value = sortVersionsByDisplayOrder<VersionItem>(Array.isArray(versionRes.result) ? versionRes.result : [])
    records.value = Array.isArray(recordRes.result) ? recordRes.result : []
    applyDefaultForm(detailRes.result)
  }
  catch (error: any) {
    console.error('load institution renewal data failed', error)
    messageService.error(error?.message || '获取续期信息失败')
  }
  finally {
    loading.value = false
    versionLoading.value = false
  }
}

function buildPayload(): InstitutionRenewalMutationPayload | null {
  const institutionId = Number(props.institutionId || 0)
  const moduleId = Number(formState.moduleId || 0)
  const openDuration = formState.renewMode === 'adjust' ? 'adjust' : String(formState.openDuration || '').trim()
  if (!institutionId || !moduleId || !openDuration)
    return null
  if (formState.renewMode === 'adjust' && !formState.customExpireEndTime)
    return null

  return {
    institutionId,
    moduleId,
    openDuration,
    customExpireEndTime: formState.renewMode === 'adjust'
      ? formState.customExpireEndTime?.format('YYYY-MM-DD HH:mm:ss')
      : undefined,
  }
}

async function openRenewConfirm() {
  if (loading.value || submitting.value || renewConfirmOpen.value)
    return
  if (!acquireRenewActionLock('openConfirm'))
    return

  try {
    await formRef.value?.validate()
  }
  catch {
    return
  }

  renewConfirmOpen.value = true
}

async function submitRenewal() {
  if (submitting.value)
    return
  if (!acquireRenewActionLock('submitRenewal'))
    return

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
    renewConfirmOpen.value = false
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

onBeforeUnmount(() => {
  Object.values(renewActionLockTimers).forEach((timer) => {
    if (timer)
      clearTimeout(timer)
  })
})

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
              <div class="renewal-fact renewal-fact--wide renewal-fact--field">
                <span class="renewal-fact__label renewal-fact__label--field">机构名称</span>
                <span class="renewal-fact__field" :title="detail?.organName || '--'">
                  {{ detail?.organName || '--' }}
                </span>
              </div>

              <div class="renewal-fact">
                <span class="renewal-fact__label">当前开通版本</span>
                <span class="renewal-fact__value">{{ detail?.currentModuleName || getOpenTypeLabel(detail?.openType) }}</span>
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
              <a-form-item label="开通版本" name="moduleId">
                <a-select
                  v-model:value="formState.moduleId"
                  :options="availableOpenTypeOptions"
                  :loading="versionLoading"
                  :disabled="versionLoading || availableOpenTypeOptions.length === 0"
                  class="renewal-inline-form__select"
                  :placeholder="availableOpenTypeOptions.length ? '请选择开通版本' : '暂无可续费版本'"
                  @change="handleOpenTypeChange"
                />
              </a-form-item>

              <a-form-item label="续期方式" name="renewMode">
                <a-segmented
                  v-model:value="formState.renewMode"
                  :options="[
                    { label: '按时长', value: 'duration' },
                    { label: '调整到期', value: 'adjust' },
                  ]"
                  @change="handleRenewModeChange"
                />
              </a-form-item>

              <a-form-item v-if="formState.renewMode === 'duration'" label="续期时长" name="openDuration">
                <a-select
                  v-model:value="formState.openDuration"
                  :options="openDurationOptions"
                  :disabled="!formState.moduleId"
                  class="renewal-inline-form__select"
                  placeholder="请选择续期时长"
                />
              </a-form-item>

              <a-form-item v-else label="到期时间" name="customExpireEndTime">
                <a-date-picker
                  v-model:value="formState.customExpireEndTime"
                  show-time
                  format="YYYY-MM-DD HH:mm"
                  class="renewal-inline-form__date"
                  placeholder="请选择到期时间"
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
            :scroll="{ x: 1020, y: 320 }"
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
                {{ getRenewDurationLabel(record) }}
              </template>

              <template v-else-if="column.key === 'beforeExpireEndTime'">
                {{ formatDateMinute(record.beforeExpireEndTime) }}
              </template>

              <template v-else-if="column.key === 'afterExpireEndTime'">
                {{ formatDateMinute(record.afterExpireEndTime) }}
              </template>

              <template v-else-if="column.key === 'operatorName'">
                <span class="operator-cell">
                  <span>{{ record.operatorName || '--' }}</span>
                  <span v-if="!record.isTenantOperator" class="assist-tag">代办</span>
                </span>
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
      <a-button type="primary" ghost :loading="submitting" @click="openRenewConfirm">
        确定续期
      </a-button>
    </template>
  </a-modal>

  <a-modal
    v-model:open="renewConfirmOpen"
    centered
    destroy-on-close
    :keyboard="false"
    :mask-closable="false"
    :confirm-loading="submitting"
    title="确认续期"
    ok-text="确认续期"
    cancel-text="取消"
    width="420px"
    @cancel="closeRenewConfirm"
    @ok="submitRenewal"
  >
    <div class="renew-confirm">
      <div class="renew-confirm__text">
        请确认本次续期信息，确认后将立即生效。
      </div>

      <div class="renew-confirm__panel">
        <div class="renew-confirm__item">
          <span class="renew-confirm__label">机构名称</span>
          <span class="renew-confirm__value" :title="detail?.organName || '--'">
            {{ detail?.organName || '--' }}
          </span>
        </div>

        <div class="renew-confirm__item">
          <span class="renew-confirm__label">开通版本</span>
          <span class="renew-confirm__value">{{ confirmOpenTypeLabel }}</span>
        </div>

        <div class="renew-confirm__item">
          <span class="renew-confirm__label">续期时长</span>
          <span class="renew-confirm__value">{{ confirmOpenDurationLabel }}</span>
        </div>

        <div class="renew-confirm__item">
          <span class="renew-confirm__label">续后到期时间</span>
          <span class="renew-confirm__value">{{ previewExpireEndTime }}</span>
        </div>
      </div>
    </div>
  </a-modal>
</template>

<style scoped lang="less">
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

.renewal-fact--field {
  gap: 12px;
  padding: 0;
  background: transparent;
  border: none;
}

.renewal-fact__label {
  flex-shrink: 0;
  color: #8c8c8c;
  font-size: 12px;
  line-height: 20px;
}

.renewal-fact__label--field {
  line-height: 32px;
}

.renewal-fact__value {
  color: #262626;
  font-size: 13px;
  font-weight: 600;
  line-height: 20px;
}

.renewal-fact__field {
  display: flex;
  align-items: center;
  flex: 1 1 auto;
  min-width: 0;
  height: 32px;
  padding: 0 11px;
  color: #262626;
  font-size: 13px;
  font-weight: 600;
  background: #fafbfc;
  border: 1px solid #d9d9d9;
  border-radius: 6px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.renewal-fact__value--ellipsis {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.renewal-inline-form {
  display: flex;
  flex-wrap: wrap;
  gap: 10px 12px;
  width: 100%;
}

.renewal-inline-form :deep(.ant-form-item) {
  margin-right: 0;
  margin-bottom: 0;
}

.renewal-inline-form__select {
  width: 152px;
}

.renewal-inline-form__date {
  width: 190px;
}

.renewal-inline-form__preview-item {
  flex: 0 0 auto;
  min-width: 0;
}

.renewal-inline-form__preview-item :deep(.ant-form-item-label) {
  flex: 0 0 auto;
  white-space: nowrap;
}

.renewal-inline-form__preview-item :deep(.ant-form-item-control) {
  flex: 0 0 auto;
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
  width: 190px;
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

.renew-confirm {
  padding-top: 4px;
}

.renew-confirm__text {
  margin-bottom: 14px;
  color: #595959;
  font-size: 13px;
  line-height: 22px;
}

.renew-confirm__panel {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 14px;
  background: #fafbfc;
  border: 1px solid #edf1f7;
  border-radius: 10px;
}

.renew-confirm__item {
  display: grid;
  grid-template-columns: 92px minmax(0, 1fr);
  gap: 10px;
  align-items: center;
}

.renew-confirm__label {
  color: #8c8c8c;
  font-size: 12px;
  line-height: 20px;
}

.renew-confirm__value {
  color: #262626;
  font-size: 13px;
  font-weight: 600;
  line-height: 20px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
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

.operator-cell {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  max-width: 100%;
  white-space: nowrap;
}

.assist-tag {
  display: inline-flex;
  align-items: center;
  height: 18px;
  padding: 0 6px;
  border: 1px solid #fed7aa;
  border-radius: 999px;
  background: #fff7ed;
  color: #ea580c;
  font-size: 12px;
  line-height: 18px;
}
</style>
