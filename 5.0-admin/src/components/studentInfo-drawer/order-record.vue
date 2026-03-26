<script setup>
const openOrderDetailDrawer = ref(false)
const dataSource = ref([{}, {}, {}, {}, {}, {}, {}, {}, {}, {}])
const allColumns = ref([
  {
    title: '订单编号',
    dataIndex: 'orderNum',
    key: 'orderNum',
    fixed: 'left',
    width: 210,
    required: true, // 新增必选标识
  },
  {
    title: '订单类型',
    key: 'orderType',
    dataIndex: 'orderType',
    width: 100,
  },
  {
    title: '订单来源',
    key: 'orderForm',
    dataIndex: 'orderForm',
    width: 100,
  },
  {
    title: '订单状态',
    key: 'orderStatus',
    dataIndex: 'orderStatus',
    width: 100,

  },
  {
    title: '办理内容',
    dataIndex: 'handleContent',
    key: 'handleContent',
    width: 120,
  },
  {
    title: '订单创建时间',
    dataIndex: 'createTime',
    key: 'createTime',
    width: 140,
  },
  {
    title: '订单总金额(元)',
    dataIndex: 'orderTotalPrice',
    key: 'orderTotalPrice',
    width: 100,
  },
])
function handleOrderDetail() {
  openOrderDetailDrawer.value = true
}
</script>

<template>
  <div class="order-record p3 pr0">
    <div class="record bg-white rounded-3 p3">
      <div class="total mb-1.5">
        共{{ dataSource.length }}条订单
      </div>
      <a-table
        :sticky="true" :data-source="dataSource" :columns="allColumns" :pagination="dataSource.length > 10" style="color: #666;"
        size="small"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'orderNum'">
            <span class="text-#06f flex-center justify-start cursor-pointer" @click="handleOrderDetail()">20250410144858420324570 {{ record.a }}
              <a-tooltip>
                <template #title>订单欠费未缴清</template>
                <span class="w-5 h-5 block text-red bg-#FBE7E6 text-3 ml-1 text-center line-height-5 rounded-1">欠</span>
              </a-tooltip>
            </span>
          </template>
          <template v-if="column.key === 'orderType'">
            报名续费
          </template>
          <template v-if="column.key === 'orderForm'">
            线下办理
          </template>
          <template v-if="column.key === 'orderStatus'">
            <div class="flex-center justify-start">
              <span class="dot" /><span>已完成</span>
            </div>
          </template>
          <template v-if="column.key === 'handleContent'">
            初级感统课
          </template>
          <template v-if="column.key === 'createTime'">
            2025-04-10 14:48
          </template>
          <template v-if="column.key === 'orderTotalPrice'">
            <div class="text-center">
              <div class="text-#222 font-500">
                +3000.00
              </div>
              <div class="text-#888">
                共1件商品
              </div>
            </div>
          </template>
        </template>
      </a-table>
    </div>
    <order-detail-drawer v-model:open="openOrderDetailDrawer" />
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

span.dot {
  border-radius: 50%;
  display: inline-block;
  height: 6px;
  position: relative;
  vertical-align: middle;
  width: 6px;
  margin-right: 4px;
  background: #06f;
}
</style>
