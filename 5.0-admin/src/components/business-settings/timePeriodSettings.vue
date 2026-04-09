<script setup lang="ts">
import { PlusOutlined } from '@ant-design/icons-vue'
import { Modal } from 'ant-design-vue'
import type { TableColumnType } from 'ant-design-vue'
import { debounce } from 'lodash-es'
import UnifiedPeriodGroupModal from '@/components/business-settings/unified-period-group-modal.vue'
import { repairInstPeriodVersionsApi, type InstConfig, setInstConfigApi } from '@/api/common/config'
import { getUserListApi } from '@/api/internal-manage/staff-manage'
import { useUserStore } from '@/stores/user'
import {
  DEFAULT_UNIFIED_TIME_PERIOD_CONFIG,
  configGroupsSorted,
  parseUnifiedTimePeriodConfig,
  slotCountActive,
  type UnifiedPeriodGroup,
  type UnifiedPeriodSlot,
  type UnifiedTimePeriodConfig,
} from '@/utils/unified-time-period'
import messageService from '@/utils/messageService'

const userStore = useUserStore()
const loading = ref(false)
const repairing = ref(false)
const quickUnifiedEnabled = ref(false)

const groupModalOpen = ref(false)
const groupModalMode = ref<'create' | 'edit'>('edit')
const editingGroupId = ref<string | null>(null)
const effectiveRuleText = '生效规则：历史周不会被覆盖；如果本周没有老师排课，新时段从本周生效，否则从下周生效。'

const periodGroups = computed<UnifiedPeriodGroup[]>(() => {
  const parsed = parseUnifiedTimePeriodConfig(userStore.instConfig?.unifiedTimePeriodJson)
  const cfg = parsed ?? DEFAULT_UNIFIED_TIME_PERIOD_CONFIG
  return configGroupsSorted(cfg)
})

const columns: TableColumnType<UnifiedPeriodGroup>[] = [
  { title: '时段名称', dataIndex: 'name', key: 'name', ellipsis: true },
  { title: '节次', key: 'slots', width: 150 },
  { title: '时间范围', key: 'span', width: 200 },
  { title: '关联老师', key: 'teachers', width: 200, ellipsis: true },
  { title: '状态', key: 'status', width: 120 },
  { title: '操作', key: 'action', width: 236 },
]

type UnifiedPeriodGroupLike = UnifiedPeriodGroup | Record<string, any>

function normalizedSlots(group: UnifiedPeriodGroupLike): UnifiedPeriodSlot[] {
  return Array.isArray(group.slots)
    ? group.slots.map((slot: any) => ({
        id: String(slot?.id || ''),
        index: Number(slot?.index || 0),
        name: String(slot?.name || ''),
        start: String(slot?.start || ''),
        end: String(slot?.end || ''),
        enabled: slot?.enabled !== false,
      }))
    : []
}

function normalizedBoundTeachers(group: UnifiedPeriodGroupLike) {
  return Array.isArray(group.boundTeachers)
    ? group.boundTeachers.map((teacher: any) => ({
        id: String(teacher?.id || ''),
        name: String(teacher?.name || ''),
      }))
    : []
}

function sortSlots(slots: UnifiedPeriodSlot[]) {
  return [...slots].sort((a, b) => a.index - b.index)
}

function groupTimeSpan(g: UnifiedPeriodGroupLike): string {
  const active = sortSlots(normalizedSlots(g)).filter(s => s.enabled !== false)
  if (!active.length)
    return '—'
  return `${active[0].start} ~ ${active[active.length - 1].end}`
}

function activeSlotCount(g: UnifiedPeriodGroupLike): number {
  return normalizedSlots(g).filter(slot => slot.enabled !== false).length
}

function slotsSummary(g: UnifiedPeriodGroupLike): string {
  const total = normalizedSlots(g).length
  const active = activeSlotCount(g)
  return `${active} / ${total} 节启用`
}

function cloneConfig(c: UnifiedTimePeriodConfig): UnifiedTimePeriodConfig {
  return {
    version: c.version,
    groups: c.groups.map(g => ({
      ...g,
      slots: g.slots.map(s => ({ ...s })),
      boundTeachers: (g.boundTeachers || []).map(t => ({ ...t })),
    })),
  }
}

function formatBoundTeachersSummary(g: UnifiedPeriodGroupLike): string {
  const list = normalizedBoundTeachers(g)
  if (!list.length)
    return '—'
  return list.map(t => t.name).join('、')
}

function hasBoundTeachers(g: UnifiedPeriodGroupLike): boolean {
  return normalizedBoundTeachers(g).length > 0
}

function loadBaseConfig(): UnifiedTimePeriodConfig {
  const raw = userStore.instConfig?.unifiedTimePeriodJson
  const parsed = parseUnifiedTimePeriodConfig(raw)
  return cloneConfig(parsed ?? DEFAULT_UNIFIED_TIME_PERIOD_CONFIG)
}

async function refreshFromServer() {
  loading.value = true
  try {
    await userStore.getInstConfig()
    quickUnifiedEnabled.value = Boolean(userStore.instConfig?.enableQuickUnifiedPeriod)
  }
  catch (e) {
    console.error('load inst config failed', e)
    messageService.error('获取机构配置失败')
  }
  finally {
    loading.value = false
  }
}

onMounted(() => {
  void refreshFromServer()
})

watch(
  () => userStore.instConfig?.enableQuickUnifiedPeriod,
  (v) => {
    if (userStore.instConfig && typeof v !== 'undefined')
      quickUnifiedEnabled.value = Boolean(v)
  },
)

async function onQuickUnifiedChange(checked: boolean) {
  const prev = quickUnifiedEnabled.value
  quickUnifiedEnabled.value = checked
  try {
    await setInstConfigApi({
      ...(userStore.instConfig as InstConfig),
      enableQuickUnifiedPeriod: checked,
    })
    await userStore.getInstConfig()
    messageService.success('已保存')
  }
  catch (e) {
    console.error(e)
    messageService.error('保存失败')
    quickUnifiedEnabled.value = prev
  }
}

function openCreateGroup() {
  groupModalMode.value = 'create'
  editingGroupId.value = null
  groupModalOpen.value = true
}

function openEditGroup(id: string) {
  groupModalMode.value = 'edit'
  editingGroupId.value = id
  groupModalOpen.value = true
}

function onGroupModalSaved() {
  void refreshFromServer()
}

function repairPeriodVersions() {
  Modal.confirm({
    title: '一键修复已排课周',
    centered: true,
    content: '会把落在已排课周上的时段版本自动顺延到该组老师的第一个空周，历史和已排课周保持不变。确定继续吗？',
    okText: '开始修复',
    cancelText: '取消',
    async onOk() {
      repairing.value = true
      try {
        const res = await repairInstPeriodVersionsApi()
        const repairedCount = Number(res.result?.repairedVersions || 0)
        if (repairedCount > 0)
          messageService.success(`修复完成，已顺延 ${repairedCount} 个时段版本`)
        else
          messageService.success('未发现需要修复的已排课周')
        await refreshFromServer()
      }
      catch (e) {
        console.error('repair inst period versions failed', e)
        messageService.error('修复失败')
        throw e
      }
      finally {
        repairing.value = false
      }
    },
  })
}

/** —— 操作列：关联老师（仅存 unifiedTimePeriodJson，无需后端新接口） */
const bindModalOpen = ref(false)
const bindGroupId = ref<string | null>(null)
const bindSaving = ref(false)
const bindTeacherIds = ref<string[]>([])

type StaffRow = { id: string, nickName: string, mobile: string }
const staffList = ref<StaffRow[]>([])
const bindStaffCache = new Map<string, StaffRow>()
const staffLoading = ref(false)
const bindStaffKeyword = ref('')
const bindPagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
})

const bindTeacherColumns: TableColumnType<StaffRow>[] = [
  { title: '姓名', dataIndex: 'nickName', key: 'nickName', ellipsis: true },
  { title: '手机号', dataIndex: 'mobile', key: 'mobile', width: 130, ellipsis: true },
]

const bindRowSelection = computed(() => ({
  selectedRowKeys: bindTeacherIds.value as unknown as (string | number)[],
  onChange: (keys: (string | number)[]) => {
    bindTeacherIds.value = keys.map(String)
  },
  preserveSelectedRowKeys: true,
}))

const bindModalTitle = computed(() => {
  const g = periodGroups.value.find(x => x.id === bindGroupId.value)
  const name = g?.name?.trim() || '该时段组'
  return `关联老师 — ${name}`
})

function cacheBindStaffRows(rows: StaffRow[]) {
  rows.forEach((row) => {
    bindStaffCache.set(row.id, row)
  })
}

async function loadStaffForBind() {
  staffLoading.value = true
  try {
    const res = await getUserListApi({
      pageRequestModel: {
        needTotal: true,
        pageSize: bindPagination.pageSize,
        pageIndex: bindPagination.current,
        skipCount: 0,
      },
      queryModel: {
        searchKey: bindStaffKeyword.value.trim() || undefined,
      },
    })
    if (res.code === 200) {
      const rows = Array.isArray(res.result) ? res.result : []
      const list = rows.map((r: { id?: unknown, nickName?: string, name?: string, mobile?: string }) => ({
        id: String(r.id ?? ''),
        nickName: String(r.nickName || r.name || r.id || '').trim() || String(r.id),
        mobile: String(r.mobile ?? '').trim(),
      })).filter((r: StaffRow) => r.id)
      staffList.value = list
      cacheBindStaffRows(list)
      bindPagination.total = Number(res.total || 0)
    }
    else {
      messageService.error(res.message || '加载老师列表失败')
    }
  }
  catch (e) {
    console.error('load staff for bind teachers', e)
    messageService.error('加载老师列表失败')
  }
  finally {
    staffLoading.value = false
  }
}

function openBindTeachers(record: UnifiedPeriodGroupLike) {
  bindStaffCache.clear()
  bindGroupId.value = String(record.id || '')
  bindTeacherIds.value = normalizedBoundTeachers(record).map(t => String(t.id))
  bindPagination.current = 1
  bindPagination.total = 0
  bindStaffKeyword.value = ''
  bindModalOpen.value = true
  void loadStaffForBind()
}

function closeBindModal() {
  debouncedSearchBindStaff.cancel()
  bindModalOpen.value = false
  bindGroupId.value = null
  bindTeacherIds.value = []
  bindStaffKeyword.value = ''
  bindPagination.current = 1
  bindPagination.total = 0
  staffList.value = []
}

function selectAllFilteredStaff() {
  const set = new Set(bindTeacherIds.value)
  staffList.value.forEach(s => set.add(s.id))
  bindTeacherIds.value = Array.from(set)
}

function clearFilteredStaffSelection() {
  const drop = new Set(staffList.value.map(s => s.id))
  bindTeacherIds.value = bindTeacherIds.value.filter(id => !drop.has(id))
}

function invertFilteredStaffSelection() {
  const visible = new Set(staffList.value.map(s => s.id))
  const cur = new Set(bindTeacherIds.value)
  const next = new Set(bindTeacherIds.value)
  visible.forEach((id) => {
    if (cur.has(id))
      next.delete(id)
    else
      next.add(id)
  })
  bindTeacherIds.value = Array.from(next)
}

function handleBindTableChange(paginationInfo: { current?: number, pageSize?: number }) {
  const nextPageSize = Number(paginationInfo.pageSize || bindPagination.pageSize)
  const pageSizeChanged = nextPageSize !== bindPagination.pageSize
  bindPagination.pageSize = nextPageSize
  bindPagination.current = pageSizeChanged ? 1 : Number(paginationInfo.current || 1)
  void loadStaffForBind()
}

const debouncedSearchBindStaff = debounce(() => {
  if (!bindModalOpen.value)
    return
  bindPagination.current = 1
  void loadStaffForBind()
}, 300)

watch(bindStaffKeyword, () => {
  if (!bindModalOpen.value)
    return
  debouncedSearchBindStaff()
})

async function saveBindTeachers() {
  if (!bindGroupId.value)
    return
  bindSaving.value = true
  try {
    const cfg = loadBaseConfig()
    const g = cfg.groups.find(x => x.id === bindGroupId.value)
    if (!g) {
      messageService.error('未找到该时段组，请刷新后重试')
      throw new Error('group missing')
    }
    const prev = g.boundTeachers || []
    const prevName = new Map(prev.map(t => [String(t.id), t.name]))
    g.boundTeachers = bindTeacherIds.value.map(id => ({
      id: String(id),
      name: bindStaffCache.get(String(id))?.nickName || prevName.get(String(id)) || String(id),
    }))
    const res = await setInstConfigApi({
      ...(userStore.instConfig as InstConfig),
      unifiedTimePeriodJson: cfg,
    })
    if (res.code !== 200) {
      messageService.error(res.message || '保存失败')
      throw new Error(res.message || 'save failed')
    }
    await userStore.getInstConfig()
    messageService.success('已保存关联老师')
    closeBindModal()
  }
  catch (e) {
    console.error(e)
    if (!(e instanceof Error && e.message === 'group missing') && !(e instanceof Error && e.message === 'save failed'))
      messageService.error('保存失败')
    throw e
  }
  finally {
    bindSaving.value = false
  }
}

function confirmDeleteGroup(item: UnifiedPeriodGroupLike) {
  if (periodGroups.value.length <= 1) {
    messageService.warning('至少保留一个时段组')
    return
  }
  if (hasBoundTeachers(item)) {
    messageService.warning('已关联老师的时段组不能删除，请先取消关联老师')
    return
  }
  Modal.confirm({
    title: '删除时段组',
    centered: true,
    content: `确定删除「${String(item.name || '')}」吗？排课中引用该组的节次可能受影响。`,
    okText: '删除',
    okType: 'danger',
    cancelText: '取消',
    async onOk() {
      try {
        const cfg = loadBaseConfig()
        cfg.groups = cfg.groups.filter(g => g.id !== String(item.id || ''))
        cfg.groups.forEach((g, i) => { g.sort = i })
        const res = await setInstConfigApi({
          ...(userStore.instConfig as InstConfig),
          unifiedTimePeriodJson: cfg,
        })
        if (res.code !== 200) {
          messageService.error(res.message || '删除失败')
          return
        }
        messageService.success('已删除')
        await refreshFromServer()
      }
      catch (e) {
        console.error(e)
        messageService.error('删除失败')
      }
    },
  })
}
</script>

<template>
  <div class="period-settings scrollbar">
    <div class="period-settings__panel">
      <div class="period-panel__head">
        <div class="period-panel__summary">
          <span class="period-panel__accent" aria-hidden="true" />
          <span class="period-panel__summary-text">当前共计 {{ periodGroups.length }} 个时段组（逐行编辑，可随时添加）</span>
        </div>
        <div class="period-panel__actions">
          <a-button :loading="repairing" @click="repairPeriodVersions">
            一键修复
          </a-button>
          <a-button type="primary" class="period-panel__edit" @click="openCreateGroup">
            <template #icon>
              <PlusOutlined />
            </template>
            添加时段组
          </a-button>
        </div>
      </div>

      <div class="period-panel__switch-row">
        <span class="period-panel__switch-label">快捷排课统一时段</span>
        <a-switch
          :checked="quickUnifiedEnabled"
          @change="onQuickUnifiedChange"
        />
      </div>

      <div class="period-panel__tip">
        {{ effectiveRuleText }}
      </div>

      <a-spin :spinning="loading">
        <a-table
          class="period-table"
          :columns="columns"
          :data-source="periodGroups"
          :pagination="false"
          row-key="id"
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'slots'">
              {{ slotsSummary(record) }}
            </template>
            <template v-else-if="column.key === 'span'">
              {{ groupTimeSpan(record) }}
            </template>
            <template v-else-if="column.key === 'teachers'">
              <span v-if="!(record.boundTeachers || []).length" class="period-teachers-cell period-teachers-cell--empty">
                —
              </span>
              <a-tooltip
                v-else
                placement="topLeft"
                :overlay-inner-style="{ maxWidth: 'min(420px, 90vw)' }"
                :title="formatBoundTeachersSummary(record)"
              >
                <span class="period-teachers-cell">{{ formatBoundTeachersSummary(record) }}</span>
              </a-tooltip>
            </template>
            <template v-else-if="column.key === 'status'">
              <span
                class="period-status"
                :class="activeSlotCount(record) > 0 ? 'period-status--on' : 'period-status--off'"
              >
                {{ activeSlotCount(record) > 0 ? '有可用节次' : '无启用节次' }}
              </span>
            </template>
            <template v-else-if="column.key === 'action'">
              <a-button type="link" size="small" class="period-action" @click="openEditGroup(record.id)">
                编辑
              </a-button>
              <a-button type="link" size="small" class="period-action" @click="openBindTeachers(record)">
                关联老师
              </a-button>
              <a-tooltip v-if="periodGroups.length > 1" :title="hasBoundTeachers(record) ? '已关联老师的时段组不能删除，请先取消关联老师' : null">
                <a-button
                  type="link"
                  size="small"
                  danger
                  class="period-action"
                  :disabled="hasBoundTeachers(record)"
                  @click="confirmDeleteGroup(record)"
                >
                  删除
                </a-button>
              </a-tooltip>
            </template>
          </template>
        </a-table>
      </a-spin>
    </div>

    <UnifiedPeriodGroupModal
      v-model:open="groupModalOpen"
      :mode="groupModalMode"
      :group-id="editingGroupId"
      @saved="onGroupModalSaved"
    />

    <a-modal
      :open="bindModalOpen"
      :title="bindModalTitle"
      :width="640"
      :mask-closable="false"
      destroy-on-close
      :confirm-loading="bindSaving"
      ok-text="保存"
      cancel-text="取消"
      class="bind-teachers-modal"
      @ok="saveBindTeachers"
      @cancel="closeBindModal"
    >
      <a-spin :spinning="staffLoading">
        <div class="bind-teachers-toolbar">
          <a-input-search
            v-model:value="bindStaffKeyword"
            allow-clear
            placeholder="搜索姓名或手机号"
            class="bind-teachers-toolbar__search"
          />
          <a-space :size="8" wrap>
            <a-button size="small" @click="selectAllFilteredStaff">
              全选列表
            </a-button>
            <a-button size="small" @click="clearFilteredStaffSelection">
              取消全选
            </a-button>
            <a-button size="small" @click="invertFilteredStaffSelection">
              反选
            </a-button>
          </a-space>
        </div>
        <a-table
          class="bind-teachers-table"
          size="small"
          :columns="bindTeacherColumns"
          :data-source="staffList"
          :row-selection="bindRowSelection"
          :pagination="{
            current: bindPagination.current,
            pageSize: bindPagination.pageSize,
            total: bindPagination.total,
            showSizeChanger: true,
            pageSizeOptions: ['10', '20', '50', '100'],
          }"
          row-key="id"
          :scroll="{ y: 320 }"
          @change="handleBindTableChange"
        />
      </a-spin>
    </a-modal>
  </div>
</template>

<style scoped lang="less">
.period-settings {
  position: relative;
  height: 100%;
  overflow-y: auto;
  background: #f2f4f7;
}

.period-settings__panel {
  margin: 12px 16px 20px;
  padding: 18px 20px 12px;
  border-radius: 12px;
  background: #fff;
  box-shadow: 0 1px 4px rgb(15 23 42 / 6%);
}

.period-panel__head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 14px;
}

.period-panel__summary {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  min-width: 0;
}

.period-panel__accent {
  display: inline-block;
  width: 4px;
  height: 16px;
  flex-shrink: 0;
  margin-top: 3px;
  border-radius: 2px;
  background: #1677ff;
}

.period-panel__summary-text {
  font-size: 14px;
  font-weight: 500;
  color: #1f2329;
  line-height: 1.5;
}

.period-panel__actions {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-shrink: 0;
}

.period-panel__edit {
  flex-shrink: 0;
  border-radius: 6px;
}

.period-panel__switch-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 12px;
  padding: 10px 12px;
  border-radius: 8px;
  background: #fafafa;
}

.period-panel__switch-label {
  font-size: 13px;
  color: #4b5563;
  font-weight: 500;
}

.period-panel__tip {
  margin-bottom: 12px;
  padding: 10px 12px;
  border-radius: 8px;
  background: #f6fbff;
  border: 1px solid #d9efff;
  color: #2f5f8f;
  font-size: 13px;
  line-height: 20px;
}

.period-table {
  :deep(.ant-table) {
    background: transparent;
  }

  :deep(.ant-table-thead > tr > th) {
    padding: 12px 16px;
    font-weight: 500;
    color: #262626;
    background: #fafafa !important;
    border-bottom: 1px solid #f0f0f0;
  }

  :deep(.ant-table-thead > tr > th::before) {
    display: none;
  }

  :deep(.ant-table-tbody > tr > td) {
    padding: 14px 16px;
    border-bottom: 1px solid #f5f5f5;
    background: #fff;
  }

  :deep(.ant-table-tbody > tr:last-child > td) {
    border-bottom: none;
  }
}

.period-status {
  font-size: 14px;
  font-weight: 500;
}

.period-status--on {
  color: #52c41a;
}

.period-status--off {
  color: #8c8c8c;
  font-weight: 400;
}

.period-action {
  padding: 0 4px !important;
  height: auto !important;
  color: #1677ff !important;
}

.period-action.ant-btn-disabled,
.period-action:disabled {
  color: #bfbfbf !important;
  cursor: not-allowed !important;
}

.period-action + .period-action {
  margin-left: 4px;
}

.period-teachers-cell {
  display: block;
  min-width: 0;
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  cursor: default;
}

.period-teachers-cell--empty {
  cursor: default;
}

.bind-teachers-toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 12px;
}

.bind-teachers-toolbar__search {
  width: 220px;
  max-width: 100%;
}

.bind-teachers-table {
  :deep(.ant-table-thead > tr > th) {
    padding: 8px 12px;
  }

  :deep(.ant-table-tbody > tr > td) {
    padding: 8px 12px;
  }
}

:deep(.bind-teachers-modal .ant-modal-body) {
  padding-top: 12px;
}
</style>
