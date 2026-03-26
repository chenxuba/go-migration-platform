import { createVNode, render } from 'vue'
import CustomMessage from '../components/common/CustomMessage.vue'

// 用于存储当前活动的消息实例
let messageInstance = null
// 用于存储消息容器
let container = null

// 显示消息的主函数
function showMessage(options) {
  // 默认配置
  const defaultOptions = {
    content: '',
    type: 'info',
    duration: 3000,
    onClose: () => {},
  }

  // 合并选项
  const mergedOptions = { ...defaultOptions, ...options }

  // 如果已经有消息实例，且类型和内容相同，则增加计数
  if (messageInstance
    && messageInstance.type === mergedOptions.type
    && messageInstance.content.value === mergedOptions.content) {
    // 检查消息是否正在关闭
    if (messageInstance.isClosing && messageInstance.isClosing.value) {
      // 如果正在关闭，创建新消息而不是增加计数
      // 继续执行后面的代码创建新消息
    }
    else {
      // 如果不在关闭状态，则增加计数
      messageInstance.incrementCount()
      return messageInstance
    }
  }

  // 如果有正在显示的消息，先移除它
  if (messageInstance) {
    // 立即关闭当前消息
    messageInstance.closeImmediately()
    messageInstance = null

    // 立即清理DOM
    if (container) {
      render(null, container)
      if (container.parentNode) {
        document.body.removeChild(container)
      }
      container = null
    }
  }

  // 始终创建新容器
  container = document.createElement('div')
  container.className = 'custom-message-wrapper'
  document.body.appendChild(container)

  // 创建虚拟节点
  const vnode = createVNode(CustomMessage, {
    ...mergedOptions,
    onClose: () => {
      // 调用原来的onClose
      mergedOptions.onClose()

      // 清理资源
      messageInstance = null

      // 清理DOM
      if (container) {
        render(null, container)
        if (container.parentNode) {
          document.body.removeChild(container)
        }
        container = null
      }
    },
  })

  // 渲染到容器中
  render(vnode, container)

  // 保存引用
  messageInstance = vnode.component.exposed

  // 显示消息
  messageInstance.show()

  // 返回消息实例，可以手动关闭
  return messageInstance
}

// 导出不同类型的消息方法
export const messageService = {
  info: (content, options = {}) => showMessage({ content, type: 'info', ...options }),
  success: (content, options = {}) => showMessage({ content, type: 'success', ...options }),
  warning: (content, options = {}) => showMessage({ content, type: 'warning', ...options }),
  error: (content, options = {}) => showMessage({ content, type: 'error', ...options }),
  // 清除所有消息
  clear: () => {
    if (messageInstance) {
      messageInstance.closeImmediately()
      messageInstance = null
    }
    if (container) {
      render(null, container)
      if (container.parentNode) {
        document.body.removeChild(container)
      }
      container = null
    }
  },
}

export default messageService
