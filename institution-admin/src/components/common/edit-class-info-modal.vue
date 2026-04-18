<script setup lang="ts">
import { CloseOutlined } from '@ant-design/icons-vue'
import dayjs from 'dayjs'
import { computed, reactive, ref, watch } from 'vue'
import scheduleClassImage from '@/assets/images/timetable/schedule-class.png'
import scheduleOneToOneImage from '@/assets/images/timetable/schedule-one2one.png'
import { listClassroomsApi } from '@/api/business-settings/classroom'
import {
  updateTeachingRecordClassInfoApi,
  type TeachingRecordDetailResult,
} from '@/api/edu-center/class-record'
import StaffSelect from './staff-select.vue'
import messageService from '~@/utils/messageService'

interface ClassroomOption {
  label: string
  value: string
}

interface StaffPresetItem {
  id: string
  name: string
  nickName: string
}

const props = withDefaults(defineProps<{
  open: boolean
  detail?: TeachingRecordDetailResult | null
}>(), {
  detail: null,
})

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'updated'): void
}>()

const formRef = ref()
const classroomLoading = ref(false)
const classroomOptions = ref<ClassroomOption[]>([])
const submitting = ref(false)

const openModal = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

const formState = reactive({
  teacherId: undefined as string | undefined,
  assistantIds: [] as string[],
  classRoomId: undefined as string | undefined,
  teacherClassTime: 0,
})

const sourceCover = computed(() => (Number(props.detail?.timetableSourceType || 0) === 2 ? scheduleOneToOneImage : scheduleClassImage))
const titleText = computed(() => props.detail?.sourceName || props.detail?.lessonName || '-')
const timeText = computed(() => {
  const start = dayjs(props.detail?.startTime)
  const end = dayjs(props.detail?.endTime)
  if (!start.isValid() || !end.isValid())
    return '-'
  const weekMap = ['周日', '周一', '周二', '周三', '周四', '周五', '周六']
  return `${start.format('YYYY-MM-DD')}(${weekMap[start.day()] || '-'})${start.format('HH:mm')} ~ ${end.format('HH:mm')}`
})
const durationText = computed(() => {
  const start = dayjs(props.detail?.startTime)
  const end = dayjs(props.detail?.endTime)
  if (!start.isValid() || !end.isValid())
    return '-'
  const minutes = Math.max(end.diff(start, 'minute'), 0)
  return `${minutes}分钟`
})
const teacherPresetStaff = computed<StaffPresetItem[]>(() => {
  const teacher = (props.detail?.teacherList || []).find(item => Number(item.type || 0) === 1)
  if (!teacher?.teacherId)
    return []
  return [{
    id: String(teacher.teacherId),
    name: teacher.teacherName || String(teacher.teacherId),
    nickName: teacher.teacherName || String(teacher.teacherId),
  }]
})
const assistantPresetStaff = computed<StaffPresetItem[]>(() =>
  (props.detail?.teacherList || [])
    .filter(item => Number(item.type || 0) !== 1 && String(item.teacherId || '').trim())
    .map(item => ({
      id: String(item.teacherId),
      name: item.teacherName || String(item.teacherId),
      nickName: item.teacherName || String(item.teacherId),
    })),
)

function fillFormState() {
  const teacherList = Array.isArray(props.detail?.teacherList) ? props.detail?.teacherList : []
  const mainTeacher = teacherList.find(item => Number(item.type || 0) === 1)
  formState.teacherId = mainTeacher?.teacherId ? String(mainTeacher.teacherId) : undefined
  formState.assistantIds = teacherList
    .filter(item => Number(item.type || 0) !== 1 && String(item.teacherId || '').trim())
    .map(item => String(item.teacherId))
  formState.classRoomId = props.detail?.classRoomId && props.detail.classRoomId !== '0'
    ? String(props.detail.classRoomId)
    : undefined
  formState.teacherClassTime = Number(props.detail?.teacherClassTime || 0)
}

async function loadClassroomOptions() {
  classroomLoading.value = true
  try {
    const res = await listClassroomsApi({ enabledOnly: true })
    if (res.code !== 200) {
      messageService.error(res.message || '获取教室列表失败')
      return
    }
    classroomOptions.value = (Array.isArray(res.result) ? res.result : []).map(item => ({
      label: item.name || String(item.id),
      value: String(item.id),
    }))
    const currentClassroomId = String(props.detail?.classRoomId || '').trim()
    const currentClassroomName = String(props.detail?.classRoomName || '').trim()
    if (currentClassroomId && currentClassroomId !== '0' && currentClassroomName && !classroomOptions.value.some(item => item.value === currentClassroomId)) {
      classroomOptions.value = [
        {
          label: currentClassroomName,
          value: currentClassroomId,
        },
        ...classroomOptions.value,
      ]
    }
  }
  catch (error: any) {
    console.error('load classroom options failed', error)
    messageService.error(error?.message || '获取教室列表失败')
  }
  finally {
    classroomLoading.value = false
  }
}

function closeFun() {
  formRef.value?.clearValidate?.()
  openModal.value = false
}

async function handleSubmit() {
  if (!props.detail?.teachingRecordId || submitting.value)
    return
  try {
    await formRef.value?.validate?.()
    if (formState.teacherId && formState.assistantIds.includes(formState.teacherId)) {
      messageService.error('上课教师与上课助教不可为同一个人，请重新选择')
      return
    }
    submitting.value = true
    const res = await updateTeachingRecordClassInfoApi({
      teachingRecordId: String(props.detail.teachingRecordId),
      teacherId: String(formState.teacherId || ''),
      assistantIds: formState.assistantIds,
      classRoomId: formState.classRoomId || '0',
      teacherClassTime: Number(formState.teacherClassTime || 0),
    })
    if (res.code !== 200 || res.result !== true)
      throw new Error(res.message || '保存失败')
    messageService.success('保存成功')
    emit('updated')
    openModal.value = false
  }
  catch (error: any) {
    if (Array.isArray(error?.errorFields))
      return
    messageService.error(error?.response?.data?.message || error?.message || '保存失败')
  }
  finally {
    submitting.value = false
  }
}

watch(
  () => [props.open, props.detail?.teachingRecordId] as const,
  async ([open]) => {
    if (!open || !props.detail)
      return
    fillFormState()
    await loadClassroomOptions()
  },
  { immediate: true },
)
</script>

<template>
  <a-modal
    v-model:open="openModal"
    centered
    class="modal-content-box"
    :keyboard="false"
    :closable="false"
    :mask-closable="false"
    :width="800"
  >
    <template #title>
      <div class="text-5 flex justify-between flex-center">
        <span>编辑上课信息</span>
        <a-button type="text" class="close-btn" @click="closeFun">
          <template #icon>
            <CloseOutlined class="text-5 close-icon" />
          </template>
        </a-button>
      </div>
    </template>
    <div class="modal-body">
      <div class="modal-summary-card">
        <div class="contenter flex flex-center bg-white px6 py3" style="margin-bottom: 0;">
          <div class="avatarBox w-16 h-16 relative">
            <img width="64" height="64" :src="sourceCover" alt="">
          </div>
          <div class="info flex flex-1 ml-4 flex-col">
            <div class="top flex justify-between flex-center flex-1">
              <a-space>
                <div class="name text-5 font-800">
                  {{ titleText }}
                </div>
              </a-space>
            </div>
            <div class="bottom flex-1 flex flex-items-center mt-2">
              <div class="birthday flex-center">
                <span class="text-4 text-#222">{{ timeText }}</span>
                <span class="bg-#e6f0ff text-#06f text-3 px2 py1 rounded-10 ml2">{{ durationText }}</span>
              </div>
            </div>
          </div>
        </div>
        <a-divider class="my-0" />
        <div class="contenter scrollbar modal-form-body">
          <a-form ref="formRef" :model="formState" :label-col="{ span: 5 }" :wrapper-col="{ span: 18 }">
            <a-form-item label="上课教师" name="teacherId" :rules="[{ required: true, message: '请选择上课教师' }]">
              <StaffSelect
                v-model="formState.teacherId"
                placeholder="请选择上课教师"
                width="100%"
                :multiple="false"
                :status="0"
                :allow-clear="false"
                :preset-staff="teacherPresetStaff"
              />
            </a-form-item>
            <a-form-item label="上课助教" name="assistantIds">
              <StaffSelect
                v-model="formState.assistantIds"
                placeholder="请选择上课助教"
                width="100%"
                :multiple="true"
                :status="0"
                :preset-staff="assistantPresetStaff"
              />
            </a-form-item>
            <a-form-item label="上课教室" name="classRoomId">
              <a-select
                v-model:value="formState.classRoomId"
                :options="classroomOptions"
                :loading="classroomLoading"
                placeholder="请选择上课教室"
                allow-clear
              />
            </a-form-item>
            <a-form-item label="教师记录课时" name="teacherClassTime">
              <div class="flex flex-items-center">
                <a-input-number
                  v-model:value="formState.teacherClassTime"
                  placeholder="请输入"
                  :precision="2"
                  :min="0"
                  :max="100"
                  class="mr-8px"
                />课时
              </div>
            </a-form-item>
          </a-form>
        </div>
      </div>
    </div>
    <template #footer>
      <a-button danger ghost :disabled="submitting" @click="closeFun">
        关闭
      </a-button>
      <a-button type="primary" ghost :loading="submitting" @click="handleSubmit">
        确定
      </a-button>
    </template>
  </a-modal>
</template>

<style lang="less" scoped>
@keyframes icon-rotate {
  from {
    transform: rotate(0deg);
  }

  to {
    transform: rotate(180deg);
  }
}

.close-btn {
  &:hover {
    background: transparent;

    .close-icon {
      animation: icon-rotate 0.3s linear;
    }
  }
}

.modal-body {
  padding: 0;
}

.modal-summary-card {
  margin: 0 24px 24px;
  border-radius: 12px;
  background: #fafafa;
  overflow: hidden;
}

.contenter {
  padding: 24px;
  background: #fafafa;
}

.modal-form-body {
  margin: 0;
}
</style>

<style>
.modal-content-box .ant-modal-header {
  padding: 10px 16px !important;
  margin-bottom: 0;
}

.modal-content-box .ant-modal-body {
  padding: 0 !important;
}
</style>
