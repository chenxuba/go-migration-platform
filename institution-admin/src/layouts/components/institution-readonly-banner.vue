<script setup lang="ts">
const userStore = useUserStore()

const visible = computed(() => ['expired_readonly', 'warning'].includes(userStore.institutionStatus))
const bannerMessage = computed(() => {
  if (userStore.institutionStatus === 'warning') {
    return '当前机构授权将在30天内到期，请及时联系售后续期。当前系统功能可正常使用，避免到期后影响业务操作。'
  }
  return '当前机构授权已到期，系统已切换为只读模式。您仍可查看数据，但编辑、创建、删除等操作已关闭，如需恢复请联系售后。'
})
</script>

<template>
  <div v-if="visible" class="institution-readonly-banner">
    <a-alert
      banner
      show-icon
      type="warning"
      :message="bannerMessage"
    />
  </div>
</template>

<style scoped lang="less">
.institution-readonly-banner {
  margin: 12px 12px 0;

  :deep(.ant-alert) {
    border: none;
    border-radius: 14px;
    background:
      linear-gradient(90deg, rgba(255, 247, 230, 0.98), rgba(255, 251, 235, 0.98));
    box-shadow: 0 8px 22px rgba(217, 119, 6, 0.12);
  }

  :deep(.ant-alert-message) {
    color: #92400e;
    font-weight: 600;
  }
}
</style>
