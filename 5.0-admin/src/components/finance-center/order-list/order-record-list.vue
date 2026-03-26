<script setup>
import { computed, onMounted, ref } from "vue";
import { useRouter } from "vue-router";
import { debounce } from "lodash-es";
import { DownOutlined, InfoCircleOutlined } from "@ant-design/icons-vue";
import dayjs from "dayjs";
import { useTableColumns } from "@/composables/useTableColumns";
import { useStudentListRefresh } from "@/composables/useStudentListRefresh";
import {
  getOrderListApi,
  setBadDebtApi,
  cancelBadDebtApi,
} from "@/api/finance-center/order-manage";
import messageService from "@/utils/messageService";
import AllFilter from "@/components/common/all-filter.vue";
const router = useRouter();
const allFilterRef = ref(null);
const tagOverflowMap = ref({});
const currentYearStart = dayjs().startOf("year").format("YYYY-MM-DD");
const today = dayjs().format("YYYY-MM-DD");
const defaultCreateTimeVals = ref([currentYearStart, today]);
const defaultOrderStatusVals = ref([1, 3, 7, 8, 6, 2]);

const dataSource = ref([]);
const loading = ref(false);
const summary = ref({
  totalPaid: 0,
  totalArrear: 0,
  totalBadDebt: 0,
});
const displayArray = ref([
  "orderNumber",
  "orderType",
  "orderTag",
  "orderSource",
  "orderStatus",
  "handleContent",
  "salesPerson",
  "createUser",
  "dealDate",
  "createTime",
  "orderArrearStatus",
  "latestPaidTime",
]);
const pagination = ref({
  current: 1,
  pageSize: 10,
  total: 0,
  showSizeChanger: true,
  showQuickJumper: true,
  showTotal: (total) => `共 ${total} 条`,
});

const queryState = ref({
  keyword: undefined,
  keywordType: undefined,
  studentId: undefined,
  orderTypeList: undefined,
  orderTagIds: undefined,
  orderStatusList: [1, 3, 7, 8, 6, 2],
  orderSourceList: undefined,
  courseIds: undefined,
  salePersonId: undefined,
  creatorId: undefined,
  orderArrearStatus: undefined,
  dealDateBegin: undefined,
  dealDateEnd: undefined,
  createdTimeBegin: currentYearStart,
  createdTimeEnd: today,
  latestPaidTimeBegin: undefined,
  latestPaidTimeEnd: undefined,
});
const allColumns = ref([
  {
    title: "订单编号",
    dataIndex: "orderNumber",
    key: "orderNumber",
    fixed: "left",
    width: 210,
    required: true,
  },
  {
    title: "报名学员",
    dataIndex: "studentName",
    key: "studentName",
    fixed: "left",
    width: 190,
    required: true,
  },
  {
    title: "订单类型",
    key: "orderType",
    dataIndex: "orderType",
    width: 130,
  },
  {
    title: "订单来源",
    key: "orderSource",
    dataIndex: "orderSource",
    width: 130,
  },
  {
    title: "订单标签",
    key: "tagNames",
    dataIndex: "tagNames",
    width: 130,
  },
  {
    title: "订单状态",
    key: "orderStatus",
    dataIndex: "orderStatus",
    width: 130,
  },
  {
    title: "办理内容",
    dataIndex: "productItems",
    key: "productItems",
    width: 200,
  },
  {
    title: "订单销售员",
    dataIndex: "salePersonName",
    key: "salePersonName",
    width: 130,
  },
  {
    title: "经办人",
    dataIndex: "staffName",
    key: "staffName",
    width: 130,
  },
  {
    title: "经办日期",
    dataIndex: "dealDate",
    key: "dealDate",
    width: 130,
  },
  {
    title: "订单创建时间",
    dataIndex: "createdTime",
    key: "createdTime",
    width: 160,
  },
  {
    title: "储值账户变动",
    key: "rechargeAccount",
    dataIndex: "rechargeAccount",
    width: 160,
  },
  {
    title: "优惠",
    dataIndex: "discount",
    key: "discount",
    width: 100,
  },
  {
    title: "整单优惠",
    dataIndex: "wideDiscountName",
    key: "wideDiscountName",
    width: 120,
  },
  {
    title: "应收/应退(元)",
    dataIndex: "totalAmount",
    key: "totalAmount",
    width: 140,
  },
  {
    title: "实收/实退(元)",
    dataIndex: "paidAmount",
    key: "paidAmount",
    width: 140,
  },
  {
    title: "欠费金额(元)",
    dataIndex: "arrearAmount",
    key: "arrearAmount",
    width: 150,
  },
  {
    title: "坏账金额(元)",
    dataIndex: "badDebtAmount",
    key: "badDebtAmount",
    width: 150,
  },
  {
    title: "最近支付时间",
    dataIndex: "latestPaidTime",
    key: "latestPaidTime",
    width: 160,
  },
  {
    title: "平账抵扣",
    dataIndex: "totalChargeAgainstAmount",
    key: "totalChargeAgainstAmount",
    width: 120,
  },
  {
    title: "对内备注",
    dataIndex: "remark",
    key: "remark",
    width: 150,
  },
  {
    title: "对外备注",
    dataIndex: "externalRemark",
    key: "externalRemark",
    width: 150,
  },
  {
    title: "学员备注",
    dataIndex: "customerRemark",
    key: "customerRemark",
    width: 150,
  },
  {
    title: "操作",
    dataIndex: "action",
    key: "action",
    fixed: "right",
    width: 140,
  },
]);

const { selectedValues, columnOptions, filteredColumns, totalWidth } =
  useTableColumns({
    storageKey: "order-record-list",
    allColumns,
    excludeKeys: ["action"],
  });

const openDrawer = ref(false);
function handleSeeStuData() {
  openDrawer.value = true;
}

const openOrderDetailDrawer = ref(false);
const currentOrderId = ref("");
const openRechargePayDrawer = ref(false);
const currentRechargeSaleOrderId = ref("");
function handleOrderDetail(orderId) {
  currentOrderId.value = orderId;
  openOrderDetailDrawer.value = true;
}

function handleOrderDetailDrawerClosed() {
  fetchOrderList();
}

function normalizeOrderItem(item) {
  const orderType = Number(item?.orderType || 0);
  const isRechargeOrder = orderType === 2;
  const isRefundRecharge = orderType === 4;
  const useStoredValueRao = isRechargeOrder || isRefundRecharge;

  return {
    ...item,
    totalAmount: item.totalAmount ?? item.amount ?? 0,
    paidAmount: item.paidAmount ?? 0,
    arrearAmount: item.arrearAmount ?? 0,
    productItems: Array.isArray(item.productItems) ? item.productItems : [],
    tagNames: Array.isArray(item.tagNames) ? item.tagNames : [],
    rechargeAccountAmount:
      item.rechargeAccountAmount ?? (useStoredValueRao ? item.amount ?? 0 : 0),
    rechargeAccountResidualAmount:
      item.rechargeAccountResidualAmount ??
      (useStoredValueRao ? item.residualAmount ?? 0 : 0),
    rechargeAccountGivingAmount:
      item.rechargeAccountGivingAmount ??
      (useStoredValueRao ? item.givingAmount ?? 0 : 0),
    totalChargeAgainstAmount: item.totalChargeAgainstAmount ?? 0,
    badDebtAmount: item.badDebtAmount ?? 0,
    isBadDebt: item.isBadDebt ?? false,
    avatar: item.avatar || undefined,
    sex: item.sex ?? 2,
  };
}

/** 储值账户充值(2) 为正；储值账户退费(4) 为负（红色展示） */
function formatRechargeChangeAmount(record, fieldKey) {
  const n = Number(record[fieldKey] || 0);
  const abs = Math.abs(n).toFixed(2);
  if (Number(record?.orderType) === 4) return `-${abs}`;
  return `+${abs}`;
}

function shouldShowRechargeAccountChange(record) {
  const t = Number(record?.orderType || 0);
  return t === 2 || t === 4;
}

/** 退费类订单：列表「应收/应退、实收/实退」展示负号与红色 */
function isRefundDisplayOrder(record) {
  return [3, 4, 6, 7].includes(Number(record?.orderType || 0));
}

function getOrderTagDisplayText(tagNames) {
  return Array.isArray(tagNames)
    ? tagNames.map((tag) => `【${tag}】`).join("、")
    : "";
}

function handleTagMouseEnter(orderId, event) {
  const el = event?.currentTarget;
  if (!orderId || !el) return;
  tagOverflowMap.value = {
    ...tagOverflowMap.value,
    [orderId]: el.scrollWidth > el.clientWidth,
  };
}

function resetQueryState() {
  queryState.value = {
    keyword: undefined,
    keywordType: undefined,
    studentId: undefined,
    orderTypeList: undefined,
    orderTagIds: undefined,
    orderStatusList: [1, 3, 7, 8, 6, 2],
    orderSourceList: undefined,
    courseIds: undefined,
    salePersonId: undefined,
    creatorId: undefined,
    orderArrearStatus: undefined,
    dealDateBegin: undefined,
    dealDateEnd: undefined,
    createdTimeBegin: currentYearStart,
    createdTimeEnd: today,
    latestPaidTimeBegin: undefined,
    latestPaidTimeEnd: undefined,
  };
}

const handleFilterUpdate = debounce(
  (updates, isClearAll = false, id, type) => {
    if (isClearAll) {
      resetQueryState();
    } else {
      Object.entries(updates).forEach(([key, value]) => {
        queryState.value[key] = value;
      });
    }
    pagination.value.current = 1;
    fetchOrderList(id, type);
  },
  300,
  { leading: true, trailing: false }
);

function mapRange(fieldPrefix, value, isClearAll, id, type) {
  if (Array.isArray(value) && value.length === 2) {
    handleFilterUpdate(
      {
        [`${fieldPrefix}Begin`]: value[0],
        [`${fieldPrefix}End`]: value[1],
      },
      isClearAll,
      id,
      type
    );
    return;
  }
  handleFilterUpdate(
    {
      [`${fieldPrefix}Begin`]: undefined,
      [`${fieldPrefix}End`]: undefined,
    },
    isClearAll,
    id,
    type
  );
}

const filterUpdateHandlers = computed(() => ({
  "update:orderNumberFilter": (val, isClearAll, id, type) =>
    handleFilterUpdate(
      {
        keyword: val || undefined,
        keywordType: val ? "orderNumber" : undefined,
      },
      isClearAll,
      id,
      type
    ),
  "update:stuPhoneSearchFilter": (val, isClearAll, id, type) =>
    handleFilterUpdate({ studentId: val || undefined }, isClearAll, id, type),
  "update:orderTypeFilter": (val, isClearAll, id, type) =>
    handleFilterUpdate(
      { orderTypeList: Array.isArray(val) && val.length ? val : undefined },
      isClearAll,
      id,
      type
    ),
  "update:orderTagFilter": (val, isClearAll, id, type) =>
    handleFilterUpdate(
      {
        orderTagIds:
          Array.isArray(val) && val.length ? val.map(String) : undefined,
      },
      isClearAll,
      id,
      type
    ),
  "update:orderStatusFilter": (val, isClearAll, id, type) =>
    handleFilterUpdate(
      { orderStatusList: Array.isArray(val) && val.length ? val : undefined },
      isClearAll,
      id,
      type
    ),
  "update:orderSourceFilter": (val, isClearAll, id, type) =>
    handleFilterUpdate(
      { orderSourceList: Array.isArray(val) && val.length ? val : undefined },
      isClearAll,
      id,
      type
    ),
  "update:handleContentFilter": (val, isClearAll, id, type) =>
    handleFilterUpdate(
      {
        courseIds:
          Array.isArray(val) && val.length ? val.map(String) : undefined,
      },
      isClearAll,
      id,
      type
    ),
  "update:salesPersonFilter": (val, isClearAll, id, type) =>
    handleFilterUpdate(
      { salePersonId: val || undefined },
      isClearAll,
      id,
      type
    ),
  "update:createUserFilter": (val, isClearAll, id, type) =>
    handleFilterUpdate({ creatorId: val || undefined }, isClearAll, id, type),
  "update:orderArrearStatusFilter": (val, isClearAll, id, type) =>
    handleFilterUpdate(
      { orderArrearStatus: Array.isArray(val) && val.length ? val : undefined },
      isClearAll,
      id,
      type
    ),
  "update:createTimeFilter": (val, isClearAll, id, type) =>
    mapRange("createdTime", val, isClearAll, id, type),
  "update:dealDateFilter": (val, isClearAll, id, type) =>
    mapRange("dealDate", val, isClearAll, id, type),
  "update:latestPaidTimeFilter": (val, isClearAll, id, type) =>
    mapRange("latestPaidTime", val, isClearAll, id, type),
}));

function openRechargePay(orderId) {
  currentRechargeSaleOrderId.value = String(orderId || "");
  openRechargePayDrawer.value = true;
}

// 去付款
function handlePayOrder(record) {
  if (Number(record?.orderType) === 2) {
    openRechargePay(record.orderId);
    return;
  }
  router.push({
    path: `/edu-center/registr-renewal/${record.orderId}`,
    query: { step: "1" }, // step=1 表示直接进入付款步骤
  });
}

// 补费
function handleRepayment(record) {
  if (Number(record?.orderType) === 2) {
    openRechargePay(record.orderId);
    return;
  }
  router.push({
    path: `/edu-center/registr-renewal/${record.orderId}`,
    query: { step: "1" }, // 直接进入付款步骤
  });
}

// 设为坏账
const badDebtModalVisible = ref(false);
const badDebtRemark = ref("");
const currentBadDebtOrderId = ref("");

function handleSetBadDebt(orderId) {
  currentBadDebtOrderId.value = orderId;
  badDebtRemark.value = "";
  badDebtModalVisible.value = true;
}

async function confirmSetBadDebt() {
  try {
    await setBadDebtApi({
      orderId: currentBadDebtOrderId.value,
      remark: badDebtRemark.value,
    });

    messageService.success("设为坏账成功");
    badDebtModalVisible.value = false;
    // 刷新列表
    fetchOrderList();
  } catch (error) {
    console.error("设为坏账失败:", error);
    messageService.error(error.message || "设为坏账失败");
  }
}

// 取消坏账
const cancelBadDebtModalVisible = ref(false);
const currentCancelBadDebtOrder = ref(null);

function handleCancelBadDebt(orderId) {
  // 从列表中找到对应的订单信息
  const order = dataSource.value.find((item) => item.orderId === orderId);
  if (order) {
    currentCancelBadDebtOrder.value = order;
    cancelBadDebtModalVisible.value = true;
  }
}

async function confirmCancelBadDebt() {
  try {
    await cancelBadDebtApi({
      orderId: currentCancelBadDebtOrder.value.orderId,
    });
    messageService.success("取消坏账成功");
    cancelBadDebtModalVisible.value = false;
    // 刷新列表
    fetchOrderList();
  } catch (error) {
    console.error("取消坏账失败:", error);
    messageService.error(error.message || "取消坏账失败");
  }
}

// 打印收据
function handlePrintReceipt(orderId) {
  console.log("打印收据:", orderId);
  messageService.info("打印收据功能开发中");
}

// 下载收据
function handleDownloadReceipt(orderId) {
  console.log("下载收据:", orderId);
  messageService.info("下载收据功能开发中");
}

// 发送短信
function handleSendSms(orderId) {
  console.log("发送短信:", orderId);
  messageService.info("发送短信功能开发中");
}

// 关闭订单
function handleCloseOrder(orderId) {
  console.log("关闭订单:", orderId);
  messageService.info("关闭订单功能开发中");
}

// 获取订单列表
async function fetchOrderList(id, type) {
  try {
    loading.value = true;
    const { result } = await getOrderListApi({
      sortModel: {},
      queryModel: queryState.value,
      pageRequestModel: {
        needTotal: true,
        pageSize: pagination.value.pageSize,
        pageIndex: pagination.value.current,
      },
    });

    if (result) {
      const orders = Array.isArray(result.list)
        ? result.list.map(normalizeOrderItem)
        : [];
      dataSource.value = orders;
      pagination.value.total = result.total || 0;
      summary.value = {
        totalPaid: Number(result.totalPaid || 0),
        totalArrear: Number(result.totalArrear || 0),
        totalBadDebt: Number(result.totalBadDebt || 0),
      };
      if (type) {
        allFilterRef.value?.clearQuickFilter(id, type);
      }
    }
  } catch (error) {
    console.error("获取订单列表失败:", error);
    messageService.error("获取订单列表失败");
  } finally {
    loading.value = false;
  }
}

// 分页变化
function handleTableChange(pag) {
  pagination.value.current = pag.current;
  pagination.value.pageSize = pag.pageSize;
  fetchOrderList();
}

// 订单状态映射
const orderStatusMap = {
  1: "待付款",
  2: "审批中",
  3: "已完成",
  4: "已关闭",
  5: "已作废",
  6: "待处理",
  7: "退费中",
  8: "已退费",
};

// 订单类型映射
const orderTypeMap = {
  1: "报名续费",
  2: "储值账户充值",
  3: "退课",
  4: "储值账户退费",
  5: "转课",
  6: "退教材费",
  7: "退学杂费",
};

// 订单来源映射
const orderSourceMap = {
  1: "线下办理",
  2: "微校报名",
  3: "线下导入",
  4: "续费订单",
};

// 格式化日期（带时分）
function formatDate(dateStr) {
  if (!dateStr) return "-";
  return dateStr.replace("T", " ").substring(0, 16);
}

// 格式化日期（仅年月日，用于经办日期）
function formatDateOnly(dateStr) {
  if (!dateStr) return "-";
  if (dateStr.includes("T")) {
    return dateStr.split("T")[0];
  }
  return dateStr.substring(0, 10);
}

function handleImportStudentOrder() {
  router.push("/import-center/starter/order");
}

// 初始化加载
onMounted(() => {
  resetQueryState();
  fetchOrderList();
});

// 学员详情编辑/关闭后统一刷新订单列表
useStudentListRefresh(fetchOrderList);

defineExpose({
  refreshData: () => fetchOrderList(),
});
</script>

<template>
  <div>
    <div class="filter-wrap bg-white pl-3 pr-3 rounded-lb-4 rounded-rb-4">
      <all-filter
        ref="allFilterRef"
        :display-array="displayArray"
        :default-create-time-vals="defaultCreateTimeVals"
        :default-order-status-vals="defaultOrderStatusVals"
        :is-quick-show="false"
        :is-show-search-stu-phonefilter="true"
        create-user-label="经办人"
        create-user-placeholder="请输入经办人"
        sales-person-label="订单销售员"
        sales-person-placeholder="请输入订单销售员"
        create-time-label="订单创建时间"
        v-on="filterUpdateHandlers"
      />
    </div>
    <div class="student-list mt-3 pt-3 pb-3 pl-6 pr-6 bg-white rounded-4">
      <div class="tab-table">
        <div class="table-title flex justify-between">
          <div class="total">
            共 {{ pagination.total }} 条订单，实收总计
            {{ summary.totalPaid.toFixed(2) }} 元，共欠费
            {{ summary.totalArrear.toFixed(2) }} 元，共坏账
            {{ summary.totalBadDebt.toFixed(2) }} 元
          </div>
          <div class="edit flex">
            <a-dropdown class="mr-2">
              <template #overlay>
                <a-menu>
                  <a-menu-item key="0" @click="handleImportStudentOrder"> 导入学员订单 </a-menu-item>
                  <a-menu-item key="1"> 批量导出 </a-menu-item>
                  <a-menu-item key="3"> 导出记录 </a-menu-item>
                </a-menu>
              </template>
              <a-button>
                导入/导出学员订单
                <DownOutlined :style="{ fontSize: '10px' }" />
              </a-button>
            </a-dropdown>
            <customize-code
              v-model:checked-values="selectedValues"
              :options="columnOptions"
              :total="allColumns.length - 1"
              :num="selectedValues.length"
            />
          </div>
        </div>
        <div class="table-content mt-2">
          <a-table
            :data-source="dataSource"
            :pagination="pagination"
            :columns="filteredColumns"
            :scroll="{ x: totalWidth }"
            :sticky="{ offsetHeader: 100 }"
            :loading="loading"
            size="small"
            @change="handleTableChange"
          >
            <template #headerCell="{ column }">
              <template v-if="column.key === 'orderStatus'">
                <span class="mr-1">{{ column.title }}</span>
                <a-popover
                  placement="top"
                  trigger="hover"
                  overlay-class-name="order-status-popover"
                >
                  <template #content>
                    <div class="w-100 text-#666 leading-6">
                      <div
                        style="
                          color: #222;
                          font-weight: 500;
                          padding-bottom: 8px;
                          margin-bottom: 12px;
                          border-bottom: 1px solid #f0f0f0;
                        "
                      >
                        字段说明
                      </div>
                      <div>待付款：尚未完成付款的订单。</div>
                      <div>
                        待处理：尚未设置授课课程开始时间或教务用品待发货的订单。
                      </div>
                      <div>已完成：已完成一次支付的订单。</div>
                      <div>退费中：尚未完成退费的订单。</div>
                      <div>已退费：已完成退费的订单。</div>
                      <div>已关闭：未完成支付已关闭的订单。</div>
                      <div>已作废：已完成支付已作废的订单。</div>
                      <div>审批中：已触发审批在审核状态中的订单。</div>
                    </div>
                  </template>
                  <InfoCircleOutlined
                    class="cursor-pointer text-#999 hover:text-#1677ff"
                  />
                </a-popover>
              </template>
              <template v-else-if="column.key === 'totalChargeAgainstAmount'">
                <span class="mr-1">{{ column.title }}</span>
                <a-popover placement="top" trigger="hover">
                  <template #content>
                    <div class="w-100 text-#666 leading-6">
                      <div
                        style="
                          color: #222;
                          font-weight: 500;
                          padding-bottom: 8px;
                          margin-bottom: 12px;
                          border-bottom: 1px solid #f0f0f0;
                        "
                      >
                        字段说明
                      </div>
                      <div>课程退课后，剩余学费抵扣订单欠费的金额</div>
                    </div>
                  </template>
                  <InfoCircleOutlined
                    class="cursor-pointer text-#999 hover:text-#1677ff"
                  />
                </a-popover>
              </template>
            </template>
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'studentName'">
                <student-avatar
                  :id="record.studentId"
                  :name="record.studentName"
                  :gender="
                    record.sex === 0 ? '女' : record.sex === 1 ? '男' : '未知'
                  "
                  :avatar-url="record.avatar"
                  :phone="record.studentPhone"
                  default-active-key="0"
                />
              </template>

              <template v-if="column.key === 'orderNumber'">
                <span
                  class="text-#06f flex-center justify-start cursor-pointer"
                  @click="handleOrderDetail(record.orderId)"
                >
                  {{ record.orderNumber }}
                  <!-- 坏账标记 -->
                  <a-tooltip v-if="record.isBadDebt">
                    <template #title>订单已设为坏账</template>
                    <span
                      class="w-5 h-5 block text-#333 bg-#DDD font-600 text-3 ml-1 text-center line-height-5 rounded-1"
                      >坏</span
                    >
                  </a-tooltip>
                  <!-- 欠费标记（非坏账） -->
                  <a-tooltip
                    v-else-if="record.isAmountOwed && record.orderStatus !== 1"
                  >
                    <template #title>订单欠费未缴清</template>
                    <span
                      class="w-5 h-5 block text-red bg-#FBE7E6 font-600 text-3 ml-1 text-center line-height-5 rounded-1"
                      >欠</span
                    >
                  </a-tooltip>
                </span>
              </template>

              <template v-else-if="column.key === 'orderType'">
                <span
                  :class="
                    Number(record.orderType) === 4
                      ? 'text-#ff3333 font-normal'
                      : ''
                  "
                >
                  {{ orderTypeMap[record.orderType] || "-" }}
                </span>
              </template>

              <template v-else-if="column.key === 'orderSource'">
                {{ orderSourceMap[record.orderSource] || "-" }}
              </template>

              <template v-else-if="column.key === 'tagNames'">
                <a-tooltip
                  v-if="record.tagNames && record.tagNames.length"
                  :title="
                    tagOverflowMap[record.orderId]
                      ? getOrderTagDisplayText(record.tagNames)
                      : null
                  "
                >
                  <div
                    @mouseenter="handleTagMouseEnter(record.orderId, $event)"
                    class="truncate whitespace-nowrap overflow-hidden text-ellipsis max-w-180px"
                  >
                    {{ getOrderTagDisplayText(record.tagNames) }}
                  </div>
                </a-tooltip>
                <span v-else>-</span>
              </template>

              <template v-else-if="column.key === 'orderStatus'">
                <span class="orderStatus">
                  <span
                    class="dot"
                    :style="{
                      background:
                        record.orderStatus === 3
                          ? '#1890ff'
                          : record.orderStatus === 6
                          ? '#1677ff'
                          : record.orderStatus === 8
                          ? '#52c41a'
                          : record.orderStatus === 7
                          ? '#fa8c16'
                          : record.orderStatus === 1
                          ? '#faad14'
                          : record.orderStatus === 2
                          ? '#faad14'
                          : record.orderStatus === 4
                          ? '#d9d9d9'
                          : record.orderStatus === 5
                          ? '#ff4d4f'
                          : '#d9d9d9',
                    }"
                  />
                  <span>{{ orderStatusMap[record.orderStatus] || "-" }}</span>
                </span>
              </template>

              <template v-else-if="column.key === 'productItems'">
                <div v-for="(item, index) in record.productItems" :key="index">
                  {{ item }}
                </div>
                <span
                  v-if="
                    !record.productItems || record.productItems.length === 0
                  "
                  >-</span
                >
              </template>

              <template v-else-if="column.key === 'salePersonName'">
                {{ record.salePersonName || "-" }}
              </template>

              <template v-else-if="column.key === 'staffName'">
                {{ record.staffName || "-" }}
              </template>

              <template v-else-if="column.key === 'dealDate'">
                {{ formatDateOnly(record.dealDate) }}
              </template>

              <template v-else-if="column.key === 'createdTime'">
                {{ formatDate(record.createdTime) }}
              </template>

              <template v-else-if="column.key === 'rechargeAccount'">
                <div
                  v-if="shouldShowRechargeAccountChange(record)"
                  class="text-3.2 leading-tight"
                  :class="Number(record.orderType) === 4 ? 'text-#ff3333' : ''"
                >
                  <div class="flex justify-between gap-2">
                    <span>充值金额</span>
                    <span class="font-600 tabular-nums">{{
                      formatRechargeChangeAmount(
                        record,
                        "rechargeAccountAmount"
                      )
                    }}</span>
                  </div>
                  <div class="flex justify-between gap-2">
                    <span>残联金额</span>
                    <span class="font-600 tabular-nums">{{
                      formatRechargeChangeAmount(
                        record,
                        "rechargeAccountResidualAmount"
                      )
                    }}</span>
                  </div>
                  <div class="flex justify-between gap-2">
                    <span>赠送金额</span>
                    <span class="font-600 tabular-nums">{{
                      formatRechargeChangeAmount(
                        record,
                        "rechargeAccountGivingAmount"
                      )
                    }}</span>
                  </div>
                </div>
                <span v-else>-</span>
              </template>

              <template v-else-if="column.key === 'discount'"> - </template>

              <template v-else-if="column.key === 'wideDiscountName'">
                <span v-if="record.wideDiscountName" class="font-600">
                  {{ record.wideDiscountName }}
                </span>
                <span v-else>-</span>
              </template>

              <template v-else-if="column.key === 'totalAmount'">
                <div
                  class="font-600"
                  :class="
                    isRefundDisplayOrder(record) ? 'text-#ff3333' : 'text-#222'
                  "
                >
                  {{ isRefundDisplayOrder(record) ? "-" : "+" }}¥
                  {{ Math.abs(Number(record.totalAmount) || 0).toFixed(2) }}
                </div>
              </template>

              <template v-else-if="column.key === 'paidAmount'">
                <div
                  class="font-600"
                  :class="
                    isRefundDisplayOrder(record) ? 'text-#ff3333' : 'text-#222'
                  "
                >
                  {{ isRefundDisplayOrder(record) ? "-" : "+" }}¥
                  {{ Math.abs(Number(record.paidAmount) || 0).toFixed(2) }}
                </div>
                <div class="text-#888">
                  共{{ record.productItems?.length || 0 }}件商品
                </div>
              </template>

              <template v-else-if="column.key === 'arrearAmount'">
                <!-- 如果是坏账订单，欠费金额不显示 -->
                <div
                  v-if="!record.isBadDebt && record.arrearAmount > 0"
                  class="text-#ff3333 font-600"
                >
                  ¥ {{ record.arrearAmount.toFixed(2) }}
                </div>
                <span v-else>-</span>
              </template>

              <template v-else-if="column.key === 'badDebtAmount'">
                <div v-if="record.isBadDebt && record.badDebtAmount > 0">
                  <span class="text-#ff3333 font-600"
                    >¥ {{ record.badDebtAmount.toFixed(2) }}</span
                  >
                  <a-tooltip v-if="record.badDebtRemark">
                    <template #title>{{ record.badDebtRemark }}</template>
                    <InfoCircleOutlined class="ml-1 text-#888" />
                  </a-tooltip>
                </div>
                <span v-else>-</span>
              </template>

              <template v-else-if="column.key === 'latestPaidTime'">
                {{
                  Number(record.orderType) === 4
                    ? "-"
                    : formatDate(record.latestPaidTime)
                }}
              </template>

              <template v-else-if="column.key === 'totalChargeAgainstAmount'">
                <span
                  v-if="record.totalChargeAgainstAmount > 0"
                  class="font-600"
                >
                  {{ record.totalChargeAgainstAmount }}
                </span>
                <span v-else>-</span>
              </template>

              <template v-else-if="column.key === 'remark'">
                <a-tooltip v-if="record.remark" :title="record.remark">
                  <div class="text-ellipsis">
                    {{ record.remark }}
                  </div>
                </a-tooltip>
                <span v-else>-</span>
              </template>

              <template v-else-if="column.key === 'externalRemark'">
                <a-tooltip
                  v-if="record.externalRemark"
                  :title="record.externalRemark"
                >
                  <div class="text-ellipsis">
                    {{ record.externalRemark }}
                  </div>
                </a-tooltip>
                <span v-else>-</span>
              </template>

              <template v-else-if="column.key === 'customerRemark'">
                <a-tooltip
                  v-if="record.customerRemark"
                  :title="record.customerRemark"
                >
                  <div class="text-ellipsis">
                    {{ record.customerRemark }}
                  </div>
                </a-tooltip>
                <span v-else>-</span>
              </template>

              <template v-if="column.key === 'action'">
                <!-- 待付款订单：显示去付款和关闭订单 -->
                <template v-if="record.orderStatus === 1">
                  <a class="text-#06f mr-2" @click="handlePayOrder(record)">
                    去付款
                  </a>
                  <a
                    class="text-#06f"
                    @click="handleCloseOrder(record.orderId)"
                  >
                    关闭订单
                  </a>
                </template>

                <!-- 审批中订单：不显示查看详情 -->
                <span v-else-if="record.orderStatus === 2">-</span>

                <!-- 已完成订单：根据坏账/欠费状态显示不同操作 -->
                <template v-else-if="record.orderStatus === 3">
                  <template v-if="record.isBadDebt">
                    <a
                      class="text-#06f"
                      @click="handleCancelBadDebt(record.orderId)"
                    >
                      取消坏账
                    </a>
                  </template>
                  <template v-else-if="record.isAmountOwed">
                    <a class="text-#06f mr-2" @click="handleRepayment(record)">
                      补费
                    </a>
                    <a
                      class="text-#06f"
                      @click="handleSetBadDebt(record.orderId)"
                    >
                      设为坏账
                    </a>
                  </template>
                  <a-dropdown v-else>
                    <a class="text-#06f">
                      查看收据
                      <DownOutlined :style="{ fontSize: '10px' }" />
                    </a>
                    <template #overlay>
                      <a-menu>
                        <a-menu-item
                          key="print"
                          @click="handlePrintReceipt(record.orderId)"
                        >
                          打印收据
                        </a-menu-item>
                        <a-menu-item
                          key="download"
                          @click="handleDownloadReceipt(record.orderId)"
                        >
                          下载收据
                        </a-menu-item>
                        <a-menu-item
                          key="sms"
                          @click="handleSendSms(record.orderId)"
                        >
                          发送短信
                        </a-menu-item>
                      </a-menu>
                    </template>
                  </a-dropdown>
                </template>

                <!-- 其他状态：显示查看详情 -->
                <a
                  v-else
                  class="text-#06f"
                  @click="handleOrderDetail(record.orderId)"
                >
                  查看详情
                </a>
              </template>
            </template>
          </a-table>
        </div>
      </div>
    </div>
    <student-info-drawer v-model:open="openDrawer" />
    <order-detail-drawer
      v-model:open="openOrderDetailDrawer"
      :order-id="currentOrderId"
      @updated="fetchOrderList"
      @closed="handleOrderDetailDrawerClosed"
    />
    <recharge-order-pay-drawer
      v-model:open="openRechargePayDrawer"
      :sale-order-id="currentRechargeSaleOrderId"
      @submitted="fetchOrderList"
    />

    <!-- 设为坏账对话框 -->
    <a-modal
      v-model:open="badDebtModalVisible"
      title="设为坏账"
      ok-text="确认"
      cancel-text="取消"
      :ok-button-props="{ danger: true }"
      @ok="confirmSetBadDebt"
    >
      <p>确认将此订单设为坏账吗？</p>
      <p class="text-#ff4d4f mt-2">设为坏账后，该订单的欠费将不再追缴。</p>
      <a-textarea
        v-model:value="badDebtRemark"
        placeholder="请输入坏账备注（选填）"
        :rows="4"
        class="mt-3"
      />
    </a-modal>

    <!-- 取消坏账对话框 -->
    <a-modal
      v-model:open="cancelBadDebtModalVisible"
      title="取消坏账"
      ok-text="确认"
      cancel-text="取消"
      width="600px"
      @ok="confirmCancelBadDebt"
    >
      <div
        v-if="currentCancelBadDebtOrder"
        class="bg-#f7f7fd p-4 pb-1 rounded-2"
      >
        <div class="mb-3">
          <span class="text-#666">订单编号：</span>
          <span class="text-#222 font-600">{{
            currentCancelBadDebtOrder.orderNumber
          }}</span>
          <span class="ml-6 text-#666">欠费金额：</span>
          <span class="text-#ff4d4f font-600"
            >¥{{
              currentCancelBadDebtOrder.badDebtAmount?.toFixed(2) || "0.00"
            }}</span
          >
        </div>
        <div>
          <span class="text-#666">坏账原因：</span>
          <span class="text-#222">{{
            currentCancelBadDebtOrder.badDebtRemark || "-"
          }}</span>
        </div>
        <p
          class="text-#666 mt-4 border-0 border-t-1px border-solid border-#ddd p-t-4"
        >
          是否确认取消该订单坏账？
        </p>
      </div>
    </a-modal>
  </div>
</template>

<style lang="less" scoped>
.total {
  position: relative;
  padding-left: 10px;
  color: #222;
  display: flex;
  align-items: center;

  &::before {
    display: inline-block;
    background: var(--pro-ant-color-primary);
    border-radius: 2px;
    content: "";
    height: 12px;
    left: 0;
    position: absolute;
    width: 4px;
  }
}

.studentStatus {
  span.dot {
    border-radius: 50%;
    display: inline-block;
    height: 6px;
    position: relative;
    vertical-align: middle;
    width: 6px;
    margin-right: 4px;
    background: var(--pro-ant-color-primary);
  }
}

.orderStatus {
  span.dot {
    border-radius: 50%;
    display: inline-block;
    height: 6px;
    position: relative;
    vertical-align: middle;
    width: 6px;
    margin-right: 4px;
  }
}

.tip {
  padding: 10px 24px 10px 14px;
  background: #e6f0ff;
  color: #333;

  a {
    color: var(--pro-ant-color-primary);
  }
}

.upNew {
  position: relative;

  &::before {
    position: absolute;
    top: -12px;
    left: -22px;
    z-index: 999;
    width: 39px;
    height: 22px;
    background: url(data:image/png;base64,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);
    background-size: contain;
    content: "";
  }
}
</style>
