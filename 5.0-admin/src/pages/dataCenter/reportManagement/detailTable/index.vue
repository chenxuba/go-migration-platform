<!-- 明细表 -->
<script setup>
// 订单明细表
import orderDetailTable from './orderDetailTable.vue'
// 上课记录明细表
import classRecordDetailTable from './classRecordDetailTable.vue'
// 学费变动记录明细表
import feeChangeDetailTable from './feeChangeDetailTable.vue'
// 分校学员报读总表
import branchReadTotalTable from './branchReadTotalTable.vue'

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
          订单明细表
        </span>
      </template>
      <div class="tab-content">
        <orderDetailTable />
      </div>
    </a-tab-pane>
    <a-tab-pane key="2">
      <template #tab>
        <span class="custom-tab">
          上课记录明细表
        </span>
      </template>
      <div class="tab-content">
        <classRecordDetailTable />
      </div>
    </a-tab-pane>
    <a-tab-pane key="3">
      <template #tab>
        <span class="custom-tab">
          学费变动记录明细表
        </span>
      </template>
      <div class="tab-content">
        <feeChangeDetailTable />
      </div>
    </a-tab-pane>
    <a-tab-pane key="4">
      <template #tab>
        <span class="custom-tab">
          分校学员报读总表
        </span>
      </template>
      <div class="tab-content">
        <branchReadTotalTable />
      </div>
    </a-tab-pane>
  </a-tabs>
</template>

<style scoped lang="less"></style>
