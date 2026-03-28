import { usePost } from '~/utils/request'

export interface ProductPackagePropertyRef {
  productPackagePropertyId: string
  productPackagePropertyValue: string
}

export interface ProductPackageItemMutation {
  productId: string
  skuId: string
  skuCount: number
  freeQuantity: number
  discountType?: number
  discountNumber: number
}

export interface ProductPackageItem {
  id: string
  productType: number
  productId: string
  productName: string
  skuId: string
  skuName: string
  lessonScope: number
  lessonType: number
  lessonMode: number
  lessonAudition: boolean
}

export interface ProductPackageProperty {
  productPackagePropertyId: string
  productPackagePropertyName?: string
  productPackagePropertyValue: string
  productPackagePropertyValueName?: string
}

export interface ProductPackageInfo {
  id: string
  name: string
  title: string
  onlineSale: boolean
  isOnlineSaleMicoSchool: boolean
  isShowMicoSchool: boolean
  orgProductPackageId: string
  editable: boolean
  isSyncOrgProductPackage: boolean
  sale: number
  totalAmount: number
  discountAmount: number
  finalAmount: number
  images: string
  subjects: Array<{ id: string, name: string }>
  extendProperties: ProductPackageProperty[]
  updatedTime?: string
  items: ProductPackageItem[]
}

export interface ProductPackagePagedResult {
  list?: ProductPackageInfo[]
  total?: number
}

export interface ProductPackageStatistics {
  totalCount?: number
  onSaleCount?: number
}

export interface ProductPackageQueryModel {
  name?: string
  searchKey?: string
  onlineSale?: boolean
  isOnlineSaleMicoSchool?: boolean
  isShowMicoSchool?: boolean
  productPackageProperties?: ProductPackagePropertyRef[]
}

export interface ProductPackageQueryParams {
  pageRequestModel: {
    pageSize: number
    pageIndex: number
    needTotal?: boolean
    skipCount?: number
  }
  sortModel?: Record<string, any>
  queryModel?: ProductPackageQueryModel
}

export interface ProductPackageCreateParams {
  name: string
  onlineSale: boolean
  isAllowEditWhenEnroll: boolean
  title: string
  images: string
  description: string
  isShowMicoSchool: boolean
  isOnlineSaleMicoSchool: boolean
  buyRule: Record<string, any>
  items: ProductPackageItemMutation[]
  subjectIds: number[]
  productPackageProperties: ProductPackagePropertyRef[]
}

export function getProductPackagePagedListApi(data: ProductPackageQueryParams) {
  return usePost<ProductPackagePagedResult>('/api/v1/product-packages/page', data)
}

export function getProductPackageStatisticsApi(data: ProductPackageQueryModel) {
  return usePost<ProductPackageStatistics>('/api/v1/product-packages/statistics', data)
}

export function createProductPackageApi(data: ProductPackageCreateParams) {
  return usePost<string>('/api/v1/product-packages/create', data)
}
