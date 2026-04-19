<template>
  <div class="default-role-page">
    <a-card title="默认角色管理" :bordered="false">
      <div class="toolbar">
        <a-space wrap>
          <a-segmented
            v-model:value="currentPortal"
            :options="portalOptions"
            @change="handlePortalChange"
          />
          <a-input
            v-model:value="keywordInput"
            placeholder="请输入角色名称"
            allow-clear
            style="width: 260px"
            @pressEnter="handleSearch"
          />
          <a-button type="primary" @click="handleSearch">
            搜索
          </a-button>
          <a-button @click="handleReset">
            重置
          </a-button>
          <a-button type="primary" @click="openCreateModal">
            <template #icon>
              <PlusOutlined />
            </template>
            新增默认角色
          </a-button>
        </a-space>
      </div>

      <a-table
        class="role-table"
        :columns="columns"
        :data-source="filteredRoleList"
        :loading="loading"
        :pagination="false"
        :scroll="{ x: 980 }"
        row-key="roleId"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'roleName'">
            <span class="ellipsis-text" :title="toRoleRecord(record).roleName || '--'">
              {{ toRoleRecord(record).roleName || '--' }}
            </span>
          </template>

          <template v-else-if="column.key === 'isDefault'">
            <a-tag :color="toRoleRecord(record).isDefault ? 'blue' : 'default'">
              {{ toRoleRecord(record).isDefault ? '系统内置' : '自定义' }}
            </a-tag>
          </template>

          <template v-else-if="column.key === 'permissionStats'">
            <div class="permission-summary">
              <span class="permission-summary__item">
                <strong>{{ toRoleRecord(record).functionalAuthorityCount }}</strong>个功能权限
              </span>
              <span class="permission-summary__divider">/</span>
              <span class="permission-summary__item">
                <strong>{{ toRoleRecord(record).dataAuthorityCount }}</strong>个数据权限
              </span>
            </div>
          </template>

          <template v-else-if="column.key === 'actions'">
            <div class="action-cell">
              <a-button type="link" size="small" @click="openEditModal(toRoleRecord(record))">
                编辑权限
              </a-button>
              <a-popconfirm
                overlay-class-name="default-role-delete-popconfirm"
                :overlay-style="deletePopconfirmOverlayStyle"
                ok-text="确认删除"
                cancel-text="取消"
                @confirm="handleDeleteRole(toRoleRecord(record))"
              >
                <template #title>
                  <div class="default-role-delete-popconfirm__title">
                    删除后会自动解除机构员工对该默认角色的绑定，不影响机构已复制的自定义角色，确定删除吗？
                  </div>
                </template>
                <a-button type="link" size="small" danger :loading="deleting">
                  删除
                </a-button>
              </a-popconfirm>
            </div>
          </template>

          <template v-else>
            {{ (toRoleRecord(record) as any)[column.dataIndex as string] ?? '--' }}
          </template>
        </template>
      </a-table>
    </a-card>

    <PlatformModalShell
      v-model:open="modalOpen"
      :width="1160"
      :title="modalMode === 'edit' ? '编辑默认角色' : '新增默认角色'"
      modal-class="default-role-modal"
      @close="closeModal"
    >
      <div class="role-modal-layout">
        <aside class="role-modal-side">
          <div class="role-form-card">
            <a-form layout="vertical">
              <a-form-item label="角色名称" required>
                <a-input
                  v-model:value="formData.roleName"
                  placeholder="请输入角色名称"
                  :maxlength="20"
                />
              </a-form-item>

              <a-form-item label="角色描述" style="margin-bottom: 0;">
                <a-textarea
                  v-model:value="formData.description"
                  :rows="2"
                  :maxlength="200"
                  placeholder="请输入角色描述"
                  show-count
                />
              </a-form-item>
            </a-form>
          </div>

          <div class="role-stat-card">
            <div class="role-stat-card__title">
              权限统计
            </div>
            <div class="role-stat-list">
              <div class="role-stat-item">
                <span class="role-stat-item__label">功能权限</span>
                <span class="role-stat-item__value">{{ selectedPermissionStats.functional }}</span>
              </div>
              <div class="role-stat-item">
                <span class="role-stat-item__label">数据权限</span>
                <span class="role-stat-item__value">{{ selectedPermissionStats.data }}</span>
              </div>
              <div class="role-stat-item">
                <span class="role-stat-item__label">总权限项</span>
                <span class="role-stat-item__value">{{ selectedPermissionStats.functional + selectedPermissionStats.data }}</span>
              </div>
            </div>
            <div class="role-stat-card__hint">
              删除默认角色时，会自动解除机构员工对该角色的绑定，但不会影响机构已创建的自定义角色。
            </div>
          </div>
        </aside>

        <section class="role-modal-main">
          <div class="permission-panel__header">
            <div class="permission-panel__titlebox">
              <div class="permission-panel__title">
                角色权限
              </div>
              <div class="permission-panel__meta">
                与权限管理口径一致：功能权限 + 数据权限
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
            <a-spin :spinning="treeLoading">
              <div v-if="filteredBoxList.length" class="permission-box">
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
                          v-for="authority in getFilteredchildren(child, item)"
                          :key="authority.id"
                          class="permission-row permission-row--authority"
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

              <a-empty v-else :image="simpleImage" description="暂无权限数据" />
            </a-spin>
          </div>
        </section>
      </div>

      <template #footer>
        <div class="role-modal-footer">
          <a-popconfirm
            v-if="modalMode === 'edit'"
            overlay-class-name="default-role-delete-popconfirm"
            :overlay-style="deletePopconfirmOverlayStyle"
            ok-text="确认删除"
            cancel-text="取消"
            @confirm="handleDeleteByModal"
          >
            <template #title>
              <div class="default-role-delete-popconfirm__title">
                删除后会自动解除机构员工对该默认角色的绑定，不影响机构已复制的自定义角色，确定删除吗？
              </div>
            </template>
            <a-button danger ghost :loading="deleting">
              删除
            </a-button>
          </a-popconfirm>
          <a-button @click="closeModal">
            取消
          </a-button>
          <a-button type="primary" :loading="submitting" @click="handleSubmit">
            保存
          </a-button>
        </div>
      </template>
    </PlatformModalShell>
  </div>
</template>

<script setup lang="ts">
import type { TableColumnsType } from 'ant-design-vue'
import {
  DownOutlined,
  PlusOutlined,
  SearchOutlined,
  UpOutlined,
} from '@ant-design/icons-vue'
import { Empty } from 'ant-design-vue'
import { computed, onMounted, reactive, ref } from 'vue'
import {
  createDefaultRoleApi,
  deleteDefaultRoleApi,
  getDefaultRoleDetailApi,
  getDefaultRoleTemplatesApi,
  getRoleMenuIDsApi,
  updateDefaultRoleApi,
  type DefaultRoleTemplateItem,
} from '@/api/platform/default-roles'
import {
  getPermissionTreeApi,
  type PermissionMenuItem,
} from '@/api/platform/permissions'
import {
  type Authority,
  type AuthorityChild,
  type AuthorityGroup,
  useRolePermissions,
} from '@/composables/useRolePermissions'
import PlatformModalShell from '../shared/platform-modal-shell.vue'
import messageService from '@/utils/messageService'

enum PortalEnum {
  INSTITUTION = 2,
  HOSPITAL = 3,
}

interface DefaultRoleRecord {
  roleId: number
  roleName: string
  isDefault: boolean
  permissionCount: number
  functionalAuthorityCount: number
  dataAuthorityCount: number
}

type ModalMode = 'create' | 'edit'

const simpleImage = Empty.PRESENTED_IMAGE_SIMPLE

const currentPortal = ref<PortalEnum>(PortalEnum.INSTITUTION)
const portalOptions = [
  { label: '机构端', value: PortalEnum.INSTITUTION },
  { label: '医院端', value: PortalEnum.HOSPITAL, disabled: true },
]

const loading = ref(false)
const deleting = ref(false)
const roleList = ref<DefaultRoleRecord[]>([])
const keywordInput = ref('')
const searchKeyword = ref('')

const treeLoading = ref(false)
const permissionTypeMap = ref<Map<number, number>>(new Map())

const modalOpen = ref(false)
const modalMode = ref<ModalMode>('create')
const submitting = ref(false)
const deletePopconfirmOverlayStyle = {
  width: '360px',
  maxWidth: 'calc(100vw - 32px)',
}

const initialPermissionData: AuthorityGroup[] = []
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

const formData = reactive<{
  roleId?: number
  roleName: string
  description: string
}>({
  roleId: undefined,
  roleName: '',
  description: '',
})

const columns: TableColumnsType<DefaultRoleRecord> = [
  {
    title: '角色名称',
    dataIndex: 'roleName',
    key: 'roleName',
    width: 200,
  },
  {
    title: '角色类型',
    key: 'isDefault',
    width: 120,
  },
  {
    title: '权限数量',
    key: 'permissionStats',
    width: 300,
  },
  {
    title: '操作',
    key: 'actions',
    width: 180,
    fixed: 'right',
  },
]

const filteredRoleList = computed(() => {
  const keyword = searchKeyword.value.trim().toLowerCase()
  if (!keyword)
    return roleList.value
  return roleList.value.filter(item => item.roleName.toLowerCase().includes(keyword))
})

const selectedPermissionStats = computed(() => {
  return calcPermissionStatsByMenuIDs(collectSelectedPermissionIDs())
})

const hasSelectedPermissions = computed(() => {
  return selectedPermissionStats.value.functional + selectedPermissionStats.value.data > 0
})

const toRoleRecord = (record: Record<string, any>) => record as DefaultRoleRecord

function normalizeMenuID(value: string | number | undefined, fallbackValue = 0) {
  const parsed = Number(value)
  if (Number.isFinite(parsed) && parsed > 0)
    return parsed
  return fallbackValue
}

function normalizeMenuIDs(values: Array<number | string> = []) {
  return Array.from(new Set(
    values
      .map(item => Number(item))
      .filter(item => Number.isFinite(item) && item > 0),
  )).sort((a, b) => a - b)
}

function buildAuthority(node: PermissionMenuItem, fallbackID: number, typeMap: Map<number, number>): Authority {
  const menuID = normalizeMenuID((node as any).id, fallbackID)
  const menuType = Number((node as any).menuType ?? (node as any).type ?? 0)
  typeMap.set(menuID, menuType === 1 ? 1 : 0)
  return {
    id: menuID,
    name: String((node as any).menuName || ''),
    remark: String((node as any).introduce || ''),
    type: menuType === 1 ? 1 : 0,
    mode: 0,
    groupCode: String((node as any).groupCode || ''),
    weight: Number((node as any).weight || 0),
    checked: false,
  }
}

function buildAuthorityChild(
  node: PermissionMenuItem,
  groupIndex: number,
  childIndex: number,
  typeMap: Map<number, number>,
): AuthorityChild {
  const childID = normalizeMenuID((node as any).id, groupIndex * 1000 + childIndex + 1)
  const rawAuthorities = Array.isArray((node as any).children) && (node as any).children.length > 0
    ? (node as any).children
    : [node]

  return {
    id: String(childID),
    menuName: String((node as any).menuName || ''),
    checked: false,
    indeterminate: false,
    children: rawAuthorities.map((authorityNode: PermissionMenuItem, authorityIndex: number) =>
      buildAuthority(authorityNode, childID * 100 + authorityIndex + 1, typeMap),
    ),
  }
}

function mapPermissionTreeToGroups(nodes: PermissionMenuItem[] = [], typeMap = new Map<number, number>()) {
  return nodes.map((groupNode, groupIndex) => {
    const groupID = normalizeMenuID((groupNode as any).id, groupIndex + 1)
    const rawChildren = Array.isArray((groupNode as any).children) && (groupNode as any).children.length > 0
      ? (groupNode as any).children
      : [groupNode]

    return {
      id: String(groupID),
      menuName: String((groupNode as any).menuName || ''),
      checked: false,
      indeterminate: false,
      children: rawChildren.map((childNode: PermissionMenuItem, childIndex: number) =>
        buildAuthorityChild(childNode, groupIndex + 1, childIndex, typeMap),
      ),
    } as AuthorityGroup
  })
}

function calcPermissionStatsByMenuIDs(values: Array<number | string> = []) {
  const menuIDs = normalizeMenuIDs(values)
  let functional = 0
  let data = 0

  menuIDs.forEach((menuID) => {
    const menuType = Number(permissionTypeMap.value.get(menuID) ?? 0)
    if (menuType === 1)
      data += 1
    else
      functional += 1
  })

  return { functional, data }
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

  return Array.from(selectedIDs).sort((a, b) => a - b)
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

function mapRoleTemplateItem(item: DefaultRoleTemplateItem): DefaultRoleRecord {
  const roleIDs = normalizeMenuIDs(item.roleIds || [])
  let functionalAuthorityCount = Number(item.functionalAuthorityCount || 0)
  let dataAuthorityCount = Number(item.dataAuthorityCount || 0)

  if (functionalAuthorityCount <= 0 && dataAuthorityCount <= 0 && roleIDs.length > 0) {
    const stats = calcPermissionStatsByMenuIDs(roleIDs)
    functionalAuthorityCount = stats.functional
    dataAuthorityCount = stats.data
  }

  if (functionalAuthorityCount <= 0 && dataAuthorityCount <= 0 && roleIDs.length > 0)
    functionalAuthorityCount = roleIDs.length

  return {
    roleId: Number(item.roleId || 0),
    roleName: String(item.roleName || '').trim(),
    isDefault: Boolean(item.isDefault),
    permissionCount: functionalAuthorityCount + dataAuthorityCount,
    functionalAuthorityCount,
    dataAuthorityCount,
  }
}

async function loadRoleList() {
  loading.value = true
  try {
    const res = await getDefaultRoleTemplatesApi({ roleType: currentPortal.value })
    if (res.code !== 200) {
      messageService.error(res.message || '加载默认角色失败')
      roleList.value = []
      return
    }

    const rows = Array.isArray(res.result) ? res.result.map(mapRoleTemplateItem) : []
    roleList.value = rows.sort((left, right) => right.roleId - left.roleId)
  }
  catch (error: any) {
    messageService.error(error?.message || '加载默认角色失败')
    roleList.value = []
  }
  finally {
    loading.value = false
  }
}

async function loadPermissionTree() {
  treeLoading.value = true
  try {
    const res = await getPermissionTreeApi({ ownType: currentPortal.value })
    if (res.code !== 200) {
      messageService.error(res.message || '加载权限树失败')
      updateData([])
      permissionTypeMap.value = new Map()
      return
    }

    const typeMap = new Map<number, number>()
    const groups = mapPermissionTreeToGroups(Array.isArray(res.result) ? res.result : [], typeMap)
    permissionTypeMap.value = typeMap
    updateData(groups)
  }
  catch (error: any) {
    messageService.error(error?.message || '加载权限树失败')
    updateData([])
    permissionTypeMap.value = new Map()
  }
  finally {
    treeLoading.value = false
  }
}

async function loadPortalData() {
  await loadPermissionTree()
  await loadRoleList()
}

async function handlePortalChange(value: number) {
  currentPortal.value = value as PortalEnum
  keywordInput.value = ''
  searchKeyword.value = ''
  searchValue.value = ''
  await loadPortalData()
}

function handleSearch() {
  searchKeyword.value = keywordInput.value.trim()
}

function handleReset() {
  keywordInput.value = ''
  searchKeyword.value = ''
}

function resetForm() {
  formData.roleId = undefined
  formData.roleName = ''
  formData.description = ''
  clearAllSelected()
  searchValue.value = ''
}

function closeModal() {
  modalOpen.value = false
}

function openCreateModal() {
  modalMode.value = 'create'
  resetForm()
  modalOpen.value = true
}

async function openEditModal(record: DefaultRoleRecord) {
  modalMode.value = 'edit'
  resetForm()
  treeLoading.value = true
  try {
    const [detailRes, menuRes] = await Promise.all([
      getDefaultRoleDetailApi({ roleId: record.roleId }),
      getRoleMenuIDsApi({ roleId: record.roleId, ownType: currentPortal.value }),
    ])
    if (detailRes.code !== 200) {
      messageService.error(detailRes.message || '加载角色详情失败')
      return
    }
    if (menuRes.code !== 200) {
      messageService.error(menuRes.message || '加载角色权限失败')
      return
    }

    formData.roleId = record.roleId
    formData.roleName = String(detailRes.result?.roleName || record.roleName || '').trim()
    formData.description = String(detailRes.result?.description || '').trim()

    const allowedSet = new Set(collectVisibleAuthorityIDs())
    const checkedMenuIDs = normalizeMenuIDs(Array.isArray(menuRes.result) ? menuRes.result : [])
      .filter(menuID => allowedSet.has(menuID))
    setDefaultCheckedByIds(checkedMenuIDs)
    modalOpen.value = true
  }
  catch (error: any) {
    messageService.error(error?.message || '加载角色详情失败')
  }
  finally {
    treeLoading.value = false
  }
}

function resolveDeleteSuccessText(detachedUsers: number) {
  if (detachedUsers > 0)
    return `删除成功，已自动解除 ${detachedUsers} 个员工绑定`
  return '删除成功'
}

async function handleDeleteRole(record: DefaultRoleRecord) {
  const roleID = Number(record.roleId || 0)
  if (!roleID)
    return

  deleting.value = true
  try {
    const res = await deleteDefaultRoleApi({ roleId: roleID })
    if (res.code !== 200) {
      messageService.error(res.message || '删除默认角色失败')
      return
    }

    const detachedUsers = Number(res.result?.detachedUsers || 0)
    messageService.success(resolveDeleteSuccessText(detachedUsers))

    if (modalOpen.value && Number(formData.roleId || 0) === roleID)
      closeModal()

    await loadRoleList()
  }
  catch (error: any) {
    messageService.error(error?.message || '删除默认角色失败')
  }
  finally {
    deleting.value = false
  }
}

async function handleDeleteByModal() {
  const roleID = Number(formData.roleId || 0)
  if (!roleID) {
    messageService.error('角色ID缺失，请重新操作')
    return
  }

  await handleDeleteRole({
    roleId: roleID,
    roleName: formData.roleName,
    isDefault: false,
    permissionCount: 0,
    functionalAuthorityCount: 0,
    dataAuthorityCount: 0,
  })
}

async function handleSubmit() {
  const roleName = String(formData.roleName || '').trim()
  if (!roleName) {
    messageService.error('请填写角色名称')
    return
  }

  const menuIDs = collectSelectedPermissionIDs()
  submitting.value = true
  try {
    if (modalMode.value === 'create') {
      const res = await createDefaultRoleApi({
        roleName,
        description: String(formData.description || '').trim(),
        menuIds: menuIDs,
        roleType: currentPortal.value,
      })
      if (res.code !== 200) {
        messageService.error(res.message || '新增默认角色失败')
        return
      }
      messageService.success('新增默认角色成功')
    }
    else {
      const roleID = Number(formData.roleId || 0)
      if (!roleID) {
        messageService.error('角色ID缺失，请重新操作')
        return
      }
      const res = await updateDefaultRoleApi({
        roleId: roleID,
        roleName,
        description: String(formData.description || '').trim(),
        menuIds: menuIDs,
      })
      if (res.code !== 200) {
        messageService.error(res.message || '更新默认角色失败')
        return
      }
      messageService.success('更新默认角色成功')
    }

    modalOpen.value = false
    await loadRoleList()
  }
  catch (error: any) {
    messageService.error(error?.message || '保存默认角色失败')
  }
  finally {
    submitting.value = false
  }
}

onMounted(async () => {
  await loadPortalData()
})
</script>

<style scoped lang="less">
.default-role-page {
  .toolbar {
    margin-bottom: 16px;
  }

  .ellipsis-text {
    display: inline-block;
    max-width: 160px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    vertical-align: middle;
  }

  .action-cell {
    display: inline-flex;
    align-items: center;
    gap: 2px;
  }

  .permission-summary {
    display: inline-flex;
    align-items: center;
    color: #475569;
    font-size: 13px;
    line-height: 20px;

    .permission-summary__item strong {
      color: #1677ff;
      font-size: 14px;
      margin-right: 2px;
    }

    .permission-summary__divider {
      margin: 0 8px;
      color: #94a3b8;
    }
  }
}

.role-modal-layout {
  display: grid;
  grid-template-columns: 320px minmax(0, 1fr);
  gap: 12px;
  height: 480px;
  align-items: stretch;
}

.role-modal-side {
  display: flex;
  flex-direction: column;
  gap: 10px;
  height: 100%;
  min-height: 0;
}

.role-form-card,
.role-stat-card,
.role-modal-main {
  border: 1px solid #e8edf5;
  border-radius: 18px;
  background: #fff;
}

.role-form-card {
  padding: 12px 12px 0;
  background: linear-gradient(180deg, #fbfdff 0%, #ffffff 100%);
}

.role-stat-card {
  padding: 12px;
  flex: 1;
  min-height: 0;
}

.role-stat-card__title {
  color: #1f2937;
  font-size: 14px;
  font-weight: 600;
  line-height: 22px;
}

.role-stat-list {
  display: grid;
  grid-template-columns: 1fr;
  gap: 8px;
  margin-top: 10px;
}

.role-stat-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 9px 10px;
  border-radius: 10px;
  background: #f8fafc;
}

.role-stat-item__label {
  color: #475569;
  font-size: 13px;
  line-height: 20px;
}

.role-stat-item__value {
  color: #1677ff;
  font-size: 16px;
  font-weight: 700;
  line-height: 22px;
}

.role-stat-card__hint {
  margin-top: 10px;
  color: #64748b;
  font-size: 12px;
  line-height: 19px;
}

.role-modal-main {
  display: flex;
  flex-direction: column;
  min-width: 0;
  height: 100%;
  min-height: 0;
  overflow: hidden;
}

.permission-panel__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 12px 18px 8px;
  border-bottom: 1px solid #eef2f7;
}

.permission-panel__titlebox {
  min-width: 0;
}

.permission-panel__title {
  color: #262626;
  font-size: 14px;
  font-weight: 600;
  line-height: 20px;
}

.permission-panel__meta {
  margin-top: 2px;
  color: #8c8c8c;
  font-size: 12px;
  line-height: 16px;
}

.permission-panel__link {
  flex-shrink: 0;
  padding-inline: 0;
  height: 24px;
}

.permission-panel__toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 10px 18px 12px 1px;
  border-bottom: 1px solid #f2f4f7;
}

.permission-panel__toolbar :deep(.ant-input-affix-wrapper) {
  width: 320px;
}

.permission-panel__body {
  flex: 1;
  min-height: 0;
  padding: 0 18px 18px;
  overflow-y: auto;
  overflow-x: hidden;
  overscroll-behavior: contain;
  scrollbar-width: thin;
  scrollbar-color: rgba(100, 116, 139, 0.4) transparent;

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
    background: rgba(100, 116, 139, 0.36);
    background-clip: padding-box;
  }

  &::-webkit-scrollbar-thumb:hover {
    background: rgba(71, 85, 105, 0.52);
    background-clip: padding-box;
  }
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

.role-modal-footer {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 8px;
}

:deep(.default-role-delete-popconfirm .ant-popover-inner) {
  max-width: 520px;
}

:deep(.default-role-delete-popconfirm .ant-popover-message-title) {
  white-space: normal;
  line-height: 22px;
}

.default-role-delete-popconfirm__title {
  max-width: 100%;
  white-space: normal;
  word-break: break-word;
  line-height: 22px;
}

@media (max-width: 1200px) {
  .role-modal-layout {
    grid-template-columns: 1fr;
    height: auto;
  }

  .role-modal-main {
    min-height: 480px;
  }
}
</style>
