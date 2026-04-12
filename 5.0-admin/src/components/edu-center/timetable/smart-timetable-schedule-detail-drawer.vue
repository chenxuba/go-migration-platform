<script setup lang="ts">
import { CloseOutlined, DownOutlined } from '@ant-design/icons-vue'
import dayjs from 'dayjs'
import type { TableColumnsType } from 'ant-design-vue'
import { Modal } from 'ant-design-vue'
import { computed, ref, watch } from 'vue'
import scheduleClassImage from '@/assets/images/timetable/schedule-class.png'
import scheduleOneToOneImage from '@/assets/images/timetable/schedule-one2one.png'
import { type TeachingScheduleBatchMeta, type TeachingScheduleDetail, type TeachingScheduleDetailStudent, getTeachingScheduleDetailApi, removeTeachingScheduleStudentCurrentApi } from '@/api/edu-center/teaching-schedule'
import ClassRecordDetails from '@/components/common/class-record-details.vue'
import RollCallAddStudentModal from '@/components/common/roll-call-add-student-modal.vue'
import RollCallDrawer from '@/components/common/roll-call-drawer.vue'
import { useStudentStore } from '@/stores/student'
import messageService from '@/utils/messageService'

interface DrawerSummary {
  scheduleId?: string
  id?: string
  batchNo?: string
  batchSize?: number
  isMain?: boolean
  lessonTitle?: string
  courseName?: string
  teacherName?: string
  assistantText?: string
  classroomName?: string
  studentText?: string
  courseType?: number
}

interface ScheduleEditPayload {
  batchMeta?: TeachingScheduleBatchMeta
  batchNo?: string
  batchSize?: number
}

const props = withDefaults(defineProps<{
  open: boolean
  detail?: DrawerSummary | null
  deleting?: boolean
  editable?: boolean
  deletable?: boolean
}>(), {
  editable: false,
  deletable: true,
})

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'delete'): void
  (e: 'delete-current'): void
  (e: 'delete-future'): void
  (e: 'copy', payload?: ScheduleEditPayload): void
  (e: 'copy-current', payload?: ScheduleEditPayload): void
  (e: 'edit', payload?: ScheduleEditPayload): void
  (e: 'edit-current', payload?: ScheduleEditPayload): void
  (e: 'updated'): void
}>()

const openDrawer = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

const loading = ref(false)
const detailData = ref<TeachingScheduleDetail | null>(null)
const studentStore = useStudentStore()
const openStudentDrawer = ref(false)
const rollCallDrawerOpen = ref(false)
const classRecordDrawerOpen = ref(false)
const activeStudentTabKey = ref('students')
const addStudentModalOpen = ref(false)
const addStudentModalTitle = ref('添加补课学员')
const addStudentType = ref(4)
const defaultStudentAvatar = 'https://pcsys.admin.ybc365.com/a369a751-2be5-4929-974d-9ae4439f54c4.png'
const studentColumns: TableColumnsType<TeachingScheduleDetailStudent> = [
  {
    title: '学员姓名',
    dataIndex: 'studentName',
    key: 'studentName',
    width: 260,
  },
  {
    title: '联系电话',
    dataIndex: 'phone',
    key: 'phone',
    width: 220,
  },
  {
    title: '状态',
    key: 'studentType',
    width: 120,
  },
  {
    title: '操作',
    key: 'action',
    width: 220,
    align: 'center',
  },
]

const repeatRuleLabelMap: Record<string, string> = {
  none: '不重复',
  weekly: '每周重复',
  biweekly: '隔周重复',
  daily: '每天重复',
  alternateDay: '隔天重复',
}

const scheduleId = computed(() => String(props.detail?.scheduleId || props.detail?.id || '').trim())
const isOneToOne = computed(() => {
  if (detailData.value)
    return Number(detailData.value.classType) === 2
  return Number(props.detail?.courseType) === 1
})
const scheduleCover = computed(() => (isOneToOne.value ? scheduleOneToOneImage : scheduleClassImage))
const students = computed(() => detailData.value?.students || [])
const leaveStudents = computed(() => detailData.value?.leaveStudents || [])
const headerTitle = computed(() => {
  if (props.detail?.lessonTitle)
    return props.detail.lessonTitle
  return detailData.value?.lessonName || detailData.value?.teachingClassName || '日程详情'
})
const timeText = computed(() => {
  if (!detailData.value)
    return '-'
  const dateText = dayjs(detailData.value.lessonDate).format('YYYY-MM-DD')
  const weekText = formatWeek(detailData.value.lessonDate)
  const startTime = dayjs(detailData.value.startAt).format('HH:mm')
  const endTime = dayjs(detailData.value.endAt).format('HH:mm')
  return `${dateText}(${weekText}) ${startTime} ~ ${endTime}`
})
const durationText = computed(() => {
  const minutes = Number(detailData.value?.durationMinutes || 0)
  return minutes > 0 ? `${minutes}分钟` : '-'
})
const assistantText = computed(() => {
  if (Array.isArray(detailData.value?.assistantNames) && detailData.value.assistantNames.length)
    return detailData.value.assistantNames.join('、')
  return props.detail?.assistantText || '-'
})
const repeatRuleText = computed(() => {
  const meta = detailData.value?.batchMeta
  if (!meta)
    return '不重复'
  if (meta.schedulingMode === 'free')
    return '自由排课'
  const base = repeatRuleLabelMap[String(meta.repeatRule || '').trim()] || String(meta.repeatRule || '').trim() || '不重复'
  if ((meta.repeatRule === 'weekly' || meta.repeatRule === 'biweekly') && Array.isArray(meta.selectedWeekdays) && meta.selectedWeekdays.length)
    return `${base} · ${meta.selectedWeekdays.join(' / ')}`
  return base
})
const remarkText = computed(() => detailData.value?.remark || '-')
const isPastSchedule = computed(() => {
  const lessonDate = String(detailData.value?.lessonDate || '').trim()
  if (!lessonDate)
    return false
  return dayjs(lessonDate).isBefore(dayjs().startOf('day'), 'day')
})
const isFutureSchedule = computed(() => {
  const lessonDate = String(detailData.value?.lessonDate || '').trim()
  if (!lessonDate)
    return false
  return dayjs(lessonDate).isAfter(dayjs().startOf('day'), 'day')
})
const editDisabledReason = computed(() => (
  isPastSchedule.value ? '过去日程不可编辑' : ''
))
const canEditSchedule = computed(() => props.editable && !isPastSchedule.value)
const rollCallDisabledReason = computed(() => {
  const serverReason = String(detailData.value?.rollCallDisabledReason || '').trim()
  if (serverReason)
    return serverReason
  return isFutureSchedule.value ? '未到日期，不可点名' : ''
})
const canManageCurrentStudents = computed(() => Number(detailData.value?.callStatus || 1) !== 2)
const canRollCall = computed(() => {
  if (Number(detailData.value?.callStatus || 1) === 2)
    return true
  if (typeof detailData.value?.canRollCall === 'boolean')
    return detailData.value.canRollCall
  return !isFutureSchedule.value
})
const rollCallButtonText = computed(() => (Number(detailData.value?.callStatus || 1) === 2 ? '点名详情' : '去点名'))
const deleteDisabledReason = computed(() => {
  if (Number(detailData.value?.callStatus || 1) === 2)
    return '当前日程已点名，不可删除'
  return ''
})
const canDeleteSchedule = computed(() => props.deletable && !deleteDisabledReason.value)
const hasBatchSchedule = computed(() => {
  return Number(detailData.value?.batchSize || props.detail?.batchSize || 0) > 1
    || String(detailData.value?.batchNo || props.detail?.batchNo || '').trim() !== ''
    || hasBatchMetaSchedule(detailData.value?.batchMeta)
})
const scheduleEditPayload = computed<ScheduleEditPayload>(() => {
  const batchMeta = detailData.value?.batchMeta
  const batchNo = String(detailData.value?.batchNo || props.detail?.batchNo || '').trim() || undefined
  const batchSize = Number(detailData.value?.batchSize || props.detail?.batchSize || 0)
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
const currentStudentList = computed(() => (activeStudentTabKey.value === 'leave' ? leaveStudents.value : students.value))
const studentCardTitle = computed(() => {
  if (isOneToOne.value)
    return '学员列表'
  return activeStudentTabKey.value === 'leave' ? '请假学员列表' : '班级学员列表'
})

let loadSeq = 0

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

function handleViewStudent(studentId?: string) {
  const id = String(studentId || '').trim()
  if (!id)
    return
  studentStore.setStudentId(id)
  openStudentDrawer.value = true
}

function getStudentTypeTagClass(studentType?: number) {
  switch (Number(studentType || 0)) {
    case 2:
      return 'student-type-tag--temporary'
    case 3:
      return 'student-type-tag--trial'
    case 4:
      return 'student-type-tag--makeup'
    default:
      return 'student-type-tag--member'
  }
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

function handleAddStudentMenuClick({ key }) {
  if (!canManageCurrentStudents.value) {
    messageService.warning('已点名日程不可添加学员')
    return
  }
  if (key === 'makeup') {
    messageService.info('补课学员功能暂未开发')
    return
  }
  else if (key === 'temporary') {
    addStudentModalTitle.value = '添加临时学员'
    addStudentType.value = 2
  }
  else {
    addStudentModalTitle.value = '添加试听学员'
    addStudentType.value = 3
  }
  addStudentModalOpen.value = true
}

async function handleAddStudentSuccess() {
  await loadDetail()
  emit('updated')
}

function handleStudentReschedule(student: Record<string, any>) {
  const name = String(student?.studentName || '').trim() || '当前学员'
  messageService.info(`${name} 的班课调课功能待接入`)
}

function handleBatchEditMenuClick({ key, domEvent }: { key: string | number, domEvent?: Event }) {
  domEvent?.stopPropagation?.()
  if (!canEditSchedule.value)
    return
  if (key === 'current')
    emit('edit-current', scheduleEditPayload.value)
  else if (key === 'future')
    emit('edit', scheduleEditPayload.value)
}

function handleBatchCopyMenuClick({ key, domEvent }: { key: string | number, domEvent?: Event }) {
  domEvent?.stopPropagation?.()
  if (key === 'copy-current')
    emit('copy-current', scheduleEditPayload.value)
  else if (key === 'copy-all')
    emit('copy', scheduleEditPayload.value)
}

function handleSingleCopyClick() {
  emit('copy-current', scheduleEditPayload.value)
}

function handleBatchDeleteMenuClick({ key, domEvent }: { key: string | number, domEvent?: Event }) {
  domEvent?.stopPropagation?.()
  if (!canDeleteSchedule.value) {
    messageService.warning(deleteDisabledReason.value || '当前日程不可删除')
    return
  }
  if (key === 'delete-current')
    emit('delete-current')
  else if (key === 'delete-future')
    emit('delete-future')
}

function handleSingleEditClick() {
  if (!canEditSchedule.value)
    return
  emit('edit', scheduleEditPayload.value)
}

function goRollCall() {
  if (Number(detailData.value?.callStatus || 1) === 2) {
    classRecordDrawerOpen.value = true
    return
  }
  if (!canRollCall.value)
    return
  rollCallDrawerOpen.value = true
}

function shouldSkipManualErrorMessage(error: any) {
  return Number(error?.response?.status || 0) === 400
}

function handleStudentRemove(student: Record<string, any>) {
  if (!canManageCurrentStudents.value) {
    messageService.warning('已点名日程不可移出本节学员')
    return
  }
  const name = String(student?.studentName || '').trim() || '当前学员'
  const currentScheduleId = scheduleId.value
  const currentStudentId = String(student?.studentId || '').trim()
  if (!currentScheduleId || !currentStudentId) {
    messageService.warning('当前学员缺少移出标识，请刷新后重试')
    return
  }
  let removing = false
  Modal.confirm({
    title: '移出本节学员',
    content: `移出后仅影响本节课，不会影响班级成员和后续未开课。确认移出“${name}”吗？`,
    okText: '确认移出',
    cancelText: '取消',
    async onOk() {
      if (removing)
        return
      removing = true
      try {
        messageService.clear()
        const res = await removeTeachingScheduleStudentCurrentApi({
          scheduleId: currentScheduleId,
          studentId: currentStudentId,
        })
        if (res.code !== 200)
          throw new Error(res.message || '移出本节失败')
        messageService.success(`已将${name}移出本节`)
        await loadDetail()
        emit('updated')
      }
      catch (error: any) {
        if (!shouldSkipManualErrorMessage(error)) {
          messageService.error(error?.response?.data?.message || error?.message || '移出本节失败')
          throw error
        }
      }
      finally {
        removing = false
      }
    },
  })
}

async function loadDetail() {
  const id = scheduleId.value
  if (!openDrawer.value || !id) {
    detailData.value = null
    return
  }

  const seq = ++loadSeq
  loading.value = true
  try {
    const res = await getTeachingScheduleDetailApi({ id })
    if (seq !== loadSeq)
      return
    if (res.code !== 200 || !res.result)
      throw new Error(res.message || '加载日程详情失败')
    detailData.value = res.result
  }
  catch (error: any) {
    if (seq !== loadSeq)
      return
    console.error('load teaching schedule detail failed', error)
    detailData.value = null
    messageService.error(error?.response?.data?.message || error?.message || '加载日程详情失败')
    openDrawer.value = false
  }
  finally {
    if (seq === loadSeq)
      loading.value = false
  }
}

watch(
  () => `${openDrawer.value}|${scheduleId.value}`,
  async () => {
    if (!openDrawer.value) {
      detailData.value = null
      loading.value = false
      activeStudentTabKey.value = 'students'
      return
    }
    activeStudentTabKey.value = 'students'
    await loadDetail()
  },
  { immediate: true },
)

watch(
  () => isOneToOne.value,
  (value) => {
    if (value)
      activeStudentTabKey.value = 'students'
  },
  { immediate: true },
)
</script>

<template>
  <a-drawer
    v-model:open="openDrawer"
    :push="{ distance: 80 }"
    :body-style="{ padding: '0', background: '#f7f7fd' }"
    :closable="false"
    width="1165px"
    placement="right"
  >
    <template #title>
      <div class="custom-header flex justify-between h-4 flex-items-center">
        <div class="text-5">
          日程详情
        </div>
        <a-button type="text" class="close-btn" @click="openDrawer = false">
          <template #icon>
            <CloseOutlined class="text-5 close-icon" />
          </template>
        </a-button>
      </div>
    </template>

    <a-spin :spinning="loading">
      <div class="contenter flex flex-center bg-white px6 py3">
        <div class="avatarBox w-16 h-16 relative">
          <img width="64" height="64" class="rounded-100" :src="scheduleCover" alt="">
        </div>
        <div class="info flex flex-1 ml-4 flex-col">
          <div class="top flex justify-between flex-center flex-1">
            <a-space>
              <div class="name text-5 font-800">
                {{ headerTitle }}
              </div>
            </a-space>
            <a-space>
              <a-dropdown
                v-if="hasBatchSchedule"
                :trigger="['hover']"
                placement="bottomLeft"
              >
                <template #overlay>
                  <a-menu :selectable="false" @click="handleBatchCopyMenuClick">
                    <a-menu-item key="copy-current">
                      仅复制当前课程
                    </a-menu-item>
                    <a-menu-item key="copy-all">
                      复制后续全部课程
                    </a-menu-item>
                  </a-menu>
                </template>

                <a-button @click.stop>
                  复制日程
                  <DownOutlined />
                </a-button>
              </a-dropdown>
              <a-button v-else type="link" @click="handleSingleCopyClick">
                仅复制当前课程
              </a-button>

              <a-dropdown
                v-if="canDeleteSchedule && hasBatchSchedule"
                :trigger="['hover']"
                placement="bottomLeft"
              >
                <template #overlay>
                  <a-menu :selectable="false" @click="handleBatchDeleteMenuClick">
                    <a-menu-item key="delete-current">
                      仅删除此日程
                    </a-menu-item>
                    <a-menu-item key="delete-future">
                      删除后续全部日程
                    </a-menu-item>
                  </a-menu>
                </template>

                <a-button @click.stop>
                  删除
                  <DownOutlined />
                </a-button>
              </a-dropdown>
              <a-tooltip v-else-if="deletable && hasBatchSchedule" :title="deleteDisabledReason || '删除此日程'" placement="top">
                <span>
                  <a-button disabled @click.stop>
                    删除
                    <DownOutlined />
                  </a-button>
                </span>
              </a-tooltip>
              <a-tooltip v-else-if="canDeleteSchedule" title="删除此日程" placement="top">
                <a-button :loading="deleting" @click="$emit('delete')">
                  删除
                </a-button>
              </a-tooltip>
              <a-tooltip v-else-if="deletable" :title="deleteDisabledReason || '删除此日程'" placement="top">
                <span>
                  <a-button disabled>
                    删除
                  </a-button>
                </span>
              </a-tooltip>
              <a-dropdown
                v-if="hasBatchSchedule && canEditSchedule"
                :trigger="['hover']"
                placement="bottomLeft"
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

                <a-button @click.stop>
                  编辑
                  <DownOutlined />
                </a-button>
              </a-dropdown>
              <a-tooltip v-else-if="hasBatchSchedule && editable" :title="editDisabledReason || '编辑日程'" placement="top">
                <span>
                  <a-button disabled @click.stop>
                    编辑
                  </a-button>
                </span>
              </a-tooltip>
              <a-tooltip v-else-if="editable" :title="editDisabledReason || '编辑此日程'" placement="top">
                <span>
                  <a-button :disabled="!canEditSchedule" @click="handleSingleEditClick">
                    编辑
                  </a-button>
                </span>
              </a-tooltip>
              <a-tooltip :title="rollCallDisabledReason || null" placement="top">
                <span>
                  <a-button type="primary" :disabled="!canRollCall" @click="goRollCall">
                    {{ rollCallButtonText }}
                  </a-button>
                </span>
              </a-tooltip>
            </a-space>
          </div>
          <div class="bottom flex-1 flex flex-items-center mt-2">
            <div class="birthday flex-center">
              <span class="text-4 text-#222">{{ timeText }}</span>
              <span class="bg-#e6f0ff text-#06f text-3 px2 py1 rounded-10 ml2">{{ durationText }}</span>
            </div>
          </div>
        </div>
      </div>

      <div class="desc pt-4 bg-white px6 py3 pb0">
        <a-descriptions :column="4" size="small" :content-style="{ color: '#888' }">
          <a-descriptions-item label="上课教师">
            {{ detailData?.teacherName || props.detail?.teacherName || '-' }}
          </a-descriptions-item>
          <a-descriptions-item label="上课助教">
            {{ assistantText }}
          </a-descriptions-item>
          <a-descriptions-item label="上课教室">
            {{ detailData?.classroomName || props.detail?.classroomName || '-' }}
          </a-descriptions-item>
          <a-descriptions-item label="重复规则">
            {{ repeatRuleText }}
          </a-descriptions-item>
          <a-descriptions-item label="对内备注">
            {{ remarkText }}
          </a-descriptions-item>
          <a-descriptions-item label="对外备注">
            -
          </a-descriptions-item>
        </a-descriptions>
      </div>

      <div class="tabs">
        <a-tabs
          v-model:active-key="activeStudentTabKey"
          size="large"
          :tab-bar-style="{ 'border-radius': '0px', 'padding-left': '24px' }"
        >
          <a-tab-pane key="students" tab="学员名单">
            <a-card :title="studentCardTitle" :bordered="false">
              <template v-if="!isOneToOne" #extra>
                <a-dropdown
                  placement="bottomRight"
                  :trigger="['hover']"
                  overlay-class-name="schedule-detail-add-student-dropdown"
                >
                  <a-button>
                    <span>添加学员</span>
                  </a-button>
                  <template #overlay>
                    <a-menu @click="handleAddStudentMenuClick">
                      <a-menu-item key="makeup">
                        补课学员
                      </a-menu-item>
                      <a-menu-item key="temporary">
                        临时学员
                      </a-menu-item>
                      <a-menu-item key="trial">
                        试听学员
                      </a-menu-item>
                    </a-menu>
                  </template>
                </a-dropdown>
              </template>
              <a-table
                :columns="studentColumns"
                :data-source="currentStudentList"
                :pagination="false"
                row-key="studentId"
                :scroll="{ x: 700 }"
              >
                <template #bodyCell="{ column, record }">
                  <template v-if="column.key === 'phone'">
                    <div class="student-phone-cell">
                      <div class="student-phone-relation">
                        {{ record.phoneRelationshipText || '-' }}
                      </div>
                      <div class="student-phone-number">
                        {{ record.maskedPhone || record.phone || '-' }}
                      </div>
                    </div>
                  </template>
                  <template v-else-if="column.key === 'studentName'">
                    <div class="student-name-cell" @click="handleViewStudent(record.studentId)">
                      <img class="student-avatar" :src="record.avatarUrl || defaultStudentAvatar" alt="">
                      <div class="student-name-text">
                        {{ record.studentName || '-' }}
                      </div>
                    </div>
                  </template>
                  <template v-else-if="column.key === 'studentType'">
                    <span class="student-type-tag" :class="getStudentTypeTagClass(record.scheduleStudentType)">
                      {{ record.scheduleStudentTypeText || (isOneToOne ? '1对1学员' : '班课学员') }}
                    </span>
                  </template>
                  <template v-else-if="column.key === 'action'">
                    <div v-if="!isOneToOne" class="student-action-links">
                      <button type="button" class="student-action-link" @click="handleViewStudent(record.studentId)">
                        详情
                      </button>
                      <span class="student-action-divider" />
                      <button type="button" class="student-action-link" @click="handleStudentReschedule(record)">
                        调课
                      </button>
                      <span class="student-action-divider" />
                      <button type="button" class="student-action-link" @click="handleStudentRemove(record)">
                        移出本节
                      </button>
                    </div>
                    <a-button v-else type="link" class="px0" @click="handleViewStudent(record.studentId)">
                      详情
                    </a-button>
                  </template>
                </template>
              </a-table>
            </a-card>
          </a-tab-pane>
          <a-tab-pane v-if="!isOneToOne" key="leave" tab="请假学员">
            <a-card :title="studentCardTitle" :bordered="false">
              <a-table
                :columns="studentColumns"
                :data-source="currentStudentList"
                :pagination="false"
                row-key="studentId"
                :scroll="{ x: 700 }"
              >
                <template #bodyCell="{ column, record }">
                  <template v-if="column.key === 'phone'">
                    <div class="student-phone-cell">
                      <div class="student-phone-relation">
                        {{ record.phoneRelationshipText || '-' }}
                      </div>
                      <div class="student-phone-number">
                        {{ record.maskedPhone || record.phone || '-' }}
                      </div>
                    </div>
                  </template>
                  <template v-else-if="column.key === 'studentName'">
                    <div class="student-name-cell" @click="handleViewStudent(record.studentId)">
                      <img class="student-avatar" :src="record.avatarUrl || defaultStudentAvatar" alt="">
                      <div class="student-name-text">
                        {{ record.studentName || '-' }}
                      </div>
                    </div>
                  </template>
                  <template v-else-if="column.key === 'studentType'">
                    <span class="student-type-tag" :class="getStudentTypeTagClass(record.scheduleStudentType)">
                      {{ record.scheduleStudentTypeText || (isOneToOne ? '1对1学员' : '班课学员') }}
                    </span>
                  </template>
                  <template v-else-if="column.key === 'action'">
                    <div class="student-action-links">
                      <button type="button" class="student-action-link" @click="handleViewStudent(record.studentId)">
                        详情
                      </button>
                      <span class="student-action-divider" />
                      <button type="button" class="student-action-link" @click="handleStudentReschedule(record)">
                        调课
                      </button>
                      <span class="student-action-divider" />
                      <button type="button" class="student-action-link" @click="handleStudentRemove(record)">
                        移出本节
                      </button>
                    </div>
                  </template>
                </template>
              </a-table>
            </a-card>
          </a-tab-pane>
        </a-tabs>
      </div>
    </a-spin>
    <student-info-drawer v-model:open="openStudentDrawer" />
    <RollCallDrawer
      v-model:open="rollCallDrawerOpen"
      :schedule-id="scheduleId"
      :lesson-day="detailData?.lessonDate || ''"
      @updated="loadDetail"
    />
    <ClassRecordDetails v-model:open="classRecordDrawerOpen" :teaching-record-id="detailData?.teachingRecordId || ''" />
    <RollCallAddStudentModal
      v-model:open="addStudentModalOpen"
      :title="addStudentModalTitle"
      :schedule-id="scheduleId"
      :student-type="addStudentType"
      @success="handleAddStudentSuccess"
    />
  </a-drawer>
</template>

<style lang="less" scoped>
@keyframes icon-rotate {
  from {
    transform: rotate(0deg);
  }

  to {
    transform: rotate(180deg);
  }
}

.close-btn {
  &:hover {
    background: transparent;

    .close-icon {
      animation: icon-rotate 0.3s linear;
    }
  }
}

.tabs {
  width: 100%;
  border-radius: 10px;

  :deep(.ant-tabs-nav) {
    background: #fff;
    margin: 0;
  }

  :deep(.ant-tabs-ink-bar) {
    text-align: center;
    height: 12px !important;
    background: transparent;
    bottom: 0 !important;

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

:deep(.tabs .ant-card-head) {
  min-height: 68px;
  padding: 0 24px;
}

:deep(.tabs .ant-card-head-title) {
  padding: 20px 0;
  color: #1f2329;
  font-size: 17px;
  font-weight: 700;
}

:deep(.tabs .ant-card-extra) {
  padding: 14px 0;
}

:global(.schedule-detail-add-student-dropdown .ant-dropdown-menu) {
  min-width: 132px;
  padding: 8px;
  border-radius: 16px;
  border: 1px solid #eef0f4;
  box-shadow: 0 12px 28px rgb(15 23 42 / 10%);
}

:global(.schedule-detail-add-student-dropdown .ant-dropdown-menu-item) {
  min-height: 40px;
  padding: 10px 14px;
  border-radius: 12px;
  color: #1f2329;
  font-size: 15px;
  font-weight: 500;
  line-height: 20px;
}

:global(.schedule-detail-add-student-dropdown .ant-dropdown-menu-item:hover) {
  background: #f5f7fb;
}

.student-name-cell {
  display: flex;
  align-items: center;
  gap: 12px;
  cursor: pointer;
}

.student-avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  object-fit: cover;
  flex: 0 0 36px;
}

.student-name-text {
  color: #222;
  font-weight: 500;
}

.student-name-cell:hover .student-name-text {
  color: #06f;
}

.student-phone-cell {
  line-height: 1.5;
}

.student-phone-relation {
  color: #222;
}

.student-phone-number {
  color: #888;
}

.student-type-tag {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 72px;
  padding: 4px 10px;
  border-radius: 999px;
  font-size: 12px;
  line-height: 1;
  white-space: nowrap;
}

.student-type-tag--member {
  color: #1668ff;
  background: #eaf3ff;
}

.student-type-tag--temporary {
  color: #ad6800;
  background: #fff3d6;
}

.student-type-tag--trial {
  color: #047857;
  background: #dcfce7;
}

.student-type-tag--makeup {
  color: #7c3aed;
  background: #f3e8ff;
}

.student-action-links {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 14px;
  white-space: nowrap;
}

.student-action-link {
  padding: 0;
  border: none;
  background: transparent;
  color: #1677ff;
  cursor: pointer;
}

.student-action-link:hover {
  color: #0958d9;
}

.student-action-divider {
  width: 1px;
  height: 16px;
  background: #e5e6eb;
}
</style>
