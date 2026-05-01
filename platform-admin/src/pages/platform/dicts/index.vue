<script setup lang="ts">
import type { TableColumnsType } from 'ant-design-vue'
import type { DictItem, DictMutationPayload, DictValueItem, DictValueMutationPayload } from '@/api/platform/dicts'
import {
  PlusOutlined,
  ReloadOutlined,
  SearchOutlined,
} from '@ant-design/icons-vue'
import { computed, onMounted, reactive, ref } from 'vue'
import {
  createDictApi,
  createDictValueApi,
  deleteDictApi,
  deleteDictValueApi,
  listDictValuesApi,
  pageDictsApi,
  updateDictApi,
  updateDictValueApi,
} from '@/api/platform/dicts'
import messageService from '@/utils/messageService'
import PlatformModalShell from '../shared/platform-modal-shell.vue'
import { PlatformAccessEnum } from '~@/constants/access'

type FormMode = 'create' | 'edit'

interface DictFormState {
  id?: number
  dictName: string
  dictCode: string
  isEnable: boolean
  remark: string
}

interface DictValueFormState {
  id?: number
  dictId?: number
  dictLabel: string
  dictValue: string
  sort: number
  isEnable: boolean
  remark: string
}

const { hasAccess } = useAccess()

const keyword = ref('')
const loading = ref(false)
const valueLoading = ref(false)
const savingDict = ref(false)
const savingValue = ref(false)
const dictModalOpen = ref(false)
const valueManageOpen = ref(false)
const valueModalOpen = ref(false)
const dictFormMode = ref<FormMode>('create')
const valueFormMode = ref<FormMode>('create')
const dictList = ref<DictItem[]>([])
const valueList = ref<DictValueItem[]>([])
const selectedDict = ref<DictItem | null>(null)

const pagination = reactive({
  current: 1,
  pageSize: 20,
  total: 0,
  showSizeChanger: true,
  showTotal: (total: number) => `共 ${total} 个字典`,
})

const dictForm = reactive<DictFormState>({
  id: undefined,
  dictName: '',
  dictCode: '',
  isEnable: true,
  remark: '',
})

const valueForm = reactive<DictValueFormState>({
  id: undefined,
  dictId: undefined,
  dictLabel: '',
  dictValue: '',
  sort: 1,
  isEnable: true,
  remark: '',
})

const dictColumns: TableColumnsType<DictItem> = [
  { title: '字典名称', dataIndex: 'dictName', key: 'dictName', width: 260, fixed: 'left' as const },
  { title: '字典编码', dataIndex: 'dictCode', key: 'dictCode', width: 220 },
  { title: '状态', dataIndex: 'isEnable', key: 'isEnable', width: 110, align: 'center' as const },
  { title: '备注', dataIndex: 'remark', key: 'remark', width: 360 },
  { title: '操作', key: 'action', width: 220, fixed: 'right' as const },
]

const valueColumns: TableColumnsType<DictValueItem> = [
  { title: '字典项名称', dataIndex: 'dictLabel', key: 'dictLabel', width: 180, fixed: 'left' as const },
  { title: '字典项值', dataIndex: 'dictValue', key: 'dictValue', width: 180 },
  { title: '排序', dataIndex: 'sort', key: 'sort', width: 88, align: 'center' as const },
  { title: '状态', dataIndex: 'isEnable', key: 'isEnable', width: 96, align: 'center' as const },
  { title: '操作', key: 'action', width: 136, fixed: 'right' as const },
]

const dictModalTitle = computed(() => dictFormMode.value === 'edit' ? '编辑字典' : '新增字典')
const valueModalTitle = computed(() => valueFormMode.value === 'edit' ? '编辑字典项' : '新增字典项')
const selectedDictTitle = computed(() => {
  if (!selectedDict.value)
    return '请选择字典'
  return `${selectedDict.value.dictName || '--'} · ${selectedDict.value.dictCode || '--'}`
})
const valueManageTitle = computed(() => selectedDict.value ? `字典项管理 · ${selectedDict.value.dictName || '--'}` : '字典项管理')
const canManageDictValues = computed(() => hasAccess([
  PlatformAccessEnum.dictValueAdd,
  PlatformAccessEnum.dictValueEdit,
  PlatformAccessEnum.dictValueDelete,
]))

function getErrorMessage(error: any, fallback: string) {
  return error?.response?.data?.message || error?.message || fallback
}

function normalizeEnable(value: unknown) {
  return value === true || value === 1 || value === '1'
}

function resetDictForm(record?: DictItem) {
  dictForm.id = record?.id
  dictForm.dictName = record?.dictName || ''
  dictForm.dictCode = record?.dictCode || ''
  dictForm.isEnable = record ? normalizeEnable(record.isEnable) : true
  dictForm.remark = record?.remark || ''
}

function resetValueForm(record?: DictValueItem) {
  valueForm.id = record?.id
  valueForm.dictId = record?.dictId || selectedDict.value?.id
  valueForm.dictLabel = record?.dictLabel || ''
  valueForm.dictValue = record?.dictValue || ''
  valueForm.sort = Number(record?.sort || 1)
  valueForm.isEnable = record ? normalizeEnable(record.isEnable) : true
  valueForm.remark = ''
}

async function fetchDictValues(dict = selectedDict.value) {
  if (!dict?.dictCode) {
    valueList.value = []
    return
  }

  valueLoading.value = true
  try {
    const res = await listDictValuesApi({ code: dict.dictCode })
    if (res.code !== 200) {
      messageService.error(res.message || '获取字典项失败')
      return
    }
    valueList.value = Array.isArray(res.result) ? res.result : []
  }
  catch (error: any) {
    console.error('fetch dict values failed', error)
    messageService.error(getErrorMessage(error, '获取字典项失败'))
  }
  finally {
    valueLoading.value = false
  }
}

async function fetchDicts() {
  loading.value = true
  try {
    const res = await pageDictsApi({
      current: pagination.current,
      size: pagination.pageSize,
      keyword: keyword.value.trim() || undefined,
      scope: 'scale',
    })
    if (res.code !== 200) {
      messageService.error(res.message || '获取字典列表失败')
      return
    }

    const list = Array.isArray(res.result) ? res.result : []
    dictList.value = list
    pagination.total = Number(res.total || 0)

    if (selectedDict.value) {
      const latestSelected = list.find(item => item.id === selectedDict.value?.id) || null
      selectedDict.value = latestSelected
      if (!latestSelected) {
        valueList.value = []
        valueManageOpen.value = false
      }
      else if (valueManageOpen.value) {
        await fetchDictValues(latestSelected)
      }
    }
  }
  catch (error: any) {
    console.error('fetch dicts failed', error)
    messageService.error(getErrorMessage(error, '获取字典列表失败'))
  }
  finally {
    loading.value = false
  }
}

function handleSearch() {
  pagination.current = 1
  fetchDicts()
}

function handleReset() {
  keyword.value = ''
  pagination.current = 1
  fetchDicts()
}

function handleTableChange(page: { current?: number, pageSize?: number }) {
  pagination.current = page.current || 1
  pagination.pageSize = page.pageSize || 20
  fetchDicts()
}

function openValueManage(record: DictItem) {
  selectedDict.value = record
  valueManageOpen.value = true
  fetchDictValues(record)
}

function openCreateDict() {
  dictFormMode.value = 'create'
  resetDictForm()
  dictModalOpen.value = true
}

function openEditDict(record: DictItem) {
  dictFormMode.value = 'edit'
  resetDictForm(record)
  dictModalOpen.value = true
}

async function submitDict() {
  const dictName = dictForm.dictName.trim()
  const dictCode = dictForm.dictCode.trim()
  if (!dictName) {
    messageService.warning('请输入字典名称')
    return
  }
  if (!dictCode) {
    messageService.warning('请输入字典编码')
    return
  }

  const payload: DictMutationPayload = {
    dictName,
    dictCode,
    isEnable: dictForm.isEnable,
    remark: dictForm.remark.trim(),
  }

  savingDict.value = true
  try {
    if (dictFormMode.value === 'edit' && dictForm.id) {
      const res = await updateDictApi({ ...payload, id: dictForm.id })
      if (res.code !== 200) {
        messageService.error(res.message || '保存字典失败')
        return
      }
      messageService.success('字典已更新')
      dictModalOpen.value = false
      await fetchDicts()
    }
    else {
      const res = await createDictApi(payload)
      if (res.code !== 200) {
        messageService.error(res.message || '新增字典失败')
        return
      }
      messageService.success('字典已新增')
      dictModalOpen.value = false
      pagination.current = 1
      await fetchDicts()
    }
  }
  catch (error: any) {
    console.error('save dict failed', error)
    messageService.error(getErrorMessage(error, '保存字典失败'))
  }
  finally {
    savingDict.value = false
  }
}

async function removeDict(record: DictItem) {
  try {
    const res = await deleteDictApi({ id: record.id })
    if (res.code !== 200) {
      messageService.error(res.message || '删除字典失败')
      return
    }
    messageService.success('字典已删除')
    if (dictList.value.length <= 1 && pagination.current > 1)
      pagination.current -= 1
    if (selectedDict.value?.id === record.id) {
      selectedDict.value = null
      valueList.value = []
      valueManageOpen.value = false
    }
    await fetchDicts()
  }
  catch (error: any) {
    console.error('delete dict failed', error)
    messageService.error(getErrorMessage(error, '删除字典失败'))
  }
}

function openCreateValue() {
  if (!selectedDict.value) {
    messageService.warning('请先选择字典')
    return
  }
  valueFormMode.value = 'create'
  resetValueForm()
  valueModalOpen.value = true
}

function openEditValue(record: DictValueItem) {
  valueFormMode.value = 'edit'
  resetValueForm(record)
  valueModalOpen.value = true
}

async function submitValue() {
  if (!selectedDict.value?.id) {
    messageService.warning('请先选择字典')
    return
  }

  const dictLabel = valueForm.dictLabel.trim()
  const dictValue = valueForm.dictValue.trim()
  if (!dictLabel) {
    messageService.warning('请输入字典项名称')
    return
  }
  if (!dictValue) {
    messageService.warning('请输入字典项值')
    return
  }

  const payload: DictValueMutationPayload = {
    dictId: selectedDict.value.id,
    dictLabel,
    dictValue,
    sort: Number(valueForm.sort || 1),
    isEnable: valueForm.isEnable,
    remark: valueForm.remark.trim(),
  }

  savingValue.value = true
  try {
    if (valueFormMode.value === 'edit' && valueForm.id) {
      const res = await updateDictValueApi({ ...payload, id: valueForm.id })
      if (res.code !== 200) {
        messageService.error(res.message || '保存字典项失败')
        return
      }
      messageService.success('字典项已更新')
    }
    else {
      const res = await createDictValueApi({ ...payload, dictId: selectedDict.value.id })
      if (res.code !== 200) {
        messageService.error(res.message || '新增字典项失败')
        return
      }
      messageService.success('字典项已新增')
    }
    valueModalOpen.value = false
    await fetchDictValues()
  }
  catch (error: any) {
    console.error('save dict value failed', error)
    messageService.error(getErrorMessage(error, '保存字典项失败'))
  }
  finally {
    savingValue.value = false
  }
}

async function removeValue(record: DictValueItem) {
  try {
    const res = await deleteDictValueApi({ id: record.id })
    if (res.code !== 200) {
      messageService.error(res.message || '删除字典项失败')
      return
    }
    messageService.success('字典项已删除')
    await fetchDictValues()
  }
  catch (error: any) {
    console.error('delete dict value failed', error)
    messageService.error(getErrorMessage(error, '删除字典项失败'))
  }
}

onMounted(() => {
  fetchDicts()
})
</script>

<template>
  <div class="dict-page">
    <div class="dict-page__header">
      <div class="dict-page__title">
        字典管理
      </div>

      <a-button v-if="hasAccess(PlatformAccessEnum.dictAdd)" type="primary" @click="openCreateDict">
        <template #icon>
          <PlusOutlined />
        </template>
        新增字典
      </a-button>
    </div>

    <section class="dict-panel">
      <div class="dict-toolbar">
        <div class="dict-filter">
          <span class="dict-filter__label">关键词搜索</span>
          <a-input
            v-model:value="keyword"
            allow-clear
            class="dict-filter__input"
            placeholder="搜索字典名称、编码"
            @press-enter="handleSearch"
          />
        </div>

        <a-button type="primary" @click="handleSearch">
          <template #icon>
            <SearchOutlined />
          </template>
          搜索
        </a-button>

        <a-button @click="handleReset">
          <template #icon>
            <ReloadOutlined />
          </template>
          重置
        </a-button>
      </div>

      <a-table
        class="dict-table"
        :columns="dictColumns"
        :data-source="dictList"
        :loading="loading"
        :pagination="pagination"
        :scroll="{ x: 1170 }"
        row-key="id"
        size="small"
        @change="handleTableChange"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'dictName'">
            <div class="dict-name-cell">
              <a-tooltip :overlay-style="{ maxWidth: '320px', whiteSpace: 'normal' }">
                <template #title>
                  {{ record.dictName || '--' }}
                </template>
                <span class="dict-name-cell__title">{{ record.dictName || '--' }}</span>
              </a-tooltip>
            </div>
          </template>

          <template v-else-if="column.key === 'dictCode'">
            <span class="mono-text" :title="record.dictCode || '--'">{{ record.dictCode || '--' }}</span>
          </template>

          <template v-else-if="column.key === 'isEnable'">
            <a-tag :color="record.isEnable ? 'green' : 'default'">
              {{ record.isEnable ? '启用' : '停用' }}
            </a-tag>
          </template>

          <template v-else-if="column.key === 'remark'">
            <span class="ellipsis-text" :title="record.remark || '--'">{{ record.remark || '--' }}</span>
          </template>

          <template v-else-if="column.key === 'action'">
            <div class="dict-actions">
              <a-button
                v-if="canManageDictValues"
                type="link"
                size="small"
                @click="openValueManage(record as DictItem)"
              >
                字典项
              </a-button>
              <a-button
                v-if="hasAccess(PlatformAccessEnum.dictEdit)"
                type="link"
                size="small"
                @click="openEditDict(record as DictItem)"
              >
                编辑
              </a-button>
              <a-popconfirm
                v-if="hasAccess(PlatformAccessEnum.dictDelete)"
                title="确定删除这个字典吗？"
                ok-text="删除"
                cancel-text="取消"
                @confirm="removeDict(record as DictItem)"
              >
                <a-button type="link" size="small" danger>
                  删除
                </a-button>
              </a-popconfirm>
            </div>
          </template>
        </template>
      </a-table>
    </section>

    <PlatformModalShell
      v-model:open="valueManageOpen"
      :title="valueManageTitle"
      :width="960"
      modal-class="dict-values-modal"
      scrollable
      @close="valueManageOpen = false"
    >
      <div class="dict-values__header">
        <div class="dict-values__heading">
          <div class="dict-values__title">
            字典项
          </div>
          <div class="dict-values__subtitle">
            {{ selectedDictTitle }}
          </div>
        </div>

        <a-button
          v-if="hasAccess(PlatformAccessEnum.dictValueAdd)"
          type="primary"
          :disabled="!selectedDict"
          @click="openCreateValue"
        >
          <template #icon>
            <PlusOutlined />
          </template>
          新增字典项
        </a-button>
      </div>

      <a-table
        class="dict-table dict-values-table"
        :columns="valueColumns"
        :data-source="valueList"
        :loading="valueLoading"
        :pagination="false"
        :scroll="{ x: 680 }"
        row-key="id"
        size="small"
      >
        <template #emptyText>
          暂无字典项
        </template>

        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'dictLabel'">
            <span class="ellipsis-text strong-text" :title="record.dictLabel || '--'">{{ record.dictLabel || '--' }}</span>
          </template>

          <template v-else-if="column.key === 'dictValue'">
            <span class="mono-text" :title="record.dictValue || '--'">{{ record.dictValue || '--' }}</span>
          </template>

          <template v-else-if="column.key === 'isEnable'">
            <a-tag :color="record.isEnable ? 'green' : 'default'">
              {{ record.isEnable ? '启用' : '停用' }}
            </a-tag>
          </template>

          <template v-else-if="column.key === 'action'">
            <div class="dict-actions">
              <a-button
                v-if="hasAccess(PlatformAccessEnum.dictValueEdit)"
                type="link"
                size="small"
                @click="openEditValue(record as DictValueItem)"
              >
                编辑
              </a-button>
              <a-popconfirm
                v-if="hasAccess(PlatformAccessEnum.dictValueDelete)"
                title="确定删除这个字典项吗？"
                ok-text="删除"
                cancel-text="取消"
                @confirm="removeValue(record as DictValueItem)"
              >
                <a-button type="link" size="small" danger>
                  删除
                </a-button>
              </a-popconfirm>
            </div>
          </template>
        </template>
      </a-table>
    </PlatformModalShell>

    <PlatformModalShell
      v-model:open="dictModalOpen"
      :title="dictModalTitle"
      :width="680"
      modal-class="dict-form-modal"
      @close="dictModalOpen = false"
    >
      <a-form :model="dictForm" layout="vertical" class="dict-form">
        <a-form-item label="字典名称" required>
          <a-input v-model:value="dictForm.dictName" placeholder="请输入字典名称" />
        </a-form-item>

        <a-form-item label="字典编码" required>
          <a-input v-model:value="dictForm.dictCode" placeholder="请输入字典编码，例如 pep3_domain_type" />
        </a-form-item>

        <a-form-item label="状态">
          <a-switch v-model:checked="dictForm.isEnable" checked-children="启用" un-checked-children="停用" />
        </a-form-item>

        <a-form-item label="备注">
          <a-textarea v-model:value="dictForm.remark" :rows="3" placeholder="请输入备注" />
        </a-form-item>
      </a-form>

      <template #footer>
        <div class="dict-modal-footer">
          <a-button @click="dictModalOpen = false">
            取消
          </a-button>
          <a-button type="primary" :loading="savingDict" @click="submitDict">
            保存
          </a-button>
        </div>
      </template>
    </PlatformModalShell>

    <PlatformModalShell
      v-model:open="valueModalOpen"
      :title="valueModalTitle"
      :width="680"
      modal-class="dict-value-form-modal"
      @close="valueModalOpen = false"
    >
      <a-form :model="valueForm" layout="vertical" class="dict-form">
        <a-form-item label="所属字典">
          <a-input :value="selectedDictTitle" disabled />
        </a-form-item>

        <a-form-item label="字典项名称" required>
          <a-input v-model:value="valueForm.dictLabel" placeholder="请输入字典项名称" />
        </a-form-item>

        <a-form-item label="字典项值" required>
          <a-input v-model:value="valueForm.dictValue" placeholder="请输入字典项值" />
        </a-form-item>

        <div class="dict-form__inline">
          <a-form-item label="排序">
            <a-input-number v-model:value="valueForm.sort" :min="0" :precision="0" style="width: 100%" />
          </a-form-item>

          <a-form-item label="状态">
            <a-switch v-model:checked="valueForm.isEnable" checked-children="启用" un-checked-children="停用" />
          </a-form-item>
        </div>

        <a-form-item label="备注">
          <a-textarea v-model:value="valueForm.remark" :rows="3" placeholder="请输入备注" />
        </a-form-item>
      </a-form>

      <template #footer>
        <div class="dict-modal-footer">
          <a-button @click="valueModalOpen = false">
            取消
          </a-button>
          <a-button type="primary" :loading="savingValue" @click="submitValue">
            保存
          </a-button>
        </div>
      </template>
    </PlatformModalShell>
  </div>
</template>

<style scoped lang="less">
.dict-page {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.dict-page__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 2px 2px 0;
}

.dict-page__title {
  color: #1f2329;
  font-size: 16px;
  font-weight: 700;
  line-height: 32px;
}

.dict-panel {
  overflow: hidden;
  border: 1px solid #e9edf3;
  border-radius: 10px;
  background: #fff;
}

.dict-toolbar {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 10px;
  padding: 16px;
}

.dict-filter {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.dict-filter__label {
  flex-shrink: 0;
  color: #262626;
  font-size: 14px;
  line-height: 32px;
  white-space: nowrap;
}

.dict-filter__input {
  width: 300px;
}

.dict-values__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  min-height: 54px;
  padding: 2px 0 14px;
  border-bottom: 1px solid #edf0f5;
}

.dict-values__heading {
  min-width: 0;
}

.dict-values__title {
  color: #1f2329;
  font-size: 15px;
  font-weight: 700;
  line-height: 22px;
}

.dict-values__subtitle {
  margin-top: 2px;
  overflow: hidden;
  color: #8c8c8c;
  font-size: 12px;
  line-height: 18px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.dict-table {
  padding: 0 8px 8px;
}

.dict-values-table {
  padding: 12px 0 0;
}

.dict-table :deep(.ant-table-thead > tr > th) {
  padding: 12px 16px;
  background: #fafafa !important;
  color: #262626;
  font-size: 14px;
  font-weight: 500;
  line-height: 20px;
  border-bottom: 1px solid #f0f0f0;
  white-space: nowrap;
}

.dict-table :deep(.ant-table-tbody > tr > td) {
  padding: 14px 16px;
  vertical-align: middle;
  border-bottom: 1px solid #f5f5f5;
}

.dict-table :deep(.ant-table-tbody > tr:hover > td) {
  background: #fcfcfc;
}

.dict-table :deep(.ant-table-cell-fix-left),
.dict-table :deep(.ant-table-cell-fix-right) {
  background: #fff;
}

.dict-name-cell {
  min-width: 0;
}

.dict-name-cell__title,
.strong-text {
  display: inline-block;
  max-width: 100%;
  overflow: hidden;
  color: #262626;
  font-size: 14px;
  font-weight: 600;
  line-height: 22px;
  text-overflow: ellipsis;
  white-space: nowrap;
  vertical-align: middle;
}

.mono-text {
  display: inline-block;
  max-width: 100%;
  overflow: hidden;
  color: #344054;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
  font-size: 12px;
  line-height: 20px;
  text-overflow: ellipsis;
  white-space: nowrap;
  vertical-align: middle;
}

.ellipsis-text {
  display: inline-block;
  max-width: 100%;
  overflow: hidden;
  color: #667085;
  font-size: 13px;
  line-height: 22px;
  text-overflow: ellipsis;
  white-space: nowrap;
  vertical-align: middle;
}

.dict-actions {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  white-space: nowrap;
}

.dict-actions :deep(.ant-btn-link) {
  padding: 0 2px;
  font-weight: 400;
}

.dict-form {
  padding-top: 4px;
}

.dict-form__inline {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 160px;
  gap: 16px;
  align-items: start;
}

.dict-modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

@media (max-width: 1180px) {
  .dict-filter__input {
    width: min(300px, calc(100vw - 220px));
  }
}

@media (max-width: 720px) {
  .dict-page__header,
  .dict-values__header {
    align-items: stretch;
    flex-direction: column;
  }

  .dict-filter,
  .dict-filter__input {
    width: 100%;
  }

  .dict-form__inline {
    grid-template-columns: 1fr;
  }
}
</style>
