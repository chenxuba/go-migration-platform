<script setup>
import { computed, reactive, ref, watch } from 'vue'
import { CloseOutlined, PlusOutlined } from '@ant-design/icons-vue'
import dayjs from 'dayjs'
import StudentSelect from './student-select.vue'

const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
})
const emit = defineEmits(['update:open'])
const formRef = ref()
const studentSelectRef = ref()
// 处理双向绑定
const openModal = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})
function getBase64(file) {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.readAsDataURL(file)
    reader.onload = () => resolve(reader.result)
    reader.onerror = error => reject(error)
  })
}
const previewVisible = ref(false)
const previewImage = ref('')
const previewTitle = ref('')
const fileList = ref([
  {
    uid: '-1',
    name: 'image.png',
    status: 'done',
    url: 'https://zos.alipayobjects.com/rmsportal/jkjgkEfvpUPVyRjUImniVslZfWPnJuuZ.png',
  },
  {
    uid: '-2',
    name: 'image.png',
    status: 'done',
    url: 'https://zos.alipayobjects.com/rmsportal/jkjgkEfvpUPVyRjUImniVslZfWPnJuuZ.png',
  },
  {
    uid: '-3',
    name: 'image.png',
    status: 'done',
    url: 'https://zos.alipayobjects.com/rmsportal/jkjgkEfvpUPVyRjUImniVslZfWPnJuuZ.png',
  },
])
function handleCancel() {
  previewVisible.value = false
  previewTitle.value = ''
}
async function handlePreview(file) {
  if (!file.url && !file.preview) {
    file.preview = await getBase64(file.originFileObj)
  }
  previewImage.value = file.url || file.preview
  previewVisible.value = true
  previewTitle.value = file.name || file.url.substring(file.url.lastIndexOf('/') + 1)
}

const formState = reactive({
  studentId: undefined,
  startTime: undefined,
  endTime: undefined,
  leaveType: undefined,
  leaveReason: undefined,
  leaveProof: [],
  remark: undefined,
})

// 计算开始时间的周几
const startWeekday = computed(() => {
  if (!formState.startTime)
    return ''
  const weekdays = ['周日', '周一', '周二', '周三', '周四', '周五', '周六']
  return weekdays[formState.startTime.day()]
})

// 计算结束时间的周几
const endWeekday = computed(() => {
  if (!formState.endTime)
    return ''
  const weekdays = ['周日', '周一', '周二', '周三', '周四', '周五', '周六']
  return weekdays[formState.endTime.day()]
})

// 禁用开始时间之前的时间
function disabledStartDate(current) {
  return current && current < dayjs().startOf('day')
}

// 禁用结束时间之前的时间
function disabledEndDate(current) {
  if (!formState.startTime) {
    return current && current < dayjs().startOf('day')
  }
  return current && current < formState.startTime
}

// 开始时间变化时的处理
function handleStartTimeChange(date) {
  if (formState.endTime && date > formState.endTime) {
    formState.endTime = undefined
  }
}

// 监听模态框打开状态
watch(() => props.open, (newVal) => {
  if (newVal && studentSelectRef.value) {
    // 弹窗打开时重置学员选择组件，确保获取最新数据
    studentSelectRef.value.reset()
  }
})

// 处理学员选择
function handleStudentSelect(student) {
  console.log('选择的学员:', student)
  // 可以在这里处理选择学员后的逻辑，比如清空之前的请假记录等
}

// 手动触发验证
async function handleSubmit() {
  try {
    await formRef.value.validate() // 关键3：通过引用调用验证方法
    console.log('验证通过，提交数据:', formState)
  }
  catch (error) {
    console.log('验证失败:', error)
  }
}
function closeFun() {
  formRef.value.resetFields()
  // 重置学员选择组件，确保重新打开弹窗时重新请求接口
  if (studentSelectRef.value) {
    studentSelectRef.value.reset()
  }
  openModal.value = false
}
</script>

<template>
  <a-modal
    v-model:open="openModal" style="top: 40px;" class="modal-content-box" :keyboard="false" :closable="false" :mask-closable="false"
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
        <!-- 选择学员 必选 -->
        <a-form-item label="选择学员" name="studentId" :rules="[{ required: true, message: '请选择学员' }]" class="w-80%">
          <StudentSelect 
            ref="studentSelectRef"
            v-model="formState.studentId" 
            placeholder="请选择学员"
            allow-clear
            @select="handleStudentSelect"
          />
        </a-form-item>
        <!-- 开始时间 必选 -->
        <a-form-item label="开始时间" name="startTime" :rules="[{ required: true, message: '请选择开始时间' }]" class="w-80%">
          <div class="flex flex-center relative">
            <a-date-picker
              v-model:value="formState.startTime" :show-time="{ format: 'HH:mm' }"
              format="YYYY-MM-DD HH:mm" placeholder="请选择开始时间" class="w-100%"
              :disabled-date="disabledStartDate" @change="handleStartTimeChange"
            />
            <span
              v-if="startWeekday"
              class="px4px py14px bg-#e6f0ff text-#06f rounded-10 text-14px ml-16px w-60px h-28px flex flex-center absolute right--18"
            >
              {{ startWeekday }}
            </span>
          </div>
        </a-form-item>
        <!-- 结束时间 必选 -->
        <a-form-item label="结束时间" name="endTime" :rules="[{ required: true, message: '请选择结束时间' }]" class="w-80%">
          <div class="flex flex-center relative">
            <a-date-picker
              v-model:value="formState.endTime" :show-now="false" :show-time="{ format: 'HH:mm' }"
              format="YYYY-MM-DD HH:mm" placeholder="请选择结束时间" class="w-100%" :disabled-date="disabledEndDate"
            />
            <span
              v-if="endWeekday"
              class="px4px py14px bg-#e6f0ff text-#06f rounded-10 text-14px ml-16px w-60px h-28px flex flex-center absolute right--18"
            >
              {{ endWeekday }}
            </span>
          </div>
          <div class="bg-#fafafa rounded-6px p-10px mt-8px max-h-300px overflow-y-auto scrollbar">
            <div class="text-14px text-#888 mb-12px">
              请假期间有 13 个相关课节
            </div>
            <a-timeline>
              <a-timeline-item>
                <div class="text-14px text-#333 font500">
                  2025-05-01 11:00 ~ 11:45
                </div>
                <div class="text-14px text-#333 mb-4px">
                  浩楠-高级言语课
                </div>
                <div class="text-14px text-#888">
                  上课课程：高级言语课
                </div>
                <div class="text-14px text-#888">
                  上课教师：何红武
                </div>
              </a-timeline-item>
              <a-timeline-item>
                <div class="text-14px text-#333 font500">
                  2025-05-01 11:00 ~ 11:45
                </div>
                <div class="text-14px text-#333 mb-4px">
                  浩楠-高级言语课
                </div>
                <div class="text-14px text-#888">
                  上课课程：高级言语课
                </div>
                <div class="text-14px text-#888">
                  上课教师：何红武
                </div>
              </a-timeline-item>
              <a-timeline-item>
                <!-- 没有更多了～ -->
                <div class="text-14px text-#888">
                  没有更多了～
                </div>
              </a-timeline-item>
            </a-timeline>
          </div>
        </a-form-item>
        <!-- 请假类型 事假  病假  休学 必选 -->
        <a-form-item label="请假类型" name="leaveType" :rules="[{ required: true, message: '请选择请假类型' }]" class="w-80%">
          <a-select v-model:value="formState.leaveType" placeholder="请选择请假类型">
            <a-select-option value="1">
              事假
            </a-select-option>
            <a-select-option value="2">
              病假
            </a-select-option>
            <a-select-option value="3">
              休学
            </a-select-option>
          </a-select>
        </a-form-item>
        <!-- 请假原因 非必选 -->
        <a-form-item label="请假原因" name="leaveReason" class="w-80%">
          <a-textarea
            v-model:value="formState.leaveReason" placeholder="请输入请假原因"
            :auto-size="{ minRows: 3, maxRows: 3 }"
          />
        </a-form-item>
        <!-- 请假佐证材料 -->
        <a-form-item label="请假佐证材料" name="leaveProof" class="w-80%">
          <a-form-item-rest>
            <div class="mt--10px">
              <a-upload
                v-model:file-list="fileList" action="https://www.mocky.io/v2/5cc8019d300000980a055e76"
                list-type="picture-card" accept=".jpg,.jpeg,.png"
                @preview="handlePreview"
              >
                <div v-if="fileList.length < 3">
                  <PlusOutlined class="text-20px" />
                </div>
              </a-upload>
              <span class="text-#888 text-12px">最多上传3张，支持JPG、JPEG、PNG，单张图片不超过 4 MB</span>
              <a-modal :open="previewVisible" :title="previewTitle" :footer="null" @cancel="handleCancel">
                <img alt="example" style="width: 100%" :src="previewImage">
              </a-modal>
            </div>
          </a-form-item-rest>
        </a-form-item>
        <!-- 备注 必选 -->
        <a-form-item label="备注" name="remark" class="w-80%">
          <!-- 最多30字 -->
          <a-textarea
            v-model:value="formState.remark" placeholder="请输入备注（最多30字）"
            :auto-size="{ minRows: 1, maxRows: 1 }" :maxlength="30"
          />
        </a-form-item>
      </a-form>
    </div>
    <template #footer>
      <a-button danger ghost @click="closeFun">
        关闭
      </a-button>
      <a-button type="primary" ghost @click="handleSubmit">
        确定
      </a-button>
    </template>
  </a-modal>
</template>

<style lang="less" scoped>
/* 添加旋转动画 */
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

/* 自定义镂空样式 */
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
.ant-timeline-item-last{
  padding-bottom: 0;
}
:deep(.ant-timeline-item-content){
  &:last-child{
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
