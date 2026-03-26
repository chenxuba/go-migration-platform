<script setup>
// 按时间汇总
import summarizedOverTime from './summarizedOverTime.vue'
// 按付费类型
import accordingToPaymentType from './accordingToPaymentType.vue'
// 按渠道
import byChannel from './byChannel.vue'
// 按销售
import accordingToSales from './accordingToSales.vue'
// 按课程
import accordingToCourse from './accordingToCourse.vue'
// 按上课老师
import accordingToTeacher from './accordingToTeacher.vue'

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
          按付费类型
        </span>
      </template>
      <div class="tab-content">
        <accordingToPaymentType />
      </div>
    </a-tab-pane>
    <a-tab-pane key="3">
      <template #tab>
        <span class="custom-tab">
          按渠道
        </span>
      </template>
      <div class="tab-content">
        <byChannel />
      </div>
    </a-tab-pane>
    <a-tab-pane key="4">
      <template #tab>
        <span class="custom-tab">
          按销售
        </span>
      </template>
      <div class="tab-content">
        <accordingToSales />
      </div>
    </a-tab-pane>
    <a-tab-pane key="5">
      <template #tab>
        <span class="custom-tab">
          按课程
        </span>
      </template>
      <div class="tab-content">
        <accordingToCourse />
      </div>
    </a-tab-pane>
    <a-tab-pane key="6">
      <template #tab>
        <span class="custom-tab">
          按上课老师
        </span>
      </template>
      <div class="tab-content">
        <accordingToTeacher />
      </div>
    </a-tab-pane>
  </a-tabs>
</template>

<style scoped lang="less"></style>
