<script setup>
import { LockOutlined, MobileOutlined, SafetyOutlined } from '@ant-design/icons-vue'

const props = defineProps({
  brand: { type: Object, required: true },
  formState: { type: Object, required: true },
  activeKey: { type: Number, required: true },
  agreeToTerms: { type: Boolean, required: true },
  submitLoading: { type: Boolean, required: true },
  errorAlert: { type: Boolean, required: true },
  loginErrorMessage: { type: String, default: '' },
  t: { type: Function, required: true },
})
const emit = defineEmits(['update:activeKey', 'update:agreeToTerms', 'submit'])
</script>

<template>
  <div class="campus-card-login">
    <div class="campus-card-login__shell">
      <section class="campus-card-login__visual">
        <div class="campus-card-login__topline">
          <div class="campus-card-login__logo-card">
            <img v-if="brand.logoUrl" :src="brand.logoUrl" alt="logo">
            <strong v-else>{{ brand.brandName }}</strong>
          </div>
          <span>{{ brand.heroBadge || '机构端' }}</span>
        </div>
        <div class="campus-card-login__copy">
          <h1>{{ brand.heroTitle || brand.loginTitle }}</h1>
          <p>{{ brand.heroDescription || brand.loginSubtitle }}</p>
        </div>
        <div class="campus-card-login__stats">
          <div>
            <b>独立域名</b>
            <span>租户专属入口</span>
          </div>
          <div>
            <b>机构隔离</b>
            <span>账号仅进本租户</span>
          </div>
        </div>
      </section>

      <section class="campus-card-login__panel">
        <div class="campus-card-login__panel-head">
          <span>{{ brand.brandName }}</span>
          <h2>{{ brand.loginTitle || '机构端登录' }}</h2>
          <p>{{ brand.loginSubtitle || '请输入机构账号登录' }}</p>
        </div>

        <a-segmented
          :value="activeKey"
          :options="[{ label: '密码登录', value: 0 }, { label: '短信登录', value: 1 }]"
          block
          class="campus-card-login__tabs"
          @update:value="value => emit('update:activeKey', value)"
        />

        <a-input v-model:value="formState.username" size="large" :placeholder="t('pages.login.mobile.placeholder', '请输入手机号')">
          <template #prefix><MobileOutlined /></template>
        </a-input>
        <a-input-password v-if="activeKey === 0" v-model:value="formState.password" size="large" :placeholder="t('pages.login.password.placeholder.simple', '请输入密码')">
          <template #prefix><LockOutlined /></template>
        </a-input-password>
        <a-input v-else v-model:value="formState.verifyCode" size="large" :placeholder="t('pages.login.verifyCode.placeholder', '请输入验证码')">
          <template #prefix><SafetyOutlined /></template>
          <template #suffix><span class="campus-card-login__code">{{ t('pages.login.verifyCode.action', '获取验证码') }}</span></template>
        </a-input>
        <a-alert v-if="errorAlert && loginErrorMessage" type="error" show-icon :message="loginErrorMessage" />
        <a-button block type="primary" size="large" :loading="submitLoading" @click="emit('submit')">
          {{ t('pages.login.submit.immediately', '立即登录') }}
        </a-button>
        <a-checkbox :checked="agreeToTerms" class="campus-card-login__agreement" @update:checked="value => emit('update:agreeToTerms', value)">
          <span>已阅读并同意《用户协议》与《隐私条款》</span>
        </a-checkbox>
        <div v-if="brand.copyright" class="campus-card-login__copyright">
          {{ brand.copyright }}
        </div>
      </section>
    </div>
  </div>
</template>

<style scoped lang="less">
.campus-card-login {
  min-height: 100vh;
  padding: 34px;
  background:
    linear-gradient(120deg, rgba(246, 251, 255, 0.94), rgba(255, 255, 255, 0.82)),
    var(--tenant-bg-image),
    radial-gradient(circle at 12% 10%, color-mix(in srgb, var(--tenant-primary) 20%, transparent), transparent 30%),
    #eef5ff;
  background-position: center;
  background-size: cover;
}

.campus-card-login__shell {
  display: grid;
  grid-template-columns: minmax(0, 1.05fr) 430px;
  gap: 0;
  width: min(1180px, 100%);
  min-height: calc(100vh - 68px);
  margin: 0 auto;
  overflow: hidden;
  border: 1px solid rgba(255, 255, 255, 0.78);
  border-radius: 34px;
  background: rgba(255, 255, 255, 0.72);
  box-shadow: 0 32px 90px rgba(15, 23, 42, 0.12);
  backdrop-filter: blur(16px);
}

.campus-card-login__visual {
  position: relative;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  padding: 34px 42px;
  background:
    linear-gradient(145deg, color-mix(in srgb, var(--tenant-primary) 18%, #ffffff), rgba(255, 255, 255, 0.92)),
    radial-gradient(circle at 18% 16%, color-mix(in srgb, var(--tenant-primary) 22%, transparent), transparent 26%);
}

.campus-card-login__visual::after {
  position: absolute;
  right: -90px;
  bottom: -90px;
  width: 360px;
  height: 360px;
  border-radius: 999px;
  background: color-mix(in srgb, var(--tenant-primary) 14%, transparent);
  content: '';
}

.campus-card-login__topline {
  position: relative;
  z-index: 1;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18px;

  span {
    padding: 6px 12px;
    border-radius: 999px;
    background: #fff;
    color: var(--tenant-primary);
    font-weight: 600;
    box-shadow: 0 10px 24px rgba(15, 23, 42, 0.08);
  }
}

.campus-card-login__logo-card {
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 72px;
  max-width: 190px;
  height: 54px;
  padding: 8px 12px;
  border: 1px solid rgba(255, 255, 255, 0.72);
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.82);
  box-shadow: 0 16px 36px rgba(15, 23, 42, 0.1);
  backdrop-filter: blur(8px);

  img {
    max-width: 166px;
    max-height: 38px;
    object-fit: contain;
  }

  strong {
    color: #0f172a;
    font-size: 19px;
  }
}

.campus-card-login__copy {
  position: relative;
  z-index: 1;
  max-width: 560px;

  h1 {
    margin: 0;
    color: #0f172a;
    font-size: 44px;
    line-height: 1.14;
  }

  p {
    margin: 18px 0 0;
    color: #475467;
    font-size: 16px;
    line-height: 1.9;
  }
}

.campus-card-login__stats {
  position: relative;
  z-index: 1;
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14px;

  div {
    padding: 16px;
    border: 1px solid rgba(255, 255, 255, 0.7);
    border-radius: 18px;
    background: rgba(255, 255, 255, 0.68);
  }

  b,
  span {
    display: block;
  }

  b {
    color: #0f172a;
    font-size: 16px;
  }

  span {
    margin-top: 5px;
    color: #667085;
  }
}

.campus-card-login__panel {
  display: flex;
  flex-direction: column;
  justify-content: center;
  padding: 44px 40px;
  background: #fff;
}

.campus-card-login__panel-head {
  margin-bottom: 24px;

  span {
    color: var(--tenant-primary);
    font-weight: 600;
  }

  h2 {
    margin: 8px 0 6px;
    color: #111827;
    font-size: 30px;
  }

  p {
    margin: 0;
    color: #667085;
  }
}

.campus-card-login__tabs {
  margin-bottom: 18px;
}

.campus-card-login__panel {
  :deep(.ant-input-affix-wrapper) {
    display: flex;
    align-items: center;
    width: 100%;
    height: 46px;
    margin-bottom: 14px;
    padding-inline: 14px;
    border-radius: 14px;
  }

  :deep(.ant-input-affix-wrapper .ant-input) {
    height: auto;
    margin: 0;
    padding: 0;
    border: 0;
    border-radius: 0;
    background: transparent;
    box-shadow: none;
  }

  :deep(.ant-btn) {
    height: 46px;
    margin: 10px 0 14px;
    border-radius: 14px;
    font-weight: 600;
  }
}

.campus-card-login__agreement {
  color: #667085;
}

.campus-card-login__code {
  color: var(--tenant-primary);
  cursor: pointer;
}

.campus-card-login__copyright {
  margin-top: 22px;
  color: #98a2b3;
  font-size: 12px;
  text-align: center;
}

@media (max-width: 900px) {
  .campus-card-login {
    padding: 18px;
  }

  .campus-card-login__shell {
    grid-template-columns: 1fr;
  }

  .campus-card-login__visual {
    min-height: 320px;
  }
}
</style>
