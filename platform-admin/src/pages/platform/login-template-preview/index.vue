<script setup lang="ts">
import type { Rule } from 'ant-design-vue/es/form'
import { computed, reactive } from 'vue'
import LoginRenderer from '@/pages/common/login-renderer.vue'
import type { TenantLoginBrandConfig } from '@/api/platform/tenants'

const route = useRoute()

const templateKey = computed(() => String(route.query.template || 'business-split'))
const entryType = computed(() => String(route.query.entryType || 'platform-admin'))
const formState = reactive({ username: 'demo_admin', password: '123456' })
const rules: Record<string, Rule[]> = {
  username: [{ required: true, message: '请输入账号', trigger: 'change' }],
  password: [{ required: true, message: '请输入密码', trigger: 'change' }],
}
const brand = computed<Required<TenantLoginBrandConfig>>(() => {
  const name = String(route.query.name || (entryType.value === 'platform-admin' ? '客户管理后台' : '机构管理端'))
  const isInstitution = entryType.value === 'institution-admin'
  return {
    template: templateKey.value,
    brandName: name,
    logoUrl: '',
    loginTitle: name,
    loginSubtitle: '请输入账号密码登录',
    backgroundUrl: '',
    primaryColor: isInstitution ? '#13ad74' : '#1677ff',
    copyright: '请使用对应账号从对应入口登录，跨域名禁止登录',
    heroBadge: isInstitution ? '机构端后台' : '子总控后台',
    heroTitle: `欢迎进入${name}`,
    heroDescription: String(route.query.desc || '这是登录页模板真实预览，使用与线上登录页一致的组件和样式。'),
  }
})

function noopSubmit() {}
</script>

<template>
  <LoginRenderer
    :brand="brand"
    :form-state="formState"
    :submit-loading="false"
    :rules="rules"
    :ready="true"
    :show-lang="false"
    @submit="noopSubmit"
  />
</template>
