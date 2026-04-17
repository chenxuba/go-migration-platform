import type { RouteRecordRaw } from 'vue-router'

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
