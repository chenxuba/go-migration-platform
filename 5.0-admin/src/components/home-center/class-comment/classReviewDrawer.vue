<script setup lang="ts">
import { CloseOutlined, InfoCircleOutlined } from '@ant-design/icons-vue'
import { Empty } from 'ant-design-vue'
import dayjs from 'dayjs'
import { computed, ref, watch } from 'vue'
import scheduleClassImage from '@/assets/images/timetable/schedule-class.png'
import scheduleOneToOneImage from '@/assets/images/timetable/schedule-one2one.png'
import type { ClassCommentStudentItem } from '@/api/edu-center/class-record'
import messageService from '@/utils/messageService'

interface ReviewStudentItem {
  id: string
  name: string
  avatar: string
  attentionText: string
  status: '到课' | '请假' | '旷课' | '未记录'
  readText?: string
}

const props = withDefaults(defineProps<{
  type?: string | number
  record?: Partial<ClassCommentStudentItem> | null
}>(), {
  type: '1',
  record: null,
})

const open = defineModel<boolean>({
  default: false,
})

const activeKey = ref('1')
const editContentModalOpen = ref(false)
const courseContent = ref('感统器材抓取与追视训练')
const editingCourseContent = ref(courseContent.value)
const selectedRowKeys = ref<string[]>([])

const defaultAvatar = 'https://cdn.schoolpal.cn/schoolpal/next-erp/avator_male.png?x-oss-process=image/resize,w_120'
const simpleImage = Empty.PRESENTED_IMAGE_SIMPLE

const sourceCover = computed(() => Number(props.record?.sourceType || 0) === 2 ? scheduleOneToOneImage : scheduleClassImage)
const headerTitle = computed(() => {
  return String(props.record?.sourceName || props.record?.lessonName || '班级感统课1班').trim() || '班级感统课1班'
})
const courseName = computed(() => {
  return String(props.record?.lessonName || '班级感统课').trim() || '班级感统课'
})
const mainTeacherText = computed(() => {
  return String(props.record?.teacherName || '丁海星').trim() || '丁海星'
})
const assistantTeacherText = computed(() => {
  const text = String(props.record?.assistants || '').trim()
  return text || '-'
})
const classRoomText = computed(() => {
  const text = String(props.record?.classRoomName || '').trim()
  return text || '-'
})

const reviewedStudents = computed<ReviewStudentItem[]>(() => [])

const pendingStudents = computed<ReviewStudentItem[]>(() => {
  const currentName = String(props.record?.studentName || '').trim()
  const currentAvatar = String(props.record?.avatar || '').trim() || defaultAvatar
  const base: ReviewStudentItem[] = [
    {
      id: 'focus-student',
      name: currentName || '刘金翠',
      avatar: currentAvatar,
      attentionText: '未关注',
      status: '到课',
    },
    {
      id: 'pending-student-2',
      name: currentName && currentName !== '张一鸣' ? '张一鸣' : '刘金翠',
      avatar: defaultAvatar,
      attentionText: '未关注',
      status: '到课',
    },
  ]
  return base.filter((item, index, array) => array.findIndex(target => target.name === item.name) === index)
})

const commentStatisticsText = computed(() => `${reviewedStudents.value.length}/${pendingStudents.value.length}`)
const pendingSummaryText = computed(() => `共 ${pendingStudents.value.length} 位待点评学员`)

const headerTimeText = computed(() => {
  const start = dayjs(props.record?.startTime)
  const end = dayjs(props.record?.endTime)
  if (!start.isValid() || !end.isValid())
    return '2026-04-16（周四）12:00 ~ 13:00'
  const weekMap = ['周日', '周一', '周二', '周三', '周四', '周五', '周六']
  return `${start.format('YYYY-MM-DD')}（${weekMap[start.day()] || '周一'}）${start.format('HH:mm')} ~ ${end.format('HH:mm')}`
})

const classDurationText = computed(() => {
  const start = dayjs(props.record?.startTime)
  const end = dayjs(props.record?.endTime)
  if (!start.isValid() || !end.isValid())
    return '60分钟'
  return `${Math.max(end.diff(start, 'minute'), 0)}分钟`
})

const rowSelection = computed(() => ({
  selectedRowKeys: selectedRowKeys.value,
  onChange: (keys: (string | number)[]) => {
    selectedRowKeys.value = keys.map(item => String(item))
  },
}))

const pendingColumns = [
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
]

function handlePendingReview() {
  messageService.info('暂未开发')
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
  () => [open.value, props.type] as const,
  ([drawerOpen, tabType]) => {
    if (!drawerOpen)
      return
    activeKey.value = String(tabType ?? '1')
    selectedRowKeys.value = []
  },
  { immediate: true },
)
</script>

<template>
  <div>
    <a-drawer
      v-model:open="open"
      :body-style="{ padding: '0', background: '#f7f7fd' }"
      width="1165px"
    >
      <template #title>
        <div class="custom-header flex justify-between h-4 flex-items-center">
          <div class="text-5">
            课堂点评详情
          </div>
          <a-button type="text" class="close-btn" @click="open = false">
            <template #icon>
              <CloseOutlined class="text-5 close-icon" />
            </template>
          </a-button>
        </div>
      </template>

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
          <a-descriptions-item label="点评统计">
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
          <a-tab-pane :key="'0'" :tab="`已点评（${reviewedStudents.length}）`">
            <div class="p-12px">
              <div class="bg-white rounded-15px p-20px">
                <a-empty :image="simpleImage" description="暂无已点评学员" />
              </div>
            </div>
          </a-tab-pane>
          <a-tab-pane :key="'1'" :tab="`待点评（${pendingStudents.length}）`">
            <div class="p-12px">
              <div class="bg-white rounded-15px p-20px">
                <div class="flex justify-between items-center mb-16px">
                  <div class="text-15px text-#222">
                    {{ pendingSummaryText }}
                  </div>
                  <a-tooltip>
                    <template #title>
                      未勾选学员时，直接点击批量点评将会默认选中到课学员
                    </template>
                    <a-button type="primary" class="px-14px" @click="handleBatchReview">
                      <InfoCircleOutlined />
                      批量点评
                    </a-button>
                  </a-tooltip>
                </div>
                <a-table
                  row-key="id"
                  :pagination="false"
                  :row-selection="rowSelection"
                  :columns="pendingColumns"
                  :data-source="pendingStudents"
                  size="small"
                >
                  <template #bodyCell="{ column, record }">
                    <template v-if="column.key === 'name'">
                      <div class="flex items-center">
                        <img width="36" height="36" class="mr-8px rounded-100" :src="record.avatar" alt="">
                        <div class="flex items-center flex-wrap">
                          <span class="text-#222 mr-6px">{{ record.name }}</span>
                          <a-tag :bordered="false" color="default">
                            {{ record.attentionText }}
                          </a-tag>
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
                      <a-button type="link" class="text-14px text-#06f px-0" @click="handlePendingReview">
                        去点评
                      </a-button>
                    </template>
                  </template>
                </a-table>
              </div>
            </div>
          </a-tab-pane>
        </a-tabs>
      </div>
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

  :deep(.ant-tabs-nav) {
    background: #fff;
    margin: 0;
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
</style>
