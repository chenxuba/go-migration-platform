<script setup lang="ts">
import type { RollCallTeachingRecordStudent } from '@/api/edu-center/roll-call'
import type { TeachingScheduleStudentCandidate } from '@/api/edu-center/teaching-schedule'
import RollCallAddStudentModal from './roll-call-add-student-modal.vue'

const props = withDefaults(defineProps<{
  open?: boolean
  title?: string
  scheduleId?: string
  studentType?: number
}>(), {
  open: false,
  title: '添加学员',
  scheduleId: '',
  studentType: 4,
})

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'success', payload?: { students?: RollCallTeachingRecordStudent[], studentIds?: string[], candidates?: TeachingScheduleStudentCandidate[] }): void
}>()
</script>

<template>
  <RollCallAddStudentModal
    :open="props.open"
    :title="props.title"
    :schedule-id="props.scheduleId"
    :student-type="props.studentType"
    defer-schedule-commit
    @update:open="emit('update:open', $event)"
    @success="emit('success', $event)"
  />
</template>
