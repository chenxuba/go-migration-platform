<script setup lang="ts">
import { ExclamationCircleOutlined } from '@ant-design/icons-vue'
import type { TableColumnsType } from 'ant-design-vue'
import { computed, ref, watch } from 'vue'
import scheduleClassRepeatImage from '@/assets/images/timetable/schedule-class.png'
import scheduleClassSingleImage from '@/assets/images/timetable/schedule-free.png'
import { getGroupClassDrawerSchedulesApi, type GroupClassDrawerScheduleItem } from '@/api/edu-center/group-class'
import SmartTimetableScheduleDetailDrawer from '@/components/edu-center/timetable/smart-timetable-schedule-detail-drawer.vue'
import messageService from '@/utils/messageService'

const props = withDefaults(defineProps<{
  open?: boolean
  classId?: string
  className?: string
}>(), {
  open: false,
  classId: '',
  className: '',
})

const columns: TableColumnsType<GroupClassDrawerScheduleItem> = [
  {
    title: '重复规则',
    dataIndex: 'repeatRule',
    key: 'repeatRule',
    width: 260,
  },
  {
    title: '上课时间',
    dataIndex: 'timeText',
    key: 'timeText',
    width: 160,
  },
  {
    title: '已上/排课',
    dataIndex: 'status',
    key: 'status',
    width: 110,
  },
  {
    title: '上课教师',
    dataIndex: 'teacherName',
    key: 'teacherName',
    width: 110,
  },
  {
    title: '上课助教',
    dataIndex: 'assistantText',
    key: 'assistantText',
    width: 120,
  },
  {
    title: '上课教室',
    dataIndex: 'classroomName',
    key: 'classroomName',
    width: 120,
  },
  {
    title: '操作',
    dataIndex: 'action',
    key: 'action',
    width: 80,
    fixed: 'right',
  },
]

const loading = ref(false)
const dataSource = ref<GroupClassDrawerScheduleItem[]>([])
const detailOpen = ref(false)
const detailState = ref<Record<string, any> | null>(null)

const totalWidth = computed(() =>
  columns.reduce((sum, item) => sum + Number(item.width || 0), 0),
)

async function loadList() {
  const classId = String(props.classId || '').trim()
  if (!props.open || !classId) {
    dataSource.value = []
    return
  }
  loading.value = true
  try {
    const res = await getGroupClassDrawerSchedulesApi({ classId })
    if (res.code !== 200)
      throw new Error(res.message || '加载班级日程失败')
    dataSource.value = Array.isArray(res.result?.list) ? res.result.list : []
  }
  catch (error: any) {
    dataSource.value = []
    messageService.error(error?.response?.data?.message || error?.message || '加载班级日程失败')
  }
  finally {
    loading.value = false
  }
}

function handleViewDetail(record: GroupClassDrawerScheduleItem | Record<string, any>) {
  detailState.value = {
    scheduleId: record.detailScheduleId,
    id: record.detailScheduleId,
    batchNo: record.batchNo,
    batchSize: record.scheduleCount,
    lessonTitle: record.lessonName || props.className || '日程详情',
    assistantText: record.assistantText,
    classroomName: record.classroomName,
    teacherName: record.teacherName,
    batchMeta: record.batchMeta,
  }
  detailOpen.value = true
}

function handleQuickSchedule() {
  messageService.info('一键排课功能待接入')
}

watch(
  () => `${props.open}|${String(props.classId || '').trim()}`,
  () => {
    loadList()
  },
  { immediate: true },
)
</script>

<template>
  <div class="m-12px">
    <div class="bg-#fff pt-18px px-20px rounded-10px">
      <div class="flex justify-between items-center">
        <custom-title :title="`共 ${dataSource.length} 个日程`" font-size="14px" class="pb-12px" />
        <a-button type="primary" class="mb-12px" @click="handleQuickSchedule">
          一键排课
        </a-button>
      </div>
      <a-table
        row-key="key"
        size="small"
        :loading="loading"
        :columns="columns"
        :data-source="dataSource"
        :pagination="false"
        :scroll="{ x: totalWidth }"
      >
        <template #headerCell="{ column }">
          <template v-if="column.key === 'status'">
            已上/排课
            <a-popover title="已上/排课">
              <template #content>
                <div>已完成日程数/排课日程总数</div>
              </template>
              <ExclamationCircleOutlined />
            </a-popover>
          </template>
        </template>
        <template #bodyCell="{ column, record }">
          <template v-if="column.dataIndex === 'repeatRule'">
            <div class="flex flex-items-center">
              <img
                class="w-34px h-34px"
                :src="Number(record.type || 0) === 1 ? scheduleClassRepeatImage : scheduleClassSingleImage"
                alt=""
              >
              <div class="ml-12px text-#666 leading-20px">
                <div class="text-14px text-#222">
                  {{ record.repeatRule || '-' }}
                </div>
                <div class="text-13px">
                  {{ record.dateRangeText || '-' }}
                </div>
              </div>
            </div>
          </template>
          <template v-if="column.dataIndex === 'timeText'">
            <div>{{ record.timeText || '-' }}</div>
            <div class="text-#888">
              {{ record.weekdayText || '-' }}
            </div>
          </template>
          <template v-if="column.dataIndex === 'status'">
            {{ `${Number(record.completedCount || 0)}/${Number(record.scheduleCount || 0)}节` }}
          </template>
          <template v-if="column.dataIndex === 'teacherName'">
            {{ record.teacherName || '-' }}
          </template>
          <template v-if="column.dataIndex === 'assistantText'">
            {{ record.assistantText || '-' }}
          </template>
          <template v-if="column.dataIndex === 'classroomName'">
            {{ record.classroomName || '-' }}
          </template>
          <template v-if="column.dataIndex === 'action'">
            <a-space :size="12">
              <a @click="handleViewDetail(record)">详情</a>
            </a-space>
          </template>
        </template>
      </a-table>
    </div>
    <SmartTimetableScheduleDetailDrawer
      v-model:open="detailOpen"
      :detail="detailState"
      @updated="loadList"
    />
  </div>
</template>
