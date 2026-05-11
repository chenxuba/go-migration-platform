package service

import (
	"fmt"
	"path/filepath"
	"strings"

	"go-migration-platform/services/education/internal/model"
)

const iepPlanPDFContentType = "application/pdf"

func (svc *Service) ExportPEP3IEPPlanPDF(userID int64, recordID int64, durationMonths int) (string, string, []byte, error) {
	fileName, _, content, err := svc.ExportPEP3IEPPlanWord(userID, recordID, durationMonths)
	if err != nil {
		return "", "", nil, err
	}
	return exportIEPPDFByDOCX(fileName, content)
}

func (svc *Service) ExportPEP3IEPPlanPDFFromAIResult(userID int64, recordID int64, planResult model.PEP3IEPPlanAIResult, durationMonths int) (string, string, []byte, error) {
	fileName, _, content, err := svc.ExportPEP3IEPPlanWordFromAIResult(userID, recordID, planResult, durationMonths)
	if err != nil {
		return "", "", nil, err
	}
	return exportIEPPDFByDOCX(fileName, content)
}

func (svc *Service) ExportPEP3ExecutionPlanPDF(userID int64, req model.PEP3ExecutionPlanWordExportRequest) (string, string, []byte, error) {
	fileName, _, content, err := svc.ExportPEP3ExecutionPlanWord(userID, req)
	if err != nil {
		return "", "", nil, err
	}
	return exportIEPPDFByDOCX(fileName, content)
}

func exportIEPPDFByDOCX(sourceFileName string, docxBytes []byte) (string, string, []byte, error) {
	pdfBytes, err := convertDOCXToPDFBytes(docxBytes, sourceFileName)
	if err != nil {
		return "", "", nil, err
	}
	baseName := strings.TrimSpace(sourceFileName)
	if baseName == "" {
		baseName = "document.docx"
	}
	baseName = strings.TrimSuffix(baseName, filepath.Ext(baseName)) + ".pdf"
	return baseName, iepPlanPDFContentType, pdfBytes, nil
}

func buildDOCXToPDFNotConfiguredError() error {
	return fmt.Errorf("DOCX转PDF转换器未配置，请安装LibreOffice并配置 `DOCX_PDF_CONVERTER_PATH`，或提供自定义脚本转换器")
}
