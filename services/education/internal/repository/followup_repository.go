package repository

import (
	"context"
	"database/sql"
	"strings"

	"go-migration-platform/services/education/internal/model"
)

func (repo *Repository) PageFollowUpRecords(ctx context.Context, instID int64, query model.StudentFollowUpQueryDTO) (model.PageResult[model.StudentFollowUpRecord], error) {
	current := query.PageRequestModel.PageIndex
	size := query.PageRequestModel.PageSize
	if current <= 0 {
		current = 1
	}
	if size <= 0 {
		size = 10
	}
	offset := (current - 1) * size

	filters := []string{"r.del_flag = 0", "r.inst_id = ?"}
	args := []any{instID}
	q := query.QueryModel
	if q.StudentID != nil {
		filters = append(filters, "r.student_id = ?")
		args = append(args, *q.StudentID)
	}
	if q.FollowUpStaffID != nil {
		filters = append(filters, "r.create_id = ?")
		args = append(args, *q.FollowUpStaffID)
	}
	if q.SalespersonID != nil {
		filters = append(filters, "s.sale_person = ?")
		args = append(args, *q.SalespersonID)
	}
	if strings.TrimSpace(q.SearchKey) != "" {
		filters = append(filters, "(s.stu_name LIKE ? OR s.mobile LIKE ? OR r.content LIKE ?)")
		kw := "%" + strings.TrimSpace(q.SearchKey) + "%"
		args = append(args, kw, kw, kw)
	}
	whereClause := strings.Join(filters, " AND ")

	var total int
	if err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM follow_record r
		LEFT JOIN inst_student s ON s.id = r.student_id
		WHERE `+whereClause, args...).Scan(&total); err != nil {
		return model.PageResult[model.StudentFollowUpRecord]{}, err
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT r.id, r.student_id, IFNULL(s.stu_name, ''), IFNULL(s.mobile, ''), IFNULL(s.student_status, 0),
		       s.sale_person, IFNULL(u.nick_name, ''), s.channel_id, IFNULL(c.channel_name, ''), c.category_id, IFNULL(cc.category_name, ''),
		       r.create_id, IFNULL(u5.nick_name, ''), r.create_time, IFNULL(r.content, ''), IFNULL(r.follow_images, ''),
		       r.follow_method, IFNULL(r.intended_course, ''), r.intention_level, r.follow_up_status, r.visit_status, r.follow_up_time, r.next_follow_up_time
		FROM follow_record r
		LEFT JOIN inst_student s ON s.id = r.student_id
		LEFT JOIN inst_channel c ON c.id = s.channel_id
		LEFT JOIN inst_channel_category cc ON cc.id = c.category_id
		LEFT JOIN inst_user u ON u.id = s.sale_person
		LEFT JOIN inst_user u5 ON u5.id = r.create_id
		WHERE `+whereClause+`
		ORDER BY r.create_time DESC
		LIMIT ? OFFSET ?`, append(args, size, offset)...)
	if err != nil {
		return model.PageResult[model.StudentFollowUpRecord]{}, err
	}
	defer rows.Close()

	items := make([]model.StudentFollowUpRecord, 0, size)
	for rows.Next() {
		var item model.StudentFollowUpRecord
		var intendedCourseRaw string
		var followUpTime sql.NullTime
		var nextFollowUpTime sql.NullTime
		if err := rows.Scan(
			&item.ID,
			&item.StudentID,
			&item.StuName,
			&item.Mobile,
			&item.StudentStatus,
			&item.SalesPersonID,
			&item.SalesPersonName,
			&item.ChannelID,
			&item.ChannelName,
			&item.CategoryID,
			&item.CategoryName,
			&item.CreateID,
			&item.CreateName,
			&item.CreateTime,
			&item.Content,
			&item.FollowImages,
			&item.FollowMethod,
			&intendedCourseRaw,
			&item.IntentionLevel,
			&item.FollowUpStatus,
			&item.VisitStatus,
			&followUpTime,
			&nextFollowUpTime,
		); err != nil {
			return model.PageResult[model.StudentFollowUpRecord]{}, err
		}
		item.IntendedCourse = parseCSVInt64(intendedCourseRaw)
		if followUpTime.Valid {
			t := followUpTime.Time
			item.FollowUpTime = &t
		}
		if nextFollowUpTime.Valid {
			t := nextFollowUpTime.Time
			item.NextFollowUpTime = &t
		}
		items = append(items, item)
	}
	return model.PageResult[model.StudentFollowUpRecord]{Items: items, Total: total, Current: current, Size: size}, rows.Err()
}

func (repo *Repository) CreateFollowUp(ctx context.Context, instID, instUserID int64, dto model.CreateFollowUpDTO) error {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	_, err = tx.ExecContext(ctx, `
		UPDATE inst_student
		SET follow_up_status = ?, intent_level = ?, last_follow_up_time = NOW(), next_follow_up_time = ?, intended_course = ?, update_id = ?
		WHERE id = ? AND inst_id = ? AND del_flag = 0
	`, dto.FollowUpStatus, dto.IntentLevel, dto.NextFollowUpTime, dto.IntentCourseIDs, instUserID, dto.StudentID, instID)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO follow_record
		(inst_id, student_id, follow_method, content, follow_images, follow_up_time, next_follow_up_time, intended_course, intention_level, follow_up_status, visit_status, create_id, create_time, del_flag, version)
		VALUES (?, ?, ?, ?, ?, NOW(), ?, ?, ?, ?, ?, ?, NOW(), 0, 0)
	`, instID, dto.StudentID, dto.FollowMethod, dto.Content, dto.FollowImages, dto.NextFollowUpTime, dto.IntentCourseIDs, dto.IntentLevel, dto.FollowUpStatus, false, instUserID)
	if err != nil {
		return err
	}

	err = tx.Commit()
	return err
}

func (repo *Repository) GetFollowUpCount(ctx context.Context, instID int64) (model.FollowUpCountVO, error) {
	row := repo.db.QueryRowContext(ctx, `
		SELECT
			COUNT(CASE WHEN DATE(r.next_follow_up_time) = CURDATE() THEN 1 END) AS to_be_followed_up_today_count,
			COUNT(CASE
			      WHEN s.create_time >= DATE_SUB(CURDATE(), INTERVAL WEEKDAY(CURDATE()) DAY)
			       AND s.create_time < DATE_ADD(DATE_SUB(CURDATE(), INTERVAL WEEKDAY(CURDATE()) DAY), INTERVAL 7 DAY)
			      THEN 1 END) AS new_inquiries_added_week_count,
			COUNT(CASE WHEN r.next_follow_up_time < NOW() AND r.visit_status = 0 THEN 1 END) AS overdue_for_follow_up_interview_count
		FROM inst_student s
		LEFT JOIN (
			SELECT DISTINCT student_id, next_follow_up_time, visit_status
			FROM follow_record
			WHERE del_flag = 0 AND inst_id = ?
		) r ON r.student_id = s.id
		WHERE s.inst_id = ? AND s.student_status = 0 AND s.del_flag = 0
	`, instID, instID)

	var result model.FollowUpCountVO
	err := row.Scan(&result.ToBeFollowedUpTodayCount, &result.NewInquiriesAddedWeekCount, &result.OverdueForFollowUpInterviewCount)
	return result, err
}

func (repo *Repository) UpdateVisitStatus(ctx context.Context, instID int64, dto model.VisitStatusUpdateDTO) error {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var studentID int64
	err = tx.QueryRowContext(ctx, `
		SELECT student_id
		FROM follow_record
		WHERE id = ? AND inst_id = ? AND del_flag = 0
		LIMIT 1
	`, dto.ID, instID).Scan(&studentID)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE follow_record
		SET visit_status = ?
		WHERE id = ? AND inst_id = ? AND del_flag = 0
	`, dto.VisitStatus, dto.ID, instID)
	if err != nil {
		return err
	}

	var latestFollowRecordID int64
	err = tx.QueryRowContext(ctx, `
		SELECT id
		FROM follow_record
		WHERE student_id = ? AND inst_id = ? AND del_flag = 0
		ORDER BY id DESC
		LIMIT 1
	`, studentID, instID).Scan(&latestFollowRecordID)
	if err != nil {
		return err
	}

	if latestFollowRecordID == dto.ID {
		followUpStatus := 0
		if dto.VisitStatus != nil && *dto.VisitStatus {
			followUpStatus = 1
		}
		_, err = tx.ExecContext(ctx, `
			UPDATE inst_student
			SET follow_up_status = ?
			WHERE id = ? AND inst_id = ? AND del_flag = 0
		`, followUpStatus, studentID, instID)
		if err != nil {
			return err
		}
	}

	err = tx.Commit()
	return err
}

func (repo *Repository) UpdateFollowUpRecord(ctx context.Context, instID int64, dto model.UpdateFollowUpDTO) error {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var studentID int64
	err = tx.QueryRowContext(ctx, `
		SELECT student_id
		FROM follow_record
		WHERE id = ? AND inst_id = ? AND del_flag = 0
		LIMIT 1
	`, dto.ID, instID).Scan(&studentID)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE follow_record
		SET follow_method = COALESCE(?, follow_method),
		    next_follow_up_time = ?,
		    content = ?,
		    follow_images = ?,
		    intended_course = ?,
		    intention_level = COALESCE(?, intention_level),
		    follow_up_status = COALESCE(?, follow_up_status),
		    update_time = NOW()
		WHERE id = ? AND inst_id = ? AND del_flag = 0
	`, dto.FollowMethod, dto.NextFollowUpTime, strings.TrimSpace(dto.Content), strings.TrimSpace(dto.FollowImages), dto.IntentCourseIDs, dto.IntentLevel, dto.FollowUpStatus, dto.ID, instID)
	if err != nil {
		return err
	}

	var latestID int64
	err = tx.QueryRowContext(ctx, `
		SELECT id
		FROM follow_record
		WHERE student_id = ? AND inst_id = ? AND del_flag = 0
		ORDER BY id DESC
		LIMIT 1
	`, studentID, instID).Scan(&latestID)
	if err != nil {
		return err
	}

	if latestID == dto.ID {
		_, err = tx.ExecContext(ctx, `
			UPDATE inst_student
			SET follow_up_status = COALESCE(?, follow_up_status),
			    intent_level = COALESCE(?, intent_level),
			    last_follow_up_time = NOW(),
			    next_follow_up_time = ?,
			    intended_course = ?
			WHERE id = ? AND inst_id = ? AND del_flag = 0
		`, dto.FollowUpStatus, dto.IntentLevel, dto.NextFollowUpTime, dto.IntentCourseIDs, studentID, instID)
		if err != nil {
			return err
		}
	}

	err = tx.Commit()
	return err
}

func (repo *Repository) GetFollowUpRecordStatistics(ctx context.Context, instID int64, query model.StudentFollowUpQueryDTO) (model.FollowVisitCountVO, error) {
	filters := []string{"r.del_flag = 0", "r.inst_id = ?"}
	args := []any{instID}
	q := query.QueryModel
	if q.StudentID != nil {
		filters = append(filters, "s.id = ?")
		args = append(args, *q.StudentID)
	}
	if q.FollowUpStaffID != nil {
		filters = append(filters, "r.create_id = ?")
		args = append(args, *q.FollowUpStaffID)
	}
	if q.SalespersonID != nil {
		filters = append(filters, "s.sale_person = ?")
		args = append(args, *q.SalespersonID)
	}
	if strings.TrimSpace(q.SearchKey) != "" {
		filters = append(filters, "(s.stu_name LIKE ? OR s.mobile LIKE ? OR r.content LIKE ?)")
		kw := "%" + strings.TrimSpace(q.SearchKey) + "%"
		args = append(args, kw, kw, kw)
	}
	whereClause := strings.Join(filters, " AND ")

	row := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT r.student_id) AS student_count,
		       COUNT(CASE WHEN r.visit_status = 1 THEN 1 END) AS interview_count,
		       COUNT(CASE WHEN r.visit_status = 0 THEN 1 END) AS not_interview_count
		FROM follow_record r
		LEFT JOIN inst_student s ON s.id = r.student_id
		WHERE `+whereClause, args...)

	var result model.FollowVisitCountVO
	err := row.Scan(&result.StudentCount, &result.InterviewCount, &result.NotInterviewCount)
	return result, err
}

func (repo *Repository) GetFollowRecordStudentID(ctx context.Context, instID, followRecordID int64) (int64, error) {
	var studentID int64
	err := repo.db.QueryRowContext(ctx, `
		SELECT student_id
		FROM follow_record
		WHERE id = ? AND inst_id = ? AND del_flag = 0
		LIMIT 1
	`, followRecordID, instID).Scan(&studentID)
	return studentID, err
}
