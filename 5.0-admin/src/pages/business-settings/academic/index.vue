<script setup lang="ts">
import { ref } from 'vue'
import CourseCategory from '~/components/edu-center/course-list/course-category.vue'
import CourseSetting from '~/components/edu-center/course-list/course-setting.vue'
import CallNameSetting from '~/components/business-settings/callNameSetting.vue'
import ClassroomSettings from '~/components/business-settings/classroomSettings.vue'

const activeKey = ref('course')

function onTabChange(key: string) {
  activeKey.value = key
}
</script>

<template>
  <div class="academic-settings-page">
    <a-tabs :active-key="activeKey" class="academic-settings-page__tabs" @change="onTabChange">
      <a-tab-pane key="course" tab="课程设置">
        <div class="academic-settings-page__pane">
          <CourseSetting />
        </div>
      </a-tab-pane>
      <a-tab-pane key="category" tab="课程类别">
        <div class="academic-settings-page__pane">
          <CourseCategory />
        </div>
      </a-tab-pane>
      <a-tab-pane key="class" tab="班级设置">
        <div class="academic-settings-page__pane academic-settings-page__pane--placeholder">
          <a-empty description="页面建设中" />
        </div>
      </a-tab-pane>
      <a-tab-pane key="schedule" tab="排课设置">
        <div class="academic-settings-page__pane academic-settings-page__pane--placeholder">
          <a-empty description="页面建设中" />
        </div>
      </a-tab-pane>
      <a-tab-pane key="roll-call" tab="点名设置">
        <div class="academic-settings-page__pane">
          <CallNameSetting />
        </div>
      </a-tab-pane>
      <a-tab-pane key="appointment" tab="约课设置">
        <div class="academic-settings-page__pane academic-settings-page__pane--placeholder">
          <a-empty description="页面建设中" />
        </div>
      </a-tab-pane>
      <a-tab-pane key="classroom" tab="教室设置">
        <div class="academic-settings-page__pane academic-settings-page__pane--classroom">
          <ClassroomSettings />
        </div>
      </a-tab-pane>
    </a-tabs>
  </div>
</template>

<style scoped lang="less">
.academic-settings-page {
  width: 100%;
  min-height: calc(100vh - 120px);
  color: #666;
}

.academic-settings-page__tabs {
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

.academic-settings-page__pane {
  min-height: 480px;
  padding-top: 8px;
}

.academic-settings-page__pane--placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 360px;
  background: #fff;
  border-radius: 0 0 16px 16px;
}

.academic-settings-page__pane--classroom {
  min-height: calc(100vh - 200px);
  border-radius: 0 0 16px 16px;
  overflow: hidden;
}
</style>
