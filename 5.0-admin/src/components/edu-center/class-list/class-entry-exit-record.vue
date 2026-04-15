<script setup lang="ts">
import { CloseOutlined, InfoCircleFilled } from '@ant-design/icons-vue'
import dayjs, { type Dayjs } from 'dayjs'
import type { TableColumnsType } from 'ant-design-vue'
import { computed, reactive, ref, watch } from 'vue'
import { ParentRelationshipLabel } from '@/enums'
import {
  pageGroupClassEntryExitRecordsApi,
  updateGroupClassEntryExitRecordTimeApi,
  type GroupClassEntryExitRecordItem,
} from '@/api/edu-center/group-class'
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

const displayArray = ['stuPhoneSearch', 'classStatus', 'classEndingTime']
const defaultAvatarUrl = 'https://pcsys.admin.ybc365.com/a369a751-2be5-4929-974d-9ae4439f54c4.png'
const entryExitStatusOptions = [
  { id: '1', value: '入班' },
  { id: '2', value: '出班' },
]

const columns: TableColumnsType<GroupClassEntryExitRecordItem> = [
  {
    title: '学员信息',
    dataIndex: 'studentName',
    key: 'studentName',
    width: 220,
  },
  {
    title: '联系电话',
    dataIndex: 'phone',
    key: 'phone',
    width: 160,
  },
  {
    title: '出入班状态',
    dataIndex: 'entryExitStatusText',
    key: 'entryExitStatusText',
    width: 120,
  },
  {
    title: '出入班时间',
    dataIndex: 'entryExitTime',
    key: 'entryExitTime',
    width: 180,
  },
  {
    title: '操作人',
    dataIndex: 'operatorName',
    key: 'operatorName',
    width: 120,
  },
  {
    title: '操作时间',
    dataIndex: 'operateTime',
    key: 'operateTime',
    width: 180,
  },
  {
    title: '操作',
    dataIndex: 'action',
    key: 'action',
    width: 100,
    fixed: 'right',
  },
]

const loading = ref(false)
const saving = ref(false)
const allFilterKey = ref(0)
const dataSource = ref<GroupClassEntryExitRecordItem[]>([])
const filterStudentId = ref<string>()
const filterEntryExitStatuses = ref<number[]>([])
const filterRecordRange = ref<[string, string] | null>(null)
const studentCount = ref(0)
const pagination = reactive({
  current: 1,
  pageSize: 20,
  total: 0,
})

const editModalOpen = ref(false)
const editingRecord = ref<GroupClassEntryExitRecordItem | null>(null)
const editDate = ref<Dayjs | null>(null)

let loadSeq = 0

const totalWidth = computed(() =>
  columns.reduce((sum, item) => sum + Number(item.width || 0), 0),
)

const tablePagination = computed(() => ({
  current: pagination.current,
  pageSize: pagination.pageSize,
  total: pagination.total,
  showSizeChanger: true,
  showQuickJumper: true,
  showTotal: (total: number) => `共 ${total} 条`,
}))

const editDateBounds = computed(() => ({
  previous: editingRecord.value?.previousRecordTime ? dayjs(editingRecord.value.previousRecordTime) : null,
  next: editingRecord.value?.nextRecordTime ? dayjs(editingRecord.value.nextRecordTime) : null,
}))

async function loadList() {
  const classId = String(props.classId || '').trim()
  if (!props.open || !classId) {
    dataSource.value = []
    pagination.total = 0
    studentCount.value = 0
    return
  }
  const currentSeq = ++loadSeq
  loading.value = true
  try {
    const res = await pageGroupClassEntryExitRecordsApi({
      queryModel: {
        classId,
        studentId: filterStudentId.value,
        recordStartDate: filterRecordRange.value?.[0],
        recordEndDate: filterRecordRange.value?.[1],
        entryExitStatuses: filterEntryExitStatuses.value,
      },
      pageRequestModel: {
        needTotal: true,
        pageIndex: pagination.current,
        pageSize: pagination.pageSize,
        skipCount: (pagination.current - 1) * pagination.pageSize,
      },
    })
    if (currentSeq !== loadSeq)
      return
    if (res.code !== 200)
      throw new Error(res.message || '加载出入班记录失败')
    dataSource.value = Array.isArray(res.result?.list) ? res.result.list : []
    pagination.total = Number(res.result?.total || 0)
    studentCount.value = Number(res.result?.studentCount || 0)
  }
  catch (error: any) {
    if (currentSeq !== loadSeq)
      return
    dataSource.value = []
    pagination.total = 0
    studentCount.value = 0
    messageService.error(error?.response?.data?.message || error?.message || '加载出入班记录失败')
  }
  finally {
    if (currentSeq === loadSeq)
      loading.value = false
  }
}

function resetFilters() {
  filterStudentId.value = undefined
  filterEntryExitStatuses.value = []
  filterRecordRange.value = null
  pagination.current = 1
}

function reloadFirstPage() {
  pagination.current = 1
  loadList()
}

function normalizeStringValue(value: unknown) {
  const normalized = String(value || '').trim()
  return normalized || undefined
}

function normalizeDateRange(value: unknown) {
  if (!Array.isArray(value) || value.length < 2)
    return null
  const start = String(value[0] || '').trim()
  const end = String(value[1] || '').trim()
  if (!start || !end)
    return null
  return [start, end] as [string, string]
}

function normalizeNumberArray(value: unknown) {
  if (!Array.isArray(value))
    return []
  return value
    .map(item => Number(item))
    .filter(item => Number.isFinite(item) && item > 0)
}

function handleStudentFilter(value: unknown) {
  const next = normalizeStringValue(value)
  if (filterStudentId.value === next)
    return
  filterStudentId.value = next
  reloadFirstPage()
}

function handleEntryExitStatusFilter(value: unknown) {
  const next = normalizeNumberArray(value)
  if (JSON.stringify(filterEntryExitStatuses.value) === JSON.stringify(next))
    return
  filterEntryExitStatuses.value = next
  reloadFirstPage()
}

function handleRecordDateFilter(value: unknown) {
  const next = normalizeDateRange(value)
  if (JSON.stringify(filterRecordRange.value) === JSON.stringify(next))
    return
  filterRecordRange.value = next
  reloadFirstPage()
}

function handleTableChange(page: { current?: number, pageSize?: number }) {
  const nextCurrent = Number(page.current || 1)
  const nextSize = Number(page.pageSize || pagination.pageSize)
  const changed = nextCurrent !== pagination.current || nextSize !== pagination.pageSize
  pagination.current = nextCurrent
  pagination.pageSize = nextSize
  if (changed)
    loadList()
}

function formatDateTime(value?: string) {
  if (!value)
    return '-'
  const date = dayjs(value)
  return date.isValid() ? date.format('YYYY-MM-DD HH:mm') : '-'
}

function formatDate(value?: string) {
  if (!value)
    return '-'
  const date = dayjs(value)
  return date.isValid() ? date.format('YYYY-MM-DD') : '-'
}

function getRelationText(value?: number) {
  if (!value)
    return '未知'
  return ParentRelationshipLabel[value as keyof typeof ParentRelationshipLabel] || ''
}

function getStudentAvatar(value?: string) {
  const avatar = String(value || '').trim()
  return avatar || defaultAvatarUrl
}

function openEditModal(record: GroupClassEntryExitRecordItem | Record<string, any>) {
  editingRecord.value = record as GroupClassEntryExitRecordItem
  editDate.value = dayjs(String(record.entryExitTime || ''))
  editModalOpen.value = true
}

function resetEditModalState() {
  editModalOpen.value = false
  editingRecord.value = null
  editDate.value = null
}

function closeEditModal() {
  if (saving.value)
    return
  resetEditModalState()
}

function disabledEditDate(current: Dayjs) {
  if (!current)
    return false
  const currentDay = current.startOf('day')
  if (currentDay.isAfter(dayjs().endOf('day')))
    return true
  const previous = editDateBounds.value.previous
  if (previous && !currentDay.isAfter(previous.startOf('day')))
    return true
  const next = editDateBounds.value.next
  if (next && !currentDay.isBefore(next.startOf('day')))
    return true
  return false
}

async function submitEdit() {
  const record = editingRecord.value
  if (!record || !editDate.value) {
    messageService.warning('请选择出入班日期')
    return
  }
  saving.value = true
  try {
    const res = await updateGroupClassEntryExitRecordTimeApi({
      id: record.id,
      entryExitTime: editDate.value.format('YYYY-MM-DD'),
    })
    if (res.code !== 200)
      throw new Error(res.message || '调整出入班日期失败')
    messageService.success('调整成功')
    resetEditModalState()
    await loadList()
  }
  catch (error: any) {
    messageService.error(error?.response?.data?.message || error?.message || '调整出入班日期失败')
  }
  finally {
    saving.value = false
  }
}

watch(
  () => `${props.open}|${String(props.classId || '').trim()}`,
  () => {
    resetFilters()
    closeEditModal()
    allFilterKey.value += 1
    loadList()
  },
  { immediate: true },
)
</script>

<template>
  <div class="group-class-history-tab">
    <div class="filter-wrap">
      <all-filter
        :key="allFilterKey"
        :display-array="displayArray"
        :class-status-options="entryExitStatusOptions"
        class-status-label="出入班状态"
        class-ending-time-label="出入班日期"
        :is-quick-show="false"
        @update:stu-phone-search-filter="handleStudentFilter"
        @update:class-status-filter="handleEntryExitStatusFilter"
        @update:class-ending-time-filter="handleRecordDateFilter"
      />
    </div>

    <div class="history-card">
      <div class="history-card__summary">
        <custom-title :title="`共 ${pagination.total} 条记录，涉及 ${studentCount} 位学员`" font-size="14px" class="pb-12px" />
      </div>

      <a-table
        row-key="id"
        size="middle"
        :loading="loading"
        :columns="columns"
        :data-source="dataSource"
        :pagination="tablePagination"
        :scroll="{ x: totalWidth }"
        @change="handleTableChange"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.dataIndex === 'studentName'">
            <div class="student-cell">
              <img :src="getStudentAvatar(record.avatar)" alt="" class="student-cell__avatar">
              <div class="student-cell__meta">
                <div class="student-cell__name">
                  {{ record.studentName || '-' }}
                </div>
                <div class="student-cell__sub">
                  未知
                </div>
              </div>
            </div>
          </template>
          <template v-else-if="column.dataIndex === 'phone'">
            <div class="phone-cell">
              <div class="sub-text">
                {{ getRelationText(record.phoneRelationship) }}
              </div>
              <div class="phone-text">
                {{ record.phone || '-' }}
              </div>
            </div>
          </template>
          <template v-else-if="column.dataIndex === 'entryExitStatusText'">
            <a-tag :color="record.entryExitStatus === 1 ? 'blue' : 'orange'" class="type-tag">
              {{ record.entryExitStatusText || '-' }}
            </a-tag>
          </template>
          <template v-else-if="column.dataIndex === 'entryExitTime'">
            {{ formatDate(record.entryExitTime) }}
          </template>
          <template v-else-if="column.dataIndex === 'operateTime'">
            {{ formatDateTime(record.operateTime) }}
          </template>
          <template v-else-if="column.dataIndex === 'operatorName'">
            {{ record.operatorName || '-' }}
          </template>
          <template v-else-if="column.dataIndex === 'action'">
            <a @click="openEditModal(record)">编辑</a>
          </template>
        </template>
      </a-table>
    </div>

    <a-modal
      v-model:open="editModalOpen"
      centered
      class="modal-content-box"
      :keyboard="false"
      :closable="false"
      :mask-closable="false"
      :confirm-loading="saving"
      :width="600"
      destroy-on-close
      @ok="submitEdit"
      @cancel="closeEditModal"
    >
      <template #title>
        <div class="text-5 flex justify-between flex-center">
          <span>编辑出入班时间</span>
          <a-button type="text" class="close-btn" @click="closeEditModal">
            <template #icon>
              <CloseOutlined class="text-5 close-icon" />
            </template>
          </a-button>
        </div>
      </template>
      <a-alert class="edit-modal__alert" type="info" show-icon>
        <template #icon>
          <InfoCircleFilled />
        </template>
        <template #message>
          如果学员有多条记录，日期调整范围只能在本次记录的上一条记录与下一条记录的日期范围之间进行调整
        </template>
      </a-alert>
      <div class="edit-modal__content">
        <a-date-picker
          v-model:value="editDate"
          style="width: 100%;"
          format="YYYY-MM-DD"
          placeholder="请选择出入班日期"
          :disabled-date="disabledEditDate"
        />
      </div>
      <template #footer>
        <a-button @click="closeEditModal">
          取消
        </a-button>
        <a-button type="primary" :loading="saving" @click="submitEdit">
          确定
        </a-button>
      </template>
    </a-modal>
  </div>
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

.group-class-history-tab {
  padding: 12px;
}

.filter-wrap {
  margin-bottom: 12px;
}

.history-card {
  border: 1px solid #eef1f6;
  border-radius: 12px;
  background: #fff;
  box-shadow: 0 8px 24px rgba(31, 35, 41, 0.06);
  overflow: hidden;
}

.history-card__summary {
  padding: 18px 20px 0;
}

.student-cell {
  display: flex;
  align-items: center;
  gap: 12px;
}

.student-cell__avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  object-fit: cover;
  flex-shrink: 0;
}

.student-cell__avatar--fallback {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: #e8f3ff;
  color: #1677ff;
  font-weight: 600;
}

.student-cell__meta {
  min-width: 0;
}

.student-cell__name {
  color: #1f2329;
  font-weight: 500;
  line-height: 22px;
}

.student-cell__sub,
.sub-text {
  color: #86909c;
  font-size: 12px;
}

.phone-text {
  color: #1f2329;
  line-height: 22px;
}

.phone-cell {
  display: inline-flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 0;
  line-height: 1;
}

.phone-cell .sub-text {
  margin: 0 0 1px;
  line-height: 14px;
  font-size: 13px;
  color: #4e4e4e;
}

.phone-cell .phone-text {
  line-height: 16px;
}

.type-tag {
  margin-inline-end: 0;
  border-radius: 999px;
}

.edit-modal__content {
  padding: 24px;
}

.edit-modal__alert {
  border: none;
  border-radius: 0;
  background: #e6f0ff;
  color: #1677ff;
}

:deep(.edit-modal__alert .ant-alert-message) {
  color: #1677ff;
  font-size: 14px;
  line-height: 22px;
}

:deep(.edit-modal__alert .ant-alert-icon) {
  color: #1677ff;
}

:deep(.ant-table-wrapper) {
  padding: 0 20px 20px;
}

:deep(.ant-table-thead > tr > th) {
  background: #f8fafc;
  color: #4e5969;
  font-weight: 600;
}
</style>
