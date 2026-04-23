<script setup lang="ts">
import { ref } from 'vue'
import CourseSetting from '~/components/edu-center/course-list/course-setting.vue'
import CourseTeachingMethodSettings from './course-teaching-method-settings.vue'
import CourseBillingSettings from './course-billing-settings.vue'

const activeKey = ref('teaching')
</script>

<template>
  <div class="course-settings-tabs">
    <a-tabs
      v-model:active-key="activeKey"
      destroy-inactive-tab-pane
      class="course-settings-tabs__tabs"
      :tab-bar-style="{
        'border-top-left-radius': '0px',
        'border-top-right-radius': '0px',
        'border-bottom-left-radius': '16px',
        'border-bottom-right-radius': '16px',
      }"
    >
      <a-tab-pane key="teaching" tab="授课方式">
        <div class="course-settings-tabs__pane">
          <CourseTeachingMethodSettings />
        </div>
      </a-tab-pane>

      <a-tab-pane key="billing" tab="收费方式">
        <div class="course-settings-tabs__pane">
          <CourseBillingSettings />
        </div>
      </a-tab-pane>

      <a-tab-pane key="property" tab="课程属性自定义">
        <div class="course-settings-tabs__pane">
          <CourseSetting />
        </div>
      </a-tab-pane>
    </a-tabs>
  </div>
</template>

<style scoped lang="less">
.course-settings-tabs {
  padding: 0;
}

.course-settings-tabs__tabs {
  :deep(.ant-tabs-nav) {
    margin: 0;
    background: #fff;
    border-radius: 0 0 16px 16px !important;

    &::before {
      display: none;
    }
  }

  :deep(.ant-tabs-nav-wrap) {
    padding-left: 10px;
    margin: 6px 0;
  }

  :deep(.ant-tabs-tab) {
    margin: 0 8px 0 0;
    padding: 6px 14px !important;
    font-size: 14px !important;
  }

  :deep(.ant-tabs-tab .ant-tabs-tab-btn) {
    font-size: 14px !important;
    line-height: 22px;
  }

  :deep(.ant-tabs-tab-active) {
    background: #e6f0ff;
    border-radius: 8px;
  }

  :deep(.ant-tabs-tab-active .ant-tabs-tab-btn) {
    color: var(--pro-ant-color-primary, #1677ff);
    font-size: 14px !important;
    font-weight: 500;
  }

  :deep(.ant-tabs-ink-bar) {
    display: none;
  }

  :deep(.ant-tabs-content-holder) {
    background: transparent;
  }
}

.course-settings-tabs__pane {
  min-height: 420px;
  margin-top: 8px;
}

@media (max-width: 768px) {
  .course-settings-tabs__tabs {
    :deep(.ant-tabs-nav-wrap) {
      padding-left: 8px;
    }
  }
}
</style>
