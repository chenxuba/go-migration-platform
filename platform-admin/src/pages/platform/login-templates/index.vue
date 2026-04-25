<script setup lang="ts">
import type { TableColumnsType } from 'ant-design-vue'
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { PlusOutlined, ReloadOutlined } from '@ant-design/icons-vue'
import { deleteLoginTemplateApi, listLoginTemplatesApi, saveLoginTemplateApi, type LoginTemplateItem } from '@/api/platform/login-templates'
import { useUserStore } from '@/stores/user'
import { pageInstitutionsApi } from '@/api/platform/institutions'
import { listTenantsApi, type TenantListItem } from '@/api/platform/tenants'
import messageService from '@/utils/messageService'
import { buildRealLoginTemplatePreviewUrl } from '../shared/login-template-real-preview'

interface InstitutionOption {
  id: number
  organName: string
  tenantName?: string
}

const userStore = useUserStore()
const isPlatformAdmin = computed(() => userStore.userInfo?.tenantRole === 'platform_admin')
const currentTenantId = computed(() => String(userStore.userInfo?.tenantId || ''))
const loading = ref(false)
const saving = ref(false)
const modalOpen = ref(false)
const rows = ref<LoginTemplateItem[]>([])
const tenants = ref<TenantListItem[]>([])
const institutions = ref<InstitutionOption[]>([])
const entryFilter = ref<string | undefined>()
const formState = reactive({
  id: undefined as number | undefined,
  templateName: '',
  templateKey: '',
  entryType: 'institution-admin',
  layoutType: 'split',
  description: '',
  previewImage: '',
  enabled: true,
  sort: 10,
  tenantIds: [] as string[],
  institutionIds: [] as number[],
})

const columns: TableColumnsType = [
  { title: '模板信息', key: 'template', width: 200 },
  { title: '适用端口', key: 'entryType', width: 130 },
  { title: '可用范围', key: 'scope', width: 220 },
  { title: '引用数量', key: 'referenceCount', width: 110 },
  { title: '状态', key: 'enabled', width: 100 },
  { title: '排序', dataIndex: 'sort', width: 80 },
  { title: '更新时间', dataIndex: 'updateTime', width: 160 },
  { title: '操作', key: 'action', width: 160, fixed: 'right' },
]

const filteredRows = computed(() => entryFilter.value ? rows.value.filter(item => item.entryType === entryFilter.value || item.entryType === 'all') : rows.value)
const tenantOptions = computed(() => tenants.value.filter(item => item.tenantId !== 'platform').map(item => ({ label: item.tenantName, value: item.tenantId })))
const institutionOptions = computed(() => institutions.value.map(item => ({ label: item.tenantName ? `${item.organName}（${item.tenantName}）` : item.organName, value: item.id })))
const isPlatformTemplate = computed(() => formState.entryType === 'platform-admin')
const isInstitutionTemplate = computed(() => formState.entryType === 'institution-admin')

function asTemplate(record: Record<string, any>) {
  return record as LoginTemplateItem
}

watch(() => formState.entryType, (entryType) => {
  if (entryType === 'platform-admin')
    formState.institutionIds = []
  if (entryType === 'institution-admin')
    formState.tenantIds = []
})

function entryTypeText(value?: string) {
  if (value === 'platform-admin')
    return '子总控后台'
  if (value === 'institution-admin')
    return '机构端'
  return value || '机构端'
}

function layoutText(value?: string) {
  const map: Record<string, string> = { split: '分屏', card: '卡片', portal: '门户' }
  return map[value || ''] || value || '分屏'
}

function scopeText(record: LoginTemplateItem) {
  const tenantCount = record.tenantIds?.length || 0
  const institutionCount = record.institutionIds?.length || 0
  if (!tenantCount && !institutionCount)
    return record.entryType === 'platform-admin' ? '全租户通用' : '全租户 / 全机构通用'
  const parts = []
  if (tenantCount)
    parts.push(`${tenantCount} 个租户`)
  if (institutionCount)
    parts.push(`${institutionCount} 个机构`)
  return parts.join('，')
}

function resetForm() {
  formState.id = undefined
  formState.templateName = ''
  formState.templateKey = ''
  formState.entryType = 'institution-admin'
  formState.layoutType = 'split'
  formState.description = ''
  formState.previewImage = ''
  formState.enabled = true
  formState.sort = 10
  formState.tenantIds = []
  formState.institutionIds = []
}

function openCreate() {
  resetForm()
  modalOpen.value = true
}

function openEdit(record: LoginTemplateItem) {
  formState.id = record.id
  formState.templateName = record.templateName
  formState.templateKey = record.templateKey
  formState.entryType = record.entryType === 'platform-admin' ? 'platform-admin' : 'institution-admin'
  formState.layoutType = String(record.layoutType || 'split')
  formState.description = record.description || ''
  formState.previewImage = record.previewImage || ''
  formState.enabled = !!record.enabled
  formState.sort = record.sort || 0
  formState.tenantIds = [...(record.tenantIds || [])]
  formState.institutionIds = [...(record.institutionIds || [])]
  modalOpen.value = true
}

function openPreview(record: LoginTemplateItem) {
  const scope = record.entryType === 'institution-admin' ? 'institution' : 'platform'
  window.open(buildRealLoginTemplatePreviewUrl({
    scope,
    template: record.templateKey,
    name: record.templateName,
    desc: record.description || '',
    layout: record.layoutType || 'split',
  }), '_blank', 'noopener,noreferrer')
}

async function loadRows() {
  loading.value = true
  try {
    const res = await listLoginTemplatesApi(isPlatformAdmin.value ? {} : { tenantId: currentTenantId.value, enabledOnly: true })
    rows.value = res.result || []
  }
  catch (error) {
    console.error(error)
    messageService.error('获取登录页模板失败')
  }
  finally {
    loading.value = false
  }
}

async function loadOptions() {
  if (!isPlatformAdmin.value)
    return
  try {
    const [tenantRes, institutionRes] = await Promise.all([
      listTenantsApi({}),
      pageInstitutionsApi({ current: 1, size: 500 }),
    ])
    tenants.value = tenantRes.result || []
    const payload = institutionRes.data as any
    institutions.value = payload?.items || institutionRes.result || []
  }
  catch (error) {
    console.warn('load template scope options failed', error)
  }
}

async function handleSave() {
  if (!formState.templateName.trim()) {
    messageService.warning('请填写模板名称')
    return
  }
  if (!formState.templateKey.trim()) {
    messageService.warning('请填写模板编码')
    return
  }
  if (!['platform-admin', 'institution-admin'].includes(formState.entryType)) {
    messageService.warning('请选择适用端口')
    return
  }
  saving.value = true
  try {
    await saveLoginTemplateApi({
      id: formState.id,
      templateName: formState.templateName.trim(),
      templateKey: formState.templateKey.trim(),
      entryType: formState.entryType,
      layoutType: formState.layoutType,
      description: formState.description.trim(),
      previewImage: formState.previewImage.trim(),
      enabled: formState.enabled,
      sort: Number(formState.sort || 0),
      tenantIds: formState.tenantIds,
      institutionIds: formState.institutionIds,
    })
    messageService.success('模板已保存')
    modalOpen.value = false
    await loadRows()
  }
  catch (error: any) {
    messageService.error(error?.response?.data?.message || error?.message || '保存模板失败')
  }
  finally {
    saving.value = false
  }
}

async function handleDelete(record: LoginTemplateItem) {
  await loadRows()
  const latest = rows.value.find(item => item.id === record.id)
  const referenceCount = latest?.referenceCount ?? record.referenceCount ?? 0
  if (referenceCount > 0) {
    messageService.warning(`当前模板已有 ${referenceCount} 个引用，请先调整引用后再删除`)
    return
  }
  try {
    await deleteLoginTemplateApi({ id: record.id })
    messageService.success('模板已删除')
    await loadRows()
  }
  catch (error: any) {
    messageService.error(error?.response?.data?.message || error?.message || '删除模板失败')
  }
}

onMounted(async () => {
  await Promise.all([loadRows(), loadOptions()])
})
</script>

<template>
  <div class="login-template-page">
    <div class="login-template-page__header">
      <div>
        <h1>登录页模板</h1>
        <p>管理模板元信息、启停状态和适用范围；模板编码需要对应前端已开发的登录组件。</p>
      </div>
      <a-space class="login-template-page__tools">
        <a-select v-model:value="entryFilter" allow-clear placeholder="全部端口" class="template-filter-select" style="width: 200px">
          <a-select-option value="platform-admin">子总控后台</a-select-option>
          <a-select-option value="institution-admin">机构端</a-select-option>
        </a-select>
        <a-button @click="loadRows"><template #icon><ReloadOutlined /></template>刷新</a-button>
        <a-button v-if="isPlatformAdmin" type="primary" @click="openCreate"><template #icon><PlusOutlined /></template>新增模板</a-button>
      </a-space>
    </div>

    <a-alert
      class="login-template-page__tip"
      type="info"
      show-icon
      :message="isPlatformAdmin ? '新增模板时，模板编码必须和前端组件注册 key 一致；不选择租户/机构时表示全租户、全机构通用。' : '当前页面仅展示本租户可用的登录页模板，租户账号仅支持真实预览，不支持新增、编辑和删除。'"
    />

    <a-table
      :columns="columns"
      :data-source="filteredRows"
      :loading="loading"
      row-key="id"
      :pagination="false"
      :scroll="{ x: 1280 }"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'template'">
          <div class="template-cell">
            <strong>{{ record.templateName }}</strong>
            <div class="template-cell__meta">
              <span>{{ record.templateKey }}</span>
              <i>{{ layoutText(record.layoutType) }}布局</i>
            </div>
          </div>
        </template>
        <template v-else-if="column.key === 'entryType'">
          <span class="template-tag template-tag--entry">{{ entryTypeText(record.entryType) }}</span>
        </template>
        <template v-else-if="column.key === 'scope'">
          <div class="scope-cell">
            <strong>{{ scopeText(asTemplate(record)) }}</strong>
          </div>
        </template>
        <template v-else-if="column.key === 'referenceCount'">
          <span class="reference-count" :class="{ 'reference-count--active': Number(record.referenceCount || 0) > 0 }">{{ record.referenceCount || 0 }}</span>
        </template>
        <template v-else-if="column.key === 'enabled'">
          <span class="template-tag" :class="record.enabled ? 'template-tag--enabled' : 'template-tag--disabled'">{{ record.enabled ? '启用' : '停用' }}</span>
        </template>
        <template v-else-if="column.key === 'action'">
          <div class="template-actions">
            <a-button type="link" size="small" class="template-action" @click="openPreview(asTemplate(record))">真实预览</a-button>
            <template v-if="isPlatformAdmin">
              <a-button type="link" size="small" class="template-action" @click="openEdit(asTemplate(record))">编辑</a-button>
              <a-popconfirm title="删除前会实时校验引用数量，确认继续？" @confirm="handleDelete(asTemplate(record))">
                <a-button type="link" danger size="small" class="template-action template-action--danger" :disabled="Number(record.referenceCount || 0) > 0">删除</a-button>
              </a-popconfirm>
            </template>
          </div>
        </template>
      </template>
    </a-table>

    <a-modal v-if="isPlatformAdmin" v-model:open="modalOpen" :width="760" centered :confirm-loading="saving" ok-text="保存模板" cancel-text="取消" @ok="handleSave">
      <template #title>
        {{ formState.id ? '编辑登录页模板' : '新增登录页模板' }}
      </template>
      <a-form layout="vertical" class="template-form">
        <div class="template-form__notice">
          模板编码需与前端组件注册 key 保持一致；子总控模板按租户限制，机构端模板按机构限制，不选择表示全局通用。
        </div>
        <div class="template-form__grid">
          <a-form-item label="模板名称" required><a-input v-model:value="formState.templateName" placeholder="例如：康复中心品牌登录" /></a-form-item>
          <a-form-item label="模板编码" required><a-input v-model:value="formState.templateKey" placeholder="例如：rehab-center-login" /></a-form-item>
          <a-form-item label="适用端口">
            <a-select v-model:value="formState.entryType">
              <a-select-option value="platform-admin">子总控后台</a-select-option>
                  <a-select-option value="institution-admin">机构端</a-select-option>
            </a-select>
          </a-form-item>
          <a-form-item label="布局类型">
            <a-select v-model:value="formState.layoutType">
              <a-select-option value="split">分屏</a-select-option>
              <a-select-option value="card">卡片</a-select-option>
              <a-select-option value="portal">门户</a-select-option>
            </a-select>
          </a-form-item>
          <a-form-item label="排序"><a-input-number v-model:value="formState.sort" :min="0" style="width: 100%" /></a-form-item>
          <a-form-item label="状态"><a-switch v-model:checked="formState.enabled" checked-children="启用" un-checked-children="停用" /></a-form-item>
        </div>
        <a-form-item label="模板说明"><a-textarea v-model:value="formState.description" :rows="3" placeholder="说明模板适用场景，便于租户配置时选择" /></a-form-item>
        <a-form-item v-if="isPlatformTemplate" label="适用租户">
          <a-select v-model:value="formState.tenantIds" mode="multiple" allow-clear show-search option-filter-prop="label" :options="tenantOptions" placeholder="不选择表示全租户通用" />
        </a-form-item>
        <a-form-item v-if="isInstitutionTemplate" label="适用机构">
          <a-select v-model:value="formState.institutionIds" mode="multiple" allow-clear show-search option-filter-prop="label" :options="institutionOptions" placeholder="不选择表示全机构通用" />
        </a-form-item>
      </a-form>
    </a-modal>

  </div>
</template>

<style scoped lang="less">
.login-template-page {
  padding: 18px;
}

.login-template-page__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 12px;

  h1 { margin: 0; color: #111827; font-size: 22px; line-height: 30px; }
  p { margin: 4px 0 0; color: #667085; font-size: 14px; }
}

.login-template-page__tip { margin-bottom: 12px; }

.login-template-page__tools {
  :deep(.ant-btn) {
    height: 32px;
    font-size: 14px;
  }
}

.template-filter-select {
  :deep(.ant-select-selector) {
    height: 32px !important;
    border-radius: 8px !important;
  }

  :deep(.ant-select-selection-item),
  :deep(.ant-select-selection-placeholder) {
    font-size: 14px;
    line-height: 30px !important;
  }

  :deep(.ant-select-arrow),
  :deep(.ant-select-clear) {
    top: 50%;
    right: 10px;
    width: 14px;
    height: 14px;
    margin-top: -7px;
    color: #98a2b3;
    font-size: 12px;
    line-height: 14px;
  }

  :deep(.ant-select-clear) {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    border-radius: 50%;
    background: #fff;
  }
}

:deep(.ant-table-tbody > tr > td) {
  color: #344054;
}

:deep(.ant-table-cell-fix-right) {
  background: #fff;
}

.template-actions {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  white-space: nowrap;
}

.template-action {
  height: 24px;
  padding: 0;
  font-size: 14px;
  line-height: 24px;
}

.template-action--danger {
  color: #ff4d4f;
}

.template-cell {
  display: flex;
  min-width: 0;
  flex-direction: column;
  justify-content: center;

  strong {
    overflow: hidden;
    color: #111827;
    font-size: 14px;
    font-weight: 600;
    line-height: 20px;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.template-cell__meta {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
  margin-top: 4px;

  span, i {
    overflow: hidden;
    color: #667085;
    font-size: 12px;
    font-style: normal;
    line-height: 18px;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  i {
    position: relative;
    flex: none;
    color: #98a2b3;
  }

  i::before {
    display: inline-block;
    width: 3px;
    height: 3px;
    margin: 0 8px 2px 0;
    border-radius: 50%;
    background: #d0d5dd;
    content: '';
  }
}

.template-tag {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  height: 24px;
  padding: 0 10px;
  border: 1px solid transparent;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 500;
  line-height: 22px;
  white-space: nowrap;
}

.template-tag--entry {
  border-color: #bfdbfe;
  background: #eff6ff;
  color: #175cd3;
}

.template-tag--enabled {
  border-color: #bbf7d0;
  background: #f0fdf4;
  color: #15803d;
}

.template-tag--disabled {
  border-color: #e5e7eb;
  background: #f9fafb;
  color: #667085;
}

.reference-count {
  color: #667085;
  font-size: 14px;
  font-weight: 500;
}

.reference-count--active {
  color: #175cd3;
}

.scope-cell {
  display: flex;
  flex-direction: column;
  gap: 3px;
  min-width: 0;

  strong { color: #344054; font-size: 14px; }
  span { overflow: hidden; color: #667085; font-size: 12px; text-overflow: ellipsis; white-space: nowrap; }
}

.template-form__grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  column-gap: 14px;
}

.template-form__notice {
  margin: 0 0 14px;
  padding: 9px 12px;
  border: 1px solid #e5edff;
  border-radius: 8px;
  background: #f8fbff;
  color: #667085;
  font-size: 13px;
  line-height: 20px;
}
</style>
