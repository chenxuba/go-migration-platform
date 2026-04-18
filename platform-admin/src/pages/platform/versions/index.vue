<script setup lang="ts">
import type { TableColumnsType } from 'ant-design-vue'
import type { VersionItem } from '@/api/platform/versions'
import { PlusOutlined } from '@ant-design/icons-vue'
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { pageVersionsApi } from '@/api/platform/versions'
import messageService from '@/utils/messageService'
import { filterSystemDefaultVersions, sortVersionsByDisplayOrder, sortVersionsByDisplayOrderDesc } from '../shared/version-order'
import VersionFormModal from './components/version-form-modal.vue'

const loading = ref(false)
const keyword = ref('')
const dataSource = ref<VersionItem[]>([])
const highlightSource = ref<VersionItem[]>([])
const versionModalOpen = ref(false)
const editingVersionId = ref<number | null>(null)

const pagination = reactive({
  current: 1,
  pageSize: 20,
  total: 0,
  showSizeChanger: true,
  showTotal: (total: number) => `共 ${total} 个版本`,
})

const highlightVersions = computed(() => highlightSource.value.slice(0, 4))
const tableVersions = computed(() => sortVersionsByDisplayOrderDesc(dataSource.value))

const columns: TableColumnsType<VersionItem> = [
  {
    title: '版本方案',
    dataIndex: 'name',
    key: 'name',
    width: 280,
  },
  {
    title: '价格',
    dataIndex: 'price',
    key: 'price',
    width: 160,
  },
  {
    title: '菜单规模',
    dataIndex: 'menuCount',
    key: 'menuCount',
    width: 140,
    align: 'center' as const,
  },
  {
    title: '绑定机构',
    dataIndex: 'orgCount',
    key: 'orgCount',
    width: 140,
    align: 'center' as const,
  },
  {
    title: '最近更新',
    dataIndex: 'updateTime',
    key: 'updateTime',
    width: 180,
  },
  {
    title: '操作',
    key: 'action',
    width: 140,
    fixed: 'right' as const,
  },
]

function formatDateMinute(value?: string) {
  const raw = String(value || '').trim()
  if (!raw)
    return '--'
  return raw.length >= 16 ? raw.slice(0, 16) : raw
}

function formatPrice(value?: number) {
  const numeric = Number(value || 0)
  return numeric.toFixed(2)
}

function formatPriceLabel(value?: number) {
  const numeric = Number(value || 0)
  if (numeric <= 0)
    return '免费'
  return `¥${numeric.toFixed(2)}`
}

function getVersionRemark(value?: string) {
  const remark = String(value || '').trim()
  return remark || '未填写版本说明'
}

async function fetchHighlightVersions() {
  try {
    const res = await pageVersionsApi({
      current: 1,
      size: 200,
      type: 1,
    })

    if (res.code !== 200) {
      messageService.error(res.message || '获取版本卡片失败')
      return
    }

    highlightSource.value = sortVersionsByDisplayOrder(
      filterSystemDefaultVersions(Array.isArray(res.result) ? res.result : []),
    )
  }
  catch (error: any) {
    console.error('fetch highlight versions failed', error)
    messageService.error(error?.message || '获取版本卡片失败')
  }
}

async function fetchVersions() {
  loading.value = true
  try {
    const res = await pageVersionsApi({
      current: pagination.current,
      size: pagination.pageSize,
      name: keyword.value.trim() || undefined,
      type: 1,
    })

    if (res.code !== 200) {
      messageService.error(res.message || '获取版本列表失败')
      return
    }

    dataSource.value = sortVersionsByDisplayOrder(Array.isArray(res.result) ? res.result : [])
    pagination.total = Number(res.total || 0)
  }
  catch (error: any) {
    console.error('fetch versions failed', error)
    messageService.error(error?.message || '获取版本列表失败')
  }
  finally {
    loading.value = false
  }
}

function openCreateModal() {
  editingVersionId.value = null
  versionModalOpen.value = true
}

function openEditModal(record: Partial<VersionItem>) {
  editingVersionId.value = Number(record.id || 0) || null
  versionModalOpen.value = true
}

function handleSaved() {
  versionModalOpen.value = false
  editingVersionId.value = null
  fetchHighlightVersions()
  fetchVersions()
}

function handleTableChange(page: { current?: number, pageSize?: number }) {
  pagination.current = page.current || 1
  pagination.pageSize = page.pageSize || 20
  fetchVersions()
}

function handleSearch() {
  pagination.current = 1
  fetchVersions()
}

onMounted(() => {
  fetchHighlightVersions()
  fetchVersions()
})

watch(versionModalOpen, (open) => {
  if (!open)
    editingVersionId.value = null
})
</script>

<template>
  <div class="version-page">
    <div class="version-header">
      <div class="version-header__title">
        版本管理
      </div>

      <div class="version-header__actions">
        <a-input-search
          v-model:value="keyword"
          allow-clear
          placeholder="搜索版本名称"
          class="version-header__search"
          @search="handleSearch"
        />

        <a-button type="primary" @click="openCreateModal">
          <template #icon>
            <PlusOutlined />
          </template>
          新建版本
        </a-button>
      </div>
    </div>

    <div v-if="highlightVersions.length" class="version-highlights">
      <div
        v-for="(item, index) in highlightVersions"
        :key="item.id"
        class="version-highlight"
        :class="`version-highlight--${(index % 4) + 1}`"
      >
        <div class="version-highlight__top">
          <div class="version-highlight__name">
            {{ item.name || '--' }}
          </div>
          <div class="version-highlight__price">
            {{ formatPriceLabel(item.price) }}
          </div>
        </div>

        <a-tooltip
          :overlay-style="{ maxWidth: '320px', whiteSpace: 'normal' }"
        >
          <template #title>
            {{ getVersionRemark(item.remark) }}
          </template>
          <div class="version-highlight__remark">
            {{ getVersionRemark(item.remark) }}
          </div>
        </a-tooltip>

        <div class="version-highlight__stats">
          <span>{{ item.menuCount || 0 }} 项菜单</span>
          <span>{{ item.orgCount || 0 }} 家机构</span>
        </div>

        <div class="version-highlight__footer">
          <span>{{ formatDateMinute(item.updateTime) }}</span>
          <a-button type="link" class="version-highlight__link" @click="openEditModal(item)">
            编辑权限
          </a-button>
        </div>
      </div>
    </div>

    <div class="version-table-card">
      <div class="version-table-card__header">
        <div class="version-table-card__title">
          共 {{ pagination.total }} 个版本
        </div>
      </div>

      <a-table
        class="version-table"
        :columns="columns"
        :data-source="tableVersions"
        :loading="loading"
        :pagination="pagination"
        :scroll="{ x: 980 }"
        row-key="id"
        size="small"
        @change="handleTableChange"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'name'">
            <div class="version-name-cell">
              <div class="version-name-cell__title">
                {{ record.name || '--' }}
              </div>
              <a-tooltip
                :overlay-style="{ maxWidth: '320px', whiteSpace: 'normal' }"
              >
                <template #title>
                  {{ getVersionRemark(record.remark) }}
                </template>
                <div class="version-name-cell__remark">
                  {{ getVersionRemark(record.remark) }}
                </div>
              </a-tooltip>
            </div>
          </template>

          <template v-else-if="column.key === 'price'">
            <div class="metric-cell">
              <div class="metric-cell__value metric-cell__value--price">
                {{ formatPriceLabel(record.price) }}
              </div>
              <div class="metric-cell__label">
                标准售价
              </div>
            </div>
          </template>

          <template v-else-if="column.key === 'menuCount'">
            <div class="metric-cell metric-cell--center">
              <div class="metric-cell__value">
                {{ record.menuCount || 0 }}
              </div>
              <div class="metric-cell__label">
                项菜单
              </div>
            </div>
          </template>

          <template v-else-if="column.key === 'orgCount'">
            <div class="metric-cell metric-cell--center">
              <div class="metric-cell__value">
                {{ record.orgCount || 0 }}
              </div>
              <div class="metric-cell__label">
                家机构
              </div>
            </div>
          </template>

          <template v-else-if="column.key === 'updateTime'">
            <div class="metric-cell">
              <div class="metric-cell__value">
                {{ formatDateMinute(record.updateTime) }}
              </div>
              <div class="metric-cell__label">
                最近更新时间
              </div>
            </div>
          </template>

          <template v-else-if="column.key === 'action'">
            <a-button type="link" class="action-link" @click="openEditModal(record)">
              编辑版本
            </a-button>
          </template>
        </template>
      </a-table>
    </div>

    <VersionFormModal
      v-model:open="versionModalOpen"
      :version-id="editingVersionId"
      @saved="handleSaved"
    />
  </div>
</template>

<style scoped lang="less">
.version-page {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.version-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 6px 2px 0;
}

.version-header__title {
  color: #1f2329;
  font-size: 24px;
  font-weight: 700;
  line-height: 34px;
}

.version-header__actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.version-header__search {
  width: 340px;
}

.version-highlights {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 14px;
}

.version-highlight {
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 8px;
  min-height: 134px;
  padding: 14px 16px 12px;
  border: 1px solid #e9edf3;
  border-radius: 16px;
  background: #fff;
  overflow: hidden;
  box-shadow: 0 14px 30px rgba(15, 23, 42, 0.06);
}

.version-highlight::before {
  position: absolute;
  inset: 0 0 auto;
  height: 4px;
  content: "";
}

.version-highlight--1 {
  background: linear-gradient(180deg, rgba(22, 119, 255, 0.06) 0%, #fff 90px);
}

.version-highlight--1::before {
  background: linear-gradient(90deg, #1677ff 0%, #69b1ff 100%);
}

.version-highlight--2 {
  background: linear-gradient(180deg, rgba(14, 116, 144, 0.08) 0%, #fff 90px);
}

.version-highlight--2::before {
  background: linear-gradient(90deg, #0891b2 0%, #67e8f9 100%);
}

.version-highlight--3 {
  background: linear-gradient(180deg, rgba(249, 115, 22, 0.08) 0%, #fff 90px);
}

.version-highlight--3::before {
  background: linear-gradient(90deg, #f97316 0%, #fdba74 100%);
}

.version-highlight--4 {
  background: linear-gradient(180deg, rgba(15, 118, 110, 0.08) 0%, #fff 90px);
}

.version-highlight--4::before {
  background: linear-gradient(90deg, #0f766e 0%, #5eead4 100%);
}

.version-highlight__top,
.version-highlight__footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}

.version-highlight__name {
  min-width: 0;
  color: #1f2329;
  font-size: 17px;
  font-weight: 700;
  line-height: 24px;
}

.version-highlight__price {
  flex-shrink: 0;
  padding: 6px 12px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.72);
  color: #1f2329;
  font-size: 13px;
  font-weight: 700;
  line-height: 18px;
}

.version-highlight__remark {
  display: block;
  width: 100%;
  min-height: 20px;
  overflow: hidden;
  color: #667085;
  font-size: 13px;
  line-height: 20px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.version-highlight__stats {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  width: 100%;
  justify-content: flex-start;
  gap: 20px;
  color: #344054;
  font-size: 13px;
  line-height: 20px;
}

.version-highlight__stats span {
  padding: 0;
  border-radius: 0;
  background: transparent;
}

.version-highlight__footer {
  margin-top: auto;
  color: #98a2b3;
  font-size: 12px;
  line-height: 18px;
}

.version-highlight__link {
  padding: 0;
  font-weight: 600;
}

.version-table-card {
  border: 1px solid #e9edf3;
  border-radius: 22px;
  background: #fff;
  box-shadow: 0 14px 30px rgba(15, 23, 42, 0.05);
  overflow: hidden;
}

.version-table-card__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 20px 24px 8px;
}

.version-table-card__title {
  position: relative;
  padding-left: 12px;
  color: #1f2329;
  font-size: 14px;
  font-weight: 700;
  line-height: 24px;
}

.version-table-card__title::before {
  position: absolute;
  left: 0;
  top: 6px;
  width: 4px;
  height: 12px;
  border-radius: 999px;
  background: var(--pro-ant-color-primary);
  content: "";
}

.version-table {
  padding: 0 8px 8px;
}

.version-table :deep(.ant-table-thead > tr > th) {
  background: #f8fafc;
  color: #344054;
  font-weight: 600;
}

.version-table :deep(.ant-table-tbody > tr > td) {
  border-bottom-color: #eef2f6;
  vertical-align: top;
}

.version-table :deep(.ant-table-tbody > tr:hover > td) {
  background: #fcfdff;
}

.version-name-cell {
  display: flex;
  flex-direction: column;
  gap: 6px;
  min-width: 0;
}

.version-name-cell__title {
  overflow: hidden;
  color: #1f2329;
  font-size: 14px;
  font-weight: 700;
  line-height: 22px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.version-name-cell__remark {
  overflow: hidden;
  color: #667085;
  font-size: 12px;
  line-height: 20px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.metric-cell {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.metric-cell--center {
  align-items: center;
}

.metric-cell__value {
  color: #1f2329;
  font-size: 13px;
  font-weight: 700;
  line-height: 20px;
}

.metric-cell__value--price {
  color: #1677ff;
}

.metric-cell__label {
  color: #98a2b3;
  font-size: 12px;
  line-height: 18px;
}

.version-page .action-link {
  padding: 0;
  font-weight: 600;
}

@media (max-width: 1200px) {
  .version-highlights {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 860px) {
  .version-header {
    flex-direction: column;
    align-items: stretch;
  }

  .version-header__actions {
    justify-content: space-between;
  }

  .version-header__search {
    flex: 1;
    width: auto;
  }

  .version-highlights {
    grid-template-columns: 1fr;
  }
}
</style>
