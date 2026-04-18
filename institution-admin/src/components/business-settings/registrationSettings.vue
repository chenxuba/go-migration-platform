<script setup>
import { RightOutlined } from '@ant-design/icons-vue'

const toggleValue1 = ref(false)
const toggleValue2 = ref(false)
</script>

<template>
  <div class="registration-settings">
    <!-- 报名优惠模式设置 -->
    <div class="bg-white settings-item">
      <div class="item-content">
        <div class="item-title" style="font-size: 16px;">
          报名优惠模式设置
        </div>
        <div class="item-description line-height-16px mt-4px">
          切换报名/续费时的优惠模式
        </div>
      </div>
      <div class="item-arrow">
        <RightOutlined />
      </div>
    </div>
    <!-- 自定义报价单 -->
    <div class="bg-white settings-item mt-10px">
      <div class="item-content">
        <div class="item-title" style="font-size: 16px;">
          自定义报价单
        </div>
        <div class="item-description line-height-16px mt-4px">
          开启后，报名/续费时可自定义填写购买数量与金额
        </div>
      </div>
      <div class="item-status ml-20px">
        <a-switch v-model:checked="toggleValue1" @click.stop />
      </div>
    </div>
    <!-- 退课支持数量为0 -->
    <div class="bg-white settings-item mt-10px">
      <div class="item-content">
        <div class="item-title" style="font-size: 16px;">
          退课支持数量为0
        </div>
        <div class="item-description line-height-16px mt-4px">
          开启后，退课支持填写退0课时/天数，当填写退课数量为0，且填写了退课费用后，退课费用将会全部记录为退课亏损费
        </div>
      </div>
      <div class="item-status ml-20px">
        <a-switch v-model:checked="toggleValue2" @click.stop />
      </div>
    </div>
  </div>
</template>

<style lang="less" scoped>
.registration-settings {
  padding: 12px 0px;
  background-color: #f6f7f8;
  height: 100%;
  :deep(.ant-collapse-header) {
    padding: 0 !important;
  }

  :deep(.ant-collapse-expand-icon) {
    position: absolute;
    right: 12px;
    padding-inline-end: 0px !important;
  }

  :deep(.ant-collapse-content-box) {
    padding: 0px !important;
  }

  .settings-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 12px 12px 12px 16px;
    border-top: 1px solid #f0f0f0;
    cursor: pointer;
  }

  .item-title {
    font-size: 14px;
    color: rgba(0, 0, 0, 0.85);
  }

  .item-description {
    font-size: 12px;
    color: rgba(0, 0, 0, 0.45);
  }
}
</style>
