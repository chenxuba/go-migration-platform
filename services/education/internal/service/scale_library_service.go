package service

import (
	"context"
	"database/sql"
	"errors"
	"sort"

	"go-migration-platform/services/education/internal/model"
)

func (svc *Service) GetScaleLibrary(userID int64, query model.ScaleLibraryQuery) (model.ScaleLibraryVO, error) {
	if svc.repo == nil {
		return model.ScaleLibraryVO{}, errors.New("repository is not configured")
	}
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.ScaleLibraryVO{}, errors.New("no institution context")
		}
		return model.ScaleLibraryVO{}, err
	}
	items, err := svc.repo.ListInstitutionScaleLibrary(context.Background(), instID, query)
	if err != nil {
		return model.ScaleLibraryVO{}, err
	}
	return model.ScaleLibraryVO{
		Items:         items,
		Summary:       buildScaleLibrarySummary(items),
		FilterOptions: buildScaleLibraryFilterOptions(items),
	}, nil
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

func buildScaleLibraryFilterOptions(items []model.ScaleLibraryItem) model.ScaleLibraryFilterOptions {
	categories := make(map[string]bool)
	scenarios := make(map[string]bool)
	statuses := make(map[string]bool)
	for _, item := range items {
		if item.Category != "" {
			categories[item.Category] = true
		}
		if item.Scenario != "" {
			scenarios[item.Scenario] = true
		}
		if item.Status != "" {
			statuses[item.Status] = true
		}
	}
	return model.ScaleLibraryFilterOptions{
		Categories: sortedScaleLibraryKeys(categories),
		Scenarios:  sortedScaleLibraryKeys(scenarios),
		Statuses:   sortedScaleLibraryKeys(statuses),
		AgeScopes:  []string{"all", "0-2", "2-6", "6-12", "12+"},
		Durations:  []string{"0-15", "15-30", "30-60", "60+"},
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
