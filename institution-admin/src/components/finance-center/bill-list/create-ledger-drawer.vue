<script setup>
import {
  AuditOutlined,
  BankOutlined,
  CheckOutlined,
  CloseOutlined,
  CustomerServiceOutlined,
  ExperimentOutlined,
  HomeOutlined,
  NotificationOutlined,
  PaperClipOutlined,
  PlaySquareOutlined,
  ReadOutlined,
  SafetyCertificateOutlined,
  ShopOutlined,
  TagOutlined,
  ThunderboltOutlined,
  WalletOutlined,
} from '@ant-design/icons-vue'
import dayjs from 'dayjs'
import { computed, reactive, ref, watch } from 'vue'
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
  ledgerType: 1,
  payDate: dayjs().format('YYYY-MM-DD'),
  dealStaffId: undefined,
})

const activeExpenseCategory = ref('management')
const activeLedgerItemKey = ref('')

const incomeCategoryItems = [
  { key: 'exam', label: '考试费用', icon: ReadOutlined },
  { key: 'show', label: '演出费用', icon: PlaySquareOutlined },
  { key: 'instrument', label: '乐器费用', icon: CustomerServiceOutlined },
  { key: 'meal', label: '餐饮费用', icon: ShopOutlined },
  { key: 'other', label: '其他', icon: TagOutlined },
]

const expenseCategoryGroups = [
  {
    key: 'management',
    label: '管理费用',
    items: [
      { key: 'office', label: '办公用品', icon: PaperClipOutlined },
      { key: 'water', label: '水费', icon: ExperimentOutlined },
      { key: 'electricity', label: '电费', icon: ThunderboltOutlined },
      { key: 'rent', label: '房租', icon: BankOutlined },
      { key: 'property', label: '物业', icon: HomeOutlined },
      { key: 'salary', label: '工资', icon: WalletOutlined },
      { key: 'fund', label: '公积金', icon: SafetyCertificateOutlined },
      { key: 'insurance', label: '社保', icon: AuditOutlined },
    ],
  },
  {
    key: 'sales',
    label: '销售费用',
    items: [{ key: 'marketing', label: '营销费用', icon: NotificationOutlined }],
  },
  {
    key: 'finance',
    label: '财务费用',
    items: [{ key: 'tax', label: '税', icon: BankOutlined }],
  },
]

const activeExpenseItems = computed(() => {
  return expenseCategoryGroups.find(item => item.key === activeExpenseCategory.value)?.items || []
})

function disabledDate(current) {
  return current > dayjs().endOf('day')
}

function selectExpenseCategory(key) {
  activeExpenseCategory.value = key
}

function selectLedgerItem(key) {
  activeLedgerItemKey.value = key
}

watch(
  () => formState.ledgerType,
  () => {
    activeExpenseCategory.value = 'management'
    activeLedgerItemKey.value = ''
  },
)
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
          <div class="middleBox pb2 bg-white">
            <div>
              <div class="lebal text-#666">
                <span
                  class="text-#f03 mr1"
                  style="font-family: SimSun, sans-serif"
                >*</span>账单金额：
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
                    <span><span
                      class="text-#f03 mr1"
                      style="font-family: SimSun, sans-serif"
                    >*</span>支付日期：</span>
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
                    <span><span
                      class="text-#f03 mr1"
                      style="font-family: SimSun, sans-serif"
                    >*</span>经办人：</span>
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
            <span class="ledger-type-text">收入</span>
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
            <span class="ledger-type-text">支出</span>
          </span>
        </a-radio-button>
      </a-radio-group>

      <div class="ledger-category-entry">
        <div v-if="formState.ledgerType === 1" class="ledger-inline-panel">
          <div class="ledger-expense-tabs">
            <button
              type="button"
              class="ledger-expense-trigger ledger-expense-trigger--income-active"
            >
              其他业务
            </button>
          </div>
          <div class="ledger-category-surface">
            <div class="ledger-category-grid">
              <button
                v-for="item in incomeCategoryItems"
                :key="item.key"
                class="ledger-category-item ledger-category-item--income"
                :class="{
                  'ledger-category-item--active': activeLedgerItemKey === item.key,
                }"
                type="button"
                @click="selectLedgerItem(item.key)"
              >
                <div class="ledger-category-item__icon">
                  <component :is="item.icon" />
                </div>
                <span class="ledger-category-item__label">{{ item.label }}</span>
              </button>
            </div>
          </div>
        </div>

        <div v-else class="ledger-inline-panel">
          <div class="ledger-expense-tabs">
            <button
              v-for="group in expenseCategoryGroups"
              :key="group.key"
              type="button"
              class="ledger-expense-trigger"
              :class="{
                'ledger-expense-trigger--active': activeExpenseCategory === group.key,
              }"
              @click="selectExpenseCategory(group.key)"
            >
              {{ group.label }}
            </button>
          </div>
          <div class="ledger-category-surface">
            <div class="ledger-category-grid">
              <button
                v-for="item in activeExpenseItems"
                :key="item.key"
                class="ledger-category-item ledger-category-item--expense"
                :class="{
                  'ledger-category-item--active': activeLedgerItemKey === item.key,
                }"
                type="button"
                @click="selectLedgerItem(item.key)"
              >
                <div class="ledger-category-item__icon">
                  <component :is="item.icon" />
                </div>
                <span class="ledger-category-item__label">{{ item.label }}</span>
              </button>
            </div>
          </div>
        </div>
      </div>
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
    text-align: center;
  }

  :deep(
      .ant-radio-button-wrapper-checked:not(.ant-radio-button-wrapper-disabled)
    ) {
    color: #fff;
  }

  :deep(
      .ledger-type-option--income.ant-radio-button-wrapper-checked:not(
          .ant-radio-button-wrapper-disabled
        )
    ) {
    background: #1677ff;
    border-color: #1677ff;
  }

  :deep(
      .ledger-type-option--expense.ant-radio-button-wrapper-checked:not(
          .ant-radio-button-wrapper-disabled
        )
    ) {
    background: #f90;
    border-color: #f90;
  }
}

.ledger-category-entry {
  padding-top: 16px;
}

.ledger-inline-panel {
  width: 100%;
}

.ledger-expense-tabs {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 14px;
}

.ledger-expense-trigger {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  position: relative;
  min-width: 88px;
  height: 28px;
  padding: 0 16px;
  border: none;
  border-radius: 999px;
  background: #fff;
  color: #222;
  cursor: pointer;
  line-height: 28px;
}

.ledger-expense-trigger--active {
  background: #fff4e6;
  color: #f90;
}

.ledger-expense-trigger--income-active {
  background: #edf5ff;
  color: #1677ff;
}

.ledger-expense-trigger--income-active::after,
.ledger-expense-trigger--active::after {
  content: "";
  width: 0;
  height: 0;
  position: absolute;
  top: 34px;
  left: 31px;
  border-right: 10px solid #0000;
  border-left: 10px solid #0000;
  border-bottom: 10px solid #fff;
}

.ledger-category-surface {
  min-height: 112px;
  margin: 0 8px 0 0;
  border-radius: 16px;
  background: #fff;
  padding: 16px;
}

.ledger-category-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}

.ledger-category-item {
  display: flex;
  width: 72px;
  padding: 0;
  border: none;
  background: transparent;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  color: #222;
  cursor: pointer;
}

.ledger-category-item__icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: #f6f7fb;
  color: #d0d4dc;
  font-size: 20px;
}

.ledger-category-item--income:hover .ledger-category-item__icon,
.ledger-category-item--income.ledger-category-item--active .ledger-category-item__icon {
  background: #edf5ff;
  color: #1677ff;
}

.ledger-category-item--expense:hover .ledger-category-item__icon,
.ledger-category-item--expense.ledger-category-item--active .ledger-category-item__icon {
  background: #fff4e6;
  color: #f90;
}

.ledger-category-item__label {
  font-size: 14px;
  line-height: 20px;
  text-align: center;
  white-space: nowrap;
}

.ledger-type-label {
  position: relative;
  display: inline-block;
}

.ledger-type-text {
  display: inline-block;
}

.ledger-type-icon {
  position: absolute;
  top: 50%;
  right: calc(100% + 6px);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 16px;
  height: 16px;
  border-radius: 999px;
  background: #fff;
  font-size: 10px;
  line-height: 1;
  transform: translateY(-50%);
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
