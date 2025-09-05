package gateway

import (
	"api-gateway/internal/config"
	"net/http"
	"net/url"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

func BuildHandler(cfg config.Config) http.Handler {
	reg := &Registry{}

	mkBackend := func(raw string) *Backend {
		u, err := url.Parse(raw)
		if err != nil {
			log.Fatalf("route target parse: %v", err)
		}
		b := &Backend{URL: u}
		b.Healthy.Store(true)
		return b
	}

	for _, rc := range cfg.Routes {
		var targets []*Backend
		for _, t := range rc.Targets {
			targets = append(targets, mkBackend(t))
		}
		rt := &Route{
			Name:           rc.Name,
			Prefix:         rc.Prefix,
			StripPrefix:    rc.StripPrefix,
			Auth:           parseAuthMode(rc.Auth),
			RequireAPIKey:  rc.RequireAPIKey,
			PublicPrefixes: rc.PublicPrefixes,
			Targets:        targets,
		}
		reg.Add(rt)
	}

	// Active health checks
	go func() {
		for {
			for _, rt := range reg.Routes {
				for _, b := range rt.Targets {
					alive := probe(b.URL)
					b.Healthy.Store(alive)
					if !alive {
						b.Fails.Add(1)
					} else {
						b.Fails.Store(0)
					}
				}
			}
			time.Sleep(5 * time.Second)
		}
	}()

	limiter := newIPLimiter(cfg.Limits.GlobalRPS)
	sem := make(chan struct{}, cfg.Limits.Concurrency)

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r)
		if !limiter.allow(ip, cfg.Limits.PerIPRPS, cfg.Limits.PerIPBurst) {
			jsonError(w, r, http.StatusTooManyRequests, "RATE_LIMITED",
				"Too many requests", "IP rate limit exceeded")
			return
		}
		select {
		case sem <- struct{}{}:
			defer func() { <-sem }()
		default:
			jsonError(w, r, http.StatusServiceUnavailable, "CONCURRENCY_LIMIT",
				"Server is busy", "Gateway concurrency cap reached")
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, int64(cfg.Limits.MaxBodyMB)*1024*1024)

		rt, rest := reg.Match(r.URL.Path)
		if rt == nil {
			jsonError(w, r, http.StatusNotFound, "ROUTE_NOT_FOUND",
				"Not Found", "No matching route for "+r.URL.Path)
			return
		}

		effectiveAuth := rt.Auth
		if isPublicPath(rt, rest) {
			if rt.RequireAPIKey && !validateAPIKey(r, cfg.APIKeys) {
				jsonError(w, r, http.StatusUnauthorized, "UNAUTHORIZED",
					"Missing or invalid API key", "X-API-Key not present or invalid")
				return
			}
			effectiveAuth = AuthNone
		}

		userHeaders := map[string]string{}
		switch effectiveAuth {
		case AuthJWT:
			token := bearerTokenFrom(r)
			info, err := validateJWTViaAuthService(r.Context(), cfg.Auth.IntrospectURL, token)
			if err != nil {
				log.WithError(err).Warn("auth failed")
				jsonError(w, r, http.StatusUnauthorized, "UNAUTHORIZED",
					"Invalid or missing access token", err.Error())
				return
			}
			for k, v := range info {
				userHeaders[k] = v
			}
		case AuthAPIKey:
			if !validateAPIKey(r, cfg.APIKeys) {
				jsonError(w, r, http.StatusUnauthorized, "UNAUTHORIZED",
					"Missing or invalid API key", "X-API-Key not present or invalid")
				return
			}
		}

		// Strip prefix before proxying
		rest = strings.TrimPrefix(r.URL.Path, rt.StripPrefix)
		if !strings.HasPrefix(rest, "/") {
			rest = "/" + rest
		}
		r.URL.Path = rest
		r.URL.RawPath = rest

		p := newReverseProxy(rt, rt.Prefix, userHeaders)
		start := time.Now()
		p.ServeHTTP(w, r)
		log.WithFields(log.Fields{
			"route":          rt.Name,
			"method":         r.Method,
			"path":           r.URL.Path,
			"duration_ms":    time.Since(start).Milliseconds(),
			"auth":           rt.Auth.String(),
			"effective_auth": effectiveAuth.String(),
			"ip":             ip,
		}).Info("proxy")
	})

	return corsMiddleware(cfg, mux)
}
