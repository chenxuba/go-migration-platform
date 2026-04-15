<script setup lang="ts">
import { InfoCircleFilled, QuestionCircleOutlined } from '@ant-design/icons-vue'
import { computed, ref, watch } from 'vue'
import {
  getGroupClassDrawerSchedulesApi,
  moveGroupClassStudentApi,
  pageMoveGroupClassCandidatesApi,
  type GroupClassDrawerScheduleItem,
  type GroupClassRow,
} from '@/api/edu-center/group-class'
import messageService from '@/utils/messageService'

interface StudentRecordLike {
  id?: string
  name?: string
}

const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
  currentClassId: {
    type: String,
    default: '',
  },
  lessonId: {
    type: String,
    default: '',
  },
  lessonName: {
    type: String,
    default: '',
  },
  student: {
    type: Object as () => StudentRecordLike | null,
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
const allFilterRef = ref()
const filterKey = ref(0)
const selectedTeacherId = ref<string | undefined>(undefined)
const classNameKeyword = ref('')
const selectedTargetClassId = ref('')
const allCandidates = ref<GroupClassRow[]>([])
const scheduleSummaryMap = ref<Record<string, string>>({})
const filterDisplayArray = ['classTeacher']
let loadRequestSeq = 0

const studentId = computed(() => String(props.student?.id || '').trim())
const studentName = computed(() => String(props.student?.name || '').trim() || '该学员')
const bannerText = computed(() => `${studentName.value}等1位学员将调整至以下为“${String(props.lessonName || '').trim() || '当前课程'}”下的班级，请选择`)

const rowSelection = computed(() => ({
  type: 'radio' as const,
  selectedRowKeys: selectedTargetClassId.value ? [selectedTargetClassId.value] : [],
  onChange: (keys: (string | number)[]) => {
    selectedTargetClassId.value = String(keys?.[0] || '').trim()
  },
}))

function handleRowClick(record: GroupClassRow) {
  selectedTargetClassId.value = String(record?.id || '').trim()
}

function resetState() {
  loadRequestSeq += 1
  loading.value = false
  confirming.value = false
  selectedTeacherId.value = undefined
  classNameKeyword.value = ''
  selectedTargetClassId.value = ''
  allCandidates.value = []
  scheduleSummaryMap.value = {}
  filterKey.value += 1
}

function syncAllFilterClear(id?: string | number, type?: string) {
  if (!type)
    return
  allFilterRef.value?.clearQuickFilter?.(id, type)
}

function formatTeacherNames(record: GroupClassRow) {
  if (!Array.isArray(record.teachers) || record.teachers.length === 0)
    return '-'
  const names = record.teachers.map(item => String(item?.name || '').trim()).filter(Boolean)
  return names.length ? names.join('、') : '-'
}

function formatScheduleSummary(list: GroupClassDrawerScheduleItem[]) {
  const segments: string[] = []
  const seen = new Set<string>()
  list.forEach((item) => {
    const repeatRule = String(item?.repeatRule || '').trim()
    if (!repeatRule || repeatRule === '单次')
      return
    const weekdayText = String(item?.weekdayText || '').trim()
    const prefix = weekdayText && weekdayText !== '-'
      ? weekdayText
      : repeatRule.replace(/重复$/, '').trim()
    const timeText = String(item?.timeText || '')
      .replace(/[～~]/g, ' ~ ')
      .replace(/\s*-\s*/g, ' ~ ')
      .replace(/\s+/g, ' ')
      .trim()
    const segment = `${prefix} ${timeText}`.trim()
    if (!segment || seen.has(segment))
      return
    seen.add(segment)
    segments.push(segment)
  })
  return segments.join('；') || '-'
}

async function buildScheduleSummaryMap(classIds: string[]) {
  const nextSummaryMap: Record<string, string> = {}
  await Promise.all(classIds.filter(Boolean).map(async (classId) => {
    if (scheduleSummaryMap.value[classId]) {
      nextSummaryMap[classId] = scheduleSummaryMap.value[classId]
      return
    }
    try {
      const res = await getGroupClassDrawerSchedulesApi({ classId })
      if (res.code !== 200) {
        nextSummaryMap[classId] = '-'
        return
      }
      nextSummaryMap[classId] = formatScheduleSummary(Array.isArray(res.result?.list) ? res.result.list : [])
    }
    catch (error) {
      console.error('load class move schedule summary failed', error)
      nextSummaryMap[classId] = '-'
    }
  }))
  return nextSummaryMap
}

async function loadCandidates() {
  const currentRequestSeq = ++loadRequestSeq
  const currentClassId = String(props.currentClassId || '').trim()
  const currentLessonId = String(props.lessonId || '').trim()
  const currentStudentId = studentId.value
  if (!currentClassId || !currentLessonId || !currentStudentId) {
    allCandidates.value = []
    scheduleSummaryMap.value = {}
    return
  }

  loading.value = true
  try {
    const listRes = await pageMoveGroupClassCandidatesApi({
      queryModel: {
        currentClassId,
        studentId: currentStudentId,
        lessonId: currentLessonId,
        className: classNameKeyword.value.trim() || undefined,
        teacherId: String(selectedTeacherId.value || '').trim() || undefined,
      },
      pageRequestModel: {
        needTotal: true,
        pageSize: 200,
        pageIndex: 1,
        skipCount: 0,
      },
    })
    if (listRes.code !== 200) {
      throw new Error(listRes.message || '获取可调整班级失败')
    }

    if (currentRequestSeq !== loadRequestSeq)
      return

    const rawList = Array.isArray(listRes.result?.list) ? listRes.result.list : []
    if (!rawList.length) {
      allCandidates.value = []
      scheduleSummaryMap.value = {}
      return
    }

    allCandidates.value = rawList
    const nextSummaryMap = await buildScheduleSummaryMap(rawList.map(item => String(item.id || '').trim()))
    if (currentRequestSeq !== loadRequestSeq)
      return
    scheduleSummaryMap.value = nextSummaryMap
  }
  catch (error: any) {
    if (currentRequestSeq !== loadRequestSeq)
      return
    console.error('load move class candidates failed', error)
    allCandidates.value = []
    scheduleSummaryMap.value = {}
    messageService.error(error?.response?.data?.message || error?.message || '获取可调整班级失败')
  }
  finally {
    if (currentRequestSeq === loadRequestSeq)
      loading.value = false
  }
}

function handleClassTeacherFilterChange(value?: string, _isClearAll?: boolean, id?: string | number, type?: string) {
  selectedTeacherId.value = value ? String(value).trim() : undefined
  syncAllFilterClear(id, type)
  if (props.open)
    loadCandidates()
}

function handleClassNameSearch(value?: string, id?: string | number, type?: string) {
  classNameKeyword.value = String(value || '').trim()
  syncAllFilterClear(id, type)
  if (props.open)
    loadCandidates()
}

async function handleConfirm() {
  const fromClassId = String(props.currentClassId || '').trim()
  const toClassId = String(selectedTargetClassId.value || '').trim()
  const currentStudentId = studentId.value
  if (!fromClassId || !toClassId || !currentStudentId) {
    messageService.warning('请选择目标班级')
    return
  }
  confirming.value = true
  try {
    const res = await moveGroupClassStudentApi({
      fromClassId,
      toClassId,
      studentId: currentStudentId,
    })
    if (res.code !== 200)
      throw new Error(res.message || '调至其他班失败')
    messageService.success('已调至其他班')
    modalOpen.value = false
    emit('success')
  }
  catch (error: any) {
    messageService.error(error?.response?.data?.message || error?.message || '调至其他班失败')
  }
  finally {
    confirming.value = false
  }
}

watch(
  () => props.open,
  (open) => {
    if (!open) {
      resetState()
      return
    }
    loadCandidates()
  },
)

watch(allCandidates, (list) => {
  if (!selectedTargetClassId.value)
    return
  const exists = list.some(item => String(item?.id || '').trim() === selectedTargetClassId.value)
  if (!exists)
    selectedTargetClassId.value = ''
})
</script>

<template>
  <a-modal
    v-model:open="modalOpen"
    title="调至其他班"
    :width="860"
    centered
    :confirm-loading="confirming"
    :ok-button-props="{ disabled: !selectedTargetClassId }"
    ok-text="确定"
    cancel-text="取消"
    @ok="handleConfirm"
  >
    <div class="move-class-modal">
      <div class="move-class-modal__banner">
        <InfoCircleFilled class="move-class-modal__banner-icon" />
        <span>{{ bannerText }}</span>
      </div>

      <div class="move-class-modal__filter-wrap">
        <all-filter
          ref="allFilterRef"
          :key="filterKey"
          :display-array="filterDisplayArray"
          :is-quick-show="false"
          :is-show-search-input="true"
          search-label="班级名称"
          search-placeholder="请输入班级名称"
          @update:classTeacherFilter="handleClassTeacherFilterChange"
          @searchInputFun="handleClassNameSearch"
        />
      </div>

      <a-table
        row-key="id"
        size="small"
        :loading="loading"
        :data-source="allCandidates"
        :pagination="false"
        :row-selection="rowSelection"
        :scroll="{ y: 320 }"
        :custom-row="record => ({
          onClick: () => {
            handleRowClick(record)
          },
        })"
      >
        <a-table-column title="班级名称" data-index="name" key="name" />
        <a-table-column title="学员数" key="studentCount" :width="120">
          <template #title>
            <span class="move-class-modal__column-title">
              学员数
              <QuestionCircleOutlined />
            </span>
          </template>
          <template #default="{ record }">
            {{ Number(record?.studentCount || 0) }}
          </template>
        </a-table-column>
        <a-table-column title="班主任" key="teachers" :width="220">
          <template #default="{ record }">
            {{ formatTeacherNames(record) }}
          </template>
        </a-table-column>
        <a-table-column key="scheduleSummary" :width="220">
          <template #title>
            <span class="move-class-modal__column-title">
              上课时间
              <a-tooltip title="仅展示重复日程，单次日程详情请查看">
                <QuestionCircleOutlined />
              </a-tooltip>
            </span>
          </template>
          <template #default="{ record }">
            {{ scheduleSummaryMap[String(record?.id || '').trim()] || '-' }}
          </template>
        </a-table-column>
      </a-table>
    </div>
  </a-modal>
</template>

<style lang="less" scoped>
.move-class-modal {
  padding-top: 4px;
}

.move-class-modal__banner {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 16px;
  padding: 10px 14px;
  background: #eaf2ff;
  border-radius: 8px;
  color: #1677ff;
  line-height: 22px;
}

.move-class-modal__banner-icon {
  font-size: 14px;
}

.move-class-modal__filter-wrap {
}

.move-class-modal__filter-wrap :deep(.selectBox) {
  margin-bottom: 0;
}

.move-class-modal__column-title {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}
</style>
