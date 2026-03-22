package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

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
	if q.QuickFilter != nil {
		switch *q.QuickFilter {
		case 1:
			startOfDay := time.Now().Truncate(24 * time.Hour)
			endOfDay := startOfDay.Add(24 * time.Hour)
			filters = append(filters, "r.next_follow_up_time >= ?", "r.next_follow_up_time < ?")
			args = append(args, startOfDay, endOfDay)
		case 2:
			filters = append(filters, "r.next_follow_up_time < NOW()", "IFNULL(r.visit_status, 0) = 0")
		}
	}
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
	if len(q.Sexes) > 0 {
		holders := make([]string, 0, len(q.Sexes))
		for _, item := range q.Sexes {
			holders = append(holders, "?")
			args = append(args, item)
		}
		filters = append(filters, "s.stu_sex IN ("+strings.Join(holders, ",")+")")
	}
	if begin := parseDateStart(q.FollowUpTimeBegin); begin != nil {
		filters = append(filters, "r.create_time >= ?")
		args = append(args, *begin)
	}
	if end := parseDateEnd(q.FollowUpTimeEnd); end != nil {
		filters = append(filters, "r.create_time <= ?")
		args = append(args, *end)
	}
	if begin := parseDateStart(q.NextFollowUpTimeBegin); begin != nil {
		filters = append(filters, "r.next_follow_up_time >= ?")
		args = append(args, *begin)
	}
	if end := parseDateEnd(q.NextFollowUpTimeEnd); end != nil {
		filters = append(filters, "r.next_follow_up_time <= ?")
		args = append(args, *end)
	}
	if len(q.FollowUpTypes) > 0 {
		holders := make([]string, 0, len(q.FollowUpTypes))
		for _, item := range q.FollowUpTypes {
			holders = append(holders, "?")
			args = append(args, item)
		}
		filters = append(filters, "r.follow_method IN ("+strings.Join(holders, ",")+")")
	}
	if len(q.VisitStatuses) > 0 {
		holders := make([]string, 0, len(q.VisitStatuses))
		for _, item := range q.VisitStatuses {
			holders = append(holders, "?")
			args = append(args, item)
		}
		filters = append(filters, "r.visit_status IN ("+strings.Join(holders, ",")+")", "r.next_follow_up_time IS NOT NULL")
	}
	if len(q.ChannelIDs) > 0 {
		holders := make([]string, 0, len(q.ChannelIDs))
		for _, item := range q.ChannelIDs {
			holders = append(holders, "?")
			args = append(args, item)
		}
		filters = append(filters, "s.channel_id IN ("+strings.Join(holders, ",")+")")
	}
	if len(q.StudentStatuses) > 0 {
		holders := make([]string, 0, len(q.StudentStatuses))
		for _, item := range q.StudentStatuses {
			holders = append(holders, "?")
			args = append(args, item)
		}
		filters = append(filters, "s.student_status IN ("+strings.Join(holders, ",")+")")
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
		SELECT r.id, r.student_id, IFNULL(s.stu_name, ''), s.stu_sex, IFNULL(s.avatar_url, ''), IFNULL(s.mobile, ''), s.phone_relationship, IFNULL(s.student_status, 0),
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
		ORDER BY `+followUpOrderClause(query.SortModel)+`
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
		var visitStatus sql.NullBool
		if err := rows.Scan(
			&item.ID,
			&item.StudentID,
			&item.StuName,
			&item.StuSex,
			&item.AvatarURL,
			&item.Mobile,
			&item.PhoneRelationship,
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
			&visitStatus,
			&followUpTime,
			&nextFollowUpTime,
		); err != nil {
			return model.PageResult[model.StudentFollowUpRecord]{}, err
		}
		if visitStatus.Valid {
			v := visitStatus.Bool
			item.VisitStatus = &v
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
		item.Mobile = maskPhoneLocal(item.Mobile)
		item.AvatarURL = normalizeStudentAvatarLocal(item.AvatarURL, item.StuSex)
		items = append(items, item)
	}
	if err := attachFollowUpIntentionLessonNames(ctx, repo.db, instID, items); err != nil {
		return model.PageResult[model.StudentFollowUpRecord]{}, err
	}
	return model.PageResult[model.StudentFollowUpRecord]{Items: items, Total: total, Current: current, Size: size}, rows.Err()
}

func attachFollowUpIntentionLessonNames(ctx context.Context, db *sql.DB, instID int64, items []model.StudentFollowUpRecord) error {
	courseIDs := make(map[int64]struct{})
	for i := range items {
		for _, id := range items[i].IntendedCourse {
			if id > 0 {
				courseIDs[id] = struct{}{}
			}
		}
	}
	if len(courseIDs) == 0 {
		return nil
	}
	idList := make([]int64, 0, len(courseIDs))
	for id := range courseIDs {
		idList = append(idList, id)
	}
	placeholders := strings.TrimRight(strings.Repeat("?,", len(idList)), ",")
	args := make([]any, 0, 1+len(idList))
	args = append(args, instID)
	for _, id := range idList {
		args = append(args, id)
	}
	rows, err := db.QueryContext(ctx, `
		SELECT id, IFNULL(name, '')
		FROM inst_course
		WHERE del_flag = 0 AND inst_id = ? AND id IN (`+placeholders+`)
	`, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	nameByID := make(map[int64]string, len(idList))
	for rows.Next() {
		var id int64
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return err
		}
		nameByID[id] = strings.TrimSpace(name)
	}
	for i := range items {
		if len(items[i].IntendedCourse) == 0 {
			continue
		}
		list := make([]model.FollowUpIntentionLesson, 0, len(items[i].IntendedCourse))
		names := make([]string, 0, len(items[i].IntendedCourse))
		for _, cid := range items[i].IntendedCourse {
			nm := nameByID[cid]
			if nm == "" {
				nm = fmt.Sprintf("未知课程(%d)", cid)
			}
			names = append(names, nm)
			list = append(list, model.FollowUpIntentionLesson{LessonID: cid, LessonName: nm})
		}
		items[i].IntendedCourseName = names
		items[i].IntentionLessonList = list
	}
	return rows.Err()
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
			0 AS new_inquiries_added_week_count,
			COUNT(CASE WHEN r.next_follow_up_time < NOW() AND r.visit_status = 0 THEN 1 END) AS overdue_for_follow_up_interview_count
		FROM follow_record r
		WHERE r.inst_id = ? AND r.del_flag = 0
	`, instID)

	var result model.FollowUpCountVO
	err := row.Scan(&result.ToBeFollowedUpTodayCount, &result.NewInquiriesAddedWeekCount, &result.OverdueForFollowUpInterviewCount)
	return result, err
}

func followUpOrderClause(sort model.SortModel) string {
	if sort.ByFollowUpTime != 0 {
		if sort.ByFollowUpTime > 0 {
			return "r.follow_up_time ASC"
		}
		return "r.follow_up_time DESC"
	}
	if sort.ByNextFlowTime != 0 {
		if sort.ByNextFlowTime > 0 {
			return "r.next_follow_up_time ASC"
		}
		return "r.next_follow_up_time DESC"
	}
	return "r.create_time DESC"
}

func normalizeStudentAvatarLocal(avatarURL string, sex *int) string {
	const (
		defaultMaleAvatar    = "https://pcsys.admin.ybc365.com/c04d0ea2-a8b0-4001-b19b-946a980cb726.png"
		defaultFemaleAvatar  = "https://pcsys.admin.ybc365.com/d92afddc-ffac-40aa-aa61-bd97d91aa1ec.png"
		defaultUnknownAvatar = "https://pcsys.admin.ybc365.com/a369a751-2be5-4929-974d-9ae4439f54c4.png"
	)
	normalized := strings.TrimSpace(avatarURL)
	if normalized == defaultUnknownAvatar {
		return defaultMaleAvatar
	}
	if normalized != "" {
		return normalized
	}
	if sex != nil {
		if *sex == 1 {
			return defaultMaleAvatar
		}
		if *sex == 0 {
			return defaultFemaleAvatar
		}
	}
	return defaultMaleAvatar
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

	// 学员主档意向课程与「该学员最新一条跟进」一致：按创建时间最新，其次 id 最大。
	var latestID int64
	var latestIntended string
	err = tx.QueryRowContext(ctx, `
		SELECT id, IFNULL(intended_course, '')
		FROM follow_record
		WHERE student_id = ? AND inst_id = ? AND del_flag = 0
		ORDER BY create_time DESC, id DESC
		LIMIT 1
	`, studentID, instID).Scan(&latestID, &latestIntended)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE inst_student
		SET intended_course = ?
		WHERE id = ? AND inst_id = ? AND del_flag = 0
	`, latestIntended, studentID, instID)
	if err != nil {
		return err
	}

	if latestID == dto.ID {
		_, err = tx.ExecContext(ctx, `
			UPDATE inst_student
			SET follow_up_status = COALESCE(?, follow_up_status),
			    intent_level = COALESCE(?, intent_level),
			    last_follow_up_time = NOW(),
			    next_follow_up_time = ?
			WHERE id = ? AND inst_id = ? AND del_flag = 0
		`, dto.FollowUpStatus, dto.IntentLevel, dto.NextFollowUpTime, studentID, instID)
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
