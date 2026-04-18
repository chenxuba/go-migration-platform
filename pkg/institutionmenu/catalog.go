package institutionmenu

type RouteCatalogGroup struct {
	Name         string
	Code         string
	Path         string
	Sort         int
	Introduce    string
	MatchNames   []string
	Children     []RouteCatalogChild
	UseAsLeaf    bool
}

type RouteCatalogChild struct {
	Name               string
	Code               string
	Path               string
	Sort               int
	Introduce          string
	MatchNames         []string
	AggregateNodeNames []string
	AggregateNodeCodes []string
	AggregateLeafNames []string
	UseDirectChildren  bool
}

var VisibleRouteCatalog = []RouteCatalogGroup{
	{
		Name:       "首页",
		Code:       "INST_ROUTE_HOME",
		Path:       "/dashboard",
		Sort:       10,
		Introduce:  "机构首页。",
		MatchNames: []string{"首页"},
		UseAsLeaf:  true,
	},
	{
		Name:       "品牌中心",
		Code:       "INST_GROUP_BRAND",
		Path:       "/dashboard/analysis",
		Sort:       20,
		Introduce:  "品牌触点与品牌展示相关权限。",
		MatchNames: []string{"品牌中心"},
		Children: []RouteCatalogChild{
			{Name: "专属公众号", Code: "INST_ROUTE_BRAND_OFFICIAL", Path: "/dashboard/analysis1", Sort: 10, Introduce: "专属公众号。", MatchNames: []string{"专属公众号"}, UseDirectChildren: true},
			{Name: "专属小程序", Code: "INST_ROUTE_BRAND_MINIAPP", Path: "/dashboard/monitor", Sort: 20, Introduce: "专属小程序。", MatchNames: []string{"专属小程序"}, UseDirectChildren: true},
			{Name: "微机构", Code: "INST_ROUTE_BRAND_MICRO", Path: "/dashboard/workplace", Sort: 30, Introduce: "微机构。", MatchNames: []string{"微机构", "微校"}, UseDirectChildren: true},
		},
	},
	{
		Name:       "招生中心",
		Code:       "INST_GROUP_ENROLL",
		Path:       "/enroll-center",
		Sort:       30,
		Introduce:  "招生相关功能与权限。",
		MatchNames: []string{"招生中心"},
		Children: []RouteCatalogChild{
			{Name: "招生自测", Code: "INST_ROUTE_ENROLL_SELF_TEST", Path: "/enroll-center/self-testing-scale", Sort: 10, Introduce: "招生自测。", MatchNames: []string{"招生自测"}, UseDirectChildren: false},
			{Name: "超级裂变", Code: "INST_ROUTE_ENROLL_CAMPAIGN", Path: "/enroll-center/basic-form", Sort: 20, Introduce: "超级裂变。", MatchNames: []string{"超级裂变", "招生宝"}, AggregateNodeNames: []string{"招生表单"}, UseDirectChildren: true},
			{Name: "意向学员", Code: "INST_ROUTE_ENROLL_INTENTION", Path: "/enroll-center/intention-student", Sort: 30, Introduce: "意向学员。", MatchNames: []string{"意向学员"}, UseDirectChildren: true},
			{Name: "跟进记录", Code: "INST_ROUTE_ENROLL_FOLLOW", Path: "/enroll-center/follow-up-list", Sort: 40, Introduce: "跟进记录。", MatchNames: []string{"跟进记录"}, UseDirectChildren: true},
			{Name: "试听管理", Code: "INST_ROUTE_ENROLL_TRIAL", Path: "/enroll-center/try-listening", Sort: 50, Introduce: "试听管理。", MatchNames: []string{"试听管理", "试听记录"}, UseDirectChildren: true},
		},
	},
	{
		Name:       "教务中心",
		Code:       "INST_GROUP_EDU",
		Path:       "/edu-center",
		Sort:       40,
		Introduce:  "教务、学员与课程运营相关权限。",
		MatchNames: []string{"教务中心"},
		Children: []RouteCatalogChild{
			{Name: "报名续费", Code: "INST_ROUTE_EDU_SIGN", Path: "/edu-center/registr-renewal", Sort: 10, Introduce: "报名续费。", MatchNames: []string{"报名续费"}, UseDirectChildren: true},
			{Name: "学员管理", Code: "INST_ROUTE_EDU_STUDENT", Path: "/edu-center/student-list", Sort: 20, Introduce: "学员管理。", MatchNames: []string{"学员管理", "学员档案"}, UseDirectChildren: true},
			{Name: "报读列表", Code: "INST_ROUTE_EDU_ENROLL_LIST", Path: "/edu-center/register-read-list", Sort: 30, Introduce: "报读列表。", MatchNames: []string{"报读列表"}, UseDirectChildren: true},
			{Name: "班级管理", Code: "INST_ROUTE_EDU_CLASS", Path: "/edu-center/class-list", Sort: 40, Introduce: "班级管理。", MatchNames: []string{"班级管理", "班级"}, UseDirectChildren: true},
			{Name: "一对一", Code: "INST_ROUTE_EDU_ONE_TO_ONE", Path: "/edu-center/oneToOne", Sort: 50, Introduce: "一对一。", MatchNames: []string{"一对一", "1对1"}, UseDirectChildren: true},
			{Name: "课表", Code: "INST_ROUTE_EDU_TIMETABLE", Path: "/edu-center/timetable", Sort: 60, Introduce: "课表。", MatchNames: []string{"课表"}, UseDirectChildren: true},
			{Name: "上课点名", Code: "INST_ROUTE_EDU_ROLL_CALL", Path: "/edu-center/roll-call-list", Sort: 70, Introduce: "上课点名。", MatchNames: []string{"上课点名"}, UseDirectChildren: true},
			{Name: "上课记录", Code: "INST_ROUTE_EDU_RECORD", Path: "/edu-center/class-record", Sort: 80, Introduce: "上课记录。", MatchNames: []string{"上课记录"}, UseDirectChildren: true},
			{Name: "补课", Code: "INST_ROUTE_EDU_MAKEUP", Path: "/edu-center/makeup-a-missedlesson", Sort: 90, Introduce: "补课。", MatchNames: []string{"补课"}, UseDirectChildren: true},
			{Name: "人脸考勤", Code: "INST_ROUTE_EDU_FACE", Path: "/edu-center/face-to-face", Sort: 100, Introduce: "人脸考勤。", MatchNames: []string{"人脸考勤"}, UseDirectChildren: true},
			{Name: "课程商品", Code: "INST_ROUTE_EDU_COURSE", Path: "/edu-center/course-list", Sort: 110, Introduce: "课程商品。", MatchNames: []string{"课程商品"}, UseDirectChildren: true},
		},
	},
	{
		Name:       "教研中心",
		Code:       "INST_GROUP_TEACHER_CENTER",
		Path:       "/teacherCenter",
		Sort:       50,
		Introduce:  "教研、评估与康复记录相关权限。",
		MatchNames: []string{"教研中心"},
		Children: []RouteCatalogChild{
			{Name: "评估量表", Code: "INST_ROUTE_TEACHER_SCALE", Path: "/teacherCenter/assessment-calendar", Sort: 10, Introduce: "评估量表。", MatchNames: []string{"评估量表"}, UseDirectChildren: false},
			{Name: "交互训练", Code: "INST_ROUTE_TEACHER_INTERACTIVE", Path: "/teacherCenter/3", Sort: 20, Introduce: "交互训练。", MatchNames: []string{"交互训练"}, UseDirectChildren: false},
			{Name: "教案中心", Code: "INST_ROUTE_TEACHER_PLAN", Path: "/teacherCenter/0", Sort: 30, Introduce: "教案中心。", MatchNames: []string{"教案中心"}, UseDirectChildren: false},
			{Name: "评估记录", Code: "INST_ROUTE_TEACHER_RECORD", Path: "/teacherCenter/evaluationRecord", Sort: 40, Introduce: "评估记录。", MatchNames: []string{"评估记录"}, UseDirectChildren: false},
			{Name: "交互记录", Code: "INST_ROUTE_TEACHER_INTERACTIVE_RECORD", Path: "/teacherCenter/5", Sort: 50, Introduce: "交互记录。", MatchNames: []string{"交互记录"}, UseDirectChildren: false},
			{Name: "作业记录", Code: "INST_ROUTE_TEACHER_HOMEWORK_RECORD", Path: "/teacherCenter/51", Sort: 60, Introduce: "作业记录。", MatchNames: []string{"作业记录"}, UseDirectChildren: false},
			{Name: "康复小结", Code: "INST_ROUTE_TEACHER_RECOVERY_SUMMARY", Path: "/teacherCenter/6", Sort: 70, Introduce: "康复小结。", MatchNames: []string{"康复小结"}, UseDirectChildren: false},
			{Name: "康复档案", Code: "INST_ROUTE_TEACHER_RECOVERY_ARCHIVE", Path: "/teacherCenter/7", Sort: 80, Introduce: "康复档案。", MatchNames: []string{"康复档案"}, UseDirectChildren: false},
		},
	},
	{
		Name:       "家校服务",
		Code:       "INST_GROUP_HOME",
		Path:       "/home-center",
		Sort:       60,
		Introduce:  "家校沟通与服务相关权限。",
		MatchNames: []string{"家校服务"},
		Children: []RouteCatalogChild{
			{Name: "康复记录", Code: "INST_ROUTE_HOME_RECOVERY", Path: "/home-center/class-comment", Sort: 10, Introduce: "康复记录。", MatchNames: []string{"康复记录", "课堂点评"}, UseDirectChildren: true},
			{Name: "课后任务", Code: "INST_ROUTE_HOME_HOMEWORK", Path: "/home-center/homework", Sort: 20, Introduce: "课后任务。", MatchNames: []string{"课后任务"}, UseDirectChildren: true},
			{Name: "通知公告", Code: "INST_ROUTE_HOME_NOTICE", Path: "/home-center/notice-list", Sort: 30, Introduce: "通知公告。", MatchNames: []string{"通知公告"}, UseDirectChildren: true},
			{Name: "请假管理", Code: "INST_ROUTE_HOME_LEAVE", Path: "/home-center/leave-list", Sort: 40, Introduce: "请假管理。", MatchNames: []string{"请假管理"}, UseDirectChildren: true},
		},
	},
	{
		Name:       "财务中心",
		Code:       "INST_GROUP_FINANCE",
		Path:       "/finance-center",
		Sort:       70,
		Introduce:  "订单、账单、审批与财务经营相关权限。",
		MatchNames: []string{"财务中心"},
		Children: []RouteCatalogChild{
			{Name: "订单管理", Code: "INST_ROUTE_FINANCE_ORDER", Path: "/finance-center/order-list", Sort: 10, Introduce: "订单管理。", MatchNames: []string{"订单管理", "订单"}, UseDirectChildren: true},
			{Name: "审批管理", Code: "INST_ROUTE_FINANCE_APPROVAL", Path: "/finance-center/approve-management", Sort: 20, Introduce: "审批管理。", MatchNames: []string{"审批管理"}, UseDirectChildren: true},
			{Name: "报名优惠", Code: "INST_ROUTE_FINANCE_DISCOUNT", Path: "/form/basic3", Sort: 30, Introduce: "报名优惠。", MatchNames: []string{"报名优惠", "报名优惠管理"}, UseDirectChildren: true},
			{Name: "业绩管理", Code: "INST_ROUTE_FINANCE_PERFORMANCE", Path: "/finance-center/performance-management", Sort: 40, Introduce: "业绩管理。", MatchNames: []string{"业绩管理"}, UseDirectChildren: true},
			{Name: "账单管理", Code: "INST_ROUTE_FINANCE_BILL", Path: "/finance-center/bill-list", Sort: 50, Introduce: "账单管理。", MatchNames: []string{"账单管理"}, UseDirectChildren: true},
			{Name: "工资管理", Code: "INST_ROUTE_FINANCE_PAYROLL", Path: "/finance-center/payroll-list", Sort: 60, Introduce: "工资管理。", MatchNames: []string{"工资管理"}, UseDirectChildren: true},
			{Name: "确认收入明细", Code: "INST_ROUTE_FINANCE_INCOME_DETAIL", Path: "/finance-center/income-details", Sort: 70, Introduce: "确认收入明细。", MatchNames: []string{"确认收入明细", "确认收入"}, UseDirectChildren: true},
			{Name: "学费变动记录", Code: "INST_ROUTE_FINANCE_TUITION_CHANGE", Path: "/finance-center/tuition-change-record", Sort: 80, Introduce: "学费变动记录。", MatchNames: []string{"学费变动记录"}, UseDirectChildren: true},
			{Name: "储值账户", Code: "INST_ROUTE_FINANCE_RECHARGE", Path: "/finance-center/recharge-account", Sort: 90, Introduce: "储值账户。", MatchNames: []string{"储值账户"}, UseDirectChildren: true},
		},
	},
	{
		Name:       "数据中心",
		Code:       "INST_GROUP_DATA_CENTER",
		Path:       "/dataCenter",
		Sort:       80,
		Introduce:  "经营分析与数据报表相关权限。",
		MatchNames: []string{"数据中心"},
		Children: []RouteCatalogChild{
			{Name: "数据大屏", Code: "INST_ROUTE_DATA_SCREEN", Path: "/dataCenter/4031", Sort: 10, Introduce: "数据大屏。", MatchNames: []string{"数据大屏"}, UseDirectChildren: false},
			{Name: "招生数据", Code: "INST_ROUTE_DATA_ENROLL", Path: "/dataCenter/enrollmentData", Sort: 20, Introduce: "招生数据。", MatchNames: []string{"招生数据"}, UseDirectChildren: false},
			{Name: "教务数据", Code: "INST_ROUTE_DATA_EDU", Path: "/dataCenter/academicAffairsData", Sort: 30, Introduce: "教务数据。", MatchNames: []string{"教务数据"}, UseDirectChildren: false},
			{Name: "课时统计", Code: "INST_ROUTE_DATA_HOURS", Path: "/dataCenter/courseHourStatistics", Sort: 40, Introduce: "课时统计。", MatchNames: []string{"课时统计"}, UseDirectChildren: false},
			{Name: "家校数据", Code: "INST_ROUTE_DATA_HOME", Path: "/exception/4036", Sort: 50, Introduce: "家校数据。", MatchNames: []string{"家校数据"}, UseDirectChildren: false},
			{Name: "财务数据", Code: "INST_ROUTE_DATA_FINANCE", Path: "/dataCenter/financialData/index", Sort: 60, Introduce: "财务数据。", MatchNames: []string{"财务数据"}, UseDirectChildren: false},
			{Name: "报表管理", Code: "INST_ROUTE_DATA_REPORT", Path: "/dataCenter/reportManagement/index", Sort: 70, Introduce: "报表管理。", MatchNames: []string{"报表管理"}, UseDirectChildren: false},
		},
	},
	{
		Name:       "内部管理",
		Code:       "INST_GROUP_INTERNAL",
		Path:       "/internal-manage",
		Sort:       90,
		Introduce:  "内部运营与协同管理权限。",
		MatchNames: []string{"内部管理"},
		Children: []RouteCatalogChild{
			{Name: "员工管理", Code: "INST_ROUTE_INTERNAL_STAFF", Path: "/internal-manage/staff-manage", Sort: 10, Introduce: "员工管理。", MatchNames: []string{"员工管理"}, UseDirectChildren: true},
			{Name: "角色管理", Code: "INST_ROUTE_INTERNAL_ROLE", Path: "/internal-manage/role-manage", Sort: 20, Introduce: "角色管理。", MatchNames: []string{"角色管理"}, AggregateLeafNames: []string{"角色管理"}, UseDirectChildren: false},
		},
	},
	{
		Name:       "业务设置",
		Code:       "INST_GROUP_BIZ_SETTING",
		Path:       "/business-settings",
		Sort:       100,
		Introduce:  "业务规则与系统设置权限。",
		MatchNames: []string{"业务设置"},
		Children: []RouteCatalogChild{
			{Name: "招生设置", Code: "INST_ROUTE_SETTING_ENROLL", Path: "/business-settings/enrollment", Sort: 10, Introduce: "招生设置。", MatchNames: []string{"招生设置"}, UseDirectChildren: true},
			{Name: "教务设置", Code: "INST_ROUTE_SETTING_EDU", Path: "/business-settings/academic", Sort: 20, Introduce: "教务设置。", MatchNames: []string{"教务设置"}, UseDirectChildren: true},
			{Name: "家校设置", Code: "INST_ROUTE_SETTING_HOME", Path: "/business-settings/home-school", Sort: 30, Introduce: "家校设置。", MatchNames: []string{"家校设置"}, UseDirectChildren: true},
			{Name: "财务设置", Code: "INST_ROUTE_SETTING_FINANCE", Path: "/business-settings/financial", Sort: 40, Introduce: "财务设置。", MatchNames: []string{"财务设置"}, UseDirectChildren: true},
			{Name: "更多设置", Code: "INST_ROUTE_SETTING_MORE", Path: "/business-settings/more-settings", Sort: 50, Introduce: "更多设置。", MatchNames: []string{"更多设置"}, UseDirectChildren: true},
		},
	},
}
