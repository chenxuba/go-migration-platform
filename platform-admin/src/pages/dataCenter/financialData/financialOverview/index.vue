<!-- 财务概况 -->
<script setup>
// 按时间汇总
import summarizedOverTime from './summarizedOverTime.vue'
// 剩余学费概况
import remainingFeeOverview from './remainingFeeOverview.vue'

const overview = ref('1')
</script>

<template>
  <a-tabs
    v-model:active-key="overview" animated :tab-bar-style="{
      'height': '46px',
      'border-top-left-radius': '0px',
      'border-top-right-radius': '0px',
    }" class="twoTab"
  >
    <a-tab-pane key="1">
      <template #tab>
        <span class="custom-tab">
          按时间汇总
        </span>
      </template>
      <div class="tab-content">
        <summarizedOverTime />
      </div>
    </a-tab-pane>
    <a-tab-pane key="2">
      <template #tab>
        <span class="custom-tab">
          剩余学费概况
        </span>
      </template>
      <div class="tab-content">
        <remainingFeeOverview />
      </div>
    </a-tab-pane>
  </a-tabs>
</template>

<style scoped lang="less"></style>
