package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"go-migration-platform/pkg/institutionmenu"
	"go-migration-platform/services/platform/internal/model"
)

type Repository struct {
	db *sql.DB
}

type rawMenu struct {
	ID        int64
	Name      string
	Code      string
	PID       int64
	Sort      int64
	Weight    int64
	MenuType  int64
	GroupCode string
	Introduce string
}

var defaultInstitutionVersionModules = []struct {
	Name  string
	Price float64
}{
	{Name: "体验版", Price: 0},
	{Name: "基础版", Price: 999},
	{Name: "高级版", Price: 2999},
	{Name: "旗舰版", Price: 9999},
}

func institutionOpenTypeModuleName(openType int) string {
	switch openType {
	case 1:
		return "体验版"
	case 2:
		return "基础版"
	case 3:
		return "高级版"
	case 4:
		return "旗舰版"
	default:
		return "基础版"
	}
}

func institutionModuleNameOpenType(moduleName string) int {
	switch strings.TrimSpace(moduleName) {
	case "体验版":
		return 1
	case "基础版":
		return 2
	case "高级版":
		return 3
	case "旗舰版":
		return 4
	default:
		return 0
	}
}

func sanitizeInstitutionMenuScope(selectedMenuIDs, moduleMenuIDs []int64) ([]int64, error) {
	if len(selectedMenuIDs) == 0 {
		return nil, nil
	}

	moduleSet := make(map[int64]struct{}, len(moduleMenuIDs))
	for _, id := range moduleMenuIDs {
		if id > 0 {
			moduleSet[id] = struct{}{}
		}
	}

	result := make([]int64, 0, len(selectedMenuIDs))
	seen := make(map[int64]struct{}, len(selectedMenuIDs))
	for _, id := range selectedMenuIDs {
		if id <= 0 {
			continue
		}
		if _, exists := moduleSet[id]; !exists {
			return nil, fmt.Errorf("存在超出版本范围的权限菜单")
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("请至少选择一个机构权限")
	}
	return result, nil
}

func isInstitutionBaseMenuCode(menuCode string) bool {
	normalized := institutionmenu.NormalizeCode(menuCode)
	return strings.HasPrefix(normalized, "grp:") || strings.HasPrefix(normalized, "page:")
}

func institutionStatusExpr(alias string) string {
	return "CASE " +
		"WHEN IFNULL(" + alias + ".status, 0) = 4 THEN 4 " +
		"WHEN " + alias + ".expire_end_time IS NOT NULL AND " + alias + ".expire_end_time < NOW() THEN 4 " +
		"WHEN IFNULL(" + alias + ".status, CASE WHEN IFNULL(" + alias + ".enabled, 0) = 1 THEN 1 ELSE 2 END) = 1 AND IFNULL(" + alias + ".enabled, 0) = 1 THEN 1 " +
		"ELSE 2 END"
}

func buildInstitutionWhereClause(keyword, mobile, registerTimeBegin, registerTimeEnd string, enabled *bool, status, openType, provinceCode, cityCode, regionCode *int) (string, []any) {
	filters := []string{"oi.del_flag = 0"}
	args := make([]any, 0, 8)
	statusExpr := institutionStatusExpr("oi")

	if trimmed := strings.TrimSpace(keyword); trimmed != "" {
		like := "%" + trimmed + "%"
		filters = append(filters, "(CAST(oi.id AS CHAR) LIKE ? OR oi.organ_name LIKE ?)")
		args = append(args, like, like)
	}

	if trimmed := strings.TrimSpace(mobile); trimmed != "" {
		filters = append(filters, "oi.mobile LIKE ?")
		args = append(args, "%"+trimmed+"%")
	}

	if trimmed := strings.TrimSpace(registerTimeBegin); trimmed != "" {
		filters = append(filters, "oi.create_time >= ?")
		args = append(args, trimmed+" 00:00:00")
	}

	if trimmed := strings.TrimSpace(registerTimeEnd); trimmed != "" {
		filters = append(filters, "oi.create_time <= ?")
		args = append(args, trimmed+" 23:59:59")
	}

	if enabled != nil {
		filters = append(filters, "IFNULL(oi.enabled, 0) = ?")
		if *enabled {
			args = append(args, 1)
		} else {
			args = append(args, 0)
		}
	}

	if status != nil && *status > 0 {
		filters = append(filters, statusExpr+" = ?")
		args = append(args, *status)
	}

	if openType != nil && *openType > 0 {
		filters = append(filters, "IFNULL(oi.open_type, 2) = ?")
		args = append(args, *openType)
	}

	if provinceCode != nil && *provinceCode > 0 {
		filters = append(filters, "IFNULL(oi.province_code, 0) = ?")
		args = append(args, *provinceCode)
	}

	if cityCode != nil && *cityCode > 0 {
		filters = append(filters, "IFNULL(oi.city_code, 0) = ?")
		args = append(args, *cityCode)
	}

	if regionCode != nil && *regionCode > 0 {
		filters = append(filters, "IFNULL(oi.region_code, 0) = ?")
		args = append(args, *regionCode)
	}

	return strings.Join(filters, " AND "), args
}

func New(db *sql.DB) (*Repository, error) {
	repo := &Repository{db: db}
	if err := repo.migrateLegacyInstitutionMenuCodes(context.Background()); err != nil {
		return nil, err
	}
	if err := repo.ensureInstitutionSchema(context.Background()); err != nil {
		return nil, err
	}
	if err := repo.ensureInstitutionProfileSchema(context.Background()); err != nil {
		return nil, err
	}
	if err := repo.ensureInstitutionRenewalSchema(context.Background()); err != nil {
		return nil, err
	}
	if err := repo.ensureInstitutionVersionChangeSchema(context.Background()); err != nil {
		return nil, err
	}
	if err := repo.ensureVersionModuleSchema(context.Background()); err != nil {
		return nil, err
	}
	if err := repo.ensureInstitutionMenuCatalog(context.Background()); err != nil {
		return nil, err
	}
	return repo, nil
}

func (repo *Repository) ensureInstitutionSchema(ctx context.Context) error {
	if err := repo.ensureColumnType(ctx, "org_institution", "expire_start_time", "datetime", `
		ALTER TABLE org_institution
		MODIFY COLUMN expire_start_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '账号有效期起始时间'
	`); err != nil {
		return err
	}

	if err := repo.ensureColumnType(ctx, "org_institution", "expire_end_time", "datetime", `
		ALTER TABLE org_institution
		MODIFY COLUMN expire_end_time DATETIME NULL DEFAULT NULL COMMENT '账号有效期结束'
	`); err != nil {
		return err
	}

	if err := repo.ensureColumnExists(ctx, "org_institution", "open_type", `
		ALTER TABLE org_institution
		ADD COLUMN open_type TINYINT NOT NULL DEFAULT 2 COMMENT '开通版本：1体验版 2基础版 3高级版 4旗舰版'
		AFTER login_name
	`); err != nil {
		return err
	}

	if err := repo.ensureColumnExists(ctx, "org_institution", "open_duration", `
		ALTER TABLE org_institution
		ADD COLUMN open_duration VARCHAR(16) NOT NULL DEFAULT '1y' COMMENT '开通时长编码'
		AFTER open_type
	`); err != nil {
		return err
	}

	_, err := repo.db.ExecContext(ctx, `
		UPDATE org_institution
		SET open_type = CASE
		        WHEN IFNULL(open_type, 0) IN (1, 2, 3, 4) THEN open_type
		        WHEN expire_end_time IS NOT NULL
		          AND TIMESTAMPDIFF(DAY, COALESCE(expire_start_time, create_time, NOW()), expire_end_time) <= 7 THEN 1
		        ELSE 2
		    END,
		    open_duration = CASE
		        WHEN NULLIF(TRIM(IFNULL(open_duration, '')), '') IS NOT NULL THEN open_duration
		        ELSE CASE
		            WHEN expire_end_time IS NULL THEN '99y'
		            WHEN TIMESTAMPDIFF(DAY, COALESCE(expire_start_time, create_time, NOW()), expire_end_time) <= 4 THEN '3d'
		            WHEN TIMESTAMPDIFF(DAY, COALESCE(expire_start_time, create_time, NOW()), expire_end_time) <= 6 THEN '5d'
		            WHEN TIMESTAMPDIFF(DAY, COALESCE(expire_start_time, create_time, NOW()), expire_end_time) <= 8 THEN '7d'
		            WHEN TIMESTAMPDIFF(DAY, COALESCE(expire_start_time, create_time, NOW()), expire_end_time) <= 548 THEN '1y'
		            WHEN TIMESTAMPDIFF(DAY, COALESCE(expire_start_time, create_time, NOW()), expire_end_time) <= 913 THEN '2y'
		            WHEN TIMESTAMPDIFF(DAY, COALESCE(expire_start_time, create_time, NOW()), expire_end_time) <= 1461 THEN '3y'
		            WHEN TIMESTAMPDIFF(DAY, COALESCE(expire_start_time, create_time, NOW()), expire_end_time) <= 3650 THEN '5y'
		            ELSE '99y'
		        END
		    END
		WHERE del_flag = 0
	`)
	return err
}

func (repo *Repository) ensureColumnExists(ctx context.Context, tableName, columnName, ddl string) error {
	var count int
	if err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE()
		  AND TABLE_NAME = ?
		  AND COLUMN_NAME = ?
	`, tableName, columnName).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	_, err := repo.db.ExecContext(ctx, ddl)
	return err
}

func (repo *Repository) ensureColumnType(ctx context.Context, tableName, columnName, expectedType, ddl string) error {
	var dataType string
	if err := repo.db.QueryRowContext(ctx, `
		SELECT DATA_TYPE
		FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE()
		  AND TABLE_NAME = ?
		  AND COLUMN_NAME = ?
		LIMIT 1
	`, tableName, columnName).Scan(&dataType); err != nil {
		return err
	}

	if strings.EqualFold(strings.TrimSpace(dataType), strings.TrimSpace(expectedType)) {
		return nil
	}

	_, err := repo.db.ExecContext(ctx, ddl)
	return err
}

func (repo *Repository) ensureInstitutionProfileSchema(ctx context.Context) error {
	if _, err := repo.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS org_institution_profile (
			id BIGINT NOT NULL AUTO_INCREMENT,
			institution_id BIGINT NOT NULL,
			description TEXT DEFAULT NULL,
			business_time VARCHAR(255) DEFAULT NULL,
			video VARCHAR(2000) DEFAULT NULL,
			gallery_images JSON DEFAULT NULL,
			create_id BIGINT DEFAULT NULL,
			create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
			update_id BIGINT DEFAULT NULL,
			update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			del_flag TINYINT(1) DEFAULT 0,
			PRIMARY KEY (id),
			UNIQUE KEY uk_org_institution_profile_inst (institution_id),
			KEY idx_org_institution_profile_inst_del (institution_id, del_flag)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
	`); err != nil {
		return err
	}

	_, err := repo.db.ExecContext(ctx, `
		INSERT INTO org_institution_profile (
			institution_id, description, business_time, video, gallery_images,
			create_id, create_time, update_id, update_time, del_flag
		)
		SELECT oi.id,
		       NULLIF(TRIM(IFNULL(oi.description, '')), ''),
		       NULLIF(TRIM(IFNULL(oi.business_time, '')), ''),
		       NULLIF(TRIM(IFNULL(oi.video, '')), ''),
		       oi.inst_images,
		       oi.create_id,
		       COALESCE(oi.create_time, NOW()),
		       oi.update_id,
		       COALESCE(oi.update_time, NOW()),
		       0
		FROM org_institution oi
		LEFT JOIN org_institution_profile oip ON oip.institution_id = oi.id AND oip.del_flag = 0
		WHERE oi.del_flag = 0
		  AND oip.id IS NULL
		  AND (
			NULLIF(TRIM(IFNULL(oi.description, '')), '') IS NOT NULL
			OR NULLIF(TRIM(IFNULL(oi.business_time, '')), '') IS NOT NULL
			OR NULLIF(TRIM(IFNULL(oi.video, '')), '') IS NOT NULL
			OR oi.inst_images IS NOT NULL
		  )
	`)
	return err
}

func (repo *Repository) ensureInstitutionRenewalSchema(ctx context.Context) error {
	_, err := repo.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS org_institution_renewal_record (
			id BIGINT NOT NULL AUTO_INCREMENT,
			institution_id BIGINT NOT NULL,
			before_open_type TINYINT NOT NULL DEFAULT 2,
			before_open_duration VARCHAR(16) NOT NULL DEFAULT '',
			before_expire_end_time DATETIME DEFAULT NULL,
			after_open_type TINYINT NOT NULL DEFAULT 2,
			renew_duration VARCHAR(16) NOT NULL DEFAULT '',
			renew_start_time DATETIME NOT NULL,
			after_expire_end_time DATETIME NOT NULL,
			operator_id BIGINT DEFAULT NULL,
			create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			del_flag TINYINT(1) DEFAULT 0,
			PRIMARY KEY (id),
			KEY idx_org_institution_renewal_record_inst_del_time (institution_id, del_flag, create_time)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
	`)
	return err
}

func (repo *Repository) ensureInstitutionVersionChangeSchema(ctx context.Context) error {
	_, err := repo.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS org_institution_version_change_record (
			id BIGINT NOT NULL AUTO_INCREMENT,
			institution_id BIGINT NOT NULL,
			before_open_type TINYINT NOT NULL DEFAULT 2,
			before_module_id BIGINT DEFAULT NULL,
			before_version_name VARCHAR(64) NOT NULL DEFAULT '',
			after_open_type TINYINT NOT NULL DEFAULT 2,
			after_module_id BIGINT DEFAULT NULL,
			after_version_name VARCHAR(64) NOT NULL DEFAULT '',
			operator_id BIGINT DEFAULT NULL,
			create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			del_flag TINYINT(1) DEFAULT 0,
			PRIMARY KEY (id),
			KEY idx_org_institution_version_change_record_inst_del_time (institution_id, del_flag, create_time)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
	`)
	return err
}

func (repo *Repository) ensureVersionModuleSchema(ctx context.Context) error {
	if _, err := repo.db.ExecContext(ctx, `
		UPDATE sys_module
		SET name = '体验版'
		WHERE del_flag = 0 AND type = 1 AND name = '试用版'
	`); err != nil {
		return err
	}

	for _, item := range defaultInstitutionVersionModules {
		var count int
		if err := repo.db.QueryRowContext(ctx, `
			SELECT COUNT(*)
			FROM sys_module
			WHERE del_flag = 0 AND type = 1 AND name = ?
		`, item.Name).Scan(&count); err != nil {
			return err
		}
		if count > 0 {
			continue
		}
		if _, err := repo.db.ExecContext(ctx, `
			INSERT INTO sys_module (uuid, version, name, type, price, create_time, update_time, del_flag, remark)
			VALUES (?, 0, ?, 1, ?, NOW(), NOW(), 0, '')
		`, uuid.NewString(), item.Name, item.Price); err != nil {
			return err
		}
	}

	return nil
}

func trimStringSlice(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		result = append(result, trimmed)
	}
	return result
}

func marshalStringSlice(values []string) string {
	cleanValues := trimStringSlice(values)
	if len(cleanValues) == 0 {
		return "[]"
	}
	payload, err := json.Marshal(cleanValues)
	if err != nil {
		return "[]"
	}
	return string(payload)
}

func unmarshalStringSlice(raw string) []string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" || trimmed == "null" {
		return nil
	}

	var direct []string
	if err := json.Unmarshal([]byte(trimmed), &direct); err == nil {
		return trimStringSlice(direct)
	}

	var generic []map[string]any
	if err := json.Unmarshal([]byte(trimmed), &generic); err == nil {
		result := make([]string, 0, len(generic))
		for _, item := range generic {
			for _, key := range []string{"url", "fileUrl", "src", "value"} {
				if rawValue, ok := item[key].(string); ok && strings.TrimSpace(rawValue) != "" {
					result = append(result, strings.TrimSpace(rawValue))
					break
				}
			}
		}
		return trimStringSlice(result)
	}

	return nil
}

func normalizedInstitutionProfile(profile *model.InstitutionProfile) model.InstitutionProfile {
	if profile == nil {
		return model.InstitutionProfile{}
	}

	return model.InstitutionProfile{
		Description:   strings.TrimSpace(profile.Description),
		BusinessTime:  strings.TrimSpace(profile.BusinessTime),
		Video:         strings.TrimSpace(profile.Video),
		GalleryImages: trimStringSlice(profile.GalleryImages),
	}
}

func (repo *Repository) PageDicts(ctx context.Context, current, size int, keyword string) (model.PageResult[model.Dict], error) {
	if current <= 0 {
		current = 1
	}
	if size <= 0 {
		size = 10
	}
	offset := (current - 1) * size

	filters := []string{"del_flag = 0"}
	args := make([]any, 0, 2)
	if strings.TrimSpace(keyword) != "" {
		filters = append(filters, "(dict_name LIKE ? OR dict_code LIKE ?)")
		args = append(args, "%"+strings.TrimSpace(keyword)+"%", "%"+strings.TrimSpace(keyword)+"%")
	}
	whereClause := strings.Join(filters, " AND ")

	var total int
	if err := repo.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM sys_dict WHERE "+whereClause, args...).Scan(&total); err != nil {
		return model.PageResult[model.Dict]{}, err
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT id, dict_name, dict_code, IFNULL(is_enable, 0), IFNULL(remark, '')
		FROM sys_dict
		WHERE `+whereClause+`
		ORDER BY id DESC
		LIMIT ? OFFSET ?`, append(args, size, offset)...)
	if err != nil {
		return model.PageResult[model.Dict]{}, err
	}
	defer rows.Close()

	items := make([]model.Dict, 0, size)
	for rows.Next() {
		var item model.Dict
		if err := rows.Scan(&item.ID, &item.DictName, &item.DictCode, &item.IsEnable, &item.Remark); err != nil {
			return model.PageResult[model.Dict]{}, err
		}
		items = append(items, item)
	}

	return model.PageResult[model.Dict]{
		Items:   items,
		Total:   total,
		Current: current,
		Size:    size,
	}, rows.Err()
}

func (repo *Repository) ListDictValuesByCode(ctx context.Context, code string) ([]model.DictValue, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT v.id, v.dict_id, v.dict_label, v.dict_value, IFNULL(v.sort, 0), IFNULL(v.is_enable, 0)
		FROM sys_dict_value v
		JOIN sys_dict d ON v.dict_id = d.id
		WHERE d.del_flag = 0 AND v.del_flag = 0 AND d.dict_code = ?
		ORDER BY v.sort ASC, v.id ASC
	`, strings.TrimSpace(code))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.DictValue, 0, 16)
	for rows.Next() {
		var item model.DictValue
		if err := rows.Scan(&item.ID, &item.DictID, &item.DictLabel, &item.DictValue, &item.Sort, &item.IsEnable); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (repo *Repository) PageNotices(ctx context.Context, query model.NoticeQuery) (model.PageResult[model.Notice], error) {
	current := query.Current
	size := query.Size
	if current <= 0 {
		current = 1
	}
	if size <= 0 {
		size = 10
	}
	offset := (current - 1) * size

	filters := []string{"del_flag = 0", "disable_id = ?"}
	args := make([]any, 0, 4)
	args = append(args, query.DisableID)
	if strings.TrimSpace(query.Title) != "" {
		filters = append(filters, "title LIKE ?")
		args = append(args, "%"+strings.TrimSpace(query.Title)+"%")
	}
	if strings.TrimSpace(query.StartTime) != "" {
		filters = append(filters, "create_time >= ?")
		args = append(args, strings.TrimSpace(query.StartTime))
	}
	if strings.TrimSpace(query.EndTime) != "" {
		filters = append(filters, "create_time <= ?")
		args = append(args, strings.TrimSpace(query.EndTime))
	}
	whereClause := strings.Join(filters, " AND ")

	var total int
	if err := repo.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM sys_notice_info WHERE "+whereClause, args...).Scan(&total); err != nil {
		return model.PageResult[model.Notice]{}, err
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT id, IFNULL(title, ''), IFNULL(content, ''), IFNULL(disable_id, -1), IFNULL(compel, 0), create_time
		FROM sys_notice_info
		WHERE `+whereClause+`
		ORDER BY id DESC
		LIMIT ? OFFSET ?`, append(args, size, offset)...)
	if err != nil {
		return model.PageResult[model.Notice]{}, err
	}
	defer rows.Close()

	items := make([]model.Notice, 0, size)
	for rows.Next() {
		var item model.Notice
		if err := rows.Scan(&item.ID, &item.Title, &item.Content, &item.DisableID, &item.Compel, &item.CreateTime); err != nil {
			return model.PageResult[model.Notice]{}, err
		}
		items = append(items, item)
	}

	return model.PageResult[model.Notice]{
		Items:   items,
		Total:   total,
		Current: current,
		Size:    size,
	}, rows.Err()
}

func (repo *Repository) PageModules(ctx context.Context, current, size int, name string, moduleType int) (model.PageResult[model.Module], error) {
	if current <= 0 {
		current = 1
	}
	if size <= 0 {
		size = 10
	}
	offset := (current - 1) * size

	filters := []string{"del_flag = 0"}
	args := make([]any, 0, 2)
	if strings.TrimSpace(name) != "" {
		filters = append(filters, "name LIKE ?")
		args = append(args, "%"+strings.TrimSpace(name)+"%")
	}
	if moduleType > 0 {
		filters = append(filters, "type = ?")
		args = append(args, moduleType)
	}
	whereClause := strings.Join(filters, " AND ")

	var total int
	if err := repo.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM sys_module WHERE "+whereClause, args...).Scan(&total); err != nil {
		return model.PageResult[model.Module]{}, err
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT m.id,
		       IFNULL(m.name, ''),
		       IFNULL(m.type, 0),
		       IFNULL(m.price, 0),
		       IFNULL(m.remark, ''),
		       COUNT(DISTINCT CASE WHEN smm.del_flag = 0 THEN smm.menu_id END) AS menu_count,
		       COUNT(DISTINCT CASE WHEN om.del_flag = 0 THEN om.org_id END) AS org_count,
		       IFNULL(DATE_FORMAT(m.create_time, '%Y-%m-%d %H:%i:%s'), ''),
		       IFNULL(DATE_FORMAT(m.update_time, '%Y-%m-%d %H:%i:%s'), '')
		FROM sys_module m
		LEFT JOIN sys_module_menu smm ON smm.module_id = m.id
		LEFT JOIN org_module om ON om.module_id = m.id
		WHERE `+strings.ReplaceAll(whereClause, "del_flag", "m.del_flag")+`
		GROUP BY m.id, m.name, m.type, m.price, m.remark, m.create_time, m.update_time
		ORDER BY m.id DESC
		LIMIT ? OFFSET ?`, append(args, size, offset)...)
	if err != nil {
		return model.PageResult[model.Module]{}, err
	}
	defer rows.Close()

	items := make([]model.Module, 0, size)
	for rows.Next() {
		var item model.Module
		if err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.Type,
			&item.Price,
			&item.Remark,
			&item.MenuCount,
			&item.OrgCount,
			&item.CreateTime,
			&item.UpdateTime,
		); err != nil {
			return model.PageResult[model.Module]{}, err
		}
		items = append(items, item)
	}

	return model.PageResult[model.Module]{
		Items:   items,
		Total:   total,
		Current: current,
		Size:    size,
	}, rows.Err()
}

func normalizeInstitutionOpenType(value *int) int {
	if value == nil {
		return 2
	}

	switch *value {
	case 1, 2, 3, 4:
		return *value
	default:
		return 2
	}
}

func institutionStoredOpenType(value sql.NullInt64) int {
	if !value.Valid {
		return 0
	}

	switch int(value.Int64) {
	case 1:
		return 1
	case 2:
		return 2
	case 3:
		return 3
	case 4:
		return 4
	default:
		return 0
	}
}

func resolveInstitutionUpdateOpenType(value *int, currentType int) int {
	if value != nil {
		return normalizeInstitutionOpenType(value)
	}
	if currentType >= 1 && currentType <= 4 {
		return currentType
	}
	return 2
}

func normalizeInstitutionOpenDuration(openType int, raw string) string {
	value := strings.TrimSpace(raw)
	switch openType {
	case 1:
		switch value {
		case "3d", "5d", "7d":
			return value
		default:
			return "7d"
		}
	default:
		switch value {
		case "1y", "2y", "3y", "5y", "99y":
			return value
		default:
			return "1y"
		}
	}
}

func buildInstitutionExpireEndTime(start time.Time, duration string) time.Time {
	switch strings.TrimSpace(duration) {
	case "3d":
		return start.AddDate(0, 0, 3)
	case "5d":
		return start.AddDate(0, 0, 5)
	case "7d":
		return start.AddDate(0, 0, 7)
	case "2y":
		return start.AddDate(2, 0, 0)
	case "3y":
		return start.AddDate(3, 0, 0)
	case "5y":
		return start.AddDate(5, 0, 0)
	case "99y":
		return start.AddDate(99, 0, 0)
	default:
		return start.AddDate(1, 0, 0)
	}
}

func buildInstitutionRenewalWindow(currentOpenType, nextOpenType int, currentExpireEnd sql.NullTime, duration string) (sql.NullTime, sql.NullTime, error) {
	if currentOpenType >= 2 && nextOpenType == 1 {
		return sql.NullTime{}, sql.NullTime{}, fmt.Errorf("已开通基础版、高级版或旗舰版的机构不支持降为体验版")
	}

	now := time.Now()
	effectiveEnd := now
	if currentExpireEnd.Valid && currentExpireEnd.Time.After(now) {
		effectiveEnd = currentExpireEnd.Time
	}

	if currentOpenType != nextOpenType {
		return sql.NullTime{Time: now, Valid: true}, sql.NullTime{Time: buildInstitutionExpireEndTime(effectiveEnd, duration), Valid: true}, nil
	}

	return sql.NullTime{Time: effectiveEnd, Valid: true}, sql.NullTime{Time: buildInstitutionExpireEndTime(effectiveEnd, duration), Valid: true}, nil
}

func institutionStatusValue(enabled bool, expireEnd sql.NullTime) int {
	if !enabled {
		return 2
	}
	if expireEnd.Valid && expireEnd.Time.Before(time.Now()) {
		return 4
	}
	return 1
}

func (repo *Repository) PageInstitutions(ctx context.Context, current, size int, keyword, mobile, registerTimeBegin, registerTimeEnd string, enabled *bool, status, openType, provinceCode, cityCode, regionCode *int) (model.InstitutionPage, error) {
	if current <= 0 {
		current = 1
	}
	if size <= 0 {
		size = 10
	}
	offset := (current - 1) * size

	whereClause, args := buildInstitutionWhereClause(keyword, mobile, registerTimeBegin, registerTimeEnd, enabled, status, openType, provinceCode, cityCode, regionCode)
	statusExpr := institutionStatusExpr("oi")

	var total int
	if err := repo.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM org_institution oi WHERE "+whereClause, args...).Scan(&total); err != nil {
		return model.InstitutionPage{}, err
	}

	var summary model.InstitutionSummary
	if err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(*),
		       COALESCE(SUM(CASE WHEN `+statusExpr+` = 1 THEN 1 ELSE 0 END), 0),
		       COALESCE(SUM(CASE WHEN `+statusExpr+` <> 1 THEN 1 ELSE 0 END), 0)
		FROM org_institution oi
		WHERE `+whereClause, args...).Scan(&summary.TotalCount, &summary.EnabledCount, &summary.DisabledCount); err != nil {
		return model.InstitutionPage{}, err
	}

	listArgs := append(append([]any{}, args...), size, offset)
	rows, err := repo.db.QueryContext(ctx, `
		SELECT oi.id,
		       IFNULL(oi.organ_name, ''),
		       IFNULL(oi.organ_code, ''),
		       IFNULL(oi.login_name, ''),
		       IFNULL(oi.mobile, ''),
		       IFNULL(oi.principal, ''),
		       IFNULL(oi.province, ''),
		       IFNULL(oi.city, ''),
		       IFNULL(oi.region, ''),
		       IFNULL(oi.address, ''),
		       IFNULL(oi.logo, ''),
		       IFNULL(oi.enabled, 0),
		       `+statusExpr+`,
		       IFNULL(oi.open_type, 2),
		       IFNULL(oi.open_duration, ''),
		       IFNULL(DATE_FORMAT(oi.create_time, '%Y-%m-%d %H:%i:%s'), ''),
		       IFNULL(DATE_FORMAT(oi.expire_end_time, '%Y-%m-%d %H:%i:%s'), ''),
		       COUNT(DISTINCT CASE WHEN iu.del_flag = 0 THEN iu.id END) AS staff_count,
		       COUNT(DISTINCT CASE WHEN iu.del_flag = 0 AND IFNULL(iu.disabled, 0) = 0 THEN iu.id END) AS active_staff_count,
		       COUNT(DISTINCT CASE WHEN iu.del_flag = 0 AND IFNULL(iu.is_admin, 0) = 1 THEN iu.id END) AS admin_count
		FROM org_institution oi
		LEFT JOIN inst_user iu ON iu.inst_id = oi.id
		WHERE `+whereClause+`
		GROUP BY oi.id, oi.organ_name, oi.organ_code, oi.login_name, oi.mobile, oi.principal, oi.province, oi.city, oi.region, oi.address, oi.logo, oi.enabled, oi.status, oi.open_type, oi.open_duration, oi.create_time, oi.expire_end_time
		ORDER BY oi.id DESC
		LIMIT ? OFFSET ?`, listArgs...)
	if err != nil {
		return model.InstitutionPage{}, err
	}
	defer rows.Close()

	items := make([]model.Institution, 0, size)
	for rows.Next() {
		var item model.Institution
		if err := rows.Scan(
			&item.ID,
			&item.OrganName,
			&item.OrganCode,
			&item.LoginName,
			&item.Mobile,
			&item.Principal,
			&item.Province,
			&item.City,
			&item.Region,
			&item.Address,
			&item.Logo,
			&item.Enabled,
			&item.Status,
			&item.OpenType,
			&item.OpenDuration,
			&item.RegisterTime,
			&item.ExpireEndTime,
			&item.StaffCount,
			&item.ActiveStaffCount,
			&item.AdminCount,
		); err != nil {
			return model.InstitutionPage{}, err
		}
		items = append(items, item)
	}

	return model.InstitutionPage{
		Items:   items,
		Total:   total,
		Current: current,
		Size:    size,
		Summary: &summary,
	}, rows.Err()
}

func (repo *Repository) GetInstitutionDetail(ctx context.Context, id int64) (model.InstitutionDetail, error) {
	row := repo.db.QueryRowContext(ctx, `
		SELECT oi.id,
		       IFNULL(oi.organ_name, ''),
		       IFNULL(oi.organ_code, ''),
		       IFNULL(oi.login_name, ''),
		       IFNULL(oi.mobile, ''),
		       IFNULL(oi.principal, ''),
		       IFNULL(oi.province_code, 0),
		       IFNULL(oi.province, ''),
		       IFNULL(oi.city_code, 0),
		       IFNULL(oi.city, ''),
		       IFNULL(oi.region_code, 0),
		       IFNULL(oi.region, ''),
		       IFNULL(oi.address, ''),
		       IFNULL(oi.concat_phone, ''),
		       IFNULL(oi.fixed_phone, ''),
		       IFNULL(oi.remark, ''),
		       IFNULL(oi.logo, ''),
		       IFNULL(oi.enabled, 0),
		       `+institutionStatusExpr("oi")+`,
		       IFNULL(oi.open_type, 2),
		       IFNULL(oi.open_duration, ''),
		       IFNULL(DATE_FORMAT(oi.expire_start_time, '%Y-%m-%d %H:%i:%s'), ''),
		       IFNULL(DATE_FORMAT(oi.expire_end_time, '%Y-%m-%d %H:%i:%s'), ''),
		       IFNULL(oi.lng, 0),
		       IFNULL(oi.lat, 0),
		       IFNULL(NULLIF(oip.description, ''), IFNULL(oi.description, '')),
		       IFNULL(NULLIF(oip.business_time, ''), IFNULL(oi.business_time, '')),
		       IFNULL(NULLIF(oip.video, ''), IFNULL(oi.video, '')),
		       CASE
		           WHEN oip.gallery_images IS NOT NULL THEN CAST(oip.gallery_images AS CHAR)
		           WHEN oi.inst_images IS NOT NULL THEN CAST(oi.inst_images AS CHAR)
		           ELSE '[]'
		       END
		FROM org_institution oi
		LEFT JOIN org_institution_profile oip ON oip.institution_id = oi.id AND oip.del_flag = 0
		WHERE oi.id = ? AND oi.del_flag = 0
		LIMIT 1
	`, id)

	var detail model.InstitutionDetail
	var galleryImagesRaw string
	if err := row.Scan(
		&detail.ID,
		&detail.OrganName,
		&detail.OrganCode,
		&detail.LoginName,
		&detail.Mobile,
		&detail.Principal,
		&detail.ProvinceCode,
		&detail.Province,
		&detail.CityCode,
		&detail.City,
		&detail.RegionCode,
		&detail.Region,
		&detail.Address,
		&detail.ConcatPhone,
		&detail.FixedPhone,
		&detail.Remark,
		&detail.Logo,
		&detail.Enabled,
		&detail.Status,
		&detail.OpenType,
		&detail.OpenDuration,
		&detail.ExpireStartTime,
		&detail.ExpireEndTime,
		&detail.Lng,
		&detail.Lat,
		&detail.Profile.Description,
		&detail.Profile.BusinessTime,
		&detail.Profile.Video,
		&galleryImagesRaw,
	); err != nil {
		return model.InstitutionDetail{}, err
	}
	detail.Profile.GalleryImages = unmarshalStringSlice(galleryImagesRaw)

	return detail, nil
}

func (repo *Repository) CreateInstitution(ctx context.Context, input model.InstitutionMutation, creatorID *int64) (int64, error) {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	nextCode, err := repo.nextInstitutionCode(ctx, tx)
	if err != nil {
		return 0, err
	}

	enabled := true
	if input.Enabled != nil {
		enabled = *input.Enabled
	}
	openType := normalizeInstitutionOpenType(input.OpenType)
	openDuration := normalizeInstitutionOpenDuration(openType, input.OpenDuration)
	expireStart := time.Now()
	expireEnd := buildInstitutionExpireEndTime(expireStart, openDuration)
	statusValue := institutionStatusValue(enabled, sql.NullTime{Time: expireEnd, Valid: true})

	creatorValue := nullableInt64Value(creatorID)
	profile := normalizedInstitutionProfile(input.Profile)
	galleryImagesJSON := marshalStringSlice(profile.GalleryImages)

	result, err := tx.ExecContext(ctx, `
		INSERT INTO org_institution (
			uuid, organ_name, organ_type, mobile, organ_code, login_name,
			open_type, open_duration, expire_start_time, expire_end_time,
			province_code, province, city_code, city, region_code, region, logo, principal, address, lng, lat,
			description, business_time, video, inst_images,
			status, enabled, concat_phone, fixed_phone, version, create_id, create_time, update_id, update_time, del_flag, remark, account_num
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, ?, NOW(), ?, NOW(), 0, ?, 5)
	`,
		uuid.NewString(),
		strings.TrimSpace(input.OrganName),
		1,
		strings.TrimSpace(input.Mobile),
		nextCode,
		strings.TrimSpace(input.LoginName),
		openType,
		openDuration,
		expireStart,
		expireEnd,
		nullableInt64Value(input.ProvinceCode),
		strings.TrimSpace(input.Province),
		nullableInt64Value(input.CityCode),
		strings.TrimSpace(input.City),
		nullableInt64Value(input.RegionCode),
		strings.TrimSpace(input.Region),
		strings.TrimSpace(input.Logo),
		strings.TrimSpace(input.Principal),
		strings.TrimSpace(input.Address),
		nullableFloat64Value(input.Lng),
		nullableFloat64Value(input.Lat),
		profile.Description,
		profile.BusinessTime,
		profile.Video,
		galleryImagesJSON,
		statusValue,
		enabled,
		strings.TrimSpace(input.ConcatPhone),
		strings.TrimSpace(input.FixedPhone),
		0,
		creatorValue,
		creatorValue,
		strings.TrimSpace(input.Remark),
	)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	if input.Profile != nil {
		if err = repo.upsertInstitutionProfileTx(ctx, tx, id, profile, creatorValue, creatorValue); err != nil {
			return 0, err
		}
	}

	if moduleID, lookupErr := repo.findInstitutionVersionModuleIDTx(ctx, tx, openType); lookupErr != nil {
		return 0, lookupErr
	} else if moduleID > 0 {
		if err = repo.bindInstitutionModuleTx(
			ctx,
			tx,
			id,
			moduleID,
			sql.NullTime{Time: expireEnd, Valid: true},
			statusValue,
			creatorID,
			nil,
			true,
			false,
		); err != nil {
			return 0, err
		}
	}

	if err = tx.Commit(); err != nil {
		return 0, err
	}

	return id, nil
}

func (repo *Repository) UpdateInstitution(ctx context.Context, input model.InstitutionMutation, updaterID *int64) error {
	if input.ID == nil {
		return fmt.Errorf("id is required")
	}

	enabled := true
	if input.Enabled != nil {
		enabled = *input.Enabled
	}

	updaterValue := nullableInt64Value(updaterID)
	profile := normalizedInstitutionProfile(input.Profile)
	galleryImagesJSON := marshalStringSlice(profile.GalleryImages)

	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var currentOpenType sql.NullInt64
	var currentOpenDuration sql.NullString
	var currentExpireStart sql.NullTime
	var currentExpireEnd sql.NullTime
	if err = tx.QueryRowContext(ctx, `
		SELECT IFNULL(open_type, 2),
		       IFNULL(open_duration, ''),
		       expire_start_time,
		       expire_end_time
		FROM org_institution
		WHERE id = ? AND del_flag = 0
		LIMIT 1
	`, *input.ID).Scan(&currentOpenType, &currentOpenDuration, &currentExpireStart, &currentExpireEnd); err != nil {
		return err
	}

	currentType := institutionStoredOpenType(currentOpenType)
	if currentType == 0 {
		currentType = 2
	}
	currentDuration := normalizeInstitutionOpenDuration(currentType, currentOpenDuration.String)
	expireStart := currentExpireStart
	expireEnd := currentExpireEnd
	if !expireStart.Valid || !expireEnd.Valid {
		nextExpireStart := time.Now()
		expireStart = sql.NullTime{Time: nextExpireStart, Valid: true}
		expireEnd = sql.NullTime{Time: buildInstitutionExpireEndTime(nextExpireStart, currentDuration), Valid: true}
	}
	statusValue := institutionStatusValue(enabled, expireEnd)

	_, err = tx.ExecContext(ctx, `
		UPDATE org_institution
		SET organ_name = ?,
		    login_name = ?,
		    mobile = ?,
		    open_type = ?,
		    open_duration = ?,
		    expire_start_time = ?,
		    expire_end_time = ?,
		    principal = ?,
		    province_code = ?,
		    province = ?,
		    city_code = ?,
		    city = ?,
		    region_code = ?,
		    region = ?,
		    address = ?,
		    concat_phone = ?,
		    fixed_phone = ?,
		    remark = ?,
		    logo = ?,
		    lng = COALESCE(?, lng),
		    lat = COALESCE(?, lat),
		    description = ?,
		    business_time = ?,
		    video = ?,
		    inst_images = ?,
		    enabled = ?,
		    status = ?,
		    update_id = ?,
		    update_time = NOW()
		WHERE id = ? AND del_flag = 0
	`,
		strings.TrimSpace(input.OrganName),
		strings.TrimSpace(input.LoginName),
		strings.TrimSpace(input.Mobile),
		currentType,
		currentDuration,
		nullableTimeValue(expireStart),
		nullableTimeValue(expireEnd),
		strings.TrimSpace(input.Principal),
		nullableInt64Value(input.ProvinceCode),
		strings.TrimSpace(input.Province),
		nullableInt64Value(input.CityCode),
		strings.TrimSpace(input.City),
		nullableInt64Value(input.RegionCode),
		strings.TrimSpace(input.Region),
		strings.TrimSpace(input.Address),
		strings.TrimSpace(input.ConcatPhone),
		strings.TrimSpace(input.FixedPhone),
		strings.TrimSpace(input.Remark),
		strings.TrimSpace(input.Logo),
		nullableFloat64Value(input.Lng),
		nullableFloat64Value(input.Lat),
		profile.Description,
		profile.BusinessTime,
		profile.Video,
		galleryImagesJSON,
		enabled,
		statusValue,
		updaterValue,
		*input.ID,
	)
	if err != nil {
		return err
	}

	if input.Profile != nil {
		if err = repo.upsertInstitutionProfileTx(ctx, tx, *input.ID, profile, nil, updaterValue); err != nil {
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (repo *Repository) UpdateInstitutionStatus(ctx context.Context, id int64, enabled bool, updaterID *int64) error {
	updaterValue := nullableInt64Value(updaterID)
	var expireEnd sql.NullTime
	if err := repo.db.QueryRowContext(ctx, `
		SELECT expire_end_time
		FROM org_institution
		WHERE id = ? AND del_flag = 0
		LIMIT 1
	`, id).Scan(&expireEnd); err != nil {
		return err
	}
	statusValue := institutionStatusValue(enabled, expireEnd)

	_, err := repo.db.ExecContext(ctx, `
		UPDATE org_institution
		SET enabled = ?,
		    status = ?,
		    update_id = ?,
		    update_time = NOW()
		WHERE id = ? AND del_flag = 0
	`, enabled, statusValue, updaterValue, id)
	if err != nil {
		return err
	}

	_, err = repo.db.ExecContext(ctx, `
		UPDATE org_module
		SET status = ?,
		    update_id = ?,
		    update_time = NOW()
		WHERE org_id = ? AND del_flag = 0
	`, statusValue, updaterValue, id)
	return err
}

func (repo *Repository) ListInstitutionRenewalRecords(ctx context.Context, institutionID int64) ([]model.InstitutionRenewalRecord, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT id,
		       institution_id,
		       IFNULL(before_open_type, 2),
		       IFNULL(before_open_duration, ''),
		       IFNULL(DATE_FORMAT(before_expire_end_time, '%Y-%m-%d %H:%i:%s'), ''),
		       IFNULL(after_open_type, 2),
		       IFNULL(renew_duration, ''),
		       IFNULL(DATE_FORMAT(renew_start_time, '%Y-%m-%d %H:%i:%s'), ''),
		       IFNULL(DATE_FORMAT(after_expire_end_time, '%Y-%m-%d %H:%i:%s'), ''),
		       IFNULL(operator_id, 0),
		       IFNULL(DATE_FORMAT(create_time, '%Y-%m-%d %H:%i:%s'), '')
		FROM org_institution_renewal_record
		WHERE institution_id = ? AND del_flag = 0
		ORDER BY id DESC
	`, institutionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.InstitutionRenewalRecord, 0, 16)
	for rows.Next() {
		var item model.InstitutionRenewalRecord
		if err := rows.Scan(
			&item.ID,
			&item.InstitutionID,
			&item.BeforeOpenType,
			&item.BeforeOpenDuration,
			&item.BeforeExpireEndTime,
			&item.AfterOpenType,
			&item.RenewDuration,
			&item.RenewStartTime,
			&item.AfterExpireEndTime,
			&item.OperatorID,
			&item.CreateTime,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (repo *Repository) ListInstitutionVersionChangeRecords(ctx context.Context, institutionID int64) ([]model.InstitutionVersionChangeRecord, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT r.id,
		       r.institution_id,
		       IFNULL(r.before_open_type, 2),
		       IFNULL(r.before_module_id, 0),
		       IFNULL(r.before_version_name, ''),
		       IFNULL(r.after_open_type, 2),
		       IFNULL(r.after_module_id, 0),
		       IFNULL(r.after_version_name, ''),
		       IFNULL(r.operator_id, 0),
		       COALESCE(NULLIF(TRIM(u.nick_name), ''), NULLIF(TRIM(u.username), ''), ''),
		       IFNULL(DATE_FORMAT(r.create_time, '%Y-%m-%d %H:%i:%s'), '')
		FROM org_institution_version_change_record r
		LEFT JOIN sso_user u ON u.id = r.operator_id AND u.del_flag = 0
		WHERE r.institution_id = ? AND r.del_flag = 0
		ORDER BY r.id DESC
	`, institutionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.InstitutionVersionChangeRecord, 0, 16)
	for rows.Next() {
		var item model.InstitutionVersionChangeRecord
		if err := rows.Scan(
			&item.ID,
			&item.InstitutionID,
			&item.BeforeOpenType,
			&item.BeforeModuleID,
			&item.BeforeVersionName,
			&item.AfterOpenType,
			&item.AfterModuleID,
			&item.AfterVersionName,
			&item.OperatorID,
			&item.OperatorName,
			&item.CreateTime,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (repo *Repository) RenewInstitution(ctx context.Context, input model.InstitutionRenewalMutation, operatorID *int64) (model.InstitutionRenewalResult, error) {
	if input.InstitutionID == nil {
		return model.InstitutionRenewalResult{}, fmt.Errorf("institutionId is required")
	}

	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return model.InstitutionRenewalResult{}, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var currentOpenType sql.NullInt64
	var currentOpenDuration sql.NullString
	var currentExpireEnd sql.NullTime
	var currentEnabled bool
	if err = tx.QueryRowContext(ctx, `
		SELECT IFNULL(open_type, 2),
		       IFNULL(open_duration, ''),
		       expire_end_time,
		       IFNULL(enabled, 0)
		FROM org_institution
		WHERE id = ? AND del_flag = 0
		LIMIT 1
		FOR UPDATE
	`, *input.InstitutionID).Scan(&currentOpenType, &currentOpenDuration, &currentExpireEnd, &currentEnabled); err != nil {
		return model.InstitutionRenewalResult{}, err
	}

	beforeOpenType := institutionStoredOpenType(currentOpenType)
	if beforeOpenType == 0 {
		beforeOpenType = 2
	}
	beforeOpenDuration := normalizeInstitutionOpenDuration(beforeOpenType, currentOpenDuration.String)
	nextOpenType := normalizeInstitutionOpenType(input.OpenType)
	renewDuration := normalizeInstitutionOpenDuration(nextOpenType, input.OpenDuration)

	renewStart, renewEnd, err := buildInstitutionRenewalWindow(beforeOpenType, nextOpenType, currentExpireEnd, renewDuration)
	if err != nil {
		return model.InstitutionRenewalResult{}, err
	}

	operatorValue := nullableInt64Value(operatorID)
	statusValue := institutionStatusValue(currentEnabled, renewEnd)

	if _, err = tx.ExecContext(ctx, `
		UPDATE org_institution
		SET open_type = ?,
		    open_duration = ?,
		    expire_start_time = ?,
		    expire_end_time = ?,
		    status = ?,
		    update_id = ?,
		    update_time = NOW()
		WHERE id = ? AND del_flag = 0
	`,
		nextOpenType,
		renewDuration,
		nullableTimeValue(renewStart),
		nullableTimeValue(renewEnd),
		statusValue,
		operatorValue,
		*input.InstitutionID,
	); err != nil {
		return model.InstitutionRenewalResult{}, err
	}

	if _, err = tx.ExecContext(ctx, `
		INSERT INTO org_institution_renewal_record (
			institution_id,
			before_open_type,
			before_open_duration,
			before_expire_end_time,
			after_open_type,
			renew_duration,
			renew_start_time,
			after_expire_end_time,
			operator_id,
			create_time,
			update_time,
			del_flag
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW(), 0)
	`,
		*input.InstitutionID,
		beforeOpenType,
		beforeOpenDuration,
		nullableTimeValue(currentExpireEnd),
		nextOpenType,
		renewDuration,
		nullableTimeValue(renewStart),
		nullableTimeValue(renewEnd),
		operatorValue,
	); err != nil {
		return model.InstitutionRenewalResult{}, err
	}

	if moduleID, lookupErr := repo.findInstitutionVersionModuleIDTx(ctx, tx, nextOpenType); lookupErr != nil {
		return model.InstitutionRenewalResult{}, lookupErr
	} else if moduleID > 0 {
		if err = repo.bindInstitutionModuleTx(
			ctx,
			tx,
			*input.InstitutionID,
			moduleID,
			renewEnd,
			statusValue,
			operatorID,
			nil,
			true,
			false,
		); err != nil {
			return model.InstitutionRenewalResult{}, err
		}
	}

	if err = tx.Commit(); err != nil {
		return model.InstitutionRenewalResult{}, err
	}

	return model.InstitutionRenewalResult{
		InstitutionID:   *input.InstitutionID,
		OpenType:        nextOpenType,
		OpenDuration:    renewDuration,
		ExpireStartTime: renewStart.Time.Format("2006-01-02 15:04:05"),
		ExpireEndTime:   renewEnd.Time.Format("2006-01-02 15:04:05"),
	}, nil
}

func (repo *Repository) FindInstitutionCoordinateByAddress(ctx context.Context, query model.InstitutionGeocodeQuery) (model.InstitutionGeocodeResult, bool, error) {
	filters := []string{"del_flag = 0", "lng IS NOT NULL", "lat IS NOT NULL"}
	args := make([]any, 0, 4)

	if trimmed := strings.TrimSpace(query.Province); trimmed != "" {
		filters = append(filters, "province = ?")
		args = append(args, trimmed)
	}
	if trimmed := strings.TrimSpace(query.City); trimmed != "" {
		filters = append(filters, "city = ?")
		args = append(args, trimmed)
	}
	if trimmed := strings.TrimSpace(query.Region); trimmed != "" {
		filters = append(filters, "region = ?")
		args = append(args, trimmed)
	}
	if trimmed := strings.TrimSpace(query.Address); trimmed != "" {
		filters = append(filters, "address = ?")
		args = append(args, trimmed)
	}

	row := repo.db.QueryRowContext(ctx, `
		SELECT IFNULL(lng, 0),
		       IFNULL(lat, 0),
		       CONCAT(IFNULL(province, ''), IFNULL(city, ''), IFNULL(region, ''), IFNULL(address, ''))
		FROM org_institution
		WHERE `+strings.Join(filters, " AND ")+`
		ORDER BY update_time DESC, id DESC
		LIMIT 1
	`, args...)

	var result model.InstitutionGeocodeResult
	if err := row.Scan(&result.Lng, &result.Lat, &result.ResolvedAddress); err != nil {
		if err == sql.ErrNoRows {
			return model.InstitutionGeocodeResult{}, false, nil
		}
		return model.InstitutionGeocodeResult{}, false, err
	}

	result.Source = "same_address"
	return result, true, nil
}

func (repo *Repository) FindInstitutionCoordinateFallback(ctx context.Context, query model.InstitutionGeocodeQuery) (model.InstitutionGeocodeResult, bool, error) {
	candidates := []struct {
		source  string
		filters []string
		args    []any
		label   string
	}{}

	if region := strings.TrimSpace(query.Region); region != "" {
		candidates = append(candidates, struct {
			source  string
			filters []string
			args    []any
			label   string
		}{
			source:  "region_average",
			filters: []string{"province = ?", "city = ?", "region = ?"},
			args:    []any{strings.TrimSpace(query.Province), strings.TrimSpace(query.City), region},
			label:   strings.TrimSpace(query.Province) + strings.TrimSpace(query.City) + region,
		})
	}

	candidates = append(candidates,
		struct {
			source  string
			filters []string
			args    []any
			label   string
		}{
			source:  "city_average",
			filters: []string{"province = ?", "city = ?"},
			args:    []any{strings.TrimSpace(query.Province), strings.TrimSpace(query.City)},
			label:   strings.TrimSpace(query.Province) + strings.TrimSpace(query.City),
		},
		struct {
			source  string
			filters []string
			args    []any
			label   string
		}{
			source:  "province_average",
			filters: []string{"province = ?"},
			args:    []any{strings.TrimSpace(query.Province)},
			label:   strings.TrimSpace(query.Province),
		},
	)

	for _, candidate := range candidates {
		filters := append([]string{"del_flag = 0", "lng IS NOT NULL", "lat IS NOT NULL"}, candidate.filters...)
		row := repo.db.QueryRowContext(ctx, `
			SELECT ROUND(AVG(lng), 6),
			       ROUND(AVG(lat), 6)
			FROM org_institution
			WHERE `+strings.Join(filters, " AND ")+`
		`, candidate.args...)

		var lng sql.NullFloat64
		var lat sql.NullFloat64
		if err := row.Scan(&lng, &lat); err != nil {
			return model.InstitutionGeocodeResult{}, false, err
		}
		if !lng.Valid || !lat.Valid {
			continue
		}

		return model.InstitutionGeocodeResult{
			Lng:             lng.Float64,
			Lat:             lat.Float64,
			Source:          candidate.source,
			ResolvedAddress: candidate.label,
		}, true, nil
	}

	return model.InstitutionGeocodeResult{}, false, nil
}

func (repo *Repository) nextInstitutionCode(ctx context.Context, tx *sql.Tx) (string, error) {
	var next int
	if err := tx.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(CAST(SUBSTRING(organ_code, 3) AS UNSIGNED)), 0) + 1
		FROM org_institution
		WHERE organ_code LIKE 'I-%'
	`).Scan(&next); err != nil {
		return "", err
	}

	return fmt.Sprintf("I-%05d", next), nil
}

func (repo *Repository) upsertInstitutionProfileTx(ctx context.Context, tx *sql.Tx, institutionID int64, profile model.InstitutionProfile, creatorID, updaterID any) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO org_institution_profile (
			institution_id, description, business_time, video, gallery_images,
			create_id, create_time, update_id, update_time, del_flag
		) VALUES (?, ?, ?, ?, ?, ?, NOW(), ?, NOW(), 0)
		ON DUPLICATE KEY UPDATE
			description = VALUES(description),
			business_time = VALUES(business_time),
			video = VALUES(video),
			gallery_images = VALUES(gallery_images),
			update_id = VALUES(update_id),
			update_time = NOW(),
			del_flag = 0
	`,
		institutionID,
		profile.Description,
		profile.BusinessTime,
		profile.Video,
		marshalStringSlice(profile.GalleryImages),
		creatorID,
		updaterID,
	)
	return err
}

func nullableInt64Value(value *int64) any {
	if value == nil || *value <= 0 {
		return nil
	}
	return *value
}

func nullableInt64(number int64) any {
	if number <= 0 {
		return nil
	}
	return number
}

func nullableInt64Deref(value *int64) int64 {
	if value == nil || *value <= 0 {
		return 0
	}
	return *value
}

func nullableFloat64Value(value *float64) any {
	if value == nil {
		return nil
	}
	return *value
}

func nullableTimeValue(value sql.NullTime) any {
	if !value.Valid {
		return nil
	}
	return value.Time
}

func (repo *Repository) CreateDict(ctx context.Context, input model.DictMutation, creatorID *int64) (int64, error) {
	result, err := repo.db.ExecContext(ctx, `
		INSERT INTO sys_dict (dict_name, dict_code, is_enable, remark, create_id, create_time, update_time, del_flag, version)
		VALUES (?, ?, ?, ?, ?, NOW(), NOW(), 0, 0)
	`, strings.TrimSpace(input.DictName), strings.TrimSpace(input.DictCode), input.IsEnable, strings.TrimSpace(input.Remark), creatorID)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (repo *Repository) UpdateDict(ctx context.Context, input model.DictMutation) error {
	if input.ID == nil {
		return fmt.Errorf("id is required")
	}
	_, err := repo.db.ExecContext(ctx, `
		UPDATE sys_dict
		SET dict_name = ?, dict_code = ?, is_enable = ?, remark = ?, update_time = NOW()
		WHERE id = ? AND del_flag = 0
	`, strings.TrimSpace(input.DictName), strings.TrimSpace(input.DictCode), input.IsEnable, strings.TrimSpace(input.Remark), *input.ID)
	return err
}

func (repo *Repository) DeleteDict(ctx context.Context, id int64) error {
	_, err := repo.db.ExecContext(ctx, `
		UPDATE sys_dict
		SET del_flag = 1, update_time = NOW()
		WHERE id = ? AND del_flag = 0
	`, id)
	return err
}

func (repo *Repository) CreateDictValue(ctx context.Context, input model.DictValueMutation, creatorID *int64) (int64, error) {
	result, err := repo.db.ExecContext(ctx, `
		INSERT INTO sys_dict_value (dict_id, dict_label, dict_value, sort, is_enable, remark, create_id, create_time, update_time, del_flag, version)
		VALUES (?, ?, ?, ?, ?, ?, ?, NOW(), NOW(), 0, 0)
	`, input.DictID, strings.TrimSpace(input.DictLabel), strings.TrimSpace(input.DictValue), input.Sort, input.IsEnable, strings.TrimSpace(input.Remark), creatorID)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (repo *Repository) UpdateDictValue(ctx context.Context, input model.DictValueMutation) error {
	if input.ID == nil {
		return fmt.Errorf("id is required")
	}
	_, err := repo.db.ExecContext(ctx, `
		UPDATE sys_dict_value
		SET dict_id = COALESCE(?, dict_id),
		    dict_label = ?,
		    dict_value = ?,
		    sort = ?,
		    is_enable = ?,
		    remark = ?,
		    update_time = NOW()
		WHERE id = ? AND del_flag = 0
	`, input.DictID, strings.TrimSpace(input.DictLabel), strings.TrimSpace(input.DictValue), input.Sort, input.IsEnable, strings.TrimSpace(input.Remark), *input.ID)
	return err
}

func (repo *Repository) DeleteDictValue(ctx context.Context, id int64) error {
	_, err := repo.db.ExecContext(ctx, `
		UPDATE sys_dict_value
		SET del_flag = 1, update_time = NOW()
		WHERE id = ? AND del_flag = 0
	`, id)
	return err
}

func (repo *Repository) GetModuleDetail(ctx context.Context, moduleID int64) (model.ModuleDetailVO, error) {
	row := repo.db.QueryRowContext(ctx, `
		SELECT m.id,
		       IFNULL(m.uuid, ''),
		       IFNULL(m.version, 0),
		       IFNULL(m.name, ''),
		       IFNULL(m.type, 0),
		       IFNULL(m.price, 0),
		       IFNULL(m.remark, ''),
		       COUNT(DISTINCT CASE WHEN smm.del_flag = 0 THEN smm.menu_id END) AS menu_count,
		       COUNT(DISTINCT CASE WHEN om.del_flag = 0 THEN om.org_id END) AS org_count,
		       IFNULL(DATE_FORMAT(m.create_time, '%Y-%m-%d %H:%i:%s'), ''),
		       IFNULL(DATE_FORMAT(m.update_time, '%Y-%m-%d %H:%i:%s'), '')
		FROM sys_module m
		LEFT JOIN sys_module_menu smm ON smm.module_id = m.id
		LEFT JOIN org_module om ON om.module_id = m.id
		WHERE m.id = ? AND m.del_flag = 0
		GROUP BY m.id, m.uuid, m.version, m.name, m.type, m.price, m.remark, m.create_time, m.update_time
		LIMIT 1
	`, moduleID)

	var detail model.ModuleDetailVO
	if err := row.Scan(
		&detail.ModuleID,
		&detail.UUID,
		&detail.Version,
		&detail.ModuleName,
		&detail.ModuleType,
		&detail.Price,
		&detail.Remark,
		&detail.MenuCount,
		&detail.OrgCount,
		&detail.CreateTime,
		&detail.UpdateTime,
	); err != nil {
		return model.ModuleDetailVO{}, err
	}

	selectedRows, err := repo.db.QueryContext(ctx, `
		SELECT menu_id
		FROM sys_module_menu
		WHERE module_id = ? AND del_flag = 0
	`, moduleID)
	if err != nil {
		return model.ModuleDetailVO{}, err
	}
	defer selectedRows.Close()
	selected := map[int64]struct{}{}
	for selectedRows.Next() {
		var id int64
		if err := selectedRows.Scan(&id); err != nil {
			return model.ModuleDetailVO{}, err
		}
		selected[id] = struct{}{}
		detail.SelectedMenuIDs = append(detail.SelectedMenuIDs, id)
	}
	if err := selectedRows.Err(); err != nil {
		return model.ModuleDetailVO{}, err
	}

	rawMenus, err := repo.listInstitutionMenus(ctx)
	if err != nil {
		return model.ModuleDetailVO{}, err
	}
	detail.MenuIDs = buildVisibleInstitutionModuleTree(rawMenus, selected)
	return detail, nil
}

func (repo *Repository) collectModuleMenus(ctx context.Context, selected map[int64]struct{}) (map[int64]rawMenu, error) {
	menuMap := make(map[int64]rawMenu)
	pending := make([]int64, 0, len(selected))
	for id := range selected {
		pending = append(pending, id)
	}

	for len(pending) > 0 {
		id := pending[0]
		pending = pending[1:]
		if _, exists := menuMap[id]; exists {
			continue
		}
		row := repo.db.QueryRowContext(ctx, `
			SELECT id, IFNULL(menu_name, ''), IFNULL(menu_code, ''), IFNULL(pid, 0), IFNULL(sort, 0), IFNULL(weight, 0), IFNULL(menu_type, 0), IFNULL(group_code, ''), IFNULL(introduce, '')
			FROM sso_menu
			WHERE id = ? AND del_flag = 0 AND own_type = 2
			LIMIT 1
		`, id)
		var item rawMenu
		if err := row.Scan(&item.ID, &item.Name, &item.Code, &item.PID, &item.Sort, &item.Weight, &item.MenuType, &item.GroupCode, &item.Introduce); err != nil {
			if err == sql.ErrNoRows {
				continue
			}
			return nil, err
		}
		menuMap[item.ID] = item
		if item.PID > 0 {
			pending = append(pending, item.PID)
		}
	}

	return menuMap, nil
}

func buildModuleTree(items []rawMenu, selected map[int64]struct{}, pid int64) []model.ModuleMenu {
	childrenItems := make([]rawMenu, 0)
	for _, item := range items {
		if item.PID == pid {
			childrenItems = append(childrenItems, item)
		}
	}
	sort.SliceStable(childrenItems, func(i, j int) bool {
		if childrenItems[i].Sort == childrenItems[j].Sort {
			if childrenItems[i].Weight == childrenItems[j].Weight {
				return childrenItems[i].ID < childrenItems[j].ID
			}
			return childrenItems[i].Weight > childrenItems[j].Weight
		}
		return childrenItems[i].Sort < childrenItems[j].Sort
	})

	result := make([]model.ModuleMenu, 0, len(childrenItems))
	for _, item := range childrenItems {
		children := buildModuleTree(items, selected, item.ID)
		menu := model.ModuleMenu{
			MenuID:    strconv.FormatInt(item.ID, 10),
			MenuName:  item.Name,
			Introduce: item.Introduce,
			MenuType:  item.MenuType,
			GroupCode: item.GroupCode,
			Weight:    item.Weight,
			Children:  children,
		}
		if len(children) == 0 {
			_, menu.IsSelect = selected[item.ID]
		} else {
			allSelected := true
			for _, child := range children {
				if !child.IsSelect {
					allSelected = false
					break
				}
			}
			menu.IsSelect = allSelected
		}
		result = append(result, menu)
	}
	return result
}

func (repo *Repository) ListModuleMenuTree(ctx context.Context, moduleType int) ([]model.ModuleMenu, error) {
	_ = moduleType
	rawMenus, err := repo.listInstitutionMenus(ctx)
	if err != nil {
		return nil, err
	}
	if len(rawMenus) == 0 {
		return []model.ModuleMenu{}, nil
	}

	selected := make(map[int64]struct{}, len(rawMenus))
	for _, item := range rawMenus {
		selected[item.ID] = struct{}{}
	}
	return buildVisibleInstitutionModuleTree(rawMenus, selected), nil
}

func (repo *Repository) listInstitutionMenus(ctx context.Context) ([]rawMenu, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT id, IFNULL(menu_name, ''), IFNULL(menu_code, ''), IFNULL(pid, 0), IFNULL(sort, 0), IFNULL(weight, 0), IFNULL(menu_type, 0), IFNULL(group_code, ''), IFNULL(introduce, '')
		FROM sso_menu
		WHERE del_flag = 0 AND own_type = 2
		ORDER BY IFNULL(level, 0) ASC, IFNULL(sort, 0) ASC, IFNULL(weight, 0) DESC, id ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]rawMenu, 0, 256)
	for rows.Next() {
		var item rawMenu
		if err := rows.Scan(&item.ID, &item.Name, &item.Code, &item.PID, &item.Sort, &item.Weight, &item.MenuType, &item.GroupCode, &item.Introduce); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (repo *Repository) listScopedInstitutionMenuIDs(ctx context.Context, moduleType int) (map[int64]struct{}, error) {
	menuIDs, err := repo.queryMenuIDSet(ctx, `
		SELECT DISTINCT scoped.menu_id
		FROM (
			SELECT smm.menu_id
			FROM sys_module sm
			JOIN sys_module_menu smm ON smm.module_id = sm.id AND smm.del_flag = 0
			JOIN sso_menu m ON m.id = smm.menu_id AND m.del_flag = 0 AND m.own_type = 2
			WHERE sm.del_flag = 0 AND sm.type = ?

			UNION

			SELECT rm.menu_id
			FROM sso_role r
			JOIN sso_role_menu rm ON rm.role_id = r.id
			JOIN sso_menu m ON m.id = rm.menu_id AND m.del_flag = 0 AND m.own_type = 2
			WHERE r.del_flag = 0 AND r.role_type = 2 AND r.is_admin = 1
		) scoped
	`, moduleType)
	if err != nil {
		return nil, err
	}
	if len(menuIDs) > 0 {
		return menuIDs, nil
	}

	return repo.queryMenuIDSet(ctx, `
		SELECT id
		FROM sso_menu
		WHERE del_flag = 0 AND own_type = 2
	`)
}

func (repo *Repository) queryMenuIDSet(ctx context.Context, query string, args ...any) (map[int64]struct{}, error) {
	rows, err := repo.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[int64]struct{})
	for rows.Next() {
		var menuID int64
		if err := rows.Scan(&menuID); err != nil {
			return nil, err
		}
		if menuID > 0 {
			result[menuID] = struct{}{}
		}
	}
	return result, rows.Err()
}

func (repo *Repository) IncreaseModuleMenus(ctx context.Context, input model.ModulePermissionMutation) error {
	if input.ID == nil {
		return fmt.Errorf("id is required")
	}
	currentMenus, err := repo.getModuleMenuIDs(ctx, *input.ID)
	if err != nil {
		return err
	}
	existing := map[int64]struct{}{}
	for _, id := range currentMenus {
		existing[id] = struct{}{}
	}
	addMenus := make([]int64, 0)
	for _, id := range input.MenuIDs {
		if _, ok := existing[id]; !ok {
			addMenus = append(addMenus, id)
		}
	}
	if len(addMenus) == 0 {
		return nil
	}
	if err := repo.insertModuleMenus(ctx, *input.ID, addMenus); err != nil {
		return err
	}
	roleIDs, err := repo.getRoleIDsByModule(ctx, *input.ID, input.IsAllRole != nil && *input.IsAllRole)
	if err != nil {
		return err
	}
	if len(roleIDs) == 0 {
		return nil
	}
	return repo.insertIgnoreRoleMenus(ctx, roleIDs, addMenus)
}

func (repo *Repository) DecreaseModuleMenus(ctx context.Context, input model.ModulePermissionMutation) error {
	if input.ID == nil {
		return fmt.Errorf("id is required")
	}
	currentMenus, err := repo.getModuleMenuIDs(ctx, *input.ID)
	if err != nil {
		return err
	}
	keep := map[int64]struct{}{}
	for _, id := range input.MenuIDs {
		keep[id] = struct{}{}
	}
	removeMenus := make([]int64, 0)
	for _, id := range currentMenus {
		if _, ok := keep[id]; !ok {
			removeMenus = append(removeMenus, id)
		}
	}
	if len(removeMenus) == 0 {
		return nil
	}
	if err := repo.deleteModuleMenus(ctx, *input.ID, removeMenus); err != nil {
		return err
	}
	roleIDs, err := repo.getRoleIDsByModule(ctx, *input.ID, true)
	if err != nil {
		return err
	}
	if len(roleIDs) == 0 {
		return nil
	}
	return repo.deleteRoleMenus(ctx, roleIDs, removeMenus)
}

func (repo *Repository) ReplaceModuleMenus(ctx context.Context, input model.ModulePermissionMutation) error {
	if input.ID == nil || *input.ID <= 0 {
		return fmt.Errorf("id is required")
	}

	currentMenus, err := repo.getModuleMenuIDs(ctx, *input.ID)
	if err != nil {
		return err
	}

	currentSet := make(map[int64]struct{}, len(currentMenus))
	for _, id := range currentMenus {
		currentSet[id] = struct{}{}
	}

	nextSet := make(map[int64]struct{}, len(input.MenuIDs))
	nextMenus := make([]int64, 0, len(input.MenuIDs))
	for _, id := range input.MenuIDs {
		if id <= 0 {
			continue
		}
		if _, exists := nextSet[id]; exists {
			continue
		}
		nextSet[id] = struct{}{}
		nextMenus = append(nextMenus, id)
	}

	addMenus := make([]int64, 0)
	for _, id := range nextMenus {
		if _, exists := currentSet[id]; !exists {
			addMenus = append(addMenus, id)
		}
	}

	removeMenus := make([]int64, 0)
	for _, id := range currentMenus {
		if _, exists := nextSet[id]; !exists {
			removeMenus = append(removeMenus, id)
		}
	}

	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if len(addMenus) > 0 {
		for _, menuID := range addMenus {
			if _, err = tx.ExecContext(ctx, `
				INSERT INTO sys_module_menu (module_id, menu_id, create_time, del_flag, version)
				VALUES (?, ?, NOW(), 0, 0)
			`, *input.ID, menuID); err != nil {
				return err
			}
		}
	}

	if len(removeMenus) > 0 {
		placeholders := make([]string, 0, len(removeMenus))
		args := make([]any, 0, len(removeMenus)+1)
		args = append(args, *input.ID)
		for _, id := range removeMenus {
			placeholders = append(placeholders, "?")
			args = append(args, id)
		}
		if _, err = tx.ExecContext(ctx, `
			DELETE FROM sys_module_menu
			WHERE module_id = ? AND menu_id IN (`+strings.Join(placeholders, ",")+`)
		`, args...); err != nil {
			return err
		}
	}

	roleIDs, err := repo.getRoleIDsByModuleTx(ctx, tx, *input.ID, true)
	if err != nil {
		return err
	}

	if len(removeMenus) > 0 && len(roleIDs) > 0 {
		if err = repo.deleteRoleMenusTx(ctx, tx, roleIDs, removeMenus); err != nil {
			return err
		}
	}

	if len(addMenus) > 0 && len(roleIDs) > 0 {
		adminRoleIDs := roleIDs
		if input.IsAllRole == nil || !*input.IsAllRole {
			adminRoleIDs, err = repo.getRoleIDsByModuleTx(ctx, tx, *input.ID, false)
			if err != nil {
				return err
			}
		}
		if len(adminRoleIDs) > 0 {
			if err = repo.insertIgnoreRoleMenusTx(ctx, tx, adminRoleIDs, addMenus); err != nil {
				return err
			}
		}
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (repo *Repository) getModuleMenuIDs(ctx context.Context, moduleID int64) ([]int64, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT menu_id
		FROM sys_module_menu
		WHERE module_id = ? AND del_flag = 0
	`, moduleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]int64, 0, 16)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		items = append(items, id)
	}
	return items, rows.Err()
}

func (repo *Repository) insertModuleMenus(ctx context.Context, moduleID int64, menuIDs []int64) error {
	for _, menuID := range menuIDs {
		if _, err := repo.db.ExecContext(ctx, `
			INSERT INTO sys_module_menu (module_id, menu_id, create_time, del_flag, version)
			VALUES (?, ?, NOW(), 0, 0)
		`, moduleID, menuID); err != nil {
			return err
		}
	}
	return nil
}

func (repo *Repository) deleteModuleMenus(ctx context.Context, moduleID int64, menuIDs []int64) error {
	placeholders := make([]string, 0, len(menuIDs))
	args := make([]any, 0, len(menuIDs)+1)
	args = append(args, moduleID)
	for _, id := range menuIDs {
		placeholders = append(placeholders, "?")
		args = append(args, id)
	}
	_, err := repo.db.ExecContext(ctx, `
		DELETE FROM sys_module_menu
		WHERE module_id = ? AND menu_id IN (`+strings.Join(placeholders, ",")+`)
	`, args...)
	return err
}

func (repo *Repository) getRoleIDsByModule(ctx context.Context, moduleID int64, allRoles bool) ([]int64, error) {
	query := `
		SELECT r.id
		FROM sso_role r
		LEFT JOIN org_module m ON r.org_id = m.org_id
		WHERE m.module_id = ? AND r.del_flag = 0`
	if !allRoles {
		query += " AND r.is_admin = 1"
	}
	rows, err := repo.db.QueryContext(ctx, query, moduleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]int64, 0, 16)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		items = append(items, id)
	}
	return items, rows.Err()
}

func (repo *Repository) getRoleIDsByModuleTx(ctx context.Context, tx *sql.Tx, moduleID int64, allRoles bool) ([]int64, error) {
	query := `
		SELECT r.id
		FROM sso_role r
		LEFT JOIN org_module m ON r.org_id = m.org_id
		WHERE m.module_id = ? AND IFNULL(m.del_flag, 0) = 0 AND r.del_flag = 0`
	if !allRoles {
		query += " AND r.is_admin = 1"
	}
	rows, err := tx.QueryContext(ctx, query, moduleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]int64, 0, 16)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		items = append(items, id)
	}
	return items, rows.Err()
}

func (repo *Repository) insertIgnoreRoleMenus(ctx context.Context, roleIDs, menuIDs []int64) error {
	for _, roleID := range roleIDs {
		for _, menuID := range menuIDs {
			if _, err := repo.db.ExecContext(ctx, `
				INSERT IGNORE INTO sso_role_menu (role_id, menu_id)
				VALUES (?, ?)
			`, roleID, menuID); err != nil {
				return err
			}
		}
	}
	return nil
}

func (repo *Repository) insertIgnoreRoleMenusTx(ctx context.Context, tx *sql.Tx, roleIDs, menuIDs []int64) error {
	for _, roleID := range roleIDs {
		for _, menuID := range menuIDs {
			if _, err := tx.ExecContext(ctx, `
				INSERT IGNORE INTO sso_role_menu (role_id, menu_id)
				VALUES (?, ?)
			`, roleID, menuID); err != nil {
				return err
			}
		}
	}
	return nil
}

func (repo *Repository) deleteRoleMenus(ctx context.Context, roleIDs, menuIDs []int64) error {
	if len(roleIDs) == 0 || len(menuIDs) == 0 {
		return nil
	}
	rolePlaceholders := make([]string, 0, len(roleIDs))
	menuPlaceholders := make([]string, 0, len(menuIDs))
	args := make([]any, 0, len(roleIDs)+len(menuIDs))
	for _, id := range roleIDs {
		rolePlaceholders = append(rolePlaceholders, "?")
		args = append(args, id)
	}
	for _, id := range menuIDs {
		menuPlaceholders = append(menuPlaceholders, "?")
		args = append(args, id)
	}
	_, err := repo.db.ExecContext(ctx, `
		DELETE FROM sso_role_menu
		WHERE role_id IN (`+strings.Join(rolePlaceholders, ",")+`)
		  AND menu_id IN (`+strings.Join(menuPlaceholders, ",")+`)
	`, args...)
	return err
}

func (repo *Repository) deleteRoleMenusTx(ctx context.Context, tx *sql.Tx, roleIDs, menuIDs []int64) error {
	if len(roleIDs) == 0 || len(menuIDs) == 0 {
		return nil
	}
	rolePlaceholders := make([]string, 0, len(roleIDs))
	menuPlaceholders := make([]string, 0, len(menuIDs))
	args := make([]any, 0, len(roleIDs)+len(menuIDs))
	for _, id := range roleIDs {
		rolePlaceholders = append(rolePlaceholders, "?")
		args = append(args, id)
	}
	for _, id := range menuIDs {
		menuPlaceholders = append(menuPlaceholders, "?")
		args = append(args, id)
	}
	_, err := tx.ExecContext(ctx, `
		DELETE FROM sso_role_menu
		WHERE role_id IN (`+strings.Join(rolePlaceholders, ",")+`)
		  AND menu_id IN (`+strings.Join(menuPlaceholders, ",")+`)
	`, args...)
	return err
}

func (repo *Repository) CreateModule(ctx context.Context, input model.ModuleMutation) (int64, error) {
	result, err := repo.db.ExecContext(ctx, `
		INSERT INTO sys_module (uuid, name, type, price, remark, del_flag, create_time, update_time, version)
		VALUES (?, ?, ?, ?, ?, 0, NOW(), NOW(), 0)
	`, uuid.NewString(), strings.TrimSpace(input.Name), input.Type, input.Price, strings.TrimSpace(input.Remark))
	if err != nil {
		return 0, err
	}
	moduleID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	if len(input.MenuIDs) > 0 {
		if err := repo.insertModuleMenus(ctx, moduleID, input.MenuIDs); err != nil {
			return 0, err
		}
	}
	return moduleID, nil
}

func (repo *Repository) UpdateModuleBasic(ctx context.Context, input model.ModuleMutation) error {
	if input.ID == nil {
		return fmt.Errorf("id is required")
	}
	_, err := repo.db.ExecContext(ctx, `
		UPDATE sys_module
		SET name = ?, type = ?, price = ?, remark = ?, update_time = NOW()
		WHERE id = ? AND del_flag = 0
	`, strings.TrimSpace(input.Name), input.Type, input.Price, strings.TrimSpace(input.Remark), *input.ID)
	return err
}

func (repo *Repository) ReplaceInstitutionModule(ctx context.Context, input model.InstitutionPermissionMutation, operatorID *int64) error {
	if input.InstitutionID == nil || *input.InstitutionID <= 0 {
		return fmt.Errorf("institutionId is required")
	}
	if input.ModuleID == nil || *input.ModuleID <= 0 {
		return fmt.Errorf("moduleId is required")
	}

	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var currentOpenType sql.NullInt64
	var currentOpenDuration sql.NullString
	var expireEnd sql.NullTime
	var statusValue sql.NullInt64
	if err = tx.QueryRowContext(ctx, `
		SELECT IFNULL(open_type, 2),
		       IFNULL(open_duration, ''),
		       expire_end_time,
		       IFNULL(status, 0)
		FROM org_institution
		WHERE id = ? AND del_flag = 0
		LIMIT 1
		FOR UPDATE
	`, *input.InstitutionID).Scan(&currentOpenType, &currentOpenDuration, &expireEnd, &statusValue); err != nil {
		return err
	}

	currentModuleID, currentModuleName, lookupCurrentModuleErr := repo.getInstitutionCurrentModuleTx(ctx, tx, *input.InstitutionID)
	if lookupCurrentModuleErr != nil {
		return lookupCurrentModuleErr
	}

	moduleName, lookupErr := repo.getModuleNameTx(ctx, tx, *input.ModuleID)
	if lookupErr != nil {
		return lookupErr
	}

	nextOpenType := institutionModuleNameOpenType(moduleName)
	storedOpenType := institutionStoredOpenType(currentOpenType)
	if storedOpenType == 0 {
		storedOpenType = 2
	}
	if nextOpenType > 0 {
		if storedOpenType >= 2 && nextOpenType == 1 {
			return fmt.Errorf("已开通基础版、高级版或旗舰版的机构不支持降为体验版")
		}
		if storedOpenType == 1 && nextOpenType >= 2 {
			return fmt.Errorf("体验版升级正式版本请通过续期操作处理")
		}

		nextOpenDuration := normalizeInstitutionOpenDuration(nextOpenType, currentOpenDuration.String)
		if _, err = tx.ExecContext(ctx, `
			UPDATE org_institution
			SET open_type = ?,
			    open_duration = ?,
			    update_id = ?,
			    update_time = NOW()
			WHERE id = ? AND del_flag = 0
		`,
			nextOpenType,
			nextOpenDuration,
			nullableInt64Value(operatorID),
			*input.InstitutionID,
		); err != nil {
			return err
		}
	}

	if err = repo.bindInstitutionModuleTx(
		ctx,
		tx,
		*input.InstitutionID,
		*input.ModuleID,
		expireEnd,
		int(statusValue.Int64),
		operatorID,
		input.MenuIDs,
		true,
		true,
	); err != nil {
		return err
	}

	shouldWriteVersionChangeLog := nextOpenType > 0 && (currentModuleID != *input.ModuleID || storedOpenType != nextOpenType)
	if shouldWriteVersionChangeLog {
		beforeVersionName := strings.TrimSpace(currentModuleName)
		if beforeVersionName == "" {
			beforeVersionName = institutionOpenTypeModuleName(storedOpenType)
		}
		if err = repo.insertInstitutionVersionChangeRecordTx(ctx, tx, model.InstitutionVersionChangeRecord{
			InstitutionID:     *input.InstitutionID,
			BeforeOpenType:    storedOpenType,
			BeforeModuleID:    currentModuleID,
			BeforeVersionName: beforeVersionName,
			AfterOpenType:     nextOpenType,
			AfterModuleID:     *input.ModuleID,
			AfterVersionName:  moduleName,
			OperatorID:        nullableInt64Deref(operatorID),
		}); err != nil {
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (repo *Repository) getModuleNameTx(ctx context.Context, tx *sql.Tx, moduleID int64) (string, error) {
	var moduleName sql.NullString
	if err := tx.QueryRowContext(ctx, `
		SELECT IFNULL(name, '')
		FROM sys_module
		WHERE id = ? AND del_flag = 0
		LIMIT 1
	`, moduleID).Scan(&moduleName); err != nil {
		return "", err
	}
	return strings.TrimSpace(moduleName.String), nil
}

func (repo *Repository) getInstitutionCurrentModuleTx(ctx context.Context, tx *sql.Tx, institutionID int64) (int64, string, error) {
	var moduleID sql.NullInt64
	var moduleName sql.NullString
	err := tx.QueryRowContext(ctx, `
		SELECT IFNULL(om.module_id, 0), IFNULL(sm.name, '')
		FROM org_module om
		LEFT JOIN sys_module sm ON sm.id = om.module_id AND sm.del_flag = 0
		WHERE om.org_id = ? AND om.del_flag = 0
		ORDER BY om.id DESC
		LIMIT 1
	`, institutionID).Scan(&moduleID, &moduleName)
	if err == sql.ErrNoRows {
		return 0, "", nil
	}
	if err != nil {
		return 0, "", err
	}
	return moduleID.Int64, strings.TrimSpace(moduleName.String), nil
}

func (repo *Repository) insertInstitutionVersionChangeRecordTx(ctx context.Context, tx *sql.Tx, record model.InstitutionVersionChangeRecord) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO org_institution_version_change_record (
			institution_id,
			before_open_type,
			before_module_id,
			before_version_name,
			after_open_type,
			after_module_id,
			after_version_name,
			operator_id,
			create_time,
			update_time,
			del_flag
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW(), 0)
	`,
		record.InstitutionID,
		record.BeforeOpenType,
		nullableInt64(record.BeforeModuleID),
		strings.TrimSpace(record.BeforeVersionName),
		record.AfterOpenType,
		nullableInt64(record.AfterModuleID),
		strings.TrimSpace(record.AfterVersionName),
		nullableInt64(record.OperatorID),
	)
	return err
}

func (repo *Repository) GetInstitutionPermissionDetail(ctx context.Context, institutionID int64) (model.InstitutionPermissionDetail, error) {
	row := repo.db.QueryRowContext(ctx, `
		SELECT oi.id,
		       IFNULL(oi.organ_name, ''),
		       IFNULL(oi.mobile, ''),
		       IFNULL(oi.open_type, 2),
		       IFNULL(oi.open_duration, ''),
		       `+institutionStatusExpr("oi")+`,
		       IFNULL(DATE_FORMAT(oi.expire_end_time, '%Y-%m-%d %H:%i:%s'), ''),
		       IFNULL(om.module_id, 0),
		       IFNULL(sm.name, ''),
		       IFNULL(sr.id, 0),
		       IFNULL(sr.role_name, '')
		FROM org_institution oi
		LEFT JOIN org_module om ON om.org_id = oi.id AND om.del_flag = 0
		LEFT JOIN sys_module sm ON sm.id = om.module_id AND sm.del_flag = 0
		LEFT JOIN sso_role sr ON sr.org_id = oi.id AND sr.role_type = 2 AND sr.is_admin = 1 AND sr.del_flag = 0
		WHERE oi.id = ? AND oi.del_flag = 0
		LIMIT 1
	`, institutionID)

	var detail model.InstitutionPermissionDetail
	if err := row.Scan(
		&detail.InstitutionID,
		&detail.OrganName,
		&detail.Mobile,
		&detail.OpenType,
		&detail.OpenDuration,
		&detail.Status,
		&detail.ExpireEndTime,
		&detail.CurrentModuleID,
		&detail.CurrentModuleName,
		&detail.AdminRoleID,
		&detail.AdminRoleName,
	); err != nil {
		return model.InstitutionPermissionDetail{}, err
	}

	if detail.CurrentModuleID > 0 {
		menuIDs, err := repo.getModuleMenuIDs(ctx, detail.CurrentModuleID)
		if err != nil {
			return model.InstitutionPermissionDetail{}, err
		}
		detail.TemplateMenuIDs = menuIDs
	}

	if detail.AdminRoleID > 0 {
		effectiveMenuIDs, err := repo.getRoleMenuIDsByRoleAndOwnType(ctx, detail.AdminRoleID, 2)
		if err != nil {
			return model.InstitutionPermissionDetail{}, err
		}
		detail.EffectiveMenuIDs = effectiveMenuIDs
	}

	return detail, nil
}

func (repo *Repository) bindInstitutionModuleTx(ctx context.Context, tx *sql.Tx, institutionID, moduleID int64, expireEnd sql.NullTime, status int, operatorID *int64, customMenuIDs []int64, syncMenus bool, strict bool) error {
	if err := repo.upsertOrgModuleTx(ctx, tx, institutionID, moduleID, expireEnd, status, operatorID); err != nil {
		return err
	}
	if !syncMenus {
		return nil
	}

	moduleMenuIDs, err := repo.getModuleMenuIDsTx(ctx, tx, moduleID)
	if err != nil {
		return err
	}

	finalMenuIDs := moduleMenuIDs
	if len(customMenuIDs) > 0 {
		scopedMenuIDs, scopeErr := sanitizeInstitutionMenuScope(customMenuIDs, moduleMenuIDs)
		if scopeErr != nil {
			return scopeErr
		}
		finalMenuIDs, err = repo.mergeInstitutionBaseMenusTx(ctx, tx, moduleMenuIDs, scopedMenuIDs)
		if err != nil {
			return err
		}
	}

	adminRoleID, _, err := repo.getInstitutionAdminRoleTx(ctx, tx, institutionID)
	if err != nil {
		if err == sql.ErrNoRows && !strict {
			return nil
		}
		if err == sql.ErrNoRows {
			return fmt.Errorf("机构管理员角色不存在，请先初始化机构账号")
		}
		return err
	}

	if err := repo.replaceRoleMenusTx(ctx, tx, adminRoleID, finalMenuIDs); err != nil {
		return err
	}

	return repo.deleteInstitutionMenusOutsideScopeTx(ctx, tx, institutionID, finalMenuIDs)
}

func (repo *Repository) mergeInstitutionBaseMenusTx(ctx context.Context, tx *sql.Tx, moduleMenuIDs, customMenuIDs []int64) ([]int64, error) {
	if len(moduleMenuIDs) == 0 {
		return customMenuIDs, nil
	}

	placeholders := make([]string, 0, len(moduleMenuIDs))
	args := make([]any, 0, len(moduleMenuIDs))
	for _, menuID := range moduleMenuIDs {
		if menuID <= 0 {
			continue
		}
		placeholders = append(placeholders, "?")
		args = append(args, menuID)
	}
	if len(placeholders) == 0 {
		return customMenuIDs, nil
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT id, IFNULL(menu_code, '')
		FROM sso_menu
		WHERE del_flag = 0 AND id IN (`+strings.Join(placeholders, ",")+`)
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	merged := make([]int64, 0, len(moduleMenuIDs))
	seen := make(map[int64]struct{}, len(moduleMenuIDs))

	for rows.Next() {
		var menuID int64
		var menuCode string
		if err := rows.Scan(&menuID, &menuCode); err != nil {
			return nil, err
		}
		if menuID <= 0 || !isInstitutionBaseMenuCode(menuCode) {
			continue
		}
		if _, exists := seen[menuID]; exists {
			continue
		}
		seen[menuID] = struct{}{}
		merged = append(merged, menuID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for _, menuID := range customMenuIDs {
		if menuID <= 0 {
			continue
		}
		if _, exists := seen[menuID]; exists {
			continue
		}
		seen[menuID] = struct{}{}
		merged = append(merged, menuID)
	}

	sort.Slice(merged, func(i, j int) bool {
		return merged[i] < merged[j]
	})

	return merged, nil
}

func (repo *Repository) upsertOrgModuleTx(ctx context.Context, tx *sql.Tx, institutionID, moduleID int64, expireEnd sql.NullTime, status int, operatorID *int64) error {
	updaterValue := nullableInt64Value(operatorID)

	var existingID int64
	err := tx.QueryRowContext(ctx, `
		SELECT id
		FROM org_module
		WHERE org_id = ? AND del_flag = 0
		ORDER BY id DESC
		LIMIT 1
		FOR UPDATE
	`, institutionID).Scan(&existingID)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	if err == sql.ErrNoRows {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO org_module (uuid, version, org_id, module_id, expire_time, status, create_id, create_time, update_id, update_time, del_flag, remark)
			VALUES (?, 0, ?, ?, ?, ?, ?, NOW(), ?, NOW(), 0, '')
		`, uuid.NewString(), institutionID, moduleID, nullableTimeValue(expireEnd), status, updaterValue, updaterValue)
		return err
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE org_module
		SET module_id = ?,
		    expire_time = ?,
		    status = ?,
		    update_id = ?,
		    update_time = NOW()
		WHERE id = ?
	`, moduleID, nullableTimeValue(expireEnd), status, updaterValue, existingID)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE org_module
		SET del_flag = 1,
		    update_id = ?,
		    update_time = NOW()
		WHERE org_id = ? AND del_flag = 0 AND id <> ?
	`, updaterValue, institutionID, existingID)
	return err
}

func (repo *Repository) getModuleMenuIDsTx(ctx context.Context, tx *sql.Tx, moduleID int64) ([]int64, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT menu_id
		FROM sys_module_menu
		WHERE module_id = ? AND del_flag = 0
	`, moduleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]int64, 0, 32)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		items = append(items, id)
	}
	return items, rows.Err()
}

func (repo *Repository) getInstitutionAdminRoleTx(ctx context.Context, tx *sql.Tx, institutionID int64) (int64, string, error) {
	var roleID int64
	var roleName string
	err := tx.QueryRowContext(ctx, `
		SELECT id, IFNULL(role_name, '')
		FROM sso_role
		WHERE org_id = ? AND role_type = 2 AND is_admin = 1 AND del_flag = 0
		LIMIT 1
	`, institutionID).Scan(&roleID, &roleName)
	if err != nil {
		return 0, "", err
	}
	return roleID, roleName, nil
}

func (repo *Repository) replaceRoleMenusTx(ctx context.Context, tx *sql.Tx, roleID int64, menuIDs []int64) error {
	if _, err := tx.ExecContext(ctx, `
		DELETE FROM sso_role_menu
		WHERE role_id = ?
	`, roleID); err != nil {
		return err
	}

	if len(menuIDs) == 0 {
		return nil
	}

	for _, menuID := range menuIDs {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO sso_role_menu (role_id, menu_id)
			VALUES (?, ?)
		`, roleID, menuID); err != nil {
			return err
		}
	}
	return nil
}

func (repo *Repository) deleteInstitutionMenusOutsideScopeTx(ctx context.Context, tx *sql.Tx, institutionID int64, allowedMenuIDs []int64) error {
	if len(allowedMenuIDs) == 0 {
		_, err := tx.ExecContext(ctx, `
			DELETE rm
			FROM sso_role_menu rm
			JOIN sso_role r ON r.id = rm.role_id
			JOIN sso_menu m ON m.id = rm.menu_id
			WHERE r.org_id = ? AND r.del_flag = 0 AND m.own_type = 2
		`, institutionID)
		return err
	}

	placeholders := make([]string, 0, len(allowedMenuIDs))
	args := make([]any, 0, len(allowedMenuIDs)+1)
	args = append(args, institutionID)
	for _, menuID := range allowedMenuIDs {
		placeholders = append(placeholders, "?")
		args = append(args, menuID)
	}

	_, err := tx.ExecContext(ctx, `
		DELETE rm
		FROM sso_role_menu rm
		JOIN sso_role r ON r.id = rm.role_id
		JOIN sso_menu m ON m.id = rm.menu_id
		WHERE r.org_id = ? AND r.del_flag = 0 AND m.own_type = 2
		  AND rm.menu_id NOT IN (`+strings.Join(placeholders, ",")+`)
	`, args...)
	return err
}

func (repo *Repository) getRoleMenuIDsByRoleAndOwnType(ctx context.Context, roleID int64, ownType int) ([]int64, error) {
	rows, err := repo.db.QueryContext(ctx, `
		SELECT rm.menu_id
		FROM sso_role_menu rm
		JOIN sso_menu m ON m.id = rm.menu_id
		WHERE rm.role_id = ? AND m.del_flag = 0 AND m.own_type = ?
		ORDER BY rm.menu_id
	`, roleID, ownType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]int64, 0, 32)
	for rows.Next() {
		var menuID int64
		if err := rows.Scan(&menuID); err != nil {
			return nil, err
		}
		items = append(items, menuID)
	}
	return items, rows.Err()
}

func (repo *Repository) findInstitutionVersionModuleIDTx(ctx context.Context, tx *sql.Tx, openType int) (int64, error) {
	moduleName := institutionOpenTypeModuleName(openType)
	if strings.TrimSpace(moduleName) == "" {
		return 0, nil
	}

	var moduleID sql.NullInt64
	if err := tx.QueryRowContext(ctx, `
		SELECT id
		FROM sys_module
		WHERE del_flag = 0 AND type = 1 AND name = ?
		ORDER BY id ASC
		LIMIT 1
	`, moduleName).Scan(&moduleID); err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}

	if !moduleID.Valid || moduleID.Int64 <= 0 {
		return 0, nil
	}
	return moduleID.Int64, nil
}

func (repo *Repository) CreateNotice(ctx context.Context, input model.NoticeMutation, creatorID *int64) (int64, error) {
	result, err := repo.db.ExecContext(ctx, `
		INSERT INTO sys_notice_info (title, content, disable_id, compel, create_id, create_time, update_time, del_flag, version)
		VALUES (?, ?, ?, ?, ?, NOW(), NOW(), 0, 0)
	`, strings.TrimSpace(input.Title), strings.TrimSpace(input.Content), input.DisableID, input.Compel, creatorID)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (repo *Repository) UpdateNotice(ctx context.Context, input model.NoticeMutation) error {
	if input.ID == nil {
		return fmt.Errorf("id is required")
	}
	_, err := repo.db.ExecContext(ctx, `
		UPDATE sys_notice_info
		SET title = ?, content = ?, compel = ?, disable_id = COALESCE(?, disable_id), update_time = NOW()
		WHERE id = ? AND del_flag = 0
	`, strings.TrimSpace(input.Title), strings.TrimSpace(input.Content), input.Compel, input.DisableID, *input.ID)
	return err
}

func (repo *Repository) DeleteNotice(ctx context.Context, id int64) error {
	_, err := repo.db.ExecContext(ctx, `
		UPDATE sys_notice_info
		SET del_flag = 1, update_time = NOW()
		WHERE id = ? AND del_flag = 0
	`, id)
	return err
}
