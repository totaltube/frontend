package middlewares

import "net/http"

type headerCheckWriter struct {
	http.ResponseWriter
	headersSent bool
}

func (hcw *headerCheckWriter) WriteHeader(statusCode int) {
	if !hcw.headersSent {
		hcw.ResponseWriter.WriteHeader(statusCode)
		hcw.headersSent = true
	}
}

func (hcw *headerCheckWriter) Write(b []byte) (int, error) {
	if !hcw.headersSent {
		hcw.WriteHeader(http.StatusOK)
	}
	return hcw.ResponseWriter.Write(b)
}

func HeadersSent(w http.ResponseWriter) bool {
	if hcw, ok := w.(*headerCheckWriter); ok {
		return hcw.headersSent
	}
	return false
}

func HeadersSentMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hcw := &headerCheckWriter{ResponseWriter: w}
		next.ServeHTTP(hcw, r)
	})
}
