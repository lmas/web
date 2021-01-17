package web

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sync"

	"github.com/julienschmidt/httprouter"
)

// RegisterFunc is a function signature used when you want to register multiple handlers under a common URL path.
// First string is method, second string is the URL and last field is the Handler you want to register.
type RegisterFunc func(string, string, Handler)

// MuxOptions contains all the optional settings for a Mux.
type MuxOptions struct {
	// Simple logger
	Log *log.Logger
	// Templates that can be rendered using context.Render()
	Templates map[string]*template.Template
	// HandleNotFound is a Handler that will be called for '404 not found" errors. If not set it will default to
	// the SimpleNotFoundHandler() handler.
	HandleNotFound Handler
	// HandleError is a ErrorHandler that will be called for all errors returned from a Handler (except for
	// "404 not found"). It defaults to SimpleErrorHandler().
	HandleError ErrorHandler
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

////////////////////////////////////////////////////////////////////////////////////////////////////

// NewMux returns a new Mux that implements the http.Handler interface and can be run with
// http.ListenAndServe(":8000", handler).
// You can optionally proved an MuxOptions struct with custom settings.
// Any panics caused by a registered handler will be caught and optionaly logged.
func NewMux(opt *MuxOptions) *Mux {
	if opt == nil {
		opt = &MuxOptions{}
	}
	if opt.HandleNotFound == nil {
		opt.HandleNotFound = SimpleNotFoundHandler
	}
	if opt.HandleError == nil {
		opt.HandleError = SimpleErrorHandler
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
	}

	opt.HandleNotFound = m.wrap(opt.HandleNotFound)
	m.mux.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.run(opt.HandleNotFound, w, r, nil)
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

func (m *Mux) run(h Handler, w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	c := m.getContext(w, r, p)
	err := h(c)
	m.putContext(c)
	if err != nil {
		m.log("Error: %+v", err)
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// ServeHTTP implements the http.Handler interface.
func (m *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.mux.ServeHTTP(w, r)
}

// Register registers a new handler for a certain http method and URL. It will also handle any errors returned from the
// handler, by responding to the erroring request with http.Error().
// You can optionally use one or more http.Handler middleware. First middleware in the list will be executed first, and
// then it loops forward through all middlewares and lasty executes the request handler last.
func (m *Mux) Register(method, url string, handler Handler, mw ...Middleware) {
	wrapped := m.wrap(handler, mw...)
	m.mux.Handle(method, url, func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		m.run(wrapped, w, r, p)
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
		panic("file doesn't exist: " + file)
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
