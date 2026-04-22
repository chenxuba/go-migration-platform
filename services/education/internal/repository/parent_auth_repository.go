package repository

import (
	"context"
	"database/sql"
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

type ParentStudentScheduleAliasRecord struct {
	StudentID     int64
	StudentStatus int
	AvatarURL     string
}

type ParentWeChatOfficialUserLinkRecord struct {
	ID             int64
	OfficialOpenID string
	MiniOpenID     string
	UnionID        string
	Phone          string
	Subscribed     bool
}

func (repo *Repository) ListParentStudentCandidatesByPhone(ctx context.Context, phone string) ([]ParentStudentLookupRecord, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT s.id, s.inst_id, IFNULL(s.stu_name, ''), IFNULL(s.avatar_url, ''), IFNULL(s.mobile, ''),
		       IFNULL(s.student_status, 0), IFNULL(s.phone_relationship, 0),
		       CASE WHEN bound.student_id IS NULL THEN 0 ELSE 1 END AS is_bound,
		       IFNULL(i.organ_name, ''), IFNULL(i.logo, '')
		FROM inst_student s
		LEFT JOIN org_institution i ON i.id = s.inst_id
		LEFT JOIN (
			SELECT DISTINCT inst_id, student_id
			FROM wechat_official_student_binding
			WHERE phone = ? AND subscribed = 1
		) bound ON bound.inst_id = s.inst_id AND bound.student_id = s.id
		WHERE s.del_flag = 0
		  AND (IFNULL(s.mobile, '') = ? OR bound.student_id IS NOT NULL)
		ORDER BY
			CASE WHEN bound.student_id IS NULL THEN 1 ELSE 0 END ASC,
			s.create_time DESC,
			s.id DESC
	`, strings.TrimSpace(phone), strings.TrimSpace(phone))
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

func (repo *Repository) GetSubscribedWeChatOfficialUserLinkByPhone(ctx context.Context, phone string) (ParentWeChatOfficialUserLinkRecord, error) {
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return ParentWeChatOfficialUserLinkRecord{}, sql.ErrNoRows
	}

	var item ParentWeChatOfficialUserLinkRecord
	err := repo.db.QueryRowContext(ctx, `
		SELECT
			IFNULL(official_openid, ''),
			IFNULL(mini_openid, ''),
			IFNULL(unionid, ''),
			IFNULL(phone, '')
		FROM wechat_official_user_link
		WHERE phone = ? AND subscribed = 1
		ORDER BY
			CASE WHEN IFNULL(official_openid, '') = '' THEN 1 ELSE 0 END ASC,
			update_time DESC,
			id DESC
		LIMIT 1
	`, phone).Scan(
		&item.OfficialOpenID,
		&item.MiniOpenID,
		&item.UnionID,
		&item.Phone,
	)
	if err != nil {
		return ParentWeChatOfficialUserLinkRecord{}, err
	}
	return item, nil
}

func (repo *Repository) GetWeChatOfficialUserLinkByMiniIdentity(ctx context.Context, miniOpenID, unionID string) (ParentWeChatOfficialUserLinkRecord, error) {
	miniOpenID = strings.TrimSpace(miniOpenID)
	unionID = strings.TrimSpace(unionID)

	if miniOpenID != "" {
		item, err := repo.getWeChatOfficialUserLinkRecord(ctx, `
			SELECT
				id,
				IFNULL(official_openid, ''),
				IFNULL(mini_openid, ''),
				IFNULL(unionid, ''),
				IFNULL(phone, ''),
				IFNULL(subscribed, 0)
			FROM wechat_official_user_link
			WHERE mini_openid = ?
			LIMIT 1
		`, miniOpenID)
		if err == nil {
			return item, nil
		}
		if err != sql.ErrNoRows {
			return ParentWeChatOfficialUserLinkRecord{}, err
		}
	}

	if unionID != "" {
		return repo.getWeChatOfficialUserLinkRecord(ctx, `
			SELECT
				id,
				IFNULL(official_openid, ''),
				IFNULL(mini_openid, ''),
				IFNULL(unionid, ''),
				IFNULL(phone, ''),
				IFNULL(subscribed, 0)
			FROM wechat_official_user_link
			WHERE unionid = ?
			LIMIT 1
		`, unionID)
	}

	return ParentWeChatOfficialUserLinkRecord{}, sql.ErrNoRows
}

func (repo *Repository) getWeChatOfficialUserLinkRecord(ctx context.Context, query string, args ...any) (ParentWeChatOfficialUserLinkRecord, error) {
	var item ParentWeChatOfficialUserLinkRecord
	var subscribed int
	err := repo.db.QueryRowContext(ctx, query, args...).Scan(
		&item.ID,
		&item.OfficialOpenID,
		&item.MiniOpenID,
		&item.UnionID,
		&item.Phone,
		&subscribed,
	)
	if err != nil {
		return ParentWeChatOfficialUserLinkRecord{}, err
	}
	item.Subscribed = subscribed != 0
	return item, nil
}

func (repo *Repository) ListParentStudentScheduleAliases(ctx context.Context, instID int64, studentName string, excludeStudentID int64) ([]ParentStudentScheduleAliasRecord, error) {
	studentName = strings.TrimSpace(studentName)
	if instID <= 0 || studentName == "" {
		return []ParentStudentScheduleAliasRecord{}, nil
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT
			s.id,
			IFNULL(s.student_status, 0),
			IFNULL(s.avatar_url, '')
		FROM inst_student s
		WHERE s.del_flag = 0
		  AND s.inst_id = ?
		  AND IFNULL(s.stu_name, '') = ?
		  AND s.id <> ?
		  AND IFNULL(s.student_status, 0) IN (1, 2)
		ORDER BY
			CASE IFNULL(s.student_status, 0)
				WHEN 1 THEN 0
				WHEN 2 THEN 1
				ELSE 9
			END ASC,
			s.update_time DESC,
			s.id DESC
	`, instID, studentName, excludeStudentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]ParentStudentScheduleAliasRecord, 0, 4)
	for rows.Next() {
		var item ParentStudentScheduleAliasRecord
		if err := rows.Scan(&item.StudentID, &item.StudentStatus, &item.AvatarURL); err != nil {
			return nil, err
		}
		if item.StudentID <= 0 {
			continue
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
