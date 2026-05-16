package service

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"go-migration-platform/pkg/logx"
)

const (
	docxPDFConverterKindScript        = "script"
	docxPDFConverterKindSoffice       = "soffice"
	docxPDFConverterKindSofficeDaemon = "soffice_daemon"
)

type docxPDFConverter struct {
	kind string
	path string
}

var defaultSofficeDaemon = &sofficeDaemonConverter{}

func convertDOCXToPDFBytes(docxBytes []byte, sourceFileName string) ([]byte, error) {
	if len(docxBytes) == 0 {
		return nil, fmt.Errorf("empty docx content")
	}
	startedAt := time.Now()
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
	case docxPDFConverterKindScript:
		err = runDOCXPDFScriptConverter(ctx, converter.path, inputPath, outputPath, tempDir)
	case docxPDFConverterKindSofficeDaemon:
		err = runSofficeDaemonDOCXPDFConverter(ctx, converter.path, inputPath, outputPath, tempDir)
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
	logx.Info("docx pdf conversion finished", logx.Entry{
		"source":       sourceFileName,
		"converter":    converter.kind,
		"docx_bytes":   len(docxBytes),
		"pdf_bytes":    len(data),
		"duration_ms":  time.Since(startedAt).Milliseconds(),
		"soffice_path": converter.path,
	})
	return data, nil
}

func resolveDOCXPDFConverter() (docxPDFConverter, error) {
	kind := strings.ToLower(strings.TrimSpace(os.Getenv("DOCX_PDF_CONVERTER_KIND")))
	if kind == "" {
		kind = docxPDFConverterKindSofficeDaemon
	}
	path := strings.TrimSpace(os.Getenv("DOCX_PDF_CONVERTER_PATH"))
	switch kind {
	case docxPDFConverterKindScript:
		if path == "" {
			return docxPDFConverter{}, fmt.Errorf("DOCX转PDF脚本未配置，请设置 `DOCX_PDF_CONVERTER_PATH`")
		}
		return docxPDFConverter{kind: kind, path: path}, nil
	case "daemon", "libreoffice_daemon":
		kind = docxPDFConverterKindSofficeDaemon
		fallthrough
	case docxPDFConverterKindSoffice, docxPDFConverterKindSofficeDaemon:
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

func runSofficeDaemonDOCXPDFConverter(ctx context.Context, execPath, inputPath, outputPath, tempDir string) error {
	if err := defaultSofficeDaemon.convert(ctx, execPath, inputPath, outputPath, tempDir); err == nil {
		return nil
	} else {
		coldStartErr := runSofficeDOCXPDFConverter(ctx, execPath, inputPath, outputPath, tempDir)
		if coldStartErr == nil {
			return nil
		}
		return fmt.Errorf("DOCX转PDF常驻转换失败: %w；单次转换兜底也失败: %v", err, coldStartErr)
	}
}

func runSofficeDOCXPDFConverter(ctx context.Context, execPath, inputPath, outputPath, tempDir string) error {
	profileDir := filepath.Join(tempDir, "lo-profile")
	if err := os.MkdirAll(profileDir, 0o755); err != nil {
		return err
	}
	return runSofficeDOCXPDFConverterWithProfile(ctx, execPath, inputPath, outputPath, tempDir, profileDir, tempDir)
}

func runSofficeDOCXPDFConverterWithProfile(ctx context.Context, execPath, inputPath, outputPath, outDir, profileDir, homeDir string) error {
	startedAt := time.Now()
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
		outDir,
		inputPath,
	)
	cmd.Env = append(os.Environ(), "HOME="+homeDir, "TMPDIR="+outDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("DOCX转PDF失败: %w, output=%s", err, strings.TrimSpace(string(output)))
	}
	if _, err := os.Stat(outputPath); err != nil {
		return fmt.Errorf("DOCX转PDF失败：未生成PDF文件，output=%s", strings.TrimSpace(string(output)))
	}
	logx.Info("soffice convert command finished", logx.Entry{
		"input":       filepath.Base(inputPath),
		"duration_ms": time.Since(startedAt).Milliseconds(),
	})
	return nil
}

type sofficeDaemonConverter struct {
	mu         sync.Mutex
	execPath   string
	rootDir    string
	profileDir string
	host       string
	port       int
	cmd        *exec.Cmd
	done       chan error
	output     strings.Builder
}

func (d *sofficeDaemonConverter) convert(ctx context.Context, execPath, inputPath, outputPath, tempDir string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if err := d.ensureStartedLocked(execPath); err != nil {
		return err
	}
	firstErr := d.convertWithDaemonLocked(ctx, inputPath, outputPath, tempDir)
	if firstErr == nil {
		return nil
	}
	d.stopLocked()
	if err := d.ensureStartedLocked(execPath); err != nil {
		return fmt.Errorf("%w；重启LibreOffice常驻进程失败: %v", firstErr, err)
	}
	if err := d.convertWithDaemonLocked(ctx, inputPath, outputPath, tempDir); err == nil {
		return nil
	} else {
		return err
	}
}

func (d *sofficeDaemonConverter) convertWithDaemonLocked(ctx context.Context, inputPath, outputPath, tempDir string) error {
	warmErr := runSofficeDOCXPDFConverterWithProfile(ctx, d.execPath, inputPath, outputPath, tempDir, d.profileDir, d.rootDir)
	if warmErr == nil {
		return nil
	}
	return fmt.Errorf("常驻profile转换失败: %w", warmErr)
}

func (d *sofficeDaemonConverter) ensureStartedLocked(execPath string) error {
	execPath = strings.TrimSpace(execPath)
	if execPath == "" {
		return buildDOCXToPDFNotConfiguredError()
	}
	if d.execPath != "" && d.execPath != execPath {
		d.stopAndRemoveLocked()
	}
	d.execPath = execPath
	if d.isReadyLocked() {
		return nil
	}
	d.stopLocked()
	if d.rootDir == "" {
		rootDir, err := os.MkdirTemp("", "iep-docx-pdf-daemon-*")
		if err != nil {
			return err
		}
		d.rootDir = rootDir
		d.profileDir = filepath.Join(rootDir, "lo-profile")
		if err := os.MkdirAll(d.profileDir, 0o755); err != nil {
			d.stopAndRemoveLocked()
			return err
		}
	}
	if d.port == 0 {
		port, err := findFreeLocalTCPPort()
		if err != nil {
			return err
		}
		d.host = "127.0.0.1"
		d.port = port
	}
	acceptArg := fmt.Sprintf("socket,host=%s,port=%d;urp;StarOffice.ComponentContext", d.host, d.port)
	userInstallationURL := "file://" + filepath.ToSlash(d.profileDir)
	cmd := exec.Command(
		d.execPath,
		"--headless",
		"--invisible",
		"--nologo",
		"--nodefault",
		"--nolockcheck",
		"--nofirststartwizard",
		"--norestore",
		"-env:UserInstallation="+userInstallationURL,
		"--accept="+acceptArg,
	)
	cmd.Env = append(os.Environ(), "HOME="+d.rootDir, "TMPDIR="+d.rootDir)
	cmd.Stdout = &d.output
	cmd.Stderr = &d.output
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动LibreOffice常驻进程失败: %w", err)
	}
	d.cmd = cmd
	d.done = make(chan error, 1)
	go func() {
		d.done <- cmd.Wait()
	}()
	if shouldSkipSofficeDaemonReadyCheck() {
		logx.Info("soffice daemon started", logx.Entry{
			"port":        d.port,
			"profile_dir": d.profileDir,
		})
		return nil
	}
	if err := d.waitUntilReadyLocked(); err != nil {
		d.stopLocked()
		return err
	}
	logx.Info("soffice daemon started", logx.Entry{
		"port":        d.port,
		"profile_dir": d.profileDir,
	})
	return nil
}

func (d *sofficeDaemonConverter) waitUntilReadyLocked() error {
	timeout := sofficeDaemonStartTimeout()
	deadline := time.Now().Add(timeout)
	address := net.JoinHostPort(d.host, strconv.Itoa(d.port))
	for time.Now().Before(deadline) {
		if d.cmdExitedLocked() {
			return fmt.Errorf("LibreOffice常驻进程启动后退出: %s", strings.TrimSpace(d.output.String()))
		}
		conn, err := net.DialTimeout("tcp", address, 200*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			return nil
		}
		time.Sleep(150 * time.Millisecond)
	}
	return fmt.Errorf("LibreOffice常驻进程启动超时（%s）", timeout)
}

func (d *sofficeDaemonConverter) isReadyLocked() bool {
	if d.host == "" || d.port == 0 || d.cmdExitedLocked() {
		return false
	}
	if shouldSkipSofficeDaemonReadyCheck() {
		return true
	}
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(d.host, strconv.Itoa(d.port)), 200*time.Millisecond)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

func (d *sofficeDaemonConverter) cmdExitedLocked() bool {
	if d.done == nil {
		return false
	}
	select {
	case <-d.done:
		d.cmd = nil
		d.done = nil
		return true
	default:
		return false
	}
}

func (d *sofficeDaemonConverter) stopLocked() {
	if d.cmd != nil && d.cmd.Process != nil {
		_ = d.cmd.Process.Kill()
	}
	if d.done != nil {
		select {
		case <-d.done:
		case <-time.After(2 * time.Second):
		}
	}
	d.cmd = nil
	d.done = nil
}

func (d *sofficeDaemonConverter) stopAndRemoveLocked() {
	d.stopLocked()
	if d.rootDir != "" {
		_ = os.RemoveAll(d.rootDir)
	}
	d.rootDir = ""
	d.profileDir = ""
	d.host = ""
	d.port = 0
	d.output.Reset()
}

func findFreeLocalTCPPort() (int, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()
	addr, ok := listener.Addr().(*net.TCPAddr)
	if !ok || addr.Port == 0 {
		return 0, errors.New("无法分配LibreOffice常驻监听端口")
	}
	return addr.Port, nil
}

func sofficeDaemonStartTimeout() time.Duration {
	raw := strings.TrimSpace(os.Getenv("DOCX_PDF_DAEMON_START_TIMEOUT_MS"))
	if raw == "" {
		return 8 * time.Second
	}
	ms, err := strconv.Atoi(raw)
	if err != nil || ms <= 0 {
		return 8 * time.Second
	}
	return time.Duration(ms) * time.Millisecond
}

func shouldSkipSofficeDaemonReadyCheck() bool {
	return strings.TrimSpace(os.Getenv("DOCX_PDF_DAEMON_SKIP_READY_CHECK")) == "1"
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
