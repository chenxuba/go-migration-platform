<!-- 删除部门 -->
<script setup>
import { CloseCircleOutlined } from '@ant-design/icons-vue'
import { deleteDepart } from '@/api/internal-manage/staff-manage'

const props = defineProps({
  params: {
    type: Object,
    required: true,
    default: () => ({}),
  },
})
const emit = defineEmits(['success'])

const open = defineModel({
  type: Boolean,
  default: false,
})

const loading = ref(false)

function handleOk() {
  const { id, uuid, version } = props.params
  loading.value = true
  deleteDepart({ id, uuid, version }).then(() => {
    emit('success')
    open.value = false
  }).finally(() => {
    loading.value = false
  })
}
</script>

<template>
  <div>
    <a-modal v-model:open="open" centered :closable="false" :mask-closable="false" :keyboard="false" width="400px">
      <div class="flex items-start gap-15px">
        <CloseCircleOutlined class="text-red-500 text-20px" />
        <div class="flex flex-col gap-20px">
          <span class="font-medium">确定删除该部门吗？</span>
          <span class="text-gray-400 text-13px">删除后不可恢复</span>
        </div>
      </div>
      <template #footer>
        <a-button :loading="loading" danger ghost @click="handleOk">
          确定
        </a-button>
        <a-button type="primary" ghost @click="open = false">
          取消
        </a-button>
      </template>
    </a-modal>
  </div>
</template>

<style scoped lang="less"></style>
