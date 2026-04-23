<script setup lang="ts">
/**
 * 单个时段组：名称 + 节次列表（由父组件传入同一引用以便双向修改）
 */
import { DeleteOutlined } from '@ant-design/icons-vue'
import { getUserListApi } from '@/api/internal-manage/staff-manage'
import {
  generateSlotsSmartFill,
  slotCountActive,
  type UnifiedPeriodGroup,
  type UnifiedPeriodSlot,
  validateUnifiedPeriodGroup,
} from '@/utils/unified-time-period'
import messageService from '@/utils/messageService'

const props = withDefaults(
  defineProps<{
    group: UnifiedPeriodGroup
    iconVariant?: 'a' | 'b'
    allowDeleteGroup?: boolean
    deleteDisabledReason?: string
    /** 为 false 时不展示「关联老师」（如单组「编辑时段组」弹窗） */
    showBoundTeachers?: boolean
  }>(),
  {
    iconVariant: 'a',
    allowDeleteGroup: false,
    deleteDisabledReason: '',
    showBoundTeachers: true,
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
  return c || '时'
}

function groupTimeRange(g: UnifiedPeriodGroup) {
  const active = sortSlots(g.slots).filter(slot => slot.enabled !== false && slot.start && slot.end)
  if (!active.length)
    return '暂未设置时间范围'
  return `${active[0].start} - ${active[active.length - 1].end}`
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

const smartModalOpen = ref(false)
const smartFirstStart = ref<string>('08:00')
const smartLessonMins = ref<number>(40)
const smartBreakMins = ref<number>(10)
const smartLunchMins = ref<number>(60)
const smartMaxSlots = ref<number>(11)

function openSmartFillModal() {
  smartModalOpen.value = true
}

function applyPresetHourly() {
  smartFirstStart.value = '08:00'
  smartLessonMins.value = 40
  smartBreakMins.value = 10
  smartLunchMins.value = 60
  smartMaxSlots.value = 11
  messageService.info('已填入康复机构模板：40分钟课时、10分钟课间、60分钟午休，点「生成」替换节次')
}

function applySmartFill() {
  const slots = generateSlotsSmartFill({
    firstStart: smartFirstStart.value || '08:00',
    lessonMinutes: smartLessonMins.value,
    breakBetweenMinutes: smartBreakMins.value,
    lunchBreakMinutes: smartLunchMins.value,
    lunchStart: '12:30',
    maxSlots: smartMaxSlots.value,
  })
  if (!slots.length) {
    messageService.error('请检查最早上课时间是否有效')
    return
  }
  props.group.slots = slots
  smartModalOpen.value = false
  messageService.success(`已生成 ${slots.length} 节课`)
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
  return validateUnifiedPeriodGroup(props.group)
}

defineExpose({ validateGroup })

type StaffOptionRow = { id: string, nickName: string }

const staffList = ref<StaffOptionRow[]>([])
const staffLoading = ref(false)

const teacherSelectOptions = computed(() =>
  staffList.value.map(s => ({ value: s.id, label: s.nickName })),
)

const teacherIdsModel = computed({
  get: () => (props.group.boundTeachers || []).map(t => String(t.id)),
  set: (ids: string[]) => {
    const safeIds = (ids || []).map(String).filter(Boolean)
    const byStaff = new Map(staffList.value.map(s => [s.id, s.nickName]))
    const prev = props.group.boundTeachers || []
    const prevName = new Map(prev.map(t => [String(t.id), t.name]))
    props.group.boundTeachers = safeIds.map((id) => {
      const name = byStaff.get(id) || prevName.get(id) || id
      return { id, name }
    })
  },
})

async function ensureStaffOptionsLoaded() {
  if (staffList.value.length)
    return
  staffLoading.value = true
  try {
    const res = await getUserListApi({
      pageRequestModel: {
        needTotal: false,
        pageSize: 500,
        pageIndex: 1,
        skipCount: 1,
      },
      queryModel: {
        isTeacher: true,
      },
    })
    if (res.code === 200) {
      const rows = Array.isArray(res.result) ? res.result : []
      staffList.value = rows.map((r: { id?: unknown, nickName?: string, name?: string }) => ({
        id: String(r.id ?? ''),
        nickName: String(r.nickName || r.name || r.id || '').trim() || String(r.id),
      })).filter((r: StaffOptionRow) => r.id)
    }
  }
  catch (e) {
    console.error('load staff for period group', e)
    messageService.error('加载老师列表失败')
  }
  finally {
    staffLoading.value = false
  }
}

function onTeacherDropdownOpen(open: boolean) {
  if (open)
    void ensureStaffOptionsLoaded()
}

function filterTeacherOption(input: string, option: { label?: string }) {
  const q = (input || '').trim().toLowerCase()
  if (!q)
    return true
  return String(option?.label || '').toLowerCase().includes(q)
}

const deleteGroupDisabled = computed(() => Boolean(props.deleteDisabledReason))
</script>

<template>
  <section class="up-group-form">
    <div class="up-group-form__setup">
      <div class="up-group-form__setup-summary">
        <div class="up-group-form__head">
          <span
            class="up-group-form__icon"
            :class="iconVariant === 'a' ? 'up-group-form__icon--a' : 'up-group-form__icon--b'"
          >
            {{ groupLetter(group) }}
          </span>
          <div class="up-group-form__head-text">
            <span class="up-group-form__name">{{ group.name || '未命名时段' }}</span>
            <span class="up-group-form__meta">时段范围 {{ groupTimeRange(group) }}</span>
          </div>
        </div>

        <div class="up-group-form__summary-bottom">
          <span class="up-group-form__hero-tag">{{ slotCountActive(group) }} 节启用</span>
          <span class="up-group-form__hero-tag up-group-form__hero-tag--muted">{{ group.slots.length }} 节总计</span>
          <a-button
            type="primary"
            ghost
            size="small"
            class="up-group-form__smart-btn"
            title="按上课时间规则一键生成整组节次，适合快速初始化。"
            @click="openSmartFillModal"
          >
            智能生成节次
          </a-button>
        </div>

        <a-tooltip v-if="allowDeleteGroup" :title="deleteGroupDisabled ? props.deleteDisabledReason : null">
          <button
            type="button"
            class="up-group-form__trash"
            :disabled="deleteGroupDisabled"
            @click="emit('removeGroup')"
          >
            <DeleteOutlined />
          </button>
        </a-tooltip>
      </div>

      <div class="up-group-form__setup-fields">
        <div class="up-group-form__setup-title">
          <div class="up-group-form__section-title">
            基础信息
          </div>
          <p v-if="showBoundTeachers" class="up-group-form__setup-desc">
            可选多名；同一老师可绑定多个时段组。
          </p>
        </div>
        <div class="up-group-form__basic-grid" :class="{ 'up-group-form__basic-grid--single': !showBoundTeachers }">
          <div class="up-group-form__field-card">
            <span class="up-group-form__label">时段名称</span>
            <a-input v-model:value="group.name" allow-clear placeholder="如 A时段" />
          </div>

          <div v-if="showBoundTeachers" class="up-group-form__field-card">
            <span class="up-group-form__label">关联老师</span>
            <a-select
              v-model:value="teacherIdsModel"
              mode="multiple"
              allow-clear
              show-search
              :max-tag-count="2"
              :options="teacherSelectOptions"
              :filter-option="filterTeacherOption"
              :loading="staffLoading"
              placeholder="打开下拉可加载机构老师，支持搜索"
              class="up-group-form__teacher-select"
              @dropdown-visible-change="onTeacherDropdownOpen"
            />
          </div>
        </div>
      </div>
    </div>

    <a-modal
      v-model:open="smartModalOpen"
      title="按规则生成节次"
      :width="620"
      :mask-closable="false"
      destroy-on-close
      class="up-group-smart-modal"
      :footer="null"
    >
      <p class="up-group-smart-modal__hint">
        一节结束后，先经过「课间休息」分钟，再开始下一节；午休从当天 <strong>12:30</strong> 起连续休息对应时长（填 0 表示不插午休）。
      </p>
      <a-form layout="vertical" class="up-group-smart-modal__form">
        <a-form-item label="最早上课时间">
          <a-time-picker
            v-model:value="smartFirstStart"
            value-format="HH:mm"
            format="HH:mm"
            :minute-step="5"
            style="width: 100%"
          />
        </a-form-item>
        <a-form-item label="每节课时长（分钟）">
          <a-input-number v-model:value="smartLessonMins" :min="5" :max="180" style="width: 100%" />
        </a-form-item>
        <a-form-item
          label="课间休息（分钟）"
          extra="上一节课下课至下一节上课之间的间隔"
        >
          <a-input-number v-model:value="smartBreakMins" :min="0" :max="120" style="width: 100%" />
        </a-form-item>
        <a-form-item
          label="午休时长（分钟）"
          extra="自 12:30 起，填 0 表示不设午休空档"
        >
          <a-input-number v-model:value="smartLunchMins" :min="0" :max="240" style="width: 100%" />
        </a-form-item>
        <a-form-item label="最多生成几节课">
          <a-input-number v-model:value="smartMaxSlots" :min="1" :max="32" style="width: 100%" />
        </a-form-item>
      </a-form>
      <div class="up-group-smart-modal__preset">
        <a-button type="link" size="small" class="up-group-smart-modal__preset-btn" @click="applyPresetHourly">
          填入康复机构模板（8:00 开始，40 分钟课时 + 10 分钟课间 + 60 分钟午休，共 11 节）
        </a-button>
      </div>
      <div class="up-group-smart-modal__footer-btns">
        <a-button @click="smartModalOpen = false">
          取消
        </a-button>
        <a-button type="primary" @click="applySmartFill">
          生成并替换
        </a-button>
      </div>
    </a-modal>

    <div class="up-group-form__section">
      <div class="up-group-form__section-head up-group-form__section-head--slots">
        <div>
          <div class="up-group-form__section-title">
            节次设置
          </div>
          <p class="up-group-form__section-desc">
            每个节次单独调整开始、结束时间和启用状态，桌面端可同时浏览多节。
          </p>
        </div>
        <div class="up-group-form__section-actions">
          <span class="up-group-form__section-chip">{{ slotCountActive(group) }}/{{ group.slots.length }} 节启用</span>
          <a-button type="primary" ghost @click="addSlot(group)">
            + 添加节次
          </a-button>
        </div>
      </div>

      <div class="up-group-form__slots-grid">
        <div
          v-for="s in sortSlots(group.slots)"
          :key="`${group.id}-${s.index}`"
          class="up-group-form__slot"
        >
          <div class="up-group-form__slot-top">
            <div class="up-group-form__slot-head">
              <span class="up-group-form__slot-num">{{ s.index }}</span>
              <div class="up-group-form__slot-head-text">
                <span class="up-group-form__slot-title">第 {{ s.index }} 节</span>
                <span class="up-group-form__slot-subtitle">{{ s.enabled !== false ? '当前启用' : '当前停用' }}</span>
              </div>
            </div>
            <a-switch
              :checked="s.enabled !== false"
              checked-children="开"
              un-checked-children="停"
              @update:checked="(v) => onEnabledChange(s, !!v)"
            />
          </div>

          <div class="up-group-form__slot-times">
            <div class="up-group-form__time-field">
              <span class="up-group-form__time-label">开始时间</span>
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
            </div>
            <div class="up-group-form__time-field">
              <span class="up-group-form__time-label">结束时间</span>
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

          <div class="up-group-form__slot-footer">
            <span class="up-group-form__slot-state" :class="{ 'up-group-form__slot-state--off': s.enabled === false }">
              {{ s.enabled !== false ? '已启用，可参与排课' : '已停用，不参与排课' }}
            </span>
            <button type="button" class="up-group-form__del" @click="removeSlot(group, s.index)">
              删除
            </button>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>

<style scoped lang="less">
.up-group-form {
  padding: 6px 2px 2px;
}

.up-group-form__setup,
.up-group-form__section {
  border-radius: 16px;
  border: 1px solid #eef2f7;
  background: #fff;
  box-shadow: 0 8px 20px rgb(15 23 42 / 4%);
}

.up-group-form__setup {
  display: grid;
  grid-template-columns: 292px minmax(0, 1fr);
  gap: 16px;
  align-items: stretch;
  margin-bottom: 16px;
  padding: 14px;
}

.up-group-form__setup-summary {
  position: relative;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  gap: 14px;
  min-height: 100%;
  padding: 14px;
  border-radius: 14px;
  border: 1px solid #dcebff;
  background: linear-gradient(180deg, #f7fbff 0%, #f2f8ff 100%);
}

.up-group-form__setup-fields {
  min-width: 0;
  padding: 4px 4px 2px;
}

.up-group-form__setup-title {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 14px;
  white-space: nowrap;
}

.up-group-form__setup-desc {
  margin: 0;
  color: #94a3b8;
  font-size: 13px;
  line-height: 20px;
}

.up-group-form__head {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 0;
}

.up-group-form__icon {
  width: 42px;
  height: 42px;
  border-radius: 14px;
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
  gap: 4px;
}

.up-group-form__name {
  font-size: 18px;
  font-weight: 700;
  color: #262626;
}

.up-group-form__meta {
  font-size: 13px;
  color: #8c8c8c;
}

.up-group-form__summary-bottom {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: nowrap;
  white-space: nowrap;
}

.up-group-form__hero-tag {
  padding: 5px 10px;
  border-radius: 999px;
  background: #eff6ff;
  color: #2563eb;
  font-size: 12px;
  font-weight: 600;
  line-height: 1;
}

.up-group-form__hero-tag--muted {
  background: #f3f4f6;
  color: #6b7280;
}

.up-group-form__trash {
  position: absolute;
  top: 8px;
  right: 8px;
  border: none;
  background: none;
  color: #8c8c8c;
  padding: 6px;
  cursor: pointer;

  &:hover {
    color: #ff4d4f;
  }

  &:disabled {
    color: #d9d9d9;
    cursor: not-allowed;
  }

  &:disabled:hover {
    color: #d9d9d9;
  }
}

.up-group-form__smart-btn {
  flex-shrink: 0;
  height: 30px;
  padding: 0 12px;
  border-radius: 10px;
  background: #fff;
  border-color: #91caff;
  color: #1677ff;
  font-size: 13px;
  font-weight: 600;
  box-shadow: 0 4px 12px rgb(22 119 255 / 10%);

  :deep(span) {
    letter-spacing: 0.01em;
  }
}

.up-group-smart-modal__hint {
  margin: 0 0 12px;
  font-size: 13px;
  line-height: 1.55;
  color: #595959;
}

.up-group-smart-modal__form {
  margin-bottom: 0;
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 2px 16px;
}

.up-group-smart-modal__form :deep(.ant-form-item) {
  margin-bottom: 14px;
}

.up-group-smart-modal__preset {
  margin-top: 2px;
  padding: 10px 12px;
  border-top: 1px dashed #f0f0f0;
  background: #f8fbff;
  border-radius: 10px;
}

.up-group-smart-modal__preset-btn {
  height: auto;
  padding: 0;
  white-space: normal;
  text-align: left;
  line-height: 20px;
}

.up-group-smart-modal__footer-btns {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 16px;
  padding-top: 12px;
  border-top: 1px solid #f0f0f0;
}

@media (max-width: 640px) {
  .up-group-smart-modal__form {
    grid-template-columns: minmax(0, 1fr);
  }
}

.up-group-form__section {
  padding: 18px;
  margin-bottom: 16px;
}

.up-group-form__section-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 16px;
}

.up-group-form__section-head--slots {
  align-items: center;
}

.up-group-form__section-title {
  color: #1f2937;
  font-size: 15px;
  font-weight: 700;
  line-height: 1.2;
}

.up-group-form__section-desc {
  margin: 6px 0 0;
  color: #94a3b8;
  font-size: 13px;
  line-height: 20px;
}

.up-group-form__section-actions {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.up-group-form__section-chip {
  padding: 6px 10px;
  border-radius: 999px;
  background: #f8fafc;
  border: 1px solid #e5edf6;
  color: #475569;
  font-size: 12px;
  font-weight: 600;
}

.up-group-form__basic-grid {
  display: grid;
  grid-template-columns: minmax(150px, 0.72fr) minmax(280px, 1.28fr);
  gap: 14px;
}

.up-group-form__basic-grid--single {
  grid-template-columns: minmax(0, 1fr);
}

.up-group-form__field-card {
  padding: 14px 16px;
  border-radius: 12px;
  background: #fafcff;
  border: 1px solid #edf2f7;
}

.up-group-form__label {
  display: block;
  margin-bottom: 8px;
  color: #334155;
  font-size: 13px;
  font-weight: 700;
}

.up-group-form__field-card :deep(.ant-input) {
  border-radius: 10px;
}

.up-group-form__field-hint {
  margin: 0 0 8px;
  font-size: 12px;
  line-height: 1.5;
  color: #8c8c8c;
}

.up-group-form__teacher-select {
  width: 100%;

  :deep(.ant-select-selector) {
    border-radius: 10px;
  }
}

.up-group-form__slots-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14px;
}

.up-group-form__slot {
  padding: 16px;
  border-radius: 16px;
  background: linear-gradient(180deg, #fbfdff 0%, #f8fbff 100%);
  border: 1px solid #e8eef6;
  box-shadow: inset 0 1px 0 rgb(255 255 255 / 75%);
}

.up-group-form__slot-top {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 14px;
}

.up-group-form__slot-head {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
}

.up-group-form__slot-num {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: #e6f7ff;
  color: #1890ff;
  font-size: 13px;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.up-group-form__slot-head-text {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}

.up-group-form__slot-title {
  color: #111827;
  font-size: 15px;
  font-weight: 700;
}

.up-group-form__slot-subtitle {
  color: #94a3b8;
  font-size: 12px;
}

.up-group-form__slot-times {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.up-group-form__time-field {
  min-width: 0;

  :deep(.ant-picker) {
    width: 100%;
    border-radius: 10px;
  }
}

.up-group-form__time-label {
  display: block;
  margin-bottom: 8px;
  color: #64748b;
  font-size: 12px;
  font-weight: 600;
  line-height: 1;
}

.up-group-form__slot-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-top: 14px;
  padding-top: 12px;
  border-top: 1px dashed #dbe5f0;
}

.up-group-form__slot-state {
  color: #2563eb;
  font-size: 12px;
  line-height: 18px;
}

.up-group-form__slot-state--off {
  color: #94a3b8;
}

.up-group-form__picker {
  width: 100%;
}

.up-group-form__del {
  border: none;
  background: none;
  color: #ef4444;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  padding: 4px 0;
}

@media (max-width: 860px) {
  .up-group-form__setup,
  .up-group-form__basic-grid,
  .up-group-form__slots-grid {
    grid-template-columns: minmax(0, 1fr);
  }

  .up-group-form__section-head {
    flex-direction: column;
    align-items: flex-start;
  }
}

@media (max-width: 520px) {
  .up-group-form__setup {
    padding: 12px;
  }

  .up-group-form__head {
    flex-wrap: wrap;
  }

  .up-group-form__summary-bottom {
    width: 100%;
  }

  .up-group-form__section {
    padding: 16px;
  }

  .up-group-form__slot-times {
    grid-template-columns: minmax(0, 1fr);
  }

  .up-group-form__slot-footer {
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>
