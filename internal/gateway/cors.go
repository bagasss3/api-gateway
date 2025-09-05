package gateway

import (
	"api-gateway/internal/config"
	"net/http"
	"strings"
)

func corsMiddleware(cfg config.Config, next http.Handler) http.Handler {
	allowedOrigins := cfg.CORS.AllowedOrigins
	allowedMethods := strings.Join(cfg.CORS.AllowedMethods, ",")
	allowedHeaders := strings.Join(cfg.CORS.AllowedHeaders, ",")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			if containsStar(allowedOrigins) || containsEq(allowedOrigins, origin) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				if cfg.CORS.Credentials {
					w.Header().Set("Access-Control-Allow-Credentials", "true")
				}
				w.Header().Set("Vary", "Origin")
			}
		}
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Methods", allowedMethods)
			w.Header().Set("Access-Control-Allow-Headers", allowedHeaders)
			w.Header().Set("Access-Control-Max-Age", "600")
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func containsStar(list []string) bool {
	for _, v := range list {
		if v == "*" {
			return true
		}
	}
	return false
}
func containsEq(list []string, v string) bool {
	for _, e := range list {
		if strings.EqualFold(e, v) {
			return true
		}
	}
	return false
}
