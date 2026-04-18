package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"go-migration-platform/pkg/institutionmenu"
)

type institutionMenuSeed struct {
	ParentName string
	ParentCode string
	ParentSort int
	ParentDesc string
	Children   []institutionMenuSeedChild
}

type institutionMenuSeedChild struct {
	Name        string
	Code        string
	Sort        int
	Title       string
	Description string
	Authorities []institutionMenuSeedAuthority
}

type institutionMenuSeedAuthority struct {
	Name      string
	Code      string
	Sort      int
	Weight    int
	MenuType  int
	GroupCode string
	Remark    string
}

type institutionMenuLookup struct {
	ID   int64
	Name string
}

var institutionMenuSeeds = []institutionMenuSeed{
	{
		ParentName: "品牌中心",
		ParentCode: "INST_GROUP_BRAND",
		ParentSort: 636,
		ParentDesc: "品牌触点与品牌展示相关权限。",
	},
	{
		ParentName: "招生中心",
		ParentCode: "INST_GROUP_ENROLL",
		ParentSort: 100,
		ParentDesc: "招生相关功能与权限。",
		Children: []institutionMenuSeedChild{
			{
				Name:        "招生表单",
				Code:        "INST_GROUP_ENROLL_FORM",
				Sort:        120,
				Title:       "招生表单",
				Description: "招生表单与线索管理权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "转为失效", Code: "INST_AUTH_ENROLL_FORM_INVALID", Sort: 60, Weight: 0, Remark: "可将招生表单线索转为失效。"},
					{Name: "线索导出", Code: "INST_AUTH_ENROLL_FORM_EXPORT", Sort: 70, Weight: 0, Remark: "可导出招生表单线索数据。"},
				},
			},
			{
				Name:        "意向学员",
				Code:        "INST_GROUP_ENROLL_INTENTION",
				Sort:        140,
				Title:       "意向学员",
				Description: "意向学员跟进与管理权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "编辑学员跟进状态", Code: "INST_AUTH_ENROLL_INTENTION_FOLLOW_STATUS", Sort: 110, Weight: 0, Remark: "支持编辑意向学员的跟进状态。"},
				},
			},
			{
				Name:        "公有池",
				Code:        "INST_GROUP_ENROLL_PUBLIC_POOL",
				Sort:        170,
				Title:       "公有池",
				Description: "公有池线索流转与设置权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "设置公有池", Code: "INST_AUTH_ENROLL_PUBLIC_POOL_SETTING", Sort: 10, Weight: 0, Remark: "支持开启关闭公有池并设置流转规则。"},
				},
			},
			{
				Name:        "跟进记录",
				Code:        "INST_GROUP_ENROLL_FOLLOW",
				Sort:        180,
				Title:       "跟进记录",
				Description: "跟进记录查看与处理权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "导出跟进记录", Code: "INST_AUTH_ENROLL_FOLLOW_EXPORT", Sort: 60, Weight: 0, Remark: "支持导出跟进记录。"},
				},
			},
			{
				Name:        "智能外呼",
				Code:        "INST_GROUP_ENROLL_AI_CALL",
				Sort:        220,
				Title:       "智能外呼",
				Description: "智能外呼通话记录与配置权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "查看我的外呼通话记录", Code: "INST_AUTH_ENROLL_AI_CALL_MY", Sort: 10, Weight: 0, GroupCode: "gp9000001", Remark: "仅查看采单员、前台、电话销售、副销售员、销售员、班主任为自己的外呼通话记录。"},
					{Name: "查看全部外呼通话记录", Code: "INST_AUTH_ENROLL_AI_CALL_ALL", Sort: 20, Weight: 10, GroupCode: "gp9000001", Remark: "可查看校区内全部外呼通话记录。"},
					{Name: "在PC端查看本部门及以下作为销售员的外呼通话记录", Code: "INST_AUTH_ENROLL_AI_CALL_DEPT", Sort: 30, Weight: 0, Remark: "可在PC端查看销售员为本部门及下级部门员工的外呼通话记录。"},
					{Name: "有效通话设置", Code: "INST_AUTH_ENROLL_AI_CALL_EFFECTIVE_RULE", Sort: 40, Weight: 0, Remark: "可以配置有效通话秒数定义，影响外呼通话记录列表与外呼数据报表。"},
				},
			},
		},
	},
	{
		ParentName: "教务中心",
		ParentCode: "INST_GROUP_EDU",
		ParentSort: 689,
		ParentDesc: "教务、学员与课程运营相关权限。",
	},
	{
		ParentName: "家校服务",
		ParentCode: "INST_GROUP_HOME",
		ParentSort: 500,
		ParentDesc: "家校沟通与服务相关权限。",
		Children: []institutionMenuSeedChild{
			{
				Name:        "课堂点评",
				Code:        "INST_GROUP_HOME_CLASS_REVIEW",
				Sort:        10,
				Title:       "课堂点评",
				Description: "课堂点评查看、编辑与反馈权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "查看分配给我的课堂点评", Code: "INST_AUTH_HOME_CLASS_REVIEW_MY", Sort: 20, Weight: 0, GroupCode: "gp900051", Remark: "查看上课教师、上课助教、班主任为自己的课堂点评和课堂点评明细。"},
					{Name: "写点评", Code: "INST_AUTH_HOME_CLASS_REVIEW_WRITE", Sort: 30, Weight: 0, Remark: "可以写点评并编辑课堂点评。"},
					{Name: "课评反馈查看", Code: "INST_AUTH_HOME_CLASS_REVIEW_FEEDBACK", Sort: 40, Weight: 0, Remark: "可在课堂点评和课堂点评明细中查看课评反馈。"},
				},
			},
			{
				Name:        "课后任务",
				Code:        "INST_GROUP_HOME_HOMEWORK",
				Sort:        20,
				Title:       "课后任务",
				Description: "课后任务布置、批改与查看权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "查看所有课后任务", Code: "INST_AUTH_HOME_HOMEWORK_ALL", Sort: 10, Weight: 10, GroupCode: "gp900071", Remark: "支持查看所有员工的课后任务。"},
					{Name: "仅查看我的课后任务", Code: "INST_AUTH_HOME_HOMEWORK_MY", Sort: 20, Weight: 0, GroupCode: "gp900071", Remark: "仅查看发布人为自己的课后任务。"},
					{Name: "布置批改课后任务", Code: "INST_AUTH_HOME_HOMEWORK_EDIT", Sort: 30, Weight: 0, Remark: "可以布置、编辑、删除和分享课后任务，也可批改任务和编辑批语。"},
				},
			},
			{
				Name:        "请假管理",
				Code:        "INST_GROUP_HOME_LEAVE",
				Sort:        30,
				Title:       "请假管理",
				Description: "请假查看、审批与处理权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "查看全部请假", Code: "INST_AUTH_HOME_LEAVE_ALL", Sort: 10, Weight: 10, GroupCode: "gp900030", Remark: "可以查看请假列表、请假详情和请假记录。"},
					{Name: "仅查看我的请假", Code: "INST_AUTH_HOME_LEAVE_MY", Sort: 20, Weight: 0, GroupCode: "gp900030", Remark: "仅查看学员顾问、请假课程老师或助教、学员班主任为本人的请假申请。"},
					{Name: "请假管理", Code: "INST_AUTH_HOME_LEAVE_MANAGE", Sort: 30, Weight: 0, Remark: "支持审核请假、请假代办和撤销请假。"},
				},
			},
			{
				Name:        "通知公告",
				Code:        "INST_GROUP_HOME_NOTICE",
				Sort:        40,
				Title:       "通知公告",
				Description: "通知公告查看与发布权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "查看通知", Code: "INST_AUTH_HOME_NOTICE_VIEW", Sort: 10, Weight: 0, Remark: "可以进入通知列表并查看通知内容。"},
					{Name: "通知管理", Code: "INST_AUTH_HOME_NOTICE_MANAGE", Sort: 20, Weight: 0, Remark: "可以发布通知、删除通知并进行二次提醒。"},
				},
			},
			{
				Name:        "电子相册",
				Code:        "INST_GROUP_HOME_ALBUM",
				Sort:        50,
				Title:       "电子相册",
				Description: "电子相册管理权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "电子相册", Code: "INST_AUTH_HOME_ALBUM", Sort: 10, Weight: 0, Remark: "可查看模板库，管理电子相册列表，并创建、编辑、删除电子相册。"},
				},
			},
			{
				Name:        "积分管理",
				Code:        "INST_GROUP_HOME_POINT",
				Sort:        60,
				Title:       "积分管理",
				Description: "积分规则、商城与发放权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "积分管理", Code: "INST_AUTH_HOME_POINT_MANAGE", Sort: 10, Weight: 0, Remark: "支持使用和查看积分功能。"},
					{Name: "设置自动获取积分规则", Code: "INST_AUTH_HOME_POINT_RULE", Sort: 20, Weight: 0, Remark: "拥有权限者可设置自动获取积分规则。"},
					{Name: "积分商城", Code: "INST_AUTH_HOME_POINT_MALL", Sort: 30, Weight: 0, Remark: "支持设置和管理积分礼品。"},
					{Name: "兑换商品", Code: "INST_AUTH_HOME_POINT_REDEEM", Sort: 40, Weight: 0, Remark: "支持进行积分礼品兑换。"},
					{Name: "发放积分", Code: "INST_AUTH_HOME_POINT_SEND", Sort: 50, Weight: 0, Remark: "拥有权限者可发放积分。"},
				},
			},
			{
				Name:        "打卡任务",
				Code:        "INST_GROUP_HOME_CLOCK_IN",
				Sort:        70,
				Title:       "打卡任务",
				Description: "打卡任务查看与管理权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "打卡任务", Code: "INST_AUTH_HOME_CLOCK_IN", Sort: 10, Weight: 0, Remark: "支持查看和管理打卡任务。"},
				},
			},
			{
				Name:        "学员测评",
				Code:        "INST_GROUP_HOME_ASSESSMENT",
				Sort:        80,
				Title:       "学员测评",
				Description: "测评模板与报告管理权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "测评模板管理", Code: "INST_AUTH_HOME_ASSESSMENT_TEMPLATE", Sort: 10, Weight: 0, Remark: "支持创建、编辑、删除和管理测评模板。"},
					{Name: "测评报告管理", Code: "INST_AUTH_HOME_ASSESSMENT_REPORT", Sort: 20, Weight: 0, Remark: "支持发起、编辑和删除测评报告。"},
				},
			},
			{
				Name:        "学员风采",
				Code:        "INST_GROUP_HOME_STYLE",
				Sort:        90,
				Title:       "学员风采",
				Description: "学员作品与展馆管理权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "学员作品集", Code: "INST_AUTH_HOME_STYLE_PORTFOLIO", Sort: 10, Weight: 0, Remark: "支持查看和管理学员作品集。"},
					{Name: "3D展览馆", Code: "INST_AUTH_HOME_STYLE_3D", Sort: 20, Weight: 0, Remark: "支持查看和管理3D展览馆。"},
				},
			},
			{
				Name:        "意见反馈",
				Code:        "INST_GROUP_HOME_FEEDBACK",
				Sort:        100,
				Title:       "意见反馈",
				Description: "意见反馈查看与回复权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "查看意见反馈", Code: "INST_AUTH_HOME_FEEDBACK_VIEW", Sort: 10, Weight: 0, Remark: "支持查看意见反馈列表和意见反馈详情。"},
					{Name: "回复意见反馈", Code: "INST_AUTH_HOME_FEEDBACK_REPLY", Sort: 20, Weight: 0, Remark: "支持给家长回复意见反馈。"},
				},
			},
		},
	},
	{
		ParentName: "财务中心",
		ParentCode: "INST_GROUP_FINANCE",
		ParentSort: 865,
		ParentDesc: "订单、账单、审批与财务经营相关权限。",
	},
	{
		ParentName: "数据中心",
		ParentCode: "INST_GROUP_DATA_CENTER",
		ParentSort: 916,
		ParentDesc: "经营分析与数据报表相关权限。",
	},
	{
		ParentName: "个人数据",
		ParentCode: "INST_GROUP_PERSONAL_DATA",
		ParentSort: 700,
		ParentDesc: "个人维度的数据查看权限。",
		Children: []institutionMenuSeedChild{
			{
				Name:        "销售跟进数据",
				Code:        "INST_GROUP_PERSONAL_DATA_SALES",
				Sort:        30,
				Title:       "销售跟进数据",
				Description: "销售跟进数据查看权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "可见跟进、意向、试听统计", Code: "INST_AUTH_PERSONAL_DATA_SALES_OVERVIEW", Sort: 20, Weight: 0, Remark: "支持查看个人的跟进记录、新增意向和新增试听统计数据。"},
				},
			},
			{
				Name:        "我的审批数据",
				Code:        "INST_GROUP_PERSONAL_DATA_APPROVAL",
				Sort:        40,
				Title:       "我的审批数据",
				Description: "个人审批数据查看权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "可见审批数据", Code: "INST_AUTH_PERSONAL_DATA_APPROVAL", Sort: 10, Weight: 0, Remark: "支持查看个人的审批数据。"},
				},
			},
		},
	},
	{
		ParentName: "内部管理",
		ParentCode: "INST_GROUP_INTERNAL",
		ParentSort: 800,
		ParentDesc: "内部运营与协同管理权限。",
		Children: []institutionMenuSeedChild{
			{
				Name:        "出入库管理",
				Code:        "INST_GROUP_INTERNAL_INVENTORY",
				Sort:        20,
				Title:       "出入库管理",
				Description: "出入库与库存管理权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "查看库存信息", Code: "INST_AUTH_INTERNAL_INVENTORY_STOCK", Sort: 20, Weight: 0, Remark: "可查看货品库存信息和出入库记录。"},
					{Name: "管理货品", Code: "INST_AUTH_INTERNAL_INVENTORY_GOODS", Sort: 30, Weight: 0, Remark: "可创建、编辑和删除货品。"},
					{Name: "出入库操作", Code: "INST_AUTH_INTERNAL_INVENTORY_OPERATE", Sort: 40, Weight: 0, Remark: "可进行出库和入库操作。"},
					{Name: "导入", Code: "INST_AUTH_INTERNAL_INVENTORY_IMPORT", Sort: 50, Weight: 0, Remark: "可导入货品库存。"},
				},
			},
			{
				Name:        "目标管理",
				Code:        "INST_GROUP_INTERNAL_TARGET",
				Sort:        40,
				Title:       "目标管理",
				Description: "校区目标查看与维护权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "查看目标", Code: "INST_AUTH_INTERNAL_TARGET_VIEW", Sort: 10, Weight: 0, Remark: "可查看目标和目标达成详情。"},
					{Name: "创建目标", Code: "INST_AUTH_INTERNAL_TARGET_CREATE", Sort: 20, Weight: 0, Remark: "可创建校区目标。"},
					{Name: "修改目标", Code: "INST_AUTH_INTERNAL_TARGET_UPDATE", Sort: 30, Weight: 0, Remark: "可修改校区目标。"},
					{Name: "删除目标", Code: "INST_AUTH_INTERNAL_TARGET_DELETE", Sort: 40, Weight: 0, Remark: "可删除校区目标。"},
					{Name: "导出目标达成详情", Code: "INST_AUTH_INTERNAL_TARGET_EXPORT", Sort: 50, Weight: 0, Remark: "可导出目标达成详情。"},
				},
			},
		},
	},
	{
		ParentName: "机构配置",
		ParentCode: "INST_GROUP_SETTING",
		ParentSort: 900,
		ParentDesc: "机构基础配置相关权限。",
		Children: []institutionMenuSeedChild{
			{
				Name:        "机构管理",
				Code:        "INST_GROUP_SETTING_ORG",
				Sort:        10,
				Title:       "机构管理",
				Description: "机构基础资料配置权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "校区信息管理", Code: "INST_AUTH_SETTING_ORG_INFO", Sort: 10, Weight: 0, Remark: "支持校区基础信息管理。"},
				},
			},
		},
	},
	{
		ParentName: "机构管理",
		ParentCode: "INST_GROUP_ORG_MANAGE",
		ParentSort: 920,
		ParentDesc: "机构内部组织与服务管理权限。",
		Children: []institutionMenuSeedChild{
			{
				Name:        "员工管理",
				Code:        "INST_GROUP_ORG_MANAGE_STAFF",
				Sort:        10,
				Title:       "员工管理",
				Description: "员工、部门、角色等组织管理权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "查看校区部门和员工", Code: "INST_AUTH_ORG_MANAGE_STAFF_VIEW", Sort: 10, Weight: 0, Remark: "查看校区下的部门和所有员工。"},
					{Name: "校区员工管理", Code: "INST_AUTH_ORG_MANAGE_STAFF_MANAGE", Sort: 20, Weight: 0, Remark: "在所属校区内创建新员工、批量编辑员工信息以及管理角色权限。"},
					{Name: "校区部门管理", Code: "INST_AUTH_ORG_MANAGE_DEPARTMENT_MANAGE", Sort: 30, Weight: 0, Remark: "在所属校区内新增、编辑和删除部门。"},
					{Name: "管理督办", Code: "INST_AUTH_ORG_MANAGE_SUPERVISE", Sort: 40, Weight: 0, Remark: "可以查看工作台的管理督办。"},
					{Name: "角色管理", Code: "INST_AUTH_ORG_MANAGE_ROLE_MANAGE", Sort: 50, Weight: 0, Remark: "支持创建和编辑角色。"},
					{Name: "导出员工", Code: "INST_AUTH_ORG_MANAGE_STAFF_EXPORT", Sort: 60, Weight: 0, Remark: "支持导出员工信息。"},
					{Name: "查看员工忙碌时段", Code: "INST_AUTH_ORG_MANAGE_BUSY_VIEW", Sort: 70, Weight: 0, Remark: "可查看员工忙碌时段。"},
					{Name: "管理员工忙碌时段", Code: "INST_AUTH_ORG_MANAGE_BUSY_MANAGE", Sort: 80, Weight: 0, Remark: "可设置、编辑和撤销员工忙碌时段。"},
				},
			},
			{
				Name:        "订购中心",
				Code:        "INST_GROUP_ORG_MANAGE_ORDER",
				Sort:        20,
				Title:       "订购中心",
				Description: "已购服务查看权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "查看已购服务", Code: "INST_AUTH_ORG_MANAGE_ORDER_VIEW", Sort: 10, Weight: 0, Remark: "查看已购服务剩余详情及到期时间。"},
				},
			},
			{
				Name:        "AI 风险预警",
				Code:        "INST_GROUP_ORG_MANAGE_AI_WARNING",
				Sort:        30,
				Title:       "AI 风险预警",
				Description: "AI 风险预警查看权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "查看预警通知", Code: "INST_AUTH_ORG_MANAGE_AI_WARNING_VIEW", Sort: 10, Weight: 0, Remark: "支持查看风险预警通知信息。"},
				},
			},
		},
	},
	{
		ParentName: "业务设置",
		ParentCode: "INST_GROUP_BIZ_SETTING",
		ParentSort: 940,
		ParentDesc: "业务规则与系统设置权限。",
		Children: []institutionMenuSeedChild{
			{
				Name:        "基础规则设置",
				Code:        "INST_GROUP_BIZ_SETTING_BASIC",
				Sort:        10,
				Title:       "基础规则设置",
				Description: "基础规则配置权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "基础规则管理", Code: "INST_AUTH_BIZ_SETTING_BASIC_RULE", Sort: 10, Weight: 0, Remark: "管理课程设置、点名设置、排课设置、渠道设置、教室设置、短信设置、约课设置、家校设置、出入库管理等基础规则。"},
				},
			},
			{
				Name:        "招生设置",
				Code:        "INST_GROUP_BIZ_SETTING_ENROLL",
				Sort:        20,
				Title:       "招生设置",
				Description: "招生业务相关规则配置权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "三方平台线索分配设置", Code: "INST_AUTH_BIZ_SETTING_ENROLL_THIRD_PARTY", Sort: 10, Weight: 0, Remark: "可配置三方平台线索自动分配规则，并支持查看与导出分配记录。"},
					{Name: "跟进状态设置", Code: "INST_AUTH_BIZ_SETTING_ENROLL_FOLLOW_STATUS", Sort: 20, Weight: 0, Remark: "支持查看和管理跟进状态。"},
					{Name: "学员属性设置", Code: "INST_AUTH_BIZ_SETTING_ENROLL_STUDENT_ATTR", Sort: 30, Weight: 0, Remark: "支持设置和管理学员属性。"},
					{Name: "自动升年级", Code: "INST_AUTH_BIZ_SETTING_ENROLL_AUTO_GRADE", Sort: 40, Weight: 0, Remark: "可以调整学员自动升年级配置。"},
					{Name: "学员关联人员设置", Code: "INST_AUTH_BIZ_SETTING_ENROLL_RELATION", Sort: 50, Weight: 0, Remark: "支持配置学员关联人员的启用和停用。"},
					{Name: "学员分类设置", Code: "INST_AUTH_BIZ_SETTING_ENROLL_STUDENT_CLASSIFY", Sort: 60, Weight: 0, Remark: "支持查看并编辑学员分类默认筛选条件定义。"},
					{Name: "意向学员录入设置", Code: "INST_AUTH_BIZ_SETTING_ENROLL_INTENTION_INPUT", Sort: 70, Weight: 0, Remark: "可以设置意向学员录入规则。"},
					{Name: "试听转化设置", Code: "INST_AUTH_BIZ_SETTING_ENROLL_TRIAL_CONVERT", Sort: 80, Weight: 0, Remark: "支持设置试听自动转化规则。"},
					{Name: "分配业绩设置", Code: "INST_AUTH_BIZ_SETTING_ENROLL_PERFORMANCE", Sort: 90, Weight: 0, Remark: "支持设置自动分配业绩及业绩分配规则。"},
				},
			},
			{
				Name:        "教务设置",
				Code:        "INST_GROUP_BIZ_SETTING_EDU",
				Sort:        30,
				Title:       "教务设置",
				Description: "教务规则与课程体系配置权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "组合课程", Code: "INST_AUTH_BIZ_SETTING_EDU_COMBO", Sort: 10, Weight: 0, Remark: "支持管理组合课程。"},
					{Name: "课程设置", Code: "INST_AUTH_BIZ_SETTING_EDU_COURSE", Sort: 20, Weight: 0, Remark: "支持查看并创建、编辑课程自定义属性及科目。"},
					{Name: "开启升期管理", Code: "INST_AUTH_BIZ_SETTING_EDU_PROMOTION", Sort: 30, Weight: 0, Remark: "支持开启升期管理功能。"},
					{Name: "课程类别管理", Code: "INST_AUTH_BIZ_SETTING_EDU_CATEGORY", Sort: 40, Weight: 0, Remark: "支持查看并创建、编辑课程类别。"},
					{Name: "班级属性设置", Code: "INST_AUTH_BIZ_SETTING_EDU_CLASS_ATTR", Sort: 50, Weight: 0, Remark: "支持设置和管理班级属性。"},
					{Name: "过滤节假日设置", Code: "INST_AUTH_BIZ_SETTING_EDU_HOLIDAY", Sort: 60, Weight: 0, Remark: "支持管理节假日列表并控制过滤规则。"},
					{Name: "人脸考勤设置", Code: "INST_AUTH_BIZ_SETTING_EDU_FACE", Sort: 70, Weight: 0, Remark: "支持调整人脸考勤设置项。"},
				},
			},
			{
				Name:        "家校设置",
				Code:        "INST_GROUP_BIZ_SETTING_HOME",
				Sort:        40,
				Title:       "家校设置",
				Description: "家校服务能力配置权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "课评反馈", Code: "INST_AUTH_BIZ_SETTING_HOME_REVIEW", Sort: 10, Weight: 0, Remark: "支持课评反馈开关与反馈方式配置。"},
					{Name: "评分模板设置", Code: "INST_AUTH_BIZ_SETTING_HOME_SCORE_TEMPLATE", Sort: 20, Weight: 0, Remark: "支持配置评分模板设置。"},
					{Name: "成长档案设置", Code: "INST_AUTH_BIZ_SETTING_HOME_GROWTH", Sort: 30, Weight: 0, Remark: "支持管理成长档案自定义类型。"},
					{Name: "上课提醒短信设置", Code: "INST_AUTH_BIZ_SETTING_HOME_SMS", Sort: 40, Weight: 0, Remark: "支持配置上课提醒短信。"},
				},
			},
			{
				Name:        "财务设置",
				Code:        "INST_GROUP_BIZ_SETTING_FINANCE",
				Sort:        50,
				Title:       "财务设置",
				Description: "财务规则与结算配置权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "订单标签管理", Code: "INST_AUTH_BIZ_SETTING_FINANCE_ORDER_TAG", Sort: 10, Weight: 0, Remark: "支持管理订单标签。"},
					{Name: "优惠活动设置", Code: "INST_AUTH_BIZ_SETTING_FINANCE_ACTIVITY", Sort: 20, Weight: 0, Remark: "支持设置报名优惠活动。"},
					{Name: "报名/退课设置", Code: "INST_AUTH_BIZ_SETTING_FINANCE_SIGN_OUT", Sort: 30, Weight: 0, Remark: "支持设置报名优惠模式、自定义报价单开关和退课数量规则。"},
					{Name: "退款支持使用收银宝原路退款", Code: "INST_AUTH_BIZ_SETTING_FINANCE_REFUND", Sort: 40, Weight: 0, Remark: "支持在退课、转课和储值账户退款中使用收银宝原路退款。"},
					{Name: "收款账户设置", Code: "INST_AUTH_BIZ_SETTING_FINANCE_COLLECTION", Sort: 50, Weight: 0, Remark: "支持管理收款账户。"},
					{Name: "微校/超级裂变收款账户设置", Code: "INST_AUTH_BIZ_SETTING_FINANCE_MICRO_COLLECTION", Sort: 60, Weight: 0, Remark: "支持管理微校和超级裂变的收款账户。"},
					{Name: "打印设置", Code: "INST_AUTH_BIZ_SETTING_FINANCE_PRINT", Sort: 70, Weight: 0, Remark: "可以针对收据进行自定义打印配置。"},
					{Name: "财务结账周期设置", Code: "INST_AUTH_BIZ_SETTING_FINANCE_CYCLE", Sort: 80, Weight: 0, Remark: "支持操作财务锁定设置。"},
					{Name: "编辑锁定数据", Code: "INST_AUTH_BIZ_SETTING_FINANCE_LOCKED_EDIT", Sort: 90, Weight: 0, Remark: "支持对已锁定的数据进行编辑。"},
					{Name: "收银宝授权对接管理", Code: "INST_AUTH_BIZ_SETTING_FINANCE_CASHIER", Sort: 100, Weight: 0, Remark: "支持查看并修改收银宝授权对接设置。"},
				},
			},
			{
				Name:        "数据中心设置",
				Code:        "INST_GROUP_BIZ_SETTING_DATA_CENTER",
				Sort:        60,
				Title:       "数据中心设置",
				Description: "数据中心报表配置权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "数据中心设置", Code: "INST_AUTH_BIZ_SETTING_DATA_CENTER", Sort: 10, Weight: 0, Remark: "支持把配置报表添加至数据中心各模块，并管理报表的添加、删除和排序。"},
				},
			},
			{
				Name:        "更多设置",
				Code:        "INST_GROUP_BIZ_SETTING_MORE",
				Sort:        70,
				Title:       "更多设置",
				Description: "更多业务设置权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "风险预警设置", Code: "INST_AUTH_BIZ_SETTING_MORE_WARNING", Sort: 10, Weight: 0, Remark: "支持编辑风险预警设置。"},
					{Name: "校区数据清空", Code: "INST_AUTH_BIZ_SETTING_MORE_CLEAR", Sort: 20, Weight: 0, Remark: "可以在业务设置中清空校区数据。"},
				},
			},
		},
	},
	{
		ParentName: "敏感数据",
		ParentCode: "INST_GROUP_SENSITIVE",
		ParentSort: 960,
		ParentDesc: "敏感数据查看范围控制权限。",
		Children: []institutionMenuSeedChild{
			{
				Name:        "手机号码可见",
				Code:        "INST_GROUP_SENSITIVE_PHONE",
				Sort:        10,
				Title:       "手机号码可见",
				Description: "手机号查看与呼出权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "可见员工手机号", Code: "INST_AUTH_SENSITIVE_PHONE_STAFF", Sort: 10, Weight: 0, Remark: "可以查看员工的完整手机号。"},
					{Name: "可见手机号码", Code: "INST_AUTH_SENSITIVE_PHONE_VIEW", Sort: 20, Weight: 0, MenuType: 1, Remark: "可以查看并拨打手机号码，包括数据明细报表。"},
					{Name: "本机呼出", Code: "INST_AUTH_SENSITIVE_PHONE_CALL", Sort: 30, Weight: 0, Remark: "支持快捷将手机号填写到本机拨号盘进行拨打。"},
				},
			},
			{
				Name:        "导入导出可见范围",
				Code:        "INST_GROUP_SENSITIVE_IMPORT_EXPORT",
				Sort:        20,
				Title:       "导入导出可见范围",
				Description: "导入导出记录查看范围权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "可查看所有导出记录", Code: "INST_AUTH_SENSITIVE_EXPORT_ALL", Sort: 10, Weight: 10, GroupCode: "gp5000010", Remark: "拥有页面导出权限时可查看对应页面的所有导出记录。"},
					{Name: "仅查看我的导出记录", Code: "INST_AUTH_SENSITIVE_EXPORT_MY", Sort: 20, Weight: 0, GroupCode: "gp5000010", Remark: "拥有页面导出权限时仅可查看导出人为自己的导出记录。"},
					{Name: "可查看所有导入记录", Code: "INST_AUTH_SENSITIVE_IMPORT_ALL", Sort: 30, Weight: 10, GroupCode: "gp5000030", Remark: "拥有页面导入权限时可查看对应页面的所有导入记录。"},
					{Name: "仅查看我的导入记录", Code: "INST_AUTH_SENSITIVE_IMPORT_MY", Sort: 40, Weight: 0, GroupCode: "gp5000030", Remark: "拥有页面导入权限时仅可查看导入人为自己的导入记录。"},
				},
			},
		},
	},
	{
		ParentName: "增值服务",
		ParentCode: "INST_GROUP_VALUE_ADDED",
		ParentSort: 980,
		ParentDesc: "增值服务功能查看与管理权限。",
		Children: []institutionMenuSeedChild{
			{
				Name:        "收银宝",
				Code:        "INST_GROUP_VALUE_ADDED_CASHIER",
				Sort:        10,
				Title:       "收银宝",
				Description: "收银宝服务权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "查看收银宝", Code: "INST_AUTH_VALUE_ADDED_CASHIER_VIEW", Sort: 10, Weight: 0, Remark: "支持查看收银宝。"},
				},
			},
			{
				Name:        "安心宝",
				Code:        "INST_GROUP_VALUE_ADDED_SAFE",
				Sort:        20,
				Title:       "安心宝",
				Description: "安心宝服务权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "查看安心宝", Code: "INST_AUTH_VALUE_ADDED_SAFE_VIEW", Sort: 10, Weight: 0, Remark: "支持查看安心宝。"},
				},
			},
			{
				Name:        "赛事考级",
				Code:        "INST_GROUP_VALUE_ADDED_EXAM",
				Sort:        30,
				Title:       "赛事考级",
				Description: "赛事考级服务权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "赛事考级管理", Code: "INST_AUTH_VALUE_ADDED_EXAM_MANAGE", Sort: 10, Weight: 0, Remark: "支持操作赛事考级申办、作品管理等权限。"},
				},
			},
			{
				Name:        "校宝商学",
				Code:        "INST_GROUP_VALUE_ADDED_SCHOOL_BIZ",
				Sort:        40,
				Title:       "校宝商学",
				Description: "校宝商学服务权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "查看校宝商学", Code: "INST_AUTH_VALUE_ADDED_SCHOOL_BIZ_VIEW", Sort: 10, Weight: 0, Remark: "支持查看校宝商学。"},
				},
			},
		},
	},
	{
		ParentName: "教研中心",
		ParentCode: "INST_GROUP_TEACHER_CENTER",
		ParentSort: 420,
		ParentDesc: "教研、评估与康复记录相关权限。",
		Children: []institutionMenuSeedChild{
			{
				Name:        "评估量表",
				Code:        "INST_GROUP_TEACHER_ASSESSMENT_SCALE",
				Sort:        10,
				Title:       "评估量表",
				Description: "评估量表查看与使用权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "查看评估量表", Code: "INST_AUTH_TEACHER_ASSESSMENT_SCALE_VIEW", Sort: 10, Weight: 0, Remark: "支持查看和使用评估量表。"},
				},
			},
			{
				Name:        "评估记录",
				Code:        "INST_GROUP_TEACHER_ASSESSMENT_RECORD",
				Sort:        20,
				Title:       "评估记录",
				Description: "评估记录查看权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "查看评估记录", Code: "INST_AUTH_TEACHER_ASSESSMENT_RECORD_VIEW", Sort: 10, Weight: 0, Remark: "支持查看评估记录。"},
				},
			},
			{
				Name:        "交互训练",
				Code:        "INST_GROUP_TEACHER_INTERACTIVE",
				Sort:        30,
				Title:       "交互训练",
				Description: "交互训练相关权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "查看交互训练", Code: "INST_AUTH_TEACHER_INTERACTIVE_VIEW", Sort: 10, Weight: 0, Remark: "支持查看交互训练内容。"},
				},
			},
			{
				Name:        "教案中心",
				Code:        "INST_GROUP_TEACHER_PLAN",
				Sort:        40,
				Title:       "教案中心",
				Description: "教案中心查看权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "查看教案中心", Code: "INST_AUTH_TEACHER_PLAN_VIEW", Sort: 10, Weight: 0, Remark: "支持查看教案中心内容。"},
				},
			},
			{
				Name:        "交互记录",
				Code:        "INST_GROUP_TEACHER_INTERACTIVE_RECORD",
				Sort:        50,
				Title:       "交互记录",
				Description: "交互记录查看权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "查看交互记录", Code: "INST_AUTH_TEACHER_INTERACTIVE_RECORD_VIEW", Sort: 10, Weight: 0, Remark: "支持查看交互记录。"},
				},
			},
			{
				Name:        "作业记录",
				Code:        "INST_GROUP_TEACHER_HOMEWORK_RECORD",
				Sort:        60,
				Title:       "作业记录",
				Description: "作业记录查看权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "查看作业记录", Code: "INST_AUTH_TEACHER_HOMEWORK_RECORD_VIEW", Sort: 10, Weight: 0, Remark: "支持查看作业记录。"},
				},
			},
			{
				Name:        "康复小结",
				Code:        "INST_GROUP_TEACHER_RECOVERY_SUMMARY",
				Sort:        70,
				Title:       "康复小结",
				Description: "康复小结查看权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "查看康复小结", Code: "INST_AUTH_TEACHER_RECOVERY_SUMMARY_VIEW", Sort: 10, Weight: 0, Remark: "支持查看康复小结。"},
				},
			},
			{
				Name:        "康复档案",
				Code:        "INST_GROUP_TEACHER_RECOVERY_ARCHIVE",
				Sort:        80,
				Title:       "康复档案",
				Description: "康复档案查看权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "查看康复档案", Code: "INST_AUTH_TEACHER_RECOVERY_ARCHIVE_VIEW", Sort: 10, Weight: 0, Remark: "支持查看康复档案。"},
				},
			},
		},
	},
}

func (repo *Repository) ensureInstitutionMenuCatalog(ctx context.Context) error {
	for _, parent := range institutionMenuSeeds {
		parentID, err := repo.upsertInstitutionMenuNode(ctx, institutionMenuNodeSpec{
			Name:       parent.ParentName,
			Code:       parent.ParentCode,
			PID:        0,
			Level:      1,
			Sort:       parent.ParentSort,
			Weight:     parent.ParentSort,
			MenuType:   0,
			Introduce:  strings.TrimSpace(parent.ParentDesc),
			Remark:     strings.TrimSpace(parent.ParentDesc),
			MatchNames: uniqueMenuNames(parent.ParentName),
		})
		if err != nil {
			return err
		}

		for _, child := range parent.Children {
			childID, err := repo.upsertInstitutionMenuNode(ctx, institutionMenuNodeSpec{
				Name:       child.Name,
				Code:       child.Code,
				PID:        parentID,
				Level:      2,
				Sort:       child.Sort,
				Weight:     child.Sort,
				MenuType:   0,
				Introduce:  defaultString(child.Description, child.Title),
				Remark:     defaultString(child.Description, child.Title),
				MatchNames: uniqueMenuNames(child.Name),
			})
			if err != nil {
				return err
			}

			for _, authority := range child.Authorities {
				if _, err := repo.upsertInstitutionMenuNode(ctx, institutionMenuNodeSpec{
					Name:       authority.Name,
					Code:       authority.Code,
					PID:        childID,
					Level:      3,
					Sort:       authority.Sort,
					Weight:     authority.Weight,
					MenuType:   authority.MenuType,
					GroupCode:  strings.TrimSpace(authority.GroupCode),
					Introduce:  strings.TrimSpace(authority.Remark),
					Remark:     strings.TrimSpace(authority.Remark),
					MatchNames: uniqueMenuNames(authority.Name),
				}); err != nil {
					return err
				}
			}
		}
	}

	return repo.ensureVisibleInstitutionRouteCatalog(ctx)
}

func (repo *Repository) ensureVisibleInstitutionRouteCatalog(ctx context.Context) error {
	routeMenuIDs := make([]int64, 0, 64)
	for _, group := range institutionmenu.VisibleRouteCatalog {
		groupID, err := repo.upsertInstitutionMenuNode(ctx, institutionMenuNodeSpec{
			Name:       group.Name,
			Code:       group.Code,
			PID:        0,
			Level:      1,
			Sort:       group.Sort,
			Weight:     group.Sort,
			MenuType:   0,
			URLPath:    group.Path,
			Introduce:  strings.TrimSpace(group.Introduce),
			Remark:     strings.TrimSpace(group.Introduce),
			MatchNames: uniqueMenuNames(append([]string{group.Name}, group.MatchNames...)...),
		})
		if err != nil {
			return err
		}
		routeMenuIDs = append(routeMenuIDs, groupID)

		if group.UseAsLeaf {
			continue
		}

		for _, child := range group.Children {
			childID, err := repo.upsertInstitutionMenuNode(ctx, institutionMenuNodeSpec{
				Name:       child.Name,
				Code:       child.Code,
				PID:        groupID,
				Level:      2,
				Sort:       child.Sort,
				Weight:     child.Sort,
				MenuType:   0,
				URLPath:    child.Path,
				Introduce:  strings.TrimSpace(child.Introduce),
				Remark:     strings.TrimSpace(child.Introduce),
				MatchNames: uniqueMenuNames(append([]string{child.Name}, child.MatchNames...)...),
			})
			if err != nil {
				return err
			}
			routeMenuIDs = append(routeMenuIDs, childID)
		}
	}

	return repo.ensureInstitutionAdminRoleMenus(ctx, routeMenuIDs)
}

func (repo *Repository) ensureInstitutionAdminRoleMenus(ctx context.Context, menuIDs []int64) error {
	if len(menuIDs) == 0 {
		return nil
	}

	roleRows, err := repo.db.QueryContext(ctx, `
		SELECT id
		FROM sso_role
		WHERE del_flag = 0 AND role_type = 2 AND is_admin = 1
	`)
	if err != nil {
		return err
	}
	defer roleRows.Close()

	roleIDs := make([]int64, 0, 8)
	for roleRows.Next() {
		var roleID int64
		if err := roleRows.Scan(&roleID); err != nil {
			return err
		}
		roleIDs = append(roleIDs, roleID)
	}
	if err := roleRows.Err(); err != nil {
		return err
	}

	for _, roleID := range roleIDs {
		for _, menuID := range menuIDs {
			if _, err := repo.db.ExecContext(ctx, `
				INSERT INTO sso_role_menu (role_id, menu_id)
				SELECT ?, ?
				FROM DUAL
				WHERE NOT EXISTS (
					SELECT 1
					FROM sso_role_menu
					WHERE role_id = ? AND menu_id = ?
				)
			`, roleID, menuID, roleID, menuID); err != nil {
				return err
			}
		}
	}

	return nil
}

type institutionMenuNodeSpec struct {
	Name       string
	Code       string
	PID        int64
	Level      int
	Sort       int
	Weight     int
	MenuType   int
	GroupCode  string
	URLPath    string
	Introduce  string
	Remark     string
	MatchNames []string
}

func (repo *Repository) upsertInstitutionMenuNode(ctx context.Context, spec institutionMenuNodeSpec) (int64, error) {
	if strings.TrimSpace(spec.Name) == "" {
		return 0, fmt.Errorf("menu name is required")
	}

	if id, err := repo.findMenuIDByCode(ctx, spec.Code); err != nil {
		return 0, err
	} else if id > 0 {
		return id, repo.updateInstitutionMenuNode(ctx, id, spec)
	}

	if id, err := repo.findMenuIDByExactPath(ctx, spec); err != nil {
		return 0, err
	} else if id > 0 {
		return id, repo.updateInstitutionMenuNode(ctx, id, spec)
	}

	if id, err := repo.findUniqueMenuIDByNames(ctx, spec.MatchNames, spec.Level); err != nil {
		return 0, err
	} else if id > 0 {
		return id, repo.updateInstitutionMenuNode(ctx, id, spec)
	}

	id, err := repo.insertInstitutionMenuNode(ctx, spec)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (repo *Repository) findMenuIDByCode(ctx context.Context, code string) (int64, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return 0, nil
	}

	var id int64
	err := repo.db.QueryRowContext(ctx, `
		SELECT id
		FROM sso_menu
		WHERE del_flag = 0 AND own_type = 2 AND menu_code = ?
		LIMIT 1
	`, code).Scan(&id)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return id, err
}

func (repo *Repository) findMenuIDByExactPath(ctx context.Context, spec institutionMenuNodeSpec) (int64, error) {
	switch spec.Level {
	case 1:
		return repo.findTopMenuIDByName(ctx, spec.Name)
	case 2:
		return repo.findChildMenuID(ctx, spec.PID, spec.Name)
	case 3:
		return repo.findChildMenuID(ctx, spec.PID, spec.Name)
	default:
		return 0, nil
	}
}

func (repo *Repository) findTopMenuIDByName(ctx context.Context, name string) (int64, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT id
		FROM sso_menu
		WHERE del_flag = 0 AND own_type = 2 AND IFNULL(pid, 0) = 0 AND TRIM(IFNULL(menu_name, '')) = ?
	`, strings.TrimSpace(name))
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	return readSingleMenuID(rows)
}

func (repo *Repository) findChildMenuID(ctx context.Context, pid int64, name string) (int64, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT id
		FROM sso_menu
		WHERE del_flag = 0 AND own_type = 2 AND pid = ? AND TRIM(IFNULL(menu_name, '')) = ?
	`, pid, strings.TrimSpace(name))
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	return readSingleMenuID(rows)
}

func (repo *Repository) findUniqueMenuIDByNames(ctx context.Context, names []string, level int) (int64, error) {
	candidates := make(map[int64]struct{})
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}

		rows, err := repo.db.QueryContext(ctx, `
			SELECT id
			FROM sso_menu
			WHERE del_flag = 0 AND own_type = 2 AND IFNULL(level, 0) = ? AND TRIM(IFNULL(menu_name, '')) = ?
		`, level, name)
		if err != nil {
			return 0, err
		}

		ids, err := readAllMenuIDs(rows)
		rows.Close()
		if err != nil {
			return 0, err
		}
		if len(ids) != 1 {
			continue
		}
		candidates[ids[0]] = struct{}{}
	}

	if len(candidates) != 1 {
		return 0, nil
	}

	for id := range candidates {
		return id, nil
	}
	return 0, nil
}

func (repo *Repository) updateInstitutionMenuNode(ctx context.Context, id int64, spec institutionMenuNodeSpec) error {
	_, err := repo.db.ExecContext(ctx, `
		UPDATE sso_menu
		SET menu_name = ?,
		    url_path = ?,
		    menu_code = ?,
		    pid = ?,
		    sort = ?,
		    is_system = 1,
		    introduce = ?,
		    own_type = 2,
		    level = ?,
		    weight = ?,
		    group_code = ?,
		    menu_type = ?,
		    del_flag = 0,
		    remark = ?,
		    update_id = 'system',
		    update_time = NOW()
		WHERE id = ?
	`, spec.Name, emptyToNullString(spec.URLPath), spec.Code, spec.PID, spec.Sort, spec.Introduce, spec.Level, spec.Weight, emptyToNullString(spec.GroupCode), spec.MenuType, spec.Remark, id)
	return err
}

func (repo *Repository) insertInstitutionMenuNode(ctx context.Context, spec institutionMenuNodeSpec) (int64, error) {
	result, err := repo.db.ExecContext(ctx, `
		INSERT INTO sso_menu (
			uuid, version, menu_name, url_path, menu_code, menu_type, pid, sort, is_system,
			introduce, own_type, level, weight, group_code, create_id, create_time,
			update_id, update_time, del_flag, remark
		)
		VALUES (?, 0, ?, ?, ?, ?, ?, ?, 1, ?, 2, ?, ?, ?, 'system', NOW(), 'system', NOW(), 0, ?)
	`, uuid.NewString(), spec.Name, emptyToNullString(spec.URLPath), spec.Code, spec.MenuType, spec.PID, spec.Sort, spec.Introduce, spec.Level, spec.Weight, emptyToNullString(spec.GroupCode), spec.Remark)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func readSingleMenuID(rows *sql.Rows) (int64, error) {
	ids, err := readAllMenuIDs(rows)
	if err != nil {
		return 0, err
	}
	if len(ids) != 1 {
		return 0, nil
	}
	return ids[0], nil
}

func readAllMenuIDs(rows *sql.Rows) ([]int64, error) {
	items := make([]int64, 0, 2)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		items = append(items, id)
	}
	return items, rows.Err()
}

func uniqueMenuNames(values ...string) []string {
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, exists := seen[trimmed]; exists {
			continue
		}
		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
	}
	return result
}

func defaultString(primary, fallback string) string {
	if strings.TrimSpace(primary) != "" {
		return strings.TrimSpace(primary)
	}
	return strings.TrimSpace(fallback)
}

func emptyToNullString(value string) any {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return value
}
