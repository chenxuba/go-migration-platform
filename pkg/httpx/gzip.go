package httpx

import (
	"compress/gzip"
	"net/http"
	"strings"
)

const gzipMinResponseBytes = 1024

type gzipResponseWriter struct {
	http.ResponseWriter
	request        *http.Request
	status         int
	gzipWriter     *gzip.Writer
	wroteHeader    bool
	decided        bool
	shouldGzip     bool
	buffer         []byte
	headerSnapshot http.Header
}

func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !clientAcceptsGzip(r) || r.Method == http.MethodHead || isStreamingPath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}
		wrapped := &gzipResponseWriter{
			ResponseWriter: w,
			request:        r,
			status:         http.StatusOK,
			buffer:         make([]byte, 0, gzipMinResponseBytes),
		}
		defer wrapped.Close()
		next.ServeHTTP(wrapped, r)
	})
}

func (w *gzipResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

func (w *gzipResponseWriter) WriteHeader(status int) {
	if w.wroteHeader {
		return
	}
	w.status = status
	w.wroteHeader = true
	w.headerSnapshot = cloneHeader(w.ResponseWriter.Header())
}

func (w *gzipResponseWriter) Write(data []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	if !w.decided {
		w.buffer = append(w.buffer, data...)
		if len(w.buffer) < gzipMinResponseBytes {
			return len(data), nil
		}
		w.decideCompression()
		if err := w.flushBuffered(); err != nil {
			return 0, err
		}
		return len(data), nil
	}
	if w.shouldGzip {
		_, err := w.gzipWriter.Write(data)
		return len(data), err
	}
	return w.ResponseWriter.Write(data)
}

func (w *gzipResponseWriter) Flush() {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	if !w.decided {
		w.decideCompression()
		_ = w.flushBuffered()
	}
	if w.gzipWriter != nil {
		_ = w.gzipWriter.Flush()
	}
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (w *gzipResponseWriter) Close() {
	if !w.wroteHeader {
		return
	}
	if !w.decided {
		w.decideCompression()
		_ = w.flushBuffered()
	}
	if w.gzipWriter != nil {
		_ = w.gzipWriter.Close()
	}
}

func (w *gzipResponseWriter) decideCompression() {
	if w.decided {
		return
	}
	w.decided = true
	header := w.ResponseWriter.Header()
	if w.headerSnapshot != nil {
		copyHeader(header, w.headerSnapshot)
	}
	w.shouldGzip = len(w.buffer) >= gzipMinResponseBytes && shouldGzipResponse(w.request, w.status, header)
	if w.shouldGzip {
		header.Del("Content-Length")
		header.Set("Content-Encoding", "gzip")
		header.Add("Vary", "Accept-Encoding")
		w.gzipWriter = gzip.NewWriter(w.ResponseWriter)
	}
	w.ResponseWriter.WriteHeader(w.status)
}

func (w *gzipResponseWriter) flushBuffered() error {
	if len(w.buffer) == 0 {
		return nil
	}
	buffer := w.buffer
	w.buffer = nil
	if w.shouldGzip {
		_, err := w.gzipWriter.Write(buffer)
		return err
	}
	_, err := w.ResponseWriter.Write(buffer)
	return err
}

func shouldGzipResponse(r *http.Request, status int, header http.Header) bool {
	if status == http.StatusNoContent || status == http.StatusNotModified {
		return false
	}
	if header.Get("Content-Encoding") != "" {
		return false
	}
	contentType := strings.ToLower(header.Get("Content-Type"))
	if contentType == "" {
		return false
	}
	if strings.Contains(contentType, "text/event-stream") {
		return false
	}
	if strings.Contains(contentType, "application/json") || strings.HasPrefix(contentType, "text/") {
		return true
	}
	return false
}

func clientAcceptsGzip(r *http.Request) bool {
	for _, part := range strings.Split(r.Header.Get("Accept-Encoding"), ",") {
		if strings.EqualFold(strings.TrimSpace(strings.Split(part, ";")[0]), "gzip") {
			return true
		}
	}
	return false
}

func isStreamingPath(path string) bool {
	return strings.Contains(strings.ToLower(path), "/stream")
}

func cloneHeader(src http.Header) http.Header {
	out := make(http.Header, len(src))
	copyHeader(out, src)
	return out
}

func copyHeader(dst, src http.Header) {
	for key := range dst {
		delete(dst, key)
	}
	for key, values := range src {
		dst[key] = append([]string(nil), values...)
	}
}
