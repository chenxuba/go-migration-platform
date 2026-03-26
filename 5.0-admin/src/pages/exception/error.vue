<script setup lang="ts">
const router = useRouter()
function back() {
  router.replace({
    path: '/dashboard',
  })
}
function refresh() {
  location.reload()
}
</script>

<template>
  <a-result status="404" title="404" sub-title="抱歉，您访问的页面不存在。">
    <template #extra>
       <a-button type="primary" @click="refresh">
        刷新页面
      </a-button>
      <a-button type="primary" @click="back">
        返回首页
      </a-button>
    </template>
  </a-result>
</template>
