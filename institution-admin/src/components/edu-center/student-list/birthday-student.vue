<script setup>
import { computed, onMounted, ref } from 'vue'
import { DownOutlined } from '@ant-design/icons-vue'
import dayjs from 'dayjs'
import { ParentRelationshipLabel, StudentStatus, StudentStatusLabel } from '@/enums'
import { useStudentListRefresh } from '@/composables/useStudentListRefresh'
import { calculateAge } from '@/utils/date'
import { getBirthdayStudentPagedListApi } from '~@/api/edu-center/student-list'
import { handleDateRangeParams } from '~@/utils/dateRangeParams'
import messageService from '~@/utils/messageService'

const displayArray = ref(['birthday', 'birthMonth', 'sex', 'age', 'studentStatus'])
const allFilterRef = ref(null)
const allFilterKey = ref(0)
const loading = ref(false)
const dataSource = ref([])

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
    title: '生日日期',
    dataIndex: 'birthDay',
    key: 'birthDay',
    width: 140,
  },
]

const totalWidth = allColumns.reduce((width, column) => width + (column.width || 0), 0)

function createDefaultQueryState() {
  return {
    sexes: undefined,
    studentStatuses: [StudentStatus.Reading],
    birthMonth: undefined,
    birthDayBegin: undefined,
    birthDayEnd: undefined,
    ageMin: undefined,
    ageMax: undefined,
    birthday: undefined,
    age: undefined,
  }
}

const queryState = ref(createDefaultQueryState())

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

const searchStudentStatus = computed(() => {
  const statuses = Array.isArray(queryState.value.studentStatuses) ? queryState.value.studentStatuses : []
  return statuses.length === 1 ? Number(statuses[0]) : undefined
})

const showDefaultHint = computed(() => {
  const statuses = Array.isArray(queryState.value.studentStatuses) ? queryState.value.studentStatuses : []
  return !queryState.value.birthMonth
    && !queryState.value.birthDayBegin
    && !queryState.value.birthDayEnd
    && statuses.length === 1
    && Number(statuses[0]) === StudentStatus.Reading
})

function resetQueryState() {
  queryState.value = createDefaultQueryState()
}

function handleFilterUpdate(updates, isClearAll = false, id, type) {
  if (isClearAll) {
    resetQueryState()
    allFilterKey.value++
  }
  else {
    Object.entries(updates).forEach(([key, value]) => {
      queryState.value[key] = value
    })
  }

  pagination.value.current = 1
  void getList(queryState.value, id, type)
}

const filterFieldMapping = {
  birthdayFilter: 'birthday',
  birthMonthFilter: 'birthMonth',
  sexFilter: 'sexes',
  ageFilter: 'age',
  stuStatusFilter: 'studentStatuses',
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
      if (Array.isArray(val) && val.length === 0)
        finalVal = undefined

      handleFilterUpdate({ [fieldName]: finalVal }, isClearAll, id, type)
    }
  })

  return handlers
})

async function getList(newQueryParams = {}, id, type) {
  const dateRangeMappings = {
    birthday: {
      begin: 'birthDayBegin',
      end: 'birthDayEnd',
    },
    age: {
      begin: 'ageMin',
      end: 'ageMax',
    },
  }

  loading.value = true
  try {
    const mergedQuery = {
      ...queryState.value,
      ...newQueryParams,
    }
    const normalizedQuery = handleDateRangeParams(mergedQuery, dateRangeMappings)

    queryState.value = {
      ...queryState.value,
      birthday: undefined,
      birthDayBegin: undefined,
      birthDayEnd: undefined,
      age: undefined,
      ageMin: undefined,
      ageMax: undefined,
      ...normalizedQuery,
    }

    const queryModel = Object.fromEntries(
      Object.entries(queryState.value)
        .filter(([, value]) => value !== undefined && value !== null && value !== '' && (!Array.isArray(value) || value.length > 0)),
    )

    const res = await getBirthdayStudentPagedListApi({
      pageRequestModel: {
        pageSize: pagination.value.pageSize,
        pageIndex: pagination.value.current,
        needTotal: true,
      },
      queryModel,
    })

    if (res.code === 200) {
      dataSource.value = res.result || []
      pagination.value.total = Number(res.total || 0)
      allFilterRef.value?.clearQuickFilter(id, type)
      return
    }

    messageService.error(res.message || '获取生日学员失败')
  }
  catch (error) {
    console.error('get birthday students failed', error)
    messageService.error('获取生日学员失败')
  }
  finally {
    loading.value = false
  }
}

function handleTableChange(paginationInfo) {
  pagination.value.current = paginationInfo.current
  pagination.value.pageSize = paginationInfo.pageSize
  void getList()
}

function getStudentStatusClass(status) {
  if (status === StudentStatus.Intention)
    return 'is-intention'
  if (status === StudentStatus.History)
    return 'is-history'
  return 'is-reading'
}

function formatBirthday(birthDay) {
  if (!birthDay)
    return '-'
  return dayjs(birthDay).format('YYYY-MM-DD')
}

useStudentListRefresh(getList)

onMounted(() => {
  void getList()
})

defineExpose({
  getList,
})
</script>

<template>
  <div>
    <div class="filter-wrap mt-2 rounded-4 bg-white pl-3 pr-3">
      <all-filter
        :key="allFilterKey"
        ref="allFilterRef"
        :display-array="displayArray"
        :is-quick-show="false"
        :is-show-search-stu-phone="false"
        :default-student-status="StudentStatus.Reading"
        :student-status="searchStudentStatus"
        birthday-label="生日日期"
        :birthday-future-mode="true"
        v-on="filterUpdateHandlers"
      />
    </div>

    <div class="student-list mt-2 rounded-4 bg-white px-6 pb-3 pt-4">
      <div class="table-title flex items-center justify-between">
        <div class="total">
          当前共{{ pagination.total || 0 }}名学员
          <span v-if="showDefaultHint" class="ml-2 text-#0066ff">
            （默认展示未来1个月的在读生日学员）
          </span>
        </div>
        <div class="actions flex items-center">
          <a-dropdown>
            <template #overlay>
              <a-menu>
                <a-menu-item key="export">
                  批量导出
                </a-menu-item>
                <a-menu-item key="record">
                  导出记录
                </a-menu-item>
              </a-menu>
            </template>
            <a-button>
              导出数据
              <DownOutlined :style="{ fontSize: '10px' }" />
            </a-button>
          </a-dropdown>
        </div>
      </div>

      <div class="table-content mt-3">
        <a-table
          row-key="id"
          :data-source="dataSource"
          :loading="loading"
          :pagination="pagination"
          :columns="allColumns"
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

            <template v-else-if="column.key === 'birthDay'">
              <div class="birthday-cell">
                {{ formatBirthday(record.birthDay) }}
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

  &.is-intention .dot {
    background: #faad14;
  }

  &.is-history .dot {
    background: #d9d9d9;
  }
}

.birthday-cell {
  color: #222;
}
</style>
