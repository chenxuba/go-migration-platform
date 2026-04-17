<script setup>
import { RightOutlined } from '@ant-design/icons-vue'
import { computed } from 'vue'

const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
  submitting: {
    type: Boolean,
    default: false,
  },
  copying: {
    type: Boolean,
    default: false,
  },
  detail: {
    type: Object,
    default: () => ({
      source: null,
      target: null,
    }),
  },
})

const emit = defineEmits(['update:open', 'confirm', 'copy'])

const modalOpen = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})
</script>

<template>
  <a-modal
    v-model:open="modalOpen"
    class="st-drag-confirm-modal"
    title="确认调整日程日期/时间？"
    :footer="null"
    :width="736"
    centered
  >
    <div class="st-drag-confirm">
      <div class="st-drag-confirm__summary">
        请确认课程将从原时段移动到新时段，确认后会立即更新课表。
      </div>

      <div v-if="detail?.warningText" class="st-drag-confirm__warning">
        {{ detail.warningText }}
      </div>

      <div class="st-drag-confirm__body">
        <div class="st-drag-confirm__card">
          <div class="st-drag-confirm__tag">
            调整前
          </div>
          <div class="st-drag-confirm__date">
            {{ detail?.source?.dateLabel || '-' }}
          </div>
          <div class="st-drag-confirm__time">
            {{ detail?.source?.timeLabel || '-' }}
          </div>
          <div class="st-drag-confirm__title">
            {{ detail?.source?.lessonTitle || '-' }}
          </div>
          <div class="st-drag-confirm__meta">
            {{ detail?.source?.courseName || '-' }}
          </div>
          <div class="st-drag-confirm__meta">
            {{ detail?.source?.studentText || '-' }}
          </div>
          <div class="st-drag-confirm__meta">
            主教：{{ detail?.source?.teacherText || '-' }}
          </div>
          <div class="st-drag-confirm__meta">
            助教：{{ detail?.source?.assistantText || '未安排' }}
          </div>
        </div>

        <div class="st-drag-confirm__arrow">
          <span class="st-drag-confirm__arrow-dot">
            <RightOutlined />
          </span>
        </div>

        <div class="st-drag-confirm__card st-drag-confirm__card--target">
          <div class="st-drag-confirm__tag st-drag-confirm__tag--target">
            调整后
          </div>
          <div class="st-drag-confirm__date st-drag-confirm__date--target">
            {{ detail?.target?.dateLabel || '-' }}
          </div>
          <div class="st-drag-confirm__time st-drag-confirm__time--target">
            {{ detail?.target?.timeLabel || '-' }}
          </div>
          <div class="st-drag-confirm__title">
            {{ detail?.target?.lessonTitle || '-' }}
          </div>
          <div class="st-drag-confirm__meta">
            {{ detail?.target?.courseName || '-' }}
          </div>
          <div class="st-drag-confirm__meta">
            {{ detail?.target?.studentText || '-' }}
          </div>
          <div class="st-drag-confirm__meta">
            主教：
            <span :class="{ 'st-drag-confirm__meta-highlight': detail?.target?.teacherChanged }">
              {{ detail?.target?.teacherText || '-' }}
            </span>
          </div>
          <div class="st-drag-confirm__meta">
            助教：
            <span :class="{ 'st-drag-confirm__meta-highlight': detail?.target?.assistantChanged }">
              {{ detail?.target?.assistantText || '未安排' }}
            </span>
          </div>
        </div>
      </div>

      <div class="st-drag-confirm__footer">
        <a-button :loading="copying" :disabled="submitting" @click="$emit('copy')">
          复制课程
        </a-button>
        <a-button @click="modalOpen = false">
          取消
        </a-button>
        <a-button type="primary" :loading="submitting" :disabled="copying" @click="$emit('confirm')">
          确定
        </a-button>
      </div>
    </div>
  </a-modal>
</template>

<style scoped lang="less">
.st-drag-confirm {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.st-drag-confirm__summary {
  color: #6b7280;
  font-size: 14px;
  line-height: 1.5;
}

.st-drag-confirm__warning {
  padding: 10px 12px;
  border-radius: 12px;
  background: #fff7e6;
  color: #ad6800;
  font-size: 13px;
  line-height: 1.5;
}

.st-drag-confirm__body {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 40px minmax(0, 1fr);
  align-items: center;
  gap: 12px;
}

.st-drag-confirm__card {
  min-height: 238px;
  padding: 18px 18px 16px;
  border: 1px solid #e8edf6;
  border-radius: 18px;
  background: linear-gradient(180deg, #ffffff 0%, #fbfcff 100%);
  box-shadow: 0 12px 24px rgba(15, 23, 42, 0.05);
  box-sizing: border-box;
}

.st-drag-confirm__card--target {
  background: linear-gradient(135deg, #f6f9ff 0%, #edf4ff 100%);
  border-color: #d8e6ff;
  box-shadow:
    0 14px 30px rgba(22, 104, 255, 0.08),
    inset 0 1px 0 rgba(255, 255, 255, 0.72);
}

.st-drag-confirm__tag {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 62px;
  height: 26px;
  padding: 0 10px;
  border-radius: 999px;
  background: #f3f6fb;
  color: #667085;
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.02em;
}

.st-drag-confirm__tag--target {
  background: rgba(22, 104, 255, 0.1);
  color: #1668ff;
}

.st-drag-confirm__date,
.st-drag-confirm__time,
.st-drag-confirm__title {
  color: #1f2329;
  font-weight: 700;
}

.st-drag-confirm__date {
  margin-top: 16px;
  font-size: 17px;
  line-height: 1.4;
}

.st-drag-confirm__time {
  margin-top: 2px;
  font-size: 18px;
  line-height: 1.45;
}

.st-drag-confirm__date--target,
.st-drag-confirm__time--target {
  color: #1668ff;
}

.st-drag-confirm__title {
  margin-top: 22px;
  font-size: 18px;
  line-height: 1.45;
  word-break: break-word;
}

.st-drag-confirm__meta {
  margin-top: 10px;
  color: #6b7280;
  font-size: 15px;
  line-height: 1.4;
  word-break: break-word;
}

.st-drag-confirm__meta-highlight {
  color: #1668ff;
  font-weight: 700;
}

.st-drag-confirm__arrow {
  display: flex;
  align-items: center;
  justify-content: center;
}

.st-drag-confirm__arrow-dot {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: 999px;
  background: #f5f7fb;
  color: #7a7a7a;
  font-size: 18px;
  box-shadow: inset 0 0 0 1px #e6ebf2;
}

.st-drag-confirm__footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  padding-top: 4px;
}

:deep(.st-drag-confirm-modal .ant-modal-content) {
  border-radius: 22px;
  overflow: hidden;
  box-shadow: 0 24px 70px rgba(15, 23, 42, 0.2);
}

:deep(.st-drag-confirm-modal .ant-modal-header) {
  padding: 18px 24px 16px;
  border-bottom: 1px solid #eef0f4;
}

:deep(.st-drag-confirm-modal .ant-modal-title) {
  color: #1f2329;
  font-size: 18px;
  font-weight: 800;
}

:deep(.st-drag-confirm-modal .ant-modal-close) {
  inset-inline-end: 18px;
  top: 18px;
  width: 28px;
  height: 28px;
  color: #8c8c8c;
}

:deep(.st-drag-confirm-modal .ant-modal-body) {
  padding: 16px 24px 16px;
}

:deep(.st-drag-confirm-modal .ant-btn) {
  min-width: 86px;
  height: 38px;
  border-radius: 12px;
  font-size: 15px;
  font-weight: 700;
}
</style>
