<script setup>
import { CloseCircleOutlined, CloseOutlined, ExclamationCircleFilled, InfoCircleOutlined } from '@ant-design/icons-vue'
import dayjs from 'dayjs'
import ClassStudentList from './class-student-list.vue'
import classRecord from './class-record.vue'
import schedule from './class-list-schedule.vue'
import waitingRollCallSchedule from './waiting-roll-call-schedule.vue'
import { getGroupClassDetailApi } from '@/api/edu-center/group-class'
import { openCloseClassConfirm } from '@/utils/closeClassConfirm'
import messageService from '@/utils/messageService'

const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
  record: {
    type: Object,
    default: () => ({}),
  },
})

const emit = defineEmits(['update:open', 'edit'])
const activeKey = ref('0')

const openDrawer = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

watch(() => openDrawer.value, (newVal) => {
  if (newVal)
    activeKey.value = '0'
})

const detailData = ref(null)
let detailReqSeq = 0

watch(
  () => [props.open, props.record],
  () => {
    if (!props.open) {
      detailReqSeq += 1
      detailData.value = null
      return
    }
    const classId = String(props.record?.id || '').trim()
    if (!classId) {
      detailReqSeq += 1
      detailData.value = null
      return
    }
    const seq = ++detailReqSeq
    getGroupClassDetailApi({ id: classId })
      .then((res) => {
        if (seq !== detailReqSeq)
          return
        if (res.code === 200 && res.result)
          detailData.value = res.result
        else
          messageService.error(res.message || '加载班级详情失败')
      })
      .catch((error) => {
        if (seq !== detailReqSeq)
          return
        console.error('load group class detail failed', error)
        messageService.error(error?.response?.data?.message || error?.message || '加载班级详情失败')
      })
  },
  { immediate: true, flush: 'sync' },
)

const displayRecord = computed(() => ({
  ...(props.record || {}),
  ...(detailData.value || {}),
  createdStaffName: detailData.value?.createdStaffName || props.record?.createdStaffName || '',
}))

const isClassClosed = computed(() => Number(displayRecord.value?.status || 0) === 2)
const studentTabLabel = computed(() => `学员(${Number(displayRecord.value?.studentCount || 0)})`)

function formatCreatedTime(value) {
  if (!value || `${value}`.startsWith('0001-01-01'))
    return '-'
  const date = dayjs(value)
  return date.isValid() ? date.format('YYYY-MM-DD HH:mm') : '-'
}

function getStatusText(status) {
  return Number(status) === 2 ? '已结班' : '开班中'
}

function formatTeacherNames(teachers) {
  if (!Array.isArray(teachers) || teachers.length === 0)
    return '-'
  const names = teachers.map(item => String(item?.name || '').trim()).filter(Boolean)
  return names.length ? names.join('、') : '-'
}

function formatClassTime(value) {
  const num = Number(value)
  if (!Number.isFinite(num))
    return '-'
  if (num <= 0)
    return '0'
  if (Number.isInteger(num))
    return String(num)
  return `${num}`.replace(/\.?0+$/, '')
}

function formatMaxCount(value) {
  const num = Number(value)
  return Number.isFinite(num) && num > 0 ? String(num) : '不限'
}

function getCourseModeText(record) {
  return record?.isMultiProduct ? '组合课' : '课程'
}

const openModal = ref(false)
function handleDelete() {
  console.log('删除')
  openModal.value = true
}

const editClassInfoModal = ref(false)
function handleEditClassInfo() {
  editClassInfoModal.value = true
}

const editRollNameModal = ref(false)
function handleEditRollName() {
  editRollNameModal.value = true
}

function handleCloseClass() {
  openCloseClassConfirm()
}

function handleEditClass() {
  emit('edit', displayRecord.value)
}

function handleQuickUpgrade() {
  messageService.info('快捷升班功能待实现')
}

function handleExportRollCallSheet() {
  messageService.info('导出点名表功能待实现')
}
</script>

<template>
  <div>
    <a-drawer
      v-model:open="openDrawer" :push="{ distance: 80 }" :body-style="{ padding: '0', background: '#f7f7fd' }"
      :closable="false" width="1165px" placement="right"
    >
      <template #title>
        <div class="custom-header flex justify-between h-4 flex-items-center">
          <div class="text-5">
            班级详情
          </div>
          <a-button type="text" class="close-btn" @click="openDrawer = false">
            <template #icon>
              <CloseOutlined class="text-5 close-icon" />
            </template>
          </a-button>
        </div>
      </template>
      <div class="contenter flex flex-center bg-white px6 py3">
        <div class="avatarBox w-16 h-16 relative">
          <img
            width="64" height="64" class="rounded-100"
            src="https://pcsys.admin.ybc365.com//e64c7fd6-2edc-412f-9141-a9904be88b4f.png" alt=""
          >
        </div>
        <div class="info flex flex-1 ml-4 flex-col min-w-0">
          <div class="top flex justify-between flex-center flex-1">
            <a-space>
              <div class="name text-5 font-800">
                {{ displayRecord?.name || '-' }}
              </div>
              <span
                class="text-3 px3 py2px rounded-10 ml2"
                :class="isClassClosed ? 'text-#666 bg-#f5f5f5' : 'bg-#e6f0ff text-#06f'"
              >
                {{ getStatusText(displayRecord?.status) }}
              </span>
            </a-space>
            <a-space>
              <a-button @click="handleQuickUpgrade">
                快捷升班
              </a-button>
              <a-button @click="handleExportRollCallSheet">
                导出点名表
              </a-button>
              <a-button v-if="!isClassClosed" @click="handleCloseClass">
                结班
              </a-button>
              <a-button v-if="!isClassClosed" type="primary" @click="handleEditClass">
                编辑班级
              </a-button>
            </a-space>
          </div>
          <div class="bottom flex-1 flex flex-items-center mt-2">
            <div class="birthday flex-center">
              <span class="text-14px text-#888">
                {{ displayRecord?.createdStaffName || '-' }} 创建于 {{ formatCreatedTime(displayRecord?.createdTime) }}
              </span>
            </div>
          </div>
        </div>
      </div>
      <div class="desc pt-4 bg-white px6 py3 pb0">
        <a-descriptions :column="4" size="small" :content-style="{ color: '#888' }">
          <a-descriptions-item label="班主任">
            {{ formatTeacherNames(displayRecord?.teachers) }}
          </a-descriptions-item>
          <a-descriptions-item label="学员记录课时">
            {{ formatClassTime(displayRecord?.defaultStudentClassTime) }}
          </a-descriptions-item>
          <a-descriptions-item label="上课教师授课课时">
            {{ formatClassTime(displayRecord?.defaultTeacherClassTime) }}
          </a-descriptions-item>
          <a-descriptions-item>
            <template #label>
              <span class="flex items-center"><span>满班人数</span>
                <a-popover title="满班人数">
                  <template #content>
                    <div class="w-280px">当开启满班限制后，报名选班时，未付款/未审批完成/未处理完成订单，将会锁定班级人数名额占用班级人数</div>
                  </template>
                  <InfoCircleOutlined class="ml-2px text-#666" />
                </a-popover></span>
            </template>
            {{ formatMaxCount(displayRecord?.maxCount) }}
          </a-descriptions-item>
          <a-descriptions-item>
            <template #label>
              <span class="flex items-center"><span>课程模式</span>
                <a-popover title="课程模式">
                  <template #content>
                    <div>
                      课程：该课程下的学员可在同一班级上课
                    </div>
                    <div>
                      组合课：该组合课程范围内，多个课程的对应学员可在同一班级上课
                    </div>
                  </template>
                  <InfoCircleOutlined class="ml-2px text-#666" />
                </a-popover></span>
            </template>
            {{ getCourseModeText(displayRecord) }}
          </a-descriptions-item>
          <a-descriptions-item label="关联课程">
            {{ displayRecord?.lessonName || '-' }}
          </a-descriptions-item>
          <a-descriptions-item label="备注">
            {{ displayRecord?.remark || '-' }}
          </a-descriptions-item>
        </a-descriptions>
      </div>
      <div class="tabs">
        <a-tabs
          v-model:active-key="activeKey" size="large" :tab-bar-style="{
            'border-radius': '0px', 'padding-left': '24px',
          }"
        >
          <a-tab-pane key="0" :tab="studentTabLabel">
            <ClassStudentList />
          </a-tab-pane>
          <a-tab-pane key="1" tab="日程">
            <schedule />
          </a-tab-pane>
          <a-tab-pane key="2" tab="待点名日程">
            <waitingRollCallSchedule />
          </a-tab-pane>
          <a-tab-pane key="3" tab="上课记录">
            <classRecord />
          </a-tab-pane>
        </a-tabs>
      </div>
    </a-drawer>
    <a-modal
      v-model:open="openModal" centered :footer="false" :closable="false" :mask-closable="false"
      :keyboard="false" width="420px" @ok="handleDelete"
    >
      <div class="text-18px mb-12px font500">
        <CloseCircleOutlined class="text-#f00 mr2 text-5" /> 删除上课点名记录？
      </div>
      <div class="pl-30px text-#666">
        <div>1.删除后已点名扣费学员将会反还学费，并减少对应的已确认收入;</div>
        <div>2.若包含试听、补课学员状态不会修改，已试听状态变成已取消状态，已补课学员将退回至未安排状态，并删除上课记录;</div>
        <div>3.删除上课点名记录后，（除了试听日程）所对应的日程中的学员点名状态变成未点名;</div>
        <div>4.删除上课点名记录后，日程状态从已点名变成未点名。</div>
        <div class="text-#f00 mt-12px">
          <ExclamationCircleFilled /> 此操作不可撤销，请谨慎操作
        </div>
      </div>
      <a-space class="mt-24px flex justify-end">
        <a-button danger ghost>
          删除
        </a-button>
        <a-button class="text-#666" @click="openModal = false">
          再想想
        </a-button>
      </a-space>
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

</style>
