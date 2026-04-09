<script setup>
import { DeleteOutlined } from '@ant-design/icons-vue'
import { Modal } from 'ant-design-vue'
import { computed, ref } from 'vue'
import TeacherMatrixApiTimetable from '@/components/edu-center/timetable/teacher-matrix-api-timetable.vue'
import { clearWeekTeachingSchedulesApi } from '@/api/edu-center/teaching-schedule'
import emitter, { EVENTS } from '@/utils/eventBus'
import messageService from '@/utils/messageService'

const activeKey = ref('1')
const clearingWeek = ref(false)
const weekRanges = ref({
  1: { startDate: '', endDate: '' },
  2: { startDate: '', endDate: '' },
  4: { startDate: '', endDate: '' },
})

const currentWeekRange = computed(() => weekRanges.value[activeKey.value] || { startDate: '', endDate: '' })
const canClearCurrentWeek = computed(() =>
  activeKey.value !== '3'
  && !!currentWeekRange.value.startDate
  && !!currentWeekRange.value.endDate
  && !clearingWeek.value,
)

function updateWeekRange(tabKey, value) {
  weekRanges.value = {
    ...weekRanges.value,
    [tabKey]: {
      startDate: String(value?.startDate || ''),
      endDate: String(value?.endDate || ''),
    },
  }
}

function handleClearCurrentWeek() {
  if (!canClearCurrentWeek.value)
    return

  const { startDate, endDate } = currentWeekRange.value
  Modal.confirm({
    title: '清空本周课表',
    centered: true,
    okText: '确认清空',
    cancelText: '取消',
    okType: 'danger',
    content: `将硬删除 ${startDate} ~ ${endDate} 的全部课表日程，以及对应的学员挂接与批次元数据。删除后不可恢复，不会保留软删除数据。`,
    async onOk() {
      clearingWeek.value = true
      try {
        const res = await clearWeekTeachingSchedulesApi({ startDate, endDate })
        if (res.code !== 200)
          throw new Error(res.message || '清空本周课表失败')
        const deleted = Number(res.result?.deleted || 0)
        messageService.success(`已硬删除 ${deleted} 条 ${startDate} ~ ${endDate} 的课表日程`)
        emitter.emit(EVENTS.REFRESH_DATA)
      }
      catch (error) {
        const message = error?.message || error?.response?.data?.message || '清空本周课表失败'
        messageService.error(message)
        return Promise.reject(error)
      }
      finally {
        clearingWeek.value = false
      }
      return undefined
    },
  })
}
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
          <smart-timetable @week-range-change="value => updateWeekRange('1', value)" />
        </a-tab-pane>
        <a-tab-pane key="2" tab="时间课表">
          <time-timetable @week-range-change="value => updateWeekRange('2', value)" />
        </a-tab-pane>
        <a-tab-pane key="3" tab="冲突日程">
          <conflict-schedule />
        </a-tab-pane>
        <a-tab-pane key="4" tab="教师矩阵">
          <TeacherMatrixApiTimetable @week-range-change="value => updateWeekRange('4', value)" />
        </a-tab-pane>
      </a-tabs>
      <a-button
        v-if="activeKey !== '3'"
        class="timetable-clear-week-btn"
        :loading="clearingWeek"
        :disabled="!canClearCurrentWeek"
        @click="handleClearCurrentWeek"
      >
        <template #icon>
          <DeleteOutlined class="timetable-clear-week-btn__icon" />
        </template>
        清空本周课表
      </a-button>
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
    position: relative;

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

.timetable-clear-week-btn {
  position: absolute;
  top: 6px;
  right: 24px;
  height: 32px;
  padding: 0 16px;
  border: none;
  border-radius: 16px;
  background: #f6f7f8;
  color: #222;
  font-weight: 500;
  box-shadow: none;

  &:hover,
  &:focus {
    color: #cf1322;
    background: #fff1f0;
  }

  &:disabled {
    color: #bfbfbf;
    background: #f5f5f5;
  }
}

.timetable-clear-week-btn__icon {
  color: #ff4d4f;
}
</style>
