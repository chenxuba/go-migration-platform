<script setup lang="ts">
import { computed, defineAsyncComponent } from 'vue'

const props = defineProps<{
  scaleCode: string
  scaleVersion?: string
}>()

const componentMap = {
  PEP3: defineAsyncComponent(() => import('./pep3-question-bank.vue')),
  SHUANGXI_A: defineAsyncComponent(() => import('./pep3-question-bank.vue')),
}

const normalizedCode = computed(() => String(props.scaleCode || '').trim().toUpperCase())
const activeComponent = computed(() => componentMap[normalizedCode.value as keyof typeof componentMap])
</script>

<template>
  <component
    :is="activeComponent"
    v-if="activeComponent"
    :scale-code="normalizedCode"
    :scale-version="scaleVersion"
  />
  <a-empty v-else class="question-bank-loader__empty" description="该量表题库管理暂未接入" />
</template>

<style scoped lang="less">
.question-bank-loader__empty {
  margin-top: 80px;
}
</style>
