<!-- 订单交易 -->
<script setup>
// 按时间汇总
import summarizedOverTime from './summarizedOverTime.vue'
// 按商品类型
import byProductType from './byProductType.vue'
// 按课程类别
import byCourseCategory from './byCourseCategory.vue'
// 按经办人
import byOperator from './byOperator.vue'
// 按单课优惠活动
import bySingleCourseDiscount from './bySingleCourseDiscount.vue'
// 按整单优惠活动
import byTotalDiscount from './byTotalDiscount.vue'

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
          按商品类型
        </span>
      </template>
      <div class="tab-content">
        <byProductType />
      </div>
    </a-tab-pane>
    <a-tab-pane key="3">
      <template #tab>
        <span class="custom-tab">
          按课程类别
        </span>
      </template>
      <div class="tab-content">
        <byCourseCategory />
      </div>
    </a-tab-pane>
    <a-tab-pane key="4">
      <template #tab>
        <span class="custom-tab">
          按经办人
        </span>
      </template>
      <div class="tab-content">
        <byOperator />
      </div>
    </a-tab-pane>
    <a-tab-pane key="5">
      <template #tab>
        <span class="custom-tab">
          按单课优惠活动
        </span>
      </template>
      <div class="tab-content">
        <bySingleCourseDiscount />
      </div>
    </a-tab-pane>
    <a-tab-pane key="6">
      <template #tab>
        <span class="custom-tab">
          按整单优惠活动
        </span>
      </template>
      <div class="tab-content">
        <byTotalDiscount />
      </div>
    </a-tab-pane>
  </a-tabs>
</template>

<style scoped lang="less"></style>
