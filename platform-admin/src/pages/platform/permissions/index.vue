<template>
  <div class="permission-management">
    <a-card title="权限管理" :bordered="false">
      <div class="search-bar">
        <a-space>
          <a-segmented
            v-model:value="currentPortal"
            :options="portalOptions"
            @change="handlePortalChange"
          />
          <a-input
            v-model:value="searchKeyword"
            placeholder="请输入菜单名称或标识"
            allow-clear
            style="width: 260px"
          />
          <a-button type="primary" @click="handleSearch">
            搜索
          </a-button>
          <a-button @click="handleReset">
            重置
          </a-button>
          <a-button v-if="hasPermission(AccessEnum.menuPermissions_add)" type="primary" @click="handleAddRoot">
            <template #icon>
              <PlusOutlined />
            </template>
            新增
          </a-button>
          <a-button @click="toggleExpandAll">
            展开 / 折叠
          </a-button>
        </a-space>
      </div>

      <a-table
        v-model:expandedRowKeys="expandedRowKeys"
        :columns="columns"
        :data-source="tableData"
        :loading="loading"
        :pagination="false"
        :scroll="{ x: 1230 }"
        table-layout="fixed"
        size="middle"
        row-key="id"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'menuName'">
            <span class="ellipsis-text" :title="record.menuName || '--'">
              {{ record.menuName || '--' }}
            </span>
          </template>

          <template v-else-if="column.key === 'menuCode'">
            <span class="ellipsis-text" :title="record.menuCode || '--'">
              {{ record.menuCode || '--' }}
            </span>
          </template>

          <template v-else-if="column.key === 'levelLabel'">
            <a-tag :color="getLevelTagColor(normalizeRecord(record))">
              {{ getLevelLabel(normalizeRecord(record)) }}
            </a-tag>
          </template>

          <template v-else-if="column.key === 'introduce'">
            <span class="ellipsis-text" :title="getIntroduceText(normalizeRecord(record))">
              {{ getIntroduceText(normalizeRecord(record)) }}
            </span>
          </template>

          <template v-else-if="column.key === 'actions'">
            <a-space class="action-cell">
              <a-button
                v-if="hasPermission(AccessEnum.menuPermissions_update)"
                type="link"
                size="small"
                @click="openDrawer('edit', normalizeRecord(record))"
              >
                修改
              </a-button>
              <a-button
                v-if="hasPermission(AccessEnum.menuPermissions_add) && Number(normalizeRecord(record).depth || 1) < 3"
                type="link"
                size="small"
                @click="handleAddChild(normalizeRecord(record))"
              >
                新增
              </a-button>
              <a-popconfirm
                v-if="hasPermission(AccessEnum.menuPermissions_delete)"
                title="确定要删除该权限吗？"
                @confirm="handleDeleteByRecord(normalizeRecord(record))"
              >
                <a-button type="link" size="small" danger>
                  删除
                </a-button>
              </a-popconfirm>
            </a-space>
          </template>

          <template v-else>
            {{ getColumnValue(record, column.dataIndex) ?? '--' }}
          </template>
        </template>
      </a-table>
    </a-card>

    <a-modal
      v-model:open="drawerVisible"
      centered
      destroy-on-close
      :keyboard="false"
      :closable="false"
      :mask-closable="false"
      :width="760"
      class="createStu-modal-content-box permission-form-modal"
      @cancel="closeFormModal"
    >
      <template #title>
        <div class="permission-form-modal__titlebar">
          <span>{{ formMode === 'edit' ? '编辑权限' : '新增权限' }}</span>
          <a-button type="text" class="close-btn" @click="closeFormModal">
            <template #icon>
              <CloseOutlined class="close-icon" />
            </template>
          </a-button>
        </div>
      </template>

      <div class="permission-form-modal__body scrollbar">
        <div class="permission-form-card">
          <a-form :model="formData" layout="vertical" class="permission-form">
            <div class="permission-form-grid">
              <div class="permission-form-grid__item permission-form-grid__item--full">
                <a-form-item label="权限名称" required>
                  <a-input v-model:value="formData.menuName" placeholder="请输入权限名称" />
                </a-form-item>
              </div>

              <div class="permission-form-grid__item permission-form-grid__item--full">
                <a-form-item label="权限编码" required>
                  <a-input
                    v-model:value="formData.menuCode"
                    :disabled="formMode === 'edit'"
                    placeholder="例如 systemModel:menuPermissions"
                  />
                </a-form-item>
              </div>

              <div class="permission-form-grid__item">
                <a-form-item label="排序">
                  <a-input-number v-model:value="formData.sort" :min="0" style="width: 100%" />
                </a-form-item>
              </div>

              <div class="permission-form-grid__item">
                <a-form-item label="权重">
                  <a-input-number v-model:value="formData.weight" :min="0" style="width: 100%" />
                </a-form-item>
              </div>

              <div class="permission-form-grid__item">
                <a-form-item label="父级菜单">
                  <a-input :value="parentName" disabled />
                </a-form-item>
              </div>

              <div class="permission-form-grid__item">
                <a-form-item label="所属端口">
                  <a-tag :color="currentPortal === PortalEnum.PLATFORM ? 'blue' : 'green'">
                    {{ currentPortal === PortalEnum.PLATFORM ? '平台端' : '机构端' }}
                  </a-tag>
                </a-form-item>
              </div>

              <div class="permission-form-grid__item permission-form-grid__item--full">
                <a-form-item label="备注">
                  <a-input v-model:value="formData.remark" placeholder="请输入备注" />
                </a-form-item>
              </div>

              <div class="permission-form-grid__item permission-form-grid__item--full">
                <a-form-item label="权限描述">
                  <a-textarea
                    v-model:value="formData.introduce"
                    :rows="4"
                    placeholder="请输入权限描述"
                  />
                </a-form-item>
              </div>
            </div>
          </a-form>
        </div>
      </div>

      <template #footer>
        <div class="permission-form-modal__footer">
          <div class="permission-form-modal__footer-left">
            <a-popconfirm
              v-if="formMode === 'edit' && hasPermission(AccessEnum.menuPermissions_delete)"
              title="确定要删除该权限吗？"
              @confirm="handleDelete"
            >
              <a-button danger ghost>
                删除
              </a-button>
            </a-popconfirm>
          </div>

          <div class="permission-form-modal__footer-right">
            <a-button @click="closeFormModal">
              取消
            </a-button>
            <a-button
              v-if="hasPermission(formMode === 'edit' ? AccessEnum.menuPermissions_update : AccessEnum.menuPermissions_add)"
              type="primary"
              @click="handleSubmit"
            >
              保存
            </a-button>
          </div>
        </div>
      </template>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import type { TableColumnsType } from 'ant-design-vue'
import { CloseOutlined, PlusOutlined } from '@ant-design/icons-vue'
import {
  createPermissionApi,
  deletePermissionApi,
  getPermissionTreeApi,
  updatePermissionApi,
  type PermissionMenuItem,
  type PermissionMutationPayload,
} from '@/api/platform/permissions'
import messageService from '@/utils/messageService'
import { AccessEnum } from '~@/utils/constant'

enum PortalEnum {
  PLATFORM = 0,
  INSTITUTION = 2,
}

interface PermissionRecord extends PermissionMenuItem {
  children?: PermissionRecord[]
  depth?: number
}

type FormMode = 'add' | 'edit' | null

const { hasAccess } = useAccess()

const hasPermission = (permissionCode: string) => {
  return hasAccess([permissionCode, AccessEnum.superAdmin])
}

const columns: TableColumnsType<PermissionRecord> = [
  {
    title: '菜单名称',
    dataIndex: 'menuName',
    key: 'menuName',
    width: 240,
  },
  {
    title: '权限标识',
    dataIndex: 'menuCode',
    key: 'menuCode',
    width: 220,
  },
  {
    title: '层级',
    key: 'levelLabel',
    width: 100,
  },
  {
    title: '排序',
    dataIndex: 'sort',
    key: 'sort',
    width: 80,
  },
  {
    title: '权重',
    dataIndex: 'weight',
    key: 'weight',
    width: 80,
  },
  {
    title: '描述',
    dataIndex: 'introduce',
    key: 'introduce',
    width: 320,
  },
  {
    title: '操作',
    key: 'actions',
    width: 190,
    fixed: 'right',
  },
]

const treeData = ref<PermissionRecord[]>([])
const searchKeyword = ref('')
const expandedRowKeys = ref<number[]>([])
const loading = ref(false)

const currentPortal = ref<PortalEnum>(PortalEnum.INSTITUTION)
const portalOptions = [
  { label: '平台端', value: PortalEnum.PLATFORM, disabled: true },
  { label: '机构端', value: PortalEnum.INSTITUTION },
]

const formMode = ref<FormMode>(null)
const drawerVisible = ref(false)
const parentName = ref<string>('根菜单')

const formData = reactive<PermissionMutationPayload>({
  id: undefined,
  pid: 0,
  menuName: '',
  menuCode: '',
  sort: 0,
  weight: 0,
  remark: '',
  introduce: '',
  ownType: PortalEnum.INSTITUTION,
})

const normalizeTree = (list: PermissionMenuItem[] = [], depth = 1): PermissionRecord[] => {
  return list.map((item) => {
    const children = Array.isArray(item.children) ? normalizeTree(item.children, depth + 1) : []
    return {
      ...item,
      sort: Number(item.sort || 0),
      weight: Number(item.weight || 0),
      pid: Number(item.pid || 0),
      ownType: Number(item.ownType || currentPortal.value),
      depth,
      children,
    }
  })
}

const normalizeRecord = (record: Record<string, any>) => record as PermissionRecord

const getColumnValue = (record: Record<string, any>, dataIndex: string | number | readonly (string | number)[] | undefined) => {
  if (typeof dataIndex === 'string' || typeof dataIndex === 'number')
    return record[dataIndex]
  return undefined
}

const getIntroduceText = (record: PermissionRecord) => {
  const introduce = String(record.introduce || '').trim()
  if (introduce)
    return introduce

  return '--'
}

const tableData = computed<PermissionRecord[]>(() => {
  const keyword = searchKeyword.value.trim().toLowerCase()
  if (!keyword)
    return treeData.value

  const filterTree = (list: PermissionRecord[]): PermissionRecord[] => {
    const result: PermissionRecord[] = []
    list.forEach((item) => {
      const children = item.children ? filterTree(item.children) : []
      const hit
        = String(item.menuName || '').toLowerCase().includes(keyword)
          || String(item.menuCode || '').toLowerCase().includes(keyword)

      if (hit || children.length) {
        result.push({
          ...item,
          children,
        })
      }
    })
    return result
  }

  return filterTree(treeData.value)
})

const loadTree = async () => {
  loading.value = true
  try {
    const res = await getPermissionTreeApi({ ownType: currentPortal.value })
    if (res.code !== 200) {
      messageService.error(res.message || '加载权限树失败')
      return
    }
    treeData.value = normalizeTree(Array.isArray(res.result) ? res.result : [])
  }
  catch (error: any) {
    messageService.error(error?.message || '加载权限树失败')
  }
  finally {
    loading.value = false
  }
}

const handlePortalChange = async (value: number) => {
  currentPortal.value = value as PortalEnum
  searchKeyword.value = ''
  expandedRowKeys.value = []
  await loadTree()
}

const resetForm = () => {
  formData.id = undefined
  formData.pid = 0
  formData.menuName = ''
  formData.menuCode = ''
  formData.sort = 0
  formData.weight = 0
  formData.remark = ''
  formData.introduce = ''
  formData.ownType = currentPortal.value
  parentName.value = '根菜单'
}

const handleAddRoot = () => {
  resetForm()
  formMode.value = 'add'
  formData.pid = 0
  parentName.value = '根菜单'
  drawerVisible.value = true
}

const handleAddChild = (parent?: PermissionRecord) => {
  const target = parent
  if (!target) {
    messageService.warning('请先选择父级节点')
    return
  }

  resetForm()
  formMode.value = 'add'
  formData.pid = Number(target.id)
  parentName.value = target.menuName
  drawerVisible.value = true
}

const findNodeById = (list: PermissionRecord[], id: number): PermissionRecord | null => {
  for (const item of list) {
    if (Number(item.id) === id)
      return item
    if (item.children && item.children.length) {
      const found = findNodeById(item.children, id)
      if (found)
        return found
    }
  }
  return null
}

const openDrawer = (mode: FormMode, node?: PermissionRecord) => {
  if (mode === 'edit') {
    if (!node) {
      messageService.warning('请先选择要编辑的权限')
      return
    }

    formData.id = Number(node.id)
    formData.pid = Number(node.pid || 0)
    if (Number(node.pid || 0) === 0) {
      parentName.value = '根菜单'
    }
    else {
      const parentNode = findNodeById(treeData.value, Number(node.pid))
      parentName.value = parentNode ? parentNode.menuName : ''
    }
    formData.menuName = node.menuName
    formData.menuCode = node.menuCode
    formData.sort = Number(node.sort || 0)
    formData.weight = Number(node.weight || 0)
    formData.remark = node.remark || ''
    formData.introduce = node.introduce || ''
    formData.ownType = Number(node.ownType || currentPortal.value)
  }
  formMode.value = mode
  drawerVisible.value = true
}

const closeFormModal = () => {
  drawerVisible.value = false
}

const handleSearch = () => {
}

const handleReset = async () => {
  searchKeyword.value = ''
  await loadTree()
}

const collectIds = (list: PermissionRecord[]): number[] => {
  const ids: number[] = []
  list.forEach((item) => {
    ids.push(Number(item.id))
    if (item.children && item.children.length)
      ids.push(...collectIds(item.children))
  })
  return ids
}

const toggleExpandAll = () => {
  if (expandedRowKeys.value.length) {
    expandedRowKeys.value = []
  }
  else {
    expandedRowKeys.value = collectIds(tableData.value)
  }
}

const handleSubmit = async () => {
  if (!String(formData.menuName || '').trim() || !String(formData.menuCode || '').trim()) {
    messageService.error('请填写权限名称和编码')
    return
  }

  const payload: PermissionMutationPayload = {
    id: formData.id || undefined,
    pid: Number(formData.pid || 0),
    menuName: String(formData.menuName || '').trim(),
    menuCode: String(formData.menuCode || '').trim(),
    sort: Number(formData.sort || 0),
    weight: Number(formData.weight || 0),
    remark: String(formData.remark || '').trim(),
    introduce: String(formData.introduce || '').trim(),
    ownType: currentPortal.value,
  }

  try {
    if (formMode.value === 'edit' && payload.id) {
      const res = await updatePermissionApi(payload as PermissionMutationPayload & { id: number })
      if (res.code !== 200) {
        messageService.error(res.message || '更新失败')
        return
      }
      messageService.success('更新成功')
    }
    else {
      const res = await createPermissionApi(payload)
      if (res.code !== 200) {
        messageService.error(res.message || '新增失败')
        return
      }
      messageService.success('新增成功')
    }
    drawerVisible.value = false
    await loadTree()
  }
  catch (error: any) {
    messageService.error(error?.message || '保存失败')
  }
}

const handleDelete = async () => {
  if (!formData.id)
    return
  try {
    const res = await deletePermissionApi({ id: Number(formData.id) })
    if (res.code !== 200) {
      messageService.error(res.message || '删除失败')
      return
    }
    messageService.success('删除成功')
    formMode.value = null
    drawerVisible.value = false
    await loadTree()
  }
  catch (error: any) {
    messageService.error(error?.message || '删除失败')
  }
}

const handleDeleteByRecord = async (record: PermissionRecord) => {
  try {
    const res = await deletePermissionApi({ id: Number(record.id) })
    if (res.code !== 200) {
      messageService.error(res.message || '删除失败')
      return
    }
    messageService.success('删除成功')
    await loadTree()
  }
  catch (error: any) {
    messageService.error(error?.message || '删除失败')
  }
}

const getLevelLabel = (record: PermissionRecord) => {
  const level = Number(record.depth || 1)
  if (level <= 1)
    return '1级菜单'
  if (level === 2)
    return '2级界面'
  return '3级权限'
}

const getLevelTagColor = (record: PermissionRecord) => {
  const level = Number(record.depth || 1)
  if (level <= 1)
    return 'blue'
  if (level === 2)
    return 'green'
  return 'orange'
}

onMounted(() => {
  loadTree()
})
</script>

<style scoped lang="less">
.permission-management {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial,
    'Noto Sans', sans-serif, 'Apple Color Emoji', 'Segoe UI Emoji', 'Segoe UI Symbol',
    'Noto Color Emoji';
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;

  .search-bar {
    margin-bottom: 16px;
  }

  .ellipsis-text {
    display: inline-block;
    max-width: 100%;
    overflow: hidden;
    color: #000;
    text-overflow: ellipsis;
    white-space: nowrap;
    vertical-align: bottom;
  }

  .action-cell {
    white-space: nowrap;
  }

  :deep(.createStu-modal-content-box.permission-form-modal .ant-modal-content) {
    border-radius: 22px;
    overflow: hidden;
    box-shadow: 0 18px 46px rgba(15, 23, 42, 0.14);
  }

  :deep(.createStu-modal-content-box.permission-form-modal .ant-modal-header) {
    padding: 24px 28px 14px;
    margin-bottom: 0;
    border-bottom: none;
  }

  :deep(.createStu-modal-content-box.permission-form-modal .ant-modal-body) {
    padding: 0 28px 0;
  }

  :deep(.createStu-modal-content-box.permission-form-modal .ant-modal-footer) {
    padding: 18px 28px 24px;
    border-top: none;
  }

  .permission-form-modal__titlebar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    width: 100%;
    color: #1f2329;
    font-size: 20px;
    font-weight: 700;
    line-height: 32px;
  }

  .permission-form-modal__body {
    max-height: calc(100vh - 220px);
    padding-top: 8px;
    overflow-y: auto;
    overflow-x: hidden;
  }

  .permission-form-card {
    border: 1px solid #e8edf5;
    border-radius: 18px;
    background:
      linear-gradient(180deg, rgba(22, 119, 255, 0.06) 0%, rgba(22, 119, 255, 0) 132px),
      #fff;
    box-shadow: 0 14px 32px rgba(15, 23, 42, 0.06);
    padding: 20px 22px 4px;
  }

  .permission-form-grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 0 24px;
  }

  .permission-form-grid__item {
    min-width: 0;
  }

  .permission-form-grid__item--full {
    grid-column: 1 / -1;
  }

  .permission-form {
    :deep(.ant-form-item) {
      margin-bottom: 18px;
    }
  }

  .permission-form-modal__footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
  }

  .permission-form-modal__footer-left,
  .permission-form-modal__footer-right {
    display: flex;
    align-items: center;
    gap: 12px;
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

  :deep(.ant-table-tbody > tr > td) {
    color: #000;
  }

  :deep(.ant-table-thead > tr > th) {
    color: #000;
  }

  :deep(.ant-card),
  :deep(.ant-card-head),
  :deep(.ant-card-body),
  :deep(.ant-space),
  :deep(.ant-segmented),
  :deep(.ant-input),
  :deep(.ant-btn),
  :deep(.ant-table),
  :deep(.ant-tag),
  :deep(.ant-drawer),
  :deep(.ant-form),
  :deep(.ant-form-item),
  :deep(.ant-form-item-label > label) {
    font-family: inherit;
  }

  @media (max-width: 768px) {
    .permission-form-grid {
      grid-template-columns: 1fr;
      gap: 0;
    }

    .permission-form-grid__item--full {
      grid-column: auto;
    }

    .permission-form-modal__footer {
      flex-direction: column;
      align-items: stretch;
    }

    .permission-form-modal__footer-left,
    .permission-form-modal__footer-right {
      justify-content: flex-end;
    }
  }
}
</style>
