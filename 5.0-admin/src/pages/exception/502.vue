<script setup lang="ts">
const router = useRouter()
function back() {
  // 回首页
  router.replace({
    path: '/dashboard',
  })
}
function refresh() {
  location.reload()
}
</script>

<template>
  <a-result status="500" title="502" sub-title="系统维护中，请稍后重试">
    <template #extra>
      <!-- <a-button type="primary" @click="refresh">
        刷新页面
      </a-button> -->
      <a-button type="primary" @click="back">
        返回首页
      </a-button>
    </template>
  </a-result>
</template>
