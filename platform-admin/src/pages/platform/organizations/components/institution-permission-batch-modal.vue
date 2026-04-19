<script setup lang="ts">
import type { AxiosError } from 'axios'
import {
  DownOutlined,
  SearchOutlined,
  UpOutlined,
} from '@ant-design/icons-vue'
import { Empty } from 'ant-design-vue'
import { computed, ref, watch } from 'vue'
import type { InstitutionPermissionDetail } from '@/api/platform/institutions'
import type { MenuTreeNode } from '@/api/platform/versions'
import PlatformModalShell from '../../shared/platform-modal-shell.vue'
import {
  getInstitutionPermissionDetailApi,
  replaceInstitutionPermissionVersionApi,
} from '@/api/platform/institutions'
import { getVersionDetailApi } from '@/api/platform/versions'
import {
  type AuthorityGroup,
  useRolePermissions,
} from '@/composables/useRolePermissions'
import messageService from '@/utils/messageService'

const props = defineProps<{
  open: boolean
  institutionIds?: number[]
}>()

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'saved'): void
}>()

const simpleImage = Empty.PRESENTED_IMAGE_SIMPLE
const initialPermissionData: AuthorityGroup[] = []
type BatchApplyMode = 'add-only' | 'remove-only'

interface BatchTargetInstitution {
  institutionId: number
  organName: string
  ownedCount: number
  missingCount: number
  totalCount: number
}

const openModal = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

const loading = ref(false)
const treeLoading = ref(false)
const submitting = ref(false)
const details = ref<InstitutionPermissionDetail[]>([])
const moduleId = ref<number>(0)
const moduleName = ref('')
const scopedMenuIds = ref<number[]>([])
const permissionOwnedCountMap = ref<Record<number, number>>({})
const mismatchMessage = ref('')
const applyMode = ref<BatchApplyMode>('add-only')
const targetInstitutions = ref<BatchTargetInstitution[]>([])
const checkedTargetInstitutionIds = ref<number[]>([])
const hasQueriedTargets = ref(false)

const {
  boxList,
  searchValue,
  filteredBoxList,
  isAllExpanded,
  isParentExpanded,
  isChildExpanded,
  toggleAllExpand,
  expandAllChildren,
  collapseAllChildren,
  toggleChildExpand,
  getFilteredChildren,
  getFilteredchildren,
  highlightText,
  handleParentChange,
  handleChildChange,
  handleAuthorityChange,
  clearAllSelected,
  updateData,
} = useRolePermissions(initialPermissionData, treeLoading, {
  enforceGroupExclusive: false,
})

const normalizedInstitutionIds = computed(() => normalizeInstitutionIDList(props.institutionIds || []))
const selectedInstitutionCount = computed(() => normalizedInstitutionIds.value.length)

const selectedInstitutionNames = computed(() => {
  const names = details.value
    .map(item => String(item.organName || '').trim())
    .filter(Boolean)
  return Array.from(new Set(names))
})

const selectedInstitutionNamePreview = computed(() => selectedInstitutionNames.value.slice(0, 8))
const hasMoreNames = computed(() => selectedInstitutionNames.value.length > selectedInstitutionNamePreview.value.length)

const currentModuleName = computed(() => {
  const normalized = String(moduleName.value || '').trim()
  if (normalized)
    return normalized
  return '--'
})

const totalLeafCount = computed(() => {
  return boxList.value.reduce((sum, parent) => {
    return sum + parent.children.reduce((childSum, child) => childSum + child.children.length, 0)
  }, 0)
})

const selectedLeafCount = computed(() => {
  return boxList.value.reduce((sum, parent) => {
    return sum + parent.children.reduce((childSum, child) => {
      return childSum + child.children.filter(authority => authority.checked).length
    }, 0)
  }, 0)
})

const hasSelectedPermissions = computed(() => {
  return boxList.value.some(parent =>
    parent.checked
    || parent.indeterminate
    || parent.children.some(child =>
      child.checked
      || child.indeterminate
      || child.children.some(authority => authority.checked),
    ),
  )
})

const partialOwnedPermissionCount = computed(() => {
  const total = selectedInstitutionCount.value
  if (!total)
    return 0
  return scopedMenuIds.value.filter((menuID) => {
    const owned = Number(permissionOwnedCountMap.value[menuID] || 0)
    return owned > 0 && owned < total
  }).length
})

const selectedActionMenuIds = computed(() => {
  const visibleAuthorityIDs = collectVisibleAuthorityIDs()
  return filterScopedMenuIDs(collectSelectedPermissionIDs(), visibleAuthorityIDs)
})

const selectedActionMenuSignature = computed(() => selectedActionMenuIds.value.join(','))
const matchedInstitutionCount = computed(() => targetInstitutions.value.length)
const checkedTargetCount = computed(() => normalizeInstitutionIDList(checkedTargetInstitutionIds.value).length)

const applyModeLabel = computed(() => {
  return applyMode.value === 'add-only' ? '仅新增勾选权限' : '仅移除勾选权限'
})

const queryRuleLabel = computed(() => {
  return applyMode.value === 'add-only'
    ? '筛出至少缺少 1 项所选权限的机构'
    : '筛出至少拥有 1 项所选权限的机构'
})

const canQueryTargets = computed(() => {
  return !mismatchMessage.value && selectedActionMenuIds.value.length > 0 && !loading.value && !treeLoading.value
})

const queryButtonText = computed(() => {
  return hasQueriedTargets.value ? '重新查询机构' : '查询命中机构'
})

const targetEmptyDescription = computed(() => {
  if (!hasSelectedPermissions.value)
    return '请先在左侧勾选要处理的权限'
  if (!hasQueriedTargets.value)
    return '点击“查询命中机构”后，这里会显示可操作机构'
  return `暂无机构满足条件：${queryRuleLabel.value}`
})

function closeModal() {
  emit('update:open', false)
}

function resetTargetQuery() {
  targetInstitutions.value = []
  checkedTargetInstitutionIds.value = []
  hasQueriedTargets.value = false
}

function resetState() {
  details.value = []
  moduleId.value = 0
  moduleName.value = ''
  scopedMenuIds.value = []
  permissionOwnedCountMap.value = {}
  mismatchMessage.value = ''
  applyMode.value = 'add-only'
  searchValue.value = ''
  resetTargetQuery()
  updateData([])
  clearAllSelected()
}

function normalizeInstitutionIDList(values: Array<number | string>) {
  const seen = new Set<number>()
  const normalized: number[] = []
  for (const item of values) {
    const value = Number(item)
    if (!Number.isFinite(value) || value <= 0 || seen.has(value))
      continue
    seen.add(value)
    normalized.push(value)
  }
  return normalized
}

function resolveRequestErrorMessage(error: unknown, fallback: string) {
  const axiosError = error as AxiosError<{ message?: string }>
  return axiosError?.response?.data?.message || (error as any)?.message || fallback
}

function normalizeMenuID(value: string | number | undefined, fallbackValue = 0) {
  const parsed = Number(value)
  if (Number.isFinite(parsed) && parsed > 0)
    return parsed
  return fallbackValue
}

function buildAuthority(node: MenuTreeNode, fallbackID: number) {
  return {
    id: normalizeMenuID(node.menuId ?? node.id, fallbackID),
    name: String(node.menuName || ''),
    remark: String(node.introduce || ''),
    type: Number(node.menuType || 1),
    mode: 0,
    groupCode: String(node.groupCode || ''),
    weight: Number(node.weight || 0),
    checked: false,
  }
}

function buildAuthorityChild(node: MenuTreeNode, groupIndex: number, childIndex: number) {
  const childID = normalizeMenuID(node.menuId ?? node.id, groupIndex * 1000 + childIndex + 1)
  const rawAuthorities = Array.isArray(node.children) && node.children.length > 0
    ? node.children
    : [node]

  return {
    id: String(childID),
    menuName: String(node.menuName || ''),
    checked: false,
    indeterminate: false,
    children: rawAuthorities.map((authorityNode, authorityIndex) =>
      buildAuthority(authorityNode, childID * 100 + authorityIndex + 1),
    ),
  }
}

function mapMenuTreeToPermissionGroups(nodes: MenuTreeNode[] = []) {
  return nodes.map((groupNode, groupIndex) => {
    const groupID = normalizeMenuID(groupNode.menuId ?? groupNode.id, groupIndex + 1)
    const rawChildren = Array.isArray(groupNode.children) && groupNode.children.length > 0
      ? groupNode.children
      : [groupNode]

    return {
      id: String(groupID),
      menuName: String(groupNode.menuName || ''),
      checked: false,
      indeterminate: false,
      children: rawChildren.map((childNode, childIndex) =>
        buildAuthorityChild(childNode, groupIndex + 1, childIndex),
      ),
    }
  })
}

function normalizeMenuIDList(values: number[] = []) {
  return Array.from(new Set(
    values
      .map(item => Number(item))
      .filter(item => Number.isFinite(item) && item > 0),
  )).sort((a, b) => a - b)
}

function isSameMenuSelection(left: number[] = [], right: number[] = []) {
  const leftList = normalizeMenuIDList(left)
  const rightList = normalizeMenuIDList(right)
  if (leftList.length !== rightList.length)
    return false
  return leftList.every((item, index) => item === rightList[index])
}

function collectSelectedPermissionIDs() {
  const selectedIDs = new Set<number>()

  boxList.value.forEach((parent) => {
    parent.children.forEach((child) => {
      child.children.forEach((authority) => {
        if (authority.checked && Number(authority.id) > 0)
          selectedIDs.add(Number(authority.id))
      })
    })
  })

  return Array.from(selectedIDs)
}

function collectVisibleAuthorityIDs() {
  const ids = new Set<number>()

  boxList.value.forEach((parent) => {
    parent.children.forEach((child) => {
      child.children.forEach((authority) => {
        if (Number(authority.id) > 0)
          ids.add(Number(authority.id))
      })
    })
  })

  return Array.from(ids)
}

function filterScopedMenuIDs(values: number[] = [], allowedValues: number[] = []) {
  const allowedSet = new Set(normalizeMenuIDList(allowedValues))
  return normalizeMenuIDList(values).filter(item => allowedSet.has(item))
}

function collectScopedEffectiveMenuIDs(detail: InstitutionPermissionDetail, allowedValues: number[]) {
  return filterScopedMenuIDs(
    Array.isArray(detail.effectiveMenuIds)
      ? detail.effectiveMenuIds.map(item => Number(item))
      : [],
    allowedValues,
  )
}

function buildPermissionOwnedCountMap(detailList: InstitutionPermissionDetail[], allowedValues: number[]) {
  const counter: Record<number, number> = {}
  detailList.forEach((detail) => {
    const scopedEffectiveMenuIDs = collectScopedEffectiveMenuIDs(detail, allowedValues)
    scopedEffectiveMenuIDs.forEach((menuID) => {
      counter[menuID] = Number(counter[menuID] || 0) + 1
    })
  })
  return counter
}

function getPermissionOwnedMeta(menuID: number) {
  const total = selectedInstitutionCount.value
  const owned = Number(permissionOwnedCountMap.value[Number(menuID) || 0] || 0)
  if (!total) {
    return {
      text: '--',
      className: 'permission-authority__meta--none',
    }
  }

  if (owned <= 0) {
    return {
      text: `0/${total} 家机构拥有`,
      className: 'permission-authority__meta--none',
    }
  }
  if (owned >= total) {
    return {
      text: `全部 ${total} 家机构已拥有`,
      className: 'permission-authority__meta--all',
    }
  }
  return {
    text: `${owned}/${total} 家机构已拥有`,
    className: 'permission-authority__meta--partial',
  }
}

function getTargetActionMeta(item: BatchTargetInstitution) {
  if (applyMode.value === 'add-only') {
    return {
      text: `将新增 ${item.missingCount} 项`,
      className: 'target-item__action--add',
    }
  }

  return {
    text: `将移除 ${item.ownedCount} 项`,
    className: 'target-item__action--remove',
  }
}

function isTargetChecked(institutionId: number) {
  return checkedTargetInstitutionIds.value.includes(institutionId)
}

async function hydratePermissionSelection(currentModuleId: number) {
  if (!currentModuleId) {
    updateData([])
    scopedMenuIds.value = []
    return
  }

  treeLoading.value = true
  try {
    const res = await getVersionDetailApi({ moduleId: currentModuleId })
    if (res.code !== 200 || !res.result) {
      messageService.error(res.message || '获取版本权限失败')
      updateData([])
      scopedMenuIds.value = []
      return
    }

    const versionScopedMenuIds = normalizeMenuIDList(
      Array.isArray(res.result.selectedMenuIds)
        ? res.result.selectedMenuIds.map(item => Number(item))
        : [],
    )
    const scopedTree = Array.isArray(res.result.menuIds) ? res.result.menuIds : []
    scopedMenuIds.value = versionScopedMenuIds
    updateData(mapMenuTreeToPermissionGroups(scopedTree))
  }
  catch (error: any) {
    messageService.error(resolveRequestErrorMessage(error, '获取版本权限失败'))
    updateData([])
    scopedMenuIds.value = []
  }
  finally {
    treeLoading.value = false
  }
}

function resolveModuleMismatchMessage(detailList: InstitutionPermissionDetail[]) {
  const uniqueModulePairs = Array.from(new Set(
    detailList
      .map(item => `${Number(item.currentModuleId || 0)}::${String(item.currentModuleName || '').trim() || '--'}`)
      .filter(Boolean),
  ))
  if (uniqueModulePairs.length <= 1)
    return ''

  const readable = uniqueModulePairs
    .map((item) => {
      const [id, name] = item.split('::')
      return `${name}(ID:${id})`
    })
    .join('、')
  return `所选机构存在多个版本：${readable}。请先按同版本筛选后再批量配置。`
}

function buildTargetInstitutions(menuIds: number[], allowedValues: number[]) {
  const list: BatchTargetInstitution[] = []

  details.value.forEach((detail) => {
    const institutionId = Number(detail.institutionId || 0)
    if (!institutionId)
      return

    const institutionName = String(detail.organName || `机构${institutionId}`).trim() || `机构${institutionId}`
    const effectiveSet = new Set(collectScopedEffectiveMenuIDs(detail, allowedValues))
    let ownedCount = 0

    menuIds.forEach((menuID) => {
      if (effectiveSet.has(menuID))
        ownedCount++
    })

    const totalCount = menuIds.length
    const missingCount = totalCount - ownedCount
    const shouldInclude = applyMode.value === 'add-only' ? missingCount > 0 : ownedCount > 0
    if (!shouldInclude)
      return

    list.push({
      institutionId,
      organName: institutionName,
      ownedCount,
      missingCount,
      totalCount,
    })
  })

  return list
}

function queryMatchedInstitutions() {
  if (mismatchMessage.value) {
    messageService.warning('请先按同版本筛选机构后再批量配置')
    return
  }

  const menuIds = selectedActionMenuIds.value
  if (!menuIds.length) {
    messageService.warning('请先勾选要处理的权限')
    return
  }

  const allowedValues = collectVisibleAuthorityIDs()
  const matched = buildTargetInstitutions(menuIds, allowedValues)
  targetInstitutions.value = matched
  checkedTargetInstitutionIds.value = matched.map(item => item.institutionId)
  hasQueriedTargets.value = true

  if (!matched.length) {
    messageService.warning('没有筛到可操作机构，请调整权限或批量模式后重试')
    return
  }

  messageService.success(`已筛出 ${matched.length} 家机构，默认已全选`)
}

function selectAllMatchedTargets() {
  checkedTargetInstitutionIds.value = targetInstitutions.value.map(item => item.institutionId)
}

function clearMatchedTargets() {
  checkedTargetInstitutionIds.value = []
}

async function loadData() {
  const institutionIds = normalizedInstitutionIds.value
  if (!institutionIds.length) {
    resetState()
    return
  }

  loading.value = true
  mismatchMessage.value = ''
  try {
    const responses = await Promise.all(
      institutionIds.map(institutionId => getInstitutionPermissionDetailApi({ institutionId })),
    )
    const failed = responses.find(item => item.code !== 200 || !item.result)
    if (failed) {
      messageService.error(failed.message || '加载机构权限信息失败')
      return
    }

    const detailList = responses.map(item => item.result as InstitutionPermissionDetail)
    details.value = detailList

    const moduleMismatch = resolveModuleMismatchMessage(detailList)
    if (moduleMismatch) {
      mismatchMessage.value = moduleMismatch
      updateData([])
      scopedMenuIds.value = []
      permissionOwnedCountMap.value = {}
      resetTargetQuery()
      return
    }

    const resolvedModuleId = Number(detailList[0]?.currentModuleId || 0)
    if (!resolvedModuleId) {
      mismatchMessage.value = '所选机构尚未绑定版本，无法批量配置权限。'
      updateData([])
      scopedMenuIds.value = []
      permissionOwnedCountMap.value = {}
      resetTargetQuery()
      return
    }

    moduleId.value = resolvedModuleId
    moduleName.value = String(detailList[0]?.currentModuleName || '').trim()
    await hydratePermissionSelection(resolvedModuleId)
    permissionOwnedCountMap.value = buildPermissionOwnedCountMap(detailList, scopedMenuIds.value)
    clearAllSelected()
    resetTargetQuery()
  }
  catch (error: any) {
    messageService.error(resolveRequestErrorMessage(error, '加载机构权限信息失败'))
  }
  finally {
    loading.value = false
  }
}

async function submitBatchPermissionConfig() {
  if (mismatchMessage.value) {
    messageService.warning('请先按同版本筛选机构后再批量配置')
    return
  }

  const institutionIds = normalizedInstitutionIds.value
  const currentModuleId = Number(moduleId.value || 0)
  if (!institutionIds.length || !currentModuleId)
    return

  if (!hasQueriedTargets.value) {
    messageService.warning('请先查询命中机构，再勾选目标机构后提交')
    return
  }

  const targetIDs = normalizeInstitutionIDList(checkedTargetInstitutionIds.value)
  if (!targetIDs.length) {
    messageService.warning('请至少选择一家目标机构')
    return
  }

  const actionMenuIds = selectedActionMenuIds.value
  if (!actionMenuIds.length) {
    messageService.warning('请先勾选要处理的权限')
    return
  }

  const visibleAuthorityIDs = collectVisibleAuthorityIDs()
  const targetSet = new Set(targetIDs)

  submitting.value = true
  try {
    let updatedCount = 0
    let skippedCount = 0

    for (const detail of details.value) {
      const institutionId = Number(detail.institutionId || 0)
      if (!targetSet.has(institutionId))
        continue

      const currentEffectiveMenuIds = collectScopedEffectiveMenuIDs(detail, visibleAuthorityIDs)
      const merged = new Set(currentEffectiveMenuIds)

      if (applyMode.value === 'add-only')
        actionMenuIds.forEach(menuID => merged.add(menuID))
      else
        actionMenuIds.forEach(menuID => merged.delete(menuID))

      const finalMenuIds = normalizeMenuIDList(Array.from(merged))
      if (isSameMenuSelection(finalMenuIds, currentEffectiveMenuIds)) {
        skippedCount++
        continue
      }

      const res = await replaceInstitutionPermissionVersionApi({
        institutionId,
        moduleId: currentModuleId,
        menuIds: finalMenuIds,
      })
      if (res.code !== 200) {
        const institutionName = String(detail.organName || `机构${institutionId}`).trim()
        messageService.error(`${institutionName} 保存失败：${res.message || '机构权限保存失败'}`)
        return
      }

      detail.effectiveMenuIds = finalMenuIds
      updatedCount++
    }

    if (!updatedCount) {
      messageService.warning('目标机构权限无需变更')
      return
    }

    permissionOwnedCountMap.value = buildPermissionOwnedCountMap(details.value, scopedMenuIds.value)
    const summary = skippedCount > 0
      ? `已更新 ${updatedCount} 家机构权限，${skippedCount} 家无需变更`
      : `已更新 ${updatedCount} 家机构权限`
    messageService.success(summary)
    emit('saved')
    closeModal()
  }
  catch (error: any) {
    messageService.error(resolveRequestErrorMessage(error, '批量机构权限保存失败'))
  }
  finally {
    submitting.value = false
  }
}

watch(
  () => [props.open, normalizedInstitutionIds.value.join(',')] as const,
  ([open]) => {
    if (!open) {
      resetState()
      return
    }
    void loadData()
  },
  { immediate: true },
)

watch(
  () => [applyMode.value, selectedActionMenuSignature.value] as const,
  () => {
    if (!props.open)
      return
    resetTargetQuery()
  },
)
</script>

<template>
  <PlatformModalShell
    v-model:open="openModal"
    :width="1320"
    title="批量机构权限配置"
    modal-class="institution-permission-batch-modal"
    scrollable
  >
    <a-spin :spinning="loading">
      <div class="batch-modal">
        <div class="batch-overview">
          <div class="batch-overview__head">
            <div class="batch-overview__title">
              已选机构 {{ selectedInstitutionCount }} 家
            </div>
            <div class="batch-overview__module">
              当前版本：{{ currentModuleName }}
            </div>
          </div>
          <div class="batch-overview__desc">
            先勾选权限，再点击“查询命中机构”，最后选择目标机构并保存。系统只会处理你勾选的权限，不会覆盖其它权限。
          </div>
          <div class="batch-overview__steps">
            <div class="batch-step" :class="{ 'batch-step--active': hasSelectedPermissions }">
              <span class="batch-step__index">1</span>
              <span class="batch-step__text">勾选权限</span>
            </div>
            <div class="batch-step" :class="{ 'batch-step--active': hasQueriedTargets }">
              <span class="batch-step__index">2</span>
              <span class="batch-step__text">查询命中机构</span>
            </div>
            <div class="batch-step" :class="{ 'batch-step--active': checkedTargetCount > 0 }">
              <span class="batch-step__index">3</span>
              <span class="batch-step__text">确认并批量保存</span>
            </div>
          </div>
          <div class="batch-overview__mode">
            <div class="batch-overview__mode-left">
              <span class="batch-overview__mode-label">批量模式</span>
              <a-radio-group v-model:value="applyMode" size="small">
                <a-radio-button value="add-only">
                  仅新增勾选
                </a-radio-button>
                <a-radio-button value="remove-only">
                  仅移除勾选
                </a-radio-button>
              </a-radio-group>
            </div>
            <a-button
              type="primary"
              size="small"
              class="batch-overview__query-btn"
              :disabled="!canQueryTargets"
              @click="queryMatchedInstitutions"
            >
              {{ queryButtonText }}
            </a-button>
          </div>
          <div class="batch-overview__metrics">
            <div class="overview-metric">
              <div class="overview-metric__label">
                当前模式
              </div>
              <div class="overview-metric__value">
                {{ applyModeLabel }}
              </div>
            </div>
            <div class="overview-metric">
              <div class="overview-metric__label">
                已勾选权限
              </div>
              <div class="overview-metric__value">
                {{ selectedLeafCount }} 项
              </div>
            </div>
            <div class="overview-metric">
              <div class="overview-metric__label">
                命中机构
              </div>
              <div class="overview-metric__value">
                {{ matchedInstitutionCount }} 家
              </div>
            </div>
            <div class="overview-metric">
              <div class="overview-metric__label">
                待执行机构
              </div>
              <div class="overview-metric__value">
                {{ checkedTargetCount }} 家
              </div>
            </div>
          </div>
          <div v-if="partialOwnedPermissionCount > 0" class="batch-overview__tip">
            当前有 {{ partialOwnedPermissionCount }} 项权限在所选机构里并不一致，请结合右侧筛选结果确认再提交。
          </div>
          <div class="batch-overview__tags-head">
            <span class="batch-overview__tags-title">已选机构池</span>
            <span v-if="hasMoreNames" class="batch-overview__more">
              等 {{ selectedInstitutionNames.length }} 家机构
            </span>
          </div>
          <div class="batch-overview__tags">
            <a-tag v-for="name in selectedInstitutionNamePreview" :key="name" class="batch-overview__tag">
              {{ name }}
            </a-tag>
          </div>
        </div>

        <a-alert
          v-if="mismatchMessage"
          type="warning"
          show-icon
          :message="mismatchMessage"
        />

        <div v-else class="batch-content">
          <div class="permission-panel">
            <div class="permission-panel__header">
              <div class="permission-panel__titlebox">
                <div class="permission-panel__title">
                  权限筛选
                </div>
                <div class="permission-panel__meta">
                  <span>仅在当前版本范围内调整</span>
                  <span>右侧会按模式筛出可操作机构</span>
                </div>
              </div>

              <a-button type="link" class="permission-panel__link" @click="toggleAllExpand">
                {{ isAllExpanded ? '一键收起' : '一键展开' }}
                <component :is="isAllExpanded ? UpOutlined : DownOutlined" />
              </a-button>
            </div>

            <div class="permission-panel__toolbar">
              <div class="permission-panel__stats">
                已选 {{ selectedLeafCount }} / {{ totalLeafCount }}，部分机构不一致 {{ partialOwnedPermissionCount }} 项
              </div>
              <div class="permission-panel__tools">
                <a-button type="link" :disabled="!hasSelectedPermissions" @click="clearAllSelected">
                  清空已选
                </a-button>
                <a-input
                  v-model:value="searchValue"
                  allow-clear
                  placeholder="搜索权限点名称或权限描述"
                >
                  <template #prefix>
                    <SearchOutlined />
                  </template>
                </a-input>
              </div>
            </div>

            <div class="permission-panel__body">
              <a-spin :spinning="treeLoading" tip="加载中...">
                <div class="permission-box">
                  <template v-for="(item, index) in filteredBoxList" :key="item.id">
                    <div class="permission-row permission-row--group" :class="{ 'permission-row--last': index === filteredBoxList.length - 1 }">
                      <div class="permission-row__main">
                        <a-checkbox
                          v-model:checked="item.checked"
                          :indeterminate="item.indeterminate"
                          @change="() => handleParentChange(item)"
                        />
                        <span class="permission-row__title permission-row__title--group" v-html="highlightText(item.menuName, searchValue)" />
                      </div>
                      <span class="permission-row__toggle" @click="isParentExpanded(item.id) ? collapseAllChildren(item.id) : expandAllChildren(item.id)">
                        {{ isParentExpanded(item.id) ? '收起全部' : '展开全部' }}
                      </span>
                    </div>

                    <template v-if="isParentExpanded(item.id)">
                      <div v-for="child in getFilteredChildren(item)" :key="child.id">
                        <div class="permission-row permission-row--child">
                          <div class="permission-row__main">
                            <a-checkbox
                              v-model:checked="child.checked"
                              :indeterminate="child.indeterminate"
                              @change="() => handleChildChange(child, item)"
                            />
                            <span class="permission-row__title permission-row__title--child" v-html="highlightText(child.menuName, searchValue)" />
                          </div>
                          <span class="permission-row__toggle" @click="toggleChildExpand(child.id)">
                            {{ isChildExpanded(child.id) ? '收起' : '展开' }}
                          </span>
                        </div>

                        <template v-if="isChildExpanded(child.id)">
                          <div
                            v-for="(authority, authorityIndex) in getFilteredchildren(child, item)"
                            :key="authority.id"
                            class="permission-row permission-row--authority"
                            :class="{ 'permission-row--last-leaf': authorityIndex === getFilteredchildren(child, item).length - 1 }"
                          >
                            <div class="permission-row__main permission-row__main--authority">
                              <a-checkbox
                                v-model:checked="authority.checked"
                                @change="() => handleAuthorityChange(authority, child, item)"
                              />
                              <div class="permission-authority">
                                <span class="permission-authority__title" v-html="highlightText(authority.name, searchValue)" />
                                <span class="permission-authority__meta" :class="getPermissionOwnedMeta(Number(authority.id)).className">
                                  {{ getPermissionOwnedMeta(Number(authority.id)).text }}
                                </span>
                                <span
                                  v-if="authority.remark"
                                  class="permission-authority__desc"
                                  v-html="highlightText(authority.remark, searchValue)"
                                />
                              </div>
                            </div>
                          </div>
                        </template>
                      </div>
                    </template>
                  </template>
                </div>

                <a-empty v-if="filteredBoxList.length === 0" :image="simpleImage" description="暂无匹配权限" />
              </a-spin>
            </div>
          </div>

          <div class="target-panel">
            <div class="target-panel__header">
              <div class="target-panel__titlebox">
                <div class="target-panel__title">
                  命中机构
                </div>
                <div class="target-panel__meta">
                  {{ queryRuleLabel }}
                </div>
                <div class="target-panel__summary">
                  <span class="target-panel__summary-item">命中 {{ matchedInstitutionCount }} 家</span>
                  <span class="target-panel__summary-item">待执行 {{ checkedTargetCount }} 家</span>
                </div>
              </div>
            </div>

            <div class="target-panel__toolbar">
              <a-button type="link" :disabled="!matchedInstitutionCount" @click="selectAllMatchedTargets">
                全选命中
              </a-button>
              <a-button type="link" :disabled="!checkedTargetCount" @click="clearMatchedTargets">
                清空勾选
              </a-button>
            </div>

            <div class="target-panel__body">
              <a-empty
                v-if="!matchedInstitutionCount"
                :image="simpleImage"
                :description="targetEmptyDescription"
              />

              <a-checkbox-group v-else v-model:value="checkedTargetInstitutionIds" class="target-list">
                <div
                  v-for="item in targetInstitutions"
                  :key="item.institutionId"
                  class="target-item"
                  :class="{ 'target-item--checked': isTargetChecked(item.institutionId) }"
                >
                  <a-checkbox :value="item.institutionId" class="target-item__checkbox">
                    <div class="target-item__content">
                      <div class="target-item__name">
                        {{ item.organName }}
                      </div>
                      <div class="target-item__desc">
                        已拥有 {{ item.ownedCount }} / {{ item.totalCount }} 项
                        <span class="target-item__action" :class="getTargetActionMeta(item).className">
                          {{ getTargetActionMeta(item).text }}
                        </span>
                      </div>
                    </div>
                  </a-checkbox>
                </div>
              </a-checkbox-group>
            </div>
          </div>
        </div>
      </div>
    </a-spin>

    <template #footer>
      <a-button @click="closeModal">
        关闭
      </a-button>
      <a-button
        type="primary"
        :loading="submitting"
        :disabled="!!mismatchMessage || !hasQueriedTargets || !checkedTargetCount"
        @click="submitBatchPermissionConfig"
      >
        批量保存
      </a-button>
    </template>
  </PlatformModalShell>
</template>

<style scoped lang="less">
.batch-modal {
  --panel-border: #e6ebf2;
  --panel-border-soft: #edf1f6;
  --text-primary: #1f2329;
  --text-secondary: #566074;
  --text-tertiary: #8b95a8;

  display: flex;
  flex-direction: column;
  gap: 14px;
  padding-top: 4px;
}

.batch-overview {
  padding: 16px 18px;
  border: 1px solid var(--panel-border);
  border-radius: 14px;
  background: #fff;
  box-shadow: 0 4px 14px rgb(12 30 68 / 4%);
}

.batch-overview__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.batch-overview__title {
  color: var(--text-primary);
  font-size: 16px;
  font-weight: 600;
  line-height: 24px;
}

.batch-overview__module {
  padding: 3px 10px;
  border: 1px solid #d2e7ff;
  border-radius: 999px;
  background: #f5faff;
  color: #0f6ad8;
  font-size: 12px;
  font-weight: 600;
  line-height: 18px;
}

.batch-overview__desc {
  margin-top: 8px;
  color: var(--text-secondary);
  font-size: 12px;
  line-height: 18px;
}

.batch-overview__steps {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10px;
  margin-top: 12px;
}

.batch-step {
  display: flex;
  align-items: center;
  gap: 8px;
  min-height: 38px;
  padding: 0 10px;
  border: 1px dashed #dbe2eb;
  border-radius: 10px;
  background: #fafcff;
}

.batch-step__index {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 18px;
  height: 18px;
  border-radius: 999px;
  background: #e9edf3;
  color: #768195;
  font-size: 11px;
  font-weight: 600;
  line-height: 1;
}

.batch-step__text {
  color: #6b7587;
  font-size: 12px;
  font-weight: 500;
  line-height: 18px;
}

.batch-step--active {
  border-style: solid;
  border-color: #b5d7ff;
  background: #f4f9ff;
}

.batch-step--active .batch-step__index {
  background: #1677ff;
  color: #fff;
}

.batch-step--active .batch-step__text {
  color: #0f6ad8;
}

.batch-overview__mode {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-top: 12px;
}

.batch-overview__mode-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.batch-overview__mode-label {
  color: var(--text-secondary);
  font-size: 12px;
  line-height: 18px;
}

.batch-overview__query-btn {
  min-width: 116px;
}

.batch-overview__metrics {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 10px;
  margin-top: 12px;
}

.overview-metric {
  min-width: 0;
  padding: 8px 10px;
  border: 1px solid #edf1f6;
  border-radius: 10px;
  background: #fcfdff;
}

.overview-metric__label {
  color: var(--text-tertiary);
  font-size: 11px;
  line-height: 16px;
}

.overview-metric__value {
  margin-top: 3px;
  color: var(--text-primary);
  font-size: 13px;
  font-weight: 600;
  line-height: 19px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.batch-overview__tip {
  margin-top: 10px;
  padding: 6px 10px;
  border-radius: 8px;
  background: #fff8ee;
  color: #b45309;
  font-size: 12px;
  line-height: 18px;
}

.batch-overview__tags-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  margin-top: 10px;
}

.batch-overview__tags-title {
  color: var(--text-secondary);
  font-size: 12px;
  font-weight: 500;
  line-height: 18px;
}

.batch-overview__tags {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 6px;
}

.batch-overview__tag {
  margin: 0;
  border: 1px solid #e4ebf5;
  border-radius: 999px;
  background: #f8fbff;
  color: #4d5a71;
}

.batch-overview__more {
  color: var(--text-tertiary);
  font-size: 12px;
}

.batch-content {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 376px;
  gap: 14px;
  min-height: 560px;
}

.permission-panel,
.target-panel {
  display: flex;
  flex-direction: column;
  min-width: 0;
  border: 1px solid var(--panel-border);
  border-radius: 14px;
  background: #fff;
  overflow: hidden;
  box-shadow: 0 2px 10px rgb(12 30 68 / 3%);
}

.permission-panel__header,
.target-panel__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 12px 16px 10px;
  border-bottom: 1px solid var(--panel-border-soft);
  background: #fcfdff;
}

.permission-panel__titlebox,
.target-panel__titlebox {
  min-width: 0;
}

.permission-panel__title,
.target-panel__title {
  color: var(--text-primary);
  font-size: 14px;
  font-weight: 600;
  line-height: 20px;
}

.permission-panel__meta,
.target-panel__meta {
  margin-top: 3px;
  color: var(--text-tertiary);
  font-size: 12px;
  line-height: 16px;
}

.permission-panel__meta {
  display: flex;
  align-items: center;
  gap: 12px;
}

.permission-panel__link {
  flex-shrink: 0;
  padding-inline: 0;
  height: 24px;
  color: #4b628b;
}

.permission-panel__toolbar,
.target-panel__toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 8px 16px;
  border-bottom: 1px solid var(--panel-border-soft);
}

.permission-panel__stats {
  color: var(--text-secondary);
  font-size: 12px;
  line-height: 20px;
}

.permission-panel__tools {
  display: flex;
  align-items: center;
  gap: 10px;
}

.permission-panel__tools :deep(.ant-input-affix-wrapper) {
  width: 320px;
}

.permission-panel__body,
.target-panel__body {
  flex: 1;
  min-height: 0;
  overflow: auto;
  padding: 10px 14px 12px;
}

.permission-box {
  display: flex;
  flex-direction: column;
}

.permission-row {
  padding-right: 6px;
  border-bottom: 1px dashed #eef2f7;
  transition: background-color 0.2s ease;
}

.permission-row:hover {
  background: #fafcff;
}

.permission-row--group {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  min-height: 44px;
  padding-left: 2px;
}

.permission-row--child {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  min-height: 38px;
  padding-left: 28px;
}

.permission-row--authority {
  min-height: 34px;
  padding: 6px 0 6px 56px;
}

.permission-row--last-leaf,
.permission-row--last {
  border-bottom: none;
}

.permission-row__main {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}

.permission-row__main--authority {
  align-items: flex-start;
}

.permission-row__title {
  min-width: 0;
  color: var(--text-primary);
  font-size: 13px;
  line-height: 20px;
}

.permission-row__title--group {
  font-weight: 600;
}

.permission-row__title--child {
  font-weight: 500;
}

.permission-row__toggle {
  flex-shrink: 0;
  color: #4b628b;
  font-size: 12px;
  line-height: 20px;
  cursor: pointer;
}

.permission-authority {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.permission-authority__title {
  color: #2c3445;
  font-size: 13px;
  line-height: 20px;
}

.permission-authority__meta {
  margin-top: 2px;
  font-size: 12px;
  line-height: 18px;
}

.permission-authority__meta--all {
  color: #15803d;
}

.permission-authority__meta--partial {
  color: #b45309;
}

.permission-authority__meta--none {
  color: #8c8c8c;
}

.permission-authority__desc {
  margin-top: 2px;
  color: #8a94a8;
  font-size: 12px;
  line-height: 18px;
}

.target-panel__summary {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 8px;
}

.target-panel__summary-item {
  display: inline-flex;
  align-items: center;
  height: 22px;
  padding: 0 8px;
  border: 1px solid #dde7f5;
  border-radius: 999px;
  background: #f7faff;
  color: #4b628b;
  font-size: 11px;
  line-height: 18px;
}

.target-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.target-item {
  padding: 9px 10px;
  border: 1px solid #e8edf5;
  border-radius: 10px;
  background: #fff;
  transition:
    border-color 0.2s ease,
    background-color 0.2s ease,
    box-shadow 0.2s ease;
}

.target-item:hover {
  border-color: #cddff7;
  background: #fafcff;
}

.target-item--checked {
  border-color: #9fc6f6;
  background: #f4f9ff;
  box-shadow: inset 0 0 0 1px rgb(22 119 255 / 10%);
}

.target-item__checkbox {
  width: 100%;
}

.target-item__checkbox :deep(.ant-checkbox + span) {
  width: 100%;
  min-width: 0;
  padding-right: 2px;
}

.target-item__content {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.target-item__name {
  color: #1f2738;
  font-size: 13px;
  line-height: 20px;
  font-weight: 500;
}

.target-item__desc {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
  color: #7c879b;
  font-size: 12px;
  line-height: 18px;
}

.target-item__action {
  font-weight: 500;
}

.target-item__action--add {
  color: #1677ff;
}

.target-item__action--remove {
  color: #dc2626;
}

:deep(mark) {
  background: #fff3bf !important;
  color: inherit;
}

@media (max-width: 1280px) {
  .batch-overview__metrics {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .batch-content {
    grid-template-columns: minmax(0, 1fr);
  }

  .target-panel {
    min-height: 280px;
  }
}

@media (max-width: 1024px) {
  .batch-overview__steps {
    grid-template-columns: 1fr;
    gap: 8px;
  }

  .permission-panel__toolbar {
    flex-direction: column;
    align-items: stretch;
  }

  .permission-panel__tools {
    justify-content: space-between;
  }

  .permission-panel__tools :deep(.ant-input-affix-wrapper) {
    width: 100%;
  }

  .batch-overview__mode {
    flex-direction: column;
    align-items: stretch;
  }

  .batch-overview__mode-left {
    flex-wrap: wrap;
  }
}

@media (max-width: 768px) {
  .batch-overview__metrics {
    grid-template-columns: 1fr;
  }
}
</style>
