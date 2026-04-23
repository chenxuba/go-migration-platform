package service

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"

	"go-migration-platform/pkg/logx"
	"go-migration-platform/services/education/internal/model"
	"go-migration-platform/services/education/internal/repository"
)

const (
	pendingRenewalReminderTemplateName = "续费提醒"
)

func (svc *Service) SendPendingRenewalWechatReminders(userID int64, dto model.PendingRenewalReminderSendDTO) (model.PendingRenewalReminderSendResult, error) {
	ctx := context.Background()
	instID, err := svc.repo.FindInstIDByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.PendingRenewalReminderSendResult{}, errors.New("no institution context")
		}
		return model.PendingRenewalReminderSendResult{}, err
	}
	if svc.wechatOfficial == nil || !svc.wechatOfficial.isEnabled() {
		return model.PendingRenewalReminderSendResult{}, errors.New("公众号模板消息未配置")
	}

	instUserID, err := svc.repo.FindInstUserIDByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.PendingRenewalReminderSendResult{}, errors.New("no institution user context")
		}
		return model.PendingRenewalReminderSendResult{}, err
	}

	tuitionAccountIDs := normalizePendingRenewalReminderIDs(dto.TuitionAccountIDs)
	if len(tuitionAccountIDs) == 0 {
		return model.PendingRenewalReminderSendResult{}, errors.New("请选择需要发送续费提醒的学员")
	}
	if len(tuitionAccountIDs) > 500 {
		return model.PendingRenewalReminderSendResult{}, errors.New("单次最多支持发送500条续费提醒，请分批操作")
	}

	targets, err := svc.repo.ListPendingRenewalReminderTargetsByTuitionAccountIDs(ctx, instID, tuitionAccountIDs)
	if err != nil {
		return model.PendingRenewalReminderSendResult{}, err
	}
	if len(targets) == 0 {
		return model.PendingRenewalReminderSendResult{}, errors.New("未找到可发送的待续费记录，请刷新后重试")
	}

	studentIDs := make([]int64, 0, len(targets))
	studentSeen := make(map[int64]struct{}, len(targets))
	for _, target := range targets {
		if target.StudentID <= 0 {
			continue
		}
		if _, ok := studentSeen[target.StudentID]; ok {
			continue
		}
		studentSeen[target.StudentID] = struct{}{}
		studentIDs = append(studentIDs, target.StudentID)
	}

	recipientMap, recipientCount, err := svc.listWeChatOfficialRecipientMapByStudentIDs(ctx, instID, studentIDs)
	if err != nil {
		return model.PendingRenewalReminderSendResult{}, err
	}

	logx.Info("pending renewal template message recipients resolved", logx.Entry{
		"instId":              instID,
		"userId":              userID,
		"templateId":          weChatOfficialTemplateIDPendingRenewalUnified,
		"targetCount":         len(targets),
		"recipientCount":      recipientCount,
		"recipientGroupCount": len(recipientMap),
	})

	operatorName := strings.TrimSpace(svc.repo.GetStaffNameByID(ctx, &instUserID))
	if operatorName == "-" {
		operatorName = ""
	}
	institutionName, err := svc.repo.GetInstitutionName(ctx, instID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return model.PendingRenewalReminderSendResult{}, err
	}
	institutionName = firstNonEmptyString(strings.TrimSpace(institutionName), svc.weChatOfficialAccountName(), "机构")

	recordItems := make([]repository.TemplateMessageRecordItemCreateInput, 0, len(targets))
	result := model.PendingRenewalReminderSendResult{
		NotifyCount: len(targets),
	}

	for _, target := range targets {
		recordItem := repository.TemplateMessageRecordItemCreateInput{
			BusinessType: model.TemplateMessageBusinessTypePendingRenewal,
			Channel:      model.TemplateMessageChannelWeChat,
			RelationID:   strings.TrimSpace(target.TuitionAccountID),
			StudentID:    target.StudentID,
			StudentName:  strings.TrimSpace(target.StudentName),
			Sex:          target.Sex,
			Avatar:       strings.TrimSpace(target.Avatar),
			Phone:        strings.TrimSpace(target.Phone),
			BizTitle:     firstNonEmptyString(strings.TrimSpace(target.LessonName), "课程"),
			BizSummary:   buildPendingRenewalReminderRemainingText(target),
		}

		openIDs := recipientMap[target.StudentID]
		if len(openIDs) == 0 {
			recordItem.Status = model.TemplateMessageRecordItemStatusSkipped
			recordItem.StatusReason = "未关注公众号，已跳过发送"
			result.SkippedCount++
			recordItems = append(recordItems, recordItem)
			continue
		}

		recordItem.RecipientCount = len(openIDs)
		successRecipientCount := 0
		var lastErr error
		for _, openID := range openIDs {
			request, err := svc.buildWeChatOfficialPendingRenewalTemplateRequest(openID, target, institutionName)
			if err != nil {
				lastErr = err
				continue
			}
			if err := svc.wechatOfficial.sendTemplateMessage(ctx, request); err != nil {
				lastErr = err
				continue
			}
			successRecipientCount++
		}

		recordItem.SuccessRecipientCount = successRecipientCount
		switch {
		case successRecipientCount > 0:
			recordItem.Status = model.TemplateMessageRecordItemStatusSuccess
			if successRecipientCount < len(openIDs) && lastErr != nil {
				recordItem.StatusReason = truncateRunesWithEllipsis("部分接收方发送失败："+lastErr.Error(), 200)
			}
			result.SuccessCount++
		default:
			recordItem.Status = model.TemplateMessageRecordItemStatusFailed
			if lastErr != nil {
				recordItem.StatusReason = truncateRunesWithEllipsis(lastErr.Error(), 200)
			}
			if strings.TrimSpace(recordItem.StatusReason) == "" {
				recordItem.StatusReason = "发送失败"
			}
			result.FailedCount++
		}
		recordItems = append(recordItems, recordItem)
	}

	recordID, err := svc.repo.CreateTemplateMessageRecordWithItems(ctx, instID, repository.TemplateMessageRecordCreateInput{
		BusinessType: model.TemplateMessageBusinessTypePendingRenewal,
		Channel:      model.TemplateMessageChannelWeChat,
		TemplateID:   weChatOfficialTemplateIDPendingRenewalUnified,
		TemplateName: pendingRenewalReminderTemplateName,
		NotifyCount:  result.NotifyCount,
		SuccessCount: result.SuccessCount,
		SkippedCount: result.SkippedCount,
		FailedCount:  result.FailedCount,
		OperatorID:   instUserID,
		OperatorName: operatorName,
	}, recordItems)
	if err != nil {
		return model.PendingRenewalReminderSendResult{}, err
	}

	result.RecordID = strconv.FormatInt(recordID, 10)
	return result, nil
}

func (svc *Service) PagePendingRenewalReminderRecords(userID int64, query model.PendingRenewalReminderRecordPageQueryDTO) (model.PendingRenewalReminderRecordPageResult, error) {
	ctx := context.Background()
	instID, err := svc.repo.FindInstIDByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.PendingRenewalReminderRecordPageResult{}, errors.New("no institution context")
		}
		return model.PendingRenewalReminderRecordPageResult{}, err
	}

	records, total, err := svc.repo.PageTemplateMessageRecords(ctx, instID, model.TemplateMessageBusinessTypePendingRenewal, query.PageRequestModel)
	if err != nil {
		return model.PendingRenewalReminderRecordPageResult{}, err
	}

	items := make([]model.PendingRenewalReminderRecordPageItem, 0, len(records))
	for _, record := range records {
		items = append(items, formatPendingRenewalReminderRecord(record))
	}
	return model.PendingRenewalReminderRecordPageResult{
		List:  items,
		Total: total,
	}, nil
}

func (svc *Service) GetPendingRenewalReminderRecordDetail(userID int64, recordIDRaw string) (model.PendingRenewalReminderRecordDetailResult, error) {
	ctx := context.Background()
	instID, err := svc.repo.FindInstIDByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.PendingRenewalReminderRecordDetailResult{}, errors.New("no institution context")
		}
		return model.PendingRenewalReminderRecordDetailResult{}, err
	}

	recordID, err := strconv.ParseInt(strings.TrimSpace(recordIDRaw), 10, 64)
	if err != nil || recordID <= 0 {
		return model.PendingRenewalReminderRecordDetailResult{}, errors.New("消息记录ID无效")
	}

	record, err := svc.repo.GetTemplateMessageRecord(ctx, instID, recordID, model.TemplateMessageBusinessTypePendingRenewal)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.PendingRenewalReminderRecordDetailResult{}, errors.New("消息记录不存在")
		}
		return model.PendingRenewalReminderRecordDetailResult{}, err
	}

	recordItems, err := svc.repo.ListTemplateMessageRecordItems(ctx, instID, recordID)
	if err != nil {
		return model.PendingRenewalReminderRecordDetailResult{}, err
	}

	result := model.PendingRenewalReminderRecordDetailResult{
		RecordID:     strconv.FormatInt(record.ID, 10),
		TemplateID:   record.TemplateID,
		TemplateName: record.TemplateName,
		Channel:      record.Channel,
		ChannelName:  templateMessageChannelName(record.Channel),
		ReadCount:    0,
		NotifyCount:  record.NotifyCount,
		SuccessCount: record.SuccessCount,
		SkippedCount: record.SkippedCount,
		FailedCount:  record.FailedCount,
		UnsentCount:  record.SkippedCount + record.FailedCount,
		OperatorID:   formatPendingRenewalReminderID(record.OperatorID),
		OperatorName: record.OperatorName,
		SendTime:     record.CreateTime,
		SentList:     make([]model.PendingRenewalReminderRecordDetailItem, 0, len(recordItems)),
		UnsentList:   make([]model.PendingRenewalReminderRecordDetailItem, 0, len(recordItems)),
	}

	for _, item := range recordItems {
		homeSchoolStatus, homeSchoolStatusText := templateMessageRecordItemHomeSchoolStatus(item)
		formatted := model.PendingRenewalReminderRecordDetailItem{
			ItemID:               strconv.FormatInt(item.ID, 10),
			TuitionAccountID:     strings.TrimSpace(item.RelationID),
			StudentID:            formatPendingRenewalReminderID(item.StudentID),
			StudentName:          strings.TrimSpace(item.StudentName),
			Sex:                  item.Sex,
			Avatar:               strings.TrimSpace(item.Avatar),
			Phone:                strings.TrimSpace(item.Phone),
			LessonName:           strings.TrimSpace(item.BizTitle),
			RemainingText:        strings.TrimSpace(item.BizSummary),
			HomeSchoolStatus:     homeSchoolStatus,
			HomeSchoolStatusText: homeSchoolStatusText,
		}
		result.UnsentList = append(result.UnsentList, formatted)
	}
	return result, nil
}

func (svc *Service) buildWeChatOfficialPendingRenewalTemplateRequest(openID string, target repository.PendingRenewalReminderTarget, institutionName string) (weChatOfficialTemplateSendRequest, error) {
	if strings.TrimSpace(openID) == "" {
		return weChatOfficialTemplateSendRequest{}, errors.New("send template message failed: empty openid")
	}

	return weChatOfficialTemplateSendRequest{
		ToUser:          strings.TrimSpace(openID),
		TemplateID:      weChatOfficialTemplateIDPendingRenewalUnified,
		ClientMessageID: "pending_renewal_reminder_" + strings.TrimSpace(target.TuitionAccountID),
		Data: map[string]weChatOfficialTemplateDataItem{
			weChatOfficialPendingRenewalKeywordStudentName: {
				Value: truncateRunes(firstNonEmptyString(strings.TrimSpace(target.StudentName), "学员"), 20),
			},
			weChatOfficialPendingRenewalKeywordCourseName: {
				Value: truncateRunesWithEllipsis(buildPendingRenewalTemplateCourseLine(target), 20),
			},
			weChatOfficialPendingRenewalKeywordProjectName: {
				Value: truncateRunesWithEllipsis(buildPendingRenewalTemplateProjectLine(target), 20),
			},
			weChatOfficialPendingRenewalKeywordInstitution: {
				Value: truncateRunesWithEllipsis(firstNonEmptyString(strings.TrimSpace(institutionName), svc.weChatOfficialAccountName(), "机构"), 20),
			},
			weChatOfficialPendingRenewalKeywordPublishTime: {
				Value: formatWeChatOfficialCoursePurchaseTime(nil),
			},
		},
	}, nil
}

func buildPendingRenewalTemplateCourseLine(target repository.PendingRenewalReminderTarget) string {
	courseName := firstNonEmptyString(strings.TrimSpace(target.LessonName), "课程")
	return firstNonEmptyString(
		courseName+"，"+buildPendingRenewalTemplateReasonSuffix(target),
		buildPendingRenewalTemplateReasonSuffix(target),
		"课程续费提醒",
	)
}

func buildPendingRenewalTemplateProjectLine(target repository.PendingRenewalReminderTarget) string {
	return buildPendingRenewalReminderRemainingText(target) + "，请及时续费"
}

func buildPendingRenewalTemplateReasonSuffix(target repository.PendingRenewalReminderTarget) string {
	switch target.LessonChargingMode {
	case 2:
		if pendingRenewalExpireOnly(target) {
			return "剩余天数即将到期"
		}
		return "剩余天数不足"
	case 3, 4:
		if pendingRenewalExpireOnly(target) {
			return "剩余金额即将到期"
		}
		return "剩余金额不足"
	default:
		if pendingRenewalExpireOnly(target) {
			return "剩余课时即将到期"
		}
		return "剩余课时不足"
	}
}

func pendingRenewalExpireOnly(target repository.PendingRenewalReminderTarget) bool {
	if !target.EnableExpireTime || target.ExpireTime == nil || target.ExpireTime.Year() <= 1 {
		return false
	}
	if target.LessonChargingMode == 3 || target.LessonChargingMode == 4 {
		return clampPendingRenewalReminderAmount(target.Tuition) >= 500
	}
	return target.LeftQuantity+target.LeftFreeQuantity >= 15
}

func buildPendingRenewalReminderRemainingText(target repository.PendingRenewalReminderTarget) string {
	if target.LessonChargingMode == 3 || target.LessonChargingMode == 4 {
		return "剩余" + formatWeChatOfficialQuantity(clampPendingRenewalReminderAmount(target.Tuition)) + "元"
	}

	total := target.LeftQuantity + target.LeftFreeQuantity
	if target.LessonChargingMode == 2 {
		return "剩余" + formatWeChatOfficialQuantity(total) + "天"
	}
	return "剩余" + formatWeChatOfficialQuantity(total) + "课时"
}

func clampPendingRenewalReminderAmount(value float64) float64 {
	if value < 0 {
		return 0
	}
	return value
}

func normalizePendingRenewalReminderIDs(values []string) []string {
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, raw := range values {
		value := strings.TrimSpace(raw)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func formatPendingRenewalReminderRecord(record repository.TemplateMessageRecord) model.PendingRenewalReminderRecordPageItem {
	return model.PendingRenewalReminderRecordPageItem{
		RecordID:     strconv.FormatInt(record.ID, 10),
		TemplateID:   strings.TrimSpace(record.TemplateID),
		TemplateName: firstNonEmptyString(strings.TrimSpace(record.TemplateName), pendingRenewalReminderTemplateName),
		Channel:      record.Channel,
		ChannelName:  templateMessageChannelName(record.Channel),
		ReadCount:    0,
		NotifyCount:  record.NotifyCount,
		SuccessCount: record.SuccessCount,
		SkippedCount: record.SkippedCount,
		FailedCount:  record.FailedCount,
		OperatorID:   formatPendingRenewalReminderID(record.OperatorID),
		OperatorName: strings.TrimSpace(record.OperatorName),
		SendTime:     record.CreateTime,
		UnsentCount:  record.SkippedCount + record.FailedCount,
	}
}

func templateMessageChannelName(channel int) string {
	switch channel {
	case model.TemplateMessageChannelWeChat:
		return "微信提醒"
	case model.TemplateMessageChannelSMS:
		return "短信提醒"
	default:
		return "未知"
	}
}

func templateMessageRecordItemHomeSchoolStatus(item repository.TemplateMessageRecordItem) (int, string) {
	if isPendingRenewalReminderRecordItemUnfollowed(item) {
		return model.PendingRenewalReminderHomeSchoolStatusUnfollowed, "未关注"
	}
	if item.RecipientCount > 0 || item.SuccessRecipientCount > 0 || item.Status == model.TemplateMessageRecordItemStatusSuccess || item.Status == model.TemplateMessageRecordItemStatusFailed {
		return model.PendingRenewalReminderHomeSchoolStatusFollowed, "已关注"
	}
	return model.PendingRenewalReminderHomeSchoolStatusUnknown, "-"
}

func isPendingRenewalReminderRecordItemUnfollowed(item repository.TemplateMessageRecordItem) bool {
	if item.Status != model.TemplateMessageRecordItemStatusSkipped {
		return false
	}
	reason := strings.TrimSpace(item.StatusReason)
	if strings.Contains(reason, "未关注") {
		return true
	}
	return item.RecipientCount == 0 && item.SuccessRecipientCount == 0
}

func formatPendingRenewalReminderID(value int64) string {
	if value <= 0 {
		return ""
	}
	return strconv.FormatInt(value, 10)
}
