import type { RouteRecordRaw } from 'vue-router'
import { PlatformAccessEnum } from '~@/constants/access'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    redirect: '/platform/control-overview',
  },
  {
    path: '/platform',
    redirect: '/platform/control-overview',
  },
  {
    path: '/platform/control-overview',
    name: 'PlatformControlOverview',
    component: () => import('~/pages/platform/control-overview/index.vue'),
    meta: {
      title: '平台总控',
      access: [PlatformAccessEnum.platformHome],
      hideInMenu: true,
      hideInBreadcrumb: true,
    },
  },
  {
    path: '/platform/organizations',
    name: 'PlatformOrganizations',
    component: () => import('~/pages/platform/organizations/index.vue'),
    meta: {
      title: '机构列表',
      access: [PlatformAccessEnum.customerOrg],
      hideInMenu: true,
      hideInBreadcrumb: true,
    },
  },
  {
    path: '/platform/government-accounts',
    name: 'PlatformGovernmentAccounts',
    component: () => import('~/pages/platform/government-accounts/index.vue'),
    meta: {
      title: '政府账户',
      access: [PlatformAccessEnum.customerGov],
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
      access: [PlatformAccessEnum.version],
      hideInMenu: true,
      hideInBreadcrumb: true,
    },
  },
  {
    path: '/platform/tenants',
    name: 'PlatformTenants',
    component: () => import('~/pages/platform/tenants/index.vue'),
    meta: {
      title: '租户管理',
      access: [PlatformAccessEnum.platformTenant],
      hideInMenu: true,
      hideInBreadcrumb: true,
    },
  },
  {
    path: '/platform/internal-manage/staff-manage',
    name: 'StaffManage',
    component: () => import('~/pages/internal-manage/staff-manage.vue'),
    meta: {
      title: '员工管理',
      hideInMenu: true,
      access: [PlatformAccessEnum.internalStaff],
      hideInBreadcrumb: true,
    },
  },
  {
    path: '/platform/internal-manage/role-manage',
    name: 'RoleManage',
    component: () => import('~/pages/internal-manage/role-manage.vue'),
    meta: {
      title: '角色管理',
      hideInMenu: true,
      access: [PlatformAccessEnum.internalRole],
      hideInBreadcrumb: true,
    },
  },
  {
    path: '/platform/login-templates',
    name: 'PlatformLoginTemplates',
    component: () => import('~/pages/platform/login-templates/index.vue'),
    meta: {
      title: '登录页模板',
      hideInMenu: true,
      access: [PlatformAccessEnum.loginTemplate],
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
      access: [PlatformAccessEnum.defaultRole],
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
      access: [PlatformAccessEnum.permission],
    },
  },
  {
    path: '/platform/login',
    component: () => import('~/pages/common/login.vue'),
    meta: {
      title: '登录',
    },
  },
  {
    path: '/platform/login-template-preview',
    name: 'LoginTemplatePreview',
    component: () => import('~/pages/platform/login-template-preview/index.vue'),
    meta: {
      title: '登录页模板预览',
      hideInMenu: true,
      hideInBreadcrumb: true,
    },
  },
  {
    path: '/platform/redirect/:path(.*)',
    name: 'Redirect',
    component: () => import('~/pages/common/redirect.vue'),
    meta: {
      title: '刷新中',
      hideInMenu: true,
      hideInBreadcrumb: true,
    },
  },
  {
    path: '/platform/401',
    name: 'Error401',
    component: () => import('~/pages/exception/401.vue'),
    meta: {
      title: '授权已过期',
    },
  },
  {
    path: '/platform/403',
    name: 'Error403',
    component: () => import('~/pages/exception/403.vue'),
    meta: {
      title: '无权访问',
    },
  },
  {
    path: '/platform/404',
    name: 'Error404',
    component: () => import('~/pages/exception/404.vue'),
    meta: {
      title: '页面不存在',
    },
  },
  {
    path: '/platform/500',
    name: 'Error500',
    component: () => import('~/pages/exception/500.vue'),
    meta: {
      title: '系统异常',
    },
  },
  {
    path: '/platform/502',
    name: 'Error502',
    component: () => import('~/pages/exception/502.vue'),
    meta: {
      title: '系统维护中',
    },
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/platform/404',
  },
]

export default routes
