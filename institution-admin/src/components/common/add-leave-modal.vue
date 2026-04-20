<script setup>
import { computed, nextTick, onBeforeUnmount, reactive, ref, watch } from 'vue'
import { CloseOutlined, PlusOutlined } from '@ant-design/icons-vue'
import { Empty, Upload, message } from 'ant-design-vue'
import dayjs from 'dayjs'
import { debounce } from 'lodash-es'
import * as qiniu from 'qiniu-js'
import StudentSelect from './student-select.vue'
import { getQiniuToken } from '@/api/qiniu'
import { createLeaveAgentApi, previewLeaveSchedulesApi } from '@/api/home-center/leave'
import messageService from '@/utils/messageService'

const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
})

const emit = defineEmits(['update:open', 'success'])

const openModal = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

const leaveTypeOptions = [
  { label: '事假', value: 1 },
  { label: '病假', value: 2 },
  { label: '休学', value: 3 },
]

const weekdayText = ['周日', '周一', '周二', '周三', '周四', '周五', '周六']
const simpleImage = Empty.PRESENTED_IMAGE_SIMPLE

const formRef = ref(null)
const studentSelectRef = ref(null)
const scheduleLoading = ref(false)
const submitLoading = ref(false)
const scheduleList = ref([])
const fileList = ref([])
const previewVisible = ref(false)
const previewImage = ref('')
const previewTitle = ref('')

const formState = reactive({
  studentId: undefined,
  startTime: undefined,
  endTime: undefined,
  leaveType: undefined,
  reason: '',
  proofMaterials: [],
  remark: '',
})

const startWeekday = computed(() => getWeekday(formState.startTime))
const endWeekday = computed(() => getWeekday(formState.endTime))
const matchedLessonCount = computed(() => scheduleList.value.length)
const canPreviewSchedules = computed(() =>
  Boolean(
    formState.studentId
    && formState.startTime
    && formState.endTime
    && !dayjs(formState.endTime).isBefore(formState.startTime),
  ),
)

let previewRequestSeed = 0

function getWeekday(value) {
  if (!value)
    return ''
  return weekdayText[dayjs(value).day()]
}

function formatRequestTime(value) {
  return value ? dayjs(value).format('YYYY-MM-DD HH:mm:ss') : ''
}

function formatDateTime(value) {
  if (!value)
    return '-'
  const parsed = dayjs(value)
  return parsed.isValid() ? parsed.format('YYYY-MM-DD HH:mm') : String(value).replace('T', ' ').slice(0, 16)
}

function formatTimeOnly(value) {
  if (!value)
    return '--:--'
  const parsed = dayjs(value)
  return parsed.isValid() ? parsed.format('HH:mm') : '--:--'
}

function formatScheduleTitle(item) {
  return item.teachingClassName || item.lessonName || '-'
}

function disabledStartDate(current) {
  return current && current < dayjs().startOf('day')
}

function disabledEndDate(current) {
  if (!formState.startTime)
    return current && current < dayjs().startOf('day')
  return current && current < dayjs(formState.startTime).startOf('day')
}

function handleStartTimeChange(date) {
  if (formState.endTime && date && dayjs(date).isAfter(formState.endTime)) {
    formState.endTime = undefined
  }
}

async function fetchPreviewSchedules() {
  if (!canPreviewSchedules.value) {
    scheduleList.value = []
    return
  }

  const currentSeed = ++previewRequestSeed

  try {
    scheduleLoading.value = true
    const res = await previewLeaveSchedulesApi({
      studentId: String(formState.studentId),
      startTime: formatRequestTime(formState.startTime),
      endTime: formatRequestTime(formState.endTime),
    })

    if (currentSeed !== previewRequestSeed)
      return

    if (res.code === 200) {
      scheduleList.value = res.result?.list || []
    }
  }
  catch (error) {
    if (currentSeed !== previewRequestSeed)
      return
    scheduleList.value = []
    console.error('预览请假课节失败:', error)
  }
  finally {
    if (currentSeed === previewRequestSeed) {
      scheduleLoading.value = false
    }
  }
}

const debouncedFetchPreviewSchedules = debounce(fetchPreviewSchedules, 300)

watch(
  () => [formState.studentId, formState.startTime, formState.endTime],
  () => {
    if (!props.open)
      return

    if (!canPreviewSchedules.value) {
      previewRequestSeed += 1
      scheduleLoading.value = false
      scheduleList.value = []
      debouncedFetchPreviewSchedules.cancel()
      return
    }

    debouncedFetchPreviewSchedules()
  },
)

watch(
  () => props.open,
  async (visible) => {
    if (visible) {
      await nextTick()
      studentSelectRef.value?.reset?.()
      return
    }

    resetState()
  },
)

function resetState() {
  debouncedFetchPreviewSchedules.cancel()
  previewRequestSeed += 1
  formRef.value?.resetFields()
  studentSelectRef.value?.reset?.()
  scheduleLoading.value = false
  submitLoading.value = false
  scheduleList.value = []
  fileList.value = []
  previewVisible.value = false
  previewImage.value = ''
  previewTitle.value = ''
  formState.studentId = undefined
  formState.startTime = undefined
  formState.endTime = undefined
  formState.leaveType = undefined
  formState.reason = ''
  formState.proofMaterials = []
  formState.remark = ''
}

function closeFun() {
  openModal.value = false
}

function getBase64(file) {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.readAsDataURL(file)
    reader.onload = () => resolve(reader.result)
    reader.onerror = error => reject(error)
  })
}

function ensureUploadUrl(raw) {
  if (raw == null || raw === '')
    return ''
  if (typeof raw === 'string')
    return raw
  if (typeof raw === 'object' && raw.url != null)
    return ensureUploadUrl(raw.url)
  return String(raw)
}

function normalizeUploadFileItem(file) {
  const url = ensureUploadUrl(file.url ?? file.response?.url ?? file.thumbUrl ?? file.preview)
  const next = { ...file, url }
  if (file.thumbUrl != null) {
    const thumbUrl = ensureUploadUrl(file.thumbUrl)
    if (thumbUrl)
      next.thumbUrl = thumbUrl
    else
      delete next.thumbUrl
  }
  return next
}

function updateProofMaterials() {
  formState.proofMaterials = fileList.value
    .map(file => ensureUploadUrl(file.url || file.response?.url || ''))
    .filter(Boolean)
}

function beforeUpload(file) {
  if (fileList.value.length >= 3) {
    message.warning('最多上传 3 张图片')
    return Upload.LIST_IGNORE
  }

  const isImage = ['image/jpeg', 'image/jpg', 'image/png', 'image/bmp'].includes(file.type)
  if (!isImage) {
    message.error('只能上传 BMP、JPG、JPEG、PNG 格式的图片')
    return Upload.LIST_IGNORE
  }

  const isLt4M = file.size / 1024 / 1024 < 4
  if (!isLt4M) {
    message.error('图片大小不能超过 4MB')
    return Upload.LIST_IGNORE
  }

  return true
}

async function handlePreview(file) {
  if (!file.url && !file.preview) {
    file.preview = await getBase64(file.originFileObj)
  }
  previewImage.value = file.url || file.preview
  previewVisible.value = true
  const urlStr = typeof file.url === 'string' ? file.url : ''
  previewTitle.value = file.name || (urlStr ? urlStr.substring(urlStr.lastIndexOf('/') + 1) : 'preview')
}

function handlePreviewCancel() {
  previewVisible.value = false
  previewTitle.value = ''
}

function handleUploadChange(info) {
  fileList.value = info.fileList.map(normalizeUploadFileItem)
  updateProofMaterials()
}

function handleRemove() {
  nextTick(() => {
    updateProofMaterials()
  })
  return true
}

function handleProofUpload(options) {
  const { file, onSuccess, onError, onProgress } = options
  const rawFile = file.originFileObj || file

  const checkResult = beforeUpload(rawFile)
  if (checkResult !== true) {
    onError?.(new Error('文件校验未通过'))
    return
  }

  ;(async () => {
    try {
      const tokenRes = await getQiniuToken()
      const { token, uuid, buckethostname } = tokenRes.result || {}

      if (!token || !uuid || !buckethostname) {
        throw new Error('上传凭证缺失')
      }

      const ext = rawFile.name?.includes('.')
        ? rawFile.name.substring(rawFile.name.lastIndexOf('.'))
        : (rawFile.type === 'image/png' ? '.png' : '.jpg')
      const key = `leave-proof/${uuid}${ext}`

      const observable = qiniu.upload(rawFile, key, token, {
        fname: rawFile.name,
        mimeType: rawFile.type,
      }, {
        useCdnDomain: true,
        region: qiniu.region.z0,
      })

      observable.subscribe({
        next(res) {
          onProgress?.({ percent: Math.floor(res.total.percent) })
        },
        error(err) {
          console.error('请假佐证上传失败:', err)
          message.error(`上传失败：${err?.message || '未知错误'}`)
          onError?.(err)
        },
        complete(res) {
          const fileUrl = `${buckethostname}${res.key}`
          onSuccess?.({ url: fileUrl }, file)
        },
      })
    }
    catch (error) {
      console.error('获取七牛上传凭证失败:', error)
      message.error('获取上传凭证失败')
      onError?.(error)
    }
  })()
}

async function handleSubmit() {
  try {
    await formRef.value?.validate()

    if (dayjs(formState.endTime).isBefore(formState.startTime)) {
      message.warning('结束时间不能早于开始时间')
      return
    }

    await fetchPreviewSchedules()
    if (!scheduleList.value.length) {
      message.warning('请假期间没有可请假的相关课节')
      return
    }

    submitLoading.value = true

    const res = await createLeaveAgentApi({
      studentId: String(formState.studentId),
      startTime: formatRequestTime(formState.startTime),
      endTime: formatRequestTime(formState.endTime),
      leaveType: formState.leaveType,
      reason: formState.reason?.trim?.() || '',
      proofMaterials: [...formState.proofMaterials],
      remark: formState.remark?.trim?.() || '',
    })

    if (res.code === 200) {
      message.success(res.result?.status === 1 ? '请假申请已提交，等待审批处理' : '请假代办成功')
      emit('success', res.result)
      openModal.value = false
    }
  }
  catch (error) {
    if (error?.errorFields) {
      return
    }
    console.error('提交请假代办失败:', error)
    messageService.error('提交请假代办失败')
  }
  finally {
    submitLoading.value = false
  }
}

onBeforeUnmount(() => {
  debouncedFetchPreviewSchedules.cancel()
})
</script>

<template>
  <a-modal
    v-model:open="openModal"
    style="top: 40px;"
    class="modal-content-box"
    :keyboard="false"
    :closable="false"
    :mask-closable="false"
    :width="700"
  >
    <template #title>
      <div class="text-5 flex justify-between flex-center">
        <span>请假代办</span>
        <a-button type="text" class="close-btn" @click="closeFun">
          <template #icon>
            <CloseOutlined class="text-5 close-icon" />
          </template>
        </a-button>
      </div>
    </template>
    <div class="contenter scrollbar">
      <a-form ref="formRef" :model="formState" label-align="right" :label-col="{ span: 7 }">
        <a-form-item label="选择学员" name="studentId" :rules="[{ required: true, message: '请选择学员' }]" class="w-80%">
          <StudentSelect
            ref="studentSelectRef"
            v-model="formState.studentId"
            width="100%"
            placeholder="请选择学员"
            allow-clear
            :student-status="1"
          />
        </a-form-item>

        <a-form-item label="开始时间" name="startTime" :rules="[{ required: true, message: '请选择开始时间' }]" class="w-80%">
          <div class="flex flex-center relative">
            <a-date-picker
              v-model:value="formState.startTime"
              :show-time="{ format: 'HH:mm' }"
              format="YYYY-MM-DD HH:mm"
              placeholder="请选择开始时间"
              class="w-100%"
              :disabled-date="disabledStartDate"
              @change="handleStartTimeChange"
            />
            <span
              v-if="startWeekday"
              class="px4px py14px bg-#e6f0ff text-#06f rounded-10 text-14px ml-16px w-60px h-28px flex flex-center absolute right--18"
            >
              {{ startWeekday }}
            </span>
          </div>
        </a-form-item>

        <a-form-item label="结束时间" name="endTime" :rules="[{ required: true, message: '请选择结束时间' }]" class="w-80%">
          <div class="flex flex-center relative">
            <a-date-picker
              v-model:value="formState.endTime"
              :show-now="false"
              :show-time="{ format: 'HH:mm' }"
              format="YYYY-MM-DD HH:mm"
              placeholder="请选择结束时间"
              class="w-100%"
              :disabled-date="disabledEndDate"
            />
            <span
              v-if="endWeekday"
              class="px4px py14px bg-#e6f0ff text-#06f rounded-10 text-14px ml-16px w-60px h-28px flex flex-center absolute right--18"
            >
              {{ endWeekday }}
            </span>
          </div>

          <div v-if="canPreviewSchedules" class="bg-#fafafa rounded-6px p-10px mt-8px max-h-300px overflow-y-auto scrollbar">
            <div v-if="canPreviewSchedules && scheduleList.length" class="text-14px text-#888 mb-12px">
              请假期间有 {{ matchedLessonCount }} 个相关课节
            </div>
            <a-spin :spinning="scheduleLoading">
              <a-timeline v-if="canPreviewSchedules && scheduleList.length">
                <a-timeline-item v-for="item in scheduleList" :key="item.scheduleId">
                  <div class="text-14px text-#333 font500">
                    {{ formatDateTime(item.startTime) }} ~ {{ formatTimeOnly(item.endTime) }}
                  </div>
                  <div class="text-14px text-#333 mb-4px">
                    {{ formatScheduleTitle(item) }}
                  </div>
                  <div class="text-14px text-#888">
                    上课课程：{{ item.lessonName || '-' }}
                  </div>
                  <div class="text-14px text-#888">
                    上课教师：{{ item.teacherName || '-' }}
                  </div>
                </a-timeline-item>
                <a-timeline-item>
                  <div class="text-14px text-#888">
                    没有更多了～
                  </div>
                </a-timeline-item>
              </a-timeline>
              <div v-else-if="canPreviewSchedules" class="leave-empty-wrap">
                <a-empty :image="simpleImage" description="请假期间无相关课节" />
              </div>
            </a-spin>
          </div>
        </a-form-item>

        <a-form-item label="请假类型" name="leaveType" :rules="[{ required: true, message: '请选择请假类型' }]" class="w-80%">
          <a-select v-model:value="formState.leaveType" placeholder="请选择请假类型" :options="leaveTypeOptions" />
        </a-form-item>

        <a-form-item label="请假原因" name="reason" class="w-80%">
          <a-textarea
            v-model:value="formState.reason"
            placeholder="请输入请假原因"
            :auto-size="{ minRows: 3, maxRows: 3 }"
            :maxlength="500"
          />
        </a-form-item>

        <a-form-item label="请假佐证材料" name="proofMaterials" class="w-80%">
          <a-form-item-rest>
            <div class="mt--10px">
              <a-upload
                v-model:file-list="fileList"
                list-type="picture-card"
                accept=".jpg,.jpeg,.png,.bmp"
                :custom-request="handleProofUpload"
                :before-upload="beforeUpload"
                @change="handleUploadChange"
                @remove="handleRemove"
                @preview="handlePreview"
              >
                <div v-if="fileList.length < 3">
                  <PlusOutlined class="text-20px" />
                </div>
              </a-upload>
              <span class="text-#888 text-12px whitespace-nowrap">最多上传3张，支持BMP、JPG、JPEG、PNG，单张图片不超过 4 MB</span>
              <a-modal :open="previewVisible" :title="previewTitle" :footer="null" @cancel="handlePreviewCancel">
                <img alt="example" style="width: 100%" :src="previewImage">
              </a-modal>
            </div>
          </a-form-item-rest>
        </a-form-item>

        <a-form-item label="备注" name="remark" class="w-80%">
          <a-textarea
            v-model:value="formState.remark"
            placeholder="请输入备注（最多30字）"
            :auto-size="{ minRows: 1, maxRows: 1 }"
            :maxlength="30"
          />
        </a-form-item>
      </a-form>
    </div>

    <template #footer>
      <a-button danger ghost @click="closeFun">
        关闭
      </a-button>
      <a-button type="primary" ghost :loading="submitLoading" @click="handleSubmit">
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

.contenter {
  padding: 24px;
  height: calc(100vh - 200px);
  overflow-y: auto;
}

.leave-empty-wrap {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 110px;
}

:deep(.leave-empty-wrap .ant-empty) {
  margin-block: 0;
}

.custom-radio ::v-deep(.ant-radio-wrapper:hover .ant-radio),
.custom-radio ::v-deep(.ant-radio:hover .ant-radio-inner),
.custom-radio ::v-deep(.ant-radio-input:focus + .ant-radio-inner) {
  border-color: var(--pro-ant-color-primary);
}

.custom-radio ::v-deep(.ant-radio-inner) {
  background-color: transparent;
  border-color: #d9d9d9;
}

.custom-radio ::v-deep(.ant-radio-checked .ant-radio-inner) {
  background-color: transparent;
  border-color: var(--pro-ant-color-primary);
}

.custom-radio ::v-deep(.ant-radio-inner::after) {
  background-color: var(--pro-ant-color-primary);
  transform: scale(0.5);
}

.ant-timeline-item-last {
  padding-bottom: 0;
}

:deep(.ant-timeline-item-content) {
  &:last-child {
    min-height: 20px !important;
  }
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
