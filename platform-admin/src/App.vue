<script setup lang="ts">
import { computed, defineAsyncComponent } from 'vue'

const route = useRoute()
const AdminAppShell = defineAsyncComponent(() => import('~/components/app/admin-app-shell.vue'))
const shelllessRoutes = new Set(['/login', '/login-template-preview'])
const isShelllessRoute = computed(() => shelllessRoutes.has(route.path))
</script>

<template>
  <RouterView v-if="isShelllessRoute" />
  <AdminAppShell v-else />
</template>

<style>
html,
body,
#app {
  overflow: auto;
}
</style>
