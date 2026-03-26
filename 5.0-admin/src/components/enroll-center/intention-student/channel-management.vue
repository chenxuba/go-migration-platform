<script setup>
import { computed, onMounted, ref } from 'vue'
import { debounce } from 'lodash-es'
import messageService from '~@/utils/messageService'
import { adjustChannelApi, createChannelApi, getChannelCategoryListApi, getChannelPCPageApi, updateChannelApi, updateChannelStatusApi } from '~@/api/enroll-center/intention-student'

const allFilterRef = ref(null)
const dataSource = ref([])
const queryParams = ref({
  pageNo: 1,
  pageSize: 50,
})
const allColumns = ref([
  {
    title: '渠道名称',
    dataIndex: 'channelName',
    key: 'channelName',
    width: 140,
  },
  {
    title: '渠道分类',
    dataIndex: 'categoryName',
    width: 120,
    key: 'categoryName',
  },
  {
    title: '渠道状态',
    key: 'isDisabled',
    dataIndex: 'isDisabled',
    width: 120,
  },
  {
    title: '渠道类型',
    key: 'isDefault',
    dataIndex: 'isDefault',
    width: 120,

  },
  {
    title: '有效线索数',
    dataIndex: 'invalidCount',
    key: 'invalidCount',
    width: 120,
  },
  {
    title: '成交转化数',
    dataIndex: 'dealTransformCount',
    key: 'dealTransformCount',
    width: 120,
  },
  {
    title: '成交转化率',
    dataIndex: 'dealTransformRate',
    key: 'dealTransformRate',
    width: 120,
  },

  {
    title: '备注',
    dataIndex: 'remark',
    key: 'remark',
    width: 140,

  },
  {
    title: '创建时间',
    dataIndex: 'createTime',
    key: 'createTime',
    width: 180,
  },
  {
    title: '操作',
    dataIndex: 'action',
    key: 'action',
    width: 175,
    fixed: 'right',
  },
])
const selectedRowsData = ref([])
const selectedRowsKeys = ref([])
const rowSelection = computed(() => ({
  selectedRowKeys: selectedRowsKeys.value,
  onChange: (selectedRowKeys, selectedRows) => {
    // console.log(`selectedRowKeys: ${selectedRowKeys}`, 'selectedRows: ', selectedRows);
    selectedRowsKeys.value = selectedRowKeys
    selectedRowsData.value = selectedRows
    // const rawData = selectedRows.map(row => toRaw(row))
    // console.log('原始数组数据:', rawData);
  },
  getCheckboxProps: record => ({
    disabled: record.isDefault, // Column configuration not to be checked
  }),
}))
// 动态计算横向滚动宽度
const totalWidth = computed(() => {
  return allColumns.value.reduce((acc, column) => acc + (column.width || 0), 0)
})

const displayArray = ref(['channelCategoryType', 'channelStatus', 'channelType'])
const openModal = ref(false)
function createChannel() {
  modalType.value = 'create'
  openModal.value = true
}
const categoryManagementModal = ref(false)
const batchAdjustChannelModal = ref(false)
function categoryManagement() {
  categoryManagementModal.value = true
}
function batchAdjustChannel() {
  if (selectedRowsKeys.value.length === 0) {
    messageService.error(`请选择要调整的渠道`)
    return
  }
  batchAdjustChannelModal.value = true
  propCategoryId.value = 0
  // 设置子组件modalType
  modalType.value = 'adjust'
}
const loading = ref(false)
// 分页参数
const pagination = ref({
  current: 1,
  pageSize: 50,
  total: 0,
  showSizeChanger: true,
  showTotal: total => `共 ${total} 条`,
  pageSizeOptions: ['5', '10', '20', '50'],
  hideOnSinglePage: true,
  showQuickJumper: true,
})
// 获取渠道列表
async function getData(id, type) {
  loading.value = true
  try {
    const res = await getChannelPCPageApi({
      'pageRequestModel': {
        'needTotal': true,
        'pageSize': pagination.value.pageSize,
        'pageIndex': pagination.value.current,
        'skipCount': (pagination.value.current - 1) * pagination.value.pageSize,
      },
      'queryModel': {
        // 应用筛选条件
        ...(queryParams.value.isDefault !== undefined && { 'isDefault': queryParams.value.isDefault }),
        ...(queryParams.value.isDisabled !== undefined && { 'isDisabled': queryParams.value.isDisabled }),
        ...(queryParams.value.categoryId !== undefined && { 'channelTypeIds': queryParams.value.categoryId }),
      },
      'sortModel': {
        'byCreatedTime': 2,
      },
    })
    dataSource.value = res.result.map(item => ({
      ...item,
      categoryName: findChannelCategoryById(item.categoryId)?.categoryName || '-',
    }))
    // console.log('dataSource.value: ', dataSource.value);
    pagination.value.total = res.total
    loading.value = false
    allFilterRef.value?.clearQuickFilter(id, type)
  }
  catch (error) {
    loading.value = false
  }
}
const getChannelList = getData
const handleDisableChannel = debounce(async (record) => {
  try {
    const res = await updateChannelStatusApi({
      id: record.id,
      isDisabled: 1,
      uuid: record.uuid,
      version: record.version,
    })
    if (res.code === 200) {
      messageService.success('停用成功')
      getChannelList(null, null)
    }
  }
  catch (error) {
    messageService.error('停用失败')
  }
}, 300, { leading: true, trailing: false })
const handleEnableChannel = debounce(async (record) => {
  try {
    const res = await updateChannelStatusApi({
      id: record.id,
      isDisabled: 0,
      uuid: record.uuid,
      version: record.version,
    })
    if (res.code === 200) {
      messageService.success('启用成功')
      getChannelList(null, null)
    }
  }
  catch (error) {
    messageService.error('启用失败')
  }
}, 300, { leading: true, trailing: false })
// 处理分页变化
function handleTableChange(pag) {
  pagination.value.current = pag.current
  pagination.value.pageSize = pag.pageSize
  getChannelList(null, null)
}
// 传入id 查找渠道分类
function findChannelCategoryById(id) {
  return channelCategoryList.value.find(item => item.id === id)
}
onMounted(() => {
  // 重置查询参数
  queryParams.value = {
    pageNo: 1,
    pageSize: 50,
  }

  // 获取渠道列表
  getChannelList(null, null)
  // 获取渠道分类
  getChannelCategoryList()
})
// 创建渠道
async function createChannelFun(formState) {
  if (modalType.value === 'create') {
    try {
      if (formState.isClassification == '1') {
        formState.categoryId = 0
      }
      const res = await createChannelApi(formState)
      if (res.code === 200) {
        messageService.success('创建成功')
        getChannelList(null, null)
        openModal.value = false
      }
    }
    catch (error) {
      messageService.error('创建失败')
    }
  }
  else {
    try {
      if (formState.categoryId == null) {
        formState.categoryId = 0
      }
      const res = await updateChannelApi(formState)
      if (res.code === 200) {
        messageService.success('编辑成功')
        getChannelList(null, null)
        openModal.value = false
      }
    }
    catch (error) {
      messageService.error('编辑失败')
    }
  }
}
const modalType = ref('create')
const data = ref({})
// 编辑渠道
function editChannel(record) {
  if (record.categoryId == 0) {
    record.categoryId = undefined
  }
  openModal.value = true
  modalType.value = 'edit'
  data.value = record
}
// 创建一个防抖的数据获取函数
const debouncedGetData = debounce((id, type) => {
  getData(id, type)
}, 300)

// 过滤器字段映射
const filterFieldMapping = {
  channelTypeFilter: 'isDefault',
  channelStatusFilter: 'isDisabled',
  channelCategoryFilter: 'categoryId',
}

// 生成所有过滤器的更新处理器
const filterUpdateHandlers = computed(() => {
  const handlers = {}
  Object.entries(filterFieldMapping).forEach(([eventKey, fieldName]) => {
    handlers[`update:${eventKey}`] = (val, isClearAll, id, type) =>
      handleFilterUpdate({ [fieldName]: val }, isClearAll, id, type)
  })
  return handlers
})

// 统一处理筛选条件变化
const handleFilterUpdate = debounce((updates, isClearAll = false, id, type) => {
  if (isClearAll) {
    // 如果是清空所有，重置所有查询条件
    Object.keys(queryParams.value).forEach((key) => {
      if (key !== 'pageNo' && key !== 'pageSize') {
        queryParams.value[key] = undefined
      }
    })
  }
  else {
    // 如果是单个更新，只更新对应的条件
    Object.assign(queryParams.value, updates)
  }

  pagination.value.current = 1
  getChannelList(id, type)
}, 300, { leading: true, trailing: false })

const channelCategoryList = ref([])
// 获取渠道分类
async function getChannelCategoryList() {
  const res = await getChannelCategoryListApi()
  channelCategoryList.value = res.result
}
// 确认调整
async function adjustChannel(categoryId) {
  // console.log('categoryId: ', categoryId);
  try {
    // 调用接口
    const res = await adjustChannelApi({
      categoryId,
      channelIds: selectedRowsData.value.map(item => item.id),
    })
    if (res.code === 200) {
      messageService.success('调整成功')
      // 清除表格选中
      selectedRowsKeys.value = []
      selectedRowsData.value = []
      // 重新获取数据
      getChannelList(null, null)
      batchAdjustChannelModal.value = false
    }
  }
  catch (error) {
    messageService.error('调整失败')
  }
}
const propCategoryId = ref(0)
function handleAdjustChannel(record) {
  // console.log('record: ', record);
  // 打开批量调整渠道弹窗
  batchAdjustChannelModal.value = true
  // 设置子组件categoryId
  propCategoryId.value = record.categoryId
  // 设置子组件选中数据
  selectedRowsData.value = [record]
}
</script>

<template>
  <div class="tab-content">
    <all-filter
      ref="allFilterRef" :display-array="displayArray" :is-quick-show="false"
      :channel-categories="channelCategoryList"
      v-on="filterUpdateHandlers"
    />
    <div class="tab-table">
      <div class="table-title flex justify-between">
        <div class="total">
          共{{ pagination.total }}个渠道
        </div>
        <div class="edit flex">
          <a-button type="primary" class="mr-2" @click="createChannel">
            创建渠道
          </a-button>
          <a-button class="mr-2" @click="categoryManagement">
            分类管理
          </a-button>
          <a-button class="mr-2" @click="batchAdjustChannel">
            批量调整渠道
          </a-button>
        </div>
      </div>
      <div class="table-content mt-2">
        <a-table
          row-key="id" :data-source="dataSource" :pagination="pagination" :columns="allColumns"
          :row-selection="rowSelection" :scroll="{ x: totalWidth }" :loading="loading" @change="handleTableChange"
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'isDisabled'">
              <span
                v-if="!record.isDisabled"
                class="text-#06f text-12px bg-#e6f0ff rounded-full px-2.5 py-2px"
              >启用中</span>
              <span
                v-if="record.isDisabled"
                class="text-#f33 text-12px bg-#ffe6e6 rounded-full px-2.5 py-2px"
              >已停用</span>
            </template>
            <!-- 成交转化数 -->
            <template v-if="column.key === 'dealTransformCount'">
              <span v-if="record.dealTransformCount"> {{ record.dealTransformCount }}</span>
              <span v-if="!record.dealTransformCount">0</span>
            </template>
            <!-- 成交转化率 -->
            <template v-if="column.key === 'dealTransformRate'">
              <span v-if="record.dealTransformRate"> {{ record.dealTransformRate }}</span>
              <span v-if="!record.dealTransformRate" class="font800">0%</span>
            </template>
            <!-- 渠道分类 -->
            <template v-if="column.key === 'categoryName'">
              <span v-if="record.categoryName"> {{ record.categoryName }}</span>
              <span v-if="!record.categoryName">-</span>
            </template>
            <template v-if="column.key === 'isDefault'">
              <span v-if="record.isDefault"> <a-badge status="default" />系统默认</span>
              <span v-if="!record.isDefault"> <a-badge status="processing" />自定义</span>
            </template>
            <!-- 备注 -->
            <template v-if="column.key === 'remark'">
              <span v-if="record.remark"> {{ record.remark }}</span>
              <span v-if="!record.remark">-</span>
            </template>
            <template v-if="column.key === 'action' && !record.isDefault">
              <a-space :size="15">
                <a @click="editChannel(record)">编辑</a>
                <a @click="handleAdjustChannel(record)">调整</a>
                <a v-if="!record.isDisabled" @click="handleDisableChannel(record)">停用</a>
                <a v-if="record.isDisabled" @click="handleEnableChannel(record)">启用</a>
              </a-space>
            </template>
            <template v-if="column.key === 'action' && record.isDefault">
              <div class="text-#bbb text-14px cursor-not-allowed">
                系统默认，不支持操作
              </div>
            </template>
          </template>
        </a-table>
      </div>
    </div>
  </div>
  <CreateChannelModal
    v-model:open="openModal" :type="modalType" :data="data" :category-list="channelCategoryList"
    @on-create-channel="createChannelFun"
  />
  <CategoryManagementModal
    v-model:open="categoryManagementModal" :channel-category-list="channelCategoryList"
    @update:channel-category-list="getChannelCategoryList"
  />
  <BatchAdjustChannelModal
    v-model:open="batchAdjustChannelModal" :channel-category-list="channelCategoryList"
    :selected-rows-data="selectedRowsData" :modal-type="modalType" :prop-category-id="propCategoryId" @update:adjust-channel="adjustChannel"
  />
</template>

<style lang="less" scoped>
.tab-content {

  .tab-table {
    background: #fff;
    margin-top: 8px;
    padding: 12px;
    border-radius: 12px;

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
        margin-top: 1px;
        position: absolute;
        width: 4px;
      }
    }

    .intention {
      display: flex;
      align-items: center;

      .statusTag {
        width: 52px;
        height: 24px;
        background-color: rgb(255, 245, 230);
        color: rgb(255, 153, 0);
        border-radius: 100px;
        display: flex;
        align-items: center;
        justify-content: center;
        font-size: 12px;
        font-family: PingFangSC-Regular, PingFang SC, sans-serif;
        font-weight: 600;
      }
    }

    .intentionTag {
      display: inline-block;
      width: 6px;
      height: 6px;
      background: #f33;
      border-radius: 100px;
      margin-right: 3px;
    }

    .action {
      a {
        color: var(--pro-ant-color-primary);
        ;
      }
    }
  }
}
</style>
