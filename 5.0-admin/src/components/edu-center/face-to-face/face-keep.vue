<script setup lang="ts">
import dayjs from 'dayjs'
import { CloseOutlined, InfoCircleFilled, RightOutlined } from '@ant-design/icons-vue'
import {
  getFaceAttendanceTodayStatisticsApi,
  pageFaceAttendanceTodayDetailsApi,
  type FaceAttendanceTodayDetailItem,
  type FaceAttendanceRelatedScheduleItem,
  type FaceAttendanceTodayStatistics,
} from '@/api/edu-center/face'
import faceIcon from '@/assets/images/face.png'
import RollCallDrawer from '@/components/common/roll-call-drawer.vue'
import StudentAvatar from '@/components/common/StudentAvatar.vue'

const AUTO_ROLL_CALL_DELAY_MINUTES = 30
const router = useRouter()
type FaceAttendanceTodayDetailLike = FaceAttendanceTodayDetailItem | Record<string, any>

let faceWindow: Window | null = null

const stats = ref<FaceAttendanceTodayStatistics>({})
const statsLoading = ref(false)
const detailModalOpen = ref(false)
const detailLoading = ref(false)
const detailType = ref<'pending' | 'success' | 'success_unrolled'>('pending')
const detailSearchKey = ref('')
const detailList = ref<FaceAttendanceTodayDetailItem[]>([])
const detailPagination = ref({
  current: 1,
  pageSize: 10,
  total: 0,
})
const previewVisible = ref(false)
const previewImage = ref('')
const previewTitle = ref('')
const previewTimeText = ref('')
const scheduleModalOpen = ref(false)
const currentScheduleRecord = ref<FaceAttendanceTodayDetailItem | null>(null)
const rollCallDrawerOpen = ref(false)
const currentRollCallScheduleId = ref('')
const currentRollCallLessonDay = ref('')

const pendingColumns = [
  { title: '学员姓名', key: 'studentName', width: 220, fixed: 'left' as const },
  { title: '人脸采集', key: 'isCollect', width: 120 },
  { title: '今日上课时间', key: 'classTime', width: 220 },
  { title: '相关日程', key: 'scheduleName', width: 260 },
  { title: '待考勤状态', key: 'prompt', width: 140 },
  { title: '操作', key: 'action', width: 110, fixed: 'right' as const },
] as any[]

const successColumns = [
  { title: '学员姓名', key: 'studentName', width: 220, fixed: 'left' as const },
  { title: '考勤时间', key: 'attendanceTime', width: 180 },
  { title: '考勤方式', key: 'attendanceType', width: 120 },
  { title: '签到签退类型', key: 'signTypes', width: 180 },
  { title: '签到签退时间', key: 'signTimes', width: 220 },
  { title: '相关日程', key: 'scheduleName', width: 260 },
  { title: '操作', key: 'action', width: 110, fixed: 'right' as const },
] as any[]

const successUnrolledColumns = [
  { title: '学员姓名', key: 'studentName', width: 220, fixed: 'left' as const },
  { title: '考勤时间', key: 'attendanceTime', width: 180 },
  { title: '签到签退时间', key: 'signTimes', width: 220 },
  { title: '相关日程', key: 'scheduleName', width: 260 },
  { title: '未点名状态', key: 'prompt', width: 150 },
  { title: '操作', key: 'action', width: 110, fixed: 'right' as const },
] as any[]

const cardItems = computed(() => [
  {
    key: 'pending' as const,
    title: '待考勤',
    count: Number(stats.value.pendingCount || 0),
  },
  {
    key: 'success' as const,
    title: '考勤成功',
    count: Number(stats.value.successCount || 0),
  },
  {
    key: 'success_unrolled' as const,
    title: '待点名',
    count: Number(stats.value.successUnrolledCount || 0),
  },
])

const detailConfigMap = {
  pending: {
    title: '今日待考勤',
    note: (total: number) => `共 ${total} 条，今日已排课的学员全部待考勤日程（排除请假学员）`,
    empty: '今日暂无待考勤学员',
    width: 1180,
    columns: pendingColumns,
  },
  success: {
    title: '今日考勤成功',
    note: (total: number) => `共 ${total} 条，按学员当天考勤记录统计，一名学员当天最多记一条；表示今日已刷脸且已完成点名，不包含纯手动点名`,
    empty: '暂无考勤成功记录',
    width: 1360,
    columns: successColumns,
  },
  success_unrolled: {
    title: '今日待点名',
    note: (total: number) => `共 ${total} 条，按学员当天考勤记录统计，一名学员当天最多记一条；表示今日已刷脸但仍有日程未完成点名，超过课后${AUTO_ROLL_CALL_DELAY_MINUTES}分钟可直接去点名`,
    empty: '暂无未点名记录',
    width: 1280,
    columns: successUnrolledColumns,
  },
}

const currentDetailConfig = computed(() => detailConfigMap[detailType.value])
const currentColumns = computed(() => currentDetailConfig.value.columns)
const currentScrollX = computed(() => currentColumns.value.reduce((sum, item) => sum + Number(item.width || 0), 0))
const modalNote = computed(() => currentDetailConfig.value.note(detailPagination.value.total))
const tablePagination = computed(() => ({
  current: detailPagination.value.current,
  pageSize: detailPagination.value.pageSize,
  total: detailPagination.value.total,
  showSizeChanger: true,
  showQuickJumper: true,
  showTotal: (total: number) => `共 ${total} 条`,
}))
const relatedScheduleColumns = [
  {
    title: '上课时间',
    dataIndex: 'classTime',
    key: 'classTime',
    width: 240,
  },
  {
    title: '相关日程',
    dataIndex: 'scheduleName',
    key: 'scheduleName',
    width: 320,
  },
  {
    title: '点名状态',
    dataIndex: 'rollCallStatus',
    key: 'rollCallStatus',
    width: 160,
  },
  {
    title: '操作',
    dataIndex: 'action',
    key: 'action',
    width: 110,
  },
] as any[]
const relatedScheduleDataSource = computed<FaceAttendanceRelatedScheduleItem[]>(() => {
  if (!currentScheduleRecord.value)
    return []
  return getRelatedScheduleItems(currentScheduleRecord.value)
})

function handleFaceSign(type: number) {
  const newUrl = router.resolve({
    path: '/pc/face',
    query: { type },
  }).href

  if (faceWindow && !faceWindow.closed) {
    if (faceWindow.location.href !== newUrl) {
      faceWindow.close()
      faceWindow = window.open(newUrl, '_blank')
    }
    else {
      faceWindow.focus()
    }
  }
  else {
    faceWindow = window.open(newUrl, '_blank')
  }

  if (!faceWindow || faceWindow.closed)
    faceWindow = null
}

function formatDateTime(value?: string) {
  if (!value)
    return '-'
  const parsed = dayjs(value)
  return parsed.isValid() ? parsed.format('YYYY-MM-DD HH:mm') : '-'
}

function formatClassTime(record: FaceAttendanceTodayDetailLike) {
  if (record.classTime)
    return record.classTime
  const start = record.lessonStartAt ? dayjs(record.lessonStartAt) : null
  const end = record.lessonEndAt ? dayjs(record.lessonEndAt) : null
  if (!start?.isValid() || !end?.isValid())
    return '-'
  return `${start.format('YYYY-MM-DD')} ${start.format('HH:mm')} ~ ${end.format('HH:mm')}`
}

function normalizeStringArray(value?: string[]) {
  if (!Array.isArray(value))
    return []
  return value.map(item => String(item || '').trim()).filter(Boolean)
}

function getRelatedScheduleItems(record?: FaceAttendanceTodayDetailLike) {
  if (!record)
    return []
  if (Array.isArray(record.relatedScheduleItems) && record.relatedScheduleItems.length)
    return record.relatedScheduleItems
  const classTimes = normalizeStringArray(String(record.classTime || '').split('\n'))
  const scheduleNames = normalizeStringArray(String(record.scheduleName || '').split('\n'))
  const length = Math.max(classTimes.length, scheduleNames.length)
  return Array.from({ length }, (_, index) => ({
    scheduleId: length === 1 ? String(record.scheduleId || '') : '',
    classTime: classTimes[index] || '-',
    scheduleName: scheduleNames[index] || '-',
    rollCallStatus: record.type === 'success' ? '已点名' : '未点名',
  }))
}

function relatedScheduleCount(record?: FaceAttendanceTodayDetailLike) {
  return getRelatedScheduleItems(record).length
}

function relatedScheduleSummaryText(record?: FaceAttendanceTodayDetailLike) {
  const items = getRelatedScheduleItems(record)
  if (!items.length)
    return '-'
  if (items.length === 1)
    return items[0]?.scheduleName || '-'
  return `共${items.length}个日程`
}

function parseRelatedScheduleEndTime(value?: string) {
  const text = String(value || '').replace(/\n/g, ' ').trim()
  const matched = text.match(/(\d{4}-\d{2}-\d{2})\s+(\d{2}:\d{2})\s*~\s*(\d{2}:\d{2})/)
  if (!matched)
    return null
  const [, dateText, _startText, endText] = matched
  const parsed = dayjs(`${dateText} ${endText}`)
  return parsed.isValid() ? parsed : null
}

function isUnsignedSchedule(record?: FaceAttendanceRelatedScheduleItem) {
  return !!record && String(record.rollCallStatus || '').trim() === '未点名'
}

function getRelatedScheduleRollCallStage(record?: FaceAttendanceRelatedScheduleItem) {
  if (!isUnsignedSchedule(record))
    return 'normal'
  const endTime = parseRelatedScheduleEndTime(record?.classTime)
  if (!endTime)
    return 'normal'
  const now = dayjs()
  if (endTime.isAfter(now))
    return 'normal'
  const autoRollCallDeadline = endTime.add(AUTO_ROLL_CALL_DELAY_MINUTES, 'minute')
  return autoRollCallDeadline.isAfter(now) ? 'auto-pending' : 'manual'
}

function isManualRollCallSchedule(record?: FaceAttendanceRelatedScheduleItem) {
  return getRelatedScheduleRollCallStage(record) === 'manual'
}

function relatedScheduleStatusText(record?: FaceAttendanceRelatedScheduleItem) {
  if (isManualRollCallSchedule(record))
    return '待手动点名'
  if (getRelatedScheduleRollCallStage(record) === 'auto-pending')
    return '待自动点名'
  return String(record?.rollCallStatus || '-').trim() || '-'
}

function relatedScheduleStatusClass(record?: FaceAttendanceRelatedScheduleItem) {
  if (isManualRollCallSchedule(record))
    return 'roll-call-status roll-call-status--manual'
  if (getRelatedScheduleRollCallStage(record) === 'auto-pending')
    return 'roll-call-status roll-call-status--auto-pending'
  if (String(record?.rollCallStatus || '') === '已点名')
    return 'roll-call-status roll-call-status--signed'
  if (String(record?.rollCallStatus || '') === '未点名')
    return 'roll-call-status roll-call-status--unsigned'
  return 'roll-call-status'
}

function relatedScheduleHintText(record?: FaceAttendanceRelatedScheduleItem) {
  if (isManualRollCallSchedule(record))
    return `已超过课后${AUTO_ROLL_CALL_DELAY_MINUTES}分钟`
  if (getRelatedScheduleRollCallStage(record) === 'auto-pending')
    return `课后${AUTO_ROLL_CALL_DELAY_MINUTES}分钟内自动点名`
  return ''
}

function getRelatedScheduleLessonDay(record?: FaceAttendanceRelatedScheduleItem) {
  const matched = String(record?.classTime || '').replace(/\n/g, ' ').match(/(\d{4}-\d{2}-\d{2})/)
  return matched?.[1] || ''
}

function formatSingleLineScheduleTime(value?: string) {
  return String(value || '').replace(/\n/g, ' ').trim() || '-'
}

function getPrimaryManualRollCallSchedule(record?: FaceAttendanceTodayDetailLike) {
  return getRelatedScheduleItems(record).find(item => isManualRollCallSchedule(item) && item.scheduleId)
}

function openRelatedScheduleModal(record: FaceAttendanceTodayDetailLike) {
  currentScheduleRecord.value = record as FaceAttendanceTodayDetailItem
  scheduleModalOpen.value = true
}

function handleGoRollCall(record?: FaceAttendanceRelatedScheduleItem) {
  const scheduleId = String(record?.scheduleId || '').trim()
  if (!scheduleId)
    return
  currentRollCallScheduleId.value = scheduleId
  currentRollCallLessonDay.value = getRelatedScheduleLessonDay(record)
  rollCallDrawerOpen.value = true
}

function isPendingSignOut(record: FaceAttendanceTodayDetailLike) {
  return !!record.hasFaceSession && Number(record.sessionStatus) === 1 && !record.signOutTime
}

function getAttendancePhoto(record: FaceAttendanceTodayDetailLike, action: 'sign_in' | 'sign_out' = 'sign_in') {
  if (action === 'sign_out')
    return String(record.signOutImage || '').trim()
  return String(record.signInImage || '').trim()
}

function getAttendancePhotoTitle(action: 'sign_in' | 'sign_out') {
  return action === 'sign_out' ? '签退照片' : '签到照片'
}

function getPreferredPreviewAction(record: FaceAttendanceTodayDetailLike) {
  if (getAttendancePhoto(record, 'sign_in'))
    return 'sign_in' as const
  if (getAttendancePhoto(record, 'sign_out'))
    return 'sign_out' as const
  return null
}

function openAttendancePhoto(record: FaceAttendanceTodayDetailLike, action: 'sign_in' | 'sign_out' = 'sign_in') {
  const image = getAttendancePhoto(record, action)
  if (!image)
    return
  previewImage.value = image
  previewTitle.value = getAttendancePhotoTitle(action)
  previewTimeText.value = formatDateTime(action === 'sign_out' ? record.signOutTime : record.signInTime)
  previewVisible.value = true
}

function handlePreviewAttendance(record: FaceAttendanceTodayDetailLike) {
  const action = getPreferredPreviewAction(record)
  if (!action)
    return
  openAttendancePhoto(record, action)
}

function getAttendancePrimaryLabel(record: FaceAttendanceTodayDetailLike) {
  if (record.hasFaceSession)
    return '刷脸签到'
  return '手动到课'
}

function getAttendanceSecondaryLabel(record: FaceAttendanceTodayDetailLike) {
  if (!record.hasFaceSession)
    return ''
  if (record.signOutTime)
    return '刷脸签退'
  if (isPendingSignOut(record))
    return '待签退'
  return ''
}

function pendingActionText(record: FaceAttendanceTodayDetailLike) {
  return record.isCollect ? '去考勤' : '去采集'
}

function handlePendingAction(record: FaceAttendanceTodayDetailLike) {
  handleFaceSign(record.isCollect ? 2 : 1)
}

function handleSuccessUnrolledAction(record: FaceAttendanceTodayDetailLike) {
  const manualSchedule = getPrimaryManualRollCallSchedule(record)
  if (manualSchedule?.scheduleId && relatedScheduleCount(record) === 1) {
    currentRollCallScheduleId.value = String(manualSchedule.scheduleId || '')
    currentRollCallLessonDay.value = getRelatedScheduleLessonDay(manualSchedule) || getLessonDay(record)
    rollCallDrawerOpen.value = true
    return
  }
  if (manualSchedule) {
    openRelatedScheduleModal(record)
    return
  }
  handlePreviewAttendance(record)
}

function getLessonDay(record: FaceAttendanceTodayDetailLike) {
  const value = record.lessonStartAt || record.classTime || ''
  const parsed = dayjs(value)
  if (parsed.isValid())
    return parsed.format('YYYY-MM-DD')
  const matched = String(record.classTime || '').match(/(\d{4}-\d{2}-\d{2})/)
  return matched?.[1] || ''
}

async function loadStats() {
  statsLoading.value = true
  try {
    const res = await getFaceAttendanceTodayStatisticsApi()
    if (res.code === 200)
      stats.value = res.result || {}
    else
      stats.value = {}
  }
  catch (error) {
    console.error('load face attendance today statistics failed', error)
    stats.value = {}
  }
  finally {
    statsLoading.value = false
  }
}

async function loadDetailList() {
  if (!detailModalOpen.value)
    return
  detailLoading.value = true
  try {
    const res = await pageFaceAttendanceTodayDetailsApi({
      pageRequestModel: {
        pageSize: detailPagination.value.pageSize,
        pageIndex: detailPagination.value.current,
      },
      queryModel: {
        type: detailType.value,
        searchKey: detailSearchKey.value,
      },
    })
    if (res.code === 200) {
      detailList.value = Array.isArray(res.result) ? res.result : []
      detailPagination.value.total = Number(res.total || 0)
      return
    }
    detailList.value = []
    detailPagination.value.total = 0
  }
  catch (error) {
    console.error('load face attendance today details failed', error)
    detailList.value = []
    detailPagination.value.total = 0
  }
  finally {
    detailLoading.value = false
  }
}

function openDetailModal(type: 'pending' | 'success' | 'success_unrolled') {
  detailType.value = type
  detailSearchKey.value = ''
  detailPagination.value.current = 1
  detailModalOpen.value = true
  loadDetailList()
}

function handleDetailTableChange(page: { current?: number, pageSize?: number }) {
  const nextCurrent = Number(page.current || 1)
  const nextSize = Number(page.pageSize || detailPagination.value.pageSize)
  const sizeChanged = nextSize !== detailPagination.value.pageSize
  detailPagination.value.pageSize = nextSize
  detailPagination.value.current = sizeChanged ? 1 : nextCurrent
  loadDetailList()
}

function handleSearch() {
  detailPagination.value.current = 1
  loadDetailList()
}

async function handleRollCallUpdated() {
  await Promise.all([loadStats(), loadDetailList()])
}

function closeDetailModal() {
  detailModalOpen.value = false
}

function closePreviewModal() {
  previewVisible.value = false
}

watch(() => detailSearchKey.value, (value) => {
  if (!detailModalOpen.value || value)
    return
  detailPagination.value.current = 1
  loadDetailList()
})

onMounted(() => {
  loadStats()
})
</script>

<template>
  <div class="bg-white h-80vh flex justify-center flex-wrap items-center">
    <div class="flex">
      <div
        class="h-56 mr-6 hover-shadow w-55 flex items-center flex-col border-1 border-color-#0000000d border-solid rounded-2 cursor-pointer"
        @click="handleFaceSign(1)"
      >
        <img
          width="120" height="120" class="mb-2"
          src="https://pcsys.admin.ybc365.com//e7cb1394-1c75-47ec-b37c-ad95f6863504.png" alt=""
        >
        <div class="text-4 text-#222 font-500 mb-1">
          人脸采集
        </div>
        <div class="text-3 text-#222">
          先采集，再考勤
        </div>
      </div>
      <div
        class="h-56 mr-6 hover-shadow w-55 flex items-center flex-col border-1 border-color-#0000000d border-solid rounded-2 cursor-pointer"
        @click="handleFaceSign(2)"
      >
        <img
          width="120" height="120" class="mb-2"
          src="https://pcsys.admin.ybc365.com//1300b671-9022-4b3f-9cc6-a1deec75d52e.png" alt=""
        >
        <div class="text-4 text-#222 font-500 mb-1">
          人脸考勤
        </div>
        <div class="text-3 text-#222">
          识别自动记录，支持多人识别
        </div>
      </div>
      <div class="h-56 mr-6 hover-shadow w-55 flex-center flex-col border-1 border-color-#0000000d border-solid rounded-2 cursor-pointer">
        <img
          width="73" height="73" class="mb-2"
          src="https://prod-tbu-next-erp-cdn.schoolpal.cn/next-pc-static/static/12087/static/more-dian-icon.4505caef.png"
          alt=""
        >
        <div class="text-4 text-#222 font-500 mb-1">
          人脸考勤设备
        </div>
        <div class="text-3 text-#06f">
          未激活设备，点此激活 >
        </div>
      </div>
    </div>
    <div class="w-88 h-83 total">
      <div class="t flex justify-between">
        <span>今日考勤统计</span>
        <a-tooltip title="按学员当天考勤记录统计，一名学员当天最多记一条；纯手动点名不计入考勤成功和待点名">
          <InfoCircleFilled class="text-#c5cee0 cursor-pointer hover-text-#06f" />
        </a-tooltip>
      </div>
      <div class="mt-8">
        <div
          v-for="item in cardItems"
          :key="item.key"
          class="mb-3.5 bg-white rounded-2 py-4 px-3 flex justify-between flex-center cursor-pointer items"
          @click="openDetailModal(item.key)"
        >
          <span>{{ item.title }}</span>
          <span class="num flex flex-items-center">
            <a-spin v-if="statsLoading" size="small" class="mr-2" />
            <template v-else>
              {{ item.count }}
            </template>
            <RightOutlined class="text-3 text-#ccc ml-1" />
          </span>
        </div>
      </div>
    </div>
  </div>

  <a-modal
    v-model:open="detailModalOpen"
    centered
    class="modal-content-box face-stat-modal"
    :keyboard="false"
    :closable="false"
    :mask-closable="false"
    :width="currentDetailConfig.width"
    :footer="null"
    destroy-on-close
  >
    <template #title>
      <div class="text-5 flex justify-between flex-center">
        <span>{{ currentDetailConfig.title }}</span>
        <a-button type="text" class="close-btn" @click="closeDetailModal">
          <template #icon>
            <CloseOutlined class="text-5 close-icon" />
          </template>
        </a-button>
      </div>
    </template>
    <div class="contenter scrollbar face-stat-modal__body">
      <div class="detail-note">
        <InfoCircleFilled class="mr-8px text-#1677ff" />
        <span>{{ modalNote }}</span>
      </div>
      <div class="face-stat-modal__search">
        <a-input-search
          v-model:value="detailSearchKey"
          allow-clear
          size="large"
          placeholder="请输入学员姓名"
          @search="handleSearch"
        />
      </div>
      <a-table
        row-key="id"
        :loading="detailLoading"
        :data-source="detailList"
        :columns="currentColumns"
        :pagination="tablePagination"
        :scroll="{ x: currentScrollX }"
        :locale="{ emptyText: currentDetailConfig.empty }"
        size="small"
        @change="handleDetailTableChange"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'studentName'">
            <StudentAvatar
              :id="record.studentId"
              :name="record.studentName || '-'"
              :avatar-url="record.avatarUrl"
              :phone="record.studentMobile || ''"
              :show-gender="false"
              :show-age="false"
              :auto-width="false"
            />
          </template>

          <template v-else-if="column.key === 'isCollect'">
            <span :class="record.isCollect ? 'text-#333' : 'text-#999'">
              {{ record.isCollect ? '已采集' : '未采集' }}
            </span>
          </template>

          <template v-else-if="column.key === 'attendanceTime'">
            {{ formatDateTime(record.attendanceTime) }}
          </template>

          <template v-else-if="column.key === 'attendanceType'">
            {{ record.attendanceType || '-' }}
          </template>

          <template v-else-if="column.key === 'classTime'">
            <span class="whitespace-pre-line">{{ formatClassTime(record) }}</span>
          </template>

          <template v-else-if="column.key === 'scheduleName'">
            <div class="schedule-name-block">
              <span class="schedule-name" :title="relatedScheduleSummaryText(record)">
                {{ relatedScheduleSummaryText(record) }}
              </span>
              <a-button
                v-if="relatedScheduleCount(record) > 1 || detailType === 'success_unrolled'"
                type="link"
                size="small"
                class="px-0"
                @click.stop="openRelatedScheduleModal(record)"
              >
                查看
              </a-button>
            </div>
          </template>

          <template v-else-if="column.key === 'prompt'">
            <span :class="record.canManualRollCall ? 'prompt-badge prompt-badge--manual' : 'prompt-badge'">
              {{ record.prompt || '-' }}
            </span>
          </template>

          <template v-else-if="column.key === 'signTypes'">
            <div class="attendance-action-cell">
              <div class="attendance-entry attendance-entry--sign-in">
                <span class="attendance-entry__tag">签到</span>
                <span class="attendance-entry__text">{{ getAttendancePrimaryLabel(record) }}</span>
                <img
                  v-if="getAttendancePhoto(record, 'sign_in')"
                  class="attendance-photo-trigger"
                  :src="faceIcon"
                  alt=""
                  @click.stop="openAttendancePhoto(record, 'sign_in')"
                >
              </div>
              <div
                v-if="getAttendanceSecondaryLabel(record)"
                class="attendance-entry"
                :class="record.signOutTime ? 'attendance-entry--sign-out' : 'attendance-entry--pending'"
              >
                <span class="attendance-entry__tag">签退</span>
                <span class="attendance-entry__text">{{ getAttendanceSecondaryLabel(record) }}</span>
                <img
                  v-if="record.signOutTime && getAttendancePhoto(record, 'sign_out')"
                  class="attendance-photo-trigger"
                  :src="faceIcon"
                  alt=""
                  @click.stop="openAttendancePhoto(record, 'sign_out')"
                >
              </div>
            </div>
          </template>

          <template v-else-if="column.key === 'signTimes'">
            <div class="attendance-action-cell">
              <div class="attendance-entry attendance-entry--sign-in">
                <span class="attendance-entry__tag">签到</span>
                <span class="attendance-entry__text attendance-entry__text--time">
                  {{ record.hasFaceSession ? formatDateTime(record.signInTime) : formatDateTime(record.attendanceTime) }}
                </span>
              </div>
              <div
                v-if="record.signOutTime || isPendingSignOut(record)"
                class="attendance-entry"
                :class="record.signOutTime ? 'attendance-entry--sign-out' : 'attendance-entry--pending'"
              >
                <span class="attendance-entry__tag">签退</span>
                <span class="attendance-entry__text attendance-entry__text--time">
                  {{ record.signOutTime ? formatDateTime(record.signOutTime) : '待签退' }}
                </span>
              </div>
            </div>
          </template>

          <template v-else-if="column.key === 'action'">
            <template v-if="detailType === 'pending'">
              <a-button type="link" size="small" class="px-0" @click.stop="handlePendingAction(record)">
                {{ pendingActionText(record) }}
              </a-button>
            </template>
            <template v-else-if="detailType === 'success'">
              -
            </template>
            <template v-else>
              <a-button
                v-if="getPrimaryManualRollCallSchedule(record)"
                type="link"
                size="small"
                class="px-0"
                @click.stop="handleSuccessUnrolledAction(record)"
              >
                {{ relatedScheduleCount(record) > 1 ? '查看日程' : '去点名' }}
              </a-button>
              <a-button
                v-else-if="getPreferredPreviewAction(record)"
                type="link"
                size="small"
                class="px-0"
                @click.stop="handleSuccessUnrolledAction(record)"
              >
                查看照片
              </a-button>
              <template v-else>
                -
              </template>
            </template>
          </template>
        </template>
      </a-table>
    </div>
  </a-modal>

  <RollCallDrawer
    v-model:open="rollCallDrawerOpen"
    :schedule-id="currentRollCallScheduleId"
    :lesson-day="currentRollCallLessonDay"
    @updated="handleRollCallUpdated"
    @confirmed="handleRollCallUpdated"
  />

  <a-modal
    v-model:open="scheduleModalOpen"
    title="相关日程"
    width="920px"
    :footer="null"
    destroy-on-close
  >
    <a-table
      :columns="relatedScheduleColumns"
      :data-source="relatedScheduleDataSource"
      :pagination="false"
      :scroll="{ x: 860 }"
      size="small"
      :row-key="record => `${record.scheduleId || ''}-${record.classTime || ''}-${record.scheduleName || ''}`"
    >
        <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'classTime'">
          <span class="related-schedule-time">{{ formatSingleLineScheduleTime(record.classTime) }}</span>
        </template>
        <template v-else-if="column.key === 'scheduleName'">
          <div class="related-schedule-name-cell">
            <span class="schedule-name" :title="record.scheduleName || '-'">{{ record.scheduleName || '-' }}</span>
            <span v-if="relatedScheduleHintText(record)" class="related-schedule-hint">
              {{ relatedScheduleHintText(record) }}
            </span>
          </div>
        </template>
        <template v-else-if="column.key === 'rollCallStatus'">
          <span :class="relatedScheduleStatusClass(record)">
            {{ relatedScheduleStatusText(record) }}
          </span>
        </template>
        <template v-else-if="column.key === 'action'">
          <a-button
            v-if="isManualRollCallSchedule(record) && record.scheduleId"
            type="link"
            size="small"
            class="px-0"
            @click="handleGoRollCall(record)"
          >
            去点名
          </a-button>
          <template v-else>
            -
          </template>
        </template>
      </template>
    </a-table>
  </a-modal>

  <a-modal
    v-model:open="previewVisible"
    centered
    class="modal-content-box face-stat-modal"
    :keyboard="false"
    :closable="false"
    :mask-closable="false"
    :footer="null"
  >
    <template #title>
      <div class="text-5 flex justify-between flex-center">
        <span>{{ previewTitle }}</span>
        <a-button type="text" class="close-btn" @click="closePreviewModal">
          <template #icon>
            <CloseOutlined class="text-5 close-icon" />
          </template>
        </a-button>
      </div>
    </template>
    <div class="contenter face-stat-modal__preview">
      <div class="attendance-preview">
        <img alt="attendance" class="attendance-preview__image" :src="previewImage">
        <div v-if="previewTimeText" class="attendance-preview__stamp">
          {{ previewTimeText }}
        </div>
      </div>
    </div>
  </a-modal>
</template>

<style lang="less" scoped>
.total {
  background: url('https://prod-tbu-next-erp-cdn.schoolpal.cn/next-pc-static/static/12087/static/today-data.82d0298c.png');
  background-size: contain;
  padding: 24px 48px 24px;
  border-radius: 16px;

  .t {
    font-family: PingFangSC-Medium, PingFang SC, sans-serif;
    font-size: 18px;
    font-weight: 500;
    color: #222;
  }

  .num {
    font-family: DINAlternate-Bold, DINAlternate, sans-serif;
    font-size: 20px;
    font-weight: bold;
  }

  .items {
    &:hover {
      .num {
        color: var(--pro-ant-color-primary);

        :deep(svg) {
          color: var(--pro-ant-color-primary) !important;
        }
      }
    }
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

.face-stat-modal__body {
  background: #fff;
  padding: 24px;
}

.face-stat-modal__search {
  margin: 16px 0;
}

.face-stat-modal__preview {
  background: #fff;
  padding: 24px;
}

.detail-note {
  display: flex;
  align-items: center;
  min-height: 40px;
  padding: 0 16px;
  border-radius: 8px;
  background: #e8f3ff;
  color: #1677ff;
  font-size: 14px;
}

.schedule-name {
  display: inline-block;
  max-width: 100%;
  color: #1f2937;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.schedule-name-block {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}

.related-schedule-name-cell {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.related-schedule-hint {
  color: #ad6800;
  font-size: 12px;
  line-height: 18px;
}

.roll-call-status {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 72px;
  min-height: 24px;
  padding: 0 10px;
  border-radius: 999px;
  background: #f5f5f5;
  color: #666;
  font-size: 12px;
  line-height: 24px;
  white-space: nowrap;
}

.roll-call-status--signed {
  color: #1f8f55;
  background: #edf9f2;
}

.roll-call-status--unsigned {
  color: #ad6800;
  background: #fff7e6;
}

.roll-call-status--auto-pending {
  color: #ad6800;
  background: #fff7e6;
}

.roll-call-status--manual {
  color: #cf1322;
  background: #fff1f0;
  box-shadow: inset 0 0 0 1px #ffccc7;
  font-weight: 600;
}

.prompt-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 24px;
  padding: 0 10px;
  border-radius: 999px;
  background: #fff7e6;
  color: #ad6800;
  font-size: 12px;
  line-height: 24px;
  white-space: nowrap;
}

.prompt-badge--manual {
  background: #fff1f0;
  color: #cf1322;
  box-shadow: inset 0 0 0 1px #ffccc7;
  font-weight: 600;
}

.attendance-action-cell {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.attendance-entry {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  min-height: 22px;
}

.attendance-entry__tag {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 34px;
  height: 20px;
  padding: 0 8px;
  border-radius: 999px;
  font-size: 12px;
  line-height: 20px;
  font-weight: 600;
  white-space: nowrap;
}

.attendance-entry__text {
  color: #334155;
  font-size: 13px;
  line-height: 20px;
}

.attendance-entry__text--time {
  font-variant-numeric: tabular-nums;
}

.attendance-entry--sign-in .attendance-entry__tag {
  color: #1668dc;
  background: #e8f3ff;
}

.attendance-entry--sign-out .attendance-entry__tag {
  color: #1f8f55;
  background: #edf9f2;
}

.attendance-entry--pending .attendance-entry__tag {
  color: #ad6800;
  background: #fff7e6;
}

.attendance-photo-trigger {
  width: 16px;
  height: 16px;
  cursor: pointer;
  flex: 0 0 auto;
}

.attendance-preview {
  position: relative;
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 260px;
}

.attendance-preview__image {
  width: 100%;
  max-height: 60vh;
  object-fit: contain;
  border-radius: 12px;
}

.attendance-preview__stamp {
  position: absolute;
  right: 16px;
  bottom: 16px;
  padding: 4px 10px;
  border-radius: 999px;
  background: rgba(15, 23, 42, 0.7);
  color: #fff;
  font-size: 12px;
  line-height: 20px;
}

@keyframes icon-rotate {
  from {
    transform: rotate(0deg);
  }

  to {
    transform: rotate(180deg);
  }
}
</style>

<style scoped>
.modal-content-box .ant-modal-header {
  padding: 10px 16px !important;
  margin-bottom: 0;
}

.modal-content-box .ant-modal-body {
  padding: 0 !important;
}
</style>
