<script setup lang="ts">
/**
 * 课程设置等入口：统一时段「全量编辑」——按 Tab 分时段组，避免单页无限滚动
 */
import {
  DEFAULT_UNIFIED_TIME_PERIOD_CONFIG,
  buildQuickHourlySlots,
  type UnifiedPeriodGroup,
  type UnifiedTimePeriodConfig,
  parseUnifiedTimePeriodConfig,
} from '@/utils/unified-time-period'
import UnifiedPeriodGroupForm from '@/components/business-settings/unified-period-group-form.vue'
import { previewInstPeriodEffectiveApi, setInstConfigApi } from '@/api/common/config'
import { useUserStore } from '@/stores/user'
import messageService from '@/utils/messageService'

const props = defineProps<{
  open: boolean
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
const draft = ref<UnifiedTimePeriodConfig>(structuredClone(DEFAULT_UNIFIED_TIME_PERIOD_CONFIG))
const activeTabKey = ref('')
const effectiveRuleText = '保存后：当前周之前保持不变；如果本周还没有老师排课，则新时段从本周生效，否则从下周生效。'
const effectivePreviewText = computed(() => {
  if (!previewWeekStart.value)
    return '正在计算本次修改会从哪一周开始生效...'
  return previewAppliedToday.value
    ? `本次预计从 ${previewWeekStart.value} 开始生效。`
    : `本次预计从 ${previewWeekStart.value} 开始生效；在此之前已排课的周不受影响。`
})

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

function loadDraftFromStore() {
  const raw = userStore.instConfig?.unifiedTimePeriodJson
  const parsed = parseUnifiedTimePeriodConfig(raw)
  draft.value = cloneConfig(parsed ?? DEFAULT_UNIFIED_TIME_PERIOD_CONFIG)
  const sorted = sortGroups(draft.value.groups)
  activeTabKey.value = sorted[0]?.id ?? ''
}

async function refreshEffectivePreview() {
  previewLoading.value = true
  try {
    const res = await previewInstPeriodEffectiveApi({
      unifiedTimePeriodJson: draft.value,
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

let previewTimer: ReturnType<typeof setTimeout> | null = null

watch(
  () => props.open,
  (open) => {
    if (open) {
      loadDraftFromStore()
      void refreshEffectivePreview()
    }
  },
)

watch(
  () => props.open ? JSON.stringify(draft.value) : '',
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

function sortGroups(list: UnifiedPeriodGroup[]) {
  return [...list].sort((a, b) => a.sort - b.sort)
}

function hasBoundTeachers(group: UnifiedPeriodGroup): boolean {
  return Array.isArray(group.boundTeachers) && group.boundTeachers.length > 0
}

function addGroupTab() {
  const n = draft.value.groups.length
  const ch = String.fromCharCode(65 + (n % 26))
  const id = `group-${Date.now()}`
  draft.value.groups.push({
    id,
    name: `${ch}时段`,
    sort: n,
    slots: [
      { index: 1, start: '08:00', end: '09:00', enabled: true },
      { index: 2, start: '09:00', end: '10:00', enabled: true },
    ],
    boundTeachers: [],
  })
  activeTabKey.value = id
}

function removeGroupFromDraft(id: string) {
  if (draft.value.groups.length <= 1) {
    messageService.warning('至少保留一个时段组')
    return
  }
  const target = draft.value.groups.find(g => g.id === id)
  if (target && hasBoundTeachers(target)) {
    messageService.warning('已关联老师的时段组不能删除，请先取消关联老师')
    return
  }
  draft.value.groups = draft.value.groups.filter(g => g.id !== id)
  draft.value.groups.forEach((g, i) => { g.sort = i })
  const sorted = sortGroups(draft.value.groups)
  activeTabKey.value = sorted[0]?.id ?? ''
}

function quickGenerateAll() {
  for (const g of draft.value.groups)
    g.slots = buildQuickHourlySlots().map(s => ({ ...s }))
  messageService.success('已为所有时段组生成整点节次（8:00–19:00，共 12 节）')
}

function validateAll(): string | null {
  for (const g of draft.value.groups) {
    if (!g.name.trim())
      return '请填写每个时段组的名称'
    for (const s of g.slots) {
      if (!s.start || !s.end)
        return `「${g.name}」存在未填写的时间`
      if (s.start >= s.end)
        return `「${g.name}」第${s.index}节结束时间须晚于开始`
    }
  }
  return null
}

async function handleSave() {
  const err = validateAll()
  if (err) {
    messageService.error(err)
    return
  }
  saving.value = true
  try {
    const res = await setInstConfigApi({
      ...(userStore.instConfig as unknown as Record<string, unknown>),
      unifiedTimePeriodJson: draft.value,
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
    class="unified-period-full-modal"
    title="编辑统一时段"
    :width="640"
    :mask-closable="false"
    destroy-on-close
    :footer="null"
    @update:open="onModalOpenUpdate"
  >
    <div class="up-full__toolbar">
      <a-button size="small" @click="quickGenerateAll">
        全部为整点节次
      </a-button>
      <a-button type="primary" size="small" @click="addGroupTab">
        + 时段组
      </a-button>
    </div>

    <div class="up-full__body-scroll">
      <a-tabs v-model:active-key="activeTabKey" class="up-full__tabs">
        <a-tab-pane
          v-for="(g, gi) in sortGroups(draft.groups)"
          :key="g.id"
          :tab="g.name.trim() || `时段组 ${gi + 1}`"
        >
          <UnifiedPeriodGroupForm
            :group="g"
            :icon-variant="gi % 2 === 0 ? 'a' : 'b'"
            :allow-delete-group="draft.groups.length > 1"
            :delete-disabled-reason="hasBoundTeachers(g) ? '已关联老师的时段组不能删除，请先取消关联老师' : ''"
            @remove-group="removeGroupFromDraft(g.id)"
          />
        </a-tab-pane>
      </a-tabs>
    </div>

    <div class="up-full__tip">
      {{ effectiveRuleText }}
    </div>

    <div class="up-full__tip up-full__tip--accent">
      {{ previewLoading ? '正在计算预计生效日期...' : effectivePreviewText }}
    </div>

    <div class="up-full__footer">
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
.up-full__toolbar {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 12px;
}

/* 仅当节次很多、接近视窗高度时才出现滚动；少量内容不再套小固定高度避免多余滚动条 */
.up-full__body-scroll {
  max-height: calc(100vh - 220px);
  overflow-y: auto;
  overscroll-behavior: contain;
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

.up-full__tabs {
  :deep(.ant-tabs-content-holder) {
    padding-top: 8px;
  }
}

.up-full__footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  padding-top: 16px;
  margin-top: 8px;
  border-top: 1px solid #f0f0f0;
}

.up-full__tip {
  margin-top: 12px;
  padding: 10px 12px;
  border-radius: 10px;
  background: #f6fbff;
  border: 1px solid #d9efff;
  color: #2f5f8f;
  font-size: 13px;
  line-height: 20px;
}

.up-full__tip--accent {
  background: #fff9ef;
  border-color: #ffe2b8;
  color: #8a5a15;
}

:deep(.unified-period-full-modal .ant-modal-body) {
  padding-top: 12px;
}
</style>
