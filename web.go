package web

import (
	"fmt"
	"log"
	"net/http"
	"path"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"
)

// HandlerFunc is a shorter convenience function signature for http handlers,
// instead of func(http.ResponseWriter, *http.Request). It also allows for
// easier error handling.
type HandlerFunc func(Context) error

// RegisterFunc is a function signature used when you want to register multiple
// handlers under a common URL path. First string is method, second string is
// the URL and last field is the HandlerFunc you want to register.
type RegisterFunc func(string, string, HandlerFunc)

// MiddlewareFunc is a function signature used when wrapping a HandlerFunc in
// one or many http.Handler middlewares that's pretty common in the go community.
type MiddlewareFunc func(http.Handler) http.Handler

// wrap wraps a HandlerFunc in one or more http.Handler middlewares.
func wrap(handler HandlerFunc, mw ...MiddlewareFunc) HandlerFunc {
	if len(mw) < 1 {
		return handler
	}

	var err error
	var params httprouter.Params
	var wrapped http.Handler
	wrapped = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err = handler(Context{w, r, params})
	})
	for i := len(mw) - 1; i >= 0; i-- {
		if mw[i] == nil {
			panic("Trying to use a nil pointer as middleware")
		}
		wrapped = mw[i](wrapped)
	}

	return HandlerFunc(func(ctx Context) error {
		params = ctx.P
		wrapped.ServeHTTP(ctx.W, ctx.R)
		return err
	})
}

////////////////////////////////////////////////////////////////////////////////

// Handler implements the http.Handler interface and allows you to easily
// register handlers and middleware with sane defaults.
// It uses github.com/julienschmidt/httprouter, for quick and easy routing.
type Handler struct {
	logger *log.Logger
	mux    *httprouter.Router
}

// New returns a new Handler ready to be used. You can provide an optional
// log.Logger to enable logging.
// Start serve requests by running it with a http.ListenAndServe(":8000", *Handler)
// call.
// Any panics caused by a registered handler will be caught and logged.
func New(l *log.Logger) *Handler {
	h := &Handler{
		logger: l,
	}

	h.mux = &httprouter.Router{
		RedirectTrailingSlash:  true,
		RedirectFixedPath:      true,
		HandleMethodNotAllowed: true,
		HandleOPTIONS:          true,
		PanicHandler: func(w http.ResponseWriter, r *http.Request, ret interface{}) {
			h.logRequest(r, fmt.Sprintf("%+v", ret))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		},
	}
	return h
}

func (h *Handler) log(msg string, args ...interface{}) {
	if h.logger == nil {
		return
	}
	if msg != "" {
		h.logger.Printf(msg+"\n", args...)
	}
}

func (h *Handler) logRequest(r *http.Request, msg string) {
	h.log("%s\t %s\t %s\t %s", r.RemoteAddr, r.Method, r.URL.Path, msg)
}

// ServeHTTP implements the http.Handler interface.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	h.mux.ServeHTTP(w, r)
	dur := time.Since(start)
	h.logRequest(r, dur.String())
}

// Register registers a new handler for a certain http method and URL.
// It will also handle any errors returned from the handler, by responsing to
// the erroring request with http.Error().
// You can optionally use one or more http.Handler middleware. First middleware
// in the list will be executed first, and then it loops forward through all
// middlewares and lasty executes the request handler last.
func (h *Handler) Register(method, path string, handler HandlerFunc, mw ...MiddlewareFunc) {
	wrapped := wrap(handler, mw...)
	h.mux.Handle(method, path, func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		err := wrapped(Context{w, r, p})
		if err != nil {
			h.logRequest(r, fmt.Sprintf("%+v", err))
			switch err := errors.Cause(err).(type) {
			case *httpError:
				http.Error(w, err.String(), err.Status())
			default:
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}
	})
}

// RegisterPrefix returns a RegisterFunc function that you can call multiple
// times to register multiple handlers under a common URL prefix.
// You can optionally use middlewares too, the same way as in Register().
func (h *Handler) RegisterPrefix(prefix string, mw ...MiddlewareFunc) RegisterFunc {
	return func(m, p string, handler HandlerFunc) {
		h.Register(m, path.Join(prefix, p), handler, mw...)
	}
}
