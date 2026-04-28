<script setup lang="ts">
import { LockOutlined, SafetyCertificateOutlined, UserOutlined } from '@ant-design/icons-vue'
import type { Rule } from 'ant-design-vue/es/form'
import { reactive, ref } from 'vue'
import SelectLang from '@/components/select-lang/index.vue'
import { useAuthorization } from '@/composables/authorization'
import messageService from '@/utils/messageService'
import { loginApi } from '~/api/common/login'
import { hasGovernmentPortalAccess } from '~/utils/government-auth'
import { reset401Status } from '~/utils/request'

const { t } = useI18nLocale()
const router = useRouter()
const route = useRoute()
const token = useAuthorization()
const userStore = useUserStore()

const formRef = ref()
const submitLoading = ref(false)

const formState = reactive({
  username: '',
  password: '',
})

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
    return '当前账号未开通政府端权限，请联系总控管理员配置监管角色'
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
    reset401Status()

    const userInfo = await userStore.getUserInfo()
    if (!hasGovernmentPortalAccess(userInfo)) {
      await userStore.logout()
      messageService.error('当前账号未分配政府端权限，请使用监管账号登录')
      return
    }

    messageService.success(t('pages.login.notification.success.title', '登录成功'), { duration: 1500 })

    const redirect = typeof route.query.redirect === 'string'
      ? decodeURIComponent(route.query.redirect)
      : '/government/overview'
    router.replace(redirect || '/government/overview')
  }
  catch (error: any) {
    console.error('government login failed', error)
    messageService.error(resolveLoginErrorMessage(error))
  }
  finally {
    submitLoading.value = false
  }
}
</script>

<template>
  <div class="platform-login">
    <div class="platform-login__lang-switch">
      <SelectLang />
    </div>

    <div class="platform-login__content">
      <section class="platform-login__hero">
        <div class="platform-login__hero-badge">
          G 端监管
        </div>
        <h1 class="platform-login__hero-title">
          统一监管省市区三级康复机构
        </h1>
        <p class="platform-login__hero-desc">
          省级、市级、区级监管账号共用一套端口，账号统一从总控端账号体系创建，通过角色和行政区划范围切换数据视角，后续可继续承接巡查、整改、统计和机构备案等模块。
        </p>

        <div class="platform-login__hero-panels">
          <div class="hero-panel">
            <div class="hero-panel__label">
              管理范围
            </div>
            <div class="hero-panel__value">
              省 / 市 / 区
            </div>
            <div class="hero-panel__desc">
              同一套代码承载三级监管视图，按行政区范围控制数据
            </div>
          </div>
          <div class="hero-panel">
            <div class="hero-panel__label">
              登录方式
            </div>
            <div class="hero-panel__value">
              监管账号
            </div>
            <div class="hero-panel__desc">
              使用总控端创建的监管账号登录，不与机构端账号入口混用
            </div>
          </div>
        </div>
      </section>

      <section class="platform-login__form-wrap">
        <div class="platform-login__form-card">
          <div class="platform-login__form-top">
            <div class="platform-login__form-icon">
              <SafetyCertificateOutlined />
            </div>
            <div>
              <div class="platform-login__form-caption">
                康复机构监管平台
              </div>
              <h2 class="platform-login__form-title">
                账号登录
              </h2>
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
                placeholder="请输入总控端创建的监管账号"
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
              登录并进入监管台
            </a-button>
          </a-form>

          <div class="platform-login__form-footer">
            仅支持总控端创建并分配监管角色的省、市、区账号登录。
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
    radial-gradient(circle at top left, rgba(59, 130, 246, 0.18) 0, rgba(59, 130, 246, 0) 32%),
    radial-gradient(circle at bottom right, rgba(15, 23, 42, 0.1) 0, rgba(15, 23, 42, 0) 28%),
    linear-gradient(135deg, #eef4ff 0%, #f8fbff 45%, #ffffff 100%);
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
  border: 1px solid rgba(191, 219, 254, 0.9);
  border-radius: 999px;
  background: rgba(239, 246, 255, 0.92);
  color: #1d4ed8;
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
  background: linear-gradient(135deg, #1d4ed8 0%, #3b82f6 100%);
  color: #fff;
  font-size: 20px;
  box-shadow: 0 12px 24px rgba(37, 99, 235, 0.26);
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
