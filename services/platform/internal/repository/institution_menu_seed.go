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

type institutionRouteAuthoritySeed struct {
	RouteCode   string
	Authorities []institutionRouteAuthority
}

type institutionRouteAuthority struct {
	Name       string
	Code       string
	Sort       int
	Weight     int
	MenuType   int
	GroupCode  string
	Remark     string
	MatchNames []string
	MatchCodes []string
}

type institutionMenuLookup struct {
	ID   int64
	Name string
}

type institutionMenuCatalogRow struct {
	ID    int64
	PID   int64
	Level int
	Name  string
	Code  string
	path  []string
}

type visibleRouteLeaf struct {
	Code string
	Name string
}

func listVisibleRouteLeaves() []visibleRouteLeaf {
	result := make([]visibleRouteLeaf, 0, 64)
	for _, group := range institutionmenu.VisibleRouteCatalog {
		if group.UseAsLeaf {
			result = append(result, visibleRouteLeaf{
				Code: group.Code,
				Name: group.Name,
			})
			continue
		}

		for _, child := range group.Children {
			result = append(result, visibleRouteLeaf{
				Code: child.Code,
				Name: child.Name,
			})
		}
	}
	return result
}

func buildRoutePageUseAuthority(routeCode, routeName string) institutionRouteAuthority {
	normalizedRouteCode := institutionmenu.NormalizeCode(routeCode)
	pageUseCode := normalizedRouteCode
	if strings.HasPrefix(pageUseCode, "page:") {
		pageUseCode = fmt.Sprintf("perm:%sUse", strings.TrimPrefix(pageUseCode, "page:"))
	}

	routeName = strings.TrimSpace(routeName)

	return institutionRouteAuthority{
		Name:     "页面功能访问",
		Code:     pageUseCode,
		Sort:     5,
		Weight:   0,
		MenuType: 0,
		Remark:   fmt.Sprintf("支持使用%s页面功能；未分配时进入页面仅展示“申请使用”。", routeName),
	}
}

var institutionMenuSeeds = []institutionMenuSeed{
	{
		ParentName: "品牌中心",
		ParentCode: "grp:brd",
		ParentSort: 636,
		ParentDesc: "品牌触点与品牌展示相关权限。",
	},
	{
		ParentName: "招生中心",
		ParentCode: "grp:enr",
		ParentSort: 100,
		ParentDesc: "招生相关功能与权限。",
		Children: []institutionMenuSeedChild{
			{
				Name:        "招生表单",
				Code:        "grp:enrFrm",
				Sort:        120,
				Title:       "招生表单",
				Description: "招生表单与线索管理权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "转为失效", Code: "perm:enrFrmInva", Sort: 60, Weight: 0, Remark: "可将招生表单线索转为失效。"},
					{Name: "线索导出", Code: "perm:enrFrmExp", Sort: 70, Weight: 0, Remark: "可导出招生表单线索数据。"},
				},
			},
			{
				Name:        "意向学员",
				Code:        "grp:enrInt",
				Sort:        140,
				Title:       "意向学员",
				Description: "意向学员跟进与管理权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "编辑学员跟进状态", Code: "perm:enrIntFlwSts", Sort: 110, Weight: 0, Remark: "支持编辑意向学员的跟进状态。"},
				},
			},
			{
				Name:        "公有池",
				Code:        "grp:enrPubPol",
				Sort:        170,
				Title:       "公有池",
				Description: "公有池线索流转与设置权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "设置公有池", Code: "perm:enrPubPolSet", Sort: 10, Weight: 0, Remark: "支持开启关闭公有池并设置流转规则。"},
					{Name: "批量认领", Code: "perm:enrPubPolClm", Sort: 20, Weight: 0, Remark: "支持认领公有池内的线索。"},
					{Name: "批量分配", Code: "perm:enrPubPolAsn", Sort: 30, Weight: 0, Remark: "支持分配公有池内的线索。"},
				},
			},
			{
				Name:        "跟进记录",
				Code:        "grp:enrFlw",
				Sort:        180,
				Title:       "跟进记录",
				Description: "跟进记录查看与处理权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "查看所有跟进记录", Code: "perm:enrFlwAll", Sort: 10, Weight: 10, GroupCode: "gp560001", Remark: "支持查看机构内全部跟进记录。"},
					{Name: "仅查看我的跟进记录", Code: "perm:enrFlwMy", Sort: 20, Weight: 0, GroupCode: "gp560001", Remark: "仅查看采单员、前台、电话销售、副销售员、销售员、班主任为自己的跟进记录。"},
					{Name: "在PC端查看本部门及以下作为销售员的跟进记录", Code: "perm:enrFlwDept", Sort: 30, Weight: 0, Remark: "可在PC端查看销售员为本部门及下级部门员工的跟进记录。"},
					{Name: "编辑跟进记录", Code: "perm:enrFlwEdt", Sort: 40, Weight: 0, Remark: "支持新增、编辑跟进记录以及更新回访状态。"},
					{Name: "导出跟进记录", Code: "perm:enrFlwExp", Sort: 60, Weight: 0, Remark: "支持导出跟进记录。"},
				},
			},
			{
				Name:        "智能外呼",
				Code:        "grp:enrAiCal",
				Sort:        220,
				Title:       "智能外呼",
				Description: "智能外呼通话记录与配置权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "查看我的外呼通话记录", Code: "perm:enrAiCalMy", Sort: 10, Weight: 0, GroupCode: "gp9000001", Remark: "仅查看采单员、前台、电话销售、副销售员、销售员、班主任为自己的外呼通话记录。"},
					{Name: "查看全部外呼通话记录", Code: "perm:enrAiCalAll", Sort: 20, Weight: 10, GroupCode: "gp9000001", Remark: "可查看校区内全部外呼通话记录。"},
					{Name: "在PC端查看本部门及以下作为销售员的外呼通话记录", Code: "perm:enrAiCalDept", Sort: 30, Weight: 0, Remark: "可在PC端查看销售员为本部门及下级部门员工的外呼通话记录。"},
					{Name: "有效通话设置", Code: "perm:enrAiCalEffRul", Sort: 40, Weight: 0, Remark: "可以配置有效通话秒数定义，影响外呼通话记录列表与外呼数据报表。"},
				},
			},
		},
	},
	{
		ParentName: "教务中心",
		ParentCode: "grp:edu",
		ParentSort: 689,
		ParentDesc: "教务、学员与课程运营相关权限。",
	},
	{
		ParentName: "家校服务",
		ParentCode: "grp:home",
		ParentSort: 500,
		ParentDesc: "家校沟通与服务相关权限。",
		Children: []institutionMenuSeedChild{
			{
				Name:        "康复记录",
				Code:        "grp:homeClsRvw",
				Sort:        10,
				Title:       "康复记录",
				Description: "康复记录查看、编辑与反馈权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "查看分配给我的康复记录", Code: "perm:homeClsRvwMy", Sort: 20, Weight: 0, GroupCode: "gp900051", Remark: "查看上课教师、上课助教、班主任为自己的康复记录和康复记录明细。"},
					{Name: "写康复记录", Code: "perm:homeClsRvwWrt", Sort: 30, Weight: 0, Remark: "可以写康复记录并编辑康复记录。"},
					{Name: "课评反馈查看", Code: "perm:homeClsRvwFdb", Sort: 40, Weight: 0, Remark: "可在康复记录和康复记录明细中查看课评反馈。"},
				},
			},
			{
				Name:        "课后任务",
				Code:        "grp:homeHwk",
				Sort:        20,
				Title:       "课后任务",
				Description: "课后任务布置、批改与查看权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "查看所有课后任务", Code: "perm:homeHwkAll", Sort: 10, Weight: 10, GroupCode: "gp900071", Remark: "支持查看所有员工的课后任务。"},
					{Name: "仅查看我的课后任务", Code: "perm:homeHwkMy", Sort: 20, Weight: 0, GroupCode: "gp900071", Remark: "仅查看发布人为自己的课后任务。"},
					{Name: "布置批改课后任务", Code: "perm:homeHwkEdt", Sort: 30, Weight: 0, Remark: "可以布置、编辑、删除和分享课后任务，也可批改任务和编辑批语。"},
				},
			},
			{
				Name:        "请假管理",
				Code:        "grp:homeLev",
				Sort:        30,
				Title:       "请假管理",
				Description: "请假查看、审批与处理权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "查看全部请假", Code: "perm:homeLevAll", Sort: 10, Weight: 10, GroupCode: "gp900030", Remark: "可以查看请假列表、请假详情和请假记录。"},
					{Name: "仅查看我的请假", Code: "perm:homeLevMy", Sort: 20, Weight: 0, GroupCode: "gp900030", Remark: "仅查看学员顾问、请假课程老师或助教、学员班主任为本人的请假申请。"},
					{Name: "请假管理", Code: "perm:homeLevMng", Sort: 30, Weight: 0, Remark: "支持审核请假、请假代办和撤销请假。"},
				},
			},
			{
				Name:        "通知公告",
				Code:        "grp:homeNtc",
				Sort:        40,
				Title:       "通知公告",
				Description: "通知公告查看与发布权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "查看通知", Code: "perm:homeNtcView", Sort: 10, Weight: 0, Remark: "可以进入通知列表并查看通知内容。"},
					{Name: "通知管理", Code: "perm:homeNtcMng", Sort: 20, Weight: 0, Remark: "可以发布通知、删除通知并进行二次提醒。"},
				},
			},
			{
				Name:        "电子相册",
				Code:        "grp:homeAlb",
				Sort:        50,
				Title:       "电子相册",
				Description: "电子相册管理权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "电子相册", Code: "perm:homeAlb", Sort: 10, Weight: 0, Remark: "可查看模板库，管理电子相册列表，并创建、编辑、删除电子相册。"},
				},
			},
			{
				Name:        "积分管理",
				Code:        "grp:homePnt",
				Sort:        60,
				Title:       "积分管理",
				Description: "积分规则、商城与发放权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "积分管理", Code: "perm:homePntMng", Sort: 10, Weight: 0, Remark: "支持使用和查看积分功能。"},
					{Name: "设置自动获取积分规则", Code: "perm:homePntRul", Sort: 20, Weight: 0, Remark: "拥有权限者可设置自动获取积分规则。"},
					{Name: "积分商城", Code: "perm:homePntMall", Sort: 30, Weight: 0, Remark: "支持设置和管理积分礼品。"},
					{Name: "兑换商品", Code: "perm:homePntRedeem", Sort: 40, Weight: 0, Remark: "支持进行积分礼品兑换。"},
					{Name: "发放积分", Code: "perm:homePntSnd", Sort: 50, Weight: 0, Remark: "拥有权限者可发放积分。"},
				},
			},
			{
				Name:        "打卡任务",
				Code:        "grp:homeClkIn",
				Sort:        70,
				Title:       "打卡任务",
				Description: "打卡任务查看与管理权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "打卡任务", Code: "perm:homeClkIn", Sort: 10, Weight: 0, Remark: "支持查看和管理打卡任务。"},
				},
			},
			{
				Name:        "学员测评",
				Code:        "grp:homeAsm",
				Sort:        80,
				Title:       "学员测评",
				Description: "测评模板与报告管理权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "测评模板管理", Code: "perm:homeAsmTpl", Sort: 10, Weight: 0, Remark: "支持创建、编辑、删除和管理测评模板。"},
					{Name: "测评报告管理", Code: "perm:homeAsmRpt", Sort: 20, Weight: 0, Remark: "支持发起、编辑和删除测评报告。"},
				},
			},
			{
				Name:        "学员风采",
				Code:        "grp:homeSty",
				Sort:        90,
				Title:       "学员风采",
				Description: "学员作品与展馆管理权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "学员作品集", Code: "perm:homeStyPort", Sort: 10, Weight: 0, Remark: "支持查看和管理学员作品集。"},
					{Name: "3D展览馆", Code: "perm:homeSty3d", Sort: 20, Weight: 0, Remark: "支持查看和管理3D展览馆。"},
				},
			},
			{
				Name:        "意见反馈",
				Code:        "grp:homeFdb",
				Sort:        100,
				Title:       "意见反馈",
				Description: "意见反馈查看与回复权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "查看意见反馈", Code: "perm:homeFdbView", Sort: 10, Weight: 0, Remark: "支持查看意见反馈列表和意见反馈详情。"},
					{Name: "回复意见反馈", Code: "perm:homeFdbRpy", Sort: 20, Weight: 0, Remark: "支持给家长回复意见反馈。"},
				},
			},
		},
	},
	{
		ParentName: "财务中心",
		ParentCode: "grp:fin",
		ParentSort: 865,
		ParentDesc: "订单、账单、审批与财务经营相关权限。",
	},
	{
		ParentName: "数据中心",
		ParentCode: "grp:dc",
		ParentSort: 916,
		ParentDesc: "经营分析与数据报表相关权限。",
	},
	{
		ParentName: "个人数据",
		ParentCode: "grp:psnDat",
		ParentSort: 700,
		ParentDesc: "个人维度的数据查看权限。",
		Children: []institutionMenuSeedChild{
			{
				Name:        "销售跟进数据",
				Code:        "grp:psnDatSls",
				Sort:        30,
				Title:       "销售跟进数据",
				Description: "销售跟进数据查看权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "可见跟进、意向、试听统计", Code: "perm:psnDatSlsOvw", Sort: 20, Weight: 0, Remark: "支持查看个人的跟进记录、新增意向和新增试听统计数据。"},
				},
			},
			{
				Name:        "我的审批数据",
				Code:        "grp:psnDatApv",
				Sort:        40,
				Title:       "我的审批数据",
				Description: "个人审批数据查看权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "可见审批数据", Code: "perm:psnDatApv", Sort: 10, Weight: 0, Remark: "支持查看个人的审批数据。"},
				},
			},
		},
	},
	{
		ParentName: "内部管理",
		ParentCode: "grp:intl",
		ParentSort: 800,
		ParentDesc: "内部运营与协同管理权限。",
		Children: []institutionMenuSeedChild{
			{
				Name:        "出入库管理",
				Code:        "grp:intlInv",
				Sort:        20,
				Title:       "出入库管理",
				Description: "出入库与库存管理权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "查看库存信息", Code: "perm:intlInvStock", Sort: 20, Weight: 0, Remark: "可查看货品库存信息和出入库记录。"},
					{Name: "管理货品", Code: "perm:intlInvGds", Sort: 30, Weight: 0, Remark: "可创建、编辑和删除货品。"},
					{Name: "出入库操作", Code: "perm:intlInvOpr", Sort: 40, Weight: 0, Remark: "可进行出库和入库操作。"},
					{Name: "导入", Code: "perm:intlInvImp", Sort: 50, Weight: 0, Remark: "可导入货品库存。"},
				},
			},
			{
				Name:        "目标管理",
				Code:        "grp:intlTgt",
				Sort:        40,
				Title:       "目标管理",
				Description: "校区目标查看与维护权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "查看目标", Code: "perm:intlTgtView", Sort: 10, Weight: 0, Remark: "可查看目标和目标达成详情。"},
					{Name: "创建目标", Code: "perm:intlTgtCrt", Sort: 20, Weight: 0, Remark: "可创建校区目标。"},
					{Name: "修改目标", Code: "perm:intlTgtUpd", Sort: 30, Weight: 0, Remark: "可修改校区目标。"},
					{Name: "删除目标", Code: "perm:intlTgtDel", Sort: 40, Weight: 0, Remark: "可删除校区目标。"},
					{Name: "导出目标达成详情", Code: "perm:intlTgtExp", Sort: 50, Weight: 0, Remark: "可导出目标达成详情。"},
				},
			},
		},
	},
	{
		ParentName: "机构配置",
		ParentCode: "grp:set",
		ParentSort: 900,
		ParentDesc: "机构基础配置相关权限。",
		Children: []institutionMenuSeedChild{
			{
				Name:        "机构管理",
				Code:        "grp:setOrg",
				Sort:        10,
				Title:       "机构管理",
				Description: "机构基础资料配置权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "校区信息管理", Code: "perm:setOrgInfo", Sort: 10, Weight: 0, Remark: "支持校区基础信息管理。"},
				},
			},
		},
	},
	{
		ParentName: "机构管理",
		ParentCode: "grp:orgMng",
		ParentSort: 920,
		ParentDesc: "机构内部组织与服务管理权限。",
		Children: []institutionMenuSeedChild{
			{
				Name:        "员工管理",
				Code:        "grp:orgMngStf",
				Sort:        10,
				Title:       "员工管理",
				Description: "员工、部门、角色等组织管理权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "查看校区部门和员工", Code: "perm:orgMngStfView", Sort: 10, Weight: 0, Remark: "查看校区下的部门和所有员工。"},
					{Name: "校区员工管理", Code: "perm:orgMngStfMng", Sort: 20, Weight: 0, Remark: "在所属校区内创建新员工、批量编辑员工信息以及管理角色权限。"},
					{Name: "校区部门管理", Code: "perm:orgMngDepaMng", Sort: 30, Weight: 0, Remark: "在所属校区内新增、编辑和删除部门。"},
					{Name: "管理督办", Code: "perm:orgMngSup", Sort: 40, Weight: 0, Remark: "可以查看工作台的管理督办。"},
					{Name: "导出员工", Code: "perm:orgMngStfExp", Sort: 60, Weight: 0, Remark: "支持导出员工信息。"},
					{Name: "查看员工忙碌时段", Code: "perm:orgMngBsyView", Sort: 70, Weight: 0, Remark: "可查看员工忙碌时段。"},
					{Name: "管理员工忙碌时段", Code: "perm:orgMngBsyMng", Sort: 80, Weight: 0, Remark: "可设置、编辑和撤销员工忙碌时段。"},
				},
			},
			{
				Name:        "订购中心",
				Code:        "grp:orgMngOrd",
				Sort:        20,
				Title:       "订购中心",
				Description: "已购服务查看权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "查看已购服务", Code: "perm:orgMngOrdView", Sort: 10, Weight: 0, Remark: "查看已购服务剩余详情及到期时间。"},
				},
			},
			{
				Name:        "AI 风险预警",
				Code:        "grp:orgMngAiWrn",
				Sort:        30,
				Title:       "AI 风险预警",
				Description: "AI 风险预警查看权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "查看预警通知", Code: "perm:orgMngAiWrnView", Sort: 10, Weight: 0, Remark: "支持查看风险预警通知信息。"},
				},
			},
		},
	},
	{
		ParentName: "业务设置",
		ParentCode: "grp:bizSet",
		ParentSort: 940,
		ParentDesc: "业务规则与系统设置权限。",
		Children: []institutionMenuSeedChild{
			{
				Name:        "基础规则设置",
				Code:        "grp:bizSetBsc",
				Sort:        10,
				Title:       "基础规则设置",
				Description: "基础规则配置权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "基础规则管理", Code: "perm:bizSetBscRul", Sort: 10, Weight: 0, Remark: "管理课程设置、点名设置、排课设置、渠道设置、教室设置、短信设置、约课设置、家校设置、出入库管理等基础规则。"},
				},
			},
			{
				Name:        "招生设置",
				Code:        "grp:bizSetEnr",
				Sort:        20,
				Title:       "招生设置",
				Description: "招生业务相关规则配置权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "三方平台线索分配设置", Code: "perm:bizSetEnrTp", Sort: 10, Weight: 0, Remark: "可配置三方平台线索自动分配规则，并支持查看与导出分配记录。"},
					{Name: "跟进状态设置", Code: "perm:bizSetEnrFlwSts", Sort: 20, Weight: 0, Remark: "支持查看和管理跟进状态。"},
					{Name: "学员属性设置", Code: "perm:bizSetEnrStuAtr", Sort: 30, Weight: 0, Remark: "支持设置和管理学员属性。"},
					{Name: "自动升年级", Code: "perm:bizSetEnrAtoGrd", Sort: 40, Weight: 0, Remark: "可以调整学员自动升年级配置。"},
					{Name: "学员关联人员设置", Code: "perm:bizSetEnrRel", Sort: 50, Weight: 0, Remark: "支持配置学员关联人员的启用和停用。"},
					{Name: "学员分类设置", Code: "perm:bizSetEnrStuClas", Sort: 60, Weight: 0, Remark: "支持查看并编辑学员分类默认筛选条件定义。"},
					{Name: "意向学员录入设置", Code: "perm:bizSetEnrIntInput", Sort: 70, Weight: 0, Remark: "可以设置意向学员录入规则。"},
					{Name: "试听转化设置", Code: "perm:bizSetEnrTrlCvt", Sort: 80, Weight: 0, Remark: "支持设置试听自动转化规则。"},
					{Name: "分配业绩设置", Code: "perm:bizSetEnrPfm", Sort: 90, Weight: 0, Remark: "支持设置自动分配业绩及业绩分配规则。"},
				},
			},
			{
				Name:        "教务设置",
				Code:        "grp:bizSetEdu",
				Sort:        30,
				Title:       "教务设置",
				Description: "教务规则与课程体系配置权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "组合课程", Code: "perm:bizSetEduCombo", Sort: 10, Weight: 0, Remark: "支持管理组合课程。"},
					{Name: "课程设置", Code: "perm:bizSetEduCrs", Sort: 20, Weight: 0, Remark: "支持查看并创建、编辑课程自定义属性及科目。"},
					{Name: "开启升期管理", Code: "perm:bizSetEduPrm", Sort: 30, Weight: 0, Remark: "支持开启升期管理功能。"},
					{Name: "课程类别管理", Code: "perm:bizSetEduCtg", Sort: 40, Weight: 0, Remark: "支持查看并创建、编辑课程类别。"},
					{Name: "班级属性设置", Code: "perm:bizSetEduClsAtr", Sort: 50, Weight: 0, Remark: "支持设置和管理班级属性。"},
					{Name: "过滤节假日设置", Code: "perm:bizSetEduHoli", Sort: 60, Weight: 0, Remark: "支持管理节假日列表并控制过滤规则。"},
					{Name: "人脸考勤设置", Code: "perm:bizSetEduFac", Sort: 70, Weight: 0, Remark: "支持调整人脸考勤设置项。"},
				},
			},
			{
				Name:        "家校设置",
				Code:        "grp:bizSetHome",
				Sort:        40,
				Title:       "家校设置",
				Description: "家校服务能力配置权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "课评反馈", Code: "perm:bizSetHomeRvw", Sort: 10, Weight: 0, Remark: "支持课评反馈开关与反馈方式配置。"},
					{Name: "评分模板设置", Code: "perm:bizSetHomeScrTpl", Sort: 20, Weight: 0, Remark: "支持配置评分模板设置。"},
					{Name: "成长档案设置", Code: "perm:bizSetHomeGrw", Sort: 30, Weight: 0, Remark: "支持管理成长档案自定义类型。"},
					{Name: "上课提醒短信设置", Code: "perm:bizSetHomeSms", Sort: 40, Weight: 0, Remark: "支持配置上课提醒短信。"},
				},
			},
			{
				Name:        "财务设置",
				Code:        "grp:bizSetFin",
				Sort:        50,
				Title:       "财务设置",
				Description: "财务规则与结算配置权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "订单标签管理", Code: "perm:bizSetFinOrdTag", Sort: 10, Weight: 0, Remark: "支持管理订单标签。"},
					{Name: "优惠活动设置", Code: "perm:bizSetFinActi", Sort: 20, Weight: 0, Remark: "支持设置报名优惠活动。"},
					{Name: "报名/退课设置", Code: "perm:bizSetFinSgnOut", Sort: 30, Weight: 0, Remark: "支持设置报名优惠模式、自定义报价单开关和退课数量规则。"},
					{Name: "退款支持使用收银宝原路退款", Code: "perm:bizSetFinRfd", Sort: 40, Weight: 0, Remark: "支持在退课、转课和储值账户退款中使用收银宝原路退款。"},
					{Name: "收款账户设置", Code: "perm:bizSetFinCol", Sort: 50, Weight: 0, Remark: "支持管理收款账户。"},
					{Name: "微校/超级裂变收款账户设置", Code: "perm:bizSetFinMicCol", Sort: 60, Weight: 0, Remark: "支持管理微校和超级裂变的收款账户。"},
					{Name: "打印设置", Code: "perm:bizSetFinPrt", Sort: 70, Weight: 0, Remark: "可以针对收据进行自定义打印配置。"},
					{Name: "财务结账周期设置", Code: "perm:bizSetFinCycle", Sort: 80, Weight: 0, Remark: "支持操作财务锁定设置。"},
					{Name: "编辑锁定数据", Code: "perm:bizSetFinLckEdt", Sort: 90, Weight: 0, Remark: "支持对已锁定的数据进行编辑。"},
					{Name: "收银宝授权对接管理", Code: "perm:bizSetFinCsh", Sort: 100, Weight: 0, Remark: "支持查看并修改收银宝授权对接设置。"},
				},
			},
			{
				Name:        "数据中心设置",
				Code:        "grp:bizSetDc",
				Sort:        60,
				Title:       "数据中心设置",
				Description: "数据中心报表配置权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "数据中心设置", Code: "perm:bizSetDc", Sort: 10, Weight: 0, Remark: "支持把配置报表添加至数据中心各模块，并管理报表的添加、删除和排序。"},
				},
			},
			{
				Name:        "更多设置",
				Code:        "grp:bizSetMor",
				Sort:        70,
				Title:       "更多设置",
				Description: "更多业务设置权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "风险预警设置", Code: "perm:bizSetMorWrn", Sort: 10, Weight: 0, Remark: "支持编辑风险预警设置。"},
					{Name: "校区数据清空", Code: "perm:bizSetMorClr", Sort: 20, Weight: 0, Remark: "可以在业务设置中清空校区数据。"},
				},
			},
		},
	},
	{
		ParentName: "敏感数据",
		ParentCode: "grp:sns",
		ParentSort: 960,
		ParentDesc: "敏感数据查看范围控制权限。",
		Children: []institutionMenuSeedChild{
			{
				Name:        "手机号码可见",
				Code:        "grp:snsPhn",
				Sort:        10,
				Title:       "手机号码可见",
				Description: "手机号查看与呼出权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "可见员工手机号", Code: "perm:snsPhnStf", Sort: 10, Weight: 0, Remark: "可以查看员工的完整手机号。"},
					{Name: "可见手机号码", Code: "perm:snsPhnView", Sort: 20, Weight: 0, MenuType: 1, Remark: "可以查看并拨打手机号码，包括数据明细报表。"},
					{Name: "本机呼出", Code: "perm:snsPhnCal", Sort: 30, Weight: 0, Remark: "支持快捷将手机号填写到本机拨号盘进行拨打。"},
				},
			},
			{
				Name:        "导入导出可见范围",
				Code:        "grp:snsImpExp",
				Sort:        20,
				Title:       "导入导出可见范围",
				Description: "导入导出记录查看范围权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "可查看所有导出记录", Code: "perm:snsExpAll", Sort: 10, Weight: 10, GroupCode: "gp5000010", Remark: "拥有页面导出权限时可查看对应页面的所有导出记录。"},
					{Name: "仅查看我的导出记录", Code: "perm:snsExpMy", Sort: 20, Weight: 0, GroupCode: "gp5000010", Remark: "拥有页面导出权限时仅可查看导出人为自己的导出记录。"},
					{Name: "可查看所有导入记录", Code: "perm:snsImpAll", Sort: 30, Weight: 10, GroupCode: "gp5000030", Remark: "拥有页面导入权限时可查看对应页面的所有导入记录。"},
					{Name: "仅查看我的导入记录", Code: "perm:snsImpMy", Sort: 40, Weight: 0, GroupCode: "gp5000030", Remark: "拥有页面导入权限时仅可查看导入人为自己的导入记录。"},
				},
			},
		},
	},
	{
		ParentName: "增值服务",
		ParentCode: "grp:va",
		ParentSort: 980,
		ParentDesc: "增值服务功能查看与管理权限。",
		Children: []institutionMenuSeedChild{
			{
				Name:        "收银宝",
				Code:        "grp:vaCsh",
				Sort:        10,
				Title:       "收银宝",
				Description: "收银宝服务权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "查看收银宝", Code: "perm:vaCshView", Sort: 10, Weight: 0, Remark: "支持查看收银宝。"},
				},
			},
			{
				Name:        "安心宝",
				Code:        "grp:vaSaf",
				Sort:        20,
				Title:       "安心宝",
				Description: "安心宝服务权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "查看安心宝", Code: "perm:vaSafView", Sort: 10, Weight: 0, Remark: "支持查看安心宝。"},
				},
			},
			{
				Name:        "赛事考级",
				Code:        "grp:vaExm",
				Sort:        30,
				Title:       "赛事考级",
				Description: "赛事考级服务权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "赛事考级管理", Code: "perm:vaExmMng", Sort: 10, Weight: 0, Remark: "支持操作赛事考级申办、作品管理等权限。"},
				},
			},
			{
				Name:        "校宝商学",
				Code:        "grp:vaSchBiz",
				Sort:        40,
				Title:       "校宝商学",
				Description: "校宝商学服务权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "查看校宝商学", Code: "perm:vaSchBizView", Sort: 10, Weight: 0, Remark: "支持查看校宝商学。"},
				},
			},
		},
	},
	{
		ParentName: "教研中心",
		ParentCode: "grp:tchCtr",
		ParentSort: 420,
		ParentDesc: "教研、评估与康复记录相关权限。",
		Children: []institutionMenuSeedChild{
			{
				Name:        "评估量表",
				Code:        "grp:tchAsmScl",
				Sort:        10,
				Title:       "评估量表",
				Description: "评估量表查看与使用权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "查看评估量表", Code: "perm:tchAsmSclView", Sort: 10, Weight: 0, Remark: "支持查看和使用评估量表。"},
				},
			},
			{
				Name:        "评估记录",
				Code:        "grp:tchAsmRec",
				Sort:        20,
				Title:       "评估记录",
				Description: "评估记录查看权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "查看评估记录", Code: "perm:tchAsmRecView", Sort: 10, Weight: 0, Remark: "支持查看评估记录。"},
				},
			},
			{
				Name:        "交互训练",
				Code:        "grp:tchIact",
				Sort:        30,
				Title:       "交互训练",
				Description: "交互训练相关权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "查看交互训练", Code: "perm:tchIactView", Sort: 10, Weight: 0, Remark: "支持查看交互训练内容。"},
				},
			},
			{
				Name:        "教案中心",
				Code:        "grp:tchPln",
				Sort:        40,
				Title:       "教案中心",
				Description: "教案中心查看权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "查看教案中心", Code: "perm:tchPlnView", Sort: 10, Weight: 0, Remark: "支持查看教案中心内容。"},
				},
			},
			{
				Name:        "交互记录",
				Code:        "grp:tchIactRec",
				Sort:        50,
				Title:       "交互记录",
				Description: "交互记录查看权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "查看交互记录", Code: "perm:tchIactRecView", Sort: 10, Weight: 0, Remark: "支持查看交互记录。"},
				},
			},
			{
				Name:        "作业记录",
				Code:        "grp:tchHwkRec",
				Sort:        60,
				Title:       "作业记录",
				Description: "作业记录查看权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "查看作业记录", Code: "perm:tchHwkRecView", Sort: 10, Weight: 0, Remark: "支持查看作业记录。"},
				},
			},
			{
				Name:        "康复小结",
				Code:        "grp:tchRcvSum",
				Sort:        70,
				Title:       "康复小结",
				Description: "康复小结查看权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "查看康复小结", Code: "perm:tchRcvSumView", Sort: 10, Weight: 0, Remark: "支持查看康复小结。"},
				},
			},
			{
				Name:        "康复档案",
				Code:        "grp:tchRcvArc",
				Sort:        80,
				Title:       "康复档案",
				Description: "康复档案查看权限。",
				Authorities: []institutionMenuSeedAuthority{
					{Name: "查看康复档案", Code: "perm:tchRcvArcView", Sort: 10, Weight: 0, Remark: "支持查看康复档案。"},
				},
			},
		},
	},
}

var institutionRouteAuthoritySeeds = []institutionRouteAuthoritySeed{
	{
		RouteCode: "page:enrInt",
		Authorities: []institutionRouteAuthority{
			{Name: "编辑学员跟进状态", Code: "perm:enrIntFlwSts", Sort: 10, Remark: "支持编辑意向学员的跟进状态。"},
			{Name: "查看所有的意向学员", Code: "perm:enrIntAll", Sort: 20, Weight: 10, GroupCode: "groupProspectiveStudent", Remark: "可查看校区内所有的意向学员。", MatchCodes: []string{"ViewAllProspectiveStudent"}},
			{Name: "仅查看我的意向学员", Code: "perm:enrIntMy", Sort: 30, GroupCode: "groupProspectiveStudent", Remark: "仅查看采单员、前台、电话销售、副销售员、销售员、班主任为自己的意向学员。", MatchCodes: []string{"OnlyViewMyProspectiveStudent"}},
			{Name: "在PC端查看本部门及以下作为销售员的意向学员", Code: "perm:enrIntDept", Sort: 40, Remark: "可在PC端查看销售员为本部门及下级部门员工的意向学员。", MatchCodes: []string{"PCProspectiveStudent"}},
			{Name: "管理意向学员", Code: "perm:enrIntMng", Sort: 50, Remark: "可以在报名续费、意向学员页面新增和编辑意向学员。", MatchCodes: []string{"ManagementProspectiveStudent"}},
			{Name: "意向学员详情", Code: "perm:enrIntDtl", Sort: 60, Remark: "支持查看意向学员详情。", MatchCodes: []string{"ProspectiveStudentDetail"}},
			{Name: "意向学员渠道编辑", Code: "perm:enrIntChnEdt", Sort: 70, Remark: "可以编辑意向学员渠道字段。", MatchCodes: []string{"ProspectiveStudentEdit"}},
			{Name: "导入意向学员", Code: "perm:enrIntImp", Sort: 80, Remark: "支持批量导入意向学员数据。", MatchCodes: []string{"ImportProspectiveStudent"}},
			{Name: "导出意向学员", Code: "perm:enrIntExp", Sort: 90, Remark: "可在PC意向学员列表中导出学员数据。", MatchCodes: []string{"ExportProspectiveStudent"}},
			{Name: "分配销售员", Code: "perm:enrIntAsnSls", Sort: 100, Remark: "支持批量分配和单独分配销售员。", MatchCodes: []string{"AssignSalespeople"}},
			{Name: "批量转入公有池", Code: "perm:enrIntTranPubPol", Sort: 110, Remark: "支持将意向学员批量转入公有池。", MatchCodes: []string{"TransferPublicPool"}},
		},
	},
	{
		RouteCode: "page:enrFlw",
		Authorities: []institutionRouteAuthority{
			{Name: "查看所有跟进记录", Code: "perm:enrFlwAll", Sort: 10, Weight: 10, GroupCode: "gp560001", Remark: "支持查看机构内全部跟进记录。", MatchNames: []string{"查看所有跟进记录"}},
			{Name: "仅查看我的跟进记录", Code: "perm:enrFlwMy", Sort: 20, GroupCode: "gp560001", Remark: "仅查看采单员、前台、电话销售、副销售员、销售员、班主任为自己的跟进记录。", MatchNames: []string{"仅查看我的跟进记录"}},
			{Name: "在PC端查看本部门及以下作为销售员的跟进记录", Code: "perm:enrFlwDept", Sort: 30, Remark: "可在PC端查看销售员为本部门及下级部门员工的跟进记录。", MatchNames: []string{"在PC端查看本部门及以下作为销售员的跟进记录"}},
			{Name: "编辑跟进记录", Code: "perm:enrFlwEdt", Sort: 40, Remark: "支持新增、编辑跟进记录以及更新回访状态。", MatchNames: []string{"编辑跟进记录"}},
			{Name: "导出跟进记录", Code: "perm:enrFlwExp", Sort: 50, Remark: "支持导出跟进记录。", MatchNames: []string{"导出跟进记录"}},
		},
	},
	{
		RouteCode: "page:eduSgn",
		Authorities: []institutionRouteAuthority{
			{Name: "可为所有学员报名续费", Code: "perm:eduSgnAllStu", Sort: 10, GroupCode: "gp200010", Remark: "可为所有学员报名续费。", MatchNames: []string{"可为所有学员报名续费"}, MatchCodes: []string{"Canregisterandrenewforallstudents"}},
			{Name: "仅可为我的学员报名续费", Code: "perm:eduSgnMyStu", Sort: 20, GroupCode: "gp200010", Remark: "仅可为采单员、前台、电话销售、副销售员、销售员、班主任、学管师、顾问为自己的学员报名续费。", MatchNames: []string{"仅可为我的学员报名续费"}, MatchCodes: []string{"Onlymystudentscanapplyforrenewal"}},
			{Name: "整单优惠设置", Code: "perm:eduSgnOrdDct", Sort: 30, Remark: "支持设置整单优惠。", MatchNames: []string{"整单优惠设置"}, MatchCodes: []string{"Wholeorderdiscountsetting"}},
		},
	},
	{
		RouteCode: "page:eduCls",
		Authorities: []institutionRouteAuthority{
			{Name: "查看所有的班级", Code: "perm:eduClsAll", Sort: 10, Weight: 10, GroupCode: "groupViewallclasses", Remark: "可查看校区内所有的班级。", MatchCodes: []string{"Viewallclasses"}},
			{Name: "仅查看我的班级", Code: "perm:eduClsMy", Sort: 20, GroupCode: "groupViewallclasses", Remark: "仅查看我作为班主任的班级。", MatchCodes: []string{"Onlyviewmyclass"}},
			{Name: "导入班级", Code: "perm:eduClsImp", Sort: 30, Remark: "支持导入班级。", MatchCodes: []string{"ImportClass"}},
			{Name: "新建/编辑/结班/调整班级学员", Code: "perm:eduClsMngWthStus", Sort: 40, Remark: "可以新建班级、编辑班级、结班，并调整班级学员。", MatchCodes: []string{"NewEditCloseAdjustClassStudents"}},
			{Name: "新建/编辑班级以及结班操作", Code: "perm:eduClsMng", Sort: 50, Remark: "可以新建班级、编辑班级并进行结班操作。", MatchCodes: []string{"NewEditClassandClosingOperations"}},
			{Name: "调整班级学员", Code: "perm:eduClsAdjustStus", Sort: 60, Remark: "可以对班级新增学员、移出学员并调整学员至其他班级。", MatchCodes: []string{"Adjustclassstudents"}},
			{Name: "编辑满班人数", Code: "perm:eduClsMaxCnt", Sort: 70, Remark: "支持修改班级满班人数。", MatchCodes: []string{"Editthenumberoffullclassmembers"}},
			{Name: "批量升期", Code: "perm:eduClsBatPrm", Sort: 80, Remark: "支持在班级内对学员进行批量升期报名操作。", MatchNames: []string{" 批量升期", "批量升期"}, MatchCodes: []string{"Batchupgradeperiod"}},
		},
	},
	{
		RouteCode: "page:eduTbl",
		Authorities: []institutionRouteAuthority{
			{Name: "查看所有的课表", Code: "perm:eduTblAll", Sort: 10, Weight: 10, GroupCode: "groupViewallclassschedules", Remark: "可以查看校区内所有的课表。", MatchCodes: []string{"Viewallclassschedules"}},
			{Name: "仅查看我的课表", Code: "perm:eduTblMy", Sort: 20, GroupCode: "groupViewallclassschedules", Remark: "仅可查看上课教师/上课助教为自己的课表。", MatchCodes: []string{"Onlyviewmyschedule"}},
			{Name: "新建/编辑/删除日程/添加补课学员时可选择所有班级/1v1", Code: "perm:eduTblAllClsOpn", Sort: 30, GroupCode: "groupallclasses", Remark: "可以新建、编辑、删除日程，添加补课学员，并选择所有班级/1v1。", MatchCodes: []string{"allclasses"}},
			{Name: "新建/编辑/删除日程/添加补课学员时可选择自己的班级/1v1", Code: "perm:eduTblOwnClsOpn", Sort: 40, GroupCode: "groupallclasses", Remark: "可以新建、编辑、删除日程，添加补课学员，并选择自己的班级/1v1。", MatchCodes: []string{"ownclass"}},
			{Name: "日程列表", Code: "perm:eduTblLst", Sort: 50, Remark: "支持PC端查看日程列表。", MatchCodes: []string{"ScheduleList"}},
			{Name: "冲突日程列表", Code: "perm:eduTblConfLst", Sort: 60, Remark: "支持PC端查看日程冲突列表。", MatchCodes: []string{"Conflictschedulelist"}},
			{Name: "导入日程", Code: "perm:eduTblImp", Sort: 70, Remark: "支持导入日程。", MatchCodes: []string{"Importschedule"}},
			{Name: "日程导出", Code: "perm:eduTblExp", Sort: 80, Remark: "支持PC端导出日程表和日程列表。", MatchCodes: []string{"Scheduleexport"}},
			{Name: "课表展示配置", Code: "perm:eduTblDispConfig", Sort: 90, Remark: "支持配置看板时间区间和颜色设置。", MatchCodes: []string{"Scheduledisplayconfiguration"}},
		},
	},
}

func (repo *Repository) ensureInstitutionMenuCatalog(ctx context.Context) error {
	for _, parent := range institutionMenuSeeds {
		parentID, err := repo.ensureInstitutionMenuNode(ctx, institutionMenuNodeSpec{
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
			childID, err := repo.ensureInstitutionMenuNode(ctx, institutionMenuNodeSpec{
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
				if _, err := repo.ensureInstitutionMenuNode(ctx, institutionMenuNodeSpec{
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

	if err := repo.ensureVisibleInstitutionRouteCatalog(ctx); err != nil {
		return err
	}

	if err := repo.ensureVisibleRouteAuthorityCatalog(ctx); err != nil {
		return err
	}

	if err := repo.clearInstitutionMenuURLPaths(ctx); err != nil {
		return err
	}

	if err := repo.cleanupInstitutionMenuCatalog(ctx); err != nil {
		return err
	}

	return nil
}

func (repo *Repository) ensureVisibleInstitutionRouteCatalog(ctx context.Context) error {
	if err := repo.removeDeprecatedInstitutionMenus(ctx); err != nil {
		return err
	}

	routeMenuIDs := make([]int64, 0, 64)
	for _, group := range institutionmenu.VisibleRouteCatalog {
		groupID, err := repo.ensureInstitutionMenuNode(ctx, institutionMenuNodeSpec{
			Name:       group.Name,
			Code:       group.Code,
			PID:        0,
			Level:      1,
			Sort:       group.Sort,
			Weight:     group.Sort,
			MenuType:   0,
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
			childID, err := repo.ensureInstitutionMenuNode(ctx, institutionMenuNodeSpec{
				Name:       child.Name,
				Code:       child.Code,
				PID:        groupID,
				Level:      2,
				Sort:       child.Sort,
				Weight:     child.Sort,
				MenuType:   0,
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

func (repo *Repository) ensureVisibleRouteAuthorityCatalog(ctx context.Context) error {
	for _, leaf := range listVisibleRouteLeaves() {
		routeID, err := repo.findMenuIDByCode(ctx, leaf.Code)
		if err != nil {
			return err
		}
		if routeID <= 0 {
			continue
		}

		if _, err := repo.ensureVisibleRouteAuthorityNode(ctx, routeID, buildRoutePageUseAuthority(leaf.Code, leaf.Name)); err != nil {
			return err
		}
	}

	for _, seed := range institutionRouteAuthoritySeeds {
		routeID, err := repo.findMenuIDByCode(ctx, seed.RouteCode)
		if err != nil {
			return err
		}
		if routeID <= 0 {
			continue
		}

		for _, authority := range seed.Authorities {
			if _, err := repo.ensureVisibleRouteAuthorityNode(ctx, routeID, authority); err != nil {
				return err
			}
		}
	}

	return nil
}

func (repo *Repository) clearInstitutionMenuURLPaths(ctx context.Context) error {
	_, err := repo.db.ExecContext(ctx, `
		UPDATE sso_menu
		SET url_path = NULL,
		    update_id = 'system',
		    update_time = NOW()
		WHERE own_type = 2
		  AND del_flag = 0
		  AND NULLIF(TRIM(IFNULL(url_path, '')), '') IS NOT NULL
	`)
	return err
}

func (repo *Repository) ensureVisibleRouteAuthorityNode(ctx context.Context, routeID int64, authority institutionRouteAuthority) (int64, error) {
	spec := institutionMenuNodeSpec{
		Name:      authority.Name,
		Code:      institutionmenu.NormalizeCode(authority.Code),
		PID:       routeID,
		Level:     3,
		Sort:      authority.Sort,
		Weight:    authority.Weight,
		MenuType:  authority.MenuType,
		GroupCode: strings.TrimSpace(authority.GroupCode),
		Introduce: strings.TrimSpace(authority.Remark),
		Remark:    strings.TrimSpace(authority.Remark),
	}

	if id, err := repo.findDirectChildMenuIDByCode(ctx, routeID, authority.Code); err != nil {
		return 0, err
	} else if id > 0 {
		return id, nil
	}

	for _, legacyCode := range authority.MatchCodes {
		if id, err := repo.findDirectChildMenuIDByCode(ctx, routeID, legacyCode); err != nil {
			return 0, err
		} else if id > 0 {
			return id, repo.updateInstitutionMenuNode(ctx, id, spec)
		}
	}

	names := uniqueMenuNames(append([]string{authority.Name}, authority.MatchNames...)...)
	if id, err := repo.findDirectChildMenuIDByNames(ctx, routeID, names); err != nil {
		return 0, err
	} else if id > 0 {
		return id, repo.updateInstitutionMenuNode(ctx, id, spec)
	}

	id, err := repo.insertInstitutionMenuNode(ctx, spec)
	if err != nil {
		return 0, err
	}

	return id, repo.copyInstitutionModuleBindings(ctx, routeID, id)
}

func (repo *Repository) removeDeprecatedInstitutionMenus(ctx context.Context) error {
	rootIDs, err := repo.findDeprecatedInstitutionMenuRootIDs(ctx)
	if err != nil {
		return err
	}
	if len(rootIDs) == 0 {
		return nil
	}

	menuIDs, err := repo.collectInstitutionMenuDescendantIDs(ctx, rootIDs)
	if err != nil {
		return err
	}
	if len(menuIDs) == 0 {
		return nil
	}

	args := make([]any, 0, len(menuIDs))
	placeholders := make([]string, 0, len(menuIDs))
	for _, id := range menuIDs {
		args = append(args, id)
		placeholders = append(placeholders, "?")
	}

	if _, err := repo.db.ExecContext(ctx, `
		DELETE FROM sys_module_menu
		WHERE menu_id IN (`+strings.Join(placeholders, ",")+`)
	`, args...); err != nil {
		return err
	}

	if _, err := repo.db.ExecContext(ctx, `
		DELETE FROM sso_role_menu
		WHERE menu_id IN (`+strings.Join(placeholders, ",")+`)
	`, args...); err != nil {
		return err
	}

	if _, err := repo.db.ExecContext(ctx, `
		DELETE FROM sso_menu
		WHERE id IN (`+strings.Join(placeholders, ",")+`)
	`, args...); err != nil {
		return err
	}

	return nil
}

func (repo *Repository) findDeprecatedInstitutionMenuRootIDs(ctx context.Context) ([]int64, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT id
		FROM sso_menu
		WHERE del_flag = 0
		  AND own_type = 2
		  AND (
			menu_code = 'page:home'
			OR (IFNULL(pid, 0) = 0 AND TRIM(IFNULL(menu_name, '')) = '首页')
		  )
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readAllMenuIDs(rows)
}

func (repo *Repository) collectInstitutionMenuDescendantIDs(ctx context.Context, rootIDs []int64) ([]int64, error) {
	seen := make(map[int64]struct{}, len(rootIDs))
	queue := make([]int64, 0, len(rootIDs))
	for _, id := range rootIDs {
		if id <= 0 {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		queue = append(queue, id)
	}

	for len(queue) > 0 {
		levelIDs := append([]int64(nil), queue...)
		queue = queue[:0]

		args := make([]any, 0, len(levelIDs))
		placeholders := make([]string, 0, len(levelIDs))
		for _, id := range levelIDs {
			args = append(args, id)
			placeholders = append(placeholders, "?")
		}

		rows, err := repo.db.QueryContext(ctx, `
			SELECT id
			FROM sso_menu
			WHERE del_flag = 0
			  AND own_type = 2
			  AND pid IN (`+strings.Join(placeholders, ",")+`)
		`, args...)
		if err != nil {
			return nil, err
		}

		childIDs, err := readAllMenuIDs(rows)
		rows.Close()
		if err != nil {
			return nil, err
		}

		for _, childID := range childIDs {
			if _, exists := seen[childID]; exists {
				continue
			}
			seen[childID] = struct{}{}
			queue = append(queue, childID)
		}
	}

	result := make([]int64, 0, len(seen))
	for id := range seen {
		result = append(result, id)
	}
	return result, nil
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
	Introduce  string
	Remark     string
	MatchNames []string
}

func (repo *Repository) ensureInstitutionMenuNode(ctx context.Context, spec institutionMenuNodeSpec) (int64, error) {
	if strings.TrimSpace(spec.Name) == "" {
		return 0, fmt.Errorf("menu name is required")
	}
	spec.Code = institutionmenu.NormalizeCode(spec.Code)

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

func (repo *Repository) findDirectChildMenuIDByCode(ctx context.Context, pid int64, code string) (int64, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return 0, nil
	}

	normalized := institutionmenu.NormalizeCode(code)
	query := `
		SELECT id
		FROM sso_menu
		WHERE del_flag = 0
		  AND own_type = 2
		  AND pid = ?
		  AND menu_code IN (?, ?)
		LIMIT 1
	`
	rows, err := repo.db.QueryContext(ctx, query, pid, code, normalized)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	return readSingleMenuID(rows)
}

func (repo *Repository) findDirectChildMenuIDByNames(ctx context.Context, pid int64, names []string) (int64, error) {
	for _, name := range names {
		id, err := repo.findChildMenuID(ctx, pid, name)
		if err != nil {
			return 0, err
		}
		if id > 0 {
			return id, nil
		}
	}

	return 0, nil
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
		    url_path = NULL,
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
	`, spec.Name, spec.Code, spec.PID, spec.Sort, spec.Introduce, spec.Level, spec.Weight, emptyToNullString(spec.GroupCode), spec.MenuType, spec.Remark, id)
	return err
}

func (repo *Repository) copyInstitutionModuleBindings(ctx context.Context, sourceMenuID, targetMenuID int64) error {
	if sourceMenuID <= 0 || targetMenuID <= 0 || sourceMenuID == targetMenuID {
		return nil
	}

	_, err := repo.db.ExecContext(ctx, `
		INSERT INTO sys_module_menu (module_id, menu_id, del_flag)
		SELECT smm.module_id, ?, 0
		FROM sys_module_menu smm
		WHERE smm.menu_id = ?
		  AND IFNULL(smm.del_flag, 0) = 0
		  AND NOT EXISTS (
		    SELECT 1
		    FROM sys_module_menu existing
		    WHERE existing.module_id = smm.module_id
		      AND existing.menu_id = ?
		      AND IFNULL(existing.del_flag, 0) = 0
		  )
	`, targetMenuID, sourceMenuID, targetMenuID)
	return err
}

func (repo *Repository) insertInstitutionMenuNode(ctx context.Context, spec institutionMenuNodeSpec) (int64, error) {
	result, err := repo.db.ExecContext(ctx, `
		INSERT INTO sso_menu (
			uuid, version, menu_name, url_path, menu_code, menu_type, pid, sort, is_system,
			introduce, own_type, level, weight, group_code, create_id, create_time,
			update_id, update_time, del_flag, remark
		)
		VALUES (?, 0, ?, NULL, ?, ?, ?, ?, 1, ?, 2, ?, ?, ?, 'system', NOW(), 'system', NOW(), 0, ?)
	`, uuid.NewString(), spec.Name, spec.Code, spec.MenuType, spec.PID, spec.Sort, spec.Introduce, spec.Level, spec.Weight, emptyToNullString(spec.GroupCode), spec.Remark)
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

func (row *institutionMenuCatalogRow) pathParts(rowsByID map[int64]*institutionMenuCatalogRow) []string {
	if row == nil {
		return nil
	}
	if row.path != nil {
		return row.path
	}
	if row.PID <= 0 {
		row.path = []string{row.Name}
		return row.path
	}

	parent := rowsByID[row.PID]
	if parent == nil {
		row.path = []string{row.Name}
		return row.path
	}

	parentPath := append([]string(nil), parent.pathParts(rowsByID)...)
	row.path = append(parentPath, row.Name)
	return row.path
}

func (row *institutionMenuCatalogRow) pathKey(rowsByID map[int64]*institutionMenuCatalogRow) string {
	return institutionMenuPathKey(row.pathParts(rowsByID))
}

func (row *institutionMenuCatalogRow) topName(rowsByID map[int64]*institutionMenuCatalogRow) string {
	parts := row.pathParts(rowsByID)
	if len(parts) == 0 {
		return ""
	}
	return parts[0]
}

func institutionMenuPathKey(parts []string) string {
	trimmed := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		trimmed = append(trimmed, part)
	}
	return strings.Join(trimmed, " / ")
}

func buildInstitutionCanonicalCodeMap() map[string]string {
	codes := make(map[string]string, 256)
	routePathByCode := make(map[string][2]string, 64)

	for _, parent := range institutionMenuSeeds {
		parentPath := []string{parent.ParentName}
		codes[institutionMenuPathKey(parentPath)] = institutionmenu.NormalizeCode(parent.ParentCode)
		for _, child := range parent.Children {
			childPath := append(append([]string(nil), parentPath...), child.Name)
			codes[institutionMenuPathKey(childPath)] = institutionmenu.NormalizeCode(child.Code)
			for _, authority := range child.Authorities {
				authorityPath := append(append([]string(nil), childPath...), authority.Name)
				codes[institutionMenuPathKey(authorityPath)] = institutionmenu.NormalizeCode(authority.Code)
			}
		}
	}

	for _, group := range institutionmenu.VisibleRouteCatalog {
		groupPath := []string{group.Name}
		codes[institutionMenuPathKey(groupPath)] = institutionmenu.NormalizeCode(group.Code)
		for _, child := range group.Children {
			childPath := append(append([]string(nil), groupPath...), child.Name)
			codes[institutionMenuPathKey(childPath)] = institutionmenu.NormalizeCode(child.Code)
			routePathByCode[child.Code] = [2]string{group.Name, child.Name}
		}
	}

	for _, seed := range institutionRouteAuthoritySeeds {
		routePath, exists := routePathByCode[seed.RouteCode]
		if !exists {
			continue
		}
		for _, authority := range seed.Authorities {
			authorityPath := []string{routePath[0], routePath[1], authority.Name}
			codes[institutionMenuPathKey(authorityPath)] = institutionmenu.NormalizeCode(authority.Code)
		}
	}

	return codes
}

func buildInstitutionCanonicalPathIndex(codes map[string]string) map[string][]string {
	index := make(map[string][]string, len(codes))
	for pathKey := range codes {
		parts := strings.Split(pathKey, " / ")
		if len(parts) == 0 {
			continue
		}
		key := fmt.Sprintf("%s|%d|%s", parts[0], len(parts), parts[len(parts)-1])
		index[key] = append(index[key], pathKey)
	}
	return index
}

func (repo *Repository) cleanupInstitutionMenuCatalog(ctx context.Context) error {
	canonicalCodes := buildInstitutionCanonicalCodeMap()
	canonicalPaths := buildInstitutionCanonicalPathIndex(canonicalCodes)

	for pass := 0; pass < 12; pass++ {
		rows, rowsByID, err := repo.loadInstitutionMenuCatalogRows(ctx)
		if err != nil {
			return err
		}

		changedCode, err := repo.rewriteInstitutionMenuCodes(ctx, rows, rowsByID, canonicalCodes)
		if err != nil {
			return err
		}

		changedDuplicate, err := repo.mergeDuplicateInstitutionMenuPaths(ctx, rows, rowsByID, canonicalCodes)
		if err != nil {
			return err
		}

		changedMisplaced, err := repo.mergeMisplacedInstitutionMenus(ctx, rows, rowsByID, canonicalPaths)
		if err != nil {
			return err
		}

		if !changedCode && !changedDuplicate && !changedMisplaced {
			return nil
		}
	}

	return nil
}

func (repo *Repository) loadInstitutionMenuCatalogRows(ctx context.Context) ([]*institutionMenuCatalogRow, map[int64]*institutionMenuCatalogRow, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT id, IFNULL(pid, 0), IFNULL(level, 0), TRIM(IFNULL(menu_name, '')), TRIM(IFNULL(menu_code, ''))
		FROM sso_menu
		WHERE own_type = 2 AND del_flag = 0
		ORDER BY IFNULL(level, 0), IFNULL(sort, 0), id
	`)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	items := make([]*institutionMenuCatalogRow, 0, 256)
	itemsByID := make(map[int64]*institutionMenuCatalogRow, 256)
	for rows.Next() {
		item := &institutionMenuCatalogRow{}
		if err := rows.Scan(&item.ID, &item.PID, &item.Level, &item.Name, &item.Code); err != nil {
			return nil, nil, err
		}
		items = append(items, item)
		itemsByID[item.ID] = item
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}
	return items, itemsByID, nil
}

func (repo *Repository) rewriteInstitutionMenuCodes(ctx context.Context, rows []*institutionMenuCatalogRow, rowsByID map[int64]*institutionMenuCatalogRow, canonicalCodes map[string]string) (bool, error) {
	desiredCodes := make(map[int64]string, len(rows))
	changed := false

	for _, row := range rows {
		pathKey := row.pathKey(rowsByID)
		desired := strings.TrimSpace(canonicalCodes[pathKey])
		parentCode := ""
		if parent := rowsByID[row.PID]; parent != nil {
			parentCode = strings.TrimSpace(desiredCodes[parent.ID])
			if parentCode == "" {
				parentCode = strings.TrimSpace(parent.Code)
			}
		}
		if desired == "" {
			desired = institutionmenu.DeriveCode(row.Level, row.Code, row.Name, parentCode)
		}
		desired = strings.TrimSpace(institutionmenu.NormalizeCode(desired))
		if desired == "" {
			desiredCodes[row.ID] = row.Code
			continue
		}

		desiredCodes[row.ID] = desired
		if desired == row.Code {
			continue
		}

		if _, err := repo.db.ExecContext(ctx, `
			UPDATE sso_menu
			SET menu_code = ?,
			    update_id = 'system',
			    update_time = NOW()
			WHERE id = ?
		`, desired, row.ID); err != nil {
			return false, err
		}
		row.Code = desired
		changed = true
	}

	return changed, nil
}

func (repo *Repository) mergeDuplicateInstitutionMenuPaths(ctx context.Context, rows []*institutionMenuCatalogRow, rowsByID map[int64]*institutionMenuCatalogRow, canonicalCodes map[string]string) (bool, error) {
	grouped := make(map[string][]*institutionMenuCatalogRow, len(rows))
	for _, row := range rows {
		key := fmt.Sprintf("%d|%d|%s", row.PID, row.Level, row.Name)
		grouped[key] = append(grouped[key], row)
	}

	changed := false
	for _, group := range grouped {
		if len(group) <= 1 {
			continue
		}

		keeper := pickInstitutionMenuKeeper(group, rowsByID, canonicalCodes)
		for _, row := range group {
			if row.ID == keeper.ID {
				continue
			}
			if err := repo.mergeInstitutionMenuNode(ctx, keeper.ID, row.ID); err != nil {
				return false, err
			}
			changed = true
		}
	}

	return changed, nil
}

func (repo *Repository) mergeMisplacedInstitutionMenus(ctx context.Context, rows []*institutionMenuCatalogRow, rowsByID map[int64]*institutionMenuCatalogRow, canonicalPaths map[string][]string) (bool, error) {
	pathToRow := make(map[string]*institutionMenuCatalogRow, len(rows))
	for _, row := range rows {
		pathToRow[row.pathKey(rowsByID)] = row
	}

	removed := make(map[int64]struct{}, 16)
	changed := false
	for _, row := range rows {
		if row.Level <= 1 {
			continue
		}
		if _, exists := removed[row.ID]; exists {
			continue
		}

		currentPath := row.pathKey(rowsByID)
		pathKey := fmt.Sprintf("%s|%d|%s", row.topName(rowsByID), row.Level, row.Name)
		targetPaths := canonicalPaths[pathKey]
		if len(targetPaths) != 1 {
			continue
		}
		targetPath := targetPaths[0]
		if targetPath == currentPath {
			continue
		}

		target := pathToRow[targetPath]
		if target == nil || target.ID == row.ID {
			continue
		}
		if _, exists := removed[target.ID]; exists {
			continue
		}

		if err := repo.mergeInstitutionMenuNode(ctx, target.ID, row.ID); err != nil {
			return false, err
		}
		removed[row.ID] = struct{}{}
		changed = true
	}

	return changed, nil
}

func pickInstitutionMenuKeeper(group []*institutionMenuCatalogRow, rowsByID map[int64]*institutionMenuCatalogRow, canonicalCodes map[string]string) *institutionMenuCatalogRow {
	if len(group) == 0 {
		return nil
	}

	keeper := group[0]
	canonicalCode := canonicalCodes[group[0].pathKey(rowsByID)]
	for _, row := range group {
		switch {
		case canonicalCode != "" && row.Code == canonicalCode:
			return row
		case canonicalCode == "" && isCanonicalMenuCode(row.Level, row.Code) && !isCanonicalMenuCode(keeper.Level, keeper.Code):
			keeper = row
		case canonicalCode == "" && isCanonicalMenuCode(row.Level, row.Code) == isCanonicalMenuCode(keeper.Level, keeper.Code) && row.ID < keeper.ID:
			keeper = row
		case canonicalCode != "" && keeper.Code != canonicalCode && row.ID < keeper.ID:
			keeper = row
		}
	}
	return keeper
}

func isCanonicalMenuCode(level int, code string) bool {
	switch level {
	case 1:
		return strings.HasPrefix(strings.TrimSpace(code), "grp:")
	case 2:
		return strings.HasPrefix(strings.TrimSpace(code), "page:")
	case 3:
		return strings.HasPrefix(strings.TrimSpace(code), "perm:")
	default:
		return false
	}
}

func (repo *Repository) mergeInstitutionMenuNode(ctx context.Context, keepID, removeID int64) error {
	if keepID <= 0 || removeID <= 0 || keepID == removeID {
		return nil
	}

	if err := repo.copyInstitutionModuleBindings(ctx, removeID, keepID); err != nil {
		return err
	}

	if _, err := repo.db.ExecContext(ctx, `
		INSERT INTO sso_role_menu (role_id, menu_id)
		SELECT rm.role_id, ?
		FROM sso_role_menu rm
		WHERE rm.menu_id = ?
		  AND NOT EXISTS (
		    SELECT 1
		    FROM sso_role_menu existing
		    WHERE existing.role_id = rm.role_id
		      AND existing.menu_id = ?
		  )
	`, keepID, removeID, keepID); err != nil {
		return err
	}

	if _, err := repo.db.ExecContext(ctx, `
		UPDATE sso_menu
		SET pid = ?,
		    update_id = 'system',
		    update_time = NOW()
		WHERE own_type = 2
		  AND del_flag = 0
		  AND pid = ?
	`, keepID, removeID); err != nil {
		return err
	}

	if _, err := repo.db.ExecContext(ctx, `
		DELETE FROM sso_role_menu
		WHERE menu_id = ?
	`, removeID); err != nil {
		return err
	}

	if _, err := repo.db.ExecContext(ctx, `
		DELETE FROM sys_module_menu
		WHERE menu_id = ?
	`, removeID); err != nil {
		return err
	}

	if _, err := repo.db.ExecContext(ctx, `
		DELETE FROM sso_menu
		WHERE id = ?
	`, removeID); err != nil {
		return err
	}

	return nil
}
