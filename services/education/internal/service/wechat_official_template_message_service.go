package service

import (
	"context"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go-migration-platform/pkg/logx"
	"go-migration-platform/services/education/internal/model"
)

const (
	weChatOfficialTemplateIDCourseConsumeComplete = "LZS6dGiiSBTAukYGupsFz1ef4M9L5JXFBZwgB6KQCzc"

	weChatOfficialCourseConsumeKeywordLessonDetail = "thing10"
	weChatOfficialCourseConsumeKeywordStudentName  = "thing17"
	weChatOfficialCourseConsumeKeywordCourseName   = "thing2"
	weChatOfficialCourseConsumeKeywordLessonTime   = "time8"
)

func (svc *Service) dispatchRollCallCourseConsumeCompleteNotifications(ctx context.Context, instID int64, confirmResult model.RollCallConfirmResult) error {
	if svc == nil || svc.repo == nil || svc.wechatOfficial == nil || !svc.wechatOfficial.isEnabled() {
		return nil
	}

	teachingRecordID := strings.TrimSpace(confirmResult.ID)
	if teachingRecordID == "" || len(confirmResult.InsertedStudentTeachingRecordIDs) == 0 {
		return nil
	}

	detail, err := svc.repo.GetTeachingRecordDetail(ctx, instID, model.TeachingRecordDetailQueryDTO{
		TeachingRecordID: teachingRecordID,
	})
	if err != nil {
		return err
	}
	if strings.TrimSpace(detail.TeachingRecordID) == "" {
		return nil
	}

	insertedRecordIDSet := make(map[string]struct{}, len(confirmResult.InsertedStudentTeachingRecordIDs))
	for _, item := range confirmResult.InsertedStudentTeachingRecordIDs {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		insertedRecordIDSet[item] = struct{}{}
	}
	if len(insertedRecordIDSet) == 0 {
		return nil
	}

	targetStudents := make([]model.TeachingRecordDetailStudent, 0, len(insertedRecordIDSet))
	studentIDs := make([]int64, 0, len(insertedRecordIDSet))
	for _, item := range detail.StudentList {
		recordID := strings.TrimSpace(item.StudentTeachingRecordID)
		if _, ok := insertedRecordIDSet[recordID]; !ok {
			continue
		}
		if !shouldSendWeChatOfficialCourseConsumeNotification(item) {
			continue
		}
		studentID, err := strconv.ParseInt(strings.TrimSpace(item.StudentID), 10, 64)
		if err != nil || studentID <= 0 {
			continue
		}
		targetStudents = append(targetStudents, item)
		studentIDs = append(studentIDs, studentID)
	}
	if len(targetStudents) == 0 {
		return nil
	}

	recipients, err := svc.repo.ListSubscribedWeChatOfficialRecipientsByStudentIDs(ctx, instID, studentIDs)
	if err != nil {
		return err
	}

	recipientMap := make(map[int64][]string, len(studentIDs))
	recipientSeen := make(map[string]struct{}, len(recipients))
	for _, item := range recipients {
		openID := strings.TrimSpace(item.OfficialOpenID)
		if item.StudentID <= 0 || openID == "" {
			continue
		}
		key := strconv.FormatInt(item.StudentID, 10) + "|" + openID
		if _, ok := recipientSeen[key]; ok {
			continue
		}
		recipientSeen[key] = struct{}{}
		recipientMap[item.StudentID] = append(recipientMap[item.StudentID], openID)
	}

	logx.Info("roll call template message recipients resolved", logx.Entry{
		"instId":              instID,
		"teachingRecordId":    teachingRecordID,
		"templateId":          weChatOfficialTemplateIDCourseConsumeComplete,
		"targetStudentCount":  len(targetStudents),
		"recipientCount":      len(recipients),
		"recipientGroupCount": len(recipientMap),
	})

	var firstErr error
	tuitionAccountReadingCache := make(map[int64]model.TuitionAccountReadingListResult, len(targetStudents))
	for _, student := range targetStudents {
		studentID, err := strconv.ParseInt(strings.TrimSpace(student.StudentID), 10, 64)
		if err != nil || studentID <= 0 {
			continue
		}
		openIDs := recipientMap[studentID]
		if len(openIDs) == 0 {
			logx.Info("roll call template message skipped because no subscribed official recipient", logx.Entry{
				"instId":                  instID,
				"teachingRecordId":        teachingRecordID,
				"studentId":               studentID,
				"studentTeachingRecordId": strings.TrimSpace(student.StudentTeachingRecordID),
				"templateId":              weChatOfficialTemplateIDCourseConsumeComplete,
			})
			continue
		}

		totalArrearQuantity := student.ArrearQuantity
		if student.ArrearQuantity > 0 {
			totalArrearQuantity, err = svc.resolveWeChatOfficialCourseConsumeTotalArrearQuantity(ctx, instID, detail, student, tuitionAccountReadingCache)
			if err != nil {
				logx.Error("roll call template message failed to resolve total arrear quantity, fallback to current record arrear quantity", logx.Entry{
					"instId":                  instID,
					"teachingRecordId":        teachingRecordID,
					"studentId":               studentID,
					"studentTeachingRecordId": strings.TrimSpace(student.StudentTeachingRecordID),
					"lessonId":                strings.TrimSpace(detail.LessonID),
					"skuMode":                 student.SkuMode,
					"error":                   err.Error(),
				})
				totalArrearQuantity = student.ArrearQuantity
			}
		}

		for _, openID := range openIDs {
			request, err := svc.buildWeChatOfficialCourseConsumeCompleteTemplateRequest(openID, detail, student, totalArrearQuantity)
			if err != nil {
				if firstErr == nil {
					firstErr = err
				}
				continue
			}
			if err := svc.wechatOfficial.sendTemplateMessage(ctx, request); err != nil && firstErr == nil {
				firstErr = err
			}
		}
	}

	return firstErr
}

func (svc *Service) buildWeChatOfficialCourseConsumeCompleteTemplateRequest(openID string, detail model.TeachingRecordDetailResult, student model.TeachingRecordDetailStudent, totalArrearQuantity float64) (weChatOfficialTemplateSendRequest, error) {
	pagePath := buildWeChatOfficialCourseConsumeDetailPagePath(student.StudentID, student.StudentTeachingRecordID)
	request := weChatOfficialTemplateSendRequest{
		ToUser:          strings.TrimSpace(openID),
		TemplateID:      weChatOfficialTemplateIDCourseConsumeComplete,
		ClientMessageID: "roll_call_consume_complete_" + strings.TrimSpace(student.StudentTeachingRecordID),
		Data: map[string]weChatOfficialTemplateDataItem{
			weChatOfficialCourseConsumeKeywordLessonDetail: {
				Value: truncateRunes(buildWeChatOfficialCourseConsumeLessonDetail(student.ActualDeduct, totalArrearQuantity, student.LeftQuantity), 20),
			},
			weChatOfficialCourseConsumeKeywordStudentName: {
				Value: truncateRunes(strings.TrimSpace(student.StudentName), 20),
			},
			weChatOfficialCourseConsumeKeywordCourseName: {
				Value: truncateRunes(buildWeChatOfficialCourseConsumeCourseName(detail), 20),
			},
			weChatOfficialCourseConsumeKeywordLessonTime: {
				Value: formatWeChatOfficialCourseConsumeLessonTime(detail.StartTime, detail.EndTime),
			},
		},
	}

	if svc != nil && svc.wechatOfficial != nil {
		appID := strings.TrimSpace(svc.wechatOfficial.config.MiniProgramAppID)
		if appID != "" && pagePath != "" {
			request.MiniProgram = &weChatOfficialTemplateMiniProgram{
				AppID:    appID,
				PagePath: pagePath,
			}
		}
	}

	return request, nil
}

func buildWeChatOfficialCourseConsumeDetailPagePath(studentID, studentTeachingRecordID string) string {
	studentID = strings.TrimSpace(studentID)
	studentTeachingRecordID = strings.TrimSpace(studentTeachingRecordID)
	if studentID == "" || studentTeachingRecordID == "" {
		return ""
	}

	values := url.Values{}
	values.Set("studentId", studentID)
	values.Set("studentTeachingRecordId", studentTeachingRecordID)
	return "pages/attendance-record/detail?" + values.Encode()
}

func buildWeChatOfficialCourseConsumeCourseName(detail model.TeachingRecordDetailResult) string {
	return firstNonEmptyString(
		strings.TrimSpace(detail.LessonName),
		strings.TrimSpace(detail.SubjectName),
		"课程",
	)
}

func (svc *Service) resolveWeChatOfficialCourseConsumeTotalArrearQuantity(ctx context.Context, instID int64, detail model.TeachingRecordDetailResult, student model.TeachingRecordDetailStudent, cache map[int64]model.TuitionAccountReadingListResult) (float64, error) {
	studentID, err := strconv.ParseInt(strings.TrimSpace(student.StudentID), 10, 64)
	if err != nil || studentID <= 0 {
		return student.ArrearQuantity, nil
	}

	readingList, ok := cache[studentID]
	if !ok {
		readingList, err = svc.repo.GetTuitionAccountReadingList(ctx, instID, strconv.FormatInt(studentID, 10))
		if err != nil {
			return 0, err
		}
		cache[studentID] = readingList
	}

	if totalArrearQuantity, found := findWeChatOfficialCourseConsumeTotalArrearQuantity(readingList.List, detail.LessonID, student.SkuMode); found {
		if totalArrearQuantity < student.ArrearQuantity {
			return student.ArrearQuantity, nil
		}
		return totalArrearQuantity, nil
	}

	return student.ArrearQuantity, nil
}

func shouldSendWeChatOfficialCourseConsumeNotification(student model.TeachingRecordDetailStudent) bool {
	return student.ActualDeduct > 0 || student.ArrearQuantity > 0
}

func findWeChatOfficialCourseConsumeTotalArrearQuantity(items []model.TuitionAccountReadingItem, lessonID string, skuMode int) (float64, bool) {
	lessonID = strings.TrimSpace(lessonID)
	chargingMode := normalizeWeChatOfficialCourseConsumeChargingMode(skuMode)
	if lessonID == "" || chargingMode == 0 {
		return 0, false
	}

	for _, item := range items {
		if strings.TrimSpace(item.LessonID) != lessonID {
			continue
		}
		if normalizeWeChatOfficialCourseConsumeChargingMode(nullableWeChatOfficialCourseConsumeChargingMode(item.LessonChargingMode)) != chargingMode {
			continue
		}
		if item.LessonConsumeArrearQuantity < 0 {
			return 0, true
		}
		return item.LessonConsumeArrearQuantity, true
	}

	return 0, false
}

func normalizeWeChatOfficialCourseConsumeChargingMode(mode int) int {
	if mode == 4 {
		return 3
	}
	switch mode {
	case 1, 2, 3:
		return mode
	default:
		return 0
	}
}

func nullableWeChatOfficialCourseConsumeChargingMode(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}

func buildWeChatOfficialCourseConsumeLessonDetail(actualDeduct, arrearQuantity, leftQuantity float64) string {
	if arrearQuantity > 0 {
		return "消耗" + formatWeChatOfficialQuantity(actualDeduct) + "课时，欠费" + formatWeChatOfficialQuantity(arrearQuantity) + "课时"
	}
	return "消耗" + formatWeChatOfficialQuantity(actualDeduct) + "课时，剩余" + formatWeChatOfficialQuantity(leftQuantity) + "课时"
}

func formatWeChatOfficialQuantity(value float64) string {
	if value < 0 {
		value = 0
	}
	return strconv.FormatFloat(value, 'f', -1, 64)
}

func formatWeChatOfficialCourseConsumeLessonTime(startTime, endTime string) string {
	start, startOK := parseWeChatOfficialCourseConsumeTime(startTime)
	end, endOK := parseWeChatOfficialCourseConsumeTime(endTime)

	switch {
	case startOK && endOK:
		if start.Format("2006-01-02") == end.Format("2006-01-02") {
			return start.Format("2006-01-02 15:04") + "~" + end.Format("15:04")
		}
		return start.Format("2006-01-02 15:04") + "~" + end.Format("2006-01-02 15:04")
	case startOK:
		return start.Format("2006-01-02 15:04")
	default:
		return strings.TrimSpace(startTime)
	}
}

func parseWeChatOfficialCourseConsumeTime(value string) (time.Time, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, false
	}

	layouts := []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
	}
	for _, layout := range layouts {
		if parsed, err := time.ParseInLocation(layout, value, noticeTimeLocation()); err == nil {
			return parsed, true
		}
	}
	return time.Time{}, false
}

func truncateRunes(value string, limit int) string {
	value = strings.TrimSpace(value)
	if limit <= 0 || value == "" {
		return ""
	}

	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit])
}
