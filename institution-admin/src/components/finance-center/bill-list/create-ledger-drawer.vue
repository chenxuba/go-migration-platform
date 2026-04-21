<script setup>
import { CheckOutlined, CloseOutlined } from "@ant-design/icons-vue";
import dayjs from "dayjs";
import { reactive } from "vue";
import StaffSelect from "@/components/common/staff-select.vue";
import { useDrawer } from "@/composables/useDrawer";

const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
});

const emit = defineEmits(["update:open"]);

const { openDrawer } = useDrawer(props, emit);

const formState = reactive({
  amount: undefined,
  ledgerType: 1,
  payDate: dayjs().format("YYYY-MM-DD"),
  dealStaffId: undefined,
});

function disabledDate(current) {
  return current > dayjs().endOf("day");
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
        <div class="text-5">记一笔</div>
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
          <div class="middleBox pb2 bg-white">
            <div>
              <div class="lebal text-#666">
                <span
                  class="text-#f03 mr1"
                  style="font-family: SimSun, sans-serif"
                  >*</span
                >账单金额：
              </div>
              <div
                class="payPrice h-20 border border-b-#eee border-solid border-x-none border-t-none"
              >
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
                    <span
                      ><span
                        class="text-#f03 mr1"
                        style="font-family: SimSun, sans-serif"
                        >*</span
                      >支付日期：</span
                    >
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
                    <span
                      ><span
                        class="text-#f03 mr1"
                        style="font-family: SimSun, sans-serif"
                        >*</span
                      >经办人：</span
                    >
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
    <div class="p-24px">
  <!-- 收入、支出 -->
    <!-- a-radio-group -->
    <a-radio-group
      v-model:value="formState.ledgerType"
      button-style="solid"
      class="ledger-type-group"
    >
      <a-radio-button
        :value="1"
        class="ledger-type-option ledger-type-option--income w-175px text-center font-size-16px"
      >
        <span class="ledger-type-label">
          <span
            v-if="formState.ledgerType === 1"
            class="ledger-type-icon ledger-type-icon--income"
          >
            <CheckOutlined />
          </span>
          <span>收入</span>
        </span>
      </a-radio-button>
      <a-radio-button
        :value="2"
        class="ledger-type-option ledger-type-option--expense w-175px text-center font-size-16px"
      >
        <span class="ledger-type-label">
          <span
            v-if="formState.ledgerType === 2"
            class="ledger-type-icon ledger-type-icon--expense"
          >
            <CheckOutlined />
          </span>
          <span>支出</span>
        </span>
      </a-radio-button>
    </a-radio-group>
    </div>
  
    <template #footer>
      <div class="flex justify-end">
        <a-button
          type="primary"
          class="w-140px h-48px font-size-18px font-weight-600"
          @click="openDrawer = false"
        >
          完成
        </a-button>
      </div>
    </template>
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
  padding: 24px 24px 0 24px;
  .form-panel {
    background: #fff;
  }
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

.ledger-type-group {
  :deep(.ant-radio-button-wrapper) {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    padding: 0 20px;
    border-color: #d9d9d9;
    color: #222;
    box-shadow: none;

    &::before {
      display: none;
    }

    &:hover {
      color: #222;
    }
  }

  :deep(
      .ant-radio-button-wrapper-checked:not(.ant-radio-button-wrapper-disabled)
    ) {
    background: #1677ff;
    border-color: #1677ff;
    color: #fff;

    &:hover {
      background: #1677ff;
      border-color: #1677ff;
      color: #fff;
    }
  }

  :deep(
      .ledger-type-option--expense.ant-radio-button-wrapper-checked:not(
          .ant-radio-button-wrapper-disabled
        )
    ) {
    background: #f90;
    border-color: #f90;

    &:hover {
      background: #f90;
      border-color: #f90;
    }
  }
}

.ledger-type-label {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
}

.ledger-type-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 16px;
  height: 16px;
  border-radius: 999px;
  background: #fff;
  font-size: 10px;
  line-height: 1;
}

.ledger-type-icon--income {
  color: #1677ff;
}

.ledger-type-icon--expense {
  color: #f90;
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
  .middleBox {
    max-width: 100%;
  }
}
</style>
