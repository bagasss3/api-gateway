package gateway

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func validateAPIKey(r *http.Request, allowed []string) bool {
	key := r.Header.Get("X-API-Key")
	if key == "" || len(allowed) == 0 {
		return false
	}
	for _, k := range allowed {
		if subtle.ConstantTimeCompare([]byte(key), []byte(k)) == 1 {
			return true
		}
	}
	return false
}

func bearerTokenFrom(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if h == "" {
		return ""
	}
	parts := strings.SplitN(h, " ", 2)
	if len(parts) != 2 {
		return ""
	}
	if !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func validateJWTViaAuthService(ctx context.Context, introspectURL string, token string) (map[string]string, error) {
	if introspectURL == "" {
		return nil, errors.New("AUTH introspect URL not set")
	}
	if token == "" {
		return nil, errors.New("missing bearer token")
	}

	reqURL := introspectURL
	if strings.Contains(introspectURL, "?") {
		reqURL += "&"
	} else {
		reqURL += "?"
	}
	reqURL += "token=" + url.QueryEscape(token)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return map[string]string{"X-User-Id": "unknown", "X-User-Name": "unknown"}, nil
	case http.StatusUnauthorized, http.StatusForbidden:
		return nil, errors.New("invalid token")
	default:
		return nil, fmt.Errorf("auth service error: %d", resp.StatusCode)
	}
}
