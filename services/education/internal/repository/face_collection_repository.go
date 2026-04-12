package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"

	"go-migration-platform/services/education/internal/model"
)

func ensureFaceCollectionTables(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS inst_student_face_profile (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			inst_id BIGINT NOT NULL,
			student_id BIGINT NOT NULL,
			face_descriptor LONGTEXT NOT NULL,
			face_image LONGTEXT NULL,
			create_id BIGINT NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_id BIGINT NOT NULL DEFAULT 0,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			UNIQUE KEY uk_inst_student_face_profile (inst_id, student_id),
			KEY idx_inst_student_face_profile_student (student_id),
			KEY idx_inst_student_face_profile_inst (inst_id)
		)
	`)
	return err
}

func (repo *Repository) PageFaceCollectionStudents(ctx context.Context, instID int64, query model.FaceCollectionStudentQueryDTO) (model.PageResult[model.FaceCollectionStudent], error) {
	current := query.PageRequestModel.PageIndex
	size := query.PageRequestModel.PageSize
	if current <= 0 {
		current = 1
	}
	if size <= 0 {
		size = 20
	}
	offset := (current - 1) * size

	filters := []string{
		"s.del_flag = 0",
		"s.inst_id = ?",
		"s.student_status IN (1, 2)",
	}
	args := []any{instID}
	if searchKey := strings.TrimSpace(query.QueryModel.SearchKey); searchKey != "" {
		filters = append(filters, "(s.stu_name LIKE ? OR s.mobile LIKE ?)")
		args = append(args, "%"+searchKey+"%", "%"+searchKey+"%")
	}
	whereClause := strings.Join(filters, " AND ")

	var total int
	if err := repo.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM inst_student s WHERE `+whereClause, args...).Scan(&total); err != nil {
		return model.PageResult[model.FaceCollectionStudent]{}, err
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT s.id, IFNULL(s.stu_name, ''), IFNULL(s.avatar_url, ''), IFNULL(s.mobile, ''), IFNULL(s.is_collect, 0)
		FROM inst_student s
		WHERE `+whereClause+`
		ORDER BY IFNULL(s.is_collect, 0) DESC, IFNULL(s.stu_name, '') ASC, s.id DESC
		LIMIT ? OFFSET ?
	`, append(args, size, offset)...)
	if err != nil {
		return model.PageResult[model.FaceCollectionStudent]{}, err
	}
	defer rows.Close()

	items := make([]model.FaceCollectionStudent, 0, size)
	for rows.Next() {
		var item model.FaceCollectionStudent
		if err := rows.Scan(&item.ID, &item.StuName, &item.AvatarURL, &item.Mobile, &item.IsCollect); err != nil {
			return model.PageResult[model.FaceCollectionStudent]{}, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return model.PageResult[model.FaceCollectionStudent]{}, err
	}
	return model.PageResult[model.FaceCollectionStudent]{
		Items:   items,
		Total:   total,
		Current: current,
		Size:    size,
	}, nil
}

func (repo *Repository) GetFaceCollectionProfile(ctx context.Context, instID, studentID int64) (model.FaceCollectionProfile, error) {
	var (
		item           model.FaceCollectionProfile
		descriptorJSON string
		updateTime     sql.NullTime
	)
	err := repo.db.QueryRowContext(ctx, `
		SELECT p.student_id, IFNULL(s.stu_name, ''), IFNULL(p.face_descriptor, ''), IFNULL(p.face_image, ''), p.update_time
		FROM inst_student_face_profile p
		LEFT JOIN inst_student s ON s.id = p.student_id
		WHERE p.inst_id = ? AND p.student_id = ? AND p.del_flag = 0
		LIMIT 1
	`, instID, studentID).Scan(&item.StudentID, &item.StuName, &descriptorJSON, &item.FaceImage, &updateTime)
	if err != nil {
		return model.FaceCollectionProfile{}, err
	}
	if strings.TrimSpace(descriptorJSON) != "" {
		if err := json.Unmarshal([]byte(descriptorJSON), &item.FaceDescriptor); err != nil {
			return model.FaceCollectionProfile{}, err
		}
	}
	if updateTime.Valid {
		t := updateTime.Time
		item.UpdatedTime = &t
	}
	return item, nil
}

func (repo *Repository) ListFaceCollectionProfiles(ctx context.Context, instID int64) ([]model.FaceCollectionProfile, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT p.student_id, IFNULL(s.stu_name, ''), IFNULL(p.face_descriptor, ''), p.update_time
		FROM inst_student_face_profile p
		LEFT JOIN inst_student s ON s.id = p.student_id
		WHERE p.inst_id = ? AND p.del_flag = 0
		ORDER BY p.update_time DESC, p.id DESC
	`, instID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.FaceCollectionProfile, 0)
	for rows.Next() {
		var (
			item           model.FaceCollectionProfile
			descriptorJSON string
			updateTime     sql.NullTime
		)
		if err := rows.Scan(&item.StudentID, &item.StuName, &descriptorJSON, &updateTime); err != nil {
			return nil, err
		}
		if strings.TrimSpace(descriptorJSON) != "" {
			if err := json.Unmarshal([]byte(descriptorJSON), &item.FaceDescriptor); err != nil {
				return nil, err
			}
		}
		if updateTime.Valid {
			t := updateTime.Time
			item.UpdatedTime = &t
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (repo *Repository) SaveFaceCollectionProfile(ctx context.Context, instID, operatorID int64, dto model.FaceCollectionProfileSaveDTO) error {
	if dto.StudentID <= 0 {
		return errors.New("invalid studentId")
	}
	if len(dto.FaceDescriptor) == 0 {
		return errors.New("faceDescriptor 不能为空")
	}
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var exists int
	if err := tx.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM inst_student
		WHERE id = ? AND inst_id = ? AND del_flag = 0 AND student_status IN (1, 2)
	`, dto.StudentID, instID).Scan(&exists); err != nil {
		return err
	}
	if exists == 0 {
		return errors.New("学员不存在或当前状态不支持人脸采集")
	}

	descriptorJSON, err := json.Marshal(dto.FaceDescriptor)
	if err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO inst_student_face_profile (
			inst_id, student_id, face_descriptor, face_image, create_id, update_id, del_flag
		) VALUES (?, ?, ?, ?, ?, ?, 0)
		ON DUPLICATE KEY UPDATE
			face_descriptor = VALUES(face_descriptor),
			face_image = VALUES(face_image),
			update_id = VALUES(update_id),
			update_time = CURRENT_TIMESTAMP,
			del_flag = 0
	`, instID, dto.StudentID, string(descriptorJSON), dto.FaceImage, operatorID, operatorID); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE inst_student
		SET is_collect = 1
		WHERE id = ? AND inst_id = ? AND del_flag = 0
	`, dto.StudentID, instID); err != nil {
		return err
	}

	return tx.Commit()
}

func (repo *Repository) DeleteFaceCollectionProfile(ctx context.Context, instID, operatorID int64, studentID int64) error {
	if studentID <= 0 {
		return errors.New("invalid studentId")
	}
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `
		UPDATE inst_student_face_profile
		SET del_flag = 1, update_id = ?, update_time = CURRENT_TIMESTAMP
		WHERE inst_id = ? AND student_id = ? AND del_flag = 0
	`, operatorID, instID, studentID); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE inst_student
		SET is_collect = 0
		WHERE id = ? AND inst_id = ? AND del_flag = 0
	`, studentID, instID); err != nil {
		return err
	}

	return tx.Commit()
}
