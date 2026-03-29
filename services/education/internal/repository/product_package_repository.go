package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"go-migration-platform/services/education/internal/model"
)

func ensureProductPackageTables(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS product_package (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			uuid VARCHAR(64) NULL,
			version BIGINT NOT NULL DEFAULT 0,
			inst_id BIGINT NOT NULL,
			name VARCHAR(120) NOT NULL DEFAULT '',
			title VARCHAR(120) NOT NULL DEFAULT '',
			online_sale TINYINT(1) NOT NULL DEFAULT 1,
			is_allow_edit_when_enroll TINYINT(1) NOT NULL DEFAULT 0,
			images LONGTEXT NOT NULL,
			description LONGTEXT NOT NULL,
			is_show_mico_school TINYINT(1) NOT NULL DEFAULT 0,
			is_online_sale_mico_school TINYINT(1) NOT NULL DEFAULT 0,
			buy_rule_json LONGTEXT NOT NULL,
			subject_ids_json LONGTEXT NOT NULL,
			org_product_package_id BIGINT NOT NULL DEFAULT 0,
			editable TINYINT(1) NOT NULL DEFAULT 1,
			is_sync_org_product_package TINYINT(1) NOT NULL DEFAULT 0,
			sale_volume INT NOT NULL DEFAULT 0,
			total_amount DECIMAL(18,2) NOT NULL DEFAULT 0,
			discount_amount DECIMAL(18,2) NOT NULL DEFAULT 0,
			final_amount DECIMAL(18,2) NOT NULL DEFAULT 0,
			create_id BIGINT NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_id BIGINT NOT NULL DEFAULT 0,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			KEY idx_product_package_inst (inst_id, del_flag),
			KEY idx_product_package_updated (inst_id, update_time, id),
			KEY idx_product_package_sale (inst_id, online_sale)
		)
	`); err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS product_package_item (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			uuid VARCHAR(64) NULL,
			version BIGINT NOT NULL DEFAULT 0,
			product_package_id BIGINT NOT NULL,
			product_type INT NOT NULL DEFAULT 1,
			product_id BIGINT NOT NULL DEFAULT 0,
			product_name VARCHAR(120) NOT NULL DEFAULT '',
			sku_id BIGINT NOT NULL DEFAULT 0,
			sku_name VARCHAR(120) NOT NULL DEFAULT '',
			sku_count DECIMAL(18,2) NOT NULL DEFAULT 0,
			free_quantity DECIMAL(18,2) NOT NULL DEFAULT 0,
			discount_type INT NULL,
			discount_number DECIMAL(18,2) NOT NULL DEFAULT 0,
			lesson_type INT NOT NULL DEFAULT 0,
			lesson_mode INT NOT NULL DEFAULT 0,
			lesson_audition TINYINT(1) NOT NULL DEFAULT 0,
			create_id BIGINT NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_id BIGINT NOT NULL DEFAULT 0,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			KEY idx_product_package_item_package (product_package_id, del_flag),
			KEY idx_product_package_item_product (product_id, sku_id)
		)
	`); err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS product_package_property_result (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			uuid VARCHAR(64) NULL,
			version BIGINT NOT NULL DEFAULT 0,
			product_package_id BIGINT NOT NULL,
			property_id BIGINT NOT NULL,
			property_value BIGINT NOT NULL,
			create_id BIGINT NOT NULL DEFAULT 0,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_id BIGINT NOT NULL DEFAULT 0,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			del_flag TINYINT(1) NOT NULL DEFAULT 0,
			KEY idx_product_package_property_package (product_package_id, del_flag),
			KEY idx_product_package_property_filter (property_id, property_value, del_flag)
		)
	`); err != nil {
		return err
	}
	if err := dropColumnIfExists(ctx, db, "product_package_item", "lesson_scope"); err != nil {
		return err
	}
	return nil
}

func buildProductPackageWhere(instID int64, filters model.ProductPackageQueryFilter) (string, []any) {
	whereParts := []string{"pp.inst_id = ?", "pp.del_flag = 0"}
	args := []any{instID}

	searchName := strings.TrimSpace(filters.Name)
	if searchName == "" {
		searchName = strings.TrimSpace(filters.SearchKey)
	}
	if searchName != "" {
		whereParts = append(whereParts, "(pp.name LIKE ? OR pp.title LIKE ?)")
		kw := "%" + searchName + "%"
		args = append(args, kw, kw)
	}
	if filters.OnlineSale != nil {
		whereParts = append(whereParts, "pp.online_sale = ?")
		args = append(args, boolValue(filters.OnlineSale))
	}
	if filters.IsOnlineSaleMicoSchool != nil {
		whereParts = append(whereParts, "pp.is_online_sale_mico_school = ?")
		args = append(args, boolValue(filters.IsOnlineSaleMicoSchool))
	}
	if filters.IsShowMicoSchool != nil {
		whereParts = append(whereParts, "pp.is_show_mico_school = ?")
		args = append(args, boolValue(filters.IsShowMicoSchool))
	}
	for _, property := range filters.ProductPackageProperties {
		propertyID := strings.TrimSpace(property.ProductPackagePropertyID)
		propertyValue := strings.TrimSpace(property.ProductPackagePropertyValue)
		if propertyID == "" || propertyValue == "" {
			continue
		}
		whereParts = append(whereParts, `
			EXISTS (
				SELECT 1
				FROM product_package_property_result ppr
				WHERE ppr.product_package_id = pp.id
				  AND ppr.del_flag = 0
				  AND CAST(ppr.property_id AS CHAR) = ?
				  AND CAST(ppr.property_value AS CHAR) = ?
			)
		`)
		args = append(args, propertyID, propertyValue)
	}
	return strings.Join(whereParts, " AND "), args
}

func (repo *Repository) PageProductPackages(ctx context.Context, instID int64, query model.ProductPackageQueryDTO) (model.ProductPackagePagedResult, error) {
	current := query.PageRequestModel.PageIndex
	size := query.PageRequestModel.PageSize
	if current <= 0 {
		current = 1
	}
	if size <= 0 {
		size = 10
	}
	offset := (current - 1) * size

	whereSQL, args := buildProductPackageWhere(instID, query.QueryModel)
	var total int
	if err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM product_package pp
		WHERE `+whereSQL, args...).Scan(&total); err != nil {
		return model.ProductPackagePagedResult{}, err
	}

	rows, err := repo.db.QueryContext(ctx, `
		SELECT
			pp.id, IFNULL(pp.name, ''), IFNULL(pp.title, ''),
			IFNULL(pp.online_sale, 0), IFNULL(pp.is_online_sale_mico_school, 0), IFNULL(pp.is_show_mico_school, 0),
			CAST(IFNULL(pp.org_product_package_id, 0) AS CHAR),
			IFNULL(pp.editable, 1), IFNULL(pp.is_sync_org_product_package, 0),
			IFNULL(pp.sale_volume, 0), IFNULL(pp.total_amount, 0), IFNULL(pp.discount_amount, 0), IFNULL(pp.final_amount, 0),
			IFNULL(pp.images, ''), pp.update_time
		FROM product_package pp
		WHERE `+whereSQL+`
		ORDER BY pp.update_time DESC, pp.id DESC
		LIMIT ? OFFSET ?
	`, append(args, size, offset)...)
	if err != nil {
		return model.ProductPackagePagedResult{}, err
	}
	defer rows.Close()

	items := make([]model.ProductPackageVO, 0, size)
	packageIDs := make([]int64, 0, size)
	indexByID := make(map[int64]int, size)
	for rows.Next() {
		var (
			item      model.ProductPackageVO
			id        int64
			updatedAt sql.NullTime
		)
		if err := rows.Scan(
			&id,
			&item.Name,
			&item.Title,
			&item.OnlineSale,
			&item.IsOnlineSaleMicoSchool,
			&item.IsShowMicoSchool,
			&item.OrgProductPackageID,
			&item.Editable,
			&item.IsSyncOrgProductPackage,
			&item.Sale,
			&item.TotalAmount,
			&item.DiscountAmount,
			&item.FinalAmount,
			&item.Images,
			&updatedAt,
		); err != nil {
			return model.ProductPackagePagedResult{}, err
		}
		item.ID = strconv.FormatInt(id, 10)
		if updatedAt.Valid {
			t := updatedAt.Time
			item.UpdatedTime = &t
		}
		items = append(items, item)
		packageIDs = append(packageIDs, id)
		indexByID[id] = len(items) - 1
	}
	if err := rows.Err(); err != nil {
		return model.ProductPackagePagedResult{}, err
	}

	if len(packageIDs) > 0 {
		if err := repo.attachProductPackageItems(ctx, packageIDs, items, indexByID); err != nil {
			return model.ProductPackagePagedResult{}, err
		}
		if err := repo.attachProductPackageProperties(ctx, packageIDs, items, indexByID); err != nil {
			return model.ProductPackagePagedResult{}, err
		}
	}

	return model.ProductPackagePagedResult{
		List:  items,
		Total: total,
	}, nil
}

func (repo *Repository) attachProductPackageItems(ctx context.Context, packageIDs []int64, items []model.ProductPackageVO, indexByID map[int64]int) error {
	holders := make([]string, 0, len(packageIDs))
	args := make([]any, 0, len(packageIDs))
	for _, id := range packageIDs {
		holders = append(holders, "?")
		args = append(args, id)
	}
	rows, err := repo.db.QueryContext(ctx, `
		SELECT
			id, product_package_id, IFNULL(product_type, 1),
			CAST(IFNULL(product_id, 0) AS CHAR), IFNULL(product_name, ''),
			CAST(IFNULL(sku_id, 0) AS CHAR), IFNULL(sku_name, ''),
			IFNULL(lesson_type, 0), IFNULL(lesson_mode, 0), IFNULL(lesson_audition, 0)
		FROM product_package_item
		WHERE del_flag = 0 AND product_package_id IN (`+strings.Join(holders, ",")+`)
		ORDER BY id ASC
	`, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			item      model.ProductPackageItemVO
			id        int64
			packageID int64
		)
		if err := rows.Scan(
			&id,
			&packageID,
			&item.ProductType,
			&item.ProductID,
			&item.ProductName,
			&item.SkuID,
			&item.SkuName,
			&item.LessonType,
			&item.LessonMode,
			&item.LessonAudition,
		); err != nil {
			return err
		}
		item.ID = strconv.FormatInt(id, 10)
		if idx, ok := indexByID[packageID]; ok {
			items[idx].Items = append(items[idx].Items, item)
		}
	}
	return rows.Err()
}

func (repo *Repository) attachProductPackageProperties(ctx context.Context, packageIDs []int64, items []model.ProductPackageVO, indexByID map[int64]int) error {
	holders := make([]string, 0, len(packageIDs))
	args := make([]any, 0, len(packageIDs))
	for _, id := range packageIDs {
		holders = append(holders, "?")
		args = append(args, id)
	}
	rows, err := repo.db.QueryContext(ctx, `
		SELECT
			ppr.product_package_id,
			CAST(ppr.property_id AS CHAR),
			IFNULL(cp.name, ''),
			CAST(ppr.property_value AS CHAR),
			IFNULL(cpo.name, '')
		FROM product_package_property_result ppr
		LEFT JOIN inst_course_property cp ON cp.id = ppr.property_id AND cp.del_flag = 0
		LEFT JOIN inst_course_property_option cpo ON cpo.id = ppr.property_value AND cpo.del_flag = 0
		WHERE ppr.del_flag = 0 AND ppr.product_package_id IN (`+strings.Join(holders, ",")+`)
		ORDER BY ppr.id ASC
	`, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			packageID int64
			item      model.ProductPackagePropertyVO
		)
		if err := rows.Scan(
			&packageID,
			&item.ProductPackagePropertyID,
			&item.ProductPackagePropertyName,
			&item.ProductPackagePropertyValue,
			&item.ProductPackagePropertyValueName,
		); err != nil {
			return err
		}
		if idx, ok := indexByID[packageID]; ok {
			items[idx].ExtendProperties = append(items[idx].ExtendProperties, item)
			if strings.TrimSpace(item.ProductPackagePropertyName) == "科目" {
				items[idx].Subjects = append(items[idx].Subjects, model.ProductPackageSubjectVO{
					ID:   item.ProductPackagePropertyValue,
					Name: item.ProductPackagePropertyValueName,
				})
			}
		}
	}
	return rows.Err()
}

func (repo *Repository) GetProductPackageStatistics(ctx context.Context, instID int64, filters model.ProductPackageQueryFilter) (model.ProductPackageStatistics, error) {
	whereSQL, args := buildProductPackageWhere(instID, filters)
	var result model.ProductPackageStatistics
	if err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM product_package pp
		WHERE `+whereSQL, args...).Scan(&result.TotalCount); err != nil {
		return result, err
	}

	onSaleWhereSQL := whereSQL + " AND pp.online_sale = 1"
	if err := repo.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM product_package pp
		WHERE `+onSaleWhereSQL, args...).Scan(&result.OnSaleCount); err != nil {
		return result, err
	}
	return result, nil
}

type productPackageSkuSnapshot struct {
	CourseID       int64
	CourseName     string
	TeachMethod    int
	QuotationID    int64
	QuotationName  string
	LessonModel    int
	LessonAudition bool
	Price          float64
}

func normalizePackageDiscountRate(value float64) float64 {
	if value <= 0 {
		return 10
	}
	if value > 10 {
		return value / 100
	}
	return value
}

func calculateProductPackageAmounts(price, skuCount float64, discountType *int, discountNumber float64) (float64, float64, float64) {
	base := roundMoney(price * skuCount)
	discountAmount := 0.0
	finalAmount := base
	if discountType != nil {
		switch *discountType {
		case 1:
			discountAmount = roundMoney(discountNumber)
			finalAmount = roundMoney(base - discountAmount)
		case 2:
			rate := normalizePackageDiscountRate(discountNumber)
			finalAmount = roundMoney(base * rate / 10)
			discountAmount = roundMoney(base - finalAmount)
		}
	}
	if finalAmount < 0 {
		finalAmount = 0
	}
	if discountAmount < 0 {
		discountAmount = 0
	}
	return base, discountAmount, finalAmount
}

func (repo *Repository) loadProductPackageSkuSnapshots(ctx context.Context, items []model.ProductPackageItemMutation) (map[string]productPackageSkuSnapshot, error) {
	uniqueSKU := make(map[string]struct{})
	skuIDs := make([]string, 0, len(items))
	for _, item := range items {
		skuID := strings.TrimSpace(item.SkuID)
		if skuID == "" {
			continue
		}
		if _, exists := uniqueSKU[skuID]; exists {
			continue
		}
		uniqueSKU[skuID] = struct{}{}
		skuIDs = append(skuIDs, skuID)
	}
	result := make(map[string]productPackageSkuSnapshot, len(skuIDs))
	if len(skuIDs) == 0 {
		return result, nil
	}

	holders := make([]string, 0, len(skuIDs))
	args := make([]any, 0, len(skuIDs))
	for _, skuID := range skuIDs {
		holders = append(holders, "?")
		args = append(args, skuID)
	}
	rows, err := repo.db.QueryContext(ctx, `
		SELECT
			c.id, IFNULL(c.name, ''), IFNULL(c.teach_method, 0),
			q.id, IFNULL(q.name, ''), IFNULL(q.lesson_model, 0), IFNULL(q.lesson_audition, 0), IFNULL(q.price, 0)
		FROM inst_course_quotation q
		INNER JOIN inst_course c ON c.id = q.course_id AND c.del_flag = 0
		WHERE q.del_flag = 0 AND CAST(q.id AS CHAR) IN (`+strings.Join(holders, ",")+`)
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var (
			item  productPackageSkuSnapshot
			skuID int64
		)
		if err := rows.Scan(
			&item.CourseID,
			&item.CourseName,
			&item.TeachMethod,
			&skuID,
			&item.QuotationName,
			&item.LessonModel,
			&item.LessonAudition,
			&item.Price,
		); err != nil {
			return nil, err
		}
		item.QuotationID = skuID
		result[strconv.FormatInt(skuID, 10)] = item
	}
	return result, rows.Err()
}

func (repo *Repository) CreateProductPackage(ctx context.Context, instID, operatorID int64, dto model.ProductPackageMutation) (int64, error) {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	buyRuleRaw, err := json.Marshal(dto.BuyRule)
	if err != nil {
		return 0, err
	}
	subjectIDsRaw, err := json.Marshal(dto.SubjectIDs)
	if err != nil {
		return 0, err
	}
	snapshots, err := repo.loadProductPackageSkuSnapshots(ctx, dto.Items)
	if err != nil {
		return 0, err
	}

	totalAmount := 0.0
	discountAmount := 0.0
	finalAmount := 0.0
	for _, item := range dto.Items {
		snapshot, ok := snapshots[strings.TrimSpace(item.SkuID)]
		if !ok {
			return 0, fmt.Errorf("报价单不存在: %s", strings.TrimSpace(item.SkuID))
		}
		base, discount, final := calculateProductPackageAmounts(snapshot.Price, item.SkuCount, item.DiscountType, item.DiscountNumber)
		totalAmount += base
		discountAmount += discount
		finalAmount += final
	}
	totalAmount = roundMoney(totalAmount)
	discountAmount = roundMoney(discountAmount)
	finalAmount = roundMoney(finalAmount)

	result, err := tx.ExecContext(ctx, `
		INSERT INTO product_package (
			uuid, version, inst_id, name, title, online_sale, is_allow_edit_when_enroll,
			images, description, is_show_mico_school, is_online_sale_mico_school, buy_rule_json, subject_ids_json,
			org_product_package_id, editable, is_sync_org_product_package, sale_volume, total_amount, discount_amount, final_amount,
			create_id, create_time, update_id, update_time, del_flag
		) VALUES (
			UUID(), 0, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, 1, 0, 0, ?, ?, ?, ?, NOW(), ?, NOW(), 0
		)
	`,
		instID,
		strings.TrimSpace(dto.Name),
		strings.TrimSpace(dto.Title),
		dto.OnlineSale,
		dto.IsAllowEditWhenEnroll,
		strings.TrimSpace(dto.Images),
		strings.TrimSpace(dto.Description),
		dto.IsShowMicoSchool,
		dto.IsOnlineSaleMicoSchool,
		string(buyRuleRaw),
		string(subjectIDsRaw),
		totalAmount,
		discountAmount,
		finalAmount,
		operatorID,
		operatorID,
	)
	if err != nil {
		return 0, err
	}
	packageID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	for _, item := range dto.Items {
		snapshot := snapshots[strings.TrimSpace(item.SkuID)]
		_, err := tx.ExecContext(ctx, `
			INSERT INTO product_package_item (
				uuid, version, product_package_id, product_type, product_id, product_name, sku_id, sku_name,
				sku_count, free_quantity, discount_type, discount_number, lesson_type, lesson_mode, lesson_audition,
				create_id, create_time, update_id, update_time, del_flag
			) VALUES (
				UUID(), 0, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), ?, NOW(), 0
			)
		`,
			packageID,
			1,
			snapshot.CourseID,
			snapshot.CourseName,
			snapshot.QuotationID,
			snapshot.QuotationName,
			item.SkuCount,
			item.FreeQuantity,
			item.DiscountType,
			item.DiscountNumber,
			snapshot.TeachMethod,
			snapshot.LessonModel,
			snapshot.LessonAudition,
			operatorID,
			operatorID,
		)
		if err != nil {
			return 0, err
		}
	}

	for _, property := range dto.ProductPackageProperties {
		propertyID, err := strconv.ParseInt(strings.TrimSpace(property.ProductPackagePropertyID), 10, 64)
		if err != nil || propertyID <= 0 {
			continue
		}
		propertyValue, err := strconv.ParseInt(strings.TrimSpace(property.ProductPackagePropertyValue), 10, 64)
		if err != nil || propertyValue <= 0 {
			continue
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO product_package_property_result (
				uuid, version, product_package_id, property_id, property_value,
				create_id, create_time, update_id, update_time, del_flag
			) VALUES (
				UUID(), 0, ?, ?, ?, ?, NOW(), ?, NOW(), 0
			)
		`, packageID, propertyID, propertyValue, operatorID, operatorID); err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return packageID, nil
}

func (repo *Repository) UpdateProductPackageSaleStatus(ctx context.Context, instID int64, id string, onlineSale bool) error {
	_, err := repo.db.ExecContext(ctx, `
		UPDATE product_package
		SET online_sale = ?, update_time = NOW()
		WHERE CAST(id AS CHAR) = ? AND inst_id = ? AND del_flag = 0
	`, onlineSale, strings.TrimSpace(id), instID)
	return err
}

func (repo *Repository) UpdateProductPackageMicroSchoolRules(ctx context.Context, instID int64, id string, isShow, isSale bool) error {
	_, err := repo.db.ExecContext(ctx, `
		UPDATE product_package
		SET is_show_mico_school = ?, is_online_sale_mico_school = ?, update_time = NOW()
		WHERE CAST(id AS CHAR) = ? AND inst_id = ? AND del_flag = 0
	`, isShow, isSale, strings.TrimSpace(id), instID)
	return err
}

func (repo *Repository) UpdateProductPackageAllowEditWhenEnroll(ctx context.Context, instID int64, id string, allow bool) error {
	_, err := repo.db.ExecContext(ctx, `
		UPDATE product_package
		SET is_allow_edit_when_enroll = ?, update_time = NOW()
		WHERE CAST(id AS CHAR) = ? AND inst_id = ? AND del_flag = 0
	`, allow, strings.TrimSpace(id), instID)
	return err
}

func (repo *Repository) DeleteProductPackage(ctx context.Context, instID int64, id string) error {
	_, err := repo.db.ExecContext(ctx, `
		UPDATE product_package
		SET del_flag = 1, update_time = NOW()
		WHERE CAST(id AS CHAR) = ? AND inst_id = ? AND del_flag = 0
	`, strings.TrimSpace(id), instID)
	return err
}
