<script setup lang="ts">
import { CaretDownOutlined, CheckOutlined, CloseOutlined, LoadingOutlined, PictureOutlined, PlayCircleOutlined, QuestionCircleOutlined, SearchOutlined } from '@ant-design/icons-vue'
import * as qiniu from 'qiniu-js'
import { Modal } from 'ant-design-vue'
import dayjs from 'dayjs'
import { computed, h, nextTick, reactive, ref, watch } from 'vue'
import type { RehabRecordMediaItem } from '@/api/edu-center/class-record'
import { pageGroupClassSelectionApi } from '@/api/edu-center/group-class'
import { pageOneToOneSelectionApi } from '@/api/edu-center/one-to-one'
import {
  batchCreateHomeworksApi,
  getHomeworkDetailApi,
  updateHomeworkApi,
  type HomeworkAttachmentItem,
  type HomeworkDetail,
  type HomeworkStudentSelection,
} from '@/api/home-center/homework'
import { getQiniuToken, getVideoUploadToken } from '@/api/qiniu'
import messageService from '@/utils/messageService'

interface StudentPickerSelection {
  sourceType: 'class' | 'one_to_one'
  sourceId: string
  sourceName: string
  studentId: string
  studentName: string
  tuitionAccountId?: string
  isBind: boolean
  selectionType?: 'student' | 'source'
}

const props = withDefaults(defineProps<{
  mode?: 'create' | 'edit'
  homeworkId?: string
}>(), {
  mode: 'create',
  homeworkId: '',
})

const emit = defineEmits<{
  (e: 'success'): void
}>()

const open = defineModel<boolean>({ default: false })

const formRef = ref()
const confirmLoading = ref(false)
const detailLoading = ref(false)
const studentPickerOpen = ref(false)
const studentPickerType = ref<'class' | 'one_to_one'>('class')
const studentPickerKeyword = ref('')
const studentPickerLoading = ref(false)
const classTargetList = ref<any[]>([])
const oneToOneTargetList = ref<any[]>([])
const expandedClassIds = ref<string[]>([])
const expandedOneToOneIds = ref<string[]>([])
const draftSelectedStudents = ref<StudentPickerSelection[]>([])
const studentPickerRequestSeq = ref(0)
const initialSnapshot = ref('')
const imageUploading = ref(false)
const videoUploading = ref(false)
const imageInputRef = ref<HTMLInputElement | null>(null)
const videoInputRef = ref<HTMLInputElement | null>(null)
const activeUploadAction = ref<number | undefined>()
const previewImageOpen = ref(false)
const previewImageSrc = ref('')
const previewImageTitle = ref('')
const previewVideoOpen = ref(false)
const previewVideoItem = ref<RehabRecordMediaItem | null>(null)
const previewVideoRef = ref<HTMLVideoElement | null>(null)

let studentPickerSearchTimer: ReturnType<typeof setTimeout> | undefined

const IMAGE_LIMIT = 12
const VIDEO_LIMIT = 9
const IMAGE_MAX_SIZE_MB = 10
const VIDEO_MAX_SIZE_MB = 100

interface InvalidSelectionItem {
  name: string
  reason: string
}

const dateOptions = Array.from({ length: 24 }, (_, index) => {
  const text = String(index).padStart(2, '0')
  return { label: `${text}:00`, value: `${text}:00` }
})

const weeks = [
  { label: '周一', value: 1 },
  { label: '周二', value: 2 },
  { label: '周三', value: 3 },
  { label: '周四', value: 4 },
  { label: '周五', value: 5 },
  { label: '周六', value: 6 },
  { label: '周日', value: 7 },
]

function createDefaultFormState() {
  return {
    title: '',
    content: '',
    rule: 1,
    students: [] as StudentPickerSelection[],
    publishAt: undefined as string | undefined,
    deadlineAt: undefined as string | undefined,
    dateRange: undefined as [string, string] | undefined,
    time: undefined as string | undefined,
    taskDurationHours: undefined as number | undefined,
    weeks: [] as number[],
    mediaList: [] as RehabRecordMediaItem[],
  }
}

const formState = reactive(createDefaultFormState())

const isEditMode = computed(() => props.mode === 'edit')
const dialogTitle = computed(() => props.mode === 'edit' ? '编辑课后任务' : '新建课后任务')
const selectedStudentButtonText = computed(() => formState.students.length > 0 ? `已选班级/学员（${formState.students.length}）` : '选择班级/学员')
const selectedStudentPreviewText = computed(() =>
  Array.from(new Set(formState.students.map(item => String(item.studentName || '').trim()).filter(Boolean))).join('、'),
)
const studentPickerTabs = [
  { key: 'class', label: '班级' },
  { key: 'one_to_one', label: '1对1' },
]
const studentPickerPlaceholder = computed(() => studentPickerType.value === 'class' ? '搜索班级名称' : '搜索1对1名称')
const isDirty = computed(() => initialSnapshot.value !== serializeFormState())

function resetFormState() {
  Object.assign(formState, createDefaultFormState())
  formState.students = []
  formState.dateRange = undefined
  formState.weeks = []
  formState.mediaList = []
}

function serializeFormState() {
  return JSON.stringify({
    title: String(formState.title || '').trim(),
    content: String(formState.content || '').trim(),
    rule: Number(formState.rule || 1),
    publishAt: formState.publishAt || '',
    deadlineAt: formState.deadlineAt || '',
    dateRange: formState.dateRange ? [...formState.dateRange] : [],
    time: formState.time || '',
    taskDurationHours: Number(formState.taskDurationHours || 0) || '',
    weeks: [...formState.weeks].sort((a, b) => a - b),
    students: [...formState.students]
      .map(item => ({
        sourceType: item.sourceType,
        sourceId: item.sourceId,
        studentId: item.studentId,
        tuitionAccountId: item.tuitionAccountId || '',
      }))
      .sort((a, b) => `${a.sourceType}:${a.sourceId}:${a.studentId}`.localeCompare(`${b.sourceType}:${b.sourceId}:${b.studentId}`)),
    mediaList: [...formState.mediaList]
      .map(item => ({
        mediaType: item.mediaType || '',
        url: item.url || '',
        fileName: item.fileName || '',
      }))
      .sort((a, b) => `${a.mediaType}:${a.url}`.localeCompare(`${b.mediaType}:${b.url}`)),
  })
}

function captureInitialSnapshot() {
  initialSnapshot.value = serializeFormState()
}

function resetStudentPickerState() {
  studentPickerOpen.value = false
  studentPickerKeyword.value = ''
  classTargetList.value = []
  oneToOneTargetList.value = []
  expandedClassIds.value = []
  expandedOneToOneIds.value = []
  studentPickerLoading.value = false
  studentPickerRequestSeq.value += 1
  if (studentPickerSearchTimer) {
    clearTimeout(studentPickerSearchTimer)
    studentPickerSearchTimer = undefined
  }
}

async function initializeForm() {
  resetFormState()
  resetStudentPickerState()
  await nextTick()
  formRef.value?.clearValidate?.()

  if (props.mode === 'edit' && props.homeworkId) {
    detailLoading.value = true
    try {
      const res = await getHomeworkDetailApi(props.homeworkId)
      if (res.code !== 200 || !res.result) {
        messageService.error(res.message || '获取课后任务详情失败')
        open.value = false
        return
      }
      applyDetailToForm(res.result)
    }
    catch (error) {
      console.error('get homework detail failed', error)
      messageService.error('获取课后任务详情失败')
      open.value = false
      return
    }
    finally {
      detailLoading.value = false
    }
  }

  await nextTick()
  captureInitialSnapshot()
}

function applyDetailToForm(detail: HomeworkDetail) {
  formState.title = detail.title || ''
  formState.content = detail.content || ''
  formState.students = (Array.isArray(detail.selectedStudents) ? detail.selectedStudents : []).map(item => transformSelectionFromDetail(item))
  formState.mediaList = homeworkAttachmentsToMedia(detail.attachments || [])
  formState.rule = detail.publishRule === 2 ? 2 : 1

  if (formState.rule === 2 && detail.repeatRule) {
    formState.dateRange = [detail.repeatRule.startDate, detail.repeatRule.endDate]
    formState.weeks = bitmaskToWeeks(detail.repeatRule.weekDays)
    formState.time = hourToTimeOption(detail.publishHour)
    formState.taskDurationHours = normalizeTaskDurationHours(detail.taskDurationHours) ?? resolveTaskDurationHours(detail.publishHour, detail.endHour)
    return
  }

  formState.publishAt = detail.publishTime ? dayjs(detail.publishTime).format('YYYY-MM-DD HH:mm') : undefined
  formState.deadlineAt = detail.endTime ? dayjs(detail.endTime).format('YYYY-MM-DD HH:mm') : undefined
}

function transformSelectionFromDetail(item: HomeworkStudentSelection): StudentPickerSelection {
  return {
    sourceType: Number(item.sourceType) === 2 ? 'one_to_one' : 'class',
    sourceId: String(item.sourceId || ''),
    sourceName: String(item.sourceName || ''),
    studentId: String(item.studentId || ''),
    studentName: String(item.studentName || ''),
    tuitionAccountId: String(item.tuitionAccountId || ''),
    isBind: item.isBind !== false,
    selectionType: 'student',
  }
}

function normalizeTextValue(value?: string | number | null) {
  return String(value ?? '').trim()
}

function matchesFileByExtension(fileName: string, patterns: RegExp[]) {
  return patterns.some(pattern => pattern.test(fileName))
}

function isSupportedImageFile(file: File) {
  const fileType = normalizeTextValue(file.type).toLowerCase()
  return fileType.startsWith('image/')
    || matchesFileByExtension(file.name || '', [/\.(jpg|jpeg|png|bmp|webp|gif)$/i])
}

function isSupportedVideoFile(file: File) {
  const fileType = normalizeTextValue(file.type).toLowerCase()
  return fileType.startsWith('video/')
    || matchesFileByExtension(file.name || '', [/\.(mp4|mov|webm|ogg|m4v)$/i])
}

const imageCount = computed(() => formState.mediaList.filter(item => item.mediaType === 'image').length)
const videoCount = computed(() => formState.mediaList.filter(item => item.mediaType === 'video').length)

function handleUploadActionHover(type: number, openStatus: boolean) {
  activeUploadAction.value = openStatus ? type : undefined
}

function openImagePicker() {
  if (imageUploading.value || imageCount.value >= IMAGE_LIMIT)
    return
  imageInputRef.value?.click()
}

function openVideoPicker() {
  if (videoUploading.value || videoCount.value >= VIDEO_LIMIT)
    return
  videoInputRef.value?.click()
}

function buildSelectionWarningContent(invalidItems: InvalidSelectionItem[], prefixText: string) {
  return h('div', { class: 'media-selection-dialog' }, [
    h('div', { class: 'media-selection-dialog__intro' }, prefixText),
    ...invalidItems.map(item => h('div', { class: 'media-selection-dialog__item' }, `${item.name}：${item.reason}`)),
  ])
}

function validateImageFiles(files: File[]) {
  const invalidItems: InvalidSelectionItem[] = []
  const acceptedFiles: File[] = []

  files.forEach((file) => {
    if (!isSupportedImageFile(file)) {
      invalidItems.push({ name: file.name, reason: '文件格式不支持' })
      return
    }

    if (file.size / 1024 / 1024 > IMAGE_MAX_SIZE_MB) {
      invalidItems.push({ name: file.name, reason: `图片需小于等于 ${IMAGE_MAX_SIZE_MB}MB` })
      return
    }

    acceptedFiles.push(file)
  })

  const remainingCount = Math.max(IMAGE_LIMIT - imageCount.value, 0)
  const validFiles = acceptedFiles.slice(0, remainingCount)
  if (acceptedFiles.length > remainingCount) {
    acceptedFiles.slice(remainingCount).forEach((file) => {
      invalidItems.push({ name: file.name, reason: `最多还能上传 ${remainingCount} 张图片` })
    })
  }

  return { validFiles, invalidItems }
}

function validateVideoFiles(files: File[]) {
  const invalidItems: InvalidSelectionItem[] = []
  const acceptedFiles: File[] = []

  files.forEach((file) => {
    if (!isSupportedVideoFile(file)) {
      invalidItems.push({ name: file.name, reason: '文件格式不支持' })
      return
    }

    if (file.size / 1024 / 1024 > VIDEO_MAX_SIZE_MB) {
      invalidItems.push({ name: file.name, reason: `视频需小于等于 ${VIDEO_MAX_SIZE_MB}MB` })
      return
    }

    acceptedFiles.push(file)
  })

  const remainingCount = Math.max(VIDEO_LIMIT - videoCount.value, 0)
  const validFiles = acceptedFiles.slice(0, remainingCount)
  if (acceptedFiles.length > remainingCount) {
    acceptedFiles.slice(remainingCount).forEach((file) => {
      invalidItems.push({ name: file.name, reason: `最多还能上传 ${remainingCount} 个视频` })
    })
  }

  return { validFiles, invalidItems }
}

function resetFileInput(input?: HTMLInputElement | null) {
  if (input)
    input.value = ''
}

function appendMediaItems(items: RehabRecordMediaItem[]) {
  if (items.length === 0)
    return
  formState.mediaList = [...formState.mediaList, ...items]
}

function uploadSingleImage(file: File) {
  return new Promise<RehabRecordMediaItem>(async (resolve, reject) => {
    try {
      const tokenRes: any = await getQiniuToken()
      const { token, uuid, buckethostname } = tokenRes.result || {}
      if (!token || !uuid || !buckethostname)
        throw new Error('图片上传凭证缺失')

      const ext = file.name?.includes('.')
        ? file.name.slice(file.name.lastIndexOf('.'))
        : (file.type === 'image/png' ? '.png' : '.jpg')
      const key = `rehab-record/${uuid}${ext}`

      const observable = qiniu.upload(file, key, token, {
        fname: file.name,
        mimeType: file.type,
      }, {
        useCdnDomain: true,
        region: qiniu.region.z0,
      })

      observable.subscribe({
        error(err) {
          reject(err)
        },
        complete(res) {
          resolve({
            mediaType: 'image',
            url: `${buckethostname}${res.key}`,
            fileName: normalizeTextValue(file.name),
            size: Number(file.size || 0) || undefined,
          })
        },
      })
    }
    catch (error) {
      reject(error)
    }
  })
}

function uploadSingleVideo(file: File) {
  return new Promise<RehabRecordMediaItem>(async (resolve, reject) => {
    try {
      const tokenRes: any = await getVideoUploadToken()
      const { token, uuid, buckethostname } = tokenRes.result || {}
      const key = normalizeTextValue(uuid)
      if (!token || !key || !buckethostname)
        throw new Error('视频上传凭证缺失')

      const observable = qiniu.upload(file, key, token, {
        fname: file.name,
        mimeType: file.type,
      }, {
        useCdnDomain: true,
        region: qiniu.region.z0,
      })

      observable.subscribe({
        error(err) {
          reject(err)
        },
        complete(res) {
          resolve({
            mediaType: 'video',
            url: `${buckethostname}${res.key}`,
            fileName: normalizeTextValue(file.name) || '视频附件',
            size: Number(file.size || 0) || undefined,
          })
        },
      })
    }
    catch (error) {
      reject(error)
    }
  })
}

async function uploadImageFiles(files: File[]) {
  if (files.length === 0)
    return

  imageUploading.value = true
  try {
    const results = await Promise.allSettled(files.map(file => uploadSingleImage(file)))
    const successItems = results
      .filter(result => result.status === 'fulfilled')
      .map(result => result.value)
    const failedCount = results.length - successItems.length

    appendMediaItems(successItems)
    if (successItems.length > 0)
      messageService.success(`已上传 ${successItems.length} 张图片`)
    if (failedCount > 0)
      messageService.error(`${failedCount} 张图片上传失败，请稍后重试`)
  }
  finally {
    imageUploading.value = false
  }
}

async function uploadVideoFiles(files: File[]) {
  if (files.length === 0)
    return

  videoUploading.value = true
  try {
    const results = await Promise.allSettled(files.map(file => uploadSingleVideo(file)))
    const successItems = results
      .filter(result => result.status === 'fulfilled')
      .map(result => result.value)
    const failedCount = results.length - successItems.length

    appendMediaItems(successItems)
    if (successItems.length > 0)
      messageService.success(`已上传 ${successItems.length} 个视频`)
    if (failedCount > 0)
      messageService.error(`${failedCount} 个视频上传失败，请稍后重试`)
  }
  finally {
    videoUploading.value = false
  }
}

function handleImageFileChange(event: Event) {
  const input = event.target as HTMLInputElement
  const files = Array.from(input.files || [])
  resetFileInput(input)
  if (files.length === 0)
    return

  const { validFiles, invalidItems } = validateImageFiles(files)
  if (invalidItems.length === 0) {
    void uploadImageFiles(validFiles)
    return
  }

  if (validFiles.length === 0) {
    Modal.warning({
      title: '图片无法上传',
      content: buildSelectionWarningContent(invalidItems, '以下图片不支持上传：'),
      okText: '知道了',
    })
    return
  }

  Modal.confirm({
    title: '部分图片无法上传',
    content: buildSelectionWarningContent(invalidItems, '以下图片不支持上传：'),
    okText: `继续上传剩余 ${validFiles.length} 张`,
    cancelText: '返回修改',
    onOk: async () => {
      await uploadImageFiles(validFiles)
    },
  })
}

function handleVideoFileChange(event: Event) {
  const input = event.target as HTMLInputElement
  const files = Array.from(input.files || [])
  resetFileInput(input)
  if (files.length === 0)
    return

  const { validFiles, invalidItems } = validateVideoFiles(files)
  if (invalidItems.length === 0) {
    void uploadVideoFiles(validFiles)
    return
  }

  if (validFiles.length === 0) {
    Modal.warning({
      title: '视频无法上传',
      content: buildSelectionWarningContent(invalidItems, '以下视频不支持上传：'),
      okText: '知道了',
    })
    return
  }

  Modal.confirm({
    title: '部分视频无法上传',
    content: buildSelectionWarningContent(invalidItems, '以下视频不支持上传：'),
    okText: `继续上传剩余 ${validFiles.length} 个`,
    cancelText: '返回修改',
    onOk: async () => {
      await uploadVideoFiles(validFiles)
    },
  })
}

function resolveMediaLabel(item: RehabRecordMediaItem, index: number) {
  const fileName = normalizeTextValue(item.fileName)
  if (fileName)
    return fileName
  return item.mediaType === 'video' ? `视频${index + 1}` : `图片${index + 1}`
}

function handleMediaPreview(item: RehabRecordMediaItem, index = 0) {
  if (item.mediaType === 'video') {
    previewVideoItem.value = item
    previewVideoOpen.value = true
    return
  }

  previewImageSrc.value = normalizeTextValue(item.url)
  previewImageTitle.value = resolveMediaLabel(item, index)
  previewImageOpen.value = true
}

function handleRemoveMedia(item: RehabRecordMediaItem) {
  formState.mediaList = formState.mediaList.filter(target => target.url !== item.url)
}

function closeVideoPreview() {
  const currentVideo = previewVideoRef.value
  if (currentVideo) {
    currentVideo.pause()
    currentVideo.currentTime = 0
    currentVideo.removeAttribute('src')
    currentVideo.load()
  }

  previewVideoOpen.value = false
  previewVideoItem.value = null
}

function resetMediaOverlayState() {
  activeUploadAction.value = undefined
  previewImageOpen.value = false
  previewImageSrc.value = ''
  previewImageTitle.value = ''
  closeVideoPreview()
  resetFileInput(imageInputRef.value)
  resetFileInput(videoInputRef.value)
}

function cloneSelectedStudents(list: StudentPickerSelection[]) {
  return Array.isArray(list)
    ? list.filter(item => item.selectionType !== 'source').map(item => ({ ...item, selectionType: 'student' as const }))
    : []
}

function buildSelectedStudentKey(item: Partial<StudentPickerSelection>) {
  return `${String(item.sourceType || '')}:${String(item.sourceId || '')}:${String(item.studentId || '')}`
}

function isDraftStudentSelected(item: StudentPickerSelection) {
  const key = buildSelectedStudentKey(item)
  return draftSelectedStudents.value.some(selectedItem => buildSelectedStudentKey(selectedItem) === key)
}

function getSelectedCountBySource(sourceType: 'class' | 'one_to_one', sourceId: string) {
  return draftSelectedStudents.value.filter(item => item.sourceType === sourceType && String(item.sourceId || '') === String(sourceId || '')).length
}

function buildClassStudentSelection(classItem: any, student: any): StudentPickerSelection {
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

function buildOneToOneSelection(item: any): StudentPickerSelection {
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

function isSourceSelected(sourceType: 'class' | 'one_to_one', sourceId: string) {
  if (sourceType === 'one_to_one')
    return getSelectedCountBySource(sourceType, sourceId) > 0

  const currentClass = classTargetList.value.find(item => String(item?.id || '') === String(sourceId || ''))
  const total = Array.isArray(currentClass?.students) ? currentClass.students.length : 0
  return total > 0 && getSelectedCountBySource(sourceType, sourceId) === total
}

function toggleDraftStudent(item: StudentPickerSelection) {
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
  toggleDraftStudent(buildClassStudentSelection(classItem, student))
}

function handleSelectOneToOne(item: any) {
  toggleDraftStudent(buildOneToOneSelection(item))
}

function handleSelectSource(sourceType: 'class' | 'one_to_one', item: any) {
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

async function loadClassTargets(currentSeq: number) {
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
  classTargetList.value = Array.isArray(res.result?.list) ? res.result.list : []
  expandedClassIds.value = classTargetList.value.filter(item => Array.isArray(item.students) && item.students.length > 0).slice(0, 1).map(item => String(item.id || ''))
}

async function loadOneToOneTargets(currentSeq: number) {
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
  catch (error: any) {
    classTargetList.value = []
    oneToOneTargetList.value = []
    expandedClassIds.value = []
    expandedOneToOneIds.value = []
    messageService.error(error?.response?.data?.message || error?.message || '加载班级/学员失败')
  }
  finally {
    if (currentSeq === studentPickerRequestSeq.value)
      studentPickerLoading.value = false
  }
}

function openStudentPicker() {
  draftSelectedStudents.value = cloneSelectedStudents(formState.students)
  studentPickerOpen.value = true
}

function closeStudentPicker() {
  studentPickerOpen.value = false
}

async function handleCompleteStudentPicker() {
  formState.students = cloneSelectedStudents(draftSelectedStudents.value)
  if (formState.students.length > 0) {
    await nextTick()
    formRef.value?.clearValidate?.(['students'])
  }
  closeStudentPicker()
}

function disabledDate(current: dayjs.Dayjs) {
  return current && current < dayjs().startOf('day')
}

function getPublishAtDate() {
  if (!formState.publishAt)
    return null
  const current = dayjs(formState.publishAt)
  return current.isValid() ? current : null
}

function buildTimeRange(start: number, end: number) {
  return Array.from({ length: Math.max(end - start, 0) }, (_, index) => start + index)
}

function disabledDeadlineDate(current: dayjs.Dayjs) {
  if (!current)
    return false
  const publishAt = getPublishAtDate()
  if (publishAt)
    return !dayjs(current).isAfter(publishAt, 'day')
  return dayjs(current).isBefore(dayjs().startOf('day'), 'day')
}

function disabledDeadlineTime(current: dayjs.Dayjs) {
  const publishAt = getPublishAtDate()
  if (!publishAt || !current || !dayjs(current).isSame(publishAt, 'day'))
    return {}
  return {
    disabledHours: () => buildTimeRange(0, publishAt.hour()),
    disabledMinutes: (selectedHour: number) => selectedHour === publishAt.hour() ? buildTimeRange(0, publishAt.minute()) : [],
  }
}

function toggleWeek(value: number) {
  const index = formState.weeks.indexOf(value)
  if (index === -1)
    formState.weeks.push(value)
  else
    formState.weeks.splice(index, 1)
}

function timeOptionToHour(value?: string) {
  if (!value)
    return undefined
  const hour = Number(String(value).split(':')[0])
  return Number.isFinite(hour) ? hour : undefined
}

function hourToTimeOption(value?: number) {
  if (value === undefined || value === null || Number.isNaN(Number(value)))
    return undefined
  return `${String(value).padStart(2, '0')}:00`
}

function normalizeTaskDurationHours(value?: string | number | null) {
  if (value === undefined || value === null || String(value).trim() === '')
    return undefined
  const current = Number(value)
  if (!Number.isFinite(current))
    return undefined
  return current
}

function resolveTaskDurationHours(publishHour?: number, endHour?: number) {
  const startHour = Number(publishHour)
  const finishHour = Number(endHour)
  if (!Number.isFinite(startHour) || !Number.isFinite(finishHour))
    return undefined
  const duration = finishHour >= startHour ? finishHour - startHour : finishHour + 24 - startHour
  return duration > 0 ? duration : undefined
}

function weeksToBitmask(values: number[]) {
  const map: Record<number, number> = { 1: 2, 2: 4, 3: 8, 4: 16, 5: 32, 6: 64, 7: 1 }
  return values.reduce((total, item) => total + (map[item] || 0), 0)
}

function bitmaskToWeeks(value?: number) {
  const current = Number(value || 0)
  const map: Array<[number, number]> = [[1, 2], [2, 4], [3, 8], [4, 16], [5, 32], [6, 64], [7, 1]]
  return map.filter(([, bit]) => (current & bit) === bit).map(([weekday]) => weekday)
}

function homeworkAttachmentsToMedia(list: HomeworkAttachmentItem[]) {
  return (Array.isArray(list) ? list : []).map(item => ({
    mediaType: Number(item.type) === 2 ? 'video' : 'image',
    url: item.url,
    fileName: item.name || '',
  }))
}

function mediaToHomeworkAttachments(list: RehabRecordMediaItem[]) {
  return (Array.isArray(list) ? list : []).map((item) => {
    const fileName = String(item.fileName || '').trim()
    return {
      type: item.mediaType === 'video' ? 2 : 1,
      url: String(item.url || '').trim(),
      duration: 0,
      name: fileName,
      extendName: fileName.includes('.') ? fileName.slice(fileName.lastIndexOf('.') + 1) : '',
    }
  }).filter(item => item.url)
}

function buildHomeworkObjects() {
  const map = new Map<string, { sourceType: number, sourceId: string, studentIds: string[] }>()
  formState.students.forEach((item) => {
    const sourceType = item.sourceType === 'one_to_one' ? 2 : 1
    const key = `${sourceType}:${item.sourceId}`
    if (!map.has(key)) {
      map.set(key, {
        sourceType,
        sourceId: item.sourceId,
        studentIds: [],
      })
    }
    const current = map.get(key)
    if (current && !current.studentIds.includes(item.studentId))
      current.studentIds.push(item.studentId)
  })
  return [...map.values()]
}

async function handleOk() {
  try {
    await formRef.value?.validate?.()
  }
  catch {
    return
  }

  if (formState.rule === 2) {
    if (formState.weeks.length === 0) {
      messageService.warning('请选择自动任务周期')
      return
    }
    if (!formState.dateRange || formState.dateRange.length !== 2) {
      messageService.warning('请选择任务日期范围')
      return
    }
    if (!formState.time) {
      messageService.warning('请选择任务推送时间')
      return
    }
    const taskDurationHours = normalizeTaskDurationHours(formState.taskDurationHours)
    if (taskDurationHours !== undefined && taskDurationHours <= 0) {
      messageService.warning('请填写正确的单次任务时长')
      return
    }
  }

  const homeworkObjects = buildHomeworkObjects()
  if (homeworkObjects.length === 0) {
    messageService.warning('请选择班级/学员')
    return
  }
  if (props.mode === 'edit' && homeworkObjects.length !== 1) {
    messageService.warning('编辑时仅支持一个班级或1对1对象')
    return
  }

  const currentDateRange = formState.dateRange || ['', '']
  const payload = {
    title: String(formState.title || '').trim(),
    content: String(formState.content || '').trim(),
    attachments: mediaToHomeworkAttachments(formState.mediaList),
    repeatRule: formState.rule === 2
      ? {
          startDate: currentDateRange[0],
          endDate: currentDateRange[1],
          repeatSpan: 1,
          weekDays: weeksToBitmask(formState.weeks),
        }
      : null,
    publishTime: formState.rule === 1 && formState.publishAt ? dayjs(formState.publishAt).format('YYYY-MM-DDTHH:mm') : undefined,
    endTime: formState.rule === 1 && formState.deadlineAt ? dayjs(formState.deadlineAt).format('YYYY-MM-DDTHH:mm') : undefined,
    publishHour: formState.rule === 2 ? timeOptionToHour(formState.time) : undefined,
    taskDurationHours: formState.rule === 2 ? normalizeTaskDurationHours(formState.taskDurationHours) : undefined,
    isVisibleStudent: false,
    homeworkObjects,
  }

  confirmLoading.value = true
  try {
    const res = props.mode === 'edit' && props.homeworkId
      ? await updateHomeworkApi({ id: props.homeworkId, ...payload })
      : await batchCreateHomeworksApi(payload)

    if (res.code !== 200) {
      messageService.error(res.message || (props.mode === 'edit' ? '编辑课后任务失败' : '创建课后任务失败'))
      return
    }

    const createdCount = Array.isArray(res.result) ? res.result.length : 0
    messageService.success(props.mode === 'edit' ? '编辑成功' : createdCount > 1 ? `创建成功，共生成${createdCount}条任务` : '创建成功')
    open.value = false
    emit('success')
  }
  catch (error) {
    console.error('submit homework failed', error)
    messageService.error(props.mode === 'edit' ? '编辑课后任务失败' : '创建课后任务失败')
  }
  finally {
    confirmLoading.value = false
  }
}

function handleRequestClose() {
  if (!isDirty.value) {
    open.value = false
    return
  }
  Modal.confirm({
    title: '关闭弹窗',
    content: '当前内容尚未保存，确认关闭吗？',
    okText: '确认关闭',
    cancelText: '继续编辑',
    onOk() {
      open.value = false
    },
  })
}

watch(() => open.value, (visible) => {
  if (!visible) {
    resetMediaOverlayState()
    resetStudentPickerState()
    return
  }
  void initializeForm()
})

watch(() => studentPickerOpen.value, (visible) => {
  if (!visible) {
    resetStudentPickerState()
    return
  }
  draftSelectedStudents.value = cloneSelectedStudents(formState.students)
  void loadStudentPickerData()
})

watch(() => studentPickerType.value, () => {
  if (!studentPickerOpen.value)
    return
  studentPickerKeyword.value = ''
  void loadStudentPickerData()
})

watch(() => studentPickerKeyword.value, () => {
  if (!studentPickerOpen.value)
    return
  if (studentPickerSearchTimer)
    clearTimeout(studentPickerSearchTimer)
  studentPickerSearchTimer = setTimeout(() => {
    void loadStudentPickerData()
  }, 300)
})

watch(() => formState.publishAt, (publishAtValue) => {
  if (!publishAtValue || !formState.deadlineAt)
    return
  const publishAt = dayjs(publishAtValue)
  const deadlineAt = dayjs(formState.deadlineAt)
  if (publishAt.isValid() && deadlineAt.isValid() && !deadlineAt.isAfter(publishAt))
    formState.deadlineAt = undefined
})
</script>

<template>
  <div>
    <a-modal
      :open="open"
      centered
      class="afterSchoolTasksModel"
      width="840px"
      :title="dialogTitle"
      :keyboard="false"
      :mask-closable="false"
      destroy-on-close
      @cancel="handleRequestClose"
    >
      <a-spin :spinning="detailLoading">
        <div class="afterSchoolTasksModel__body">
          <a-form ref="formRef" layout="vertical" :model="formState">
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
                v-model:value="formState.content"
                class="afterSchoolTasksModel__content-textarea"
                :maxlength="2000"
                :show-count="true"
                :auto-size="{ minRows: 4, maxRows: 4 }"
                placeholder="请输入任务内容，最多2000字"
              />
            </a-form-item>

            <a-form-item class="afterSchoolTasksModel__upload-form-item">
              <div class="afterSchoolTasksModel__upload-panel">
                <input
                  ref="imageInputRef"
                  class="afterSchoolTasksModel__upload-hidden-input"
                  type="file"
                  hidden
                  accept=".jpg,.jpeg,.png,.bmp,.webp,.gif"
                  multiple
                  @change="handleImageFileChange"
                >
                <input
                  ref="videoInputRef"
                  class="afterSchoolTasksModel__upload-hidden-input"
                  type="file"
                  hidden
                  accept=".mp4,.mov,.webm,.ogg,.m4v"
                  multiple
                  @change="handleVideoFileChange"
                >

                <div class="afterSchoolTasksModel__upload-actions">
                  <a-tooltip placement="right" @open-change="(show) => handleUploadActionHover(1, show)">
                    <template #title>
                      图片需小于等于 10MB
                    </template>
                    <div
                      class="afterSchoolTasksModel__upload-trigger"
                      :class="{
                        'is-active': activeUploadAction === 1,
                        'is-disabled': imageUploading || imageCount >= IMAGE_LIMIT,
                      }"
                      @click="openImagePicker"
                    >
                      <LoadingOutlined v-if="imageUploading" spin class="afterSchoolTasksModel__upload-trigger-icon" />
                      <PictureOutlined v-else class="afterSchoolTasksModel__upload-trigger-icon" />
                      <span>添加图片({{ imageCount }}/{{ IMAGE_LIMIT }})</span>
                    </div>
                  </a-tooltip>

                  <a-tooltip placement="right" @open-change="(show) => handleUploadActionHover(2, show)">
                    <template #title>
                      视频需小于等于 100MB
                    </template>
                    <div
                      class="afterSchoolTasksModel__upload-trigger"
                      :class="{
                        'is-active': activeUploadAction === 2,
                        'is-disabled': videoUploading || videoCount >= VIDEO_LIMIT,
                      }"
                      @click="openVideoPicker"
                    >
                      <LoadingOutlined v-if="videoUploading" spin class="afterSchoolTasksModel__upload-trigger-icon" />
                      <PlayCircleOutlined v-else class="afterSchoolTasksModel__upload-trigger-icon" />
                      <span>添加视频({{ videoCount }}/{{ VIDEO_LIMIT }})</span>
                    </div>
                  </a-tooltip>
                </div>

                <div v-if="formState.mediaList.length > 0" class="afterSchoolTasksModel__media-list">
                  <div
                    v-for="(item, index) in formState.mediaList"
                    :key="`${item.mediaType}-${item.url}-${index}`"
                    class="afterSchoolTasksModel__media-card"
                    @click="handleMediaPreview(item, index)"
                  >
                    <img
                      v-if="item.mediaType === 'image'"
                      class="afterSchoolTasksModel__media-cover"
                      :src="item.url"
                      :alt="resolveMediaLabel(item, index)"
                    >
                    <div v-else class="afterSchoolTasksModel__media-video">
                      <video
                        class="afterSchoolTasksModel__media-cover"
                        :src="item.url"
                        muted
                        playsinline
                        preload="metadata"
                      />
                      <div class="afterSchoolTasksModel__media-play">
                        <PlayCircleOutlined />
                      </div>
                    </div>

                    <div class="afterSchoolTasksModel__media-label">
                      {{ resolveMediaLabel(item, index) }}
                    </div>

                    <button
                      type="button"
                      class="afterSchoolTasksModel__media-remove"
                      @click.stop="handleRemoveMedia(item)"
                    >
                      <CloseOutlined />
                    </button>
                  </div>
                </div>
              </div>
            </a-form-item>

            <a-form-item label="选择班级/学员" name="students" :rules="[{ required: true, message: '请选择班级/学员' }]">
              <div class="afterSchoolTasksModel__student-selector">
                <a-button type="primary" ghost @click="openStudentPicker">
                  {{ selectedStudentButtonText }}
                </a-button>
                <div v-if="selectedStudentPreviewText" class="afterSchoolTasksModel__student-preview">
                  {{ selectedStudentPreviewText }}
                </div>
              </div>
            </a-form-item>

            <a-form-item v-if="!isEditMode" label="发布规则" :required="true">
              <a-radio-group v-model:value="formState.rule" class="custom-radio">
                <a-radio :value="1">
                  仅本次发布
                </a-radio>
                <a-radio :value="2">
                  设置自动任务
                </a-radio>
              </a-radio-group>
            </a-form-item>

            <a-form-item
              v-if="formState.rule === 1"
              class="afterSchoolTasksModel__rule-card-item"
              :class="{ 'afterSchoolTasksModel__rule-card-item--edit': isEditMode }"
            >
              <div class="afterSchoolTasksModel__rule-card">
                <div class="afterSchoolTasksModel__rule-card-title">
                  设置本次发布时间（非必填）
                </div>
                <a-row :gutter="[16, 16]">
                  <a-col :xs="24" :sm="12">
                    <div class="afterSchoolTasksModel__rule-field">
                      <div class="afterSchoolTasksModel__rule-label">
                        <span>定时发布日期</span>
                        <a-popover color="#fff" placement="topLeft" title="定时发布日期">
                          <template #content>
                            <div class="afterSchoolTasksModel__rule-popover">
                              设置后，任务创建完成会按设置时间发送；不设置则立即发送。
                            </div>
                          </template>
                          <QuestionCircleOutlined class="afterSchoolTasksModel__rule-tip" />
                        </a-popover>
                        <span>：</span>
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
                        <a-popover color="#fff" placement="topLeft" title="任务截止日期">
                          <template #content>
                            <div class="afterSchoolTasksModel__rule-popover">
                              设置后，学员超时提交会在系统中标记为超时提交。
                            </div>
                          </template>
                          <QuestionCircleOutlined class="afterSchoolTasksModel__rule-tip" />
                        </a-popover>
                        <span>：</span>
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

            <a-form-item
              v-if="formState.rule === 2"
              class="afterSchoolTasksModel__rule-card-item"
              :class="{ 'afterSchoolTasksModel__rule-card-item--edit': isEditMode }"
            >
              <div class="afterSchoolTasksModel__rule-card">
                <div class="afterSchoolTasksModel__rule-card-title">
                  设置自动任务周期
                </div>
                <div class="afterSchoolTasksModel__week-list">
                  <div
                    v-for="week in weeks"
                    :key="week.value"
                    class="afterSchoolTasksModel__week-item"
                    :class="{ 'is-active': formState.weeks.includes(week.value) }"
                    @click="toggleWeek(week.value)"
                  >
                    {{ week.label }}
                  </div>
                </div>
                <a-row :gutter="[16, 16]" class="mt-4">
                  <a-col :xs="24" :sm="12">
                    <a-form-item label="任务日期范围" name="dateRange" :rules="[{ required: true, message: '请选择任务日期范围' }]">
                      <a-range-picker v-model:value="formState.dateRange" class="w-full" value-format="YYYY-MM-DD" :disabled-date="disabledDate" />
                    </a-form-item>
                  </a-col>
                  <a-col :xs="24" :sm="6">
                    <a-form-item label="任务推送时间" name="time" :rules="[{ required: true, message: '请选择任务推送时间' }]">
                      <a-select v-model:value="formState.time" placeholder="请选择" :options="dateOptions" />
                    </a-form-item>
                  </a-col>
                  <a-col :xs="24" :sm="6">
                    <a-form-item name="taskDurationHours">
                      <template #label>
                        <span>单次任务时长</span>
                        <a-popover color="#fff" placement="topLeft" title="单次任务时长">
                          <template #content>
                            <div class="afterSchoolTasksModel__rule-popover">
                              举例：设置任务推送时间为今日6:00发送，单次任务时长8小时，学员在今日14:00点前为正常提交，超过14:00点为超时提交。
                            </div>
                          </template>
                          <QuestionCircleOutlined class="afterSchoolTasksModel__rule-tip" />
                        </a-popover>
                        <span>：</span>
                      </template>
                      <a-input-number
                        v-model:value="formState.taskDurationHours"
                        class="w-full afterSchoolTasksModel__duration-input"
                        :min="1"
                        :precision="0"
                        :controls="false"
                        placeholder="请输入"
                      >
                        <template #addonAfter>
                          <span>小时</span>
                        </template>
                      </a-input-number>
                    </a-form-item>
                  </a-col>
                </a-row>
              </div>
            </a-form-item>
          </a-form>
        </div>
      </a-spin>

      <template #footer>
        <div class="afterSchoolTasksModel__footer">
          <a-button @click="handleRequestClose">
            取消
          </a-button>
          <a-button type="primary" :loading="confirmLoading" @click="handleOk">
            {{ props.mode === 'edit' ? '保存' : '发布' }}
          </a-button>
        </div>
      </template>
    </a-modal>

    <a-modal
      :open="studentPickerOpen"
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
            @click="studentPickerType = item.key as 'class' | 'one_to_one'"
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

    <a-modal
      :open="previewImageOpen"
      :title="previewImageTitle"
      :footer="null"
      @cancel="previewImageOpen = false"
    >
      <img
        alt="图片预览"
        style="width: 100%;"
        :src="previewImageSrc"
      >
    </a-modal>

    <a-modal
      :open="previewVideoOpen"
      :title="previewVideoItem?.fileName || '视频预览'"
      :footer="null"
      width="720px"
      :destroy-on-close="true"
      @cancel="closeVideoPreview"
      @after-close="closeVideoPreview"
    >
      <video
        v-if="previewVideoItem?.url"
        ref="previewVideoRef"
        class="afterSchoolTasksModel__preview-video"
        :src="previewVideoItem.url"
        controls
        playsinline
      />
    </a-modal>
  </div>
</template>

<style scoped lang="less">
.afterSchoolTasksModel__body {
  max-height: 580px;
  overflow-y: auto;
  padding-right: 8px;
  margin-right: -4px;
  overscroll-behavior: contain;
  scrollbar-gutter: stable;
  scrollbar-width: thin;
  scrollbar-color: #cfd6e4 transparent;
}

.afterSchoolTasksModel__content-item {
  margin-bottom: 10px;
}

.afterSchoolTasksModel__content-textarea {
  min-height: 66px !important;
}

.afterSchoolTasksModel__upload-form-item {
  margin-top: -12px;
  margin-bottom: 14px;
}

.afterSchoolTasksModel__rule-card-item {
  margin-top: -20px;
}

.afterSchoolTasksModel__rule-card-item--edit {
  margin-top: 8px;
}

.afterSchoolTasksModel__upload-panel {
  width: 100%;
}

.afterSchoolTasksModel__upload-hidden-input {
  display: none;
}

.afterSchoolTasksModel__upload-actions {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-start;
  gap: 10px;
}

.afterSchoolTasksModel__upload-trigger {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  min-width: 135px;
  height: 34px;
  padding: 0 12px;
  border-radius: 999px;
  background: #f6f7f8;
  color: #595959;
  font-size: 14px;
  line-height: 20px;
  cursor: pointer;
  transition: all 0.2s ease;

  &.is-active {
    background: var(--pro-ant-color-primary);
    color: #fff;
  }

  &.is-disabled {
    opacity: 0.56;
    cursor: not-allowed;
  }
}

.afterSchoolTasksModel__upload-trigger-icon {
  color: var(--pro-ant-color-primary);
  font-size: 15px;
  transition: color 0.2s ease;
}

.afterSchoolTasksModel__upload-trigger.is-active .afterSchoolTasksModel__upload-trigger-icon {
  color: #fff;
}

.afterSchoolTasksModel__media-list {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  margin-top: 12px;
}

.afterSchoolTasksModel__media-card {
  position: relative;
  width: 108px;
  height: 72px;
  flex: 0 0 auto;
  overflow: hidden;
  border: 1px solid #edf0f5;
  border-radius: 10px;
  background: #f5f7fb;
  cursor: pointer;
}

.afterSchoolTasksModel__media-cover {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.afterSchoolTasksModel__media-video {
  position: relative;
  width: 100%;
  height: 100%;
  background: #0f172a;
}

.afterSchoolTasksModel__media-play {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-size: 22px;
  background: rgba(15, 23, 42, 0.16);
}

.afterSchoolTasksModel__media-label {
  position: absolute;
  right: 0;
  bottom: 0;
  left: 0;
  height: 24px;
  padding: 0 8px;
  overflow: hidden;
  color: #fff;
  font-size: 12px;
  line-height: 24px;
  text-overflow: ellipsis;
  white-space: nowrap;
  background: linear-gradient(180deg, rgba(15, 23, 42, 0) 0%, rgba(15, 23, 42, 0.72) 100%);
}

.afterSchoolTasksModel__media-remove {
  position: absolute;
  top: 6px;
  right: 6px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  padding: 0;
  border: none;
  border-radius: 999px;
  background: rgba(15, 23, 42, 0.52);
  color: #fff;
  cursor: pointer;

  &:hover {
    background: rgba(15, 23, 42, 0.72);
  }
}

.afterSchoolTasksModel__student-selector {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 8px;
}

.afterSchoolTasksModel__student-selector :deep(.ant-btn) {
  width: auto;
  min-width: 0;
}

.afterSchoolTasksModel__student-preview {
  width: 100%;
  min-height: 42px;
  padding: 9px 14px;
  border: 1px solid #d9d9d9;
  border-radius: 8px;
  background: #fafafa;
  color: #595959;
  font-size: 14px;
  line-height: 22px;
  word-break: break-all;
}

.afterSchoolTasksModel__rule-card {
  border: 1px solid #edf0f5;
  border-radius: 12px;
  background: #fff;
  padding: 16px;
}

.afterSchoolTasksModel__rule-card-title {
  margin-bottom: 14px;
  color: #262626;
  font-size: 14px;
  line-height: 22px;
  font-weight: 500;
}

.afterSchoolTasksModel__rule-field {
  display: flex;
  flex-direction: column;
  gap: 8px;
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
  max-width: 320px;
  color: #595959;
  font-size: 14px;
  line-height: 22px;
}

.afterSchoolTasksModel__duration-input :deep(.ant-input-number-group-addon) {
  min-width: 48px;
  padding: 0 12px;
  color: #595959;
}

.afterSchoolTasksModel__week-list {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.afterSchoolTasksModel__week-item {
  min-width: 64px;
  height: 34px;
  padding: 0 14px;
  border: 1px solid #d9d9d9;
  border-radius: 999px;
  color: #595959;
  font-size: 13px;
  line-height: 32px;
  text-align: center;
  cursor: pointer;
  transition: all 0.2s ease;

  &.is-active {
    border-color: var(--pro-ant-color-primary);
    background: #edf5ff;
    color: var(--pro-ant-color-primary);
  }
}

.afterSchoolTasksModel__footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

.afterSchoolTasksModel__preview-video {
  width: 100%;
  max-height: 70vh;
  border-radius: 12px;
  background: #000;
}

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

:deep(.afterSchoolTasksModel .ant-modal-content) {
  padding-bottom: 12px;
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

:deep(.afterSchoolTasksModel__content-item .ant-form-item-explain-error) {
  margin-top: 4px;
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

:deep(.afterSchoolTasksModel__body::-webkit-scrollbar),
:deep(.afterSchoolTasksModel__student-picker-content::-webkit-scrollbar) {
  width: 8px;
}

:deep(.afterSchoolTasksModel__body::-webkit-scrollbar-track),
:deep(.afterSchoolTasksModel__student-picker-content::-webkit-scrollbar-track) {
  background: transparent;
}

:deep(.afterSchoolTasksModel__body::-webkit-scrollbar-thumb),
:deep(.afterSchoolTasksModel__student-picker-content::-webkit-scrollbar-thumb) {
  border: 2px solid transparent;
  border-radius: 999px;
  background: #cfd6e4;
  background-clip: padding-box;
}

:deep(.afterSchoolTasksModel__body::-webkit-scrollbar-thumb:hover),
:deep(.afterSchoolTasksModel__student-picker-content::-webkit-scrollbar-thumb:hover) {
  background: #b9c3d4;
  background-clip: padding-box;
}

:deep(.media-selection-dialog__intro) {
  margin-bottom: 8px;
  color: #595959;
  line-height: 22px;
}

:deep(.media-selection-dialog__item) {
  color: #262626;
  line-height: 22px;
  word-break: break-all;
}
</style>
