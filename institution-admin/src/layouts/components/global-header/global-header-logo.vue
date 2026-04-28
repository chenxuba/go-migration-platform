<script setup lang="ts">
import InstitutionVersionChip from '../institution-version-chip.vue'
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
      <div v-if="!isMobile" class="ant-pro-brand-meta">
        <h1>{{ title }}</h1>
        <InstitutionVersionChip />
      </div>
    </a>
  </div>
</template>
