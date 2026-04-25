<script setup lang="ts">
import type { AxiosError } from 'axios'
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
import { sortVersionsByDisplayOrder } from '../../shared/version-order'
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

const statusLabelMap: Record<number, string> = {
  1: '启用',
  2: '停用',
  3: '即将到期',
  4: '已过期',
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

const versionSelectOptions = computed(() => versionOptions.value.map(item => ({ value: item.id, label: item.name })))

const targetVersionName = computed(() => {
  const targetId = Number(selectedModuleId.value || 0)
  return versionOptions.value.find(item => Number(item.id) === targetId)?.name || '--'
})

const changeHint = computed(() => '切换后会按目标版本的默认权限模板重新生效；如需做机构级裁剪，请再进入“机构权限”调整。')

function closeModal() {
  emit('update:open', false)
}

function resetState() {
  detail.value = null
  versionOptions.value = []
  versionChangeRecords.value = []
  selectedModuleId.value = undefined
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
  if (normalized === 3)
    return 'status-chip--warning'
  if (normalized === 4)
    return 'status-chip--expired'
  return 'status-chip--disabled'
}

function getVersionBadgeClass(name?: string) {
  const normalized = String(name || '').trim()
  if (normalized.includes('旗舰'))
    return 'version-badge--flagship'
  if (normalized.includes('高级'))
    return 'version-badge--advanced'
  if (normalized.includes('基础'))
    return 'version-badge--basic'
  if (normalized.includes('体验'))
    return 'version-badge--trial'
  return 'version-badge--basic'
}

function getVersionLabelByRecord(record: Partial<InstitutionVersionChangeRecord>, key: 'before' | 'after') {
  const versionName = key === 'before' ? record.beforeVersionName : record.afterVersionName
  const openType = key === 'before' ? record.beforeOpenType : record.afterOpenType
  const normalizedName = String(versionName || '').trim()
  return normalizedName || getOpenTypeLabel(openType)
}

function resolveRequestErrorMessage(error: unknown, fallback: string) {
  const axiosError = error as AxiosError<{ message?: string }>
  return axiosError?.response?.data?.message || (error as any)?.message || fallback
}

function resolveSelectedModuleId(detailData: InstitutionPermissionDetail, versions: VersionItem[]) {
  const currentModuleId = Number(detailData.currentModuleId || 0)
  if (currentModuleId && versions.some(item => Number(item.id) === currentModuleId))
    return currentModuleId

  return versions[0]?.id
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
        institutionId,
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

    const versions = sortVersionsByDisplayOrder<VersionItem>(Array.isArray(versionRes.result) ? versionRes.result : [])
    detail.value = detailRes.result
    versionOptions.value = versions
    versionChangeRecords.value = Array.isArray(recordRes.result) ? recordRes.result : []
    selectedModuleId.value = resolveSelectedModuleId(detailRes.result, versions)
  }
  catch (error: any) {
    console.error('load institution version change data failed', error)
    messageService.error(resolveRequestErrorMessage(error, '获取机构版本信息失败'))
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
    messageService.error(resolveRequestErrorMessage(error, '机构版本切换失败'))
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
  <PlatformModalShell
    v-model:open="openModal"
    :width="1080"
    title="切换版本"
    modal-class="institution-version-modal"
    scrollable
  >
    <a-spin :spinning="loading">
      <div class="version-modal">
        <div class="version-overview">
          <div class="version-overview__main">
            <div class="version-overview__name">
              {{ detail?.organName || '--' }}
            </div>
            <div class="version-overview__meta">
              <span>登录账号：{{ detail?.mobile || '--' }}</span>
              <span>当前版本：{{ currentVersionName }}</span>
              <span>过期时间：{{ formatDateMinute(detail?.expireEndTime) }}</span>
            </div>
          </div>

          <span class="status-chip" :class="getStatusClass(detail?.status)">
            {{ getStatusLabel(detail?.status) }}
          </span>
        </div>

        <div class="version-main">
          <div class="version-info-card">
            <div class="version-info-card__label">
              当前版本
            </div>
            <div class="version-info-card__value version-badge" :class="getVersionBadgeClass(currentVersionName)">
              {{ currentVersionName }}
            </div>
            <div class="version-info-card__grid">
              <div class="version-info-card__item">
                <span class="version-info-card__item-label">账号状态</span>
                <span class="version-info-card__item-value">{{ getStatusLabel(detail?.status) }}</span>
              </div>
              <div class="version-info-card__item">
                <span class="version-info-card__item-label">到期时间</span>
                <span class="version-info-card__item-value">{{ formatDateMinute(detail?.expireEndTime) }}</span>
              </div>
            </div>
          </div>

          <div class="version-switch-card">
            <div class="version-switch-card__title">
              目标版本
            </div>
            <div class="version-switch-card__desc">
              {{ changeHint }}
            </div>

            <a-select
              v-model:value="selectedModuleId"
              class="version-switch-card__select"
              placeholder="请选择目标版本"
              :options="versionSelectOptions"
            />

            <div class="version-flow">
              <div class="version-flow__item">
                <span class="version-flow__label">当前</span>
                <span class="version-flow__value">{{ currentVersionName }}</span>
              </div>
              <div class="version-flow__arrow">
                →
              </div>
              <div class="version-flow__item version-flow__item--target">
                <span class="version-flow__label">切换后</span>
                <span class="version-flow__value">{{ targetVersionName }}</span>
              </div>
            </div>

            <div class="version-switch-card__tip">
              保存后会立即更新机构版本，并使用目标版本的默认权限模板。
            </div>
          </div>
        </div>

        <div class="version-log-card">
          <div class="version-log-card__title">
            切换日志
          </div>
          <a-table
            :columns="columns"
            :data-source="versionChangeRecords"
            :pagination="false"
            :scroll="{ x: 600 }"
            row-key="id"
            size="small"
            class="version-log-table"
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
.version-modal {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding-top: 4px;
}

.version-overview {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 16px 18px;
  border: 1px solid #e8edf5;
  border-radius: 18px;
  background: #fff;
}

.version-overview__main {
  min-width: 0;
}

.version-overview__name {
  color: #262626;
  font-size: 16px;
  font-weight: 600;
  line-height: 24px;
}

.version-overview__meta {
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

.status-chip--warning {
  background: rgba(245, 158, 11, 0.14);
  color: #b45309;
}

.status-chip--expired {
  background: rgba(234, 88, 12, 0.12);
  color: #c2410c;
}

.version-main {
  display: grid;
  grid-template-columns: 300px minmax(0, 1fr);
  gap: 12px;
}

.version-info-card,
.version-switch-card,
.version-log-card {
  border: 1px solid #e8edf5;
  border-radius: 18px;
  background: #fff;
}

.version-info-card {
  padding: 18px;
  background: linear-gradient(180deg, #fafcff 0%, #ffffff 100%);
}

.version-info-card__label {
  color: #8c8c8c;
  font-size: 12px;
  line-height: 18px;
}

.version-info-card__value {
  margin-top: 6px;
  color: #1f2329;
  font-size: 22px;
  font-weight: 700;
  line-height: 30px;
}

.version-badge {
  position: relative;
  display: inline-flex;
  align-items: center;
  gap: 8px;
  width: fit-content;
  min-height: 36px;
  padding: 4px 12px;
  border: 1px solid transparent;
  border-radius: 12px;
  font-size: 18px;
  font-weight: 700;
  letter-spacing: 0.5px;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.72);
}

.version-badge::before {
  content: '';
  width: 6px;
  height: 6px;
  border-radius: 50%;
  flex-shrink: 0;
}

.version-badge--trial {
  color: #475569;
  background: linear-gradient(135deg, #f8fafc 0%, #eef2f7 100%);
  border-color: #dbe3ee;
}

.version-badge--trial::before {
  background: #94a3b8;
}

.version-badge--basic {
  color: #245bdb;
  background: linear-gradient(135deg, #eff5ff 0%, #e0ecff 100%);
  border-color: #cdddff;
}

.version-badge--basic::before {
  background: #3b82f6;
}

.version-badge--advanced {
  color: #1d4ed8;
  background: linear-gradient(135deg, #eef4ff 0%, #dce9ff 100%);
  border-color: #c8dbff;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.84),
    0 6px 16px rgba(59, 130, 246, 0.08);
}

.version-badge--advanced::before {
  background: linear-gradient(180deg, #60a5fa 0%, #2563eb 100%);
}

.version-badge--flagship {
  color: #7a4b00;
  background: linear-gradient(135deg, #fff7e7 0%, #ffe7b3 48%, #fff2d3 100%);
  border-color: #f3d28b;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.9),
    0 6px 18px rgba(214, 158, 46, 0.12);
}

.version-badge--flagship::before {
  background: linear-gradient(180deg, #fbbf24 0%, #d97706 100%);
  box-shadow: 0 0 0 3px rgba(251, 191, 36, 0.14);
}

.version-info-card__grid {
  display: grid;
  gap: 10px;
  margin-top: 16px;
}

.version-info-card__item {
  padding: 10px 12px;
  border: 1px solid #eef2f7;
  border-radius: 12px;
  background: rgba(250, 250, 250, 0.9);
}

.version-info-card__item-label {
  display: block;
  color: #8c8c8c;
  font-size: 12px;
  line-height: 18px;
}

.version-info-card__item-value {
  display: block;
  margin-top: 4px;
  color: #262626;
  font-size: 13px;
  font-weight: 600;
  line-height: 20px;
}

.version-switch-card {
  padding: 18px 20px;
}

.version-switch-card__title {
  color: #262626;
  font-size: 15px;
  font-weight: 600;
  line-height: 22px;
}

.version-switch-card__desc {
  margin-top: 6px;
  color: #5b6475;
  font-size: 12px;
  line-height: 20px;
}

.version-switch-card__select {
  width: 100%;
  margin-top: 18px;
}

.version-flow {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 18px;
  padding: 14px 16px;
  border-radius: 14px;
  background: #f8fbff;
  border: 1px solid #e6efff;
}

.version-flow__item {
  min-width: 0;
  flex: 1;
}

.version-flow__label {
  display: block;
  color: #8c8c8c;
  font-size: 12px;
  line-height: 18px;
}

.version-flow__value {
  display: block;
  margin-top: 4px;
  color: #1f2329;
  font-size: 14px;
  font-weight: 600;
  line-height: 22px;
}

.version-flow__arrow {
  flex-shrink: 0;
  color: #8c8c8c;
  font-size: 18px;
  line-height: 1;
}

.version-switch-card__tip {
  margin-top: 16px;
  color: #8c8c8c;
  font-size: 12px;
  line-height: 20px;
}

.version-log-card {
  padding: 16px 18px 8px;
}

.version-log-card__title {
  color: #262626;
  font-size: 14px;
  font-weight: 600;
  line-height: 22px;
  margin-bottom: 12px;
}

.version-log-card :deep(.ant-table-thead > tr > th) {
  background: #fafafa !important;
  font-weight: 500;
}

.version-log-card :deep(.ant-table-tbody > tr > td) {
  padding-top: 12px;
  padding-bottom: 12px;
}

@media (max-width: 1200px) {
  .version-main {
    grid-template-columns: 1fr;
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
