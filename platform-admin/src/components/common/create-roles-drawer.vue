<script setup lang="ts">
import {
  CloseOutlined,
  DownOutlined,
  SearchOutlined,
  UpOutlined,
} from '@ant-design/icons-vue'
import { watch } from 'vue'
import { Empty } from 'ant-design-vue'
import { useDrawer } from '@/composables/useDrawer'
import {
  type Authority,
  type AuthorityChild,
  type AuthorityGroup,
  useRolePermissions,
} from '@/composables/useRolePermissions'
import {
  getDefaultRoleDetailApi,
  getFullMenuListApi,
} from '~@/api/internal-manage/role-manage'
import emitter, { EVENTS } from '~@/utils/eventBus'
import { useUserStore } from '~@/stores/user'
import { useQueryBreakpoints } from '@/composables/query-breakpoints'
import { useConsoleOwnType } from '@/utils/console-permission'

const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
  roleId: {
    type: Number,
    default: null,
  },
})

const emit = defineEmits(['update:open', 'onSuccess'])

const simpleImage = Empty.PRESENTED_IMAGE_SIMPLE
const consoleOwnType = useConsoleOwnType()

function getPermissionNodeId(node: any) {
  return Number(node?.menuId || node?.id || node?.menu?.id || 0)
}

// 详情接口返回的是已选权限树，叶子节点即需要回显选中的权限点
function getLastLevelCheckedIds(permissions: any[]): number[] {
  const result = new Set<number>()

  const traverse = (nodes: any[]) => {
    nodes.forEach((node) => {
      const children = Array.isArray(node?.children) ? node.children : []
      if (children.length > 0) {
        traverse(children)
        return
      }

      const nodeId = getPermissionNodeId(node)
      if (nodeId > 0)
        result.add(nodeId)
    })
  }

  traverse(Array.isArray(permissions) ? permissions : [])
  return Array.from(result)
}
const btnLoading = ref(false)
// 使用抽屉状态管理
const { openDrawer } = useDrawer(props, emit)
const loading = ref(false)
// 权限数据
const initialData: AuthorityGroup[] = []

// 使用权限管理 hook
const {
  boxList,
  formState,
  searchValue,
  filteredBoxList,
  isAllExpanded,
  isParentExpanded,
  isChildExpanded,
  toggleAllExpand,
  expandAllChildren,
  collapseAllChildren,
  toggleChildExpand,
  getFilteredChildren,
  getFilteredchildren,
  highlightText,
  handleParentChange: originalHandleParentChange,
  handleChildChange: originalHandleChildChange,
  handleAuthorityChange: originalHandleAuthorityChange,
  clearAllSelected,
  resetAllStates,
  updateData,
  setDefaultCheckedByIds,
} = useRolePermissions(initialData, loading)

// 包装权限处理函数，添加验证清除逻辑
function handleParentChange(item: AuthorityGroup) {
  originalHandleParentChange(item)
  clearPermissionValidation()
}

function handleChildChange(child: AuthorityChild, parent: AuthorityGroup) {
  originalHandleChildChange(child, parent)
  clearPermissionValidation()
}

function handleAuthorityChange(authority: Authority, child: AuthorityChild, parent: AuthorityGroup) {
  originalHandleAuthorityChange(authority, child, parent)
  clearPermissionValidation()
}

// 包装清空已选函数
function handleClearAllSelected() {
  clearAllSelected()
  clearPermissionValidation()
}
async function getMenuList() {
  loading.value = true
  try {
    const res = await getFullMenuListApi({ ownType: consoleOwnType.value })
    // console.log("API原始数据:", res);
    if (res.code === 200) {
      // 递归给数据添加checked和indeterminate属性
      const processMenuData = (data: any[]): any[] => {
        if (!data || !Array.isArray(data))
          return []

        return data.map((item) => {
          const isLeafNode = !item.children || item.children.length === 0

          const processedItem = {
            ...item,
            // 确保基本字段存在
            id: item.id || item.menuId || '',
            menuName: item.menuName || item.name || '',
            checked: false,
            // 如果是叶子节点，确保有权限相关字段
            name: item.menuName || '',
            remark: item.introduce || '',
            type: item.type || 1,
            mode: item.mode || 0,
            groupCode: item.groupCode || '',
            weight: item.weight || 0,
          }

          // 非叶子节点才添加indeterminate属性
          if (!isLeafNode) {
            processedItem.indeterminate = false
          }

          if (item.children && item.children.length > 0) {
            processedItem.children = processMenuData(item.children)
          }
          else {
            // 确保children是数组
            processedItem.children = []
          }

          return processedItem
        })
      }

      const processedData = processMenuData(res.result)
      // console.log("处理后的数据:", processedData);
      updateData(processedData)
      if (props.roleId) {
        // 根据角色id查询详情，获取最后一级权限id
        await getRoleDetail()
      }
      else {
        formState.roleId = null
      }
    }
  }
  catch (error) {
    console.log(error)
  }
  finally {
    loading.value = false
  }
}
async function getRoleDetail() {
  try {
    const res = await getDefaultRoleDetailApi({ roleId: props.roleId })
    if (res.code === 200) {
      // console.log(res.result);
      //  赋值 表单
      formState.roleId = res.result.roleId
      formState.roleName = res.result.roleName
      formState.description = res.result.description
      // 获取最后一级权限id
      const lastLevelPermissionIds = getLastLevelCheckedIds(
        res.result.menuIds || [],
      )
      // console.log("最后一级checked权限IDs:", lastLevelPermissionIds);
      // 赋值 权限
      formState.menuIds = lastLevelPermissionIds
      // 设置权限树的选中状态
      setDefaultCheckedByIds(lastLevelPermissionIds)
    }
  }
  catch (error) {
    console.log(error)
  }
}
// 监听抽屉关闭，重置所有状态
watch(openDrawer, (newVal) => {
  if (!newVal) {
    resetAllStates()
  }
  else {
    // 获取所有权限列表
    getMenuList()
  }
})
const formRef = ref(null)

// 权限验证函数
function validatePermissions() {
  const hasSelectedPermissions = boxList.value.some(
    parent =>
      parent.checked
      || parent.indeterminate
      || (parent.children
        && parent.children.some(
          child =>
            child.checked
            || child.indeterminate
            || (child.children
              && child.children.some(authority => authority.checked)),
        )),
  )

  if (!hasSelectedPermissions) {
    return Promise.reject('请选择功能与权限')
  }
  return Promise.resolve()
}

// 清除权限验证错误
function clearPermissionValidation() {
  if (formRef.value) {
    formRef.value.clearValidate(['menuIds'])
  }
}
// 保存
async function handleSave() {
  btnLoading.value = true
  try {
    // 1. 使用表单验证进行所有校验（包括角色名称和权限选择）
    await formRef.value.validate()

    // 2. 收集所有选中和半选中的权限ID
    const menuIds: number[] = []

    boxList.value.forEach((parent) => {
      // 收集一级权限ID（选中或半选中）
      if (parent.checked || parent.indeterminate) {
        // 注意：这里假设一级权限也有数字ID，如果没有可以跳过
        menuIds.push(Number(parent.id))
      }

      if (parent.children) {
        parent.children.forEach((child) => {
          // 收集二级权限ID（选中或半选中）
          if (child.checked || child.indeterminate) {
            // 注意：这里假设二级权限也有数字ID，如果没有可以跳过
            menuIds.push(Number(child.id))
          }

          // 收集三级权限ID（只收集选中的）
          if (child.children) {
            child.children.forEach((authority) => {
              if (authority.checked) {
                menuIds.push(authority.id)
              }
            })
          }
        })
      }
    })

    // 3. 更新formState
    formState.menuIds = menuIds

    // 4. 打印结果供调试
    // console.log("表单数据：", formState);
    // console.log("选中的权限ID列表：", menuIds);
    // 5. 这里可以调用API保存数据
    emit('onSuccess', formState)
  }
  catch (error) {
    console.log('表单验证失败:', error)
    btnLoading.value = false
    // Ant Design Vue 的表单验证失败时会自动显示错误信息，无需手动处理
  }
}
// 监听关闭loading事件
onMounted(() => {
  emitter.on(EVENTS.CLOSE_LOADING_EVENT, () => {
    btnLoading.value = false
  })
})
// 组件卸载时移除事件监听
onUnmounted(() => {
  emitter.off(EVENTS.CLOSE_LOADING_EVENT)
})

// 获取用户信息
const userStore = useUserStore()
// 计算属性获取机构名称
const orgName = computed(() => userStore.userInfo?.orgName || '总机构')

// 响应式布局
const { isMobile, isPad, isDesktop } = useQueryBreakpoints()

// 响应式抽屉宽度
const drawerWidth = computed(() => {
  if (isMobile.value) {
    return '100%'
  } else if (isPad.value) {
    return '90%'
  } else {
    return '800px'
  }
})
</script>

<template>
  <div>
    <a-drawer v-model:open="openDrawer" :push="{ distance: isMobile ? 0 : 80 }"
      :body-style="{ padding: '0', background: '#f7f7fd' }" :closable="false" :width="drawerWidth" placement="right">
      <!-- 自定义头部 -->
      <template #title>
        <div class="custom-header flex justify-between h-4 flex-items-center">
          <div :class="isMobile ? 'text-4' : 'text-5'">
            {{ roleId ? "编辑角色" : "新建角色" }}
          </div>
          <a-button type="text" class="close-btn" @click="openDrawer = false">
            <template #icon>
              <CloseOutlined :class="isMobile ? 'text-4' : 'text-5'" class="close-icon" />
            </template>
          </a-button>
        </div>
      </template>
      <div class="contenter scrollbar">
        <div class="text-14px text-#222 font-400">
          所属机构：{{ orgName }}
        </div>
        <!-- 分割线 -->
        <a-divider />
        <!-- 角色名称  角色描述 -->
        <a-form ref="formRef" :model="formState">
          <div :class="isMobile ? 'block' : 'flex'">
            <a-form-item label="角色名称" name="roleName" :label-col="{ span: 24 }"
              :class="isMobile ? 'mb-4' : 'flex-1 mr-100px'" :rules="[{ required: true, message: '请输入角色名称' }]">
              <a-input v-model:value="formState.roleName" placeholder="请输入角色名称" />
            </a-form-item>
            <a-form-item label="角色描述" name="description" :label-col="{ span: 24 }" :class="isMobile ? '' : 'flex-1'">
              <a-input v-model:value="formState.description" placeholder="请输入角色描述" />
            </a-form-item>
          </div>
          <a-form-item label="功能与权限" class="permissions" name="menuIds" :rules="[
            {
              required: true,
              validator: validatePermissions,
              message: '请选择功能与权限',
            },
          ]">
            <div class="flex justify-end items-center">
              <a-button type="link" @click="toggleAllExpand">
                {{ isAllExpanded ? "一键收起" : "一键展开" }}
                <component :is="isAllExpanded ? UpOutlined : DownOutlined" />
              </a-button>
            </div>
          </a-form-item>
          <a-form-item class="relative top--12px ml--5px">
            <div class="flex justify-between items-center">
              <a-button type="link" :disabled="!filteredBoxList.some(
                (item) => item.checked || item.indeterminate,
              )
                " @click="handleClearAllSelected">
                清空已选
              </a-button>
              <a-input v-model:value="searchValue" placeholder="搜索权限点名称或权限描述" allow-clear>
                <template #prefix>
                  <SearchOutlined />
                </template>
              </a-input>
            </div>
            <div class="box">
              <a-spin :spinning="loading" tip="加载中...">
                <div v-for="(item, index) in filteredBoxList" :key="item.id">
                  <div class="text-14px text-#222 font-400 h-40px flex items-center justify-between shadow-box" :class="{
                    'last-child': index === filteredBoxList.length - 1,
                  }">
                    <div class="pl-16px">
                      <a-checkbox v-model:checked="item.checked" :indeterminate="item.indeterminate"
                        @change="() => handleParentChange(item)" />
                      <span class="ml-8px font-500" v-html="highlightText(item.menuName, searchValue)" />
                    </div>
                    <div class="pr-16px">
                      <span
                        class="block text-#666 text-12px border-1px border-#ddd border-solid cursor-pointer rounded-14px px-10px py-1px"
                        @click="
                          isParentExpanded(item.id)
                            ? collapseAllChildren(item.id)
                            : expandAllChildren(item.id)
                          ">
                        {{
                          isParentExpanded(item.id) ? "收起全部" : "展开全部"
                        }}
                      </span>
                    </div>
                  </div>
                  <template v-if="isParentExpanded(item.id)">
                    <div v-for="child in getFilteredChildren(item)" :key="child.id">
                      <div
                        class="text-14px text-#222 font-400 h-40px flex items-center justify-between shadow-box pl-38px">
                        <div>
                          <a-checkbox v-model:checked="child.checked" :indeterminate="child.indeterminate"
                            @change="() => handleChildChange(child, item)" />
                          <span class="ml-8px font-500" v-html="highlightText(child.menuName, searchValue)" />
                        </div>
                        <div class="pr-16px">
                          <span
                            class="block text-#666 text-12px border-1px border-#ddd border-solid cursor-pointer rounded-14px px-10px py-1px"
                            @click="toggleChildExpand(child.id)">
                            {{ isChildExpanded(child.id) ? "收起" : "展开" }}
                          </span>
                        </div>
                      </div>
                      <template v-if="isChildExpanded(child.id)">
                        <div v-for="(authority, idx) in getFilteredchildren(
                          child,
                          item,
                        )" :key="authority.id">
                          <div
                            class="text-12px text-#222 font-400 min-h-58px py-6px flex items-center justify-between shadow-box bg-#fcfcfc"
                            :class="{
                              'last-child':
                                idx
                                === getFilteredchildren(child, item).length - 1,
                            }">
                            <div class="flex items-center pl-60px">
                              <a-checkbox v-model:checked="authority.checked" @change="
                                () =>
                                  handleAuthorityChange(
                                    authority,
                                    child,
                                    item,
                                  )
                              " />
                              <div class="flex flex-col">
                                <span class="ml-8px text-13px" v-html="highlightText(authority.name, searchValue)
                                  " />
                                <span class="ml-8px text-12px text-#888 pr-120px" v-html="highlightText(authority.remark, searchValue)
                                  " />
                              </div>
                            </div>
                          </div>
                        </div>
                      </template>
                    </div>
                  </template>
                </div>
                <a-empty v-if="filteredBoxList.length === 0" :image="simpleImage" />
              </a-spin>
            </div>
          </a-form-item>
        </a-form>
      </div>
      <template #footer>
        <div :class="isMobile ? 'p-4' : 'flex justify-end'">
          <a-button type="primary" :class="isMobile ? 'w-full h-44px' : 'w-100px h-40px'" :loading="btnLoading"
            @click="handleSave">
            保存
          </a-button>
        </div>
      </template>
    </a-drawer>
  </div>
</template>

<style lang="less" scoped>
.contenter {
  padding: 24px;
  background-color: #fff;

  :deep(.ant-form-item-label) {
    padding-bottom: 0;
  }

  .box {
    margin-top: 10px;
    border-radius: 8px;
    border: 1px solid #ddd;
    margin-left: 15px;
  }

  .shadow-box {
    box-shadow: inset 0 -1px 0 0 #eee;
  }

  .last-child {
    box-shadow: none;
    border-radius: 0 0 8px 8px;
  }

  .permissions {
    :deep(.ant-form-show-help) {
      position: absolute;
      top: 5px;
      left: 160px;
      width: 50%;
    }
  }

  // 响应式样式
  @media (max-width: 767px) {
    padding: 16px;

    .box {
      margin-left: 0;
    }

    .permissions {
      :deep(.ant-form-show-help) {
        position: static;
        width: 100%;
      }
    }

    // 调整权限树的内边距
    .shadow-box {
      &.pl-16px {
        padding-left: 12px;
      }

      &.pl-38px {
        padding-left: 28px;
      }

      &.pl-60px {
        padding-left: 40px;
      }
    }
  }

  @media (max-width: 575px) {
    padding: 12px;

    // 进一步减小字体大小
    .text-14px {
      font-size: 13px;
    }

    .text-12px {
      font-size: 11px;
    }
  }
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

/* 角色模板搜索框样式 */
:deep(.template-search-input) {
  .ant-input {
    border-color: #d9d9d9 !important;
    box-shadow: none !important;

    &:hover {
      border-color: #40a9ff !important;
    }

    &:focus {
      border-color: #40a9ff !important;
      box-shadow: 0 0 0 2px rgba(24, 144, 255, 0.2) !important;
    }
  }

  /* 强制覆盖任何错误状态样式 */
  &.ant-input-status-error .ant-input,
  &.ant-input-affix-wrapper-status-error,
  .ant-input-affix-wrapper-status-error .ant-input {
    border-color: #d9d9d9 !important;
    box-shadow: none !important;
    background-color: #fff !important;
  }
}

/* 确保下拉框容器不受表单验证影响 */
.role-template-dropdown {

  /* 重置任何可能继承的表单验证样式 */
  .ant-input-affix-wrapper {
    border-color: #d9d9d9 !important;
    box-shadow: none !important;

    &:hover {
      border-color: #40a9ff !important;
    }

    &:focus,
    &:focus-within {
      border-color: #40a9ff !important;
      box-shadow: 0 0 0 2px rgba(24, 144, 255, 0.2) !important;
    }
  }
}
</style>
