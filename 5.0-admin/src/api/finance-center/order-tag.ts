import { usePost } from '~/utils/request'

export interface OrderTagItem {
  id: string
  name: string
  enable: boolean
  orgOrderTagId: string
  createdTime?: string
  updatedTime?: string
}

export interface OrderTagPagedResult {
  list?: OrderTagItem[]
  total?: number
}

export function getOrderTagListPagedApi(data: {
  queryModel?: {
    enable?: boolean
  }
  sortModel?: Record<string, any>
  pageRequestModel: {
    needTotal?: boolean
    pageSize: number
    pageIndex: number
    skipCount?: number
  }
}) {
  return usePost<OrderTagPagedResult>('/api/v1/order-tags/list-paged', data)
}

export function createOrderTagApi(data: {
  name: string
}) {
  return usePost<{ id: number }>('/api/v1/order-tags/create', data)
}

export function updateOrderTagApi(data: {
  id: string | number
  name?: string
  enable?: boolean
}) {
  return usePost<boolean>('/api/v1/order-tags/update', data)
}
