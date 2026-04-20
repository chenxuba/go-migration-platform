<script setup lang="ts">
import { CaretDownOutlined, CheckOutlined, CloseOutlined, SearchOutlined } from '@ant-design/icons-vue'
import { computed, nextTick, ref, watch } from 'vue'
import type { NoticePickerCompletePayload, NoticePickerSelection, NoticePickerSource } from './notice-picker.types'
import messageService from '@/utils/messageService'

const props = withDefaults(defineProps<{
  title?: string
  selectedStudents?: NoticePickerSelection[]
  selectedSources?: NoticePickerSource[]
}>(), {
  title: '选择学员',
  selectedStudents: () => [],
  selectedSources: () => [],
})

const emit = defineEmits<{
  (e: 'complete', payload: NoticePickerCompletePayload): void
}>()

const open = defineModel<boolean>({ default: false })

const pickerType = ref<'class' | 'one_to_one'>('class')
const pickerKeyword = ref('')
const expandedClassIds = ref<string[]>([])
const expandedOneToOneIds = ref<string[]>([])
const draftSelectedStudents = ref<NoticePickerSelection[]>([])
const sourceRegistry = ref<Record<string, NoticePickerSource>>({})

const pickerTabs = [
  { key: 'class', label: '班级' },
  { key: 'one_to_one', label: '1对1' },
]

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

function buildSourceKey(source: Pick<NoticePickerSelection, 'sourceType' | 'sourceId'>) {
  return `${String(source.sourceType || '')}:${String(source.sourceId || '')}`
}

function buildSelectedStudentKey(item: Pick<NoticePickerSelection, 'sourceType' | 'sourceId' | 'studentId'>) {
  return `${String(item.sourceType || '')}:${String(item.sourceId || '')}:${String(item.studentId || '')}`
}

function resetPickerState() {
  pickerKeyword.value = ''
  expandedClassIds.value = []
  expandedOneToOneIds.value = []
  sourceRegistry.value = {}
}

function rememberSources(list: NoticePickerSource[]) {
  const nextRegistry: Record<string, NoticePickerSource> = {}
  cloneSources(list).forEach((source) => {
    nextRegistry[buildSourceKey(source)] = source
  })
  sourceRegistry.value = nextRegistry
}

const classTargetList = computed(() => {
  const keyword = String(pickerKeyword.value || '').trim().toLowerCase()
  return Object.values(sourceRegistry.value)
    .filter(source => source.sourceType === 'class')
    .map((source) => {
      const students = !keyword
        ? source.students
        : source.students.filter((student) => {
            const sourceName = String(source.sourceName || '').toLowerCase()
            const studentName = String(student.studentName || '').toLowerCase()
            return sourceName.includes(keyword) || studentName.includes(keyword)
          })
      return {
        ...source,
        students,
      }
    })
    .filter(source => source.students.length > 0)
})

const oneToOneTargetList = computed(() => {
  const keyword = String(pickerKeyword.value || '').trim().toLowerCase()
  return Object.values(sourceRegistry.value)
    .filter(source => source.sourceType === 'one_to_one')
    .map((source) => {
      const students = !keyword
        ? source.students
        : source.students.filter((student) => {
            const sourceName = String(source.sourceName || '').toLowerCase()
            const studentName = String(student.studentName || '').toLowerCase()
            return sourceName.includes(keyword) || studentName.includes(keyword)
          })
      return {
        ...source,
        students,
      }
    })
    .filter(source => source.students.length > 0)
})

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
  const currentSource = sourceRegistry.value[`${sourceType}:${sourceId}`]
  const total = Array.isArray(currentSource?.students) ? currentSource.students.length : 0
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

function handleSelectClassStudent(source: NoticePickerSource, student: NoticePickerSelection) {
  toggleDraftStudent({
    ...student,
    sourceType: source.sourceType,
    sourceId: source.sourceId,
    sourceName: source.sourceName,
  })
}

function handleSelectOneToOne(source: NoticePickerSource) {
  const student = source.students[0]
  if (!student)
    return
  toggleDraftStudent({
    ...student,
    sourceType: source.sourceType,
    sourceId: source.sourceId,
    sourceName: source.sourceName,
  })
}

function handleSelectSource(sourceType: 'class' | 'one_to_one', source: NoticePickerSource) {
  if (sourceType === 'one_to_one') {
    handleSelectOneToOne(source)
    return
  }

  const students = Array.isArray(source.students) ? source.students : []
  const allSelected = students.length > 0 && students.every(student => isDraftStudentSelected(student))
  if (allSelected) {
    const selectedKeys = new Set(students.map(student => buildSelectedStudentKey(student)))
    draftSelectedStudents.value = draftSelectedStudents.value.filter(selectedItem => !selectedKeys.has(buildSelectedStudentKey(selectedItem)))
    return
  }

  students.forEach((student) => {
    if (!isDraftStudentSelected(student))
      draftSelectedStudents.value.push({ ...student, selectionType: 'student' })
  })
}

function handleInviteFollow() {
  messageService.info('邀请关注功能待接入')
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
  rememberSources(props.selectedSources)
  draftSelectedStudents.value = cloneSelectedStudents(props.selectedStudents)
  expandedClassIds.value = []
  expandedOneToOneIds.value = oneToOneTargetList.value.slice(0, 1).map(item => String(item.sourceId || ''))
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
            placeholder="搜索学员名称"
            allow-clear
          >
            <template #prefix>
              <SearchOutlined />
            </template>
          </a-input>
        </div>

        <div class="afterSchoolTasksModel__student-picker-content">
          <template v-if="pickerType === 'class'">
            <template v-if="classTargetList.length > 0">
              <div
                v-for="classItem in classTargetList"
                :key="classItem.sourceId"
                class="afterSchoolTasksModel__student-group"
              >
                <div class="afterSchoolTasksModel__student-group-header">
                  <span
                    class="afterSchoolTasksModel__student-checkbox afterSchoolTasksModel__student-checkbox--button"
                    :class="{ 'is-selected': isSourceSelected('class', classItem.sourceId) }"
                    @click.stop="handleSelectSource('class', classItem)"
                  >
                    <CheckOutlined v-if="isSourceSelected('class', classItem.sourceId)" />
                  </span>
                  <div class="afterSchoolTasksModel__student-group-header-main" @click="toggleClassExpanded(classItem.sourceId)">
                    <span class="afterSchoolTasksModel__student-group-title">
                      {{ classItem.sourceName }}（{{ getSelectedCountBySource('class', classItem.sourceId) }}/{{ classItem.students?.length || 0 }}）
                    </span>
                    <CaretDownOutlined
                      class="afterSchoolTasksModel__student-group-arrow afterSchoolTasksModel__student-group-arrow--inline"
                      :class="{ 'is-collapsed': !expandedClassIds.includes(String(classItem.sourceId || '')) }"
                    />
                  </div>
                </div>

                <div v-show="expandedClassIds.includes(String(classItem.sourceId || ''))" class="afterSchoolTasksModel__student-list">
                  <div
                    v-for="student in classItem.students || []"
                    :key="student.studentId"
                    class="afterSchoolTasksModel__student-row"
                    @click="handleSelectClassStudent(classItem, student)"
                  >
                    <span class="afterSchoolTasksModel__student-radio" :class="{ 'is-selected': isDraftStudentSelected(student) }">
                      <CheckOutlined v-if="isDraftStudentSelected(student)" />
                    </span>
                    <span class="afterSchoolTasksModel__student-name">
                      {{ student.studentName }}
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
                :key="item.sourceId"
                class="afterSchoolTasksModel__student-group afterSchoolTasksModel__student-group--plain"
              >
                <div class="afterSchoolTasksModel__student-group-header afterSchoolTasksModel__student-group-header--plain">
                  <span
                    class="afterSchoolTasksModel__student-checkbox afterSchoolTasksModel__student-checkbox--button"
                    :class="{ 'is-selected': isSourceSelected('one_to_one', item.sourceId) }"
                    @click.stop="handleSelectSource('one_to_one', item)"
                  >
                    <CheckOutlined v-if="isSourceSelected('one_to_one', item.sourceId)" />
                  </span>
                  <div class="afterSchoolTasksModel__student-group-header-main" @click="toggleOneToOneExpanded(item.sourceId)">
                    <span class="afterSchoolTasksModel__student-group-title">
                      {{ item.sourceName || '1对1' }}（{{ isDraftStudentSelected(item.students[0]) ? 1 : 0 }}/1）
                    </span>
                    <CaretDownOutlined
                      class="afterSchoolTasksModel__student-group-arrow afterSchoolTasksModel__student-group-arrow--inline"
                      :class="{ 'is-collapsed': !expandedOneToOneIds.includes(String(item.sourceId || '')) }"
                    />
                  </div>
                </div>

                <div v-show="expandedOneToOneIds.includes(String(item.sourceId || ''))" class="afterSchoolTasksModel__student-list">
                  <div
                    class="afterSchoolTasksModel__student-row"
                    @click="handleSelectOneToOne(item)"
                  >
                    <span class="afterSchoolTasksModel__student-radio" :class="{ 'is-selected': isDraftStudentSelected(item.students[0]) }">
                      <CheckOutlined v-if="isDraftStudentSelected(item.students[0])" />
                    </span>
                    <span class="afterSchoolTasksModel__student-name">
                      {{ item.students[0]?.studentName || '-' }}
                    </span>
                    <template v-if="item.students[0]?.isBind === false">
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
