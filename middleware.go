package web

import (
	"fmt"

	"github.com/pkg/errors"
)

// Middleware is a function signature used when wrapping a Handler in one or many Handler middlewares.
type Middleware func(Handler) Handler

// recoverErrors is a middleware that catches panics and handle errors
func (m *Mux) recoverErrors(next Handler) Handler {
	rec := func(c *Context) (err error) {
		// Do some recover magic with named returns
		defer func() {
			if r := recover(); r != nil {
				err = errors.New(fmt.Sprintf("panic: %v", r))
			}
		}()
		err = next(c)
		return
	}
	return Handler(func(c *Context) error {
		err := rec(c)
		if err != nil {
			err = m.opt.HandleError(c, err)
		}
		return err
	})
}

// wrap a Handler in one or more Middlewares.
func (m *Mux) wrap(h Handler, mw ...Middleware) Handler {
	h = m.recoverErrors(h)
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
