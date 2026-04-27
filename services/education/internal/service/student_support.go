package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"go-migration-platform/pkg/logx"
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

func (svc *Service) studentImportDuplicateCheck(ctx context.Context, instID int64, stuName, mobile string, excludeID *int64) (int, int, error) {
	rule, err := svc.repo.GetAddImportStudentRule(ctx, instID)
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

func studentWeChatDuplicateMessage() string {
	return "当前机构已存在微信号相同的学员"
}

func (svc *Service) refreshStudentBindChildStatusByPhones(ctx context.Context, phones ...string) {
	if svc == nil || svc.repo == nil {
		return
	}

	seen := make(map[string]struct{}, len(phones))
	for _, phone := range phones {
		trimmed := strings.TrimSpace(phone)
		if trimmed == "" {
			continue
		}
		if _, exists := seen[trimmed]; exists {
			continue
		}
		seen[trimmed] = struct{}{}
		if err := svc.repo.RefreshStudentBindChildStatusByPhone(ctx, trimmed); err != nil {
			logx.Error("refresh student bind child status by phone failed", logx.Entry{
				"phone": trimmed,
				"error": err.Error(),
			})
		}
	}
}

func buildStudentStatusSnapshotAfter(before repository.StudentSnapshot, dtoIntentionLevel, dtoFollowUpStatus *int) repository.StudentSnapshot {
	after := before
	if dtoIntentionLevel != nil {
		value := *dtoIntentionLevel
		after.IntentLevel = &value
	}
	if dtoFollowUpStatus != nil {
		value := *dtoFollowUpStatus
		after.FollowUpStatus = &value
	}
	return after
}

func buildStudentStatusChangeText(before, after repository.StudentSnapshot) string {
	changes := make([]string, 0, 2)
	if studentIntentLevelLabel(before.IntentLevel) != studentIntentLevelLabel(after.IntentLevel) {
		changes = append(changes, fmt.Sprintf(`意向度从"%s"修改为"%s"`, studentIntentLevelLabel(before.IntentLevel), studentIntentLevelLabel(after.IntentLevel)))
	}
	if studentFollowUpStatusLabel(before.FollowUpStatus) != studentFollowUpStatusLabel(after.FollowUpStatus) {
		changes = append(changes, fmt.Sprintf(`跟进状态从"%s"修改为"%s"`, studentFollowUpStatusLabel(before.FollowUpStatus), studentFollowUpStatusLabel(after.FollowUpStatus)))
	}
	if len(changes) == 0 {
		return ""
	}
	return strings.Join(changes, ";") + ";"
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
	appendChange("督导", svc.repo.GetStaffNameByID(ctx, before.SupervisorID), svc.repo.GetStaffNameByID(ctx, after.SupervisorID))
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
