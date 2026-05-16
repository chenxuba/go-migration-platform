package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/signintech/gopdf"
	"go-migration-platform/pkg/logx"
	"go-migration-platform/services/education/internal/model"
)

const (
	AutismDevReportSectionAssessmentInfo     = "assessmentInfo"
	AutismDevReportSectionResultAnalysis     = "resultAnalysis"
	AutismDevReportSectionTraining           = "training"
	AutismDevReportSectionDevelopmentProfile = "developmentProfile"
	AutismDevReportSectionBehaviorProfile    = "behaviorProfile"
)

func (svc *Service) ExportAutismDevSelectedReportPDF(userID int64, recordID int64, sections []string, analysis *model.AutismDevResultAnalysisVO) (string, string, []byte, error) {
	startedAt := time.Now()
	normalizedSections := normalizeAutismDevSelectedReportSections(sections)
	if len(normalizedSections) == 0 {
		return "", "", nil, errors.New("请选择打印内容")
	}

	record, err := svc.GetAutismDevAssessmentRecord(userID, recordID)
	if err != nil {
		return "", "", nil, err
	}

	if len(normalizedSections) == 1 {
		sectionStartedAt := time.Now()
		content, err := svc.autismDevSelectedReportSectionOrWordPDF(userID, recordID, normalizedSections[0], analysis)
		if err != nil {
			return "", "", nil, err
		}
		if len(content) == 0 {
			return "", "", nil, fmt.Errorf("%s暂无可打印内容", autismDevSelectedReportSectionLabel(normalizedSections[0]))
		}
		logx.Info("autismdev selected report single section pdf finished", logx.Entry{
			"record_id":   recordID,
			"section":     normalizedSections[0],
			"pdf_bytes":   len(content),
			"duration_ms": time.Since(sectionStartedAt).Milliseconds(),
		})
		fileName := fmt.Sprintf("%s-孤独症儿童发展评估报告-%s.pdf", sanitizeExportFileName(record.StudentName), time.Now().Format("20060102150405"))
		logx.Info("autismdev selected report pdf finished", logx.Entry{
			"record_id":   recordID,
			"sections":    strings.Join(normalizedSections, ","),
			"pdf_bytes":   len(content),
			"duration_ms": time.Since(startedAt).Milliseconds(),
		})
		return fileName, iepPlanPDFContentType, content, nil
	}

	institutionName := ""
	if svc.repo != nil {
		institutionName, err = svc.repo.GetInstitutionName(context.Background(), record.InstID)
		if err != nil {
			return "", "", nil, err
		}
	}

	builder := newAutismDevSelectedReportPDFBuilder()
	for index := 0; index < len(normalizedSections); {
		section := normalizedSections[index]
		if isAutismDevSelectedReportWordSection(section) {
			end := index + 1
			for end < len(normalizedSections) && isAutismDevSelectedReportWordSection(normalizedSections[end]) {
				end++
			}
			sectionStartedAt := time.Now()
			content, err := svc.autismDevSelectedReportWordSectionsPDF(userID, recordID, normalizedSections[index:end], analysis)
			if err != nil {
				return "", "", nil, err
			}
			if len(content) == 0 {
				return "", "", nil, errors.New("Word报告内容暂无可打印内容")
			}
			logx.Info("autismdev selected report word sections pdf finished", logx.Entry{
				"record_id":    recordID,
				"sections":     strings.Join(normalizedSections[index:end], ","),
				"pdf_bytes":    len(content),
				"duration_ms":  time.Since(sectionStartedAt).Milliseconds(),
				"section_from": index,
				"section_to":   end,
			})
			if err := builder.appendPDFBytes(content); err != nil {
				return "", "", nil, err
			}
			index = end
			continue
		}
		sectionStartedAt := time.Now()
		if isAutismDevSelectedReportDirectDrawSection(section) {
			if err := svc.appendAutismDevSelectedReportDirectSectionPDF(builder, record, section, institutionName); err != nil {
				return "", "", nil, err
			}
			logx.Info("autismdev selected report direct section pdf finished", logx.Entry{
				"record_id":    recordID,
				"section":      section,
				"duration_ms":  time.Since(sectionStartedAt).Milliseconds(),
				"section_from": index,
			})
		} else {
			content, err := svc.autismDevSelectedReportSectionPDF(userID, recordID, section, analysis)
			if err != nil {
				return "", "", nil, err
			}
			if len(content) == 0 {
				return "", "", nil, fmt.Errorf("%s暂无可打印内容", autismDevSelectedReportSectionLabel(section))
			}
			logx.Info("autismdev selected report section pdf finished", logx.Entry{
				"record_id":    recordID,
				"section":      section,
				"pdf_bytes":    len(content),
				"duration_ms":  time.Since(sectionStartedAt).Milliseconds(),
				"section_from": index,
			})
			if err := builder.appendPDFBytes(content); err != nil {
				return "", "", nil, err
			}
		}
		index++
	}
	if builder.partCount == 0 {
		return "", "", nil, errors.New("暂无可打印内容")
	}

	buildStartedAt := time.Now()
	content, err := builder.bytes()
	if err != nil {
		return "", "", nil, err
	}
	logx.Info("autismdev selected report pdf build finished", logx.Entry{
		"record_id":   recordID,
		"part_count":  builder.partCount,
		"pdf_bytes":   len(content),
		"duration_ms": time.Since(buildStartedAt).Milliseconds(),
	})
	fileName := fmt.Sprintf("%s-孤独症儿童发展评估报告-%s.pdf", sanitizeExportFileName(record.StudentName), time.Now().Format("20060102150405"))
	logx.Info("autismdev selected report pdf finished", logx.Entry{
		"record_id":   recordID,
		"sections":    strings.Join(normalizedSections, ","),
		"pdf_bytes":   len(content),
		"duration_ms": time.Since(startedAt).Milliseconds(),
	})
	return fileName, iepPlanPDFContentType, content, nil
}

func (svc *Service) autismDevSelectedReportSectionOrWordPDF(userID int64, recordID int64, section string, analysis *model.AutismDevResultAnalysisVO) ([]byte, error) {
	if isAutismDevSelectedReportWordSection(section) {
		return svc.autismDevSelectedReportWordSectionsPDF(userID, recordID, []string{section}, analysis)
	}
	return svc.autismDevSelectedReportSectionPDF(userID, recordID, section, analysis)
}

func (svc *Service) autismDevSelectedReportWordSectionsPDF(userID int64, recordID int64, sections []string, analysis *model.AutismDevResultAnalysisVO) ([]byte, error) {
	if len(sections) == 0 {
		return nil, nil
	}
	_, record, score, data, itemScores, err := svc.autismDevResultAnalysisContext(userID, recordID)
	if err != nil {
		return nil, err
	}
	exports := make([]autismDevSelectedReportWordSectionExport, 0, len(sections))
	for _, section := range sections {
		switch section {
		case AutismDevReportSectionAssessmentInfo:
			export := buildAutismDevAssessmentSituationWordExport(record, score, data, itemScores)
			if len(export.DevelopmentRows) == 0 && len(export.BehaviorRows) == 0 {
				return nil, errors.New("评估情况暂无可打印内容")
			}
			exports = append(exports, autismDevSelectedReportWordSectionExport{
				Section:             section,
				AssessmentSituation: export,
			})
		case AutismDevReportSectionResultAnalysis:
			export, err := svc.autismDevResultAnalysisWordExport(userID, recordID, analysis)
			if err != nil {
				return nil, err
			}
			exports = append(exports, autismDevSelectedReportWordSectionExport{
				Section:        section,
				ResultAnalysis: export,
			})
		}
	}
	docx, err := buildAutismDevSelectedReportWordDocx(exports)
	if err != nil {
		return nil, err
	}
	fileName := fmt.Sprintf("%s-孤独症儿童发展评估报告-%s.docx", sanitizeExportFileName(record.StudentName), time.Now().Format("20060102150405"))
	_, _, pdf, err := exportIEPPDFByDOCX(fileName, docx)
	return pdf, err
}

func (svc *Service) autismDevSelectedReportSectionPDF(userID int64, recordID int64, section string, analysis *model.AutismDevResultAnalysisVO) ([]byte, error) {
	switch section {
	case AutismDevReportSectionAssessmentInfo:
		_, _, content, err := svc.ExportAutismDevAssessmentSituationPDF(userID, recordID)
		return content, err
	case AutismDevReportSectionResultAnalysis:
		_, _, content, err := svc.ExportAutismDevResultAnalysisPDF(userID, recordID, analysis)
		return content, err
	case AutismDevReportSectionTraining:
		_, content, err := svc.GenerateAutismDevProfilePDF(userID, recordID, "training")
		return content, err
	case AutismDevReportSectionDevelopmentProfile:
		_, content, err := svc.GenerateAutismDevProfilePDF(userID, recordID, "development")
		return content, err
	case AutismDevReportSectionBehaviorProfile:
		_, content, err := svc.GenerateAutismDevProfilePDF(userID, recordID, "behavior")
		return content, err
	default:
		return nil, fmt.Errorf("不支持的打印内容：%s", section)
	}
}

func (svc *Service) appendAutismDevSelectedReportDirectSectionPDF(builder *autismDevSelectedReportPDFBuilder, record model.AssessmentRecordDetailVO, section string, institutionName string) error {
	switch section {
	case AutismDevReportSectionTraining:
		records := []model.AssessmentRecordDetailVO{record}
		if svc.repo != nil {
			history, err := svc.repo.ListAssessmentRecordsForStudentScale(
				context.Background(),
				record.InstID,
				record.StudentID,
				record.AssessmentCode,
				record.ID,
				3,
			)
			if err != nil {
				return err
			}
			if len(history) > 0 {
				records = history
			}
		}
		return builder.appendDirectDraw(func(pdf *gopdf.GoPdf) error {
			return drawAutismDevTrainingEffectPDFPages(pdf, record, records, institutionName)
		})
	case AutismDevReportSectionDevelopmentProfile:
		return builder.appendDirectDraw(func(pdf *gopdf.GoPdf) error {
			return drawAutismDevProfilePDFPage(pdf, record, autismDevDevelopmentProfilePDF, institutionName)
		})
	case AutismDevReportSectionBehaviorProfile:
		return builder.appendDirectDraw(func(pdf *gopdf.GoPdf) error {
			return drawAutismDevProfilePDFPage(pdf, record, autismDevBehaviorProfilePDF, institutionName)
		})
	default:
		return fmt.Errorf("不支持直接绘制的打印内容：%s", section)
	}
}

func isAutismDevSelectedReportWordSection(section string) bool {
	return section == AutismDevReportSectionAssessmentInfo ||
		section == AutismDevReportSectionResultAnalysis
}

func isAutismDevSelectedReportDirectDrawSection(section string) bool {
	return section == AutismDevReportSectionTraining ||
		section == AutismDevReportSectionDevelopmentProfile ||
		section == AutismDevReportSectionBehaviorProfile
}

func normalizeAutismDevSelectedReportSections(sections []string) []string {
	allowed := map[string]bool{
		AutismDevReportSectionAssessmentInfo:     true,
		AutismDevReportSectionResultAnalysis:     true,
		AutismDevReportSectionTraining:           true,
		AutismDevReportSectionDevelopmentProfile: true,
		AutismDevReportSectionBehaviorProfile:    true,
	}
	out := make([]string, 0, len(sections))
	seen := make(map[string]bool, len(sections))
	for _, raw := range sections {
		section := strings.TrimSpace(raw)
		if !allowed[section] || seen[section] {
			continue
		}
		seen[section] = true
		out = append(out, section)
	}
	return out
}

func autismDevSelectedReportSectionLabel(section string) string {
	switch section {
	case AutismDevReportSectionAssessmentInfo:
		return "评估情况"
	case AutismDevReportSectionResultAnalysis:
		return "评估结果分析"
	case AutismDevReportSectionTraining:
		return "训练效果"
	case AutismDevReportSectionDevelopmentProfile:
		return "发展情况剖面图"
	case AutismDevReportSectionBehaviorProfile:
		return "情绪行为表现图"
	default:
		return "所选内容"
	}
}

type autismDevSelectedReportWordSectionExport struct {
	Section             string
	AssessmentSituation autismDevAssessmentSituationWordExport
	ResultAnalysis      autismDevResultAnalysisWordExport
}

func buildAutismDevSelectedReportWordDocx(exports []autismDevSelectedReportWordSectionExport) ([]byte, error) {
	if len(exports) == 0 {
		return nil, errors.New("暂无可导出的报告内容")
	}
	entries := map[string][]byte{
		"[Content_Types].xml":          []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/></Types>`),
		"_rels/.rels":                  []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/></Relationships>`),
		"word/_rels/document.xml.rels": []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"></Relationships>`),
		"word/document.xml":            []byte(buildAutismDevSelectedReportWordDocumentXML(exports)),
	}
	return writeDocxZipEntries(entries)
}

func buildAutismDevSelectedReportWordDocumentXML(exports []autismDevSelectedReportWordSectionExport) string {
	var builder strings.Builder
	builder.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	builder.WriteString(`<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">`)
	builder.WriteString(`<w:body>`)
	for index, export := range exports {
		if index > 0 {
			builder.WriteString(buildIEPPageBreakParagraph())
		}
		switch export.Section {
		case AutismDevReportSectionAssessmentInfo:
			builder.WriteString(buildAutismDevAssessmentSituationTitleParagraph(firstNonEmptyExportValue(export.AssessmentSituation.Title, autismDevAssessmentSituationTitle)))
			builder.WriteString(buildAutismDevAssessmentSituationTable(export.AssessmentSituation))
		case AutismDevReportSectionResultAnalysis:
			builder.WriteString(buildAutismDevResultAnalysisTitleParagraph(firstNonEmptyExportValue(export.ResultAnalysis.Title, "孤独症儿童评估结果分析表")))
			builder.WriteString(buildAutismDevResultAnalysisMetaParagraph(export.ResultAnalysis))
			builder.WriteString(buildAutismDevResultAnalysisTable(export.ResultAnalysis.Rows))
		}
	}
	builder.WriteString(`<w:sectPr><w:pgSz w:w="11906" w:h="16838"/><w:pgMar w:top="720" w:right="720" w:bottom="720" w:left="720" w:header="851" w:footer="720" w:gutter="0"/></w:sectPr>`)
	builder.WriteString(`</w:body></w:document>`)
	return builder.String()
}

func mergeAutismDevReportPDFBytes(pdfs [][]byte) ([]byte, error) {
	var pdf gopdf.GoPdf
	pdf.Start(gopdf.Config{
		Unit:     gopdf.UnitPT,
		PageSize: *gopdf.PageSizeA4,
	})
	for _, content := range pdfs {
		if err := importAutismDevReportPDFPages(&pdf, content); err != nil {
			return nil, err
		}
	}
	var output bytes.Buffer
	if err := pdf.Write(&output); err != nil {
		return nil, fmt.Errorf("合并PDF失败：%w", err)
	}
	return output.Bytes(), nil
}

func importAutismDevReportPDFPages(pdf *gopdf.GoPdf, content []byte) error {
	if len(content) == 0 {
		return nil
	}
	sizeReader := io.ReadSeeker(bytes.NewReader(content))
	pageSizes := pdf.GetStreamPageSizes(&sizeReader)
	if len(pageSizes) == 0 {
		return errors.New("合并PDF失败：无法读取页面尺寸")
	}
	for pageNo := 1; pageNo <= len(pageSizes); pageNo++ {
		box := pageSizes[pageNo]["/MediaBox"]
		if box == nil {
			box = pageSizes[pageNo]["/CropBox"]
		}
		if box == nil {
			return errors.New("合并PDF失败：无法读取页面边界")
		}
		width := box["w"]
		height := box["h"]
		if width <= 0 || height <= 0 {
			return errors.New("合并PDF失败：页面尺寸无效")
		}
		pageSize := &gopdf.Rect{W: width, H: height}
		pdf.AddPageWithOption(gopdf.PageOption{PageSize: pageSize})
		pageReader := io.ReadSeeker(bytes.NewReader(content))
		templateID := pdf.ImportPageStream(&pageReader, pageNo, "/MediaBox")
		pdf.UseImportedTemplate(templateID, 0, 0, width, height)
	}
	return nil
}

type autismDevSelectedReportPDFBuilder struct {
	pdf        gopdf.GoPdf
	fontReady  bool
	partCount  int
	directDraw int
}

func newAutismDevSelectedReportPDFBuilder() *autismDevSelectedReportPDFBuilder {
	builder := &autismDevSelectedReportPDFBuilder{}
	builder.pdf.Start(gopdf.Config{
		Unit:     gopdf.UnitPT,
		PageSize: gopdf.Rect{W: autismDevProfilePDFPageWidth, H: autismDevProfilePDFPageHeight},
	})
	return builder
}

func (b *autismDevSelectedReportPDFBuilder) appendPDFBytes(content []byte) error {
	if len(content) == 0 {
		return nil
	}
	if err := importAutismDevReportPDFPages(&b.pdf, content); err != nil {
		return err
	}
	b.partCount++
	return nil
}

func (b *autismDevSelectedReportPDFBuilder) appendDirectDraw(draw func(*gopdf.GoPdf) error) error {
	if !b.fontReady {
		if err := addAutismDevProfilePDFFont(&b.pdf); err != nil {
			return err
		}
		b.fontReady = true
	}
	if err := draw(&b.pdf); err != nil {
		return err
	}
	b.partCount++
	b.directDraw++
	return nil
}

func (b *autismDevSelectedReportPDFBuilder) bytes() ([]byte, error) {
	var output bytes.Buffer
	if err := b.pdf.Write(&output); err != nil {
		return nil, fmt.Errorf("生成PDF失败：%w", err)
	}
	return output.Bytes(), nil
}
