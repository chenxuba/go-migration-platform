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
  status: 1,
  editRecord: 1,
  internalNote: undefined,
  externalNote: undefined,
})
// 手动触发验证
async function handleSubmit() {
  try {
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
        <span>编辑点名</span>
        <a-button type="text" class="close-btn" @click="closeFun">
          <template #icon>
            <CloseOutlined class="text-5 close-icon" />
          </template>
        </a-button>
      </div>
    </template>
    <div class="rounded-12px bg-#fafafa mx-24px">
      <div class="contenter   px6 py3" style="margin-bottom: 0;">
        <div class="avatar flex flex-items-center">
          <img
            src="https://cdn.schoolpal.cn/schoolpal/next-erp/avator_male.png"
            className="w-40px h-40px rounded-full mr-8px" alt=""
          >
          <span class="text-5 text-#222 font-800">张三</span>
        </div>
        <div class="text-14px text-#222 flex mt-12px flex-wrap">
          <div class="mr-60px mb-10px flex flex-items-center">
            <span class="text-#888">学员身份：</span>
            <span class="flex flex-items-center">试听学员
              <span class="mt-1px ml4px bg-#fff5e6 text-#f90  text-10px px2 py2px rounded-10">免费试听</span>
            </span>
          </div>
          <div class="mr-60px">
            <span class="text-#888">课消方式：</span>
            <span>按课时</span>
          </div>
          <div class="mr-60px">
            <span class="text-#888">扣费课程账户：</span>
            <span>视知觉训练</span>
          </div>
          <div class="mr-60px">
            <span class="text-#888">剩余课时：</span>
            <span>20课时</span>
          </div>
        </div>
      </div>
      <div class="w-752px">
        <a-divider class="my-0" />
      </div>
      <div class="contenter scrollbar" style="margin: 0 24px 24px 24px;">
        <a-form ref="formRef" :model="formState" :label-col="{ span: 4 }" :wrapper-col="{ span: 18 }">
          <!-- 编辑状态 必选- 单选框 -->
          <a-form-item label="编辑状态" name="status" :rules="[{ required: true, message: '请选择编辑状态' }]" class="mb-40px">
            <div class="flex flex-col relative">
              <a-radio-group v-model:value="formState.status" class="custom-radio">
                <a-radio :value="1">
                  到课
                </a-radio>
                <a-radio :value="2">
                  请假
                </a-radio>
                <a-radio :value="3">
                  旷课
                </a-radio>
              </a-radio-group>
              <span class="text-14px text-#888 absolute bottom--22px">免费试听学员，不支持记录课时</span>
            </div>
          </a-form-item>
          <!-- 编辑记录 -->
          <a-form-item label="编辑记录" name="editRecord">
            <div class="flex flex-items-center">
              <a-input-number v-model:value="formState.editRecord" :precision="2" placeholder="选填（200字以内）" :min="0" :max="100" /> <span class="ml-4px">课时</span>
            </div>
          </a-form-item>
          <!-- 编辑对内备注 -->
          <a-form-item label="编辑对内备注" name="internalNote">
            <a-input v-model:value="formState.internalNote" placeholder="选填（200字以内）" :maxlength="200" />
          </a-form-item>
          <!-- 编辑对外备注 -->
          <a-form-item label="编辑对外备注" name="externalNote">
            <a-input v-model:value="formState.externalNote" placeholder="选填（200字以内）" :maxlength="200" />
          </a-form-item>
        </a-form>
      </div>
    </div>
    <template #footer>
      <a-button danger ghost @click="closeFun">
        取消
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
