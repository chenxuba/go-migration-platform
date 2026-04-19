<script setup lang="ts">
import { CloseOutlined, InfoCircleOutlined } from '@ant-design/icons-vue'
import { Empty } from 'ant-design-vue'
import dayjs from 'dayjs'
import { computed, ref, watch } from 'vue'
import scheduleClassImage from '@/assets/images/timetable/schedule-class.png'
import scheduleOneToOneImage from '@/assets/images/timetable/schedule-one2one.png'
import { getTeachingRecordDetailApi, type TeachingRecordDetailResult, type TeachingRecordDetailStudent, type TeachingRecordDetailTeacher } from '@/api/edu-center/class-record'
import RehabRecordEditorDrawer from '@/components/home-center/class-comment/rehab-record-editor-drawer.vue'
import messageService from '@/utils/messageService'

interface ReviewStudentItem {
  id: string
  studentTeachingRecordId: string
  name: string
  avatar: string
  attentionText: string
  status: '到课' | '请假' | '旷课' | '未记录'
  sex?: number
  birthday?: string
  readText?: string
}

interface ReviewDrawerRecord {
  teachingRecordId?: string
  sourceType?: string | number
  sourceName?: string
  lessonName?: string
  teacherName?: string
  assistants?: string
  classRoomName?: string
  startTime?: string
  endTime?: string
  studentName?: string
  avatar?: string
}

const props = withDefaults(defineProps<{
  type?: string | number
  record?: Partial<ReviewDrawerRecord> | null
}>(), {
  type: '1',
  record: null,
})

const emit = defineEmits<{
  (e: 'updated'): void
}>()

const open = defineModel<boolean>({
  default: false,
})

const loading = ref(false)
const detailData = ref<TeachingRecordDetailResult | null>(null)
const activeKey = ref('1')
const editContentModalOpen = ref(false)
const editorDrawerOpen = ref(false)
const editorMode = ref<'create' | 'view' | 'edit'>('create')
const courseContent = ref('')
const editingCourseContent = ref(courseContent.value)
const selectedRowKeys = ref<string[]>([])
const statusFilteredValues = ref<string[]>(['到课'])
const currentEditingStudent = ref<ReviewStudentItem | null>(null)

const defaultAvatar = 'https://cdn.schoolpal.cn/schoolpal/next-erp/avator_male.png?x-oss-process=image/resize,w_120'
const simpleImage = Empty.PRESENTED_IMAGE_SIMPLE

const currentTeachingRecordId = computed(() => String(props.record?.teachingRecordId || '').trim())
const teacherList = computed(() => Array.isArray(detailData.value?.teacherList) ? detailData.value?.teacherList || [] : [])
const studentList = computed(() => Array.isArray(detailData.value?.studentList) ? detailData.value?.studentList || [] : [])

const sourceCover = computed(() => Number(detailData.value?.timetableSourceType || detailData.value?.sourceType || props.record?.sourceType || 0) === 2 ? scheduleOneToOneImage : scheduleClassImage)
const headerTitle = computed(() => {
  return String(detailData.value?.sourceName || detailData.value?.lessonName || props.record?.sourceName || props.record?.lessonName || '班级感统课1班').trim() || '班级感统课1班'
})
const courseName = computed(() => {
  return String(detailData.value?.lessonName || props.record?.lessonName || '班级感统课').trim() || '班级感统课'
})
const mainTeacherText = computed(() => {
  const names = formatTeacherNames(teacherList.value.filter(item => Number(item.type || 0) === 1))
  if (names !== '-')
    return names
  return String(props.record?.teacherName || '丁海星').trim() || '丁海星'
})
const assistantTeacherText = computed(() => {
  const names = formatTeacherNames(teacherList.value.filter(item => Number(item.type || 0) !== 1))
  if (names !== '-')
    return names
  const text = String(props.record?.assistants || '').trim()
  return text || '-'
})
const classRoomText = computed(() => {
  const text = String(detailData.value?.classRoomName || props.record?.classRoomName || '').trim()
  return text || '-'
})

const reviewedStudents = computed<ReviewStudentItem[]>(() => {
  return studentList.value
    .filter(item => hasComment(item))
    .map(mapReviewStudent)
})

const pendingStudents = computed<ReviewStudentItem[]>(() => {
  return studentList.value
    .filter(item => !hasComment(item))
    .map(mapReviewStudent)
})

const commentStatisticsText = computed(() => `${reviewedStudents.value.length}/${studentList.value.length}`)
const pendingSummaryText = computed(() => `共 ${pendingStudents.value.length} 位待记录学员`)
const reviewedSummaryText = computed(() => `共 ${reviewedStudents.value.length} 位已记录学员`)

const headerTimeText = computed(() => {
  const start = dayjs(detailData.value?.startTime || props.record?.startTime)
  const end = dayjs(detailData.value?.endTime || props.record?.endTime)
  if (!start.isValid() || !end.isValid())
    return '2026-04-16（周四）12:00 ~ 13:00'
  const weekMap = ['周日', '周一', '周二', '周三', '周四', '周五', '周六']
  return `${start.format('YYYY-MM-DD')}（${weekMap[start.day()] || '周一'}）${start.format('HH:mm')} ~ ${end.format('HH:mm')}`
})

const classDurationText = computed(() => {
  const start = dayjs(detailData.value?.startTime || props.record?.startTime)
  const end = dayjs(detailData.value?.endTime || props.record?.endTime)
  if (!start.isValid() || !end.isValid())
    return '60分钟'
  return `${Math.max(end.diff(start, 'minute'), 0)}分钟`
})
const editorSession = computed(() => ({
  sourceName: headerTitle.value,
  lessonName: courseName.value,
  teacherName: mainTeacherText.value,
  classRoomName: classRoomText.value,
  startTime: detailData.value?.startTime || props.record?.startTime,
  endTime: detailData.value?.endTime || props.record?.endTime,
}))

const reviewedColumns = [
  {
    title: '学员/性别',
    dataIndex: 'name',
    key: 'name',
    width: 520,
  },
  {
    title: '上课状态',
    dataIndex: 'status',
    key: 'status',
    width: 260,
  },
  {
    title: '操作',
    key: 'action',
    width: 180,
  },
]

const rowSelection = computed(() => ({
  selectedRowKeys: selectedRowKeys.value,
  onChange: (keys: (string | number)[]) => {
    selectedRowKeys.value = keys.map(item => String(item))
  },
}))

const pendingColumns = computed(() => [
  {
    title: '学员/性别',
    dataIndex: 'name',
    key: 'name',
    width: 520,
  },
  {
    title: '上课状态',
    dataIndex: 'status',
    key: 'status',
    width: 260,
    filteredValue: statusFilteredValues.value,
    filters: [
      { text: '到课', value: '到课' },
      { text: '请假', value: '请假' },
      { text: '旷课', value: '旷课' },
      { text: '未记录', value: '未记录' },
    ],
    onFilter: (value: string | number | boolean, record: ReviewStudentItem) => record.status === value,
  },
  {
    title: '操作',
    key: 'action',
    width: 180,
  },
])

function handlePendingTableChange(_: unknown, filters: Record<string, (string | number | boolean)[] | null>) {
  const nextValues = Array.isArray(filters?.status)
    ? filters.status.map(item => String(item)).filter(Boolean)
    : []
  statusFilteredValues.value = nextValues
}

function formatTeacherNames(list: Array<Partial<TeachingRecordDetailTeacher>>) {
  const names = list.map(item => String(item.teacherName || '').trim()).filter(Boolean)
  return names.length ? names.join('、') : '-'
}

function resolveStudentStatus(status?: number): ReviewStudentItem['status'] {
  const currentStatus = Number(status || 0)
  if (currentStatus === 2)
    return '旷课'
  if (currentStatus === 3)
    return '请假'
  if (currentStatus === 4)
    return '未记录'
  return '到课'
}

function hasComment(student: Partial<TeachingRecordDetailStudent>) {
  if (student.isComment === true)
    return true
  return String(student.externalRemark || '').trim() !== ''
}

function mapReviewStudent(student: Partial<TeachingRecordDetailStudent>): ReviewStudentItem {
  const studentTeachingRecordId = String(student.studentTeachingRecordId || '').trim()
  return {
    id: studentTeachingRecordId || String(student.studentId || student.studentName || Math.random()).trim(),
    studentTeachingRecordId,
    name: String(student.studentName || '-').trim() || '-',
    avatar: String(student.avatar || '').trim() || defaultAvatar,
    attentionText: '未关注',
    status: resolveStudentStatus(student.status),
    sex: typeof student.sex === 'number' ? student.sex : undefined,
    birthday: String(student.birthday || '').trim() || undefined,
  }
}

function syncCourseContent() {
  const text = String(detailData.value?.teachingContent || '').trim()
  courseContent.value = text
  editingCourseContent.value = text
}

async function loadDetail() {
  const teachingRecordId = currentTeachingRecordId.value
  if (!open.value || !teachingRecordId) {
    detailData.value = null
    courseContent.value = ''
    editingCourseContent.value = ''
    return
  }

  loading.value = true
  try {
    const res = await getTeachingRecordDetailApi({ teachingRecordId })
    if (res.code !== 200)
      throw new Error(res.message || '加载康复记录详情失败')
    const data = res.result
    detailData.value = data && String(data.teachingRecordId || '').trim() ? data : null
    syncCourseContent()
  }
  catch (error: any) {
    detailData.value = null
    courseContent.value = ''
    editingCourseContent.value = ''
    messageService.error(error?.response?.data?.message || error?.message || '加载康复记录详情失败')
  }
  finally {
    loading.value = false
  }
}

function handlePendingReview(student: ReviewStudentItem) {
  editorMode.value = 'create'
  currentEditingStudent.value = student
  editorDrawerOpen.value = true
}

function handleViewReviewed(student: ReviewStudentItem) {
  editorMode.value = 'view'
  currentEditingStudent.value = student
  editorDrawerOpen.value = true
}

function handleEditReviewed(student: ReviewStudentItem) {
  editorMode.value = 'edit'
  currentEditingStudent.value = student
  editorDrawerOpen.value = true
}

async function handleEditorPublished() {
  currentEditingStudent.value = null
  await loadDetail()
  emit('updated')
}

function handleBatchReview() {
  messageService.info('暂未开发')
}

function handleOpenEditContent() {
  editingCourseContent.value = courseContent.value
  editContentModalOpen.value = true
}

function handleSaveContent() {
  courseContent.value = editingCourseContent.value
  editContentModalOpen.value = false
}

watch(
  () => `${open.value}|${props.type}|${currentTeachingRecordId.value}`,
  async () => {
    if (!open.value) {
      detailData.value = null
      loading.value = false
      courseContent.value = ''
      editingCourseContent.value = ''
      editorDrawerOpen.value = false
      editorMode.value = 'create'
      currentEditingStudent.value = null
      activeKey.value = '1'
      selectedRowKeys.value = []
      statusFilteredValues.value = ['到课']
      return
    }
    activeKey.value = String(props.type ?? '1')
    selectedRowKeys.value = []
    statusFilteredValues.value = ['到课']
    await loadDetail()
  },
  { immediate: true },
)
</script>

<template>
  <div>
    <a-drawer
      v-model:open="open"
      :body-style="{ padding: '0', background: '#f7f7fd', display: 'flex', 'flex-direction': 'column' }"
      :closable='false'
      width="1165px"
    >
      <template #title>
        <div class="custom-header flex justify-between h-4 flex-items-center">
          <div class="text-5">
            康复记录详情
          </div>
          <a-button type="text" class="close-btn" @click="open = false">
            <template #icon>
              <CloseOutlined class="text-5 close-icon" />
            </template>
          </a-button>
        </div>
      </template>

      <a-spin class="review-drawer-spin" :spinning="loading">
        <div class="review-drawer-content">
        <div class="contenter flex flex-center bg-white px6 py3">
          <div class="avatarBox w-16 h-16 relative">
            <img width="64" height="64" class="rounded-100" :src="sourceCover" alt="">
          </div>
          <div class="info flex flex-1 ml-4 flex-col">
            <div class="top flex justify-between flex-center flex-1">
              <a-space>
                <div class="name text-5 font-800">
                  {{ headerTitle }}
                </div>
              </a-space>
            </div>
            <div class="bottom flex-1 flex flex-items-center mt-2">
              <div class="birthday flex-center">
                <span class="text-4 text-#222">{{ headerTimeText }}</span>
                <span class="bg-#e6f0ff text-#06f text-3 px2 py1 rounded-10 ml2">{{ classDurationText }}</span>
              </div>
            </div>
          </div>
        </div>

        <div class="desc pt-4 bg-white px6 py3 pb0">
          <a-descriptions :column="3" size="small" :content-style="{ color: '#888' }">
            <a-descriptions-item label="上课教师">
              {{ mainTeacherText }}
            </a-descriptions-item>
            <a-descriptions-item label="上课助教">
              {{ assistantTeacherText }}
            </a-descriptions-item>
            <a-descriptions-item label="上课教室">
              {{ classRoomText }}
            </a-descriptions-item>
            <a-descriptions-item label="所属课程">
              {{ courseName }}
            </a-descriptions-item>
            <a-descriptions-item label="记录统计">
              {{ commentStatisticsText }}
            </a-descriptions-item>
            <a-descriptions-item label="上课内容">
              <div class="flex items-start gap-8px">
                <span>{{ courseContent || '-' }}</span>
                <span class="text-#06f cursor-pointer" @click="handleOpenEditContent">编辑</span>
              </div>
            </a-descriptions-item>
          </a-descriptions>
        </div>

        <div class="tabs">
          <a-tabs
            v-model:active-key="activeKey"
            size="large"
            :tab-bar-style="{ 'border-radius': '0px', 'padding-left': '24px' }"
          >
            <a-tab-pane :key="'0'" :tab="`已记录（${reviewedStudents.length}）`">
              <div class="tab-pane-wrap p-12px">
                <div class="tab-pane-card bg-white rounded-15px px-20px pt-12px" :class="{ 'tab-pane-card--empty': reviewedStudents.length === 0 }">
                  <template v-if="reviewedStudents.length > 0">
                    <custom-title class="mb-16px" :title="reviewedSummaryText" font-size="14px" />
                    <a-table
                      row-key="id"
                      :pagination="false"
                      :columns="reviewedColumns"
                      :data-source="reviewedStudents"
                      size="small"
                    >
                      <template #bodyCell="{ column, record }">
                        <template v-if="column.key === 'name'">
                          <div class="flex items-center">
                            <img width="36" height="36" class="mr-8px rounded-100" :src="record.avatar" alt="">
                            <div class="flex items-center flex-wrap">
                              <span class="text-#222 mr-6px">{{ record.name }}</span>
                              <span class="inline-flex h-24px items-center rounded-10px bg-#eee px-8px text-12px text-#888">
                                {{ record.attentionText }}
                              </span>
                            </div>
                          </div>
                        </template>
                        <template v-else-if="column.key === 'status'">
                          <a-tag
                            :bordered="false"
                            :color="record.status === '到课' ? 'processing' : 'default'"
                          >
                            {{ record.status }}
                          </a-tag>
                        </template>
                        <template v-else-if="column.key === 'action'">
                          <a-space :size="14">
                            <a-button type="link" class="text-14px text-#06f px-0" @click="handleViewReviewed(record as ReviewStudentItem)">
                              查看
                            </a-button>
                            <a-button type="link" class="text-14px text-#06f px-0" @click="handleEditReviewed(record as ReviewStudentItem)">
                              编辑
                            </a-button>
                          </a-space>
                        </template>
                      </template>
                    </a-table>
                  </template>
                  <a-empty v-else :image="simpleImage" description="暂无已记录学员" />
                </div>
              </div>
            </a-tab-pane>
            <a-tab-pane :key="'1'" :tab="`待记录（${pendingStudents.length}）`">
              <div class="tab-pane-wrap p-12px">
                <div class="tab-pane-card bg-white rounded-15px px-20px pt-12px">
                  <custom-title class="mb-16px" :title="pendingSummaryText" font-size="14px">
                    <template #right>
                      <a-tooltip>
                        <template #title>
                          未勾选学员时，直接点击批量记录将会默认选中到课学员
                        </template>
                        <a-button type="primary" class="px-14px" @click="handleBatchReview">
                          <InfoCircleOutlined />
                          批量记录
                        </a-button>
                      </a-tooltip>
                    </template>
                  </custom-title>
                  <a-table
                    row-key="id"
                    :pagination="false"
                    :row-selection="rowSelection"
                    :columns="pendingColumns"
                    :data-source="pendingStudents"
                    size="small"
                    @change="handlePendingTableChange"
                  >
                    <template #bodyCell="{ column, record }">
                      <template v-if="column.key === 'name'">
                        <div class="flex items-center">
                          <img width="36" height="36" class="mr-8px rounded-100" :src="record.avatar" alt="">
                          <div class="flex items-center flex-wrap">
                            <span class="text-#222 mr-6px">{{ record.name }}</span>
                            <span class="inline-flex h-24px items-center rounded-10px bg-#eee px-8px text-12px text-#888">
                              {{ record.attentionText }}
                            </span>
                          </div>
                        </div>
                      </template>
                      <template v-else-if="column.key === 'status'">
                        <a-tag
                          :bordered="false"
                          :color="record.status === '到课' ? 'processing' : 'default'"
                        >
                          {{ record.status }}
                        </a-tag>
                      </template>
                      <template v-else-if="column.key === 'action'">
                        <a-button type="link" class="text-14px text-#06f px-0" @click="handlePendingReview(record as ReviewStudentItem)">
                          去记录
                        </a-button>
                      </template>
                    </template>
                  </a-table>
                </div>
              </div>
            </a-tab-pane>
          </a-tabs>
        </div>
        </div>
      </a-spin>
    </a-drawer>

    <a-modal
      v-model:open="editContentModalOpen"
      :mask-closable="false"
      :keyboard="false"
      width="550px"
      @ok="handleSaveContent"
      @cancel="editingCourseContent = courseContent"
    >
      <template #title>
        <span class="font-400 text-15px">编辑上课内容</span>
      </template>
      <div class="p-10px flex items-start">
        <span class="text-#888">上课内容：</span>
        <a-textarea
          v-model:value="editingCourseContent"
          placeholder="选填（1000字以内）"
          style="flex:1;height: 120px;"
          show-count
          :maxlength="1000"
        />
      </div>
    </a-modal>

    <RehabRecordEditorDrawer
      v-model="editorDrawerOpen"
      :mode="editorMode"
      :student-teaching-record-id="currentEditingStudent?.studentTeachingRecordId"
      :student="currentEditingStudent"
      :session="editorSession"
      @published="handleEditorPublished"
    />
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

.tabs {
  width: 100%;
  border-radius: 10px;
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;

  :deep(.ant-tabs) {
    height: 100%;
    display: flex;
    flex-direction: column;
  }

  :deep(.ant-tabs-nav) {
    background: #fff;
    margin: 0;
  }

  :deep(.ant-tabs-content-holder) {
    flex: 1;
    min-height: 0;
  }

  :deep(.ant-tabs-content) {
    height: 100%;
  }

  :deep(.ant-tabs-tabpane-active) {
    height: 100%;
  }

  :deep(.ant-tabs-ink-bar) {
    text-align: center;
    height: 12px !important;
    background: transparent;
    bottom: 0 !important;

    &::after {
      position: absolute;
      top: 0;
      left: calc(50% - 12px);
      width: 24px !important;
      height: 4px !important;
      border-radius: 2px;
      background-color: var(--pro-ant-color-primary);
      content: "";
    }
  }
}

.review-drawer-spin {
  height: 100%;

  :deep(.ant-spin-container) {
    height: 100%;
    display: flex;
    flex-direction: column;
    min-height: 0;
  }
}

.review-drawer-content {
  height: 100%;
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.tab-pane-wrap {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.tab-pane-card {
  flex: 1;
}

.tab-pane-card--empty {
  display: flex;
  align-items: center;
  justify-content: center;
}
</style>
