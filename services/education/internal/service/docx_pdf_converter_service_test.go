package service

import (
	"os"
	"path/filepath"
	"strings"
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

func TestResolveDOCXPDFConverterDefaultsToSofficeDaemon(t *testing.T) {
	t.Setenv("DOCX_PDF_CONVERTER_KIND", "")
	t.Setenv("DOCX_PDF_CONVERTER_PATH", filepath.Join(t.TempDir(), "soffice"))

	converter, err := resolveDOCXPDFConverter()
	if err != nil {
		t.Fatalf("resolveDOCXPDFConverter failed: %v", err)
	}
	if converter.kind != docxPDFConverterKindSofficeDaemon {
		t.Fatalf("expected default converter kind %q, got %q", docxPDFConverterKindSofficeDaemon, converter.kind)
	}
}

func TestConvertDOCXToPDFBytesWithSofficeDaemonReusesProcess(t *testing.T) {
	stopDefaultSofficeDaemonForTest(t)
	t.Cleanup(func() {
		stopDefaultSofficeDaemonForTest(t)
	})

	scriptDir := t.TempDir()
	startLogPath := filepath.Join(scriptDir, "starts.log")
	sofficePath := filepath.Join(scriptDir, "fake-soffice.sh")
	sofficeScript := "#!/bin/sh\n" +
		"is_convert=0\n" +
		"outdir=''\n" +
		"input=''\n" +
		"while [ \"$#\" -gt 0 ]; do\n" +
		"  case \"$1\" in\n" +
		"    --convert-to)\n" +
		"      is_convert=1\n" +
		"      shift 2\n" +
		"      ;;\n" +
		"    --outdir)\n" +
		"      outdir=\"$2\"\n" +
		"      shift 2\n" +
		"      ;;\n" +
		"    *)\n" +
		"      input=\"$1\"\n" +
		"      shift\n" +
		"      ;;\n" +
		"  esac\n" +
		"done\n" +
		"if [ \"$is_convert\" = '1' ]; then\n" +
		"  base=\"$(basename \"$input\" .docx)\"\n" +
		"  printf '%s\\n' '%PDF-1.4' '1 0 obj<<>>endobj' 'trailer<<>>' '%%EOF' > \"$outdir/$base.pdf\"\n" +
		"  exit 0\n" +
		"fi\n" +
		"printf 'start\\n' >> \"$FAKE_SOFFICE_START_LOG\"\n" +
		"while :; do sleep 1; done\n"
	if err := os.WriteFile(sofficePath, []byte(sofficeScript), 0o755); err != nil {
		t.Fatalf("write fake soffice failed: %v", err)
	}

	t.Setenv("DOCX_PDF_CONVERTER_KIND", docxPDFConverterKindSofficeDaemon)
	t.Setenv("DOCX_PDF_CONVERTER_PATH", sofficePath)
	t.Setenv("DOCX_PDF_DAEMON_SKIP_READY_CHECK", "1")
	t.Setenv("FAKE_SOFFICE_START_LOG", startLogPath)

	for i := 0; i < 2; i++ {
		pdfBytes, err := convertDOCXToPDFBytes([]byte("fake-docx-bytes"), "常驻转换测试.docx")
		if err != nil {
			t.Fatalf("convertDOCXToPDFBytes #%d failed: %v", i+1, err)
		}
		if len(pdfBytes) < 8 || string(pdfBytes[:8]) != "%PDF-1.4" {
			t.Fatalf("unexpected pdf header on conversion #%d: %q", i+1, string(pdfBytes[:minIntForTest(len(pdfBytes), 8)]))
		}
	}

	content, err := os.ReadFile(startLogPath)
	if err != nil {
		t.Fatalf("read fake soffice start log failed: %v", err)
	}
	if starts := strings.Count(string(content), "start"); starts != 1 {
		t.Fatalf("expected one daemon process start for two conversions, got %d", starts)
	}
}

func stopDefaultSofficeDaemonForTest(t *testing.T) {
	t.Helper()
	defaultSofficeDaemon.mu.Lock()
	defer defaultSofficeDaemon.mu.Unlock()
	defaultSofficeDaemon.stopAndRemoveLocked()
}

func minIntForTest(a, b int) int {
	if a < b {
		return a
	}
	return b
}
