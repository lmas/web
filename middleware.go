package web

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// Middleware is a function signature used when wrapping a HandlerFunc in one or many http.Handler middlewares
// that's pretty common in the go community.
type Middleware func(http.Handler) http.Handler

// wrap wraps a HandlerFunc in one or more http.Handler middlewares.
func (m *Mux) wrapMiddleware(handler HandlerFunc, mw ...Middleware) HandlerFunc {
	mw = append(m.opt.Middlewares, mw...)
	if len(mw) < 1 {
		return handler
	}

	var err error
	var params httprouter.Params
	var wrapped http.Handler
	wrapped = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c := m.getContext(w, r, params)
		err = handler(c)
		m.putContext(c)
	})
	for i := len(mw) - 1; i >= 0; i-- {
		if mw[i] == nil {
			panic("Trying to use a nil pointer as middleware")
		}
		wrapped = mw[i](wrapped)
	}

	return HandlerFunc(func(ctx *Context) error {
		params = ctx.P
		wrapped.ServeHTTP(ctx.W, ctx.R)
		return err
	})
}
