<script setup lang="ts">
import type { TableColumnsType } from 'ant-design-vue'
import {
  BookOutlined,
  DownOutlined,
  PlusOutlined,
  ReloadOutlined,
  TeamOutlined,
} from '@ant-design/icons-vue'
import { computed, ref } from 'vue'
import messageService from '@/utils/messageService'
import PlatformModalShell from '../shared/platform-modal-shell.vue'
import { PlatformAccessEnum } from '~@/constants/access'

type ScaleStatus = 'available' | 'disabled'
type DetailTab = 'base' | 'versions' | 'auth' | 'iep' | 'citation'
type LooseScaleRecord = ScaleRecord | Record<string, any>

interface ScaleVersionRow {
  version: string
  status: string
  itemCount: number
  updatedAt: string
}

interface ScaleInstitutionRow {
  name: string
  contact: string
  authState: string
  expireAt: string
}

interface ScaleIepRow {
  name: string
  count: number
  owner: string
  updatedAt: string
}

interface ScaleRecord {
  id: number
  name: string
  code: string
  category: string
  scenario: string
  ageRange: string
  status: ScaleStatus
  engineType: string
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
  references: string[]
  acknowledgements: string[]
  versionRows: ScaleVersionRow[]
  authInstitutions: ScaleInstitutionRow[]
  iepLibraries: ScaleIepRow[]
}

const { hasAccess } = useAccess()

const keyword = ref('')
const categoryFilter = ref('')
const scenarioFilter = ref('')
const statusFilter = ref<'all' | ScaleStatus>('all')
const engineFilter = ref('')
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
    status: 'available',
    engineType: '内置评分引擎',
    currentVersion: '2025-92mo-draft',
    itemCount: 172,
    domainCount: 13,
    institutionCount: 29,
    monthUsage: 418,
    dataStatus: '题库、常模、评分规则和机构端入口已串联',
    updatedAt: '2026-05-01 09:20',
    summary: '面向儿童心理教育与康复评估的标准化量表，已接入机构端测评工作台。',
    executionEntry: '机构端 /teacherCenter/assessment-calendar',
    apiPackage: '/api/v1/assessments/pep3/*',
    references: [
      'Schopler, E. et al. PEP-3 Clinical Guide',
      '儿童心理教育评核相关本土化译注',
      '康复评估与教育干预常模整理稿',
    ],
    acknowledgements: ['张老师', '李博士', '王主任', '周老师'],
    versionRows: [
      { version: '2025-92mo-draft', status: '已发布', itemCount: 172, updatedAt: '2026-05-01 09:20' },
      { version: '2025-92mo-rev01', status: '草稿', itemCount: 172, updatedAt: '2026-04-28 16:12' },
      { version: '2024-90mo-initial', status: '停用', itemCount: 168, updatedAt: '2026-04-10 10:30' },
    ],
    authInstitutions: [
      { name: '星河康复中心', contact: '主任 138****1024', authState: '已授权', expireAt: '2026-12-31' },
      { name: '启明特殊教育学校', contact: '教务 176****2311', authState: '已授权', expireAt: '2026-10-15' },
      { name: '晨曦儿童发展中心', contact: '院长 139****9088', authState: '待复核', expireAt: '2026-08-30' },
    ],
    iepLibraries: [
      { name: 'PEP3-IEP 基础版', count: 42, owner: '教研中心', updatedAt: '2026-04-30 14:20' },
      { name: 'PEP3-IEP 语言训练', count: 18, owner: '语训组', updatedAt: '2026-04-29 17:45' },
      { name: 'PEP3-IEP 认知干预', count: 25, owner: '康复组', updatedAt: '2026-04-28 11:15' },
    ],
  },
])

const columns: TableColumnsType<ScaleRecord> = [
  { title: '量表信息', key: 'scale', width: 220, fixed: 'left' as const },
  { title: '分类 / 场景', key: 'meta', width: 160 },
  { title: '当前版本', key: 'version', width: 140 },
  { title: '题库', key: 'data', width: 120, align: 'center' as const },
  { title: '授权机构', key: 'auth', width: 110, align: 'center' as const },
  { title: '最近更新', key: 'updatedAt', width: 140 },
  { title: '操作', key: 'action', width: 180, fixed: 'right' as const },
]

const filteredScaleRecords = computed(() => {
  const key = keyword.value.trim().toLowerCase()
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
    if (statusFilter.value !== 'all' && item.status !== statusFilter.value)
      return false
    if (engineFilter.value && item.engineType !== engineFilter.value)
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

const engineOptions = computed(() => ['全部引擎', ...new Set(scaleRecords.value.map(item => item.engineType))].map(item => ({
  label: item,
  value: item === '全部引擎' ? '' : item,
})))

const summaryCards = computed(() => {
  const total = scaleRecords.value.length
  const availableCount = scaleRecords.value.filter(item => item.status === 'available').length
  const disabledCount = scaleRecords.value.filter(item => item.status === 'disabled').length
  const institutionCount = scaleRecords.value.reduce((sum, item) => sum + item.institutionCount, 0)
  const monthUsage = scaleRecords.value.reduce((sum, item) => sum + item.monthUsage, 0)

  return [
    { label: '全部量表', value: total, hint: '量表包总数' },
    { label: '可用量表', value: availableCount, hint: '可供机构授权' },
    { label: '停用量表', value: disabledCount, hint: '当前不可执行' },
    { label: '已授权机构', value: institutionCount, hint: '跨量表授权数' },
    { label: '本月测评', value: monthUsage, hint: '按量表汇总' },
  ]
})

function formatStatus(status: ScaleStatus) {
  return status === 'available' ? '可用' : '停用'
}

function statusColor(status: ScaleStatus) {
  return status === 'available' ? 'green' : 'default'
}

function formatDateOnly(value: string) {
  return value?.slice(0, 10) || '--'
}

function formatTimeOnly(value: string) {
  return value?.slice(11, 19) || '--'
}

function asScaleRecord(record: LooseScaleRecord) {
  return record as ScaleRecord
}

function openScaleDetail(record: LooseScaleRecord, tab: DetailTab = 'base') {
  selectedScale.value = asScaleRecord(record)
  activeDetailTab.value = tab
  detailOpen.value = true
}

function handleToggleStatus(record: LooseScaleRecord) {
  const scale = asScaleRecord(record)
  scale.status = scale.status === 'available' ? 'disabled' : 'available'
  messageService.success(`已切换为${formatStatus(scale.status)}状态`)
}

function handlePublishVersion(record: LooseScaleRecord) {
  const scale = asScaleRecord(record)
  messageService.success(`已触发 ${scale.name} 的版本发布流程`)
}

function resetFilters() {
  keyword.value = ''
  categoryFilter.value = ''
  scenarioFilter.value = ''
  statusFilter.value = 'all'
  engineFilter.value = ''
}
</script>

<template>
  <div class="scale-page">
    <div class="scale-page__header">
      <div class="scale-page__heading">
        <div class="scale-page__title">
          量表管理
        </div>
        <div class="scale-page__count">
          共 {{ scaleRecords.length }} 个量表包
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
          <a-input
            v-model:value="keyword"
            allow-clear
            placeholder="搜索量表名称、编码、版本、场景"
            class="scale-toolbar__keyword"
          />

          <a-select
            v-model:value="categoryFilter"
            :options="categoryOptions"
            placeholder="分类"
            allow-clear
            class="scale-toolbar__select"
          />

          <a-select
            v-model:value="scenarioFilter"
            :options="scenarioOptions"
            placeholder="使用场景"
            allow-clear
            class="scale-toolbar__select"
          />

          <a-select
            v-model:value="engineFilter"
            :options="engineOptions"
            placeholder="评分引擎"
            allow-clear
            class="scale-toolbar__select"
          />

          <a-segmented
            v-model:value="statusFilter"
            class="scale-toolbar__status"
            :options="[
              { label: '全部', value: 'all' },
              { label: '可用', value: 'available' },
              { label: '停用', value: 'disabled' },
            ]"
          />
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
        :scroll="{ x: 1070 }"
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
                <span class="scale-state" :class="`scale-state--${record.status}`">
                  <span class="scale-state__dot" />
                  {{ formatStatus(record.status) }}
                </span>
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
              <div class="meta-cell__sub">
                {{ record.engineType }}
              </div>
            </div>
          </template>

          <template v-else-if="column.key === 'data'">
            <div class="metric-cell metric-cell--center">
              <div class="metric-cell__value">
                {{ record.itemCount }}题
              </div>
              <div class="metric-cell__label">
                {{ record.domainCount }} 个维度
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

              <a-dropdown placement="bottomRight" :trigger="['click']">
                <a class="scale-actions__link scale-actions__more">
                  更多
                  <DownOutlined class="scale-actions__arrow" />
                </a>
                <template #overlay>
                  <a-menu class="scale-actions__menu" @click="({ key }) => {
                    if (key === 'versions')
                      openScaleDetail(record, 'versions')
                    if (key === 'iep')
                      openScaleDetail(record, 'iep')
                    if (key === 'publish')
                      handlePublishVersion(record)
                    if (key === 'toggle')
                      handleToggleStatus(record)
                  }">
                    <a-menu-item key="versions">
                      版本数据
                    </a-menu-item>
                    <a-menu-item key="iep">
                      IEP库
                    </a-menu-item>
                    <a-menu-item v-if="hasAccess(PlatformAccessEnum.scaleManagePublish)" key="publish">
                      发布版本
                    </a-menu-item>
                    <a-menu-item v-if="hasAccess(PlatformAccessEnum.scaleManageEdit)" key="toggle">
                      {{ record.status === 'available' ? '停用量表' : '启用量表' }}
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

          <div class="detail-top__tags">
            <a-tag :color="statusColor(selectedScale.status)">
              {{ formatStatus(selectedScale.status) }}
            </a-tag>
            <a-tag color="blue">
              {{ selectedScale.currentVersion }}
            </a-tag>
          </div>
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
                <a-descriptions-item label="评分引擎">
                  {{ selectedScale.engineType }}
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

          <a-tab-pane key="versions" tab="版本数据">
            <div class="detail-section">
              <div class="detail-section__head">
                <div class="detail-section__title">
                  版本列表
                </div>
                <a-tag color="blue">
                  {{ selectedScale.versionRows.length }} 个版本
                </a-tag>
              </div>

              <a-table
                :columns="[
                  { title: '版本', dataIndex: 'version', key: 'version' },
                  { title: '状态', dataIndex: 'status', key: 'status', width: 110 },
                  { title: '题数', dataIndex: 'itemCount', key: 'itemCount', width: 100, align: 'center' },
                  { title: '更新时间', dataIndex: 'updatedAt', key: 'updatedAt', width: 160 },
                ]"
                :data-source="selectedScale.versionRows"
                :pagination="false"
                row-key="version"
                size="small"
              />
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

          <a-tab-pane key="iep" tab="IEP库">
            <div class="detail-section">
              <div class="detail-section__head">
                <div class="detail-section__title">
                  IEP 库入口
                </div>
                <a-tag color="blue">
                  {{ selectedScale.iepLibraries.length }} 个库
                </a-tag>
              </div>

              <div class="iep-grid">
                <div v-for="item in selectedScale.iepLibraries" :key="item.name" class="iep-card">
                  <div class="iep-card__title">
                    {{ item.name }}
                  </div>
                  <div class="iep-card__meta">
                    <span>{{ item.count }} 条目标</span>
                    <span>{{ item.owner }}</span>
                  </div>
                  <div class="iep-card__time">
                    {{ item.updatedAt }}
                  </div>
                </div>
              </div>
            </div>
          </a-tab-pane>

          <a-tab-pane key="citation" tab="引用与鸣谢">
            <div class="detail-section">
              <div class="detail-section__group">
                <div class="detail-section__head">
                  <div class="detail-section__title">
                    引用文献
                  </div>
                </div>
                <div class="detail-list">
                  <div v-for="item in selectedScale.references" :key="item" class="detail-list__item">
                    <BookOutlined />
                    <span>{{ item }}</span>
                  </div>
                </div>
              </div>

              <a-divider />

              <div class="detail-section__group">
                <div class="detail-section__head">
                  <div class="detail-section__title">
                    特别鸣谢
                  </div>
                </div>
                <div class="detail-list detail-list--people">
                  <div v-for="person in selectedScale.acknowledgements" :key="person" class="detail-list__item">
                    <TeamOutlined />
                    <span>{{ person }}</span>
                  </div>
                </div>
              </div>
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
  grid-template-columns: repeat(5, minmax(0, 1fr));
  overflow: hidden;
  border: 1px solid #e9edf3;
  border-radius: 10px;
  background: #fff;
}

.scale-summary__item {
  min-height: 58px;
  padding: 9px 16px;
  border-right: 1px solid #eef2f6;
}

.scale-summary__item:last-child {
  border-right: 0;
}

.scale-summary__label {
  color: #667085;
  font-size: 12px;
  line-height: 18px;
}

.scale-summary__value {
  margin-top: 2px;
  color: #1f2329;
  font-size: 20px;
  font-weight: 700;
  line-height: 24px;
}

.scale-summary__hint {
  color: #98a2b3;
  font-size: 12px;
  line-height: 18px;
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
  gap: 10px;
  min-width: 0;
}

.scale-toolbar__keyword {
  width: 340px;
}

.scale-toolbar__select {
  width: 142px;
}

.scale-toolbar__status {
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

.scale-state {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  flex-shrink: 0;
  color: #667085;
  font-size: 12px;
  line-height: 18px;
}

.scale-state__dot {
  width: 6px;
  height: 6px;
  border-radius: 999px;
  background: #98a2b3;
}

.scale-state--available {
  color: #389e0d;
}

.scale-state--available .scale-state__dot {
  background: #52c41a;
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

.detail-top__tags {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
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

.iep-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 12px;
}

.iep-card {
  padding: 14px 16px;
  border: 1px solid #e9edf3;
  border-radius: 12px;
  background: #fff;
}

.iep-card__title {
  color: #1f2329;
  font-size: 14px;
  font-weight: 600;
  line-height: 22px;
}

.iep-card__meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  margin-top: 8px;
  color: #667085;
  font-size: 12px;
  line-height: 18px;
}

.iep-card__time {
  margin-top: 10px;
  color: #98a2b3;
  font-size: 12px;
  line-height: 18px;
}

.detail-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.detail-list--people {
  gap: 8px;
}

.detail-list__item {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #344054;
  font-size: 13px;
  line-height: 20px;
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
