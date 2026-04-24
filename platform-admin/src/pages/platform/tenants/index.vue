<script setup lang="ts">
import type { TableColumnsType } from 'ant-design-vue'
import { Modal } from 'ant-design-vue'
import { ExclamationCircleOutlined, PlusOutlined } from '@ant-design/icons-vue'
import { createVNode, computed, reactive, ref } from 'vue'
import messageService from '@/utils/messageService'

interface TenantBrand {
  appName?: string
  logo?: string
  primaryColor?: string
}

interface TenantItem {
  tenantId: string
  name: string
  enabled: boolean
  brand: TenantBrand
  features: string[]
  institutionIds: number[]
  institutionNames: string[]
  remark?: string
  updateTime?: string
}

interface FeatureDefinition {
  code: string
  name: string
  group: string
  description: string
}

const featureDefinitions: FeatureDefinition[] = [
  {
    code: 'tenant-branding',
    name: '开通品牌定制',
    group: '品牌开通',
    description: '租户开通后，机构端才会使用这里配置的系统名称、Logo 和主题色',
  },
  {
    code: 'compose-course',
    name: '开通组合课程模块',
    group: '教务开通',
    description: '租户开通后，机构端才可能展示组合课程；员工能否操作仍看权限',
  },
]

const institutionOptions = [
  { id: 1, organName: 'A机构' },
  { id: 2, organName: 'B机构' },
  { id: 3, organName: 'C集团校区' },
]

const dataSource = ref<TenantItem[]>([
  {
    tenantId: 'tenant-a',
    name: 'A客户',
    enabled: true,
    brand: {
      appName: 'A客户教培后台',
      logo: '/logo.svg',
      primaryColor: '#1677FF',
    },
    features: ['tenant-branding', 'compose-course'],
    institutionIds: [1],
    institutionNames: ['A机构'],
    remark: '演示：A客户开通品牌定制和组合课程模块',
    updateTime: '2026-04-24 12:00:00',
  },
  {
    tenantId: 'tenant-b',
    name: 'B客户',
    enabled: true,
    brand: {
      appName: 'B客户招生后台',
      logo: '/logo.svg',
      primaryColor: '#13C2C2',
    },
    features: ['tenant-branding'],
    institutionIds: [2],
    institutionNames: ['B机构'],
    remark: '演示：B客户只开通品牌定制，不开通组合课程',
    updateTime: '2026-04-24 12:00:00',
  },
])

const saving = ref(false)
const keyword = ref('')
const modalOpen = ref(false)
const editingTenantId = ref('')

const pagination = reactive({
  current: 1,
  pageSize: 20,
  total: computed(() => filteredTenants.value.length),
  showSizeChanger: true,
  showTotal: (total: number) => `共 ${total} 个租户`,
})

const formState = reactive<TenantItem>({
  tenantId: '',
  name: '',
  enabled: true,
  brand: {
    appName: '',
    logo: '',
    primaryColor: '#1677FF',
  },
  features: [],
  institutionIds: [],
  institutionNames: [],
  remark: '',
})

const columns: TableColumnsType<TenantItem> = [
  { title: '租户信息', key: 'tenant', width: 260, fixed: 'left' as const },
  { title: '机构绑定', key: 'institutions', width: 260 },
  { title: '产品开通', key: 'features', width: 320 },
  { title: '状态', key: 'enabled', width: 100, align: 'center' as const },
  { title: '更新时间', dataIndex: 'updateTime', key: 'updateTime', width: 180 },
  { title: '操作', key: 'action', width: 150, fixed: 'right' as const },
]

const filteredTenants = computed(() => {
  const value = keyword.value.trim().toLowerCase()
  if (!value)
    return dataSource.value
  return dataSource.value.filter(item => [item.tenantId, item.name, item.remark].some(text => String(text || '').toLowerCase().includes(value)))
})

const featureNameMap = computed(() => new Map(featureDefinitions.map(item => [item.code, item.name])))
const selectedFeatureSet = computed(() => new Set(formState.features || []))
const featureGroups = computed(() => {
  const groups = new Map<string, FeatureDefinition[]>()
  featureDefinitions.forEach((item) => {
    if (!groups.has(item.group))
      groups.set(item.group, [])
    groups.get(item.group)!.push(item)
  })
  return Array.from(groups.entries()).map(([module, items]) => ({ module, items }))
})

function resetForm() {
  editingTenantId.value = ''
  formState.tenantId = ''
  formState.name = ''
  formState.enabled = true
  formState.brand = {
    appName: '',
    logo: '',
    primaryColor: '#1677FF',
  }
  formState.features = []
  formState.institutionIds = []
  formState.institutionNames = []
  formState.remark = ''
  formState.updateTime = ''
}

function fillForm(record: TenantItem) {
  editingTenantId.value = record.tenantId
  formState.tenantId = record.tenantId
  formState.name = record.name
  formState.enabled = record.enabled
  formState.brand = { ...record.brand }
  formState.features = [...record.features]
  formState.institutionIds = [...record.institutionIds]
  formState.institutionNames = [...record.institutionNames]
  formState.remark = record.remark || ''
  formState.updateTime = record.updateTime || ''
}

function formatFeatureLabel(code: string) {
  return featureNameMap.value.get(code) || code
}

function normalizeTenantId() {
  formState.tenantId = String(formState.tenantId || '').trim().toLowerCase().replace(/\s+/g, '-')
}

function syncInstitutionNames() {
  const nameMap = new Map(institutionOptions.map(item => [item.id, item.organName]))
  formState.institutionNames = formState.institutionIds.map(id => nameMap.get(id) || `机构${id}`)
}

function toggleFeature(code: string, checked: boolean) {
  const set = new Set(formState.features || [])
  if (checked)
    set.add(code)
  else
    set.delete(code)
  formState.features = Array.from(set)
}

function openCreateModal() {
  resetForm()
  modalOpen.value = true
}

function openEditModal(record: TenantItem | Record<string, any>) {
  fillForm(record as TenantItem)
  modalOpen.value = true
}

function handleSave() {
  normalizeTenantId()
  syncInstitutionNames()
  if (!formState.tenantId) {
    messageService.warning('请填写租户标识')
    return
  }
  if (!formState.name?.trim()) {
    messageService.warning('请填写租户名称')
    return
  }

  saving.value = true
  window.setTimeout(() => {
    const payload: TenantItem = {
      tenantId: formState.tenantId,
      name: formState.name,
      enabled: formState.enabled,
      brand: { ...formState.brand },
      features: [...formState.features],
      institutionIds: [...formState.institutionIds],
      institutionNames: [...formState.institutionNames],
      remark: formState.remark,
      updateTime: new Date().toLocaleString('zh-CN', { hour12: false }).replace(/\//g, '-'),
    }
    const index = dataSource.value.findIndex(item => item.tenantId === payload.tenantId)
    if (index >= 0)
      dataSource.value[index] = payload
    else
      dataSource.value.unshift(payload)
    saving.value = false
    modalOpen.value = false
    messageService.success('静态页面保存成功，仅当前页面临时生效')
  }, 300)
}

function handleDelete(record: TenantItem | Record<string, any>) {
  Modal.confirm({
    title: '删除租户配置',
    icon: createVNode(ExclamationCircleOutlined),
    content: `确认删除 ${record.name || record.tenantId}？当前是静态页面，仅会从本页列表移除。`,
    okText: '确认删除',
    okType: 'danger',
    cancelText: '取消',
    onOk() {
      dataSource.value = dataSource.value.filter(item => item.tenantId !== record.tenantId)
      messageService.success('已删除')
    },
  })
}

function handleSearch() {
  pagination.current = 1
}
</script>

<template>
  <div class="tenant-page">
    <div class="tenant-header">
      <div>
        <div class="tenant-header__title">
          租户管理
        </div>
        <div class="tenant-header__desc">
          静态示例页：用于确认 OEM 租户管理的页面结构，暂不接后端和数据库。
        </div>
      </div>
      <div class="tenant-header__actions">
        <a-input-search
          v-model:value="keyword"
          allow-clear
          placeholder="搜索租户名称 / 标识"
          class="tenant-header__search"
          @search="handleSearch"
        />
        <a-button type="primary" @click="openCreateModal">
          <template #icon>
            <PlusOutlined />
          </template>
          新建租户
        </a-button>
      </div>
    </div>

    <div class="tenant-overview">
      <div class="tenant-card tenant-card--blue">
        <span class="tenant-card__label">租户总数</span>
        <strong>{{ filteredTenants.length }}</strong>
      </div>
      <div class="tenant-card tenant-card--green">
        <span class="tenant-card__label">产品开通项</span>
        <strong>{{ featureDefinitions.length }}</strong>
      </div>
      <div class="tenant-card tenant-card--purple">
        <span class="tenant-card__label">当前状态</span>
        <strong>静态预览</strong>
      </div>
    </div>

    <a-alert
      show-icon
      type="info"
      class="tenant-static-alert"
      message="当前页面为静态预览"
      description="新增、编辑、删除只会在当前页面临时生效，刷新后恢复演示数据；后续确认方案后再接后端接口。"
    />

    <a-table
      row-key="tenantId"
      :columns="columns"
      :data-source="filteredTenants"
      :pagination="pagination"
      :scroll="{ x: 1080 }"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'tenant'">
          <div class="tenant-info">
            <div class="tenant-info__name">
              {{ record.name || '--' }}
            </div>
            <div class="tenant-info__code">
              {{ record.tenantId }}
            </div>
            <div v-if="record.remark" class="tenant-info__remark">
              {{ record.remark }}
            </div>
          </div>
        </template>

        <template v-if="column.key === 'institutions'">
          <div class="tenant-tags">
            <a-tag v-for="(name, index) in record.institutionNames" :key="record.institutionIds[index] || name">
              {{ name || `机构${record.institutionIds[index]}` }}
            </a-tag>
            <span v-if="!record.institutionIds?.length" class="tenant-empty">未绑定机构</span>
          </div>
        </template>

        <template v-if="column.key === 'features'">
          <div class="tenant-tags">
            <a-tag v-for="feature in record.features" :key="feature" color="blue">
              {{ formatFeatureLabel(feature) }}
            </a-tag>
            <span v-if="!record.features?.length" class="tenant-empty">未开通产品</span>
          </div>
        </template>

        <template v-if="column.key === 'enabled'">
          <a-tag :color="record.enabled ? 'success' : 'default'">
            {{ record.enabled ? '启用' : '停用' }}
          </a-tag>
        </template>

        <template v-if="column.key === 'action'">
          <a-space>
            <a-button type="link" size="small" @click="openEditModal(record)">
              编辑
            </a-button>
            <a-button type="link" danger size="small" @click="handleDelete(record)">
              删除
            </a-button>
          </a-space>
        </template>
      </template>
    </a-table>

    <a-modal
      v-model:open="modalOpen"
      centered
      :width="920"
      :confirm-loading="saving"
      :body-style="{ maxHeight: '72vh', overflowY: 'auto', padding: '0' }"
      wrap-class-name="tenant-config-modal"
      ok-text="保存配置"
      cancel-text="取消"
      @ok="handleSave"
    >
      <template #title>
        <div class="tenant-modal-title">
          <div class="tenant-modal-title__main">
            {{ editingTenantId ? '编辑租户' : '新建租户' }}
          </div>
          <div class="tenant-modal-title__desc">
            静态配置示例：一个租户对应一套品牌配置和一组机构。
          </div>
        </div>
      </template>

      <a-form layout="vertical" class="tenant-form tenant-form--modal">
        <section class="tenant-config-section">
          <div class="tenant-config-section__head">
            <div>
              <div class="tenant-config-section__title">
                基础信息
              </div>
              <div class="tenant-config-section__desc">
                用来识别这个 OEM 客户，租户标识保存后不可修改。
              </div>
            </div>
            <a-switch v-model:checked="formState.enabled" checked-children="启用" un-checked-children="停用" />
          </div>
          <div class="tenant-form__grid">
            <a-form-item label="租户标识" required>
              <a-input
                v-model:value="formState.tenantId"
                :disabled="Boolean(editingTenantId)"
                placeholder="例如 tenant-a"
                @blur="normalizeTenantId"
              />
            </a-form-item>
            <a-form-item label="租户名称" required>
              <a-input v-model:value="formState.name" placeholder="例如 A客户 / 某某集团" />
            </a-form-item>
          </div>
        </section>

        <section class="tenant-config-section">
          <div class="tenant-config-section__head">
            <div>
              <div class="tenant-config-section__title">
                品牌定制
              </div>
              <div class="tenant-config-section__desc">
                勾选“开通品牌定制”后，机构端才会使用下面的系统名称、Logo 和主题色。
              </div>
            </div>
          </div>
          <div class="brand-config-layout">
            <div class="brand-config-form">
              <a-form-item label="系统名称">
                <a-input v-model:value="formState.brand.appName" placeholder="例如 A客户教培后台" />
              </a-form-item>
              <a-form-item label="品牌 Logo">
                <a-input v-model:value="formState.brand.logo" placeholder="图片地址，例如 /logo.svg 或 https://..." />
              </a-form-item>
              <a-form-item label="主题色">
                <div class="brand-color-field">
                  <a-input v-model:value="formState.brand.primaryColor" type="color" class="brand-color-field__picker" />
                  <a-input v-model:value="formState.brand.primaryColor" placeholder="#1677FF" />
                </div>
              </a-form-item>
            </div>
            <div class="brand-preview-card" :style="{ '--tenant-preview-color': formState.brand.primaryColor || '#1677FF' }">
              <div class="brand-preview-card__label">
                机构端预览
              </div>
              <div class="brand-preview-card__shell">
                <div class="brand-preview-card__side">
                  <div class="brand-preview-card__logo">
                    <img v-if="formState.brand.logo" :src="formState.brand.logo" alt="logo">
                    <span v-else>{{ (formState.brand.appName || formState.name || '租').slice(0, 1) }}</span>
                  </div>
                  <div class="brand-preview-card__name">
                    {{ formState.brand.appName || formState.name || '机构后台' }}
                  </div>
                </div>
                <div class="brand-preview-card__content">
                  <span />
                  <strong />
                  <em />
                </div>
              </div>
            </div>
          </div>
        </section>

        <section class="tenant-config-section">
          <div class="tenant-config-section__title">
            机构绑定
          </div>
          <div class="tenant-config-section__desc">
            静态示例：这些机构后续会作为租户归属关系保存。
          </div>
          <a-form-item class="tenant-form-item--last">
            <a-select
              v-model:value="formState.institutionIds"
              mode="multiple"
              show-search
              option-filter-prop="label"
              placeholder="选择哪些机构属于这个租户"
              :options="institutionOptions.map(item => ({ value: item.id, label: `${item.organName}（ID:${item.id}）` }))"
            />
          </a-form-item>
        </section>

        <section class="tenant-config-section">
          <div class="tenant-config-section__title">
            产品开通（不是权限）
          </div>
          <div class="tenant-config-section__desc">
            这里表达客户有没有开通这个产品模块；权限仍然在“版本管理/角色权限”里配置。
          </div>
          <div class="feature-groups feature-groups--modal">
            <div v-for="group in featureGroups" :key="group.module" class="feature-group">
              <div class="feature-group__title">
                {{ group.module }}
              </div>
              <div class="feature-group__items">
                <a-checkbox
                  v-for="feature in group.items"
                  :key="feature.code"
                  :checked="selectedFeatureSet.has(feature.code)"
                  @change="event => toggleFeature(feature.code, event.target.checked)"
                >
                  <span class="feature-option">
                    <span>{{ feature.name }}</span>
                    <small>{{ feature.description }}</small>
                  </span>
                </a-checkbox>
              </div>
            </div>
          </div>
        </section>

        <section class="tenant-config-section">
          <div class="tenant-config-section__title">
            备注
          </div>
          <a-form-item class="tenant-form-item--last">
            <a-textarea v-model:value="formState.remark" :rows="3" placeholder="记录客户品牌定制说明、交付约定等" />
          </a-form-item>
        </section>
      </a-form>
    </a-modal>
  </div>
</template>

<style scoped lang="less">
.tenant-page {
  padding: 24px;
}

.tenant-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 16px;

  &__title {
    color: rgba(0, 0, 0, 0.88);
    font-size: 22px;
    font-weight: 600;
    line-height: 32px;
  }

  &__desc {
    margin-top: 4px;
    color: rgba(0, 0, 0, 0.45);
  }

  &__actions {
    display: flex;
    gap: 12px;
  }

  &__search {
    width: 280px;
  }
}

.tenant-overview {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 16px;
  margin-bottom: 16px;
}

.tenant-card {
  padding: 18px 20px;
  border: 1px solid rgba(0, 0, 0, 0.06);
  border-radius: 14px;
  background: #fff;

  &__label {
    display: block;
    margin-bottom: 8px;
    color: rgba(0, 0, 0, 0.45);
  }

  strong {
    color: rgba(0, 0, 0, 0.88);
    font-size: 24px;
  }

  &--blue {
    background: linear-gradient(135deg, #eef5ff, #fff);
  }

  &--green {
    background: linear-gradient(135deg, #effff7, #fff);
  }

  &--purple {
    background: linear-gradient(135deg, #f7f1ff, #fff);
  }
}

.tenant-static-alert {
  margin-bottom: 16px;
}

.tenant-info {
  &__name {
    color: rgba(0, 0, 0, 0.88);
    font-weight: 600;
  }

  &__code,
  &__remark {
    margin-top: 4px;
    color: rgba(0, 0, 0, 0.45);
    font-size: 12px;
  }
}

.tenant-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.tenant-empty {
  color: rgba(0, 0, 0, 0.35);
}

.tenant-form {
  &__grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 0 16px;
  }
}

.feature-groups {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.feature-group {
  padding: 12px;
  border: 1px solid rgba(0, 0, 0, 0.06);
  border-radius: 10px;
  background: rgba(0, 0, 0, 0.015);

  &__title {
    margin-bottom: 8px;
    color: rgba(0, 0, 0, 0.65);
    font-weight: 600;
  }

  &__items {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
}

.feature-option {
  display: inline-flex;
  flex-direction: column;
  gap: 2px;

  small {
    color: rgba(0, 0, 0, 0.35);
  }
}

.tenant-form--modal {
  padding: 20px 24px 8px;
}

.tenant-modal-title {
  &__main {
    color: rgba(0, 0, 0, 0.88);
    font-size: 18px;
    font-weight: 600;
    line-height: 26px;
  }

  &__desc {
    margin-top: 4px;
    color: rgba(0, 0, 0, 0.45);
    font-size: 13px;
    font-weight: 400;
  }
}

.tenant-config-section {
  padding: 18px 18px 2px;
  border: 1px solid rgba(0, 0, 0, 0.06);
  border-radius: 12px;
  background: #fff;

  & + & {
    margin-top: 14px;
  }

  &__head {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 16px;
    margin-bottom: 16px;
  }

  &__title {
    color: rgba(0, 0, 0, 0.88);
    font-weight: 600;
    line-height: 22px;
  }

  &__desc {
    margin: 4px 0 16px;
    color: rgba(0, 0, 0, 0.45);
    font-size: 13px;
    line-height: 20px;
  }
}

.brand-config-layout {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 280px;
  gap: 18px;
}

.brand-config-form {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0 16px;
}

.brand-color-field {
  display: flex;
  gap: 8px;

  &__picker {
    width: 56px;
    flex: 0 0 56px;
    padding: 4px 6px;
  }
}

.brand-preview-card {
  --tenant-preview-color: #1677ff;

  &__label {
    margin-bottom: 8px;
    color: rgba(0, 0, 0, 0.45);
    font-size: 13px;
  }

  &__shell {
    height: 148px;
    overflow: hidden;
    border: 1px solid rgba(0, 0, 0, 0.06);
    border-radius: 12px;
    background: #f5f7fb;
    display: grid;
    grid-template-columns: 96px 1fr;
  }

  &__side {
    padding: 14px 12px;
    background: var(--tenant-preview-color);
    color: #fff;
  }

  &__logo {
    width: 34px;
    height: 34px;
    display: grid;
    place-items: center;
    overflow: hidden;
    border-radius: 8px;
    background: rgba(255, 255, 255, 0.9);
    color: var(--tenant-preview-color);
    font-weight: 700;

    img {
      width: 100%;
      height: 100%;
      object-fit: contain;
    }
  }

  &__name {
    margin-top: 10px;
    font-size: 12px;
    font-weight: 600;
    line-height: 18px;
  }

  &__content {
    padding: 18px;

    span,
    strong,
    em {
      display: block;
      border-radius: 999px;
      background: rgba(0, 0, 0, 0.08);
    }

    span {
      width: 88px;
      height: 10px;
    }

    strong {
      width: 130px;
      height: 20px;
      margin-top: 18px;
      background: color-mix(in srgb, var(--tenant-preview-color) 18%, white);
    }

    em {
      width: 100%;
      height: 46px;
      margin-top: 16px;
    }
  }
}

.feature-groups--modal {
  margin-bottom: 16px;
}

.tenant-form-item--last {
  margin-bottom: 16px;
}
</style>
