package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
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
		content, err := svc.autismDevSelectedReportSectionPDF(userID, recordID, normalizedSections[0], analysis)
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
	for index, section := range normalizedSections {
		sectionStartedAt := time.Now()
		if isAutismDevSelectedReportDirectDrawSection(section) {
			if err := svc.appendAutismDevSelectedReportDirectSectionPDF(builder, userID, recordID, record, section, institutionName, analysis); err != nil {
				return "", "", nil, err
			}
			logx.Info("autismdev selected report direct section pdf finished", logx.Entry{
				"record_id":    recordID,
				"section":      section,
				"duration_ms":  time.Since(sectionStartedAt).Milliseconds(),
				"section_from": index,
			})
			continue
		}
		return "", "", nil, fmt.Errorf("不支持的打印内容：%s", section)
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

func (svc *Service) autismDevSelectedReportSectionPDF(userID int64, recordID int64, section string, analysis *model.AutismDevResultAnalysisVO) ([]byte, error) {
	switch section {
	case AutismDevReportSectionAssessmentInfo:
		_, content, err := svc.buildAutismDevAssessmentSituationPDF(userID, recordID)
		return content, err
	case AutismDevReportSectionResultAnalysis:
		export, err := svc.autismDevResultAnalysisWordExport(userID, recordID, analysis)
		if err != nil {
			return nil, err
		}
		content, err := buildAutismDevResultAnalysisPDF(export)
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

func (svc *Service) appendAutismDevSelectedReportDirectSectionPDF(builder *autismDevSelectedReportPDFBuilder, userID int64, recordID int64, record model.AssessmentRecordDetailVO, section string, institutionName string, analysis *model.AutismDevResultAnalysisVO) error {
	switch section {
	case AutismDevReportSectionAssessmentInfo:
		_, record, score, data, itemScores, err := svc.autismDevResultAnalysisContext(userID, recordID)
		if err != nil {
			return err
		}
		export := buildAutismDevAssessmentSituationWordExport(record, score, data, itemScores)
		return builder.appendDirectDraw(func(pdf *gopdf.GoPdf) error {
			return drawAutismDevAssessmentSituationPDFPages(pdf, export)
		})
	case AutismDevReportSectionResultAnalysis:
		export, err := svc.autismDevResultAnalysisWordExport(userID, recordID, analysis)
		if err != nil {
			return err
		}
		return builder.appendDirectDraw(func(pdf *gopdf.GoPdf) error {
			return drawAutismDevResultAnalysisPDFPages(pdf, export)
		})
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

func isAutismDevSelectedReportDirectDrawSection(section string) bool {
	return section == AutismDevReportSectionAssessmentInfo ||
		section == AutismDevReportSectionResultAnalysis ||
		section == AutismDevReportSectionTraining ||
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
