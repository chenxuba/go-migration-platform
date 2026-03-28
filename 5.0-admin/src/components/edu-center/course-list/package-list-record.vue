<script setup>
import { CloseOutlined, DeleteOutlined, DownOutlined } from '@ant-design/icons-vue'
import { computed, ref } from 'vue'
import { debounce } from 'lodash-es'
import dayjs from 'dayjs'
import { batchSaleStatusApi, getCoursePropertyOptionsApi } from '~@/api/edu-center/course-list'
import { getProductPackagePagedListApi, getProductPackageStatisticsApi } from '@/api/edu-center/product-package'
import { useCourseAttribute } from '@/composables/useCourseAttribute'
import { useTableColumns } from '@/composables/useTableColumns'
import messageService from '@/utils/messageService'

const saleStatusOptions = [
  { id: 1, value: '在售' },
  { id: 0, value: '停售' },
]

const yesNoOptions = [
  { id: 1, value: '是' },
  { id: 0, value: '否' },
]

const propertyNameOrder = ['科目', '学季', '学年', '年级', '班型']

const { enabledCourseProperties, getEnabledCourseProperties } = useCourseAttribute()
const propertyOptionsMap = ref({})
const loading = ref(false)
const dataSource = ref([])
const selectedRowKeys = ref([])
const selectedRows = ref([])
const createDrawerOpen = ref(false)
const afterCreateModalVisible = ref(false)
const afterCreateStopSale = ref(false)
const createdPackageProductIds = ref([])

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

const dynamicColumns = computed(() => {
  const ordered = [...enabledCourseProperties.value].sort((a, b) => {
    const aIndex = propertyNameOrder.includes(a.name) ? propertyNameOrder.indexOf(a.name) : propertyNameOrder.length + 1
    const bIndex = propertyNameOrder.includes(b.name) ? propertyNameOrder.indexOf(b.name) : propertyNameOrder.length + 1
    if (aIndex !== bIndex)
      return aIndex - bIndex
    return a.id - b.id
  })
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

const selectedFilterList = computed(() => {
  const list = []
  if (filterState.value.onlineSale !== undefined) {
    list.push({
      key: 'onlineSale',
      label: '售卖状态',
      value: filterState.value.onlineSale ? '在售' : '停售',
    })
  }
  if (filterState.value.isOnlineSaleMicoSchool !== undefined) {
    list.push({
      key: 'isOnlineSaleMicoSchool',
      label: '是否开启微校售卖',
      value: filterState.value.isOnlineSaleMicoSchool ? '是' : '否',
    })
  }
  if (filterState.value.isShowMicoSchool !== undefined) {
    list.push({
      key: 'isShowMicoSchool',
      label: '是否开启微校展示',
      value: filterState.value.isShowMicoSchool ? '是' : '否',
    })
  }
  if (filterState.value.packageName.trim()) {
    list.push({
      key: 'packageName',
      label: '套餐名称',
      value: filterState.value.packageName.trim(),
    })
  }
  enabledCourseProperties.value.forEach((property) => {
    const current = filterState.value.propertyFilters[property.id]
    if (!current)
      return
    const option = (propertyOptionsMap.value[property.id] || []).find(item => String(item.id) === String(current))
    list.push({
      key: `property_${property.id}`,
      label: property.name,
      value: option?.name || String(current),
    })
  })
  return list
})

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

const debouncedRefresh = debounce(() => {
  pagination.value.current = 1
  refreshList()
}, 300)

async function loadPropertyOptions() {
  await getEnabledCourseProperties()
  await Promise.all(enabledCourseProperties.value.map(async (property) => {
    const res = await getCoursePropertyOptionsApi({ propertyId: property.id })
    if (res.code === 200) {
      propertyOptionsMap.value[property.id] = res.result || []
    }
  }))
}

function clearFilter(key) {
  if (key === 'onlineSale' || key === 'isOnlineSaleMicoSchool' || key === 'isShowMicoSchool' || key === 'packageName') {
    filterState.value[key] = key === 'packageName' ? '' : undefined
  }
  if (key.startsWith('property_')) {
    const propertyId = key.replace('property_', '')
    delete filterState.value.propertyFilters[propertyId]
  }
  debouncedRefresh()
}

function clearAllFilters() {
  filterState.value.onlineSale = undefined
  filterState.value.isOnlineSaleMicoSchool = undefined
  filterState.value.isShowMicoSchool = undefined
  filterState.value.packageName = ''
  filterState.value.propertyFilters = {}
  debouncedRefresh()
}

function handleSearchInput(value) {
  filterState.value.packageName = value || ''
  debouncedRefresh()
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

const rowSelection = computed(() => ({
  selectedRowKeys: selectedRowKeys.value,
  onChange: (keys, rows) => {
    selectedRowKeys.value = keys
    selectedRows.value = rows
  },
}))

loadPropertyOptions().then(() => refreshList())
</script>

<template>
  <div class="tab-content">
    <div class="filter-wrap bg-white rounded-4 px-4 py-3">
      <div class="filter-section">
        <div class="standard-filters">
          <span class="section-title">筛选条件：</span>
          <checkbox-filter v-model:checked-values="filterState.onlineSale" :options="saleStatusOptions" label="售卖状态" type="radio" category="noSearchRadio" @radio-change="debouncedRefresh" />
          <checkbox-filter v-model:checked-values="filterState.isOnlineSaleMicoSchool" :options="yesNoOptions" label="是否开启微校售卖" type="radio" category="noSearchRadio" @radio-change="debouncedRefresh" />
          <checkbox-filter v-model:checked-values="filterState.isShowMicoSchool" :options="yesNoOptions" label="是否开启微校展示" type="radio" category="noSearchRadio" @radio-change="debouncedRefresh" />
          <checkbox-filter
            v-for="property in [...enabledCourseProperties].sort((a, b) => {
              const aIndex = propertyNameOrder.includes(a.name) ? propertyNameOrder.indexOf(a.name) : propertyNameOrder.length + 1
              const bIndex = propertyNameOrder.includes(b.name) ? propertyNameOrder.indexOf(b.name) : propertyNameOrder.length + 1
              if (aIndex !== bIndex) return aIndex - bIndex
              return a.id - b.id
            })"
            :key="property.id"
            v-model:checked-values="filterState.propertyFilters[property.id]"
            :options="(propertyOptionsMap[property.id] || []).map(item => ({ id: item.id, value: item.name }))"
            :label="property.name"
            type="radio"
            category="noSearchRadio"
            @radio-change="debouncedRefresh"
          />
        </div>
        <div class="search-wrap">
          <div class="label">
            套餐名称
          </div>
          <a-input
            v-model:value="filterState.packageName"
            placeholder="请输入套餐名称"
            class="search-input"
            allow-clear
            @input="handleSearchInput($event.target.value)"
          />
        </div>
      </div>
      <div v-if="selectedFilterList.length > 0" class="selected-conditions">
        <span class="section-title">已选条件：</span>
        <div class="condition-tags">
          <a-popconfirm title="确定要清空所有条件吗？" @confirm="clearAllFilters">
            <a-tag color="red" class="clear-all mb-2">
              清空已选
              <DeleteOutlined class="text-3 ml-4px mt-0.6px" />
            </a-tag>
          </a-popconfirm>
          <a-tag v-for="item in selectedFilterList" :key="item.key" color="blue" class="condition-tag mb-2">
            <div class="tag-content">
              <span class="condition-label">{{ item.label }}：</span>
              <span class="value-item">
                {{ item.value }}
                <CloseOutlined class="close-icon" @click.stop="clearFilter(item.key)" />
              </span>
            </div>
          </a-tag>
        </div>
      </div>
    </div>

    <div class="tab-table mt-3 bg-white rounded-4 px-4 py-3">
      <div class="table-title flex justify-between items-center">
        <div class="total">
          当前 {{ stats.totalCount }} 个套餐，{{ stats.onSaleCount }} 个在售
        </div>
        <div class="edit ml10px flex overflow-x-auto">
          <a-button class="mr-2">
            批量操作
            <DownOutlined :style="{ fontSize: '10px' }" />
          </a-button>
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
  </div>
</template>

<style scoped lang="less">
.filter-wrap {
  .filter-section {
    display: flex;
    justify-content: space-between;
    gap: 16px;
  }

  .standard-filters {
    display: flex;
    flex-wrap: wrap;
    align-items: flex-start;
  }

  .section-title {
    white-space: nowrap;
    color: #222;
    margin-right: 12px;
    padding-top: 4px;
  }
}

.search-wrap {
  display: flex;
  align-items: center;
  margin-left: auto;

  .label {
    border: 1px solid #f0f0f0;
    border-right: 0;
    border-radius: 8px 0 0 8px;
    height: 32px;
    line-height: 32px;
    padding: 0 14px;
    white-space: nowrap;
  }

  .search-input {
    width: 320px;
  }
}

:deep(.search-input .ant-input) {
  border-radius: 0 8px 8px 0 !important;
}

.selected-conditions {
  display: flex;
  align-items: flex-start;
  margin-top: 8px;
}

.condition-tags {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
}

.clear-all {
  cursor: pointer;
  display: flex;
  align-items: center;
}

.condition-tag {
  border-radius: 4px;
}

.tag-content {
  display: flex;
  align-items: center;
}

.condition-label {
  color: #4b5563;
}

.value-item {
  display: inline-flex;
  align-items: center;
}

.close-icon {
  margin-left: 6px;
  font-size: 12px;
  cursor: pointer;
  color: rgba(92, 92, 92, 0.45);
}

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
</style>
