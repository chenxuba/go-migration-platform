<script setup lang="ts">
import {
  AppstoreOutlined,
  BarsOutlined,
  BookOutlined,
  CheckCircleFilled,
  ClockCircleOutlined,
  ExperimentOutlined,
  FileDoneOutlined,
  FileTextOutlined,
  SearchOutlined,
  TeamOutlined,
} from '@ant-design/icons-vue'
import { computed, onMounted, ref } from 'vue'
import {
  getScaleLibraryApi,
  type ScaleLibraryItem,
  type ScaleLibraryQuery,
  type ScaleLibraryResponse,
  type ScaleLibraryStatus,
} from '@/api/teacher-center/scale-library'
import scaleIntroImage from '@/assets/images/image.png'
import messageService from '@/utils/messageService'

interface ChildOption {
  id: number
  shortName: string
  name: string
  gender: string
  age: string
  contactPhone: string
  className: string
  latestAssessment: string
}

const searchText = ref('')
const childSearchText = ref('')
const ageScope = ref('all')
const categoryFilter = ref('all')
const scenarioFilter = ref('')
const statusFilter = ref('')
const durationScope = ref('')
const libraryLoading = ref(false)
const selectedChildId = ref<number>(10015)
const startModalOpen = ref(false)
const detailModalOpen = ref(false)
const activeScale = ref<ScaleLibraryItem>()
const library = ref<ScaleLibraryResponse>({
  items: [],
  summary: { total: 0, available: 0, unavailable: 0, monthUsage: 0, usageCount: 0, reservedAuths: 0 },
  filterOptions: { categories: [], categoryCounts: {}, scenarios: [], statuses: [], ageScopes: ['all', '0-2', '2-6', '6-12', '12+'], durations: ['0-15', '15-30', '30-60', '60+'] },
})
let libraryRequestSeq = 0

const childOptions: ChildOption[] = [
  { id: 10012, shortName: '乐', name: '乐乐', gender: '男', age: '5岁2个月', contactPhone: '爸爸 176****0124', className: '小海豚班', latestAssessment: '2025-04-28' },
  { id: 10015, shortName: '小', name: '小宇', gender: '男', age: '4岁8个月', contactPhone: '妈妈 188****5630', className: '小海豚班', latestAssessment: '2025-05-12' },
  { id: 10009, shortName: '安', name: '安安', gender: '女', age: '6岁1个月', contactPhone: '妈妈 139****8902', className: '小海豚班', latestAssessment: '2025-04-15' },
  { id: 10007, shortName: '浩', name: '浩浩', gender: '男', age: '7岁3个月', contactPhone: '爸爸 185****7721', className: '海星班', latestAssessment: '2025-05-10' },
  { id: 10021, shortName: '糖', name: '糖糖', gender: '女', age: '3岁11个月', contactPhone: '妈妈 177****3056', className: '海星班', latestAssessment: '2025-04-20' },
]

const categoryTabs = computed(() => {
  const categories = library.value.filterOptions.categories || []
  const counts = library.value.filterOptions.categoryCounts || {}
  const totalCount = Object.values(counts).reduce((sum, count) => sum + Number(count || 0), 0)
  return [
    { key: 'all', label: '全部', count: totalCount || library.value.summary.total, color: 'blue', icon: AppstoreOutlined },
    ...categories.map((category, index) => ({
      key: category,
      label: category,
      count: counts[category] || 0,
      color: index % 2 === 0 ? 'blue' : 'orange',
      icon: index % 2 === 0 ? CheckCircleFilled : ExperimentOutlined,
    })),
  ]
})

const summaryCards = computed(() => [
  { label: '全部量表', value: library.value.summary.total, suffix: '个', desc: '来自系统量表库', icon: AppstoreOutlined, tone: 'blue' },
  { label: '可用量表', value: library.value.summary.available, suffix: '个', desc: '已配置测评入口', icon: CheckCircleFilled, tone: 'green' },
  { label: '预留授权', value: library.value.summary.reservedAuths, suffix: '个', desc: '后续接入分配/授权', icon: FileTextOutlined, tone: 'orange' },
  { label: '本月已测评', value: library.value.summary.monthUsage, suffix: '次', desc: '按当前机构统计', icon: BarsOutlined, tone: 'purple' },
])

const scenarioOptions = computed(() => library.value.filterOptions.scenarios || [])
const selectedChild = computed(() => childOptions.find(item => item.id === selectedChildId.value))

const filteredChildOptions = computed(() => {
  const keyword = childSearchText.value.trim()
  if (!keyword)
    return childOptions
  return childOptions.filter(item => `${item.name}${item.contactPhone}${item.className}${item.latestAssessment}${item.gender}${item.age}`.includes(keyword))
})

function unwrap<T>(res: any): T {
  return (res?.data ?? res?.result ?? res) as T
}

function statusMeta(status: ScaleLibraryStatus, statusText?: string) {
  const map: Record<string, { text: string, className: string }> = {
    available: { text: statusText || '可用', className: 'is-available' },
    unavailable: { text: statusText || '暂不可用', className: 'is-disabled' },
  }
  return map[status] || { text: statusText || '未知', className: 'is-disabled' }
}

function tagClass(index: number) {
  return ['tag-blue', 'tag-green', 'tag-purple'][index % 3]
}

async function fetchScaleLibrary() {
  const requestSeq = ++libraryRequestSeq
  libraryLoading.value = true
  try {
    const res = await getScaleLibraryApi(buildScaleLibraryQuery())
    if (requestSeq === libraryRequestSeq)
      library.value = unwrap<ScaleLibraryResponse>(res)
  } catch (error: any) {
    messageService.error(error?.response?.data?.message || error?.message || '获取量表库失败')
  } finally {
    if (requestSeq === libraryRequestSeq)
      libraryLoading.value = false
  }
}

function buildScaleLibraryQuery(): ScaleLibraryQuery {
  return {
    keyword: searchText.value.trim() || undefined,
    category: categoryFilter.value === 'all' ? undefined : categoryFilter.value,
    scenario: scenarioFilter.value || undefined,
    status: statusFilter.value || undefined,
    ageScope: ageScope.value === 'all' ? undefined : ageScope.value,
    duration: durationScope.value || undefined,
  }
}

function resetFilters() {
  searchText.value = ''
  ageScope.value = 'all'
  categoryFilter.value = 'all'
  scenarioFilter.value = ''
  statusFilter.value = ''
  durationScope.value = ''
  void fetchScaleLibrary()
}

function setCategoryFilter(value: string) {
  categoryFilter.value = value
  void fetchScaleLibrary()
}

function setStatusFilter(value: string) {
  statusFilter.value = value
  void fetchScaleLibrary()
}

function setAgeScope(value: string) {
  ageScope.value = value
  void fetchScaleLibrary()
}

function setScenarioFilter(value: string) {
  scenarioFilter.value = value
  void fetchScaleLibrary()
}

function setDurationScope(value: string) {
  durationScope.value = value
  void fetchScaleLibrary()
}

function handleAgeScopeChange(event: any) {
  setAgeScope(event?.target?.value || 'all')
}

function openStartModal(scale: ScaleLibraryItem) {
  if (scale.status !== 'available') {
    messageService.warning('该量表暂不可发起测评')
    return
  }
  activeScale.value = scale
  startModalOpen.value = true
}

function openDetailModal(scale: ScaleLibraryItem) {
  activeScale.value = scale
  detailModalOpen.value = true
}

function openIepLibrary(scale: ScaleLibraryItem) {
  messageService.info(`${scale.name} 的 IEP 库入口已预留`)
}

function confirmStartAssessment() {
  if (!activeScale.value || !selectedChild.value)
    return
  messageService.success(`已选择 ${selectedChild.value.name}，准备开始 ${activeScale.value.name}`)
  startModalOpen.value = false
}

onMounted(fetchScaleLibrary)
</script>

<template>
  <div class="scale-library-page">
    <section class="category-tabs">
      <button
        v-for="item in categoryTabs"
        :key="item.key"
        type="button"
        class="category-tab"
        :class="[`is-${item.color}`, { 'is-active': item.key === categoryFilter }]"
        @click="setCategoryFilter(item.key)"
      >
        <component :is="item.icon" />
        <span>{{ item.label }}</span>
        <b>{{ item.count }}</b>
      </button>
    </section>

    <section class="summary-grid">
      <div v-for="item in summaryCards" :key="item.label" class="summary-card">
        <div class="summary-icon" :class="`is-${item.tone}`">
          <component :is="item.icon" />
        </div>
        <div class="summary-card__content">
          <div>
            <div class="summary-label">{{ item.label }}</div>
            <div class="summary-value">{{ item.value }}<span>{{ item.suffix }}</span></div>
          </div>
          <div class="summary-desc">{{ item.desc }}</div>
        </div>
      </div>
    </section>

    <section class="library-workbench">
      <aside class="filter-panel">
        <div class="filter-title">
          <strong>快速筛选</strong>
          <a @click="resetFilters">清空</a>
        </div>

        <div class="filter-group">
          <div class="filter-group__title">状态</div>
          <a-checkbox :checked="statusFilter === ''" @change="setStatusFilter('')">全部状态</a-checkbox>
          <a-checkbox :checked="statusFilter === 'available'" @change="setStatusFilter('available')">可用 <span>{{ library.summary.available }}</span></a-checkbox>
          <a-checkbox :checked="statusFilter === 'unavailable'" @change="setStatusFilter('unavailable')">停用 <span>{{ library.summary.unavailable }}</span></a-checkbox>
        </div>

        <div class="filter-group">
          <div class="filter-group__title">适用年龄</div>
          <a-radio-group v-model:value="ageScope" class="custom-radio filter-radio-group age-radio-group" @change="handleAgeScopeChange">
            <a-radio value="all">全部年龄</a-radio>
            <a-radio value="0-2">0-2岁</a-radio>
            <a-radio value="2-6">2-6岁</a-radio>
            <a-radio value="6-12">6-12岁</a-radio>
            <a-radio value="12+">12岁以上</a-radio>
          </a-radio-group>
        </div>

        <div class="filter-group">
          <div class="filter-group__title">使用场景</div>
          <a-checkbox :checked="scenarioFilter === ''" @change="setScenarioFilter('')">全部场景</a-checkbox>
          <a-checkbox
            v-for="scenario in scenarioOptions"
            :key="scenario"
            :checked="scenarioFilter === scenario"
            @change="setScenarioFilter(scenario)"
          >
            {{ scenario }}
          </a-checkbox>
        </div>

        <div class="filter-group">
          <div class="filter-group__title">测评时长</div>
          <a-checkbox :checked="durationScope === '0-15'" @change="setDurationScope(durationScope === '0-15' ? '' : '0-15')">15分钟以内</a-checkbox>
          <a-checkbox :checked="durationScope === '15-30'" @change="setDurationScope(durationScope === '15-30' ? '' : '15-30')">15-30分钟</a-checkbox>
          <a-checkbox :checked="durationScope === '30-60'" @change="setDurationScope(durationScope === '30-60' ? '' : '30-60')">30-60分钟</a-checkbox>
          <a-checkbox :checked="durationScope === '60+'" @change="setDurationScope(durationScope === '60+' ? '' : '60+')">60分钟以上</a-checkbox>
        </div>
      </aside>

      <main class="scale-content">
        <div class="scale-content-scroll">
          <a-spin :spinning="libraryLoading">
            <div v-if="library.items.length" class="scale-card-grid">
              <article v-for="scale in library.items" :key="scale.id" class="scale-card">
              <div class="scale-card__head">
                <div>
                  <h2>{{ scale.name }}</h2>
                  <div class="tag-list">
                    <span v-if="scale.category" :class="tagClass(0)">{{ scale.category }}</span>
                    <span v-if="scale.ageRange" :class="tagClass(1)">{{ scale.ageRange }}</span>
                    <span v-if="scale.scenario" :class="tagClass(2)">{{ scale.scenario }}</span>
                  </div>
                </div>
                <div class="card-side">
                  <div class="card-status">
                    <span class="status-pill" :class="statusMeta(scale.status, scale.statusText).className">
                      {{ statusMeta(scale.status, scale.statusText).text }}
                    </span>
                  </div>
                  <div class="card-side-actions">
                    <button type="button" class="iep-trigger" @click="openIepLibrary(scale)">
                      <BookOutlined />
                      <span>IEP库</span>
                    </button>
                    <a-popover placement="bottomRight" trigger="hover" overlay-class-name="reference-popover">
                      <template #content>
                        <div class="reference-popover-content">
                          <div class="reference-section">
                            <div class="reference-section__title">引用文献</div>
                            <div
                              v-for="(reference, referenceIndex) in scale.references"
                              :key="reference.content"
                              class="reference-item"
                            >
                              <span>{{ referenceIndex + 1 }}</span>
                              <p>{{ reference.content }}</p>
                            </div>
                          </div>
                          <div class="reference-divider"></div>
                          <div class="reference-section">
                            <div class="reference-section__title">特别鸣谢</div>
                            <div
                              v-for="(acknowledgement, acknowledgementIndex) in scale.acknowledgements"
                              :key="acknowledgement.content"
                              class="acknowledgement-item"
                            >
                              <span>{{ acknowledgementIndex + 1 }}</span>
                              <p>{{ acknowledgement.content }}</p>
                            </div>
                          </div>
                        </div>
                      </template>
                      <button type="button" class="reference-trigger">
                        <FileTextOutlined />
                        <span>引用文献</span>
                      </button>
                    </a-popover>
                  </div>
                </div>
              </div>

              <p class="scale-desc">{{ scale.summary || scale.dataStatus || '暂无量表说明' }}</p>

              <div class="scale-meta">
                <div><FileTextOutlined /><span>题目数量</span><b>{{ scale.itemCount }}题</b></div>
                <div><TeamOutlined /><span>评估维度</span><b>{{ scale.domainCount }}个</b></div>
                <div><ClockCircleOutlined /><span>测评时长</span><b>{{ scale.duration || '--' }}</b></div>
                <div><FileDoneOutlined /><span>最近使用</span><b>{{ scale.latestUse || '--' }}</b></div>
              </div>

              <div class="scale-card__footer">
                <span>使用次数 {{ scale.usageCount }}次，本月 {{ scale.monthUsage }}次</span>
                <div class="card-actions">
                  <a-button @click="openDetailModal(scale)">量表介绍</a-button>
                  <a-button
                    type="primary"
                    :disabled="scale.status !== 'available'"
                    @click="openStartModal(scale)"
                  >
                    开始测评
                  </a-button>
                </div>
              </div>
              </article>
            </div>
            <a-empty v-else description="暂无量表" class="scale-empty" />
          </a-spin>
        </div>
      </main>
    </section>

    <a-modal
      v-model:open="startModalOpen"
      width="760px"
      wrap-class-name="child-picker-modal"
      :footer="null"
      centered
    >
      <div class="modal-header">
        <h2>选择测评儿童</h2>
        <p>开始 {{ activeScale?.name || '量表' }} 测评前，请先选择本次测评对象。</p>
      </div>

      <div class="child-filter-bar">
        <a-input
          v-model:value="childSearchText"
          allow-clear
          placeholder="搜索儿童姓名 / 联系电话"
        >
          <template #suffix>
            <SearchOutlined />
          </template>
        </a-input>
        <a-select placeholder="班级" style="width: 108px">
          <a-select-option value="小海豚班">小海豚班</a-select-option>
          <a-select-option value="海星班">海星班</a-select-option>
        </a-select>
        <a-select placeholder="在读状态" style="width: 118px">
          <a-select-option value="在读">在读</a-select-option>
        </a-select>
        <a-select placeholder="测评状态" style="width: 118px">
          <a-select-option value="未测评">未测评</a-select-option>
        </a-select>
      </div>

      <div class="child-table">
        <div class="child-table__head">
          <span>儿童信息</span>
          <span>性别</span>
          <span>年龄</span>
          <span>联系电话</span>
          <span>班级</span>
          <span>最近测评</span>
          <span>操作</span>
        </div>
        <button
          v-for="child in filteredChildOptions"
          :key="child.id"
          type="button"
          class="child-row"
          :class="{ 'is-selected': child.id === selectedChildId }"
          @click="selectedChildId = child.id"
        >
          <span class="child-name">
            <i>{{ child.shortName }}</i>
            <strong>{{ child.name }}</strong>
          </span>
          <span>{{ child.gender }}</span>
          <span>{{ child.age }}</span>
          <span>{{ child.contactPhone }}</span>
          <span>{{ child.className }}</span>
          <span>{{ child.latestAssessment }}</span>
          <span class="radio-mark">
            <b v-if="child.id === selectedChildId"></b>
          </span>
        </button>
      </div>

      <div class="modal-footer">
        <div class="selected-child">
          已选择：<strong>{{ selectedChild?.name || '未选择' }}</strong>
        </div>
        <a-space>
          <a-button @click="startModalOpen = false">取消</a-button>
          <a-button type="primary" :disabled="!selectedChild" @click="confirmStartAssessment">
            确认并开始测评
          </a-button>
        </a-space>
      </div>
    </a-modal>

    <a-modal
      v-model:open="detailModalOpen"
      width="760px"
      title="量表介绍"
      :centered='true'
      wrap-class-name="scale-intro-modal"
      :footer="null"
      :body-style="{ maxHeight: 'calc(100vh - 150px)', overflowY: 'auto', padding: '16px' }"
    >
      <div v-if="activeScale" class="scale-intro-preview">
        <img :src="activeScale.posterUrl || scaleIntroImage" alt="量表介绍">
      </div>
    </a-modal>
  </div>
</template>

<style scoped lang="less">
.scale-library-page {
  color: #111827;
}

.category-tabs {
  display: flex;
  flex-wrap: nowrap;
  gap: 12px;
  width: 100%;
  max-width: 100%;
  padding: 0 0 12px 0;
  overflow-x: auto;
  overflow-y: hidden;
  scrollbar-width: thin;
  scrollbar-color: #c9d5ea transparent;

  &::-webkit-scrollbar {
    height: 8px;
  }

  &::-webkit-scrollbar-track {
    background: transparent;
  }

  &::-webkit-scrollbar-thumb {
    background: #c9d5ea;
    border-radius: 999px;
  }

  &::-webkit-scrollbar-thumb:hover {
    background: #b3c4df;
  }
}

.category-tab {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  flex: 0 1 auto;
  width: auto;
  min-width: max-content;
  max-width: none;
  height: 46px;
  padding: 0 18px;
  color: #344054;
  background: #fff;
  border: 1px solid #dfe6f1;
  border-radius: 8px;
  cursor: pointer;

  span {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-size: 14px;
    font-weight: 600;
  }

  b {
    flex: 0 0 auto;
    min-width: 28px;
    padding: 2px 8px;
    color: #667085;
    background: #f2f4f7;
    border-radius: 999px;
    font-weight: 600;
  }

  &.is-active {
    color: #145dff;
    border-color: #2f6bff;
    box-shadow: 0 6px 16px rgb(47 107 255 / 12%);
  }
}

.is-blue :deep(svg) {
  color: #2f6bff;
}

.is-orange :deep(svg) {
  color: #fb7b2b;
}

.is-purple :deep(svg) {
  color: #8a5cf6;
}

.summary-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 16px;
}

.summary-card {
  display: flex;
  align-items: center;
  gap: 10px;
  min-height: 62px;
  padding: 8px 12px;
  background: #fff;
  border: 1px solid #e6eaf0;
  border-radius: 8px;
}

.summary-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  border-radius: 8px;
  font-size: 21px;

  &.is-blue {
    color: #2367e8;
    background: #eef5ff;
  }

  &.is-green {
    color: #109a54;
    background: #eaf8f0;
  }

  &.is-orange {
    color: #ef6c00;
    background: #fff2e6;
  }

  &.is-purple {
    color: #7c4de8;
    background: #f2ebff;
  }
}

.summary-card__content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
  flex: 1 1 auto;
  min-width: 0;
}

.summary-label {
  color: #667085;
  font-size: 11px;
  line-height: 18px;
}

.summary-value {
  margin-top: 0;
  color: #111827;
  font-size: 20px;
  font-weight: 750;
  line-height: 22px;

  span {
    margin-left: 4px;
    color: #667085;
    font-size: 11px;
    font-weight: 500;
  }
}

.summary-desc {
  margin-top: 0;
  color: #667085;
  font-size: 12px;
  line-height: 18px;
  text-align: right;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  flex: 0 0 auto;
  max-width: 50%;
}

.library-workbench {
  display: grid;
  grid-template-columns: 208px minmax(0, 1fr);
  gap: 18px;
  margin-top: 18px;
  align-items: stretch;
  height: calc(100dvh - 248px);
  min-height: 560px;
  overflow: hidden;
}

.filter-panel {
  height: 100%;
  max-height: 100%;
  overflow-x: hidden;
  overflow-y: auto;
  padding: 16px;
  background: #fff;
  border: 1px solid #e6eaf0;
  border-radius: 8px;
  scrollbar-width: thin;
  scrollbar-color: #c9d5ea transparent;

  &::-webkit-scrollbar {
    width: 8px;
  }

  &::-webkit-scrollbar-track {
    background: transparent;
  }

  &::-webkit-scrollbar-thumb {
    background: #c9d5ea;
    border-radius: 999px;
  }

  &::-webkit-scrollbar-thumb:hover {
    background: #b3c4df;
  }
}

.filter-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 14px;

  strong {
    font-size: 15px;
  }
}

.filter-group {
  padding: 12px 0;
  border-top: 1px solid #edf1f6;

  :deep(.ant-checkbox-wrapper),
  :deep(.ant-radio-wrapper) {
    display: flex;
    align-items: center;
    justify-content: flex-start;
    width: 100%;
    margin: 8px 0;
    color: #475467;
    text-align: left;
    line-height: 20px;
  }

  :deep(.ant-checkbox),
  :deep(.ant-radio) {
    flex: 0 0 auto;
    margin-inline-end: 0;
  }

  span {
    color: #667085;
  }
}

.filter-radio-group {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
}

.filter-radio-group :deep(.ant-radio-wrapper:hover .ant-radio),
.filter-radio-group :deep(.ant-radio:hover .ant-radio-inner),
.filter-radio-group :deep(.ant-radio-input:focus + .ant-radio-inner) {
  border-color: var(--pro-ant-color-primary);
}

.filter-radio-group :deep(.ant-radio-inner) {
  background-color: transparent;
  border-color: #d9d9d9;
}

.filter-radio-group :deep(.ant-radio-checked .ant-radio-inner) {
  background-color: transparent;
  border-color: var(--pro-ant-color-primary);
}

.filter-radio-group :deep(.ant-radio-inner::after) {
  background-color: var(--pro-ant-color-primary);
  transform: scale(0.5);
}

.filter-radio-group :deep(.ant-radio-disabled .ant-radio-inner) {
  border-color: #d9d9d9;
}

.age-radio-group :deep(.ant-radio-wrapper) {
  margin: 4px 0;
}

.filter-group__title {
  margin-bottom: 8px;
  color: #111827;
  font-weight: 650;
}

.filter-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  height: 34px;
  padding: 0;
  color: #475467;
  background: transparent;
  border: 0;
  cursor: pointer;
}

.scale-content {
  min-width: 0;
  min-height: 0;
  display: flex;
  height: 100%;
}

.scale-content-scroll {
  flex: 1 1 auto;
  min-width: 0;
  height: 100%;
  min-height: 0;
  overflow-y: auto;
  padding-right: 4px;
  scrollbar-width: thin;
  scrollbar-color: #c9d5ea transparent;

  &::-webkit-scrollbar {
    width: 8px;
  }

  &::-webkit-scrollbar-track {
    background: transparent;
  }

  &::-webkit-scrollbar-thumb {
    background: #c9d5ea;
    border-radius: 999px;
  }

  &::-webkit-scrollbar-thumb:hover {
    background: #b3c4df;
  }
}

.scale-card-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(483.5px, 483.5px));
  justify-content: start;
  gap: 18px;
}

.scale-card {
  display: flex;
  flex-direction: column;
  min-height: 244px;
  padding: 16px 18px;
  background: #fff;
  border: 1px solid #e6eaf0;
  border-radius: 8px;
  transition: border-color .16s, box-shadow .16s;

  &:hover {
    border-color: #b8cdfa;
    box-shadow: 0 12px 28px rgb(15 23 42 / 8%);
  }
}

.scale-card__head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 14px;

  h2 {
    margin: 0 0 8px;
    color: #111827;
    font-size: 16px;
    font-weight: 700;
    line-height: 26px;
  }
}

.card-side {
  display: flex;
  flex: 0 0 auto;
  flex-direction: column;
  align-items: flex-end;
  gap: 8px;
}

.card-side-actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 8px;
}

.card-status {
  display: flex;
  align-items: center;
  flex: 0 0 auto;
  color: #667085;
}

.iep-trigger {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 5px;
  height: 26px;
  padding: 0 10px;
  color: #127a3d;
  background: #eefaf2;
  border: 1px solid #cfe9d8;
  border-radius: 6px;
  cursor: pointer;
  font-size: 12px;
  line-height: 1;

  &:hover {
    color: #0d6b33;
    background: #e5f6eb;
    border-color: #9fd5b2;
  }
}

.reference-trigger {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 5px;
  height: 26px;
  padding: 0 10px;
  color: #145dff;
  background: #f4f8ff;
  border: 1px solid #d7e5ff;
  border-radius: 6px;
  cursor: pointer;
  font-size: 12px;
  line-height: 1;
  min-width: 0;

  &:hover {
    color: #0f4fd7;
    background: #edf5ff;
    border-color: #9fc1ff;
  }
}

:global(.reference-popover) {
  max-width: 390px;
}

:global(.reference-popover .ant-popover-inner) {
  border-radius: 8px;
  box-shadow: 0 12px 30px rgb(15 23 42 / 12%);
}

.reference-popover-content {
  display: grid;
  gap: 12px;
  width: 340px;
  max-width: 72vw;
}

.reference-section {
  display: grid;
  gap: 8px;
}

.reference-section__title {
  color: #667085;
  font-size: 12px;
  font-weight: 650;
  line-height: 18px;
}

.reference-divider {
  height: 1px;
  background: #edf1f6;
}

.reference-item {
  display: grid;
  grid-template-columns: 20px minmax(0, 1fr);
  gap: 8px;
  color: #344054;

  span {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 20px;
    height: 20px;
    color: #145dff;
    background: #eef5ff;
    border-radius: 50%;
    font-size: 12px;
    font-weight: 650;
  }

  p {
    margin: 0;
    color: #475467;
    font-size: 13px;
    line-height: 20px;
  }
}

.acknowledgement-item {
  display: grid;
  grid-template-columns: 20px minmax(0, 1fr);
  gap: 8px;
  color: #344054;

  span {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 20px;
    height: 20px;
    color: #14804a;
    background: #eaf8f0;
    border-radius: 50%;
    font-size: 12px;
    font-weight: 650;
  }

  p {
    margin: 0;
    color: #475467;
    font-size: 13px;
    line-height: 20px;
  }
}

.status-pill {
  display: inline-flex;
  align-items: center;
  height: 26px;
  padding: 0 10px;
  border-radius: 999px;
  font-size: 13px;
  font-weight: 650;

  &.is-available {
    color: #127a3d;
    background: #dff7e7;
  }

  &.is-disabled {
    color: #667085;
    background: #f2f4f7;
  }
}

.tag-list,
.detail-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.tag-list span,
.detail-tags span {
  padding: 4px 6px;
  border-radius: 6px;
  font-size: 12px;
  line-height: 18px;
}

.tag-blue {
  color: #145dff;
  background: #eef5ff;
}

.tag-green {
  color: #14804a;
  background: #eaf8f0;
}

.tag-purple {
  color: #7c4de8;
  background: #f2ebff;
}

.scale-desc {
  margin: 12px 0 0px;
  color: #475467;
  font-size: 14px;
  line-height: 23px;
}

.scale-meta {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 8px 14px;
  margin-top: 12px;

  div {
    display: grid;
    grid-template-columns: 18px 64px minmax(0, 1fr);
    align-items: center;
    gap: 8px;
    color: #667085;
    font-size: 13px;
  }

  b {
    overflow: hidden;
    color: #344054;
    font-weight: 600;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.scale-card__footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid #edf1f6;
  color: #667085;
  font-size: 13px;
}

.card-actions {
  display: grid;
  grid-template-columns: 118px 132px;
  gap: 10px;
}

.modal-header {
  margin-bottom: 16px;

  h2 {
    margin: 0;
    color: #111827;
    font-size: 22px;
    font-weight: 700;
  }

  p {
    margin: 6px 0 0;
    color: #667085;
  }
}

.child-filter-bar {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 108px 118px 118px;
  gap: 12px;
  margin-bottom: 14px;
}

.child-table {
  overflow: hidden;
  border: 1px solid #e6eaf0;
  border-radius: 8px;
}

.child-table__head,
.child-row {
  display: grid;
  grid-template-columns: 1.08fr .52fr .75fr 1.28fr .84fr .9fr .38fr;
  align-items: center;
  column-gap: 10px;
}

.child-table__head {
  min-height: 40px;
  padding: 0 12px;
  color: #667085;
  background: #f8fafc;
  font-size: 13px;
}

.child-row {
  width: 100%;
  min-height: 62px;
  padding: 0 12px;
  color: #344054;
  text-align: left;
  background: #fff;
  border: 0;
  border-top: 1px solid #edf1f6;
  cursor: pointer;

  &.is-selected {
    position: relative;
    z-index: 1;
    background: #f4f8ff;
    box-shadow: inset 0 0 0 1px #2f6bff;
  }
}

.child-name {
  display: flex;
  align-items: center;
  gap: 5px;

  i {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 34px;
    height: 34px;
    color: #145dff;
    background: #eef5ff;
    border-radius: 50%;
    font-style: normal;
    font-weight: 700;
  }
}

.radio-mark {
  justify-self: center;
  width: 18px;
  height: 18px;
  border: 1px solid #b8c2d6;
  border-radius: 50%;

  b {
    display: block;
    width: 10px;
    height: 10px;
    margin: 3px;
    background: #2f6bff;
    border-radius: 50%;
  }
}

.modal-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 18px;
}

.selected-child {
  color: #111827;
  font-size: 16px;

  strong {
    color: #145dff;
  }
}

.scale-intro-preview {
  overflow: hidden;
  background: #f8fafc;
  border: 1px solid #edf1f6;
  border-radius: 8px;

  img {
    display: block;
    width: 100%;
    height: auto;
  }
}

.detail-content {
  h2 {
    margin: 0 0 8px;
    color: #111827;
    font-size: 21px;
    font-weight: 700;
  }

  p {
    color: #475467;
    line-height: 24px;
  }
}

.detail-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
  margin-top: 18px;

  div {
    padding: 14px;
    background: #f8fafc;
    border: 1px solid #edf1f6;
    border-radius: 8px;
  }

  span {
    display: block;
    color: #667085;
    font-size: 13px;
  }

  b {
    display: block;
    margin-top: 6px;
    color: #111827;
    font-size: 18px;
  }
}

.detail-reference {
  margin-top: 16px;
  padding: 14px;
  background: #f8fafc;
  border: 1px solid #edf1f6;
  border-radius: 8px;

  span {
    display: block;
    color: #667085;
    font-size: 13px;
  }

  p {
    margin: 6px 0 0;
    color: #344054;
    font-size: 14px;
    line-height: 22px;
  }

  ul {
    display: grid;
    gap: 8px;
    margin: 8px 0 0;
    padding: 0;
    list-style: none;
  }

  li {
    position: relative;
    padding-left: 16px;
    color: #344054;
    font-size: 14px;
    line-height: 22px;

    &::before {
      position: absolute;
      top: 9px;
      left: 2px;
      width: 5px;
      height: 5px;
      background: #98a2b3;
      border-radius: 50%;
      content: '';
    }
  }
}

.detail-acknowledgements {
  margin-top: 12px;
  padding: 14px;
  background: #f8fafc;
  border: 1px solid #edf1f6;
  border-radius: 8px;

  span {
    display: block;
    color: #667085;
    font-size: 13px;
  }

  ul {
    display: grid;
    gap: 8px;
    margin: 8px 0 0;
    padding: 0;
    list-style: none;
  }

  li {
    position: relative;
    padding-left: 16px;
    color: #344054;
    font-size: 14px;
    line-height: 22px;

    &::before {
      position: absolute;
      top: 9px;
      left: 2px;
      width: 5px;
      height: 5px;
      background: #98a2b3;
      border-radius: 50%;
      content: '';
    }
  }
}

@media (max-width: 1280px) {
  .library-workbench {
    grid-template-columns: 1fr;
    height: auto;
    min-height: 0;
    overflow: visible;
  }

  .filter-panel {
    display: none;
  }

  .scale-content-scroll {
    overflow: visible;
    padding-right: 0;
  }

  .summary-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .category-tabs {
    flex-wrap: nowrap;
  }

  .scale-card-grid {
    grid-template-columns: 1fr;
  }
}
</style>
