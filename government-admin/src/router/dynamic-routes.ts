import type { RouteRecordRaw } from 'vue-router'
import { basicRouteMap } from './router-modules'

const routes: RouteRecordRaw[] = [
  {
    path: '/government/dashboard',
    redirect: '/government/overview',
    name: 'GovernmentDashboardCenter',
    meta: {
      title: '监管驾驶舱',
      icon: 'DashboardOutlined',
    },
    component: basicRouteMap.RouteView,
    children: [
      {
        path: '/government/overview',
        name: 'GovernmentOverview',
        component: () => import('~/pages/government/overview/index.vue'),
        meta: {
          title: '监管总览',
        },
      },
    ],
  },
  {
    path: '/government/institution-center',
    redirect: '/government/institutions',
    name: 'GovernmentInstitutionCenter',
    meta: {
      title: '机构监管',
      icon: 'BankOutlined',
    },
    component: basicRouteMap.RouteView,
    children: [
      {
        path: '/government/institutions',
        name: 'GovernmentInstitutions',
        component: () => import('~/pages/government/institutions/index.vue'),
        meta: {
          title: '机构列表',
        },
      },
    ],
  },
  {
    path: '/government/supervision-center',
    redirect: '/government/supervision',
    name: 'GovernmentSupervisionCenter',
    meta: {
      title: '督导管理',
      icon: 'AlertOutlined',
    },
    component: basicRouteMap.RouteView,
    children: [
      {
        path: '/government/supervision',
        name: 'GovernmentSupervision',
        component: () => import('~/pages/government/supervision/index.vue'),
        meta: {
          title: '督导任务',
        },
      },
    ],
  },
  {
    path: '/government/account-center',
    redirect: '/government/accounts',
    name: 'GovernmentAccountCenter',
    meta: {
      title: '账号权限',
      icon: 'TeamOutlined',
    },
    component: basicRouteMap.RouteView,
    children: [
      {
        path: '/government/accounts',
        name: 'GovernmentAccounts',
        component: () => import('~/pages/government/accounts/index.vue'),
        meta: {
          title: '层级权限',
        },
      },
    ],
  },
]

export default routes
