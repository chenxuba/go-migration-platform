<script setup>
import { CaretDownOutlined, DownOutlined } from '@ant-design/icons-vue'
import dayjs from 'dayjs'
import { debounce } from 'lodash-es'
import {
  groupClassStatisticsApi,
  pageGroupClassesApi,
} from '@/api/edu-center/group-class'
import CreateClassModal from '@/components/common/create-class-modal.vue'
import ClassAddStudentModal from '@/components/edu-center/class-list/class-add-student-modal.vue'
import ClassListDrawer from '@/components/edu-center/class-list/class-list-drawer.vue'
import { useTableColumns } from '@/composables/useTableColumns'
import messageService from '@/utils/messageService'

const defaultOpenClassStatus = 1

const allFilterRef = ref()
const createClassModal = ref(false)
const editClassRecord = ref(null)
const classListDrawerFlag = ref(false)
const addStudentModalOpen = ref(false)
const addStudentModalTitle = ref('')
const addStudentModalLessonName = ref('')
const addStudentModalClassId = ref('')
const addStudentModalLessonId = ref('')
const listLoading = ref(false)
const dataSource = ref([])
const selectedRowKeys = ref([])
const selectedRows = ref([])

const displayArray = ref([
  'customSearch',
  'classTeacher',
  'createUser',
  'doYouSchedule',
  'openClassStatus',
  'classEndingTime',
  'createTime',
  'salesPerson',
])

const customSearchFilters = ref([
  {
    id: 'lessonKey',
    fieldKey: '关联课程',
    fieldType: 4,
    optionsList: [],
  },
  {
    id: 'classRoomName',
    fieldKey: '上课教室',
    fieldType: 4,
    optionsList: [],
  },
  {
    id: 'courseType',
    fieldKey: '关联课程类型',
    fieldType: 4,
    optionsList: [
      { id: 'single', value: '课程' },
      { id: 'compose', value: '组合课' },
    ],
  },
])

const stats = ref({
  classCount: 0,
  openClassCount: 0,
  studentCount: 0,
  studentPersonTime: 0,
})

const pagination = reactive({
  current: 1,
  pageSize: 20,
  total: 0,
  showSizeChanger: true,
  showTotal: total => `共 ${total} 条`,
})

const queryState = ref({
  classIds: undefined,
  lessonKey: undefined,
  teacherId: undefined,
  defaultTeacherId: undefined,
  classRoomName: undefined,
  courseType: undefined,
  isScheduled: undefined,
  statues: [defaultOpenClassStatus],
  className: undefined,
  createdStaffIds: undefined,
  createdTime: undefined,
  closedTime: undefined,
})

function resetQueryState() {
  Object.keys(queryState.value).forEach((key) => {
    queryState.value[key] = undefined
  })
}

function updateCustomSearchOptions(id, optionsList) {
  customSearchFilters.value = customSearchFilters.value.map(item => item.id === id
    ? { ...item, optionsList }
    : item)
}

/**
 * 不再单独请求课程/组合课/额外班级 page：从当前列表结果合并筛选项，随翻页、换筛选累积。
 * （未出现在已加载班级里的课程/教室不会出现在下拉里，若需要全量可后续做懒加载或独立轻量接口。）
 */
function mergeCustomFilterOptionsFromClassList(list) {
  if (!Array.isArray(list) || list.length === 0)
    return

  const lessonItem = customSearchFilters.value.find(f => f.id === 'lessonKey')
  const roomItem = customSearchFilters.value.find(f => f.id === 'classRoomName')
  const lessonMap = new Map((lessonItem?.optionsList || []).map(o => [o.id, o]))
  const roomMap = new Map((roomItem?.optionsList || []).map(o => [o.id, o]))

  for (const item of list) {
    const roomName = String(item.classRoomName || '').trim()
    if (roomName && !roomMap.has(roomName)) {
      roomMap.set(roomName, {
        id: roomName,
        value: roomName,
      })
    }
    const lid = String(item.lessonId ?? '').trim()
    if (lid) {
      const optId = item.isMultiProduct ? `compose:${lid}` : `single:${lid}`
      if (!lessonMap.has(optId)) {
        lessonMap.set(optId, {
          id: optId,
          value: item.lessonName || lid,
        })
      }
    }
  }

  updateCustomSearchOptions('lessonKey', Array.from(lessonMap.values()))
  updateCustomSearchOptions('classRoomName', Array.from(roomMap.values()))
}

const handleFilterUpdate = debounce((updates = {}, isClearAll = false, id, type) => {
  if (isClearAll) {
    resetQueryState()
  }
  else {
    Object.entries(updates).forEach(([key, value]) => {
      queryState.value[key] = value
    })
  }

  pagination.current = 1
  selectedRows.value = []
  selectedRowKeys.value = []
  getClassList(queryState.value, id, type)
}, 200, { leading: true, trailing: false })

const filterUpdateHandlers = computed(() => ({
  'update:customSearchInputFilter': (payload, isClearAll, id, type) => {
    if (isClearAll) {
      handleFilterUpdate({}, true, id, type)
      return
    }

    const fieldId = id || payload?.item?.id
    const value = payload?.value
    if (fieldId === 'lessonKey') {
      handleFilterUpdate({ lessonKey: value || undefined }, false, id, type)
      return
    }
    if (fieldId === 'classRoomName') {
      handleFilterUpdate({ classRoomName: value || undefined }, false, id, type)
      return
    }
    if (fieldId === 'courseType') {
      handleFilterUpdate({ courseType: value || undefined }, false, id, type)
    }
  },
  'update:classTeacherFilter': (val, isClearAll, id, type) => {
    handleFilterUpdate({ teacherId: val || undefined }, isClearAll, id, type)
  },
  'update:stuPhoneSearchFilter': (val, isClearAll, id, type) => {
    handleFilterUpdate({ classIds: val ? [String(val)] : undefined }, isClearAll, id, type)
  },
  'update:createUserFilter': (val, isClearAll, id, type) => {
    handleFilterUpdate({ defaultTeacherId: val || undefined }, isClearAll, id, type)
  },
  'update:salesPersonFilter': (val, isClearAll, id, type) => {
    handleFilterUpdate({ createdStaffIds: val ? [String(val)] : undefined }, isClearAll, id, type)
  },
  'update:doYouScheduleFilter': (val, isClearAll, id, type) => {
    if (val === 1) {
      handleFilterUpdate({ isScheduled: true }, isClearAll, id, type)
      return
    }
    if (val === 2) {
      handleFilterUpdate({ isScheduled: false }, isClearAll, id, type)
      return
    }
    handleFilterUpdate({ isScheduled: undefined }, isClearAll, id, type)
  },
  'update:openClassStatusFilter': (val, isClearAll, id, type) => {
    handleFilterUpdate({ statues: val ? [val] : undefined }, isClearAll, id, type)
  },
  'update:createTimeFilter': (val, isClearAll, id, type) => {
    handleFilterUpdate({ createdTime: Array.isArray(val) && val.length ? val : undefined }, isClearAll, id, type)
  },
  'update:classEndingTimeFilter': (val, isClearAll, id, type) => {
    handleFilterUpdate({ closedTime: Array.isArray(val) && val.length ? val : undefined }, isClearAll, id, type)
  },
}))

async function handleClassSearch(searchParams) {
  try {
    const res = await pageGroupClassesApi({
      pageRequestModel: searchParams.pageRequestModel || {
        needTotal: true,
        pageSize: 10,
        pageIndex: 1,
        skipCount: 0,
      },
      queryModel: {
        className: searchParams.searchKey || undefined,
      },
    })
    if (res.code !== 200) {
      messageService.error(res.message || '搜索班级失败')
      return
    }
    const list = Array.isArray(res.result?.list) ? res.result.list : []
    allFilterRef.value?.updateStaffSearchData?.({
      result: list.map(item => ({
        id: item.id,
        name: item.name,
      })),
      total: Number(res.result?.total || 0),
    })
  }
  catch (error) {
    console.error('search class failed', error)
    messageService.error('搜索班级失败')
  }
}

function buildQueryModel(source = {}) {
  const queryModel = {}

  if (source.className) {
    queryModel.className = String(source.className).trim()
  }
  if (Array.isArray(source.classIds) && source.classIds.length > 0) {
    queryModel.classIds = source.classIds
  }
  if (source.teacherId) {
    queryModel.teacherId = String(source.teacherId)
  }
  if (source.defaultTeacherId) {
    queryModel.defaultTeacherId = String(source.defaultTeacherId)
  }
  if (source.classRoomName) {
    queryModel.classRoomName = String(source.classRoomName).trim()
  }
  if (Array.isArray(source.statues) && source.statues.length > 0) {
    queryModel.statues = source.statues
  }
  if (Array.isArray(source.createdStaffIds) && source.createdStaffIds.length > 0) {
    queryModel.createdStaffIds = source.createdStaffIds
  }
  if (typeof source.isScheduled === 'boolean') {
    queryModel.isScheduled = source.isScheduled
  }
  if (source.courseType === 'single') {
    queryModel.isMultiProduct = false
  }
  else if (source.courseType === 'compose') {
    queryModel.isMultiProduct = true
  }
  if (source.lessonKey) {
    const raw = String(source.lessonKey)
    const split = raw.split(':')
    queryModel.lessonIds = [split.length > 1 ? split[1] : raw]
  }
  if (Array.isArray(source.createdTime) && source.createdTime.length === 2) {
    queryModel.createdStartTime = source.createdTime[0]
    queryModel.createdEndTime = source.createdTime[1]
  }
  if (Array.isArray(source.closedTime) && source.closedTime.length === 2) {
    queryModel.closedStartDate = source.closedTime[0]
    queryModel.closedEndDate = source.closedTime[1]
  }

  return queryModel
}

async function getClassList(newQueryParams = {}, id, type) {
  listLoading.value = true
  try {
    const queryModel = buildQueryModel(newQueryParams)
    const [listRes, statsRes] = await Promise.all([
      pageGroupClassesApi({
        queryModel,
        pageRequestModel: {
          needTotal: true,
          pageSize: pagination.pageSize,
          pageIndex: pagination.current,
          skipCount: 0,
        },
      }),
      groupClassStatisticsApi(queryModel),
    ])

    if (listRes.code === 200 && listRes.result) {
      dataSource.value = Array.isArray(listRes.result.list) ? listRes.result.list : []
      pagination.total = Number(listRes.result.total || 0)
      mergeCustomFilterOptionsFromClassList(dataSource.value)
      allFilterRef.value?.clearQuickFilter?.(id, type)
    }
    else {
      dataSource.value = []
      pagination.total = 0
      messageService.error(listRes.message || '获取班级列表失败')
    }

    if (statsRes.code === 200 && statsRes.result) {
      stats.value = {
        classCount: Number(statsRes.result.classCount || 0),
        openClassCount: Number(statsRes.result.openClassCount || 0),
        studentCount: Number(statsRes.result.studentCount || 0),
        studentPersonTime: Number(statsRes.result.studentPersonTime || 0),
      }
    }
    else {
      stats.value = {
        classCount: 0,
        openClassCount: 0,
        studentCount: 0,
        studentPersonTime: 0,
      }
    }
  }
  catch (error) {
    console.error('get class list failed', error)
    dataSource.value = []
    pagination.total = 0
    stats.value = {
      classCount: 0,
      openClassCount: 0,
      studentCount: 0,
      studentPersonTime: 0,
    }
    messageService.error('获取班级列表失败')
  }
  finally {
    listLoading.value = false
  }
}

function onTableChange(pageInfo) {
  pagination.current = pageInfo.current
  pagination.pageSize = pageInfo.pageSize
  getClassList(queryState.value)
}

function formatDt(value) {
  if (value == null || value === '')
    return '-'
  const date = dayjs(value)
  return date.isValid() ? date.format('YYYY-MM-DD HH:mm') : '-'
}

function formatClosed(value) {
  if (value == null || value === '')
    return '-'
  const date = dayjs(value)
  if (!date.isValid() || date.year() < 1900)
    return '-'
  return date.format('YYYY-MM-DD')
}

function formatClassTime(times) {
  if (!Array.isArray(times) || times.length === 0)
    return '-'

  return times.map((item) => {
    const startTime = item?.startTime ? dayjs(item.startTime).format('MM-DD HH:mm') : ''
    const endTime = item?.endTime ? dayjs(item.endTime).format('HH:mm') : ''
    if (startTime && endTime)
      return `${startTime}-${endTime}`
    return item?.name || '日程'
  }).join('；')
}

function statusLabel(status) {
  if (status === 1)
    return '开班中'
  if (status === 2)
    return '已结班'
  return `状态${status}`
}

function teacherNames(teachers) {
  if (!Array.isArray(teachers) || teachers.length === 0)
    return '-'
  return teachers.map(item => item.name).filter(Boolean).join('、')
}

function createClass() {
  editClassRecord.value = null
  createClassModal.value = true
}

function openClassListDrawer() {
  classListDrawerFlag.value = true
}

function openAddStudentModal(record) {
  addStudentModalTitle.value = String(record?.name || '').trim() || '班级'
  addStudentModalLessonName.value = String(record?.lessonName || '').trim()
  addStudentModalClassId.value = String(record?.id ?? '').trim()
  addStudentModalLessonId.value = String(record?.lessonId ?? '').trim()
  addStudentModalOpen.value = true
}

function onClassRowMenuClick({ key }, record) {
  if (key === '3') {
    editClassRecord.value = record
    createClassModal.value = true
    return
  }
  console.log(key, record)
}

function afterClassModalSave() {
  getClassList(queryState.value)
}

watch(createClassModal, (open) => {
  if (!open)
    editClassRecord.value = null
})

const allColumns = ref([
  {
    title: '班级名称',
    dataIndex: 'name',
    key: 'name',
    fixed: 'left',
    width: 200,
    required: true,
  },
  {
    title: '关联课程',
    key: 'linkCourse',
    dataIndex: 'linkCourse',
    width: 160,
  },
  {
    title: '学员数',
    key: 'studentNum',
    dataIndex: 'studentNum',
    width: 110,
  },
  {
    title: '班主任',
    key: 'headTeacher',
    dataIndex: 'headTeacher',
    width: 140,
  },
  {
    title: '默认上课教师',
    key: 'defaultTeacher',
    dataIndex: 'defaultTeacher',
    width: 140,
  },
  {
    title: '上课教室',
    dataIndex: 'classRoom',
    key: 'classRoom',
    width: 140,
  },
  {
    title: '上课时间',
    dataIndex: 'classTime',
    key: 'classTime',
    width: 220,
  },
  {
    title: '是否排课',
    dataIndex: 'doYouSchedule',
    key: 'doYouSchedule',
    width: 120,
  },
  {
    title: '已上/日程总数',
    dataIndex: 'alreadyOnOrtotal',
    key: 'alreadyOnOrtotal',
    width: 150,
  },
  {
    title: '状态',
    dataIndex: 'openClassStatus',
    key: 'openClassStatus',
    width: 120,
  },
  {
    title: '创建时间',
    dataIndex: 'createTime',
    key: 'createTime',
    width: 180,
  },
  {
    title: '创建人',
    key: 'createUser',
    dataIndex: 'createUser',
    width: 120,
  },
  {
    title: '备注',
    key: 'remark',
    dataIndex: 'remark',
    width: 160,
  },
  {
    title: '结班日期',
    key: 'classEndingTime',
    dataIndex: 'classEndingTime',
    width: 120,
  },
  {
    title: '操作',
    dataIndex: 'action',
    key: 'action',
    fixed: 'right',
    width: 220,
  },
])

const { selectedValues, columnOptions, filteredColumns, totalWidth }
  = useTableColumns({
    storageKey: 'class-list',
    allColumns,
    excludeKeys: ['action'],
  })

const rowSelection = computed(() => ({
  selectedRowKeys: selectedRowKeys.value,
  onChange: (keys, rows) => {
    selectedRowKeys.value = keys
    selectedRows.value = rows
  },
}))

onMounted(async () => {
  await getClassList(queryState.value)
})
</script>

<template>
  <div>
    <div class="filter-wrap bg-white pl-3 pr-3 rounded-4">
      <all-filter
        ref="allFilterRef"
        :display-array="displayArray"
        :is-quick-show="false"
        :is-show-clsss-or-course-search="true"
        :custom-is-display-list="customSearchFilters"
        search-label="班级名称"
        search-placeholder="请选择班级名称"
        create-user-label="默认上课教师"
        create-user-placeholder="请输入默认上课教师"
        sales-person-label="创建人"
        sales-person-placeholder="请输入创建人"
        :default-open-class-status="defaultOpenClassStatus"
        @staff-search="handleClassSearch"
        v-on="filterUpdateHandlers"
      />
    </div>

    <div class="student-list mt-2 pt-3 pb-3 pl-6 pr-6 bg-white rounded-4">
      <div class="tab-table">
        <div class="table-title flex justify-between">
          <div class="total">
            总计 {{ stats.classCount }} 个班级，{{ stats.openClassCount }} 个开班中，在读学员 {{ stats.studentCount }} 人，在读人次 {{ stats.studentPersonTime }} 人
          </div>
          <div class="edit flex">
            <a-button class="mr-2">
              批量结班
            </a-button>
            <a-dropdown class="mr-2">
              <template #overlay>
                <a-menu>
                  <a-menu-item key="0">
                    导入班级
                  </a-menu-item>
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
            <a-button type="primary" class="mr-2" @click="createClass">
              创建班级
            </a-button>
            <customize-code
              v-model:checked-values="selectedValues"
              :options="columnOptions"
              :total="allColumns.length - 1"
              :num="selectedValues.length - 1"
            />
          </div>
        </div>
        <div class="table-content mt-2">
          <a-table
            :data-source="dataSource"
            :loading="listLoading"
            :pagination="pagination"
            :columns="filteredColumns"
            :row-selection="rowSelection"
            :scroll="{ x: totalWidth }"
            row-key="id"
            size="small"
            @change="onTableChange"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'name'">
                <a-button type="link" class="text-#222" @click="openClassListDrawer">
                  {{ record.name || '-' }}
                </a-button>
              </template>
              <template v-else-if="column.key === 'linkCourse'">
                <div class="text-#222">
                  {{ record.lessonName || '-' }}
                </div>
                <div class="text-3 text-#888 flex flex-items-center">
                  {{ record.isMultiProduct ? '组合课程' : '课程' }}
                </div>
              </template>
              <template v-else-if="column.key === 'studentNum'">
                {{ record.studentCount ?? 0 }}
              </template>
              <template v-else-if="column.key === 'headTeacher'">
                {{ teacherNames(record.teachers) }}
              </template>
              <template v-else-if="column.key === 'defaultTeacher'">
                {{ record.defaultTeacherName || '-' }}
              </template>
              <template v-else-if="column.key === 'classRoom'">
                {{ record.classRoomName || '-' }}
              </template>
              <template v-else-if="column.key === 'classTime'">
                {{ formatClassTime(record.classLessonTimes) }}
              </template>
              <template v-else-if="column.key === 'doYouSchedule'">
                <div class="studentStatus">
                  <span class="dot" />
                  <span>{{ record.isScheduled ? '已排课' : '未排课' }}</span>
                </div>
              </template>
              <template v-else-if="column.key === 'alreadyOnOrtotal'">
                {{ record.classLessonDayInfos?.completeLessonDayCount ?? 0 }}/{{ record.classLessonDayInfos?.lessonDayCount ?? 0 }}
              </template>
              <template v-else-if="column.key === 'openClassStatus'">
                <div
                  class="rounded-2.5 inline-block text-3 pt-0.5 pb-0.5 pl-2 pr-2"
                  :class="record.status === 1 ? 'text-#06f bg-#e6f0ff' : 'text-#666 bg-#f5f5f5'"
                >
                  {{ statusLabel(record.status) }}
                </div>
              </template>
              <template v-else-if="column.key === 'createTime'">
                {{ formatDt(record.createdTime) }}
              </template>
              <template v-else-if="column.key === 'createUser'">
                {{ record.createdStaffName || '-' }}
              </template>
              <template v-else-if="column.key === 'remark'">
                {{ record.remark || '-' }}
              </template>
              <template v-else-if="column.key === 'classEndingTime'">
                {{ formatClosed(record.closedTime) }}
              </template>
              <template v-else-if="column.key === 'action'">
                <span class="flex action">
                  <a class="mr-3">排课</a>
                  <a class="mr-3" @click.prevent="openAddStudentModal(record)">添加学员</a>
                  <div style="cursor: pointer;">
                    <a-dropdown :trigger="['click']" placement="bottom">
                      <a @click.prevent>
                        <div class="intention">
                          更多
                          <CaretDownOutlined
                            class="text-#1677ff"
                            :style="{ fontSize: '12px' }"
                          />
                        </div>
                      </a>
                      <template #overlay>
                        <a-menu style="text-align: center;width: 120px;" @click="(e) => onClassRowMenuClick(e, record)">
                          <a-menu-item key="1">
                            上课点名
                          </a-menu-item>
                          <a-menu-item key="2">
                            未排课点名
                          </a-menu-item>
                          <a-menu-item key="3">
                            编辑班级
                          </a-menu-item>
                          <a-menu-item key="4" danger>
                            结班
                          </a-menu-item>
                        </a-menu>
                      </template>
                    </a-dropdown>
                  </div>
                </span>
              </template>
            </template>
          </a-table>
        </div>
      </div>
    </div>
    <CreateClassModal
      v-model:open="createClassModal"
      :edit-record="editClassRecord"
      @created="afterClassModalSave"
      @updated="afterClassModalSave"
    />
    <ClassListDrawer v-model:open="classListDrawerFlag" />
    <ClassAddStudentModal
      v-model:open="addStudentModalOpen"
      :title="addStudentModalTitle"
      :lesson-name="addStudentModalLessonName"
      :class-id="addStudentModalClassId"
      :lesson-id="addStudentModalLessonId"
      @success="afterClassModalSave"
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
    content: '';
    height: 12px;
    left: 0;
    position: absolute;
    width: 4px;
  }
}

.studentStatus {
  display: flex;
  align-items: center;

  span.dot {
    border-radius: 50%;
    display: inline-block;
    height: 6px;
    position: relative;
    vertical-align: middle;
    width: 6px;
    margin-right: 4px;
    background: var(--pro-ant-color-primary);
  }
}
</style>
