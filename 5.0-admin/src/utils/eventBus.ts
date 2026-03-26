import mitt from 'mitt'

const emitter = mitt() as any

// 定义事件常量
export const EVENTS: Record<string, string> = {
  REFRESH_STUDENT_LIST: 'REFRESH_STUDENT_LIST',
  CLOSE_LOADING_EVENT: 'CLOSE_LOADING_EVENT',
  REFRESH_DATA: 'REFRESH_DATA',
}

export default emitter as any
