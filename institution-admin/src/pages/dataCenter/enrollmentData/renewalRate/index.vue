<script setup>
// 整体续费率
import overallRenewalRate from './overallRenewalRate.vue'
// 课程续费率
import courseRenewalRate from './courseRenewalRate.vue'
// 课程类别续费率
import courseCategoryRenewalRate from './courseCategoryRenewalRate.vue'
// 班级续费率
import classRenewalRate from './classRenewalRate.vue'
// 班主任续费率
import homeroomTeacherRenewalRate from './homeroomTeacherRenewalRate.vue'

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
          整体续费率
        </span>
      </template>
      <div class="tab-content">
        <overallRenewalRate />
      </div>
    </a-tab-pane>
    <a-tab-pane key="2">
      <template #tab>
        <span class="custom-tab">
          课程续费率
        </span>
      </template>
      <div class="tab-content">
        <courseRenewalRate />
      </div>
    </a-tab-pane>
    <a-tab-pane key="3">
      <template #tab>
        <span class="custom-tab">
          课程类别续费率
        </span>
      </template>
      <div class="tab-content">
        <courseCategoryRenewalRate />
      </div>
    </a-tab-pane>
    <a-tab-pane key="4">
      <template #tab>
        <span class="custom-tab">
          班级续费率
        </span>
      </template>
      <div class="tab-content">
        <classRenewalRate />
      </div>
    </a-tab-pane>
    <a-tab-pane key="5">
      <template #tab>
        <span class="custom-tab">
          班主任续费率
        </span>
      </template>
      <div class="tab-content">
        <homeroomTeacherRenewalRate />
      </div>
    </a-tab-pane>
  </a-tabs>
</template>

<style scoped lang="less"></style>
