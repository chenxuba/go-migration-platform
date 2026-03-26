package service

import (
	"strings"

	"go-migration-platform/services/education/internal/model"
)

type orderImportMode string

const (
	orderImportModeUnknown    orderImportMode = ""
	orderImportModeLessonHour orderImportMode = "lesson_hour"
	orderImportModeTimeSlot   orderImportMode = "time_slot"
	orderImportModeAmount     orderImportMode = "amount"
)

func detectOrderImportModeByColumns(columns []model.IntentionStudentImportColumn) orderImportMode {
	titles := make([]string, 0, len(columns))
	for _, column := range columns {
		titles = append(titles, column.Title)
	}
	return detectOrderImportModeByTitles(titles)
}

func detectOrderImportModeByTitles(titles []string) orderImportMode {
	titleSet := make(map[string]struct{}, len(titles))
	for _, title := range titles {
		trimmed := strings.TrimSpace(strings.TrimPrefix(title, "*"))
		if trimmed == "" {
			continue
		}
		titleSet[trimmed] = struct{}{}
	}

	switch {
	case hasImportColumnTitle(titleSet, "购买课时数"):
		return orderImportModeLessonHour
	case hasImportColumnTitle(titleSet, "有效开始日期") || hasImportColumnTitle(titleSet, "有效结束日期(含赠送天数)"):
		return orderImportModeTimeSlot
	case hasImportColumnTitle(titleSet, "购买金额"):
		return orderImportModeAmount
	default:
		return orderImportModeUnknown
	}
}

func hasImportColumnTitle(titleSet map[string]struct{}, title string) bool {
	_, ok := titleSet[strings.TrimSpace(title)]
	return ok
}

func lessonModelByOrderImportMode(mode orderImportMode) int {
	switch mode {
	case orderImportModeTimeSlot:
		return 2
	case orderImportModeAmount:
		return 3
	default:
		return 1
	}
}

func orderImportModeLabel(mode orderImportMode) string {
	switch mode {
	case orderImportModeTimeSlot:
		return "按时段"
	case orderImportModeAmount:
		return "按金额"
	default:
		return "按课时"
	}
}
