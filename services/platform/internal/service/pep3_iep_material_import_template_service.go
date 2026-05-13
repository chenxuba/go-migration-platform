package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
	"go-migration-platform/services/platform/internal/model"
)

const pep3IEPMaterialImportRowLimit = 1000

type pep3IEPImportDomainOption struct {
	Code string
	Name string
}

type pep3IEPImportLookups struct {
	Domains           []pep3IEPImportDomainOption
	DomainByName      map[string]pep3IEPImportDomainOption
	Questions         []model.ScaleQuestionBankItem
	QuestionByName    map[string]model.ScaleQuestionBankItem
	QuestionsByDomain map[string][]model.ScaleQuestionBankItem
}

func (svc *Service) BuildPlatformPEP3IEPMaterialImportTemplate() (string, error) {
	lookups, err := svc.loadPEP3IEPImportLookups()
	if err != nil {
		return "", err
	}
	columns := buildPEP3IEPImportTemplateColumns(lookups)
	content, err := buildPEP3IEPImportTemplateWorkbook(columns, lookups)
	if err != nil {
		return "", err
	}
	filename := sanitizePlatformImportFileName(fmt.Sprintf("PEP3-IEP素材库导入模板-%s.xlsx", time.Now().Format("20060102")))
	ticket := savePlatformTemplateDownloadFile(platformTemplateDownloadFile{
		Filename:    filename,
		ContentType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		Data:        content,
		ExpiresAt:   time.Now().Add(30 * time.Minute),
	})
	return "/api/v1/platform/scales/pep3-iep-material/import-template/file?ticket=" + url.QueryEscape(ticket), nil
}

func (svc *Service) LoadPlatformPEP3IEPMaterialImportTemplate(ticket string) (string, string, []byte, bool) {
	file, ok := loadPlatformTemplateDownloadFile(ticket)
	if !ok {
		return "", "", nil, false
	}
	return file.Filename, file.ContentType, file.Data, true
}

func (svc *Service) ParsePlatformPEP3IEPMaterialImportFile(filename string, reader io.Reader) (model.PEP3IEPMaterialImportParseResult, error) {
	lookups, err := svc.loadPEP3IEPImportLookups()
	if err != nil {
		return model.PEP3IEPMaterialImportParseResult{}, err
	}

	raw, err := io.ReadAll(reader)
	if err != nil {
		return model.PEP3IEPMaterialImportParseResult{}, err
	}
	file, err := excelize.OpenReader(bytes.NewReader(raw))
	if err != nil {
		return model.PEP3IEPMaterialImportParseResult{}, errors.New("文件解析失败，请上传有效的 xlsx 文件")
	}
	defer file.Close()

	sheetName := file.GetSheetName(0)
	rows, err := file.GetRows(sheetName)
	if err != nil {
		return model.PEP3IEPMaterialImportParseResult{}, err
	}
	if len(rows) == 0 {
		return model.PEP3IEPMaterialImportParseResult{}, errors.New("导入文件为空")
	}

	templateColumns := buildPEP3IEPImportTemplateColumns(lookups)
	templateColumnMap := make(map[string]model.PEP3IEPMaterialImportTemplateColumn, len(templateColumns))
	for _, column := range templateColumns {
		templateColumnMap[column.Title] = column
	}

	headerRow := rows[0]
	columns := make([]model.PEP3IEPMaterialImportColumn, 0, len(templateColumns))
	headerIndexes := make([]int, 0, len(templateColumns))
	seenTitle := map[string]bool{}
	for idx, item := range headerRow {
		title := strings.TrimSpace(strings.TrimPrefix(item, "*"))
		column, ok := templateColumnMap[title]
		if !ok || seenTitle[title] {
			continue
		}
		seenTitle[title] = true
		columns = append(columns, model.PEP3IEPMaterialImportColumn{
			Key:       buildPEP3IEPImportColumnKey(title, len(columns)),
			Title:     title,
			Required:  column.Required,
			FieldType: column.FieldType,
			Options:   append([]string{}, column.Options...),
		})
		headerIndexes = append(headerIndexes, idx)
	}
	if len(columns) == 0 {
		return model.PEP3IEPMaterialImportParseResult{}, errors.New("未识别到可导入字段，请使用最新模板")
	}
	for _, column := range templateColumns {
		if column.Required && !seenTitle[column.Title] {
			return model.PEP3IEPMaterialImportParseResult{}, fmt.Errorf("缺少必填字段「%s」，请使用最新模板", column.Title)
		}
	}

	result := model.PEP3IEPMaterialImportParseResult{
		ImportID: time.Now().Format("20060102150405") + fmt.Sprintf("%09d", time.Now().UnixNano()%1e9),
		FileName: strings.TrimSpace(filename),
		InstName: "平台总控",
		Columns:  columns,
		Rows:     make([]model.PEP3IEPMaterialImportRow, 0, len(rows)),
	}

	for rowIdx := 1; rowIdx < len(rows) && len(result.Rows) < pep3IEPMaterialImportRowLimit; rowIdx++ {
		rawRow := rows[rowIdx]
		cells := make([]model.PEP3IEPMaterialImportCell, 0, len(columns))
		hasRawValue := false
		for colIdx, column := range columns {
			value := ""
			sourceIndex := headerIndexes[colIdx]
			if sourceIndex < len(rawRow) {
				value = strings.TrimSpace(rawRow[sourceIndex])
			}
			if value != "" {
				hasRawValue = true
			}
			cells = append(cells, model.PEP3IEPMaterialImportCell{
				Key:   column.Key,
				Title: column.Title,
				Value: value,
			})
		}
		if !hasRawValue {
			continue
		}
		row := model.PEP3IEPMaterialImportRow{
			ID:    fmt.Sprintf("%s_%d", result.ImportID, rowIdx+1),
			RowNo: rowIdx + 1,
			Cells: cells,
		}
		row = normalizePEP3IEPImportRow(row, columns, lookups)
		result.Rows = append(result.Rows, row)
		if row.HasError {
			result.AbnormalCount++
		} else {
			result.NormalCount++
		}
	}

	if len(result.Rows) == 0 {
		return model.PEP3IEPMaterialImportParseResult{}, errors.New("请勿上传空文件")
	}
	if len(result.Rows) >= pep3IEPMaterialImportRowLimit && len(rows)-1 > pep3IEPMaterialImportRowLimit {
		return model.PEP3IEPMaterialImportParseResult{}, errors.New("每次最多支持导入1000条数据")
	}
	return result, nil
}

func (svc *Service) loadPEP3IEPImportLookups() (pep3IEPImportLookups, error) {
	if svc.repo == nil {
		return pep3IEPImportLookups{}, errors.New("repository is not configured")
	}
	bank, err := svc.repo.GetScaleQuestionBank(context.Background(), "PEP3", "")
	if err != nil {
		return pep3IEPImportLookups{}, err
	}

	lookups := pep3IEPImportLookups{
		Domains:           make([]pep3IEPImportDomainOption, 0, len(bank.Domains)),
		DomainByName:      map[string]pep3IEPImportDomainOption{},
		Questions:         make([]model.ScaleQuestionBankItem, 0, len(bank.Items)),
		QuestionByName:    map[string]model.ScaleQuestionBankItem{},
		QuestionsByDomain: map[string][]model.ScaleQuestionBankItem{},
	}
	for _, item := range bank.Domains {
		option := pep3IEPImportDomainOption{
			Code: strings.TrimSpace(item.ScaleCode),
			Name: strings.TrimSpace(item.ScaleName),
		}
		if option.Name == "" {
			option.Name = option.Code
		}
		if option.Code == "" || option.Name == "" {
			continue
		}
		lookups.Domains = append(lookups.Domains, option)
		lookups.DomainByName[option.Name] = option
	}
	for _, item := range bank.Items {
		title := pep3IEPImportQuestionTitle(item)
		if item.ItemNo <= 0 || title == "" {
			continue
		}
		item.ItemTitle = title
		lookups.Questions = append(lookups.Questions, item)
		lookups.QuestionByName[title] = item
		lookups.QuestionsByDomain[strings.TrimSpace(item.DomainCode)] = append(lookups.QuestionsByDomain[strings.TrimSpace(item.DomainCode)], item)
	}
	sort.Slice(lookups.Domains, func(i, j int) bool { return lookups.Domains[i].Code < lookups.Domains[j].Code })
	sort.Slice(lookups.Questions, func(i, j int) bool { return lookups.Questions[i].ItemNo < lookups.Questions[j].ItemNo })
	for domainCode := range lookups.QuestionsByDomain {
		items := lookups.QuestionsByDomain[domainCode]
		sort.Slice(items, func(i, j int) bool { return items[i].ItemNo < items[j].ItemNo })
		lookups.QuestionsByDomain[domainCode] = items
	}
	return lookups, nil
}

func buildPEP3IEPImportTemplateColumns(lookups pep3IEPImportLookups) []model.PEP3IEPMaterialImportTemplateColumn {
	return []model.PEP3IEPMaterialImportTemplateColumn{
		{Title: "领域", Required: true, FieldType: 4, Options: pep3IEPImportDomainNames(lookups)},
		{Title: "题目", Required: true, FieldType: 4, Options: pep3IEPImportQuestionTitles(lookups)},
		{Title: "选项", Required: true, FieldType: 4, Options: []string{"0分", "1分", "2分"}},
		{Title: "长期目标", Required: true, FieldType: 1},
		{Title: "短期目标", FieldType: 1},
		{Title: "课程形式", FieldType: 4, Options: []string{"个训", "集体课"}},
		{Title: "训练项目", FieldType: 1},
		{Title: "训练内容", FieldType: 1},
		{Title: "状态", FieldType: 4, Options: []string{"启用", "停用"}},
	}
}

func buildPEP3IEPImportTemplateWorkbook(columns []model.PEP3IEPMaterialImportTemplateColumn, lookups pep3IEPImportLookups) ([]byte, error) {
	file := excelize.NewFile()
	sheetName := file.GetSheetName(0)
	if err := file.SetSheetName(sheetName, "PEP3素材导入"); err == nil {
		sheetName = "PEP3素材导入"
	}
	const rowCount = pep3IEPMaterialImportRowLimit

	headerStyle, err := file.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 10, Family: "Microsoft YaHei", Color: "#222222"},
		Fill: excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#D8E5F7"}},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	if err != nil {
		return nil, err
	}
	dataCellStyle, err := file.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 11, Family: "Microsoft YaHei", Color: "#222222"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
	})
	if err != nil {
		return nil, err
	}
	noteStyle, err := file.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 12, Family: "Microsoft YaHei", Color: "#111111"},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "top", WrapText: true},
	})
	if err != nil {
		return nil, err
	}

	helperColIndex := 1
	domainCol := ""
	questionCol := ""
	for idx, column := range columns {
		cell, _ := excelize.CoordinatesToCellName(idx+1, 1)
		col := platformExcelColumnName(idx + 1)
		switch column.Title {
		case "领域":
			domainCol = col
		case "题目":
			questionCol = col
		}
		file.SetColWidth(sheetName, col, col, pep3IEPImportColumnWidth(column.Title))
		if err := file.SetCellRichText(sheetName, cell, buildPlatformImportHeaderRichText(column.Title, column.Required)); err != nil {
			return nil, err
		}
		if err := file.SetCellStyle(sheetName, cell, cell, headerStyle); err != nil {
			return nil, err
		}
		dataStartCell := fmt.Sprintf("%s%d", col, 2)
		dataEndCell := fmt.Sprintf("%s%d", col, rowCount+1)
		if err := file.SetCellStyle(sheetName, dataStartCell, dataEndCell, dataCellStyle); err != nil {
			return nil, err
		}
		if len(column.Options) > 0 && column.Title != "题目" {
			helperCol := platformExcelColumnName(helperColIndex)
			helperColIndex++
			if err := addPlatformTemplateDropdownValidationBySheetRange(file, sheetName, "template_options", helperCol, col, 2, rowCount+1, column.Options, !column.Required); err != nil {
				return nil, err
			}
		}
	}
	if err := addPEP3IEPQuestionDependentDropdown(file, sheetName, "domain_question_options", domainCol, questionCol, 2, rowCount+1, lookups); err != nil {
		return nil, err
	}

	noteCol := len(columns) + 1
	noteHeaderCell, _ := excelize.CoordinatesToCellName(noteCol, 1)
	file.SetColWidth(sheetName, platformExcelColumnName(noteCol), platformExcelColumnName(noteCol), 64)
	if err := file.SetCellValue(sheetName, noteHeaderCell, "填写说明"); err != nil {
		return nil, err
	}
	if err := file.SetCellStyle(sheetName, noteHeaderCell, noteHeaderCell, headerStyle); err != nil {
		return nil, err
	}
	noteStartCell, _ := excelize.CoordinatesToCellName(noteCol, 2)
	noteEndCell, _ := excelize.CoordinatesToCellName(noteCol, rowCount+1)
	if err := file.MergeCell(sheetName, noteStartCell, noteEndCell); err != nil {
		return nil, err
	}
	if err := file.SetCellRichText(sheetName, noteStartCell, buildPEP3IEPImportNotesRichText(lookups)); err != nil {
		return nil, err
	}
	if err := file.SetCellStyle(sheetName, noteStartCell, noteStartCell, noteStyle); err != nil {
		return nil, err
	}
	for row := 1; row <= rowCount+1; row++ {
		height := 24.0
		if row == 1 {
			height = 22
		}
		if err := file.SetRowHeight(sheetName, row, height); err != nil {
			return nil, err
		}
	}

	buffer, err := file.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func buildPEP3IEPImportNotesRichText(lookups pep3IEPImportLookups) []excelize.RichTextRun {
	black := &excelize.Font{Size: 12, Family: "Microsoft YaHei", Color: "#111111"}
	red := &excelize.Font{Size: 12, Family: "Microsoft YaHei", Color: "#FF3B30"}
	title := &excelize.Font{Size: 12, Family: "Microsoft YaHei", Bold: true, Color: "#111111"}
	return []excelize.RichTextRun{
		{Text: "【导入提示】\n", Font: title},
		{Text: "1、请勿修改顶部字段标题及顺序，带 * 的「领域」「题目」「选项」「长期目标」为必填。\n", Font: black},
		{Text: "2、请先选择「领域」，「题目」下拉会自动筛选该领域下的题目；系统也会校验题目是否属于所选领域。\n", Font: black},
		{Text: "3、每行会导入一条题目选项对应的长期目标；同一长期目标下可继续填写短期目标、课程形式、训练项目和训练内容。\n", Font: black},
		{Text: "4、填写短期目标时必须选择课程形式；填写训练项目或训练内容时必须同时填写短期目标、训练项目和训练内容。\n", Font: black},
		{Text: "5、「状态」不填默认启用，最多导入1000条数据。\n\n", Font: black},
		{Text: "【选项说明】\n", Font: title},
		{Text: "选项：0分、1分、2分；课程形式：个训、集体课；状态：启用、停用。\n", Font: red},
		{Text: fmt.Sprintf("当前模板包含 %d 个领域、%d 道题目。", len(lookups.Domains), len(lookups.Questions)), Font: black},
	}
}

func normalizePEP3IEPImportRow(row model.PEP3IEPMaterialImportRow, columns []model.PEP3IEPMaterialImportColumn, lookups pep3IEPImportLookups) model.PEP3IEPMaterialImportRow {
	columnMap := make(map[string]model.PEP3IEPMaterialImportColumn, len(columns))
	for _, column := range columns {
		columnMap[column.Key] = column
	}
	rowData := pep3IEPImportRowValueMap(row)
	hasError := false
	for idx := range row.Cells {
		cell := &row.Cells[idx]
		column := columnMap[cell.Key]
		cell.Value = strings.TrimSpace(cell.Value)
		cell.Error = validatePEP3IEPImportCell(column, *cell, lookups)
		if cell.Error != "" {
			hasError = true
		}
	}
	if errText := validatePEP3IEPImportRowRelation(rowData, lookups); errText != "" {
		hasError = true
		applyPEP3IEPImportRowError(row.Cells, errText)
	}
	row.HasError = hasError
	return row
}

func validatePEP3IEPImportCell(column model.PEP3IEPMaterialImportColumn, cell model.PEP3IEPMaterialImportCell, lookups pep3IEPImportLookups) string {
	text := strings.TrimSpace(cell.Value)
	if column.Required && text == "" {
		return "请填写"
	}
	if text == "" {
		return ""
	}
	switch column.Title {
	case "领域":
		if _, ok := lookups.DomainByName[text]; !ok {
			return "请选择预设领域"
		}
	case "题目":
		if _, ok := lookups.QuestionByName[text]; !ok {
			return "请选择预设题目"
		}
	case "选项":
		if _, ok := parsePEP3IEPScoreValue(text); !ok {
			return "请选择0分、1分或2分"
		}
	case "课程形式":
		if text != "个训" && text != "集体课" {
			return "请选择课程形式"
		}
	case "状态":
		if text != "启用" && text != "停用" {
			return "请选择启用或停用"
		}
	}
	return ""
}

func validatePEP3IEPImportRowRelation(rowData map[string]string, lookups pep3IEPImportLookups) string {
	domainName := strings.TrimSpace(rowData["领域"])
	questionName := strings.TrimSpace(rowData["题目"])
	shortGoal := strings.TrimSpace(rowData["短期目标"])
	courseForm := strings.TrimSpace(rowData["课程形式"])
	trainingProject := strings.TrimSpace(rowData["训练项目"])
	trainingContent := strings.TrimSpace(rowData["训练内容"])

	if domainName != "" && questionName != "" {
		domain, domainOK := lookups.DomainByName[domainName]
		question, questionOK := lookups.QuestionByName[questionName]
		if domainOK && questionOK && strings.TrimSpace(question.DomainCode) != strings.TrimSpace(domain.Code) {
			return "题目不属于所选领域"
		}
	}
	if shortGoal != "" && courseForm == "" {
		return "填写短期目标时必须选择课程形式"
	}
	if courseForm != "" && shortGoal == "" {
		return "选择课程形式时必须填写短期目标"
	}
	if trainingProject != "" || trainingContent != "" {
		if shortGoal == "" {
			return "填写训练内容时必须先填写短期目标"
		}
		if trainingProject == "" || trainingContent == "" {
			return "训练项目和训练内容必须同时填写"
		}
	}
	return ""
}

func applyPEP3IEPImportRowError(cells []model.PEP3IEPMaterialImportCell, errText string) {
	for idx := range cells {
		if cells[idx].Error == "" {
			cells[idx].Error = errText
			return
		}
	}
}

func pep3IEPImportRowValueMap(row model.PEP3IEPMaterialImportRow) map[string]string {
	result := make(map[string]string, len(row.Cells))
	for _, cell := range row.Cells {
		result[cell.Title] = strings.TrimSpace(cell.Value)
	}
	return result
}

func buildPEP3IEPImportColumnKey(title string, index int) string {
	return fmt.Sprintf("pep3_iep_%d_%s", index+1, strings.ToLower(strconv.FormatInt(int64(len(title)), 36)))
}

func pep3IEPImportDomainNames(lookups pep3IEPImportLookups) []string {
	result := make([]string, 0, len(lookups.Domains))
	for _, item := range lookups.Domains {
		result = append(result, item.Name)
	}
	return result
}

func pep3IEPImportQuestionTitles(lookups pep3IEPImportLookups) []string {
	result := make([]string, 0, len(lookups.Questions))
	for _, item := range lookups.Questions {
		result = append(result, pep3IEPImportQuestionTitle(item))
	}
	return result
}

func pep3IEPImportQuestionTitle(item model.ScaleQuestionBankItem) string {
	title := strings.TrimSpace(item.ItemTitle)
	if title == "" {
		title = strings.TrimSpace(item.TestItem)
	}
	return title
}

func parsePEP3IEPScoreValue(text string) (int, bool) {
	value := strings.TrimSpace(strings.TrimSuffix(text, "分"))
	score, err := strconv.Atoi(value)
	if err != nil || (score != 0 && score != 1 && score != 2) {
		return 0, false
	}
	return score, true
}

func pep3IEPImportStatusValue(text string) string {
	switch strings.TrimSpace(text) {
	case "停用", "inactive":
		return "inactive"
	default:
		return "active"
	}
}

func pep3IEPImportColumnWidth(title string) float64 {
	switch strings.TrimSpace(title) {
	case "领域", "选项", "课程形式", "状态":
		return 16
	case "题目":
		return 40
	case "长期目标", "短期目标", "训练内容":
		return 36
	case "训练项目":
		return 24
	default:
		return 18
	}
}

func buildPlatformImportHeaderRichText(titleText string, required bool) []excelize.RichTextRun {
	if !required {
		return []excelize.RichTextRun{{Text: titleText, Font: &excelize.Font{Bold: true, Size: 10, Family: "Microsoft YaHei", Color: "#222222"}}}
	}
	return []excelize.RichTextRun{
		{Text: "*", Font: &excelize.Font{Bold: true, Size: 10, Family: "Microsoft YaHei", Color: "#FF4D4F"}},
		{Text: titleText, Font: &excelize.Font{Bold: true, Size: 10, Family: "Microsoft YaHei", Color: "#222222"}},
	}
}

func platformExcelColumnName(index int) string {
	name, _ := excelize.ColumnNumberToName(index)
	return name
}

func addPlatformTemplateDropdownValidationBySheetRange(file *excelize.File, sheetName, helperSheetName, helperCol, targetCol string, startRow, endRow int, options []string, allowBlank bool) error {
	if len(options) == 0 {
		return nil
	}
	helperSheetIndex, err := file.GetSheetIndex(helperSheetName)
	if err != nil {
		return err
	}
	if helperSheetIndex == -1 {
		if _, err := file.NewSheet(helperSheetName); err != nil {
			return err
		}
	}
	for idx, option := range options {
		cell := fmt.Sprintf("%s%d", helperCol, idx+1)
		if err := file.SetCellValue(helperSheetName, cell, option); err != nil {
			return err
		}
	}
	if err := file.SetSheetVisible(helperSheetName, false); err != nil {
		return err
	}
	dv := excelize.NewDataValidation(allowBlank)
	dv.Sqref = fmt.Sprintf("%s%d:%s%d", targetCol, startRow, targetCol, endRow)
	dv.SetSqrefDropList(fmt.Sprintf("%s$%s$1:$%s$%d", excelFormulaSheetRef(helperSheetName), helperCol, helperCol, len(options)))
	dv.ShowErrorMessage = true
	dv.SetError(excelize.DataValidationErrorStyleStop, "填写有误", "请选择下拉选项中的预设值")
	dv.ShowInputMessage = true
	if allowBlank {
		dv.SetInput("可选项", "请从下拉列表中选择")
	} else {
		dv.SetInput("必填项", "请从下拉列表中选择，不能为空")
	}
	return file.AddDataValidation(sheetName, dv)
}

func addPEP3IEPQuestionDependentDropdown(file *excelize.File, sheetName, helperSheetName, domainCol, questionCol string, startRow, endRow int, lookups pep3IEPImportLookups) error {
	if domainCol == "" || questionCol == "" || len(lookups.Domains) == 0 {
		return nil
	}
	helperSheetIndex, err := file.GetSheetIndex(helperSheetName)
	if err != nil {
		return err
	}
	if helperSheetIndex == -1 {
		if _, err := file.NewSheet(helperSheetName); err != nil {
			return err
		}
	}

	for idx, domain := range lookups.Domains {
		col := platformExcelColumnName(idx + 1)
		if err := file.SetCellValue(helperSheetName, fmt.Sprintf("%s1", col), domain.Name); err != nil {
			return err
		}
		for questionIdx, question := range lookups.QuestionsByDomain[domain.Code] {
			cell := fmt.Sprintf("%s%d", col, questionIdx+2)
			if err := file.SetCellValue(helperSheetName, cell, pep3IEPImportQuestionTitle(question)); err != nil {
				return err
			}
		}
	}
	if err := file.SetSheetVisible(helperSheetName, false); err != nil {
		return err
	}

	lastCol := platformExcelColumnName(len(lookups.Domains))
	helperSheetRef := excelFormulaSheetRef(helperSheetName)
	helperHeaderRange := fmt.Sprintf("%s$A$1:$%s$1", helperSheetRef, lastCol)
	helperColumnRange := fmt.Sprintf("%s$A:$%s", helperSheetRef, lastCol)
	for row := startRow; row <= endRow; row++ {
		dv := excelize.NewDataValidation(false)
		dv.Sqref = fmt.Sprintf("%s%d", questionCol, row)
		// The formula picks the helper-sheet column whose header equals the current row's domain.
		// It keeps the question dropdown limited to the selected PEP3 domain.
		formula := fmt.Sprintf(
			"OFFSET(%s$A$2,0,MATCH($%s%d,%s,0)-1,COUNTA(INDEX(%s,0,MATCH($%s%d,%s,0)))-1,1)",
			helperSheetRef,
			domainCol,
			row,
			helperHeaderRange,
			helperColumnRange,
			domainCol,
			row,
			helperHeaderRange,
		)
		dv.SetSqrefDropList(formula)
		dv.ShowErrorMessage = true
		dv.SetError(excelize.DataValidationErrorStyleStop, "题目不属于领域", "请先选择领域，再从题目下拉中选择该领域对应的题目")
		dv.ShowInputMessage = true
		dv.SetInput("题目联动", "题目下拉会根据当前行的领域自动筛选")
		if err := file.AddDataValidation(sheetName, dv); err != nil {
			return err
		}
	}
	return nil
}

func excelFormulaSheetRef(sheetName string) string {
	return "'" + strings.ReplaceAll(sheetName, "'", "''") + "'!"
}

func sanitizePlatformImportFileName(name string) string {
	replacer := strings.NewReplacer("/", "-", "\\", "-", "?", "", "*", "", ":", "-", "\"", "", "<", "", ">", "", "|", "")
	return replacer.Replace(strings.TrimSpace(name))
}
