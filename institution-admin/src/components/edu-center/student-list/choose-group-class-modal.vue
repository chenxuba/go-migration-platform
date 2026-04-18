<script setup lang="ts">
import { CloseOutlined, QuestionCircleOutlined, SearchOutlined } from '@ant-design/icons-vue'
import { computed, ref, watch } from 'vue'
import type { GroupClassDrawerScheduleItem, GroupClassRow } from '@/api/edu-center/group-class'
import {
  batchAssignGroupClassStudentsApi,
  getGroupClassDrawerSchedulesApi,
  pageGroupClassesApi,
} from '@/api/edu-center/group-class'
import emitter, { EVENTS } from '@/utils/eventBus'
import messageService from '@/utils/messageService'
import CreateClassModal from '@/components/common/create-class-modal.vue'

interface PendingClassRow {
  tuitionAccountId?: string
  studentId?: string
  studentName?: string
  lessonId?: string
  lessonName?: string
}

interface ClassScheduleSummaryItem {
  summary: string
  dateRange: string
}

interface ClassScheduleSummaryDisplay {
  primaryText: string
  extraText: string
  items: ClassScheduleSummaryItem[]
}

const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
  record: {
    type: Object as () => PendingClassRow | null,
    default: null,
  },
})

const emit = defineEmits(['update:open', 'success'])

const modalOpen = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

const loading = ref(false)
const confirming = ref(false)
const createClassOpen = ref(false)
const classNameKeyword = ref('')
const selectedTargetClassId = ref('')
const classOptions = ref<GroupClassRow[]>([])
const scheduleSummaryMap = ref<Record<string, ClassScheduleSummaryDisplay>>({})
const pendingSelectClassId = ref('')

const lessonId = computed(() => String(props.record?.lessonId || '').trim())
const lessonName = computed(() => String(props.record?.lessonName || '').trim() || '当前课程')
const studentId = computed(() => String(props.record?.studentId || '').trim())
const studentName = computed(() => String(props.record?.studentName || '').trim() || '该学员')
const tuitionAccountId = computed(() => String(props.record?.tuitionAccountId || '').trim())
const createClassPreset = computed(() => ({
  lessonId: lessonId.value,
  lessonName: lessonName.value,
  isMultiProduct: false,
}))

const rowSelection = computed(() => ({
  type: 'radio' as const,
  selectedRowKeys: selectedTargetClassId.value ? [selectedTargetClassId.value] : [],
  onChange: (keys: (string | number)[]) => {
    selectedTargetClassId.value = String(keys?.[0] || '').trim()
  },
}))

function resetState() {
  loading.value = false
  confirming.value = false
  createClassOpen.value = false
  classNameKeyword.value = ''
  selectedTargetClassId.value = ''
  classOptions.value = []
  scheduleSummaryMap.value = {}
  pendingSelectClassId.value = ''
}

function formatTeacherNames(record: GroupClassRow) {
  if (!Array.isArray(record.teachers) || record.teachers.length === 0)
    return '-'
  const names = record.teachers.map(item => String(item?.name || '').trim()).filter(Boolean)
  return names.length ? names.join('、') : '-'
}

function compactWeekdayLabel(text?: string) {
  const value = String(text || '').trim()
  if (!value)
    return ''
  return value.startsWith('周') ? value : value.replace(/^星期/, '周')
}

function formatRepeatWeekdayText(prefix: string, weekdays?: string[]) {
  const labels = (Array.isArray(weekdays) ? weekdays : [])
    .map(item => compactWeekdayLabel(item))
    .filter(Boolean)
  return labels.length ? `${prefix}${labels.join('、')}` : prefix
}

function getScheduleRuleText(item?: GroupClassDrawerScheduleItem) {
  const repeatRule = String(item?.repeatRule || '').trim()
  const metaRepeatRule = String(item?.batchMeta?.repeatRule || '').trim().toLowerCase()
  const selectedWeekdays = Array.isArray(item?.batchMeta?.selectedWeekdays)
    ? item.batchMeta.selectedWeekdays
    : []
  const weekdayText = String(item?.weekdayText || '').trim()

  if (metaRepeatRule === 'daily' || repeatRule === '每天重复')
    return '每天'
  if (metaRepeatRule === 'alternateday' || metaRepeatRule === 'alternate_day' || repeatRule === '隔天重复')
    return '隔天'
  if (metaRepeatRule === 'weekly' || repeatRule === '每周重复')
    return formatRepeatWeekdayText('每周', selectedWeekdays)
  if (metaRepeatRule === 'biweekly' || repeatRule === '隔周重复')
    return formatRepeatWeekdayText('隔周', selectedWeekdays)
  if (repeatRule === '单次')
    return '单次'
  if (weekdayText && weekdayText !== '-')
    return weekdayText
  return repeatRule.replace(/重复$/, '').trim()
}

function normalizeTimeText(value?: string) {
  return String(value || '')
    .replace(/[～~]/g, ' ~ ')
    .replace(/\s*-\s*/g, ' ~ ')
    .replace(/\s+/g, ' ')
    .trim()
}

function buildScheduleSummaryDisplay(list: GroupClassDrawerScheduleItem[]): ClassScheduleSummaryDisplay {
  const items: ClassScheduleSummaryItem[] = []
  const seen = new Set<string>()
  list.forEach((item) => {
    const ruleText = getScheduleRuleText(item)
    const timeText = normalizeTimeText(item?.timeText)
    const summary = `${ruleText} ${timeText}`.trim() || timeText || ruleText
    if (!summary)
      return
    const dateRange = String(item?.dateRangeText || '').trim()
    const uniqueKey = `${summary}__${dateRange}`
    if (seen.has(uniqueKey))
      return
    seen.add(uniqueKey)
    items.push({
      summary,
      dateRange,
    })
  })

  if (!items.length) {
    return {
      primaryText: '-',
      extraText: '',
      items: [],
    }
  }

  return {
    primaryText: items[0].summary,
    extraText: items.length > 1 ? `另${items.length - 1}个时段` : '',
    items,
  }
}

async function buildScheduleSummaryMap(list: GroupClassRow[]) {
  const nextSummaryMap: Record<string, ClassScheduleSummaryDisplay> = {}
  await Promise.all(list.map(async (item) => {
    const classId = String(item?.id || '').trim()
    if (!classId) {
      return
    }
    try {
      const res = await getGroupClassDrawerSchedulesApi({ classId })
      nextSummaryMap[classId] = res.code === 200
        ? buildScheduleSummaryDisplay(Array.isArray(res.result?.list) ? res.result.list : [])
        : { primaryText: '-', extraText: '', items: [] }
    }
    catch {
      nextSummaryMap[classId] = { primaryText: '-', extraText: '', items: [] }
    }
  }))
  return nextSummaryMap
}

function getScheduleDisplay(record: GroupClassRow): ClassScheduleSummaryDisplay {
  return scheduleSummaryMap.value[String(record?.id || '').trim()] || {
    primaryText: '-',
    extraText: '',
    items: [],
  }
}

async function loadClasses() {
  const currentLessonId = lessonId.value
  if (!currentLessonId) {
    classOptions.value = []
    scheduleSummaryMap.value = {}
    return
  }

  loading.value = true
  try {
    const res = await pageGroupClassesApi({
      queryModel: {
        lessonIds: [currentLessonId],
        statues: [1],
        isMultiProduct: false,
        className: classNameKeyword.value.trim() || undefined,
      },
      pageRequestModel: {
        needTotal: true,
        pageSize: 200,
        pageIndex: 1,
        skipCount: 0,
      },
    })
    if (res.code !== 200) {
      throw new Error(res.message || '获取班级列表失败')
    }
    const list = Array.isArray(res.result?.list) ? res.result.list : []
    classOptions.value = list
    scheduleSummaryMap.value = await buildScheduleSummaryMap(list)
    if (pendingSelectClassId.value && list.some(item => String(item.id) === pendingSelectClassId.value)) {
      selectedTargetClassId.value = pendingSelectClassId.value
      pendingSelectClassId.value = ''
    }
    if (selectedTargetClassId.value && !list.some(item => String(item.id) === selectedTargetClassId.value))
      selectedTargetClassId.value = ''
  }
  catch (error: any) {
    classOptions.value = []
    scheduleSummaryMap.value = {}
    messageService.error(error?.response?.data?.message || error?.message || '获取班级列表失败')
  }
  finally {
    loading.value = false
  }
}

async function handleConfirm() {
  if (!selectedTargetClassId.value) {
    messageService.warning('请选择班级')
    return
  }
  if (!studentId.value || !tuitionAccountId.value) {
    messageService.warning('缺少学员或课程账户信息')
    return
  }

  confirming.value = true
  try {
    const res = await batchAssignGroupClassStudentsApi({
      classIds: [selectedTargetClassId.value],
      students: [{
        studentId: studentId.value,
        tuitionAccountId: tuitionAccountId.value,
      }],
      enforceClassAssign: true,
    })
    const success = res.code === 200 && (res.result?.success === true || res.data?.success === true)
    if (!success) {
      throw new Error(res.message || '分班失败')
    }
    messageService.success('分班成功')
    emitter.emit(EVENTS.REFRESH_STUDENT_LIST)
    modalOpen.value = false
    emit('success')
  }
  catch (error: any) {
    messageService.error(error?.response?.data?.message || error?.message || '分班失败')
  }
  finally {
    confirming.value = false
  }
}

function handleSearch() {
  loadClasses()
}

function handleRowClick(record: GroupClassRow) {
  selectedTargetClassId.value = String(record?.id || '').trim()
}

function handleClassCreated(payload?: { id?: string }) {
  pendingSelectClassId.value = String(payload?.id || '').trim()
  loadClasses()
}

watch(
  () => props.open,
  (open) => {
    if (!open) {
      resetState()
      return
    }
    loadClasses()
  },
)

watch(
  () => classNameKeyword.value,
  (value, oldValue) => {
    if (!props.open)
      return
    if (oldValue && !String(value || '').trim())
      loadClasses()
  },
)
</script>

<template>
  <div class="choose-group-class-modal-root">
    <a-modal
      v-model:open="modalOpen"
      centered
      class="choose-group-class-modal"
      :keyboard="false"
      :closable="false"
      :mask-closable="false"
      :footer="false"
      :width="800"
    >
      <template #title>
        <div class="choose-group-class-modal__title">
          <span>选择班级</span>
          <a-button type="text" class="choose-group-class-modal__close" @click="modalOpen = false">
            <template #icon>
              <CloseOutlined />
            </template>
          </a-button>
        </div>
      </template>

      <div class="choose-group-class-modal__body">
        <div class="choose-group-class-modal__toolbar">
          <a-input
            v-model:value="classNameKeyword"
            placeholder="请输入班级名称"
            allow-clear
            @press-enter="handleSearch"
          >
            <template #prefix>
              <SearchOutlined />
            </template>
          </a-input>
        </div>

        <div class="choose-group-class-modal__summary">
          <div class="choose-group-class-modal__total">
            共计 {{ classOptions.length }} 个班级
          </div>
          <a-button type="primary" @click="createClassOpen = true">
            创建班级
          </a-button>
        </div>

        <a-table
          row-key="id"
          size="small"
          :loading="loading"
          :data-source="classOptions"
          :pagination="false"
          :row-selection="rowSelection"
          :scroll="{ y: 280 }"
          :custom-row="record => ({
            onClick: () => handleRowClick(record),
          })"
        >
          <a-table-column title="班级名称" data-index="name" key="name" />
          <a-table-column key="studentCount" :width="120">
            <template #title>
              <span class="choose-group-class-modal__column-title">
                学员数
                <QuestionCircleOutlined />
              </span>
            </template>
            <template #default="{ record }">
              {{ Number(record?.studentCount || 0) }}
            </template>
          </a-table-column>
          <a-table-column title="班主任" key="teachers" :width="180">
            <template #default="{ record }">
              {{ formatTeacherNames(record) }}
            </template>
          </a-table-column>
          <a-table-column key="scheduleSummary" :width="220">
            <template #title>
              <span class="choose-group-class-modal__column-title">
                上课时间
                <QuestionCircleOutlined />
              </span>
            </template>
            <template #default="{ record }">
              <template v-if="getScheduleDisplay(record).items.length">
                <a-tooltip placement="topLeft">
                  <template #title>
                    <div class="choose-group-class-modal__schedule-tooltip">
                      <div
                        v-for="(item, index) in getScheduleDisplay(record).items"
                        :key="`${item.summary}-${item.dateRange}-${index}`"
                        class="choose-group-class-modal__schedule-tooltip-item"
                      >
                        <div>{{ item.summary }}</div>
                        <div v-if="item.dateRange" class="choose-group-class-modal__schedule-tooltip-date">
                          {{ item.dateRange }}
                        </div>
                      </div>
                    </div>
                  </template>
                  <div class="choose-group-class-modal__schedule-cell">
                    <div class="choose-group-class-modal__schedule-primary">
                      {{ getScheduleDisplay(record).primaryText }}
                    </div>
                    <div v-if="getScheduleDisplay(record).extraText" class="choose-group-class-modal__schedule-extra">
                      {{ getScheduleDisplay(record).extraText }}
                    </div>
                  </div>
                </a-tooltip>
              </template>
              <template v-else>
                -
              </template>
            </template>
          </a-table-column>
        </a-table>

        <div class="choose-group-class-modal__footer">
          <a-button @click="modalOpen = false">
            取消
          </a-button>
          <a-button type="primary" :loading="confirming" :disabled="!selectedTargetClassId" @click="handleConfirm">
            确定
          </a-button>
        </div>
      </div>
    </a-modal>

    <CreateClassModal
      v-model:open="createClassOpen"
      :edit-record="createClassPreset"
      @created="handleClassCreated"
    />
  </div>
</template>

<style lang="less" scoped>
.choose-group-class-modal-root {
  display: contents;
}

.choose-group-class-modal__title {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.choose-group-class-modal__close {
  margin-right: -8px;
}

.choose-group-class-modal__body {
  padding-top: 8px;
}

.choose-group-class-modal__toolbar {
  margin-bottom: 16px;
}

.choose-group-class-modal__summary {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.choose-group-class-modal__total {
  position: relative;
  padding-left: 10px;
  color: #222;

  &::before {
    position: absolute;
    left: 0;
    top: 4px;
    width: 4px;
    height: 12px;
    border-radius: 2px;
    background: var(--pro-ant-color-primary);
    content: '';
  }
}

.choose-group-class-modal__column-title {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.choose-group-class-modal__footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 16px;
}

.choose-group-class-modal__schedule-cell {
  min-width: 0;
  cursor: help;
}

.choose-group-class-modal__schedule-primary {
  overflow: hidden;
  color: #262626;
  line-height: 20px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.choose-group-class-modal__schedule-extra {
  margin-top: 2px;
  color: #8c8c8c;
  font-size: 12px;
  line-height: 18px;
}

.choose-group-class-modal__schedule-tooltip {
  max-width: 360px;
}

.choose-group-class-modal__schedule-tooltip-item + .choose-group-class-modal__schedule-tooltip-item {
  margin-top: 8px;
  padding-top: 8px;
  border-top: 1px solid rgba(255, 255, 255, 0.12);
}

.choose-group-class-modal__schedule-tooltip-date {
  margin-top: 2px;
  color: rgba(255, 255, 255, 0.75);
  font-size: 12px;
}
</style>
