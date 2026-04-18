<script setup>
const toggleValue1 = ref(false)
const toggleValue2 = ref(false)
</script>

<template>
  <div>
    <div class="bg-white settings-item mt-10px">
      <div class="item-content">
        <div class="item-title" style="font-size: 16px;">
          消超点名
        </div>
        <div class="item-description line-height-16px mt-4px">
          开启后，学员的上课点名课时支持消超记录（支持欠费点名）
        </div>
      </div>
      <div class="item-status ml-20px">
        <a-switch v-model:checked="toggleValue1" @click.stop />
      </div>
    </div>
    <div class="bg-white settings-item mt-10px">
      <div class="item-content">
        <div class="item-title" style="font-size: 16px;">
          自动点名
        </div>
        <div class="item-description line-height-16px mt-4px">
          开启后，未点名课程在结束后30分钟内，系统将进行自动点名并记录课消。（若不支持消超点名，日程内任意学员剩余数量不足，此日程将不自动点名，需手动点名）
        </div>
      </div>
      <div class="item-status ml-20px">
        <a-switch v-model:checked="toggleValue2" @click.stop />
      </div>
    </div>
  </div>
</template>

<style lang="less" scoped>
.settings-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 2px 12px 12px 16px;
    border-bottom: 1px solid #f0f0f0;
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
</style>
