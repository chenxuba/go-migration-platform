<script setup>
import { CloseOutlined } from '@ant-design/icons-vue'

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

})
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
  openModal.value = false
}
</script>

<template>
  <a-modal
    v-model:open="openModal" style="top:12px" class="modal-content-box" :keyboard="false" :closable="false"
    :mask-closable="false" :width="800"
  >
    <template #title>
      <div class="text-5 flex justify-between flex-center">
        <span>添加跟进记录</span>
        <a-button type="text" class="close-btn" @click="closeFun">
          <template #icon>
            <CloseOutlined class="text-5 close-icon" />
          </template>
        </a-button>
      </div>
    </template>
    <div class="contenter scrollbar">
      <a-form ref="formRef" layout="vertical" :model="formState">
        11
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
