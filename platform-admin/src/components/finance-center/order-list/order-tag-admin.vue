<script setup>
import { computed, ref } from 'vue'
import { createOrderTagApi, getOrderTagListPagedApi, updateOrderTagApi } from '@/api/finance-center/order-tag'
import messageService from '@/utils/messageService'

const loading = ref(false)
const dataSource = ref([])
const displayArray = ref(['enableStatus'])
const pagination = ref({
  current: 1,
  pageSize: 50,
  total: 0,
  showSizeChanger: false,
  showTotal: total => `共 ${total} 条`,
})

const enableFilter = ref(1)
const saving = ref(false)
const modalOpen = ref(false)
const editingRecord = ref(null)
const tagName = ref('')
const switchLoadingMap = ref({})
const filterUpdateHandlers = computed(() => ({
  'update:enableStatusFilter': (val) => {
    enableFilter.value = val ?? undefined
    pagination.value.current = 1
    fetchOrderTags()
  },
}))

const columns = [
  {
    title: '订单标签',
    dataIndex: 'name',
    key: 'name',
  },
  {
    title: '启用状态',
    dataIndex: 'enable',
    key: 'enable',
    width: 180,
  },
  {
    title: '操作',
    dataIndex: 'action',
    key: 'action',
    width: 120,
  },
]

const enabledCount = computed(() => dataSource.value.filter(item => item.enable).length)
const modalTitle = computed(() => editingRecord.value ? '编辑标签' : '新建标签')

async function fetchOrderTags() {
  try {
    loading.value = true
    const res = await getOrderTagListPagedApi({
      queryModel: enableFilter.value === undefined ? {} : { enable: enableFilter.value === 1 },
      sortModel: {},
      pageRequestModel: {
        pageSize: pagination.value.pageSize,
        pageIndex: pagination.value.current,
        needTotal: true,
        skipCount: (pagination.value.current - 1) * pagination.value.pageSize,
      },
    })
    if (res.code === 200) {
      dataSource.value = Array.isArray(res.result?.list) ? res.result.list : []
      pagination.value.total = res.result?.total || 0
      return
    }
    messageService.error(res.message || '获取订单标签失败')
  }
  catch (error) {
    console.error('获取订单标签失败:', error)
    messageService.error('获取订单标签失败')
  }
  finally {
    loading.value = false
  }
}

function handleTableChange(pag) {
  pagination.value.current = pag.current
  fetchOrderTags()
}

function openCreateModal() {
  editingRecord.value = null
  tagName.value = ''
  modalOpen.value = true
}

function openEditModal(record) {
  editingRecord.value = record
  tagName.value = record.name || ''
  modalOpen.value = true
}

async function handleModalOk() {
  const name = String(tagName.value || '').trim()
  if (!name) {
    messageService.error('请输入标签名称')
    return
  }
  try {
    saving.value = true
    if (editingRecord.value) {
      const res = await updateOrderTagApi({
        id: editingRecord.value.id,
        name,
      })
      if (res.code !== 200) {
        messageService.error(res.message || '编辑标签失败')
        return
      }
      messageService.success('编辑标签成功')
    }
    else {
      const res = await createOrderTagApi({ name })
      if (res.code !== 200) {
        messageService.error(res.message || '新建标签失败')
        return
      }
      messageService.success('新建标签成功')
    }
    modalOpen.value = false
    await fetchOrderTags()
  }
  catch (error) {
    console.error('保存订单标签失败:', error)
    messageService.error('保存订单标签失败')
  }
  finally {
    saving.value = false
  }
}

async function handleEnableChange(record, checked) {
  const previous = !!record.enable
  record.enable = checked
  switchLoadingMap.value = { ...switchLoadingMap.value, [record.id]: true }
  try {
    const res = await updateOrderTagApi({
      id: record.id,
      enable: checked,
    })
    if (res.code !== 200) {
      record.enable = previous
      messageService.error(res.message || '更新启用状态失败')
      return
    }
    messageService.success(checked ? '已启用标签' : '已停用标签')
    await fetchOrderTags()
  }
  catch (error) {
    console.error('更新启用状态失败:', error)
    record.enable = previous
    messageService.error('更新启用状态失败')
  }
  finally {
    switchLoadingMap.value = { ...switchLoadingMap.value, [record.id]: false }
  }
}

fetchOrderTags()
</script>

<template>
  <div>
    <div class="filter-wrap bg-white pl-3 pr-3 rounded-lb-4 rounded-rb-4">
      <all-filter
        :display-array="displayArray"
        :default-enable-status="enableFilter"
        :is-quick-show="false"
        v-on="filterUpdateHandlers"
      />
    </div>

    <div class="student-list mt-3 pt-3 pb-3 pl-6 pr-6 bg-white rounded-4">
      <div class="tab-table">
        <div class="table-title flex justify-between">
          <div class="total">
            总计 {{ pagination.total }} 个标签，{{ enabledCount }} 个标签已启用
          </div>
          <div class="edit flex">
            <a-button type="primary" @click="openCreateModal">
              创建标签
            </a-button>
          </div>
        </div>

        <div class="table-content mt-2">
          <a-table
            row-key="id"
            :data-source="dataSource"
            :pagination="pagination"
            :columns="columns"
            :loading="loading"
            size="small"
            @change="handleTableChange"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'enable'">
                <a-switch
                  :checked="record.enable"
                  :loading="!!switchLoadingMap[record.id]"
                  @change="handleEnableChange(record, $event)"
                />
              </template>
              <template v-else-if="column.key === 'action'">
                <a class="font500" @click="openEditModal(record)">编辑</a>
              </template>
            </template>
          </a-table>
        </div>
      </div>
    </div>

    <a-modal
      v-model:open="modalOpen"
      :title="modalTitle"
      :confirm-loading="saving"
      ok-text="保存"
      cancel-text="取消"
      @ok="handleModalOk"
    >
      <a-form layout="vertical">
        <a-form-item label="标签名称" required>
          <a-input v-model:value="tagName" placeholder="请输入标签名称" :maxlength="20" />
        </a-form-item>
      </a-form>
    </a-modal>
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
</style>
