import type { RouteRecordRaw } from 'vue-router'
import { basicRouteMap } from './router-modules'

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
    ],
  },
]

export default routes
