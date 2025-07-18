package middleware

import (
	"net/http"
)

func NewJsonMiddleware() func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Здесь логика middleware

			// Вызов следующей функции в цепочке
			next(w, r)
		})
	}
}

func NewUrlPathMiddleware() func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Здесь логика middleware

			// Вызов следующей функции в цепочке
			next(w, r)
		})
	}
}
