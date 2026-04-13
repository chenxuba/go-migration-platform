<script setup lang="ts">
import { CloseCircleOutlined, CloseOutlined, ExclamationCircleFilled } from '@ant-design/icons-vue'
import dayjs from 'dayjs'
import { computed, ref, watch } from 'vue'
import scheduleClassImage from '@/assets/images/timetable/schedule-class.png'
import scheduleOneToOneImage from '@/assets/images/timetable/schedule-one2one.png'
import { deleteTeachingRecordApi, getTeachingRecordDetailApi, type TeachingRecordDetailResult, type TeachingRecordDetailTeacher } from '@/api/edu-center/class-record'
import messageService from '@/utils/messageService'
import EditClassInfoModal from './edit-class-info-modal.vue'
import EditRollNameModal from './edit-roll-name-modal.vue'

const props = withDefaults(defineProps<{
  open: boolean
  teachingRecordId?: string
}>(), {
  teachingRecordId: '',
})

const emit = defineEmits(['update:open', 'updated', 'deleted'])
const activeKey = ref('0')
const openModal = ref(false)
const editClassInfoModal = ref(false)
const editRollNameModal = ref(false)
const loading = ref(false)
const deleting = ref(false)
const detailData = ref<TeachingRecordDetailResult | null>(null)
const hasDetail = computed(() => String(detailData.value?.teachingRecordId || '').trim() !== '')

const openDrawer = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

const currentTeachingRecordId = computed(() => String(props.teachingRecordId || '').trim())
const sourceCover = computed(() => (Number(detailData.value?.timetableSourceType || 0) === 2 ? scheduleOneToOneImage : scheduleClassImage))
const headerTitle = computed(() => detailData.value?.sourceName || detailData.value?.lessonName || '-')
const teacherList = computed(() => Array.isArray(detailData.value?.teacherList) ? detailData.value?.teacherList || [] : [])
const mainTeacherText = computed(() => formatTeacherNames(teacherList.value.filter(item => Number(item.type || 0) === 1)))
const assistantTeacherText = computed(() => formatTeacherNames(teacherList.value.filter(item => Number(item.type || 0) !== 1)))
const classDurationText = computed(() => {
  const start = dayjs(detailData.value?.startTime)
  const end = dayjs(detailData.value?.endTime)
  if (!start.isValid() || !end.isValid())
    return '-'
  const minutes = Math.max(end.diff(start, 'minute'), 0)
  return `${minutes}分钟`
})
const timeText = computed(() => {
  const start = dayjs(detailData.value?.startTime)
  const end = dayjs(detailData.value?.endTime)
  if (!start.isValid() || !end.isValid())
    return '-'
  const weekMap = ['周日', '周一', '周二', '周三', '周四', '周五', '周六']
  return `${start.format('YYYY-MM-DD')}(${weekMap[start.day()] || '-'})${start.format('HH:mm')} ~ ${end.format('HH:mm')}`
})
const teacherClassTimeText = computed(() => `教师记录${formatClassTime(detailData.value?.teacherClassTime)}`)
const createdText = computed(() => {
  const rawTime = String(detailData.value?.createdTime || '').trim()
  const time = dayjs(rawTime).isValid() ? dayjs(rawTime).format('YYYY-MM-DD HH:mm') : rawTime
  return time || '-'
})

let loadSeq = 0

function handleDelete() {
  openModal.value = true
}

async function handleConfirmDelete() {
  const teachingRecordId = currentTeachingRecordId.value
  if (!teachingRecordId || deleting.value)
    return
  deleting.value = true
  try {
    const res = await deleteTeachingRecordApi({ teachingRecordId })
    if (res.code !== 200 || res.result !== true)
      throw new Error(res.message || '删除上课点名记录失败')
    messageService.success('删除成功')
    openModal.value = false
    openDrawer.value = false
    emit('updated')
    emit('deleted')
  }
  catch (error: any) {
    messageService.error(error?.response?.data?.message || error?.message || '删除上课点名记录失败')
  }
  finally {
    deleting.value = false
  }
}

function handleEditClassInfo() {
  editClassInfoModal.value = true
}

function handleEditRollName() {
  editRollNameModal.value = true
}

function formatTeacherNames(list: TeachingRecordDetailTeacher[]) {
  const names = list.map(item => String(item.teacherName || '').trim()).filter(Boolean)
  return names.length ? names.join('、') : '-'
}

function formatClassTime(value?: number) {
  const num = Number(value || 0)
  if (!Number.isFinite(num))
    return '0课时'
  const text = Number.isInteger(num) ? String(num) : num.toFixed(2).replace(/\.?0+$/, '')
  return `${text}课时`
}

async function loadDetail() {
  const teachingRecordId = currentTeachingRecordId.value
  if (!openDrawer.value || !teachingRecordId) {
    detailData.value = null
    return
  }

  const seq = ++loadSeq
  loading.value = true
  try {
    const res = await getTeachingRecordDetailApi({ teachingRecordId })
    if (seq !== loadSeq)
      return
    if (res.code !== 200)
      throw new Error(res.message || '加载上课记录详情失败')
    const data = res.result
    detailData.value = data && String(data.teachingRecordId || '').trim() ? data : null
  }
  catch (error: any) {
    if (seq !== loadSeq)
      return
    detailData.value = null
    messageService.error(error?.response?.data?.message || error?.message || '加载上课记录详情失败')
  }
  finally {
    if (seq === loadSeq)
      loading.value = false
  }
}

watch(
  () => `${openDrawer.value}|${currentTeachingRecordId.value}`,
  async () => {
    if (!openDrawer.value) {
      detailData.value = null
      loading.value = false
      deleting.value = false
      activeKey.value = '0'
      return
    }
    activeKey.value = '0'
    await loadDetail()
  },
  { immediate: true },
)
</script>

<template>
  <div>
    <a-drawer
      v-model:open="openDrawer"
      :push="{ distance: 80 }"
      :body-style="{ padding: '0', background: hasDetail ? '#f7f7fd' : '#fff' }"
      :closable="false"
      width="1165px"
      placement="right"
    >
      <template #title>
        <div class="custom-header flex justify-between h-4 flex-items-center">
          <div class="text-5">
            上课记录详情
          </div>
          <a-button type="text" class="close-btn" @click="openDrawer = false">
            <template #icon>
              <CloseOutlined class="text-5 close-icon" />
            </template>
          </a-button>
        </div>
      </template>

      <a-spin :spinning="loading">
        <template v-if="hasDetail">
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
                <a-space>
                  <a-button danger ghost @click="handleDelete">
                    删除
                  </a-button>
                  <a-button type="primary" @click="handleEditRollName">
                    编辑点名
                  </a-button>
                  <a-button type="primary">
                    课堂点评
                  </a-button>
                  <a-button type="primary">
                    课后任务
                  </a-button>
                </a-space>
              </div>
              <div class="bottom flex-1 flex flex-items-center mt-2">
                <div class="birthday flex-center">
                  <span class="text-4 text-#222">{{ timeText }}</span>
                  <span class="bg-#e6f0ff text-#06f text-3 px2 py1 rounded-10 ml2">{{ classDurationText }}</span>
                </div>
              </div>
            </div>
          </div>

          <div class="desc pt-4 bg-white px6 py3 pb0">
            <a-descriptions :column="4" size="small" :content-style="{ color: '#888' }">
              <a-descriptions-item label="上课老师">
                {{ mainTeacherText }}
              </a-descriptions-item>
              <a-descriptions-item label="上课助教">
                {{ assistantTeacherText }}
              </a-descriptions-item>
              <a-descriptions-item label="上课教室">
                {{ detailData?.classRoomName || '-' }}
              </a-descriptions-item>
              <a-descriptions-item label="本次上课">
                {{ teacherClassTimeText }}
              </a-descriptions-item>
              <a-descriptions-item label="创建时间">
                {{ createdText }}
              </a-descriptions-item>
              <a-descriptions-item label="科目">
                {{ detailData?.subjectName || '-' }}
              </a-descriptions-item>
              <a-descriptions-item>
                <span class="text-#06f cursor-pointer" @click="handleEditClassInfo">编辑上课信息</span>
              </a-descriptions-item>
            </a-descriptions>
          </div>

          <div class="tabs">
            <a-tabs
              v-model:active-key="activeKey"
              size="large"
              :tab-bar-style="{ 'border-radius': '0px', 'padding-left': '24px' }"
            >
              <a-tab-pane key="0" tab="点名详情">
                <call-name-details :detail="detailData" :loading="loading" />
              </a-tab-pane>
              <a-tab-pane key="1" tab="点名变更记录">
                <call-name-change-details />
              </a-tab-pane>
            </a-tabs>
          </div>
        </template>
        <div v-else class="deleted-empty-state flex flex-center bg-white">
          <a-empty description="当前上课点名记录已被删除" />
        </div>
      </a-spin>
    </a-drawer>

    <a-modal
      v-model:open="openModal"
      centered
      :footer="false"
      :closable="false"
      :mask-closable="false"
      :keyboard="false"
      width="440px"
    >
      <div class="text-18px mb-12px font500">
        <CloseCircleOutlined class="text-#f00 mr2 text-5" /> 删除上课点名记录？
      </div>
      <div class="pl-30px text-#666">
        <div>1.删除后已点名扣费学员将会返还学费，并减少对应的已确认收入;</div>
        <div>2.若包含试听学员，已试听状态将变成已取消状态，并删除上课记录；若包含补课学员，已补课状态将变成已安排或未安排状态，并删除上课记录;</div>
        <div>3.删除上课点名记录后，所对应的日程中的学员点名状态变成未点名;</div>
        <div>4.删除上课点名记录后，日程状态从已点名变成未点名。</div>
        <div class="text-#f00 mt-12px">
          <ExclamationCircleFilled /> 此操作不可撤销，请谨慎操作
        </div>
      </div>
      <a-space class="mt-24px flex justify-end">
        <a-button danger ghost :loading="deleting" @click="handleConfirmDelete">
          删除
        </a-button>
        <a-button class="text-#666" :disabled="deleting" @click="openModal = false">
          再想想
        </a-button>
      </a-space>
    </a-modal>

    <EditClassInfoModal v-model:open="editClassInfoModal" />
    <EditRollNameModal v-model:open="editRollNameModal" :detail="detailData" />
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
    bottom: 0px !important;

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

.deleted-empty-state {
  min-height: calc(100vh - 55px);
}

</style>
