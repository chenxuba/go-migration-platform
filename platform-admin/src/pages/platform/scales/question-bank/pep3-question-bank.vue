<script setup lang="ts">
import type { TableColumnsType } from 'ant-design-vue'
import { Empty } from 'ant-design-vue'
import { CloseOutlined, EditOutlined, ReloadOutlined, SearchOutlined } from '@ant-design/icons-vue'
import { computed, onMounted, reactive, ref, watch } from 'vue'
import {
  getScaleQuestionBankApi,
  type ScaleQuestionBank,
  type ScaleQuestionBankItem,
  type ScaleQuestionBankRecordField,
  type ScaleQuestionBankRecordFieldOption,
  type ScaleQuestionBankScoreOption,
  updateScaleQuestionBankItemApi,
} from '@/api/platform/scales'
import { useQueryBreakpoints } from '@/composables/query-breakpoints'
import messageService from '@/utils/messageService'

const props = defineProps<{
  scaleCode: string
  scaleVersion?: string
}>()

const loading = ref(false)
const saving = ref(false)
const questionBank = ref<ScaleQuestionBank | null>(null)
const activeDomainCode = ref('all')
const keyword = ref('')
const appliedKeyword = ref('')
const editOpen = ref(false)
const simpleEmptyImage = Empty.PRESENTED_IMAGE_SIMPLE
const { isMobile, isPad } = useQueryBreakpoints()

const drawerWidth = computed(() => {
  if (isMobile.value)
    return '100%'
  if (isPad.value)
    return '90%'
  return '920px'
})

const columns: TableColumnsType<ScaleQuestionBankItem> = [
  { title: '题号', key: 'itemNo', width: 72, fixed: 'left' as const, align: 'center' as const },
  { title: '题目', key: 'title', width: 240 },
  { title: '操作标准', key: 'method', width: 320 },
  { title: '指导语', key: 'guidance', width: 240 },
  { title: '选项', key: 'scoreOptions', width: 320 },
  { title: '儿童表现记录', key: 'recordFields', width: 220 },
  { title: '操作', key: 'action', width: 86, fixed: 'right' as const, align: 'center' as const },
]

const fieldTypeOptions = [
  { label: '填空', value: 'text' },
  { label: '长文本', value: 'textarea' },
  { label: '数字', value: 'number' },
  { label: '单选', value: 'radio' },
  { label: '多选', value: 'checkbox_group' },
]

const textConversionMap: Record<string, string> = {
  兒: '儿',
  兩: '两',
  內: '内',
  個: '个',
  冊: '册',
  則: '则',
  創: '创',
  劃: '划',
  動: '动',
  務: '务',
  區: '区',
  協: '协',
  單: '单',
  問: '问',
  圖: '图',
  圓: '圆',
  報: '报',
  塊: '块',
  墊: '垫',
  實: '实',
  對: '对',
  導: '导',
  將: '将',
  尋: '寻',
  張: '张',
  彈: '弹',
  後: '后',
  從: '从',
  復: '复',
  慣: '惯',
  應: '应',
  戲: '戏',
  擊: '击',
  擇: '择',
  擾: '扰',
  於: '于',
  時: '时',
  書: '书',
  會: '会',
  標: '标',
  樣: '样',
  機: '机',
  檢: '检',
  歡: '欢',
  歲: '岁',
  氣: '气',
  測: '测',
  為: '为',
  無: '无',
  獨: '独',
  現: '现',
  畫: '画',
  當: '当',
  發: '发',
  皺: '皱',
  監: '监',
  盤: '盘',
  眾: '众',
  睜: '睁',
  確: '确',
  禮: '礼',
  稱: '称',
  積: '积',
  續: '续',
  聯: '联',
  聽: '听',
  聲: '声',
  聳: '耸',
  肌: '肌',
  與: '与',
  興: '兴',
  舉: '举',
  著: '着',
  處: '处',
  見: '见',
  視: '视',
  規: '规',
  覺: '觉',
  觀: '观',
  觸: '触',
  訊: '讯',
  記: '记',
  設: '设',
  試: '试',
  話: '话',
  認: '认',
  語: '语',
  說: '说',
  説: '说',
  課: '课',
  請: '请',
  論: '论',
  識: '识',
  護: '护',
  貼: '贴',
  資: '资',
  過: '过',
  達: '达',
  適: '适',
  選: '选',
  還: '还',
  邊: '边',
  開: '开',
  閉: '闭',
  間: '间',
  關: '关',
  雙: '双',
  雜: '杂',
  題: '题',
  額: '额',
  顏: '颜',
  顯: '显',
  體: '体',
  這: '这',
  給: '给',
  蓋: '盖',
  範: '范',
  嘗: '尝',
  嚐: '尝',
  萬: '万',
  轉: '转',
  穩: '稳',
  輪: '轮',
  搖: '摇',
  鈴: '铃',
  響: '响',
  輕: '轻',
  膠: '胶',
  餅: '饼',
  狀: '状',
  燭: '烛',
  隻: '只',
  貓: '猫',
  親: '亲',
  參: '参',
  經: '经',
  讓: '让',
  遊: '游',
  進: '进',
  疊: '叠',
  詢: '询',
  異: '异',
  們: '们',
  須: '须',
  風: '风',
  裏: '里',
  類: '类',
  質: '质',
  環: '环',
  境: '境',
  遲: '迟',
  難: '难',
  線: '线',
  錯: '错',
  雖: '虽',
  詞: '词',
  彙: '汇',
  溝: '沟',
  邏: '逻',
  輯: '辑',
}

const textConversionPattern = new RegExp(`[${Object.keys(textConversionMap).join('')}]`, 'g')

const form = reactive<ScaleQuestionBankItem & { scaleCode: string, scaleVersion: string }>({
  scaleCode: '',
  scaleVersion: '',
  itemNo: 0,
  itemTitle: '',
  testItem: '',
  materials: '',
  method: '',
  guidance: '',
  domainCode: '',
  domainName: '',
  standard: '',
  scoreOptions: [],
  scoreOptionText: '',
  recordFields: [],
  sourcePages: [],
})

const domainOptions = computed(() => {
  const domains = questionBank.value?.domains || []
  return [
    { label: '全部维度', value: 'all' },
    ...domains.map(item => ({
      label: `${item.scaleCode} ${item.scaleName}`,
      value: item.scaleCode,
    })),
  ]
})

const domainNameMap = computed(() => {
  const map = new Map<string, string>()
  for (const item of questionBank.value?.domains || [])
    map.set(item.scaleCode, item.scaleName)
  return map
})

const itemCountByDomain = computed(() => {
  const map = new Map<string, number>()
  for (const item of questionBank.value?.items || [])
    map.set(item.domainCode, (map.get(item.domainCode) || 0) + 1)
  return map
})

const recordFieldItemCount = computed(() => {
  return (questionBank.value?.items || []).filter(item => item.recordFields?.length).length
})

const filteredItems = computed(() => {
  const key = appliedKeyword.value.trim().toLowerCase()
  return (questionBank.value?.items || []).filter((item) => {
    if (activeDomainCode.value !== 'all' && item.domainCode !== activeDomainCode.value)
      return false
    if (!key)
      return true
    return [
      item.itemNo,
      item.itemTitle,
      item.testItem,
      item.domainCode,
      item.domainName,
      item.guidance,
      item.method,
      item.standard,
    ].join(' ').toLowerCase().includes(key)
  })
})

function recordFieldTypeLabel(type: string) {
  return fieldTypeOptions.find(item => item.value === type)?.label || type || '--'
}

function normalizeQuestionBankText(text?: string | number | null) {
  return String(text ?? '')
    .replace(textConversionPattern, char => textConversionMap[char] || char)
    .replace(/[「」]/g, match => (match === '「' ? '“' : '”'))
    .replace(/\s*\r?\n+\s*/g, ' ')
    .replace(/[ \t]{2,}/g, ' ')
    .trim()
}

function normalizeOptions(options?: ScaleQuestionBankRecordFieldOption[] | null) {
  return (Array.isArray(options) ? options : []).map(item => ({
    value: normalizeQuestionBankText(item.value || item.label || ''),
    label: normalizeQuestionBankText(item.label || item.value || ''),
  })).filter(item => item.value || item.label)
}

function cloneRecordFields(fields?: ScaleQuestionBankRecordField[] | null) {
  return (Array.isArray(fields) ? fields : []).map(field => ({
    key: normalizeQuestionBankText(field.key || ''),
    label: normalizeQuestionBankText(field.label || ''),
    fieldType: field.fieldType || 'text',
    displayType: field.displayType || recordFieldTypeLabel(field.fieldType || 'text'),
    required: Boolean(field.required),
    placeholder: normalizeQuestionBankText(field.placeholder || ''),
    options: normalizeOptions(field.options),
  }))
}

function cloneScoreOptions(options?: ScaleQuestionBankScoreOption[] | null) {
  return (Array.isArray(options) ? options : []).map(option => ({
    value: Number(option.value),
    label: normalizeQuestionBankText(option.label || `${option.value}分`),
    description: normalizeQuestionBankText(option.description || ''),
  }))
}

function normalizeNumberArray(values?: number[] | null) {
  return (Array.isArray(values) ? values : [])
    .map(value => Number(value))
    .filter(value => Number.isFinite(value))
}

function normalizeQuestionBank(data: ScaleQuestionBank): ScaleQuestionBank {
  const domains = (Array.isArray(data.domains) ? data.domains : []).map(domain => ({
    ...domain,
    scaleCode: domain.scaleCode || '',
    scaleName: domain.scaleName || '',
    category: domain.category || '',
    itemCount: Number(domain.itemCount) || 0,
    maxRawScore: Number(domain.maxRawScore) || 0,
    itemNumbers: normalizeNumberArray(domain.itemNumbers),
    compositeCode: domain.compositeCode || '',
  }))
  const items = (Array.isArray(data.items) ? data.items : []).map((item) => {
    const scoreOptions = cloneScoreOptions(item.scoreOptions)
    return {
      ...item,
      itemNo: Number(item.itemNo) || 0,
      itemTitle: normalizeQuestionBankText(item.itemTitle),
      testItem: normalizeQuestionBankText(item.testItem),
      materials: normalizeQuestionBankText(item.materials),
      method: normalizeQuestionBankText(item.method),
      guidance: normalizeQuestionBankText(item.guidance),
      domainCode: normalizeQuestionBankText(item.domainCode),
      domainName: normalizeQuestionBankText(item.domainName),
      standard: normalizeQuestionBankText(item.standard),
      scoreOptions,
      scoreOptionText: normalizeQuestionBankText(item.scoreOptionText || scoreOptions.map(option => option.value).join('/')),
      recordFields: cloneRecordFields(item.recordFields),
      sourcePdf: normalizeQuestionBankText(item.sourcePdf),
      sourcePages: normalizeNumberArray(item.sourcePages),
      ocrStatus: normalizeQuestionBankText(item.ocrStatus),
      updatedAt: normalizeQuestionBankText(item.updatedAt),
    }
  })

  return {
    ...data,
    scaleCode: data.scaleCode || props.scaleCode,
    scaleVersion: data.scaleVersion || props.scaleVersion || '',
    dataStatus: data.dataStatus || '',
    itemCount: Number(data.itemCount) || items.length,
    domainCount: Number(data.domainCount) || domains.length,
    domains,
    items,
    sourceTables: Array.isArray(data.sourceTables) ? data.sourceTables : [],
  }
}

function fillForm(record: ScaleQuestionBankItem) {
  form.scaleCode = questionBank.value?.scaleCode || props.scaleCode
  form.scaleVersion = questionBank.value?.scaleVersion || props.scaleVersion || ''
  form.itemNo = record.itemNo
  form.itemTitle = normalizeQuestionBankText(record.itemTitle)
  form.testItem = normalizeQuestionBankText(record.testItem)
  form.materials = normalizeQuestionBankText(record.materials)
  form.method = normalizeQuestionBankText(record.method)
  form.guidance = normalizeQuestionBankText(record.guidance)
  form.domainCode = normalizeQuestionBankText(record.domainCode)
  form.domainName = normalizeQuestionBankText(record.domainName)
  form.standard = normalizeQuestionBankText(record.standard)
  form.scoreOptions = cloneScoreOptions(record.scoreOptions)
  form.scoreOptionText = normalizeQuestionBankText(record.scoreOptionText || form.scoreOptions.map(item => item.value).join('/'))
  form.recordFields = cloneRecordFields(record.recordFields)
  form.sourcePages = [...(record.sourcePages || [])]
}

function openEdit(record: ScaleQuestionBankItem) {
  fillForm(record)
  editOpen.value = true
}

function openEditRecord(record: Record<string, any>) {
  openEdit(record as ScaleQuestionBankItem)
}

function closeEdit() {
  editOpen.value = false
}

function handleSearch() {
  appliedKeyword.value = keyword.value.trim()
}

function resetSearch() {
  keyword.value = ''
  appliedKeyword.value = ''
}

function syncDomainName() {
  form.domainName = domainNameMap.value.get(form.domainCode) || form.domainName
}

function buildScoreStandard(options: ScaleQuestionBankScoreOption[]) {
  const lines = options
    .map(option => ({
      value: Number(option.value),
      description: normalizeQuestionBankText(option.description || ''),
    }))
    .filter(option => Number.isFinite(option.value))
    .map(option => `${option.value}- ${option.description}`.trim())
  return lines.length ? lines.join('\n') : normalizeQuestionBankText(form.standard)
}

function buildPayload() {
  const scoreOptions = form.scoreOptions.map(item => ({
    value: Number(item.value),
    label: normalizeQuestionBankText(item.label || `${item.value}分`),
    description: normalizeQuestionBankText(item.description || ''),
  }))
  return {
    scaleCode: form.scaleCode,
    scaleVersion: form.scaleVersion,
    itemNo: form.itemNo,
    itemTitle: normalizeQuestionBankText(form.itemTitle),
    testItem: normalizeQuestionBankText(form.testItem),
    materials: normalizeQuestionBankText(form.materials),
    method: normalizeQuestionBankText(form.method),
    guidance: normalizeQuestionBankText(form.guidance),
    domainCode: normalizeQuestionBankText(form.domainCode),
    domainName: normalizeQuestionBankText(domainNameMap.value.get(form.domainCode) || form.domainName || ''),
    standard: buildScoreStandard(scoreOptions),
    scoreOptions,
    scoreOptionText: normalizeQuestionBankText(form.scoreOptionText),
    recordFields: form.recordFields.map(field => ({
      key: normalizeQuestionBankText(field.key),
      label: normalizeQuestionBankText(field.label),
      fieldType: field.fieldType,
      displayType: field.displayType || recordFieldTypeLabel(field.fieldType),
      required: Boolean(field.required),
      placeholder: normalizeQuestionBankText(field.placeholder || ''),
      options: normalizeOptions(field.options),
    })).filter(field => field.key && field.label && field.fieldType),
    sourcePages: form.sourcePages,
  }
}

async function submitEdit() {
  if (!form.itemTitle.trim()) {
    messageService.warning('请输入题目')
    return
  }
  if (!form.domainCode.trim()) {
    messageService.warning('请选择维度')
    return
  }
  saving.value = true
  try {
    const res = await updateScaleQuestionBankItemApi(buildPayload())
    if (res.code !== 200) {
      messageService.error(res.message || '题目保存失败')
      return
    }
    messageService.success('题目已保存')
    editOpen.value = false
    await loadQuestionBank()
  }
  catch (error: any) {
    console.error('save question bank item failed', error)
    messageService.error(error?.response?.data?.message || error?.message || '题目保存失败')
  }
  finally {
    saving.value = false
  }
}

async function loadQuestionBank() {
  loading.value = true
  try {
    const res = await getScaleQuestionBankApi({
      scaleCode: props.scaleCode,
      scaleVersion: props.scaleVersion,
    })
    if (res.code !== 200 || !res.result) {
      messageService.error(res.message || '加载题库失败')
      return
    }
    const normalized = normalizeQuestionBank(res.result)
    questionBank.value = normalized
    if (activeDomainCode.value !== 'all' && !normalized.domains.some(item => item.scaleCode === activeDomainCode.value))
      activeDomainCode.value = 'all'
  }
  catch (error: any) {
    console.error('load question bank failed', error)
    messageService.error(error?.response?.data?.message || error?.message || '加载题库失败')
  }
  finally {
    loading.value = false
  }
}

watch(() => [props.scaleCode, props.scaleVersion], () => {
  activeDomainCode.value = 'all'
  resetSearch()
  loadQuestionBank()
})

onMounted(loadQuestionBank)
</script>

<template>
  <div class="pep3-question-bank">
    <div class="pep3-question-bank__summary">
      <div>
        <span>题目总数</span>
        <strong>{{ questionBank?.itemCount || 0 }}</strong>
      </div>
      <div>
        <span>维度数量</span>
        <strong>{{ questionBank?.domainCount || 0 }}</strong>
      </div>
      <div>
        <span>表现记录题</span>
        <strong>{{ recordFieldItemCount }}</strong>
      </div>
    </div>

    <div class="pep3-question-bank__toolbar">
      <div class="pep3-question-bank__filters">
        <a-select
          v-model:value="activeDomainCode"
          :options="domainOptions"
          class="pep3-question-bank__domain-select"
        />
        <a-input
          v-model:value="keyword"
          allow-clear
          placeholder="搜索题号、题目、指导语、操作标准、选项"
          class="pep3-question-bank__keyword"
          @press-enter="handleSearch"
        />
        <a-button type="primary" @click="handleSearch">
          <template #icon>
            <SearchOutlined />
          </template>
          搜索
        </a-button>
        <a-button @click="resetSearch">
          重置
        </a-button>
      </div>
      <a-button :loading="loading" @click="loadQuestionBank">
        <template #icon>
          <ReloadOutlined />
        </template>
        刷新
      </a-button>
    </div>

    <div class="pep3-question-bank__body">
      <aside class="pep3-question-bank__domains">
        <button
          type="button"
          class="pep3-question-bank__domain"
          :class="{ 'is-active': activeDomainCode === 'all' }"
          @click="activeDomainCode = 'all'"
        >
          <span>全部维度</span>
          <em>{{ questionBank?.itemCount || 0 }}</em>
        </button>
        <button
          v-for="domain in questionBank?.domains || []"
          :key="domain.scaleCode"
          type="button"
          class="pep3-question-bank__domain"
          :class="{ 'is-active': activeDomainCode === domain.scaleCode }"
          @click="activeDomainCode = domain.scaleCode"
        >
          <span>{{ domain.scaleName }}</span>
          <em>{{ itemCountByDomain.get(domain.scaleCode) || 0 }}</em>
        </button>
      </aside>

      <div class="pep3-question-bank__table-wrap">
        <a-table
          class="pep3-question-bank__table"
          :columns="columns"
          :data-source="filteredItems"
          :loading="loading"
          :pagination="false"
          :scroll="{ x: 1498, y: 'calc(100vh - 230px)' }"
          row-key="itemNo"
          size="small"
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'itemNo'">
              <span class="pep3-question-bank__item-no">{{ record.itemNo }}</span>
            </template>
            <template v-else-if="column.key === 'title'">
              <div class="pep3-question-bank__title-cell">
                <strong>{{ record.itemTitle }}</strong>
                <span>{{ record.domainCode }} · {{ record.domainName }}</span>
              </div>
            </template>
            <template v-else-if="column.key === 'method'">
              <a-tooltip :overlay-style="{ maxWidth: '520px', whiteSpace: 'pre-line' }">
                <template #title>
                  {{ record.method || '--' }}
                </template>
                <div class="pep3-question-bank__clamp">
                  {{ record.method || '--' }}
                </div>
              </a-tooltip>
            </template>
            <template v-else-if="column.key === 'guidance'">
              <a-tooltip :overlay-style="{ maxWidth: '520px', whiteSpace: 'pre-line' }">
                <template #title>
                  {{ record.guidance || '--' }}
                </template>
                <div class="pep3-question-bank__clamp">
                  {{ record.guidance || '--' }}
                </div>
              </a-tooltip>
            </template>
            <template v-else-if="column.key === 'scoreOptions'">
              <div class="pep3-question-bank__option-list">
                <div v-for="option in record.scoreOptions || []" :key="`${option.value}-${option.label}`" class="pep3-question-bank__score-option">
                  <a-tag color="blue">
                    {{ option.label }}
                  </a-tag>
                  <a-tooltip :title="option.description || record.standard || '--'">
                    <span>{{ option.description || record.standard || '--' }}</span>
                  </a-tooltip>
                </div>
              </div>
            </template>
            <template v-else-if="column.key === 'recordFields'">
              <div v-if="(record.recordFields || []).length" class="pep3-question-bank__record-fields">
                <a-tag v-for="field in (record.recordFields || []).slice(0, 3)" :key="field.key || field.label">
                  {{ field.label }}
                </a-tag>
                <span v-if="(record.recordFields || []).length > 3">+{{ (record.recordFields || []).length - 3 }}</span>
              </div>
              <span v-else class="pep3-question-bank__muted">无</span>
            </template>
            <template v-else-if="column.key === 'action'">
              <a-button type="link" size="small" class="pep3-question-bank__edit" @click="openEditRecord(record)">
                <template #icon>
                  <EditOutlined />
                </template>
                编辑
              </a-button>
            </template>
          </template>
        </a-table>
      </div>
    </div>

    <a-drawer
      v-model:open="editOpen"
      :push="{ distance: isMobile ? 0 : 80 }"
      :body-style="{ padding: '0', background: '#f7f7fd' }"
      :closable="false"
      :width="drawerWidth"
      :destroy-on-close="false"
      placement="right"
      class="pep3-question-bank-editor-drawer"
      @close="closeEdit"
    >
      <template #title>
        <div class="pep3-question-bank-editor__header">
          <div class="pep3-question-bank-editor__title">
            编辑题目
          </div>
          <a-button type="text" class="pep3-question-bank-editor__close close-btn" @click="closeEdit">
            <template #icon>
              <CloseOutlined class="close-icon" />
            </template>
          </a-button>
        </div>
      </template>

      <div class="pep3-question-bank-editor scrollbar">
        <a-form layout="vertical" class="pep3-question-bank-form">
          <section class="pep3-question-bank-form__card">
            <div class="pep3-question-bank-form__card-head">
              <span>基础信息</span>
            </div>
            <div class="pep3-question-bank-form__grid">
              <a-form-item label="题号">
                <a-input-number v-model:value="form.itemNo" disabled class="pep3-question-bank-form__full" />
              </a-form-item>
              <a-form-item label="维度">
                <a-select
                  v-model:value="form.domainCode"
                  disabled
                  :options="domainOptions.filter(item => item.value !== 'all')"
                  @change="syncDomainName"
                />
              </a-form-item>
            </div>

            <a-form-item label="题目" required>
              <a-input v-model:value="form.itemTitle" :maxlength="120" />
            </a-form-item>

            <a-form-item label="测试项目">
              <a-input v-model:value="form.testItem" />
            </a-form-item>

            <a-form-item label="材料">
              <a-input v-model:value="form.materials" />
            </a-form-item>

            <a-form-item label="指导语">
              <a-input v-model:value="form.guidance" :maxlength="160" />
            </a-form-item>
          </section>

          <section class="pep3-question-bank-form__card">
            <div class="pep3-question-bank-form__card-head">
              <span>施测与评分</span>
            </div>
            <a-form-item label="操作标准">
              <a-textarea v-model:value="form.method" class="pep3-question-bank-form__textarea" :auto-size="{ minRows: 3, maxRows: 5 }" />
            </a-form-item>

            <div class="pep3-question-bank-form__section">
              <div class="pep3-question-bank-form__section-head">
                <span>评分选项</span>
              </div>
              <a-form-item label="选项组合">
                <a-input v-model:value="form.scoreOptionText" disabled class="pep3-question-bank-form__score-text" />
              </a-form-item>
              <div v-for="(option, index) in form.scoreOptions" :key="index" class="pep3-question-bank-form__score-row">
                <a-input-number v-model:value="option.value" disabled :precision="0" />
                <a-input v-model:value="option.label" disabled />
                <a-input v-model:value="option.description" placeholder="请输入选项说明" />
              </div>
            </div>
          </section>

          <section class="pep3-question-bank-form__card">
            <div class="pep3-question-bank-form__section-head">
              <span>儿童表现记录</span>
            </div>

            <div v-if="form.recordFields.length" class="pep3-question-bank-form__record-list">
              <div v-for="(field, fieldIndex) in form.recordFields" :key="`${field.key}-${fieldIndex}`" class="pep3-question-bank-form__record">
                <div class="pep3-question-bank-form__record-head">
                  <span>{{ field.label || `字段 ${fieldIndex + 1}` }}</span>
                </div>
                <div class="pep3-question-bank-form__record-grid">
                  <a-input v-model:value="field.key" disabled placeholder="字段标识" />
                  <a-input v-model:value="field.label" placeholder="字段名称" />
                  <a-select v-model:value="field.fieldType" disabled :options="fieldTypeOptions" />
                  <a-input v-model:value="field.placeholder" placeholder="占位提示" />
                </div>
                <div v-if="field.fieldType === 'radio' || field.fieldType === 'checkbox_group'" class="pep3-question-bank-form__field-options">
                  <div class="pep3-question-bank-form__field-option-head">
                    <span>字段选项</span>
                  </div>
                  <div v-for="(option, optionIndex) in field.options || []" :key="optionIndex" class="pep3-question-bank-form__field-option">
                    <a-input v-model:value="option.value" disabled placeholder="选项值" />
                    <a-input v-model:value="option.label" placeholder="显示名称" />
                  </div>
                </div>
              </div>
            </div>
            <a-empty v-else :image="simpleEmptyImage" description="暂无儿童表现记录字段" class="pep3-question-bank-form__empty" />
          </section>
        </a-form>
      </div>

      <template #footer>
        <div class="pep3-question-bank-form__footer">
          <a-button @click="closeEdit">
            取消
          </a-button>
          <a-button type="primary" :loading="saving" @click="submitEdit">
            保存
          </a-button>
        </div>
      </template>
    </a-drawer>
  </div>
</template>

<style scoped lang="less">
.pep3-question-bank {
  --pep3-question-bank-panel-height: calc(100vh - 176px);

  display: flex;
  flex-direction: column;
  gap: 12px;
}

.pep3-question-bank__summary {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  overflow: hidden;
  border: 1px solid #edf0f5;
  border-radius: 8px;
  background: #fff;
}

.pep3-question-bank__summary > div {
  display: flex;
  align-items: center;
  gap: 10px;
  min-height: 42px;
  padding: 8px 14px;
  border-right: 1px solid #edf0f5;
}

.pep3-question-bank__summary > div:last-child {
  border-right: 0;
}

.pep3-question-bank__summary span {
  color: #667085;
  font-size: 13px;
}

.pep3-question-bank__summary strong {
  color: #1f2329;
  font-size: 18px;
  font-weight: 700;
}

.pep3-question-bank__toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 12px 14px;
  border: 1px solid #edf0f5;
  border-radius: 8px;
  background: #fff;
}

.pep3-question-bank__filters {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
  min-width: 0;
}

.pep3-question-bank__domain-select {
  width: 220px;
}

.pep3-question-bank__keyword {
  width: 320px;
}

.pep3-question-bank__body {
  display: grid;
  grid-template-columns: 220px minmax(0, 1fr);
  gap: 12px;
  align-items: stretch;
}

.pep3-question-bank__domains {
  position: sticky;
  top: 12px;
  display: flex;
  flex-direction: column;
  height: var(--pep3-question-bank-panel-height);
  overflow-y: auto;
  border: 1px solid #edf0f5;
  border-radius: 8px;
  background: #fff;
  scrollbar-width: thin;
  scrollbar-color: #c9d3df transparent;
}

.pep3-question-bank__domain {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  min-height: 38px;
  padding: 8px 12px;
  border: 0;
  border-bottom: 1px solid #f1f3f7;
  background: transparent;
  color: #344054;
  font-size: 13px;
  text-align: left;
  cursor: pointer;
}

.pep3-question-bank__domain:last-child {
  border-bottom: 0;
}

.pep3-question-bank__domain:hover,
.pep3-question-bank__domain.is-active {
  color: #1677ff;
  background: #f4f8ff;
}

.pep3-question-bank__domain span {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.pep3-question-bank__domain em {
  flex: 0 0 auto;
  color: #98a2b3;
  font-style: normal;
}

.pep3-question-bank__table-wrap {
  min-width: 0;
  height: var(--pep3-question-bank-panel-height);
  overflow: hidden;
  border: 1px solid #edf0f5;
  border-radius: 8px;
  background: #fff;
}

.pep3-question-bank__table {
  height: 100%;
}

.pep3-question-bank__table :deep(.ant-table-thead > tr > th) {
  background: #fafafa !important;
  color: #262626;
  font-size: 13px;
  font-weight: 500;
}

.pep3-question-bank__table :deep(.ant-table-tbody > tr > td) {
  vertical-align: middle;
}

.pep3-question-bank__table :deep(.ant-table-body) {
  scrollbar-width: thin;
  scrollbar-color: #c9d3df transparent;
}

.pep3-question-bank__domains::-webkit-scrollbar,
.pep3-question-bank__table :deep(.ant-table-body::-webkit-scrollbar) {
  width: 8px;
  height: 8px;
}

.pep3-question-bank__domains::-webkit-scrollbar-track,
.pep3-question-bank__table :deep(.ant-table-body::-webkit-scrollbar-track) {
  background: transparent;
}

.pep3-question-bank__domains::-webkit-scrollbar-thumb,
.pep3-question-bank__table :deep(.ant-table-body::-webkit-scrollbar-thumb) {
  border-radius: 999px;
  background: #c9d3df;
}

.pep3-question-bank__domains::-webkit-scrollbar-thumb:hover,
.pep3-question-bank__table :deep(.ant-table-body::-webkit-scrollbar-thumb:hover) {
  background: #aeb9c8;
}

.pep3-question-bank__item-no {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 28px;
  height: 22px;
  border-radius: 999px;
  background: #f2f5fb;
  color: #344054;
  font-size: 12px;
  font-weight: 600;
}

.pep3-question-bank__title-cell {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.pep3-question-bank__title-cell strong {
  color: #262626;
  font-size: 13px;
  font-weight: 600;
  line-height: 20px;
}

.pep3-question-bank__title-cell span,
.pep3-question-bank__muted {
  color: #8c8c8c;
  font-size: 12px;
  line-height: 18px;
}

.pep3-question-bank__clamp {
  display: -webkit-box;
  overflow: hidden;
  color: #475467;
  font-size: 12px;
  line-height: 20px;
  white-space: pre-line;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 3;
}

.pep3-question-bank__option-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.pep3-question-bank__score-option {
  display: grid;
  grid-template-columns: 44px minmax(0, 1fr);
  align-items: center;
  gap: 6px;
  min-height: 24px;
  color: #475467;
  font-size: 12px;
  line-height: 20px;
}

.pep3-question-bank__score-option :deep(.ant-tag) {
  width: 40px;
  height: 24px;
  margin-right: 0;
  padding: 0;
  line-height: 22px;
  text-align: center;
}

.pep3-question-bank__score-option span {
  display: block;
  overflow: hidden;
  min-width: 0;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.pep3-question-bank__record-fields {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 4px;
}

.pep3-question-bank__record-fields span {
  color: #8c8c8c;
  font-size: 12px;
}

.pep3-question-bank__edit {
  padding: 0;
}

.pep3-question-bank-editor {
  height: calc(100vh - 112px);
  padding: 24px;
  overflow-y: auto;
  background: #f7f7fd;
  scrollbar-width: thin;
  scrollbar-color: #c9d3df transparent;
}

.pep3-question-bank-editor::-webkit-scrollbar {
  width: 8px;
}

.pep3-question-bank-editor::-webkit-scrollbar-track {
  background: transparent;
}

.pep3-question-bank-editor::-webkit-scrollbar-thumb {
  border-radius: 999px;
  background: #c9d3df;
}

.pep3-question-bank-editor__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 20px;
}

.pep3-question-bank-editor__title {
  color: #1f2329;
  font-size: 20px;
  font-weight: 700;
  line-height: 28px;
}

.pep3-question-bank-editor__close {
  width: 32px;
  height: 32px;
  color: #667085;
}

.pep3-question-bank-editor__close .close-icon {
  font-size: 18px;
}

:deep(.pep3-question-bank-editor-drawer .ant-drawer-header) {
  height: 64px;
  padding: 22px 24px;
  border-bottom: 1px solid #edf0f5;
}

:deep(.pep3-question-bank-editor-drawer .ant-drawer-footer) {
  padding: 14px 24px;
  border-top: 1px solid #edf0f5;
  box-shadow: 0 -8px 20px rgba(15, 23, 42, 0.05);
}

.close-btn:hover {
  background: transparent;
}

.close-btn:hover .close-icon {
  animation: icon-rotate 0.3s linear;
}

@keyframes icon-rotate {
  from {
    transform: rotate(0deg);
  }

  to {
    transform: rotate(180deg);
  }
}

.pep3-question-bank-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.pep3-question-bank-form__card {
  padding: 18px 18px 8px;
  border: 1px solid #edf0f5;
  border-radius: 8px;
  background: #fff;
}

.pep3-question-bank-form__card-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 14px;
  padding-bottom: 12px;
  border-bottom: 1px solid #f1f3f7;
}

.pep3-question-bank-form__card-head span {
  color: #1f2329;
  font-size: 15px;
  font-weight: 700;
  line-height: 22px;
}

.pep3-question-bank-form :deep(.ant-form-item) {
  margin-bottom: 14px;
}

.pep3-question-bank-form :deep(.ant-form-item-label) {
  padding-bottom: 6px;
}

.pep3-question-bank-form :deep(.ant-form-item-label > label) {
  height: 22px;
  color: #344054;
  font-size: 13px;
  font-weight: 600;
}

.pep3-question-bank-form :deep(.ant-input),
.pep3-question-bank-form :deep(.ant-input-number),
.pep3-question-bank-form :deep(.ant-select-selector) {
  border-radius: 8px;
}

.pep3-question-bank-form :deep(.ant-input[disabled]),
.pep3-question-bank-form :deep(.ant-input-number-disabled),
.pep3-question-bank-form :deep(.ant-select-disabled .ant-select-selector) {
  color: #667085;
  background: #f7f9fc !important;
  border-color: #e5e7eb !important;
}

.pep3-question-bank-form__grid,
.pep3-question-bank-form__record-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0 14px;
}

.pep3-question-bank-form__full {
  width: 100%;
}

.pep3-question-bank-form__section {
  margin-top: 4px;
  padding-top: 14px;
  border-top: 1px solid #edf0f5;
}

.pep3-question-bank-form__section-head,
.pep3-question-bank-form__record-head,
.pep3-question-bank-form__field-option-head,
.pep3-question-bank-form__footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.pep3-question-bank-form__section-head {
  margin-bottom: 10px;
}

.pep3-question-bank-form__section-head span,
.pep3-question-bank-form__record-head span {
  color: #1f2329;
  font-size: 14px;
  font-weight: 600;
}

.pep3-question-bank-form__score-text {
  margin-bottom: 8px;
}

.pep3-question-bank-form__score-row {
  display: grid;
  grid-template-columns: 84px 112px minmax(0, 1fr);
  gap: 8px;
  align-items: center;
  margin-top: 8px;
}

.pep3-question-bank-form__field-option {
  display: grid;
  grid-template-columns: minmax(0, 0.7fr) minmax(0, 1fr);
  gap: 8px;
  align-items: center;
  margin-top: 8px;
}

.pep3-question-bank-form__record-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.pep3-question-bank-form__record {
  padding: 12px;
  border: 1px solid #edf0f5;
  border-radius: 8px;
  background: #fbfcfe;
}

.pep3-question-bank-form__record-head {
  margin-bottom: 10px;
}

.pep3-question-bank-form__field-options {
  margin-top: 10px;
  padding-top: 10px;
  border-top: 1px dashed #d9e1ec;
}

.pep3-question-bank-form__field-option-head {
  color: #667085;
  font-size: 12px;
}

.pep3-question-bank-form__empty {
  margin: 8px 0 14px;
  padding: 26px 16px;
  border: 1px dashed #d9e6ef;
  border-radius: 8px;
  background: #fbfefd;
}

.pep3-question-bank-form__empty :deep(.ant-empty-description) {
  color: #667085;
  font-size: 13px;
}

.pep3-question-bank-form__footer {
  justify-content: flex-end;
}

.pep3-question-bank-form__footer :deep(.ant-btn) {
  min-width: 96px;
  height: 40px;
  border-radius: 8px;
}

@media (max-width: 1100px) {
  .pep3-question-bank__body {
    grid-template-columns: 1fr;
  }

  .pep3-question-bank__domains {
    position: static;
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    height: auto;
  }

  .pep3-question-bank__table-wrap {
    height: auto;
  }
}

@media (max-width: 760px) {
  .pep3-question-bank__summary,
  .pep3-question-bank-form__grid,
  .pep3-question-bank-form__record-grid,
  .pep3-question-bank-form__score-row,
  .pep3-question-bank-form__field-option {
    grid-template-columns: 1fr;
  }

  .pep3-question-bank__summary > div {
    border-right: 0;
    border-bottom: 1px solid #edf0f5;
  }

  .pep3-question-bank__toolbar,
  .pep3-question-bank__filters {
    align-items: stretch;
  }

  .pep3-question-bank__toolbar {
    flex-direction: column;
  }

  .pep3-question-bank__domain-select,
  .pep3-question-bank__keyword {
    width: 100%;
  }
}
</style>
