<script setup>
import { MoreOutlined, PlusOutlined } from '@ant-design/icons-vue'
import { onMounted, ref, watch } from 'vue'
import { Empty, Modal } from 'ant-design-vue'
import AddDepartments from './components/addDepartments.vue'
import EditDepartments from './components/editDepartments.vue'
import { deleteDepart, getListTreeDepartApi, getUserListApi } from '~@/api/internal-manage/staff-manage'
import messageService from '~@/utils/messageService'
import StaffList from '~@/components/internal-manage/staff-manage/staff-list.vue'

const searchValue = ref('')
const isOpen = ref(true)
function handleCloseLeft() {
  isOpen.value = !isOpen.value
}
const selectedKeys = ref(['1'])
const selectedDepartment = ref(null)
const expandedKeys = ref([])

function handleSelect(keys, event) {
  selectedKeys.value = keys
  if (keys.length > 0 && event.selectedNodes.length > 0) {
    selectedDepartment.value = event.selectedNodes[0]
    // console.log('选中的部门:', selectedDepartment.value)
  }
  else {
    selectedDepartment.value = null
  }
}

// 处理展开/收起事件
function onExpand(keys) {
  expandedKeys.value = keys
}
function getParentKey(key, tree) {
  let parentKey
  for (let i = 0; i < tree.length; i++) {
    const node = tree[i]
    if (node.children) {
      if (node.children.some(item => item.key === key)) {
        parentKey = node.key
      }
      else if (getParentKey(key, node.children)) {
        parentKey = getParentKey(key, node.children)
      }
    }
  }
  return parentKey
}
watch(searchValue, (value) => {
  if (value) {
    // 搜索时只展开匹配节点的路径，其他节点收起
    const matchedPaths = []
    const shouldExpand = []

    // 递归查找匹配的节点并记录路径
    function findMatchingNodesWithPath(nodes, currentPath = []) {
      for (const node of nodes) {
        const nodePath = [...currentPath, node.key]

        if (node.departName && node.departName.includes(value)) {
          // 找到匹配的节点，记录完整路径
          matchedPaths.push({
            node,
            path: nodePath,
            isTopLevel: currentPath.length === 0,
          })
        }

        // 继续搜索子节点
        if (node.children && node.children.length > 0) {
          findMatchingNodesWithPath(node.children, nodePath)
        }
      }
    }

    findMatchingNodesWithPath(treeData.value)

    if (matchedPaths.length > 0) {
      // 有匹配结果时，只展开必要的父节点路径
      hasSearchResult.value = true
      matchedPaths.forEach(({ path, isTopLevel }) => {
        if (isTopLevel) {
          // 如果匹配的是顶级节点，不展开它（保持收起）
          // 不添加任何展开节点
        }
        else {
          // 如果匹配的不是顶级节点，展开到其父节点（不包括自己）
          const parentPath = path.slice(0, -1)
          shouldExpand.push(...parentPath)
        }
      })

      // 去重并设置展开节点
      expandedKeys.value = [...new Set(shouldExpand)]
      console.log(`搜索"${value}"找到 ${matchedPaths.length} 个匹配项，展开 ${expandedKeys.value.length} 个父节点`)
    }
    else {
      // 搜索不到数据时，全部收起并显示无数据状态
      hasSearchResult.value = false
      expandedKeys.value = []
      console.log(`搜索"${value}"无匹配结果，全部收起`)
    }
  }
  else {
    // 没有搜索词时，恢复全部展开
    hasSearchResult.value = true
    expandedKeys.value = getAllKeys(treeData.value)
    console.log('清空搜索，恢复全部展开')
  }
})
// 默认部门数据（当API无数据时使用）
const defaultTreeData = []

const treeData = ref(defaultTreeData)
const loading = ref(false)
const hasSearchResult = ref(true)

// 初始化时设置默认展开
function initExpandedKeys() {
  expandedKeys.value = getAllKeys(treeData.value)
}

// 获取部门树
async function getTreeData() {
  loading.value = true
  try {
    const res = await getListTreeDepartApi()
    if (res.code === 200 && res.result && res.result.length > 0) {
      // console.log('API返回的部门数据:', res.result)
      // 将API返回的数据转换为tree组件需要的格式
      treeData.value = transformTreeData(res.result)
    }
    else {
      console.log('API无数据，使用默认部门数据')
      treeData.value = defaultTreeData
    }

    // 设置默认展开所有节点
    expandedKeys.value = getAllKeys(treeData.value)
  }
  catch (error) {
    console.error('获取部门数据失败:', error)
    // 如果API调用失败，使用默认数据
    treeData.value = defaultTreeData
    // 设置默认展开所有节点
    expandedKeys.value = getAllKeys(treeData.value)
  }
  finally {
    loading.value = false
  }
}

// 收集所有节点的key，用于全部展开
function getAllKeys(data) {
  const keys = []

  function traverse(nodes) {
    nodes.forEach((node) => {
      keys.push(node.key)
      if (node.children && node.children.length > 0) {
        traverse(node.children)
      }
    })
  }

  traverse(data)
  return keys
}

// 转换数据格式以适配ant-design-vue的tree组件
function transformTreeData(data) {
  if (!Array.isArray(data))
    return []

  return data.map((item) => {
    const transformed = {
      key: item.id || item.departId || item.key || Math.random().toString(36).substr(2, 9),
      departName: item.departName || item.name || item.title || item.label,
      id: item.id || item.departId,
      parentId: item.parentId || item.pid,
      ...item,
    }

    // 如果有子部门，递归转换
    if (item.children && Array.isArray(item.children) && item.children.length > 0) {
      transformed.children = transformTreeData(item.children)
    }

    return transformed
  })
}
const addDepartmentsOpen = ref(false)
const editDepartmentsOpen = ref(false)
const params = ref({})
// 新增部门
function addDepartmentsFunc({ id, orgId, data }) {
  console.log('addDepartmentsFunc', id, orgId, data)
  addDepartmentsOpen.value = true
  params.value = { pid: id, orgId, data }
}
// 编辑部门
function editDepartmentsFunc(props) {
  editDepartmentsOpen.value = true
  params.value = props
}
// 删除部门
async function delDepartmentsFunc(props) {
  console.log('delDepartmentsFunc', props)
  // props.children.length > 0 提示不能删除
  if (props.data.children.length > 0) {
    messageService.error('该部门下有子部门，不能删除')
    return
  }

  // 调用 getUserListApi 接口检查部门下是否有员工
  try {
    const userListParams = {
      pageRequestModel: {
        needTotal: true,
        pageSize: 1, // 只需要检查是否有数据，设置为1即可
        pageIndex: 1,
        skipCount: 1,
      },
      queryModel: {
        deptId: props.id, // 使用部门ID查询
      },
    }
    
    const userListRes = await getUserListApi(userListParams)
    
    if (userListRes.code === 200 && userListRes.total > 0) {
      messageService.error('暂不可删除，当前部门下有员工（含离职员工）')
      return
    }
  }
  catch (error) {
    console.error('查询部门员工失败:', error)
    messageService.error('查询部门员工失败，请稍后重试')
    return
  }

  // 二次确认弹窗
  Modal.confirm({
    title: '确定删除该部门？',
    content: '删除后不可恢复',
    centered: true,
    okText: '确定',
    cancelText: '再想想',
    onOk: async () => {
      try {
        const res = await deleteDepart({ id: props.id, uuid: props.uuid, version: props.version })
        if (res.code === 200) {
          messageService.success('删除成功')
          // 删除成功后重新获取部门树数据
          getTreeData()
        }
        else {
          messageService.error(res.message || '删除失败')
        }
      }
      catch (error) {
        console.error('删除部门失败:', error)
        messageService.error('删除失败')
      }
    },
  })
}
function departmentsSuccess() {
  // 在这里重新掉获取部门树的api
  getTreeData()
}
onMounted(() => {
  // 先设置初始展开状态
  initExpandedKeys()
  // 然后获取API数据
  getTreeData()
})
</script>

<template>
  <div class="staff-manage flex">
    <div class="left flex h-100% mr4">
      <div class="left-wrap bg-white rounded-lt-4 transition-all duration-300 ease-in-out overflow-hidden"
        :class="isOpen ? 'w-60' : 'w-0'">
        <div v-if="isOpen">
          <div class="left-t text-4 font800 text-#222 pl-6 h14 flex flex-items-center" :class="isOpen ? 'w-60' : 'w-0'">
            部门架构
          </div>
          <div class="px4.5 pr0">
            <a-input v-model:value="searchValue" style="margin-bottom: 8px" placeholder="输入名称关键字" />
            <a-spin :spinning="loading" tip="加载中...">
              <a-tree :tree-data="treeData" block-node default-expand-all :expanded-keys="expandedKeys"
                :selected-keys="selectedKeys" @expand="onExpand" @select="handleSelect">
                <template #title="{ departName, parentId, ...rest }">
                  <div class="flex justify-between">
                    <span v-if="departName.indexOf(searchValue) > -1" class="flex">
                      <clamped-text :lines="1" :text="departName.substring(0, departName.indexOf(searchValue))"></clamped-text>
                      <div class="flex whitespace-nowrap mr-4px">
                        <span style="color: #f03;">{{ searchValue }}</span>
                        <clamped-text :lines="1" :text="departName.substring(
                            departName.indexOf(searchValue) + searchValue.length,
                          )"></clamped-text>
                      </div>
                    </span>
                    <span v-else>{{ departName }}</span>
                    <div class="op-icons">
                      <PlusOutlined :class="parentId ? '' : 'mr10px'" @click.stop="addDepartmentsFunc(rest)" />
                      <!-- 顶级部门不显示更多操作图标 -->
                      <!-- <MoreOutlined @click.stop class="transform-rotate-90 ml2.5" /> -->
                      <a-dropdown>
                        <MoreOutlined v-if="parentId" class="transform-rotate-90 ml2.5" @click.stop />
                        <template #overlay>
                          <a-menu>
                            <a-menu-item>
                              <a href="javascript:;" @click="editDepartmentsFunc({ ...rest, departName, pid })">编辑部门</a>
                            </a-menu-item>
                            <a-menu-item>
                              <a href="javascript:;" @click="delDepartmentsFunc(rest)">删除部门</a>
                            </a-menu-item>
                          </a-menu>
                        </template>
                      </a-dropdown>
                    </div>
                  </div>
                </template>
              </a-tree>
              <div v-if="treeData.length == 0" class="h-200px flex items-center justify-center">
                <a-empty :image="Empty.PRESENTED_IMAGE_SIMPLE" />
              </div>
            </a-spin>
          </div>
        </div>
      </div>
      <div class="left-close w-7 bg-white flex-center rounded-rt-4 bg-#fafafa  " :class="!isOpen ? 'rounded-lt-4' : ''">
        <img class="cursor-pointer py-8" :class="!isOpen ? 'rotate-180' : ''" width="16"
          src="https://pcsys.admin.ybc365.com/3a66d1e6-3e92-48c8-a082-9ecbe657b73d.png" alt="" @click="handleCloseLeft">
      </div>
    </div>
    <div class="right flex-1 overflow-auto">
      <StaffList :selected-department="selectedDepartment" :department-list="treeData" />
    </div>
    <AddDepartments v-model:open="addDepartmentsOpen" :params="params" @success="departmentsSuccess" />
    <EditDepartments v-model:open="editDepartmentsOpen" :params="params" :tree-data="treeData"
      @success="departmentsSuccess" />
  </div>
</template>

<style lang="less" scoped>
.staff-manage {
  height: calc(100vh - 120px);
  display: flex;
  
  .left {
    height: 100%;
  }
  
  .right {
    height: 100%;
    overflow: auto;
    padding-right: 4px; /* Add some padding for the scrollbar */
  }

  :deep(.ant-tree-treenode-selected) {
    background-color: #06f;
    border-radius: 4px;
    height: 24px;

    &:hover {
      background-color: #06f !important;
    }

    .ant-tree-node-selected {
      background: transparent;
      color: #fff;
    }

    .ant-tree-switcher-icon {
      color: #fff;
    }
  }

  :deep(.ant-tree-treenode) {
    height: 24px;
    margin: 2px 0;

    &:hover {
      background-color: rgba(0, 0, 0, 0.05);
    }
  }

  .op-icons {
    display: none;
  }

  :deep(.ant-tree-node-content-wrapper:hover) .op-icons {
    display: block;
  }
}
</style>
