<script setup lang="ts">
import type { FormInstance, Rule } from 'ant-design-vue/es/form'
import { CloseOutlined } from '@ant-design/icons-vue'
import { computed, h, nextTick, onBeforeUnmount, reactive, ref, watch } from 'vue'
import type { VNodeChild } from 'vue'
import {
  createVersionApi,
  getVersionDetailApi,
  getVersionMenuTreeApi,
  saveVersionMenusApi,
  updateVersionApi,
} from '@/api/platform/versions'
import {
  buildPermissionTreeData,
  collectAllKeys,
  collectLeafCheckedKeys,
  collectLeafKeysBySelectedSet,
  countLeafNodes,
  type PermissionTreeNode,
} from '../../shared/permission-tree'
import messageService from '@/utils/messageService'

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

interface DisplayPermissionTreeNode extends Omit<PermissionTreeNode, 'title' | 'children'> {
  title: string | VNodeChild
  children?: DisplayPermissionTreeNode[]
}

const formRef = ref<FormInstance>()
const formCardRef = ref<HTMLElement>()
const submitting = ref(false)
const detailLoading = ref(false)
const treeLoading = ref(false)
const treeKeyword = ref('')
const baseMenuTree = ref<PermissionTreeNode[]>([])
const checkedKeys = ref<number[]>([])
const halfCheckedKeys = ref<number[]>([])
const expandedKeys = ref<number[]>([])
const autoExpandParent = ref(false)
const permissionCardHeight = ref<number>()
let formCardResizeObserver: ResizeObserver | null = null

const openModal = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

const isEdit = computed(() => Number(props.versionId || 0) > 0)
const modalTitle = computed(() => (isEdit.value ? '编辑版本' : '新建版本'))
const totalLeafCount = computed(() => countLeafNodes(baseMenuTree.value))
const selectedLeafCount = computed(() => collectLeafKeysBySelectedSet(baseMenuTree.value, [...checkedKeys.value, ...halfCheckedKeys.value]).length)
const rootExpandedKeys = computed(() => baseMenuTree.value.map(node => Number(node.key)))
const allTreeKeys = computed(() => collectAllKeys(baseMenuTree.value))
const permissionCardStyle = computed(() => (
  permissionCardHeight.value
    ? { height: `${permissionCardHeight.value}px` }
    : undefined
))
const displayMenuTree = computed<DisplayPermissionTreeNode[]>(() => filterMenuTreeByKeyword(baseMenuTree.value, treeKeyword.value))

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

function normalizeKeyword(value?: string) {
  return String(value || '').trim()
}

function renderHighlightedTitle(title: string, keyword: string) {
  const normalizedKeyword = normalizeKeyword(keyword)
  if (!normalizedKeyword)
    return title

  const source = String(title || '')
  const lowerSource = source.toLowerCase()
  const lowerKeyword = normalizedKeyword.toLowerCase()
  const children: Array<string | VNodeChild> = []
  let cursor = 0

  while (cursor < source.length) {
    const matchedIndex = lowerSource.indexOf(lowerKeyword, cursor)
    if (matchedIndex < 0) {
      children.push(source.slice(cursor))
      break
    }

    if (matchedIndex > cursor)
      children.push(source.slice(cursor, matchedIndex))

    children.push(
      h(
        'span',
        { class: 'permission-tree-title__highlight' },
        source.slice(matchedIndex, matchedIndex + normalizedKeyword.length),
      ),
    )
    cursor = matchedIndex + normalizedKeyword.length
  }

  return h('span', { class: 'permission-tree-title' }, children)
}

function clonePermissionSubtree(nodes: PermissionTreeNode[] = [], keyword = ''): DisplayPermissionTreeNode[] {
  return nodes.map(node => ({
    ...node,
    title: renderHighlightedTitle(String(node.title || ''), keyword),
    children: Array.isArray(node.children) && node.children.length
      ? clonePermissionSubtree(node.children, keyword)
      : undefined,
  }))
}

function filterMenuTreeByKeyword(nodes: PermissionTreeNode[] = [], keyword = ''): DisplayPermissionTreeNode[] {
  const normalizedKeyword = normalizeKeyword(keyword)
  if (!normalizedKeyword)
    return nodes as DisplayPermissionTreeNode[]

  const lowerKeyword = normalizedKeyword.toLowerCase()

  const walk = (items: PermissionTreeNode[]): DisplayPermissionTreeNode[] => {
    const result: DisplayPermissionTreeNode[] = []

    items.forEach((item) => {
      const title = String(item.title || '')
      const matched = title.toLowerCase().includes(lowerKeyword)
      const filteredChildren = Array.isArray(item.children) ? walk(item.children) : []

      if (!matched && !filteredChildren.length)
        return

      result.push({
        ...item,
        title: renderHighlightedTitle(title, normalizedKeyword),
        children: matched
          ? (Array.isArray(item.children) && item.children.length ? clonePermissionSubtree(item.children, normalizedKeyword) : undefined)
          : filteredChildren,
      })
    })

    return result
  }

  return walk(nodes)
}

function collectDisplayTreeKeys(nodes: DisplayPermissionTreeNode[] = []) {
  const result: number[] = []

  const walk = (items: DisplayPermissionTreeNode[]) => {
    items.forEach((item) => {
      result.push(Number(item.key))
      if (Array.isArray(item.children) && item.children.length > 0)
        walk(item.children)
    })
  }

  walk(nodes)
  return result
}

function resetState() {
  formState.id = undefined
  formState.name = ''
  formState.price = 0
  formState.remark = ''
  detailMeta.orgCount = 0
  detailMeta.updateTime = ''
  checkedKeys.value = []
  halfCheckedKeys.value = []
  treeKeyword.value = ''
  expandedKeys.value = rootExpandedKeys.value
  autoExpandParent.value = false
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

function expandAllTree() {
  expandedKeys.value = allTreeKeys.value
  autoExpandParent.value = false
}

function collapseTree() {
  expandedKeys.value = rootExpandedKeys.value
  autoExpandParent.value = false
}

function collapseAllTree() {
  expandedKeys.value = []
  autoExpandParent.value = false
}

function getMergedMenuIds() {
  return Array.from(new Set([...checkedKeys.value, ...halfCheckedKeys.value]))
    .filter(key => Number(key) > 0)
    .map(key => Number(key))
}

async function ensureMenuTreeLoaded() {
  if (baseMenuTree.value.length)
    return

  treeLoading.value = true
  try {
    const res = await getVersionMenuTreeApi({ type: 1 })
    if (res.code !== 200) {
      messageService.error(res.message || '获取权限树失败')
      return
    }
    baseMenuTree.value = buildPermissionTreeData(res.result || [])
    expandedKeys.value = rootExpandedKeys.value
    autoExpandParent.value = false
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
    checkedKeys.value = collectLeafCheckedKeys(res.result.menuIds || [])
    halfCheckedKeys.value = []
    expandedKeys.value = treeKeyword.value
      ? collectDisplayTreeKeys(displayMenuTree.value)
      : rootExpandedKeys.value
    autoExpandParent.value = !!treeKeyword.value
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

  const menuIds = getMergedMenuIds()
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

function handleTreeCheck(value: any, info: any) {
  checkedKeys.value = Array.isArray(value) ? value.map(Number) : (value?.checked || []).map(Number)
  halfCheckedKeys.value = (info?.halfCheckedKeys || []).map((key: string | number) => Number(key))
}

function handleExpandedKeysChange(keys: Array<string | number>) {
  expandedKeys.value = (keys || []).map(key => Number(key))
  autoExpandParent.value = false
}

watch(
  () => treeKeyword.value,
  (keyword) => {
    const normalizedKeyword = normalizeKeyword(keyword)
    expandedKeys.value = normalizedKeyword
      ? collectDisplayTreeKeys(displayMenuTree.value)
      : rootExpandedKeys.value
    autoExpandParent.value = !!normalizedKeyword
  },
)

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
  <a-modal
    v-model:open="openModal"
    centered
    destroy-on-close
    :keyboard="false"
    :closable="false"
    :mask-closable="false"
    :width="1240"
    class="createStu-modal-content-box platform-version-modal"
  >
    <template #title>
      <div class="version-modal__titlebar">
        <span>{{ modalTitle }}</span>
        <a-button type="text" class="close-btn" @click="closeModal">
          <template #icon>
            <CloseOutlined class="close-icon" />
          </template>
        </a-button>
      </div>
    </template>

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

              <div class="permission-card__actions">
                <a-button type="link" class="permission-card__link" @click="expandAllTree">
                  展开全部
                </a-button>
                <a-button type="link" class="permission-card__link" @click="collapseTree">
                  收起层级
                </a-button>
                <a-button type="link" class="permission-card__link" @click="collapseAllTree">
                  收起全部
                </a-button>
              </div>
            </div>

            <div class="permission-card__toolbar">
              <a-input
                v-model:value="treeKeyword"
                allow-clear
                placeholder="搜索菜单名称"
              />
            </div>

            <div class="permission-card__body">
              <a-empty
                v-if="!treeLoading && !displayMenuTree.length"
                :description="treeKeyword ? '未匹配到相关菜单' : '暂无可配置菜单'"
              />
              <a-spin v-else :spinning="treeLoading">
                <a-tree
                  :tree-data="displayMenuTree"
                  :checked-keys="checkedKeys"
                  :expanded-keys="expandedKeys"
                  :auto-expand-parent="autoExpandParent"
                  :selectable="false"
                  checkable
                  @check="handleTreeCheck"
                  @update:expanded-keys="handleExpandedKeysChange"
                />
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
  </a-modal>
</template>

<style scoped lang="less">
:deep(.createStu-modal-content-box.platform-version-modal .ant-modal-content) {
  border-radius: 22px;
  overflow: hidden;
  box-shadow: 0 18px 46px rgba(15, 23, 42, 0.14);
}

:deep(.createStu-modal-content-box.platform-version-modal .ant-modal-header) {
  padding: 24px 28px 14px;
  margin-bottom: 0;
  border-bottom: none;
}

:deep(.createStu-modal-content-box.platform-version-modal .ant-modal-body) {
  padding: 0 28px 0;
  max-height: calc(100vh - 220px);
  overflow-y: auto;
  overflow-x: hidden;
}

:deep(.createStu-modal-content-box.platform-version-modal .ant-modal-footer) {
  padding: 18px 28px 24px;
  border-top: none;
}

.version-modal__titlebar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  color: #1f2329;
  font-size: 20px;
  font-weight: 700;
  line-height: 32px;
}

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

.permission-card__actions {
  display: flex;
  align-items: center;
  gap: 6px;
}

.permission-card__link {
  padding: 0;
}

.permission-card__toolbar {
  padding: 16px 22px 0;
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

.permission-card__body :deep(.ant-tree-checkbox) {
  margin-block-start: 0;
}

.permission-card__body :deep(.permission-tree-title__highlight) {
  padding: 0 2px;
  border-radius: 4px;
  background: rgba(22, 119, 255, 0.14);
  color: #1677ff;
  font-weight: 700;
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
