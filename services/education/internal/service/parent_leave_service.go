package service

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"go-migration-platform/services/education/internal/model"
)

const (
	parentLeaveDefaultPageIndex = 1
	parentLeaveDefaultPageSize  = 20
	parentLeaveMaxPageSize      = 100
)

func (svc *Service) ListParentLeavesByPhone(ctx context.Context, phone string, query model.ParentLeaveQueryDTO) (model.ParentLeaveSummaryVO, error) {
	if svc == nil || svc.repo == nil {
		return model.ParentLeaveSummaryVO{}, errors.New("家长端请假记录服务未初始化")
	}

	phone = normalizeParentPhone(phone)
	if phone == "" {
		return model.ParentLeaveSummaryVO{}, errors.New("手机号不能为空")
	}

	rows, err := svc.repo.ListParentStudentCandidatesByPhone(ctx, phone)
	if err != nil {
		return model.ParentLeaveSummaryVO{}, err
	}
	displayProfiles, err := svc.resolveParentStudentDisplayProfiles(ctx, rows)
	if err != nil {
		return model.ParentLeaveSummaryVO{}, err
	}

	targets := buildParentScheduleTargets(rows, displayProfiles)
	students := buildParentBoundStudents(targets)
	if len(targets) == 0 {
		return model.ParentLeaveSummaryVO{
			Students: students,
			Items:    []model.ParentLeaveVO{},
		}, nil
	}

	target, ok := resolveParentLeaveTarget(targets, query.StudentID)
	if !ok {
		return model.ParentLeaveSummaryVO{}, errors.New("未找到对应学员")
	}

	pageIndex := normalizeParentLeavePageIndex(query.PageIndex)
	pageSize := normalizeParentLeavePageSize(query.PageSize)
	items, total, err := svc.listParentLeaveItemsForTarget(ctx, target, pageIndex, pageSize)
	if err != nil {
		return model.ParentLeaveSummaryVO{}, err
	}

	return model.ParentLeaveSummaryVO{
		Students:  students,
		Items:     items,
		PageIndex: pageIndex,
		PageSize:  pageSize,
		Total:     total,
		HasMore:   pageIndex*pageSize < total,
	}, nil
}

func (svc *Service) listParentLeaveItemsForTarget(ctx context.Context, target parentScheduleTarget, pageIndex, pageSize int) ([]model.ParentLeaveVO, int, error) {
	sourceStudentIDs := []int64{target.StudentID}
	result, err := svc.repo.PageParentLeaveRequestsByStudentIDs(ctx, target.InstID, sourceStudentIDs, pageIndex, pageSize)
	if err != nil {
		return nil, 0, err
	}

	if result.Total <= 0 && shouldFallbackToParentScheduleAliases(target) {
		aliasIDs, err := svc.resolveParentScheduleAliasStudentIDs(ctx, target)
		if err != nil {
			return nil, 0, err
		}
		if len(aliasIDs) > 0 {
			sourceStudentIDs = aliasIDs
			result, err = svc.repo.PageParentLeaveRequestsByStudentIDs(ctx, target.InstID, sourceStudentIDs, pageIndex, pageSize)
			if err != nil {
				return nil, 0, err
			}
		}
	}

	if result.Total <= 0 || len(result.List) == 0 {
		return []model.ParentLeaveVO{}, result.Total, nil
	}

	items := make([]model.ParentLeaveVO, 0, len(result.List))
	for _, item := range result.List {
		items = append(items, buildParentLeaveVO(target, item))
	}
	return items, result.Total, nil
}

func resolveParentLeaveTarget(targets []parentScheduleTarget, studentID string) (parentScheduleTarget, bool) {
	studentID = strings.TrimSpace(studentID)
	if studentID == "" {
		return targets[0], true
	}
	for _, target := range targets {
		if strconv.FormatInt(target.StudentID, 10) == studentID {
			return target, true
		}
	}
	return parentScheduleTarget{}, false
}

func normalizeParentLeavePageIndex(pageIndex int) int {
	if pageIndex <= 0 {
		return parentLeaveDefaultPageIndex
	}
	return pageIndex
}

func normalizeParentLeavePageSize(pageSize int) int {
	if pageSize <= 0 {
		return parentLeaveDefaultPageSize
	}
	if pageSize > parentLeaveMaxPageSize {
		return parentLeaveMaxPageSize
	}
	return pageSize
}

func buildParentLeaveVO(target parentScheduleTarget, item model.LeavePagedItem) model.ParentLeaveVO {
	return model.ParentLeaveVO{
		ID:               strings.TrimSpace(item.ID),
		StudentID:        firstNonEmptyString(strings.TrimSpace(item.StudentID), strconv.FormatInt(target.StudentID, 10)),
		StudentName:      firstNonEmptyString(strings.TrimSpace(item.StudentName), target.StudentName, "学员"),
		StudentAvatarURL: firstNonEmptyString(strings.TrimSpace(item.StudentAvatarURL), target.AvatarURL),
		LeaveType:        item.LeaveType,
		LeaveTypeText:    parentLeaveTypeText(item.LeaveType),
		Status:           item.Status,
		StatusText:       parentLeaveStatusText(item.Status, item.StatusText),
		StartTime:        formatParentLeaveTime(item.StartTime),
		EndTime:          formatParentLeaveTime(item.EndTime),
		ApplyTime:        formatParentLeaveTime(item.ApplyTime),
	}
}

func parentLeaveTypeText(value int) string {
	switch value {
	case model.LeaveTypePersonal:
		return "事假"
	case model.LeaveTypeSick:
		return "病假"
	case model.LeaveTypeSuspend:
		return "休学"
	default:
		return "-"
	}
}

func parentLeaveStatusText(status int, fallback string) string {
	switch status {
	case model.LeaveStatusPending:
		return "待处理"
	case model.LeaveStatusApproved:
		return "已通过"
	case model.LeaveStatusRejected:
		return "已拒绝"
	case model.LeaveStatusRevoked:
		return "已撤销"
	default:
		return firstNonEmptyString(strings.TrimSpace(fallback), "-")
	}
}

func formatParentLeaveTime(value *time.Time) string {
	if value == nil || value.IsZero() {
		return "-"
	}

	weekdayText := ""
	switch value.Weekday() {
	case time.Monday:
		weekdayText = "周一"
	case time.Tuesday:
		weekdayText = "周二"
	case time.Wednesday:
		weekdayText = "周三"
	case time.Thursday:
		weekdayText = "周四"
	case time.Friday:
		weekdayText = "周五"
	case time.Saturday:
		weekdayText = "周六"
	case time.Sunday:
		weekdayText = "周日"
	}
	if weekdayText == "" {
		return value.Format("2006-01-02 15:04")
	}
	return value.Format("2006-01-02") + "（" + weekdayText + "）" + value.Format("15:04")
}
