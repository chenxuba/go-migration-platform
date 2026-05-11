package service

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type docxPDFConverter struct {
	kind string
	path string
}

func convertDOCXToPDFBytes(docxBytes []byte, sourceFileName string) ([]byte, error) {
	if len(docxBytes) == 0 {
		return nil, fmt.Errorf("empty docx content")
	}
	converter, err := resolveDOCXPDFConverter()
	if err != nil {
		return nil, err
	}
	tempDir, err := os.MkdirTemp("", "iep-docx-pdf-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tempDir)

	baseName := sanitizeExportFileName(strings.TrimSuffix(strings.TrimSpace(sourceFileName), filepath.Ext(strings.TrimSpace(sourceFileName))))
	if baseName == "" {
		baseName = "document"
	}
	inputPath := filepath.Join(tempDir, baseName+".docx")
	outputPath := filepath.Join(tempDir, baseName+".pdf")
	if err := os.WriteFile(inputPath, docxBytes, 0o644); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	switch converter.kind {
	case "script":
		err = runDOCXPDFScriptConverter(ctx, converter.path, inputPath, outputPath, tempDir)
	default:
		err = runSofficeDOCXPDFConverter(ctx, converter.path, inputPath, outputPath, tempDir)
	}
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(outputPath)
	if err != nil {
		return nil, fmt.Errorf("读取转换后的PDF失败: %w", err)
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("DOCX转PDF失败：输出PDF为空")
	}
	return data, nil
}

func resolveDOCXPDFConverter() (docxPDFConverter, error) {
	kind := strings.ToLower(strings.TrimSpace(os.Getenv("DOCX_PDF_CONVERTER_KIND")))
	if kind == "" {
		kind = "soffice"
	}
	path := strings.TrimSpace(os.Getenv("DOCX_PDF_CONVERTER_PATH"))
	switch kind {
	case "script":
		if path == "" {
			return docxPDFConverter{}, fmt.Errorf("DOCX转PDF脚本未配置，请设置 `DOCX_PDF_CONVERTER_PATH`")
		}
		return docxPDFConverter{kind: kind, path: path}, nil
	case "soffice":
		if path == "" {
			path = findSofficeExecPath()
		}
		if path == "" {
			return docxPDFConverter{}, buildDOCXToPDFNotConfiguredError()
		}
		return docxPDFConverter{kind: kind, path: path}, nil
	default:
		return docxPDFConverter{}, fmt.Errorf("unsupported DOCX_PDF_CONVERTER_KIND: %s", kind)
	}
}

func runSofficeDOCXPDFConverter(ctx context.Context, execPath, inputPath, outputPath, tempDir string) error {
	profileDir := filepath.Join(tempDir, "lo-profile")
	if err := os.MkdirAll(profileDir, 0o755); err != nil {
		return err
	}
	userInstallationURL := "file://" + filepath.ToSlash(profileDir)
	cmd := exec.CommandContext(
		ctx,
		execPath,
		"--headless",
		"--nologo",
		"--nodefault",
		"--nolockcheck",
		"--nofirststartwizard",
		"-env:UserInstallation="+userInstallationURL,
		"--convert-to",
		"pdf:writer_pdf_Export",
		"--outdir",
		tempDir,
		inputPath,
	)
	cmd.Env = append(os.Environ(), "HOME="+tempDir, "TMPDIR="+tempDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("DOCX转PDF失败: %w, output=%s", err, strings.TrimSpace(string(output)))
	}
	if _, err := os.Stat(outputPath); err != nil {
		return fmt.Errorf("DOCX转PDF失败：未生成PDF文件，output=%s", strings.TrimSpace(string(output)))
	}
	return nil
}

func runDOCXPDFScriptConverter(ctx context.Context, execPath, inputPath, outputPath, tempDir string) error {
	cmd := exec.CommandContext(ctx, execPath, inputPath, outputPath)
	cmd.Env = append(os.Environ(), "TMPDIR="+tempDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("DOCX转PDF脚本执行失败: %w, output=%s", err, strings.TrimSpace(string(output)))
	}
	if _, err := os.Stat(outputPath); err != nil {
		return fmt.Errorf("DOCX转PDF脚本未生成PDF文件，output=%s", strings.TrimSpace(string(output)))
	}
	return nil
}

func findSofficeExecPath() string {
	candidates := []string{
		strings.TrimSpace(os.Getenv("SOFFICE_PATH")),
		strings.TrimSpace(os.Getenv("LIBREOFFICE_PATH")),
		"soffice",
		"libreoffice",
		"/Applications/LibreOffice.app/Contents/MacOS/soffice",
		"/usr/bin/soffice",
		"/usr/bin/libreoffice",
		"/usr/local/bin/soffice",
		"/usr/local/bin/libreoffice",
		"/opt/homebrew/bin/soffice",
		"/opt/homebrew/bin/libreoffice",
		"/snap/bin/libreoffice",
		"/opt/libreoffice/program/soffice",
	}
	for _, candidate := range candidates {
		candidate = strings.TrimSpace(candidate)
		if candidate == "" {
			continue
		}
		if filepath.IsAbs(candidate) {
			if _, err := os.Stat(candidate); err == nil {
				return candidate
			}
			continue
		}
		if resolved, err := exec.LookPath(candidate); err == nil {
			return resolved
		}
	}
	return ""
}
