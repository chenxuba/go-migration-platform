<script setup lang="ts">
import { CloseOutlined } from '@ant-design/icons-vue'
import { Modal } from 'ant-design-vue'
import dayjs from 'dayjs'
import { computed, reactive, ref, watch } from 'vue'
import { getStudentRehabRecordDetailApi, publishStudentRehabRecordApi, saveStudentRehabRecordDraftApi, type RehabRecordContent, type RehabRecordMediaItem, type RehabRecordTrainingItem, type StudentRehabRecordSnapshot } from '@/api/edu-center/class-record'
import RehabRecordMediaUpload from './rehab-record-media-upload.vue'
import messageService from '@/utils/messageService'

interface RehabRecordStudent {
  id?: string
  name?: string
  avatar?: string
  status?: string
  gender?: string | number
  sex?: string | number
  stuSex?: string | number
  birthday?: string
  birthDay?: string
  birthDate?: string
}

interface RehabRecordSession {
  sourceName?: string
  lessonName?: string
  teacherName?: string
  classRoomName?: string
  startTime?: string
  endTime?: string
}

interface RehabRecordParentFeedback {
  content?: string
  parentFeedbackContent?: string
  signature?: string
  parentSignature?: string
  date?: string
  feedbackDate?: string
}

const props = withDefaults(defineProps<{
  studentTeachingRecordId?: string
  mode?: 'create' | 'view' | 'edit'
  student?: Partial<RehabRecordStudent> | null
  session?: Partial<RehabRecordSession> | null
  parentFeedback?: Partial<RehabRecordParentFeedback> | null
}>(), {
  studentTeachingRecordId: '',
  mode: 'create',
  student: null,
  session: null,
  parentFeedback: null,
})

const emit = defineEmits<{
  (e: 'published'): void
}>()

const open = defineModel<boolean>({
  default: false,
})

const genderOptions = [
  { label: '男', value: '男' },
  { label: '女', value: '女' },
]

interface TrainingModuleItem {
  id: number
  title: string
  content: string
}

const trainingModuleSeed = ref(0)
const trainingModules = ref<TrainingModuleItem[]>([])
const loading = ref(false)
const submittingAction = ref<'draft' | 'publish' | ''>('')
const loadSequence = ref(0)
const previousPublishedSnapshot = ref<StudentRehabRecordSnapshot | null>(null)

const formModel = reactive({
  studentName: '',
  gender: undefined as string | undefined,
  birthDate: undefined as string | undefined,
  className: '',
  teacherName: '',
  trainingDate: undefined as string | undefined,
  trainingTarget: '',
  trainingMediaList: [] as RehabRecordMediaItem[],
  performance: '',
  performanceMediaList: [] as RehabRecordMediaItem[],
  suggestion: '',
  suggestionMediaList: [] as RehabRecordMediaItem[],
  parentFeedback: '',
  parentSignature: '',
  feedbackDate: undefined as string | undefined,
})

function normalizeTextValue(value?: string | number | null) {
  return String(value ?? '').trim()
}

function inferMediaType(url: string) {
  if (/\.(mp4|mov|webm|ogg|m4v)(\?.*)?$/i.test(url))
    return 'video'
  return 'image'
}

function normalizeMediaList(items?: RehabRecordMediaItem[] | null) {
  return Array.isArray(items)
    ? items
        .map((item) => {
          const url = normalizeTextValue(item?.url)
          if (!url)
            return null

          const rawMediaType = normalizeTextValue(item?.mediaType).toLowerCase()
          const mediaType = rawMediaType === 'image' || rawMediaType === 'video'
            ? rawMediaType
            : inferMediaType(url)
          const size = Number(item?.size || 0)
          return {
            mediaType,
            url,
            fileName: normalizeTextValue(item?.fileName),
            size: Number.isFinite(size) && size > 0 ? size : undefined,
          } satisfies RehabRecordMediaItem
        })
        .filter(Boolean) as RehabRecordMediaItem[]
    : []
}

function resolveTrainingMediaList(content?: Partial<RehabRecordContent> | null) {
  const mediaList = normalizeMediaList(content?.trainingMediaList)
  if (mediaList.length > 0)
    return mediaList

  return Array.isArray(content?.trainingItems)
    ? normalizeMediaList(content.trainingItems.flatMap(item => item?.mediaList || []))
    : []
}

function createTrainingModule(title = '', content = '') {
  trainingModuleSeed.value += 1
  return {
    id: trainingModuleSeed.value,
    title,
    content,
  }
}

function resetTrainingModules(items?: RehabRecordTrainingItem[]) {
  trainingModuleSeed.value = 0
  const normalized = Array.isArray(items)
    ? items
        .map(item => ({
          title: normalizeTextValue(item?.title),
          content: normalizeTextValue(item?.content),
        }))
        .filter(item => item.title || item.content)
    : []

  trainingModules.value = normalized.length > 0
    ? normalized.map(item => createTrainingModule(item.title, item.content))
    : [createTrainingModule(), createTrainingModule()]
}

function resolveGenderValue() {
  const rawValue = props.student?.gender ?? props.student?.sex ?? props.student?.stuSex
  if (rawValue === undefined || rawValue === null || rawValue === '')
    return undefined

  if (rawValue === 1 || rawValue === '1' || rawValue === '男')
    return '男'

  if (rawValue === 0 || rawValue === '0' || rawValue === '女')
    return '女'

  if (rawValue === 2 || rawValue === '2' || rawValue === '未知')
    return undefined

  return normalizeTextValue(rawValue) || undefined
}

function normalizeDateValue(value?: string | null) {
  if (!value)
    return undefined
  const date = dayjs(value)
  return date.isValid() ? date.format('YYYY-MM-DD') : undefined
}

function hydrateBasicInfo() {
  formModel.studentName = normalizeTextValue(props.student?.name)
  formModel.gender = resolveGenderValue()
  formModel.birthDate = normalizeDateValue(props.student?.birthDate || props.student?.birthDay || props.student?.birthday)
  formModel.className = normalizeTextValue(props.session?.sourceName)
  formModel.teacherName = normalizeTextValue(props.session?.teacherName)
  formModel.trainingDate = normalizeDateValue(props.session?.startTime)
}

function hydrateParentFeedback() {
  formModel.parentFeedback = normalizeTextValue(props.parentFeedback?.content || props.parentFeedback?.parentFeedbackContent)
  formModel.parentSignature = normalizeTextValue(props.parentFeedback?.signature || props.parentFeedback?.parentSignature)
  formModel.feedbackDate = normalizeDateValue(props.parentFeedback?.date || props.parentFeedback?.feedbackDate)
}

function hydrateDefaultForm() {
  hydrateBasicInfo()
  hydrateParentFeedback()
  formModel.trainingTarget = ''
  formModel.trainingMediaList = []
  formModel.performance = ''
  formModel.performanceMediaList = []
  formModel.suggestion = ''
  formModel.suggestionMediaList = []
  resetTrainingModules()
}

const currentStudentTeachingRecordId = computed(() => {
  return normalizeTextValue(props.studentTeachingRecordId || props.student?.id)
})
const isReadonly = computed(() => props.mode === 'view')
const isEditMode = computed(() => props.mode === 'edit')
const drawerTitle = computed(() => {
  if (props.mode === 'view')
    return '查看康复训练记录'
  if (props.mode === 'edit')
    return '编辑康复训练记录'
  return '康复训练记录'
})
const publishButtonText = computed(() => isEditMode.value ? '更新记录' : '发布记录')
const canReusePreviousPublished = computed(() => {
  return !isReadonly.value && Boolean(previousPublishedSnapshot.value?.content)
})
const signatureImageSrc = computed(() => {
  const value = normalizeTextValue(formModel.parentSignature)
  if (!value)
    return ''
  if (/^data:image\//i.test(value))
    return value
  if (/^(https?:)?\/\//i.test(value))
    return value
  if (value.startsWith('/'))
    return value
  if (/\.(png|jpe?g|gif|webp|bmp|svg)(\?.*)?$/i.test(value))
    return value
  return ''
})

function handleAddTrainingModule() {
  trainingModules.value.push(createTrainingModule())
}

function handleRemoveTrainingModule(id: number) {
  if (trainingModules.value.length <= 1)
    return
  trainingModules.value = trainingModules.value.filter(item => item.id !== id)
}

function getTrainingModuleColSpan(index: number) {
  const count = trainingModules.value.length
  if (count === 1)
    return 24
  if (count % 2 === 1 && index === count - 1)
    return 24
  return 12
}

function applyContent(content?: Partial<RehabRecordContent> | null) {
  if (!content)
    return

  formModel.studentName = normalizeTextValue(content.studentName) || formModel.studentName
  formModel.gender = normalizeTextValue(content.gender) || formModel.gender
  formModel.birthDate = normalizeDateValue(content.birthDate) || formModel.birthDate
  formModel.className = normalizeTextValue(content.className) || formModel.className
  formModel.teacherName = normalizeTextValue(content.teacherName) || formModel.teacherName
  formModel.trainingDate = normalizeDateValue(content.trainingDate) || formModel.trainingDate
  formModel.trainingTarget = normalizeTextValue(content.trainingTarget)
  formModel.trainingMediaList = resolveTrainingMediaList(content)
  formModel.performance = normalizeTextValue(content.performance)
  formModel.performanceMediaList = normalizeMediaList(content.performanceMediaList)
  formModel.suggestion = normalizeTextValue(content.suggestion)
  formModel.suggestionMediaList = normalizeMediaList(content.suggestionMediaList)
  formModel.parentFeedback = normalizeTextValue(content.parentFeedback)
  formModel.parentSignature = normalizeTextValue(content.parentSignature)
  formModel.feedbackDate = normalizeDateValue(content.feedbackDate)
  resetTrainingModules(content.trainingItems)
}

function applySnapshot(snapshot?: StudentRehabRecordSnapshot | null) {
  applyContent(snapshot?.content)
}

function applyReusableContent(snapshot?: StudentRehabRecordSnapshot | null) {
  const content = snapshot?.content
  if (!content)
    return

  formModel.trainingTarget = normalizeTextValue(content.trainingTarget)
  formModel.trainingMediaList = resolveTrainingMediaList(content)
  formModel.performance = normalizeTextValue(content.performance)
  formModel.performanceMediaList = normalizeMediaList(content.performanceMediaList)
  formModel.suggestion = normalizeTextValue(content.suggestion)
  formModel.suggestionMediaList = normalizeMediaList(content.suggestionMediaList)
  resetTrainingModules(content.trainingItems)
}

function buildContentPayload(): RehabRecordContent {
  return {
    studentName: normalizeTextValue(formModel.studentName),
    gender: normalizeTextValue(formModel.gender),
    birthDate: normalizeTextValue(formModel.birthDate),
    className: normalizeTextValue(formModel.className),
    teacherName: normalizeTextValue(formModel.teacherName),
    trainingDate: normalizeTextValue(formModel.trainingDate),
    trainingTarget: normalizeTextValue(formModel.trainingTarget),
    trainingMediaList: normalizeMediaList(formModel.trainingMediaList),
    trainingItems: trainingModules.value.map(item => ({
      title: normalizeTextValue(item.title),
      content: normalizeTextValue(item.content),
    })),
    performance: normalizeTextValue(formModel.performance),
    performanceMediaList: normalizeMediaList(formModel.performanceMediaList),
    suggestion: normalizeTextValue(formModel.suggestion),
    suggestionMediaList: normalizeMediaList(formModel.suggestionMediaList),
    parentFeedback: normalizeTextValue(formModel.parentFeedback),
    parentSignature: normalizeTextValue(formModel.parentSignature),
    feedbackDate: normalizeTextValue(formModel.feedbackDate),
  }
}

function applySnapshotByMode(options: {
  draft?: StudentRehabRecordSnapshot | null
  published?: StudentRehabRecordSnapshot | null
  currentLoad: number
}) {
  const { draft, published, currentLoad } = options
  if (props.mode === 'view') {
    applySnapshot(published || draft)
    return
  }

  if (draft?.content) {
    const title = draft.updatedTime ? '检测到未发布草稿' : '检测到草稿'
    const content = draft.updatedTime
      ? `检测到 ${draft.updatedTime} 保存的草稿，是否恢复并继续编辑？`
      : '检测到未发布草稿，是否恢复并继续编辑？'
    Modal.confirm({
      title,
      content,
      okText: '恢复草稿',
      cancelText: props.mode === 'edit' && published?.content ? '使用已发布内容' : '暂不恢复',
      onOk: () => {
        if (!open.value || currentLoad !== loadSequence.value)
          return
        applySnapshot(draft)
      },
      onCancel: () => {
        if (!open.value || currentLoad !== loadSequence.value)
          return
        if (props.mode === 'edit' && published?.content)
          applySnapshot(published)
      },
    })
    return
  }

  if (props.mode === 'edit' && published?.content)
    applySnapshot(published)
}

async function loadRecordIfNeeded() {
  const studentTeachingRecordId = currentStudentTeachingRecordId.value
  if (!open.value || !studentTeachingRecordId)
    return

  const currentLoad = ++loadSequence.value
  loading.value = true
  try {
    const res = await getStudentRehabRecordDetailApi({ studentTeachingRecordId })
    if (!open.value || currentLoad !== loadSequence.value)
      return
    if (res.code !== 200)
      throw new Error(res.message || '加载康复记录失败')

    const draft = res.result?.hasDraft ? res.result?.draft : null
    const published = res.result?.hasPublished ? res.result?.published : null
    previousPublishedSnapshot.value = res.result?.hasPreviousPublished ? (res.result?.previousPublished || null) : null
    applySnapshotByMode({ draft, published, currentLoad })
  }
  catch (error: any) {
    previousPublishedSnapshot.value = null
    messageService.error(error?.response?.data?.message || error?.message || '加载康复记录失败')
  }
  finally {
    if (currentLoad === loadSequence.value)
      loading.value = false
  }
}

async function handleSaveDraft() {
  if (isReadonly.value)
    return
  const studentTeachingRecordId = currentStudentTeachingRecordId.value
  if (!studentTeachingRecordId) {
    messageService.warning('缺少有效的学员记录')
    return
  }

  submittingAction.value = 'draft'
  try {
    const res = await saveStudentRehabRecordDraftApi({
      studentTeachingRecordId,
      content: buildContentPayload(),
    })
    if (res.code !== 200)
      throw new Error(res.message || '保存草稿失败')
    messageService.success('草稿已保存')
  }
  catch (error: any) {
    messageService.error(error?.response?.data?.message || error?.message || '保存草稿失败')
  }
  finally {
    submittingAction.value = ''
  }
}

async function handlePublish() {
  if (isReadonly.value)
    return
  const studentTeachingRecordId = currentStudentTeachingRecordId.value
  if (!studentTeachingRecordId) {
    messageService.warning('缺少有效的学员记录')
    return
  }

  submittingAction.value = 'publish'
  try {
    const res = await publishStudentRehabRecordApi({
      studentTeachingRecordId,
      content: buildContentPayload(),
    })
    if (res.code !== 200)
      throw new Error(res.message || '发布记录失败')
    messageService.success('康复记录发布成功')
    open.value = false
    emit('published')
  }
  catch (error: any) {
    messageService.error(error?.response?.data?.message || error?.message || '发布记录失败')
  }
  finally {
    submittingAction.value = ''
  }
}

function handleReusePreviousPublished() {
  if (!canReusePreviousPublished.value || !previousPublishedSnapshot.value) {
    messageService.warning('同课程暂无可复用的已发布康复记录')
    return
  }

  Modal.confirm({
    title: '复用上一篇',
    content: '将使用同课程最近一篇已发布康复记录覆盖当前训练内容，基础信息和家长反馈会保留当前值。',
    okText: '确认复用',
    cancelText: '取消',
    onOk: () => {
      applyReusableContent(previousPublishedSnapshot.value)
      messageService.success('已复用上一篇康复记录')
    },
  })
}

watch(
  [() => open.value, () => props.mode, currentStudentTeachingRecordId, () => props.student, () => props.session, () => props.parentFeedback],
  async ([isOpen]) => {
    if (!isOpen) {
      loadSequence.value += 1
      loading.value = false
      submittingAction.value = ''
      previousPublishedSnapshot.value = null
      resetTrainingModules()
      return
    }
    hydrateDefaultForm()
    await loadRecordIfNeeded()
  },
  { deep: true, immediate: true },
)
</script>

<template>
  <a-drawer
    v-model:open="open"
    :body-style="{ padding: '0', background: '#f7f7fd', display: 'flex', flexDirection: 'column' }"
    :closable="false"
    width="980px"
    placement="right"
    :z-index="1100"
  >
    <template #title>
      <div class="custom-header flex justify-between h-4 flex-items-center">
        <div class="text-5">
          {{ drawerTitle }}
        </div>
        <a-button type="text" class="close-btn" @click="open = false">
          <template #icon>
            <CloseOutlined class="text-5 close-icon" />
          </template>
        </a-button>
      </div>
    </template>

    <div class="h-full flex flex-col min-h-0">
      <a-spin class="flex-1 min-h-0" :spinning="loading">
        <div class="flex-1 min-h-0 overflow-auto p-12px">
          <div v-if="isReadonly" class="flex flex-col gap-16px">
            <div class="rounded-14px border border-solid border-#e9edf5 bg-white p-18px">
              <div class="text-15px leading-22px font-600 text-#222">
                基础信息
              </div>

              <a-row :gutter="[12, 12]" class="mt-14px">
                <a-col :xs="24" :sm="12" :lg="8">
                  <div class="rounded-12px border border-solid border-#edf0f5 bg-#fafbfc p-14px">
                    <div class="mb-10px text-12px leading-18px text-#8c8c8c">姓名</div>
                    <a-input v-model:value="formModel.studentName" disabled placeholder="请输入学员姓名" />
                  </div>
                </a-col>
                <a-col :xs="24" :sm="12" :lg="8">
                  <div class="rounded-12px border border-solid border-#edf0f5 bg-#fafbfc p-14px">
                    <div class="mb-10px text-12px leading-18px text-#8c8c8c">性别</div>
                    <a-select
                      v-model:value="formModel.gender"
                      :options="genderOptions"
                      disabled
                      allow-clear
                      style="width: 100%;"
                      placeholder="请选择性别"
                    />
                  </div>
                </a-col>
                <a-col :xs="24" :sm="12" :lg="8">
                  <div class="rounded-12px border border-solid border-#edf0f5 bg-#fafbfc p-14px">
                    <div class="mb-10px text-12px leading-18px text-#8c8c8c">出生年月</div>
                    <a-date-picker
                      v-model:value="formModel.birthDate"
                      disabled
                      value-format="YYYY-MM-DD"
                      style="width: 100%;"
                      placeholder="请选择出生年月"
                    />
                  </div>
                </a-col>
                <a-col :xs="24" :sm="12" :lg="8">
                  <div class="rounded-12px border border-solid border-#edf0f5 bg-#fafbfc p-14px">
                    <div class="mb-10px text-12px leading-18px text-#8c8c8c">班别</div>
                    <a-input v-model:value="formModel.className" disabled placeholder="请输入班别" />
                  </div>
                </a-col>
                <a-col :xs="24" :sm="12" :lg="8">
                  <div class="rounded-12px border border-solid border-#edf0f5 bg-#fafbfc p-14px">
                    <div class="mb-10px text-12px leading-18px text-#8c8c8c">任教老师</div>
                    <a-input v-model:value="formModel.teacherName" disabled placeholder="请输入任教老师" />
                  </div>
                </a-col>
                <a-col :xs="24" :sm="12" :lg="8">
                  <div class="rounded-12px border border-solid border-#edf0f5 bg-#fafbfc p-14px">
                    <div class="mb-10px text-12px leading-18px text-#8c8c8c">训练日期</div>
                    <a-date-picker
                      v-model:value="formModel.trainingDate"
                      disabled
                      value-format="YYYY-MM-DD"
                      style="width: 100%;"
                      placeholder="请选择训练日期"
                    />
                  </div>
                </a-col>
              </a-row>
            </div>

            <div class="rounded-14px border border-solid border-#e9edf5 bg-white p-18px">
              <div class="text-15px leading-22px font-600 text-#222">
                训练目标
              </div>
              <div class="mt-14px">
                <a-textarea
                  v-model:value="formModel.trainingTarget"
                  disabled
                  :auto-size="{ minRows: 4, maxRows: 6 }"
                  placeholder="请输入本次训练目标"
                />
              </div>
            </div>

            <div class="rounded-14px border border-solid border-#e9edf5 bg-white p-18px">
              <div class="text-15px leading-22px font-600 text-#222">
                训练项目
              </div>

              <a-row :gutter="[12, 12]" class="mt-14px">
                <a-col
                  v-for="(item, index) in trainingModules"
                  :key="item.id"
                  :xs="24"
                  :lg="getTrainingModuleColSpan(index)"
                >
                  <div class="h-full rounded-12px border border-solid border-#edf0f5 bg-#fafbfc p-16px">
                    <div class="flex items-center justify-between gap-12px mb-12px">
                      <a-input
                        v-model:value="item.title"
                        disabled
                        class="flex-1"
                        placeholder="请输入训练项目名称"
                      />
                    </div>

                    <a-textarea
                      v-model:value="item.content"
                      disabled
                      :auto-size="{ minRows: 5, maxRows: 8 }"
                      placeholder="请输入训练项目内容"
                    />
                  </div>
                </a-col>
              </a-row>

              <div v-if="formModel.trainingMediaList.length > 0" class="mt-14px">
                <RehabRecordMediaUpload
                  :model-value="formModel.trainingMediaList"
                  disabled
                />
              </div>
            </div>

            <a-row :gutter="[16, 16]">
              <a-col :xs="24" :lg="12">
                <div class="h-full rounded-14px border border-solid border-#e9edf5 bg-white p-18px">
                  <div class="text-15px leading-22px font-600 text-#222">
                    学生综合表现
                  </div>
                  <div class="mt-14px">
                    <a-textarea
                      v-model:value="formModel.performance"
                      disabled
                      :auto-size="{ minRows: 7, maxRows: 10 }"
                      placeholder="请输入学生综合表现"
                    />

                    <div v-if="formModel.performanceMediaList.length > 0" class="mt-14px">
                      <RehabRecordMediaUpload
                        :model-value="formModel.performanceMediaList"
                        disabled
                      />
                    </div>
                  </div>
                </div>
              </a-col>
              <a-col :xs="24" :lg="12">
                <div class="h-full rounded-14px border border-solid border-#e9edf5 bg-white p-18px">
                  <div class="text-15px leading-22px font-600 text-#222">
                    康复建议
                  </div>
                  <div class="mt-14px">
                    <a-textarea
                      v-model:value="formModel.suggestion"
                      disabled
                      :auto-size="{ minRows: 7, maxRows: 10 }"
                      placeholder="请输入康复建议"
                    />

                    <div v-if="formModel.suggestionMediaList.length > 0" class="mt-14px">
                      <RehabRecordMediaUpload
                        :model-value="formModel.suggestionMediaList"
                        disabled
                      />
                    </div>
                  </div>
                </div>
              </a-col>
            </a-row>

            <div class="rounded-14px border border-solid border-#e9edf5 bg-white p-18px">
              <div class="text-15px leading-22px font-600 text-#222">
                家长意见反馈
              </div>

              <div class="mt-14px">
                <a-textarea
                  :value="formModel.parentFeedback || '未填写'"
                  disabled
                  :auto-size="{ minRows: 4, maxRows: 6 }"
                  placeholder="请输入家长意见反馈"
                />
              </div>

              <a-row :gutter="[12, 12]" class="mt-16px">
                <a-col :xs="24" :lg="12">
                  <div class="rounded-12px border border-solid border-#edf0f5 bg-#fafbfc p-14px">
                    <div class="parent-signature-row">
                      <div class="text-12px leading-18px text-#8c8c8c">家长签名</div>
                      <div v-if="signatureImageSrc" class="parent-signature-preview">
                        <a-image
                          :src="signatureImageSrc"
                          class="parent-signature-image"
                        />
                      </div>
                      <span v-else class="text-14px leading-22px text-#262626">
                        {{ formModel.parentSignature || '未填写' }}
                      </span>
                    </div>
                  </div>
                </a-col>
                <a-col :xs="24" :lg="12">
                  <div class="rounded-12px border border-solid border-#edf0f5 bg-#fafbfc p-14px">
                    <div class="mb-10px text-12px leading-18px text-#8c8c8c">反馈时间</div>
                    <a-input
                      :value="formModel.feedbackDate || '未填写'"
                      disabled
                    />
                  </div>
                </a-col>
              </a-row>
            </div>
          </div>

          <div v-else class="flex flex-col gap-16px">
            <div class="rounded-14px border border-solid border-#e9edf5 bg-white p-18px">
              <div class="flex items-center justify-between gap-12px">
                <div class="text-15px leading-22px font-600 text-#222">
                  基础信息
                </div>
                <a-button
                  type="primary"
                  ghost
                  :disabled="!canReusePreviousPublished"
                  @click="handleReusePreviousPublished"
                >
                  复用上一篇
                </a-button>
              </div>

              <a-row :gutter="[12, 12]" class="mt-14px">
                <a-col :xs="24" :sm="12" :lg="8">
                  <div class="rounded-12px border border-solid border-#edf0f5 bg-#fafbfc p-14px">
                    <div class="mb-10px text-12px leading-18px text-#8c8c8c">姓名</div>
                    <a-input v-model:value="formModel.studentName" placeholder="请输入学员姓名" />
                  </div>
                </a-col>
                <a-col :xs="24" :sm="12" :lg="8">
                  <div class="rounded-12px border border-solid border-#edf0f5 bg-#fafbfc p-14px">
                    <div class="mb-10px text-12px leading-18px text-#8c8c8c">性别</div>
                    <a-select
                      v-model:value="formModel.gender"
                      :options="genderOptions"
                      allow-clear
                      style="width: 100%;"
                      placeholder="请选择性别"
                    />
                  </div>
                </a-col>
                <a-col :xs="24" :sm="12" :lg="8">
                  <div class="rounded-12px border border-solid border-#edf0f5 bg-#fafbfc p-14px">
                    <div class="mb-10px text-12px leading-18px text-#8c8c8c">出生年月</div>
                    <a-date-picker
                      v-model:value="formModel.birthDate"
                      value-format="YYYY-MM-DD"
                      style="width: 100%;"
                      placeholder="请选择出生年月"
                    />
                  </div>
                </a-col>
                <a-col :xs="24" :sm="12" :lg="8">
                  <div class="rounded-12px border border-solid border-#edf0f5 bg-#fafbfc p-14px">
                    <div class="mb-10px text-12px leading-18px text-#8c8c8c">班别</div>
                    <a-input v-model:value="formModel.className" placeholder="请输入班别" />
                  </div>
                </a-col>
                <a-col :xs="24" :sm="12" :lg="8">
                  <div class="rounded-12px border border-solid border-#edf0f5 bg-#fafbfc p-14px">
                    <div class="mb-10px text-12px leading-18px text-#8c8c8c">任教老师</div>
                    <a-input v-model:value="formModel.teacherName" placeholder="请输入任教老师" />
                  </div>
                </a-col>
                <a-col :xs="24" :sm="12" :lg="8">
                  <div class="rounded-12px border border-solid border-#edf0f5 bg-#fafbfc p-14px">
                    <div class="mb-10px text-12px leading-18px text-#8c8c8c">训练日期</div>
                    <a-date-picker
                      v-model:value="formModel.trainingDate"
                      value-format="YYYY-MM-DD"
                      style="width: 100%;"
                      placeholder="请选择训练日期"
                    />
                  </div>
                </a-col>
              </a-row>
            </div>

            <div class="rounded-14px border border-solid border-#e9edf5 bg-white p-18px">
              <div class="text-15px leading-22px font-600 text-#222">
                训练目标
              </div>
              <div class="mt-14px">
                <a-textarea
                  v-model:value="formModel.trainingTarget"
                  :auto-size="{ minRows: 4, maxRows: 6 }"
                  placeholder="请输入本次训练目标"
                />
              </div>
            </div>

            <div class="rounded-14px border border-solid border-#e9edf5 bg-white p-18px">
              <div class="flex items-center justify-between gap-12px">
                <div class="text-15px leading-22px font-600 text-#222">
                  训练项目
                </div>
                <a-button type="dashed" size="small" @click="handleAddTrainingModule">
                  新增项目
                </a-button>
              </div>

              <a-row :gutter="[12, 12]" class="mt-14px">
                <a-col
                  v-for="(item, index) in trainingModules"
                  :key="item.id"
                  :xs="24"
                  :lg="getTrainingModuleColSpan(index)"
                >
                  <div class="h-full rounded-12px border border-solid border-#edf0f5 bg-#fafbfc p-16px">
                    <div class="flex items-center justify-between gap-12px mb-12px">
                      <a-input
                        v-model:value="item.title"
                        class="flex-1"
                        placeholder="请输入训练项目名称"
                      />
                      <a-button
                        v-if="trainingModules.length > 1"
                        type="link"
                        danger
                        class="px-0"
                        @click="handleRemoveTrainingModule(item.id)"
                      >
                        删除
                      </a-button>
                    </div>

                    <a-textarea
                      v-model:value="item.content"
                      :auto-size="{ minRows: 5, maxRows: 8 }"
                      placeholder="请输入训练项目内容"
                    />
                  </div>
                </a-col>
              </a-row>

              <div class="mt-14px">
                <RehabRecordMediaUpload v-model="formModel.trainingMediaList" />
              </div>
            </div>

            <a-row :gutter="[16, 16]">
              <a-col :xs="24" :lg="12">
                <div class="h-full rounded-14px border border-solid border-#e9edf5 bg-white p-18px">
                  <div class="text-15px leading-22px font-600 text-#222">
                    学生综合表现
                  </div>
                  <div class="mt-14px">
                    <a-textarea
                      v-model:value="formModel.performance"
                      :auto-size="{ minRows: 7, maxRows: 10 }"
                      placeholder="请输入学生综合表现"
                    />

                    <div class="mt-14px">
                      <RehabRecordMediaUpload v-model="formModel.performanceMediaList" />
                    </div>
                  </div>
                </div>
              </a-col>

              <a-col :xs="24" :lg="12">
                <div class="h-full rounded-14px border border-solid border-#e9edf5 bg-white p-18px">
                  <div class="text-15px leading-22px font-600 text-#222">
                    康复建议
                  </div>
                  <div class="mt-14px">
                    <a-textarea
                      v-model:value="formModel.suggestion"
                      :auto-size="{ minRows: 7, maxRows: 10 }"
                      placeholder="请输入康复建议"
                    />

                    <div class="mt-14px">
                      <RehabRecordMediaUpload v-model="formModel.suggestionMediaList" />
                    </div>
                  </div>
                </div>
              </a-col>
            </a-row>

          </div>
        </div>
      </a-spin>

      <div class="flex justify-end px-20px pt-12px pb-16px bg-white border-t border-solid border-#eef0f5">
        <a-space>
          <a-button :disabled="Boolean(submittingAction)" @click="open = false">
            {{ isReadonly ? '关闭' : '取消' }}
          </a-button>
          <template v-if="!isReadonly">
            <a-button
              :loading="submittingAction === 'draft'"
              :disabled="submittingAction === 'publish'"
              @click="handleSaveDraft"
            >
              保存草稿
            </a-button>
            <a-button
              type="primary"
              :loading="submittingAction === 'publish'"
              :disabled="submittingAction === 'draft'"
              @click="handlePublish"
            >
              {{ publishButtonText }}
            </a-button>
          </template>
        </a-space>
      </div>
    </div>
  </a-drawer>
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

.parent-signature-preview {
  display: flex;
  align-items: center;
  min-height: 45px;
}

.parent-signature-row {
  display: flex;
  align-items: center;
  justify-content: flex-start;
  gap: 16px;
  min-height: 60px;
}

:deep(.parent-signature-image) {
  max-width: 110px;
}

:deep(.parent-signature-image .ant-image-img) {
  width: auto;
  max-width: 100%;
  max-height: 45px;
  object-fit: contain;
}
</style>
