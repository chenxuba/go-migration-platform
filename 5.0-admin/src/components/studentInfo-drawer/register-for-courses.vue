<script setup>
import Icon, { QuestionCircleOutlined } from '@ant-design/icons-vue'
import { computed, ref, watch } from 'vue'
import dayjs from 'dayjs'
import { useStudentStore } from '@/stores/student'
import { getTuitionAccountReadingListApi } from '@/api/edu-center/tuition-account'
import messageService from '~@/utils/messageService'
import revokeCloseCourseModal from '@/components/common/revoke-close-course-modal.vue'
// 转课抽屉
import transferClassDrawer from './transferClassDrawer.vue'
// 退课抽屉
import dropTheClassDrawer from './dropTheClassDrawer.vue'
// 停课抽屉
import stopTheClassModal from './stopTheClassModal.vue'
// 结课抽屉
import endTheClassModal from './endTheClassModal.vue'
// 停/复课记录抽屉
import suspensionResumeModal from './suspensionResumeModal.vue'
// 学费变动记录抽屉
import feeChangeModal from './feeChangeModal.vue'
// 剩余详情抽屉
import remainingDetailsModal from './remainingDetailsModal.vue'
// 一对一模态框
import oneToOneModal from './oneToOneModal.vue'
import closeCourseRecordModal from './close-course-record-modal.vue'

const STATUS_END_STAMP_SRC = 'https://prod-tbu-next-erp-cdn.schoolpal.cn/next-pc-static/static/14983/static/svg/status-end-D4a6u7st.svg'
const STATUS_SUSPEND_STAMP_SRC = 'https://prod-tbu-next-erp-cdn.schoolpal.cn/next-pc-static/static/14983/static/svg/status-suspend-C7VTrkvF.svg'

const props = defineProps({
  emptyText: {
    type: String,
    default: '没有报读课程',
  },
})

const studentStore = useStudentStore()
const transferClassDrawerOpen = ref(false)
const dropTheClassDrawerOpen = ref(false)
const stopTheClassDrawerOpen = ref(false)
const endTheClassDrawerOpen = ref(false)
const revokeCloseCourseModalOpen = ref(false)
const closeCourseRecordModalOpen = ref(false)
const suspensionResumeDrawerOpen = ref(false)
const feeChangeDrawerOpen = ref(false)
const remainingDetailsModalOpen = ref(false)
const oneToOneModalOpen = ref(false)
const loading = ref(false)
const tuitionAccountList = ref([])
const listLoaded = ref(false)
const currentRevokeCourseRecord = ref(null)
const currentCloseRecordCourse = ref(null)
const currentEndCourseRecord = ref(null)
const currentStopCourseRecord = ref(null)

function handleOneToOne() {
  oneToOneModalOpen.value = true
}

function resetListState() {
  tuitionAccountList.value = []
  listLoaded.value = false
}

// 获取报读列表
async function getTuitionAccountList() {
  const studentId = studentStore.studentId
  if (!studentId) {
    return
  }

  loading.value = true
  listLoaded.value = false
  try {
    const res = await getTuitionAccountReadingListApi({
      sortModel: {},
      queryModel: {
        studentId: String(studentId),
      },
      pageRequestModel: {
        needTotal: true,
        pageSize: 100,
        pageIndex: 1,
        skipCount: 0,
      },
    })

    if (res.code === 200 && res.result) {
      const accounts =
        res.result.studentTutionAccounts
        ?? res.result.list
        ?? []
      tuitionAccountList.value = Array.isArray(accounts) ? accounts : []
      listLoaded.value = true
    } else {
      tuitionAccountList.value = []
      messageService.error(res.message || '获取报读列表失败')
    }
  } catch (error) {
    console.error('获取报读列表失败:', error)
    messageService.error('获取数据失败')
  } finally {
    loading.value = false
  }
}

// 格式化日期
function formatDate(dateStr) {
  if (!dateStr || dateStr === '0001-01-01T00:00:00' || dateStr === '') return '-'
  return dayjs(dateStr).format('YYYY-MM-DD')
}

// 格式化金额
function formatMoney(amount) {
  if (amount === null || amount === undefined) return '0.00'
  const num = Number(amount)
  if (isNaN(num)) return '0.00'
  return num.toFixed(2)
}

function formatCount(value) {
  const num = Number(value || 0)
  if (Number.isInteger(num)) {
    return String(num)
  }
  return num.toFixed(2)
}

function hasArrearTuition(item) {
  return Number(item?.arrearTuition || 0) > 0
}

function getArrearTuitionTooltip(item) {
  return `欠费学费金额：¥ ${formatMoney(item?.arrearTuition || 0)}`
}

function isTuitionAccountClosedByStatus(item) {
  return Number(item?.status || 0) === 3
}

function isTuitionAccountSuspended(item) {
  return Number(item?.status || 0) === 2
}

function hasPlannedSuspend(item) {
  if (isTuitionAccountSuspended(item))
    return false
  if (!item?.planSuspendTime)
    return false
  const parsed = dayjs(item.planSuspendTime)
  return parsed.isValid() && parsed.isAfter(dayjs(), 'day')
}

function getDisplayedRemainQuantity(item) {
  if (isTuitionAccountClosedByStatus(item))
    return 0
  return Number(item?.remainQuantity || 0)
}

function getDisplayedRemainFreeQuantity(item) {
  if (isTuitionAccountClosedByStatus(item))
    return 0
  return Number(item?.remainFreeQuantity || 0)
}

function getDisplayedRemainTuition(item) {
  if (isTuitionAccountClosedByStatus(item))
    return 0
  return Number(item?.tuition || 0)
}

/** 报读列表按课程聚合：剩余数量与剩余学费均为 0、曾有报读且无欠费时视为已结课（含手动结课扣完） */
function isTuitionAccountCourseEnded(item) {
  const totalTuition = Number(item?.totalTuition || 0)
  const totalQty = Number(item?.totalQuantity || 0)
  const totalFree = Number(item?.totalFreeQuantity || 0)
  if (totalTuition <= 0 && totalQty + totalFree <= 0)
    return false
  if (Number(item?.arrearTuition || 0) > 0.01)
    return false
  if (isTuitionAccountClosedByStatus(item))
    return true
  const remainQty = getDisplayedRemainQuantity(item) + getDisplayedRemainFreeQuantity(item)
  const remainTuition = getDisplayedRemainTuition(item)
  return remainQty <= 0.0001 && remainTuition <= 0.0001
}

function onEndedMenuRenew() {
  messageService.info('续费功能开发中')
}
function onEndedMenuRevokeGraduate(item) {
  currentRevokeCourseRecord.value = item || null
  revokeCloseCourseModalOpen.value = true
}

async function handleRevokeCloseCourseSuccess() {
  revokeCloseCourseModalOpen.value = false
  currentRevokeCourseRecord.value = null
  await getTuitionAccountList()
}

async function handleCloseCourseSuccess() {
  endTheClassDrawerOpen.value = false
  currentEndCourseRecord.value = null
  await getTuitionAccountList()
}

async function handleStopCourseSuccess() {
  stopTheClassDrawerOpen.value = false
  currentStopCourseRecord.value = null
  await getTuitionAccountList()
}

async function handleCloseCourseRecordSuccess() {
  await getTuitionAccountList()
}

function onMenuCloseCourse(item) {
  currentEndCourseRecord.value = item || null
  endTheClassDrawerOpen.value = true
}

function onMenuStopCourse(item) {
  currentStopCourseRecord.value = item || null
  stopTheClassDrawerOpen.value = true
}

function onEndedMenuCloseRecord(item) {
  currentCloseRecordCourse.value = item || null
  closeCourseRecordModalOpen.value = true
}

// 格式化有效期文本
function formatValidityText(item) {
  const enableExpireTime = item.enableExpireTime
  
  if (!enableExpireTime) {
    return '有效期至：不限制'
  }
  
  const expireTime = formatDate(item.expireTime)
  return `有效期至：${expireTime}`
}

// 格式化有效时段文本
function formatValidPeriodText(item) {
  const validDate = formatDate(item.validDate || item.activedAt)
  const endDate = formatDate(item.endDate || item.expireTime)
  
  if (validDate === '-' || endDate === '-') {
    return '有效时段：-'
  }
  
  return `有效时段：${validDate} ~ ${endDate}`
}

function getSuspendStatusText(item) {
  if (isTuitionAccountSuspended(item)) {
    const resumeDate = formatDate(item?.planResumeTime)
    return resumeDate !== '-' ? `有效期至：将于${resumeDate}复课` : '有效期至：停课中'
  }
  if (hasPlannedSuspend(item)) {
    const suspendDate = formatDate(item?.planSuspendTime)
    return suspendDate !== '-' ? `有效期至：将于${suspendDate}停课` : ''
  }
  return ''
}

function getSuspendStatusClass(item) {
  if (isTuitionAccountSuspended(item))
    return 'text-#f90'
  if (hasPlannedSuspend(item))
    return 'text-#ff4d4f'
  return 'text-#888'
}

// 获取授课方式文本
function getLessonTypeText(type) {
  const typeMap = {
    1: '班级授课',
    2: '1v1授课',
  }
  return typeMap[type] || '-'
}

// 获取计费模式文本
function getChargingModeText(mode) {
  const modeMap = {
    1: '课时',
    2: '时段',
    3: '金额',
  }
  return modeMap[mode] || '-'
}

// 格式化剩余数量显示文本
function formatRemainQuantityText(item) {
  const mode = item.lessonChargingMode
  const remainQty = getDisplayedRemainQuantity(item)
  const remainFreeQty = getDisplayedRemainFreeQuantity(item)
  const totalQty = Number(item.totalQuantity || 0)
  const totalFreeQty = Number(item.totalFreeQuantity || 0)
  
  // 获取单位和标签
  let unit = '课时'
  let label = '剩余课时'
  let prefix = ''
  if (mode === 2) {
    unit = '天'
    label = '剩余天数'
  } else if (mode === 3) {
    unit = ''
    label = '剩余金额'
    prefix = '¥'
  }
  
  // 当前剩余 = 剩余购买 + 剩余赠送
  const remainTotal = remainQty + remainFreeQty
  // 总计 = 总购买 + 总赠送
  const total = totalQty + totalFreeQty
  
  if (remainFreeQty > 0) {
    // 有赠送：剩余课时：29（含赠 2 课时）|（总计 31 课时）
    // 按金额：剩余金额：¥41000（含¥11400）|（总¥51000）
    if (mode === 3) {
      return `${label}：${prefix}${formatCount(remainTotal)}（含赠${prefix}${formatCount(remainFreeQty)}）|（总计${prefix}${formatCount(totalQty)}）`
    } else {
      return `${label}：${formatCount(remainTotal)}（含赠 ${formatCount(remainFreeQty)} ${unit}）|（总计 ${formatCount(total)} ${unit}）`
    }
  } else {
    // 没有赠送：剩余课时：29（总计 31 课时）
    // 按金额：剩余金额：¥960（总¥1000）
    if (mode === 3) {
      return `${label}：${prefix}${formatCount(remainQty)}（总${prefix}${formatCount(total)}）`
    } else {
      return `${label}：${formatCount(remainQty)}（总计 ${formatCount(total)} ${unit}）`
    }
  }
}

// 暴露方法和数据给父组件调用
defineExpose({
  getTuitionAccountList,
  tuitionAccountList,
  listLoaded,
  resetListState,
})

watch(endTheClassDrawerOpen, (value) => {
  if (!value)
    currentEndCourseRecord.value = null
})
</script>

<template>
  <div class="register-for-courses h-full flex flex-col">
    <a-spin :spinning="loading" class="flex-1 flex flex-col">
      <!-- 空状态 -->
      <div v-if="!loading && tuitionAccountList.length === 0" class="empty-state bg-white rounded-2 flex flex-col items-center justify-center flex-1">
        <img 
          src="https://prod-tbu-next-erp-cdn.schoolpal.cn/next-pc-static/static/14147/static/png/no-data-C70XJ9II.png" 
          alt="暂无数据"
          style="width: 120px; height: 120px; margin-bottom: 16px;"
        >
        <div class="text-#999 text-4  text-sm mt--15px">{{ props.emptyText }}</div>
      </div>
      
      <!-- 列表数据 -->
      <div v-else class="p3 pr0">
        <a-space class="flex flex-wrap">
        <div v-for="(item, index) in tuitionAccountList" :key="item.id || index" class=" bg-white w-92.5 pb-2 rounded-2">
          <div class="t flex justify-between p4 pb0 flex-items-center">
            <span
              class="text-4 text-#222 font-500 flex-1 line-clamp-1 mr-10"
              :class="{ 'course-ended-faded': isTuitionAccountCourseEnded(item) }"
            >{{ item.lessonName || '-' }}</span>
          <a-dropdown placement="bottom" :arrow="{ pointAtCenter: true }">
            <a class="ant-dropdown-link rounded-10 hover:bg-#e6f0ff w-24px h-24px" @click.prevent>
              <img
                src="https://prod-tbu-next-erp-cdn.schoolpal.cn/next-pc-static/static/12181/static/menu-normal.3cae936f.svg"
                alt=""
              >
            </a>
            <template #overlay>
              <a-space direction="vertical" :size="1">
                <template v-if="!isTuitionAccountCourseEnded(item)">
                <div class="flex items-center gap-2" @click="transferClassDrawerOpen = true">
                  <div>
                    <Icon :style="{ color: 'hotpink' }">
                      <template #component>
                        <svg
                          width="14px" height="14px" viewBox="0 0 14 14" version="1.1"
                          xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink"
                        >
                          <title>编组 18备份</title>
                          <g id="页面-2" stroke="none" stroke-width="1" fill="none" fill-rule="evenodd">
                            <g id="画板备份-25" transform="translate(-1314.000000, -823.000000)">
                              <g id="编组-18备份" transform="translate(1314.000000, 823.000000)">
                                <circle id="椭圆形备份-2" fill="#0066FF" cx="7" cy="7" r="7" />
                                <path
                                  id="路径备份"
                                  d="M9.7570082,6.46593375 L7.49232985,4.2012554 C7.31545958,4.02438512 7.02874679,4.02438512 6.85187651,4.2012554 L4.58719816,6.46593375 C4.41032789,6.64280403 4.41032789,6.92951682 4.58719816,7.10638709 C4.76406844,7.28325737 5.05078123,7.28325737 5.22765151,7.10638709 L6.71917543,5.61486317 L6.71917543,9.95670747 C6.71917543,10.2069556 6.92186824,10.4096484 7.17211638,10.4096484 C7.42236451,10.4096484 7.62505732,10.2069556 7.62505732,9.95670747 L7.62505732,5.61486317 L9.11658124,7.10638709 C9.20490205,7.19471669 9.32085718,7.23909697 9.43681231,7.23909697 C9.55276745,7.23909697 9.66871379,7.19493657 9.75704338,7.10638709 C9.93391366,6.92951682 9.93391366,6.64280403 9.75704338,6.46593375 L9.7570082,6.46593375 Z"
                                  fill="#FFFFFF" fill-rule="nonzero"
                                  transform="translate(7.172121, 7.239126) scale(-1, -1) rotate(-90.000000) translate(-7.172121, -7.239126) "
                                />
                              </g>
                            </g>
                          </g>
                        </svg>
                      </template>
                    </Icon>
                  </div>
                  <span class="font-size-14px text-#666">转课</span>
                </div>
                <div class="flex items-center gap-2" @click="dropTheClassDrawerOpen = true">
                  <div>
                    <Icon :style="{ color: 'hotpink' }">
                      <template #component>
                        <svg
                          xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="14px"
                          height="14px" viewBox="0 0 14 14" version="1.1"
                        >
                          <title>编组 44备份</title>
                          <g id="页面-2" stroke="none" stroke-width="1" fill="none" fill-rule="evenodd">
                            <g id="画板备份-25" transform="translate(-1382.000000, -823.000000)">
                              <g id="编组-42" transform="translate(1382.000000, 823.000000)">
                                <circle id="椭圆形备份-2" fill="#FF3333" cx="7" cy="7" r="7" />
                                <path
                                  id="路径备份"
                                  d="M9.7570082,6.21551287 L7.49232985,3.95083453 C7.31545958,3.77396425 7.02874679,3.77396425 6.85187651,3.95083453 L4.58719816,6.21551287 C4.41032789,6.39238315 4.41032789,6.67909594 4.58719816,6.85596622 C4.76406844,7.0328365 5.05078123,7.0328365 5.22765151,6.85596622 L6.71917543,5.3644423 L6.71917543,9.70628659 C6.71917543,9.95653473 6.92186824,10.1592275 7.17211638,10.1592275 C7.42236451,10.1592275 7.62505732,9.95653473 7.62505732,9.70628659 L7.62505732,5.3644423 L9.11658124,6.85596622 C9.20490205,6.94429582 9.32085718,6.9886761 9.43681231,6.9886761 C9.55276745,6.9886761 9.66871379,6.94451569 9.75704338,6.85596622 C9.93391366,6.67909594 9.93391366,6.39238315 9.75704338,6.21551287 L9.7570082,6.21551287 Z"
                                  fill="#FFFFFF" fill-rule="nonzero"
                                  transform="translate(7.172121, 6.988705) scale(-1, -1) translate(-7.172121, -6.988705) "
                                />
                              </g>
                            </g>
                          </g>
                        </svg>
                      </template>
                    </Icon>
                  </div>
                  <span class="font-size-14px text-#666  w-90px ">退课</span>
                </div>
                <div class="flex items-center gap-2" @click="onMenuStopCourse(item)">
                  <div>
                    <Icon :style="{ color: 'hotpink' }">
                      <template #component>
                        <svg
                          xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="14px"
                          height="14px" viewBox="0 0 14 14" version="1.1"
                        >
                          <title>编组 19</title>
                          <g id="页面-2" stroke="none" stroke-width="1" fill="none" fill-rule="evenodd">
                            <g id="画板备份-25" transform="translate(-1348.000000, -823.000000)">
                              <g id="编组-19" transform="translate(1348.000000, 823.000000)">
                                <circle id="椭圆形" fill="#FF9900" cx="7" cy="7" r="7" />
                                <rect
                                  id="矩形备份-2" fill="#FFFFFF" x="4.45454545" y="6.36363636" width="5.09090909"
                                  height="1.27272727" rx="0.636363636"
                                />
                              </g>
                            </g>
                          </g>
                        </svg>
                      </template>
                    </Icon>
                  </div>
                  <span class="font-size-14px text-#666 w-90px">停课</span>
                </div>
                <div class="flex items-center gap-2" @click="onMenuCloseCourse(item)">
                  <div>
                    <Icon :style="{ color: 'hotpink' }">
                      <template #component>
                        <svg
                          width="14px" height="14px" viewBox="0 0 14 14" version="1.1"
                          xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink"
                        >
                          <title>编组 39备份</title>
                          <g id="页面-2" stroke="none" stroke-width="1" fill="none" fill-rule="evenodd">
                            <g id="画板备份-25" transform="translate(-1445.000000, -823.000000)">
                              <g id="编组-39备份" transform="translate(1445.000000, 823.000000)">
                                <circle id="椭圆形" fill="#666666" cx="7" cy="7" r="7" />
                                <g id="编组" transform="translate(3.818182, 3.432239)" fill="#FFFFFF" fill-rule="nonzero">
                                  <path
                                    id="形状"
                                    d="M5.00560859,1.01470833 C5.21888805,1.17156667 5.40853247,1.34925 5.57442903,1.54781667 C5.74032559,1.74638333 5.8825307,1.96081667 6.00104436,2.19123333 C6.11955802,2.42165 6.20958553,2.66548333 6.2711833,2.92285 C6.33283748,3.18021667 6.36363636,3.4412 6.36363636,3.70591667 C6.36363636,4.16179167 6.27953172,4.58949167 6.1112096,4.98901667 C5.9430003,5.38854167 5.71539316,5.73655833 5.42861379,6.033125 C5.14183443,6.32969167 4.80524661,6.56500833 4.41896315,6.73901667 C4.0326797,6.913025 3.61915107,7 3.17826446,7 C2.74222896,7 2.33095667,6.913025 1.94467321,6.73895833 C1.55833335,6.56500833 1.22061736,6.32963333 0.931468854,6.03306667 C0.642320344,5.7365 0.414826013,5.388425 0.248873044,4.98895833 C0.082920076,4.58949167 0,4.16173333 0,3.70585833 C0,3.4461 0.0296143129,3.191125 0.0888993468,2.94116667 C0.148127973,2.69115 0.232232621,2.45344167 0.341326109,2.22798333 C0.450363188,2.00246667 0.585404455,1.79165 0.746619133,1.59565 C0.907777404,1.39953333 1.08788883,1.223075 1.28695342,1.06615833 C1.39130862,0.987758333 1.50384301,0.958358333 1.62472582,0.977958333 C1.74560862,0.997558333 1.84392814,1.05886667 1.91974078,1.16176667 C1.99560983,1.264725 2.02409598,1.37993333 2.00514282,1.50739167 C1.98613325,1.63485 1.92690462,1.73780833 1.82734412,1.81620833 C1.52866263,2.04166667 1.29998372,2.31863333 1.1411946,2.64705 C0.982405472,2.97546667 0.902926297,3.32844167 0.902926297,3.70585833 C0.902926297,4.02943333 0.962211331,4.33451667 1.08072499,4.62128333 C1.19923865,4.90805 1.36158149,5.15806667 1.56775352,5.371275 C1.77398195,5.58448333 2.01569115,5.75365 2.29299394,5.8786 C2.57029672,6.00366667 2.86531169,6.06614167 3.17815165,6.06614167 C3.49104802,6.06614167 3.78611939,6.00366667 4.06336577,5.8786 C4.34066855,5.75359167 4.58243416,5.58448333 4.78860619,5.371275 C4.99483462,5.15806667 5.15836203,4.90805 5.27918843,4.62128333 C5.40012764,4.33451667 5.46054084,4.02943333 5.46054084,3.70585833 C5.46054084,3.32348333 5.37519521,2.96199167 5.20456036,2.621325 C5.03392551,2.28065833 4.79452905,2.00001667 4.48648379,1.7794 C4.38212859,1.7059 4.31934624,1.60539167 4.29802394,1.47799167 C4.27670163,1.350475 4.30157766,1.232875 4.37265201,1.12501667 C4.44378277,1.02205833 4.54091771,0.959583333 4.66416966,0.937533333 C4.78753443,0.915425 4.9013098,0.941208333 5.00560859,1.01470833 L5.00560859,1.01470833 Z M3.17826446,3.72790833 C3.05501251,3.72790833 2.94958556,3.68258333 2.86192719,3.59193333 C2.77421242,3.501225 2.73038324,3.39214167 2.73038324,3.26468333 L2.73038324,0.470575 C2.73038324,0.343116667 2.77421242,0.232866667 2.86192719,0.139708333 C2.94964197,0.04655 3.05501251,0 3.17826446,0 C3.30631111,0 3.41416362,0.04655 3.50176558,0.139708333 C3.58953676,0.232866667 3.63336594,0.343116667 3.63336594,0.470575 L3.63336594,3.26468333 C3.63336594,3.39214167 3.58953676,3.501225 3.50176558,3.59193333 C3.41410721,3.68258333 3.3062547,3.72790833 3.17826446,3.72790833 L3.17826446,3.72790833 Z"
                                  />
                                </g>
                              </g>
                            </g>
                          </g>
                        </svg>
                      </template>
                    </Icon>
                  </div>
                  <span class="font-size-14px text-#666 w-90px">结课</span>
                </div>
                <div class="flex items-center gap-2">
                  <div>
                    <Icon :style="{ color: 'hotpink' }">
                      <template #component>
                        <svg
                          width="14px" height="14px" viewBox="0 0 14 14" version="1.1"
                          xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink"
                        >
                          <title>编组 2</title>
                          <g id="页面-1" stroke="none" stroke-width="1" fill="none" fill-rule="evenodd">
                            <g id="UC12-学员详情增加快速续费" transform="translate(-466.000000, -466.000000)">
                              <g id="编组-2" transform="translate(466.000000, 466.000000)">
                                <circle id="椭圆形" fill="#00CC33" cx="7" cy="7" r="7" />
                                <g id="编组" transform="translate(3.000000, 3.000000)" fill="#FFFFFF" fill-rule="nonzero">
                                  <path
                                    id="形状结合"
                                    d="M5.81333333,0.693333333 C5.87224371,0.614786169 5.98367506,0.598867404 6.06222222,0.657777778 L6.06222222,0.657777778 L7.36592593,1.63555556 C7.37940249,1.64566298 7.39137406,1.65763455 7.40148148,1.67111111 C7.46039185,1.74965828 7.44447309,1.86108963 7.36592593,1.92 L7.36592593,1.92 L6.06222222,2.89777778 C6.03144957,2.92085727 5.99402137,2.93333333 5.95555556,2.93333333 C5.8573716,2.93333333 5.77777778,2.85373951 5.77777778,2.75555556 L5.77777778,2.75555556 L5.77744444,2.22221176 L2.66666667,2.22222222 C1.97360345,2.22222222 1.40404562,2.75101163 1.33943696,3.42714666 L1.33333333,3.55555556 L1.33333333,5.33333333 C1.33333333,6.02639655 1.86212274,6.59595438 2.53825777,6.66056304 L2.66666667,6.66666667 L5.33333333,6.66666667 C6.02639655,6.66666667 6.59595438,6.13787726 6.66056304,5.46174223 L6.66666667,5.33333333 L6.66666667,4.44444444 C6.66666667,4.19898456 6.86565122,4 7.11111111,4 C7.32929768,4 7.510763,4.15722236 7.54839496,4.36455499 L7.55555556,4.44444444 L7.55555556,5.33333333 C7.55555556,6.5115408 6.63863472,7.47558993 5.47944506,7.55082873 L5.33333333,7.55555556 L2.66666667,7.55555556 C1.4884592,7.55555556 0.524410069,6.63863472 0.449171271,5.47944506 L0.444444444,5.33333333 L0.444444444,3.55555556 C0.444444444,2.37734809 1.36136528,1.41329896 2.52055494,1.33806016 L2.66666667,1.33333333 L5.77744444,1.33321176 L5.77777778,0.8 C5.77777778,0.774356123 5.7833227,0.749173409 5.79385768,0.72611643 Z"
                                  />
                                </g>
                              </g>
                            </g>
                          </g>
                        </svg>
                      </template>
                    </Icon>
                  </div>
                  <span class="font-size-14px text-#666 w-90px">续费</span>
                </div>
                <div class="flex items-center gap-2" @click="feeChangeDrawerOpen = true">
                  <span class="image-wrapper" />
                  <span class="font-size-14px text-#666 w-90px">学费变动记录</span>
                </div>
                <div class="flex items-center gap-2" @click="suspensionResumeDrawerOpen = true">
                  <span class="image-wrapper suspendResumen" />
                  <span class="font-size-14px text-#666 w-90px">停/复课记录</span>
                </div>
                <div class="flex items-center gap-2" @click="onEndedMenuCloseRecord(item)">
                  <div>
                    <Icon :style="{ color: 'hotpink' }">
                      <template #component>
                        <svg width="14px" height="14px" viewBox="0 0 14 14" xmlns="http://www.w3.org/2000/svg">
                          <circle fill="#666666" cx="7" cy="7" r="7" />
                          <rect fill="#FFFFFF" x="4.2" y="3.8" width="5.6" height="6.4" rx="0.6" />
                          <rect fill="#666666" x="5.1" y="5.2" width="3.8" height="0.55" rx="0.2" />
                          <rect fill="#666666" x="5.1" y="6.35" width="2.8" height="0.55" rx="0.2" />
                        </svg>
                      </template>
                    </Icon>
                  </div>
                  <span class="font-size-14px text-#666 w-90px">结课记录</span>
                </div>
                </template>
                <template v-else>
                <div class="flex items-center gap-2" @click="onEndedMenuRenew">
                  <div>
                    <Icon :style="{ color: 'hotpink' }">
                      <template #component>
                        <svg
                          width="14px" height="14px" viewBox="0 0 14 14" version="1.1"
                          xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink"
                        >
                          <g stroke="none" stroke-width="1" fill="none" fill-rule="evenodd">
                            <g transform="translate(-466.000000, -466.000000)">
                              <g transform="translate(466.000000, 466.000000)">
                                <circle fill="#00CC33" cx="7" cy="7" r="7" />
                                <g transform="translate(3.000000, 3.000000)" fill="#FFFFFF" fill-rule="nonzero">
                                  <path
                                    d="M5.81333333,0.693333333 C5.87224371,0.614786169 5.98367506,0.598867404 6.06222222,0.657777778 L7.36592593,1.63555556 C7.46039185,1.74965828 7.44447309,1.86108963 7.36592593,1.92 L6.06222222,2.89777778 C5.99402137,2.93333333 5.8573716,2.93333333 5.77777778,2.75555556 L5.77744444,2.22221176 L2.66666667,2.22222222 C1.97360345,2.22222222 1.40404562,2.75101163 1.33943696,3.42714666 L1.33333333,3.55555556 L1.33333333,5.33333333 C1.33333333,6.02639655 1.86212274,6.59595438 2.53825777,6.66056304 L2.66666667,6.66666667 L5.33333333,6.66666667 C6.02639655,6.66666667 6.59595438,6.13787726 6.66056304,5.46174223 L6.66666667,5.33333333 L6.66666667,4.44444444 C6.66666667,4.19898456 6.86565122,4 7.11111111,4 C7.32929768,4 7.510763,4.15722236 7.54839496,4.36455499 L7.55555556,4.44444444 L7.55555556,5.33333333 C7.55555556,6.5115408 6.63863472,7.47558993 5.47944506,7.55082873 L5.33333333,7.55555556 L2.66666667,7.55555556 C1.4884592,7.55555556 0.524410069,6.63863472 0.449171271,5.47944506 L0.444444444,5.33333333 L0.444444444,3.55555556 C0.444444444,2.37734809 1.36136528,1.41329896 2.52055494,1.33806016 L2.66666667,1.33333333 L5.77744444,1.33321176 L5.77777778,0.8 C5.77777778,0.774356123 5.7833227,0.749173409 5.79385768,0.72611643 Z"
                                  />
                                </g>
                              </g>
                            </g>
                          </g>
                        </svg>
                      </template>
                    </Icon>
                  </div>
                  <span class="font-size-14px text-#666 w-90px">续费</span>
                </div>
                <div class="flex items-center gap-2" @click="onEndedMenuRevokeGraduate(item)">
                  <div>
                    <Icon :style="{ color: 'hotpink' }">
                      <template #component>
                        <svg width="14px" height="14px" viewBox="0 0 14 14" xmlns="http://www.w3.org/2000/svg">
                          <circle fill="#666666" cx="7" cy="7" r="7" />
                          <path
                            fill="#FFFFFF"
                            fill-rule="nonzero"
                            d="M9.7570082,6.46593375 L7.49232985,4.2012554 C7.31545958,4.02438512 7.02874679,4.02438512 6.85187651,4.2012554 L4.58719816,6.46593375 C4.41032789,6.64280403 4.41032789,6.92951682 4.58719816,7.10638709 C4.76406844,7.28325737 5.05078123,7.28325737 5.22765151,7.10638709 L6.71917543,5.61486317 L6.71917543,9.95670747 C6.71917543,10.2069556 6.92186824,10.4096484 7.17211638,10.4096484 C7.42236451,10.4096484 7.62505732,10.2069556 7.62505732,9.95670747 L7.62505732,5.61486317 L9.11658124,7.10638709 C9.20490205,7.19471669 9.32085718,7.23909697 9.43681231,7.23909697 C9.55276745,7.23909697 9.66871379,7.19493657 9.75704338,7.10638709 C9.93391366,6.92951682 9.93391366,6.64280403 9.75704338,6.46593375 Z"
                            transform="translate(7.172121, 7.239126) scale(-1, -1) rotate(-90.000000) translate(-7.172121, -7.239126)"
                          />
                        </svg>
                      </template>
                    </Icon>
                  </div>
                  <span class="font-size-14px text-#666 w-90px">撤销结课</span>
                </div>
                <div class="flex items-center gap-2" @click="feeChangeDrawerOpen = true">
                  <span class="image-wrapper" />
                  <span class="font-size-14px text-#666 w-90px">学费变动记录</span>
                </div>
                <div class="flex items-center gap-2" @click="suspensionResumeDrawerOpen = true">
                  <span class="image-wrapper suspendResumen" />
                  <span class="font-size-14px text-#666 w-90px">停/复课记录</span>
                </div>
                <div class="flex items-center gap-2" @click="onEndedMenuCloseRecord(item)">
                  <div>
                    <Icon :style="{ color: 'hotpink' }">
                      <template #component>
                        <svg width="14px" height="14px" viewBox="0 0 14 14" xmlns="http://www.w3.org/2000/svg">
                          <circle fill="#666666" cx="7" cy="7" r="7" />
                          <rect fill="#FFFFFF" x="4.2" y="3.8" width="5.6" height="6.4" rx="0.6" />
                          <rect fill="#666666" x="5.1" y="5.2" width="3.8" height="0.55" rx="0.2" />
                          <rect fill="#666666" x="5.1" y="6.35" width="2.8" height="0.55" rx="0.2" />
                        </svg>
                      </template>
                    </Icon>
                  </div>
                  <span class="font-size-14px text-#666 w-90px">结课记录</span>
                </div>
                </template>
              </a-space>
            </template>
          </a-dropdown>
        </div>
        <div class="course-card-body">
          <div class="course-card-main min-w-0 flex-1">
            <div class="tag px3 mt1" :class="{ 'course-ended-faded': isTuitionAccountCourseEnded(item) }">
              <a-space>
                <span v-if="item.lessonType" class="bg-#e6f0ff text-#06f text-3 rounded-10 px3 py1">{{ getLessonTypeText(item.lessonType) }}</span>
                <span v-if="item.lessonChargingMode" class="bg-#e6f0ff text-#06f text-3 rounded-10 px3 py1">{{ getChargingModeText(item.lessonChargingMode) }}</span>
              </a-space>
            </div>
            <div class="remaining-class-hours px3 text-3 text-#888 mt-5 flex justify-between">
              <span>{{ formatRemainQuantityText(item) }}</span>
            </div>
            <div class="remaining-tuition px3 text-#888 flex justify-start flex-items-center mt-1">
              <span class="text-3">
                剩余学费：¥ {{ formatMoney(getDisplayedRemainTuition(item)) }}（总计 ¥ {{ formatMoney(item.totalTuition) }}
                <a-tooltip v-if="hasArrearTuition(item)" placement="top">
                  <template #title>
                    {{ getArrearTuitionTooltip(item) }}
                  </template>
                  <span class="arrear-badge ml-1">欠</span>
                </a-tooltip>
                ）
              </span>
            </div>
            <div v-if="getSuspendStatusText(item)" class="validity px3 flex justify-start flex-items-center mt-1">
              <span class="text-3" :class="getSuspendStatusClass(item)">{{ getSuspendStatusText(item) }}</span>
              <a-popover>
                <template #title>
                  <div style="border-bottom:1px solid #eee;padding-bottom: 6px;">
                    字段说明
                  </div>
                </template>
                <template #content>
                  <div class="text-#666">
                    停课后学员将暂停计费，并在计划复课日恢复。
                  </div>
                </template>
                <QuestionCircleOutlined class="ml-2 mb-0.5" />
              </a-popover>
            </div>
            <!-- 时段模式显示有效时段，非时段模式显示有效期至 -->
            <div v-else-if="item.lessonChargingMode === 2" class="valid-period px3 text-#888 flex justify-start flex-items-center mt-1">
              <span class="text-3">{{ formatValidPeriodText(item) }}</span>
            </div>
            <div v-else class="validity px3 text-#888 flex justify-start flex-items-center mt-1">
              <span class="text-3">{{ formatValidityText(item) }}</span>
              <a-popover>
                <template #title>
                  <div style="border-bottom:1px solid #eee;padding-bottom: 6px;">
                    字段说明
                  </div>
                </template>
                <template #content>
                  <div class="text-#666">
                    此有效期展示为历史报名订单中，最长的有效期。
                  </div>
                  <div class="text-#666">
                    具体每次报名有效期，请点击剩余详情进行查看。
                  </div>
                </template>
                <QuestionCircleOutlined class="ml-2 mb-0.5" />
              </a-popover>
            </div>
          </div>
          <div v-if="isTuitionAccountCourseEnded(item)" class="course-ended-stamp-wrap">
            <img
              class="course-ended-stamp"
              :src="STATUS_END_STAMP_SRC"
              alt="已结课"
              width="88"
              height="88"
            >
          </div>
          <div v-else-if="isTuitionAccountSuspended(item)" class="course-ended-stamp-wrap">
            <img
              class="course-ended-stamp"
              :src="STATUS_SUSPEND_STAMP_SRC"
              alt="停课中"
              width="120"
              height="120"
            >
          </div>
        </div>
        <div class="line w-100% h-0.25 bg-#eee mt2" />
        <div class="btn mt2 flex justify-end pr3">
          <a-tooltip>
            <template #title>
              点击查看剩余详情
            </template>
            <span
              class="bg-#e6f0ff mr-3 text-#06f text-3 rounded-10 px2 py0.5 pb0 flex-center cursor-pointer"
              @click="remainingDetailsModalOpen = true"
            ><img
              class="mt--0.6"
              src="https://prod-tbu-next-erp-cdn.schoolpal.cn/next-pc-static/static/12087/static/left-detail.8a056096.svg"
              alt=""
            > 剩余详情</span>
          </a-tooltip>
          <a-button
            v-if="item.lessonType === 2 && !isTuitionAccountCourseEnded(item)"
            size="small" class=" text-#666" style="font-size: 12px;border-radius: 10px;padding: 0 10px;"
            @click="handleOneToOne"
          >
            查看分班
          </a-button>
        </div>
        </div>
      </a-space>
      </div>
    </a-spin>
    <transferClassDrawer v-model:open="transferClassDrawerOpen" />
    <dropTheClassDrawer v-model:open="dropTheClassDrawerOpen" />
    <stopTheClassModal
      v-model:open="stopTheClassDrawerOpen"
      :record="currentStopCourseRecord"
      @success="handleStopCourseSuccess"
    />
    <endTheClassModal
      v-model:open="endTheClassDrawerOpen"
      :record="currentEndCourseRecord"
      @success="handleCloseCourseSuccess"
    />
    <revokeCloseCourseModal
      v-model:open="revokeCloseCourseModalOpen"
      :record="currentRevokeCourseRecord"
      @success="handleRevokeCloseCourseSuccess"
    />
    <closeCourseRecordModal
      v-model:open="closeCourseRecordModalOpen"
      :record="currentCloseRecordCourse"
      @success="handleCloseCourseRecordSuccess"
    />
    <suspensionResumeModal v-model:open="suspensionResumeDrawerOpen" />
    <feeChangeModal v-model:open="feeChangeDrawerOpen" />
    <remainingDetailsModal v-model:open="remainingDetailsModalOpen" />
    <oneToOneModal v-model:open="oneToOneModalOpen" />
  </div>
</template>

<style lang="less" scoped>
.register-for-courses {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.empty-state {
  // min-height: 500px; 
  height: calc(100vh - 300px);
  margin: 12px;
}

.image-wrapper {
  display: inline-flex;
  width: 14px;
  height: 14px;
  border-radius: 50%;
  background-color: #06f;
  background-image: url(data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABAAAAAQCAYAAAAf8/9hAAAAAXNSR0IArs4c6QAAAR5JREFUOE+t0kFLlUEUxvHfI0ipoKsWiTtBhNSN0MdpEbQMAgV3QbQKCQqXCq77BtEHENdtIlq2iTYFZoWBJwbuG7frfe2Cns1wZs7855nnnLhiZNz9qrqHL0netPOquouNJAej9X2Ar7iBVXzDR9zCXJKfw5A+wFM8wWt8wjYOk9yfVMEsPmAJv/ELK0k+9wKqagdbmBoUTTfJaCpPcTbYP8eLJM9b/vcLVdUkN6kdoJ3PDPIG6KIBdpM8+wfQ0413WMd8kpNxNWNN7AqranJAVT3Aw5EvLOMm3qNJb9HWvST7ox68xKMJB/NVkscXPKiq1r5hE49xB7fxvVOQ5Ef30PV50NOFI2xiIUkbpgvxPwVrWEzyts+bSwGTGPoHcjBZEcLVOqIAAAAASUVORK5CYII=);
  background-repeat: no-repeat;
  background-position: 50%;
  background-size: 8px auto;
}

.suspendResumen {
  background-size: 14px auto;
  background-image: url(data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABwAAAAcCAYAAAByDd+UAAAAAXNSR0IArs4c6QAAAQlJREFUSEvt1b0uRUEUhuHni1CoNW7BFShcgIRW6yfR+akkKgWJRKJ0aIVGoncRoncJElHoVGJkJ5ugOPaRfXYUZ5ppVtY76501a6LjlY55RsDWjY+U/l+lpZSb6nRJFvqdsu8dllImcIBljOMSe0lefiYtpZQa2Dfnb8Aj7P5Ifo+lJNX+udoCPmAalabHusIZVBWuJbn+ILYF/KaplDKJM6zgFXNJbivoUIBfqulhA70kW10AT7CJkyTbQwPWSk+xijfMJrlrE/iEKczjGef4aJr1JFdtN80xdrp8FmM4rBVW3AvsD+3hDzJI69FWkiz+ebQNAmwaO/oPm5pqHDdS2lhV08DOlb4DI/5/HWsZb+UAAAAASUVORK5CYII=);
}

.arrear-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 16px;
  height: 16px;
  padding: 0 4px;
  border-radius: 4px;
  background: #fbe7e6;
  color: #ff4d4f;
  font-size: 12px;
  line-height: 16px;
  cursor: pointer;
}

.course-card-body {
  display: flex;
  align-items: center;
  gap: 4px;
}

.course-ended-stamp-wrap {
  flex-shrink: 0;
  align-self: center;
  padding-right: 12px;
  padding-top: 8px;
  pointer-events: none;
}

.course-ended-stamp {
  display: block;
  width: 88px;
  height: 88px;
  object-fit: contain;
  opacity: 0.95;
}

.course-ended-faded {
  opacity: 0.4;
}

</style>
