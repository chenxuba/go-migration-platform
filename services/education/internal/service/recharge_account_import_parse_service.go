package service

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
	"go-migration-platform/services/education/internal/model"
)

func (svc *Service) ParseRechargeAccountImportFile(userID int64, filename string, reader io.Reader) (model.IntentionStudentImportParseResult, error) {
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.IntentionStudentImportParseResult{}, errors.New("no institution context")
		}
		return model.IntentionStudentImportParseResult{}, err
	}

	orgName, err := svc.repo.GetInstitutionName(context.Background(), instID)
	if err != nil {
		return model.IntentionStudentImportParseResult{}, err
	}
	if strings.TrimSpace(orgName) == "" {
		orgName = "总校区"
	}

	channels, err := svc.repo.GetChannels(context.Background(), instID)
	if err != nil {
		return model.IntentionStudentImportParseResult{}, err
	}
	defaultFields, err := svc.repo.ListStudentFields(context.Background(), instID, true)
	if err != nil {
		return model.IntentionStudentImportParseResult{}, err
	}
	staffNames, err := svc.repo.ListActiveStaffNames(context.Background(), instID)
	if err != nil {
		return model.IntentionStudentImportParseResult{}, err
	}
	orderTagNames, err := svc.repo.ListEnabledOrderTagNames(context.Background(), instID)
	if err != nil {
		return model.IntentionStudentImportParseResult{}, err
	}

	raw, err := io.ReadAll(reader)
	if err != nil {
		return model.IntentionStudentImportParseResult{}, err
	}
	file, err := excelize.OpenReader(bytes.NewReader(raw))
	if err != nil {
		return model.IntentionStudentImportParseResult{}, errors.New("文件解析失败，请上传有效的 xlsx 文件")
	}
	defer file.Close()

	sheetName := file.GetSheetName(0)
	rows, err := file.GetRows(sheetName)
	if err != nil {
		return model.IntentionStudentImportParseResult{}, err
	}
	if len(rows) == 0 {
		return model.IntentionStudentImportParseResult{}, errors.New("导入文件为空")
	}

	importMode := detectRechargeAccountImportModeByColumns(rows[0])
	if importMode == rechargeAccountImportModeUnknown {
		return model.IntentionStudentImportParseResult{}, errors.New("未识别到可导入字段，请使用最新模板")
	}

	var templateColumns []model.IntentionStudentImportTemplateColumn
	if importMode == rechargeAccountImportModeByAccount {
		templateColumns = buildRechargeAccountImportByAccountColumns(defaultFields, channels, staffNames, orderTagNames)
	} else {
		templateColumns = buildRechargeAccountImportByStudentColumns(defaultFields, channels, staffNames, orderTagNames)
	}

	columnMap := make(map[string]model.IntentionStudentImportTemplateColumn, len(templateColumns))
	for _, column := range templateColumns {
		columnMap[strings.TrimSpace(column.Title)] = column
	}

	columns := make([]model.IntentionStudentImportColumn, 0, len(rows[0]))
	headerIndexes := make([]int, 0, len(rows[0]))
	for idx, item := range rows[0] {
		title := strings.TrimSpace(strings.TrimPrefix(item, "*"))
		column, ok := columnMap[title]
		if !ok || title == "填写说明" {
			continue
		}
		columns = append(columns, model.IntentionStudentImportColumn{
			Key:       buildIntentionStudentImportColumnKey(title, len(columns)),
			Title:     title,
			Required:  column.Required,
			FieldType: column.FieldType,
			FieldID:   column.FieldID,
			Options:   column.Options,
		})
		headerIndexes = append(headerIndexes, idx)
	}
	if len(columns) == 0 {
		return model.IntentionStudentImportParseResult{}, errors.New("未识别到可导入字段，请使用最新模板")
	}

	result := model.IntentionStudentImportParseResult{
		ImportID: time.Now().Format("20060102150405") + fmt.Sprintf("%09d", time.Now().UnixNano()%1e9),
		FileName: strings.TrimSpace(filename),
		InstName: orgName,
		Columns:  columns,
		Rows:     make([]model.IntentionStudentImportRow, 0, len(rows)),
	}

	defaultBusinessDate := time.Now().Format("2006-01-02")
	for rowIdx := 1; rowIdx < len(rows); rowIdx++ {
		rawRow := rows[rowIdx]
		cells := make([]model.IntentionStudentImportCell, 0, len(columns))
		hasRawValue := false
		hasError := false
		for colIdx, column := range columns {
			value := ""
			sourceIndex := headerIndexes[colIdx]
			if sourceIndex < len(rawRow) {
				value = strings.TrimSpace(rawRow[sourceIndex])
			}
			if column.FieldType == 3 {
				if cellName, err := excelize.CoordinatesToCellName(sourceIndex+1, rowIdx+1); err == nil {
					if rawValue, err := file.GetCellValue(sheetName, cellName, excelize.Options{RawCellValue: true}); err == nil {
						rawValue = strings.TrimSpace(rawValue)
						if rawValue != "" {
							value = rawValue
						}
					}
				}
				value = normalizeImportDateText(value)
			}
			if strings.TrimSpace(value) != "" {
				hasRawValue = true
			}
			switch column.Title {
			case "经办日期":
				if strings.TrimSpace(value) == "" {
					value = defaultBusinessDate
				}
			case "收款方式":
				if strings.TrimSpace(value) == "其他" {
					value = "其他方式"
				}
				if strings.TrimSpace(value) == "" {
					value = "其他方式"
				}
			case "收款账户":
				if strings.TrimSpace(value) == "" {
					value = "默认账户"
				}
			}
			errText := validateOrderImportValue(column, value)
			if errText != "" {
				hasError = true
			}
			cells = append(cells, model.IntentionStudentImportCell{
				Key:   column.Key,
				Title: column.Title,
				Value: value,
				Error: errText,
			})
		}
		if !hasRawValue {
			continue
		}
		applyRechargeAccountImportRowValidation(importMode, cells, &hasError)
		result.Rows = append(result.Rows, model.IntentionStudentImportRow{
			ID:       fmt.Sprintf("%s_%d", result.ImportID, rowIdx+1),
			RowNo:    rowIdx + 1,
			HasError: hasError,
			Cells:    cells,
		})
		if hasError {
			result.AbnormalCount++
		} else {
			result.NormalCount++
		}
	}

	if len(result.Rows) == 0 {
		return model.IntentionStudentImportParseResult{}, errors.New("请勿上传空文件")
	}
	return result, nil
}

func applyRechargeAccountImportRowValidation(mode rechargeAccountImportMode, cells []model.IntentionStudentImportCell, hasError *bool) {
	if len(cells) == 0 {
		return
	}
	rowData := make(map[string]*model.IntentionStudentImportCell, len(cells))
	for idx := range cells {
		rowData[cells[idx].Title] = &cells[idx]
	}

	if mode == rechargeAccountImportModeByAccount {
		if accountCell, ok := rowData["储值账户号"]; ok && strings.TrimSpace(accountCell.Value) == "" {
			accountCell.Error = "请填写"
			*hasError = true
		}
	}

	rechargeCell, hasRecharge := rowData["充值金额"]
	residualCell, hasResidual := rowData["残联金额"]
	givingCell, hasGiving := rowData["赠送金额"]
	if !hasRecharge || !hasResidual || !hasGiving {
		return
	}

	rechargeAmount, rechargeOK := parseOrderImportValidationFloat(rechargeCell.Value)
	residualAmount, residualOK := parseOrderImportValidationFloat(residualCell.Value)
	givingAmount, givingOK := parseOrderImportValidationFloat(givingCell.Value)
	if !rechargeOK || !residualOK || !givingOK {
		return
	}
	if rechargeAmount <= 0 && residualAmount <= 0 && givingAmount <= 0 {
		rechargeCell.Error = "充值金额、残联金额、赠送金额至少填写一项"
		residualCell.Error = "充值金额、残联金额、赠送金额至少填写一项"
		givingCell.Error = "充值金额、残联金额、赠送金额至少填写一项"
		*hasError = true
	}
}
