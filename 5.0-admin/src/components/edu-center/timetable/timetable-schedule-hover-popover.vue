<script setup lang="ts">
import { CopyOutlined, EditOutlined } from '@ant-design/icons-vue'
import dayjs from 'dayjs'
import { computed, getCurrentInstance, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { type TeachingScheduleDetail, type TeachingScheduleDetailStudent, getTeachingScheduleDetailApi } from '@/api/edu-center/teaching-schedule'

const props = withDefaults(defineProps<{
  open?: boolean
  scheduleId?: string
  modeLabel?: string
  lessonTitle?: string
  teacherName?: string
  courseName?: string
  assistantText?: string
  studentText?: string
  trialStudentText?: string
  leaveStudentText?: string
  remarkText?: string
  classroomName?: string
  timeText?: string
  conflictText?: string
  showCopyAction?: boolean
}>(), {
  scheduleId: '',
  modeLabel: '课程',
  lessonTitle: '课程',
  teacherName: '-',
  courseName: '-',
  assistantText: '未安排',
  studentText: '-',
  trialStudentText: '-',
  leaveStudentText: '-',
  remarkText: '-',
  classroomName: '-',
  timeText: '-',
  conflictText: '',
  showCopyAction: true,
})

const emit = defineEmits<{
  (e: 'detail'): void
  (e: 'openChange', value: boolean): void
}>()

const router = useRouter()
const instance = getCurrentInstance()
const popoverInnerStyle = {
  padding: '0px',
}
const innerOpen = ref(false)
const detailLoading = ref(false)
const detailData = ref<TeachingScheduleDetail | null>(null)
let detailLoadSeq = 0
const isOpenControlled = computed(() => {
  const vnodeProps = instance?.vnode.props
  return Boolean(vnodeProps && Object.prototype.hasOwnProperty.call(vnodeProps, 'open'))
})
const popoverOpenProps = computed(() => (
  { open: isOpenControlled.value ? props.open : innerOpen.value }
))
const currentOpen = computed(() => (isOpenControlled.value ? Boolean(props.open) : innerOpen.value))

function formatWeek(date: string) {
  const day = dayjs(date).day()
  const weekMap: Record<number, string> = {
    0: '周日',
    1: '周一',
    2: '周二',
    3: '周三',
    4: '周四',
    5: '周五',
    6: '周六',
  }
  return weekMap[day] || '-'
}

function firstNonEmptyText(...values: Array<string | undefined | null>) {
  for (const value of values) {
    const text = String(value || '').trim()
    if (text)
      return text
  }
  return '-'
}

function formatNameBucket(names: string[]) {
  const validNames = names.map(item => String(item || '').trim()).filter(Boolean)
  if (!validNames.length)
    return '-'
  return `${validNames.length}人，${validNames.join('、')}`
}

function formatStudentBucket(students: TeachingScheduleDetailStudent[]) {
  return formatNameBucket(students.map(item => item.studentName))
}

const activeStudents = computed(() => {
  if (!detailData.value)
    return []
  return (detailData.value.students || []).filter(item => Number(item.scheduleStudentType) !== 3)
})
const trialStudents = computed(() => {
  if (!detailData.value)
    return []
  return (detailData.value.students || []).filter(item => Number(item.scheduleStudentType) === 3)
})
const leaveStudents = computed(() => detailData.value?.leaveStudents || [])
const displayLessonTitle = computed(() => firstNonEmptyText(
  detailData.value?.teachingClassName,
  detailData.value?.lessonName,
  props.lessonTitle,
))
const displayTeacherName = computed(() => firstNonEmptyText(detailData.value?.teacherName, props.teacherName))
const displayCourseName = computed(() => firstNonEmptyText(detailData.value?.lessonName, props.courseName))
const displayAssistantText = computed(() => {
  if (detailData.value) {
    const assistantNames = Array.isArray(detailData.value.assistantNames) ? detailData.value.assistantNames : []
    return formatNameBucket(assistantNames)
  }
  return firstNonEmptyText(props.assistantText)
})
const displayStudentText = computed(() => (
  detailData.value ? formatStudentBucket(activeStudents.value) : firstNonEmptyText(props.studentText)
))
const displayTrialStudentText = computed(() => (
  detailData.value ? formatStudentBucket(trialStudents.value) : firstNonEmptyText(props.trialStudentText)
))
const displayLeaveStudentText = computed(() => (
  detailData.value ? formatStudentBucket(leaveStudents.value) : firstNonEmptyText(props.leaveStudentText)
))
const displayRemarkText = computed(() => firstNonEmptyText(detailData.value?.remark, props.remarkText))
const displayTimeText = computed(() => {
  if (!detailData.value)
    return firstNonEmptyText(props.timeText)
  const dateText = dayjs(detailData.value.lessonDate).format('M月D日')
  const weekText = formatWeek(detailData.value.lessonDate)
  const startTime = dayjs(detailData.value.startAt).format('HH:mm')
  const endTime = dayjs(detailData.value.endAt).format('HH:mm')
  return `${startTime} ~ ${endTime}(${weekText}) ${dateText}`
})

async function loadLatestDetail() {
  const scheduleId = String(props.scheduleId || '').trim()
  if (!scheduleId) {
    detailData.value = null
    return
  }
  const seq = ++detailLoadSeq
  detailLoading.value = true
  try {
    const res = await getTeachingScheduleDetailApi({ id: scheduleId })
    if (seq !== detailLoadSeq)
      return
    if (res.code !== 200 || !res.result)
      throw new Error(res.message || '加载日程详情失败')
    detailData.value = res.result
  }
  catch (error) {
    if (seq !== detailLoadSeq)
      return
    detailData.value = null
    console.error('load hover schedule detail failed', error)
  }
  finally {
    if (seq === detailLoadSeq)
      detailLoading.value = false
  }
}

function closePopover() {
  if (!isOpenControlled.value)
    innerOpen.value = false
  emit('openChange', false)
}

function handleOpenChange(value: boolean) {
  if (!isOpenControlled.value)
    innerOpen.value = value
  emit('openChange', value)
}

function openDetail() {
  closePopover()
  emit('detail')
}

function goRollCall() {
  closePopover()
  router.push('/edu-center/roll-call-list')
}

watch(
  () => `${currentOpen.value}|${String(props.scheduleId || '').trim()}`,
  async () => {
    if (!currentOpen.value) {
      detailLoading.value = false
      return
    }
    await loadLatestDetail()
  },
  { immediate: true },
)
</script>

<template>
  <a-popover
    trigger="hover"
    placement="rightTop"
    overlay-class-name="st-schedule-cell-popover"
    :overlay-inner-style="popoverInnerStyle"
    :mouse-enter-delay="0.12"
    :mouse-leave-delay="0.06"
    v-bind="popoverOpenProps"
    @open-change="handleOpenChange"
  >
    <template #content>
      <a-spin :spinning="detailLoading">
        <div class="st-schedule-hover-card">
          <div class="st-schedule-hover-card__header">
            <div class="st-schedule-hover-card__hero">
              <div class="st-schedule-hover-card__badge-shell">
                <div class="st-schedule-hover-card__badge">
                  {{ modeLabel }}
                </div>
              </div>

              <div class="st-schedule-hover-card__hero-main">
                <div class="st-schedule-hover-card__hero-top">
                  <div class="st-schedule-hover-card__title" :title="displayLessonTitle">
                    {{ displayLessonTitle }}
                  </div>
                  <button
                    type="button"
                    class="st-schedule-hover-card__detail-link"
                    @click.stop="openDetail"
                  >
                    详情
                  </button>
                </div>
                <div class="st-schedule-hover-card__time" :title="displayTimeText">
                  {{ displayTimeText }}
                </div>
              </div>
            </div>
          </div>

          <div class="st-schedule-hover-card__body">
            <div class="st-schedule-hover-card__row">
              <span>上课教师：</span>
              <strong :title="displayTeacherName">{{ displayTeacherName }}</strong>
            </div>
            <div class="st-schedule-hover-card__row">
              <span>课程：</span>
              <strong :title="displayCourseName">{{ displayCourseName }}</strong>
            </div>
            <div class="st-schedule-hover-card__row">
              <span>上课助教：</span>
              <strong :title="displayAssistantText">{{ displayAssistantText }}</strong>
            </div>
            <div class="st-schedule-hover-card__row">
              <span>上课学员：</span>
              <strong class="st-schedule-hover-card__value--primary" :title="displayStudentText">{{ displayStudentText }}</strong>
            </div>
            <div class="st-schedule-hover-card__row">
              <span>试听学员：</span>
              <strong :title="displayTrialStudentText">{{ displayTrialStudentText }}</strong>
            </div>
            <div class="st-schedule-hover-card__row">
              <span>请假学员：</span>
              <strong :title="displayLeaveStudentText">{{ displayLeaveStudentText }}</strong>
            </div>
            <div class="st-schedule-hover-card__row">
              <span>对内备注：</span>
              <strong :title="displayRemarkText">{{ displayRemarkText }}</strong>
            </div>
            <div v-if="conflictText" class="st-schedule-hover-card__row st-schedule-hover-card__row--danger">
              <span>冲突说明：</span>
              <strong :title="conflictText">{{ conflictText }}</strong>
            </div>
          </div>

          <div class="st-schedule-hover-card__footer">
            <div class="st-schedule-hover-card__actions">
              <a-tooltip title="编辑日程" placement="top">
                <button
                  type="button"
                  class="st-schedule-hover-card__icon-btn"
                  @click.stop="openDetail"
                >
                  <EditOutlined />
                </button>
              </a-tooltip>

              <a-tooltip v-if="showCopyAction" title="复制日程" placement="top">
                <button
                  type="button"
                  class="st-schedule-hover-card__icon-btn"
                  @click.stop
                >
                  <CopyOutlined />
                </button>
              </a-tooltip>
            </div>

            <button
              type="button"
              class="st-schedule-hover-card__primary-btn"
              @click.stop="goRollCall"
            >
              去点名
            </button>
          </div>
        </div>
      </a-spin>
    </template>

    <slot />
  </a-popover>
</template>

<style scoped lang="less">
:deep(.st-schedule-cell-popover .ant-popover-inner) {
  padding: 0 !important;
  border-radius: 8px;
  overflow: hidden;
  box-shadow:
    0 14px 32px rgba(15, 23, 42, 0.14),
    0 4px 12px rgba(15, 23, 42, 0.08);
}

:deep(.st-schedule-cell-popover .ant-popover-inner-content) {
  padding: 0 !important;
}

.st-schedule-hover-card {
  width: 344px;
  max-width: min(344px, 90vw);
  min-height: 273px;
  background: #fff;
  border-radius: 8px;
  overflow: hidden;
}

.st-schedule-hover-card__header {
  padding: 0 0 1px;
  background: linear-gradient(135deg, #166dff 0%, #1d98ff 100%);
}

.st-schedule-hover-card__hero {
  display: flex;
  gap: 14px;
  align-items: flex-start;
  padding: 16px 18px 14px;
  color: #fff;
}

.st-schedule-hover-card__badge-shell {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 46px;
  height: 46px;
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.92);
  box-shadow: 0 8px 18px rgba(7, 55, 143, 0.16);
}

.st-schedule-hover-card__badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: 8px;
  background: linear-gradient(180deg, #ff8a85 0%, #ff5353 100%);
  color: #fff;
  font-size: 9px;
  font-weight: 700;
  line-height: 1;
}

.st-schedule-hover-card__hero-main {
  min-width: 0;
  flex: 1;
}

.st-schedule-hover-card__hero-top {
  display: flex;
  gap: 12px;
  align-items: flex-start;
  justify-content: space-between;
}

.st-schedule-hover-card__detail-link {
  padding: 0;
  border: 0;
  background: transparent;
  color: #fff;
  font-size: 13px;
  font-weight: 600;
  line-height: 24px;
  cursor: pointer;
  white-space: nowrap;
}

.st-schedule-hover-card__detail-link::after {
  content: ' >';
}

.st-schedule-hover-card__title {
  overflow: hidden;
  color: #fff;
  font-size: 16px;
  font-weight: 700;
  line-height: 24px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.st-schedule-hover-card__time {
  margin-top: 4px;
  overflow: hidden;
  color: rgba(255, 255, 255, 0.96);
  font-size: 13px;
  font-weight: 600;
  line-height: 18px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.st-schedule-hover-card__body {
  display: flex;
  flex-direction: column;
  gap: 0;
  padding: 10px 18px 2px;
}

.st-schedule-hover-card__row {
  display: grid;
  grid-template-columns: max-content minmax(0, 1fr);
  column-gap: 8px;
  row-gap: 0;
  align-items: start;
  font-size: 12px;
  line-height: 22px;
}

.st-schedule-hover-card__row > span {
  color: #8f8f8f;
  font-weight: 400;
}

.st-schedule-hover-card__row > strong {
  overflow: hidden;
  color: #6c6c6c;
  font-weight: 400;
  text-overflow: ellipsis;
  white-space: nowrap;
  min-width: 0;
}

.st-schedule-hover-card__value--primary {
  color: #166dff !important;
}

.st-schedule-hover-card__row--danger > strong {
  color: #cf1322;
}

.st-schedule-hover-card__footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 18px 14px;
  margin-top: auto;
}

.st-schedule-hover-card__actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.st-schedule-hover-card__icon-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  padding: 0;
  border: 0;
  border-radius: 50%;
  background: transparent;
  color: #9f9f9f;
  font-size: 18px;
  cursor: pointer;
  transition: background-color 0.18s ease, color 0.18s ease;
}

.st-schedule-hover-card__icon-btn:hover,
.st-schedule-hover-card__icon-btn--active {
  background: #e8f1ff;
  color: #166dff;
}

.st-schedule-hover-card__primary-btn {
  width: 74px;
  min-width: 74px;
  height: 28px;
  padding: 0;
  border: 0;
  border-radius: 6px;
  background: linear-gradient(180deg, #1970ff 0%, #1660e8 100%);
  color: #fff;
  font-size: 12px;
  font-weight: 700;
  line-height: 28px;
  cursor: pointer;
}
</style>
