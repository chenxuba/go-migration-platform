<script setup>
// 按时间汇总
import summarizedOverTime from './summarizedOverTime.vue'
// 按渠道
import byChannel from './byChannel.vue'
// 按销售
import accordingToSales from './accordingToSales.vue'
// 按订单销售员
import byCreator from './byCreator.vue'

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
          按渠道
        </span>
      </template>
      <div class="tab-content">
        <byChannel />
      </div>
    </a-tab-pane>
    <a-tab-pane key="3">
      <template #tab>
        <span class="custom-tab">
          按销售
        </span>
      </template>
      <div class="tab-content">
        <accordingToSales />
      </div>
    </a-tab-pane>
    <a-tab-pane key="4">
      <template #tab>
        <span class="custom-tab">
          按创建人
        </span>
      </template>
      <div class="tab-content">
        <byCreator />
      </div>
    </a-tab-pane>
  </a-tabs>
</template>

<style scoped lang="less"></style>
