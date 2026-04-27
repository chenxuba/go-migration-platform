<script setup lang="ts">
import type { Rule } from 'ant-design-vue/es/form'
import { onMounted, reactive, ref } from 'vue'
import LoginRenderer from './login-renderer.vue'
import { useAuthorization } from '@/composables/authorization'
import { loginApi } from '~/api/common/login'
import { getLoginThemeApi, type TenantLoginBrandConfig } from '~/api/common/login-theme'
import { preloadPlatformHomeEntry, schedulePlatformHomePreload } from '~/router/route-preload'
import { reset401Status } from '~/utils/request'

const LOGIN_THEME_CACHE_KEY = 'PLATFORM_ADMIN_LOGIN_THEME:platform-admin'

const { t } = useI18nLocale()
const router = useRouter()
const route = useRoute()
const token = useAuthorization()
const userStore = useUserStore()
const submitLoading = ref(false)
const loginThemeReady = ref(true)

const formState = reactive({
  username: '',
  password: '',
})

const defaultBrand: Required<TenantLoginBrandConfig> = {
  template: 'business-split',
  brandName: '总控平台',
  logoUrl: '',
  loginTitle: '总控管理后台',
  loginSubtitle: '请输入账号密码登录',
  backgroundUrl: '',
  primaryColor: '#1677ff',
  copyright: '仅支持总部/总控账号登录。如需开通权限，请联系系统管理员。',
  heroBadge: '总控平台',
  heroTitle: '统一管理全部机构与平台能力',
  heroDescription: '从一个入口完成机构总览、启停筛查和总部级管理协同，保留现有后台的操作节奏，但切换为更适合总控场景的登录与信息层级。',
}
const loginBrand = reactive<Required<TenantLoginBrandConfig>>({ ...defaultBrand })

function mergeLoginBrand(next?: TenantLoginBrandConfig) {
  Object.assign(loginBrand, defaultBrand, next || {})
}

function readCachedLoginBrand() {
  try {
    const raw = window.localStorage.getItem(LOGIN_THEME_CACHE_KEY)
    if (!raw)
      return undefined
    return JSON.parse(raw) as TenantLoginBrandConfig
  }
  catch {
    return undefined
  }
}

function writeCachedLoginBrand(next?: TenantLoginBrandConfig) {
  if (!next)
    return
  try {
    window.localStorage.setItem(LOGIN_THEME_CACHE_KEY, JSON.stringify(next))
  }
  catch {
    // Ignore storage quota and privacy-mode errors; the login page can use defaults.
  }
}

async function loadLoginTheme() {
  mergeLoginBrand(readCachedLoginBrand())

  try {
    const res = await getLoginThemeApi('platform-admin')
    const brand = res.result?.loginBrand || res.data?.loginBrand
    writeCachedLoginBrand(brand)
    mergeLoginBrand(brand)
  }
  catch (error) {
    console.warn('load login theme failed', error)
    mergeLoginBrand()
  }
}

onMounted(() => {
  void loadLoginTheme()
  schedulePlatformHomePreload()
})

const rules: Record<string, Rule[]> = {
  username: [{ required: true, message: '请输入账号', trigger: 'change' }],
  password: [{ required: true, message: '请输入密码', trigger: 'change' }],
}

async function showLoginMessage(type: 'success' | 'error', content: string, options: Record<string, any> = {}) {
  const { default: messageService } = await import('@/utils/messageService')
  messageService[type](content, options)
}

function resolveLoginErrorMessage(error: any) {
  const backendMessage = String(error?.response?.data?.message || error?.message || '').trim()
  if (!backendMessage)
    return '登录失败，请稍后重试'
  if (backendMessage === '无权限')
    return '当前账号未开通总控端权限，请联系系统管理员配置总部/总控角色'
  return backendMessage
}

async function onSubmit() {
  submitLoading.value = true
  try {
    const { result } = await loginApi({
      username: formState.username.trim(),
      password: formState.password,
    })

    if (!result?.token) {
      await showLoginMessage('error', '登录失败，请检查账号或密码')
      return
    }

    token.value = result.token
    if (result.user)
      userStore.setUserInfo({ ...result.user, loginType: result.loginType })

    if (result.tenantId) {
      const hostname = window.location.hostname.toLowerCase()
      window.localStorage.setItem(`PLATFORM_ADMIN_TENANT_ID:${hostname}`, result.tenantId)
      window.localStorage.setItem('PLATFORM_ADMIN_TENANT_ID', result.tenantId)
    }
    reset401Status()
    void preloadPlatformHomeEntry()

    void showLoginMessage('success', t('pages.login.notification.success.title', '登录成功'), { duration: 1500 })

    const redirect = typeof route.query.redirect === 'string'
      ? decodeURIComponent(route.query.redirect)
      : '/'
    const safeRedirect = ['/401', '/403', '/404', '/500', '/502'].includes(redirect) ? '/' : redirect
    await router.replace(safeRedirect || '/')
  }
  catch (error: any) {
    console.error('platform login failed', error)
    await showLoginMessage('error', resolveLoginErrorMessage(error))
  }
  finally {
    submitLoading.value = false
  }
}
</script>

<template>
  <LoginRenderer
    :brand="loginBrand"
    :form-state="formState"
    :submit-loading="submitLoading"
    :rules="rules"
    :ready="loginThemeReady"
    @submit="onSubmit"
  />
</template>
