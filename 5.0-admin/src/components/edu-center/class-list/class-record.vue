<script setup lang="ts">
import { CloseCircleOutlined, ExclamationCircleFilled, ExclamationCircleOutlined } from '@ant-design/icons-vue'
import dayjs, { type Dayjs } from 'dayjs'
import type { TableColumnsType } from 'ant-design-vue'
import { computed, ref, watch } from 'vue'
import { deleteTeachingRecordApi, getScheduleTeachingRecordPagedListApi, getTeachingRecordDetailApi, type ScheduleTeachingRecordItem, type TeachingRecordDetailResult } from '@/api/edu-center/class-record'
import ClassRecordDetails from '@/components/common/class-record-details.vue'
import EditRollNameModal from '@/components/common/edit-roll-name-modal.vue'
import messageService from '@/utils/messageService'

const props = withDefaults(defineProps<{
  open?: boolean
  classId?: string
  className?: string
}>(), {
  open: false,
  classId: '',
  className: '',
})

const displayArray = ref(['classEndingTime', 'classStopTime'])
const columns: TableColumnsType<any> = [
  {
    title: '上课日期/时段',
    dataIndex: 'date',
    key: 'date',
    fixed: 'left',
    width: 200,
  },
  {
    title: '授课课时',
    dataIndex: 'classTime',
    key: 'classTime',
    width: 120,
  },
  {
    title: '学员消耗课时',
    dataIndex: 'studentClassTime',
    key: 'studentClassTime',
    width: 140,
  },
  {
    title: '消耗学费',
    dataIndex: 'consumeFee',
    key: 'consumeFee',
    width: 120,
  },
  {
    title: '上课教师',
    dataIndex: 'teacher',
    key: 'teacher',
    width: 120,
  },
  {
    title: '上课助教',
    dataIndex: 'assistant',
    key: 'assistant',
    width: 120,
  },
  {
    title: '出勤率',
    dataIndex: 'attendanceRate',
    key: 'attendanceRate',
    width: 170,
  },
  {
    title: '到课（人）',
    dataIndex: 'attendance',
    key: 'attendance',
    width: 120,
  },
  {
    title: '请假（人）',
    dataIndex: 'leave',
    key: 'leave',
    width: 120,
  },
  {
    title: '旷课（人）',
    dataIndex: 'absent',
    key: 'absent',
    width: 120,
  },
  {
    title: '未记录（人）',
    dataIndex: 'unrecorded',
    key: 'unrecorded',
    width: 120,
  },
  {
    title: '点名时间',
    dataIndex: 'rollCallTime',
    key: 'rollCallTime',
    width: 160,
  },
  {
    title: '操作',
    dataIndex: 'action',
    key: 'action',
    width: 180,
    fixed: 'right',
  },
]

const loading = ref(false)
const deleting = ref(false)
const rawList = ref<ScheduleTeachingRecordItem[]>([])
const scheduleDateRange = ref<[Dayjs, Dayjs] | null>(null)
const createDateRange = ref<[Dayjs, Dayjs] | null>(null)
const recordDrawerOpen = ref(false)
const currentTeachingRecordId = ref('')
const deleteModalOpen = ref(false)
const deletingTeachingRecordId = ref('')
const editRollNameOpen = ref(false)
const editRollNameLoading = ref(false)
const editRollNameDetail = ref<TeachingRecordDetailResult | null>(null)

const totalWidth = computed(() =>
  columns.reduce((acc, col) => acc + Number(col.width || 0), 0),
)

const displaySummary = computed(() => ({
  total: filteredList.value.length,
  totalClassTimes: filteredList.value.reduce((sum, item) => sum + Number(item.actualQuantity || 0), 0),
  totalTeacherTimes: filteredList.value.reduce((sum, item) => sum + Number(item.teacherClassTime || 0), 0),
  totalTuition: filteredList.value.reduce((sum, item) => sum + Number(item.actualTuition || 0), 0),
}))

const filteredList = computed(() => {
  return rawList.value.filter((item) => {
    const start = dayjs(String(item.startTime || '').trim())
    const created = dayjs(String(item.createdTime || '').trim())
    const matchSchedule = !scheduleDateRange.value || (
      start.isValid()
      && (start.isAfter(scheduleDateRange.value[0], 'day') || start.isSame(scheduleDateRange.value[0], 'day'))
      && (start.isBefore(scheduleDateRange.value[1], 'day') || start.isSame(scheduleDateRange.value[1], 'day'))
    )
    const matchCreate = !createDateRange.value || (
      created.isValid()
      && (created.isAfter(createDateRange.value[0], 'day') || created.isSame(createDateRange.value[0], 'day'))
      && (created.isBefore(createDateRange.value[1], 'day') || created.isSame(createDateRange.value[1], 'day'))
    )
    return matchSchedule && matchCreate
  })
})

async function loadList() {
  const classId = String(props.classId || '').trim()
  if (!props.open || !classId) {
    rawList.value = []
    return
  }
  loading.value = true
  try {
    const res = await getScheduleTeachingRecordPagedListApi({
      queryModel: {
        classIds: [classId],
      },
      pageRequestModel: {
        needTotal: true,
        pageIndex: 1,
        pageSize: 500,
        skipCount: 0,
      },
      sortModel: {
        startTime: 2,
        updatedTime: 0,
      },
    })
    if (res.code !== 200)
      throw new Error(res.message || '加载上课记录失败')
    rawList.value = Array.isArray(res.result?.list) ? res.result.list : []
  }
  catch (error: any) {
    rawList.value = []
    messageService.error(error?.response?.data?.message || error?.message || '加载上课记录失败')
  }
  finally {
    loading.value = false
  }
}

function handleClassEndingTimeFilter(value: unknown) {
  scheduleDateRange.value = normalizeDayRange(value)
}

function handleClassStopTimeFilter(value: unknown) {
  createDateRange.value = normalizeDayRange(value)
}

async function handleEditRollCall(record: ScheduleTeachingRecordItem | Record<string, any>) {
  const teachingRecordId = String(record.teachingRecordId || '').trim()
  if (!teachingRecordId) {
    messageService.info('当前记录缺少上课记录，暂不可编辑点名')
    return
  }
  if (editRollNameLoading.value)
    return
  editRollNameLoading.value = true
  try {
    const res = await getTeachingRecordDetailApi({ teachingRecordId })
    if (res.code !== 200)
      throw new Error(res.message || '加载编辑点名详情失败')
    const detail = res.result
    if (!detail || !String(detail.teachingRecordId || '').trim())
      throw new Error('当前上课记录不存在')
    editRollNameDetail.value = detail
    editRollNameOpen.value = true
  }
  catch (error: any) {
    editRollNameDetail.value = null
    messageService.error(error?.response?.data?.message || error?.message || '加载编辑点名详情失败')
  }
  finally {
    editRollNameLoading.value = false
  }
}

function handleViewDetail(record: ScheduleTeachingRecordItem | Record<string, any>) {
  currentTeachingRecordId.value = String(record.teachingRecordId || '').trim()
  if (!currentTeachingRecordId.value)
    return
  recordDrawerOpen.value = true
}

function handleDelete(record: ScheduleTeachingRecordItem | Record<string, any>) {
  const teachingRecordId = String(record.teachingRecordId || '').trim()
  if (!teachingRecordId || deleting.value)
    return
  deletingTeachingRecordId.value = teachingRecordId
  deleteModalOpen.value = true
}

async function handleConfirmDelete() {
  const teachingRecordId = deletingTeachingRecordId.value
  if (!teachingRecordId || deleting.value)
    return
  deleting.value = true
  try {
    const res = await deleteTeachingRecordApi({ teachingRecordId })
    if (res.code !== 200 || res.result !== true)
      throw new Error(res.message || '删除上课记录失败')
    messageService.success('删除成功')
    deleteModalOpen.value = false
    deletingTeachingRecordId.value = ''
    await loadList()
  }
  catch (error: any) {
    messageService.error(error?.response?.data?.message || error?.message || '删除上课记录失败')
  }
  finally {
    deleting.value = false
  }
}

function handleCancelDelete() {
  if (deleting.value)
    return
  deleteModalOpen.value = false
  deletingTeachingRecordId.value = ''
}

watch(editRollNameOpen, (open) => {
  if (!open)
    editRollNameDetail.value = null
})

function normalizeDayRange(value: unknown) {
  if (!Array.isArray(value) || value.length < 2)
    return null
  const start = dayjs(String(value[0] || ''))
  const end = dayjs(String(value[1] || ''))
  if (!start.isValid() || !end.isValid())
    return null
  return [start.startOf('day'), end.endOf('day')] as [Dayjs, Dayjs]
}

function formatDateTime(record: ScheduleTeachingRecordItem | Record<string, any>) {
  const start = dayjs(record.startTime)
  const end = dayjs(record.endTime)
  if (!start.isValid() || !end.isValid()) {
    return {
      dateText: '-',
      timeText: '--:-- ～ --:--',
    }
  }
  const weeks = ['周日', '周一', '周二', '周三', '周四', '周五', '周六']
  return {
    dateText: `${start.format('YYYY-MM-DD')}(${weeks[start.day()]})`,
    timeText: `${start.format('HH:mm')} ～ ${end.format('HH:mm')}`,
  }
}

function formatNumber(value?: number, suffix = '') {
  const num = Number(value || 0)
  if (!Number.isFinite(num))
    return suffix ? `0${suffix}` : '0'
  const text = Number.isInteger(num) ? String(num) : num.toFixed(2).replace(/\.?0+$/, '')
  return suffix ? `${text}${suffix}` : text
}

function formatCurrency(value?: number) {
  return `￥${Number(value || 0).toFixed(2)}`
}

function formatMinuteDateTime(value?: string) {
  const text = String(value || '').trim()
  if (!text)
    return '-'
  const date = dayjs(text)
  return date.isValid() ? date.format('YYYY-MM-DD HH:mm') : text
}

function attendanceRateText(record: ScheduleTeachingRecordItem | Record<string, any>) {
  const shouldAttend = Number(record.shouldAttendCount || 0)
  if (shouldAttend <= 0)
    return '0%'
  return `${Math.round(Number(record.attendanceRate || 0) * 100)}%`
}

function attendanceRateSummaryText(record: ScheduleTeachingRecordItem | Record<string, any>) {
  return `实到${Number(record.attendCount || 0)}人 / 应到${Number(record.shouldAttendCount || 0)}人`
}

watch(
  () => `${props.open}|${String(props.classId || '').trim()}`,
  () => {
    loadList()
  },
  { immediate: true },
)
</script>

<template>
  <div class="m-12px">
    <all-filter
      :display-array="displayArray"
      @update:class-ending-time-filter="handleClassEndingTimeFilter"
      @update:class-stop-time-filter="handleClassStopTimeFilter"
    />
  </div>
  <div class="m-12px">
    <div class="bg-#fff pt-18px px-20px rounded-10px">
      <div class="flex justify-between items-center">
        <custom-title :title="`共 ${displaySummary.total} 条上课记录 学员总计 ${formatNumber(displaySummary.totalClassTimes, '课时')}，上课教师总计 ${formatNumber(displaySummary.totalTeacherTimes, '课时')} ，共消耗学费 ${formatCurrency(displaySummary.totalTuition)}`" font-size="14px" class="pb-12px" />
      </div>
      <a-table
        row-key="teachingRecordId"
        size="small"
        :loading="loading"
        :columns="columns"
        :data-source="filteredList"
        :pagination="false"
        :scroll="{ x: totalWidth }"
      >
        <template #headerCell="{ column }">
          <template v-if="column.key === 'studentClassTime'">
            学员消耗课时
            <a-popover title="学员消耗课时">
              <template #content>
                <div>本日程全部学员的消耗总课时</div>
              </template>
              <ExclamationCircleOutlined />
            </a-popover>
          </template>
          <template v-if="column.key === 'attendanceRate'">
            出勤率
            <a-popover title="出勤率">
              <template #content>
                <div>实到人数 / 应到人数 = 出勤率</div>
              </template>
              <ExclamationCircleOutlined />
            </a-popover>
          </template>
          <template v-if="column.key === 'consumeFee'">
            消耗学费
            <a-popover title="消耗学费">
              <template #content>
                <div>本次点名数量对应的实际确认收入</div>
              </template>
              <ExclamationCircleOutlined />
            </a-popover>
          </template>
        </template>
        <template #bodyCell="{ column, record }">
          <template v-if="column.dataIndex === 'date'">
            <div>{{ formatDateTime(record).dateText }}</div>
            <div>{{ formatDateTime(record).timeText }}</div>
          </template>
          <template v-if="column.dataIndex === 'classTime'">
            {{ formatNumber(record.teacherClassTime, '课时') }}
          </template>
          <template v-if="column.dataIndex === 'studentClassTime'">
            {{ formatNumber(record.actualQuantity, '课时') }}
          </template>
          <template v-if="column.dataIndex === 'consumeFee'">
            {{ formatCurrency(record.actualTuition) }}
          </template>
          <template v-if="column.dataIndex === 'teacher'">
            {{ record.teacherName || '-' }}
          </template>
          <template v-if="column.dataIndex === 'assistant'">
            {{ record.assistants || '-' }}
          </template>
          <template v-if="column.dataIndex === 'attendanceRate'">
            <div class="leading-20px">
              <div>{{ attendanceRateText(record) }}</div>
              <div class="text-3 text-#888 whitespace-nowrap">
                {{ attendanceRateSummaryText(record) }}
              </div>
            </div>
          </template>
          <template v-if="column.dataIndex === 'attendance'">
            {{ Number(record.attendCount || 0) }}
          </template>
          <template v-if="column.dataIndex === 'leave'">
            {{ Number(record.leaveCount || 0) }}
          </template>
          <template v-if="column.dataIndex === 'absent'">
            {{ Number(record.absentCount || 0) }}
          </template>
          <template v-if="column.dataIndex === 'unrecorded'">
            {{ Number(record.unrecordedCount || 0) }}
          </template>
          <template v-if="column.dataIndex === 'rollCallTime'">
            {{ formatMinuteDateTime(record.updatedTime || record.createdTime) }}
          </template>
          <template v-if="column.dataIndex === 'action'">
            <a-space :size="12">
              <a @click="handleEditRollCall(record)">编辑点名</a>
              <a @click="handleViewDetail(record)">详情</a>
              <a @click="handleDelete(record)">删除</a>
            </a-space>
          </template>
        </template>
      </a-table>
    </div>
    <ClassRecordDetails
      v-model:open="recordDrawerOpen"
      :teaching-record-id="currentTeachingRecordId"
      @updated="loadList"
      @deleted="loadList"
    />
    <EditRollNameModal
      v-model:open="editRollNameOpen"
      :detail="editRollNameDetail"
    />
    <a-modal
      v-model:open="deleteModalOpen"
      centered
      :footer="false"
      :closable="false"
      :mask-closable="false"
      :keyboard="false"
      width="440px"
    >
      <div class="text-18px mb-12px font500">
        <CloseCircleOutlined class="text-#f00 mr2 text-5" /> 删除上课点名记录？
      </div>
      <div class="pl-30px text-#666">
        <div>1.删除后已点名扣费学员将会返还学费，并减少对应的已确认收入;</div>
        <div>2.若包含试听学员，已试听状态将变成已取消状态，并删除上课记录；若包含补课学员，已补课状态将变成已安排或未安排状态，并删除上课记录;</div>
        <div>3.删除上课点名记录后，所对应的日程中的学员点名状态变成未点名;</div>
        <div>4.删除上课点名记录后，日程状态从已点名变成未点名。</div>
        <div class="text-#f00 mt-12px">
          <ExclamationCircleFilled /> 此操作不可撤销，请谨慎操作
        </div>
      </div>
      <a-space class="mt-24px flex justify-end">
        <a-button danger ghost :loading="deleting" @click="handleConfirmDelete">
          删除
        </a-button>
        <a-button
          class="text-#666"
          :disabled="deleting"
          @click="handleCancelDelete"
        >
          再想想
        </a-button>
      </a-space>
    </a-modal>
  </div>
</template>
