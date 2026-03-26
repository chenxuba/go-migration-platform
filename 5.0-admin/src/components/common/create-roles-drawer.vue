<script setup lang="ts">
import {
  CloseOutlined,
  DownOutlined,
  QuestionCircleOutlined,
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
  getMenuListApi,
  getDefaultRole,
} from '~@/api/internal-manage/role-manage'
import emitter, { EVENTS } from '~@/utils/eventBus'
import { useUserStore } from '~@/stores/user'
import { useQueryBreakpoints } from '@/composables/query-breakpoints'

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

// 递归获取最后一级checked为true的权限ID
function getLastLevelCheckedIds(permissions: any[]): number[] {
  const result: number[] = []

  const traverse = (nodes: any[]) => {
    nodes.forEach((node) => {
      // 如果没有children或children为空数组，说明是最后一级
      if (!node.children || node.children.length === 0) {
        // 如果当前节点checked为true，添加到结果中
        if (node.checked === true) {
          result.push(node.menuId)
        }
      }
      else {
        // 如果有子节点，继续递归遍历
        traverse(node.children)
      }
    })
  }

  traverse(permissions)
  return result
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
    const res = await getMenuListApi({ ownType: 'INSTITUTION' })
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
        //  根据角色id查询详情，获取最后一级权限id
        loading.value = true
        getRoleDetail()
      }
      else {
        formState.roleId = null
      }
    }
  }
  catch (error) {
    console.log(error)
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
    // 获取角色模板列表
    getRoleTemplateList()
  }
})
const formRef = ref(null)

// 权限验证函数
function validatePermissions() {
  // 防止在角色模板选择过程中触发验证
  if (templateDropdownVisible.value) {
    return Promise.resolve()
  }

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

// 角色模板数据
const roleTemplates = ref([])

// 角色模板加载状态
const templatesLoading = ref(false)

// 角色模板搜索关键词
const templateSearchValue = ref('')

// 下拉框显示状态
const templateDropdownVisible = ref(false)

// 过滤后的角色模板
const filteredRoleTemplates = computed(() => {
  if (!templateSearchValue.value) {
    return roleTemplates.value
  }
  return roleTemplates.value.filter(template =>
    template.label.toLowerCase().includes(templateSearchValue.value.toLowerCase())
  )
})

// 获取角色模板列表
async function getRoleTemplateList() {
  try {
    templatesLoading.value = true
    const res = await getDefaultRole({
      pageSize: 500,
      pageIndex: 1
    })

    if (res.code === 200) {
      // 转换API数据格式为组件所需格式
      roleTemplates.value = (res.result || []).map((item, index) => ({
        key: item.id || `role_${index}`,
        label: `${item.roleName || item.name || '未命名角色'}${item.isDefault ? '（系统默认）' : ''}`,
        isSystem: item.isDefault || false,
        checked: false,
        roleId: item.id,
        description: item.description,
        menuIds: item.roleIds || []
      }))
    }
  } catch (error) {
    console.error('获取角色模板失败:', error)
  } finally {
    templatesLoading.value = false
  }
}

// 获取当前权限树中已选中的权限ID
function getCurrentSelectedMenuIds(): number[] {
  const currentMenuIds: number[] = []

  boxList.value.forEach((parent) => {
    if (parent.children) {
      parent.children.forEach((child) => {
        if (child.children) {
          child.children.forEach((authority) => {
            if (authority.checked) {
              currentMenuIds.push(authority.id)
            }
          })
        }
      })
    }
  })

  return currentMenuIds
}

// 处理角色模板复选框变化
function handleTemplateCheckChange(templateKey: string, checked: boolean) {
  const template = roleTemplates.value.find(t => t.key === templateKey)
  if (template) {
    template.checked = checked
  }
}

// 确定选择角色模板
async function handleConfirmTemplateSelect() {
  const selectedTemplates = roleTemplates.value.filter(t => t.checked)
  // console.log('选择的角色模板:', selectedTemplates)

  // 获取当前已选中的权限ID（从权限树中获取）
  const existingMenuIds = new Set<number>(getCurrentSelectedMenuIds())
  // console.log('现有权限ID:', Array.from(existingMenuIds))

  // 收集新选中模板的权限ID
  const newSelectedMenuIds = new Set<number>()

  selectedTemplates.forEach(template => {
    if (template.menuIds && Array.isArray(template.menuIds)) {
      template.menuIds.forEach(menuId => {
        if (typeof menuId === 'number') {
          newSelectedMenuIds.add(menuId)
        }
      })
    }
  })

  // console.log('新增权限ID:', Array.from(newSelectedMenuIds))

  // 合并现有权限和新权限（使用Set自动去重）
  const allMenuIds = new Set<number>([...existingMenuIds, ...newSelectedMenuIds])
  const finalMenuIdsArray = Array.from(allMenuIds)

  // console.log('合并后的权限ID:', finalMenuIdsArray)

  if (finalMenuIdsArray.length > 0) {
    // 更新表单状态
    formState.menuIds = finalMenuIdsArray

    // 设置权限树的选中状态（在现有基础上新增）
    setDefaultCheckedByIds(finalMenuIdsArray)

    console.log('应用模板权限 - 总计:', finalMenuIdsArray.length, '个权限')
  }

  // 先关闭下拉框，再清除验证错误
  templateDropdownVisible.value = false

  // 等待下一个tick后清除权限验证错误
  nextTick(() => {
    clearPermissionValidation()
  })

  // 重置搜索关键词
  templateSearchValue.value = ''

  // 重置所有角色模板的选中状态
  roleTemplates.value.forEach(template => {
    template.checked = false
  })
}

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
            <div class="flex justify-between justify-between items-center">
              <div>
                <a-dropdown v-model:open="templateDropdownVisible" >
                  <template #overlay>
                    <div class="role-template-dropdown "
                      style="width: 220px; background: white; border-radius: 8px; box-shadow: 0 2px 8px rgba(0,0,0,0.15);"
                      @click.stop>
                      <!-- 搜索框 -->
                      <div class="p-12px border-b border-#f0f0f0">
                        <a-form-item-rest>
                          <a-input v-model:value="templateSearchValue" placeholder="请输入角色名称" @click.stop>
                            <template #prefix>
                              <SearchOutlined class="text-#bfbfbf" />
                            </template>
                          </a-input>
                        </a-form-item-rest>
                      </div>

                      <!-- 角色列表 -->
                      <div class="max-h-200px overflow-y-auto scrollbar">
                        <template v-if="templatesLoading">
                          <div class="p-16px text-center text-#666">
                            <a-spin size="small" />
                            <span class="ml-8px">加载角色模板中...</span>
                          </div>
                        </template>
                        <template v-else-if="filteredRoleTemplates.length > 0">
                          <div v-for="template in filteredRoleTemplates" :key="template.key"
                            class="p-8px px-12px hover:bg-#f5f5f5 cursor-pointer flex items-center" @click.stop>
                            <a-form-item-rest>
                              <a-checkbox :checked="template.checked"
                                @change="(e) => handleTemplateCheckChange(template.key, e.target.checked)" @click.stop
                                class="mr-8px" />
                            </a-form-item-rest>
                            <div class="flex flex-col flex-1">
                              <span class="text-14px text-#333">{{ template.label }}</span>
                              <span v-if="template.description" class="text-12px text-#999 mt-2px">{{
                                template.description }}</span>
                            </div>
                          </div>
                        </template>
                        <template v-else>
                          <div class="p-16px text-center text-#999">
                            暂无角色模板
                          </div>
                        </template>
                      </div>

                      <!-- 确定按钮 -->
                      <div class="p-12px border-t border-#f0f0f0">
                        <a-button type="primary" block @click="handleConfirmTemplateSelect">
                          确定
                        </a-button>
                      </div>
                    </div>
                  </template>
                  <a-button type="primary" ghost :style="{ width: isMobile ? '140px' : '120px', borderRadius: '8px' }">
                    使用角色模版
                  </a-button>
                </a-dropdown>
                <a-popover title="说明">
                  <template #content>
                    <div class="text-12px text-#222">
                      选择角色模板，再进行调整的规则如下：
                    </div>
                    <div class="text-12px text-#222">
                      1、支持同时选择多个角色模板
                    </div>
                    <div class="text-12px text-#222">
                      2、每次勾选都是在现有基础上新增权限
                    </div>
                    <div class="text-12px text-#222">
                      3、不会覆盖已手动选择的权限
                    </div>
                    <div class="text-12px text-#222">
                      4、如需重新设置，可点击"清空已选"
                    </div>
                  </template>
                  <QuestionCircleOutlined class="text-#06f ml-12px" />
                </a-popover>
              </div>
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
