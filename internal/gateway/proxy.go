package gateway

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"
)

type lbTransport struct {
	route *Route
	base  *http.Transport
}

func (t *lbTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	b := t.route.pickHealthyBackend()
	if b == nil {
		return nil, errors.New("no backends")
	}

	clone := *req
	urlCopy := *req.URL
	urlCopy.Scheme = b.URL.Scheme
	urlCopy.Host = b.URL.Host
	urlCopy.Path = singleJoin(b.URL.Path, req.URL.Path)
	clone.URL = &urlCopy

	clone.Header = req.Header.Clone()
	clone.Header.Set("X-Forwarded-Host", req.Host)
	clone.Header.Set("X-Forwarded-Proto", schemeOf(req))
	clone.Header.Set("X-Forwarded-For", clientIP(req))

	resp, err := t.base.RoundTrip(&clone)
	if err != nil {
		b.Fails.Add(1)
		if b.Fails.Load() > 3 {
			b.Healthy.Store(false)
		}
		return nil, err
	}

	if resp.StatusCode == 502 || resp.StatusCode == 503 || resp.StatusCode == 504 {
		b.Fails.Add(1)
		if b.Fails.Load() > 3 {
			b.Healthy.Store(false)
		}
	} else {
		b.Fails.Store(0)
		b.Healthy.Store(true)
	}
	return resp, nil
}

func newReverseProxy(rt *Route, basePath string, userHeaders map[string]string) *httputil.ReverseProxy {
	return &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.Header.Set("X-Forwarded-Prefix", basePath)
			for k, v := range userHeaders {
				req.Header.Set(k, v)
			}
			// path already stripped in handler
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			jsonError(w, r, http.StatusBadGateway, "BAD_GATEWAY", "Upstream service error", err.Error())
		},
		Transport: &lbTransport{route: rt, base: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			DialContext:           (&netDialerWithTimeout{}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          500,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		}},
	}
}

func singleJoin(a, b string) string {
	if a == "" {
		a = "/"
	}
	if !strings.HasPrefix(a, "/") {
		a = "/" + a
	}
	if !strings.HasPrefix(b, "/") {
		b = "/" + b
	}
	return strings.TrimRight(a, "/") + "/" + strings.TrimLeft(b, "/")
}

func schemeOf(r *http.Request) string {
	if v := r.Header.Get("X-Forwarded-Proto"); v != "" {
		return v
	}
	if r.TLS != nil {
		return "https"
	}
	return "http"
}

func clientIP(r *http.Request) string {
	if xf := r.Header.Get("X-Forwarded-For"); xf != "" {
		return xf
	}
	host := r.RemoteAddr
	if idx := strings.LastIndex(host, ":"); idx > 0 {
		return host[:idx]
	}
	return host
}

type netDialerWithTimeout struct{}

func (d *netDialerWithTimeout) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	var nd net.Dialer
	nd.Timeout = 10 * time.Second
	nd.KeepAlive = 30 * time.Second
	return nd.DialContext(ctx, network, address)
}
