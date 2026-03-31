import { ExclamationCircleOutlined } from '@ant-design/icons-vue'
import { Modal } from 'ant-design-vue'
import { createVNode } from 'vue'

type CloseClassConfirmHandler = () => void | Promise<unknown>

interface CloseClassConfirmOptions {
  title?: string
  content?: string
  okText?: string
  cancelText?: string
  onOk?: CloseClassConfirmHandler
  onCancel?: CloseClassConfirmHandler
  onDismiss?: (event?: Event) => void
}

export function openCloseClassConfirm(options: CloseClassConfirmOptions = {}) {
  const {
    title = '结班并删除以后日程？',
    content = '是否确认对班级进行结班且结课，结班后会同步删除相关日程，被删除的日程不可恢复，请谨慎操作',
    okText = '结班并结课',
    cancelText = '仅结班',
    onOk,
    onCancel,
    onDismiss,
  } = options

  return Modal.confirm({
    title,
    centered: true,
    closable: true,
    maskClosable: false,
    keyboard: false,
    icon: createVNode(ExclamationCircleOutlined, { style: { color: '#fa8c16' } }),
    content,
    okText,
    cancelText,
    onOk() {
      return onOk?.()
    },
    onCancel(...args) {
      const [, event] = args
      if (event)
        return onDismiss?.(event)
      return onCancel?.()
    },
  })
}
