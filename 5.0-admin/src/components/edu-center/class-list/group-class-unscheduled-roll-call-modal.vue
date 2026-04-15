<script setup lang="ts">
import { ref } from 'vue'
import { getGroupClassUnscheduledRollCallContextApi, type GroupClassUnscheduledRollCallContextParams, type GroupClassUnscheduledRollCallContextResult } from '@/api/edu-center/roll-call'
import RollCallDrawer from '@/components/common/roll-call-drawer.vue'
import GroupClassScheduleModal from '@/components/edu-center/timetable/group-class-schedule-modal.vue'
import messageService from '@/utils/messageService'

const props = withDefaults(defineProps<{
  open?: boolean
  classId?: string
}>(), {
  open: false,
  classId: '',
})

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'updated'): void
  (e: 'confirmed'): void
}>()

const rollCallOpen = ref(false)
const rollCallContext = ref<GroupClassUnscheduledRollCallContextResult | null>(null)

async function handleStartRollCall(payload: GroupClassUnscheduledRollCallContextParams) {
  try {
    const res = await getGroupClassUnscheduledRollCallContextApi(payload)
    if (res.code !== 200 || !res.result?.detail || !res.result?.record)
      throw new Error(res.message || '加载未排课点名数据失败')
    const studentCount = Array.isArray(res.result.record?.students) ? res.result.record.students.length : 0
    if (studentCount <= 0) {
      const className = String(res.result.detail?.className || '').trim() || '当前班级'
      throw new Error(`${className}暂无在班学员，不能发起未排课点名`)
    }
    rollCallContext.value = res.result
    rollCallOpen.value = true
  }
  catch (error: any) {
    rollCallContext.value = null
    messageService.error(error?.response?.data?.message || error?.message || '加载未排课点名数据失败')
  }
}

function handleRollCallUpdated() {
  emit('updated')
}

function handleRollCallConfirmed() {
  emit('confirmed')
}
</script>

<template>
  <GroupClassScheduleModal
    :open="open"
    scenario="unscheduledRollCall"
    :initial-group-class-id="classId"
    @update:open="value => emit('update:open', value)"
    @updated="emit('updated')"
    @start-roll-call="handleStartRollCall"
  />

  <RollCallDrawer
    v-model:open="rollCallOpen"
    :unscheduled-context="rollCallContext"
    @updated="handleRollCallUpdated"
    @confirmed="handleRollCallConfirmed"
  />
</template>
