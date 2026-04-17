<script setup>
import {
  DownOutlined,
  ExclamationCircleOutlined,
  CloseOutlined,
  ExclamationCircleFilled,
  QuestionCircleOutlined,
} from "@ant-design/icons-vue";
import { ref, watch, reactive, onMounted, computed, nextTick } from "vue";
import { debounce } from "lodash-es";
import dayjs from "dayjs";
import { getUserListApi, batchDisabledApi } from "@/api/internal-manage/staff-manage";
import ResignConfirmModal from "./components/ResignConfirmModal.vue";
import BatchEditDepartment from "./components/batchEditDepartment.vue";
import BatchEditRole from "./components/batchEditRole.vue";
import detailEmployees from "./components/detailEmployees.vue";
import editEmployees from "./components/editEmployees.vue";
import addEmployees from "./components/addEmployees.vue";
import ClampedText from "@/components/common/clamped-text.vue";
import AllFilter from "@/components/common/all-filter.vue";
import { getListTreeDepartApi } from "@/api/internal-manage/staff-manage";

// 接收选中的部门信息
const props = defineProps({
  selectedDepartment: {
    type: Object,
    default: null
  },
  departmentList: {
    type: Array,
    default: () => []
  }
});

// 监听部门变化
watch(() => props.selectedDepartment, (newDepartment) => {
  // console.log('员工列表接收到部门变化:', newDepartment);

  if (newDepartment) {
    // 选中部门时，根据部门筛选员工数据
    state.query.deptId = newDepartment.id;
  } else {
    // 取消选中部门时，清空部门筛选条件，显示所有员工
    state.query.deptId = undefined;
  }

  // 重置到第一页并清空选中状态
  pagination.value.current = 1;
  selectedRowKeys.value = [];
  selectedRows.value = [];

  // 触发查询（延迟执行，确保函数已定义）
  nextTick(() => {
    getStaffList(state.query);
  });
}, { immediate: false }); // 改为 false，避免在组件初始化时立即执行

const loading = ref(false);
const allFilterRef = ref(null);
const displayArray = ref([
  'createTime',
  'accountStatus',
  'userType',
  'positionRole',
]);

// 查询状态
const state = reactive({
  query: {
    status: '0',
  },
});

// 分页参数
const pagination = ref({
  current: 1,
  pageSize: 10,
  total: 0,
  showSizeChanger: true,
  showTotal: (total) => `共 ${total} 条`,
  pageSizeOptions: ["5", "10", "20", "50"],
  hideOnSinglePage: true,
  showQuickJumper: true,
});

// 多选相关状态
const selectedRowKeys = ref([]);
const selectedRows = ref([]);

const rowSelection = {
  selectedRowKeys: selectedRowKeys,
  preserveSelectedRowKeys: true, // 保留已选中的行，支持跨页选择
  onChange: (keys, rows) => {
    selectedRowKeys.value = keys;
    selectedRows.value = rows;
    console.log('已选择的员工:', keys, rows);
  },
  onSelectAll: (selected, selectedRows, changeRows) => {
    console.log('全选/取消全选:', selected, selectedRows, changeRows);
  },
  onSelect: (record, selected, selectedRows, nativeEvent) => {
    console.log('单选:', record, selected, selectedRows);
  },
  getCheckboxProps: (record) => ({
    disabled: record.isAdmin, // 超级管理员禁止选择
  }),
};

const dataSource = ref([]);

const allColumns = ref([
  {
    title: '员工姓名/手机号',
    dataIndex: 'nickName',
    key: 'nickName',
    width: 240,
  },
  {
    title: '所属部门',
    dataIndex: 'departNames',
    key: 'departNames',
    width: 160,
  },
  {
    title: '任职角色',
    dataIndex: 'roleName',
    key: 'roleName',
    ellipsis: true,
    width: 160,
  },
  {
    title: '账号状态',
    dataIndex: 'disabled',
    key: 'disabled',
    width: 100,
  },
  {
    title: '员工类型',
    dataIndex: 'userType',
    key: 'userType',
    width: 100,
  },
  {
    title: '创建时间',
    dataIndex: 'createTime',
    key: 'createTime',
    width: 155,
  },
  {
    title: '操作',
    dataIndex: 'action',
    fixed: 'right',
    key: 'action',
    width: 140,
  },
])

import { useTableColumns } from "@/composables/useTableColumns";
import { getInstRolePageApi } from "~@/api/internal-manage/role-manage";
import messageService from "~@/utils/messageService";
const { selectedValues, columnOptions, filteredColumns, totalWidth } =
  useTableColumns({
    storageKey: "staff-list", // 本地存储键名
    allColumns: allColumns, // 原始列配置
    excludeKeys: ["action"], // 需要排除的列键
  });

// 获取员工列表
async function getStaffList(query = {}, id, type) {
  loading.value = true;
  const params = {
    pageRequestModel: {
      needTotal: true,
      pageSize: pagination.value.pageSize,
      pageIndex: pagination.value.current,
      skipCount: 1,
    },
    queryModel: {
      ...query,
    },
  };

  try {
    const res = await getUserListApi(params);
    if (res.code === 200) {
      dataSource.value = res.result || [];
      pagination.value.total = res.total || 0;
    } else {
      messageService.error(res.message || '获取员工列表失败');
    }
  } catch (error) {
    console.error('获取员工列表失败:', error);
    // messageService.error('获取员工列表失败');
  } finally {
    loading.value = false;
    // 清除快捷筛选标记
    if (allFilterRef.value && id && type) {
      allFilterRef.value.clearQuickFilter(id, type);
    }
  }
}

// 处理员工搜索请求
async function handleStaffSearch(searchParams) {
  try {
    // 构建查询参数，将 searchKey 映射到对应的查询字段
    const params = {
      pageRequestModel: searchParams.pageRequestModel,
      queryModel: {
        // 如果有searchKey，使用它作为关键字搜索（可能匹配姓名或手机号）
        searchKey: searchParams.searchKey || undefined,
        // 排除当前列表查询的部分条件，避免冲突
        status: state.query.status,
        deptId: state.query.deptId,
      },
    };

    const res = await getUserListApi(params);
    if (res.code === 200) {
      // 将搜索结果传递给allFilter组件
      if (allFilterRef.value && allFilterRef.value.updateStaffSearchData) {
        allFilterRef.value.updateStaffSearchData(res);
      }
    } else {
      messageService.error(res.message || '搜索员工失败');
    }
  } catch (error) {
    console.error('搜索员工失败:', error);
    messageService.error('搜索员工失败');
  }
}

// 处理表格排序和分页变化
async function handleTableChange(paginationInfo) {
  // 更新分页信息
  pagination.value.current = paginationInfo.current;
  pagination.value.pageSize = paginationInfo.pageSize;

  // 不清空选中状态，支持跨页选择

  // 重新获取数据
  await getStaffList(state.query);
}

// 过滤器字段映射
const filterFieldMapping = {
  createTimeFilter: "createTime",
  channelAccountStatus: "status",
  channelUserType: "userType",
  channelPositionRoleFilter: "roleIds",
  stuPhoneSearchFilter: "id",
};

// 统一处理筛选条件变化
const handleFilterUpdate = debounce(
  async (updates, isClearAll = false, id, type) => {
    if (isClearAll) {
      // 如果是清空所有，重置所有查询条件
      Object.keys(state.query).forEach((key) => {
        if (key !== "pageIndex" && key !== "pageSize") {
          state.query[key] = undefined;
        }
      });
    } else {
      // 如果是单个更新，只更新对应的条件
      Object.assign(state.query, updates);
    }

    // 处理时间范围查询
    state.query.createTimeBegin = state.query?.createTime?.[0];
    state.query.createTimeEnd = state.query?.createTime?.[1];
    delete state.query.createTime;

    // 筛选条件变化时重置到第一页
    pagination.value.current = 1;

    // 筛选条件变化时清空选中状态（因为筛选结果可能不包含之前选中的数据）
    selectedRowKeys.value = [];
    selectedRows.value = [];

    await getStaffList(state.query, id, type);
  },
  300,
  { leading: true, trailing: false }
);

// 生成所有过滤器的更新处理器
const filterUpdateHandlers = computed(() => {
  const handlers = {};
  Object.entries(filterFieldMapping).forEach(([eventKey, fieldName]) => {
    handlers[`update:${eventKey}`] = (val, isClearAll, id, type) =>
      handleFilterUpdate({ [fieldName]: val }, isClearAll, id, type);
  });
  return handlers;
});

const channelPositionRole = ref([])
async function getRoleList(query = { queryModel: {} }) {
  const pages = {
    pageRequestModel: {
      needTotal: true,
      pageSize: 500,
      pageIndex: 1,
      skipCount: 1,
    },
  }
  try {
    const res = await getInstRolePageApi({ ...pages, ...query })
    if (res.code === 200) {
      channelPositionRole.value = res.result || []
    } else {
      messageService.error(res.message || '获取角色列表失败');
    }
  } catch (error) {
    console.error('获取角色列表失败:', error);
    // messageService.error('获取角色列表失败');
  }
}
// 详情按钮触发
function detailEmployeesFunc(record) {
  // console.log('查看员工详情:', record);
  currentEmployeeDetail.value = record;
  detailEmployeesVisible.value = true;
}

// 清空选择
function clearSelection() {
  selectedRowKeys.value = [];
  selectedRows.value = [];
}
// 批量操作
function handleBatchAction({ key }) {
  // console.log('批量操作:', key);
  switch (key) {
    case '1':
      // 批量离职
      batchDisabledFun()
      break;
    case '2':
      // 批量复职
      batchRehireFun()
      break;
    case '3':
      // 批量修改所属部门
      batchEditDepartmentFun()
      break;
    case '4':
      // 批量修改任职角色
      batchEditRoleFun()
      break;
    default:
      break;
  }
}
// 离职确认弹窗相关状态
const resignConfirmVisible = ref(false);
const resignLoading = ref(false);
const currentResignEmployee = ref(null);
const batchResignEmployees = ref([]);

// 批量离职
function batchDisabledFun() {
  // console.log('批量离职:', selectedRowKeys.value);
  if (selectedRowKeys.value.length === 0) {
    messageService.error('请选择员工后，再进行"批量离职"');
    return;
  }

  // 筛选出在职的员工
  const workingEmployees = dataSource.value.filter(item =>
    selectedRowKeys.value.includes(item.id) && !item.disabled && !item.isAdmin
  );

  if (workingEmployees.length === 0) {
    messageService.error('请选择在职中的员工');
    return;
  }

  batchResignEmployees.value = workingEmployees;
  currentResignEmployee.value = null;
  resignConfirmVisible.value = true;
}

// 单个离职
function handleSingleResign(record) {
  if (record.isAdmin) {
    messageService.error('超级管理员不能进行离职操作');
    return;
  }

  if (record.disabled) {
    // 复职操作
    handleRehire(record);
    return;
  }

  currentResignEmployee.value = record;
  batchResignEmployees.value = [];
  resignConfirmVisible.value = true;
}

// 复职操作
async function handleRehire(record) {
  try {
    const res = await batchDisabledApi({
      userIds: [record.id],
      isWork: false,
    });

    if (res.code === 200) {
      messageService.success('复职成功');
      await getStaffList(state.query);
    } else {
      messageService.error(res.message || '复职失败');
    }
  } catch (error) {
    messageService.error('复职失败');
    console.error('复职失败:', error);
  }
}

// 确认离职
async function handleResignConfirm() {
  try {
    resignLoading.value = true;

    const userIds = currentResignEmployee.value
      ? [currentResignEmployee.value.id]
      : batchResignEmployees.value.map(item => item.id);

    const res = await batchDisabledApi({
      userIds,
      isWork: true,
    });

    if (res.code === 200) {
      messageService.success('离职成功');
      resignConfirmVisible.value = false;

      // 清空选中状态
      selectedRowKeys.value = [];
      selectedRows.value = [];

      // 刷新列表
      await getStaffList(state.query);
    } else {
      messageService.error(res.message || '离职失败');
    }
  } catch (error) {
    messageService.error('离职失败');
    console.error('离职失败:', error);
  } finally {
    resignLoading.value = false;
  }
}

// 取消离职
function handleResignCancel() {
  resignConfirmVisible.value = false;
  currentResignEmployee.value = null;
  batchResignEmployees.value = [];
}

// 批量复职
function batchRehireFun() {
  // console.log('批量复职:', selectedRowKeys.value);
  if (selectedRowKeys.value.length === 0) {
    messageService.error('请选择员工后，再进行"批量复职"');
    return;
  }

  // 筛选出离职的员工
  const resignedEmployees = dataSource.value.filter(item =>
    selectedRowKeys.value.includes(item.id) && item.disabled && !item.isAdmin
  );

  if (resignedEmployees.length === 0) {
    messageService.error('请选择已离职的员工');
    return;
  }

  // 直接执行复职操作
  handleBatchRehire(resignedEmployees);
}

// 批量复职操作
async function handleBatchRehire(employees) {
  try {
    const userIds = employees.map(item => item.id);
    const res = await batchDisabledApi({
      userIds,
      isWork: false,
    });

    if (res.code === 200) {
      messageService.success(`已成功复职 ${employees.length} 名员工`);

      // 清空选中状态
      selectedRowKeys.value = [];
      selectedRows.value = [];

      // 刷新列表
      await getStaffList(state.query);
    } else {
      messageService.error(res.message || '批量复职失败');
    }
  } catch (error) {
    messageService.error('批量复职失败');
    console.error('批量复职失败:', error);
  }
}

// 批量修改部门相关状态
const batchEditDepartmentVisible = ref(false);
const batchEditDepartmentEmployees = ref([]);
// const departmentList = ref([]);

// 批量修改所属部门
function batchEditDepartmentFun() {
  // console.log('批量修改所属部门:', selectedRowKeys.value);
  if (selectedRowKeys.value.length === 0) {
    messageService.error('请选择员工后，再进行"批量修改所属部门"');
    return;
  }

  // 筛选出选中的员工
  const selectedEmployees = dataSource.value.filter(item =>
    selectedRowKeys.value.includes(item.id) && !item.isAdmin
  );

  if (selectedEmployees.length === 0) {
    messageService.error('请选择有效的员工');
    return;
  }

  batchEditDepartmentEmployees.value = selectedEmployees;
  batchEditDepartmentVisible.value = true;
}

// 批量修改部门成功回调
async function handleBatchEditDepartmentSuccess() {
  batchEditDepartmentVisible.value = false;

  // 清空选中状态
  selectedRowKeys.value = [];
  selectedRows.value = [];

  // 刷新列表
  await getStaffList(state.query);
}

// 批量修改任职角色相关状态
const batchEditRoleVisible = ref(false);
const batchEditRoleEmployees = ref([]);

// 员工详情相关状态
const detailEmployeesVisible = ref(false);
const currentEmployeeDetail = ref({});

// 员工编辑相关状态
const editEmployeesVisible = ref(false);
const currentEditEmployee = ref({});

// 新增员工相关状态
const addEmployeesVisible = ref(false);

// 批量修改任职角色
function batchEditRoleFun() {
  // console.log('批量修改任职角色:', selectedRowKeys.value);
  if (selectedRowKeys.value.length === 0) {
    messageService.error('请选择员工后，再进行"批量修改任职角色"');
    return;
  }

  // 筛选出选中的员工（排除超级管理员）
  const selectedEmployees = dataSource.value.filter(item =>
    selectedRowKeys.value.includes(item.id) && !item.isAdmin
  );

  if (selectedEmployees.length === 0) {
    messageService.error('请选择有效的员工（超级管理员不支持修改角色）');
    return;
  }

  batchEditRoleEmployees.value = selectedEmployees;
  batchEditRoleVisible.value = true;
}

// 批量修改任职角色成功回调
async function handleBatchEditRoleSuccess() {
  batchEditRoleVisible.value = false;

  // 清空选中状态
  selectedRowKeys.value = [];
  selectedRows.value = [];

  // 刷新列表
  await getStaffList(state.query);
}

// 编辑员工
function handleEditEmployee(record) {
  // console.log('编辑员工:', record);
  currentEditEmployee.value = record;
  editEmployeesVisible.value = true;
}

// 编辑员工成功回调
async function handleEditEmployeeSuccess() {
  editEmployeesVisible.value = false;

  // 刷新列表
  await getStaffList(state.query);
}

// 新建员工
function handleAddEmployee() {
  // console.log('新建员工');
  addEmployeesVisible.value = true;
}

// 新建员工成功回调
async function handleAddEmployeeSuccess() {
  addEmployeesVisible.value = false;

  // 刷新列表
  await getStaffList(state.query);
}

// 获取部门列表
// async function getDepartmentList() {
//   try {
//     const res = await getListTreeDepartApi();
//     if (res.code === 200) {
//       departmentList.value = res.result || [];
//     }
//   } catch (error) {
//     console.error('获取部门列表失败:', error);
//   }
// }

onMounted(async () => {
  await getStaffList(state.query);
  await getRoleList();
  // await getDepartmentList(); // 获取部门列表

  // 如果有初始选中的部门，触发一次部门查询
  if (props.selectedDepartment) {
    state.query.deptId = props.selectedDepartment.id;
    pagination.value.current = 1;
    selectedRowKeys.value = [];
    selectedRows.value = [];
    await getStaffList(state.query);
  }
});
</script>
<template>
  <div>
    <!-- 学员筛选条件 -->
    <div class="filter-wrap bg-white pl-3 pr-3 rounded-4">
      <AllFilter ref="allFilterRef" default-account-status="0" :channel-position-role="channelPositionRole"
        :displayArray="displayArray" :is-quick-show="false" :is-show-clsss-or-course-search="true" search-label="员工姓名"
        search-placeholder="请输入员工姓名"  v-on="filterUpdateHandlers"
        @staff-search="handleStaffSearch" />
    </div>
    <div class="student-list mt-3 pt-3 pb-3 pl-6 pr-6 bg-white rounded-4">
      <div class="tab-table">
        <div class="table-title flex justify-between">
          <div class="total whitespace-nowrap">
            当前共计 {{ pagination.total }} 个员工
          </div>
          <div class="edit flex overflow-x-auto">
            <a-dropdown class="mr-2">
              <template #overlay>
                <a-menu @click="handleBatchAction">
                  <a-menu-item key="1"> 批量离职 </a-menu-item>
                  <a-menu-item key="2"> 批量复职 </a-menu-item>
                  <a-menu-item key="3"> 批量修改所属部门 </a-menu-item>
                  <a-menu-item key="4"> 批量修改任职角色 </a-menu-item>
                </a-menu>
              </template>
              <a-button>
                批量操作{{ selectedRowKeys.length > 0 ? `(${selectedRowKeys.length})` : '' }}
                <DownOutlined :style="{ fontSize: '10px' }" />
              </a-button>
            </a-dropdown>
            <a-button class="mr-2">批量导入</a-button>
            <a-button v-if="selectedRowKeys.length > 0" class="mr-2" @click="clearSelection">
              清空选择
            </a-button>
            <a-button type="primary" @click="handleAddEmployee">新建员工</a-button>
            <!-- 自定义字段 -->
            <!-- <customize-code v-model:checkedValues="selectedValues" :options="columnOptions"
                :total="allColumns.length - 1" :num="selectedValues.length - 1" /> -->
          </div>
        </div>
        <div class="table-content mt-2">
          <a-table :dataSource="dataSource" :pagination="pagination" row-key="id" :columns="filteredColumns"
            :scroll="{ x: totalWidth }" :loading="loading" size="small" :row-selection="rowSelection"
            :sticky="{ offsetHeader: 0 }" @change="handleTableChange">
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'nickName'">
                <div class="flex flex-items-center">
                  <div class="w8 h8 rounded-10 bg-#06f text-#fff flex-center font-500 mr2">
                    {{ record.nickName.slice(0, 1) }}
                  </div>
                  <div>
                    <div class="text-#222">
                      {{ record.nickName }}
                    </div>
                    <div class="text-3 text-#222">
                      {{ record.mobile }}
                      <span v-if="record.isAdmin"
                        class="bg-#fff5e6 text-#f90 font-500 text-3 px2 py1 rounded-10 ml2px">超级管理员</span>
                      <span v-if="!record.isAdmin && !record.activatedStatus" class="text-#ff3333 ml2px">未激活
                        <a-popover placement="top">
                          <template #title>
                            <div style="border-bottom:1px solid #eee;padding-bottom: 6px;">
                              说明
                            </div>
                          </template>
                          <template #content>
                            <div>未激活：当前员工未激活登录过系统，如手机号不正确，员</div>
                            <div>工自己可登录 App 点击"我的-个人信息-账号与安全"修</div>
                            <div>改手机号码或超级管理员修改员工手机号码。</div>
                          </template>
                          <QuestionCircleOutlined />
                        </a-popover>
                      </span>
                    </div>
                  </div>
                </div>
              </template>
              <template v-if="column.key === 'departNames'">
                <a-tooltip placement="topLeft">
                  <template #title>
                    <span>{{ record.departNames }}</span>
                  </template>
                  <div class="flex">
                    <div class="schoolDepartments w-90%">
                      <clamped-text :text="record.departNames" class="text-ellipsis whitespace-nowrap overflow-hidden"
                        :lines="1" />
                    </div>
                  </div>
                </a-tooltip>
              </template>
              <template v-if="column.key === 'roleName'">
                <a-tooltip placement="topLeft">
                  <template #title>
                    <span>点击查看详情</span>
                  </template>
                  <div class="cursor-pointer" @click="detailEmployeesFunc(record)">
                    <div>
                      {{ record.roleNum }} 个
                    </div>
                    <div class="w-90% ">
                      <clamped-text :text="record.roleName" class="text-ellipsis whitespace-nowrap overflow-hidden"
                        :lines="1" />
                    </div>
                  </div>
                </a-tooltip>
              </template>
              <template v-if="column.key === 'disabled'">
                <span :class="record.disabled ? 'text-#ff3333 bg-#ffe6e6' : 'text-#06f bg-#e6f0ff'"
                  class=" text-3 px2 py1 rounded-10 ml2 font500">{{ record.disabled ? '已离职' : '在职中'
                  }}</span>
              </template>
              <template v-if="column.key === 'userType'">
                {{ record.userType === 1 ? '正式员工' : '兼职员工' }}
              </template>
              <template v-if="column.key === 'createTime'">
                {{ dayjs(record.createTime).format('YYYY-MM-DD HH:mm') }}
              </template>
              <template v-if="column.key === 'action'">
                <a-space :size="14">
                  <a-button type="link" class="font500 flex-1 p-0" @click="detailEmployeesFunc(record)">
                    详情
                  </a-button>
                  <a-button type="link" class="font500 flex-1 p-0" @click="handleEditEmployee(record)">
                    编辑
                  </a-button>
                  <a-button type="link" :disabled="record.isAdmin" class="font500 flex-1 p-0"
                    @click="handleSingleResign(record)">
                    {{ record.disabled ? '复职' : '离职' }}
                  </a-button>
                </a-space>
              </template>
            </template>
          </a-table>
        </div>
      </div>
    </div>

    <!-- 离职确认弹窗 -->
    <ResignConfirmModal v-model:open="resignConfirmVisible"
      :employee-names="currentResignEmployee ? currentResignEmployee.nickName : batchResignEmployees.map(emp => emp.nickName).join('、')"
      :employee-count="currentResignEmployee ? 1 : batchResignEmployees.length" :loading="resignLoading"
      @confirm="handleResignConfirm" @cancel="handleResignCancel" />

    <!-- 批量修改部门弹窗 -->
    <BatchEditDepartment v-model="batchEditDepartmentVisible" :batch-user-list="batchEditDepartmentEmployees"
      :department-list="departmentList" @success="handleBatchEditDepartmentSuccess" />

    <!-- 批量修改任职角色弹窗 -->
    <BatchEditRole v-model="batchEditRoleVisible" :batch-user-list="batchEditRoleEmployees"
      @success="handleBatchEditRoleSuccess" />

    <!-- 员工详情抽屉 -->
    <detailEmployees v-model="detailEmployeesVisible" :detail="currentEmployeeDetail" :department-list="departmentList"
      @refresh-list="() => getStaffList(state.query)" />

    <!-- 员工编辑弹窗 -->
    <editEmployees v-model="editEmployeesVisible" :detail="currentEditEmployee" :department-list="departmentList"
      @success="handleEditEmployeeSuccess" @refresh-list="() => getStaffList(state.query)" />

    <!-- 新增员工弹窗 -->
    <addEmployees v-model="addEmployeesVisible" :department-list="departmentList" @success="handleAddEmployeeSuccess" />
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

:deep(.ant-table-sticky-scroll) {
  position: absolute;
  top: 38px;
  bottom: 8px !important;
  background: none;
  border: none;
  display: flex;
  opacity: .6;
  align-items: center;
  z-index: 3;
  display: none;

}

:deep(.ant-table-sticky-scroll-bar) {
  cursor: move !important;
  background-color: rgba(0, 0, 0, .35);
  border-radius: 4px;
  height: 8px;
}
</style>