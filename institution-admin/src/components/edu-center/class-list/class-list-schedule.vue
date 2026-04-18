<script setup lang="ts">
import { ExclamationCircleOutlined } from '@ant-design/icons-vue'
import type { TableColumnsType } from 'ant-design-vue'
import { Modal } from 'ant-design-vue'
import dayjs from 'dayjs'
import { computed, ref, watch } from 'vue'
import scheduleClassRepeatImage from '@/assets/images/timetable/schedule-class-repeat-local.png'
import scheduleClassSingleImage from '@/assets/images/timetable/schedule-class-single.png'
import { getGroupClassDrawerSchedulesApi, type GroupClassDrawerScheduleItem } from '@/api/edu-center/group-class'
import {
  cancelTeachingScheduleScopedApi,
  getTeachingScheduleBatchDetailApi,
  type TeachingScheduleBatchDetail,
  type TeachingScheduleItem,
} from '@/api/edu-center/teaching-schedule'
import {
  inferGroupClassBatchPlanPreset,
  type GroupClassBatchPlanModalPreset,
} from '@/components/edu-center/timetable/group-class-batch-plan-preset'
import GroupClassScheduleModal from '@/components/edu-center/timetable/group-class-schedule-modal.vue'
import {
  loadTeachingScheduleDeleteTargetCount,
  sortTeachingScheduleItemsByTimeline,
} from '@/components/edu-center/timetable/schedule-delete-scope'
import SmartTimetableScheduleDetailDrawer from '@/components/edu-center/timetable/smart-timetable-schedule-detail-drawer.vue'
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

type ScheduleActionMode = 'view' | 'edit' | 'copy' | 'delete'
type ScheduleScope = 'current' | 'future'

interface ResolvedScheduleActionContext {
  detail: TeachingScheduleBatchDetail
  anchor: TeachingScheduleItem
  scope: ScheduleScope
}

const columns: TableColumnsType<GroupClassDrawerScheduleItem> = [
  {
    title: '重复规则',
    dataIndex: 'repeatRule',
    key: 'repeatRule',
    width: 260,
  },
  {
    title: '上课时间',
    dataIndex: 'timeText',
    key: 'timeText',
    width: 160,
  },
  {
    title: '已上/排课',
    dataIndex: 'status',
    key: 'status',
    width: 110,
  },
  {
    title: '上课教师',
    dataIndex: 'teacherName',
    key: 'teacherName',
    width: 110,
  },
  {
    title: '上课助教',
    dataIndex: 'assistantText',
    key: 'assistantText',
    width: 120,
  },
  {
    title: '上课教室',
    dataIndex: 'classroomName',
    key: 'classroomName',
    width: 120,
  },
  {
    title: '操作',
    dataIndex: 'action',
    key: 'action',
    width: 220,
    fixed: 'right',
  },
]

const loading = ref(false)
const dataSource = ref<GroupClassDrawerScheduleItem[]>([])
const detailOpen = ref(false)
const detailState = ref<Record<string, any> | null>(null)
const currentDetailRecord = ref<GroupClassDrawerScheduleItem | null>(null)
const scheduleModalOpen = ref(false)
const scheduleModalMode = ref<'create' | 'editBatch'>('create')
const scheduleBatchPlanPreset = ref<GroupClassBatchPlanModalPreset | null>(null)
const resolvingAction = ref(false)
const deleting = ref(false)

const totalWidth = computed(() =>
  columns.reduce((sum, item) => sum + Number(item.width || 0), 0),
)

function getRemainingScheduleCount(record?: Partial<GroupClassDrawerScheduleItem> | null) {
  return Math.max(0, Number(record?.scheduleCount || 0) - Number(record?.completedCount || 0))
}

function canOperateRecord(record?: Partial<GroupClassDrawerScheduleItem> | null) {
  return getRemainingScheduleCount(record) > 0
}

function getEditActionText(record?: Partial<GroupClassDrawerScheduleItem> | null) {
  return '编辑'
}

function getViewActionTooltip(record?: Partial<GroupClassDrawerScheduleItem> | null) {
  return hasBatchScheduleRecord(record) ? '查看最近一节日程详情' : '查看日程详情'
}

function getEditActionDisabledReason(record?: Partial<GroupClassDrawerScheduleItem> | null) {
  return hasBatchScheduleRecord(record) ? '当前批次已无可编辑的后续日程' : '当前日程不可编辑'
}

function getEditActionTooltip(record?: Partial<GroupClassDrawerScheduleItem> | null) {
  return canOperateRecord(record)
    ? (resolveScheduleScope(record) === 'future' ? '编辑以后日程' : '编辑日程')
    : getEditActionDisabledReason(record)
}

function getDeleteActionDisabledReason(record?: Partial<GroupClassDrawerScheduleItem> | null) {
  return hasBatchScheduleRecord(record) ? '当前批次已无可删除的后续日程' : '当前日程不可删除'
}

function compactWeekdayLabel(value?: string) {
  const text = String(value || '').trim()
  return text.startsWith('周') ? text.slice(1) : text
}

function formatRepeatWeekdayText(prefix: string, weekdays?: string[]) {
  const labels = (Array.isArray(weekdays) ? weekdays : [])
    .map(item => compactWeekdayLabel(item))
    .filter(Boolean)
  return labels.length ? `${prefix}${labels.join('、')}` : prefix
}

function getTimeRuleText(record?: Partial<GroupClassDrawerScheduleItem> | null) {
  const repeatRule = String(record?.repeatRule || '').trim()
  const metaRepeatRule = String(record?.batchMeta?.repeatRule || '').trim().toLowerCase()
  const schedulingMode = String(record?.batchMeta?.schedulingMode || '').trim().toLowerCase()
  const selectedWeekdays = Array.isArray(record?.batchMeta?.selectedWeekdays)
    ? record.batchMeta.selectedWeekdays
    : []

  if (metaRepeatRule === 'daily' || repeatRule === '每天重复')
    return '每天'
  if (metaRepeatRule === 'alternateday' || metaRepeatRule === 'alternate_day' || repeatRule === '隔天重复')
    return '隔天'
  if (metaRepeatRule === 'weekly' || repeatRule === '每周重复')
    return formatRepeatWeekdayText('每周', selectedWeekdays)
  if (metaRepeatRule === 'biweekly' || repeatRule === '隔周重复')
    return formatRepeatWeekdayText('隔周', selectedWeekdays)
  if (schedulingMode === 'free' && hasBatchScheduleRecord(record))
    return '自由排课'
  return String(record?.weekdayText || '-').trim() || '-'
}

function hasBatchScheduleRecord(record?: Partial<GroupClassDrawerScheduleItem> | null) {
  return Number(record?.scheduleCount || 0) > 1 || String(record?.batchNo || '').trim() !== ''
}

function resolveScheduleScope(record?: Partial<GroupClassDrawerScheduleItem> | null): ScheduleScope {
  return hasBatchScheduleRecord(record) && getRemainingScheduleCount(record) > 1 ? 'future' : 'current'
}

function isRolledCallSchedule(item?: Partial<TeachingScheduleItem> | null) {
  return Number(item?.callStatus || 1) === 2
}

function isPastSchedule(item?: Partial<TeachingScheduleItem> | null) {
  const lessonDate = String(item?.lessonDate || '').trim()
  if (!lessonDate)
    return false
  return dayjs(lessonDate).isBefore(dayjs().startOf('day'), 'day')
}

function resolveViewAnchor(list: TeachingScheduleItem[]) {
  const sorted = sortTeachingScheduleItemsByTimeline(list)
  const upcomingPending = sorted.find(item => !isRolledCallSchedule(item) && !isPastSchedule(item))
  if (upcomingPending)
    return upcomingPending

  const latestPending = [...sorted].reverse().find(item => !isRolledCallSchedule(item))
  if (latestPending)
    return latestPending

  return sorted[sorted.length - 1] || null
}

function resolveActionAnchor(list: TeachingScheduleItem[], scope: ScheduleScope, mode: ScheduleActionMode) {
  const sorted = sortTeachingScheduleItemsByTimeline(list)
  const pending = sorted.filter(item => !isRolledCallSchedule(item))
  if (!pending.length)
    return null
  if (scope === 'current')
    return pending[0]

  const todayOrFuture = pending.find(item => !isPastSchedule(item))
  if (todayOrFuture)
    return todayOrFuture
  if (mode === 'delete' || mode === 'view')
    return pending[0]
  return null
}

async function loadScheduleActionContext(record: GroupClassDrawerScheduleItem, mode: ScheduleActionMode): Promise<ResolvedScheduleActionContext> {
  const scheduleId = String(record?.detailScheduleId || '').trim()
  const batchNo = String(record?.batchNo || '').trim()
  if (!scheduleId && !batchNo)
    throw new Error('当前日程缺少标识，请刷新后重试')

  const res = await getTeachingScheduleBatchDetailApi({
    batchNo: batchNo || undefined,
    id: batchNo ? undefined : (scheduleId || undefined),
  })
  if (res.code !== 200 || !res.result)
    throw new Error(res.message || '加载日程详情失败')

  const detail = res.result
  const scope = resolveScheduleScope(record)
  const anchor = mode === 'view'
    ? resolveViewAnchor(detail.schedules || [])
    : resolveActionAnchor(detail.schedules || [], scope, mode)
  if (!anchor) {
    if (mode === 'edit')
      throw new Error(scope === 'future' ? '当前批次暂无可编辑的后续日程' : '过去日程不可编辑')
    if (mode === 'copy')
      throw new Error(scope === 'future' ? '当前批次暂无可复制的后续日程' : '当前日程不可复制')
    if (mode === 'delete')
      throw new Error(scope === 'future' ? '当前批次暂无可删除的后续日程' : '当前日程不可删除')
    throw new Error('当前日程暂无可查看内容')
  }

  return {
    detail,
    anchor,
    scope,
  }
}

function openScheduleCreateModal(preset: GroupClassBatchPlanModalPreset) {
  scheduleModalMode.value = 'create'
  scheduleBatchPlanPreset.value = preset
  scheduleModalOpen.value = true
}

function openScheduleEditModal(preset: GroupClassBatchPlanModalPreset) {
  scheduleModalMode.value = 'editBatch'
  scheduleBatchPlanPreset.value = preset
  scheduleModalOpen.value = true
}

async function loadList() {
  const classId = String(props.classId || '').trim()
  if (!props.open || !classId) {
    dataSource.value = []
    return
  }
  loading.value = true
  try {
    const res = await getGroupClassDrawerSchedulesApi({ classId })
    if (res.code !== 200)
      throw new Error(res.message || '加载班级日程失败')
    dataSource.value = Array.isArray(res.result?.list) ? res.result.list : []
  }
  catch (error: any) {
    dataSource.value = []
    messageService.error(error?.response?.data?.message || error?.message || '加载班级日程失败')
  }
  finally {
    loading.value = false
  }
}

async function handleViewDetail(record: GroupClassDrawerScheduleItem | Record<string, any>) {
  const target = record as GroupClassDrawerScheduleItem
  currentDetailRecord.value = target
  try {
    const { detail, anchor } = await loadScheduleActionContext(target, 'view')
    detailState.value = {
      scheduleId: anchor.id,
      id: anchor.id,
      batchNo: detail.batchNo || target.batchNo,
      batchSize: detail.batchSize || target.scheduleCount,
      lessonTitle: target.lessonName || props.className || '日程详情',
      assistantText: target.assistantText,
      classroomName: target.classroomName,
      teacherName: target.teacherName,
      batchMeta: detail.batchMeta || target.batchMeta,
    }
    detailOpen.value = true
  }
  catch (error: any) {
    messageService.error(error?.response?.data?.message || error?.message || '加载日程详情失败')
  }
}

function handleQuickSchedule() {
  if (!String(props.classId || '').trim()) {
    messageService.warning('当前班级信息不完整，暂不可排课')
    return
  }
  scheduleModalMode.value = 'create'
  scheduleBatchPlanPreset.value = null
  scheduleModalOpen.value = true
}

async function openSchedulePreset(record: GroupClassDrawerScheduleItem | Record<string, any>, mode: 'edit' | 'copy', scopeOverride?: ScheduleScope) {
  if (resolvingAction.value)
    return
  resolvingAction.value = true
  try {
    const target = record as GroupClassDrawerScheduleItem
    const { detail, anchor, scope } = await loadScheduleActionContext(target, mode)
    const preset = inferGroupClassBatchPlanPreset(detail, anchor.id)
    preset.editScope = (scopeOverride || scope) === 'future' ? 'batch' : 'current'
    if (mode === 'edit')
      openScheduleEditModal(preset)
    else
      openScheduleCreateModal(preset)
  }
  catch (error: any) {
    messageService.warning(error?.response?.data?.message || error?.message || (mode === 'edit' ? '打开编辑失败' : '打开复制失败'))
  }
  finally {
    resolvingAction.value = false
  }
}

async function handleEdit(record: GroupClassDrawerScheduleItem | Record<string, any> | null | undefined, scopeOverride?: ScheduleScope) {
  const target = (record as GroupClassDrawerScheduleItem | null | undefined) || currentDetailRecord.value
  if (!target)
    return
  await openSchedulePreset(target, 'edit', scopeOverride)
}

async function handleCopy(record: GroupClassDrawerScheduleItem | Record<string, any> | null | undefined, scopeOverride?: ScheduleScope) {
  const target = (record as GroupClassDrawerScheduleItem | null | undefined) || currentDetailRecord.value
  if (!target)
    return
  await openSchedulePreset(target, 'copy', scopeOverride)
}

async function handleDelete(record: GroupClassDrawerScheduleItem | Record<string, any> | null | undefined, scopeOverride?: ScheduleScope) {
  const target = (record as GroupClassDrawerScheduleItem | null | undefined) || currentDetailRecord.value
  if (!target || resolvingAction.value || deleting.value)
    return

  resolvingAction.value = true
  try {
    const { anchor, scope } = await loadScheduleActionContext(target, 'delete')
    const finalScope = scopeOverride || scope
    let deleteCount = 1
    if (finalScope === 'future') {
      deleteCount = await loadTeachingScheduleDeleteTargetCount({
        id: anchor.id,
        batchNo: anchor.batchNo,
      }, 'future')
    }

    Modal.confirm({
      title: finalScope === 'future' ? '删除后续全部日程?' : '删除日程?',
      content: finalScope === 'future'
        ? `后续 ${deleteCount} 个日程将被全部删除，删除后不可恢复，请谨慎操作`
        : '删除后将不可恢复，请谨慎操作',
      okText: '删除',
      cancelText: '取消',
      async onOk() {
        deleting.value = true
        try {
          const res = await cancelTeachingScheduleScopedApi({
            id: anchor.id,
            scope: finalScope,
          })
          if (res.code !== 200)
            throw new Error(res.message || '删除日程失败')
          detailOpen.value = false
          messageService.success(finalScope === 'future' ? `已删除后续 ${deleteCount} 节班课日程` : '已删除班课日程')
          await loadList()
        }
        catch (error: any) {
          messageService.error(error?.response?.data?.message || error?.message || '删除日程失败')
          throw error
        }
        finally {
          deleting.value = false
        }
      },
    })
  }
  catch (error: any) {
    messageService.warning(error?.response?.data?.message || error?.message || '删除日程失败')
  }
  finally {
    resolvingAction.value = false
  }
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
    <div class="bg-#fff pt-18px px-20px rounded-10px">
      <div class="flex justify-between items-center">
        <custom-title :title="`共 ${dataSource.length} 个日程`" font-size="14px" class="pb-12px" />
        <a-button type="primary" class="mb-12px" @click="handleQuickSchedule">
          一键排课
        </a-button>
      </div>
      <a-table
        row-key="key"
        size="small"
        :loading="loading"
        :columns="columns"
        :data-source="dataSource"
        :pagination="false"
        :scroll="{ x: totalWidth }"
      >
        <template #headerCell="{ column }">
          <template v-if="column.key === 'status'">
            已上/排课
            <a-popover title="已上/排课">
              <template #content>
                <div>已完成日程数/排课日程总数</div>
              </template>
              <ExclamationCircleOutlined />
            </a-popover>
          </template>
        </template>
        <template #bodyCell="{ column, record }">
          <template v-if="column.dataIndex === 'repeatRule'">
            <div class="flex flex-items-center">
              <img
                class="w-34px h-34px"
                :src="Number(record.type || 0) === 1 ? scheduleClassRepeatImage : scheduleClassSingleImage"
                alt=""
              >
              <div class="ml-12px text-#666 leading-20px">
                <div class="text-14px text-#222">
                  {{ record.repeatRule || '-' }}
                </div>
                <div class="text-13px">
                  {{ record.dateRangeText || '-' }}
                </div>
              </div>
            </div>
          </template>
          <template v-if="column.dataIndex === 'timeText'">
            <div>{{ record.timeText || '-' }}</div>
            <div class="text-#888">
              {{ getTimeRuleText(record) }}
            </div>
          </template>
          <template v-if="column.dataIndex === 'status'">
            {{ `${Number(record.completedCount || 0)}/${Number(record.scheduleCount || 0)}节` }}
          </template>
          <template v-if="column.dataIndex === 'teacherName'">
            {{ record.teacherName || '-' }}
          </template>
          <template v-if="column.dataIndex === 'assistantText'">
            {{ record.assistantText || '-' }}
          </template>
          <template v-if="column.dataIndex === 'classroomName'">
            {{ record.classroomName || '-' }}
          </template>
          <template v-if="column.dataIndex === 'action'">
            <a-space :size="12">
              <a-tooltip :title="getViewActionTooltip(record)">
                <a @click="handleViewDetail(record)">详情</a>
              </a-tooltip>
              <a-tooltip :title="getEditActionTooltip(record)">
                <a v-if="canOperateRecord(record)" @click="handleEdit(record)">{{ getEditActionText(record) }}</a>
                <span v-else class="class-list-schedule__action class-list-schedule__action--disabled">{{ getEditActionText(record) }}</span>
              </a-tooltip>
              <a-tooltip :title="canOperateRecord(record) ? '删除日程' : getDeleteActionDisabledReason(record)">
                <a v-if="canOperateRecord(record)" @click="handleDelete(record)">删除</a>
                <span v-else class="class-list-schedule__action class-list-schedule__action--disabled">删除</span>
              </a-tooltip>
            </a-space>
          </template>
        </template>
      </a-table>
    </div>
    <SmartTimetableScheduleDetailDrawer
      v-model:open="detailOpen"
      :detail="detailState"
      :editable="Boolean(currentDetailRecord)"
      :deletable="Boolean(currentDetailRecord)"
      :deleting="deleting"
      @delete="handleDelete(undefined, 'current')"
      @delete-current="handleDelete(undefined, 'current')"
      @delete-future="handleDelete(undefined, 'future')"
      @copy="handleCopy(undefined, 'future')"
      @copy-current="handleCopy(undefined, 'current')"
      @edit="handleEdit(undefined, 'future')"
      @edit-current="handleEdit(undefined, 'current')"
      @updated="loadList"
    />
    <GroupClassScheduleModal
      v-model:open="scheduleModalOpen"
      :mode="scheduleModalMode"
      :batch-plan-preset="scheduleBatchPlanPreset"
      :initial-group-class-id="scheduleModalMode === 'create' ? String(props.classId || '') : ''"
      @updated="loadList"
    />
  </div>
</template>

<style scoped>
.class-list-schedule__action {
  color: #1677ff;
  cursor: pointer;
  transition: color 0.2s ease;
}

.class-list-schedule__action--disabled {
  color: #bfbfbf;
  cursor: not-allowed;
}
</style>
