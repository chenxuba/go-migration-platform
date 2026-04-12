<script setup>
import { CloseOutlined, EyeInvisibleOutlined, EyeOutlined, RightOutlined, CaretDownOutlined, CaretUpOutlined, ExclamationCircleOutlined } from '@ant-design/icons-vue'
import { computed, nextTick, ref, watch } from 'vue'
import { useWindowSize } from '@vueuse/core'
import { useRouter } from 'vue-router'
import { debounce } from 'lodash-es'
import DeleteConfirmModal from './DeleteConfirmModal.vue'
import AssignSalesModal from './assign-sales-modal.vue'
import CreateStudent from './create-student.vue'
import RegisterForCourses from '@/components/studentInfo-drawer/register-for-courses.vue'
import StudentClassRecord from '@/components/studentInfo-drawer/student-class-record.vue'
import OrderRecord from '@/components/studentInfo-drawer/order-record.vue'
import TryListeningRecord from '@/components/studentInfo-drawer/try-listening-record.vue'
import FollowUpRecord from '@/components/studentInfo-drawer/follow-up-record.vue'
import { getOneToOneListApi } from '@/api/edu-center/one-to-one'
import { useStudentFields } from '@/composables/useStudentFields'
import { batchAssignSalespersonApi, batchDeleteIntendedStudentApi, batchTransferToPublicPoolApi, getIntentStudentDetailApi, listStudentChangeInfoApi, updateIntendedStudentApi, updateStatusApi } from '@/api/enroll-center/intention-student'
import { getStudentPhoneNumberApi } from '@/api/common/config'
import { useStudentStore } from '@/stores/student'
import { useUserStore } from '@/stores/user'
import messageService from '~@/utils/messageService'
import emitter, { EVENTS } from '~@/utils/eventBus'
import { calculateAge } from '@/utils/date'
import { ParentRelationshipLabel, IntentionLevel, IntentionLevelLabel, IntentionLevelStyle, FollowUpStatus, FollowUpStatusLabel, FollowUpStatusStyle, Sex, SexLabel, StudentStatus, StudentStatusLabel } from '@/enums'
import dayjs from 'dayjs'
import { Empty } from 'ant-design-vue'
const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
  defaultActiveKey: {
    type: String,
    default: '0',
  },
})
const emit = defineEmits(['update:open'])
const activeKey = ref(props.defaultActiveKey)
// 处理双向绑定
const openDrawer = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

// defineEmits(['update:open']);
const openModal = ref(false)
function seeAllData() {
  // console.log('seeAllData')
  openModal.value = true
}
const formRef = ref(null)
const openDeleteModal = ref(false)
const openDropStudentModal = ref(false)

function handleDelete() {
  // console.log('handleDelete')
  openDeleteModal.value = true
}

async function handleSubmitDelete() {
  btnLoading.value = true
  try {
    // 直接调用删除API
    const res = await batchDeleteIntendedStudentApi({ studentIds: [studentId.value] })
    if (res.code === 200) {
      messageService.success('删除成功')
      // 发送刷新列表事件
      emitter.emit(EVENTS.REFRESH_STUDENT_LIST)
      // 成功后再关闭弹窗和清空表单
      openDeleteModal.value = false
      openDrawer.value = false
    }
    btnLoading.value = false
  }
  catch (error) {
    console.log(error)
    messageService.error('删除失败')
    btnLoading.value = false
  }
}

const studentStore = useStudentStore()
const userStore = useUserStore()
const { systemDefaultIsDisplayList, getAllStuFields } = useStudentFields()
const studentId = computed(() => studentStore.studentId)
const studentDetail = ref({})
const oneToOneClassTeacherNames = ref([])
const router = useRouter()
const hasFaceCollected = computed(() => !!studentDetail.value?.isCollect)

// 判断是否是自己的学员
const isMyStudent = computed(() => {
  if (!studentDetail.value || !userStore.userInfo?.instUserId) {
    return false
  }
  // salePerson 是字符串，instUserId 是数字，需要转换类型后比较
  return String(studentDetail.value.salePerson) === String(userStore.userInfo.instUserId)
})

// 处理报名按钮点击
function handleEnrollment() {
  if (studentId.value) {
    // 跳转到报名续费页面（通过 query 传学员 ID，避免被当成订单 ID 查询）
    router.push({
      path: '/edu-center/registr-renewal',
      query: {
        id: String(studentId.value),
      },
    })
  } else {
    messageService.error('学生信息不完整，无法跳转')
  }
}
// 获取意向学员详情
// getIntentStudentDetailApi
async function getIntentStudentDetail(flag = true) {
  const currentStudentId = String(studentId.value || '').trim()
  if (!currentStudentId) {
    studentDetail.value = {}
    return
  }
  try {
    const res = await getIntentStudentDetailApi({ studentId: currentStudentId })
    if (res.code === 200) {
      if (String(studentId.value || '').trim() !== currentStudentId)
        return
      studentDetail.value = res.result
      nextTick(() => {
        if (activeKey.value === '4' && flag) {
          followUpRecordRef.value.getFollowUpRecord(true)
        }
      })
    }
  }
  catch (error) {
    console.log(error)
  }
}

async function getOneToOneClassTeachers() {
  if (!studentId.value) {
    oneToOneClassTeacherNames.value = []
    return
  }
  try {
    const res = await getOneToOneListApi({
      pageRequestModel: {
        needTotal: true,
        pageSize: 100,
        pageIndex: 1,
        skipCount: 0,
      },
      queryModel: {
        studentId: String(studentId.value),
      },
    })
    if (res.code !== 200) {
      oneToOneClassTeacherNames.value = []
      return
    }
    const names = (Array.isArray(res.result?.list) ? res.result.list : [])
      .flatMap((item) => {
        const raw = String(item?.classTeacherName || '').trim()
        if (!raw)
          return []
        return raw.split('、').map(name => name.trim()).filter(Boolean)
      })
    oneToOneClassTeacherNames.value = Array.from(new Set(names))
  }
  catch (error) {
    console.log(error)
    oneToOneClassTeacherNames.value = []
  }
}

function isSystemFieldVisible(fieldKey) {
  return systemDefaultIsDisplayList.value.some(item => item.fieldKey === fieldKey)
}

function refreshPrimaryCourseList() {
  nextTick(() => {
    registerForCoursesRef.value?.getTuitionAccountList()
  })
}

watch(() => props.open, (newVal) => {
  if (!newVal) {
    studentStore.clearStudentId()
    oneToOneClassTeacherNames.value = []
    // 关闭抽屉时重置activeKey为默认值
    activeKey.value = props.defaultActiveKey
    registerForCoursesRef.value?.resetListState?.()
    // 关闭抽屉时刷新学员列表，确保数据准确性
    emitter.emit(EVENTS.REFRESH_STUDENT_LIST)
    // 清理下拉框状态
    openIntentionDropdown.value = false
    openStatusDropdown.value = false
    // 重置手机号解密状态
    isPhoneDecrypted.value = false
    decryptedPhone.value = ''
    phoneLoading.value = false
  }
})

// 监听openDrawer的变化
watch(() => openDrawer.value, (newVal) => {
  if (newVal) {
    getIntentStudentDetail()
    getOneToOneClassTeachers()
    getAllStuFields({ filter: 3 })
    if (activeKey.value === '0') {
      refreshPrimaryCourseList()
    }
  }
})

watch(() => studentId.value, (newVal, oldVal) => {
  if (!openDrawer.value || newVal === oldVal)
    return
  getIntentStudentDetail()
  getOneToOneClassTeachers()
})

// 监听 activeKey 变化，切换到报读课程时刷新数据
watch(() => activeKey.value, (newVal) => {
  if (newVal === '0' && openDrawer.value) {
    refreshPrimaryCourseList()
  }
})

const btnLoading = ref(false)

const { width } = useWindowSize()

// 响应式抽屉宽度
const drawerWidth = computed(() => {
  if (width.value < 990) {
    return '100%'
  }
  else if (width.value < 1165) {
    return '80%'
  }
  else {
    return '1165px'
  }
})
// 响应式描述列表列数
const responsiveColumns = computed(() => {
  if (width.value < 550) {
    return 2
  }
  else if (width.value < 800) {
    return 3
  }
  else if (width.value < 1024) {
    return 4
  }
  else {
    return 4
  }
})
function handleEdit() {
  // openDrawer.value = false
  openEditStudentModal.value = true
  // emitter.emit(EVENTS.REFRESH_STUDENT_LIST)
}
const followUpRecordRef = ref(null)
const createStudentRef = ref(null)
const registerForCoursesRef = ref(null)
const openEditStudentModal = ref(false)
const assignSalesVisible = ref(false)
const assignSalesRef = ref(null)
async function handleUpdateStudent(data) {
  try {
    // 创建一个新对象用于API调用，避免修改原始数据
    const apiData = { ...data }

    // 处理渠道ID
    if (apiData.channelId && apiData.channelId.length === 1) {
      apiData.channelId = apiData.channelId[0]
    }
    else if (apiData.channelId && apiData.channelId.length > 1) {
      apiData.channelId = apiData.channelId[1]
    }
    const res = await updateIntendedStudentApi(apiData)

    if (res.code === 200) {
      messageService.success('编辑成功')
      openEditStudentModal.value = false
      // emitter.emit(EVENTS.REFRESH_STUDENT_LIST)
      getIntentStudentDetail()
      createStudentRef.value.resetForm()
      isPhoneDecrypted.value = false
      decryptedPhone.value = ''
      phoneLoading.value = false
    }
  }
  catch (error) {
    console.log('error: ', error)
  }
  finally {
    createStudentRef.value.closeSpinning()
  }
}

// 报读课程数量（根据子组件暴露的 tuitionAccountList 实时计算）
const registerCourseCount = computed(() => {
  const detailCount = Number(studentDetail.value?.primaryCourseCount || 0)
  const listLoaded = registerForCoursesRef.value?.listLoaded
  const exposed = registerForCoursesRef.value?.tuitionAccountList
  const hasLoadedList = listLoaded === true || listLoaded?.value === true
  if (!hasLoadedList) {
    return detailCount
  }

  // 情况 1：子组件暴露的是响应式 ref([])，这里拿到的是 ref
  // 情况 2：子组件暴露的 ref 在父组件已被自动解包，这里拿到的是普通数组
  if (Array.isArray(exposed)) {
    return exposed.length
  }
  if (Array.isArray(exposed.value)) {
    return exposed.value.length
  }

  return detailCount
})

const isIntentionStudent = computed(() => Number(studentDetail.value?.studentStatus) === StudentStatus.Intention)
const primaryCourseTabLabel = computed(() => (isIntentionStudent.value ? '体验课程' : '报读课程'))
const primaryCourseEmptyText = computed(() => (isIntentionStudent.value ? '没有体验课程' : '没有报读课程'))
const studentStatusText = computed(() => StudentStatusLabel[Number(studentDetail.value?.studentStatus)] || '-')
const classTeacherNames = computed(() => {
  if (oneToOneClassTeacherNames.value.length)
    return oneToOneClassTeacherNames.value
  const fallback = String(studentDetail.value?.teacherName || '').trim()
  return fallback ? [fallback] : []
})
const classTeacherSummary = computed(() => (
  classTeacherNames.value.length ? `${classTeacherNames.value.length}位` : '-'
))
const classTeacherTooltipTitle = computed(() => classTeacherNames.value.join('、'))

// 处理分配销售
function handleAssignSales() {
  assignSalesVisible.value = true
}

// 处理分配销售提交
async function handleAssignSalesSubmit(data) {
  try {
    const res = await batchAssignSalespersonApi(data)
    if (res.code === 200) {
      messageService.success('分配销售成功')
      // 刷新学员详情
      getIntentStudentDetail()
      // 发送刷新列表事件
      // emitter.emit(EVENTS.REFRESH_STUDENT_LIST)
      // 关闭弹窗
      assignSalesVisible.value = false
    }
  }
  catch (error) {
    console.log(error)
    messageService.error('分配销售失败')
  }
  finally {
    // 关闭modal中的loading
    assignSalesRef.value?.closeLoading()
  }
}

// 处理放弃学员（打开确认弹窗）
function handleDropStudent() {
  openDropStudentModal.value = true
}

// 确认放弃学员
async function handleConfirmDropStudent() {
  if (!studentId.value) {
    messageService.error('学员信息不完整')
    return
  }
  btnLoading.value = true
  try {
    const res = await batchTransferToPublicPoolApi({ studentIds: [studentId.value] })
    if (res.code === 200) {
      messageService.success('放弃成功')
      // 关闭弹窗
      openDropStudentModal.value = false
      // 刷新学员详情
      getIntentStudentDetail()
      // 发送刷新列表事件
      emitter.emit(EVENTS.REFRESH_STUDENT_LIST)
    }
  } catch (error) {
    console.log(error)
    messageService.error('放弃失败')
  } finally {
    btnLoading.value = false
  }
}

// 监听 openModal 的变化
watch(() => openModal.value, (newVal) => {
  if (newVal) {
    getStudentChangeInfo()
  }
})

// 提供给子组件的刷新回调函数
function handleRefreshStudentDetail() {
  getIntentStudentDetail()
  getOneToOneClassTeachers()
}

const studentChangeInfo = ref([])
const changeInfoLoading = ref(false)
async function getStudentChangeInfo() {
  changeInfoLoading.value = true
  try {
    const res = await listStudentChangeInfoApi({ stuId: studentId.value })
    if (res.code === 200) {
      studentChangeInfo.value = res.result
    }
  }
  catch (error) {
    console.log(error)
  }
  finally {
    changeInfoLoading.value = false
  }
}

// 下拉框状态管理
const openIntentionDropdown = ref(false)
const openStatusDropdown = ref(false)

// 手机号解密相关状态
const isPhoneDecrypted = ref(false)
const decryptedPhone = ref('')
const phoneLoading = ref(false)

const DEFAULT_AVATAR_URL = 'https://pcsys.admin.ybc365.com/a369a751-2be5-4929-974d-9ae4439f54c4.png'

function maskPhoneNumber(value) {
  if (!value) {
    return ''
  }
  if (value.includes('*')) {
    return value
  }
  if (value.length >= 11) {
    return value.replace(/(\d{3})\d{4}(\d{4})/, '$1****$2')
  }
  return value
}

function formatBirthDayForDisplay(value) {
  if (!value) {
    return ''
  }
  const normalized = value.split('T')[0]
  if (normalized === '0001-01-01') {
    return ''
  }
  const parsed = dayjs(normalized)
  if (!parsed.isValid() || parsed.year() <= 1) {
    return ''
  }
  return parsed.format('YYYY-MM-DD')
}

const studentAvatarSrc = computed(() => {
  return studentDetail.value?.avatarUrl || studentDetail.value?.avatar || DEFAULT_AVATAR_URL
})

const hasFollowedOfficialAccount = computed(() => !!studentDetail.value?.isBindChild)

const studentGender = computed(() => {
  return (studentDetail.value?.stuSex ?? studentDetail.value?.sex) ?? Sex.Unknown
})

const studentNameDisplay = computed(() => {
  return (
    studentDetail.value?.stuName
    || studentDetail.value?.studentName
    || studentDetail.value?.name
    || '-'
  )
})

const studentBirthDayRaw = computed(() => {
  return studentDetail.value?.birthDay ?? studentDetail.value?.birthday ?? ''
})

const formattedBirthDay = computed(() => formatBirthDayForDisplay(studentBirthDayRaw.value))

const displayAge = computed(() => (formattedBirthDay.value ? calculateAge(formattedBirthDay.value) : '-'))

const studentPhoneRaw = computed(() => {
  return (
    studentDetail.value?.mobile
    || studentDetail.value?.phone
    || studentDetail.value?.phoneNumber
    || ''
  )
})

const studentMaskedPhone = computed(() => maskPhoneNumber(studentPhoneRaw.value))

const studentPhoneRelationship = computed(() => {
  return (
    studentDetail.value?.phoneRelationship
    ?? studentDetail.value?.phoneRelation
    ?? studentDetail.value?.relationship
    ?? 0
  )
})

const displayPhoneRelationship = computed(() => {
  return ParentRelationshipLabel[studentPhoneRelationship.value] || '-'
})

const displayPhoneNumber = computed(() => {
  if (isPhoneDecrypted.value) {
    return decryptedPhone.value || '-'
  }
  return studentMaskedPhone.value || '-'
})

const rechargeAccountBalanceTotal = computed(() => {
  return Number(studentDetail.value?.rechargeAccountBalanceTotal || 0)
})

function formatMoney(value) {
  return Number(value || 0).toFixed(2)
}

// 切换意向度下拉框
function toggleIntentionDropdown() {
  openStatusDropdown.value = false
  openIntentionDropdown.value = !openIntentionDropdown.value
}

// 切换跟进状态下拉框
function toggleStatusDropdown() {
  openIntentionDropdown.value = false
  openStatusDropdown.value = !openStatusDropdown.value
}

// 处理下拉框显示状态变化
function handleDropdownVisibleChange(visible, type) {
  if (!visible) {
    if (type === 'intention') {
      openIntentionDropdown.value = false
    } else {
      openStatusDropdown.value = false
    }
  }
}

// 获取状态配置的计算属性
const getStatusConfig = computed(() => (status) => {
  return {
    text: FollowUpStatusLabel[status] || '-',
    className: FollowUpStatusStyle[status]?.className || '',
  }
})

// 使用防抖包装状态变更处理函数
const debouncedStatusChange = debounce(async (status) => {
  const previousStatus = studentDetail.value.followUpStatus
  studentDetail.value = {
    ...studentDetail.value,
    followUpStatus: status,
  }
  openStatusDropdown.value = false
  btnLoading.value = true
  try {
    const res = await updateStatusApi({
      id: studentDetail.value.id,
      uuid: studentDetail.value.uuid,
      version: studentDetail.value.version,
      followUpStatus: status,
    })
    if (res.code === 200) {
      messageService.success('更新状态成功')
      return
    }
    throw new Error(res.message || '更新状态失败')
  } catch (error) {
    studentDetail.value = {
      ...studentDetail.value,
      followUpStatus: previousStatus,
    }
    openStatusDropdown.value = false
    messageService.error('更新状态失败')
    console.error('更新状态失败:', error)
  } finally {
    btnLoading.value = false
  }
}, 300, { leading: true, trailing: false })

// 处理跟进状态修改
async function handleStatusChange(status) {
  if (studentDetail.value.followUpStatus === status) {
    openStatusDropdown.value = false
    return
  }
  debouncedStatusChange(status)
}

// 使用防抖包装意向度变更处理函数
const debouncedIntentionChange = debounce(async (intentLevel) => {
  const previousIntentLevel = studentDetail.value.intentLevel
  studentDetail.value = {
    ...studentDetail.value,
    intentLevel,
  }
  openIntentionDropdown.value = false
  btnLoading.value = true
  try {
    const res = await updateStatusApi({
      id: studentDetail.value.id,
      uuid: studentDetail.value.uuid,
      version: studentDetail.value.version,
      intentLevel,
    })
    if (res.code === 200) {
      messageService.success('更新意向度成功')
      return
    }
    throw new Error(res.message || '更新意向度失败')
  } catch (error) {
    studentDetail.value = {
      ...studentDetail.value,
      intentLevel: previousIntentLevel,
    }
    openIntentionDropdown.value = false
    messageService.error('更新意向度失败')
    console.error('更新意向度失败:', error)
  } finally {
    btnLoading.value = false
  }
}, 300, { leading: true, trailing: false })

// 处理意向度修改
async function handleIntentionChange(intentLevel) {
  if (studentDetail.value.intentLevel === intentLevel) {
    openIntentionDropdown.value = false
    return
  }
  debouncedIntentionChange(intentLevel)
}

// 处理手机号解密/隐藏切换
async function handlePhoneToggle() {
  if (isPhoneDecrypted.value) {
    // 如果已解密，切换为隐藏状态
    isPhoneDecrypted.value = false
    return
  }
  
  // 如果已经解密过，直接显示
  if (decryptedPhone.value) {
    isPhoneDecrypted.value = true
    return
  }
  
  // 首次解密，调用接口
  phoneLoading.value = true
  try {
    const res = await getStudentPhoneNumberApi({ studentId: studentId.value })
    if (res.code === 200) {
      decryptedPhone.value = res.result
      isPhoneDecrypted.value = true
    }
  } catch (error) {
    console.error('解密手机号失败:', error)
    messageService.error('解密手机号失败')
  } finally {
    phoneLoading.value = false
  }
}
</script>

<template>
  <div>
    <a-drawer v-model:open="openDrawer" :push="false" :body-style="{ padding: '0', background: '#f7f7fd' }"
      :closable="false" :width="drawerWidth" placement="right">
      <!-- 自定义头部 -->
      <template #title>
        <div class="custom-header flex justify-between h-4 flex-items-center">
          <div class="text-5">
            学员详情
          </div>
          <a-button type="text" class="close-btn" @click="openDrawer = false">
            <template #icon>
              <CloseOutlined class="text-5 close-icon" />
            </template>
          </a-button>
        </div>
      </template>
      <div class="contenter flex flex-col lg:flex-row flex-center bg-white px6 py3">
        <div class="avatarBox w-16 h-16 relative mx-auto lg:mx-0">
          <img width="64" height="64" class="rounded-100" :src="studentAvatarSrc" alt="">
          <img v-if="studentGender === Sex.Female" width="24" height="24" class="absolute bottom-0 right-0"
            src="~@/assets/images/girl.png" alt="">
          <img v-else-if="studentGender === Sex.Male" width="24" height="24" class="absolute bottom-0 right-0"
            src="~@/assets/images/boy.png" alt="">
          <img v-else width="24" height="24" class="absolute bottom-0 right-0" src="~@/assets/images/unknown.svg"
            alt="">
        </div>
        <div class="info flex flex-1 lg:ml-4 flex-col mt-4 lg:mt-0">
          <div class="top flex flex-col lg:flex-row justify-between flex-center flex-1">
            <a-space class="mb-4 lg:mb-0">
              <div class="name text-5 font-800 whitespace-nowrap">
                {{ studentNameDisplay }}
              </div>
              <a-tooltip :title="hasFollowedOfficialAccount ? '已关注' : '未关注'">
                <img
                  class="mt-0.5 cursor-pointer ml1 follow-status-icon"
                  :class="{ 'follow-status-icon--inactive': !hasFollowedOfficialAccount }"
                  src="~@/assets/images/follow.svg"
                  alt=""
                >
              </a-tooltip>
              <a-tooltip v-if="hasFaceCollected" title="已采集">
                <span class="face-collect-status ml0.5">
                  <svg class="face-collect-status__svg" width="16px" height="16px" viewBox="0 0 16 16" xmlns="http://www.w3.org/2000/svg">
                    <g stroke="none" stroke-width="1" fill="none" fill-rule="evenodd">
                      <g transform="translate(-594.000000, -464.000000)">
                        <g transform="translate(518.000000, 310.000000)">
                          <g transform="translate(0.000000, 126.000000)">
                            <g transform="translate(76.000000, 21.600000)">
                              <g transform="translate(0.000000, 6.400000)">
                                <polygon fill="#000000" fill-rule="nonzero" opacity="0" points="0 0 16 0 16 16 8 16 0 16" />
                                <path
                                  d="M1.49983336,11 C1.74529324,10.9999182 1.94950067,11.1767253 1.99191437,11.4099604 L2,11.4998334 L2,14 L4.5,14 C4.74545992,14 4.9496084,14.1768752 4.99194436,14.4101244 L5,14.5 C5,14.7454599 4.82312487,14.9496084 4.58987566,14.9919444 L4.5,15 L1.50100003,15 C1.25559799,15 1.05147725,14.8232051 1.00908211,14.5900195 L1.00100006,14.5001667 L1,11.5001667 C0.999908009,11.2240243 1.223691,11.0000921 1.49983336,11 Z M14.4988336,11 C14.7442935,10.9999183 14.9485009,11.1767254 14.9909146,11.4099605 L14.9990002,11.4998334 L15,14.4998334 C15.0000818,14.7453511 14.8231944,14.9495863 14.5898958,14.9919408 L14.5,15 L11.5,15 C11.2238576,15 11,14.7761424 11,14.5 C11,14.2545401 11.1768752,14.0503917 11.4101244,14.0080557 L11.5,14 L14,14 L13.9990003,11.5001667 C13.9989185,11.2547068 14.1757256,11.0504994 14.4089607,11.0080857 L14.4988336,11 Z M4.5,9 L11.5,9 L11.4931641,9.38828125 L11.4769287,9.60498047 L11.4453125,9.83125 C11.28125,10.75 10.625,11.8 8,11.8 C5.484375,11.8 4.77685547,10.8356771 4.5778656,9.94669189 L4.53663635,9.71717529 C4.53140259,9.67943522 4.5269165,9.64200846 4.52307129,9.60498047 L4.50683594,9.38828125 L4.5,9 Z M11,5.5 C11.5522847,5.5 12,5.94771525 12,6.5 C12,7.05228475 11.5522847,7.5 11,7.5 C10.4477153,7.5 10,7.05228475 10,6.5 C10,5.94771525 10.4477153,5.5 11,5.5 Z M5,5.5 C5.55228475,5.5 6,5.94771525 6,6.5 C6,7.05228475 5.55228475,7.5 5,7.5 C4.44771525,7.5 4,7.05228475 4,6.5 C4,5.94771525 4.44771525,5.5 5,5.5 Z M14.5,1 C14.7455177,1 14.9496939,1.17695541 14.9919707,1.41026814 L15,1.50016663 L14.9990002,4.50016663 C14.9989082,4.77630898 14.774976,5.000092 14.4988336,5 C14.2533737,4.99991817 14.0492842,4.82297499 14.007026,4.58971169 L13.9990003,4.49983337 L14,2 L11.5,2 C11.2545401,2 11.0503916,1.82312484 11.0080557,1.58987563 L11,1.5 C11,1.25454011 11.1768752,1.05039163 11.4101244,1.00805567 L11.5,1 L14.5,1 Z M4.5,1 C4.77614235,1 5,1.22385763 5,1.5 C5,1.74545989 4.82312481,1.94960837 4.5898756,1.99194433 L4.5,2 L2,2 L2,4.50016667 C1.99991812,4.74562654 1.82297492,4.94971605 1.58971162,4.99197426 L1.49983331,5 C1.25437343,4.99991815 1.05028392,4.82297495 1.00802571,4.58971165 L1,4.49983333 L1.001,1.49983333 C1.0010818,1.25443131 1.17794474,1.05036951 1.41114451,1.0080521 L1.50099997,1 L4.5,1 Z"
                                  fill="#0066FF"
                                />
                              </g>
                            </g>
                          </g>
                        </g>
                      </g>
                    </g>
                  </svg>
                </span>
              </a-tooltip>
              <a-tooltip v-else title="未采集">
                <img class="mt-0.5 cursor-pointer ml0.5" src="~@/assets/images/face.svg" alt="">
              </a-tooltip>
            </a-space>
            <a-space class="flex flex-wrap lg:flex-nowrap justify-center lg:justify-start gap-2">
              <a-button @click="handleDelete">
                删除
              </a-button>
              <a-button @click="handleEdit">
                编辑
              </a-button>
              <a-button>未排课点名</a-button>
              <a-button @click="handleAssignSales">分配销售</a-button>
              <a-tooltip v-if="isMyStudent" title="放弃该学员">
                <a-button @click="handleDropStudent" :loading="btnLoading">
                  放弃
                </a-button>
              </a-tooltip>
              <a-button>试听</a-button>
              <a-button type="primary" @click="handleEnrollment">
                报名
              </a-button>
            </a-space>
          </div>
          <div
            class="bottom  flex-1 flex flex-wrap justify-center lg:justify-start whitespace-nowrap flex-items-center mt-2">
            <div class="birthday flex-center lg:mb-0">
              <img src="~@/assets/images/birthday.svg" alt="">
              <span class="text-15px text-#000000e0 ml-2 ">
                {{ displayAge }}
                <span v-if="formattedBirthDay">({{ formattedBirthDay }})</span>
              </span>
            </div>
            <span class="hidden lg:flex w-0.25 h-3.5 bg-#ccc mx-2.5" />
            <div class="acount flex-center lg:mb-0 mx-12px">
              <img src="~@/assets/images/bag.svg" alt="">
              <span class="price mx-1 text-15px text-#000000e0">{{ formatMoney(rechargeAccountBalanceTotal) }}</span>
              <span class="text-#06f cursor-pointer text-3.5">储值
                <RightOutlined class="text-3" />
              </span>
            </div>
            <span class="hidden lg:flex w-0.25 h-3.5 bg-#ccc mr-2.5 mx-2.5" />
            <div class="phone lg:mb-0 flex-center">
              <span class="text-15px text-#000000e0">{{ displayPhoneNumber }}<span class="text-14px">({{
                displayPhoneRelationship }})</span></span>
              <a-spin :spinning="phoneLoading" size="small">
                <EyeOutlined v-if="isPhoneDecrypted" class="cursor-pointer text-#06f ml-2" @click="handlePhoneToggle" />
                <EyeInvisibleOutlined v-else class="cursor-pointer text-#06f ml-2" @click="handlePhoneToggle" />
              </a-spin>
            </div>
            <span class="hidden lg:flex w-0.25 h-3.5 bg-#ccc mr-2.5 mx-2.5" />
            <div class="flex items-center ml-10px">
              <img width="18" height="18" src="https://pcsys.admin.ybc365.com//06365255-087c-423c-bf37-33a4ceab650b.png"
                alt="">
              <span class="ml-1 text-15px text-#000000e0 ">9999</span>
            </div>
          </div>
        </div>
      </div>
      <div class="desc pt-4 bg-white px6 py3 pb0">
        <a-descriptions :column="responsiveColumns" size="small" :content-style="{ color: '#888' }">
          <a-descriptions-item label="销售员">
            {{ studentDetail.salePersonName || '-' }}
          </a-descriptions-item>
          <a-descriptions-item label="班主任">
            <a-tooltip v-if="classTeacherNames.length" :title="classTeacherTooltipTitle" placement="topLeft">
              <span class="cursor-pointer">{{ classTeacherSummary }}</span>
            </a-tooltip>
            <span v-else>-</span>
          </a-descriptions-item>
          <a-descriptions-item label="意向度">
            <div style="cursor: pointer;">
              <a-dropdown :trigger="['click']" :open="openIntentionDropdown"
                @update:open="(val) => handleDropdownVisibleChange(val, 'intention')">
                <div @click.prevent="toggleIntentionDropdown">
                  <div class="intention">
                    <span class="intentionTag"
                      :style="{ background: IntentionLevelStyle[studentDetail?.intentLevel]?.color }" />
                    {{ IntentionLevelLabel[studentDetail?.intentLevel] || IntentionLevelLabel[IntentionLevel.Unknown] }}
                    <CaretDownOutlined v-if="!openIntentionDropdown" class="ml-1 text-#ccc"
                      :style="{ 'font-size': '10px' }" />
                    <CaretUpOutlined v-if="openIntentionDropdown" class="ml-1 text-#1677ff"
                      :style="{ 'font-size': '10px' }" />
                  </div>
                </div>
                <template #overlay>
                  <a-menu style="text-align: center;width: 70px;"
                    @click="({ key }) => handleIntentionChange(Number(key))">
                    <a-menu-item v-for="level in Object.values(IntentionLevel).filter(v => !isNaN(Number(v)))"
                      :key="level">
                      {{ IntentionLevelLabel[level] }}
                    </a-menu-item>
                  </a-menu>
                </template>
              </a-dropdown>
            </div>
          </a-descriptions-item>
          <a-descriptions-item label="来源渠道">
            {{ studentDetail.channelName || '-' }}
          </a-descriptions-item>
          <a-descriptions-item label="跟进状态">
            <div style="cursor: pointer;">
              <a-dropdown :trigger="['click']" :open="openStatusDropdown"
                @update:open="(val) => handleDropdownVisibleChange(val, 'status')">
                <div @click.prevent="toggleStatusDropdown">
                  <div class="intention">
                    <span class="statusTag" :class="getStatusConfig(studentDetail?.followUpStatus).className">
                      {{ getStatusConfig(studentDetail?.followUpStatus).text }}
                    </span>
                    <component :is="openStatusDropdown ? CaretUpOutlined : CaretDownOutlined"
                      class="ml-1px" :class="openStatusDropdown ? 'text-#1677ff' : 'text-#ccc'"
                      :style="{ fontSize: '10px' }" />
                  </div>
                </div>
                <template #overlay>
                  <a-menu style="text-align: center;width: 100px;"
                    @click="({ key }) => handleStatusChange(Number(key))">
                    <a-menu-item v-for="status in Object.values(FollowUpStatus).filter(v => !isNaN(Number(v)))"
                      :key="status">
                      {{ FollowUpStatusLabel[status] }}
                    </a-menu-item>
                  </a-menu>
                </template>
              </a-dropdown>
            </div>
          </a-descriptions-item>
          <a-descriptions-item v-if="isSystemFieldVisible('微信号')" label="微信号">
            {{ studentDetail.weChatNumber || '-' }}
          </a-descriptions-item>
          <a-descriptions-item v-if="isSystemFieldVisible('年级')" label="年级">
            {{ studentDetail.grade || '-' }}
          </a-descriptions-item>
          <a-descriptions-item>
            <span class="text-#06f cursor-pointer" @click="seeAllData">查看全部资料</span>
          </a-descriptions-item>
        </a-descriptions>
      </div>
      <div class="tabs">
        <a-tabs v-model:active-key="activeKey" size="large" :tab-bar-style="{
          'border-radius': '0px', 'padding-left': '24px'
        }">
          <a-tab-pane
            key="0"
            :tab="`${primaryCourseTabLabel}（${registerCourseCount}）`"
          >
            <RegisterForCourses ref="registerForCoursesRef" :empty-text="primaryCourseEmptyText" />
          </a-tab-pane>
          <a-tab-pane key="1" tab="上课记录">
            <StudentClassRecord />
          </a-tab-pane>
          <a-tab-pane key="2" tab="订单记录">
            <OrderRecord />
          </a-tab-pane>
          <a-tab-pane key="3" tab="试听记录">
            <TryListeningRecord />
          </a-tab-pane>
          <a-tab-pane key="4" tab="跟进记录">
            <FollowUpRecord ref="followUpRecordRef" :student-id="studentId" :student-detail="studentDetail" 
              @refresh-student-detail="handleRefreshStudentDetail" />
          </a-tab-pane>
          <a-tab-pane key="5" tab="康复档案">
            3
          </a-tab-pane>

          <a-tab-pane key="6" tab="评估报告">
            3
          </a-tab-pane>
          <a-tab-pane key="7" tab="康复记录">
            3
          </a-tab-pane>
          <a-tab-pane key="8" tab="作业记录">
            3
          </a-tab-pane>
          <a-tab-pane key="9" tab="交互记录">
            3
          </a-tab-pane>
        </a-tabs>
      </div>
    </a-drawer>
    <!-- 查看全部资料 modal 弹窗 -->
    <a-modal v-model:open="openModal" title="学员资料" :width="800"
      :body-style="{ padding: '10px 24px', height: '76vh', overflow: 'auto' }" style="top:50px;">
      <a-descriptions :column="4" size="small" :content-style="{ color: '#888' }">
        <a-descriptions-item label="销售员">
          {{ studentDetail.salePersonName || '-' }}
        </a-descriptions-item>
        <a-descriptions-item label="班主任">
          <a-tooltip v-if="classTeacherNames.length" :title="classTeacherTooltipTitle" placement="topLeft">
            <span class="cursor-pointer">{{ classTeacherSummary }}</span>
          </a-tooltip>
          <span v-else>-</span>
        </a-descriptions-item>
      </a-descriptions>
      <a-divider class="my-5px" />
      <div class="avatar my-15px">
        <img :src="studentAvatarSrc" alt="" width="80" height="80" class="rounded-10">
      </div>
      <a-descriptions title="学员信息" :column="2" size="small" :content-style="{ color: '#888' }" :label-style="{ whiteSpace: 'nowrap' }">
        <a-descriptions-item label="学员姓名">
          {{ studentNameDisplay }}
        </a-descriptions-item>
        <a-descriptions-item label="学员性别">
          {{ SexLabel[studentGender] || SexLabel[Sex.Unknown] }}
        </a-descriptions-item>
        <a-descriptions-item label="手机号码">
          <span>{{ displayPhoneNumber }}（{{ displayPhoneRelationship }}）</span>
          <a-spin :spinning="phoneLoading" size="small" class="ml-2">
            <EyeOutlined v-if="isPhoneDecrypted" class="cursor-pointer text-#06f" @click="handlePhoneToggle" />
            <EyeInvisibleOutlined v-else class="cursor-pointer text-#06f" @click="handlePhoneToggle" />
          </a-spin>
        </a-descriptions-item>
        <a-descriptions-item label="学员状态">
          {{ studentStatusText }}
        </a-descriptions-item>
        <a-descriptions-item v-if="isSystemFieldVisible('微信号')" label="微信号">
          {{ studentDetail.weChatNumber || '-' }}
        </a-descriptions-item>
        <a-descriptions-item v-if="isSystemFieldVisible('家庭住址')" label="家庭住址">
          <a-tooltip v-if="studentDetail.address" :title="studentDetail.address" placement="topLeft">
            <div class="address-text">
              {{ studentDetail.address }}
            </div>
          </a-tooltip>
          <span v-else>-</span>
        </a-descriptions-item>
        <a-descriptions-item label="渠道分类">
          {{ studentDetail.channelCategoryName || '-' }}
        </a-descriptions-item>
        <a-descriptions-item label="来源渠道">
          {{ studentDetail.channelName || '-' }}
        </a-descriptions-item>
        <a-descriptions-item label="学员备注" :span="2">
          {{ studentDetail.remark || '-' }}
        </a-descriptions-item>
        <!-- 动态渲染自定义字段 -->
        <a-descriptions-item v-for="customField in studentDetail.customInfo || []" :key="customField.fieldId"
          :label="customField.fieldName">
          {{ customField.value || '-' }}
        </a-descriptions-item>
        <a-descriptions-item label="创建时间">
          {{ studentDetail.createTime ? dayjs(studentDetail.createTime).format('YYYY-MM-DD HH:mm') : '-' }}
        </a-descriptions-item>
        <a-descriptions-item label="创建人">
          {{ studentDetail.createName || '-' }}
        </a-descriptions-item>
        <a-descriptions-item label="分配销售时间">
          {{ studentDetail.salesAssignedTime ? dayjs(studentDetail.salesAssignedTime).format('YYYY-MM-DD HH:mm') : '-'
          }}
        </a-descriptions-item>
        <a-descriptions-item label="体验课购买状态">
          未购买
        </a-descriptions-item>
        <a-descriptions-item label="首次报读时间">
          {{ studentDetail?.firstEnrolledTime ? dayjs(studentDetail.firstEnrolledTime).format('YYYY-MM-DD HH:mm') : '-'
          }}
        </a-descriptions-item>
        <a-descriptions-item label="最后转为历史学员时间">
          {{ studentDetail.turnedHistoryTime ? dayjs(studentDetail.turnedHistoryTime).format('YYYY-MM-DD HH:mm') : '-'
          }}
        </a-descriptions-item>
      </a-descriptions>
      <div class="border border-solid border-color-#eee rounded-6px mt-12px">
        <div
          class="header bg-#fafafa h-40px flex-items-center flex pl-24px text-14px text-#333 font500 border-b border-t-0 border-x-0 border-solid border-color-#eee">
          信息变更记录
        </div>
        <a-spin :spinning="changeInfoLoading">
          <div v-if="changeInfoLoading" class="min-h-100px flex-center">
            <!-- loading时的占位内容 -->
          </div>
          <div v-else>
            <div v-for="item in studentChangeInfo" :key="item.id">
              <div class="px-24px py-12px">
                <div class="text-14px text-#666">
                  <span>{{ dayjs(item.createTime).format('YYYY-MM-DD HH:mm') }}</span> <span>{{ item.changeName }}</span>
                </div>
                <div class="text-12px text-#888 mt-2px">
                  变更内容：{{ item.changeContent }}
                </div>
              </div>
              <a-divider v-if="item.id !== studentChangeInfo[studentChangeInfo.length - 1].id" class="my-0px" />
            </div>
            <div v-if="studentChangeInfo.length === 0">
              <a-empty description="暂无变更记录" :image="Empty.PRESENTED_IMAGE_SIMPLE" />
            </div>
          </div>
        </a-spin>
      </div>
      <template #footer>
        <a-button @click="openModal = false">关闭</a-button>
      </template>
    </a-modal>
    <!-- 确认删除学员弹窗 -->
    <DeleteConfirmModal v-model:open="openDeleteModal" :student-names="studentDetail.stuName || ''"
      :loading="btnLoading" @confirm="handleSubmitDelete" @cancel="openDeleteModal = false" />
    <!-- 编辑学生 -->
    <CreateStudent ref="createStudentRef" v-model:open="openEditStudentModal" :record="studentDetail" :type="2"
      @submit="handleUpdateStudent" />

    <!-- 分配销售弹窗 -->
    <AssignSalesModal ref="assignSalesRef" v-model:open="assignSalesVisible" title="分配销售" :type="2"
      :selected-students="[studentDetail]" @submit="handleAssignSalesSubmit" />

    <!-- 确认放弃学员弹窗 -->
    <a-modal
      v-model:open="openDropStudentModal"
      centered
      :footer="false"
      :closable="false"
      :mask-closable="false"
      :keyboard="false"
      width="480px"
    >
      <div class="text-18px mb-12px font500 flex items-center">
        <ExclamationCircleOutlined class="text-#ff4d4f mr-2 text-5" />
        确定放弃该学员?
      </div>
      <div class="pl-30px text-#666 mb-24px">
        放弃后,您就不再是该学员的销售员,此学员将进入公有池
      </div>
      <div class="flex justify-end">
        <a-space>
          <a-button @click="openDropStudentModal = false">
            再想想
          </a-button>
          <a-button type="primary" danger :loading="btnLoading" @click="handleConfirmDropStudent">
            放弃
          </a-button>
        </a-space>
      </div>
    </a-modal>
  </div>
</template>

<style lang="less" scoped>
/* 添加旋转动画 */
@keyframes icon-rotate {
  from {
    transform: rotate(0deg);
  }

  to {
    transform: rotate(180deg);
  }
}

.close-btn {
  &:hover {
    background: transparent;

    .close-icon {
      animation: icon-rotate 0.3s linear;
    }
  }
}

.tabs {
  width: 100%;
  border-radius: 10px;

  :deep(.ant-tabs-nav) {
    background: #fff;
    margin: 0;
  }

  :deep(.ant-tabs-ink-bar) {
    text-align: center;
    height: 12px !important;
    background: transparent;
    bottom: 0px !important;

    &::after {
      position: absolute;
      top: 0;
      left: calc(50% - 12px);
      width: 24px !important;
      height: 4px !important;
      border-radius: 2px;
      background-color: var(--pro-ant-color-primary);
      content: "";
    }
  }

  :deep(.ant-tabs-tab) {
    font-size: 14px;
  }
}

.bottom {
  font-family: Arial, Helvetica, sans-serif;
}

@media (max-width: 800px) {
  .tabs {
    :deep(.ant-tabs-nav) {
      padding-left: 12px;
    }
  }
}

.intention {
  display: flex;
  align-items: center;

  .statusTag {
    width: 54px;
    height: 22px;
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
  border-radius: 100px;
  margin-right: 3px;
}

.address-text {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 300px;
}

.follow-status-icon--inactive {
  filter: grayscale(1) opacity(0.45);
}

.face-collect-status {
  display: inline-flex;
  align-items: center;
  min-height: 24px;
}

.face-collect-status__icon {
  width: 20px;
  height: 20px;
}

.face-collect-status__svg {
  display: block;
}

// 防止 descriptions label 换行
:deep(.ant-descriptions-item-label) {
  white-space: nowrap;
}
</style>
