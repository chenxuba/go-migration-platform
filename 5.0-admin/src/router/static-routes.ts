import type { RouteRecordRaw } from 'vue-router'

const Layout = () => import('~/layouts/index.vue')

export default [
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
    path: '/502',
    name: 'Error502',
    component: () => import('~/pages/exception/502.vue'),
    meta: {
      title: '系统维护中',
    },
  },
  {
    path: '/pc/face',
    name: 'Face',
    component: () => import('~/pages/edu-center/face.vue'),
    meta: {
      title: '人脸采集',
    },
  },
  // pc/import-center/starter/intentionStudent
  {
    path: '/import-center/starter/intentionStudent',
    name: 'ImportCenterStarterIntentionStudent',
    component: () => import('~/pages/import-center/intention-student/import-intention-student.vue'),
    meta: {
      title: '导入意向学员',
    },
  },
  {
    path: '/pc/import-center/import-intention-student',
    redirect: '/import-center/starter/intentionStudent',
    meta: {
      title: '导入意向学员',
    },
  },
  {
    path: '/import-center/import-intention-student-starter',
    name: 'ImportIntentionStudentStarter',
    component: () => import('~/pages/import-center/intention-student/import-intention-student-starter.vue'),
    meta: {
      title: '导入意向学员',
    },
  },
  {
    path: '/import-center/starter/order',
    name: 'ImportCenterStarterOrder',
    component: () => import('~/pages/import-center/order/import-order.vue'),
    meta: {
      title: '导入学员订单',
    },
  },
  {
    path: '/pc/import-center/starter/order',
    redirect: '/import-center/starter/order',
    meta: {
      title: '导入学员订单',
    },
  },
  {
    path: '/import-center/import-order-starter',
    name: 'ImportOrderStarter',
    component: () => import('~/pages/import-center/order/import-order-starter.vue'),
    meta: {
      title: '导入学员订单',
    },
  },
  {
    path: '/import-center/import-order/edit/:id',
    name: 'ImportOrderEdit',
    component: () => import('~/pages/import-center/order/import-order-edit.vue'),
    meta: {
      title: '导入学员订单编辑',
    },
  },
  {
    path: '/import-center/import-order/record',
    name: 'ImportOrderRecord',
    component: () => import('~/pages/import-center/order/import-order-record.vue'),
    meta: {
      title: '学员订单导入记录',
    },
  },
  {
    path: '/import-center/import-order/record/:id',
    name: 'ImportOrderRecordDetail',
    component: () => import('~/pages/import-center/order/import-order-record-detail.vue'),
    meta: {
      title: '学员订单导入记录详情',
    },
  },
  {
    path: '/import-center/import-intention-student/edit/:id',
    name: 'ImportIntentionStudentEdit',
    component: () => import('~/pages/import-center/intention-student/import-intention-student-edit.vue'),
    meta: {
      title: '导入意向学员编辑',
    },
  },
  {
    path: '/import-center/import-intention-student/record',
    name: 'ImportIntentionStudentRecord',
    component: () => import('~/pages/import-center/intention-student/import-intention-student-record.vue'),
    meta: {
      title: '意向学员导入记录',
    },
  },
  {
    path: '/import-center/import-intention-student/record/:id',
    name: 'ImportIntentionStudentRecordDetail',
    component: () => import('~/pages/import-center/intention-student/import-intention-student-record-detail.vue'),
    meta: {
      title: '意向学员导入记录详情',
    },
  },
  {
    path: '/common',
    name: 'LayoutBasicRedirect',
    component: Layout,
    redirect: '/common/redirect',
    children: [
      {
        path: '/common/redirect',
        component: () => import('~/pages/common/route-view.vue'),
        name: 'CommonRedirect',
        redirect: '/redirect',
        children: [
          {
            path: '/redirect/:path(.*)',
            name: 'RedirectPath',
            component: () => import('~/pages/common/redirect.vue'),
          },
        ],
      },

    ],
  },
  {
    path: '/:pathMatch(.*)',
    meta: {
      title: '系统维护中',
    },
    component: () => import('~/pages/exception/error.vue'),
  },
] as RouteRecordRaw[]
