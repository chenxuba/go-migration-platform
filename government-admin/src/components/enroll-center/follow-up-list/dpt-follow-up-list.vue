<script setup>
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { debounce } from 'lodash-es'
import { ExclamationCircleOutlined } from '@ant-design/icons-vue'
import dayjs from 'dayjs'
import { getFollowRecordCountApi, getFollowUpRecordPagedApi, updateVisitStatusApi } from '~@/api/enroll-center/intention-student'
import { useTableColumns } from '@/composables/useTableColumns'
import emitter, { EVENTS } from '~@/utils/eventBus'
import { FollowMethodLabel, ParentRelationshipLabel, StudentStatusLabel } from '@/enums'
import { handleDateRangeParams } from '~@/utils/dateRangeParams'
import { useUserStore } from '~@/stores/user'

const dataSource = ref([])
const allColumns = ref([
  {
    title: '学员/性别',
    dataIndex: 'name',
    key: 'name',
    fixed: 'left',
    width: 140,
    required: true, // 新增必选标识
  },
  {
    title: '联系电话',
    dataIndex: 'phone',
    width: 140,
    key: 'phone',
  },
  {
    title: '跟进人',
    dataIndex: 'createUser',
    key: 'createUser',
    width: 120,
  },
  {
    title: '跟进内容',
    dataIndex: 'content',
    key: 'content',
    width: 280,
  },
  {
    title: '学员状态',
    dataIndex: 'studentStatus',
    key: 'studentStatus',
    width: 120,
  },
  {
    title: '渠道分类',
    dataIndex: 'categoryName',
    key: 'categoryName',
    width: 130,
  },
  {
    title: '渠道',
    dataIndex: 'channel',
    key: 'channel',
    width: 130,

  },
  {
    title: '图片/音频',
    dataIndex: 'followImages',
    key: 'followImages',
    width: 100,
  },
  {
    title: '意向课程',
    key: 'intentionCourse',
    dataIndex: 'intentionCourse',
    width: 120,

  },
  {
    title: '跟进时间',
    dataIndex: 'followUpTime',
    key: 'followUpTime',
    width: 160,
    sorter: true,

  },
  {
    title: '跟进方式',
    dataIndex: 'followMethod',
    key: 'followMethod',
    width: 100,
  },
  {
    title: '销售员',
    dataIndex: 'teacher',
    key: 'teacher',
    width: 100,

  },
  {
    title: '下次跟进',
    dataIndex: 'nextFollowUpTime',
    key: 'nextFollowUpTime',
    fixed: 'right',
    width: 160,
    sorter: true,

  },
  {
    title: '回访状态',
    dataIndex: 'visitStatus',
    key: 'visitStatus',
    fixed: 'right',
    width: 120,
  },
  {
    title: '操作',
    key: 'action',
    fixed: 'right',
    width: 120,
  },
])

const { selectedValues, columnOptions, filteredColumns, totalWidth }
  = useTableColumns({
    storageKey: 'follow-up-list', // 本地存储键名
    allColumns, // 原始列配置
    excludeKeys: ['action'], // 需要排除的列键
    defaultSelectedKeys: ['name', 'phone', 'createUser', 'content', 'studentStatus', 'categoryName', 'channel', 'followImages', 'intentionCourse', 'followUpTime', 'followMethod', 'teacher', 'nextFollowUpTime', 'visitStatus'],
  })

const userStore = useUserStore()
const displayArray = ref(['stuPhoneSearch', 'followTime', 'nextFollowTime', 'visitStatus', 'sex', 'channelCategory', 'followMethod', 'stuStatus', 'department'])
const loading = ref(false)
// 定义字段映射关系
const fieldMappings = {
  followUpTime: ['followUpTime', 'followUpTimeBegin', 'followUpTimeEnd'],
  nextFollowTime: ['nextFollowTime', 'nextFollowUpTimeBegin', 'nextFollowUpTimeEnd'],
}

// 存储所有查询条件的响应式对象
const queryState = ref({
  queryAllOrDepartment: 2,
  deptId: userStore.deptIds?.[0] || 0, // 部门id
  followUpTypes: undefined,
  visitStatuses: undefined,
  studentStatuses: undefined,
  quickFilter: undefined,
  intentionLevels: undefined,
  followUpTimeBegin: undefined,
  followUpTimeEnd: undefined,
  nextFollowUpTimeBegin: undefined,
  nextFollowUpTimeEnd: undefined,
  sexes: undefined,
  channelIds: undefined,
  // 原始字段
  followUpTime: undefined,
  nextFollowTime: undefined,
  // 可以添加更多查询条件...
})

// 重置所有查询条件
function resetQueryState() {
  Object.keys(queryState.value).forEach((key) => {
    if (key !== 'deptId' && key !== 'queryAllOrDepartment') {
      queryState.value[key] = undefined
    }
  })
}

// 使用防抖处理所有筛选条件的更新
const handleFilterUpdate = debounce((updates, isClearAll = false, id, type) => {
  if (isClearAll) {
    // 如果是清空所有，重置所有查询条件
    resetQueryState()
  }
  else {
    // 处理更新
    Object.entries(updates).forEach(([key, value]) => {
      if (Array.isArray(value) && value.length === 0 && fieldMappings[key]) {
        // 如果是空数组且在映射中存在，清除所有相关字段
        fieldMappings[key].forEach((field) => {
          queryState.value[field] = undefined
        })
      }
      else {
        queryState.value[key] = value
      }
    })
  }

  pagination.value.current = 1
  getFollowUpRecordPaged(queryState.value, id, type)
}, 300, { leading: true, trailing: false })

// 分页参数
const pagination = ref({
  current: 1,
  pageSize: 50,
  total: 0,
  showSizeChanger: true,
  showTotal: total => `共 ${total} 条`,
  pageSizeOptions: ['5', '10', '20', '50'],
  hideOnSinglePage: true,
  showQuickJumper: true,
})

// 排序状态管理
const sortModel = ref({
  byFollowUpTime: 0,
  byNextFlowTime: 0,
})

// 获取跟进记录列表
async function getFollowUpRecordPaged(newQueryParams = {}, id, type) {
  // 定义时间范围字段映射
  const dateRangeMappings = {
    followUpTime: {
      begin: 'followUpTimeBegin',
      end: 'followUpTimeEnd',
    },
    nextFollowTime: {
      begin: 'nextFollowUpTimeBegin',
      end: 'nextFollowUpTimeEnd',
    },

  }

  loading.value = true
  try {
    // 先清除 queryState 中的所有时间范围字段
    Object.values(dateRangeMappings).forEach(({ begin, end }) => {
      queryState.value[begin] = undefined
      queryState.value[end] = undefined
    })

    // 如果有新的查询参数，则处理时间范围
    if (Object.keys(newQueryParams).length > 0) {
      newQueryParams = handleDateRangeParams(newQueryParams, dateRangeMappings)
    }

    // 合并新的查询参数到queryState
    Object.assign(queryState.value, newQueryParams)

    // 过滤掉undefined的值和原始字段，只传递有效的查询条件
    const originalFields = ['followUpTime', 'nextFollowTime']
    const validQueryParams = Object.fromEntries(
      Object.entries(queryState.value)
        .filter(([key, value]) => value !== undefined && !originalFields.includes(key)),
    )

    const res = await getFollowUpRecordPagedApi({
      'pageRequestModel': {
        'pageSize': pagination.value.pageSize,
        'pageIndex': pagination.value.current,
      },
      'sortModel': sortModel.value,
      'queryModel': {
        ...validQueryParams, // 展开所有有效的查询条件
      },
    })
    dataSource.value = res.result || []
    pagination.value.total = res.total
    allFilterRef.value.clearQuickFilter(id, type)

    getFollowUpCount()
  }
  catch (error) {
    console.log('error: ', error)
  }
  finally {
    loading.value = false
  }
}

// 处理表格排序变化
function handleTableChange(paginationInfo, filters, sorter) {
  // 重置所有排序字段为0
  Object.keys(sortModel.value).forEach((key) => {
    sortModel.value[key] = 0
  })

  // 处理排序逻辑（支持单列排序和多列排序）
  const sortFieldMap = {
    'followUpTime': 'byFollowUpTime',
    'nextFollowUpTime': 'byNextFlowTime',
  }

  // 如果是数组，说明是多列排序，我们只取第一个（因为后端只支持单列排序）
  const currentSorter = Array.isArray(sorter) ? sorter[0] : sorter

  if (currentSorter && currentSorter.order) {
    const sortField = sortFieldMap[currentSorter.field]
    if (sortField) {
      // ascend: 升序(1), descend: 降序(2)
      sortModel.value[sortField] = currentSorter.order === 'ascend' ? 1 : 2
    }
  }

  // 更新分页信息
  pagination.value.current = paginationInfo.current
  pagination.value.pageSize = paginationInfo.pageSize

  // 重新获取数据
  getFollowUpRecordPaged()
}

const followUpCount = ref({})
// 获取跟进数量 getFollowRecordCountApi
async function getFollowUpCount() {
  try {
    const res = await getFollowRecordCountApi({ queryAllOrDepartment: 2, deptId: queryState.value.deptId })
    followUpCount.value = res.result
  }
  catch (error) {
    console.log('error: ', error)
  }
}

const allFilterRef = ref(null)

// 过滤器字段映射
const filterFieldMapping = {
  quickFilter: 'quickFilter',
  intentionLevelFilter: 'intentionLevels',
  followTimeFilter: 'followUpTime',
  nextFollowTimeFilter: 'nextFollowTime',
  sexFilter: 'sexes',
  followMethodFilter: 'followUpTypes',
  channelFilter: 'channelIds',
  stuPhoneSearchFilter: 'studentId',
  stuStatusFilter: 'studentStatuses',
  visitStatusFilter: 'visitStatuses',
  departmentFilter: 'deptId',
}

// 生成所有过滤器的更新处理器
const filterUpdateHandlers = computed(() => {
  const handlers = {}
  Object.entries(filterFieldMapping).forEach(([eventKey, fieldName]) => {
    handlers[`update:${eventKey}`] = (val, isClearAll, id, type) =>
      handleFilterUpdate({ [fieldName]: val }, isClearAll, id, type)
  })
  return handlers
})

function formatTime(time) {
  return time ? dayjs(time).format('YYYY-MM-DD HH:mm') : '-'
}

function formatFollowUpIntendedCourseText(record) {
  if (record?.intendedCourseName?.length)
    return record.intendedCourseName.join('、')
  if (record?.intentionLessonList?.length)
    return record.intentionLessonList.map(l => l.lessonName).join('、')
  if (record?.intentionCourse)
    return typeof record.intentionCourse === 'string' ? record.intentionCourse : String(record.intentionCourse)
  if (record?.intendedCourse?.length)
    return record.intendedCourse.join('、')
  return ''
}

// 处理跟进图片/音频数据
function formatFollowImages(followImages) {
  if (!followImages)
    return '-'

  try {
    // 如果是字符串，尝试解析为JSON
    const images = typeof followImages === 'string' ? JSON.parse(followImages) : followImages

    // 检查是否为数组且有长度
    if (Array.isArray(images) && images.length > 0) {
      return `${images.length}个`
    }

    return '-'
  }
  catch (error) {
    console.warn('解析followImages失败:', error)
    return '-'
  }
}

function isFollowUpVisited(record) {
  const v = record?.visitStatus
  return v === true || v === 1 || v === '1'
}

function isFollowUpNotVisited(record) {
  if (!record?.nextFollowUpTime)
    return false
  const v = record.visitStatus
  if (v === undefined || v === null)
    return true
  return v === false || v === 0 || v === '0'
}

function isOverdue(record) {
  if (!record?.nextFollowUpTime)
    return false
  if (isFollowUpVisited(record))
    return false

  const targetTime = dayjs(record.nextFollowUpTime)
  const now = dayjs()

  return targetTime.isValid() && targetTime.isBefore(now)
}

const openDrawer = ref(false)
function handleSeeStuData() {
  openDrawer.value = true
}

const handleMarkAsVisited = debounce(async (item) => {
  try {
    loading.value = true
    const res = await updateVisitStatusApi({
      'id': item.id,
      'uuid': item.uuid,
      'version': item.version,
      'visitStatus': true,
    })
    if (res.code === 200) {
      getFollowUpRecordPaged()
    }
  }
  catch (error) {
    console.log(error)
  }
}, 300, {
  leading: true,
  trailing: false,
})

const handleMarkAsUnvisited = debounce(async (item) => {
  try {
    loading.value = true
    const res = await updateVisitStatusApi({
      'id': item.id,
      'uuid': item.uuid,
      'version': item.version,
      'visitStatus': false,
    })
    if (res.code === 200) {
      getFollowUpRecordPaged()
    }
  }
  catch (error) {
    console.log(error)
  }
}, 300, {
  leading: true,
  trailing: false,
})

function handleSeeFollowImages(data) {
  record.value = JSON.parse(data.followImages) || []
  openPerviewImageModal.value = true
}
const openPerviewImageModal = ref(false)
const record = ref([])

onMounted(() => {
  getFollowUpRecordPaged()
  // 监听刷新列表事件
  emitter.on(EVENTS.REFRESH_STUDENT_LIST, getFollowUpRecordPaged)
})

onUnmounted(() => {
  emitter.off(EVENTS.REFRESH_STUDENT_LIST, getFollowUpRecordPaged)
})
</script>

<template>
  <div class="tab-content">
    <all-filter ref="allFilterRef" type="dpt" :select-dpt-vals="userStore.deptIds?.[0] || 0"
      :display-array="displayArray" :is-quick-show="true" :student-status="0" :follow-up-count="followUpCount"
      :hide-quick-filters="[2]" :custom-quick-filter-values="{ 3: 2 }" v-on="filterUpdateHandlers" />
    <div class="tab-table">
      <div class="table-title flex justify-between flex-items-center">
        <div class="total whitespace-nowrap mr-12px">
          当前共{{ dataSource.length || 0 }}条跟进记录, 跟进学员 1 人，0 条已标记回访，1 条未回访
        </div>
        <div class="edit flex overflow-x-auto">
          <a-button class="mr-2">
            导出数据
          </a-button>
          <!-- 自定义字段 -->
          <customize-code v-model:checked-values="selectedValues" :options="columnOptions"
            :total="allColumns.length - 1" :num="selectedValues.length" />
        </div>
      </div>
      <div class="table-content mt-2">
        <a-table :data-source="dataSource" row-key="id" :loading="loading" :pagination="pagination"
          :columns="filteredColumns" :scroll="{ x: totalWidth }" size="small" @change="handleTableChange">
          <template #headerCell="{ column }">
            <template v-if="column.key === 'visitStatus'">
              <span class="mr-1">{{ column.title }}</span>
              <a-popover color="#fff" title="回访状态">
                <template #content>
                  添加下次跟进时间后，后续标记回访的状态
                </template>
                <ExclamationCircleOutlined />
              </a-popover>
            </template>
          </template>
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'name'">
              <student-avatar :id="record.studentId" :name="record.stuName"
                :gender="record.stuSex == 1 ? '男' : record.stuSex == 0 ? '女' : '未知'" :avatar-url="record.avatarUrl"
                :show-age="false" default-active-key="4" />
            </template>
            <template v-if="column.key === 'phone'">
              <div class="name">
                <div class="text-#222">
                  {{ ParentRelationshipLabel[record.phoneRelationship] || '爸爸' }}
                </div>
                <div class="text-3 text-#666">
                  {{ record.mobile || '176****1636' }}
                </div>
              </div>
            </template>
            <template v-if="column.key === 'intentionCourse'">
              <clamped-text :text="formatFollowUpIntendedCourseText(record) || '-'" />
            </template>
            <!-- 跟进内容 -->
            <template v-if="column.key === 'content'">
              <div class="w-90%">
                <clamped-text :text="record.content" />
              </div>
            </template>
            <template v-if="column.key === 'studentStatus'">
              <div class="intention">
                <a-badge status="processing" />
                {{ StudentStatusLabel[record.studentStatus] }}
              </div>
            </template>
            <template v-if="column.key === 'createUser'">
              <clamped-text :text="record.createName" />
            </template>
            <!-- 销售员 -->
            <template v-if="column.key === 'teacher'">
              <clamped-text :text="record.salePersonName || '-'" />
            </template>
            <!-- 图片/音频 -->
            <template v-if="column.key === 'followImages'">
              <template v-if="record.followImages && formatFollowImages(record.followImages) !== '-'">
                <a-tooltip title="点击可查看具体内容">
                  <span class="cursor-pointer hover:text-#06f" @click="handleSeeFollowImages(record)">{{
                    formatFollowImages(record.followImages) }}</span>
                </a-tooltip>
              </template>
              <span v-else>{{ formatFollowImages(record.followImages) }}</span>
            </template>
            <!-- 跟进方式 -->
            <template v-if="column.key === 'followMethod'">
              {{ FollowMethodLabel[record.followMethod] || '-' }}
            </template>
            <!-- 跟进时间 -->
            <template v-if="column.key == 'followUpTime'">
              <clamped-text :text="formatTime(record.followUpTime) || '-'" />
            </template>
            <!-- 下次跟进 -->
            <template v-if="column.key == 'nextFollowUpTime'">
              <clamped-text :class="{ 'text-#f90': isOverdue(record) }"
                :text="formatTime(record.nextFollowUpTime) || '-'" />
            </template>
            <!-- 渠道分类 -->
            <template v-if="column.key === 'categoryName'">
              <clamped-text :text="record.categoryName || '-'" />
            </template>
            <!-- 渠道 -->
            <template v-if="column.key === 'channel'">
              <clamped-text :text="record.channelName || '-'" />
            </template>
            <!-- 回访状态（兼容 bool / 数字） -->
            <template v-if="column.key === 'visitStatus'">
              <template v-if="record.nextFollowUpTime">
                <span
                  v-if="isFollowUpNotVisited(record)"
                  class="text-#ff3333 bg-#ffe6e6 px1.5 py0.5 mx2 rounded-10 text-12px"
                >未回访</span>
                <span
                  v-else-if="isFollowUpVisited(record)"
                  class="text-#0c3 bg-#e6ffec px1.5 py0.5 mx2 rounded-10 text-12px"
                >已回访</span>
                <span v-else>-</span>
              </template>
              <span v-else>-</span>
            </template>
            <template v-else-if="column.key === 'action'">
              <span class="flex action">
                <a-popconfirm
                  v-if="isFollowUpNotVisited(record)" title="标记为已回访？"
                  @confirm="handleMarkAsVisited(record)"
                >
                  <a>标记已回访</a>
                </a-popconfirm>
                <a-popconfirm
                  v-else-if="isFollowUpVisited(record)" title="取消标记已回访？"
                  @confirm="handleMarkAsUnvisited(record)"
                >
                  <a>取消回访</a>
                </a-popconfirm>
                <span v-else>-</span>
              </span>
            </template>
          </template>
        </a-table>
      </div>
    </div>
  </div>
  <student-info-drawer v-model:open="openDrawer" default-active-key="4" />
  <perview-image-modal v-model:open="openPerviewImageModal" :record="record" />
</template>

<style lang="less" scoped>
.tab-content {
  .tab-table {
    background: #fff;
    margin-top: 8px;
    padding: 12px;
    border-radius: 12px;

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
        margin-top: 1px;
        position: absolute;
        width: 4px;
      }
    }

    .intention {
      display: flex;
      align-items: center;

      .statusTag {
        padding: 0 10px;
        height: 24px;
        background-color: rgb(255, 245, 230);
        color: rgb(255, 153, 0);
        border-radius: 100px;
        display: flex;
        align-items: center;
        justify-content: center;
        font-size: 12px;
        font-family: PingFangSC-Regular, PingFang SC, sans-serif;
        font-weight: 600;
      }
    }

    .intentionTag {
      display: inline-block;
      width: 6px;
      height: 6px;
      background: #f33;
      border-radius: 100px;
      margin-right: 3px;
    }

    .action {
      a {
        color: var(--pro-ant-color-primary);
        ;
      }
    }
  }
}

.hover {
  &:hover {
    .name {
      color: #06f;
    }
  }
}
</style>
