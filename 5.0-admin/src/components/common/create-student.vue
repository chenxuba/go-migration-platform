<script setup>
import { CloseOutlined, LoadingOutlined } from '@ant-design/icons-vue'
import dayjs from 'dayjs'
import { debounce } from 'lodash-es'
import * as qiniu from 'qiniu-js'
import { message } from 'ant-design-vue'
import { useStudentFields } from '@/composables/useStudentFields'
import { checkStudentRepeatApi, getChannelTreeApi, getIntentStudentDetailApi, getRecommenderPageApi } from '@/api/enroll-center/intention-student'
import { getStudentPhoneNumberApi } from '@/api/common/config'
import { getQiniuToken } from '@/api/qiniu'
import { calculateAge } from '@/utils/date'
import { ParentRelationshipLabel, Sex, SexLabel } from '@/enums'
import StaffSelect from './staff-select.vue'
import messageService from '~@/utils/messageService'

const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
  record: {
    type: Object,
    default: () => ({}),
  },
  type: {
    type: Number,
    default: 1,
  },
})
const emit = defineEmits(['update:open', 'submit'])
const formRef = ref()
// 处理双向绑定
const openModal = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})
const spinning = ref(false)
// 向外部暴露一个方法，用于关闭loading和重置表单
function closeSpinning() {
  spinning.value = false
}

function createInitialFormState() {
  return {
    avatar: 'https://pcsys.admin.ybc365.com/a369a751-2be5-4929-974d-9ae4439f54c4.png',
    sex: 2,
    stuName: '',
    phoneRelationship: 1,
    mobile: '',
    channelId: undefined,
    birthday: '',
    salespersonId: undefined,
    weChatNumber: '',
    grade: undefined,
    studySchool: '',
    interest: '',
    recommendStudentId: undefined,
    address: '',
    remark: '',
    customInfo: [],
  }
}

function resetFormState() {
  const initialState = createInitialFormState()
  Object.keys(formState).forEach((key) => {
    if (!(key in initialState))
      delete formState[key]
  })
  Object.assign(formState, initialState)
}

// 重置表单
function resetForm() {
  resetFormState()
  formRef.value.resetFields()
}
function setChannelId(channelId) {
  formState.channelId = channelId
}
defineExpose({
  closeSpinning,
  resetForm,
  setChannelId,
})
const channelOptions = ref([])
// 渠道树下拉框可见状态 请求数据
// 递归处理渠道数据，设置禁用状态
function processChannelData(channels) {
  return channels.map(channel => ({
    ...channel,
    disabled: channel.isDisabled === true,
    channelList: channel.channelList ? processChannelData(channel.channelList) : []
  }))
}

async function getChannelTree() {
  try {
    const res = await getChannelTreeApi()
    // 过滤 channelList长度为0的数据，并处理禁用状态
    const filteredData = res.result.filter(item => item.channelList.length > 0)
    channelOptions.value = processChannelData(filteredData)
  }
  catch (error) {
    console.log(error)
  }
}

// 处理销售员选择变化
function handleSalespersonChange(value, staffInfo) {
  formState.salespersonId = value
}
function filter(inputValue, path) {
  return path.some(option => option.name.toLowerCase().includes(inputValue.toLowerCase()))
}
// 显示最后一级节点
function displayRender({ labels }) {
  return labels[labels.length - 1] // 取最后一个标签
}
// 机构配置
const userStore = useUserStore()
const instConfig = ref(userStore.instConfig)
const isSalesperson = ref(true)
const isReference = ref(true)
const formState = reactive(createInitialFormState())

function normalizeSelectEmptyValue(value) {
  return value === '' || value === null ? undefined : value
}

// 头像上传相关状态
const uploadingAvatar = ref(false)
const uploadProgress = ref(0)
const hasCustomAvatar = ref(false) // 标记是否上传了自定义头像

/**
 * 上传前校验
 */
const beforeAvatarUpload = (file) => {
  const isImage = file.type.startsWith('image/')
  if (!isImage) {
    message.error('只能上传图片文件!')
    return false
  }

  const isLt5M = file.size / 1024 / 1024 < 5
  if (!isLt5M) {
    message.error('图片大小不能超过 5MB!')
    return false
  }

  return true
}

/**
 * 自定义上传处理
 */
const handleAvatarUpload = async ({ file }) => {
  if (!beforeAvatarUpload(file)) {
    return
  }

  uploadingAvatar.value = true
  uploadProgress.value = 0

  try {
    // 1. 获取七牛云上传token
    const tokenRes = await getQiniuToken()
    const { token, uuid, buckethostname } = tokenRes.result

    // 2. 生成文件名（使用uuid + 原文件扩展名）
    const ext = file.name.substring(file.name.lastIndexOf('.'))
    const key = `avatar/${uuid}${ext}`

    // 3. 配置七牛云上传参数
    const config = {
      useCdnDomain: true,
      region: qiniu.region.z0 // 华东区域
    }

    const putExtra = {
      fname: file.name,
      mimeType: file.type
    }

    // 4. 创建上传observable
    const observable = qiniu.upload(file, key, token, putExtra, config)

    // 5. 订阅上传进度
    observable.subscribe({
      next(res) {
        // 更新上传进度
        uploadProgress.value = Math.floor(res.total.percent)
        console.log('上传进度:', uploadProgress.value + '%')
      },
      error(err) {
        console.error('上传失败:', err)
        message.error('上传失败: ' + err.message)
        uploadingAvatar.value = false
        uploadProgress.value = 0
      },
      complete(res) {
        // 上传完成
        const fileUrl = buckethostname + res.key
        console.log('上传成功:', fileUrl)

        formState.avatar = fileUrl
        hasCustomAvatar.value = true // 标记已上传自定义头像
        messageService.success('头像上传成功')
        
        // 延迟重置状态，让用户看到100%的进度
        setTimeout(() => {
          uploadingAvatar.value = false
          uploadProgress.value = 0
        }, 500)
      }
    })
  } catch (error) {
    console.error('获取token失败:', error)
    message.error('获取上传凭证失败')
    uploadingAvatar.value = false
    uploadProgress.value = 0
  }
}
function handlePhoneInput(e) {
  formState.mobile = e.target.value
    .replace(/\D/g, '') // 移除非数字
    .slice(0, 11) // 限制11位
}
// 禁选未来日期
function disabledDate(current) {
  return current > dayjs().endOf('day')
}
// 手动触发验证
async function handleSubmit() {
  try {
    spinning.value = true

    // 确保 channelOptions 已加载
    if (!channelOptions.value || channelOptions.value.length === 0) {
      await getChannelTree()
    }

    // 处理 channelId
    if (formState.channelId) {
      const channelPath = getChannelId(formState.channelId)
      if (channelPath.length > 0) {
        formState.channelId = channelPath
      }
    }

    // 强制触发表单验证
    try {
      // await formRef.value.validateFields(['stuName', 'mobile']);
      await formRef.value.validateFields()
    }
    catch (error) {
      console.log('表单验证失败:', error)
      spinning.value = false
      return
    }

    // 检查学生重复
    if (formState.stuName && formState.mobile) {
      const checkResult = await checkStudentRepeatApi({
        id: props.record.id || undefined,
        stuName: formState.stuName,
        mobile: formState.mobile,
      })

      if (checkResult.result && checkResult.result.addStudentRepeatRuleEnum !== 0) {
        // 设置重复状态
        repeatCheckState.value = {
          nameRepeat: checkResult.result.addStudentRepeatRuleEnum === 1 || checkResult.result.addStudentRepeatRuleEnum === 3,
          phoneRepeat: checkResult.result.addStudentRepeatRuleEnum === 1 || checkResult.result.addStudentRepeatRuleEnum === 2,
          errorMessage: checkResult.result.addStudentRepeatRuleEnum === 1
            ? '本机构已存在姓名和手机号同时相同的学员'
            : checkResult.result.addStudentRepeatRuleEnum === 2
              ? '本机构已存在手机号相同的学员'
              : '本机构已存在姓名相同的学员',
        }

        // 再次触发验证以显示错误信息
        await formRef.value.validateFields(['stuName', 'mobile'])
        spinning.value = false
        return
      }
    }

    // Process custom fields before submission
    const customInfo = customIsDisplayList.value.map(field => ({
      fieldId: field.id,
      fieldName: field.fieldKey,
      value: (formState[`${field.fieldKey}-${field.id}`]?.toString() || ''),
    }))

    // Update formState with processed customInfo
    formState.customInfo = customInfo
    console.log('验证通过，提交数据:', formState)
    emit('submit', formState)
  }
  catch (error) {
    console.log('验证失败:', error)
    spinning.value = false
  }
}
function closeFun() {
  resetFormState()
  formRef.value.resetFields()
  // 重置重复校验状态
  repeatCheckState.value = {
    nameRepeat: false,
    phoneRepeat: false,
    errorMessage: '',
  }
  // 重置自定义头像标记
  hasCustomAvatar.value = false
  // 重置销售员相关状态
  if (props.type === 1) {
    // formState.salespersonId = userStore.instUserId
  }
  openModal.value = false
}
const { systemDefaultIsDisplayList, customIsDisplayList, getAllStuFields } = useStudentFields()

function findSystemField(fieldKey) {
  return systemDefaultIsDisplayList.value.find(item => item.fieldKey === fieldKey)
}

function isSystemFieldVisible(fieldKey) {
  return !!findSystemField(fieldKey)
}

function isSystemFieldRequired(fieldKey) {
  return !!findSystemField(fieldKey)?.required
}

const dynamicSystemFields = computed(() => {
  const items = []
  if (isSystemFieldVisible('渠道')) {
    items.push({ key: 'channel', label: '渠道：', name: 'channelId' })
  }
  if (isSystemFieldVisible('生日')) {
    items.push({ key: 'birthday', label: '生日：', name: 'birthday' })
  }
  if (props.type === 1 && isSalesperson.value) {
    items.push({ key: 'salesperson', label: '销售员：', name: 'salespersonId' })
  }
  if (props.type === 1 && isReference.value) {
    items.push({ key: 'recommend', label: '推荐人：', name: 'recommendStudentId' })
  }
  if (isSystemFieldVisible('微信号')) {
    items.push({ key: 'wechat', label: '微信号：', name: 'weChatNumber' })
  }
  if (isSystemFieldVisible('年级')) {
    items.push({ key: 'grade', label: '年级：', name: 'grade' })
  }
  if (isSystemFieldVisible('就读学校')) {
    items.push({ key: 'school', label: '就读学校：', name: 'studySchool' })
  }
  if (isSystemFieldVisible('兴趣爱好')) {
    items.push({ key: 'interest', label: '兴趣爱好：', name: 'interest' })
  }
  if (isSystemFieldVisible('家庭住址')) {
    items.push({ key: 'address', label: '家庭住址：', name: 'address' })
  }
  return items
})

// 对自定义字段进行排序，fieldType为1的排在前面
const sortedCustomFields = computed(() => {
  return [...customIsDisplayList.value].sort((a, b) => {
    if (a.fieldType === 1 && b.fieldType !== 1)
      return -1
    if (a.fieldType !== 1 && b.fieldType === 1)
      return 1
    return 0
  })
})

const mergedGridFields = computed(() => {
  const systemItems = dynamicSystemFields.value.map(item => ({
    ...item,
    source: 'system',
  }))
  const customItems = sortedCustomFields.value.map(item => ({
    source: 'custom',
    key: `custom-${item.id}`,
    field: item,
  }))
  return [
    ...systemItems,
    ...customItems,
    {
      source: 'remark',
      key: 'remark',
      label: '学员备注：',
      name: 'remark',
    },
  ]
})

// Watch customIsDisplayList and initialize formState.customInfo
watch(customIsDisplayList, (newVal) => {
  if (newVal && newVal.length > 0) {
    formState.customInfo = newVal.map(item => ({
      fieldId: item.id,
      value: '',
    }))
  }
}, { immediate: true })

// 注意：不要在 systemDefaultIsDisplayList 上关 loading。弹窗打开时先拉字段再设 spinning=true，
// 若在此处关 loading 会与 openModal 里 spinning=true 竞态，导致创建模式 loading 常亮。

// 写一个方法 拿渠道id 反查渠道分类id 返回数组【渠道分类id，渠道id】 树结构 递归
function getChannelId(targetChannelId) {
  // 确保 channelOptions 已加载
  if (!channelOptions.value || channelOptions.value.length === 0) {
    console.warn('Channel options not loaded yet')
    return []
  }

  // 递归查找渠道ID的路径
  const findPath = (nodes, targetId, path = []) => {
    for (const node of nodes) {
      // 检查当前节点
      if (node.id == targetId) {
        return [...path, node.id]
      }
      // 检查子节点
      if (node.channelList && node.channelList.length > 0) {
        const result = findPath(node.channelList, targetId, [...path, node.id])
        if (result) {
          return result
        }
      }
    }
    return null
  }

  // 从根节点开始查找
  const path = findPath(channelOptions.value, targetChannelId)
  return path || []
}
// 获取学生信息
async function getStudentInfo() {
  try {
    const res = await getIntentStudentDetailApi({ studentId: props.record.id })
    if (res.code === 200) {
      const { result } = res
      // 使用对象映射来更新表单状态  系统返回字段：表单字段
      const fieldMapping = {
        id: 'studentId',
        uuid: 'uuid',
        version: 'version',
        stuName: 'stuName',
        mobile: 'mobile',
        stuSex: 'sex',
        birthDay: 'birthday',
        channelId: 'channelId',
        salePerson: 'salespersonId',
        weChatNumber: 'weChatNumber',
        phoneRelationship: 'phoneRelationship',
        grade: 'grade',
        studySchool: 'studySchool',
        interest: 'interest',
        recommendStudentId: 'recommendStudentId',
        address: 'address',
        remark: 'remark',
        avatarUrl: 'avatar',
      }
      // 使用 Object.entries 遍历映射并更新表单状态
      Object.entries(fieldMapping).forEach(([apiField, formField]) => {
        if (result[apiField] !== undefined) {
          if (['channelId', 'salespersonId', 'grade', 'recommendStudentId'].includes(formField)) {
            formState[formField] = normalizeSelectEmptyValue(result[apiField])
          }
          else {
            formState[formField] = result[apiField]
          }
          // 如果是编辑模式且有自定义头像，标记为已上传
          if (formField === 'avatar' && result[apiField]) {
            hasCustomAvatar.value = true
          }
        }
      })

      // 特殊字段处理
      if (result.channelId) {
        formState.channelId = getChannelId(result.channelId)
      }
      else {
        formState.channelId = undefined
      }
      // 处理自定义字段
      if (result.customInfo && Array.isArray(result.customInfo)) {
        result.customInfo.forEach((customField) => {
          const { fieldId, fieldName, value } = customField
          // 直接按 fieldName-id 回填，避免字段配置异步加载导致回填丢失
          formState[`${fieldName}-${fieldId}`] = value || undefined
        })
      }

      // 在编辑模式下，调用API解密手机号
      try {
        const phoneRes = await getStudentPhoneNumberApi({ studentId: props.record.id })
        if (phoneRes.code === 200) {
          formState.mobile = phoneRes.result
        }
      } catch (error) {
        console.error('解密手机号失败:', error)
        // 如果解密失败，保持原来的手机号
      }
    }
    else {
      messageService.error(res.message || '获取学员信息失败')
    }
  }
  catch (error) {
    console.error('获取学生信息失败:', error)
  }
  finally {
    spinning.value = false
  }
}

// 监听性别变化更新头像
watch(() => formState.sex, (newVal) => {
  // 如果是编辑模式或已上传自定义头像，不自动切换
  if (props.type === 2 || hasCustomAvatar.value) {
    return
  }
  
  // 只在创建模式且未上传自定义头像时，根据性别切换默认头像
  if (newVal === 1) {
    formState.avatar = 'https://pcsys.admin.ybc365.com/c04d0ea2-a8b0-4001-b19b-946a980cb726.png'
  }
  else if (newVal === 0) {
    formState.avatar = 'https://pcsys.admin.ybc365.com/d92afddc-ffac-40aa-aa61-bd97d91aa1ec.png'
  }
  else {
    formState.avatar = 'https://pcsys.admin.ybc365.com/a369a751-2be5-4929-974d-9ae4439f54c4.png'
  }
})

// 添加重复校验状态
const repeatCheckState = ref({
  nameRepeat: false,
  phoneRepeat: false,
  errorMessage: '',
})

// 检查学生重复
async function checkStudentRepeat() {
  // 只有当学生姓名和手机号都有值，且手机号格式正确时才进行校验
  if (!formState.stuName || !formState.mobile || !/^1[3-9]\d{9}$/.test(formState.mobile)) {
    // 重置校验状态
    repeatCheckState.value = {
      nameRepeat: false,
      phoneRepeat: false,
      errorMessage: '',
    }
    return
  }

  try {
    const res = await checkStudentRepeatApi({
      id: props.record.id || undefined,
      stuName: formState.stuName,
      mobile: formState.mobile,
    })

    // 重置校验状态
    repeatCheckState.value = {
      nameRepeat: false,
      phoneRepeat: false,
      errorMessage: '',
    }

    if (res.result) {
      const { addStudentRepeatRuleEnum } = res.result

      // 如果 addStudentRepeatRuleEnum 为 0，保持状态清空
      if (addStudentRepeatRuleEnum === 0) {
        // 清除之前的验证错误
        formRef.value?.clearValidate(['stuName', 'mobile'])
        return
      }

      // 根据不同的重复规则设置不同的错误信息
      switch (addStudentRepeatRuleEnum) {
        case 1: // 姓名和手机号同时重复
          repeatCheckState.value.nameRepeat = true
          repeatCheckState.value.phoneRepeat = true
          repeatCheckState.value.errorMessage = '本机构已存在姓名和手机号同时相同的学员'
          break
        case 2: // 手机号重复
          repeatCheckState.value.phoneRepeat = true
          repeatCheckState.value.errorMessage = '本机构已存在手机号相同的学员'
          break
        case 3: // 姓名重复
          repeatCheckState.value.nameRepeat = true
          repeatCheckState.value.errorMessage = '本机构已存在姓名相同的学员'
          break
      }

      // 如果有重复，触发表单验证
      if (addStudentRepeatRuleEnum > 0) {
        formRef.value?.validateFields(['stuName', 'mobile'])
      }
    }
  }
  catch (error) {
    console.error('检查学生重复出错：', error)
  }
}

// 更新表单验证规则
const getFormRules = computed(() => ({
  stuName: [
    { required: true, message: '学生姓名不能为空' },
    {
      validator: async (_, value) => {
        // 只做重复校验，不再做非空判断
        if (repeatCheckState.value.nameRepeat && formState.mobile && /^1[3-9]\d{9}$/.test(formState.mobile)) {
          return Promise.reject(repeatCheckState.value.errorMessage)
        }
        return Promise.resolve()
      },
    },
  ],
  mobile: [
    { required: true, message: '联系电话不能为空' },
    {
      validator: async (_, value) => {
        // 只做格式和重复校验，不再做非空判断
        if (value && !/^1[3-9]\d{9}$/.test(value)) {
          repeatCheckState.value.phoneRepeat = false // 清除重复校验状态
          return Promise.reject('无效的手机号码')
        }
        if (repeatCheckState.value.phoneRepeat && formState.stuName) {
          return Promise.reject(repeatCheckState.value.errorMessage)
        }
        return Promise.resolve()
      },
    },
  ],
}))

// 修改监听器
watch(openModal, async (newVal) => {
  if (!newVal)
    return
  try {
    resetFormState()
    await getAllStuFields({ filter: 3 })
    spinning.value = true
    instConfig.value = userStore.instConfig
    await getChannelTree()

    if (props.type === 1) {
      formState.salespersonId = userStore.instUserId
    }
    else if (props.record && props.record.id) {
      await getStudentInfo()
    }

    if (props.type === 2) {
      nextTick(() => {
        formRef.value?.clearValidate(['stuName', 'mobile'])
      })
    }
  }
  catch (error) {
    console.error('创建学员弹窗初始化失败:', error)
    spinning.value = false
  }
  finally {
    // 创建模式：拉完渠道等即结束 loading（不再依赖字段列表 watch，避免与上面 spinning=true 竞态）
    if (props.type === 1) {
      spinning.value = false
    }
    // 编辑模式但未带 id：无法走 getStudentInfo 的 finally
    if (props.type === 2 && !(props.record && props.record.id)) {
      spinning.value = false
    }
  }
})

// 学员搜索相关
const stuListOptions = ref([])
const pagination = ref({
  current: 1,
  pageSize: 5,
  total: 0,
  showTotal: total => `共 ${total} 条`,
})
const finished = ref(false)
const isLoading = ref(false)

// 处理下拉菜单打开状态变化
async function dropdownVisibleChangeFun(visible) {
  if (!visible || formState.recommendStudentId || finished.value)
    return
  pagination.value.current = 1
  stuListOptions.value = []
  finished.value = false
  await getRecommenderPage()
}

// 获取推荐人列表
async function getRecommenderPage(params = { key: undefined, studentStatus: undefined }) {
  try {
    if (finished.value)
      return
    isLoading.value = true
    const res = await getRecommenderPageApi({
      'pageRequestModel': {
        'needTotal': true,
        'pageSize': pagination.value.pageSize,
        'pageIndex': pagination.value.current,
        'skipCount': 0,
      },
      'queryModel': {
        'searchKey': params.key,
        'studentStatus': params.studentStatus,
      },
      'sortModel': {},
    })

    if (res.code === 200) {
      // 保留首次加载的清空逻辑
      if (pagination.value.current === 1) {
        stuListOptions.value = res.result
      }
      else {
        stuListOptions.value = [...stuListOptions.value, ...res.result]
      }
      pagination.value.total = res.total

      if (stuListOptions.value.length >= pagination.value.total) {
        finished.value = true
      }
    }
  }
  catch (error) {
    console.error('加载数据失败:', error)
    // 发生错误时回退页码
    if (pagination.value.current > 1) {
      pagination.value.current -= 1
    }
  }
  finally {
    isLoading.value = false
  }
}

// 文本框变化触发搜索
const handleSearchStuPhone = debounce((value) => {
  pagination.value.current = 1
  finished.value = false
  getRecommenderPage({ key: value })
}, 300)

// 处理滚动加载更多
function handlePopupScroll(event) {
  const { target } = event
  const { scrollTop, scrollHeight, clientHeight } = target
  // 判断是否滚动到底部
  if (scrollHeight - scrollTop - clientHeight < 1) {
    // 检查是否正在加载且还有更多数据
    if (!isLoading.value && pagination.value.current * pagination.value.pageSize < pagination.value.total) {
      isLoading.value = true
      pagination.value.current += 1
      getRecommenderPage()
    }
  }
}

// 选择推荐人
function handleRecommendStudentIdChange(value) {
  finished.value = false
}


</script>

<template>
  <a-modal v-model:open="openModal" centered class="createStu-modal-content-box" :keyboard="false" :closable="false"
    :mask-closable="false" :width="800">
    <template #title>
      <div class="text-5 flex justify-between flex-center">
        <span>{{ type === 1 ? '创建学员' : '编辑学员' }}</span>
        <a-button type="text" class="close-btn" @click="closeFun">
          <template #icon>
            <CloseOutlined class="text-5 close-icon" />
          </template>
        </a-button>
      </div>
    </template>
    <a-spin :spinning="spinning">
      <div class="stu-content scrollbar">
        <a-form ref="formRef" layout="vertical" :model="formState">
          <a-row :gutter="24" class="flex flex-items-center">
            <a-col :xs="24" :sm="12" :md="12" :lg="12" :xl="12">
              <div class="leftInfo">
                <a-form-item>
                  <div class="avatar-upload-wrapper">
                    <div class="avatar-container" :class="{ 'uploading': uploadingAvatar }">
                      <a-image :width="80" class="rounded-10" :src="formState.avatar" />
                      <svg 
                        v-if="uploadingAvatar" 
                        class="progress-ring" 
                        width="88" 
                        height="88"
                      >
                        <circle
                          class="progress-ring-circle-bg"
                          stroke="#e6e6e6"
                          stroke-width="3"
                          fill="transparent"
                          r="42"
                          cx="44"
                          cy="44"
                        />
                        <circle
                          class="progress-ring-circle"
                          stroke="url(#gradient)"
                          stroke-width="3"
                          fill="transparent"
                          r="42"
                          cx="44"
                          cy="44"
                          :style="{
                            strokeDasharray: `${2 * Math.PI * 42}`,
                            strokeDashoffset: `${2 * Math.PI * 42 * (1 - uploadProgress / 100)}`
                          }"
                        />
                        <defs>
                          <linearGradient id="gradient" x1="0%" y1="0%" x2="100%" y2="100%">
                            <stop offset="0%" style="stop-color:#1890ff;stop-opacity:1" />
                            <stop offset="100%" style="stop-color:#52c41a;stop-opacity:1" />
                          </linearGradient>
                        </defs>
                      </svg>
                      <div v-if="uploadingAvatar" class="progress-text">
                        {{ uploadProgress }}%
                      </div>
                    </div>
                    <a-upload
                      :custom-request="handleAvatarUpload"
                      :show-upload-list="false"
                      accept="image/*"
                      :disabled="uploadingAvatar"
                    >
                      <a-button class="ml4 text-#666" :loading="uploadingAvatar">
                        <template v-if="!uploadingAvatar">
                          上传学员头像
                        </template>
                        <template v-else>
                          上传中...
                        </template>
                      </a-button>
                    </a-upload>
                  </div>
                  <div class="upload-tip">支持jpg、png格式，大小不超过5MB</div>
                </a-form-item>
                <div class="stusex">
                  <a-form-item label="学员性别：" name="sex"
                    :rules="systemDefaultIsDisplayList.find(item => item.fieldKey === '性别')?.required ? [{ required: true, message: '性别不能为空' }] : []">
                    <a-radio-group v-model:value="formState.sex" class="custom-radio">
                      <a-radio :value="Sex.Unknown">
                        {{ SexLabel[Sex.Unknown] }}
                      </a-radio>
                      <a-radio :value="Sex.Male">
                        {{ SexLabel[Sex.Male] }}
                      </a-radio>
                      <a-radio :value="Sex.Female">
                        {{ SexLabel[Sex.Female] }}
                      </a-radio>
                    </a-radio-group>
                  </a-form-item>
                </div>
              </div>
            </a-col>
            <a-col :xs="24" :sm="12" :md="12" :lg="12" :xl="12">
              <div class="rightInfo">
                <!-- 学生姓名： 必填 -->
                <a-form-item label="学生姓名：" name="stuName" :rules="getFormRules.stuName">
                  <a-input v-model:value="formState.stuName" placeholder="请输入学生姓名" @blur="checkStudentRepeat" />
                </a-form-item>
                <a-form-item label="联系人电话：" name="mobile" :rules="getFormRules.mobile">
                  <a-input-group compact style="width: 100%; display: flex;">
                    <a-form-item-rest>
                      <a-select v-model:value="formState.phoneRelationship" style="width: 70px;text-align: center;">
                        <a-select-option v-for="(label, value) in ParentRelationshipLabel" :key="value"
                          :value="Number(value)">
                          {{ label }}
                        </a-select-option>
                      </a-select>
                    </a-form-item-rest>
                    <a-input v-model:value="formState.mobile" type="tel" :maxlength="11" placeholder="请输入"
                      class="flex-1" @input="handlePhoneInput" @blur="checkStudentRepeat" />
                  </a-input-group>
                </a-form-item>
              </div>
            </a-col>
          </a-row>
          <div class="system-grid">
            <div
              v-for="item in mergedGridFields"
              :key="item.key"
              class="system-grid__item"
            >
              <a-form-item
                v-if="item.source === 'system'"
                :label="item.label"
                :name="item.name"
                :rules="item.key === 'salesperson'
                  ? [{ required: true, message: '销售员不能为空' }]
                  : isSystemFieldRequired(item.key === 'channel' ? '渠道' : item.key === 'birthday' ? '生日' : item.key === 'wechat' ? '微信号' : item.key === 'grade' ? '年级' : item.key === 'school' ? '就读学校' : item.key === 'interest' ? '兴趣爱好' : item.key === 'recommend' ? '推荐人' : item.key === 'address' ? '家庭住址' : '') ? [{ required: true, message: `${item.label.replace('：', '')}不能为空` }] : []"
              >
                <a-cascader
                  v-if="item.key === 'channel'"
                  v-model:value="formState.channelId"
                  :field-names="{ children: 'channelList', label: 'name', value: 'id' }"
                  :show-search="{ filter }"
                  :display-render="displayRender"
                  :options="channelOptions"
                  placeholder="搜索渠道"
                  style="width: 100%"
                />
                <a-date-picker
                  v-else-if="item.key === 'birthday'"
                  v-model:value="formState.birthday"
                  value-format="YYYY-MM-DD"
                  :disabled-date="disabledDate"
                  placeholder="请选择日期"
                  style="width: 100%"
                  :format="(date) => {
                    if (date) {
                      const birthDate = dayjs(date);
                      const now = dayjs();
                      const years = now.diff(birthDate, 'year');
                      const months = now.diff(birthDate, 'month') % 12;
                      return `${calculateAge(birthDate)} ${birthDate.format('YYYY年MM月DD日')}`;
                    }
                    return '';
                  }"
                />
                <StaffSelect
                  v-else-if="item.key === 'salesperson'"
                  v-model="formState.salespersonId"
                  placeholder="搜索姓名/手机号"
                  width="100%"
                  :status="0"
                  @change="handleSalespersonChange"
                />
                <a-select
                  v-else-if="item.key === 'recommend'"
                  v-model:value="formState.recommendStudentId"
                  allow-clear
                  style="width: 100%"
                  placeholder="搜索姓名/手机号"
                  show-search
                  :filter-option="false"
                  option-label-prop="label"
                  @change="handleRecommendStudentIdChange"
                  @dropdown-visible-change="dropdownVisibleChangeFun"
                  @search="handleSearchStuPhone"
                  @popup-scroll="handlePopupScroll"
                >
                  <a-select-option v-for="option in stuListOptions" :key="option.id" :value="option.id" :data="option" :label="option.stuName">
                    <div class="flex flex-center mb-1 justify-between">
                      <div class="flex">
                        <div>
                          <img class="w-10 rounded-10" :src="option.avatarUrl" alt="">
                        </div>
                        <div class="ml-2 mr-3">
                          <div class="text-sm text-#666 leading-7">
                            {{ option.stuName }}
                          </div>
                          <div class="text-xs text-#888">
                            {{ option.mobile }}
                          </div>
                        </div>
                      </div>
                      <div>
                        <a-tag v-if="option.studentStatus == 1" :bordered="false" color="processing">
                          在读学员
                        </a-tag>
                        <a-tag v-else-if="option.studentStatus == 0" :bordered="false" color="orange">
                          意向学员
                        </a-tag>
                      </div>
                    </div>
                  </a-select-option>
                </a-select>
                <a-input
                  v-else-if="item.key === 'wechat'"
                  v-model:value="formState.weChatNumber"
                  placeholder="请输入微信号"
                />
                <a-select
                  v-else-if="item.key === 'grade'"
                  v-model:value="formState.grade"
                  style="width: 100%"
                  placeholder="请选择"
                >
                  <a-select-option v-for="option in findSystemField('年级')?.optionsJson?.split(',') || []" :key="option" :value="option">
                    {{ option }}
                  </a-select-option>
                </a-select>
                <a-input
                  v-else-if="item.key === 'school'"
                  v-model:value="formState.studySchool"
                  placeholder="请输入就读学校"
                />
                <a-input
                  v-else-if="item.key === 'interest'"
                  v-model:value="formState.interest"
                  placeholder="请输入兴趣爱好"
                />
                <a-input
                  v-else-if="item.key === 'address'"
                  v-model:value="formState.address"
                  placeholder="请输入地址"
                />
              </a-form-item>
              <a-form-item
                v-else-if="item.source === 'custom'"
                :label="`${item.field.fieldKey}：`"
                :name="`${item.field.fieldKey}-${item.field.id}`"
                :rules="item.field.required ? [{ required: true, message: `${item.field.fieldKey}不能为空` }] : []"
              >
                <a-input
                  v-if="item.field.fieldType == '1'"
                  v-model:value="formState[`${item.field.fieldKey}-${item.field.id}`]"
                  :placeholder="`请输入${item.field.fieldKey}`"
                />
                <a-input-number
                  v-else-if="item.field.fieldType == '2'"
                  v-model:value="formState[`${item.field.fieldKey}-${item.field.id}`]"
                  :placeholder="`请输入${item.field.fieldKey}`"
                  class="w-full"
                />
                <a-date-picker
                  v-else-if="item.field.fieldType == '3'"
                  v-model:value="formState[`${item.field.fieldKey}-${item.field.id}`]"
                  format="YYYY-MM-DD"
                  value-format="YYYY-MM-DD"
                  :placeholder="`请选择${item.field.fieldKey}`"
                  style="width: 100%"
                />
                <a-select
                  v-else-if="item.field.fieldType == '4'"
                  v-model:value="formState[`${item.field.fieldKey}-${item.field.id}`]"
                  allow-clear
                  :placeholder="`请选择${item.field.fieldKey}`"
                  style="width: 100%"
                >
                  <a-select-option v-for="option in item.field.optionsList" :key="option.id" :value="option.value">
                    {{ option.value }}
                  </a-select-option>
                  </a-select>
              </a-form-item>
              <a-form-item v-else-if="item.source === 'remark'" :label="item.label" :name="item.name">
                <a-input v-model:value="formState.remark" placeholder="请输入备注" />
              </a-form-item>
            </div>
          </div>
        </a-form>
      </div>
    </a-spin>
    <template #footer>
      <a-button danger ghost @click="closeFun">
        关闭
      </a-button>
      <a-button type="primary" ghost :loading="spinning" @click="handleSubmit">
        确定
      </a-button>
    </template>
  </a-modal>
</template>

<style lang="less" scoped>
.createStu-modal-content-box {
  .stu-content {
    // min-width:800px;
    max-height: calc(100vh - 155px);
    padding: 24px 40px 0 !important;
    overflow: auto;

  }
}

.avatar-upload-wrapper {
  display: flex;
  align-items: center;
  gap: 16px;
}

.system-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0 24px;
}

.system-grid__item {
  min-width: 0;
}

.system-grid__item--full {
  grid-column: 1 / -1;
}

@media (max-width: 768px) {
  .system-grid {
    grid-template-columns: 1fr;
    gap: 0;
  }

  .system-grid__item--full {
    grid-column: auto;
  }
}

.avatar-container {
  position: relative;
  width: 88px;
  height: 88px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.progress-ring {
  position: absolute;
  top: 0;
  left: 0;
  transform: rotate(-90deg);
  z-index: 1;
}

.progress-ring-circle-bg {
  opacity: 0.3;
}

.progress-ring-circle {
  transition: stroke-dashoffset 0.3s ease;
  stroke-linecap: round;
}

.progress-text {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  font-size: 12px;
  font-weight: 600;
  color: #1890ff;
  z-index: 2;
  background: rgba(255, 255, 255, 0.9);
  padding: 2px 6px;
  border-radius: 4px;
  pointer-events: none;
}

.upload-tip {
  margin-top: 8px;
  color: #999;
  font-size: 12px;
}

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

:deep(.ant-image-mask) {
  border-radius: 100px;
}

:deep(.ant-picker .ant-picker-input >input) {
  color: inherit !important;
}

/* 自定义镂空样式 */
.custom-radio ::v-deep(.ant-radio-wrapper:hover .ant-radio),
.custom-radio ::v-deep(.ant-radio:hover .ant-radio-inner),
.custom-radio ::v-deep(.ant-radio-input:focus + .ant-radio-inner) {
  border-color: var(--pro-ant-color-primary);
}

.custom-radio ::v-deep(.ant-radio-inner) {
  background-color: transparent;
  border-color: #d9d9d9;
}

.custom-radio ::v-deep(.ant-radio-checked .ant-radio-inner) {
  background-color: transparent;
  border-color: var(--pro-ant-color-primary);
}

.custom-radio ::v-deep(.ant-radio-inner::after) {
  background-color: var(--pro-ant-color-primary);
  transform: scale(0.5);
}

/* 学员搜索下拉样式 */
:deep(.ant-select-item-option-content) {
  display: inline-block;
  width: 100%;
}

.w-10 {
  width: 2.5rem;
}

.w-8 {
  width: 2rem;
}

.h-8 {
  height: 2rem;
}

.rounded-10 {
  border-radius: 10px;
}

.rounded-full {
  border-radius: 50%;
}

.flex-center {
  display: flex;
  align-items: center;
}

.ml-2 {
  margin-left: 0.5rem;
}

.mr-3 {
  margin-right: 0.75rem;
}

.text-sm {
  font-size: 0.875rem;
  line-height: 1.25rem;
}

.text-xs {
  font-size: 0.75rem;
  line-height: 1rem;
}

.leading-7 {
  line-height: 1.75rem;
}

.mr-2 {
  margin-right: 0.5rem;
}

.p-2 {
  padding: 0.5rem;
}

.border {
  border: 1px solid #d9d9d9;
}

.rounded {
  border-radius: 0.375rem;
}

.font-medium {
  font-weight: 500;
}

.text-gray-500 {
  color: #6b7280;
}
</style>

<style>
.createStu-modal-content-box .ant-modal-header {
  padding: 10px 16px !important;
  margin-bottom: 0;
}

.createStu-modal-content-box .ant-modal-body {
  padding: 0 !important;
}
</style>
