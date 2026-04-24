package tenant

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync/atomic"
	"time"
)

type contextKey string

const contextTenantKey contextKey = "tenant_context"

var requestCounter uint64

type Context struct {
	TenantID     string `json:"tenantId"`
	UserID       string `json:"userId"`
	RequestID    string `json:"requestId"`
	TenantSource string `json:"tenantSource,omitempty"`
	Host         string `json:"host,omitempty"`
}

var domainMapCache atomic.Value

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host := normalizeHost(firstNonEmpty(
			r.Header.Get("X-Tenant-Domain"),
			r.Header.Get("X-Forwarded-Host"),
			hostFromURLHeader(r.Header.Get("Origin")),
			hostFromURLHeader(r.Header.Get("Referer")),
			r.Host,
		))
		tenantID, tenantSource := resolveTenantFromDomain(host)
		if tenantID == "" {
			tenantID = strings.TrimSpace(r.Header.Get("X-Tenant-ID"))
			if tenantID != "" {
				tenantSource = "header"
			}
		}
		if tenantID == "" {
			tenantID = "tenant-a"
			tenantSource = "default"
		}

		userID := r.Header.Get("X-User-ID")
		if userID == "" {
			userID = "anonymous"
		}

		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			seq := atomic.AddUint64(&requestCounter, 1)
			requestID = fmt.Sprintf("req-%d-%d", time.Now().UnixMilli(), seq)
		}

		w.Header().Set("X-Request-ID", requestID)
		w.Header().Set("X-Tenant-ID", tenantID)
		ctx := context.WithValue(r.Context(), contextTenantKey, Context{
			TenantID:     tenantID,
			UserID:       userID,
			RequestID:    requestID,
			TenantSource: tenantSource,
			Host:         host,
		})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func hostFromURLHeader(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" {
		return ""
	}
	return parsed.Host
}

func normalizeHost(host string) string {
	host = strings.TrimSpace(strings.ToLower(host))
	if host == "" {
		return ""
	}
	if hostname, _, err := net.SplitHostPort(host); err == nil {
		return strings.TrimSpace(strings.ToLower(hostname))
	}
	return host
}

func resolveTenantFromDomain(host string) (string, string) {
	if host == "" {
		return "", ""
	}
	value := domainMapCache.Load()
	if value == nil {
		value = parseDomainMap(os.Getenv("TENANT_DOMAIN_MAP"))
		domainMapCache.Store(value)
	}
	domainMap, _ := value.(map[string]string)
	if tenantID := strings.TrimSpace(domainMap[host]); tenantID != "" {
		return tenantID, "domain"
	}
	return "", ""
}

func parseDomainMap(raw string) map[string]string {
	result := make(map[string]string)
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return result
	}

	var jsonMap map[string]string
	if err := json.Unmarshal([]byte(trimmed), &jsonMap); err == nil {
		for domain, tenantID := range jsonMap {
			if domain = normalizeHost(domain); domain != "" && strings.TrimSpace(tenantID) != "" {
				result[domain] = strings.TrimSpace(tenantID)
			}
		}
		return result
	}

	for _, pair := range strings.Split(trimmed, ",") {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			continue
		}
		domain := normalizeHost(parts[0])
		tenantID := strings.TrimSpace(parts[1])
		if domain != "" && tenantID != "" {
			result[domain] = tenantID
		}
	}
	return result
}

func FromContext(ctx context.Context) Context {
	value, _ := ctx.Value(contextTenantKey).(Context)
	return value
}
