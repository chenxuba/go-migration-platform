<script setup lang="ts">
import type { AccessCodeLike } from '~@/constants/access'

defineProps<{
  access: AccessCodeLike
}>()
const { hasAccess } = useAccess()
</script>

<template>
  <slot v-if="hasAccess(access)" />
</template>
