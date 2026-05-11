package service

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConvertDOCXToPDFBytesWithScriptConverter(t *testing.T) {
	scriptDir := t.TempDir()
	scriptPath := filepath.Join(scriptDir, "docx-to-pdf.sh")
	script := "#!/bin/sh\n" +
		"input=\"$1\"\n" +
		"output=\"$2\"\n" +
		"printf '%s\n' '%PDF-1.4' '1 0 obj<<>>endobj' 'trailer<<>>' '%%EOF' > \"$output\"\n" +
		"exit 0\n"
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatalf("write script failed: %v", err)
	}
	t.Setenv("DOCX_PDF_CONVERTER_KIND", "script")
	t.Setenv("DOCX_PDF_CONVERTER_PATH", scriptPath)

	pdfBytes, err := convertDOCXToPDFBytes([]byte("fake-docx-bytes"), "测试计划.docx")
	if err != nil {
		t.Fatalf("convertDOCXToPDFBytes failed: %v", err)
	}
	if len(pdfBytes) == 0 {
		t.Fatalf("expected pdf bytes")
	}
	if string(pdfBytes[:8]) != "%PDF-1.4" {
		t.Fatalf("unexpected pdf header: %q", string(pdfBytes[:8]))
	}
}
