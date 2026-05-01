<script setup lang="ts">
import type { TableColumnsType } from 'ant-design-vue'
import {
  DownOutlined,
  PlusOutlined,
  ReloadOutlined,
  SearchOutlined,
} from '@ant-design/icons-vue'
import { computed, ref } from 'vue'
import messageService from '@/utils/messageService'
import PlatformModalShell from '../shared/platform-modal-shell.vue'
import { PlatformAccessEnum } from '~@/constants/access'

type DetailTab = 'base' | 'auth'
type LooseScaleRecord = ScaleRecord | Record<string, any>

interface ScaleInstitutionRow {
  name: string
  contact: string
  authState: string
  expireAt: string
}

interface ScaleRecord {
  id: number
  name: string
  code: string
  category: string
  scenario: string
  ageRange: string
  currentVersion: string
  itemCount: number
  domainCount: number
  institutionCount: number
  monthUsage: number
  dataStatus: string
  updatedAt: string
  summary: string
  executionEntry: string
  apiPackage: string
  authInstitutions: ScaleInstitutionRow[]
}

const { hasAccess } = useAccess()

const keyword = ref('')
const appliedKeyword = ref('')
const categoryFilter = ref('')
const scenarioFilter = ref('')
const activeDetailTab = ref<DetailTab>('base')
const detailOpen = ref(false)
const selectedScale = ref<ScaleRecord | null>(null)

const scaleRecords = ref<ScaleRecord[]>([
  {
    id: 1,
    name: 'PEP-3 儿童心理教育评核',
    code: 'PEP3',
    category: '标准化测评',
    scenario: '现场测评',
    ageRange: '2岁6个月 - 6岁',
    currentVersion: '2025-92题版',
    itemCount: 172,
    domainCount: 13,
    institutionCount: 29,
    monthUsage: 418,
    dataStatus: '题库、常模、评分规则和机构端入口已串联',
    updatedAt: '2026-05-01 09:20',
    summary: '面向儿童心理教育与康复评估的标准化量表，已接入机构端测评工作台。',
    executionEntry: '机构端 /teacherCenter/assessment-calendar',
    apiPackage: '/api/v1/assessments/pep3/*',
    authInstitutions: [
      { name: '星河康复中心', contact: '主任 138****1024', authState: '已授权', expireAt: '2026-12-31' },
      { name: '启明特殊教育学校', contact: '教务 176****2311', authState: '已授权', expireAt: '2026-10-15' },
      { name: '晨曦儿童发展中心', contact: '院长 139****9088', authState: '待复核', expireAt: '2026-08-30' },
    ],
  },
])

const columns: TableColumnsType<ScaleRecord> = [
  { title: '量表信息', key: 'scale', width: 300, fixed: 'left' as const },
  { title: '分类 / 场景', key: 'meta', width: 170 },
  { title: '当前版本', key: 'version', width: 170 },
  { title: '题库', key: 'data', width: 130, align: 'center' as const },
  { title: '授权机构', key: 'auth', width: 130, align: 'center' as const },
  { title: '最近更新', key: 'updatedAt', width: 160 },
  { title: '操作', key: 'action', width: 190, fixed: 'right' as const },
]

const filteredScaleRecords = computed(() => {
  const key = appliedKeyword.value.trim().toLowerCase()
  return scaleRecords.value.filter((item) => {
    if (key) {
      const hit = [item.name, item.code, item.currentVersion, item.category, item.scenario, item.ageRange]
        .join(' ')
        .toLowerCase()
        .includes(key)
      if (!hit)
        return false
    }
    if (categoryFilter.value && item.category !== categoryFilter.value)
      return false
    if (scenarioFilter.value && item.scenario !== scenarioFilter.value)
      return false
    return true
  })
})

const categoryOptions = computed(() => ['全部分类', ...new Set(scaleRecords.value.map(item => item.category))].map(item => ({
  label: item,
  value: item === '全部分类' ? '' : item,
})))

const scenarioOptions = computed(() => ['全部场景', ...new Set(scaleRecords.value.map(item => item.scenario))].map(item => ({
  label: item,
  value: item === '全部场景' ? '' : item,
})))

const summaryCards = computed(() => {
  const total = scaleRecords.value.length
  const institutionCount = scaleRecords.value.reduce((sum, item) => sum + item.institutionCount, 0)
  const monthUsage = scaleRecords.value.reduce((sum, item) => sum + item.monthUsage, 0)

  return [
    { label: '全部量表', value: total, hint: '量表包总数' },
    { label: '已授权机构', value: institutionCount, hint: '跨量表授权数' },
    { label: '本月测评', value: monthUsage, hint: '按量表汇总' },
  ]
})

function formatDateOnly(value: string) {
  return value?.slice(0, 10) || '--'
}

function formatTimeOnly(value: string) {
  const time = value?.slice(11, 19) || ''
  if (!time)
    return '--'
  return time.length === 5 ? `${time}:00` : time
}

function asScaleRecord(record: LooseScaleRecord) {
  return record as ScaleRecord
}

function openScaleDetail(record: LooseScaleRecord, tab: DetailTab = 'base') {
  selectedScale.value = asScaleRecord(record)
  activeDetailTab.value = tab
  detailOpen.value = true
}

function handleSearch() {
  appliedKeyword.value = keyword.value.trim()
}

function handlePendingAction(actionName: string) {
  messageService.info(`${actionName}功能暂未开放`)
}

function resetFilters() {
  keyword.value = ''
  appliedKeyword.value = ''
  categoryFilter.value = ''
  scenarioFilter.value = ''
}
</script>

<template>
  <div class="scale-page">
    <div class="scale-page__header">
      <div class="scale-page__heading">
        <div class="scale-page__title">
          量表管理
        </div>
      </div>

      <div class="scale-page__actions">
        <a-button v-if="hasAccess(PlatformAccessEnum.scaleManageAdd)" type="primary">
          <template #icon>
            <PlusOutlined />
          </template>
          新增量表
        </a-button>
      </div>
    </div>

    <div class="scale-summary">
      <div v-for="item in summaryCards" :key="item.label" class="scale-summary__item">
        <div class="scale-summary__label">
          {{ item.label }}
        </div>
        <div class="scale-summary__value">
          {{ item.value }}
        </div>
        <div class="scale-summary__hint">
          {{ item.hint }}
        </div>
      </div>
    </div>

    <div class="scale-panel">
      <div class="scale-toolbar">
        <div class="scale-toolbar__filters">
          <div class="scale-filter-item scale-filter-item--keyword">
            <span class="scale-filter-item__label">关键词搜索</span>
            <a-input
              v-model:value="keyword"
              allow-clear
              placeholder="搜索量表名称、编码、版本、场景"
              class="scale-toolbar__keyword"
              @press-enter="handleSearch"
            />
          </div>

          <div class="scale-filter-item">
            <span class="scale-filter-item__label">量表分类</span>
            <a-select
              v-model:value="categoryFilter"
              :options="categoryOptions"
              placeholder="分类"
              allow-clear
              class="scale-toolbar__select"
            />
          </div>

          <div class="scale-filter-item">
            <span class="scale-filter-item__label">使用场景</span>
            <a-select
              v-model:value="scenarioFilter"
              :options="scenarioOptions"
              placeholder="使用场景"
              allow-clear
              class="scale-toolbar__select"
            />
          </div>

          <a-button type="primary" class="scale-toolbar__search" @click="handleSearch">
            <template #icon>
              <SearchOutlined />
            </template>
            搜索
          </a-button>

        </div>

        <a-button class="scale-toolbar__reset" @click="resetFilters">
          <template #icon>
            <ReloadOutlined />
          </template>
          重置
        </a-button>
      </div>

      <a-table
        class="scale-table"
        :columns="columns"
        :data-source="filteredScaleRecords"
        :pagination="false"
        :scroll="{ x: 1250 }"
        row-key="id"
        size="small"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'scale'">
            <div class="scale-cell">
              <div class="scale-cell__head">
                <a-tooltip :overlay-style="{ maxWidth: '360px', whiteSpace: 'normal' }">
                  <template #title>
                    {{ record.name }}
                  </template>
                  <div class="scale-cell__name">
                    {{ record.name }}
                  </div>
                </a-tooltip>
              </div>

              <div class="scale-cell__meta">
                <span>{{ record.code }}</span>
                <span>{{ record.ageRange }}</span>
              </div>
            </div>
          </template>

          <template v-else-if="column.key === 'meta'">
            <div class="meta-cell">
              <div class="meta-cell__main">
                {{ record.category }}
              </div>
              <div class="meta-cell__sub">
                {{ record.scenario }}
              </div>
            </div>
          </template>

          <template v-else-if="column.key === 'version'">
            <div class="meta-cell">
              <div class="meta-cell__main">
                {{ record.currentVersion }}
              </div>
            </div>
          </template>

          <template v-else-if="column.key === 'data'">
            <div class="metric-cell metric-cell--center">
              <div class="metric-cell__value">
                {{ record.itemCount }}题
              </div>
              <div class="metric-cell__label">
                {{ record.domainCount }}个维度
              </div>
            </div>
          </template>

          <template v-else-if="column.key === 'auth'">
            <div class="metric-cell metric-cell--center">
              <div class="metric-cell__value">
                {{ record.institutionCount }}
              </div>
              <div class="metric-cell__label">
                家机构
              </div>
            </div>
          </template>

          <template v-else-if="column.key === 'updatedAt'">
            <div class="meta-cell">
              <div class="meta-cell__main">
                {{ formatDateOnly(record.updatedAt) }}
              </div>
              <div class="meta-cell__sub">
                {{ formatTimeOnly(record.updatedAt) }}
              </div>
            </div>
          </template>

          <template v-else-if="column.key === 'action'">
            <div class="scale-actions scale-actions--text">
              <a class="scale-actions__link" @click="openScaleDetail(record, 'base')">
                详情
              </a>

              <a v-if="hasAccess(PlatformAccessEnum.scaleManageAuth)" class="scale-actions__link" @click="openScaleDetail(record, 'auth')">
                授权机构
              </a>

              <a-dropdown
                v-if="hasAccess([
                  PlatformAccessEnum.scaleManageIepTarget,
                  PlatformAccessEnum.scaleManageReference,
                  PlatformAccessEnum.scaleManageThanks,
                ])"
                placement="bottomRight"
                :trigger="['click']"
              >
                <a class="scale-actions__link scale-actions__more">
                  更多
                  <DownOutlined class="scale-actions__arrow" />
                </a>
                <template #overlay>
                  <a-menu class="scale-actions__menu">
                    <a-menu-item v-if="hasAccess(PlatformAccessEnum.scaleManageIepTarget)" key="iep" @click="handlePendingAction('IEP目标库')">
                      IEP目标库
                    </a-menu-item>
                    <a-menu-item v-if="hasAccess(PlatformAccessEnum.scaleManageReference)" key="references" @click="handlePendingAction('引用文献')">
                      引用文献
                    </a-menu-item>
                    <a-menu-item v-if="hasAccess(PlatformAccessEnum.scaleManageThanks)" key="acknowledgements" @click="handlePendingAction('特别鸣谢')">
                      特别鸣谢
                    </a-menu-item>
                  </a-menu>
                </template>
              </a-dropdown>
            </div>
          </template>
        </template>
      </a-table>
    </div>

    <PlatformModalShell
      v-model:open="detailOpen"
      :title="selectedScale ? `${selectedScale.name} · 量表详情` : '量表详情'"
      :width="1040"
      :scrollable="true"
      modal-class="scale-detail-modal"
    >
      <template v-if="selectedScale">
        <div class="detail-top">
          <div>
            <div class="detail-top__title">
              {{ selectedScale.name }}
            </div>
            <div class="detail-top__sub">
              {{ selectedScale.code }} · {{ selectedScale.category }} · {{ selectedScale.scenario }} · {{ selectedScale.ageRange }}
            </div>
          </div>

          <a-tag color="blue">
            {{ selectedScale.currentVersion }}
          </a-tag>
        </div>

        <a-tabs v-model:activeKey="activeDetailTab" class="detail-tabs">
          <a-tab-pane key="base" tab="基础信息">
            <div class="detail-section">
              <div class="detail-section__head">
                <div class="detail-section__title">
                  接入概览
                </div>
              </div>

              <a-descriptions bordered size="small" :column="2">
                <a-descriptions-item label="量表说明">
                  {{ selectedScale.summary }}
                </a-descriptions-item>
                <a-descriptions-item label="当前版本">
                  {{ selectedScale.currentVersion }}
                </a-descriptions-item>
                <a-descriptions-item label="题库 / 维度">
                  {{ selectedScale.itemCount }} 题 / {{ selectedScale.domainCount }} 个维度
                </a-descriptions-item>
                <a-descriptions-item label="适用年龄">
                  {{ selectedScale.ageRange }}
                </a-descriptions-item>
                <a-descriptions-item label="执行入口">
                  {{ selectedScale.executionEntry }}
                </a-descriptions-item>
                <a-descriptions-item label="接口包">
                  {{ selectedScale.apiPackage }}
                </a-descriptions-item>
                <a-descriptions-item label="数据状态">
                  {{ selectedScale.dataStatus }}
                </a-descriptions-item>
              </a-descriptions>
            </div>
          </a-tab-pane>

          <a-tab-pane v-if="hasAccess(PlatformAccessEnum.scaleManageAuth)" key="auth" tab="机构授权">
            <div class="detail-section">
              <div class="detail-section__head">
                <div class="detail-section__title">
                  授权机构
                </div>
                <a-tag color="green">
                  {{ selectedScale.institutionCount }} 家
                </a-tag>
              </div>

              <a-table
                :columns="[
                  { title: '机构名称', dataIndex: 'name', key: 'name' },
                  { title: '联系人', dataIndex: 'contact', key: 'contact', width: 180 },
                  { title: '授权状态', dataIndex: 'authState', key: 'authState', width: 120 },
                  { title: '到期时间', dataIndex: 'expireAt', key: 'expireAt', width: 130 },
                ]"
                :data-source="selectedScale.authInstitutions"
                :pagination="false"
                row-key="name"
                size="small"
              />
            </div>
          </a-tab-pane>

        </a-tabs>
      </template>
    </PlatformModalShell>
  </div>
</template>

<style scoped lang="less">
.scale-page {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.scale-page__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 2px 2px 0;
}

.scale-page__heading {
  display: flex;
  align-items: baseline;
  gap: 10px;
  min-width: 0;
}

.scale-page__title {
  color: #1f2329;
  font-size: 16px;
  font-weight: 700;
  line-height: 32px;
}

.scale-page__count {
  color: #98a2b3;
  font-size: 12px;
  line-height: 18px;
}

.scale-page__actions {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-shrink: 0;
}

.scale-summary {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  overflow: hidden;
  border: 1px solid #e9edf3;
  border-radius: 10px;
  background: #fff;
}

.scale-summary__item {
  display: flex;
  align-items: baseline;
  gap: 10px;
  min-height: 42px;
  padding: 8px 18px;
  border-right: 1px solid #eef2f6;
}

.scale-summary__item:last-child {
  border-right: 0;
}

.scale-summary__label {
  color: #667085;
  font-size: 13px;
  line-height: 22px;
  white-space: nowrap;
}

.scale-summary__value {
  color: #1f2329;
  font-size: 20px;
  font-weight: 700;
  line-height: 24px;
  white-space: nowrap;
}

.scale-summary__hint {
  color: #98a2b3;
  font-size: 12px;
  line-height: 20px;
  white-space: nowrap;
}

.scale-panel {
  overflow: hidden;
  border: 1px solid #e9edf3;
  border-radius: 10px;
  background: #fff;
}

.scale-toolbar {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 10px;
  padding: 16px;
}

.scale-toolbar__filters {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 10px 16px;
  min-width: 0;
}

.scale-filter-item {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.scale-filter-item--keyword {
  gap: 8px;
}

.scale-filter-item__label {
  flex-shrink: 0;
  color: #262626;
  font-size: 14px;
  line-height: 32px;
  white-space: nowrap;
}

.scale-toolbar__keyword {
  width: 300px;
}

.scale-toolbar__select {
  width: 150px;
}

.scale-toolbar__search {
  flex-shrink: 0;
}

.scale-toolbar__reset {
  flex-shrink: 0;
}

.scale-table {
  padding: 0 8px 8px;
}

.scale-table :deep(.ant-table-thead > tr > th) {
  padding: 12px 16px;
  background: #fafafa !important;
  color: #262626;
  font-size: 14px;
  font-weight: 500;
  line-height: 20px;
  border-bottom: 1px solid #f0f0f0;
  white-space: nowrap;
}

.scale-table :deep(.ant-table-thead > tr > th .ant-table-column-title) {
  color: #262626;
  font-weight: 500;
}

.scale-table :deep(.ant-table-tbody > tr > td) {
  padding: 16px;
  vertical-align: middle;
  border-bottom: 1px solid #f5f5f5;
}

.scale-table :deep(.ant-table-tbody > tr:hover > td) {
  background: #fcfcfc;
}

.scale-table :deep(.ant-table-cell-fix-left),
.scale-table :deep(.ant-table-cell-fix-right) {
  background: #fff;
}

.scale-table :deep(.ant-table-tbody > tr:hover .ant-table-cell-fix-left),
.scale-table :deep(.ant-table-tbody > tr:hover .ant-table-cell-fix-right) {
  background: #fcfcfc;
}

.scale-cell {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.scale-cell__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}

.scale-cell__name {
  min-width: 0;
  color: #262626;
  font-size: 14px;
  font-weight: 600;
  line-height: 22px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.scale-cell__meta {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
  color: #8c8c8c;
  font-size: 12px;
  line-height: 20px;
}

.meta-cell {
  display: flex;
  flex-direction: column;
}

.meta-cell__main {
  color: #262626;
  font-size: 13px;
  font-weight: 500;
  line-height: 22px;
}

.meta-cell__sub {
  max-width: 100%;
  overflow: hidden;
  color: #8c8c8c;
  font-size: 12px;
  line-height: 20px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.metric-cell {
  display: flex;
  flex-direction: column;
}

.metric-cell--center {
  align-items: center;
}

.metric-cell__value {
  color: #262626;
  font-size: 13px;
  font-weight: 500;
  line-height: 22px;
}

.metric-cell__label {
  color: #8c8c8c;
  font-size: 12px;
  line-height: 20px;
}

.scale-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  padding-right: 4px;
  flex-wrap: wrap;
  white-space: nowrap;
}

.scale-actions--text {
  gap: 12px;
}

.scale-actions__link {
  color: #1677ff;
  font-size: 14px;
  line-height: 22px;
  cursor: pointer;
  white-space: nowrap;
  transition: color 0.2s ease;
}

.scale-actions__link:hover {
  color: #4096ff;
}

.scale-actions__more {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.scale-actions__arrow {
  font-size: 10px;
}

:deep(.scale-actions__menu.ant-dropdown-menu) {
  min-width: 124px;
  padding: 8px 0;
  border-radius: 12px;
  box-shadow: 0 10px 28px rgba(15, 35, 95, 0.12);
}

:deep(.scale-actions__menu .ant-dropdown-menu-item) {
  min-height: 40px;
  padding: 8px 16px;
  border-radius: 0;
  color: #262626;
  font-size: 14px;
}

.detail-top {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 16px;
}

.detail-top__title {
  color: #1f2329;
  font-size: 18px;
  font-weight: 700;
  line-height: 28px;
}

.detail-top__sub {
  margin-top: 4px;
  color: #667085;
  font-size: 12px;
  line-height: 20px;
}

.detail-tabs :deep(.ant-tabs-nav) {
  margin-bottom: 12px;
}

.detail-section {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.detail-section__group {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.detail-section__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.detail-section__title {
  color: #1f2329;
  font-size: 14px;
  font-weight: 700;
  line-height: 22px;
}

@media (max-width: 1200px) {
  .scale-summary {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }

  .scale-summary__item {
    border-bottom: 1px solid #eef2f6;
  }

  .scale-toolbar__keyword {
    flex: 1 1 280px;
    width: auto;
  }
}

@media (max-width: 900px) {
  .scale-page__header {
    flex-direction: column;
    align-items: stretch;
  }

  .scale-page__actions {
    width: 100%;
    justify-content: flex-start;
  }

  .scale-summary {
    grid-template-columns: 1fr;
  }

  .scale-summary__item {
    border-right: 0;
  }

  .scale-toolbar__select {
    flex: 1 1 140px;
  }
}
</style>
