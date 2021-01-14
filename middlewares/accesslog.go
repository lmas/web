package middlewares

import (
	"log"
	"net/http"
	"strconv"
	"time"
)

// recorder wraps ResponseWriter so we can get the response status code
type recorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (w *recorder) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *recorder) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.bytes += n
	return n, err
}

func (w *recorder) Status() string {
	return strconv.Itoa(w.status)
}

func (w *recorder) Bytes() string {
	return strconv.FormatFloat(float64(w.bytes)/1024.0, 'f', 1, 64)
	// If printing a float (in kb) will ever cause any problems or something, you could always stick with plain and
	// simple bytes as an int
	//return strconv.Itoa(w.bytes)
}

// AccessLog is a middleware that prints request access logs to a log.Logger
// of your choice.
// If *log.Logger is nil, a panic will be raised.
//
// Columns for each output line:
// Host
// Remote client addr
// HTTP request (method, path, protocol)
// HTTP response code
// Response body size (in kb)
// Response run time
// HTTP Referer
// Client User-Agent
func AccessLog(l *log.Logger) func(http.Handler) http.Handler {
	if l == nil {
		// There's no point trying to run this middleware without providing a logger, hence the hard panic
		panic("accesslog: missing *log.Logger")
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w2 := &recorder{w, http.StatusOK, 0}
			start := time.Now()
			next.ServeHTTP(w2, r)
			dur := time.Since(start)
			// It's more efficient (speed, memory allocs) to concat a string like this..
			l.Println(r.Host +
				" " + r.RemoteAddr +
				" \"" + r.Method + " " + r.URL.Path + " " + r.Proto + "\"" +
				" " + w2.Status() +
				" " + w2.Bytes() + "kb" +
				" " + dur.String() +
				" \"" + r.Referer() + "\"" +
				" \"" + r.UserAgent() + "\"")
		})
	}
}
