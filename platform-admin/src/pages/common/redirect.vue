<script setup lang="ts">
const route = useRoute()
const router = useRouter()

function resolveRedirectPath() {
  const value = route.params.path
  const raw = Array.isArray(value) ? value.join('/') : String(value || '')
  if (!raw)
    return '/'

  const decoded = decodeURIComponent(raw)
  return decoded.startsWith('/') ? decoded : `/${decoded}`
}

router.replace(resolveRedirectPath())
</script>

<template>
  <div />
</template>
