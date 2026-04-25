<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { CloudUploadOutlined, SafetyCertificateOutlined } from '@ant-design/icons-vue'
import { getTenantStorageConfigApi, saveTenantStorageConfigApi, type TenantStorageConfig } from '@/api/platform/storage'
import { listTenantsApi, type TenantListItem } from '@/api/platform/tenants'
import { useUserStore } from '@/stores/user'
import messageService from '@/utils/messageService'
import { PlatformAccessEnum } from '~@/constants/access'

const userStore = useUserStore()
const { hasAccess } = useAccess()
const loading = ref(false)
const saving = ref(false)
const tenantLoading = ref(false)
const tenantOptions = ref<TenantListItem[]>([])
const selectedTenantId = ref('')
const imageMimeOptions = [
  { label: '全部图片（image/*）', value: 'image/*' },
  { label: 'JPG / JPEG', value: 'image/jpeg' },
  { label: 'PNG', value: 'image/png' },
  { label: 'WEBP', value: 'image/webp' },
  { label: 'GIF', value: 'image/gif' },
]
const videoMimeOptions = [
  { label: '全部视频（video/*）', value: 'video/*' },
  { label: 'MP4', value: 'video/mp4' },
  { label: 'MOV', value: 'video/quicktime' },
  { label: 'AVI', value: 'video/x-msvideo' },
  { label: 'WEBM', value: 'video/webm' },
]

const BYTES_PER_KB = 1024
const KB_PER_MB = 1024

function bytesToKB(bytes?: number) {
  const value = Number(bytes || 0)
  return value > 0 ? Math.round(value / BYTES_PER_KB) : undefined
}

function kbToBytes(kb?: number) {
  const value = Number(kb || 0)
  return value > 0 ? Math.round(value * BYTES_PER_KB) : 0
}

function formatKBToMB(kb?: number) {
  const value = Number(kb || 0)
  if (value <= 0)
    return '约 0 MB'
  return `约 ${Number((value / KB_PER_MB).toFixed(2))} MB`
}

const isPlatformAdmin = computed(() => userStore.userInfo?.tenantRole === 'platform_admin')

const imageMaxSizeKB = computed({
  get: () => bytesToKB(formState.imageMaxSize),
  set: value => formState.imageMaxSize = kbToBytes(value),
})
const videoMaxSizeKB = computed({
  get: () => bytesToKB(formState.videoMaxSize),
  set: value => formState.videoMaxSize = kbToBytes(value),
})

const formState = reactive<TenantStorageConfig>({
  tenantId: '',
  provider: 'qiniu',
  accessKey: '',
  secretKey: '',
  bucket: '',
  bucketHost: '',
  uploadPrefix: '',
  expiresSeconds: 72000,
  imageMaxSize: 10485760,
  imageMimeTypes: 'image/*',
  videoMaxSize: 104857600,
  videoMimeTypes: 'video/*',
  enabled: true,
  remark: '',
})

function fillForm(data?: Partial<TenantStorageConfig>) {
  formState.tenantId = data?.tenantId || selectedTenantId.value || ''
  formState.provider = 'qiniu'
  formState.accessKey = data?.accessKey || ''
  formState.secretKey = data?.secretKey || ''
  formState.bucket = data?.bucket || ''
  formState.bucketHost = data?.bucketHost || ''
  formState.uploadPrefix = data?.uploadPrefix || ''
  formState.expiresSeconds = data?.expiresSeconds || 72000
  formState.imageMaxSize = data?.imageMaxSize || 10485760
  formState.imageMimeTypes = data?.imageMimeTypes || 'image/*'
  formState.videoMaxSize = data?.videoMaxSize || 104857600
  formState.videoMimeTypes = data?.videoMimeTypes || 'video/*'
  formState.enabled = data?.enabled ?? true
  formState.remark = data?.remark || ''
  formState.updateTime = data?.updateTime || ''
}

async function loadTenants() {
  if (!isPlatformAdmin.value)
    return
  tenantLoading.value = true
  try {
    const res = await listTenantsApi()
    const rows = Array.isArray(res.result) ? res.result : []
    tenantOptions.value = rows.filter(item => item.tenantType !== 'platform')
    if (!selectedTenantId.value && tenantOptions.value.length)
      selectedTenantId.value = tenantOptions.value[0].tenantId
  }
  catch (error) {
    console.warn('load tenants failed', error)
    messageService.error('租户列表加载失败')
  }
  finally {
    tenantLoading.value = false
  }
}

async function loadConfig() {
  loading.value = true
  try {
    const res = await getTenantStorageConfigApi({ tenantId: selectedTenantId.value || undefined })
    fillForm(res.result || { tenantId: selectedTenantId.value })
  }
  catch (error: any) {
    console.warn('load tenant storage failed', error)
    fillForm({ tenantId: selectedTenantId.value })
    if (error?.response?.status !== 404)
      messageService.warning('当前租户暂未配置云存储')
  }
  finally {
    loading.value = false
  }
}

async function saveConfig() {
  if (!formState.tenantId && selectedTenantId.value)
    formState.tenantId = selectedTenantId.value
  if (!formState.tenantId) {
    messageService.warning('请选择租户')
    return
  }
  if (!formState.accessKey.trim() || !formState.bucket.trim() || !formState.bucketHost.trim()) {
    messageService.warning('请填写 AccessKey、存储桶和访问域名')
    return
  }
  if (!formState.secretKey?.trim() && formState.secretKey !== '******') {
    messageService.warning('首次配置必须填写 SecretKey')
    return
  }
  saving.value = true
  try {
    const payload = { ...formState, secretKey: formState.secretKey === '******' ? '' : formState.secretKey }
    await saveTenantStorageConfigApi(payload)
    messageService.success('云存储配置已保存')
    await loadConfig()
  }
  catch (error: any) {
    console.error(error)
    messageService.error(error?.response?.data?.message || '保存失败')
  }
  finally {
    saving.value = false
  }
}

onMounted(async () => {
  await loadTenants()
  await loadConfig()
})

watch(selectedTenantId, () => {
  if (selectedTenantId.value)
    loadConfig()
})
</script>

<template>
  <div class="storage-page">
    <div class="storage-header">
      <div>
        <h1>云存储配置</h1>
        <p>每个租户使用自己的七牛云存储桶，上传凭证按当前租户独立签发。</p>
      </div>
      <a-button v-if="hasAccess(PlatformAccessEnum.storageEdit)" type="primary" :loading="saving" @click="saveConfig">保存配置</a-button>
    </div>

    <a-alert
      show-icon
      type="warning"
      class="storage-alert"
      message="生产环境不再兜底使用平台云存储：租户未配置或已停用时，上传接口会直接拒绝。"
    />

    <a-card :bordered="false" class="storage-card">
      <a-spin :spinning="loading">
        <div class="storage-section-title storage-section-title--with-switch">
          <div class="storage-section-title__main">
            <CloudUploadOutlined />
            <span>七牛云配置</span>
          </div>
          <div class="storage-section-title__switch">
            <span>启用状态</span>
            <a-switch v-model:checked="formState.enabled" checked-children="启用" un-checked-children="停用" />
          </div>
        </div>

        <a-form layout="vertical" class="storage-form">
          <a-form-item v-if="isPlatformAdmin" label="合作租户">
            <a-select
              v-model:value="selectedTenantId"
              show-search
              option-filter-prop="label"
              :loading="tenantLoading"
              :options="tenantOptions.map(item => ({ value: item.tenantId, label: `${item.tenantName}（${item.tenantId}）` }))"
              placeholder="请选择租户"
            />
          </a-form-item>

          <a-form-item label="AccessKey" required>
            <a-input v-model:value="formState.accessKey" placeholder="填写租户自己的七牛 AccessKey" />
          </a-form-item>

          <a-form-item label="SecretKey" required>
            <a-input-password v-model:value="formState.secretKey" placeholder="已保存后留空不修改；首次配置必填" />
          </a-form-item>

          <a-form-item label="Bucket" required>
            <a-input v-model:value="formState.bucket" placeholder="例如：tenant-a-assets" />
          </a-form-item>

          <a-form-item label="访问域名" required>
            <a-input v-model:value="formState.bucketHost" placeholder="例如：https://cdn.tenant-a.com/" />
          </a-form-item>

          <a-form-item label="上传前缀">
            <a-input v-model:value="formState.uploadPrefix" placeholder="例如：tenant-a/，可选" />
          </a-form-item>

          <a-form-item label="Token 有效期（秒）">
            <a-input-number v-model:value="formState.expiresSeconds" :min="60" :controls="false" class="full-input" />
          </a-form-item>

          <a-form-item>
            <template #label>
              <span class="limit-label">图片限制 <span class="limit-label__value">{{ formatKBToMB(imageMaxSizeKB) }}</span></span>
            </template>
            <a-input-group class="limit-input-group">
              <a-input-number v-model:value="imageMaxSizeKB" :min="1" :controls="false" class="size-input" addon-before="大小" addon-after="KB" />
              <a-select
                v-model:value="formState.imageMimeTypes"
                class="mime-input"
                :options="imageMimeOptions"
                placeholder="选择图片格式"
              />
            </a-input-group>
          </a-form-item>

          <a-form-item>
            <template #label>
              <span class="limit-label limit-label--video">视频限制 <span class="limit-label__value">{{ formatKBToMB(videoMaxSizeKB) }}</span></span>
            </template>
            <a-input-group class="limit-input-group">
              <a-input-number v-model:value="videoMaxSizeKB" :min="1" :controls="false" class="size-input" addon-before="大小" addon-after="KB" />
              <a-select
                v-model:value="formState.videoMimeTypes"
                class="mime-input"
                :options="videoMimeOptions"
                placeholder="选择视频格式"
              />
            </a-input-group>
          </a-form-item>

          <a-form-item v-if="isPlatformAdmin" label="备注">
            <a-input v-model:value="formState.remark" placeholder="记录客户云账号、桶用途或交付说明" />
          </a-form-item>
        </a-form>
      </a-spin>
    </a-card>

    <a-card :bordered="false" class="storage-side-card">
      <div class="storage-section-title">
        <SafetyCertificateOutlined />
        <span>生效范围</span>
      </div>
      <ul>
        <li>子总控后台上传：使用当前租户配置。</li>
        <li>机构端上传：使用机构所属租户配置。</li>
        <li>未配置或停用：上传 token 接口返回错误。</li>
        <li>SecretKey 不回显，留空保存表示不修改。</li>
      </ul>
    </a-card>
  </div>
</template>

<style scoped lang="less">
.storage-page {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 360px;
  gap: 14px;
  padding: 14px 16px;
}

.storage-header {
  grid-column: 1 / -1;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 18px;
  border-radius: 12px;
  background: #fff;

  h1 {
    margin: 0;
    color: rgba(0, 0, 0, 0.88);
    font-size: 20px;
    font-weight: 650;
    line-height: 28px;
  }

  p {
    margin: 3px 0 0;
    color: rgba(0, 0, 0, 0.45);
    font-size: 13px;
  }
}

.storage-alert {
  grid-column: 1 / -1;
}

.storage-card,
.storage-side-card {
  border-radius: 12px;
}

.storage-section-title {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 16px;
  color: rgba(0, 0, 0, 0.88);
  font-size: 15px;
  font-weight: 600;
}

.storage-section-title--with-switch {
  justify-content: space-between;
}

.storage-section-title__main,
.storage-section-title__switch {
  display: flex;
  align-items: center;
}

.storage-section-title__main {
  gap: 8px;
}

.storage-section-title__switch {
  gap: 10px;
  color: rgba(0, 0, 0, 0.65);
  font-size: 14px;
  font-weight: 400;
}

.storage-form {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0 16px;
}

.storage-form__full-row {
  grid-column: 1 / -1;
}

.full-input {
  width: 100%;
}

.limit-label {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  white-space: nowrap;
}

.limit-label__value {
  display: inline-flex;
  align-items: center;
  height: 20px;
  padding: 0 7px;
  border: 1px solid #91caff;
  border-radius: 10px;
  background: #e6f4ff;
  color: #1677ff;
  font-size: 12px;
  font-style: normal;
  font-weight: 500;
  line-height: 18px;
}

.limit-label--video .limit-label__value {
  border-color: #b7eb8f;
  background: #f6ffed;
  color: #52c41a;
}

.limit-input-group {
  display: flex !important;
  width: 100%;
  gap: 8px;
}

.limit-input-group :deep(.ant-input-number-group-wrapper),
.limit-input-group :deep(.ant-select) {
  flex: 1 1 0;
  width: 0 !important;
  min-width: 0;
}

.storage-side-card ul {
  margin: 0;
  padding-left: 18px;
  color: rgba(0, 0, 0, 0.55);
  line-height: 28px;
}

@media (max-width: 1100px) {
  .storage-page,
  .storage-form {
    display: block;
  }

  .storage-side-card {
    margin-top: 14px;
  }
}
</style>
