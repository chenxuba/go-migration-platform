const menuData = [
  {
    id: 2,
    parentId: null,
    title: '分析页',
    icon: 'DashboardOutlined',
    component: '/dashboard/analysis',
    path: '/dashboard/analysis',
    name: 'DashboardAnalysis',
    keepAlive: true,
    locale: 'menu.dashboard.analysis',
  },
  {
    id: 1,
    parentId: null,
    title: '仪表盘',
    icon: 'DashboardOutlined',
    component: 'RouteView',
    redirect: '/dashboard/analysis',
    path: '/dashboard',
    name: 'Dashboard',
    locale: 'menu.dashboard',
  },
  {
    id: 3,
    parentId: null,
    title: '表单页',
    icon: 'FormOutlined',
    component: 'RouteView',
    redirect: '/form/basic',
    path: '/form',
    name: 'Form',
    locale: 'menu.form',
  },

  {
    id: 19,
    parentId: null,
    title: '异常页',
    icon: 'WarningOutlined',
    component: 'RouteView',
    redirect: '/exception/403',
    path: '/exception',
    name: 'Exception',
    locale: 'menu.exception',
  },
  {
    id: 20,
    parentId: 19,
    path: '/exception/403',
    title: '403',
    name: '403',
    component: '/exception/403',
    locale: 'menu.exception.not-permission',
  },
  {
    id: 21,
    parentId: 19,
    path: '/exception/404',
    title: '404',
    name: '404',
    component: '/exception/404',
    locale: 'menu.exception.not-find',
  },
  {
    id: 22,
    parentId: 19,
    path: '/exception/500',
    title: '500',
    name: '500',
    component: '/exception/500',
    locale: 'menu.exception.server-error',
  },

  {
    id: 4,
    parentId: 3,
    title: '基础表单',
    component: '/form/basic-form/index',
    path: '/form/basic-form',
    name: 'FormBasic',
    keepAlive: false,
    locale: 'menu.form.basic-form',
  },

  {
    id: 39,
    parentId: 3,
    title: '分步表单',
    component: '/form/step-form/index',
    path: '/form/step-form',
    name: 'FormStep',
    keepAlive: false,
    locale: 'menu.form.step-form',
  },
  {
    id: 40,
    parentId: 3,
    title: '高级表单',
    component: '/form/advanced-form/index',
    path: '/form/advanced-form',
    name: 'FormAdvanced',
    keepAlive: false,
    locale: 'menu.form.advanced-form',
  },

  {
    id: 42,
    parentId: 1,
    title: '监控页',
    component: '/dashboard/monitor',
    path: '/dashboard/monitor',
    name: 'DashboardMonitor',
    keepAlive: true,
    locale: 'menu.dashboard.monitor',
  },
  {
    id: 43,
    parentId: 1,
    title: '工作台',
    component: '/dashboard/workplace',
    path: '/dashboard/workplace',
    name: 'DashboardWorkplace',
    keepAlive: true,
    locale: 'menu.dashboard.workplace',
  },

]

export const accessMenuData = [
  {
    id: 18,
    parentId: 15,
    path: '/access/admin',
    title: '管理员',
    name: 'AccessAdmin',
    component: '/access/admin',
    locale: 'menu.access.admin',
  },

]

export default eventHandler((event) => {
  const token = getHeader(event, 'Authorization')
  // eslint-disable-next-line node/prefer-global/buffer
  const username = Buffer.from(token as any, 'base64').toString('utf-8')
  return {
    code: 200,
    msg: '获取成功',
    data: [...menuData, ...(username === 'admin' ? accessMenuData : [])],
  }
})
