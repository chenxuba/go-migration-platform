<script setup>
import { ref, computed } from 'vue'

// Props
const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
})

// Emits
const emit = defineEmits(['update:open', 'confirm'])

// 处理双向绑定
const openModal = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

// 优惠券码
const couponCode = ref('')

// 关闭弹窗
function handleClose() {
  couponCode.value = ''
  openModal.value = false
}

// 确认使用优惠券
function handleConfirm() {
  // 这里可以添加验证逻辑
  console.log('输入的优惠券码:', couponCode.value)
  emit('confirm', couponCode.value)
  handleClose()
}
</script>

<template>
  <a-modal 
    v-model:open="openModal" 
    title="请输入优惠券码" 
    centered
    :footer="null"
    @cancel="handleClose"
    width="440px"
  >
    <div class="coupon-modal-body">
      <!-- 提示信息 -->
      <div class="coupon-tip">
        <div class="tip-icon"></div>
        <div class="tip-text">
          仅可使用学员已有的优惠券核销<br>
          学员可至"家校 - 我的 - 优惠券"中查看券码
        </div>
      </div>
      
      <!-- 输入框和按钮 -->
      <div class="coupon-input-group">
        <a-input 
          v-model:value="couponCode" 
          placeholder="请输入优惠券码"
          class="coupon-input"
          @press-enter="handleConfirm"
        />
        <a-button 
          type="primary" 
          class="verify-btn"
          @click="handleConfirm"
        >
          验证使用
        </a-button>
      </div>
    </div>
  </a-modal>
</template>

<style lang="less" scoped>
/* 优惠券弹窗样式 */
.coupon-modal-body {
  .coupon-tip {
    display: flex;
    align-items: flex-start;
    margin-bottom: 20px;
    padding: 16px;
    background: #f0f5ff;
    border-radius: 8px;
    
    .tip-icon {
      width: 8px;
      height: 8px;
      background: #1890ff;
      border-radius: 50%;
      margin-top: 6px;
      margin-right: 8px;
      flex-shrink: 0;
    }
    
    .tip-text {
      color: #666;
      font-size: 13px;
      line-height: 20px;
    }
  }
  
  .coupon-input-group {
    display: flex;
    gap: 12px;
    align-items: center;
    margin-bottom: 20px;
    
    .coupon-input {
      flex: 1;
      height: 40px;
    }
    
    .verify-btn {
      height: 40px;
      padding: 0 20px;
    }
  }
}
</style> 