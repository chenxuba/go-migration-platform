package repository

import (
	"context"
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
}

func (repo *Repository) ListParentStudentCandidatesByPhone(ctx context.Context, phone string) ([]ParentStudentLookupRecord, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT s.id, s.inst_id, IFNULL(s.stu_name, ''), IFNULL(s.avatar_url, ''), IFNULL(s.mobile, ''),
		       IFNULL(s.student_status, 0), IFNULL(s.phone_relationship, 0), IFNULL(s.is_bind_child, 0),
		       IFNULL(i.organ_name, '')
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
		); err != nil {
			return nil, err
		}
		item.IsBound = isBound != 0
		items = append(items, item)
	}
	return items, rows.Err()
}
