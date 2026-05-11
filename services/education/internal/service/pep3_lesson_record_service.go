package service

import (
	"strings"
	"time"

	"go-migration-platform/services/education/internal/model"
	"go-migration-platform/services/education/internal/repository"
)

func buildPEP3LessonRecordEntities(
	record model.AssessmentRecordDetailVO,
	plan model.PEP3WeeklyPlanAIResult,
	durationMonths int,
	targetMonthIndex int,
	targetWeekIndex int,
	userID int64,
	instID int64,
) []repository.PEP3LessonRecordEntity {
	if durationMonths <= 0 {
		durationMonths = 3
	}
	weekDates := make([]time.Time, 0, len(plan.WeekDates))
	for _, rawDate := range plan.WeekDates {
		parsed := parseIEPPlanDateValue(strings.TrimSpace(rawDate))
		weekDates = append(weekDates, parsed)
	}
	result := make([]repository.PEP3LessonRecordEntity, 0)
	for rowIndex, row := range plan.Rows {
		if strings.TrimSpace(row.Project) == "" &&
			strings.TrimSpace(row.Content) == "" &&
			len(row.Completion) == 0 {
			continue
		}
		for dateIndex, lessonDate := range weekDates {
			if lessonDate.IsZero() {
				continue
			}
			completionCode := ""
			if dateIndex < len(row.Completion) {
				completionCode = strings.TrimSpace(row.Completion[dateIndex])
			}
			if completionCode == "" {
				continue
			}
			result = append(result, repository.PEP3LessonRecordEntity{
				InstID:           instID,
				RecordID:         record.ID,
				StudentID:        record.StudentID,
				StudentName:      strings.TrimSpace(record.StudentName),
				AssessmentCode:   strings.TrimSpace(record.AssessmentCode),
				AssessmentName:   strings.TrimSpace(record.AssessmentName),
				DurationMonths:   durationMonths,
				TargetMonthIndex: targetMonthIndex,
				TargetWeekIndex:  targetWeekIndex,
				LessonDate:       lessonDate,
				WeekDateIndex:    dateIndex + 1,
				WeeklyRowIndex:   rowIndex,
				Project:          strings.TrimSpace(row.Project),
				Content:          strings.TrimSpace(row.Content),
				CompletionCode:   completionCode,
				TeacherName:      strings.TrimSpace(plan.TeacherName),
				CourseName:       strings.TrimSpace(plan.CourseName),
				CreatedBy:        userID,
				UpdatedBy:        userID,
			})
		}
	}
	return result
}
