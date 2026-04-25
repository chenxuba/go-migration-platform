package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"go-migration-platform/pkg/institutionmenu"
	"go-migration-platform/pkg/tenantstorage"
	"go-migration-platform/services/platform/internal/model"
	"golang.org/x/crypto/bcrypt"
)

type Repository struct {
	db *sql.DB
}

type loginNameQueryer interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type governmentOverviewContext struct {
	Level    string
	Disabled bool
	IsAdmin  bool
	Scopes   []governmentOverviewScope
}

type governmentOverviewScope struct {
	ScopeLevel   string
	ProvinceCode string
	ProvinceName string
	CityCode     string
	CityName     string
	DistrictCode string
	DistrictName string
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

const (
	// defaultInstitutionBootstrapPassword is the initial login password for the institution admin (login_name).
	defaultInstitutionBootstrapPassword = "123456"
	// defaultInstitutionAdminRoleName is the hidden super-admin role name for institution side.
	defaultInstitutionAdminRoleName = "超级管理员"
	// defaultInstitutionAdminRoleDescription is the fixed description for the hidden super-admin role.
	defaultInstitutionAdminRoleDescription = "系统内置角色，拥有机构全部权限"
)

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
	normalized := strings.TrimSpace(moduleName)
	switch normalized {
	case "体验版":
		return 1
	case "基础版":
		return 2
	case "高级版":
		return 3
	case "旗舰版":
		return 4
	}

	switch {
	case strings.Contains(normalized, "体验"):
		return 1
	case strings.Contains(normalized, "基础"):
		return 2
	case strings.Contains(normalized, "高级"):
		return 3
	case strings.Contains(normalized, "旗舰"):
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
		"WHEN IFNULL(" + alias + ".enabled, 0) = 0 THEN 2 " +
		"WHEN " + alias + ".expire_end_time IS NOT NULL AND " + alias + ".expire_end_time < NOW() THEN 4 " +
		"WHEN " + alias + ".expire_end_time IS NOT NULL AND " + alias + ".expire_end_time <= DATE_ADD(NOW(), INTERVAL 1 MONTH) THEN 3 " +
		"ELSE 1 END"
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

func normalizeGovernmentLevel(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "super", "超级监管":
		return "super"
	case "province", "省级":
		return "province"
	case "city", "市级":
		return "city"
	case "district", "区县级", "区级", "县级":
		return "district"
	default:
		return ""
	}
}

func governmentLevelLabel(level string) string {
	switch normalizeGovernmentLevel(level) {
	case "super":
		return "超级监管"
	case "province":
		return "省级"
	case "city":
		return "市级"
	case "district":
		return "区县级"
	default:
		return "--"
	}
}

func governmentOverviewGroupMeta(level string) (codeExpr, nameExpr, levelLabel string, hasChildren bool) {
	switch normalizeGovernmentLevel(level) {
	case "super":
		return "CAST(IFNULL(oi.province_code, 0) AS CHAR)", "COALESCE(NULLIF(TRIM(oi.province), ''), '未设置省份')", "省级", true
	case "province":
		return "CAST(IFNULL(oi.city_code, 0) AS CHAR)", "COALESCE(NULLIF(TRIM(oi.city), ''), '未设置城市')", "市级", true
	case "city":
		return "CAST(IFNULL(oi.region_code, 0) AS CHAR)", "COALESCE(NULLIF(TRIM(oi.region), ''), '未设置区县')", "区县级", true
	case "district":
		return "CAST(IFNULL(oi.region_code, 0) AS CHAR)", "COALESCE(NULLIF(TRIM(oi.region), ''), '未设置区县')", "区县级", false
	default:
		return "CAST(IFNULL(oi.province_code, 0) AS CHAR)", "COALESCE(NULLIF(TRIM(oi.province), ''), '未设置省份')", "--", true
	}
}

func governmentScopeDisplayName(scope governmentOverviewScope) string {
	switch normalizeGovernmentLevel(scope.ScopeLevel) {
	case "province":
		return strings.TrimSpace(scope.ProvinceName)
	case "city":
		return strings.Trim(strings.Join([]string{strings.TrimSpace(scope.ProvinceName), strings.TrimSpace(scope.CityName)}, "-"), "-")
	case "district":
		return strings.Trim(strings.Join([]string{strings.TrimSpace(scope.ProvinceName), strings.TrimSpace(scope.CityName), strings.TrimSpace(scope.DistrictName)}, "-"), "-")
	default:
		if strings.TrimSpace(scope.DistrictName) != "" {
			return strings.Trim(strings.Join([]string{strings.TrimSpace(scope.ProvinceName), strings.TrimSpace(scope.CityName), strings.TrimSpace(scope.DistrictName)}, "-"), "-")
		}
		if strings.TrimSpace(scope.CityName) != "" {
			return strings.Trim(strings.Join([]string{strings.TrimSpace(scope.ProvinceName), strings.TrimSpace(scope.CityName)}, "-"), "-")
		}
		return strings.TrimSpace(scope.ProvinceName)
	}
}

func governmentScopeCodeText(scope governmentOverviewScope, fallbackLevel string) string {
	switch normalizeGovernmentLevel(firstNonEmptyGovernmentString(scope.ScopeLevel, fallbackLevel)) {
	case "province":
		return strings.TrimSpace(scope.ProvinceCode)
	case "city":
		return strings.TrimSpace(scope.CityCode)
	case "district":
		return strings.TrimSpace(scope.DistrictCode)
	default:
		return ""
	}
}

func firstNonEmptyGovernmentString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func summarizeGovernmentScopes(level string, scopes []governmentOverviewScope) (string, string, int) {
	if normalizeGovernmentLevel(level) == "super" {
		return "全部辖区", "全域", 0
	}

	names := make([]string, 0, len(scopes))
	seen := make(map[string]struct{}, len(scopes))
	codeText := ""
	for _, scope := range scopes {
		displayName := strings.TrimSpace(governmentScopeDisplayName(scope))
		if displayName == "" {
			continue
		}
		if _, ok := seen[displayName]; ok {
			continue
		}
		seen[displayName] = struct{}{}
		names = append(names, displayName)
		if codeText == "" {
			codeText = governmentScopeCodeText(scope, level)
		}
	}

	switch len(names) {
	case 0:
		return "未配置辖区", "--", 0
	case 1:
		if codeText == "" {
			codeText = "--"
		}
		return names[0], codeText, 1
	default:
		preview := strings.Join(names[:minInt(len(names), 2)], "、")
		if len(names) > 2 {
			preview = fmt.Sprintf("%s 等 %d 个区域", preview, len(names))
		}
		return preview, "多区域", len(names)
	}
}

func buildGovernmentInstitutionWhereClause(alias string, level string, scopes []governmentOverviewScope) (string, []any) {
	filters := []string{alias + ".del_flag = 0"}
	normalizedLevel := normalizeGovernmentLevel(level)
	if normalizedLevel == "super" {
		return strings.Join(filters, " AND "), nil
	}

	conditions := make([]string, 0, len(scopes))
	args := make([]any, 0, len(scopes)*3)
	for _, scope := range scopes {
		effectiveLevel := normalizeGovernmentLevel(firstNonEmptyGovernmentString(scope.ScopeLevel, normalizedLevel))
		switch effectiveLevel {
		case "province":
			if strings.TrimSpace(scope.ProvinceCode) == "" {
				continue
			}
			conditions = append(conditions, "IFNULL("+alias+".province_code, 0) = ?")
			args = append(args, scope.ProvinceCode)
		case "city":
			if strings.TrimSpace(scope.ProvinceCode) == "" || strings.TrimSpace(scope.CityCode) == "" {
				continue
			}
			conditions = append(conditions, "(IFNULL("+alias+".province_code, 0) = ? AND IFNULL("+alias+".city_code, 0) = ?)")
			args = append(args, scope.ProvinceCode, scope.CityCode)
		case "district":
			if strings.TrimSpace(scope.ProvinceCode) == "" || strings.TrimSpace(scope.CityCode) == "" || strings.TrimSpace(scope.DistrictCode) == "" {
				continue
			}
			conditions = append(conditions, "(IFNULL("+alias+".province_code, 0) = ? AND IFNULL("+alias+".city_code, 0) = ? AND IFNULL("+alias+".region_code, 0) = ?)")
			args = append(args, scope.ProvinceCode, scope.CityCode, scope.DistrictCode)
		}
	}

	if len(conditions) == 0 {
		filters = append(filters, "1 = 0")
		return strings.Join(filters, " AND "), args
	}
	filters = append(filters, "("+strings.Join(conditions, " OR ")+")")
	return strings.Join(filters, " AND "), args
}

func buildGovernmentInstitutionListWhereClause(baseWhere string, baseArgs []any, keyword string, status, openType *int) (string, []any) {
	filters := []string{baseWhere}
	args := append([]any{}, baseArgs...)
	statusExpr := institutionStatusExpr("oi")

	if trimmed := strings.TrimSpace(keyword); trimmed != "" {
		like := "%" + trimmed + "%"
		filters = append(filters, `(CAST(oi.id AS CHAR) LIKE ? OR oi.organ_name LIKE ? OR oi.organ_code LIKE ? OR oi.principal LIKE ? OR oi.mobile LIKE ? OR oi.province LIKE ? OR oi.city LIKE ? OR oi.region LIKE ?)`)
		args = append(args, like, like, like, like, like, like, like, like)
	}

	if status != nil && *status > 0 {
		filters = append(filters, statusExpr+" = ?")
		args = append(args, *status)
	}

	if openType != nil && *openType > 0 {
		filters = append(filters, "IFNULL(oi.open_type, 2) = ?")
		args = append(args, *openType)
	}

	return strings.Join(filters, " AND "), args
}

func minInt(left, right int) int {
	if left < right {
		return left
	}
	return right
}

func New(db *sql.DB) (*Repository, error) {
	repo := &Repository{db: db}
	if err := repo.ensureTenantControlPlaneSchema(context.Background()); err != nil {
		return nil, err
	}
	if err := tenantstorage.EnsureSchema(context.Background(), db); err != nil {
		return nil, err
	}
	if err := repo.migrateLegacyInstitutionMenuCodes(context.Background()); err != nil {
		return nil, err
	}
	if err := repo.ensureInstitutionSchema(context.Background()); err != nil {
		return nil, err
	}
	if err := repo.ensureInstConfigHomeSchoolSchema(context.Background()); err != nil {
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
	if err := repo.migrateInstitutionSuperAdminRole(context.Background()); err != nil {
		return nil, err
	}
	return repo, nil
}

func (repo *Repository) ensureTenantControlPlaneSchema(ctx context.Context) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS tenant_profile (
			id BIGINT NOT NULL AUTO_INCREMENT,
			tenant_id VARCHAR(64) NOT NULL,
			tenant_name VARCHAR(128) NOT NULL,
			tenant_type VARCHAR(32) NOT NULL DEFAULT 'partner',
			parent_tenant_id VARCHAR(64) DEFAULT NULL,
			edition VARCHAR(64) NOT NULL DEFAULT 'enterprise',
			status VARCHAR(32) NOT NULL DEFAULT 'active',
			isolation_mode VARCHAR(32) NOT NULL DEFAULT 'shared_db',
			brand_config JSON DEFAULT NULL,
			remark VARCHAR(500) DEFAULT NULL,
			create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			PRIMARY KEY (id),
			UNIQUE KEY uk_tenant_profile_tenant (tenant_id),
			KEY idx_tenant_profile_parent (parent_tenant_id, del_flag)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS tenant_domain (
			id BIGINT NOT NULL AUTO_INCREMENT,
			tenant_id VARCHAR(64) NOT NULL,
			domain VARCHAR(255) NOT NULL,
			entry_type VARCHAR(32) NOT NULL DEFAULT 'institution-admin',
			is_primary TINYINT(1) NOT NULL DEFAULT 0,
			create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			PRIMARY KEY (id),
			UNIQUE KEY uk_tenant_domain_domain (domain),
			KEY idx_tenant_domain_tenant (tenant_id, del_flag)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS tenant_institution (
			id BIGINT NOT NULL AUTO_INCREMENT,
			tenant_id VARCHAR(64) NOT NULL,
			institution_id BIGINT NOT NULL,
			create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			PRIMARY KEY (id),
			UNIQUE KEY uk_tenant_institution_inst (institution_id),
			KEY idx_tenant_institution_tenant (tenant_id, del_flag)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS tenant_menu (
			id BIGINT NOT NULL AUTO_INCREMENT,
			tenant_id VARCHAR(64) NOT NULL,
			menu_id BIGINT NOT NULL,
			create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			PRIMARY KEY (id),
			UNIQUE KEY uk_tenant_menu (tenant_id, menu_id),
			KEY idx_tenant_menu_tenant (tenant_id, del_flag)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS tenant_module (
			id BIGINT NOT NULL AUTO_INCREMENT,
			tenant_id VARCHAR(64) NOT NULL,
			module_id BIGINT NOT NULL,
			create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			PRIMARY KEY (id),
			UNIQUE KEY uk_tenant_module (tenant_id, module_id),
			KEY idx_tenant_module_tenant (tenant_id, del_flag)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS tenant_user (
			id BIGINT NOT NULL AUTO_INCREMENT,
			tenant_id VARCHAR(64) NOT NULL,
			user_id BIGINT NOT NULL,
			user_role VARCHAR(32) NOT NULL DEFAULT 'tenant_admin',
			create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			PRIMARY KEY (id),
			UNIQUE KEY uk_tenant_user (tenant_id, user_id),
			KEY idx_tenant_user_tenant (tenant_id, del_flag)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	}
	for _, statement := range statements {
		if _, err := repo.db.ExecContext(ctx, statement); err != nil {
			return err
		}
	}
	if err := repo.ensureColumnExists(ctx, "tenant_profile", "brand_config", `
		ALTER TABLE tenant_profile ADD COLUMN brand_config JSON DEFAULT NULL AFTER isolation_mode
	`); err != nil {
		return err
	}
	return repo.seedTenantAControlPlane(ctx)
}

func (repo *Repository) seedTenantAControlPlane(ctx context.Context) error {
	if _, err := repo.db.ExecContext(ctx, `
		INSERT INTO tenant_profile (tenant_id, tenant_name, tenant_type, parent_tenant_id, edition, status, isolation_mode, remark)
		SELECT 'platform', '公司平台总控', 'platform', NULL, 'platform', 'active', 'shared_db', '公司最高权限控制台'
		FROM DUAL
		WHERE NOT EXISTS (SELECT 1 FROM tenant_profile WHERE tenant_id = 'platform' AND del_flag = 0)
	`); err != nil {
		return err
	}
	if _, err := repo.db.ExecContext(ctx, `
		INSERT INTO tenant_profile (tenant_id, tenant_name, tenant_type, parent_tenant_id, edition, status, isolation_mode, remark)
		SELECT 'tenant-a', 'A租户', 'partner', 'platform', 'enterprise', 'active', 'shared_db', '默认迁移租户'
		FROM DUAL
		WHERE NOT EXISTS (SELECT 1 FROM tenant_profile WHERE tenant_id = 'tenant-a' AND del_flag = 0)
	`); err != nil {
		return err
	}
	if _, err := repo.db.ExecContext(ctx, `
		INSERT INTO tenant_institution (tenant_id, institution_id, create_time, update_time, del_flag)
		SELECT 'tenant-a', oi.id, NOW(), NOW(), 0
		FROM org_institution oi
		WHERE oi.del_flag = 0
		  AND NOT EXISTS (SELECT 1 FROM tenant_institution WHERE del_flag = 0)
		ON DUPLICATE KEY UPDATE tenant_id = VALUES(tenant_id), update_time = NOW(), del_flag = 0
	`); err != nil {
		return err
	}
	if _, err := repo.db.ExecContext(ctx, `
		INSERT INTO tenant_menu (tenant_id, menu_id, create_time, update_time, del_flag)
		SELECT 'tenant-a', sm.id, NOW(), NOW(), 0
		FROM sso_menu sm
		WHERE sm.del_flag = 0
		  AND NOT EXISTS (SELECT 1 FROM tenant_menu WHERE tenant_id = 'tenant-a' AND del_flag = 0)
		ON DUPLICATE KEY UPDATE update_time = NOW(), del_flag = 0
	`); err != nil {
		return err
	}
	if _, err := repo.db.ExecContext(ctx, `
		INSERT INTO tenant_module (tenant_id, module_id, create_time, update_time, del_flag)
		SELECT 'tenant-a', sm.id, NOW(), NOW(), 0
		FROM sys_module sm
		WHERE sm.del_flag = 0
		  AND NOT EXISTS (SELECT 1 FROM tenant_module WHERE tenant_id = 'tenant-a' AND del_flag = 0)
		ON DUPLICATE KEY UPDATE update_time = NOW(), del_flag = 0
	`); err != nil {
		return err
	}
	if _, err := repo.db.ExecContext(ctx, `
		INSERT INTO tenant_user (tenant_id, user_id, user_role)
		SELECT 'platform', su.id, 'platform_admin'
		FROM sso_user su
		WHERE su.del_flag = 0 AND su.username = 'admin'
		ON DUPLICATE KEY UPDATE user_role = 'platform_admin', update_time = NOW(), del_flag = 0
	`); err != nil {
		return err
	}
	if _, err := repo.db.ExecContext(ctx, `
		UPDATE tenant_user tu
		JOIN sso_user su ON su.id = tu.user_id AND su.del_flag = 0
		SET tu.del_flag = 1, tu.update_time = NOW()
		WHERE tu.tenant_id = 'tenant-a'
		  AND su.username = 'admin'
		  AND tu.user_role = 'platform_admin'
	`); err != nil {
		return err
	}
	if _, err := repo.db.ExecContext(ctx, `
		UPDATE sso_user su
		JOIN tenant_user tu ON tu.user_id = su.id AND tu.del_flag = 0
		SET su.is_admin = 0, su.update_time = NOW()
		WHERE tu.user_role = 'tenant_admin'
		  AND su.del_flag = 0
	`); err != nil {
		return err
	}
	_, err := repo.db.ExecContext(ctx, `
		INSERT INTO tenant_domain (tenant_id, domain, entry_type, is_primary)
		SELECT 'tenant-a', 'tenant-a.localhost', 'institution-admin', 1
		FROM DUAL
		WHERE NOT EXISTS (SELECT 1 FROM tenant_domain WHERE domain = 'tenant-a.localhost' AND del_flag = 0)
	`)
	return err
}

func (repo *Repository) migrateInstitutionSuperAdminRole(ctx context.Context) error {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO sso_role (uuid, version, role_name, description, org_id, role_type, is_admin, is_default, del_flag, create_time, update_time)
		SELECT UUID(), 0, ?, ?, oi.id, 2, 1, 0, 0, NOW(), NOW()
		FROM org_institution oi
		WHERE oi.del_flag = 0
		  AND NOT EXISTS (
		    SELECT 1
		    FROM sso_role sr
		    WHERE sr.org_id = oi.id
		      AND sr.role_type = 2
		      AND sr.is_admin = 1
		      AND sr.del_flag = 0
		  )
	`, defaultInstitutionAdminRoleName, defaultInstitutionAdminRoleDescription); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE sso_role
		SET role_name = ?, description = ?, update_time = NOW()
		WHERE del_flag = 0
		  AND role_type = 2
		  AND is_admin = 1
		  AND (
		    TRIM(REPLACE(REPLACE(IFNULL(role_name, ''), '\r', ''), '\n', '')) <> ?
		    OR TRIM(REPLACE(REPLACE(IFNULL(description, ''), '\r', ''), '\n', '')) <> ?
		  )
	`, defaultInstitutionAdminRoleName, defaultInstitutionAdminRoleDescription, defaultInstitutionAdminRoleName, defaultInstitutionAdminRoleDescription); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO sso_user_role (user_id, role_id)
		SELECT iu.user_id, sr.id
		FROM inst_user iu
		JOIN sso_role sr
		  ON sr.org_id = iu.inst_id
		 AND sr.role_type = 2
		 AND sr.is_admin = 1
		 AND sr.del_flag = 0
		WHERE iu.del_flag = 0
		  AND iu.is_admin = 1
		  AND NOT EXISTS (
		    SELECT 1
		    FROM sso_user_role sur
		    WHERE sur.user_id = iu.user_id
		      AND sur.role_id = sr.id
		  )
	`); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `
		DELETE sur
		FROM sso_user_role sur
		JOIN sso_role sr ON sr.id = sur.role_id
		JOIN inst_user iu ON iu.user_id = sur.user_id
		WHERE iu.del_flag = 0
		  AND iu.is_admin = 1
		  AND sr.del_flag = 0
		  AND sr.role_type = 2
		  AND IFNULL(sr.is_admin, 0) = 0
		  AND TRIM(REPLACE(REPLACE(IFNULL(sr.role_name, ''), '\r', ''), '\n', '')) = '校区管理员'
		  AND (sr.org_id = 0 OR sr.org_id = iu.inst_id)
	`); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO sso_role_menu (role_id, menu_id)
		SELECT sr.id, smm.menu_id
		FROM sso_role sr
		JOIN org_module om
		  ON om.org_id = sr.org_id
		 AND om.del_flag = 0
		JOIN sys_module_menu smm
		  ON smm.module_id = om.module_id
		 AND smm.del_flag = 0
		JOIN sso_menu m
		  ON m.id = smm.menu_id
		 AND m.del_flag = 0
		 AND m.own_type = 2
		WHERE sr.del_flag = 0
		  AND sr.role_type = 2
		  AND sr.is_admin = 1
		  AND NOT EXISTS (
		    SELECT 1
		    FROM sso_role_menu rm
		    WHERE rm.role_id = sr.id
		      AND rm.menu_id = smm.menu_id
		  )
	`); err != nil {
		return err
	}

	return tx.Commit()
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

func (repo *Repository) ensureInstConfigHomeSchoolSchema(ctx context.Context) error {
	columns := map[string]string{
		"auto_send_birthday_message":                 "ALTER TABLE inst_config ADD COLUMN auto_send_birthday_message TINYINT(1) NOT NULL DEFAULT 0",
		"enable_recharge_account_change_message":     "ALTER TABLE inst_config ADD COLUMN enable_recharge_account_change_message TINYINT(1) NOT NULL DEFAULT 0",
		"enabled_class_reminder":                     "ALTER TABLE inst_config ADD COLUMN enabled_class_reminder TINYINT(1) NOT NULL DEFAULT 0",
		"enabled_class_consumption_reminder":         "ALTER TABLE inst_config ADD COLUMN enabled_class_consumption_reminder TINYINT(1) NOT NULL DEFAULT 0",
		"enable_audition_sms_remind":                 "ALTER TABLE inst_config ADD COLUMN enable_audition_sms_remind TINYINT(1) NOT NULL DEFAULT 0",
		"enable_send_coupon_remind_sms":              "ALTER TABLE inst_config ADD COLUMN enable_send_coupon_remind_sms TINYINT(1) NOT NULL DEFAULT 0",
		"enable_send_child_bind_notice_to_admin":     "ALTER TABLE inst_config ADD COLUMN enable_send_child_bind_notice_to_admin TINYINT(1) NOT NULL DEFAULT 0",
		"enable_teaching_bill_remind_sms":            "ALTER TABLE inst_config ADD COLUMN enable_teaching_bill_remind_sms TINYINT(1) NOT NULL DEFAULT 0",
		"student_absent_class_switch":                "ALTER TABLE inst_config ADD COLUMN student_absent_class_switch TINYINT(1) NOT NULL DEFAULT 0",
		"enabled_renew_reminder":                     "ALTER TABLE inst_config ADD COLUMN enabled_renew_reminder TINYINT(1) NOT NULL DEFAULT 0",
		"enable_arrearaged_send_message":             "ALTER TABLE inst_config ADD COLUMN enable_arrearaged_send_message TINYINT(1) NOT NULL DEFAULT 0",
		"enable_liquidation_remind_message":          "ALTER TABLE inst_config ADD COLUMN enable_liquidation_remind_message TINYINT(1) NOT NULL DEFAULT 0",
		"enable_point_change_remind_message":         "ALTER TABLE inst_config ADD COLUMN enable_point_change_remind_message TINYINT(1) NOT NULL DEFAULT 0",
		"enable_org_send_child_bind_notice_to_admin": "ALTER TABLE inst_config ADD COLUMN enable_org_send_child_bind_notice_to_admin TINYINT(1) NOT NULL DEFAULT 0",
		"send_class_reminder_msg_hour":               "ALTER TABLE inst_config ADD COLUMN send_class_reminder_msg_hour VARCHAR(16) NOT NULL DEFAULT '19:00'",
		"enable_leave_apply_number_limit":            "ALTER TABLE inst_config ADD COLUMN enable_leave_apply_number_limit TINYINT(1) NOT NULL DEFAULT 0",
		"leave_apply_cycle_limit":                    "ALTER TABLE inst_config ADD COLUMN leave_apply_cycle_limit VARCHAR(32) NOT NULL DEFAULT 'month'",
		"leave_apply_number_limit":                   "ALTER TABLE inst_config ADD COLUMN leave_apply_number_limit VARCHAR(32) NOT NULL DEFAULT '2'",
		"leave_apply_type_limit":                     "ALTER TABLE inst_config ADD COLUMN leave_apply_type_limit VARCHAR(32) NOT NULL DEFAULT 'course'",
		"enable_leave_apply_time_limit":              "ALTER TABLE inst_config ADD COLUMN enable_leave_apply_time_limit TINYINT(1) NOT NULL DEFAULT 0",
		"leave_apply_time_limit":                     "ALTER TABLE inst_config ADD COLUMN leave_apply_time_limit VARCHAR(32) NOT NULL DEFAULT '1.0'",
		"enable_renew_class_num":                     "ALTER TABLE inst_config ADD COLUMN enable_renew_class_num TINYINT(1) NOT NULL DEFAULT 0",
		"renew_class_num":                            "ALTER TABLE inst_config ADD COLUMN renew_class_num VARCHAR(32) NOT NULL DEFAULT '5'",
		"enable_renew_validity_day":                  "ALTER TABLE inst_config ADD COLUMN enable_renew_validity_day TINYINT(1) NOT NULL DEFAULT 0",
		"renew_validity_day":                         "ALTER TABLE inst_config ADD COLUMN renew_validity_day VARCHAR(32) NOT NULL DEFAULT '15'",
		"enable_renew_price":                         "ALTER TABLE inst_config ADD COLUMN enable_renew_price TINYINT(1) NOT NULL DEFAULT 0",
		"renew_price":                                "ALTER TABLE inst_config ADD COLUMN renew_price VARCHAR(32) NOT NULL DEFAULT '500'",
	}
	for columnName, ddl := range columns {
		if err := repo.ensureColumnExists(ctx, "inst_config", columnName, ddl); err != nil {
			return err
		}
	}
	if err := repo.ensureInstConfigStringFieldTypes(ctx); err != nil {
		return err
	}
	return nil
}

func (repo *Repository) ensureInstConfigStringFieldTypes(ctx context.Context) error {
	type fieldSpec struct {
		Column       string
		Definition   string
		NumericValue string
		StringValue  string
	}
	fields := []fieldSpec{
		{Column: "send_class_reminder_msg_hour", Definition: "VARCHAR(16) NOT NULL DEFAULT '19:00'", NumericValue: "19", StringValue: "19:00"},
		{Column: "face_attendance_interval", Definition: "VARCHAR(32) NOT NULL DEFAULT '1'", NumericValue: "1", StringValue: "1"},
		{Column: "leave_apply_cycle_limit", Definition: "VARCHAR(32) NOT NULL DEFAULT 'month'", NumericValue: "1", StringValue: "month"},
		{Column: "leave_apply_number_limit", Definition: "VARCHAR(32) NOT NULL DEFAULT '2'", NumericValue: "2", StringValue: "2"},
		{Column: "leave_apply_type_limit", Definition: "VARCHAR(32) NOT NULL DEFAULT 'course'", NumericValue: "1", StringValue: "course"},
		{Column: "leave_apply_time_limit", Definition: "VARCHAR(32) NOT NULL DEFAULT '1.0'", NumericValue: "1", StringValue: "1.0"},
		{Column: "renew_class_num", Definition: "VARCHAR(32) NOT NULL DEFAULT '5'", NumericValue: "5", StringValue: "5"},
		{Column: "renew_validity_day", Definition: "VARCHAR(32) NOT NULL DEFAULT '15'", NumericValue: "15", StringValue: "15"},
		{Column: "renew_price", Definition: "VARCHAR(32) NOT NULL DEFAULT '500'", NumericValue: "500", StringValue: "500"},
	}
	for _, field := range fields {
		if _, err := repo.db.ExecContext(ctx, fmt.Sprintf("UPDATE inst_config SET %s = %s WHERE %s IS NULL", field.Column, field.NumericValue, field.Column)); err != nil {
			return err
		}
		if err := repo.ensureColumnType(ctx, "inst_config", field.Column, "varchar", fmt.Sprintf("ALTER TABLE inst_config MODIFY COLUMN %s %s", field.Column, field.Definition)); err != nil {
			return err
		}
		if _, err := repo.db.ExecContext(ctx, fmt.Sprintf("UPDATE inst_config SET %s = ? WHERE TRIM(%s) = ''", field.Column, field.Column), field.StringValue); err != nil {
			return err
		}
	}
	if _, err := repo.db.ExecContext(ctx, `
		UPDATE inst_config
		SET send_class_reminder_msg_hour = CONCAT(LPAD(send_class_reminder_msg_hour, 2, '0'), ':00')
		WHERE send_class_reminder_msg_hour REGEXP '^[0-9]{1,2}$'
	`); err != nil {
		return err
	}
	_, err := repo.db.ExecContext(ctx, `
		UPDATE inst_config
		SET leave_apply_cycle_limit = 'month'
		WHERE leave_apply_cycle_limit REGEXP '^[0-9]+$'
	`)
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
			login_slug VARCHAR(64) NOT NULL DEFAULT '',
			login_brand_config JSON DEFAULT NULL,
			create_id BIGINT DEFAULT NULL,
			create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
			update_id BIGINT DEFAULT NULL,
			update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			del_flag TINYINT(1) DEFAULT 0,
			PRIMARY KEY (id),
			UNIQUE KEY uk_org_institution_profile_inst (institution_id),
			KEY idx_org_institution_profile_inst_del (institution_id, del_flag),
			KEY idx_org_institution_profile_login_slug (login_slug, del_flag)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
	`); err != nil {
		return err
	}
	if err := repo.ensureColumnExists(ctx, "org_institution_profile", "login_slug", `
		ALTER TABLE org_institution_profile ADD COLUMN login_slug VARCHAR(64) NOT NULL DEFAULT '' AFTER gallery_images
	`); err != nil {
		return err
	}
	if err := repo.ensureColumnExists(ctx, "org_institution_profile", "login_brand_config", `
		ALTER TABLE org_institution_profile ADD COLUMN login_brand_config JSON DEFAULT NULL AFTER login_slug
	`); err != nil {
		return err
	}
	if _, err := repo.db.ExecContext(ctx, `
		CREATE INDEX idx_org_institution_profile_login_slug ON org_institution_profile (login_slug, del_flag)
	`); err != nil && !strings.Contains(strings.ToLower(err.Error()), "duplicate") {
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
	if err := repo.ensureColumnExists(ctx, "sys_module", "tenant_id", `
		ALTER TABLE sys_module
		ADD COLUMN tenant_id VARCHAR(64) NOT NULL DEFAULT 'platform' COMMENT '版本归属租户'
		AFTER id
	`); err != nil {
		return err
	}
	if err := repo.ensureColumnExists(ctx, "sys_module", "owner_type", `
		ALTER TABLE sys_module
		ADD COLUMN owner_type VARCHAR(32) NOT NULL DEFAULT 'platform_template' COMMENT '版本类型：平台模板/租户售卖版本'
		AFTER tenant_id
	`); err != nil {
		return err
	}
	if err := repo.ensureColumnExists(ctx, "sys_module", "source_module_id", `
		ALTER TABLE sys_module
		ADD COLUMN source_module_id BIGINT DEFAULT NULL COMMENT '来源平台模板ID'
		AFTER owner_type
	`); err != nil {
		return err
	}
	if _, err := repo.db.ExecContext(ctx, `
		UPDATE sys_module
		SET tenant_id = 'platform', owner_type = 'platform_template'
		WHERE del_flag = 0
		  AND (tenant_id IS NULL OR tenant_id = '' OR owner_type IS NULL OR owner_type = '')
	`); err != nil {
		return err
	}
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
			INSERT INTO sys_module (uuid, version, tenant_id, owner_type, name, type, price, create_time, update_time, del_flag, remark)
			VALUES (?, 0, 'platform', 'platform_template', ?, 1, ?, NOW(), NOW(), 0, '')
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
		LoginSlug:     normalizeLoginSlug(profile.LoginSlug),
		LoginBrand:    trimLoginBrandConfig(profile.LoginBrand),
	}
}

func normalizeLoginSlug(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, " ", "-")
	var builder strings.Builder
	lastDash := false
	for _, char := range value {
		valid := (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9')
		if valid {
			builder.WriteRune(char)
			lastDash = false
			continue
		}
		if (char == '-' || char == '_') && !lastDash {
			builder.WriteRune('-')
			lastDash = true
		}
	}
	return strings.Trim(builder.String(), "-")
}

func trimLoginBrandConfig(input model.TenantLoginBrandConfig) model.TenantLoginBrandConfig {
	return model.TenantLoginBrandConfig{
		Template:        strings.TrimSpace(input.Template),
		BrandName:       strings.TrimSpace(input.BrandName),
		LogoURL:         strings.TrimSpace(input.LogoURL),
		LoginTitle:      strings.TrimSpace(input.LoginTitle),
		LoginSubtitle:   strings.TrimSpace(input.LoginSubtitle),
		BackgroundURL:   strings.TrimSpace(input.BackgroundURL),
		PrimaryColor:    strings.TrimSpace(input.PrimaryColor),
		Copyright:       strings.TrimSpace(input.Copyright),
		HeroBadge:       strings.TrimSpace(input.HeroBadge),
		HeroTitle:       strings.TrimSpace(input.HeroTitle),
		HeroDescription: strings.TrimSpace(input.HeroDescription),
	}
}

func marshalLoginBrandConfig(input model.TenantLoginBrandConfig) string {
	normalized := trimLoginBrandConfig(input)
	if normalized == (model.TenantLoginBrandConfig{}) {
		return "{}"
	}
	payload, err := json.Marshal(normalized)
	if err != nil {
		return "{}"
	}
	return string(payload)
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
	args := make([]any, 0, 3)
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

func (repo *Repository) GetTenantLoginTheme(ctx context.Context, domain, entryType string) (model.TenantPublicLoginTheme, error) {
	domain = strings.ToLower(strings.TrimSpace(domain))
	entryType = firstNonEmpty(strings.TrimSpace(entryType), "platform-admin")

	if domain != "" && entryType == "institution-admin" {
		if result, ok, err := repo.getInstitutionWildcardLoginTheme(ctx, domain, entryType); err != nil {
			return model.TenantPublicLoginTheme{}, err
		} else if ok {
			return result, nil
		}
	}

	query := `
		SELECT tp.tenant_id, tp.tenant_name, IFNULL(CAST(tp.brand_config AS CHAR), '')
		FROM tenant_profile tp
	`
	args := make([]any, 0, 2)
	matchedBy := "default"
	if domain != "" {
		query += `
			JOIN tenant_domain td ON td.tenant_id = tp.tenant_id AND td.del_flag = 0
			WHERE td.domain = ? AND td.entry_type = ? AND tp.del_flag = 0 AND tp.status = 'active'
			ORDER BY td.is_primary DESC, td.id ASC
			LIMIT 1
		`
		args = append(args, domain, entryType)
		matchedBy = "domain"
	} else {
		query += `
			WHERE tp.tenant_id = 'platform' AND tp.del_flag = 0
			LIMIT 1
		`
	}

	var result model.TenantPublicLoginTheme
	var brandConfigRaw string
	err := repo.db.QueryRowContext(ctx, query, args...).Scan(&result.TenantID, &result.TenantName, &brandConfigRaw)
	if err == sql.ErrNoRows && domain != "" {
		err = repo.db.QueryRowContext(ctx, `
			SELECT tenant_id, tenant_name, IFNULL(CAST(brand_config AS CHAR), '')
			FROM tenant_profile
			WHERE tenant_id = 'platform' AND del_flag = 0
			LIMIT 1
		`).Scan(&result.TenantID, &result.TenantName, &brandConfigRaw)
		matchedBy = "fallback"
	}
	if err != nil {
		return model.TenantPublicLoginTheme{}, err
	}
	result.EntryType = entryType
	result.MatchedBy = matchedBy
	brandSet := parseLoginBrandSet(brandConfigRaw)
	result.LoginBrand = selectLoginBrandConfig(brandSet, result.EntryType, result.TenantName)
	return result, nil
}

func (repo *Repository) getInstitutionWildcardLoginTheme(ctx context.Context, domain, entryType string) (model.TenantPublicLoginTheme, bool, error) {
	var tenantID string
	var tenantName string
	var baseDomain string
	var brandConfigRaw string
	err := repo.db.QueryRowContext(ctx, `
		SELECT tp.tenant_id, tp.tenant_name, td.domain, IFNULL(CAST(tp.brand_config AS CHAR), '')
		FROM tenant_domain td
		JOIN tenant_profile tp ON tp.tenant_id = td.tenant_id AND tp.del_flag = 0 AND tp.status = 'active'
		WHERE td.entry_type = ?
		  AND td.del_flag = 0
		  AND ? LIKE CONCAT('%.', td.domain)
		ORDER BY CHAR_LENGTH(td.domain) DESC, td.is_primary DESC, td.id ASC
		LIMIT 1
	`, entryType, domain).Scan(&tenantID, &tenantName, &baseDomain, &brandConfigRaw)
	if err == sql.ErrNoRows {
		return model.TenantPublicLoginTheme{}, false, nil
	}
	if err != nil {
		return model.TenantPublicLoginTheme{}, false, err
	}

	prefix := strings.TrimSuffix(domain, "."+baseDomain)
	if prefix == "" || strings.Contains(prefix, ".") {
		return model.TenantPublicLoginTheme{}, false, nil
	}

	brandSet := parseLoginBrandSet(brandConfigRaw)
	tenantBrand := selectLoginBrandConfig(brandSet, entryType, tenantName)
	result := model.TenantPublicLoginTheme{
		TenantID:   tenantID,
		TenantName: tenantName,
		EntryType:  entryType,
		LoginBrand: tenantBrand,
		MatchedBy:  "tenant-wildcard",
	}

	var institutionBrandRaw string
	err = repo.db.QueryRowContext(ctx, `
		SELECT oi.id, IFNULL(oi.organ_name, ''), IFNULL(CAST(oip.login_brand_config AS CHAR), '')
		FROM org_institution_profile oip
		JOIN org_institution oi ON oi.id = oip.institution_id AND oi.del_flag = 0 AND IFNULL(oi.enabled, 0) = 1
		JOIN tenant_institution ti ON ti.institution_id = oi.id AND ti.tenant_id = ? AND ti.del_flag = 0
		WHERE oip.del_flag = 0 AND oip.login_slug = ?
		LIMIT 1
	`, tenantID, prefix).Scan(&result.InstitutionID, &result.InstitutionName, &institutionBrandRaw)
	if err == sql.ErrNoRows {
		return result, true, nil
	}
	if err != nil {
		return model.TenantPublicLoginTheme{}, false, err
	}
	result.LoginBrand = mergeLoginBrandConfig(tenantBrand, parseLoginBrandConfig(institutionBrandRaw), result.InstitutionName)
	if strings.TrimSpace(result.LoginBrand.HeroBadge) == "" || strings.TrimSpace(result.LoginBrand.HeroBadge) == strings.TrimSpace(tenantBrand.HeroBadge) {
		result.LoginBrand.HeroBadge = buildInstitutionAllianceBadge(tenantName, result.InstitutionName)
	}
	result.MatchedBy = "institution-wildcard"
	return result, true, nil
}

func buildInstitutionAllianceBadge(tenantName, institutionName string) string {
	tenantName = strings.TrimSpace(tenantName)
	institutionName = strings.TrimSpace(institutionName)
	if tenantName == "" {
		return institutionName
	}
	if institutionName == "" {
		return tenantName
	}
	if strings.HasSuffix(tenantName, "联盟") {
		return tenantName + " X " + institutionName
	}
	return tenantName + "联盟 X " + institutionName
}

func mergeLoginBrandConfig(base, override model.TenantLoginBrandConfig, fallbackName string) model.TenantLoginBrandConfig {
	merged := base
	if strings.TrimSpace(override.Template) != "" {
		merged.Template = strings.TrimSpace(override.Template)
	}
	if strings.TrimSpace(override.BrandName) != "" {
		merged.BrandName = strings.TrimSpace(override.BrandName)
	} else if strings.TrimSpace(fallbackName) != "" {
		merged.BrandName = strings.TrimSpace(fallbackName)
	}
	if strings.TrimSpace(override.LogoURL) != "" {
		merged.LogoURL = strings.TrimSpace(override.LogoURL)
	}
	if strings.TrimSpace(override.LoginTitle) != "" {
		merged.LoginTitle = strings.TrimSpace(override.LoginTitle)
	}
	if strings.TrimSpace(override.LoginSubtitle) != "" {
		merged.LoginSubtitle = strings.TrimSpace(override.LoginSubtitle)
	}
	if strings.TrimSpace(override.BackgroundURL) != "" {
		merged.BackgroundURL = strings.TrimSpace(override.BackgroundURL)
	}
	if strings.TrimSpace(override.PrimaryColor) != "" {
		merged.PrimaryColor = strings.TrimSpace(override.PrimaryColor)
	}
	if strings.TrimSpace(override.Copyright) != "" {
		merged.Copyright = strings.TrimSpace(override.Copyright)
	}
	if strings.TrimSpace(override.HeroBadge) != "" {
		merged.HeroBadge = strings.TrimSpace(override.HeroBadge)
	}
	if strings.TrimSpace(override.HeroTitle) != "" {
		merged.HeroTitle = strings.TrimSpace(override.HeroTitle)
	}
	if strings.TrimSpace(override.HeroDescription) != "" {
		merged.HeroDescription = strings.TrimSpace(override.HeroDescription)
	}
	return merged
}

func (repo *Repository) GetTenantBootstrapSummary(ctx context.Context, tenantID string) (model.TenantBootstrapSummary, error) {
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		tenantID = "tenant-a"
	}

	var summary model.TenantBootstrapSummary
	var brandConfigRaw string
	if err := repo.db.QueryRowContext(ctx, `
		SELECT tenant_id, tenant_name, tenant_type, edition, status, isolation_mode, IFNULL(CAST(brand_config AS CHAR), '')
		FROM tenant_profile
		WHERE tenant_id = ? AND del_flag = 0
		LIMIT 1
	`, tenantID).Scan(&summary.TenantID, &summary.TenantName, &summary.TenantType, &summary.Edition, &summary.Status, &summary.IsolationMode, &brandConfigRaw); err != nil {
		return model.TenantBootstrapSummary{}, err
	}
	brandSet := parseLoginBrandSet(brandConfigRaw)
	summary.PlatformLoginBrand = selectLoginBrandConfig(brandSet, "platform-admin", summary.TenantName)
	summary.InstitutionLoginBrand = selectLoginBrandConfig(brandSet, "institution-admin", summary.TenantName)
	summary.LoginBrand = summary.PlatformLoginBrand

	if err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(1)
		FROM tenant_institution
		WHERE tenant_id = ? AND del_flag = 0
	`, tenantID).Scan(&summary.InstitutionCount); err != nil {
		return model.TenantBootstrapSummary{}, err
	}
	institutionRows, err := repo.db.QueryContext(ctx, `
		SELECT institution_id
		FROM tenant_institution
		WHERE tenant_id = ? AND del_flag = 0
		ORDER BY institution_id ASC
	`, tenantID)
	if err != nil {
		return model.TenantBootstrapSummary{}, err
	}
	defer institutionRows.Close()
	for institutionRows.Next() {
		var institutionID int64
		if err := institutionRows.Scan(&institutionID); err != nil {
			return model.TenantBootstrapSummary{}, err
		}
		summary.InstitutionIDs = append(summary.InstitutionIDs, institutionID)
	}
	if err := institutionRows.Err(); err != nil {
		return model.TenantBootstrapSummary{}, err
	}
	if err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(1)
		FROM tenant_menu
		WHERE tenant_id = ? AND del_flag = 0
	`, tenantID).Scan(&summary.MenuCount); err != nil {
		return model.TenantBootstrapSummary{}, err
	}
	moduleRows, err := repo.db.QueryContext(ctx, `
		SELECT tm.module_id, IFNULL(sm.name, '')
		FROM tenant_module tm
		JOIN sys_module sm ON sm.id = tm.module_id AND sm.del_flag = 0
		WHERE tm.tenant_id = ? AND tm.del_flag = 0
		ORDER BY sm.id ASC
	`, tenantID)
	if err != nil {
		return model.TenantBootstrapSummary{}, err
	}
	defer moduleRows.Close()
	for moduleRows.Next() {
		var moduleID int64
		var moduleName string
		if err := moduleRows.Scan(&moduleID, &moduleName); err != nil {
			return model.TenantBootstrapSummary{}, err
		}
		summary.ModuleIDs = append(summary.ModuleIDs, moduleID)
		summary.ModuleNames = append(summary.ModuleNames, moduleName)
	}
	if err := moduleRows.Err(); err != nil {
		return model.TenantBootstrapSummary{}, err
	}
	summary.ModuleCount = len(summary.ModuleIDs)

	userRows, err := repo.db.QueryContext(ctx, `
		SELECT COALESCE(NULLIF(TRIM(su.username), ''), CAST(su.id AS CHAR))
		FROM tenant_user tu
		JOIN sso_user su ON su.id = tu.user_id AND su.del_flag = 0
		WHERE tu.tenant_id = ? AND tu.del_flag = 0
		ORDER BY tu.id
	`, tenantID)
	if err != nil {
		return model.TenantBootstrapSummary{}, err
	}
	defer userRows.Close()
	for userRows.Next() {
		var username string
		if err := userRows.Scan(&username); err != nil {
			return model.TenantBootstrapSummary{}, err
		}
		summary.AdminUsernames = append(summary.AdminUsernames, username)
	}
	if err := userRows.Err(); err != nil {
		return model.TenantBootstrapSummary{}, err
	}

	domainRows, err := repo.db.QueryContext(ctx, `
		SELECT domain, IFNULL(entry_type, '')
		FROM tenant_domain
		WHERE tenant_id = ? AND del_flag = 0
		ORDER BY is_primary DESC, id ASC
	`, tenantID)
	if err != nil {
		return model.TenantBootstrapSummary{}, err
	}
	defer domainRows.Close()
	for domainRows.Next() {
		var domain string
		var entryType string
		if err := domainRows.Scan(&domain, &entryType); err != nil {
			return model.TenantBootstrapSummary{}, err
		}
		summary.Domains = append(summary.Domains, domain)
		if entryType == "institution-admin" {
			summary.InstitutionDomains = append(summary.InstitutionDomains, domain)
		} else {
			summary.AdminDomains = append(summary.AdminDomains, domain)
		}
	}
	if err := domainRows.Err(); err != nil {
		return model.TenantBootstrapSummary{}, err
	}

	return summary, nil
}

func (repo *Repository) ListTenants(ctx context.Context, keyword string) ([]model.TenantListItem, error) {
	filters := []string{"tp.del_flag = 0"}
	args := make([]any, 0, 2)
	if trimmed := strings.TrimSpace(keyword); trimmed != "" {
		filters = append(filters, "(tp.tenant_id LIKE ? OR tp.tenant_name LIKE ?)")
		like := "%" + trimmed + "%"
		args = append(args, like, like)
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT tp.tenant_id,
		       tp.tenant_name,
		       tp.tenant_type,
		       tp.edition,
		       tp.status,
		       tp.isolation_mode,
		       IFNULL(CAST(tp.brand_config AS CHAR), '') AS brand_config,
		       IFNULL(ti.institution_count, 0) AS institution_count,
		       IFNULL(ti.institution_ids, '') AS institution_ids,
		       IFNULL(tm.menu_count, 0) AS menu_count,
		       IFNULL(tmod.module_count, 0) AS module_count,
		       IFNULL(tmod.module_ids, '') AS module_ids,
		       IFNULL(tmod.module_names, '') AS module_names,
		       IFNULL(tu.admins, '') AS admins,
		       IFNULL(td.domains, '') AS domains,
		       IFNULL(td.admin_domains, '') AS admin_domains,
		       IFNULL(td.institution_domains, '') AS institution_domains
		FROM tenant_profile tp
		LEFT JOIN (
			SELECT tenant_id,
			       COUNT(DISTINCT institution_id) AS institution_count,
			       GROUP_CONCAT(DISTINCT CAST(institution_id AS CHAR) ORDER BY institution_id ASC SEPARATOR ',') AS institution_ids
			FROM tenant_institution
			WHERE del_flag = 0
			GROUP BY tenant_id
		) ti ON ti.tenant_id = tp.tenant_id
		LEFT JOIN (
			SELECT tenant_id, COUNT(DISTINCT menu_id) AS menu_count
			FROM tenant_menu
			WHERE del_flag = 0
			GROUP BY tenant_id
		) tm ON tm.tenant_id = tp.tenant_id
		LEFT JOIN (
			SELECT tenant_module.tenant_id,
			       COUNT(DISTINCT tenant_module.module_id) AS module_count,
			       GROUP_CONCAT(DISTINCT CAST(tenant_module.module_id AS CHAR) ORDER BY sys_module.id ASC SEPARATOR ',') AS module_ids,
			       GROUP_CONCAT(DISTINCT sys_module.name ORDER BY sys_module.id ASC SEPARATOR ',') AS module_names
			FROM tenant_module
			LEFT JOIN sys_module ON sys_module.id = tenant_module.module_id AND sys_module.del_flag = 0
			WHERE tenant_module.del_flag = 0
			GROUP BY tenant_module.tenant_id
		) tmod ON tmod.tenant_id = tp.tenant_id
		LEFT JOIN (
			SELECT tenant_user.tenant_id,
			       GROUP_CONCAT(DISTINCT COALESCE(NULLIF(TRIM(sso_user.username), ''), CAST(sso_user.id AS CHAR)) ORDER BY sso_user.id SEPARATOR ',') AS admins
			FROM tenant_user
			LEFT JOIN sso_user ON sso_user.id = tenant_user.user_id AND sso_user.del_flag = 0
			WHERE tenant_user.del_flag = 0
			GROUP BY tenant_user.tenant_id
		) tu ON tu.tenant_id = tp.tenant_id
		LEFT JOIN (
			SELECT tenant_id,
			       GROUP_CONCAT(DISTINCT domain ORDER BY is_primary DESC, id ASC SEPARATOR ',') AS domains,
			       GROUP_CONCAT(DISTINCT CASE WHEN entry_type = 'platform-admin' THEN domain END ORDER BY is_primary DESC, id ASC SEPARATOR ',') AS admin_domains,
			       GROUP_CONCAT(DISTINCT CASE WHEN entry_type = 'institution-admin' THEN domain END ORDER BY is_primary DESC, id ASC SEPARATOR ',') AS institution_domains
			FROM tenant_domain
			WHERE del_flag = 0
			GROUP BY tenant_id
		) td ON td.tenant_id = tp.tenant_id
		WHERE `+strings.Join(filters, " AND ")+`
		ORDER BY tp.id ASC
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.TenantListItem, 0, 16)
	for rows.Next() {
		var item model.TenantListItem
		var brandConfigRaw string
		var institutionIDs string
		var moduleIDs string
		var moduleNames string
		var admins string
		var domains string
		var adminDomains string
		var institutionDomains string
		if err := rows.Scan(
			&item.TenantID,
			&item.TenantName,
			&item.TenantType,
			&item.Edition,
			&item.Status,
			&item.IsolationMode,
			&brandConfigRaw,
			&item.InstitutionCount,
			&institutionIDs,
			&item.MenuCount,
			&item.ModuleCount,
			&moduleIDs,
			&moduleNames,
			&admins,
			&domains,
			&adminDomains,
			&institutionDomains,
		); err != nil {
			return nil, err
		}
		brandSet := parseLoginBrandSet(brandConfigRaw)
		item.PlatformLoginBrand = selectLoginBrandConfig(brandSet, "platform-admin", item.TenantName)
		item.InstitutionLoginBrand = selectLoginBrandConfig(brandSet, "institution-admin", item.TenantName)
		item.LoginBrand = item.PlatformLoginBrand
		item.InstitutionIDs = splitCommaInt64List(institutionIDs)
		item.ModuleIDs = splitCommaInt64List(moduleIDs)
		item.ModuleNames = splitCommaList(moduleNames)
		item.AdminUsernames = splitCommaList(admins)
		item.Domains = splitCommaList(domains)
		item.AdminDomains = splitCommaList(adminDomains)
		item.InstitutionDomains = splitCommaList(institutionDomains)
		items = append(items, item)
	}
	return items, rows.Err()
}

func parseLoginBrandConfig(raw string) model.TenantLoginBrandConfig {
	var config model.TenantLoginBrandConfig
	if strings.TrimSpace(raw) == "" {
		return config
	}
	if err := json.Unmarshal([]byte(raw), &config); err != nil {
		return model.TenantLoginBrandConfig{}
	}
	return config
}

func parseLoginBrandSet(raw string) model.TenantLoginBrandSet {
	var brandSet model.TenantLoginBrandSet
	if strings.TrimSpace(raw) == "" {
		return brandSet
	}
	if err := json.Unmarshal([]byte(raw), &brandSet); err == nil && (hasLoginBrandConfig(brandSet.PlatformAdmin) || hasLoginBrandConfig(brandSet.InstitutionAdmin)) {
		return brandSet
	}
	var legacy model.TenantLoginBrandConfig
	if err := json.Unmarshal([]byte(raw), &legacy); err != nil {
		return model.TenantLoginBrandSet{}
	}
	return model.TenantLoginBrandSet{PlatformAdmin: legacy, InstitutionAdmin: legacy}
}

func hasLoginBrandConfig(config model.TenantLoginBrandConfig) bool {
	return strings.TrimSpace(config.Template) != "" || strings.TrimSpace(config.BrandName) != "" || strings.TrimSpace(config.LogoURL) != "" || strings.TrimSpace(config.LoginTitle) != "" || strings.TrimSpace(config.PrimaryColor) != ""
}

func selectLoginBrandConfig(brandSet model.TenantLoginBrandSet, entryType, tenantName string) model.TenantLoginBrandConfig {
	if entryType == "institution-admin" {
		return normalizeLoginBrandConfig(brandSet.InstitutionAdmin, tenantName, entryType)
	}
	return normalizeLoginBrandConfig(brandSet.PlatformAdmin, tenantName, "platform-admin")
}

func normalizeLoginBrandConfig(config model.TenantLoginBrandConfig, tenantName, entryType string) model.TenantLoginBrandConfig {
	brandName := firstNonEmpty(config.BrandName, tenantName, "总控平台")
	primaryColor := firstNonEmpty(config.PrimaryColor, "#1677ff")
	template := firstNonEmpty(config.Template, defaultLoginTemplate(entryType))
	return model.TenantLoginBrandConfig{
		Template:        template,
		BrandName:       brandName,
		LogoURL:         strings.TrimSpace(config.LogoURL),
		LoginTitle:      firstNonEmpty(config.LoginTitle, brandName+"管理后台"),
		LoginSubtitle:   firstNonEmpty(config.LoginSubtitle, "请输入账号密码登录"),
		BackgroundURL:   strings.TrimSpace(config.BackgroundURL),
		PrimaryColor:    primaryColor,
		Copyright:       strings.TrimSpace(config.Copyright),
		HeroBadge:       firstNonEmpty(config.HeroBadge, brandName),
		HeroTitle:       firstNonEmpty(config.HeroTitle, "欢迎进入"+brandName),
		HeroDescription: firstNonEmpty(config.HeroDescription, "独立租户后台，按客户域名、菜单权限和业务配置隔离运行。"),
	}
}

func defaultLoginTemplate(entryType string) string {
	if entryType == "institution-admin" {
		return "education-split"
	}
	return "business-split"
}

func splitCommaInt64List(raw string) []int64 {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	items := make([]int64, 0, len(parts))
	for _, part := range parts {
		value, err := strconv.ParseInt(strings.TrimSpace(part), 10, 64)
		if err == nil && value > 0 {
			items = append(items, value)
		}
	}
	return items
}

func splitCommaList(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			items = append(items, trimmed)
		}
	}
	return items
}

func (repo *Repository) GetTenantUserRole(ctx context.Context, tenantID string, userID int64) (string, error) {
	if strings.TrimSpace(tenantID) == "" || userID <= 0 {
		return "", nil
	}
	var role string
	err := repo.db.QueryRowContext(ctx, `
		SELECT IFNULL(user_role, '')
		FROM tenant_user
		WHERE tenant_id = ? AND user_id = ? AND del_flag = 0
		LIMIT 1
	`, strings.TrimSpace(tenantID), userID).Scan(&role)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return role, err
}

func (repo *Repository) GetTenantStorageConfig(ctx context.Context, tenantID, provider string) (tenantstorage.Config, error) {
	return tenantstorage.Get(ctx, repo.db, tenantID, provider)
}

func (repo *Repository) SaveTenantStorageConfig(ctx context.Context, input tenantstorage.Config) error {
	return tenantstorage.Save(ctx, repo.db, input)
}

func (repo *Repository) SaveTenant(ctx context.Context, input model.TenantMutation, operatorID *int64) error {
	tenantID := normalizeTenantID(input.TenantID)
	tenantName := strings.TrimSpace(input.TenantName)
	if tenantID == "" {
		return fmt.Errorf("tenantId is required")
	}
	if tenantName == "" {
		return fmt.Errorf("tenantName is required")
	}

	tenantType := firstNonEmpty(strings.TrimSpace(input.TenantType), "partner")
	edition := firstNonEmpty(strings.TrimSpace(input.Edition), "enterprise")
	status := firstNonEmpty(strings.TrimSpace(input.Status), "active")
	isolationMode := firstNonEmpty(strings.TrimSpace(input.IsolationMode), "shared_db")
	remark := strings.TrimSpace(input.Remark)
	platformLoginBrand := input.PlatformLoginBrand
	institutionLoginBrand := input.InstitutionLoginBrand
	if !hasLoginBrandConfig(platformLoginBrand) && hasLoginBrandConfig(input.LoginBrand) {
		platformLoginBrand = input.LoginBrand
	}
	if !hasLoginBrandConfig(institutionLoginBrand) && hasLoginBrandConfig(input.LoginBrand) {
		institutionLoginBrand = input.LoginBrand
	}
	brandSet := model.TenantLoginBrandSet{
		PlatformAdmin:    normalizeLoginBrandConfig(platformLoginBrand, tenantName, "platform-admin"),
		InstitutionAdmin: normalizeLoginBrandConfig(institutionLoginBrand, tenantName, "institution-admin"),
	}
	brandConfigBytes, err := json.Marshal(brandSet)
	if err != nil {
		return err
	}

	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO tenant_profile (tenant_id, tenant_name, tenant_type, parent_tenant_id, edition, status, isolation_mode, brand_config, remark, create_time, update_time, del_flag)
		VALUES (?, ?, ?, 'platform', ?, ?, ?, ?, ?, NOW(), NOW(), 0)
		ON DUPLICATE KEY UPDATE
		  tenant_name = VALUES(tenant_name),
		  tenant_type = VALUES(tenant_type),
		  edition = VALUES(edition),
		  status = VALUES(status),
		  isolation_mode = VALUES(isolation_mode),
		  brand_config = VALUES(brand_config),
		  remark = VALUES(remark),
		  update_time = NOW(),
		  del_flag = 0
	`, tenantID, tenantName, tenantType, edition, status, isolationMode, string(brandConfigBytes), remark); err != nil {
		return err
	}

	adminDomains := input.AdminDomains
	if adminDomains == nil {
		adminDomains = input.Domains
	}
	if err := repo.replaceTenantDomainsTx(ctx, tx, tenantID, "platform-admin", adminDomains); err != nil {
		return err
	}
	if err := repo.replaceTenantDomainsTx(ctx, tx, tenantID, "institution-admin", input.InstitutionDomains); err != nil {
		return err
	}
	if err := repo.replaceTenantInstitutionsTx(ctx, tx, tenantID, input.InstitutionIDs); err != nil {
		return err
	}
	menuIDs := input.MenuIDs
	if input.ModuleIDs != nil {
		if err := repo.replaceTenantModulesTx(ctx, tx, tenantID, input.ModuleIDs); err != nil {
			return err
		}
		menuIDs, err = repo.listModuleMenuIDsTx(ctx, tx, input.ModuleIDs)
		if err != nil {
			return err
		}
	}
	if err := repo.replaceTenantMenusTx(ctx, tx, tenantID, menuIDs); err != nil {
		return err
	}
	if strings.TrimSpace(input.AdminUsername) != "" {
		if err := repo.upsertTenantAdminTx(ctx, tx, tenantID, input, operatorID); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func normalizeTenantID(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, " ", "-")
	return value
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func (repo *Repository) replaceTenantDomainsTx(ctx context.Context, tx *sql.Tx, tenantID, entryType string, domains []string) error {
	entryType = firstNonEmpty(strings.TrimSpace(entryType), "platform-admin")
	if _, err := tx.ExecContext(ctx, `UPDATE tenant_domain SET del_flag = 1, update_time = NOW() WHERE tenant_id = ? AND entry_type = ? AND del_flag = 0`, tenantID, entryType); err != nil {
		return err
	}
	seen := map[string]struct{}{}
	for index, domain := range domains {
		domain = strings.ToLower(strings.TrimSpace(domain))
		if domain == "" {
			continue
		}
		if _, ok := seen[domain]; ok {
			continue
		}
		seen[domain] = struct{}{}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO tenant_domain (tenant_id, domain, entry_type, is_primary, create_time, update_time, del_flag)
			VALUES (?, ?, ?, ?, NOW(), NOW(), 0)
			ON DUPLICATE KEY UPDATE tenant_id = VALUES(tenant_id), entry_type = VALUES(entry_type), is_primary = VALUES(is_primary), update_time = NOW(), del_flag = 0
		`, tenantID, domain, entryType, index == 0); err != nil {
			return err
		}
	}
	return nil
}

func (repo *Repository) replaceTenantInstitutionsTx(ctx context.Context, tx *sql.Tx, tenantID string, institutionIDs []int64) error {
	if institutionIDs == nil {
		return nil
	}
	normalized := normalizeInstitutionIDs(institutionIDs)
	if err := repo.ensureInstitutionTenantAssignableTx(ctx, tx, tenantID, normalized); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE tenant_institution SET del_flag = 1, update_time = NOW() WHERE tenant_id = ? AND del_flag = 0`, tenantID); err != nil {
		return err
	}
	for _, institutionID := range normalized {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO tenant_institution (tenant_id, institution_id, create_time, update_time, del_flag)
			VALUES (?, ?, NOW(), NOW(), 0)
			ON DUPLICATE KEY UPDATE tenant_id = VALUES(tenant_id), update_time = NOW(), del_flag = 0
		`, tenantID, institutionID); err != nil {
			return err
		}
	}
	return nil
}

func (repo *Repository) ensureInstitutionTenantAssignableTx(ctx context.Context, tx *sql.Tx, tenantID string, institutionIDs []int64) error {
	if len(institutionIDs) == 0 {
		return nil
	}
	placeholders := make([]string, 0, len(institutionIDs))
	args := make([]any, 0, len(institutionIDs)+1)
	for _, institutionID := range institutionIDs {
		placeholders = append(placeholders, "?")
		args = append(args, institutionID)
	}
	args = append(args, strings.TrimSpace(tenantID))

	var organName string
	var tenantName string
	err := tx.QueryRowContext(ctx, `
		SELECT IFNULL(oi.organ_name, ''), IFNULL(tp.tenant_name, ti.tenant_id)
		FROM tenant_institution ti
		JOIN org_institution oi ON oi.id = ti.institution_id AND oi.del_flag = 0
		LEFT JOIN tenant_profile tp ON tp.tenant_id = ti.tenant_id AND tp.del_flag = 0
		WHERE ti.del_flag = 0
		  AND ti.institution_id IN (`+strings.Join(placeholders, ",")+`)
		  AND ti.tenant_id <> ?
		LIMIT 1
	`, args...).Scan(&organName, &tenantName)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return err
	}
	return fmt.Errorf("机构%s已归属%s，不能重复绑定", strings.TrimSpace(organName), strings.TrimSpace(tenantName))
}

func (repo *Repository) replaceTenantModulesTx(ctx context.Context, tx *sql.Tx, tenantID string, moduleIDs []int64) error {
	if _, err := tx.ExecContext(ctx, `UPDATE tenant_module SET del_flag = 1, update_time = NOW() WHERE tenant_id = ? AND del_flag = 0`, tenantID); err != nil {
		return err
	}
	seen := map[int64]struct{}{}
	for _, moduleID := range moduleIDs {
		if moduleID <= 0 {
			continue
		}
		if _, ok := seen[moduleID]; ok {
			continue
		}
		seen[moduleID] = struct{}{}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO tenant_module (tenant_id, module_id, create_time, update_time, del_flag)
			VALUES (?, ?, NOW(), NOW(), 0)
			ON DUPLICATE KEY UPDATE update_time = NOW(), del_flag = 0
		`, tenantID, moduleID); err != nil {
			return err
		}
	}
	return nil
}

func (repo *Repository) listModuleMenuIDsTx(ctx context.Context, tx *sql.Tx, moduleIDs []int64) ([]int64, error) {
	normalized := normalizeInt64IDs(moduleIDs)
	if len(normalized) == 0 {
		return []int64{}, nil
	}
	placeholders := make([]string, 0, len(normalized))
	args := make([]any, 0, len(normalized))
	for _, moduleID := range normalized {
		placeholders = append(placeholders, "?")
		args = append(args, moduleID)
	}
	rows, err := tx.QueryContext(ctx, `
		SELECT DISTINCT smm.menu_id
		FROM sys_module_menu smm
		JOIN sso_menu sm ON sm.id = smm.menu_id AND sm.del_flag = 0
		WHERE smm.del_flag = 0 AND smm.module_id IN (`+strings.Join(placeholders, ",")+`)
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	menuIDs := make([]int64, 0, 64)
	for rows.Next() {
		var menuID int64
		if err := rows.Scan(&menuID); err != nil {
			return nil, err
		}
		menuIDs = append(menuIDs, menuID)
	}
	return menuIDs, rows.Err()
}

func (repo *Repository) CountAllowedTenantMenus(ctx context.Context, tenantID string, menuIDs []int64) (int, error) {
	normalized := normalizeInt64IDs(menuIDs)
	if len(normalized) == 0 {
		return 0, nil
	}
	placeholders := make([]string, 0, len(normalized))
	args := make([]any, 0, len(normalized)+1)
	args = append(args, strings.TrimSpace(tenantID))
	for _, menuID := range normalized {
		placeholders = append(placeholders, "?")
		args = append(args, menuID)
	}
	var total int
	err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT smm.menu_id)
		FROM tenant_module tm
		JOIN sys_module m ON m.id = tm.module_id AND m.del_flag = 0
		JOIN sys_module_menu smm ON smm.module_id = tm.module_id AND smm.del_flag = 0
		WHERE tm.tenant_id = ? AND tm.del_flag = 0 AND smm.menu_id IN (`+strings.Join(placeholders, ",")+`)
	`, args...).Scan(&total)
	return total, err
}

func normalizeInt64IDs(values []int64) []int64 {
	seen := map[int64]struct{}{}
	result := make([]int64, 0, len(values))
	for _, value := range values {
		if value <= 0 {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func (repo *Repository) replaceTenantMenusTx(ctx context.Context, tx *sql.Tx, tenantID string, menuIDs []int64) error {
	if _, err := tx.ExecContext(ctx, `UPDATE tenant_menu SET del_flag = 1, update_time = NOW() WHERE tenant_id = ? AND del_flag = 0`, tenantID); err != nil {
		return err
	}
	if len(menuIDs) == 0 {
		return nil
	}
	seen := map[int64]struct{}{}
	for _, menuID := range menuIDs {
		if menuID <= 0 {
			continue
		}
		if _, ok := seen[menuID]; ok {
			continue
		}
		seen[menuID] = struct{}{}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO tenant_menu (tenant_id, menu_id, create_time, update_time, del_flag)
			VALUES (?, ?, NOW(), NOW(), 0)
			ON DUPLICATE KEY UPDATE update_time = NOW(), del_flag = 0
		`, tenantID, menuID); err != nil {
			return err
		}
	}
	return nil
}

func (repo *Repository) upsertTenantAdminTx(ctx context.Context, tx *sql.Tx, tenantID string, input model.TenantMutation, operatorID *int64) error {
	username := strings.TrimSpace(input.AdminUsername)
	password := strings.TrimSpace(input.AdminPassword)
	nickName := firstNonEmpty(strings.TrimSpace(input.AdminNickName), strings.TrimSpace(input.TenantName)+"管理员")
	mobile := strings.TrimSpace(input.AdminMobile)

	var userID int64
	err := tx.QueryRowContext(ctx, `SELECT id FROM sso_user WHERE del_flag = 0 AND username = ? LIMIT 1`, username).Scan(&userID)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if err == sql.ErrNoRows {
		if password == "" {
			password = "123456"
		}
		hashBytes, hashErr := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if hashErr != nil {
			return hashErr
		}
		result, insErr := tx.ExecContext(ctx, `
			INSERT INTO sso_user (uuid, version, username, password, mobile, avatar, nick_name, user_type, is_admin, del_flag, create_time, update_time)
			VALUES (?, 0, ?, ?, ?, '', ?, 0, 0, 0, NOW(), NOW())
		`, uuid.NewString(), username, string(hashBytes), mobile, nickName)
		if insErr != nil {
			return insErr
		}
		userID, insErr = result.LastInsertId()
		if insErr != nil {
			return insErr
		}
	} else {
		if password != "" {
			hashBytes, hashErr := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			if hashErr != nil {
				return hashErr
			}
			if _, err := tx.ExecContext(ctx, `
				UPDATE sso_user
				SET password = ?, mobile = CASE WHEN ? = '' THEN mobile ELSE ? END, nick_name = ?, user_type = 0, is_admin = 0, update_time = NOW()
				WHERE id = ? AND del_flag = 0
			`, string(hashBytes), mobile, mobile, nickName, userID); err != nil {
				return err
			}
		} else if _, err := tx.ExecContext(ctx, `
			UPDATE sso_user
			SET mobile = CASE WHEN ? = '' THEN mobile ELSE ? END, nick_name = ?, user_type = 0, is_admin = 0, update_time = NOW()
			WHERE id = ? AND del_flag = 0
		`, mobile, mobile, nickName, userID); err != nil {
			return err
		}
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO tenant_user (tenant_id, user_id, user_role, create_time, update_time, del_flag)
		VALUES (?, ?, 'tenant_admin', NOW(), NOW(), 0)
		ON DUPLICATE KEY UPDATE user_role = VALUES(user_role), update_time = NOW(), del_flag = 0
	`, tenantID, userID)
	return err
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

func (repo *Repository) PageModules(ctx context.Context, current, size int, name string, moduleType int, tenantID string) (model.PageResult[model.Module], error) {
	if current <= 0 {
		current = 1
	}
	if size <= 0 {
		size = 10
	}
	offset := (current - 1) * size

	filters := []string{"m.del_flag = 0"}
	args := make([]any, 0, 4)
	if strings.TrimSpace(name) != "" {
		filters = append(filters, "m.name LIKE ?")
		args = append(args, "%"+strings.TrimSpace(name)+"%")
	}
	if moduleType > 0 {
		filters = append(filters, "m.type = ?")
		args = append(args, moduleType)
	}
	if tenantID = strings.TrimSpace(tenantID); tenantID != "" {
		filters = append(filters, "m.tenant_id = ?")
		args = append(args, tenantID)
	} else {
		filters = append(filters, "m.tenant_id = 'platform'")
	}
	whereClause := strings.Join(filters, " AND ")

	var total int
	if err := repo.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM sys_module m WHERE "+whereClause, args...).Scan(&total); err != nil {
		return model.PageResult[model.Module]{}, err
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT m.id,
		       IFNULL(m.tenant_id, ''),
		       IFNULL(m.owner_type, ''),
		       IFNULL(m.source_module_id, 0),
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
		WHERE `+whereClause+`
		GROUP BY m.id, m.tenant_id, m.owner_type, m.source_module_id, m.name, m.type, m.price, m.remark, m.create_time, m.update_time
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
			&item.TenantID,
			&item.OwnerType,
			&item.SourceModuleID,
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
	if err := rows.Err(); err != nil {
		return model.PageResult[model.Module]{}, err
	}

	if len(items) > 0 {
		counts, err := repo.countVisibleModuleLeavesByID(ctx, tenantID, items)
		if err != nil {
			return model.PageResult[model.Module]{}, err
		}
		for index := range items {
			if count, ok := counts[items[index].ID]; ok {
				items[index].MenuCount = count
			}
		}
	}

	return model.PageResult[model.Module]{
		Items:   items,
		Total:   total,
		Current: current,
		Size:    size,
	}, nil
}

func (repo *Repository) countVisibleModuleLeavesByID(ctx context.Context, tenantID string, modules []model.Module) (map[int64]int, error) {
	result := make(map[int64]int, len(modules))
	if len(modules) == 0 {
		return result, nil
	}

	rawMenus, err := repo.listInstitutionMenus(ctx, strings.TrimSpace(tenantID))
	if err != nil {
		return nil, err
	}
	if len(rawMenus) == 0 {
		return result, nil
	}

	moduleIDs := make([]int64, 0, len(modules))
	for _, item := range modules {
		if item.ID > 0 {
			moduleIDs = append(moduleIDs, item.ID)
		}
	}
	if len(moduleIDs) == 0 {
		return result, nil
	}

	placeholders := make([]string, 0, len(moduleIDs))
	args := make([]any, 0, len(moduleIDs))
	for _, id := range moduleIDs {
		placeholders = append(placeholders, "?")
		args = append(args, id)
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT module_id, menu_id
		FROM sys_module_menu
		WHERE del_flag = 0 AND module_id IN (`+strings.Join(placeholders, ",")+`)
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	selectedByModule := make(map[int64]map[int64]struct{}, len(moduleIDs))
	for rows.Next() {
		var moduleID int64
		var menuID int64
		if err := rows.Scan(&moduleID, &menuID); err != nil {
			return nil, err
		}
		if _, ok := selectedByModule[moduleID]; !ok {
			selectedByModule[moduleID] = make(map[int64]struct{})
		}
		selectedByModule[moduleID][menuID] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for _, moduleID := range moduleIDs {
		result[moduleID] = countVisibleInstitutionModuleLeaves(rawMenus, selectedByModule[moduleID])
	}
	return result, nil
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

func (repo *Repository) PageInstitutions(ctx context.Context, current, size int, keyword, mobile, registerTimeBegin, registerTimeEnd string, enabled *bool, status, openType, provinceCode, cityCode, regionCode *int, tenantID string) (model.InstitutionPage, error) {
	if current <= 0 {
		current = 1
	}
	if size <= 0 {
		size = 10
	}
	offset := (current - 1) * size

	whereClause, args := buildInstitutionWhereClause(keyword, mobile, registerTimeBegin, registerTimeEnd, enabled, status, openType, provinceCode, cityCode, regionCode)
	statusExpr := institutionStatusExpr("oi")
	tenantJoinClause := ""
	queryArgs := append([]any{}, args...)
	if tenantID = strings.TrimSpace(tenantID); tenantID != "" {
		tenantJoinClause = "JOIN tenant_institution ti ON ti.institution_id = oi.id AND ti.del_flag = 0 AND ti.tenant_id = ?"
		queryArgs = append([]any{tenantID}, queryArgs...)
	}

	var total int
	if err := repo.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM org_institution oi "+tenantJoinClause+" WHERE "+whereClause, queryArgs...).Scan(&total); err != nil {
		return model.InstitutionPage{}, err
	}

	var summary model.InstitutionSummary
	if err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(*),
		       COALESCE(SUM(CASE WHEN `+statusExpr+` IN (1, 3) THEN 1 ELSE 0 END), 0),
		       COALESCE(SUM(CASE WHEN `+statusExpr+` IN (2, 4) THEN 1 ELSE 0 END), 0)
		FROM org_institution oi
		`+tenantJoinClause+`
		WHERE `+whereClause, queryArgs...).Scan(&summary.TotalCount, &summary.EnabledCount, &summary.DisabledCount); err != nil {
		return model.InstitutionPage{}, err
	}

	listArgs := append(append([]any{}, queryArgs...), size, offset)
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
		       IFNULL(tenant_scope.tenant_id, ''),
		       IFNULL(tenant_scope.tenant_name, ''),
		       IFNULL(oi.logo, ''),
		       IFNULL(oi.enabled, 0),
		       `+statusExpr+`,
		       IFNULL(oi.open_type, 2),
		       IFNULL(oi.open_duration, ''),
		       IFNULL(om_current.module_id, 0),
		       IFNULL(sm_current.name, ''),
		       IFNULL(DATE_FORMAT(oi.create_time, '%Y-%m-%d %H:%i:%s'), ''),
		       IFNULL(DATE_FORMAT(oi.expire_end_time, '%Y-%m-%d %H:%i:%s'), ''),
		       COUNT(DISTINCT CASE WHEN iu.del_flag = 0 THEN iu.id END) AS staff_count,
		       COUNT(DISTINCT CASE WHEN iu.del_flag = 0 AND IFNULL(iu.disabled, 0) = 0 THEN iu.id END) AS active_staff_count,
		       COUNT(DISTINCT CASE WHEN iu.del_flag = 0 AND IFNULL(iu.is_admin, 0) = 1 THEN iu.id END) AS admin_count
		FROM org_institution oi
		`+tenantJoinClause+`
		LEFT JOIN org_module om_current ON om_current.id = (
			SELECT om2.id FROM org_module om2 WHERE om2.org_id = oi.id AND om2.del_flag = 0 ORDER BY om2.id DESC LIMIT 1
		)
		LEFT JOIN sys_module sm_current ON sm_current.id = om_current.module_id AND sm_current.del_flag = 0
		LEFT JOIN (
			SELECT ti_latest.institution_id,
			       ti_latest.tenant_id,
			       IFNULL(tp.tenant_name, '') AS tenant_name
			FROM tenant_institution ti_latest
			JOIN (
				SELECT institution_id, MAX(id) AS id
				FROM tenant_institution
				WHERE del_flag = 0
				GROUP BY institution_id
			) latest_tenant ON latest_tenant.id = ti_latest.id
			LEFT JOIN tenant_profile tp ON tp.tenant_id = ti_latest.tenant_id AND tp.del_flag = 0
			WHERE ti_latest.del_flag = 0
		) tenant_scope ON tenant_scope.institution_id = oi.id
		LEFT JOIN inst_user iu ON iu.inst_id = oi.id
		WHERE `+whereClause+`
		GROUP BY oi.id, oi.organ_name, oi.organ_code, oi.login_name, oi.mobile, oi.principal, oi.province, oi.city, oi.region, oi.address, tenant_scope.tenant_id, tenant_scope.tenant_name, oi.logo, oi.enabled, oi.status, oi.open_type, oi.open_duration, om_current.module_id, sm_current.name, oi.create_time, oi.expire_end_time
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
			&item.TenantID,
			&item.TenantName,
			&item.Logo,
			&item.Enabled,
			&item.Status,
			&item.OpenType,
			&item.OpenDuration,
			&item.CurrentModuleID,
			&item.CurrentModuleName,
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

func (repo *Repository) GetGovernmentOverview(ctx context.Context, userID int64) (model.GovernmentOverview, error) {
	contextInfo, err := repo.getGovernmentOverviewContext(ctx, userID)
	if err != nil {
		return model.GovernmentOverview{}, err
	}
	if contextInfo.Disabled {
		return model.GovernmentOverview{}, errors.New("政府账号已停用")
	}
	if normalizeGovernmentLevel(contextInfo.Level) == "" {
		return model.GovernmentOverview{}, errors.New("政府账号未配置监管层级")
	}

	scopeText, scopeCodeText, scopeCount := summarizeGovernmentScopes(contextInfo.Level, contextInfo.Scopes)
	whereClause, args := buildGovernmentInstitutionWhereClause("oi", contextInfo.Level, contextInfo.Scopes)
	groupCodeExpr, groupNameExpr, groupLevelLabel, hasChildren := governmentOverviewGroupMeta(contextInfo.Level)

	rows, err := repo.db.QueryContext(ctx, `
		SELECT grouped.region_code,
		       grouped.region_name,
		       COUNT(*) AS institution_count,
		       COALESCE(SUM(grouped.reading_student_count), 0) AS reading_student_count,
		       COALESCE(SUM(grouped.intent_student_count), 0) AS intent_student_count,
		       COALESCE(SUM(grouped.order_count), 0) AS order_count
		FROM (
			SELECT `+groupCodeExpr+` AS region_code,
			       `+groupNameExpr+` AS region_name,
			       COALESCE(student_stats.reading_student_count, 0) AS reading_student_count,
			       COALESCE(student_stats.intent_student_count, 0) AS intent_student_count,
			       COALESCE(order_stats.order_count, 0) AS order_count
			FROM org_institution oi
			LEFT JOIN (
				SELECT inst_id,
				       SUM(CASE WHEN student_status = 1 THEN 1 ELSE 0 END) AS reading_student_count,
				       SUM(CASE WHEN student_status = 0 THEN 1 ELSE 0 END) AS intent_student_count
				FROM inst_student
				WHERE del_flag = 0
				GROUP BY inst_id
			) student_stats ON student_stats.inst_id = oi.id
			LEFT JOIN (
				SELECT inst_id,
				       COUNT(*) AS order_count
				FROM sale_order
				WHERE del_flag = 0
				GROUP BY inst_id
			) order_stats ON order_stats.inst_id = oi.id
			WHERE `+whereClause+`
		) grouped
		GROUP BY grouped.region_code, grouped.region_name
		ORDER BY institution_count DESC, region_code ASC
	`, args...)
	if err != nil {
		return model.GovernmentOverview{}, err
	}
	defer rows.Close()

	summary := make([]model.GovernmentOverviewEntry, 0, 16)
	overview := model.GovernmentOverview{
		Level:           normalizeGovernmentLevel(contextInfo.Level),
		LevelLabel:      governmentLevelLabel(contextInfo.Level),
		ScopeText:       scopeText,
		ScopeCodeText:   scopeCodeText,
		ScopeCount:      scopeCount,
		RegionalSummary: summary,
	}

	for rows.Next() {
		var item model.GovernmentOverviewEntry
		if err := rows.Scan(
			&item.RegionCode,
			&item.RegionName,
			&item.InstitutionCount,
			&item.ReadingStudentCount,
			&item.IntentStudentCount,
			&item.OrderCount,
		); err != nil {
			return model.GovernmentOverview{}, err
		}
		item.LevelLabel = groupLevelLabel
		summary = append(summary, item)
		overview.InstitutionCount += item.InstitutionCount
		overview.ReadingStudentCount += item.ReadingStudentCount
		overview.OrderCount += item.OrderCount
	}
	if err := rows.Err(); err != nil {
		return model.GovernmentOverview{}, err
	}

	overview.RegionalSummary = summary
	if hasChildren {
		overview.SubordinateRegionCount = len(summary)
	}
	return overview, nil
}

func (repo *Repository) PageGovernmentInstitutions(ctx context.Context, userID int64, current, size int, keyword string, status, openType *int) (model.GovernmentInstitutionPage, error) {
	if current <= 0 {
		current = 1
	}
	if size <= 0 {
		size = 10
	}
	offset := (current - 1) * size

	contextInfo, err := repo.getGovernmentOverviewContext(ctx, userID)
	if err != nil {
		return model.GovernmentInstitutionPage{}, err
	}
	if contextInfo.Disabled {
		return model.GovernmentInstitutionPage{}, errors.New("政府账号已停用")
	}
	if normalizeGovernmentLevel(contextInfo.Level) == "" {
		return model.GovernmentInstitutionPage{}, errors.New("政府账号未配置监管层级")
	}

	scopeText, scopeCodeText, scopeCount := summarizeGovernmentScopes(contextInfo.Level, contextInfo.Scopes)
	scopeWhereClause, scopeArgs := buildGovernmentInstitutionWhereClause("oi", contextInfo.Level, contextInfo.Scopes)
	whereClause, args := buildGovernmentInstitutionListWhereClause(scopeWhereClause, scopeArgs, keyword, status, openType)
	statusExpr := institutionStatusExpr("oi")

	var summary model.GovernmentInstitutionSummary
	if err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(*),
		       COALESCE(SUM(CASE WHEN summary_source.status = 1 THEN 1 ELSE 0 END), 0),
		       COALESCE(SUM(CASE WHEN summary_source.status = 3 THEN 1 ELSE 0 END), 0),
		       COALESCE(SUM(CASE WHEN summary_source.status = 2 THEN 1 ELSE 0 END), 0),
		       COALESCE(SUM(CASE WHEN summary_source.status = 4 THEN 1 ELSE 0 END), 0),
		       COALESCE(SUM(summary_source.reading_student_count), 0),
		       COALESCE(SUM(summary_source.intent_student_count), 0),
		       COALESCE(SUM(summary_source.order_count), 0)
		FROM (
			SELECT `+statusExpr+` AS status,
			       COALESCE(student_stats.reading_student_count, 0) AS reading_student_count,
			       COALESCE(student_stats.intent_student_count, 0) AS intent_student_count,
			       COALESCE(order_stats.order_count, 0) AS order_count
			FROM org_institution oi
			LEFT JOIN (
				SELECT inst_id,
				       SUM(CASE WHEN student_status = 1 THEN 1 ELSE 0 END) AS reading_student_count,
				       SUM(CASE WHEN student_status = 0 THEN 1 ELSE 0 END) AS intent_student_count
				FROM inst_student
				WHERE del_flag = 0
				GROUP BY inst_id
			) student_stats ON student_stats.inst_id = oi.id
			LEFT JOIN (
				SELECT inst_id,
				       COUNT(*) AS order_count
				FROM sale_order
				WHERE del_flag = 0
				GROUP BY inst_id
			) order_stats ON order_stats.inst_id = oi.id
			WHERE `+whereClause+`
		) summary_source
	`, args...).Scan(
		&summary.TotalCount,
		&summary.EnabledCount,
		&summary.WarningCount,
		&summary.DisabledCount,
		&summary.ExpiredCount,
		&summary.ReadingStudentCount,
		&summary.IntentStudentCount,
		&summary.OrderCount,
	); err != nil {
		return model.GovernmentInstitutionPage{}, err
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
		       IFNULL(oi.enabled, 0),
		       `+statusExpr+`,
		       IFNULL(oi.open_type, 2),
		       IFNULL(oi.open_duration, ''),
		       IFNULL(om_current.module_id, 0),
		       IFNULL(sm_current.name, ''),
		       IFNULL(DATE_FORMAT(oi.create_time, '%Y-%m-%d %H:%i:%s'), ''),
		       IFNULL(DATE_FORMAT(oi.expire_end_time, '%Y-%m-%d %H:%i:%s'), ''),
		       COALESCE(staff_stats.staff_count, 0),
		       COALESCE(staff_stats.active_staff_count, 0),
		       COALESCE(staff_stats.admin_count, 0),
		       COALESCE(student_stats.reading_student_count, 0),
		       COALESCE(student_stats.intent_student_count, 0),
		       COALESCE(order_stats.order_count, 0)
		FROM org_institution oi
		LEFT JOIN (
			SELECT inst_id,
			       COUNT(*) AS staff_count,
			       SUM(CASE WHEN IFNULL(disabled, 0) = 0 THEN 1 ELSE 0 END) AS active_staff_count,
			       SUM(CASE WHEN IFNULL(is_admin, 0) = 1 THEN 1 ELSE 0 END) AS admin_count
			FROM inst_user
			WHERE del_flag = 0
			GROUP BY inst_id
		) staff_stats ON staff_stats.inst_id = oi.id
		LEFT JOIN (
			SELECT inst_id,
			       SUM(CASE WHEN student_status = 1 THEN 1 ELSE 0 END) AS reading_student_count,
			       SUM(CASE WHEN student_status = 0 THEN 1 ELSE 0 END) AS intent_student_count
			FROM inst_student
			WHERE del_flag = 0
			GROUP BY inst_id
		) student_stats ON student_stats.inst_id = oi.id
		LEFT JOIN (
			SELECT inst_id,
			       COUNT(*) AS order_count
			FROM sale_order
			WHERE del_flag = 0
			GROUP BY inst_id
		) order_stats ON order_stats.inst_id = oi.id
		WHERE `+whereClause+`
		ORDER BY oi.id DESC
		LIMIT ? OFFSET ?
	`, listArgs...)
	if err != nil {
		return model.GovernmentInstitutionPage{}, err
	}
	defer rows.Close()

	items := make([]model.GovernmentInstitution, 0, size)
	for rows.Next() {
		var item model.GovernmentInstitution
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
			&item.Enabled,
			&item.Status,
			&item.OpenType,
			&item.OpenDuration,
			&item.RegisterTime,
			&item.ExpireEndTime,
			&item.StaffCount,
			&item.ActiveStaffCount,
			&item.AdminCount,
			&item.ReadingStudentCount,
			&item.IntentStudentCount,
			&item.OrderCount,
		); err != nil {
			return model.GovernmentInstitutionPage{}, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return model.GovernmentInstitutionPage{}, err
	}

	return model.GovernmentInstitutionPage{
		Items:         items,
		Total:         summary.TotalCount,
		Current:       current,
		Size:          size,
		Level:         normalizeGovernmentLevel(contextInfo.Level),
		LevelLabel:    governmentLevelLabel(contextInfo.Level),
		ScopeText:     scopeText,
		ScopeCodeText: scopeCodeText,
		ScopeCount:    scopeCount,
		Summary:       &summary,
	}, nil
}

func (repo *Repository) getGovernmentOverviewContext(ctx context.Context, userID int64) (governmentOverviewContext, error) {
	row := repo.db.QueryRowContext(ctx, `
		SELECT IFNULL(gp.level, ''),
		       IFNULL(gp.disabled, 0),
		       IFNULL((
		           SELECT MAX(IFNULL(r.is_admin, 0))
		           FROM sso_user_role ur
		           INNER JOIN sso_role r ON r.id = ur.role_id
		           WHERE ur.user_id = u.id
		             AND r.del_flag = 0
		             AND r.role_type = 3
		             AND r.org_id = 1
		       ), 0)
		FROM sso_user u
		LEFT JOIN government_user_profile gp ON gp.user_id = u.id AND gp.del_flag = 0
		WHERE u.id = ? AND u.del_flag = 0
		LIMIT 1
	`, userID)

	var (
		result  governmentOverviewContext
		level   string
		isAdmin int
	)
	if err := row.Scan(&level, &result.Disabled, &isAdmin); err != nil {
		return governmentOverviewContext{}, err
	}
	result.IsAdmin = isAdmin > 0

	rows, err := repo.db.QueryContext(ctx, `
		SELECT IFNULL(scope_level, ''),
		       IFNULL(province_code, ''),
		       IFNULL(province_name, ''),
		       IFNULL(city_code, ''),
		       IFNULL(city_name, ''),
		       IFNULL(district_code, ''),
		       IFNULL(district_name, '')
		FROM government_user_scope
		WHERE user_id = ? AND del_flag = 0
		ORDER BY province_code ASC, city_code ASC, district_code ASC, id ASC
	`, userID)
	if err != nil {
		return governmentOverviewContext{}, err
	}
	defer rows.Close()

	scopes := make([]governmentOverviewScope, 0, 8)
	for rows.Next() {
		var item governmentOverviewScope
		if err := rows.Scan(
			&item.ScopeLevel,
			&item.ProvinceCode,
			&item.ProvinceName,
			&item.CityCode,
			&item.CityName,
			&item.DistrictCode,
			&item.DistrictName,
		); err != nil {
			return governmentOverviewContext{}, err
		}
		scopes = append(scopes, item)
	}
	if err := rows.Err(); err != nil {
		return governmentOverviewContext{}, err
	}

	result.Level = normalizeGovernmentLevel(level)
	if result.Level == "" {
		switch {
		case result.IsAdmin:
			result.Level = "super"
		case len(scopes) > 0:
			result.Level = normalizeGovernmentLevel(scopes[0].ScopeLevel)
		}
	}
	result.Scopes = scopes
	return result, nil
}

func (repo *Repository) ResolveTenantIDByInstitution(ctx context.Context, institutionID int64) (string, error) {
	if institutionID <= 0 {
		return "", nil
	}
	var tenantID string
	err := repo.db.QueryRowContext(ctx, `
		SELECT ti.tenant_id
		FROM tenant_institution ti
		JOIN tenant_profile tp ON tp.tenant_id = ti.tenant_id AND tp.del_flag = 0 AND tp.status = 'active'
		WHERE ti.institution_id = ? AND ti.del_flag = 0
		ORDER BY ti.id DESC
		LIMIT 1
	`, institutionID).Scan(&tenantID)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return strings.TrimSpace(tenantID), err
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
		       IFNULL(om_current.module_id, 0),
		       IFNULL(sm_current.name, ''),
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
		       END,
		       IFNULL(oip.login_slug, ''),
		       IFNULL(CAST(oip.login_brand_config AS CHAR), '')
		FROM org_institution oi
		LEFT JOIN org_institution_profile oip ON oip.institution_id = oi.id AND oip.del_flag = 0
		LEFT JOIN org_module om_current ON om_current.id = (
			SELECT om2.id FROM org_module om2 WHERE om2.org_id = oi.id AND om2.del_flag = 0 ORDER BY om2.id DESC LIMIT 1
		)
		LEFT JOIN sys_module sm_current ON sm_current.id = om_current.module_id AND sm_current.del_flag = 0
		WHERE oi.id = ? AND oi.del_flag = 0
		LIMIT 1
	`, id)

	var detail model.InstitutionDetail
	var galleryImagesRaw string
	var loginBrandRaw string
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
		&detail.CurrentModuleID,
		&detail.CurrentModuleName,
		&detail.ExpireStartTime,
		&detail.ExpireEndTime,
		&detail.Lng,
		&detail.Lat,
		&detail.Profile.Description,
		&detail.Profile.BusinessTime,
		&detail.Profile.Video,
		&galleryImagesRaw,
		&detail.Profile.LoginSlug,
		&loginBrandRaw,
	); err != nil {
		return model.InstitutionDetail{}, err
	}
	detail.Profile.GalleryImages = unmarshalStringSlice(galleryImagesRaw)
	detail.Profile.LoginBrand = parseLoginBrandConfig(loginBrandRaw)

	return detail, nil
}

func (repo *Repository) CheckInstitutionLoginNameAvailable(ctx context.Context, loginName string, excludeInstitutionID *int64) (model.InstitutionLoginNameAvailability, error) {
	available, message, err := repo.checkInstitutionLoginNameAvailable(ctx, repo.db, loginName, excludeInstitutionID)
	if err != nil {
		return model.InstitutionLoginNameAvailability{}, err
	}
	return model.InstitutionLoginNameAvailability{
		LoginName: strings.TrimSpace(loginName),
		Available: available,
		Message:   message,
	}, nil
}

func (repo *Repository) ensureInstitutionLoginNameAvailableTx(ctx context.Context, tx *sql.Tx, loginName string, excludeInstitutionID *int64) error {
	available, message, err := repo.checkInstitutionLoginNameAvailable(ctx, tx, loginName, excludeInstitutionID)
	if err != nil {
		return err
	}
	if available {
		return nil
	}
	if strings.TrimSpace(message) == "" {
		message = "登录账号已存在，请更换"
	}
	return fmt.Errorf("%s", message)
}

func (repo *Repository) checkInstitutionLoginNameAvailable(ctx context.Context, queryer loginNameQueryer, loginName string, excludeInstitutionID *int64) (bool, string, error) {
	loginName = strings.TrimSpace(loginName)
	if loginName == "" {
		return false, "登录账号不能为空", nil
	}

	findInstitutionSQL := `
		SELECT id
		FROM org_institution
		WHERE del_flag = 0 AND login_name = ?
	`
	findInstitutionArgs := []any{loginName}
	if excludeInstitutionID != nil && *excludeInstitutionID > 0 {
		findInstitutionSQL += " AND id <> ?"
		findInstitutionArgs = append(findInstitutionArgs, *excludeInstitutionID)
	}
	findInstitutionSQL += " ORDER BY id LIMIT 1"

	var occupiedInstitutionID int64
	err := queryer.QueryRowContext(ctx, findInstitutionSQL, findInstitutionArgs...).Scan(&occupiedInstitutionID)
	if err != nil && err != sql.ErrNoRows {
		return false, "", err
	}
	if err == nil {
		return false, "登录账号已存在，请更换", nil
	}

	var userID int64
	err = queryer.QueryRowContext(ctx, `
		SELECT id
		FROM sso_user
		WHERE del_flag = 0 AND username = ?
		ORDER BY id
		LIMIT 1
	`, loginName).Scan(&userID)
	if err != nil && err != sql.ErrNoRows {
		return false, "", err
	}
	if err == sql.ErrNoRows {
		return true, "", nil
	}

	if excludeInstitutionID != nil && *excludeInstitutionID > 0 {
		currentInstitutionID := *excludeInstitutionID

		var sameInstitutionCount int
		if err := queryer.QueryRowContext(ctx, `
			SELECT COUNT(1)
			FROM inst_user
			WHERE del_flag = 0 AND user_id = ? AND inst_id = ?
		`, userID, currentInstitutionID).Scan(&sameInstitutionCount); err != nil {
			return false, "", err
		}

		var otherInstitutionCount int
		if err := queryer.QueryRowContext(ctx, `
			SELECT COUNT(1)
			FROM inst_user
			WHERE del_flag = 0 AND user_id = ? AND inst_id <> ?
		`, userID, currentInstitutionID).Scan(&otherInstitutionCount); err != nil {
			return false, "", err
		}

		if sameInstitutionCount > 0 && otherInstitutionCount == 0 {
			return true, "", nil
		}
	}

	return false, "登录账号已存在，请更换", nil
}

// ensureInstitutionBootstrapAdminTx creates the org-scoped admin role, SSO user (login_name), inst_user link,
// and assigns the admin role. The SSO password is always set to defaultInstitutionBootstrapPassword ("123456").
func (repo *Repository) ensureInstitutionBootstrapAdminTx(
	ctx context.Context,
	tx *sql.Tx,
	institutionID int64,
	loginName, mobile, organName, principal string,
) error {
	loginName = strings.TrimSpace(loginName)
	if loginName == "" {
		return nil
	}
	organName = strings.TrimSpace(organName)
	hash, err := bcrypt.GenerateFromPassword([]byte(defaultInstitutionBootstrapPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	hashStr := string(hash)
	mobile = strings.TrimSpace(mobile)
	nick := strings.TrimSpace(principal)
	if nick == "" {
		nick = organName
	}
	if nick == "" {
		nick = loginName
	}

	var roleID int64
	var roleName string
	var roleDescription string
	err = tx.QueryRowContext(ctx, `
		SELECT id, IFNULL(role_name, ''), IFNULL(description, '')
		FROM sso_role
		WHERE org_id = ? AND role_type = 2 AND is_admin = 1 AND del_flag = 0
		LIMIT 1
	`, institutionID).Scan(&roleID, &roleName, &roleDescription)
	if err == sql.ErrNoRows {
		res, insErr := tx.ExecContext(ctx, `
			INSERT INTO sso_role (uuid, version, role_name, description, org_id, role_type, is_admin, is_default, del_flag, create_time, update_time)
			VALUES (?, 0, ?, ?, ?, 2, 1, 0, 0, NOW(), NOW())
		`, uuid.NewString(), defaultInstitutionAdminRoleName, defaultInstitutionAdminRoleDescription, institutionID)
		if insErr != nil {
			return insErr
		}
		newID, insErr := res.LastInsertId()
		if insErr != nil {
			return insErr
		}
		roleID = newID
		roleName = defaultInstitutionAdminRoleName
		roleDescription = defaultInstitutionAdminRoleDescription
	} else if err != nil {
		return err
	}
	normalizedRoleName := strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(roleName, "\r", ""), "\n", ""))
	normalizedRoleDescription := strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(roleDescription, "\r", ""), "\n", ""))
	if normalizedRoleName != defaultInstitutionAdminRoleName || normalizedRoleDescription != defaultInstitutionAdminRoleDescription {
		if _, err := tx.ExecContext(ctx, `
			UPDATE sso_role
			SET role_name = ?, description = ?, update_time = NOW()
			WHERE id = ? AND del_flag = 0
		`, defaultInstitutionAdminRoleName, defaultInstitutionAdminRoleDescription, roleID); err != nil {
			return err
		}
	}

	var userID int64
	err = tx.QueryRowContext(ctx, `
		SELECT id FROM sso_user WHERE del_flag = 0 AND username = ? LIMIT 1
	`, loginName).Scan(&userID)
	if err == sql.ErrNoRows {
		userType := 1
		res, insErr := tx.ExecContext(ctx, `
			INSERT INTO sso_user (uuid, version, username, password, mobile, avatar, nick_name, user_type, is_admin, del_flag, create_time)
			VALUES (?, 0, ?, ?, ?, '', ?, ?, 1, 0, NOW())
		`, uuid.NewString(), loginName, hashStr, mobile, nick, userType)
		if insErr != nil {
			return insErr
		}
		newUID, insErr := res.LastInsertId()
		if insErr != nil {
			return insErr
		}
		userID = newUID
	} else if err != nil {
		return err
	} else {
		var otherCount int
		if qErr := tx.QueryRowContext(ctx, `
			SELECT COUNT(1) FROM inst_user WHERE del_flag = 0 AND user_id = ? AND inst_id <> ?
		`, userID, institutionID).Scan(&otherCount); qErr != nil {
			return qErr
		}
		if otherCount > 0 {
			return fmt.Errorf("登录名 %s 已被其他机构占用", loginName)
		}
		if _, err := tx.ExecContext(ctx, `
			UPDATE sso_user
			SET password = ?,
			    mobile = CASE WHEN NULLIF(?, '') IS NULL THEN mobile ELSE ? END,
			    nick_name = CASE WHEN NULLIF(?, '') IS NULL THEN nick_name ELSE ? END
			WHERE id = ? AND del_flag = 0
		`, hashStr, mobile, mobile, nick, nick, userID); err != nil {
			return err
		}
	}

	var instUserID int64
	err = tx.QueryRowContext(ctx, `
		SELECT id
		FROM inst_user
		WHERE del_flag = 0 AND user_id = ? AND inst_id = ?
		LIMIT 1
	`, userID, institutionID).Scan(&instUserID)
	if err == sql.ErrNoRows {
		userType := 1
		res, insErr := tx.ExecContext(ctx, `
			INSERT INTO inst_user (uuid, version, user_id, inst_id, nick_name, username, avatar, mobile, is_admin, disabled, user_type, activated_status, del_flag, create_time)
			VALUES (?, 0, ?, ?, ?, ?, '', ?, 1, 0, ?, 0, 0, NOW())
		`, uuid.NewString(), userID, institutionID, nick, loginName, mobile, userType)
		if insErr != nil {
			return insErr
		}
		newInstUserID, insErr := res.LastInsertId()
		if insErr != nil {
			return insErr
		}
		instUserID = newInstUserID
	} else if err != nil {
		return err
	}

	rootDeptID, err := repo.ensureInstitutionRootDepartTx(ctx, tx, institutionID, organName)
	if err != nil {
		return err
	}
	if rootDeptID > 0 {
		if _, err := tx.ExecContext(ctx, `
			UPDATE sso_user
			SET dept_id = ?, update_time = NOW()
			WHERE id = ? AND del_flag = 0
		`, rootDeptID, userID); err != nil {
			return err
		}
		if instUserID > 0 {
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO inst_user_dept (uuid, version, inst_user_id, dept_id, del_flag, create_time)
				SELECT ?, 0, ?, ?, 0, NOW()
				FROM DUAL
				WHERE NOT EXISTS (
					SELECT 1
					FROM inst_user_dept
					WHERE inst_user_id = ? AND dept_id = ? AND del_flag = 0
				)
			`, uuid.NewString(), instUserID, rootDeptID, instUserID, rootDeptID); err != nil {
				return err
			}
		}
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO sso_user_role (user_id, role_id)
		SELECT ?, ?
		FROM DUAL
		WHERE NOT EXISTS (SELECT 1 FROM sso_user_role WHERE user_id = ? AND role_id = ?)
	`, userID, roleID, userID, roleID); err != nil {
		return err
	}

	return nil
}

func (repo *Repository) ensureInstitutionRootDepartTx(ctx context.Context, tx *sql.Tx, institutionID int64, organName string) (int64, error) {
	trimmedName := strings.TrimSpace(organName)
	if trimmedName == "" {
		trimmedName = "默认部门"
	}

	var deptID int64
	var deptName string
	err := tx.QueryRowContext(ctx, `
		SELECT id, IFNULL(depart_name, '')
		FROM sys_depart
		WHERE org_id = ? AND pid = 0 AND del_flag = 0
		ORDER BY id ASC
		LIMIT 1
	`, institutionID).Scan(&deptID, &deptName)
	if err == sql.ErrNoRows {
		res, insErr := tx.ExecContext(ctx, `
			INSERT INTO sys_depart (depart_name, depart_code, depart_man, depart_concat, org_id, pid, is_enable, sort, remark, del_flag, create_time)
			VALUES (?, '', '', '', ?, 0, 1, 1, '', 0, NOW())
		`, trimmedName, institutionID)
		if insErr != nil {
			return 0, insErr
		}
		newID, insErr := res.LastInsertId()
		if insErr != nil {
			return 0, insErr
		}
		return newID, nil
	}
	if err != nil {
		return 0, err
	}

	if strings.TrimSpace(deptName) == "" {
		if _, err := tx.ExecContext(ctx, `
			UPDATE sys_depart
			SET depart_name = ?, update_time = NOW()
			WHERE id = ? AND del_flag = 0
		`, trimmedName, deptID); err != nil {
			return 0, err
		}
	}
	return deptID, nil
}

func (repo *Repository) getScopedInstitutionModuleTx(ctx context.Context, tx *sql.Tx, moduleID int64, tenantID string) (int64, string, int, error) {
	if moduleID <= 0 {
		return 0, "", 0, fmt.Errorf("请选择开通版本")
	}
	tenantID = strings.TrimSpace(tenantID)
	query := `
		SELECT sm.id, IFNULL(sm.name, '')
		FROM sys_module sm
		WHERE sm.id = ? AND sm.del_flag = 0 AND sm.type = 1`
	args := []any{moduleID}
	if tenantID != "" {
		query += ` AND sm.tenant_id = ? AND sm.owner_type = 'tenant_package'`
		args = append(args, tenantID)
	}
	query += ` LIMIT 1`

	var scopedModuleID sql.NullInt64
	var moduleName sql.NullString
	if err := tx.QueryRowContext(ctx, query, args...).Scan(&scopedModuleID, &moduleName); err != nil {
		if err == sql.ErrNoRows {
			return 0, "", 0, fmt.Errorf("所选版本不在当前租户可用范围内")
		}
		return 0, "", 0, err
	}
	name := strings.TrimSpace(moduleName.String)
	openType := institutionModuleNameOpenType(name)
	if openType == 0 {
		openType = 2
	}
	return scopedModuleID.Int64, name, openType, nil
}

func (repo *Repository) ensureInstitutionTenantScopeTx(ctx context.Context, tx *sql.Tx, institutionID int64, tenantID string) error {
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return nil
	}
	var exists int
	if err := tx.QueryRowContext(ctx, `
		SELECT COUNT(1)
		FROM tenant_institution
		WHERE tenant_id = ? AND institution_id = ? AND del_flag = 0
	`, tenantID, institutionID).Scan(&exists); err != nil {
		return err
	}
	if exists <= 0 {
		return fmt.Errorf("机构不属于当前租户")
	}
	return nil
}

func (repo *Repository) bindInstitutionTenantTx(ctx context.Context, tx *sql.Tx, tenantID string, institutionID int64) error {
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" || institutionID <= 0 {
		return nil
	}
	_, err := tx.ExecContext(ctx, `
		INSERT INTO tenant_institution (tenant_id, institution_id, create_time, update_time, del_flag)
		VALUES (?, ?, NOW(), NOW(), 0)
		ON DUPLICATE KEY UPDATE del_flag = 0, update_time = NOW()
	`, tenantID, institutionID)
	return err
}

func (repo *Repository) CreateInstitution(ctx context.Context, input model.InstitutionMutation, creatorID *int64, tenantID string) (int64, error) {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if err = repo.ensureInstitutionLoginNameAvailableTx(ctx, tx, strings.TrimSpace(input.LoginName), nil); err != nil {
		return 0, err
	}

	nextCode, err := repo.nextInstitutionCode(ctx, tx)
	if err != nil {
		return 0, err
	}

	enabled := true
	if input.Enabled != nil {
		enabled = *input.Enabled
	}
	tenantID = strings.TrimSpace(tenantID)
	moduleID := int64(0)
	moduleName := ""
	openType := normalizeInstitutionOpenType(input.OpenType)
	if input.ModuleID != nil && *input.ModuleID > 0 {
		var lookupErr error
		moduleID, moduleName, openType, lookupErr = repo.getScopedInstitutionModuleTx(ctx, tx, *input.ModuleID, tenantID)
		if lookupErr != nil {
			return 0, lookupErr
		}
	}
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
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, ?, NOW(), ?, NOW(), 0, ?, 5)
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

	if err = repo.bindInstitutionTenantTx(ctx, tx, tenantID, id); err != nil {
		return 0, err
	}

	if err = repo.createDefaultInstitutionConfigTx(ctx, tx, id); err != nil {
		return 0, err
	}

	if err = repo.ensureInstitutionBootstrapAdminTx(
		ctx,
		tx,
		id,
		strings.TrimSpace(input.LoginName),
		strings.TrimSpace(input.Mobile),
		strings.TrimSpace(input.OrganName),
		strings.TrimSpace(input.Principal),
	); err != nil {
		return 0, err
	}

	if moduleID <= 0 {
		var lookupErr error
		moduleID, lookupErr = repo.findInstitutionVersionModuleIDTx(ctx, tx, openType)
		if lookupErr != nil {
			return 0, lookupErr
		}
	}
	_ = moduleName
	if moduleID > 0 {
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

func (repo *Repository) createDefaultInstitutionConfigTx(ctx context.Context, tx *sql.Tx, instID int64) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO inst_config (
			inst_id,
			add_import_student_rule,
			add_intention_student_rule,
			enable_classroom_teaching,
			enabled_one2one,
			enable_compose_lesson,
			enable_charge_by_hours,
			enable_by_date_lesson,
			enable_charge_by_price,
			enable_filter_holiday,
			enabled_arrears_rollcall,
			enable_by_auto_teaching,
			enable_one_to_one_schedule_limit,
			enable_schedule_conflict_continue,
			schedule_teacher_selection_range,
			enable_face_attendance_relate_teaching,
			enable_face_attendance_check_in_notice,
			enable_face_attendance_check_out_notice,
			enable_by_voice_tips,
			enable_send_face_attend_notice_to_admin,
			face_attendance_interval,
			default_class_time_record_mode,
			default_student_class_time,
			default_teacher_class_time,
			charge_by_price_default_price,
			enable_limit_single_order_arrears_deduct,
			enable_hour_leave_normal_record,
			enable_hour_truancy_normal_record,
			enable_period_makeup,
			enable_period_auto_finish_when_zero,
			enable_price_leave_normal_record,
			enable_price_truancy_normal_record,
			enable_price_makeup,
			enable_collector_staff,
			enable_phone_sell_staff,
			enable_foreground,
			enable_vice_sell_staff,
			enable_advisor,
			enable_student_manager,
			limit_same_weChat,
			limit_import_same_weChat,
			enable_public_pool,
			auto_send_birthday_message,
			enable_recharge_account_change_message,
			enabled_class_reminder,
			enabled_class_consumption_reminder,
			enable_audition_sms_remind,
			enable_send_coupon_remind_sms,
			enable_send_child_bind_notice_to_admin,
			enable_teaching_bill_remind_sms,
			student_absent_class_switch,
			enabled_renew_reminder,
			enable_arrearaged_send_message,
			enable_liquidation_remind_message,
			enable_point_change_remind_message,
			enable_org_send_child_bind_notice_to_admin,
			send_class_reminder_msg_hour,
			enable_leave_apply_number_limit,
			leave_apply_cycle_limit,
			leave_apply_number_limit,
			leave_apply_type_limit,
			enable_leave_apply_time_limit,
			leave_apply_time_limit,
			enable_renew_class_num,
			renew_class_num,
			enable_renew_validity_day,
			renew_validity_day,
			enable_renew_price,
			renew_price,
			del_flag,
			create_time,
			version
		)
		SELECT ?, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 'all', 0, 0, 0, 0, 0, '1', 1, 1.00, 0.00, 100.00, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, '19:00', 0, 'month', '2', 'course', 0, '1.0', 0, '5', 0, '15', 0, '500', 0, NOW(), 0
		FROM DUAL
		WHERE NOT EXISTS (
			SELECT 1
			FROM inst_config
			WHERE inst_id = ? AND del_flag = 0
		)
	`, instID, instID)
	return err
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
		_ = tx.Rollback()
	}()

	var currentLoginName string
	var currentOpenType sql.NullInt64
	var currentOpenDuration sql.NullString
	var currentExpireStart sql.NullTime
	var currentExpireEnd sql.NullTime
	if err = tx.QueryRowContext(ctx, `
		SELECT IFNULL(login_name, ''),
		       IFNULL(open_type, 2),
		       IFNULL(open_duration, ''),
		       expire_start_time,
		       expire_end_time
		FROM org_institution
		WHERE id = ? AND del_flag = 0
		LIMIT 1
	`, *input.ID).Scan(&currentLoginName, &currentOpenType, &currentOpenDuration, &currentExpireStart, &currentExpireEnd); err != nil {
		return err
	}
	currentLoginName = strings.TrimSpace(currentLoginName)
	requestedLoginName := strings.TrimSpace(input.LoginName)
	if requestedLoginName != "" && requestedLoginName != currentLoginName {
		return fmt.Errorf("编辑机构不支持修改登录账号")
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
		currentLoginName,
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

	updatedMobile := strings.TrimSpace(input.Mobile)
	if updatedMobile != "" {
		var loginUserID int64
		err = tx.QueryRowContext(ctx, `
			SELECT iu.user_id
			FROM inst_user iu
			JOIN sso_user su ON su.id = iu.user_id AND su.del_flag = 0
			WHERE iu.inst_id = ?
			  AND iu.del_flag = 0
			  AND su.username = ?
			ORDER BY iu.is_admin DESC, iu.id ASC
			LIMIT 1
		`, *input.ID, currentLoginName).Scan(&loginUserID)
		if err != nil && err != sql.ErrNoRows {
			return err
		}
		if err == nil && loginUserID > 0 {
			var conflictEmployeeCount int
			if err = tx.QueryRowContext(ctx, `
				SELECT COUNT(1)
				FROM inst_user iu
				LEFT JOIN sso_user su ON su.id = iu.user_id AND su.del_flag = 0
				WHERE iu.inst_id = ?
				  AND iu.del_flag = 0
				  AND iu.is_admin = 0
				  AND iu.user_id <> ?
				  AND (
				      iu.mobile = ?
				      OR IFNULL(su.mobile, '') = ?
				  )
			`, *input.ID, loginUserID, updatedMobile, updatedMobile).Scan(&conflictEmployeeCount); err != nil {
				return err
			}
			if conflictEmployeeCount > 0 {
				return fmt.Errorf("该手机号已被本机构员工使用，不能设置为超级管理员手机号")
			}

			if _, err = tx.ExecContext(ctx, `
				UPDATE inst_user
				SET mobile = ?,
				    update_time = NOW()
				WHERE inst_id = ?
				  AND user_id = ?
				  AND del_flag = 0
			`, updatedMobile, *input.ID, loginUserID); err != nil {
				return err
			}

			if _, err = tx.ExecContext(ctx, `
				UPDATE sso_user
				SET mobile = ?,
				    update_time = NOW()
				WHERE id = ?
				  AND del_flag = 0
			`, updatedMobile, loginUserID); err != nil {
				return err
			}
		} else if err == sql.ErrNoRows {
			return fmt.Errorf("机构登录账号未绑定到超级管理员，无法同步管理员手机号")
		}
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
		SELECT r.id,
		       r.institution_id,
		       IFNULL(r.before_open_type, 2),
		       IFNULL(r.before_open_duration, ''),
		       IFNULL(DATE_FORMAT(r.before_expire_end_time, '%Y-%m-%d %H:%i:%s'), ''),
		       IFNULL(r.after_open_type, 2),
		       IFNULL(r.renew_duration, ''),
		       IFNULL(DATE_FORMAT(r.renew_start_time, '%Y-%m-%d %H:%i:%s'), ''),
		       IFNULL(DATE_FORMAT(r.after_expire_end_time, '%Y-%m-%d %H:%i:%s'), ''),
		       IFNULL(r.operator_id, 0),
		       COALESCE(NULLIF(TRIM(u.nick_name), ''), NULLIF(TRIM(u.username), ''), ''),
		       IF(tu.user_role = 'tenant_admin', 1, 0),
		       IFNULL(DATE_FORMAT(r.create_time, '%Y-%m-%d %H:%i:%s'), '')
		FROM org_institution_renewal_record r
		LEFT JOIN sso_user u ON u.id = r.operator_id AND u.del_flag = 0
		LEFT JOIN tenant_institution ti ON ti.institution_id = r.institution_id AND ti.del_flag = 0
		LEFT JOIN tenant_user tu ON tu.tenant_id = ti.tenant_id AND tu.user_id = r.operator_id AND tu.del_flag = 0
		WHERE r.institution_id = ? AND r.del_flag = 0
		ORDER BY id DESC
	`, institutionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.InstitutionRenewalRecord, 0, 16)
	for rows.Next() {
		var item model.InstitutionRenewalRecord
		var isTenantOperator int
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
			&item.OperatorName,
			&isTenantOperator,
			&item.CreateTime,
		); err != nil {
			return nil, err
		}
		item.IsTenantOperator = isTenantOperator == 1
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
		       IF(tu.user_role = 'tenant_admin', 1, 0),
		       IFNULL(DATE_FORMAT(r.create_time, '%Y-%m-%d %H:%i:%s'), '')
		FROM org_institution_version_change_record r
		LEFT JOIN sso_user u ON u.id = r.operator_id AND u.del_flag = 0
		LEFT JOIN tenant_institution ti ON ti.institution_id = r.institution_id AND ti.del_flag = 0
		LEFT JOIN tenant_user tu ON tu.tenant_id = ti.tenant_id AND tu.user_id = r.operator_id AND tu.del_flag = 0
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
		var isTenantOperator int
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
			&isTenantOperator,
			&item.CreateTime,
		); err != nil {
			return nil, err
		}
		item.IsTenantOperator = isTenantOperator == 1
		items = append(items, item)
	}

	return items, rows.Err()
}

func (repo *Repository) RenewInstitution(ctx context.Context, input model.InstitutionRenewalMutation, operatorID *int64, tenantID string) (model.InstitutionRenewalResult, error) {
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

	if err = repo.ensureInstitutionTenantScopeTx(ctx, tx, *input.InstitutionID, tenantID); err != nil {
		return model.InstitutionRenewalResult{}, err
	}

	beforeOpenType := institutionStoredOpenType(currentOpenType)
	if beforeOpenType == 0 {
		beforeOpenType = 2
	}
	beforeOpenDuration := normalizeInstitutionOpenDuration(beforeOpenType, currentOpenDuration.String)
	moduleID := int64(0)
	moduleName := ""
	nextOpenType := normalizeInstitutionOpenType(input.OpenType)
	if input.ModuleID != nil && *input.ModuleID > 0 {
		var lookupErr error
		moduleID, moduleName, nextOpenType, lookupErr = repo.getScopedInstitutionModuleTx(ctx, tx, *input.ModuleID, tenantID)
		if lookupErr != nil {
			return model.InstitutionRenewalResult{}, lookupErr
		}
	}
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

	if moduleID <= 0 {
		var lookupErr error
		moduleID, lookupErr = repo.findInstitutionVersionModuleIDTx(ctx, tx, nextOpenType)
		if lookupErr != nil {
			return model.InstitutionRenewalResult{}, lookupErr
		}
	}
	if moduleID > 0 {
		if strings.TrimSpace(moduleName) == "" {
			moduleName, _ = repo.getModuleNameTx(ctx, tx, moduleID)
		}
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
		ModuleID:        moduleID,
		ModuleName:      moduleName,
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
			institution_id, description, business_time, video, gallery_images, login_slug, login_brand_config,
			create_id, create_time, update_id, update_time, del_flag
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, NOW(), ?, NOW(), 0)
		ON DUPLICATE KEY UPDATE
			description = VALUES(description),
			business_time = VALUES(business_time),
			video = VALUES(video),
			gallery_images = VALUES(gallery_images),
			login_slug = VALUES(login_slug),
			login_brand_config = VALUES(login_brand_config),
			update_id = VALUES(update_id),
			update_time = NOW(),
			del_flag = 0
	`,
		institutionID,
		profile.Description,
		profile.BusinessTime,
		profile.Video,
		marshalStringSlice(profile.GalleryImages),
		profile.LoginSlug,
		marshalLoginBrandConfig(profile.LoginBrand),
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

func (repo *Repository) GetModuleDetail(ctx context.Context, moduleID int64, tenantID string) (model.ModuleDetailVO, error) {
	tenantID = strings.TrimSpace(tenantID)
	ownerFilter := "m.tenant_id = 'platform'"
	args := []any{moduleID}
	if tenantID == "*" {
		ownerFilter = "1 = 1"
	} else if tenantID != "" {
		ownerFilter = "m.tenant_id = ?"
		args = []any{tenantID, moduleID}
	}

	row := repo.db.QueryRowContext(ctx, `
		SELECT m.id,
		       IFNULL(m.tenant_id, ''),
		       IFNULL(m.owner_type, ''),
		       IFNULL(m.source_module_id, 0),
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
		WHERE `+ownerFilter+` AND m.id = ? AND m.del_flag = 0
		GROUP BY m.id, m.tenant_id, m.owner_type, m.source_module_id, m.uuid, m.version, m.name, m.type, m.price, m.remark, m.create_time, m.update_time
		LIMIT 1
	`, args...)

	var detail model.ModuleDetailVO
	if err := row.Scan(
		&detail.ModuleID,
		&detail.TenantID,
		&detail.OwnerType,
		&detail.SourceModuleID,
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

	menuScopeTenantID := tenantID
	if tenantID == "*" {
		menuScopeTenantID = strings.TrimSpace(detail.TenantID)
		if menuScopeTenantID == "platform" {
			menuScopeTenantID = ""
		}
	}
	rawMenus, err := repo.listInstitutionMenus(ctx, menuScopeTenantID)
	if err != nil {
		return model.ModuleDetailVO{}, err
	}
	detail.MenuIDs = buildVisibleInstitutionModuleTree(rawMenus, selected)
	detail.MenuCount = countVisibleInstitutionModuleLeaves(rawMenus, selected)
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

func (repo *Repository) ListModuleMenuTree(ctx context.Context, moduleType int, tenantID string) ([]model.ModuleMenu, error) {
	_ = moduleType
	rawMenus, err := repo.listInstitutionMenus(ctx, strings.TrimSpace(tenantID))
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

func (repo *Repository) listInstitutionMenus(ctx context.Context, tenantID string) ([]rawMenu, error) {
	tenantID = strings.TrimSpace(tenantID)
	joinClause := ""
	args := []any{}
	if tenantID != "" {
		joinClause = `JOIN (
			SELECT DISTINCT smm.menu_id
			FROM tenant_module tm
			JOIN sys_module m ON m.id = tm.module_id AND m.del_flag = 0
			JOIN sys_module_menu smm ON smm.module_id = tm.module_id AND smm.del_flag = 0
			WHERE tm.tenant_id = ? AND tm.del_flag = 0
		) allowed_menu ON allowed_menu.menu_id = sm.id`
		args = append(args, tenantID)
	}
	rows, err := repo.db.QueryContext(ctx, `
		SELECT sm.id, IFNULL(sm.menu_name, ''), IFNULL(sm.menu_code, ''), IFNULL(sm.pid, 0), IFNULL(sm.sort, 0), IFNULL(sm.weight, 0), IFNULL(sm.menu_type, 0), IFNULL(sm.group_code, ''), IFNULL(sm.introduce, '')
		FROM sso_menu sm
		`+joinClause+`
		WHERE sm.del_flag = 0 AND sm.own_type = 2
		ORDER BY IFNULL(sm.level, 0) ASC, IFNULL(sm.sort, 0) ASC, IFNULL(sm.weight, 0) DESC, sm.id ASC
	`, args...)
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
	tenantID := firstNonEmpty(input.TenantID, "platform")
	ownerType := firstNonEmpty(input.OwnerType, "platform_template")
	result, err := repo.db.ExecContext(ctx, `
		INSERT INTO sys_module (uuid, tenant_id, owner_type, source_module_id, name, type, price, remark, del_flag, create_time, update_time, version)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, 0, NOW(), NOW(), 0)
	`, uuid.NewString(), tenantID, ownerType, nullableInt64Value(input.SourceModuleID), strings.TrimSpace(input.Name), input.Type, input.Price, strings.TrimSpace(input.Remark))
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
	tenantID := strings.TrimSpace(input.TenantID)
	whereTenant := "tenant_id = 'platform'"
	args := []any{strings.TrimSpace(input.Name), input.Type, input.Price, strings.TrimSpace(input.Remark), *input.ID}
	if tenantID != "" {
		whereTenant = "tenant_id = ?"
		args = []any{strings.TrimSpace(input.Name), input.Type, input.Price, strings.TrimSpace(input.Remark), tenantID, *input.ID}
	}
	_, err := repo.db.ExecContext(ctx, `
		UPDATE sys_module
		SET name = ?, type = ?, price = ?, remark = ?, update_time = NOW()
		WHERE `+whereTenant+` AND id = ? AND del_flag = 0
	`, args...)
	return err
}

func normalizeInstitutionIDs(institutionIDs []int64) []int64 {
	if len(institutionIDs) == 0 {
		return []int64{}
	}
	seen := make(map[int64]struct{}, len(institutionIDs))
	normalized := make([]int64, 0, len(institutionIDs))
	for _, institutionID := range institutionIDs {
		if institutionID <= 0 {
			continue
		}
		if _, exists := seen[institutionID]; exists {
			continue
		}
		seen[institutionID] = struct{}{}
		normalized = append(normalized, institutionID)
	}
	return normalized
}

func (repo *Repository) ReplaceInstitutionModule(ctx context.Context, input model.InstitutionPermissionMutation, operatorID *int64, tenantID string) error {
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
		_ = tx.Rollback()
	}()

	if err := repo.ensureInstitutionTenantScopeTx(ctx, tx, *input.InstitutionID, tenantID); err != nil {
		return err
	}
	moduleID, _, _, err := repo.getScopedInstitutionModuleTx(ctx, tx, *input.ModuleID, tenantID)
	if err != nil {
		return err
	}

	if err := repo.replaceInstitutionModuleTx(ctx, tx, *input.InstitutionID, moduleID, input.MenuIDs, operatorID); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (repo *Repository) ReplaceInstitutionModulesBatch(ctx context.Context, input model.InstitutionPermissionBatchMutation, operatorID *int64, tenantID string) error {
	if input.ModuleID == nil || *input.ModuleID <= 0 {
		return fmt.Errorf("moduleId is required")
	}

	institutionIDs := normalizeInstitutionIDs(input.InstitutionIDs)
	if len(institutionIDs) == 0 {
		return fmt.Errorf("institutionIds is required")
	}

	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	moduleID, _, _, err := repo.getScopedInstitutionModuleTx(ctx, tx, *input.ModuleID, tenantID)
	if err != nil {
		return err
	}

	for _, institutionID := range institutionIDs {
		if err := repo.ensureInstitutionTenantScopeTx(ctx, tx, institutionID, tenantID); err != nil {
			return fmt.Errorf("机构 %d 处理失败: %w", institutionID, err)
		}
		if err := repo.replaceInstitutionModuleTx(ctx, tx, institutionID, moduleID, input.MenuIDs, operatorID); err != nil {
			return fmt.Errorf("机构 %d 处理失败: %w", institutionID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (repo *Repository) replaceInstitutionModuleTx(ctx context.Context, tx *sql.Tx, institutionID, moduleID int64, customMenuIDs []int64, operatorID *int64) error {
	var currentOpenType sql.NullInt64
	var currentOpenDuration sql.NullString
	var expireEnd sql.NullTime
	var statusValue sql.NullInt64
	if err := tx.QueryRowContext(ctx, `
		SELECT IFNULL(open_type, 2),
		       IFNULL(open_duration, ''),
		       expire_end_time,
		       IFNULL(status, 0)
		FROM org_institution
		WHERE id = ? AND del_flag = 0
		LIMIT 1
		FOR UPDATE
	`, institutionID).Scan(&currentOpenType, &currentOpenDuration, &expireEnd, &statusValue); err != nil {
		return err
	}

	currentModuleID, currentModuleName, lookupCurrentModuleErr := repo.getInstitutionCurrentModuleTx(ctx, tx, institutionID)
	if lookupCurrentModuleErr != nil {
		return lookupCurrentModuleErr
	}

	moduleName, lookupErr := repo.getModuleNameTx(ctx, tx, moduleID)
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
		if _, err := tx.ExecContext(ctx, `
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
			institutionID,
		); err != nil {
			return err
		}
	}

	if err := repo.bindInstitutionModuleTx(
		ctx,
		tx,
		institutionID,
		moduleID,
		expireEnd,
		int(statusValue.Int64),
		operatorID,
		customMenuIDs,
		true,
		true,
	); err != nil {
		return err
	}

	shouldWriteVersionChangeLog := currentModuleID != moduleID || (nextOpenType > 0 && storedOpenType != nextOpenType)
	if shouldWriteVersionChangeLog {
		beforeVersionName := strings.TrimSpace(currentModuleName)
		if beforeVersionName == "" {
			beforeVersionName = institutionOpenTypeModuleName(storedOpenType)
		}
		afterOpenType := nextOpenType
		if afterOpenType == 0 {
			afterOpenType = storedOpenType
		}
		if err := repo.insertInstitutionVersionChangeRecordTx(ctx, tx, model.InstitutionVersionChangeRecord{
			InstitutionID:     institutionID,
			BeforeOpenType:    storedOpenType,
			BeforeModuleID:    currentModuleID,
			BeforeVersionName: beforeVersionName,
			AfterOpenType:     afterOpenType,
			AfterModuleID:     moduleID,
			AfterVersionName:  moduleName,
			OperatorID:        nullableInt64Deref(operatorID),
		}); err != nil {
			return err
		}
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
			return fmt.Errorf("超级管理员角色不存在，请先初始化机构账号")
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
