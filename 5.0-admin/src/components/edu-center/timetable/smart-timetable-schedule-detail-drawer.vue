<script setup lang="ts">
import { CloseOutlined } from '@ant-design/icons-vue'
import dayjs from 'dayjs'
import type { TableColumnsType } from 'ant-design-vue'
import { computed, ref, watch } from 'vue'
import scheduleClassImage from '@/assets/images/timetable/schedule-class.png'
import scheduleOneToOneImage from '@/assets/images/timetable/schedule-one2one.png'
import { getTeachingScheduleDetailApi, type TeachingScheduleDetail, type TeachingScheduleDetailStudent } from '@/api/edu-center/teaching-schedule'
import { useStudentStore } from '@/stores/student'
import messageService from '@/utils/messageService'

type DrawerSummary = {
  scheduleId?: string
  id?: string
  lessonTitle?: string
  courseName?: string
  teacherName?: string
  assistantText?: string
  classroomName?: string
  studentText?: string
  courseType?: number
}

const props = defineProps<{
  open: boolean
  detail?: DrawerSummary | null
  deleting?: boolean
}>()

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'delete'): void
}>()

const openDrawer = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

const loading = ref(false)
const detailData = ref<TeachingScheduleDetail | null>(null)
const studentStore = useStudentStore()
const openStudentDrawer = ref(false)
const defaultStudentAvatar = 'https://pcsys.admin.ybc365.com/a369a751-2be5-4929-974d-9ae4439f54c4.png'
const studentColumns: TableColumnsType<TeachingScheduleDetailStudent> = [
  {
    title: '学员姓名',
    dataIndex: 'studentName',
    key: 'studentName',
    width: 260,
  },
  {
    title: '联系电话',
    dataIndex: 'phone',
    key: 'phone',
    width: 220,
  },
  {
    title: '状态',
    key: 'studentType',
    width: 120,
  },
  {
    title: '操作',
    key: 'action',
    width: 100,
    align: 'center',
  },
]

const repeatRuleLabelMap: Record<string, string> = {
  none: '不重复',
  weekly: '每周重复',
  biweekly: '隔周重复',
  daily: '每天重复',
  alternateDay: '隔天重复',
}

const scheduleId = computed(() => String(props.detail?.scheduleId || props.detail?.id || '').trim())
const isOneToOne = computed(() => {
  if (detailData.value)
    return Number(detailData.value.classType) === 2
  return Number(props.detail?.courseType) === 1
})
const scheduleCover = computed(() => (isOneToOne.value ? scheduleOneToOneImage : scheduleClassImage))
const students = computed(() => detailData.value?.students || [])
const headerTitle = computed(() => {
  if (props.detail?.lessonTitle)
    return props.detail.lessonTitle
  return detailData.value?.lessonName || detailData.value?.teachingClassName || '日程详情'
})
const timeText = computed(() => {
  if (!detailData.value)
    return '-'
  const dateText = dayjs(detailData.value.lessonDate).format('YYYY-MM-DD')
  const weekText = formatWeek(detailData.value.lessonDate)
  const startTime = dayjs(detailData.value.startAt).format('HH:mm')
  const endTime = dayjs(detailData.value.endAt).format('HH:mm')
  return `${dateText}(${weekText}) ${startTime} ~ ${endTime}`
})
const durationText = computed(() => {
  const minutes = Number(detailData.value?.durationMinutes || 0)
  return minutes > 0 ? `${minutes}分钟` : '-'
})
const assistantText = computed(() => {
  if (Array.isArray(detailData.value?.assistantNames) && detailData.value.assistantNames.length)
    return detailData.value.assistantNames.join('、')
  return props.detail?.assistantText || '-'
})
const repeatRuleText = computed(() => {
  const meta = detailData.value?.batchMeta
  if (!meta)
    return '不重复'
  if (meta.schedulingMode === 'free')
    return '自由排课'
  const base = repeatRuleLabelMap[String(meta.repeatRule || '').trim()] || String(meta.repeatRule || '').trim() || '不重复'
  if ((meta.repeatRule === 'weekly' || meta.repeatRule === 'biweekly') && Array.isArray(meta.selectedWeekdays) && meta.selectedWeekdays.length)
    return `${base} · ${meta.selectedWeekdays.join(' / ')}`
  return base
})
const remarkText = computed(() => detailData.value?.remark || '-')
const studentTypeText = computed(() => (Number(detailData.value?.classType) === 2 ? '1对1学员' : '班课学员'))

let loadSeq = 0

function formatWeek(date: string) {
  const day = dayjs(date).day()
  const weekMap: Record<number, string> = {
    0: '周日',
    1: '周一',
    2: '周二',
    3: '周三',
    4: '周四',
    5: '周五',
    6: '周六',
  }
  return weekMap[day] || '-'
}

function handleViewStudent(studentId?: string) {
  const id = String(studentId || '').trim()
  if (!id)
    return
  studentStore.setStudentId(id)
  openStudentDrawer.value = true
}

async function loadDetail() {
  const id = scheduleId.value
  if (!openDrawer.value || !id) {
    detailData.value = null
    return
  }

  const seq = ++loadSeq
  loading.value = true
  try {
    const res = await getTeachingScheduleDetailApi({ id })
    if (seq !== loadSeq)
      return
    if (res.code !== 200 || !res.result)
      throw new Error(res.message || '加载日程详情失败')
    detailData.value = res.result
  }
  catch (error: any) {
    if (seq !== loadSeq)
      return
    console.error('load teaching schedule detail failed', error)
    detailData.value = null
    messageService.error(error?.response?.data?.message || error?.message || '加载日程详情失败')
    openDrawer.value = false
  }
  finally {
    if (seq === loadSeq)
      loading.value = false
  }
}

watch(
  () => `${openDrawer.value}|${scheduleId.value}`,
  async () => {
    if (!openDrawer.value) {
      detailData.value = null
      loading.value = false
      return
    }
    await loadDetail()
  },
  { immediate: true },
)
</script>

<template>
  <a-drawer
    v-model:open="openDrawer"
    :push="{ distance: 80 }"
    :body-style="{ padding: '0', background: '#f7f7fd' }"
    :closable="false"
    width="1165px"
    placement="right"
  >
    <template #title>
      <div class="custom-header flex justify-between h-4 flex-items-center">
        <div class="text-5">
          日程详情
        </div>
        <a-button type="text" class="close-btn" @click="openDrawer = false">
          <template #icon>
            <CloseOutlined class="text-5 close-icon" />
          </template>
        </a-button>
      </div>
    </template>

    <a-spin :spinning="loading">
      <div class="contenter flex flex-center bg-white px6 py3">
        <div class="avatarBox w-16 h-16 relative">
          <img width="64" height="64" class="rounded-100" :src="scheduleCover" alt="">
        </div>
        <div class="info flex flex-1 ml-4 flex-col">
          <div class="top flex justify-between flex-center flex-1">
            <a-space>
              <div class="name text-5 font-800">
                {{ headerTitle }}
              </div>
             
            </a-space>
            <a-space>
              <a-button type="link" ghost>
                仅复制此日程
              </a-button>
              <a-button danger ghost :loading="deleting" @click="$emit('delete')">
                删除
              </a-button>
              <!-- 编辑 -->
              <a-button type="primary" >
                编辑
              </a-button>
              <!-- 去点名 -->
              <a-button type="primary" >
                去点名
              </a-button>
            </a-space>
          </div>
          <div class="bottom flex-1 flex flex-items-center mt-2">
            <div class="birthday flex-center">
              <span class="text-4 text-#222">{{ timeText }}</span>
              <span class="bg-#e6f0ff text-#06f text-3 px2 py1 rounded-10 ml2">{{ durationText }}</span>
            </div>
          </div>
        </div>
      </div>

      <div class="desc pt-4 bg-white px6 py3 pb0">
        <a-descriptions :column="4" size="small" :content-style="{ color: '#888' }">
          <a-descriptions-item label="上课教师">
            {{ detailData?.teacherName || props.detail?.teacherName || '-' }}
          </a-descriptions-item>
          <a-descriptions-item label="上课助教">
            {{ assistantText }}
          </a-descriptions-item>
          <a-descriptions-item label="上课教室">
            {{ detailData?.classroomName || props.detail?.classroomName || '-' }}
          </a-descriptions-item>
          <a-descriptions-item label="重复规则">
            {{ repeatRuleText }}
          </a-descriptions-item>
          <a-descriptions-item label="对内备注">
            {{ remarkText }}
          </a-descriptions-item>
          <a-descriptions-item label="对外备注">
            -
          </a-descriptions-item>
        </a-descriptions>
      </div>

      <div class="tabs">
        <a-tabs
          active-key="0"
          size="large"
          :tab-bar-style="{ 'border-radius': '0px', 'padding-left': '24px' }"
        >
          <a-tab-pane key="0" tab="学员名单">
            <a-card title="上课学员" :bordered="false">
              <a-table
                :columns="studentColumns"
                :data-source="students"
                :pagination="false"
                row-key="studentId"
                :scroll="{ x: 700 }"
              >
                <template #bodyCell="{ column, record }">
                  <template v-if="column.key === 'phone'">
                    <div class="student-phone-cell">
                      <div class="student-phone-relation">
                        {{ record.phoneRelationshipText || '-' }}
                      </div>
                      <div class="student-phone-number">
                        {{ record.maskedPhone || record.phone || '-' }}
                      </div>
                    </div>
                  </template>
                  <template v-else-if="column.key === 'studentName'">
                    <div class="student-name-cell" @click="handleViewStudent(record.studentId)">
                      <img class="student-avatar" :src="record.avatarUrl || defaultStudentAvatar" alt="">
                      <div class="student-name-text">
                        {{ record.studentName || '-' }}
                      </div>
                    </div>
                  </template>
                  <template v-else-if="column.key === 'studentType'">
                    <span class="student-type-text">
                      {{ studentTypeText }}
                    </span>
                  </template>
                  <template v-else-if="column.key === 'action'">
                    <a-button type="link" class="px0" @click="handleViewStudent(record.studentId)">
                      详情
                    </a-button>
                  </template>
                </template>
              </a-table>
            </a-card>
          </a-tab-pane>
        </a-tabs>
      </div>
    </a-spin>
    <student-info-drawer v-model:open="openStudentDrawer" />
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

.student-name-cell {
  display: flex;
  align-items: center;
  gap: 12px;
  cursor: pointer;
}

.student-avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  object-fit: cover;
  flex: 0 0 36px;
}

.student-name-text {
  color: #222;
  font-weight: 500;
}

.student-name-cell:hover .student-name-text {
  color: #06f;
}

.student-phone-cell {
  line-height: 1.5;
}

.student-phone-relation {
  color: #222;
}

.student-phone-number {
  color: #888;
}

.student-type-text {
  color: #222;
}
</style>
