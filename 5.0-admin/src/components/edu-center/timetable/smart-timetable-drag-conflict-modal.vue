<script setup>
import { computed } from 'vue'

const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
  detail: {
    type: Object,
    default: () => ({
      summary: '',
      attempted: null,
      items: [],
    }),
  },
})

const emit = defineEmits(['update:open', 'jump'])

const modalOpen = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})
</script>

<template>
  <a-modal
    v-model:open="modalOpen"
    class="st-drag-conflict-modal"
    title="冲突详情"
    :footer="null"
    :width="700"
    :body-style="{ padding: '0 20px 18px' }"
    centered
  >
    <div class="st-drag-conflict">
      <div class="st-drag-conflict__summary">
        {{ detail.summary }}
      </div>

      <div v-if="detail?.attempted" class="st-drag-conflict__section">
        <div class="st-drag-conflict__section-title">
          你正在选择的空位
        </div>

        <div class="st-drag-conflict__attempt-card">
          <div class="st-drag-conflict__attempt-head">
            <span class="st-drag-conflict__badge">{{ detail.attempted.modeLabel || '1v1' }}</span>
            <span class="st-drag-conflict__attempt-title">待调课程信息</span>
          </div>

          <div class="st-drag-conflict__attempt-date">
            {{ detail.attempted.date }} {{ detail.attempted.week }} 第{{ detail.attempted.lessonIndex }}节
          </div>

          <div class="st-drag-conflict__attempt-student">
            排课学员
            <strong>{{ detail.attempted.studentText || detail.attempted.targetValue || '-' }}</strong>
          </div>

          <div class="st-drag-conflict__attempt-facts">
            <span class="st-drag-conflict__fact-chip">上课课程 {{ detail.attempted.courseName || '-' }}</span>
            <span class="st-drag-conflict__fact-chip">上课时间 {{ detail.attempted.timeText || '-' }}</span>
            <span class="st-drag-conflict__fact-chip">上课老师 {{ detail.attempted.teacherName || '-' }}</span>
            <span class="st-drag-conflict__fact-chip">上课助教 {{ detail.attempted.assistantText || '未安排' }}</span>
            <span class="st-drag-conflict__fact-chip">所在组别 {{ detail.attempted.groupLabel || '-' }}</span>
          </div>

          <div class="st-drag-conflict__attempt-tip">
            系统检测到这条调课信息与已有日程发生冲突，请先处理冲突后再继续调课。
          </div>
        </div>
      </div>

      <div class="st-drag-conflict__section">
        <div class="st-drag-conflict__section-title">
          冲突课程
        </div>

        <div class="st-drag-conflict__list">
          <div v-for="item in detail?.items || []" :key="item.key" class="st-drag-conflict__item">
            <div class="st-drag-conflict__item-main">
              <div class="st-drag-conflict__item-head">
                <div class="st-drag-conflict__item-name">
                  {{ item.name || '-' }}
                </div>
                <div class="st-drag-conflict__item-tags">
                  <a-tag color="blue" :bordered="false">
                    {{ item.classTypeText || '日程' }}
                  </a-tag>
                  <a-tag color="orange" :bordered="false">
                    {{ item.groupLabel || '当前组' }}
                  </a-tag>
                </div>
              </div>

              <div class="st-drag-conflict__item-time">
                {{ item.date }} {{ item.week }} · {{ item.timeText }}
              </div>

              <div class="st-drag-conflict__item-meta">
                教师：
                <span :class="{ 'st-drag-conflict__danger': item.hasTeacherConflict }">{{ item.teacherName || '-' }}</span>
                <span class="st-drag-conflict__sep">｜</span>
                助教：
                <span :class="{ 'st-drag-conflict__danger': item.hasAssistantConflict }">{{ item.assistantText || '-' }}</span>
                <span class="st-drag-conflict__sep">｜</span>
                学员：
                <span :class="{ 'st-drag-conflict__danger': item.hasStudentConflict }">{{ item.studentText || '-' }}</span>
              </div>

              <div v-if="item.classroomName && item.classroomName !== '-'" class="st-drag-conflict__item-meta">
                教室：
                <span :class="{ 'st-drag-conflict__danger': item.hasClassroomConflict }">{{ item.classroomName }}</span>
              </div>

              <div class="st-drag-conflict__reasons">
                <span class="st-drag-conflict__reasons-label">冲突原因：</span>
                <span
                  v-for="type in item.conflictTypes || []"
                  :key="type"
                  class="st-drag-conflict__reason-chip"
                >
                  {{ type }}冲突
                </span>
              </div>
            </div>

            <div class="st-drag-conflict__item-side">
              <a-button type="primary" ghost @click="$emit('jump', item)">
                定位到课程
              </a-button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </a-modal>
</template>

<style scoped lang="less">
.st-drag-conflict {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.st-drag-conflict__summary {
  padding: 12px 16px;
  border-radius: 14px;
  background: linear-gradient(180deg, #fff9ea 0%, #fff5dc 100%);
  color: #b86800;
  font-size: 13px;
  font-weight: 700;
  line-height: 1.45;
}

.st-drag-conflict__section {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.st-drag-conflict__section-title {
  color: #1f2329;
  font-size: 15px;
  font-weight: 800;
}

.st-drag-conflict__attempt-card,
.st-drag-conflict__item {
  border: 1px solid #e8eef7;
  border-radius: 16px;
  background: #fff;
  box-shadow: 0 10px 24px rgba(15, 23, 42, 0.04);
}

.st-drag-conflict__attempt-card {
  padding: 14px 16px 12px;
}

.st-drag-conflict__attempt-head {
  display: flex;
  align-items: center;
  gap: 8px;
}

.st-drag-conflict__badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 52px;
  height: 30px;
  padding: 0 12px;
  border-radius: 999px;
  background: linear-gradient(135deg, #2f7bff 0%, #1d64f2 100%);
  color: #fff;
  font-size: 14px;
  font-weight: 800;
}

.st-drag-conflict__attempt-title {
  color: #1f2329;
  font-size: 15px;
  font-weight: 800;
}

.st-drag-conflict__attempt-date {
  margin-top: 12px;
  color: #256be9;
  font-size: 15px;
  font-weight: 800;
}

.st-drag-conflict__attempt-student {
  margin-top: 12px;
  padding: 12px 14px;
  border-radius: 14px;
  background: #f7f9fd;
  color: #8a909c;
  font-size: 13px;
  font-weight: 600;
}

.st-drag-conflict__attempt-student strong {
  margin-left: 10px;
  color: #1f2329;
  font-size: 16px;
  font-weight: 800;
}

.st-drag-conflict__attempt-facts {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 12px;
}

.st-drag-conflict__fact-chip {
  display: inline-flex;
  align-items: center;
  min-height: 30px;
  padding: 0 12px;
  border: 1px solid #e6ecf5;
  border-radius: 12px;
  background: #f9fbff;
  color: #6b7280;
  font-size: 12px;
  font-weight: 600;
  line-height: 1.4;
}

.st-drag-conflict__attempt-tip {
  margin-top: 12px;
  color: #5b6475;
  font-size: 12px;
  line-height: 1.55;
}

.st-drag-conflict__list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.st-drag-conflict__item {
  display: flex;
  gap: 12px;
  justify-content: space-between;
  padding: 14px 16px;
}

.st-drag-conflict__item-main {
  flex: 1;
  min-width: 0;
}

.st-drag-conflict__item-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 8px;
}

.st-drag-conflict__item-name {
  color: #1f2329;
  font-size: 15px;
  font-weight: 800;
  line-height: 1.4;
}

.st-drag-conflict__item-tags {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.st-drag-conflict__item-time,
.st-drag-conflict__item-meta {
  margin-top: 6px;
  color: #4b5563;
  font-size: 12px;
  line-height: 1.55;
}

.st-drag-conflict__sep {
  margin: 0 6px;
  color: #c3cad8;
}

.st-drag-conflict__danger {
  color: #ff4d4f;
  font-weight: 700;
}

.st-drag-conflict__reasons {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
  margin-top: 8px;
}

.st-drag-conflict__reasons-label {
  color: #4b5563;
  font-size: 12px;
  font-weight: 600;
}

.st-drag-conflict__reason-chip {
  display: inline-flex;
  align-items: center;
  height: 24px;
  padding: 0 10px;
  border-radius: 999px;
  background: #fff1f0;
  color: #ff4d4f;
  font-size: 12px;
  font-weight: 700;
}

.st-drag-conflict__item-side {
  display: flex;
  align-items: flex-start;
  flex-shrink: 0;
  padding-top: 2px;
}

:deep(.st-drag-conflict-modal.ant-modal) {
  max-width: calc(100vw - 32px);
}

:deep(.st-drag-conflict-modal .ant-modal-content) {
  border-radius: 18px;
  overflow: hidden;
}

:deep(.st-drag-conflict-modal .ant-modal-header) {
  padding: 16px 20px 14px;
  border-bottom: 1px solid #eef2f7;
}

:deep(.st-drag-conflict-modal .ant-modal-title) {
  color: #1f2329;
  font-size: 17px;
  font-weight: 800;
}

:deep(.st-drag-conflict-modal .ant-modal-close) {
  inset-inline-end: 14px;
  top: 14px;
  width: 28px;
  height: 28px;
}

:deep(.st-drag-conflict-modal .ant-btn) {
  height: 34px;
  padding: 0 14px;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 700;
}

@media (max-width: 768px) {
  .st-drag-conflict__item {
    flex-direction: column;
  }

  .st-drag-conflict__item-side {
    padding-top: 0;
  }
}
</style>
