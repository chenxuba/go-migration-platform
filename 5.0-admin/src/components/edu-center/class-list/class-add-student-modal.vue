<script setup>
import { CloseOutlined, FilterFilled, SearchOutlined } from '@ant-design/icons-vue'
import { debounce } from 'lodash-es'
import dayjs from 'dayjs'
import { batchAssignGroupClassStudentsApi, listGroupClassStudentsByClassIdsApi } from '@/api/edu-center/group-class'
import { pageTuitionAccountsByLessonIdApi } from '@/api/edu-center/tuition-account'
import messageService from '@/utils/messageService'

const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
  /** 弹窗标题后缀：班级名称，如「感统集体课2班2」 */
  title: {
    type: String,
    default: '',
  },
  /** 顶部提示条中的课程/报读名称，默认同 title */
  lessonName: {
    type: String,
    default: '',
  },
  /** 当前集体班 id，与 lessonId 同时有值时拉接口 */
  classId: {
    type: String,
    default: '',
  },
  /** 列表行上的 lessonId（单课 id 或组合课 id） */
  lessonId: {
    type: String,
    default: '',
  },
})

const emit = defineEmits(['update:open', 'success'])

const openModal = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

const bannerName = computed(() => {
  const n = (props.lessonName || props.title || '').trim()
  return n || '课程'
})

const userName = ref('')
const listLoading = ref(false)
const submitLoading = ref(false)

/** 本班学员 id，与接口返回的 assignedClass 一起用于禁用勾选 */
const inClassStudentIds = ref(new Set())

const paginationState = reactive({
  current: 1,
  pageSize: 20,
  total: 0,
})

/** 性别表头筛选，与后端 stu_sex 一致：1 男 2 女 0 未知 */
const sexFilter = ref([])

/** 已生效的年龄区间（岁），服务端 TIMESTAMPDIFF YEAR */
const appliedAgeMin = ref(null)
const appliedAgeMax = ref(null)

/** 避免打开弹窗时清空姓名触发 debounce 与首屏 load 重复请求 */
const allowNameTriggeredFetch = ref(false)

const genderFilteredValue = computed(() =>
  sexFilter.value.length ? sexFilter.value.map(String) : null,
)

const tablePagination = computed(() => {
  const cid = String(props.classId || '').trim()
  const lid = String(props.lessonId || '').trim()
  if (!cid || !lid)
    return false
  return {
    current: paginationState.current,
    pageSize: paginationState.pageSize,
    total: paginationState.total,
    showSizeChanger: true,
    showTotal: t => `共 ${t} 条`,
    pageSizeOptions: ['10', '20', '50', '100'],
  }
})

/** 无班级/课程上下文时的演示数据 */
function staticRows() {
  return [
    {
      id: 'static-1',
      displayName: '陈瑞',
      phone: '176****1111',
      avatarUrl: '',
      courseAccount: '时段课程',
      age: '-',
      ageYears: null,
      gender: '0',
      remainingQuantity: '8天',
      classStatus: '未分班',
      classStatusType: 2,
      disabled: false,
    },
  ]
}

function sexToGenderKey(sex) {
  if (sex === 1)
    return '1'
  if (sex === 2)
    return '2'
  return '0'
}

function ageLabelFromBirthday(birthday) {
  if (birthday == null || birthday === '')
    return { label: '-', years: null }
  const d = dayjs(birthday)
  if (!d.isValid() || d.year() <= 1)
    return { label: '-', years: null }
  const y = dayjs().diff(d, 'year')
  if (y < 0 || y > 120)
    return { label: '-', years: null }
  return { label: String(y), years: y }
}

function formatRemainQuantity(quantity, lessonChargingMode) {
  const q = Number(quantity)
  if (Number.isNaN(q))
    return '-'
  if (lessonChargingMode === 2)
    return `${Math.floor(q)}天`
  if (lessonChargingMode === 3 || lessonChargingMode === 4)
    return `${q}`
  return `${q}课时`
}

function mapTuitionRows(list, inClass) {
  return list.map((item) => {
    const sid = String(item.studentId || '')
    const inThis = item.assignedClass === true || inClass.has(sid)
    const { label: ageLabel, years: ageYears } = ageLabelFromBirthday(item.birthday)
    return {
      id: `${sid}-${item.tuitionAccountId}`,
      studentId: sid,
      tuitionAccountId: String(item.tuitionAccountId || ''),
      displayName: item.studentName || '-',
      phone: item.phone || '-',
      avatarUrl: item.avatar || '',
      courseAccount: item.productName || '-',
      age: ageLabel,
      ageYears,
      gender: sexToGenderKey(item.sex),
      remainingQuantity: formatRemainQuantity(item.quantity, item.lessonChargingMode),
      lessonChargingMode: item.lessonChargingMode,
      classStatus: inThis ? '已分班' : '未分班',
      classStatusType: inThis ? 1 : 2,
      disabled: inThis,
    }
  })
}

/** 组装 queryModel；姓名走服务端模糊匹配字段 studentName（与竞品一致） */
function buildTuitionQueryModel(cid, lid) {
  const q = {
    lessonId: lid,
    studentIds: [],
    classId: cid,
  }
  if (sexFilter.value.length)
    q.sex = [...sexFilter.value]
  if (appliedAgeMin.value != null)
    q.ageMin = appliedAgeMin.value
  if (appliedAgeMax.value != null)
    q.ageMax = appliedAgeMax.value
  const nm = userName.value.trim()
  if (nm)
    q.studentName = nm
  return q
}

async function fetchTuitionAccountsPage() {
  const cid = String(props.classId || '').trim()
  const lid = String(props.lessonId || '').trim()
  if (!cid || !lid)
    return
  listLoading.value = true
  try {
    const accRes = await pageTuitionAccountsByLessonIdApi({
      pageRequestModel: {
        needTotal: true,
        pageSize: paginationState.pageSize,
        pageIndex: paginationState.current,
        skipCount: (paginationState.current - 1) * paginationState.pageSize,
      },
      queryModel: buildTuitionQueryModel(cid, lid),
    })
    if (accRes.code !== 200) {
      messageService.error(accRes.message || '获取可添加学员失败')
      data.value = []
      paginationState.total = 0
      return
    }
    const list = Array.isArray(accRes.result?.list) ? accRes.result.list : []
    paginationState.total = Number(accRes.result?.total) || 0
    const inClass = inClassStudentIds.value
    data.value = mapTuitionRows(list, inClass)
  }
  catch (e) {
    console.error(e)
    messageService.error('加载学员列表失败')
    data.value = []
    paginationState.total = 0
  }
  finally {
    listLoading.value = false
  }
}

async function loadTableData() {
  const cid = String(props.classId || '').trim()
  const lid = String(props.lessonId || '').trim()
  if (!cid || !lid) {
    const rows = staticRows()
    data.value = [...rows]
    inClassStudentIds.value = new Set()
    paginationState.total = 0
    return
  }
  listLoading.value = true
  try {
    const stuRes = await listGroupClassStudentsByClassIdsApi({ classIds: [cid] })
    if (stuRes.code !== 200) {
      messageService.error(stuRes.message || '获取本班学员失败')
      data.value = []
      inClassStudentIds.value = new Set()
      paginationState.total = 0
      return
    }
    const inClass = new Set()
    const buckets = Array.isArray(stuRes.result) ? stuRes.result : []
    for (const b of buckets) {
      if (String(b.classId) !== cid)
        continue
      for (const s of b.students || [])
        inClass.add(String(s.id))
    }
    inClassStudentIds.value = inClass
    await fetchTuitionAccountsPage()
  }
  catch (e) {
    console.error(e)
    messageService.error('加载学员列表失败')
    data.value = []
    inClassStudentIds.value = new Set()
    paginationState.total = 0
  }
  finally {
    listLoading.value = false
  }
}

const columns = computed(() => [
  {
    title: '学员姓名',
    dataIndex: 'name',
    key: 'name',
    width: 200,
  },
  {
    title: '课程账户',
    dataIndex: 'courseAccount',
    key: 'courseAccount',
    width: 140,
  },
  {
    title: '年龄',
    dataIndex: 'age',
    key: 'age',
    width: 100,
  },
  {
    title: '性别',
    dataIndex: 'gender',
    key: 'gender',
    width: 100,
    filters: [
      { text: '男', value: '1' },
      { text: '女', value: '2' },
      { text: '未知', value: '0' },
    ],
    filteredValue: genderFilteredValue.value,
  },
  {
    title: '剩余数量',
    dataIndex: 'remainingQuantity',
    key: 'remainingQuantity',
    width: 120,
  },
  {
    title: '分班状态',
    dataIndex: 'classStatus',
    key: 'classStatus',
    width: 140,
  },
])

const data = ref([])

const selectedRowKeys = ref([])
const selectedRows = ref([])

const rowSelection = computed(() => ({
  selectedRowKeys: selectedRowKeys.value,
  onChange: (keys, rows) => {
    selectedRowKeys.value = keys
    selectedRows.value = rows
  },
  getCheckboxProps: record => ({ disabled: record.disabled }),
}))

const minAge = ref(null)
const maxAge = ref(null)
const ageDropdownVisible = ref(false)
const isAgeFiltered = ref(false)

function resetAgeStateOnly() {
  minAge.value = null
  maxAge.value = null
  appliedAgeMin.value = null
  appliedAgeMax.value = null
  isAgeFiltered.value = false
}

function resetAge() {
  resetAgeStateOnly()
  paginationState.current = 1
  fetchTuitionAccountsPage()
}

function handleAgeConfirm() {
  if (minAge.value == null && maxAge.value == null) {
    appliedAgeMin.value = null
    appliedAgeMax.value = null
    isAgeFiltered.value = false
    ageDropdownVisible.value = false
    paginationState.current = 1
    fetchTuitionAccountsPage()
    return
  }
  let min = minAge.value
  let max = maxAge.value
  if (min != null && max != null && min > max) {
    [min, max] = [max, min]
    minAge.value = min
    maxAge.value = max
  }
  appliedAgeMin.value = min
  appliedAgeMax.value = max
  isAgeFiltered.value = true
  paginationState.current = 1
  fetchTuitionAccountsPage()
  ageDropdownVisible.value = false
}

function onTableChange(pag, filters) {
  let needFetch = false
  if (pag) {
    if (pag.current !== paginationState.current || pag.pageSize !== paginationState.pageSize) {
      paginationState.current = pag.current
      paginationState.pageSize = pag.pageSize
      needFetch = true
    }
  }
  if (filters && Object.prototype.hasOwnProperty.call(filters, 'gender')) {
    const g = filters.gender
    const next = Array.isArray(g)
      ? g.map(v => parseInt(String(v), 10)).filter(n => !Number.isNaN(n))
      : []
    const same = next.length === sexFilter.value.length && next.every((v, i) => v === sexFilter.value[i])
    if (!same) {
      sexFilter.value = next
      paginationState.current = 1
      needFetch = true
    }
  }
  if (needFetch)
    fetchTuitionAccountsPage()
}

const debouncedRefetchTuition = debounce(() => {
  paginationState.current = 1
  fetchTuitionAccountsPage()
}, 300)

/** 回车立即按 queryModel.studentName 搜索（不等 debounce） */
function onStudentNameSearchEnter() {
  if (!openModal.value)
    return
  const cid = String(props.classId || '').trim()
  const lid = String(props.lessonId || '').trim()
  if (!cid || !lid)
    return
  debouncedRefetchTuition.cancel()
  paginationState.current = 1
  fetchTuitionAccountsPage()
}

watch(userName, () => {
  if (!openModal.value || !allowNameTriggeredFetch.value)
    return
  const cid = String(props.classId || '').trim()
  const lid = String(props.lessonId || '').trim()
  if (!cid || !lid)
    return
  debouncedRefetchTuition()
})

onBeforeUnmount(() => {
  debouncedRefetchTuition.cancel()
})

async function handleSubmit() {
  const cid = String(props.classId || '').trim()
  const lid = String(props.lessonId || '').trim()
  if (!cid || !lid) {
    messageService.warning('缺少班级或课程信息')
    return
  }
  const rows = selectedRows.value.filter(r => r && !r.disabled)
  if (rows.length === 0) {
    messageService.warning('请选择可添加的学员')
    return
  }
  submitLoading.value = true
  try {
    const res = await batchAssignGroupClassStudentsApi({
      classIds: [cid],
      students: rows.map(r => ({
        studentId: String(r.studentId || ''),
        tuitionAccountId: String(r.tuitionAccountId || ''),
      })),
      enforceClassAssign: true,
    })
    const ok = res.code === 200 && (res.result?.success === true || res.data?.success === true)
    if (ok) {
      messageService.success('添加学员成功')
      emit('success')
      closeFun()
      return
    }
    if (res.code !== 500)
      messageService.error(res.message || '添加学员失败')
  }
  catch (e) {
    const msg = e?.response?.data?.message || e?.message || '添加学员失败'
    messageService.error(msg)
  }
  finally {
    submitLoading.value = false
  }
}

function closeFun() {
  openModal.value = false
  allowNameTriggeredFetch.value = false
  selectedRowKeys.value = []
  selectedRows.value = []
  debouncedRefetchTuition.cancel()
  resetAgeStateOnly()
  sexFilter.value = []
  paginationState.current = 1
  paginationState.pageSize = 20
  paginationState.total = 0
}

watch(
  () => props.open,
  async (v) => {
    if (!v)
      return
    allowNameTriggeredFetch.value = false
    debouncedRefetchTuition.cancel()
    userName.value = ''
    selectedRowKeys.value = []
    selectedRows.value = []
    resetAgeStateOnly()
    sexFilter.value = []
    paginationState.current = 1
    paginationState.pageSize = 20
    await loadTableData()
    await nextTick()
    allowNameTriggeredFetch.value = true
  },
)

function isRemainingZero(v) {
  if (v === 0 || v === '0')
    return true
  const s = String(v)
  return s === '0课时' || s === '0天'
}
</script>

<template>
  <a-modal
    v-model:open="openModal"
    centered
    class="add-class-student-modal"
    :keyboard="false"
    :closable="false"
    :mask-closable="false"
    :width="920"
    :footer="null"
    @cancel="closeFun"
  >
    <template #title>
      <div class="text-5 flex justify-between flex-center pr-1">
        <span class="font-500 text-#222">添加学员-{{ title || '班级' }}</span>
        <a-button type="text" class="close-btn" @click="closeFun">
          <template #icon>
            <CloseOutlined class="text-4.5 text-#8c8c8c" />
          </template>
        </a-button>
      </div>
    </template>

    <a-alert
      class="add-student-banner border-0 rounded-0 !mb-0 pl-30px"
      type="info"
      show-icon
    >
      <template #message>
        <span class="text-#1677ff">以下为「{{ bannerName }}」报读下的学员</span>
      </template>
    </a-alert>

    <div class="modal-body-inner">
      <a-input
        v-model:value="userName"
        allow-clear
        placeholder="请输入学员姓名（模糊搜索）"
        class="student-search-input"
        @press-enter="onStudentNameSearchEnter"
      >
        <template #prefix>
          <SearchOutlined class="text-#bfbfbf" />
        </template>
      </a-input>

      <div class="table-wrap mt-3">
        <a-table
          :columns="columns"
          :data-source="data"
          row-key="id"
          :pagination="tablePagination"
          :loading="listLoading"
          :row-selection="rowSelection"
          :scroll="{ x: 840 }"
          size="middle"
          @change="onTableChange"
        >
          <template #headerCell="{ column }">
            <template v-if="column.dataIndex === 'age'">
              <div class="flex items-center justify-between gap-1">
                <span>{{ column.title }}</span>
                <a-dropdown v-model:open="ageDropdownVisible" placement="bottomRight" :trigger="['click']">
                  <FilterFilled
                    class="text-12px cursor-pointer shrink-0"
                    :class="[
                      isAgeFiltered ? 'text-#1677ff' : 'text-#00000040 hover:text-#00000073',
                    ]"
                  />
                  <template #overlay>
                    <a-menu class="!p-0">
                      <div class="age-filter">
                        <div class="flex justify-between mb-2 text-3 text-#666">
                          <span>最小年龄（岁）</span>
                          <span>最大年龄（岁）</span>
                        </div>
                        <div class="flex items-center mb-3 gap-2">
                          <a-input-number
                            v-model:value="minAge"
                            :controls="false"
                            class="flex-1"
                            :min="0"
                            :max="100"
                            :precision="0"
                            placeholder="最小"
                          />
                          <span class="text-#d9d9d9">-</span>
                          <a-input-number
                            v-model:value="maxAge"
                            :controls="false"
                            class="flex-1"
                            :min="0"
                            :max="100"
                            :precision="0"
                            placeholder="最大"
                          />
                        </div>
                        <div class="flex justify-between gap-2">
                          <a-button
                            size="small"
                            type="link"
                            class="!px-1"
                            :disabled="minAge == null && maxAge == null"
                            @click="resetAge"
                          >
                            重置
                          </a-button>
                          <a-button size="small" type="primary" @click="handleAgeConfirm">
                            确定
                          </a-button>
                        </div>
                      </div>
                    </a-menu>
                  </template>
                </a-dropdown>
              </div>
            </template>
          </template>

          <template #bodyCell="{ column, record }">
            <template v-if="column.dataIndex === 'name'">
              <div class="flex items-center gap-3">
                <a-avatar
                  v-if="record.avatarUrl"
                  class="shrink-0"
                  :size="36"
                  :src="record.avatarUrl"
                />
                <a-avatar
                  v-else
                  class="shrink-0 !bg-#1677ff"
                  :size="36"
                >
                  {{ (record.displayName || '?').slice(0, 1) }}
                </a-avatar>
                <div class="min-w-0">
                  <div class="text-#222 truncate">
                    {{ record.displayName }}
                  </div>
                  <div class="text-3 text-#8c8c8c truncate">
                    {{ record.phone || '-' }}
                  </div>
                </div>
              </div>
            </template>
            <template v-else-if="column.dataIndex === 'age'">
              {{ record.age === '-' ? '-' : `${record.age} 岁` }}
            </template>
            <template v-else-if="column.dataIndex === 'gender'">
              {{ record.gender === '1' ? '男' : record.gender === '2' ? '女' : '未知' }}
            </template>
            <template v-else-if="column.dataIndex === 'remainingQuantity'">
              <span
                :class="isRemainingZero(record.remainingQuantity) ? 'text-#ff4d4f' : 'text-#222'"
              >
                {{ record.remainingQuantity }}
              </span>
            </template>
            <template v-else-if="column.dataIndex === 'classStatus'">
              <div class="flex items-center">
                <span
                  class="inline-block h-6px w-6px rounded-full shrink-0"
                  :class="record.classStatusType === 1 ? 'bg-#1677ff' : 'bg-#fa8c16'"
                />
                <span
                  class="text-3.5 px-1"
                  :class="record.classStatusType === 1 ? 'text-#1677ff' : ' text-#fa8c16'"
                >
                  {{ record.classStatus }}
                </span>
              </div>
            </template>
          </template>
        </a-table>
      </div>
    </div>

    <div class="modal-footer-bar flex items-center justify-between border-t border-#f0f0f0 bg-white px-6 py-3">
      <span class="text-#222">已选择: {{ selectedRowKeys.length }} 人</span>
      <a-space :size="12">
        <a-button @click="closeFun">
          取消
        </a-button>
        <a-button
          type="primary"
          :loading="submitLoading"
          :disabled="selectedRowKeys.length === 0"
          @click="handleSubmit"
        >
          确定
        </a-button>
      </a-space>
    </div>
  </a-modal>
</template>

<style lang="less" scoped>
.close-btn {
  &:hover {
    background: transparent;
  }
}

.modal-body-inner {
  padding: 16px 24px 8px;
}

.student-search-input {
  height: 44px;
  border-radius: 10px;

 
}

.table-wrap {
  :deep(.ant-table-thead > tr > th) {
    background: #fafafa;
    color: #595959;
    font-weight: 500;
  }
}

.age-filter {
  padding: 12px;
  min-width: 260px;
}
</style>

<style lang="less">
.add-class-student-modal .ant-modal-content {
  padding: 0;
  overflow: hidden;
}

.add-class-student-modal .ant-modal-header {
  padding: 16px 24px;
  margin-bottom: 0;
  border-bottom: 1px solid #f0f0f0;
}

.add-class-student-modal .ant-modal-body {
  padding: 0;
}

.add-student-banner.ant-alert {
  margin: 0;
  border-radius: 0;
  border: none;
  background: #e6f4ff;
}

.add-student-banner .ant-alert-icon {
  color: #1677ff;
}

/* 选中行无高亮底；hover 仅极浅灰（覆盖 Ant Design 默认选中蓝、hover 蓝） */
.add-class-student-modal .ant-table-tbody > tr > td {
  background: #fff !important;
}

.add-class-student-modal .ant-table-tbody > tr:hover > td {
  background: #f7f7f7 !important;
}

.add-class-student-modal .ant-table-tbody > tr.ant-table-row-selected > td {
  background: #fff !important;
}

.add-class-student-modal .ant-table-tbody > tr.ant-table-row-selected:hover > td {
  background: #f7f7f7 !important;
}
</style>
