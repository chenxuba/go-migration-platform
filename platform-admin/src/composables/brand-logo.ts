import type { Ref } from 'vue'
import { computed, ref, watch } from 'vue'

const DEFAULT_LOGO_URL = './logo.svg'

function normalizeBrandLogo(value?: string) {
  const logo = String(value || '').trim()
  if (!logo || logo === '/logo.svg')
    return DEFAULT_LOGO_URL
  return logo
}

export function useBrandLogo(logo: Ref<string | undefined>) {
  const primaryFailed = ref(false)
  const hideLogo = ref(false)
  const primaryLogo = computed(() => normalizeBrandLogo(logo.value))
  const logoSrc = computed(() => primaryFailed.value ? DEFAULT_LOGO_URL : primaryLogo.value)

  watch(primaryLogo, () => {
    primaryFailed.value = false
    hideLogo.value = false
  })

  function handleLogoError() {
    if (!primaryFailed.value && logoSrc.value !== DEFAULT_LOGO_URL) {
      primaryFailed.value = true
      return
    }
    hideLogo.value = true
  }

  return {
    logoSrc,
    hideLogo,
    handleLogoError,
  }
}
