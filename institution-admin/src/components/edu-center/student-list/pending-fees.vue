<script setup lang="ts">
import { computed, nextTick, onMounted, ref } from 'vue'
import { DownOutlined } from '@ant-design/icons-vue'
import dayjs from 'dayjs'
import messageService from '@/utils/messageService'
import { pageGroupClassesApi } from '@/api/edu-center/group-class'
import {
  getPendingRenewalStudentsPagedListApi,
  type PendingRenewalStudentItem,
  sendPendingRenewalWechatReminderApi,
} from '@/api/edu-center/student-list'
import { useTableColumns } from '@/composables/useTableColumns'
import { useStudentListRefresh } from '@/composables/useStudentListRefresh'
import { Sex, SexLabel } from '@/enums'
import PendingRenewalMessageRecordDrawer from './pending-renewal-message-record-drawer.vue'

const tipsText = '（规则：剩余课时<15 / 剩余天数<15 / 剩余金额<500元）'
const displayArray = ['intentionCourse', 'className', 'classTeacher', 'currentStatus']

const allFilterRef = ref<{
  clearQuickFilter?: (id?: string | number, type?: string) => void
} | null>(null)
const loading = ref(false)
const dataSource = ref<PendingRenewalStudentItem[]>([])
const classNameOptionsData = ref<Array<{ id: string, value: string }>>([])
const selectedRowKeys = ref<string[]>([])
const selectedRows = ref<PendingRenewalStudentItem[]>([])
const messageRecordOpen = ref(false)
const sendingWechatReminder = ref(false)
const messageRecordDrawerRef = ref<InstanceType<typeof PendingRenewalMessageRecordDrawer> | null>(null)
const summary = ref({
  total: 0,
  studentCount: 0,
})

const pagination = ref({
  current: 1,
  pageSize: 50,
  total: 0,
  showSizeChanger: true,
  showQuickJumper: true,
  pageSizeOptions: ['20', '50', '100'],
  showTotal: (total: number) => `共 ${total} 条`,
})

const queryState = ref({
  studentId: undefined as string | undefined,
  productId: undefined as string | undefined,
  classTeacherId: undefined as string | undefined,
  classIds: undefined as string[] | undefined,
  statusList: undefined as number[] | undefined,
})

const allColumns = ref([
  {
    title: '学员/性别',
    dataIndex: 'student',
    key: 'student',
    fixed: 'left',
    width: 180,
    required: true,
  },
  {
    title: '联系电话',
    dataIndex: 'phone',
    key: 'phone',
    width: 140,
  },
  {
    title: '当前状态',
    dataIndex: 'status',
    key: 'status',
    width: 120,
  },
  {
    title: '在读课程',
    dataIndex: 'lessonName',
    key: 'lessonName',
    width: 180,
  },
  {
    title: '班主任',
    dataIndex: 'classTeacherList',
    key: 'classTeacherList',
    width: 180,
  },
  {
    title: '剩余数量',
    dataIndex: 'remaining',
    key: 'remaining',
    width: 180,
  },
  {
    title: '到期时间',
    dataIndex: 'expireTime',
    key: 'expireTime',
    width: 140,
  },
])

const { selectedValues, columnOptions, filteredColumns, totalWidth } = useTableColumns({
  storageKey: 'pending-renewal-student-list',
  allColumns,
  excludeKeys: ['action'],
})

const rowSelection = computed(() => ({
  selectedRowKeys: selectedRowKeys.value,
  onChange: (keys: Array<string | number>, rows: PendingRenewalStudentItem[]) => {
    selectedRowKeys.value = keys.map(item => String(item))
    selectedRows.value = rows
  },
}))

const selectedCount = computed(() => selectedRowKeys.value.length)

function resetQueryState() {
  queryState.value.studentId = undefined
  queryState.value.productId = undefined
  queryState.value.classTeacherId = undefined
  queryState.value.classIds = undefined
  queryState.value.statusList = undefined
}

function normalizeStringValue(value: unknown) {
  if (value === undefined || value === null)
    return undefined
  const text = String(value).trim()
  return text || undefined
}

function normalizeStringArray(value: unknown) {
  if (!Array.isArray(value))
    return undefined
  const list = value
    .map(item => String(item ?? '').trim())
    .filter(Boolean)
  return list.length ? list : undefined
}

function normalizeStatusList(value: unknown) {
  if (!Array.isArray(value))
    return undefined
  const list = value
    .map(item => Number(item))
    .filter(item => Number.isFinite(item))
  return list.length ? list : undefined
}

async function loadClassNameOptions() {
  try {
    const lessonIds = queryState.value.productId ? [queryState.value.productId] : undefined
    const res = await pageGroupClassesApi({
      pageRequestModel: {
        needTotal: true,
        pageSize: 200,
        pageIndex: 1,
        skipCount: 0,
      },
      queryModel: {
        lessonIds,
      },
    })

    if (res.code !== 200) {
      throw new Error(res.message || '获取班级筛选项失败')
    }

    const list = Array.isArray(res.result?.list) ? res.result.list : []
    const optionMap = new Map<string, { id: string, value: string }>()
    list.forEach((item) => {
      const id = String(item?.id ?? '').trim()
      const value = String(item?.name ?? '').trim()
      if (!id || !value || optionMap.has(id))
        return
      optionMap.set(id, { id, value })
    })
    classNameOptionsData.value = [...optionMap.values()]
  }
  catch (error) {
    console.error('加载待续费班级筛选项失败:', error)
    classNameOptionsData.value = []
  }
}

function formatNumber(value?: number) {
  const num = Number(value || 0)
  if (!Number.isFinite(num))
    return '0'
  return Number.isInteger(num) ? String(num) : num.toFixed(2).replace(/\.?0+$/, '')
}

function formatMoney(value?: number) {
  const num = Number(value || 0)
  return num.toFixed(2)
}

function isAmountMode(mode?: number) {
  const value = Number(mode || 0)
  return value === 3 || value === 4
}

function getRemainingUnit(mode?: number) {
  const value = Number(mode || 0)
  if (value === 2)
    return '天'
  if (isAmountMode(value))
    return '元'
  return '课时'
}

function getGenderText(sex?: number) {
  const value = Number.isFinite(Number(sex)) ? Number(sex) : Sex.Unknown
  return SexLabel[value as Sex] || SexLabel[Sex.Unknown]
}

function getStatusInfo(status?: number) {
  const statusValue = Number(status || 0)
  const map: Record<number, { text: string, className: string }> = {
    1: { text: '正常', className: 'text-#0c3 bg-#e6ffec' },
    2: { text: '已停课', className: 'text-#f90 bg-#fff5e6' },
    3: { text: '已结课', className: 'text-#888 bg-#f5f5f5' },
  }
  return map[statusValue] || { text: '未知', className: 'text-#888 bg-#f5f5f5' }
}

function formatExpireDate(record: PendingRenewalStudentItem) {
  if (!record.enableExpireTime || !record.expireTime)
    return '-'
  const date = dayjs(record.expireTime)
  if (!date.isValid() || date.year() <= 1)
    return '-'
  return date.format('YYYY-MM-DD')
}

function getClassTeacherText(record: PendingRenewalStudentItem) {
  const teacherNames = Array.isArray(record.classTeacherList)
    ? record.classTeacherList
      .map(item => String(item?.name ?? '').trim())
      .filter(Boolean)
    : []
  return teacherNames.length ? teacherNames.join('、') : '-'
}

function getRemainingText(record: PendingRenewalStudentItem) {
  if (isAmountMode(record.lessonChargingMode))
      return `¥ ${formatMoney(record.tuition || record.leftQuantity)}`
  const total = Number(record.leftQuantity || 0) + Number(record.leftFreeQuantity || 0)
  return `${formatNumber(total)}${getRemainingUnit(record.lessonChargingMode)}`
}

function getRemainingSubText(record: PendingRenewalStudentItem) {
  if (isAmountMode(record.lessonChargingMode))
    return ''
  const freeQuantity = Number(record.leftFreeQuantity || 0)
  if (freeQuantity > 0) {
    return `正课${formatNumber(record.leftQuantity)}${getRemainingUnit(record.lessonChargingMode)} + 赠送${formatNumber(freeQuantity)}${getRemainingUnit(record.lessonChargingMode)}`
  }
  return ''
}

function handleMessageRecord() {
  messageRecordOpen.value = true
}

async function handleWechatRemind() {
  if (!selectedCount.value) {
    messageService.warning('请先选择待发送的学员')
    return
  }
  if (sendingWechatReminder.value)
    return
  sendingWechatReminder.value = true
  try {
    const res = await sendPendingRenewalWechatReminderApi({
      tuitionAccountIds: selectedRowKeys.value,
    })
    if (res.code !== 200) {
      throw new Error(res.message || '发送续费提醒失败')
    }
    const result = res.result
    const successCount = Number(result?.successCount || 0)
    const skippedCount = Number(result?.skippedCount || 0)
    const failedCount = Number(result?.failedCount || 0)

    if (successCount > 0) {
      messageService.success(`已发送 ${successCount} 条微信提醒${skippedCount > 0 ? `，未关注跳过 ${skippedCount} 条` : ''}${failedCount > 0 ? `，失败 ${failedCount} 条` : ''}`)
    }
    else if (skippedCount > 0 || failedCount > 0) {
      messageService.warning(`本次未成功发送${skippedCount > 0 ? `，未关注跳过 ${skippedCount} 条` : ''}${failedCount > 0 ? `，失败 ${failedCount} 条` : ''}`)
    }
    else {
      messageService.info('本次没有可发送的续费提醒')
    }

    selectedRowKeys.value = []
    selectedRows.value = []
    messageRecordOpen.value = true
    await nextTick()
    await messageRecordDrawerRef.value?.getList?.()
  }
  catch (error: any) {
    console.error('send pending renewal wechat reminder failed', error)
    messageService.error(error?.message || '发送续费提醒失败')
  }
  finally {
    sendingWechatReminder.value = false
  }
}

function handleSmsRemind() {
  if (!selectedCount.value) {
    messageService.warning('请先选择待发送的学员')
    return
  }
  messageService.info('短信提醒功能待接入')
}

async function getList(id?: string | number, type?: string) {
  loading.value = true
  try {
    const queryModel = Object.fromEntries(
      Object.entries(queryState.value).filter(([, value]) => value !== undefined),
    )

    const res = await getPendingRenewalStudentsPagedListApi({
      pageRequestModel: {
        needTotal: true,
        pageSize: pagination.value.pageSize,
        pageIndex: pagination.value.current,
        skipCount: 0,
      },
      queryModel,
      sortModel: {
        expriedTime: 0,
      },
    })

    if (res.code !== 200) {
      throw new Error(res.message || '获取待续费学员列表失败')
    }

    const result = res.result || {}
    dataSource.value = Array.isArray(result.list) ? result.list : []
    pagination.value.total = Number(result.total || 0)
    summary.value.total = Number(result.total || 0)
    summary.value.studentCount = Number(result.studentCount || 0)
    allFilterRef.value?.clearQuickFilter?.(id, type)
  }
  catch (error: any) {
    console.error('获取待续费学员列表失败:', error)
    messageService.error(error?.message || '获取待续费学员列表失败')
  }
  finally {
    loading.value = false
  }
}

function applyFilterChange(
  updates: Partial<typeof queryState.value>,
  id?: string | number,
  type?: string,
  options?: { reloadClassOptions?: boolean },
) {
  Object.assign(queryState.value, updates)
  pagination.value.current = 1
  selectedRowKeys.value = []
  selectedRows.value = []
  if (options?.reloadClassOptions)
    void loadClassNameOptions()
  void getList(id, type)
}

const filterUpdateHandlers = computed(() => ({
  'update:stuPhoneSearchFilter': (val: unknown, isClearAll?: boolean, id?: string | number, type?: string) => {
    if (isClearAll) {
      resetQueryState()
      void loadClassNameOptions()
      void getList(id, type)
      return
    }
    applyFilterChange({ studentId: normalizeStringValue(val) }, id, type)
  },
  'update:intentionCourseFilter': (val: unknown, isClearAll?: boolean, id?: string | number, type?: string) => {
    if (isClearAll) {
      resetQueryState()
      void loadClassNameOptions()
      void getList(id, type)
      return
    }
    applyFilterChange({ productId: normalizeStringValue(val) }, id, type, { reloadClassOptions: true })
  },
  'update:classTeacherFilter': (val: unknown, isClearAll?: boolean, id?: string | number, type?: string) => {
    if (isClearAll) {
      resetQueryState()
      void loadClassNameOptions()
      void getList(id, type)
      return
    }
    applyFilterChange({ classTeacherId: normalizeStringValue(val) }, id, type)
  },
  'update:classNameFilter': (val: unknown, isClearAll?: boolean, id?: string | number, type?: string) => {
    if (isClearAll) {
      resetQueryState()
      void loadClassNameOptions()
      void getList(id, type)
      return
    }
    applyFilterChange({ classIds: normalizeStringArray(val) }, id, type)
  },
  'update:currentStatusFilter': (val: unknown, isClearAll?: boolean, id?: string | number, type?: string) => {
    if (isClearAll) {
      resetQueryState()
      void loadClassNameOptions()
      void getList(id, type)
      return
    }
    applyFilterChange({ statusList: normalizeStatusList(val) }, id, type)
  },
}))

function handleTableChange(paginationInfo: any) {
  pagination.value.current = Number(paginationInfo?.current || 1)
  pagination.value.pageSize = Number(paginationInfo?.pageSize || pagination.value.pageSize)
  void getList()
}

onMounted(async () => {
  await loadClassNameOptions()
  await getList()
})

useStudentListRefresh(getList)

defineExpose({
  getList,
})
</script>

<template>
  <div>
    <div class="filter-wrap mt-2 bg-white pl-3 pr-3 rounded-4">
      <all-filter
        ref="allFilterRef"
        :display-array="displayArray"
        :is-quick-show="false"
        :is-show-search-stu-phonefilter="true"
        :class-name-options-data="classNameOptionsData"
        v-on="filterUpdateHandlers"
      />
    </div>

    <div class="student-list mt-2 pt-3 pb-3 pl-6 pr-6 bg-white rounded-4">
      <div class="tab-table">
        <div class="table-title flex justify-between">
          <div class="total">
            当前共{{ summary.studentCount }}名学员，{{ summary.total }}条待续费记录
            <span class="text-#0066ff">{{ tipsText }}</span>
          </div>
          <div class="edit flex">
            <a-button class="mr-2" @click="handleMessageRecord">
              消息记录
            </a-button>
            <a-dropdown class="mr-2">
              <template #overlay>
                <a-menu>
                  <a-menu-item key="1">
                    批量导出
                  </a-menu-item>
                  <a-menu-item key="2">
                    导出记录
                  </a-menu-item>
                </a-menu>
              </template>
              <a-button>
                导出数据
                <DownOutlined :style="{ fontSize: '10px' }" />
              </a-button>
            </a-dropdown>
            <a-dropdown class="mr-2">
              <template #overlay>
                <a-menu>
                  <a-menu-item key="1" @click="handleWechatRemind">
                    微信提醒
                  </a-menu-item>
                  <a-menu-item key="2" @click="handleSmsRemind">
                    短信提醒
                  </a-menu-item>
                </a-menu>
              </template>
              <a-button :loading="sendingWechatReminder">
                批量发送续费提醒{{ selectedCount > 0 ? `(${selectedCount})` : '' }}
                <DownOutlined :style="{ fontSize: '10px' }" />
              </a-button>
            </a-dropdown>
            <customize-code
              :checked-values="selectedValues"
              :options="columnOptions"
              :total="allColumns.length"
              :num="selectedValues.length"
              @update:checked-values="(val) => { selectedValues = val }"
            />
          </div>
        </div>

        <div class="table-content mt-2">
          <a-table
            :data-source="dataSource"
            :loading="loading"
            :pagination="pagination"
            :columns="filteredColumns"
            :row-selection="rowSelection"
            :scroll="{ x: totalWidth }"
            row-key="tuitionAccountId"
            size="small"
            @change="handleTableChange"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'student'">
                <student-avatar
                  :id="record.studentId"
                  :name="record.studentName || '-'"
                  :gender="getGenderText(record.sex)"
                  :avatar-url="record.avatar || ''"
                  :show-age="false"
                  default-active-key="0"
                />
              </template>

              <template v-if="column.key === 'phone'">
                <div class="text-#222">
                  {{ record.phone || '-' }}
                </div>
              </template>

              <template v-if="column.key === 'status'">
                <span :class="`${getStatusInfo(record.status).className} rounded-2.5 inline-block text-3 pt-0.5 pb-0.5 pl-2 pr-2`">
                  {{ getStatusInfo(record.status).text }}
                </span>
              </template>

              <template v-if="column.key === 'lessonName'">
                <div class="text-#222">
                  {{ record.lessonName || '-' }}
                </div>
              </template>

              <template v-if="column.key === 'classTeacherList'">
                <div class="text-#222">
                  {{ getClassTeacherText(record) }}
                </div>
              </template>

              <template v-if="column.key === 'remaining'">
                <div class="text-#222">
                  {{ getRemainingText(record) }}
                </div>
                <div v-if="getRemainingSubText(record)" class="text-3 text-#888">
                  {{ getRemainingSubText(record) }}
                </div>
              </template>

              <template v-if="column.key === 'expireTime'">
                {{ formatExpireDate(record) }}
              </template>
            </template>
          </a-table>
        </div>
      </div>
    </div>

    <pending-renewal-message-record-drawer
      ref="messageRecordDrawerRef"
      v-model:open="messageRecordOpen"
    />
  </div>
</template>

<style lang="less" scoped>
.total {
  position: relative;
  padding-left: 10px;
  color: #222;
  display: flex;
  align-items: center;

  &::before {
    display: inline-block;
    background: var(--pro-ant-color-primary);
    border-radius: 2px;
    content: "";
    height: 12px;
    left: 0;
    position: absolute;
    width: 4px;
  }
}
</style>
