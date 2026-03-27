package service

import "strings"

type rechargeAccountImportMode string

const (
	rechargeAccountImportModeUnknown   rechargeAccountImportMode = ""
	rechargeAccountImportModeByStudent rechargeAccountImportMode = "by_student"
	rechargeAccountImportModeByAccount rechargeAccountImportMode = "by_account"
)

func detectRechargeAccountImportModeByColumns(columns []string) rechargeAccountImportMode {
	titleSet := make(map[string]struct{}, len(columns))
	for _, title := range columns {
		trimmed := strings.TrimSpace(strings.TrimPrefix(title, "*"))
		if trimmed == "" {
			continue
		}
		titleSet[trimmed] = struct{}{}
	}
	if _, ok := titleSet["储值账户号"]; ok {
		return rechargeAccountImportModeByAccount
	}
	if _, ok := titleSet["学员姓名"]; ok {
		return rechargeAccountImportModeByStudent
	}
	return rechargeAccountImportModeUnknown
}
