<script setup lang="ts">
import type { FormInstance, Rule } from 'ant-design-vue/es/form'
import type { UploadRequestOption } from 'ant-design-vue/es/vc-upload/interface'
import type {
  InstitutionDetail,
  InstitutionGeocodePayload,
  InstitutionMutationPayload,
  InstitutionProfile,
} from '@/api/platform/institutions'
import { CloseOutlined } from '@ant-design/icons-vue'
import * as qiniu from 'qiniu-js'
import {
  createInstitutionApi,
  geocodeInstitutionApi,
  getInstitutionDetailApi,
  updateInstitutionApi,
} from '@/api/platform/institutions'
import { regionData } from '@/constants/region-data'
import { getQiniuToken } from '@/api/qiniu'
import { debounce } from 'lodash-es'
import messageService from '@/utils/messageService'

const props = defineProps<{
  open: boolean
  institutionId?: number | null
}>()

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'saved'): void
}>()

interface InstitutionFormState {
  id?: number
  organCode: string
  organName: string
  loginName: string
  mobile: string
  principal: string
  provinceCode?: string
  province: string
  cityCode?: string
  city: string
  regionCode?: string
  region: string
  address: string
  lng?: number
  lat?: number
  concatPhone: string
  fixedPhone: string
  logo: string
  remark: string
  enabled: boolean
  organLabel: string
  businessTime: string
  description: string
  video: string
  galleryImages: string[]
}

interface RegionSelection {
  provinceCode?: string
  cityCode?: string
  regionCode?: string
}

const geocodeSourceMap: Record<string, string> = {
  same_address: '同地址匹配',
  amap: '高德解析',
  nominatim: '地址解析',
  region_average: '区县均值',
  city_average: '城市均值',
  province_average: '省份均值',
}

const directCountyCityLabels = new Set([
  '市辖区',
  '县',
  '省直辖县级行政区划',
  '自治区直辖县级行政区划',
])

function normalizeRegionCode(value?: number | string | null) {
  const raw = String(value ?? '').trim()
  if (!raw || raw === '0')
    return ''

  if (raw.length >= 6)
    return raw

  if (raw.length === 2)
    return `${raw}0000`

  if (raw.length === 4)
    return `${raw}00`

  return raw
}

function matchCityName(provinceName: string, optionName: string, currentName: string) {
  const name = String(currentName || '').trim()
  if (!name)
    return false

  if (optionName === name)
    return true

  if (optionName === '市辖区' && name === provinceName)
    return true

  if ((optionName === '省直辖县级行政区划' || optionName === '自治区直辖县级行政区划') && name === '直辖县级')
    return true

  return false
}

function findRegionSelectionByNames(provinceName: string, cityName: string, regionName: string): RegionSelection | null {
  const provinceOption = regionData.find(item => item.label === String(provinceName || '').trim())
  if (!provinceOption)
    return null

  const provinceChildren = provinceOption.children || []
  const cityOption = provinceChildren.find(item => matchCityName(provinceOption.label, item.label, cityName))
    || (provinceChildren.length === 1 ? provinceChildren[0] : undefined)

  if (!cityOption) {
    return {
      provinceCode: provinceOption.value,
      cityCode: '',
      regionCode: '',
    }
  }

  const regionOption = (cityOption.children || []).find(item => item.label === String(regionName || '').trim())

  return {
    provinceCode: provinceOption.value,
    cityCode: cityOption.value,
    regionCode: regionOption?.value || '',
  }
}

function resolveRegionSelection(detail: InstitutionDetail): RegionSelection {
  const provinceCode = normalizeRegionCode(detail.provinceCode)
  const cityCode = normalizeRegionCode(detail.cityCode)
  const regionCode = normalizeRegionCode(detail.regionCode)

  if (provinceCode || cityCode || regionCode) {
    return {
      provinceCode,
      cityCode,
      regionCode,
    }
  }

  return findRegionSelectionByNames(detail.province, detail.city, detail.region || '')
    || { provinceCode: '', cityCode: '', regionCode: '' }
}

function createInitialFormState(): InstitutionFormState {
  return {
    organCode: '',
    organName: '',
    loginName: '',
    mobile: '',
    principal: '',
    provinceCode: undefined,
    province: '',
    cityCode: undefined,
    city: '',
    regionCode: undefined,
    region: '',
    address: '',
    lng: undefined,
    lat: undefined,
    concatPhone: '',
    fixedPhone: '',
    logo: '',
    remark: '',
    enabled: true,
    organLabel: '',
    businessTime: '',
    description: '',
    video: '',
    galleryImages: [],
  }
}

const openModal = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

const formRef = ref<FormInstance>()
const detailLoading = ref(false)
const submitting = ref(false)
const geocoding = ref(false)
const uploadingLogo = ref(false)
const logoUploadProgress = ref(0)
const geocodeSource = ref('')
const geocodeResolvedAddress = ref('')
const suspendAutoResolve = ref(false)
const lastResolvedAddressKey = ref('')

const formState = reactive<InstitutionFormState>(createInitialFormState())

const isEdit = computed(() => Number(props.institutionId || 0) > 0)
const modalTitle = computed(() => (isEdit.value ? '编辑机构' : '创建机构'))
const geocodeSourceLabel = computed(() => geocodeSourceMap[geocodeSource.value] || '')
const hasLogo = computed(() => !!String(formState.logo || '').trim())
const logoInitial = computed(() => String(formState.organName || '').trim().slice(0, 1) || '机')
const provinceOptions = computed(() => regionData.map(({ value, label }) => ({ value, label })))
const selectedProvinceOption = computed(() => regionData.find(item => item.value === formState.provinceCode))
const cityOptions = computed(() => (selectedProvinceOption.value?.children || []).map(({ value, label }) => ({ value, label })))
const selectedCityOption = computed(() => selectedProvinceOption.value?.children?.find(item => item.value === formState.cityCode))
const regionOptions = computed(() => (selectedCityOption.value?.children || []).map(({ value, label }) => ({ value, label })))
const selectedRegionOption = computed(() => selectedCityOption.value?.children?.find(item => item.value === formState.regionCode))

const rules: Record<string, Rule[]> = {
  organName: [{ required: true, message: '请输入机构名称', trigger: 'blur' }],
  loginName: [{ required: true, message: '请输入登录账号', trigger: 'blur' }],
  mobile: [
    { required: true, message: '请输入联系人电话', trigger: 'blur' },
    { pattern: /^\d{11}$/, message: '联系电话需为 11 位手机号', trigger: 'blur' },
  ],
  principal: [{ required: true, message: '请输入负责人姓名', trigger: 'blur' }],
  provinceCode: [{ required: true, message: '请选择省份', trigger: 'change' }],
  cityCode: [{ required: true, message: '请选择城市', trigger: 'change' }],
  regionCode: [{ required: true, message: '请选择区县', trigger: 'change' }],
  address: [{ required: true, message: '请输入详细地址', trigger: 'blur' }],
}

function resetForm() {
  Object.assign(formState, createInitialFormState())
  geocodeSource.value = ''
  geocodeResolvedAddress.value = ''
  suspendAutoResolve.value = false
  lastResolvedAddressKey.value = ''
  uploadingLogo.value = false
  logoUploadProgress.value = 0
  debouncedResolveCoordinates.cancel()
  nextTick(() => {
    formRef.value?.clearValidate?.()
  })
}

function closeModal() {
  emit('update:open', false)
}

function syncRegionLabels(fallback?: Partial<Pick<InstitutionFormState, 'province' | 'city' | 'region'>>) {
  formState.province = selectedProvinceOption.value?.label || String(fallback?.province || '').trim()
  formState.city = selectedCityOption.value?.label || String(fallback?.city || '').trim()
  formState.region = selectedRegionOption.value?.label || String(fallback?.region || '').trim()
}

function applyRegionSelection(selection: Partial<RegionSelection>, fallback?: Partial<Pick<InstitutionFormState, 'province' | 'city' | 'region'>>) {
  formState.provinceCode = selection.provinceCode ? String(selection.provinceCode) : undefined
  formState.cityCode = selection.cityCode ? String(selection.cityCode) : undefined
  formState.regionCode = selection.regionCode ? String(selection.regionCode) : undefined
  syncRegionLabels(fallback)
}

function handleProvinceChange(value?: string | number) {
  formState.provinceCode = value ? String(value) : undefined
  formState.cityCode = undefined
  formState.regionCode = undefined
  syncRegionLabels()
}

function handleCityChange(value?: string | number) {
  formState.cityCode = value ? String(value) : undefined
  formState.regionCode = undefined
  syncRegionLabels()
}

function handleRegionChange(value?: string | number) {
  formState.regionCode = value ? String(value) : undefined
  syncRegionLabels()
}

function buildGeocodeLocation() {
  const province = formState.province.trim()
  let city = formState.city.trim()
  let region = formState.region.trim()

  if (city === '市辖区') {
    city = province
  }
  else if (directCountyCityLabels.has(city)) {
    city = region || province
    region = ''
  }

  return { province, city, region }
}

function buildAddressKey() {
  const location = buildGeocodeLocation()
  return [
    location.province,
    location.city,
    location.region,
    formState.address.trim(),
  ].join('|')
}

function hasResolvedCoordinate() {
  return Number.isFinite(Number(formState.lng))
    && Number.isFinite(Number(formState.lat))
    && Number(formState.lng) !== 0
    && Number(formState.lat) !== 0
}

function buildGeocodePayload(): InstitutionGeocodePayload | null {
  const { province, city, region } = buildGeocodeLocation()
  const address = formState.address.trim()
  if (!province || !city || !address)
    return null

  return {
    province,
    city,
    region: region || undefined,
    address,
  }
}

function formatCoordinate(value?: number) {
  if (!Number.isFinite(Number(value)) || Number(value) === 0)
    return ''
  return Number(value).toFixed(6)
}

function buildProfilePayload(): InstitutionProfile | undefined {
  const galleryImages = formState.galleryImages
    .map(item => String(item || '').trim())
    .filter(Boolean)

  const profile: InstitutionProfile = {
    organLabel: formState.organLabel.trim() || undefined,
    businessTime: formState.businessTime.trim() || undefined,
    description: formState.description.trim() || undefined,
    video: formState.video.trim() || undefined,
    galleryImages: galleryImages.length ? galleryImages : undefined,
  }

  return Object.values(profile).some(value => Array.isArray(value) ? value.length > 0 : !!value)
    ? profile
    : undefined
}

function handleMobileInput(event: Event) {
  const target = event.target as HTMLInputElement
  formState.mobile = String(target.value || '').replace(/\D/g, '').slice(0, 11)
}

function beforeLogoUpload(file: File) {
  if (!file.type.startsWith('image/')) {
    messageService.error('只能上传图片文件')
    return false
  }

  if (file.size / 1024 / 1024 >= 5) {
    messageService.error('图片大小不能超过 5MB')
    return false
  }

  return true
}

async function handleLogoUpload(options: UploadRequestOption) {
  const rawFile = options.file as File
  if (!rawFile || !beforeLogoUpload(rawFile)) {
    options.onError?.(new Error('invalid file'))
    return
  }

  uploadingLogo.value = true
  logoUploadProgress.value = 0

  try {
    const tokenRes: any = await getQiniuToken()
    const { token, uuid, buckethostname } = tokenRes.result || {}
    if (!token || !uuid || !buckethostname)
      throw new Error('获取上传凭证失败')

    const ext = rawFile.name.includes('.') ? rawFile.name.slice(rawFile.name.lastIndexOf('.')) : '.png'
    const key = `institution/logo/${uuid}${ext}`
    const config = {
      useCdnDomain: true,
      region: qiniu.region.z0,
    }
    const putExtra = {
      fname: rawFile.name,
      mimeType: rawFile.type,
    }

    const observable = qiniu.upload(rawFile, key, token, putExtra, config)
    observable.subscribe({
      next(result) {
        logoUploadProgress.value = Math.floor(result.total.percent)
      },
      error(error) {
        console.error('upload institution logo failed', error)
        messageService.error(error?.message || '上传机构 Logo 失败')
        uploadingLogo.value = false
        logoUploadProgress.value = 0
        options.onError?.(error)
      },
      complete(result) {
        formState.logo = `${buckethostname}${result.key}`
        uploadingLogo.value = false
        logoUploadProgress.value = 100
        messageService.success('机构 Logo 上传成功')
        options.onSuccess?.(result as any)
      },
    })
  }
  catch (error: any) {
    console.error('prepare institution logo upload failed', error)
    messageService.error(error?.message || '上传机构 Logo 失败')
    uploadingLogo.value = false
    logoUploadProgress.value = 0
    options.onError?.(error)
  }
}

async function resolveCoordinates(manual = false) {
  const payload = buildGeocodePayload()
  if (!payload) {
    if (manual)
      messageService.warning('请先选择完整的省市区并填写详细地址')
    return false
  }

  geocoding.value = true
  try {
    const res = await geocodeInstitutionApi(payload)
    if (res.code !== 200 || !res.result) {
      if (manual)
        messageService.error(res.message || '未获取到机构坐标')
      return false
    }

    formState.lng = Number(res.result.lng || 0) || undefined
    formState.lat = Number(res.result.lat || 0) || undefined
    geocodeSource.value = String(res.result.source || '')
    geocodeResolvedAddress.value = String(res.result.resolvedAddress || '')
    lastResolvedAddressKey.value = buildAddressKey()

    if (manual)
      messageService.success('机构坐标已更新')

    return hasResolvedCoordinate()
  }
  catch (error: any) {
    console.error('resolve institution coordinates failed', error)
    if (manual)
      messageService.error(error?.message || '未获取到机构坐标')
    return false
  }
  finally {
    geocoding.value = false
  }
}

const debouncedResolveCoordinates = debounce(() => {
  void resolveCoordinates(false)
}, 700)

async function loadInstitutionDetail(id: number) {
  detailLoading.value = true
  suspendAutoResolve.value = true
  try {
    const res = await getInstitutionDetailApi({ id })
    if (res.code !== 200 || !res.result) {
      messageService.error(res.message || '获取机构详情失败')
      return
    }

    const detail = res.result
    formState.id = detail.id
    formState.organCode = String(detail.organCode || '')
    formState.organName = String(detail.organName || '')
    formState.loginName = String(detail.loginName || '')
    formState.mobile = String(detail.mobile || '')
    formState.principal = String(detail.principal || '')
    applyRegionSelection(resolveRegionSelection(detail), {
      province: String(detail.province || ''),
      city: String(detail.city || ''),
      region: String(detail.region || ''),
    })
    formState.address = String(detail.address || '')
    formState.lng = Number(detail.lng || 0) || undefined
    formState.lat = Number(detail.lat || 0) || undefined
    formState.concatPhone = String(detail.concatPhone || '')
    formState.fixedPhone = String(detail.fixedPhone || '')
    formState.remark = String(detail.remark || '')
    formState.logo = String(detail.logo || '')
    formState.enabled = !!detail.enabled
    formState.organLabel = String(detail.profile?.organLabel || '')
    formState.businessTime = String(detail.profile?.businessTime || '')
    formState.description = String(detail.profile?.description || '')
    formState.video = String(detail.profile?.video || '')
    formState.galleryImages = Array.isArray(detail.profile?.galleryImages)
      ? detail.profile.galleryImages.filter(Boolean)
      : []

    geocodeSource.value = ''
    geocodeResolvedAddress.value = ''
    lastResolvedAddressKey.value = buildAddressKey()
  }
  catch (error: any) {
    console.error('load institution detail failed', error)
    messageService.error(error?.message || '获取机构详情失败')
  }
  finally {
    detailLoading.value = false
    nextTick(() => {
      suspendAutoResolve.value = false
    })
  }
}

watch(
  () => [props.open, props.institutionId] as const,
  ([open, institutionId]) => {
    if (!open) {
      resetForm()
      return
    }

    if (institutionId) {
      void loadInstitutionDetail(Number(institutionId))
      return
    }

    resetForm()
  },
  { immediate: true },
)

watch(
  () => buildAddressKey(),
  (value) => {
    if (!props.open || detailLoading.value || suspendAutoResolve.value)
      return

    if (!value) {
      formState.lng = undefined
      formState.lat = undefined
      geocodeSource.value = ''
      geocodeResolvedAddress.value = ''
      lastResolvedAddressKey.value = ''
      return
    }

    if (value === lastResolvedAddressKey.value)
      return

    formState.lng = undefined
    formState.lat = undefined
    geocodeSource.value = ''
    geocodeResolvedAddress.value = ''
    debouncedResolveCoordinates()
  },
)

function buildPayload(): InstitutionMutationPayload {
  return {
    organName: formState.organName.trim(),
    loginName: formState.loginName.trim(),
    mobile: formState.mobile.trim(),
    principal: formState.principal.trim(),
    provinceCode: formState.provinceCode ? Number(formState.provinceCode) : undefined,
    province: formState.province.trim(),
    cityCode: formState.cityCode ? Number(formState.cityCode) : undefined,
    city: formState.city.trim(),
    regionCode: formState.regionCode ? Number(formState.regionCode) : undefined,
    region: formState.region.trim() || undefined,
    address: formState.address.trim(),
    lng: hasResolvedCoordinate() ? Number(formState.lng) : undefined,
    lat: hasResolvedCoordinate() ? Number(formState.lat) : undefined,
    concatPhone: formState.concatPhone.trim() || undefined,
    fixedPhone: formState.fixedPhone.trim() || undefined,
    remark: formState.remark.trim() || undefined,
    logo: formState.logo.trim() || undefined,
    enabled: !!formState.enabled,
    profile: buildProfilePayload(),
  }
}

async function submitForm() {
  try {
    await formRef.value?.validate()
  }
  catch {
    return
  }

  if (!hasResolvedCoordinate()) {
    const resolved = await resolveCoordinates(true)
    if (!resolved)
      return
  }

  submitting.value = true
  try {
    const payload = buildPayload()
    const res = isEdit.value && formState.id
      ? await updateInstitutionApi({ id: formState.id, ...payload })
      : await createInstitutionApi(payload)

    if (res.code !== 200) {
      messageService.error(res.message || (isEdit.value ? '更新机构失败' : '新增机构失败'))
      return
    }

    messageService.success(isEdit.value ? '机构更新成功' : '机构新增成功')
    emit('saved')
    closeModal()
  }
  catch (error: any) {
    console.error('submit institution failed', error)
    messageService.error(error?.message || (isEdit.value ? '更新机构失败' : '新增机构失败'))
  }
  finally {
    submitting.value = false
  }
}
</script>

<template>
  <a-modal
    v-model:open="openModal"
    centered
    destroy-on-close
    :keyboard="false"
    :closable="false"
    :mask-closable="false"
    :width="980"
    class="createStu-modal-content-box institution-create-modal"
  >
    <template #title>
      <div class="institution-modal__titlebar">
        <span>{{ modalTitle }}</span>
        <a-button type="text" class="close-btn" @click="closeModal">
          <template #icon>
            <CloseOutlined class="close-icon" />
          </template>
        </a-button>
      </div>
    </template>

    <a-spin :spinning="detailLoading">
      <div class="institution-content scrollbar">
        <a-form ref="formRef" layout="vertical" :model="formState" :rules="rules" class="institution-form">
          <a-row :gutter="24" class="institution-form__top">
            <a-col :xs="24" :sm="24" :md="10" :lg="8" :xl="8">
              <div class="left-panel">
                <a-form-item>
                  <div class="avatar-upload-wrapper">
                    <a-avatar v-if="hasLogo" :src="formState.logo" :size="88" />
                    <a-avatar v-else :size="88" class="institution-avatar-placeholder">
                      {{ logoInitial }}
                    </a-avatar>

                    <a-upload
                      :custom-request="handleLogoUpload"
                      :show-upload-list="false"
                      accept="image/*"
                      :disabled="uploadingLogo"
                    >
                      <a-button :loading="uploadingLogo">
                        上传机构 Logo
                      </a-button>
                    </a-upload>
                  </div>

                  <div class="upload-tip">
                    支持 jpg、png 格式，大小不超过 5MB
                  </div>
                  <div v-if="uploadingLogo" class="upload-progress">
                    上传中 {{ logoUploadProgress }}%
                  </div>
                </a-form-item>
              </div>
            </a-col>

            <a-col :xs="24" :sm="24" :md="14" :lg="16" :xl="16">
              <div class="right-panel">
                <a-row :gutter="24">
                  <a-col :xs="24" :md="12">
                    <a-form-item label="机构名称：" name="organName">
                      <a-input v-model:value="formState.organName" :maxlength="64" placeholder="请输入机构名称" />
                    </a-form-item>
                  </a-col>

                  <a-col :xs="24" :md="12">
                    <a-form-item label="登录账号：" name="loginName">
                      <a-input v-model:value="formState.loginName" :maxlength="22" placeholder="请输入登录账号" />
                    </a-form-item>
                  </a-col>

                  <a-col :xs="24" :md="12">
                    <a-form-item label="联系人电话：" name="mobile">
                      <a-input
                        v-model:value="formState.mobile"
                        :maxlength="11"
                        placeholder="请输入手机号"
                        @input="handleMobileInput"
                      />
                    </a-form-item>
                  </a-col>

                  <a-col :xs="24" :md="12">
                    <a-form-item label="负责人：" name="principal">
                      <a-input v-model:value="formState.principal" :maxlength="64" placeholder="请输入负责人姓名" />
                    </a-form-item>
                  </a-col>
                </a-row>
              </div>
            </a-col>
          </a-row>

          <div class="system-grid">
            <div class="system-grid__item">
              <a-form-item label="省份：" name="provinceCode">
                <a-select
                  v-model:value="formState.provinceCode"
                  :options="provinceOptions"
                  show-search
                  option-filter-prop="label"
                  placeholder="请选择省份"
                  @change="handleProvinceChange"
                />
              </a-form-item>
            </div>

            <div class="system-grid__item">
              <a-form-item label="城市：" name="cityCode">
                <a-select
                  v-model:value="formState.cityCode"
                  :options="cityOptions"
                  :disabled="!formState.provinceCode"
                  show-search
                  option-filter-prop="label"
                  placeholder="请选择城市"
                  @change="handleCityChange"
                />
              </a-form-item>
            </div>

            <div class="system-grid__item">
              <a-form-item label="区县：" name="regionCode">
                <a-select
                  v-model:value="formState.regionCode"
                  :options="regionOptions"
                  :disabled="!formState.cityCode"
                  show-search
                  option-filter-prop="label"
                  placeholder="请选择区县"
                  @change="handleRegionChange"
                />
              </a-form-item>
            </div>

            <div class="system-grid__item">
              <a-form-item label="机构简称：">
                <a-input v-model:value="formState.organLabel" :maxlength="127" placeholder="请输入机构简称" />
              </a-form-item>
            </div>

            <div class="system-grid__item system-grid__item--full">
              <a-form-item label="详细地址：" name="address">
                <a-input v-model:value="formState.address" :maxlength="256" placeholder="请输入详细地址，用于自动获取经纬度" />
              </a-form-item>
            </div>

            <div class="system-grid__item system-grid__item--full">
              <a-form-item label="坐标定位：">
                <div class="coordinate-inline">
                  <a-input class="coordinate-inline__field" :value="formatCoordinate(formState.lng)" disabled placeholder="自动获取经度" />
                  <a-input class="coordinate-inline__field" :value="formatCoordinate(formState.lat)" disabled placeholder="自动获取纬度" />
                  <a-button :loading="geocoding" @click="resolveCoordinates(true)">
                    获取坐标
                  </a-button>
                </div>

                <div v-if="geocodeSourceLabel || geocodeResolvedAddress" class="coordinate-extra">
                  <span v-if="geocodeSourceLabel">{{ geocodeSourceLabel }}</span>
                  <span v-if="geocodeResolvedAddress">{{ geocodeResolvedAddress }}</span>
                </div>
              </a-form-item>
            </div>

            <div class="system-grid__item system-grid__item--full">
              <a-form-item label="机构备注：">
                <a-textarea v-model:value="formState.remark" :rows="3" :maxlength="255" placeholder="请输入备注" />
              </a-form-item>
            </div>
          </div>
        </a-form>
      </div>
    </a-spin>

    <template #footer>
      <a-button danger ghost @click="closeModal">
        关闭
      </a-button>
      <a-button type="primary" ghost :loading="submitting" @click="submitForm">
        确定
      </a-button>
    </template>
  </a-modal>
</template>

<style scoped>
.institution-modal__titlebar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  color: #1f2329;
  font-size: 22px;
  font-weight: 700;
  line-height: 32px;
}

.institution-content {
  max-height: calc(100vh - 155px);
  padding: 24px 40px 0 !important;
  overflow: auto;
}

.institution-form__top {
  margin-bottom: 4px;
}

.avatar-upload-wrapper {
  display: flex;
  align-items: center;
  gap: 16px;
}

.institution-avatar-placeholder {
  background: #d6e4ff;
  color: #2f54eb;
  font-size: 32px;
  font-weight: 700;
}

.upload-tip,
.upload-progress {
  margin-top: 8px;
  color: #999;
  font-size: 12px;
}

.system-grid {
  display: flex;
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0 24px;
}

.system-grid__item {
  min-width: 0;
}

.system-grid__item--full {
  grid-column: 1 / -1;
}

.coordinate-inline {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;
}

.coordinate-inline__field {
  flex: 1;
  min-width: 0;
}

.coordinate-extra {
  display: flex;
  gap: 12px;
  margin-top: 8px;
  color: #999;
  font-size: 12px;
  line-height: 20px;
  flex-wrap: wrap;
}

.close-btn:hover {
  background: transparent;
}

.close-btn {
  width: 40px;
  height: 40px;
  color: #1f2329;
  font-size: 22px;
}

.close-btn:hover .close-icon {
  animation: icon-rotate 0.3s linear;
}

@keyframes icon-rotate {
  from {
    transform: rotate(0deg);
  }

  to {
    transform: rotate(180deg);
  }
}

@media (max-width: 992px) {
  .institution-content {
    padding: 20px 20px 0 !important;
  }

  .system-grid {
    grid-template-columns: 1fr;
    gap: 0;
  }

  .system-grid__item--full {
    grid-column: auto;
  }

  .coordinate-inline {
    flex-direction: column;
    align-items: stretch;
  }
}

@media (max-width: 576px) {
  .avatar-upload-wrapper {
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>

<style>
.createStu-modal-content-box .ant-modal-header {
  padding: 10px 16px !important;
  margin-bottom: 0;
}

.createStu-modal-content-box .ant-modal-body {
  padding: 0 !important;
}
</style>
