<script setup lang="ts">
const router = useRouter()
function back() {
  router.replace({
    path: '/platform/control-overview',
  })
}
</script>

<template>
  <a-result status="404" title="404" sub-title="Sorry, the page you visited does not exist.">
    <template #extra>
      <a-button type="primary" @click="back">
        Back Home
      </a-button>
    </template>
  </a-result>
</template>
