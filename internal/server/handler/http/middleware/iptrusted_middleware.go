package middleware

import (
	"net"
	"net/http"
	"strings"
)

func TrustedSubnetMiddleware(subnet string) func(http.Handler) http.Handler {
	_, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		panic("invalid subnet: " + err.Error())
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if subnet == "" {
				next.ServeHTTP(w, r)
			}
			realIP := r.Header.Get("X-Real-IP")
			if realIP == "" {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			ip := net.ParseIP(strings.TrimSpace(realIP))
			if ip == nil || !ipNet.Contains(ip) {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
