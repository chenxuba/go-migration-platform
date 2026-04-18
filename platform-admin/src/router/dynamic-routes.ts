import type { RouteRecordRaw } from 'vue-router'
import { basicRouteMap } from './router-modules'
import { AccessEnum } from '~@/utils/constant'

const routes: RouteRecordRaw[] = [
  {
    path: '/platform',
    redirect: '/platform/organizations',
    name: 'PlatformCenter',
    meta: {
      title: '总控管理',
      icon: 'BankOutlined',
    },
    component: basicRouteMap.RouteView,
    children: [
      {
        path: '/platform/organizations',
        name: 'PlatformOrganizations',
        component: () => import('~/pages/platform/organizations/index.vue'),
        meta: {
          title: '机构列表',
        },
      },
      {
        path: '/platform/versions',
        name: 'PlatformVersions',
        component: () => import('~/pages/platform/versions/index.vue'),
        meta: {
          title: '版本管理',
        },
      },
      {
        path: '/platform/permissions',
        name: 'PlatformPermissions',
        component: () => import('~/pages/platform/permissions/index.vue'),
        meta: {
          title: '权限管理',
          access: [AccessEnum.systemModel_menuPermissions],
        },
      },
    ],
  },
]

export default routes
