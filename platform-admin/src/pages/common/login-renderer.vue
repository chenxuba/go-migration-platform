<script setup lang="ts">
import { LockOutlined, SafetyCertificateOutlined, UserOutlined } from '@ant-design/icons-vue'
import { computed, defineAsyncComponent } from 'vue'
import type { TenantLoginBrandConfig } from '~/api/common/login-theme'

const SelectLang = defineAsyncComponent(() => import('@/components/select-lang/index.vue'))

const props = withDefaults(defineProps<{
  brand: Required<TenantLoginBrandConfig>
  formState: { username: string; password: string }
  submitLoading: boolean
  rules: any
  ready?: boolean
  showLang?: boolean
}>(), {
  ready: true,
  showLang: true,
})

const emit = defineEmits<{
  (e: 'submit'): void
}>()

const loginTemplateComponents: Record<string, any> = {
  'center-card': defineAsyncComponent(() => import('./login-templates/center-card.vue')),
  'minimal-portal': defineAsyncComponent(() => import('./login-templates/minimal-portal.vue')),
}

const loginThemeStyle = computed(() => ({
  '--tenant-primary': props.brand.primaryColor || '#1677ff',
  '--tenant-bg-image': props.brand.backgroundUrl ? `url(${props.brand.backgroundUrl})` : 'none',
}))
const currentLoginComponent = computed(() => loginTemplateComponents[props.brand.template] || null)
const dynamicLoginProps = computed(() => ({ brand: props.brand, formState: props.formState, submitLoading: props.submitLoading, rules: props.rules }))
</script>

<template>
  <div class="platform-login" :style="loginThemeStyle">
    <component
      :is="currentLoginComponent"
      v-if="ready && currentLoginComponent"
      v-bind="dynamicLoginProps"
      @submit="emit('submit')"
    />
    <div v-if="!ready" class="login-theme-loading" />

    <div v-if="ready && !currentLoginComponent && showLang" class="platform-login__lang-switch">
      <SelectLang />
    </div>

    <div v-if="ready && !currentLoginComponent" class="platform-login__content">
      <section class="platform-login__hero">
        <div class="platform-login__hero-badge">
          {{ brand.heroBadge }}
        </div>
        <h1 class="platform-login__hero-title">
          {{ brand.heroTitle }}
        </h1>
        <p class="platform-login__hero-desc">
          {{ brand.heroDescription }}
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
              <img v-if="brand.logoUrl" :src="brand.logoUrl" alt="logo">
              <SafetyCertificateOutlined v-else />
            </div>
            <div>
              <div class="platform-login__form-caption">
                {{ brand.loginTitle }}
              </div>
              <h2 class="platform-login__form-title">
                账号登录
              </h2>
              <div class="platform-login__form-subtitle">
                {{ brand.loginSubtitle }}
              </div>
            </div>
          </div>

          <a-form
            :model="formState"
            :rules="rules"
            layout="vertical"
            autocomplete="off"
            @finish="emit('submit')"
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
            {{ brand.copyright }}
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

.platform-login :deep(.ant-btn-primary) {
  background: var(--tenant-primary) !important;
  border-color: var(--tenant-primary) !important;
  box-shadow: 0 12px 24px color-mix(in srgb, var(--tenant-primary) 24%, transparent);
}

.platform-login :deep(.ant-btn-primary:not(:disabled):not(.ant-btn-disabled):hover),
.platform-login :deep(.ant-btn-primary:not(:disabled):not(.ant-btn-disabled):focus-visible) {
  background: color-mix(in srgb, var(--tenant-primary) 86%, white) !important;
  border-color: color-mix(in srgb, var(--tenant-primary) 86%, white) !important;
}

.platform-login :deep(.ant-btn-primary:not(:disabled):not(.ant-btn-disabled):active) {
  background: color-mix(in srgb, var(--tenant-primary) 90%, black) !important;
  border-color: color-mix(in srgb, var(--tenant-primary) 90%, black) !important;
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
