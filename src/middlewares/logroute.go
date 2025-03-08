package middlewares

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func LogRouteMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		routePattern := chi.RouteContext(r.Context()).RoutePattern()
		log.Printf("Обработчик вызван для маршрута: %s %s -> %s", r.Method, r.URL.Path, routePattern)
		next.ServeHTTP(w, r)
	})
}
