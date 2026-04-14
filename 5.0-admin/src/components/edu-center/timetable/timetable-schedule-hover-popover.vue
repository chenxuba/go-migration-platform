<script setup lang="ts">
import { CopyOutlined, EditOutlined } from '@ant-design/icons-vue'
import dayjs from 'dayjs'
import { computed, getCurrentInstance, nextTick, onMounted, onUnmounted, onUpdated, ref, watch } from 'vue'
import { type TeachingScheduleBatchMeta, type TeachingScheduleDetail, type TeachingScheduleDetailStudent, getTeachingScheduleDetailApi } from '@/api/edu-center/teaching-schedule'
import ClassRecordDetails from '@/components/common/class-record-details.vue'
import RollCallDrawer from '@/components/common/roll-call-drawer.vue'

interface ScheduleEditPayload {
  batchMeta?: TeachingScheduleBatchMeta
  batchNo?: string
  batchSize?: number
}

const props = withDefaults(defineProps<{
  open?: boolean
  scheduleId?: string
  editable?: boolean
  batchNo?: string
  batchSize?: number
  lessonDate?: string
  callStatusKey?: string
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
  editable: true,
  batchNo: '',
  batchSize: 0,
  lessonDate: '',
  callStatusKey: 'unsigned',
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
  (e: 'copy', payload?: ScheduleEditPayload): void
  (e: 'copy-current', payload?: ScheduleEditPayload): void
  (e: 'edit', payload?: ScheduleEditPayload): void
  (e: 'edit-current', payload?: ScheduleEditPayload): void
  (e: 'openChange', value: boolean): void
}>()

const instance = getCurrentInstance()
const hoverActionTooltipStyle = {
  zIndex: 1301,
}
const popoverSafeWidth = 376
const popoverSafeHeight = 332
const innerOpen = ref(false)
const detailLoading = ref(false)
const detailData = ref<TeachingScheduleDetail | null>(null)
const detailCache = new Map<string, TeachingScheduleDetail>()
const triggerWrapperRef = ref<HTMLElement | null>(null)
const floatingCardRef = ref<HTMLElement | null>(null)
const floatingCardStyle = ref<Record<string, string>>({
  left: '-9999px',
  top: '-9999px',
})
const rollCallDrawerOpen = ref(false)
const classRecordDrawerOpen = ref(false)
const currentTeachingRecordId = ref('')
let detailLoadSeq = 0
let floatingPositionFrame = 0
const isOpenControlled = computed(() => {
  const vnodeProps = instance?.vnode.props
  return Boolean(vnodeProps && Object.prototype.hasOwnProperty.call(vnodeProps, 'open'))
})
const currentScheduleId = computed(() => String(props.scheduleId || '').trim())
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

function hasBatchMetaSchedule(meta?: TeachingScheduleBatchMeta | null) {
  if (!meta)
    return false
  const schedulingMode = String(meta.schedulingMode || '').trim()
  const repeatRule = String(meta.repeatRule || '').trim()
  const plannedClassCount = Number(meta.plannedClassCount || 0)
  const freeSelectedDates = Array.isArray(meta.freeSelectedDates) ? meta.freeSelectedDates.filter(Boolean) : []
  return plannedClassCount > 1
    || freeSelectedDates.length > 1
    || schedulingMode === 'free'
    || (schedulingMode === 'repeat' && repeatRule !== '' && repeatRule !== 'none')
}

function resolveCallStatusKeyFromNumber(status?: number | null) {
  if (Number(status || 0) === 2)
    return 'signed'
  if (Number(status || 0) === 3)
    return 'partial'
  return 'unsigned'
}

const activeStudents = computed(() => detailData.value?.students || [])
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
const currentLessonDate = computed(() => String(detailData.value?.lessonDate || props.lessonDate || '').trim())
const currentCallStatusKey = computed(() => {
  if (detailData.value)
    return resolveCallStatusKeyFromNumber(detailData.value.callStatus)
  return String(props.callStatusKey || 'unsigned').trim() || 'unsigned'
})
const hasInternalDrawerOpen = computed(() => rollCallDrawerOpen.value || classRecordDrawerOpen.value)
const showFloatingCard = computed(() => currentOpen.value && !hasInternalDrawerOpen.value)
const isPastSchedule = computed(() => {
  const lessonDate = currentLessonDate.value
  if (!lessonDate)
    return false
  return dayjs(lessonDate).isBefore(dayjs().startOf('day'), 'day')
})
const isFutureSchedule = computed(() => {
  const lessonDate = currentLessonDate.value
  if (!lessonDate)
    return false
  return dayjs(lessonDate).isAfter(dayjs().startOf('day'), 'day')
})
const hasBatchSchedule = computed(() => {
  const batchSize = Number(detailData.value?.batchSize || props.batchSize || 0)
  const batchNo = String(detailData.value?.batchNo || props.batchNo || '').trim()
  return batchSize > 1 || batchNo !== '' || hasBatchMetaSchedule(detailData.value?.batchMeta)
})
const canEditByContext = computed(() => Boolean(String(props.scheduleId || '').trim()) && props.editable)
const isRolledCallSchedule = computed(() => currentCallStatusKey.value === 'signed')
const canEditSchedule = computed(() => canEditByContext.value && !isPastSchedule.value && !isRolledCallSchedule.value)
const scheduleEditPayload = computed<ScheduleEditPayload>(() => {
  const batchMeta = detailData.value?.batchMeta
  const batchNo = String(detailData.value?.batchNo || props.batchNo || '').trim() || undefined
  const batchSize = Number(detailData.value?.batchSize || props.batchSize || 0)
  return {
    batchMeta: batchMeta ? {
      ...batchMeta,
      selectedWeekdays: Array.isArray(batchMeta.selectedWeekdays) ? [...batchMeta.selectedWeekdays] : undefined,
      freeSelectedDates: Array.isArray(batchMeta.freeSelectedDates) ? [...batchMeta.freeSelectedDates] : undefined,
    } : undefined,
    batchNo,
    batchSize: batchSize > 0 ? batchSize : undefined,
  }
})
const editDisabledReason = computed(() => (
  isRolledCallSchedule.value ? '已点名日程不可编辑' : (isPastSchedule.value ? '过去日程不可编辑' : (canEditByContext.value ? '编辑日程' : '当前日程不可编辑'))
))
const rollCallDisabledReason = computed(() => {
  const serverReason = String(detailData.value?.rollCallDisabledReason || '').trim()
  if (serverReason)
    return serverReason
  return isFutureSchedule.value ? '未到日期，不可点名' : ''
})
const canRollCall = computed(() => {
  if (currentCallStatusKey.value === 'signed')
    return true
  if (typeof detailData.value?.canRollCall === 'boolean')
    return detailData.value.canRollCall
  return !isFutureSchedule.value
})
const rollCallButtonText = computed(() => (currentCallStatusKey.value === 'signed' ? '点名详情' : '去点名'))

async function loadLatestDetail(force = true) {
  const scheduleId = currentScheduleId.value
  if (!scheduleId) {
    detailData.value = null
    return null
  }
  const cached = detailCache.get(scheduleId)
  if (cached)
    detailData.value = cached
  else if (!detailData.value || String(detailData.value.id || '').trim() !== scheduleId)
    detailData.value = null
  if (cached && !force)
    return cached
  const seq = ++detailLoadSeq
  detailLoading.value = true
  try {
    const res = await getTeachingScheduleDetailApi({ id: scheduleId })
    if (seq !== detailLoadSeq)
      return detailData.value
    if (res.code !== 200 || !res.result)
      throw new Error(res.message || '加载日程详情失败')
    detailData.value = res.result
    detailCache.set(scheduleId, res.result)
    return res.result
  }
  catch (error) {
    if (seq !== detailLoadSeq)
      return detailData.value
    if (!cached)
      detailData.value = null
    console.error('load hover schedule detail failed', error)
    return detailData.value
  }
  finally {
    if (seq === detailLoadSeq)
      detailLoading.value = false
  }
}

async function ensureDetailLoaded() {
  const scheduleId = currentScheduleId.value
  if (!scheduleId)
    return null
  if (String(detailData.value?.id || '').trim() === scheduleId)
    return detailData.value
  const cached = detailCache.get(scheduleId)
  if (cached) {
    detailData.value = cached
    return cached
  }
  return await loadLatestDetail(true)
}

function closePopover() {
  if (!isOpenControlled.value)
    innerOpen.value = false
  emit('openChange', false)
}

function getTriggerNode() {
  const wrapper = triggerWrapperRef.value
  if (!wrapper)
    return null
  const firstChild = wrapper.firstElementChild
  return firstChild instanceof HTMLElement ? firstChild : wrapper
}

function clamp(value: number, min: number, max: number) {
  return Math.min(Math.max(value, min), max)
}

function resolveTooltipContainer() {
  if (typeof document === 'undefined')
    return undefined as unknown as HTMLElement
  return floatingCardRef.value || document.body
}

function resolveDropdownContainer(triggerNode?: HTMLElement) {
  if (triggerNode instanceof HTMLElement) {
    const card = triggerNode.closest('.st-schedule-hover-card')
    if (card instanceof HTMLElement)
      return card
  }
  if (typeof document === 'undefined')
    return undefined as unknown as HTMLElement
  return floatingCardRef.value || document.body
}

function updateFloatingCardPosition() {
  if (!showFloatingCard.value || typeof window === 'undefined')
    return
  const triggerNode = getTriggerNode()
  if (!triggerNode)
    return
  const triggerRect = triggerNode.getBoundingClientRect()
  const cardWidth = Math.max(344, Number(floatingCardRef.value?.offsetWidth || 0))
  const cardHeight = Math.max(273, Number(floatingCardRef.value?.offsetHeight || 0))
  const gap = 2
  const overlap = 6
  const viewportPadding = 8

  const spaceRight = window.innerWidth - triggerRect.right
  const spaceLeft = triggerRect.left
  const placeOnRight = spaceRight >= cardWidth + gap || spaceRight >= spaceLeft

  const topAlignedSpace = window.innerHeight - triggerRect.top
  const bottomAlignedSpace = triggerRect.bottom
  const alignToTop = topAlignedSpace >= cardHeight || topAlignedSpace >= bottomAlignedSpace

  const preferredLeft = placeOnRight
    ? triggerRect.right + gap - overlap
    : triggerRect.left - cardWidth - gap + overlap
  const preferredTop = alignToTop
    ? triggerRect.top
    : triggerRect.bottom - cardHeight

  floatingCardStyle.value = {
    left: `${Math.round(clamp(preferredLeft, viewportPadding, Math.max(viewportPadding, window.innerWidth - cardWidth - viewportPadding)))}px`,
    top: `${Math.round(clamp(preferredTop, viewportPadding, Math.max(viewportPadding, window.innerHeight - cardHeight - viewportPadding)))}px`,
  }
}

function scheduleFloatingCardPositionUpdate() {
  if (typeof window === 'undefined')
    return
  if (floatingPositionFrame)
    window.cancelAnimationFrame(floatingPositionFrame)
  floatingPositionFrame = window.requestAnimationFrame(() => {
    floatingPositionFrame = 0
    updateFloatingCardPosition()
  })
}

function handleCardMouseEnter() {
  emit('openChange', true)
  scheduleFloatingCardPositionUpdate()
}

function handleCardMouseLeave() {
  if (hasInternalDrawerOpen.value)
    return
  closePopover()
}

function openDetail() {
  closePopover()
  emit('detail')
}

async function openEdit() {
  if (hasBatchSchedule.value)
    await ensureDetailLoaded()
  closePopover()
  emit('edit', scheduleEditPayload.value)
}

async function openEditCurrent() {
  if (hasBatchSchedule.value)
    await ensureDetailLoaded()
  closePopover()
  emit('edit-current', scheduleEditPayload.value)
}

async function openCopy() {
  if (hasBatchSchedule.value)
    await ensureDetailLoaded()
  closePopover()
  emit('copy', scheduleEditPayload.value)
}

async function openCopyCurrent() {
  if (hasBatchSchedule.value)
    await ensureDetailLoaded()
  closePopover()
  emit('copy-current', scheduleEditPayload.value)
}

async function handleBatchEditMenuClick({ key, domEvent }: { key: string | number, domEvent?: Event }) {
  domEvent?.stopPropagation()
  if (!canEditSchedule.value)
    return
  if (String(key) === 'current')
    await openEditCurrent()
  else
    await openEdit()
}

async function handleBatchCopyMenuClick({ key, domEvent }: { key: string | number, domEvent?: Event }) {
  domEvent?.stopPropagation()
  if (String(key) === 'current')
    await openCopyCurrent()
  else
    await openCopy()
}

async function goRollCall() {
  if (currentCallStatusKey.value === 'signed') {
    const detail = await ensureDetailLoaded()
    currentTeachingRecordId.value = String(detail?.teachingRecordId || '').trim()
    if (!currentTeachingRecordId.value)
      return
    classRecordDrawerOpen.value = true
    return
  }
  if (!canRollCall.value)
    return
  rollCallDrawerOpen.value = true
}

async function handleRollCallConfirmed(teachingRecordId?: string) {
  const nextTeachingRecordId = String(teachingRecordId || '').trim() || String(detailData.value?.teachingRecordId || '').trim()
  if (nextTeachingRecordId)
    currentTeachingRecordId.value = nextTeachingRecordId
  await loadLatestDetail()
  if (currentTeachingRecordId.value)
    classRecordDrawerOpen.value = true
}

watch(
  () => `${currentOpen.value}|${currentScheduleId.value}`,
  () => {
    if (!currentOpen.value) {
      detailLoading.value = false
      return
    }
    currentTeachingRecordId.value = ''
    detailData.value = detailCache.get(currentScheduleId.value) || null
    nextTick(() => scheduleFloatingCardPositionUpdate())
    void loadLatestDetail(true)
  },
  { immediate: true },
)

watch(
  hasInternalDrawerOpen,
  (open, previousOpen) => {
    if (!open && previousOpen)
      closePopover()
  },
)

onMounted(() => {
  if (typeof window === 'undefined')
    return
  window.addEventListener('resize', scheduleFloatingCardPositionUpdate)
  window.addEventListener('scroll', scheduleFloatingCardPositionUpdate, true)
})

onUpdated(() => {
  if (showFloatingCard.value)
    scheduleFloatingCardPositionUpdate()
})

onUnmounted(() => {
  if (typeof window === 'undefined')
    return
  if (floatingPositionFrame)
    window.cancelAnimationFrame(floatingPositionFrame)
  window.removeEventListener('resize', scheduleFloatingCardPositionUpdate)
  window.removeEventListener('scroll', scheduleFloatingCardPositionUpdate, true)
})
</script>

<template>
  <div ref="triggerWrapperRef" class="st-schedule-hover-trigger">
    <slot />
  </div>

  <Teleport to="body">
    <div
      v-if="currentOpen"
      ref="floatingCardRef"
      v-show="showFloatingCard"
      class="st-schedule-hover-floating"
      :style="floatingCardStyle"
      @mouseenter="handleCardMouseEnter"
      @mouseleave="handleCardMouseLeave"
    >
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
            <a-dropdown
              v-if="hasBatchSchedule && canEditSchedule"
              :trigger="['click']"
              placement="topLeft"
              :get-popup-container="resolveDropdownContainer"
            >
              <template #overlay>
                <a-menu :selectable="false" @click="handleBatchEditMenuClick">
                  <a-menu-item key="current">
                    仅编辑此日程
                  </a-menu-item>
                  <a-menu-item key="future">
                    编辑以后日程
                  </a-menu-item>
                </a-menu>
              </template>

              <button
                type="button"
                class="st-schedule-hover-card__icon-btn"
                @click.stop
              >
                <EditOutlined />
              </button>
            </a-dropdown>
            <a-tooltip
              v-else
              :title="editDisabledReason"
              placement="top"
              :auto-adjust-overflow="false"
              :get-popup-container="resolveTooltipContainer"
              :overlay-style="hoverActionTooltipStyle"
            >
              <button
                type="button"
                class="st-schedule-hover-card__icon-btn"
                :disabled="!canEditSchedule"
                @click.stop="canEditSchedule && openEdit()"
              >
                <EditOutlined />
              </button>
            </a-tooltip>

            <a-dropdown
              v-if="showCopyAction && hasBatchSchedule"
              :trigger="['click']"
              placement="topLeft"
              :get-popup-container="resolveDropdownContainer"
            >
              <template #overlay>
                <a-menu :selectable="false" @click="handleBatchCopyMenuClick">
                  <a-menu-item key="current">
                    仅复制当前课程
                  </a-menu-item>
                  <a-menu-item key="future">
                    复制后续全部课程
                  </a-menu-item>
                </a-menu>
              </template>

              <button
                type="button"
                class="st-schedule-hover-card__icon-btn"
                @click.stop
              >
                <CopyOutlined />
              </button>
            </a-dropdown>
            <a-tooltip
              v-else-if="showCopyAction"
              title="仅复制当前课程"
              placement="top"
              :auto-adjust-overflow="false"
              :get-popup-container="resolveTooltipContainer"
              :overlay-style="hoverActionTooltipStyle"
            >
              <button
                type="button"
                class="st-schedule-hover-card__icon-btn"
                @click.stop="openCopyCurrent()"
              >
                <CopyOutlined />
              </button>
            </a-tooltip>
          </div>

          <a-tooltip
            :title="rollCallDisabledReason || null"
            placement="top"
            :auto-adjust-overflow="false"
            :get-popup-container="resolveTooltipContainer"
            :overlay-style="hoverActionTooltipStyle"
          >
            <span class="st-schedule-hover-card__primary-wrap">
              <button
                type="button"
                class="st-schedule-hover-card__primary-btn"
                :disabled="!canRollCall"
                @click.stop="goRollCall"
              >
                {{ rollCallButtonText }}
              </button>
            </span>
          </a-tooltip>
        </div>
      </div>
    </div>
  </Teleport>
  <RollCallDrawer
    v-model:open="rollCallDrawerOpen"
    :schedule-id="String(props.scheduleId || '')"
    :lesson-day="detailData?.lessonDate || lessonDate || ''"
    @updated="loadLatestDetail"
    @confirmed="handleRollCallConfirmed"
  />
  <ClassRecordDetails v-model:open="classRecordDrawerOpen" :teaching-record-id="currentTeachingRecordId || detailData?.teachingRecordId || ''" @updated="loadLatestDetail" />
</template>

<style scoped lang="less">
.st-schedule-hover-trigger {
  display: contents;
}

.st-schedule-hover-floating {
  position: fixed;
  z-index: 1080;
  pointer-events: auto;
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

.st-schedule-hover-card__primary-wrap {
  display: inline-flex;
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

.st-schedule-hover-card__primary-btn:disabled {
  cursor: not-allowed;
  opacity: 0.56;
}
</style>
