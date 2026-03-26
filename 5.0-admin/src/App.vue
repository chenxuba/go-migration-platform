<script setup lang="ts">
import { computed, onUnmounted, ref } from 'vue'
import { useLayoutMenuProvide } from '~/components/page-container/context'
import GlobalModal from '~/components/global-modal/index.vue'

const appStore = useAppStore()
const { theme } = storeToRefs(appStore)
const { antd } = useI18nLocale()
const layoutMenu = useLayoutMenu()
useLayoutMenuProvide(layoutMenu, appStore)
// 初始位置
const initialPosition = {
  x: window.innerWidth - 22 - 58, // 基于原 right:22px 计算
  y: window.innerHeight - 100 - 58, // 基于原 bottom:100px 计算
}

// 响应式状态
const isDragging = ref(false)
const position = ref({ x: initialPosition.x, y: initialPosition.y })
const draggableElement = ref(null)

// 动态位置样式
const positionStyle = computed(() => ({
  left: `${position.value.x}px`,
  top: `${position.value.y}px`,
}))

// 事件处理
let startX = 0
let startY = 0

function startDrag(e) {
  isDragging.value = true
  startX = e.clientX - position.value.x
  startY = e.clientY - position.value.y

  document.addEventListener('mousemove', handleDrag)
  document.addEventListener('mouseup', stopDrag)
}

function handleDrag(e) {
  if (!isDragging.value)
    return

  let newX = e.clientX - startX
  let newY = e.clientY - startY

  // 边界检查
  const maxX = window.innerWidth - draggableElement.value.offsetWidth
  const maxY = window.innerHeight - draggableElement.value.offsetHeight
  newX = Math.max(0, Math.min(newX, maxX))
  newY = Math.max(0, Math.min(newY, maxY))

  position.value = { x: newX, y: newY }
}

function stopDrag() {
  isDragging.value = false
  document.removeEventListener('mousemove', handleDrag)
  document.removeEventListener('mouseup', stopDrag)
}

// 组件卸载时清理事件监听
onUnmounted(() => {
  document.removeEventListener('mousemove', handleDrag)
  document.removeEventListener('mouseup', stopDrag)
})
</script>

<template>
  <a-config-provider :theme="theme" :locale="antd">
    <a-app class="h-full font-chinese antialiased">
      <TokenProvider>
        <RouterView />
        <GlobalModal />
      </TokenProvider>
    </a-app>
  </a-config-provider>
  <div ref="draggableElement" class="fixed cursor-pointer" :style="positionStyle" @mousedown="startDrag">
    <img
      width="58" height="58" src="https://xiaoguanai.oss-cn-hangzhou.aliyuncs.com/xai/xiaobao.png" alt=""
      style="pointer-events: none"
    >
  </div>
</template>

<style>
html,
body,
#app {
  /* min-width: 1100px;  */
  overflow: auto;
}
</style>

<style lang="less" scoped>
.fixed {
  position: fixed;
  user-select: none;
  /* 禁止文字选中 */
  cursor: move;
  /* 拖拽时显示移动光标 */
  z-index:99999;
}
</style>
