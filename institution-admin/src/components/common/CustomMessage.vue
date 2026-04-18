<script setup>
import { onBeforeUnmount, onMounted, ref } from 'vue'
import {
  CheckCircleFilled,
  CloseCircleFilled,
  ExclamationCircleFilled,
  InfoCircleFilled,
} from '@ant-design/icons-vue'

const props = defineProps({
  content: {
    type: String,
    default: '',
  },
  type: {
    type: String,
    default: 'info', // info, success, warning, error
  },
  duration: {
    type: Number,
    default: 3000,
  },
  onClose: {
    type: Function,
    default: () => {},
  },
})

const visible = ref(false)
const isClosing = ref(false)
const count = ref(1)
let timer = null
// 用于存储消息内容的响应式变量
const messageContent = ref(props.content)

// 显示消息
function show() {
  isClosing.value = false
  visible.value = true
  clearTimeout(timer)

  // 设置自动关闭
  if (props.duration > 0) {
    timer = setTimeout(() => {
      close()
    }, props.duration)
  }
}

// 关闭消息
function close() {
  isClosing.value = true

  // 等待动画完成后再隐藏元素
  setTimeout(() => {
    visible.value = false
    // 调用关闭回调
    props.onClose()
  }, 300)
}

// 立即关闭消息（不等待动画）
function closeImmediately() {
  visible.value = false
  // 不调用onClose回调，由外部处理
}

// 增加计数
function incrementCount() {
  count.value += 1

  // 重新显示并重置计时器
  isClosing.value = false // 确保不在关闭状态
  clearTimeout(timer) // 清除之前的计时器

  // 重新设置自动关闭计时器
  if (props.duration > 0) {
    timer = setTimeout(() => {
      close()
    }, props.duration)
  }

  // 添加动画效果到计数器
  setTimeout(() => {
    const badgeElement = document.querySelector('.custom-counter .ant-badge-count')
    if (badgeElement) {
      // 移除之前的动画类
      badgeElement.classList.remove('ant-badge-count-animate')
      // 触发浏览器重排
      void badgeElement.offsetWidth
      // 添加动画类
      badgeElement.classList.add('ant-badge-count-animate')
    }
  }, 10)
}

// 更新消息内容
function setContent(newContent) {
  messageContent.value = newContent
}

// 暴露方法供外部使用
defineExpose({
  show,
  close,
  closeImmediately,
  incrementCount,
  setContent,
  type: props.type,
  content: messageContent,
  isClosing,
})

// 组件挂载时自动显示
onMounted(() => {
  show()
})

// 组件卸载前清除定时器
onBeforeUnmount(() => {
  clearTimeout(timer)
})
</script>

<template>
  <div v-if="visible" class="custom-message-container" :class="{ closing: isClosing }">
    <div class="custom-message-content" :class="type">
      <i v-if="type === 'warning'" class="custom-message-icon">
        <ExclamationCircleFilled />
      </i>
      <i v-else-if="type === 'success'" class="custom-message-icon">
        <CheckCircleFilled />
      </i>
      <i v-else-if="type === 'error'" class="custom-message-icon">
        <CloseCircleFilled />
      </i>
      <i v-else-if="type === 'info'" class="custom-message-icon">
        <InfoCircleFilled />
      </i>
      <span class="custom-message-text text-14px">{{ messageContent }}</span>
      <a-badge v-if="count > 1" :count="count" :show-zero="false" class="custom-counter" />
    </div>
  </div>
</template>

<style scoped>
.custom-message-container {
  position: fixed !important;
  top: 0 !important;
  left: 50% !important;
  transform: translateX(-50%) !important;
  z-index: 9999 !important;
  pointer-events: none !important;
  text-align: center !important;
  width: auto !important;
  transform-origin: top center !important;
  padding-top: 18px !important;
  perspective: 1000px !important;
  animation: none !important;
}

/* 消息容器本身不再有动画，动画转移到内容上 */
.custom-message-container:not(.closing) .custom-message-content {
  animation: drawerSlideIn 0.4s cubic-bezier(0.23, 1, 0.32, 1) forwards !important;
}

.custom-message-container.closing .custom-message-content {
  animation: drawerSlideOut 0.3s cubic-bezier(0.23, 1, 0.32, 1) forwards !important;
}

@keyframes drawerSlideIn {
  0% {
    transform: rotateX(-30deg) translateY(-60px);
    box-shadow: 0 2px 4px -2px rgba(0, 0, 0, 0.1);
    opacity: 0;
  }
  40% {
    opacity: 1;
  }
  70% {
    transform: rotateX(5deg) translateY(5px);
    box-shadow: 0 3px 6px -4px rgba(0, 0, 0, 0.12), 0 6px 16px 0 rgba(0, 0, 0, 0.08);
  }
  100% {
    transform: rotateX(0) translateY(0);
    box-shadow: 0 3px 6px -4px rgba(0, 0, 0, 0.12), 0 6px 16px 0 rgba(0, 0, 0, 0.08), 0 9px 28px 8px rgba(0, 0, 0, 0.05);
    opacity: 1;
  }
}

@keyframes drawerSlideOut {
  0% {
    transform: rotateX(0) translateY(0);
    box-shadow: 0 3px 6px -4px rgba(0, 0, 0, 0.12), 0 6px 16px 0 rgba(0, 0, 0, 0.08), 0 9px 28px 8px rgba(0, 0, 0, 0.05);
    opacity: 1;
  }
  100% {
    transform: rotateX(-30deg) translateY(-60px);
    box-shadow: 0 2px 4px -2px rgba(0, 0, 0, 0.1);
    opacity: 0;
  }
}

.custom-message-content {
  position: relative !important;
  display: inline-flex !important;
  align-items: center !important;
  padding: 10px 16px !important;
  background: #fff !important;
  border-radius: 8px !important;
  box-shadow: 0 3px 6px -4px rgba(0, 0, 0, 0.12), 0 6px 16px 0 rgba(0, 0, 0, 0.08), 0 9px 28px 8px rgba(0, 0, 0, 0.05) !important;
  pointer-events: all !important;
  text-align: left !important;
  margin: 0 auto !important;
  transition: all 0.3s !important;
  transform-style: preserve-3d !important;
  backface-visibility: hidden !important;
  will-change: transform, opacity, box-shadow !important;
}

/* 自定义计数器位置 */
.custom-counter {
  position: absolute !important;
  top: -10px !important;
  right: -5px !important;
  transform: translate(25%, 0) !important;
}

:deep(.ant-badge-count) {
  background-color: #ff4d4f !important;
  box-shadow: 0 0 0 1px #fff !important;
  transition: transform 0.3s cubic-bezier(0.68, -0.55, 0.27, 1.55) !important;
}

/* 当计数增加时，为徽章添加一个弹跳动画效果 */
@keyframes badgeBounce {
  0% {
    transform: scale(0.5);
  }
  50% {
    transform: scale(1.2);
  }
  100% {
    transform: scale(1);
  }
}

/* 通过在组件上添加一个类，触发徽章动画 */
:deep(.ant-badge-count-animate) {
  animation: badgeBounce 0.5s cubic-bezier(0.68, -0.55, 0.27, 1.55);
}

/* 消息图标样式 */
.custom-message-icon {
  margin-right: 8px !important;
  font-size: 16px !important;
}

/* 不同类型消息的样式 */
.custom-message-content.info {
  border: 1px solid #91caff !important;
}
.custom-message-content.info .custom-message-icon {
  color: #1677ff !important;
}

.custom-message-content.success {
  border: 1px solid #b7eb8f !important;
}
.custom-message-content.success .custom-message-icon {
  color: #52c41a !important;
}

.custom-message-content.warning {
  border: 1px solid #ffe58f !important;
}
.custom-message-content.warning .custom-message-icon {
  color: #faad14 !important;
}

.custom-message-content.error {
  border: 1px solid #ffccc7 !important;
}
.custom-message-content.error .custom-message-icon {
  color: #ff4d4f !important;
}
</style>
