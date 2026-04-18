import { onMounted, onUnmounted } from 'vue'
import emitter, { EVENTS } from '~@/utils/eventBus'

/**
 * 统一的学员相关列表刷新 hook
 * 在学员详情抽屉编辑/关闭后，会触发 EVENTS.REFRESH_STUDENT_LIST 事件
 * 使用该 hook 可在任意列表页面自动监听该事件并执行传入的刷新函数
 */
export function useStudentListRefresh(refreshFn: () => void) {
  onMounted(() => {
    emitter.on(EVENTS.REFRESH_STUDENT_LIST, refreshFn)
  })

  onUnmounted(() => {
    emitter.off(EVENTS.REFRESH_STUDENT_LIST, refreshFn)
  })
}

