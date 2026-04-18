<!-- 财务数据 -->
<script setup>
// 财务状况
import financialOverview from './financialOverview/index.vue'
// 订单交易
import orderTransaction from './orderTransaction/index.vue'
// 账单流水
import billStatement from './billStatement/index.vue'
// 学费变动
import feeChange from './feeChange/index.vue'

const activeKey = ref('1')
</script>

<template>
  <div class="home">
    <div class="tabs">
      <a-tabs
        v-model:active-key="activeKey" :tab-bar-style="{
          'border-bottom-left-radius': '0px',
          'border-bottom-right-radius': '0px',
        }"
      >
        <!-- force-render 强制渲染 -->
        <a-tab-pane key="1" tab="财务状况">
          <financialOverview />
        </a-tab-pane>
        <a-tab-pane key="2" tab="订单交易">
          <orderTransaction />
        </a-tab-pane>
        <a-tab-pane key="3" tab="账单流水">
          <billStatement />
        </a-tab-pane>
        <a-tab-pane key="4" tab="学费变动">
          <feeChange />
        </a-tab-pane>
      </a-tabs>
    </div>
  </div>
</template>

<style scoped lang="less">
.home {
  color: #666;

  .tabs {
    width: 100%;
    border-radius: 10px;
    line-height: 40px;

    :deep(.ant-tabs-nav) {
      background: #fff;
      border-radius: 16px;
      margin: 0;
    }

    :deep(.ant-tabs-nav-wrap) {
      padding-left: 36px;
    }

    :deep(.ant-tabs-ink-bar) {
      text-align: center;
      height: 9px !important;
      background: transparent;
      bottom: 1px !important;

      &::after {
        position: absolute;
        top: 0;
        left: calc(50% - 12px);
        width: 24px !important;
        height: 4px !important;
        border-radius: 2px;
        background-color: var(--pro-ant-color-primary);
        content: "";
      }
    }

    .twoTab {
      padding: 0;

      :deep(.ant-tabs-nav-wrap) {
        padding-left: 10px;
        margin: 6px 0;
      }

      :deep(.ant-tabs-nav) {
        &::before {
          display: none;
        }
      }

      :deep(.ant-tabs-tab) {
        padding: 6px 14px !important;
      }

      :deep(.ant-tabs-tab-active) {
        background: #e6f0ff;
        border-radius: 8px;
      }

      :deep(.ant-tabs-ink-bar) {
        display: none;
      }

    }
  }
}
</style>
