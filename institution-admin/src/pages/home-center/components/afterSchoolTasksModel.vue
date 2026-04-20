<script setup>
import { CaretDownOutlined, CheckOutlined, CloseOutlined, PictureOutlined, PlayCircleOutlined, QuestionCircleOutlined, SearchOutlined } from '@ant-design/icons-vue'
import dayjs from 'dayjs'
import { pageOneToOneSelectionApi } from '@/api/edu-center/one-to-one'
import { pageGroupClassSelectionApi } from '@/api/edu-center/group-class'
import messageService from '@/utils/messageService'

const props = defineProps({
  title: {
    type: String,
    default: '新建课后任务',
  },
})
const open = defineModel({
  type: Boolean,
  default: false,
})

const confirmLoading = ref(false)
const formRef = ref(null)
const studentPickerOpen = ref(false)
const studentPickerType = ref('class')
const studentPickerKeyword = ref('')
const studentPickerLoading = ref(false)
const classTargetList = ref([])
const oneToOneTargetList = ref([])
const expandedClassIds = ref([])
const expandedOneToOneIds = ref([])
const draftSelectedStudents = ref([])
const studentPickerRequestSeq = ref(0)
const formState = reactive({
  title: '',
  content: '',
  rule: 1,
  students: [],
  publishAt: undefined,
  deadlineAt: undefined,
  dateRange: [],
  time: undefined,
  weeks: [],
  imgList: [],
  videoList: [],
  audioList: [],
})

const activeFile = ref(undefined)
let studentPickerSearchTimer = undefined

const studentPickerTabs = [
  { key: 'class', label: '班级' },
  { key: 'one_to_one', label: '1对1' },
]

const studentPickerPlaceholder = computed(() => studentPickerType.value === 'class' ? '搜索班级名称' : '搜索1对1名称')
const selectedStudentButtonText = computed(() => formState.students.length > 0 ? `已选班级/学员(${formState.students.length})` : '选择班级/学员')

const weeks = [{ label: '星期一', value: 1 }, { label: '星期二', value: 2 }, { label: '星期三', value: 3 }, { label: '星期四', value: 4 }, { label: '星期五', value: 5 }, { label: '星期六', value: 6 }, { label: '星期日', value: 7 }]
const dateOptions = [{ label: '00:00', value: '00:00' }, { label: '01:00', value: '01:00' }, { label: '02:00', value: '02:00' }, { label: '03:00', value: '03:00' }, { label: '04:00', value: '04:00' }, { label: '05:00', value: '05:00' }, { label: '06:00', value: '06:00' }, { label: '07:00', value: '07:00' }, { label: '08:00', value: '08:00' }, { label: '09:00', value: '09:00' }, { label: '10:00', value: '10:00' }, { label: '11:00', value: '11:00' }, { label: '12:00', value: '12:00' }, { label: '13:00', value: '13:00' }, { label: '14:00', value: '14:00' }, { label: '15:00', value: '15:00' }, { label: '16:00', value: '16:00' }, { label: '17:00', value: '17:00' }, { label: '18:00', value: '18:00' }, { label: '19:00', value: '19:00' }, { label: '20:00', value: '20:00' }, { label: '21:00', value: '21:00' }, { label: '22:00', value: '22:00' }, { label: '23:00', value: '23:00' }]

function handleWeek(value) {
  const index = formState.weeks.indexOf(value)
  if (index === -1) {
    formState.weeks.push(value)
  }
  else {
    formState.weeks.splice(index, 1)
  }
}

// 图片预览
function handlePreview(file) {
  console.log(file)
}

// 鼠标悬停样式处理
function handleOpenChange(value, show) {
  if (show) {
    activeFile.value = value
  }
  else {
    activeFile.value = undefined
  }
}

function cloneSelectedStudents(list) {
  return Array.isArray(list)
    ? list.filter(item => String(item?.selectionType || 'student') === 'student').map(item => ({ ...item, selectionType: 'student' }))
    : []
}

function buildSelectedStudentKey(item) {
  return `${String(item?.sourceType || '')}:${String(item?.sourceId || '')}:${String(item?.studentId || '')}`
}

function isDraftStudentSelected(item) {
  const key = buildSelectedStudentKey(item)
  return draftSelectedStudents.value.some(selectedItem => buildSelectedStudentKey(selectedItem) === key)
}

function getSelectedCountBySource(sourceType, sourceId) {
  return draftSelectedStudents.value.filter(item =>
    String(item?.sourceType || '') === sourceType
    && String(item?.sourceId || '') === String(sourceId || '')).length
}

function buildClassStudentSelection(classItem, student) {
  return {
    sourceType: 'class',
    sourceId: String(classItem?.id || ''),
    sourceName: String(classItem?.name || ''),
    studentId: String(student?.id || ''),
    studentName: String(student?.name || ''),
    tuitionAccountId: String(student?.tuitionAccountId || ''),
    isBind: student?.isBind !== false,
  }
}

function buildOneToOneSelection(item) {
  return {
    sourceType: 'one_to_one',
    sourceId: String(item?.id || ''),
    sourceName: String(item?.name || item?.lessonName || item?.studentName || '1对1'),
    studentId: String(item?.studentId || item?.id || ''),
    studentName: String(item?.studentName || item?.name || ''),
    tuitionAccountId: String(item?.tuitionAccountId || ''),
    isBind: item?.isBindChild !== false,
  }
}

function isSourceSelected(sourceType, sourceId) {
  if (sourceType === 'one_to_one') {
    return getSelectedCountBySource(sourceType, sourceId) > 0
  }

  const targetClass = classTargetList.value.find(item => String(item?.id || '') === String(sourceId || ''))
  const total = Array.isArray(targetClass?.students) ? targetClass.students.length : 0
  return total > 0 && getSelectedCountBySource(sourceType, sourceId) === total
}

function toggleDraftStudent(item) {
  const key = buildSelectedStudentKey(item)
  const index = draftSelectedStudents.value.findIndex(selectedItem => buildSelectedStudentKey(selectedItem) === key)
  if (index >= 0) {
    draftSelectedStudents.value.splice(index, 1)
    return
  }
  draftSelectedStudents.value.push({ ...item })
}

function toggleClassExpanded(classId) {
  const normalizedId = String(classId || '')
  const index = expandedClassIds.value.indexOf(normalizedId)
  if (index >= 0) {
    expandedClassIds.value.splice(index, 1)
    return
  }
  expandedClassIds.value.push(normalizedId)
}

function toggleOneToOneExpanded(itemId) {
  const normalizedId = String(itemId || '')
  const index = expandedOneToOneIds.value.indexOf(normalizedId)
  if (index >= 0) {
    expandedOneToOneIds.value.splice(index, 1)
    return
  }
  expandedOneToOneIds.value.push(normalizedId)
}

function handleSelectClassStudent(classItem, student) {
  toggleDraftStudent(buildClassStudentSelection(classItem, student))
}

function handleSelectOneToOne(item) {
  toggleDraftStudent(buildOneToOneSelection(item))
}

function handleSelectSource(sourceType, item) {
  if (sourceType === 'one_to_one') {
    toggleDraftStudent(buildOneToOneSelection(item))
    return
  }

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

async function loadClassTargets(currentSeq) {
  const res = await pageGroupClassSelectionApi({
    queryModel: {
      className: String(studentPickerKeyword.value || '').trim() || undefined,
      status: [1],
    },
    pageRequestModel: {
      needTotal: true,
      pageSize: 50,
      pageIndex: 1,
      skipCount: 0,
    },
  })
  if (currentSeq !== studentPickerRequestSeq.value)
    return
  if (res.code !== 200)
    throw new Error(res.message || '获取班级列表失败')

  const classRows = Array.isArray(res.result?.list) ? res.result.list : []
  if (currentSeq !== studentPickerRequestSeq.value)
    return
  classTargetList.value = classRows
  expandedClassIds.value = classRows.filter(item => Array.isArray(item.students) && item.students.length > 0).slice(0, 1).map(item => String(item.id || ''))
}

async function loadOneToOneTargets(currentSeq) {
  const res = await pageOneToOneSelectionApi({
    queryModel: {
      searchKey: String(studentPickerKeyword.value || '').trim() || undefined,
      status: [1],
    },
    pageRequestModel: {
      needTotal: true,
      pageSize: 50,
      pageIndex: 1,
      skipCount: 0,
    },
  })
  if (currentSeq !== studentPickerRequestSeq.value)
    return
  if (res.code !== 200)
    throw new Error(res.message || '获取1对1列表失败')
  oneToOneTargetList.value = Array.isArray(res.result?.list) ? res.result.list : []
  expandedOneToOneIds.value = oneToOneTargetList.value.slice(0, 1).map(item => String(item?.id || ''))
}

async function loadStudentPickerData() {
  const currentSeq = ++studentPickerRequestSeq.value
  studentPickerLoading.value = true
  try {
    if (studentPickerType.value === 'class') {
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
  catch (error) {
    classTargetList.value = []
    oneToOneTargetList.value = []
    expandedOneToOneIds.value = []
    messageService.error(error?.response?.data?.message || error?.message || '加载班级/学员失败')
  }
  finally {
    if (currentSeq === studentPickerRequestSeq.value)
      studentPickerLoading.value = false
  }
}

function openStudentPicker() {
  studentPickerOpen.value = true
}

function closeStudentPicker() {
  studentPickerOpen.value = false
}

function handleCompleteStudentPicker() {
  formState.students = cloneSelectedStudents(draftSelectedStudents.value)
  closeStudentPicker()
}

function handleOk() {
  formRef.value.validate().then(() => {
    console.log('验证通过')
  })
}

function buildTimeRange(start, end) {
  return Array.from({ length: Math.max(end - start, 0) }, (_, index) => start + index)
}

function getPublishAtDate() {
  if (!formState.publishAt)
    return null
  const publishAt = dayjs(formState.publishAt)
  return publishAt.isValid() ? publishAt : null
}

// 禁用今天之前的日期
function disabledDate(current) {
  return current && current < dayjs().startOf('day')
}

function disabledDeadlineDate(current) {
  if (!current)
    return false

  const publishAt = getPublishAtDate()
  if (publishAt)
    return !dayjs(current).isAfter(publishAt, 'day')

  return dayjs(current).isBefore(dayjs().startOf('day'), 'day')
}

function disabledDeadlineTime(current) {
  const publishAt = getPublishAtDate()
  if (!publishAt || !current || !dayjs(current).isSame(publishAt, 'day'))
    return {}

  return {
    disabledHours: () => buildTimeRange(0, publishAt.hour()),
    disabledMinutes: selectedHour => selectedHour === publishAt.hour() ? buildTimeRange(0, publishAt.minute()) : [],
  }
}

watch(() => formState.publishAt, (publishAtValue) => {
  if (!publishAtValue || !formState.deadlineAt)
    return

  const publishAt = dayjs(publishAtValue)
  const deadlineAt = dayjs(formState.deadlineAt)
  if (publishAt.isValid() && deadlineAt.isValid() && !deadlineAt.isAfter(publishAt, 'day'))
    formState.deadlineAt = undefined
})

watch(() => studentPickerOpen.value, (open) => {
  if (!open) {
    studentPickerRequestSeq.value += 1
    studentPickerLoading.value = false
    studentPickerKeyword.value = ''
    classTargetList.value = []
    oneToOneTargetList.value = []
    expandedClassIds.value = []
    expandedOneToOneIds.value = []
    if (studentPickerSearchTimer) {
      clearTimeout(studentPickerSearchTimer)
      studentPickerSearchTimer = undefined
    }
    return
  }

  draftSelectedStudents.value = cloneSelectedStudents(formState.students)
  loadStudentPickerData()
})

watch(() => studentPickerType.value, () => {
  if (!studentPickerOpen.value)
    return
  studentPickerKeyword.value = ''
  loadStudentPickerData()
})

watch(() => studentPickerKeyword.value, () => {
  if (!studentPickerOpen.value)
    return
  if (studentPickerSearchTimer)
    clearTimeout(studentPickerSearchTimer)
  studentPickerSearchTimer = setTimeout(() => {
    loadStudentPickerData()
  }, 300)
})
</script>

<template>
  <div>
    <a-modal
      v-model:open="open"
      centered
      class="afterSchoolTasksModel"
      :body-style="{ height: '580px', overflowY: 'auto' }"
      width="800px"
      :title="props.title"
      destroy-on-close
      @ok="handleOk"
    >
      <a-form ref="formRef" layout="vertical" :model="formState" v-bind="formItemLayout">
        <a-form-item label="任务标题" name="title" :rules="[{ required: true, message: '请输入任务标题' }]">
          <a-input v-model:value="formState.title" :maxlength="20" placeholder="请输入任务标题，最多20字" />
        </a-form-item>
        <a-form-item
          label="任务内容"
          name="content"
          class="afterSchoolTasksModel__content-item"
          :rules="[{ required: true, message: '请输入任务内容' }]"
        >
          <a-textarea
            class="afterSchoolTasksModel__content-textarea"
            v-model:value="formState.content"
            :show-count="true"
            style="height: 66px; min-height: 66px;"
            :maxlength="2000"
            placeholder="请输入任务内容，最多2000字"
            :auto-size="{ minRows: 4, maxRows: 4 }"
          />
        </a-form-item>

        <a-form-item class="afterSchoolTasksModel__upload-form-item">
          <div class="afterSchoolTasksModel__upload-actions flex flex-wrap items-center gap-8px">
            <a-upload
              v-model:file-list="formState.imgList"
              class="afterSchoolTasksModel__upload"
              list-type="picture-card"
              :max-count="12"
              action="https://www.mocky.io/v2/5cc8019d300000980a055e76"
            >
              <a-tooltip placement="right" @open-change="(show) => handleOpenChange(1, show)">
                <template #title>
                  限制单张 9 M
                </template>
                <div
                  :class="{ 'bg-#06f!important': activeFile === 1, 'text-#fff!important': activeFile === 1 }"
                  class="w-135px cursor-pointer flex items-center gap-5px bg-#f6f7f8 px-10px py-3px rounded-12px"
                >
                  <PictureOutlined :class="activeFile === 1 ? 'text-#fff' : 'text-#06f'" />
                  <span>添加图片({{ formState.imgList.length }}/12)</span>
                </div>
              </a-tooltip>
            </a-upload>

            <a-upload
              v-model:file-list="formState.videoList"
              class="afterSchoolTasksModel__upload"
              list-type="picture-card"
              :max-count="9"
              action="https://www.mocky.io/v2/5cc8019d300000980a055e76"
            >
              <a-tooltip placement="right" @open-change="(show) => handleOpenChange(2, show)">
                <template #title>
                  限制每个 500 M
                </template>
                <div
                  :class="{ 'bg-#06f!important': activeFile === 2, 'text-#fff!important': activeFile === 2 }"
                  class="w-135px cursor-pointer flex items-center gap-5px bg-#f6f7f8 px-10px py-3px rounded-12px"
                >
                  <PlayCircleOutlined :class="activeFile === 2 ? 'text-#fff' : 'text-#06f'" />
                  <span>添加视频({{ formState.videoList.length }}/9)</span>
                </div>
              </a-tooltip>
            </a-upload>
            <!-- <a-upload
              v-model:file-list="formState.audioList" list-type="picture-card" :max-count="12"
              action="https://www.mocky.io/v2/5cc8019d300000980a055e76"
            >
              <a-tooltip placement="right" @open-change="(show) => handleOpenChange(3, show)">
                <template #title>
                  限制每个 10 M
                </template>
                <div
                  :class="{ 'bg-#06f!important': activeFile === 3, 'text-#fff!important': activeFile === 3 }"
                  class="w-135px cursor-pointer flex items-center gap-5px bg-#f6f7f8 px-10px py-3px rounded-12px"
                >
                  <AudioOutlined :class="activeFile === 3 ? 'text-#fff' : 'text-#06f'" />
                  <span>添加音频(0/10)</span>
                </div>
              </a-tooltip>
            </a-upload> -->
          </div>
        </a-form-item>

        <a-form-item label="选择班级/学员" name="students" :rules="[{ required: true, message: '请选择班级/学员' }]">
          <a-button type="primary" ghost @click="openStudentPicker">
            {{ selectedStudentButtonText }}
          </a-button>
        </a-form-item>
        <a-form-item label="发布规则" :required="true">
          <a-radio-group v-model:value="formState.rule" class="custom-radio">
            <a-radio :value="1">
              仅本次发布
            </a-radio>
            <a-radio :value="2">
              设置自动任务
            </a-radio>
          </a-radio-group>
        </a-form-item>
        <a-form-item v-if="formState.rule === 1">
          <div class="afterSchoolTasksModel__rule-card">
            <div class="afterSchoolTasksModel__rule-card-title">
              设置本次发布时间（非必填）
            </div>
            <a-row :gutter="[20, 16]">
              <a-col :xs="24" :sm="12">
                <div class="afterSchoolTasksModel__rule-field">
                  <div class="afterSchoolTasksModel__rule-label">
                    <span>定时发布日期</span>
                    <a-popover
                      color="#fff"
                      placement="topLeft"
                      title="定时发布日期"
                    >
                      <template #content>
                        <div class="afterSchoolTasksModel__rule-popover">
                          设置后，任务创建完成，会按设置的时间点发送，如果不设置，任务创建完成会立即发送
                        </div>
                      </template>
                      <QuestionCircleOutlined class="afterSchoolTasksModel__rule-tip" />
                    </a-popover>
                    <span>:</span>
                  </div>
                  <a-date-picker
                    v-model:value="formState.publishAt"
                    class="w-full"
                    :show-time="{ format: 'HH:mm' }"
                    value-format="YYYY-MM-DD HH:mm"
                    format="YYYY-MM-DD HH:mm"
                    placeholder="请选择日期时间"
                    :disabled-date="disabledDate"
                  />
                </div>
              </a-col>
              <a-col :xs="24" :sm="12">
                <div class="afterSchoolTasksModel__rule-field">
                  <div class="afterSchoolTasksModel__rule-label">
                    <span>设置任务截止日期</span>
                    <a-popover
                      color="#fff"
                      placement="topLeft"
                      title="任务截止日期"
                    >
                      <template #content>
                        <div class="afterSchoolTasksModel__rule-popover">
                          设置后，学员仍可以上传任务，超时提交的学员，在系统上为学员打上超时提交标签
                        </div>
                      </template>
                      <QuestionCircleOutlined class="afterSchoolTasksModel__rule-tip" />
                    </a-popover>
                    <span>:</span>
                  </div>
                  <a-date-picker
                    v-model:value="formState.deadlineAt"
                    class="w-full"
                    :show-time="{ format: 'HH:mm' }"
                    value-format="YYYY-MM-DD HH:mm"
                    format="YYYY-MM-DD HH:mm"
                    placeholder="请选择日期时间"
                    :disabled-date="disabledDeadlineDate"
                    :disabled-time="disabledDeadlineTime"
                  />
                </div>
              </a-col>
            </a-row>
          </div>
        </a-form-item>
        <a-form-item v-if="formState.rule === 2">
          <div class="border border-gray-200 rounded-8px border-solid">
            <div class="bg-#fafafa  px-15px py-10px">
              设置自动任务周期
            </div>
            <div class="flex items-center gap-30px p-15px">
              <div v-for="(week, index) in weeks" :key="index" class="flex flex-col items-center gap-5px">
                <div class="text-#888 text-12px">
                  {{ week.label }}
                </div>
                <div
                  class="week-day" :class="{ 'week-active': formState.weeks.includes(week.value) }"
                  @click="handleWeek(week.value)"
                />
              </div>
            </div>
            <div class="flex items-center  justify-between gap-10px p-15px">
              <a-form-item label="任务日期范围" name="dateRange" :rules="[{ required: true, message: '请选择周期' }]">
                <a-range-picker v-model:value="formState.dateRange" :disabled-date="disabledDate" />
              </a-form-item>
              <a-form-item
                class=" w-340px" label="任务推送时间：" name="time"
                :rules="[{ required: true, message: '请选择任务推送时间' }]"
              >
                <a-select v-model:value="formState.time" placeholder="请选择" :options="dateOptions" />
              </a-form-item>
            </div>
          </div>
        </a-form-item>
      </a-form>
    </a-modal>

    <a-modal
      v-model:open="studentPickerOpen"
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
          <span>选择班级/学员</span>
          <a-button type="text" class="afterSchoolTasksModel__student-picker-close" @click="closeStudentPicker">
            <template #icon>
              <CloseOutlined />
            </template>
          </a-button>
        </div>
      </template>

      <div class="afterSchoolTasksModel__student-picker">
        <div class="afterSchoolTasksModel__student-picker-sidebar">
          <div
            v-for="item in studentPickerTabs"
            :key="item.key"
            :class="{ 'is-active': studentPickerType === item.key }"
            class="afterSchoolTasksModel__student-picker-tab"
            @click="studentPickerType = item.key"
          >
            {{ item.label }}
          </div>
        </div>

        <div class="afterSchoolTasksModel__student-picker-main">
          <div class="afterSchoolTasksModel__student-picker-toolbar">
            <a-input
              v-model:value="studentPickerKeyword"
              :placeholder="studentPickerPlaceholder"
              allow-clear
            >
              <template #prefix>
                <SearchOutlined />
              </template>
            </a-input>
          </div>

          <a-spin :spinning="studentPickerLoading" class="afterSchoolTasksModel__student-picker-spin">
            <div class="afterSchoolTasksModel__student-picker-content">
              <template v-if="studentPickerType === 'class'">
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
          <a-button @click="closeStudentPicker">
            关闭
          </a-button>
          <a-button type="primary" @click="handleCompleteStudentPicker">
            完成
          </a-button>
        </div>
      </template>
    </a-modal>
  </div>
</template>

<style>
.afterSchoolTasksModel {
  padding-bottom: 0;
  text-align: left;
}
</style>

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

.afterSchoolTasksModel__upload-form-item {
  margin-top: -14px;
  margin-bottom: 12px;
}

.afterSchoolTasksModel__content-item {
  margin-bottom: 10px;
}

.afterSchoolTasksModel__upload-actions {
  align-items: flex-start;
}

.afterSchoolTasksModel__upload {
  display: inline-flex;
  width: auto;
  flex: none;
  margin-bottom: 0;
}

.afterSchoolTasksModel__rule-card {
  border: 1px solid #e8e8e8;
  border-radius: 10px;
  background: #fff;
  padding: 16px;
}

.afterSchoolTasksModel__rule-card-title {
  margin-bottom: 18px;
  color: #262626;
  font-size: 14px;
  line-height: 22px;
  font-weight: 500;
}

.afterSchoolTasksModel__rule-field {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.afterSchoolTasksModel__rule-label {
  display: inline-flex;
  align-items: center;
  gap: 2px;
  color: #262626;
  font-size: 14px;
  line-height: 22px;
}

.afterSchoolTasksModel__rule-tip {
  color: #999;
  font-size: 14px;
  cursor: pointer;
}

.afterSchoolTasksModel__rule-popover {
  max-width: 360px;
  color: #595959;
  font-size: 14px;
  line-height: 22px;
}

:deep(.afterSchoolTasksModel__content-item .ant-input-textarea-show-count::after) {
  margin-top: 2px;
}

:deep(.afterSchoolTasksModel__content-item .ant-form-item-control-input + div) {
  min-height: auto;
  margin-top: -24px;
  margin-bottom: 5px;
}

:deep(.afterSchoolTasksModel__content-textarea.ant-input-textarea-show-count::after) {
  margin-top: 2px;
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

.week-day {
  width: 48px;
  height: 48px;
  border-radius: 8px;
  background-color: #eee;
  background-image: url('https://pcsys.admin.ybc365.com/64344ed6-b8db-43a2-8488-4c18a6095a50.png');
  background-repeat: no-repeat;
  background-position: center;
  background-size: 24px;
  cursor: pointer;
}

.week-active {
  background-color: #06f;
}

::v-deep(.ant-upload-select) {
  border: none !important;
  flex: 1;
  width: 135px !important;
  height: 100% !important;
  display: block;
}

::v-deep(.afterSchoolTasksModel__upload .ant-upload-list) {
  display: inline-flex;
  align-items: flex-start;
  flex-wrap: wrap;
}

::v-deep(.ant-upload-list-item-container) {
  width: 80px !important;
  height: 80px !important;
}
</style>
