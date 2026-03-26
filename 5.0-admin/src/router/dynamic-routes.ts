import type { RouteRecordRaw } from 'vue-router'
import { basicRouteMap } from './router-modules'

export default [
  {
    path: '/dashboard',
    name: 'Homepage',
    parentKeys: [],
    meta: {
      title: '首页',
      icon: 'HomeOutlined',
      // access: [AccessEnum.structureModel],
    },
    component: () => import('~/pages/dashboard/homepage/index.vue'),
  },
  {
    path: '/dashboard/analysis',
    redirect: '/dashboard/analysis',
    name: 'Dashboard',
    meta: {
      title: '品牌中心',
      icon: 'DashboardOutlined',
      // access: [AccessEnum.ADMIN], // 配置只允许管理员访问
    },
    component: basicRouteMap.RouteView,
    children: [
      {
        path: '/dashboard/analysis1',
        name: 'DashboardAnalysis1',
        component: () => import('~/pages/dashboard/analysis/index.vue'),
        meta: {
          title: '专属公众号',
          // access: [AccessEnum.ADMIN], // 配置只允许管理员访问
        },
      },
      {
        path: '/dashboard/monitor',
        name: 'DashboardMonitor',
        new: 'new',
        component: () => import('~/pages/dashboard/monitor/index.vue'),
        meta: {
          title: '专属小程序',
          // access: [AccessEnum.ADMIN], // 配置只允许管理员访问

        },
      },
      {
        path: '/dashboard/workplace',
        name: 'DashboardWorkplaces',
        component: () => import('~/pages/dashboard/workplace/index.vue'),
        meta: {
          title: '微机构',
          new: true,
          // access: [AccessEnum.ADMIN], // 配置只允许管理员访问

        },
      },
    ],
  },
  {
    path: '/enroll-center',
    redirect: '/enroll-center/basic-form',
    name: 'Form',
    meta: {
      title: '招生中心',
      icon: 'FormOutlined',
      // access: [AccessEnum.USER, AccessEnum.ADMIN], // 配置只允许用户访问
    },
    component: basicRouteMap.RouteView,
    children: [
      {
        path: '/enroll-center/self-testing-scale',
        name: 'SelfTestingScale',
        component: () => import('~/pages/enroll-center/self-testing-scale.vue'),
        meta: {
          title: '招生自测',
          // locale: 'menu.form.basic-form',
        },
      },
      {
        path: '/enroll-center/basic-form',
        name: 'FormBasic',
        component: () => import('~/pages/form/basic-form/index.vue'),
        meta: {
          title: '超级裂变',
          // locale: 'menu.form.basic-form',
        },
      },
      {
        path: '/enroll-center/intention-student',
        name: 'Intention-student',
        component: () => import('~/pages/enroll-center/intention-student.vue'),
        meta: {
          title: '意向学员',
        },
      },
      {
        path: '/enroll-center/follow-up-list',
        name: 'FollowUpList',
        component: () => import('~/pages/enroll-center/follow-up-list.vue'),
        meta: {
          title: '跟进记录',
        },
      },
      {
        path: '/enroll-center/try-listening',
        name: 'FormAdvanced2',
        component: () => import('~/pages/enroll-center/try-listening.vue'),
        meta: {
          title: '试听管理',
        },
      },
    ],
  },
  {
    path: '/edu-center',
    redirect: '/edu-center/registr-renewal',
    name: 'EduCenter',
    meta: {
      title: '教务中心',
      icon: 'WarningOutlined',
    },
    children: [
      {
        path: '/edu-center/registr-renewal',
        name: 'RegistrRenewal',
        component: () => import('~/pages/edu-center/registr-renewal.vue'),
        meta: {
          title: '报名续费',
        },
      },
      {
        path: '/edu-center/registr-renewal/:id',
        name: 'RegistrRenewalId',
        component: () => import('~/pages/edu-center/registr-renewal.vue'),
        meta: {
          title: '报名续费',
        hideInMenu: true,
          parentKeys: ['/edu-center/registr-renewal'],
        },
      },
    
      {
        path: '/edu-center/student-list',
        name: 'Exception404',
        component: () => import('~/pages/edu-center/student-list.vue'),
        meta: {
          title: '学员管理',
        },
      },
      {
        path: '/edu-center/register-read-list',
        name: 'RegisterReadList',
        component: () => import('~/pages/edu-center/register-read-list.vue'),
        meta: {
          title: '报读列表',
        },
      },
      {
        path: '/edu-center/class-list',
        name: 'ClassList',
        component: () => import('~/pages/edu-center/class-list.vue'),
        meta: {
          title: '班级管理',
        },
      },
      {
        path: '/edu-center/oneToOne',
        name: 'OneToOne',
        component: () => import('~/pages/edu-center/oneToOne.vue'),
        meta: {
          title: '一对一',
        },
      },
      {
        path: '/edu-center/timetable',
        name: 'Timetable',
        component: () => import('~/pages/edu-center/timetable.vue'),
        meta: {
          title: '课表',
        },
      },
      {
        path: '/edu-center/roll-call-list',
        name: 'RollCall',
        component: () => import('~/pages/edu-center/roll-call-list.vue'),
        meta: {
          title: '上课点名',
        },
      },
      {
        path: '/edu-center/class-record',
        name: 'ClassRecord',
        component: () => import('~/pages/edu-center/class-record.vue'),
        meta: {
          title: '上课记录',
        },
      },
      {
        path: '/edu-center/makeup-a-missedlesson',
        name: 'MakeupAmissedLesson',
        component: () => import('~/pages/edu-center/makeup-a-missedlesson.vue'),
        meta: {
          title: '补课',
        },
      },
      {
        path: '/edu-center/face-to-face',
        name: 'FaceToFace',
        component: () => import('~/pages/edu-center/face-to-face.vue'),
        meta: {
          title: '人脸考勤',
          new: true,
        },
      },

      {
        path: '/edu-center/course-list',
        name: 'CourseList',
        component: () => import('~/pages/edu-center/course-list.vue'),
        meta: {
          title: '课程商品',
        },
      },
    ],
  },
  {
    path: '/teacherCenter',
    redirect: '/teacherCenter/1',
    name: 'TeacherCenter',
    meta: {
      title: '教研中心',
      icon: 'WarningOutlined',
    },
    children: [
      {
        path: '/teacherCenter/assessment-calendar',
        name: 'AssessmentCalendar',
        component: () => import('~/pages/teacher-center/assessment-calendar.vue'),
        meta: {
          title: '评估量表',
        },
      },
      {
        path: '/teacherCenter/3',
        name: 'TeacherCenter2',
        component: () => import('~/pages/exception/403.vue'),
        meta: {
          title: '交互训练',
        },
      },
      {
        path: '/teacherCenter/0',
        name: 'TeacherCenter10',
        component: () => import('~/pages/exception/403.vue'),
        meta: {
          title: '教案中心',
        },
      },
      {
        path: '/teacherCenter/evaluationRecord',
        name: 'EvaluationRecord',
        component: () => import('~/pages/teacher-center/evaluationRecord.vue'),
        meta: {
          title: '评估记录',
        },
      },
      {
        path: '/teacherCenter/5',
        name: 'TeacherCenter5',
        component: () => import('~/pages/exception/403.vue'),
        meta: {
          title: '交互记录',
        },
      },
      {
        path: '/teacherCenter/51',
        name: 'TeacherCenter51',
        component: () => import('~/pages/exception/403.vue'),
        meta: {
          title: '作业记录',
        },
      },
      {
        path: '/teacherCenter/6',
        name: 'TeacherCenter6',
        component: () => import('~/pages/exception/403.vue'),
        meta: {
          title: '康复小结',
        },
      },
      {
        path: '/teacherCenter/7',
        name: 'TeacherCenter7',
        component: () => import('~/pages/exception/403.vue'),
        meta: {
          title: '康复档案',
        },
      },
    ],
  },
  {
    path: '/home-center',
    redirect: '/home-center/class-comment',
    name: 'HomeCenter',
    meta: {
      title: '家校服务',
      icon: 'FormOutlined',
    },
    component: basicRouteMap.RouteView,
    children: [
      {
        path: '/home-center/class-comment',
        name: 'ClassComment',
        component: () => import('~/pages/home-center/class-comment.vue'),
        meta: {
          title: '课堂点评',
        },
      },
      {
        path: '/home-center/homework',
        name: 'Homework',
        component: () => import('~/pages/home-center/homework.vue'),
        meta: {
          title: '课后任务',
        },
      },
      {
        path: '/home-center/notice-list',
        name: 'NoticeList',
        component: () => import('~/pages/home-center/notice-list.vue'),
        meta: {
          title: '通知公告',
        },
      },
      {
        path: '/home-center/leave-list',
        name: 'LeaveList',
        component: () => import('~/pages/home-center/leave-list.vue'),
        meta: {
          title: '请假管理',
        },
      },
    ],
  },
  {
    path: '/finance-center',
    redirect: '/finance-center/order-list',
    name: 'FinanceCenter',
    meta: {
      title: '财务中心',
      icon: 'FormOutlined',
    },
    component: basicRouteMap.RouteView,
    children: [
      {
        path: '/finance-center/order-list',
        name: 'OrderList',
        component: () => import('~/pages/finance-center/order-list.vue'),
        meta: {
          title: '订单管理',
        },
      },
      {
        path: '/finance-center/approve-management',
        name: 'ApproveManagement',
        component: () => import('~/pages/finance-center/approve-management.vue'),
        meta: {
          title: '审批管理',
        },
      },
      {
        path: '/form/basic3',
        name: 'Basic3',
        component: () => import('~/pages/form/basic-form/index.vue'),
        meta: {
          title: '报名优惠',
        },
      },
      {
        path: '/form/basic4',
        name: 'Basic4',
        component: () => import('~/pages/form/basic-form/index.vue'),
        meta: {
          title: '电子合同',
        },
      },
      {
        path: '/finance-center/performance-management',
        name: 'PerformanceManagement',
        component: () => import('~/pages/finance-center/performance-management.vue'),
        meta: {
          title: '业绩管理',
        },
      },
      {
        path: '/finance-center/performance-edit/:id',
        name: 'PerformanceEdit',
        component: () => import('~/pages/finance-center/performance-edit.vue'),
        meta: {
          title: '分配业绩',
          hideInMenu: true,
          parentKeys: ['/finance-center/performance-management'],
        },
      },
      {
        path: '/finance-center/bill-list',
        name: 'BillList',
        component: () => import('~/pages/finance-center/bill-list.vue'),
        meta: {
          title: '账单管理',
        },
      },
      {
        path: '/finance-center/payroll-list',
        name: 'PayrollList',
        component: () => import('~/pages/finance-center/payroll-list.vue'),
        meta: {
          title: '工资管理',
        },
      },
      {
        path: '/finance-center/income-details',
        name: 'IncomeDetails',
        component: () => import('~/pages/finance-center/income-details.vue'),
        meta: {
          title: '确认收入明细',
        },
      },
      {
        path: '/finance-center/tuition-change-record',
        name: 'TuitionChangeRecord',
        component: () => import('~/pages/finance-center/tuition-change-record.vue'),
        meta: {
          title: '学费变动记录',
        },
      },
      {
        path: '/finance-center/recharge-account',
        name: 'RechargeAccount',
        component: () => import('~/pages/finance-center/recharge-account.vue'),
        meta: {
          title: '储值账户',
        },
      },
    ],
  },
  {
    path: '/dataCenter',
    redirect: '/dataCenter/4031',
    name: 'DataCenter',
    meta: {
      title: '数据中心',
      icon: 'WarningOutlined',
    },
    children: [
      {
        path: '/dataCenter/4031',
        name: 'dataCenter',
        component: () => import('~/pages/exception/403.vue'),
        meta: {
          title: '数据大屏',
        },
      },
      {
        path: '/dataCenter/enrollmentData',
        name: 'enrollmentData',
        component: () => import('~/pages/dataCenter/enrollmentData/index.vue'),
        meta: {
          title: '招生数据',
        },
      },
      {
        path: '/dataCenter/academicAffairsData',
        name: 'academicAffairsData',
        component: () => import('~/pages/dataCenter/academicAffairsData/index.vue'),
        meta: {
          title: '教务数据',
        },
      },
      {
        path: '/dataCenter/courseHourStatistics',
        name: 'CourseHourStatistics',
        component: () => import('~/pages/dataCenter/courseHourStatistics/index.vue'),
        meta: {
          title: '课时统计',
        },
      },
      {
        path: '/exception/4036',
        name: 'Exception4036',
        component: () => import('~/pages/exception/403.vue'),
        meta: {
          title: '家校数据',
        },
      },
      {
        path: '/dataCenter/financialData/index',
        name: 'FinancialData',
        component: () => import('~/pages/dataCenter/financialData/index.vue'),
        meta: {
          title: '财务数据',
        },
      },
      {
        path: '/dataCenter/reportManagement/index',
        name: 'ReportManagement',
        component: () => import('~/pages/dataCenter/reportManagement/index.vue'),
        meta: {
          title: '报表管理',
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
    },
    children: [
      {
        path: '/internal-manage/staff-manage',
        name: 'StaffManage',
        component: () => import('~/pages/internal-manage/staff-manage.vue'),
        meta: {
          title: '员工管理',
        },
      },
      {
        path: '/internal-manage/role-manage',
        name: 'RoleManage',
        component: () => import('~/pages/internal-manage/role-manage.vue'),
        meta: {
          title: '角色管理',
        },
      },
      {
        path: '/internal-manage/businessSettings',
        name: 'BusinessSettings',
        component: () => import('~/pages/exception/403.vue'),
        meta: {
          title: '业务设置',
          isModal: true,
          noHighlight: true,
        },
      },
    ],
  },
] as RouteRecordRaw[]
