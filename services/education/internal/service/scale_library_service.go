package service

import (
	"context"
	"database/sql"
	"errors"
	"sort"
	"strings"
	"time"

	"go-migration-platform/services/education/internal/model"
)

func (svc *Service) GetScaleLibrary(userID int64, query model.ScaleLibraryQuery) (model.ScaleLibraryVO, error) {
	if svc.repo == nil {
		return model.ScaleLibraryVO{}, errors.New("repository is not configured")
	}
	ctx := context.Background()
	instID, err := svc.repo.FindInstIDByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.ScaleLibraryVO{}, errors.New("no institution context")
		}
		return model.ScaleLibraryVO{}, err
	}
	items, err := svc.repo.ListInstitutionScaleLibrary(ctx, instID, query, true)
	if err != nil {
		return model.ScaleLibraryVO{}, err
	}
	filterOptionItems, err := svc.repo.ListScaleLibraryFilterItems(ctx)
	if err != nil {
		return model.ScaleLibraryVO{}, err
	}
	categoryOptions, err := svc.repo.ListScaleLibraryCategoryOptions(ctx)
	if err != nil {
		return model.ScaleLibraryVO{}, err
	}
	return model.ScaleLibraryVO{
		Items:         items,
		Summary:       buildScaleLibrarySummary(items),
		FilterOptions: buildScaleLibraryFilterOptions(filterOptionItems, categoryOptions),
	}, nil
}

func (svc *Service) ListScaleAssessmentStudentCandidates(userID int64, query model.ScaleAssessmentStudentCandidateQuery) (model.PageResult[model.ScaleAssessmentStudentCandidate], error) {
	if svc.repo == nil {
		return model.PageResult[model.ScaleAssessmentStudentCandidate]{}, errors.New("repository is not configured")
	}
	ctx := context.Background()
	instID, err := svc.repo.FindInstIDByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.PageResult[model.ScaleAssessmentStudentCandidate]{}, errors.New("no institution context")
		}
		return model.PageResult[model.ScaleAssessmentStudentCandidate]{}, err
	}
	return svc.repo.ListScaleAssessmentStudentCandidates(ctx, instID, query)
}

func (svc *Service) UpdateScaleAssessmentStudentGender(userID int64, req model.ScaleAssessmentStudentGenderUpdateRequest) (string, error) {
	if svc.repo == nil {
		return "", errors.New("repository is not configured")
	}
	if req.StudentID <= 0 {
		return "", errors.New("studentId is required")
	}
	sex, gender, err := scaleAssessmentStudentGenderValue(req.Gender)
	if err != nil {
		return "", err
	}
	ctx := context.Background()
	instID, err := svc.repo.FindInstIDByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", errors.New("no institution context")
		}
		return "", err
	}
	operatorID, err := svc.repo.FindInstUserIDByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", errors.New("no institution user context")
		}
		return "", err
	}
	beforeGender, _ := svc.repo.GetStudentGenderText(ctx, instID, req.StudentID)
	updated, err := svc.repo.UpdateStudentGender(ctx, instID, req.StudentID, sex, operatorID)
	if err != nil {
		return "", err
	}
	if !updated {
		return "", errors.New("学生信息不存在")
	}
	if strings.TrimSpace(beforeGender) != gender {
		_ = svc.repo.InsertStudentChangeRecord(ctx, instID, req.StudentID, operatorID, `性别从"`+displayStudentChangeValue(beforeGender)+`"修改为"`+gender+`";`)
	}
	return gender, nil
}

func scaleAssessmentStudentGenderValue(value string) (int, string, error) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if normalized == "0" || normalized == "f" || normalized == "female" || strings.Contains(normalized, "女") {
		return 0, "女", nil
	}
	if normalized == "1" || normalized == "m" || normalized == "male" || strings.Contains(normalized, "男") {
		return 1, "男", nil
	}
	return -1, "", errors.New("gender must be 男 or 女")
}

func (svc *Service) UpdateScaleAssessmentStudentBirthDate(userID int64, req model.ScaleAssessmentStudentBirthDateUpdateRequest) (string, error) {
	if svc.repo == nil {
		return "", errors.New("repository is not configured")
	}
	if req.StudentID <= 0 {
		return "", errors.New("studentId is required")
	}
	birthDate, err := scaleAssessmentStudentBirthDateValue(req.BirthDate)
	if err != nil {
		return "", err
	}
	ctx := context.Background()
	instID, err := svc.repo.FindInstIDByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", errors.New("no institution context")
		}
		return "", err
	}
	operatorID, err := svc.repo.FindInstUserIDByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", errors.New("no institution user context")
		}
		return "", err
	}
	updatedBirthDate := birthDate.Format(time.DateOnly)
	beforeBirthDate, _ := svc.repo.GetStudentBirthDateText(ctx, instID, req.StudentID)
	updated, err := svc.repo.UpdateStudentBirthDate(ctx, instID, req.StudentID, birthDate, operatorID)
	if err != nil {
		return "", err
	}
	if !updated {
		return "", errors.New("学生信息不存在")
	}
	if strings.TrimSpace(beforeBirthDate) != updatedBirthDate {
		_ = svc.repo.InsertStudentChangeRecord(ctx, instID, req.StudentID, operatorID, `出生日期从"`+displayStudentChangeValue(beforeBirthDate)+`"修改为"`+updatedBirthDate+`";`)
	}
	return updatedBirthDate, nil
}

func scaleAssessmentStudentBirthDateValue(value string) (time.Time, error) {
	raw := strings.TrimSpace(value)
	if raw == "" {
		return time.Time{}, errors.New("birthDate is required")
	}
	for _, layout := range []string{time.DateOnly, "2006/01/02", time.RFC3339} {
		if parsed, err := time.ParseInLocation(layout, raw, time.Local); err == nil {
			birthDate := time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, time.Local)
			if birthDate.After(time.Now()) {
				return time.Time{}, errors.New("birthDate cannot be in the future")
			}
			return birthDate, nil
		}
	}
	return time.Time{}, errors.New("birthDate format should be YYYY-MM-DD")
}

func (svc *Service) ListScaleCategoryOptions(userID int64) ([]string, error) {
	if svc.repo == nil {
		return nil, errors.New("repository is not configured")
	}
	if _, err := svc.repo.FindInstIDByUserID(context.Background(), userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("no institution context")
		}
		return nil, err
	}
	return svc.repo.ListScaleLibraryCategoryOptions(context.Background())
}

func buildScaleLibrarySummary(items []model.ScaleLibraryItem) model.ScaleLibrarySummary {
	summary := model.ScaleLibrarySummary{Total: len(items)}
	for _, item := range items {
		if item.Status == "available" {
			summary.Available++
		} else {
			summary.Unavailable++
		}
		summary.MonthUsage += item.MonthUsage
		summary.UsageCount += item.UsageCount
		if item.AuthReserved {
			summary.ReservedAuths++
		}
	}
	return summary
}

func buildScaleLibraryFilterOptions(items []model.ScaleLibraryItem, categoryOptions []string) model.ScaleLibraryFilterOptions {
	categories := make(map[string]bool)
	categoryCounts := make(map[string]int)
	scenarios := make(map[string]bool)
	statuses := make(map[string]bool)
	for _, category := range categoryOptions {
		category = strings.TrimSpace(category)
		if category != "" {
			categories[category] = true
			categoryCounts[category] = 0
		}
	}
	for _, item := range items {
		if item.Category != "" {
			categories[item.Category] = true
			categoryCounts[item.Category]++
		}
		if item.Scenario != "" {
			scenarios[item.Scenario] = true
		}
		if item.Status != "" {
			statuses[item.Status] = true
		}
	}
	return model.ScaleLibraryFilterOptions{
		Categories:     sortedScaleLibraryKeys(categories),
		CategoryCounts: categoryCounts,
		Scenarios:      sortedScaleLibraryKeys(scenarios),
		Statuses:       sortedScaleLibraryKeys(statuses),
		AgeScopes:      []string{"all", "0-2", "2-6", "6-12", "12+"},
		Durations:      []string{"0-15", "15-30", "30-60", "60+"},
	}
}

func sortedScaleLibraryKeys(values map[string]bool) []string {
	out := make([]string, 0, len(values))
	for value := range values {
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}
