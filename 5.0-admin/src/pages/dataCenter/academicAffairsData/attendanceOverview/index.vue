<!-- 出勤概况 -->
<script setup>
// 按时间汇总
import summarizedOverTime from './summarizedOverTime.vue'
// 按上课老师
import AsTeacherIsInstruction from './AsTeacherIsInstruction.vue'
// 按班主任
import AccordingToHeadTeacher from './AccordingToHeadTeacher.vue'
// 按班级/1对1
import ByClass_1on1 from './ByClass_1on1.vue'
// 按上课课程
import AccordingToClassSchedule from './AccordingToClassSchedule.vue'
// 按教室
import AccordingToClassroom from './AccordingToClassroom.vue'

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
          按时间汇总
        </span>
      </template>
      <div class="tab-content">
        <summarizedOverTime />
      </div>
    </a-tab-pane>
    <a-tab-pane key="2">
      <template #tab>
        <span class="custom-tab">
          按上课老师
        </span>
      </template>
      <div class="tab-content">
        <AsTeacherIsInstruction />
      </div>
    </a-tab-pane>
    <a-tab-pane key="3">
      <template #tab>
        <span class="custom-tab">
          按班主任
        </span>
      </template>
      <div class="tab-content">
        <AccordingToHeadTeacher />
      </div>
    </a-tab-pane>
    <a-tab-pane key="4">
      <template #tab>
        <span class="custom-tab">
          按班级/1对1
        </span>
      </template>
      <div class="tab-content">
        <ByClass_1on1 />
      </div>
    </a-tab-pane>
    <a-tab-pane key="5">
      <template #tab>
        <span class="custom-tab">
          按上课课程
        </span>
      </template>
      <div class="tab-content">
        <AccordingToClassSchedule />
      </div>
    </a-tab-pane>
    <a-tab-pane key="6">
      <template #tab>
        <span class="custom-tab">
          按教室
        </span>
      </template>
      <div class="tab-content">
        <AccordingToClassroom />
      </div>
    </a-tab-pane>
  </a-tabs>
</template>

<style scoped lang="less"></style>
