<script setup>
import { ref } from 'vue'

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
        <a-tab-pane key="1" tab="课堂点评">
          <class-comment-list />
        </a-tab-pane>
        <a-tab-pane key="2" tab="课堂点评明细">
          <class-comment-details />
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
  }
}
</style>
