<script setup lang="ts">
import { CaretDownOutlined, CheckOutlined, CloseOutlined, SearchOutlined } from '@ant-design/icons-vue'
import { computed, nextTick, ref, watch } from 'vue'
import { pageGroupClassSelectionApi } from '@/api/edu-center/group-class'
import { pageOneToOneSelectionApi } from '@/api/edu-center/one-to-one'
import type { NoticePickerCompletePayload, NoticePickerSelection, NoticePickerSource } from './notice-picker.types'
import messageService from '@/utils/messageService'

const props = withDefaults(defineProps<{
  title?: string
  selectedStudents?: NoticePickerSelection[]
  selectedSources?: NoticePickerSource[]
}>(), {
  title: '选择班级/1v1',
  selectedStudents: () => [],
  selectedSources: () => [],
})

const emit = defineEmits<{
  (e: 'complete', payload: NoticePickerCompletePayload): void
}>()

const open = defineModel<boolean>({ default: false })

const pickerType = ref<'class' | 'one_to_one'>('class')
const pickerKeyword = ref('')
const pickerLoading = ref(false)
const classTargetList = ref<any[]>([])
const oneToOneTargetList = ref<any[]>([])
const expandedClassIds = ref<string[]>([])
const expandedOneToOneIds = ref<string[]>([])
const draftSelectedStudents = ref<NoticePickerSelection[]>([])
const pickerRequestSeq = ref(0)
const sourceRegistry = ref<Record<string, NoticePickerSource>>({})

let pickerSearchTimer: ReturnType<typeof setTimeout> | undefined

const pickerTabs = [
  { key: 'class', label: '班级' },
  { key: 'one_to_one', label: '1对1' },
]

const pickerPlaceholder = computed(() => pickerType.value === 'class' ? '搜索班级名称' : '搜索1对1名称')

function cloneSelectedStudents(list: NoticePickerSelection[]) {
  return Array.isArray(list)
    ? list.filter(item => item.selectionType !== 'source').map(item => ({ ...item, selectionType: 'student' as const }))
    : []
}

function cloneSources(list: NoticePickerSource[]) {
  return Array.isArray(list)
    ? list.map(item => ({
        ...item,
        students: Array.isArray(item.students) ? item.students.map(student => ({ ...student })) : [],
      }))
    : []
}

function resetPickerState() {
  pickerKeyword.value = ''
  classTargetList.value = []
  oneToOneTargetList.value = []
  expandedClassIds.value = []
  expandedOneToOneIds.value = []
  pickerLoading.value = false
  pickerRequestSeq.value += 1
  sourceRegistry.value = {}
  if (pickerSearchTimer) {
    clearTimeout(pickerSearchTimer)
    pickerSearchTimer = undefined
  }
}

function buildSourceKey(source: Pick<NoticePickerSelection, 'sourceType' | 'sourceId'>) {
  return `${String(source.sourceType || '')}:${String(source.sourceId || '')}`
}

function buildSelectedStudentKey(item: Pick<NoticePickerSelection, 'sourceType' | 'sourceId' | 'studentId'>) {
  return `${String(item.sourceType || '')}:${String(item.sourceId || '')}:${String(item.studentId || '')}`
}

function buildClassStudentSelection(classItem: any, student: any): NoticePickerSelection {
  return {
    sourceType: 'class',
    sourceId: String(classItem?.id || ''),
    sourceName: String(classItem?.name || ''),
    studentId: String(student?.id || ''),
    studentName: String(student?.name || ''),
    tuitionAccountId: String(student?.tuitionAccountId || ''),
    isBind: student?.isBind !== false,
    selectionType: 'student',
  }
}

function buildOneToOneSelection(item: any): NoticePickerSelection {
  return {
    sourceType: 'one_to_one',
    sourceId: String(item?.id || ''),
    sourceName: String(item?.name || item?.lessonName || item?.studentName || '1对1'),
    studentId: String(item?.studentId || item?.id || ''),
    studentName: String(item?.studentName || item?.name || ''),
    tuitionAccountId: String(item?.tuitionAccountId || ''),
    isBind: item?.isBindChild !== false,
    selectionType: 'student',
  }
}

function buildClassSource(classItem: any): NoticePickerSource {
  return {
    sourceType: 'class',
    sourceId: String(classItem?.id || ''),
    sourceName: String(classItem?.name || ''),
    students: (Array.isArray(classItem?.students) ? classItem.students : []).map((student: any) => buildClassStudentSelection(classItem, student)),
  }
}

function buildOneToOneSource(item: any): NoticePickerSource {
  return {
    sourceType: 'one_to_one',
    sourceId: String(item?.id || ''),
    sourceName: String(item?.name || item?.lessonName || item?.studentName || '1对1'),
    students: [buildOneToOneSelection(item)].filter(student => student.studentId),
  }
}

function rememberSource(source: NoticePickerSource) {
  const key = buildSourceKey(source)
  sourceRegistry.value = {
    ...sourceRegistry.value,
    [key]: {
      ...source,
      students: cloneSelectedStudents(source.students),
    },
  }
}

function hydrateSourceRegistry() {
  sourceRegistry.value = {}
  cloneSources(props.selectedSources).forEach(source => rememberSource(source))
}

function isDraftStudentSelected(item: NoticePickerSelection) {
  const key = buildSelectedStudentKey(item)
  return draftSelectedStudents.value.some(selectedItem => buildSelectedStudentKey(selectedItem) === key)
}

function getSelectedCountBySource(sourceType: 'class' | 'one_to_one', sourceId: string) {
  return draftSelectedStudents.value.filter(item => item.sourceType === sourceType && String(item.sourceId || '') === String(sourceId || '')).length
}

function getSelectedSourceCount(sourceType: 'class' | 'one_to_one') {
  return draftSelectedStudents.value.filter(item => item.sourceType === sourceType).length
}

function isSourceSelected(sourceType: 'class' | 'one_to_one', sourceId: string) {
  if (sourceType === 'one_to_one')
    return getSelectedCountBySource(sourceType, sourceId) > 0

  const currentClass = classTargetList.value.find(item => String(item?.id || '') === String(sourceId || ''))
  const total = Array.isArray(currentClass?.students) ? currentClass.students.length : 0
  return total > 0 && getSelectedCountBySource(sourceType, sourceId) === total
}

function toggleDraftStudent(item: NoticePickerSelection) {
  const key = buildSelectedStudentKey(item)
  const index = draftSelectedStudents.value.findIndex(selectedItem => buildSelectedStudentKey(selectedItem) === key)
  if (index >= 0) {
    draftSelectedStudents.value.splice(index, 1)
    return
  }
  draftSelectedStudents.value.push({ ...item, selectionType: 'student' })
}

function toggleClassExpanded(classId: string) {
  const currentId = String(classId || '')
  const index = expandedClassIds.value.indexOf(currentId)
  if (index >= 0) {
    expandedClassIds.value.splice(index, 1)
    return
  }
  expandedClassIds.value.push(currentId)
}

function toggleOneToOneExpanded(itemId: string) {
  const currentId = String(itemId || '')
  const index = expandedOneToOneIds.value.indexOf(currentId)
  if (index >= 0) {
    expandedOneToOneIds.value.splice(index, 1)
    return
  }
  expandedOneToOneIds.value.push(currentId)
}

function handleSelectClassStudent(classItem: any, student: any) {
  rememberSource(buildClassSource(classItem))
  toggleDraftStudent(buildClassStudentSelection(classItem, student))
}

function handleSelectOneToOne(item: any) {
  rememberSource(buildOneToOneSource(item))
  toggleDraftStudent(buildOneToOneSelection(item))
}

function handleSelectSource(sourceType: 'class' | 'one_to_one', item: any) {
  if (sourceType === 'one_to_one') {
    rememberSource(buildOneToOneSource(item))
    toggleDraftStudent(buildOneToOneSelection(item))
    return
  }

  rememberSource(buildClassSource(item))

  const students = Array.isArray(item?.students) ? item.students : []
  const allSelected = students.length > 0 && students.every(student => isDraftStudentSelected(buildClassStudentSelection(item, student)))
  if (allSelected) {
    const selectedKeys = new Set(students.map(student => buildSelectedStudentKey(buildClassStudentSelection(item, student))))
    draftSelectedStudents.value = draftSelectedStudents.value.filter(selectedItem => !selectedKeys.has(buildSelectedStudentKey(selectedItem)))
    return
  }

  students.forEach((student) => {
    const selection = buildClassStudentSelection(item, student)
    if (!isDraftStudentSelected(selection))
      draftSelectedStudents.value.push(selection)
  })
}

function handleInviteFollow() {
  messageService.info('邀请关注功能待接入')
}

async function loadClassTargets(currentSeq: number) {
  const res = await pageGroupClassSelectionApi({
    queryModel: {
      className: String(pickerKeyword.value || '').trim() || undefined,
      status: [1],
    },
    pageRequestModel: {
      needTotal: true,
      pageSize: 50,
      pageIndex: 1,
      skipCount: 0,
    },
  })
  if (currentSeq !== pickerRequestSeq.value)
    return
  if (res.code !== 200)
    throw new Error(res.message || '获取班级列表失败')
  classTargetList.value = Array.isArray(res.result?.list) ? res.result.list : []
  classTargetList.value.forEach(item => rememberSource(buildClassSource(item)))
  expandedClassIds.value = []
}

async function loadOneToOneTargets(currentSeq: number) {
  const res = await pageOneToOneSelectionApi({
    queryModel: {
      searchKey: String(pickerKeyword.value || '').trim() || undefined,
      status: [1],
    },
    pageRequestModel: {
      needTotal: true,
      pageSize: 50,
      pageIndex: 1,
      skipCount: 0,
    },
  })
  if (currentSeq !== pickerRequestSeq.value)
    return
  if (res.code !== 200)
    throw new Error(res.message || '获取1对1列表失败')
  oneToOneTargetList.value = Array.isArray(res.result?.list) ? res.result.list : []
  oneToOneTargetList.value.forEach(item => rememberSource(buildOneToOneSource(item)))
  expandedOneToOneIds.value = oneToOneTargetList.value.slice(0, 1).map(item => String(item?.id || ''))
}

async function loadPickerData() {
  const currentSeq = ++pickerRequestSeq.value
  pickerLoading.value = true
  try {
    if (pickerType.value === 'class') {
      oneToOneTargetList.value = []
      expandedOneToOneIds.value = []
      await loadClassTargets(currentSeq)
    }
    else {
      classTargetList.value = []
      expandedClassIds.value = []
      await loadOneToOneTargets(currentSeq)
    }
  }
  catch (error: any) {
    classTargetList.value = []
    oneToOneTargetList.value = []
    expandedClassIds.value = []
    expandedOneToOneIds.value = []
    messageService.error(error?.response?.data?.message || error?.message || '加载班级/学员失败')
  }
  finally {
    if (currentSeq === pickerRequestSeq.value)
      pickerLoading.value = false
  }
}

function buildSelectedSources() {
  const grouped = new Map<string, NoticePickerSource>()
  draftSelectedStudents.value.forEach((student) => {
    const key = buildSourceKey(student)
    if (grouped.has(key))
      return

    const cached = sourceRegistry.value[key]
    if (cached) {
      grouped.set(key, {
        ...cached,
        students: cloneSelectedStudents(cached.students),
      })
      return
    }

    grouped.set(key, {
      sourceType: student.sourceType,
      sourceId: student.sourceId,
      sourceName: student.sourceName,
      students: [{ ...student }],
    })
  })
  return [...grouped.values()]
}

async function handleCompletePicker() {
  const selectedStudents = cloneSelectedStudents(draftSelectedStudents.value)
  const selectedSources = buildSelectedSources()
  await nextTick()
  emit('complete', { selectedStudents, selectedSources })
  open.value = false
}

watch(() => open.value, (visible) => {
  if (!visible) {
    resetPickerState()
    return
  }
  hydrateSourceRegistry()
  draftSelectedStudents.value = cloneSelectedStudents(props.selectedStudents)
  void loadPickerData()
})

watch(() => pickerType.value, () => {
  if (!open.value)
    return
  pickerKeyword.value = ''
  void loadPickerData()
})

watch(() => pickerKeyword.value, () => {
  if (!open.value)
    return
  if (pickerSearchTimer)
    clearTimeout(pickerSearchTimer)
  pickerSearchTimer = setTimeout(() => {
    void loadPickerData()
  }, 300)
})
</script>

<template>
  <a-modal
    :open="open"
    centered
    class="afterSchoolTasksModel__student-picker-modal"
    :body-style="{ padding: 0 }"
    :keyboard="false"
    :closable="false"
    :mask-closable="false"
    width="800px"
    destroy-on-close
  >
    <template #title>
      <div class="afterSchoolTasksModel__student-picker-title">
        <span>{{ title }}</span>
        <a-button type="text" class="afterSchoolTasksModel__student-picker-close" @click="open = false">
          <template #icon>
            <CloseOutlined />
          </template>
        </a-button>
      </div>
    </template>

    <div class="afterSchoolTasksModel__student-picker">
      <div class="afterSchoolTasksModel__student-picker-sidebar">
        <div
          v-for="item in pickerTabs"
          :key="item.key"
          :class="{ 'is-active': pickerType === item.key }"
          class="afterSchoolTasksModel__student-picker-tab"
          @click="pickerType = item.key as 'class' | 'one_to_one'"
        >
          <span class="afterSchoolTasksModel__student-picker-tab-label">{{ item.label }}</span>
          <span
            v-if="getSelectedSourceCount(item.key as 'class' | 'one_to_one') > 0"
            class="afterSchoolTasksModel__student-picker-tab-count"
          >
            {{ getSelectedSourceCount(item.key as 'class' | 'one_to_one') }}
          </span>
        </div>
      </div>

      <div class="afterSchoolTasksModel__student-picker-main">
        <div class="afterSchoolTasksModel__student-picker-toolbar">
          <a-input
            v-model:value="pickerKeyword"
            :placeholder="pickerPlaceholder"
            allow-clear
          >
            <template #prefix>
              <SearchOutlined />
            </template>
          </a-input>
        </div>

        <a-spin :spinning="pickerLoading" class="afterSchoolTasksModel__student-picker-spin">
          <div class="afterSchoolTasksModel__student-picker-content">
            <template v-if="pickerType === 'class'">
              <template v-if="classTargetList.length > 0">
                <div
                  v-for="classItem in classTargetList"
                  :key="classItem.id"
                  class="afterSchoolTasksModel__student-group"
                >
                  <div class="afterSchoolTasksModel__student-group-header">
                    <span
                      class="afterSchoolTasksModel__student-checkbox afterSchoolTasksModel__student-checkbox--button"
                      :class="{ 'is-selected': isSourceSelected('class', classItem.id) }"
                      @click.stop="handleSelectSource('class', classItem)"
                    >
                      <CheckOutlined v-if="isSourceSelected('class', classItem.id)" />
                    </span>
                    <div class="afterSchoolTasksModel__student-group-header-main" @click="toggleClassExpanded(classItem.id)">
                      <span class="afterSchoolTasksModel__student-group-title">
                        {{ classItem.name }}（{{ getSelectedCountBySource('class', classItem.id) }}/{{ classItem.students?.length || 0 }}）
                      </span>
                      <CaretDownOutlined
                        class="afterSchoolTasksModel__student-group-arrow afterSchoolTasksModel__student-group-arrow--inline"
                        :class="{ 'is-collapsed': !expandedClassIds.includes(String(classItem.id || '')) }"
                      />
                    </div>
                  </div>

                  <div v-show="expandedClassIds.includes(String(classItem.id || ''))" class="afterSchoolTasksModel__student-list">
                    <div
                      v-for="student in classItem.students || []"
                      :key="student.id"
                      class="afterSchoolTasksModel__student-row"
                      @click="handleSelectClassStudent(classItem, student)"
                    >
                      <span class="afterSchoolTasksModel__student-radio" :class="{ 'is-selected': isDraftStudentSelected(buildClassStudentSelection(classItem, student)) }">
                        <CheckOutlined v-if="isDraftStudentSelected(buildClassStudentSelection(classItem, student))" />
                      </span>
                      <span class="afterSchoolTasksModel__student-name">
                        {{ student.name }}
                      </span>
                      <template v-if="student.isBind === false">
                        <span class="afterSchoolTasksModel__student-warning">
                          未关注家校平台，无法发送通知
                        </span>
                        <a class="afterSchoolTasksModel__student-link" @click.stop="handleInviteFollow">邀请关注</a>
                      </template>
                    </div>
                  </div>
                </div>
              </template>
              <a-empty v-else description="暂无班级数据" />
            </template>

            <template v-else>
              <template v-if="oneToOneTargetList.length > 0">
                <div
                  v-for="item in oneToOneTargetList"
                  :key="item.id"
                  class="afterSchoolTasksModel__student-group afterSchoolTasksModel__student-group--plain"
                >
                  <div class="afterSchoolTasksModel__student-group-header afterSchoolTasksModel__student-group-header--plain">
                    <span
                      class="afterSchoolTasksModel__student-checkbox afterSchoolTasksModel__student-checkbox--button"
                      :class="{ 'is-selected': isSourceSelected('one_to_one', item.id) }"
                      @click.stop="handleSelectSource('one_to_one', item)"
                    >
                      <CheckOutlined v-if="isSourceSelected('one_to_one', item.id)" />
                    </span>
                    <div class="afterSchoolTasksModel__student-group-header-main" @click="toggleOneToOneExpanded(item.id)">
                      <span class="afterSchoolTasksModel__student-group-title">
                        {{ item.name || item.lessonName || item.studentName || '1对1' }}（{{ isDraftStudentSelected(buildOneToOneSelection(item)) ? 1 : 0 }}/1）
                      </span>
                      <CaretDownOutlined
                        class="afterSchoolTasksModel__student-group-arrow afterSchoolTasksModel__student-group-arrow--inline"
                        :class="{ 'is-collapsed': !expandedOneToOneIds.includes(String(item.id || '')) }"
                      />
                    </div>
                  </div>

                  <div v-show="expandedOneToOneIds.includes(String(item.id || ''))" class="afterSchoolTasksModel__student-list">
                    <div
                      class="afterSchoolTasksModel__student-row"
                      @click="handleSelectOneToOne(item)"
                    >
                      <span class="afterSchoolTasksModel__student-radio" :class="{ 'is-selected': isDraftStudentSelected(buildOneToOneSelection(item)) }">
                        <CheckOutlined v-if="isDraftStudentSelected(buildOneToOneSelection(item))" />
                      </span>
                      <span class="afterSchoolTasksModel__student-name">
                        {{ item.studentName || item.name || '-' }}
                      </span>
                      <template v-if="item.isBindChild === false">
                        <span class="afterSchoolTasksModel__student-warning">
                          未关注家校平台，无法发送通知
                        </span>
                        <a class="afterSchoolTasksModel__student-link" @click.stop="handleInviteFollow">邀请关注</a>
                      </template>
                    </div>
                  </div>
                </div>
              </template>
              <a-empty v-else description="暂无1对1数据" />
            </template>
          </div>
        </a-spin>
      </div>
    </div>

    <template #footer>
      <div class="afterSchoolTasksModel__student-picker-footer">
        <a-button @click="open = false">
          关闭
        </a-button>
        <a-button type="primary" @click="handleCompletePicker">
          完成
        </a-button>
      </div>
    </template>
  </a-modal>
</template>

<style scoped lang="less">
.afterSchoolTasksModel__student-picker-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.afterSchoolTasksModel__student-picker-close {
  margin-right: -8px;
}

.afterSchoolTasksModel__student-picker {
  display: flex;
  height: 650px;
}

.afterSchoolTasksModel__student-picker-sidebar {
  width: 160px;
  border-right: 1px solid #f0f0f0;
  background: #fff;
}

.afterSchoolTasksModel__student-picker-tab {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 22px 28px;
  color: #595959;
  font-size: 16px;
  line-height: 24px;
  cursor: pointer;
  transition: all 0.2s ease;

  &.is-active {
    background: #f2f8ff;
    color: var(--pro-ant-color-primary);
    font-weight: 600;

    &::before {
      position: absolute;
      left: 12px;
      top: 50%;
      width: 4px;
      height: 14px;
      border-radius: 2px;
      background: var(--pro-ant-color-primary);
      transform: translateY(-50%);
      content: '';
    }
  }
}

.afterSchoolTasksModel__student-picker-tab-label {
  min-width: 0;
}

.afterSchoolTasksModel__student-picker-tab-count {
  min-width: 22px;
  height: 22px;
  padding: 0 6px;
  border-radius: 999px;
  background: rgba(255, 77, 79, 0.12);
  color: #ff4d4f;
  font-size: 12px;
  line-height: 22px;
  text-align: center;
  font-weight: 600;
  flex: none;
}

.afterSchoolTasksModel__student-picker-tab.is-active .afterSchoolTasksModel__student-picker-tab-count {
  background: #ff4d4f;
  color: #fff;
}

.afterSchoolTasksModel__student-picker-main {
  display: flex;
  flex: 1;
  min-width: 0;
  flex-direction: column;
}

.afterSchoolTasksModel__student-picker-toolbar {
  padding: 16px 24px;
  border-bottom: 1px solid #f0f0f0;
}

.afterSchoolTasksModel__student-picker-spin {
  flex: 1;
  min-height: 0;
}

.afterSchoolTasksModel__student-picker-content {
  height: 100%;
  overflow: auto;
  padding: 18px 24px 20px;
  overscroll-behavior: contain;
  scrollbar-gutter: stable;
  scrollbar-width: thin;
  scrollbar-color: #cfd6e4 transparent;
}

.afterSchoolTasksModel__student-group + .afterSchoolTasksModel__student-group {
  margin-top: 18px;
}

.afterSchoolTasksModel__student-group-header {
  display: flex;
  align-items: center;
  justify-content: flex-start;
  gap: 12px;
  color: #262626;
  font-size: 16px;
  line-height: 24px;
  font-weight: 600;
}

.afterSchoolTasksModel__student-group-header--plain {
  cursor: default;
}

.afterSchoolTasksModel__student-group-header-main {
  display: inline-flex;
  align-items: center;
  gap: 2px;
  min-width: 0;
  flex: none;
  cursor: pointer;
}

.afterSchoolTasksModel__student-group-arrow {
  color: #999;
  font-size: 12px;
  flex: none;
  transition: transform 0.2s ease;

  &.is-collapsed {
    transform: rotate(-90deg);
  }
}

.afterSchoolTasksModel__student-group-arrow--inline {
  margin-left: 2px;
}

.afterSchoolTasksModel__student-group-title {
  color: #262626;
}

.afterSchoolTasksModel__student-list {
  padding-top: 8px;
  padding-left: 24px;
}

.afterSchoolTasksModel__student-row {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 6px 0;
  cursor: pointer;
}

.afterSchoolTasksModel__student-checkbox,
.afterSchoolTasksModel__student-radio {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 18px;
  height: 18px;
  flex: none;
  border: 1px solid #d9d9d9;
  background: #fff;
  color: #fff;
  font-size: 12px;
  transition: all 0.2s ease;
}

.afterSchoolTasksModel__student-checkbox {
  border-radius: 4px;

  &.is-selected {
    border-color: var(--pro-ant-color-primary);
    background: var(--pro-ant-color-primary);
  }
}

.afterSchoolTasksModel__student-checkbox--button {
  cursor: pointer;
}

.afterSchoolTasksModel__student-radio {
  border-radius: 50%;

  &.is-selected {
    border-color: var(--pro-ant-color-primary);
    background: var(--pro-ant-color-primary);
  }
}

:deep(.afterSchoolTasksModel__student-checkbox .anticon),
:deep(.afterSchoolTasksModel__student-radio .anticon) {
  transform: scale(0.85);
}

.afterSchoolTasksModel__student-name {
  color: #262626;
  font-size: 14px;
  line-height: 22px;
}

.afterSchoolTasksModel__student-warning {
  color: #fa8c16;
  font-size: 13px;
  line-height: 20px;
}

.afterSchoolTasksModel__student-link {
  color: var(--pro-ant-color-primary);
  font-size: 13px;
  line-height: 20px;
}

.afterSchoolTasksModel__student-picker-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 12px 16px;
  border-top: 1px solid #f0f0f0;
}

:deep(.afterSchoolTasksModel__student-picker-modal .ant-modal-header) {
  margin-bottom: 0;
}

:deep(.afterSchoolTasksModel__student-picker-modal .ant-modal-body) {
  padding: 0 !important;
}

:deep(.afterSchoolTasksModel__student-picker-spin.ant-spin-nested-loading) {
  height: 100%;
}

:deep(.afterSchoolTasksModel__student-picker-spin .ant-spin-container) {
  height: 100%;
}

:deep(.afterSchoolTasksModel__student-picker-content::-webkit-scrollbar) {
  width: 8px;
}

:deep(.afterSchoolTasksModel__student-picker-content::-webkit-scrollbar-track) {
  background: transparent;
}

:deep(.afterSchoolTasksModel__student-picker-content::-webkit-scrollbar-thumb) {
  border: 2px solid transparent;
  border-radius: 999px;
  background: #cfd6e4;
  background-clip: padding-box;
}

:deep(.afterSchoolTasksModel__student-picker-content::-webkit-scrollbar-thumb:hover) {
  background: #b9c3d4;
  background-clip: padding-box;
}
</style>
