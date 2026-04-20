<script setup lang="ts">
import type { FormInstance, Rule } from 'ant-design-vue/es/form'
import { computed, reactive, ref, watch } from 'vue'
import type {
  GovernmentAccountMutationPayload,
  GovernmentAccountScopeItem,
  GovernmentLevel,
  GovernmentRoleOption,
} from '@/api/platform/government-accounts'
import {
  checkGovernmentUsernameAvailableApi,
  createGovernmentAccountApi,
  getGovernmentAccountDetailApi,
  getGovernmentRoleOptionsApi,
  updateGovernmentAccountApi,
} from '@/api/platform/government-accounts'
import { regionData } from '@/constants/region-data'
import PlatformModalShell from '../../shared/platform-modal-shell.vue'
import messageService from '@/utils/messageService'

const props = defineProps<{
  open: boolean
  accountId?: number | null
}>()

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'saved'): void
}>()

interface GovernmentAccountFormState {
  nickName: string
  mobile: string
  username: string
  password: string
  disabled: boolean
  level: GovernmentLevel
  roleId?: number
}

interface ScopePickerState {
  provinceCode?: string
  cityCode?: string
  districtCode?: string
}

const levelOptions: Array<{ label: string, value: GovernmentLevel }> = [
  { label: '超级监管', value: 'super' },
  { label: '省级', value: 'province' },
  { label: '市级', value: 'city' },
  { label: '区县级', value: 'district' },
]

function createInitialFormState(): GovernmentAccountFormState {
  return {
    nickName: '',
    mobile: '',
    username: '',
    password: '',
    disabled: false,
    level: 'province',
    roleId: undefined,
  }
}

const openModal = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

const formRef = ref<FormInstance>()
const detailLoading = ref(false)
const submitting = ref(false)
const roleLoading = ref(false)
const roleOptions = ref<GovernmentRoleOption[]>([])
const scopeList = ref<GovernmentAccountScopeItem[]>([])
const scopePicker = reactive<ScopePickerState>({
  provinceCode: undefined,
  cityCode: undefined,
  districtCode: undefined,
})

const formState = reactive<GovernmentAccountFormState>(createInitialFormState())

const isEdit = computed(() => Number(props.accountId || 0) > 0)
const modalTitle = computed(() => (isEdit.value ? '编辑政府账户' : '新建政府账户'))
const accountEnabled = computed({
  get: () => !formState.disabled,
  set: value => {
    formState.disabled = !value
  },
})
const selectedLevelLabel = computed(() => levelOptions.find(item => item.value === formState.level)?.label || '--')
const filteredRoleOptions = computed(() => roleOptions.value.filter(item => item.level === formState.level))
const provinceOptions = computed(() => regionData.map(item => ({
  label: item.label,
  value: item.value,
})))
const selectedProvince = computed(() => regionData.find(item => item.value === scopePicker.provinceCode))
const cityOptions = computed(() => (selectedProvince.value?.children || []).map(item => ({
  label: item.label,
  value: item.value,
})))
const selectedCity = computed(() => selectedProvince.value?.children?.find(item => item.value === scopePicker.cityCode))
const districtOptions = computed(() => (selectedCity.value?.children || []).map(item => ({
  label: item.label,
  value: item.value,
})))
const canAddScope = computed(() => {
  if (formState.level === 'super')
    return false
  if (formState.level === 'province')
    return Boolean(scopePicker.provinceCode)
  if (formState.level === 'city')
    return Boolean(scopePicker.provinceCode && scopePicker.cityCode)
  return Boolean(scopePicker.provinceCode && scopePicker.cityCode && scopePicker.districtCode)
})

const rules: Record<string, Rule[]> = {
  nickName: [
    { required: true, message: '请输入姓名', trigger: 'blur' },
  ],
  mobile: [
    { required: true, message: '请输入手机号', trigger: 'blur' },
    {
      validator: async (_rule, value) => {
        const text = String(value || '').trim()
        if (!text)
          return Promise.resolve()
        if (!/^1\d{10}$/.test(text))
          return Promise.reject(new Error('请输入正确的手机号'))
        return Promise.resolve()
      },
      trigger: 'blur',
    },
  ],
  username: [
    { required: true, message: '请输入登录账号', trigger: 'blur' },
    {
      validator: async (_rule, value) => {
        const text = String(value || '').trim()
        if (!text)
          return Promise.resolve()
        if (/\s/.test(text))
          return Promise.reject(new Error('登录账号不能包含空格'))
        return Promise.resolve()
      },
      trigger: 'blur',
    },
    {
      validator: async (_rule, value) => {
        const text = String(value || '').trim()
        if (!text)
          return Promise.resolve()

        const result = await checkGovernmentUsernameAvailability(text)
        if (result.available)
          return Promise.resolve()

        return Promise.reject(new Error(result.message || '登录账号已存在，请更换'))
      },
      trigger: 'blur',
    },
  ],
  password: [
    {
      validator: async (_rule, value) => {
        const text = String(value || '').trim()
        if (!isEdit.value && !text)
          return Promise.reject(new Error('请输入初始密码'))
        if (text && text.length < 6)
          return Promise.reject(new Error('密码长度不能少于6位'))
        return Promise.resolve()
      },
      trigger: 'blur',
    },
  ],
  roleId: [
    { required: true, message: '请选择监管角色', trigger: 'change' },
  ],
}

function resolveRequestErrorMessage(error: any, fallback: string) {
  return String(error?.response?.data?.message || error?.message || fallback).trim() || fallback
}

async function checkGovernmentUsernameAvailability(username: string) {
  const trimmed = String(username || '').trim()
  if (!trimmed) {
    return {
      available: false,
      message: '登录账号不能为空',
    }
  }

  try {
    const res = await checkGovernmentUsernameAvailableApi({
      username: trimmed,
      userId: props.accountId ? Number(props.accountId) : undefined,
    })
    if (res.code !== 200 || !res.result) {
      return {
        available: false,
        message: String(res.message || '登录账号校验失败，请稍后重试').trim(),
      }
    }

    return {
      available: !!res.result.available,
      message: String(res.result.message || '').trim(),
    }
  }
  catch (error: any) {
    return {
      available: false,
      message: resolveRequestErrorMessage(error, '登录账号校验失败，请稍后重试'),
    }
  }
}

function resetScopePicker() {
  scopePicker.provinceCode = undefined
  scopePicker.cityCode = undefined
  scopePicker.districtCode = undefined
}

function resetForm() {
  Object.assign(formState, createInitialFormState())
  scopeList.value = []
  resetScopePicker()
  formRef.value?.clearValidate()
}

function normalizeLevel(value?: string): GovernmentLevel {
  const raw = String(value || '').trim().toLowerCase()
  if (raw === 'super')
    return 'super'
  if (raw === 'city')
    return 'city'
  if (raw === 'district')
    return 'district'
  return 'province'
}

function normalizeScopeDisplay(scope: GovernmentAccountScopeItem) {
  if (scope.displayName)
    return scope.displayName
  if (scope.scopeLevel === 'province')
    return String(scope.provinceName || '').trim()
  if (scope.scopeLevel === 'city')
    return [scope.provinceName, scope.cityName].filter(Boolean).join('-')
  if (scope.scopeLevel === 'district')
    return [scope.provinceName, scope.cityName, scope.districtName].filter(Boolean).join('-')
  return '全部辖区'
}

function buildScopeKey(scope: GovernmentAccountScopeItem) {
  return [
    scope.scopeLevel,
    scope.provinceCode || '',
    scope.cityCode || '',
    scope.districtCode || '',
  ].join('|')
}

function handleLevelChange(value: GovernmentLevel) {
  formState.level = value
  formState.roleId = undefined
  scopeList.value = []
  resetScopePicker()
}

function handleProvinceChange(value?: string) {
  scopePicker.provinceCode = value
  scopePicker.cityCode = undefined
  scopePicker.districtCode = undefined
}

function handleCityChange(value?: string) {
  scopePicker.cityCode = value
  scopePicker.districtCode = undefined
}

function addScope() {
  if (formState.level === 'super')
    return

  const province = selectedProvince.value
  const city = selectedCity.value
  const district = selectedCity.value?.children?.find(item => item.value === scopePicker.districtCode)

  const scope: GovernmentAccountScopeItem = {
    scopeLevel: formState.level,
    provinceCode: scopePicker.provinceCode,
    provinceName: province?.label || '',
    cityCode: formState.level === 'province' ? '' : scopePicker.cityCode,
    cityName: formState.level === 'province' ? '' : city?.label || '',
    districtCode: formState.level === 'district' ? scopePicker.districtCode : '',
    districtName: formState.level === 'district' ? district?.label || '' : '',
  }
  scope.displayName = normalizeScopeDisplay(scope)

  const exists = scopeList.value.some(item => buildScopeKey(item) === buildScopeKey(scope))
  if (exists) {
    messageService.warning('该管辖范围已添加')
    return
  }

  scopeList.value = [...scopeList.value, scope]
  resetScopePicker()
}

function removeScope(index: number) {
  scopeList.value = scopeList.value.filter((_item, currentIndex) => currentIndex !== index)
}

async function loadRoleOptions() {
  roleLoading.value = true
  try {
    const res = await getGovernmentRoleOptionsApi()
    if (res.code !== 200) {
      messageService.error(res.message || '获取监管角色失败')
      roleOptions.value = []
      return
    }
    roleOptions.value = Array.isArray(res.result)
      ? res.result.map(item => ({
          ...item,
          level: normalizeLevel(item.level),
        }))
      : []
  }
  catch (error: any) {
    console.error('load government roles failed', error)
    messageService.error(resolveRequestErrorMessage(error, '获取监管角色失败'))
    roleOptions.value = []
  }
  finally {
    roleLoading.value = false
  }
}

async function loadDetail(accountId: number) {
  detailLoading.value = true
  try {
    const res = await getGovernmentAccountDetailApi({ id: accountId })
    if (res.code !== 200 || !res.result) {
      messageService.error(res.message || '获取政府账户详情失败')
      return
    }
    formState.nickName = String(res.result.nickName || '').trim()
    formState.mobile = String(res.result.mobile || '').trim()
    formState.username = String(res.result.username || '').trim()
    formState.password = ''
    formState.disabled = Boolean(res.result.disabled)
    formState.level = normalizeLevel(res.result.level)
    formState.roleId = Number(res.result.roleId || 0) || undefined
    scopeList.value = Array.isArray(res.result.scopes)
      ? res.result.scopes.map(item => ({
          ...item,
          scopeLevel: normalizeLevel(item.scopeLevel),
          displayName: normalizeScopeDisplay(item),
        }))
      : []
  }
  catch (error: any) {
    console.error('load government account detail failed', error)
    messageService.error(resolveRequestErrorMessage(error, '获取政府账户详情失败'))
  }
  finally {
    detailLoading.value = false
  }
}

function buildPayload(): GovernmentAccountMutationPayload {
  return {
    username: String(formState.username || '').trim(),
    password: String(formState.password || '').trim() || undefined,
    mobile: String(formState.mobile || '').trim(),
    nickName: String(formState.nickName || '').trim(),
    disabled: Boolean(formState.disabled),
    level: formState.level,
    roleId: Number(formState.roleId || 0),
    scopes: formState.level === 'super'
      ? []
      : scopeList.value.map(item => ({
          scopeLevel: formState.level,
          provinceCode: item.provinceCode,
          provinceName: item.provinceName,
          cityCode: item.cityCode,
          cityName: item.cityName,
          districtCode: item.districtCode,
          districtName: item.districtName,
          displayName: item.displayName,
        })),
  }
}

async function submitForm() {
  try {
    await formRef.value?.validate()
  }
  catch {
    return
  }

  if (!Number(formState.roleId || 0)) {
    messageService.error('请选择监管角色')
    return
  }

  if (formState.level !== 'super' && !scopeList.value.length) {
    messageService.error('请至少添加一个管辖范围')
    return
  }

  submitting.value = true
  try {
    const payload = buildPayload()
    if (isEdit.value && props.accountId) {
      const res = await updateGovernmentAccountApi({
        ...payload,
        id: Number(props.accountId),
      })
      if (res.code !== 200) {
        messageService.error(res.message || '更新政府账户失败')
        return
      }
      messageService.success('更新成功')
    }
    else {
      const res = await createGovernmentAccountApi(payload)
      if (res.code !== 200) {
        messageService.error(res.message || '创建政府账户失败')
        return
      }
      messageService.success('创建成功')
    }
    openModal.value = false
    emit('saved')
  }
  catch (error: any) {
    console.error('submit government account failed', error)
    messageService.error(resolveRequestErrorMessage(error, isEdit.value ? '更新政府账户失败' : '创建政府账户失败'))
  }
  finally {
    submitting.value = false
  }
}

watch(() => props.open, async (open) => {
  if (!open) {
    resetForm()
    return
  }

  resetForm()
  await loadRoleOptions()
  if (props.accountId)
    await loadDetail(Number(props.accountId))
}, { immediate: true })
</script>

<template>
  <PlatformModalShell
    v-model:open="openModal"
    :width="980"
    :title="modalTitle"
    modal-class="government-account-modal-shell"
    scrollable
  >
    <a-spin :spinning="detailLoading">
      <div class="government-account-modal">
        <div class="government-account-modal__aside">
          <div class="gov-card gov-card--form">
            <div class="gov-card__title">
              基础信息
            </div>

            <a-form ref="formRef" layout="vertical" :model="formState" :rules="rules">
              <a-form-item label="姓名" name="nickName">
                <a-input v-model:value="formState.nickName" :maxlength="20" placeholder="请输入姓名" />
              </a-form-item>

              <a-form-item label="手机号" name="mobile">
                <a-input v-model:value="formState.mobile" :maxlength="11" placeholder="请输入手机号" />
              </a-form-item>

              <a-form-item label="登录账号" name="username">
                <a-input v-model:value="formState.username" :maxlength="30" placeholder="请输入登录账号" />
              </a-form-item>

              <a-form-item name="password">
                <template #label>
                  <span v-if="!isEdit" class="password-label password-label--required">初始密码</span>
                  <span v-else>重置密码</span>
                </template>
                <a-input-password
                  v-model:value="formState.password"
                  :maxlength="20"
                  :placeholder="isEdit ? '不填写则保持原密码' : '请输入初始密码'"
                />
              </a-form-item>

              <a-form-item v-if="isEdit" label="账号状态">
                <a-switch
                  v-model:checked="accountEnabled"
                  checked-children="启用"
                  un-checked-children="停用"
                />
              </a-form-item>
            </a-form>
          </div>
        </div>

        <div class="government-account-modal__main">
          <div class="gov-card gov-card--scope">
            <div class="gov-card__header">
              <div>
                <div class="gov-card__title">
                  监管权限
                </div>
                <div class="gov-card__meta">
                  先选择层级与监管角色，再配置可监管的行政区划范围
                </div>
              </div>
            </div>

            <a-form layout="vertical" :model="formState">
              <a-form-item label="监管层级">
                <a-segmented
                  v-model:value="formState.level"
                  :options="levelOptions"
                  block
                  @change="handleLevelChange($event as GovernmentLevel)"
                />
              </a-form-item>

              <a-form-item name="roleId">
                <template #label>
                  <span class="form-label-required">监管角色</span>
                </template>
                <a-select
                  v-model:value="formState.roleId"
                  :loading="roleLoading"
                  :options="filteredRoleOptions.map(item => ({ label: item.roleName, value: item.roleId }))"
                  placeholder="请选择监管角色"
                />
              </a-form-item>
            </a-form>

            <div class="scope-panel">
              <div class="scope-panel__head">
                <div class="scope-panel__title" :class="{ 'scope-panel__title--required': formState.level !== 'super' }">
                  管辖范围
                </div>
                <div class="scope-panel__subtitle">
                  当前层级：{{ selectedLevelLabel }}
                </div>
              </div>

              <div v-if="formState.level === 'super'" class="scope-panel__empty">
                超级监管默认可查看全部辖区，无需额外配置范围。
              </div>

              <template v-else>
                <div class="scope-picker">
                  <a-select
                    :value="scopePicker.provinceCode"
                    class="scope-picker__item"
                    :options="provinceOptions"
                    placeholder="选择省份"
                    show-search
                    @update:value="handleProvinceChange"
                  />

                  <a-select
                    v-if="formState.level === 'city' || formState.level === 'district'"
                    :value="scopePicker.cityCode"
                    class="scope-picker__item"
                    :options="cityOptions"
                    :disabled="!scopePicker.provinceCode"
                    placeholder="选择城市"
                    show-search
                    @update:value="handleCityChange"
                  />

                  <a-select
                    v-if="formState.level === 'district'"
                    v-model:value="scopePicker.districtCode"
                    class="scope-picker__item"
                    :options="districtOptions"
                    :disabled="!scopePicker.cityCode"
                    placeholder="选择区县"
                    show-search
                  />

                  <a-button type="primary" :disabled="!canAddScope" @click="addScope">
                    添加范围
                  </a-button>
                </div>

                <div v-if="scopeList.length" class="scope-list">
                  <a-tag
                    v-for="(item, index) in scopeList"
                    :key="buildScopeKey(item)"
                    closable
                    class="scope-tag"
                    @close.prevent="removeScope(index)"
                  >
                    {{ item.displayName }}
                  </a-tag>
                </div>

                <div v-else class="scope-panel__empty">
                  还没有添加管辖范围，请先选择区域并点击“添加范围”。
                </div>
              </template>
            </div>
          </div>
        </div>
      </div>
    </a-spin>

    <template #footer>
      <a-button @click="openModal = false">
        取消
      </a-button>
      <a-button type="primary" :loading="submitting" @click="submitForm">
        {{ isEdit ? '保存' : '创建' }}
      </a-button>
    </template>
  </PlatformModalShell>
</template>

<style scoped lang="less">
.government-account-modal {
  display: grid;
  grid-template-columns: 320px minmax(0, 1fr);
  gap: 18px;
  padding-top: 8px;
  align-items: stretch;
}

.government-account-modal__aside,
.government-account-modal__main {
  min-width: 0;
  display: flex;
}

.gov-card {
  width: 100%;
  height: 100%;
  border: 1px solid #e8edf5;
  border-radius: 18px;
  background: #fff;
  box-shadow: 0 14px 32px rgba(15, 23, 42, 0.06);
}

.gov-card--form {
  padding: 20px;
  background:
    linear-gradient(180deg, rgba(22, 119, 255, 0.06) 0%, rgba(22, 119, 255, 0) 132px),
    #fff;
}

.gov-card--scope {
  padding: 20px 22px 22px;
}

.gov-card__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 18px;
}

.gov-card__title {
  color: #1f2329;
  font-size: 18px;
  font-weight: 700;
  line-height: 28px;
  margin-bottom: 8px;
}

.gov-card__meta {
  color: #8a919f;
  font-size: 12px;
  line-height: 20px;
}

.scope-panel {
  margin-top: 12px;
  padding: 18px;
  border-radius: 16px;
  background: #f8fafc;
}

.scope-panel__head {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 14px;
}

.scope-panel__title {
  color: #1f2329;
  font-size: 15px;
  font-weight: 600;
}

.scope-panel__title--required {
  position: relative;
  padding-left: 12px;
}

.scope-panel__title--required::before {
  position: absolute;
  left: 0;
  top: 50%;
  color: #ff4d4f;
  font-size: 14px;
  line-height: 1;
  transform: translateY(-46%);
  content: '*';
}

.scope-panel__subtitle {
  color: #8a919f;
  font-size: 12px;
}

.scope-picker {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.scope-picker__item {
  width: 180px;
}

.scope-list {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
  margin-top: 16px;
}

.scope-tag {
  padding: 5px 10px;
  border-radius: 999px;
  border-color: #d0dbe8;
  background: #fff;
  color: #425466;
}

.scope-panel__empty {
  margin-top: 16px;
  padding: 18px 16px;
  border: 1px dashed #d7deea;
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.88);
  color: #8a919f;
  font-size: 13px;
  line-height: 22px;
}

:deep(.government-account-modal-shell .ant-form-item) {
  margin-bottom: 18px;
}

:deep(.government-account-modal-shell .ant-form-item-label > label) {
  color: #1f2329;
  font-weight: 600;
}

.form-label-required,
.password-label--required {
  position: relative;
  padding-left: 12px;
}

.form-label-required::before,
.password-label--required::before {
  position: absolute;
  left: 0;
  top: 50%;
  color: #ff4d4f;
  font-size: 14px;
  line-height: 1;
  transform: translateY(-46%);
  content: '*';
}

@media (max-width: 1100px) {
  .government-account-modal {
    grid-template-columns: 1fr;
  }

  .scope-picker__item {
    width: 100%;
  }
}
</style>
