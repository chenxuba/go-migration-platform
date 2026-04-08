<script setup lang="ts">
/**
 * 单个时段组新增/编辑，保存机构统一时段配置
 */
import UnifiedPeriodGroupForm from '@/components/business-settings/unified-period-group-form.vue'
import { previewInstPeriodEffectiveApi, setInstConfigApi } from '@/api/common/config'
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
const previewLoading = ref(false)
const previewWeekStart = ref('')
const previewAppliedToday = ref<boolean | null>(null)
const formRef = ref<InstanceType<typeof UnifiedPeriodGroupForm> | null>(null)

const localGroup = ref<UnifiedPeriodGroup>(emptyNewGroup())

const modalTitle = computed(() =>
  props.mode === 'create' ? '添加时段组' : '编辑时段组',
)

const effectiveRuleText = '保存后：当前周之前保持不变；如果本周还没有老师排课，则新时段从本周生效，否则从下周生效。'

const effectivePreviewText = computed(() => {
  if (!previewWeekStart.value)
    return '正在计算本次修改会从哪一周开始生效...'
  return previewAppliedToday.value
    ? `本次预计从 ${previewWeekStart.value} 开始生效。`
    : `本次预计从 ${previewWeekStart.value} 开始生效；在此之前已排课的周不受影响。`
})

function emptyNewGroup(): UnifiedPeriodGroup {
  return {
    id: `group-${Date.now()}`,
    name: '',
    sort: 0,
    slots: buildQuickHourlySlots().map(s => ({ ...s })),
    boundTeachers: [],
  }
}

function cloneGroup(g: UnifiedPeriodGroup): UnifiedPeriodGroup {
  return {
    id: g.id,
    name: g.name,
    sort: g.sort,
    slots: g.slots.map(s => ({ ...s })),
    boundTeachers: (g.boundTeachers || []).map(t => ({ ...t })),
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

function buildEditingConfig(): UnifiedTimePeriodConfig | null {
  const cfg = loadBaseConfig()
  if (props.mode === 'edit' && props.groupId) {
    const i = cfg.groups.findIndex(x => x.id === props.groupId)
    if (i < 0)
      return null
    cfg.groups[i] = cloneGroup(localGroup.value)
    return cfg
  }
  const g = cloneGroup(localGroup.value)
  g.id = `group-${Date.now()}`
  g.sort = cfg.groups.length
  cfg.groups.push(g)
  return cfg
}

let previewTimer: ReturnType<typeof setTimeout> | null = null

async function refreshEffectivePreview() {
  const cfg = buildEditingConfig()
  if (!cfg) {
    previewWeekStart.value = ''
    previewAppliedToday.value = null
    return
  }
  previewLoading.value = true
  try {
    const res = await previewInstPeriodEffectiveApi({
      unifiedTimePeriodJson: cfg,
    })
    previewWeekStart.value = String(res.result?.periodWeekStart || '').trim()
    previewAppliedToday.value = typeof res.result?.periodAppliedToday === 'boolean' ? res.result.periodAppliedToday : null
  }
  catch (e) {
    console.error('preview inst period effective failed', e)
    previewWeekStart.value = ''
    previewAppliedToday.value = null
  }
  finally {
    previewLoading.value = false
  }
}

watch(
  () => props.open,
  (open) => {
    if (!open)
      return
    if (props.mode === 'create') {
      localGroup.value = emptyNewGroup()
      void refreshEffectivePreview()
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
    void refreshEffectivePreview()
  },
)

watch(
  () => props.open
    ? JSON.stringify({
        mode: props.mode,
        groupId: props.groupId,
        group: localGroup.value,
      })
    : '',
  () => {
    if (!props.open)
      return
    if (previewTimer)
      clearTimeout(previewTimer)
    previewTimer = setTimeout(() => {
      void refreshEffectivePreview()
    }, 250)
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

  const cfg = buildEditingConfig()
  if (!cfg) {
    messageService.error('时段组不存在，请刷新后重试')
    return
  }

  const fullErr = validateFullConfig(cfg)
  if (fullErr) {
    messageService.error(fullErr)
    return
  }

  saving.value = true
  try {
    const res = await setInstConfigApi({
      ...(userStore.instConfig as unknown as Record<string, unknown>),
      unifiedTimePeriodJson: cfg,
    } as never)
    await userStore.getInstConfig()
    const appliedWeek = res.result?.periodWeekStart
    if (appliedWeek) {
      messageService.success(res.result?.periodAppliedToday
        ? `保存成功，已从本周 ${appliedWeek} 生效`
        : `保存成功，已从 ${appliedWeek} 这一周开始生效，之前已排课周不受影响`)
    }
    else {
      messageService.success('保存成功')
    }
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
    :width="600"
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
        :show-bound-teachers="mode === 'create'"
      />
    </div>
    <div class="upgm-tip">
      {{ effectiveRuleText }}
    </div>
    <div class="upgm-tip upgm-tip--accent">
      {{ previewLoading ? '正在计算预计生效日期...' : effectivePreviewText }}
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

.upgm-tip {
  margin-top: 8px;
  padding: 10px 12px;
  border-radius: 10px;
  background: #f6fbff;
  border: 1px solid #d9efff;
  color: #2f5f8f;
  font-size: 13px;
  line-height: 20px;
}

.upgm-tip--accent {
  background: #fff9ef;
  border-color: #ffe2b8;
  color: #8a5a15;
}

:deep(.unified-period-group-modal .ant-modal-body) {
  padding-top: 8px;
}
</style>
