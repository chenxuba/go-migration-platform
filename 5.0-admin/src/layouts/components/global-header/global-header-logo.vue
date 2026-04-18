<script setup lang="ts">
import InstitutionVersionChip from '../institution-version-chip.vue'
import { useLayoutState } from '../../basic-layout/context'

const { logo, title, layout, isMobile } = useLayoutState()
const cls = computed(() => ({
  'ant-pro-global-header-logo': layout.value === 'mix' || isMobile.value,
  'ant-pro-top-nav-header-logo': layout.value === 'top' && !isMobile.value,
}))
</script>

<template>
  <div :class="cls">
    <a c-primary>
      <img :src="logo">
      <div v-if="!isMobile" class="ant-pro-brand-meta">
        <h1>{{ title }}</h1>
        <InstitutionVersionChip />
      </div>
    </a>
  </div>
</template>
