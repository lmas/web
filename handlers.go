package web

import (
	"net/http"

	"github.com/pkg/errors"
)

// Handler is a shorter convenience function signature for http handlers, instead of
// func(http.ResponseWriter, *http.Request). It also allows for easier error handling.
type Handler func(*Context) error

// ErrorHandler is a sort of a Handler, but it also takes an error.
type ErrorHandler func(*Context, error) error

// SimpleNotFoundHandler is the default "404 not found" handler. It simply calls http.NotFound()
func SimpleNotFoundHandler(c *Context) error {
	http.NotFound(c.W, c.R)
	return nil
}

// SimpleErrorHandler is the default handler for handling handler errors (that sounded sick).
// It checks if an error is a Error and sends it's status code and msg as the http response. If it's not, it simply
// sends an "500 internal server error" for all other errors.
func SimpleErrorHandler(c *Context, err error) error {
	switch e := errors.Cause(err).(type) {
	case *Error:
		http.Error(c.W, e.Error(), e.Status())
		return nil // Don't wanna log client errors
	default:
		http.Error(c.W, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return err
	}
}
