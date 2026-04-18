package institutionmenu

import (
	"strings"
	"unicode"
)

const (
	groupCodePrefix = "grp:"
	routeCodePrefix = "page:"
	authCodePrefix  = "perm:"
)

var compactPhraseRules = []struct {
	From []string
	To   string
}{
	{From: []string{"ONE", "TO", "ONE"}, To: "o2o"},
	{From: []string{"VALUE", "ADDED"}, To: "va"},
	{From: []string{"THIRD", "PARTY"}, To: "tp"},
	{From: []string{"DATA", "CENTER"}, To: "dc"},
}

var compactTokenMap = map[string]string{
	"ACADEMIC":      "acd",
	"ADDED":         "add",
	"AI":            "ai",
	"ALBUM":         "alb",
	"ALL":           "all",
	"ALLOCATION":    "aln",
	"APPLY":         "apy",
	"APPROVAL":      "apv",
	"ARCHIVE":       "arc",
	"ASSESSMENT":    "asm",
	"ASSIGN":        "asn",
	"ASSISTANT":     "ast",
	"ATTENDANCE":    "att",
	"ATTR":          "atr",
	"AUTO":          "ato",
	"BASIC":         "bsc",
	"BATCH":         "bat",
	"BILL":          "bil",
	"BIZ":           "biz",
	"BRAND":         "brd",
	"BUSY":          "bsy",
	"CALL":          "cal",
	"CAMPAIGN":      "camp",
	"CASHIER":       "csh",
	"CATEGORY":      "ctg",
	"CENTER":        "ctr",
	"CHANGE":        "chg",
	"CHANNEL":       "chn",
	"CLAIM":         "clm",
	"CLEAR":         "clr",
	"CLASS":         "cls",
	"CLOCK":         "clk",
	"COLLECTION":    "col",
	"CONFIRM":       "cfm",
	"CONSULTANT":    "cst",
	"CONSULTANTS":   "cst",
	"CONVERT":       "cvt",
	"COUNT":         "cnt",
	"COURSE":        "crs",
	"COURSES":       "crs",
	"CREATE":        "crt",
	"DATA":          "dat",
	"DELETE":        "del",
	"DEPT":          "dept",
	"DETAIL":        "dtl",
	"DETAILS":       "dtl",
	"DEVICE":        "dev",
	"DISCOUNT":      "dct",
	"EDU":           "edu",
	"EDIT":          "edt",
	"EFFECTIVE":     "eff",
	"ENROLL":        "enr",
	"ENROLLMENT":    "enr",
	"EXAM":          "exm",
	"EXPORT":        "exp",
	"FACE":          "fac",
	"FACIAL":        "fac",
	"FEEDBACK":      "fdb",
	"FINANCE":       "fin",
	"FOLLOW":        "flw",
	"FORM":          "frm",
	"GOODS":         "gds",
	"GRADE":         "grd",
	"GROWTH":        "grw",
	"HOME":          "home",
	"HOMEWORK":      "hwk",
	"HOURS":         "hrs",
	"IMPORT":        "imp",
	"INCOME":        "inc",
	"INFO":          "info",
	"INTERNAL":      "intl",
	"INTENTION":     "int",
	"INTERACTIVE":   "iact",
	"INVENTORY":     "inv",
	"LEAVE":         "lev",
	"LIST":          "lst",
	"LISTENING":     "lsn",
	"LOCKED":        "lck",
	"MAKEUP":        "mkp",
	"MANAGE":        "mng",
	"MANAGEMENT":    "mng",
	"MAX":           "max",
	"MEITUAN":       "mt",
	"MICRO":         "mic",
	"MINIAPP":       "mini",
	"MISCELLANEOUS": "msc",
	"MODIFY":        "mdf",
	"MORE":          "mor",
	"MY":            "my",
	"NOTICE":        "ntc",
	"OFFICIAL":      "off",
	"OPERATE":       "opr",
	"OPERATION":     "opn",
	"ORDER":         "ord",
	"ORG":           "org",
	"OVERVIEW":      "ovw",
	"PARTY":         "pty",
	"PAYROLL":       "pay",
	"PERFORMANCE":   "pfm",
	"PERSONAL":      "psn",
	"PHONE":         "phn",
	"PLAN":          "pln",
	"POINT":         "pnt",
	"POOL":          "pol",
	"PRINT":         "prt",
	"PROMOTION":     "prm",
	"PUBLIC":        "pub",
	"RECHARGE":      "rch",
	"RECORD":        "rec",
	"RECOVERY":      "rcv",
	"REFUND":        "rfd",
	"RELATION":      "rel",
	"RELATIONSHIP":  "rel",
	"RENEW":         "rnw",
	"RENEWAL":       "rnw",
	"REPLY":         "rpy",
	"REPORT":        "rpt",
	"REVENUE":       "rev",
	"REVIEW":        "rvw",
	"ROLE":          "role",
	"ROLL":          "rol",
	"RULE":          "rul",
	"SAFE":          "saf",
	"SALARY":        "sal",
	"SALES":         "sls",
	"SCALE":         "scl",
	"SCHEDULE":      "sch",
	"SCHOOL":        "sch",
	"SCORE":         "scr",
	"SCREEN":        "scr",
	"SELF":          "self",
	"SEND":          "snd",
	"SENSITIVE":     "sns",
	"SETTING":       "set",
	"SIGN":          "sgn",
	"SMS":           "sms",
	"STAFF":         "stf",
	"STATUS":        "sts",
	"STUDENT":       "stu",
	"STUDENTS":      "stus",
	"STYLE":         "sty",
	"SUMMARY":       "sum",
	"SUPERVISE":     "sup",
	"SUPERVISOR":    "sup",
	"TARGET":        "tgt",
	"TEACHER":       "tch",
	"TEMPLATE":      "tpl",
	"TEST":          "tst",
	"TEXTBOOK":      "txt",
	"THIRD":         "thd",
	"TIMETABLE":     "tbl",
	"TRIAL":         "trl",
	"TUITION":       "tui",
	"UPDATE":        "upd",
	"VALIDITY":      "vld",
	"VALUE":         "val",
	"VIEW":          "view",
	"WARNING":       "wrn",
	"WECHAT":        "wc",
	"WHOLE":         "whl",
	"WITH":          "wth",
	"WITHDRAWAL":    "wdr",
	"WRITE":         "wrt",
}

var chineseCompactPhraseRules = []struct {
	From string
	To   string
}{
	{From: "查看校区部门和员工", To: "viewCampusDeptStaff"},
	{From: "管理员工忙碌时段", To: "staffBusyMng"},
	{From: "查看员工忙碌时段", To: "viewStaffBusy"},
	{From: "校区员工管理", To: "campusStaffMng"},
	{From: "校区部门管理", To: "campusDeptMng"},
	{From: "管理督办", To: "superviseMng"},
	{From: "可见跟进数据", To: "visFlwDat"},
	{From: "可见授课数据", To: "visTeachDat"},
	{From: "招生工具", To: "enrTool"},
	{From: "到账确认", To: "receiptCfm"},
	{From: "人脸考勤设备", To: "facAttDev"},
	{From: "人脸采集", To: "facCap"},
	{From: "人脸考勤", To: "facAtt"},
	{From: "招生表单推广", To: "enrFrmPrm"},
	{From: "招生表单", To: "enrFrm"},
	{From: "推荐人更换", To: "refChg"},
	{From: "收款查询", To: "payQry"},
	{From: "云打印", To: "cloudPrt"},
	{From: "学费码", To: "tuiCode"},
	{From: "可为所有学员报名续费", To: "allStuSgnRnw"},
	{From: "仅可为我的学员报名续费", To: "myStuSgnRnw"},
	{From: "查看校区所有报读列表", To: "viewCampusEnrLst"},
	{From: "查看校区所有分班列表", To: "viewCampusClsLst"},
	{From: "仅查看我的学员报读列表", To: "viewMyStuEnrLst"},
	{From: "仅查看我的学员分班列表", To: "viewMyStuClsLst"},
	{From: "报读列表导出", To: "enrLstExp"},
	{From: "分班列表导出", To: "clsLstExp"},
	{From: "插班补课可选择所有日程", To: "mkpAllSch"},
	{From: "插班补课仅可选择自己的日程", To: "mkpOwnSch"},
	{From: "查看课时汇总学费消耗", To: "viewHrsSumTuiUse"},
	{From: "点名编辑限制当天", To: "rolCalEdtDay"},
	{From: "上课点名限制当天", To: "rolCalDay"},
	{From: "未排课点名可选择所有班级/1v1", To: "rolCalAllCls1v1"},
	{From: "未排课点名仅可选择自己的班级/1v1", To: "rolCalOwnCls1v1"},
	{From: "欠费学员补费", To: "arrStuPay"},
	{From: "欠费学员仅清算", To: "arrStuClr"},
	{From: "欠费学员抵扣并清算", To: "arrStuDedClr"},
	{From: "修改经办、支付日期", To: "mdfHdlPayDt"},
	{From: "应续生成记录", To: "rnwGenRec"},
	{From: "作废应续记录", To: "voidRnwRec"},
	{From: "学员分班调班", To: "stuClsChg"},
	{From: "待关注学员", To: "needFlwStu"},
	{From: "修改升期关系", To: "mdfPrmRel"},
	{From: "学杂教材", To: "miscTxt"},
	{From: "退学杂教材费", To: "rfdMiscTxtFee"},
	{From: "学杂费", To: "miscFee"},
	{From: "查看系统账单统计", To: "viewSysBilStat"},
	{From: "查看系统账单", To: "viewSysBil"},
	{From: "订单备注编辑", To: "ordRemarkEdt"},
	{From: "订单报读类型编辑", To: "ordEnrTypeEdt"},
	{From: "订单标签编辑", To: "ordTagEdt"},
	{From: "作废订单", To: "ordVoid"},
	{From: "关闭订单", To: "ordClose"},
	{From: "优惠券管理", To: "couponMng"},
	{From: "管理场地预约", To: "venueResMng"},
	{From: "场地管理", To: "venueMng"},
	{From: "约课管理", To: "bookClsMng"},
	{From: "数据中心", To: "dc"},
	{From: "招生数据", To: "enrDat"},
	{From: "教务数据", To: "eduDat"},
	{From: "家校数据", To: "homeDat"},
	{From: "财务数据", To: "finDat"},
	{From: "数据概览", To: "datOvw"},
	{From: "报表管理", To: "rptMng"},
	{From: "本部门及以下", To: "dept"},
	{From: "仅查看我的学员报名续费", To: "viewMyStuSgnRnw"},
	{From: "仅查看我的", To: "viewMy"},
	{From: "查看所有的", To: "viewAll"},
	{From: "查看所有", To: "viewAll"},
	{From: "查看全部", To: "viewAll"},
	{From: "查看我的", To: "viewMy"},
	{From: "查看并管理", To: "viewMng"},
	{From: "查看和管理", To: "viewMng"},
	{From: "报名续费", To: "sgnRnw"},
	{From: "整单优惠", To: "ordDct"},
	{From: "评分模板", To: "scrTpl"},
	{From: "成长档案", To: "grwArc"},
	{From: "课程商品", To: "crsGds"},
	{From: "康复记录", To: "rcvRec"},
	{From: "导出记录", To: "expRec"},
	{From: "导入记录", To: "impRec"},
	{From: "跟进状态", To: "flwSts"},
	{From: "跟进记录", To: "flwRec"},
	{From: "上课教师", To: "tch"},
	{From: "上课助教", To: "ast"},
	{From: "学管师", To: "sup"},
	{From: "公有池", To: "pubPol"},
	{From: "销售员", To: "sls"},
	{From: "班主任", To: "clsTea"},
	{From: "意向学员", To: "intStu"},
	{From: "学员", To: "stu"},
	{From: "班级", To: "cls"},
	{From: "课表", To: "tbl"},
	{From: "日程", To: "sch"},
	{From: "试听", To: "trl"},
	{From: "补课", To: "mkp"},
	{From: "点名", To: "rolCal"},
	{From: "订单", To: "ord"},
	{From: "账单", To: "bil"},
	{From: "审批", To: "apv"},
	{From: "业绩", To: "pfm"},
	{From: "工资", To: "pay"},
	{From: "收入", To: "inc"},
	{From: "通知", To: "ntc"},
	{From: "请假", To: "lev"},
	{From: "短信", To: "sms"},
	{From: "专属公众号", To: "off"},
	{From: "专属小程序", To: "mini"},
	{From: "微机构", To: "micOrg"},
	{From: "转介绍", To: "ref"},
	{From: "美团", To: "mt"},
	{From: "操作", To: "opr"},
	{From: "安排", To: "arn"},
	{From: "办理", To: "hnd"},
	{From: "绑定", To: "bnd"},
	{From: "解绑", To: "ubd"},
	{From: "新建", To: "crt"},
	{From: "新增", To: "crt"},
	{From: "编辑", To: "edt"},
	{From: "删除", To: "del"},
	{From: "管理", To: "mng"},
	{From: "设置", To: "set"},
	{From: "分配", To: "asn"},
	{From: "认领", To: "clm"},
	{From: "批量", To: "bat"},
	{From: "导入", To: "imp"},
	{From: "导出", To: "exp"},
	{From: "查看", To: "view"},
	{From: "转入", To: "in"},
	{From: "转为失效", To: "invalid"},
	{From: "转课", To: "chgCrs"},
	{From: "退课", To: "rfdCrs"},
	{From: "停复课", To: "spnRsm"},
	{From: "结课", To: "cls"},
	{From: "撤销", To: "rvk"},
	{From: "修改", To: "mdf"},
	{From: "详情", To: "dtl"},
	{From: "状态", To: "sts"},
	{From: "记录", To: "rec"},
	{From: "列表", To: "lst"},
	{From: "规则", To: "rul"},
	{From: "有效期", To: "vld"},
	{From: "生日", To: "bdy"},
	{From: "缺课", To: "abs"},
	{From: "欠费", To: "arr"},
	{From: "待续费", To: "rnw"},
	{From: "待分班", To: "clsAsn"},
	{From: "关联人员", To: "relStaff"},
	{From: "渠道", To: "chn"},
}

var legacyStopWords = map[string]struct{}{
	"A":     {},
	"AN":    {},
	"AND":   {},
	"ARE":   {},
	"AT":    {},
	"BE":    {},
	"BY":    {},
	"CAN":   {},
	"FOR":   {},
	"FROM":  {},
	"IN":    {},
	"INTO":  {},
	"IS":    {},
	"OF":    {},
	"ON":    {},
	"OR":    {},
	"SHALL": {},
	"THE":   {},
	"TO":    {},
	"WITH":  {},
}

var legacyWordBuckets = buildLegacyWordBuckets()

func NormalizeCode(code string) string {
	trimmed := strings.TrimSpace(code)
	if trimmed == "" {
		return ""
	}

	switch {
	case strings.HasPrefix(trimmed, groupCodePrefix):
		return groupCodePrefix + normalizeCamelSuffix(strings.TrimPrefix(trimmed, groupCodePrefix))
	case strings.HasPrefix(trimmed, routeCodePrefix):
		return routeCodePrefix + normalizeCamelSuffix(strings.TrimPrefix(trimmed, routeCodePrefix))
	case strings.HasPrefix(trimmed, authCodePrefix):
		return authCodePrefix + normalizeCamelSuffix(strings.TrimPrefix(trimmed, authCodePrefix))
	default:
		return trimmed
	}
}

func DeriveCode(level int, currentCode, menuName, parentCode string) string {
	currentCode = strings.TrimSpace(currentCode)
	menuName = strings.TrimSpace(menuName)
	parentCode = strings.TrimSpace(parentCode)

	switch {
	case strings.HasPrefix(currentCode, groupCodePrefix),
		strings.HasPrefix(currentCode, routeCodePrefix),
		strings.HasPrefix(currentCode, authCodePrefix):
		return NormalizeCode(currentCode)
	}

	suffix := ""
	if !strings.HasPrefix(currentCode, groupCodePrefix) && !strings.HasPrefix(currentCode, routeCodePrefix) && !strings.HasPrefix(currentCode, authCodePrefix) {
		suffix = compactLooseCode(currentCode)
	}
	if suffix == "" && hasChinese(menuName) {
		suffix = compactChineseLabel(menuName)
	}
	if suffix == "" && menuName != "" {
		suffix = compactLooseCode(menuName)
	}
	if suffix == "" && currentCode != "" {
		suffix = codeSuffix(currentCode)
	}
	if suffix == "" {
		return currentCode
	}

	switch level {
	case 1:
		return groupCodePrefix + normalizeCamelSuffix(suffix)
	case 2:
		return routeCodePrefix + normalizeCamelSuffix(suffix)
	case 3:
		parentSuffix := codeSuffix(NormalizeCode(parentCode))
		suffix = trimCompactPrefix(normalizeCamelSuffix(suffix), parentSuffix)
		if parentSuffix == "" {
			return authCodePrefix + suffix
		}
		if suffix == "" {
			return authCodePrefix + parentSuffix
		}
		return authCodePrefix + parentSuffix + upperFirst(suffix)
	default:
		return currentCode
	}
}

func upperSnakeToCompactCamel(value string) string {
	parts := strings.FieldsFunc(strings.TrimSpace(value), func(r rune) bool {
		return r == '_' || r == ':' || r == '-' || unicode.IsSpace(r)
	})
	if len(parts) == 0 {
		return ""
	}

	return compactJoinedTokens(compactParts(parts))
}

func matchCompactPhrase(parts []string, start int) (string, bool, int) {
	for _, rule := range compactPhraseRules {
		if start+len(rule.From) > len(parts) {
			continue
		}

		matched := true
		for offset, expect := range rule.From {
			if strings.ToUpper(strings.TrimSpace(parts[start+offset])) != expect {
				matched = false
				break
			}
		}
		if matched {
			return rule.To, true, start + len(rule.From)
		}
	}
	return "", false, start
}

func compactParts(parts []string) []string {
	result := make([]string, 0, len(parts))
	for index := 0; index < len(parts); {
		if phrase, matched, nextIndex := matchCompactPhrase(parts, index); matched {
			result = append(result, phrase)
			index = nextIndex
			continue
		}

		result = append(result, abbreviateToken(parts[index]))
		index++
	}
	return result
}

func abbreviateToken(token string) string {
	token = strings.ToUpper(strings.TrimSpace(token))
	if token == "" {
		return ""
	}
	if value, exists := compactTokenMap[token]; exists {
		return value
	}

	lower := strings.ToLower(token)
	runes := []rune(lower)
	if len(runes) <= 3 {
		return lower
	}
	if len(runes) <= 6 {
		return string(runes)
	}
	return string(runes[:4])
}

func normalizeCamelSuffix(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if strings.ContainsAny(value, "_:- ") {
		return upperSnakeToCompactCamel(value)
	}

	runes := []rune(value)
	if len(runes) == 0 {
		return ""
	}

	firstLetterFound := false
	for index, current := range runes {
		if !unicode.IsLetter(current) {
			continue
		}
		if !firstLetterFound {
			runes[index] = unicode.ToLower(current)
			firstLetterFound = true
		}
	}
	return string(runes)
}

func compactChineseLabel(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	tokens := make([]string, 0, 6)
	for len(value) > 0 {
		matched := false
		for _, rule := range chineseCompactPhraseRules {
			if strings.HasPrefix(value, rule.From) {
				if rule.To != "" {
					tokens = append(tokens, rule.To)
				}
				value = strings.TrimSpace(strings.TrimPrefix(value, rule.From))
				matched = true
				break
			}
		}
		if matched {
			continue
		}

		runes := []rune(value)
		if len(runes) == 0 {
			break
		}
		current := runes[0]
		if unicode.IsSpace(current) || strings.ContainsRune("()（）/、，：:,-_", current) {
			value = strings.TrimSpace(string(runes[1:]))
			continue
		}
		if current < unicode.MaxASCII {
			asciiEnd := 1
			for asciiEnd < len(runes) && runes[asciiEnd] < unicode.MaxASCII {
				asciiEnd++
			}
			asciiToken := compactLooseCode(string(runes[:asciiEnd]))
			if asciiToken != "" {
				tokens = append(tokens, asciiToken)
			}
			value = strings.TrimSpace(string(runes[asciiEnd:]))
			continue
		}

		value = strings.TrimSpace(string(runes[1:]))
	}

	return compactJoinedTokens(tokens)
}

func compactLooseCode(value string) string {
	parts := splitLooseCodeParts(value)
	if len(parts) == 0 {
		return ""
	}

	filtered := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.ToUpper(strings.TrimSpace(part))
		if part == "" {
			continue
		}
		if _, skip := legacyStopWords[part]; skip {
			continue
		}
		filtered = append(filtered, part)
	}
	if len(filtered) == 0 {
		return ""
	}

	return compactJoinedTokens(compactParts(filtered))
}

func splitLooseCodeParts(value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	var builder strings.Builder
	var previous rune
	for index, current := range value {
		switch {
		case current == '/' || current == '\\' || current == '-' || current == '_' || current == ':' || unicode.IsSpace(current) || current == '（' || current == '）' || current == '(' || current == ')' || current == '，' || current == '、':
			builder.WriteRune(' ')
		case index > 0 && shouldSplitLooseRune(previous, current):
			builder.WriteRune(' ')
			builder.WriteRune(current)
		default:
			builder.WriteRune(current)
		}
		previous = current
	}

	fields := strings.Fields(builder.String())
	parts := make([]string, 0, len(fields)*2)
	for _, field := range fields {
		if field == "" || hasChinese(field) {
			continue
		}
		if isSimpleWord(field) {
			parts = append(parts, segmentLegacyWord(field)...)
			continue
		}
		parts = append(parts, strings.ToUpper(field))
	}

	return parts
}

func segmentLegacyWord(value string) []string {
	lower := strings.ToLower(strings.TrimSpace(value))
	if lower == "" {
		return nil
	}

	parts := make([]string, 0, 4)
	for len(lower) > 0 {
		matched, size := matchLegacyPrefix(lower)
		if matched != "" {
			parts = append(parts, matched)
			lower = lower[size:]
			continue
		}

		size = 1
		for size < len(lower) {
			if next, _ := matchLegacyPrefix(lower[size:]); next != "" {
				break
			}
			size++
		}

		parts = append(parts, strings.ToUpper(lower[:size]))
		lower = lower[size:]
	}

	return parts
}

func matchLegacyPrefix(value string) (string, int) {
	if value == "" {
		return "", 0
	}

	bucket := legacyWordBuckets[value[0]]
	for _, candidate := range bucket {
		if strings.HasPrefix(value, candidate) {
			return strings.ToUpper(candidate), len(candidate)
		}
	}
	return "", 0
}

func buildLegacyWordBuckets() map[byte][]string {
	words := make([]string, 0, len(compactTokenMap)+32)
	seen := make(map[string]struct{}, len(compactTokenMap)+32)

	appendWord := func(word string) {
		word = strings.TrimSpace(strings.ToLower(word))
		if word == "" {
			return
		}
		if _, exists := seen[word]; exists {
			return
		}
		seen[word] = struct{}{}
		words = append(words, word)
	}

	for word := range compactTokenMap {
		appendWord(word)
	}

	for _, extra := range []string{
		"allclasses", "attendance", "birthday", "classroom", "completion", "confirm", "conflict",
		"consultants", "consultant", "countdown", "details", "dropout", "effective", "enrollment",
		"experience", "export", "facial", "followup", "homeschool", "interaction", "listening",
		"management", "meituan", "microschool", "miniapp", "miscellaneous", "oneonone",
		"operation", "performance", "prospective", "references", "renewal", "restriction",
		"rollcall", "salesperson", "schedule", "supervisors", "supervisor", "templates",
		"textbooks", "valueadded", "wechat", "zinquiry",
	} {
		appendWord(extra)
	}

	for index := 0; index < len(words); index++ {
		for inner := index + 1; inner < len(words); inner++ {
			if len(words[inner]) > len(words[index]) {
				words[index], words[inner] = words[inner], words[index]
			}
		}
	}

	buckets := make(map[byte][]string, 26)
	for _, word := range words {
		buckets[word[0]] = append(buckets[word[0]], word)
	}
	return buckets
}

func shouldSplitLooseRune(previous, current rune) bool {
	if previous == 0 {
		return false
	}
	if unicode.IsDigit(previous) != unicode.IsDigit(current) {
		return true
	}
	return unicode.IsLower(previous) && unicode.IsUpper(current)
}

func isSimpleWord(value string) bool {
	hasLetter := false
	for _, current := range value {
		if unicode.IsLetter(current) {
			hasLetter = true
			continue
		}
		if !unicode.IsDigit(current) {
			return false
		}
	}
	return hasLetter
}

func hasChinese(value string) bool {
	for _, current := range value {
		if current > unicode.MaxASCII {
			return true
		}
	}
	return false
}

func compactJoinedTokens(parts []string) string {
	var builder strings.Builder
	for index, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		runes := []rune(part)
		if len(runes) == 0 {
			continue
		}
		if index == 0 {
			runes[0] = unicode.ToLower(runes[0])
		} else if unicode.IsLetter(runes[0]) {
			runes[0] = unicode.ToUpper(runes[0])
		}
		builder.WriteString(string(runes))
	}
	return builder.String()
}

func codeSuffix(code string) string {
	switch {
	case strings.HasPrefix(code, groupCodePrefix):
		return strings.TrimPrefix(code, groupCodePrefix)
	case strings.HasPrefix(code, routeCodePrefix):
		return strings.TrimPrefix(code, routeCodePrefix)
	case strings.HasPrefix(code, authCodePrefix):
		return strings.TrimPrefix(code, authCodePrefix)
	default:
		return normalizeCamelSuffix(code)
	}
}

func trimCompactPrefix(value, prefix string) string {
	value = normalizeCamelSuffix(value)
	prefix = normalizeCamelSuffix(prefix)
	if prefix == "" {
		return value
	}
	if !strings.HasPrefix(strings.ToLower(value), strings.ToLower(prefix)) {
		return value
	}

	trimmed := value[len(prefix):]
	if trimmed == "" {
		return ""
	}
	return normalizeCamelSuffix(trimmed)
}

func upperFirst(value string) string {
	runes := []rune(strings.TrimSpace(value))
	if len(runes) == 0 {
		return ""
	}
	if unicode.IsLetter(runes[0]) {
		runes[0] = unicode.ToUpper(runes[0])
	}
	return string(runes)
}
