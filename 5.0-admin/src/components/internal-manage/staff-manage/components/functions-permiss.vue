<script setup lang="ts">
import {
  CheckOutlined,
  DownOutlined,
  SearchOutlined,
  UpOutlined,
} from '@ant-design/icons-vue'
import { Empty } from 'ant-design-vue'
import {
  type AuthorityGroup,
  useRolePermissions,
} from '@/composables/useRolePermissions'
import {
  getDefaultRoleDetailApi,
  getMenuListApi,
} from '~@/api/internal-manage/role-manage'

// 定义props接收角色ID
const props = defineProps({
  roleId: {
    type: Number,
    default: null,
  },
  details: {
    type: Object,
    default: () => ({}),
  },
})

// 定义emit
const emit = defineEmits(['update:details'])

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

const loading = ref(false)
// 权限数据
const initialData: AuthorityGroup[] = []

// 使用权限管理 hook
const {
  boxList,
  searchValue,
  filteredBoxList,
  expandedParentIds,
  expandedChildIds,
  expandedAuthorityIds,
  isAllExpanded,
  isParentExpanded,
  isChildExpanded,
  toggleAllExpand,
  getFilteredChildren,
  getFilteredchildren,
  highlightText,
  updateData,
  setDefaultCheckedByIds,
} = useRolePermissions(initialData, loading)

// 统计选中的权限数量
const selectedPermissionCount = computed(() => {
  let functionCount = 0
  const dataCount = 0

  boxList.value.forEach((parent) => {
    parent.children?.forEach((child) => {
      child.children?.forEach((authority) => {
        if (authority.checked) {
          // 根据权限类型统计 - 这里假设type=1是功能权限，type=2是数据权限
          // 可以根据实际业务需求调整统计逻辑
          functionCount++
        }
      })
    })
  })

  return { functionCount, dataCount }
})

// 获取权限列表
async function getMenuList() {
  loading.value = true
  try {
    const res = await getMenuListApi({ ownType: 'INSTITUTION' })
    // console.log('API原始数据:', res)
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
      // console.log('处理后的数据:', processedData)
      updateData(processedData)

      if (props.roleId) {
        // 根据角色id查询详情，获取最后一级权限id
        getRoleDetail()
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

// 获取角色详情
async function getRoleDetail() {
  try {
    const res = await getDefaultRoleDetailApi({ roleId: props.roleId })
    if (res.code === 200) {
      // console.log('角色详情:', res.result)
      // 获取最后一级权限id
      const lastLevelPermissionIds = getLastLevelCheckedIds(
        res.result.menuIds || [],
      )
      // console.log('最后一级checked权限IDs:', lastLevelPermissionIds)
      // 设置权限树的选中状态
      setDefaultCheckedByIds(lastLevelPermissionIds)

      // 将isDefault值传递给父组件
      const updatedDetails = {
        ...props.details,
        isDefault: res.result.isDefault || false,
        updateName: res.result.updateName || '',
        updateTime: res.result.updateTime || ''
      }
      emit('update:details', updatedDetails)
    }
  }
  catch (error) {
    console.log(error)
    loading.value = false
  }
}

// 重置展开状态的方法
function resetExpandState() {
  // 收起全部
  expandedParentIds.value = []
  expandedChildIds.value = []
  expandedAuthorityIds.value = []
  isAllExpanded.value = false
}

// 刷新权限数据的方法
function refreshPermissions() {
  if (props.roleId) {
    getMenuList()
  }
}

// 暴露方法给父组件调用
defineExpose({
  resetExpandState,
  refreshPermissions,
})

// 监听roleId变化，重新加载数据
watch(
  () => props.roleId,
  (newRoleId) => {
    if (newRoleId) {
      getMenuList()
    }
  },
  { immediate: true },
)
</script>

<template>
  <div class="px-24px">
    <div class="flex justify-between items-center mb-10px">
      <span class="text-14px text-#222 font500">
        <span class="text-#06f">{{ details.functionalAuthorityCount || 0 }}</span>个功能权限，
        <span class="text-#06f">{{ details.dataAuthorityCount || 0 }}</span>个数据权限
      </span>
      <a-button type="link" @click="toggleAllExpand">
        {{ isAllExpanded ? "一键收起" : "一键展开" }}
        <component :is="isAllExpanded ? UpOutlined : DownOutlined" />
      </a-button>
    </div>
    <div class="tree pb-24px">
      <a-input v-model:value="searchValue" class="mb-8px" placeholder="搜索权限点名称或权限描述" allow-clear>
        <template #prefix>
          <SearchOutlined />
        </template>
      </a-input>
      <div class="box">
        <a-spin :spinning="loading" tip="加载中...">
          <div v-for="(item, index) in filteredBoxList" :key="item.id">
            <div class="text-14px text-#222 font-400 h-40px flex items-center justify-between shadow-box" :class="{
              'last-child': index === filteredBoxList.length - 1,
            }">
              <div class="pl-16px">
                <span class="font-500" v-html="highlightText(item.menuName, searchValue)" />
              </div>
            </div>
            <template v-if="isParentExpanded(item.id)">
              <div v-for="child in getFilteredChildren(item)" :key="child.id">
                <div class="text-14px text-#222 font-400 h-40px flex items-center justify-between shadow-box pl-16px">
                  <div>
                    <span class="font-500" v-html="highlightText(child.menuName, searchValue)" />
                  </div>
                </div>
                <template v-if="isChildExpanded(child.id)">
                  <div v-for="(authority, idx) in getFilteredchildren(child, item)" :key="authority.id">
                                        <div
                      class="text-12px text-#222 font-400 min-h-58px py-6px flex items-center justify-between shadow-box bg-#fcfcfc"
                      :class="{
                        'last-child':
                          idx === getFilteredchildren(child, item).length - 1,
                      }">
                      <div class="flex items-center pl-16px">
                        <div class="w-16px h-16px flex items-center justify-center">
                          <CheckOutlined :class="authority.checked ? 'text-#06f' : 'text-#fcfcfc'" class="text-16px mt--8px" />
                        </div>
                        <div class="flex flex-col">
                          <div class="flex items-center">
                            <span class="ml-8px text-13px" v-html="highlightText(authority.name, searchValue)" />
                            <span v-if="authority.weight" class="ml-8px px-6px py-2px text-10px bg-#f0f0f0 text-#666 rounded">
                              权重: {{ authority.weight }}
                            </span>
                          </div>
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
    </div>
  </div>
</template>

<style lang="less" scoped>
.tree {
  .box {
    margin-top: 10px;
    border-radius: 8px;
    border: 1px solid #ddd;
  }

  .shadow-box {
    box-shadow: inset 0 -1px 0 0 #eee;
  }

  .last-child {
    box-shadow: none;
    border-radius: 0 0 8px 8px;
  }
}
</style>
