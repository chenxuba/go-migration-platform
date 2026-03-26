<script setup>
import dayjs from 'dayjs'
import { debounce } from 'lodash-es'
import {
  getInstRolePageApi,
  saveRoleApi,
  updateRoleApi,
} from '@/api/internal-manage/role-manage'

import { useTableColumns } from '@/composables/useTableColumns'
import messageService from '~@/utils/messageService'
import emitter, { EVENTS } from '~@/utils/eventBus'

const loading = ref(false)
const allFilterRef = ref(null)
const displayArray = ref(['lastEditedTime'])
const state = reactive({
  query: {
    lastEditedTime: '',
    roleId: '',
  },
})

// 分页参数
const pagination = ref({
  current: 1,
  pageSize: 10,
  total: 0,
  showSizeChanger: true,
  showTotal: total => `共 ${total} 条`,
  pageSizeOptions: ['5', '10', '20', '50'],
  hideOnSinglePage: true,
  showQuickJumper: true,
})

const dataSource = ref([])
const allColumns = ref([
  {
    title: '角色名称',
    dataIndex: 'roleName',
    key: 'roleName',
    width: 180,
  },

  {
    title: '角色描述',
    dataIndex: 'description',
    key: 'description',
    width: 240,
  },
  {
    title: '任职员工',
    dataIndex: 'staffNum',
    key: 'staffNum',
    width: 160,
  },
  {
    title: '最近编辑时间',
    dataIndex: 'updateTime',
    key: 'updateTime',
    width: 180,
  },
  {
    title: '操作',
    dataIndex: 'action',
    fixed: 'right',
    key: 'action',
    width: 120,
  },
])
const { filteredColumns, totalWidth } = useTableColumns({
  storageKey: 'role-list', // 本地存储键名
  allColumns, // 原始列配置
  excludeKeys: ['action'], // 需要排除的列键
})


async function getRoleList(query = {}, id, type) {
  
  loading.value = true
  const pages = {
    pageRequestModel: {
      needTotal: true,
      pageSize: pagination.value.pageSize,
      pageIndex: pagination.value.current,
      skipCount: 1,
    },
    queryModel: {
      ...query,
    },
  }
  const res = await getInstRolePageApi({ ...pages })
  dataSource.value = res.result || []
  pagination.value.total = res.total || 0
  loading.value = false
  // 清除快捷筛选标记
  if (allFilterRef.value && id && type) {
    allFilterRef.value.clearQuickFilter(id, type)
  }
}

// 处理角色搜索请求
async function handleRoleSearch(searchParams) {
  try {
    // 构建查询参数，将 searchKey 映射到角色名称搜索
    const params = {
      pageRequestModel: searchParams.pageRequestModel,
      queryModel: {
        // 使用 searchKey 作为角色名称搜索
        searchKey: searchParams.searchKey || undefined,
        // 保持其他现有的查询条件
        ...state.query,
      },
    }

    const res = await getInstRolePageApi(params)
    if (res.code === 200) {
      // 将搜索结果传递给allFilter组件
      if (allFilterRef.value && allFilterRef.value.updateStaffSearchData) {
        allFilterRef.value.updateStaffSearchData(res)
      }
    } else {
      messageService.error(res.message || '搜索角色失败')
    }
  } catch (error) {
    console.error('搜索角色失败:', error)
    messageService.error('搜索角色失败')
  }
}

onMounted(async () => {
  await getRoleList()
})

// 过滤器字段映射
const filterFieldMapping = {
  lastEditedTimeFilter: 'lastEditedTime',
  stuPhoneSearchFilter: 'roleId',
}
// 统一处理筛选条件变化
const handleFilterUpdate = debounce(
  (updates, isClearAll = false, id, type) => {
    if (isClearAll) {
      // 如果是清空所有，重置所有查询条件
      Object.keys(state.query).forEach((key) => {
        if (key !== 'pageNo' && key !== 'pageSize') {
          state.query[key] = undefined
        }
      })
    }
    else {
      // 如果是单个更新，只更新对应的条件
      Object.assign(state.query, updates)
    }
    state.query.updateTimeBegin = state.query?.lastEditedTime?.[0]
    state.query.updateTimeEnd = state.query?.lastEditedTime?.[1]
    delete state.query.lastEditedTime

    // 筛选条件变化时重置到第一页
    pagination.value.current = 1
    getRoleList(state.query, id, type)
  },
  300,
  { leading: true, trailing: false },
)
// 生成所有过滤器的更新处理器
const filterUpdateHandlers = computed(() => {
  const handlers = {}
  Object.entries(filterFieldMapping).forEach(([eventKey, fieldName]) => {
    handlers[`update:${eventKey}`] = (val, isClearAll, id, type) =>
      handleFilterUpdate({ [fieldName]: val }, isClearAll, id, type)
  })
  return handlers
})
const createRolesDrawerOpen = ref(false)

const roleId = ref(null)
function handleOpenCreateRole() {
  createRolesDrawerOpen.value = true
  roleId.value = null
}
// 处理表格排序和分页变化
function handleTableChange(paginationInfo) {
  // 更新分页信息
  pagination.value.current = paginationInfo.current
  pagination.value.pageSize = paginationInfo.pageSize

  // 重新获取数据
  getRoleList(state.query)
}

async function handleSaveFun(data) {
  console.log(data)

  try {
    const res = data.roleId
      ? await updateRoleApi(data)
      : await saveRoleApi(data)
    if (res.code === 200) {
      await getRoleList()
      createRolesDrawerOpen.value = false
      messageService.success(data.roleId ? '更新成功' : '创建成功')

      // 如果是编辑操作且详情抽屉是打开的，更新details数据
      if (data.roleId && roleDetailsDrawerOpen.value && details.value?.id === data.roleId) {
        const updatedRole = dataSource.value.find(role => role.id === data.roleId)
        if (updatedRole) {
          details.value = { ...updatedRole }
        }
      }
    }
    // 关闭按钮loaidng
    emitter.emit(EVENTS.CLOSE_LOADING_EVENT)
  }
  catch (err) {
    console.log(err)
    messageService.error('创建失败')
  }
}
// 编辑按钮触发
function handleEditRole(record) {
  // console.log(record);
  createRolesDrawerOpen.value = true
  roleId.value = record.id
}
const roleDetailsDrawerOpen = ref(false)
const details = ref(null)
// 详情按钮触发
function handleSeeDetail(record) {
  roleDetailsDrawerOpen.value = true
  roleId.value = record.id
  details.value = record
}
</script>

<template>
  <div>
    <!-- 筛选条件 -->
    <div class="filter-wrap bg-white pl-3 pr-3 rounded-4">
      <all-filter
        ref="allFilterRef"
        :display-array="displayArray"
        :is-quick-show="false"
        :is-show-clsss-or-course-search="true"
        search-label="角色名称"
        search-placeholder="请输入角色名称"
        v-on="filterUpdateHandlers"
        @staff-search="handleRoleSearch"
      />
    </div>
    <div class="student-list mt-3 pt-3 pb-3 pl-6 pr-6 bg-white rounded-4">
      <div class="tab-table">
        <div class="table-title flex justify-between">
          <div class="total">
            当前共计 {{ pagination.total }} 个角色
          </div>
          <div class="edit flex">
            <a-button type="primary" @click="handleOpenCreateRole()">
              新建角色
            </a-button>
          </div>
        </div>
        <div class="table-content mt-2">
          <a-table
            size="small"
            :data-source="dataSource"
            :pagination="pagination"
            :columns="filteredColumns"
            :scroll="{ x: totalWidth }"
            :sticky="{ offsetHeader: 100 }"
            :loading="loading"
            @change="handleTableChange"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'roleName'">
                <div class="text-#222">
                  {{ record.roleName }}
                  <span
                    v-if="record.isDefault"
                    class="bg-#e6f0ff text-#06f text-3 px2 py1 rounded-10 font500"
                  >系统默认</span>
                </div>
                <div class="text-3 text-#666">
                  {{ record.functionalAuthorityCount }}个功能权限，{{
                    record.dataAuthorityCount
                  }}个数据权限
                </div>
              </template>
              <!-- 角色描述 -->
              <template v-if="column.key === 'description'">
                <div class="text-#555 w-80%">
                  <clamped-text
                    :text="record.description || '-'"
                    :lines="2"
                  />
                </div>
              </template>
              <!-- 任职员工 -->
              <template v-if="column.key === 'staffNum'">
                <div class="cursor-pointer">
                  <div class="text-#222">
                    {{ record.staffCount||0 }} 个
                  </div>
                  <clamped-text
                    :lines="1"
                    class="text-3 text-#666 w-80%"
                    :text="record.staffNames && record.staffNames.length > 0 ? record.staffNames.join('、') : ''"
                  />
                </div>
              </template>
              <!-- 最近编辑时间 -->
              <template v-if="column.key === 'updateTime'">
                <div v-if="record.isDefault" class="text-#ccc">
                  默认角色不可编辑
                </div>
                <div v-else class="text-#222">
                  <div>
                    {{ dayjs(record.updateTime).format("YYYY-MM-DD HH:mm") }}
                  </div>
                  <div>{{ record.updateName || record.createName }}</div>
                </div>
              </template>

              <template v-if="column.key === 'action'">
                <a-space :size="16">
                  <a @click="handleSeeDetail(record)">详情</a>
                  <a v-if="!record.isDefault" @click="handleEditRole(record)">编辑</a>
                </a-space>
              </template>
            </template>
          </a-table>
        </div>
      </div>
    </div>
    <create-roles-drawer
      v-model:open="createRolesDrawerOpen"
      :role-id="roleId"
      @on-success="handleSaveFun"
    />
    <roles-details-drawer
      v-model:open="roleDetailsDrawerOpen"
      :role-id="roleId"
      :details="details"
      @on-edit-success="handleSaveFun"
    />
  </div>
</template>

<style lang="less" scoped>
.total {
  position: relative;
  padding-left: 10px;
  color: #222;
  display: flex;
  align-items: center;

  &::before {
    display: inline-block;
    background: var(--pro-ant-color-primary);
    border-radius: 2px;
    content: "";
    height: 12px;
    left: 0;
    position: absolute;
    width: 4px;
  }
}

span.dot {
  border-radius: 50%;
  display: inline-block;
  height: 6px;
  position: relative;
  vertical-align: middle;
  width: 6px;
  margin-right: 4px;
  background: #06f;
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
