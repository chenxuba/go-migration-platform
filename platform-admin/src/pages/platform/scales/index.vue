<script setup lang="ts">
import type { UploadRequestOption } from 'ant-design-vue/es/vc-upload/interface'
import type { TableColumnsType } from 'ant-design-vue'
import {
  DeleteOutlined,
  DownOutlined,
  EditOutlined,
  FileTextOutlined,
  HeartOutlined,
  PlusOutlined,
  ReloadOutlined,
  SearchOutlined,
  UploadOutlined,
} from '@ant-design/icons-vue'
import { computed, onMounted, reactive, ref } from 'vue'
import * as qiniu from 'qiniu-js'
import { useRouter } from 'vue-router'
import { listDictValuesApi } from '@/api/platform/dicts'
import {
  createScaleApi,
  createScaleAcknowledgementApi,
  createScaleReferenceApi,
  deleteScaleAcknowledgementApi,
  deleteScaleReferenceApi,
  listScalesApi,
  type ScaleMutationPayload,
  type ScaleRecord,
  type ScaleTextResourceItem,
  updateScaleApi,
  updateScaleAcknowledgementApi,
  updateScaleReferenceApi,
} from '@/api/platform/scales'
import { getQiniuToken } from '@/api/qiniu'
import { resolveUploadErrorMessage, validateUploadFileByToken } from '@/utils/upload-limit'
import messageService from '@/utils/messageService'
import PlatformModalShell from '../shared/platform-modal-shell.vue'
import { PlatformAccessEnum } from '~@/constants/access'

type ResourceKind = 'references' | 'acknowledgements'
type LooseScaleRecord = ScaleRecord | Record<string, any>
type ScaleFormMode = 'create' | 'edit'

const { hasAccess } = useAccess()
const router = useRouter()

const keyword = ref('')
const appliedKeyword = ref('')
const categoryFilter = ref('')
const scenarioFilter = ref('')
const authManageOpen = ref(false)
const activeAuthScale = ref<ScaleRecord | null>(null)
const scaleLoading = ref(false)
const scaleFormOpen = ref(false)
const scaleFormMode = ref<ScaleFormMode>('create')
const scaleSaving = ref(false)
const posterUploading = ref(false)
const posterUploadProgress = ref(0)
const referenceManageOpen = ref(false)
const thanksManageOpen = ref(false)
const activeResourceScale = ref<ScaleRecord | null>(null)
const resourceFormKind = ref<ResourceKind | null>(null)
const resourceFormId = ref<number | null>(null)
const resourceFormIndex = ref<number | null>(null)
const resourceFormContent = ref('')
const resourceSaving = ref(false)
const categoryOptions = ref([
  { label: '全部分类', value: '' },
  { label: '标准化测评', value: '标准化测评' },
])
const scenarioOptions = ref([
  { label: '全部场景', value: '' },
  { label: '现场测评', value: '现场测评' },
])

const scaleRecords = ref<ScaleRecord[]>([])

const columns: TableColumnsType<ScaleRecord> = [
  { title: '量表信息', key: 'scale', width: 300, fixed: 'left' as const },
  { title: '分类 / 场景', key: 'meta', width: 170 },
  { title: '当前版本', key: 'version', width: 150 },
  { title: '题库', key: 'data', width: 150 },
  { title: '授权机构', key: 'auth', width: 150 },
  { title: '最近更新', key: 'updatedAt', width: 160 },
  { title: '操作', key: 'action', width: 190, fixed: 'right' as const },
]

const scaleForm = reactive<ScaleMutationPayload>({
  name: '',
  code: '',
  category: '',
  scenario: '',
  ageRange: '',
  ageMinMonths: 0,
  ageMaxMonths: 0,
  currentVersion: '',
  itemCount: 0,
  domainCount: 0,
  summary: '',
  posterUrl: '',
  executionEntry: '',
  apiPackage: '',
})

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

const scaleFormTitle = computed(() => scaleFormMode.value === 'create' ? '新增量表' : '编辑量表')
const categoryFormOptions = computed(() => categoryOptions.value.filter(item => item.value))
const scenarioFormOptions = computed(() => scenarioOptions.value.filter(item => item.value))

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

function openAuthManage(record: LooseScaleRecord) {
  activeAuthScale.value = asScaleRecord(record)
  authManageOpen.value = true
}

function resetScaleForm() {
  scaleForm.id = undefined
  scaleForm.name = ''
  scaleForm.code = ''
  scaleForm.category = categoryFormOptions.value[0]?.value || ''
  scaleForm.scenario = scenarioFormOptions.value[0]?.value || ''
  scaleForm.ageRange = ''
  scaleForm.ageMinMonths = 0
  scaleForm.ageMaxMonths = 0
  scaleForm.currentVersion = ''
  scaleForm.itemCount = 0
  scaleForm.domainCount = 0
  scaleForm.summary = ''
  scaleForm.posterUrl = ''
  scaleForm.executionEntry = ''
  scaleForm.apiPackage = ''
  posterUploadProgress.value = 0
}

function openCreateScale() {
  scaleFormMode.value = 'create'
  resetScaleForm()
  scaleFormOpen.value = true
}

function openEditScale(record: LooseScaleRecord) {
  const scale = asScaleRecord(record)
  scaleFormMode.value = 'edit'
  scaleForm.id = scale.id
  scaleForm.name = scale.name || ''
  scaleForm.code = scale.code || ''
  scaleForm.category = scale.category || ''
  scaleForm.scenario = scale.scenario || ''
  scaleForm.ageRange = scale.ageRange || ''
  scaleForm.ageMinMonths = Number(scale.ageMinMonths || 0)
  scaleForm.ageMaxMonths = Number(scale.ageMaxMonths || 0)
  scaleForm.currentVersion = scale.currentVersion || ''
  scaleForm.itemCount = Number(scale.itemCount || 0)
  scaleForm.domainCount = Number(scale.domainCount || 0)
  scaleForm.summary = scale.summary || ''
  scaleForm.posterUrl = scale.posterUrl || ''
  scaleForm.executionEntry = scale.executionEntry || ''
  scaleForm.apiPackage = scale.apiPackage || ''
  posterUploadProgress.value = 0
  scaleFormOpen.value = true
}

function closeScaleForm() {
  scaleFormOpen.value = false
}

function validateScaleForm() {
  if (!scaleForm.name.trim())
    return '请输入量表名称'
  if (scaleFormMode.value === 'create' && !String(scaleForm.code || '').trim())
    return '请输入量表编码'
  if (!scaleForm.category)
    return '请选择量表分类'
  if (!scaleForm.scenario)
    return '请选择使用场景'
  if (Number(scaleForm.ageMinMonths) < 0 || Number(scaleForm.ageMaxMonths) < 0)
    return '适用年龄不能小于0'
  if (Number(scaleForm.ageMaxMonths) > 0 && Number(scaleForm.ageMinMonths) > Number(scaleForm.ageMaxMonths))
    return '最小月龄不能大于最大月龄'
  if (!scaleForm.currentVersion.trim())
    return '请输入当前版本'
  if (Number(scaleForm.itemCount) < 0)
    return '题库数量不能小于0'
  if (Number(scaleForm.domainCount) < 0)
    return '维度数量不能小于0'
  return ''
}

function buildScalePayload() {
  const ageMinMonths = Number(scaleForm.ageMinMonths || 0)
  const ageMaxMonths = Number(scaleForm.ageMaxMonths || 0)
  return {
    id: scaleForm.id,
    name: scaleForm.name.trim(),
    code: String(scaleForm.code || '').trim(),
    category: scaleForm.category,
    scenario: scaleForm.scenario,
    ageRange: formatAgeRange(ageMinMonths, ageMaxMonths),
    ageMinMonths,
    ageMaxMonths,
    currentVersion: scaleForm.currentVersion.trim(),
    itemCount: Number(scaleForm.itemCount || 0),
    domainCount: Number(scaleForm.domainCount || 0),
    summary: String(scaleForm.summary || '').trim(),
    posterUrl: String(scaleForm.posterUrl || '').trim(),
    executionEntry: String(scaleForm.executionEntry || '').trim(),
    apiPackage: String(scaleForm.apiPackage || '').trim(),
  }
}

function beforePosterUpload(file: File) {
  if (!file.type?.startsWith('image/')) {
    messageService.warning('宣传海报只能上传图片文件')
    return false
  }
  if (file.size / 1024 / 1024 > 8) {
    messageService.warning('宣传海报大小不能超过 8MB')
    return false
  }
  return true
}

async function handlePosterUpload(options: UploadRequestOption) {
  const rawFile = options.file as File
  if (!rawFile || !beforePosterUpload(rawFile)) {
    options.onError?.(new Error('invalid file'))
    return
  }

  posterUploading.value = true
  posterUploadProgress.value = 0
  try {
    const tokenRes: any = await getQiniuToken()
    const { token, uuid, buckethostname } = tokenRes.result || {}
    if (!token || !uuid || !buckethostname)
      throw new Error(tokenRes?.message || '获取上传凭证失败')
    validateUploadFileByToken(rawFile, tokenRes.result, '宣传海报')

    const ext = rawFile.name.includes('.') ? rawFile.name.slice(rawFile.name.lastIndexOf('.')) : '.png'
    const key = `scale/poster/${uuid}${ext}`
    const observable = qiniu.upload(rawFile, key, token, {
      fname: rawFile.name,
      mimeType: rawFile.type,
    }, {
      useCdnDomain: true,
      region: qiniu.region.z0,
    })

    observable.subscribe({
      next(result) {
        posterUploadProgress.value = Math.floor(result.total.percent)
      },
      error(error) {
        console.error('upload scale poster failed', error)
        messageService.error(resolveUploadErrorMessage(error, '宣传海报上传失败'))
        posterUploading.value = false
        posterUploadProgress.value = 0
        options.onError?.(error)
      },
      complete(result) {
        scaleForm.posterUrl = `${buckethostname}${result.key}`
        posterUploading.value = false
        posterUploadProgress.value = 100
        messageService.success('宣传海报上传成功')
        options.onSuccess?.(result as any)
      },
    })
  }
  catch (error: any) {
    console.error('prepare scale poster upload failed', error)
    messageService.error(resolveUploadErrorMessage(error, '宣传海报上传失败'))
    posterUploading.value = false
    posterUploadProgress.value = 0
    options.onError?.(error)
  }
}

function clearPosterUrl() {
  scaleForm.posterUrl = ''
  posterUploadProgress.value = 0
}

function formatAgeRange(minMonths: number, maxMonths: number) {
  if (minMonths <= 0 && maxMonths <= 0)
    return ''
  if (maxMonths > 0 && minMonths > maxMonths)
    [minMonths, maxMonths] = [maxMonths, minMonths]
  if (maxMonths <= 0 || minMonths === maxMonths)
    return formatAgeLabel(minMonths)
  if (minMonths <= 0)
    return `${formatAgeLabel(maxMonths)}以下`
  return `${formatAgeLabel(minMonths)}-${formatAgeLabel(maxMonths)}`
}

function formatAgeLabel(months: number) {
  if (months <= 0)
    return '0岁'
  const years = Math.floor(months / 12)
  const remainMonths = months % 12
  if (!remainMonths)
    return `${years}岁`
  if (!years)
    return `${remainMonths}个月`
  return `${years}.${remainMonths}岁`
}

async function submitScaleForm() {
  const warning = validateScaleForm()
  if (warning) {
    messageService.warning(warning)
    return
  }

  scaleSaving.value = true
  try {
    const payload = buildScalePayload()
    const res = scaleFormMode.value === 'create'
      ? await createScaleApi(payload)
      : await updateScaleApi({ ...payload, id: Number(payload.id || 0) })
    if (res.code !== 200) {
      messageService.error(res.message || '量表保存失败')
      return
    }
    messageService.success(scaleFormMode.value === 'create' ? '量表已新增' : '量表已保存')
    scaleFormOpen.value = false
    await loadScaleRecords()
  }
  catch (error: any) {
    console.error('save scale failed', error)
    messageService.error(getErrorMessage(error, '量表保存失败'))
  }
  finally {
    scaleSaving.value = false
  }
}

function openQuestionBankPage(record: LooseScaleRecord) {
  const scale = asScaleRecord(record)
  router.push({
    name: 'PlatformScaleQuestionBank',
    query: {
      scaleCode: scale.code,
      scaleVersion: scale.currentVersion,
      scaleId: String(scale.id),
    },
  })
}

function isPEP3Scale(record: LooseScaleRecord) {
  return String(asScaleRecord(record).code || '').trim().toUpperCase() === 'PEP3'
}

function openPEP3IEPMaterialPage(record: LooseScaleRecord) {
  const scale = asScaleRecord(record)
  router.push({
    name: 'PlatformPEP3IEPMaterials',
    query: {
      scaleCode: scale.code,
      scaleVersion: scale.currentVersion,
      scaleId: String(scale.id),
    },
  })
}

function handleSearch() {
  appliedKeyword.value = keyword.value.trim()
}

function resetResourceForm() {
  resourceFormKind.value = null
  resourceFormId.value = null
  resourceFormIndex.value = null
  resourceFormContent.value = ''
}

function openReferenceManage(record: LooseScaleRecord) {
  const scale = asScaleRecord(record)
  activeResourceScale.value = scale
  thanksManageOpen.value = false
  referenceManageOpen.value = true
  resetResourceForm()
}

function openThanksManage(record: LooseScaleRecord) {
  const scale = asScaleRecord(record)
  activeResourceScale.value = scale
  referenceManageOpen.value = false
  thanksManageOpen.value = true
  resetResourceForm()
}

function resourceKindLabel(kind: ResourceKind) {
  return kind === 'references' ? '引用文献' : '特别鸣谢'
}

function getResourceList(kind: ResourceKind) {
  if (!activeResourceScale.value)
    return [] as ScaleTextResourceItem[]
  return kind === 'references'
    ? activeResourceScale.value.references
    : activeResourceScale.value.acknowledgements
}

function startCreateResource(kind: ResourceKind) {
  resourceFormKind.value = kind
  resourceFormId.value = null
  resourceFormIndex.value = null
  resourceFormContent.value = ''
}

function startEditResource(kind: ResourceKind, index: number) {
  const item = getResourceList(kind)[index]
  if (!item)
    return
  resourceFormKind.value = kind
  resourceFormId.value = item.id
  resourceFormIndex.value = index
  resourceFormContent.value = item.content || ''
}

function isEditingResource(kind: ResourceKind, index: number) {
  return resourceFormKind.value === kind && resourceFormIndex.value === index
}

function formatResourceIndex(index: number) {
  return String(index + 1).padStart(2, '0')
}

function getErrorMessage(error: any, fallback: string) {
  return error?.response?.data?.message || error?.message || fallback
}

async function submitResourceForm(kind: ResourceKind) {
  if (!activeResourceScale.value?.id) {
    messageService.warning('请先选择量表')
    return
  }

  const content = resourceFormContent.value.trim()
  if (!content) {
    messageService.warning(`请输入${resourceKindLabel(kind)}内容`)
    return
  }

  const list = getResourceList(kind)
  const editIndex = resourceFormIndex.value
  const editItem = editIndex !== null && editIndex >= 0 ? list[editIndex] : undefined
  const payload = {
    scaleId: activeResourceScale.value.id,
    content,
    sort: editItem?.sort || list.length + 1,
  }

  resourceSaving.value = true
  try {
    let res
    if (kind === 'references') {
      res = resourceFormId.value
        ? await updateScaleReferenceApi({ ...payload, id: resourceFormId.value })
        : await createScaleReferenceApi(payload)
    }
    else {
      res = resourceFormId.value
        ? await updateScaleAcknowledgementApi({ ...payload, id: resourceFormId.value })
        : await createScaleAcknowledgementApi(payload)
    }

    if (res.code !== 200) {
      messageService.error(res.message || `${resourceKindLabel(kind)}保存失败`)
      return
    }
    messageService.success(`${resourceKindLabel(kind)}已保存`)
    resetResourceForm()
    await loadScaleRecords()
  }
  catch (error: any) {
    console.error('save scale resource failed', error)
    messageService.error(getErrorMessage(error, `${resourceKindLabel(kind)}保存失败`))
  }
  finally {
    resourceSaving.value = false
  }
}

async function removeResource(kind: ResourceKind, index: number) {
  const item = getResourceList(kind)[index]
  if (!item?.id)
    return

  resourceSaving.value = true
  try {
    const res = kind === 'references'
      ? await deleteScaleReferenceApi({ id: item.id })
      : await deleteScaleAcknowledgementApi({ id: item.id })
    if (res.code !== 200) {
      messageService.error(res.message || `${resourceKindLabel(kind)}删除失败`)
      return
    }
    if (isEditingResource(kind, index))
      resetResourceForm()
    messageService.success(`${resourceKindLabel(kind)}已删除`)
    await loadScaleRecords()
  }
  catch (error: any) {
    console.error('delete scale resource failed', error)
    messageService.error(getErrorMessage(error, `${resourceKindLabel(kind)}删除失败`))
  }
  finally {
    resourceSaving.value = false
  }
}

async function loadDictOptions() {
  try {
    const [categoryRes, scenarioRes] = await Promise.all([
      listDictValuesApi({ code: 'scale_category' }),
      listDictValuesApi({ code: 'scale_usage_scenario' }),
    ])

    if (categoryRes.code === 200 && Array.isArray(categoryRes.result)) {
      const options = categoryRes.result
        .filter(item => item.isEnable)
        .map(item => ({ label: item.dictLabel, value: item.dictValue }))
      categoryOptions.value = [{ label: '全部分类', value: '' }, ...options]
    }

    if (scenarioRes.code === 200 && Array.isArray(scenarioRes.result)) {
      const options = scenarioRes.result
        .filter(item => item.isEnable)
        .map(item => ({ label: item.dictLabel, value: item.dictValue }))
      scenarioOptions.value = [{ label: '全部场景', value: '' }, ...options]
    }
  }
  catch (error) {
    console.error('load scale dict options failed', error)
  }
}

async function loadScaleRecords() {
  scaleLoading.value = true
  try {
    const res = await listScalesApi()
    if (res.code !== 200 || !Array.isArray(res.result)) {
      messageService.error(res.message || '加载量表列表失败')
      return
    }
    scaleRecords.value = res.result
    if (activeResourceScale.value) {
      activeResourceScale.value = res.result.find(item => item.id === activeResourceScale.value?.id) || activeResourceScale.value
    }
    if (activeAuthScale.value) {
      activeAuthScale.value = res.result.find(item => item.id === activeAuthScale.value?.id) || activeAuthScale.value
    }
  }
  catch (error) {
    console.error('load scale records failed', error)
    messageService.error('加载量表列表失败')
  }
  finally {
    scaleLoading.value = false
  }
}

function resetFilters() {
  keyword.value = ''
  appliedKeyword.value = ''
  categoryFilter.value = ''
  scenarioFilter.value = ''
}

onMounted(() => {
  loadDictOptions()
  loadScaleRecords()
})
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
        <a-button v-if="hasAccess(PlatformAccessEnum.scaleManageAdd)" type="primary" @click="openCreateScale">
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
        :loading="scaleLoading"
        :pagination="false"
        :scroll="{ x: 1270 }"
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
                  <button
                    v-if="hasAccess(PlatformAccessEnum.scaleManageQuestionBank)"
                    type="button"
                    class="scale-cell__name scale-cell__name-button"
                    @click="openQuestionBankPage(record)"
                  >
                    {{ record.name }}
                  </button>
                  <div v-else class="scale-cell__name">
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
            <div class="metric-cell">
              <div class="metric-cell__value">
                {{ record.itemCount }}题
              </div>
              <div class="metric-cell__label">
                {{ record.domainCount }}个维度
              </div>
            </div>
          </template>

          <template v-else-if="column.key === 'auth'">
            <div class="metric-cell">
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
              <a v-if="hasAccess(PlatformAccessEnum.scaleManageEdit)" class="scale-actions__link" @click="openEditScale(record)">
                编辑
              </a>

              <a v-if="hasAccess(PlatformAccessEnum.scaleManageAuth)" class="scale-actions__link" @click="openAuthManage(record)">
                授权机构
              </a>

              <a-dropdown
                v-if="(isPEP3Scale(record) && hasAccess(PlatformAccessEnum.scaleManageIepTarget)) || hasAccess([
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
                    <a-menu-item v-if="isPEP3Scale(record) && hasAccess(PlatformAccessEnum.scaleManageIepTarget)" key="pep3IepMaterials" @click="openPEP3IEPMaterialPage(record)">
                      IEP素材库
                    </a-menu-item>
                    <a-menu-item v-if="hasAccess(PlatformAccessEnum.scaleManageReference)" key="references" @click="openReferenceManage(record)">
                      引用文献
                    </a-menu-item>
                    <a-menu-item v-if="hasAccess(PlatformAccessEnum.scaleManageThanks)" key="acknowledgements" @click="openThanksManage(record)">
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
      v-model:open="scaleFormOpen"
      :title="scaleFormTitle"
      :width="760"
      :scrollable="true"
      modal-class="scale-form-modal"
      @close="closeScaleForm"
    >
      <a-form :model="scaleForm" layout="vertical" class="scale-form">
        <div class="scale-form__grid">
          <a-form-item label="量表名称" required>
            <a-input v-model:value="scaleForm.name" :maxlength="80" placeholder="请输入量表名称" />
          </a-form-item>

          <a-form-item label="量表编码" required>
            <a-input
              v-model:value="scaleForm.code"
              :disabled="scaleFormMode === 'edit'"
              :maxlength="40"
              placeholder="例如 PEP3"
            />
          </a-form-item>

          <a-form-item label="量表分类" required>
            <a-select v-model:value="scaleForm.category" :options="categoryFormOptions" placeholder="请选择量表分类" />
          </a-form-item>

          <a-form-item label="使用场景" required>
            <a-select v-model:value="scaleForm.scenario" :options="scenarioFormOptions" placeholder="请选择使用场景" />
          </a-form-item>

          <a-form-item label="最小月龄 - 最大月龄">
            <div class="scale-form__range">
              <a-input-number
                v-model:value="scaleForm.ageMinMonths"
                :min="0"
                :precision="0"
                class="scale-form__number"
                placeholder="例如 30"
              />
              <span class="scale-form__range-separator">-</span>
              <a-input-number
                v-model:value="scaleForm.ageMaxMonths"
                :min="0"
                :precision="0"
                class="scale-form__number"
                placeholder="例如 72"
              />
            </div>
          </a-form-item>

          <a-form-item label="当前版本" required>
            <a-input v-model:value="scaleForm.currentVersion" :maxlength="60" placeholder="例如 2025-92题版" />
          </a-form-item>

          <a-form-item label="题库数量">
            <a-input-number v-model:value="scaleForm.itemCount" :min="0" :precision="0" class="scale-form__number" />
          </a-form-item>

          <a-form-item label="维度数量">
            <a-input-number v-model:value="scaleForm.domainCount" :min="0" :precision="0" class="scale-form__number" />
          </a-form-item>
        </div>

        <a-form-item label="量表说明">
          <a-input v-model:value="scaleForm.summary" :maxlength="120" placeholder="用于量表详情展示的简短说明" />
        </a-form-item>

        <a-form-item label="宣传海报">
          <div class="scale-poster-field">
            <a-upload
              :custom-request="handlePosterUpload"
              :show-upload-list="false"
              accept="image/*"
              :disabled="posterUploading"
            >
              <div class="scale-poster-card">
                <img v-if="scaleForm.posterUrl" :src="scaleForm.posterUrl" alt="宣传海报">
                <div v-else class="scale-poster-card__empty">
                  <UploadOutlined />
                  <span>上传海报</span>
                  <em>点击上传或替换图片</em>
                </div>
                <div v-if="posterUploading" class="scale-poster-card__mask">
                  上传中 {{ posterUploadProgress }}%
                </div>
              </div>
            </a-upload>

            <div class="scale-poster-field__hint-row">
              <span>建议使用竖版图片，大小不超过 8MB。</span>
              <a-button v-if="scaleForm.posterUrl" size="small" :disabled="posterUploading" @click="clearPosterUrl">
                <template #icon>
                  <DeleteOutlined />
                </template>
                清除
              </a-button>
            </div>
          </div>
        </a-form-item>
      </a-form>

      <template #footer>
        <div class="scale-modal-footer">
          <a-button @click="closeScaleForm">
            取消
          </a-button>
          <a-button type="primary" :loading="scaleSaving" @click="submitScaleForm">
            保存
          </a-button>
        </div>
      </template>
    </PlatformModalShell>

    <PlatformModalShell
      v-model:open="authManageOpen"
      :title="activeAuthScale ? `${activeAuthScale.name} · 授权机构` : '授权机构'"
      :width="760"
      :scrollable="true"
      modal-class="scale-auth-modal"
    >
      <template v-if="activeAuthScale">
        <div class="auth-manage">
          <div class="auth-manage__head">
            <div>
              <div class="auth-manage__title">
                {{ activeAuthScale.name }}
              </div>
              <div class="auth-manage__meta">
                {{ activeAuthScale.code }} · {{ activeAuthScale.currentVersion }}
              </div>
            </div>
            <a-tag color="green">
              {{ activeAuthScale.institutionCount }} 家
            </a-tag>
          </div>

          <a-table
            :columns="[
              { title: '机构名称', dataIndex: 'name', key: 'name' },
              { title: '联系人', dataIndex: 'contact', key: 'contact', width: 160 },
              { title: '授权状态', dataIndex: 'authState', key: 'authState', width: 120 },
              { title: '到期时间', dataIndex: 'expireAt', key: 'expireAt', width: 130 },
            ]"
            :data-source="activeAuthScale.authInstitutions"
            :pagination="false"
            row-key="name"
            size="small"
          />
        </div>
      </template>
    </PlatformModalShell>

    <PlatformModalShell
      v-model:open="referenceManageOpen"
      :title="activeResourceScale ? `${activeResourceScale.name} · 引用文献管理` : '引用文献管理'"
      :width="720"
      :scrollable="true"
      modal-class="scale-reference-modal"
      @close="resetResourceForm"
    >
      <template v-if="activeResourceScale">
        <div class="resource-manage resource-manage--reference">
          <div class="resource-hero">
            <div class="resource-hero__icon">
              <FileTextOutlined />
            </div>

            <div class="resource-hero__content">
              <div class="resource-hero__meta">
                <span>{{ activeResourceScale.code }}</span>
                <span>{{ activeResourceScale.currentVersion }}</span>
              </div>
              <div class="resource-hero__title">
                引用文献
              </div>
            </div>

            <div class="resource-hero__count">
              <strong>{{ activeResourceScale.references.length }}</strong>
              <span>条</span>
            </div>

            <a-button type="primary" class="resource-hero__action" @click="startCreateResource('references')">
              <template #icon>
                <PlusOutlined />
              </template>
              新增引用文献
            </a-button>
          </div>

          <div v-if="resourceFormKind === 'references'" class="resource-editor">
            <div class="resource-editor__head">
              <span>{{ resourceFormIndex === null ? '新增引用文献' : '编辑引用文献' }}</span>
            </div>
            <div class="resource-editor__body">
              <a-textarea
                v-model:value="resourceFormContent"
                :auto-size="{ minRows: 4, maxRows: 7 }"
                placeholder="请输入引用文献内容"
              />
              <div class="resource-editor__actions">
                <a-button @click="resetResourceForm">
                  取消
                </a-button>
                <a-button type="primary" :loading="resourceSaving" @click="submitResourceForm('references')">
                  保存
                </a-button>
              </div>
            </div>
          </div>

          <div class="resource-list-panel">
            <div v-if="activeResourceScale.references.length" class="resource-list">
              <div
                v-for="(reference, index) in activeResourceScale.references"
                :key="reference.id"
                class="resource-list__item"
                :class="{ 'is-editing': isEditingResource('references', index) }"
              >
                <span class="resource-list__index">{{ formatResourceIndex(index) }}</span>
                <div class="resource-list__body">
                  <p>{{ reference.content }}</p>
                </div>
                <div class="resource-list__actions">
                  <a-button type="link" size="small" @click="startEditResource('references', index)">
                    <template #icon>
                      <EditOutlined />
                    </template>
                    编辑
                  </a-button>
                  <a-popconfirm
                    title="确认删除这条引用文献？"
                    ok-text="删除"
                    cancel-text="取消"
                    @confirm="removeResource('references', index)"
                  >
                    <a-button type="link" danger size="small" :loading="resourceSaving">
                      <template #icon>
                        <DeleteOutlined />
                      </template>
                      删除
                    </a-button>
                  </a-popconfirm>
                </div>
              </div>
            </div>
            <div v-else class="resource-empty">
              <a-empty description="暂无引用文献">
                <a-button type="primary" @click="startCreateResource('references')">
                  <template #icon>
                    <PlusOutlined />
                  </template>
                  新增引用文献
                </a-button>
              </a-empty>
            </div>
          </div>
        </div>
      </template>
    </PlatformModalShell>

    <PlatformModalShell
      v-model:open="thanksManageOpen"
      :title="activeResourceScale ? `${activeResourceScale.name} · 特别鸣谢管理` : '特别鸣谢管理'"
      :width="720"
      :scrollable="true"
      modal-class="scale-thanks-modal"
      @close="resetResourceForm"
    >
      <template v-if="activeResourceScale">
        <div class="resource-manage resource-manage--thanks">
          <div class="resource-hero">
            <div class="resource-hero__icon">
              <HeartOutlined />
            </div>

            <div class="resource-hero__content">
              <div class="resource-hero__meta">
                <span>{{ activeResourceScale.code }}</span>
                <span>{{ activeResourceScale.currentVersion }}</span>
              </div>
              <div class="resource-hero__title">
                特别鸣谢
              </div>
            </div>

            <div class="resource-hero__count">
              <strong>{{ activeResourceScale.acknowledgements.length }}</strong>
              <span>条</span>
            </div>

            <a-button type="primary" class="resource-hero__action" @click="startCreateResource('acknowledgements')">
              <template #icon>
                <PlusOutlined />
              </template>
              新增特别鸣谢
            </a-button>
          </div>

          <div v-if="resourceFormKind === 'acknowledgements'" class="resource-editor">
            <div class="resource-editor__head">
              <span>{{ resourceFormIndex === null ? '新增特别鸣谢' : '编辑特别鸣谢' }}</span>
            </div>
            <div class="resource-editor__body">
              <a-textarea
                v-model:value="resourceFormContent"
                :auto-size="{ minRows: 4, maxRows: 7 }"
                placeholder="请输入特别鸣谢内容"
              />
              <div class="resource-editor__actions">
                <a-button @click="resetResourceForm">
                  取消
                </a-button>
                <a-button type="primary" :loading="resourceSaving" @click="submitResourceForm('acknowledgements')">
                  保存
                </a-button>
              </div>
            </div>
          </div>

          <div class="resource-list-panel">
            <div v-if="activeResourceScale.acknowledgements.length" class="resource-list">
              <div
                v-for="(acknowledgement, index) in activeResourceScale.acknowledgements"
                :key="acknowledgement.id"
                class="resource-list__item"
                :class="{ 'is-editing': isEditingResource('acknowledgements', index) }"
              >
                <span class="resource-list__index">{{ formatResourceIndex(index) }}</span>
                <div class="resource-list__body">
                  <p>{{ acknowledgement.content }}</p>
                </div>
                <div class="resource-list__actions">
                  <a-button type="link" size="small" @click="startEditResource('acknowledgements', index)">
                    <template #icon>
                      <EditOutlined />
                    </template>
                    编辑
                  </a-button>
                  <a-popconfirm
                    title="确认删除这条特别鸣谢？"
                    ok-text="删除"
                    cancel-text="取消"
                    @confirm="removeResource('acknowledgements', index)"
                  >
                    <a-button type="link" danger size="small" :loading="resourceSaving">
                      <template #icon>
                        <DeleteOutlined />
                      </template>
                      删除
                    </a-button>
                  </a-popconfirm>
                </div>
              </div>
            </div>
            <div v-else class="resource-empty">
              <a-empty description="暂无特别鸣谢">
                <a-button type="primary" @click="startCreateResource('acknowledgements')">
                  <template #icon>
                    <PlusOutlined />
                  </template>
                  新增特别鸣谢
                </a-button>
              </a-empty>
            </div>
          </div>
        </div>
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

.scale-cell__name-button {
  display: block;
  width: 100%;
  padding: 0;
  border: 0;
  background: transparent;
  text-align: left;
  cursor: pointer;
  transition: color 0.16s ease;
}

.scale-cell__name-button:hover {
  color: #1677ff;
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
  flex-wrap: nowrap;
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

.scale-form {
  padding-top: 2px;
}

.scale-form__grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0 16px;
}

.scale-form__range {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 18px minmax(0, 1fr);
  gap: 6px;
  align-items: center;
}

.scale-form__range-separator {
  color: #8c8c8c;
  text-align: center;
  line-height: 32px;
}

.scale-form__number {
  width: 100%;
}

.scale-poster-field {
  display: flex;
  flex-direction: column;
  gap: 8px;
  width: 100%;

  :deep(.ant-upload-wrapper),
  :deep(.ant-upload) {
    display: block;
    width: 100%;
    line-height: 0;
  }
}

.scale-poster-card {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 220px;
  overflow: hidden;
  border: 1px dashed #d9d9d9;
  border-radius: 8px;
  background: #fafafa;
  cursor: pointer;

  img {
    display: block;
    width: 100%;
    height: 100%;
    object-fit: contain;
  }
}

.scale-poster-card__empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  color: #8c8c8c;
  font-size: 13px;
  line-height: 18px;

  .anticon {
    color: var(--pro-ant-color-primary, #1677ff);
    font-size: 20px;
  }

  em {
    color: #bfbfbf;
    font-size: 12px;
    font-style: normal;
    line-height: 18px;
  }
}

.scale-poster-card__mask {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-size: 13px;
  line-height: 18px;
  background: rgba(0, 0, 0, 0.46);
}

.scale-poster-field__hint-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  color: #8c8c8c;
  font-size: 12px;
  line-height: 20px;
}

.scale-modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

.auth-manage {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.auth-manage__head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  padding: 14px 16px;
  border: 1px solid #e6edf7;
  border-radius: 8px;
  background: #fbfcfe;
}

.auth-manage__title {
  color: #1f2329;
  font-size: 15px;
  font-weight: 700;
  line-height: 24px;
}

.auth-manage__meta {
  margin-top: 2px;
  color: #8c8c8c;
  font-size: 12px;
  line-height: 18px;
}

.resource-manage {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.resource-hero {
  display: grid;
  grid-template-columns: 44px minmax(0, 1fr) auto auto;
  gap: 12px;
  align-items: center;
  padding: 14px 16px;
  border: 1px solid #e6edf7;
  border-radius: 10px;
  background: #fbfcfe;
}

.resource-manage--thanks .resource-hero {
  border-color: #f5e3ef;
  background: #fffafc;
}

.resource-hero__icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 44px;
  height: 44px;
  border-radius: 12px;
  color: #1677ff;
  background: #eef4ff;
  font-size: 18px;
}

.resource-manage--thanks .resource-hero__icon {
  color: #d4380d;
  background: #fff1f0;
}

.resource-hero__content {
  min-width: 0;
}

.resource-hero__meta {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #8c8c8c;
  font-size: 12px;
  line-height: 18px;
}

.resource-hero__meta span + span::before {
  content: '·';
  margin-right: 8px;
  color: #c8cdd6;
}

.resource-hero__title {
  margin-top: 2px;
  color: #1f2329;
  font-size: 15px;
  font-weight: 700;
  line-height: 24px;
}

.resource-hero__count {
  display: inline-flex;
  align-items: baseline;
  gap: 4px;
  padding: 4px 10px;
  border-radius: 999px;
  background: #f2f5fb;
  color: #475467;
  white-space: nowrap;
}

.resource-manage--thanks .resource-hero__count {
  background: #fff1f5;
}

.resource-hero__count strong {
  color: #1f2329;
  font-size: 18px;
  line-height: 24px;
}

.resource-hero__count span {
  font-size: 12px;
  line-height: 18px;
}

.resource-hero__action {
  justify-self: end;
}

.resource-editor {
  padding: 14px 16px 12px;
  border: 1px solid #d6e4ff;
  border-radius: 10px;
  background: #f8fbff;
}

.resource-manage--thanks .resource-editor {
  border-color: #f5d6e2;
  background: #fff8fb;
}

.resource-editor__head {
  margin-bottom: 10px;
  color: #1f2329;
  font-size: 13px;
  font-weight: 600;
  line-height: 20px;
}

.resource-editor__body {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.resource-editor__actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

.resource-list-panel {
  padding: 2px 0 0;
}

.resource-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin: 0;
  list-style: none;
}

.resource-list__item {
  display: grid;
  grid-template-columns: 28px minmax(0, 1fr) auto;
  gap: 10px 12px;
  align-items: start;
  padding: 14px 14px 13px;
  border: 1px solid #edf0f5;
  border-radius: 10px;
  background: #fff;
}

.resource-list__item.is-editing {
  border-color: #91caff;
  background: #f5f9ff;
}

.resource-list__index {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  margin-top: 1px;
  border-radius: 999px;
  background: #eef4ff;
  color: #1677ff;
  font-size: 12px;
  font-weight: 600;
  line-height: 24px;
}

.resource-manage--thanks .resource-list__index {
  color: #c41d7f;
  background: #fff0f6;
}

.resource-list__body {
  min-width: 0;
}

.resource-list__item p {
  margin: 0;
  color: #344054;
  font-size: 14px;
  line-height: 22px;
  word-break: break-word;
}

.resource-list__actions {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding-top: 0;
  white-space: nowrap;
}

.resource-list__actions :deep(.ant-btn-link) {
  height: 28px;
  padding: 0 4px;
  color: #667085;
  font-size: 13px;
  font-weight: 400;
  line-height: 22px;
}

.resource-list__actions :deep(.ant-btn-link:hover) {
  color: #1677ff;
}

.resource-list__actions :deep(.ant-btn-link.ant-btn-dangerous) {
  color: #ff4d4f;
}

.resource-empty {
  padding: 16px 0 4px;
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

  .scale-form__grid {
    grid-template-columns: 1fr;
  }

  .scale-poster-field {
    grid-template-columns: 1fr;
  }

  .scale-poster-card {
    height: 180px;
  }

  .resource-hero {
    grid-template-columns: 44px minmax(0, 1fr);
    align-items: start;
  }

  .resource-hero__count {
    justify-self: start;
  }

  .resource-hero__action {
    justify-self: start;
  }

  .resource-list__item {
    grid-template-columns: 28px minmax(0, 1fr);
  }

  .resource-list__actions {
    grid-column: 2 / -1;
    justify-content: flex-start;
    padding-top: 2px;
  }
}
</style>
