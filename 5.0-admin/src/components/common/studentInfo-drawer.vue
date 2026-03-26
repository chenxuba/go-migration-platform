<script setup>
import { CloseOutlined, EyeInvisibleOutlined, EyeOutlined, RightOutlined, CaretDownOutlined, CaretUpOutlined, ExclamationCircleOutlined } from '@ant-design/icons-vue'
import { computed, nextTick, ref, watch } from 'vue'
import { useWindowSize } from '@vueuse/core'
import { useRouter } from 'vue-router'
import { debounce } from 'lodash-es'
import DeleteConfirmModal from './DeleteConfirmModal.vue'
import AssignSalesModal from './assign-sales-modal.vue'
import CreateStudent from './create-student.vue'
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
const studentId = computed(() => studentStore.studentId)
const studentDetail = ref({})
const router = useRouter()

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
  try {
    const res = await getIntentStudentDetailApi({ studentId: studentId.value })
    if (res.code === 200) {
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

function refreshPrimaryCourseList() {
  nextTick(() => {
    registerForCoursesRef.value?.getTuitionAccountList()
  })
}

watch(() => props.open, (newVal) => {
  if (!newVal) {
    studentStore.clearStudentId()
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
    if (activeKey.value === '0') {
      refreshPrimaryCourseList()
    }
  }
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
      openStatusDropdown.value = false
      // 刷新学员详情
      getIntentStudentDetail(false)
    }
  } catch (error) {
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
      openIntentionDropdown.value = false
      // 刷新学员详情
      getIntentStudentDetail(false)
    }
  } catch (error) {
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
              <a-tooltip title="未关注">
                <img class="mt-0.5 cursor-pointer ml1" src="~@/assets/images/follow.svg" alt="">
              </a-tooltip>
              <a-tooltip title="未采集">
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
            {{ studentDetail.teacherName || '-' }}
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
          <a-descriptions-item label="微信号">
            {{ studentDetail.weChatNumber || '-' }}
          </a-descriptions-item>
          <a-descriptions-item label="年级">
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
            <register-for-courses ref="registerForCoursesRef" :empty-text="primaryCourseEmptyText" />
          </a-tab-pane>
          <a-tab-pane key="1" tab="上课记录">
            <class-record />
          </a-tab-pane>
          <a-tab-pane key="2" tab="订单记录">
            <order-record />
          </a-tab-pane>
          <a-tab-pane key="3" tab="试听记录">
            <try-listening-record />
          </a-tab-pane>
          <a-tab-pane key="4" tab="跟进记录">
            <follow-up-record ref="followUpRecordRef" :student-id="studentId" :student-detail="studentDetail" 
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
          {{ studentDetail.teacherName || '-' }}
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
        <a-descriptions-item label="微信号">
          {{ studentDetail.weChatNumber || '-' }}
        </a-descriptions-item>
        <a-descriptions-item label="家庭住址">
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

// 防止 descriptions label 换行
:deep(.ant-descriptions-item-label) {
  white-space: nowrap;
}
</style>
