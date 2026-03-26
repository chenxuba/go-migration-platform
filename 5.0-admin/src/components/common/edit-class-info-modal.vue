<script setup>
import { CloseOutlined } from '@ant-design/icons-vue'
import StaffSelect from './staff-select.vue'
import messageService from '~@/utils/messageService'

const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
})
const emit = defineEmits(['update:open'])
const formRef = ref()
// 处理双向绑定
const openModal = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})
const formState = reactive({
  teacher: undefined,
  assistant: [],
  teacherRecordHour: 1,
})
// 手动触发验证
async function handleSubmit() {
  try {
    // 新增：检测上课助教是否包含上课教师
    if (formState.assistant && Array.isArray(formState.assistant) && formState.assistant.includes(formState.teacher)) {
      messageService.error('上课教师与上课助教不可为同一个人，请重新选择')
      return
    }
    await formRef.value.validate() // 关键3：通过引用调用验证方法
    console.log('验证通过，提交数据:', formState)
    // 关闭modal
    openModal.value = false
  }
  catch (error) {
    console.log('验证失败:', error)
  }
}
function closeFun() {
  formRef.value.resetFields()
  openModal.value = false
}
</script>

<template>
  <a-modal
    v-model:open="openModal" centered class="modal-content-box" :keyboard="false" :closable="false"
    :mask-closable="false" :width="800"
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
    <div class="rounded-12px bg-#fafafa mx-24px">
      <div class="contenter flex flex-center bg-white px6 py3" style="margin-bottom: 0;">
        <div class="avatarBox w-16 h-16 relative">
          <img
            width="64" height="64" src="https://pcsys.admin.ybc365.com/83b8fd68-2f9b-4a35-979f-1fd0ea349889.png"
            alt=""
          >
        </div>
        <div class="info flex flex-1 ml-4 flex-col ">
          <div class="top flex justify-between flex-center flex-1">
            <a-space>
              <div class="name text-5 font-800">
                陈陈-初级感统课
              </div>
            </a-space>
          </div>
          <div class="bottom flex-1 flex flex-items-center mt-2">
            <div class="birthday flex-center">
              <span class="text-4 text-#222">2025-04-14(周一)10:00 ~ 10:30</span>
              <span class="bg-#e6f0ff text-#06f text-3 px2 py1 rounded-10 ml2">30分钟</span>
            </div>
          </div>
        </div>
      </div>
      <div class="w-752px">
        <a-divider class="my-0" />
      </div>
      <div class="contenter scrollbar" style="margin: 0 24px 24px 24px;">
        <a-form ref="formRef" :model="formState" :label-col="{ span: 5 }" :wrapper-col="{ span: 18 }">
          <!-- 上课教师 必选 -->
          <a-form-item label="上课教师" name="teacher" :rules="[{ required: true, message: '请选择上课教师' }]">
            <StaffSelect
              v-model="formState.teacher"
              placeholder="请选择上课教师"
              :width="'240px'"
              :multiple="false"
              :status="0"
            />
          </a-form-item>
          <!-- 上课助教 非必选 支持多选 -->
          <a-form-item label="上课助教" name="assistant">
            <StaffSelect
              v-model="formState.assistant"
              placeholder="请选择上课助教"
              :width="'100%'"
              :multiple="true"
              :status="0"
            />
          </a-form-item>

          <!-- 教师记录课时  非必填 -->
          <a-form-item label="教师记录课时" name="teacherRecordHour">
            <div class="flex flex-items-center">
              <a-input-number
                v-model:value="formState.teacherRecordHour"
                placeholder="请输入" :precision="2" :min="0" :max="100" class="mr-8px"
              /> 课时
            </div>
          </a-form-item>
        </a-form>
      </div>
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
  background: #fafafa;
  margin: 24px;

  .multiple-select {
    :deep(.ant-select-selection-item) {
      background-color: #e6f0ff;
      border: 1px solid #99c2ff;
    }
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
