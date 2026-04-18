<script setup>
import { CloseOutlined, QuestionCircleOutlined } from '@ant-design/icons-vue'
import dayjs from 'dayjs'
import editEmployees from './editEmployees.vue'
import ResignConfirmModal from './ResignConfirmModal.vue'
import { batchDisabledApi, getInstUserDetail } from '@/api/internal-manage/staff-manage'
import messageService from '~@/utils/messageService'
import { updateRoleApi } from '~@/api/internal-manage/role-manage'
import emitter, { EVENTS } from '~@/utils/eventBus'
import { useUserStore } from '~@/stores/user'
const { detail, departmentList } = defineProps({
  detail: {
    type: Object,
    default: () => ({}),
  },
  departmentList: {
    type: Array,
    default: () => ([]),
  },
})

const emit = defineEmits(['refresh-list'])

const open = defineModel({
  type: Boolean,
  default: false,
})

const editEmployeesOpen = ref(false)
const roleDetailsOpen = ref(false)
const selectedRoleId = ref(null)
const selectedRoleDetails = ref({})
const detailInfo = ref({})
const loading = ref(false)
const activeKey = ref('0')
// 角色选项（这里可以从props或API获取）
const roleInfos = ref([])

// 离职确认弹窗相关状态
const resignConfirmVisible = ref(false)
const resignLoading = ref(false)

const drawerRef = ref(null)

// 响应式宽度
const drawerWidth = ref('800px')

// 计算响应式宽度
const updateDrawerWidth = () => {
  const windowWidth = window.innerWidth
  if (windowWidth <= 768) {
    // 移动端 - 全屏
    drawerWidth.value = '100vw'
  } else if (windowWidth <= 1024) {
    // 平板 - 80%宽度，最大600px
    drawerWidth.value = Math.min(windowWidth * 0.8, 600) + 'px'
  } else if (windowWidth <= 1440) {
    // 小桌面 - 60%宽度，最大700px
    drawerWidth.value = Math.min(windowWidth * 0.6, 700) + 'px'
  } else {
    // 大桌面 - 固定800px
    drawerWidth.value = '800px'
  }
}

// 组件挂载时计算初始宽度并添加监听器
onMounted(() => {
  updateDrawerWidth()
  window.addEventListener('resize', updateDrawerWidth)
})

// 组件卸载时移除监听器
onUnmounted(() => {
  window.removeEventListener('resize', updateDrawerWidth)
})

function editEmployeesFunc() {
  editEmployeesOpen.value = true
}

function delEmployeesFunc() {
  if (detailInfo.value.disabled) {
    // 复职操作
    handleRehire()
  }
  else {
    // 离职操作 - 检查是否是超级管理员
    if (detailInfo.value.isAdmin) {
      messageService.error('超级管理员不能进行离职操作')
      return
    }
    resignConfirmVisible.value = true
  }
}

// 复职操作
async function handleRehire() {
  try {
    await batchDisabledApi({
      userIds: [detailInfo.value.id],
      isWork: false,
    })
    messageService.success('复职成功')
    getInstUserDetailFunc()
    // 通知父组件刷新列表
    emit('refresh-list')
  } catch (error) {
    messageService.error('复职失败')
    console.error('复职失败:', error)
  }
}

// 确认离职
async function handleResignConfirm() {
  try {
    resignLoading.value = true

    await batchDisabledApi({
      userIds: [detailInfo.value.id],
      isWork: true,
    })

    messageService.success('离职成功')
    resignConfirmVisible.value = false

    // 刷新员工详情
    getInstUserDetailFunc()
    // 通知父组件刷新列表
    emit('refresh-list')
  } catch (error) {
    messageService.error('离职失败')
    console.error('离职失败:', error)
  } finally {
    resignLoading.value = false
  }
}

// 取消离职
function handleResignCancel() {
  resignConfirmVisible.value = false
}

function roleDetailsDrawerFunc(roleId) {
  // 根据roleId找到对应的角色详情
  const roleInfo = roleInfos.value.find(role => role.roleId === roleId)
  selectedRoleId.value = roleId
  selectedRoleDetails.value = roleInfo || {}
  roleDetailsOpen.value = true
}

function handleAfterOpenChange(open) {
  if (open) {
    getInstUserDetailFunc()
  }
}

function getInstUserDetailFunc() {
  loading.value = true
  getInstUserDetail({ id: detail.id }).then(({ result }) => {
    detailInfo.value = result
    roleInfos.value = result.roles
  }).finally(() => {
    loading.value = false
  })
}
async function handleSaveFun(data) {
  try {
    const res = await updateRoleApi(data)
    if (res.code === 200) {
      messageService.success('更新成功')
      getInstUserDetailFunc()
    }

    // 关闭按钮loaidng
    emitter.emit(EVENTS.CLOSE_LOADING_EVENT)
  }
  catch (err) {
    console.log(err)
    messageService.error('创建失败')
  }
}
// 编辑员工的回调
function handleEditSuccess() {
  getInstUserDetailFunc()
  // 通知父组件刷新列表
  emit('refresh-list')
}
// 获取用户信息
const userStore = useUserStore()
// 计算属性获取机构名称
const orgName = computed(() => userStore.userInfo?.orgName || '总机构')

// 计算属性获取部门标题
const getDeptTitle = computed(() => {
  const filteredDepts = detailInfo.value.deptNames?.filter(name => name !== orgName.value)
  if (filteredDepts && filteredDepts.length > 0) {
    return `${orgName.value}（${filteredDepts.join('、')}）`
  }
  return orgName.value
})
</script>

<template>
  <div ref="drawerRef">
    <a-drawer v-model:open="open" :body-style="{ padding: '0' }" :closable="false" :width="drawerWidth"
      placement="right" :push="false" @after-open-change="handleAfterOpenChange">
      <!-- 自定义头部 -->
      <template #title>
        <div class="custom-header flex justify-between h-4 flex-items-center">
          <div class="text-5">
            员工详情
          </div>
          <a-button type="text" class="close-btn" @click="open = false">
            <template #icon>
              <CloseOutlined class="text-5 close-icon" />
            </template>
          </a-button>
        </div>
      </template>
      <div class="contenter">
        <a-spin :spinning="loading">
          <div class="p-24px">
            <div class="flex pl-16px overflow-hidden">
              <div class="flex flex-1">
                <div class="text-30px">
                  <a-avatar
                    style="width: 80px; height: 80px; line-height: 80px; font-size: 28px; background-color: rgb(0, 102, 255);"
                    :size="{ xs: 24, sm: 32, md: 40, lg: 64, xl: 80, xxl: 100 }">
                    <template #icon>
                      <div style="line-height: 80px;">
                        {{ detailInfo.nickName?.slice(0, 1) }}
                      </div>
                    </template>
                  </a-avatar>
                </div>
                <div class="flex flex-col ml-16px flex-1 justify-center">
                  <div class="staffName">
                    {{ detailInfo.nickName }}
                  </div>
                  <div class="staffPhone mb-2px">
                    <span>{{ detailInfo.mobile }}</span>
                    <div class="notActivation">
                      <span v-if="!detail.isAdmin && !detail.activatedStatus" class="text-#ff3333 ml2">未激活
                        <a-popover placement="top">
                          <template #title>
                            <div style="border-bottom:1px solid #eee;padding-bottom: 6px;">
                              说明
                            </div>
                          </template>
                          <template #content>
                            <div>未激活：当前员工未激活登录过系统，如手机号不正确，员</div>
                            <div>工自己可登录 App 点击“我的-个人信息-账号与安全”修</div>
                            <div>改手机号码或超级管理员修改员工手机号码。</div>
                          </template>
                          <QuestionCircleOutlined />
                        </a-popover>
                      </span>
                    </div>
                  </div>
                  <div class="flex">
                    <span class="working" :class="detailInfo.disabled ? 'text-#ff3333! bg-#ffe6e6!' : ''">{{
                      detailInfo.disabled
                        ? '已离职' : '在职中'
                    }}</span>
                  </div>
                </div>
              </div>
              <div class="ml-16px">
                <a-button :danger="!detailInfo.disabled" :ghost="!detailInfo.disabled" @click="delEmployeesFunc">
                  {{ detailInfo.disabled ? '复职' : '离职' }}
                </a-button>
                <a-button type="primary" class="ml-16px" @click="editEmployeesFunc">
                  编辑
                </a-button>
              </div>
            </div>
            <div class="detailInfo">
              <div class="itemInfo">
                创建时间：<span class="text-#888">{{ dayjs(detailInfo.createTime).format("YYYY-MM-DD HH:mm") }}</span>
              </div>
            </div>
          </div>
          <div class="tabs">
            <a-tabs v-model:active-key="activeKey" size="large" :tab-bar-style="{
              'border-radius': '0px', 'padding-left': '24px',
            }">
              <a-tab-pane key="0" tab="任职信息">
                <div class="mx-20px border-1px border-solid border-gray-200 rounded-10px">
                  <div class="p15px">
                    <custom-title :title="getDeptTitle" font-size="16px"
                      font-weight="500" />
                  </div>

                  <template v-for="roleInfo in roleInfos" :key="roleInfo.roleId">
                    <div class="border-t-1px border-#eee border-solid border-0px">
                      <div class="flex justify-between px-25px py-15px">
                        <div class="flex items-start gap-10px">
                          <div>
                            <div class="font-bold">
                              {{ roleInfo.roleName }}
                            </div>
                            <div>
                              <span class="text-#06f font-500">{{ roleInfo.functionalAuthorityCount || 0 }}</span>
                              <span>个功能权限,</span>
                              <span class="text-#06f font-500">{{ roleInfo.dataAuthorityCount || 0 }}</span>
                              <span>个数据权限</span>
                            </div>
                            <div v-if="roleInfo.description" class="mt-4px">
                              <span >
                                <clamped-text :lines="2" class="text-#666 text-13px line-height-18px"
                                  :text="'角色描述：' + roleInfo.description"></clamped-text>
                              </span>
                            </div>
                          </div>
                        </div>
                        <div
                          class="text-#06f cursor-pointer  text-12px px-8px max-h-25px border-#06f border-solid border-1px flex items-center justify-center rounded-15px text-center"
                          @click="roleDetailsDrawerFunc(roleInfo.roleId)">
                          <span class="whitespace-nowrap">角色权限详情</span>
                        </div>
                      </div>
                    </div>
                  </template>
                </div>
              </a-tab-pane>
            </a-tabs>
          </div>
        </a-spin>
      </div>
      <!-- 编辑员工弹窗 -->
      <editEmployees v-model="editEmployeesOpen" :detail="detail" :department-list="departmentList"
        @success="handleEditSuccess" @refresh-list="handleEditSuccess" />

      <!-- 离职确认弹窗 -->
      <ResignConfirmModal v-model:open="resignConfirmVisible" :employee-names="detailInfo.nickName" :employee-count="1"
        :loading="resignLoading" @confirm="handleResignConfirm" @cancel="handleResignCancel" />
      <!-- 角色权限详情抽屉 -->
      <roles-details-drawer v-model:open="roleDetailsOpen" :role-id="selectedRoleId" :details="selectedRoleDetails"
        @on-edit-success="handleSaveFun" />
    </a-drawer>
  </div>
</template>

<style lang="less" scoped>
.itemInfo {
  display: flex;
  align-items: center;
  margin-right: 50px;
}

.tabs {
  width: 100%;
  border-radius: 10px;
  line-height: 40px;

  :deep(.ant-tabs-nav) {
    border-radius: 16px;
    padding: 0 24px;
  }

  :deep(.ant-tabs-tab) {
    font-size: 16px;
  }

  :deep(.ant-tabs-ink-bar) {
    text-align: center;
    height: 9px !important;
    background: transparent;
    bottom: 1px !important;

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

}

.working {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 2px 8px;
  margin-right: 8px;
  border-radius: 10px;
  font-size: 12px;
  font-weight: 400;
  color: #06f;
  background: #e6f0ff;
}

.staffPhone {
  display: flex;
  align-items: center;
  height: 24px;
  font-size: 16px;
  font-weight: 400;
  color: #222;
}

.staffName {
  height: 28px;
  line-height: 28px;
  font-size: 20px;
  font-weight: 500;
  color: #222;
}

.detailInfo {
  display: flex;
  height: 22px;
  padding-left: 8px;
  margin-top: 16px;
  line-height: 22px;
  font-size: 14px;
  font-weight: 400;
  color: #222;
}

.notActivation {
  display: flex;
  align-items: center;
  padding: 2px 8px;
  font-size: 12px;
  font-weight: 400;
  color: #f33;
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
</style>
