<script setup lang="ts">
import { LockOutlined, SafetyCertificateOutlined, UserOutlined } from '@ant-design/icons-vue'
import type { TenantLoginBrandConfig } from '~/api/common/login-theme'

defineProps<{
  brand: Required<TenantLoginBrandConfig>
  formState: { username: string; password: string }
  submitLoading: boolean
  rules: any
}>()
const emit = defineEmits(['submit'])
</script>

<template>
  <div class="center-card-login">
    <section class="center-card-login__card">
      <div class="center-card-login__icon">
        <img v-if="brand.logoUrl" :src="brand.logoUrl" alt="logo">
        <SafetyCertificateOutlined v-else />
      </div>
      <h1>{{ brand.loginTitle }}</h1>
      <p>{{ brand.loginSubtitle || brand.heroDescription }}</p>
      <a-form :model="formState" :rules="rules" layout="vertical" autocomplete="off" @finish="emit('submit')">
        <a-form-item label="账号" name="username">
          <a-input v-model:value="formState.username" size="large" placeholder="请输入账号">
            <template #prefix><UserOutlined /></template>
          </a-input>
        </a-form-item>
        <a-form-item label="密码" name="password">
          <a-input-password v-model:value="formState.password" size="large" placeholder="请输入登录密码">
            <template #prefix><LockOutlined /></template>
          </a-input-password>
        </a-form-item>
        <a-button block type="primary" html-type="submit" size="large" :loading="submitLoading">登录并进入后台</a-button>
      </a-form>
    </section>
  </div>
</template>

<style scoped lang="less">
.center-card-login {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  padding: 32px;
  background:
    linear-gradient(rgba(248, 251, 255, 0.86), rgba(255, 255, 255, 0.92)),
    var(--tenant-bg-image),
    radial-gradient(circle at top left, color-mix(in srgb, var(--tenant-primary) 18%, transparent), transparent 34%),
    #f8fbff;
  background-position: center;
  background-size: cover;
}

.center-card-login__card {
  width: min(460px, 100%);
  padding: 36px;
  border-radius: 32px;
  background: rgba(255, 255, 255, 0.96);
  box-shadow: 0 30px 80px rgba(15, 23, 42, 0.12);
  text-align: center;

  h1 {
    margin: 18px 0 8px;
    color: #0f172a;
    font-size: 30px;
  }

  p {
    margin: 0 0 24px;
    color: #64748b;
  }

  :deep(.ant-form) {
    text-align: left;
  }

  :deep(.ant-input-affix-wrapper),
  :deep(.ant-btn) {
    height: 46px;
    border-radius: 14px;
  }
}

.center-card-login__icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 58px;
  height: 58px;
  border-radius: 18px;
  background: linear-gradient(135deg, var(--tenant-primary), color-mix(in srgb, var(--tenant-primary) 70%, white));
  color: #fff;
  font-size: 24px;

  img {
    max-width: 38px;
    max-height: 38px;
    object-fit: contain;
  }
}
</style>
