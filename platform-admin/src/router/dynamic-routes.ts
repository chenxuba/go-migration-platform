import type { RouteRecordRaw } from 'vue-router'
import { basicRouteMap } from './router-modules'
import { AccessEnum } from '~@/utils/constant'

const routes: RouteRecordRaw[] = [
  {
    path: '/platform/customers',
    redirect: '/platform/organizations',
    name: 'PlatformCustomerCenter',
    meta: {
      title: '客户管理',
      icon: 'TeamOutlined',
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
  {
    path: '/platform/system-config',
    redirect: '/platform/versions',
    name: 'PlatformSystemConfigCenter',
    meta: {
      title: '系统配置',
      icon: 'SettingOutlined',
    },
    component: basicRouteMap.RouteView,
    children: [
      {
        path: '/platform/default-roles',
        name: 'PlatformDefaultRoles',
        component: () => import('~/pages/platform/default-roles/index.vue'),
        meta: {
          title: '默认角色',
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
