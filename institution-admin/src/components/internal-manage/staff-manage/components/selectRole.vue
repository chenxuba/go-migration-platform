<!-- 选择角色 -->
<script setup>
import createRolesDrawer from '@/components/common/create-roles-drawer.vue'
import { saveRoleApi, getInstRolePageApi, getMenuListApi } from '@/api/internal-manage/role-manage'
import messageService from '@/utils/messageService'
import emitter, { EVENTS } from '@/utils/eventBus'

const { selectRoleList, roleInfos } = defineProps({
  selectRoleList: {
    type: Array,
    default: () => [],
  },
  roleInfos: {
    type: Array,
    default: () => [],
  },
})

const emit = defineEmits(['success', 'update:roleInfos'])

const open = defineModel({
  type: Boolean,
  default: false,
})

const createRoleOpen = ref(false)

// 展开的角色
const activeRoles = ref([])

// 选中的角色
const selectRole = ref(selectRoleList)

// 存储所有菜单数据
const allMenus = ref([])

// 存储每个角色的权限树
const roleMenusMap = ref(new Map())

// 权限数据加载状态
const permissionsLoading = ref(false)

// 展开/收起角色
function selectRoleFunc(roleId) {
  if (activeRoles.value.includes(roleId)) {
    activeRoles.value.splice(activeRoles.value.indexOf(roleId), 1)
  }
  else {
    activeRoles.value.push(roleId)
  }
}

function handleOk() {
  emit('success', selectRole.value)
  open.value = false
  selectRole.value = []
  activeRoles.value = []
}

function cancel() {
  open.value = false
  selectRole.value = []
  activeRoles.value = []
}

async function createRoleSuccess(data) {
  try {
    const res = await saveRoleApi(data)
    if (res.code === 200) {
      // Close the drawer
      createRoleOpen.value = false
      messageService.success("创建成功")
      
      // Refresh role list
      await refreshRoleList(res.result?.id)
    }
    // 关闭按钮loading
    emitter.emit(EVENTS.CLOSE_LOADING_EVENT)
  }
  catch (err) {
    console.log(err)
    messageService.error('创建失败')
    // 关闭按钮loading
    emitter.emit(EVENTS.CLOSE_LOADING_EVENT)
  }
}

// 刷新角色列表
async function refreshRoleList(newRoleId) {
  try {
    const pages = {
      pageRequestModel: {
        needTotal: true,
        pageSize: 500,
        pageIndex: 1,
        skipCount: 1,
      },
      queryModel:{}
    }
    
    const rolesRes = await getInstRolePageApi(pages)
    if (rolesRes.code === 200) {
      // Update the parent component with the new role list
      emit('update:roleInfos', rolesRes.result || [])
      
      // Add the new role to the selected roles if available
      if (newRoleId) {
        selectRole.value = [...selectRole.value, newRoleId]
      }
    }
  } catch (error) {
    console.error('刷新角色列表失败:', error)
  }
}

// 获取菜单列表
async function getMenuList() {
  try {
    permissionsLoading.value = true
    const res = await getMenuListApi({ ownType: 'INSTITUTION' })
    if (res.code === 200) {
      allMenus.value = res.result || []
      // 菜单数据加载完成后，预处理所有角色的权限
      preloadAllRoleMenus()
    }
  }
  catch (error) {
    console.log(error)
  }
  finally {
    permissionsLoading.value = false
  }
}



// 根据权限ID数组构建菜单树
function buildMenuTree(menuIds, allMenus) {
  if (!menuIds || !Array.isArray(menuIds) || menuIds.length === 0) {
    return []
  }

  const menuIdSet = new Set(menuIds)

  // 检查菜单及其子菜单是否有权限
  function hasMenuPermission(menu, menuIdSet) {
    if (menuIdSet.has(menu.id) || menuIdSet.has(menu.menuId)) {
      return true
    }
    if (menu.children && menu.children.length > 0) {
      return menu.children.some(child => hasMenuPermission(child, menuIdSet))
    }
    return false
  }

  // 递归处理菜单树 - 创建新对象，不修改原始数据
  function processMenus(menus, level = 0) {
    if (!menus || !Array.isArray(menus)) return []
    
    return menus
      .filter(menu => {
        // 检查当前菜单是否有权限或其子菜单有权限
        const hasPermission = menuIdSet.has(menu.id) || menuIdSet.has(menu.menuId)
        const hasChildPermission = menu.children && menu.children.some(child => 
          hasMenuPermission(child, menuIdSet)
        )
        return hasPermission || hasChildPermission
      })
      .map(menu => {
        // 创建新的菜单对象，避免修改原始数据
        const newMenu = {
          ...menu,
          id: menu.id,
          menuId: menu.menuId,
          menuName: menu.menuName || menu.name,
          name: menu.menuName || menu.name
        }
        
        // 递归处理子菜单
        if (menu.children && menu.children.length > 0) {
          newMenu.children = processMenus(menu.children, level + 1)
        } else {
          newMenu.children = []
        }
        
        return newMenu
      })
  }

  return processMenus(allMenus)
}

// 预加载所有角色的权限数据
async function preloadAllRoleMenus() {
  if (!roleInfos || roleInfos.length === 0) {
    return
  }

  // 直接处理角色数据，不需要额外API调用
  roleInfos.forEach((roleInfo) => {
    try {
      // 获取角色的权限ID列表
      const menuIds = roleInfo.menuIds || []
      
      // 构建该角色的菜单树
      const roleMenuTree = buildMenuTree(menuIds, allMenus.value)
      
      // 存储到Map中
      roleMenusMap.value.set(roleInfo.id, roleMenuTree)
    } catch (error) {
      console.error(`处理角色 ${roleInfo.roleName} 权限失败:`, error)
      // 即使某个角色失败，也不影响其他角色的处理
      roleMenusMap.value.set(roleInfo.id, [])
    }
  })
}

// 获取单个角色的权限信息（备用函数）
function loadSingleRoleMenu(roleId) {
  if (roleMenusMap.value.has(roleId)) {
    return // 已经处理过了
  }

  try {
    // 从roleInfos中找到对应的角色
    const roleInfo = roleInfos.find(role => role.id === roleId)
    if (roleInfo) {
      // 获取角色的权限ID列表
      const menuIds = roleInfo.menuIds || []
      
      // 构建该角色的菜单树
      const roleMenuTree = buildMenuTree(menuIds, allMenus.value)
      
      // 存储到Map中
      roleMenusMap.value.set(roleId, roleMenuTree)
    }
  } catch (error) {
    console.error('处理角色权限失败:', error)
    roleMenusMap.value.set(roleId, [])
  }
}

// 获取角色的权限菜单树
function getRoleMenus(roleId) {
  return roleMenusMap.value.get(roleId) || []
}

// 格式化菜单树为展示用的结构
function formatMenuForDisplay(menus) {
  if (!menus || menus.length === 0) return []
  
  return menus.map(menu => {
    const children = menu.children ? menu.children.map(child => child.menuName || child.name).filter(Boolean) : []
    return {
      name: menu.menuName || menu.name,
      children: children
    }
  })
}

watch(
  () => open.value,
  async (val) => {
    if (val) {
      activeRoles.value = []
      selectRole.value = selectRoleList
      // 查询权限树
      await getMenuList()
    }
  },
)
</script>

<template>
  <a-modal
    v-model:open="open" :body-style="{ padding: 0 }" title="任职角色" :mask-closable="false" :keyboard="false"
    width="800px" :destroy-on-close="true" :centered="true" class="select-role-modal" @cancel="cancel" @ok="handleOk"
  >
    <div class="text-16px">
      <div class="flex justify-between px-25px py-15px">
        <div>已选角色：{{ selectRole.length }}</div>
        <div>
          <span>没有符合的角色？</span>
          <span class="text-#06f cursor-pointer" @click="createRoleOpen = true">立即创建</span>
        </div>
      </div>
      <!-- 角色列表 -->
      <div class="w-100%">
        <a-checkbox-group
          v-model:value="selectRole"
          class="w-100% flex-col h-500px flex-nowrap overflow-y-auto scrollbar"
        >
          <div
            v-for="roleInfo in roleInfos" :key="roleInfo.id"
            class="border-t-1px border-#eee border-solid border-0px"
          >
            <div class="flex justify-between px-25px py-15px">
              <div class="flex items-start gap-10px flex-1 min-w-0">
                <a-checkbox :value="roleInfo.id" class="mt--3px flex-shrink-0" />
                <div class="flex flex-col gap-5px flex-1 min-w-0">
                  <div class="font-bold">
                    {{ roleInfo.roleName }}
                  </div>
                  <div class="leading-24px">
                    <span class="text-#06f font-500 mr-2px">{{ roleInfo.functionalAuthorityCount }}</span>
                    <span>个功能权限,</span>
                    <span class="text-#06f font-500 mx-2px">{{ roleInfo.dataAuthorityCount }}</span>
                    <span>个数据权限</span>
                  </div>
                  <div class="overflow-hidden text-ellipsis whitespace-nowrap max-w-600px">
                    <span class="text-#666 text-13px">角色描述：{{ roleInfo.description }}</span>
                  </div>
                </div>
              </div>
              <div
                class="cursor-pointer text-12px px-8px max-h-24px border-#ddd border-solid border-1px flex items-center justify-center rounded-15px text-center text-#666 flex-shrink-0 ml-10px"
                @click="selectRoleFunc(roleInfo.id)"
              >
                {{ activeRoles.includes(roleInfo.id) ? '收起' : '展开' }}
              </div>
            </div>
            <template v-if="activeRoles.includes(roleInfo.id)">
              <div class="bg-#fafafa px-50px">
                <template v-if="permissionsLoading">
                  <div class="py-20px text-center">
                    <a-spin size="small" />
                    <span class="ml-10px text-#999">权限数据加载中...</span>
                  </div>
                </template>
                <template v-else-if="getRoleMenus(roleInfo.id).length > 0">
                  <div 
                    v-for="(menuGroup, index) in formatMenuForDisplay(getRoleMenus(roleInfo.id))" 
                    :key="index" 
                    class="py-10px border-b-1px border-#eee border-solid border-0px"
                    :style="index === formatMenuForDisplay(getRoleMenus(roleInfo.id)).length - 1 ? 'border: none' : ''"
                  >
                    <span class="font-500">{{ menuGroup.name }}</span>
                    <p class="text-13px text-#666 m-0 mt-7px" v-if="menuGroup.children.length > 0">
                      {{ menuGroup.children.join('、') }}
                    </p>
                  </div>
                </template>
                <template v-else>
                  <div class="py-20px text-center text-#999">
                    暂无权限信息
                  </div>
                </template>
              </div>
            </template>
          </div>
        </a-checkbox-group>
      </div>
    </div>
    <create-roles-drawer v-model:open="createRoleOpen" @on-success="createRoleSuccess" />
  </a-modal>
</template>

<style lang="less">
.select-role-modal {
  :deep(.ant-modal) {
    width: 800px !important;
    max-width: 95vw !important;
  }
}

@media (max-width: 768px) {
  .select-role-modal {
    :deep(.ant-modal) {
      width: 90vw !important;
      max-width: 90vw !important;
    }
  }
}

@media (max-width: 480px) {
  .select-role-modal {
    :deep(.ant-modal) {
      width: 95vw !important;
      max-width: 95vw !important;
    }
  }
}
</style>
