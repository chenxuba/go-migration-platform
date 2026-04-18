<script setup lang="ts">
import type { PermissionTreeNode } from '../../shared/permission-tree'
import { CloseOutlined } from '@ant-design/icons-vue'
import { computed, ref, watch } from 'vue'
import {
  getInstitutionPermissionDetailApi,
  replaceInstitutionPermissionVersionApi,
  type InstitutionPermissionDetail,
} from '@/api/platform/institutions'
import {
  getVersionDetailApi,
  getVersionMenuTreeApi,
  pageVersionsApi,
  type VersionItem,
} from '@/api/platform/versions'
import {
  buildPermissionTreeData,
  collectAllKeys,
  collectExpandKeysByKeyword,
  collectLeafKeysBySelectedSet,
  countLeafNodes,
} from '../../shared/permission-tree'
import { sortVersionsByDisplayOrder } from '../../shared/version-order'
import messageService from '@/utils/messageService'

const props = defineProps<{
  open: boolean
  institutionId?: number | null
}>()

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'saved'): void
}>()

const openTypeLabelMap: Record<number, string> = {
  1: '体验版',
  2: '基础版',
  3: '高级版',
  4: '旗舰版',
}

const statusLabelMap: Record<number, string> = {
  1: '启用',
  2: '停用',
  4: '过期',
}

const openModal = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

const detailLoading = ref(false)
const submitting = ref(false)
const menuTreeLoading = ref(false)
const versionLoading = ref(false)
const treeKeyword = ref('')
const expandedKeys = ref<number[]>([])
const baseMenuTree = ref<PermissionTreeNode[]>([])
const versionOptions = ref<VersionItem[]>([])
const detail = ref<InstitutionPermissionDetail | null>(null)
const selectedModuleId = ref<number | undefined>()
const templateCheckedKeys = ref<number[]>([])
const editableCheckedKeys = ref<number[]>([])
const editableHalfCheckedKeys = ref<number[]>([])

const rootExpandedKeys = computed(() => baseMenuTree.value.map(node => Number(node.key)))
const totalLeafCount = computed(() => countLeafNodes(baseMenuTree.value))
const templateLeafCount = computed(() => collectLeafKeysBySelectedSet(baseMenuTree.value, templateCheckedKeys.value).length)
const editableLeafCount = computed(() => collectLeafKeysBySelectedSet(baseMenuTree.value, [...editableCheckedKeys.value, ...editableHalfCheckedKeys.value]).length)

function closeModal() {
  emit('update:open', false)
}

function formatDateMinute(value?: string) {
  const raw = String(value || '').trim()
  if (!raw)
    return '--'
  return raw.length >= 16 ? raw.slice(0, 16) : raw
}

function getOpenTypeLabel(value?: number) {
  return openTypeLabelMap[Number(value || 0)] || '--'
}

function getStatusLabel(value?: number) {
  return statusLabelMap[Number(value || 0)] || '--'
}

function getStatusClass(value?: number) {
  const normalized = Number(value || 0)
  if (normalized === 1)
    return 'status-chip--enabled'
  if (normalized === 4)
    return 'status-chip--expired'
  return 'status-chip--disabled'
}

function isSameMenuScope(left: number[] = [], right: number[] = []) {
  if (left.length !== right.length)
    return false

  const rightSet = new Set(right.map(item => Number(item)))
  return left.every(item => rightSet.has(Number(item)))
}

const syncStatusText = computed(() => {
  if (!detail.value?.currentModuleId || !detail.value?.effectiveMenuIds?.length)
    return '未同步'
  return isSameMenuScope(detail.value.templateMenuIds || [], detail.value.effectiveMenuIds || []) ? '已同步' : '未同步'
})

const editableTreeData = computed(() => {
  const allowedKeySet = new Set((templateCheckedKeys.value || []).map(item => Number(item)))
  if (!allowedKeySet.size)
    return baseMenuTree.value

  const mapNodes = (nodes: PermissionTreeNode[]): PermissionTreeNode[] => nodes.map(node => ({
    ...node,
    disableCheckbox: !allowedKeySet.has(Number(node.key)),
    children: Array.isArray(node.children) ? mapNodes(node.children) : undefined,
  }))

  return mapNodes(baseMenuTree.value)
})

async function ensureMenuTreeLoaded() {
  if (baseMenuTree.value.length)
    return

  menuTreeLoading.value = true
  try {
    const res = await getVersionMenuTreeApi({ type: 1 })
    if (res.code !== 200) {
      messageService.error(res.message || '获取权限树失败')
      return
    }
    baseMenuTree.value = buildPermissionTreeData(res.result || [])
    expandedKeys.value = rootExpandedKeys.value
  }
  catch (error: any) {
    console.error('load version menu tree failed', error)
    messageService.error(error?.message || '获取权限树失败')
  }
  finally {
    menuTreeLoading.value = false
  }
}

async function loadVersionOptions() {
  versionLoading.value = true
  try {
    const res = await pageVersionsApi({
      current: 1,
      size: 200,
      type: 1,
    })
    if (res.code !== 200) {
      messageService.error(res.message || '获取版本列表失败')
      return
    }
    versionOptions.value = sortVersionsByDisplayOrder(Array.isArray(res.result) ? res.result : [])
  }
  catch (error: any) {
    console.error('load version options failed', error)
    messageService.error(error?.message || '获取版本列表失败')
  }
  finally {
    versionLoading.value = false
  }
}

async function loadTemplateDetail(moduleId?: number) {
  const targetId = Number(moduleId || 0)
  if (!targetId) {
    templateCheckedKeys.value = []
    return
  }

  try {
    const res = await getVersionDetailApi({ moduleId: targetId })
    if (res.code !== 200 || !res.result) {
      messageService.error(res.message || '获取版本权限失败')
      return
    }
    templateCheckedKeys.value = Array.isArray(res.result.selectedMenuIds)
      ? res.result.selectedMenuIds.map(item => Number(item))
      : []
  }
  catch (error: any) {
    console.error('load template detail failed', error)
    messageService.error(error?.message || '获取版本权限失败')
  }
}

async function loadDetail(institutionId: number) {
  detailLoading.value = true
  try {
    const res = await getInstitutionPermissionDetailApi({ institutionId })
    if (res.code !== 200 || !res.result) {
      messageService.error(res.message || '获取机构权限失败')
      return
    }
    detail.value = res.result
    selectedModuleId.value = Number(res.result.currentModuleId || 0) || undefined
    templateCheckedKeys.value = Array.isArray(res.result.templateMenuIds)
      ? res.result.templateMenuIds.map(item => Number(item))
      : []
    editableCheckedKeys.value = collectLeafKeysBySelectedSet(baseMenuTree.value, res.result.effectiveMenuIds || [])
    editableHalfCheckedKeys.value = []
    expandedKeys.value = treeKeyword.value
      ? collectExpandKeysByKeyword(baseMenuTree.value, treeKeyword.value)
      : rootExpandedKeys.value
  }
  catch (error: any) {
    console.error('load institution permission detail failed', error)
    messageService.error(error?.message || '获取机构权限失败')
  }
  finally {
    detailLoading.value = false
  }
}

async function submitVersionBinding() {
  const institutionId = Number(props.institutionId || 0)
  const moduleId = Number(selectedModuleId.value || 0)
  if (!institutionId || !moduleId)
    return

  submitting.value = true
  try {
    const menuIds = Array.from(new Set([...editableCheckedKeys.value, ...editableHalfCheckedKeys.value]))
      .map(item => Number(item))
      .filter(item => item > 0)

    if (!menuIds.length) {
      messageService.warning('请至少选择一个机构权限')
      return
    }

    const res = await replaceInstitutionPermissionVersionApi({ institutionId, moduleId, menuIds })
    if (res.code !== 200) {
      messageService.error(res.message || '同步机构权限失败')
      return
    }
    messageService.success('机构权限同步成功')
    await loadDetail(institutionId)
    emit('saved')
  }
  catch (error: any) {
    console.error('replace institution permission version failed', error)
    messageService.error(error?.message || '同步机构权限失败')
  }
  finally {
    submitting.value = false
  }
}

function handleEditableTreeCheck(value: any, info: any) {
  editableCheckedKeys.value = Array.isArray(value) ? value.map(Number) : (value?.checked || []).map(Number)
  editableHalfCheckedKeys.value = (info?.halfCheckedKeys || []).map((key: string | number) => Number(key))
}

function applyTemplateScope() {
  editableCheckedKeys.value = collectLeafKeysBySelectedSet(baseMenuTree.value, templateCheckedKeys.value)
  editableHalfCheckedKeys.value = []
}

function handleExpandedKeysChange(keys: Array<string | number>) {
  expandedKeys.value = (keys || []).map(key => Number(key))
}

function expandAllTree() {
  expandedKeys.value = collectAllKeys(baseMenuTree.value)
}

function collapseTree() {
  expandedKeys.value = rootExpandedKeys.value
}

watch(
  () => treeKeyword.value,
  (keyword) => {
    expandedKeys.value = keyword
      ? collectExpandKeysByKeyword(baseMenuTree.value, keyword)
      : rootExpandedKeys.value
  },
)

watch(
  () => selectedModuleId.value,
  async (moduleId, previousId) => {
    if (!openModal.value)
      return
    if (!moduleId)
      return
    if (moduleId === previousId && templateCheckedKeys.value.length)
      return
    await loadTemplateDetail(moduleId)
    if (Number(moduleId) !== Number(detail.value?.currentModuleId || 0))
      applyTemplateScope()
  },
)

watch(
  () => [props.open, props.institutionId] as const,
  async ([open, institutionId]) => {
    if (!open) {
      detail.value = null
      selectedModuleId.value = undefined
      templateCheckedKeys.value = []
      editableCheckedKeys.value = []
      editableHalfCheckedKeys.value = []
      treeKeyword.value = ''
      expandedKeys.value = rootExpandedKeys.value
      return
    }

    await Promise.all([ensureMenuTreeLoaded(), loadVersionOptions()])
    if (institutionId)
      await loadDetail(Number(institutionId))
  },
  { immediate: true },
)
</script>

<template>
  <a-modal
    v-model:open="openModal"
    centered
    destroy-on-close
    :keyboard="false"
    :closable="false"
    :mask-closable="false"
    :width="1260"
    class="createStu-modal-content-box institution-permission-modal"
  >
    <template #title>
      <div class="permission-modal__titlebar">
        <span>机构权限</span>
        <a-button type="text" class="close-btn" @click="closeModal">
          <template #icon>
            <CloseOutlined class="close-icon" />
          </template>
        </a-button>
      </div>
    </template>

    <a-spin :spinning="detailLoading || menuTreeLoading || versionLoading">
      <div class="permission-modal">
        <div class="permission-overview">
          <div class="permission-overview__main">
            <div class="permission-overview__name">
              {{ detail?.organName || '--' }}
            </div>
            <div class="permission-overview__meta">
              <span>登录账号：{{ detail?.mobile || '--' }}</span>
              <span>开通版本：{{ getOpenTypeLabel(detail?.openType) }}</span>
              <span>过期时间：{{ formatDateMinute(detail?.expireEndTime) }}</span>
            </div>
          </div>

          <div class="permission-overview__chips">
            <span class="status-chip" :class="getStatusClass(detail?.status)">
              {{ getStatusLabel(detail?.status) }}
            </span>
            <span class="status-chip" :class="syncStatusText === '已同步' ? 'status-chip--sync' : 'status-chip--warning'">
              {{ syncStatusText }}
            </span>
          </div>
        </div>

        <div class="permission-summary">
          <div class="summary-card">
            <span class="summary-card__label">权限版本</span>
            <span class="summary-card__value">{{ detail?.currentModuleName || '--' }}</span>
          </div>
          <div class="summary-card">
            <span class="summary-card__label">管理员角色</span>
            <span class="summary-card__value">{{ detail?.adminRoleName || '--' }}</span>
          </div>
          <div class="summary-card">
            <span class="summary-card__label">模板菜单</span>
            <span class="summary-card__value">{{ templateLeafCount }} / {{ totalLeafCount }}</span>
          </div>
          <div class="summary-card">
            <span class="summary-card__label">机构权限</span>
            <span class="summary-card__value">{{ editableLeafCount }} / {{ totalLeafCount }}</span>
          </div>
        </div>

        <div class="permission-action">
          <a-select
            v-model:value="selectedModuleId"
            class="permission-action__select"
            placeholder="请选择权限版本"
            :options="versionOptions.map(item => ({ value: item.id, label: item.name }))"
          />
          <a-button type="primary" :loading="submitting" @click="submitVersionBinding">
            同步权限
          </a-button>
        </div>

        <div class="permission-toolbar">
          <a-input
            v-model:value="treeKeyword"
            allow-clear
            placeholder="搜索菜单名称"
            class="permission-toolbar__search"
          />

          <div class="permission-toolbar__actions">
            <a-button type="link" class="permission-toolbar__link" @click="expandAllTree">
              展开全部
            </a-button>
            <a-button type="link" class="permission-toolbar__link" @click="collapseTree">
              收起层级
            </a-button>
          </div>
        </div>

        <div class="permission-panels">
          <div class="permission-panel">
            <div class="permission-panel__header">
              <span class="permission-panel__title">版本权限</span>
              <span class="permission-panel__meta">{{ templateLeafCount }} 项</span>
            </div>
            <div class="permission-panel__body tree-readonly">
              <a-empty v-if="!menuTreeLoading && !baseMenuTree.length" description="暂无菜单" />
              <a-tree
                v-else
                :tree-data="baseMenuTree"
                :checked-keys="templateCheckedKeys"
                :expanded-keys="expandedKeys"
                checkable
                check-strictly
                block-node
                @update:expanded-keys="handleExpandedKeysChange"
              />
            </div>
          </div>

          <div class="permission-panel">
            <div class="permission-panel__header">
              <span class="permission-panel__title">机构权限设置</span>
              <div class="permission-panel__actions">
                <span class="permission-panel__meta">{{ editableLeafCount }} 项</span>
                <a-button type="link" class="permission-panel__reset" @click="applyTemplateScope">
                  重置为版本权限
                </a-button>
              </div>
            </div>
            <div class="permission-panel__body">
              <a-empty v-if="!menuTreeLoading && !editableTreeData.length" description="暂无菜单" />
              <a-tree
                v-else
                :tree-data="editableTreeData"
                :checked-keys="editableCheckedKeys"
                :expanded-keys="expandedKeys"
                checkable
                block-node
                @check="handleEditableTreeCheck"
                @update:expanded-keys="handleExpandedKeysChange"
              />
            </div>
          </div>
        </div>
      </div>
    </a-spin>

    <template #footer>
      <a-button @click="closeModal">
        关闭
      </a-button>
    </template>
  </a-modal>
</template>

<style scoped lang="less">
:deep(.createStu-modal-content-box.institution-permission-modal .ant-modal-content) {
  border-radius: 22px;
  overflow: hidden;
  box-shadow: 0 18px 46px rgba(15, 23, 42, 0.14);
}

:deep(.createStu-modal-content-box.institution-permission-modal .ant-modal-header) {
  padding: 24px 28px 14px;
  margin-bottom: 0;
  border-bottom: none;
}

:deep(.createStu-modal-content-box.institution-permission-modal .ant-modal-body) {
  padding: 0 28px 0;
}

:deep(.createStu-modal-content-box.institution-permission-modal .ant-modal-footer) {
  padding: 18px 28px 24px;
  border-top: none;
}

.permission-modal__titlebar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  color: #1f2329;
  font-size: 20px;
  font-weight: 700;
  line-height: 32px;
}

.permission-modal {
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding-top: 8px;
}

.permission-overview {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 18px 22px;
  border: 1px solid #e8edf5;
  border-radius: 18px;
  background: linear-gradient(180deg, rgba(22, 119, 255, 0.05) 0%, #fff 120px);
  box-shadow: 0 14px 32px rgba(15, 23, 42, 0.05);
}

.permission-overview__main {
  min-width: 0;
}

.permission-overview__name {
  color: #1f2329;
  font-size: 20px;
  font-weight: 700;
  line-height: 30px;
}

.permission-overview__meta {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
  margin-top: 8px;
  color: #667085;
  font-size: 13px;
  line-height: 20px;
}

.permission-overview__chips {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 10px;
}

.status-chip {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 7px 14px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 600;
  line-height: 18px;
}

.status-chip--enabled {
  background: rgba(22, 163, 74, 0.12);
  color: #15803d;
}

.status-chip--disabled {
  background: rgba(71, 85, 105, 0.12);
  color: #475569;
}

.status-chip--expired {
  background: rgba(234, 88, 12, 0.12);
  color: #c2410c;
}

.status-chip--sync {
  background: rgba(22, 119, 255, 0.12);
  color: #1677ff;
}

.status-chip--warning {
  background: rgba(250, 173, 20, 0.16);
  color: #ad6800;
}

.permission-summary {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
}

.summary-card {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 16px 18px;
  border: 1px solid #e8edf5;
  border-radius: 16px;
  background: #fff;
}

.summary-card__label {
  color: #667085;
  font-size: 12px;
  line-height: 18px;
}

.summary-card__value {
  color: #1f2329;
  font-size: 14px;
  font-weight: 600;
  line-height: 22px;
}

.permission-action,
.permission-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.permission-action {
  padding: 18px 20px;
  border: 1px solid #e8edf5;
  border-radius: 18px;
  background: #fff;
}

.permission-action__select {
  width: 340px;
}

.permission-toolbar__search {
  width: 320px;
}

.permission-toolbar__actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.permission-toolbar__link {
  padding: 0;
}

.permission-panels {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px;
}

.permission-panel {
  display: flex;
  flex-direction: column;
  min-height: 520px;
  border: 1px solid #e8edf5;
  border-radius: 18px;
  background: #fff;
  overflow: hidden;
  box-shadow: 0 12px 28px rgba(15, 23, 42, 0.04);
}

.permission-panel__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 18px 20px 12px;
  border-bottom: 1px solid #edf1f7;
}

.permission-panel__title {
  color: #1f2329;
  font-size: 15px;
  font-weight: 700;
  line-height: 24px;
}

.permission-panel__meta {
  color: #667085;
  font-size: 12px;
  line-height: 18px;
}

.permission-panel__actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.permission-panel__reset {
  padding: 0;
}

.permission-panel__body {
  flex: 1;
  min-height: 0;
  padding: 12px 14px 18px;
  overflow: auto;
}

.permission-panel__body :deep(.ant-tree-checkbox) {
  margin-block-start: 0;
}

.permission-panel__body :deep(.ant-tree) {
  padding: 6px 6px 0;
}

.permission-panel__body :deep(.ant-tree-treenode) {
  padding: 4px 0;
}

.permission-panel__body :deep(.ant-tree-node-content-wrapper) {
  min-height: 34px;
  border-radius: 10px;
}

.permission-panel__body :deep(.ant-tree-checkbox + span) {
  color: #1f2329;
}

.permission-panel__body :deep(.ant-tree-checkbox-disabled + span) {
  color: #b0b8c4;
}

.tree-readonly {
  pointer-events: none;
}

.close-btn {
  width: 40px;
  height: 40px;
  color: #1f2329;
  font-size: 22px;
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

@media (max-width: 1180px) {
  .permission-summary {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .permission-panels {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 860px) {
  .permission-overview,
  .permission-action,
  .permission-toolbar {
    flex-direction: column;
    align-items: stretch;
  }

  .permission-action__select,
  .permission-toolbar__search {
    width: 100%;
  }

  .permission-summary {
    grid-template-columns: 1fr;
  }
}
</style>
