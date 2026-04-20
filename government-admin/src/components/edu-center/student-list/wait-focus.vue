<script setup>
import { computed, onMounted, ref } from 'vue'
import { calculateAge } from '@/utils/date'
import { ParentRelationshipLabel, StudentStatus, StudentStatusLabel } from '@/enums'
import { getPendingAttentionStudentPagedListApi } from '~@/api/edu-center/student-list'
import messageService from '~@/utils/messageService'
import { handleDateRangeParams } from '~@/utils/dateRangeParams'
import { useStudentListRefresh } from '@/composables/useStudentListRefresh'

const displayArray = ref(['className', 'sex', 'age', 'studentStatus'])
const allFilterRef = ref(null)
const loading = ref(false)
const dataSource = ref([])
const selectedRows = ref([])
const selectedRowKeys = ref([])

const allColumns = [
  {
    title: '学员/性别/年龄',
    dataIndex: 'name',
    key: 'name',
    fixed: 'left',
    width: 180,
  },
  {
    title: '联系电话',
    dataIndex: 'mobile',
    key: 'mobile',
    width: 150,
  },
  {
    title: '学员状态',
    dataIndex: 'studentStatus',
    key: 'studentStatus',
    width: 130,
  },
  {
    title: '家校云',
    dataIndex: 'isBindChild',
    key: 'isBindChild',
    width: 120,
  },
]

const totalWidth = allColumns.reduce((width, column) => width + (column.width || 0), 0)

const queryState = ref({
  studentId: undefined,
  sexes: undefined,
  studentStatuses: [StudentStatus.Reading],
  classIds: undefined,
  ageMin: undefined,
  ageMax: undefined,
  age: undefined,
})

const pagination = ref({
  current: 1,
  pageSize: 10,
  total: 0,
  showSizeChanger: true,
  showTotal: total => `共 ${total} 条`,
  pageSizeOptions: ['10', '20', '50'],
  hideOnSinglePage: false,
  showQuickJumper: true,
})

const rowSelection = computed(() => ({
  selectedRowKeys: selectedRowKeys.value,
  onChange: (keys, rows) => {
    selectedRowKeys.value = keys
    selectedRows.value = rows
  },
}))

const searchStudentStatus = computed(() => {
  const statuses = Array.isArray(queryState.value.studentStatuses) ? queryState.value.studentStatuses : []
  return statuses.length === 1 ? Number(statuses[0]) : undefined
})

function resetQueryState() {
  queryState.value = {
    studentId: undefined,
    sexes: undefined,
    studentStatuses: [StudentStatus.Reading],
    classIds: undefined,
    ageMin: undefined,
    ageMax: undefined,
    age: undefined,
  }
}

function clearSelection() {
  selectedRowKeys.value = []
  selectedRows.value = []
}

function handleFilterUpdate(updates, isClearAll = false, id, type) {
  if (isClearAll) {
    resetQueryState()
  }
  else {
    Object.entries(updates).forEach(([key, value]) => {
      queryState.value[key] = value
    })
  }
  pagination.value.current = 1
  clearSelection()
  getList(queryState.value, id, type)
}

const filterFieldMapping = {
  classNameFilter: 'classIds',
  sexFilter: 'sexes',
  ageFilter: 'age',
  stuStatusFilter: 'studentStatuses',
  stuPhoneSearchFilter: 'studentId',
}

const filterUpdateHandlers = computed(() => {
  const handlers = {}
  Object.entries(filterFieldMapping).forEach(([eventKey, fieldName]) => {
    handlers[`update:${eventKey}`] = (val, isClearAll, id, type) => {
      if (fieldName === 'studentStatuses') {
        if (isClearAll) {
          handleFilterUpdate({}, true, id, type)
          return
        }
        if (val === undefined || val === null || val === '') {
          handleFilterUpdate({ studentStatuses: undefined }, false, id, type)
          return
        }
        handleFilterUpdate({ studentStatuses: [Number(val)] }, false, id, type)
        return
      }

      let finalVal = val
      if (Array.isArray(val) && val.length === 0) {
        finalVal = undefined
      }
      if (fieldName === 'classIds' && Array.isArray(finalVal)) {
        finalVal = finalVal.map(item => `${item}`)
      }
      if (fieldName === 'studentId' && finalVal !== undefined && finalVal !== null && finalVal !== '') {
        finalVal = String(finalVal)
      }
      handleFilterUpdate({ [fieldName]: finalVal }, isClearAll, id, type)
    }
  })
  return handlers
})

async function getList(newQueryParams = {}, id, type) {
  const dateRangeMappings = {
    age: {
      begin: 'ageMin',
      end: 'ageMax',
    },
  }

  loading.value = true
  try {
    let mergedQuery = { ...queryState.value }
    if (Object.keys(newQueryParams).length > 0) {
      mergedQuery = { ...mergedQuery, ...newQueryParams }
    }
    const normalizedQuery = handleDateRangeParams(mergedQuery, dateRangeMappings)
    Object.assign(queryState.value, normalizedQuery)
    const queryModel = Object.fromEntries(
      Object.entries(queryState.value)
        .filter(([key, value]) => key !== 'age' && value !== undefined && value !== null && value !== '' && (!Array.isArray(value) || value.length > 0)),
    )

    const res = await getPendingAttentionStudentPagedListApi({
      pageRequestModel: {
        pageSize: pagination.value.pageSize,
        pageIndex: pagination.value.current,
        needTotal: true,
      },
      queryModel,
    })

    if (res.code === 200) {
      dataSource.value = res.result || []
      pagination.value.total = res.total || 0
      allFilterRef.value?.clearQuickFilter(id, type)
      return
    }

    messageService.error(res.message || '获取待关注学员失败')
  }
  catch (error) {
    console.error('get pending attention students failed', error)
    messageService.error('获取待关注学员失败')
  }
  finally {
    loading.value = false
  }
}

function handleTableChange(paginationInfo) {
  pagination.value.current = paginationInfo.current
  pagination.value.pageSize = paginationInfo.pageSize
  getList()
}

function handleInvite(type) {
  if (selectedRowKeys.value.length === 0) {
    messageService.warning('请先选择学员')
    return
  }
  messageService.info(`${type}邀请功能开发中`)
}

function getStudentStatusClass(status) {
  return status === StudentStatus.History ? 'is-history' : 'is-reading'
}

useStudentListRefresh(getList)

onMounted(() => {
  getList()
})

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
        :default-student-status="StudentStatus.Reading"
        :student-status="searchStudentStatus"
        v-on="filterUpdateHandlers"
      />
    </div>

    <div class="student-list mt-2 rounded-4 bg-white px-6 pb-3 pt-4">
      <div class="table-title flex items-center justify-between">
        <div class="total">
          当前共{{ pagination.total || 0 }}名学员
          <span v-if="selectedRowKeys.length > 0" class="ml-2 text-blue-600">
            （已选中{{ selectedRowKeys.length }}名）
            <a-button type="link" size="small" class="ml-1 p-0" @click="clearSelection">
              清空选择
            </a-button>
          </span>
        </div>
        <div class="actions flex items-center">
          <a-button class="mr-2" @click="handleInvite('二维码')">
            二维码邀请
          </a-button>
          <a-button type="primary" @click="handleInvite('短信')">
            短信邀请
          </a-button>
        </div>
      </div>

      <div class="table-content mt-3">
        <a-table
          row-key="id"
          :data-source="dataSource"
          :loading="loading"
          :pagination="pagination"
          :columns="allColumns"
          :row-selection="rowSelection"
          :scroll="{ x: totalWidth }"
          :sticky="{ offsetHeader: 100 }"
          size="small"
          @change="handleTableChange"
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'name'">
              <student-avatar
                :id="record.id"
                :name="record.stuName"
                :gender="record.stuSex === 0 ? '女' : record.stuSex === 1 ? '男' : '未知'"
                :age="calculateAge(record.birthDay)"
                :avatar-url="record.avatarUrl"
                default-active-key="0"
              />
            </template>

            <template v-else-if="column.key === 'mobile'">
              <div class="contact-cell">
                <div v-if="record.phoneRelationship" class="relation">
                  {{ ParentRelationshipLabel[record.phoneRelationship] }}
                </div>
                <div class="phone">
                  {{ record.mobile || '-' }}
                </div>
              </div>
            </template>

            <template v-else-if="column.key === 'studentStatus'">
              <div class="status-cell" :class="getStudentStatusClass(record.studentStatus)">
                <span class="dot" />
                <span>{{ StudentStatusLabel[record.studentStatus] || '-' }}</span>
              </div>
            </template>

            <template v-else-if="column.key === 'isBindChild'">
              <div class="cloud-cell">
                <span class="unbound-text">未关注</span>
              </div>
            </template>
          </template>
        </a-table>
      </div>
    </div>
  </div>
</template>

<style lang="less" scoped>
.table-title {
  min-height: 32px;
}

.total {
  position: relative;
  display: flex;
  align-items: center;
  padding-left: 10px;
  color: #222;
  font-weight: 500;

  &::before {
    position: absolute;
    left: 0;
    display: inline-block;
    width: 4px;
    height: 12px;
    content: "";
    border-radius: 2px;
    background: var(--pro-ant-color-primary);
  }
}

.contact-cell {
  line-height: 1.5;

  .relation {
    color: #222;
  }

  .phone {
    color: #666;
    font-size: 12px;
  }
}

.status-cell {
  display: inline-flex;
  align-items: center;
  color: #222;

  .dot {
    width: 6px;
    height: 6px;
    margin-right: 6px;
    border-radius: 999px;
    background: #52c41a;
  }

  &.is-history .dot {
    background: #d9d9d9;
  }
}

.cloud-cell {
  display: inline-flex;
  align-items: center;
}

.unbound-text {
  color: #999;
}
</style>
