<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()
const title = computed(() => String(route.meta?.title ?? '业务设置'))
</script>

<template>
  <div class="business-settings-placeholder">
    <h2 class="business-settings-placeholder__title">
      {{ title }}
    </h2>
    <p class="business-settings-placeholder__desc">
      页面建设中，后续将在此配置相关业务项。
    </p>
    <a-empty description="敬请期待" />
  </div>
</template>

<style scoped lang="less">
.business-settings-placeholder {
  padding: 24px;
  min-height: 280px;
  background: #fff;
  border-radius: 12px;
}

.business-settings-placeholder__title {
  margin: 0 0 8px;
  font-size: 18px;
  font-weight: 600;
  color: #1f2329;
}

.business-settings-placeholder__desc {
  margin: 0 0 24px;
  font-size: 14px;
  color: #8c8c8c;
}
</style>
