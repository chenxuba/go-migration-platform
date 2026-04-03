<script setup lang="ts">
/**
 * 单个时段组：名称 + 节次列表（由父组件传入同一引用以便双向修改）
 */
import { DeleteOutlined } from '@ant-design/icons-vue'
import {
  buildQuickHourlySlots,
  slotCountActive,
  type UnifiedPeriodGroup,
  type UnifiedPeriodSlot,
} from '@/utils/unified-time-period'

const props = withDefaults(
  defineProps<{
    group: UnifiedPeriodGroup
    iconVariant?: 'a' | 'b'
    allowDeleteGroup?: boolean
  }>(),
  {
    iconVariant: 'a',
    allowDeleteGroup: false,
  },
)

const emit = defineEmits<{
  (e: 'removeGroup'): void
}>()

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

function onEnabledChange(s: UnifiedPeriodSlot, v: boolean) {
  s.enabled = v
}

function quickGenerateThisGroup() {
  props.group.slots = buildQuickHourlySlots().map(s => ({ ...s }))
}

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

function validateGroup(): string | null {
  const g = props.group
  if (!g.name.trim())
    return '请填写时段名称'
  for (const s of g.slots) {
    if (!s.start || !s.end)
      return '存在未填写的时间'
    if (s.start >= s.end)
      return `第${s.index}节结束时间须晚于开始`
  }
  return null
}

defineExpose({ validateGroup })
</script>

<template>
  <section class="up-group-form">
    <div class="up-group-form__head">
      <span
        class="up-group-form__icon"
        :class="iconVariant === 'a' ? 'up-group-form__icon--a' : 'up-group-form__icon--b'"
      >
        {{ groupLetter(group) }}
      </span>
      <div class="up-group-form__head-text">
        <span class="up-group-form__name">{{ group.name || '未命名时段' }}</span>
        <span class="up-group-form__meta">共 {{ slotCountActive(group) }} 节启用</span>
      </div>
      <button
        v-if="allowDeleteGroup"
        type="button"
        class="up-group-form__trash"
        @click="emit('removeGroup')"
      >
        <DeleteOutlined />
      </button>
    </div>

    <div class="up-group-form__toolbar">
      <a-button type="primary" size="small" @click="quickGenerateThisGroup">
        本组填充满整点（8:00–19:00）
      </a-button>
    </div>

    <div class="up-group-form__field">
      <span class="up-group-form__label">时段名称</span>
      <a-input v-model:value="group.name" allow-clear placeholder="如 A时段" />
    </div>

    <div class="up-group-form__slots">
      <div
        v-for="s in sortSlots(group.slots)"
        :key="`${group.id}-${s.index}`"
        class="up-group-form__slot"
      >
        <div class="up-group-form__slot-main">
          <span class="up-group-form__slot-num">{{ s.index }}</span>
          <div class="up-group-form__slot-times">
            <a-time-picker
              v-model:value="s.start"
              value-format="HH:mm"
              format="HH:mm"
              placeholder="开始"
              :minute-step="5"
              :input-read-only="true"
              class="up-group-form__picker"
              @change="() => clampSlotEndAfterStart(s)"
            />
            <span class="up-group-form__dash">—</span>
            <a-time-picker
              v-model:value="s.end"
              value-format="HH:mm"
              format="HH:mm"
              placeholder="结束"
              :minute-step="5"
              :input-read-only="true"
              :disabled="!s.start"
              :disabled-time="() => disabledEndTimeByStart(s.start)"
              class="up-group-form__picker"
            />
          </div>
        </div>
        <div class="up-group-form__slot-row2">
          <a-switch
            :checked="s.enabled !== false"
            checked-children="开"
            un-checked-children="停"
            @update:checked="(v) => onEnabledChange(s, !!v)"
          />
          <button type="button" class="up-group-form__del" @click="removeSlot(group, s.index)">
            删除
          </button>
        </div>
      </div>
    </div>

    <button type="button" class="up-group-form__add-line" @click="addSlot(group)">
      + 添加节次
    </button>
  </section>
</template>

<style scoped lang="less">
.up-group-form {
  padding: 14px;
  border-radius: 14px;
  background: #fff;
  border: 1px solid #f0f0f0;
}

.up-group-form__head {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 12px;
}

.up-group-form__icon {
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

.up-group-form__head-text {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.up-group-form__name {
  font-size: 16px;
  font-weight: 600;
  color: #262626;
}

.up-group-form__meta {
  font-size: 12px;
  color: #8c8c8c;
}

.up-group-form__trash {
  border: none;
  background: none;
  color: #8c8c8c;
  padding: 8px;
  cursor: pointer;

  &:hover {
    color: #ff4d4f;
  }
}

.up-group-form__toolbar {
  margin-bottom: 12px;
}

.up-group-form__field {
  margin-bottom: 12px;
}

.up-group-form__label {
  display: block;
  margin-bottom: 8px;
  color: #4b5563;
  font-size: 13px;
  font-weight: 600;
}

.up-group-form__field :deep(.ant-input) {
  border-radius: 10px;
}

.up-group-form__slots {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.up-group-form__slot {
  padding: 12px;
  border-radius: 12px;
  background: #f8fafc;
  border: 1px solid #eef2f6;
}

.up-group-form__slot-main {
  display: flex;
  align-items: flex-start;
  gap: 10px;
}

.up-group-form__slot-num {
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

.up-group-form__slot-times {
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

.up-group-form__dash {
  display: none;
  text-align: center;
  color: #bfbfbf;
}

@media (min-width: 480px) {
  .up-group-form__slot-times {
    flex-direction: row;
    align-items: center;

    :deep(.ant-picker) {
      flex: 1;
      min-width: 0;
      width: auto;
    }
  }

  .up-group-form__dash {
    display: block;
    flex: 0 0 20px;
  }
}

.up-group-form__slot-row2 {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 10px;
  padding-top: 10px;
  border-top: 1px solid #f0f0f0;
}

.up-group-form__del {
  border: none;
  background: none;
  color: #ff4d4f;
  font-size: 14px;
  cursor: pointer;
  padding: 4px 8px;
}

.up-group-form__add-line {
  width: 100%;
  margin-top: 12px;
  padding: 10px;
  border: 1px dashed #91d5ff;
  border-radius: 12px;
  background: #fafcff;
  color: #1890ff;
  font-size: 14px;
  cursor: pointer;
}
</style>
