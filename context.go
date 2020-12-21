package web

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// Context is a convenience struct for easing the handling of a http request.
type Context struct {
	H *Handler
	W http.ResponseWriter
	R *http.Request
	P httprouter.Params
}

////////////////////////////////////////////////////////////////////////////////
// Utilize a sync.Pool to keep recently unused context objects for later reuse.

func (h *Handler) newContext() *Context {
	return &Context{
		H: h,
	}
}

func (h *Handler) getContext() *Context {
	return h.contextPool.Get().(*Context)
}

func (h *Handler) putContext(c *Context) {
	h.contextPool.Put(c)
}

func (c *Context) reset(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	c.W = w
	c.R = r
	c.P = p
}

////////////////////////////////////////////////////////////////////////////////

// Error returns a special http error, for which you can specify the http
// response status.
func (c *Context) Error(status int, msg string) error {
	return &httpError{status, msg}
}

////////////////////////////////////////////////////////////////////////////////

// JSON is a helper for JSON encoding the data and sending it with a response status.
func (c *Context) JSON(status int, data interface{}) error {
	// Basicly copied from http.Error()
	c.SetHeader("Content-Type", "application/json; charset=utf-8")
	c.SetHeader("X-Content-Type-Options", "nosniff")
	c.W.WriteHeader(status)
	return json.NewEncoder(c.W).Encode(data)
}

// DecodeJSON is a helper for JSON decoding a request body.
func (c *Context) DecodeJSON(data interface{}) error {
	defer c.R.Body.Close()
	return json.NewDecoder(c.R.Body).Decode(data)
}

////////////////////////////////////////////////////////////////////////////////

// SetHeader is a shortcut to set a header value for a response.
func (c *Context) SetHeader(key, value string) {
	c.W.Header().Set(key, value)
}

// GetHeader is a shortcut to get a header value from a request.
func (c *Context) GetHeader(key string) string {
	return c.R.Header.Get(key)
}

////////////////////////////////////////////////////////////////////////////////

// GetParams is a shortcut to get URL params, first one given by key.
// See https://pkg.go.dev/github.com/julienschmidt/httprouter#Param for more info.
func (c *Context) GetParams(key string) string {
	return c.P.ByName(key)
}
