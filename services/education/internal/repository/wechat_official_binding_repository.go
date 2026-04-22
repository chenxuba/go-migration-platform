package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"go-migration-platform/services/education/internal/model"
)

const (
	weChatOfficialBindTicketStatusPending = 0
	weChatOfficialBindTicketStatusUsed    = 1
)

type WeChatOfficialBindTicketRecord struct {
	Ticket         string
	OfficialOpenID string
	EventKey       string
	SceneValue     string
	InstID         int64
	StudentID      int64
	Status         int
	ExpiresAt      *time.Time
	UsedAt         *time.Time
}

type WeChatOfficialPhoneBindingStatus struct {
	BoundStudentCount   int
	SubscribedBindCount int
	LastUnsubscribeTime *time.Time
}

func ensureWeChatOfficialBindingTables(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS wechat_official_bind_ticket (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			ticket VARCHAR(64) NOT NULL,
			official_openid VARCHAR(128) NOT NULL DEFAULT '',
			event_key VARCHAR(255) NOT NULL DEFAULT '',
			scene_value VARCHAR(255) NOT NULL DEFAULT '',
			inst_id BIGINT NOT NULL DEFAULT 0,
			student_id BIGINT NOT NULL DEFAULT 0,
			status TINYINT NOT NULL DEFAULT 0,
			expires_at DATETIME NULL,
			used_at DATETIME NULL,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE KEY uk_wechat_official_bind_ticket_ticket (ticket),
			KEY idx_wechat_official_bind_ticket_openid (official_openid, create_time),
			KEY idx_wechat_official_bind_ticket_status (status, expires_at)
		)
	`)
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS wechat_official_student_binding (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			inst_id BIGINT NOT NULL,
			student_id BIGINT NOT NULL,
			official_openid VARCHAR(128) NOT NULL DEFAULT '',
			mini_openid VARCHAR(128) NOT NULL DEFAULT '',
			unionid VARCHAR(128) NOT NULL DEFAULT '',
			phone VARCHAR(32) NOT NULL DEFAULT '',
			subscribed TINYINT(1) NOT NULL DEFAULT 1,
			last_bind_ticket VARCHAR(64) NOT NULL DEFAULT '',
			last_subscribe_time DATETIME NULL,
			last_unsubscribe_time DATETIME NULL,
			bind_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE KEY uk_wechat_official_student_binding (inst_id, student_id, official_openid),
			KEY idx_wechat_official_student_binding_openid (official_openid, subscribed),
			KEY idx_wechat_official_student_binding_student (inst_id, student_id, subscribed)
		)
	`)
	if err != nil {
		return err
	}

	return nil
}

func (repo *Repository) CreateWeChatOfficialBindTicket(ctx context.Context, ticket, officialOpenID, eventKey, sceneValue string, instID, studentID int64, expiresAt time.Time) error {
	_, err := repo.db.ExecContext(ctx, `
		INSERT INTO wechat_official_bind_ticket (
			ticket, official_openid, event_key, scene_value, inst_id, student_id, status, expires_at, create_time, update_time
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`,
		strings.TrimSpace(ticket),
		strings.TrimSpace(officialOpenID),
		strings.TrimSpace(eventKey),
		strings.TrimSpace(sceneValue),
		instID,
		studentID,
		weChatOfficialBindTicketStatusPending,
		expiresAt,
	)
	return err
}

func (repo *Repository) GetWeChatOfficialBindTicket(ctx context.Context, ticket string) (WeChatOfficialBindTicketRecord, error) {
	var record WeChatOfficialBindTicketRecord
	var expiresAt sql.NullTime
	var usedAt sql.NullTime
	err := repo.db.QueryRowContext(ctx, `
		SELECT ticket, official_openid, event_key, scene_value, inst_id, student_id, status, expires_at, used_at
		FROM wechat_official_bind_ticket
		WHERE ticket = ?
		LIMIT 1
	`, strings.TrimSpace(ticket)).Scan(
		&record.Ticket,
		&record.OfficialOpenID,
		&record.EventKey,
		&record.SceneValue,
		&record.InstID,
		&record.StudentID,
		&record.Status,
		&expiresAt,
		&usedAt,
	)
	if err != nil {
		return WeChatOfficialBindTicketRecord{}, err
	}
	if expiresAt.Valid {
		value := expiresAt.Time
		record.ExpiresAt = &value
	}
	if usedAt.Valid {
		value := usedAt.Time
		record.UsedAt = &value
	}
	return record, nil
}

func (repo *Repository) MarkWeChatOfficialBindTicketUsed(ctx context.Context, ticket string) error {
	_, err := repo.db.ExecContext(ctx, `
		UPDATE wechat_official_bind_ticket
		SET status = ?, used_at = NOW(), update_time = NOW()
		WHERE ticket = ? AND status = ?
	`, weChatOfficialBindTicketStatusUsed, strings.TrimSpace(ticket), weChatOfficialBindTicketStatusPending)
	return err
}

func (repo *Repository) FindInstitutionAndStudentByScene(ctx context.Context, sceneValue string) (int64, int64, error) {
	sceneValue = strings.TrimSpace(sceneValue)
	if sceneValue == "" {
		return 0, 0, nil
	}

	lower := strings.ToLower(sceneValue)
	if strings.HasPrefix(lower, "student_") {
		var studentID int64
		if _, err := fmt.Sscanf(sceneValue, "student_%d", &studentID); err == nil && studentID > 0 {
			instID, err := repo.FindInstIDByStudentID(ctx, studentID)
			if err != nil {
				if err == sql.ErrNoRows {
					return 0, studentID, nil
				}
				return 0, 0, err
			}
			return instID, studentID, nil
		}
	}

	for _, pattern := range []string{"inst_%d", "institution_%d", "campus_%d", "school_%d"} {
		var instID int64
		if _, err := fmt.Sscanf(sceneValue, pattern, &instID); err == nil && instID > 0 {
			return instID, 0, nil
		}
	}

	return 0, 0, nil
}

func (repo *Repository) FindInstIDByStudentID(ctx context.Context, studentID int64) (int64, error) {
	var instID int64
	err := repo.db.QueryRowContext(ctx, `
		SELECT inst_id
		FROM inst_student
		WHERE id = ? AND del_flag = 0
		LIMIT 1
	`, studentID).Scan(&instID)
	return instID, err
}

func (repo *Repository) ListWeChatOfficialBindStudentCandidatesByPhone(ctx context.Context, instID int64, phone string) ([]model.WeChatOfficialBindStudentCandidateVO, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT s.id, IFNULL(s.stu_name, ''), IFNULL(s.avatar_url, ''), IFNULL(s.mobile, ''), IFNULL(s.student_status, 0), IFNULL(s.is_bind_child, 0)
		FROM inst_student s
		WHERE s.inst_id = ? AND s.del_flag = 0 AND IFNULL(s.mobile, '') = ?
		ORDER BY s.create_time DESC, s.id DESC
	`, instID, strings.TrimSpace(phone))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.WeChatOfficialBindStudentCandidateVO, 0, 4)
	for rows.Next() {
		var item model.WeChatOfficialBindStudentCandidateVO
		var isBindChild int
		if err := rows.Scan(&item.ID, &item.StuName, &item.AvatarURL, &item.Mobile, &item.StudentStatus, &isBindChild); err != nil {
			return nil, err
		}
		item.IsBound = isBindChild != 0
		items = append(items, item)
	}
	return items, rows.Err()
}

func (repo *Repository) GetStudentBaseInfo(ctx context.Context, studentID int64) (model.WeChatOfficialBindStudentCandidateVO, int64, error) {
	var item model.WeChatOfficialBindStudentCandidateVO
	var instID int64
	var isBindChild int
	err := repo.db.QueryRowContext(ctx, `
		SELECT s.id, IFNULL(s.stu_name, ''), IFNULL(s.avatar_url, ''), IFNULL(s.mobile, ''), IFNULL(s.student_status, 0), IFNULL(s.is_bind_child, 0), s.inst_id
		FROM inst_student s
		WHERE s.id = ? AND s.del_flag = 0
		LIMIT 1
	`, studentID).Scan(&item.ID, &item.StuName, &item.AvatarURL, &item.Mobile, &item.StudentStatus, &isBindChild, &instID)
	if err != nil {
		return model.WeChatOfficialBindStudentCandidateVO{}, 0, err
	}
	item.IsBound = isBindChild != 0
	return item, instID, nil
}

func (repo *Repository) UpsertWeChatOfficialStudentBinding(ctx context.Context, instID, studentID int64, officialOpenID, miniOpenID, unionID, phone, bindTicket string, subscribed bool) error {
	subscribedValue := 0
	if subscribed {
		subscribedValue = 1
	}
	_, err := repo.db.ExecContext(ctx, `
		INSERT INTO wechat_official_student_binding (
			inst_id, student_id, official_openid, mini_openid, unionid, phone, subscribed, last_bind_ticket,
			last_subscribe_time, last_unsubscribe_time, bind_time, create_time, update_time
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, CASE WHEN ? = 1 THEN NOW() ELSE NULL END, CASE WHEN ? = 1 THEN NULL ELSE NOW() END, NOW(), NOW(), NOW())
		ON DUPLICATE KEY UPDATE
			mini_openid = CASE WHEN VALUES(mini_openid) = '' THEN mini_openid ELSE VALUES(mini_openid) END,
			unionid = CASE WHEN VALUES(unionid) = '' THEN unionid ELSE VALUES(unionid) END,
			phone = CASE WHEN VALUES(phone) = '' THEN phone ELSE VALUES(phone) END,
			subscribed = VALUES(subscribed),
			last_bind_ticket = VALUES(last_bind_ticket),
			last_subscribe_time = CASE WHEN VALUES(subscribed) = 1 THEN NOW() ELSE last_subscribe_time END,
			last_unsubscribe_time = CASE WHEN VALUES(subscribed) = 0 THEN NOW() ELSE last_unsubscribe_time END,
			update_time = NOW()
	`,
		instID,
		studentID,
		strings.TrimSpace(officialOpenID),
		strings.TrimSpace(miniOpenID),
		strings.TrimSpace(unionID),
		strings.TrimSpace(phone),
		subscribedValue,
		strings.TrimSpace(bindTicket),
		subscribedValue,
		subscribedValue,
	)
	return err
}

func (repo *Repository) UpdateWeChatOfficialBindingSubscriptionByOpenID(ctx context.Context, officialOpenID string, subscribed bool) ([]int64, error) {
	officialOpenID = strings.TrimSpace(officialOpenID)
	if officialOpenID == "" {
		return nil, nil
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT DISTINCT student_id
		FROM wechat_official_student_binding
		WHERE official_openid = ?
	`, officialOpenID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	studentIDs := make([]int64, 0, 8)
	for rows.Next() {
		var studentID int64
		if err := rows.Scan(&studentID); err != nil {
			return nil, err
		}
		if studentID > 0 {
			studentIDs = append(studentIDs, studentID)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	subscribedValue := 0
	if subscribed {
		subscribedValue = 1
	}
	_, err = repo.db.ExecContext(ctx, `
		UPDATE wechat_official_student_binding
		SET subscribed = ?,
			last_subscribe_time = CASE WHEN ? = 1 THEN NOW() ELSE last_subscribe_time END,
			last_unsubscribe_time = CASE WHEN ? = 0 THEN NOW() ELSE last_unsubscribe_time END,
			update_time = NOW()
		WHERE official_openid = ?
	`, subscribedValue, subscribedValue, subscribedValue, officialOpenID)
	if err != nil {
		return nil, err
	}
	return studentIDs, nil
}

func (repo *Repository) RefreshStudentBindChildStatus(ctx context.Context, studentID int64) error {
	if studentID <= 0 {
		return nil
	}

	var bindCount int
	if err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM wechat_official_student_binding
		WHERE student_id = ? AND subscribed = 1
	`, studentID).Scan(&bindCount); err != nil {
		return err
	}

	isBindChild := 0
	if bindCount > 0 {
		isBindChild = 1
	}
	_, err := repo.db.ExecContext(ctx, `
		UPDATE inst_student
		SET is_bind_child = ?, update_time = NOW()
		WHERE id = ? AND del_flag = 0
	`, isBindChild, studentID)
	return err
}

func (repo *Repository) HasWeChatOfficialBoundStudent(ctx context.Context, officialOpenID string, instID int64) (bool, error) {
	var count int
	err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM wechat_official_student_binding
		WHERE official_openid = ? AND inst_id = ? AND subscribed = 1
	`, strings.TrimSpace(officialOpenID), instID).Scan(&count)
	return count > 0, err
}

func (repo *Repository) GetWeChatOfficialBindingStatusByPhone(ctx context.Context, phone string) (WeChatOfficialPhoneBindingStatus, error) {
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return WeChatOfficialPhoneBindingStatus{}, nil
	}

	var status WeChatOfficialPhoneBindingStatus
	var lastUnsubscribeAt sql.NullTime
	err := repo.db.QueryRowContext(ctx, `
		SELECT
			COUNT(DISTINCT student_id) AS bound_student_count,
			COUNT(DISTINCT CASE WHEN subscribed = 1 THEN CONCAT(inst_id, ':', student_id, ':', official_openid) END) AS subscribed_bind_count,
			MAX(last_unsubscribe_time) AS last_unsubscribe_time
		FROM wechat_official_student_binding
		WHERE phone = ?
	`, phone).Scan(
		&status.BoundStudentCount,
		&status.SubscribedBindCount,
		&lastUnsubscribeAt,
	)
	if err != nil {
		return WeChatOfficialPhoneBindingStatus{}, err
	}
	if lastUnsubscribeAt.Valid {
		value := lastUnsubscribeAt.Time
		status.LastUnsubscribeTime = &value
	}
	return status, nil
}
