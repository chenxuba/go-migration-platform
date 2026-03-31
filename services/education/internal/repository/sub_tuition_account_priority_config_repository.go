package repository

import (
	"context"

	"go-migration-platform/services/education/internal/model"
)

func defaultSubTuitionAccountPriorityConfigs() []model.SubTuitionAccountPriorityConfigItem {
	return []model.SubTuitionAccountPriorityConfigItem{
		{
			PriorityType:  1,
			SortDirection: 1,
			SortWeight:    100,
			IsEnabled:     true,
		},
		{
			PriorityType:  2,
			SortDirection: 1,
			SortWeight:    200,
			IsEnabled:     false,
		},
		{
			PriorityType:  3,
			SortDirection: 1,
			SortWeight:    300,
			IsEnabled:     false,
		},
	}
}

func (repo *Repository) ListSubTuitionAccountPriorityConfigs(ctx context.Context, instID int64) (model.SubTuitionAccountPriorityConfigResult, error) {
	_ = ctx
	_ = instID
	return model.SubTuitionAccountPriorityConfigResult{
		List: defaultSubTuitionAccountPriorityConfigs(),
	}, nil
}
