package middleware

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"
)

func APIKeyAuth(apiKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			got := r.Header.Get("Api-Key")
			if subtle.ConstantTimeCompare([]byte(got), []byte(apiKey)) != 1 {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": "bad api key"})

				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
