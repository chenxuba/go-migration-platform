<script setup lang="ts">
import type { FormInstance, Rule } from 'ant-design-vue/es/form'
import {
  DownOutlined,
  SearchOutlined,
  UpOutlined,
} from '@ant-design/icons-vue'
import { Empty } from 'ant-design-vue'
import { computed, nextTick, onBeforeUnmount, reactive, ref, watch } from 'vue'
import {
  createVersionApi,
  getVersionDetailApi,
  getVersionMenuTreeApi,
  saveVersionMenusApi,
  updateVersionApi,
  type MenuTreeNode,
} from '@/api/platform/versions'
import {
  type AuthorityGroup,
  useRolePermissions,
} from '@/composables/useRolePermissions'
import messageService from '@/utils/messageService'
import PlatformModalShell from '../../shared/platform-modal-shell.vue'

const props = defineProps<{
  open: boolean
  versionId?: number | null
}>()

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'saved'): void
}>()

interface FormState {
  id?: number
  name: string
  price?: number
  remark: string
}

interface DetailMetaState {
  orgCount: number
  updateTime: string
}

const simpleImage = Empty.PRESENTED_IMAGE_SIMPLE
const initialPermissionData: AuthorityGroup[] = []

const formRef = ref<FormInstance>()
const formCardRef = ref<HTMLElement>()
const submitting = ref(false)
const detailLoading = ref(false)
const treeLoading = ref(false)
const permissionCardHeight = ref<number>()
let formCardResizeObserver: ResizeObserver | null = null

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

const openModal = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

const isEdit = computed(() => Number(props.versionId || 0) > 0)
const modalTitle = computed(() => (isEdit.value ? '编辑版本' : '新建版本'))
const totalLeafCount = computed(() => {
  return collectAuthorityIDSet(false).size
})
const selectedLeafCount = computed(() => {
  return collectAuthorityIDSet(true).size
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
const permissionCardStyle = computed(() => (
  permissionCardHeight.value
    ? { height: `${permissionCardHeight.value}px` }
    : undefined
))

const formState = reactive<FormState>({
  name: '',
  price: 0,
  remark: '',
})

const detailMeta = reactive<DetailMetaState>({
  orgCount: 0,
  updateTime: '',
})

const rules: Record<string, Rule[]> = {
  name: [{ required: true, message: '请输入版本名称', trigger: 'blur' }],
}

function formatDateMinute(value?: string) {
  const raw = String(value || '').trim()
  if (!raw)
    return '--'
  return raw.length >= 16 ? raw.slice(0, 16) : raw
}

function formatPrice(value?: number) {
  return `¥${Number(value || 0).toFixed(2)}`
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
        buildAuthorityChild(childNode, groupIndex+1, childIndex),
      ),
    }
  })
}

function collectSelectedMenuIDs() {
  const selectedIDs = new Set<number>()

  boxList.value.forEach((parent) => {
    const parentID = Number(parent.id)
    if ((parent.checked || parent.indeterminate) && parentID > 0)
      selectedIDs.add(parentID)

    parent.children.forEach((child) => {
      const childID = Number(child.id)
      if ((child.checked || child.indeterminate) && childID > 0)
        selectedIDs.add(childID)

      child.children.forEach((authority) => {
        if (authority.checked && Number(authority.id) > 0)
          selectedIDs.add(Number(authority.id))
      })
    })
  })

  return Array.from(selectedIDs)
}

function collectAuthorityIDSet(checkedOnly = false) {
  const authorityIDs = new Set<number>()

  boxList.value.forEach((parent) => {
    parent.children.forEach((child) => {
      child.children.forEach((authority) => {
        const authorityID = Number(authority.id)
        if (!Number.isFinite(authorityID) || authorityID <= 0)
          return
        if (checkedOnly && !authority.checked)
          return
        authorityIDs.add(authorityID)
      })
    })
  })

  return authorityIDs
}

function resetState() {
  formState.id = undefined
  formState.name = ''
  formState.price = 0
  formState.remark = ''
  detailMeta.orgCount = 0
  detailMeta.updateTime = ''
  searchValue.value = ''
  clearAllSelected()
  nextTick(() => {
    formRef.value?.clearValidate?.()
  })
}

function closeModal() {
  emit('update:open', false)
}

function cleanupFormCardResizeObserver() {
  if (formCardResizeObserver) {
    formCardResizeObserver.disconnect()
    formCardResizeObserver = null
  }
}

function syncPermissionCardHeight() {
  permissionCardHeight.value = formCardRef.value?.offsetHeight || undefined
}

function setupFormCardResizeObserver() {
  cleanupFormCardResizeObserver()
  if (!formCardRef.value || typeof ResizeObserver === 'undefined')
    return

  formCardResizeObserver = new ResizeObserver(() => {
    syncPermissionCardHeight()
  })
  formCardResizeObserver.observe(formCardRef.value)
  syncPermissionCardHeight()
}

async function ensureMenuTreeLoaded() {
  if (boxList.value.length)
    return

  treeLoading.value = true
  try {
    const res = await getVersionMenuTreeApi({ type: 1 })
    if (res.code !== 200) {
      messageService.error(res.message || '获取权限树失败')
      return
    }
    updateData(mapMenuTreeToPermissionGroups(res.result || []))
  }
  catch (error: any) {
    console.error('load version menu tree failed', error)
    messageService.error(error?.message || '获取权限树失败')
  }
  finally {
    treeLoading.value = false
  }
}

async function loadDetail(versionId: number) {
  detailLoading.value = true
  try {
    const res = await getVersionDetailApi({ moduleId: versionId })
    if (res.code !== 200 || !res.result) {
      messageService.error(res.message || '获取版本详情失败')
      return
    }

    formState.id = res.result.moduleId
    formState.name = String(res.result.moduleName || '')
    formState.price = Number(res.result.price || 0)
    formState.remark = String(res.result.remark || '')
    detailMeta.orgCount = Number(res.result.orgCount || 0)
    detailMeta.updateTime = String(res.result.updateTime || '')
    setDefaultCheckedByIds(
      Array.isArray(res.result.selectedMenuIds)
        ? res.result.selectedMenuIds.map(item => Number(item))
        : [],
    )
  }
  catch (error: any) {
    console.error('load version detail failed', error)
    messageService.error(error?.message || '获取版本详情失败')
  }
  finally {
    detailLoading.value = false
  }
}

async function submitForm() {
  try {
    await formRef.value?.validate()
  }
  catch {
    return
  }

  const menuIds = collectSelectedMenuIDs()
  if (!menuIds.length) {
    messageService.warning('请至少选择一个菜单权限')
    return
  }

  submitting.value = true
  try {
    const payload = {
      name: formState.name.trim(),
      type: 1,
      price: Number(formState.price || 0),
      remark: formState.remark.trim() || undefined,
    }

    let versionId = Number(formState.id || 0)
    const basicRes = versionId
      ? await updateVersionApi({ id: versionId, ...payload })
      : await createVersionApi(payload)

    if (basicRes.code !== 200) {
      messageService.error(basicRes.message || (versionId ? '更新版本失败' : '新增版本失败'))
      return
    }

    if (!versionId)
      versionId = Number((basicRes.result as any)?.id || 0)

    if (!versionId) {
      messageService.error('保存版本失败')
      return
    }

    const menuRes = await saveVersionMenusApi({ id: versionId, menuIds })
    if (menuRes.code !== 200) {
      messageService.error(menuRes.message || '保存版本权限失败')
      return
    }

    messageService.success(isEdit.value ? '版本更新成功' : '版本创建成功')
    emit('saved')
    closeModal()
  }
  catch (error: any) {
    console.error('submit version form failed', error)
    messageService.error(error?.message || '保存版本失败')
  }
  finally {
    submitting.value = false
  }
}

watch(
  () => [props.open, props.versionId] as const,
  async ([open, versionId]) => {
    if (!open) {
      cleanupFormCardResizeObserver()
      permissionCardHeight.value = undefined
      resetState()
      return
    }

    await ensureMenuTreeLoaded()
    resetState()

    if (versionId)
      await loadDetail(Number(versionId))

    await nextTick()
    setupFormCardResizeObserver()
  },
  { immediate: true },
)

onBeforeUnmount(() => {
  cleanupFormCardResizeObserver()
})
</script>

<template>
  <PlatformModalShell
    v-model:open="openModal"
    :width="1240"
    :title="modalTitle"
    modal-class="platform-version-modal"
    scrollable
  >
    <a-spin :spinning="detailLoading">
      <div class="version-modal">
        <div class="version-modal__aside">
          <div ref="formCardRef" class="version-card version-card--form">
            <div class="version-card__header">
              <div class="version-card__title">
                基础信息
              </div>
              <div class="version-card__price">
                {{ formatPrice(formState.price) }}
              </div>
            </div>

            <a-form ref="formRef" layout="vertical" :model="formState" :rules="rules" class="version-form">
              <a-form-item label="版本名称" name="name">
                <a-input v-model:value="formState.name" :maxlength="40" placeholder="请输入版本名称" />
              </a-form-item>

              <a-form-item label="版本价格">
                <a-input-number
                  v-model:value="formState.price"
                  class="version-form__number"
                  :min="0"
                  :precision="2"
                  :controls="false"
                  placeholder="请输入版本价格"
                />
              </a-form-item>

              <a-form-item label="版本备注">
                <a-textarea
                  v-model:value="formState.remark"
                  :maxlength="120"
                  :rows="4"
                  placeholder="请输入版本说明"
                />
              </a-form-item>
            </a-form>

            <div class="version-stats">
              <div class="version-stat">
                <span class="version-stat__label">菜单覆盖</span>
                <span class="version-stat__value">
                  {{ totalLeafCount ? `${selectedLeafCount}/${totalLeafCount}` : '--' }}
                </span>
              </div>
              <div class="version-stat">
                <span class="version-stat__label">绑定机构</span>
                <span class="version-stat__value">{{ detailMeta.orgCount }}</span>
              </div>
              <div class="version-stat">
                <span class="version-stat__label">最近更新</span>
                <span class="version-stat__value">{{ formatDateMinute(detailMeta.updateTime) }}</span>
              </div>
            </div>
          </div>
        </div>

        <div class="version-modal__main">
          <div class="version-card version-card--permission" :style="permissionCardStyle">
            <div class="permission-card__header">
              <div class="permission-card__titlebox">
                <div class="permission-card__title">
                  菜单权限
                </div>
                <div class="permission-card__meta">
                  <span>已选 {{ selectedLeafCount }} 项</span>
                  <span>共 {{ totalLeafCount }} 项</span>
                </div>
              </div>

              <a-button type="link" class="permission-card__link" @click="toggleAllExpand">
                {{ isAllExpanded ? '一键收起' : '一键展开' }}
                <component :is="isAllExpanded ? UpOutlined : DownOutlined" />
              </a-button>
            </div>

            <div class="permission-card__toolbar">
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

            <div class="permission-card__body">
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
        取消
      </a-button>
      <a-button type="primary" :loading="submitting" @click="submitForm">
        保存版本
      </a-button>
    </template>
  </PlatformModalShell>
</template>

<style scoped lang="less">
.version-modal {
  display: grid;
  grid-template-columns: 320px minmax(0, 1fr);
  gap: 18px;
  padding-top: 8px;
  align-items: stretch;
}

.version-modal__aside,
.version-modal__main {
  min-height: 0;
}

.version-modal__aside {
  align-self: start;
}

.version-modal__main {
  display: flex;
}

.version-card {
  width: 100%;
  border: 1px solid #e8edf5;
  border-radius: 18px;
  background: #fff;
  overflow: hidden;
  box-shadow: 0 14px 32px rgba(15, 23, 42, 0.06);
}

.version-card--form {
  padding: 20px 20px 18px;
  background:
    linear-gradient(180deg, rgba(22, 119, 255, 0.06) 0%, rgba(22, 119, 255, 0) 132px),
    #fff;
}

.version-card__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 18px;
}

.version-card__title {
  color: #1f2329;
  font-size: 16px;
  font-weight: 700;
}

.version-card__price {
  padding: 8px 12px;
  border-radius: 999px;
  background: rgba(22, 119, 255, 0.08);
  color: #1677ff;
  font-size: 13px;
  font-weight: 600;
  line-height: 20px;
}

.version-form :deep(.ant-form-item) {
  margin-bottom: 18px;
}

.version-form__number {
  width: 100%;
}

.version-stats {
  display: grid;
  gap: 10px;
  margin-top: 18px;
}

.version-stat {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 12px 14px;
  border: 1px solid rgba(22, 119, 255, 0.08);
  border-radius: 14px;
  background: rgba(248, 250, 252, 0.92);
}

.version-stat__label {
  color: #667085;
  font-size: 12px;
  line-height: 18px;
}

.version-stat__value {
  color: #1f2329;
  font-size: 13px;
  font-weight: 600;
  line-height: 20px;
}

.version-card--permission {
  display: flex;
  flex-direction: column;
  height: 100%;
  min-height: 0;
}

.permission-card__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 20px 22px 14px;
  border-bottom: 1px solid #edf1f7;
  background: linear-gradient(180deg, rgba(15, 23, 42, 0.02) 0%, rgba(15, 23, 42, 0) 100%);
}

.permission-card__titlebox {
  min-width: 0;
}

.permission-card__title {
  color: #1f2329;
  font-size: 16px;
  font-weight: 700;
  line-height: 24px;
}

.permission-card__meta {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
  margin-top: 6px;
  color: #667085;
  font-size: 12px;
  line-height: 18px;
}

.permission-card__link {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 0;
}

.permission-card__toolbar {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 14px 22px 0;
}

.permission-card__toolbar :deep(.ant-input-affix-wrapper) {
  flex: 1;
}

.permission-card__body {
  flex: 1;
  min-height: 0;
  padding: 12px 14px 18px;
  overflow: auto;
  scrollbar-width: thin;
  scrollbar-color: rgba(15, 23, 42, 0.24) transparent;

  &::-webkit-scrollbar {
    width: 8px;
    height: 8px;
  }

  &::-webkit-scrollbar-track {
    background: transparent;
  }

  &::-webkit-scrollbar-thumb {
    border: 2px solid transparent;
    border-radius: 999px;
    background: rgba(15, 23, 42, 0.22);
    background-clip: padding-box;
  }

  &::-webkit-scrollbar-thumb:hover {
    background: rgba(15, 23, 42, 0.34);
    background-clip: padding-box;
  }
}

.permission-box {
  overflow: hidden;
  border: 1px solid #e5e7eb;
  border-radius: 14px;
  background: #fff;
}

.permission-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-height: 38px;
  padding: 0 14px;
  box-shadow: inset 0 -1px 0 0 #eef2f7;
}

.permission-row--group {
  min-height: 38px;
}

.permission-row--child {
  min-height: 38px;
  padding-left: 34px;
}

.permission-row--authority {
  min-height: 58px;
  padding-left: 52px;
  background: #fcfcfd;
}

.permission-row__main {
  display: flex;
  align-items: center;
  min-width: 0;
}

.permission-row__main--authority {
  align-items: flex-start;
}

.permission-row__title {
  margin-left: 8px;
  color: #1f2329;
}

.permission-row__title--group,
.permission-row__title--child {
  font-weight: 600;
}

.permission-row__title--group {
  font-size: 14px;
}

.permission-row__title--child {
  font-size: 13px;
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
  margin-left: 8px;
}

.permission-authority__title {
  color: #1f2329;
  font-size: 12px;
  line-height: 18px;
}

.permission-authority__desc {
  margin-top: 2px;
  padding-right: 120px;
  color: #8a94a6;
  font-size: 12px;
  line-height: 16px;
}

.permission-card__body :deep(mark) {
  padding: 0 2px;
  border-radius: 4px;
  background: rgba(22, 119, 255, 0.14);
  color: #1677ff;
  font-weight: 700;
}

@media (max-width: 1100px) {
  .version-modal {
    grid-template-columns: 1fr;
  }

  .version-modal__aside,
  .version-modal__main,
  .version-card--form,
  .version-card--permission {
    height: auto;
  }

  .version-card--permission {
    min-height: 540px;
  }
}
</style>
