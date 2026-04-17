<!-- 续费变动 -->
<script setup>
// 按时间汇总
import summarizedOverTime from './summarizedOverTime.vue'
// 按变动类型
import remainingFeeOverview from './remainingFeeOverview.vue'
// 按课程
import courseFeeChange from './courseFeeChange.vue'
// 按上课老师
import teacherFeeChange from './teacherFeeChange.vue'

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
          按变动类型
        </span>
      </template>
      <div class="tab-content">
        <remainingFeeOverview />
      </div>
    </a-tab-pane>
    <a-tab-pane key="3">
      <template #tab>
        <span class="custom-tab">
          按课程
        </span>
      </template>
      <div class="tab-content">
        <courseFeeChange />
      </div>
    </a-tab-pane>
    <a-tab-pane key="4">
      <template #tab>
        <span class="custom-tab">
          按上课老师
        </span>
      </template>
      <div class="tab-content">
        <teacherFeeChange />
      </div>
    </a-tab-pane>
  </a-tabs>
</template>

<style scoped lang="less"></style>
