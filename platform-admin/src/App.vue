<script setup lang="ts">
import { computed, defineAsyncComponent, ref } from 'vue'

const route = useRoute()
const router = useRouter()
const AdminAppShell = defineAsyncComponent(() => import('~/components/app/admin-app-shell.vue'))
const shelllessRoutes = new Set(['/platform/login', '/platform/login-template-preview'])
const shelllessPrefixes = ['/platform/scales/pep3-iep-materials/import']
const routeReady = ref(false)
const isShelllessRoute = computed(() => {
  if (shelllessRoutes.has(route.path))
    return true
  return shelllessPrefixes.some(prefix => route.path === prefix || route.path.startsWith(`${prefix}/`))
})

void router.isReady().then(() => {
  routeReady.value = true
})
</script>

<template>
  <RouterView v-if="routeReady && isShelllessRoute" />
  <AdminAppShell v-else-if="routeReady" />
</template>

<style>
html,
body,
#app {
  overflow: auto;
}
</style>
