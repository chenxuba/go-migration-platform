package service

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
	"go-migration-platform/services/education/internal/model"
)

func (svc *Service) ParseLessonHourOrderImportFile(userID int64, filename string, reader io.Reader) (model.IntentionStudentImportParseResult, error) {
	return svc.ParseOrderImportFile(userID, filename, reader)
}

func (svc *Service) ParseOrderImportFile(userID int64, filename string, reader io.Reader) (model.IntentionStudentImportParseResult, error) {
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
	customFields, err := svc.repo.ListStudentFields(context.Background(), instID, false)
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

	headerRow := rows[0]
	importMode := detectOrderImportModeByTitles(headerRow)
	switch importMode {
	case orderImportModeUnknown:
		return model.IntentionStudentImportParseResult{}, errors.New("未识别到可导入字段，请使用最新模板")
	case orderImportModeAmount:
		return model.IntentionStudentImportParseResult{}, errors.New("当前仅支持按课时、按时段订单模板导入")
	}

	courseNames, err := svc.loadOrderImportCourseNames(context.Background(), instID, orderImportModeLabel(importMode))
	if err != nil {
		return model.IntentionStudentImportParseResult{}, err
	}

	var templateColumns []model.IntentionStudentImportTemplateColumn
	switch importMode {
	case orderImportModeTimeSlot:
		templateColumns = buildTimeSlotOrderImportColumns(defaultFields, customFields, channels, courseNames, staffNames, orderTagNames)
	default:
		templateColumns = buildLessonHourOrderImportColumns(defaultFields, customFields, channels, courseNames, staffNames, orderTagNames)
	}

	columnMap := make(map[string]model.IntentionStudentImportTemplateColumn, len(templateColumns))
	for _, column := range templateColumns {
		columnMap[strings.TrimSpace(column.Title)] = column
	}

	columns := make([]model.IntentionStudentImportColumn, 0, len(headerRow))
	headerIndexes := make([]int, 0, len(headerRow))
	for idx, item := range headerRow {
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
		applyOrderImportRowValidation(importMode, cells, &hasError)
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

func applyOrderImportRowValidation(mode orderImportMode, cells []model.IntentionStudentImportCell, hasError *bool) {
	if len(cells) == 0 {
		return
	}
	rowData := make(map[string]*model.IntentionStudentImportCell, len(cells))
	for idx := range cells {
		rowData[cells[idx].Title] = &cells[idx]
	}

	switch mode {
	case orderImportModeTimeSlot:
		startCell, hasStart := rowData["有效开始日期"]
		endCell, hasEnd := rowData["有效结束日期(含赠送天数)"]
		if !hasStart || !hasEnd {
			return
		}
		startDate, startOK := parseImportDateValue(startCell.Value)
		endDate, endOK := parseImportDateValue(endCell.Value)
		if !startOK || !endOK {
			return
		}
		if endDate.Before(startDate) {
			endCell.Error = "结束日期不能早于开始日期"
			*hasError = true
			return
		}

		giftDays := int64(0)
		if giftCell, ok := rowData["赠送天数"]; ok {
			text := strings.TrimSpace(giftCell.Value)
			if text != "" {
				parsed, err := strconv.ParseInt(text, 10, 64)
				if err == nil && parsed >= 0 {
					giftDays = parsed
				}
			}
		}
		totalDays := int64(endDate.Sub(startDate).Hours()/24) + 1
		if totalDays-giftDays <= 0 {
			endCell.Error = "有效时段需大于赠送天数"
			*hasError = true
		}
	}
}

func validateOrderImportValue(column model.IntentionStudentImportColumn, value string) string {
	if column.Required && strings.TrimSpace(value) == "" {
		return "请填写"
	}
	if strings.TrimSpace(value) == "" {
		return ""
	}
	if len(column.Options) > 0 && !containsImportOption(column.Options, value) {
		return "请选择预设值"
	}
	if column.Title == "手机号" && !phoneDigitsPattern.MatchString(value) {
		return "手机号格式错误"
	}
	if requiresIntegerPrecision(column.Title) && !isValidIntegerNumber(value) {
		return "请输入整数"
	}
	if requiresTwoDecimalPrecision(column.Title) && !isValidTwoDecimalNumber(value) {
		return "最多保留2位小数"
	}
	switch column.FieldType {
	case 2:
		if !isNumericImportValue(value) {
			return "请输入数字"
		}
	case 3:
		if _, ok := parseImportDateValue(value); !ok {
			return "日期格式错误"
		}
	}
	return ""
}
