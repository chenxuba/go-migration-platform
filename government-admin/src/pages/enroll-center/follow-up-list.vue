<script setup>
import { ref } from 'vue'
import { QuestionCircleOutlined } from '@ant-design/icons-vue'

const activeKey = ref('1')
const activeKey2 = ref('1')
</script>

<template>
  <div class="tabs">
    <a-tabs
      v-model:active-key="activeKey2" animated :tab-bar-style="{
        'height': '46px',
        'border-bottom-left-radius': '0px',
        'border-bottom-right-radius': '0px',
      }" class="twoTab"
    >
      <a-tab-pane key="1">
        <template #tab>
          <span class="custom-tab">
            <a-tooltip
              :overlay-style="{
                maxWidth: '300px', // 最大宽度
                whiteSpace: 'normal', // 允许换行
              }"
            >
              <template #title>查看机构内所有的跟进记录</template>
              全部跟进记录
              <QuestionCircleOutlined />
            </a-tooltip>
          </span>
        </template>
        <div class="tab-content">
          <all-follow-up-list />
        </div>
      </a-tab-pane>
      <a-tab-pane key="2">
        <template #tab>
          <a-tooltip
            :overlay-style="{
              maxWidth: '300px', // 最大宽度
              whiteSpace: 'normal', // 允许换行
            }"
          >
            <template #title>
              查看所在部门及下级部门人员的跟进记录
            </template>
            部门跟进记录
            <QuestionCircleOutlined />
          </a-tooltip>
        </template>
        <div class="tab-content">
          <dpt-follow-up-list />
        </div>
      </a-tab-pane>
    </a-tabs>
  </div>
</template>

<style lang="less" scoped>
.tabs {
  width: 100%;
  border-radius: 10px;
  line-height: 40px;

  :deep(.ant-tabs-nav) {
    background: #fff;
    border-radius: 16px;
    margin: 0;
  }

  .twoTab {
    :deep(.ant-tabs-nav-wrap) {
      padding-left: 10px;
      margin: 6px 0;
    }

    :deep(.ant-tabs-tab) {
      padding: 6px 14px !important;
    }

    :deep(.ant-tabs-ink-bar) {
      text-align: center;
      height: 3px !important;
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
