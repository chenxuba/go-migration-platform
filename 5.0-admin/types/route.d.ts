import 'vue-router'
import type { AccessCodeLike } from '~@/constants/access'

declare module 'vue-router'{
  import type { RouteRecordRaw } from 'vue-router'

  interface RouteMeta {
    title?: string
    icon?: string
    hideInMenu?: boolean
    parentKeys?: string[]
    isIframe?: boolean
    url?: string
    hideInBreadcrumb?: boolean
    hideChildrenInMenu?: boolean
    keepAlive?: boolean
    target?: '_blank' | '_self' | '_parent'
    affix?: boolean
    id?: string | number
    parentId?: string | number | null
    access?: AccessCodeLike
    locale?: string
    parentName?: string
    parentComps?: RouteRecordRaw['component'][]
    originPath?: string
  }
}
