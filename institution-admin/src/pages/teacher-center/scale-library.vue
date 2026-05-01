<script setup lang="ts">
import {
  AppstoreOutlined,
  BarsOutlined,
  CheckCircleFilled,
  ClockCircleOutlined,
  ExperimentOutlined,
  BookOutlined,
  FileDoneOutlined,
  FileTextOutlined,
  HeartOutlined,
  SearchOutlined,
  TeamOutlined,
} from '@ant-design/icons-vue'
import { computed, ref } from 'vue'
import messageService from '@/utils/messageService'

type ScaleStatus = 'available' | 'disabled'

interface ScaleCard {
  id: number
  name: string
  description: string
  tags: string[]
  category: string
  ageRange: string
  status: ScaleStatus
  itemCount: number
  domainCount: number
  duration: string
  latestUse: string
  usageCount: number
  references: string[]
  acknowledgements: string[]
}

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
const selectedChildId = ref<number>(10015)
const startModalOpen = ref(false)
const detailModalOpen = ref(false)
const activeScale = ref<ScaleCard>()

const categoryTabs = [
  { key: 'all', label: '全部', count: 68, color: 'blue', icon: AppstoreOutlined },
  { key: 'standard', label: '标准化测评', count: 24, color: 'blue', icon: CheckCircleFilled },
  { key: 'screening', label: '筛查量表', count: 18, color: 'orange', icon: ExperimentOutlined },
  { key: 'development', label: '发展评估', count: 10, color: 'orange', icon: TeamOutlined },
  { key: 'social', label: '社交行为', count: 8, color: 'purple', icon: HeartOutlined },
  { key: 'sensory', label: '感觉统合', count: 6, color: 'orange', icon: ExperimentOutlined },
  { key: 'emotion', label: '情绪行为', count: 7, color: 'purple', icon: HeartOutlined },
]

const summaryCards = [
  { label: '全部量表', value: 68, suffix: '个', desc: '覆盖8大类评估领域', icon: AppstoreOutlined, tone: 'blue' },
  { label: '可用量表', value: 58, suffix: '个', desc: '可正常发起测评', icon: CheckCircleFilled, tone: 'green' },
  { label: '停用量表', value: 10, suffix: '个', desc: '暂不可发起测评', icon: FileTextOutlined, tone: 'orange' },
  { label: '本月已测评', value: 156, suffix: '次', desc: '较上月 ↑ 23%', icon: BarsOutlined, tone: 'purple' },
]

const scaleCards: ScaleCard[] = [
  {
    id: 1,
    name: 'PEP-3 儿童心理教育评核',
    description: '评估儿童认知、语言、运动、行为及照顾者报告表现。',
    tags: ['标准化测评', '2-7岁', '阶段复测'],
    category: '标准化测评',
    ageRange: '2-7岁',
    status: 'available',
    itemCount: 172,
    domainCount: 13,
    duration: '45-90分钟',
    latestUse: '2025-05-12',
    usageCount: 128,
    references: [
      'Schopler, E., Lansing, M. D., Reichler, R. J., & Marcus, L. M. (2005). Psychoeducational Profile: Third Edition (PEP-3). PRO-ED.',
      'PEP-3 中文版手册及机构本土化施测记录规范。',
    ],
    acknowledgements: [
      '王晓琳博士（儿童发展评估顾问）',
      '陈志远老师（PEP-3 施测支持）',
    ],
  },
  {
    id: 2,
    name: '感觉统合能力筛查',
    description: '筛查儿童感觉统合发展状况与功能水平。',
    tags: ['筛查量表', '3-12岁', '专项筛查'],
    category: '筛查量表',
    ageRange: '3-12岁',
    status: 'available',
    itemCount: 62,
    domainCount: 16,
    duration: '15-30分钟',
    latestUse: '2025-05-08',
    usageCount: 96,
    references: [
      'Dunn, W. (1999). Sensory Profile: User\'s Manual. Psychological Corporation.',
      '儿童感觉统合评估工具与机构专项筛查条目。',
    ],
    acknowledgements: [
      '刘文静老师（感觉统合顾问）',
      '赵明教授（筛查工具修订）',
    ],
  },
  {
    id: 3,
    name: '社交互动能力评估',
    description: '评估儿童社交理解、社交表达及社交互动能力。',
    tags: ['社交行为', '4-12岁', '日常跟踪'],
    category: '社交行为',
    ageRange: '4-12岁',
    status: 'available',
    itemCount: 88,
    domainCount: 6,
    duration: '20-40分钟',
    latestUse: '2025-05-10',
    usageCount: 67,
    references: [
      '参考社交沟通与互动观察量表编制规范。',
      '结合机构本土化评估条目与课堂观察记录形成。',
    ],
    acknowledgements: [
      '周嘉宁老师（社交行为观察）',
      '何雪梅老师（课堂记录支持）',
    ],
  },
  {
    id: 4,
    name: '语言沟通能力评估',
    description: '覆盖语言理解、语言表达、语用沟通及互动意图。',
    tags: ['发展评估', '2-8岁', '阶段复测'],
    category: '发展评估',
    ageRange: '2-8岁',
    status: 'disabled',
    itemCount: 96,
    domainCount: 11,
    duration: '30-45分钟',
    latestUse: '2025-05-09',
    usageCount: 12,
    references: [
      '参考儿童语言发展评估与语用沟通观察资料。',
      '结合机构教学评估场景与阶段复测记录形成。',
    ],
    acknowledgements: [
      '唐可欣博士（语言发展顾问）',
      '宋雨晴老师（语用沟通支持）',
    ],
  },
  {
    id: 5,
    name: '行为观察记录表',
    description: '用于记录儿童在自然情境中的行为表现与频率。',
    tags: ['情绪行为', '2-12岁', '日常跟踪'],
    category: '情绪行为',
    ageRange: '2-12岁',
    status: 'disabled',
    itemCount: 36,
    domainCount: 4,
    duration: '10-20分钟',
    latestUse: '2025-05-07',
    usageCount: 8,
    references: [
      '参考应用行为分析观察记录方法。',
      '结合机构课堂、训练场景与日常跟踪记录形成。',
    ],
    acknowledgements: [
      '吴思远老师（行为观察顾问）',
      '郑雅宁老师（ABA 记录支持）',
    ],
  },
  {
    id: 6,
    name: '生活自理能力评估',
    description: '评估儿童进食、穿脱、如厕、清洁等日常自理能力。',
    tags: ['康复评估', '3-10岁', '结案评估'],
    category: '康复评估',
    ageRange: '3-10岁',
    status: 'available',
    itemCount: 54,
    domainCount: 7,
    duration: '20-30分钟',
    latestUse: '2025-05-06',
    usageCount: 43,
    references: [
      '参考儿童日常生活能力评估与康复训练记录资料。',
      '结合机构康复目标与结案评估场景形成。',
    ],
    acknowledgements: [
      '林安然老师（康复训练顾问）',
      '黄若彤老师（日常生活能力支持）',
    ],
  },
]

const childOptions: ChildOption[] = [
  { id: 10012, shortName: '乐', name: '乐乐', gender: '男', age: '5岁2个月', contactPhone: '爸爸 176****0124', className: '小海豚班', latestAssessment: '2025-04-28' },
  { id: 10015, shortName: '小', name: '小宇', gender: '男', age: '4岁8个月', contactPhone: '妈妈 188****5630', className: '小海豚班', latestAssessment: '2025-05-12' },
  { id: 10009, shortName: '安', name: '安安', gender: '女', age: '6岁1个月', contactPhone: '妈妈 139****8902', className: '小海豚班', latestAssessment: '2025-04-15' },
  { id: 10007, shortName: '浩', name: '浩浩', gender: '男', age: '7岁3个月', contactPhone: '爸爸 185****7721', className: '海星班', latestAssessment: '2025-05-10' },
  { id: 10021, shortName: '糖', name: '糖糖', gender: '女', age: '3岁11个月', contactPhone: '妈妈 177****3056', className: '海星班', latestAssessment: '2025-04-20' },
]

const filteredScales = computed(() => {
  const keyword = searchText.value.trim()
  if (!keyword)
    return scaleCards
  return scaleCards.filter(item => `${item.name}${item.description}${item.tags.join('')}${item.ageRange}`.includes(keyword))
})

const filteredChildOptions = computed(() => {
  const keyword = childSearchText.value.trim()
  if (!keyword)
    return childOptions
  return childOptions.filter(item => `${item.name}${item.contactPhone}${item.className}${item.latestAssessment}${item.gender}${item.age}`.includes(keyword))
})

const selectedChild = computed(() => childOptions.find(item => item.id === selectedChildId.value))

function statusMeta(status: ScaleStatus) {
  const map: Record<ScaleStatus, { text: string, className: string }> = {
    available: { text: '可用', className: 'is-available' },
    disabled: { text: '停用', className: 'is-disabled' },
  }
  return map[status]
}

function tagClass(index: number) {
  return ['tag-blue', 'tag-green', 'tag-purple'][index % 3]
}

function openStartModal(scale: ScaleCard) {
  if (scale.status === 'disabled') {
    messageService.warning('该量表已停用，暂不可发起测评')
    return
  }
  activeScale.value = scale
  startModalOpen.value = true
}

function openDetailModal(scale: ScaleCard) {
  activeScale.value = scale
  detailModalOpen.value = true
}

function openIepLibrary(scale: ScaleCard) {
  messageService.info(`进入 ${scale.name} 的 IEP 库配置`)
}

function confirmStartAssessment() {
  if (!activeScale.value || !selectedChild.value)
    return
  messageService.success(`已选择 ${selectedChild.value.name}，准备开始 ${activeScale.value.name}`)
  startModalOpen.value = false
}
</script>

<template>
  <div class="scale-library-page">
    <section class="category-tabs">
      <button
        v-for="item in categoryTabs"
        :key="item.key"
        type="button"
        class="category-tab"
        :class="[`is-${item.color}`, { 'is-active': item.key === 'all' }]"
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
          <a>清空</a>
        </div>

        <div class="filter-group">
          <div class="filter-group__title">状态</div>
          <a-checkbox checked>全部状态</a-checkbox>
          <a-checkbox>可用 <span>58</span></a-checkbox>
          <a-checkbox>停用 <span>10</span></a-checkbox>
        </div>

        <div class="filter-group">
          <div class="filter-group__title">适用年龄</div>
          <a-radio-group v-model:value="ageScope" class="custom-radio filter-radio-group age-radio-group">
            <a-radio value="all">全部年龄</a-radio>
            <a-radio value="0-2">0-2岁</a-radio>
            <a-radio value="2-6">2-6岁</a-radio>
            <a-radio value="6-12">6-12岁</a-radio>
            <a-radio value="12+">12岁以上</a-radio>
          </a-radio-group>
        </div>

        <div class="filter-group">
          <div class="filter-group__title">使用场景</div>
          <a-checkbox>入园评估</a-checkbox>
          <a-checkbox>阶段复测</a-checkbox>
          <a-checkbox>结案评估</a-checkbox>
          <a-checkbox>日常跟踪</a-checkbox>
          <a-checkbox>专项筛查</a-checkbox>
        </div>

        <div class="filter-group">
          <div class="filter-group__title">测评时长</div>
          <a-checkbox>15分钟以内</a-checkbox>
          <a-checkbox>15-30分钟</a-checkbox>
          <a-checkbox>30-60分钟</a-checkbox>
          <a-checkbox>60分钟以上</a-checkbox>
        </div>
      </aside>

      <main class="scale-content">
        <div class="scale-content-scroll">
          <div class="scale-card-grid">
            <article v-for="scale in filteredScales" :key="scale.id" class="scale-card">
              <div class="scale-card__head">
                <div>
                  <h2>{{ scale.name }}</h2>
                  <div class="tag-list">
                    <span v-for="(tag, index) in scale.tags" :key="tag" :class="tagClass(index)">{{ tag }}</span>
                  </div>
                </div>
                <div class="card-side">
                  <div class="card-status">
                    <span class="status-pill" :class="statusMeta(scale.status).className">
                      {{ statusMeta(scale.status).text }}
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
                            :key="reference"
                            class="reference-item"
                          >
                            <span>{{ referenceIndex + 1 }}</span>
                            <p>{{ reference }}</p>
                          </div>
                        </div>
                        <div class="reference-divider"></div>
                        <div class="reference-section">
                          <div class="reference-section__title">特别鸣谢</div>
                          <div
                            v-for="(acknowledgement, acknowledgementIndex) in scale.acknowledgements"
                            :key="acknowledgement"
                            class="acknowledgement-item"
                          >
                            <span>{{ acknowledgementIndex + 1 }}</span>
                            <p>{{ acknowledgement }}</p>
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

              <p class="scale-desc">{{ scale.description }}</p>

              <div class="scale-meta">
                <div><FileTextOutlined /><span>题目数量</span><b>{{ scale.itemCount }}题</b></div>
                <div><TeamOutlined /><span>评估维度</span><b>{{ scale.domainCount }}个</b></div>
                <div><ClockCircleOutlined /><span>测评时长</span><b>{{ scale.duration }}</b></div>
                <div><FileDoneOutlined /><span>最近使用</span><b>{{ scale.latestUse }}</b></div>
              </div>

              <div class="scale-card__footer">
                <span>使用次数 {{ scale.usageCount }}次</span>
                <div class="card-actions">
                  <a-button @click="openDetailModal(scale)">查看详情</a-button>
                  <a-button
                    type="primary"
                    :disabled="scale.status === 'disabled'"
                    @click="openStartModal(scale)"
                  >
                    开始测评
                  </a-button>
                </div>
              </div>
            </article>
          </div>
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
      width="680px"
      title="量表详情"
      ok-text="开始测评"
      cancel-text="关闭"
      :ok-button-props="{ disabled: activeScale?.status === 'disabled' }"
      @ok="activeScale && openStartModal(activeScale); detailModalOpen = false"
    >
      <div v-if="activeScale" class="detail-content">
        <h2>{{ activeScale.name }}</h2>
        <p>{{ activeScale.description }}</p>
        <div class="detail-tags">
          <span v-for="(tag, index) in activeScale.tags" :key="tag" :class="tagClass(index)">
            {{ tag }}
          </span>
        </div>
        <div class="detail-grid">
          <div><span>题目数量</span><b>{{ activeScale.itemCount }}题</b></div>
          <div><span>评估维度</span><b>{{ activeScale.domainCount }}个</b></div>
          <div><span>测评时长</span><b>{{ activeScale.duration }}</b></div>
          <div><span>最近使用</span><b>{{ activeScale.latestUse }}</b></div>
        </div>
        <div class="detail-reference">
          <span>引用文献</span>
          <ul>
            <li v-for="reference in activeScale.references" :key="reference">
              {{ reference }}
            </li>
          </ul>
        </div>
        <div class="detail-acknowledgements">
          <span>特别鸣谢</span>
          <ul>
            <li v-for="acknowledgement in activeScale.acknowledgements" :key="acknowledgement">
              {{ acknowledgement }}
            </li>
          </ul>
        </div>
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
  flex: 1 0 148px;
  min-width: 148px;
  height: 46px;
  padding: 0 18px;
  color: #344054;
  background: #fff;
  border: 1px solid #dfe6f1;
  border-radius: 8px;
  cursor: pointer;

  span {
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
  grid-template-columns: repeat(2, minmax(0, 1fr));
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
  padding: 4px 10px;
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
}
</style>
