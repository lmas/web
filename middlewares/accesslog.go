package middlewares

import (
	"log"
	"net/http"
	"time"
)

// AccessLog is a middleware that prints request access logs to a log.Logger
// of your choice.
func AccessLog(l *log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			dur := time.Since(start)
			l.Printf("%s\t %s\t %s\t %s\n",
				r.RemoteAddr, r.Method, r.URL.Path, dur)
		})
	}
}
