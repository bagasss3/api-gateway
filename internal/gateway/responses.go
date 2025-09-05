package gateway

import (
	crypto_rand "crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
)

type JsonResponse struct {
	RequestId string      `json:"request_id"`
	Status    int         `json:"status_code"`
	Messages  string      `json:"messages"`
	Data      interface{} `json:"data"`
}

type JsonResponseTotal struct {
	RequestId string      `json:"request_id"`
	Status    int         `json:"status_code"`
	Messages  string      `json:"messages"`
	Total     int         `json:"total"`
	Data      interface{} `json:"data"`
}

type JsonResponsError struct {
	RequestId        string      `json:"request_id"`
	StatusCode       int         `json:"status_code"`
	ErrorCode        interface{} `json:"error_code"`
	ErrorMessage     string      `json:"error_message"`
	DeveloperMessage interface{} `json:"developer_message"`
}

func requestIDFrom(r *http.Request) string {
	if v := r.Header.Get("X-Request-Id"); v != "" {
		return v
	}
	var b [16]byte
	_, _ = crypto_rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func jsonError(w http.ResponseWriter, r *http.Request, status int, code string, msg string, dev interface{}) {
	writeJSON(w, status, JsonResponsError{
		RequestId:        requestIDFrom(r),
		StatusCode:       status,
		ErrorCode:        code,
		ErrorMessage:     msg,
		DeveloperMessage: dev,
	})
}
