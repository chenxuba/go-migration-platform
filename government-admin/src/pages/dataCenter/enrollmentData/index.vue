<!-- 招生数据 -->
<script setup>
// 招生概述
import overview from './overview/index.vue'
// 学员跟进
import studentFollowUP from './studentFollowUp/index.vue'
// 试听转化
import trialConversion from './trialConversion/index.vue'
// 业绩排行
import performanceRanking from './performanceRanking/index.vue'
// 续费率
import renewalRate from './renewalRate/index.vue'
// 机构流失率
import campusAttritionRate from './campusAttritionRate/index.vue'

const activeKey = ref('1')

const followUp = ref('1')
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
        <a-tab-pane key="1" tab="招生概述">
          <overview />
        </a-tab-pane>
        <a-tab-pane key="2" tab="学员跟进">
          <studentFollowUP />
        </a-tab-pane>
        <a-tab-pane key="3" tab="试听转化">
          <trialConversion />
        </a-tab-pane>
        <a-tab-pane key="4" tab="业绩排行">
          <performanceRanking />
        </a-tab-pane>
        <a-tab-pane key="5" tab="续费率">
          <renewalRate />
        </a-tab-pane>
        <a-tab-pane key="6" tab="机构流失率">
          <campusAttritionRate />
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
