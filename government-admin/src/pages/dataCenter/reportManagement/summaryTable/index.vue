<!-- 汇总表 -->
<script setup>
// 招生数据报表
import enrollmentDataReport from './enrollmentDataReport.vue'
// 满班/出勤报表
import fullClassAttendanceReport from './fullClassAttendanceReport.vue'
// 剩余学费报表
import remainingFeeReport from './remainingFeeReport.vue'
// 学费消耗报表
import feeConsumptionReport from './feeConsumptionReport.vue'
// 教材数据报表
import textbookDataReport from './textbookDataReport.vue'
// 班级数据报表
import classDataReport from './classDataReport.vue'
// 课时汇总数据报表
import courseSummaryDataReport from './courseSummaryDataReport.vue'

const overview = ref('1')
</script>

<template>
  <a-tabs
    v-model:active-key="overview" animated :tab-bar-style="{
      'height': '46px',
      'border-radius': '0px',
    }" class="twoTab"
  >
    <a-tab-pane key="1">
      <template #tab>
        <span class="custom-tab">
          招生数据报表
        </span>
      </template>
      <div class="tab-content">
        <enrollmentDataReport />
      </div>
    </a-tab-pane>
    <a-tab-pane key="2">
      <template #tab>
        <span class="custom-tab">
          满班/出勤报表
        </span>
      </template>
      <div class="tab-content">
        <fullClassAttendanceReport />
      </div>
    </a-tab-pane>
    <a-tab-pane key="3">
      <template #tab>
        <span class="custom-tab">
          剩余学费报表
        </span>
      </template>
      <div class="tab-content">
        <remainingFeeReport />
      </div>
    </a-tab-pane>
    <a-tab-pane key="4">
      <template #tab>
        <span class="custom-tab">
          学费消耗报表
        </span>
      </template>
      <div class="tab-content">
        <feeConsumptionReport />
      </div>
    </a-tab-pane>
    <a-tab-pane key="5">
      <template #tab>
        <span class="custom-tab">
          教材数据报表
        </span>
      </template>
      <div class="tab-content">
        <textbookDataReport />
      </div>
    </a-tab-pane>
    <a-tab-pane key="6">
      <template #tab>
        <span class="custom-tab">
          班级数据报表
        </span>
      </template>
      <div class="tab-content">
        <classDataReport />
      </div>
    </a-tab-pane>
    <a-tab-pane key="7">
      <template #tab>
        <span class="custom-tab">
          课时汇总数据报表
        </span>
      </template>
      <div class="tab-content">
        <courseSummaryDataReport />
      </div>
    </a-tab-pane>
  </a-tabs>
</template>

<style scoped lang="less"></style>
