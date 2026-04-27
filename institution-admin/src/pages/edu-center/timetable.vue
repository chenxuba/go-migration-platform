<script setup>
import { DeleteOutlined } from '@ant-design/icons-vue'
import { Modal } from 'ant-design-vue'
import { computed, defineAsyncComponent, onMounted, onUnmounted, ref, watch } from 'vue'
import { clearWeekTeachingSchedulesApi } from '@/api/edu-center/teaching-schedule'
import { AccessEnum, AccessGroup } from '@/constants/access'
import emitter, { EVENTS } from '@/utils/eventBus'
import messageService from '@/utils/messageService'

const loadSmartTimetableTab = () => import('@/components/edu-center/timetable/smart-timetable.vue')
const loadTimeTimetableTab = () => import('@/components/edu-center/timetable/time-timetable.vue')
const loadTeacherMatrixApiTimetableTab = () => import('@/components/edu-center/timetable/teacher-matrix-api-timetable.vue')
const loadConflictScheduleTab = () => import('@/components/edu-center/timetable/conflict-schedule.vue')

const SmartTimetableTabPane = defineAsyncComponent(loadSmartTimetableTab)
const TimeTimetableTabPane = defineAsyncComponent(loadTimeTimetableTab)
const TeacherMatrixApiTimetableTabPane = defineAsyncComponent(loadTeacherMatrixApiTimetableTab)
const ConflictScheduleTabPane = defineAsyncComponent(loadConflictScheduleTab)

const activeKey = ref('1')
const clearingWeek = ref(false)
const { hasAccess } = useAccess()
const weekRanges = ref({
  1: { startDate: '', endDate: '' },
  2: { startDate: '', endDate: '' },
  4: { startDate: '', endDate: '' },
})

const canManageTimetable = computed(() => hasAccess(AccessGroup.edu_timetable_manage))
const canViewConflictSchedule = computed(() => hasAccess(AccessEnum.edu_timetable_conflict_list))

const currentWeekRange = computed(() => weekRanges.value[activeKey.value] || { startDate: '', endDate: '' })
const canClearCurrentWeek = computed(() =>
  activeKey.value !== '3'
  && !!currentWeekRange.value.startDate
  && !!currentWeekRange.value.endDate
  && canManageTimetable.value
  && !clearingWeek.value,
)

watch(canViewConflictSchedule, (visible) => {
  if (!visible && activeKey.value === '3')
    activeKey.value = '1'
}, { immediate: true })

let tabPrefetchTimer = null
let tabPrefetchIdleHandle = null

function prefetchSecondaryTabs() {
  void loadTimeTimetableTab()
  void loadTeacherMatrixApiTimetableTab()
  if (canViewConflictSchedule.value)
    void loadConflictScheduleTab()
}

function clearTabPrefetchTask() {
  if (typeof window === 'undefined')
    return
  if (tabPrefetchTimer) {
    window.clearTimeout(tabPrefetchTimer)
    tabPrefetchTimer = null
  }
  if (tabPrefetchIdleHandle && typeof window.cancelIdleCallback === 'function') {
    window.cancelIdleCallback(tabPrefetchIdleHandle)
    tabPrefetchIdleHandle = null
  }
}

function scheduleTabPrefetch() {
  clearTabPrefetchTask()
  if (typeof window === 'undefined') {
    prefetchSecondaryTabs()
    return
  }
  tabPrefetchTimer = window.setTimeout(() => {
    tabPrefetchTimer = null
    if (typeof window.requestIdleCallback === 'function') {
      tabPrefetchIdleHandle = window.requestIdleCallback(() => {
        tabPrefetchIdleHandle = null
        prefetchSecondaryTabs()
      }, { timeout: 1500 })
      return
    }
    prefetchSecondaryTabs()
  }, 240)
}

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

onMounted(() => {
  scheduleTabPrefetch()
})

onUnmounted(() => {
  clearTabPrefetchTask()
})
</script>

<template>
  <div class="home">
    <div class="tabs">
      <a-tabs
        v-model:active-key="activeKey"
        :animated="{ inkBar: true, tabPane: false }"
        :destroy-inactive-tab-pane="false"
        :tab-bar-style="{
          'border-bottom-left-radius': '0px',
          'border-bottom-right-radius': '0px',
        }"
      >
        <a-tab-pane key="1" tab="智慧课表">
          <Suspense>
            <SmartTimetableTabPane @week-range-change="value => updateWeekRange('1', value)" />
            <template #fallback>
              <div class="timetable-tab-loading">
                <a-spin size="large" />
              </div>
            </template>
          </Suspense>
        </a-tab-pane>
        <a-tab-pane key="2" tab="时间课表">
          <Suspense>
            <TimeTimetableTabPane @week-range-change="value => updateWeekRange('2', value)" />
            <template #fallback>
              <div class="timetable-tab-loading">
                <a-spin size="large" />
              </div>
            </template>
          </Suspense>
        </a-tab-pane>
        <a-tab-pane key="4" tab="教师矩阵">
          <Suspense>
            <TeacherMatrixApiTimetableTabPane @week-range-change="value => updateWeekRange('4', value)" />
            <template #fallback>
              <div class="timetable-tab-loading">
                <a-spin size="large" />
              </div>
            </template>
          </Suspense>
        </a-tab-pane>
        <a-tab-pane v-if="canViewConflictSchedule" key="3" tab="冲突日程">
          <Suspense>
            <ConflictScheduleTabPane />
            <template #fallback>
              <div class="timetable-tab-loading">
                <a-spin size="large" />
              </div>
            </template>
          </Suspense>
        </a-tab-pane>
      </a-tabs>
      <a-button
        v-if="canManageTimetable && activeKey !== '3'"
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

    :deep(.ant-tabs-content-holder) {
      background: #fff;
      border-bottom-left-radius: 16px;
      border-bottom-right-radius: 16px;
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

.timetable-tab-loading {
  min-height: 420px;
  display: flex;
  align-items: center;
  justify-content: center;
  background:
    linear-gradient(180deg, rgba(248, 250, 252, 0.9) 0%, rgba(255, 255, 255, 1) 100%);
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
