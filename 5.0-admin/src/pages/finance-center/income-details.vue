<script setup>
import { CloseOutlined, DeleteOutlined } from '@ant-design/icons-vue'
import dayjs from 'dayjs'
import { debounce } from 'lodash-es'
import { computed, nextTick, onMounted, onUnmounted, reactive, ref } from 'vue'
import StudentSelect from '@/components/common/student-select.vue'
import { getCourseCategoryPageApi } from '@/api/edu-center/course-list'
import { getCourseIdAndNameApi } from '@/api/edu-center/registr-renewal'
import { getLessonIncomePagedListApi, getLessonIncomeStatisticsApi } from '@/api/finance-center/lesson-income'
import { getUserListApi } from '@/api/internal-manage/staff-manage'
import { useTableColumns } from '@/composables/useTableColumns'
import messageService from '@/utils/messageService'

const TEACHING_METHOD_MAP = {
  1: '班课',
  2: '1v1',
  3: '直播课',
}

const CHARGING_MODE_UNIT_MAP = {
  1: '课时',
  2: '天',
  3: '元',
}

const DETAIL_TYPE_MAP = {
  1: '课时课消',
  2: '手动结课',
  3: '过期结课',
  4: '导入课消',
  5: '课消退还',
  6: '课消补扣',
  7: '按天自动课消',
  8: '课消欠费清算',
  9: '退课手续费',
  10: '撤销结课',
  11: '过期撤回返还',
  12: '作废返还',
  13: '撤销退课手续费',
}

const detailTypeOptions = Object.entries(DETAIL_TYPE_MAP).map(([id, value]) => ({
  id: Number(id),
  value,
}))

const currentMonthStart = dayjs().startOf('month').format('YYYY-MM-DD')
const today = dayjs().format('YYYY-MM-DD')

function createEmptyFilters() {
  return {
    createTime: [],
    studentId: undefined,
    lessonId: undefined,
    productCategoryId: undefined,
    classId: undefined,
    sourceTypes: [],
    staffId: undefined,
    lessonDay: [],
    conformIncomeTime: [currentMonthStart, today],
  }
}

const filterState = reactive(createEmptyFilters())

const loading = ref(false)
const dataSource = ref([])
const stats = ref({
  totalCount: 0,
  totalTuition: 0,
})
const childRefs = ref({})

const pagination = ref({
  current: 1,
  pageSize: 10,
  total: 0,
  showSizeChanger: true,
  showQuickJumper: true,
  showTotal: total => `共 ${total} 条`,
})

const courseOptions = ref([])
const courseFinished = ref(false)
const teacherOptions = ref([])
const teacherFinished = ref(false)
const teacherSearchKey = ref('')
const courseCategoryOptions = ref([])
const classOptions = ref([])

const teacherPagination = ref({
  current: 1,
  pageSize: 20,
  total: 0,
})
const selectedStudentOption = ref(null)

function handleRef(el, key) {
  if (el) {
    childRefs.value[key] = el
  }
}

function mergeOptions(targetRef, source, normalize) {
  const map = new Map((targetRef.value || []).map(item => [String(item.id), item]))
  ;(source || []).forEach((item) => {
    const normalized = normalize(item)
    if (!normalized || normalized.id === undefined || normalized.id === null || normalized.id === '') {
      return
    }
    map.set(String(normalized.id), normalized)
  })
  targetRef.value = Array.from(map.values())
}

function buildQueryModel() {
  return {
    startDate: filterState.createTime?.[0] || undefined,
    endDate: filterState.createTime?.[1] || undefined,
    sourceTypes: filterState.sourceTypes.length ? filterState.sourceTypes : undefined,
    studentId: filterState.studentId || undefined,
    staffId: filterState.staffId || undefined,
    lessonId: filterState.lessonId || undefined,
    lessonDayStartDate: filterState.lessonDay?.[0] || undefined,
    lessonDayEndDate: filterState.lessonDay?.[1] || undefined,
    classId: filterState.classId || undefined,
    productCategoryId: filterState.productCategoryId || undefined,
    conformIncomeTimeStartDate: filterState.conformIncomeTime?.[0] || undefined,
    conformIncomeTimeEndDate: filterState.conformIncomeTime?.[1] || undefined,
  }
}

function formatMoney(value) {
  return Number(value || 0).toFixed(2)
}

function formatDateTime(value, pattern = 'YYYY-MM-DD HH:mm:ss') {
  if (!value) {
    return '-'
  }
  const date = dayjs(value)
  if (!date.isValid() || date.year() <= 1) {
    return '-'
  }
  return date.format(pattern)
}

function minutesToTime(minutes) {
  const hour = String(Math.floor(Number(minutes || 0) / 60)).padStart(2, '0')
  const minute = String(Number(minutes || 0) % 60).padStart(2, '0')
  return `${hour}:${minute}`
}

function hasMeaningfulTimeRange(record) {
  return Number(record.startMinutes || 0) > 0 || Number(record.endMinutes || 0) > 0
}

function formatCourseConsumption(record) {
  return `${Number(record.quantity || 0)}${CHARGING_MODE_UNIT_MAP[record.lessonChargingMode] || ''}`
}

function getDetailTypeText(sourceType) {
  return DETAIL_TYPE_MAP[sourceType] || `类型${sourceType}`
}

function getTeacherText(record) {
  if (Array.isArray(record.teachers) && record.teachers.length) {
    return record.teachers.map(item => item.name).filter(Boolean).join('、')
  }
  return record.teacherName || '-'
}

function getAssistantText(record) {
  if (Array.isArray(record.assistantTeachers) && record.assistantTeachers.length) {
    return record.assistantTeachers.map(item => item.name).filter(Boolean).join('、')
  }
  return record.assistantName || '-'
}

function findOptionLabel(options, id) {
  const item = options.find(option => String(option.id) === String(id))
  return item?.value || item?.nickName || item?.name || String(id || '')
}

const selectedFilterList = computed(() => {
  const list = [
    {
      key: 'conformIncomeTime',
      label: '确认收入时间',
      value: `${filterState.conformIncomeTime[0]}~${filterState.conformIncomeTime[1]}`,
      fixed: true,
    },
  ]

  if (filterState.studentId) {
    list.push({
      key: 'studentId',
      label: '学员/电话',
      value: selectedStudentOption.value?.stuName || String(filterState.studentId),
    })
  }
  if (filterState.lessonId) {
    list.push({
      key: 'lessonId',
      label: '扣费课程账户',
      value: findOptionLabel(courseOptions.value, filterState.lessonId),
    })
  }
  if (filterState.productCategoryId) {
    list.push({
      key: 'productCategoryId',
      label: '课程类别',
      value: findOptionLabel(courseCategoryOptions.value, filterState.productCategoryId),
    })
  }
  if (filterState.classId) {
    list.push({
      key: 'classId',
      label: '课消所属班级',
      value: findOptionLabel(classOptions.value, filterState.classId),
    })
  }
  if (filterState.sourceTypes.length) {
    list.push({
      key: 'sourceTypes',
      label: '明细类型',
      value: filterState.sourceTypes.map(item => DETAIL_TYPE_MAP[item]).filter(Boolean).join('、'),
    })
  }
  if (filterState.lessonDay.length === 2) {
    list.push({
      key: 'lessonDay',
      label: '上课时间',
      value: `${filterState.lessonDay[0]} ~ ${filterState.lessonDay[1]}`,
    })
  }
  if (filterState.createTime.length === 2) {
    list.push({
      key: 'createTime',
      label: '确认收入创建时间',
      value: `${filterState.createTime[0]} ~ ${filterState.createTime[1]}`,
    })
  }
  if (filterState.staffId) {
    list.push({
      key: 'staffId',
      label: '上课教师',
      value: findOptionLabel(teacherOptions.value, filterState.staffId),
    })
  }
  return list
})

const hasSelectedFilters = computed(() => selectedFilterList.value.length > 0)

async function fetchLessonIncomeList() {
  try {
    loading.value = true
    const { result } = await getLessonIncomePagedListApi({
      queryModel: buildQueryModel(),
      sortModel: {
        orderByCreatedTime: 0,
      },
      pageRequestModel: {
        needTotal: true,
        pageSize: pagination.value.pageSize,
        pageIndex: pagination.value.current,
        skipCount: (pagination.value.current - 1) * pagination.value.pageSize,
      },
    })
    const list = Array.isArray(result?.list) ? result.list : []
    dataSource.value = list
    pagination.value.total = result?.total || 0
    mergeOptions(
      classOptions,
      list.filter(item => item.classId && item.className),
      item => ({
        id: item.classId,
        value: item.className,
      }),
    )
  }
  catch (error) {
    console.error('获取确认收入列表失败:', error)
    messageService.error('获取确认收入列表失败')
  }
  finally {
    loading.value = false
  }
}

async function fetchLessonIncomeStatistics() {
  try {
    const { result } = await getLessonIncomeStatisticsApi(buildQueryModel())
    stats.value = {
      totalCount: result?.totalCount || 0,
      totalTuition: result?.totalTuition || 0,
    }
  }
  catch (error) {
    console.error('获取确认收入统计失败:', error)
  }
}

async function refreshPageData() {
  await Promise.all([
    fetchLessonIncomeList(),
    fetchLessonIncomeStatistics(),
  ])
}

const debouncedRefresh = debounce(() => {
  pagination.value.current = 1
  refreshPageData()
}, 250)

async function loadCourseOptions(searchKey = '') {
  try {
    const res = await getCourseIdAndNameApi({
      pageRequestModel: {
        needTotal: true,
        pageSize: 50,
        pageIndex: 1,
        skipCount: 0,
      },
      queryModel: {
        delFlag: false,
        productType: 1,
        searchKey,
        saleStatus: true,
      },
      sortModel: {},
    })
    if (res.code === 200) {
      courseOptions.value = (res.result || []).map(item => ({
        id: item.id,
        value: item.name,
      }))
      courseFinished.value = true
    }
  }
  catch (error) {
    console.error('加载课程列表失败:', error)
  }
}

async function loadCourseCategoryOptions(searchKey = '') {
  try {
    const res = await getCourseCategoryPageApi({
      pageRequestModel: {
        needTotal: true,
        pageSize: 50,
        pageIndex: 1,
        skipCount: 0,
      },
      queryModel: {
        searchKey,
      },
      sortModel: {
        orderBySortNumber: 2,
      },
    })
    if (res.code === 200) {
      courseCategoryOptions.value = (res.result || []).map(item => ({
        id: item.id,
        value: item.name,
      }))
    }
  }
  catch (error) {
    console.error('加载课程类别失败:', error)
  }
}

async function loadTeacherOptions(searchKey = '') {
  try {
    teacherSearchKey.value = searchKey
    childRefs.value.teacher?.openSpinning?.()
    const res = await getUserListApi({
      pageRequestModel: {
        needTotal: true,
        pageSize: teacherPagination.value.pageSize,
        pageIndex: teacherPagination.value.current,
        skipCount: 0,
      },
      queryModel: {
        searchKey,
      },
      sortModel: {},
    })
    if (res.code === 200) {
      const resultData = (res.result || []).map(item => ({
        ...item,
        value: item.value || item.nickName || item.name || '-',
        nickName: item.nickName || item.value || item.name || '-',
        mobile: item.mobile || item.phone || '',
      }))
      if (teacherPagination.value.current === 1) {
        teacherOptions.value = resultData
      }
      else {
        teacherOptions.value = [...teacherOptions.value, ...resultData]
      }
      teacherPagination.value.total = res.total || 0
      teacherFinished.value = teacherOptions.value.length >= teacherPagination.value.total
    }
  }
  catch (error) {
    console.error('加载上课教师数据失败:', error)
    if (teacherPagination.value.current > 1) {
      teacherPagination.value.current -= 1
    }
  }
  finally {
    childRefs.value.teacher?.resetSpinning?.()
  }
}

function onTeacherDropdownVisibleChange(searchKey = '') {
  teacherPagination.value.current = 1
  teacherFinished.value = false
  teacherOptions.value = []
  teacherSearchKey.value = String(searchKey || '')
  loadTeacherOptions(teacherSearchKey.value)
}

function onTeacherSearch(value) {
  teacherPagination.value.current = 1
  teacherFinished.value = false
  teacherOptions.value = []
  teacherSearchKey.value = value
  loadTeacherOptions(value)
}

function loadMoreTeacher() {
  if (teacherFinished.value) {
    return
  }
  if (teacherPagination.value.current * teacherPagination.value.pageSize >= teacherPagination.value.total) {
    return
  }
  teacherPagination.value.current += 1
  loadTeacherOptions(teacherSearchKey.value)
}

function handleStudentChange(value) {
  if (!value) {
    selectedStudentOption.value = null
  }
  nextTick(() => {
    debouncedRefresh()
  })
}

function handleStudentSelect(student) {
  selectedStudentOption.value = student || null
}

function clearFilter(key) {
  switch (key) {
    case 'lessonDay':
    case 'createTime':
      filterState[key] = []
      break
    case 'sourceTypes':
      filterState.sourceTypes = []
      break
    case 'studentId':
      filterState.studentId = undefined
      selectedStudentOption.value = null
      break
    default:
      filterState[key] = undefined
  }
  debouncedRefresh()
}

function clearSelectableFilters() {
  const fixedConformIncomeTime = [...filterState.conformIncomeTime]
  Object.assign(filterState, createEmptyFilters())
  filterState.conformIncomeTime = fixedConformIncomeTime
  selectedStudentOption.value = null
  debouncedRefresh()
}

function handleTableChange(pag) {
  pagination.value.current = pag.current
  pagination.value.pageSize = pag.pageSize
  fetchLessonIncomeList()
}

function handleExport() {
  messageService.info('导出逻辑待接入')
}

const allColumns = ref([
  { title: '确认收入创建时间', dataIndex: 'createdTime', key: 'createdTime', fixed: 'left', width: 180, required: true },
  { title: '学员', dataIndex: 'studentName', key: 'studentName', fixed: 'left', width: 180, required: true },
  { title: '上课课程', dataIndex: 'teachingCourseName', key: 'teachingCourseName', width: 160 },
  { title: '扣费课程账户', dataIndex: 'lessonName', key: 'lessonName', width: 180 },
  { title: '课程类别', dataIndex: 'productCategoryName', key: 'productCategoryName', width: 140 },
  { title: '授课方式', dataIndex: 'teachingMethod', key: 'teachingMethod', width: 120 },
  { title: '明细类型', dataIndex: 'sourceType', key: 'sourceType', width: 140 },
  { title: '上课教师', dataIndex: 'teachers', key: 'teachers', width: 140 },
  { title: '上课助教', dataIndex: 'assistantTeachers', key: 'assistantTeachers', width: 140 },
  { title: '课消所属班级', dataIndex: 'className', key: 'className', width: 150 },
  { title: '上课时间', dataIndex: 'lessonTime', key: 'lessonTime', width: 200 },
  { title: '点名时间', dataIndex: 'rollCallTime', key: 'rollCallTime', width: 180 },
  { title: '课程消耗', dataIndex: 'quantity', key: 'quantity', fixed: 'right', width: 120 },
  { title: '确认收入（元）', dataIndex: 'tuition', key: 'tuition', fixed: 'right', width: 140, required: true },
])

const { selectedValues, columnOptions, filteredColumns, totalWidth } = useTableColumns({
  storageKey: 'income-details-v2',
  allColumns,
  excludeKeys: [],
})

onMounted(async () => {
  await Promise.all([
    loadCourseOptions(),
    loadCourseCategoryOptions(),
    refreshPageData(),
  ])
})

onUnmounted(() => {
  debouncedRefresh.cancel()
})
</script>

<template>
  <div>
    <div class="filter-wrap bg-white pl-3 pr-3 pb-3 rounded-4">
      <div class="filter-section pt-3">
        <span class="section-title mt-0.5 text-#222">筛选条件：</span>
        <div class="filter-toolbar">
          <div class="standard-filters">
            <checkbox-filter
              v-model:checked-values="filterState.conformIncomeTime"
              label="确认收入时间"
              type="dateTime"
              @date-picker-change="debouncedRefresh"
            />
            <checkbox-filter
              v-model:checked-values="filterState.lessonId"
              :options="courseOptions"
              label="扣费课程账户"
              type="radio"
              category="course"
              placeholder="请输入扣费课程账户"
              :finished="courseFinished"
              @radio-change="debouncedRefresh"
              @on-search="loadCourseOptions"
              @on-dropdown-visible-change="loadCourseOptions"
            />
            <checkbox-filter
              v-model:checked-values="filterState.productCategoryId"
              :options="courseCategoryOptions"
              label="课程类别"
              type="radio"
              category="course"
              placeholder="请输入课程类别"
              @radio-change="debouncedRefresh"
              @on-search="loadCourseCategoryOptions"
              @on-dropdown-visible-change="loadCourseCategoryOptions"
            />
            <checkbox-filter
              v-model:checked-values="filterState.classId"
              :options="classOptions"
              label="课消所属班级"
              type="radio"
              category="noSearchRadio"
              @radio-change="debouncedRefresh"
            />
            <checkbox-filter
              v-model:checked-values="filterState.sourceTypes"
              :options="detailTypeOptions"
              label="明细类型"
              type="checkbox"
              @change="debouncedRefresh"
            />
            <checkbox-filter
              v-model:checked-values="filterState.lessonDay"
              label="上课时间"
              type="dateTime"
              @date-picker-change="debouncedRefresh"
            />
            <checkbox-filter
              v-model:checked-values="filterState.createTime"
              label="确认收入创建时间"
              type="dateTime"
              @date-picker-change="debouncedRefresh"
            />
            <checkbox-filter
              :ref="(el) => handleRef(el, 'teacher')"
              v-model:checked-values="filterState.staffId"
              :options="teacherOptions"
              label="上课教师"
              type="radio"
              category="teacher"
              placeholder="请输入上课教师"
              :finished="teacherFinished"
              @radio-change="debouncedRefresh"
              @on-dropdown-visible-change="onTeacherDropdownVisibleChange"
              @on-search="onTeacherSearch"
              @load-more="loadMoreTeacher"
            />
          </div>
          <div class="student-filter-wrap">
            <div class="label">
              学员/电话
            </div>
            <StudentSelect
              v-model="filterState.studentId"
              allow-clear
              placeholder="搜索姓名/手机号"
              width="240px"
              :student-status="1"
              @change="handleStudentChange"
              @select="handleStudentSelect"
            />
          </div>
        </div>
      </div>

      <div v-if="hasSelectedFilters" class="selected-conditions">
        <span class="section-title text-#222">已选条件：</span>
        <div class="condition-tags">
          <a-popconfirm title="确定要清空所有条件吗？" @confirm="clearSelectableFilters">
            <a-tag color="red" class="clear-all mb-2">
              清空已选
              <DeleteOutlined class="text-3 ml-4px mt-0.6px" />
            </a-tag>
          </a-popconfirm>
          <a-tag color="blue" class="condition-tag mb-2">
            <div class="tag-content">
              <span class="condition-label">确认收入时间：</span>
              <div class="condition-values">
                <span class="value-item">{{ filterState.conformIncomeTime[0] }}~{{ filterState.conformIncomeTime[1] }}</span>
              </div>
            </div>
          </a-tag>
          <a-tag v-for="item in selectedFilterList.filter(tag => !tag.fixed)" :key="item.key" color="blue" class="condition-tag mb-2">
            <div class="tag-content">
              <span class="condition-label">{{ item.label }}：</span>
              <div class="condition-values">
                <span class="value-item">
                  {{ item.value }}
                  <CloseOutlined class="close-icon" @click.stop="clearFilter(item.key)" />
                </span>
              </div>
            </div>
          </a-tag>
        </div>
      </div>
    </div>

    <div class="student-list mt-3 pt-3 pb-3 pl-6 pr-6 bg-white rounded-4">
      <div class="tab-table">
        <div class="table-title flex justify-between">
          <div class="total">
            共 {{ stats.totalCount }} 条确认收入记录，确认收入金额总计：￥{{ formatMoney(stats.totalTuition) }}
          </div>
          <div class="edit flex">
            <a-button class="mr-2" @click="handleExport">
              导出数据
            </a-button>
            <customize-code
              v-model:checked-values="selectedValues"
              :options="columnOptions"
              :total="allColumns.length"
              :num="selectedValues.length"
            />
          </div>
        </div>
        <div class="table-content mt-2">
          <a-table
            :data-source="dataSource"
            :pagination="pagination"
            :columns="filteredColumns"
            :scroll="{ x: totalWidth }"
            :loading="loading"
            row-key="id"
            size="small"
            @change="handleTableChange"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'createdTime'">
                {{ formatDateTime(record.createdTime) }}
              </template>
              <template v-if="column.key === 'studentName'">
                <student-avatar
                  :id="record.studentId"
                  :name="record.studentName || '-'"
                  :phone="record.studentPhone || ''"
                  :avatar-url="record.studentAvatar"
                  :show-gender="false"
                  :show-age="false"
                  default-active-key="0"
                />
              </template>
              <template v-if="column.key === 'teachingCourseName'">
                {{ record.teachingCourseName || '-' }}
              </template>
              <template v-if="column.key === 'lessonName'">
                {{ record.lessonName || '-' }}
              </template>
              <template v-if="column.key === 'productCategoryName'">
                {{ record.productCategoryName || '-' }}
              </template>
              <template v-if="column.key === 'teachingMethod'">
                {{ TEACHING_METHOD_MAP[record.teachingMethod] || '-' }}
              </template>
              <template v-if="column.key === 'sourceType'">
                <div class="detail-type-cell">
                  <span class="dot" />
                  <span>{{ getDetailTypeText(record.sourceType) }}</span>
                </div>
              </template>
              <template v-if="column.key === 'teachers'">
                {{ getTeacherText(record) }}
              </template>
              <template v-if="column.key === 'assistantTeachers'">
                {{ getAssistantText(record) }}
              </template>
              <template v-if="column.key === 'className'">
                {{ record.className || '-' }}
              </template>
              <template v-if="column.key === 'lessonTime'">
                <div class="text-#222">
                  {{ formatDateTime(record.lessonDay, 'YYYY-MM-DD') }}
                </div>
                <div v-if="hasMeaningfulTimeRange(record)" class="text-3 text-#888">
                  时段：{{ minutesToTime(record.startMinutes) }}~{{ minutesToTime(record.endMinutes) }}
                </div>
              </template>
              <template v-if="column.key === 'rollCallTime'">
                {{ formatDateTime(record.rollCallTime) }}
              </template>
              <template v-if="column.key === 'quantity'">
                {{ formatCourseConsumption(record) }}
              </template>
              <template v-if="column.key === 'tuition'">
                {{ formatMoney(record.tuition) }}
              </template>
            </template>
          </a-table>
        </div>
      </div>
    </div>
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

.filter-section {
  display: flex;
  align-items: flex-start;
}

.section-title {
  white-space: nowrap;
}

.filter-toolbar {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  width: 100%;
}

.standard-filters {
  display: flex;
  flex-wrap: wrap;
}

.student-filter-wrap {
  align-items: center;
  display: flex;
  margin-left: auto;
  padding-right: 8px;
}

.student-filter-wrap .label {
  border: 1px solid #f0f0f0;
  border-right: 0;
  border-radius: 8px 0 0 8px !important;
  color: #222;
  font-size: 14px;
  height: 32px;
  line-height: 32px;
  min-width: 104px;
  padding: 0 16px 0 8px;
  text-align: center;
  white-space: nowrap;
}

:deep(.student-filter-wrap .ant-select-selector) {
  border-radius: 0 8px 8px 0 !important;
}

.avatar-fallback {
  width: 40px;
  height: 40px;
  border-radius: 999px;
  align-items: center;
  background: #e6f4ff;
  color: #1677ff;
  display: flex;
  font-size: 14px;
  font-weight: 600;
  justify-content: center;
}

.selected-conditions {
  display: flex;
  align-items: flex-start;
}

.condition-tags {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
}

.condition-tag {
  display: flex;
  align-items: center;
  border-radius: 4px;
}

.clear-all {
  cursor: pointer;
  display: flex;
  align-items: center;
}

.tag-content {
  display: flex;
  align-items: center;
}

.condition-values {
  display: flex;
  align-items: center;
}

.value-item {
  display: inline-flex;
  align-items: center;
}

.close-icon {
  margin-left: 6px;
  font-size: 12px;
  cursor: pointer;
  color: rgba(92, 92, 92, 0.45);
  transition: color 0.3s;
}

.close-icon:hover {
  color: rgba(0, 0, 0, 0.75);
}

.detail-type-cell {
  display: flex;
  align-items: center;
}

span.dot {
  border-radius: 50%;
  display: inline-block;
  height: 6px;
  margin-right: 4px;
  width: 6px;
  background: #06f;
}
</style>
