package gateway

import (
	"net/url"
	"strings"
	"sync/atomic"
)

type AuthMode int

const (
	AuthNone AuthMode = iota
	AuthAPIKey
	AuthJWT
)

func (a AuthMode) String() string {
	switch a {
	case AuthNone:
		return "none"
	case AuthAPIKey:
		return "api_key"
	case AuthJWT:
		return "jwt"
	default:
		return "unknown"
	}
}

func parseAuthMode(s string) AuthMode {
	switch strings.ToLower(s) {
	case "jwt":
		return AuthJWT
	case "api_key":
		return AuthAPIKey
	case "none", "":
		return AuthNone
	default:
		return AuthNone
	}
}

type Backend struct {
	URL     *url.URL
	Healthy atomic.Bool
	Fails   atomic.Int64
}

type Route struct {
	Name           string
	Prefix         string
	StripPrefix    string
	Auth           AuthMode
	RequireAPIKey  bool
	PublicPrefixes []string
	Targets        []*Backend
	rrIdx          atomic.Uint64
}

type Registry struct {
	Routes []*Route
}

func (r *Registry) Add(rt *Route) {
	r.Routes = append(r.Routes, rt)
}

func (r *Registry) Match(path string) (*Route, string) {
	var best *Route
	bestLen := -1
	rest := "/"
	for _, rt := range r.Routes {
		if strings.HasPrefix(path, rt.Prefix) {
			if len(rt.Prefix) > bestLen {
				best = rt
				bestLen = len(rt.Prefix)
				rest = strings.TrimPrefix(path, rt.Prefix)
				if !strings.HasPrefix(rest, "/") {
					rest = "/" + rest
				}
			}
		}
	}
	return best, rest
}

func isPublicPath(rt *Route, rest string) bool {
	for _, p := range rt.PublicPrefixes {
		if strings.HasPrefix(rest, p) {
			return true
		}
	}
	return false
}

func (rt *Route) pickHealthyBackend() *Backend {
	n := len(rt.Targets)
	if n == 0 {
		return nil
	}
	start := int(rt.rrIdx.Add(1))
	for i := 0; i < n; i++ {
		b := rt.Targets[(start+i)%n]
		if b.Healthy.Load() {
			return b
		}
	}
	return rt.Targets[(start)%n]
}
