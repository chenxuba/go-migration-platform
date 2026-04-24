<script setup lang="ts">
import { LockOutlined, UserOutlined } from '@ant-design/icons-vue'
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
  <div class="minimal-portal-login">
    <section class="minimal-portal-login__copy">
      <div class="minimal-portal-login__badge">{{ brand.heroBadge || brand.brandName }}</div>
      <h1>{{ brand.heroTitle }}</h1>
      <p>{{ brand.heroDescription }}</p>
    </section>
    <section class="minimal-portal-login__form">
      <img v-if="brand.logoUrl" :src="brand.logoUrl" alt="logo">
      <strong v-else>{{ brand.brandName }}</strong>
      <span>{{ brand.loginSubtitle }}</span>
      <a-form :model="formState" :rules="rules" layout="vertical" autocomplete="off" @finish="emit('submit')">
        <a-form-item label="账号" name="username">
          <a-input v-model:value="formState.username" size="large" placeholder="请输入账号">
            <template #prefix><UserOutlined /></template>
          </a-input>
        </a-form-item>
        <a-form-item label="密码" name="password">
          <a-input-password v-model:value="formState.password" size="large" placeholder="请输入密码">
            <template #prefix><LockOutlined /></template>
          </a-input-password>
        </a-form-item>
        <a-button block type="primary" html-type="submit" size="large" :loading="submitLoading">进入管理后台</a-button>
      </a-form>
    </section>
  </div>
</template>

<style scoped lang="less">
.minimal-portal-login {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 430px;
  gap: 58px;
  align-items: center;
  min-height: 100vh;
  padding: 48px max(32px, calc((100vw - 1160px) / 2));
  background:
    linear-gradient(115deg, color-mix(in srgb, var(--tenant-primary) 18%, #0f172a) 0%, #0f172a 45%, #f8fafc 45%, #fff 100%),
    var(--tenant-bg-image);
  background-position: center;
  background-size: cover;
}

.minimal-portal-login__copy {
  color: #fff;

  h1 {
    max-width: 600px;
    margin: 18px 0 0;
    font-size: 46px;
    line-height: 1.15;
  }

  p {
    max-width: 620px;
    margin-top: 18px;
    color: rgba(255, 255, 255, 0.78);
    font-size: 16px;
    line-height: 1.8;
  }
}

.minimal-portal-login__badge {
  display: inline-flex;
  padding: 6px 13px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.12);
  font-weight: 600;
}

.minimal-portal-login__form {
  padding: 32px;
  border-radius: 28px;
  background: #fff;
  box-shadow: 0 30px 80px rgba(15, 23, 42, 0.14);

  img {
    max-width: 170px;
    max-height: 46px;
    object-fit: contain;
  }

  strong {
    display: block;
    color: #0f172a;
    font-size: 26px;
  }

  > span {
    display: block;
    margin: 6px 0 24px;
    color: #64748b;
  }

  :deep(.ant-input-affix-wrapper),
  :deep(.ant-btn) {
    height: 46px;
    border-radius: 12px;
  }
}
</style>
