package institutionmenu

type RouteCatalogGroup struct {
	Name       string
	Code       string
	Sort       int
	Introduce  string
	MatchNames []string
	Children   []RouteCatalogChild
	UseAsLeaf  bool
}

type RouteCatalogChild struct {
	Name               string
	Code               string
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
		Name:       "品牌中心",
		Code:       "grp:brd",
		Sort:       20,
		Introduce:  "品牌触点与品牌展示相关权限。",
		MatchNames: []string{"品牌中心"},
		Children: []RouteCatalogChild{
			{Name: "专属公众号", Code: "page:brdOff", Sort: 10, Introduce: "专属公众号。", MatchNames: []string{"专属公众号"}, UseDirectChildren: true},
			{Name: "专属小程序", Code: "page:brdMini", Sort: 20, Introduce: "专属小程序。", MatchNames: []string{"专属小程序"}, UseDirectChildren: true},
			{Name: "微机构", Code: "page:brdMic", Sort: 30, Introduce: "微机构。", MatchNames: []string{"微机构", "微校"}, UseDirectChildren: true},
		},
	},
	{
		Name:       "招生中心",
		Code:       "grp:enr",
		Sort:       30,
		Introduce:  "招生相关功能与权限。",
		MatchNames: []string{"招生中心"},
		Children: []RouteCatalogChild{
			{Name: "招生自测", Code: "page:enrSelfTst", Sort: 10, Introduce: "招生自测。", MatchNames: []string{"招生自测"}, UseDirectChildren: false},
			{Name: "超级裂变", Code: "page:enrCamp", Sort: 20, Introduce: "超级裂变。", MatchNames: []string{"超级裂变", "招生宝"}, AggregateNodeNames: []string{"招生表单"}, UseDirectChildren: true},
			{Name: "意向学员", Code: "page:enrInt", Sort: 30, Introduce: "意向学员。", MatchNames: []string{"意向学员"}, AggregateNodeNames: []string{"公有池"}, UseDirectChildren: true},
			{Name: "跟进记录", Code: "page:enrFlw", Sort: 40, Introduce: "跟进记录。", MatchNames: []string{"跟进记录"}, UseDirectChildren: true},
			{Name: "试听管理", Code: "page:enrTrl", Sort: 50, Introduce: "试听管理。", MatchNames: []string{"试听管理", "试听记录"}, UseDirectChildren: true},
		},
	},
	{
		Name:       "教务中心",
		Code:       "grp:edu",
		Sort:       40,
		Introduce:  "教务、学员与课程运营相关权限。",
		MatchNames: []string{"教务中心"},
		Children: []RouteCatalogChild{
			{Name: "报名续费", Code: "page:eduSgn", Sort: 10, Introduce: "报名续费。", MatchNames: []string{"报名续费"}, UseDirectChildren: true},
			{Name: "学员管理", Code: "page:eduStu", Sort: 20, Introduce: "学员管理。", MatchNames: []string{"学员管理", "学员档案"}, UseDirectChildren: true},
			{Name: "报读列表", Code: "page:eduEnrLst", Sort: 30, Introduce: "报读列表。", MatchNames: []string{"报读列表"}, UseDirectChildren: true},
			{Name: "班级管理", Code: "page:eduCls", Sort: 40, Introduce: "班级管理。", MatchNames: []string{"班级管理", "班级"}, UseDirectChildren: true},
			{Name: "一对一", Code: "page:eduO2o", Sort: 50, Introduce: "一对一。", MatchNames: []string{"一对一", "1对1"}, UseDirectChildren: true},
			{Name: "课表", Code: "page:eduTbl", Sort: 60, Introduce: "课表。", MatchNames: []string{"课表"}, UseDirectChildren: true},
			{Name: "上课点名", Code: "page:eduRolCal", Sort: 70, Introduce: "上课点名。", MatchNames: []string{"上课点名"}, UseDirectChildren: true},
			{Name: "上课记录", Code: "page:eduRec", Sort: 80, Introduce: "上课记录。", MatchNames: []string{"上课记录"}, UseDirectChildren: true},
			{Name: "补课", Code: "page:eduMkp", Sort: 90, Introduce: "补课。", MatchNames: []string{"补课"}, UseDirectChildren: true},
			{Name: "人脸考勤", Code: "page:eduFac", Sort: 100, Introduce: "人脸考勤。", MatchNames: []string{"人脸考勤"}, UseDirectChildren: true},
			{Name: "课程商品", Code: "page:eduCrs", Sort: 110, Introduce: "课程商品。", MatchNames: []string{"课程商品"}, UseDirectChildren: true},
		},
	},
	{
		Name:       "教研中心",
		Code:       "grp:tchCtr",
		Sort:       50,
		Introduce:  "教研、评估与康复记录相关权限。",
		MatchNames: []string{"教研中心"},
		Children: []RouteCatalogChild{
			{Name: "量表库", Code: "page:tchScl", Sort: 10, Introduce: "量表库。", MatchNames: []string{"量表库", "评估量表"}, UseDirectChildren: false},
			{Name: "交互训练", Code: "page:tchIact", Sort: 20, Introduce: "交互训练。", MatchNames: []string{"交互训练"}, UseDirectChildren: false},
			{Name: "教案中心", Code: "page:tchPln", Sort: 30, Introduce: "教案中心。", MatchNames: []string{"教案中心"}, UseDirectChildren: false},
			{Name: "评估记录", Code: "page:tchRec", Sort: 40, Introduce: "评估记录。", MatchNames: []string{"评估记录"}, UseDirectChildren: false},
			{Name: "交互记录", Code: "page:tchIactRec", Sort: 50, Introduce: "交互记录。", MatchNames: []string{"交互记录"}, UseDirectChildren: false},
			{Name: "作业记录", Code: "page:tchHwkRec", Sort: 60, Introduce: "作业记录。", MatchNames: []string{"作业记录"}, UseDirectChildren: false},
			{Name: "康复小结", Code: "page:tchRcvSum", Sort: 70, Introduce: "康复小结。", MatchNames: []string{"康复小结"}, UseDirectChildren: false},
			{Name: "康复档案", Code: "page:tchRcvArc", Sort: 80, Introduce: "康复档案。", MatchNames: []string{"康复档案"}, UseDirectChildren: false},
		},
	},
	{
		Name:       "家校服务",
		Code:       "grp:home",
		Sort:       60,
		Introduce:  "家校沟通与服务相关权限。",
		MatchNames: []string{"家校服务"},
		Children: []RouteCatalogChild{
			{Name: "康复记录", Code: "page:homeRcv", Sort: 10, Introduce: "康复记录。", MatchNames: []string{"康复记录", "课堂点评"}, UseDirectChildren: true},
			{Name: "课后任务", Code: "page:homeHwk", Sort: 20, Introduce: "课后任务。", MatchNames: []string{"课后任务"}, UseDirectChildren: true},
			{Name: "通知公告", Code: "page:homeNtc", Sort: 30, Introduce: "通知公告。", MatchNames: []string{"通知公告"}, UseDirectChildren: true},
			{Name: "请假管理", Code: "page:homeLev", Sort: 40, Introduce: "请假管理。", MatchNames: []string{"请假管理"}, UseDirectChildren: true},
		},
	},
	{
		Name:       "财务中心",
		Code:       "grp:fin",
		Sort:       70,
		Introduce:  "订单、账单、审批与财务经营相关权限。",
		MatchNames: []string{"财务中心"},
		Children: []RouteCatalogChild{
			{Name: "订单管理", Code: "page:finOrd", Sort: 10, Introduce: "订单管理。", MatchNames: []string{"订单管理", "订单"}, UseDirectChildren: true},
			{Name: "审批管理", Code: "page:finApv", Sort: 20, Introduce: "审批管理。", MatchNames: []string{"审批管理"}, UseDirectChildren: true},
			{Name: "报名优惠", Code: "page:finDct", Sort: 30, Introduce: "报名优惠。", MatchNames: []string{"报名优惠", "报名优惠管理"}, UseDirectChildren: true},
			{Name: "业绩管理", Code: "page:finPfm", Sort: 40, Introduce: "业绩管理。", MatchNames: []string{"业绩管理"}, UseDirectChildren: true},
			{Name: "账单管理", Code: "page:finBil", Sort: 50, Introduce: "账单管理。", MatchNames: []string{"账单管理"}, UseDirectChildren: true},
			{Name: "工资管理", Code: "page:finPay", Sort: 60, Introduce: "工资管理。", MatchNames: []string{"工资管理"}, UseDirectChildren: true},
			{Name: "确认收入明细", Code: "page:finIncDtl", Sort: 70, Introduce: "确认收入明细。", MatchNames: []string{"确认收入明细", "确认收入"}, UseDirectChildren: true},
			{Name: "学费变动记录", Code: "page:finTuiChg", Sort: 80, Introduce: "学费变动记录。", MatchNames: []string{"学费变动记录"}, UseDirectChildren: true},
			{Name: "储值账户", Code: "page:finRch", Sort: 90, Introduce: "储值账户。", MatchNames: []string{"储值账户"}, UseDirectChildren: true},
		},
	},
	{
		Name:       "数据中心",
		Code:       "grp:dc",
		Sort:       80,
		Introduce:  "经营分析与数据报表相关权限。",
		MatchNames: []string{"数据中心"},
		Children: []RouteCatalogChild{
			{Name: "数据大屏", Code: "page:datScr", Sort: 10, Introduce: "数据大屏。", MatchNames: []string{"数据大屏"}, UseDirectChildren: false},
			{Name: "招生数据", Code: "page:datEnr", Sort: 20, Introduce: "招生数据。", MatchNames: []string{"招生数据"}, UseDirectChildren: false},
			{Name: "教务数据", Code: "page:datEdu", Sort: 30, Introduce: "教务数据。", MatchNames: []string{"教务数据"}, UseDirectChildren: false},
			{Name: "课时统计", Code: "page:datHrs", Sort: 40, Introduce: "课时统计。", MatchNames: []string{"课时统计"}, UseDirectChildren: false},
			{Name: "家校数据", Code: "page:datHome", Sort: 50, Introduce: "家校数据。", MatchNames: []string{"家校数据"}, UseDirectChildren: false},
			{Name: "财务数据", Code: "page:datFin", Sort: 60, Introduce: "财务数据。", MatchNames: []string{"财务数据"}, UseDirectChildren: false},
			{Name: "报表管理", Code: "page:datRpt", Sort: 70, Introduce: "报表管理。", MatchNames: []string{"报表管理"}, UseDirectChildren: false},
		},
	},
	{
		Name:       "内部管理",
		Code:       "grp:intl",
		Sort:       90,
		Introduce:  "内部运营与协同管理权限。",
		MatchNames: []string{"内部管理"},
		Children: []RouteCatalogChild{
			{Name: "员工管理", Code: "page:intlStf", Sort: 10, Introduce: "员工管理。", MatchNames: []string{"员工管理"}, UseDirectChildren: true},
			{Name: "角色管理", Code: "page:intlRole", Sort: 20, Introduce: "角色管理。", MatchNames: []string{"角色管理"}, UseDirectChildren: true},
		},
	},
	{
		Name:       "业务设置",
		Code:       "grp:bizSet",
		Sort:       100,
		Introduce:  "业务规则与系统设置权限。",
		MatchNames: []string{"业务设置"},
		Children: []RouteCatalogChild{
			{Name: "招生设置", Code: "page:setEnr", Sort: 10, Introduce: "招生设置。", MatchNames: []string{"招生设置"}, UseDirectChildren: true},
			{Name: "教务设置", Code: "page:setEdu", Sort: 20, Introduce: "教务设置。", MatchNames: []string{"教务设置"}, UseDirectChildren: true},
			{Name: "家校设置", Code: "page:setHome", Sort: 30, Introduce: "家校设置。", MatchNames: []string{"家校设置"}, UseDirectChildren: true},
			{Name: "财务设置", Code: "page:setFin", Sort: 40, Introduce: "财务设置。", MatchNames: []string{"财务设置"}, UseDirectChildren: true},
			{Name: "更多设置", Code: "page:setMor", Sort: 50, Introduce: "更多设置。", MatchNames: []string{"更多设置"}, UseDirectChildren: true},
		},
	},
}
