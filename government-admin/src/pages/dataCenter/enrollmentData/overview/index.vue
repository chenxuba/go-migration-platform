<script setup>
// 按时间汇总
import summarizedOverTime from './summarizedOverTime.vue'
// 按报读类型
import reportReadingType from './reportReadingType.vue'
// 按学员销售员
import traineeSalesperson from './traineeSalesperson.vue'
// 按订单销售员
import orderSalesperson from './orderSalesperson.vue'

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
          按报读类型
        </span>
      </template>
      <div class="tab-content">
        <reportReadingType />
      </div>
    </a-tab-pane>
    <a-tab-pane key="3">
      <template #tab>
        <span class="custom-tab">
          按学员销售员
        </span>
      </template>
      <div class="tab-content">
        <traineeSalesperson />
      </div>
    </a-tab-pane>
    <a-tab-pane key="4">
      <template #tab>
        <span class="custom-tab">
          按订单销售员
        </span>
      </template>
      <div class="tab-content">
        <orderSalesperson />
      </div>
    </a-tab-pane>
  </a-tabs>
</template>

<style scoped lang="less"></style>
