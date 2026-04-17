<script setup>
import { CloseOutlined } from '@ant-design/icons-vue'
import dayjs from 'dayjs'
const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
  config: {
    type: Object,
    default: () => ({}),
  },
})
const emit = defineEmits(['update:open', 'submit'])
const formRef = ref()
// 处理双向绑定
const openModal = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})
const formState = reactive({
  validDate: "",
  endDate: "",
  freeQuantity: 0,
})
// 手动触发验证
async function handleSubmit() {
  try {
    await formRef.value.validate() // 关键3：通过引用调用验证方法
    console.log('验证通过，提交数据:', formState)
    emit('submit', formState)
  }
  catch (error) {
    console.log('验证失败:', error)
  }
}
function closeFun() {
  formRef.value.resetFields()
  openModal.value = false
}
watch(() => props.config, (newVal) => {
  formState.validDate = newVal.validDate
  formState.endDate = newVal.endDate
  formState.beforeTotalDays = newVal.endDate
  formState.freeQuantity = newVal.freeQuantity
}, { immediate: true })
function disabledDate(current) {
  return current && current < dayjs(formState?.validDate).add(formState?.freeQuantity, 'day')
}
const totalDays = computed(() => {
  return dayjs(formState?.endDate).diff(dayjs(formState?.validDate), 'day') + 1
})
const beforeTotalDays = computed(() => {
  return dayjs(formState?.beforeTotalDays).diff(dayjs(formState?.validDate), 'day') + 1
})
</script>

<template>
  <a-modal v-model:open="openModal" centered class="modal-content-box" :keyboard="false" :closable="false"
    :mask-closable="false" :width="800">
    <template #title>
      <div class="text-5 flex justify-between flex-center">
        <span>修改结束时间</span>
        <a-button type="text" class="close-btn" @click="closeFun">
          <template #icon>
            <CloseOutlined class="text-5 close-icon" />
          </template>
        </a-button>
      </div>
    </template>
    <div class="contenter scrollbar">
      <div class="box">
        <a-form ref="formRef" :model="formState">
          <a-form-item class="mb-0">
            <span class="text-#888">系统根据您的报价单和购买份数自动计算出购买总天数（含赠）： <span class="text-#06f">{{ totalDays }}</span>
              天</span>
          </a-form-item>
          <a-form-item class="mt-12px" label="结束时间" name="beforeTotalDays"
            :rules="[{ required: true, message: '请选择结束时间' }]">
            <div class="flex flex-center">
              <a-date-picker v-model:value="formState.beforeTotalDays" :disabled-date="disabledDate" format="YYYY-MM-DD"
                value-format="YYYY-MM-DD" style="width: 200px" />
              <span class="ml-12px text-#888">调整后总天数（含赠）：<span class="text-#06f">{{ beforeTotalDays || totalDays }}</span>
                天</span>
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

  .box {
    background: #fafafa;
    border-radius: 16px;
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 24px 0 12px 0;
  }
}
</style>
