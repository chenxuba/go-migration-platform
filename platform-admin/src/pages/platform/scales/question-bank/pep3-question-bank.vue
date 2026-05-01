<script setup lang="ts">
import type { TableColumnsType } from 'ant-design-vue'
import { EditOutlined, PlusOutlined, ReloadOutlined, SearchOutlined } from '@ant-design/icons-vue'
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

const columns: TableColumnsType<ScaleQuestionBankItem> = [
  { title: '题号', key: 'itemNo', width: 72, fixed: 'left' as const, align: 'center' as const },
  { title: '题目', key: 'title', width: 240 },
  { title: '操作标准', key: 'standard', width: 300 },
  { title: '指导语', key: 'method', width: 260 },
  { title: '选项', key: 'scoreOptions', width: 210 },
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

const form = reactive<ScaleQuestionBankItem & { scaleCode: string, scaleVersion: string }>({
  scaleCode: '',
  scaleVersion: '',
  itemNo: 0,
  itemTitle: '',
  testItem: '',
  materials: '',
  method: '',
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
      item.method,
      item.standard,
    ].join(' ').toLowerCase().includes(key)
  })
})

function recordFieldTypeLabel(type: string) {
  return fieldTypeOptions.find(item => item.value === type)?.label || type || '--'
}

function normalizeOptions(options?: ScaleQuestionBankRecordFieldOption[] | null) {
  return (Array.isArray(options) ? options : []).map(item => ({
    value: item.value || item.label || '',
    label: item.label || item.value || '',
  })).filter(item => item.value || item.label)
}

function cloneRecordFields(fields?: ScaleQuestionBankRecordField[] | null) {
  return (Array.isArray(fields) ? fields : []).map(field => ({
    key: field.key || '',
    label: field.label || '',
    fieldType: field.fieldType || 'text',
    displayType: field.displayType || recordFieldTypeLabel(field.fieldType || 'text'),
    required: Boolean(field.required),
    placeholder: field.placeholder || '',
    options: normalizeOptions(field.options),
  }))
}

function cloneScoreOptions(options?: ScaleQuestionBankScoreOption[] | null) {
  return (Array.isArray(options) ? options : []).map(option => ({
    value: Number(option.value),
    label: option.label || `${option.value}分`,
    description: option.description || '',
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
      itemTitle: item.itemTitle || '',
      testItem: item.testItem || '',
      materials: item.materials || '',
      method: item.method || '',
      domainCode: item.domainCode || '',
      domainName: item.domainName || '',
      standard: item.standard || '',
      scoreOptions,
      scoreOptionText: item.scoreOptionText || scoreOptions.map(option => option.value).join('/'),
      recordFields: cloneRecordFields(item.recordFields),
      sourcePdf: item.sourcePdf || '',
      sourcePages: normalizeNumberArray(item.sourcePages),
      ocrStatus: item.ocrStatus || '',
      updatedAt: item.updatedAt || '',
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
  form.itemTitle = record.itemTitle
  form.testItem = record.testItem
  form.materials = record.materials || ''
  form.method = record.method || ''
  form.domainCode = record.domainCode
  form.domainName = record.domainName
  form.standard = record.standard || ''
  form.scoreOptions = cloneScoreOptions(record.scoreOptions)
  form.scoreOptionText = record.scoreOptionText || form.scoreOptions.map(item => item.value).join('/')
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

function addScoreOption() {
  form.scoreOptions.push({
    value: 0,
    label: '0分',
    description: '',
  })
  form.scoreOptionText = form.scoreOptions.map(item => item.value).join('/')
}

function removeScoreOption(index: number) {
  form.scoreOptions.splice(index, 1)
  form.scoreOptionText = form.scoreOptions.map(item => item.value).join('/')
}

function addRecordField() {
  const nextIndex = form.recordFields.length + 1
  form.recordFields.push({
    key: `record_${nextIndex}`,
    label: '',
    fieldType: 'text',
    displayType: '填空',
    placeholder: '',
    options: [],
  })
}

function removeRecordField(index: number) {
  form.recordFields.splice(index, 1)
}

function addFieldOption(field: ScaleQuestionBankRecordField) {
  if (!field.options)
    field.options = []
  field.options.push({ value: '', label: '' })
}

function removeFieldOption(field: ScaleQuestionBankRecordField, index: number) {
  field.options?.splice(index, 1)
}

function syncDomainName() {
  form.domainName = domainNameMap.value.get(form.domainCode) || form.domainName
}

function buildPayload() {
  return {
    scaleCode: form.scaleCode,
    scaleVersion: form.scaleVersion,
    itemNo: form.itemNo,
    itemTitle: form.itemTitle.trim(),
    testItem: form.testItem.trim(),
    materials: form.materials.trim(),
    method: form.method.trim(),
    domainCode: form.domainCode.trim(),
    domainName: (domainNameMap.value.get(form.domainCode) || form.domainName || '').trim(),
    standard: form.standard.trim(),
    scoreOptions: form.scoreOptions.map(item => ({
      value: Number(item.value),
      label: item.label || `${item.value}分`,
      description: item.description || '',
    })),
    scoreOptionText: form.scoreOptionText.trim(),
    recordFields: form.recordFields.map(field => ({
      key: field.key.trim(),
      label: field.label.trim(),
      fieldType: field.fieldType,
      displayType: field.displayType || recordFieldTypeLabel(field.fieldType),
      required: Boolean(field.required),
      placeholder: field.placeholder || '',
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
          placeholder="搜索题号、题目、指导语、操作标准"
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
          :scroll="{ x: 1388, y: 'calc(100vh - 312px)' }"
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
            <template v-else-if="column.key === 'standard'">
              <a-tooltip :overlay-style="{ maxWidth: '520px', whiteSpace: 'pre-line' }">
                <template #title>
                  {{ record.standard || '--' }}
                </template>
                <div class="pep3-question-bank__clamp">
                  {{ record.standard || '--' }}
                </div>
              </a-tooltip>
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
            <template v-else-if="column.key === 'scoreOptions'">
              <div class="pep3-question-bank__option-list">
                <a-tag v-for="option in record.scoreOptions || []" :key="`${option.value}-${option.label}`" color="blue">
                  {{ option.label }}
                </a-tag>
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
      width="760"
      :destroy-on-close="false"
      title="编辑题目"
      class="pep3-question-bank-drawer"
      @close="closeEdit"
    >
      <a-form layout="vertical" class="pep3-question-bank-form">
        <div class="pep3-question-bank-form__grid">
          <a-form-item label="题号">
            <a-input-number v-model:value="form.itemNo" disabled class="pep3-question-bank-form__full" />
          </a-form-item>
          <a-form-item label="维度">
            <a-select v-model:value="form.domainCode" :options="domainOptions.filter(item => item.value !== 'all')" @change="syncDomainName" />
          </a-form-item>
        </div>

        <a-form-item label="题目" required>
          <a-input v-model:value="form.itemTitle" :maxlength="120" />
        </a-form-item>

        <a-form-item label="测试项目">
          <a-textarea v-model:value="form.testItem" :auto-size="{ minRows: 2, maxRows: 4 }" />
        </a-form-item>

        <a-form-item label="指导语">
          <a-textarea v-model:value="form.method" :auto-size="{ minRows: 3, maxRows: 6 }" />
        </a-form-item>

        <a-form-item label="材料">
          <a-textarea v-model:value="form.materials" :auto-size="{ minRows: 2, maxRows: 4 }" />
        </a-form-item>

        <a-form-item label="操作标准">
          <a-textarea v-model:value="form.standard" :auto-size="{ minRows: 5, maxRows: 10 }" />
        </a-form-item>

        <div class="pep3-question-bank-form__section">
          <div class="pep3-question-bank-form__section-head">
            <span>选项</span>
            <a-button size="small" @click="addScoreOption">
              <template #icon>
                <PlusOutlined />
              </template>
              新增选项
            </a-button>
          </div>
          <a-input v-model:value="form.scoreOptionText" placeholder="例如 2/1/0" class="pep3-question-bank-form__score-text" />
          <div v-for="(option, index) in form.scoreOptions" :key="index" class="pep3-question-bank-form__score-row">
            <a-input-number v-model:value="option.value" :precision="0" />
            <a-input v-model:value="option.label" placeholder="标签" />
            <a-input v-model:value="option.description" placeholder="说明" />
            <a-button danger type="link" @click="removeScoreOption(index)">
              删除
            </a-button>
          </div>
        </div>

        <div class="pep3-question-bank-form__section">
          <div class="pep3-question-bank-form__section-head">
            <span>儿童表现记录</span>
            <a-button size="small" @click="addRecordField">
              <template #icon>
                <PlusOutlined />
              </template>
              新增字段
            </a-button>
          </div>

          <div v-if="form.recordFields.length" class="pep3-question-bank-form__record-list">
            <div v-for="(field, fieldIndex) in form.recordFields" :key="`${field.key}-${fieldIndex}`" class="pep3-question-bank-form__record">
              <div class="pep3-question-bank-form__record-head">
                <span>{{ field.label || `字段 ${fieldIndex + 1}` }}</span>
                <a-button danger type="link" @click="removeRecordField(fieldIndex)">
                  删除
                </a-button>
              </div>
              <div class="pep3-question-bank-form__record-grid">
                <a-input v-model:value="field.key" placeholder="字段标识" />
                <a-input v-model:value="field.label" placeholder="字段名称" />
                <a-select v-model:value="field.fieldType" :options="fieldTypeOptions" />
                <a-input v-model:value="field.placeholder" placeholder="占位提示" />
              </div>
              <div v-if="field.fieldType === 'radio' || field.fieldType === 'checkbox_group'" class="pep3-question-bank-form__field-options">
                <div class="pep3-question-bank-form__field-option-head">
                  <span>字段选项</span>
                  <a-button size="small" @click="addFieldOption(field)">
                    新增
                  </a-button>
                </div>
                <div v-for="(option, optionIndex) in field.options || []" :key="optionIndex" class="pep3-question-bank-form__field-option">
                  <a-input v-model:value="option.value" placeholder="选项值" />
                  <a-input v-model:value="option.label" placeholder="显示名称" />
                  <a-button danger type="link" @click="removeFieldOption(field, optionIndex)">
                    删除
                  </a-button>
                </div>
              </div>
            </div>
          </div>
          <a-empty v-else description="暂无儿童表现记录字段" />
        </div>
      </a-form>

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
  align-items: start;
}

.pep3-question-bank__domains {
  position: sticky;
  top: 12px;
  display: flex;
  flex-direction: column;
  max-height: calc(100vh - 236px);
  overflow-y: auto;
  border: 1px solid #edf0f5;
  border-radius: 8px;
  background: #fff;
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
  border: 1px solid #edf0f5;
  border-radius: 8px;
  background: #fff;
}

.pep3-question-bank__table :deep(.ant-table-thead > tr > th) {
  background: #fafafa !important;
  color: #262626;
  font-size: 13px;
  font-weight: 500;
}

.pep3-question-bank__table :deep(.ant-table-tbody > tr > td) {
  vertical-align: top;
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

.pep3-question-bank__option-list,
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

.pep3-question-bank-form__grid,
.pep3-question-bank-form__record-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0 12px;
}

.pep3-question-bank-form__full {
  width: 100%;
}

.pep3-question-bank-form__section {
  margin-top: 18px;
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

.pep3-question-bank-form__score-row,
.pep3-question-bank-form__field-option {
  display: grid;
  grid-template-columns: 84px minmax(0, 0.8fr) minmax(0, 1fr) 52px;
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

.pep3-question-bank-form__field-option {
  grid-template-columns: minmax(0, 1fr) minmax(0, 1fr) 52px;
}

.pep3-question-bank-form__footer {
  justify-content: flex-end;
}

@media (max-width: 1100px) {
  .pep3-question-bank__body {
    grid-template-columns: 1fr;
  }

  .pep3-question-bank__domains {
    position: static;
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    max-height: none;
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
