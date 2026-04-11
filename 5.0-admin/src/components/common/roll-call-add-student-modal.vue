<script setup lang="ts">
import type { TableColumnsType } from 'ant-design-vue'
import { CloseOutlined } from '@ant-design/icons-vue'
import { debounce } from 'lodash-es'
import { addTeachingScheduleStudentsCurrentApi, pageTeachingScheduleStudentCandidatesApi, type TeachingScheduleStudentCandidate } from '@/api/edu-center/teaching-schedule'
import messageService from '@/utils/messageService'

const props = withDefaults(defineProps<{
  open?: boolean
  title?: string
  scheduleId?: string
  studentType?: number
}>(), {
  open: false,
  title: '添加学员',
  scheduleId: '',
  studentType: 4,
})

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'success'): void
}>()

const openModal = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

const keyword = ref('')
const listLoading = ref(false)
const submitLoading = ref(false)
const data = ref<TeachingScheduleStudentCandidate[]>([])
const selectedRowKeys = ref<string[]>([])
const selectedRows = ref<TeachingScheduleStudentCandidate[]>([])
const defaultStudentAvatar = 'https://cdn.schoolpal.cn/schoolpal/next-erp/avator_male.png'

const paginationState = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
})

const isDynamicMode = computed(() => Boolean(String(props.scheduleId || '').trim()))

const columns = computed<TableColumnsType<TeachingScheduleStudentCandidate>>(() => [
  {
    title: '学员姓名',
    dataIndex: 'studentName',
    key: 'studentName',
  },
  {
    title: '联系电话',
    dataIndex: 'phone',
    key: 'phone',
  },
  {
    title: '学员状态',
    dataIndex: 'studentStatusText',
    key: 'studentStatusText',
    width: 140,
  },
])

const tablePagination = computed(() => {
  if (!isDynamicMode.value)
    return false
  return {
    current: paginationState.current,
    pageSize: paginationState.pageSize,
    total: paginationState.total,
    showSizeChanger: true,
    showTotal: (total: number) => `共 ${total} 条`,
    pageSizeOptions: ['10', '20', '50'],
  }
})

const rowSelection = computed(() => ({
  selectedRowKeys: selectedRowKeys.value,
  onChange: (keys: (string | number)[], rows: TeachingScheduleStudentCandidate[]) => {
    selectedRowKeys.value = keys.map(key => String(key))
    selectedRows.value = rows
  },
}))

function getStudentStatusTagClass(status?: number) {
  switch (Number(status ?? -1)) {
    case 1:
      return 'student-status-tag student-status-tag--enrolled'
    case 2:
      return 'student-status-tag student-status-tag--history'
    default:
      return 'student-status-tag student-status-tag--intent'
  }
}

function staticRows(): TeachingScheduleStudentCandidate[] {
  return [
    {
      studentId: 'static-1',
      studentName: '演示学员',
      phone: '176****1111',
      maskedPhone: '176****1111',
      phoneRelationshipText: '妈妈',
    },
  ]
}

async function loadTableData() {
  if (!isDynamicMode.value) {
    data.value = staticRows()
    paginationState.total = data.value.length
    return
  }

  listLoading.value = true
  try {
    const res = await pageTeachingScheduleStudentCandidatesApi({
      pageRequestModel: {
        pageIndex: paginationState.current,
        pageSize: paginationState.pageSize,
      },
      queryModel: {
        scheduleId: String(props.scheduleId || '').trim(),
        studentType: Number(props.studentType || 0),
        keyword: keyword.value.trim() || undefined,
      },
    })
    if (res.code !== 200) {
      data.value = []
      paginationState.total = 0
      messageService.error(res.message || '获取可添加学员失败')
      return
    }
    data.value = Array.isArray(res.result?.list) ? res.result.list : []
    paginationState.total = Number(res.result?.total) || 0
  }
  catch (error: any) {
    console.error(error)
    data.value = []
    paginationState.total = 0
    messageService.error(error?.response?.data?.message || error?.message || '获取可添加学员失败')
  }
  finally {
    listLoading.value = false
  }
}

function resetState() {
  keyword.value = ''
  selectedRowKeys.value = []
  selectedRows.value = []
  paginationState.current = 1
  paginationState.pageSize = 10
  paginationState.total = 0
  debouncedReload.cancel()
}

function closeFun() {
  openModal.value = false
  resetState()
}

function handleTableChange(pagination: { current?: number, pageSize?: number }) {
  const nextCurrent = Number(pagination?.current || 1)
  const nextPageSize = Number(pagination?.pageSize || 10)
  if (nextCurrent === paginationState.current && nextPageSize === paginationState.pageSize)
    return
  paginationState.current = nextCurrent
  paginationState.pageSize = nextPageSize
  loadTableData()
}

const debouncedReload = debounce(() => {
  paginationState.current = 1
  loadTableData()
}, 300)

watch(keyword, () => {
  if (!openModal.value)
    return
  debouncedReload()
})

watch(
  () => props.open,
  (open) => {
    if (!open)
      return
    resetState()
    loadTableData()
  },
)

onBeforeUnmount(() => {
  debouncedReload.cancel()
})

async function handleSubmit() {
  if (selectedRows.value.length === 0) {
    messageService.warning('请选择学员')
    return
  }
  if (!isDynamicMode.value) {
    messageService.info('当前页面暂未接入真实添加逻辑')
    return
  }

  submitLoading.value = true
  try {
    const res = await addTeachingScheduleStudentsCurrentApi({
      scheduleId: String(props.scheduleId || '').trim(),
      studentType: Number(props.studentType || 0),
      studentIds: selectedRows.value.map(item => String(item.studentId || '')).filter(Boolean),
    })
    if (res.code !== 200) {
      messageService.error(res.message || '添加学员失败')
      return
    }
    messageService.success('添加学员成功')
    emit('success')
    closeFun()
  }
  catch (error: any) {
    messageService.error(error?.response?.data?.message || error?.message || '添加学员失败')
  }
  finally {
    submitLoading.value = false
  }
}
</script>

<template>
  <a-modal
    v-model:open="openModal"
    centered
    class="modal-content-box"
    :keyboard="false"
    :closable="false"
    :mask-closable="false"
    :width="800"
    @cancel="closeFun"
  >
    <template #title>
      <div class="text-5 flex justify-between flex-center">
        <span>{{ title }}</span>
        <a-button type="text" class="close-btn" @click="closeFun">
          <template #icon>
            <CloseOutlined class="text-5 close-icon" />
          </template>
        </a-button>
      </div>
    </template>
    <div class="contenter scrollbar">
      <a-input v-model:value="keyword" placeholder="搜索学员姓名/手机号" class="h-48px rounded-12px">
        <template #prefix>
          <img
            src="https://prod-tbu-next-erp-cdn.schoolpal.cn/next-pc-static/static/12181/static/magnifying.2bcc08ab.svg"
            alt="" class="pr-6px mt--2px"
          >
        </template>
      </a-input>
      <a-table
        :columns="columns"
        :data-source="data"
        row-key="studentId"
        class="mt-12px"
        :loading="listLoading"
        :pagination="tablePagination"
        :row-selection="rowSelection"
        @change="handleTableChange"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.dataIndex === 'studentName'">
            <div class="student-name-cell">
              <img
                :src="record.avatarUrl || defaultStudentAvatar"
                class="student-avatar"
                alt=""
              >
              {{ record.studentName || '-' }}
            </div>
          </template>
          <template v-else-if="column.dataIndex === 'phone'">
            <div class="student-phone-cell">
              <span v-if="record.phoneRelationshipText">{{ record.phoneRelationshipText }}</span>
              <span>{{ record.maskedPhone || '-' }}</span>
            </div>
          </template>
          <template v-else-if="column.dataIndex === 'studentStatusText'">
            <span :class="getStudentStatusTagClass(record.studentStatus)">
              {{ record.studentStatusText || '-' }}
            </span>
          </template>
        </template>
      </a-table>
    </div>
    <template #footer>
      <div class="flex justify-between">
        <span>已勾选{{ selectedRows.length }}人</span>
        <div>
          <a-button danger ghost @click="closeFun">
            关闭
          </a-button>
          <a-button type="primary" ghost :loading="submitLoading" :disabled="selectedRows.length === 0" @click="handleSubmit">
            确定
          </a-button>
        </div>
      </div>
    </template>
  </a-modal>
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

.contenter {
  padding: 24px;
}

.student-name-cell {
  display: flex;
  align-items: center;
  gap: 10px;
  color: #1f2329;
  font-weight: 500;
}

.student-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  flex: 0 0 32px;
  object-fit: cover;
}

.student-phone-cell {
  display: flex;
  gap: 6px;
  align-items: center;
  color: #5c6570;
}

.student-status-tag {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 72px;
  padding: 2px 10px;
  border-radius: 999px;
  font-size: 12px;
  line-height: 20px;
  white-space: nowrap;
}

.student-status-tag--enrolled {
  color: #1677ff;
  background: #e8f3ff;
}

.student-status-tag--history {
  color: #666d7a;
  background: #f1f3f5;
}

.student-status-tag--intent {
  color: #d46b08;
  background: #fff3e8;
}
</style>

<style>
.modal-content-box .ant-modal-header {
  padding: 10px 16px !important;
  margin-bottom: 0;
}

.modal-content-box .ant-modal-body {
  padding: 0 !important;
}
</style>
