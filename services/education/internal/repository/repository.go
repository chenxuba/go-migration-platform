package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"
)

type Repository struct {
	db *sql.DB
}

type StudentSnapshot struct {
	ID                 int64
	InstID             int64
	StuName            string
	Mobile             string
	PhoneRelationship  *int
	SalePerson         *int64
	ChannelID          *int64
	CollectorStaffID   *int64
	PhoneSellStaffID   *int64
	ForegroundStaffID  *int64
	ViceSellStaffID    *int64
	StudentManagerID   *int64
	AdvisorID          *int64
	RecommendStudentID *int64
	WeChatNumber       string
	Grade              string
	StudySchool        string
	Interest           string
	Address            string
	Remark             string
	FollowUpStatus     *int
	IntentLevel        *int
}

func New(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (repo *Repository) EnsureInfrastructureTables(ctx context.Context) error {
	_, err := repo.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS mq_event_log (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			topic VARCHAR(255) NOT NULL,
			tag VARCHAR(255) NULL,
			payload LONGTEXT NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}
	var exists int
	if err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE()
		  AND TABLE_NAME = 'approval_record'
		  AND COLUMN_NAME = 'initiate_reason'
	`).Scan(&exists); err != nil {
		return err
	}
	if exists > 0 {
		return nil
	}
	_, err = repo.db.ExecContext(ctx, `
		ALTER TABLE approval_record
		ADD COLUMN initiate_reason VARCHAR(1000) NULL DEFAULT NULL COMMENT '审批发起时的触发条件快照'
	`)
	return err
}

func (repo *Repository) FindInstIDByUserID(ctx context.Context, userID int64) (int64, error) {
	row := repo.db.QueryRowContext(ctx, `
		SELECT u.inst_id
		FROM inst_user u
		LEFT JOIN org_institution i ON u.inst_id = i.id
		WHERE u.del_flag = 0 AND u.disabled = 0
		  AND i.del_flag = 0 AND i.enabled = 1
		  AND i.expire_end_time > NOW()
		  AND u.user_id = ?
		  AND i.organ_type != 2 AND i.organ_type != 10 AND i.organ_type != 11
		ORDER BY u.id
		LIMIT 1
	`, userID)

	var instID int64
	err := row.Scan(&instID)
	return instID, err
}

func (repo *Repository) FindInstUserIDByUserID(ctx context.Context, userID int64) (int64, error) {
	row := repo.db.QueryRowContext(ctx, `
		SELECT u.id
		FROM inst_user u
		LEFT JOIN org_institution i ON u.inst_id = i.id
		WHERE u.del_flag = 0 AND u.disabled = 0
		  AND i.del_flag = 0 AND i.enabled = 1
		  AND i.expire_end_time > NOW()
		  AND u.user_id = ?
		  AND i.organ_type != 2 AND i.organ_type != 10 AND i.organ_type != 11
		ORDER BY u.id
		LIMIT 1
	`, userID)

	var instUserID int64
	err := row.Scan(&instUserID)
	return instUserID, err
}

func (repo *Repository) GetStaffNameByID(ctx context.Context, staffID *int64) string {
	if staffID == nil {
		return "-"
	}
	var nickName string
	err := repo.db.QueryRowContext(ctx, "SELECT IFNULL(nick_name, '') FROM inst_user WHERE id = ? LIMIT 1", *staffID).Scan(&nickName)
	if err != nil || strings.TrimSpace(nickName) == "" {
		return fmt.Sprintf("未知(%d)", *staffID)
	}
	return nickName
}

func (repo *Repository) GetChannelNameByID(ctx context.Context, channelID *int64) string {
	if channelID == nil {
		return "-"
	}
	var name string
	err := repo.db.QueryRowContext(ctx, "SELECT IFNULL(channel_name, '') FROM inst_channel WHERE id = ? LIMIT 1", *channelID).Scan(&name)
	if err != nil || strings.TrimSpace(name) == "" {
		return fmt.Sprintf("未知渠道(%d)", *channelID)
	}
	return name
}

func parseInt64CSV(raw string) []int64 {
	parts := strings.Split(strings.TrimSpace(raw), ",")
	result := make([]int64, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		value, err := strconv.ParseInt(part, 10, 64)
		if err != nil {
			continue
		}
		result = append(result, value)
	}
	return result
}

func boolValue(value *bool) bool {
	if value == nil {
		return false
	}
	return *value
}

func interfaceToInt64Slice(value any) []int64 {
	switch typed := value.(type) {
	case []int64:
		return typed
	case []int:
		res := make([]int64, len(typed))
		for idx, item := range typed {
			res[idx] = int64(item)
		}
		return res
	case []any:
		result := make([]int64, 0, len(typed))
		for _, item := range typed {
			if parsed, ok := parseAnyToInt64(item); ok {
				result = append(result, parsed)
			}
		}
		return result
	default:
		if parsed, ok := parseAnyToInt64(value); ok {
			return []int64{parsed}
		}
	}
	return nil
}

func parseAnyToInt64(value any) (int64, bool) {
	switch typed := value.(type) {
	case int64:
		return typed, true
	case int:
		return int64(typed), true
	case float64:
		return int64(typed), true
	case string:
		text := strings.TrimSpace(typed)
		if text == "" {
			return 0, false
		}
		if parsed, err := strconv.ParseInt(text, 10, 64); err == nil {
			return parsed, true
		}
	}
	return 0, false
}

func camelToSnake(input string) string {
	if input == "" {
		return input
	}
	var builder strings.Builder
	for idx, r := range input {
		if r >= 'A' && r <= 'Z' {
			if idx > 0 {
				builder.WriteRune('_')
			}
			builder.WriteRune(r + ('a' - 'A'))
			continue
		}
		builder.WriteRune(r)
	}
	return builder.String()
}

func snakeToCamel(input string) string {
	if input == "" {
		return input
	}
	parts := strings.Split(input, "_")
	for idx := 1; idx < len(parts); idx++ {
		if parts[idx] == "" {
			continue
		}
		runes := []rune(parts[idx])
		runes[0] = unicode.ToUpper(runes[0])
		parts[idx] = string(runes)
	}
	return strings.Join(parts, "")
}

func normalizeDBValue(value any) any {
	switch typed := value.(type) {
	case nil:
		return nil
	case []byte:
		return string(typed)
	case time.Time:
		return typed
	default:
		return typed
	}
}

func normalizeUpdateValue(value any) any {
	switch typed := value.(type) {
	case map[string]any, []any:
		blob, err := json.Marshal(typed)
		if err != nil {
			return fmt.Sprintf("%v", typed)
		}
		return string(blob)
	default:
		return typed
	}
}

func parseNullableTime(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	t := value.Time
	return &t
}
