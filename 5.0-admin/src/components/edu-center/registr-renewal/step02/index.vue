<script setup>
import { computed, reactive, ref, watch } from "vue";
import dayjs from "dayjs";
import { debounce } from "lodash-es";
import { postPayOrderApi } from "~@/api/edu-center/registr-renewal";
import messageService from "~@/utils/messageService";

// Props
const props = defineProps({
  paymentSummary: {
    type: String,
    default: "",
  },
  orderData: {
    type: Object,
    default: () => ({}),
  },
});

// Emits
const emit = defineEmits(["submit-payment"]);

const payType = ref("1");
const pay = ref([]);
const checkOptions = reactive([
  {
    id: 1,
    label: "微信",
    img: "https://pcsys.admin.ybc365.com/e068b5e2-e27e-4228-8437-fae315326ced.png",
  },
  {
    id: 2,
    label: "支付宝",
    img: "https://pcsys.admin.ybc365.com/3f396285-ac3a-43fe-94a8-adfef395d47d.png",
  },
  {
    id: 3,
    label: "银行转帐",
    img: "https://pcsys.admin.ybc365.com/b05c2b92-6c09-46a0-8123-42446727d480.png",
  },
  {
    id: 4,
    label: "POS机",
    img: "https://pcsys.admin.ybc365.com/2a6137fb-3e35-470f-b368-2fe71e5439f1.png",
  },
  {
    id: 5,
    label: "现金",
    img: "https://pcsys.admin.ybc365.com/bab59869-17ea-42c9-9354-d88a59ecf18a.png",
  },
  {
    id: 6,
    label: "其他",
    img: "https://pcsys.admin.ybc365.com/6d6faf00-aaea-4f6c-b3ab-e581a9bcf7f6.png",
  },
]);

const accountData = reactive([]);
const accountFormRefs = ref([]);
const submitLoading = ref(false);
const payAmountErrors = reactive({});

// 原始的提交函数
async function submitPayment() {
  if (submitLoading.value) return;

  // 清除之前的错误提示
  Object.keys(payAmountErrors).forEach((key) => {
    delete payAmountErrors[key];
  });

  // 检查每个收款方式的实收金额不能为0
  let hasError = false;
  for (const item of accountData) {
    const amount = Number(item.payAmount) || 0;
    if (amount <= 0) {
      payAmountErrors[item.payMethodId] = true;
      hasError = true;
    }
  }

  if (hasError) {
    messageService.error("实收金额不可为0");
    return;
  }

  try {
    submitLoading.value = true;

    // 表单验证
    await Promise.all(
      accountFormRefs.value.map((formRef) => {
        return formRef.validate();
      })
    );
    console.log("所有表单验证通过");

    // 准备提交的支付数据
    const paymentData = {
      payType: payType.value,
      payAmount: payAmount.value,
      owedAmount: owedAmount.value,
      receivableAmount: receivableAmount.value,
      paymentMethods: accountData.map((item) => ({
        payMethod: item.payMethodId,
        payTitle: item.payTitle,
        img: item.img,
        payAmount: Number(item.payAmount),
        amountId: item.amountId,
        payTime: item.payTime,
        paymentVoucher: item.paymentVoucher,
        orderId: props.orderData.orderId,
      })),
      paymentSummary: paymentSummary.value,
      orderId: props.orderData.orderId,
    };

    // 根据API要求的格式组装数据
    const apiData = {
      orderId: props.orderData.orderId,
      payAmount: payAmount.value,
      payAccounts: accountData.map((item) => ({
        orderId: props.orderData.orderId,
        amountId: item.amountId,
        payMethod: item.payMethodId,
        payAmount: Number(item.payAmount),
        payTime: item.payTime,
        paymentVoucher: item.paymentVoucher?.text || "",
      })),
    };

    console.log("原始数据:", paymentData);
    console.log("API数据:", apiData);

    // 调用API
    const res = await postPayOrderApi(apiData);
    console.log("提交成功:", res);

    // 成功后触发事件
    emit("submit-payment", paymentData);
  } catch (error) {
    console.log("提交失败:", error);
    // 可以在这里添加错误提示
  } finally {
    submitLoading.value = false;
  }
}

// 带防抖的提交函数
const validateForms = debounce(submitPayment, 500, {
  leading: true,
  trailing: false,
});

const accountList = ref([{ value: "1", label: "默认账户" }]);

// 禁止选择今天之后的日期
function disabledDate(current) {
  return current > dayjs().endOf("day");
}

// 根据日期获取对应的星期几
function getWeekDay(dateString) {
  if (!dateString) return "";
  const weekdays = ["周日", "周一", "周二", "周三", "周四", "周五", "周六"];
  const date = dayjs(dateString);
  return weekdays[date.day()];
}

// 监听支付方式变化
watch(
  () => [...pay.value],
  (newVal, oldVal) => {
    // 处理新增项
    newVal.forEach((id) => {
      if (!oldVal.includes(id)) {
        const option = checkOptions.find((opt) => opt.id === id);

        // 计算当前已收金额总和
        const currentTotal = accountData.reduce((total, item) => {
          return total + (Number(item.payAmount) || 0);
        }, 0);

        // 计算剩余未收金额
        const totalAmount = props.orderData?.amountInfo?.totalAmount || 0;
        const remainingAmount = totalAmount - currentTotal;

        // 确定默认填充金额
        let defaultAmount = 0;
        if (accountData.length === 0) {
          // 第一个收款方式，填充全部应收金额
          defaultAmount = totalAmount;
        } else {
          // 后续收款方式，填充剩余金额（如果剩余金额大于0）
          defaultAmount = remainingAmount > 0 ? remainingAmount : 0;
        }

        accountData.push({
          payMethodId: option.id,
          payTitle: option.label,
          img: option.img,
          payAmount: defaultAmount,
          amountId: accountList.value[0]?.value || null,
          payTime: dayjs().format("YYYY-MM-DD"),
          paymentVoucher: { text: "", images: [] },
        });
      }
    });

    // 处理移除项
    const removed = oldVal.filter((id) => !newVal.includes(id));
    removed.forEach((id) => {
      const index = accountData.findIndex(
        (item) =>
          item.payTitle === checkOptions.find((opt) => opt.id === id)?.label
      );
      if (index > -1) {
        const removedItem = accountData[index];
        // 清除被移除项的错误提示
        if (payAmountErrors[removedItem.payMethodId]) {
          delete payAmountErrors[removedItem.payMethodId];
        }
        accountData.splice(index, 1);
      }
    });
  },
  { deep: true }
);

// 取消选择处理
function handleCancel(item) {
  const target = checkOptions.find((opt) => opt.label === item.payTitle);
  pay.value = pay.value.filter((id) => id !== target?.id);

  // 清除当前项的错误提示
  if (payAmountErrors[item.payMethodId]) {
    delete payAmountErrors[item.payMethodId];
  }
}

// 计算支付方式汇总信息
const paymentSummary = computed(() => {
  const groups = accountData
    .filter((item) => Number(item.payAmount) > 0)
    .map((item) => `${item.payTitle} ¥${Number(item.payAmount).toFixed(2)}`);

  return groups.length > 0 ? groups.join("，") : "";
});

// 计算应收金额
const receivableAmount = computed(() => {
  return props.orderData?.amountInfo?.totalAmount || 0;
});

// 计算实收金额总和
const payAmount = computed(() => {
  return accountData.reduce((total, item) => {
    const amount = Number(item.payAmount) || 0;
    return total + amount;
  }, 0);
});

// 计算欠费金额
const owedAmount = computed(() => {
  const receivable = receivableAmount.value;
  const actual = payAmount.value;
  const owed = receivable - actual;
  return owed > 0 ? owed : 0;
});

// 计算每个收款方式的最大可输入金额
function getMaxAmount(currentItem) {
  const totalAmount = receivableAmount.value;

  // 计算除当前项之外的其他项的金额总和
  const otherItemsTotal = accountData.reduce((total, item) => {
    if (item !== currentItem) {
      return total + (Number(item.payAmount) || 0);
    }
    return total;
  }, 0);

  // 最大可输入金额 = 应收总金额 - 其他项金额总和
  const maxAmount = totalAmount - otherItemsTotal;
  return maxAmount > 0 ? maxAmount : 0;
}

// 处理金额输入变化，确保不超过最大值
function handleAmountChange(item) {
  const maxAmount = getMaxAmount(item);
  const currentAmount = Number(item.payAmount) || 0;

  if (currentAmount > maxAmount) {
    item.payAmount = maxAmount.toFixed(2);
  }

  // 清除当前项的错误提示
  if (payAmountErrors[item.payMethodId]) {
    delete payAmountErrors[item.payMethodId];
  }
}
</script>

<template>
  <div class="current1 current0">
    <div class="step current1-auto bg-white rounded-4 mt-3 justify-start p-6">
      <div class="Yprice"><span>*</span>应收金额：</div>
      <div class="Yprice-num">
        ¥ {{ Number(receivableAmount || 0).toFixed(2) }}
      </div>
      <div class="radio-pay-type">
        <a-radio-group v-model:value="payType" button-style="solid">
          <a-radio-button value="1"> 已收款只记账 </a-radio-button>
          <a-radio-button value="2"> 面对面收款 </a-radio-button>
        </a-radio-group>
        <div class="payList mt-4">
          <div class="payList-title">
            <span>*</span>收款方式
            <span class="payList-tip ml-2">请选择</span>
          </div>
          <div class="pay">
            <a-checkbox-group v-model:value="pay">
              <label
                v-for="(item, index) in checkOptions"
                :key="index"
                class="pay-box"
                :class="{ active: pay.includes(item.id) }"
              >
                <span> <img :src="item.img" alt="" /> {{ item.label }}</span>
                <a-checkbox :value="item.id" />
              </label>
            </a-checkbox-group>
          </div>
        </div>
        <div class="payDetail">
          <div
            v-for="(item, index) in accountData"
            :key="index"
            class="step bg-white rounded-4 justify-start"
          >
            <div class="conductContent">
              <div class="container-box mb-4">
                <div class="container-box-top">
                  <div class="flex flex-items-center">
                    <span
                      class="font-600 text-#222 text-4 flex flex-items-center"
                    >
                      <img width="23" class="mt--0.5" :src="item.img" alt="" />
                      <span class="ml-1">{{ item.payTitle }}</span>
                    </span>
                  </div>
                  <div class="right">
                    <span class="price flex flex-items-center relative"
                      >实收金额：
                      <a-input-number
                        v-model:value="item.payAmount"
                        style="width: 150px"
                        :max="getMaxAmount(item)"
                        :min="0"
                        :precision="2"
                        step="0.01"
                        :status="
                          payAmountErrors[item.payMethodId] ? 'error' : ''
                        "
                        @change="handleAmountChange(item)"
                        @input="handleAmountChange(item)"
                      >
                        <template #addonAfter> 元 </template>
                      </a-input-number>
                      <span class="max-tip ml-2 text-#999 text-3">
                        (最多 ¥{{ getMaxAmount(item).toFixed(2) }})
                      </span>
                      <span
                        v-if="payAmountErrors[item.payMethodId]"
                        class="text-red text-12px absolute bottom--22px left-68px"
                        >实收金额不可为0</span
                      >
                    </span>
                    <span class="cancel ml-5" @click="handleCancel(item)"
                      >取消选择</span
                    >
                  </div>
                </div>
                <div class="container-box-bottom">
                  <a-form ref="accountFormRefs" layout="vertical" :model="item">
                    <a-space :size="24">
                      <!-- 收款账户必选校验 -->
                      <a-form-item
                        label="收款账户"
                        name="amountId"
                        :rules="[
                          {
                            required: true,
                            message: '请选择收款账户',
                            trigger: 'change',
                          },
                        ]"
                      >
                        <a-select
                          v-model:value="item.amountId"
                          placeholder="请选择收款账户"
                          :options="accountList"
                          style="width: 170px"
                        />
                      </a-form-item>
                      <div style="border-right: 1px solid #eee; height: 80px" />
                      <!-- 支付日期 -->
                      <a-form-item
                        label="支付日期"
                        name="payTime"
                        :rules="[
                          {
                            required: true,
                            message: '请选择支付日期',
                            trigger: 'change',
                          },
                        ]"
                      >
                        <div class="flex flex-items-center week-wrap">
                          <a-date-picker
                            v-model:value="item.payTime"
                            :disabled-date="disabledDate"
                            format="YYYY-MM-DD"
                            value-format="YYYY-MM-DD"
                            class="w-35"
                            placeholder="请选择日期"
                          />
                          <div v-if="item.payTime" class="week">
                            {{ getWeekDay(item.payTime) }}
                          </div>
                        </div>
                      </a-form-item>
                      <div style="border-right: 1px solid #eee; height: 80px" />
                      <a-form-item label="账单备注（选填）">
                        <div class="flex flex-items-center week-wrap">
                          <a-button>填写</a-button>
                        </div>
                      </a-form-item>
                    </a-space>
                  </a-form>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
    <a-affix :offset-bottom="0">
      <div class="mt-3 fixedcss">
        <div class="h-20 flex flex-center justify-end pr-10 bg-white rounded-4">
          <div class="text-5 totalPrice mr-12 flex">
            <div>
              实收金额：<span>¥ {{ payAmount.toFixed(2) }}</span>
            </div>
            <div class="text-3.5 text-#666 font-400 flex">
              {{ paymentSummary }}
              <div v-if="owedAmount > 0" class="ml-12px">
                欠费
                <b class="text-#f03 text-3.5">¥ {{ owedAmount.toFixed(2) }}</b>
              </div>
            </div>
          </div>
          <a-button
            class="h-11 w-35 text-5 font-600"
            :disabled="payAmount <= 0 || submitLoading"
            :loading="submitLoading"
            type="primary"
            @click="validateForms"
          >
            {{ submitLoading ? "提交中..." : "提交" }}
          </a-button>
        </div>
      </div>
    </a-affix>
  </div>
</template>

<style lang="less" scoped>
.step {
  box-shadow: 0 0 8px 0 rgba(94, 188, 255, 0.08);
}

.current1 {
  .current1-auto {
    max-height: calc(100vh - 300px);
    overflow: auto;
  }

  .Yprice {
    font-size: 14px;

    span {
      color: red;
    }
  }

  .Yprice-num {
    padding-top: 12px;
    width: 560px;
    font-family: DINAlternate-Bold, DINAlternate, sans-serif;
    font-weight: 700;
    font-size: 48px;
    color: #000;
    line-height: 56px;
  }

  .radio-pay-type {
    padding-top: 20px;
  }

  .payList {
    span {
      color: red;
    }

    span.payList-tip {
      color: var(--pro-ant-color-primary);
    }
  }

  .pay {
    margin-top: 10px;
    margin-bottom: 20px;
    .pay-box {
      border: 1px solid #eee;
      padding: 12px 16px;
      display: flex;
      align-items: center;
      border-radius: 6px;
      margin-right: 6px;
      user-select: none;
      cursor: pointer;
      margin-bottom: 16px;

      &:hover {
        border-color: var(--pro-ant-color-primary);
      }

      span {
        color: #000;
        margin-right: 20px;
        display: flex;
        align-items: center;

        img {
          width: 20px;
          height: 20px;
          margin-right: 6px;
        }
      }
    }

    .active {
      border-color: var(--pro-ant-color-primary);
    }
  }

  .week-wrap {
    display: flex;

    .week {
      background: #e6f0ff;
      border-radius: 14px;
      color: var(--pro-ant-color-primary);
      font-size: 14px;
      font-weight: 400;
      line-height: 20px;
      margin-left: 8px;
      padding: 2px 14px;
    }
  }

  .fixedcss {
    position: absolute;
    bottom: 0;
    left: 0;
    right: 0;
  }

  .conductContent {
    .container-box {
      background: #fafafa;
      overflow-x: auto;

      .container-box-top {
        height: 44px;
        background: #f0f5fe;
        padding: 10px 24px;
        display: flex;
        justify-content: space-between;
        align-items: center;
        border-top-left-radius: 16px;
        border-top-right-radius: 16px;

        .right {
          display: flex;
          align-items: center;

          span.price {
            font-weight: 500;
            color: #222;
            white-space: nowrap;
            font-size: 14px;
          }

          .cancel {
            color: var(--pro-ant-color-primary);
            font-weight: bold;
            cursor: pointer;
          }
        }
      }

      .container-box-bottom {
        padding: 10px 8px 0 24px;

        :deep(.ant-form-item-label label) {
          white-space: nowrap;
        }
      }
    }
  }

  .totalPrice {
    font-weight: bold;
    display: flex;
    align-items: flex-end;
    flex-direction: column;
    line-height: 1.2;
    font-family: "DIN alternate", sans-serif;

    span {
      color: var(--pro-ant-color-primary);
      font-family: "DIN alternate", sans-serif;
      font-size: 38px;
    }
  }
}
</style>
