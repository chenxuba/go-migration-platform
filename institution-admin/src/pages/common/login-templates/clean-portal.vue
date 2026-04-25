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
  <div class="clean-portal-login">
    <header class="clean-portal-login__header">
      <div class="clean-portal-login__brand">
        <div class="clean-portal-login__logo-card">
          <img v-if="brand.logoUrl" :src="brand.logoUrl" alt="logo">
          <strong v-else>{{ brand.brandName }}</strong>
        </div>
        <div class="clean-portal-login__brand-name">{{ brand.brandName }}</div>
      </div>
      <span>{{ brand.heroBadge || '机构端登录' }}</span>
    </header>

    <main class="clean-portal-login__main">
      <section class="clean-portal-login__hero">
        <div class="clean-portal-login__eyebrow">{{ brand.loginSubtitle || '独立机构入口' }}</div>
        <h1>{{ brand.heroTitle || brand.loginTitle }}</h1>
        <p>{{ brand.heroDescription }}</p>
      </section>

      <section class="clean-portal-login__panel">
        <div class="clean-portal-login__panel-title">
          <span>{{ brand.brandName }}</span>
          <h2>{{ brand.loginTitle || '机构端登录' }}</h2>
        </div>
        <a-segmented
          :value="activeKey"
          :options="[{ label: '密码登录', value: 0 }, { label: '短信登录', value: 1 }]"
          block
          @update:value="value => emit('update:activeKey', value)"
        />
        <a-input v-model:value="formState.username" size="large" placeholder="请输入手机号">
          <template #prefix><MobileOutlined /></template>
        </a-input>
        <a-input-password v-if="activeKey === 0" v-model:value="formState.password" size="large" placeholder="请输入密码">
          <template #prefix><LockOutlined /></template>
        </a-input-password>
        <a-input v-else v-model:value="formState.verifyCode" size="large" placeholder="请输入验证码">
          <template #prefix><SafetyOutlined /></template>
          <template #suffix><span class="clean-portal-login__code">获取验证码</span></template>
        </a-input>
        <a-alert v-if="errorAlert && loginErrorMessage" type="error" show-icon :message="loginErrorMessage" />
        <a-button block type="primary" size="large" :loading="submitLoading" @click="emit('submit')">登录机构后台</a-button>
        <a-checkbox :checked="agreeToTerms" @update:checked="value => emit('update:agreeToTerms', value)">
          同意《用户协议》与《隐私条款》
        </a-checkbox>
      </section>
    </main>

    <footer v-if="brand.copyright" class="clean-portal-login__footer">
      {{ brand.copyright }}
    </footer>
  </div>
</template>

<style scoped lang="less">
.clean-portal-login {
  min-height: 100vh;
  padding: 28px max(28px, calc((100vw - 1180px) / 2));
  background:
    linear-gradient(120deg, rgba(15, 23, 42, 0.84), rgba(15, 23, 42, 0.56)),
    var(--tenant-bg-image),
    radial-gradient(circle at 74% 18%, color-mix(in srgb, var(--tenant-primary) 26%, transparent), transparent 30%),
    #0f172a;
  background-position: center;
  background-size: cover;
  color: #fff;
}

.clean-portal-login__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-height: 48px;

  > span {
    padding: 7px 13px;
    border: 1px solid rgba(255, 255, 255, 0.18);
    border-radius: 999px;
    background: rgba(255, 255, 255, 0.1);
    color: rgba(255, 255, 255, 0.88);
  }
}

.clean-portal-login__brand {
  display: flex;
  align-items: center;
  gap: 12px;
}

.clean-portal-login__logo-card {
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 64px;
  max-width: 176px;
  height: 54px;
  padding: 8px 12px;
  border: 1px solid rgba(255, 255, 255, 0.28);
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.92);
  box-shadow: 0 16px 36px rgba(0, 0, 0, 0.18);
  backdrop-filter: blur(10px);

  img {
    max-width: 152px;
    max-height: 38px;
    object-fit: contain;
  }

  strong {
    color: #0f172a;
    font-size: 18px;
  }
}

.clean-portal-login__brand-name {
  max-width: 220px;
  color: rgba(255, 255, 255, 0.92);
  font-size: 18px;
  font-weight: 700;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.clean-portal-login__main {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 410px;
  gap: 64px;
  align-items: center;
  min-height: calc(100vh - 150px);
}

.clean-portal-login__hero {
  max-width: 650px;

  h1 {
    margin: 18px 0 0;
    font-size: 52px;
    line-height: 1.12;
    letter-spacing: -1px;
  }

  p {
    max-width: 590px;
    margin-top: 20px;
    color: rgba(255, 255, 255, 0.76);
    font-size: 17px;
    line-height: 1.9;
  }
}

.clean-portal-login__eyebrow {
  display: inline-flex;
  padding: 7px 13px;
  border-radius: 999px;
  background: color-mix(in srgb, var(--tenant-primary) 82%, white);
  color: #fff;
  font-weight: 600;
}

.clean-portal-login__panel {
  padding: 30px;
  border: 1px solid rgba(255, 255, 255, 0.15);
  border-radius: 28px;
  background: rgba(255, 255, 255, 0.96);
  box-shadow: 0 30px 80px rgba(0, 0, 0, 0.24);
  color: #111827;

  :deep(.ant-segmented) {
    margin-bottom: 14px;
    border-radius: 12px;
  }

  :deep(.ant-input-affix-wrapper) {
    display: flex;
    align-items: center;
    width: 100%;
    height: 46px;
    margin-bottom: 14px;
    padding-inline: 14px;
    border-radius: 12px;
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
    margin: 8px 0 14px;
    border-radius: 12px;
    font-weight: 600;
  }
}

.clean-portal-login__panel-title {
  margin-bottom: 22px;

  span {
    color: var(--tenant-primary);
    font-weight: 600;
  }

  h2 {
    margin: 8px 0 0;
    color: #0f172a;
    font-size: 28px;
  }
}

.clean-portal-login__code {
  color: var(--tenant-primary);
  cursor: pointer;
}

.clean-portal-login__footer {
  color: rgba(255, 255, 255, 0.6);
  font-size: 12px;
  text-align: center;
}

@media (max-width: 900px) {
  .clean-portal-login__main {
    grid-template-columns: 1fr;
    gap: 28px;
  }

  .clean-portal-login__hero h1 {
    font-size: 38px;
  }
}
</style>
