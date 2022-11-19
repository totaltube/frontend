package middlewares

import (
	"context"
	"log"
	"net/http"
	"time"
)

func Timeout(timeout time.Duration) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer func() {
				cancel()
				if ctx.Err() == context.DeadlineExceeded {
					log.Println("timeout requesting", r.URL.String())
					w.WriteHeader(http.StatusGatewayTimeout)
				}
			}()
			next.ServeHTTP(w, r.WithContext(ctx))
		}
		return http.HandlerFunc(fn)
	}
}