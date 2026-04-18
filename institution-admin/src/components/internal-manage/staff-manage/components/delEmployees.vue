<script setup>
import { batchDisabledApi } from '@/api/internal-manage/staff-manage'

const { detail, batchUserList } = defineProps({
  detail: {
    type: Object,
    default: () => {
      return {}
    },
  },
  batchUserList: {
    type: Array,
    default: () => {
      return []
    },
  },
})

const emit = defineEmits(['success'])

const open = defineModel({
  type: Boolean,
  default: false,
})

const resign = ref(false)

const isBatch = computed(() => {
  return batchUserList.length > 0
})

function onCancel() {
  open.value = false
  resign.value = false
}

function batchDisabledApiUser() {
  const userIds = isBatch.value
    ? batchUserList.map((item) => { return item.id })
    : [detail.id]
  batchDisabledApi({
    userIds,
    isWork: true,
  }).then(() => {
    emit('success')
    onCancel()
  })
}
</script>

<template>
  <div>
    <a-modal
      v-model:open="open" :keyboard="false" :mask-closable="false" centered :closable="false" width="500px"
      :after-close="onCancel" :footer="null" :destroy-on-close="true"
    >
      <div class="ant-modal-confirm-body-wrapper">
        <div>
          <div class="flex items-start gap-15px">
            <span role="img" aria-label="exclamation-circle" class="text-#f90  text-18px  text-center">
              <svg
                viewBox="64 64 896 896" focusable="false" data-icon="exclamation-circle" width="1em" height="1em"
                fill="currentColor" aria-hidden="true"
              >
                <path
                  d="M512 64C264.6 64 64 264.6 64 512s200.6 448 448 448 448-200.6 448-448S759.4 64 512 64zm0 820c-205.4 0-372-166.6-372-372s166.6-372 372-372 372 166.6 372 372-166.6 372-372 372z"
                />
                <path
                  d="M464 688a48 48 0 1096 0 48 48 0 10-96 0zm24-112h48c4.4 0 8-3.6 8-8V296c0-4.4-3.6-8-8-8h-48c-4.4 0-8 3.6-8 8v272c0 4.4 3.6 8 8 8z"
                />
              </svg>
            </span>
            <div v-if="isBatch" class="flex flex-col items-start gap-3px justify-start">
              <span class="ant-modal-confirm-title">确定批量离职 {{ batchUserList.length }} 为员工嘛？</span>
              <span>员工姓名：{{ batchUserList.map(item => { return item.nickName }).join('、') }}</span>
            </div>
            <span v-else class="ant-modal-confirm-title">确定要将“{{ detail.nickName }}”设置为离职吗？</span>
          </div>
          <div class="ml-38px mt-8px text-#666 text-14px">
            <div>
              1、离职后，该账户将无法登录系统;<br>2、该员工的离职，不会影响历史已生成的相关工作记录，也不会影响未来已安排的工作，如需更改负责老师，请去相关日程进行手动更改。<div
                class="resignCheckBox"
              >
                <a-checkbox v-model:checked="resign">
                  我已阅读并知晓以上风险
                </a-checkbox>
              </div>
              <div class="flex justify-end mt-24px gap-10px">
                <a-button @click="onCancel">
                  取消
                </a-button>
                <a-button type="primary" :disabled="!resign" @click="batchDisabledApiUser">
                  确认离职
                </a-button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </a-modal>
  </div>
</template>

<style scoped lang='less'>
.ant-modal-confirm-title {
  color: #222;
  font-size: 16px;
  font-weight: 400;
  // line-height: 24px;
  overflow: hidden;
}

.resignCheckBox {
  height: 24px;
  margin-top: 16px;
  line-height: 24px;
  font-size: 16px;
  font-weight: 500;
  color: #222;
}
</style>
