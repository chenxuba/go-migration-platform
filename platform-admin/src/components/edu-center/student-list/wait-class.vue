<script setup lang="ts">
import dayjs from 'dayjs'
import { computed, reactive, ref } from 'vue'
import type { TableColumnsType } from 'ant-design-vue'
import type { RegisterReadInfo } from '@/api/edu-center/register-read-list'
import { getRegisterReadListApi } from '@/api/edu-center/register-read-list'
import ChooseGroupClassModal from './choose-group-class-modal.vue'
import { useStudentListRefresh } from '@/composables/useStudentListRefresh'
import messageService from '@/utils/messageService'

const displayArray = ['enrolledCourse', 'createTime', 'orNotFenClass']

const columns: TableColumnsType<RegisterReadInfo> = [
  {
    title: '学员/性别',
    dataIndex: 'studentName',
    key: 'studentName',
    width: 240,
  },
  {
    title: '联系电话',
    dataIndex: 'phone',
    key: 'phone',
    width: 180,
  },
  {
    title: '课程名称',
    dataIndex: 'lessonName',
    key: 'lessonName',
    width: 220,
  },
  {
    title: '授课类型',
    dataIndex: 'lessonType',
    key: 'lessonType',
    width: 140,
  },
  {
    title: '创建时间',
    dataIndex: 'createTime',
    key: 'createTime',
    width: 180,
  },
  {
    title: '操作',
    dataIndex: 'action',
    key: 'action',
    width: 100,
    fixed: 'right',
  },
]

const loading = ref(false)
const dataSource = ref<RegisterReadInfo[]>([])
const chooseClassOpen = ref(false)
const selectedRecord = ref<RegisterReadInfo | null>(null)

const queryState = reactive({
  productIds: undefined as string[] | undefined,
  createdTimeBegin: undefined as string | undefined,
  createdTimeEnd: undefined as string | undefined,
  hasAssignedClassCourse: undefined as boolean | undefined,
})

const pagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
  showSizeChanger: true,
  pageSizeOptions: ['10', '20', '50', '100'],
  hideOnSinglePage: false,
  showQuickJumper: true,
  showTotal: (total: number) => `共 ${total} 条`,
})

const totalWidth = computed(() =>
  columns.reduce((sum, item) => sum + Number(item.width || 0), 0),
)

function formatDateTime(value?: string) {
  if (!value)
    return '-'
  const date = dayjs(value)
  return date.isValid() ? date.format('YYYY-MM-DD HH:mm') : '-'
}

function getSexText(sex?: number) {
  if (sex === 1)
    return '男'
  if (sex === 0 || sex === 2)
    return '女'
  return '未知'
}

function getLessonTypeText(lessonType?: number) {
  if (lessonType === 1)
    return '班级授课'
  if (lessonType === 2)
    return '1对1授课'
  return '-'
}

async function getList() {
  loading.value = true
  try {
    const res = await getRegisterReadListApi({
      pageRequestModel: {
        needTotal: true,
        pageSize: pagination.pageSize,
        pageIndex: pagination.current,
        skipCount: (pagination.current - 1) * pagination.pageSize,
      },
      queryModel: {
        assignedClass: false,
        lessonType: 1,
        statusList: [1, 2],
        productIds: queryState.productIds,
        createdTimeBegin: queryState.createdTimeBegin,
        createdTimeEnd: queryState.createdTimeEnd,
        hasAssignedClassCourse: queryState.hasAssignedClassCourse,
      },
    })

    if (res.code !== 200) {
      throw new Error(res.message || '获取待分班学员失败')
    }

    dataSource.value = Array.isArray(res.result?.studentTutionAccounts) ? res.result.studentTutionAccounts : []
    pagination.total = Number(res.result?.total || 0)
  }
  catch (error: any) {
    dataSource.value = []
    pagination.total = 0
    messageService.error(error?.response?.data?.message || error?.message || '获取待分班学员失败')
  }
  finally {
    loading.value = false
  }
}

function reloadFirstPage() {
  pagination.current = 1
  getList()
}

function handleCourseFilter(val?: Array<string | number>, _isClearAll?: boolean) {
  queryState.productIds = Array.isArray(val) && val.length > 0 ? val.map(item => String(item)) : undefined
  reloadFirstPage()
}

function handleCreateTimeFilter(val?: string[], _isClearAll?: boolean) {
  if (Array.isArray(val) && val.length >= 2) {
    queryState.createdTimeBegin = val[0] || undefined
    queryState.createdTimeEnd = val[1] || undefined
  }
  else {
    queryState.createdTimeBegin = undefined
    queryState.createdTimeEnd = undefined
  }
  reloadFirstPage()
}

function handleHasAssignedClassCourseFilter(val?: number[], _isClearAll?: boolean) {
  if (Array.isArray(val) && val.length > 0) {
    if (val.includes(0) && val.includes(1))
      queryState.hasAssignedClassCourse = undefined
    else if (val.includes(0))
      queryState.hasAssignedClassCourse = false
    else if (val.includes(1))
      queryState.hasAssignedClassCourse = true
    else
      queryState.hasAssignedClassCourse = undefined
  }
  else {
    queryState.hasAssignedClassCourse = undefined
  }
  reloadFirstPage()
}

function handleTableChange(page: { current?: number, pageSize?: number }) {
  const nextCurrent = Number(page.current || 1)
  const nextSize = Number(page.pageSize || pagination.pageSize)
  const changed = nextCurrent !== pagination.current || nextSize !== pagination.pageSize
  pagination.current = nextCurrent
  pagination.pageSize = nextSize
  if (changed)
    getList()
}

function handleChooseClass(record: RegisterReadInfo) {
  selectedRecord.value = record
  chooseClassOpen.value = true
}

function handleAssignSuccess() {
  chooseClassOpen.value = false
  selectedRecord.value = null
  getList()
}

defineExpose({ getList })

useStudentListRefresh(() => {
  getList()
})

getList()
</script>

<template>
  <div>
    <div class="filter-wrap mt-2 bg-white pl-3 pr-3 rounded-4">
      <all-filter
        :display-array="displayArray"
        :is-quick-show="false"
        :is-show-search-stu-phone="false"
        or-not-fen-class-label="学员是否有已分班课程"
        :or-not-fen-class-options-override="[{ id: 0, value: '否' }, { id: 1, value: '是' }]"
        @update:enrolledCourseFilter="handleCourseFilter"
        @update:createTimeFilter="handleCreateTimeFilter"
        @update:orNotFenClassFilter="handleHasAssignedClassCourseFilter"
      />
    </div>

    <div class="student-list mt-2 pt-3 pb-3 pl-6 pr-6 bg-white rounded-4">
      <div class="table-title">
        <div class="total">
          当前共 {{ pagination.total }} 条数据
        </div>
      </div>

      <div class="table-content mt-2">
        <a-table
          row-key="tuitionAccountId"
          size="small"
          :loading="loading"
          :columns="columns"
          :data-source="dataSource"
          :pagination="pagination"
          :scroll="{ x: totalWidth }"
          @change="handleTableChange"
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'studentName'">
              <div class="student-cell">
                <student-avatar
                  :id="record.studentId"
                  :name="record.studentName || '-'"
                  :gender="getSexText(record.sex)"
                  :avatar-url="record.avatar || 'https://pcsys.admin.ybc365.com/c04d0ea2-a8b0-4001-b19b-946a980cb726.png'"
                  :show-age="false"
                  default-active-key="0"
                />
              </div>
            </template>

            <template v-else-if="column.key === 'phone'">
              <div class="text-#333">
                {{ record.phone || '-' }}
              </div>
            </template>

            <template v-else-if="column.key === 'lessonName'">
              {{ record.lessonName || '-' }}
            </template>

            <template v-else-if="column.key === 'lessonType'">
              {{ getLessonTypeText(record.lessonType) }}
            </template>

            <template v-else-if="column.key === 'createTime'">
              {{ formatDateTime(record.createTime) }}
            </template>

            <template v-else-if="column.key === 'action'">
              <a-button type="link" size="small" @click="handleChooseClass(record)">
                去分班
              </a-button>
            </template>
          </template>
        </a-table>
      </div>
    </div>

    <ChooseGroupClassModal
      v-model:open="chooseClassOpen"
      :record="selectedRecord"
      @success="handleAssignSuccess"
    />
  </div>
</template>

<style lang="less" scoped>
.total {
  position: relative;
  display: flex;
  align-items: center;
  padding-left: 10px;
  color: #222;

  &::before {
    position: absolute;
    left: 0;
    width: 4px;
    height: 12px;
    border-radius: 2px;
    background: var(--pro-ant-color-primary);
    content: '';
  }
}

.student-cell {
  display: flex;
  flex-direction: column;
}
</style>
