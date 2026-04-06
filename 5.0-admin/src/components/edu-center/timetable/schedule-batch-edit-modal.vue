<script setup lang="ts">
import { CloseOutlined } from '@ant-design/icons-vue'
import dayjs from 'dayjs'
import { computed, reactive, ref, watch } from 'vue'
import { listClassroomsApi } from '@/api/business-settings/classroom'
import { type TeachingScheduleItem, batchUpdateTeachingSchedulesApi } from '@/api/edu-center/teaching-schedule'
import StaffSelect from '@/components/common/staff-select.vue'
import messageService from '@/utils/messageService'

interface ClassroomOption {
  label: string
  value: string
}

const props = defineProps<{
  open: boolean
  schedule?: TeachingScheduleItem | null
}>()

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'updated'): void
  (e: 'editBatchPlan', schedule: TeachingScheduleItem): void
}>()

const modalOpen = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

const classroomOptions = ref<ClassroomOption[]>([])
const classroomLoading = ref(false)
const submitting = ref(false)

const formState = reactive({
  teacherId: undefined as number | undefined,
  assistantIds: [] as number[],
  classroomId: undefined as string | undefined,
  startTime: undefined as string | undefined,
  endTime: undefined as string | undefined,
})

const teacherPresetStaff = computed(() => {
  const row = props.schedule
  if (!row?.teacherId)
    return []
  return [{
    id: row.teacherId,
    name: row.teacherName || row.teacherId,
    nickName: row.teacherName || row.teacherId,
  }]
})

const assistantPresetStaff = computed(() => {
  const ids = Array.isArray(props.schedule?.assistantIds) ? props.schedule?.assistantIds : []
  const names = Array.isArray(props.schedule?.assistantNames) ? props.schedule?.assistantNames : []
  return ids.map((id, index) => ({
    id,
    name: names[index] || id,
    nickName: names[index] || id,
  }))
})

const titleText = computed(() =>
  Number(props.schedule?.batchSize || 1) > 1 ? '批量修改本批次日程' : '编辑日程',
)

const canEditBatchPlan = computed(() =>
  Number(props.schedule?.batchSize || 1) > 1 && Number(props.schedule?.classType || 0) === 2,
)

const helperText = computed(() => {
  const count = Number(props.schedule?.batchSize || 1)
  if (count > 1)
    return `本次会同步修改该批次下的 ${count} 节日程，系统会再次校验老师与教室冲突。`
  return '本次修改仅影响当前这节日程，保存前会再次校验老师与教室冲突。'
})

async function loadClassroomOptions() {
  classroomLoading.value = true
  try {
    const res = await listClassroomsApi({ enabledOnly: true })
    if (res.code === 200) {
      classroomOptions.value = (Array.isArray(res.result) ? res.result : []).map(item => ({
        label: item.name || String(item.id),
        value: String(item.id),
      }))
      const current = props.schedule
      if (current?.classroomId && current.classroomName && !classroomOptions.value.some(item => item.value === String(current.classroomId))) {
        classroomOptions.value = [{
          label: current.classroomName,
          value: String(current.classroomId),
        }, ...classroomOptions.value]
      }
      return
    }
    messageService.error(res.message || '获取教室列表失败')
  }
  catch (error: any) {
    console.error('load classroom options failed', error)
    messageService.error(error?.message || '获取教室列表失败')
  }
  finally {
    classroomLoading.value = false
  }
}

watch(() => [props.open, props.schedule], async ([open]) => {
  if (!open || !props.schedule)
    return
  formState.teacherId = props.schedule.teacherId ? Number(props.schedule.teacherId) : undefined
  formState.assistantIds = Array.isArray(props.schedule.assistantIds)
    ? props.schedule.assistantIds.map(id => Number(id)).filter(Boolean)
    : []
  formState.classroomId = props.schedule.classroomId ? String(props.schedule.classroomId) : undefined
  formState.startTime = props.schedule.startAt ? dayjs(props.schedule.startAt).format('HH:mm') : undefined
  formState.endTime = props.schedule.endAt ? dayjs(props.schedule.endAt).format('HH:mm') : undefined
  await loadClassroomOptions()
}, { immediate: true })

function closeModal() {
  modalOpen.value = false
}

function handleEditBatchPlan() {
  if (!props.schedule)
    return
  emit('editBatchPlan', props.schedule)
  closeModal()
}

async function submitForm() {
  if (!formState.teacherId) {
    messageService.warning('请选择上课教师')
    return
  }
  if (!formState.startTime || !formState.endTime) {
    messageService.warning('请补全开始与结束时间')
    return
  }
  submitting.value = true
  try {
    const current = props.schedule
    const res = await batchUpdateTeachingSchedulesApi({
      batchNo: current?.batchNo || '',
      ids: current?.batchNo ? [] : [String(current?.id || '')],
      teacherId: String(formState.teacherId),
      assistantIds: formState.assistantIds.map(id => String(id)),
      classroomId: formState.classroomId || '',
      startTime: formState.startTime,
      endTime: formState.endTime,
    })
    if (res.code !== 200)
      throw new Error(res.message || '修改日程失败')
    messageService.success(Number(current?.batchSize || 1) > 1 ? '批量修改成功' : '修改成功')
    emit('updated')
    closeModal()
  }
  catch (error: any) {
    console.error('batch update teaching schedules failed', error)
    messageService.error(error?.message || '修改日程失败')
  }
  finally {
    submitting.value = false
  }
}
</script>

<template>
  <a-modal
    v-model:open="modalOpen"
    centered
    class="schedule-batch-edit-modal"
    :keyboard="false"
    :closable="false"
    :mask-closable="false"
    :width="760"
    :footer="null"
  >
    <template #title>
      <div class="schedule-batch-edit__titlebar">
        <span>{{ titleText }}</span>
        <a-button type="text" @click="closeModal">
          <template #icon>
            <CloseOutlined />
          </template>
        </a-button>
      </div>
    </template>

    <div class="schedule-batch-edit">
      <div class="schedule-batch-edit__hero">
        <div class="schedule-batch-edit__name">
          {{ schedule?.teachingClassName || '-' }}
        </div>
        <div class="schedule-batch-edit__meta">
          {{ schedule?.lessonName || '-' }} · {{ schedule?.studentName || '-' }}
        </div>
        <div class="schedule-batch-edit__hint">
          {{ helperText }}
        </div>
        <a-button
          v-if="canEditBatchPlan"
          type="link"
          class="schedule-batch-edit__hero-action"
          @click="handleEditBatchPlan"
        >
          编辑生成条件
        </a-button>
      </div>

      <div class="schedule-batch-edit__grid">
        <div class="schedule-batch-edit__field">
          <span class="schedule-batch-edit__label required">上课教师</span>
          <StaffSelect
            v-model="formState.teacherId"
            placeholder="请选择上课教师"
            width="100%"
            :multiple="false"
            :status="0"
            :allow-clear="false"
            :preset-staff="teacherPresetStaff"
          />
        </div>

        <div class="schedule-batch-edit__field">
          <span class="schedule-batch-edit__label">上课助教</span>
          <StaffSelect
            v-model="formState.assistantIds"
            placeholder="可不选"
            width="100%"
            :multiple="true"
            :status="0"
            :allow-clear="true"
            :preset-staff="assistantPresetStaff"
          />
        </div>

        <div class="schedule-batch-edit__field">
          <span class="schedule-batch-edit__label">上课教室</span>
          <a-select
            v-model:value="formState.classroomId"
            :options="classroomOptions"
            :loading="classroomLoading"
            placeholder="可不选"
            allow-clear
          />
        </div>

        <div class="schedule-batch-edit__field schedule-batch-edit__field--time">
          <span class="schedule-batch-edit__label required">上课时间</span>
          <div class="schedule-batch-edit__time-row">
            <a-time-picker
              v-model:value="formState.startTime"
              value-format="HH:mm"
              format="HH:mm"
              placeholder="开始"
            />
            <span>至</span>
            <a-time-picker
              v-model:value="formState.endTime"
              value-format="HH:mm"
              format="HH:mm"
              placeholder="结束"
            />
          </div>
        </div>
      </div>

      <div class="schedule-batch-edit__footer">
        <a-button @click="closeModal">
          取消
        </a-button>
        <a-button type="primary" :loading="submitting" @click="submitForm">
          保存
        </a-button>
      </div>
    </div>
  </a-modal>
</template>

<style scoped lang="less">
.schedule-batch-edit {
  display: flex;
  flex-direction: column;
  gap: 18px;
  padding: 6px 0 4px;
}

.schedule-batch-edit__titlebar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.schedule-batch-edit__hero {
  padding: 16px;
  border: 1px solid #e8edf3;
  border-radius: 16px;
  background: #f8fbff;
}

.schedule-batch-edit__name {
  color: #1f2329;
  font-size: 18px;
  font-weight: 700;
  line-height: 1.4;
}

.schedule-batch-edit__meta {
  margin-top: 4px;
  color: #6b7280;
  font-size: 13px;
  line-height: 1.6;
}

.schedule-batch-edit__hint {
  margin-top: 10px;
  color: #1677ff;
  font-size: 12px;
  line-height: 1.6;
}

.schedule-batch-edit__hero-action {
  margin-top: 8px;
  padding-left: 0;
}

.schedule-batch-edit__grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14px 16px;
}

.schedule-batch-edit__field {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.schedule-batch-edit__field--time {
  grid-column: 1 / -1;
}

.schedule-batch-edit__label {
  color: #4b5563;
  font-size: 13px;
  font-weight: 600;
}

.schedule-batch-edit__label.required::before {
  margin-right: 4px;
  color: #ff4d4f;
  content: '*';
}

.schedule-batch-edit__time-row {
  display: flex;
  align-items: center;
  gap: 10px;
}

.schedule-batch-edit__time-row :deep(.ant-picker) {
  flex: 1;
}

.schedule-batch-edit__footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

@media (max-width: 768px) {
  .schedule-batch-edit__grid {
    grid-template-columns: minmax(0, 1fr);
  }

  .schedule-batch-edit__time-row {
    flex-direction: column;
    align-items: stretch;
  }
}
</style>
