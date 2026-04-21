package repository

import (
	"context"
	"strings"

	"go-migration-platform/services/education/internal/model"
)

type ParentTimeSlotConsumeRecord struct {
	StudentID int64
	LessonID  int64
	Date      string
	Quantity  float64
}

func (repo *Repository) ListParentTimeSlotConsumeRecords(ctx context.Context, instID int64, studentIDs []int64, startDate, endDate string) ([]ParentTimeSlotConsumeRecord, error) {
	if instID <= 0 {
		return []ParentTimeSlotConsumeRecord{}, nil
	}
	startDate = strings.TrimSpace(startDate)
	endDate = strings.TrimSpace(endDate)
	if startDate == "" || endDate == "" {
		return []ParentTimeSlotConsumeRecord{}, nil
	}

	normalizedIDs := make([]int64, 0, len(studentIDs))
	seen := make(map[int64]struct{}, len(studentIDs))
	for _, studentID := range studentIDs {
		if studentID <= 0 {
			continue
		}
		if _, exists := seen[studentID]; exists {
			continue
		}
		seen[studentID] = struct{}{}
		normalizedIDs = append(normalizedIDs, studentID)
	}
	if len(normalizedIDs) == 0 {
		return []ParentTimeSlotConsumeRecord{}, nil
	}

	args := make([]any, 0, len(normalizedIDs)+4)
	args = append(args, instID, model.TuitionAccountFlowSourceAutoConsume)
	for _, studentID := range normalizedIDs {
		args = append(args, studentID)
	}
	args = append(args, startDate, endDate)

	rows, err := repo.db.QueryContext(ctx, `
		SELECT
			student_id,
			product_id,
			DATE_FORMAT(
				STR_TO_DATE(
					LPAD(MOD(ABS(source_id), 100000000), 8, '0'),
					'%Y%m%d'
				),
				'%Y-%m-%d'
			) AS consume_date,
			IFNULL(SUM(quantity), 0)
		FROM tuition_account_flow
		WHERE inst_id = ?
		  AND del_flag = 0
		  AND source_type = ?
		  AND IFNULL(lesson_charging_mode, 0) = 2
		  AND ABS(source_id) >= 20000101
		  AND student_id IN (`+sqlPlaceholders(len(normalizedIDs))+`)
		  AND STR_TO_DATE(
				LPAD(MOD(ABS(source_id), 100000000), 8, '0'),
				'%Y%m%d'
			) BETWEEN ? AND ?
		GROUP BY student_id, product_id, consume_date
		ORDER BY consume_date DESC, product_id ASC, student_id ASC
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]ParentTimeSlotConsumeRecord, 0, len(normalizedIDs)*4)
	for rows.Next() {
		var item ParentTimeSlotConsumeRecord
		if err := rows.Scan(&item.StudentID, &item.LessonID, &item.Date, &item.Quantity); err != nil {
			return nil, err
		}
		if item.StudentID <= 0 || strings.TrimSpace(item.Date) == "" {
			continue
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
