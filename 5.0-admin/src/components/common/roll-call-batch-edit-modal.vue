<script setup>
import { CloseOutlined, InfoCircleFilled } from '@ant-design/icons-vue'

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
const formState = ref({
  classNumber: 1,
  editRange: '1',
})
const editRangeOptions = [
  { label: '所有上课状态的学员', value: '1' },
  { label: '仅对到课状态的学员', value: '2' },
  { label: '仅对请假状态的学员', value: '3' },
  { label: '仅对旷课状态的学员', value: '4' },
]
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
    v-model:open="openModal" centered class="modal-content-box" :keyboard="false" :closable="false"
    :mask-closable="false" :width="800"
  >
    <template #title>
      <div class="text-5 flex justify-between flex-center">
        <span>批量编辑</span>
        <a-button type="text" class="close-btn" @click="closeFun">
          <template #icon>
            <CloseOutlined class="text-5 close-icon" />
          </template>
        </a-button>
      </div>
    </template>
    <a-alert type="info" show-icon class="border-none rounded-0 text-#06f text-14px" message="批量编辑不包含补课学员、试听学员">
      <template #icon>
        <InfoCircleFilled />
      </template>
    </a-alert>
    <div class="contenter ">
      <div class="bg-#fafafa rounded-12px p-24px pb-1px">
        <a-form ref="formRef" :model="formState" :label-col="{ span: 4 }" :wrapper-col="{ span: 18 }">
          <!-- 上课点名数量 -->
          <a-form-item label="上课点名数量" name="classNumber" :rules="[{ required: true, message: '请输入上课点名数量' }]">
            <a-input-number v-model:value="formState.classNumber" :min="1" :max="100" :precision="2" placeholder="请输入" class="w-180px" />
          </a-form-item>
          <!-- 编辑生效范围 -->
          <a-form-item label="编辑生效范围" name="editRange">
            <a-select v-model:value="formState.editRange" placeholder="请选择" style="width: 180px;">
              <a-select-option v-for="item in editRangeOptions" :key="item.value" :value="item.value">
                {{ item.label }}
              </a-select-option>
            </a-select>
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
  padding: 16px 24px 24px;
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
