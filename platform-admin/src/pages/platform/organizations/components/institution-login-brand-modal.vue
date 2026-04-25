<script setup lang="ts">
import type { InstitutionDetail, InstitutionMutationPayload, TenantLoginBrandConfig } from '@/api/platform/institutions'
import { CopyOutlined } from '@ant-design/icons-vue'
import { computed, reactive, ref, watch } from 'vue'
import { getInstitutionDetailApi, updateInstitutionApi } from '@/api/platform/institutions'
import { listTenantsApi } from '@/api/platform/tenants'
import messageService from '@/utils/messageService'

const props = defineProps<{
  open: boolean
  institutionId?: number | null
}>()

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'saved'): void
}>()

const openModal = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

const loading = ref(false)
const saving = ref(false)
const institutionDetail = ref<InstitutionDetail | null>(null)
const institutionDomain = ref('')
const formState = reactive({
  loginSlug: '',
  template: '',
  primaryColor: '#1677ff',
  loginTitle: '',
  heroTitle: '',
  heroDescription: '',
})

const templateOptions = [
  { label: '跟随租户默认', value: '' },
  { label: '轻量门户登录', value: 'clean-portal' },
  { label: '校区品牌卡片', value: 'campus-card' },
  { label: '教务分屏登录', value: 'education-split' },
]

const colorOptions = [
  { label: '科技蓝', value: '#1677ff' },
  { label: '活力橙', value: '#fe8130' },
  { label: '教育绿', value: '#13ad74' },
  { label: '品牌紫', value: '#7c3aed' },
  { label: '商务青', value: '#08979c' },
  { label: '高级黑', value: '#1f2937' },
]

const institutionName = computed(() => institutionDetail.value?.organName || '机构')
const logoUrl = computed(() => institutionDetail.value?.logo || '')
const loginDomainSuffix = computed(() => institutionDomain.value ? `.${institutionDomain.value}` : '.租户机构端域名')
const fullLoginDomain = computed(() => {
  const slug = normalizeSlug(formState.loginSlug)
  return slug && institutionDomain.value ? `${slug}.${institutionDomain.value}` : ''
})

function normalizeSlug(value: string) {
  return String(value || '')
    .trim()
    .toLowerCase()
    .replace(/\s+/g, '-')
    .replace(/[^a-z0-9-_]/g, '-')
    .replace(/_+/g, '-')
    .replace(/-+/g, '-')
    .replace(/^-|-$/g, '')
}

function resetForm() {
  institutionDetail.value = null
  institutionDomain.value = ''
  formState.loginSlug = ''
  formState.template = ''
  formState.primaryColor = '#1677ff'
  formState.loginTitle = ''
  formState.heroTitle = ''
  formState.heroDescription = ''
}

function applyDetail(detail: InstitutionDetail) {
  institutionDetail.value = detail
  const brand = detail.profile?.loginBrand || {}
  formState.loginSlug = String(detail.profile?.loginSlug || '')
  formState.template = String(brand.template || '')
  formState.primaryColor = String(brand.primaryColor || '#1677ff')
  formState.loginTitle = String(brand.loginTitle || '')
  formState.heroTitle = String(brand.heroTitle || '')
  formState.heroDescription = String(brand.heroDescription || '')
}

async function loadTenantInstitutionDomain(institutionId: number) {
  try {
    const res = await listTenantsApi()
    const tenants = res.result || res.data || []
    const matchedTenant = tenants.find((tenant: any) => Array.isArray(tenant.institutionIds) && tenant.institutionIds.map(Number).includes(institutionId))
      || tenants.find((tenant: any) => Array.isArray(tenant.institutionDomains) && tenant.institutionDomains.length)
    institutionDomain.value = String(matchedTenant?.institutionDomains?.[0] || '').trim()
  }
  catch (error) {
    console.warn('load tenant institution domain failed', error)
    institutionDomain.value = ''
  }
}

async function loadDetail() {
  const id = Number(props.institutionId || 0)
  if (!id)
    return

  loading.value = true
  try {
    const res = await getInstitutionDetailApi({ id })
    if (res.code !== 200 || !res.result) {
      messageService.error(res.message || '获取机构详情失败')
      return
    }
    applyDetail(res.result)
    await loadTenantInstitutionDomain(id)
  }
  catch (error: any) {
    console.error('load institution login brand detail failed', error)
    messageService.error(error?.message || '获取机构详情失败')
  }
  finally {
    loading.value = false
  }
}

function buildPayload(detail: InstitutionDetail): InstitutionMutationPayload & { id: number } {
  const brand: TenantLoginBrandConfig = {
    template: formState.template.trim() || undefined,
    brandName: detail.organName,
    logoUrl: detail.logo || undefined,
    loginTitle: formState.loginTitle.trim() || undefined,
    primaryColor: formState.primaryColor.trim() || undefined,
    heroTitle: formState.heroTitle.trim() || undefined,
    heroDescription: formState.heroDescription.trim() || undefined,
  }

  return {
    id: detail.id,
    organName: detail.organName,
    loginName: detail.loginName,
    mobile: detail.mobile,
    principal: detail.principal || '',
    provinceCode: detail.provinceCode,
    province: detail.province,
    cityCode: detail.cityCode,
    city: detail.city,
    regionCode: detail.regionCode,
    region: detail.region,
    address: detail.address,
    lng: detail.lng,
    lat: detail.lat,
    concatPhone: detail.concatPhone,
    fixedPhone: detail.fixedPhone,
    remark: detail.remark,
    logo: detail.logo,
    enabled: detail.enabled,
    profile: {
      businessTime: detail.profile?.businessTime || undefined,
      description: detail.profile?.description || undefined,
      video: detail.profile?.video || undefined,
      galleryImages: detail.profile?.galleryImages?.length ? detail.profile.galleryImages : undefined,
      loginSlug: normalizeSlug(formState.loginSlug) || undefined,
      loginBrand: Object.values(brand).some(Boolean) ? brand : undefined,
    },
  }
}

async function copyFullLoginDomain() {
  if (!fullLoginDomain.value) {
    messageService.warning('请先填写一级子域名')
    return
  }
  try {
    await navigator.clipboard.writeText(fullLoginDomain.value)
    messageService.success('访问域名已复制')
  }
  catch (error) {
    console.warn('copy login domain failed', error)
    messageService.error('复制失败，请手动复制')
  }
}

function validateLoginBrandForm() {
  const hasIndependentDomain = !!normalizeSlug(formState.loginSlug)
  if (!hasIndependentDomain)
    return true

  if (!institutionDomain.value) {
    messageService.warning('请先在租户管理中配置机构端登录域名')
    return false
  }
  if (!formState.loginTitle.trim()) {
    messageService.warning('请填写登录标题')
    return false
  }
  if (!formState.heroTitle.trim()) {
    messageService.warning('请填写宣传标题')
    return false
  }
  if (!formState.primaryColor.trim()) {
    messageService.warning('请选择主色调')
    return false
  }
  if (!formState.heroDescription.trim()) {
    messageService.warning('请填写宣传文案')
    return false
  }
  return true
}

async function handleSave() {
  const detail = institutionDetail.value
  if (!detail)
    return

  formState.loginSlug = normalizeSlug(formState.loginSlug)
  if (!validateLoginBrandForm())
    return

  saving.value = true
  try {
    const res = await updateInstitutionApi(buildPayload(detail))
    if (res.code !== 200) {
      messageService.error(res.message || '保存机构独立登录页失败')
      return
    }
    messageService.success('机构独立登录页已保存')
    emit('saved')
    openModal.value = false
  }
  catch (error: any) {
    console.error('save institution login brand failed', error)
    messageService.error(error?.message || '保存机构独立登录页失败')
  }
  finally {
    saving.value = false
  }
}

watch(
  () => props.open,
  (open) => {
    if (!open) {
      resetForm()
      return
    }
    resetForm()
    void loadDetail()
  },
)
</script>

<template>
  <a-modal
    v-model:open="openModal"
    :width="760"
    centered
    destroy-on-close
    ok-text="保存配置"
    cancel-text="取消"
    :confirm-loading="saving"
    @ok="handleSave"
  >
    <template #title>
      <div class="login-brand-title">
        <strong>机构独立登录页</strong>
        <span>{{ institutionName }}</span>
      </div>
    </template>

    <a-spin :spinning="loading">
      <div class="login-brand-form">
        <a-alert
          type="info"
          show-icon
          message="独立域名选填；不填写时继续使用租户机构端登录地址。填写独立域名后，标题、主色和文案需要完整配置；页面模板可选择跟随租户默认。"
        />

        <div class="login-brand-logo-row">
          <a-avatar v-if="logoUrl" :src="logoUrl" :size="52" />
          <a-avatar v-else :size="52">{{ institutionName.slice(0, 1) }}</a-avatar>
          <div>
            <strong>{{ institutionName }}</strong>
            <span>登录页默认复用机构 Logo，如需更换请在机构编辑里修改 Logo。</span>
          </div>
        </div>

        <a-row :gutter="16">
          <a-col :xs="24" :md="12">
            <a-form-item label="独立域名" class="login-domain-form-item">
              <div class="login-domain-input-row">
                <a-input v-model:value="formState.loginSlug" placeholder="例如kaina" @blur="formState.loginSlug = normalizeSlug(formState.loginSlug)" />
                <span class="login-domain-suffix">{{ loginDomainSuffix }}</span>
              </div>
            </a-form-item>
            <div class="login-domain-preview">
              <span>访问地址：</span>
              <template v-if="fullLoginDomain">
                <strong>{{ fullLoginDomain }}</strong>
                <button type="button" class="login-domain-copy" title="复制访问域名" @click="copyFullLoginDomain">
                  <CopyOutlined />
                </button>
              </template>
              <em v-else>未配置独立机构域名，使用租户机构端登录地址</em>
            </div>
          </a-col>
          <a-col :xs="24" :md="12">
            <a-form-item label="页面模板">
              <a-select v-model:value="formState.template" :options="templateOptions" />
            </a-form-item>
          </a-col>
          <a-col :xs="24" :md="12">
            <a-form-item label="登录标题" :required="!!formState.loginSlug">
              <a-input v-model:value="formState.loginTitle" placeholder="默认使用机构名称" />
            </a-form-item>
          </a-col>
          <a-col :xs="24" :md="12">
            <a-form-item label="宣传标题" :required="!!formState.loginSlug">
              <a-input v-model:value="formState.heroTitle" placeholder="例如：欢迎进入本校区" />
            </a-form-item>
          </a-col>
          <a-col :xs="24">
            <a-form-item label="主色调" :required="!!formState.loginSlug">
              <div class="login-brand-colors">
                <button
                  v-for="color in colorOptions"
                  :key="color.value"
                  type="button"
                  class="login-brand-color"
                  :class="{ 'login-brand-color--active': formState.primaryColor === color.value }"
                  :style="{ '--brand-color': color.value }"
                  @click="formState.primaryColor = color.value"
                >
                  <span />{{ color.label }}
                </button>
              </div>
            </a-form-item>
          </a-col>
          <a-col :xs="24">
            <a-form-item label="宣传文案" :required="!!formState.loginSlug">
              <a-input v-model:value="formState.heroDescription" placeholder="机构登录页展示文案" />
            </a-form-item>
          </a-col>
        </a-row>
      </div>
    </a-spin>
  </a-modal>
</template>

<style scoped lang="less">
.login-brand-title {
  display: flex;
  align-items: baseline;
  gap: 10px;

  strong {
    font-size: 16px;
  }

  span {
    color: rgba(0, 0, 0, 0.45);
    font-size: 13px;
  }
}

.login-brand-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.login-brand-logo-row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 14px;
  background: #fafafa;
  border: 1px solid #eef0f4;
  border-radius: 12px;

  div {
    display: flex;
    flex-direction: column;
    gap: 3px;
  }

  strong {
    color: rgba(0, 0, 0, 0.88);
  }

  span {
    color: rgba(0, 0, 0, 0.45);
    font-size: 13px;
  }
}

.login-domain-form-item {
  position: relative;

  :deep(.ant-form-item-row) {
    flex-flow: row nowrap;
    align-items: center;
  }

  :deep(.ant-form-item-label) {
    flex: 0 0 auto;
    white-space: nowrap;
  }

  :deep(.ant-form-item-control) {
    flex: 1 1 auto;
    min-width: 0;
  }
}

.login-domain-input-row {
  display: flex;
  align-items: center;
  width: 100%;

  :deep(.ant-input) {
    flex: 0 0 96px;
    border-top-right-radius: 0;
    border-bottom-right-radius: 0;
  }
}

.login-domain-suffix {
  flex: 0 1 190px;
  min-width: 0;
  height: 32px;
  padding: 4px 12px;
  overflow: hidden;
  color: rgba(0, 0, 0, 0.88);
  line-height: 22px;
  text-overflow: ellipsis;
  white-space: nowrap;
  background: #fafafa;
  border: 1px solid #d9d9d9;
  border-radius: 0 6px 6px 0;
}

.login-domain-preview {
  position: absolute;
  top: 34px;
  left: 80px;
  z-index: 1;
  max-width: 360px;
  margin: 0;
  overflow: hidden;
  color: rgba(0, 0, 0, 0.45);
  font-size: 12px;
  line-height: 18px;
  text-overflow: ellipsis;
  white-space: nowrap;

  strong {
    color: var(--pro-ant-color-primary, #1677ff);
    font-weight: 500;
  }

  :deep(.anticon) {
    font-size: 13px;
  }

  em {
    color: rgba(0, 0, 0, 0.4);
    font-style: normal;
  }
}

.login-domain-copy {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  margin-left: 6px;
  padding: 0;
  color: var(--pro-ant-color-primary, #1677ff);
  vertical-align: -3px;
  background: transparent;
  border: 0;
  border-radius: 4px;
  cursor: pointer;

  &:hover {
    background: rgba(22, 119, 255, 0.08);
  }
}

.login-brand-colors {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.login-brand-color {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 30px;
  padding: 0 11px;
  color: rgba(0, 0, 0, 0.65);
  background: #fff;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;

  span {
    width: 13px;
    height: 13px;
    background: var(--brand-color);
    border-radius: 50%;
  }

  &:hover,
  &--active {
    color: var(--brand-color);
    border-color: var(--brand-color);
  }

  &--active {
    background: color-mix(in srgb, var(--brand-color) 8%, #fff);
    box-shadow: 0 0 0 2px color-mix(in srgb, var(--brand-color) 12%, transparent);
  }
}
</style>
