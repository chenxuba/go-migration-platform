<script setup>
import { CloseOutlined, QuestionCircleOutlined } from '@ant-design/icons-vue'
import emitter, { EVENTS } from '~@/utils/eventBus'
import { TreeSelect, message } from 'ant-design-vue'
import selectRole from './selectRole.vue'
import { getInstUserDetail, updateInstUserDetail, checkPhoneUsedApi, changePhoneWithOtherApi } from '@/api/internal-manage/staff-manage'
import { getInstRolePageApi, updateRoleApi } from '@/api/internal-manage/role-manage'
import RolesDetailsDrawer from '@/components/common/roles-details-drawer.vue'
import messageService from '~@/utils/messageService'
import { useThrottleFn } from '@vueuse/core'
import { useUserStore } from '@/stores/user'

const props = defineProps({
  detail: {
    type: Object,
    default: () => ({}),
  },
  departmentList: {
    type: Array,
    default: () => [],
  },
})

const emit = defineEmits(['success', 'refresh-list'])
const open = defineModel({
  type: Boolean,
  default: false,
})

const userStore = useUserStore()
const safeSelectRoleList = computed(() => Array.isArray(selectRoleList.value) ? selectRoleList.value : [])

// 角色详情相关
const roleDetailsOpen = ref(false)
const currentRoleId = ref(null)
const currentRoleDetails = ref({})
const roleInfos = ref([])
const selectRoleList = ref([])
const detailInfo = ref({})
const form = ref(null)
const addEmployeesOpen = ref(false)
const submitLoading = ref(false)

// 表单数据
const formState = reactive({
  nickName: '',
  mobile: '',
  deptIds: undefined,
  roleIds: [],
  disabled: false, // false: 在职中, true: 已离职
  userType: 1, // 1: 正式员工, 2: 兼职员工
  avatar: '',
})

// 修改手机号相关状态
const changeMobileOpen = ref(false)
const changeMobileLoading = ref(false)
const changeMobileForm = ref(null)
const changeMobileFormState = reactive({
  mobile: '',
})
const mobileValidationState = ref({
  validateStatus: '',
  help: '',
})

// 打开角色详情抽屉
function roleDetailsFunc(roleInfo) {
  currentRoleId.value = roleInfo.id
  currentRoleDetails.value = roleInfo
  roleDetailsOpen.value = true
}

// 获取员工详情数据
async function fetchUserDetail() {
  try {
    const { result } = await getInstUserDetail({ id: props.detail.id })
    detailInfo.value = result

    // 更新表单数据
    Object.keys(formState).forEach(key => {
      if (key in result) {
        formState[key] = result[key]
      }
    })

    formState.roleIds = Array.isArray(result.roleIds) ? result.roleIds : []
    selectRoleList.value = Array.isArray(result.roleIds) ? result.roleIds : []
  } catch (error) {
    messageService.error('获取员工详情失败')
    console.error('获取员工详情失败:', error)
  }
}

// 获取角色列表
async function fetchRoleList() {
  try {
    const { result } = await getInstRolePageApi({
      pageRequestModel: {
        needTotal: true,
        pageSize: 500,
        pageIndex: 1,
        skipCount: 1,
      },
      queryModel: {},
    })
    roleInfos.value = result
  } catch (error) {
    messageService.error('获取角色列表失败')
    // console.error('获取角色列表失败:', error)
  }
}

// 角色权限编辑成功回调
async function handleRoleEditSuccess(data) {
  try {
    const res = await updateRoleApi(data)
    if (res.code === 200) {
      messageService.success('更新成功')
      await Promise.all([fetchUserDetail(), fetchRoleList()])
    }

    // 关闭按钮loading
    emitter.emit(EVENTS.CLOSE_LOADING_EVENT)
  }
  catch (err) {
    console.error('更新角色失败:', err)
    messageService.error('更新失败')
  }
}

// 提交表单 - 添加防抖处理
const handleOk = useThrottleFn(async () => {
  if (submitLoading.value) return
  submitLoading.value = true

  try {
    const valid = await form.value.validate()
    if (valid) {
      await updateInstUserDetail({ ...formState, id: detailInfo.value.id })
      
      // 如果编辑的是当前登录用户，则更新store中的用户信息
      if (userStore.instUserId === detailInfo.value.id) {
        // 重新获取用户信息以更新store
        await userStore.getUserInfo()
      }
      
      form.value.resetFields()
      emit('success')
      messageService.success('员工信息修改成功')
      open.value = false
    }
  } catch (error) {
    console.error('提交表单失败:', error)
    messageService.error('提交失败')
  } finally {
    submitLoading.value = false
  }
}, 500)

// 处理选择角色成功
function handleSelectRoleSuccess(roleIds) {
  selectRoleList.value = roleIds
  formState.roleIds = roleIds
}

// 关闭弹窗
function cancel() {
  open.value = false
  form.value?.resetFields()
  selectRoleList.value = []
}

// 头像上传处理
function handleAvatarChange(info) {
  // 处理头像上传逻辑
  if (info.file.status === 'done') {
    // 假设上传成功后，服务器返回了头像URL
    const response = info.file.response
    if (response && response.url) {
      formState.avatar = response.url
    }
  }
}

// 打开修改手机号弹窗
function openChangeMobile() {
  changeMobileOpen.value = true
  changeMobileFormState.mobile = formState.mobile || detailInfo.value.mobile || ''
  mobileValidationState.value = {
    validateStatus: '',
    help: '',
  }
}

// 关闭修改手机号弹窗
function closeChangeMobile() {
  changeMobileOpen.value = false
  changeMobileFormState.mobile = ''
  mobileValidationState.value = {
    validateStatus: '',
    help: '',
  }
}

// 实时验证手机号格式
function validateMobileFormat(mobile) {
  const mobileRegex = /^1[3-9]\d{9}$/
  if (!mobile) {
    return {
      validateStatus: '',
      help: '',
    }
  }
  if (!mobileRegex.test(mobile)) {
    return {
      validateStatus: 'error',
      help: '请输入正确的手机号',
    }
  }
  return {
    validateStatus: 'success',
    help: '',
  }
}

// 手机号输入变化处理
function handleMobileChange(e) {
  // 只允许输入数字
  const value = e.target.value.replace(/\D/g, '').slice(0, 11)
  changeMobileFormState.mobile = value
  
  const validation = validateMobileFormat(value)
  mobileValidationState.value = validation
}

// 提交修改手机号
async function submitChangeMobile() {
  // 验证手机号格式
  const validation = validateMobileFormat(changeMobileFormState.mobile)
  if (validation.validateStatus === 'error' || !changeMobileFormState.mobile) {
    mobileValidationState.value = validation.validateStatus === 'error' 
      ? validation 
      : { validateStatus: 'error', help: '请输入手机号' }
    return
  }

  changeMobileLoading.value = true

  try {
    // 1. 检查手机号是否被占用
    let checkResult
    try {
      checkResult = await checkPhoneUsedApi({ mobile: changeMobileFormState.mobile })
    } catch (error) {
      console.error('检查手机号是否被占用失败:', error)
      messageService.error('检查手机号失败，请重试')
      return
    }
    
    if (checkResult.code!== 200) {
      // 手机号已被占用
      mobileValidationState.value = {
        validateStatus: 'error',
        help: '手机号已被占用',
      }
      return
    }

    // 2. 如果没被占用，执行更换手机号
    const changeResult = await changePhoneWithOtherApi({
      mobile: changeMobileFormState.mobile,
      userId: detailInfo.value.id,
    })

    if (changeResult.code === 200) {
      messageService.success('手机号修改成功')
      closeChangeMobile()
      // 重新获取员工详情更新显示
      await fetchUserDetail()
      // 通知父组件刷新列表
      emit('refresh-list')
      // emit('success')
    }
  } catch (error) {
    console.error('修改手机号失败:', error)
    messageService.error('修改手机号失败')
  } finally {
    changeMobileLoading.value = false
  }
}

// 监听弹窗打开，加载数据
watch(
  () => open.value,
  async (newValue) => {
    if (newValue) {
      await Promise.all([fetchUserDetail(), fetchRoleList()])
    }
  },
)
</script>

<template>
  <a-modal v-model:open="open" class="modal-content-box business-settings-modal responsive-modal" :width="800"
    :destroy-on-close="true" :centered="true" :keyboard="false" :closable="false" :mask-closable="false"
    style="top:12px">
    <template #title>
      <div class="text-5 flex justify-between flex-center">
        <span>编辑员工</span>
        <a-button type="text" class="close-btn" @click="cancel">
          <template #icon>
            <CloseOutlined class="text-5 close-icon" />
          </template>
        </a-button>
      </div>
    </template>
    <div class="contenter scrollbar">
      <!-- 表单部分 -->
      <a-form ref="form" :model="formState" autocomplete="off" layout="vertical">
        <div class="form-container flex justify-between w-100% gap-30px">
          <!-- 左侧表单 -->
          <div class="form-left w-50%">
            <!-- 头像上传 -->
            <a-form-item>
              <div class="flex items-center gap-20px">
                <div class="upload-area">
                  <div class="text-#fff text-30px font-medium">
                    {{ formState.nickName ? formState.nickName[0] : '-' }}
                  </div>
                </div>
                <a-upload name="avatar" list-type="picture-circle" :show-upload-list="false" :disabled="detail.isAdmin"
                  @change="handleAvatarChange">
                  <a-button :disabled="detail.isAdmin">上传照片</a-button>
                </a-upload>
              </div>
            </a-form-item>

            <!-- 员工姓名 -->
            <a-form-item label="员工姓名：" name="nickName" :rules="[{ required: true, message: '请输入员工姓名' }]">
              <a-input v-model:value="formState.nickName" :disabled="detail.isAdmin" placeholder="请输入" />
            </a-form-item>

            <!-- 所属部门 -->
            <a-form-item label="所属部门：" name="deptIds" :rules="[{ required: true, message: '请选择所属部门' }]">
              <a-tree-select v-model:value="formState.deptIds" style="width: 100%"
                :dropdown-style="{ maxHeight: '400px', overflow: 'auto' }" placeholder="请选择" allow-clear multiple
                tree-default-expand-all :tree-data="departmentList" tree-node-filter-prop="departName"
                :field-names="{ children: 'children', label: 'departName', value: 'id' }"
                :show-checked-strategy="TreeSelect.SHOW_ALL" />
            </a-form-item>

            <!-- 任职角色 -->
            <a-form-item style="margin: 0;" class="position" name="roleIds"
              :rules="[{ required: true, message: '请设置任职角色' }]">
              <template #label>
                <div class="flex items-center">
                  <div>任职角色：</div>
                  <a-button type="primary" @click="addEmployeesOpen = true">编辑</a-button>
                </div>
              </template>
            </a-form-item>
          </div>

          <!-- 右侧表单 -->
          <div class="form-right w-50% mt-18px">
            <!-- 账号状态 -->
            <a-form-item label="账号状态：" :required="true">
              <a-radio-group v-model:value="formState.disabled" :disabled="detail.isAdmin" class="custom-radio">
                <a-radio :value="false">在职中</a-radio>
                <a-radio :value="true">已离职</a-radio>
              </a-radio-group>
            </a-form-item>

            <!-- 员工手机号 -->
            <a-form-item name="mobile" :rules="[
              { required: true, message: '请输入员工手机号' },
              { pattern: /^1[3-9]\d{9}$/, message: '请输入正确的手机号' },
            ]">
              <template #label>
                <div class="flex items-center">
                  <div>员工手机号：</div>
                  <span v-if="!detail.activatedStatus" :class="detail.activatedStatus ? 'text-#06f' : 'text-#f03333'">
                    {{ detail.activatedStatus ? '已激活' : '未激活' }}
                    <a-popover placement="top">
                      <template #title>
                        <div style="border-bottom:1px solid #eee;padding-bottom: 6px;">说明</div>
                      </template>
                      <template #content>
                        <div class="w-300px">
                          <div>修改手机号码方式：</div>
                          <div>1.建议员工自己登录 App 点击"我的-个人信息-账号与安全"修改自己手机号码。</div>
                          <div>2.支持超级管理员修改员工手机号码。</div>
                        </div>
                      </template>
                      <QuestionCircleOutlined />
                    </a-popover>
                  </span>
                </div>
              </template>
              <div class="flex">
                <a-input v-model:value="formState.mobile" :disabled="true" placeholder="请输入" />
                <a-button type="link" @click="openChangeMobile">修改手机号</a-button>
              </div>
            </a-form-item>

            <!-- 员工类型 -->
            <a-form-item label="员工类型：" :required="true">
              <a-radio-group v-model:value="formState.userType" class="custom-radio">
                <a-radio :value="1">正式员工</a-radio>
                <a-radio :value="2">兼职员工</a-radio>
              </a-radio-group>
            </a-form-item>
          </div>
        </div>
      </a-form>

      <!-- 角色列表展示 -->
      <template v-if="safeSelectRoleList.length !== 0">
        <div class="border-1px border-solid border-gray-200 rounded-10px mt-10px">
          <div class="p15px">
            <custom-title title="总机构" font-size="16px" font-weight="500" />
          </div>
          <template v-for="roleInfo in roleInfos" :key="roleInfo.id">
            <div v-if="safeSelectRoleList.includes(roleInfo.id)" class="border-t-1px border-#eee border-solid border-0px">
              <div class="flex justify-between px-25px py-15px">
                <div class="flex items-start gap-10px">
                  <div>
                    <div class="font-bold">{{ roleInfo.roleName }}</div>
                    <div>
                      <span class="text-#06f font-500">{{ roleInfo.functionalAuthorityCount }}</span>
                      <span>个功能权限,</span>
                      <span class="text-#06f font-500">{{ roleInfo.dataAuthorityCount }}</span>
                      <span>个数据权限</span>
                    </div>
                    <div v-if="roleInfo.description" class="mt-4px">
                      <span>
                        <clamped-text :lines="2" class="text-#666 text-13px line-height-18px"
                          :text="'角色描述：' + roleInfo.description"></clamped-text>
                      </span>
                    </div>
                  </div>
                </div>
                <div
                  class="text-#06f cursor-pointer whitespace-nowrap text-12px px-8px max-h-25px border-#06f border-solid border-1px flex items-center justify-center rounded-15px text-center"
                  @click="roleDetailsFunc(roleInfo)">
                  <span class="whitespace-nowrap">角色权限详情</span>
                </div>
              </div>
            </div>
          </template>
        </div>
      </template>
    </div>
    <!-- 底部按钮 -->
    <template #footer>
      <a-button danger ghost @click="cancel" :disabled="submitLoading">关闭</a-button>
      <a-button type="primary" ghost @click="handleOk" :loading="submitLoading">确定</a-button>
    </template>

    <!-- 子组件 -->
    <selectRole v-model="addEmployeesOpen" :role-infos="roleInfos" :select-role-list="selectRoleList"
      @success="handleSelectRoleSuccess" />

    <!-- 角色权限详情抽屉 -->
    <roles-details-drawer v-model:open="roleDetailsOpen" :role-id="currentRoleId" :details="currentRoleDetails"
      @on-edit-success="handleRoleEditSuccess" />

    <!-- 修改手机号弹窗 -->
    <a-modal v-model:open="changeMobileOpen" title="修改手机号" :width="400" :destroy-on-close="true" :centered="true"
      :keyboard="false" :mask-closable="false">
      <a-form ref="changeMobileForm" :model="changeMobileFormState" layout="vertical">
        <a-form-item label="员工手机号：" name="mobile" required
          :validate-status="mobileValidationState.validateStatus"
          :help="mobileValidationState.help">
          <a-input v-model:value="changeMobileFormState.mobile" placeholder="请输入新手机号" 
            :maxlength="11" @input="handleMobileChange" />
        </a-form-item>
      </a-form>
      <template #footer>
        <a-button @click="closeChangeMobile" :disabled="changeMobileLoading">取消</a-button>
        <a-button type="primary" @click="submitChangeMobile" :loading="changeMobileLoading">确定</a-button>
      </template>
    </a-modal>
  </a-modal>
</template>

<style lang="less" scoped>
.contenter {
  padding: 24px;
}

/* 动画和交互效果 */
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

/* 头像上传区域样式 */
.upload-area {
  width: 80px;
  height: 80px;
  border-radius: 50%;
  background: #06f;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
}

/* 表单样式覆盖 */
:deep(.ant-form-item-explain-error) {
  font-size: 12px;
}

:deep(.ant-modal-body) {
  padding: 12px 24px;
}

/* 角色选择表单项样式 */
.position {
  margin-bottom: 0;

  :deep(.ant-form-item-control),
  :deep(.ant-form-item-control-input) {
    min-height: 1px;
    height: 1px;
  }
}

/* z-index 控制 */
:global(.z-index-1000) {
  z-index: 1000 !important;
}

:global(.z-index) {
  z-index: 800 !important;
}

/* 响应式样式 */
.responsive-modal {
  :deep(.ant-modal) {
    width: 800px !important;
    max-width: 95vw !important;
  }
}

@media (max-width: 560px) {
  .responsive-modal {
    :deep(.ant-modal) {
      width: 95vw !important;
      margin: 0 auto;
    }
  }

  .form-container {
    flex-direction: column !important;
    gap: 20px !important;
  }

  .form-left,
  .form-right {
    width: 100% !important;
    margin-top: 0 !important;
  }

  .form-right {
    margin-top: 0 !important;
  }
}
</style>
