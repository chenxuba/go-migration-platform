<script setup>
import { computed, ref, watch } from 'vue'
import dayjs from 'dayjs'
import { CloseOutlined } from '@ant-design/icons-vue'
import { cancelLeaveApi, getLeaveDetailApi, getLeaveDetailSchedulesApi } from '@/api/home-center/leave'
import { Sex } from '@/enums'
import messageService from '@/utils/messageService'

const open = defineModel({
  type: Boolean,
  default: false,
})

const props = defineProps({
  leaveId: {
    type: [String, Number],
    default: '',
  },
})

const emit = defineEmits(['closed', 'changed'])

const defaultStudentAvatar = 'https://pcsys.admin.ybc365.com/a369a751-2be5-4929-974d-9ae4439f54c4.png'
const approvedStampUrl = 'https://prod-tbu-next-erp-cdn.schoolpal.cn/next-pc-static/static/15233/static/png/leave-approve-B7HZREJ9.png'
const revokedStampUrl = 'https://prod-tbu-next-erp-cdn.schoolpal.cn/next-pc-static/static/15233/static/png/leave-cancled-ClErrj5T.png'

const loading = ref(false)
const cancelLoading = ref(false)
const cancelModalOpen = ref(false)
const cancelRemark = ref('')
const detail = ref(createEmptyDetail())
const scheduleList = ref([])

let requestSeed = 0

const leaveInfo = computed(() => {
  const processes = buildLeaveProcessList(detail.value)
  return {
    studentName: detail.value.studentName || '-',
    studentPhone: detail.value.studentPhone || '-',
    studentAvatarUrl: detail.value.studentAvatarUrl || defaultStudentAvatar,
    studentGender: getStudentGender(detail.value.studentSex),
    startTime: formatDateTime(detail.value.startDate || detail.value.startTime),
    endTime: formatDateTime(detail.value.endDate || detail.value.endTime),
    leaveTitle: `请假课节 (${scheduleList.value.length})`,
    leaveClasses: scheduleList.value.map(item => ({
      time: formatScheduleTime(item),
      className: item.title || item.lessonName || '-',
      courseName: item.lessonName || '-',
      teacherName: item.mainTeacherName || item.teachers?.[0]?.teacherName || '-',
    })),
    leaveProcess: processes.map((item, index) => ({
      ...item,
      showLine: index < processes.length - 1,
    })),
    stampUrl: getStampUrl(detail.value.status),
    statusText: detail.value.statusText || '-',
    statusClass: getStatusClass(detail.value.status),
  }
})

const canCancelLeave = computed(() => Number(detail.value.status) === 2)

function createEmptyDetail() {
  return {
    id: '',
    studentId: '',
    studentName: '',
    studentPhone: '',
    studentAvatarUrl: '',
    studentSex: 0,
    startTime: '',
    endTime: '',
    startDate: '',
    endDate: '',
    isAgent: false,
    leaveType: 0,
    leaveTypeText: '',
    reason: '',
    proofMaterials: [],
    remark: '',
    status: 1,
    statusText: '',
    initiateStaffName: '',
    operatorId: '',
    operatorName: '',
    operatorAvatar: '',
    operationDate: '',
    currentApproverName: '',
    approverName: '',
    approve: null,
    approves: [],
    applyTime: '',
    schedules: [],
    processes: [],
  }
}

function resetState() {
  requestSeed += 1
  cancelModalOpen.value = false
  cancelLoading.value = false
  cancelRemark.value = ''
  detail.value = createEmptyDetail()
  scheduleList.value = []
}

function getStudentGender(studentSex) {
  if (Number(studentSex) === Sex.Female)
    return Sex.Female
  if (Number(studentSex) === Sex.Male)
    return Sex.Male
  return Sex.Unknown
}

function formatDateTime(value) {
  if (!value || String(value).startsWith('0001-01-01'))
    return '-'
  const parsed = dayjs(value)
  return parsed.isValid() ? parsed.format('YYYY-MM-DD HH:mm') : String(value).replace('T', ' ').slice(0, 16)
}

function formatTimeOnly(value) {
  if (!value)
    return '--:--'
  const parsed = dayjs(value)
  return parsed.isValid() ? parsed.format('HH:mm') : '--:--'
}

function formatMinutes(minutes) {
  const total = Number(minutes)
  if (!Number.isFinite(total) || total < 0)
    return '--:--'
  const hour = `${Math.floor(total / 60)}`.padStart(2, '0')
  const minute = `${total % 60}`.padStart(2, '0')
  return `${hour}:${minute}`
}

function formatScheduleTime(item) {
  const startDateText = item.startTime ? dayjs(item.startTime).format('YYYY-MM-DD') : (item.lessonDay ? dayjs(item.lessonDay).format('YYYY-MM-DD') : '')
  const startText = item.startTime ? formatTimeOnly(item.startTime) : formatMinutes(item.startMinutes)
  const endText = item.endTime ? formatTimeOnly(item.endTime) : formatMinutes(item.endMinutes)
  return startDateText ? `${startDateText} ${startText} ~ ${endText}` : `${startText} ~ ${endText}`
}

function formatInitiatorName(name, isAgent) {
  const normalized = String(name || '').trim()
  if (!normalized)
    return '-'
  if (!isAgent)
    return normalized
  if (normalized.includes('（代办）') || normalized.includes('(代办)'))
    return normalized
  return `${normalized}（代办）`
}

function normalizeProcessRemark(actionType, remark) {
  const normalized = String(remark || '').trim()
  if (!normalized)
    return ''
  if (Number(actionType) === 5 && normalized === '未配置审批人，系统自动通过')
    return ''
  return normalized
}

function getStampUrl(status) {
  switch (Number(status)) {
    case 2:
      return approvedStampUrl
    case 4:
      return revokedStampUrl
    default:
      return ''
  }
}

function getStatusClass(status) {
  switch (Number(status)) {
    case 2:
      return 'success'
    case 3:
      return 'danger'
    case 4:
      return 'warning'
    default:
      return 'info'
  }
}

function buildLeaveProcessList(detailData) {
  const approveList = Array.isArray(detailData.approves) ? detailData.approves : []
  const processList = approveList.map((item) => {
    const actionType = Number(item?.actionType)
    const isPending = Number(item?.approveStatus) === 1 && actionType !== 1 && !item?.operationDate
    const displayName = actionType === 1
      ? formatInitiatorName(item?.operatorName || detailData.operatorName || detailData.initiateStaffName, detailData.isAgent)
      : (item?.operatorName || '-')

    return {
      name: displayName,
      status: item?.approveStatusText || getProcessStatusText(actionType, isPending),
      time: isPending ? '待处理' : formatDateTime(item?.operationDate),
      timeLabel: actionType === 1 ? '发起时间' : isPending ? '当前状态' : '处理时间',
      nodeText: getProcessNodeText(actionType, isPending),
      nodeClass: getProcessNodeClass(actionType, isPending),
      tagClass: getProcessTagClass(actionType, isPending),
      remark: normalizeProcessRemark(actionType, item?.remark),
      warningText: getProcessWarningText(actionType),
    }
  })

  const currentApprove = detailData.approve
  const shouldAppendPending = Number(currentApprove?.approveStatus) === 1
    && Number(currentApprove?.actionType) !== 1
    && !currentApprove?.operationDate

  if (shouldAppendPending) {
    processList.push({
      name: currentApprove?.operatorName || detailData.currentApproverName || '-',
      status: currentApprove?.approveStatusText || '待处理',
      time: '待处理',
      timeLabel: '当前状态',
      nodeText: getProcessNodeText(currentApprove?.actionType, true),
      nodeClass: getProcessNodeClass(currentApprove?.actionType, true),
      tagClass: getProcessTagClass(currentApprove?.actionType, true),
      remark: '',
      warningText: '',
    })
  }

  if (processList.length)
    return processList

  return (Array.isArray(detailData.processes) ? detailData.processes : []).map((item, index) => {
    const actionType = Number(item?.actionType)
    const isPending = Boolean(item?.pending)
    return {
      name: actionType === 1
        ? formatInitiatorName(item?.name || detailData.operatorName || detailData.initiateStaffName, detailData.isAgent)
        : (item?.name || '-'),
      status: item?.status || getProcessStatusText(actionType, isPending),
      time: isPending ? '待处理' : formatDateTime(item?.actionTime),
      timeLabel: index === 0 ? '发起时间' : isPending ? '当前状态' : '处理时间',
      nodeText: getProcessNodeText(actionType, isPending),
      nodeClass: getProcessNodeClass(actionType, isPending),
      tagClass: getProcessTagClass(actionType, isPending),
      remark: normalizeProcessRemark(actionType, item?.remark),
      warningText: getProcessWarningText(actionType),
    }
  })
}

function getProcessNodeText(actionType, isPending) {
  if (isPending)
    return '批'

  switch (Number(actionType)) {
    case 1:
      return '发'
    case 3:
      return '拒'
    case 4:
      return '撤'
    default:
      return '批'
  }
}

function getProcessNodeClass(actionType, isPending) {
  if (isPending)
    return 'process-node-info'

  switch (Number(actionType)) {
    case 1:
      return 'process-node-info'
    case 3:
      return 'process-node-danger'
    case 4:
      return 'process-node-warning'
    default:
      return 'process-node-success'
  }
}

function getProcessTagClass(actionType, isPending) {
  if (isPending)
    return 'process-tag-info'

  switch (Number(actionType)) {
    case 1:
      return 'process-tag-info'
    case 3:
      return 'process-tag-danger'
    case 4:
      return 'process-tag-warning'
    default:
      return 'process-tag-success'
  }
}

function getProcessStatusText(actionType, isPending) {
  if (isPending)
    return '待处理'

  switch (Number(actionType)) {
    case 1:
      return '发起'
    case 3:
      return '已拒绝'
    case 4:
      return '已撤销'
    default:
      return '已通过'
  }
}

function getProcessWarningText(actionType) {
  if (Number(actionType) === 4)
    return '此操作不可撤销，请谨慎操作'
  return ''
}

async function fetchLeaveDetail() {
  if (!props.leaveId)
    return

  const currentSeed = ++requestSeed
  loading.value = true

  try {
    const detailRes = await getLeaveDetailApi({ id: String(props.leaveId) })
    if (currentSeed !== requestSeed)
      return

    if (detailRes.code !== 200) {
      messageService.error(detailRes.message || '获取请假详情失败')
      return
    }

    const detailResult = detailRes.result || {}
    detail.value = {
      ...createEmptyDetail(),
      ...detailResult,
      approve: detailResult.approve || null,
      approves: Array.isArray(detailResult.approves) ? detailResult.approves : [],
      processes: Array.isArray(detailResult.processes) ? detailResult.processes : [],
      schedules: Array.isArray(detailResult.schedules) ? detailResult.schedules : [],
    }

    const studentId = String(detail.value.studentId || '').trim()
    const startDateTime = String(detail.value.startDate || detail.value.startTime || '').trim()
    const endDateTime = String(detail.value.endDate || detail.value.endTime || '').trim()

    if (!studentId || !startDateTime || !endDateTime) {
      scheduleList.value = Array.isArray(detail.value.schedules) ? detail.value.schedules : []
      return
    }

    const scheduleRes = await getLeaveDetailSchedulesApi({
      pageRequestModel: {
        needTotal: true,
        pageSize: 1000,
        pageIndex: 1,
        skipCount: 0,
      },
      queryModel: {
        studentId,
        startDateTime,
        endDateTime,
        timeRangeSearchType: 1,
      },
      sortModel: {
        byStartDate: 1,
      },
    })

    if (currentSeed !== requestSeed)
      return

    if (scheduleRes.code === 200) {
      scheduleList.value = Array.isArray(scheduleRes.result?.list) ? scheduleRes.result.list : []
      return
    }

    scheduleList.value = Array.isArray(detail.value.schedules) ? detail.value.schedules : []
  }
  catch (error) {
    if (currentSeed !== requestSeed)
      return
    scheduleList.value = []
    detail.value = createEmptyDetail()
    messageService.error(error?.response?.data?.message || error?.message || '获取请假详情失败')
  }
  finally {
    if (currentSeed === requestSeed) {
      loading.value = false
    }
  }
}

function openCancelLeaveModal() {
  if (!canCancelLeave.value)
    return
  cancelRemark.value = ''
  cancelModalOpen.value = true
}

function closeCancelLeaveModal() {
  if (cancelLoading.value)
    return
  cancelModalOpen.value = false
}

async function handleCancelLeave() {
  const leaveID = String(detail.value.id || props.leaveId || '').trim()
  if (!leaveID)
    return

  cancelLoading.value = true
  try {
    const res = await cancelLeaveApi({
      id: leaveID,
      remark: cancelRemark.value.trim(),
    })

    if (res.code !== 200)
      throw new Error(res.message || '撤销请假失败')

    messageService.success('撤销请假成功')
    cancelModalOpen.value = false
    emit('changed')
    await fetchLeaveDetail()
  }
  catch (error) {
    messageService.error(error?.response?.data?.message || error?.message || '撤销请假失败')
  }
  finally {
    cancelLoading.value = false
  }
}

watch(
  () => [open.value, props.leaveId],
  ([visible, leaveId]) => {
    if (visible && leaveId) {
      fetchLeaveDetail()
      return
    }

    if (!visible) {
      resetState()
    }
  },
  { immediate: true },
)

function handleClose() {
  cancelModalOpen.value = false
  open.value = false
  emit('closed')
}
</script>

<template>
  <a-drawer
    v-model:open="open" :body-style="{ padding: '0', background: '#fff' }" :keyboard="false"
    :mask-closable="false" :closable="false" width="800px"
  >
    <template #title>
      <div class="custom-header flex justify-between h-4 flex-items-center">
        <div class="text-5">
          请假详情
        </div>
        <a-button type="text" class="close-btn" @click="handleClose">
          <template #icon>
            <CloseOutlined class="text-5 close-icon" />
          </template>
        </a-button>
      </div>
    </template>

    <div class="custom-content py-20px px-30px">
      <div class="student-info flex items-center justify-between mb-10px">
        <div class="flex items-center">
          <div class="avatar-container mr-15px">
            <div class="avatarBox w-16 h-16 relative">
              <img
                width="64" height="64" class=" rounded-100"
                :src="leaveInfo.studentAvatarUrl"
                alt=""
              >
              <img
                v-if="leaveInfo.studentGender === Sex.Female"
                width="24" height="24" class="absolute bottom-0 right-0"
                src="~@/assets/images/girl.png" alt=""
              >
              <img
                v-else-if="leaveInfo.studentGender === Sex.Male"
                width="24" height="24" class="absolute bottom-0 right-0"
                src="~@/assets/images/boy.png" alt=""
              >
              <img
                v-else
                width="24" height="24" class="absolute bottom-0 right-0"
                src="~@/assets/images/unknown.svg" alt=""
              >
            </div>
          </div>
          <div class="info-container">
            <div class="name text-20px font-medium mb-5px">
              {{ leaveInfo.studentName }}
            </div>
            <div class="phone text-16px font-450">
              {{ leaveInfo.studentPhone }}
            </div>
          </div>
        </div>
        <div class="status-action-wrap">
          <img
            v-if="leaveInfo.stampUrl"
            class="leave-stamp-image"
            :src="leaveInfo.stampUrl"
            alt=""
          >
          <div v-else class="status-text-chip" :class="leaveInfo.statusClass">
            {{ leaveInfo.statusText }}
          </div>
          <a-button v-if="canCancelLeave" danger class="revoke-leave-btn" @click="openCancelLeaveModal">
            撤销请假
          </a-button>
        </div>
      </div>

      <div class="time-info flex items-center gap-25% mb-25px">
        <div class="flex items-center gap-1 text-14px">
          <div class="label ">
            开始时间：
          </div>
          <div class="value text-#999">
            {{ leaveInfo.startTime }}
          </div>
        </div>
        <div class="flex items-center gap-1 .dark:">
          <div class="label">
            结束时间：
          </div>
          <div class="value text-#999">
            {{ leaveInfo.endTime }}
          </div>
        </div>
      </div>

      <div class=" mb-20px">
        <div class="text-20px font-700 mb-15px">
          {{ leaveInfo.leaveTitle }}
        </div>

        <div class="bg-#fafafa rounded-8px p-15px">
          <div v-for="(item, index) in leaveInfo.leaveClasses" :key="index" class="class-item">
            <div class="time-label flex items-center mb-10px">
              <div class="dot w-6px h-6px rounded-full bg-#0066ff mr-8px" />
              <div class="time text-15px font-medium">
                {{ item.time }}
              </div>
              <div class="class-name text-14px font-medium ml-10px">
                {{ item.className }}
              </div>
            </div>

            <div class="pl-2px">
              <div class="px-10px border-0px border-l-1px border-l-#e5e6eb border-solid">
                <div class="course mb-8px text-#999">
                  <span class="label text-14px ">上课课程：</span>
                  <span class="value text-14px ml-5px">{{ item.courseName }}</span>
                </div>
                <div class="text-#999">
                  <span class="label text-14px">上课教师：</span>
                  <span class="value text-14px ml-5px">{{ item.teacherName }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="">
        <div class="text-20px font-700 mb-15px">
          请假流程
        </div>

        <div class="process-timeline">
          <div
            v-for="(item, index) in leaveInfo.leaveProcess" :key="index" class="process-item flex relative pb-20px"
            :class="{ 'last-item': index === leaveInfo.leaveProcess.length - 1, 'with-line': item.showLine }"
          >
            <div class="flex items-start gap-20px flex-1">
              <div
                :class="[item.nodeClass]"
                class="process-node p-8px text-12px text-#fff rounded-50% w-28px h-28px flex items-center justify-center"
              >
                {{ item.nodeText }}
              </div>
              <div class="process-card bg-#f6f6f6 rounded-8px p-15px py-20px flex-1">
                <div class="flex items-center justify-between">
                  <div class="flex items-center gap-20px">
                    <span class="text-16px font-500">{{ item.name }}</span>
                    <div class="text-12px font-400 py-3px px-8px rounded-8px" :class="item.tagClass">
                      {{ item.status }}
                    </div>
                  </div>
                  <div class="text-#999">
                    {{ item.timeLabel }}：{{ item.time }}
                  </div>
                </div>

                <div v-if="item.remark" class="process-remark">
                  备注：{{ item.remark }}
                </div>

                <div v-if="item.warningText" class="process-warning">
                  <span class="process-warning-icon">!</span>
                  <span>{{ item.warningText }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </a-drawer>

  <a-modal
    v-model:open="cancelModalOpen"
    title="撤销请假"
    :mask-closable="false"
    :keyboard="false"
    :confirm-loading="cancelLoading"
    width="520px"
    @cancel="closeCancelLeaveModal"
  >
    <div class="revoke-leave-modal-content">
      <div class="revoke-leave-modal-title">
        撤销请假后
      </div>
      <div class="revoke-leave-modal-rule">
        1. 不影响已经标记为请假的上课记录
      </div>
      <div class="revoke-leave-modal-rule">
        2. 不影响已有的补课数据
      </div>
      <div class="revoke-leave-modal-rule">
        3. 撤销后本次请假不计入请假限制/次数
      </div>
      <div class="revoke-leave-modal-warning">
        谨慎操作，撤销后不可恢复
      </div>
      <a-textarea
        v-model:value="cancelRemark"
        :maxlength="200"
        :auto-size="{ minRows: 4, maxRows: 4 }"
        show-count
        placeholder="选填，备注最多200字"
      />
    </div>

    <template #footer>
      <div class="revoke-leave-modal-footer">
        <a-button @click="closeCancelLeaveModal">
          再想想
        </a-button>
        <a-button danger type="primary" :loading="cancelLoading" @click="handleCancelLeave">
          撤 销
        </a-button>
      </div>
    </template>
  </a-modal>
</template>

<style lang="less" scoped>
.close-btn {
  &:hover {
    background: transparent;
  }
}

.custom-content {
  height: calc(100vh - 56px);
  overflow-y: auto;
}

.status-action-wrap {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 12px;
}

.revoke-leave-btn {
  flex-shrink: 0;
}

.leave-stamp-image {
  display: block;
  width: 110px;
  height: 110px;
  object-fit: contain;
  flex-shrink: 0;
}

.class-item + .class-item {
  margin-top: 14px;
}

.process-item {
  &.last-item {
    padding-bottom: 0;
  }

  &.with-line::before {
    content: '';
    position: absolute;
    left: 14px;
    top: 54px;
    bottom: -10px;
    width: 1px;
    background-color: #ccc;
  }
}

.process-node {
  margin-top: 18px;
  position: relative;
  z-index: 1;
}

.process-card {
  display: flex;
  flex-direction: column;
}

.process-remark {
  margin-top: 14px;
  color: #666;
  font-size: 14px;
  line-height: 22px;
}

.process-warning {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 12px;
  color: #ff8a00;
  font-size: 14px;
  font-weight: 500;
  line-height: 22px;
}

.process-warning-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  border-radius: 50%;
  background: #ff8a00;
  color: #fff;
  font-size: 14px;
  font-weight: 700;
  flex-shrink: 0;
}

.status-text-chip {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 96px;
  min-height: 48px;
  padding: 0 16px;
  border-radius: 999px;
  font-size: 16px;
  font-weight: 600;
}

.revoke-leave-modal-content {
  color: #222;
}

.revoke-leave-modal-title {
  margin-bottom: 12px;
  font-size: 16px;
  font-weight: 600;
}

.revoke-leave-modal-rule {
  margin-bottom: 8px;
  line-height: 22px;
  color: #666;
}

.revoke-leave-modal-warning {
  margin: 16px 0 12px;
  padding: 10px 12px;
  border-radius: 8px;
  background: #fff7e6;
  color: #d46b08;
  line-height: 22px;
}

.revoke-leave-modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

.process-node-info {
  background-color: #0066ff;
}

.process-node-success {
  background-color: #00cc66;
}

.process-node-danger {
  background-color: #ff4d4f;
}

.process-node-warning {
  background-color: #ff8a00;
}

.process-tag-info {
  color: #0066ff;
  background-color: #e6f0ff;
}

.process-tag-success {
  color: #00cc66;
  background-color: #e6ffed;
}

.process-tag-danger {
  color: #ff4d4f;
  background-color: #fff1f0;
}

.process-tag-warning {
  color: #ff8a00;
  background-color: #fff7e6;
}

</style>
