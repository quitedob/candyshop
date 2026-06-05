package realtime

import (
	"net/http"
	"strings"
)

// OriginAllowList builds an OriginChecker that permits same-origin requests
// (no Origin header, e.g. non-browser clients) and any Origin present in the
// configured allow-list. A "*" entry allows every origin (development only).
func OriginAllowList(allowed []string) OriginChecker {
	set := make(map[string]struct{}, len(allowed))
	wildcard := false
	for _, o := range allowed {
		o = strings.TrimSpace(o)
		if o == "" {
			continue
		}
		if o == "*" {
			wildcard = true
			continue
		}
		set[strings.ToLower(strings.TrimRight(o, "/"))] = struct{}{}
	}

	return func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			// Non-browser clients (and same-origin server-to-server) omit Origin.
			return true
		}
		if wildcard {
			return true
		}
		_, ok := set[strings.ToLower(strings.TrimRight(origin, "/"))]
		return ok
	}
}
