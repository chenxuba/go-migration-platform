<script setup lang="ts">
/**
 * 单个时段组新增/编辑：写入 inst_config.unifiedTimePeriodJson
 */
import UnifiedPeriodGroupForm from '@/components/business-settings/unified-period-group-form.vue'
import { setInstConfigApi } from '@/api/common/config'
import { useUserStore } from '@/stores/user'
import {
  DEFAULT_UNIFIED_TIME_PERIOD_CONFIG,
  buildQuickHourlySlots,
  parseUnifiedTimePeriodConfig,
  type UnifiedPeriodGroup,
  type UnifiedTimePeriodConfig,
} from '@/utils/unified-time-period'
import messageService from '@/utils/messageService'

const props = defineProps<{
  open: boolean
  mode: 'create' | 'edit'
  /** 编辑必填 */
  groupId?: string | null
}>()

const emit = defineEmits<{
  (e: 'update:open', v: boolean): void
  (e: 'saved'): void
}>()

const userStore = useUserStore()
const saving = ref(false)
const formRef = ref<InstanceType<typeof UnifiedPeriodGroupForm> | null>(null)

const localGroup = ref<UnifiedPeriodGroup>(emptyNewGroup())

const modalTitle = computed(() =>
  props.mode === 'create' ? '添加时段组' : '编辑时段组',
)

function emptyNewGroup(): UnifiedPeriodGroup {
  return {
    id: `group-${Date.now()}`,
    name: '',
    sort: 0,
    slots: buildQuickHourlySlots().map(s => ({ ...s })),
  }
}

function cloneGroup(g: UnifiedPeriodGroup): UnifiedPeriodGroup {
  return {
    id: g.id,
    name: g.name,
    sort: g.sort,
    slots: g.slots.map(s => ({ ...s })),
  }
}

function cloneConfig(c: UnifiedTimePeriodConfig): UnifiedTimePeriodConfig {
  return {
    version: c.version,
    groups: c.groups.map(g => cloneGroup(g)),
  }
}

function loadBaseConfig(): UnifiedTimePeriodConfig {
  const raw = userStore.instConfig?.unifiedTimePeriodJson
  const parsed = parseUnifiedTimePeriodConfig(raw)
  return cloneConfig(parsed ?? DEFAULT_UNIFIED_TIME_PERIOD_CONFIG)
}

function validateFullConfig(cfg: UnifiedTimePeriodConfig): string | null {
  for (const g of cfg.groups) {
    if (!g.name.trim())
      return '存在未命名时段组'
    for (const s of g.slots) {
      if (!s.start || !s.end)
        return `「${g.name}」存在未填写的时间`
      if (s.start >= s.end)
        return `「${g.name}」第${s.index}节结束时间须晚于开始`
    }
  }
  return null
}

watch(
  () => props.open,
  (open) => {
    if (!open)
      return
    if (props.mode === 'create') {
      localGroup.value = emptyNewGroup()
      return
    }
    if (!props.groupId) {
      messageService.error('缺少时段组信息')
      close()
      return
    }
    const cfg = loadBaseConfig()
    const g = cfg.groups.find(x => x.id === props.groupId)
    if (!g) {
      messageService.error('未找到该时段组，请刷新后重试')
      close()
      return
    }
    localGroup.value = cloneGroup(g)
  },
)

function close() {
  emit('update:open', false)
}

function onModalOpenUpdate(v: boolean) {
  emit('update:open', v)
}

async function handleSave() {
  const err = formRef.value?.validateGroup()
  if (err) {
    messageService.error(err)
    return
  }

  const cfg = loadBaseConfig()

  if (props.mode === 'edit' && props.groupId) {
    const i = cfg.groups.findIndex(x => x.id === props.groupId)
    if (i < 0) {
      messageService.error('时段组不存在，请刷新后重试')
      return
    }
    cfg.groups[i] = cloneGroup(localGroup.value)
  }
  else {
    const g = cloneGroup(localGroup.value)
    g.id = `group-${Date.now()}`
    g.sort = cfg.groups.length
    cfg.groups.push(g)
  }

  const fullErr = validateFullConfig(cfg)
  if (fullErr) {
    messageService.error(fullErr)
    return
  }

  saving.value = true
  try {
    await setInstConfigApi({
      ...(userStore.instConfig as Record<string, unknown>),
      unifiedTimePeriodJson: cfg,
    } as never)
    await userStore.getInstConfig()
    messageService.success('保存成功')
    emit('saved')
    close()
  }
  catch (e) {
    console.error(e)
    messageService.error('保存失败')
  }
  finally {
    saving.value = false
  }
}
</script>

<template>
  <a-modal
    :open="open"
    class="unified-period-group-modal"
    :title="modalTitle"
    :width="560"
    :mask-closable="false"
    destroy-on-close
    :footer="null"
    @update:open="onModalOpenUpdate"
  >
    <div class="upgm-body">
      <UnifiedPeriodGroupForm
        ref="formRef"
        :group="localGroup"
        icon-variant="a"
      />
    </div>
    <div class="upgm-footer">
      <a-button @click="close">
        取消
      </a-button>
      <a-button type="primary" :loading="saving" @click="handleSave">
        保存
      </a-button>
    </div>
  </a-modal>
</template>

<style scoped lang="less">
/* 随内容增高，只有快超出视窗时才滚动，避免两三条节次也出现丑滚动条 */
.upgm-body {
  max-height: calc(100vh - 260px);
  overflow-y: auto;
  overscroll-behavior: contain;
  padding: 4px 0 8px;
  scrollbar-width: thin;
  scrollbar-color: rgb(15 23 42 / 22%) transparent;

  &::-webkit-scrollbar {
    width: 5px;
  }

  &::-webkit-scrollbar-track {
    background: transparent;
  }

  &::-webkit-scrollbar-thumb {
    background: rgb(15 23 42 / 22%);
    border-radius: 5px;
  }
}

.upgm-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  padding-top: 16px;
  margin-top: 8px;
  border-top: 1px solid #f0f0f0;
}

:deep(.unified-period-group-modal .ant-modal-body) {
  padding-top: 8px;
}
</style>
