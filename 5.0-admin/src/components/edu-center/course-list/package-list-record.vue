<script setup>
import { DownOutlined, ExclamationCircleOutlined } from '@ant-design/icons-vue'
import { computed, ref } from 'vue'
import { debounce } from 'lodash-es'
import dayjs from 'dayjs'
import { batchSaleStatusApi } from '~@/api/edu-center/course-list'
import {
  deleteProductPackageApi,
  getProductPackagePagedListApi,
  getProductPackageStatisticsApi,
  updateProductPackageEnrollEditApi,
  updateProductPackageMicroSchoolRulesApi,
  updateProductPackageSaleStatusApi,
} from '@/api/edu-center/product-package'
import { useCourseAttribute } from '@/composables/useCourseAttribute'
import { useTableColumns } from '@/composables/useTableColumns'
import messageService from '@/utils/messageService'

const propertyNameOrder = ['科目', '学季', '学年', '年级', '班型']

const displayArray = ref(['sellStatus', 'isMicroSchoolSale', 'isMicroSchoolDisplay', 'courseAttribute'])

const { enabledCourseProperties, getEnabledCourseProperties } = useCourseAttribute()
const loading = ref(false)
const dataSource = ref([])
const selectedRowKeys = ref([])
const selectedRows = ref([])
const createDrawerOpen = ref(false)
const afterCreateModalVisible = ref(false)
const afterCreateStopSale = ref(false)
const createdPackageProductIds = ref([])
const batchSaleStatusModalOpen = ref(false)
const batchMicroSchoolModalOpen = ref(false)
const batchEnrollEditModalOpen = ref(false)
const batchDeleteModalOpen = ref(false)
const batchRunningModalOpen = ref(false)
const batchRunning = ref(false)
const batchProgressCurrent = ref(0)
const batchProgressTotal = ref(0)
const batchProgressTitle = ref('')
const batchSaleStatusValue = ref(true)
const batchMicroSchoolShow = ref(false)
const batchMicroSchoolSale = ref(false)
const batchAllowEditWhenEnroll = ref(false)

const filterState = ref({
  onlineSale: undefined,
  isOnlineSaleMicoSchool: undefined,
  isShowMicoSchool: undefined,
  packageName: '',
  propertyFilters: {},
})

const stats = ref({
  totalCount: 0,
  onSaleCount: 0,
})

const pagination = ref({
  current: 1,
  pageSize: 50,
  total: 0,
  showSizeChanger: true,
  showQuickJumper: true,
  showTotal: total => `共 ${total} 条`,
})

const baseColumns = [
  { title: '套餐名称', dataIndex: 'name', key: 'name', fixed: 'left', width: 220, required: true },
  { title: '是否关联总部模板', dataIndex: 'isSyncOrgProductPackage', key: 'isSyncOrgProductPackage', width: 180 },
  { title: '售卖状态', dataIndex: 'onlineSale', key: 'onlineSale', width: 130 },
  { title: '是否开启微校售卖', dataIndex: 'isOnlineSaleMicoSchool', key: 'isOnlineSaleMicoSchool', width: 160 },
  { title: '是否开启微校展示', dataIndex: 'isShowMicoSchool', key: 'isShowMicoSchool', width: 160 },
  { title: '总销量', dataIndex: 'sale', key: 'sale', width: 120 },
  { title: '最近更新时间', dataIndex: 'updatedTime', key: 'updatedTime', width: 180 },
]

function sortCourseProperties(list) {
  return [...list].sort((a, b) => {
    const aIndex = propertyNameOrder.includes(a.name) ? propertyNameOrder.indexOf(a.name) : propertyNameOrder.length + 1
    const bIndex = propertyNameOrder.includes(b.name) ? propertyNameOrder.indexOf(b.name) : propertyNameOrder.length + 1
    if (aIndex !== bIndex)
      return aIndex - bIndex
    return a.id - b.id
  })
}

const orderedCourseAttributes = computed(() => sortCourseProperties(enabledCourseProperties.value))

const dynamicColumns = computed(() => {
  const ordered = sortCourseProperties(enabledCourseProperties.value)
  return ordered.map(item => ({
    title: item.name,
    dataIndex: `property_${item.id}`,
    key: `property_${item.id}`,
    width: 140,
    isDynamic: true,
    propertyId: item.id,
    propertyName: item.name,
  }))
})

const allColumns = computed(() => [
  ...baseColumns,
  ...dynamicColumns.value,
  { title: '操作', dataIndex: 'action', key: 'action', fixed: 'right', width: 180 },
])

const { selectedValues, columnOptions, filteredColumns, totalWidth } = useTableColumns({
  storageKey: 'product-package-list-record',
  allColumns,
  excludeKeys: ['action'],
  defaultSelectedKeys: [
    'name',
    'isSyncOrgProductPackage',
    'onlineSale',
    'isOnlineSaleMicoSchool',
    'isShowMicoSchool',
    'sale',
    'updatedTime',
  ],
})

function formatDate(value) {
  if (!value)
    return '-'
  return dayjs(value).format('YYYY-MM-DD HH:mm')
}

function getBooleanText(value) {
  return value ? '是' : '否'
}

function getBooleanClass(value) {
  return value ? 'status-yes' : 'status-no'
}

function getSaleStatusClass(value) {
  return value ? 'sale-on' : 'sale-off'
}

function getPropertyDisplay(record, propertyId) {
  const row = (record.extendProperties || []).find(item => String(item.productPackagePropertyId) === String(propertyId))
  return row?.productPackagePropertyValueName || '-'
}

function buildPropertyQueryRows() {
  return Object.entries(filterState.value.propertyFilters)
    .filter(([, value]) => value !== undefined && value !== null && value !== '')
    .map(([propertyId, value]) => ({
      productPackagePropertyId: String(propertyId),
      productPackagePropertyValue: String(value),
    }))
}

function buildQueryModel() {
  return {
    name: filterState.value.packageName.trim() || undefined,
    onlineSale: filterState.value.onlineSale,
    isOnlineSaleMicoSchool: filterState.value.isOnlineSaleMicoSchool,
    isShowMicoSchool: filterState.value.isShowMicoSchool,
    productPackageProperties: buildPropertyQueryRows(),
  }
}

async function fetchPackageList() {
  loading.value = true
  try {
    const res = await getProductPackagePagedListApi({
      pageRequestModel: {
        pageSize: pagination.value.pageSize,
        pageIndex: pagination.value.current,
        needTotal: true,
        skipCount: (pagination.value.current - 1) * pagination.value.pageSize,
      },
      sortModel: {},
      queryModel: buildQueryModel(),
    })
    dataSource.value = res.result?.list || []
    pagination.value.total = res.result?.total || 0
  }
  catch (error) {
    console.error('获取套餐列表失败:', error)
    messageService.error('获取套餐列表失败')
  }
  finally {
    loading.value = false
  }
}

async function fetchPackageStatistics() {
  try {
    const res = await getProductPackageStatisticsApi(buildQueryModel())
    stats.value = {
      totalCount: res.result?.totalCount || 0,
      onSaleCount: res.result?.onSaleCount || 0,
    }
  }
  catch (error) {
    console.error('获取套餐统计失败:', error)
  }
}

async function refreshList() {
  await Promise.all([
    fetchPackageList(),
    fetchPackageStatistics(),
  ])
}

const handleFilterUpdate = debounce((updates, isClearAll = false) => {
  if (isClearAll) {
    filterState.value.onlineSale = undefined
    filterState.value.isOnlineSaleMicoSchool = undefined
    filterState.value.isShowMicoSchool = undefined
    filterState.value.packageName = ''
    filterState.value.propertyFilters = {}
  }
  else {
    Object.entries(updates).forEach(([key, value]) => {
      filterState.value[key] = value
    })
  }
  pagination.value.current = 1
  refreshList()
}, 300, { leading: true, trailing: false })

const filterUpdateHandlers = computed(() => ({
  'update:sellStatusFilter': (val, isClearAll) => handleFilterUpdate({ onlineSale: val }, isClearAll),
  'update:isMicroSchoolSaleFilter': (val, isClearAll) => handleFilterUpdate({ isOnlineSaleMicoSchool: val }, isClearAll),
  'update:isMicroSchoolDisplayFilter': (val, isClearAll) => handleFilterUpdate({ isShowMicoSchool: val }, isClearAll),
  'update:courseAttributeFilter': (payload) => {
    const pid = payload?.itemId
    const next = { ...filterState.value.propertyFilters }
    if (payload?.value === undefined || payload?.value === null || payload?.value === '') {
      delete next[pid]
    }
    else {
      next[pid] = payload.value
    }
    filterState.value.propertyFilters = next
    pagination.value.current = 1
    refreshList()
  },
}))

function handleSearchInput(value) {
  filterState.value.packageName = value || ''
  pagination.value.current = 1
  refreshList()
}

async function init() {
  await getEnabledCourseProperties()
  await refreshList()
}

function handleTableChange(pag) {
  pagination.value.current = pag.current
  pagination.value.pageSize = pag.pageSize
  refreshList()
}

function handleCreated(payload) {
  createdPackageProductIds.value = payload?.productIds || []
  afterCreateStopSale.value = false
  afterCreateModalVisible.value = true
  refreshList()
}

async function handleAfterCreateConfirm() {
  try {
    if (afterCreateStopSale.value && createdPackageProductIds.value.length > 0) {
      await batchSaleStatusApi({
        courseIds: createdPackageProductIds.value,
        saleStatus: false,
      })
      messageService.success('已将套餐内商品设为停售')
    }
  }
  catch (error) {
    console.error('套餐内商品停售失败:', error)
    messageService.error('套餐内商品停售失败')
  }
  finally {
    afterCreateModalVisible.value = false
    createdPackageProductIds.value = []
    refreshList()
  }
}

function handleAction(action) {
  messageService.info(`${action}功能待实现`)
}

const batchProgressPercent = computed(() => {
  if (!batchProgressTotal.value) return 0
  return Math.min(100, Math.round((batchProgressCurrent.value / batchProgressTotal.value) * 100))
})

function getSelectedPackageIds() {
  return selectedRows.value.map(row => String(row.id)).filter(Boolean)
}

function ensurePackageSelected(actionLabel) {
  if (selectedRows.value.length === 0) {
    messageService.warning(`请选择要${actionLabel}的套餐`)
    return false
  }
  return true
}

function resetSelection() {
  selectedRows.value = []
  selectedRowKeys.value = []
}

function openBatchSaleStatusModal() {
  if (!ensurePackageSelected('批量设置售卖状态'))
    return
  batchSaleStatusValue.value = true
  batchSaleStatusModalOpen.value = true
}

function openBatchMicroSchoolModal() {
  if (!ensurePackageSelected('批量设置微校售卖规则'))
    return
  batchMicroSchoolShow.value = false
  batchMicroSchoolSale.value = false
  batchMicroSchoolModalOpen.value = true
}

function openBatchEnrollEditModal() {
  if (!ensurePackageSelected('批量设置报名时修改办理内容'))
    return
  batchAllowEditWhenEnroll.value = false
  batchEnrollEditModalOpen.value = true
}

function openBatchDeleteModal() {
  if (!ensurePackageSelected('批量删除'))
    return
  batchDeleteModalOpen.value = true
}

async function runSingleBatchAction(actionTitle, ids, worker) {
  batchProgressTitle.value = actionTitle
  batchProgressCurrent.value = 0
  batchProgressTotal.value = ids.length
  batchRunning.value = true
  batchRunningModalOpen.value = true

  const failedIds = []

  for (let index = 0; index < ids.length; index += 1) {
    const id = ids[index]
    try {
      await worker(id)
    }
    catch (error) {
      console.error(`${actionTitle}失败:`, error)
      failedIds.push(id)
    }
    finally {
      batchProgressCurrent.value = index + 1
    }
  }

  batchRunning.value = false
  batchRunningModalOpen.value = false

  if (failedIds.length === 0) {
    messageService.success(`${actionTitle}成功`)
  } else if (failedIds.length === ids.length) {
    messageService.error(`${actionTitle}失败`)
  } else {
    messageService.warning(`${actionTitle}部分成功，失败 ${failedIds.length} 个`)
  }

  resetSelection()
  await refreshList()
}

async function confirmBatchSaleStatus() {
  batchSaleStatusModalOpen.value = false
  const ids = getSelectedPackageIds()
  await runSingleBatchAction('批量设置售卖状态', ids, id =>
    updateProductPackageSaleStatusApi({
      id,
      onlineSale: batchSaleStatusValue.value,
    }))
}

async function confirmBatchMicroSchoolRules() {
  batchMicroSchoolModalOpen.value = false
  const ids = getSelectedPackageIds()
  await runSingleBatchAction('批量设置微校售卖规则', ids, id =>
    updateProductPackageMicroSchoolRulesApi({
      id,
      isShowMicoSchool: batchMicroSchoolShow.value,
      isOnlineSaleMicoSchool: batchMicroSchoolSale.value,
    }))
}

async function confirmBatchEnrollEdit() {
  batchEnrollEditModalOpen.value = false
  const ids = getSelectedPackageIds()
  await runSingleBatchAction('批量设置报名时修改办理内容', ids, id =>
    updateProductPackageEnrollEditApi({
      id,
      isAllowEditWhenEnroll: batchAllowEditWhenEnroll.value,
    }))
}

async function confirmBatchDelete() {
  batchDeleteModalOpen.value = false
  const ids = getSelectedPackageIds()
  await runSingleBatchAction('批量删除套餐', ids, id =>
    deleteProductPackageApi({ id }))
}

function onBatchActionClick({ key }) {
  switch (String(key)) {
    case 'sale-status':
      openBatchSaleStatusModal()
      break
    case 'micro-school-rules':
      openBatchMicroSchoolModal()
      break
    case 'enroll-edit':
      openBatchEnrollEditModal()
      break
    case 'delete':
      openBatchDeleteModal()
      break
    default:
      break
  }
}

const rowSelection = computed(() => ({
  selectedRowKeys: selectedRowKeys.value,
  onChange: (keys, rows) => {
    selectedRowKeys.value = keys
    selectedRows.value = rows
  },
}))

init()
</script>

<template>
  <div class="tab-content">
    <div class="filter-wrap bg-white pl-3 pr-3 rounded-4 rounded-tl-0 rounded-tr-0">
      <all-filter
        :display-array="displayArray"
        :is-show-search-input="true"
        search-label="套餐名称"
        search-placeholder="请输入..."
        :render-class-list-options="false"
        :course-attribute-list="orderedCourseAttributes"
        v-on="filterUpdateHandlers"
        @searchInputFun="handleSearchInput"
      />
    </div>

    <div class="tab-table mt-3 bg-white rounded-4 px-4 py-3">
      <div class="table-title flex justify-between items-center">
        <div class="total">
          当前 {{ stats.totalCount }} 个套餐，{{ stats.onSaleCount }} 个在售
        </div>
        <div class="edit ml10px flex overflow-x-auto">
          <a-dropdown class="mr-2">
            <template #overlay>
              <a-menu @click="onBatchActionClick">
                <a-menu-item key="sale-status">
                  批量设置售卖状态
                </a-menu-item>
                <a-menu-item key="micro-school-rules">
                  批量设置微校售卖规则
                </a-menu-item>
                <a-menu-item key="enroll-edit">
                  设置报名时修改办理内容
                </a-menu-item>
                <a-menu-item key="delete">
                  批量删除
                </a-menu-item>
              </a-menu>
            </template>
            <a-button class="mr-2">
              批量操作
              <DownOutlined :style="{ fontSize: '10px' }" />
            </a-button>
          </a-dropdown>
          <a-button type="primary" class="mr-2" @click="createDrawerOpen = true">
            创建套餐
          </a-button>
          <customize-code v-model:checked-values="selectedValues" :options="columnOptions" :total="allColumns.length - 1" :num="selectedValues.length" />
        </div>
      </div>
      <div class="table-content mt-2">
        <a-table
          :data-source="dataSource"
          :loading="loading"
          :pagination="pagination"
          :columns="filteredColumns"
          :row-selection="rowSelection"
          :scroll="{ x: totalWidth }"
          row-key="id"
          @change="handleTableChange"
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'name'">
              <clamped-text :text="record.name" :lines="1" />
            </template>
            <template v-if="column.key === 'isSyncOrgProductPackage'">
              <span class="dot-status" :class="getBooleanClass(record.isSyncOrgProductPackage)">{{ getBooleanText(record.isSyncOrgProductPackage) }}</span>
            </template>
            <template v-if="column.key === 'onlineSale'">
              <span class="sale-tag" :class="getSaleStatusClass(record.onlineSale)">{{ record.onlineSale ? '在售' : '停售' }}</span>
            </template>
            <template v-if="column.key === 'isOnlineSaleMicoSchool'">
              <span class="dot-status" :class="getBooleanClass(record.isOnlineSaleMicoSchool)">{{ getBooleanText(record.isOnlineSaleMicoSchool) }}</span>
            </template>
            <template v-if="column.key === 'isShowMicoSchool'">
              <span class="dot-status" :class="getBooleanClass(record.isShowMicoSchool)">{{ getBooleanText(record.isShowMicoSchool) }}</span>
            </template>
            <template v-if="column.key === 'sale'">
              {{ record.sale || 0 }}
            </template>
            <template v-if="column.key === 'updatedTime'">
              {{ formatDate(record.updatedTime) }}
            </template>
            <template v-if="String(column.key).startsWith('property_')">
              {{ getPropertyDisplay(record, column.propertyId) }}
            </template>
            <template v-if="column.key === 'action'">
              <div class="action-links">
                <a @click="handleAction('编辑')">编辑</a>
                <a @click="handleAction('分享')">分享</a>
                <a @click="handleAction('更多')">更多</a>
              </div>
            </template>
          </template>
        </a-table>
      </div>
    </div>

    <package-create-drawer v-model:open="createDrawerOpen" @created="handleCreated" />

    <a-modal
      v-model:open="afterCreateModalVisible"
      title="提示"
      ok-text="确定"
      cancel-text="取消"
      centered
      @ok="handleAfterCreateConfirm"
      @cancel="afterCreateModalVisible = false"
    >
      <div class="after-create-body">
        <div class="after-create-desc">
          套餐创建后，如果您希望套餐内的商品只允许通过套餐进行售卖，可以选择自动将本套餐内的商品设为停售
        </div>
        <div class="after-create-radio">
          <span>套餐内商品在售状态：</span>
          <a-radio-group v-model:value="afterCreateStopSale">
            <a-radio :value="false">
              不设为停售
            </a-radio>
            <a-radio :value="true">
              设为停售
            </a-radio>
          </a-radio-group>
        </div>
      </div>
    </a-modal>

    <a-modal
      v-model:open="batchSaleStatusModalOpen"
      centered
      title="批量设置售卖状态"
      :width="520"
      :closable="false"
      :mask-closable="false"
      ok-text="确定"
      cancel-text="取消"
      @ok="confirmBatchSaleStatus"
      @cancel="batchSaleStatusModalOpen = false"
    >
      <div class="batch-modal-body">
        <a-radio-group v-model:value="batchSaleStatusValue" class="batch-radio-group custom-radio">
          <a-radio :value="true">
            售卖
          </a-radio>
          <a-radio :value="false">
            停售
          </a-radio>
        </a-radio-group>
      </div>
    </a-modal>

    <a-modal
      v-model:open="batchMicroSchoolModalOpen"
      centered
      title="批量设置微校售卖规则"
      :width="620"
      :closable="false"
      :mask-closable="false"
      ok-text="确定"
      cancel-text="取消"
      @ok="confirmBatchMicroSchoolRules"
      @cancel="batchMicroSchoolModalOpen = false"
    >
      <div class="batch-micro-body">
        <div class="batch-switch-row">
          <span class="batch-switch-row__label">微校展示：</span>
          <a-switch v-model:checked="batchMicroSchoolShow" />
          <a-popover content="关闭后，套餐将不在微校展示" title="微校展示">
            <ExclamationCircleOutlined class="batch-switch-row__icon" />
          </a-popover>
        </div>
        <div class="batch-switch-row">
          <span class="batch-switch-row__label">微校售卖：</span>
          <a-switch v-model:checked="batchMicroSchoolSale" />
          <a-popover content="开启后，套餐可在微校售卖" title="微校售卖">
            <ExclamationCircleOutlined class="batch-switch-row__icon" />
          </a-popover>
        </div>
      </div>
    </a-modal>

    <a-modal
      v-model:open="batchEnrollEditModalOpen"
      centered
      title="批量设置报名时修改办理内容"
      :width="560"
      :closable="false"
      :mask-closable="false"
      ok-text="确定"
      cancel-text="取消"
      @ok="confirmBatchEnrollEdit"
      @cancel="batchEnrollEditModalOpen = false"
    >
      <div class="batch-modal-body">
        <a-radio-group v-model:value="batchAllowEditWhenEnroll" class="batch-radio-group custom-radio">
          <a-radio :value="true">
            允许修改
          </a-radio>
          <a-radio :value="false">
            禁止修改
          </a-radio>
        </a-radio-group>
      </div>
    </a-modal>

    <a-modal
      v-model:open="batchDeleteModalOpen"
      centered
      title="确定批量删除？"
      :width="560"
      :closable="false"
      :mask-closable="false"
      :footer="null"
      @cancel="batchDeleteModalOpen = false"
    >
      <div class="batch-delete-modal">
        <div class="batch-delete-modal__desc">
          套餐被删除后将无法恢复，是否确认删除套餐？
        </div>
        <div class="batch-delete-modal__footer">
          <a-button danger ghost @click="confirmBatchDelete">删除</a-button>
          <a-button type="primary" ghost @click="batchDeleteModalOpen = false">取消</a-button>
        </div>
      </div>
    </a-modal>

    <a-modal
      v-model:open="batchRunningModalOpen"
      centered
      :closable="false"
      :mask-closable="false"
      :keyboard="false"
      :footer="null"
      :width="560"
    >
      <div class="batch-progress-modal">
        <div class="batch-progress-modal__title">{{ batchProgressTitle }}</div>
        <div class="batch-progress-modal__desc">正在处理 {{ batchProgressCurrent }} / {{ batchProgressTotal }}</div>
        <a-progress :percent="batchProgressPercent" status="active" />
      </div>
    </a-modal>
  </div>
</template>

<style scoped lang="less">
.tab-content {
  margin: 0;

  .tab-table {
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
  }
}

.dot-status::before {
  content: "";
  display: inline-block;
  width: 8px;
  height: 8px;
  border-radius: 999px;
  margin-right: 6px;
  background: #d9d9d9;
}

.status-yes::before {
  background: #52c41a;
}

.sale-tag {
  border-radius: 999px;
  padding: 3px 10px;
  font-size: 12px;
}

.sale-on {
  color: #1677ff;
  background: #e6f4ff;
}

.sale-off {
  color: #8c8c8c;
  background: #f5f5f5;
}

.action-links {
  display: flex;
  gap: 14px;
  white-space: nowrap;
}

.after-create-body {
  padding-top: 8px;
}

.after-create-desc {
  color: #444;
  line-height: 1.8;
  margin-bottom: 18px;
}

.after-create-radio {
  display: flex;
  align-items: center;
  gap: 12px;
}

.batch-modal-body {
  padding: 8px 0 4px;
}

.batch-radio-group {
  display: flex;
  gap: 48px;
  flex-wrap: wrap;
}

.batch-micro-body {
  padding: 8px 0 4px;
}

.batch-switch-row {
  display: flex;
  align-items: center;
  gap: 14px;
  min-height: 48px;

}

.batch-switch-row__label {
  min-width: 96px;
  color: #222;
  font-size: 15px;
  font-weight: 500;
  text-align: right;
}

.batch-switch-row__icon {
  color: #6b7280;
  font-size: 18px;
}

.batch-delete-modal {
  padding: 4px 0 0;
}

.batch-delete-modal__desc {
  margin: 8px 0 28px;
  color: #666;
  font-size: 15px;
  line-height: 1.8;
}

.batch-delete-modal__footer {
  display: flex;
  justify-content: flex-end;
  gap: 16px;
}

.batch-progress-modal {
  padding: 8px 4px;
}

.batch-progress-modal__title {
  font-size: 18px;
  font-weight: 600;
  color: #222;
  margin-bottom: 10px;
}

.batch-progress-modal__desc {
  color: #666;
  margin-bottom: 16px;
}

.custom-radio ::v-deep(.ant-radio-wrapper:hover .ant-radio),
.custom-radio ::v-deep(.ant-radio:hover .ant-radio-inner),
.custom-radio ::v-deep(.ant-radio-input:focus + .ant-radio-inner) {
  border-color: var(--pro-ant-color-primary);
}

.custom-radio ::v-deep(.ant-radio-inner) {
  background-color: transparent;
  border-color: #d9d9d9;
}

.custom-radio ::v-deep(.ant-radio-checked .ant-radio-inner) {
  background-color: transparent;
  border-color: var(--pro-ant-color-primary);
}

.custom-radio ::v-deep(.ant-radio-inner::after) {
  background-color: var(--pro-ant-color-primary);
  transform: scale(0.5);
}
</style>
