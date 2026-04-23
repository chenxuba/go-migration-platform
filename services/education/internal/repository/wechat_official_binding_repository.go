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

type WeChatOfficialUserFollowStatus struct {
	SubscribedUserCount int
	LastUnsubscribeTime *time.Time
}

type WeChatOfficialStudentRecipient struct {
	StudentID      int64
	OfficialOpenID string
}

type weChatOfficialUserLinkRow struct {
	ID             int64
	OfficialOpenID string
	MiniOpenID     string
	UnionID        string
	Phone          string
	Subscribed     bool
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

	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS wechat_official_user_link (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			official_openid VARCHAR(128) NULL DEFAULT NULL,
			mini_openid VARCHAR(128) NULL DEFAULT NULL,
			unionid VARCHAR(128) NULL DEFAULT NULL,
			phone VARCHAR(32) NOT NULL DEFAULT '',
			subscribed TINYINT(1) NOT NULL DEFAULT 0,
			last_subscribe_time DATETIME NULL,
			last_unsubscribe_time DATETIME NULL,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE KEY uk_wechat_official_user_link_official_openid (official_openid),
			UNIQUE KEY uk_wechat_official_user_link_mini_openid (mini_openid),
			UNIQUE KEY uk_wechat_official_user_link_unionid (unionid),
			KEY idx_wechat_official_user_link_phone (phone, subscribed),
			KEY idx_wechat_official_user_link_subscribed (subscribed, update_time)
		)
	`)
	if err != nil {
		return err
	}

	return nil
}

func nullableTrimmedString(value string) any {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return trimmed
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

func (repo *Repository) ListSubscribedWeChatOfficialRecipientsByStudentIDs(ctx context.Context, instID int64, studentIDs []int64) ([]WeChatOfficialStudentRecipient, error) {
	if instID <= 0 {
		return []WeChatOfficialStudentRecipient{}, nil
	}

	ids := uniquePositiveInt64s(studentIDs)
	if len(ids) == 0 {
		return []WeChatOfficialStudentRecipient{}, nil
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT DISTINCT
			wsb.student_id,
			wsb.official_openid
		FROM wechat_official_student_binding wsb
		INNER JOIN wechat_official_user_link ul
			ON ul.official_openid = wsb.official_openid
		WHERE wsb.inst_id = ?
		  AND wsb.student_id IN (`+sqlPlaceholders(len(ids))+`)
		  AND wsb.subscribed = 1
		  AND IFNULL(wsb.official_openid, '') <> ''
		  AND IFNULL(ul.subscribed, 0) = 1
	`, append([]any{instID}, int64SliceToAny(ids)...)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]WeChatOfficialStudentRecipient, 0, len(ids))
	for rows.Next() {
		var item WeChatOfficialStudentRecipient
		if err := rows.Scan(&item.StudentID, &item.OfficialOpenID); err != nil {
			return nil, err
		}
		item.OfficialOpenID = strings.TrimSpace(item.OfficialOpenID)
		if item.StudentID <= 0 || item.OfficialOpenID == "" {
			continue
		}
		result = append(result, item)
	}
	return result, rows.Err()
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

func (repo *Repository) UpsertWeChatOfficialUserLinkByOfficialProfile(ctx context.Context, officialOpenID, unionID string, subscribed bool) error {
	return repo.upsertWeChatOfficialUserLinkMerged(ctx, officialOpenID, "", unionID, "", subscribed)
}

func (repo *Repository) GetWeChatOfficialUserLinkByOfficialOpenID(ctx context.Context, officialOpenID string) (ParentWeChatOfficialUserLinkRecord, error) {
	record, err := repo.findWeChatOfficialUserLinkByOfficialOpenID(ctx, officialOpenID)
	if err != nil {
		return ParentWeChatOfficialUserLinkRecord{}, err
	}
	return ParentWeChatOfficialUserLinkRecord{
		ID:             record.ID,
		OfficialOpenID: record.OfficialOpenID,
		MiniOpenID:     record.MiniOpenID,
		UnionID:        record.UnionID,
		Phone:          record.Phone,
		Subscribed:     record.Subscribed,
	}, nil
}

func (repo *Repository) UpsertWeChatOfficialUserLinkByMiniProfile(ctx context.Context, miniOpenID, unionID, phone string) error {
	miniOpenID = strings.TrimSpace(miniOpenID)
	unionID = strings.TrimSpace(unionID)
	phone = strings.TrimSpace(phone)
	if miniOpenID == "" && unionID == "" && phone == "" {
		return nil
	}

	_, err := repo.db.ExecContext(ctx, `
		INSERT INTO wechat_official_user_link (
			mini_openid, unionid, phone, create_time, update_time
		) VALUES (?, ?, ?, NOW(), NOW())
		ON DUPLICATE KEY UPDATE
			mini_openid = CASE WHEN VALUES(mini_openid) IS NULL THEN mini_openid ELSE VALUES(mini_openid) END,
			unionid = CASE WHEN VALUES(unionid) IS NULL THEN unionid ELSE VALUES(unionid) END,
			phone = CASE WHEN VALUES(phone) = '' THEN phone ELSE VALUES(phone) END,
			update_time = NOW()
	`,
		nullableTrimmedString(miniOpenID),
		nullableTrimmedString(unionID),
		phone,
	)
	return err
}

func (repo *Repository) UpsertWeChatOfficialUserLink(ctx context.Context, officialOpenID, miniOpenID, unionID, phone string, subscribed bool) error {
	return repo.upsertWeChatOfficialUserLinkMerged(ctx, officialOpenID, miniOpenID, unionID, phone, subscribed)
}

func (repo *Repository) RepairWeChatOfficialStudentBindingsByUserLink(ctx context.Context, officialOpenID, miniOpenID, unionID, phone string) ([]int64, error) {
	officialOpenID = strings.TrimSpace(officialOpenID)
	miniOpenID = strings.TrimSpace(miniOpenID)
	unionID = strings.TrimSpace(unionID)
	phone = strings.TrimSpace(phone)
	if officialOpenID == "" {
		return nil, nil
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT DISTINCT student_id
		FROM wechat_official_student_binding
		WHERE official_openid = ?
		   OR (
				IFNULL(official_openid, '') = ''
				AND (
					(? <> '' AND mini_openid = ?)
				 OR (? <> '' AND unionid = ?)
				 OR (? <> '' AND phone = ?)
				)
		   )
	`, officialOpenID, miniOpenID, miniOpenID, unionID, unionID, phone, phone)
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

	if miniOpenID == "" && unionID == "" && phone == "" {
		return studentIDs, nil
	}

	_, err = repo.db.ExecContext(ctx, `
		UPDATE wechat_official_student_binding wsb
		LEFT JOIN (
			SELECT DISTINCT inst_id, student_id
			FROM wechat_official_student_binding
			WHERE official_openid = ?
		) existing ON existing.inst_id = wsb.inst_id AND existing.student_id = wsb.student_id
		SET wsb.official_openid = ?,
			wsb.update_time = NOW()
		WHERE IFNULL(wsb.official_openid, '') = ''
		  AND existing.student_id IS NULL
		  AND (
				(? <> '' AND wsb.mini_openid = ?)
			 OR (? <> '' AND wsb.unionid = ?)
			 OR (? <> '' AND wsb.phone = ?)
		  )
	`, officialOpenID, officialOpenID, miniOpenID, miniOpenID, unionID, unionID, phone, phone)
	if err != nil {
		return nil, err
	}

	return studentIDs, nil
}

func (repo *Repository) RepairWeChatOfficialUserLinkByPhone(ctx context.Context, miniOpenID, unionID, phone string) error {
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return nil
	}

	record, err := repo.findSubscribedWeChatOfficialUserLinkByPhone(ctx, phone)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}
	if strings.TrimSpace(record.OfficialOpenID) == "" {
		return nil
	}

	return repo.upsertWeChatOfficialUserLinkMerged(ctx, record.OfficialOpenID, miniOpenID, unionID, phone, true)
}

func (repo *Repository) upsertWeChatOfficialUserLinkMerged(ctx context.Context, officialOpenID, miniOpenID, unionID, phone string, subscribed bool) error {
	officialOpenID = strings.TrimSpace(officialOpenID)
	miniOpenID = strings.TrimSpace(miniOpenID)
	unionID = strings.TrimSpace(unionID)
	phone = strings.TrimSpace(phone)
	if officialOpenID == "" && miniOpenID == "" && unionID == "" && phone == "" {
		return nil
	}

	officialRow, err := repo.findWeChatOfficialUserLinkByOfficialOpenID(ctx, officialOpenID)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	miniRow, err := repo.findWeChatOfficialUserLinkByMiniIdentity(ctx, miniOpenID, unionID)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	finalOfficialOpenID := firstNonEmptyTrimmedString(officialOpenID, officialRow.OfficialOpenID, miniRow.OfficialOpenID)
	finalMiniOpenID := firstNonEmptyTrimmedString(miniOpenID, officialRow.MiniOpenID, miniRow.MiniOpenID)
	finalUnionID := firstNonEmptyTrimmedString(unionID, officialRow.UnionID, miniRow.UnionID)
	finalPhone := firstNonEmptyTrimmedString(phone, officialRow.Phone, miniRow.Phone)

	if officialRow.ID <= 0 && miniRow.ID <= 0 {
		subscribedValue := 0
		if subscribed {
			subscribedValue = 1
		}

		_, err := repo.db.ExecContext(ctx, `
		INSERT INTO wechat_official_user_link (
			official_openid, mini_openid, unionid, phone, subscribed, last_subscribe_time, last_unsubscribe_time, create_time, update_time
		) VALUES (?, ?, ?, ?, ?, CASE WHEN ? = 1 THEN NOW() ELSE NULL END, CASE WHEN ? = 0 THEN NOW() ELSE NULL END, NOW(), NOW())
		ON DUPLICATE KEY UPDATE
			official_openid = CASE WHEN VALUES(official_openid) IS NULL THEN official_openid ELSE VALUES(official_openid) END,
			mini_openid = CASE WHEN VALUES(mini_openid) IS NULL THEN mini_openid ELSE VALUES(mini_openid) END,
			unionid = CASE WHEN VALUES(unionid) IS NULL THEN unionid ELSE VALUES(unionid) END,
			phone = CASE WHEN VALUES(phone) = '' THEN phone ELSE VALUES(phone) END,
			subscribed = VALUES(subscribed),
			last_subscribe_time = CASE WHEN VALUES(subscribed) = 1 THEN NOW() ELSE last_subscribe_time END,
			last_unsubscribe_time = CASE WHEN VALUES(subscribed) = 0 THEN NOW() ELSE last_unsubscribe_time END,
			update_time = NOW()
		`,
			nullableTrimmedString(finalOfficialOpenID),
			nullableTrimmedString(finalMiniOpenID),
			nullableTrimmedString(finalUnionID),
			finalPhone,
			subscribedValue,
			subscribedValue,
			subscribedValue,
		)
		return err
	}

	target := officialRow
	source := miniRow
	if target.ID <= 0 {
		target = miniRow
		source = weChatOfficialUserLinkRow{}
	}
	if target.ID > 0 && !canMergeWeChatOfficialUserLinkRow(target, finalOfficialOpenID, finalMiniOpenID, finalUnionID) {
		return fmt.Errorf("wechat official user link conflict for official_openid=%s", finalOfficialOpenID)
	}
	if source.ID > 0 && source.ID != target.ID && strings.TrimSpace(source.OfficialOpenID) != "" && strings.TrimSpace(source.OfficialOpenID) != finalOfficialOpenID {
		return fmt.Errorf("wechat official user link conflict for mini_openid=%s", finalMiniOpenID)
	}

	if source.ID > 0 && source.ID != target.ID {
		if err := repo.detachWeChatOfficialUserLinkMiniIdentity(ctx, source.ID); err != nil {
			return err
		}
	}
	if err := repo.updateWeChatOfficialUserLinkByID(ctx, target.ID, finalOfficialOpenID, finalMiniOpenID, finalUnionID, finalPhone, subscribed); err != nil {
		return err
	}
	if source.ID > 0 && source.ID != target.ID {
		if err := repo.deleteWeChatOfficialUserLinkIfDetached(ctx, source.ID); err != nil {
			return err
		}
	}
	return nil
}

func (repo *Repository) findWeChatOfficialUserLinkByOfficialOpenID(ctx context.Context, officialOpenID string) (weChatOfficialUserLinkRow, error) {
	officialOpenID = strings.TrimSpace(officialOpenID)
	if officialOpenID == "" {
		return weChatOfficialUserLinkRow{}, sql.ErrNoRows
	}

	return repo.findWeChatOfficialUserLinkRow(ctx, `
		SELECT
			id,
			IFNULL(official_openid, ''),
			IFNULL(mini_openid, ''),
			IFNULL(unionid, ''),
			IFNULL(phone, ''),
			IFNULL(subscribed, 0)
		FROM wechat_official_user_link
		WHERE official_openid = ?
		LIMIT 1
	`, officialOpenID)
}

func (repo *Repository) findWeChatOfficialUserLinkByMiniIdentity(ctx context.Context, miniOpenID, unionID string) (weChatOfficialUserLinkRow, error) {
	miniOpenID = strings.TrimSpace(miniOpenID)
	unionID = strings.TrimSpace(unionID)

	if miniOpenID != "" {
		record, err := repo.findWeChatOfficialUserLinkRow(ctx, `
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
			return record, nil
		}
		if err != sql.ErrNoRows {
			return weChatOfficialUserLinkRow{}, err
		}
	}

	if unionID != "" {
		return repo.findWeChatOfficialUserLinkRow(ctx, `
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

	return weChatOfficialUserLinkRow{}, sql.ErrNoRows
}

func (repo *Repository) findSubscribedWeChatOfficialUserLinkByPhone(ctx context.Context, phone string) (weChatOfficialUserLinkRow, error) {
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return weChatOfficialUserLinkRow{}, sql.ErrNoRows
	}

	return repo.findWeChatOfficialUserLinkRow(ctx, `
		SELECT
			id,
			IFNULL(official_openid, ''),
			IFNULL(mini_openid, ''),
			IFNULL(unionid, ''),
			IFNULL(phone, ''),
			IFNULL(subscribed, 0)
		FROM wechat_official_user_link
		WHERE phone = ? AND subscribed = 1 AND IFNULL(official_openid, '') <> ''
		ORDER BY update_time DESC, id DESC
		LIMIT 1
	`, phone)
}

func (repo *Repository) findWeChatOfficialUserLinkRow(ctx context.Context, query string, args ...any) (weChatOfficialUserLinkRow, error) {
	var record weChatOfficialUserLinkRow
	var subscribed int
	err := repo.db.QueryRowContext(ctx, query, args...).Scan(
		&record.ID,
		&record.OfficialOpenID,
		&record.MiniOpenID,
		&record.UnionID,
		&record.Phone,
		&subscribed,
	)
	if err != nil {
		return weChatOfficialUserLinkRow{}, err
	}
	record.Subscribed = subscribed != 0
	return record, nil
}

func (repo *Repository) updateWeChatOfficialUserLinkByID(ctx context.Context, rowID int64, officialOpenID, miniOpenID, unionID, phone string, subscribed bool) error {
	subscribedValue := 0
	if subscribed {
		subscribedValue = 1
	}

	_, err := repo.db.ExecContext(ctx, `
		UPDATE wechat_official_user_link
		SET official_openid = CASE WHEN ? = '' THEN official_openid ELSE ? END,
			mini_openid = CASE WHEN ? = '' THEN mini_openid ELSE ? END,
			unionid = CASE WHEN ? = '' THEN unionid ELSE ? END,
			phone = CASE WHEN ? = '' THEN phone ELSE ? END,
			subscribed = ?,
			last_subscribe_time = CASE WHEN ? = 1 THEN NOW() ELSE last_subscribe_time END,
			last_unsubscribe_time = CASE WHEN ? = 0 THEN NOW() ELSE last_unsubscribe_time END,
			update_time = NOW()
		WHERE id = ?
	`,
		officialOpenID,
		nullableTrimmedString(officialOpenID),
		miniOpenID,
		nullableTrimmedString(miniOpenID),
		unionID,
		nullableTrimmedString(unionID),
		phone,
		phone,
		subscribedValue,
		subscribedValue,
		subscribedValue,
		rowID,
	)
	return err
}

func (repo *Repository) detachWeChatOfficialUserLinkMiniIdentity(ctx context.Context, rowID int64) error {
	_, err := repo.db.ExecContext(ctx, `
		UPDATE wechat_official_user_link
		SET mini_openid = NULL,
			unionid = NULL,
			update_time = NOW()
		WHERE id = ? AND IFNULL(official_openid, '') = ''
	`, rowID)
	return err
}

func (repo *Repository) deleteWeChatOfficialUserLinkIfDetached(ctx context.Context, rowID int64) error {
	_, err := repo.db.ExecContext(ctx, `
		DELETE FROM wechat_official_user_link
		WHERE id = ?
		  AND official_openid IS NULL
		  AND mini_openid IS NULL
		  AND unionid IS NULL
	`, rowID)
	return err
}

func canMergeWeChatOfficialUserLinkRow(row weChatOfficialUserLinkRow, officialOpenID, miniOpenID, unionID string) bool {
	if row.ID <= 0 {
		return true
	}
	if officialOpenID != "" && strings.TrimSpace(row.OfficialOpenID) != "" && strings.TrimSpace(row.OfficialOpenID) != officialOpenID {
		return false
	}
	if miniOpenID != "" && strings.TrimSpace(row.MiniOpenID) != "" && strings.TrimSpace(row.MiniOpenID) != miniOpenID {
		return false
	}
	if unionID != "" && strings.TrimSpace(row.UnionID) != "" && strings.TrimSpace(row.UnionID) != unionID {
		return false
	}
	return true
}

func firstNonEmptyTrimmedString(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
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

	_, err = repo.db.ExecContext(ctx, `
		UPDATE wechat_official_student_binding
		SET subscribed = ?,
			last_subscribe_time = CASE WHEN ? = 1 THEN NOW() ELSE last_subscribe_time END,
			last_unsubscribe_time = CASE WHEN ? = 0 THEN NOW() ELSE last_unsubscribe_time END,
			update_time = NOW()
		WHERE official_openid = ?
	`, boolToTinyInt(subscribed), boolToTinyInt(subscribed), boolToTinyInt(subscribed), officialOpenID)
	if err != nil {
		return nil, err
	}
	return studentIDs, nil
}

func (repo *Repository) HasWeChatOfficialSubscribedBindingByOpenID(ctx context.Context, officialOpenID string) (bool, error) {
	officialOpenID = strings.TrimSpace(officialOpenID)
	if officialOpenID == "" {
		return false, nil
	}

	var count int
	err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM wechat_official_student_binding
		WHERE official_openid = ? AND subscribed = 1
	`, officialOpenID).Scan(&count)
	return count > 0, err
}

func (repo *Repository) ListWeChatOfficialOpenIDsByPhone(ctx context.Context, phone string) ([]string, error) {
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return nil, nil
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT DISTINCT openid
		FROM (
			SELECT official_openid AS openid
			FROM wechat_official_user_link
			WHERE phone = ? AND IFNULL(official_openid, '') <> ''
			UNION ALL
			SELECT official_openid AS openid
			FROM wechat_official_student_binding
			WHERE phone = ? AND IFNULL(official_openid, '') <> ''
		) records
		WHERE IFNULL(openid, '') <> ''
		ORDER BY openid
	`, phone, phone)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	openIDs := make([]string, 0, 4)
	for rows.Next() {
		var openID string
		if err := rows.Scan(&openID); err != nil {
			return nil, err
		}
		openID = strings.TrimSpace(openID)
		if openID == "" {
			continue
		}
		openIDs = append(openIDs, openID)
	}
	return openIDs, rows.Err()
}

func (repo *Repository) FindWeChatOfficialLinkedPhoneByOpenID(ctx context.Context, officialOpenID string) (string, error) {
	officialOpenID = strings.TrimSpace(officialOpenID)
	if officialOpenID == "" {
		return "", nil
	}

	var phone string
	err := repo.db.QueryRowContext(ctx, `
		SELECT IFNULL(phone, '')
		FROM wechat_official_user_link
		WHERE official_openid = ?
		LIMIT 1
	`, officialOpenID).Scan(&phone)
	if err != nil && err != sql.ErrNoRows {
		return "", err
	}
	phone = strings.TrimSpace(phone)
	if phone != "" {
		return phone, nil
	}

	err = repo.db.QueryRowContext(ctx, `
		SELECT IFNULL(phone, '')
		FROM wechat_official_student_binding
		WHERE official_openid = ? AND IFNULL(phone, '') <> ''
		ORDER BY update_time DESC, id DESC
		LIMIT 1
	`, officialOpenID).Scan(&phone)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}
	return strings.TrimSpace(phone), nil
}

func (repo *Repository) DeleteWeChatOfficialStudentBindingsByIdentity(ctx context.Context, officialOpenID, miniOpenID, unionID string) ([]int64, error) {
	officialOpenID = strings.TrimSpace(officialOpenID)
	miniOpenID = strings.TrimSpace(miniOpenID)
	unionID = strings.TrimSpace(unionID)
	if officialOpenID == "" && miniOpenID == "" && unionID == "" {
		return nil, nil
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT DISTINCT student_id
		FROM wechat_official_student_binding
		WHERE (? <> '' AND official_openid = ?)
		   OR (? <> '' AND mini_openid = ?)
		   OR (? <> '' AND unionid = ?)
	`, officialOpenID, officialOpenID, miniOpenID, miniOpenID, unionID, unionID)
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

	if _, err := repo.db.ExecContext(ctx, `
		DELETE FROM wechat_official_student_binding
		WHERE (? <> '' AND official_openid = ?)
		   OR (? <> '' AND mini_openid = ?)
		   OR (? <> '' AND unionid = ?)
	`, officialOpenID, officialOpenID, miniOpenID, miniOpenID, unionID, unionID); err != nil {
		return nil, err
	}
	return studentIDs, nil
}

func (repo *Repository) DeleteWeChatOfficialUserLinkByID(ctx context.Context, rowID int64) error {
	if rowID <= 0 {
		return nil
	}

	_, err := repo.db.ExecContext(ctx, `
		DELETE FROM wechat_official_user_link
		WHERE id = ?
	`, rowID)
	return err
}

func (repo *Repository) ClearWeChatOfficialUserLinkMiniIdentityByID(ctx context.Context, rowID int64) error {
	if rowID <= 0 {
		return nil
	}

	_, err := repo.db.ExecContext(ctx, `
		UPDATE wechat_official_user_link
		SET mini_openid = NULL,
			unionid = NULL,
			update_time = NOW()
		WHERE id = ?
	`, rowID)
	return err
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

func (repo *Repository) RefreshStudentBindChildStatusByPhone(ctx context.Context, phone string) error {
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return nil
	}

	_, err := repo.db.ExecContext(ctx, `
		UPDATE inst_student s
		LEFT JOIN (
			SELECT DISTINCT student_id
			FROM wechat_official_student_binding
			WHERE subscribed = 1
		) bound ON bound.student_id = s.id
		SET s.is_bind_child = CASE WHEN bound.student_id IS NULL THEN 0 ELSE 1 END,
			s.update_time = NOW()
		WHERE IFNULL(s.mobile, '') = ? AND s.del_flag = 0
	`, phone)
	return err
}

func (repo *Repository) HasWeChatOfficialBoundStudent(ctx context.Context, officialOpenID string, instID int64) (bool, error) {
	var count int
	err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM wechat_official_student_binding
		WHERE official_openid = ? AND inst_id = ? 
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
			COUNT(DISTINCT CASE
				WHEN IFNULL(sb.official_openid, '') <> ''
				THEN sb.student_id
			END) AS bound_student_count,
			COUNT(DISTINCT CASE
				WHEN sb.subscribed = 1
				 AND IFNULL(sb.official_openid, '') <> ''
				 AND IFNULL(ul.subscribed, 0) = 1
				THEN CONCAT(sb.inst_id, ':', sb.student_id, ':', sb.official_openid)
			END) AS subscribed_bind_count,
			MAX(sb.last_unsubscribe_time) AS last_unsubscribe_time
		FROM wechat_official_student_binding sb
		LEFT JOIN wechat_official_user_link ul ON ul.official_openid = sb.official_openid
		WHERE sb.phone = ?
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

func (repo *Repository) GetWeChatOfficialUserFollowStatusByPhone(ctx context.Context, phone string) (WeChatOfficialUserFollowStatus, error) {
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return WeChatOfficialUserFollowStatus{}, nil
	}

	var status WeChatOfficialUserFollowStatus
	var lastUnsubscribeAt sql.NullTime
	err := repo.db.QueryRowContext(ctx, `
		SELECT
			COUNT(DISTINCT CASE
				WHEN subscribed = 1 AND IFNULL(official_openid, '') <> ''
				THEN COALESCE(unionid, official_openid, mini_openid)
			END) AS subscribed_user_count,
			MAX(last_unsubscribe_time) AS last_unsubscribe_time
		FROM wechat_official_user_link
		WHERE phone = ?
	`, phone).Scan(
		&status.SubscribedUserCount,
		&lastUnsubscribeAt,
	)
	if err != nil {
		return WeChatOfficialUserFollowStatus{}, err
	}
	if lastUnsubscribeAt.Valid {
		value := lastUnsubscribeAt.Time
		status.LastUnsubscribeTime = &value
	}
	return status, nil
}
