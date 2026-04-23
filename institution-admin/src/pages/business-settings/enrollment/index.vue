<script setup lang="ts">
import { ref } from 'vue'
import ChannelManagement from '~/components/enroll-center/intention-student/channel-management.vue'
import IntentionStudentSetting from '~/components/enroll-center/intention-student/setting.vue'
import StudentCode from '~/components/edu-center/student-list/student-code.vue'
import StudentLinkSetting from '~/components/edu-center/student-list/student-link-setting.vue'

const activeKey = ref('student')
const studentSettingKey = ref('attribute')
</script>

<template>
  <div class="enrollment-settings-page">
    <a-tabs v-model:active-key="activeKey" destroy-inactive-tab-pane class="enrollment-settings-page__tabs">
      <a-tab-pane key="student" tab="学员设置">
        <div class="enrollment-settings-page__pane">
          <a-tabs
            v-model:active-key="studentSettingKey"
            destroy-inactive-tab-pane
            class="enrollment-settings-page__student-tabs"
          >
            <a-tab-pane key="attribute" tab="学员属性自定义">
              <StudentCode />
            </a-tab-pane>
            <a-tab-pane key="relation" tab="学员关联人员设置">
              <StudentLinkSetting />
            </a-tab-pane>
          </a-tabs>
        </div>
      </a-tab-pane>

      <a-tab-pane key="rules" tab="招生规则设置">
        <div class="enrollment-settings-page__pane enrollment-settings-page__pane--rules">
          <IntentionStudentSetting />
        </div>
      </a-tab-pane>

      <a-tab-pane key="channel" tab="渠道管理">
        <div class="enrollment-settings-page__pane">
          <ChannelManagement />
        </div>
      </a-tab-pane>
    </a-tabs>
  </div>
</template>

<style scoped lang="less">
.enrollment-settings-page {
  width: 100%;
  min-height: calc(100vh - 120px);
  color: #666;
}

.enrollment-settings-page__tabs {
  :deep(.ant-tabs-nav) {
    margin: 0;
    padding: 0 12px;
    background: #fff;
    border-radius: 16px 16px 0 0;
  }

  :deep(.ant-tabs-nav-wrap) {
    padding-left: 24px;
  }

  :deep(.ant-tabs-tab) {
    padding: 12px 0;
    font-size: 15px;
  }

  :deep(.ant-tabs-ink-bar) {
    height: 9px !important;
    background: transparent !important;
    bottom: 1px !important;

    &::after {
      position: absolute;
      top: 0;
      left: calc(50% - 12px);
      width: 24px !important;
      height: 4px !important;
      border-radius: 2px;
      background-color: var(--pro-ant-color-primary, #1677ff);
      content: '';
    }
  }

  :deep(.ant-tabs-content-holder) {
    background: transparent;
  }
}

.enrollment-settings-page__pane {
  min-height: 480px;
  padding-top: 8px;
}

.enrollment-settings-page__student-tabs {
  :deep(.ant-tabs-nav) {
    margin: 0;
    padding: 0 12px;
    background: #fff;
    border-radius: 14px 14px 0 0;
  }

  :deep(.ant-tabs-nav-wrap) {
    padding-left: 24px;
  }

  :deep(.ant-tabs-tab) {
    padding: 12px 0;
    font-size: 14px;
  }

  :deep(.ant-tabs-ink-bar) {
    height: 8px !important;
    background: transparent !important;
    bottom: 1px !important;

    &::after {
      position: absolute;
      top: 0;
      left: calc(50% - 10px);
      width: 20px !important;
      height: 4px !important;
      border-radius: 2px;
      background-color: var(--pro-ant-color-primary, #1677ff);
      content: '';
    }
  }

  :deep(.ant-tabs-content-holder) {
    background: transparent;
  }
}

@media (max-width: 768px) {
  .enrollment-settings-page__tabs {
    :deep(.ant-tabs-nav-wrap) {
      padding-left: 12px;
    }
  }

  .enrollment-settings-page__student-tabs {
    :deep(.ant-tabs-nav-wrap) {
      padding-left: 12px;
    }
  }
}
</style>
