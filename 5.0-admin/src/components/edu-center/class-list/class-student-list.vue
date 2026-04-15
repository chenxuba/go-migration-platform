<script setup lang="ts">
import { DownOutlined } from '@ant-design/icons-vue'
import dayjs from 'dayjs'
import { computed, reactive, ref, watch } from 'vue'
import {
  getGroupClassStudentStatisticsApi,
  getGroupClassStudentTeachingRecordCountApi,
  pageGroupClassStudentsApi,
  type GroupClassStudentPagedItem,
  type GroupClassStudentTeachingRecordCountItem,
} from '@/api/edu-center/group-class'
import ClassAddStudentModal from './class-add-student-modal.vue'
import { ParentRelationshipLabel } from '@/enums'
import messageService from '@/utils/messageService'

const props = defineProps({
  drawerOpen: {
    type: Boolean,
    default: false,
  },
  classId: {
    type: String,
    default: '',
  },
  className: {
    type: String,
    default: '',
  },
  lessonId: {
    type: String,
    default: '',
  },
  lessonName: {
    type: String,
    default: '',
  },
})

const checked = ref(false)
const loading = ref(false)
const addStudentVisible = ref(false)
const dataSource = ref<GroupClassStudentPagedItem[]>([])
const stats = ref({
  studentCount: 0,
  noneBindCount: 0,
  noneFaceCount: 0,
})
const recordCountMap = ref<Record<string, GroupClassStudentTeachingRecordCountItem>>({})
const pagination = reactive({
  current: 1,
  pageSize: 20,
  total: 0,
})

let requestSeq = 0

const columns = computed(() => [
  {
    title: '学员/性别',
    dataIndex: 'name',
    key: 'name',
    fixed: 'left' as const,
    width: 170,
  },
  {
    title: '联系电话',
    dataIndex: 'phone',
    key: 'phone',
    fixed: 'left' as const,
    width: 130,
  },
  {
    title: '家校通',
    dataIndex: 'isBind',
    key: 'isBind',
    width: 120,
  },
  {
    title: '人脸采集',
    dataIndex: 'isStudentFace',
    key: 'isStudentFace',
    width: 120,
  },
  {
    title: '积分',
    dataIndex: 'point',
    key: 'point',
    width: 100,
  },
  {
    title: '储值账户余额',
    dataIndex: 'balance',
    key: 'balance',
    width: 130,
  },
  {
    title: '默认课程账户',
    dataIndex: 'classStudentTuitionAccountInfo',
    key: 'classStudentTuitionAccountInfo',
    width: 180,
  },
  {
    title: '报读数量',
    dataIndex: 'totalQuantity',
    key: 'totalQuantity',
    width: 100,
  },
  {
    title: '已用数量',
    dataIndex: 'usedQuantity',
    key: 'usedQuantity',
    width: 100,
  },
  {
    title: '班内已用课时',
    dataIndex: 'usedClassTime',
    key: 'usedClassTime',
    width: 110,
  },
  {
    title: '剩余数量',
    dataIndex: 'quantity',
    key: 'quantity',
    width: 100,
  },
  {
    title: '报读学费',
    dataIndex: 'totalTuition',
    key: 'totalTuition',
    width: 120,
  },
  {
    title: '已用学费',
    dataIndex: 'confirmedTuition',
    key: 'confirmedTuition',
    width: 120,
  },
  {
    title: '剩余学费',
    dataIndex: 'tuition',
    key: 'tuition',
    width: 120,
  },
  {
    title: '有效期至',
    dataIndex: 'expireTime',
    key: 'expireTime',
    width: 120,
  },
  {
    title: '停课日期',
    dataIndex: 'suspendedTime',
    key: 'suspendedTime',
    width: 120,
  },
  {
    title: '结课日期',
    dataIndex: 'classEndingTime',
    key: 'classEndingTime',
    width: 120,
  },
  {
    title: '状态',
    dataIndex: 'status',
    key: 'status',
    width: 100,
  },
  {
    title: '上课次数',
    dataIndex: 'studentAttendCount',
    key: 'studentAttendCount',
    width: 100,
  },
  {
    title: '请假次数',
    dataIndex: 'studentLeaveCount',
    key: 'studentLeaveCount',
    width: 100,
  },
  {
    title: '入班时间',
    dataIndex: 'joinTime',
    key: 'joinTime',
    width: 120,
  },
  {
    title: '操作',
    dataIndex: 'action',
    key: 'action',
    fixed: 'right' as const,
    width: 170,
  },
])

const totalWidth = computed(() =>
  columns.value.reduce((acc, column) => acc + (column.width || 0), 0),
)

const tablePagination = computed(() => ({
  current: pagination.current,
  pageSize: pagination.pageSize,
  total: pagination.total,
  showSizeChanger: true,
  showQuickJumper: true,
  showTotal: (total: number) => `共 ${total} 条`,
}))

function buildQueryModel() {
  return {
    id: props.classId,
    classId: props.classId,
    status: checked.value ? [1, 2, 3] : [1],
    ignoreSuspendedTuitionAccount: false,
  }
}

function formatGender(value?: number) {
  if (value === 1)
    return '男'
  if (value === 0)
    return '女'
  return '未知'
}

function formatDate(value?: string) {
  if (!value || `${value}`.startsWith('0001-01-01'))
    return '-'
  const date = dayjs(value)
  return date.isValid() ? date.format('YYYY-MM-DD') : '-'
}

function formatExpireDate(record: Partial<GroupClassStudentPagedItem>) {
  if (!record.enableExpireTime)
    return '不限'
  return formatDate(record.expireTime)
}

function formatCurrency(value?: number) {
  return `¥ ${Number(value || 0).toFixed(2)}`
}

function getQuantityUnit(lessonChargingMode?: number) {
  if (lessonChargingMode === 2)
    return '天'
  if (lessonChargingMode === 3 || lessonChargingMode === 4)
    return '元'
  return '课时'
}

function formatQuantity(value?: number, lessonChargingMode?: number) {
  return `${formatNumber(value)}${getQuantityUnit(lessonChargingMode)}`
}

function formatNumber(value?: number) {
  const num = Number(value || 0)
  if (!Number.isFinite(num))
    return '-'
  if (Number.isInteger(num))
    return String(num)
  return num.toFixed(2).replace(/\.?0+$/, '')
}

function getUsedQuantity(record: Partial<GroupClassStudentPagedItem>) {
  const total = Number(record.totalQuantity || 0)
  return Math.max(total - getRemainQuantity(record), 0)
}

function getRemainQuantity(record: Partial<GroupClassStudentPagedItem>) {
  return Number(record.quantity || 0) + Number(record.freeQuantity || 0)
}

function getClassStatusInfo(record: Partial<GroupClassStudentPagedItem>) {
  if (Number(record.tuitionAccountStatus || 0) === 3)
    return { text: '结课', className: 'text-#888 bg-#f5f5f5' }
  if (record.status === 3) {
    if (record.classEndingTime && `${record.classEndingTime}` !== '0001-01-01T00:00:00')
      return { text: '结课', className: 'text-#888 bg-#f5f5f5' }
    return { text: '转出', className: 'text-#0c3 bg-#e6ffec' }
  }
  if (record.status === 2) {
    return { text: '停课', className: 'text-#f90 bg-#fff5e6' }
  }
  return { text: '在读', className: 'bg-#e6f0ff text-#06f' }
}

function getPhoneRelationshipText(value?: number) {
  if (!value)
    return ''
  return ParentRelationshipLabel[value as keyof typeof ParentRelationshipLabel] || ''
}

function getRecordCount(studentId: string) {
  return recordCountMap.value[studentId] || {
    studentId,
    studentAttendCount: 0,
    studentLeaveCount: 0,
    studentTruancyCount: 0,
  }
}

async function loadStudentTable() {
  const currentSeq = ++requestSeq
  const classId = String(props.classId || '').trim()
  if (!classId) {
    dataSource.value = []
    pagination.total = 0
    stats.value = {
      studentCount: 0,
      noneBindCount: 0,
      noneFaceCount: 0,
    }
    recordCountMap.value = {}
    return
  }

  loading.value = true
  try {
    const queryModel = buildQueryModel()
    const [statsRes, listRes] = await Promise.all([
      getGroupClassStudentStatisticsApi(queryModel),
      pageGroupClassStudentsApi({
        queryModel,
        pageRequestModel: {
          needTotal: true,
          pageSize: pagination.pageSize,
          pageIndex: pagination.current,
          skipCount: (pagination.current - 1) * pagination.pageSize,
        },
      }),
    ])

    if (currentSeq !== requestSeq)
      return

    if (statsRes.code !== 200) {
      throw new Error(statsRes.message || '加载班级学员统计失败')
    }
    if (listRes.code !== 200) {
      throw new Error(listRes.message || '加载班级学员列表失败')
    }

    stats.value = {
      studentCount: Number(statsRes.result?.studentCount || 0),
      noneBindCount: Number(statsRes.result?.noneBindCount || 0),
      noneFaceCount: Number(statsRes.result?.noneFaceCount || 0),
    }

    const list = Array.isArray(listRes.result?.list) ? listRes.result.list : []
    dataSource.value = list
    pagination.total = Number(listRes.result?.total || 0)

    const studentIds = list.map(item => String(item.id || '')).filter(Boolean)
    if (!studentIds.length) {
      recordCountMap.value = {}
      return
    }

    const countRes = await getGroupClassStudentTeachingRecordCountApi({
      studentIds,
      classId,
      studentTeachingRecordStatuses: [1, 2, 3],
    })

    if (currentSeq !== requestSeq)
      return

    if (countRes.code !== 200) {
      throw new Error(countRes.message || '加载上课次数失败')
    }

    const nextMap: Record<string, GroupClassStudentTeachingRecordCountItem> = {}
    const countList = Array.isArray(countRes.result) ? countRes.result : []
    countList.forEach((item) => {
      if (item?.studentId)
        nextMap[item.studentId] = item
    })
    recordCountMap.value = nextMap
  }
  catch (error: any) {
    console.error('load group class students failed', error)
    dataSource.value = []
    pagination.total = 0
    recordCountMap.value = {}
    messageService.error(error?.message || '加载班级学员列表失败')
  }
  finally {
    if (currentSeq === requestSeq)
      loading.value = false
  }
}

function addStudent() {
  addStudentVisible.value = true
}

function handleBatchAction() {
  messageService.info('批量操作待实现')
}

function handleMoveClass() {
  messageService.info('调至其他班待实现')
}

function handleRemoveClass() {
  messageService.info('移出本班待实现')
}

function handleSwitchAccount() {
  messageService.info('切换课程账户待实现')
}

function getAccountName(record: Partial<GroupClassStudentPagedItem>) {
  return record.classStudentTuitionAccountInfo?.productName || '-'
}

function getQuantityDetailText(record: Partial<GroupClassStudentPagedItem>) {
  const unit = getQuantityUnit(record.classStudentTuitionAccountInfo?.lessonChargingMode)
  const buyQuantity = Math.max(Number(record.totalQuantity || 0) - Number(record.totalFreeQuantity || 0), 0)
  const freeQuantity = Number(record.totalFreeQuantity || 0)
  const segments = [`购${formatNumber(buyQuantity)}${unit}`]
  if (freeQuantity > 0)
    segments.push(`赠${formatNumber(freeQuantity)}${unit}`)
  return segments.join(' ')
}

function handleTableChange(pageInfo: { current?: number, pageSize?: number }) {
  const nextPageSize = Number(pageInfo.pageSize || pagination.pageSize)
  const pageSizeChanged = nextPageSize !== pagination.pageSize
  pagination.pageSize = nextPageSize
  pagination.current = pageSizeChanged ? 1 : Number(pageInfo.current || 1)
  loadStudentTable()
}

function handleAddStudentSuccess() {
  addStudentVisible.value = false
  loadStudentTable()
}

watch(
  () => [props.classId, checked.value, props.drawerOpen],
  ([classId, checkedValue, drawerOpen], prev = []) => {
    const [prevClassId, prevCheckedValue, prevDrawerOpen] = prev
    if (!drawerOpen)
      return
    const openedNow = drawerOpen && !prevDrawerOpen
    const classChanged = classId !== prevClassId
    const checkedChanged = checkedValue !== prevCheckedValue
    if (!openedNow && !classChanged && !checkedChanged)
      return
    pagination.current = 1
    loadStudentTable()
  },
  { immediate: true },
)
</script>

<template>
  <div>
    <div class="m-12px">
      <div class="bg-#fff pt-18px px-20px rounded-10px">
        <div class="flex justify-between items-center pb-12px">
          <custom-title
            :title="`共 ${stats.studentCount} 人，${stats.noneBindCount} 人未关注家校通，${stats.noneFaceCount} 人未人脸采集`"
            font-size="14px"
          />
          <div class="flex items-center">
            <a-checkbox v-model:checked="checked">
              显示转出、停课、结课学员
            </a-checkbox>
            <a-dropdown class="mx-2">
              <template #overlay>
                <a-menu>
                  <a-menu-item key="1" @click="handleBatchAction">
                    批量调至其他班
                  </a-menu-item>
                  <a-menu-item key="2" @click="handleBatchAction">
                    批量移出本班
                  </a-menu-item>
                </a-menu>
              </template>
              <a-button>
                批量操作
                <DownOutlined :style="{ fontSize: '10px' }" />
              </a-button>
            </a-dropdown>
            <a-button type="primary" @click="addStudent">
              添加学员
            </a-button>
          </div>
        </div>

        <a-table
          row-key="id"
          size="small"
          :columns="columns"
          :data-source="dataSource"
          :loading="loading"
          :pagination="pagination.total > pagination.pageSize ? tablePagination : false"
          :scroll="{ x: totalWidth }"
          @change="handleTableChange"
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'name'">
              <student-avatar
                :id="record.id"
                :name="record.name"
                :gender="formatGender(record.sex)"
                :avatar-url="record.avatar"
                :show-age="false"
                default-active-key="0"
              />
            </template>

            <template v-else-if="column.key === 'phone'">
              <div class="text-#222">
                {{ getPhoneRelationshipText(record.phoneRelationship) || '-' }}
              </div>
              <div class="text-#888">
                {{ record.phone || '-' }}
              </div>
            </template>

            <template v-else-if="column.key === 'isBind'">
              <div class="flex flex-items-center">
                <span class="whitespace-nowrap" :class="record.isBind ? 'text-green-600' : 'text-#ccc'">
                  {{ record.isBind ? '已关注' : '未关注' }}
                </span>
                <svg v-if="!record.isBind" width="16px" height="16px" class="ml-2" viewBox="0 0 16 16">
                  <g fill="none" fill-rule="evenodd">
                    <g transform="translate(-474 -608)" fill="#CCC">
                      <g transform="translate(398 580)">
                        <g transform="translate(76 21.6)">
                          <g transform="translate(0 6.4)">
                            <path d="M12.5488957 14.2844713H11.5010486C11.1341596 14.280754 10.8398076 13.9883197 10.843536 13.6312425 10.8398076 13.2741654 11.1341596 12.9817311 11.5010486 12.9780138H12.5488957C13.1929132 12.9707828 13.7094253 12.457622 13.7035882 11.8308133V5.51659915C13.7049584 5.07149643 13.4426881 4.66546656 13.0299588 4.47372986L8.49973266 2.41098807C8.19497588 2.2717625 7.84236314 2.2717625 7.53760636 2.41098807L3.00725203 4.47372986C2.59455747 4.66549483 2.33231941 5.07151473 2.33368895 5.51659915V8.11331051C2.33741739 8.47038769 2.04306536 8.76282195 1.67617635 8.76653928 1.30928733 8.76282195 1.0149353 8.47038769 1.01862871 8.11331051V5.51659915C1.01573797 4.56462047 1.57664141 3.69619985 2.4593492 3.28605311L6.98970297 1.22331132C7.64157303.925562892 8.39577156.925562892 9.04764162 1.22331132L13.5778672 3.28605311C14.460609 3.69617215 15.0215439 4.56460257 15.0186287 5.51659915V11.8308133C15.0186287 13.1837748 13.9107309 14.2844713 12.5488957 14.2844713Z" />
                            <path d="M1.56733162 10.2194036C1.40127346 10.2195233 1.23544109 10.2313173 1.07112909 10.2546935 1.02383678 10.4730961 1 10.6956916 1 10.9188916 1 11.7700045 1.34739282 12.5862583 1.96575882 13.1880863 2.58412481 13.7899143 3.42280952 14.1280178 4.29731162 14.1280178 4.51607755 14.1280178 4.73430678 14.1069519 4.94880175 14.0650857 4.97598308 13.8947261 4.98963678 13.7225811 4.98963678 13.5501797 4.98963678 12.666803 4.6290758 11.8196066 3.98726883 11.1949647 3.34546185 10.5703228 2.4749842 10.2194036 1.56733162 10.2194036Z" />
                            <path d="M4.04965057 14.1112242C4.36580361 14.1624804 4.68574686 14.1883875 5.00625618 14.1886844 6.55014367 14.1886844 8.03079835 13.5917848 9.12249298 12.529291 10.2141876 11.4667972 10.8274979 10.0257453 10.8274979 8.5231503 10.8271446 8.25723104 10.8076999 7.99166078 10.7693057 7.72837983 10.4531526 7.67712408 10.1332094 7.65121719 9.81270009 7.65092023 8.26881275 7.65092023 6.78815822 8.24781981 5.69646369 9.3103135 4.60476916 10.3728072 3.99145897 11.813859 3.99145897 13.3164538 3.99181199 13.582373 4.01125654 13.8479433 4.04965057 14.1112242Z" />
                          </g>
                        </g>
                      </g>
                    </g>
                  </g>
                </svg>
              </div>
            </template>

            <template v-else-if="column.key === 'isStudentFace'">
              <div class="flex flex-items-center">
                <span class="whitespace-nowrap" :class="record.isStudentFace ? 'text-green-600' : 'text-#ccc'">
                  {{ record.isStudentFace ? '已采集' : '未采集' }}
                </span>
                <svg v-if="!record.isStudentFace" width="16px" height="16px" viewBox="0 0 16 16" class="ml-2">
                  <g fill="none" fill-rule="evenodd">
                    <g transform="translate(-594 -608)">
                      <g transform="translate(518 310)">
                        <g transform="translate(0 270)">
                          <g transform="translate(76 21.6)">
                            <g transform="translate(0 6.4)">
                              <polygon fill="#000" fill-rule="nonzero" opacity="0" points="0 0 16 0 16 16 8 16 0 16" />
                              <path fill="#CCC" d="M1.49983336 11C1.74529324 10.9999182 1.94950067 11.1767253 1.99191437 11.4099604L2 11.4998334V14H4.5C4.74545992 14 4.9496084 14.1768752 4.99194436 14.4101244L5 14.5C5 14.7454599 4.82312487 14.9496084 4.58987566 14.9919444L4.5 15H1.50100003C1.25559799 15 1.05147725 14.8232051 1.00908211 14.5900195L1.00100006 14.5001667 1 11.5001667C.999908009 11.2240243 1.223691 11.0000921 1.49983336 11ZM14.4988336 11C14.7442935 10.9999183 14.9485009 11.1767254 14.9909146 11.4099605L14.9990002 11.4998334 15 14.4998334C15.0000818 14.7453511 14.8231944 14.9495863 14.5898958 14.9919408L14.5 15H11.5C11.2238576 15 11 14.7761424 11 14.5 11 14.2545401 11.1768752 14.0503917 11.4101244 14.0080557L11.5 14H14L13.9990003 11.5001667C13.9989185 11.2547068 14.1757256 11.0504994 14.4089607 11.0080857L14.4988336 11ZM4.5 9H11.5L11.4931641 9.38828125 11.4769287 9.60498047 11.4453125 9.83125C11.28125 10.75 10.625 11.8 8 11.8 5.484375 11.8 4.77685547 10.8356771 4.5778656 9.94669189L4.53663635 9.71717529C4.53140259 9.67943522 4.5269165 9.64200846 4.52307129 9.60498047L4.50683594 9.38828125 4.5 9ZM11 5.5C11.5522847 5.5 12 5.94771525 12 6.5 12 7.05228475 11.5522847 7.5 11 7.5 10.4477153 7.5 10 7.05228475 10 6.5 10 5.94771525 10.4477153 5.5 11 5.5ZM5 5.5C5.55228475 5.5 6 5.94771525 6 6.5 6 7.05228475 5.55228475 7.5 5 7.5 4.44771525 7.5 4 7.05228475 4 6.5 4 5.94771525 4.44771525 5.5 5 5.5ZM14.5 1C14.7455177 1 14.9496939 1.17695541 14.9919707 1.41026814L15 1.50016663 14.9990002 4.50016663C14.9989082 4.77630898 14.774976 5.000092 14.4988336 5 14.2533737 4.99991817 14.0492842 4.82297499 14.007026 4.58971169L13.9990003 4.49983337 14 2H11.5C11.2545401 2 11.0503916 1.82312484 11.0080557 1.58987563L11 1.5C11 1.25454011 11.1768752 1.05039163 11.4101244 1.00805567L11.5 1H14.5ZM4.5 1C4.77614235 1 5 1.22385763 5 1.5 5 1.74545989 4.82312481 1.94960837 4.5898756 1.99194433L4.5 2H2V4.50016667C1.99991812 4.74562654 1.82297492 4.94971605 1.58971162 4.99197426L1.49983331 5C1.25437343 4.99991815 1.05028392 4.82297495 1.00802571 4.58971165L1 4.49983333 1.001 1.49983333C1.0010818 1.25443131 1.17794474 1.05036951 1.41114451 1.0080521L1.50099997 1 4.5 1Z" />
                            </g>
                          </g>
                        </g>
                      </g>
                    </g>
                  </g>
                </svg>
              </div>
            </template>

            <template v-else-if="column.key === 'point'">
              {{ record.point || '0' }}
            </template>

            <template v-else-if="column.key === 'balance'">
              {{ formatCurrency(record.balance) }}
            </template>

            <template v-else-if="column.key === 'classStudentTuitionAccountInfo'">
              <div class="text-12px leading-20px">
                <div>
                  {{ getAccountName(record) }}
                  <a class="ml-6px" @click="handleSwitchAccount">切换</a>
                </div>
                <div class="text-#666">
                  剩余课时：{{ formatQuantity(getRemainQuantity(record), record.classStudentTuitionAccountInfo?.lessonChargingMode) }}
                </div>
                <div class="text-#666">
                  有效期至：{{ formatExpireDate(record) }}
                </div>
              </div>
            </template>

            <template v-else-if="column.key === 'usedQuantity'">
              {{ formatQuantity(getUsedQuantity(record), record.classStudentTuitionAccountInfo?.lessonChargingMode) }}
            </template>

            <template v-else-if="column.key === 'usedClassTime'">
              {{ formatQuantity(Number(record.usedClassTime || 0), 1) }}
            </template>

            <template v-else-if="column.key === 'totalQuantity'">
              <div class="leading-20px">
                <div>{{ formatQuantity(Number(record.totalQuantity || 0), record.classStudentTuitionAccountInfo?.lessonChargingMode) }}</div>
                <div class="text-#999 text-12px">
                  {{ getQuantityDetailText(record) }}
                </div>
              </div>
            </template>

            <template v-else-if="column.key === 'quantity'">
              {{ formatQuantity(getRemainQuantity(record), record.classStudentTuitionAccountInfo?.lessonChargingMode) }}
            </template>

            <template v-else-if="column.key === 'totalTuition'">
              {{ formatCurrency(Number(record.totalTuition || 0)) }}
            </template>

            <template v-else-if="column.key === 'confirmedTuition'">
              {{ formatCurrency(Number(record.confirmedTuition || 0)) }}
            </template>

            <template v-else-if="column.key === 'tuition'">
              {{ formatCurrency(Number(record.tuition || 0)) }}
            </template>

            <template v-else-if="column.key === 'expireTime'">
              {{ formatExpireDate(record) }}
            </template>

            <template v-else-if="column.key === 'suspendedTime'">
              {{ formatDate(record.suspendedTime) }}
            </template>

            <template v-else-if="column.key === 'classEndingTime'">
              {{ formatDate(record.classEndingTime) }}
            </template>

            <template v-else-if="column.key === 'joinTime'">
              {{ formatDate(record.joinTime) }}
            </template>

            <template v-else-if="column.key === 'status'">
              <span class="rounded-2.5 inline-block text-3 pt-0.5 pb-0.5 pl-2 pr-2" :class="getClassStatusInfo(record).className">
                {{ getClassStatusInfo(record).text }}
              </span>
            </template>

            <template v-else-if="column.key === 'studentAttendCount'">
              {{ getRecordCount(record.id).studentAttendCount }}
            </template>

            <template v-else-if="column.key === 'studentLeaveCount'">
              {{ getRecordCount(record.id).studentLeaveCount }}
            </template>

            <template v-else-if="column.key === 'action'">
              <a-space :size="12">
                <a @click="handleMoveClass">调至其他班</a>
                <a @click="handleRemoveClass">移出本班</a>
              </a-space>
            </template>
          </template>
        </a-table>
      </div>

      <ClassAddStudentModal
        v-model:open="addStudentVisible"
        :title="className"
        :lesson-name="lessonName"
        :class-id="classId"
        :lesson-id="lessonId"
        @success="handleAddStudentSuccess"
      />
    </div>
  </div>
</template>

<style lang="less" scoped></style>
