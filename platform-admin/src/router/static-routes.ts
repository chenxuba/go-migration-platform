import type { RouteRecordRaw } from 'vue-router'
import { AccessEnum } from '~@/utils/constant'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    redirect: '/platform/organizations',
  },
  {
    path: '/platform',
    redirect: '/platform/organizations',
  },
  {
    path: '/platform/organizations',
    name: 'PlatformOrganizations',
    component: () => import('~/pages/platform/organizations/index.vue'),
    meta: {
      title: '机构列表',
      hideInMenu: true,
      hideInBreadcrumb: true,
    },
  },
  {
    path: '/platform/versions',
    name: 'PlatformVersions',
    component: () => import('~/pages/platform/versions/index.vue'),
    meta: {
      title: '版本管理',
      hideInMenu: true,
      hideInBreadcrumb: true,
    },
  },
  {
    path: '/platform/default-roles',
    name: 'PlatformDefaultRoles',
    component: () => import('~/pages/platform/default-roles/index.vue'),
    meta: {
      title: '默认角色',
      hideInMenu: true,
      hideInBreadcrumb: true,
    },
  },
  {
    path: '/platform/permissions',
    name: 'PlatformPermissions',
    component: () => import('~/pages/platform/permissions/index.vue'),
    meta: {
      title: '权限管理',
      hideInMenu: true,
      hideInBreadcrumb: true,
      access: [AccessEnum.systemModel_menuPermissions],
    },
  },
  {
    path: '/login',
    component: () => import('~/pages/common/login.vue'),
    meta: {
      title: '登录',
    },
  },
  {
    path: '/401',
    name: 'Error401',
    component: () => import('~/pages/exception/401.vue'),
    meta: {
      title: '授权已过期',
    },
  },
  {
    path: '/403',
    name: 'Error403',
    component: () => import('~/pages/exception/403.vue'),
    meta: {
      title: '无权访问',
    },
  },
  {
    path: '/404',
    name: 'Error404',
    component: () => import('~/pages/exception/404.vue'),
    meta: {
      title: '页面不存在',
    },
  },
  {
    path: '/500',
    name: 'Error500',
    component: () => import('~/pages/exception/500.vue'),
    meta: {
      title: '系统异常',
    },
  },
  {
    path: '/502',
    name: 'Error502',
    component: () => import('~/pages/exception/502.vue'),
    meta: {
      title: '系统维护中',
    },
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/404',
  },
]

export default routes
