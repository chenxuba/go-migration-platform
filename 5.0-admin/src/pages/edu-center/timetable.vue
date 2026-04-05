<script setup>
import { SettingFilled } from '@ant-design/icons-vue'
import { ref } from 'vue'
import TeacherMatrixApiTimetable from '@/components/edu-center/timetable/teacher-matrix-api-timetable.vue'

const activeKey = ref('1')
</script>

<template>
  <div class="home">
    <div class="tabs">
      <a-tabs
        v-model:active-key="activeKey" animated :tab-bar-style="{
          'border-bottom-left-radius': '0px',
          'border-bottom-right-radius': '0px',
        }"
      >
        <a-tab-pane key="1" tab="智慧课表">
          <smart-timetable />
        </a-tab-pane>
        <a-tab-pane key="2" tab="时间课表">
          <time-timetable />
        </a-tab-pane>
        <a-tab-pane key="3" tab="冲突日程">
          <conflict-schedule />
        </a-tab-pane>
        <a-tab-pane key="4" tab="教师矩阵">
          <TeacherMatrixApiTimetable />
        </a-tab-pane>
      </a-tabs>
      <div class="absolute right-6 top-1.5 bg-#f6f7f8 h-8 px-4 flex-center rounded-4 text-#222 cursor-pointer font-500">
        <SettingFilled class="mr1 text-#06f" />课表展示配置
      </div>
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
