<script setup>
import { Modal } from 'ant-design-vue'
import { PlusCircleFilled, QuestionCircleOutlined } from '@ant-design/icons-vue'
import { onBeforeRouteLeave, useRoute } from 'vue-router'
import { h, ref } from 'vue'

const route = useRoute()
const orderNum = route.params.id
const checked = ref(false)
const formState = reactive({
  employeeId: null, // 员工
  performanceAmount: undefined, // 分配业绩金额
  remark: null, // 备注
})
const details = reactive({
  'id': '611311192488555520',
  'assignedAmount': 0,
  'status': 0,
  'orderNumber': '20250617144623560897530',
  'orderType': 1,
  'orderAmount': 360,
  'orderStaffId': '576940597315664926',
  'orderStaffName': '陈瑞',
  'orderCreatedTime': '2025-06-17T14:46:24',
  'orderSalePersonId': '580458472080941056',
  'orderSalePersonName': '何红武',
  'students': [
    {
      'advisorId': '0',
      'advisorName': '',
      'salespersonId': '580458472080941056',
      'salespersonName': '何红武',
      'collectorName': '',
      'phoneSellName': '',
      'viceSellName': '',
      'foregroundName': '',
      'studentManagerName': '',
      'performanceId': '0',
      'id': '609417555739700224',
      'name': '何黑五',
      'phone': '199****0001',
    },
  ],
  'subPerformances': [
    {
      'assignmentRecords': [],
      'id': '611311192488555522',
      'sourceId': '609878475498638336',
      'sourceName': '编程课',
      'sourceType': 1,
      'sourceDescType': 1,
      'sourceDescDetailModel': {
        'productSourceModel': {
          'skuCount': 1,
          'skuUnit': 1,
          'freeQuantity': 0,
          'discountType': 0,
          'discountNumber': 0,
          'hasValidDate': false,
          'validDate': '0001-01-01T00:00:00',
          'shareDiscount': 0,
          'tuition': 60,
          'quantity': 1,
          'realQuantity': 1,
          'realTuition': 60,
          'shareRechargeAccountGivingAmount': 0,
        },
        'refundOrderSourceModel': null,
        'rechargeAccountSourceModel': null,
        'transferOrderSourceModel': null,
        'materialsProductRefundOrderSourceModel': null,
      },
      'sourceRealAmount': 60,
      'assignedAmount': 0,
      'enrollType': 3,
    },
    {
      'assignmentRecords': [],
      'id': '611311192488555523',
      'sourceId': '608479202068436992',
      'sourceName': '书本费',
      'sourceType': 4,
      'sourceDescType': 11,
      'sourceDescDetailModel': {
        'productSourceModel': {
          'skuCount': 1,
          'skuUnit': 0,
          'freeQuantity': 0,
          'discountType': 0,
          'discountNumber': 0,
          'hasValidDate': false,
          'validDate': '0001-01-01T00:00:00',
          'shareDiscount': 0,
          'tuition': 300,
          'quantity': 1,
          'realQuantity': 1,
          'realTuition': 300,
          'shareRechargeAccountGivingAmount': 0,
        },
        'refundOrderSourceModel': null,
        'rechargeAccountSourceModel': null,
        'transferOrderSourceModel': null,
        'materialsProductRefundOrderSourceModel': null,
      },
      'sourceRealAmount': 300,
      'assignedAmount': 0,
      'enrollType': 0,
    },
  ],
})
onBeforeRouteLeave((to, from, next) => {
  Modal.confirm({
    title: '退出业绩分配?',
    centered: true,
    content: '退出后，当前页面内容将被清空',
    okText: '退出',
    cancelText: '取消',
    okButtonProps: {
      danger: true,
    },
    onOk: () => {
      next()
    },
    onCancel: () => {
      next(false)
    },
  })
})
</script>

<template>
  <div class="w-full h-full flex flex-col ">
    <div class="content flex-1 overflow-auto scrollbar">
      <a-card :bordered="false" class="h-full">
        <div class="title text-20px font-bold text-#000 flex items-center">
          订单编号：{{ orderNum }}
          <a-tag color="#e6f0ff" class="ml-10px rounded-10 font-400" style="color: #06f;">
            报名续费
          </a-tag>
        </div>
        <a-space :size="80" class="flex text-14px py-14px text-#000">
          <div>
            <span>订单总金额：</span>
            <span class="text-#888">¥ 2000</span>
          </div>
          <div>
            <span>订单创建时间：</span>
            <span class="text-#888">2025-06-13 14:46</span>
          </div>
          <div>
            <span>经办人：</span>
            <span class="text-#888">陈瑞</span>
          </div>
        </a-space>
        <a-space class="mt-8px bg-#fbfbfb flex items-center px-14px py-14px rounded-1" :size="180">
          <div>
            <span>关联学员：</span>
            <span class="text-#888">何黑五（199****0001）</span>
          </div>
          <div>
            <span>销售员：</span>
            <span class="text-#888">陈瑞</span>
          </div>
        </a-space>
        <div class="mt-12px text-20px font-bold text-#000 flex items-center">
          交易内容 <span class="text-14px text-#666 ml-12px font-400">请根据以下订单交易项设置业绩归属人并分配业绩金额</span>
        </div>
        <div v-for="item in details.subPerformances" :key="item.id" class="orderItem">
          <div
            class="header h-54px px-18px bg-#005ce60f flex items-center justify-between rounded-lt-10px rounded-rt-10px"
          >
            <div class="flex items-center">
              <span class="text-16px text-#000 font-bold mr-12px">编程课</span>
              <a-tag color="#e6f0ff" class="rounded-10 font-400" style="color: #06f;">
                编程课
              </a-tag>
              <a-tag color="#e6f0ff" class="rounded-10 font-400" style="color: #06f;">
                书本费
              </a-tag>
            </div>
            <div class="flex items-center">
              <span class="text-14px text-#222">分配业绩小计：</span>
              <span class="text-26px text-#06f font-bold custom-num-font-family price">2000</span>
            </div>
          </div>
          <div class="content">
            <div class="itemInfo">
              <a-row :gutter="12">
                <a-col :span="4">
                  <div
                    class="px-16px py-24px pl-24px border-r-1px border-b-1px border-t-0 border-l-0  border-#f0f5fe border-solid "
                  >
                    <div class="title">
                      报价单
                    </div>
                    <div class="text-#000000a6 text-14px mt-4px">
                      ¥60 / 1课时
                    </div>
                  </div>
                </a-col>
                <a-col :span="3">
                  <div
                    class="px-16px py-24px border-r-1px border-b-1px border-t-0 border-l-0  border-#f0f5fe border-solid"
                  >
                    <div class="title">
                      购买份数
                    </div>
                    <div class="text-#000000a6 text-14px mt-4px">
                      *1
                    </div>
                  </div>
                </a-col>
                <a-col :span="3">
                  <div
                    class="px-16px py-24px border-r-1px border-b-1px border-t-0 border-l-0  border-#f0f5fe border-solid"
                  >
                    <div class="title">
                      赠送课时
                    </div>
                    <div class="text-#000000a6 text-14px mt-4px">
                      -
                    </div>
                  </div>
                </a-col>
                <a-col :span="4">
                  <div
                    class="px-16px py-24px border-r-1px border-b-1px border-t-0 border-l-0  border-#f0f5fe border-solid"
                  >
                    <div class="title">
                      有效期至
                    </div>
                    <div class="text-#000000a6 text-14px  mt-4px">
                      -
                    </div>
                  </div>
                </a-col>
                <a-col :span="3">
                  <div
                    class="px-16px py-24px border-r-1px border-b-1px border-t-0 border-l-0  border-#f0f5fe border-solid flex  flex-col items-end"
                  >
                    <div class="title">
                      单课优惠
                    </div>
                    <div class="text-#000000a6 text-14px mt-4px">
                      -
                    </div>
                  </div>
                </a-col>
                <a-col :span="3">
                  <div
                    class="px-16px py-24px border-r-1px border-b-1px border-t-0 border-l-0  border-#f0f5fe border-solid flex  flex-col items-end"
                  >
                    <div class="title">
                      分摊整单优惠
                    </div>
                    <div class="text-#000000a6 text-14px mt-4px">
                      -
                    </div>
                  </div>
                </a-col>
                <a-col :span="4">
                  <div
                    class="px-16px py-24px pr-24px border-r-1px border-b-1px border-t-0 border-l-0  border-#f0f5fe border-solid flex  flex-col items-end"
                  >
                    <div class="title">
                      实收金额
                    </div>
                    <div class="text-#000000a6 text-14px mt-4px">
                      ¥ 60.00
                    </div>
                  </div>
                </a-col>
              </a-row>
              <!-- 设置业绩归属人 -->
              <div class="setting p-20px text-#16px text-#000 font-500">
                设置业绩归属人
                <a-tooltip>
                  <template #title>
                    默认展示与此订单相关员工，可根据实际情况调整
                  </template>
                  <QuestionCircleOutlined />
                </a-tooltip>
                <a-checkbox v-model:checked="checked" class="ml-24px text-#666 font400">
                  显示离职员工
                </a-checkbox>
              </div>
              <!-- 选择员工 -->
              <a-form :model="formState" class="flex px-24px">
                <a-form-item label="选择员工" class="w-20%">
                  <a-select v-model:value="formState.employeeId" placeholder="请选择员工" style="width: 100%;">
                    <a-select-option value="1">
                      Option 1
                    </a-select-option>
                    <a-select-option value="2">
                      Option 2
                    </a-select-option>
                    <a-select-option value="3">
                      Option 3
                    </a-select-option>
                  </a-select>
                </a-form-item>
                <a-form-item label="分配业绩金额" class="w-20% mx-60px">
                  <a-input-number
                    v-model:value="formState.performanceAmount" style="width: 100%;" placeholder="请输入金额"
                    :min="0.01" prefix="¥" :precision="2"
                  />
                </a-form-item>
                <a-form-item label="备注" class="flex-1 mr-100px">
                  <div class="flex items-center">
                    <a-input v-model:value="formState.remark" placeholder="请输入备注" style="width: 100%;" />
                    <span class="text-#06fc cursor-pointer whitespace-nowrap ml-20px">删除</span>
                  </div>
                </a-form-item>
              </a-form>
              <div class="add pl-10px pb-12px">
                <a-button :icon="h(PlusCircleFilled)" type="link">
                  添加业绩归属人
                </a-button>
              </div>
            </div>
          </div>
        </div>
        <div class="flex justify-end items-center mr-20px">
          <span class="text-14px text-#222">分配业绩总计：</span>
          <span class="text-26px text-#06f font-bold custom-num-font-family price">2000</span>
        </div>
      </a-card>
    </div>
    <a-affix :offset-bottom="0">
      <div class="footer flex justify-end items-center p-12px bg-#fff rounded-3 mt-14px">
        <a-space :size="14">
          <a-button>返回</a-button>
          <a-button type="primary">
            提交
          </a-button>
        </a-space>
      </div>
    </a-affix>
  </div>
</template>

<style lang="less" scoped>
.orderItem {
  margin-top: 14px;
  border: 1px solid #f0f5fe;

  .itemInfo {
    border-top: 0;
    overflow: hidden;

    :deep(.ant-col) {
      padding: 0 !important;
    }

    .title {
      color: #222;
      font-size: 14px;
      font-weight: bold;
    }
  }
}

.price {
  display: flex;
  align-items: center;

  &::before {
    content: "¥";
    margin-right: 6px;
    font-size: 20px;
    font-family: D-DINExp, D;
    font-weight: 500;
    color: #06f;
    line-height: 22px;
  }
}
</style>
