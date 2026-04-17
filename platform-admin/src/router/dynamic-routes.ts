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
    ],
  },
]

export default routes
