<script setup lang="ts">
import { LockOutlined, SafetyCertificateOutlined, UserOutlined } from '@ant-design/icons-vue'
import type { Rule } from 'ant-design-vue/es/form'
import { computed, defineAsyncComponent, onMounted, reactive, ref } from 'vue'
import SelectLang from '@/components/select-lang/index.vue'
import { useAuthorization } from '@/composables/authorization'
import messageService from '@/utils/messageService'
import { loginApi } from '~/api/common/login'
import { getLoginThemeApi, type TenantLoginBrandConfig } from '~/api/common/login-theme'
import { useAppStore } from '~/stores/app'
import { reset401Status } from '~/utils/request'

const { t } = useI18nLocale()

const loginTemplateComponents = {
  'center-card': defineAsyncComponent(() => import('./login-templates/center-card.vue')),
  'minimal-portal': defineAsyncComponent(() => import('./login-templates/minimal-portal.vue')),
}
const router = useRouter()
const route = useRoute()
const token = useAuthorization()
const appStore = useAppStore()

const formRef = ref()
const submitLoading = ref(false)

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
const loginThemeReady = ref(false)
const loginBrand = reactive<Required<TenantLoginBrandConfig>>({ ...defaultBrand })
const loginThemeStyle = computed(() => ({
  '--tenant-primary': loginBrand.primaryColor || defaultBrand.primaryColor,
  '--tenant-bg-image': loginBrand.backgroundUrl ? `url(${loginBrand.backgroundUrl})` : 'none',
}))
const currentLoginComponent = computed(() => loginTemplateComponents[loginBrand.template] || null)
const dynamicLoginProps = computed(() => ({ brand: loginBrand, formState, submitLoading: submitLoading.value, rules }))

function mergeLoginBrand(next?: TenantLoginBrandConfig) {
  Object.assign(loginBrand, defaultBrand, next || {})
  if (loginBrand.primaryColor)
    appStore.toggleColorPrimary(loginBrand.primaryColor)
}

async function loadLoginTheme() {
  try {
    const res = await getLoginThemeApi('platform-admin')
    mergeLoginBrand(res.result?.loginBrand || res.data?.loginBrand)
  }
  catch (error) {
    console.warn('load login theme failed', error)
    mergeLoginBrand()
  }
  finally {
    loginThemeReady.value = true
  }
}

onMounted(loadLoginTheme)

const rules: Record<string, Rule[]> = {
  username: [
    {
      required: true,
      message: '请输入账号',
      trigger: 'change',
    },
  ],
  password: [
    {
      required: true,
      message: '请输入密码',
      trigger: 'change',
    },
  ],
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
    await formRef.value?.validate()
    const { result } = await loginApi({
      username: formState.username.trim(),
      password: formState.password,
    })

    if (!result?.token) {
      messageService.error('登录失败，请检查账号或密码')
      return
    }

    token.value = result.token
    if (result.tenantId)
      window.localStorage.setItem('PLATFORM_ADMIN_TENANT_ID', result.tenantId)
    reset401Status()

    messageService.success(t('pages.login.notification.success.title', '登录成功'), { duration: 1500 })

    const redirect = typeof route.query.redirect === 'string'
      ? decodeURIComponent(route.query.redirect)
      : '/'
    router.replace(redirect || '/')
  }
  catch (error: any) {
    console.error('platform login failed', error)
    messageService.error(resolveLoginErrorMessage(error))
  }
  finally {
    submitLoading.value = false
  }
}
</script>

<template>
  <div class="platform-login" :style="loginThemeStyle">
    <component
      :is="currentLoginComponent"
      v-if="loginThemeReady && currentLoginComponent"
      v-bind="dynamicLoginProps"
      @submit="onSubmit"
    />
    <div v-if="!loginThemeReady" class="login-theme-loading" />

    <div v-if="loginThemeReady && !currentLoginComponent" class="platform-login__lang-switch">
      <SelectLang />
    </div>

    <div v-if="loginThemeReady && !currentLoginComponent" class="platform-login__content">
      <section class="platform-login__hero">
        <div class="platform-login__hero-badge">
          {{ loginBrand.heroBadge }}
        </div>
        <h1 class="platform-login__hero-title">
          {{ loginBrand.heroTitle }}
        </h1>
        <p class="platform-login__hero-desc">
          {{ loginBrand.heroDescription }}
        </p>

        <div class="platform-login__hero-panels">
          <div class="hero-panel">
            <div class="hero-panel__label">
              管理范围
            </div>
            <div class="hero-panel__value">
              全部机构
            </div>
            <div class="hero-panel__desc">
              统一查看总部管辖下的机构状态与人员规模
            </div>
          </div>
          <div class="hero-panel">
            <div class="hero-panel__label">
              登录方式
            </div>
            <div class="hero-panel__value">
              总部账号
            </div>
            <div class="hero-panel__desc">
              使用总控或管理账号登录，不再混入校区端扫码/短信入口
            </div>
          </div>
        </div>
      </section>

      <section class="platform-login__form-wrap">
        <div class="platform-login__form-card">
          <div class="platform-login__form-top">
            <div class="platform-login__form-icon">
              <img v-if="loginBrand.logoUrl" :src="loginBrand.logoUrl" alt="logo">
              <SafetyCertificateOutlined v-else />
            </div>
            <div>
              <div class="platform-login__form-caption">
                {{ loginBrand.loginTitle }}
              </div>
              <h2 class="platform-login__form-title">
                账号登录
              </h2>
              <div class="platform-login__form-subtitle">
                {{ loginBrand.loginSubtitle }}
              </div>
            </div>
          </div>

          <a-form
            ref="formRef"
            :model="formState"
            :rules="rules"
            layout="vertical"
            autocomplete="off"
            @finish="onSubmit"
          >
            <a-form-item label="账号" name="username">
              <a-input
                v-model:value="formState.username"
                size="large"
                placeholder="请输入总部账号"
              >
                <template #prefix>
                  <UserOutlined />
                </template>
              </a-input>
            </a-form-item>

            <a-form-item label="密码" name="password">
              <a-input-password
                v-model:value="formState.password"
                size="large"
                placeholder="请输入登录密码"
              >
                <template #prefix>
                  <LockOutlined />
                </template>
              </a-input-password>
            </a-form-item>

            <a-button
              block
              type="primary"
              html-type="submit"
              size="large"
              class="platform-login__submit"
              :loading="submitLoading"
            >
              登录并进入总控台
            </a-button>
          </a-form>

          <div class="platform-login__form-footer">
            {{ loginBrand.copyright }}
          </div>
        </div>
      </section>
    </div>
  </div>
</template>

<style scoped>
.platform-login {
  position: relative;
  min-height: 100vh;
  overflow: hidden;
  background:
    linear-gradient(rgba(248, 251, 255, 0.88), rgba(255, 255, 255, 0.92)),
    var(--tenant-bg-image),
    radial-gradient(circle at top left, color-mix(in srgb, var(--tenant-primary) 18%, transparent) 0, transparent 32%),
    radial-gradient(circle at bottom right, rgba(15, 23, 42, 0.1) 0, transparent 28%),
    linear-gradient(135deg, #eef4ff 0%, #f8fbff 45%, #ffffff 100%);
  background-position: center;
  background-size: cover;
}

.login-theme-loading {
  min-height: 100vh;
  background: #f8fbff;
}

.platform-login::before,
.platform-login::after {
  position: absolute;
  border-radius: 999px;
  content: '';
  filter: blur(2px);
}

.platform-login::before {
  top: -180px;
  left: -120px;
  width: 420px;
  height: 420px;
  background: rgba(37, 99, 235, 0.08);
}

.platform-login::after {
  right: -120px;
  bottom: -140px;
  width: 360px;
  height: 360px;
  background: rgba(15, 23, 42, 0.06);
}

.platform-login__lang-switch {
  position: absolute;
  top: 20px;
  right: 24px;
  z-index: 3;
  padding: 4px 8px;
  border: 1px solid rgba(226, 232, 240, 0.9);
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.92);
  box-shadow: 0 10px 24px rgba(15, 23, 42, 0.06);
}

.platform-login__content {
  position: relative;
  z-index: 1;
  display: grid;
  grid-template-columns: minmax(0, 1.2fr) minmax(420px, 480px);
  gap: 48px;
  align-items: center;
  max-width: 1240px;
  min-height: 100vh;
  margin: 0 auto;
  padding: 48px 32px;
}

.platform-login__hero {
  padding-right: 12px;
}

.platform-login__hero-badge {
  display: inline-flex;
  align-items: center;
  height: 32px;
  padding: 0 14px;
  border: 1px solid color-mix(in srgb, var(--tenant-primary) 28%, white);
  border-radius: 999px;
  background: color-mix(in srgb, var(--tenant-primary) 10%, white);
  color: var(--tenant-primary);
  font-size: 13px;
  font-weight: 600;
}

.platform-login__hero-title {
  max-width: 620px;
  margin: 18px 0 0;
  color: #0f172a;
  font-size: 44px;
  line-height: 1.15;
  font-weight: 700;
}

.platform-login__hero-desc {
  max-width: 640px;
  margin-top: 18px;
  color: #475467;
  font-size: 16px;
  line-height: 1.8;
}

.platform-login__hero-panels {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px;
  margin-top: 28px;
}

.hero-panel {
  padding: 22px 22px 20px;
  border: 1px solid rgba(226, 232, 240, 0.9);
  border-radius: 22px;
  background: rgba(255, 255, 255, 0.84);
  box-shadow: 0 18px 40px rgba(15, 23, 42, 0.05);
  backdrop-filter: blur(10px);
}

.hero-panel__label {
  color: #64748b;
  font-size: 13px;
}

.hero-panel__value {
  margin-top: 12px;
  color: #111827;
  font-size: 26px;
  font-weight: 700;
}

.hero-panel__desc {
  margin-top: 8px;
  color: #667085;
  font-size: 13px;
  line-height: 1.7;
}

.platform-login__form-wrap {
  display: flex;
  justify-content: center;
}

.platform-login__form-card {
  width: 100%;
  padding: 30px 30px 24px;
  border: 1px solid rgba(226, 232, 240, 0.95);
  border-radius: 28px;
  background: rgba(255, 255, 255, 0.94);
  box-shadow: 0 28px 56px rgba(15, 23, 42, 0.1);
  backdrop-filter: blur(12px);
}

.platform-login__form-top {
  display: flex;
  align-items: center;
  gap: 14px;
  margin-bottom: 24px;
}

.platform-login__form-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 48px;
  height: 48px;
  border-radius: 16px;
  background: linear-gradient(135deg, var(--tenant-primary) 0%, color-mix(in srgb, var(--tenant-primary) 72%, white) 100%);
  color: #fff;
  font-size: 20px;
  box-shadow: 0 12px 24px color-mix(in srgb, var(--tenant-primary) 26%, transparent);
}

.platform-login__form-icon img {
  width: 30px;
  height: 30px;
  object-fit: contain;
}

.platform-login__form-subtitle {
  margin-top: 5px;
  color: #667085;
  font-size: 13px;
}

.platform-login__form-caption {
  color: #64748b;
  font-size: 13px;
}

.platform-login__form-title {
  margin: 4px 0 0;
  color: #0f172a;
  font-size: 28px;
  font-weight: 700;
}

.platform-login__submit {
  height: 48px;
  margin-top: 8px;
  border-radius: 12px;
  font-size: 15px;
  font-weight: 600;
}

.platform-login__form-footer {
  margin-top: 18px;
  color: #667085;
  font-size: 13px;
  line-height: 1.7;
}

:deep(.ant-form-item) {
  margin-bottom: 18px;
}

:deep(.ant-form-item-label > label) {
  color: #344054;
  font-weight: 600;
}

:deep(.ant-input-affix-wrapper),
:deep(.ant-input) {
  border-radius: 12px;
}


@media (max-width: 1024px) {
  .platform-login__content {
    grid-template-columns: 1fr;
    gap: 28px;
    padding-top: 88px;
    padding-bottom: 40px;
  }

  .platform-login__hero {
    padding-right: 0;
  }

  .platform-login__hero-title {
    max-width: none;
    font-size: 34px;
  }
}

@media (max-width: 640px) {
  .platform-login__content {
    padding-inline: 18px;
  }

  .platform-login__hero-panels {
    grid-template-columns: 1fr;
  }

  .platform-login__form-card {
    padding: 24px 20px 20px;
    border-radius: 22px;
  }
}
</style>
