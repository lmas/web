package web

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path"
	"sync"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"
)

// HandlerFunc is a shorter convenience function signature for http handlers,
// instead of func(http.ResponseWriter, *http.Request). It also allows for
// easier error handling.
type HandlerFunc func(*Context) error

// RegisterFunc is a function signature used when you want to register multiple
// handlers under a common URL path. First string is method, second string is
// the URL and last field is the HandlerFunc you want to register.
type RegisterFunc func(string, string, HandlerFunc)

// Options contains all the optional settings for a Handler.
type Options struct {
	// Simple logger
	Log *log.Logger
	// Templates that can be rendered using context.Render()
	Templates map[string]*template.Template
}

////////////////////////////////////////////////////////////////////////////////

type httpError struct {
	status int
	msg    string
}

func (e *httpError) Error() string {
	return e.msg
}

func (e *httpError) String() string {
	return fmt.Sprintf("Error: %q", e.msg)
}

func (e *httpError) Status() int {
	return e.status
}

////////////////////////////////////////////////////////////////////////////////

// Handler implements the http.Handler interface and allows you to easily
// register handlers and middleware with sane defaults.
// It uses github.com/julienschmidt/httprouter, for quick and easy routing.
type Handler struct {
	mux          *httprouter.Router
	opt          *Options
	contextPool  sync.Pool
	templatePool sync.Pool
}

// New returns a new Handler that implements the http.Handler interface and can
// be run with http.ListenAndServe(":8000", handler).
// You can optionally proved an Options struct with custom settings.
// Any panics caused by a registered handler will be caught and optionaly logged.
func New(opt *Options) *Handler {
	if opt == nil {
		opt = &Options{}
	}
	h := &Handler{
		opt: opt,
	}
	h.contextPool.New = h.newContext
	h.templatePool.New = h.newTemplateBuff

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
	if h.opt.Log == nil {
		return
	}
	if msg != "" {
		h.opt.Log.Printf(msg+"\n", args...)
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
// It will also handle any errors returned from the handler, by responding to
// the erroring request with http.Error().
// You can optionally use one or more http.Handler middleware. First middleware
// in the list will be executed first, and then it loops forward through all
// middlewares and lasty executes the request handler last.
func (h *Handler) Register(method, path string, handler HandlerFunc, mw ...MiddlewareFunc) {
	wrapped := h.wrapMiddleware(handler, mw...)
	h.mux.Handle(method, path, func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		c := h.getContext(w, r, p)
		defer h.putContext(c)
		err := wrapped(c)
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
