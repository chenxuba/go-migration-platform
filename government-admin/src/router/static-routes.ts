import type { RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    redirect: '/government/overview',
  },
  {
    path: '/government',
    redirect: '/government/overview',
  },
  {
    path: '/government/overview',
    name: 'GovernmentOverview',
    component: () => import('~/pages/government/overview/index.vue'),
    meta: {
      title: '监管总览',
      hideInMenu: true,
      hideInBreadcrumb: true,
    },
  },
  {
    path: '/government/institutions',
    name: 'GovernmentInstitutions',
    component: () => import('~/pages/government/institutions/index.vue'),
    meta: {
      title: '机构监管',
      hideInMenu: true,
      hideInBreadcrumb: true,
    },
  },
  {
    path: '/government/supervision',
    name: 'GovernmentSupervision',
    component: () => import('~/pages/government/supervision/index.vue'),
    meta: {
      title: '督导任务',
      hideInMenu: true,
      hideInBreadcrumb: true,
    },
  },
  {
    path: '/government/accounts',
    name: 'GovernmentAccounts',
    component: () => import('~/pages/government/accounts/index.vue'),
    meta: {
      title: '账号权限',
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
