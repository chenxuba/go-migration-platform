package tenant

import (
	"context"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"
)

type contextKey string

const contextTenantKey contextKey = "tenant_context"

var requestCounter uint64

type Context struct {
	TenantID  string `json:"tenantId"`
	UserID    string `json:"userId"`
	RequestID string `json:"requestId"`
}

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "tenant-a"
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
		ctx := context.WithValue(r.Context(), contextTenantKey, Context{
			TenantID:  tenantID,
			UserID:    userID,
			RequestID: requestID,
		})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func FromContext(ctx context.Context) Context {
	value, _ := ctx.Value(contextTenantKey).(Context)
	return value
}
