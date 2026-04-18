<script setup lang="ts">
import type { AxiosError } from 'axios'
import {
  DownOutlined,
  SearchOutlined,
  UpOutlined,
} from '@ant-design/icons-vue'
import { Empty } from 'ant-design-vue'
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue'
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
  institutionId?: number | null
}>()

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'saved'): void
}>()

const simpleImage = Empty.PRESENTED_IMAGE_SIMPLE
const initialPermissionData: AuthorityGroup[] = []

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

const loading = ref(false)
const submitting = ref(false)
const treeLoading = ref(false)
const detail = ref<InstitutionPermissionDetail | null>(null)
const scopedMenuIds = ref<number[]>([])
const infoPanelRef = ref<HTMLElement>()
const permissionPanelHeight = ref<number>()
let infoPanelResizeObserver: ResizeObserver | null = null

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
  setDefaultCheckedByIds,
} = useRolePermissions(initialPermissionData, treeLoading, {
  enforceGroupExclusive: false,
})

const currentVersionName = computed(() => {
  const moduleName = String(detail.value?.currentModuleName || '').trim()
  if (moduleName)
    return moduleName
  return getOpenTypeLabel(detail.value?.openType)
})

const scopeHint = computed(() => {
  return '菜单入口跟随当前版本模板保留，页面和按钮能力按这里的实际权限生效。'
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

function closeModal() {
  emit('update:open', false)
}

function resetState() {
  detail.value = null
  scopedMenuIds.value = []
  permissionPanelHeight.value = undefined
  searchValue.value = ''
  updateData([])
  clearAllSelected()
}

function syncPermissionPanelHeight() {
  permissionPanelHeight.value = infoPanelRef.value?.offsetHeight || undefined
}

function cleanupInfoPanelResizeObserver() {
  if (infoPanelResizeObserver) {
    infoPanelResizeObserver.disconnect()
    infoPanelResizeObserver = null
  }
}

function setupInfoPanelResizeObserver() {
  cleanupInfoPanelResizeObserver()
  if (!infoPanelRef.value || typeof ResizeObserver === 'undefined')
    return

  infoPanelResizeObserver = new ResizeObserver(() => {
    syncPermissionPanelHeight()
  })
  infoPanelResizeObserver.observe(infoPanelRef.value)
  syncPermissionPanelHeight()
}

const permissionPanelStyle = computed(() => {
  if (!permissionPanelHeight.value)
    return undefined
  return {
    height: `${permissionPanelHeight.value}px`,
    minHeight: `${permissionPanelHeight.value}px`,
  }
})

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

function filterScopedMenuIDs(values: number[] = [], allowedValues: number[] = []) {
  const allowedSet = new Set(normalizeMenuIDList(allowedValues))
  return normalizeMenuIDList(values).filter(item => allowedSet.has(item))
}

async function hydratePermissionSelection(moduleId: number, effectiveMenuIds: number[] = []) {
  if (!moduleId) {
    updateData([])
    scopedMenuIds.value = []
    return
  }

  treeLoading.value = true
  try {
    const res = await getVersionDetailApi({ moduleId })
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
    setDefaultCheckedByIds(filterScopedMenuIDs(effectiveMenuIds, versionScopedMenuIds))
  }
  catch (error: any) {
    console.error('load module permission detail failed', error)
    messageService.error(resolveRequestErrorMessage(error, '获取版本权限失败'))
    updateData([])
    scopedMenuIds.value = []
  }
  finally {
    treeLoading.value = false
  }
}

async function loadData(institutionId: number) {
  loading.value = true
  try {
    const detailRes = await getInstitutionPermissionDetailApi({ institutionId })
    if (detailRes.code !== 200 || !detailRes.result) {
      messageService.error(detailRes.message || '获取机构权限信息失败')
      return
    }

    detail.value = detailRes.result
    await hydratePermissionSelection(
      Number(detailRes.result.currentModuleId || 0),
      Array.isArray(detailRes.result.effectiveMenuIds)
        ? detailRes.result.effectiveMenuIds.map(item => Number(item))
        : [],
    )
  }
  catch (error: any) {
    console.error('load institution permission data failed', error)
    messageService.error(resolveRequestErrorMessage(error, '获取机构权限信息失败'))
  }
  finally {
    loading.value = false
  }
}

async function submitPermissionConfig() {
  const institutionId = Number(props.institutionId || 0)
  const moduleId = Number(detail.value?.currentModuleId || 0)
  if (!institutionId || !moduleId)
    return

  const visibleAuthorityIDs = collectVisibleAuthorityIDs()
  const menuIds = filterScopedMenuIDs(collectSelectedPermissionIDs(), visibleAuthorityIDs)
  if (!menuIds.length) {
    messageService.warning('请至少选择一个机构权限')
    return
  }

  const currentEffectiveMenuIds = filterScopedMenuIDs(
    Array.isArray(detail.value?.effectiveMenuIds)
      ? detail.value!.effectiveMenuIds.map(item => Number(item))
      : [],
    visibleAuthorityIDs,
  )

  if (isSameMenuSelection(menuIds, currentEffectiveMenuIds)) {
    messageService.warning('机构权限没有变化')
    return
  }

  submitting.value = true
  try {
    const res = await replaceInstitutionPermissionVersionApi({
      institutionId,
      moduleId,
      menuIds,
    })
    if (res.code !== 200) {
      messageService.error(res.message || '机构权限保存失败')
      return
    }

    messageService.success('机构权限保存成功')
    emit('saved')
    await loadData(institutionId)
  }
  catch (error: any) {
    console.error('save institution permission failed', error)
    messageService.error(resolveRequestErrorMessage(error, '机构权限保存失败'))
  }
  finally {
    submitting.value = false
  }
}

watch(
  () => [props.open, props.institutionId] as const,
  ([open, institutionId]) => {
    if (!open) {
      cleanupInfoPanelResizeObserver()
      resetState()
      return
    }

    if (institutionId)
      void loadData(Number(institutionId))
  },
  { immediate: true },
)

watch(
  () => [props.open, detail.value?.institutionId, selectedLeafCount.value, totalLeafCount.value] as const,
  async ([open]) => {
    if (!open)
      return
    await nextTick()
    setupInfoPanelResizeObserver()
  },
)

onBeforeUnmount(() => {
  cleanupInfoPanelResizeObserver()
})
</script>

<template>
  <PlatformModalShell
    v-model:open="openModal"
    :width="1240"
    title="机构权限配置"
    modal-class="institution-permission-modal"
    scrollable
  >
    <a-spin :spinning="loading">
      <div class="permission-modal">
        <div class="permission-overview">
          <div class="permission-overview__main">
            <div class="permission-overview__name">
              {{ detail?.organName || '--' }}
            </div>
            <div class="permission-overview__meta">
              <span>登录账号：{{ detail?.mobile || '--' }}</span>
              <span>当前版本：{{ currentVersionName }}</span>
              <span>过期时间：{{ formatDateMinute(detail?.expireEndTime) }}</span>
            </div>
          </div>

          <span class="status-chip" :class="getStatusClass(detail?.status)">
            {{ getStatusLabel(detail?.status) }}
          </span>
        </div>

        <div class="permission-layout">
          <div ref="infoPanelRef" class="permission-info-panel">
            <div class="info-card">
              <div class="info-card__label">
                当前版本
              </div>
              <div class="info-card__value">
                {{ currentVersionName }}
              </div>
              <div class="info-card__grid">
                <div class="info-card__item">
                  <span class="info-card__item-label">账号状态</span>
                  <span class="info-card__item-value">{{ getStatusLabel(detail?.status) }}</span>
                </div>
                <div class="info-card__item">
                  <span class="info-card__item-label">到期时间</span>
                  <span class="info-card__item-value">{{ formatDateMinute(detail?.expireEndTime) }}</span>
                </div>
              </div>
            </div>

            <div class="info-note">
              <div class="info-note__title">
                配置说明
              </div>
              <div class="info-note__text">
                {{ scopeHint }}
              </div>
            </div>

            <div class="info-stats">
              <div class="info-stats__item">
                <span class="info-stats__label">已选权限</span>
                <span class="info-stats__value">{{ selectedLeafCount }}</span>
              </div>
              <div class="info-stats__item">
                <span class="info-stats__label">可配权限</span>
                <span class="info-stats__value">{{ totalLeafCount }}</span>
              </div>
            </div>
          </div>

          <div class="permission-panel" :style="permissionPanelStyle">
            <div class="permission-panel__header">
              <div class="permission-panel__titlebox">
                <div class="permission-panel__title">
                  机构实际权限
                </div>
                <div class="permission-panel__meta">
                  <span>仅在当前版本范围内调整</span>
                  <span>页面权限与按钮权限一起生效</span>
                </div>
              </div>

              <a-button type="link" class="permission-panel__link" @click="toggleAllExpand">
                {{ isAllExpanded ? '一键收起' : '一键展开' }}
                <component :is="isAllExpanded ? UpOutlined : DownOutlined" />
              </a-button>
            </div>

            <div class="permission-panel__toolbar">
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
        </div>
      </div>
    </a-spin>

    <template #footer>
      <a-button @click="closeModal">
        关闭
      </a-button>
      <a-button type="primary" :loading="submitting" @click="submitPermissionConfig">
        保存权限
      </a-button>
    </template>
  </PlatformModalShell>
</template>

<style scoped lang="less">
.permission-modal {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding-top: 4px;
}

.permission-layout {
  display: grid;
  grid-template-columns: 300px minmax(0, 1fr);
  gap: 12px;
  align-items: start;
}

.permission-overview {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 16px 18px;
  border: 1px solid #e8edf5;
  border-radius: 18px;
  background: #fff;
}

.permission-overview__main {
  min-width: 0;
}

.permission-overview__name {
  color: #262626;
  font-size: 16px;
  font-weight: 600;
  line-height: 24px;
}

.permission-overview__meta {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px 12px;
  margin-top: 6px;
  color: #595959;
  font-size: 12px;
  line-height: 18px;
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

.permission-info-panel {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.info-card,
.info-note,
.info-stats {
  border: 1px solid #e8edf5;
  border-radius: 18px;
  background: #fff;
}

.info-card {
  padding: 18px;
  background: linear-gradient(180deg, #fafcff 0%, #ffffff 100%);
}

.info-card__label {
  color: #8c8c8c;
  font-size: 12px;
  line-height: 18px;
}

.info-card__value {
  margin-top: 6px;
  color: #1f2329;
  font-size: 22px;
  font-weight: 700;
  line-height: 30px;
}

.info-card__grid {
  display: grid;
  gap: 10px;
  margin-top: 16px;
}

.info-card__item {
  padding: 10px 12px;
  border: 1px solid #eef2f7;
  border-radius: 12px;
  background: rgba(250, 250, 250, 0.9);
}

.info-card__item-label {
  display: block;
  color: #8c8c8c;
  font-size: 12px;
  line-height: 18px;
}

.info-card__item-value {
  display: block;
  margin-top: 4px;
  color: #262626;
  font-size: 13px;
  font-weight: 600;
  line-height: 20px;
}

.info-note {
  padding: 16px 18px;
  background: #f8fbff;
}

.info-note__title {
  color: #262626;
  font-size: 14px;
  font-weight: 600;
  line-height: 22px;
}

.info-note__text {
  margin-top: 8px;
  color: #5b6475;
  font-size: 12px;
  line-height: 20px;
}

.info-stats {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0;
  overflow: hidden;
}

.info-stats__item {
  padding: 16px 14px;
}

.info-stats__item + .info-stats__item {
  border-left: 1px solid #eef2f7;
}

.info-stats__label {
  display: block;
  color: #8c8c8c;
  font-size: 12px;
  line-height: 18px;
}

.info-stats__value {
  display: block;
  margin-top: 6px;
  color: #1f2329;
  font-size: 22px;
  font-weight: 700;
  line-height: 30px;
}

.permission-panel {
  display: flex;
  flex-direction: column;
  min-width: 0;
  min-height: 580px;
  border: 1px solid #e8edf5;
  border-radius: 18px;
  background: #fff;
  overflow: hidden;
}

.permission-panel__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 16px 18px 10px;
  border-bottom: 1px solid #eef2f7;
}

.permission-panel__titlebox {
  min-width: 0;
}

.permission-panel__title {
  color: #262626;
  font-size: 15px;
  font-weight: 600;
  line-height: 22px;
}

.permission-panel__meta {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 4px;
  color: #8c8c8c;
  font-size: 12px;
  line-height: 18px;
}

.permission-panel__link {
  flex-shrink: 0;
  padding-inline: 0;
}

.permission-panel__toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 10px 18px 12px;
  border-bottom: 1px solid #f2f4f7;
}

.permission-panel__toolbar :deep(.ant-input-affix-wrapper) {
  width: 320px;
}

.permission-panel__body {
  flex: 1;
  min-height: 0;
  padding: 0 18px 18px;
  overflow: auto;
}

.permission-box {
  padding-top: 8px;
}

.permission-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  min-height: 38px;
  border-bottom: 1px solid #f3f5f8;
}

.permission-row--group {
  min-height: 38px;
}

.permission-row--child {
  min-height: 38px;
  padding-left: 20px;
}

.permission-row--authority {
  min-height: 38px;
  padding-left: 44px;
}

.permission-row__main {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
  flex: 1;
}

.permission-row__main--authority {
  align-items: flex-start;
  padding: 8px 0;
}

.permission-row__title {
  min-width: 0;
  color: #262626;
  line-height: 22px;
}

.permission-row__title--group {
  font-size: 14px;
  font-weight: 600;
}

.permission-row__title--child {
  font-size: 13px;
  font-weight: 500;
}

.permission-row__toggle {
  flex-shrink: 0;
  color: #1677ff;
  font-size: 12px;
  line-height: 18px;
  cursor: pointer;
}

.permission-authority {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}

.permission-authority__title {
  color: #262626;
  font-size: 13px;
  line-height: 20px;
}

.permission-authority__desc {
  color: #8c8c8c;
  font-size: 12px;
  line-height: 18px;
}

.permission-row :deep(mark) {
  color: inherit;
  border-radius: 4px;
}

@media (max-width: 1200px) {
  .permission-layout {
    grid-template-columns: 1fr;
  }

  .permission-panel {
    height: auto !important;
    min-height: 480px;
  }
}
</style>
