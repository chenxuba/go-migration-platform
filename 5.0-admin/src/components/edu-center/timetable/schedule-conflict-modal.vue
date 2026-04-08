<script setup lang="ts">
import { CloseOutlined, ExclamationCircleFilled } from '@ant-design/icons-vue'
import type { TeachingScheduleValidationResult } from '@/api/edu-center/teaching-schedule'

type ConflictScheduleItem = NonNullable<TeachingScheduleValidationResult['existingSchedules']>[number]

const props = defineProps<{
  open: boolean
  validation?: TeachingScheduleValidationResult | null
  title?: string
  currentTitle?: string
  existingTitle?: string
  fallbackMessage?: string
  jumpingKey?: string
  getTimeExtraLabel?: (item: ConflictScheduleItem) => string
}>()

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'jump', item: ConflictScheduleItem): void
}>()

const modalOpen = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

const currentSchedules = computed(() => props.validation?.currentSchedules || [])
const existingSchedules = computed(() => props.validation?.existingSchedules || [])

function hasConflictType(item: { conflictTypes?: string[] }, type: string) {
  return (item.conflictTypes || []).includes(type)
}

function canJumpToItem(item: ConflictScheduleItem) {
  return Boolean(String(item?.date || '').trim() && String(item?.timeText || '').trim())
}

function jumpActionKey(item: ConflictScheduleItem) {
  return [
    String(item?.teacherId || '').trim(),
    String(item?.teacherName || '').trim(),
    String(item?.date || '').trim(),
    String(item?.timeText || '').trim(),
  ].join('|')
}

function resolveTimeExtraLabel(item: ConflictScheduleItem) {
  return props.getTimeExtraLabel?.(item) || ''
}
</script>

<template>
  <a-modal
    v-model:open="modalOpen"
    centered
    class="schedule-conflict-modal"
    :footer="null"
    :width="1180"
    :body-style="{ paddingTop: '0px' }"
    :keyboard="false"
    :closable="false"
    :mask-closable="true"
  >
    <template #title>
      <div class="schedule-conflict__titlebar">
        <span>{{ props.title || '冲突提示' }}</span>
        <a-button type="text" @click="modalOpen = false">
          <template #icon>
            <CloseOutlined />
          </template>
        </a-button>
      </div>
    </template>

    <div class="schedule-conflict">
      <div class="schedule-conflict__banner">
        <ExclamationCircleFilled />
        <span>{{ validation?.message || props.fallbackMessage || '当前日程与已有日程冲突' }}</span>
      </div>

      <section class="schedule-conflict__section">
        <div class="schedule-conflict__section-title">
          {{ props.currentTitle || '当前冲突日程' }}
        </div>
        <div class="schedule-conflict__table">
          <div class="schedule-conflict__head schedule-conflict__head--with-action">
            <span>日程名称</span>
            <span>日程类型</span>
            <span>{{ props.getTimeExtraLabel ? '上课时间 / 时段名称' : '上课时间' }}</span>
            <span>上课教师</span>
            <span>上课教室</span>
            <span>冲突类型</span>
            <span class="schedule-conflict__action-head">操作</span>
          </div>
          <div
            v-for="(item, index) in currentSchedules"
            :key="`${item.date}-${item.timeText}-${index}`"
            class="schedule-conflict__row schedule-conflict__row--with-action"
          >
            <span>{{ item.name }}</span>
            <span>{{ item.classTypeText }}</span>
            <span class="schedule-conflict__time-cell">
              <span>{{ item.date }} {{ item.timeText }}</span>
              <span v-if="resolveTimeExtraLabel(item)" class="schedule-conflict__time-extra">
                {{ resolveTimeExtraLabel(item) }}
              </span>
            </span>
            <span
              :class="{
                'schedule-conflict__cell--danger': hasConflictType(item, '老师'),
              }"
            >{{ item.teacherName || '-' }}</span>
            <span
              :class="{
                'schedule-conflict__cell--danger': hasConflictType(item, '教室'),
              }"
            >{{ item.classroomName || '-' }}</span>
            <span class="schedule-conflict__tags">
              <a-tag
                v-for="tag in item.conflictTypes || []"
                :key="tag"
                color="error"
                :bordered="false"
              >
                {{ tag }}冲突
              </a-tag>
            </span>
            <span class="schedule-conflict__action-cell">-</span>
          </div>
        </div>
      </section>

      <section class="schedule-conflict__section">
        <div class="schedule-conflict__section-title">
          {{ props.existingTitle || '与其冲突的日程' }}
        </div>
        <div class="schedule-conflict__table">
          <div class="schedule-conflict__head schedule-conflict__head--with-action">
            <span>日程名称</span>
            <span>日程类型</span>
            <span>{{ props.getTimeExtraLabel ? '上课时间 / 时段名称' : '上课时间' }}</span>
            <span>上课教师</span>
            <span>上课教室</span>
            <span>冲突学员</span>
            <span class="schedule-conflict__action-head">操作</span>
          </div>
          <div
            v-for="(item, index) in existingSchedules"
            :key="`${item.date}-${item.timeText}-${index}`"
            class="schedule-conflict__row schedule-conflict__row--with-action"
          >
            <span>{{ item.name }}</span>
            <span>{{ item.classTypeText }}</span>
            <span class="schedule-conflict__time-cell">
              <span>{{ item.date }} {{ item.timeText }}</span>
              <span v-if="resolveTimeExtraLabel(item)" class="schedule-conflict__time-extra">
                {{ resolveTimeExtraLabel(item) }}
              </span>
            </span>
            <span
              :class="{
                'schedule-conflict__cell--danger': hasConflictType(item, '老师'),
              }"
            >{{ item.teacherName || '-' }}</span>
            <span
              :class="{
                'schedule-conflict__cell--danger': hasConflictType(item, '教室'),
              }"
            >{{ item.classroomName || '-' }}</span>
            <span
              :class="{
                'schedule-conflict__cell--danger': hasConflictType(item, '学员'),
              }"
            >{{ (item.studentNames || []).join('、') || '-' }}</span>
            <span class="schedule-conflict__action-cell">
              <a-button
                v-if="canJumpToItem(item)"
                type="primary"
                ghost
                class="schedule-conflict__jump"
                :loading="props.jumpingKey === jumpActionKey(item)"
                @click.stop="emit('jump', item)"
              >
                定位到课程
              </a-button>
              <template v-else>
                -
              </template>
            </span>
          </div>
        </div>
      </section>
    </div>
  </a-modal>
</template>

<style scoped lang="less">
.schedule-conflict__titlebar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.schedule-conflict {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.schedule-conflict__banner {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 14px;
  border-radius: 12px;
  background: #fff7f7;
  color: #ff7875;
  font-size: 13px;
  font-weight: 600;
  border: 1px solid #ffe1e0;
}

.schedule-conflict__section {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.schedule-conflict__section-title {
  position: relative;
  padding-left: 14px;
  color: #1f2329;
  font-size: 16px;
  font-weight: 700;
}

.schedule-conflict__section-title::before {
  position: absolute;
  left: 0;
  top: 5px;
  width: 5px;
  height: 16px;
  border-radius: 999px;
  background: #1677ff;
  content: '';
}

.schedule-conflict__table {
  overflow: hidden;
  border-radius: 18px;
  background: #fff;
  border: 1px solid #edf2f7;
}

.schedule-conflict__head,
.schedule-conflict__row {
  display: grid;
  grid-template-columns: 1.5fr 1fr 1.5fr 1fr 1fr 1fr;
  gap: 16px;
  align-items: center;
  padding: 18px 20px;
}

.schedule-conflict__head {
  background: #f8fafc;
  color: #4b5563;
  font-size: 13px;
  font-weight: 700;
}

.schedule-conflict__row {
  border-top: 1px solid #f0f2f5;
  color: #1f2329;
  font-size: 14px;
  line-height: 1.6;
}

.schedule-conflict__time-cell {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.schedule-conflict__time-extra {
  color: #8c8c8c;
  font-size: 12px;
  line-height: 1.5;
}

.schedule-conflict__head--with-action,
.schedule-conflict__row--with-action {
  grid-template-columns: 1.4fr 1fr 1.5fr 1fr 1fr 1fr 156px;
}

.schedule-conflict__cell--danger {
  color: #ff4d4f;
  font-weight: 700;
}

.schedule-conflict__tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.schedule-conflict__action-head,
.schedule-conflict__action-cell {
  display: flex;
  align-items: center;
  justify-content: center;
}

.schedule-conflict__jump {
  width: 102px;
  min-width: 102px;
  height: 32px;
  padding: 0;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 400;
  flex: 0 0 auto;
}

.schedule-conflict__action-cell :deep(.ant-btn.schedule-conflict__jump) {
  width: 102px;
  min-width: 102px;
  height: 32px;
  padding: 0;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 400;
}

.schedule-conflict__jump:disabled {
  opacity: 0.5;
}

@media (max-width: 1200px) {
  .schedule-conflict__table {
    overflow-x: auto;
  }

  .schedule-conflict__head,
  .schedule-conflict__row {
    min-width: 960px;
  }

  .schedule-conflict__head--with-action,
  .schedule-conflict__row--with-action {
    min-width: 1116px;
  }
}
</style>
