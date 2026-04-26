<script setup lang="ts">
import { onMounted } from 'vue'
import { useLayoutMenuProvide } from '~/components/page-container/context'
import { getLoginThemeApi } from '~/api/common/login-theme'

const appStore = useAppStore()
const { theme } = storeToRefs(appStore)
const { antd } = useI18nLocale()
const layoutMenu = useLayoutMenu()
useLayoutMenuProvide(layoutMenu, appStore)

async function applyTenantTheme() {
  try {
    const res = await getLoginThemeApi('platform-admin')
    const primaryColor = res.result?.loginBrand?.primaryColor || res.data?.loginBrand?.primaryColor
    if (primaryColor)
      appStore.toggleColorPrimary(primaryColor)
  }
  catch (error) {
    console.warn('apply tenant theme failed', error)
  }
}

onMounted(applyTenantTheme)
</script>

<template>
  <a-config-provider :theme="theme" :locale="antd">
    <a-app class="h-full font-chinese antialiased">
      <TokenProvider>
        <RouterView />
      </TokenProvider>
    </a-app>
  </a-config-provider>
</template>

<style>
html,
body,
#app {
  overflow: auto;
}
</style>
