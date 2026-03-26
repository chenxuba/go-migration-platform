<!-- 新增部门 -->
<script setup>
import { saveDepartApi } from '@/api/internal-manage/staff-manage'

const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
  params: {
    type: Object,
    required: true,
    default: () => ({}),
  },
})
const emit = defineEmits(['update:open', 'success'])

// 处理双向绑定
const openModal = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})
const form = ref(null)

const formState = reactive({
  departName: '',
})

const loading = ref(false)

async function handleOk() {
  try {
    const res = await form.value.validate()
    if (res) {
      loading.value = true
      try {
        // 在这里调用新增部门接口
        const res = await saveDepartApi({ ...formState, ...props.params })
        if (res.code === 200) {
          emit('success')
          cancel()
        }
      }
      finally {
        openModal.value = false
        loading.value = false
      }
    }
  }
  catch (err) {
    console.log(err)
  }
}
const formLabel = computed(() => {
  return props.params.data.pname ? `${props.params.data.pname}/${props.params.data.departName}` : props.params.data.departName
})
function cancel() {
  openModal.value = false
  form.value.resetFields()
}
</script>

<template>
  <div>
    <a-modal v-model:open="openModal" title="新增部门" :mask-closable="false" :keyboard="false" :centered="true">
      <a-form ref="form" :model="formState" layout="vertical" autocomplete="off">
        <a-form-item :label="formLabel" name="departName" :rules="[{ required: true, message: '请输入部门名称' }]">
          <a-input v-model:value="formState.departName" style="width:100%;" placeholder="请输入（15字以内）" />
        </a-form-item>
      </a-form>
      <template #footer>
        <div>
          <a-button @click="cancel">
            取消
          </a-button>
          <a-button :loading="loading" type="primary" @click="handleOk">
            确定
          </a-button>
        </div>
      </template>
    </a-modal>
  </div>
</template>

<style scoped lang="less">
:deep(.ant-form-item-explain-error) {
  font-size: 12px;
}
</style>
