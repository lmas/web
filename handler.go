package web

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sync"

	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"
)

// Handler is a shorter convenience function signature for http handlers, instead of
// func(http.ResponseWriter, *http.Request). It also allows for easier error handling.
type Handler func(*Context) error

// Middleware is a function signature used when wrapping a Handler in one or many Handler middlewares.
type Middleware func(Handler) Handler

// RegisterFunc is a function signature used when you want to register multiple handlers under a common URL path.
// First string is method, second string is the URL and last field is the Handler you want to register.
type RegisterFunc func(string, string, Handler)

// NotFound is the default "404 not found" handler. It simply calls http.NotFound()
func NotFound(c *Context) error {
	http.NotFound(c.W, c.R)
	return nil
}

// MuxOptions contains all the optional settings for a Mux.
type MuxOptions struct {
	// Simple logger
	Log *log.Logger
	// Templates that can be rendered using context.Render()
	Templates map[string]*template.Template
	// NotFound is a Handler that will be called for '404 not found" errors. If not set it will default to the
	// NotFound() handler.
	NotFound Handler
	// Middlewares is a list of middlewares that will be globaly added to all handlers
	Middlewares []Middleware
}

// Mux implements the http.Handler interface and allows you to easily register handlers and middleware with sane
// defaults. It uses github.com/julienschmidt/httprouter, for quick and easy routing.
type Mux struct {
	mux          *httprouter.Router
	opt          *MuxOptions
	contextPool  sync.Pool
	templatePool sync.Pool
}

// HTTPError is a custom error which also contains a http status code. Whenever a Handler returns this error, the error
// message and status code will be sent back to the client unaltered.
type HTTPError struct {
	status int
	msg    string
}

func (e *HTTPError) Error() string {
	return e.msg
}

func (e *HTTPError) String() string {
	return fmt.Sprintf("Error: %q", e.msg)
}

// Status returns the HTTP status code for this error.
func (e *HTTPError) Status() int {
	return e.status
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// NewMux returns a new Mux that implements the http.Handler interface and can be run with
// http.ListenAndServe(":8000", handler).
// You can optionally proved an MuxOptions struct with custom settings.
// Any panics caused by a registered handler will be caught and optionaly logged.
func NewMux(opt *MuxOptions) *Mux {
	if opt == nil {
		opt = &MuxOptions{}
	}
	if opt.NotFound == nil {
		opt.NotFound = NotFound
	}
	m := &Mux{
		opt: opt,
	}

	m.contextPool.New = m.newContext
	m.templatePool.New = m.newTemplateBuff
	m.mux = &httprouter.Router{
		RedirectTrailingSlash:  true,
		RedirectFixedPath:      true,
		HandleMethodNotAllowed: true,
		HandleOPTIONS:          true,
		PanicHandler: func(w http.ResponseWriter, r *http.Request, ret interface{}) {
			m.logPanic(ret)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		},
	}
	opt.NotFound = m.wrap(opt.NotFound)
	m.mux.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c := &Context{m, w, r, nil}
		if err := opt.NotFound(c); err != nil {
			m.handlerError(c, err)
		}
	})
	return m
}

func (m *Mux) log(msg string, args ...interface{}) {
	if m.opt.Log == nil {
		return
	}
	if msg != "" {
		m.opt.Log.Printf(msg+"\n", args...)
	}
}

func (m *Mux) logError(err error) {
	m.log("Error: %+v", err)
}

func (m *Mux) logPanic(ret interface{}) {
	m.log("Panic: %+v", ret)
}

// ServeHTTP implements the http.Handler interface.
func (m *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.mux.ServeHTTP(w, r)
}

// wrap wraps a Handler in one or more Middlewares.
func (m *Mux) wrap(h Handler, mw ...Middleware) Handler {
	mw = append(m.opt.Middlewares, mw...)
	if len(mw) < 1 {
		return h
	}

	for i := len(mw) - 1; i >= 0; i-- {
		if mw[i] == nil {
			panic("Trying to use a nil pointer as middleware")
		}
		h = mw[i](h)
	}
	return h
}

func (m *Mux) handlerError(c *Context, e error) {
	m.logError(e)
	switch err := errors.Cause(e).(type) {
	case *HTTPError:
		http.Error(c.W, err.Error(), err.Status())
	default:
		http.Error(c.W, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// Register registers a new handler for a certain http method and URL. It will also handle any errors returned from the
// handler, by responding to the erroring request with http.Error().
// You can optionally use one or more http.Handler middleware. First middleware in the list will be executed first, and
// then it loops forward through all middlewares and lasty executes the request handler last.
func (m *Mux) Register(method, url string, handler Handler, mw ...Middleware) {
	wrapped := m.wrap(handler, mw...)
	m.mux.Handle(method, url, func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		c := m.getContext(w, r, p)
		defer m.putContext(c)
		if err := wrapped(c); err != nil {
			m.handlerError(c, err)
		}
	})
}

// RegisterPrefix returns a RegisterFunc function that you can call multiple times to register multiple handlers under
// a common URL prefix.
// You can optionally use middlewares too, the same way as in Register().
func (m *Mux) RegisterPrefix(prefix string, mw ...Middleware) RegisterFunc {
	return func(method, url string, handler Handler) {
		m.Register(method, path.Join(prefix, url), handler, mw...)
	}
}

// File is a helper to serve a simple http GET response for a single file. If the file "disappears" while the server is
// running, a 404 Not found will be returned.
// NOTE: if the file doesn't exist at start up, it will cause a panic instead.
// You can optionally use middlewares too, the same way as in Register().
func (m *Mux) File(url, file string, mw ...Middleware) {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		panic(fmt.Errorf("file doesn't exist: %s", file))
	}
	fs := http.Dir(filepath.Dir(file))
	m.Register("GET", url, func(c *Context) error {
		return c.File(fs, file)
	}, mw...)
}

// Static is a helper to serve a whole directory with static files.
// You can optionally use middlewares too, the same way as in Register().
func (m *Mux) Static(dir string, fs http.FileSystem, mw ...Middleware) {
	m.Register("GET", path.Join(dir, "/*filepath"), func(c *Context) error {
		fp := httprouter.CleanPath(c.GetParams("filepath")) // Not sure if it's already cleaned but oh well...
		return c.File(fs, fp)
	}, mw...)
}
