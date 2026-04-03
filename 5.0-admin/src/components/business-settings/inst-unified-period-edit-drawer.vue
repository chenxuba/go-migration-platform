<script setup lang="ts">
/**
 * 与教室设置一致：底部弹出抽屉（placement=bottom、无遮罩、圆角顶、底部操作栏）
 */
import { CloseOutlined, DeleteOutlined } from '@ant-design/icons-vue'
import {
  DEFAULT_UNIFIED_TIME_PERIOD_CONFIG,
  buildQuickHourlySlots,
  type UnifiedPeriodGroup,
  type UnifiedPeriodSlot,
  type UnifiedTimePeriodConfig,
  parseUnifiedTimePeriodConfig,
  slotCountActive,
} from '@/utils/unified-time-period'
import { setInstConfigApi } from '@/api/common/config'
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
const draft = ref<UnifiedTimePeriodConfig>(structuredClone(DEFAULT_UNIFIED_TIME_PERIOD_CONFIG))

function cloneConfig(c: UnifiedTimePeriodConfig): UnifiedTimePeriodConfig {
  return {
    version: c.version,
    groups: c.groups.map(g => ({
      ...g,
      slots: g.slots.map(s => ({ ...s })),
    })),
  }
}

function loadDraftFromStore() {
  const raw = userStore.instConfig?.unifiedTimePeriodJson
  const parsed = parseUnifiedTimePeriodConfig(raw)
  draft.value = cloneConfig(parsed ?? DEFAULT_UNIFIED_TIME_PERIOD_CONFIG)
}

watch(
  () => props.open,
  (open) => {
    if (open)
      loadDraftFromStore()
  },
)

function close() {
  emit('update:open', false)
}

function sortGroups(list: UnifiedPeriodGroup[]) {
  return [...list].sort((a, b) => a.sort - b.sort)
}

function sortSlots(slots: UnifiedPeriodSlot[]) {
  return [...slots].sort((a, b) => a.index - b.index)
}

function groupLetter(g: UnifiedPeriodGroup) {
  const c = (g.name || '').trim().charAt(0)
  return c || '—'
}

function addSlot(group: UnifiedPeriodGroup) {
  const maxIdx = group.slots.reduce((m, s) => Math.max(m, s.index), 0)
  group.slots.push({ index: maxIdx + 1, start: '08:00', end: '09:00', enabled: true })
}

function removeSlot(g: UnifiedPeriodGroup, idx: number) {
  g.slots = g.slots.filter(s => s.index !== idx)
  let n = 1
  for (const s of sortSlots(g.slots)) {
    s.index = n
    n++
  }
}

function addGroup() {
  const n = draft.value.groups.length
  const ch = String.fromCharCode(65 + (n % 26))
  draft.value.groups.push({
    id: `group-${Date.now()}`,
    name: `${ch}时段`,
    sort: n,
    slots: [
      { index: 1, start: '08:00', end: '09:00', enabled: true },
      { index: 2, start: '09:00', end: '10:00', enabled: true },
    ],
  })
}

function removeGroup(id: string) {
  if (draft.value.groups.length <= 1) {
    messageService.warning('至少保留一个时段组')
    return
  }
  draft.value.groups = draft.value.groups.filter(g => g.id !== id)
  draft.value.groups.forEach((g, i) => { g.sort = i })
}

function onEnabledChange(s: UnifiedPeriodSlot, v: boolean) {
  s.enabled = v
}

/** 保留已有时段组与名称，每组节次重置为 8:00–19:00 共 12 节整点 */
function quickGenerate() {
  for (const g of draft.value.groups)
    g.slots = buildQuickHourlySlots().map(s => ({ ...s }))
  messageService.success('已为各时段组生成整点节次（8:00–19:00，共 12 节）')
}

/** 解析「HH:mm」 */
function parseHHmm(str: string): { h: number, m: number } | null {
  const parts = String(str || '').trim().split(':')
  if (parts.length < 2)
    return null
  const h = Number(parts[0])
  const m = Number(parts[1])
  if (!Number.isFinite(h) || !Number.isFinite(m))
    return null
  if (h < 0 || h > 23 || m < 0 || m > 59)
    return null
  return { h, m }
}

/** 结束时间：不得早于或等于开始（与 one-to-one-schedule-modal disabledCustomEndTime 一致） */
function disabledEndTimeByStart(startStr: string) {
  const hm = parseHHmm(startStr)
  if (!hm) {
    return {
      disabledHours: () => [] as number[],
      disabledMinutes: () => [] as number[],
      disabledSeconds: () => [] as number[],
    }
  }
  const { h: startHour, m: startMinute } = hm
  return {
    disabledHours: () => Array.from({ length: startHour }, (_, i) => i),
    disabledMinutes: (selectedHour: number) => {
      if (selectedHour === startHour)
        return Array.from({ length: startMinute + 1 }, (_, i) => i)
      return []
    },
    disabledSeconds: () => [] as number[],
  }
}

/** 开始改晚后，若结束 ≤ 开始则自动推到开始后一格（5 分钟） */
function clampSlotEndAfterStart(s: UnifiedPeriodSlot) {
  const sh = parseHHmm(s.start || '')
  const eh = parseHHmm(s.end || '')
  if (!sh || !eh)
    return
  const sm = sh.h * 60 + sh.m
  const em = eh.h * 60 + eh.m
  if (em <= sm) {
    const next = sm + 5
    if (next >= 24 * 60) {
      s.end = '23:59'
      return
    }
    const nh = Math.floor(next / 60)
    const nm = next % 60
    s.end = `${String(nh).padStart(2, '0')}:${String(nm).padStart(2, '0')}`
  }
}

async function handleSave() {
  for (const g of draft.value.groups) {
    if (!g.name.trim()) {
      messageService.error('请填写每个时段组的名称')
      return
    }
    for (const s of g.slots) {
      if (!s.start || !s.end) {
        messageService.error(`「${g.name}」存在未填写的时间`)
        return
      }
      if (s.start >= s.end) {
        messageService.error(`「${g.name}」第${s.index}节结束时间须晚于开始`)
        return
      }
    }
  }
  saving.value = true
  try {
    await setInstConfigApi({
      ...(userStore.instConfig as Record<string, unknown>),
      unifiedTimePeriodJson: draft.value,
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
  <a-drawer
    root-class-name="unified-period-setting-drawer"
    :closable="false"
    :mask="false"
    :mask-closable="false"
    placement="bottom"
    :open="open"
    :get-container="false"
    :height="'72vh'"
    :body-style="{ padding: 0 }"
    @close="close"
  >
    <div class="unified-period-drawer">
      <div class="unified-period-drawer__header">
        <span class="unified-period-drawer__title">编辑统一时段</span>
        <span class="unified-period-drawer__close" @click="close">
          <CloseOutlined />
        </span>
      </div>

      <div class="unified-period-drawer__body">
        <div class="unified-period-drawer__quick">
          <a-button type="primary" @click="quickGenerate">
            快捷生成
          </a-button>
        </div>

        <section
          v-for="(g, gi) in sortGroups(draft.groups)"
          :key="g.id"
          class="unified-period-block"
        >
          <div class="unified-period-block__head">
            <span
              class="unified-period-block__icon"
              :class="gi % 2 === 0 ? 'unified-period-block__icon--a' : 'unified-period-block__icon--b'"
            >
              {{ groupLetter(g) }}
            </span>
            <div class="unified-period-block__head-text">
              <span class="unified-period-block__name">{{ g.name }}</span>
              <span class="unified-period-block__meta">共 {{ slotCountActive(g) }} 节</span>
            </div>
            <button
              v-if="draft.groups.length > 1"
              type="button"
              class="unified-period-block__trash"
              @click="removeGroup(g.id)"
            >
              <DeleteOutlined />
            </button>
          </div>

          <div class="unified-period-field">
            <span class="unified-period-field__label">时段名称</span>
            <a-input v-model:value="g.name" allow-clear placeholder="如 A时段" />
          </div>

          <div class="unified-period-slots">
            <div
              v-for="s in sortSlots(g.slots)"
              :key="`${g.id}-${s.index}`"
              class="unified-period-slot"
            >
              <div class="unified-period-slot__main">
                <span class="unified-period-slot__num">{{ s.index }}</span>
                <div class="unified-period-slot__times">
                  <a-time-picker
                    v-model:value="s.start"
                    value-format="HH:mm"
                    format="HH:mm"
                    placeholder="开始"
                    :minute-step="5"
                    :input-read-only="true"
                    class="unified-period-slot__picker"
                    @change="() => clampSlotEndAfterStart(s)"
                  />
                  <span class="unified-period-slot__dash">—</span>
                  <a-time-picker
                    v-model:value="s.end"
                    value-format="HH:mm"
                    format="HH:mm"
                    placeholder="结束"
                    :minute-step="5"
                    :input-read-only="true"
                    :disabled="!s.start"
                    :disabled-time="() => disabledEndTimeByStart(s.start)"
                    class="unified-period-slot__picker"
                  />
                </div>
              </div>
              <div class="unified-period-slot__row2">
                <a-switch
                  :checked="s.enabled !== false"
                  checked-children="开"
                  un-checked-children="停"
                  @update:checked="(v) => onEnabledChange(s, !!v)"
                />
                <button type="button" class="unified-period-slot__del" @click="removeSlot(g, s.index)">
                  删除
                </button>
              </div>
            </div>
          </div>

          <button type="button" class="unified-period-add-line" @click="addSlot(g)">
            + 添加节次
          </button>
        </section>

        <button type="button" class="unified-period-add-group" @click="addGroup">
          + 添加时段组
        </button>
      </div>

      <div class="unified-period-drawer__footer">
        <a-button @click="close">
          取消
        </a-button>
        <a-button type="primary" :loading="saving" @click="handleSave">
          保存
        </a-button>
      </div>
    </div>
  </a-drawer>
</template>

<style scoped lang="less">
/* 与 classroomSettings.vue 中 .classroom-drawer 一致：外层灰底，边距在 __body 内收 */
.unified-period-drawer {
  display: flex;
  height: 100%;
  flex-direction: column;
  background: #f6f7f8;
}

.unified-period-drawer__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 16px;
  background: #fff;
}

.unified-period-drawer__title {
  color: #1f2329;
  font-size: 18px;
  font-weight: 700;
}

.unified-period-drawer__close {
  cursor: pointer;
  color: #8c8c8c;
  font-size: 18px;
}

.unified-period-drawer__body {
  flex: 1;
  overflow-y: auto;
  padding: 12px 16px 24px;
}

.unified-period-drawer__quick {
  display: flex;
  margin-bottom: 12px;
}

.unified-period-drawer__footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  padding: 14px 16px calc(env(safe-area-inset-bottom) + 14px);
  border-top: 1px solid #eef2f6;
  background: #fff;
}

.unified-period-block {
  margin-bottom: 12px;
  padding: 14px;
  border-radius: 18px;
  background: #fff;
}

.unified-period-block__head {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 14px;
}

.unified-period-block__icon {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-size: 14px;
  font-weight: 700;
  flex-shrink: 0;

  &--a {
    background: linear-gradient(135deg, #1890ff 0%, #40a9ff 100%);
  }

  &--b {
    background: linear-gradient(135deg, #52c41a 0%, #73d13d 100%);
  }
}

.unified-period-block__head-text {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.unified-period-block__name {
  font-size: 16px;
  font-weight: 600;
  color: #262626;
}

.unified-period-block__meta {
  font-size: 12px;
  color: #8c8c8c;
}

.unified-period-block__trash {
  border: none;
  background: none;
  color: #8c8c8c;
  padding: 8px;
  cursor: pointer;

  &:hover {
    color: #ff4d4f;
  }
}

.unified-period-field {
  margin: 0 0 12px;
}

.unified-period-field__label {
  display: block;
  margin-bottom: 10px;
  color: #4b5563;
  font-size: 13px;
  font-weight: 600;
}

.unified-period-field :deep(.ant-input),
.unified-period-field :deep(.ant-input-affix-wrapper) {
  border-radius: 14px;
}

.unified-period-slots {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.unified-period-slot {
  padding: 12px;
  border-radius: 14px;
  background: #f8fafc;
  border: 1px solid #f0f0f0;
}

.unified-period-slot__main {
  display: flex;
  align-items: flex-start;
  gap: 10px;
}

.unified-period-slot__num {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background: #e6f7ff;
  color: #1890ff;
  font-size: 13px;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  margin-top: 5px;
}

.unified-period-slot__times {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 8px;
  min-width: 0;

  :deep(.ant-picker) {
    width: 100%;
    border-radius: 10px;
  }
}

.unified-period-slot__dash {
  display: none;
  text-align: center;
  color: #bfbfbf;
}

@media (min-width: 400px) {
  .unified-period-slot__times {
    flex-direction: row;
    align-items: center;

    :deep(.ant-picker) {
      flex: 1;
      min-width: 0;
      width: auto;
    }
  }

  .unified-period-slot__dash {
    display: block;
    flex: 0 0 20px;
  }
}

.unified-period-slot__row2 {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 10px;
  padding-top: 10px;
  border-top: 1px solid #f0f0f0;
}

.unified-period-slot__del {
  border: none;
  background: none;
  color: #ff4d4f;
  font-size: 14px;
  cursor: pointer;
  padding: 4px 8px;
}

.unified-period-add-line {
  width: 100%;
  margin-top: 12px;
  padding: 10px;
  border: 1px dashed #91d5ff;
  border-radius: 14px;
  background: #fff;
  color: #1890ff;
  font-size: 14px;
  cursor: pointer;
}

.unified-period-add-group {
  width: 100%;
  padding: 12px;
  border: 1px dashed #d9d9d9;
  border-radius: 14px;
  background: #fff;
  color: #595959;
  font-size: 14px;
  cursor: pointer;
}

/* 高度由 a-drawer :height 写入内联样式（默认 378px 会盖过仅 CSS 的修改） */
:deep(.unified-period-setting-drawer .ant-drawer-content) {
  border-radius: 20px 20px 0 0;
  overflow: hidden;
}
</style>
