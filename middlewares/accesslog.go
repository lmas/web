package middlewares

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/lmas/web"
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
func AccessLog(l *log.Logger) func(web.Handler) web.Handler {
	if l == nil {
		// There's no point trying to run this middleware without providing a logger, hence the hard panic
		panic("accesslog: missing *log.Logger")
	}
	return func(next web.Handler) web.Handler {
		return web.Handler(func(c *web.Context) error {
			c.W = &recorder{c.W, http.StatusOK, 0}
			start := time.Now()
			err := next(c)
			dur := time.Since(start)

			// It's more efficient (speed, memory allocs) to concat a string like this..
			l.Println(c.R.Host +
				" " + c.R.RemoteAddr +
				" \"" + c.R.Method + " " + c.R.URL.Path + " " + c.R.Proto + "\"" +
				" " + c.W.(*recorder).Status() +
				" " + c.W.(*recorder).Bytes() + "kb" +
				" " + dur.String() +
				" \"" + c.R.Referer() + "\"" +
				" \"" + c.R.UserAgent() + "\"")
			c.W = c.W.(*recorder).ResponseWriter
			return err
		})
	}
}
