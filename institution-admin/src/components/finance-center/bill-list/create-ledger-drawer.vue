<script setup>
import { CloseOutlined } from '@ant-design/icons-vue'
import dayjs from 'dayjs'
import { reactive } from 'vue'
import StaffSelect from '@/components/common/staff-select.vue'
import { useDrawer } from '@/composables/useDrawer'

const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
})

const emit = defineEmits(['update:open'])

const { openDrawer } = useDrawer(props, emit)

const formState = reactive({
  amount: undefined,
  payDate: dayjs().format('YYYY-MM-DD'),
  dealStaffId: undefined,
})

function disabledDate(current) {
  return current > dayjs().endOf('day')
}
</script>

<template>
  <a-drawer
    v-model:open="openDrawer"
    :push="{ distance: 80 }"
    :body-style="{ padding: '0', background: '#f7f7fd' }"
    :closable="false"
    width="800px"
    placement="right"
  >
    <template #title>
      <div class="custom-header flex justify-between h-4 flex-items-center">
        <div class="text-5">
          记一笔
        </div>
        <a-button type="text" class="close-btn" @click="openDrawer = false">
          <template #icon>
            <CloseOutlined class="text-5 close-icon" />
          </template>
        </a-button>
      </div>
    </template>
    <div class="contenter">
      <div class="form-panel">
        <a-form layout="vertical" :model="formState" class="drawer-form">
          <div class="middleBox mt2 py6 pb2 px8 bg-white">
            <div>
              <div class="lebal text-#666">
                <span class="text-#f03 mr1" style="font-family: SimSun, sans-serif;">*</span>账单金额：
              </div>
              <div class="payPrice h-20 border border-b-#eee border-solid border-x-none border-t-none">
                <a-input-number
                  v-model:value="formState.amount"
                  :bordered="false"
                  :controls="false"
                  class="h-100% w-100% text-12"
                  :min="0"
                  :max="100000"
                  :precision="2"
                  placeholder="输入金额"
                >
                  <template #addonBefore>
                    <span class="text-12">¥</span>
                  </template>
                </a-input-number>
              </div>
            </div>
            <div class="mt6">
              <div class="flex justify-between w-75%">
                <a-form-item class="drawer-form-item">
                  <template #label>
                    <span><span class="text-#f03 mr1" style="font-family: SimSun, sans-serif;">*</span>支付日期：</span>
                  </template>
                  <a-date-picker
                    v-model:value="formState.payDate"
                    class="w-60"
                    :allow-clear="false"
                    :disabled-date="disabledDate"
                    format="YYYY-MM-DD"
                    value-format="YYYY-MM-DD"
                  />
                </a-form-item>
                <a-form-item class="drawer-form-item">
                  <template #label>
                    <span><span class="text-#f03 mr1" style="font-family: SimSun, sans-serif;">*</span>经办人：</span>
                  </template>
                  <StaffSelect
                    v-model="formState.dealStaffId"
                    placeholder="请选择经办人"
                    width="240px"
                  />
                </a-form-item>
              </div>
            </div>
          </div>
        </a-form>
      </div>
    </div>
  </a-drawer>
</template>

<style lang="less" scoped>
@keyframes icon-rotate {
  from {
    transform: rotate(0deg);
  }

  to {
    transform: rotate(180deg);
  }
}

.contenter {
  background: #fff;
}



.middleBox {
  max-width: 760px;
}

.close-btn {
  &:hover {
    background: transparent;

    .close-icon {
      animation: icon-rotate 0.3s linear;
    }
  }
}

.payPrice {
  :deep(.ant-input-number) {
    width: 100%;
  }

  :deep(.ant-input-number-input) {
    height: 80px;
    line-height: 80px;
    font-family: "DIN alternate", sans-serif;
  }

  :deep(.ant-input-number-group-addon) {
    background: transparent;
    border: none;
  }
}

@media (max-width: 960px) {
  .form-panel {
    padding: 24px;
  }

  .middleBox {
    max-width: 100%;
  }
}
</style>
