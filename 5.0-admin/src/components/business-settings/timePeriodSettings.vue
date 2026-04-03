<script setup lang="ts">
import { PlusOutlined } from '@ant-design/icons-vue'
import { Modal } from 'ant-design-vue'
import type { TableColumnType } from 'ant-design-vue'
import UnifiedPeriodGroupModal from '@/components/business-settings/unified-period-group-modal.vue'
import { type InstConfig, setInstConfigApi } from '@/api/common/config'
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
const quickUnifiedEnabled = ref(false)

const groupModalOpen = ref(false)
const groupModalMode = ref<'create' | 'edit'>('edit')
const editingGroupId = ref<string | null>(null)

const periodGroups = computed<UnifiedPeriodGroup[]>(() => {
  const parsed = parseUnifiedTimePeriodConfig(userStore.instConfig?.unifiedTimePeriodJson)
  const cfg = parsed ?? DEFAULT_UNIFIED_TIME_PERIOD_CONFIG
  return configGroupsSorted(cfg)
})

const columns: TableColumnType<UnifiedPeriodGroup>[] = [
  { title: '时段名称', dataIndex: 'name', key: 'name', ellipsis: true },
  { title: '节次', key: 'slots', width: 150 },
  { title: '时间范围', key: 'span', width: 200 },
  { title: '状态', key: 'status', width: 120 },
  { title: '操作', key: 'action', width: 160 },
]

function sortSlots(slots: UnifiedPeriodSlot[]) {
  return [...slots].sort((a, b) => a.index - b.index)
}

function groupTimeSpan(g: UnifiedPeriodGroup): string {
  const active = sortSlots(g.slots).filter(s => s.enabled !== false)
  if (!active.length)
    return '—'
  return `${active[0].start} ~ ${active[active.length - 1].end}`
}

function slotsSummary(g: UnifiedPeriodGroup): string {
  const total = g.slots.length
  const active = slotCountActive(g)
  return `${active} / ${total} 节启用`
}

function cloneConfig(c: UnifiedTimePeriodConfig): UnifiedTimePeriodConfig {
  return {
    version: c.version,
    groups: c.groups.map(g => ({
      ...g,
      slots: g.slots.map(s => ({ ...s })),
    })),
  }
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

function confirmDeleteGroup(item: UnifiedPeriodGroup) {
  if (periodGroups.value.length <= 1) {
    messageService.warning('至少保留一个时段组')
    return
  }
  Modal.confirm({
    title: '删除时段组',
    centered: true,
    content: `确定删除「${item.name}」吗？排课中引用该组的节次可能受影响。`,
    okText: '删除',
    okType: 'danger',
    cancelText: '取消',
    async onOk() {
      try {
        const cfg = loadBaseConfig()
        cfg.groups = cfg.groups.filter(g => g.id !== item.id)
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
        <a-button type="primary" class="period-panel__edit" @click="openCreateGroup">
          <template #icon>
            <PlusOutlined />
          </template>
          添加时段组
        </a-button>
      </div>

      <div class="period-panel__switch-row">
        <span class="period-panel__switch-label">快捷排课统一时段</span>
        <a-switch
          :checked="quickUnifiedEnabled"
          @change="onQuickUnifiedChange"
        />
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
            <template v-else-if="column.key === 'status'">
              <span
                class="period-status"
                :class="slotCountActive(record) > 0 ? 'period-status--on' : 'period-status--off'"
              >
                {{ slotCountActive(record) > 0 ? '有可用节次' : '无启用节次' }}
              </span>
            </template>
            <template v-else-if="column.key === 'action'">
              <a-button type="link" size="small" class="period-action" @click="openEditGroup(record.id)">
                编辑
              </a-button>
              <a-button
                v-if="periodGroups.length > 1"
                type="link"
                size="small"
                danger
                class="period-action"
                @click="confirmDeleteGroup(record)"
              >
                删除
              </a-button>
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

.period-action + .period-action {
  margin-left: 4px;
}
</style>
