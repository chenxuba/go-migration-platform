<script setup lang="ts">
import { useLayoutState } from '../../basic-layout/context'
import { useBrandLogo } from '~@/composables/brand-logo'

const { logo, title, layout, isMobile } = useLayoutState()
const { logoSrc, hideLogo, handleLogoError } = useBrandLogo(logo)
const cls = computed(() => ({
  'ant-pro-global-header-logo': layout.value === 'mix' || isMobile.value,
  'ant-pro-top-nav-header-logo': layout.value === 'top' && !isMobile.value,
}))
</script>

<template>
  <div :class="cls">
    <a c-primary>
      <img v-if="!hideLogo" :src="logoSrc" @error="handleLogoError">
      <h1 v-if="!isMobile">{{ title }}</h1>
    </a>
  </div>
</template>
