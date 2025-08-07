package middleware

import (
	"net/http"
	"regexp"
)

func ValidateRoute(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !isValidRoute(r.URL.Path) {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func isValidRoute(path string) bool {
	if match, _ := regexp.MatchString(`^/update/[a-zA-Z0-9_-]+/[a-zA-Z0-9_-]+/(-?)[a-zA-Z0-9\\.]+$`, path); match {
		return true
	}

	if path == "/update" || path == "/update/" || path == "/updates" || path == "/updates/" {
		return true
	}

	return false
}
