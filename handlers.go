package web

import (
	"net/http"

	"github.com/pkg/errors"
)

// NotFoundHandler is the default "404 not found" handler. It simply calls http.NotFound()
func NotFoundHandler(c *Context) error {
	http.NotFound(c.W, c.R)
	return nil
}

// ErrorHandler is the default error handler. The error is never nil.
// It will log the error and call http.Error(). Defaults to "500 internal server error", but if it's an ErrorClient it
// will send a custom status code and error message (from ErrorClient).
func ErrorHandler(c *Context, e error) {
	switch err := errors.Cause(e).(type) {
	case *ErrorClient:
		c.M.logError("Client", err, err.Stack())
		http.Error(c.W, err.Error(), err.Status())
	case *ErrorServer:
		c.M.logError("Server", err, err.Stack())
		http.Error(c.W, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	case *ErrorPanic:
		c.M.logError("Panic", err, err.Stack())
		http.Error(c.W, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	default:
		c.M.logError("Unknown", err)
		http.Error(c.W, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

// PanicHandler is the default handler called whenever a panic was recovered inside a handler. It simply calls
// HandleError() with a ErrorPanic.
func PanicHandler(c *Context, ret interface{}) {
	c.M.opt.HandleError(c, NewErrorPanic(ret))
}
