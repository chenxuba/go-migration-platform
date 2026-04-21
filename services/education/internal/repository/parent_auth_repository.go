package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

type ParentStudentLookupRecord struct {
	StudentID         int64
	InstID            int64
	StudentName       string
	AvatarURL         string
	Mobile            string
	StudentStatus     int
	PhoneRelationship int
	IsBound           bool
	InstitutionName   string
	InstitutionLogo   string
}

func (repo *Repository) ListParentStudentCandidatesByPhone(ctx context.Context, phone string) ([]ParentStudentLookupRecord, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT s.id, s.inst_id, IFNULL(s.stu_name, ''), IFNULL(s.avatar_url, ''), IFNULL(s.mobile, ''),
		       IFNULL(s.student_status, 0), IFNULL(s.phone_relationship, 0), IFNULL(s.is_bind_child, 0),
		       IFNULL(i.organ_name, ''), IFNULL(i.logo, '')
		FROM inst_student s
		LEFT JOIN org_institution i ON i.id = s.inst_id
		WHERE s.del_flag = 0 AND IFNULL(s.mobile, '') = ?
		ORDER BY s.create_time DESC, s.id DESC
	`, strings.TrimSpace(phone))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]ParentStudentLookupRecord, 0, 4)
	for rows.Next() {
		var item ParentStudentLookupRecord
		var isBound int
		if err := rows.Scan(
			&item.StudentID,
			&item.InstID,
			&item.StudentName,
			&item.AvatarURL,
			&item.Mobile,
			&item.StudentStatus,
			&item.PhoneRelationship,
			&isBound,
			&item.InstitutionName,
			&item.InstitutionLogo,
		); err != nil {
			return nil, err
		}
		item.IsBound = isBound != 0
		items = append(items, item)
	}
	return items, rows.Err()
}

func (repo *Repository) ConfirmParentStudentsByPhone(ctx context.Context, phone string, studentIDs []int64) error {
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return errors.New("手机号不能为空")
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
		return errors.New("请选择至少一位学员")
	}

	placeholders := make([]string, 0, len(normalizedIDs))
	args := make([]any, 0, len(normalizedIDs)+1)
	args = append(args, phone)
	for _, studentID := range normalizedIDs {
		placeholders = append(placeholders, "?")
		args = append(args, studentID)
	}

	querySuffix := fmt.Sprintf(" AND id IN (%s)", strings.Join(placeholders, ","))

	var matchedCount int
	if err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM inst_student
		WHERE del_flag = 0 AND IFNULL(mobile, '') = ?`+querySuffix,
		args...,
	).Scan(&matchedCount); err != nil {
		return err
	}
	if matchedCount != len(normalizedIDs) {
		return errors.New("存在无效学员，或该学员不属于当前手机号")
	}

	updateArgs := make([]any, 0, len(args))
	updateArgs = append(updateArgs, args...)
	_, err := repo.db.ExecContext(ctx, `
		UPDATE inst_student
		SET is_bind_child = 1,
			update_time = NOW()
		WHERE del_flag = 0 AND IFNULL(mobile, '') = ?`+querySuffix,
		updateArgs...,
	)
	return err
}
