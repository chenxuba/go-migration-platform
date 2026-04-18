<script setup lang="ts">
import type { TableColumnsType } from 'ant-design-vue'
import { computed, ref, watch } from 'vue'
import type {
  InstitutionPermissionDetail,
  InstitutionVersionChangeRecord,
} from '@/api/platform/institutions'
import type { VersionItem } from '@/api/platform/versions'
import PlatformModalShell from '../../shared/platform-modal-shell.vue'
import {
  getInstitutionPermissionDetailApi,
  getInstitutionVersionChangeRecordsApi,
  replaceInstitutionPermissionVersionApi,
} from '@/api/platform/institutions'
import { pageVersionsApi } from '@/api/platform/versions'
import { filterSystemDefaultVersions, sortVersionsByDisplayOrder } from '../../shared/version-order'
import messageService from '@/utils/messageService'

const props = defineProps<{
  open: boolean
  institutionId?: number | null
}>()

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'saved'): void
}>()

const openTypeLabelMap: Record<number, string> = {
  1: '体验版',
  2: '基础版',
  3: '高级版',
  4: '旗舰版',
}

const standardVersionOpenTypeMap: Record<string, number> = {
  体验版: 1,
  基础版: 2,
  高级版: 3,
  旗舰版: 4,
}

const statusLabelMap: Record<number, string> = {
  1: '启用',
  2: '停用',
  4: '过期',
}

const openModal = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

const loading = ref(false)
const submitting = ref(false)
const detail = ref<InstitutionPermissionDetail | null>(null)
const versionOptions = ref<VersionItem[]>([])
const versionChangeRecords = ref<InstitutionVersionChangeRecord[]>([])
const selectedModuleId = ref<number | undefined>()

const columns: TableColumnsType<InstitutionVersionChangeRecord> = [
  {
    title: '切换时间',
    dataIndex: 'createTime',
    key: 'createTime',
    width: 180,
  },
  {
    title: '切换前版本',
    dataIndex: 'beforeVersionName',
    key: 'beforeVersionName',
    width: 140,
  },
  {
    title: '切换后版本',
    dataIndex: 'afterVersionName',
    key: 'afterVersionName',
    width: 140,
  },
  {
    title: '操作人',
    dataIndex: 'operatorName',
    key: 'operatorName',
    width: 140,
  },
]

const currentVersionName = computed(() => {
  const moduleName = String(detail.value?.currentModuleName || '').trim()
  if (moduleName)
    return moduleName
  return getOpenTypeLabel(detail.value?.openType)
})

const versionSelectOptions = computed(() => {
  const currentOpenType = Number(detail.value?.openType || 0)
  return filterSystemDefaultVersions(versionOptions.value)
    .filter((item) => {
      const mappedOpenType = getVersionOpenTypeByName(item.name)
      if (!mappedOpenType || !currentOpenType)
        return true
      if (currentOpenType === 1)
        return mappedOpenType === 1
      return mappedOpenType >= 2
    })
    .map(item => ({ value: item.id, label: item.name }))
})

const changeHint = computed(() => {
  const currentOpenType = Number(detail.value?.openType || 0)
  if (currentOpenType === 1)
    return '体验版如需升级正式版本，请继续走续期流程。'
  return '切换版本后会立即生效，并自动记录切换日志，不会改动当前到期时间。'
})

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
  return openTypeLabelMap[Number(value || 0)] || '--'
}

function getStatusLabel(value?: number) {
  return statusLabelMap[Number(value || 0)] || '--'
}

function getStatusClass(value?: number) {
  const normalized = Number(value || 0)
  if (normalized === 1)
    return 'status-chip--enabled'
  if (normalized === 4)
    return 'status-chip--expired'
  return 'status-chip--disabled'
}

function getVersionOpenTypeByName(name?: string) {
  return standardVersionOpenTypeMap[String(name || '').trim()] || 0
}

function getVersionLabelByRecord(record: Partial<InstitutionVersionChangeRecord>, key: 'before' | 'after') {
  const versionName = key === 'before' ? record.beforeVersionName : record.afterVersionName
  const openType = key === 'before' ? record.beforeOpenType : record.afterOpenType
  const normalizedName = String(versionName || '').trim()
  return normalizedName || getOpenTypeLabel(openType)
}

function resolveSelectedModuleId(detailData: InstitutionPermissionDetail, versions: VersionItem[]) {
  const systemVersions = filterSystemDefaultVersions(versions)
  const currentModuleId = Number(detailData.currentModuleId || 0)
  if (currentModuleId && systemVersions.some(item => Number(item.id) === currentModuleId))
    return currentModuleId

  const currentOpenType = Number(detailData.openType || 0)
  return systemVersions.find(item => getVersionOpenTypeByName(item.name) === currentOpenType)?.id
}

async function loadData(institutionId: number) {
  loading.value = true
  try {
    const [detailRes, versionRes, recordRes] = await Promise.all([
      getInstitutionPermissionDetailApi({ institutionId }),
      pageVersionsApi({
        current: 1,
        size: 200,
        type: 1,
      }),
      getInstitutionVersionChangeRecordsApi({ institutionId }),
    ])

    if (detailRes.code !== 200 || !detailRes.result) {
      messageService.error(detailRes.message || '获取机构版本信息失败')
      return
    }
    if (versionRes.code !== 200) {
      messageService.error(versionRes.message || '获取版本列表失败')
      return
    }
    if (recordRes.code !== 200) {
      messageService.error(recordRes.message || '获取版本切换日志失败')
      return
    }

    const versions = sortVersionsByDisplayOrder(Array.isArray(versionRes.result) ? versionRes.result : [])
    detail.value = detailRes.result
    versionOptions.value = versions
    versionChangeRecords.value = Array.isArray(recordRes.result) ? recordRes.result : []
    selectedModuleId.value = resolveSelectedModuleId(detailRes.result, versions)
  }
  catch (error: any) {
    console.error('load institution version change data failed', error)
    messageService.error(error?.message || '获取机构版本信息失败')
  }
  finally {
    loading.value = false
  }
}

async function submitVersionChange() {
  const institutionId = Number(props.institutionId || 0)
  const moduleId = Number(selectedModuleId.value || 0)
  if (!institutionId || !moduleId)
    return

  if (Number(detail.value?.currentModuleId || 0) === moduleId) {
    messageService.warning('当前已是该版本，无需重复切换')
    return
  }

  submitting.value = true
  try {
    const res = await replaceInstitutionPermissionVersionApi({
      institutionId,
      moduleId,
    })
    if (res.code !== 200) {
      messageService.error(res.message || '机构版本切换失败')
      return
    }

    messageService.success('机构版本切换成功')
    emit('saved')
    await loadData(institutionId)
  }
  catch (error: any) {
    console.error('change institution version failed', error)
    messageService.error(error?.message || '机构版本切换失败')
  }
  finally {
    submitting.value = false
  }
}

watch(
  () => [props.open, props.institutionId] as const,
  ([open, institutionId]) => {
    if (!open) {
      detail.value = null
      versionOptions.value = []
      versionChangeRecords.value = []
      selectedModuleId.value = undefined
      return
    }

    if (institutionId)
      void loadData(Number(institutionId))
  },
  { immediate: true },
)
</script>

<template>
  <PlatformModalShell
    v-model:open="openModal"
    :width="980"
    title="切换版本"
    modal-class="institution-permission-modal"
  >
    <a-spin :spinning="loading">
      <div class="version-switch-modal">
        <div class="version-switch-overview">
          <div class="version-switch-overview__main">
            <div class="version-switch-overview__name">
              {{ detail?.organName || '--' }}
            </div>
            <div class="version-switch-overview__meta">
              <span>登录账号：{{ detail?.mobile || '--' }}</span>
              <span>当前版本：{{ currentVersionName }}</span>
              <span>过期时间：{{ formatDateMinute(detail?.expireEndTime) }}</span>
            </div>
          </div>

          <span class="status-chip" :class="getStatusClass(detail?.status)">
            {{ getStatusLabel(detail?.status) }}
          </span>
        </div>

        <div class="version-switch-panel">
          <div class="switch-inline-field">
            <div class="switch-inline-field__label">
              当前版本
            </div>
            <div class="switch-inline-field__value" :title="currentVersionName">
              {{ currentVersionName }}
            </div>
          </div>

          <div class="switch-inline-field">
            <div class="switch-inline-field__label">
              当前到期时间
            </div>
            <div class="switch-inline-field__value">
              {{ formatDateMinute(detail?.expireEndTime) }}
            </div>
          </div>

          <div class="switch-inline-field switch-inline-field--select">
            <div class="switch-inline-field__label">
              目标版本
            </div>
            <a-select
              v-model:value="selectedModuleId"
              class="switch-inline-field__control"
              placeholder="请选择目标版本"
              :options="versionSelectOptions"
            />
          </div>

          <div class="switch-inline-note">
            <div class="switch-inline-note__label">
              说明
            </div>
            <div class="switch-inline-note__text">
              {{ changeHint }}
            </div>
          </div>
        </div>

        <div class="version-switch-log">
          <div class="version-switch-log__title">
            切换日志
          </div>
          <a-table
            :columns="columns"
            :data-source="versionChangeRecords"
            :pagination="false"
            :scroll="{ x: 600 }"
            row-key="id"
            size="small"
            class="version-switch-table"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'createTime'">
                {{ formatDateMinute(record.createTime) }}
              </template>

              <template v-else-if="column.key === 'beforeVersionName'">
                {{ getVersionLabelByRecord(record, 'before') }}
              </template>

              <template v-else-if="column.key === 'afterVersionName'">
                {{ getVersionLabelByRecord(record, 'after') }}
              </template>

              <template v-else-if="column.key === 'operatorName'">
                {{ record.operatorName || '--' }}
              </template>
            </template>
          </a-table>
        </div>
      </div>
    </a-spin>

    <template #footer>
      <a-button @click="closeModal">
        关闭
      </a-button>
      <a-button type="primary" :loading="submitting" @click="submitVersionChange">
        确认切换
      </a-button>
    </template>
  </PlatformModalShell>
</template>

<style scoped lang="less">
.version-switch-modal {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding-top: 4px;
}

.version-switch-overview {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 16px 18px;
  border: 1px solid #e8edf5;
  border-radius: 18px;
  background: #fff;
}

.version-switch-overview__main {
  min-width: 0;
}

.version-switch-overview__name {
  color: #262626;
  font-size: 16px;
  font-weight: 600;
  line-height: 24px;
}

.version-switch-overview__meta {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px 12px;
  margin-top: 6px;
  color: #595959;
  font-size: 12px;
  line-height: 18px;
}

.status-chip {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 7px 14px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 600;
  line-height: 18px;
}

.status-chip--enabled {
  background: rgba(22, 163, 74, 0.12);
  color: #15803d;
}

.status-chip--disabled {
  background: rgba(71, 85, 105, 0.12);
  color: #475569;
}

.status-chip--expired {
  background: rgba(234, 88, 12, 0.12);
  color: #c2410c;
}

.version-switch-panel {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1.15fr) minmax(260px, 1.2fr);
  gap: 10px 14px;
  padding: 16px 18px;
  border: 1px solid #e8edf5;
  border-radius: 18px;
  background: #fff;
}

.switch-inline-field {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
}

.switch-inline-field--select {
  min-width: 0;
}

.switch-inline-field__label {
  flex-shrink: 0;
  color: #8c8c8c;
  font-size: 12px;
  line-height: 18px;
}

.switch-inline-field__value {
  display: flex;
  align-items: center;
  flex: 1 1 auto;
  min-width: 0;
  min-height: 36px;
  padding: 7px 12px;
  color: #262626;
  font-size: 13px;
  line-height: 20px;
  background: #fafafa;
  border: 1px solid #d9d9d9;
  border-radius: 8px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.switch-inline-field__control {
  width: 100%;
  min-width: 0;
}

.switch-inline-note {
  grid-column: 1 / -1;
  display: flex;
  align-items: flex-start;
  gap: 10px;
  min-width: 0;
  padding-top: 2px;
}

.switch-inline-note__label {
  flex-shrink: 0;
  color: #8c8c8c;
  font-size: 12px;
  line-height: 20px;
}

.switch-inline-note__text {
  min-width: 0;
  color: #595959;
  font-size: 12px;
  line-height: 20px;
}

.version-switch-log {
  padding: 16px 18px 6px;
  border: 1px solid #e8edf5;
  border-radius: 18px;
  background: #fff;
}

.version-switch-log__title {
  margin-bottom: 10px;
  color: #262626;
  font-size: 14px;
  font-weight: 600;
  line-height: 22px;
}

:deep(.switch-inline-field__control .ant-select-selector) {
  min-height: 36px !important;
  padding-top: 1px !important;
  padding-bottom: 1px !important;
  border-radius: 8px !important;
}

:deep(.version-switch-table .ant-table-thead > tr > th) {
  background: #fafafa !important;
  color: #262626;
  font-size: 13px;
  font-weight: 500;
}

:deep(.version-switch-table .ant-table-tbody > tr > td) {
  color: #262626;
  font-size: 13px;
  line-height: 20px;
}

@media (max-width: 960px) {
  .version-switch-overview {
    align-items: flex-start;
  }

  .version-switch-panel {
    grid-template-columns: 1fr;
  }

  .switch-inline-field {
    flex-direction: column;
    align-items: stretch;
    gap: 6px;
  }
}
</style>
