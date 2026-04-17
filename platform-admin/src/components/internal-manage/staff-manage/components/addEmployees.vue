<!-- 新增员工 -->
<script setup>
import { CloseOutlined } from '@ant-design/icons-vue'
import { TreeSelect, message } from 'ant-design-vue'
import selectRole from './selectRole.vue'
import { saveInstUser } from '@/api/internal-manage/staff-manage'
import { getInstRolePageApi, updateRoleApi } from '@/api/internal-manage/role-manage'
import RolesDetailsDrawer from '@/components/common/roles-details-drawer.vue'
import messageService from '~@/utils/messageService'
import emitter, { EVENTS } from '~@/utils/eventBus'
import { useThrottleFn } from '@vueuse/core'

const props = defineProps({
  departmentList: {
    type: Array,
    default: () => [],
  },
})
const emit = defineEmits(['success'])
const open = defineModel({
  type: Boolean,
  default: false,
})

const addEmployeesOpen = ref(false)
const roleDetailsOpen = ref(false)
const currentRoleId = ref(null)
const currentRoleDetails = ref({})
const confirmLoading = ref(false)

const roleInfos = ref()

// 选择的角色
const selectRoleList = ref([])

const form = ref(null)

// 部门选项（这里可以从props或API获取）

const formState = reactive({
  nickName: '', // 员工姓名
  mobile: '', // 员工手机号
  deptIds: [], // 所属部门 - 初始为空，将在modal打开时设置
  disabled: 0, // 是否禁用,
  userType: '1', // 1: 正式员工 2: 兼职员工
  avatar: '', // 头像
  roleIds: [], // 角色
})

function roleDetailsFunc(roleInfo) {
  currentRoleId.value = roleInfo.id
  currentRoleDetails.value = roleInfo
  roleDetailsOpen.value = true
}

async function getRoleList(query = { queryModel: {} }) {
  try {
    const pages = {
      pageRequestModel: {
        needTotal: true,
        pageSize: 500,
        pageIndex: 1,
        skipCount: 1,
      },
    }
    const res = await getInstRolePageApi({ ...pages, ...query })
    if (res.code === 200) {
      roleInfos.value = res.result
    } else {
      messageService.error('获取角色列表失败')
    }
  } catch (error) {
    console.error('获取角色列表失败:', error)
    // messageService.error('获取角色列表失败')
  }
}

// 添加防抖处理
const handleOk = useThrottleFn(async () => {
  if (confirmLoading.value) return
  try {
    const valid = await form.value.validate()
    if (valid) {
      confirmLoading.value = true
      console.log('提交数据:', formState)
      const res = await saveInstUser({ ...formState })
      
      if (res.code === 200) {
        // 重置表单
        form.value.resetFields()
        message.success('新增员工成功')
        emit('success')
        // 重置角色选择状态
        selectRoleList.value = []
        formState.roleIds = []
        // 关闭模态框
        open.value = false
      } else {
        messageService.error(res.message || '新增员工失败')
      }
    }
  } catch (error) {
    console.log('新增员工失败:', error)
  } finally {
    confirmLoading.value = false
  }
}, 1000)

// 处理选择角色成功的逻辑
function handleSelectRoleSuccess(roleIds) {
  selectRoleList.value = roleIds
  formState.roleIds = roleIds
  
  // 清除任职角色的校验错误
  if (roleIds && roleIds.length > 0) {
    form.value.clearValidate('roleIds')
  }
}

// 更新角色列表
function updateRoleInfos(newRoleInfos) {
  roleInfos.value = newRoleInfos
}

// 角色权限编辑成功回调
async function handleRoleEditSuccess(data) {
  try {
    const res = await updateRoleApi(data)
    if (res.code === 200) {
      messageService.success('角色权限更新成功')
      // 刷新角色列表
      await getRoleList()
    } else {
      messageService.error(res.message || '角色权限更新失败')
    }

    // 关闭按钮loading
    emitter.emit(EVENTS.CLOSE_LOADING_EVENT)
  }
  catch (err) {
    console.error('更新角色失败:', err)
    messageService.error('角色权限更新失败')
  }
}

function cancel() {
  open.value = false
  form.value.resetFields()
  // 重置角色选择状态
  selectRoleList.value = []
  formState.roleIds = []
}

// 头像上传处理
function handleAvatarChange(info) {
  // 处理头像上传逻辑
  console.log('头像上传:', info)
}

// 监听modal打开状态
watch(() => open.value, (newVal) => {
  if (newVal && props.departmentList.length > 0) {
    // 当modal打开且部门列表有数据时，设置默认选中的部门为第一个部门（顶级部门）
    formState.deptIds = [props.departmentList[0]?.id]
  }
  // 当modal关闭时，重置角色选择状态
  if (!newVal) {
    selectRoleList.value = []
    formState.roleIds = []
  }
})

// 监听部门列表变化
watch(() => props.departmentList, (newList) => {
  if (open.value && newList.length > 0) {
    // 如果modal已打开且部门列表有变化，设置默认部门
    formState.deptIds = [newList[0]?.id]
  }
}, { immediate: true })

onMounted(async () => {
  await getRoleList()
})
</script>

<template>
  <a-modal
    v-model:open="open" :mask-closable="false" :keyboard="false" width="800px" :destroy-on-close="true"
    :closable="false" :centered="true" @cancel="cancel"
    @ok="handleOk" :confirm-loading="confirmLoading" class="responsive-modal"
  >
    <template #title>
      <div class=" text-5 flex justify-between flex-center">
        <span>新建员工</span>
        <a-button type="text" class="close-btn" @click="cancel">
          <template #icon>
            <CloseOutlined class="text-5 close-icon" />
          </template>
        </a-button>
      </div>
    </template>
    <div class="overflow-auto pb-20px scrollbar">
      <a-form ref="form" :model="formState" autocomplete="off" layout="vertical">
        <div class="form-container flex justify-between w-100% gap-30px">
          <div class="form-left w-50%">
            <!-- 头像上传 -->
            <a-form-item>
              <div class="flex items-center gap-20px">
                <div class="upload-area">
                  <div class="text-#fff text-20px font-medium">
                    {{ formState.nickName ? formState.nickName[0] : '-' }}
                  </div>
                </div>
                <a-upload
                  name="avatar" list-type="picture-circle" :show-upload-list="false"
                  @change="handleAvatarChange"
                >
                  <a-button>
                    上传照片
                  </a-button>
                </a-upload>
              </div>
            </a-form-item>
            <!-- 员工姓名 -->
            <a-form-item label="员工姓名：" name="nickName" :rules="[{ required: true, message: '请输入员工姓名' }]">
              <a-input v-model:value="formState.nickName" placeholder="请输入" />
            </a-form-item>
            <!-- 所属部门 -->
            <a-form-item label="所属部门：" name="deptIds" :rules="[{ required: true, message: '请选择所属部门' }]">
              <a-tree-select
                v-model:value="formState.deptIds" style="width: 100%"
                :dropdown-style="{ maxHeight: '400px', overflow: 'auto' }" placeholder="请选择" allow-clear multiple
                tree-default-expand-all :tree-data="departmentList" tree-node-filter-prop="departName"
                :field-names="{ children: 'children', label: 'departName', value: 'id' }"
                :show-checked-strategy="TreeSelect.SHOW_ALL"
              />
            </a-form-item>
            <!-- 任职角色 -->
            <a-form-item
              style="margin: 0;" class="position" name="roleIds"
              :rules="[{ required: true, message: '请设置任职角色' }]"
            >
              <template #label>
                <div class="flex items-center">
                  <div>任职角色：</div>
                  <a-button type="primary" @click="addEmployeesOpen = true">
                    编辑
                  </a-button>
                </div>
              </template>
            </a-form-item>
          </div>
          <div class="form-right w-50% mt-18px mr-20px">
            <!-- 账号状态 -->
            <a-form-item label="账号状态：" :required="true">
              <a-radio-group v-model:value="formState.disabled" class="custom-radio">
                <a-radio :value="0">
                  在职中
                </a-radio>
                <!-- <a-radio value="1">
                  离职
                </a-radio> -->
              </a-radio-group>
            </a-form-item>
            <!-- 员工手机号 -->
            <a-form-item
              label="员工手机号：" name="mobile" :rules="[
                { required: true, message: '请输入员工手机号' },
                { pattern: /^1[3-9]\d{9}$/, message: '请输入正确的手机号' },
              ]"
            >
              <a-input v-model:value="formState.mobile" :maxlength="11" placeholder="请输入" />
            </a-form-item>
            <!-- 员工类型 -->
            <a-form-item label="员工类型：" :required="true">
              <a-radio-group v-model:value="formState.userType" class="custom-radio">
                <a-radio value="1">
                  正式员工
                </a-radio>
                <a-radio value="2">
                  兼职员工
                </a-radio>
              </a-radio-group>
            </a-form-item>
          </div>
        </div>
      </a-form>
      <template v-if="selectRoleList.length !== 0">
        <div class="border-1px border-solid border-gray-200 rounded-10px mr-20px mt-20px">
          <div class="p15px">
            <custom-title title="总机构" font-size="16px" font-weight="500" />
          </div>

          <template v-for="roleInfo in roleInfos" :key="roleInfo.id">
            <div v-if="selectRoleList.includes(roleInfo.id)" class="border-t-1px border-#eee border-solid border-0px">
              <div class="flex justify-between px-25px py-15px">
                <div class="flex items-start gap-10px">
                  <div>
                    <div class="font-bold">
                      {{ roleInfo.roleName }}
                    </div>
                    <div>
                      <span class="text-#06f font-500">{{ roleInfo.functionalAuthorityCount }}</span>
                      <span>个功能权限,</span>
                      <span class="text-#06f font-500">{{ roleInfo.dataAuthorityCount || 1 }}</span>
                      <span>个数据权限</span>
                    </div>
                    <div v-if="roleInfo.description">
                      <span class="text-#666">角色描述：{{ roleInfo.description }}</span>
                    </div>
                  </div>
                </div>
                <div
                  class="text-#06f cursor-pointer text-12px px-8px max-h-25px border-#06f border-solid border-1px flex items-center justify-center rounded-15px text-center"
                  @click="roleDetailsFunc(roleInfo)"
                >
                  角色权限详情
                </div>
              </div>
            </div>
          </template>
        </div>
      </template>
    </div>
    <selectRole
      v-model="addEmployeesOpen" 
      :role-infos="roleInfos" 
      :select-role-list="selectRoleList"
      @success="handleSelectRoleSuccess"
      @update:roleInfos="updateRoleInfos"
    />
    <!-- 角色权限详情抽屉 -->
    <roles-details-drawer 
      v-model:open="roleDetailsOpen" 
      :role-id="currentRoleId" 
      :details="currentRoleDetails"
      @on-edit-success="handleRoleEditSuccess" 
    />
  </a-modal>
</template>

<style>
.z-index {
  z-index: 800 !important;
}
</style>

<style scoped lang="less">
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

:deep(.ant-form-item-explain-error) {
  font-size: 12px;
}

:deep(.ant-modal-body) {
  padding: 12px 24px;
}

.position {
  margin-bottom: 0;
}

:deep(.position .ant-form-item-control) {
  min-height: 1px;
  height: 1px;
}

:deep(.position .ant-form-item-control-input) {
  min-height: 1px;
  height: 1px;
}

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

:deep(.modal-wrap .ant-modal) {
  top: 30px;
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
    margin-right: 0 !important;
  }

  .form-right {
    margin-top: 0 !important;
    margin-right: 0 !important;
  }
}
</style>
