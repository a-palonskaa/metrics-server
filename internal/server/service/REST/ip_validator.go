package serverREST

import (
	"net"
	"net/http"

	"github.com/rs/zerolog/log"
)

// ValidateIP creates a middleware wrapper for validating IP
// func checks that IP is on a trusted subnet
func ValidateIP(trustedSubnet string) func(fn http.Handler) http.Handler {
	return func(fn http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if trustedSubnet != "" {
				ip := ""
				if ip = r.Header.Get("X-Real-IP"); ip == "" {
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				if status := isIPtrustful(ip, trustedSubnet); status != http.StatusOK {
					w.WriteHeader(status)
					return
				}
			}

			fn.ServeHTTP(w, r)
		})
	}
}

func isIPtrustful(ipRaw string, trustedSubnet string) int {
	_, subnet, err := net.ParseCIDR(trustedSubnet)
	if err != nil {
		log.Info().Err(err).Msg("invalid trusted subnet")
		return http.StatusInternalServerError
	}

	ip := net.ParseIP(ipRaw)
	if ip == nil {
		log.Info().Err(err).Msgf("ip %s is not a valid textual representation of an IP address", ipRaw)
		return http.StatusForbidden
	}

	if !subnet.Contains(ip) {
		return http.StatusForbidden
	}
	return http.StatusOK
}
