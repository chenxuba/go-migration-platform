<script setup lang="ts">
import { computed } from 'vue'
import { getLoginTemplateMeta, type LoginTemplateScope } from './login-template-registry'
import type { TenantLoginBrandConfig } from '@/api/platform/tenants'

const props = withDefaults(defineProps<{
  templateKey?: string
  scope?: LoginTemplateScope
  brand?: TenantLoginBrandConfig
  tenantName?: string
  compact?: boolean
}>(), {
  templateKey: '',
  scope: 'platform',
  tenantName: '',
  compact: false,
})

const meta = computed(() => getLoginTemplateMeta(props.templateKey))
const primaryColor = computed(() => props.brand?.primaryColor || '#1677ff')
const brandName = computed(() => props.brand?.brandName || props.tenantName || (props.scope === 'platform' ? '客户后台' : '机构中心'))
const loginTitle = computed(() => props.brand?.loginTitle || `${brandName.value}${props.scope === 'platform' ? '管理后台' : '机构端'}`)
const heroTitle = computed(() => props.brand?.heroTitle || `欢迎进入${brandName.value}`)
const heroDescription = computed(() => props.brand?.heroDescription || meta.value?.description || '按客户品牌、域名和权限独立呈现登录入口。')
const layout = computed(() => meta.value?.layout || 'split')
</script>

<template>
  <div
    class="login-template-preview"
    :class="[
      `login-template-preview--${layout}`,
      { 'login-template-preview--compact': compact },
    ]"
    :style="{ '--preview-primary': primaryColor }"
  >
    <div class="login-template-preview__canvas">
      <div class="login-template-preview__hero">
        <div class="login-template-preview__badge">{{ brandName }}</div>
        <strong>{{ heroTitle }}</strong>
        <p>{{ heroDescription }}</p>
      </div>
      <div class="login-template-preview__card">
        <div class="login-template-preview__logo">
          <img v-if="brand?.logoUrl" :src="brand.logoUrl" alt="logo">
          <span v-else>{{ brandName.slice(0, 1) }}</span>
        </div>
        <b>{{ loginTitle }}</b>
        <i />
        <i />
        <button type="button">登录</button>
      </div>
    </div>
    <div class="login-template-preview__meta">
      <strong>{{ meta?.label || '跟随租户默认' }}</strong>
      <span>{{ meta?.description || '使用租户机构端默认登录页模板' }}</span>
    </div>
  </div>
</template>

<style scoped lang="less">
.login-template-preview {
  --preview-primary: #1677ff;
  overflow: hidden;
  border: 1px solid #e8edf5;
  border-radius: 12px;
  background: #fff;
}

.login-template-preview__canvas {
  display: grid;
  grid-template-columns: 1fr 132px;
  gap: 12px;
  min-height: 112px;
  padding: 14px;
  background:
    radial-gradient(circle at 12% 8%, color-mix(in srgb, var(--preview-primary) 18%, transparent), transparent 28%),
    linear-gradient(135deg, #f8fbff 0%, #fff 64%);
}

.login-template-preview__hero {
  min-width: 0;

  .login-template-preview__badge {
    display: inline-flex;
    max-width: 120px;
    height: 20px;
    align-items: center;
    padding: 0 8px;
    overflow: hidden;
    border-radius: 999px;
    background: color-mix(in srgb, var(--preview-primary) 10%, #fff);
    color: var(--preview-primary);
    font-size: 11px;
    line-height: 20px;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  strong {
    display: block;
    margin-top: 8px;
    overflow: hidden;
    color: #1f2937;
    font-size: 16px;
    line-height: 22px;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  p {
    display: -webkit-box;
    margin: 5px 0 0;
    overflow: hidden;
    color: #667085;
    font-size: 12px;
    line-height: 18px;
    -webkit-box-orient: vertical;
    -webkit-line-clamp: 2;
  }
}

.login-template-preview__card {
  display: flex;
  flex-direction: column;
  gap: 7px;
  padding: 10px;
  border: 1px solid #eef2f7;
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.92);
  box-shadow: 0 12px 30px rgba(15, 23, 42, 0.08);

  b {
    overflow: hidden;
    color: #1f2937;
    font-size: 12px;
    font-style: normal;
    line-height: 16px;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  i {
    display: block;
    height: 12px;
    border-radius: 6px;
    background: #edf2f7;
  }

  button {
    height: 18px;
    border: 0;
    border-radius: 6px;
    background: var(--preview-primary);
    color: #fff;
    font-size: 10px;
  }
}

.login-template-preview__logo {
  display: inline-flex;
  width: 28px;
  height: 28px;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  border-radius: 9px;
  background: color-mix(in srgb, var(--preview-primary) 12%, #fff);
  color: var(--preview-primary);
  font-weight: 700;

  img {
    width: 100%;
    height: 100%;
    object-fit: contain;
  }
}

.login-template-preview__meta {
  display: flex;
  flex-direction: column;
  gap: 2px;
  padding: 10px 12px;

  strong {
    color: #1f2937;
    font-size: 13px;
    line-height: 18px;
  }

  span {
    overflow: hidden;
    color: #667085;
    font-size: 12px;
    line-height: 18px;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.login-template-preview--card {
  .login-template-preview__canvas {
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .login-template-preview__hero {
    display: none;
  }

  .login-template-preview__card {
    width: 156px;
  }
}

.login-template-preview--portal {
  .login-template-preview__canvas {
    grid-template-columns: 1fr 116px;
    background: linear-gradient(135deg, color-mix(in srgb, var(--preview-primary) 12%, #fff), #fff 62%);
  }

  .login-template-preview__card {
    box-shadow: none;
  }
}

.login-template-preview--compact {
  .login-template-preview__canvas {
    min-height: 84px;
    padding: 10px;
  }

  .login-template-preview__meta {
    padding: 8px 10px;
  }
}
</style>
