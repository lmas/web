package web

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// MiddlewareFunc is a function signature used when wrapping a HandlerFunc in
// one or many http.Handler middlewares that's pretty common in the go community.
type MiddlewareFunc func(http.Handler) http.Handler

// wrap wraps a HandlerFunc in one or more http.Handler middlewares.
func (h *Handler) wrapMiddleware(handler HandlerFunc, mw ...MiddlewareFunc) HandlerFunc {
	if len(mw) < 1 {
		return handler
	}

	var err error
	var params httprouter.Params
	var wrapped http.Handler
	wrapped = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c := h.getContext()
		c.reset(w, r, params)
		err = handler(c)
		h.putContext(c)
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
