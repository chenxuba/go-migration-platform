<script setup>
const props = defineProps({
  record: {
    type: Object,
    default: () => ({}),
  },
})
</script>

<template>
  <div class="m-12px">
    <div class="bg-#fff pt-18px px-20px pb-18px rounded-10px">
      <div class="text-#222 text-5 font-600 mb-18px">
        通用1v1
      </div>
      <div class="flex flex-wrap gap-10px mb-18px">
        <span class="tag-pill tag-pill--primary">通用课</span>
        <span class="tag-pill">全部课程通用</span>
        <span class="tag-pill">1对1授课</span>
        <span class="tag-pill">课时</span>
      </div>

      <div class="detail-grid">
        <div class="detail-column">
          <div class="detail-line">
            <span class="detail-label">当前状态：</span>
            <span>在读</span>
          </div>
          <div class="detail-line">
            <span class="detail-label">已用数量：</span>
            <span>4.00 课时</span>
          </div>
          <div class="detail-line">
            <span class="detail-label">已用学费金额：</span>
            <span>￥ 400.00</span>
          </div>
        </div>

        <div class="detail-column">
          <div class="detail-line">
            <span class="detail-label">报读数量：</span>
            <span>购 11 课时</span>
          </div>
          <div class="detail-line">
            <span class="detail-label">剩余数量：</span>
            <span>7.00 课时</span>
          </div>
          <div class="detail-line">
            <span class="detail-label">剩余学费金额：</span>
            <span>￥ 700</span>
          </div>
        </div>

        <div class="detail-column">
          <div class="detail-line">
            <span class="detail-label">有效期至：</span>
            <span>不限制</span>
          </div>
          <div class="detail-line">
            <span class="detail-label">总学费：</span>
            <span>￥1100</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style lang="less" scoped>
.tag-pill {
  display: inline-flex;
  align-items: center;
  height: 34px;
  padding: 0 16px;
  border-radius: 999px;
  background: #edf3ff;
  color: #2b6de5;
  font-size: 14px;
  font-weight: 500;

  &--primary {
    background: #1f6fff;
    color: #fff;
  }
}

.detail-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 32px;
}

.detail-column {
  min-width: 0;
}

.detail-line {
  display: flex;
  align-items: baseline;
  margin-bottom: 20px;
  color: #555;
  font-size: 16px;
}

.detail-label {
  color: #222;
  font-weight: 500;
  margin-right: 10px;
}
</style>
