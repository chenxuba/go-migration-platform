<script setup lang="ts">
import { computed, defineAsyncComponent, ref } from 'vue'

const route = useRoute()
const router = useRouter()
const AdminAppShell = defineAsyncComponent(() => import('~/components/app/admin-app-shell.vue'))
const shelllessRoutes = new Set(['/platform/login', '/platform/login-template-preview'])
const routeReady = ref(false)
const isShelllessRoute = computed(() => shelllessRoutes.has(route.path))

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
