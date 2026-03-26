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
  endTheClassDate: '2025-04-14',
  remark: '',
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
    v-model:open="openModal" centered class="modal-content-box" :keyboard="false" :closable="false"
    :mask-closable="false" :width="800"
  >
    <template #title>
      <div class="text-5 flex justify-between flex-center">
        <span>结课</span>
        <a-button type="text" class="close-btn" @click="closeFun">
          <template #icon>
            <CloseOutlined class="text-5 close-icon" />
          </template>
        </a-button>
      </div>
    </template>
    <a-alert message="结课后，学员报读课程的剩余课时将全部扣除，机构获得相应课消收入。" show-icon type="warning" class="text-#f90 border-none rounded-0 bg-#fff5e6" />
    <div class="contenter scrollbar">
      <a-form ref="formRef" :model="formState" :label-col="{ span: 3 }" :wrapper-col="{ span: 20 }">
        <div class="text-20px font-800 mb-4px">
          视知觉训练
        </div>
        <a-space>
          <span class="bg-#e6f0ff text-#06f text-3 px3 py2px rounded-10 ">班级授课</span>
          <span class="bg-#e6f0ff text-#06f text-3 px3 py2px rounded-10 ">课时</span>
        </a-space>
        <a-descriptions class="mt-20px" :column="2" size="small" :content-style="{ color: '#888' }">
          <a-descriptions-item label="剩余课时">
            21课时
          </a-descriptions-item>
          <a-descriptions-item label="剩余学费">
            ¥ 2300.00
          </a-descriptions-item>
          <a-descriptions-item label="有效期至">
            2025-04-14
          </a-descriptions-item>
        </a-descriptions>
        <a-divider class="my-16px" />
        <a-form-item label="结课日期" name="endTheClassDate">
          <span class="text-#ff3333">{{ formState.endTheClassDate }}（今天）</span>
        </a-form-item>
        <!-- 结课备注 -->
        <a-form-item label="结课备注">
          <a-input v-model:value="formState.remark" placeholder="选填" />
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
  margin: 24px;
  border-radius: 14px;
  background: #fafafa;
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
