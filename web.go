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

type HandlerFunc func(Context) error

type RegisterFunc func(string, string, HandlerFunc)

type Handler struct {
	logger *log.Logger
	mux    *httprouter.Router
}

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

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	h.mux.ServeHTTP(w, r)
	dur := time.Since(start)
	h.logRequest(r, dur.String())
}

func (h *Handler) Register(method, path string, handler HandlerFunc) {
	h.mux.Handle(method, path, func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		err := handler(Context{w, r, p})
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

func (h *Handler) RegisterPrefix(prefix string, f func(RegisterFunc)) {
	f(func(m, p string, handler HandlerFunc) {
		h.Register(m, path.Join(prefix, p), handler)
	})
}
