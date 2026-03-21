package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"go-migration-platform/services/education/internal/repository"
)

func (svc *Service) studentDuplicateCheck(ctx context.Context, instID int64, stuName, mobile string, excludeID *int64) (int, int, error) {
	rule, err := svc.repo.GetAddIntentionStudentRule(ctx, instID)
	if err != nil {
		return 0, 0, err
	}
	count, err := svc.repo.CountStudentDuplicatesByRule(ctx, instID, int64(rule), stuName, mobile, excludeID)
	if err != nil {
		return 0, 0, err
	}
	return rule, count, nil
}

func studentDuplicateMessage(rule int) string {
	switch rule {
	case 2:
		return "当前机构已存在手机号相同的学员"
	case 3:
		return "当前机构已存在姓名相同的学员"
	default:
		return "当前机构已存在姓名和手机号同时相同的学员"
	}
}

func (svc *Service) buildStudentSnapshotChangeText(ctx context.Context, before, after repository.StudentSnapshot) string {
	changes := make([]string, 0, 12)

	appendChange := func(fieldName, oldValue, newValue string) {
		if oldValue == newValue {
			return
		}
		changes = append(changes, fmt.Sprintf(`%s从"%s"修改为"%s"`, fieldName, oldValue, newValue))
	}

	appendChange("姓名", displayStudentChangeValue(before.StuName), displayStudentChangeValue(after.StuName))
	appendChange("手机号码", maskStudentMobile(before.Mobile), maskStudentMobile(after.Mobile))
	appendChange("渠道来源", svc.repo.GetChannelNameByID(ctx, before.ChannelID), svc.repo.GetChannelNameByID(ctx, after.ChannelID))
	appendChange("采单员", svc.repo.GetStaffNameByID(ctx, before.CollectorStaffID), svc.repo.GetStaffNameByID(ctx, after.CollectorStaffID))
	appendChange("电话销售", svc.repo.GetStaffNameByID(ctx, before.PhoneSellStaffID), svc.repo.GetStaffNameByID(ctx, after.PhoneSellStaffID))
	appendChange("前台", svc.repo.GetStaffNameByID(ctx, before.ForegroundStaffID), svc.repo.GetStaffNameByID(ctx, after.ForegroundStaffID))
	appendChange("副销售员", svc.repo.GetStaffNameByID(ctx, before.ViceSellStaffID), svc.repo.GetStaffNameByID(ctx, after.ViceSellStaffID))
	appendChange("学管师", svc.repo.GetStaffNameByID(ctx, before.StudentManagerID), svc.repo.GetStaffNameByID(ctx, after.StudentManagerID))
	appendChange("顾问", svc.repo.GetStaffNameByID(ctx, before.AdvisorID), svc.repo.GetStaffNameByID(ctx, after.AdvisorID))
	appendChange("推荐人", svc.repo.GetStudentNameByID(ctx, before.RecommendStudentID), svc.repo.GetStudentNameByID(ctx, after.RecommendStudentID))
	appendChange("手机关联人关系", studentPhoneRelationshipLabel(before.PhoneRelationship), studentPhoneRelationshipLabel(after.PhoneRelationship))
	appendChange("意向度", studentIntentLevelLabel(before.IntentLevel), studentIntentLevelLabel(after.IntentLevel))
	appendChange("跟进状态", studentFollowUpStatusLabel(before.FollowUpStatus), studentFollowUpStatusLabel(after.FollowUpStatus))

	oldSaleName := salePersonDisplayName(ctx, svc.repo, before.SalePerson)
	newSaleName := salePersonDisplayName(ctx, svc.repo, after.SalePerson)
	if oldSaleName != newSaleName {
		if after.SalePerson == nil {
			changes = append(changes, fmt.Sprintf(`销售员从"%s"修改为"-"，已进入公有池`, oldSaleName))
		} else {
			changes = append(changes, fmt.Sprintf(`销售员从"%s"修改为"%s"`, oldSaleName, newSaleName))
		}
	}

	if len(changes) == 0 {
		return ""
	}
	return strings.Join(changes, ";") + ";"
}

func displayStudentChangeValue(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "-"
	}
	return trimmed
}

func maskStudentMobile(mobile string) string {
	trimmed := strings.TrimSpace(mobile)
	if trimmed == "" {
		return "-"
	}
	if len(trimmed) == 11 {
		return trimmed[:3] + "****" + trimmed[7:]
	}
	return trimmed
}

func studentPhoneRelationshipLabel(value *int) string {
	if value == nil {
		return "-"
	}
	switch *value {
	case 1:
		return "爸爸"
	case 2:
		return "妈妈"
	case 3:
		return "爷爷"
	case 4:
		return "奶奶"
	case 5:
		return "外公"
	case 6:
		return "外婆"
	case 7:
		return "其他"
	default:
		return strconv.Itoa(*value)
	}
}

func studentIntentLevelLabel(value *int) string {
	if value == nil {
		return "-"
	}
	switch *value {
	case 1:
		return "未知"
	case 2:
		return "低"
	case 3:
		return "中"
	case 4:
		return "高"
	default:
		return strconv.Itoa(*value)
	}
}

func studentFollowUpStatusLabel(value *int) string {
	if value == nil {
		return "-"
	}
	switch *value {
	case 0:
		return "待跟进"
	case 1:
		return "跟进中"
	case 2:
		return "未接听"
	case 3:
		return "已邀约"
	case 4:
		return "已试听"
	case 5:
		return "已到访"
	case 6:
		return "已失效"
	default:
		return strconv.Itoa(*value)
	}
}

func salePersonDisplayName(ctx context.Context, repo *repository.Repository, value *int64) string {
	if value == nil {
		return "无"
	}
	return repo.GetStaffNameByID(ctx, value)
}
