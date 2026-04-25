import type { RouteRecordRaw } from 'vue-router'
import { basicRouteMap } from './router-modules'
import { AccessEnum } from '~@/utils/constant'

const routes: RouteRecordRaw[] = [
  {
    path: '/platform/control',
    redirect: '/platform/control-overview',
    name: 'PlatformControlCenter',
    meta: {
      title: '平台总控',
      icon: 'ControlOutlined',
      tenantRoles: ['platform_admin', 'tenant_admin'],
    },
    component: basicRouteMap.RouteView,
    children: [
      {
        path: '/platform/control-overview',
        name: 'PlatformControlOverview',
        component: () => import('~/pages/platform/control-overview/index.vue'),
        meta: {
          title: '总控首页',
          tenantRoles: ['platform_admin', 'tenant_admin'],
        },
      },
      {
        path: '/platform/tenants',
        name: 'PlatformTenants',
        component: () => import('~/pages/platform/tenants/index.vue'),
        meta: {
          title: '租户管理',
          tenantRoles: ['platform_admin'],
        },
      },
    ],
  },
  {
    path: '/platform/customers',
    redirect: '/platform/organizations',
    name: 'PlatformCustomerCenter',
    meta: {
      title: '客户管理',
      icon: 'TeamOutlined',
      tenantRoles: ['platform_admin', 'tenant_admin'],
    },
    component: basicRouteMap.RouteView,
    children: [
      {
        path: '/platform/organizations',
        name: 'PlatformOrganizations',
        component: () => import('~/pages/platform/organizations/index.vue'),
        meta: {
          title: '机构列表',
          tenantRoles: ['platform_admin', 'tenant_admin'],
        },
      },
      {
        path: '/platform/government-accounts',
        name: 'PlatformGovernmentAccounts',
        component: () => import('~/pages/platform/government-accounts/index.vue'),
        meta: {
          title: '政府账户',
          tenantRoles: ['platform_admin'],
        },
      },
    ],
  },
  {
    path: '/internal-manage',
    redirect: '/internal-manage/staff-manage',
    name: 'InternalManage',
    meta: {
      title: '内部管理',
      icon: 'WarningOutlined',
      tenantRoles: ['platform_admin', 'tenant_admin'],
    },
    component: basicRouteMap.RouteView,
    children: [
      {
        path: '/internal-manage/staff-manage',
        name: 'StaffManage',
        component: () => import('~/pages/internal-manage/staff-manage.vue'),
        meta: {
          title: '员工管理',
          tenantRoles: ['platform_admin', 'tenant_admin'],
        },
      },
      {
        path: '/internal-manage/role-manage',
        name: 'RoleManage',
        component: () => import('~/pages/internal-manage/role-manage.vue'),
        meta: {
          title: '角色管理',
          tenantRoles: ['platform_admin', 'tenant_admin'],
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
      tenantRoles: ['platform_admin', 'tenant_admin'],
    },
    component: basicRouteMap.RouteView,
    children: [
      {
        path: '/platform/default-roles',
        name: 'PlatformDefaultRoles',
        component: () => import('~/pages/platform/default-roles/index.vue'),
        meta: {
          title: '默认角色',
          tenantRoles: ['platform_admin', 'tenant_admin'],
        },
      },
      {
        path: '/platform/versions',
        name: 'PlatformVersions',
        component: () => import('~/pages/platform/versions/index.vue'),
        meta: {
          title: '版本管理',
          tenantRoles: ['platform_admin', 'tenant_admin'],
        },
      },
      {
        path: '/platform/storage',
        name: 'PlatformStorageConfig',
        component: () => import('~/pages/platform/storage/index.vue'),
        meta: {
          title: '云存储配置',
          tenantRoles: ['platform_admin', 'tenant_admin'],
        },
      },
      {
        path: '/platform/login-templates',
        name: 'PlatformLoginTemplates',
        component: () => import('~/pages/platform/login-templates/index.vue'),
        meta: {
          title: '登录页模板',
          tenantRoles: ['platform_admin', 'tenant_admin'],
        },
      },
      {
        path: '/platform/permissions',
        name: 'PlatformPermissions',
        component: () => import('~/pages/platform/permissions/index.vue'),
        meta: {
          title: '权限管理',
          access: [AccessEnum.systemModel_menuPermissions],
          tenantRoles: ['platform_admin'],
        },
      },
    ],
  },
]

export default routes
