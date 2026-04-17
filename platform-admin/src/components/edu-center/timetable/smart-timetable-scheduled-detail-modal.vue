<script setup>
import { computed } from 'vue'

const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
  detail: {
    type: Object,
    default: () => ({}),
  },
  deleting: {
    type: Boolean,
    default: false,
  },
})

const emit = defineEmits(['update:open', 'confirm'])

const modalOpen = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})
</script>

<template>
  <a-modal
    v-model:open="modalOpen"
    title="课程详情"
    width="620px"
    centered
    :ok-button-props="{ danger: true, loading: deleting }"
    :ok-text="detail.courseType === 1 ? (detail.isMain ? '删除本节' : '移除助教') : '删除本节'"
    cancel-text="关闭"
    :ok-cancel="true"
    @ok="$emit('confirm')"
  >
    <div class="st-scheduled-detail">
      <div class="st-scheduled-detail__hero">
        <span
          class="st-scheduled-detail__badge"
          :style="{ background: detail.modeColor || '#1677ff' }"
        >
          {{ detail.modeLabel }}
        </span>
        <span class="st-scheduled-detail__title">{{ detail.lessonTitle }}</span>
      </div>

      <div class="st-scheduled-detail__card">
        <div class="st-scheduled-detail__row">
          <span>上课时间</span>
          <strong>{{ detail.dateLabel }} · {{ detail.timeLabel }}</strong>
        </div>
        <div class="st-scheduled-detail__row">
          <span>上课老师</span>
          <strong>{{ detail.teacherName }}</strong>
        </div>
        <div class="st-scheduled-detail__row">
          <span>上课助教</span>
          <strong>{{ detail.assistantText }}</strong>
        </div>
        <div class="st-scheduled-detail__row">
          <span>所在组别</span>
          <strong>{{ detail.groupLabel }}</strong>
        </div>
        <div class="st-scheduled-detail__row">
          <span>上课学员</span>
          <strong>{{ detail.studentText }}</strong>
        </div>
      </div>

      <div v-if="detail.courseType === 1" class="st-scheduled-detail__hint st-scheduled-detail__hint--danger">
        {{ detail.isMain ? '删除这节 1v1 日程后，会立即从主教与助教课表中同步移除。' : '当前为助教视角。确认后仅移除这节课的当前助教，不会删除整节课。' }}
      </div>
      <div v-else class="st-scheduled-detail__hint st-scheduled-detail__hint--danger">
        删除这节班课后，会同步从主教与全部助教课表中移除，仅删除当前这节，不会删除整批班课。
      </div>
    </div>
  </a-modal>
</template>

<style scoped lang="less">
.st-scheduled-detail {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.st-scheduled-detail__hero {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.st-scheduled-detail__badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 56px;
  height: 30px;
  padding: 0 12px;
  border-radius: 999px;
  color: #fff;
  font-size: 14px;
  font-weight: 700;
}

.st-scheduled-detail__title {
  color: #262626;
  font-size: 18px;
  font-weight: 700;
}

.st-scheduled-detail__card {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 16px 18px;
  border-radius: 16px;
  background: #f8fafc;
  border: 1px solid #edf2f7;
}

.st-scheduled-detail__row {
  display: grid;
  grid-template-columns: 76px 1fr;
  gap: 12px;
  align-items: start;
  font-size: 14px;
  line-height: 22px;
}

.st-scheduled-detail__row > span {
  color: #8c8c8c;
}

.st-scheduled-detail__row > strong {
  color: #262626;
}

.st-scheduled-detail__hint {
  padding: 12px 14px;
  border-radius: 12px;
  background: #f6f8fb;
  color: #5b6475;
  font-size: 13px;
  line-height: 22px;
}

.st-scheduled-detail__hint--danger {
  background: #fff2f0;
  color: #cf1322;
}
</style>
