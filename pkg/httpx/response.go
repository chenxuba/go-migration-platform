package httpx

import (
	"encoding/json"
	"net/http"
)

type Envelope struct {
	Success   bool   `json:"success"`
	Message   string `json:"message,omitempty"`
	RequestID string `json:"requestId,omitempty"`
	Data      any    `json:"data,omitempty"`
}

func WriteJSON(w http.ResponseWriter, status int, data any, requestID string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(Envelope{
		Success:   status < http.StatusBadRequest,
		RequestID: requestID,
		Data:      data,
	})
}

func WriteError(w http.ResponseWriter, status int, message, requestID string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(Envelope{
		Success:   false,
		Message:   message,
		RequestID: requestID,
	})
}
