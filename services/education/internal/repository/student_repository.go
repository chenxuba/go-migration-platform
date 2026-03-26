package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"go-migration-platform/services/education/internal/model"
)

func (repo *Repository) GetStudentSnapshot(ctx context.Context, instID, studentID int64) (StudentSnapshot, error) {
	row := repo.db.QueryRowContext(ctx, `
		SELECT id, inst_id, IFNULL(stu_name, ''), IFNULL(mobile, ''), phone_relationship, sale_person, channel_id,
		       collector_staff_id, phone_sell_staff_id, foreground_staff_id, vice_sell_staff_Id,
		       student_manager_id, advisor_id, recommend_student_id, IFNULL(wechat_number, ''), IFNULL(grade, ''),
		       IFNULL(study_school, ''), IFNULL(interest, ''), IFNULL(address, ''), IFNULL(remark, ''),
		       follow_up_status, intent_level
		FROM inst_student
		WHERE id = ? AND inst_id = ? AND del_flag = 0
		LIMIT 1
	`, studentID, instID)
	var item StudentSnapshot
	err := row.Scan(
		&item.ID,
		&item.InstID,
		&item.StuName,
		&item.Mobile,
		&item.PhoneRelationship,
		&item.SalePerson,
		&item.ChannelID,
		&item.CollectorStaffID,
		&item.PhoneSellStaffID,
		&item.ForegroundStaffID,
		&item.ViceSellStaffID,
		&item.StudentManagerID,
		&item.AdvisorID,
		&item.RecommendStudentID,
		&item.WeChatNumber,
		&item.Grade,
		&item.StudySchool,
		&item.Interest,
		&item.Address,
		&item.Remark,
		&item.FollowUpStatus,
		&item.IntentLevel,
	)
	return item, err
}

func (repo *Repository) GetStudentNameByID(ctx context.Context, studentID *int64) string {
	if studentID == nil {
		return "-"
	}
	var name string
	err := repo.db.QueryRowContext(ctx, "SELECT IFNULL(stu_name, '') FROM inst_student WHERE id = ? LIMIT 1", *studentID).Scan(&name)
	if err != nil || strings.TrimSpace(name) == "" {
		return fmt.Sprintf("未知学员(%d)", *studentID)
	}
	return name
}

func (repo *Repository) InsertStudentChangeRecord(ctx context.Context, instID, stuID, changeID int64, content string) error {
	if strings.TrimSpace(content) == "" {
		return nil
	}
	_, err := repo.db.ExecContext(ctx, `
		INSERT INTO inst_student_record (inst_id, stu_id, change_content, change_id, create_time, del_flag)
		VALUES (?, ?, ?, ?, NOW(), 0)
	`, instID, stuID, strings.TrimSpace(content), changeID)
	return err
}

func (repo *Repository) UpdateStudentStatus(ctx context.Context, instID int64, dto model.StudentStatusUpdateDTO) error {
	_, err := repo.db.ExecContext(ctx, `
		UPDATE inst_student
		SET follow_up_status = COALESCE(?, follow_up_status),
		    intent_level = COALESCE(?, intent_level)
		WHERE id = ? AND inst_id = ? AND del_flag = 0
	`, dto.FollowUpStatus, dto.IntentLevel, dto.ID, instID)
	return err
}

func (repo *Repository) BatchAssignSalesperson(ctx context.Context, instID int64, salespersonID int64, studentIDs []int64) error {
	if len(studentIDs) == 0 {
		return nil
	}
	placeholders := make([]string, 0, len(studentIDs))
	args := make([]any, 0, len(studentIDs)+2)
	args = append(args, salespersonID)
	for _, id := range studentIDs {
		placeholders = append(placeholders, "?")
		args = append(args, id)
	}
	args = append(args, instID)

	query := `
		UPDATE inst_student
		SET sale_person = ?, sale_assigned_time = NOW()
		WHERE id IN (` + strings.Join(placeholders, ",") + `)
		  AND inst_id = ?
		  AND del_flag = 0`

	_, err := repo.db.ExecContext(ctx, query, args...)
	return err
}

func (repo *Repository) BatchTransferToPublicPool(ctx context.Context, instID int64, studentIDs []int64) error {
	if len(studentIDs) == 0 {
		return nil
	}
	placeholders := make([]string, 0, len(studentIDs))
	args := make([]any, 0, len(studentIDs)+1)
	for _, id := range studentIDs {
		placeholders = append(placeholders, "?")
		args = append(args, id)
	}
	args = append(args, instID)

	query := `
		UPDATE inst_student
		SET sale_person = NULL,
		    sale_assigned_time = NULL
		WHERE id IN (` + strings.Join(placeholders, ",") + `)
		  AND inst_id = ?
		  AND del_flag = 0`

	_, err := repo.db.ExecContext(ctx, query, args...)
	return err
}

func (repo *Repository) BatchDeleteIntentStudents(ctx context.Context, instID int64, studentIDs []int64) error {
	if len(studentIDs) == 0 {
		return nil
	}
	placeholders := make([]string, 0, len(studentIDs))
	args := make([]any, 0, len(studentIDs)+1)
	for _, id := range studentIDs {
		placeholders = append(placeholders, "?")
		args = append(args, id)
	}
	args = append(args, instID)

	query := `
		UPDATE inst_student
		SET del_flag = 1
		WHERE id IN (` + strings.Join(placeholders, ",") + `)
		  AND inst_id = ?
		  AND del_flag = 0`

	_, err := repo.db.ExecContext(ctx, query, args...)
	return err
}

func (repo *Repository) GetAddIntentionStudentRule(ctx context.Context, instID int64) (int, error) {
	var rule sql.NullInt64
	err := repo.db.QueryRowContext(ctx, `
		SELECT add_intention_student_rule
		FROM inst_config
		WHERE inst_id = ? AND del_flag = 0
		ORDER BY id DESC
		LIMIT 1
	`, instID).Scan(&rule)
	if err != nil {
		if err == sql.ErrNoRows {
			return 1, nil
		}
		return 0, err
	}
	if !rule.Valid || rule.Int64 < 1 || rule.Int64 > 3 {
		return 1, nil
	}
	return int(rule.Int64), nil
}

func (repo *Repository) GetAddImportStudentRule(ctx context.Context, instID int64) (int, error) {
	var rule sql.NullInt64
	err := repo.db.QueryRowContext(ctx, `
		SELECT add_import_student_rule
		FROM inst_config
		WHERE inst_id = ? AND del_flag = 0
		ORDER BY id DESC
		LIMIT 1
	`, instID).Scan(&rule)
	if err != nil {
		if err == sql.ErrNoRows {
			return 1, nil
		}
		return 0, err
	}
	if !rule.Valid || rule.Int64 < 1 || rule.Int64 > 3 {
		return 1, nil
	}
	return int(rule.Int64), nil
}

func (repo *Repository) GetLimitSameWeChat(ctx context.Context, instID int64) (bool, error) {
	var value sql.NullInt64
	err := repo.db.QueryRowContext(ctx, `
		SELECT IFNULL(limit_same_weChat, 0)
		FROM inst_config
		WHERE inst_id = ? AND del_flag = 0
		ORDER BY id DESC
		LIMIT 1
	`, instID).Scan(&value)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return value.Valid && value.Int64 != 0, nil
}

func (repo *Repository) GetLimitImportSameWeChat(ctx context.Context, instID int64) (bool, error) {
	var value sql.NullInt64
	err := repo.db.QueryRowContext(ctx, `
		SELECT IFNULL(limit_import_same_weChat, 0)
		FROM inst_config
		WHERE inst_id = ? AND del_flag = 0
		ORDER BY id DESC
		LIMIT 1
	`, instID).Scan(&value)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return value.Valid && value.Int64 != 0, nil
}

func (repo *Repository) CountStudentDuplicatesByRule(ctx context.Context, instID, rule int64, stuName, mobile string, excludeID *int64) (int, error) {
	filters := []string{"inst_id = ?", "del_flag = 0"}
	args := []any{instID}
	switch rule {
	case 1:
		filters = append(filters, "stu_name = ?", "mobile = ?")
		args = append(args, strings.TrimSpace(stuName), strings.TrimSpace(mobile))
	case 2:
		filters = append(filters, "mobile = ?")
		args = append(args, strings.TrimSpace(mobile))
	case 3:
		filters = append(filters, "stu_name = ?")
		args = append(args, strings.TrimSpace(stuName))
	default:
		filters = append(filters, "stu_name = ?", "mobile = ?")
		args = append(args, strings.TrimSpace(stuName), strings.TrimSpace(mobile))
	}
	if excludeID != nil {
		filters = append(filters, "id <> ?")
		args = append(args, *excludeID)
	}

	query := "SELECT COUNT(*) FROM inst_student WHERE " + strings.Join(filters, " AND ")
	var count int
	err := repo.db.QueryRowContext(ctx, query, args...).Scan(&count)
	return count, err
}

func (repo *Repository) FindStudentIDByNameMobile(ctx context.Context, instID int64, stuName, mobile string) (int64, error) {
	var studentID int64
	err := repo.db.QueryRowContext(ctx, `
		SELECT id
		FROM inst_student
		WHERE inst_id = ? AND del_flag = 0 AND stu_name = ? AND mobile = ?
		ORDER BY id ASC
		LIMIT 1
	`, instID, strings.TrimSpace(stuName), strings.TrimSpace(mobile)).Scan(&studentID)
	return studentID, err
}

func (repo *Repository) CountStudentByWeChat(ctx context.Context, instID int64, weChat string, excludeID *int64) (int, error) {
	filters := []string{"inst_id = ?", "del_flag = 0", "wechat_number = ?"}
	args := []any{instID, strings.TrimSpace(weChat)}
	if excludeID != nil {
		filters = append(filters, "id <> ?")
		args = append(args, *excludeID)
	}
	var count int
	err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM inst_student
		WHERE `+strings.Join(filters, " AND "),
		args...,
	).Scan(&count)
	return count, err
}

func (repo *Repository) UpdateStudentStatusValue(ctx context.Context, instID, studentID int64, status int) error {
	_, err := repo.db.ExecContext(ctx, `
		UPDATE inst_student
		SET student_status = ?, update_time = NOW()
		WHERE id = ? AND inst_id = ? AND del_flag = 0
	`, status, studentID, instID)
	return err
}

func (repo *Repository) GetStudentPhone(ctx context.Context, instID, studentID int64) (string, error) {
	var mobile string
	err := repo.db.QueryRowContext(ctx, `
		SELECT IFNULL(mobile, '')
		FROM inst_student
		WHERE id = ? AND inst_id = ? AND del_flag = 0
		LIMIT 1
	`, studentID, instID).Scan(&mobile)
	return mobile, err
}

func (repo *Repository) PageRecommenders(ctx context.Context, instID int64, query model.RecommenderQueryDTO) (model.PageResult[model.RecommenderQueryVO], error) {
	current := query.PageRequestModel.PageIndex
	size := query.PageRequestModel.PageSize
	if current <= 0 {
		current = 1
	}
	if size <= 0 {
		size = 10
	}
	offset := (current - 1) * size

	filters := []string{"inst_id = ?", "del_flag = 0"}
	args := []any{instID}
	if query.QueryModel.StudentID != nil {
		filters = append(filters, "id = ?")
		args = append(args, *query.QueryModel.StudentID)
	}
	if query.QueryModel.StudentStatus != nil {
		filters = append(filters, "student_status = ?")
		args = append(args, *query.QueryModel.StudentStatus)
	}
	if strings.TrimSpace(query.QueryModel.SearchKey) != "" {
		filters = append(filters, "(stu_name LIKE ? OR mobile LIKE ?)")
		kw := "%" + strings.TrimSpace(query.QueryModel.SearchKey) + "%"
		args = append(args, kw, kw)
	}
	whereClause := strings.Join(filters, " AND ")

	var total int
	if err := repo.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM inst_student WHERE "+whereClause, args...).Scan(&total); err != nil {
		return model.PageResult[model.RecommenderQueryVO]{}, err
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT id, IFNULL(stu_name, ''), IFNULL(avatar_url, ''), IFNULL(mobile, ''), IFNULL(student_status, 0)
		FROM inst_student
		WHERE `+whereClause+`
		ORDER BY create_time DESC
		LIMIT ? OFFSET ?`, append(args, size, offset)...)
	if err != nil {
		return model.PageResult[model.RecommenderQueryVO]{}, err
	}
	defer rows.Close()

	items := make([]model.RecommenderQueryVO, 0, size)
	for rows.Next() {
		var item model.RecommenderQueryVO
		if err := rows.Scan(&item.ID, &item.StuName, &item.AvatarURL, &item.Mobile, &item.StudentStatus); err != nil {
			return model.PageResult[model.RecommenderQueryVO]{}, err
		}
		item.Mobile = maskPhoneLocal(item.Mobile)
		items = append(items, item)
	}
	return model.PageResult[model.RecommenderQueryVO]{Items: items, Total: total, Current: current, Size: size}, rows.Err()
}

func (repo *Repository) PageBirthdayStudents(ctx context.Context, instID int64, query model.BirthdayStudentQueryDTO) (model.PageResult[model.BirthdayStudentQueryVO], error) {
	current := query.PageRequestModel.PageIndex
	size := query.PageRequestModel.PageSize
	if current <= 0 {
		current = 1
	}
	if size <= 0 {
		size = 10
	}
	offset := (current - 1) * size

	filters := []string{"s.inst_id = ?", "s.del_flag = 0"}
	args := []any{instID}
	if query.QueryModel.StudentManagerID != nil {
		filters = append(filters, "s.student_manager_id = ?")
		args = append(args, *query.QueryModel.StudentManagerID)
	}
	if query.QueryModel.AdvisorID != nil {
		filters = append(filters, "s.advisor_id = ?")
		args = append(args, *query.QueryModel.AdvisorID)
	}
	if query.QueryModel.BirthMonth != nil {
		filters = append(filters, "MONTH(s.birthday) = ?")
		args = append(args, *query.QueryModel.BirthMonth)
	}
	whereClause := strings.Join(filters, " AND ")

	var total int
	if err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM inst_student s
		WHERE `+whereClause, args...).Scan(&total); err != nil {
		return model.PageResult[model.BirthdayStudentQueryVO]{}, err
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT s.id, IFNULL(s.stu_name, ''), IFNULL(s.avatar_url, ''), s.stu_sex, IFNULL(s.mobile, ''), s.phone_relationship,
		       IFNULL(s.student_status, 0), s.birthday, s.student_manager_id, IFNULL(u1.nick_name, ''), s.advisor_id, IFNULL(u2.nick_name, '')
		FROM inst_student s
		LEFT JOIN inst_user u1 ON u1.id = s.student_manager_id
		LEFT JOIN inst_user u2 ON u2.id = s.advisor_id
		WHERE `+whereClause+`
		ORDER BY s.birthday ASC, s.id DESC
		LIMIT ? OFFSET ?`, append(args, size, offset)...)
	if err != nil {
		return model.PageResult[model.BirthdayStudentQueryVO]{}, err
	}
	defer rows.Close()

	items := make([]model.BirthdayStudentQueryVO, 0, size)
	for rows.Next() {
		var item model.BirthdayStudentQueryVO
		var birthday sql.NullTime
		if err := rows.Scan(&item.ID, &item.StuName, &item.AvatarURL, &item.StuSex, &item.Mobile, &item.PhoneRelationship, &item.StudentStatus, &birthday, &item.StudentManagerID, &item.StudentManagerName, &item.AdvisorID, &item.AdvisorName); err != nil {
			return model.PageResult[model.BirthdayStudentQueryVO]{}, err
		}
		if birthday.Valid {
			t := birthday.Time
			item.BirthDay = &t
		}
		items = append(items, item)
	}
	return model.PageResult[model.BirthdayStudentQueryVO]{Items: items, Total: total, Current: current, Size: size}, rows.Err()
}

func (repo *Repository) ListStudentChangeRecords(ctx context.Context, instID, stuID int64) ([]model.StudentChangeRecord, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT r.id, r.stu_id, IFNULL(r.change_content, ''), r.change_id, IFNULL(u.nick_name, ''), r.create_time, IFNULL(r.remark, '')
		FROM inst_student_record r
		LEFT JOIN inst_user u ON r.change_id = u.id
		WHERE r.stu_id = ? AND r.inst_id = ? AND r.del_flag = 0
		ORDER BY r.create_time DESC
	`, stuID, instID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.StudentChangeRecord, 0, 16)
	for rows.Next() {
		var item model.StudentChangeRecord
		if err := rows.Scan(&item.ID, &item.StuID, &item.ChangeContent, &item.ChangeID, &item.ChangeName, &item.CreateTime, &item.Remark); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (repo *Repository) CreateIntentStudent(ctx context.Context, instID, operatorID int64, dto model.StudentSaveDTO) (int64, error) {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx, `
		INSERT INTO inst_student
		(inst_id, stu_name, stu_sex, birthday, mobile, phone_relationship, avatar_url, channel_id, sale_person,
		 sale_assigned_time, follow_up_status, intent_level, student_status, wechat_number, grade, study_school,
		 interest, address, recommend_student_id, collector_staff_id, phone_sell_staff_id, foreground_staff_id,
		 vice_sell_staff_id, student_manager_id, advisor_id, remark, del_flag, create_id, create_time, update_id, update_time)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?,
		        CASE WHEN ? IS NULL THEN NULL ELSE NOW() END, ?, ?, 0, ?, ?, ?,
		        ?, ?, ?, ?, ?, ?,
		        ?, ?, ?, ?, 0, ?, NOW(), ?, NOW())
	`,
		instID,
		strings.TrimSpace(dto.StuName),
		dto.Sex,
		dto.Birthday,
		strings.TrimSpace(dto.Mobile),
		dto.PhoneRelationship,
		strings.TrimSpace(dto.Avatar),
		dto.ChannelID,
		dto.SalespersonID,
		dto.SalespersonID,
		0,
		1,
		strings.TrimSpace(dto.WeChatNumber),
		strings.TrimSpace(dto.Grade),
		strings.TrimSpace(dto.StudySchool),
		strings.TrimSpace(dto.Interest),
		strings.TrimSpace(dto.Address),
		dto.RecommendStudentID,
		dto.CollectorStaffID,
		dto.PhoneSellStaffID,
		dto.ForegroundStaffID,
		dto.ViceSellStaffID,
		dto.StudentManagerID,
		dto.AdvisorID,
		strings.TrimSpace(dto.Remark),
		operatorID,
		operatorID,
	)
	if err != nil {
		return 0, err
	}
	studentID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	if err := repo.replaceStudentCustomFieldValuesTx(ctx, tx, studentID, operatorID, dto.CustomInfo); err != nil {
		return 0, err
	}
	if err := repo.ensureRechargeAccountTx(ctx, tx, instID, studentID, operatorID); err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return studentID, nil
}

func (repo *Repository) UpdateIntentStudent(ctx context.Context, instID int64, dto model.StudentSaveDTO) error {
	if dto.StudentID == nil {
		return fmt.Errorf("studentId is required")
	}
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
		UPDATE inst_student
		SET stu_name = ?, stu_sex = ?, birthday = ?, mobile = ?, phone_relationship = ?, avatar_url = ?, channel_id = ?,
		    sale_person = ?, wechat_number = ?, grade = ?, study_school = ?, interest = ?, address = ?,
		    recommend_student_id = ?, collector_staff_id = ?, phone_sell_staff_id = ?, foreground_staff_id = ?,
		    vice_sell_staff_id = ?, student_manager_id = ?, advisor_id = ?, remark = ?, update_id = ?, update_time = NOW()
		WHERE id = ? AND inst_id = ? AND del_flag = 0
	`,
		strings.TrimSpace(dto.StuName),
		dto.Sex,
		dto.Birthday,
		strings.TrimSpace(dto.Mobile),
		dto.PhoneRelationship,
		strings.TrimSpace(dto.Avatar),
		dto.ChannelID,
		dto.SalespersonID,
		strings.TrimSpace(dto.WeChatNumber),
		strings.TrimSpace(dto.Grade),
		strings.TrimSpace(dto.StudySchool),
		strings.TrimSpace(dto.Interest),
		strings.TrimSpace(dto.Address),
		dto.RecommendStudentID,
		dto.CollectorStaffID,
		dto.PhoneSellStaffID,
		dto.ForegroundStaffID,
		dto.ViceSellStaffID,
		dto.StudentManagerID,
		dto.AdvisorID,
		strings.TrimSpace(dto.Remark),
		dto.OperatorID,
		*dto.StudentID,
		instID,
	)
	if err != nil {
		return err
	}
	if err := repo.replaceStudentCustomFieldValuesTx(ctx, tx, *dto.StudentID, derefInt64ForCustom(dto.OperatorID), dto.CustomInfo); err != nil {
		return err
	}
	return tx.Commit()
}

func (repo *Repository) replaceStudentCustomFieldValuesTx(ctx context.Context, tx *sql.Tx, studentID, operatorID int64, values []model.CustomInfo) error {
	if _, err := tx.ExecContext(ctx, `
		UPDATE inst_student_field_value
		SET del_flag = 1, update_id = ?, update_time = NOW()
		WHERE student_id = ? AND del_flag = 0
	`, operatorID, studentID); err != nil {
		return err
	}
	for _, item := range values {
		if item.FieldID <= 0 || strings.TrimSpace(item.Value) == "" {
			continue
		}
		fieldKey := strings.TrimSpace(item.FieldName)
		if fieldKey == "" {
			_ = tx.QueryRowContext(ctx, `
				SELECT IFNULL(field_key, '')
				FROM inst_student_field_key
				WHERE id = ? AND del_flag = 0
				LIMIT 1
			`, item.FieldID).Scan(&fieldKey)
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO inst_student_field_value (
				uuid, version, student_id, field_id, field_key, field_value,
				create_id, create_time, update_id, update_time, del_flag
			) VALUES (
				UUID(), 0, ?, ?, ?, ?, ?, NOW(), ?, NOW(), 0
			)
		`, studentID, item.FieldID, fieldKey, strings.TrimSpace(item.Value), operatorID, operatorID); err != nil {
			return err
		}
	}
	return nil
}

func derefInt64ForCustom(value *int64) int64 {
	if value == nil {
		return 0
	}
	return *value
}

func maskPhoneLocal(value string) string {
	if len(value) == 11 {
		return value[:3] + "****" + value[7:]
	}
	return value
}
